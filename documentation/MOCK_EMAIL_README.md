# Mock Email Testing Feature

## Overview

When `MOCK_EMAIL=true` is set in the environment, the email service will:
1. Log all emails to the console
2. Save all emails to a JSON file at `tmp/mock_emails.json`
3. Expose a `/api/v1/emails/latest-emails` endpoint to retrieve all sent emails

This is useful for:
- Development and testing
- Automated testing of email functionality
- Debugging email content without sending real emails

## Setup

Add to your `.env` file:
```env
MOCK_EMAIL=true
```

## File Structure

The mock emails are stored in `tmp/mock_emails.json` with the following structure:

```json
[
  {
    "to": "user@example.com",
    "time": "2026-01-31T10:30:00Z",
    "subject": "Welcome to Our App",
    "text": "Plain text version of the email",
    "html": "<html>HTML version of the email</html>"
  }
]
```

## Features

### Automatic File Initialization

- The `tmp/mock_emails.json` file is created empty on server start
- Each server restart clears previous emails
- The file is automatically created if it doesn't exist

### API Endpoint

**GET** `/api/v1/emails/latest-emails`

Returns all emails sent during the current server session.

**Response:**
```json
{
  "success": true,
  "message": "Mock emails retrieved successfully",
  "data": [
    {
      "to": "user@example.com",
      "time": "2026-01-31T10:30:00Z",
      "subject": "Welcome",
      "text": "Welcome to our app!",
      "html": "<h1>Welcome</h1><p>Welcome to our app!</p>"
    }
  ]
}
```

**Note:** This endpoint only works when `MOCK_EMAIL=true`. If mock email is disabled, it returns:
```json
{
  "success": false,
  "error": "Mock email not enabled",
  "details": "This endpoint is only available when MOCK_EMAIL=true"
}
```

## Testing

Use the provided test script:

```bash
./base-server/tests/test-mock-emails.sh
```

Or manually test with curl:

```bash
# Send an email
curl -X POST http://localhost:8080/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -d '{
    "to_email": "test@example.com",
    "subject": "Test",
    "body": "Test body"
  }'

# Retrieve all mock emails
curl http://localhost:8080/api/v1/emails/latest-emails | jq '.'
```

## Implementation Details

### Email Service (`modules/email/services/email_service.go`)

- `initMockEmailsFile()` - Creates empty JSON file on service initialization
- `saveMockEmail()` - Appends each email to the JSON file
- `GetLatestEmails()` - Reads and returns all emails from the file

### Email Handler (`modules/email/handlers/mock_emails.go`)

- `GetLatestEmails()` - HTTP handler for the `/latest-emails` endpoint
- Only registered when `MOCK_EMAIL=true`

### File Location

- **Development:** `base-server/tmp/mock_emails.json`
- **Production:** Should not be enabled (use real SMTP)

## Security Considerations

- **DO NOT** enable `MOCK_EMAIL=true` in production
- The endpoint is not authenticated (add auth if needed)
- Email content may contain sensitive information
- The `tmp/` directory should be in `.gitignore`

## Troubleshooting

### Endpoint returns 503 error
- Verify `MOCK_EMAIL=true` in your `.env` file
- Restart the server after changing environment variables

### File not found errors
- The `tmp/` directory is automatically created
- Check file permissions on the `tmp/` directory

### Emails not appearing
- Check server console for error messages
- Verify the email service is using `ProviderMock`
- Check the `tmp/mock_emails.json` file directly
