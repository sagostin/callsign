# CallSign API - Status

## ✅ Complete

### Core Infrastructure
- PostgreSQL + GORM migrations
- ESL Manager (6 services including BLF)
- Logging (Loki)
- CDR sync
- WebSocket Hub (notifications)
- Encryption (AES-256-GCM)
- Audit Log Middleware

### Authentication
- JWT + Refresh tokens
- RequirePermission middleware
- Role-based (system_admin, tenant_admin, user)
- Admin login + User login endpoints

### Tenant-Scoped API Endpoints
- Extensions (CRUD + status + call handling rules)
- Extension Profiles (CRUD + call handling rules)
- Devices (CRUD + provisioning + call control)
- Device Profiles/Templates
- Voicemail Boxes (CRUD + messages)
- IVR Menus (CRUD)
- Queues (CRUD)
- Ring Groups (CRUD + **distinct ring name**)
- Speed Dials (CRUD)
- Conferences (CRUD)
- Numbers/DIDs (CRUD)
- Routing (inbound/outbound + call blocks)
- Dial Plans (CRUD)
- Audio Library (CRUD + streaming)
- Music on Hold (CRUD)
- Feature Codes (CRUD + system codes)
- Time Conditions (CRUD)
- Holiday Lists (CRUD + sync)
- Call Flows (CRUD + toggle)
- CDR / Call Records (list, get, export)
- Audit Logs (list)
- Messaging / Conversations (list, get, send)
- Contacts (CRUD + sync + lookup)
- Chat (threads, rooms, queues)
- **Paging Groups (CRUD)** ✨
- **Provisioning Templates (CRUD)** ✨
- **Tenant Settings (General, Branding, SMTP, Messaging, Hospitality)** ✨
- **Tenant Media (Sounds & Music Overrides)** ✨

### Device Control API ✨
- `/api/devices/{mac}/hangup`
- `/api/devices/{mac}/transfer`
- `/api/devices/{mac}/hold`
- `/api/devices/{mac}/dial`
- `/api/devices/{mac}/call-status`

### Provisioning System ✨
- `/provisioning/{mac}/{filename}` - Public endpoint
- `/api/provision/{tenant}/{secret}/{mac}` - Secure endpoint
- Multi-vendor support (Polycom, Yealink, Grandstream, Cisco, etc.)
- Template variable substitution
- Tenant/device-specific variables

### Call Recording & Transcription ✨
- CallRecording model (multi-storage: local, S3, GCS, Azure)
- Transcription model with segments
- Multi-provider: Whisper, OpenAI, Google, AWS, Azure, Deepgram
- Diarization, sentiment, summary support
- RecordingConfig & TranscriptionConfig per tenant

### BLF/Presence (ESL Service)
- PRESENCE_PROBE handling
- DND, Forward, Voicemail, Call Flow status
- Extension presence tracking

### System Admin API
- Tenants (CRUD)
- Tenant Profiles (CRUD with live usage counts)
- Users (CRUD)
- Gateways (CRUD + status)
- Bridges (CRUD)
- SIP Profiles (CRUD + sync from disk)
- Sofia Live Control (status, restart, reload)
- Global Dial Plans (CRUD)
- ACLs (CRUD + nodes)
- Device Templates (CRUD)
- Device Manufacturers (CRUD)
- Firmware (CRUD + upload + set default)
- Messaging Providers (CRUD)
- System Settings/Logs
- System Media (Sounds & Music)
- Security (Banned IPs)
- Config Inspector (XML debug, file browser)

### User Portal API
- GetUserDevices
- GetUserCallHistory
- GetUserVoicemail
- GetUserSettings / UpdateUserSettings
- GetUserContacts / CreateUserContact

### Out-of-Box Seeding
- Default admin, tenant profiles
- Outbound routes, sounds, feature codes, chatplans

### FreeSWITCH Integration
- mod_xml_curl (directory, dialplan, configuration)
- mod_xml_cdr (CDR ingestion)
- Cache management (flush, stats)

## 🔧 Remaining
- Fax API (handlers stub exists)
- Transcription service worker (processing queue)
- WebSocket event broadcasting expansion
- Reports/Analytics expansion
- Hospitality module (wake-up calls, room management)
- Call Broadcast campaigns
- Billing integration