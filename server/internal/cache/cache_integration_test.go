//go:build integration

package cache

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

func newTestCache(t *testing.T) *RedisCache {
	t.Helper()
	cfg := DefaultRedisConfig()
	cfg.Prefix = "test:"

	cache, err := NewRedisCache(cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		_ = cache.Clear(ctx)
		_ = cache.Close()
	})

	return cache
}

func TestRedisCache_Integration_SetAndGet(t *testing.T) {
	requireRedisAvailable(t)
	c := newTestCache(t)
	ctx := context.Background()

	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	input := testData{Name: "hello", Value: 42}
	err := c.Set(ctx, "setget:key1", input, 10*time.Second)
	require.NoError(t, err)

	var got testData
	err = c.Get(ctx, "setget:key1", &got)
	require.NoError(t, err)
	assert.Equal(t, input, got)
}

func TestRedisCache_Integration_Delete(t *testing.T) {
	requireRedisAvailable(t)
	c := newTestCache(t)
	ctx := context.Background()

	err := c.Set(ctx, "del:key1", "some-value", 10*time.Second)
	require.NoError(t, err)

	err = c.Delete(ctx, "del:key1")
	require.NoError(t, err)

	var got string
	err = c.Get(ctx, "del:key1", &got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestRedisCache_Integration_GetNotFound(t *testing.T) {
	requireRedisAvailable(t)
	c := newTestCache(t)
	ctx := context.Background()

	var got string
	err := c.Get(ctx, "nonexistent:key", &got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestRedisCache_Integration_Clear(t *testing.T) {
	requireRedisAvailable(t)
	c := newTestCache(t)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "clear:a", "val-a", 10*time.Second))
	require.NoError(t, c.Set(ctx, "clear:b", "val-b", 10*time.Second))
	require.NoError(t, c.Set(ctx, "clear:c", "val-c", 10*time.Second))

	err := c.Clear(ctx)
	require.NoError(t, err)

	var got string
	for _, key := range []string{"clear:a", "clear:b", "clear:c"} {
		err = c.Get(ctx, key, &got)
		assert.Error(t, err, "expected key %s to be cleared", key)
	}
}

func TestRedisCache_Integration_GetOrSet(t *testing.T) {
	requireRedisAvailable(t)
	c := newTestCache(t)
	ctx := context.Background()

	callCount := 0
	fn := func() (any, error) {
		callCount++
		return map[string]string{"status": "computed"}, nil
	}

	// First call: cache miss, function should be invoked
	result, err := c.GetOrSet(ctx, "getorset:key1", 10*time.Second, fn)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	resultMap, ok := result.(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "computed", resultMap["status"])

	// Wait briefly for the background goroutine to write to cache
	time.Sleep(100 * time.Millisecond)

	// Second call: cache hit, function should NOT be invoked
	_, err = c.GetOrSet(ctx, "getorset:key1", 10*time.Second, fn)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRedisCache_Integration_Ping(t *testing.T) {
	requireRedisAvailable(t)
	c := newTestCache(t)
	ctx := context.Background()

	err := c.Ping(ctx)
	assert.NoError(t, err)
}

func TestRedisCache_Integration_Close(t *testing.T) {
	requireRedisAvailable(t)
	cfg := DefaultRedisConfig()
	cfg.Prefix = "test:close:"

	c, err := NewRedisCache(cfg)
	require.NoError(t, err)

	err = c.Close()
	assert.NoError(t, err)

	// After close, operations should fail
	ctx := context.Background()
	err = c.Ping(ctx)
	assert.Error(t, err)
}
