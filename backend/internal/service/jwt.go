package service

 import (
	"crypto/rand"
	"encoding/base64"
 	"errors"
 	"fmt"
	"time"

 	"github.com/golang-jwt/jwt/v5"
 	"github.com/imageforge/imageforge/internal/config"
 	"github.com/imageforge/imageforge/internal/domain"
 )

// Claims is the ImageForge JWT payload. Minimal: user id + role + token version. No PII, no secrets.
type Claims struct {
	UserID        int64  `json:"uid"`
	Role          string `json:"role"`
	TokenVersion  int    `json:"tv"` // Token version - incremented on password change
	jwt.RegisteredClaims
}

// JWTService signs and parses access tokens.
type JWTService struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTService builds a JWTService from auth config.
// In production mode (release), JWTSecret must be configured or the function returns an error.
func NewJWTService(cfg config.AuthConfig, mode string) (*JWTService, error) {
	ttl := time.Duration(cfg.AccessTokenMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	secret := cfg.JWTSecret
	if secret == "" {
		if mode == "release" || mode == "production" {
			return nil, fmt.Errorf("JWTSecret must be configured in production mode")
		}
		// Only use auto-generated secret in debug/test mode.
		// 32 crypto-random bytes, base64-encoded. Not for production.
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return nil, fmt.Errorf("failed to generate debug JWT secret: %w", err)
		}
		secret = base64.RawURLEncoding.EncodeToString(buf)
 	}
	return &JWTService{secret: []byte(secret), ttl: ttl}, nil
}

// Generate signs a token for the given user.
func (s *JWTService) Generate(userID int64, role string, tokenVersion int) (string, error) {
	if len(s.secret) == 0 {
		return "", errors.New("jwt secret not configured")
	}
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			Issuer:    "imageforge",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// Parse validates a token and returns its claims.
func (s *JWTService) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method: " + t.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// HasAdmin reports whether the role is admin.
func HasAdmin(role string) bool {
	return role == domain.RoleAdmin
}
