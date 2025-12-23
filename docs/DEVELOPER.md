# CallSign Developer Guide

This guide covers the architecture, patterns, and key functions for developers working on the CallSign PBX platform.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Backend (Go API)](#backend-go-api)
- [Frontend (Vue.js)](#frontend-vuejs)
- [FreeSWITCH Integration](#freeswitch-integration)
- [Database Models](#database-models)
- [Authentication & Authorization](#authentication--authorization)
- [Adding New Features](#adding-new-features)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                               Caddy (Reverse Proxy)                          │
│                          TLS Termination + WebSocket                         │
└──────────────────────────────┬───────────────────────────────────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
┌───────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue.js UI   │    │   Go Iris API   │    │   FreeSWITCH    │
│   (Vite)      │◄──►│   (Port 8080)   │◄──►│   (ESL/XML)     │
└───────────────┘    └────────┬────────┘    └─────────────────┘
                              │
                     ┌────────▼────────┐
                     │   PostgreSQL    │
                     │   (GORM ORM)    │
                     └─────────────────┘
```

### Tech Stack

| Layer | Technology |
|-------|------------|
| Web Server | Caddy |
| Frontend | Vue 3 + Vite |
| API Framework | Go Iris |
| ORM | GORM |
| Database | PostgreSQL |
| PBX Engine | FreeSWITCH |
| Auth | JWT (HS256) |

---

## Backend (Go API)

### Directory Structure

```
api/
├── config/           # Configuration loading (.env)
├── handlers/         # HTTP request handlers
│   └── freeswitch/   # XML CURL handlers
├── middleware/       # Auth, CORS, audit, tenant
├── models/           # GORM database models
├── router/           # Route definitions
├── services/         # Business logic
│   ├── esl/          # FreeSWITCH ESL services
│   ├── encryption/   # AES-256-GCM encryption
│   ├── logging/      # Loki log manager
│   └── xmlcache/     # XML response caching
└── utils/            # Helper functions
```

### Handler Pattern

All handlers are methods on the `Handler` struct:

```go
// api/handlers/handlers.go
type Handler struct {
    DB         *gorm.DB
    Config     *config.Config
    Auth       *middleware.AuthMiddleware
    ESLManager *esl.Manager
    LogManager *logging.LogManager
}

// Create handler
h := handlers.NewHandler(db, cfg)
```

### Creating a New Handler

```go
// api/handlers/my_handlers.go
package handlers

// ListMyResources lists all resources for the current tenant
func (h *Handler) ListMyResources(ctx iris.Context) {
    // Get tenant from context (set by middleware)
    tenantID := middleware.GetTenantID(ctx)
    
    var resources []models.MyResource
    if err := h.DB.Where("tenant_id = ?", tenantID).Find(&resources).Error; err != nil {
        ctx.StatusCode(http.StatusInternalServerError)
        ctx.JSON(iris.Map{"error": "Database error"})
        return
    }
    
    ctx.JSON(iris.Map{"data": resources})
}

// CreateMyResource creates a new resource
func (h *Handler) CreateMyResource(ctx iris.Context) {
    var resource models.MyResource
    if err := ctx.ReadJSON(&resource); err != nil {
        ctx.StatusCode(http.StatusBadRequest)
        ctx.JSON(iris.Map{"error": "Invalid request payload"})
        return
    }
    
    resource.TenantID = middleware.GetTenantID(ctx)
    
    if err := h.DB.Create(&resource).Error; err != nil {
        ctx.StatusCode(http.StatusInternalServerError)
        ctx.JSON(iris.Map{"error": "Failed to create resource"})
        return
    }
    
    ctx.StatusCode(http.StatusCreated)
    ctx.JSON(iris.Map{"data": resource, "message": "Resource created"})
}
```

### Registering Routes

```go
// api/router/router.go
func (r *Router) Init() {
    // Tenant-scoped routes (require auth + tenant)
    tenantScoped := protected.Party("")
    tenantScoped.Use(r.Tenant.RequireTenant())
    {
        myResources := tenantScoped.Party("/my-resources")
        {
            myResources.Get("/", r.Handler.ListMyResources)
            myResources.Post("/", r.Handler.CreateMyResource)
            myResources.Get("/{id}", r.Handler.GetMyResource)
            myResources.Put("/{id}", r.Handler.UpdateMyResource)
            myResources.Delete("/{id}", r.Handler.DeleteMyResource)
        }
    }
}
```

### Key Middleware Functions

```go
// Get authenticated user's claims
claims := middleware.GetClaims(ctx)  // Returns *Claims or nil

// Get current user ID
userID := middleware.GetUserID(ctx)  // Returns uint

// Get current tenant ID (handles X-Tenant-ID header for system admins)
tenantID := middleware.GetTenantID(ctx)  // Returns uint

// Get user role
role := middleware.GetRole(ctx)  // Returns models.UserRole
```

---

## Database Models

### Creating a New Model

```go
// api/models/my_resource.go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type MyResource struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    UUID      uuid.UUID      `json:"uuid" gorm:"type:uuid;uniqueIndex"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    
    // Tenant association (required for multi-tenancy)
    TenantID uint   `json:"tenant_id" gorm:"index;not null"`
    Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID"`
    
    // Resource fields
    Name        string `json:"name" gorm:"not null"`
    Description string `json:"description"`
    Enabled     bool   `json:"enabled" gorm:"default:true"`
}

// BeforeCreate generates UUID
func (r *MyResource) BeforeCreate(tx *gorm.DB) error {
    r.UUID = uuid.New()
    return nil
}
```

### Key Models

| Model | Description |
|-------|-------------|
| `User` | System/tenant users with roles |
| `Tenant` | Multi-tenant organizations |
| `Extension` | SIP extensions with call settings |
| `Device` | Provisioned phones |
| `IVRMenu` | Auto-attendant menus |
| `Queue` | Call center queues |
| `Gateway` | SIP trunks |
| `Dialplan` | Call routing rules |

---

## Authentication & Authorization

### Roles

```go
const (
    RoleSystemAdmin models.UserRole = "system_admin"  // Full access
    RoleTenantAdmin models.UserRole = "tenant_admin"  // Tenant management
    RoleUser        models.UserRole = "user"          // End user
)
```

### JWT Claims

```go
type Claims struct {
    UserID   uint            `json:"user_id"`
    Username string          `json:"username"`
    Email    string          `json:"email"`
    Role     models.UserRole `json:"role"`
    TenantID *uint           `json:"tenant_id,omitempty"`
    jwt.RegisteredClaims
}
```

### Permission Middleware

```go
// Require system admin only
system.Use(r.Auth.RequireSystemAdmin())

// Require tenant admin or higher
tenantAdmin.Use(r.Auth.RequireTenantAdmin())

// Require specific permissions
route.Use(auth.RequirePermission(models.PermExtensionManage))

// Require ALL permissions
route.Use(auth.RequireAllPermissions(models.PermRecordingView, models.PermRecordingDelete))
```

---

## Frontend (Vue.js)

### Directory Structure

```
ui/src/
├── components/       # Reusable components
│   ├── common/       # DataTable, StatusBadge, etc.
│   └── layout/       # Sidebar, TopBar
├── views/            # Page components
│   ├── admin/        # Tenant admin views
│   ├── system/       # System admin views
│   └── user/         # User portal views
├── services/         # API communication
├── styles/           # Global CSS
└── router.js         # Route definitions
```

### API Service Pattern

```javascript
// ui/src/services/api.js

// All APIs follow CRUD pattern:
export const myResourcesAPI = {
    list: (params) => api.get('/my-resources', { params }),
    get: (id) => api.get(`/my-resources/${id}`),
    create: (data) => api.post('/my-resources', data),
    update: (id, data) => api.put(`/my-resources/${id}`, data),
    delete: (id) => api.delete(`/my-resources/${id}`),
}
```

### Using in Components

```vue
<script setup>
import { ref, onMounted } from 'vue'
import { myResourcesAPI } from '@/services/api'

const resources = ref([])
const loading = ref(false)

onMounted(async () => {
    loading.value = true
    try {
        const { data } = await myResourcesAPI.list()
        resources.value = data.data
    } catch (err) {
        console.error('Failed to load resources:', err)
    } finally {
        loading.value = false
    }
})

const createResource = async (resource) => {
    await myResourcesAPI.create(resource)
    // Refresh list...
}
</script>
```

### Key Frontend APIs

| API Object | Purpose |
|------------|---------|
| `authAPI` | Login, logout, profile |
| `extensionsAPI` | Extension CRUD |
| `devicesAPI` | Device management + call control |
| `systemAPI` | System admin operations |
| `tenantSettingsAPI` | Tenant configuration |

---

## FreeSWITCH Integration

### XML CURL Flow

```
FreeSWITCH ──POST──► /api/freeswitch/directory   ──► SIP Auth
FreeSWITCH ──POST──► /api/freeswitch/dialplan    ──► Call Routing
FreeSWITCH ──POST──► /api/freeswitch/configuration ──► Module Configs
FreeSWITCH ──POST──► /api/freeswitch/cdr         ──► Call Records
```

### ESL Services

Each ESL service handles a specific call type:

```go
// api/services/esl/
services/esl/
├── manager.go        # Service orchestration
├── callcontrol/      # General call routing
├── voicemail/        # Voicemail handling
├── queue/            # Call center queues
└── conference/       # Conference rooms
```

---

## Adding New Features

### Checklist

1. **Model**: Create `api/models/my_feature.go`
2. **Handlers**: Create `api/handlers/my_feature_handlers.go`
3. **Routes**: Add routes in `api/router/router.go`
4. **Frontend API**: Add to `ui/src/services/api.js`
5. **Vue Component**: Create in `ui/src/views/`
6. **Route**: Add to `ui/src/router.js`
7. **Sidebar**: Update `ui/src/components/layout/Sidebar.vue`

### Testing

```bash
# Run API tests
cd api
go test ./...

# Run with verbose
go test -v ./models/...
```

### Environment Variables

See `.env.example` for all configuration options:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=callsign
DB_USER=callsign
DB_PASSWORD=secret

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24  # hours

# FreeSWITCH
FREESWITCH_HOST=localhost
FREESWITCH_PORT=8021
FREESWITCH_PASSWORD=ClueCon
```

---

## Common Patterns

### Error Handling

```go
// Handler error response
ctx.StatusCode(http.StatusBadRequest)
ctx.JSON(iris.Map{"error": "Descriptive error message"})
return

// Success with data
ctx.JSON(iris.Map{"data": resource, "message": "Operation successful"})
```

### Pagination

```go
// Query params: ?page=1&limit=50
page, _ := strconv.Atoi(ctx.URLParamDefault("page", "1"))
limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "50"))
offset := (page - 1) * limit

var total int64
h.DB.Model(&models.Resource{}).Where("tenant_id = ?", tenantID).Count(&total)

var resources []models.Resource
h.DB.Where("tenant_id = ?", tenantID).
    Offset(offset).
    Limit(limit).
    Find(&resources)

ctx.JSON(iris.Map{
    "data":  resources,
    "total": total,
    "page":  page,
    "limit": limit,
})
```

### Tenant Scoping

All tenant-scoped queries MUST include tenant_id:

```go
// Correct - scoped to tenant
h.DB.Where("tenant_id = ?", tenantID).Find(&resources)

// Wrong - could leak data between tenants
h.DB.Find(&resources)
```
