# Booking Token Validation with Limited Use

## Overview

The booking module now supports tokens with limited use and expiration. This allows you to create booking links that can only be used:
- A specific number of times (e.g., one-time use, 5 times, etc.)
- Within a specific time period (e.g., valid for 7 days)
- Or both

## Features

### 1. Token Usage Tracking
- Each token usage is tracked in the `booking_token_usage` table
- Tracks: use count, max use count, last used timestamp, and expiration
- Automatic validation on each token use

### 2. Token Types

#### One-Time Booking Links
- Default expiration: 24 hours
- Can be configured with custom validity period
- Automatically tracked for usage

#### Timed Booking Links
- Default expiration: 180 days
- Can be configured with custom validity period or max use count
- Useful for reusable booking links with time limits

### 3. Usage Limits
- `max_use_count = 0`: Unlimited uses (default for permanent links)
- `max_use_count = 1`: One-time use only
- `max_use_count = N`: Can be used N times

### 4. Validity Period
- `validity_days = 0`: No expiration
- `validity_days = N`: Valid for N days from creation
- Default for timed links: 180 days

## API Usage

### Creating a Booking Link

**Endpoint**: `POST /api/v1/booking/link`

**Request Body**:
```json
{
  "template_id": 1,
  "client_id": 123,
  "token_purpose": "one-time-booking-link",
  "max_use_count": 1,
  "validity_days": 7
}
```

**Parameters**:
- `template_id` (required): The booking template ID
- `client_id` (required): The client ID for whom the link is generated
- `token_purpose` (required): Either `"one-time-booking-link"` or `"timed-booking-link"`
- `max_use_count` (optional): Maximum number of times the token can be used (0 = unlimited)
- `validity_days` (optional): Number of days the token is valid (0 = no expiration, default 180 for timed links)

**Response**:
```json
{
  "success": true,
  "message": "Booking link created successfully",
  "data": {
    "token": "eyJhbGci...",
    "url": "http://localhost:3005/booking/eyJhbGci...",
    "purpose": "one-time-booking-link",
    "expires_at": "2025-12-29T10:00:00Z",
    "created_at": "2025-12-22T10:00:00Z"
  }
}
```

### Using a Booking Link

When a booking link is accessed (e.g., `GET /api/v1/booking/freeslots/{token}`):

1. **Token validation** - Verifies signature and structure
2. **Blacklist check** - Ensures token hasn't been revoked
3. **Expiration check** - Validates token hasn't expired
4. **Usage limit check** - Ensures token hasn't exceeded max uses
5. **Usage tracking** - Increments use count and updates last used timestamp

When a booking is created with a token (e.g., `POST /api/v1/sessions/book/{token}`):

1. All the above validation steps occur
2. **Booking creation** - Calendar entries and sessions are created
3. **Auto-invalidation** - One-time tokens are automatically marked as exhausted
   - Tokens with `token_purpose: "one-time-booking-link"` are invalidated
   - Tokens with `max_use_count: 1` are invalidated
   - The token's `use_count` is set to `max_use_count` to prevent further use
4. Subsequent attempts to use the token will fail with "Token limit reached" error

## Error Messages

The system returns specific error messages for different token validation failures:

### Invalid Token Format
```json
{
  "success": false,
  "message": "Invalid token",
  "error": "invalid token format"
}
```

### Expired Token
```json
{
  "success": false,
  "message": "Token expired",
  "error": "This booking link has expired and is no longer valid"
}
```

### Usage Limit Reached
```json
{
  "success": false,
  "message": "Token limit reached",
  "error": "This booking link has already been used the maximum number of times"
}
```

### Revoked Token
```json
{
  "success": false,
  "message": "Token revoked",
  "error": "This booking link has been revoked"
}
```

### Invalid Signature
```json
{
  "success": false,
  "message": "Invalid token",
  "error": "invalid token signature"
}
```

## Examples

### Example 1: One-Time Use Token (24 hours)
```json
{
  "template_id": 1,
  "client_id": 123,
  "token_purpose": "one-time-booking-link"
}
```
- Can be used only once
- Expires in 24 hours
- Perfect for single booking confirmations
- **Automatically invalidated after first booking is created**
- Viewing free slots does NOT invalidate the token
- Creating a session/booking DOES invalidate the token

### Example 2: Limited Use Token (5 uses, 30 days)
```json
{
  "template_id": 1,
  "client_id": 123,
  "token_purpose": "timed-booking-link",
  "max_use_count": 5,
  "validity_days": 30
}
```
- Can be used 5 times
- Valid for 30 days
- Great for package deals or multiple appointment bookings

### Example 3: Time-Limited Reusable Token (7 days)
```json
{
  "template_id": 1,
  "client_id": 123,
  "token_purpose": "timed-booking-link",
  "max_use_count": 0,
  "validity_days": 7
}
```
- Unlimited uses
- Valid for 7 days
- Useful for short-term access

### Example 4: Timed Unlimited Token (180 days default)
```json
{
  "template_id": 1,
  "client_id": 123,
  "token_purpose": "timed-booking-link"
}
```
- Unlimited uses
- Valid for 180 days (default)
- For medium-term client access

## Database Schema

### booking_token_usage Table

| Column | Type | Description |
|--------|------|-------------|
| id | BIGSERIAL | Primary key |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |
| deleted_at | TIMESTAMP | Soft delete timestamp |
| token_id | VARCHAR(64) | SHA256 hash of the token (unique) |
| tenant_id | BIGINT | Tenant ID |
| template_id | BIGINT | Template ID |
| client_id | BIGINT | Client ID |
| use_count | INTEGER | Current usage count |
| max_use_count | INTEGER | Maximum allowed uses (0 = unlimited) |
| last_used_at | TIMESTAMP | Last usage timestamp |
| expires_at | TIMESTAMP | Expiration timestamp |

## Migration

Run the migration to create the token usage tracking table:

```bash
psql -d your_database -f modules/booking/migrations/001_create_token_usage_table.sql
```

Or if using automatic migrations, the table will be created when the application starts.

## Security Considerations

1. **Token IDs** are SHA256 hashes of the actual token, not the token itself
2. **Blacklisting** is checked before usage validation
3. **Usage tracking** happens after all validations pass
4. **Atomic operations** ensure usage count is accurate even under concurrent access
5. **Expiration** is checked both at token level and usage tracking level

## Monitoring Token Usage

You can query token usage statistics:

```sql
-- Check usage for a specific token
SELECT * FROM booking_token_usage WHERE token_id = 'your_token_hash';

-- Find tokens nearing their limit
SELECT * FROM booking_token_usage 
WHERE max_use_count > 0 
  AND use_count >= max_use_count * 0.8
  AND deleted_at IS NULL;

-- Find expired tokens
SELECT * FROM booking_token_usage 
WHERE expires_at IS NOT NULL 
  AND expires_at < NOW()
  AND deleted_at IS NULL;
```

## Future Enhancements

Potential future improvements:
- Rate limiting per IP address
- Usage analytics and reporting
- Automatic cleanup of expired tokens
- Token refresh/extension functionality
- Webhook notifications when tokens expire or reach limits
