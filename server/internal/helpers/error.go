package helpers

import (
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
    Success bool `json:"success"`
    Error   any  `json:"error"`
}

func Error(c echo.Context, error any, status int) error {
    return c.JSON(status, ErrorResponse{
        Success: false,
        Error:   error,
    })
}
