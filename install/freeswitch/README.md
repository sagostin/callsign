# FreeSWITCH Installation for CallSign

This directory contains the FreeSWITCH installation script and minimal configuration templates for CallSign PBX.

## Quick Install

```bash
# Set your CallSign API URL
export CALLSIGN_API_URL="http://localhost:8080"

# Optional: Set SignalWire token for official packages
export SIGNALWIRE_TOKEN="your_token_here"

# Run installer as root
sudo -E ./install.sh
```

## What it does

1. **Installs FreeSWITCH** from SignalWire packages (with token) or system repos
2. **Configures mod_xml_curl** to fetch directory/dialplan from CallSign API
3. **Configures mod_event_socket** for ESL connections
4. **Sets up required modules** for PBX functionality

## Required Modules

The installer enables these FreeSWITCH modules:

| Module | Purpose |
|--------|---------|
| mod_xml_curl | Dynamic config from API |
| mod_event_socket | ESL connection |
| mod_sofia | SIP handling |
| mod_verto | WebRTC support |
| mod_conference | Conferencing |
| mod_callcenter | Queue/ACD |
| mod_voicemail | Voicemail |
| mod_lua | Scripting (optional) |

## Manual Configuration

If you prefer manual setup, configure these files:

### /etc/freeswitch/autoload_configs/xml_curl.conf.xml
```xml
<configuration name="xml_curl.conf" description="cURL XML Gateway">
  <bindings>
    <binding name="directory">
      <param name="gateway-url" value="http://YOUR_API:8080/api/freeswitch/directory" bindings="directory"/>
    </binding>
    <binding name="dialplan">
      <param name="gateway-url" value="http://YOUR_API:8080/api/freeswitch/dialplan" bindings="dialplan"/>
    </binding>
    <binding name="configuration">
      <param name="gateway-url" value="http://YOUR_API:8080/api/freeswitch/configuration" bindings="configuration"/>
    </binding>
  </bindings>
</configuration>
```

### /etc/freeswitch/autoload_configs/event_socket.conf.xml
```xml
<configuration name="event_socket.conf" description="Socket Client">
  <settings>
    <param name="listen-ip" value="127.0.0.1"/>
    <param name="listen-port" value="8021"/>
    <param name="password" value="ClueCon"/>
  </settings>
</configuration>
```

## Verification

After installation, verify FreeSWITCH is running:

```bash
# Check service status
systemctl status freeswitch

# Connect to CLI
fs_cli

# Test ESL connection
fs_cli -x "status"
```

## Troubleshooting

```bash
# View logs
journalctl -u freeswitch -f

# Check module loading
fs_cli -x "module_exists mod_xml_curl"
fs_cli -x "module_exists mod_event_socket"

# Test API connectivity
curl http://localhost:8080/api/freeswitch/configuration
```
