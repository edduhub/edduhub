package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"DB_HOST":           "localhost",
				"DB_PORT":           "5432",
				"DB_USER":           "testuser",
				"DB_PASSWORD":       "testpass",
				"DB_NAME":           "testdb",
				"KRATOS_PUBLIC_URL": "http://public.example.com",
				"KRATOS_ADMIN_URL":  "http://admin.example.com",
				"KRATOS_DOMAIN":     "example.com",
				"PORT":              "3000",
				"APP_PORT":          "3000",
			},
			expectError: false,
		},
		{
			name: "missing database config",
			envVars: map[string]string{
				"KRATOS_PUBLIC_URL": "http://public.example.com",
				"KRATOS_ADMIN_URL":  "http://admin.example.com",
				"APP_PORT":          "3000",
			},
			expectError: true,
		},
		{
			name: "missing auth config",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"APP_PORT":    "3000",
			},
			expectError: true,
		},
		{
			name: "missing app port",
			envVars: map[string]string{
				"DB_HOST":           "localhost",
				"DB_PORT":           "5432",
				"DB_USER":           "testuser",
				"DB_PASSWORD":       "testpass",
				"DB_NAME":           "testdb",
				"KRATOS_PUBLIC_URL": "http://public.example.com",
				"KRATOS_ADMIN_URL":  "http://admin.example.com",
				"KRATOS_DOMAIN":     "example.com",
				"PORT":              "3000",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			os.Setenv("DB_SKIP_CONNECT", "1")

			config, err := LoadConfig()

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && config != nil {
				if len(tt.envVars["APP_PORT"]) > 0 && config.AppPort != tt.envVars["APP_PORT"] {
					t.Errorf("expected AppPort %s, got %s", tt.envVars["APP_PORT"], config.AppPort)
				}
				if config.DBConfig == nil {
					t.Error("expected DBConfig to not be nil")
				}
				if config.AuthConfig == nil {
					t.Error("expected AuthConfig to not be nil")
				}
				if config.DB == nil {
					t.Error("expected DB to not be nil")
				}
			}
		})
	}
}
