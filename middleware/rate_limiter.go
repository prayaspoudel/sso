package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rate     time.Duration
	limit    int
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// NewRateLimiter creates a new rate limiter
// rate: time window (e.g., 1 minute)
// limit: max requests within the window
func NewRateLimiter(rate time.Duration, limit int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		limit:    limit,
	}

	// Cleanup goroutine to remove old visitors
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware handler
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()

		// Check rate limit
		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow checks if the IP is allowed to make a request
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	// Reset counter if time window has passed
	if now.Sub(v.lastSeen) > rl.rate {
		v.count = 1
		v.lastSeen = now
		return true
	}

	// Check if limit exceeded
	if v.count >= rl.limit {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

// cleanup removes old visitors periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.rate*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// EndpointRateLimiter creates a rate limiter for specific endpoints
func EndpointRateLimiter(requests int, duration time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(duration, requests)
	return limiter.Middleware()
}
