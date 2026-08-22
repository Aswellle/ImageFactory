package service

import (
	"fmt"

	"github.com/imageforge/imageforge/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// Password provides bcrypt hashing. Plaintext passwords are never stored.
type Password struct {
	cost int
}

// NewPassword builds a Password hasher from auth config.
func NewPassword(cfg config.AuthConfig) *Password {
	cost := cfg.BcryptCost
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &Password{cost: cost}
}

// Hash returns a bcrypt hash of the plaintext password.
func (p *Password) Hash(plaintext string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), p.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(b), nil
}

// Verify reports whether plaintext matches the stored hash.
func (p *Password) Verify(plaintext, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}
