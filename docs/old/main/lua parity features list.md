# Lua Middleman Feasibility Analysis

## Executive Summary

After deep-diving into FusionPBX's production Lua scripts and comparing them against CallSign's current ESL-based modules, there is a **significant feature gap**. However, the right solution is **not** to add Lua as a middleman. Instead, the gap should be closed by expanding the existing ESL modules — everything Lua does, ESL can do, often better. Below is the full analysis with a third option (hybrid Lua) discussed for completeness.

---

## What FusionPBX's Lua Scripts Actually Do

### Ring Groups (`ring_groups/index.lua` — ~900 lines)

The production ring group script handles far more than basic call distribution:

| Feature | Description |
|---------|-------------|
| **5 Ring Strategies** | `simultaneous`, `sequence`, `rollover`, `random`, `enterprise` (each with different delimiters and timing logic) |
| **Call Screening** | Records caller's name before connecting; plays recording to answering party |
| **Follow-Me Integration** | If a ring group member has follow-me enabled, merges follow-me destinations into the ring group |
| **DND Checking** | Checks each destination's `do_not_disturb` status via `user_data` API; skips DND-enabled extensions |
| **Multi-Level Call Forwarding** | Recursively follows `forward_all_enabled` → `forward_all_destination` (up to 3 levels deep) |
| **Distinctive Ring / Alert-Info** | Sets `Alert-Info` header for different ring tones per ring group |
| **CID Prefix Manipulation** | Prepends configurable prefix to caller ID name and number |
| **Exit Key** | DTMF key binding (configurable) to escape to timeout destination during ringing |
| **Greeting Playback** | Plays audio greeting before ringing (e.g., "Please hold while we transfer your call") |
| **Ring Group Forwarding** | Entire ring group can be forwarded to a single destination |
| **Diversion Headers** | Sets SIP `Diversion` header for compliance with carrier call-forwarding disclosure |
| **Missed Call Email** | Queries `v_email_templates` from DB, renders template with call variables, sends via SMTP |
| **Custom FreeSWITCH Events** | Fires `CUSTOM RING_GROUPS` events with status, useful for BLF/presence |
| **Toll Allow Enforcement** | Checks toll_allow per-extension for external ring group destinations |
| **Verto Support** | Adds verto_contact endpoints alongside sofia_contact for WebRTC clients |
| **Enterprise Strategy** | Uses `:_:` delimiter with `originate_delay_start`/`originate_timeout` for true enterprise bridging |
| **SIP URI Destinations** | Supports external SIP URIs as ring group destinations alongside local extensions |

### Voicemail (`voicemail/index.lua` + 20 sub-modules)

The production voicemail system is a full IVR application:

| Feature | Description |
|---------|-------------|
| **PIN Authentication** | `check_password()` with configurable max attempts, lockout |
| **Main Menu IVR** | Full DTMF menu: 1=listen, 2=advanced, 5=repeat, 7=delete, 9=save, *=return |
| **Greeting Management** | Record new greeting, choose from multiple greetings (numbered), record name |
| **Message Forwarding** | Forward a voicemail to another extension's mailbox with optional intro recording |
| **Return Call** | Press key to call back the person who left the message |
| **MWI Notification** | `message_waiting()` sends SIP NOTIFY to phones for message lamp |
| **BLF Notification** | `blf_notify()` updates BLF status for voicemail buttons |
| **Email with Attachment** | Full SMTP email with `.wav` attachment, HTML template from `v_email_templates` |
| **SMS Notification** | `send_sms()` text notification when new voicemail arrives |
| **Tutorial Mode** | First-time user walkthrough for setting up voicemail |
| **Disk Quota** | Configurable storage limits per mailbox |
| **Timezone Handling** | Says date/time of messages in user's configured timezone |
| **Message Categories** | New vs saved messages, with separate navigation |
| **Voicemail Transcription** | Integration hook for speech-to-text services |
| **Delete After Email** | Configurable: keep or delete local copy after emailing |
| **Stream Seeking** | Fast-forward/rewind during message playback |

### Other Notable Lua Scripts

| Script | Key Features Not in CallSign |
|--------|------------------------------|
| **`follow_me/index.lua`** | Confirmation prompt ("Press 1 to accept this call"), sequential dialing with per-destination delays, caller ID screening |
| **`call_block/index.lua`** | Real-time caller ID matching against block list during call, with reject/voicemail/play-message actions |
| **`toll_allow/index.lua`** | Checks dialed number against extension's toll_allow rules before bridging |
| **`failure_handler/index.lua`** | Handles bridge failures — retries, fallback routing, failure-specific actions |
| **`missed_calls/index.lua`** | Logs missed calls to DB with CDR correlation |
| **`caller_id/index.lua`** | CNAM lookup (caller name from external DB/API), outbound CID manipulation |
| **`valet_park/index.lua`** | Call parking with lot number announcement |
| **`speed_dial/index.lua`** | Real-time speed-dial code lookup and bridging |
| **`emergency/index.lua`** | E911 routing with location data from extension's assigned location |

---

## What CallSign's ESL Modules Currently Cover

| Module | Lines | Status |
|--------|-------|--------|
| `voicemail/service.go` | 259 | Basic deposit/check. Has TODOs for: PIN auth, email, MWI, save/delete menu |
| `callcontrol/service.go` | — | General call operations |
| `conference/service.go` | — | Live conference management (well-developed) |
| `featurecodes/service.go` | — | Feature code processing |
| `ivr/service.go` | — | IVR menu processing |
| `queue/service.go` | — | ACD queue handling |
| `blf/service.go` | — | Presence/BLF monitoring |

---

## The Three Options

### Option 1: Expand ESL Modules (Recommended ✅)

**Add the missing Lua features directly into the Go ESL modules.** ESL outbound mode gives you everything `session:execute()` gives Lua, plus:

| Advantage | Detail |
|-----------|--------|
| **Same call control** | `conn.Execute("bridge"...)`, `conn.Execute("playback"...)`, `conn.Execute("record"...)` — identical to Lua's `session:execute()` |
| **DB access** | Already have — GORM queries are cleaner than Lua's string-concatenated SQL |
| **DTMF collection** | `conn.Execute("read"...)` or `conn.Execute("play_and_get_digits"...)` works via ESL |
| **FreeSWITCH API calls** | `conn.Send("api user_exists id 1001 example.com")` — same as Lua's `api:execute()` |
| **Email/SMS** | Go has mature libraries for SMTP, SMS APIs — better than Lua's |
| **Type safety** | Compile-time checks vs Lua's runtime errors |
| **Testing** | Go's testing framework vs no testing in Lua scripts |
| **Single deployment** | One binary, no file sync to FreeSWITCH scripts dir |
| **Observability** | Structured logging, metrics, traces — Lua has console logs only |

**What needs to be built:**
- Ring group: call screening, DND check, follow-me merge, CID prefix, distinctive ring, missed call email, exit key
- Voicemail: PIN auth, DTMF menu, greeting management, message forwarding, MWI/BLF, email+attachment, SMS
- Follow-me: confirm prompt, sequential dial
- Call parking, speed dial, toll-allow enforcement, failure handler, E911 routing

### Option 2: Lua Middleman (Not recommended ❌)

**Deploy FusionPBX Lua scripts alongside your Go backend.** FreeSWITCH's dialplan would call Lua scripts (via `mod_lua`) for call processing while your Go API serves `mod_xml_curl` for config.

| Disadvantage | Detail |
|--------------|--------|
| **Two runtimes** | Lua scripts + Go binary — doubled maintenance surface |
| **DB connection duplication** | Lua connects to PostgreSQL independently (via `config.conf`), separate connection pools |
| **Schema conflicts** | FusionPBX Lua expects `v_` prefixed tables with specific columns; your GORM models use different names |
| **Config file dependency** | Lua scripts need `/etc/fusionpbx/config.conf` — not env-var driven |
| **Incompatible models** | FusionPBX uses `domain_uuid` for multi-tenancy; you use `tenant_id`. Every query would need adaptation |
| **Deployment complexity** | Must sync Lua scripts to FreeSWITCH's scripts directory, manage versions |
| **No compile checks** | Lua errors surface at call-time, potentially dropping live calls |
| **Duplicate logic** | UI-facing API handlers + Lua scripts both touching same data |

**The Lua scripts were written for FusionPBX's data model. Adapting them to CallSign's schema would require rewriting most of the SQL — at which point you're better off writing Go.**

### Option 3: Hybrid — Lua for edge cases only (Possible but niche ⚠️)

Use Lua only for features where **in-process execution** matters:

| Use Case | Why Lua helps |
|----------|---------------|
| **Inline dialplan actions** | `<action application="lua" data="..."/>` executes synchronously within the dialplan phase, before call is established — ESL outbound only works after the call is set up |
| **Pre-answer manipulation** | Setting variables, manipulating SIP headers before `answer` — technically possible via ESL but Lua is simpler |
| **Ultra-low-latency lookups** | Lua runs in-process with zero network hop; ESL has TCP overhead (~1-2ms) |

**Practically**, the latency difference is negligible and ESL's `pre_answer` commands handle most pre-answer scenarios. The only real edge case would be highly custom dialplan logic that needs to run before the call enters an ESL socket.

---

## Recommendation

**Option 1 — Expand your ESL modules.** The feature gap is a *completeness* issue, not an *architecture* issue. Your Go+ESL approach can do 100% of what FusionPBX's Lua does. The priority features to implement (roughly in order of user impact):

### High Priority
1. **Ring group enhancements** — DND checking, follow-me integration, multi-level forwarding, call screening, missed call notifications
2. **Voicemail full IVR** — PIN auth, DTMF menu (save/delete/repeat/forward), greeting management, MWI/BLF events
3. **Email notifications** — Missed call and voicemail email via DB-stored templates

### Medium Priority  
4. **Follow-me with confirm** — "Press 1 to accept this call"
5. **Failure handler** — Fallback routing on bridge failure
6. **Toll-allow enforcement** — Check before bridging to external numbers
7. **CID prefix/manipulation** — Ring group and destination-level caller ID control
8. **Distinctive ring** — Alert-Info headers per ring group or destination

### Lower Priority
9. **Call parking** — Valet park with lot numbers
10. **Speed dial** — Real-time code lookup
11. **Call screening** — Record caller name before connecting
12. **E911 integration** — Location-aware emergency routing
