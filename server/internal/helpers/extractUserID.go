package helpers

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// ExtractUserID extracts the user ID from the Echo context
// User ID is typically set by the authentication middleware
func ExtractUserID(c echo.Context) (int, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return 0, fmt.Errorf("user ID not found in context")
	}

	id, ok := userID.(int)
	if !ok {
		return 0, fmt.Errorf("user ID is not an integer")
	}

	return id, nil
}
