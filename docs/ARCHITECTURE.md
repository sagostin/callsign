# CallSign PBX — Architecture Documentation

## System Overview

CallSign is a multi-tenant cloud PBX platform built on FreeSWITCH. It provides a complete telephony management solution with three distinct web portals (user, tenant admin, system admin), a RESTful JSON API, and deep FreeSWITCH integration via both the Event Socket Layer (ESL) and XML CURL.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Caddy (Reverse Proxy)             │
│          :80/:443 — TLS termination, routing         │
│   /api/* → API (:8080)    /* → UI (:5173)            │
└───────────┬─────────────────────────┬───────────────┘
            │                         │
   ┌────────▼────────┐     ┌─────────▼─────────┐
   │   Go API Server  │     │  Vue 3 SPA (Vite)  │
   │     (:8080)       │     │     (:5173)         │
   │                   │     │                     │
   │  Fiber HTTP       │     │  3 Portal Layouts   │
   │  Handlers         │     │  100+ Views         │
   │  Middleware        │     │  Axios API Client   │
   │  Services          │     │  WebSocket Client   │
   └───┬───┬───┬───────┘     └─────────────────────┘
       │   │   │
       │   │   └─── ESL (Inbound + Outbound) ──► FreeSWITCH
       │   │
       │   └─── XML CURL ◄── FreeSWITCH (mod_xml_curl)
       │
  ┌────▼────────────────────────────────────┐
  │            Data Stores                   │
  │  PostgreSQL (primary)    ClickHouse (CDR)│
  │  Redis (cache/sessions)  Loki (logs)     │
  └──────────────────────────────────────────┘
```

### Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| Reverse Proxy | Caddy 2 | TLS termination, path-based routing |
| Frontend | Vue 3 + Vite | Single-page application |
| Backend | Go + Fiber v2 | REST API, WebSocket, ESL services |
| ORM | GORM | Database abstraction and migrations |
| Primary DB | PostgreSQL 15 | All application data |
| Analytics DB | ClickHouse 23.8 | CDR/call analytics (optional) |
| Cache | Redis 7 | Session cache, pub/sub |
| Logging | Loki + Grafana | Centralized log aggregation |
| Telephony | FreeSWITCH | SIP, media, call processing |
| Auth | JWT (HS256) | Bearer token authentication |

---

## Deployment Topology

All services run in **host network mode** via Docker Compose for loopback connectivity with FreeSWITCH (which runs directly on the host). The `docker-compose.yml` defines 8 services:

| Service | Container | Port | Notes |
|---|---|---|---|
| `caddy` | callsign-caddy | 80, 443 | Routes `/api/*` → API, `/*` → UI |
| `ui` | callsign-ui | 5173 | Static Vue build served by Vite |
| `api` | callsign-api | 8080 | Go binary, bind-mounts FS dirs |
| `postgres` | callsign-postgres | 5432 | Persistent volume |
| `redis` | callsign-redis | 6379 | AOF persistence |
| `clickhouse` | callsign-clickhouse | 9000/8123 | CDR analytics (optional) |
| `loki` | callsign-loki | 3100 | Log aggregation |
| `grafana` | callsign-grafana | 3000 | Monitoring dashboards |

The API container bind-mounts FreeSWITCH directories (`/etc/freeswitch`, `/usr/share/freeswitch`, `/var/lib/freeswitch`, `/var/log/freeswitch`) so it can read/write configuration files, sounds, and recordings.

---

## Backend Architecture

The backend is a Go application located in `api/`. Source code layout:

```
api/
├── main.go              # Application entrypoint & startup orchestration
├── config/
│   └── config.go        # Environment-based configuration (Config struct)
├── router/
│   └── router.go        # Route definitions (~825 lines, 200+ endpoints)
├── handlers/
│   ├── handlers.go      # Main Handler struct, shared dependencies
│   ├── system_handlers.go     # System admin endpoints
│   ├── tenant_handlers.go     # Tenant management
│   ├── routing_handlers.go    # Inbound/outbound route management
│   ├── device_handlers.go     # Device provisioning & management
│   ├── freeswitch/            # XML CURL & FreeSWITCH handlers
│   │   ├── xmlcurl.go         # Main XML CURL dispatcher
│   │   ├── configuration.go   # sofia.conf, event_socket, etc.
│   │   ├── dialplan.go        # Dynamic dialplan generation
│   │   ├── directory.go       # SIP registration directory
│   │   ├── cdr.go             # CDR ingestion from mod_xml_cdr
│   │   └── ...
│   └── ...               # 29 handler files total
├── middleware/
│   ├── auth.go           # JWT auth, role checks, permissions
│   ├── tenant.go         # X-Tenant-ID scoping for system admins
│   ├── cors.go           # CORS configuration
│   ├── audit.go          # Audit log middleware
│   ├── logging.go        # Request logging (recovery, etc.)
│   └── permissions.go    # Permission-based access control
├── models/               # 42 GORM model files (PostgreSQL)
│   ├── base.go           # DB init, AutoMigrate, seeds
│   ├── extension.go      # SIP extensions
│   ├── tenant.go         # Multi-tenant isolation
│   ├── device.go         # Phones, softphones, registrations
│   ├── feature_code.go   # Star codes (*67, *72, etc.)
│   ├── ivr.go            # IVR menus & options
│   ├── queue.go          # Call center queues
│   └── ...
├── services/
│   ├── esl/              # Event Socket Layer integration
│   ├── cdr/              # ClickHouse CDR sync
│   ├── email/            # SMTP notifications
│   ├── encryption/       # Data-at-rest encryption
│   ├── fax/              # Fax manager & gofaxlib
│   ├── logging/          # Loki log shipping
│   ├── messaging/        # SMS/MMS via Telnyx
│   ├── tts/              # Text-to-speech caching
│   ├── websocket/        # WebSocket hub for real-time events
│   └── xmlcache/         # XML response caching
└── utils/
    ├── pagination.go     # Pagination helpers
    └── response.go       # Standardized JSON responses
```

### Startup Sequence (`main.go`)

The application initializes in this order:

1. **Environment** — Load `.env` via godotenv
2. **Configuration** — Parse all env vars into `config.Config`
3. **Logging** — Initialize structured logging with optional Loki shipping
4. **Database** — Connect to PostgreSQL via GORM
5. **Migrations** — Run `AutoMigrate` for all 80+ model structs
6. **Seeds** — Create default system admin if no users exist
7. **SIP Profiles** — Import XML profiles from disk on first boot
8. **ESL Manager** — Create manager, register 6 modules, connect to FreeSWITCH
9. **TTS Service** — Initialize text-to-speech cache, warm system phrases
10. **Email Service** — Configure SMTP for voicemail-to-email
11. **Fax Manager** — Start fax routing and queue processing
12. **ClickHouse** — Connect for CDR analytics, start periodic sync (5 min)
13. **Router** — Initialize Fiber app, register all routes, wire dependencies
14. **WebSocket** — Wire WebSocket hub to ESL manager for event broadcasting
15. **Listen** — Start HTTP server, set up graceful shutdown handlers

### Authentication & Authorization

The system uses JWT tokens (HS256) with role-based access control. Three roles exist:

| Role | Scope | Token Source |
|---|---|---|
| `system_admin` | Global — full access to all tenants and system config | Admin login |
| `tenant_admin` | Single tenant — manage extensions, routing, devices, etc. | Admin login |
| `user` | Single extension — softphone, voicemail, contacts | Extension login |

**JWT Claims** include: `user_id`, `username`, `email`, `role`, `tenant_id` (optional), `extension_id` (optional).

**Middleware chain for protected routes:**
1. `RequireAuth()` — Validates JWT Bearer token
2. `AuditMiddleware()` — Logs write operations to audit trail
3. `RequireTenant()` — Resolves tenant context (from JWT or `X-Tenant-ID` header for system admins)
4. Role-specific: `RequireSystemAdmin()`, `RequireTenantAdmin()`, `RequirePermission(...)`

### Route Groups

Routes are organized into 5 authorization tiers in `router.go`:

| Group | Prefix | Auth | Description |
|---|---|---|---|
| Public | `/api/auth/*`, `/api/health` | None | Login, registration, health check |
| FreeSWITCH | `/api/freeswitch/*` | API key or localhost | XML CURL, CDR ingestion |
| Internal | `/api/internal/*` | `X-Internal-Key` header | fail2ban reporting |
| Tenant-scoped | `/api/extensions/*`, `/api/routing/*`, etc. | JWT + tenant | All tenant feature management |
| System admin | `/api/system/*` | JWT + system_admin role | Tenants, gateways, SIP profiles, etc. |

### Data Models

All models use GORM with PostgreSQL. Key model groups:

- **Core**: `User`, `Tenant`, `TenantProfile`
- **Directory**: `Extension`, `ExtensionSetting`, `ExtensionProfile`
- **SIP/Sofia**: `SIPProfile`, `SIPProfileSetting`, `SIPProfileDomain`, `Gateway`, `ACL`, `ACLNode`
- **Dialplan**: `Dialplan`, `DialplanDetail`, `Destination`
- **Call Features**: `VoicemailBox`, `Queue`, `QueueAgent`, `Conference`, `RingGroup`, `FeatureCode`, `CallFlow`, `TimeCondition`, `HolidayList`, `IVRMenu`, `SpeedDialGroup`, `CallHandlingRule`, `CallBlock`, `BroadcastCampaign`
- **Device Management**: `Device`, `DeviceLine`, `DeviceTemplate`, `DeviceManufacturer`, `DeviceProfile`, `Firmware`, `ClientRegistration`
- **Messaging**: `Conversation`, `Message`, `MessageMedia`, `MessagingProvider`, `MessagingNumber`
- **Fax**: `FaxBox`, `FaxEndpoint`, `FaxJob`, `FaxPageResult`
- **CDR/Audit**: `CallRecord`, `AuditLog`, `BannedIP`, `Recording`, `CallRecording`, `Transcription`
- **Provisioning**: `ProvisioningTemplate`, `ProvisioningVariable`
- **System Numbers**: `SystemNumber`, `NumberGroup`

All tenant-scoped models include a `TenantID` field for multi-tenant isolation.

---

## Frontend Architecture

The frontend is a Vue 3 Single-Page Application built with Vite, located in `ui/`. It does not use a component library — all components are custom-built.

```
ui/src/
├── main.js               # Vue app entry point
├── App.vue               # Root component, <router-view>
├── router.js             # Route definitions with auth guard
├── style.css             # Global styles
├── styles/               # Additional CSS partials
├── services/
│   ├── api.js            # Axios client with 30+ API modules (~870 lines)
│   ├── auth.js           # Authentication state management
│   ├── notifications.js  # WebSocket-based notification service
│   └── sipService.js     # WebRTC SIP client (JsSIP integration)
├── components/
│   ├── layout/
│   │   └── LayoutShell.vue    # Shared layout for admin & system portals
│   ├── common/                # Reusable UI components
│   ├── features/              # Feature-specific components
│   ├── flow/                  # Call flow builder components
│   └── ivr/                   # IVR visual editor components
├── layouts/
│   └── UserLayout.vue         # User portal layout (softphone-centric)
└── views/
    ├── auth/         # Login.vue, AdminLogin.vue
    ├── admin/        # 64 views — tenant admin portal
    ├── system/       # 26 views — system admin portal
    └── user/         # 8 views — end-user portal
```

### Three Portals

| Portal | Path Prefix | Layout | Target User |
|---|---|---|---|
| **User Portal** | `/` (root) | `UserLayout.vue` | End users / extensions |
| **Tenant Admin** | `/admin` | `LayoutShell.vue` | Tenant administrators |
| **System Admin** | `/system` | `LayoutShell.vue` | Platform operators |

**User Portal views**: Softphone/Dialer, Messages, Voicemail, Conferences, Fax, Contacts, Recordings, History, Settings.

**Tenant Admin views**: Overview (dashboard), Extensions, IVR, Queues, Ring Groups, Routing, Feature Codes, Devices, Conferences, Voicemail, Recordings, Music-on-Hold, Call Flows/Toggles, Schedules/Time Conditions, CDR, Reports, Messaging, Fax Server, Hospitality, Audit Log, Tenant Settings, and more.

**System Admin views**: Dashboard, Tenants, SIP Profiles, Trunks/Gateways, Routing, ACLs, Sounds/Media, Firmware, Provisioning Templates, Messaging Providers, Config Inspector, System Logs, Security, Settings.

### API Client (`services/api.js`)

The API client uses Axios with interceptors:

- **Request interceptor**: Attaches JWT `Authorization: Bearer <token>` header and `X-Tenant-ID` header (for system admins scoping to a specific tenant).
- **Response interceptor**: Auto-unwraps the `{ data: [...] }` envelope returned by Go handlers; guards against `null` (Go nil slices serialize as JSON `null`); handles 401 with token refresh.

The file exports 30+ named API modules (e.g., `extensionsAPI`, `routingAPI`, `systemAPI`, `faxAPI`, etc.), each providing CRUD methods that map 1:1 to backend routes.

### Real-Time Communication

- **WebSocket Hub** (`services/websocket/hub.go`): Server-side pub/sub hub that broadcasts events to connected clients. Events include call state changes, voicemail MWI, conference join/leave, and queue events.
- **Notification Service** (`services/notifications.js`): Client-side WebSocket consumer that connects to `/api/ws/notifications` and dispatches events to Vue components.
- **SIP Service** (`services/sipService.js`): WebRTC softphone client using JsSIP for the user portal dialer.

---

## FreeSWITCH Integration

CallSign interfaces with FreeSWITCH through two complementary channels:

### 1. XML CURL (FreeSWITCH → CallSign)

FreeSWITCH's `mod_xml_curl` module makes HTTP POST requests to the API whenever it needs configuration. The API acts as a dynamic configuration source, replacing static XML files.

```
FreeSWITCH (mod_xml_curl)
    │
    POST /api/freeswitch/{section}
    │
    ├── section=directory   → handleDirectory()   — SIP registration & auth
    ├── section=configuration → handleConfiguration() — sofia.conf, event_socket, etc.
    ├── section=dialplan    → handleDialplan()     — call routing rules
    └── section=phrases     → handlePhrases()      — language phrases
```

**Entry Point**: `handlers/freeswitch/xmlcurl.go` → `HandleXMLCurl()`. Parses the multipart form data from FreeSWITCH and dispatches to section-specific handlers.

**Authentication**: Requests from `127.0.0.1` are allowed without auth (typical for co-located FreeSWITCH). Remote requests require HTTP Basic Auth with the configured `FREESWITCH_API_KEY`.

**Section Handlers**:

| Section | File | Generates |
|---|---|---|
| `directory` | `directory.go` | User/extension XML for SIP auth, registration, voicemail |
| `dialplan` | `dialplan.go` | Context-specific dialplan XML (inbound routes, outbound routes, feature codes, internal dialing) |
| `configuration` | `configuration.go` | Sofia profiles, event_socket, callcenter queues, ACLs |
| CDR | `cdr.go` | Receives POST from `mod_xml_cdr` and stores call records |

**Caching**: Responses are cached with TTLs (configuration: 1h, directory: 5m, dialplan: 30m) via `services/xmlcache/`. Cache is flushed automatically on relevant data changes.

**Additional Files**:
- `sip_profile_importer.go` — Imports existing SIP profile XML files into the database on first boot
- `sip_profile_writer.go` — Writes DB-managed SIP profiles back to XML files for FreeSWITCH
- `debug.go` — Debugging endpoints for the config inspector

---

### 2. Event Socket Layer — ESL (CallSign → FreeSWITCH)

The ESL service provides bidirectional communication with FreeSWITCH. It uses the `go-eventsocket` library and supports both **inbound** (API sends commands to FS) and **outbound** (FS connects to API for call handling) modes.

#### ESL Architecture

```
                    ┌──────────────────────────┐
                    │      ESL Manager         │
                    │  (services/esl/manager.go)│
                    ├──────────────────────────┤
                    │  ┌──────────────────┐    │
  API commands ────▶│  │  Inbound Client   │    │◀─── FreeSWITCH ESL (:8021)
  (API/BgAPI)       │  │  (client.go)      │    │
                    │  └──────────────────┘    │
                    │  ┌──────────────────┐    │
                    │  │  Event Processor  │    │
                    │  │  (events.go)      │    │── Session tracking
                    │  └──────────────────┘    │
                    │  ┌──────────────────┐    │
                    │  │ Module Registry   │    │
                    │  │  (service.go)     │    │
                    │  └───┬──────────────┘    │
                    └──────┼───────────────────┘
                           │
          ┌────────────────┼────────────────────┐
          │                │                    │
  ┌───────▼───────┐ ┌─────▼──────┐ ┌──────────▼──────┐
  │  Call Control  │ │  Voicemail │ │  Queue Module    │
  │  127.0.0.1:9001│ │ 127.0.0.2: │ │  127.0.0.5:9001  │
  └───────────────┘ │ 9001       │ └──────────────────┘
                    └────────────┘
  ┌───────────────┐ ┌────────────┐ ┌──────────────────┐
  │  IVR Module    │ │ Conference │ │  Feature Codes   │
  │                │ │ 127.0.0.4: │ │  Module           │
  └───────────────┘ │ 9001       │ └──────────────────┘
                    └────────────┘
```

#### Components

**Client** (`client.go`): Maintains a persistent inbound ESL connection to FreeSWITCH on port 8021.
- Sends synchronous API commands (`api <command>`) and background commands (`bgapi <command>`)
- Subscribes to events: `CHANNEL_CREATE`, `CHANNEL_ANSWER`, `CHANNEL_BRIDGE`, `CHANNEL_UNBRIDGE`, `CHANNEL_HANGUP_COMPLETE`, `CHANNEL_STATE`, `DTMF`, `RECORD_START`, `RECORD_STOP`, `PLAYBACK_START`, `PLAYBACK_STOP`, `CUSTOM`
- Has automatic reconnection with exponential backoff (up to 10 attempts)
- Buffered event channel (capacity: 1000)

**Event Processor** (`events.go`): Consumes events from the client's channel and dispatches them to registered handlers. Default handlers track call sessions:
- `CHANNEL_CREATE` → Creates new `CallSession` with A-leg
- `CHANNEL_ANSWER` → Updates answer timestamp
- `CHANNEL_BRIDGE` → Registers B-leg, sets state to bridged
- `CHANNEL_HANGUP_COMPLETE` → Records hangup cause, removes session

**Session Manager** (`session.go`): Tracks active calls as `CallSession` objects with A-leg and B-leg `ChannelState`. Sessions are keyed by A-leg UUID with a reverse index for B-leg lookup. States: `initiating → ringing → early → answered → bridged → held → transferring → hangup`.

**Module Registry** (`service.go`): All call handling modules implement the `Service` interface:

```go
type Service interface {
    Name() string
    Address() string
    Init(manager *Manager) error
    Handle(conn *eventsocket.Connection)
    Shutdown()
}
```

Each module runs an outbound ESL server on a unique loopback address. FreeSWITCH connects to these addresses when the dialplan routes calls via `<action application="socket" data="127.0.0.x:9001 async full"/>`.

**Registered Modules**:

| Module | Address | Purpose |
|---|---|---|
| `callcontrol` | `127.0.0.1:9001` | General call control, B2BUA bridging |
| `voicemail` | `127.0.0.2:9001` | Voicemail recording, playback, MWI |
| `queue` | `127.0.0.5:9001` | ACD queue handling, agent dispatch |
| `ivr` | (dynamic) | IVR menu traversal, DTMF collection |
| `conference` | `127.0.0.4:9001` | Conference management with live control |
| `featurecodes` | (dynamic) | Star-code handling (*67, *72, *98, etc.) |
| `blf` | (dynamic) | Busy Lamp Field subscriptions |

**Manager Convenience Methods** (`manager.go`):
- `API(command)` — Send synchronous ESL command
- `BgAPI(command)` — Send async ESL command (for long-running operations like `sofia profile restart`)
- `ReloadXML()` — Trigger FreeSWITCH to re-fetch XML CURL data
- `SofiaRescan(profile)` — Rescan profile for new gateways
- `SofiaRestart(profile)` — Restart SIP profile (async to avoid blocking)
- `CallcenterReload()` — Reload mod_callcenter config
- `ReloadACL()` — Reload access control lists
- `NotifyCallEvent()` — Broadcast call events via WebSocket hub

---

### Integration Flow — Inbound Call Example

```
1. SIP INVITE arrives at FreeSWITCH
2. FreeSWITCH → POST /api/freeswitch/directory (section=directory)
   → API returns XML with user credentials for authentication
3. FreeSWITCH → POST /api/freeswitch/dialplan (section=dialplan)
   → API generates dialplan XML based on:
     - Inbound routes (DID matching)
     - Feature codes (star codes)
     - Extension dialing (internal)
     - Outbound routes (PSTN via gateways)
4. Dialplan action routes call to ESL module:
   <action application="socket" data="127.0.0.1:9001 async full"/>
5. Call Control module handles the call:
   - Lookup extension → check call handling rules
   - Ring endpoints → bridge to B-leg
   - On no answer → forward to voicemail module
6. ESL events flow back through the Inbound Client:
   CHANNEL_CREATE → CHANNEL_ANSWER → CHANNEL_BRIDGE → CHANNEL_HANGUP_COMPLETE
7. Event Processor updates SessionManager
8. Manager broadcasts events to WebSocket Hub → frontend clients
```

---

## Service Layer Details

### CDR Service (`services/cdr/`)
- **ClickHouseClient**: Connects to ClickHouse for analytical CDR storage
- **SyncJob**: Periodically syncs CDRs from PostgreSQL → ClickHouse (every 5 minutes)
- Automatic cleanup of synced records older than 90 days from PostgreSQL

### Email Service (`services/email/`)
- SMTP-based email delivery for voicemail-to-email notifications
- Configurable per tenant via SMTP settings

### Encryption Service (`services/encryption/`)
- AES-based data-at-rest encryption for sensitive fields (passwords, API keys)
- Key derivation from `ENCRYPTION_KEY` and `ENCRYPTION_SALT` environment variables

### Fax Service (`services/fax/`)
- Fax queue manager with retry strategy
- Uses `gofaxlib` for T.38 fax processing
- Supports send/receive with per-tenant fax boxes and endpoints

### Messaging Service (`services/messaging/`)
- SMS/MMS gateway integration (primary: Telnyx)
- Provider abstraction (`provider.go`) for multi-provider support
- Message queue with media transcoding (FFmpeg for MMS size optimization)
- WebSocket integration for real-time message delivery

### TTS Service (`services/tts/`)
- Text-to-speech caching for IVR prompts and system phrases
- Pre-warms cache with system phrases on startup
- Cache stored at `TTS_CACHE_PATH`

### WebSocket Hub (`services/websocket/`)
- Pub/sub hub for real-time event broadcasting
- Tenant-scoped event channels
- Events: call state, voicemail MWI, conference updates, queue events

### XML Cache (`services/xmlcache/`)
- In-memory cache for FreeSWITCH XML CURL responses
- Pattern-based cache invalidation
- Configurable TTLs per section type

---

## Configuration Reference

All configuration is via environment variables. See `.env.example` for the complete list. Key groups:

| Category | Variables | Notes |
|---|---|---|
| Server | `API_HOST`, `API_PORT` | Defaults: `0.0.0.0:8080` |
| Database | `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Required |
| JWT | `JWT_SECRET`, `JWT_EXPIRATION_HOURS` | Secret must be set in production |
| FreeSWITCH | `FREESWITCH_HOST`, `FREESWITCH_ESL_PORT`, `FREESWITCH_ESL_PASSWORD`, `FREESWITCH_API_KEY` | ESL defaults: `127.0.0.1:8021` |
| ESL Addresses | `ESL_CALLCONTROL_ADDR`, `ESL_VOICEMAIL_ADDR`, `ESL_CONFERENCE_ADDR`, `ESL_QUEUE_ADDR` | Loopback addresses for outbound modules |
| ClickHouse | `CLICKHOUSE_ENABLED`, `CLICKHOUSE_HOST`, `CLICKHOUSE_PORT` | Optional analytics |
| Logging | `LOG_LEVEL`, `LOG_FORMAT`, `LOKI_ENABLED`, `LOKI_URL` | Loki optional |
| Storage | `MEDIA_PATH`, `FIRMWARE_PATH`, `PROVISIONING_PATH`, `SIP_PROFILES_PATH` | FreeSWITCH shared paths |
| Encryption | `ENCRYPTION_KEY`, `ENCRYPTION_SALT` | Required for data-at-rest encryption |
| Messaging | `TELNYX_API_KEY`, `TELNYX_MESSAGING_PROFILE`, `TELNYX_WEBHOOK_SECRET` | SMS/MMS gateway |
