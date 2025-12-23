# CallSign Architecture

This document provides a technical overview of the CallSign PBX platform architecture.

## System Components

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                   CLIENT LAYER                                        │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                       │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐    │
│   │   Browser    │  │  SIP Phone   │  │   WebRTC     │  │   Provisioning       │    │
│   │   (Vue.js)   │  │   (Yealink)  │  │   Softphone  │  │   Config Request     │    │
│   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘    │
│          │                 │                 │                      │                 │
└──────────┼─────────────────┼─────────────────┼──────────────────────┼─────────────────┘
           │                 │                 │                      │
           ▼                 ▼                 ▼                      ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              EDGE / PROXY LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                       │
│   ┌───────────────────────────────────────────────────────────────────────────┐      │
│   │                          Caddy (Reverse Proxy)                              │      │
│   │   - TLS Termination (Let's Encrypt)                                         │      │
│   │   - HTTP/2 + WebSocket Upgrade                                              │      │
│   │   - Route: /api/* → Go API (8080)                                           │      │
│   │   - Route: /* → Vue UI (static/5173)                                        │      │
│   └───────────────────────────────────────────────────────────────────────────┘      │
│                                                                                       │
└──────────────────────────────────┬──────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              APPLICATION LAYER                                        │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                       │
│   ┌─────────────────────────────────────────────────────────────────────────────┐    │
│   │                           Go Iris API (Port 8080)                             │    │
│   │                                                                               │    │
│   │   ┌────────────────┐ ┌────────────────┐ ┌────────────────┐ ┌──────────────┐  │    │
│   │   │  Auth Handler  │ │ Tenant Handler │ │ System Handler │ │ User Handler │  │    │
│   │   └────────────────┘ └────────────────┘ └────────────────┘ └──────────────┘  │    │
│   │                                                                               │    │
│   │   ┌────────────────┐ ┌────────────────┐ ┌────────────────┐ ┌──────────────┐  │    │
│   │   │ Device Handler │ │ Routing Handler│ │  CDR Handler   │ │  FS Handler  │  │    │
│   │   └────────────────┘ └────────────────┘ └────────────────┘ └──────────────┘  │    │
│   │                                                                               │    │
│   ├─────────────────────────────────────────────────────────────────────────────┤    │
│   │                              MIDDLEWARE                                       │    │
│   │   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │    │
│   │   │  JWT Auth   │ │   Tenant    │ │   Audit     │ │  CORS / Recovery    │    │    │
│   │   └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────────┘    │    │
│   │                                                                               │    │
│   ├─────────────────────────────────────────────────────────────────────────────┤    │
│   │                              SERVICES                                         │    │
│   │   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │    │
│   │   │ ESL Manager │ │   Logging   │ │ Encryption  │ │    XML Cache        │    │    │
│   │   └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────────┘    │    │
│   └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                       │
└──────────────────────────────────┬──────────────────────────────────────────────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
        ▼                          ▼                          ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────────────────────┐
│                   │   │                   │   │                                   │
│    PostgreSQL     │   │     Grafana       │   │           FreeSWITCH              │
│                   │   │      Loki         │   │                                   │
│   - Users         │   │                   │   │   ┌──────────────────────────┐    │
│   - Tenants       │   │   - API Logs      │   │   │     mod_xml_curl         │    │
│   - Extensions    │   │   - CDR Logs      │   │   │  (directory/dialplan/    │    │
│   - Devices       │   │   - Error Logs    │   │   │   configuration)         │    │
│   - Dialplans     │   │                   │   │   └──────────────────────────┘    │
│   - CDR           │   │                   │   │                                   │
│   - Recordings    │   │                   │   │   ┌──────────────────────────┐    │
│                   │   │                   │   │   │     mod_xml_cdr          │    │
│                   │   │                   │   │   │   (POST to /freeswitch/  │    │
│                   │   │                   │   │   │    cdr)                  │    │
│                   │   │                   │   │   └──────────────────────────┘    │
│                   │   │                   │   │                                   │
│                   │   │                   │   │   ┌──────────────────────────┐    │
│                   │   │                   │   │   │    ESL (Event Socket)    │    │
│                   │   │                   │   │   │   - Call Control         │    │
│                   │   │                   │   │   │   - Voicemail            │    │
│                   │   │                   │   │   │   - Conference           │    │
│                   │   │                   │   │   │   - Queue                │    │
│                   │   │                   │   │   └──────────────────────────┘    │
└───────────────────┘   └───────────────────┘   └───────────────────────────────────┘
```

---

## Multi-Tenant Architecture

### Tenant Isolation

All data is scoped by `tenant_id`:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SYSTEM ADMIN                                     │
│   - Full access to all tenants                                                │
│   - Can impersonate tenant admins (X-Tenant-ID header)                        │
│   - Manages: Gateways, SIP Profiles, Global Dial Plans                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
        ▼                          ▼                          ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│    Tenant A       │   │    Tenant B       │   │    Tenant C       │
│                   │   │                   │   │                   │
│ - Extensions      │   │ - Extensions      │   │ - Extensions      │
│ - Devices         │   │ - Devices         │   │ - Devices         │
│ - IVR Menus       │   │ - IVR Menus       │   │ - IVR Menus       │
│ - Queues          │   │ - Queues          │   │ - Queues          │
│ - Numbers/DIDs    │   │ - Numbers/DIDs    │   │ - Numbers/DIDs    │
│ - Users           │   │ - Users           │   │ - Users           │
│                   │   │                   │   │                   │
│ tenant_id = 1     │   │ tenant_id = 2     │   │ tenant_id = 3     │
└───────────────────┘   └───────────────────┘   └───────────────────┘
```

### Tenant Profiles

Tenant profiles define limits and features:

| Limit | Description |
|-------|-------------|
| `max_extensions` | Maximum extensions allowed |
| `max_devices` | Maximum provisioned devices |
| `max_queues` | Maximum call queues |
| `max_conferences` | Maximum conference rooms |
| `max_call_duration` | Maximum call duration (seconds) |
| `recording_enabled` | Allow call recording |
| `sms_enabled` | Allow SMS messaging |

---

## Data Flow

### Authentication Flow

```
┌──────────┐                  ┌──────────┐                  ┌──────────┐
│  Client  │                  │   API    │                  │  Database│
└────┬─────┘                  └────┬─────┘                  └────┬─────┘
     │                             │                             │
     │  POST /auth/login           │                             │
     │  {username, password}       │                             │
     │────────────────────────────►│                             │
     │                             │  Find user by username      │
     │                             │────────────────────────────►│
     │                             │                             │
     │                             │  Verify password (bcrypt)   │
     │                             │◄────────────────────────────│
     │                             │                             │
     │  {token: "eyJ...", user}    │                             │
     │◄────────────────────────────│                             │
     │                             │                             │
     │  GET /extensions            │                             │
     │  Authorization: Bearer ...  │                             │
     │────────────────────────────►│                             │
     │                             │  Verify JWT                 │
     │                             │  Extract tenant_id          │
     │                             │────────────────────────────►│
     │                             │                             │
     │  {data: [...]}              │  WHERE tenant_id = X        │
     │◄────────────────────────────│◄────────────────────────────│
```

### Call Flow (Inbound)

```
┌────────────┐     ┌────────────┐     ┌────────────┐     ┌────────────┐
│   PSTN     │     │ FreeSWITCH │     │  Go API    │     │  Database  │
└─────┬──────┘     └─────┬──────┘     └─────┬──────┘     └─────┬──────┘
      │                  │                  │                  │
      │  INVITE          │                  │                  │
      │─────────────────►│                  │                  │
      │                  │                  │                  │
      │                  │  POST /freeswitch/dialplan          │
      │                  │  {DID, caller_id, ...}              │
      │                  │─────────────────►│                  │
      │                  │                  │  Lookup route    │
      │                  │                  │─────────────────►│
      │                  │                  │                  │
      │                  │  <dialplan XML>  │  Return action   │
      │                  │◄─────────────────│◄─────────────────│
      │                  │                  │                  │
      │                  │  Bridge to ext   │                  │
      │                  │═════════════════════════════════════│
      │  200 OK          │                  │                  │
      │◄─────────────────│                  │                  │
```

---

## Service Architecture

### ESL Services

Each ESL service binds to a unique loopback IP:

```
┌────────────────────────────────────────────────────────────────┐
│                        ESL Manager                               │
│                                                                  │
│   ┌─────────────────────┐  ┌─────────────────────┐             │
│   │  Call Control       │  │  Voicemail          │             │
│   │  127.0.0.1:9001     │  │  127.0.0.2:9001     │             │
│   └─────────────────────┘  └─────────────────────┘             │
│                                                                  │
│   ┌─────────────────────┐  ┌─────────────────────┐             │
│   │  Conference         │  │  Queue              │             │
│   │  127.0.0.4:9001     │  │  127.0.0.5:9001     │             │
│   └─────────────────────┘  └─────────────────────┘             │
│                                                                  │
│   ┌─────────────────────┐                                       │
│   │  Feature Codes      │                                       │
│   │  127.0.0.6:9001     │                                       │
│   └─────────────────────┘                                       │
└────────────────────────────────────────────────────────────────┘
```

FreeSWITCH routes calls to these services via `socket` application:

```xml
<extension name="voicemail">
  <condition field="destination_number" expression="^\*97$">
    <action application="socket" data="127.0.0.2:9001 async full"/>
  </condition>
</extension>
```

---

## Security

### Authentication

| Method | Usage |
|--------|-------|
| JWT Bearer | API requests |
| Basic Auth | FreeSWITCH XML CURL |
| X-Internal-Key | Internal service calls |
| MAC Address | Device provisioning |

### Data Protection

- Passwords: bcrypt hashed
- SIP passwords: AES-256-GCM encrypted at rest
- JWT: HS256 signed, configurable expiration
- HTTPS: Caddy with Let's Encrypt
