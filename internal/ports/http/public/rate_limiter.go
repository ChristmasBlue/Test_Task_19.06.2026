package public

import (
	"net/http"
	"sync"
	"test_task/tools/config"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(cfg config.Config) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     rate.Limit(cfg.RateLimiterRate()),
		burst:    cfg.RateLimiterBurst(),
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter = rate.NewLimiter(rl.rate, rl.burst)
	rl.visitors[ip] = limiter
	return limiter
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(1 * time.Minute)
		rl.mu.Lock()
		for ip, limiter := range rl.visitors {
			if limiter.Allow() {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.getLimiter(ip)

		if !l.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
