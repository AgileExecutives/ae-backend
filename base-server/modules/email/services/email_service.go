package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ae-base-server/pkg/utils"
)

// EmailProvider represents different email service providers
type EmailProvider string

const (
	ProviderSMTP     EmailProvider = "smtp"
	ProviderSendGrid EmailProvider = "sendgrid"
	ProviderMailgun  EmailProvider = "mailgun"
	ProviderMock     EmailProvider = "mock"
)

// EmailTemplate represents different email templates
type EmailTemplate string

const (
	TemplateVerification        EmailTemplate = "verification"
	TemplatePasswordReset       EmailTemplate = "password_reset"
	TemplateWelcome             EmailTemplate = "welcome"
	TemplateInvoice             EmailTemplate = "invoice"
	TemplateAppointment         EmailTemplate = "appointment"
	TemplateBookingConfirmation EmailTemplate = "booking_confirmation"
	TemplateNotification        EmailTemplate = "notification"
)

// EmailData contains data to be passed to email templates
type EmailData struct {
	RecipientName   string
	SenderName      string
	Subject         string
	VerificationURL string
	ResetURL        string
	AppName         string
	SupportEmail    string
	CompanyName     string
	CompanyAddress  string
	CustomData      map[string]interface{}
}

// EmailService handles email operations
type EmailService struct {
	provider  EmailProvider
	templates map[EmailTemplate]*template.Template
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	service := &EmailService{
		templates: make(map[EmailTemplate]*template.Template),
	}

	// Determine provider based on environment
	mockEmail := utils.GetEnv("MOCK_EMAIL", "false")
	if strings.ToLower(mockEmail) == "true" {
		service.provider = ProviderMock
	} else {
		service.provider = ProviderSMTP
	}

	service.loadTemplates()
	return service
}

// loadTemplates loads email templates from files
func (e *EmailService) loadTemplates() {
	templatesDir := utils.GetEnv("TEMPLATES_DIR", "./statics/templates")

	templates := map[EmailTemplate]string{
		TemplateVerification:        "email_verification.html",
		TemplatePasswordReset:       "password_reset.html",
		TemplateWelcome:             "welcome.html",
		TemplateInvoice:             "invoice.html",
		TemplateAppointment:         "appointment.html",
		TemplateBookingConfirmation: "booking_confirmation.html",
		TemplateNotification:        "notification.html",
	}

	for templateType, filename := range templates {
		path := filepath.Join(templatesDir, filename)
		if _, err := os.Stat(path); err == nil {
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				log.Printf("Failed to parse template %s: %v", filename, err)
				continue
			}
			e.templates[templateType] = tmpl
		}
	}
}

// SendEmail sends an email using the configured provider
func (e *EmailService) SendEmail(to, subject, htmlBody, textBody string) error {
	var result error
	switch e.provider {
	case ProviderSMTP:
		result = e.sendSMTP(to, subject, htmlBody, textBody)
	case ProviderMock:
		result = e.sendMock(to, subject, htmlBody, textBody)
	default:
		result = fmt.Errorf("unsupported email provider: %s", e.provider)
	}
	log.Printf("Email sent to %s with subject %q: %v", to, subject, result)
	return result
}

// SendTemplateEmail sends an email using a predefined template
func (e *EmailService) SendTemplateEmail(to string, template EmailTemplate, data EmailData) error {
	// Set default data
	if data.AppName == "" {
		data.AppName = utils.GetEnv("APP_NAME", "Unburdy")
	}
	if data.SupportEmail == "" {
		data.SupportEmail = utils.GetEnv("SUPPORT_EMAIL", "support@unburdy.de")
	}
	if data.CompanyName == "" {
		data.CompanyName = utils.GetEnv("COMPANY_NAME", "Unburdy")
	}

	// Try to use loaded template first
	if tmpl, exists := e.templates[template]; exists {
		var htmlBuffer bytes.Buffer
		err := tmpl.Execute(&htmlBuffer, data)
		if err != nil {
			log.Printf("Failed to execute template %s: %v", template, err)
			return e.sendDefaultTemplate(to, template, data)
		}

		textBody := e.htmlToText(htmlBuffer.String())
		return e.SendEmail(to, data.Subject, htmlBuffer.String(), textBody)
	}

	// Fall back to default template
	return e.sendDefaultTemplate(to, template, data)
}

// SendContactFormEmail sends a contact form submission to support
func (e *EmailService) SendContactFormEmail(name, email, subject, message, timestamp, source string) error {
	supportEmail := utils.GetEnv("SUPPORT_EMAIL", "support@unburdy.de")

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
	return e.SendEmail(supportEmail, emailSubject, htmlBody, textBody)
}

// Convenience methods for specific email types

// SendVerificationEmail sends a verification email
func (e *EmailService) SendVerificationEmail(to, recipientName, verificationURL string) error {
	data := EmailData{
		RecipientName:   recipientName,
		Subject:         "Please verify your email address",
		VerificationURL: verificationURL,
	}
	return e.SendTemplateEmail(to, TemplateVerification, data)
}

// SendPasswordResetEmail sends a password reset email
func (e *EmailService) SendPasswordResetEmail(to, recipientName, resetURL string) error {
	data := EmailData{
		RecipientName: recipientName,
		Subject:       "Password Reset Request",
		ResetURL:      resetURL,
	}
	return e.SendTemplateEmail(to, TemplatePasswordReset, data)
}

// SendWelcomeEmail sends a welcome email
func (e *EmailService) SendWelcomeEmail(to, recipientName string) error {
	data := EmailData{
		RecipientName: recipientName,
		Subject:       "Welcome to " + utils.GetEnv("APP_NAME", "Unburdy"),
	}
	return e.SendTemplateEmail(to, TemplateWelcome, data)
}

// SendNotificationEmail sends a notification email
func (e *EmailService) SendNotificationEmail(to, recipientName, subject, message string, customData map[string]interface{}) error {
	data := EmailData{
		RecipientName: recipientName,
		Subject:       subject,
		CustomData:    customData,
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>%s</title>
</head>
<body>
    <h1>%s</h1>
    <p>Hello %s,</p>
    <p>%s</p>
    <p>Best regards,<br>%s Team</p>
</body>
</html>`, subject, subject, recipientName, message, data.AppName)

	textBody := fmt.Sprintf("%s\n\nHello %s,\n\n%s\n\nBest regards,\n%s Team",
		subject, recipientName, message, data.AppName)

	return e.SendEmail(to, subject, htmlBody, textBody)
}

// Private helper methods

// sendSMTP sends email via SMTP
func (e *EmailService) sendSMTP(to, subject, htmlBody, textBody string) error {
	smtpHost := utils.GetEnv("SMTP_HOST", "")
	smtpPort := utils.GetEnv("SMTP_PORT", "587")
	smtpUser := utils.GetEnv("SMTP_USER", "")
	smtpPassword := utils.GetEnv("SMTP_PASSWORD", "")
	smtpFrom := utils.GetEnv("SMTP_FROM", smtpUser)
	authEnabled := strings.ToLower(utils.GetEnv("SMTP_AUTH_ENABLED", "true")) == "true"

	if smtpHost == "" {
		return fmt.Errorf("SMTP configuration missing")
	}

	var auth smtp.Auth
	if authEnabled && smtpUser != "" && smtpPassword != "" {
		auth = smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	}

	message := e.composeMessage(to, subject, htmlBody, textBody)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	// Use STARTTLS for port 587, direct TLS for port 465, plain for other ports
	if smtpPort == "587" {
		return e.sendSMTPSTARTTLS(addr, auth, smtpFrom, []string{to}, []byte(message))
	} else if smtpPort == "465" {
		return e.sendSMTPTLS(addr, auth, smtpFrom, []string{to}, []byte(message))
	}

	return smtp.SendMail(addr, auth, smtpFrom, []string{to}, []byte(message))
}

// sendSMTPTLS sends email via SMTP with TLS
func (e *EmailService) sendSMTPTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsEnabled := strings.ToLower(utils.GetEnv("SMTP_TLS_ENABLED", "true")) == "true"

	if !tlsEnabled {
		// Fall back to plain SMTP if TLS is disabled
		return smtp.SendMail(addr, auth, from, to, msg)
	}

	// Create TLS connection
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         strings.Split(addr, ":")[0],
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, tlsConfig.ServerName)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %v", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", addr, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	return nil
}

// sendSMTPSTARTTLS sends email via SMTP with STARTTLS (for port 587)
func (e *EmailService) sendSMTPSTARTTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsEnabled := strings.ToLower(utils.GetEnv("SMTP_TLS_ENABLED", "true")) == "true"

	if !tlsEnabled {
		// Fall back to plain SMTP if TLS is disabled
		return smtp.SendMail(addr, auth, from, to, msg)
	}

	// Connect to SMTP server without TLS
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Create SMTP client
	host := strings.Split(addr, ":")[0]
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %v", err)
		}
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// Set recipients
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", addr, err)
		}
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	return nil
}

// sendMock logs email content instead of sending (for development)
func (e *EmailService) sendMock(to, subject, htmlBody, textBody string) error {
	from := utils.GetEnv("SMTP_FROM", "Unburdy <no-reply@unburdy.de>")

	emailType := e.detectEmailType(subject, htmlBody)

	fmt.Println("\n================================================================================")
	fmt.Println("🚀 MOCK EMAIL SERVICE - EMAIL DATA")
	fmt.Println("================================================================================")
	fmt.Printf("📧 To: %s\n", to)
	fmt.Printf("📤 From: %s\n", from)
	fmt.Printf("📋 Subject: %s\n", subject)
	fmt.Println("--------------------------------------------------------------------------------")

	e.displayEmailTypeInfo(emailType, htmlBody)

	fmt.Println("📄 HTML CONTENT:")
	fmt.Println("----------------------------------------")
	fmt.Println(htmlBody)
	fmt.Println("----------------------------------------")
	fmt.Println("📝 TEXT CONTENT:")
	fmt.Println("----------------------------------------")
	fmt.Println(textBody)
	fmt.Println("----------------------------------------")
	fmt.Println("✅ Email processing completed successfully (Mock Mode)")
	fmt.Println("================================================================================")

	return nil
}

// detectEmailType analyzes the email content to determine its type
func (e *EmailService) detectEmailType(subject, htmlBody string) string {
	subjectLower := strings.ToLower(subject)
	bodyLower := strings.ToLower(htmlBody)

	if strings.Contains(subjectLower, "verification") || strings.Contains(bodyLower, "verify") {
		return "📧 EMAIL VERIFICATION"
	} else if strings.Contains(subjectLower, "reset") || strings.Contains(bodyLower, "reset") {
		return "🔑 PASSWORD RESET"
	} else if strings.Contains(subjectLower, "welcome") || strings.Contains(bodyLower, "welcome") {
		return "👋 WELCOME EMAIL"
	} else if strings.Contains(subjectLower, "invoice") || strings.Contains(bodyLower, "invoice") {
		return "🧾 INVOICE"
	} else if strings.Contains(subjectLower, "appointment") || strings.Contains(bodyLower, "appointment") {
		return "📅 APPOINTMENT"
	} else if strings.Contains(subjectLower, "contact") || strings.Contains(bodyLower, "contact form") {
		return "📝 CONTACT FORM SUBMISSION"
	} else if strings.Contains(subjectLower, "notification") {
		return "🔔 NOTIFICATION"
	}
	return "📧 GENERAL EMAIL"
}

// displayEmailTypeInfo shows contextual information based on email type
func (e *EmailService) displayEmailTypeInfo(emailType, htmlBody string) {
	fmt.Printf("🏷️  EMAIL TYPE: %s\n", emailType)
	fmt.Println("--------------------------------------------------------------------------------")

	switch {
	case strings.Contains(emailType, "VERIFICATION"):
		link := e.extractLink(htmlBody, []string{"verify", "confirmation", "activate"})
		if link != "" {
			fmt.Printf("🔗 VERIFICATION LINK: %s\n", link)
		}
		fmt.Println("💡 INFO: User account verification email")
	case strings.Contains(emailType, "PASSWORD"):
		link := e.extractLink(htmlBody, []string{"reset", "password"})
		if link != "" {
			fmt.Printf("🔗 RESET LINK: %s\n", link)
		}
		fmt.Println("💡 INFO: Password reset request")
	case strings.Contains(emailType, "CONTACT"):
		fmt.Println("💡 INFO: Contact form submission received and forwarded")
	case strings.Contains(emailType, "INVOICE"):
		fmt.Println("💡 INFO: Invoice generated and sent to customer")
	default:
		fmt.Println("💡 INFO: General email communication")
	}
	fmt.Println("--------------------------------------------------------------------------------")
}

// extractLink finds relevant links in email HTML
func (e *EmailService) extractLink(htmlBody string, keywords []string) string {
	// Look for href attributes containing keywords
	re := regexp.MustCompile(`href=["']([^"']*(?:` + strings.Join(keywords, "|") + `)[^"']*)["']`)
	matches := re.FindStringSubmatch(htmlBody)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// composeMessage creates the full email message with headers
func (e *EmailService) composeMessage(to, subject, htmlBody, textBody string) string {
	fromEmail := utils.GetEnv("SMTP_FROM", "no-reply@unburdy.de")
	fromName := utils.GetEnv("SMTP_FROM_NAME", "")

	// Format the From header with name if provided
	var from string
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	} else {
		from = fromEmail
	}

	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n")
	buf.WriteString("\r\n")

	// Text part
	buf.WriteString("--boundary123\r\n")
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(textBody)
	buf.WriteString("\r\n\r\n")

	// HTML part
	buf.WriteString("--boundary123\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n\r\n")

	// End boundary
	buf.WriteString("--boundary123--\r\n")

	return buf.String()
}

// sendDefaultTemplate sends email using built-in default templates
func (e *EmailService) sendDefaultTemplate(to string, template EmailTemplate, data EmailData) error {
	var htmlBody, textBody string

	switch template {
	case TemplateVerification:
		htmlBody = e.getDefaultVerificationTemplate(data)
	case TemplatePasswordReset:
		htmlBody = e.getDefaultPasswordResetTemplate(data)
	case TemplateWelcome:
		htmlBody = e.getDefaultWelcomeTemplate(data)
	default:
		return fmt.Errorf("unsupported template: %s", template)
	}

	textBody = e.htmlToText(htmlBody)
	return e.SendEmail(to, data.Subject, htmlBody, textBody)
}

// htmlToText converts HTML to plain text (simplified)
func (e *EmailService) htmlToText(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	// Clean up whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Clean up line breaks
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.TrimSpace(text)

	return text
}

// Default template methods

// getDefaultVerificationTemplate returns a default verification email template
func (e *EmailService) getDefaultVerificationTemplate(data EmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Email Verification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .button { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
        </div>
        <div class="content">
            <p>Hello %s,</p>
            <p>Thank you for signing up! Please click the button below to verify your email address:</p>
            <p><a href="%s" class="button">Verify Email</a></p>
            <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
            <p><a href="%s">%s</a></p>
        </div>
        <div class="footer">
            <p>This email was sent by %s. If you didn't create an account, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>`, data.AppName, data.RecipientName, data.VerificationURL, data.VerificationURL, data.VerificationURL, data.CompanyName)
}

// getDefaultPasswordResetTemplate returns a default password reset email template
func (e *EmailService) getDefaultPasswordResetTemplate(data EmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Password Reset</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .button { background: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset</h1>
        </div>
        <div class="content">
            <p>Hello %s,</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p><a href="%s" class="button">Reset Password</a></p>
            <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
            <p><a href="%s">%s</a></p>
            <p>If you didn't request a password reset, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>This email was sent by %s.</p>
        </div>
    </div>
</body>
</html>`, data.RecipientName, data.ResetURL, data.ResetURL, data.ResetURL, data.CompanyName)
}

// getDefaultWelcomeTemplate returns a default welcome email template
func (e *EmailService) getDefaultWelcomeTemplate(data EmailData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Welcome</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to %s!</h1>
        </div>
        <div class="content">
            <p>Hello %s,</p>
            <p>Welcome to %s! We're excited to have you on board.</p>
            <p>You can now access all the features of your account. If you have any questions, don't hesitate to contact our support team at %s.</p>
        </div>
        <div class="footer">
            <p>Thank you for choosing %s!</p>
        </div>
    </div>
</body>
</html>`, data.AppName, data.RecipientName, data.AppName, data.SupportEmail, data.CompanyName)
}
