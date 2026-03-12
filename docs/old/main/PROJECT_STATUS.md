# CallSign - Project Status

> Last updated: 2026-03-11

## Quick Summary

| Area | Status |
|------|--------|
| **UI** | 100 Vue views (64 admin, 26 system, 8 user, 2 auth) |
| **API Handlers** | 29 handler files, ~400+ endpoints |
| **API Models** | 40 model files |
| **Services** | 9 packages (ESL, fax, messaging, TTS, encryption, CDR, logging, websocket, xmlcache) |
| **FreeSWITCH** | mod_xml_curl + mod_xml_cdr + ESL + TTS + Fax |
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
| Recordings | ✅ | - | ✅ | ✅ | ✅ |
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
| Inbound Routes | ✅ | ✅ | ✅ | ✅ | ✅ |
| Outbound Routes | ✅ | ✅ | ✅ | ✅ | ✅ |
| Call Blocks | ✅ | ✅ | - | ✅ | ✅ |

### Backend Handlers (routing_handlers.go)
| Feature | Status |
|---------|:------:|
| Inbound Route CRUD + Reorder | ✅ |
| Outbound Route CRUD + Reorder + US/CAN Defaults | ✅ |
| Call Blocks CRUD | ✅ |
| Route Debugger | ✅ |
| Feature Codes CRUD + System Codes | ✅ |
| Time Conditions CRUD | ✅ |
| Call Flows CRUD + Toggle | ✅ |
| Dial Code Collision Check | ✅ |

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
| Messaging Numbers | ✅ | ✅ | - | ✅ | ✅ |

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

### Backend Handlers (extension_portal_handlers.go)
| Endpoint | Status |
|----------|:------:|
| GetExtensionDevices | ✅ |
| GetExtensionCallHistory | ✅ |
| GetExtensionVoicemail | ✅ |
| GetExtensionSettings | ✅ |
| UpdateExtensionSettings | ✅ |
| ChangeExtensionPassword | ✅ |
| GetExtensionContacts | ✅ |
| CreateExtensionContact | ✅ |

### Backend Handlers (Other Files)
| File | Status |
|------|--------|
| handlers.go | ✅ Auth, Health, Profile, WebSocket |
| cdr_handlers.go | ✅ CDR list/export, Audit logs |
| routing_handlers.go | ✅ Full route CRUD, Feature Codes, Time Conditions, Call Flows, Call Blocks |
| messaging_handlers.go | ✅ SMS/MMS, Contacts, Chat |
| chat_handlers.go | ✅ Chat threads, rooms, queues |
| paging_handlers.go | ✅ Paging Groups, Provisioning Templates |
| device_handlers.go | ✅ Device CRUD, Provisioning, Call Control |
| conference.go | ✅ Conference CRUD + Live Control (mute/kick/lock/record/floor/stats) |
| media_handlers.go | ✅ System Sounds/Music |
| media_db_handlers.go | ✅ Audio Library, Media management |
| tenant_settings_handlers.go | ✅ Tenant Settings, Branding, SMTP, Messaging, Hospitality |
| fax_handlers.go | ✅ Fax boxes, jobs, endpoints, send, stats |
| hospitality_handlers.go | ✅ Rooms, check-in/out, wake-up calls |
| reports_handlers.go | ✅ Call volume, agent performance, queue stats, KPI, export |
| broadcast_handlers.go | ✅ Campaigns CRUD, start/stop, stats |
| live_handlers.go | ✅ Recording control, active calls, queue stats, registrations |
| location_handlers.go | ✅ E911 Locations CRUD |
| operator_panel_handlers.go | ✅ Operator panel data |
| sms_handlers.go | ✅ SMS number management |
| sms_number_handlers.go | ✅ SMS number config/assignment |
| client_registration_handlers.go | ✅ Client/device registration management |
| call_handling_handlers.go | ✅ Call handling rules CRUD + reorder |
| webhook_handlers.go | ✅ Telnyx inbound/status webhooks |
| console_ws.go | ✅ FreeSWITCH console WebSocket |
| notification_ws.go | ✅ Real-time notification WebSocket |

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
| FaxServer.vue | faxAPI | ✅ |
| Hospitality.vue | hospitalityAPI | ✅ |
| CallBroadcast.vue | broadcastAPI | ✅ |
| Reports.vue | reportsAPI | ✅ |
| CallRecordings.vue | recordingsAPI | ✅ |
| Messaging.vue | messagingAPI | ✅ |
| system/Tenants.vue | tenantsAPI | ✅ |
| system/TenantProfiles.vue | profilesAPI | ✅ |
| system/SystemGateways.vue | gatewaysAPI | ✅ |
| system/SipProfiles.vue | sipProfilesAPI | ✅ |
| system/ConfigInspector.vue | configAPI | ✅ |
| system/SystemSounds.vue | mediaAPI | ✅ |
| system/MessagingProviders.vue | messagingProvidersAPI | ✅ |

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

### High Priority
- [ ] WebSocket expansion (operator panel, queue, conference, device feeds)
- [ ] Voicemail delivery tracking & system settings
- [ ] Conference profiles CRUD

### Medium Priority
- [ ] Transcription service worker (processing queue)
- [ ] Device reboot/logs/status endpoints
- [ ] Dark mode toggle

### Lower Priority
- [ ] Billing integration
- [ ] Multi-language phrases
- [ ] Mobile-optimized views

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
| `docs/DEVELOPER.md` | Developer guide |
| `docs/ARCHITECTURE.md` | System architecture |
