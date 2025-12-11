# Backend API Development Todo List

This document tracks the backend API endpoints needed to support the CallSign-UI frontend application.

---

## 1. Authentication & Authorization

- [ ] `POST /api/auth/login` - User login (returns JWT token)
- [ ] `POST /api/auth/admin/login` - Admin login (system/tenant admins)
- [ ] `POST /api/auth/logout` - Logout / invalidate token
- [ ] `POST /api/auth/refresh` - Refresh JWT token
- [ ] `GET /api/auth/me` - Get current user profile & permissions
- [ ] `PUT /api/auth/password` - Change password
- [ ] `POST /api/auth/password/reset` - Request password reset

---

## 2. Tenants (Multi-Tenant Management)

- [ ] `GET /api/tenants` - List all tenants (system admin)
- [ ] `POST /api/tenants` - Create new tenant
- [ ] `GET /api/tenants/:id` - Get tenant details
- [ ] `PUT /api/tenants/:id` - Update tenant
- [ ] `DELETE /api/tenants/:id` - Delete tenant
- [ ] `GET /api/tenants/:id/users` - List tenant users
- [ ] `POST /api/tenants/:id/users` - Add user to tenant
- [ ] `GET /api/tenants/:id/devices` - List tenant devices
- [ ] `GET /api/tenants/:id/stats` - Get tenant usage statistics

### Tenant Profiles (Service Plans)

- [ ] `GET /api/tenant-profiles` - List all tenant profiles
- [ ] `POST /api/tenant-profiles` - Create tenant profile
- [ ] `GET /api/tenant-profiles/:id` - Get profile details
- [ ] `PUT /api/tenant-profiles/:id` - Update profile
- [ ] `DELETE /api/tenant-profiles/:id` - Delete profile

---

## 3. Extensions

- [ ] `GET /api/extensions` - List all extensions
- [ ] `POST /api/extensions` - Create new extension
- [ ] `GET /api/extensions/:ext` - Get extension details
- [ ] `PUT /api/extensions/:ext` - Update extension
- [ ] `DELETE /api/extensions/:ext` - Delete extension
- [ ] `GET /api/extensions/:ext/status` - Get real-time status (SIP registration, in-call)
- [ ] `PUT /api/extensions/:ext/profile` - Quick profile assignment

### Extension Profiles

- [ ] `GET /api/extension-profiles` - List extension profiles
- [ ] `POST /api/extension-profiles` - Create profile
- [ ] `GET /api/extension-profiles/:id` - Get profile details
- [ ] `PUT /api/extension-profiles/:id` - Update profile
- [ ] `DELETE /api/extension-profiles/:id` - Delete profile

---

## 4. IVR / Auto Attendant

### IVR Menus

- [ ] `GET /api/ivr/menus` - List all IVR menus
- [ ] `POST /api/ivr/menus` - Create IVR menu
- [ ] `GET /api/ivr/menus/:id` - Get menu details
- [ ] `PUT /api/ivr/menus/:id` - Update menu
- [ ] `DELETE /api/ivr/menus/:id` - Delete menu

### Time Conditions

- [ ] `GET /api/ivr/time-conditions` - List time conditions
- [ ] `POST /api/ivr/time-conditions` - Create time condition
- [ ] `GET /api/ivr/time-conditions/:id` - Get details
- [ ] `PUT /api/ivr/time-conditions/:id` - Update time condition
- [ ] `DELETE /api/ivr/time-conditions/:id` - Delete time condition
- [ ] `GET /api/ivr/time-conditions/:id/status` - Check if currently matched

### Mode Toggles

- [ ] `GET /api/ivr/toggles` - List mode toggles
- [ ] `POST /api/ivr/toggles` - Create toggle
- [ ] `GET /api/ivr/toggles/:id` - Get toggle details
- [ ] `PUT /api/ivr/toggles/:id` - Update toggle
- [ ] `DELETE /api/ivr/toggles/:id` - Delete toggle
- [ ] `POST /api/ivr/toggles/:id/activate` - Activate toggle
- [ ] `POST /api/ivr/toggles/:id/deactivate` - Deactivate toggle

### Schedules

- [ ] `GET /api/schedules` - List schedules
- [ ] `POST /api/schedules` - Create schedule
- [ ] `GET /api/schedules/:id` - Get schedule
- [ ] `PUT /api/schedules/:id` - Update schedule
- [ ] `DELETE /api/schedules/:id` - Delete schedule

---

## 5. Call Routing

### Phone Numbers (DIDs)

- [ ] `GET /api/numbers` - List phone numbers
- [ ] `POST /api/numbers` - Add/provision number
- [ ] `GET /api/numbers/:id` - Get number details
- [ ] `PUT /api/numbers/:id` - Update number routing
- [ ] `DELETE /api/numbers/:id` - Remove number
- [ ] `GET /api/numbers/:id/stats` - Number usage statistics

### Inbound Routes

- [ ] `GET /api/routing/inbound` - List inbound routes (ordered)
- [ ] `POST /api/routing/inbound` - Create inbound route
- [ ] `GET /api/routing/inbound/:id` - Get route details
- [ ] `PUT /api/routing/inbound/:id` - Update route
- [ ] `DELETE /api/routing/inbound/:id` - Delete route
- [ ] `PUT /api/routing/inbound/reorder` - Reorder routes (batch)

### Outbound Routes

- [ ] `GET /api/routing/outbound` - List outbound routes
- [ ] `POST /api/routing/outbound` - Create outbound route
- [ ] `GET /api/routing/outbound/:id` - Get route details
- [ ] `PUT /api/routing/outbound/:id` - Update route
- [ ] `DELETE /api/routing/outbound/:id` - Delete route
- [ ] `PUT /api/routing/outbound/reorder` - Reorder routes (batch)

### Dial Plans

- [ ] `GET /api/dial-plans` - List dial plans
- [ ] `POST /api/dial-plans` - Create dial plan
- [ ] `GET /api/dial-plans/:id` - Get dial plan
- [ ] `PUT /api/dial-plans/:id` - Update dial plan
- [ ] `DELETE /api/dial-plans/:id` - Delete dial plan

### Routing Settings

- [ ] `GET /api/routing/settings` - Get routing settings
- [ ] `PUT /api/routing/settings` - Update routing settings

---

## 6. Devices (Provisioning)

- [ ] `GET /api/devices` - List all provisioned devices
- [ ] `POST /api/devices` - Add device (manual)
- [ ] `GET /api/devices/:mac` - Get device details
- [ ] `PUT /api/devices/:mac` - Update device
- [ ] `DELETE /api/devices/:mac` - Delete device
- [ ] `POST /api/devices/:mac/reprovision` - Trigger reprovision
- [ ] `POST /api/devices/:mac/reboot` - Send reboot command
- [ ] `GET /api/devices/:mac/logs` - Get device logs
- [ ] `GET /api/devices/:mac/status` - Get registration status
- [ ] `POST /api/devices/scan` - Scan network for unprovisioned devices

### Device Templates

- [ ] `GET /api/device-templates` - List device templates
- [ ] `POST /api/device-templates` - Create template
- [ ] `GET /api/device-templates/:id` - Get template details
- [ ] `PUT /api/device-templates/:id` - Update template
- [ ] `DELETE /api/device-templates/:id` - Delete template
- [ ] `GET /api/device-templates/:id/preview` - Preview provisioning config

### Provisioning Templates (System-level)

- [ ] `GET /api/system/provisioning-templates` - List master templates
- [ ] `POST /api/system/provisioning-templates` - Create template
- [ ] `GET /api/system/provisioning-templates/:id` - Get template
- [ ] `PUT /api/system/provisioning-templates/:id` - Update template
- [ ] `DELETE /api/system/provisioning-templates/:id` - Delete template

---

## 7. Call Queues & Ring Groups

### Queues

- [ ] `GET /api/queues` - List call queues
- [ ] `POST /api/queues` - Create queue
- [ ] `GET /api/queues/:id` - Get queue details
- [ ] `PUT /api/queues/:id` - Update queue
- [ ] `DELETE /api/queues/:id` - Delete queue
- [ ] `GET /api/queues/:id/stats` - Real-time queue statistics
- [ ] `GET /api/queues/:id/agents` - List queue agents
- [ ] `POST /api/queues/:id/agents` - Add agent to queue
- [ ] `DELETE /api/queues/:id/agents/:agentId` - Remove agent from queue
- [ ] `PUT /api/queues/:id/agents/:agentId/pause` - Pause agent
- [ ] `PUT /api/queues/:id/agents/:agentId/unpause` - Unpause agent

### Ring Groups

- [ ] `GET /api/ring-groups` - List ring groups
- [ ] `POST /api/ring-groups` - Create ring group
- [ ] `GET /api/ring-groups/:id` - Get ring group details
- [ ] `PUT /api/ring-groups/:id` - Update ring group
- [ ] `DELETE /api/ring-groups/:id` - Delete ring group

---

## 8. Conferencing

- [ ] `GET /api/conferences` - List conference rooms
- [ ] `POST /api/conferences` - Create conference room
- [ ] `GET /api/conferences/:id` - Get conference details
- [ ] `PUT /api/conferences/:id` - Update conference
- [ ] `DELETE /api/conferences/:id` - Delete conference
- [ ] `GET /api/conferences/:id/participants` - List active participants (real-time)
- [ ] `POST /api/conferences/:id/mute-all` - Mute all participants
- [ ] `POST /api/conferences/:id/lock` - Lock conference
- [ ] `POST /api/conferences/:id/unlock` - Unlock conference
- [ ] `POST /api/conferences/:id/kick/:participantId` - Kick participant
- [ ] `POST /api/conferences/:id/mute/:participantId` - Mute participant
- [ ] `POST /api/conferences/:id/unmute/:participantId` - Unmute participant
- [ ] `POST /api/conferences/:id/record/start` - Start recording
- [ ] `POST /api/conferences/:id/record/stop` - Stop recording

### Conference Profiles

- [ ] `GET /api/conference-profiles` - List conference profiles
- [ ] `POST /api/conference-profiles` - Create profile
- [ ] `GET /api/conference-profiles/:id` - Get profile
- [ ] `PUT /api/conference-profiles/:id` - Update profile
- [ ] `DELETE /api/conference-profiles/:id` - Delete profile

---

## 9. Voicemail

### Voicemail Boxes

- [ ] `GET /api/voicemail/boxes` - List voicemail boxes
- [ ] `POST /api/voicemail/boxes` - Create voicemail box
- [ ] `GET /api/voicemail/boxes/:ext` - Get box details
- [ ] `PUT /api/voicemail/boxes/:ext` - Update box
- [ ] `DELETE /api/voicemail/boxes/:ext` - Delete box
- [ ] `GET /api/voicemail/boxes/:ext/messages` - List messages
- [ ] `GET /api/voicemail/boxes/:ext/messages/:id` - Get message
- [ ] `DELETE /api/voicemail/boxes/:ext/messages/:id` - Delete message
- [ ] `GET /api/voicemail/boxes/:ext/messages/:id/audio` - Stream message audio
- [ ] `PUT /api/voicemail/boxes/:ext/messages/:id/read` - Mark as read

### Voicemail Delivery

- [ ] `GET /api/voicemail/delivery-attempts` - List delivery attempts
- [ ] `POST /api/voicemail/delivery-attempts/:id/retry` - Retry failed delivery
- [ ] `POST /api/voicemail/delivery/retry-all-failed` - Retry all failed

### Voicemail Settings

- [ ] `GET /api/voicemail/settings` - Get voicemail settings
- [ ] `PUT /api/voicemail/settings` - Update settings
- [ ] `GET /api/voicemail/status` - System status (SMTP, transcription)

---

## 10. Fax

### Fax Servers

- [ ] `GET /api/fax/servers` - List fax servers
- [ ] `POST /api/fax/servers` - Create fax server
- [ ] `GET /api/fax/servers/:id` - Get server details
- [ ] `PUT /api/fax/servers/:id` - Update server
- [ ] `DELETE /api/fax/servers/:id` - Delete server
- [ ] `GET /api/fax/servers/:id/inbox` - List received faxes
- [ ] `GET /api/fax/servers/:id/sent` - List sent faxes

### Fax Operations

- [ ] `POST /api/fax/send` - Send fax
- [ ] `GET /api/fax/:id` - Get fax details
- [ ] `GET /api/fax/:id/download` - Download fax PDF
- [ ] `DELETE /api/fax/:id` - Delete fax

---

## 11. Messaging (SMS/MMS)

### Conversations

- [ ] `GET /api/messages/conversations` - List conversations
- [ ] `GET /api/messages/conversations/:id` - Get conversation messages
- [ ] `POST /api/messages/send` - Send message
- [ ] `PUT /api/messages/conversations/:id/read` - Mark as read
- [ ] `DELETE /api/messages/conversations/:id` - Delete conversation

### Messaging Providers (System)

- [ ] `GET /api/system/messaging-providers` - List providers
- [ ] `POST /api/system/messaging-providers` - Add provider
- [ ] `GET /api/system/messaging-providers/:id` - Get provider
- [ ] `PUT /api/system/messaging-providers/:id` - Update provider
- [ ] `DELETE /api/system/messaging-providers/:id` - Delete provider
- [ ] `POST /api/system/messaging-providers/:id/test` - Test provider

---

## 12. Call Recordings

- [ ] `GET /api/recordings` - List recordings (with filtering)
- [ ] `GET /api/recordings/:id` - Get recording metadata
- [ ] `GET /api/recordings/:id/audio` - Stream recording audio
- [ ] `GET /api/recordings/:id/download` - Download recording
- [ ] `DELETE /api/recordings/:id` - Delete recording
- [ ] `POST /api/recordings/:id/email` - Email recording
- [ ] `POST /api/recordings/export` - Batch export
- [ ] `GET /api/recordings/stats` - Recording statistics

---

## 13. Audio Library (Prompts/Greetings)

- [ ] `GET /api/audio-library` - List audio files
- [ ] `POST /api/audio-library` - Upload audio file
- [ ] `GET /api/audio-library/:id` - Get audio metadata
- [ ] `GET /api/audio-library/:id/stream` - Stream audio
- [ ] `PUT /api/audio-library/:id` - Update metadata
- [ ] `DELETE /api/audio-library/:id` - Delete audio
- [ ] `POST /api/audio-library/record` - Record audio (TTS or phone)

---

## 14. Music on Hold / Streams

- [ ] `GET /api/music-on-hold` - List MoH streams
- [ ] `POST /api/music-on-hold` - Create stream
- [ ] `GET /api/music-on-hold/:id` - Get stream details
- [ ] `PUT /api/music-on-hold/:id` - Update stream
- [ ] `DELETE /api/music-on-hold/:id` - Delete stream
- [ ] `POST /api/music-on-hold/:id/files` - Add file to stream
- [ ] `DELETE /api/music-on-hold/:id/files/:fileId` - Remove file

### System Streams

- [ ] `GET /api/system/streams` - List system streams
- [ ] `POST /api/system/streams` - Create system stream
- [ ] `PUT /api/system/streams/:id` - Update stream
- [ ] `DELETE /api/system/streams/:id` - Delete stream

---

## 15. Feature Codes

- [ ] `GET /api/feature-codes` - List feature codes
- [ ] `POST /api/feature-codes` - Create feature code
- [ ] `GET /api/feature-codes/:id` - Get feature code
- [ ] `PUT /api/feature-codes/:id` - Update feature code
- [ ] `DELETE /api/feature-codes/:id` - Delete feature code

---

## 16. Call Block

- [ ] `GET /api/call-block` - List blocked numbers/patterns
- [ ] `POST /api/call-block` - Add block rule
- [ ] `GET /api/call-block/:id` - Get block rule
- [ ] `PUT /api/call-block/:id` - Update rule
- [ ] `DELETE /api/call-block/:id` - Delete rule

---

## 17. Call Broadcast

- [ ] `GET /api/call-broadcast` - List broadcast campaigns
- [ ] `POST /api/call-broadcast` - Create campaign
- [ ] `GET /api/call-broadcast/:id` - Get campaign details
- [ ] `PUT /api/call-broadcast/:id` - Update campaign
- [ ] `DELETE /api/call-broadcast/:id` - Delete campaign
- [ ] `POST /api/call-broadcast/:id/start` - Start campaign
- [ ] `POST /api/call-broadcast/:id/stop` - Stop campaign
- [ ] `GET /api/call-broadcast/:id/stats` - Campaign statistics

---

## 18. Speed Dials

- [ ] `GET /api/speed-dials` - List speed dials
- [ ] `POST /api/speed-dials` - Create speed dial
- [ ] `PUT /api/speed-dials/:id` - Update speed dial
- [ ] `DELETE /api/speed-dials/:id` - Delete speed dial

---

## 19. Reports & Analytics

- [ ] `GET /api/reports/call-volume` - Call volume report
- [ ] `GET /api/reports/call-detail` - CDR (Call Detail Records)
- [ ] `GET /api/reports/agent-performance` - Agent performance metrics
- [ ] `GET /api/reports/queue-stats` - Queue statistics
- [ ] `GET /api/reports/extension-usage` - Extension usage
- [ ] `GET /api/reports/kpi` - Key performance indicators
- [ ] `POST /api/reports/export` - Export report (CSV/PDF)

---

## 20. Audit Log

- [ ] `GET /api/audit-log` - List audit log entries
- [ ] `GET /api/audit-log/:id` - Get log entry details
- [ ] `POST /api/audit-log/export` - Export audit log

---

## 21. User Portal (End User)

### Softphone / Dialer

- [ ] `GET /api/user/devices` - List user's devices
- [ ] `POST /api/user/click-to-call` - Initiate click-to-call
- [ ] `GET /api/user/sip-credentials` - Get WebRTC SIP credentials
- [ ] `POST /api/user/bind-device` - Bind to specific device

### Call History

- [ ] `GET /api/user/call-history` - User's call history
- [ ] `GET /api/user/call-history/:id` - Call details

### Contacts

- [ ] `GET /api/user/contacts` - List user contacts
- [ ] `POST /api/user/contacts` - Add contact
- [ ] `PUT /api/user/contacts/:id` - Update contact
- [ ] `DELETE /api/user/contacts/:id` - Delete contact

### User Settings

- [ ] `GET /api/user/settings` - Get user settings
- [ ] `PUT /api/user/settings` - Update settings
- [ ] `PUT /api/user/settings/call-forwarding` - Update call forwarding
- [ ] `PUT /api/user/settings/dnd` - Toggle Do Not Disturb

### User Voicemail

- [ ] `GET /api/user/voicemail` - List voicemail messages
- [ ] `GET /api/user/voicemail/:id` - Get message
- [ ] `DELETE /api/user/voicemail/:id` - Delete message
- [ ] `GET /api/user/voicemail/:id/audio` - Stream audio

### User Fax

- [ ] `GET /api/user/fax/inbox` - User's received faxes
- [ ] `GET /api/user/fax/sent` - User's sent faxes
- [ ] `POST /api/user/fax/send` - Send fax

### User Recordings

- [ ] `GET /api/user/recordings` - User's call recordings
- [ ] `GET /api/user/recordings/:id/audio` - Stream recording

### User Conferences

- [ ] `GET /api/user/conferences` - User's conference rooms
- [ ] `POST /api/user/conferences/:id/join` - Join conference

---

## 22. System Administration

### Gateways / Trunks

- [ ] `GET /api/system/gateways` - List gateways
- [ ] `POST /api/system/gateways` - Create gateway
- [ ] `GET /api/system/gateways/:id` - Get gateway
- [ ] `PUT /api/system/gateways/:id` - Update gateway
- [ ] `DELETE /api/system/gateways/:id` - Delete gateway
- [ ] `GET /api/system/gateways/:id/status` - Gateway status
- [ ] `POST /api/system/gateways/:id/test` - Test gateway

### Bridges

- [ ] `GET /api/system/bridges` - List bridges
- [ ] `POST /api/system/bridges` - Create bridge
- [ ] `GET /api/system/bridges/:id` - Get bridge
- [ ] `PUT /api/system/bridges/:id` - Update bridge
- [ ] `DELETE /api/system/bridges/:id` - Delete bridge

### SIP Profiles

- [ ] `GET /api/system/sip-profiles` - List SIP profiles
- [ ] `POST /api/system/sip-profiles` - Create profile
- [ ] `GET /api/system/sip-profiles/:id` - Get profile
- [ ] `PUT /api/system/sip-profiles/:id` - Update profile
- [ ] `DELETE /api/system/sip-profiles/:id` - Delete profile

### Global Dial Plans

- [ ] `GET /api/system/dial-plans` - List global dial plans
- [ ] `POST /api/system/dial-plans` - Create dial plan
- [ ] `PUT /api/system/dial-plans/:id` - Update dial plan
- [ ] `DELETE /api/system/dial-plans/:id` - Delete dial plan

### Phrases (System Audio)

- [ ] `GET /api/system/phrases` - List system phrases
- [ ] `POST /api/system/phrases` - Create phrase
- [ ] `GET /api/system/phrases/:id` - Get phrase
- [ ] `PUT /api/system/phrases/:id` - Update phrase
- [ ] `DELETE /api/system/phrases/:id` - Delete phrase

### System Settings

- [ ] `GET /api/system/settings` - Get all system settings
- [ ] `PUT /api/system/settings` - Update settings
- [ ] `GET /api/system/settings/smtp` - SMTP settings
- [ ] `PUT /api/system/settings/smtp` - Update SMTP
- [ ] `POST /api/system/settings/smtp/test` - Test SMTP
- [ ] `GET /api/system/settings/freeswitch` - FreeSWITCH settings
- [ ] `PUT /api/system/settings/freeswitch` - Update FreeSWITCH
- [ ] `GET /api/system/settings/database` - Database info
- [ ] `GET /api/system/settings/cluster` - Cluster nodes info

### System Logs

- [ ] `GET /api/system/logs` - Get system logs
- [ ] `GET /api/system/logs/sip` - SIP logs
- [ ] `GET /api/system/logs/errors` - Error logs
- [ ] `GET /api/system/logs/security` - Security logs

### Infrastructure Status

- [ ] `GET /api/system/status` - Overall system status
- [ ] `GET /api/system/status/freeswitch` - FreeSWITCH status
- [ ] `GET /api/system/status/database` - Database status
- [ ] `GET /api/system/status/nodes` - Cluster node status
- [ ] `GET /api/system/stats` - Real-time statistics

---

## 23. Hospitality Module

- [ ] `GET /api/hospitality/rooms` - List hotel rooms
- [ ] `POST /api/hospitality/rooms` - Add room
- [ ] `PUT /api/hospitality/rooms/:id` - Update room
- [ ] `DELETE /api/hospitality/rooms/:id` - Delete room
- [ ] `POST /api/hospitality/rooms/:id/checkin` - Check-in guest
- [ ] `POST /api/hospitality/rooms/:id/checkout` - Check-out guest

### Wake-Up Calls

- [ ] `GET /api/hospitality/wake-up-calls` - List wake-up calls
- [ ] `POST /api/hospitality/wake-up-calls` - Create wake-up call
- [ ] `GET /api/hospitality/wake-up-calls/:id` - Get details
- [ ] `PUT /api/hospitality/wake-up-calls/:id` - Update
- [ ] `DELETE /api/hospitality/wake-up-calls/:id` - Delete

---

## 24. Locations (E911)

- [ ] `GET /api/locations` - List locations
- [ ] `POST /api/locations` - Create location
- [ ] `GET /api/locations/:id` - Get location
- [ ] `PUT /api/locations/:id` - Update location
- [ ] `DELETE /api/locations/:id` - Delete location

---

## 25. WebSocket / Real-time Events

- [ ] `WSS /ws/events` - Real-time event stream (calls, status)
- [ ] `WSS /ws/operator-panel` - Operator panel events
- [ ] `WSS /ws/queue/:id` - Queue real-time updates
- [ ] `WSS /ws/conference/:id` - Conference real-time updates
- [ ] `WSS /ws/device/:mac` - Device status updates

---

## Progress Summary

| Category | Total | Completed |
|----------|-------|-----------|
| Authentication | 7 | 0 |
| Tenants | 14 | 0 |
| Extensions | 12 | 0 |
| IVR/Auto Attendant | 22 | 0 |
| Call Routing | 20 | 0 |
| Devices | 21 | 0 |
| Queues & Ring Groups | 16 | 0 |
| Conferencing | 18 | 0 |
| Voicemail | 16 | 0 |
| Fax | 11 | 0 |
| Messaging | 11 | 0 |
| Call Recordings | 8 | 0 |
| Audio Library | 7 | 0 |
| Music on Hold | 11 | 0 |
| Feature Codes | 5 | 0 |
| Call Block | 5 | 0 |
| Call Broadcast | 8 | 0 |
| Speed Dials | 4 | 0 |
| Reports | 7 | 0 |
| Audit Log | 3 | 0 |
| User Portal | 22 | 0 |
| System Admin | 36 | 0 |
| Hospitality | 11 | 0 |
| Locations | 5 | 0 |
| WebSocket | 5 | 0 |
| **TOTAL** | **~300** | **0** |

---

## Implementation Notes

### Architecture Recommendations

1. **RESTful Design**: All endpoints follow REST conventions
2. **Tenant Scoping**: Most endpoints should be scoped by tenant (via JWT claims)
3. **Pagination**: List endpoints should support `?page=&limit=&sort=&order=`
4. **Filtering**: Support query params for filtering (e.g., `?direction=inbound&dateFrom=&dateTo=`)
5. **Rate Limiting**: Apply appropriate rate limits per endpoint category
6. **Authentication**: JWT Bearer tokens with role-based access control
7. **WebSocket Auth**: Token-based authentication for WebSocket connections

### Priority Order (Suggested)

1. **Phase 1 - Core**: Authentication, Extensions, Devices, Basic Routing
2. **Phase 2 - Communication**: Voicemail, Recordings, Messaging
3. **Phase 3 - Advanced Call Handling**: IVR, Queues, Conferences
4. **Phase 4 - Administration**: Reports, Audit Log, System Settings
5. **Phase 5 - Specialized**: Fax, Hospitality, Call Broadcast
