# CallSign API - Status

## ✅ Complete

### Core Infrastructure
- PostgreSQL + GORM migrations
- ESL Manager (6+ services including BLF)
- Logging (Loki)
- CDR sync (PostgreSQL + ClickHouse)
- WebSocket Hub (notifications + event broadcasting)
- Encryption (AES-256-GCM)
- Audit Log Middleware
- TTS Caching Service
- Fax Manager (queue processing, retry strategy)
- Messaging Manager (SMS/MMS, webhook-based)

### Authentication
- JWT + Refresh tokens
- RequirePermission middleware
- Role-based (system_admin, tenant_admin, user)
- Admin login + User login + Extension login endpoints
- Password reset flow

### Tenant-Scoped API Endpoints
- Extensions (CRUD + status + call handling rules)
- Extension Profiles (CRUD + call handling rules)
- Devices (CRUD + provisioning + call control)
- Device Profiles/Templates
- Client Registrations (provision, assign, unassigned list)
- Voicemail Boxes (CRUD + messages + streaming)
- IVR Menus (CRUD)
- Queues (CRUD + agent management: add/remove/pause/unpause)
- Ring Groups (CRUD + distinct ring name)
- Speed Dials (CRUD)
- Conferences (CRUD + live control: mute/kick/lock/record/floor + stats/sessions)
- Numbers/DIDs (CRUD)
- Routing (full inbound/outbound CRUD + reorder + call blocks + debug)
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
- SMS Number Management (config, assign, unassign)
- Contacts (CRUD + sync + lookup)
- Chat (threads, rooms, queues)
- Paging Groups (CRUD)
- Provisioning Templates (CRUD)
- Tenant Settings (General, Branding, SMTP, Messaging, Hospitality)
- Tenant Media (Sounds & Music Overrides)
- Fax (boxes, jobs, endpoints, send, retry, stats, download)
- Reports & Analytics (call-volume, agent-performance, queue-stats, extension-usage, kpi, number-usage, export)
- Hospitality (rooms CRUD, check-in/out, wake-up calls)
- Call Broadcast (campaigns CRUD, start/stop, stats)
- E911 Locations (CRUD)
- Recordings (list, get, delete, stream, download, notes, transcription, config)

### Device Control API
- `/api/devices/{mac}/hangup`
- `/api/devices/{mac}/transfer`
- `/api/devices/{mac}/hold`
- `/api/devices/{mac}/dial`
- `/api/devices/{mac}/call-status`

### Provisioning System
- `/provisioning/{mac}/{filename}` - Public endpoint
- `/api/provision/{tenant}/{secret}/{mac}` - Secure endpoint
- Multi-vendor support (Polycom, Yealink, Grandstream, Cisco, etc.)
- Template variable substitution
- Tenant/device-specific variables

### Call Recording & Transcription
- CallRecording model (multi-storage: local, S3, GCS, Azure)
- Transcription model with segments
- Multi-provider: Whisper, OpenAI, Google, AWS, Azure, Deepgram
- Diarization, sentiment, summary support
- RecordingConfig & TranscriptionConfig per tenant
- Stream, download, notes endpoints

### BLF/Presence (ESL Service)
- PRESENCE_PROBE handling
- DND, Forward, Voicemail, Call Flow status
- Extension presence tracking

### Live Operations
- Start/stop call recording via ESL
- Active calls data
- Live queue statistics
- Device registrations
- Wake-up call scheduling via ESL

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
- Messaging Numbers (CRUD)
- System Settings/Logs
- System Media (Sounds & Music)
- Security (Banned IPs)
- Config Inspector (XML debug, file browser)
- System Status & Stats

### User Portal API
- GetUserDevices
- GetUserCallHistory
- GetUserVoicemail
- GetUserSettings / UpdateUserSettings
- GetUserContacts / CreateUserContact

### Extension Portal API
- GetExtensionDevices
- GetExtensionCallHistory
- GetExtensionVoicemail
- GetExtensionSettings / UpdateExtensionSettings
- ChangeExtensionPassword
- GetExtensionContacts / CreateExtensionContact

### Webhooks
- Telnyx inbound SMS handler
- Telnyx delivery status handler

### Operator Panel
- Operator panel data endpoint

### Out-of-Box Seeding
- Default admin, tenant profiles
- Outbound routes, sounds, feature codes, chatplans

### FreeSWITCH Integration
- mod_xml_curl (directory, dialplan, configuration)
- mod_xml_cdr (CDR ingestion)
- Cache management (flush, stats)

## 🔧 Remaining
- Voicemail delivery tracking & system settings
- Conference profiles CRUD
- Transcription service worker (processing queue)
- WebSocket event broadcasting expansion (operator panel, queue, conference, device feeds)
- Device reboot/logs/status endpoints
- Billing integration
- Multi-language phrases