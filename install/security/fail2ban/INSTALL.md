# CallSign Fail2ban Integration

## Overview

Fail2ban protects CallSign from brute-force attacks on:
- FreeSWITCH SIP registration
- API authentication
- Web server floods

## Quick Install

```bash
# Copy configuration
cp install/security/fail2ban/jail.conf /etc/fail2ban/jail.local
cp install/security/fail2ban/filter.d/* /etc/fail2ban/filter.d/

# Restart fail2ban
systemctl restart fail2ban

# Verify jails
fail2ban-client status
```

## Jails

| Jail | Log | Triggers |
|------|-----|----------|
| `freeswitch` | freeswitch.log | Auth failures |
| `freeswitch-ip` | freeswitch.log | ACL violations |
| `sip-auth-failure` | freeswitch.log | SIP auth |
| `callsign-api` | api.log | 401/403 errors |
| `nginx-dos` | access.log | Request flood |
| `ssh` | auth.log | SSH brute force |

## Commands

```bash
# Check status
fail2ban-client status freeswitch

# Unban IP
fail2ban-client set freeswitch unbanip 1.2.3.4

# Show banned
iptables -L -n | grep f2b
```

## Whitelist

Edit `/etc/fail2ban/jail.local`:
```ini
ignoreip = 127.0.0.1/8 10.0.0.0/8 YOUR_OFFICE_IP
```
