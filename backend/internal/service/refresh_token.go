// Copyright 2024 ImageForge
// Refresh Token 服务 — 支持 Access Token 刷新和吊销。

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RefreshTokenStore 管理 Refresh Token 的存储和吊销。
type RefreshTokenStore struct {
	rdb *redis.Client
}

// NewRefreshTokenStore 创建 Refresh Token 存储。
// 如果 rdb 为 nil，使用内存存储（重启丢失，仅开发用）。
func NewRefreshTokenStore(rdb *redis.Client) *RefreshTokenStore {
	return &RefreshTokenStore{rdb: rdb}
}

// refreshTokenRecord 存储在 Redis 中的记录。
type refreshTokenRecord struct {
	UserID       int64     `json:"uid"`
	TokenVersion int       `json:"tv"`
	Role         string    `json:"role"`
	ExpiresAt    time.Time `json:"exp"`
	FamilyID     string    `json:"family"` // 令牌家族，用于检测重放
}

// Generate 生成新的 Refresh Token。
// 返回 token 字符串和家族 ID。
func (s *RefreshTokenStore) Generate(userID int64, role string, tokenVersion int, ttl time.Duration) (token string, familyID string, err error) {
	tokenBytes := make([]byte, 32)
	if _, err = rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	token = hex.EncodeToString(tokenBytes)

	familyBytes := make([]byte, 16)
	if _, err = rand.Read(familyBytes); err != nil {
		return "", "", fmt.Errorf("generate family: %w", err)
	}
	familyID = hex.EncodeToString(familyBytes)

	record := refreshTokenRecord{
		UserID:       userID,
		TokenVersion: tokenVersion,
		Role:         role,
		ExpiresAt:    time.Now().Add(ttl),
		FamilyID:     familyID,
	}

	key := refreshTokenKey(token)
	if s.rdb != nil {
		// Redis 存储
		data, _ := jsonMarshal(record)
		if err = s.rdb.Set(ctx(), key, data, ttl).Err(); err != nil {
			return "", "", fmt.Errorf("store refresh token: %w", err)
		}
		// 同时存储反向索引（用户 -> token 列表），便于吊销
		s.rdb.SAdd(ctx(), userTokensKey(userID), token)
		s.rdb.Expire(ctx(), userTokensKey(userID), ttl)
	} else {
		// 内存存储
		memoryStore.mu.Lock()
		memoryStore.tokens[key] = record
		memoryStore.mu.Unlock()
	}

	return token, familyID, nil
}

// Validate 验证 Refresh Token 是否有效。
// 返回关联的用户信息和家族 ID。
func (s *RefreshTokenStore) Validate(token string) (userID int64, role string, tokenVersion int, familyID string, err error) {
	key := refreshTokenKey(token)

	var record refreshTokenRecord
	if s.rdb != nil {
		data, e := s.rdb.Get(ctx(), key).Result()
		if e == redis.Nil {
			return 0, "", 0, "", errors.New("refresh token not found or expired")
		} else if e != nil {
			return 0, "", 0, "", fmt.Errorf("get refresh token: %w", e)
		}
		if e := jsonUnmarshal([]byte(data), &record); e != nil {
			return 0, "", 0, "", fmt.Errorf("decode refresh token: %w", e)
		}
	} else {
		memoryStore.mu.RLock()
		rec, ok := memoryStore.tokens[key]
		memoryStore.mu.RUnlock()
		if !ok {
			return 0, "", 0, "", errors.New("refresh token not found or expired")
		}
		record = rec
	}

	if time.Now().After(record.ExpiresAt) {
		s.Revoke(token) // 清理过期 token
		return 0, "", 0, "", errors.New("refresh token expired")
	}

	return record.UserID, record.Role, record.TokenVersion, record.FamilyID, nil
}

// Revoke 吊销指定的 Refresh Token。
func (s *RefreshTokenStore) Revoke(token string) error {
	key := refreshTokenKey(token)
	if s.rdb != nil {
		// 获取用户 ID 以清理反向索引
		data, _ := s.rdb.Get(ctx(), key).Result()
		if data != "" {
			var record refreshTokenRecord
			if err := jsonUnmarshal([]byte(data), &record); err == nil {
				s.rdb.SRem(ctx(), userTokensKey(record.UserID), token)
			}
		}
		return s.rdb.Del(ctx(), key).Err()
	}
	memoryStore.mu.Lock()
	delete(memoryStore.tokens, key)
	memoryStore.mu.Unlock()
	return nil
}

// RevokeAllUserTokens 吊销用户的所有 Refresh Token（密码重置时调用）。
func (s *RefreshTokenStore) RevokeAllUserTokens(userID int64) error {
	if s.rdb != nil {
		tokens, err := s.rdb.SMembers(ctx(), userTokensKey(userID)).Result()
		if err != nil {
			return err
		}
		for _, token := range tokens {
			s.rdb.Del(ctx(), refreshTokenKey(token))
		}
		return s.rdb.Del(ctx(), userTokensKey(userID)).Err()
	}
	// 内存模式：遍历清理
	memoryStore.mu.Lock()
	for key, rec := range memoryStore.tokens {
		if rec.UserID == userID {
			delete(memoryStore.tokens, key)
		}
	}
	memoryStore.mu.Unlock()
	return nil
}

// 辅助函数

func refreshTokenKey(token string) string {
	return "refresh_token:" + token
}

func userTokensKey(userID int64) string {
	return fmt.Sprintf("user_refresh_tokens:%d", userID)
}

func ctx() context.Context {
	return context.Background()
}

// 内存存储（开发用）
var memoryStore struct {
	mu     sync.RWMutex
	tokens map[string]refreshTokenRecord
}

func init() {
	memoryStore.tokens = make(map[string]refreshTokenRecord)
}

// JSON 辅助（避免循环导入 encoding/json）
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
