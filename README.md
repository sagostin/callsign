# CallSign PBX

A modern, multi-tenant cloud PBX platform built on FreeSWITCH with Vue 3 frontend and Go backend.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/vue-3.5+-4FC08D.svg)](https://vuejs.org)

## Features

**Three Distinct Portals:**
- **User Portal** - WebRTC softphone, voicemail, call history, personal settings
- **Tenant Admin** - Extension management, routing, IVR, queues, devices, reporting
- **System Admin** - Multi-tenant management, SIP profiles, infrastructure monitoring

**Core Capabilities:**
- Complete telephony management (extensions, devices, routing)
- Advanced call features (queues, conferences, IVR, ring groups)
- Messaging (SMS/MMS via Telnyx)
- Fax server with retry logic
- Call recording with transcription
- Real-time operations via WebSocket
- Device auto-provisioning
- Hotel/PMS integration (hospitality features)

## Quick Start

```bash
# Clone repository
git clone https://github.com/yourorg/callsign.git
cd callsign

# Setup environment
cp .env.example .env
cp api/.env.example api/.env
# Edit .env files with your configuration

# Start services
docker-compose up -d postgres redis

# Run API
cd api && go run main.go

# Run UI (new terminal)
cd ui && npm run dev
```

Access:
- User Portal: http://localhost:5173
- Admin Portal: http://localhost:5173/admin
- API: http://localhost:8080

## Documentation

- **[ROADMAP](ROADMAP.md)** - Development roadmap and priorities
- **[Architecture](docs/ARCHITECTURE.md)** - System design and technical details
- **[Frontend Guide](docs/FRONTEND.md)** - UI development patterns
- **[API Reference](docs/API_REFERENCE.md)** - REST API documentation
- **[Setup Guide](docs/SETUP_AND_USAGE.md)** - Detailed installation instructions
- **[FreeSWITCH Integration](docs/FREESWITCH_INTEGRATION.md)** - Telephony integration

## Project Structure

```
callsign/
├── api/                    # Go backend
│   ├── handlers/          # HTTP handlers (100+ endpoints)
│   ├── models/            # GORM models (80+)
│   ├── services/          # Business logic
│   ├── middleware/        # Auth, logging, audit
│   └── router/            # Route definitions
├── ui/                     # Vue 3 frontend
│   ├── src/views/         # 100+ view components
│   ├── src/components/    # Reusable components
│   ├── src/services/      # API client, WebSocket
│   └── src/layouts/       # Portal layouts
├── docs/                   # Documentation
├── install/               # Installation scripts
├── docker/                # Docker configs
└── .claude/skills/        # AI assistant context
```

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Vue 3 + Vite |
| Backend | Go + Fiber |
| Database | PostgreSQL 15 |
| Cache | Redis 7 |
| Telephony | FreeSWITCH |
| Logging | Loki + Grafana |
| Analytics | ClickHouse (optional) |

## Development

### Prerequisites

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose
- FreeSWITCH (or use provided Docker config)

### Running Tests

```bash
# Backend
cd api
go test ./...

# Frontend
cd ui
npm test

# E2E
cd ui
npm run test:e2e
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [Contributing Guidelines](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/yourorg/callsign/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourorg/callsign/discussions)
- **Documentation**: [docs/](docs/)

## Acknowledgments

- [FreeSWITCH](https://freeswitch.com/) - Open source telephony platform
- [Fiber](https://gofiber.io/) - Express inspired web framework for Go
- [Vue.js](https://vuejs.org/) - Progressive JavaScript framework
- [Lucide](https://lucide.dev/) - Beautiful icons

---

**Note**: This project is under active development. See [ROADMAP.md](ROADMAP.md) for current status and planned features.
