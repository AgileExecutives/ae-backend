package services

import (
	"fmt"
	"os"
	"strings"
)

// EmailService handles email operations
type EmailService struct {
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendContactFormEmail sends a contact form submission to support
func (e *EmailService) SendContactFormEmail(name, email, subject, message, timestamp, source string) error {
	supportEmail := getEnv("SUPPORT_EMAIL", "support@unburdy.de")

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Contact Form Submission</title>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #007bff; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background: #f9f9f9; }
		.field { margin-bottom: 15px; }
		.label { font-weight: bold; color: #555; }
		.value { margin-top: 5px; }
		.footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Contact Form Submission</h1>
		</div>
		<div class="content">
			<h2>New Contact Form Message</h2>
			
			<div class="field">
				<div class="label">Name:</div>
				<div class="value">%s</div>
			</div>
			
			<div class="field">
				<div class="label">Email:</div>
				<div class="value">%s</div>
			</div>
			
			<div class="field">
				<div class="label">Subject:</div>
				<div class="value">%s</div>
			</div>
			
			<div class="field">
				<div class="label">Message:</div>
				<div class="value">%s</div>
			</div>
			
			<div class="field">
				<div class="label">Submitted:</div>
				<div class="value">%s</div>
			</div>
			
			<div class="field">
				<div class="label">Source:</div>
				<div class="value">%s</div>
			</div>
		</div>
		<div class="footer">
			<p>This message was sent from the website contact form.</p>
		</div>
	</div>
</body>
</html>`,
		name, email, subject, strings.ReplaceAll(message, "\n", "<br>"), timestamp, source,
	)

	textBody := fmt.Sprintf(`Contact Form Submission

Name: %s
Email: %s
Subject: %s
Message: %s
Submitted: %s
Source: %s

This message was sent from the website contact form.`,
		name, email, subject, message, timestamp, source,
	)

	emailSubject := fmt.Sprintf("Contact Form: %s", subject)

	// For now, just log the email since we need to implement the actual email sending
	fmt.Printf("Email to be sent to: %s\nSubject: %s\nHTML Body: %s\nText Body: %s\n",
		supportEmail, emailSubject, htmlBody, textBody)

	// TODO: Implement actual email sending (SMTP, SendGrid, etc.)
	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
