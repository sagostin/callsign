# CallSign PBX

A modern multi-tenant cloud PBX platform built with Go (API) and Vue.js (UI).

## Project Structure

```
callsign/
├── api/                    # Go Iris API backend
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # Auth, CORS, logging
│   ├── models/             # GORM models
│   ├── services/           # Business logic (ESL, logging)
│   └── router/             # Route definitions
├── ui/                     # Vue.js frontend
│   └── src/views/          # 96 Vue components
├── install/freeswitch/     # FreeSWITCH configs & scripts
├── docs/                   # Documentation
│   └── reference/          # FusionPBX reference notes
└── docker/                 # Docker configs
```

## Documentation

| Document | Purpose |
|----------|---------|
| [PROJECT_STATUS.md](PROJECT_STATUS.md) | What's done vs pending |
| [BACKEND_TODO.md](BACKEND_TODO.md) | All 300 API endpoints |
| [docs/README.md](docs/README.md) | Documentation index |

## Quick Start

### API
```bash
cd api
cp .env.example .env
# Edit .env with DB credentials
go run main.go
```

### UI
```bash
cd ui
npm install
npm run dev
```

## Tech Stack

- **Backend**: Go + Iris + GORM + PostgreSQL
- **Frontend**: Vue 3 + Vite
- **Telephony**: FreeSWITCH (ESL + mod_xml_curl)
- **Auth**: JWT with RBAC

## License

See [LICENSE.md](LICENSE.md)
