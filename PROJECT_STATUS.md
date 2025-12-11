# CallSign - Project Status

> Last updated: 2025-12-11

## Quick Summary

| Area | Status |
|------|--------|
| **UI** | 96 Vue views (mostly complete) |
| **API Handlers** | ~90 of ~300 implemented |
| **FreeSWITCH** | mod_xml_curl + ESL foundation done |
| **Docker** | ✅ Ready for deployment |
| **Tests** | ✅ Passing |

---

## ✅ Completed

### Backend Handlers (tenant_handlers.go)
| Resource | List | Create | Get | Update | Delete |
|----------|:----:|:------:|:---:|:------:|:------:|
| Extensions | ✅ | ✅ | ✅ | ✅ | ✅ |
| Devices | ✅ | ✅ | ✅ | ✅ | ✅ |
| Voicemail Boxes | ✅ | ✅ | ✅ | ✅ | ✅ |
| Recordings | ✅ | - | ✅ | - | ✅ |
| IVR Menus | ✅ | ✅ | ✅ | ✅ | ✅ |
| Queues | ✅ | ✅ | ✅ | ✅ | ✅ |
| Ring Groups | ✅ | ✅ | ✅ | ✅ | ✅ |
| Conferences | ✅ | ✅ | ✅ | ✅ | ✅ |
| Numbers/DIDs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Dial Plans | ✅ | ✅ | ✅ | ✅ | ✅ |
| Audio Library | ✅ | ✅ | ✅ | - | ✅ |
| MOH Streams | ✅ | ✅ | ✅ | ✅ | ✅ |
| Inbound Routes | ✅ | ✅ | - | - | - |
| Outbound Routes | ✅ | ✅ | - | - | - |

### Backend Handlers (system_handlers.go)
| Resource | List | Create | Get | Update | Delete |
|----------|:----:|:------:|:---:|:------:|:------:|
| Tenants | ✅ | ✅ | ✅ | ✅ | ✅ |
| Tenant Profiles | ✅ | ✅ | ✅ | ✅ | ✅ |
| Users (System) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Gateways | ✅ | ✅ | ✅ | ✅ | ✅ |
| SIP Profiles | ✅ | ✅ | ✅ | ✅ | ✅ |
| Bridges | - | - | - | - | - |

### Backend Handlers (user_handlers.go)
| Endpoint | Status |
|----------|:------:|
| GetUserDevices | ✅ |
| GetUserCallHistory | ✅ |
| GetUserVoicemail | ✅ |
| GetUserSettings | ✅ |
| UpdateUserSettings | ✅ |
| GetUserContacts | ✅ |
| CreateUserContact | ✅ |

### Backend Handlers (Other Files)
| File | Status |
|------|--------|
| handlers.go | ✅ Auth, Health, Profile |
| cdr_handlers.go | ✅ CDR list/export, Audit logs |
| routing_handlers.go | ✅ Feature Codes, Time Conditions, Call Flows |
| messaging_handlers.go | ✅ SMS/MMS, Contacts, Chat |
| paging_handlers.go | ✅ Paging Groups, Provisioning Templates |

### UI Views Wired to API
| View | API Module | Status |
|------|-----------|:------:|
| Extensions.vue | extensionsAPI | ✅ |
| Queues.vue | queuesAPI, ringGroupsAPI | ✅ |
| Devices.vue | devicesAPI | ✅ |
| IVR.vue | ivrAPI | ✅ |
| VoicemailBoxes.vue | voicemailAPI | ✅ |
| Conferences.vue | conferencesAPI | ✅ |
| admin/CDR.vue | cdrAPI | ✅ |

### Docker & Environment
| File | Status |
|------|:------:|
| docker-compose.yml | ✅ |
| api/Dockerfile | ✅ |
| ui/Dockerfile | ✅ |
| ui/nginx.conf | ✅ |
| .env.example | ✅ |
| docker/caddy/Caddyfile | ✅ |

### Setup & Deployment Scripts
| File | Purpose |
|------|---------|
| configure.sh | Interactive setup script |
| install/freeswitch/install.sh | FreeSWITCH installer |

---

## 🔲 Pending Work

### High Priority (Phase 1)
- [x] Tenants CRUD (system admin) ✅
- [x] System users management ✅
- [x] Gateways/Trunks management ✅
- [x] SIP Profiles management ✅
- [x] Deployment setup scripts ✅
- [ ] Runtime ESL integration

### Medium Priority (Phase 2)
- [ ] Fax server handlers
- [ ] Speed dial handlers
- [ ] Call block handlers
- [ ] Transcription service implementation
- [ ] TTS service implementation
- [ ] WebSocket real-time events

### Lower Priority (Phase 3)
- [ ] Reports/Analytics endpoints
- [ ] Hospitality module
- [ ] Billing integration
- [ ] Multi-language phrases

---

## UI Views Needing Wiring

| View | Priority |
|------|----------|
| user/UserDashboard.vue | High |
| user/UserFax.vue | Medium |
| user/UserMessages.vue | Medium |
| admin/Dashboard.vue | High |
| admin/Tenants.vue | High |
| admin/Users.vue | High |
| system/Gateways.vue | High |
| system/SIPProfiles.vue | High |
| system/Bridges.vue | Medium |

---

## Quick Start

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

---

## Files Reference

| File | Purpose |
|------|---------|
| `BACKEND_TODO.md` | Complete endpoint checklist |
| `api/FREESWITCH_INTEGRATION.md` | FreeSWITCH architecture |
| `README_UI.md` | UI documentation |
| `CALLFLOW.MD` | Call flow diagrams |
