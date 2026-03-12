# Call Flow Architecture & Destination Calculation

This document describes the call flow structure, node types, and how destinations are calculated for the CallSign PBX system. The call flows are designed to be interpreted by a FreeSWITCH mod_python3 script.

---

## Flow Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           INBOUND CALL                                       │
│                         (DID / Number)                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          INITIAL ROUTING                                     │
│            Toggle / Time Condition / Direct IVR / Direct Queue              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
              ┌─────────┐     ┌─────────┐     ┌─────────┐
              │  IVR    │     │  Queue  │     │ Direct  │
              │  Menu   │     │         │     │  Ext    │
              └─────────┘     └─────────┘     └─────────┘
                    │               │               │
        ┌───────────┼───────────┐   │               │
        │           │           │   │               │
        ▼           ▼           ▼   ▼               ▼
   ┌────────┐  ┌────────┐  ┌────────────┐     ┌──────────┐
   │ Queue  │  │ Ring   │  │ Voicemail  │     │ Complete │
   │        │  │ Group  │  │            │     │          │
   └────────┘  └────────┘  └────────────┘     └──────────┘
```

---

## JSON Flow Schema

```json
{
  "id": "ivr_main_menu",
  "name": "Main Menu",
  "extension": "8000",
  "enabled": true,
  "settings": {
    "greetLong": "main_greeting.wav",
    "greetShort": "main_greeting_short.wav",
    "timeout": 3000,
    "maxTimeouts": 3,
    "maxFailures": 3,
    "digitLength": 4,
    "interDigitTimeout": 2000,
    "directDial": false,
    "exitAction": { "type": "hangup", "value": null },
    "invalidSound": "invalid_option.wav",
    "exitSound": "goodbye.wav"
  },
  "nodes": [...],
  "connections": [...]
}
```

---

## Node Types

### Input Nodes

| Type | Description | Outputs | Config |
|------|-------------|---------|--------|
| `gather` | Collect DTMF digits | `Match`, `Timeout`, `Invalid` | `maxDigits`, `timeout`, `terminator`, `validPattern` |
| `speech` | Speech recognition input | `Match`, `No Match` | `provider`, `grammar`, `timeout`, `confidence` |

### Audio Nodes

| Type | Description | Outputs | Config |
|------|-------------|---------|--------|
| `play_audio` | Play WAV/MP3 file | `Next` | `audioFile`, `loop` |
| `play_tts` | Text-to-Speech | `Next` | `text`, `engine`, `voice` |
| `say_digits` | Speak digits/numbers | `Next` | `value`, `format` |

### Logic Nodes

| Type | Description | Outputs | Config |
|------|-------------|---------|--------|
| `web_request` | HTTP API call | `Success`, `Error`, `Timeout` | `method`, `url`, `headers`, `body`, `timeout`, `responseVar` |
| `send_sms` | Send SMS message | `Sent`, `Failed` | `provider`, `from`, `to`, `body` |
| `database` | Database query | `Success`, `No Results`, `Error` | `connection`, `query`, `timeout`, `resultVar` |
| `condition` | Conditional branch | `True`, `False` | `variable`, `operator`, `value` |
| `set_variable` | Set channel variable | `Next` | `name`, `value` |
| `subflow` | Jump to another IVR | `Return`, `End` | `menuId`, `returnAfter` |

### Destination Nodes

| Type | Description | Outputs | Config |
|------|-------------|---------|--------|
| `extension` | Transfer to extension | (terminal) | `extension` |
| `queue` | Place in ACD queue | (terminal) | `queueId` |
| `ring_group` | Ring group dial | (terminal) | `groupId` |
| `external` | External number | (terminal) | `number` |
| `voicemail` | Send to voicemail | (terminal) | `mailboxId` |
| `hangup` | Disconnect call | (terminal) | — |

---

## Connection Schema

```json
{
  "id": "conn_1234567890",
  "sourceId": "node_gather_1",
  "targetId": "node_queue_1",
  "sourceOutput": "match",
  "label": "1"
}
```

---

## Destination Calculation

### Priority Order

When calculating the final destination, the system evaluates in this order:

1. **Toggle Check** — Is there an active toggle override?
2. **Time Condition** — Does current time match any schedule?
3. **IVR Input** — User DTMF selection
4. **Fallback** — Exit action (hangup, voicemail, etc.)

### Variable Substitution

Variables can be used throughout the flow using `${variable_name}` syntax:

| Variable | Description |
|----------|-------------|
| `${caller_id}` | Caller's phone number |
| `${caller_name}` | Caller ID name (if available) |
| `${destination}` | Dialed number |
| `${uuid}` | Unique call ID |
| `${epoch}` | Unix timestamp |
| `${api_response}` | Last API response (JSON) |
| `${db_result}` | Last database result |
| `${digits}` | User input digits |

### Example: API-Driven Routing

```
1. Call comes in to +1 (555) 100-1000
2. Web Request node calls CRM API with ${caller_id}
3. API returns customer tier: { "tier": "premium" }
4. Condition node checks ${api_response.tier} == "premium"
5. If true → Route to Premium Support Queue
6. If false → Route to Standard Support Queue
```

---

## Python Script Integration

The JSON flow is interpreted by a FreeSWITCH mod_python3 script:

```python
# Pseudocode for flow execution
def execute_flow(call, flow_json):
    current_node = find_start_node(flow_json)
    
    while current_node:
        result = execute_node(call, current_node)
        
        # Find connection based on output
        connection = find_connection(
            flow_json['connections'],
            current_node['id'],
            result.output
        )
        
        if connection:
            current_node = find_node(flow_json['nodes'], connection['targetId'])
        else:
            # Terminal node or no matching connection
            current_node = None
    
    # Apply exit action if no terminal reached
    apply_exit_action(call, flow_json['settings']['exitAction'])
```

---

## Flow Validation Rules

1. **No Orphan Nodes** — Every node must be reachable from start
2. **No Cycles** — Flows must not create infinite loops (except subflow with return)
3. **Terminal Required** — Every path must end at a destination node
4. **Unique IDs** — Node and connection IDs must be unique
5. **Valid Outputs** — Connections must reference valid output types

---

## Export Format

When saving, the editor exports:

```json
{
  "version": "1.0",
  "type": "ivr_menu",
  "metadata": {
    "created": "2024-12-10T09:00:00Z",
    "modified": "2024-12-10T09:30:00Z",
    "author": "admin"
  },
  "menu": {
    "id": "...",
    "name": "...",
    "settings": { ... },
    "nodes": [ ... ],
    "connections": [ ... ]
  }
}
```

---

## Related Files

- `NodePalette.vue` — Drag-and-drop node library
- `FlowNode.vue` — Individual node component with config popup
- `FlowCanvas.vue` — Canvas for node placement and connections
- `IVRMenuForm.vue` — Main visual editor page
- `CallFlows.vue` — Read-only routing overview

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2024-12-10 | Initial schema with basic nodes |
| 1.1 | 2024-12-10 | Added Web Request, SMS, Database nodes |
| 1.2 | 2024-12-10 | Added Condition, Set Variable, Subflow |
