package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	e := echo.New()
	mw := SecurityHeaders()

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.NoError(t, err)

	headers := rec.Header()

	t.Run("sets Content-Security-Policy", func(t *testing.T) {
		csp := headers.Get("Content-Security-Policy")
		assert.Contains(t, csp, "default-src 'self'")
		assert.Contains(t, csp, "frame-ancestors 'none'")
	})

	t.Run("sets Strict-Transport-Security", func(t *testing.T) {
		hsts := headers.Get("Strict-Transport-Security")
		assert.Contains(t, hsts, "max-age=31536000")
		assert.Contains(t, hsts, "includeSubDomains")
	})

	t.Run("sets X-Frame-Options", func(t *testing.T) {
		assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	})

	t.Run("sets X-Content-Type-Options", func(t *testing.T) {
		assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	})

	t.Run("sets X-XSS-Protection", func(t *testing.T) {
		assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	})

	t.Run("sets Referrer-Policy", func(t *testing.T) {
		assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
	})

	t.Run("sets Permissions-Policy", func(t *testing.T) {
		pp := headers.Get("Permissions-Policy")
		assert.Contains(t, pp, "geolocation=()")
		assert.Contains(t, pp, "camera=()")
	})

	t.Run("sets X-Permitted-Cross-Domain-Policies", func(t *testing.T) {
		assert.Equal(t, "none", headers.Get("X-Permitted-Cross-Domain-Policies"))
	})

	t.Run("sets Cache-Control", func(t *testing.T) {
		assert.Contains(t, headers.Get("Cache-Control"), "no-store")
		assert.Equal(t, "no-cache", headers.Get("Pragma"))
	})
}

func TestSecurityHeaders_CallsNext(t *testing.T) {
	e := echo.New()
	mw := SecurityHeaders()
	called := false

	handler := mw(func(c echo.Context) error {
		called = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = handler(c)
	assert.True(t, called)
}

func TestPublicCacheHeaders(t *testing.T) {
	e := echo.New()

	t.Run("sets cache-control with maxAge", func(t *testing.T) {
		mw := PublicCacheHeaders(3600)
		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, "public, max-age=3600", rec.Header().Get("Cache-Control"))
	})

	t.Run("sets zero maxAge", func(t *testing.T) {
		mw := PublicCacheHeaders(0)
		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, "public, max-age=0", rec.Header().Get("Cache-Control"))
	})
}
