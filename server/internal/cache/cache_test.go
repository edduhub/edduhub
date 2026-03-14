package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRedisConfig(t *testing.T) {
	cfg := DefaultRedisConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "6379", cfg.Port)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, "eduhub:", cfg.Prefix)
	assert.Equal(t, 10, cfg.PoolSize)
	assert.Equal(t, 2, cfg.MinIdleConns)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout)
	assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 4*time.Second, cfg.PoolTimeout)
}

func TestCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		components []any
		expected   string
	}{
		{
			name:       "single string",
			components: []any{"user"},
			expected:   `["user"]`,
		},
		{
			name:       "single int",
			components: []any{42},
			expected:   `[42]`,
		},
		{
			name:       "multiple strings",
			components: []any{"user", "profile"},
			expected:   `["user","profile"]`,
		},
		{
			name:       "mixed types",
			components: []any{"course", 1, "lecture", 5},
			expected:   `["course",1,"lecture",5]`,
		},
		{
			name:       "empty components",
			components: []any{},
			expected:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CacheKey(tt.components...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildStudentKey(t *testing.T) {
	key := BuildStudentKey(1, 42)
	assert.Equal(t, "student:1:42", key)
}

func TestBuildCourseKey(t *testing.T) {
	key := BuildCourseKey(3, 10)
	assert.Equal(t, "course:3:10", key)
}

func TestBuildCourseListKey(t *testing.T) {
	key := BuildCourseListKey(1, 2, 25)
	assert.Equal(t, "course:list:1:p2:l25", key)
}

func TestBuildAttendanceKey(t *testing.T) {
	key := BuildAttendanceKey(1, 42, 10)
	assert.Equal(t, "attendance:1:42:10", key)
}

func TestBuildCalendarKey(t *testing.T) {
	key := BuildCalendarKey(1, "2026-01-01", "2026-01-31")
	assert.Equal(t, "calendar:1:2026-01-01:2026-01-31", key)
}

func TestBuildSessionKey(t *testing.T) {
	key := BuildSessionKey("abc-123")
	assert.Equal(t, "session:abc-123", key)
}

func TestRedisCache_buildKey(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		c := &RedisCache{prefix: "test:"}
		assert.Equal(t, "test:mykey", c.buildKey("mykey"))
	})

	t.Run("without prefix", func(t *testing.T) {
		c := &RedisCache{prefix: ""}
		assert.Equal(t, "mykey", c.buildKey("mykey"))
	})

	t.Run("with nested key", func(t *testing.T) {
		c := &RedisCache{prefix: "app:"}
		assert.Equal(t, "app:user:42:profile", c.buildKey("user:42:profile"))
	})
}
