package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"eduhub/server/internal/cache"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Enabled      bool
	Host         string
	Port         string
	Password     string
	DB           int
	Prefix       string
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
}

// LoadRedisConfig loads Redis configuration from environment variables
func LoadRedisConfig() (*RedisConfig, error) {
	// Check if Redis is enabled
	enabled := os.Getenv("REDIS_ENABLED")
	if enabled == "" || enabled == "false" {
		return &RedisConfig{Enabled: false}, nil
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // default
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379" // default
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbStr := os.Getenv("REDIS_DB")
	db := 0
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_DB value: %w", err)
		}
	}

	prefix := os.Getenv("REDIS_PREFIX")
	if prefix == "" {
		prefix = "eduhub:" // default
	}

	// Parse pool size
	poolSizeStr := os.Getenv("REDIS_POOL_SIZE")
	poolSize := 10 // default for low-resource environments
	if poolSizeStr != "" {
		var err error
		poolSize, err = strconv.Atoi(poolSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_POOL_SIZE value: %w", err)
		}
	}

	// Parse min idle connections
	minIdleConnsStr := os.Getenv("REDIS_MIN_IDLE_CONNS")
	minIdleConns := 2 // default
	if minIdleConnsStr != "" {
		var err error
		minIdleConns, err = strconv.Atoi(minIdleConnsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS value: %w", err)
		}
	}

	return &RedisConfig{
		Enabled:      true,
		Host:         host,
		Port:         port,
		Password:     password,
		DB:           db,
		Prefix:       prefix,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}, nil
}

// ToRedisCacheConfig converts RedisConfig to cache.RedisConfig
func (c *RedisConfig) ToRedisCacheConfig() *cache.RedisConfig {
	return &cache.RedisConfig{
		Host:         c.Host,
		Port:         c.Port,
		Password:     c.Password,
		DB:           c.DB,
		Prefix:       c.Prefix,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
		MaxRetries:   c.MaxRetries,
		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		PoolTimeout:  c.PoolTimeout,
	}
}

// Validate validates Redis configuration
func (c *RedisConfig) Validate() error {
	if !c.Enabled {
		return nil // No validation needed if disabled
	}

	if c.Host == "" {
		return fmt.Errorf("RedisConfig.Host cannot be empty")
	}
	if c.Port == "" {
		return fmt.Errorf("RedisConfig.Port cannot be empty")
	}
	if c.PoolSize < 1 {
		return fmt.Errorf("RedisConfig.PoolSize must be at least 1")
	}
	if c.MinIdleConns < 0 {
		return fmt.Errorf("RedisConfig.MinIdleConns cannot be negative")
	}
	return nil
}
