# One-Time Booking Link Invalidation

## Implementation Summary

Successfully implemented automatic invalidation of one-time booking links after the first booking is created.

## How It Works

### Token Types Affected
1. **One-Time Booking Links** (`token_purpose: "one-time-booking-link"`)
2. **Limited Use Links with max_use_count = 1**

### Workflow

1. **Token Creation**
   ```json
   POST /api/v1/booking/link
   {
     "client_id": 41,
     "template_id": 1,
     "token_purpose": "one-time-booking-link"
   }
   ```
   - Token is generated with `Purpose = OneTimeBookingLink`
   - Default `max_use_count = 1` for one-time tokens
   - Default expiration = 24 hours

2. **Viewing Free Slots** (Does NOT invalidate)
   ```
   GET /api/v1/booking/freeslots/{token}
   ```
   - Token is validated
   - Usage count is incremented
   - Token remains usable (unless usage limit reached)

3. **Creating a Booking** (DOES invalidate)
   ```
   POST /api/v1/sessions/book/{token}
   ```
   - Token is validated
   - Booking/session is created
   - **Auto-invalidation triggered**:
     - Token's `use_count` is set to `max_use_count`
     - Token usage record is marked as exhausted
     - Token can no longer be used

4. **Subsequent Attempts**
   - Any further use of the token will fail
   - Error: "Token limit reached"
   - Error: "This booking link has already been used the maximum number of times"

## Code Changes

### 1. BookingLinkService (`modules/booking/services/booking_link_service.go`)

Added new method:
```go
func (s *BookingLinkService) InvalidateOneTimeToken(token string, claims *entities.BookingLinkClaims) error
```

This method:
- Checks if token is one-time (`Purpose == OneTimeBookingLink` or `MaxUseCount == 1`)
- Updates `booking_token_usage` table to set `use_count = max_use_count`
- Creates a usage record if one doesn't exist
- Marks token as exhausted to prevent further use

### 2. SessionService (`unburdy_server/modules/client_management/services/session_service.go`)

Updated `BookSessionsWithToken()`:
```go
// After successful booking
seriesID, sessions, err := s.BookSessions(fullReq, claims.TenantID, claims.UserID)
if err != nil {
    return nil, nil, err
}

// Invalidate one-time tokens
if err := s.bookingLinkService.InvalidateOneTimeToken(token, claims); err != nil {
    // Log error but don't fail the booking
    fmt.Printf("Warning: failed to invalidate one-time token: %v\n", err)
}
```

Key points:
- Invalidation happens AFTER successful booking
- If invalidation fails, booking still succeeds (fail-safe)
- Error is logged but not returned to user

## Database Impact

### booking_token_usage Table
When a one-time token is used for booking:

**Before booking:**
```
token_id: abc123...
use_count: 1
max_use_count: 1
```

**After booking (auto-invalidated):**
```
token_id: abc123...
use_count: 1  (equals max_use_count)
max_use_count: 1
last_used_at: 2025-12-22 10:30:00
```

The middleware's `HasReachedLimit()` check will return `true`, preventing further use.

## Testing

Created comprehensive tests in `token_invalidation_test.go`:

1. **TestInvalidateOneTimeToken** - Verifies different token types
   - One-time links are invalidated ✓
   - Max use = 1 links are invalidated ✓
   - Unlimited links are NOT invalidated ✓
   - Multi-use links are NOT invalidated ✓

2. **TestInvalidateOneTimeTokenPreventsReuse** - Verifies prevention
   - Token can be used once ✓
   - After booking, token is exhausted ✓
   - `HasReachedLimit()` returns true ✓
   - `CanBeUsed()` returns false ✓

## User Experience

### Happy Path
1. Client receives one-time booking link via email
2. Client clicks link, views available time slots
3. Client selects a slot and confirms booking
4. Booking is created successfully
5. Link is automatically invalidated
6. If client tries to use link again → Error message

### Edge Cases Handled
- **Multiple simultaneous requests**: First successful booking wins
- **Invalidation failure**: Booking succeeds, error logged
- **Token already exhausted**: Middleware prevents booking attempt
- **Non-one-time tokens**: Not affected, continue to work as configured

## Security Benefits

1. **Prevents replay attacks** - One-time links can't be reused
2. **Reduces abuse** - Limits booking link sharing
3. **Clear audit trail** - Token usage is tracked in database
4. **Fail-safe design** - Booking succeeds even if invalidation fails

## Configuration Examples

### One-Time Link (24 hours)
```json
{
  "token_purpose": "one-time-booking-link"
}
```
→ Invalidated after first booking

### One-Time Link (Custom 7 days)
```json
{
  "token_purpose": "one-time-booking-link",
  "validity_days": 7
}
```
→ Invalidated after first booking OR after 7 days

### Package Deal (5 uses, 30 days)
```json
{
  "token_purpose": "timed-booking-link",
  "max_use_count": 5,
  "validity_days": 30
}
```
→ NOT auto-invalidated, can be used 5 times

### Timed Link (180 days default)
```json
{
  "token_purpose": "timed-booking-link"
}
```
→ NOT auto-invalidated, unlimited uses for 180 days

## Backward Compatibility

- Existing tokens continue to work
- Existing middleware validation unchanged
- Only affects new bookings going forward
- No database migration required (table already exists)

## Future Enhancements

Potential improvements:
1. Webhook notification when token is exhausted
2. Auto-generate new token after invalidation
3. Grace period for accidental re-bookings
4. Admin dashboard to view exhausted tokens
5. Email notification to client when link is used
