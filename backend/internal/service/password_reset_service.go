package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/imageforge/imageforge/internal/pkg/errors"
	"github.com/imageforge/imageforge/internal/repository"
)

// PasswordResetService orchestrates the password reset flow:
// 1. Generate a verification code
// 2. Send it via email
// 3. Verify the code and reset the password
type PasswordResetService struct {
	users       *repository.UserRepository
	codes       *repository.ResetCodeStore
	email       *EmailService
	password    *Password
	rateLimiter *repository.RateLimiter
}

// NewPasswordResetService builds a PasswordResetService.
func NewPasswordResetService(
	users *repository.UserRepository,
	codes *repository.ResetCodeStore,
	email *EmailService,
	password *Password,
	rateLimiter *repository.RateLimiter,
) *PasswordResetService {
	return &PasswordResetService{
		users:       users,
		codes:       codes,
		email:       email,
		password:    password,
		rateLimiter: rateLimiter,
	}
}

// SendResetCodeInput is the request to send a reset code.
type SendResetCodeInput struct {
	Email string
}

// SendResetCode generates a 6-digit code and emails it to the user.
// For privacy, it returns success even if the email is not registered.
func (s *PasswordResetService) SendResetCode(ctx context.Context, in SendResetCodeInput) error {
	email := strings.TrimSpace(strings.ToLower(in.Email))

	// Check rate limit
	if s.rateLimiter != nil {
		if err := s.rateLimiter.AllowResetCode(ctx, email); err != nil {
			return errors.New(errors.ErrInvalidRequest, err.Error())
		}
	}

	// Check if user exists — but don't reveal this to the caller.
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// User not found — log for security auditing but return nil to avoid enumeration.
		return nil
	}

	// Generate and store the code.
	code, err := s.codes.Generate(email)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to generate reset code", err)
	}

	// Send the email.
	err = s.email.Send(ctx, Email{
		To:      email,
		Subject: "Your ImageForge Password Reset Code",
		HTML:    resetEmailHTML(code),
		Text:    resetEmailText(code),
	})
	if err != nil {
		// Invalidate the code so it can't be used.
		_ = s.codes.Invalidate(email)
		return errors.Wrap(errors.ErrInternal, "failed to send reset email", err)
	}

	_ = user // used for existence check above
	return nil
}

// ResetPasswordInput is the request to reset a password.
type ResetPasswordInput struct {
	Email       string
	Code        string
	NewPassword string
}

// ResetPassword verifies the code and updates the user's password.
func (s *PasswordResetService) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	email := strings.TrimSpace(strings.ToLower(in.Email))

	// Verify the code.
	valid, err := s.codes.Verify(email, in.Code)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to verify code", err)
	}
	if !valid {
		return errors.New(errors.ErrInvalidRequest, "invalid or expired verification code")
	}

	// Validate new password strength.
	if len(in.NewPassword) < 8 {
		return errors.New(errors.ErrInvalidRequest, "password must be at least 8 characters")
	}

	// Find the user.
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return errors.New(errors.ErrNotFound, "user not found")
	}

	// Increment token version FIRST to invalidate all existing tokens.
	// This must succeed before password change to prevent the window where
	// password is changed but old tokens remain valid.
	if err := s.users.IncrementTokenVersion(ctx, user.ID); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to invalidate existing tokens", err)
	}

	// Hash and update the password.
	hash, err := s.password.Hash(in.NewPassword)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to process password", err)
	}

	if err := s.users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to update password", err)
	}

	// Invalidate any remaining codes for this email.
	_ = s.codes.Invalidate(email)

	return nil
}

// resetEmailHTML returns the HTML body for the reset email.
func resetEmailHTML(code string) string {
	return fmt.Sprintf(`<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 480px; margin: 0 auto; padding: 24px;">
		<h2 style="color: #1a1a1a; margin-bottom: 16px;">Reset Your Password</h2>
		<p style="color: #4a4a4a; line-height: 1.6;">You requested a password reset for your ImageForge account. Use the verification code below:</p>
		<div style="background: #f5f5f5; border-radius: 8px; padding: 20px; text-align: center; margin: 24px 0;">
			<span style="font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #1a1a1a;">%s</span>
		</div>
		<p style="color: #4a4a4a; line-height: 1.6;">This code expires in <strong>5 minutes</strong>.</p>
		<p style="color: #888; font-size: 14px; margin-top: 24px;">If you didn't request this, you can safely ignore this email.</p>
	</div>`, code)
}

// resetEmailText returns the plain-text body for the reset email.
func resetEmailText(code string) string {
	return fmt.Sprintf("Your ImageForge password reset code is: %s\n\nThis code expires in 5 minutes.\n\nIf you didn't request this, you can safely ignore this email.", code)
}
