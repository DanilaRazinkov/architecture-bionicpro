package models

import "time"

type Session struct {
	ID           string    `json:"id"`
	AuthCode     string    `json:"auth_code"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserInfo     UserInfo  `json:"user_info"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserInfo struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	UserID    int      `json:"user_id"`
	CRMUserID *int     `json:"crm_user_id"`
}

type LoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
