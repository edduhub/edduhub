package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestRateLimiter_Middleware_AllowsRequests(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 10)
	defer rl.Stop()

	mw := rl.Middleware()
	e := echo.New()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-Ip", "1.2.3.4")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiter_Middleware_BlocksExcessRequests(t *testing.T) {
	// 1 request per 10 seconds, burst of 1
	rl := NewRateLimiter(rate.Every(10*time.Second), 1)
	defer rl.Stop()

	mw := rl.Middleware()
	e := echo.New()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request should pass
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Real-Ip", "5.6.7.8")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	err := handler(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Real-Ip", "5.6.7.8")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = handler(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
}

func TestRateLimiter_Middleware_DifferentIPsAreSeparate(t *testing.T) {
	rl := NewRateLimiter(rate.Every(10*time.Second), 1)
	defer rl.Stop()

	mw := rl.Middleware()
	e := echo.New()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First IP
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Real-Ip", "10.0.0.1")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	_ = handler(c1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Second IP should also pass
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Real-Ip", "10.0.0.2")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	_ = handler(c2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestRateLimiter_Middleware_StartsOnce(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 5)
	defer rl.Stop()

	_ = rl.Middleware()
	_ = rl.Middleware()

	rl.mu.RLock()
	defer rl.mu.RUnlock()
	assert.True(t, rl.started)
}

func TestRateLimiter_StopWhenStarted(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 5)
	_ = rl.Middleware() // This starts the cleanup goroutine

	rl.mu.RLock()
	started := rl.started
	rl.mu.RUnlock()
	require.True(t, started)

	assert.NotPanics(t, func() {
		rl.Stop()
	})

	rl.mu.RLock()
	defer rl.mu.RUnlock()
	assert.False(t, rl.started)
}

// --- Preset limiter factories ---

func TestStrictRateLimiter(t *testing.T) {
	rl := StrictRateLimiter()
	assert.NotNil(t, rl)
	assert.Equal(t, rate.Every(12*time.Second), rl.rate)
	assert.Equal(t, 5, rl.burst)
}

func TestModerateRateLimiter(t *testing.T) {
	rl := ModerateRateLimiter()
	assert.NotNil(t, rl)
	assert.Equal(t, rate.Every(3*time.Second), rl.rate)
	assert.Equal(t, 20, rl.burst)
}

func TestLenientRateLimiter(t *testing.T) {
	rl := LenientRateLimiter()
	assert.NotNil(t, rl)
	assert.Equal(t, rate.Every(600*time.Millisecond), rl.rate)
	assert.Equal(t, 100, rl.burst)
}
