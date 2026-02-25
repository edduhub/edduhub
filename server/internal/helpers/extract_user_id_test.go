package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestExtractUserID(t *testing.T) {
	t.Run("uses user_id when available", func(t *testing.T) {
		c := newTestContext()
		c.Set("user_id", 42)

		userID, err := ExtractUserID(c)
		require.NoError(t, err)
		assert.Equal(t, 42, userID)
	})

	t.Run("falls back to identity user id", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", &auth.Identity{UserID: 77})

		userID, err := ExtractUserID(c)
		require.NoError(t, err)
		assert.Equal(t, 77, userID)
	})

	t.Run("falls back to student_id", func(t *testing.T) {
		c := newTestContext()
		c.Set("student_id", 91)

		userID, err := ExtractUserID(c)
		require.NoError(t, err)
		assert.Equal(t, 91, userID)
	})

	t.Run("returns error when missing", func(t *testing.T) {
		c := newTestContext()

		_, err := ExtractUserID(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
