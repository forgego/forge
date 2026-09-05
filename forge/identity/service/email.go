package service

import (
	"context"
	"log"
)

// EmailSender defines the interface for sending emails
type EmailSender interface {
	// SendVerificationEmail sends an email verification message
	SendVerificationEmail(ctx context.Context, to, token string) error
}

// LogEmailSender is a mock implementation that logs emails instead of sending them
// Useful for development and testing
type LogEmailSender struct{}

// SendVerificationEmail logs the verification email details
func (s *LogEmailSender) SendVerificationEmail(ctx context.Context, to, token string) error {
	log.Printf("[EMAIL] Verification email for %s: token=%s", to, token)
	return nil
}

// NoOpEmailSender is an email sender that does nothing
// Useful when email functionality is disabled
type NoOpEmailSender struct{}

// SendVerificationEmail does nothing
func (s *NoOpEmailSender) SendVerificationEmail(ctx context.Context, to, token string) error {
	return nil
}
