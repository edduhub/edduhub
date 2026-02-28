package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- AppError ---

func TestAppError_Error(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		inner := errors.New("db failure")
		appErr := NewAppError(500, "INTERNAL", "something failed", inner)
		assert.Equal(t, "something failed: db failure", appErr.Error())
	})

	t.Run("without wrapped error", func(t *testing.T) {
		appErr := NewAppError(400, "BAD_REQUEST", "bad input", nil)
		assert.Equal(t, "bad input", appErr.Error())
	})
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := NewAppError(422, "VALIDATION_ERROR", "invalid", nil)
	details := map[string]any{"field": "name"}
	result := appErr.WithDetails(details)

	assert.Same(t, appErr, result)
	assert.Equal(t, details, appErr.Details)
}

// --- Error constructors ---

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string, error) *AppError
		status   int
		code     string
	}{
		{"BadRequestError", BadRequestError, http.StatusBadRequest, "BAD_REQUEST"},
		{"UnauthorizedError", UnauthorizedError, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"ForbiddenError", ForbiddenError, http.StatusForbidden, "FORBIDDEN"},
		{"NotFoundError", NotFoundError, http.StatusNotFound, "NOT_FOUND"},
		{"ConflictError", ConflictError, http.StatusConflict, "CONFLICT"},
		{"InternalServerError", InternalServerError, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"},
		{"ServiceUnavailableError", ServiceUnavailableError, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn("test message", errors.New("inner"))
			assert.Equal(t, tc.status, err.Status)
			assert.Equal(t, tc.code, err.Code)
			assert.Equal(t, "test message", err.Message)
			assert.NotNil(t, err.Err)
		})
	}
}

func TestValidationError(t *testing.T) {
	details := map[string]any{"email": []string{"is required"}}
	err := ValidationError("Validation failed", details)

	assert.Equal(t, http.StatusUnprocessableEntity, err.Status)
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, details, err.Details)
	assert.Nil(t, err.Err)
}

// --- getErrorCode ---

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		status int
		code   string
	}{
		{http.StatusBadRequest, "BAD_REQUEST"},
		{http.StatusUnauthorized, "UNAUTHORIZED"},
		{http.StatusForbidden, "FORBIDDEN"},
		{http.StatusNotFound, "NOT_FOUND"},
		{http.StatusConflict, "CONFLICT"},
		{http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED"},
		{http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"},
		{http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE"},
		{http.StatusTeapot, "UNKNOWN_ERROR"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("status_%d", tc.status), func(t *testing.T) {
			assert.Equal(t, tc.code, getErrorCode(tc.status))
		})
	}
}

// --- ErrorHandlerMiddleware ---

func newEchoContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestErrorHandlerMiddleware_NoError(t *testing.T) {
	mw := ErrorHandlerMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestErrorHandlerMiddleware_AppError(t *testing.T) {
	mw := ErrorHandlerMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		return BadRequestError("invalid input", nil)
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "BAD_REQUEST", resp.Code)
	assert.Equal(t, "invalid input", resp.Message)
}

func TestErrorHandlerMiddleware_HTTPError(t *testing.T) {
	mw := ErrorHandlerMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "page not found")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_FOUND", resp.Code)
}

func TestErrorHandlerMiddleware_JSONParseError(t *testing.T) {
	mw := ErrorHandlerMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		return &json.UnmarshalTypeError{Value: "string", Type: nil}
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "INVALID_JSON", resp.Code)
}

func TestErrorHandlerMiddleware_UnknownError(t *testing.T) {
	mw := ErrorHandlerMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		return errors.New("something unexpected")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "INTERNAL_ERROR", resp.Code)
}

// --- RecoverMiddleware ---

func TestRecoverMiddleware_NoPanic(t *testing.T) {
	mw := RecoverMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecoverMiddleware_PanicWithError(t *testing.T) {
	mw := RecoverMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		panic(errors.New("boom"))
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "PANIC_RECOVERED", resp.Code)
}

func TestRecoverMiddleware_PanicWithString(t *testing.T) {
	mw := RecoverMiddleware()
	c, rec := newEchoContext(http.MethodGet, "/")

	handler := mw(func(c echo.Context) error {
		panic("string panic")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
