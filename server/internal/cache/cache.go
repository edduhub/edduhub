package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface defines cache operations
type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (any, error)) (any, error)
	Ping(ctx context.Context) error
	Close() error
}

// RedisCache implements Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	prefix string // Optional key prefix for namespace isolation
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Prefix   string
	// Connection pool settings
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	// Timeouts
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
}

// DefaultRedisConfig returns a configuration optimized for low-resource environments
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         "localhost",
		Port:         "6379",
		Password:     "",
		DB:           0,
		Prefix:       "eduhub:",
		PoolSize:     10, // Reduced for low-resource environments
		MinIdleConns: 2,  // Minimum idle connections
		MaxRetries:   3,  // Retry failed commands
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	if config == nil {
		config = DefaultRedisConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: config.Prefix,
	}, nil
}

// buildKey adds prefix to the key
func (c *RedisCache) buildKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + key
}

// Set stores a value in Redis with expiration
func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, c.buildKey(key), data, ttl).Err()
}

// Get retrieves a value from Redis
func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := c.client.Get(ctx, c.buildKey(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value from Redis
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.buildKey(key)).Err()
}

// Clear removes all keys with the configured prefix
func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.buildKey("*")
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GetOrSet retrieves from cache or sets if not exists (cache-aside pattern)
func (c *RedisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (any, error)) (any, error) {
	// Try to get from cache first
	var result any
	err := c.Get(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	// Not in cache, execute function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache (fire and forget)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.Set(bgCtx, key, value, ttl)
	}()

	return value, nil
}

// Ping checks if Redis is accessible
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client for advanced operations
func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}

// CacheKey generates a cache key from components
func CacheKey(components ...any) string {
	data, _ := json.Marshal(components)
	return string(data)
}

// Common cache key patterns for EduHub
const (
	// TTL configurations optimized for low-resource environments
	TTLShort  = 5 * time.Minute  // For frequently changing data
	TTLMedium = 30 * time.Minute // For moderately stable data
	TTLLong   = 2 * time.Hour    // For stable data
	TTLDay    = 24 * time.Hour   // For very stable data

	// Cache key prefixes
	PrefixStudent    = "student:"
	PrefixCourse     = "course:"
	PrefixLecture    = "lecture:"
	PrefixAttendance = "attendance:"
	PrefixGrade      = "grade:"
	PrefixCalendar   = "calendar:"
	PrefixDepartment = "department:"
	PrefixCollege    = "college:"
	PrefixUser       = "user:"
	PrefixSession    = "session:"
)

// Helper functions for common cache operations

// BuildStudentKey creates a cache key for student data
func BuildStudentKey(collegeID, studentID int) string {
	return fmt.Sprintf("%s%d:%d", PrefixStudent, collegeID, studentID)
}

// BuildCourseKey creates a cache key for course data
func BuildCourseKey(collegeID, courseID int) string {
	return fmt.Sprintf("%s%d:%d", PrefixCourse, collegeID, courseID)
}

// BuildCourseListKey creates a cache key for course list
func BuildCourseListKey(collegeID int, page, limit int) string {
	return fmt.Sprintf("%slist:%d:p%d:l%d", PrefixCourse, collegeID, page, limit)
}

// BuildAttendanceKey creates a cache key for attendance data
func BuildAttendanceKey(collegeID, studentID, courseID int) string {
	return fmt.Sprintf("%s%d:%d:%d", PrefixAttendance, collegeID, studentID, courseID)
}

// BuildCalendarKey creates a cache key for calendar events
func BuildCalendarKey(collegeID int, startDate, endDate string) string {
	return fmt.Sprintf("%s%d:%s:%s", PrefixCalendar, collegeID, startDate, endDate)
}

// BuildSessionKey creates a cache key for user session
func BuildSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", PrefixSession, sessionID)
}
