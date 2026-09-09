// Copyright 2024 ImageForge
// Refresh Token 服务 — 支持 Access Token 刷新和吊销。

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)


// RefreshTokenStore 管理 Refresh Token 的存储和吊销。
type RefreshTokenStore struct {
	rdb *redis.Client
	ctx context.Context
}

// NewRefreshTokenStore 创建 Refresh Token 存储。
// 如果 rdb 为 nil，使用内存存储（重启丢失，仅开发用）。
func NewRefreshTokenStore(rdb *redis.Client) *RefreshTokenStore {
	return &RefreshTokenStore{rdb: rdb, ctx: context.Background()}
}

// NewRefreshTokenStoreWithContext 创建带上下文的 Refresh Token 存储。
func NewRefreshTokenStoreWithContext(rdb *redis.Client, ctx context.Context) *RefreshTokenStore {
	return &RefreshTokenStore{rdb: rdb, ctx: ctx}
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
		if err = s.rdb.Set(s.ctx, key, data, ttl).Err(); err != nil {
			return "", "", fmt.Errorf("store refresh token: %w", err)
		}
		// 同时存储反向索引（用户 -> token 列表），便于吊销
		s.rdb.SAdd(s.ctx, userTokensKey(userID), token)
		s.rdb.Expire(s.ctx, userTokensKey(userID), ttl)
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
	if s.rdb != nil {
		data, err := s.rdb.Get(s.ctx, key).Result()
		if err != nil {
			return 0, "", 0, "", fmt.Errorf("refresh token not found")
		}
		var record refreshTokenRecord
		if err := jsonUnmarshal([]byte(data), &record); err != nil {
			return 0, "", 0, "", fmt.Errorf("invalid token record")
		}
		return record.UserID, record.Role, record.TokenVersion, record.FamilyID, nil
	}
	// 内存模式
	memoryStore.mu.RLock()
	record, ok := memoryStore.tokens[key]
	memoryStore.mu.RUnlock()
	if !ok {
		return 0, "", 0, "", fmt.Errorf("refresh token not found")
	}
	return record.UserID, record.Role, record.TokenVersion, record.FamilyID, nil
}

// Revoke 吊销指定的 Refresh Token。
func (s *RefreshTokenStore) Revoke(token string) error {
	key := refreshTokenKey(token)
	if s.rdb != nil {
		// 获取用户 ID 以清理反向索引
		data, _ := s.rdb.Get(s.ctx, key).Result()
		if data != "" {
			var record refreshTokenRecord
			if err := jsonUnmarshal([]byte(data), &record); err == nil {
				s.rdb.SRem(s.ctx, userTokensKey(record.UserID), token)
			}
		}
		return s.rdb.Del(s.ctx, key).Err()
	}
	memoryStore.mu.Lock()
	delete(memoryStore.tokens, key)
	memoryStore.mu.Unlock()
	return nil
}

// RevokeAllUserTokens 吊销用户的所有 Refresh Token（密码重置时调用）。
func (s *RefreshTokenStore) RevokeAllUserTokens(userID int64) error {
	if s.rdb != nil {
		tokens, err := s.rdb.SMembers(s.ctx, userTokensKey(userID)).Result()
		if err != nil {
			return err
		}
		for _, token := range tokens {
			s.rdb.Del(s.ctx, refreshTokenKey(token))
		}
		return s.rdb.Del(s.ctx, userTokensKey(userID)).Err()
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

// 内存存储（开发用）
var memoryStore struct {
	mu     sync.RWMutex
	tokens map[string]refreshTokenRecord
}

func init() {
	memoryStore.tokens = make(map[string]refreshTokenRecord)
}

// JSON 辅助（避免循环导入 encoding/json）
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
