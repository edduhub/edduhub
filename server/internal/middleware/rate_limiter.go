package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiter holds the configuration for rate limiting
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter instance
// rate: number of requests per second
// burst: maximum burst size
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getVisitor returns the rate limiter for a given IP address
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// cleanupVisitors removes old entries from the visitors map
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, limiter := range rl.visitors {
			if limiter.Allow() {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an Echo middleware function for rate limiting
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP
			ip := c.RealIP()
			limiter := rl.getVisitor(ip)

			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":   "Too many requests",
					"message": "Rate limit exceeded. Please try again later.",
				})
			}

			return next(c)
		}
	}
}

// StrictRateLimiter creates a strict rate limiter for sensitive endpoints
// 5 requests per minute
func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(rate.Every(12*time.Second), 5)
}

// ModerateRateLimiter creates a moderate rate limiter
// 20 requests per minute
func ModerateRateLimiter() *RateLimiter {
	return NewRateLimiter(rate.Every(3*time.Second), 20)
}

// LenientRateLimiter creates a lenient rate limiter
// 100 requests per minute
func LenientRateLimiter() *RateLimiter {
	return NewRateLimiter(rate.Every(600*time.Millisecond), 100)
}
