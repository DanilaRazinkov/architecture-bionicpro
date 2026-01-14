package storage

import (
	"context"
	"fmt"
	"report/internal/config"
	"report/internal/models"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseClient struct {
	conn clickhouse.Conn
}

func NewClickHouseClient(cfg config.ClickHouseConfig) *ClickHouseClient {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to ClickHouse: %v", err))
	}
	return &ClickHouseClient{conn: conn}
}

func (c *ClickHouseClient) GetUserReport(ctx context.Context, userID int) (*models.UserReport, error) {
	username, email := c.getUserInfo(ctx, userID)

	report, err := c.getReportData(ctx, userID)
	if err != nil {
		return c.createEmptyReport(userID, username, email), nil
	}

	report.Username = username
	report.Email = email
	report.ReportGeneratedAt = time.Now()
	report.HasData = true

	return report, nil
}

func (c *ClickHouseClient) GetReportSummary(ctx context.Context) (*models.ReportSummary, error) {
	query := `
		SELECT
			COUNT(DISTINCT user_id) as total_users,
			COUNT(DISTINCT CASE WHEN last_activity > now() - INTERVAL 30 DAY THEN user_id END) as active_users,
			SUM(total_sessions) as total_sessions,
			AVG(total_usage_time) as average_usage
		FROM default.report_patient_activity_mart
	`

	row := c.conn.QueryRow(ctx, query)

	var summary models.ReportSummary
	err := row.Scan(
		&summary.TotalUsers,
		&summary.ActiveUsers,
		&summary.TotalSessions,
		&summary.AverageUsage,
	)

	return &summary, err
}

func (c *ClickHouseClient) getUserInfo(ctx context.Context, userID int) (string, string) {
	query := `SELECT name, email FROM dim_customers WHERE id = ?`
	row := c.conn.QueryRow(ctx, query, userID)

	var username, email string
	row.Scan(&username, &email)
	return username, email
}

func (c *ClickHouseClient) getReportData(ctx context.Context, userID int) (*models.UserReport, error) {
	query := `
		SELECT
			user_id,
			session_count as total_sessions,
			total_signals,
			avg_duration * session_count as total_usage_time,
			avg_duration as average_session_time,
			muscle_groups,
			avg_accuracy,
			report_date as last_activity
		FROM bionicpro.user_reports_realtime
		WHERE user_id = ?
		ORDER BY report_date DESC
		LIMIT 1
	`

	row := c.conn.QueryRow(ctx, query, userID)

	var report models.UserReport
	err := row.Scan(
		&report.UserID,
		&report.TotalSessions,
		&report.TotalSignals,
		&report.TotalUsageTime,
		&report.AverageSessionTime,
		&report.MuscleGroups,
		&report.AverageAccuracy,
		&report.LastActivity,
	)

	return &report, err
}

func (c *ClickHouseClient) createEmptyReport(userID int, username, email string) *models.UserReport {
	return &models.UserReport{
		UserID:            uint32(userID),
		Username:          username,
		Email:             email,
		HasData:           false,
		Message:           "Данные отчёта пока отсутствуют",
		ReportGeneratedAt: time.Now(),
	}
}
