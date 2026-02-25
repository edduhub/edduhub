package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore defines the interface for token storage operations
type TokenStore interface {
	// StoreRefreshToken stores a refresh token with metadata
	StoreRefreshToken(ctx context.Context, tokenID string, userID int, expiresAt time.Duration) error

	// GetRefreshToken retrieves refresh token metadata
	GetRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenData, error)

	// DeleteRefreshToken deletes a refresh token (revocation)
	DeleteRefreshToken(ctx context.Context, tokenID string) error

	// IsTokenRevoked checks if a token has been revoked
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)

	// RevokeToken marks a token as revoked
	RevokeToken(ctx context.Context, jti string, expiry time.Duration) error
}

// RefreshTokenData holds refresh token information
type RefreshTokenData struct {
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RedisTokenStore implements TokenStore using Redis
type RedisTokenStore struct {
	client *redis.Client
}

// NewRedisTokenStore creates a new Redis-based token store
func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{
		client: client,
	}
}

// tokenKeys generates Redis keys for token storage
func (r *RedisTokenStore) tokenKeys() string {
	return "edduhub:tokens"
}

// refreshTokenKey generates the key for a refresh token
func (r *RedisTokenStore) refreshTokenKey(tokenID string) string {
	return fmt.Sprintf("edduhub:refresh:%s", tokenID)
}

// revokedTokenKey generates the key for a revoked token
func (r *RedisTokenStore) revokedTokenKey(jti string) string {
	return fmt.Sprintf("edduhub:revoked:%s", jti)
}

// StoreRefreshToken stores a refresh token with metadata
func (r *RedisTokenStore) StoreRefreshToken(ctx context.Context, tokenID string, userID int, expiresAt time.Duration) error {
	data := RefreshTokenData{
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiresAt),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	err = r.client.Set(ctx, r.refreshTokenKey(tokenID), jsonData, expiresAt).Err()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves refresh token metadata
func (r *RedisTokenStore) GetRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenData, error) {
	data, err := r.client.Get(ctx, r.refreshTokenKey(tokenID)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var tokenData RefreshTokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	return &tokenData, nil
}

// DeleteRefreshToken deletes a refresh token (revocation)
func (r *RedisTokenStore) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	err := r.client.Del(ctx, r.refreshTokenKey(tokenID)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// IsTokenRevoked checks if a token has been revoked
func (r *RedisTokenStore) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	result, err := r.client.Exists(ctx, r.revokedTokenKey(jti)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}
	return result > 0, nil
}

// RevokeToken marks a token as revoked
func (r *RedisTokenStore) RevokeToken(ctx context.Context, jti string, expiry time.Duration) error {
	// Store with a small TTL to auto-cleanup
	err := r.client.Set(ctx, r.revokedTokenKey(jti), "1", expiry).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}
