package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// resetCodeRateLimit is the minimum time between reset code requests per email.
	resetCodeRateLimit = 60 * time.Second
	// resetCodeDailyLimit is the maximum number of reset code requests per email per day.
	resetCodeDailyLimit = 5
)

// RateLimiter provides rate limiting for API operations.
// When Redis is available it uses Redis for distributed rate limiting;
// otherwise it falls back to an in-memory map.
type RateLimiter struct {
	rdb *redis.Client

	mu       sync.RWMutex
	memLimit map[string]time.Time // email -> last request time
	memDaily map[string]int       // email -> count today
}

// NewRateLimiter builds a RateLimiter.
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{
		rdb:      rdb,
		memLimit: make(map[string]time.Time),
		memDaily: make(map[string]int),
	}
}

// AllowResetCode checks if a reset code request is allowed for the given email.
// Returns nil if allowed, or an error describing the rate limit.
func (r *RateLimiter) AllowResetCode(ctx context.Context, email string) error {
	if r.rdb != nil {
		return r.allowRedis(ctx, email)
	}
	return r.allowMemory(email)
}

func (r *RateLimiter) allowRedis(ctx context.Context, email string) error {
	// Check rate limit (60 seconds between requests)
	limitKey := fmt.Sprintf("pwdreset:limit:%s", email)
	exists, err := r.rdb.Exists(ctx, limitKey).Result()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if exists > 0 {
		ttl, _ := r.rdb.TTL(ctx, limitKey).Result()
		return fmt.Errorf("please wait %d seconds before requesting another code", int(ttl.Seconds())+1)
	}

	// Check daily limit (5 per day)
	dailyKey := fmt.Sprintf("pwdreset:daily:%s:%s", email, time.Now().Format("2006-01-02"))
	count, err := r.rdb.Incr(ctx, dailyKey).Result()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if count == 1 {
		// Set expiry at end of day
		r.rdb.Expire(ctx, dailyKey, 24*time.Hour)
	}
	if count > int64(resetCodeDailyLimit) {
		return fmt.Errorf("daily limit exceeded, please try again tomorrow")
	}

	// Set rate limit key
	r.rdb.Set(ctx, limitKey, "1", resetCodeRateLimit)

	return nil
}

func (r *RateLimiter) allowMemory(email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Check rate limit (60 seconds between requests)
	if last, ok := r.memLimit[email]; ok {
		if now.Sub(last) < resetCodeRateLimit {
			wait := int(resetCodeRateLimit.Seconds() - now.Sub(last).Seconds())
			return fmt.Errorf("please wait %d seconds before requesting another code", wait)
		}
	}

	// Check daily limit (5 per day) - reset count if it's a new day
	today := now.Format("2006-01-02")
	dailyKey := email + ":" + today
	if count, ok := r.memDaily[dailyKey]; ok && count >= resetCodeDailyLimit {
		return fmt.Errorf("daily limit exceeded, please try again tomorrow")
	}

	// Update tracking
	r.memLimit[email] = now
	r.memDaily[dailyKey]++

	return nil
}
