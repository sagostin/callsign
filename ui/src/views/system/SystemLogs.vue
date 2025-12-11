<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Logs</h2>
      <p class="text-muted text-sm">Real-time FreeSWITCH console and live log streaming.</p>
    </div>
    <div class="header-actions">
       <div class="connection-status" :class="connected ? 'online' : 'offline'">
         <span class="dot"></span>
         {{ connected ? 'Connected' : 'Disconnected' }}
       </div>
       <button class="btn-secondary small" @click="togglePause">{{ paused ? 'Resume' : 'Pause' }}</button>
       <button class="btn-secondary small" @click="clearLogs">Clear</button>
    </div>
  </div>

  <div class="logs-container">
    <div class="filter-bar">
      <select v-model="levelFilter" class="level-filter">
        <option value="">All Levels</option>
        <option value="DEBUG">DEBUG</option>
        <option value="INFO">INFO</option>
        <option value="NOTICE">NOTICE</option>
        <option value="WARNING">WARNING</option>
        <option value="ERROR">ERROR</option>
      </select>
      <input v-model="searchFilter" type="text" class="search-input" placeholder="Filter logs...">
    </div>

    <div class="console-output" ref="consoleRef">
      <div v-for="(log, idx) in filteredLogs" :key="idx" class="line" :class="log.level?.toLowerCase()">
        <span class="timestamp">{{ formatTime(log.timestamp) }}</span>
        <span class="level-badge" :class="log.level?.toLowerCase()">{{ log.level }}</span>
        <span class="module" v-if="log.module">[{{ log.module }}]</span>
        <span class="message">{{ log.message }}</span>
      </div>
      <div v-if="logs.length === 0" class="empty-console">
        <span v-if="!connected">Connecting to FreeSWITCH...</span>
        <span v-else>Waiting for log events...</span>
      </div>
    </div>

    <div class="command-bar">
      <span class="prompt">fs&gt;</span>
      <input 
        v-model="command" 
        @keyup.enter="sendCommand"
        type="text" 
        class="command-input" 
        placeholder="Enter FreeSWITCH command (e.g. sofia status, show channels)"
        :disabled="!connected">
      <button class="btn-send" @click="sendCommand" :disabled="!connected || !command">Send</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'

const logs = ref([])
const connected = ref(false)
const paused = ref(false)
const command = ref('')
const levelFilter = ref('')
const searchFilter = ref('')
const consoleRef = ref(null)
const autoScroll = ref(true)

let ws = null
const MAX_LOGS = 1000

const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    if (levelFilter.value && log.level !== levelFilter.value) return false
    if (searchFilter.value && !log.message?.toLowerCase().includes(searchFilter.value.toLowerCase())) return false
    return true
  })
})

const connect = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = localStorage.getItem('token')
  const wsUrl = `${protocol}//${window.location.host}/api/system/console?token=${token}`
  
  ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    connected.value = true
    addLog({ type: 'status', level: 'INFO', message: 'Connected to FreeSWITCH console', timestamp: new Date().toISOString() })
  }
  
  ws.onmessage = (event) => {
    if (paused.value) return
    
    try {
      const msg = JSON.parse(event.data)
      addLog(msg)
    } catch (e) {
      addLog({ level: 'DEBUG', message: event.data, timestamp: new Date().toISOString() })
    }
  }
  
  ws.onclose = () => {
    connected.value = false
    addLog({ type: 'status', level: 'WARNING', message: 'Disconnected from FreeSWITCH console', timestamp: new Date().toISOString() })
    // Attempt reconnect after 5 seconds
    setTimeout(connect, 5000)
  }
  
  ws.onerror = (err) => {
    console.error('WebSocket error:', err)
    addLog({ type: 'error', level: 'ERROR', message: 'WebSocket connection error', timestamp: new Date().toISOString() })
  }
}

const addLog = (log) => {
  logs.value.push(log)
  if (logs.value.length > MAX_LOGS) {
    logs.value.shift()
  }
  if (autoScroll.value) {
    scrollToBottom()
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (consoleRef.value) {
      consoleRef.value.scrollTop = consoleRef.value.scrollHeight
    }
  })
}

const sendCommand = () => {
  if (!ws || !command.value || ws.readyState !== WebSocket.OPEN) return
  
  ws.send(JSON.stringify({ type: 'command', command: command.value }))
  addLog({ type: 'command', level: 'INFO', message: `> ${command.value}`, timestamp: new Date().toISOString() })
  command.value = ''
}

const togglePause = () => {
  paused.value = !paused.value
}

const clearLogs = () => {
  logs.value = []
}

const formatTime = (ts) => {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleTimeString('en-US', { hour12: false }) + '.' + String(d.getMilliseconds()).padStart(3, '0')
}

onMounted(() => {
  connect()
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
})

// Watch for new logs and handle command responses
watch(() => logs.value.length, () => {
  const lastLog = logs.value[logs.value.length - 1]
  if (lastLog?.type === 'response' && lastLog.body) {
    // Command response - show the body as separate lines
    const lines = lastLog.body.split('\n').filter(l => l.trim())
    lines.forEach(line => {
      addLog({ level: 'DEBUG', message: line, timestamp: new Date().toISOString() })
    })
  }
})
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 12px;
}
.connection-status.online { background: #dcfce7; color: #16a34a; }
.connection-status.offline { background: #fee2e2; color: #dc2626; }
.dot { width: 8px; height: 8px; border-radius: 50%; }
.connection-status.online .dot { background: #16a34a; animation: pulse 2s infinite; }
.connection-status.offline .dot { background: #dc2626; }
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.logs-container {
  background: #0d1117;
  border-radius: var(--radius-md);
  height: calc(100vh - 200px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border: 1px solid #30363d;
}

.filter-bar {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  background: #161b22;
  border-bottom: 1px solid #30363d;
}

.level-filter, .search-input {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 6px;
  padding: 6px 12px;
  color: #c9d1d9;
  font-size: 12px;
}
.search-input { flex: 1; }
.level-filter:focus, .search-input:focus { outline: none; border-color: #58a6ff; }

.console-output {
  color: #c9d1d9;
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 12px;
  overflow-y: auto;
  flex: 1;
  padding: 12px 16px;
  line-height: 1.6;
}

.line { display: flex; align-items: flex-start; gap: 8px; margin-bottom: 2px; }
.timestamp { color: #6e7681; min-width: 100px; }
.level-badge { 
  font-size: 10px; 
  font-weight: 700; 
  padding: 1px 6px; 
  border-radius: 3px; 
  text-transform: uppercase;
  min-width: 50px;
  text-align: center;
}
.level-badge.debug { background: #21262d; color: #8b949e; }
.level-badge.info { background: #1f3d5c; color: #58a6ff; }
.level-badge.notice { background: #1a3826; color: #3fb950; }
.level-badge.warning { background: #3b2e1a; color: #d29922; }
.level-badge.error { background: #492828; color: #f85149; }
.module { color: #a371f7; }
.message { color: #c9d1d9; word-break: break-all; }

.empty-console { color: #6e7681; font-style: italic; text-align: center; padding: 40px; }

.command-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #161b22;
  border-top: 1px solid #30363d;
}

.prompt { color: #3fb950; font-weight: 700; font-family: monospace; }

.command-input {
  flex: 1;
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 6px;
  padding: 8px 12px;
  color: #c9d1d9;
  font-family: 'Fira Code', monospace;
  font-size: 13px;
}
.command-input:focus { outline: none; border-color: #58a6ff; }
.command-input:disabled { opacity: 0.5; }

.btn-send {
  background: #238636;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.btn-send:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-send:hover:not(:disabled) { background: #2ea043; }

.btn-secondary {
  background: #21262d;
  border: 1px solid #30363d;
  color: #c9d1d9;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
}
.btn-secondary:hover { background: #30363d; }
</style>
