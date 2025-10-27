# AE SaaS Basic

A foundational Go module for building multi-tenant SaaS applications with essential features like user management, authentication, customer billing, plans, email system, and contact management.

## Features

- 🔐 **JWT Authentication** - Secure user authentication with token blacklisting
- 👥 **Multi-tenant Architecture** - Organization-level data separation
- 👤 **User Management** - Role-based access control and user settings
- 💳 **Customer & Billing** - Customer management with subscription plans
- 📧 **Email System** - Email tracking and management
- 📞 **Contact Management** - Generic contact management system
- 📄 **PDF Generation** - Template-based PDF document generation with dynamic data
- 🔍 **Fuzzy Search** - Intelligent multi-entity search with relevance scoring and highlighting
- 🎨 **Static Assets** - Professional email templates, CSS, and images
- 🏥 **Health Checks** - API health monitoring
- 🛡️ **Security** - Password hashing, CORS, and middleware protection

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database

### Installation

```bash
git clone https://github.com/your-org/ae-saas-basic.git
cd ae-saas-basic
go mod download
```

### Environment Variables

Create a `.env` file in the project root:

```env
# Server Configuration
PORT=8080
HOST=0.0.0.0
GIN_MODE=debug

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ae_saas_basic
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOUR=24

# Email Configuration (Optional)
SMTP_HOST=localhost
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
FROM_EMAIL=noreply@ae-saas-basic.com
FROM_NAME=AE SaaS Basic
```

### Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE ae_saas_basic;
```

### Running the Application

```bash
# Run the application
go run main.go
```

The API will be available at `http://localhost:8080`

### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

## API Endpoints

### Public Endpoints

- `GET /api/v1/health` - Health check
- `GET /api/v1/ping` - Simple ping
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/plans` - List available plans
- `GET /api/v1/plans/:id` - Get plan by ID

### Protected Endpoints (Require Authentication)

#### Authentication
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/change-password` - Change password
- `GET /api/v1/auth/me` - Get current user info

#### Customers
- `GET /api/v1/customers` - List customers (tenant-isolated)
- `GET /api/v1/customers/:id` - Get customer by ID
- `POST /api/v1/customers` - Create customer
- `PUT /api/v1/customers/:id` - Update customer
- `DELETE /api/v1/customers/:id` - Delete customer

#### Contacts
- `GET /api/v1/contacts` - List contacts
- `GET /api/v1/contacts/:id` - Get contact by ID
- `POST /api/v1/contacts` - Create contact
- `PUT /api/v1/contacts/:id` - Update contact
- `DELETE /api/v1/contacts/:id` - Delete contact

#### Emails
- `GET /api/v1/emails` - List emails
- `GET /api/v1/emails/:id` - Get email by ID
- `POST /api/v1/emails/send` - Send email
- `GET /api/v1/emails/stats` - Get email statistics

#### User Settings
- `GET /api/v1/user-settings` - Get user settings
- `PUT /api/v1/user-settings` - Update user settings
- `POST /api/v1/user-settings/reset` - Reset to defaults

### Admin Endpoints (Require Admin Role)

#### Plans Management
- `POST /api/v1/admin/plans` - Create plan
- `PUT /api/v1/admin/plans/:id` - Update plan
- `DELETE /api/v1/admin/plans/:id` - Delete plan

## Usage as a Module

### Integration in Your Project

1. **Add as dependency:**
```bash
go mod init your-project
go get github.com/your-org/ae-saas-basic
```

2. **Import and use:**
```go
package main

import (
    "github.com/ae-saas-basic/ae-saas-basic/internal/config"
    "github.com/ae-saas-basic/ae-saas-basic/internal/database"
    "github.com/ae-saas-basic/ae-saas-basic/internal/router"
    "github.com/ae-saas-basic/ae-saas-basic/pkg/auth"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Set JWT secret
    auth.SetJWTSecret(cfg.JWT.Secret)
    
    // Connect to database
    db, err := database.Connect(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Run migrations
    database.Migrate(db)
    
    // Setup router with your custom routes
    r := router.SetupRouter(db)
    
    // Add your custom routes
    yourGroup := r.Group("/api/v1/your-feature")
    yourGroup.Use(middleware.AuthMiddleware(db))
    {
        yourGroup.GET("", yourHandler.GetYourData)
        // ... more routes
    }
    
    // Start server
    r.Run(":8080")
}
```

### Extending the Module

You can extend this module by:

1. **Adding new models** - Define your domain-specific models that reference User/Organization
2. **Creating new handlers** - Build on top of the authentication middleware
3. **Utilizing tenant isolation** - Use the organization context for multi-tenant data separation

### Database Models

The module provides these core models:

- `Organization` - Tenant separation
- `User` - User accounts with roles
- `Plan` - Subscription plans
- `Customer` - Billing customers
- `Contact` - Contact management
- `Email` - Email tracking
- `UserSettings` - User preferences
- `TokenBlacklist` - JWT token management

## Architecture

```
ae-saas-basic/
├── cmd/server/           # Alternative main entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # Authentication and CORS middleware
│   ├── models/          # Data models and request/response structures
│   └── router/          # Route definitions
├── pkg/
│   ├── auth/           # JWT utilities
│   └── utils/          # Helper functions
└── main.go             # Application entry point
```

## Security Features

- **JWT Authentication** with token blacklisting on logout
- **Role-based Access Control** (user, admin, super-admin)
- **Multi-tenant Data Isolation** at organization level
- **Password Hashing** using bcrypt
- **CORS Protection**
- **Input Validation** with Gin binding

## PDF Generation System

AE SaaS Basic includes a comprehensive, generalized PDF generation system that can be used for creating professional documents like invoices, reports, certificates, and more.

### Features

- **Template-based Generation** - HTML templates with Go template syntax
- **Dynamic Data Injection** - Pass any data structure to templates
- **Configurable Output** - Page size, orientation, margins, quality settings
- **Multiple Output Modes** - Download, save to file, or stream
- **Template Validation** - Built-in validation for template data
- **Common Document Types** - Pre-built models for invoices, reports, certificates
- **REST API Endpoints** - Easy integration via HTTP API
- **Error Handling** - Comprehensive error reporting and logging

### System Requirements

The PDF generation system uses `wkhtmltopdf` for HTML to PDF conversion:

```bash
# Ubuntu/Debian
sudo apt-get install wkhtmltopdf

# macOS
brew install wkhtmltopdf

# CentOS/RHEL
sudo yum install wkhtmltopdf
```

### Configuration

Add PDF configuration to your `.env` file:

```env
# PDF Generation Configuration
PDF_TEMPLATE_DIR=./statics/templates/pdf
PDF_OUTPUT_DIR=./output/pdf
PDF_PAGE_SIZE=A4
PDF_ORIENTATION=Portrait
PDF_MARGIN_TOP=1cm
PDF_MARGIN_RIGHT=1cm
PDF_MARGIN_BOTTOM=1cm
PDF_MARGIN_LEFT=1cm
PDF_QUALITY=80
PDF_ENABLE_JS=true
PDF_LOAD_TIMEOUT=30
PDF_MAX_FILE_SIZE=52428800
```

### PDF API Endpoints

#### Template Management
- `GET /api/v1/pdf/templates` - List available templates
- `GET /api/v1/pdf/templates/{template}` - Get template information
- `POST /api/v1/pdf/templates/{template}/preview` - Preview template with data

#### PDF Generation
- `POST /api/v1/pdf/generate` - Generate PDF from template and data
- `POST /api/v1/pdf/generate/html` - Generate PDF from raw HTML
- `POST /api/v1/pdf/generate/stream` - Stream PDF generation (for large files)

#### Configuration
- `GET /api/v1/pdf/config` - Get current PDF configuration

### Usage Example

Generate an invoice PDF:

```bash
curl -X POST http://localhost:8080/api/v1/pdf/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "template": "invoice",
    "data": {
      "InvoiceNumber": "INV-001",
      "InvoiceDate": "2024-01-15",
      "Company": {
        "Name": "Your Company",
        "Address": "123 Business St"
      },
      "Customer": {
        "Name": "John Doe",
        "Email": "john@example.com"
      },
      "Items": [
        {
          "Description": "Consultation",
          "Quantity": 1,
          "UnitPrice": 150.00,
          "Total": 150.00
        }
      ],
      "Total": 150.00,
      "Currency": "USD"
    }
  }' --output invoice.pdf
```

### Template Structure

Create HTML templates in `./statics/templates/pdf/`:

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Invoice #{{.InvoiceNumber}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { border-bottom: 2px solid #333; }
        .total { font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Company.Name}}</h1>
        <p>Invoice #{{.InvoiceNumber}}</p>
    </div>
    
    <div class="customer">
        <h3>Bill To:</h3>
        <p>{{.Customer.Name}}<br>{{.Customer.Email}}</p>
    </div>
    
    <table>
        <thead>
            <tr><th>Description</th><th>Qty</th><th>Price</th><th>Total</th></tr>
        </thead>
        <tbody>
        {{range .Items}}
            <tr>
                <td>{{.Description}}</td>
                <td>{{.Quantity}}</td>
                <td>{{FormatCurrency .UnitPrice .Currency}}</td>
                <td>{{FormatCurrency .Total .Currency}}</td>
            </tr>
        {{end}}
        </tbody>
        <tfoot>
            <tr class="total">
                <td colspan="3">Total:</td>
                <td>{{FormatCurrency .Total .Currency}}</td>
            </tr>
        </tfoot>
    </table>
    
    <p>Generated on {{.GeneratedDate}}</p>
</body>
</html>
```

### Integration in Code

```go
package main

import (
    "github.com/ae-saas-basic/ae-saas-basic/internal/services"
    "github.com/ae-saas-basic/ae-saas-basic/internal/handlers"
)

func setupPDFService() *handlers.PDFHandler {
    // Initialize PDF service
    pdfService := services.NewPDFService(
        "./statics/templates/pdf",
        "./output/pdf",
        nil, // Use default config
    )
    
    return handlers.NewPDFHandler(pdfService)
}
```

The PDF generation system is designed to be flexible and reusable across different projects. You can create custom templates for any document type and integrate the system into existing applications easily.

For more detailed examples and integration patterns, see [INTEGRATION_EXAMPLE.md](./INTEGRATION_EXAMPLE.md).

## Fuzzy Search System

AE SaaS Basic includes a powerful, generalized fuzzy search system that enables intelligent search across multiple data types with relevance scoring, highlighting, and advanced filtering capabilities.

### Features

- **Multi-Entity Search** - Search across users, customers, contacts, plans, emails, and custom entities
- **Relevance Scoring** - Configurable scoring algorithms with exact match and prefix match boosts
- **Text Highlighting** - Highlight matching text in search results
- **Advanced Filters** - Date ranges, numeric ranges, boolean filters, and custom criteria
- **Search Preferences** - User-specific search settings and saved searches
- **Search Analytics** - Track search queries, results, and user behavior
- **Caching System** - Configurable result caching for improved performance
- **Admin Controls** - Manage entity types, search configuration, and analytics
- **Multi-tenant Support** - Organization-level data isolation

### Configuration

Add fuzzy search configuration to your `.env` file:

```env
# Fuzzy Search Configuration
FUZZY_MIN_SEARCH_LENGTH=2
FUZZY_MAX_RESULTS=50
FUZZY_SCORE_THRESHOLD=0.3
FUZZY_ENABLE_HIGHLIGHT=true
FUZZY_CASE_SENSITIVE=false
FUZZY_EXACT_MATCH_BOOST=2.0
FUZZY_PREFIX_MATCH_BOOST=1.5
FUZZY_ENABLE_STEMMING=true
FUZZY_ENABLE_SYNONYMS=false
FUZZY_ENABLE_LOGGING=false
FUZZY_CACHE_RESULTS=true
FUZZY_CACHE_TIMEOUT_MIN=30
```

### Search API Endpoints

#### General Search
- `GET /api/v1/search` - Multi-entity search with query parameter
- `POST /api/v1/search/advanced` - Advanced search with filters
- `GET /api/v1/search/quick?q=query` - Quick search for autocomplete

#### Entity-Specific Search
- `GET /api/v1/search/users` - Search users only
- `GET /api/v1/search/customers` - Search customers only
- `GET /api/v1/search/contacts` - Search contacts only
- `GET /api/v1/search/plans` - Search plans only
- `GET /api/v1/search/emails` - Search emails only

#### Search Management
- `GET /api/v1/search/preferences` - Get user search preferences
- `POST /api/v1/search/preferences` - Update search preferences
- `GET /api/v1/search/saved` - Get saved searches
- `POST /api/v1/search/saved` - Save search query
- `DELETE /api/v1/search/saved/{id}` - Delete saved search

#### Admin Endpoints (Requires Admin Role)
- `GET /api/v1/admin/search/entities` - List searchable entity types
- `PUT /api/v1/admin/search/config` - Update search configuration
- `GET /api/v1/admin/search/analytics` - Get search analytics

### Usage Examples

#### Basic Search

Search across all entity types:

```bash
curl -X GET "http://localhost:8080/api/v1/search?q=john&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "query": "john",
  "total_results": 15,
  "results": [
    {
      "entity_type": "user",
      "entity_id": "123",
      "score": 0.95,
      "title": "John Doe",
      "description": "Software Developer at AcmeCorp",
      "highlighted_text": "<mark>John</mark> Doe",
      "metadata": {
        "email": "john.doe@example.com",
        "role": "user"
      }
    },
    {
      "entity_type": "customer",
      "entity_id": "456",
      "score": 0.87,
      "title": "Johnson Inc.",
      "description": "Enterprise customer since 2023",
      "highlighted_text": "<mark>John</mark>son Inc.",
      "metadata": {
        "plan": "enterprise",
        "status": "active"
      }
    }
  ]
}
```

#### Advanced Search with Filters

```bash
curl -X POST http://localhost:8080/api/v1/search/advanced \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "query": "developer",
    "entity_types": ["user", "customer"],
    "filters": {
      "date_range": {
        "field": "created_at",
        "start": "2024-01-01",
        "end": "2024-12-31"
      },
      "boolean_filters": {
        "active": true
      },
      "numeric_range": {
        "field": "age",
        "min": 25,
        "max": 45
      }
    },
    "sort_by": "relevance",
    "limit": 20
  }'
```

#### Entity-Specific Search

Search only within customers:

```bash
curl -X GET "http://localhost:8080/api/v1/search/customers?q=enterprise&status=active" \
  -H "Authorization: Bearer $TOKEN"
```

#### Quick Search for Autocomplete

```bash
curl -X GET "http://localhost:8080/api/v1/search/quick?q=jo&limit=5" \
  -H "Authorization: Bearer $TOKEN"
```

### Integration in Code

```go
package main

import (
    "github.com/ae-saas-basic/ae-saas-basic/internal/services"
    "github.com/ae-saas-basic/ae-saas-basic/internal/handlers"
    "github.com/ae-saas-basic/ae-saas-basic/internal/models"
)

func setupFuzzySearchService(db *gorm.DB, config config.FuzzySearchConfig) *handlers.FuzzySearchHandler {
    // Initialize fuzzy search service
    fuzzyService := services.NewFuzzySearchService(db, config)
    
    // Register entity types for search
    fuzzyService.RegisterEntity("user", models.User{})
    fuzzyService.RegisterEntity("customer", models.Customer{})
    fuzzyService.RegisterEntity("contact", models.Contact{})
    fuzzyService.RegisterEntity("plan", models.Plan{})
    fuzzyService.RegisterEntity("email", models.Email{})
    
    return handlers.NewFuzzySearchHandler(fuzzyService)
}
```

### Custom Entity Integration

To make your custom entities searchable, implement the `Searchable` interface:

```go
type MyCustomEntity struct {
    ID          uint   `gorm:"primaryKey"`
    Title       string `gorm:"size:255"`
    Description string `gorm:"type:text"`
    Tags        string `gorm:"size:500"`
    CreatedAt   time.Time
}

// Implement Searchable interface
func (e MyCustomEntity) GetSearchFields() map[string]float64 {
    return map[string]float64{
        "title":       1.0,  // Highest weight
        "description": 0.7,  // Medium weight
        "tags":        0.5,  // Lower weight
    }
}

func (e MyCustomEntity) GetTitle() string {
    return e.Title
}

func (e MyCustomEntity) GetDescription() string {
    return e.Description
}

func (e MyCustomEntity) GetMetadata() map[string]interface{} {
    return map[string]interface{}{
        "tags": e.Tags,
        "created_at": e.CreatedAt,
    }
}

// Register the entity
fuzzyService.RegisterEntity("my_entity", MyCustomEntity{})
```

### Search Preferences and Personalization

Users can customize their search experience:

```bash
# Update search preferences
curl -X POST http://localhost:8080/api/v1/search/preferences \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "preferred_entities": ["user", "customer"],
    "results_per_page": 25,
    "highlight_enabled": true,
    "save_history": true
  }'
```

### Search Analytics

Track search performance and user behavior:

```bash
# Get search analytics (Admin only)
curl -X GET http://localhost:8080/api/v1/admin/search/analytics \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

Response:
```json
{
  "total_searches": 1250,
  "popular_queries": [
    {"query": "john", "count": 45},
    {"query": "enterprise", "count": 32},
    {"query": "developer", "count": 28}
  ],
  "entity_popularity": [
    {"entity": "user", "searches": 567},
    {"entity": "customer", "searches": 423},
    {"entity": "contact", "searches": 260}
  ],
  "avg_results_per_search": 8.3,
  "avg_search_time_ms": 45.2
}
```

The fuzzy search system is designed to be highly configurable and extensible. You can easily integrate it with any Go application and customize it for specific domain needs.

## Testing

AE SaaS Basic includes a comprehensive HURL test suite that covers all endpoints and functionality.

### Running Tests

```bash
# Install HURL
brew install hurl  # macOS
# or
sudo apt-get install hurl  # Ubuntu

# Run all tests
./run-hurl-tests.sh

# Run specific test category
hurl --variable host=http://localhost:8080 tests/hurl/08_pdf_generation.hurl
hurl --variable host=http://localhost:8080 tests/hurl/09_static_assets.hurl
```

### Test Coverage

The HURL test suite includes:

- **Core System Tests** - Health checks, authentication, user management
- **Business Logic Tests** - Plans, customers, contacts, emails, user settings
- **PDF Generation Tests** - Template management, PDF generation, streaming
- **Static Asset Tests** - File serving, security, performance

For detailed testing documentation, see [HURL_TESTING.md](./HURL_TESTING.md).

### Prerequisites for Tests

- Server running at `http://localhost:8080`
- PostgreSQL database connected
- wkhtmltopdf installed (for PDF tests)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For questions and support, please open an issue in the GitHub repository.