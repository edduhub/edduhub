package config

import (
	"os"
	"testing"

	"eduhub/server/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- AppConfig.Validate ---

func TestAppConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &AppConfig{Port: "8080", LogLevel: "info"}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty port fails", func(t *testing.T) {
		cfg := &AppConfig{Port: "", LogLevel: "info"}
		require.Error(t, cfg.Validate())
		assert.Contains(t, cfg.Validate().Error(), "Port")
	})

	t.Run("empty log level fails", func(t *testing.T) {
		cfg := &AppConfig{Port: "8080", LogLevel: ""}
		require.Error(t, cfg.Validate())
		assert.Contains(t, cfg.Validate().Error(), "LogLevel")
	})
}

// --- LoadAppConfig ---

func TestLoadAppConfig(t *testing.T) {
	t.Run("defaults when no env vars set", func(t *testing.T) {
		os.Clearenv()
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.Port)
		assert.False(t, cfg.Debug)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, []string{"http://localhost:3000"}, cfg.CORSOrigins)
	})

	t.Run("custom port", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_PORT", "3000")
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.Equal(t, "3000", cfg.Port)
	})

	t.Run("invalid port returns error", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_PORT", "not-a-number")
		_, err := LoadAppConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APP_PORT")
	})

	t.Run("port out of range", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_PORT", "99999")
		_, err := LoadAppConfig()
		require.Error(t, err)
	})

	t.Run("port zero", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_PORT", "0")
		_, err := LoadAppConfig()
		require.Error(t, err)
	})

	t.Run("debug mode true", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_DEBUG", "true")
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.True(t, cfg.Debug)
	})

	t.Run("debug mode false by default", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_DEBUG", "false")
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.False(t, cfg.Debug)
	})

	t.Run("valid log levels", func(t *testing.T) {
		for _, level := range []string{"debug", "info", "warn", "error"} {
			os.Clearenv()
			os.Setenv("APP_LOG_LEVEL", level)
			cfg, err := LoadAppConfig()
			require.NoError(t, err, "level %s should be valid", level)
			assert.Equal(t, level, cfg.LogLevel)
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_LOG_LEVEL", "trace")
		_, err := LoadAppConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APP_LOG_LEVEL")
	})

	t.Run("custom CORS origins", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CORS_ORIGINS", "https://example.com, https://app.example.com")
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, cfg.CORSOrigins)
	})

	t.Run("CORS origins trims whitespace", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CORS_ORIGINS", "  https://a.com , https://b.com  ")
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.Equal(t, []string{"https://a.com", "https://b.com"}, cfg.CORSOrigins)
	})

	t.Run("CORS origins with empty values after split", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CORS_ORIGINS", ",,")
		_, err := LoadAppConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no valid origins")
	})

	t.Run("production requires razorpay secrets", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_ENV", "production")
		os.Setenv("CORS_ORIGINS", "https://example.com")
		_, err := LoadAppConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "RAZORPAY")
	})

	t.Run("production rejects localhost CORS", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_ENV", "production")
		os.Setenv("CORS_ORIGINS", "http://localhost:3000")
		os.Setenv("RAZORPAY_KEY_ID", "key")
		os.Setenv("RAZORPAY_KEY_SECRET", "secret")
		os.Setenv("RAZORPAY_WEBHOOK_SECRET", "webhook")
		_, err := LoadAppConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "localhost")
	})

	t.Run("production with valid config", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("APP_ENV", "production")
		os.Setenv("CORS_ORIGINS", "https://example.com")
		os.Setenv("RAZORPAY_KEY_ID", "key")
		os.Setenv("RAZORPAY_KEY_SECRET", "secret")
		os.Setenv("RAZORPAY_WEBHOOK_SECRET", "webhook")
		cfg, err := LoadAppConfig()
		require.NoError(t, err)
		assert.Equal(t, "key", cfg.RazorpayKey)
	})
}

// --- DBConfig.Validate ---

func TestDBConfig_Validate(t *testing.T) {
	valid := DBConfig{
		Host: "localhost", Port: "5432", User: "user",
		Password: "pass", DBName: "db", SSLMode: "disable",
	}

	t.Run("valid config", func(t *testing.T) {
		assert.NoError(t, valid.Validate())
	})

	tests := []struct {
		name  string
		field string
	}{
		{"empty host", "Host"},
		{"empty port", "Port"},
		{"empty user", "User"},
		{"empty password", "Password"},
		{"empty dbname", "DBName"},
		{"empty sslmode", "SSLMode"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			switch tc.field {
			case "Host":
				cfg.Host = ""
			case "Port":
				cfg.Port = ""
			case "User":
				cfg.User = ""
			case "Password":
				cfg.Password = ""
			case "DBName":
				cfg.DBName = ""
			case "SSLMode":
				cfg.SSLMode = ""
			}
			require.Error(t, cfg.Validate())
			assert.Contains(t, cfg.Validate().Error(), tc.field)
		})
	}
}

// --- AuthConfig.Validate ---

func TestAuthConfig_Validate(t *testing.T) {
	valid := AuthConfig{
		PublicURL: "http://localhost", AdminURL: "http://localhost",
		Domain: "example.com", Port: "8080",
	}

	t.Run("valid config", func(t *testing.T) {
		assert.NoError(t, valid.Validate())
	})

	tests := []struct {
		name  string
		field string
	}{
		{"empty public url", "PublicURL"},
		{"empty admin url", "AdminURL"},
		{"empty domain", "Domain"},
		{"empty port", "Port"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			switch tc.field {
			case "PublicURL":
				cfg.PublicURL = ""
			case "AdminURL":
				cfg.AdminURL = ""
			case "Domain":
				cfg.Domain = ""
			case "Port":
				cfg.Port = ""
			}
			require.Error(t, cfg.Validate())
			assert.Contains(t, cfg.Validate().Error(), tc.field)
		})
	}
}

// --- RedisConfig.Validate ---

func TestRedisConfig_Validate(t *testing.T) {
	t.Run("disabled config always valid", func(t *testing.T) {
		cfg := &RedisConfig{Enabled: false}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid enabled config", func(t *testing.T) {
		cfg := &RedisConfig{Enabled: true, Host: "localhost", Port: "6379", PoolSize: 10, MinIdleConns: 2}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty host fails", func(t *testing.T) {
		cfg := &RedisConfig{Enabled: true, Host: "", Port: "6379", PoolSize: 10}
		require.Error(t, cfg.Validate())
	})

	t.Run("empty port fails", func(t *testing.T) {
		cfg := &RedisConfig{Enabled: true, Host: "localhost", Port: "", PoolSize: 10}
		require.Error(t, cfg.Validate())
	})

	t.Run("zero pool size fails", func(t *testing.T) {
		cfg := &RedisConfig{Enabled: true, Host: "localhost", Port: "6379", PoolSize: 0}
		require.Error(t, cfg.Validate())
	})

	t.Run("negative min idle conns fails", func(t *testing.T) {
		cfg := &RedisConfig{Enabled: true, Host: "localhost", Port: "6379", PoolSize: 10, MinIdleConns: -1}
		require.Error(t, cfg.Validate())
	})
}

// --- LoadRedisConfig ---

func TestLoadRedisConfig(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		os.Clearenv()
		cfg, err := LoadRedisConfig()
		require.NoError(t, err)
		assert.False(t, cfg.Enabled)
	})

	t.Run("disabled explicitly", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_ENABLED", "false")
		cfg, err := LoadRedisConfig()
		require.NoError(t, err)
		assert.False(t, cfg.Enabled)
	})

	t.Run("enabled with defaults", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_ENABLED", "true")
		cfg, err := LoadRedisConfig()
		require.NoError(t, err)
		assert.True(t, cfg.Enabled)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, "6379", cfg.Port)
		assert.Equal(t, 0, cfg.DB)
		assert.Equal(t, "eduhub:", cfg.Prefix)
		assert.Equal(t, 10, cfg.PoolSize)
		assert.Equal(t, 2, cfg.MinIdleConns)
		assert.Equal(t, 3, cfg.MaxRetries)
	})

	t.Run("custom values", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_ENABLED", "true")
		os.Setenv("REDIS_HOST", "redis.example.com")
		os.Setenv("REDIS_PORT", "6380")
		os.Setenv("REDIS_DB", "2")
		os.Setenv("REDIS_PREFIX", "app:")
		os.Setenv("REDIS_POOL_SIZE", "20")
		os.Setenv("REDIS_MIN_IDLE_CONNS", "5")
		cfg, err := LoadRedisConfig()
		require.NoError(t, err)
		assert.Equal(t, "redis.example.com", cfg.Host)
		assert.Equal(t, "6380", cfg.Port)
		assert.Equal(t, 2, cfg.DB)
		assert.Equal(t, "app:", cfg.Prefix)
		assert.Equal(t, 20, cfg.PoolSize)
		assert.Equal(t, 5, cfg.MinIdleConns)
	})

	t.Run("invalid DB number", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_ENABLED", "true")
		os.Setenv("REDIS_DB", "not-a-number")
		_, err := LoadRedisConfig()
		require.Error(t, err)
	})

	t.Run("invalid pool size", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_ENABLED", "true")
		os.Setenv("REDIS_POOL_SIZE", "abc")
		_, err := LoadRedisConfig()
		require.Error(t, err)
	})

	t.Run("invalid min idle conns", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("REDIS_ENABLED", "true")
		os.Setenv("REDIS_MIN_IDLE_CONNS", "xyz")
		_, err := LoadRedisConfig()
		require.Error(t, err)
	})
}

// --- RedisConfig.ToRedisCacheConfig ---

func TestRedisConfig_ToRedisCacheConfig(t *testing.T) {
	cfg := &RedisConfig{
		Host: "redis.local", Port: "6379", Password: "secret",
		DB: 1, Prefix: "test:", PoolSize: 15, MinIdleConns: 3, MaxRetries: 5,
	}

	cacheConfig := cfg.ToRedisCacheConfig()
	assert.Equal(t, "redis.local", cacheConfig.Host)
	assert.Equal(t, "6379", cacheConfig.Port)
	assert.Equal(t, "secret", cacheConfig.Password)
	assert.Equal(t, 1, cacheConfig.DB)
	assert.Equal(t, "test:", cacheConfig.Prefix)
	assert.Equal(t, 15, cacheConfig.PoolSize)
	assert.Equal(t, 3, cacheConfig.MinIdleConns)
	assert.Equal(t, 5, cacheConfig.MaxRetries)
}

// --- StorageConfig.Validate ---

func TestStorageConfig_Validate(t *testing.T) {
	t.Run("valid minimal config", func(t *testing.T) {
		cfg := &StorageConfig{Endpoint: "localhost:9000", Bucket: "eduhub", PresignedURLExpirySeconds: 3600}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid with both keys", func(t *testing.T) {
		cfg := &StorageConfig{
			Endpoint: "s3.amazonaws.com", Bucket: "eduhub",
			AccessKey: "key", SecretKey: "secret", PresignedURLExpirySeconds: 3600,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty endpoint fails", func(t *testing.T) {
		cfg := &StorageConfig{Endpoint: "", Bucket: "eduhub", PresignedURLExpirySeconds: 3600}
		require.Error(t, cfg.Validate())
	})

	t.Run("empty bucket fails", func(t *testing.T) {
		cfg := &StorageConfig{Endpoint: "localhost:9000", Bucket: "", PresignedURLExpirySeconds: 3600}
		require.Error(t, cfg.Validate())
	})

	t.Run("access key without secret fails", func(t *testing.T) {
		cfg := &StorageConfig{
			Endpoint: "localhost:9000", Bucket: "eduhub",
			AccessKey: "key", SecretKey: "", PresignedURLExpirySeconds: 3600,
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("secret key without access fails", func(t *testing.T) {
		cfg := &StorageConfig{
			Endpoint: "localhost:9000", Bucket: "eduhub",
			AccessKey: "", SecretKey: "secret", PresignedURLExpirySeconds: 3600,
		}
		require.Error(t, cfg.Validate())
	})

	t.Run("zero expiry fails", func(t *testing.T) {
		cfg := &StorageConfig{Endpoint: "localhost:9000", Bucket: "eduhub", PresignedURLExpirySeconds: 0}
		require.Error(t, cfg.Validate())
	})
}

// --- LoadStorageConfig ---

func TestLoadStorageConfig(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		os.Clearenv()
		cfg, err := LoadStorageConfig()
		require.NoError(t, err)
		assert.Equal(t, "localhost:9000", cfg.Endpoint)
		assert.Equal(t, "eduhub", cfg.Bucket)
		assert.Equal(t, "us-east-1", cfg.Region)
		assert.False(t, cfg.UseSSL)
		assert.Equal(t, int64(3600), cfg.PresignedURLExpirySeconds)
	})

	t.Run("custom values", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("STORAGE_ENDPOINT", "s3.example.com")
		os.Setenv("STORAGE_BUCKET", "mybucket")
		os.Setenv("STORAGE_USE_SSL", "true")
		os.Setenv("STORAGE_REGION", "eu-west-1")
		os.Setenv("STORAGE_PRESIGNED_URL_EXPIRY", "7200")
		cfg, err := LoadStorageConfig()
		require.NoError(t, err)
		assert.Equal(t, "s3.example.com", cfg.Endpoint)
		assert.Equal(t, "mybucket", cfg.Bucket)
		assert.True(t, cfg.UseSSL)
		assert.Equal(t, "eu-west-1", cfg.Region)
		assert.Equal(t, int64(7200), cfg.PresignedURLExpirySeconds)
	})

	t.Run("invalid expiry", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("STORAGE_PRESIGNED_URL_EXPIRY", "not-a-number")
		_, err := LoadStorageConfig()
		require.Error(t, err)
	})
}

// --- EmailConfig.Validate ---

func TestEmailConfig_Validate(t *testing.T) {
	valid := EmailConfig{
		Host: "smtp.gmail.com", Port: "587",
		Username: "user", Password: "pass", FromAddress: "test@example.com",
	}

	t.Run("valid config", func(t *testing.T) {
		assert.NoError(t, valid.Validate())
	})

	tests := []struct {
		name  string
		field string
	}{
		{"empty host", "Host"},
		{"empty port", "Port"},
		{"empty username", "Username"},
		{"empty password", "Password"},
		{"empty from address", "FromAddress"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			switch tc.field {
			case "Host":
				cfg.Host = ""
			case "Port":
				cfg.Port = ""
			case "Username":
				cfg.Username = ""
			case "Password":
				cfg.Password = ""
			case "FromAddress":
				cfg.FromAddress = ""
			}
			require.Error(t, cfg.Validate())
		})
	}
}

// --- LoadEmailConfig ---

func TestLoadEmailConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMTP_HOST", "smtp.gmail.com")
		os.Setenv("SMTP_USERNAME", "user")
		os.Setenv("SMTP_PASSWORD", "pass")
		os.Setenv("EMAIL_FROM", "test@example.com")
		cfg, err := LoadEmailConfig()
		require.NoError(t, err)
		assert.Equal(t, "smtp.gmail.com", cfg.Host)
		assert.Equal(t, "587", cfg.Port) // default
		assert.True(t, cfg.EnableStartTLS)
	})

	t.Run("custom port and TLS", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMTP_HOST", "smtp.example.com")
		os.Setenv("SMTP_PORT", "465")
		os.Setenv("SMTP_USERNAME", "user")
		os.Setenv("SMTP_PASSWORD", "pass")
		os.Setenv("EMAIL_FROM", "noreply@example.com")
		os.Setenv("SMTP_STARTTLS", "false")
		cfg, err := LoadEmailConfig()
		require.NoError(t, err)
		assert.Equal(t, "465", cfg.Port)
		assert.False(t, cfg.EnableStartTLS)
	})

	t.Run("missing host fails", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SMTP_USERNAME", "user")
		os.Setenv("SMTP_PASSWORD", "pass")
		os.Setenv("EMAIL_FROM", "test@example.com")
		_, err := LoadEmailConfig()
		require.Error(t, err)
	})
}

// --- LoadAnalyticsConfig ---

func TestLoadAnalyticsConfig(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		os.Clearenv()
		cfg := LoadAnalyticsConfig()
		assert.Equal(t, 0.45, cfg.RiskWeightGradeVeryLow)
		assert.Equal(t, 0.30, cfg.RiskWeightGradeLow)
		assert.Equal(t, 0.15, cfg.RiskWeightGradeMedium)
		assert.Equal(t, 0.75, cfg.RiskLevelHighThreshold)
		assert.Equal(t, 0.45, cfg.RiskLevelLowThreshold)
		assert.Equal(t, 0.05, cfg.RiskMinScore)
		assert.Equal(t, 0.99, cfg.RiskMaxScore)
	})

	t.Run("custom values from env", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ANALYTICS_RISK_GRADE_VERY_LOW_WEIGHT", "0.50")
		os.Setenv("ANALYTICS_RISK_HIGH_THRESHOLD", "0.80")
		cfg := LoadAnalyticsConfig()
		assert.Equal(t, 0.50, cfg.RiskWeightGradeVeryLow)
		assert.Equal(t, 0.80, cfg.RiskLevelHighThreshold)
	})

	t.Run("invalid float uses default", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("ANALYTICS_RISK_GRADE_VERY_LOW_WEIGHT", "not-a-float")
		cfg := LoadAnalyticsConfig()
		assert.Equal(t, 0.45, cfg.RiskWeightGradeVeryLow)
	})
}

// --- getEnvOrDefault ---

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_KEY", "value")
		assert.Equal(t, "value", getEnvOrDefault("TEST_KEY", "default"))
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Clearenv()
		assert.Equal(t, "default", getEnvOrDefault("MISSING_KEY", "default"))
	})
}

// --- getEnvFloat ---

func TestGetEnvFloat(t *testing.T) {
	t.Run("returns parsed float", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("FLOAT_KEY", "3.14")
		assert.Equal(t, 3.14, getEnvFloat("FLOAT_KEY", 0.0))
	})

	t.Run("returns default for missing", func(t *testing.T) {
		os.Clearenv()
		assert.Equal(t, 1.5, getEnvFloat("MISSING", 1.5))
	})

	t.Run("returns default for invalid", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("BAD_FLOAT", "abc")
		assert.Equal(t, 2.5, getEnvFloat("BAD_FLOAT", 2.5))
	})
}

// --- buildDSN ---

func TestBuildDSN_Extended(t *testing.T) {
	t.Run("basic DSN", func(t *testing.T) {
		cfg := DBConfig{
			Host: "localhost", Port: "5432", User: "admin",
			Password: "pass", DBName: "mydb", SSLMode: "disable",
		}
		dsn := buildDSN(cfg)
		assert.Equal(t, "postgres://admin:pass@localhost:5432/mydb?sslmode=disable", dsn)
	})

	t.Run("DSN with SSL certs", func(t *testing.T) {
		cfg := DBConfig{
			Host: "db.example.com", Port: "5432", User: "user",
			Password: "pass", DBName: "prod", SSLMode: "verify-full",
			SSLRootCert: "/certs/root.crt", SSLCert: "/certs/client.crt", SSLKey: "/certs/client.key",
		}
		dsn := buildDSN(cfg)
		assert.Contains(t, dsn, "sslmode=verify-full")
		assert.Contains(t, dsn, "sslrootcert=/certs/root.crt")
		assert.Contains(t, dsn, "sslcert=/certs/client.crt")
		assert.Contains(t, dsn, "sslkey=/certs/client.key")
	})

	t.Run("DSN with partial SSL", func(t *testing.T) {
		cfg := DBConfig{
			Host: "localhost", Port: "5432", User: "user",
			Password: "pass", DBName: "db", SSLMode: "require",
			SSLRootCert: "/certs/root.crt",
		}
		dsn := buildDSN(cfg)
		assert.Contains(t, dsn, "sslrootcert=/certs/root.crt")
		assert.NotContains(t, dsn, "sslcert=")
		assert.NotContains(t, dsn, "sslkey=")
	})
}

// --- LoadDatabaseConfig ---

func TestLoadDatabaseConfig_Extended(t *testing.T) {
	t.Run("production rejects disabled SSL", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "pass")
		os.Setenv("DB_NAME", "db")
		os.Setenv("DB_SSLMODE", "disable")
		os.Setenv("APP_ENV", "production")
		_, err := LoadDatabaseConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SECURITY ERROR")
	})

	t.Run("production allows require SSL", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "pass")
		os.Setenv("DB_NAME", "db")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("APP_ENV", "production")
		cfg, err := LoadDatabaseConfig()
		require.NoError(t, err)
		assert.Equal(t, "require", cfg.SSLMode)
	})

	t.Run("loads SSL cert paths", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "pass")
		os.Setenv("DB_NAME", "db")
		os.Setenv("DB_SSL_ROOT_CERT", "/certs/root.crt")
		os.Setenv("DB_SSL_CERT", "/certs/client.crt")
		os.Setenv("DB_SSL_KEY", "/certs/client.key")
		cfg, err := LoadDatabaseConfig()
		require.NoError(t, err)
		assert.Equal(t, "/certs/root.crt", cfg.SSLRootCert)
		assert.Equal(t, "/certs/client.crt", cfg.SSLCert)
		assert.Equal(t, "/certs/client.key", cfg.SSLKey)
	})
}

// --- Config.Validate extended ---

func TestConfig_Validate_Extended(t *testing.T) {
	t.Run("nil DB", func(t *testing.T) {
		cfg := &Config{DBConfig: &DBConfig{}, AuthConfig: &AuthConfig{}, AppConfig: &AppConfig{}}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DB cannot be nil")
	})

	t.Run("nil DBConfig", func(t *testing.T) {
		cfg := &Config{DB: newDummyDB(), AuthConfig: &AuthConfig{}, AppConfig: &AppConfig{}}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DBConfig cannot be nil")
	})

	t.Run("nil AuthConfig", func(t *testing.T) {
		cfg := &Config{DB: newDummyDB(), DBConfig: &DBConfig{}, AppConfig: &AppConfig{}}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AuthConfig cannot be nil")
	})

	t.Run("nil AppConfig", func(t *testing.T) {
		cfg := &Config{DB: newDummyDB(), DBConfig: &DBConfig{}, AuthConfig: &AuthConfig{}}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AppConfig cannot be nil")
	})

	t.Run("validates optional redis config", func(t *testing.T) {
		cfg := &Config{
			DB:         newDummyDB(),
			DBConfig:   &DBConfig{Host: "h", Port: "5432", User: "u", Password: "p", DBName: "d", SSLMode: "disable"},
			AuthConfig: &AuthConfig{PublicURL: "u", AdminURL: "u", Domain: "d", Port: "8080"},
			AppConfig:  &AppConfig{Port: "8080", LogLevel: "info"},
			RedisConfig: &RedisConfig{Enabled: true, Host: "", Port: "6379", PoolSize: 10},
		}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "RedisConfig")
	})
}

// dummyDB creates a non-nil *repository.DB for testing Config.Validate
func newDummyDB() *repository.DB {
	return &repository.DB{}
}
