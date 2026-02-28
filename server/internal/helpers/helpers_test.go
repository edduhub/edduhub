package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ExtractCollegeID ---

func TestExtractCollegeID(t *testing.T) {
	t.Run("returns college_id when set as int", func(t *testing.T) {
		c := newTestContext()
		c.Set("college_id", 5)

		id, err := ExtractCollegeID(c)
		require.NoError(t, err)
		assert.Equal(t, 5, id)
	})

	t.Run("returns error when college_id is nil", func(t *testing.T) {
		c := newTestContext()

		_, err := ExtractCollegeID(c)
		require.Error(t, err)
	})

	t.Run("returns error when college_id is wrong type", func(t *testing.T) {
		c := newTestContext()
		c.Set("college_id", "not-an-int")

		_, err := ExtractCollegeID(c)
		require.Error(t, err)
	})
}

// --- ExtractStudentID ---
// Note: ExtractStudentID uses Error() which writes JSON response and returns nil from c.JSON().

func TestExtractStudentID(t *testing.T) {
	t.Run("returns student_id when set as int", func(t *testing.T) {
		c := newTestContext()
		c.Set("student_id", 10)

		id, err := ExtractStudentID(c)
		require.NoError(t, err)
		assert.Equal(t, 10, id)
	})

	t.Run("returns zero when student_id is nil", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		id, _ := ExtractStudentID(c)
		assert.Equal(t, 0, id)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("returns zero when student_id is wrong type", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("student_id", "not-an-int")

		id, _ := ExtractStudentID(c)
		assert.Equal(t, 0, id)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

// --- GetIDFromParam ---
// Note: GetIDFromParam uses Error() which writes JSON response and returns nil from c.JSON().
// So we check the returned id is 0 and the recorder's status code for error cases.

func TestGetIDFromParam(t *testing.T) {
	t.Run("returns parsed int for valid param", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("42")

		id, err := GetIDFromParam(c, "id")
		require.NoError(t, err)
		assert.Equal(t, 42, id)
	})

	t.Run("returns zero for empty param", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		id, _ := GetIDFromParam(c, "id")
		assert.Equal(t, 0, id)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns zero for non-numeric param", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("abc")

		id, _ := GetIDFromParam(c, "id")
		assert.Equal(t, 0, id)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns zero for zero param", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("0")

		id, _ := GetIDFromParam(c, "id")
		assert.Equal(t, 0, id)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns zero for negative param", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("-1")

		id, _ := GetIDFromParam(c, "id")
		assert.Equal(t, 0, id)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// --- GetKratosID ---

func TestGetKratosID(t *testing.T) {
	t.Run("returns kratos ID from identity", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", &auth.Identity{ID: "abc-123"})

		id, err := GetKratosID(c)
		require.NoError(t, err)
		assert.Equal(t, "abc-123", id)
	})

	t.Run("returns error when identity is nil", func(t *testing.T) {
		c := newTestContext()

		_, err := GetKratosID(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "identity not found")
	})

	t.Run("returns error when identity is wrong type", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", "not-an-identity")

		_, err := GetKratosID(c)
		require.Error(t, err)
	})

	t.Run("returns error when kratos ID is empty", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", &auth.Identity{ID: ""})

		_, err := GetKratosID(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kratos ID is empty")
	})
}

// --- GetUserRole ---

func TestGetUserRole(t *testing.T) {
	t.Run("returns role from identity", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", &auth.Identity{
			Traits: auth.Traits{Role: "admin"},
		})

		role, err := GetUserRole(c)
		require.NoError(t, err)
		assert.Equal(t, "admin", role)
	})

	t.Run("returns error when identity is missing", func(t *testing.T) {
		c := newTestContext()

		_, err := GetUserRole(c)
		require.Error(t, err)
	})

	t.Run("returns empty string when role is empty", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", &auth.Identity{
			Traits: auth.Traits{Role: ""},
		})

		role, err := GetUserRole(c)
		require.NoError(t, err)
		assert.Equal(t, "", role)
	})
}

// --- Error ---

func TestErrorResponse(t *testing.T) {
	t.Run("returns JSON error response", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = Error(c, "something went wrong", http.StatusBadRequest)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "something went wrong", resp.Error)
	})

	t.Run("returns 500 error response", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = Error(c, "internal error", http.StatusInternalServerError)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

// --- Success ---

func TestSuccessResponse(t *testing.T) {
	t.Run("returns JSON success response", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = Success(c, map[string]string{"key": "value"}, http.StatusOK)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp SuccessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("returns 201 created", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = Success(c, nil, http.StatusCreated)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

// --- NotFound ---

func TestNotFoundResponse(t *testing.T) {
	t.Run("returns 404 JSON response", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = NotFound(c, "resource not found", http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var resp NotFoundResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.Status)
		assert.Equal(t, "resource not found", resp.Data)
	})
}

// --- ExtractUserID edge cases ---

func TestExtractUserID_TypeErrors(t *testing.T) {
	t.Run("returns error when user_id is wrong type", func(t *testing.T) {
		c := newTestContext()
		c.Set("user_id", "not-an-int")

		_, err := ExtractUserID(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not an integer")
	})

	t.Run("returns error when student_id fallback is wrong type", func(t *testing.T) {
		c := newTestContext()
		c.Set("student_id", "not-an-int")

		_, err := ExtractUserID(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not an integer")
	})

	t.Run("skips identity with zero UserID", func(t *testing.T) {
		c := newTestContext()
		c.Set("identity", &auth.Identity{UserID: 0})
		c.Set("student_id", 99)

		userID, err := ExtractUserID(c)
		require.NoError(t, err)
		assert.Equal(t, 99, userID)
	})
}
