# Client Registration Workflow

## Overview

The client registration system allows organizations to generate secure, permanent registration tokens that enable prospective clients to self-register on a waiting list with email verification. This system is organization-scoped with automatic blacklisting of old tokens when new ones are generated.

## Architecture

### Key Components

1. **Registration Token System**: Organization-based permanent tokens with automatic blacklisting
2. **Public Registration Endpoint**: No authentication required for clients to register
3. **Email Verification**: JWT-based email verification using base-server token service
4. **Waiting List Management**: Clients start with `status="waiting"` and `email_verified=false`

### Entities

#### RegistrationToken
```go
type RegistrationToken struct {
    ID             uint
    OrganizationID uint      // Scope token to specific organization
    TenantID       uint      // Multi-tenant isolation
    Token          string    // Secure random token (base64 encoded)
    Email          string    // Optional pre-filled email
    UsedCount      int       // Track how many clients registered with this token
    Blacklisted    bool      // Revoked status (set when new token generated)
    CreatedBy      uint      // User who generated the token
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

#### ClientRegistrationRequest
```go
type ClientRegistrationRequest struct {
    FirstName            string  `json:"first_name" binding:"required"`
    LastName             string  `json:"last_name" binding:"required"`
    Email                string  `json:"email" binding:"required,email"`
    PhoneNumber          *string `json:"phone_number,omitempty"`
    MobileNumber         *string `json:"mobile_number,omitempty"`
    Street               *string `json:"street,omitempty"`
    Zip                  *string `json:"zip,omitempty"`
    City                 *string `json:"city,omitempty"`
    Country              *string `json:"country,omitempty"`
    DateOfBirth          *string `json:"date_of_birth,omitempty"`
    Nationality          *string `json:"nationality,omitempty"`
    PreferredLanguage    *string `json:"preferred_language,omitempty"`
    MaritalStatus        *string `json:"marital_status,omitempty"`
    Gender               *string `json:"gender,omitempty"`
    Profession           *string `json:"profession,omitempty"`
    Notes                *string `json:"notes,omitempty"`
}
```

## Complete Workflow

### Step 1: Generate Registration Token (Admin)

**Endpoint**: `GET /api/v1/clients/registrationtoken`

**Authentication**: Required (Bearer token)

**Authorization**: Admin users only

**Request**:
```http
GET /api/v1/clients/registrationtoken?organization_id=123&email=jane@example.com
Authorization: Bearer <admin_jwt_token>
```

**Query Parameters**:
- `organization_id` (required): The organization ID for which to generate the token
- `email` (optional): Pre-fill email address for the registration

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Registration token generated successfully. Previous tokens have been revoked.",
  "data": {
    "token": "Xy7kP9mN2sQ4vB8zA1cR6fT3gH5jL0",
    "email": "jane@example.com",
    "organization_id": 123
  }
}
```

**Behavior**:
1. Validates admin authentication
2. **Blacklists all existing tokens** for the specified organization (sets `blacklisted=true`)
3. Generates new secure 30-character token using `crypto/rand`
4. Stores token with organization scope (permanent, no expiration)
5. Returns token to admin for distribution

**Swagger ID**: `generateRegistrationToken`

---

### Step 2: Client Self-Registration

**Endpoint**: `POST /api/v1/clients/registration/{token}`

**Authentication**: None (public endpoint)

**Validation**: Registration token must be valid (not blacklisted)

**Request**:
```http
POST /api/v1/clients/registration/Xy7kP9mN2sQ4vB8zA1cR6fT3gH5jL0
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Smith",
  "email": "jane@example.com",
  "phone_number": "+49 30 12345678",
  "mobile_number": "+49 170 9876543",
  "street": "Hauptstraße 123",
  "zip": "10115",
  "city": "Berlin",
  "country": "Germany",
  "date_of_birth": "1990-05-15",
  "nationality": "German",
  "preferred_language": "de",
  "marital_status": "single",
  "gender": "female",
  "profession": "Software Engineer",
  "notes": "Interested in therapy sessions"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "message": "Client registered successfully. Please check your email to verify your address.",
  "data": {
    "id": 456,
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane@example.com",
    "email_verified": false,
    "status": "waiting",
    "phone_number": "+49 30 12345678",
    "mobile_number": "+49 170 9876543",
    "street": "Hauptstraße 123",
    "zip": "10115",
    "city": "Berlin",
    "country": "Germany",
    "date_of_birth": "1990-05-15",
    "nationality": "German",
    "preferred_language": "de",
    "marital_status": "single",
    "gender": "female",
    "profession": "Software Engineer",
    "notes": "Interested in therapy sessions",
    "organization_id": 123,
    "created_at": "2026-01-27T22:00:00Z",
    "updated_at": "2026-01-27T22:00:00Z"
  }
}
```

**Behavior**:
1. Validates registration token (checks if blacklisted)
2. Verifies email in request matches token email (if token has pre-filled email)
3. Creates client with:
   - `status = "waiting"`
   - `email_verified = false`
   - Organization ID from token
   - Tenant ID from token
4. Increments `used_count` on registration token
5. Generates email verification JWT token (24-hour expiry)
6. **TODO**: Send verification email to client
7. Returns client data with verification instructions

**Error Cases**:
- `400 Bad Request`: Invalid token, blacklisted token, or email mismatch
- `500 Internal Server Error`: Database or service errors

**Swagger ID**: `registerClient`

---

### Step 3: Email Verification

**Endpoint**: `POST /api/v1/clients/emailverification/{token}`

**Authentication**: None (public endpoint)

**Validation**: JWT token must be valid and not expired

**Request**:
```http
POST /api/v1/clients/emailverification/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response** (200 OK):
```json
{
  "status": "success",
  "message": "Email verified successfully",
  "data": {
    "id": 456,
    "first_name": "Jane",
    "last_name": "Smith",
    "email": "jane@example.com",
    "email_verified": true,
    "status": "waiting",
    "organization_id": 123,
    ...
  }
}
```

**Behavior**:
1. Validates JWT verification token using `baseAuth.ValidateVerificationToken`
2. Extracts `client_id` claim from token
3. Retrieves client from database
4. Updates `email_verified = true`
5. Returns updated client data

**Error Cases**:
- `400 Bad Request`: Invalid or expired token
- `404 Not Found`: Client not found
- `500 Internal Server Error`: Database errors

**Swagger ID**: `verifyClientEmail`

---

## Token Security

### Registration Token Generation
```go
// Generate cryptographically secure random token
tokenBytes := make([]byte, 32)
_, err := rand.Read(tokenBytes)
token := base64.URLEncoding.EncodeToString(tokenBytes)[:30]
```

**Properties**:
- 32 random bytes from `crypto/rand` (cryptographically secure)
- Base64 URL-safe encoding
- Truncated to 30 characters
- High entropy (approximately 180 bits)

### Email Verification Token
Uses base-server JWT token service with:
- **Algorithm**: HMAC-SHA256
- **Expiry**: 24 hours
- **Claims**: `client_id`
- **Secret**: From environment configuration

---

## Token Lifecycle Management

### Registration Token States

1. **Active**: `blacklisted = false`, available for client registrations
2. **Blacklisted**: `blacklisted = true`, rejected at validation

### Blacklisting Behavior

When a new registration token is generated for an organization:

```go
// Step 1: Blacklist all existing tokens for organization
UPDATE registration_tokens 
SET blacklisted = true 
WHERE organization_id = ? AND blacklisted = false

// Step 2: Create new token
INSERT INTO registration_tokens (organization_id, token, ...)
```

**Guarantees**:
- Organization always has exactly ONE valid token
- Old tokens are immediately invalidated
- No race conditions (blacklisting happens in transaction)
- Revocation is instant and permanent

### Usage Tracking

Each successful registration increments `used_count`:

```go
UPDATE registration_tokens 
SET used_count = used_count + 1 
WHERE id = ?
```

**Use Cases**:
- Analytics: How many clients registered per token
- Monitoring: Detect unusual registration patterns
- Auditing: Track token effectiveness

---

## Database Schema

```sql
CREATE TABLE registration_tokens (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    tenant_id INTEGER NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255),
    used_count INTEGER DEFAULT 0,
    blacklisted BOOLEAN DEFAULT false,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (organization_id) REFERENCES organizations(id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    
    INDEX idx_organization_blacklisted (organization_id, blacklisted),
    INDEX idx_token (token)
);

-- Client table includes email_verified field
ALTER TABLE clients ADD COLUMN email_verified BOOLEAN DEFAULT false;
```

---

## API Endpoints Summary

| Endpoint | Method | Auth | Purpose | Swagger ID |
|----------|--------|------|---------|-----------|
| `/api/v1/clients/registrationtoken` | GET | Required | Generate registration token | `generateRegistrationToken` |
| `/api/v1/clients/registration/{token}` | POST | None | Client self-registration | `registerClient` |
| `/api/v1/clients/emailverification/{token}` | POST | None | Verify email address | `verifyClientEmail` |

---

## Request/Response Types

### RegistrationTokenResponse
```go
type RegistrationTokenResponse struct {
    Token          string `json:"token" example:"Xy7kP9mN2sQ4vB8zA1cR6fT3gH5jL0"`
    Email          string `json:"email" example:"jane@example.com"`
    OrganizationID uint   `json:"organization_id" example:"123"`
}
```

**Swagger Schema**: `entities.RegistrationTokenResponse`

### ClientRegistrationRequest
```go
type ClientRegistrationRequest struct {
    FirstName         string  `json:"first_name" binding:"required" example:"Jane"`
    LastName          string  `json:"last_name" binding:"required" example:"Smith"`
    Email             string  `json:"email" binding:"required,email" example:"jane@example.com"`
    PhoneNumber       *string `json:"phone_number,omitempty" example:"+49 30 12345678"`
    MobileNumber      *string `json:"mobile_number,omitempty" example:"+49 170 9876543"`
    Street            *string `json:"street,omitempty" example:"Hauptstraße 123"`
    Zip               *string `json:"zip,omitempty" example:"10115"`
    City              *string `json:"city,omitempty" example:"Berlin"`
    Country           *string `json:"country,omitempty" example:"Germany"`
    DateOfBirth       *string `json:"date_of_birth,omitempty" example:"1990-05-15"`
    Nationality       *string `json:"nationality,omitempty" example:"German"`
    PreferredLanguage *string `json:"preferred_language,omitempty" example:"de"`
    MaritalStatus     *string `json:"marital_status,omitempty" example:"single"`
    Gender            *string `json:"gender,omitempty" example:"female"`
    Profession        *string `json:"profession,omitempty" example:"Software Engineer"`
    Notes             *string `json:"notes,omitempty" example:"Interested in therapy sessions"`
}
```

**Swagger Schema**: `entities.ClientRegistrationRequest`

### ClientResponse
```go
type ClientResponse struct {
    ID                uint       `json:"id"`
    FirstName         string     `json:"first_name"`
    LastName          string     `json:"last_name"`
    Email             string     `json:"email"`
    EmailVerified     bool       `json:"email_verified"`
    Status            string     `json:"status"`
    OrganizationID    *uint      `json:"organization_id,omitempty"`
    // ... all other client fields
}
```

**Swagger Schema**: `entities.ClientResponse`

---

## Error Handling

### Common Error Response Format
```json
{
  "status": "error",
  "message": "Brief error description",
  "error": "Detailed error message"
}
```

**Swagger Schema**: `models.ErrorResponse`

### Error Scenarios

#### Invalid Registration Token
```json
{
  "status": "error",
  "message": "Registration failed",
  "error": "invalid registration token"
}
```

#### Blacklisted Token
```json
{
  "status": "error",
  "message": "Registration failed",
  "error": "registration token has been revoked"
}
```

#### Email Mismatch
```json
{
  "status": "error",
  "message": "Registration failed",
  "error": "email does not match registration token"
}
```

#### Invalid Verification Token
```json
{
  "status": "error",
  "message": "Verification failed",
  "error": "invalid or expired token"
}
```

---

## Integration Points

### Base-Server Authentication Service

The registration workflow integrates with the base-server authentication service for email verification:

```go
import "github.com/ae-base-server/pkg/auth"

// Generate 24-hour JWT token
token, err := auth.GenerateVerificationToken(clientID, 24*time.Hour)

// Validate JWT token and extract client_id
claims, err := auth.ValidateVerificationToken(token)
clientID := claims.ClientID
```

### Email Service (Future)

Currently, email verification tokens are logged. Production implementation should:

1. Send welcome email with verification link
2. Include verification token in URL
3. Use email templates with branding
4. Track email delivery status

**Example Integration**:
```go
verificationURL := fmt.Sprintf("https://app.example.com/verify?token=%s", verificationToken)
err = emailService.SendVerificationEmail(client.Email, verificationURL)
```

---

## Security Considerations

### Token Security

1. **Registration Tokens**:
   - Cryptographically secure random generation
   - 180 bits of entropy (30 characters base64)
   - Organization-scoped to prevent cross-organization abuse
   - Automatic blacklisting prevents old token reuse

2. **Email Verification Tokens**:
   - Time-limited (24 hours)
   - JWT-based with HMAC signature
   - Client-specific (cannot be reused for other clients)

### Rate Limiting

**Recommended**:
- Limit registration attempts per token
- IP-based rate limiting on public endpoints
- Email verification retry limits

### Data Privacy

- Client data stored with tenant isolation
- Email verification required before activation
- GDPR compliance: right to deletion, data export

---

## Testing

### Manual Testing Flow

1. **Generate Token** (as admin):
   ```bash
   curl -X GET "http://localhost:8080/api/v1/clients/registrationtoken?organization_id=1" \
     -H "Authorization: Bearer <admin_token>"
   ```

2. **Register Client** (as public user):
   ```bash
   curl -X POST "http://localhost:8080/api/v1/clients/registration/<token>" \
     -H "Content-Type: application/json" \
     -d '{
       "first_name": "Jane",
       "last_name": "Smith",
       "email": "jane@example.com"
     }'
   ```

3. **Verify Email** (as public user):
   ```bash
   curl -X POST "http://localhost:8080/api/v1/clients/emailverification/<verification_token>"
   ```

### Test Scenarios

- ✅ Generate token with organization_id
- ✅ Generate second token blacklists first
- ✅ Register with valid token
- ✅ Register multiple clients with same token (used_count increments)
- ✅ Verify email with valid JWT
- ❌ Register with blacklisted token
- ❌ Register with invalid token
- ❌ Register with email mismatch
- ❌ Verify with expired JWT
- ❌ Verify with invalid JWT

---

## Migration Path

### Database Migration Required

```sql
-- Add new columns to registration_tokens table
ALTER TABLE registration_tokens 
  ADD COLUMN organization_id INTEGER,
  ADD COLUMN used_count INTEGER DEFAULT 0,
  ADD COLUMN blacklisted BOOLEAN DEFAULT false;

-- Remove old expiration columns (if migrating from old system)
ALTER TABLE registration_tokens 
  DROP COLUMN IF EXISTS expires_at,
  DROP COLUMN IF EXISTS used_at;

-- Add indexes for performance
CREATE INDEX idx_registration_tokens_org_blacklisted 
  ON registration_tokens(organization_id, blacklisted);

-- Add foreign key constraint
ALTER TABLE registration_tokens 
  ADD CONSTRAINT fk_registration_tokens_organization 
  FOREIGN KEY (organization_id) REFERENCES organizations(id);

-- Add email_verified to clients table
ALTER TABLE clients 
  ADD COLUMN email_verified BOOLEAN DEFAULT false;
```

---

## Future Enhancements

1. **Email Integration**:
   - Automated verification email sending
   - Customizable email templates
   - Email delivery tracking

2. **Admin Dashboard**:
   - View active registration tokens
   - Monitor registration statistics
   - Manual token revocation

3. **Multi-Organization Support**:
   - Generate tokens for multiple organizations
   - Organization-specific registration forms
   - Custom branding per organization

4. **Advanced Security**:
   - IP-based rate limiting
   - CAPTCHA integration
   - Fraud detection

5. **Analytics**:
   - Registration conversion rates
   - Token usage statistics
   - Email verification success rates

---

## Swagger Documentation

All endpoints are automatically documented in Swagger UI at `/swagger/index.html` with:

- Complete request/response schemas
- Example values for all fields
- Authentication requirements
- Error response formats
- Interactive API testing

**Access Swagger**: `http://localhost:8080/swagger/index.html`
