package helpers

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// ExtractUserID extracts the user ID from the Echo context
// User ID is typically set by the authentication middleware
// Checks multiple keys: user_id, student_id
func ExtractUserID(c echo.Context) (int, error) {
	// Check for user_id first
	userID := c.Get("user_id")
	if userID != nil {
		id, ok := userID.(int)
		if !ok {
			return 0, fmt.Errorf("user ID is not an integer")
		}
		return id, nil
	}

	// Fallback: check for student_id
	studentID := c.Get("student_id")
	if studentID != nil {
		id, ok := studentID.(int)
		if !ok {
			return 0, fmt.Errorf("student ID is not an integer")
		}
		return id, nil
	}

	return 0, fmt.Errorf("user ID not found in context")
}
