#!/bin/bash

# ============================================================================
# CallSign - FreeSWITCH Installation Script
# ============================================================================
# Installs FreeSWITCH and configures it for CallSign PBX
# Supports Debian/Ubuntu systems
# ============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ============================================================================
# Check Prerequisites
# ============================================================================

if [ "$EUID" -ne 0 ]; then
    log_error "This script must be run as root"
    exit 1
fi

if [ ! -f /etc/os-release ]; then
    log_error "Cannot detect OS version"
    exit 1
fi

. /etc/os-release
log_info "Detected OS: $ID $VERSION_ID"

# ============================================================================
# Configuration
# ============================================================================

CALLSIGN_API_URL="${CALLSIGN_API_URL:-http://127.0.0.1:8080}"
ESL_PASSWORD="${ESL_PASSWORD:-ClueCon}"
ESL_LISTEN_IP="${ESL_LISTEN_IP:-127.0.0.1}"

echo ""
echo "Configuration:"
echo "  CallSign API URL: $CALLSIGN_API_URL"
echo "  ESL Password: $ESL_PASSWORD"
echo "  ESL Listen IP: $ESL_LISTEN_IP"
echo ""

# ============================================================================
# Install Dependencies
# ============================================================================

log_info "Installing dependencies..."

apt-get update
apt-get install -y \
    wget curl gnupg2 lsb-release \
    ca-certificates apt-transport-https \
    sox libsox-fmt-all \
    sngrep ntp memcached

log_success "Dependencies installed"

# ============================================================================
# Install FreeSWITCH
# ============================================================================

if [ -z "$SIGNALWIRE_TOKEN" ]; then
    log_warning "No SIGNALWIRE_TOKEN set - using system packages if available"
    
    # Try system repos
    if apt-cache show freeswitch >/dev/null 2>&1; then
        log_info "Installing FreeSWITCH from system repositories..."
        apt-get install -y freeswitch freeswitch-mod-commands \
            freeswitch-mod-console freeswitch-mod-sofia \
            freeswitch-mod-event-socket
    else
        log_error "FreeSWITCH not found in repositories"
        log_info ""
        log_info "To use official SignalWire packages:"
        log_info "  1. Get token from https://freeswitch.signalwire.com/"
        log_info "  2. Run: export SIGNALWIRE_TOKEN=your_token"
        log_info "  3. Re-run this script"
        exit 1
    fi
else
    log_info "Installing from SignalWire repository..."
    
    # Add SignalWire repo
    wget --http-user=signalwire --http-password=$SIGNALWIRE_TOKEN \
        -O /usr/share/keyrings/signalwire-freeswitch-repo.gpg \
        https://freeswitch.signalwire.com/repo/deb/debian-release/signalwire-freeswitch-repo.gpg

    echo "machine freeswitch.signalwire.com login signalwire password $SIGNALWIRE_TOKEN" > /etc/apt/auth.conf
    chmod 600 /etc/apt/auth.conf

    CODENAME=$(lsb_release -sc)
    echo "deb [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ $CODENAME main" > /etc/apt/sources.list.d/freeswitch.list

    apt-get update

    # Install core packages
    apt-get install -y \
        freeswitch-meta-bare freeswitch-conf-vanilla \
        freeswitch-mod-commands freeswitch-mod-console freeswitch-mod-logfile \
        freeswitch-lang-en freeswitch-mod-say-en freeswitch-sounds-en-us-callie

    # Install essential modules
    apt-get install -y \
        freeswitch-mod-event-socket freeswitch-mod-sofia freeswitch-mod-loopback \
        freeswitch-mod-conference freeswitch-mod-db freeswitch-mod-dptools \
        freeswitch-mod-expr freeswitch-mod-fifo freeswitch-mod-hash \
        freeswitch-mod-httapi freeswitch-mod-valet-parking freeswitch-mod-dialplan-xml \
        freeswitch-mod-sndfile freeswitch-mod-native-file freeswitch-mod-local-stream \
        freeswitch-mod-tone-stream freeswitch-mod-lua freeswitch-meta-mod-say \
        freeswitch-mod-xml-curl freeswitch-mod-xml-cdr freeswitch-mod-json-cdr \
        freeswitch-mod-verto freeswitch-mod-rtc freeswitch-mod-callcenter \
        freeswitch-mod-voicemail freeswitch-mod-directory \
        freeswitch-mod-flite freeswitch-mod-tts-commandline \
        freeswitch-meta-codecs freeswitch-mod-pgsql \
        freeswitch-music-default
fi

log_success "FreeSWITCH installed"

# ============================================================================
# Configure FreeSWITCH for CallSign
# ============================================================================

log_info "Configuring FreeSWITCH..."

# Backup original config
[ -d /etc/freeswitch.orig ] || cp -r /etc/freeswitch /etc/freeswitch.orig

# Configure Event Socket
cat > /etc/freeswitch/autoload_configs/event_socket.conf.xml << EOF
<configuration name="event_socket.conf" description="Socket Client">
  <settings>
    <param name="nat-map" value="false"/>
    <param name="listen-ip" value="$ESL_LISTEN_IP"/>
    <param name="listen-port" value="8021"/>
    <param name="password" value="$ESL_PASSWORD"/>
  </settings>
</configuration>
EOF

# Configure XML CURL for CallSign API
cat > /etc/freeswitch/autoload_configs/xml_curl.conf.xml << EOF
<configuration name="xml_curl.conf" description="cURL XML Gateway">
  <bindings>
    <binding name="directory">
      <param name="gateway-url" value="$CALLSIGN_API_URL/api/freeswitch/directory" bindings="directory"/>
      <param name="method" value="POST"/>
      <param name="timeout" value="10"/>
    </binding>
    <binding name="dialplan">
      <param name="gateway-url" value="$CALLSIGN_API_URL/api/freeswitch/dialplan" bindings="dialplan"/>
      <param name="method" value="POST"/>
      <param name="timeout" value="10"/>
    </binding>
    <binding name="configuration">
      <param name="gateway-url" value="$CALLSIGN_API_URL/api/freeswitch/configuration" bindings="configuration"/>
      <param name="method" value="POST"/>
      <param name="timeout" value="10"/>
    </binding>
  </bindings>
</configuration>
EOF

# Set permissions
chown -R freeswitch:freeswitch /etc/freeswitch
chmod -R 755 /etc/freeswitch

# Create storage directories
mkdir -p /var/lib/freeswitch/{recordings,voicemail,fax}
chown -R freeswitch:freeswitch /var/lib/freeswitch

log_success "FreeSWITCH configured"

# ============================================================================
# Start FreeSWITCH
# ============================================================================

log_info "Starting FreeSWITCH..."

systemctl daemon-reload
systemctl enable freeswitch
systemctl restart freeswitch

sleep 3

if systemctl is-active --quiet freeswitch; then
    log_success "FreeSWITCH is running"
else
    log_error "FreeSWITCH failed to start"
    log_info "Check logs: journalctl -u freeswitch"
    exit 1
fi

# ============================================================================
# Summary
# ============================================================================

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}FreeSWITCH Installation Complete!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo "Configuration:"
echo "  ESL Port: 8021"
echo "  ESL Password: $ESL_PASSWORD"
echo "  API URL: $CALLSIGN_API_URL"
echo ""
echo "Useful commands:"
echo "  fs_cli              - FreeSWITCH CLI"
echo "  fs_cli -x 'status'  - Check status"
echo "  journalctl -u freeswitch -f  - View logs"
echo ""
