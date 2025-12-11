# CallSign FreeSWITCH Installation Guide

## Quick Start

1. Copy configuration files to FreeSWITCH:
```bash
cp -r install/freeswitch/conf/* /etc/freeswitch/
```

2. Update API credentials in:
   - `/etc/freeswitch/autoload_configs/xml_curl.conf.xml`
   - `/etc/freeswitch/autoload_configs/xml_cdr.conf.xml`

3. Restart FreeSWITCH:
```bash
systemctl restart freeswitch
```

## Required Modules

| Module | Purpose |
|--------|---------|
| `mod_xml_curl` | Dynamic config from CallSign API |
| `mod_xml_cdr` | CDR posting to CallSign API |
| `mod_callcenter` | Queue/ACD features |
| `mod_conference` | Conference rooms |
| `mod_voicemail` | Voicemail boxes |
| `mod_event_socket` | ESL for real-time control |

## Configuration Files

### xml_curl.conf.xml
Points to: `http://api:8080/freeswitch/xmlapi`
Handles: `directory|dialplan|configuration`

### xml_cdr.conf.xml
Points to: `http://api:8080/freeswitch/cdr`
Handles: Call Detail Records

## Environment Variables

Set in CallSign API `.env`:
```
FREESWITCH_HOST=localhost
FREESWITCH_PORT=8021
FREESWITCH_PASSWORD=ClueCon
FREESWITCH_API_KEY=your-api-key
```

## Feature Status

| Feature | Status |
|---------|--------|
| Directory (SIP auth) | ✅ |
| Dialplans | ✅ |
| SIP Profiles | ✅ |
| Gateways | ✅ |
| Voicemail | ✅ |
| Call Center | ✅ |
| Conferences | ✅ |
| CDR | ✅ |
| IVR Menus | 🚧 |
| Ring Groups | ✅ |
| Time Conditions | 🚧 |
