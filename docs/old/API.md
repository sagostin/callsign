# CallSign API Documentation

## Overview

CallSign is a multi-tenant VoIP platform API with FreeSWITCH integration.

**Base URL**: `http://localhost:8080/api`

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

### Admin Login
```
POST /api/auth/admin/login
{
  "username": "admin",
  "password": "secret"
}
```

### Extension Login
```
POST /api/auth/extension/login
{
  "extension": "1001",
  "password": "secret",
  "tenant_domain": "example.com"
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

### Tenant Settings
```
GET    /api/tenant/settings
PUT    /api/tenant/settings
GET    /api/tenant/branding
PUT    /api/tenant/branding
GET    /api/tenant/smtp
PUT    /api/tenant/smtp
POST   /api/tenant/smtp/test
GET    /api/tenant/messaging
PUT    /api/tenant/messaging
GET    /api/tenant/hospitality
PUT    /api/tenant/hospitality
```

### E911 Locations
```
GET    /api/tenant/locations
POST   /api/tenant/locations
GET    /api/tenant/locations/{id}
PUT    /api/tenant/locations/{id}
DELETE /api/tenant/locations/{id}
```

### Extensions
```
GET    /api/extensions
POST   /api/extensions
GET    /api/extensions/{ext}
PUT    /api/extensions/{ext}
DELETE /api/extensions/{ext}
GET    /api/extensions/{ext}/status

# Call Handling Rules
GET    /api/extensions/{ext}/call-rules
POST   /api/extensions/{ext}/call-rules
PUT    /api/extensions/{ext}/call-rules/{ruleId}
DELETE /api/extensions/{ext}/call-rules/{ruleId}
POST   /api/extensions/{ext}/call-rules/reorder
```

### Extension Profiles
```
GET    /api/extension-profiles
POST   /api/extension-profiles
GET    /api/extension-profiles/{id}
PUT    /api/extension-profiles/{id}
DELETE /api/extension-profiles/{id}

# Profile Call Handling Rules
GET    /api/extension-profiles/{id}/call-rules
POST   /api/extension-profiles/{id}/call-rules
PUT    /api/extension-profiles/{id}/call-rules/{ruleId}
DELETE /api/extension-profiles/{id}/call-rules/{ruleId}
POST   /api/extension-profiles/{id}/call-rules/reorder
```

### Devices
```
GET    /api/devices
POST   /api/devices
GET    /api/devices/{id}
PUT    /api/devices/{id}
DELETE /api/devices/{id}
POST   /api/devices/{id}/reprovision
POST   /api/devices/{id}/assign-user
POST   /api/devices/{id}/assign-profile
PUT    /api/devices/{id}/lines

# Device Call Control
POST   /api/devices/{mac}/hangup
POST   /api/devices/{mac}/transfer
POST   /api/devices/{mac}/hold
POST   /api/devices/{mac}/dial
GET    /api/devices/{mac}/call-status
```

### Client Registrations
```
GET    /api/registrations
POST   /api/registrations/provision
DELETE /api/registrations/{id}
GET    /api/registrations/unassigned
POST   /api/registrations/{id}/assign
GET    /api/registrations/extension/{id}
```

### Device Profiles
```
GET    /api/device-profiles
POST   /api/device-profiles
GET    /api/device-profiles/{id}
PUT    /api/device-profiles/{id}
DELETE /api/device-profiles/{id}
```

### Feature Codes
```
GET    /api/feature-codes
POST   /api/feature-codes
GET    /api/feature-codes/{id}
PUT    /api/feature-codes/{id}
DELETE /api/feature-codes/{id}
GET    /api/feature-codes/system
```

### Conferences
```
GET    /api/conferences
POST   /api/conferences
GET    /api/conferences/{id}
PUT    /api/conferences/{id}
DELETE /api/conferences/{id}
GET    /api/conferences/{id}/stats
GET    /api/conferences/{id}/sessions
GET    /api/conferences/sessions/{sessionId}/participants

# Live Conference Control
GET    /api/conferences/live
GET    /api/conferences/live/{name}
POST   /api/conferences/live/{name}/mute/{memberId}
POST   /api/conferences/live/{name}/unmute/{memberId}
POST   /api/conferences/live/{name}/deaf/{memberId}
POST   /api/conferences/live/{name}/undeaf/{memberId}
POST   /api/conferences/live/{name}/kick/{memberId}
POST   /api/conferences/live/{name}/lock
POST   /api/conferences/live/{name}/unlock
POST   /api/conferences/live/{name}/record/start
POST   /api/conferences/live/{name}/record/stop
POST   /api/conferences/live/{name}/mute-all
POST   /api/conferences/live/{name}/unmute-all
POST   /api/conferences/live/{name}/floor/{memberId}
```

### Queues
```
GET    /api/queues
POST   /api/queues
GET    /api/queues/{id}
PUT    /api/queues/{id}
DELETE /api/queues/{id}

# Queue Agent Management
GET    /api/queues/{id}/agents
POST   /api/queues/{id}/agents
DELETE /api/queues/{id}/agents/{agentId}
POST   /api/queues/{id}/agents/{agentId}/pause
POST   /api/queues/{id}/agents/{agentId}/unpause
```

### Ring Groups
```
GET    /api/ring-groups
POST   /api/ring-groups
GET    /api/ring-groups/{id}
PUT    /api/ring-groups/{id}
DELETE /api/ring-groups/{id}
```

### IVR Menus
```
GET    /api/ivr/menus
POST   /api/ivr/menus
GET    /api/ivr/menus/{id}
PUT    /api/ivr/menus/{id}
DELETE /api/ivr/menus/{id}
```

### Time Conditions
```
GET    /api/time-conditions
POST   /api/time-conditions
GET    /api/time-conditions/{id}
PUT    /api/time-conditions/{id}
DELETE /api/time-conditions/{id}
```

### Holiday Lists
```
GET    /api/holidays
POST   /api/holidays
GET    /api/holidays/{id}
PUT    /api/holidays/{id}
DELETE /api/holidays/{id}
POST   /api/holidays/{id}/sync
```

### Call Flows
```
GET    /api/call-flows
POST   /api/call-flows
GET    /api/call-flows/{id}
PUT    /api/call-flows/{id}
DELETE /api/call-flows/{id}
POST   /api/call-flows/{id}/toggle
```

### Numbers/DIDs
```
GET    /api/numbers
POST   /api/numbers
GET    /api/numbers/{id}
PUT    /api/numbers/{id}
DELETE /api/numbers/{id}
```

### Routing
```
# Inbound Routes
GET    /api/routing/inbound
POST   /api/routing/inbound
GET    /api/routing/inbound/{id}
PUT    /api/routing/inbound/{id}
DELETE /api/routing/inbound/{id}
POST   /api/routing/inbound/reorder

# Outbound Routes
GET    /api/routing/outbound
POST   /api/routing/outbound
GET    /api/routing/outbound/{id}
PUT    /api/routing/outbound/{id}
DELETE /api/routing/outbound/{id}
POST   /api/routing/outbound/reorder
POST   /api/routing/outbound/defaults

# Call Blocks
GET    /api/routing/blocks
POST   /api/routing/blocks
PUT    /api/routing/blocks/{id}
DELETE /api/routing/blocks/{id}

# Debug
GET    /api/routing/debug

# Dial Code Check
POST   /api/check-dial-code
```

### Dial Plans
```
GET    /api/dial-plans
POST   /api/dial-plans
GET    /api/dial-plans/{id}
PUT    /api/dial-plans/{id}
DELETE /api/dial-plans/{id}
```

### Voicemail
```
GET    /api/voicemail/boxes
POST   /api/voicemail/boxes
GET    /api/voicemail/boxes/{ext}
PUT    /api/voicemail/boxes/{ext}
DELETE /api/voicemail/boxes/{ext}

# Messages
GET    /api/voicemail/boxes/{ext}/messages
GET    /api/voicemail/messages/{id}
DELETE /api/voicemail/messages/{id}
POST   /api/voicemail/messages/{id}/read
GET    /api/voicemail/messages/{id}/stream
```

### Recordings
```
GET    /api/recordings
GET    /api/recordings/config
GET    /api/recordings/{id}
DELETE /api/recordings/{id}
GET    /api/recordings/{id}/stream
GET    /api/recordings/{id}/download
PUT    /api/recordings/{id}/notes
GET    /api/recordings/{id}/transcription
```

### Audio Library
```
GET    /api/audio-library
POST   /api/audio-library
PUT    /api/audio-library/{id}
DELETE /api/audio-library/{id}
GET    /api/audio-library/{id}/stream
```

### Music on Hold
```
GET    /api/music-on-hold
POST   /api/music-on-hold
GET    /api/music-on-hold/{id}
PUT    /api/music-on-hold/{id}
DELETE /api/music-on-hold/{id}
```

### CDR
```
GET    /api/cdr
GET    /api/cdr/{id}
GET    /api/cdr/export
```

### Audit Logs
```
GET    /api/audit-logs
```

### Messaging (SMS/MMS)
```
GET    /api/messaging/conversations
GET    /api/messaging/conversations/{id}
POST   /api/messaging/send

# SMS Number Management
GET    /api/messaging/numbers
PUT    /api/messaging/numbers/{id}/sms
GET    /api/messaging/numbers/{id}/assignments
POST   /api/messaging/numbers/{id}/assignments
DELETE /api/messaging/numbers/{id}/assignments/{assignId}
```

### Contacts
```
GET    /api/contacts
POST   /api/contacts
GET    /api/contacts/{id}
PUT    /api/contacts/{id}
DELETE /api/contacts/{id}
POST   /api/contacts/{id}/sync
GET    /api/contacts/lookup
```

### Chat
```
GET    /api/chat/threads
POST   /api/chat/threads
GET    /api/chat/threads/{id}
POST   /api/chat/threads/{id}/messages
GET    /api/chat/rooms
POST   /api/chat/rooms
POST   /api/chat/rooms/{id}/join
GET    /api/chat/queues
POST   /api/chat/queues
```

### Paging Groups
```
GET    /api/page-groups
POST   /api/page-groups
GET    /api/page-groups/{id}
PUT    /api/page-groups/{id}
DELETE /api/page-groups/{id}
```

### Speed Dials
```
GET    /api/speed-dials
POST   /api/speed-dials
GET    /api/speed-dials/{id}
PUT    /api/speed-dials/{id}
DELETE /api/speed-dials/{id}
```

### Provisioning Templates
```
GET    /api/provisioning-templates
POST   /api/provisioning-templates
GET    /api/provisioning-templates/{id}
PUT    /api/provisioning-templates/{id}
DELETE /api/provisioning-templates/{id}
```

### Tenant Media
```
GET    /api/media/sounds
POST   /api/media/sounds
DELETE /api/media/sounds
GET    /api/media/music
POST   /api/media/music
DELETE /api/media/music
```

### Fax
```
# Fax Boxes
GET    /api/fax/boxes
POST   /api/fax/boxes
GET    /api/fax/boxes/{boxId}
PUT    /api/fax/boxes/{boxId}
DELETE /api/fax/boxes/{boxId}

# Fax Jobs
GET    /api/fax/jobs
GET    /api/fax/jobs/{jobId}
DELETE /api/fax/jobs/{jobId}
GET    /api/fax/jobs/{jobId}/download
POST   /api/fax/jobs/{jobId}/retry

# Fax Actions
POST   /api/fax/send
GET    /api/fax/active
GET    /api/fax/stats

# Fax Endpoints
GET    /api/fax/endpoints
POST   /api/fax/endpoints
PUT    /api/fax/endpoints/{epId}
DELETE /api/fax/endpoints/{epId}
```

### Reports & Analytics
```
GET    /api/reports/call-volume
GET    /api/reports/agent-performance
GET    /api/reports/queue-stats
GET    /api/reports/extension-usage
GET    /api/reports/kpi
GET    /api/reports/number-usage
GET    /api/reports/export
```

### Hospitality
```
GET    /api/hospitality/rooms
POST   /api/hospitality/rooms
GET    /api/hospitality/rooms/{id}
PUT    /api/hospitality/rooms/{id}
DELETE /api/hospitality/rooms/{id}
POST   /api/hospitality/rooms/{id}/checkin
POST   /api/hospitality/rooms/{id}/checkout
POST   /api/hospitality/rooms/{id}/wakeup
```

### Call Broadcast
```
GET    /api/broadcast
POST   /api/broadcast
GET    /api/broadcast/{id}
PUT    /api/broadcast/{id}
DELETE /api/broadcast/{id}
POST   /api/broadcast/{id}/start
POST   /api/broadcast/{id}/stop
GET    /api/broadcast/{id}/stats
```

### Operator Panel
```
GET    /api/operator-panel
```

### Live Operations
```
POST   /api/live/recording/start
POST   /api/live/recording/stop
GET    /api/live/calls
GET    /api/live/queue-stats
POST   /api/live/wakeup/schedule
GET    /api/live/registrations
```

---

## System Admin Endpoints

### Tenants
```
GET    /api/system/tenants
POST   /api/system/tenants
GET    /api/system/tenants/{id}
PUT    /api/system/tenants/{id}
DELETE /api/system/tenants/{id}
```

### Tenant Profiles
```
GET    /api/system/tenant-profiles
POST   /api/system/tenant-profiles
GET    /api/system/tenant-profiles/{id}
PUT    /api/system/tenant-profiles/{id}
DELETE /api/system/tenant-profiles/{id}
```

### System Numbers
```
GET    /api/system/numbers
```

### Gateways
```
GET    /api/system/gateways
POST   /api/system/gateways
GET    /api/system/gateways/{id}
PUT    /api/system/gateways/{id}
DELETE /api/system/gateways/{id}
GET    /api/system/gateways/status
```

### Bridges
```
GET    /api/system/bridges
POST   /api/system/bridges
GET    /api/system/bridges/{id}
PUT    /api/system/bridges/{id}
DELETE /api/system/bridges/{id}
```

### SIP Profiles
```
GET    /api/system/sip-profiles
POST   /api/system/sip-profiles
POST   /api/system/sip-profiles/sync
GET    /api/system/sip-profiles/{id}
PUT    /api/system/sip-profiles/{id}
DELETE /api/system/sip-profiles/{id}
```

### Sofia Control
```
GET    /api/system/sofia/status
GET    /api/system/sofia/profiles/{name}/status
GET    /api/system/sofia/profiles/{name}/registrations
GET    /api/system/sofia/profiles/{name}/gateways
POST   /api/system/sofia/profiles/{name}/restart
POST   /api/system/sofia/profiles/{name}/start
POST   /api/system/sofia/profiles/{name}/stop
POST   /api/system/sofia/reload-xml
```

### Global Dial Plans
```
GET    /api/system/dialplans
POST   /api/system/dialplans
GET    /api/system/dialplans/{id}
PUT    /api/system/dialplans/{id}
DELETE /api/system/dialplans/{id}
```

### ACLs
```
GET    /api/system/acls
POST   /api/system/acls
GET    /api/system/acls/{id}
PUT    /api/system/acls/{id}
DELETE /api/system/acls/{id}
POST   /api/system/acls/{id}/nodes
PUT    /api/system/acls/{id}/nodes/{nodeId}
DELETE /api/system/acls/{id}/nodes/{nodeId}
```

### System Media
```
GET    /api/system/media/sounds
POST   /api/system/media/sounds
GET    /api/system/media/sounds/stream
GET    /api/system/media/music
POST   /api/system/media/music
GET    /api/system/media/music/stream
```

### System Settings & Status
```
GET    /api/system/settings
PUT    /api/system/settings
GET    /api/system/logs
GET    /api/system/status
GET    /api/system/stats
```

### Security
```
GET    /api/system/security/banned-ips
POST   /api/system/security/banned-ips
DELETE /api/system/security/banned-ips/{ip}
```

### Device Templates
```
GET    /api/system/device-templates
POST   /api/system/device-templates
GET    /api/system/device-templates/{id}
PUT    /api/system/device-templates/{id}
DELETE /api/system/device-templates/{id}
```

### Device Manufacturers
```
GET    /api/system/device-manufacturers
POST   /api/system/device-manufacturers
PUT    /api/system/device-manufacturers/{id}
DELETE /api/system/device-manufacturers/{id}
```

### Firmware
```
GET    /api/system/firmware
POST   /api/system/firmware
GET    /api/system/firmware/{id}
PUT    /api/system/firmware/{id}
DELETE /api/system/firmware/{id}
POST   /api/system/firmware/{id}/upload
POST   /api/system/firmware/{id}/set-default
```

### Messaging Providers
```
GET    /api/system/messaging-providers
POST   /api/system/messaging-providers
GET    /api/system/messaging-providers/{id}
PUT    /api/system/messaging-providers/{id}
DELETE /api/system/messaging-providers/{id}
```

### Messaging Numbers
```
GET    /api/system/messaging-numbers
POST   /api/system/messaging-numbers
PUT    /api/system/messaging-numbers/{id}
DELETE /api/system/messaging-numbers/{id}
```

### Config Inspector
```
GET    /api/system/xml/debug
GET    /api/system/config/files
GET    /api/system/config/file
```

---

## User Portal Endpoints

```
GET    /api/user/devices
GET    /api/user/call-history
GET    /api/user/voicemail
GET    /api/user/settings
PUT    /api/user/settings
GET    /api/user/contacts
POST   /api/user/contacts
```

---

## Extension Portal Endpoints

```
GET    /api/extension/portal/devices
GET    /api/extension/portal/call-history
GET    /api/extension/portal/voicemail
GET    /api/extension/portal/settings
PUT    /api/extension/portal/settings
PUT    /api/extension/portal/password
GET    /api/extension/portal/contacts
POST   /api/extension/portal/contacts
```

---

## Tenant Users

```
GET    /api/users
POST   /api/users
GET    /api/users/{id}
PUT    /api/users/{id}
DELETE /api/users/{id}
```

---

## FreeSWITCH Integration

### XML Curl (mod_xml_curl)
```
POST /api/freeswitch/directory
POST /api/freeswitch/dialplan
POST /api/freeswitch/configuration
POST /api/freeswitch/xmlapi          # Legacy combined
```

### CDR (mod_xml_cdr)
```
POST /api/freeswitch/cdr
```

### Cache Management
```
GET /api/freeswitch/cache/flush
GET /api/freeswitch/cache/stats
```

---

## Webhooks

```
POST /api/webhooks/telnyx/inbound    # Telnyx inbound SMS
POST /api/webhooks/telnyx/status     # Telnyx delivery status
```

---

## Provisioning

```
GET /api/provision/{tenant}/{secret}/{mac}   # Secure (API)
GET /provisioning/{mac}/{filename}           # Public (device)
```

---

## Internal API

```
POST /api/internal/fail2ban/report   # X-Internal-Key auth
```

---

## WebSocket Endpoints

```
WSS /api/ws/notifications    # Real-time notifications
WSS /api/system/console      # FreeSWITCH console
WSS /api/ws                  # Generic WebSocket
```

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
