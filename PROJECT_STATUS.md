# CallSign - Project Status

> Last updated: 2025-12-16

## Quick Summary

| Area | Status |
|------|--------|
| **UI** | 98 Vue views (complete) |
| **API Handlers** | ~180+ of ~300 endpoints |
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
| Voicemail Messages | ✅ | - | ✅ | ✅ | ✅ |
| Recordings | ✅ | - | ✅ | - | ✅ |
| IVR Menus | ✅ | ✅ | ✅ | ✅ | ✅ |
| Queues | ✅ | ✅ | ✅ | ✅ | ✅ |
| Ring Groups | ✅ | ✅ | ✅ | ✅ | ✅ |
| Speed Dials | ✅ | ✅ | ✅ | ✅ | ✅ |
| Conferences | ✅ | ✅ | ✅ | ✅ | ✅ |
| Numbers/DIDs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Dial Plans | ✅ | ✅ | ✅ | ✅ | ✅ |
| Audio Library | ✅ | ✅ | ✅ | ✅ | ✅ |
| MOH Streams | ✅ | ✅ | ✅ | ✅ | ✅ |
| Feature Codes | ✅ | ✅ | ✅ | ✅ | ✅ |
| Time Conditions | ✅ | ✅ | ✅ | ✅ | ✅ |
| Holiday Lists | ✅ | ✅ | ✅ | ✅ | ✅ |
| Call Flows | ✅ | ✅ | ✅ | ✅ | ✅ |
| Inbound Routes | ✅ | ✅ | - | - | - |
| Outbound Routes | ✅ | ✅ | - | - | - |
| Call Blocks | ✅ | ✅ | - | ✅ | ✅ |

### Backend Handlers (system_handlers.go)
| Resource | List | Create | Get | Update | Delete |
|----------|:----:|:------:|:---:|:------:|:------:|
| Tenants | ✅ | ✅ | ✅ | ✅ | ✅ |
| Tenant Profiles | ✅ | ✅ | ✅ | ✅ | ✅ |
| Users (System) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Gateways | ✅ | ✅ | ✅ | ✅ | ✅ |
| SIP Profiles | ✅ | ✅ | ✅ | ✅ | ✅ |
| Bridges | ✅ | ✅ | ✅ | ✅ | ✅ |
| Global Dialplans | ✅ | ✅ | ✅ | ✅ | ✅ |
| ACLs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Device Templates | ✅ | ✅ | ✅ | ✅ | ✅ |
| Device Manufacturers | ✅ | ✅ | - | ✅ | ✅ |
| Firmware | ✅ | ✅ | ✅ | ✅ | ✅ |
| Messaging Providers | ✅ | ✅ | ✅ | ✅ | ✅ |

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
| routing_handlers.go | ✅ Feature Codes, Time Conditions, Call Flows, Call Blocks |
| messaging_handlers.go | ✅ SMS/MMS, Contacts, Chat |
| paging_handlers.go | ✅ Paging Groups, Provisioning Templates |
| device_handlers.go | ✅ Device CRUD, Provisioning, Call Control |
| conference.go | ✅ Conference CRUD + Live Control |
| media_handlers.go | ✅ System Sounds/Music |
| tenant_settings_handlers.go | ✅ Tenant Settings, Branding, SMTP |

### UI Views Wired to API
| View | API Module | Status |
|------|-----------|:------:|
| Extensions.vue | extensionsAPI | ✅ |
| ExtensionDetail.vue | extensionsAPI | ✅ |
| Queues.vue | queuesAPI, ringGroupsAPI | ✅ |
| Devices.vue | devicesAPI | ✅ |
| IVR.vue | ivrAPI | ✅ |
| VoicemailBoxes.vue | voicemailAPI | ✅ |
| Conferences.vue | conferencesAPI | ✅ |
| admin/CDR.vue | cdrAPI | ✅ |
| Routing.vue | routingAPI | ✅ |
| TimeConditions.vue | timeConditionsAPI | ✅ |
| CallFlows.vue | callFlowsAPI | ✅ |
| TenantSettings.vue | tenantAPI | ✅ |
| system/Tenants.vue | tenantsAPI | ✅ |
| system/TenantProfiles.vue | profilesAPI | ✅ |
| system/SystemGateways.vue | gatewaysAPI | ✅ |
| system/SipProfiles.vue | sipProfilesAPI | ✅ |
| system/ConfigInspector.vue | configAPI | ✅ |
| system/SystemSounds.vue | mediaAPI | ✅ |

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
- [x] Config Inspector ✅
- [ ] Runtime ESL integration (in progress)

### Medium Priority (Phase 2)
- [ ] Fax server handlers
- [ ] WebSocket real-time events (notifications WebSocket exists, expand)
- [ ] Transcription service implementation
- [ ] TTS service implementation

### Lower Priority (Phase 3)
- [ ] Reports/Analytics expansion (basic reports exist)
- [ ] Hospitality module
- [ ] Billing integration
- [ ] Multi-language phrases

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
| `docs/BACKEND_TODO.md` | Complete endpoint checklist |
| `api/FREESWITCH_INTEGRATION.md` | FreeSWITCH architecture |
| `docs/README_UI.md` | UI documentation |
| `docs/CALLFLOW.MD` | Call flow diagrams |
