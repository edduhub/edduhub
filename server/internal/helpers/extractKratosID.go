package helpers

import (
	"fmt"

	"eduhub/server/internal/services/auth"
	"github.com/labstack/echo/v4"
)

// GetKratosID extracts the Kratos identity ID from the Echo context
// The identity is set by authentication middleware (JWT middleware)
func GetKratosID(c echo.Context) (string, error) {
	identity, ok := c.Get("identity").(*auth.Identity)
	if !ok || identity == nil {
		return "", fmt.Errorf("identity not found in context")
	}

	if identity.ID == "" {
		return "", fmt.Errorf("kratos ID is empty")
	}

	return identity.ID, nil
}
