# CallSign PBX — Setup, Usage & Feature Guide

## Table of Contents

1. [Initial Setup](#initial-setup)
2. [System Administration](#system-administration)
3. [Tenant Configuration](#tenant-configuration)
4. [Extensions & Users](#extensions--users)
5. [Call Routing](#call-routing)
6. [Feature Modules](#feature-modules)
7. [Device Management](#device-management)
8. [Communication Services](#communication-services)
9. [Monitoring & Reporting](#monitoring--reporting)

---

## Initial Setup

### Prerequisites

The following must be installed on the host server:

- **FreeSWITCH** (installed directly on the host, not in Docker)
- **Docker** and **Docker Compose** (for all other services)
- FreeSWITCH `mod_xml_curl`, `mod_event_socket`, `mod_xml_cdr`, `mod_callcenter`, `mod_conference` modules loaded

### First Boot Process

1. **Start services**: `docker compose up -d`
2. The API container automatically:
   - Connects to PostgreSQL and runs all database migrations
   - Seeds a **default system admin** (`admin` / `changeme123` unless overridden via `DEFAULT_ADMIN_USERNAME`, `DEFAULT_ADMIN_PASSWORD`, `DEFAULT_ADMIN_EMAIL`)
   - Seeds **3 default tenant profiles** (Starter, Professional, Enterprise)
   - Seeds **default outbound routes** (Local/Toll-Free, Long Distance, International, Emergency)
   - Seeds **default system sounds** (silence, ring, beep, IVR greeting, voicemail greeting)
   - Seeds **default chatplans** for SMS routing
   - Imports existing **SIP profiles** from `/etc/freeswitch/sip_profiles/` into the database
   - Connects to FreeSWITCH via ESL (:8021) and starts all call handling modules

3. **FreeSWITCH configuration**: Point `mod_xml_curl` to `http://127.0.0.1:8080/api/freeswitch/xmlapi` for `directory|dialplan|configuration` bindings. Point `mod_xml_cdr` to `http://127.0.0.1:8080/api/freeswitch/cdr`.

4. **Access the UI**: Navigate to the server's hostname. Log in at `/admin/login` with the default admin credentials.

### Environment Variables

Key variables that must be set for production:

| Variable | Required | Description |
|---|---|---|
| `JWT_SECRET` | **Yes** | Secret for JWT signing — must not be default |
| `ENCRYPTION_KEY` | **Yes** | AES key for data-at-rest encryption |
| `ENCRYPTION_SALT` | **Yes** | Salt for key derivation |
| `POSTGRES_PASSWORD` | **Yes** | Database password |
| `FREESWITCH_ESL_PASSWORD` | **Yes** | ESL authentication password (must match FreeSWITCH) |
| `FREESWITCH_API_KEY` | Recommended | API key for XML CURL authentication |
| `DEFAULT_ADMIN_PASSWORD` | Recommended | Override default admin password |

See `docs/ARCHITECTURE.md` for the full configuration reference.

---

## System Administration

System administrators have global access. After first login, the recommended setup order is:

### 1. SIP Profiles

SIP profiles define how FreeSWITCH listens for SIP traffic. Two are typically needed:

| Profile | Port | Purpose | Auth |
|---|---|---|---|
| **Internal** | 5060 | Extension registrations, WebRTC | Required (SIP digest) |
| **External** | 5080 | Trunk connections (inbound from carriers) | Not required |

Profiles are imported from disk on first boot and then managed via the UI at **System → SIP Profiles**. Settings include codec preferences, NAT handling, TLS, RTP ranges, and more. Changes require a profile rescan or restart to take effect.

### 2. SIP Trunks (Gateways)

Trunks connect CallSign to PSTN carriers. Configure at **System → Trunks**.

Each gateway defines:
- **Connection**: Proxy address, transport (UDP/TCP/TLS), registration
- **Authentication**: Username, password, realm
- **Caller ID**: From-user, from-domain, caller-ID-in-FROM
- **Dial Format**: E.164, 10-digit, 11-digit, or custom number formatting
- **Routing**: Priority/weight for LCR (Least Cost Routing), route tags

After creating a trunk, run **Sofia Rescan** to activate it without restarting the profile.

### 3. System Numbers (DID Pool)

System numbers are the central pool of phone numbers (DIDs). Managed at **System → Routing → Numbers**.

- System admin adds numbers in E.164 format (e.g., `+14155551234`)
- Numbers have capability flags: SMS, MMS, Fax, E911
- Numbers are assigned to **tenants** — tenants cannot add their own numbers
- Numbers can be grouped into **Number Groups** for outbound routing

**Number Groups** define which trunk(s) to use for outbound calls on numbers in that group, with priority/weight for failover and load balancing.

### 4. Tenant Profiles

Tenant profiles define resource limits and feature access. Three are seeded by default:

| Profile | Extensions | Queues | Conferences | Recording | Fax | SMS | WebRTC |
|---|---|---|---|---|---|---|---|
| **Starter** | 10 | 2 | 2 | ❌ | ❌ | ❌ | ✅ |
| **Professional** | 50 | 10 | 10 | ✅ | ✅ | ❌ | ✅ |
| **Enterprise** | Unlimited | Unlimited | Unlimited | ✅ | ✅ | ✅ | ✅ |

Custom profiles can be created. Limits are enforced at the API level — attempts to exceed a limit return an error.

Limits available: extensions, devices, queues, conferences, ring groups, IVR menus, voicemail boxes, fax servers, users. Feature flags: recording, fax, SMS, WebRTC, conferencing, call broadcast.

### 5. Access Control Lists (ACLs)

ACLs restrict SIP traffic by IP address. Managed at **System → ACLs**. Each ACL has nodes (CIDR rules) that allow or deny traffic. Used for:
- Restricting which IPs can register
- Trusted carrier IPs (bypass auth)
- Blocking known bad actors

### 6. Additional System Configuration

- **Messaging Providers**: Configure SMS/MMS gateways (Telnyx integration). System → Messaging.
- **Messaging Numbers**: Assign phone numbers to messaging providers. System → Messaging → Numbers.
- **Global Dial Plans**: System-wide dialplan overrides. System → Routing → Dial Plans.
- **Sounds & Media**: Upload system-wide sound files, music on hold. System → Sounds.
- **Firmware**: Manage device firmware for auto-provisioning. System → Firmware.
- **Device Templates**: Create system-level provisioning templates. System → Provisioning Templates.
- **Config Inspector**: Browse FreeSWITCH configuration files. System → Config Inspector.
- **Security**: View and manage banned IPs (from fail2ban integration). System → Security.

---

## Tenant Configuration

### Creating a Tenant

From **System → Tenants → Add Tenant**, provide:

| Field | Required | Description |
|---|---|---|
| **Name** | Yes | Organization/company name |
| **Domain** | Yes | SIP domain (e.g., `acme.callsign.io`) — must be unique |
| **Profile** | Recommended | Assign a tenant profile for resource limits |
| **Enabled** | Yes | Enable/disable the tenant |

After creation, assign system numbers to the tenant and create a tenant admin user.

### Creating a Tenant Admin

Tenant admins are `User` records with role `tenant_admin` and a `tenant_id`. Create one from the system admin panel or bulk-provision. The admin can then log in at `/admin/login` and manage the tenant.

### Tenant Settings

Within the admin portal, tenant-level settings include:

- **General Settings**: Timezone, country code, area code, caller ID defaults
- **Branding**: Whitelabel logo, name, primary color (for custom-branded portals)
- **SMTP**: Per-tenant email settings for voicemail-to-email and notifications
- **Messaging**: SMS/MMS configuration (number assignment, messaging provider)
- **Hospitality**: Enable hotel/hospitality mode (room management, wake-up calls)
- **E911 Locations**: Physical locations linked to system numbers for emergency caller ID

### Provisioning

Each tenant has a **provisioning secret** for secure phone auto-provisioning. Phones fetch their configuration via:
```
https://{server}/api/provision/{tenant_uuid}/{provisioning_secret}/{mac}.cfg
```

Provisioning templates define the XML/config format for different phone models (Yealink, Polycom, Grandstream, etc.).

---

## Extensions & Users

### Extension Model

An extension represents a phone line (SIP endpoint). Each extension lives within a single tenant.

**Creating an Extension** (Admin Portal → Extensions → Add):

| Field | Description | Default |
|---|---|---|
| **Extension** | Number (e.g., `1001`) | Required |
| **SIP Password** | Used by physical phones for SIP registration | Required |
| **Web Password** | Used for user portal / app login (bcrypt-hashed, separate from SIP) | Optional |
| **Effective Caller ID** | Internal display name/number | — |
| **Outbound Caller ID** | PSTN-facing display name/number | — |
| **Emergency Caller ID** | Used for 911 calls | — |

**Call Settings**:

| Setting | Description | Default |
|---|---|---|
| Call Timeout | Seconds to ring before no-answer action | 30 |
| Toll Allow | Allowed call types: `domestic`, `local`, `emergency`, `international` | — |
| Ring Strategy | `simultaneous` (all devices at once) or `sequential` (ordered) | simultaneous |
| No-Answer Action | `voicemail`, `forward`, `hangup`, `queue` | voicemail |
| Max Registrations | Simultaneous device registrations | 5 |
| Max Concurrent Calls | Per-extension concurrent call limit | 5 |

**Call Forwarding**:

| Type | When Triggered |
|---|---|
| Forward All | Unconditional — all calls immediately forwarded |
| Forward Busy | When extension is on another call |
| Forward No Answer | After ring timeout expires |
| Forward Unregistered | When no devices are registered |

**Other Features**:
- Do Not Disturb (DND) toggle
- Voicemail (enable/disable, PIN, email-to-voicemail, custom greeting)
- Follow-Me (ring external numbers after local devices)
- Call Recording (inbound and/or outbound)
- Hold Music (per-extension override)
- Call/Pickup Groups (answered elsewhere)
- E911 Location assignment
- Bypass Media (direct RTP between endpoints)

### Extension Profiles

Extension profiles are reusable templates that set defaults for extensions. They also support call handling rules that apply to all extensions using the profile.

### Call Handling Rules

Advanced call routing rules that override default behavior based on events and conditions:

**Events** (triggers):
- `Any Call` — evaluate for every incoming call
- `On Phone` — caller is busy on another line
- `No Answer` — ring timeout was reached
- `Unavailable` — DND, offline, or unregistered

**Conditions** (optional filters):
- Caller ID match (equals, contains, regex)
- Time of day (between hours)
- Day of week
- Date range
- Holiday list membership
- Presence state

**Actions**:
- `forward` — Forward to number with ring timeout
- `voicemail` — Send to voicemail with optional greeting selection
- `find_me` — Ring multiple numbers (simultaneous or sequential)
- `ring_devices` — Override which device types ring (softphone, desk, mobile)
- `reject` — Reject with reason (busy, unavailable, declined)

Rules are evaluated in priority order (lower number = higher priority). First matching rule wins.

### Users vs. Extensions

- **Users** are web login accounts (username/email/password) with roles
- **Extensions** are phone lines (SIP credentials, call settings)
- A user can optionally be linked to an extension
- Extensions can also log in directly using extension-based auth (for the user portal)

---

## Call Routing

### Inbound Routes

Inbound routes map incoming DID numbers to destinations. Managed at **Admin → Routing → Inbound**.

| Field | Description |
|---|---|
| DID Number/Pattern | The incoming number to match (exact or regex) |
| Caller ID Pattern | Optional: filter by caller number |
| Destination Type | Extension, Ring Group, Queue, IVR, Voicemail, External, Time Condition, Call Flow |
| Destination Value | The specific target (extension number, queue ID, etc.) |
| Priority | Order of evaluation (lower = first) |

Routes are reorderable via drag-and-drop. First matching route wins.

### Outbound Routes

Outbound routes match dialed patterns and send calls out through trunks. Managed at **Admin → Routing → Outbound**.

| Field | Description |
|---|---|
| Name | Route name |
| Digit Pattern | Regex to match (e.g., `^1?(\d{10})$` for US/CAN) |
| Digit Prefix | Required prefix (e.g., `1` for long distance) |
| Min/Max Digits | Digit length range |
| Strip Digits | Remove N leading digits before sending to trunk |
| Prepend | Add prefix before sending |
| Gateway | Which trunk to use |
| Priority | Route evaluation order |

Default outbound routes are seeded on first boot:
1. **Emergency** (911) — highest priority
2. **Local/Toll-Free** — 7-11 digits
3. **Long Distance** — 1 + 10 digits
4. **International** — 011 + variable

### Call Blocks

Block specific caller IDs or patterns from reaching the system.

---

## Feature Modules

### IVR / Auto-Attendant

Interactive Voice Response menus allow callers to navigate options via DTMF. Managed at **Admin → IVR**.

Features:
- Multi-level menus (IVR can lead to another IVR)
- Custom greetings (upload or TTS)
- Direct extension dialing within IVR
- Configurable timeout, max failures, max timeouts
- Visual flow editor for building IVR trees
- 9 option actions: transfer to extension, send to IVR, voicemail, queue, ring group, playback, hangup, repeat, exit

### Queues (Call Center)

ACD (Automatic Call Distribution) queues route calls to agents. Managed at **Admin → Queues**.

**Queue Settings**:
- **Strategy**: Ring All, Longest Idle, Round Robin, Top Down, Least Talk Time, Random
- **Timeouts**: Max wait time, max wait with no agents
- **Agent Settings**: Wrap-up time, reject/busy/no-answer delays
- **Announcements**: Position in queue, estimated wait time, periodic announcements
- **Callbacks**: Callers can request a callback instead of waiting
- **Exit Action**: What happens when queue times out (hangup, transfer)

**Agent Management**:
- Agents are extensions assigned to a queue with a tier level and position
- States: Logged Out, Available, Available (On Demand), On Break
- Agent login/logout via feature codes (*51/*52)
- Pause/unpause via API or feature codes

### Conferences

Conference bridges for multi-party calling. Managed at **Admin → Conferences**.

Features:
- Dial-in access via extension number
- Participant and moderator PINs
- Wait for moderator option
- Mute on join
- Recording (per-session)
- Live control (mute/unmute/deaf/kick/lock individual or all participants)
- Session history with participant tracking
- Member permissions (mute others, kick, lock/unlock, barge, set floor)
- Conference statistics (total sessions, participants, peak concurrent, durations)

### Ring Groups

Ring groups ring multiple extensions for incoming calls. Managed at **Admin → Ring Groups**.

- **Ring strategy**: Simultaneous, sequential, random
- **Destinations**: List of extensions with individual timeouts
- **Timeout action**: What to do if no one answers

### Feature Codes (Star Codes)

Feature codes are dialed star codes that trigger PBX actions. Managed at **Admin → Feature Codes**.

Feature codes are provisioned **per-tenant** — each tenant gets their own set. Codes start with `*` or `#` and can use regex for variable capture (e.g., `*57(\d{2})` for parking to a specific slot).

**Built-in Feature Code Actions**:

| Code | Action | Description |
|---|---|---|
| `*78` | DND Toggle | Toggle Do Not Disturb on/off |
| `*72` | Call Forward | Set unconditional call forwarding |
| `*73` | Cancel Forward | Cancel call forwarding |
| `*98` | Voicemail | Check your voicemail |
| `*97` | Voicemail (direct) | Check voicemail for a specific extension |
| `*55` | Park | Park current call in next available slot |
| `*57XX` | Park Slot | Park in specific slot (e.g., `*5701`) |
| `*58XX` | Park Retrieve | Retrieve call from specific slot |
| `*67` | Transfer (blind) | Blind transfer current call |
| `*86` | Intercom | One-way intercom to extension |
| `*80` | Speed Dial | Dial a speed dial entry |
| `*44` | Record Toggle | Toggle call recording on/off |
| `*51` | Queue Login | Log in to all assigned queues |
| `*52` | Queue Logout | Log out of all queues |
| `*99` | Conference | Join conference room |
| `*88` | Pickup | Directed call pickup |
| `*85` | Page Group | Page a group of extensions |

Custom feature codes can be created with actions:
- **Webhook**: Call an external URL
- **Lua**: Execute a Lua script
- **Custom**: Arbitrary FreeSWITCH dialplan application

### Call Flows / Toggles

Call flows are multi-state routing switches (e.g., Day/Night mode). Managed at **Admin → Call Flows**.

- Define 2+ states, each with a destination (IVR, queue, extension, voicemail, etc.)
- Toggle between states via feature code or API
- Current state is displayed in the admin UI
- Use in routing: inbound routes can target a call flow

### Time Conditions & Schedules

Time-based routing that changes destinations based on schedule. Managed at **Admin → Time Conditions**.

- **Schedule**: Days of week, start/end time, timezone
- **Match destination**: Where to route during scheduled hours
- **No-match destination**: Where to route outside hours
- **Holiday override**: Link to a holiday list for special routing on holidays

### Holiday Lists

Named lists of dates for holiday-specific routing. Managed at **Admin → Holidays**.
Can be synced from external sources. Linked to time conditions for automatic holiday overrides.

### Speed Dials

Speed dial groups with assigned extensions/numbers. Dialed via feature code (e.g., `*801`, `*802`).

### Paging Groups

Broadcast audio to multiple extensions simultaneously (intercom). Managed at **Admin → Paging Groups**.

### Call Broadcasts

Mass outbound call campaigns. Managed at **Admin → Call Broadcasts**.
Upload a list of numbers, assign a message/recording, and start the campaign. Start/stop control via API.

### Call Recording

Per-extension or per-route call recording. Managed at **Admin → Recordings**.

Features:
- Record inbound, outbound, or both
- Playback and download
- Recording notes
- Transcription (if configured)
- Feature code toggle (`*44`) for on-demand recording
- Tenant-level recording configuration

---

## Device Management

### Devices

Physical phones, softphones, and generic devices. Managed at **Admin → Devices**.

- Assign device to an extension (device registers as the extension)
- Support for multiple devices per extension (simultaneous ringing)
- Line configuration (which extensions appear on which buttons)
- Remote call control via API: hangup, transfer, hold, dial

### Device Profiles

Templates for device settings applied to multiple devices. Include early media, codec preferences, and other SIP parameters.

### Device Templates

Model-specific provisioning templates (Yealink T54W, Polycom VVX, etc.). System-level templates can be copied to tenant level for customization.

### Auto-Provisioning

Devices can auto-provision using the provisioning URL. The system generates device-specific configuration files based on templates, substituting variables for:
- SIP credentials (from extension)
- Server address
- Line assignments
- Firmware URL
- Tenant branding
- Feature keys / BLF subscriptions

### Client Registrations

WebRTC and softphone clients register through client registration records. These track:
- MAC/device identifier
- Extension assignment
- Registration status
- Last registration time

---

## Communication Services

### Voicemail

Per-extension voicemail boxes. Managed at **Admin → Voicemail**.

- Custom greeting files
- PIN protection
- Email notifications (voicemail-to-email with audio attachment)
- MWI (Message Waiting Indicator) updates to registered phones
- Web playback in admin and user portals
- Check via feature code (*98) or user portal

### SMS/MMS Messaging

Two-way SMS/MMS messaging. Managed at **Admin → Messaging**.

- Conversation-based UI (threaded by contact)
- Supports Telnyx as messaging provider
- Media transcoding for MMS (FFmpeg-based size optimization)
- Message queue with retry logic
- Real-time delivery via WebSocket
- Chatplans for automated reply routing (pattern matching, auto-reply, forwarding)
- Per-tenant messaging number assignment

### Fax

Send and receive faxes via T.38. Managed at **Admin → Fax Server**.

- **Fax Boxes**: Inbound fax destinations (linked to extension or DID)
- **Fax Endpoints**: Devices/extensions that can send faxes
- **Send Fax**: Upload document, enter destination, send
- **Job Tracking**: View pending, completed, and failed fax jobs with retry
- **Download**: Download received/sent fax documents

### Chat System

Internal chat rooms and queues. Managed at **Admin → Chat**.

- Chat rooms with members
- Chat queues for agent-based chat routing
- Message threading
- File attachments
- Read receipts
- Contact webhooks for external integrations

---

## Monitoring & Reporting

### CDR (Call Detail Records)

Call history with searchable/filterable views. Managed at **Admin → CDR**.

- Filter by date range, caller, destination, duration, direction
- Export to CSV
- Stored in PostgreSQL (primary) with optional ClickHouse sync for analytics
- CDRs ingested from FreeSWITCH via `mod_xml_cdr`

### Reports & Analytics

Pre-built reports at **Admin → Reports**:

| Report | Description |
|---|---|
| Call Volume | Calls per hour/day/month with trends |
| Agent Performance | Queue agent metrics (answer rate, talk time, hold time) |
| Queue Statistics | Queue wait times, abandonment rates, service levels |
| Extension Usage | Per-extension call counts and durations |
| KPI Dashboard | Key performance indicators summary |
| Number Usage | DID utilization metrics |

All reports support date range filtering and CSV export.

### Operator Panel

Real-time view of all extensions and their states (idle, ringing, busy, on hold, DND, offline). Available at **Admin → Live Ops → Operator Panel**.

### Live Operations

Real-time call monitoring and control:

- **Active Calls**: View all in-progress calls with metadata
- **Live Recording**: Start/stop recording on active calls
- **Queue Dashboard**: Real-time queue metrics and agent states
- **Conference Console**: Real-time conference participant management

### Audit Log

Full audit trail of administrative actions. Every create, update, and delete operation is logged with:
- Timestamp
- User who performed the action
- Action type
- Target resource
- Before/after values

### System Monitoring

- **System Status**: FreeSWITCH status, ESL connection, database health
- **System Stats**: Active calls, registrations, uptime
- **System Logs**: FreeSWITCH console output (live via WebSocket)
- **Grafana**: Preconfigured dashboards via Loki log aggregation

---

## User Portal

End users access the user portal at `/` after logging in with extension credentials.

Available features:
- **Softphone**: WebRTC dialer with call controls (answer, hold, transfer, mute, DTMF)
- **Messages**: SMS/MMS conversations
- **Voicemail**: Listen to messages, mark read/unread, delete
- **Conferences**: View and join personal conference rooms
- **Fax**: Send/receive faxes, view inbox/outbox
- **Contacts**: Personal contact directory with phone lookup
- **Recordings**: Listen to and download call recordings
- **History**: View call history with search/filter
- **Settings**: Update personal settings, change web password
