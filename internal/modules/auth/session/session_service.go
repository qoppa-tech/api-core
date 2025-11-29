package session

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TokenPrefix        = "jwt:token:"
	RefreshTokenPrefix = "jwt:refresh:"
	BlacklistPrefix    = "jwt:blacklist:"
)

type SessionService struct {
	client *redis.Client
}

func NewSessionService() (*SessionService, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")

	redisDB := 0
	if redisDBStr != "" {
		db, err := strconv.Atoi(redisDBStr)
		if err == nil {
			redisDB = db
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &SessionService{
		client: client,
	}, nil
}

func (s *SessionService) StoreToken(ctx context.Context, userID, token string, expiration time.Duration) error {
	key := TokenPrefix + userID
	return s.client.Set(ctx, key, token, expiration).Err()
}

func (s *SessionService) GetToken(ctx context.Context, userID string) (string, error) {
	key := TokenPrefix + userID
	token, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	}
	return token, err
}

func (s *SessionService) DeleteToken(ctx context.Context, userID string) error {
	key := TokenPrefix + userID
	return s.client.Del(ctx, key).Err()
}

func (s *SessionService) StoreRefreshToken(ctx context.Context, userID, token string, expiration time.Duration) error {
	key := RefreshTokenPrefix + userID
	return s.client.Set(ctx, key, token, expiration).Err()
}

func (s *SessionService) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := RefreshTokenPrefix + userID
	token, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("refresh token not found")
	}
	return token, err
}

func (s *SessionService) DeleteRefreshToken(ctx context.Context, userID string) error {
	key := RefreshTokenPrefix + userID
	return s.client.Del(ctx, key).Err()
}

func (s *SessionService) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	key := BlacklistPrefix + token
	return s.client.Set(ctx, key, "1", expiration).Err()
}

func (s *SessionService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := BlacklistPrefix + token
	_, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *SessionService) UpdateTokenExpiry(ctx context.Context, userID string, expiration time.Duration) error {
	key := TokenPrefix + userID
	return s.client.Expire(ctx, key, expiration).Err()
}

func (s *SessionService) DeleteAllUserTokens(ctx context.Context, userID string) error {
	tokenKey := TokenPrefix + userID
	refreshKey := RefreshTokenPrefix + userID

	pipe := s.client.Pipeline()
	pipe.Del(ctx, tokenKey)
	pipe.Del(ctx, refreshKey)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *SessionService) GetTokenTTL(ctx context.Context, userID string) (time.Duration, error) {
	key := TokenPrefix + userID
	return s.client.TTL(ctx, key).Result()
}

func (s *SessionService) Close() error {
	return s.client.Close()
}

func (s *SessionService) Health(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
