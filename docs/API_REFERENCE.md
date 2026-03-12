# CallSign PBX — Backend API Reference

## Overview

The CallSign API is a RESTful JSON service built with Go and the **Fiber v2** framework. All endpoints live under the `/api` prefix. Responses use a consistent `{ "data": ... }` envelope for list/get operations and `{ "error": "..." }` for errors.

---

## Authentication Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/auth/login` | Public | User login (username + password + domain) |
| POST | `/api/auth/admin/login` | Public | Admin login (for tenant_admin and system_admin) |
| POST | `/api/auth/extension/login` | Public | Extension-based login (extension + password + domain) |
| POST | `/api/auth/register` | Public | Self-registration (if enabled) |
| POST | `/api/auth/password/reset` | Public | Request password reset |
| GET | `/api/auth/me` | JWT | Get current user profile |
| PUT | `/api/auth/password` | JWT | Change password |
| POST | `/api/auth/logout` | JWT | Logout |
| POST | `/api/auth/refresh` | JWT | Refresh token |

---

## Tenant-Scoped Endpoints

All tenant-scoped endpoints require a valid JWT token. Tenant context is determined from the JWT claims (`tenant_id`) or the `X-Tenant-ID` header (for system admins operating on a specific tenant).

### Extensions

| Method | Path | Description |
|---|---|---|
| GET | `/api/extensions` | List extensions (with pagination) |
| POST | `/api/extensions` | Create extension |
| GET | `/api/extensions/:ext` | Get extension details |
| PUT | `/api/extensions/:ext` | Update extension |
| DELETE | `/api/extensions/:ext` | Delete extension |
| GET | `/api/extensions/:ext/status` | Get extension registration status |
| GET | `/api/extensions/:ext/call-rules` | List call handling rules |
| POST | `/api/extensions/:ext/call-rules` | Create call handling rule |
| PUT | `/api/extensions/:ext/call-rules/:ruleId` | Update call handling rule |
| DELETE | `/api/extensions/:ext/call-rules/:ruleId` | Delete call handling rule |
| POST | `/api/extensions/:ext/call-rules/reorder` | Reorder call handling rules |

### Extension Profiles

| Method | Path | Description |
|---|---|---|
| GET | `/api/extension-profiles` | List extension profiles |
| POST | `/api/extension-profiles` | Create profile |
| GET | `/api/extension-profiles/:id` | Get profile |
| PUT | `/api/extension-profiles/:id` | Update profile |
| DELETE | `/api/extension-profiles/:id` | Delete profile |
| GET/POST/PUT/DELETE | `/api/extension-profiles/:id/call-rules[/:ruleId]` | Profile call handling rules |

### Devices

| Method | Path | Description |
|---|---|---|
| GET | `/api/devices` | List devices |
| POST | `/api/devices` | Create device |
| GET | `/api/devices/:id` | Get device |
| PUT | `/api/devices/:id` | Update device |
| DELETE | `/api/devices/:id` | Delete device |
| POST | `/api/devices/:id/assign-user` | Assign device to user |
| POST | `/api/devices/:id/assign-profile` | Assign device to profile |
| POST | `/api/devices/:id/reprovision` | Trigger reprovisioning |
| PUT | `/api/devices/:id/lines` | Update device line configuration |
| POST | `/api/devices/:mac/hangup` | Hangup active call on device |
| POST | `/api/devices/:mac/transfer` | Transfer call on device |
| POST | `/api/devices/:mac/hold` | Hold/unhold call on device |
| POST | `/api/devices/:mac/dial` | Initiate call from device |
| GET | `/api/devices/:mac/call-status` | Get device call status |

### Device Profiles & Templates

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/device-profiles[/:id]` | Tenant-level device profiles |
| GET/POST | `/api/device-templates` | Tenant-level device templates |

### Client Registrations

| Method | Path | Description |
|---|---|---|
| GET | `/api/registrations` | List client registrations |
| POST | `/api/registrations/provision` | Provision new registration |
| DELETE | `/api/registrations/:id` | Delete registration |
| GET | `/api/registrations/unassigned` | List unassigned registrations |
| POST | `/api/registrations/:id/assign` | Assign registration to extension |
| GET | `/api/registrations/extension/:id` | List registrations for extension |

### Voicemail

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/voicemail/boxes[/:ext]` | Voicemail box management |
| GET | `/api/voicemail/boxes/:ext/messages` | List messages in box |
| GET | `/api/voicemail/messages/:id` | Get message details |
| DELETE | `/api/voicemail/messages/:id` | Delete message |
| POST | `/api/voicemail/messages/:id/read` | Mark message as read |
| GET | `/api/voicemail/messages/:id/stream` | Stream message audio |

### Recordings

| Method | Path | Description |
|---|---|---|
| GET | `/api/recordings` | List recordings |
| GET | `/api/recordings/config` | Get recording configuration |
| GET | `/api/recordings/:id` | Get recording |
| DELETE | `/api/recordings/:id` | Delete recording |
| GET | `/api/recordings/:id/stream` | Stream recording |
| GET | `/api/recordings/:id/download` | Download recording |
| PUT | `/api/recordings/:id/notes` | Update recording notes |
| GET | `/api/recordings/:id/transcription` | Get transcription |

### IVR Menus

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/ivr/menus[/:id]` | IVR menu management |

### Queues

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/queues[/:id]` | Queue management |
| GET | `/api/queues/:id/agents` | List queue agents |
| POST | `/api/queues/:id/agents` | Add agent to queue |
| DELETE | `/api/queues/:id/agents/:agentId` | Remove agent |
| POST | `/api/queues/:id/agents/:agentId/pause` | Pause agent |
| POST | `/api/queues/:id/agents/:agentId/unpause` | Unpause agent |

### Ring Groups, Speed Dials, Conferences

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/ring-groups[/:id]` | Ring group management |
| CRUD | `/api/speed-dials[/:id]` | Speed dial group management |
| CRUD | `/api/conferences[/:id]` | Conference room management |
| GET | `/api/conferences/:id/stats` | Conference statistics |
| GET | `/api/conferences/:id/sessions` | Session history |
| Various | `/api/conferences/live/*` | Live conference control (mute, kick, lock, record, etc.) |

### Numbers/DIDs

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/numbers[/:id]` | DID number management |
| POST | `/api/numbers/:id/location` | Assign number to E911 location |
| DELETE | `/api/numbers/:id/location` | Unassign from location |

### Routing

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/routing/inbound[/:id]` | Inbound route management |
| POST | `/api/routing/inbound/reorder` | Reorder inbound routes |
| CRUD | `/api/routing/outbound[/:id]` | Outbound route management |
| POST | `/api/routing/outbound/reorder` | Reorder outbound routes |
| POST | `/api/routing/outbound/defaults` | Create default US/CAN routes |
| CRUD | `/api/routing/blocks[/:id]` | Call block management |
| GET | `/api/routing/debug` | Dialplan debug tool |

### Dial Plans, Feature Codes, Time Conditions, Holidays, Call Flows

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/dial-plans[/:id]` | Dial plan management |
| CRUD | `/api/feature-codes[/:id]` | Feature code management |
| GET | `/api/feature-codes/system` | List system feature codes |
| GET | `/api/feature-codes/modules` | List available modules |
| POST | `/api/feature-codes/provision` | Provision feature code modules |
| DELETE | `/api/feature-codes/deprovision` | Deprovision modules |
| CRUD | `/api/time-conditions[/:id]` | Time condition management |
| CRUD | `/api/holidays[/:id]` | Holiday list management |
| POST | `/api/holidays/:id/sync` | Sync holidays from external source |
| CRUD | `/api/call-flows[/:id]` | Call flow management |
| POST | `/api/call-flows/:id/toggle` | Toggle call flow state |

### Audio Library, Music on Hold, Tenant Media

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/audio-library[/:id]` | Audio file library |
| GET | `/api/audio-library/:id/stream` | Stream audio file |
| CRUD | `/api/music-on-hold[/:id]` | Music on hold streams |
| GET/POST/DELETE | `/api/media/sounds` | Tenant sound overrides |
| GET/POST/DELETE | `/api/media/music` | Tenant music overrides |

### CDR, Audit, Reports

| Method | Path | Description |
|---|---|---|
| GET | `/api/cdr` | List call detail records |
| GET | `/api/cdr/:id` | Get CDR detail |
| GET | `/api/cdr/export` | Export CDR as file |
| GET | `/api/audit-logs` | List audit logs |
| GET | `/api/reports/call-volume` | Call volume report |
| GET | `/api/reports/agent-performance` | Agent performance report |
| GET | `/api/reports/queue-stats` | Queue statistics report |
| GET | `/api/reports/extension-usage` | Extension usage report |
| GET | `/api/reports/kpi` | KPI dashboard report |
| GET | `/api/reports/number-usage` | Number usage report |
| GET | `/api/reports/export` | Export report data |

### Messaging, Chat, Contacts, Fax

| Method | Path | Description |
|---|---|---|
| GET | `/api/messaging/conversations[/:id]` | SMS/MMS conversations |
| POST | `/api/messaging/send` | Send SMS/MMS |
| GET/PUT/POST/DELETE | `/api/messaging/numbers/*` | SMS number management |
| CRUD | `/api/chat/threads[/:id]` | Chat threads |
| POST | `/api/chat/threads/:id/messages` | Send chat message |
| CRUD | `/api/chat/rooms[/:id]` | Chat rooms |
| CRUD | `/api/chat/queues[/:id]` | Chat queues |
| CRUD | `/api/contacts[/:id]` | Contact management |
| GET | `/api/contacts/lookup` | Lookup contact by phone |
| Various | `/api/fax/*` | Fax boxes, jobs, endpoints, send/receive |

### Paging, Broadcast, Hospitality, Provisioning, Live Ops

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/page-groups[/:id]` | Paging groups |
| CRUD | `/api/broadcast[/:id]` | Call broadcast campaigns |
| POST | `/api/broadcast/:id/start\|stop` | Start/stop broadcast |
| CRUD | `/api/hospitality/rooms[/:id]` | Hotel room management |
| POST | `/api/hospitality/rooms/:id/checkin\|checkout` | Guest check-in/check-out |
| POST | `/api/hospitality/rooms/:id/wakeup` | Schedule wake-up call |
| CRUD | `/api/provisioning-templates[/:id]` | Provisioning templates |
| Various | `/api/live/*` | Live recording, calls, queue stats |
| GET | `/api/operator-panel` | Operator panel data |

### Tenant Settings

| Method | Path | Description |
|---|---|---|
| GET/PUT | `/api/tenant/settings` | General tenant settings |
| GET/PUT | `/api/tenant/branding` | Tenant branding |
| GET/PUT | `/api/tenant/smtp` | SMTP configuration |
| POST | `/api/tenant/smtp/test` | Test SMTP |
| GET/PUT | `/api/tenant/messaging` | Messaging settings |
| GET/PUT | `/api/tenant/hospitality` | Hospitality settings |
| CRUD | `/api/tenant/locations[/:id]` | E911 locations |

---

## System Admin Endpoints

All system endpoints require `system_admin` role. Prefix: `/api/system`.

### Tenants & Profiles

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/system/tenants[/:id]` | Tenant management |
| CRUD | `/api/system/tenant-profiles[/:id]` | Tenant profiles (limits, features) |

### System Numbers & Number Groups

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/system/numbers[/:id]` | System number pool (DIDs) |
| POST | `/api/system/numbers/:id/assign` | Assign number to tenant |
| POST | `/api/system/numbers/:id/unassign` | Unassign number |
| CRUD | `/api/system/number-groups[/:id]` | Number groups for outbound routing |
| POST | `/api/system/number-groups/:id/reorder-gateways` | Reorder gateway priority |

### SIP Infrastructure

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/system/gateways[/:id]` | SIP trunk/gateway management |
| GET | `/api/system/gateways/status` | Live gateway status |
| POST | `/api/system/gateways/reorder` | Reorder gateway priority |
| CRUD | `/api/system/bridges[/:id]` | Bridge management |
| CRUD | `/api/system/sip-profiles[/:id]` | SIP profile management |
| POST | `/api/system/sip-profiles/sync` | Import profiles from disk |

### Sofia Control (Live FreeSWITCH)

| Method | Path | Description |
|---|---|---|
| GET | `/api/system/sofia/status` | Overall sofia status |
| GET | `/api/system/sofia/profiles/:name/status` | Profile status |
| GET | `/api/system/sofia/profiles/:name/registrations` | Profile registrations |
| GET | `/api/system/sofia/profiles/:name/gateways` | Profile gateways |
| POST | `/api/system/sofia/profiles/:name/restart` | Restart profile |
| POST | `/api/system/sofia/profiles/:name/start` | Start profile |
| POST | `/api/system/sofia/profiles/:name/stop` | Stop profile |
| POST | `/api/system/sofia/reload-xml` | Reload XML configuration |

### Messaging, Dialplans, ACLs, Media, Security, Settings

| Method | Path | Description |
|---|---|---|
| CRUD | `/api/system/messaging-providers[/:id]` | Messaging provider management |
| CRUD | `/api/system/messaging-numbers[/:id]` | Messaging number management |
| CRUD | `/api/system/dialplans[/:id]` | Global dial plans |
| CRUD | `/api/system/acls[/:id]` | Access control lists |
| CRUD | `/api/system/acls/:id/nodes[/:nodeId]` | ACL nodes/entries |
| GET/POST | `/api/system/media/sounds` | System sound management |
| GET/POST | `/api/system/media/music` | System music management |
| CRUD | `/api/system/device-templates[/:id]` | System device templates |
| CRUD | `/api/system/device-manufacturers[/:id]` | Device manufacturers |
| CRUD | `/api/system/firmware[/:id]` | Firmware management |
| GET/DELETE | `/api/system/security/banned-ips[/:ip]` | IP ban management |
| GET/PUT | `/api/system/settings` | System settings |
| GET | `/api/system/status` | System status |
| GET | `/api/system/stats` | System statistics |
| GET | `/api/system/logs` | System logs |
| GET | `/api/system/xml/debug` | XML debug output |
| GET | `/api/system/config/files` | Config file browser |
| GET | `/api/system/config/file` | Read config file |

---

## User & Extension Portal Endpoints

### User Portal (User-scoped)

| Method | Path | Description |
|---|---|---|
| GET | `/api/user/devices` | User's devices |
| GET | `/api/user/call-history` | Call history |
| GET | `/api/user/voicemail` | Voicemail messages |
| GET/PUT | `/api/user/settings` | User settings |
| GET/POST | `/api/user/contacts` | Contacts |

### Extension Portal (Extension-scoped)

| Method | Path | Description |
|---|---|---|
| GET | `/api/extension/portal/devices` | Extension's devices |
| GET | `/api/extension/portal/call-history` | Call history |
| GET | `/api/extension/portal/voicemail` | Voicemail |
| GET/PUT | `/api/extension/portal/settings` | Extension settings |
| PUT | `/api/extension/portal/password` | Change extension password |
| GET/POST | `/api/extension/portal/contacts` | Contacts |

---

## FreeSWITCH Internal Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/freeswitch/directory` | API key / localhost | SIP user directory |
| POST | `/api/freeswitch/dialplan` | API key / localhost | Dialplan generation |
| POST | `/api/freeswitch/configuration` | API key / localhost | Module configuration |
| POST | `/api/freeswitch/xmlapi` | API key / localhost | Combined handler (legacy) |
| POST | `/api/freeswitch/cdr` | API key / localhost | CDR ingestion |
| GET | `/api/freeswitch/cache/flush` | API key / localhost | Flush XML cache |
| GET | `/api/freeswitch/cache/stats` | API key / localhost | Cache statistics |

---

## WebSocket & Webhook Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/system/console` | JWT (via first message) | FreeSWITCH console WebSocket |
| GET | `/api/ws/notifications` | JWT (via first message) | Real-time notification WebSocket |
| GET | `/api/ws` | JWT | General WebSocket |
| POST | `/api/webhooks/telnyx/inbound` | Webhook signature | Inbound SMS/MMS webhook |
| POST | `/api/webhooks/telnyx/status` | Webhook signature | SMS delivery status webhook |

---

## Provisioning Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/provision/:tenant/:secret/:mac` | Tenant secret in URL | Device config (secure) |
| GET | `/provisioning/:mac/:filename` | MAC-based | Serve provisioning config file |

---

## Internal Service Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/internal/fail2ban/report` | `X-Internal-Key` | Report banned IP from fail2ban |
