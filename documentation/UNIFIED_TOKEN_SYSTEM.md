# Unified Token System Implementation

## Overview

Successfully implemented a unified token system that consolidates token generation and validation across the base server and booking module, while maintaining backward compatibility with existing tokens.

## Changes Summary

### Base Server Changes

#### 1. New Generic Token Service (`pkg/auth/token_service.go`)
- Created `TokenService` struct with generic token operations
- Supports custom claim payloads implementing `jwt.Claims` interface
- Methods:
  - `GenerateToken(claims interface{})` - generates JWT with custom claims
  - `ValidateToken(tokenString, claims interface{})` - validates and populates claims
  - `ParseTokenID(tokenString)` - extracts token ID without full validation
  - `GetTokenExpiration(tokenString)` - extracts expiration without full validation
  - `GetSharedTokenService()` - returns singleton instance using shared JWT secret

#### 2. Core Interface Updates (`pkg/core/interfaces.go`)
- Added `TokenService` interface for dependency injection
- Added `TokenService` field to `ModuleContext`
- Added `time` import for expiration handling

#### 3. Bootstrap Integration (`pkg/bootstrap/application.go`)
- Created `TokenService` instance during initialization
- Injected `TokenService` into `ModuleContext` for all modules
- Uses same JWT secret as existing auth system

### Booking Module Changes

#### 1. Updated Claims Structure (`entities/booking_link.go`)
- Modified `BookingLinkClaims` to embed `jwt.RegisteredClaims`
- Removed custom `IssuedAt` and `ExpiresAt` fields (now in `RegisteredClaims`)
- Maintains backward compatibility with JSON serialization
- Now implements `jwt.Claims` interface

#### 2. Enhanced BookingLinkService (`services/booking_link_service.go`)
- Added `TokenServiceInterface` for abstraction
- New constructors:
  - `NewBookingLinkService()` - legacy implementation (backward compatible)
  - `NewBookingLinkServiceWithTokenService()` - uses unified token service
- Updated `GenerateBookingLinkWithOptions()`:
  - Uses `jwt.RegisteredClaims` for standard fields
  - Falls back to legacy implementation if TokenService unavailable
- Updated `ValidateBookingLink()`:
  - Tries unified service first, then falls back to legacy
  - New `validateLegacyToken()` method for backward compatibility
- Maintains legacy `generateToken()` for fallback

#### 3. Module Initialization (`module.go`)
- Auto-detects TokenService availability in ModuleContext
- Creates `tokenServiceAdapter` to bridge interfaces
- Logs which token system is being used
- Graceful fallback to legacy implementation

#### 4. Middleware Updates (`middleware/token_auth.go`)
- Updated to handle new `RegisteredClaims` structure
- Changed expiration check from `claims.ExpiresAt > 0` to `claims.ExpiresAt != nil`
- Maintains all existing validation logic (blacklist, usage tracking, etc.)

### Token Usage Tracking (Unchanged)
- `entities/token_usage.go` - unchanged, still tracks usage limits
- Migration file - unchanged
- Usage validation in middleware - updated for new claims format only

## Backward Compatibility

### How It Works:
1. **New Tokens**: Generated using unified TokenService with standard JWT library
2. **Old Tokens**: Still validated using legacy custom implementation
3. **Automatic Fallback**: Service tries unified validation first, falls back to legacy
4. **No Data Migration**: Existing tokens remain valid without changes

### Migration Path:
- **Phase 1** (Current): Both systems work side-by-side
- **Phase 2** (Future): New tokens use unified system, old tokens expire naturally
- **Phase 3** (Future): Remove legacy implementation after all old tokens expire

## Benefits

1. **Single Source of Truth**: All modules use same token generation logic
2. **Consistent Security**: Same secret, same algorithm across all token types
3. **Module Flexibility**: Each module defines its own payload structure
4. **Easy Extension**: New modules can easily add custom token types
5. **Standard Library**: Uses industry-standard `golang-jwt/jwt/v5`
6. **Usage Tracking**: Booking-specific features remain module-scoped
7. **Zero Downtime**: Backward compatible, no migration required

## Token Types

### Base Server Tokens:
- **Authentication** (`JWTClaims`) - User session tokens
- **Password Reset** (`ResetTokenClaims`) - Password reset flow
- **Email Verification** (`JWTClaims` with type) - Email confirmation
- **User Signup** (`JWTClaims`) - Invitation links

### Booking Module Tokens:
- **Booking Links** (`BookingLinkClaims`) - Booking slot access
  - One-time use tokens
  - Limited use tokens (N times)
  - Time-limited tokens
  - Permanent tokens

## Usage Example

### For New Modules:

```go
// 1. Define your claims (must implement jwt.Claims)
type MyModuleClaims struct {
    CustomField string `json:"custom_field"`
    jwt.RegisteredClaims
}

// 2. In module initialization
func (m *MyModule) Initialize(ctx core.ModuleContext) error {
    if ctx.TokenService != nil {
        // Use unified token service
        claims := &MyModuleClaims{
            CustomField: "value",
            RegisteredClaims: jwt.RegisteredClaims{
                ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
                IssuedAt:  jwt.NewNumericDate(time.Now()),
            },
        }
        
        token, err := ctx.TokenService.GenerateToken(claims)
        // ...
    }
}
```

## Testing

### Compilation:
- ✅ Base server builds successfully
- ✅ Booking module builds successfully
- ✅ Go workspace synchronized

### To Test Runtime:
1. Start the server
2. Create a new booking link (uses new system)
3. Use an old booking link (uses legacy validation)
4. Both should work correctly

## Future Improvements

1. **Standardize Base Server**: Update password reset and email verification to use TokenService
2. **Remove Legacy Code**: After all old tokens expire, remove fallback implementation
3. **Generic Usage Tracking**: Consider making token usage tracking a base feature
4. **Token Refresh**: Add token refresh capability to TokenService
5. **Token Revocation**: Integrate blacklist checking into TokenService
