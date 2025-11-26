package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter はIPアドレスベースのレート制限を実装するミドルウェア
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter は新しいRateLimiterを作成する
// ratePerSecond: 1秒あたりのリクエスト数
// burst: バースト許容数
func NewRateLimiter(ratePerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(ratePerSecond),
		burst:    burst,
	}
}

// getLimiter は指定されたIPアドレスのリミッターを取得または作成する
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// Limit はレート制限を適用するミドルウェア関数を返す
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "リクエストが多すぎます。しばらくしてから再度お試しください。",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupOldLimiters は古いリミッターを定期的にクリーンアップする
// 本番環境では、定期的にこの関数を呼び出すことを推奨
func (rl *RateLimiter) CleanupOldLimiters() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limiters = make(map[string]*rate.Limiter)
}
