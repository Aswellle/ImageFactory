package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/imageforge/imageforge/internal/config"
	"github.com/imageforge/imageforge/internal/domain"
)

// Claims is the ImageForge JWT payload. Minimal: user id + role. No PII, no secrets.
type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService signs and parses access tokens.
type JWTService struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTService builds a JWTService from auth config.
func NewJWTService(cfg config.AuthConfig) *JWTService {
	ttl := time.Duration(cfg.AccessTokenMinutes) * time.Minute
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	secret := cfg.JWTSecret
	if secret == "" {
		secret = "imageforge-dev-secret-change-in-production-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return &JWTService{secret: []byte(secret), ttl: ttl}
}

// Generate signs a token for the given user.
func (s *JWTService) Generate(userID int64, role string) (string, error) {
	if len(s.secret) == 0 {
		return "", errors.New("jwt secret not configured")
	}
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
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
