#!/bin/bash

# ============================================================================
# CallSign - Interactive Configuration Script
# ============================================================================
# This script helps you configure CallSign for deployment.
# It generates the .env file and starts Docker Compose.
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"

# ============================================================================
# Helper Functions
# ============================================================================

print_banner() {
    echo -e "${CYAN}"
    echo "  ██████╗ █████╗ ██╗     ██╗     ███████╗██╗ ██████╗ ███╗   ██╗"
    echo " ██╔════╝██╔══██╗██║     ██║     ██╔════╝██║██╔════╝ ████╗  ██║"
    echo " ██║     ███████║██║     ██║     ███████╗██║██║  ███╗██╔██╗ ██║"
    echo " ██║     ██╔══██║██║     ██║     ╚════██║██║██║   ██║██║╚██╗██║"
    echo " ╚██████╗██║  ██║███████╗███████╗███████║██║╚██████╔╝██║ ╚████║"
    echo "  ╚═════╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚═╝ ╚═════╝ ╚═╝  ╚═══╝"
    echo -e "${NC}"
    echo -e "${GREEN}CallSign PBX - Setup & Configuration Script${NC}"
    echo ""
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

prompt() {
    local message="$1"
    local default="$2"
    local var_name="$3"
    
    if [ -n "$default" ]; then
        read -p "$message [$default]: " value
        value="${value:-$default}"
    else
        read -p "$message: " value
    fi
    
    eval "$var_name='$value'"
}

prompt_password() {
    local message="$1"
    local var_name="$2"
    
    read -s -p "$message: " value
    echo ""
    eval "$var_name='$value'"
}

prompt_yes_no() {
    local message="$1"
    local default="$2"
    local var_name="$3"
    
    while true; do
        read -p "$message (y/n) [$default]: " value
        value="${value:-$default}"
        case $value in
            [Yy]* ) eval "$var_name=true"; break;;
            [Nn]* ) eval "$var_name=false"; break;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

generate_random_string() {
    openssl rand -hex "$1" 2>/dev/null || head -c "$1" /dev/urandom | xxd -p | head -c $(($1 * 2))
}

# ============================================================================
# Configuration Collection
# ============================================================================

collect_domain_config() {
    echo ""
    echo -e "${CYAN}=== Domain Configuration ===${NC}"
    echo ""
    
    prompt "Enter your domain (use 'localhost' for local development)" "localhost" DOMAIN
    
    if [ "$DOMAIN" != "localhost" ]; then
        prompt_yes_no "Enable SSL with Let's Encrypt?" "y" SSL_ENABLED
    else
        SSL_ENABLED=false
        log_info "SSL disabled for localhost"
    fi
}

collect_database_config() {
    echo ""
    echo -e "${CYAN}=== Database Configuration ===${NC}"
    echo ""
    
    prompt "PostgreSQL username" "callsign" POSTGRES_USER
    
    log_info "Generating secure PostgreSQL password..."
    POSTGRES_PASSWORD=$(generate_random_string 16)
    log_success "Password generated"
    
    prompt "PostgreSQL database name" "callsign" POSTGRES_DB
}

collect_security_config() {
    echo ""
    echo -e "${CYAN}=== Security Configuration ===${NC}"
    echo ""
    
    log_info "Generating JWT secret..."
    JWT_SECRET=$(generate_random_string 32)
    log_success "JWT secret generated"
    
    log_info "Generating encryption key..."
    ENCRYPTION_KEY=$(generate_random_string 32)
    log_success "Encryption key generated"
}

collect_freeswitch_config() {
    echo ""
    echo -e "${CYAN}=== FreeSWITCH Configuration ===${NC}"
    echo ""
    
    prompt_yes_no "Is FreeSWITCH running on this host?" "y" FS_LOCAL
    
    if [ "$FS_LOCAL" = "true" ]; then
        FREESWITCH_HOST="host.docker.internal"
    else
        prompt "FreeSWITCH host IP/hostname" "127.0.0.1" FREESWITCH_HOST
    fi
    
    prompt "FreeSWITCH ESL port" "8021" FREESWITCH_ESL_PORT
    prompt "FreeSWITCH ESL password" "ClueCon" FREESWITCH_ESL_PASSWORD
}

collect_email_config() {
    echo ""
    echo -e "${CYAN}=== Email Configuration (Optional) ===${NC}"
    echo ""
    
    prompt_yes_no "Configure SMTP for email notifications?" "n" CONFIGURE_EMAIL
    
    if [ "$CONFIGURE_EMAIL" = "true" ]; then
        prompt "SMTP host" "smtp.gmail.com" SMTP_HOST
        prompt "SMTP port" "587" SMTP_PORT
        prompt "SMTP username" "" SMTP_USER
        prompt_password "SMTP password" SMTP_PASSWORD
        prompt "From email address" "noreply@$DOMAIN" SMTP_FROM
    else
        SMTP_HOST="smtp.example.com"
        SMTP_PORT="587"
        SMTP_USER=""
        SMTP_PASSWORD=""
        SMTP_FROM="noreply@example.com"
    fi
}

collect_transcription_config() {
    echo ""
    echo -e "${CYAN}=== Transcription Configuration (Optional) ===${NC}"
    echo ""
    
    prompt_yes_no "Enable voicemail/recording transcription?" "n" ENABLE_TRANSCRIPTION
    
    if [ "$ENABLE_TRANSCRIPTION" = "true" ]; then
        echo ""
        echo "Available providers:"
        echo "  1) whisper  - Local Whisper or OpenAI API"
        echo "  2) deepgram - Deepgram (cloud)"
        echo "  3) assemblyai - AssemblyAI (cloud)"
        echo "  4) custom - Custom API endpoint"
        echo ""
        prompt "Select provider (1-4)" "1" TRANSCRIPTION_CHOICE
        
        case $TRANSCRIPTION_CHOICE in
            1) TRANSCRIPTION_PROVIDER="whisper";;
            2) TRANSCRIPTION_PROVIDER="deepgram"
               prompt "Deepgram API key" "" DEEPGRAM_API_KEY;;
            3) TRANSCRIPTION_PROVIDER="assemblyai"
               prompt "AssemblyAI API key" "" ASSEMBLYAI_API_KEY;;
            4) TRANSCRIPTION_PROVIDER="custom"
               prompt "Custom transcription API URL" "" CUSTOM_TRANSCRIPTION_URL;;
            *) TRANSCRIPTION_PROVIDER="whisper";;
        esac
    else
        TRANSCRIPTION_PROVIDER="whisper"
    fi
}

collect_tts_config() {
    echo ""
    echo -e "${CYAN}=== Text-to-Speech Configuration (Optional) ===${NC}"
    echo ""
    
    prompt_yes_no "Enable Text-to-Speech for IVR/announcements?" "n" ENABLE_TTS
    
    if [ "$ENABLE_TTS" = "true" ]; then
        echo ""
        echo "Available providers:"
        echo "  1) edge    - Microsoft Edge TTS (free)"
        echo "  2) openai  - OpenAI TTS"
        echo "  3) elevenlabs - ElevenLabs"
        echo "  4) piper   - Piper TTS (local)"
        echo ""
        prompt "Select provider (1-4)" "1" TTS_CHOICE
        
        case $TTS_CHOICE in
            1) TTS_PROVIDER="edge";;
            2) TTS_PROVIDER="openai"
               prompt "OpenAI API key" "" OPENAI_API_KEY;;
            3) TTS_PROVIDER="elevenlabs"
               prompt "ElevenLabs API key" "" ELEVENLABS_API_KEY;;
            4) TTS_PROVIDER="piper";;
            *) TTS_PROVIDER="edge";;
        esac
    else
        TTS_PROVIDER="edge"
    fi
}

# ============================================================================
# Generate Configuration Files
# ============================================================================

generate_env_file() {
    log_info "Generating .env file..."
    
    cat > "$ENV_FILE" << EOF
# ============================================================================
# CallSign Environment Configuration
# Generated by configure.sh on $(date)
# ============================================================================

# ===================================
# Domain & SSL
# ===================================
DOMAIN=$DOMAIN
SSL_ENABLED=$SSL_ENABLED
ADMIN_PANEL_URL=http${SSL_ENABLED:+s}://$DOMAIN/admin

# ===================================
# Database (PostgreSQL)
# ===================================
POSTGRES_USER=$POSTGRES_USER
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_DB=$POSTGRES_DB
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# ===================================
# Redis
# ===================================
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# ===================================
# ClickHouse (Analytics)
# ===================================
CLICKHOUSE_HOST=clickhouse
CLICKHOUSE_PORT=9000
CLICKHOUSE_HTTP_PORT=8123
CLICKHOUSE_DB=callsign
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=

# ===================================
# API Server
# ===================================
API_PORT=8080
API_HOST=0.0.0.0
API_ENV=production
API_DEBUG=false

# JWT Settings
JWT_SECRET=$JWT_SECRET
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# ===================================
# FreeSWITCH
# ===================================
FREESWITCH_HOST=$FREESWITCH_HOST
FREESWITCH_ESL_PORT=$FREESWITCH_ESL_PORT
FREESWITCH_ESL_PASSWORD=$FREESWITCH_ESL_PASSWORD
FREESWITCH_XML_CURL_URL=http://api:8080/api/freeswitch

# ===================================
# Encryption
# ===================================
ENCRYPTION_KEY=$ENCRYPTION_KEY

# ===================================
# Monitoring
# ===================================
GRAFANA_USER=admin
GRAFANA_PASSWORD=$(generate_random_string 12)
LOKI_URL=http://loki:3100

# ===================================
# CORS
# ===================================
CORS_ORIGINS=http://$DOMAIN,https://$DOMAIN

# ===================================
# File Storage
# ===================================
STORAGE_PATH=/var/lib/callsign
RECORDINGS_PATH=/var/lib/callsign/recordings
VOICEMAIL_PATH=/var/lib/callsign/voicemail
FAX_PATH=/var/lib/callsign/fax
TRANSCRIPTIONS_PATH=/var/lib/callsign/transcriptions

# ===================================
# Email (SMTP)
# ===================================
SMTP_HOST=$SMTP_HOST
SMTP_PORT=$SMTP_PORT
SMTP_USER=$SMTP_USER
SMTP_PASSWORD=$SMTP_PASSWORD
SMTP_FROM=$SMTP_FROM
SMTP_TLS=true

# ===================================
# Transcription
# ===================================
ENABLE_TRANSCRIPTION=$ENABLE_TRANSCRIPTION
TRANSCRIPTION_PROVIDER=$TRANSCRIPTION_PROVIDER
TRANSCRIPTION_MODE=simple
WHISPER_MODEL=base
DEEPGRAM_API_KEY=${DEEPGRAM_API_KEY:-}
ASSEMBLYAI_API_KEY=${ASSEMBLYAI_API_KEY:-}
CUSTOM_TRANSCRIPTION_URL=${CUSTOM_TRANSCRIPTION_URL:-}

# ===================================
# Text-to-Speech
# ===================================
ENABLE_TTS=$ENABLE_TTS
TTS_PROVIDER=$TTS_PROVIDER
TTS_DEFAULT_VOICE=en-US-JennyNeural
TTS_DEFAULT_LANGUAGE=en-US
OPENAI_API_KEY=${OPENAI_API_KEY:-}
ELEVENLABS_API_KEY=${ELEVENLABS_API_KEY:-}

# ===================================
# Feature Flags
# ===================================
ENABLE_FAX=true
ENABLE_SMS=true
ENABLE_WEBRTC=true
ENABLE_HOSPITALITY=false
ENABLE_CALL_RECORDING=true
ENABLE_VOICEMAIL_TRANSCRIPTION=${ENABLE_TRANSCRIPTION}
EOF

    log_success ".env file generated"
}

generate_caddyfile() {
    log_info "Generating Caddyfile..."
    
    mkdir -p "$SCRIPT_DIR/docker/caddy"
    
    if [ "$SSL_ENABLED" = "true" ]; then
        cat > "$SCRIPT_DIR/docker/caddy/Caddyfile" << EOF
# CallSign Caddyfile - SSL Enabled
{
    email admin@$DOMAIN
}

$DOMAIN {
    # API routes
    handle /api/* {
        reverse_proxy api:8080
    }
    
    # WebSocket
    handle /ws/* {
        reverse_proxy api:8080
    }
    
    # FreeSWITCH XML handler
    handle /freeswitch/* {
        reverse_proxy api:8080
    }
    
    # UI (default)
    handle {
        reverse_proxy ui:80
    }
    
    # Security headers
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
        Referrer-Policy strict-origin-when-cross-origin
    }
    
    # Logging
    log {
        output file /var/log/caddy/access.log
        format json
    }
}
EOF
    else
        cat > "$SCRIPT_DIR/docker/caddy/Caddyfile" << EOF
# CallSign Caddyfile - No SSL (Development)
{
    auto_https off
    admin off
}

:80 {
    # API routes
    handle /api/* {
        reverse_proxy api:8080
    }
    
    # WebSocket
    handle /ws/* {
        reverse_proxy api:8080
    }
    
    # FreeSWITCH XML handler  
    handle /freeswitch/* {
        reverse_proxy api:8080
    }
    
    # UI (default)
    handle {
        reverse_proxy ui:80
    }
    
    # Security headers
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options SAMEORIGIN
        X-XSS-Protection "1; mode=block"
    }
    
    # Logging
    log {
        output stdout
        format console
    }
}
EOF
    fi
    
    log_success "Caddyfile generated"
}

# ============================================================================
# Docker Compose Update
# ============================================================================

update_docker_compose() {
    log_info "Updating docker-compose.yml with Caddy..."
    
    cat > "$SCRIPT_DIR/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  # ========================
  # Reverse Proxy (Caddy)
  # ========================
  caddy:
    image: caddy:2-alpine
    container_name: callsign-caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./docker/caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - api
      - ui
    networks:
      - callsign-network

  # ========================
  # API Server
  # ========================
  api:
    build:
      context: ./api
      dockerfile: Dockerfile
    container_name: callsign-api
    restart: unless-stopped
    environment:
      - POSTGRES_HOST=${POSTGRES_HOST:-postgres}
      - POSTGRES_PORT=${POSTGRES_PORT:-5432}
      - POSTGRES_USER=${POSTGRES_USER:-callsign}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB:-callsign}
      - REDIS_HOST=${REDIS_HOST:-redis}
      - REDIS_PORT=${REDIS_PORT:-6379}
      - JWT_SECRET=${JWT_SECRET}
      - ENCRYPTION_KEY=${ENCRYPTION_KEY}
      - FREESWITCH_HOST=${FREESWITCH_HOST:-host.docker.internal}
      - FREESWITCH_ESL_PORT=${FREESWITCH_ESL_PORT:-8021}
      - FREESWITCH_ESL_PASSWORD=${FREESWITCH_ESL_PASSWORD:-ClueCon}
      - API_ENV=${API_ENV:-production}
      - ENABLE_TRANSCRIPTION=${ENABLE_TRANSCRIPTION:-false}
      - TRANSCRIPTION_PROVIDER=${TRANSCRIPTION_PROVIDER:-whisper}
      - ENABLE_TTS=${ENABLE_TTS:-false}
      - TTS_PROVIDER=${TTS_PROVIDER:-edge}
    volumes:
      - storage_data:/var/lib/callsign
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - callsign-network
    extra_hosts:
      - "host.docker.internal:host-gateway"

  # ========================
  # UI (Vue.js)
  # ========================
  ui:
    build:
      context: ./ui
      dockerfile: Dockerfile
    container_name: callsign-ui
    restart: unless-stopped
    networks:
      - callsign-network

  # ========================
  # PostgreSQL Database
  # ========================
  postgres:
    image: postgres:15-alpine
    container_name: callsign-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_USER=${POSTGRES_USER:-callsign}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB:-callsign}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-callsign}"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - callsign-network

  # ========================
  # Redis
  # ========================
  redis:
    image: redis:7-alpine
    container_name: callsign-redis
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - callsign-network

  # ========================
  # ClickHouse (Analytics)
  # ========================
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    container_name: callsign-clickhouse
    restart: unless-stopped
    environment:
      - CLICKHOUSE_DB=${CLICKHOUSE_DB:-callsign}
      - CLICKHOUSE_USER=${CLICKHOUSE_USER:-default}
      - CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD:-}
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    networks:
      - callsign-network

  # ========================
  # Loki (Log Aggregation)
  # ========================
  loki:
    image: grafana/loki:2.9.0
    container_name: callsign-loki
    restart: unless-stopped
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki_data:/loki
    networks:
      - callsign-network

  # ========================
  # Grafana (Monitoring)
  # ========================
  grafana:
    image: grafana/grafana:latest
    container_name: callsign-grafana
    restart: unless-stopped
    environment:
      - GF_SECURITY_ADMIN_USER=${GRAFANA_USER:-admin}
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "3000:3000"
    depends_on:
      - loki
    networks:
      - callsign-network

networks:
  callsign-network:
    driver: bridge

volumes:
  postgres_data:
  redis_data:
  clickhouse_data:
  loki_data:
  grafana_data:
  storage_data:
  caddy_data:
  caddy_config:
EOF

    log_success "docker-compose.yml updated with Caddy"
}

# ============================================================================
# FreeSWITCH Setup Script
# ============================================================================

generate_freeswitch_install_script() {
    log_info "Generating FreeSWITCH installation script..."
    
    mkdir -p "$SCRIPT_DIR/install/freeswitch"
    
    cat > "$SCRIPT_DIR/install/freeswitch/install.sh" << 'FSEOF'
#!/bin/bash

# ============================================================================
# CallSign - FreeSWITCH Installation Script
# ============================================================================
# This script installs FreeSWITCH on Debian/Ubuntu systems
# and configures it for use with CallSign PBX.
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

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "This script must be run as root"
    exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION=$VERSION_ID
else
    log_error "Cannot detect OS version"
    exit 1
fi

log_info "Detected OS: $OS $VERSION"

# ============================================================================
# Configuration
# ============================================================================

CALLSIGN_API_URL="${CALLSIGN_API_URL:-http://localhost:8080}"
ESL_PASSWORD="${ESL_PASSWORD:-ClueCon}"

# ============================================================================
# Install Dependencies
# ============================================================================

log_info "Installing dependencies..."

apt-get update
apt-get install -y \
    wget curl gnupg2 lsb-release \
    ca-certificates apt-transport-https \
    sox libsox-fmt-all \
    sngrep ntp \
    lua5.3 liblua5.3-dev \
    libcurl4-openssl-dev \
    memcached

log_success "Dependencies installed"

# ============================================================================
# Install FreeSWITCH from Packages
# ============================================================================

log_info "Adding FreeSWITCH repository..."

# Check for SignalWire token
if [ -z "$SIGNALWIRE_TOKEN" ]; then
    log_warning "No SIGNALWIRE_TOKEN provided. Using community packages."
    log_info "For official packages, set SIGNALWIRE_TOKEN environment variable"
    
    # Use community/alternative sources
    # Add FreeSWITCH from Debian/Ubuntu repos if available
    apt-get install -y software-properties-common
    
    # Try installing from default repos first
    if apt-cache show freeswitch > /dev/null 2>&1; then
        log_info "Installing FreeSWITCH from system repositories..."
        apt-get install -y freeswitch freeswitch-mod-commands freeswitch-mod-console
    else
        log_error "FreeSWITCH not found in repositories."
        log_info "Please provide SIGNALWIRE_TOKEN for official packages"
        log_info "Get a token from: https://freeswitch.signalwire.com/"
        exit 1
    fi
else
    log_info "Using SignalWire repository with provided token..."
    
    # Download signing key
    wget --http-user=signalwire --http-password=$SIGNALWIRE_TOKEN \
        -O /usr/share/keyrings/signalwire-freeswitch-repo.gpg \
        https://freeswitch.signalwire.com/repo/deb/debian-release/signalwire-freeswitch-repo.gpg
    
    # Configure authentication
    echo "machine freeswitch.signalwire.com login signalwire password $SIGNALWIRE_TOKEN" > /etc/apt/auth.conf
    chmod 600 /etc/apt/auth.conf
    
    # Add repository
    CODENAME=$(lsb_release -sc)
    echo "deb [signed-by=/usr/share/keyrings/signalwire-freeswitch-repo.gpg] https://freeswitch.signalwire.com/repo/deb/debian-release/ $CODENAME main" > /etc/apt/sources.list.d/freeswitch.list
    
    apt-get update
    
    # Install FreeSWITCH packages
    log_info "Installing FreeSWITCH packages..."
    
    apt-get install -y \
        freeswitch-meta-bare \
        freeswitch-conf-vanilla \
        freeswitch-mod-commands \
        freeswitch-mod-console \
        freeswitch-mod-logfile \
        freeswitch-lang-en \
        freeswitch-mod-say-en \
        freeswitch-sounds-en-us-callie \
        freeswitch-mod-enum \
        freeswitch-mod-cdr-csv \
        freeswitch-mod-event-socket \
        freeswitch-mod-sofia \
        freeswitch-mod-loopback \
        freeswitch-mod-conference \
        freeswitch-mod-db \
        freeswitch-mod-dptools \
        freeswitch-mod-expr \
        freeswitch-mod-fifo \
        freeswitch-mod-httapi \
        freeswitch-mod-hash \
        freeswitch-mod-esl \
        freeswitch-mod-esf \
        freeswitch-mod-fsv \
        freeswitch-mod-valet-parking \
        freeswitch-mod-dialplan-xml \
        freeswitch-mod-sndfile \
        freeswitch-mod-native-file \
        freeswitch-mod-local-stream \
        freeswitch-mod-tone-stream \
        freeswitch-mod-lua \
        freeswitch-meta-mod-say \
        freeswitch-mod-xml-cdr \
        freeswitch-mod-xml-curl \
        freeswitch-mod-verto \
        freeswitch-mod-callcenter \
        freeswitch-mod-rtc \
        freeswitch-mod-png \
        freeswitch-mod-json-cdr \
        freeswitch-mod-shout \
        freeswitch-mod-sms \
        freeswitch-mod-cidlookup \
        freeswitch-mod-memcache \
        freeswitch-mod-tts-commandline \
        freeswitch-mod-directory \
        freeswitch-mod-av \
        freeswitch-mod-flite \
        freeswitch-mod-distributor \
        freeswitch-meta-codecs \
        freeswitch-mod-pgsql \
        freeswitch-music-default
fi

log_success "FreeSWITCH installed"

# ============================================================================
# Configure FreeSWITCH for CallSign
# ============================================================================

log_info "Configuring FreeSWITCH for CallSign..."

# Backup original config
cp -r /etc/freeswitch /etc/freeswitch.orig

# Configure ESL (Event Socket)
cat > /etc/freeswitch/autoload_configs/event_socket.conf.xml << EOF
<configuration name="event_socket.conf" description="Socket Client">
  <settings>
    <param name="nat-map" value="false"/>
    <param name="listen-ip" value="0.0.0.0"/>
    <param name="listen-port" value="8021"/>
    <param name="password" value="$ESL_PASSWORD"/>
  </settings>
</configuration>
EOF

# Configure XML CURL for dynamic config
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

# Enable required modules
cat > /etc/freeswitch/autoload_configs/modules.conf.xml << 'EOF'
<configuration name="modules.conf" description="Modules">
  <modules>
    <!-- Loggers -->
    <load module="mod_console"/>
    <load module="mod_logfile"/>
    
    <!-- XML Interfaces -->
    <load module="mod_xml_curl"/>
    
    <!-- Event Handlers -->
    <load module="mod_event_socket"/>
    <load module="mod_cdr_csv"/>
    <load module="mod_json_cdr"/>
    
    <!-- Endpoints -->
    <load module="mod_sofia"/>
    <load module="mod_loopback"/>
    <load module="mod_verto"/>
    <load module="mod_rtc"/>
    
    <!-- Applications -->
    <load module="mod_commands"/>
    <load module="mod_conference"/>
    <load module="mod_db"/>
    <load module="mod_dptools"/>
    <load module="mod_expr"/>
    <load module="mod_fifo"/>
    <load module="mod_hash"/>
    <load module="mod_httapi"/>
    <load module="mod_valet_parking"/>
    <load module="mod_callcenter"/>
    <load module="mod_voicemail"/>
    <load module="mod_directory"/>
    <load module="mod_distributor"/>
    <load module="mod_esf"/>
    <load module="mod_fsv"/>
    
    <!-- Dialplan Interfaces -->
    <load module="mod_dialplan_xml"/>
    <load module="mod_enum"/>
    
    <!-- Codec Interfaces -->
    <load module="mod_g723_1"/>
    <load module="mod_g729"/>
    <load module="mod_amr"/>
    <load module="mod_opus"/>
    
    <!-- File Format Interfaces -->
    <load module="mod_sndfile"/>
    <load module="mod_native_file"/>
    <load module="mod_png"/>
    <load module="mod_shout"/>
    <load module="mod_local_stream"/>
    <load module="mod_tone_stream"/>
    
    <!-- Languages -->
    <load module="mod_spandsp"/>
    
    <!-- Say -->
    <load module="mod_say_en"/>
    
    <!-- TTS -->
    <load module="mod_flite"/>
    <load module="mod_tts_commandline"/>
    
    <!-- Scripting -->
    <load module="mod_lua"/>
    
    <!-- Databases -->
    <load module="mod_pgsql"/>
  </modules>
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
# Enable and Start FreeSWITCH
# ============================================================================

log_info "Starting FreeSWITCH service..."

systemctl daemon-reload
systemctl enable freeswitch
systemctl start freeswitch

# Wait for startup
sleep 3

if systemctl is-active --quiet freeswitch; then
    log_success "FreeSWITCH is running"
else
    log_error "FreeSWITCH failed to start. Check: journalctl -u freeswitch"
    exit 1
fi

# ============================================================================
# Print Summary
# ============================================================================

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}FreeSWITCH Installation Complete!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo "Configuration:"
echo "  ESL Port: 8021"
echo "  ESL Password: $ESL_PASSWORD"
echo "  Config Dir: /etc/freeswitch"
echo "  Storage: /var/lib/freeswitch"
echo ""
echo "Useful commands:"
echo "  fs_cli                    - FreeSWITCH CLI"
echo "  systemctl status freeswitch"
echo "  journalctl -u freeswitch -f"
echo ""
FSEOF

    chmod +x "$SCRIPT_DIR/install/freeswitch/install.sh"
    log_success "FreeSWITCH installation script generated"
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    clear
    print_banner
    
    echo "This script will configure CallSign PBX for deployment."
    echo ""
    echo -e "${YELLOW}Prerequisites:${NC}"
    echo "  • Docker and Docker Compose installed"
    echo "  • FreeSWITCH installed (or will be installed separately)"
    echo ""
    
    read -p "Press Enter to continue or Ctrl+C to cancel..."
    
    # Collect configuration
    collect_domain_config
    collect_database_config
    collect_security_config
    collect_freeswitch_config
    collect_email_config
    collect_transcription_config
    collect_tts_config
    
    # Generate files
    echo ""
    echo -e "${CYAN}=== Generating Configuration Files ===${NC}"
    echo ""
    
    generate_env_file
    generate_caddyfile
    update_docker_compose
    generate_freeswitch_install_script
    
    # Summary
    echo ""
    echo -e "${GREEN}============================================${NC}"
    echo -e "${GREEN}Configuration Complete!${NC}"
    echo -e "${GREEN}============================================${NC}"
    echo ""
    echo "Generated files:"
    echo "  • .env                           - Environment configuration"
    echo "  • docker/caddy/Caddyfile        - Caddy reverse proxy config"
    echo "  • docker-compose.yml            - Docker services"
    echo "  • install/freeswitch/install.sh - FreeSWITCH installer"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Install FreeSWITCH (if not already installed):"
    echo "     sudo ./install/freeswitch/install.sh"
    echo ""
    echo "  2. Start CallSign services:"
    echo "     docker compose up -d"
    echo ""
    echo "  3. Access the admin panel:"
    if [ "$SSL_ENABLED" = "true" ]; then
        echo "     https://$DOMAIN/admin"
    else
        echo "     http://$DOMAIN/admin"
    fi
    echo ""
    echo "  4. Default admin credentials:"
    echo "     Email: admin@$DOMAIN"
    echo "     Password: admin (change immediately!)"
    echo ""
    
    # Ask to start
    prompt_yes_no "Start CallSign now?" "y" START_NOW
    
    if [ "$START_NOW" = "true" ]; then
        log_info "Building and starting services..."
        docker compose up -d --build
        
        echo ""
        log_success "CallSign is starting!"
        echo ""
        echo "View logs: docker compose logs -f"
        echo ""
    fi
}

# Run main function
main "$@"
