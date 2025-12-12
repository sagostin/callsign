#!/bin/bash
# CallSign Fail2Ban Installation Script
# Run as root on the FreeSWITCH server

set -e

echo "=== Installing Fail2Ban for CallSign ==="

# Install fail2ban if not present
if ! command -v fail2ban-client &> /dev/null; then
    echo "Installing fail2ban..."
    apt-get update
    apt-get install -y fail2ban
fi

# Copy filter
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cp "$SCRIPT_DIR/filter.d/freeswitch-callsign.conf" /etc/fail2ban/filter.d/

# Copy jail config
cp "$SCRIPT_DIR/jail.d/freeswitch-callsign.conf" /etc/fail2ban/jail.d/

# Enable fail2ban
systemctl enable fail2ban
systemctl restart fail2ban

echo "=== Fail2Ban installed and configured ==="
echo ""
echo "Testing filter against freeswitch.log..."
fail2ban-regex /var/log/freeswitch/freeswitch.log /etc/fail2ban/filter.d/freeswitch-callsign.conf --print-all-matched || true
echo ""
echo "Current status:"
fail2ban-client status freeswitch-callsign 2>/dev/null || echo "Jail will be active after restart"

echo ""
echo "=== Setup Complete ==="
echo "Ban settings: 3 failures in 5 mins = 1 hour ban"
echo "Check status: fail2ban-client status freeswitch-callsign"
echo "Unban IP:     fail2ban-client set freeswitch-callsign unbanip <IP>"
