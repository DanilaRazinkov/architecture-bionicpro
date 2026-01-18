package services

import (
	"auth/internal/models"
	"auth/internal/storage"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type SessionService struct {
	redis    *storage.RedisClient
	keycloak *KeycloakService
}

func NewSessionService(redis *storage.RedisClient, keycloak *KeycloakService) *SessionService {
	return &SessionService{
		redis:    redis,
		keycloak: keycloak,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, code, redirectURI string) (*models.Session, error) {
	tokenResp, err := s.keycloak.ExchangeCodeForToken(ctx, code, redirectURI)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.keycloak.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	sessionID := s.generateSessionID()
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	session := &models.Session{
		ID:           sessionID,
		AuthCode:     code,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt,
		UserInfo:     *userInfo,
		CreatedAt:    time.Now(),
	}

	err = s.redis.SetSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	return s.redis.GetSession(ctx, sessionID)
}

func (s *SessionService) GetSessionByCode(ctx context.Context, code string) (*models.Session, error) {
	return s.redis.GetSessionByCode(ctx, code)
}

func (s *SessionService) RefreshSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.redis.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if time.Now().Before(session.ExpiresAt.Add(-30 * time.Second)) {
		return session, nil
	}

	tokenResp, err := s.keycloak.RefreshToken(ctx, session.RefreshToken)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.keycloak.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	session.AccessToken = tokenResp.AccessToken
	session.RefreshToken = tokenResp.RefreshToken
	session.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	session.UserInfo = *userInfo

	err = s.redis.SetSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) RotateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.redis.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	newSessionID := s.generateSessionID()

	newSession := &models.Session{
		ID:           newSessionID,
		AuthCode:     session.AuthCode,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt,
		UserInfo:     session.UserInfo,
		CreatedAt:    time.Now(),
	}

	err = s.redis.SetSession(ctx, newSession)
	if err != nil {
		return nil, err
	}

	err = s.redis.DeleteSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return newSession, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.redis.DeleteSession(ctx, sessionID)
}

func (s *SessionService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
