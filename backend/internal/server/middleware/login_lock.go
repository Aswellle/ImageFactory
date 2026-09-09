// Copyright 2024 ImageForge
// 登录失败次数限制 — 防止暴力破解。

package middleware

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginAttemptTracker 跟踪登录失败次数并在超过阈值时锁定账户。
type LoginAttemptTracker struct {
	rdb         *redis.Client
	mu          sync.Mutex
	attempts    map[string]*loginAttempt
	maxAttempts int
	window      time.Duration
	lockout     time.Duration
}

type loginAttempt struct {
	count       int
	firstFail   time.Time
	lockedUntil *time.Time
}

// NewLoginAttemptTracker 创建登录失败追踪器。
func NewLoginAttemptTracker(rdb *redis.Client, maxAttempts int, window, lockout time.Duration) *LoginAttemptTracker {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	if lockout <= 0 {
		lockout = 15 * time.Minute
	}
	return &LoginAttemptTracker{
		rdb:         rdb,
		attempts:    make(map[string]*loginAttempt),
		maxAttempts: maxAttempts,
		window:      window,
		lockout:     lockout,
	}
}

// IsLocked 检查指定标识是否被锁定。
func (t *LoginAttemptTracker) IsLocked(key string) (bool, time.Duration) {
	if t.rdb != nil {
		data, err := t.rdb.Get(context.Background(), lockKey(key)).Result()
		if err == redis.Nil {
			return false, 0
		} else if err != nil {
			return false, 0
		}
		var until time.Time
		if err := json.Unmarshal([]byte(data), &until); err == nil {
			if time.Now().Before(until) {
				return true, time.Until(until)
			}
			t.rdb.Del(context.Background(), lockKey(key))
			t.rdb.Del(context.Background(), attemptsKey(key))
		}
		return false, 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	attempt, ok := t.attempts[key]
	if !ok {
		return false, 0
	}
	if attempt.lockedUntil != nil && time.Now().Before(*attempt.lockedUntil) {
		return true, time.Until(*attempt.lockedUntil)
	}
	if attempt.lockedUntil != nil && time.Now().After(*attempt.lockedUntil) {
		delete(t.attempts, key)
	}
	return false, 0
}

// RecordFailure 记录一次登录失败。
func (t *LoginAttemptTracker) RecordFailure(key string) (count int, locked bool) {
	if t.rdb != nil {
		pipe := t.rdb.Pipeline()
		pipe.Incr(context.Background(), attemptsKey(key))
		pipe.Expire(context.Background(), attemptsKey(key), t.window)
		_, _ = pipe.Exec(context.Background())
		c, _ := t.rdb.Get(context.Background(), attemptsKey(key)).Int()
		if c >= t.maxAttempts {
			until := time.Now().Add(t.lockout)
			_ = t.rdb.Set(context.Background(), lockKey(key), mustMarshal(until), t.lockout).Err()
			return c, true
		}
		return c, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	attempt, ok := t.attempts[key]
	if !ok || time.Since(attempt.firstFail) > t.window {
		attempt = &loginAttempt{firstFail: time.Now()}
		t.attempts[key] = attempt
	}
	attempt.count++
	if attempt.count >= t.maxAttempts {
		until := time.Now().Add(t.lockout)
		attempt.lockedUntil = &until
		return attempt.count, true
	}
	return attempt.count, false
}

// RecordSuccess 清除失败记录。
func (t *LoginAttemptTracker) RecordSuccess(key string) {
	if t.rdb != nil {
		t.rdb.Del(context.Background(), attemptsKey(key), lockKey(key))
		return
	}
	t.mu.Lock()
	delete(t.attempts, key)
	t.mu.Unlock()
}

func lockKey(key string) string     { return "login_lock:" + key }
func attemptsKey(key string) string { return "login_attempts:" + key }

func mustMarshal(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
