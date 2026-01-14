package models

import "time"

type UserReport struct {
	UserID             uint32    `json:"user_id"`
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	TotalSessions      uint32    `json:"total_sessions"`
	TotalSignals       uint32    `json:"total_signals"`
	TotalUsageTime     float64   `json:"total_usage_time"`
	AverageSessionTime float64   `json:"average_session_time"`
	MuscleGroups       string    `json:"muscle_groups"`
	AverageAccuracy    float64   `json:"average_accuracy"`
	LastActivity       time.Time `json:"last_activity"`
	ReportGeneratedAt  time.Time `json:"report_generated_at"`
	HasData            bool      `json:"has_data"`
	Message            string    `json:"message,omitempty"`
}

type ReportSummary struct {
	TotalUsers    int     `json:"total_users"`
	ActiveUsers   int     `json:"active_users"`
	TotalSessions int     `json:"total_sessions"`
	AverageUsage  float64 `json:"average_usage"`
}
