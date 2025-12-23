# Backend API Development Todo List

This document tracks the backend API endpoints needed to support the CallSign-UI frontend application.

> **Last updated**: 2025-12-16  
> **Status**: ~180+ of ~300 endpoints implemented

---

## 1. Authentication & Authorization

- [x] `POST /api/auth/login` - User login (returns JWT token)
- [x] `POST /api/auth/admin/login` - Admin login (system/tenant admins)
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
- [ ] `GET /api/numbers/:id/stats` - Number usage statistics

### Inbound Routes

- [x] `GET /api/routing/inbound` - List inbound routes (ordered)
- [x] `POST /api/routing/inbound` - Create inbound route
- [ ] `GET /api/routing/inbound/:id` - Get route details
- [ ] `PUT /api/routing/inbound/:id` - Update route
- [ ] `DELETE /api/routing/inbound/:id` - Delete route
- [ ] `PUT /api/routing/inbound/reorder` - Reorder routes (batch)

### Outbound Routes

- [x] `GET /api/routing/outbound` - List outbound routes
- [x] `POST /api/routing/outbound` - Create outbound route
- [x] `POST /api/routing/outbound/defaults` - Create default US/CAN routes
- [ ] `GET /api/routing/outbound/:id` - Get route details
- [ ] `PUT /api/routing/outbound/:id` - Update route
- [ ] `DELETE /api/routing/outbound/:id` - Delete route
- [ ] `PUT /api/routing/outbound/reorder` - Reorder routes (batch)

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

### Routing Settings

- [x] `GET /api/routing/debug` - Route debugger

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

---

## 7. Call Queues & Ring Groups

### Queues

- [x] `GET /api/queues` - List call queues
- [x] `POST /api/queues` - Create queue
- [x] `GET /api/queues/:id` - Get queue details
- [x] `PUT /api/queues/:id` - Update queue
- [x] `DELETE /api/queues/:id` - Delete queue
- [ ] `GET /api/queues/:id/stats` - Real-time queue statistics
- [ ] `GET /api/queues/:id/agents` - List queue agents
- [ ] `POST /api/queues/:id/agents` - Add agent to queue
- [ ] `DELETE /api/queues/:id/agents/:agentId` - Remove agent from queue
- [ ] `PUT /api/queues/:id/agents/:agentId/pause` - Pause agent
- [ ] `PUT /api/queues/:id/agents/:agentId/unpause` - Unpause agent

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

- [x] `GET /api/messaging/conversations` - List conversations
- [x] `GET /api/messaging/conversations/:id` - Get conversation messages
- [x] `POST /api/messaging/send` - Send message
- [ ] `PUT /api/messaging/conversations/:id/read` - Mark as read
- [ ] `DELETE /api/messaging/conversations/:id` - Delete conversation

### Messaging Providers (System)

- [x] `GET /api/system/messaging-providers` - List providers
- [x] `POST /api/system/messaging-providers` - Add provider
- [x] `GET /api/system/messaging-providers/:id` - Get provider
- [x] `PUT /api/system/messaging-providers/:id` - Update provider
- [x] `DELETE /api/system/messaging-providers/:id` - Delete provider
- [ ] `POST /api/system/messaging-providers/:id/test` - Test provider

---

## 12. Call Recordings

- [x] `GET /api/recordings` - List recordings (with filtering)
- [x] `GET /api/recordings/:id` - Get recording metadata
- [ ] `GET /api/recordings/:id/audio` - Stream recording audio
- [ ] `GET /api/recordings/:id/download` - Download recording
- [x] `DELETE /api/recordings/:id` - Delete recording
- [ ] `POST /api/recordings/:id/email` - Email recording
- [ ] `POST /api/recordings/export` - Batch export
- [ ] `GET /api/recordings/stats` - Recording statistics

---

## 13. Audio Library (Prompts/Greetings)

- [x] `GET /api/audio-library` - List audio files
- [x] `POST /api/audio-library` - Upload audio file
- [ ] `GET /api/audio-library/:id` - Get audio metadata
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

- [ ] `GET /api/reports/call-volume` - Call volume report
- [ ] `GET /api/reports/call-detail` - CDR (Call Detail Records)
- [ ] `GET /api/reports/agent-performance` - Agent performance metrics
- [ ] `GET /api/reports/queue-stats` - Queue statistics
- [ ] `GET /api/reports/extension-usage` - Extension usage
- [ ] `GET /api/reports/kpi` - Key performance indicators
- [ ] `POST /api/reports/export` - Export report (CSV/PDF)

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

- [ ] `GET /api/call-broadcast` - List broadcast campaigns
- [ ] `POST /api/call-broadcast` - Create campaign
- [ ] `GET /api/call-broadcast/:id` - Get campaign details
- [ ] `PUT /api/call-broadcast/:id` - Update campaign
- [ ] `DELETE /api/call-broadcast/:id` - Delete campaign
- [ ] `POST /api/call-broadcast/:id/start` - Start campaign
- [ ] `POST /api/call-broadcast/:id/stop` - Stop campaign
- [ ] `GET /api/call-broadcast/:id/stats` - Campaign statistics

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

## 21. Contacts

- [x] `GET /api/contacts` - List contacts
- [x] `POST /api/contacts` - Add contact
- [x] `GET /api/contacts/:id` - Get contact
- [x] `PUT /api/contacts/:id` - Update contact
- [x] `DELETE /api/contacts/:id` - Delete contact
- [x] `POST /api/contacts/:id/sync` - Sync contact
- [x] `GET /api/contacts/lookup` - Lookup by phone

---

## 22. Chat

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

## 23. Paging Groups

- [x] `GET /api/page-groups` - List paging groups
- [x] `POST /api/page-groups` - Create paging group
- [x] `GET /api/page-groups/:id` - Get paging group
- [x] `PUT /api/page-groups/:id` - Update paging group
- [x] `DELETE /api/page-groups/:id` - Delete paging group

---

## 24. System Administration

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

## 25. Hospitality Module

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

## 26. Locations (E911)

- [ ] `GET /api/locations` - List locations
- [ ] `POST /api/locations` - Create location
- [ ] `GET /api/locations/:id` - Get location
- [ ] `PUT /api/locations/:id` - Update location
- [ ] `DELETE /api/locations/:id` - Delete location

---

## 27. WebSocket / Real-time Events

- [x] `WSS /api/ws/notifications` - Real-time notifications
- [x] `WSS /api/system/console` - FreeSWITCH console
- [ ] `WSS /ws/operator-panel` - Operator panel events
- [ ] `WSS /ws/queue/:id` - Queue real-time updates
- [ ] `WSS /ws/conference/:id` - Conference real-time updates
- [ ] `WSS /ws/device/:mac` - Device status updates

---

## 28. FreeSWITCH Integration

- [x] `POST /api/freeswitch/directory` - XML CURL directory
- [x] `POST /api/freeswitch/dialplan` - XML CURL dialplan
- [x] `POST /api/freeswitch/configuration` - XML CURL configuration
- [x] `POST /api/freeswitch/xmlapi` - Legacy combined handler
- [x] `POST /api/freeswitch/cdr` - CDR ingestion
- [x] `GET /api/freeswitch/cache/flush` - Flush cache
- [x] `GET /api/freeswitch/cache/stats` - Cache statistics

---

## Progress Summary

| Category | Total | Completed |
|----------|-------|-----------|
| Authentication | 7 | 7 |
| Tenants | 16 | 16 |
| Tenant Settings | 11 | 11 |
| Tenant Profiles | 5 | 5 |
| Extensions | 11 | 11 |
| Extension Profiles | 10 | 10 |
| IVR/Auto Attendant | 5 | 5 |
| Time Conditions | 6 | 5 |
| Holiday Lists | 6 | 6 |
| Call Flows | 6 | 6 |
| Call Routing | 21 | 13 |
| Devices | 45 | 40 |
| Queues & Ring Groups | 16 | 10 |
| Conferencing | 18 | 5 |
| Voicemail | 16 | 10 |
| Fax | 11 | 0 |
| Messaging | 11 | 8 |
| Call Recordings | 8 | 3 |
| Audio Library | 7 | 5 |
| Music on Hold | 13 | 11 |
| Feature Codes | 6 | 6 |
| Speed Dials | 5 | 5 |
| CDR & Audit | 4 | 4 |
| Call Broadcast | 8 | 0 |
| User Portal | 9 | 7 |
| Contacts | 7 | 7 |
| Chat | 9 | 9 |
| Paging Groups | 5 | 5 |
| System Admin | 48 | 48 |
| Hospitality | 11 | 0 |
| Locations | 5 | 0 |
| WebSocket | 6 | 2 |
| FreeSWITCH | 7 | 7 |
| **TOTAL** | **~375** | **~280** |

---

## Implementation Notes

### Architecture Recommendations

1. **RESTful Design**: All endpoints follow REST conventions
2. **Tenant Scoping**: Most endpoints are scoped by tenant (via JWT claims)
3. **Pagination**: List endpoints support `?page=&limit=&sort=&order=`
4. **Filtering**: Support query params for filtering (e.g., `?direction=inbound&dateFrom=&dateTo=`)
5. **Rate Limiting**: Apply appropriate rate limits per endpoint category
6. **Authentication**: JWT Bearer tokens with role-based access control
7. **WebSocket Auth**: Token-based authentication for WebSocket connections

### Priority Order (Current Focus)

1. **Phase 1 - Core**: ✅ Authentication, Extensions, Devices, Basic Routing
2. **Phase 2 - Communication**: ✅ Voicemail, Recordings, Messaging (partial)
3. **Phase 3 - Advanced Call Handling**: ✅ IVR, Queues, Conferences (partial - need live control)
4. **Phase 4 - Administration**: ✅ Reports (basic), Audit Log, System Settings
5. **Phase 5 - Specialized**: 🔲 Fax, Hospitality, Call Broadcast
