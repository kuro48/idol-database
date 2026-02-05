package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter はIPアドレスベースのレート制限を実装するミドルウェア
type RateLimiter struct {
	limiters        map[string]*clientLimiter
	mu              sync.Mutex
	rate            rate.Limit
	burst           int
	ttl             time.Duration
	cleanupInterval time.Duration
	now             func() time.Time
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter は新しいRateLimiterを作成する
// ratePerSecond: 1秒あたりのリクエスト数
// burst: バースト許容数
func NewRateLimiter(ratePerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters:        make(map[string]*clientLimiter),
		rate:            rate.Limit(ratePerSecond),
		burst:           burst,
		ttl:             10 * time.Minute,
		cleanupInterval: time.Minute,
		now:             time.Now,
	}
	rl.startCleanup()
	return rl
}

// getLimiter は指定されたIPアドレスのリミッターを取得または作成する
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	entry, exists := rl.limiters[ip]
	if !exists || rl.isExpired(entry, now) {
		entry = &clientLimiter{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: now,
		}
		rl.limiters[ip] = entry
	} else {
		entry.lastSeen = now
	}

	return entry.limiter
}

// Limit はレート制限を適用するミドルウェア関数を返す
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, NewTooManyRequestsError("リクエストが多すぎます。しばらくしてから再度お試しください。"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupOldLimiters は古いリミッターをクリーンアップする
func (rl *RateLimiter) CleanupOldLimiters() {
	rl.cleanup()
}

func (rl *RateLimiter) startCleanup() {
	if rl.cleanupInterval <= 0 || rl.ttl <= 0 {
		return
	}

	ticker := time.NewTicker(rl.cleanupInterval)
	go func() {
		for range ticker.C {
			rl.cleanup()
		}
	}()
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	for ip, entry := range rl.limiters {
		if rl.isExpired(entry, now) {
			delete(rl.limiters, ip)
		}
	}
}

func (rl *RateLimiter) isExpired(entry *clientLimiter, now time.Time) bool {
	if rl.ttl <= 0 {
		return false
	}
	return now.Sub(entry.lastSeen) > rl.ttl
}
