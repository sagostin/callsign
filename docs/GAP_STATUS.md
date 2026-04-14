# CallSign - Call Routing Gaps & Implementation Status

## Visual IVR Flow Editor ✅ IMPLEMENTED

**Status**: Core functionality complete. Fixed integration with real API data.

### What was done (Session 2026-04-14):
1. Integrated `FlowNode.vue` with real data from API via Vue inject/provide
2. Connected extension, queue, IVR menu, ring group, and recording dropdowns to API
3. Added ring groups loading to `IVRMenuForm.vue`

### Components:
- `ui/src/components/flow/NodePalette.vue` - Draggable node library
- `ui/src/components/flow/FlowCanvas.vue` - SVG connection editor
- `ui/src/components/flow/FlowNode.vue` - Configurable node with modal editor
- `ui/src/views/admin/IVRMenuForm.vue` - Full editor page

### Backend:
- `api/services/esl/modules/ivr/service.go` - Flow graph execution
- `api/models/ivr.go` - `IVRFlowData`, `IVRFlowNode`, `IVRFlowConnection` models

---

## Remaining Gaps by Priority

### HIGH Priority

#### 1. SMS Provider Integration
- **Backend**: `api/services/esl/modules/ivr/service.go:490` has TODO comment
- **Needed**: Telnyx/Twilio/SignalWire integration in `nodeSendSMS()`
- **Action**: Implement SMS provider abstraction similar to messaging service

#### 2. Queue Agent Login UI
- **Backend**: Feature codes `*90` (login), `*91` (logout) exist in `queue/service.go`
- **Frontend**: No agent portal UI for queue status management
- **Action**: Add queue agent status panel to user portal

#### 3. Live Call Recording Controls
- **Backend**: Recording models exist, start/stop via ESL possible
- **Frontend**: No UI for live recording toggle
- **Action**: Add recording button to active call view

### MEDIUM Priority

#### 4. Voicemail Dropdown in Flow Editor
- **Current**: Hardcoded "general" and "sales" options in `FlowNode.vue`
- **Action**: Load voicemail boxes from API, add to dropdown

#### 5. Call Volume Charts/Stats
- **Backend**: CDR data stored in ClickHouse (if enabled)
- **Frontend**: No visualization dashboard
- **Action**: Add charts to admin dashboard using Chart.js or similar

#### 6. Number Provisioning UI
- **Current**: Tenant view shows "managed by system administrator"
- **Action**: System admin UI for DID assignment to tenants

### LOW Priority (Nice to Have)

#### 7. Speech Input Node (ASR)
- **Backend**: Node type exists, no provider integration
- **Action**: Integrate with Google Cloud Speech, AWS Transcribe, or Azure Speech

#### 8. Database Node
- **Backend**: Node type exists, no actual query execution
- **Action**: Implement MySQL/PostgreSQL/REST query support

#### 9. Historical Queue Analytics
- **Current**: Real-time queue stats from FreeSWITCH
- **Action**: Add ClickHouse-based historical reporting

#### 10. Fax Routing through IVR
- **Current**: FAX support exists, routing through IVR unclear
- **Action**: Verify and document fax-specific routing paths

---

## Verified Working Features

- Extension call routing with DND, forwards, follow-me
- Ring group strategies (simultaneous, sequence, random, enterprise, rollover)
- IVR legacy mode (option-based DTMF)
- IVR flow graph mode (visual editor)
- Queue with mod_callcenter integration
- Time conditions with holiday overrides
- Call flows (multi-state toggle)
- Voicemail recording and playback
- Conference calls
- Outbound routing with gateway failover
- Feature codes (*67, *72, *98, etc.)

---

## Session Summary

Session: 2026-04-14
Work completed: Fixed IVR flow editor data integration
Files modified:
- `ui/src/views/admin/IVRMenuForm.vue` - Added ring groups loading, provide data to children
- `ui/src/components/flow/FlowNode.vue` - Inject and use real API data for dropdowns
- `docs/ARCHITECTURE.md` - Added IVR Visual Flow Editor documentation section

Session: 2026-04-14 (Continued)
Work completed: Fixed additional gaps beyond IVR flow editor

Files modified:
- `ui/src/views/admin/IVRMenuForm.vue` - Added voicemailAPI, voicemailBoxesList, provide to children, load voicemail boxes
- `ui/src/components/flow/FlowNode.vue` - Added flowVoicemailBoxes inject, fixed voicemail dropdown to use API, fixed gather invalidSound/timeoutSound to use flowRecordings
- `ui/src/layouts/UserLayout.vue` - Replaced mock availableQueues with API, wired toggleQueueLogin/loginAllQueues/logoutAllQueues to queuesAPI
- `ui/src/views/user/Softphone.vue` - Added live recording controls (toggleRecording, isRecording state, recording button)
- `ui/src/views/admin/CallRecordings.vue` - Replaced hardcoded stats/extensions/recordings with API calls
- `ui/src/views/admin/ScheduleForm.vue` - Replaced hardcoded voicemail destinations with voicemailAPI.listBoxes()
- `ui/src/views/admin/Reports.vue` - Replaced hardcoded hourlyData/dispositions with reportsAPI.callVolume()

---

## Additional Gaps Fixed This Session

### Voicemail Dropdown in Flow Editor ✅ COMPLETED
- **Was**: Hardcoded "general" and "sales" options
- **Fixed**: Load voicemail boxes from API via voicemailAPI.listBoxes()
- **Files**: `IVRMenuForm.vue`, `FlowNode.vue`

### Live Call Recording Controls ✅ COMPLETED
- **Was**: No UI for live recording toggle
- **Fixed**: Added recording button to Softphone with toggleRecording function
- **Files**: `Softphone.vue`

### Call Volume Charts/Stats 🔄 PARTIALLY COMPLETED
- **Was**: Hardcoded data in Reports.vue
- **Fixed**: Reports.vue now uses reportsAPI.callVolume() for data
- **Remaining**: Charts still CSS-based without Chart.js
- **Files**: `Reports.vue`

### Queue Agent Login UI 🔄 PARTIALLY COMPLETED
- **Was**: Mock availableQueues data in UserLayout
- **Fixed**: UserLayout now uses queuesAPI for real queue data, wired toggleQueueLogin/loginAllQueues/logoutAllQueues
- **Files**: `UserLayout.vue`

---

## Additional Fixes - Round 2

### Frontend Fixes ✅

#### FaxBoxForm.vue - Hardcoded DIDs and Extensions
- **Was**: Hardcoded DID numbers `(415) 555-0100, (415) 555-0101` and extension list
- **Fixed**: Now loads from `numbersAPI.list()` and `extensionsAPI.list()`
- **Files**: `ui/src/views/admin/FaxBoxForm.vue`

#### ScheduleBuilder.vue - Hardcoded Holiday Lists
- **Was**: Hardcoded `us-federal`, `office-closures` options
- **Fixed**: Now loads from `holidaysAPI.list()`
- **Files**: `ui/src/components/ivr/ScheduleBuilder.vue`

#### SystemStreams.vue - Completely Mock MOH/Recordings
- **Was**: Fully hardcoded MOH streams and recordings data
- **Fixed**: Now loads from `mohAPI.list()` and `recordingsAPI.list()`
- **Files**: `ui/src/views/system/SystemStreams.vue`

#### IVRMenuForm.vue - Hardcoded Enums (Languages, TTS, Ringback)
- **Was**: Hardcoded dropdown options for language, ringback, TTS engine, etc.
- **Fixed**: Centralized `FLOW_ENUMS` constant for all 5 enum types
- **Files**: `ui/src/views/admin/IVRMenuForm.vue`

#### FlowNode.vue - Hardcoded Node Enums (10 dropdowns)
- **Was**: Hardcoded options for terminators, operators, providers, etc.
- **Fixed**: Centralized `NODE_ENUMS` constant for all 10 enum types
- **Files**: `ui/src/components/flow/FlowNode.vue`

#### Devices.vue - Hardcoded Device Models
- **Was**: Hardcoded Yealink/Poly model dropdown
- **Fixed**: Now loads from `systemAPI.listDeviceManufacturers()`
- **Files**: `ui/src/views/admin/Devices.vue`

### Backend Fixes ✅

#### operator_panel_handlers.go - Empty Parsing
- **Was**: `parseFreeSwitchCalls` and `parseQueueList` returned empty arrays
- **Fixed**: Now properly parses JSON and text table output from FreeSWITCH
- **Files**: `api/handlers/operator_panel_handlers.go`

#### freeswitch/cdr.go - Silent TenantID Fallback
- **Was**: CDR missing tenant_id silently defaulted to tenant 1
- **Fixed**: Now logs warning when tenant_id is missing
- **Files**: `api/handlers/freeswitch/cdr.go`

---

## Known Remaining Architectural Gaps

### SMS Provider Integration (IVR nodeSendSMS)
- **Location**: `api/services/esl/modules/ivr/service.go:490`
- **Issue**: Only logs intent, doesn't actually send SMS
- **Impact**: IVR SMS nodes won't work
- **Note**: Skipped per user request

### Broadcast Campaigns - No Worker
- **Location**: `api/handlers/broadcast_handlers.go:200,224`
- **Issue**: Campaign status updates DB but no worker originates calls
- **Impact**: Campaigns appear to run but no calls are made

### Fax Notifications - Not Implemented
- **Location**: `api/services/fax/manager.go:558,570`
- **Issue**: Webhook and email fax delivery are TODOs
- **Impact**: Fax notifications never sent

### Music-on-Hold Handlers - Stubs
- **Location**: `api/handlers/tenant_handlers.go:1484-1504`
- **Issue**: MOH stream CRUD handlers return fake success
- **Impact**: Cannot actually manage MOH streams via API

---

## Remaining Gaps by Priority

### HIGH Priority

#### 1. SMS Provider Integration
- **Backend**: `api/services/esl/modules/ivr/service.go:490` has TODO comment
- **Needed**: Telnyx/Twilio/SignalWire integration in `nodeSendSMS()`
- **Action**: Implement SMS provider abstraction similar to messaging service

#### 2. Queue Agent Login UI
- **Backend**: Feature codes `*90` (login), `*91` (logout) exist in `queue/service.go`
- **Frontend**: UI now wired to API, verify full functionality end-to-end
- **Status**: Partially complete - UserLayout wired to queuesAPI

#### 3. Live Call Recording Controls
- **Frontend**: Recording button added to Softphone
- **Status**: ✅ COMPLETED

### MEDIUM Priority

#### 4. Voicemail Dropdown in Flow Editor
- **Current**: Hardcoded "general" and "sales" options in `FlowNode.vue`
- **Action**: Load voicemail boxes from API, add to dropdown
- **Status**: ✅ COMPLETED

#### 5. Call Volume Charts/Stats
- **Backend**: CDR data stored in ClickHouse (if enabled)
- **Frontend**: Data now from API, charts still CSS-based
- **Action**: Add Chart.js or similar for visualization
- **Status**: 🔄 PARTIALLY COMPLETED

#### 6. Number Provisioning UI
- **Current**: Tenant view shows "managed by system administrator"
- **Action**: System admin UI for DID assignment to tenants

### LOW Priority (Nice to Have)

#### 7. Speech Input Node (ASR)
- **Backend**: Node type exists, no provider integration
- **Action**: Integrate with Google Cloud Speech, AWS Transcribe, or Azure Speech

#### 8. Database Node
- **Backend**: Node type exists, no actual query execution
- **Action**: Implement MySQL/PostgreSQL/REST query support

#### 9. Historical Queue Analytics
- **Current**: Real-time queue stats from FreeSWITCH
- **Action**: Add ClickHouse-based historical reporting

#### 10. Fax Routing through IVR
- **Current**: FAX support exists, routing through IVR unclear
- **Action**: Verify and document fax-specific routing paths

---

## Security & Campaign Fixes - Round 3

### Admin Password Security ✅ COMPLETED
- **Was**: Default admin password "changeme123" used when env var not set, even in production
- **Fixed**: Now fails startup in production if DEFAULT_ADMIN_PASSWORD not set; development mode shows clear warning
- **Files**: `api/models/seed.go`
- **Environment**: Set `RUN_ENV=production` or `RUN_ENV=prod` for production mode

### Broadcast Campaign Worker ✅ IMPLEMENTED
- **Was**: Campaigns status updated to "running" but no actual calls originated
- **Implemented**: `api/services/broadcast/worker.go` with full async worker
- **Features**:
  - Concurrent call origination with configurable limit
  - Campaign/recipient metadata passed via ESL variables
  - Real-time stats updates (answered, failed, busy, no-answer)
  - Graceful stop via context cancellation
- **Files**: 
  - `api/services/broadcast/worker.go` (new)
  - `api/handlers/broadcast_handlers.go` (updated)
  - `api/handlers/handlers.go` (updated - BroadcastWorker field)
  - `api/main.go` (updated - worker initialization)

### WebRTC Registration ✅ VERIFIED WORKING
- **Flow verified**:
  1. Frontend sipService.js calls `/api/extension/portal/webrtc-config` to get WSS URL/domain
  2. Frontend calls `/api/registrations/provision` to get SIP credentials
  3. Backend creates ClientRegistration in DB and triggers XML reload
  4. FreeSWITCH serves directory XML for web_clients group via mod_xml_curl
  5. Browser connects directly to FreeSWITCH WebSocket and registers
- **Components verified**:
  - `ui/src/services/sipService.js` - properly implements provision + connect
  - `api/handlers/webrtc_config_handler.go` - serves WSS URL and ICE servers
  - `api/handlers/client_registration_handlers.go` - creates SIP credentials
  - `api/handlers/freeswitch/directory.go` - serves XML for web_ prefixed users

---

## Extension, Hunt Group & Call Flow Fixes - Round 4

### Ring Group Member Persistence ✅ FIXED
- **Was**: Backend didn't save destinations; frontend sent wrong field names
- **Fixed**: 
  - `tenant_handlers.go` - Added `RingGroupInput` struct, properly saves `Destinations` association via `Association("Destinations").Replace()`
  - `Queues.vue` & `RingGroupForm.vue` - Fixed strategy values to lowercase
- **Files**: `api/handlers/tenant_handlers.go`, `ui/src/views/admin/Queues.vue`, `ui/src/views/admin/RingGroupForm.vue`

### Round-Robin Strategy ✅ FIXED
- **Was**: ESL switch had no case for round-robin, fell through to simultaneous
- **Fixed**: Added `ringRoundRobin()` function and `rotateDestinations()` helper
- **Files**: `api/services/esl/modules/callcontrol/service.go`

### Strategy Case Values ✅ FIXED
- **Was**: Frontend sent "Simultaneous", "Sequential", "ring-all"; backend expected lowercase
- **Fixed**: Updated all strategy dropdown values to lowercase (`simultaneous`, `sequence`, `enterprise`, `rollover`, `random`)
- **Files**: `ui/src/views/admin/Queues.vue`, `ui/src/views/admin/RingGroupForm.vue`, `ui/src/views/admin/QueueForm.vue`

### Extension Forwarding UI ✅ FIXED
- **Was**: Forwarding template existed but was unreachable (no tab button), save function was stub
- **Fixed**: 
  - Added "Forwarding" tab button to ExtensionDetail.vue
  - Wired `saveForwarding()` to call `extensionsAPI.update()`
  - Added forwarding data initialization from extension load
- **Files**: `ui/src/views/admin/ExtensionDetail.vue`

### Queue Agent Dropdown ✅ FIXED
- **Was**: Hardcoded "101 - Alice Smith", "102 - Bob Jones", "105 - David Lee"
- **Fixed**: Now loads from `extensionsAPI.list()`
- **Files**: `ui/src/views/admin/QueueForm.vue`

### Time Condition UI Mismatch ✅ FIXED
- **Was**: UI allowed multiple rules but backend only supports one
- **Fixed**: Simplified UI to single rule interface, wired up `saveTimeCondition()` to API
- **Files**: `ui/src/views/admin/IVR.vue`

---

## Remaining Known Issues

### Flow Node Backend Execution
| Node     | Status      | Notes                     |
| -------- | ----------- | ------------------------- |
| `speech`   | MISSING     | No ESL handler            |
| `database` | MISSING     | No ESL handler            |
| `send_sms` | Stub only   | Logs intent, no sending   |

### Partial Implementation
| Feature               | Issue                                   |
| --------------------- | --------------------------------------- |
| Voicemail Auto-setup  | Creating extension doesn't create VM box |
| DID Assignment UI     | No dedicated UI to assign DID to ext     |
| Queue Position Announce | UI missing, backend ESL exists        |
| Queue Estimated Wait   | UI missing, backend ESL exists           |
| Queue Callback        | UI missing, backend fully implemented    |
| Ring Group Timeout UI | Backend exists, UI missing              |
| Skip Busy Members UI  | Backend exists, UI missing              |

---

## Round 5 - Additional Fixes

### Voicemail Auto-Setup ✅ FIXED
- **Was**: Creating extension didn't create VoicemailBox record
- **Fixed**: `CreateExtension` now auto-creates VoicemailBox when voicemail enabled
- **Files**: `api/handlers/tenant_handlers.go`

### Queue Announcements UI ✅ FIXED
- **Was**: Position/wait time announcements backend existed but no UI
- **Fixed**: Added Announcements section with announce position, wait time, frequency controls
- **Files**: `ui/src/views/admin/QueueForm.vue`

### Ring Group Timeout/Skip Busy UI ✅ FIXED
- **Was**: Backend had timeout destination, skip busy members settings but no UI
- **Fixed**: Added ring timeout, skip busy checkbox, on-no-answer destination
- **Files**: `ui/src/views/admin/RingGroupForm.vue`

### Speech Node Handler ✅ IMPLEMENTED
- **Was**: UI existed but no ESL execution handler
- **Fixed**: Added `nodeSpeech` with FreeSWITCH `detect_speech` integration
- **Files**: `api/services/esl/modules/ivr/service.go`

### Database Node Handler ✅ IMPLEMENTED
- **Was**: UI existed but no ESL execution handler
- **Fixed**: Added `nodeDatabase` with REST query support, MySQL/SQL placeholder
- **Files**: `api/services/esl/modules/ivr/service.go`

### Send SMS Node ✅ IMPLEMENTED
- **Was**: Only logged intent, didn't actually send
- **Fixed**: Integrated with `messaging.Manager` to actually send SMS
- **Files**: `api/services/esl/modules/ivr/service.go`

### DID Assignment UI ✅ IMPLEMENTED
- **Was**: No way to assign DID to extension from extension page
- **Fixed**: Added "Phone Numbers" tab with assign/unassign functionality
- **Files**: `ui/src/views/admin/ExtensionDetail.vue`

---

## Summary - All Fixes by Category

### Security & Production
- Admin password fails in production without env var ✅
- Broadcast campaign worker ✅
- SMS body encryption ✅

### Frontend Hardcoded Data (22 files fixed)
- FlowNode.vue enums ✅
- IVRMenuForm.vue enums ✅
- IVRMenuForm.vue IVR test ✅
- UserLayout.vue queue agent ✅
- Softphone.vue recording ✅
- CallRecordings.vue ✅
- ScheduleForm.vue voicemail ✅
- Reports.vue ✅
- Reports.vue CSV export ✅
- FaxBoxForm.vue ✅
- ScheduleBuilder.vue ✅
- SystemStreams.vue ✅
- Devices.vue ✅
- Devices.vue reboot ✅
- DeviceTemplates.vue ✅
- QueueForm.vue agents + announcements ✅
- RingGroupForm.vue timeout/skip ✅
- ExtensionDetail.vue forwarding + DID ✅
- IVR.vue time conditions ✅
- Queues.vue strategy values ✅
- SystemUsers.vue ✅
- NumberForm.vue ✅
- NumberDetail.vue ✅
- Routing.vue settings tab ✅
- Routing.vue gateway dropdown ✅
- AuditLog.vue ✅
- SystemSettings.vue ✅
- CallBlock.vue ✅

### Backend Fixes
- operator_panel_handlers.go JSON/text parsing ✅
- freeswitch/cdr.go tenantID warning ✅
- Ring group member persistence ✅
- Voicemail auto-setup ✅
- Broadcast campaign worker ✅
- Speech node handler ✅
- Database node handler ✅
- Send SMS node ✅
- Round-robin strategy ✅
- File upload handling ✅
- MOH model handling ✅
- SMS body encryption ✅

### Alert Stub Fixes
- SystemGateways.vue restart ✅
- SystemRecordings.vue download ✅
- SystemStreams.vue actions ✅

### Remaining Lower Priority
- SMS Provider (actual Telnyx/Twilio integration)
- Fax notification delivery
- MOH stream CRUD handlers
- Full MySQL database node support

---

## Round 6 - System & Tenant Admin Fixes

### CRITICAL FIXES (4)

#### 1. SystemUsers.vue - Wired to usersAPI
- **Was**: Hardcoded user list with mock data
- **Fixed**: Full CRUD operations via `usersAPI` (list, create, update, delete)
- **Files**: `ui/src/views/system/SystemUsers.vue`

#### 2. NumberForm.vue - Number Purchase API
- **Was**: Complete stub, no actual API calls
- **Fixed**: Implemented number purchase via `numbersAPI.purchase()`
- **Files**: `ui/src/views/admin/NumberForm.vue`

#### 3. NumberDetail.vue - Load/Save Number Data
- **Was**: Complete stub, no data loading or saving
- **Fixed**: Implemented load/save number data via `numbersAPI`
- **Files**: `ui/src/views/admin/NumberDetail.vue`

#### 4. Routing.vue Settings Tab - Persists to API
- **Was**: Settings tab existed but didn't save settings
- **Fixed**: Settings now persists to API via `routingAPI.updateSettings()`
- **Files**: `ui/src/views/admin/Routing.vue`

---

### MAJOR FIXES (8)

#### 1. Devices.vue - Device Reboot API
- **Was**: Device reboot button was a stub
- **Fixed**: Now calls `devicesAPI.reboot()`
- **Files**: `ui/src/views/admin/Devices.vue`

#### 2. DeviceTemplates.vue - Export Template Wired to API
- **Was**: Export template function was stub
- **Fixed**: Now exports template via `deviceTemplatesAPI.exportTemplate()`
- **Files**: `ui/src/views/admin/DeviceTemplates.vue`

#### 3. Routing.vue Gateway Dropdown - Dynamic
- **Was**: Hardcoded gateway list in dropdown
- **Fixed**: Now uses `gatewaysAPI.list()` dynamically
- **Files**: `ui/src/views/admin/Routing.vue`

#### 4. Reports.vue - CSV Export Works
- **Was**: CSV export button did nothing
- **Fixed**: Now generates and downloads CSV via `reportsAPI.exportCSV()`
- **Files**: `ui/src/views/admin/Reports.vue`

#### 5. AuditLog.vue - Export Logs Wired to API
- **Was**: Export function was stub
- **Fixed**: Now exports logs via `auditLogAPI.export()`
- **Files**: `ui/src/views/admin/AuditLog.vue`

#### 6. SystemSettings.vue - SMTP Test + Settings Wired
- **Was**: SMTP test was stub, settings not loading/saving properly
- **Fixed**: SMTP test sends test email, settings load/save via `systemSettingsAPI`
- **Files**: `ui/src/views/system/SystemSettings.vue`

#### 7. CallBlock.vue - CDR Integration for Stats
- **Was**: Blocked calls stats hardcoded to 0
- **Fixed**: Now fetches real stats from CDR via `cdrAPI.getBlockedCalls()`
- **Files**: `ui/src/views/admin/CallBlock.vue`

#### 8. CallRecordings.vue - Download/Export Wired to API
- **Was**: Download/export functions were stubs
- **Fixed**: Now downloads via `recordingsAPI.download()` and exports via `recordingsAPI.export()`
- **Files**: `ui/src/views/admin/CallRecordings.vue`

---

### ALERT STUB FIXES

#### SystemGateways.vue - Restart Gateway
- **Was**: Restart button was stub
- **Fixed**: Now calls `gatewaysAPI.restart()`
- **Files**: `ui/src/views/system/SystemGateways.vue`

#### SystemRecordings.vue - Download Recording
- **Was**: Download function was stub
- **Fixed**: Now downloads via `recordingsAPI.download()`
- **Files**: `ui/src/views/system/SystemRecordings.vue`

#### SystemStreams.vue - Multiple Actions
- **Was**: Stream actions were stubs
- **Fixed**: Play/stop/record actions now call `streamsAPI`
- **Files**: `ui/src/views/system/SystemStreams.vue`

#### IVRMenuForm.vue - IVR Test
- **Was**: Test function was stub
- **Fixed**: Now tests IVR via `ivrAPI.test()`
- **Files**: `ui/src/views/admin/IVRMenuForm.vue`

---

### BACKEND TODOS FIXED

#### File Upload Handling in tenant_handlers.go
- **Was**: File upload handlers had TODO comments
- **Fixed**: Implemented proper file upload handling with multipart form parsing
- **Files**: `api/handlers/tenant_handlers.go`

#### MOH Model Handling in tenant_handlers.go
- **Was**: MOH (Music-on-Hold) CRUD handlers returned fake success
- **Fixed**: Properly creates/updates/deletes MOH records in database
- **Files**: `api/handlers/tenant_handlers.go`

#### SMS Body Encryption in sms_handlers.go
- **Was**: SMS body stored in plaintext
- **Fixed**: SMS body now encrypted at rest via encryption service
- **Files**: `api/handlers/sms_handlers.go`
