package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// resetCodeKeyPrefix is the Redis key prefix for password reset codes.
	resetCodeKeyPrefix = "pwdreset:"
	// resetCodeTTL is how long a verification code remains valid.
	resetCodeTTL = 5 * time.Minute
	// resetCodeLength is the number of digits in the code.
	resetCodeLength = 6
)

// ResetCodeStore persists password reset verification codes.
// When Redis is unavailable it transparently falls back to an in-memory map.
type ResetCodeStore struct {
	rdb *redis.Client

	mu  sync.RWMutex
	mem map[string]*memCodeEntry
}

// maxResetAttempts is the maximum number of failed attempts before a code is invalidated.
const maxResetAttempts = 5

// memCodeEntry is an in-memory verification code with expiry.
type memCodeEntry struct {
	code      string
	expiresAt time.Time
	attempts  int
}

// NewResetCodeStore builds a ResetCodeStore.
// If the Redis client is nil or the connection fails, the store falls back to in-memory.
func NewResetCodeStore(rdb *redis.Client) *ResetCodeStore {
	store := &ResetCodeStore{rdb: rdb, mem: make(map[string]*memCodeEntry)}
	if rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			store.rdb = nil
		}
	}
	return store
}

// Generate creates a new 6-digit verification code for the given email and stores it.
func (s *ResetCodeStore) Generate(email string) (string, error) {
	code := randomCode(resetCodeLength)

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.rdb.Set(ctx, resetCodeKey(email), code, resetCodeTTL).Err(); err != nil {
			return "", fmt.Errorf("failed to store reset code: %w", err)
		}
		return code, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.mem[email] = &memCodeEntry{code: code, expiresAt: time.Now().Add(resetCodeTTL)}
	return code, nil
}

// Verify checks whether the provided code matches the stored one for the email.
// On success, the code is consumed (deleted) to prevent reuse.
func (s *ResetCodeStore) Verify(email, code string) (bool, error) {
	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stored, err := s.rdb.Get(ctx, resetCodeKey(email)).Result()
		if err == redis.Nil {
			return false, nil // code expired or never existed
		}
		if err != nil {
			return false, fmt.Errorf("failed to read reset code: %w", err)
		}

		if stored != code {
			return false, nil
		}

		// Consume the code — one-time use.
		_ = s.rdb.Del(ctx, resetCodeKey(email)).Err()
		return true, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.mem[email]
	if !ok || time.Now().After(entry.expiresAt) {
		return false, nil
	}
	// Check attempt count to prevent brute force
	if entry.attempts >= maxResetAttempts {
		delete(s.mem, email)
		return false, nil
	}
	if entry.code != code {
		entry.attempts++
		s.mem[email] = entry
		return false, nil
	}
	delete(s.mem, email) // one-time use
	return true, nil
}

// Invalidate removes any existing code for the email (e.g. on password change).
func (s *ResetCodeStore) Invalidate(email string) error {
	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.rdb.Del(ctx, resetCodeKey(email)).Err()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.mem, email)
	return nil
}

// randomCode generates an n-digit numeric code using crypto/rand.
func randomCode(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b)
}

func resetCodeKey(email string) string {
	return resetCodeKeyPrefix + email
}
