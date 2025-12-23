# CallSign PBX

A modern multi-tenant cloud PBX platform built with Go (API) and Vue.js (UI).

## Project Structure

```
callsign/
├── api/                    # Go Iris API backend
│   ├── handlers/           # HTTP handlers (~18 files)
│   │   └── freeswitch/     # FreeSWITCH XML CURL handlers
│   ├── middleware/         # Auth, CORS, logging, audit
│   ├── models/             # GORM models (~32 files)
│   ├── services/           # Business logic (ESL, logging)
│   └── router/             # Route definitions (~180+ endpoints)
├── ui/                     # Vue.js frontend
│   └── src/views/          # Vue components
│       ├── admin/          # 64 tenant admin views
│       ├── system/         # 26 system admin views
│       └── user/           # 8 user portal views
├── install/                # FreeSWITCH configs & scripts
├── docs/                   # Documentation
│   └── reference/          # FusionPBX reference notes
├── docker/                 # Docker configs (Caddy)
└── configure.sh            # Interactive setup script
```

## Documentation

| Document | Purpose |
|----------|---------|
| [PROJECT_STATUS.md](PROJECT_STATUS.md) | What's done vs pending |
| [docs/BACKEND_TODO.md](docs/BACKEND_TODO.md) | API endpoint checklist |
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/API.md](docs/API.md) | API reference |

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

- **Backend**: Go 1.21+ / Iris Framework / GORM / PostgreSQL
- **Frontend**: Vue 3 / Vite 5 / Vue Router 4
- **Telephony**: FreeSWITCH (ESL + mod_xml_curl + mod_xml_cdr)
- **Auth**: JWT with RBAC (system_admin, tenant_admin, user)
- **Provisioning**: Multi-vendor (Yealink, Polycom, Grandstream, etc.)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CallSign API                              │
│                   (Go Iris + PostgreSQL)                         │
├──────────────────┬────────────────────┬─────────────────────────┤
│  mod_xml_curl    │   Go ESL Client    │   Go ESL Servers        │
│  (config/dialplan│   (event monitor)  │   (call handling)       │
│   /directory)    │                    │                         │
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
