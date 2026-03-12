# CallSign PBX — FreeSWITCH Integration Guide

## Overview

CallSign interfaces with FreeSWITCH through two complementary communication channels. Together, they replace static configuration files with a fully dynamic, database-driven system.

| Channel | Direction | Protocol | Purpose |
|---|---|---|---|
| **XML CURL** | FreeSWITCH → CallSign | HTTP POST | Dynamic configuration (directory, dialplan, sofia config) |
| **ESL (Event Socket)** | CallSign → FreeSWITCH | TCP socket | Command execution, event subscription, call control |

---

## XML CURL Integration

### How It Works

FreeSWITCH's `mod_xml_curl` module replaces static XML file lookups with HTTP requests. Whenever FreeSWITCH needs to resolve a user, generate a dialplan, or load module configuration, it POSTs a request to the CallSign API.

### FreeSWITCH Configuration

FreeSWITCH must be configured to use XML CURL. The relevant configuration files:

**`/etc/freeswitch/autoload_configs/xml_curl.conf.xml`**:
```xml
<configuration name="xml_curl.conf" description="cURL XML Gateway">
  <bindings>
    <binding name="callsign">
      <param name="gateway-url" value="http://127.0.0.1:8080/api/freeswitch/xmlapi"/>
      <param name="bindings" value="directory|dialplan|configuration"/>
      <param name="method" value="POST"/>
    </binding>
  </bindings>
</configuration>
```

### API Endpoint

All XML CURL requests hit **`POST /api/freeswitch/xmlapi`** (or the section-specific routes `/api/freeswitch/directory`, `/api/freeswitch/dialplan`, `/api/freeswitch/configuration`).

**Authentication**: Requests from localhost (`127.0.0.1`, `::1`) are allowed without authentication. Non-local requests require HTTP Basic Auth where the password matches `FREESWITCH_API_KEY`.

### Request Format

FreeSWITCH sends multipart form data with 100+ fields. The `XMLCurlRequest` struct in `xmlcurl.go` extracts the relevant ones:

| Field | Used In | Description |
|---|---|---|
| `section` | All | Which section: `directory`, `dialplan`, `configuration`, `phrases` |
| `tag_name` | Configuration | e.g., `configuration` |
| `key_name` | Configuration | e.g., `name` |
| `key_value` | Configuration | e.g., `sofia.conf` |
| `user` | Directory | SIP username being looked up |
| `domain` | Directory | SIP domain / tenant domain |
| `action` | Directory | `sip_auth`, `reverse-auth`, `message-count` |
| `purpose` | Directory | `gateways`, `network-list` |
| `sip_profile` | Directory | Which SIP profile is requesting |
| `context` | Dialplan | Dialplan context (e.g., `default`, `public`) |
| `destination_number` | Dialplan | The dialed number |
| `caller_id_number` | Dialplan | Caller's number |

### Response Format

All responses are XML documents in FreeSWITCH's expected format:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<document type="freeswitch/xml">
  <section name="directory">
    <!-- content -->
  </section>
</document>
```

If no data is found, a "not found" result is returned, which tells FreeSWITCH to fall back to static files:

```xml
<document type="freeswitch/xml">
  <section name="result">
    <result status="not found"/>
  </section>
</document>
```

---

## Section: Directory (`directory.go`)

Handles SIP user registration, authentication, and presence.

### When Called

- SIP REGISTER request arrives
- SIP INVITE needs authentication
- Voicemail message count query
- BLF/presence subscription

### What It Does

1. Looks up the **tenant** by domain name
2. Finds the **extension** by the `user` field within that tenant
3. Generates XML with:
   - SIP authentication credentials (password from DB)
   - Extension parameters (CID, effective CID, voicemail settings)
   - Extension variables (tenant_id, domain_name, user_context, etc.)
   - Call handling chain variables (call forwarding, DND, etc.)

### Key Variables Set

| Variable | Purpose |
|---|---|
| `tenant_id` | Multi-tenant isolation in dialplan |
| `domain_name` | Tenant's SIP domain |
| `user_context` | Dialplan context for this user (typically `default`) |
| `effective_caller_id_name/number` | Outbound caller ID |
| `callgroup` | Extension's call group |
| `voicemail_enabled` | Whether voicemail is active |
| `call_forward_*` | Call forwarding settings (busy, no-answer, unconditional) |

---

## Section: Dialplan (`dialplan.go`)

Generates dynamic dialplan XML based on the request context and destination number.

### Contexts

The dialplan generates rules for different contexts:

| Context | Triggered By | Purpose |
|---|---|---|
| `default` | Internal extension calls | Extension-to-extension dialing, feature codes, outbound |
| `public` | Inbound calls from trunks | DID → routing destination mapping |
| `features` | Feature code execution | Star-code handling (*67, *72, *98, etc.) |

### Dialplan Priority Order (Default Context)

1. **Emergency numbers** (911, 112) — Highest priority
2. **Feature codes** — Star codes (*67, *72, *98, etc.)
3. **Internal extensions** — Direct extension dialing
4. **Ring groups** — Group dialing patterns
5. **Conference bridges** — Conference access numbers
6. **Queue access** — Call center queue numbers
7. **IVR menus** — Auto-attendant access
8. **Outbound routes** — PSTN dialing via gateways (pattern-matched)
9. **Call blocks** — Block specific patterns

### Inbound Route Matching (Public Context)

For incoming calls, the dialplan matches the `destination_number` (DID) against configured inbound routes and generates actions based on the route's destination:

| Destination Type | Dialplan Action |
|---|---|
| Extension | `bridge` to extension's endpoints |
| Ring Group | `bridge` with ring strategy |
| Queue | `callcenter` application |
| IVR Menu | `socket` to IVR module |
| Voicemail | `socket` to voicemail module |
| External Number | `bridge` via gateway |
| Time Condition | Conditional branching |
| Call Flow | Toggle-based routing |

### Generated Dialplan Structure

```xml
<document type="freeswitch/xml">
  <section name="dialplan">
    <context name="default">
      <!-- Feature codes -->
      <extension name="feature-dnd-toggle">
        <condition field="destination_number" expression="^*78$">
          <action application="socket" data="127.0.0.1:9001 async full"/>
        </condition>
      </extension>
      
      <!-- Internal dialing -->
      <extension name="ext-1001">
        <condition field="destination_number" expression="^1001$">
          <action application="set" data="tenant_id=1"/>
          <action application="socket" data="127.0.0.1:9001 async full"/>
        </condition>
      </extension>
      
      <!-- Outbound routes -->
      <extension name="outbound-us-domestic">
        <condition field="destination_number" expression="^1?(\d{10})$">
          <action application="bridge" data="sofia/gateway/main-trunk/$1"/>
        </condition>
      </extension>
    </context>
  </section>
</document>
```

---

## Section: Configuration (`configuration.go`)

Provides dynamic module configuration to FreeSWITCH.

### Handled Configurations

| `key_value` | What It Returns |
|---|---|
| `sofia.conf` | SIP profile definitions (internal, external, WebRTC, custom) |
| `event_socket.conf` | ESL connection settings (port 8021, password) |
| `callcenter.conf` | Queue definitions, tier/agent assignments |
| `acl.conf` | Access control list rules |
| `conference.conf` | Conference profile definitions |
| `local_stream.conf` | Music-on-hold stream definitions |
| `xml_cdr.conf` | CDR logging URL (→ `/api/freeswitch/cdr`) |

### Sofia Profile Generation

For `sofia.conf`, the handler generates complete SIP profile XML from database records:

1. Queries all `SIPProfile` records with their `SIPProfileSetting` values
2. For each profile, generates `<profile>` with:
   - Settings (SIP port, RTP range, codecs, TLS, etc.)
   - Gateways attached to the profile
   - Domains served by the profile

### Gateway Generation

Gateways (`Gateway` model) are rendered as Sofia gateway XML within their assigned SIP profile:

```xml
<gateway name="trunk-provider">
  <param name="username" value="user"/>
  <param name="password" value="pass"/>
  <param name="realm" value="sip.provider.com"/>
  <param name="proxy" value="sip.provider.com"/>
  <param name="register" value="true"/>
</gateway>
```

---

## Section: CDR (`cdr.go`)

Receives call detail records from `mod_xml_cdr` via HTTP POST.

### FreeSWITCH Configuration

```xml
<!-- autoload_configs/xml_cdr.conf.xml -->
<configuration name="xml_cdr.conf">
  <settings>
    <param name="url" value="http://127.0.0.1:8080/api/freeswitch/cdr"/>
    <param name="encode" value="true"/>
    <param name="retries" value="3"/>
  </settings>
</configuration>
```

### What It Does

1. Parses the XML CDR payload
2. Extracts call details (caller, destination, duration, hangup cause, timestamps)
3. Resolves tenant from domain/variables
4. Stores as `CallRecord` in PostgreSQL
5. Triggers CDR sync to ClickHouse (if enabled, via periodic job)

---

## ESL — Event Socket Layer

### Architecture Summary

The ESL integration has two modes:

**Inbound Mode** (CallSign connects TO FreeSWITCH):
- Single persistent TCP connection to port 8021
- Used for: sending API commands, subscribing to events
- Managed by `esl.Client`

**Outbound Mode** (FreeSWITCH connects TO CallSign):
- Multiple TCP servers on loopback addresses
- FreeSWITCH connects when dialplan action is `socket`
- Each module (call control, voicemail, queue, etc.) runs its own server

### Inbound Client

The inbound client (`services/esl/client.go`) manages a single connection:

```
CallSign API ─── TCP ───► FreeSWITCH ESL (:8021)
                          Authenticate with password
                          Subscribe to events
                          Send api/bgapi commands
                          Receive event stream
```

**Connection lifecycle**:
1. `Connect()` — Dial FreeSWITCH ESL, authenticate
2. `Subscribe(events...)` — Subscribe to event types
3. `StartEventLoop()` — Begin reading events into buffered channel
4. Events flow to `EventProcessor` for dispatch
5. Auto-reconnect on connection loss (10 attempts with backoff)

**API Commands**:
- `API(command)` — Synchronous: blocks until result. E.g., `api status`, `api sofia status`
- `BgAPI(command)` — Asynchronous: returns job UUID. Used for long-running commands like `sofia profile restart` that would otherwise block the connection.

### Event Processing

The `EventProcessor` (`events.go`) reads from the client's event channel and dispatches to handlers:

```
Event Channel (1000 buffer)
    │
    ▼
EventProcessor.processEvents()
    │
    ├── Lookup channel UUID → find CallSession
    ├── Dispatch to registered handlers by event name
    └── Dispatch to wildcard (*) handlers
```

**Default Handlers** track call sessions through their lifecycle:

| Event | Handler Action |
|---|---|
| `CHANNEL_CREATE` | Creates `CallSession`, sets A-leg, state → `ringing` |
| `CHANNEL_ANSWER` | Records answer time, state → `answered` |
| `CHANNEL_BRIDGE` | Registers B-leg, state → `bridged` |
| `CHANNEL_HANGUP_COMPLETE` | Records hangup cause, removes session from tracker |

### Outbound Modules

Each module implements the `Service` interface and runs an ESL server on a unique loopback address. FreeSWITCH connects to these when the dialplan routes via `<action application="socket" data="address async full"/>`.

**Module handler flow**:

```
FreeSWITCH ──── TCP connect ────► Module Server (127.0.0.x:9001)
                                      │
                                      ▼
                                  module.Handle(conn)
                                      │
                                  conn.Execute("answer")
                                  conn.Execute("playback", "greeting.wav")
                                  conn.Execute("bridge", "user/1001@domain")
                                      │
                                  conn closes when call ends
```

Each connection represents a single call session. The module has full control over the call, including:
- Reading channel variables
- Executing applications (answer, bridge, playback, record, etc.)
- Collecting DTMF input
- Transferring to other modules
- Setting variables that persist across the call

### Module Details

**Call Control** (`modules/callcontrol/`):
- Handles general extension-to-extension bridging
- Implements call handling rules (find-me/follow-me, DND, forwarding)
- Manages SIP endpoint bridging with failover
- Bridge timeout/no-answer handling with voicemail fallback

**Voicemail** (`modules/voicemail/`):
- Records voicemail messages
- Plays greeting (personal or default)
- Stores messages in DB as `VoicemailMessage`
- Sends MWI (Message Waiting Indicator) events
- Triggers email notification (voicemail-to-email)

**Queue** (`modules/queue/`):
- ACD queue handling via mod_callcenter integration
- Agent status management
- Queue statistics tracking
- Hold music during wait

**IVR** (`modules/ivr/`):
- Multi-level IVR menu traversal
- DTMF collection and option routing
- Timeout and invalid input handling
- Dynamic prompt playback (TTS or recorded)

**Conference** (`modules/conference/`):
- Conference room management
- Live participant control (mute, deaf, kick)
- Recording control
- Session and participant tracking via DB
- Uses the ESL Manager for API commands

**Feature Codes** (`modules/featurecodes/`):
- Star-code handling (*67, *72, *78, *98, etc.)
- DB lookup of feature code definitions per tenant
- Executes the mapped action (toggle DND, check voicemail, speed dial, etc.)
- Supports dynamic feature code creation

**BLF** (`modules/blf/`):
- Busy Lamp Field subscriptions
- Extension presence state updates

---

## Data Flow Diagrams

### Inbound PSTN Call → Extension

```
PSTN ──► SIP Trunk ──► FreeSWITCH
                          │
                          ├─ POST /freeswitch/dialplan (context=public)
                          │   └─ API returns: match DID → route to extension 1001
                          │
                          ├─ POST /freeswitch/directory (user=1001, domain=tenant.com)
                          │   └─ API returns: auth credentials, extension settings
                          │
                          ├─ Dialplan: <action application="socket" data="127.0.0.1:9001"/>
                          │
                          └─► Call Control module handles bridging
                                │
                                ├─ Query call handling rules
                                ├─ Ring extension endpoints
                                ├─ If no answer → transfer to voicemail module
                                └─ ESL events → WebSocket → UI
```

### Extension → Outbound PSTN Call

```
Extension ──► SIP REGISTER + INVITE ──► FreeSWITCH
                                            │
                                            ├─ POST /freeswitch/directory (authentication)
                                            │
                                            ├─ POST /freeswitch/dialplan (context=default)
                                            │   └─ API returns: match outbound pattern → gateway route
                                            │
                                            └─► Bridge via gateway to PSTN
```

### Feature Code (*98 Check Voicemail)

```
Extension dials *98 ──► FreeSWITCH
                           │
                           ├─ POST /freeswitch/dialplan (dest=*98, context=default)
                           │   └─ API returns: route to feature codes module
                           │
                           ├─ <action application="socket" data="featurecodes_addr"/>
                           │
                           └─► Feature Codes module
                                 │
                                 ├─ Look up *98 = "check voicemail" action
                                 ├─ Transfer call to voicemail module
                                 └─ Voicemail plays messages, handles DTMF
```

### Configuration Change Propagation

```
Admin changes extension password in UI
    │
    ├─ PUT /api/extensions/1001 → API updates PostgreSQL
    │
    ├─ API calls eslManager.ReloadXML()
    │   └─ ESL Client sends: "api reloadxml" to FreeSWITCH
    │
    ├─ XML cache is flushed for affected patterns
    │
    └─ Next SIP REGISTER → FreeSWITCH re-fetches directory via XML CURL
         └─ New password is returned
```

---

## Cache Management

XML CURL responses are cached with configurable TTLs to reduce database load:

| Section | Default TTL | Invalidation |
|---|---|---|
| Configuration | 1 hour | Manual flush or API change triggers |
| Directory | 5 minutes | Extension/user updates |
| Dialplan | 30 minutes | Route/feature code changes |

**Cache endpoints**:
- `GET /api/freeswitch/cache/flush` — Flush all or by pattern
- `GET /api/freeswitch/cache/stats` — Cache hit/miss statistics

The API automatically flushes relevant cache entries when data is modified through the admin interface (e.g., creating an extension flushes directory cache for that tenant's domain).

---

## SIP Profile Management

SIP profiles are managed through a hybrid approach:

1. **First boot**: `SIPProfileImporter` scans `/etc/freeswitch/sip_profiles/` for existing XML files and imports them into the database
2. **After import**: Profiles are managed exclusively via the database and admin UI
3. **On change**: `SIPProfileWriter` syncs DB profiles back to XML files (FreeSWITCH reads these for `sofia.conf`)
4. **Reload**: `SofiaRescan` or `SofiaRestart` commands are sent via ESL to apply changes

Protected system profiles (`internal`, `external`) cannot be deleted but can be modified.
