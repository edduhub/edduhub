package middleware

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorSanitizationMiddleware_Development(t *testing.T) {
	m := &ErrorSanitizationMiddleware{IsProduction: false}

	t.Run("passes through errors in development", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/test")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "sql: connection refused at /var/db")
		})

		err := handler(c)
		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, 500, he.Code)
		assert.Contains(t, he.Message, "sql")
	})

	t.Run("returns nil when no error", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		assert.NoError(t, err)
	})
}

func TestErrorSanitizationMiddleware_Production(t *testing.T) {
	m := &ErrorSanitizationMiddleware{IsProduction: true}

	t.Run("sanitizes database errors", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "sql: no rows in result set")
		})

		err := handler(c)
		require.Error(t, err)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "A database error occurred", he.Message)
	})

	t.Run("sanitizes postgres errors", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "postgres: duplicate key value violates unique constraint")
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "A database error occurred", he.Message)
	})

	t.Run("sanitizes file path errors", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "open /var/data/config.yml: permission denied")
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "An error occurred while processing your request", he.Message)
	})

	t.Run("sanitizes stack trace errors", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "goroutine 1 [running]:")
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "An internal error occurred", he.Message)
	})

	t.Run("sanitizes connection errors", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "dial tcp 10.0.0.1:5432: connection refused")
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "A connectivity error occurred", he.Message)
	})

	t.Run("sanitizes timeout errors", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, "context deadline exceeded: timeout")
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "A connectivity error occurred", he.Message)
	})

	t.Run("passes through safe messages", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(400, "name is required")
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, "name is required", he.Message)
	})

	t.Run("non-HTTP errors become generic 500", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return assert.AnError
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		assert.Equal(t, 500, he.Code)
		assert.Equal(t, "An internal error occurred", he.Message)
	})

	t.Run("sanitizes map messages", func(t *testing.T) {
		c, _ := newEchoContext("GET", "/")

		handler := m.Middleware(func(c echo.Context) error {
			return echo.NewHTTPError(500, map[string]any{
				"error": "sql: connection lost",
				"code":  42,
			})
		})

		err := handler(c)
		he := err.(*echo.HTTPError)
		msgMap, ok := he.Message.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "A database error occurred", msgMap["error"])
		// non-string values are passed through
		assert.Equal(t, 42, msgMap["code"])
	})
}

func TestErrorSanitizationMiddleware_RecoverMiddleware(t *testing.T) {
	t.Run("production panic returns generic error", func(t *testing.T) {
		m := &ErrorSanitizationMiddleware{IsProduction: true}
		c, rec := newEchoContext("GET", "/")

		handler := m.RecoverMiddleware(func(c echo.Context) error {
			panic("unexpected panic")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, 500, rec.Code)
	})

	t.Run("development panic includes details", func(t *testing.T) {
		m := &ErrorSanitizationMiddleware{IsProduction: false}
		c, rec := newEchoContext("GET", "/")

		handler := m.RecoverMiddleware(func(c echo.Context) error {
			panic("dev panic")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, 500, rec.Code)
	})

	t.Run("no panic passes through", func(t *testing.T) {
		m := &ErrorSanitizationMiddleware{IsProduction: true}
		c, rec := newEchoContext("GET", "/")

		handler := m.RecoverMiddleware(func(c echo.Context) error {
			return c.String(200, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, 200, rec.Code)
	})
}
