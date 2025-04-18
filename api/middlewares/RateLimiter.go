package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     int
	window   time.Duration
}

type visitor struct {
	lastSeen time.Time
	count    int
}

func NewRateLimiter(rate int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// 定期清理旧的访问者
	go rl.cleanupVisitors()
	return rl
}

func (rl *rateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.window)

		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Limit(c *gin.Context) {
	ip := c.ClientIP()

	rl.mu.Lock()
	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{}
		rl.visitors[ip] = v
	}

	v.count++
	v.lastSeen = time.Now()

	if v.count > rl.rate {
		rl.mu.Unlock()
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		return
	}
	rl.mu.Unlock()

	c.Next()
}

func RateLimiter(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)
	return limiter.Limit
}
