// Copyright 2024 ImageForge
// AES-256-GCM 加密工具，用于敏感凭证的 at-rest 加密。

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// 敏感凭证字段名（这些字段在存储时会被加密）。
var sensitiveFields = []string{
	"api_key",
	"access_token",
	"refresh_token",
	"secret_key",
	"client_secret",
	"password",
}

// keyStore 保存全局加密密钥（启动时从环境变量加载）。
var (
	keyStore struct {
		mu  sync.RWMutex
		key []byte
	}
)

// SetKey 设置加密密钥（应在应用启动时调用一次）。
// 密钥必须是 32 字节（AES-256）。如果 keyStr 为空，则禁用加密（开发模式）。
func SetKey(keyStr string) error {
	if keyStr == "" {
		// 空密钥 = 不加密（开发模式）
		keyStore.mu.Lock()
		keyStore.key = nil
		keyStore.mu.Unlock()
		return nil
	}

	// 支持 base64 编码的密钥或原始字符串
	var key []byte
	decoded, err := base64.StdEncoding.DecodeString(keyStr)
	if err == nil && len(decoded) == 32 {
		key = decoded
	} else if len(keyStr) == 32 {
		key = []byte(keyStr)
	} else {
		return fmt.Errorf("encryption key must be 32 bytes (got %d bytes); use base64 or raw string", len(keyStr))
	}

	keyStore.mu.Lock()
	keyStore.key = key
	keyStore.mu.Unlock()
	return nil
}

// IsEnabled 返回加密是否已启用。
func IsEnabled() bool {
	keyStore.mu.RLock()
	defer keyStore.mu.RUnlock()
	return keyStore.key != nil
}

// GenerateKey 生成一个随机的 32 字节密钥（base64 编码），用于初始化配置。
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// Encrypt 加密明文数据，返回 base64 编码的密文。
// 如果加密未启用，直接返回原始数据的 base64 编码。
func Encrypt(plaintext []byte) (string, error) {
	if !IsEnabled() {
		return base64.StdEncoding.EncodeToString(plaintext), nil
	}

	keyStore.mu.RLock()
	key := keyStore.key
	keyStore.mu.RUnlock()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的密文，返回明文。
// 如果加密未启用，直接 base64 解码。
func Decrypt(encoded string) ([]byte, error) {
	if !IsEnabled() {
		return base64.StdEncoding.DecodeString(encoded)
	}

	keyStore.mu.RLock()
	key := keyStore.key
	keyStore.mu.RUnlock()

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptCredentials 加密凭证 map 中的敏感字段。
// 非敏感字段保持明文。返回新的 map（不修改原始 map）。
func EncryptCredentials(creds map[string]any) (map[string]any, error) {
	if !IsEnabled() || len(creds) == 0 {
		return creds, nil
	}

	result := make(map[string]any, len(creds))
	for k, v := range creds {
		if isSensitiveField(k) && v != nil {
			// 将值序列化为 JSON 后加密
			plain, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("marshal credential %s: %w", k, err)
			}
			encrypted, err := Encrypt(plain)
			if err != nil {
				return nil, fmt.Errorf("encrypt credential %s: %w", k, err)
			}
			result[k] = "__ENC__:" + encrypted
		} else {
			result[k] = v
		}
	}
	return result, nil
}

// DecryptCredentials 解密凭证 map 中的敏感字段。
// 识别 "__ENC__:" 前缀并解密。
func DecryptCredentials(creds map[string]any) (map[string]any, error) {
	if !IsEnabled() || len(creds) == 0 {
		return creds, nil
	}

	result := make(map[string]any, len(creds))
	for k, v := range creds {
		if strVal, ok := v.(string); ok && strings.HasPrefix(strVal, "__ENC__:") {
			encrypted := strings.TrimPrefix(strVal, "__ENC__:")
			plain, err := Decrypt(encrypted)
			if err != nil {
				return nil, fmt.Errorf("decrypt credential %s: %w", k, err)
			}
			// 尝试解析为 JSON（可能是字符串、数字等）
			var parsed any
			if err := json.Unmarshal(plain, &parsed); err == nil {
				result[k] = parsed
			} else {
				result[k] = string(plain)
			}
		} else {
			result[k] = v
		}
	}
	return result, nil
}

// isSensitiveField 检查字段名是否为敏感字段（不区分大小写）。
func isSensitiveField(name string) bool {
	for _, s := range sensitiveFields {
		if subtle.ConstantTimeCompare([]byte(strings.ToLower(name)), []byte(s)) == 1 {
			return true
		}
	}
	return false
}

// init 从环境变量 IF_CREDENTIAL_KEY 自动加载密钥（如果存在）。
func init() {
	if key := os.Getenv("IF_CREDENTIAL_KEY"); key != "" {
		_ = SetKey(key)
	}
}
