package helpers

import (
    "github.com/labstack/echo/v4"
)

type SuccessResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    Data    any    `json:"data"`
}

func Success(c echo.Context, data any, status int) error {
    return c.JSON(status, SuccessResponse{
        Success: true,
        Data:    data,
    })
}
