# Database Schema Migration - Multi-Tenant Hierarchie

## Übersicht

Das Datenbankschema wurde angepasst, um die folgende Multi-Tenant-Hierarchie abzubilden:

```
Customer (Kunde/Mandant)
  └── Tenant (1:n - Ein Customer kann mehrere Tenants haben)
      └── Organization (1:n - Ein Tenant hat mehrere Organisationen)
          └── User (1:n - Ein User gehört zu einem Tenant und einer Organisation)
```

## Geänderte Modelle

### 1. Tenant Model (`base-server/internal/models/tenant.go`)

**Neu hinzugefügt:**
- `CustomerID uint` - Verknüpfung zum Customer (erforderlich)

**Geänderte Indizes:**
- `Name` und `Slug` sind nicht mehr global unique, sondern nur noch innerhalb eines Customers
- Neuer Index: `idx_tenant_customer` auf `customer_id`
- Neuer unique Index: `idx_tenant_slug` auf `slug`

**Beispiel:**
```go
type Tenant struct {
    ID         uint
    CustomerID uint   `gorm:"not null;index:idx_tenant_customer"`
    Name       string
    Slug       string `gorm:"not null;uniqueIndex:idx_tenant_slug"`
}
```

### 2. User Model (`base-server/internal/models/user.go`)

**Neu hinzugefügt:**
- `OrganizationID uint` - Verknüpfung zur Organization (erforderlich)

**Neue Indizes:**
- `idx_user_tenant` auf `tenant_id`
- `idx_user_organization` auf `organization_id`

**Beispiel:**
```go
type User struct {
    ID             uint
    TenantID       uint `gorm:"not null;index:idx_user_tenant"`
    OrganizationID uint `gorm:"not null;index:idx_user_organization"`
    Username       string
    Email          string
    // ... weitere Felder
}
```

### 3. Organization Model (`modules/organization/entities/organization.go`)

**Entfernt:**
- `UserID uint` - User gehört zur Organization, nicht umgekehrt

**Behalten:**
- `TenantID uint` - Organization gehört zu einem Tenant

**Beispiel:**
```go
type Organization struct {
    ID       uint
    TenantID uint `gorm:"not null;index:idx_organization_tenant"`
    Name     string
    // ... weitere Felder
}
```

## Geänderte Services & Handler

### Organization Service

Alle Methoden wurden angepasst, um `userID` Parameter zu entfernen:

- `CreateOrganization(req, tenantID, userID)` → `CreateOrganization(req, tenantID)`
- `GetOrganizationByID(id, tenantID, userID)` → `GetOrganizationByID(id, tenantID)`
- `GetOrganizations(page, limit, tenantID, userID)` → `GetOrganizations(page, limit, tenantID)`
- `UpdateOrganization(id, tenantID, userID, req)` → `UpdateOrganization(id, tenantID, req)`
- `DeleteOrganization(id, tenantID, userID)` → `DeleteOrganization(id, tenantID)`

### Organization Handler

Alle Handler wurden angepasst, um keine `userID` mehr aus dem Context zu lesen.

## Migrationsstrategie

### Option 1: Neue Datenbank (empfohlen für Entwicklung)

```bash
# Alte Datenbank löschen und neu erstellen
dropdb ae_saas_basic_test
createdb ae_saas_basic_test

# Server starten (führt Auto-Migration aus)
cd /Users/alex/src/ae/backend/unburdy_server
./bin/unburdy-server-api
```

### Option 2: Manuelle Migration (für Produktionsdaten)

```sql
-- 1. Customer-Tabelle erstellen (falls noch nicht vorhanden)
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 2. Standard-Customer erstellen
INSERT INTO customers (name, email) 
VALUES ('Default Customer', 'admin@example.com')
ON CONFLICT DO NOTHING;

-- 3. Tenants-Tabelle anpassen
ALTER TABLE tenants ADD COLUMN customer_id INTEGER;
UPDATE tenants SET customer_id = 1 WHERE customer_id IS NULL;
ALTER TABLE tenants ALTER COLUMN customer_id SET NOT NULL;
CREATE INDEX idx_tenant_customer ON tenants(customer_id);

-- 4. Tenants unique constraint anpassen
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_name_key;
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_slug_key;
CREATE UNIQUE INDEX idx_tenant_slug ON tenants(slug);

-- 5. Users-Tabelle anpassen
ALTER TABLE users ADD COLUMN organization_id INTEGER;

-- Default-Organization erstellen für jeden Tenant
INSERT INTO organizations (tenant_id, name, created_at, updated_at)
SELECT id, CONCAT('Default Organization - ', name), NOW(), NOW()
FROM tenants
WHERE NOT EXISTS (SELECT 1 FROM organizations WHERE organizations.tenant_id = tenants.id);

-- User mit Default-Organization verknüpfen
UPDATE users u
SET organization_id = (
    SELECT id FROM organizations o 
    WHERE o.tenant_id = u.tenant_id 
    LIMIT 1
)
WHERE organization_id IS NULL;

ALTER TABLE users ALTER COLUMN organization_id SET NOT NULL;
CREATE INDEX idx_user_tenant ON users(tenant_id);
CREATE INDEX idx_user_organization ON users(organization_id);

-- 6. Organizations-Tabelle bereinigen
ALTER TABLE organizations DROP COLUMN IF EXISTS user_id;
DROP INDEX IF EXISTS idx_organization_user;
```

## Seed-Daten anpassen

Die Seed-Skripte müssen angepasst werden:

### 1. Seed-Data JSON anpassen

```json
{
  "customers": [
    {
      "name": "Default Customer",
      "email": "admin@example.com"
    }
  ],
  "tenants": [
    {
      "customer_id": 1,
      "name": "Admin Tenant",
      "slug": "admin"
    }
  ],
  "organizations": [
    {
      "tenant_id": 1,
      "name": "Default Organization"
    }
  ],
  "users": [
    {
      "tenant_id": 1,
      "organization_id": 1,
      "username": "admin",
      "email": "admin@example.com"
    }
  ]
}
```

### 2. Seed-Funktion anpassen

```go
// Erst Customers erstellen
for _, customerData := range seedData.Customers {
    customer := models.Customer{
        Name:  customerData.Name,
        Email: customerData.Email,
    }
    db.Create(&customer)
}

// Dann Tenants mit CustomerID
for _, tenantData := range seedData.Tenants {
    tenant := models.Tenant{
        CustomerID: tenantData.CustomerID,
        Name:       tenantData.Name,
        Slug:       tenantData.Slug,
    }
    db.Create(&tenant)
}

// Dann Organizations mit TenantID
for _, orgData := range seedData.Organizations {
    org := entities.Organization{
        TenantID: orgData.TenantID,
        Name:     orgData.Name,
    }
    db.Create(&org)
}

// Zuletzt Users mit TenantID und OrganizationID
for _, userData := range seedData.Users {
    user := models.User{
        TenantID:       userData.TenantID,
        OrganizationID: userData.OrganizationID,
        Username:       userData.Username,
        Email:          userData.Email,
    }
    db.Create(&user)
}
```

## Auswirkungen

### Betroffene Module

- ✅ **base-server**: User, Tenant Models
- ✅ **organization**: Organization Model, Service, Handler
- ⚠️ **documents**: Templates, Documents (verwendet `OrganizationID` - sollte funktionieren)
- ⚠️ **client_management**: Sessions, Invoices (verwendet `TenantID` - sollte funktionieren)
- ⚠️ **calendar**: CalendarEntries (verwendet `TenantID` und `UserID` - sollte funktionieren)

### Breaking Changes

1. **API Änderungen**: Organization-Endpunkte benötigen keine `userID` mehr in Query-Filtern
2. **Authentifizierung**: User-Context muss jetzt auch `OrganizationID` enthalten
3. **Tenant-Erstellung**: Benötigt jetzt `CustomerID`
4. **User-Erstellung**: Benötigt jetzt `OrganizationID`

### Vorteile der neuen Struktur

1. **Klare Hierarchie**: Customer → Tenant → Organization → User
2. **Flexibilität**: Ein Customer kann mehrere Tenants verwalten
3. **Multi-Org Support**: Tenants können mehrere Organisationen haben
4. **Saubere Zuordnung**: Users sind eindeutig einer Organization zugeordnet

## Nächste Schritte

1. ✅ Models angepasst (Tenant, User, Organization)
2. ✅ Services angepasst (OrganizationService)
3. ✅ Handler angepasst (OrganizationHandler)
4. ⏳ Seed-Daten anpassen
5. ⏳ Tests aktualisieren
6. ⏳ API-Dokumentation aktualisieren (Swagger)
7. ⏳ Frontend anpassen (falls vorhanden)
8. ⏳ Migrations-SQL-Skript testen
