// Copyright 2024 ImageForge
// TOTP 2FA 服务 — 基于时间的一次性密码。

package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPService 管理 TOTP 2FA 的设置和验证。
type TOTPService struct {
	issuer string
}

// NewTOTPService 创建 TOTP 服务。
func NewTOTPService(issuer string) *TOTPService {
	if issuer == "" {
		issuer = "ImageForge"
	}
	return &TOTPService{issuer: issuer}
}

// SetupResult 返回给客户端的设置信息。
type SetupResult struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"` // Base64 PNG
	BackupCodes []string `json:"backup_codes"`
}

// Setup 为用户生成新的 TOTP 密钥。
// 返回密钥、二维码（Base64 PNG）和备用恢复码。
func (s *TOTPService) Setup(userID int64, email string) (*SetupResult, error) {
	// 生成随机密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20,
		Digits:      6,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP key: %w", err)
	}

	// 生成二维码
	qrBuf, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, qrBuf); err != nil {
		return nil, fmt.Errorf("encode QR PNG: %w", err)
	}
	qrBase64 := base64.StdEncoding.EncodeToString(pngBuf.Bytes())

	// 生成 8 个备用恢复码
	backupCodes := make([]string, 8)
	for i := range backupCodes {
		code := make([]byte, 4)
		if _, err := rand.Read(code); err != nil {
			return nil, fmt.Errorf("generate backup code: %w", err)
		}
		backupCodes[i] = fmt.Sprintf("%04x-%04x", code[:2], code[2:])
	}

	return &SetupResult{
		Secret:      key.Secret(),
		QRCode:      "data:image/png;base64," + qrBase64,
		BackupCodes: backupCodes,
	}, nil
}

// Validate 验证 TOTP 码或备用恢复码。
func (s *TOTPService) Validate(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	// 清理输入
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	return totp.Validate(code, secret)
}

// ValidateBackupCode 验证备用恢复码是否在列表中。
func ValidateBackupCode(backupCodesJSON, code string) (valid bool, remaining []string) {
	if backupCodesJSON == "" {
		return false, nil
	}
	var codes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &codes); err != nil {
		return false, nil
	}
	for i, c := range codes {
		if strings.EqualFold(c, code) {
			// 移除已使用的恢复码
			remaining = append(codes[:i], codes[i+1:]...)
			return true, remaining
		}
	}
	return false, codes
}

// GenerateBackupCodesJSON 生成备用恢复码 JSON。
func GenerateBackupCodesJSON(codes []string) string {
	data, _ := json.Marshal(codes)
	return string(data)
}
