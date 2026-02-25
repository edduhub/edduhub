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
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	SSLRootCert string // Path to SSL root certificate for production
	SSLCert     string // Path to SSL client certificate (optional)
	SSLKey      string // Path to SSL client key (optional)
}

// LoadDatabaseConfig loads database configuration from environment variables
// SECURITY: Supports SSL/TLS configuration for production databases
func LoadDatabaseConfig() (*DBConfig, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE") // Often "disable" for local dev, "require"/"verify-full" for prod

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

	// Load SSL certificate paths for production
	sslRootCert := os.Getenv("DB_SSL_ROOT_CERT") // e.g., "/path/to/root.crt"
	sslCert := os.Getenv("DB_SSL_CERT")          // e.g., "/path/to/client.crt"
	sslKey := os.Getenv("DB_SSL_KEY")            // e.g., "/path/to/client.key"

	// SECURITY: Enforce SSL in production - disabled SSL is a security vulnerability
	if os.Getenv("APP_ENV") == "production" && dbSSLMode == "disable" {
		return nil, fmt.Errorf("SECURITY ERROR: Database SSL cannot be disabled in production environment. Set DB_SSL_MODE to 'require' or higher")
	}

	return &DBConfig{
		Host:        dbHost,
		Port:        strconv.Itoa(dbPort),
		User:        dbUser,
		Password:    dbPassword,
		DBName:      dbName,
		SSLMode:     dbSSLMode,
		SSLRootCert: sslRootCert,
		SSLCert:     sslCert,
		SSLKey:      sslKey,
	}, nil
}

// LoadDatabaseWithRetry loads the database with proper error handling instead of panics.
// It supports retry logic and graceful failure.
func LoadDatabaseWithRetry(maxRetries int) (*repository.DB, error) {
	if os.Getenv("DB_SKIP_CONNECT") == "1" {
		return &repository.DB{}, nil
	}

	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	dsn := buildDSN(*dbConfig)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
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

	var pool *pgxpool.Pool
	var lastErr error

	// Retry logic with exponential backoff
	for attempt := range maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			cancel()
			lastErr = fmt.Errorf("attempt %d: failed to connect to database: %w", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
			continue
		}

		// ping the database to ensure the connection is healthy
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = pool.Ping(pingCtx)
		pingCancel()
		cancel()

		if err != nil {
			pool.Close()
			lastErr = fmt.Errorf("attempt %d: failed to ping database: %w", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Success
		return &repository.DB{Pool: pool}, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, lastErr)
}

// LoadDatabase loads the database connection, panics on failure for backward compatibility.
// DEPRECATED: Prefer using LoadDatabaseWithRetry for graceful error handling.
func LoadDatabase() *repository.DB {
	db, err := LoadDatabaseWithRetry(3)
	if err != nil {
		// Log the error before panicking for better debugging
		fmt.Fprintf(os.Stderr, "FATAL: Database connection failed: %v\n", err)
		panic(err)
	}
	return db
}

func buildDSN(config DBConfig) string {
	// Base DSN
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)

	// Add SSL certificate parameters if configured
	if config.SSLRootCert != "" {
		dsn += fmt.Sprintf("&sslrootcert=%s", config.SSLRootCert)
	}
	if config.SSLCert != "" {
		dsn += fmt.Sprintf("&sslcert=%s", config.SSLCert)
	}
	if config.SSLKey != "" {
		dsn += fmt.Sprintf("&sslkey=%s", config.SSLKey)
	}

	return dsn
}

// Validate is a method on DBConfig for validation.
func (c *DBConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("DBConfig.Host cannot be empty")
	}
	if c.Port == "" {
		return fmt.Errorf("DBConfig.Port cannot be empty")
	}
	if c.User == "" {
		return fmt.Errorf("DBConfig.User cannot be empty")
	}
	if c.Password == "" {
		return fmt.Errorf("DBConfig.Password cannot be empty")
	}
	if c.DBName == "" {
		return fmt.Errorf("DBConfig.DBName cannot be empty")
	}
	if c.SSLMode == "" {
		return fmt.Errorf("DBConfig.SSLMode cannot be empty")
	}
	return nil
}
