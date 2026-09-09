package repository

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	image_task "github.com/imageforge/imageforge/internal/domain/image_task"
)

const imageTaskKeyPrefix = "image_task:"

// RedisImageTaskStore is the Redis-backed implementation of imagetask.Store.
// When Redis is unavailable it transparently falls back to an in-memory map.
type RedisImageTaskStore struct {
	rdb *redis.Client

	mu  sync.RWMutex
	mem map[string]*memEntry
}

// memEntry is an in-memory task record with expiry.
type memEntry struct {
	task      *image_task.Record
	expiresAt time.Time
}

// NewRedisImageTaskStore builds a RedisImageTaskStore.
// If the Redis client is nil or the connection fails, the store falls back to in-memory.
func NewRedisImageTaskStore(rdb *redis.Client) *RedisImageTaskStore {
	store := &RedisImageTaskStore{rdb: rdb, mem: make(map[string]*memEntry)}
	if rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			store.rdb = nil
		}
	}
	return store
}

// Save persists a task record with the given TTL.
func (s *RedisImageTaskStore) Save(task *image_task.Record, ttl time.Duration) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.rdb.Set(ctx, imageTaskKey(task.ID), data, ttl).Err()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.mem[task.ID] = &memEntry{task: task, expiresAt: time.Now().Add(ttl)}
	return nil
}

// Get retrieves a task by its ID.
func (s *RedisImageTaskStore) Get(id string) (*image_task.Record, error) {
	if s.rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		data, err := s.rdb.Get(ctx, imageTaskKey(id)).Bytes()
		if err != nil {
			return nil, err
		}
		var rec image_task.Record
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, err
		}
		return &rec, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.mem[id]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, errTaskNotFound
	}
	return entry.task, nil
}

func imageTaskKey(id string) string {
	return imageTaskKeyPrefix + strings.TrimSpace(id)
}

// Ensure image_task types are referenced.
var _ = image_task.StatusProcessing

var errTaskNotFound = image_task.ErrImageTaskNotFound
