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

// resetCodeScript atomically checks and updates rate limits in Redis.
// Returns: 0 = allowed, 1 = rate limited, 2 = daily limit exceeded.
const resetCodeScript = `
local limit_key = KEYS[1]
local daily_key = KEYS[2]
local limit_ttl = tonumber(ARGV[1])
local daily_ttl = tonumber(ARGV[2])
local daily_max = tonumber(ARGV[3])

-- Check rate limit
if redis.call('EXISTS', limit_key) == 1 then
  return 1
end

-- Check and increment daily limit
local count = redis.call('INCR', daily_key)
if count == 1 then
  redis.call('EXPIRE', daily_key, daily_ttl)
end
if count > daily_max then
  return 2
end

-- Set rate limit key
redis.call('SET', limit_key, '1', 'EX', limit_ttl)
return 0
`

func (r *RateLimiter) allowRedis(ctx context.Context, email string) error {
	limitKey := fmt.Sprintf("pwdreset:limit:%s", email)
	dailyKey := fmt.Sprintf("pwdreset:daily:%s:%s", email, time.Now().Format("2006-01-02"))
	
	result, err := r.rdb.Eval(ctx, resetCodeScript, []string{limitKey, dailyKey}, 
		int(resetCodeRateLimit.Seconds()), int(24*time.Hour.Seconds()), resetCodeDailyLimit).Int64()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	
	switch result {
	case 1:
		ttl, _ := r.rdb.TTL(ctx, limitKey).Result()
		return fmt.Errorf("please wait %d seconds before requesting another code", int(ttl.Seconds())+1)
	case 2:
		return fmt.Errorf("daily limit exceeded, please try again tomorrow")
	default:
		return nil
	}
}

func (r *RateLimiter) allowMemory(email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	// Periodic cleanup of stale entries (every 100 calls)
	if len(r.memLimit) > 1000 {
		r.cleanupStaleEntries(now)
	}

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

// cleanupStaleEntries removes expired entries to prevent memory leaks.
func (r *RateLimiter) cleanupStaleEntries(now time.Time) {
	for email, last := range r.memLimit {
		if now.Sub(last) > 24*time.Hour {
			delete(r.memLimit, email)
		}
	}
	today := now.Format("2006-01-02")
	for key := range r.memDaily {
		if len(key) > 10 && key[len(key)-10:] != today {
			delete(r.memDaily, key)
		}
	}
}
