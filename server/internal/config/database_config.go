package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

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
	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		// It's generally better to return an error from LoadDatabase
		// and handle panics at a higher level (e.g., main), but matching
		// the original panic behavior.
		panic(fmt.Errorf("failed to load database config: %w", err))
	}

	dsn := buildDSN(*dbConfig)

	// Use a context with timeout for connection attempts in production
	// For this example, using context.Background() as in original
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		// Same note about panic vs return error applies
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	// ping the database to ensure the connection is healthy
	err = pool.Ping(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	// --- Complete the return statement ---
	return &repository.DB{
		Pool: pool, // Assign the connected pool to the Pool field
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
