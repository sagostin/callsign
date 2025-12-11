# FreeSWITCH Lua App Scripts - Call Processing

> Documentation of FusionPBX's Lua scripts that run during live calls to handle IVR, ring groups, voicemail, and other call features.

---

## Table of Contents
1. [Overview](#overview)
2. [Script Directory Structure](#script-directory-structure)
3. [Core Call Processing Scripts](#core-call-processing-scripts)
4. [Resources Functions](#resources-functions)
5. [Database Interaction](#database-interaction)
6. [How Dialplan Triggers Scripts](#how-dialplan-triggers-scripts)

---

## Overview

FusionPBX's Lua scripts are located in:
```
/var/www/fusionpbx/app/switch/resources/scripts/app/
```

These scripts run **during live calls** via FreeSWITCH's `mod_lua`. When the dialplan executes `<action application="lua" data="app/script.lua"/>`, these scripts process the call in real-time.

### Key Difference from XML Handlers

| XML Handlers | App Scripts |
|--------------|-------------|
| Run when FreeSWITCH starts or requests config | Run during live calls |
| Generate static XML | Process real-time call flow |
| Cached heavily | Dynamic per-call |
| Examples: sofia.conf.lua, dialplan.lua | Examples: voicemail.lua, ring_groups.lua |

---

## Script Directory Structure

```
app/switch/resources/scripts/app/
├── voicemail/              # Voicemail handling
│   └── index.lua
├── ring_groups/            # Hunt group call distribution
│   └── index.lua
├── follow_me/              # Find-me/follow-me routing
│   └── index.lua
├── call_block/             # Call blocking logic
│   └── index.lua
├── caller_id/              # Caller ID manipulation
│   └── index.lua
├── conference_center/      # Conference room management
│   └── index.lua
├── conferences/            # Simple conference bridges
│   └── index.lua
├── dialplan/               # Dialplan helpers
│   └── index.lua
├── emergency/              # E911 routing
│   └── index.lua
├── emergency_notify/       # Emergency call notifications
│   └── index.lua
├── event_notify/           # Event notifications (webhooks)
│   └── index.lua
├── failure_handler/        # Call failure handling
│   └── index.lua
├── fax/                    # Fax transmission
│   └── index.lua
├── feature_event/          # BLF/feature events
│   └── index.lua
├── fifo/                   # FIFO queue handling
│   └── index.lua
├── hangup/                 # Hangup event processing
│   └── index.lua
├── is_local/               # Local extension check
│   └── index.lua
├── missed_calls/           # Missed call logging
│   └── index.lua
├── provision/              # Device provisioning
│   └── index.lua
├── servers/                # Multi-server support
│   └── index.lua
├── speed_dial/             # Speed dial handling
│   └── index.lua
├── toll_allow/             # Toll restriction checking
│   └── index.lua
├── valet_park/             # Call parking
│   └── index.lua
├── xml_handler/            # XML generation (covered separately)
│   └── index.lua
├── agent_status/           # Call center agent status
│   └── index.lua
├── avmd/                   # Answering machine detection
│   └── index.lua
└── call_control/           # In-call controls
    └── index.lua
```

---

## Core Call Processing Scripts

### Ring Groups (`ring_groups/index.lua`)

Handles hunt group call distribution with strategies like simultaneous, sequential, and random.

```lua
-- ring_groups/index.lua (simplified)

-- Get channel variables
domain_name = session:getVariable("domain_name")
ring_group_uuid = session:getVariable("ring_group_uuid")

-- Connect to database
require "resources.functions.config"
local Database = require "resources.functions.database"
dbh = Database.new('system')

-- Get ring group settings
sql = [[SELECT * FROM v_ring_groups 
        WHERE ring_group_uuid = :uuid AND ring_group_enabled = true]]
dbh:query(sql, {uuid = ring_group_uuid}, function(row)
    ring_group_strategy = row.ring_group_strategy
    ring_group_timeout = row.ring_group_timeout
    ring_group_forward_destination = row.ring_group_forward_destination
end)

-- Get ring group destinations
sql = [[SELECT * FROM v_ring_group_destinations 
        WHERE ring_group_uuid = :uuid ORDER BY destination_delay]]
local destinations = {}
dbh:query(sql, {uuid = ring_group_uuid}, function(row)
    table.insert(destinations, {
        number = row.destination_number,
        delay = row.destination_delay,
        timeout = row.destination_timeout
    })
end)

-- Build dial string based on strategy
if ring_group_strategy == "simultaneous" then
    -- Ring all at once
    local dial_string = ""
    for _, dest in ipairs(destinations) do
        dial_string = dial_string .. "," .. build_dial_string(dest)
    end
    session:execute("bridge", dial_string)
    
elseif ring_group_strategy == "sequence" then
    -- Ring one at a time
    for _, dest in ipairs(destinations) do
        session:execute("set", "call_timeout="..dest.timeout)
        session:execute("bridge", build_dial_string(dest))
        if session:getVariable("bridge_hangup_cause") == "SUCCESS" then
            break
        end
    end
end

-- Handle no answer - forward or voicemail
if not answered then
    session:execute("transfer", ring_group_forward_destination)
end
```

### Voicemail (`voicemail/index.lua`)

Handles voicemail check and record functionality.

```lua
-- voicemail/index.lua (simplified)

-- Get parameters
domain_name = session:getVariable("domain_name")
voicemail_id = session:getVariable("voicemail_id")
voicemail_action = session:getVariable("voicemail_action")

-- Connect to database
require "resources.functions.config"
local Database = require "resources.functions.database"
dbh = Database.new('system')

-- Get voicemail settings
sql = [[SELECT * FROM v_voicemails 
        WHERE voicemail_id = :id AND domain_uuid = :domain_uuid]]
dbh:query(sql, {id = voicemail_id, domain_uuid = domain_uuid}, function(row)
    voicemail_password = row.voicemail_password
    voicemail_mail_to = row.voicemail_mail_to
    voicemail_attach_file = row.voicemail_attach_file
end)

if voicemail_action == "check" then
    -- Play greeting, ask for PIN
    session:answer()
    session:streamFile("voicemail/vm-enter_pass.wav")
    local pin = session:getDigits(10, "#", 5000)
    
    if pin == voicemail_password then
        -- List and play messages
        sql = [[SELECT * FROM v_voicemail_messages 
                WHERE voicemail_uuid = :uuid ORDER BY created_epoch DESC]]
        -- ... play messages, handle deletion
    end
    
elseif voicemail_action == "record" then
    -- Record voicemail
    session:answer()
    session:sleep(500)
    session:streamFile(greeting_file)
    session:streamFile("voicemail/vm-record_message.wav")
    
    local record_file = recordings_dir.."/"..uuid()..".wav"
    session:recordFile(record_file, 180, 0, 3)
    
    -- Save to database
    sql = [[INSERT INTO v_voicemail_messages 
            (message_uuid, voicemail_uuid, message_filename, ...) 
            VALUES (:uuid, :vm_uuid, :filename, ...)]]
    dbh:query(sql, {uuid = new_uuid, vm_uuid = voicemail_uuid, filename = record_file})
    
    -- Send email notification
    if voicemail_mail_to then
        send_voicemail_email(voicemail_mail_to, record_file)
    end
end
```

### Follow Me (`follow_me/index.lua`)

Implements find-me/follow-me with sequential dialing of multiple numbers.

```lua
-- follow_me/index.lua (simplified)

-- Get settings
follow_me_uuid = session:getVariable("follow_me_uuid")

-- Get follow me destinations
sql = [[SELECT * FROM v_follow_me_destinations 
        WHERE follow_me_uuid = :uuid ORDER BY follow_me_order]]
local destinations = {}
dbh:query(sql, {uuid = follow_me_uuid}, function(row)
    table.insert(destinations, {
        destination = row.follow_me_destination,
        delay = row.follow_me_delay,
        timeout = row.follow_me_timeout,
        prompt = row.follow_me_prompt
    })
end)

-- Try each destination in order
for _, dest in ipairs(destinations) do
    -- Wait for delay
    if dest.delay > 0 then
        session:sleep(dest.delay * 1000)
    end
    
    -- Set timeout
    session:execute("set", "call_timeout="..dest.timeout)
    
    -- If prompt enabled, play "Press 1 to accept"
    if dest.prompt == "true" then
        session:execute("set", "group_confirm_key=1")
        session:execute("set", "group_confirm_file=ivr/ivr-accept_this_call.wav")
    end
    
    -- Bridge the call
    session:execute("bridge", build_dial_string(dest.destination))
    
    if session:answered() then
        break
    end
end
```

### IVR Menu (`ivr_menu.lua`)

Located at `/scripts/ivr_menu.lua` (top level), handles IVR menu logic.

```lua
-- ivr_menu.lua (simplified)

-- Get IVR UUID from channel variable
ivr_menu_uuid = session:getVariable("ivr_menu_uuid")

-- Query IVR from database
sql = [[SELECT * FROM v_ivr_menus WHERE ivr_menu_uuid = :uuid]]
dbh:query(sql, {uuid = ivr_menu_uuid}, function(row)
    ivr_menu_greet_long = row.ivr_menu_greet_long
    ivr_menu_greet_short = row.ivr_menu_greet_short
    ivr_menu_timeout = row.ivr_menu_timeout
    ivr_menu_max_failures = row.ivr_menu_max_failures
    ivr_menu_digit_len = row.ivr_menu_digit_len
    ivr_menu_direct_dial = row.ivr_menu_direct_dial
end)

-- Query IVR options
sql = [[SELECT * FROM v_ivr_menu_options 
        WHERE ivr_menu_uuid = :uuid ORDER BY ivr_menu_option_order]]
local options = {}
dbh:query(sql, {uuid = ivr_menu_uuid}, function(row)
    options[row.ivr_menu_option_digits] = {
        action = row.ivr_menu_option_action,
        param = row.ivr_menu_option_param
    }
end)

-- Answer and play greeting
session:answer()
session:sleep(500)
session:streamFile(ivr_menu_greet_long)

-- Start menu loop
local failures = 0
local timeouts = 0
while failures < ivr_menu_max_failures and timeouts < ivr_menu_max_timeouts do
    -- Collect digits
    local digits = session:getDigits(ivr_menu_digit_len, "#", ivr_menu_timeout * 1000)
    
    if digits == "" then
        -- Timeout
        timeouts = timeouts + 1
        session:streamFile(ivr_menu_greet_short)
    elseif options[digits] then
        -- Valid option found
        local opt = options[digits]
        execute_ivr_action(session, opt.action, opt.param)
        break
    elseif ivr_menu_direct_dial == "true" and is_extension(digits) then
        -- Direct dial to extension
        session:execute("transfer", digits.." XML "..domain_name)
        break
    else
        -- Invalid option
        failures = failures + 1
        session:streamFile(ivr_menu_invalid_sound)
    end
end

-- Max failures reached - exit
if failures >= ivr_menu_max_failures then
    session:streamFile(ivr_menu_exit_sound)
    execute_ivr_action(session, ivr_menu_exit_app, ivr_menu_exit_data)
end
```

---

## Resources Functions

Common helper functions used by all scripts, located in:
```
/scripts/resources/functions/
```

### Database Connection (`database.lua`)

```lua
-- resources/functions/database.lua

local Database = {}

function Database.new(database, mode)
    -- Read config from /etc/fusionpbx/config.conf
    local dsn = database_dsn()  -- e.g., "pgsql://user:pass@host/dbname"
    
    -- Create database handle
    local dbh = freeswitch.Dbh(dsn)
    
    return dbh
end

return Database
```

### Configuration (`config.lua`)

```lua
-- resources/functions/config.lua

-- Parse /etc/fusionpbx/config.conf
local config_file = "/etc/fusionpbx/config.conf"
local config = {}

for line in io.lines(config_file) do
    local key, value = line:match("([^=]+)=(.+)")
    if key and value then
        config[key:trim()] = value:trim()
    end
end

-- Database settings
database = {
    type = config["database.0.type"],
    host = config["database.0.host"],
    port = config["database.0.port"],
    name = config["database.0.name"],
    username = config["database.0.username"],
    password = config["database.0.password"]
}

-- Build DSN
function database_dsn()
    return database.type.."://"..database.username..":"..database.password
        .."@"..database.host..":"..database.port.."/"..database.name
end
```

### Cache (`cache.lua`)

```lua
-- resources/functions/cache.lua

local cache = {}
local api = freeswitch.API()

function cache.get(key)
    local result = api:execute("memcache", "get "..key)
    if result == "-ERR NOT FOUND" then
        return nil, "NOT FOUND"
    end
    return result
end

function cache.set(key, value, expire)
    expire = expire or 3600
    api:execute("memcache", "set "..key.." "..value.." "..expire)
end

function cache.delete(key)
    api:execute("memcache", "delete "..key)
end

return cache
```

---

## Database Interaction

All scripts use parameterized queries for security:

```lua
-- Correct - parameterized query
local sql = "SELECT * FROM v_extensions WHERE extension = :ext AND domain_uuid = :domain"
local params = {ext = extension, domain = domain_uuid}
dbh:query(sql, params, function(row)
    -- process row
end)

-- Helper for single value
local value = dbh:first_value(sql, params)
```

### Common Tables Accessed by Scripts

| Script | Tables Read |
|--------|-------------|
| `ring_groups` | `v_ring_groups`, `v_ring_group_destinations` |
| `voicemail` | `v_voicemails`, `v_voicemail_messages`, `v_voicemail_greetings` |
| `follow_me` | `v_follow_me`, `v_follow_me_destinations` |
| `ivr_menu` | `v_ivr_menus`, `v_ivr_menu_options` |
| `call_block` | `v_call_block` |
| `fax` | `v_fax`, `v_fax_files` |
| `conferences` | `v_conferences`, `v_conference_rooms` |

---

## How Dialplan Triggers Scripts

When you create a ring group or IVR, the dialplan entry calls the Lua script:

### Ring Group Dialplan XML

```xml
<extension name="Ring Group 500" continue="false">
    <condition field="destination_number" expression="^500$">
        <!-- Set ring group UUID -->
        <action application="set" data="ring_group_uuid=abc-123-uuid"/>
        <!-- Transfer to domain context -->
        <action application="set" data="hangup_after_bridge=true"/>
        <!-- Execute ring group script -->
        <action application="lua" data="app/ring_groups/index.lua"/>
    </condition>
</extension>
```

### IVR Menu Dialplan XML

```xml
<extension name="Main IVR" continue="false">
    <condition field="destination_number" expression="^5000$">
        <action application="answer"/>
        <action application="sleep" data="500"/>
        <action application="set" data="ivr_menu_uuid=def-456-uuid"/>
        <action application="lua" data="ivr_menu.lua"/>
    </condition>
</extension>
```

### Voicemail Dialplan XML

```xml
<extension name="Voicemail" continue="false">
    <condition field="destination_number" expression="^\*99(\d+)$">
        <action application="set" data="voicemail_id=$1"/>
        <action application="set" data="voicemail_action=record"/>
        <action application="lua" data="app/voicemail/index.lua"/>
    </condition>
</extension>

<extension name="Check Voicemail" continue="false">
    <condition field="destination_number" expression="^\*98$">
        <action application="set" data="voicemail_action=check"/>
        <action application="lua" data="app/voicemail/index.lua"/>
    </condition>
</extension>
```

---

## Summary

| Category | Scripts | Purpose |
|----------|---------|---------|
| **Call Routing** | `ring_groups`, `follow_me`, `dialplan` | Distribute calls |
| **Messaging** | `voicemail`, `fax` | Handle messaging |
| **IVR** | `ivr_menu.lua` (top-level) | Auto-attendant |
| **Conferencing** | `conferences`, `conference_center` | Audio/video bridging |
| **Security** | `call_block`, `toll_allow` | Restrict calls |
| **Monitoring** | `hangup`, `missed_calls`, `event_notify` | Track events |
| **Hardware** | `provision` | Device provisioning |
| **ACD** | `agent_status`, `fifo` | Call center |
