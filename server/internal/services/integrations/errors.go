package integrations

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrIntegrationDisabled      = errors.New("integration disabled")
	ErrIntegrationMisconfigured = errors.New("integration misconfigured")
)

type ConfigError struct {
	Integration string
	Reason      error
	Missing     []string
}

func (e *ConfigError) Error() string {
	name := strings.TrimSpace(e.Integration)
	if name == "" {
		name = "integration"
	}
	if len(e.Missing) > 0 {
		return fmt.Sprintf("%s %v: missing %s", name, e.Reason, strings.Join(e.Missing, ", "))
	}
	return fmt.Sprintf("%s %v", name, e.Reason)
}

func (e *ConfigError) Unwrap() error {
	return e.Reason
}

func NewDisabledError(integration string) error {
	return &ConfigError{Integration: integration, Reason: ErrIntegrationDisabled}
}

func NewMisconfiguredError(integration string, missing ...string) error {
	return &ConfigError{Integration: integration, Reason: ErrIntegrationMisconfigured, Missing: missing}
}

func IsDisabled(err error) bool {
	return errors.Is(err, ErrIntegrationDisabled)
}

func IsMisconfigured(err error) bool {
	return errors.Is(err, ErrIntegrationMisconfigured)
}
