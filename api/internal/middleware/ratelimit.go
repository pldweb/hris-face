package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit is a fixed-window limiter keyed by client IP (docs/PRD.md F1).
// In-memory is the right size here: one VPS, one process, <200 employees. It
// resets on restart, which is acceptable for slowing brute force but is NOT a
// defence against a distributed attacker -- that belongs at the reverse proxy.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	type bucket struct {
		count int
		reset time.Time
	}

	var (
		mu      sync.Mutex
		buckets = map[string]*bucket{}
	)

	// Sweep expired keys so a long-lived process does not accumulate one entry
	// per IP that ever touched the endpoint.
	go func() {
		for range time.Tick(window) {
			mu.Lock()
			now := time.Now()
			for k, b := range buckets {
				if now.After(b.reset) {
					delete(buckets, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, ok := buckets[key]
		if !ok || now.After(b.reset) {
			b = &bucket{reset: now.Add(window)}
			buckets[key] = b
		}
		b.count++
		exceeded := b.count > limit
		retryAfter := int(time.Until(b.reset).Seconds()) + 1
		mu.Unlock()

		if exceeded {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				gin.H{"error": "Terlalu banyak percobaan. Coba lagi beberapa saat lagi."})
			return
		}
		c.Next()
	}
}
