package config

import (
	"fmt"
	"os"
)

type AuthConfig struct {
	PublicURL         string
	AdminURL          string
	Domain            string
	Port              string
	College           CollegeConfig
	HydraPublicURL    string
	HydraAdminURL     string
	HydraClientID     string
	HydraClientSecret string
	KetoReadURL       string
	KetoWriteURL      string
}

type CollegeConfig struct {
	RequireVerification bool
	AllowedRoles        []string
}

func LoadAuthConfig() (*AuthConfig, error) {
	config := &AuthConfig{
		PublicURL:         os.Getenv("KRATOS_PUBLIC_URL"),
		AdminURL:          os.Getenv("KRATOS_ADMIN_URL"),
		Domain:            os.Getenv("KRATOS_DOMAIN"),
		Port:              os.Getenv("PORT"),
		HydraPublicURL:    os.Getenv("HYDRA_PUBLIC_URL"),
		HydraAdminURL:     os.Getenv("HYDRA_ADMIN_URL"),
		HydraClientID:     os.Getenv("HYDRA_CLIENT_ID"),
		HydraClientSecret: os.Getenv("HYDRA_CLIENT_SECRET"),
		KetoReadURL:       os.Getenv("KETO_READ_URL"),
		KetoWriteURL:      os.Getenv("KETO_WRITE_URL"),
		College: CollegeConfig{
			RequireVerification: true,
			AllowedRoles:        []string{"admin", "faculty", "student"},
		},
	}

	// Validate required fields
	if config.PublicURL == "" || config.AdminURL == "" {
		return nil, fmt.Errorf("missing required Kratos configuration")
	}

	// Ensure full validation is applied
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate is a method on AuthConfig for validation.
func (c *AuthConfig) Validate() error {
	if c.PublicURL == "" {
		return fmt.Errorf("AuthConfig.PublicURL cannot be empty")
	}
	if c.AdminURL == "" {
		return fmt.Errorf("AuthConfig.AdminURL cannot be empty")
	}
	if c.Domain == "" {
		return fmt.Errorf("AuthConfig.Domain cannot be empty")
	}
	if c.Port == "" {
		return fmt.Errorf("AuthConfig.Port cannot be empty")
	}
	return nil
}
