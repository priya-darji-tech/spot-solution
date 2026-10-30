package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters = make(map[string]*ipLimiter)
	mu       sync.Mutex
)

func init() {
	// Background cleanup of stale IP limiters every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			for ip, entry := range limiters {
				if time.Since(entry.lastSeen) > 15*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func IPRateLimiter(rps float64, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		entry, exists := limiters[ip]
		if !exists {
			entry = &ipLimiter{
				limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
			limiters[ip] = entry
		}
		entry.lastSeen = time.Now()
		limiter := entry.limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests. Please slow down.",
			})
			return
		}

		c.Next()
	}
}
