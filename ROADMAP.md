# CallSign PBX - Roadmap

> **Last Updated**: January 2026  
> A comprehensive development roadmap for the CallSign multi-tenant cloud PBX platform.

---

## Project Summary

| Metric | Current Status |
|--------|---------------|
| **API Endpoints** | ~280 of ~375 implemented (75%) |
| **Vue.js Views** | 100 views (64 admin, 26 system, 8 user, 2 auth) |
| **Backend Handlers** | 18 handler files operational |
| **FreeSWITCH Integration** | mod_xml_curl + mod_xml_cdr + ESL foundation complete |
| **Docker Deployment** | ✅ Production ready |
| **Multi-Tenant Support** | ✅ Full tenant isolation |

---

## Current Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CallSign Platform                          │
├─────────────────────────────────────────────────────────────────────┤
│  Frontend: Vue.js 3 + Vite    │  Backend: Go Iris + PostgreSQL      │
│  Styling: CSS + Inter Font     │  Telephony: FreeSWITCH ESL          │
│  Reverse Proxy: Caddy          │  Auth: JWT + RBAC                    │
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

### Multi-Tenant Foundation
- [x] Tenant CRUD operations
- [x] Tenant profiles with limits (extensions, devices, queues, etc.)
- [x] Tenant-scoped data isolation
- [x] Tenant branding & settings
- [x] SMTP configuration per tenant

### Extensions & Devices
- [x] Extension CRUD with profiles
- [x] Call handling rules
- [x] Device CRUD with multi-vendor provisioning
- [x] Device profiles & templates
- [x] Device call control (dial, hangup, transfer, hold)
- [x] Firmware management

### Basic Voice Features
- [x] Voicemail boxes & messages
- [x] IVR menus
- [x] Call queues
- [x] Ring groups
- [x] Conferences (basic CRUD)
- [x] Speed dials
- [x] Feature codes

### Routing & Numbers
- [x] Phone numbers/DIDs management
- [x] Inbound routes (list, create)
- [x] Outbound routes (list, create, defaults)
- [x] Dial plans
- [x] Call blocks
- [x] Time conditions
- [x] Holiday lists
- [x] Call flows with toggle

### FreeSWITCH Integration
- [x] mod_xml_curl (directory, dialplan, configuration)
- [x] mod_xml_cdr (CDR ingestion)
- [x] ESL Manager with 6 services
- [x] XML cache with flush/stats
- [x] BLF/Presence support

### System Administration
- [x] Gateway/trunk management
- [x] SIP profiles with Sofia control
- [x] Bridges
- [x] Global dial plans
- [x] ACLs with nodes
- [x] System settings & logs
- [x] Config inspector
- [x] Banned IP management

---

## 🔄 Phase 2 - Enhanced Features (IN PROGRESS)

### Call Routing Improvements
- [ ] Inbound routes: Get, Update, Delete, Reorder
- [ ] Outbound routes: Get, Update, Delete, Reorder
- [ ] Time condition status checking
- [ ] Route debugging tools

### Device Enhancements
- [ ] Device reboot command
- [ ] Device logs viewing
- [ ] Device registration status
- [ ] Network device scanning

### Queue & Conference Live Control
- [ ] Queue real-time statistics
- [ ] Queue agent management (add, remove, pause, unpause)
- [ ] Conference participants list
- [ ] Conference controls (mute, kick, lock, record)
- [ ] Conference profiles

### Voicemail Improvements
- [ ] Delivery attempts tracking
- [ ] Retry failed deliveries
- [ ] Voicemail system settings
- [ ] System status monitoring

### Messaging Enhancements
- [ ] Mark conversation as read
- [ ] Delete conversations
- [ ] Provider testing endpoint

### Audio & Media
- [ ] Get audio file metadata
- [ ] Record audio (TTS or phone)
- [ ] Add/remove music on hold files
- [ ] Recording streaming & download
- [ ] Recording email & export

### User Portal Expansion
- [ ] Update/delete user contacts
- [ ] Enhanced user settings

---

## 📅 Phase 3 - Advanced Capabilities (PLANNED)

### Real-Time Features
- [ ] Operator panel WebSocket
- [ ] Queue real-time WebSocket
- [ ] Conference real-time WebSocket
- [ ] Device status WebSocket
- [ ] Enhanced notification events

### Reports & Analytics
- [ ] Call volume reports
- [ ] Agent performance metrics
- [ ] Queue statistics dashboard
- [ ] Extension usage reports
- [ ] KPI dashboard
- [ ] Report export (CSV/PDF)
- [ ] Number usage statistics

### Call Recording System
- [ ] Recording streaming
- [ ] Recording download
- [ ] Recording email
- [ ] Batch export
- [ ] Recording statistics

---

## 📅 Phase 4 - Specialized Modules (FUTURE)

### Fax Server
- [ ] Fax server management
- [ ] Send/receive fax
- [ ] Fax inbox/sent folders
- [ ] Fax download (PDF)
- [ ] Fax deletion

### Call Broadcast
- [ ] Campaign management
- [ ] Campaign scheduling
- [ ] Start/stop campaigns
- [ ] Campaign statistics

### Hospitality Module
- [ ] Room management
- [ ] Guest check-in/out
- [ ] Wake-up call scheduling
- [ ] Room status tracking

### E911 Locations
- [ ] Location management
- [ ] Extension-location mapping
- [ ] Emergency routing

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
| Authentication | 7 | 7 | ████████████ 100% |
| Tenants | 16 | 17 | ███████████░ 94% |
| Extensions | 21 | 21 | ████████████ 100% |
| IVR/Time Conditions | 17 | 18 | ███████████░ 94% |
| Call Routing | 13 | 21 | ████████░░░░ 62% |
| Devices | 40 | 45 | ██████████░░ 89% |
| Queues/Ring Groups | 10 | 16 | ████████░░░░ 63% |
| Conferencing | 5 | 18 | ███░░░░░░░░░ 28% |
| Voicemail | 10 | 16 | ████████░░░░ 63% |
| Fax | 0 | 11 | ░░░░░░░░░░░░ 0% |
| Messaging | 8 | 11 | █████████░░░ 73% |
| Recordings | 3 | 8 | █████░░░░░░░ 38% |
| Audio/Music | 16 | 20 | ██████████░░ 80% |
| Reports | 0 | 7 | ░░░░░░░░░░░░ 0% |
| User Portal | 7 | 9 | ██████████░░ 78% |
| System Admin | 48 | 48 | ████████████ 100% |
| Hospitality | 0 | 11 | ░░░░░░░░░░░░ 0% |
| WebSocket | 2 | 6 | ████░░░░░░░░ 33% |
| **Total** | **~280** | **~375** | **████████░░░░ 75%** |

---

## Priority Matrix

### 🔴 High Priority (Q1)
1. **Inbound/Outbound Route CRUD** - Complete routing management
2. **Queue Agent Management** - Essential for call center operations
3. **Conference Live Control** - Real-time meeting management
4. **Recording Streaming** - Access call recordings

### 🟡 Medium Priority (Q2)
1. **Reports Module** - Analytics and KPIs
2. **WebSocket Expansion** - Real-time operator panel
3. **Device Status/Logs** - Enhanced device management
4. **Voicemail Delivery Tracking** - Reliability improvements

### 🟢 Lower Priority (Q3-Q4)
1. **Fax Server** - Legacy support
2. **Hospitality Module** - Vertical market feature
3. **Call Broadcast** - Marketing campaigns
4. **E911 Locations** - Compliance feature

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

---

## License

See [LICENSE.md](LICENSE.md) for licensing information.
