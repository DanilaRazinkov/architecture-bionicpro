package storage

import (
	"auth/internal/config"
	"auth/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg config.RedisConfig) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &RedisClient{client: rdb}
}

func (r *RedisClient) SetSession(ctx context.Context, session *models.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("session:%s", session.ID), data, 24*time.Hour)
	if session.AuthCode != "" {
		pipe.Set(ctx, fmt.Sprintf("code:%s", session.AuthCode), session.ID, 24*time.Hour)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisClient) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	data, err := r.client.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
	if err != nil {
		return nil, err
	}

	var session models.Session
	err = json.Unmarshal([]byte(data), &session)
	return &session, err
}

func (r *RedisClient) GetSessionByCode(ctx context.Context, code string) (*models.Session, error) {
	sessionID, err := r.client.Get(ctx, fmt.Sprintf("code:%s", code)).Result()
	if err != nil {
		return nil, err
	}

	return r.GetSession(ctx, sessionID)
}

func (r *RedisClient) DeleteSession(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, fmt.Sprintf("session:%s", sessionID)).Err()
}

func (r *RedisClient) SetAccessToken(ctx context.Context, userID, token string, expiration time.Duration) error {
	return r.client.Set(ctx, fmt.Sprintf("access_token:%s", userID), token, expiration).Err()
}

func (r *RedisClient) GetAccessToken(ctx context.Context, userID string) (string, error) {
	return r.client.Get(ctx, fmt.Sprintf("access_token:%s", userID)).Result()
}

func (r *RedisClient) DeleteAccessToken(ctx context.Context, userID string) error {
	return r.client.Del(ctx, fmt.Sprintf("access_token:%s", userID)).Err()
}
