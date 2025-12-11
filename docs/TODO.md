# CallSign API - Status

## ✅ Complete

### Core Infrastructure
- PostgreSQL + GORM migrations
- ESL Manager (6 services including BLF)
- Logging (Loki)
- CDR sync
- WebSocket Hub
- Encryption (AES-256-GCM)

### Authentication
- JWT + Refresh tokens
- RequirePermission middleware
- Role-based (system_admin, tenant_admin, user)

### Tenant-Scoped API Endpoints
- Extensions (CRUD + status)
- Devices (CRUD + call control)
- Voicemail Boxes (CRUD)
- IVR Menus (CRUD)
- Queues (CRUD)
- Ring Groups (CRUD + **distinct ring name**)
- Conferences (CRUD)
- Numbers/DIDs (CRUD)
- Routing (inbound/outbound)
- Dial Plans (CRUD)
- Audio Library (CRUD)
- Music on Hold (CRUD)
- Feature Codes (CRUD + system codes)
- Time Conditions (CRUD)
- Call Flows (CRUD + toggle)
- CDR / Call Records (list, get, export)
- Audit Logs (list)
- Messaging / Conversations (list, get, send)
- Contacts (CRUD + sync + lookup)
- Chat (threads, rooms, queues)
- **Paging Groups (CRUD)** ✨
- **Provisioning Templates (CRUD)** ✨

### Device Control API ✨
- `/api/devices/{mac}/hangup`
- `/api/devices/{mac}/transfer`
- `/api/devices/{mac}/hold`
- `/api/devices/{mac}/dial`
- `/api/devices/{mac}/call-status`

### Provisioning System ✨
- `/provisioning/{mac}/{filename}` - Public endpoint
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
- Tenant Profiles (CRUD)
- Gateways, Bridges, SIP Profiles (CRUD)
- System Settings/Logs

### Out-of-Box Seeding
- Default admin, tenant profiles
- Outbound routes, sounds, feature codes, chatplans

## 🔧 Remaining
- Call Block, Speed Dials, Fax API
- Transcription service worker (processing queue)
- WebSocket event broadcasting
- UI integration for new features