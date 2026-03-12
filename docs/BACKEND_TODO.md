# Backend API Development Todo List

This document tracks the backend API endpoints needed to support the CallSign-UI frontend application.

> **Last updated**: 2026-03-11  
> **Status**: ~400+ of ~420 endpoints implemented (95%+)

---

## 1. Authentication & Authorization

- [x] `POST /api/auth/login` - User login (returns JWT token)
- [x] `POST /api/auth/admin/login` - Admin login (system/tenant admins)
- [x] `POST /api/auth/extension/login` - Extension login (web client)
- [x] `POST /api/auth/register` - Self-registration (if enabled)
- [x] `POST /api/auth/logout` - Logout / invalidate token
- [x] `POST /api/auth/refresh` - Refresh JWT token
- [x] `GET /api/auth/me` - Get current user profile & permissions
- [x] `PUT /api/auth/password` - Change password
- [x] `POST /api/auth/password/reset` - Request password reset

---

## 2. Tenants (Multi-Tenant Management)

- [x] `GET /api/system/tenants` - List all tenants (system admin)
- [x] `POST /api/system/tenants` - Create new tenant
- [x] `GET /api/system/tenants/:id` - Get tenant details
- [x] `PUT /api/system/tenants/:id` - Update tenant
- [x] `DELETE /api/system/tenants/:id` - Delete tenant
- [x] `GET /api/users` - List tenant users
- [x] `POST /api/users` - Add user to tenant
- [x] `GET /api/users/:id` - Get user details
- [x] `PUT /api/users/:id` - Update user
- [x] `DELETE /api/users/:id` - Delete user
- [x] `GET /api/devices` - List tenant devices
- [ ] `GET /api/tenants/:id/stats` - Get tenant usage statistics

### Tenant Settings
- [x] `GET /api/tenant/settings` - Get tenant settings
- [x] `PUT /api/tenant/settings` - Update tenant settings
- [x] `GET /api/tenant/branding` - Get branding settings
- [x] `PUT /api/tenant/branding` - Update branding settings
- [x] `GET /api/tenant/smtp` - Get SMTP settings
- [x] `PUT /api/tenant/smtp` - Update SMTP settings
- [x] `POST /api/tenant/smtp/test` - Test SMTP settings
- [x] `GET /api/tenant/messaging` - Get messaging settings
- [x] `PUT /api/tenant/messaging` - Update messaging settings
- [x] `GET /api/tenant/hospitality` - Get hospitality settings
- [x] `PUT /api/tenant/hospitality` - Update hospitality settings

### Tenant Profiles (Service Plans)

- [x] `GET /api/system/tenant-profiles` - List all tenant profiles
- [x] `POST /api/system/tenant-profiles` - Create tenant profile
- [x] `GET /api/system/tenant-profiles/:id` - Get profile details
- [x] `PUT /api/system/tenant-profiles/:id` - Update profile
- [x] `DELETE /api/system/tenant-profiles/:id` - Delete profile

---

## 3. Extensions

- [x] `GET /api/extensions` - List all extensions
- [x] `POST /api/extensions` - Create new extension
- [x] `GET /api/extensions/:ext` - Get extension details
- [x] `PUT /api/extensions/:ext` - Update extension
- [x] `DELETE /api/extensions/:ext` - Delete extension
- [x] `GET /api/extensions/:ext/status` - Get real-time status (SIP registration, in-call)
- [x] `GET /api/extensions/:ext/call-rules` - List call handling rules
- [x] `POST /api/extensions/:ext/call-rules` - Create call handling rule
- [x] `PUT /api/extensions/:ext/call-rules/:ruleId` - Update call handling rule
- [x] `DELETE /api/extensions/:ext/call-rules/:ruleId` - Delete call handling rule
- [x] `POST /api/extensions/:ext/call-rules/reorder` - Reorder call handling rules

### Extension Profiles

- [x] `GET /api/extension-profiles` - List extension profiles
- [x] `POST /api/extension-profiles` - Create profile
- [x] `GET /api/extension-profiles/:id` - Get profile details
- [x] `PUT /api/extension-profiles/:id` - Update profile
- [x] `DELETE /api/extension-profiles/:id` - Delete profile
- [x] `GET /api/extension-profiles/:id/call-rules` - List profile call rules
- [x] `POST /api/extension-profiles/:id/call-rules` - Create profile call rule
- [x] `PUT /api/extension-profiles/:id/call-rules/:ruleId` - Update profile call rule
- [x] `DELETE /api/extension-profiles/:id/call-rules/:ruleId` - Delete profile call rule
- [x] `POST /api/extension-profiles/:id/call-rules/reorder` - Reorder profile call rules

---

## 4. IVR / Auto Attendant

### IVR Menus

- [x] `GET /api/ivr/menus` - List all IVR menus
- [x] `POST /api/ivr/menus` - Create IVR menu
- [x] `GET /api/ivr/menus/:id` - Get menu details
- [x] `PUT /api/ivr/menus/:id` - Update menu
- [x] `DELETE /api/ivr/menus/:id` - Delete menu

### Time Conditions

- [x] `GET /api/time-conditions` - List time conditions
- [x] `POST /api/time-conditions` - Create time condition
- [x] `GET /api/time-conditions/:id` - Get details
- [x] `PUT /api/time-conditions/:id` - Update time condition
- [x] `DELETE /api/time-conditions/:id` - Delete time condition
- [ ] `GET /api/time-conditions/:id/status` - Check if currently matched

### Holiday Lists

- [x] `GET /api/holidays` - List holiday lists
- [x] `POST /api/holidays` - Create holiday list
- [x] `GET /api/holidays/:id` - Get holiday list
- [x] `PUT /api/holidays/:id` - Update holiday list
- [x] `DELETE /api/holidays/:id` - Delete holiday list
- [x] `POST /api/holidays/:id/sync` - Sync holiday list

### Call Flows (Mode Toggles)

- [x] `GET /api/call-flows` - List call flows
- [x] `POST /api/call-flows` - Create call flow
- [x] `GET /api/call-flows/:id` - Get call flow details
- [x] `PUT /api/call-flows/:id` - Update call flow
- [x] `DELETE /api/call-flows/:id` - Delete call flow
- [x] `POST /api/call-flows/:id/toggle` - Toggle call flow state

---

## 5. Call Routing

### Phone Numbers (DIDs)

- [x] `GET /api/numbers` - List phone numbers
- [x] `POST /api/numbers` - Add/provision number
- [x] `GET /api/numbers/:id` - Get number details
- [x] `PUT /api/numbers/:id` - Update number routing
- [x] `DELETE /api/numbers/:id` - Remove number
- [x] `GET /api/system/numbers` - List all numbers (system admin)

### Inbound Routes

- [x] `GET /api/routing/inbound` - List inbound routes (ordered)
- [x] `POST /api/routing/inbound` - Create inbound route
- [x] `GET /api/routing/inbound/:id` - Get route details
- [x] `PUT /api/routing/inbound/:id` - Update route
- [x] `DELETE /api/routing/inbound/:id` - Delete route
- [x] `POST /api/routing/inbound/reorder` - Reorder routes

### Outbound Routes

- [x] `GET /api/routing/outbound` - List outbound routes
- [x] `POST /api/routing/outbound` - Create outbound route
- [x] `POST /api/routing/outbound/defaults` - Create default US/CAN routes
- [x] `GET /api/routing/outbound/:id` - Get route details
- [x] `PUT /api/routing/outbound/:id` - Update route
- [x] `DELETE /api/routing/outbound/:id` - Delete route
- [x] `POST /api/routing/outbound/reorder` - Reorder routes

### Call Blocks

- [x] `GET /api/routing/blocks` - List call blocks
- [x] `POST /api/routing/blocks` - Create call block
- [x] `PUT /api/routing/blocks/:id` - Update call block
- [x] `DELETE /api/routing/blocks/:id` - Delete call block

### Dial Plans

- [x] `GET /api/dial-plans` - List dial plans
- [x] `POST /api/dial-plans` - Create dial plan
- [x] `GET /api/dial-plans/:id` - Get dial plan
- [x] `PUT /api/dial-plans/:id` - Update dial plan
- [x] `DELETE /api/dial-plans/:id` - Delete dial plan

### Routing Utilities

- [x] `GET /api/routing/debug` - Route debugger
- [x] `POST /api/check-dial-code` - Dial code collision check

---

## 6. Devices (Provisioning)

- [x] `GET /api/devices` - List all provisioned devices
- [x] `POST /api/devices` - Add device (manual)
- [x] `GET /api/devices/:id` - Get device details
- [x] `PUT /api/devices/:id` - Update device
- [x] `DELETE /api/devices/:id` - Delete device
- [x] `POST /api/devices/:id/reprovision` - Trigger reprovision
- [x] `POST /api/devices/:id/assign-user` - Assign device to user
- [x] `POST /api/devices/:id/assign-profile` - Assign device to profile
- [x] `PUT /api/devices/:id/lines` - Update device lines
- [ ] `POST /api/devices/:id/reboot` - Send reboot command
- [ ] `GET /api/devices/:id/logs` - Get device logs
- [ ] `GET /api/devices/:id/status` - Get registration status
- [ ] `POST /api/devices/scan` - Scan network for unprovisioned devices

### Device Call Control

- [x] `POST /api/devices/:mac/hangup` - Hang up call
- [x] `POST /api/devices/:mac/transfer` - Transfer call
- [x] `POST /api/devices/:mac/hold` - Hold call
- [x] `POST /api/devices/:mac/dial` - Initiate call
- [x] `GET /api/devices/:mac/call-status` - Get call status

### Device Profiles (Tenant)

- [x] `GET /api/device-profiles` - List device profiles
- [x] `POST /api/device-profiles` - Create profile
- [x] `GET /api/device-profiles/:id` - Get profile details
- [x] `PUT /api/device-profiles/:id` - Update profile
- [x] `DELETE /api/device-profiles/:id` - Delete profile

### Device Templates (Tenant)

- [x] `GET /api/device-templates` - List device templates
- [x] `POST /api/device-templates` - Create template

### System Device Templates

- [x] `GET /api/system/device-templates` - List master templates
- [x] `POST /api/system/device-templates` - Create template
- [x] `GET /api/system/device-templates/:id` - Get template
- [x] `PUT /api/system/device-templates/:id` - Update template
- [x] `DELETE /api/system/device-templates/:id` - Delete template

### Device Manufacturers

- [x] `GET /api/system/device-manufacturers` - List manufacturers
- [x] `POST /api/system/device-manufacturers` - Create manufacturer
- [x] `PUT /api/system/device-manufacturers/:id` - Update manufacturer
- [x] `DELETE /api/system/device-manufacturers/:id` - Delete manufacturer

### Firmware Management

- [x] `GET /api/system/firmware` - List firmware
- [x] `POST /api/system/firmware` - Create firmware entry
- [x] `GET /api/system/firmware/:id` - Get firmware
- [x] `PUT /api/system/firmware/:id` - Update firmware
- [x] `DELETE /api/system/firmware/:id` - Delete firmware
- [x] `POST /api/system/firmware/:id/upload` - Upload firmware file
- [x] `POST /api/system/firmware/:id/set-default` - Set as default

### Provisioning Templates (Tenant)

- [x] `GET /api/provisioning-templates` - List templates
- [x] `POST /api/provisioning-templates` - Create template
- [x] `GET /api/provisioning-templates/:id` - Get template
- [x] `PUT /api/provisioning-templates/:id` - Update template
- [x] `DELETE /api/provisioning-templates/:id` - Delete template

### Client Registrations

- [x] `GET /api/registrations` - List client registrations
- [x] `POST /api/registrations/provision` - Provision client registration
- [x] `DELETE /api/registrations/:id` - Delete registration
- [x] `GET /api/registrations/unassigned` - List unassigned registrations
- [x] `POST /api/registrations/:id/assign` - Assign registration to extension
- [x] `GET /api/registrations/extension/:id` - List extension registrations

---

## 7. Call Queues & Ring Groups

### Queues

- [x] `GET /api/queues` - List call queues
- [x] `POST /api/queues` - Create queue
- [x] `GET /api/queues/:id` - Get queue details
- [x] `PUT /api/queues/:id` - Update queue
- [x] `DELETE /api/queues/:id` - Delete queue
- [x] `GET /api/queues/:id/agents` - List queue agents
- [x] `POST /api/queues/:id/agents` - Add agent to queue
- [x] `DELETE /api/queues/:id/agents/:agentId` - Remove agent from queue
- [x] `POST /api/queues/:id/agents/:agentId/pause` - Pause agent
- [x] `POST /api/queues/:id/agents/:agentId/unpause` - Unpause agent

### Ring Groups

- [x] `GET /api/ring-groups` - List ring groups
- [x] `POST /api/ring-groups` - Create ring group
- [x] `GET /api/ring-groups/:id` - Get ring group details
- [x] `PUT /api/ring-groups/:id` - Update ring group
- [x] `DELETE /api/ring-groups/:id` - Delete ring group

---

## 8. Conferencing

- [x] `GET /api/conferences` - List conference rooms
- [x] `POST /api/conferences` - Create conference room
- [x] `GET /api/conferences/:id` - Get conference details
- [x] `PUT /api/conferences/:id` - Update conference
- [x] `DELETE /api/conferences/:id` - Delete conference
- [x] `GET /api/conferences/:id/stats` - Conference statistics
- [x] `GET /api/conferences/:id/sessions` - Conference sessions
- [x] `GET /api/conferences/sessions/:sessionId/participants` - Session participants

### Live Conference Control

- [x] `GET /api/conferences/live` - List live conferences
- [x] `GET /api/conferences/live/:name` - Get live conference
- [x] `POST /api/conferences/live/:name/mute/:memberId` - Mute participant
- [x] `POST /api/conferences/live/:name/unmute/:memberId` - Unmute participant
- [x] `POST /api/conferences/live/:name/deaf/:memberId` - Deaf participant
- [x] `POST /api/conferences/live/:name/undeaf/:memberId` - Undeaf participant
- [x] `POST /api/conferences/live/:name/kick/:memberId` - Kick participant
- [x] `POST /api/conferences/live/:name/lock` - Lock conference
- [x] `POST /api/conferences/live/:name/unlock` - Unlock conference
- [x] `POST /api/conferences/live/:name/record/start` - Start recording
- [x] `POST /api/conferences/live/:name/record/stop` - Stop recording
- [x] `POST /api/conferences/live/:name/mute-all` - Mute all
- [x] `POST /api/conferences/live/:name/unmute-all` - Unmute all
- [x] `POST /api/conferences/live/:name/floor/:memberId` - Set floor

### Conference Profiles

- [ ] `GET /api/conference-profiles` - List conference profiles
- [ ] `POST /api/conference-profiles` - Create profile
- [ ] `GET /api/conference-profiles/:id` - Get profile
- [ ] `PUT /api/conference-profiles/:id` - Update profile
- [ ] `DELETE /api/conference-profiles/:id` - Delete profile

---

## 9. Voicemail

### Voicemail Boxes

- [x] `GET /api/voicemail/boxes` - List voicemail boxes
- [x] `POST /api/voicemail/boxes` - Create voicemail box
- [x] `GET /api/voicemail/boxes/:ext` - Get box details
- [x] `PUT /api/voicemail/boxes/:ext` - Update box
- [x] `DELETE /api/voicemail/boxes/:ext` - Delete box
- [x] `GET /api/voicemail/boxes/:ext/messages` - List messages
- [x] `GET /api/voicemail/messages/:id` - Get message
- [x] `DELETE /api/voicemail/messages/:id` - Delete message
- [x] `GET /api/voicemail/messages/:id/stream` - Stream message audio
- [x] `POST /api/voicemail/messages/:id/read` - Mark as read

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

### Fax Boxes

- [x] `GET /api/fax/boxes` - List fax boxes
- [x] `POST /api/fax/boxes` - Create fax box
- [x] `GET /api/fax/boxes/:boxId` - Get fax box details
- [x] `PUT /api/fax/boxes/:boxId` - Update fax box
- [x] `DELETE /api/fax/boxes/:boxId` - Delete fax box

### Fax Jobs

- [x] `GET /api/fax/jobs` - List fax jobs
- [x] `GET /api/fax/jobs/:jobId` - Get fax job details
- [x] `DELETE /api/fax/jobs/:jobId` - Delete fax job
- [x] `GET /api/fax/jobs/:jobId/download` - Download fax
- [x] `POST /api/fax/jobs/:jobId/retry` - Retry fax

### Fax Actions

- [x] `POST /api/fax/send` - Send fax
- [x] `GET /api/fax/active` - Get active faxes
- [x] `GET /api/fax/stats` - Fax statistics

### Fax Endpoints

- [x] `GET /api/fax/endpoints` - List fax endpoints
- [x] `POST /api/fax/endpoints` - Create fax endpoint
- [x] `PUT /api/fax/endpoints/:epId` - Update fax endpoint
- [x] `DELETE /api/fax/endpoints/:epId` - Delete fax endpoint

---

## 11. Messaging (SMS/MMS)

### Conversations

- [x] `GET /api/messaging/conversations` - List conversations
- [x] `GET /api/messaging/conversations/:id` - Get conversation messages
- [x] `POST /api/messaging/send` - Send message
- [ ] `PUT /api/messaging/conversations/:id/read` - Mark as read
- [ ] `DELETE /api/messaging/conversations/:id` - Delete conversation

### SMS Number Management (Tenant)

- [x] `GET /api/messaging/numbers` - List SMS numbers
- [x] `PUT /api/messaging/numbers/:id/sms` - Configure SMS number
- [x] `GET /api/messaging/numbers/:id/assignments` - List number assignments
- [x] `POST /api/messaging/numbers/:id/assignments` - Assign number
- [x] `DELETE /api/messaging/numbers/:id/assignments/:assignId` - Unassign number

### Messaging Providers (System)

- [x] `GET /api/system/messaging-providers` - List providers
- [x] `POST /api/system/messaging-providers` - Add provider
- [x] `GET /api/system/messaging-providers/:id` - Get provider
- [x] `PUT /api/system/messaging-providers/:id` - Update provider
- [x] `DELETE /api/system/messaging-providers/:id` - Delete provider
- [ ] `POST /api/system/messaging-providers/:id/test` - Test provider

### Messaging Numbers (System)

- [x] `GET /api/system/messaging-numbers` - List messaging numbers
- [x] `POST /api/system/messaging-numbers` - Create messaging number
- [x] `PUT /api/system/messaging-numbers/:id` - Update messaging number
- [x] `DELETE /api/system/messaging-numbers/:id` - Delete messaging number

---

## 12. Call Recordings

- [x] `GET /api/recordings` - List recordings (with filtering)
- [x] `GET /api/recordings/config` - Get recording configuration
- [x] `GET /api/recordings/:id` - Get recording metadata
- [x] `DELETE /api/recordings/:id` - Delete recording
- [x] `GET /api/recordings/:id/stream` - Stream recording audio
- [x] `GET /api/recordings/:id/download` - Download recording
- [x] `PUT /api/recordings/:id/notes` - Update recording notes
- [x] `GET /api/recordings/:id/transcription` - Get transcription

---

## 13. Audio Library (Prompts/Greetings)

- [x] `GET /api/audio-library` - List audio files
- [x] `POST /api/audio-library` - Upload audio file
- [x] `GET /api/audio-library/:id/stream` - Stream audio
- [x] `PUT /api/audio-library/:id` - Update metadata
- [x] `DELETE /api/audio-library/:id` - Delete audio
- [ ] `POST /api/audio-library/record` - Record audio (TTS or phone)

---

## 14. Music on Hold / Streams

- [x] `GET /api/music-on-hold` - List MoH streams
- [x] `POST /api/music-on-hold` - Create stream
- [x] `GET /api/music-on-hold/:id` - Get stream details
- [x] `PUT /api/music-on-hold/:id` - Update stream
- [x] `DELETE /api/music-on-hold/:id` - Delete stream
- [ ] `POST /api/music-on-hold/:id/files` - Add file to stream
- [ ] `DELETE /api/music-on-hold/:id/files/:fileId` - Remove file

### System Media

- [x] `GET /api/system/media/sounds` - List system sounds
- [x] `POST /api/system/media/sounds` - Upload system sound
- [x] `GET /api/system/media/sounds/stream` - Stream sound
- [x] `GET /api/system/media/music` - List system music
- [x] `POST /api/system/media/music` - Upload system music
- [x] `GET /api/system/media/music/stream` - Stream music

### Tenant Media

- [x] `GET /api/media/sounds` - List tenant sounds
- [x] `POST /api/media/sounds` - Upload tenant sound
- [x] `DELETE /api/media/sounds` - Delete tenant sound
- [x] `GET /api/media/music` - List tenant music
- [x] `POST /api/media/music` - Upload tenant music
- [x] `DELETE /api/media/music` - Delete tenant music

---

## 15. Feature Codes

- [x] `GET /api/feature-codes` - List feature codes
- [x] `GET /api/feature-codes/system` - List system feature codes
- [x] `POST /api/feature-codes` - Create feature code
- [x] `GET /api/feature-codes/:id` - Get feature code
- [x] `PUT /api/feature-codes/:id` - Update feature code
- [x] `DELETE /api/feature-codes/:id` - Delete feature code

---

## 16. Speed Dials

- [x] `GET /api/speed-dials` - List speed dial groups
- [x] `POST /api/speed-dials` - Create speed dial group
- [x] `GET /api/speed-dials/:id` - Get speed dial group
- [x] `PUT /api/speed-dials/:id` - Update speed dial group
- [x] `DELETE /api/speed-dials/:id` - Delete speed dial group

---

## 17. Reports & Analytics

- [x] `GET /api/reports/call-volume` - Call volume report
- [x] `GET /api/reports/agent-performance` - Agent performance metrics
- [x] `GET /api/reports/queue-stats` - Queue statistics
- [x] `GET /api/reports/extension-usage` - Extension usage
- [x] `GET /api/reports/kpi` - Key performance indicators
- [x] `GET /api/reports/number-usage` - Number usage statistics
- [x] `GET /api/reports/export` - Export report (CSV/PDF)

---

## 18. CDR & Audit Log

### CDR

- [x] `GET /api/cdr` - List CDR entries
- [x] `GET /api/cdr/:id` - Get CDR entry details
- [x] `GET /api/cdr/export` - Export CDR

### Audit Log

- [x] `GET /api/audit-logs` - List audit log entries

---

## 19. Call Broadcast

- [x] `GET /api/broadcast` - List broadcast campaigns
- [x] `POST /api/broadcast` - Create campaign
- [x] `GET /api/broadcast/:id` - Get campaign details
- [x] `PUT /api/broadcast/:id` - Update campaign
- [x] `DELETE /api/broadcast/:id` - Delete campaign
- [x] `POST /api/broadcast/:id/start` - Start campaign
- [x] `POST /api/broadcast/:id/stop` - Stop campaign
- [x] `GET /api/broadcast/:id/stats` - Campaign statistics

---

## 20. User Portal (End User)

### User Devices

- [x] `GET /api/user/devices` - List user's devices

### Call History

- [x] `GET /api/user/call-history` - User's call history

### Contacts

- [x] `GET /api/user/contacts` - List user contacts
- [x] `POST /api/user/contacts` - Add contact
- [ ] `PUT /api/user/contacts/:id` - Update contact
- [ ] `DELETE /api/user/contacts/:id` - Delete contact

### User Settings

- [x] `GET /api/user/settings` - Get user settings
- [x] `PUT /api/user/settings` - Update settings

### User Voicemail

- [x] `GET /api/user/voicemail` - List voicemail messages

---

## 21. Extension Portal

- [x] `GET /api/extension/portal/devices` - List extension devices
- [x] `GET /api/extension/portal/call-history` - Extension call history
- [x] `GET /api/extension/portal/voicemail` - Extension voicemail
- [x] `GET /api/extension/portal/settings` - Get extension settings
- [x] `PUT /api/extension/portal/settings` - Update extension settings
- [x] `PUT /api/extension/portal/password` - Change extension password
- [x] `GET /api/extension/portal/contacts` - List extension contacts
- [x] `POST /api/extension/portal/contacts` - Create extension contact

---

## 22. Contacts

- [x] `GET /api/contacts` - List contacts
- [x] `POST /api/contacts` - Add contact
- [x] `GET /api/contacts/:id` - Get contact
- [x] `PUT /api/contacts/:id` - Update contact
- [x] `DELETE /api/contacts/:id` - Delete contact
- [x] `POST /api/contacts/:id/sync` - Sync contact
- [x] `GET /api/contacts/lookup` - Lookup by phone

---

## 23. Chat

- [x] `GET /api/chat/threads` - List chat threads
- [x] `POST /api/chat/threads` - Create chat thread
- [x] `GET /api/chat/threads/:id` - Get chat thread
- [x] `POST /api/chat/threads/:id/messages` - Send chat message
- [x] `GET /api/chat/rooms` - List chat rooms
- [x] `POST /api/chat/rooms` - Create chat room
- [x] `POST /api/chat/rooms/:id/join` - Join chat room
- [x] `GET /api/chat/queues` - List chat queues
- [x] `POST /api/chat/queues` - Create chat queue

---

## 24. Paging Groups

- [x] `GET /api/page-groups` - List paging groups
- [x] `POST /api/page-groups` - Create paging group
- [x] `GET /api/page-groups/:id` - Get paging group
- [x] `PUT /api/page-groups/:id` - Update paging group
- [x] `DELETE /api/page-groups/:id` - Delete paging group

---

## 25. System Administration

### Gateways / Trunks

- [x] `GET /api/system/gateways` - List gateways
- [x] `POST /api/system/gateways` - Create gateway
- [x] `GET /api/system/gateways/:id` - Get gateway
- [x] `PUT /api/system/gateways/:id` - Update gateway
- [x] `DELETE /api/system/gateways/:id` - Delete gateway
- [x] `GET /api/system/gateways/status` - Gateway status

### Bridges

- [x] `GET /api/system/bridges` - List bridges
- [x] `POST /api/system/bridges` - Create bridge
- [x] `GET /api/system/bridges/:id` - Get bridge
- [x] `PUT /api/system/bridges/:id` - Update bridge
- [x] `DELETE /api/system/bridges/:id` - Delete bridge

### SIP Profiles

- [x] `GET /api/system/sip-profiles` - List SIP profiles
- [x] `POST /api/system/sip-profiles` - Create profile
- [x] `POST /api/system/sip-profiles/sync` - Sync from disk
- [x] `GET /api/system/sip-profiles/:id` - Get profile
- [x] `PUT /api/system/sip-profiles/:id` - Update profile
- [x] `DELETE /api/system/sip-profiles/:id` - Delete profile

### Sofia Control

- [x] `GET /api/system/sofia/status` - Get Sofia status
- [x] `GET /api/system/sofia/profiles/:name/status` - Get profile status
- [x] `GET /api/system/sofia/profiles/:name/registrations` - Get registrations
- [x] `GET /api/system/sofia/profiles/:name/gateways` - Get gateway status
- [x] `POST /api/system/sofia/profiles/:name/restart` - Restart profile
- [x] `POST /api/system/sofia/profiles/:name/start` - Start profile
- [x] `POST /api/system/sofia/profiles/:name/stop` - Stop profile
- [x] `POST /api/system/sofia/reload-xml` - Reload XML

### Global Dial Plans

- [x] `GET /api/system/dialplans` - List global dial plans
- [x] `POST /api/system/dialplans` - Create dial plan
- [x] `GET /api/system/dialplans/:id` - Get dial plan
- [x] `PUT /api/system/dialplans/:id` - Update dial plan
- [x] `DELETE /api/system/dialplans/:id` - Delete dial plan

### ACLs

- [x] `GET /api/system/acls` - List ACLs
- [x] `POST /api/system/acls` - Create ACL
- [x] `GET /api/system/acls/:id` - Get ACL
- [x] `PUT /api/system/acls/:id` - Update ACL
- [x] `DELETE /api/system/acls/:id` - Delete ACL
- [x] `POST /api/system/acls/:id/nodes` - Add ACL node
- [x] `PUT /api/system/acls/:id/nodes/:nodeId` - Update ACL node
- [x] `DELETE /api/system/acls/:id/nodes/:nodeId` - Delete ACL node

### System Settings

- [x] `GET /api/system/settings` - Get all system settings
- [x] `PUT /api/system/settings` - Update settings

### System Logs

- [x] `GET /api/system/logs` - Get system logs

### Infrastructure Status

- [x] `GET /api/system/status` - Overall system status
- [x] `GET /api/system/stats` - Real-time statistics

### Security

- [x] `GET /api/system/security/banned-ips` - List banned IPs
- [x] `POST /api/system/security/banned-ips` - Report banned IP
- [x] `DELETE /api/system/security/banned-ips/:ip` - Unban IP

### Config Inspector

- [x] `GET /api/system/xml/debug` - Debug XML generation
- [x] `GET /api/system/config/files` - List config files
- [x] `GET /api/system/config/file` - Read config file

---

## 26. Hospitality Module

### Rooms

- [x] `GET /api/hospitality/rooms` - List hotel rooms
- [x] `POST /api/hospitality/rooms` - Add room
- [x] `GET /api/hospitality/rooms/:id` - Get room details
- [x] `PUT /api/hospitality/rooms/:id` - Update room
- [x] `DELETE /api/hospitality/rooms/:id` - Delete room
- [x] `POST /api/hospitality/rooms/:id/checkin` - Check-in guest
- [x] `POST /api/hospitality/rooms/:id/checkout` - Check-out guest
- [x] `POST /api/hospitality/rooms/:id/wakeup` - Schedule wake-up call

---

## 27. Locations (E911)

- [x] `GET /api/tenant/locations` - List locations
- [x] `POST /api/tenant/locations` - Create location
- [x] `GET /api/tenant/locations/:id` - Get location
- [x] `PUT /api/tenant/locations/:id` - Update location
- [x] `DELETE /api/tenant/locations/:id` - Delete location

---

## 28. Live Operations

- [x] `POST /api/live/recording/start` - Start call recording
- [x] `POST /api/live/recording/stop` - Stop call recording
- [x] `GET /api/live/calls` - Get active calls
- [x] `GET /api/live/queue-stats` - Get live queue statistics
- [x] `POST /api/live/wakeup/schedule` - Schedule wake-up via ESL
- [x] `GET /api/live/registrations` - Get device registrations

---

## 29. Operator Panel

- [x] `GET /api/operator-panel` - Get operator panel data

---

## 30. WebSocket / Real-time Events

- [x] `WSS /api/ws/notifications` - Real-time notifications
- [x] `WSS /api/system/console` - FreeSWITCH console
- [x] `WSS /api/ws` - Generic WebSocket endpoint
- [ ] `WSS /ws/operator-panel` - Operator panel events
- [ ] `WSS /ws/queue/:id` - Queue real-time updates
- [ ] `WSS /ws/conference/:id` - Conference real-time updates
- [ ] `WSS /ws/device/:mac` - Device status updates

---

## 31. FreeSWITCH Integration

- [x] `POST /api/freeswitch/directory` - XML CURL directory
- [x] `POST /api/freeswitch/dialplan` - XML CURL dialplan
- [x] `POST /api/freeswitch/configuration` - XML CURL configuration
- [x] `POST /api/freeswitch/xmlapi` - Legacy combined handler
- [x] `POST /api/freeswitch/cdr` - CDR ingestion
- [x] `GET /api/freeswitch/cache/flush` - Flush cache
- [x] `GET /api/freeswitch/cache/stats` - Cache statistics

---

## 32. Webhooks

- [x] `POST /api/webhooks/telnyx/inbound` - Telnyx inbound webhook
- [x] `POST /api/webhooks/telnyx/status` - Telnyx status webhook

---

## 33. Provisioning (Public)

- [x] `GET /api/provision/:tenant/:secret/:mac` - Secure provisioning
- [x] `GET /provisioning/:mac/:filename` - Public provisioning

---

## 34. Internal API

- [x] `POST /api/internal/fail2ban/report` - Report banned IP (X-Internal-Key auth)

---

## Progress Summary

| Category | Total | Completed |
|----------|-------|-----------|
| Authentication | 9 | 9 |
| Tenants | 12 | 11 |
| Tenant Settings | 11 | 11 |
| Tenant Profiles | 5 | 5 |
| Extensions | 11 | 11 |
| Extension Profiles | 10 | 10 |
| IVR/Auto Attendant | 5 | 5 |
| Time Conditions | 6 | 5 |
| Holiday Lists | 6 | 6 |
| Call Flows | 6 | 6 |
| Call Routing | 22 | 22 |
| Devices | 44 | 40 |
| Client Registrations | 6 | 6 |
| Queues & Ring Groups | 15 | 15 |
| Conferencing | 23 | 18 |
| Voicemail | 16 | 10 |
| Fax | 15 | 15 |
| Messaging/SMS | 17 | 15 |
| Call Recordings | 8 | 8 |
| Audio Library | 6 | 5 |
| Music on Hold | 13 | 11 |
| Feature Codes | 6 | 6 |
| Speed Dials | 5 | 5 |
| Reports | 7 | 7 |
| CDR & Audit | 4 | 4 |
| Call Broadcast | 8 | 8 |
| User Portal | 9 | 7 |
| Extension Portal | 8 | 8 |
| Contacts | 7 | 7 |
| Chat | 9 | 9 |
| Paging Groups | 5 | 5 |
| System Admin | 55 | 55 |
| Hospitality | 8 | 8 |
| Locations | 5 | 5 |
| Live Operations | 6 | 6 |
| Operator Panel | 1 | 1 |
| WebSocket | 7 | 3 |
| FreeSWITCH | 7 | 7 |
| Webhooks | 2 | 2 |
| Provisioning | 2 | 2 |
| Internal API | 1 | 1 |
| **TOTAL** | **~420** | **~400+** |

---

## Implementation Notes

### Architecture

1. **RESTful Design**: All endpoints follow REST conventions
2. **Tenant Scoping**: Most endpoints are scoped by tenant (via JWT claims)
3. **Pagination**: List endpoints support `?page=&limit=&sort=&order=`
4. **Filtering**: Support query params for filtering (e.g., `?direction=inbound&dateFrom=&dateTo=`)
5. **Rate Limiting**: Apply appropriate rate limits per endpoint category
6. **Authentication**: JWT Bearer tokens with role-based access control
7. **WebSocket Auth**: Token-based authentication for WebSocket connections
8. **Framework**: Go Fiber (migrated from Iris)

### Current Focus

1. **Phase 1 - Core**: ✅ Complete
2. **Phase 2 - Communication**: ✅ Voicemail, Recordings, Messaging, Chat, Fax
3. **Phase 3 - Advanced Call Handling**: ✅ IVR, Queues + agents, Conferences + live control
4. **Phase 4 - Administration**: ✅ Reports, Audit Log, System Settings, Security
5. **Phase 5 - Specialized**: ✅ Fax, Hospitality, Call Broadcast, Locations
6. **Phase 6 - Refinement**: 🔄 WebSocket expansion, voicemail delivery, conference profiles
