# CallSign PBX — Frontend Documentation

## Overview

The CallSign frontend is a **Vue 3 Single-Page Application** built with Vite. It uses vanilla CSS (no Tailwind or component library), Axios for HTTP, and JsSIP for WebRTC. The UI provides three distinct portals with role-based routing.

---

## Project Structure

```
ui/
├── index.html            # HTML shell
├── vite.config.js        # Vite configuration (proxy /api → localhost:8080)
├── package.json          # Dependencies: vue, vue-router, axios, jssip
└── src/
    ├── main.js            # Vue app bootstrap
    ├── App.vue            # Root component (<router-view>)
    ├── router.js          # Route definitions with auth guard
    ├── style.css          # Global base styles
    ├── styles/            # Additional CSS modules
    ├── services/
    │   ├── api.js          # Axios client (30+ API modules, ~870 lines)
    │   ├── auth.js         # Auth state management (login/logout/token)
    │   ├── notifications.js # WebSocket notification consumer
    │   └── sipService.js   # WebRTC SIP client (JsSIP)
    ├── components/
    │   ├── layout/
    │   │   └── LayoutShell.vue   # Admin/system portal layout (sidebar + topbar)
    │   ├── common/              # Reusable UI components
    │   ├── features/            # Feature-specific components
    │   ├── flow/                # Visual call flow builder
    │   └── ivr/                 # IVR menu visual editor
    ├── layouts/
    │   └── UserLayout.vue       # User portal layout (softphone-focused)
    ├── utils/                   # Utility functions
    ├── assets/                  # Static assets (images, icons)
    └── views/
        ├── auth/                # 2 views
        ├── admin/               # 64 views
        ├── system/              # 26 views
        └── user/                # 8 views
```

---

## Portal Architecture

### User Portal (`/`)

Layout: `UserLayout.vue` — A softphone-centric layout with a sidebar navigation and integrated WebRTC dialer.

| Route | View | Description |
|---|---|---|
| `/dialer` (default) | `Softphone.vue` | WebRTC softphone with call controls |
| `/messages` | `Messaging.vue` | SMS/MMS conversations |
| `/voicemail` | `Voicemail.vue` | Voicemail inbox and playback |
| `/conferences` | `UserConferences.vue` | Personal conference rooms |
| `/fax` | `UserFax.vue` | Fax inbox/outbox with send |
| `/contacts` | `Contacts.vue` | Personal contact directory |
| `/recordings` | `UserRecordings.vue` | Call recordings |
| `/history` | `History.vue` | Call history / CDR |
| `/settings` | `UserSettings.vue` | Extension and user settings |

### Tenant Admin Portal (`/admin`)

Layout: `LayoutShell.vue` — Full admin layout with sidebar navigation, topbar with tenant context, and notification center.

The admin portal provides management for a single tenant's telephony configuration. Key views organized by function:

**Call Management**: Extensions, Extension Profiles, Ring Groups, Queues, Speed Dials, Paging Groups, Call Flows/Toggles, Call Block, Call Broadcast

**Routing & Dialplan**: Routing (inbound + outbound), Dial Plans, Feature Codes, Routing Debugger

**Conferencing**: Conference Rooms, Live Conference Console

**IVR & Scheduling**: IVR Menu Builder, Time Conditions, Holiday Lists

**Media**: Audio Library, Music on Hold, Call Recordings

**Voicemail & Fax**: Voicemail Box Manager, Fax Server (boxes, endpoints, jobs)

**Devices**: Device Manager, Device Profiles, Device Templates, Provisioning

**Communication**: Messaging (SMS/MMS), Contacts

**Hospitality**: Room Management, Wake-Up Calls

**Reports & Audit**: CDR Viewer, Reports/Analytics, Audit Log

**Settings**: Tenant Settings (general, branding, SMTP, messaging, hospitality), E911 Locations

### System Admin Portal (`/system`)

Layout: `LayoutShell.vue` — Same shell as admin but with system-level navigation.

| Route | View | Description |
|---|---|---|
| `/system` | `Admin.vue` | System dashboard (status, stats) |
| `/system/tenants[/new\|:id]` | `Tenants.vue`, `TenantForm.vue` | Tenant management |
| `/system/profiles` | `TenantProfiles.vue` | Tenant feature profiles |
| `/system/trunks` | `SystemGateways.vue` | SIP trunk management |
| `/system/sip-profiles` | `SipProfiles.vue` | SIP profile management with sofia control |
| `/system/acls` | `ACLProfiles.vue` | Access control lists |
| `/system/routing` | `SystemRoutes.vue` | System-level routing (numbers, groups) |
| `/system/sounds` | `SystemSounds.vue` | System sounds, music, phrases |
| `/system/firmware` | `FirmwareUpdates.vue` | Firmware management |
| `/system/provisioning-templates` | `ProvisioningTemplates.vue` | Device provisioning templates |
| `/system/messaging` | `MessagingProviders.vue` | SMS/MMS provider management |
| `/system/logs` | `SystemLogs.vue` | FreeSWITCH console & API logs |
| `/system/security` | `SystemSecurity.vue` | Banned IP management |
| `/system/settings` | `SystemSettings.vue` | Global system settings |
| `/system/config-inspector` | `ConfigInspector.vue` | FreeSWITCH config file browser |

---

## Authentication Flow

### Router Guard (`router.js`)

The Vue Router has a `beforeEach` guard that enforces authentication:

1. **Public routes** (`Login`, `AdminLogin`): Redirects authenticated users to their portal based on role.
2. **Protected routes**: Redirects unauthenticated users to the appropriate login page.
3. **Role enforcement**:
   - `system_admin` is blocked from user portal routes and redirected to `/system`
   - `system_admin` accessing `/admin` requires a tenant selected (stored in `localStorage.tenantId`)
   - Non-admins are blocked from `/system` routes
   - Non-admin/non-system-admin users are blocked from `/admin` routes

### Auth State

Authentication state is stored in `localStorage`:
- `token` — JWT access token
- `refreshToken` — JWT refresh token
- `user` — Serialized user object (`{ role, username, ... }`)
- `tenantId` — Selected tenant ID (for system admins operating on a specific tenant)

---

## API Client (`services/api.js`)

The API client is an Axios instance with the following features:

### Request Interceptor
- Adds `Authorization: Bearer <token>` header from localStorage
- Adds `X-Tenant-ID` header for system admins scoping to a tenant

### Response Interceptor
- **Auto-unwrap**: Backend responses wrap data in `{ data: [...] }`. The interceptor extracts `response.data = body.data` so views receive the payload directly.
- **Null guard**: Go nil slices serialize as JSON `null`. The interceptor converts `null`/`undefined` to `[]`.
- **Meta preservation**: Sibling fields (e.g., `message`, `box`, `interval`) are stored in `response._meta`.
- **Token refresh**: 401 responses trigger a token refresh attempt; on failure, redirects to login.

### API Modules

The file exports 30+ named API modules. Each module is an object with methods like `list()`, `get(id)`, `create(data)`, `update(id, data)`, `delete(id)`:

| Module | Prefix | Description |
|---|---|---|
| `authAPI` | `/auth` | Login, logout, profile, password |
| `extensionsAPI` | `/extensions` | Extension CRUD + call rules |
| `extensionProfilesAPI` | `/extension-profiles` | Extension profiles + call rules |
| `devicesAPI` | `/devices` | Devices, lines, call control |
| `deviceProfilesAPI` | `/device-profiles` | Device profiles |
| `deviceTemplatesAPI` | `/device-templates` | Device templates |
| `queuesAPI` | `/queues` | Queues + agent management |
| `conferencesAPI` | `/conferences` | Conferences + live control |
| `ivrAPI` | `/ivr/menus` | IVR menus |
| `featureCodesAPI` | `/feature-codes` | Feature codes + modules |
| `timeConditionsAPI` | `/time-conditions` | Time conditions |
| `holidaysAPI` | `/holidays` | Holiday lists |
| `togglesAPI` / `callFlowsAPI` | `/call-flows` | Call flows |
| `voicemailAPI` | `/voicemail` | Voicemail boxes + messages |
| `cdrAPI` | `/cdr` | CDR records |
| `messagingAPI` | `/messaging` | SMS/MMS conversations |
| `faxAPI` | `/fax` | Fax boxes, jobs, endpoints |
| `contactsAPI` | `/contacts` | Contacts |
| `pagingAPI` | `/page-groups` | Paging groups |
| `ringGroupsAPI` | `/ring-groups` | Ring groups |
| `numbersAPI` | `/numbers` | DIDs/numbers |
| `routingAPI` | `/routing` | Inbound/outbound routes + blocks |
| `dialPlansAPI` | `/dial-plans` | Dial plans |
| `audioLibraryAPI` | `/audio-library` | Audio files |
| `mohAPI` | `/music-on-hold` | Music on hold |
| `recordingsAPI` | `/recordings` | Recordings + transcription |
| `provisioningAPI` | `/provisioning-templates` | Provisioning templates |
| `tenantSettingsAPI` | `/tenant` | Tenant settings (all sub-resources) |
| `systemAPI` | `/system` | All system admin operations |
| `usersAPI` | `/users` | Tenant user management |
| `tenantMediaAPI` | `/media` | Tenant sound/music overrides |
| `userPortalAPI` | `/user` | User-scoped portal data |
| `extensionPortalAPI` | `/extension/portal` | Extension-scoped portal data |
| `reportsAPI` | `/reports` | Reports & analytics |
| `liveOpsAPI` | `/live` | Live call/queue/recording ops |

---

## Real-Time Services

### Notification Service (`services/notifications.js`)

Connects to `/api/ws/notifications` via WebSocket. Handles:
- Call event notifications (ringing, answered, ended)
- Voicemail MWI (Message Waiting Indicator) updates
- Conference join/leave events
- Queue agent events
- Chat/messaging notifications

Authentication is handled within the WebSocket by sending the JWT token as the first message.

### SIP Service (`services/sipService.js`)

WebRTC softphone client built on JsSIP. Used by the user portal's Softphone view for:
- SIP registration via WebSocket transport
- Making and receiving calls
- DTMF tone generation
- Call hold, transfer, mute
- Audio device management
- Call event callbacks for UI state updates

---

## Key Component Patterns

### View Pattern
Most admin/system views follow a consistent pattern:
1. `onMounted()` → Fetch data via API module
2. Table/list display with search/filter
3. Modal dialogs for create/edit forms
4. Delete confirmation dialogs
5. Toast notifications for success/error feedback

### Layout Shell
`LayoutShell.vue` provides:
- Collapsible sidebar navigation (contextual to admin vs system portal)
- Top navigation bar with user profile dropdown
- Notification center (WebSocket-driven)
- Content area with `<router-view>` slot
- CSS-based responsive layout

### Form Components
Dedicated form views (e.g., `IVRMenuForm.vue`, `QueueForm.vue`, `ExtensionDetail.vue`) handle:
- Both create and edit modes (detected by route params)
- Validation
- Related sub-resource management (e.g., queue agents, IVR options)
- Navigation guards for unsaved changes
