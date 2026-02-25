package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAuditTestContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestExtractUserIDFromContext(t *testing.T) {
	t.Run("uses canonical user_id key", func(t *testing.T) {
		c := newAuditTestContext()
		c.Set("user_id", 12)

		id, err := extractUserID(c)
		require.NoError(t, err)
		assert.Equal(t, 12, id)
	})

	t.Run("supports legacy userID key", func(t *testing.T) {
		c := newAuditTestContext()
		c.Set("userID", 34)

		id, err := extractUserID(c)
		require.NoError(t, err)
		assert.Equal(t, 34, id)
	})
}

func TestExtractCollegeIDFromContext(t *testing.T) {
	t.Run("uses canonical college_id key", func(t *testing.T) {
		c := newAuditTestContext()
		c.Set("college_id", 9)

		id, err := extractCollegeID(c)
		require.NoError(t, err)
		assert.Equal(t, 9, id)
	})

	t.Run("supports legacy collegeID key", func(t *testing.T) {
		c := newAuditTestContext()
		c.Set("collegeID", 19)

		id, err := extractCollegeID(c)
		require.NoError(t, err)
		assert.Equal(t, 19, id)
	})
}
