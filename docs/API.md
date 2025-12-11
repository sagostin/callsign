# CallSign API Documentation

## Overview

CallSign is a multi-tenant VoIP platform API with FreeSWITCH integration.

## Authentication

### JWT Token

All protected endpoints require a JWT Bearer token:
```
Authorization: Bearer <token>
```

### Login
```
POST /api/auth/login
{
  "username": "user@domain.com",
  "password": "secret"
}
```

Response:
```json
{
  "token": "eyJ...",
  "user": { ... }
}
```

---

## Roles & Permissions

| Role | Description |
|------|-------------|
| `system_admin` | Full system access |
| `tenant_admin` | Manage own tenant |
| `user` | Personal features only |

### Permission Middleware

```go
// Single permission
protected.Use(auth.RequirePermission(models.PermExtensionManage))

// Multiple (requires ANY)
protected.Use(auth.RequirePermission(models.PermUserCreate, models.PermUserManage))

// Requires ALL
protected.Use(auth.RequireAllPermissions(models.PermRecordingView, models.PermRecordingDelete))
```

---

## Tenant-Scoped Endpoints

### Extensions
```
GET    /api/extensions
POST   /api/extensions
GET    /api/extensions/{ext}
PUT    /api/extensions/{ext}
DELETE /api/extensions/{ext}
```

### Feature Codes
```
GET    /api/feature-codes          # List all
POST   /api/feature-codes          # Create (with validation)
GET    /api/feature-codes/{id}
PUT    /api/feature-codes/{id}     # Cannot modify system codes
DELETE /api/feature-codes/{id}     # Cannot delete system codes
GET    /api/feature-codes/system   # List reserved codes
```

### Conferences
```
GET    /api/conferences
POST   /api/conferences
GET    /api/conferences/{id}
PUT    /api/conferences/{id}
DELETE /api/conferences/{id}

# Live Control
GET    /api/conferences/live                    # Active conferences
GET    /api/conferences/live/{name}             # With members
POST   /api/conferences/live/{name}/lock
POST   /api/conferences/live/{name}/mute/{id}
POST   /api/conferences/live/{name}/kick/{id}
POST   /api/conferences/live/{name}/record
```

### Queues
```
GET    /api/queues
POST   /api/queues
GET    /api/queues/{id}
PUT    /api/queues/{id}
DELETE /api/queues/{id}
```

### IVR Menus
```
GET    /api/ivr/menus
POST   /api/ivr/menus
GET    /api/ivr/menus/{id}
PUT    /api/ivr/menus/{id}
DELETE /api/ivr/menus/{id}
```

---

## System Admin Endpoints

### Tenants
```
GET    /api/admin/tenants
POST   /api/admin/tenants
GET    /api/admin/tenants/{id}
PUT    /api/admin/tenants/{id}
DELETE /api/admin/tenants/{id}
```

### SIP Profiles
```
GET    /api/admin/sip-profiles
POST   /api/admin/sip-profiles
GET    /api/admin/sip-profiles/{id}
PUT    /api/admin/sip-profiles/{id}
```

### Gateways
```
GET    /api/admin/gateways
POST   /api/admin/gateways
GET    /api/admin/gateways/{id}
PUT    /api/admin/gateways/{id}
DELETE /api/admin/gateways/{id}
```

---

## FreeSWITCH Integration

### XML Curl (mod_xml_curl)
```
POST /freeswitch/xmlapi
```
Provides: directory, dialplan, configuration

### CDR (mod_xml_cdr)
```
POST /freeswitch/cdr
```
Receives call detail records

---

## ESL Services

| Service | Address | Function |
|---------|---------|----------|
| callcontrol | 127.0.0.1:9001 | Routing |
| voicemail | 127.0.0.2:9001 | VM |
| queue | 127.0.0.3:9001 | Call center |
| conference | 127.0.0.4:9001 | Conferences |
| featurecodes | 127.0.0.6:9001 | *XX codes |

---

## Feature Codes

### System Codes (Reserved)
| Code | Action |
|------|--------|
| *97 | Voicemail check |
| *72/*73 | Call forward on/off |
| *78/*79 | DND on/off |
| *30 | Call flow toggle |
| *70/*85 | Park/retrieve |

### Custom Codes
Tenants can create custom codes with actions:
- `webhook` - Call external URL
- `lua` - Run Lua script
- `transfer` - Transfer to destination
