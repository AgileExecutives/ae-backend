# Booking Link API

## Overview
The booking link endpoint creates self-contained JWT tokens that allow clients to book appointments without authentication. These tokens encode all necessary information and are signed with HMAC-SHA256 for security.

## Endpoint

### POST /booking/link

Creates a booking link token for a client to book appointments.

**Authentication Required:** Yes (Bearer Token)

**Request Body:**
```json
{
  "template_id": 1,
  "client_id": 123,
  "token_purpose": "one-time-booking-link"
}
```

**Request Fields:**
- `template_id` (uint, required): The ID of the booking template to use
- `client_id` (uint, required): The ID of the client who will use this link
- `token_purpose` (string, required): Either `"one-time-booking-link"` or `"permanent-booking-link"`

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Booking link created successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "url": "/booking/book?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "purpose": "one-time-booking-link",
    "expires_at": "2024-01-16T19:30:00Z",
    "created_at": "2024-01-15T19:30:00Z"
  },
  "error": null
}
```

**Response Fields:**
- `token`: The self-contained JWT token
- `url`: The booking URL with the token as a query parameter
- `purpose`: The token purpose (one-time or permanent)
- `expires_at`: Expiration time (only for one-time links, null for permanent)
- `created_at`: Token creation timestamp

## Token Types

### One-Time Booking Link
- Purpose: `"one-time-booking-link"`
- Expires after 24 hours
- Intended for single use (validation on use should be implemented)

### Permanent Booking Link
- Purpose: `"permanent-booking-link"`
- No expiration
- Can be used multiple times

## JWT Token Structure

The token is a standard JWT with three parts: header, payload, and signature.

### Header
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

### Payload (Claims)
```json
{
  "tenant_id": 1,
  "user_id": 5,
  "calendar_id": 10,
  "template_id": 1,
  "client_id": 123,
  "purpose": "one-time-booking-link",
  "iat": 1705349400,
  "exp": 1705435800
}
```

**Claims:**
- `tenant_id`: The tenant ID (for multi-tenancy)
- `user_id`: The user who owns the calendar
- `calendar_id`: The calendar ID for the booking
- `template_id`: The booking template ID
- `client_id`: The client who will book appointments
- `purpose`: Token purpose (one-time or permanent)
- `iat`: Issued at timestamp (Unix time)
- `exp`: Expiration timestamp (Unix time, only for one-time links)

### Signature
HMAC-SHA256 signature using a secret key configured in the booking module.

## Security

- **JWT Signing:** All tokens are signed with HMAC-SHA256
- **Self-Contained:** No database storage required; all data is in the token
- **Validation:** The `ValidateBookingLink` service method verifies:
  - Token format (3 parts: header.payload.signature)
  - Signature authenticity
  - Expiration (for one-time links)
- **Secret Key:** Configured via the module initialization (production should use environment variable)

## Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "message": "Invalid request data",
  "data": null,
  "error": "validation error details"
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "message": "Tenant information required",
  "data": null,
  "error": ""
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "booking template not found",
  "data": null,
  "error": ""
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "",
  "data": null,
  "error": "error details"
}
```

## Example Usage

### cURL Example
```bash
curl -X POST https://api.example.com/booking/link \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "one-time-booking-link"
  }'
```

### JavaScript Example
```javascript
const response = await fetch('https://api.example.com/booking/link', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_AUTH_TOKEN',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    template_id: 1,
    client_id: 123,
    token_purpose: 'one-time-booking-link'
  })
});

const data = await response.json();
console.log('Booking URL:', data.data.url);
console.log('Token:', data.data.token);
```

## Implementation Details

### Files
- **Entities:** `/modules/booking/entities/booking_link.go`
  - `TokenPurpose` enum
  - `BookingLinkClaims` struct
  - `CreateBookingLinkRequest` struct
  - `BookingLinkResponse` struct

- **Service:** `/modules/booking/services/booking_link_service.go`
  - `GenerateBookingLink()`: Creates JWT token
  - `ValidateBookingLink()`: Validates and decodes token
  - `generateToken()`: Internal JWT generation
  - `createSignature()`: HMAC-SHA256 signing

- **Handler:** `/modules/booking/handlers/booking_handler.go`
  - `CreateBookingLink()`: HTTP handler

- **Routes:** `/modules/booking/routes/routes.go`
  - POST /booking/link endpoint registration

### Token Validation Example
```go
// In a booking handler that accepts tokens
token := c.Query("token")
claims, err := bookingLinkService.ValidateBookingLink(token)
if err != nil {
    // Handle invalid token
    return
}

// Use the claims
tenantID := claims.TenantID
userID := claims.UserID
calendarID := claims.CalendarID
templateID := claims.TemplateID
clientID := claims.ClientID
```

## Production Considerations

1. **Secret Key:** Configure via environment variable
   - Currently uses default: `"booking-link-secret-key-change-in-production"`
   - Should be a strong, randomly generated secret in production

2. **One-Time Link Usage:** Implement usage tracking
   - Current implementation validates expiration but not single-use
   - Consider adding a used token blacklist or tracking in database

3. **Token Revocation:** For permanent links
   - Implement a token blacklist for revoked permanent links
   - Store revoked token signatures in database

4. **Rate Limiting:** Protect the endpoint
   - Implement rate limiting to prevent abuse
   - Consider per-user or per-tenant limits

5. **HTTPS Only:** Ensure tokens are transmitted securely
   - Always use HTTPS in production
   - Never expose tokens in logs or URLs that might be cached

## Swagger/OpenAPI

The endpoint is fully documented in Swagger with:
- **Operation ID:** `createBookingLink`
- **Tags:** `booking-templates`
- **Security:** Bearer authentication required
- Complete request/response schemas
