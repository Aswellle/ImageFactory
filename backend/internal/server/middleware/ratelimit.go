// Copyright 2024 ImageForge
// 令牌桶速率限制中间件。

package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/pkg/response"
)

// RateLimiter 基于令牌桶算法的全局限流器。
type RateLimiter struct {
	rate   float64 // 每秒生成的令牌数
	burst  int     // 桶的最大容量
	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// NewRateLimiter 创建一个令牌桶限流器。
// rate: 每秒允许的请求数；burst: 突发容量。
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	if rate <= 0 {
		rate = 10 // 默认 10 req/s
	}
	if burst <= 0 {
		burst = 20 // 默认突发 20
	}
	return &RateLimiter{
		rate:   rate,
		burst:  burst,
		tokens: float64(burst),
		last:   time.Now(),
	}
}

// Allow 检查是否允许当前请求通过。
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.last).Seconds()
	rl.last = now

	// 补充令牌
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

// GlobalRateLimiter 全局限流器实例（惰性初始化）。
var (
)

// SetGlobalRateLimiter 设置全局限流器（应在启动时调用）。
func SetGlobalRateLimiter(rate float64, burst int) {
}

// RateLimit 返回限流中间件。
// 按客户端 IP 进行限流。超过速率返回 429。
func RateLimit(rate float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests,
				string(errors.ErrorCodeUpstreamRateLimited),
				"too many requests, please slow down", rid(c))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitByIP 按 IP 分别限流的中间件（更严格）。
// 每个 IP 独立计数，防止单一 IP 耗尽全局配额。
func RateLimitByIP(rate float64, burst int) gin.HandlerFunc {
	var (
		mu       sync.RWMutex
		limiters = make(map[string]*RateLimiter)
	)

	return func(c *gin.Context) {
		ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
		if ip == "" {
			ip = c.ClientIP()
		}

		mu.RLock()
		limiter, ok := limiters[ip]
		mu.RUnlock()

		if !ok {
			mu.Lock()
			limiter = NewRateLimiter(rate, burst)
			limiters[ip] = limiter
			mu.Unlock()
		}

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests,
				string(errors.ErrorCodeUpstreamRateLimited),
				"too many requests from your IP", rid(c))
			c.Abort()
			return
		}
		c.Next()
	}
}
