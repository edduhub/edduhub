package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"eduhub/server/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoadDatabaseConfig loads database configuration from environment variables
func LoadDatabaseConfig() (*DBConfig, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE") // Often "disable" for local dev, "require" for prod

	if dbHost == "" || dbPortStr == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return nil, fmt.Errorf("database environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME) must be set")
	}

	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT value: %w", err)
	}

	if dbSSLMode == "" {
		dbSSLMode = "disable" // Default SSLMode if not set
	}

	return &DBConfig{
		Host:     dbHost,
		Port:     strconv.Itoa(dbPort),
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  dbSSLMode,
	}, nil
}

func LoadDatabase() *repository.DB {
	if os.Getenv("DB_SKIP_CONNECT") == "1" {
		return &repository.DB{}
	}

	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load database config: %w", err))
	}

	dsn := buildDSN(*dbConfig)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Errorf("unable to parse config: %w", err))
	}

	// OPTIMIZED: Low-resource configuration for better performance on limited hardware
	// Reduce connection overhead and memory usage
	poolConfig.MaxConnIdleTime = 5 * time.Minute    // Reduced from 10 minutes - close idle connections faster
	poolConfig.MaxConnLifetime = 30 * time.Minute   // Reduced from 1 hour - prevent connection buildup
	poolConfig.MinConns = 2                         // Reduced from 4 - lower baseline memory usage
	poolConfig.MaxConns = 20                        // Reduced from 100 - prevent resource exhaustion
	poolConfig.HealthCheckPeriod = 30 * time.Second // Reduced from 1 hour - faster detection of failed connections

	// OPTIMIZED: Connection timeouts for better resource management
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second
	poolConfig.ConnConfig.RuntimeParams = map[string]string{
		// Optimize for faster queries on low-end hardware
		"statement_timeout": "30000", // 30 seconds max per query
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	// ping the database to ensure the connection is healthy
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pingCancel()
	err = pool.Ping(pingCtx)
	if err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	return &repository.DB{
		Pool: pool,
	}
}

func buildDSN(config DBConfig) string {
	// Using fmt.Sprintf is often cleaner for DSN construction
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)
}
