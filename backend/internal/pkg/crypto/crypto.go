// Copyright 2024 ImageForge
// AES-256-GCM 加密工具，用于敏感凭证的 at-rest 加密。

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"
)

// 敏感凭证字段名（这些字段在存储时会被加密）。
var sensitiveFields = []string{
	"api_key",
	"service_account",
	"credentials",
	"secret",
	"access_token",
	"refresh_token",
}

// keyStore 保存全局加密密钥（启动时从环境变量加载）。
var (
	mu       sync.RWMutex
	key      []byte
	enabled  bool
)

// SetKey 设置加密密钥（应在应用启动时调用一次）。
// 密钥必须是 32 字节（AES-256）。如果 keyStr 为空，则禁用加密（开发模式）。
func SetKey(keyStr string) error {
	mu.Lock()
	defer mu.Unlock()

	if keyStr == "" {
		key = nil
		enabled = false
		return nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return fmt.Errorf("invalid key (must be base64-encoded 32-byte key): %w", err)
	}

	if len(keyBytes) != 32 {
		return fmt.Errorf("invalid key length: got %d bytes, want 32", len(keyBytes))
	}

	key = keyBytes
	enabled = true
	return nil
}

// IsEnabled 返回加密是否已启用。
func IsEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

// GenerateKey 生成一个随机的 32 字节密钥（base64 编码），用于初始化配置。
func GenerateKey() (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("generate random key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(keyBytes), nil
}

// ErrCryptoDisabled 在加密未启用时返回的错误。
var ErrCryptoDisabled = fmt.Errorf("crypto is disabled: set IF_CREDENTIAL_KEY to enable encryption")

// Key version prefix for key rotation support.
const keyVersionPrefix = "v1:"

// Encrypt 加密明文数据，返回 base64 编码的密文。
// 如果加密未启用，返回 ErrCryptoDisabled 错误。
func Encrypt(plaintext []byte) (string, error) {
	if !IsEnabled() {
		return "", ErrCryptoDisabled
	}

	mu.RLock()
	k := key
	mu.RUnlock()

	block, err := aes.NewCipher(k)
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
	return keyVersionPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的密文，返回明文。
// 如果加密未启用，返回 ErrCryptoDisabled 错误。
func Decrypt(encoded string) ([]byte, error) {
	if !IsEnabled() {
		return nil, ErrCryptoDisabled
	}

	// Strip version prefix if present
	if len(encoded) > len(keyVersionPrefix) && encoded[:len(keyVersionPrefix)] == keyVersionPrefix {
		encoded = encoded[len(keyVersionPrefix):]
	}

	mu.RLock()
	k := key
	mu.RUnlock()

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(k)
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
	result := make(map[string]any, len(creds))
	for k, v := range creds {
		if isSensitiveField(k) {
			plain, ok := v.(string)
			if !ok {
				result[k] = v
				continue
			}
			enc, err := Encrypt([]byte(plain))
			if err != nil {
				return nil, fmt.Errorf("encrypt field %q: %w", k, err)
			}
			result[k] = "__ENC__:" + enc
		} else {
			result[k] = v
		}
	}
	return result, nil
}

// DecryptCredentials 解密凭证 map 中的敏感字段。
// 识别 "__ENC__:" 前缀并解密。
func DecryptCredentials(creds map[string]any) (map[string]any, error) {
	result := make(map[string]any, len(creds))
	for k, v := range creds {
		str, ok := v.(string)
		if !ok || len(str) < 7 || str[:7] != "__ENC__:" {
			result[k] = v
			continue
		}
		plain, err := Decrypt(str[7:])
		if err != nil {
			return nil, fmt.Errorf("decrypt field %q: %w", k, err)
		}
		result[k] = string(plain)
	}
	return result, nil
}

// isSensitiveField 检查字段名是否为敏感字段（不区分大小写）。
func isSensitiveField(name string) bool {
	for _, s := range sensitiveFields {
		if len(name) == len(s) {
			match := true
			for i := range name {
				if name[i]|0x20 != s[i]|0x20 {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// init 从环境变量 IF_CREDENTIAL_KEY 自动加载密钥（如果存在）。
func init() {
	if keyStr := os.Getenv("IF_CREDENTIAL_KEY"); keyStr != "" {
		_ = SetKey(keyStr)
	}
}
