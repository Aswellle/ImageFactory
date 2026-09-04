package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/imageforge/imageforge/internal/config"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

// EmailService sends transactional emails (verification codes, notifications).
// It abstracts the delivery backend so handlers don't depend on SendGrid directly.
type EmailService struct {
	cfg    config.EmailConfig
	log    *zap.Logger
	client *sendgrid.Client
}

// emailRegex is a simple email validation pattern.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmailService builds an EmailService. When cfg.Provider is "console", emails
// are logged instead of sent — useful for local development.
func NewEmailService(cfg config.EmailConfig, log *zap.Logger) *EmailService {
	// Validate sender email format
	if cfg.Provider == "sendgrid" && !emailRegex.MatchString(cfg.FromEmail) {
		log.Warn("invalid from_email configured, emails may fail", zap.String("from_email", cfg.FromEmail))
	}

	svc := &EmailService{cfg: cfg, log: log}
	if cfg.Provider == "sendgrid" && cfg.SendGridAPIKey != "" {
		svc.client = sendgrid.NewSendClient(cfg.SendGridAPIKey)
	}
	return svc
}

// Email represents a simple transactional email.
type Email struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// Send delivers an email using the configured backend.
func (s *EmailService) Send(ctx context.Context, email Email) error {
	switch s.cfg.Provider {
	case "sendgrid":
		return s.sendSendGrid(ctx, email)
	case "console":
		return s.sendConsole(email)
	default:
		return fmt.Errorf("email provider %q not configured", s.cfg.Provider)
	}
}

// sendSendGrid delivers via SendGrid's API.
func (s *EmailService) sendSendGrid(ctx context.Context, email Email) error {
	if s.client == nil {
		return fmt.Errorf("sendgrid client not initialized — check SENDGRID_API_KEY")
	}

	from := mail.NewEmail(s.cfg.FromName, s.cfg.FromEmail)
	to := mail.NewEmail("", email.To)
	plain := email.Text
	if plain == "" {
		plain = stripHTML(email.HTML)
	}
	html := email.HTML
	if html == "" {
		html = "<p>" + plain + "</p>"
	}

	message := mail.NewSingleEmail(from, email.Subject, to, plain, html)

	resp, err := s.client.SendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("sendgrid send failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid API error: status=%d body=%s", resp.StatusCode, resp.Body)
	}

	s.log.Info("email sent",
		zap.String("to", email.To),
		zap.String("subject", email.Subject),
		zap.Int("status", resp.StatusCode),
	)
	return nil
}

// sendConsole logs the email instead of sending — for local development.
func (s *EmailService) sendConsole(email Email) error {
	s.log.Info("=== EMAIL (console mode) ===",
		zap.String("to", email.To),
		zap.String("subject", email.Subject),
	)
	s.log.Info("--- HTML ---", zap.String("body", email.HTML))
	s.log.Info("--- Text ---", zap.String("body", email.Text))
	s.log.Info("=== END EMAIL ===")
	return nil
}

// stripHTML is a minimal HTML-to-text fallback for when only HTML is provided.
func stripHTML(html string) string {
	// Simple tag stripping — sufficient for our transactional emails.
	// For production, consider bluemonday or goquery.
	result := ""
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result += string(r)
		}
	}
	return result
}
