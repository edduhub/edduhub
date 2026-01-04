package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	r := rate.Limit(10)
	b := 5

	rl := NewRateLimiter(r, b)

	assert.NotNil(t, rl)
	assert.NotNil(t, rl.visitors)
	assert.Equal(t, r, rl.rate)
	assert.Equal(t, b, rl.burst)
	assert.NotNil(t, rl.done)
	assert.False(t, rl.started)
}

func TestRateLimiter_SetCleanupInterval(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 5)
	interval := 10 * time.Minute

	rl.SetCleanupInterval(interval)

	assert.Equal(t, interval, rl.cleanup)
}

func TestRateLimiter_getVisitor_NewVisitor(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 5)
	ip := "192.168.1.1"

	limiter := rl.getVisitor(ip)

	assert.NotNil(t, limiter)

	rl.mu.RLock()
	defer rl.mu.RUnlock()
	visitor, exists := rl.visitors[ip]
	assert.True(t, exists)
	assert.NotNil(t, visitor)
	assert.NotNil(t, visitor.limiter)
	assert.False(t, visitor.lastSeen.IsZero())
}

func TestRateLimiter_getVisitor_ExistingVisitor(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 5)
	ip := "192.168.1.1"

	firstLimiter := rl.getVisitor(ip)

	time.Sleep(10 * time.Millisecond)

	secondLimiter := rl.getVisitor(ip)

	assert.Same(t, firstLimiter, secondLimiter)

	rl.mu.RLock()
	defer rl.mu.RUnlock()
	visitor := rl.visitors[ip]
	assert.NotNil(t, visitor)
	assert.True(t, time.Since(visitor.lastSeen) < 100*time.Millisecond)
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 5)

	assert.NotPanics(t, func() {
		rl.Stop()
	})

	rl.mu.RLock()
	defer rl.mu.RUnlock()
	assert.False(t, rl.started)
}
