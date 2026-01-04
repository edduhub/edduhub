package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	done     chan struct{}
	started  bool
	logger   zerolog.Logger
	cleanup  time.Duration
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        zerolog.ConsoleWriter{Out: nil}.Out,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	return &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    b,
		done:     make(chan struct{}),
		started:  false,
		logger:   logger,
		cleanup:  5 * time.Minute,
	}
}

func (rl *RateLimiter) SetCleanupInterval(d time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.cleanup = d
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
		return v.limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.done:
			rl.logger.Debug().Msg("Rate limiter cleanup stopping")
			return
		case <-ticker.C:
			rl.mu.Lock()
			var removed int
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.Sub(v.lastSeen) > rl.cleanup {
					delete(rl.visitors, ip)
					removed++
				}
			}
			rl.mu.Unlock()

			if removed > 0 {
				rl.logger.Debug().Int("removed", removed).Int("active", len(rl.visitors)).Msg("Rate limiter cleanup completed")
			}
		}
	}
}

func (rl *RateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.started {
		rl.logger.Info().Msg("Stopping rate limiter cleanup goroutine...")
		close(rl.done)
		rl.started = false
		rl.logger.Info().Msg("Rate limiter stopped")
	}
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	rl.mu.Lock()
	if !rl.started {
		rl.started = true
		go rl.cleanupVisitors()
		rl.logger.Info().
			Float64("rate", float64(rl.rate)).
			Int("burst", rl.burst).
			Dur("cleanup", rl.cleanup).
			Msg("Rate limiter started")
	}
	rl.mu.Unlock()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getVisitor(ip)

			if !limiter.Allow() {
				rl.logger.Warn().Str("ip", ip).Msg("Rate limit exceeded")
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":   "Too many requests",
					"message": "Rate limit exceeded. Please try again later.",
				})
			}

			return next(c)
		}
	}
}

func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(rate.Every(12*time.Second), 5)
}

func ModerateRateLimiter() *RateLimiter {
	return NewRateLimiter(rate.Every(3*time.Second), 20)
}

func LenientRateLimiter() *RateLimiter {
	return NewRateLimiter(rate.Every(600*time.Millisecond), 100)
}
