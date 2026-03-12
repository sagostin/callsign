# CallSign PBX - Roadmap

> **Last Updated**: March 2026  
> A comprehensive development roadmap for the CallSign multi-tenant cloud PBX platform.

---

## Project Summary

| Metric | Current Status |
|--------|---------------|
| **API Endpoints** | ~400+ of ~420 implemented (95%+) |
| **Vue.js Views** | 100 views (64 admin, 26 system, 8 user, 2 auth) |
| **Backend Handlers** | 29 handler files operational |
| **Backend Models** | 40 model files |
| **Backend Services** | 9 service packages |
| **FreeSWITCH Integration** | mod_xml_curl + mod_xml_cdr + ESL + TTS + Fax |
| **Docker Deployment** | ✅ Production ready |
| **Multi-Tenant Support** | ✅ Full tenant isolation |

---

## Current Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CallSign Platform                          │
├─────────────────────────────────────────────────────────────────────┤
│  Frontend: Vue.js 3 + Vite    │  Backend: Go Fiber + PostgreSQL     │
│  Styling: CSS + Inter Font     │  Telephony: FreeSWITCH ESL          │
│  Reverse Proxy: Caddy          │  Auth: JWT + RBAC                    │
│  CDR Analytics: ClickHouse     │  Logging: Loki                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## ✅ Phase 1 - Core Platform (COMPLETE)

### Authentication & Authorization
- [x] JWT token-based authentication
- [x] Role-based access control (system_admin, tenant_admin, user)
- [x] Refresh token support
- [x] Password reset flow
- [x] Audit logging middleware
- [x] Admin login endpoint
- [x] Extension login endpoint
- [x] Permission-based middleware (RequirePermission, RequireAllPermissions)

### Multi-Tenant Foundation
- [x] Tenant CRUD operations
- [x] Tenant profiles with limits (extensions, devices, queues, etc.)
- [x] Tenant-scoped data isolation
- [x] Tenant branding & settings
- [x] SMTP configuration per tenant
- [x] Messaging configuration per tenant
- [x] Hospitality configuration per tenant

### Extensions & Devices
- [x] Extension CRUD with profiles
- [x] Extension status (SIP registration, in-call)
- [x] Call handling rules (CRUD + reorder)
- [x] Extension profiles with call handling rules
- [x] Device CRUD with multi-vendor provisioning
- [x] Device profiles & templates
- [x] Device call control (dial, hangup, transfer, hold, call-status)
- [x] Firmware management
- [x] Client registrations (provision, assign, unassigned list)

### Voice Features
- [x] Voicemail boxes & messages (CRUD + streaming + mark read)
- [x] IVR menus (CRUD)
- [x] Call queues (CRUD + agent management: add/remove/pause/unpause)
- [x] Ring groups (CRUD)
- [x] Conferences (CRUD + live control: mute/kick/lock/record/floor + stats/sessions)
- [x] Speed dials (CRUD)
- [x] Feature codes (CRUD + system codes)
- [x] Paging groups (CRUD)

### Routing & Numbers
- [x] Phone numbers/DIDs management (CRUD)
- [x] Inbound routes (full CRUD + reorder)
- [x] Outbound routes (full CRUD + reorder + default US/CAN routes)
- [x] Dial plans (CRUD)
- [x] Call blocks (CRUD)
- [x] Time conditions (CRUD)
- [x] Holiday lists (CRUD + sync)
- [x] Call flows with toggle (CRUD)
- [x] Dial code collision checking
- [x] Route debugging

### FreeSWITCH Integration
- [x] mod_xml_curl (directory, dialplan, configuration)
- [x] mod_xml_cdr (CDR ingestion)
- [x] ESL Manager with 6+ services (call control, voicemail, queue, conference, feature codes, BLF)
- [x] XML cache with flush/stats
- [x] BLF/Presence support
- [x] TTS caching service
- [x] SIP profile import/sync from disk

### System Administration
- [x] Gateway/trunk management (CRUD + status)
- [x] SIP profiles with Sofia control (status, restart, start, stop, reload)
- [x] Bridges (CRUD)
- [x] Global dial plans (CRUD)
- [x] ACLs with nodes (CRUD)
- [x] System settings & logs
- [x] Config inspector (XML debug, file browser)
- [x] Banned IP management
- [x] System media (sounds & music streaming)
- [x] Device templates (system-level CRUD)
- [x] Device manufacturers (CRUD)
- [x] System status & stats
- [x] FreeSWITCH console WebSocket

### Messaging & Communication
- [x] SMS/MMS conversations (list, get, send)
- [x] Messaging providers (system-level CRUD)
- [x] Messaging numbers (system-level CRUD)
- [x] SMS number management (tenant-level config/assign)
- [x] Contacts (CRUD + sync + lookup)
- [x] Chat system (threads, rooms, queues)
- [x] Webhook handlers (Telnyx inbound/status)

### Recordings & Media
- [x] Call recordings (list, get, delete, stream, download, notes, transcription, config)
- [x] Audio library (CRUD + streaming)
- [x] Music on Hold streams (CRUD)
- [x] Tenant media overrides (sounds & music)
- [x] CDR / Call Records (list, get, export)
- [x] Audit logs

### Reports & Analytics
- [x] Call volume reports
- [x] Agent performance metrics
- [x] Queue statistics
- [x] Extension usage reports
- [x] KPI dashboard
- [x] Number usage statistics
- [x] Report export

### Specialized Modules
- [x] Fax server (boxes CRUD, jobs, send/retry, endpoints, active/stats)
- [x] Call broadcast campaigns (CRUD + start/stop/stats)
- [x] Hospitality (rooms CRUD, check-in/out, wake-up calls)
- [x] E911 Locations (CRUD)
- [x] Operator panel data endpoint

### User / Extension Portal
- [x] User portal (devices, call history, voicemail, settings, contacts)
- [x] Extension portal (devices, call history, voicemail, settings, password, contacts)

### Live Operations
- [x] Start/stop call recording
- [x] Active calls data
- [x] Live queue statistics
- [x] Device registrations
- [x] Wake-up call scheduling via ESL

### Infrastructure
- [x] ClickHouse CDR sync (periodic PG → ClickHouse, cleanup)
- [x] WebSocket hub (real-time notifications, event broadcasting)
- [x] Encryption service (AES-256-GCM for SIP passwords)
- [x] Loki logging integration
- [x] Graceful shutdown handling

---

## 🔄 Phase 2 - Refinement & Expansion (IN PROGRESS)

### Remaining API Gaps
- [ ] Tenant usage statistics endpoint
- [ ] Time condition status checking (is currently matched?)
- [ ] Number usage statistics endpoint
- [ ] Voicemail delivery attempts tracking
- [ ] Voicemail system settings/status
- [ ] Messaging provider testing endpoint
- [ ] Mark conversation as read
- [ ] Delete conversations
- [ ] Audio library record via TTS or phone
- [ ] Music on Hold file add/remove per stream
- [ ] Update/delete user contacts (user portal)
- [ ] Conference profiles CRUD

### Device Enhancements
- [ ] Device reboot command
- [ ] Device logs viewing
- [ ] Device registration status check
- [ ] Network device scanning

### Real-Time Features
- [ ] Operator panel WebSocket
- [ ] Queue real-time WebSocket
- [ ] Conference real-time WebSocket
- [ ] Device status WebSocket
- [ ] Enhanced notification events

---

## 📅 Phase 3 - Future (PLANNED)

### Transcription Service
- [ ] Processing queue worker
- [ ] Multi-provider support (Whisper, OpenAI, Google, AWS, Azure, Deepgram)
- [ ] Diarization & sentiment analysis
- [ ] Transcription summaries

### Billing Integration
- [ ] Usage tracking
- [ ] Invoice generation
- [ ] Payment processing
- [ ] Billing reports

### Multi-Language
- [ ] Multi-language phrase system
- [ ] Locale-specific sound packs

---

## 🎨 UI/UX Improvements

### Current Theme
- Premium OSS aesthetic with Inter font
- Sophisticated blue primary (#3b82f6)
- Light slate backgrounds
- Responsive grid system (1-3 columns)

### Planned Improvements
- [ ] Dark mode toggle
- [ ] Mobile-optimized views
- [ ] Accessibility enhancements
- [ ] Loading state improvements
- [ ] Toast notification system

### Landing Page
- ✅ Modern light theme with indigo/lavender gradients
- ✅ Feature grid with colorful icons
- ✅ Pricing section (Self-Hosted, Professional Support, Managed)
- ✅ Community & roadmap sections

---

## 📊 API Completion by Category

| Category | Completed | Total | Progress |
|----------|-----------|-------|----------|
| Authentication | 8 | 8 | ████████████ 100% |
| Tenants | 17 | 18 | ███████████░ 94% |
| Extensions | 21 | 21 | ████████████ 100% |
| IVR/Time Conditions | 17 | 18 | ███████████░ 94% |
| Call Routing | 20 | 21 | ███████████░ 95% |
| Devices | 40 | 44 | ███████████░ 91% |
| Client Registrations | 6 | 6 | ████████████ 100% |
| Queues/Ring Groups | 15 | 16 | ███████████░ 94% |
| Conferencing | 19 | 23 | ██████████░░ 83% |
| Voicemail | 10 | 16 | ████████░░░░ 63% |
| Fax | 15 | 15 | ████████████ 100% |
| Messaging/SMS | 14 | 16 | ██████████░░ 88% |
| Recordings | 8 | 10 | ██████████░░ 80% |
| Audio/Music | 17 | 20 | ██████████░░ 85% |
| Reports | 7 | 7 | ████████████ 100% |
| User Portal | 7 | 9 | ██████████░░ 78% |
| Extension Portal | 8 | 8 | ████████████ 100% |
| System Admin | 55 | 55 | ████████████ 100% |
| Hospitality | 8 | 8 | ████████████ 100% |
| Locations | 5 | 5 | ████████████ 100% |
| Broadcast | 8 | 8 | ████████████ 100% |
| Live Operations | 6 | 6 | ████████████ 100% |
| WebSocket | 3 | 7 | █████░░░░░░░ 43% |
| Webhooks | 2 | 2 | ████████████ 100% |
| **Total** | **~400+** | **~420** | **███████████░ 95%** |

---

## Priority Matrix

### 🔴 High Priority
1. **WebSocket Expansion** - Real-time operator panel, queue/conference feeds
2. **Voicemail Delivery Tracking** - Reliability improvements
3. **Conference Profiles** - Multi-profile support

### 🟡 Medium Priority
1. **Transcription Worker** - Process recordings queue
2. **Device Reboot/Logs** - Enhanced device management
3. **Dark Mode** - UI theme toggle

### 🟢 Lower Priority
1. **Billing Integration** - Usage-based billing
2. **Multi-Language Phrases** - Internationalization
3. **Mobile-Optimized Views** - Responsive improvements

---

## Contributing

See [DEVELOPER.md](docs/DEVELOPER.md) for development setup and contribution guidelines.

## Documentation Index

| Document | Purpose |
|----------|---------|
| [PROJECT_STATUS.md](PROJECT_STATUS.md) | Current completion status |
| [docs/BACKEND_TODO.md](docs/BACKEND_TODO.md) | Detailed API endpoint checklist |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System architecture overview |
| [docs/API.md](docs/API.md) | API reference documentation |
| [docs/CALLFLOW.md](docs/CALLFLOW.MD) | Call flow diagrams |
| [docs/DEVELOPER.md](docs/DEVELOPER.md) | Developer guide |

---

## License

See [LICENSE.md](LICENSE.md) for licensing information.
