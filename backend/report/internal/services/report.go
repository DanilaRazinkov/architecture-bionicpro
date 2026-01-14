package services

import (
	"context"
	"report/internal/config"
	"report/internal/models"
	"report/internal/storage"
)

type ReportService struct {
	clickhouse *storage.ClickHouseClient
	s3         *storage.S3Client
	config     *config.Config
}

func NewReportService(clickhouse *storage.ClickHouseClient, s3 *storage.S3Client, cfg *config.Config) *ReportService {
	return &ReportService{
		clickhouse: clickhouse,
		s3:         s3,
		config:     cfg,
	}
}

func (s *ReportService) GetUserReport(ctx context.Context, userID int) (*models.UserReport, error) {
	if exists, _ := s.s3.ReportExists(ctx, userID); exists {
		return s.s3.GetReport(ctx, userID)
	}

	report, err := s.clickhouse.GetUserReport(ctx, userID)
	if err != nil {
		return nil, err
	}

	if report.HasData {
		s.s3.StoreReport(ctx, userID, report)
	}

	return report, nil
}

func (s *ReportService) GenerateUserReport(ctx context.Context, userID int) (string, error) {
	if exists, _ := s.s3.ReportExists(ctx, userID); exists {
		return s.s3.GenerateCDNLink(userID, s.config.CDN.Host), nil
	}

	report, err := s.clickhouse.GetUserReport(ctx, userID)
	if err != nil {
		return "", err
	}

	if !report.HasData {
		return "", nil
	}

	if err := s.s3.StoreReport(ctx, userID, report); err != nil {
		return "", err
	}

	return s.s3.GenerateCDNLink(userID, s.config.CDN.Host), nil
}

func (s *ReportService) GetReportSummary(ctx context.Context) (*models.ReportSummary, error) {
	return s.clickhouse.GetReportSummary(ctx)
}
