#!/bin/bash
# CallSign Fail2Ban Action Script
# Reports banned IPs to the CallSign backend API
# Put this in /etc/fail2ban/action.d/callsign-report.conf

# Action configuration
CALLSIGN_API_URL="${CALLSIGN_API_URL:-http://localhost:8080/api}"
CALLSIGN_INTERNAL_KEY="${CALLSIGN_INTERNAL_KEY:-callsign-internal-key}"

# Called when an IP is banned
actionban() {
    IP="$1"
    JAIL="$2"
    FAILURES="${3:-1}"
    TIMESTAMP=$(date -Iseconds)
    
    echo "[$(date)] Banning IP: $IP from jail: $JAIL after $FAILURES failures"
    
    # Add to iptables
    iptables -I INPUT -s "$IP" -j DROP 2>/dev/null || true
    
    # Report to CallSign API
    curl -s -X POST "$CALLSIGN_API_URL/internal/fail2ban/report" \
        -H "X-Internal-Key: $CALLSIGN_INTERNAL_KEY" \
        -H "Content-Type: application/json" \
        -d "{
            \"ip\": \"$IP\",
            \"jail\": \"$JAIL\",
            \"failures\": $FAILURES,
            \"banned_at\": \"$TIMESTAMP\",
            \"action\": \"ban\"
        }" 2>/dev/null || echo "[$(date)] Warning: Failed to report ban to CallSign API"
}

# Called when an IP is unbanned
actionunban() {
    IP="$1"
    JAIL="$2"
    
    echo "[$(date)] Unbanning IP: $IP from jail: $JAIL"
    
    # Remove from iptables
    iptables -D INPUT -s "$IP" -j DROP 2>/dev/null || true
    
    # Report to CallSign API
    curl -s -X POST "$CALLSIGN_API_URL/internal/fail2ban/report" \
        -H "X-Internal-Key: $CALLSIGN_INTERNAL_KEY" \
        -H "Content-Type: application/json" \
        -d "{
            \"ip\": \"$IP\",
            \"jail\": \"$JAIL\",
            \"action\": \"unban\"
        }" 2>/dev/null || echo "[$(date)] Warning: Failed to report unban to CallSign API"
}

# Handle command line args for fail2ban integration
case "$1" in
    ban)
        actionban "$2" "$3" "$4"
        ;;
    unban)
        actionunban "$2" "$3"
        ;;
    *)
        echo "Usage: $0 {ban|unban} <ip> <jail> [failures]"
        exit 1
        ;;
esac

