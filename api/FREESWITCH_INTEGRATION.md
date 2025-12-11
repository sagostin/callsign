# CallSign API - FreeSWITCH Integration

## Overview

This document describes how CallSign integrates with FreeSWITCH for telephony functionality.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CallSign API                             │
│                    (Go Iris + PostgreSQL)                        │
├──────────────────┬────────────────────┬─────────────────────────┤
│  mod_xml_curl    │   Go ESL Client    │   Go ESL Servers        │
│  (config/dialplan│   (event monitor)  │   (call handling)       │
│   /directory)    │                    │                         │
└────────┬─────────┴─────────┬──────────┴────────┬────────────────┘
         │                   │                    │
         └───────────────────┴────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   FreeSWITCH    │
                    └─────────────────┘
```

## Components

### 1. mod_xml_curl Configuration Provider

**Endpoint:** `POST /freeswitch/xmlapi`

Serves dynamic XML for:
- **directory** - SIP user authentication (a1-hash)
- **configuration** - sofia.conf, ACL, IVR menus
- **dialplan** - Call routing rules

See: `handlers/freeswitch/`

### 2. Event Socket Layer (Modular)

Go services for call handling via outbound ESL using a **modular Service interface**:

| Service | Port | Purpose |
|---------|------|---------|
| `callcontrol` | 127.0.0.1:9001 | General call routing |
| `voicemail` | 127.0.0.2:9001 | VM deposit/retrieval + DB |
| `conference` | 127.0.0.4:9001 | Conference rooms |
| `queue` | 127.0.0.5:9001 | Call center queues |

**Architecture:**
```
services/esl/
├── service.go         # Service interface + ModuleRegistry
├── client.go          # Inbound ESL connection
├── server.go          # Outbound socket server
├── session.go         # B2B session tracking
├── events.go          # Event processor
└── modules/
    ├── callcontrol/   # General routing
    ├── voicemail/     # DB-integrated VM
    ├── queue/         # mod_callcenter sync
    └── conference/    # Conference handling
```

**Adding a new module:**
```go
type MyService struct {
    *esl.BaseService
}

func (s *MyService) Name() string { return "myservice" }
func (s *MyService) Address() string { return "127.0.0.6:9001" }
func (s *MyService) Handle(conn *eventsocket.Connection) {
    // Handle call
}
```

### 3. SIP Profiles

| Profile | Port | Use Case |
|---------|------|----------|
| `internal` | 5060 | Desk phones, ATAs |
| `webrtc` | 7443/WSS | Browser softphones |
| `public` | 5080 | SIP trunks (inbound/outbound) |

## FreeSWITCH Configuration

### xml_curl.conf.xml

```xml
<configuration name="xml_curl.conf">
  <bindings>
    <binding name="all">
      <param name="gateway-url" 
             value="http://127.0.0.1:8080/freeswitch/xmlapi"
             bindings="configuration,directory,dialplan"/>
      <param name="gateway-credentials" value="freeswitch:API_KEY"/>
      <param name="auth-scheme" value="basic"/>
    </binding>
  </bindings>
</configuration>
```

### Socket Application (for ESL services)

```xml
<!-- Route to Go voicemail service -->
<extension name="voicemail">
  <condition field="destination_number" expression="^\*97$">
    <action application="socket" data="127.0.0.2:9001 async full"/>
  </condition>
</extension>
```

## Database Models

### SIP-Related

- `Extension` - SIP users with auth, caller ID, forwarding
- `SIPProfile` - Sofia profiles with settings
- `Gateway` - SIP trunks
- `Dialplan` - Call routing with pre-generated XML

### Call Tracking

- `CallSession` - B2B session with A/B leg tracking (planned)
- `CDR` - Call detail records (planned)

## Environment Variables

```bash
# FreeSWITCH ESL
FREESWITCH_HOST=localhost
FREESWITCH_PORT=8021
FREESWITCH_PASSWORD=ClueCon

# XML CURL Auth
FREESWITCH_API_KEY=your-secret-key
```

## Why Go ESL Instead of Lua?

| Factor | Lua (FusionPBX) | Go ESL |
|--------|-----------------|--------|
| Scalability | mod_lua threads | Goroutines |
| DB Access | Separate pool | Shared pool |
| Debugging | FS logs only | Go debugger |
| Testing | Manual | Unit tests |
| Deployment | FS restart | Hot reload |

## Service Separation

Using multiple loopback IPs maximizes port space:

```
127.0.0.1:9001  → callcontrol
127.0.0.2:9001  → voicemail
127.0.0.3:9001  → hospitality
127.0.0.4:9001  → conference
127.0.0.5:9001  → queue
```

Each service runs as a goroutine in the main process but can be split out later.

## References

- [go-eventsocket](https://github.com/fiorix/go-eventsocket)
- [mod_xml_curl](https://developer.signalwire.com/freeswitch/FreeSWITCH-Explained/Modules/mod_xml_curl_1049001/)
- [mod_event_socket](https://developer.signalwire.com/freeswitch/FreeSWITCH-Explained/Modules/mod_event_socket_1048924/)
