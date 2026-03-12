# CallSign PBX

A modern multi-tenant cloud PBX platform built with Go (API) and Vue.js (UI).

## Project Structure

```
callsign/
├── api/                    # Go Fiber API backend
│   ├── handlers/           # HTTP handlers (29 files)
│   │   └── freeswitch/     # FreeSWITCH XML CURL handlers
│   ├── middleware/         # Auth, CORS, logging, audit, tenant
│   ├── models/             # GORM models (40 files)
│   ├── services/           # Business logic (9 packages)
│   │   ├── cdr/            # ClickHouse CDR sync
│   │   ├── encryption/     # AES-256-GCM encryption
│   │   ├── esl/            # FreeSWITCH ESL services
│   │   ├── fax/            # Fax processing & routing
│   │   ├── logging/        # Loki integration
│   │   ├── messaging/      # SMS/MMS messaging
│   │   ├── tts/            # Text-to-speech caching
│   │   ├── websocket/      # Real-time event hub
│   │   └── xmlcache/       # FreeSWITCH XML cache
│   └── router/             # Route definitions (~400+ endpoints)
├── ui/                     # Vue.js frontend
│   └── src/views/          # Vue components
│       ├── admin/          # 64 tenant admin views
│       ├── system/         # 26 system admin views
│       ├── user/           # 8 user portal views
│       └── auth/           # 2 auth views
├── install/                # FreeSWITCH configs & scripts
├── docs/                   # Documentation
│   └── reference/          # FusionPBX reference notes
├── docker/                 # Docker configs (Caddy)
├── landingpage/            # Marketing landing page
└── configure.sh            # Interactive setup script
```

## Documentation

| Document | Purpose |
|----------|---------|
| [ROADMAP.md](ROADMAP.md) | Development roadmap & progress |
| [PROJECT_STATUS.md](PROJECT_STATUS.md) | What's done vs pending |
| [docs/BACKEND_TODO.md](docs/BACKEND_TODO.md) | API endpoint checklist |
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/API.md](docs/API.md) | API reference |
| [docs/DEVELOPER.md](docs/DEVELOPER.md) | Developer guide |

## Quick Start

### Using Docker (Recommended)

```bash
# Clone and setup
cd callsign
cp .env.example .env
# Edit .env with your settings

# Start all services
docker compose up -d

# View logs
docker compose logs -f api

# Access
# UI: http://localhost
# API: http://localhost:8080
# Grafana: http://localhost:3000
```

### Manual Development

```bash
# API
cd api
cp .env.example .env
# Edit .env with DB credentials
go run main.go

# UI (separate terminal)
cd ui
npm install
npm run dev
```

### Interactive Setup

```bash
./configure.sh
```

## Tech Stack

- **Backend**: Go 1.21+ / Fiber / GORM / PostgreSQL
- **Frontend**: Vue 3 / Vite 5 / Vue Router 4
- **Telephony**: FreeSWITCH (ESL + mod_xml_curl + mod_xml_cdr)
- **Auth**: JWT with RBAC (system_admin, tenant_admin, user)
- **Provisioning**: Multi-vendor (Yealink, Polycom, Grandstream, etc.)
- **Messaging**: SMS/MMS via Telnyx (webhook-based)
- **CDR Analytics**: ClickHouse (optional)
- **Logging**: Loki
- **Reverse Proxy**: Caddy (TLS + WebSocket)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CallSign API                              │
│                   (Go Fiber + PostgreSQL)                         │
├──────────────────┬────────────────────┬─────────────────────────┤
│  mod_xml_curl    │   Go ESL Client    │   Go ESL Servers        │
│  (config/dialplan│   (event monitor)  │   (call handling)       │
│   /directory)    │                    │                         │
├──────────────────┼────────────────────┼─────────────────────────┤
│  Fax Manager     │  Messaging Manager │  TTS Cache Service      │
│  (queue/retry)   │  (SMS/MMS/webhook) │  (phrase rendering)     │
└────────┬─────────┴─────────┬──────────┴────────┬────────────────┘
         │                   │                    │
         └───────────────────┴────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   FreeSWITCH    │
                    └─────────────────┘
```

## License

See [LICENSE.md](LICENSE.md)
