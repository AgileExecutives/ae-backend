# OpenAPI/Swagger Documentation for AE SaaS Basic API

## Overview

This project now includes comprehensive OpenAPI 3.0/Swagger documentation for all API endpoints. The documentation provides interactive API exploration, request/response schemas, and authentication details.

## Accessing the Documentation

### Swagger UI (Interactive)
- **URL**: `http://localhost:8080/swagger/index.html`
- **Description**: Interactive web interface to explore and test API endpoints
- **Features**: 
  - Try out API calls directly from the browser
  - View request/response schemas
  - Authentication support
  - Parameter validation

### Raw Documentation
- **JSON**: `http://localhost:8080/swagger/doc.json`
- **YAML**: Available in `docs/swagger.yaml`

## API Overview

### Base Information
- **Title**: AE SaaS Basic API
- **Version**: 1.0
- **Base URL**: `http://localhost:8080/api/v1`
- **License**: MIT

### Authentication
The API uses JWT Bearer token authentication:
```
Authorization: Bearer <your-jwt-token>
```

## Documented Endpoints

### Authentication (`/auth`)
- `POST /auth/login` - User login
- `POST /auth/register` - User registration  
- `POST /auth/logout` - User logout (requires authentication)

### Users (`/users`)
- `GET /users` - List users (paginated)
- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

### Customers (`/customers`)
- `GET /customers` - List customers (paginated)
- `GET /customers/{id}` - Get customer by ID
- `POST /customers` - Create customer
- `PUT /customers/{id}` - Update customer
- `DELETE /customers/{id}` - Delete customer

### Contacts (`/contacts`)
- `GET /contacts` - List contacts (paginated)
- `GET /contacts/{id}` - Get contact by ID
- `POST /contacts` - Create contact
- `PUT /contacts/{id}` - Update contact
- `DELETE /contacts/{id}` - Delete contact

### Plans (`/plans`)
- `GET /plans` - List subscription plans
- `GET /plans/{id}` - Get plan by ID
- `POST /plans` - Create plan
- `PUT /plans/{id}` - Update plan
- `DELETE /plans/{id}` - Delete plan

### Emails (`/emails`)
- `GET /emails` - List emails (paginated)
- `GET /emails/{id}` - Get email by ID
- `POST /emails/send` - Send email

### User Settings (`/user-settings`)
- `GET /user-settings` - Get user settings
- `PUT /user-settings` - Update user settings
- `POST /user-settings/reset` - Reset to defaults

### PDF Generation (`/pdf`)
- `GET /pdf/config` - Get PDF configuration
- `GET /pdf/templates` - List available templates
- `POST /pdf/generate` - Generate PDF from template
- `POST /pdf/generate/html` - Generate PDF from HTML

### Search (`/search`)
- `POST /search` - Advanced search
- `GET /search/quick` - Quick search
- `GET /search/types` - Get entity types
- `GET /search/health` - Search service health

### Static Files (`/static`)
- Static file serving for CSS, JS, images, templates

### System (`/health`)
- `GET /health` - System health check

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {...}
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message",
  "details": "Detailed error information"
}
```

### Paginated Response
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "pages": 10,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## Generating Documentation

### Prerequisites
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate/Update Documentation
```bash
# Generate swagger docs from code comments
swag init

# This creates/updates:
# - docs/docs.go
# - docs/swagger.json  
# - docs/swagger.yaml
```

### Adding Documentation to New Endpoints

Add swagger comments above your handler functions:

```go
// CreateUser creates a new user
// @Summary Create user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UserCreateRequest true "User data"
// @Success 201 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    // Implementation...
}
```

### Swagger Comment Tags
- `@Summary`: Brief endpoint description
- `@Description`: Detailed description
- `@Tags`: Group endpoints by functionality
- `@Accept`: Request content type
- `@Produce`: Response content type
- `@Security`: Authentication requirement
- `@Param`: Parameters (path, query, body, header)
- `@Success`: Success response format
- `@Failure`: Error response formats
- `@Router`: HTTP method and path

## Development Workflow

1. **Add new endpoints** with proper swagger comments
2. **Run** `swag init` to regenerate documentation
3. **Test** the API documentation at `/swagger/index.html`
4. **Commit** both code and generated docs

## Best Practices

### Documentation
- Always add swagger comments to public endpoints
- Use descriptive summaries and descriptions
- Document all parameters and their validation rules
- Include all possible response codes
- Group related endpoints with consistent tags

### Models
- Create dedicated request/response models
- Use proper JSON tags and validation
- Document model fields with comments
- Keep models in `internal/models/` package

### Testing
- Test endpoints through Swagger UI during development
- Verify request/response schemas match documentation
- Ensure authentication flows work correctly

## Integration Examples

### Using the API with curl
```bash
# 1. Login to get token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | \
  jq -r '.data.token')

# 2. Use token for authenticated requests
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/users
```

### Using with JavaScript/Frontend
```javascript
// Login and store token
const response = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'admin', password: 'admin123' })
});
const { data } = await response.json();
localStorage.setItem('token', data.token);

// Use token for requests
const users = await fetch('/api/v1/users', {
  headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
});
```

## Troubleshooting

### Documentation Not Updating
1. Run `swag init` to regenerate
2. Restart the server
3. Clear browser cache

### Swagger UI Not Loading
1. Check that imports are correct in main.go
2. Verify swagger route in router setup
3. Ensure docs package is imported

### Authentication Issues in Swagger UI
1. Use the "Authorize" button in Swagger UI
2. Enter: `Bearer <your-jwt-token>`
3. Get token from `/auth/login` endpoint first

## Production Considerations

### Security
- Disable Swagger UI in production environments
- Use environment-based conditional routing
- Implement proper CORS policies
- Secure sensitive endpoints appropriately

### Performance
- Consider serving static swagger assets from CDN
- Implement caching for documentation endpoints
- Monitor API usage through documentation access

This comprehensive OpenAPI/Swagger documentation system provides a solid foundation for API development, testing, and client integration.