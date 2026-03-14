//go:build integration

package auth

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireRedisAvailable(t *testing.T) {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis not available at localhost:6379 for integration test: %v", err)
	}
}

func newTestTokenStore(t *testing.T) *RedisTokenStore {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	t.Cleanup(func() {
		ctx := context.Background()
		// Clean up test keys
		iter := client.Scan(ctx, 0, "edduhub:refresh:test-*", 0).Iterator()
		var keys []string
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		iter = client.Scan(ctx, 0, "edduhub:revoked:test-*", 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if len(keys) > 0 {
			client.Del(ctx, keys...)
		}
		client.Close()
	})

	return NewRedisTokenStore(client)
}

func TestRedisTokenStore_StoreAndGetRefreshToken(t *testing.T) {
	requireRedisAvailable(t)
	store := newTestTokenStore(t)
	ctx := context.Background()

	tokenID := "test-store-get-" + time.Now().Format("150405")
	userID := 42
	ttl := 10 * time.Second

	beforeStore := time.Now()
	err := store.StoreRefreshToken(ctx, tokenID, userID, ttl)
	require.NoError(t, err)

	data, err := store.GetRefreshToken(ctx, tokenID)
	require.NoError(t, err)
	require.NotNil(t, data)

	assert.Equal(t, userID, data.UserID)
	assert.WithinDuration(t, beforeStore, data.CreatedAt, 2*time.Second)
	assert.WithinDuration(t, beforeStore.Add(ttl), data.ExpiresAt, 2*time.Second)
}

func TestRedisTokenStore_GetRefreshToken_NotFound(t *testing.T) {
	requireRedisAvailable(t)
	store := newTestTokenStore(t)
	ctx := context.Background()

	data, err := store.GetRefreshToken(ctx, "test-nonexistent-token")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "refresh token not found")
}

func TestRedisTokenStore_DeleteRefreshToken(t *testing.T) {
	requireRedisAvailable(t)
	store := newTestTokenStore(t)
	ctx := context.Background()

	tokenID := "test-delete-" + time.Now().Format("150405")
	err := store.StoreRefreshToken(ctx, tokenID, 1, 10*time.Second)
	require.NoError(t, err)

	err = store.DeleteRefreshToken(ctx, tokenID)
	require.NoError(t, err)

	data, err := store.GetRefreshToken(ctx, tokenID)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "refresh token not found")
}

func TestRedisTokenStore_RevokeAndCheckToken(t *testing.T) {
	requireRedisAvailable(t)
	store := newTestTokenStore(t)
	ctx := context.Background()

	jti := "test-revoke-" + time.Now().Format("150405")

	err := store.RevokeToken(ctx, jti, 10*time.Second)
	require.NoError(t, err)

	revoked, err := store.IsTokenRevoked(ctx, jti)
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestRedisTokenStore_IsTokenRevoked_NotRevoked(t *testing.T) {
	requireRedisAvailable(t)
	store := newTestTokenStore(t)
	ctx := context.Background()

	revoked, err := store.IsTokenRevoked(ctx, "test-never-revoked-token")
	require.NoError(t, err)
	assert.False(t, revoked)
}
