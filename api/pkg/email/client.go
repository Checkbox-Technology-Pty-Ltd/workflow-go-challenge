package email

import (
	"context"
	"fmt"
	"log/slog"
)

// Client defines the interface for sending emails
type Client interface {
	Send(ctx context.Context, to, subject, body string) error
}

// MockClient is a mock implementation that logs instead of sending
type MockClient struct{}

// NewMockClient creates a new MockClient
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Send logs the email details instead of actually sending
func (c *MockClient) Send(ctx context.Context, to, subject, body string) error {
	slog.Info("mock email sent",
		"to", to,
		"subject", subject,
		"body_length", len(body),
	)
	return nil
}

// SMTPClient sends emails via SMTP (placeholder for real implementation)
type SMTPClient struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewSMTPClient creates a new SMTP client
func NewSMTPClient(cfg SMTPConfig) *SMTPClient {
	return &SMTPClient{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
	}
}

// Send sends an email via SMTP
func (c *SMTPClient) Send(ctx context.Context, to, subject, body string) error {
	// In production, this would use net/smtp or a library like gomail
	// For now, return an error indicating it's not configured
	if c.host == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// TODO: Implement actual SMTP sending
	// msg := gomail.NewMessage()
	// msg.SetHeader("From", c.from)
	// msg.SetHeader("To", to)
	// msg.SetHeader("Subject", subject)
	// msg.SetBody("text/plain", body)

	slog.Info("SMTP email would be sent",
		"to", to,
		"subject", subject,
		"from", c.from,
	)
	return nil
}
