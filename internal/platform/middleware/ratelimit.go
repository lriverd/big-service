package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lriverd/big-service/internal/platform/cache"
)

type RateLimiter struct {
	cache         *cache.Cache
	limitPerMin   int
	mu            sync.Mutex
}

func NewRateLimiter(c *cache.Cache, limitPerMin int) *RateLimiter {
	return &RateLimiter{cache: c, limitPerMin: limitPerMin}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s", c.ClientIP())
		if userID, exists := c.Get("userID"); exists {
			key = fmt.Sprintf("ratelimit:user:%s", userID.(string))
		}

		rl.mu.Lock()
		val, found := rl.cache.Get(key)
		if !found {
			rl.cache.Set(key, 1, time.Minute)
			rl.mu.Unlock()
			rl.setHeaders(c, rl.limitPerMin, rl.limitPerMin-1)
			c.Next()
			return
		}

		count := val.(int)
		if count >= rl.limitPerMin {
			rl.mu.Unlock()
			rl.setHeaders(c, rl.limitPerMin, 0)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests",
				},
			})
			return
		}

		rl.cache.Set(key, count+1, time.Minute)
		rl.mu.Unlock()
		rl.setHeaders(c, rl.limitPerMin, rl.limitPerMin-count-1)
		c.Next()
	}
}

func (rl *RateLimiter) setHeaders(c *gin.Context, limit, remaining int) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
}

