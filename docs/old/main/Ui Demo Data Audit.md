# UI Demo Data Audit

Full audit of all UI views & services to identify non-functional / demo-data portions.

---

## Severity Legend

| Level | Meaning |
|---|---|
| 🔴 **Fully Static** | 100% hardcoded HTML/JS — no API calls at all |
| 🟠 **Fallback Demo** | Tries API, falls back to hardcoded demo data on failure/empty |
| 🟡 **Mock Logic** | Has form/interaction logic but uses `alert()` / `console.log()` instead of API |
| 🔵 **TODO Gaps** | Wired to API but has specific missing features flagged with `// TODO` |

---

## 1. User Portal (`/dialer`, `/messages`, `/voicemail`, etc.)

### 🔴 Softphone — [Softphone.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/Softphone.vue)
- SIP connection is **fully mocked** — simulates registration with `setTimeout`
- Call flow (dial → ringing → connected → hangup) is **mock timers**, no actual SIP.js or WebRTC
- Uses mock data from [sipService.js](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/services/sipService.js) which has `// For now, mock data` and `// Mock successful connection`

### 🔴 Call History — [History.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/History.vue)
- 8 hardcoded call records (`Alice Smith`, `Bob Jones`, etc.)
- `makeCall()`, `sendMessage()`, `addToContacts()` all use `alert()`

### 🔴 Contacts — [Contacts.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/Contacts.vue)
- Company Directory (5 hardcoded contacts: `Alice Smith`, `Bob Jones`, etc.)
- Speed Dials (2 hardcoded groups: `Executive Directory`, `Vendor Support`)
- System Resources (3 hardcoded: `Main IVR`, `Sales Queue`, `Support Queue`)
- Call/Message buttons use `alert()`

### 🔴 My Recordings — [UserRecordings.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/UserRecordings.vue)
- 5 hardcoded recording entries
- `sampleTranscript` array with 4 hardcoded transcript lines
- Download, delete, play all mock — `alert()` or local state only
- Audio player waveform is random `Math.random()` bars

### 🟡 User Fax — [UserFax.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/UserFax.vue)
- Document processing labeled `// Document Processing (Mock)`
- Preview marked `// Mock preview ready`
- Fax list may attempt API, but core send/receive logic is mock

### 🟡 User Conferences — [UserConferences.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/UserConferences.vue)
- Needs verification — may pull real data but live controls (mute/kick) likely mock

### ✅ Voicemail — [Voicemail.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/Voicemail.vue)
- Appears to use API for data

### ✅ User Settings — [UserSettings.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/user/UserSettings.vue)
- Appears wired to API

---

## 2. Tenant Admin (`/admin/...`)

### 🔴 Overview Dashboard — [Overview.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Overview.vue)
- All 4 stat cards hardcoded (`Active Calls: 24`, `Registrations: 856`, `Failed Calls: 3`, `System Health: Operational`)
- Recent Alerts section (2 entries) hardcoded in HTML
- Live Calls table (2 rows) hardcoded in HTML
- **No API calls**

### 🔴 Reports & Analytics — [Reports.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Reports.vue)
- All KPI values hardcoded (`Total Calls: 1,248`, `Avg Handle Time: 4m 12s`, `Missed: 24`, `SLA Breached: 2`)
- Bar chart is CSS mockup with hardcoded heights
- Donut chart is hardcoded `conic-gradient` (85%/10%/5%)
- Export CSV button non-functional
- **No script logic at all** (just icon imports)

### 🔴 Messaging / Chat — [Messaging.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Messaging.vue)
- 3 hardcoded conversations in sidebar (`John Smith`, `(555) 123-4567`, `Support Ticket #99`)
- 3 hardcoded message bubbles in chat area
- Send button, search, attachments all non-functional
- **No script logic** (just icon imports)

### 🔴 Operator Panel — [OperatorPanel.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/OperatorPanel.vue)
- 5 hardcoded extensions (`Alice Smith`, `Bob Jones`, etc.)
- "Live" indicator is cosmetic animation only
- No WebSocket/ESL connection for real-time status

### 🔴 Conference Console — [ConferenceConsole.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/ConferenceConsole.vue)
- 4 hardcoded participants (`Alice Smith`, `Bob Jones`, `External Caller`, `Dave Wilson`)
- Conference name/timer hardcoded (`Weekly Sales #3001`, `00:12:43`)
- Mute toggle works locally but doesn't send ESL commands
- Lock Room, Mute All, Stop Rec buttons non-functional

### 🔴 Call Flows — [CallFlows.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/CallFlows.vue)
- 4 hardcoded phone numbers with routing topology
- `flowGroups` has hardcoded node positions and connections
- Visual diagram is beautiful but entirely static sample data

### 🔴 Hospitality & PMS — [Hospitality.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Hospitality.vue)
All 4 tabs static:
- **Room Status**: 7 hardcoded rooms, stats hardcoded (`Total: 124`, `Clean: 85`, `Dirty: 32`, `Insp: 7`)
- **PMS Configuration**: Form fields with no save logic
- **Service Codes**: Hardcoded values, save button non-functional
- **Wake Up Calls**: 1 hardcoded wake-up entry, schedule uses `prompt()` / `alert()`

### 🔴 Location Manager — [LocationManager.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/LocationManager.vue)
- 2 hardcoded locations (`HQ - San Francisco`, `Warehouse - Nevada`)
- Edit/Remove/Add buttons non-functional

### 🔴 Call Broadcast — [CallBroadcast.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/CallBroadcast.vue)
- 3 hardcoded campaigns
- Delete button uses `alert()`

### 🔴 Bridges List — [Bridges.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Bridges.vue)
- 3 hardcoded bridge extensions
- Delete uses `confirm()` + `alert()`

### 🟠 CDR (Call Detail Records) — [CDR.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/CDR.vue)
- Attempts API call, falls back to demo data on failure

### 🟠 Queues — [Queues.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Queues.vue)
- Attempts API, falls back to demo data **twice** (queues + ring groups)

### 🟠 IVR — [IVR.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/IVR.vue)
- Attempts API, falls back to demo data
- Save form has `// TODO: Map form to API format and save`

### 🟠 Devices — [Devices.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Devices.vue)
- Attempts API, falls back to demo data
- Reboot device has `// TODO: Call API to reboot device`

### 🟠 VoicemailBoxes — [VoicemailBoxes.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/VoicemailBoxes.vue)
- Attempts API, falls back to demo data

### 🟠 Conferences — [Conferences.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/Conferences.vue)
- Attempts API, falls back to demo data

### 🟠 Time Conditions — [TimeConditions.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/TimeConditions.vue)
- Attempts API, falls back to demo data

### 🟠 Audit Log — [AuditLog.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/AuditLog.vue)
- Contains explicit `demoLogs` array with 5+ entries
- Falls back to demo data on error OR empty response

### 🟡 Bridge Form — [BridgeForm.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/BridgeForm.vue)
- Edit mode uses `// Mock Load` with hardcoded values
- Save uses `alert('Bridge saved successfully.')`

### 🟡 Ring Group Form — [RingGroupForm.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/RingGroupForm.vue)
- Edit mode uses `// Mock load` with hardcoded values
- Save uses `alert('Ring Group Saved')`

### 🟡 Queue Form — [QueueForm.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/QueueForm.vue)
- `// Mock hydration` for edit mode

### 🟡 Recording Form — [RecordingForm.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/RecordingForm.vue)
- `// Mock upload logic` — just `console.log()`

### 🟡 Call Broadcast Form — [CallBroadcastForm.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/CallBroadcastForm.vue)
- Likely mock save logic

### 🔵 Music On Hold — [MusicOnHold.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/MusicOnHold.vue)
- `// TODO: API call to create folder` (×2)

### 🔵 Call Block — [CallBlock.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/CallBlock.vue)
- `totalHits = computed(() => 0) // CDR integration TODO`

### 🔵 Device Templates — [DeviceTemplates.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/DeviceTemplates.vue)
- `// TODO: Implement template export endpoint`

### 🔵 Global Dial Plans — [GlobalDialPlans.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/GlobalDialPlans.vue)
- `// TODO: Update order in backend`

### 🔵 Tenant Provisioning — [TenantProvisioning.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/admin/TenantProvisioning.vue)
- `// TODO: Call API to generate tenant secret`
- `// TODO: Call API to save syslog settings`

---

## 3. System Admin (`/system/...`)

### 🔵 System Routes — [SystemRoutes.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/system/SystemRoutes.vue)
- `// TODO: Save new order to backend`
- Outbound save has `// ... placeholder for outbound save logic ...`

### 🔵 System Media — [SystemMedia.vue](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/views/system/SystemMedia.vue)
- `<!-- Modals (Placeholders for now) -->`

---

## 4. Services

### 🔴 SIP Service — [sipService.js](file:///Users/shaunagostinho/Antigravity/callsign/ui/src/services/sipService.js)
- `// For now, mock data` — presence/BLF is simulated
- `// Mock successful connection for now` — no real SIP.js integration
- `// Mock call progress` — call state machine is timers only

---

## Summary Table

| Section | Total Views | 🔴 Fully Static | 🟠 Fallback | 🟡 Mock Logic | 🔵 TODO | ✅ Wired |
|---|---|---|---|---|---|---|
| **User Portal** | 9 | 4 | 0 | 2 | 0 | 3 |
| **Tenant Admin** | 64 | 11 | 7 | 4 | 5 | ~37 |
| **System Admin** | 26 | 0 | 0 | 0 | 2 | ~24 |
| **Services** | 4 | 1 | 0 | 0 | 0 | 3 |
| **Total** | **103** | **16** | **7** | **6** | **7** | **~67** |

> [!WARNING]
> Approximately **36 files** (35% of UI) have some level of demo data, mock logic, or incomplete API integration. The most critical are the 16 fully static views that have zero API connectivity.
