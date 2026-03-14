package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthConfig(t *testing.T) {
	t.Setenv("KRATOS_PUBLIC_URL", "http://localhost:4433")
	t.Setenv("KRATOS_ADMIN_URL", "http://localhost:4434")
	t.Setenv("KRATOS_DOMAIN", "example.com")
	t.Setenv("PORT", "8080")
	t.Setenv("HYDRA_PUBLIC_URL", "http://localhost:4444")
	t.Setenv("HYDRA_ADMIN_URL", "http://localhost:4445")
	t.Setenv("HYDRA_CLIENT_ID", "client-id")
	t.Setenv("HYDRA_CLIENT_SECRET", "client-secret")
	t.Setenv("KETO_READ_URL", "http://localhost:4467")
	t.Setenv("KETO_WRITE_URL", "http://localhost:4466")

	cfg, err := LoadAuthConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "http://localhost:4433", cfg.PublicURL)
	assert.Equal(t, "http://localhost:4434", cfg.AdminURL)
	assert.Equal(t, "example.com", cfg.Domain)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "http://localhost:4444", cfg.HydraPublicURL)
	assert.Equal(t, "http://localhost:4445", cfg.HydraAdminURL)
	assert.Equal(t, "client-id", cfg.HydraClientID)
	assert.Equal(t, "client-secret", cfg.HydraClientSecret)
	assert.Equal(t, "http://localhost:4467", cfg.KetoReadURL)
	assert.Equal(t, "http://localhost:4466", cfg.KetoWriteURL)
	assert.True(t, cfg.College.RequireVerification)
	assert.Equal(t, []string{"admin", "faculty", "student"}, cfg.College.AllowedRoles)
}

func TestLoadAuthConfig_RequiresKratosURLs(t *testing.T) {
	os.Clearenv()

	cfg, err := LoadAuthConfig()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "missing required Kratos configuration")
}

func TestAuthConfigValidate(t *testing.T) {
	valid := &AuthConfig{
		PublicURL: "http://localhost:4433",
		AdminURL:  "http://localhost:4434",
		Domain:    "example.com",
		Port:      "8080",
	}

	t.Run("valid", func(t *testing.T) {
		assert.NoError(t, valid.Validate())
	})

	t.Run("missing public url", func(t *testing.T) {
		cfg := *valid
		cfg.PublicURL = ""
		require.Error(t, cfg.Validate())
		assert.Contains(t, cfg.Validate().Error(), "PublicURL")
	})

	t.Run("missing admin url", func(t *testing.T) {
		cfg := *valid
		cfg.AdminURL = ""
		require.Error(t, cfg.Validate())
		assert.Contains(t, cfg.Validate().Error(), "AdminURL")
	})

	t.Run("missing domain", func(t *testing.T) {
		cfg := *valid
		cfg.Domain = ""
		require.Error(t, cfg.Validate())
		assert.Contains(t, cfg.Validate().Error(), "Domain")
	})

	t.Run("missing port", func(t *testing.T) {
		cfg := *valid
		cfg.Port = ""
		require.Error(t, cfg.Validate())
		assert.Contains(t, cfg.Validate().Error(), "Port")
	})
}
