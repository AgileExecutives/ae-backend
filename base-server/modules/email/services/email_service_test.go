package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ae-base-server/modules/email/services"
)

// TestEmailService_SendEmail tests basic email sending
func TestEmailService_SendEmail(t *testing.T) {
	// Skip if no SMTP configured
	t.Skip("SMTP configuration required for email tests")

	// This is a placeholder test that would require actual SMTP configuration
	service := &services.EmailService{}

	err := service.SendEmail(
		"test@example.com",
		"Test Subject",
		"<p>Test Body</p>",
		"Test Body",
	)

	// In a real test, you'd mock the SMTP server
	require.NoError(t, err)
}

// TestEmailService_SendVerificationEmail tests verification email template
func TestEmailService_SendVerificationEmail(t *testing.T) {
	t.Skip("SMTP configuration required for email tests")

	service := &services.EmailService{}

	err := service.SendVerificationEmail("test@example.com", "Test User", "https://example.com/verify?token=token123")
	assert.NoError(t, err)
}

// TestEmailService_SendPasswordResetEmail tests password reset email template
func TestEmailService_SendPasswordResetEmail(t *testing.T) {
	t.Skip("SMTP configuration required for email tests")

	service := &services.EmailService{}

	err := service.SendPasswordResetEmail("test@example.com", "Test User", "https://example.com/reset?token=reset-token-456")
	assert.NoError(t, err)
}

// TestEmailService_SendTemplateEmail tests template-based email sending
func TestEmailService_SendTemplateEmail(t *testing.T) {
	t.Skip("SMTP configuration required for email tests")

	service := &services.EmailService{}

	data := services.EmailData{
		Subject:       "Welcome",
		RecipientName: "Test User",
	}

	err := service.SendTemplateEmail("test@example.com", services.TemplateWelcome, data)
	assert.NoError(t, err)
}

// BenchmarkEmailService_SendEmail benchmarks email sending
func BenchmarkEmailService_SendEmail(b *testing.B) {
	b.Skip("SMTP configuration required")

	service := &services.EmailService{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SendEmail(
			"bench@example.com",
			"Benchmark Test",
			"<p>Content</p>",
			"Content",
		)
	}
}
