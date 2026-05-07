<template>
  <div class="operator-panel">
    <!-- View Header -->
    <div class="view-header">
      <div class="header-content">
        <h2>Operator Panel</h2>
        <p class="text-muted text-sm">Real-time view of active extensions, calls, and queues.</p>
      </div>
      <div class="header-actions">
        <div class="last-updated" :class="{ stale: isStale }">
          <ClockIcon class="icon-sm" />
          {{ lastUpdatedText }}
        </div>
        <div class="status-indicator">
          <span class="dot pulse"></span>
          Live
        </div>
      </div>
    </div>

    <!-- Active Calls Panel -->
    <div class="section-card" v-if="parsedActiveCalls.length > 0">
      <div class="section-header">
        <PhoneIcon class="icon" />
        <h3>Active Calls</h3>
        <span class="badge-count">{{ parsedActiveCalls.length }}</span>
      </div>
      <div class="calls-table-wrapper">
        <table class="calls-table">
          <thead>
            <tr>
              <th>Direction</th>
              <th>Caller ID</th>
              <th>Destination</th>
              <th>Duration</th>
              <th>State</th>
              <th class="actions-col">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="call in parsedActiveCalls" :key="call.uuid">
              <td>
                <span class="direction-badge" :class="call.direction">
                  <ArrowUpRightIcon v-if="call.direction === 'outbound'" class="icon-xs" />
                  <ArrowDownLeftIcon v-else class="icon-xs" />
                  {{ call.directionLabel }}
                </span>
              </td>
              <td>
                <div class="caller-info">
                  <strong>{{ call.callerName || 'Unknown' }}</strong>
                  <span class="mono">{{ call.callerNumber }}</span>
                </div>
              </td>
              <td class="mono">{{ call.destination }}</td>
              <td class="duration-cell">{{ call.duration }}</td>
              <td>
                <StatusBadge :status="call.state" />
              </td>
              <td class="actions-col">
                <button class="btn-icon-sm danger" title="Hangup" @click="hangupCall(call.uuid)">
                  <PhoneOffIcon class="icon-xs" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Queues Panel -->
    <div class="section-card" v-if="parsedQueues.length > 0">
      <div class="section-header">
        <ListIcon class="icon" />
        <h3>Queues</h3>
        <span class="badge-count">{{ parsedQueues.length }}</span>
      </div>
      <div class="queues-grid">
        <div v-for="queue in parsedQueues" :key="queue.id || queue.name" class="queue-stat-card" :class="{ 'has-waiting': queue.waiting > 0 }">
          <div class="queue-stat-header">
            <h4>{{ queue.name }}</h4>
            <span class="queue-ext">Ext. {{ queue.extension }}</span>
          </div>
          <div class="queue-stat-body">
            <div class="queue-stat-item">
              <span class="queue-stat-value" :class="{ alert: queue.waiting > 0 }">{{ queue.waiting }}</span>
              <span class="queue-stat-label">Waiting</span>
            </div>
            <div class="queue-stat-item">
              <span class="queue-stat-value">{{ queue.avgWait }}</span>
              <span class="queue-stat-label">Avg Wait</span>
            </div>
            <div class="queue-stat-item">
              <span class="queue-stat-value">{{ queue.agents }}</span>
              <span class="queue-stat-label">Agents</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Extensions Header with Search -->
    <div class="section-header extensions-header">
      <div class="extensions-title">
        <UsersIcon class="icon" />
        <h3>Extensions</h3>
        <span class="badge-count">{{ filteredExtensions.length }}</span>
      </div>
      <div class="search-bar">
        <SearchIcon class="icon-sm search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by name, number, or status..."
          class="search-input"
        />
        <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
          <XIcon class="icon-xs" />
        </button>
      </div>
      <div class="filter-pills">
        <button
          v-for="filter in statusFilters"
          :key="filter.value"
          class="filter-pill"
          :class="{ active: activeFilter === filter.value }"
          @click="activeFilter = filter.value"
        >
          {{ filter.label }}
        </button>
      </div>
    </div>

    <!-- Extensions Grid -->
    <div class="panel-grid">
      <div
        v-for="ext in filteredExtensions"
        :key="ext.id"
        class="ext-card"
        :class="[ext.status.toLowerCase(), { 'has-active-call': ext.activeCall }]"
      >
        <div class="ext-header">
          <span class="ext-number">{{ ext.number }}</span>
          <StatusBadge :status="ext.status" :show-dot="true" />
        </div>
        <div class="ext-name">{{ ext.name }}</div>
        <div class="ext-call-info" v-if="ext.activeCall">
          <PhoneIcon class="icon-xs" />
          <span class="call-detail">{{ ext.activeCall.callerNumber }} → {{ ext.activeCall.destination }}</span>
          <span class="call-duration">{{ ext.activeCall.duration }}</span>
        </div>
        <div class="ext-actions">
          <button class="btn-icon-sm" title="Dial" @click="openDialModal(ext)">
            <PhoneIcon class="icon-xs" />
          </button>
          <button
            v-if="ext.activeCall"
            class="btn-icon-sm danger"
            title="Hangup"
            @click="hangupCall(ext.activeCall.uuid)"
          >
            <PhoneOffIcon class="icon-xs" />
          </button>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredExtensions.length === 0" class="empty-state">
      <UsersIcon class="icon-lg" />
      <p>No extensions match your search.</p>
    </div>

    <!-- Dial Modal -->
    <div class="modal-overlay" v-if="dialModalOpen" @click.self="dialModalOpen = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Click-to-Dial</h3>
          <button class="close-btn" @click="dialModalOpen = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>From Extension</label>
            <input :value="dialForm.fromExtension" class="input-field" readonly />
          </div>
          <div class="form-group">
            <label>Destination Number</label>
            <input v-model="dialForm.toNumber" class="input-field" placeholder="Enter number..." @keyup.enter="executeDial" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="dialModalOpen = false">Cancel</button>
          <button class="btn-primary" @click="executeDial" :disabled="!dialForm.toNumber">
            <PhoneIcon class="icon-xs" /> Dial
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, inject } from 'vue'
import { operatorPanelAPI, liveAPI } from '../../services/api'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { formatTime } from '../../utils/formatters'
import {
  Phone as PhoneIcon,
  PhoneOff as PhoneOffIcon,
  ArrowUpRight as ArrowUpRightIcon,
  ArrowDownLeft as ArrowDownLeftIcon,
  List as ListIcon,
  Users as UsersIcon,
  Search as SearchIcon,
  X as XIcon,
  Clock as ClockIcon
} from 'lucide-vue-next'

const toast = inject('toast')

const extensions = ref([])
const activeCalls = ref([])
const queues = ref([])
const lastUpdated = ref(null)
const refreshTimer = null
const ws = ref(null)
const searchQuery = ref('')
const activeFilter = ref('all')
const dialModalOpen = ref(false)
const dialForm = ref({ fromExtension: '', toNumber: '' })

const statusFilters = [
  { label: 'All', value: 'all' },
  { label: 'Online', value: 'online' },
  { label: 'Busy', value: 'busy' },
  { label: 'Ringing', value: 'ringing' },
  { label: 'Offline', value: 'offline' }
]

// ============ Data Fetching ============

const fetchData = async () => {
  try {
    const [panelRes, callsRes, queueRes] = await Promise.all([
      operatorPanelAPI.getData(),
      liveAPI.getActiveCalls().catch(() => ({ data: { raw: '[]' } })),
      liveAPI.getQueueStats().catch(() => ({ data: [] }))
    ])

    const data = panelRes.data || {}

    // Extensions
    const extList = (data.extensions || []).map(e => {
      const extCalls = (data.active_calls || []).filter(c => {
        const extNum = String(e.extension)
        const callerNum = String(c.caller_id_number || c.cid_num || '')
        const destNum = String(c.destination_number || c.dest || '')
        return callerNum === extNum || destNum === extNum || String(c.presence_id || '') === extNum
      })
      const activeCall = extCalls.length > 0 ? parseCall(extCalls[0]) : null
      return {
        id: e.id,
        number: e.extension,
        name: e.name || `Ext ${e.extension}`,
        status: e.presence || 'offline',
        enabled: e.enabled,
        activeCall
      }
    })
    extensions.value = extList

    // Active calls from panel data
    const panelCalls = (data.active_calls || []).map(parseCall)

    // Also try to parse live calls endpoint (returns raw JSON string)
    let liveCalls = []
    try {
      const raw = callsRes.data?.raw || '[]'
      const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
      liveCalls = (Array.isArray(parsed) ? parsed : []).map(parseCall)
    } catch {
      liveCalls = []
    }

    // Merge calls deduplicated by uuid
    const callMap = new Map()
    panelCalls.forEach(c => callMap.set(c.uuid, c))
    liveCalls.forEach(c => callMap.set(c.uuid, c))
    activeCalls.value = Array.from(callMap.values())

    // Queues: merge panel queue list with live stats
    const panelQueues = (data.queues || []).map(q => ({
      id: q.name,
      name: q.name,
      extension: q.extension || '',
      waiting: parseInt(q.waiting || q.ram_queue || 0, 10),
      avgWait: q.max_wait || '0:00',
      agents: parseInt(q.total_act || q.ram_act || 0, 10)
    }))

    const liveQueueStats = Array.isArray(queueRes.data) ? queueRes.data : []
    const liveQueueMap = new Map()
    liveQueueStats.forEach(q => {
      liveQueueMap.set(q.name || q.extension, q)
    })

    queues.value = panelQueues.map(q => {
      const live = liveQueueMap.get(q.name) || liveQueueMap.get(q.extension)
      if (live) {
        return {
          ...q,
          waiting: live.waiting_calls ?? q.waiting,
          agents: live.available_agents ?? live.total_agents ?? q.agents,
          avgWait: live.avg_wait_time ? formatTime(live.avg_wait_time) : q.avgWait
        }
      }
      return q
    })

    lastUpdated.value = new Date()
  } catch (err) {
    console.error('Failed to load operator panel data:', err)
    toast?.error(err.message, 'Failed to load panel data')
  }
}

function parseCall(call) {
  const uuid = call.uuid || call.call_uuid || call['Call-UUID'] || ''
  const createdEpoch = parseInt(call.created_epoch || call.created_time || call.call_created_epoch || 0, 10)
  const now = Math.floor(Date.now() / 1000)
  const durationSec = createdEpoch > 0 ? now - createdEpoch : 0

  const directionRaw = String(call.direction || call.call_direction || 'inbound').toLowerCase()
  const direction = directionRaw.includes('out') ? 'outbound' : 'inbound'

  return {
    uuid,
    callerName: call.cid_name || call.caller_id_name || call.caller_id || 'Unknown',
    callerNumber: call.cid_num || call.caller_id_number || call.caller_number || '',
    destination: call.destination_number || call.dest || call.called_num || '',
    duration: formatTime(durationSec),
    durationSec,
    direction,
    directionLabel: direction === 'outbound' ? 'Outbound' : 'Inbound',
    state: call.channel_state || call.state || 'active'
  }
}

// ============ WebSocket ============

const connectWebSocket = () => {
  const token = localStorage.getItem('token')
  if (!token) return

  const apiUrl = import.meta.env.VITE_API_URL || ''
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = apiUrl ? new URL(apiUrl).host : window.location.host
  const wsUrl = `${wsProtocol}//${wsHost}/api/ws/notifications?token=${token}`

  try {
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('[OperatorPanel] WebSocket connected')
    }

    ws.value.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'presence' || msg.type === 'call' || msg.type === 'queue') {
          fetchData()
        }
      } catch {
        // ignore
      }
    }

    ws.value.onerror = (err) => {
      console.warn('[OperatorPanel] WebSocket error:', err)
    }

    ws.value.onclose = () => {
      console.log('[OperatorPanel] WebSocket closed, falling back to polling')
      ws.value = null
    }
  } catch (err) {
    console.warn('[OperatorPanel] Failed to connect WebSocket:', err)
  }
}

// ============ Filtering ============

const filteredExtensions = computed(() => {
  let list = extensions.value

  if (activeFilter.value !== 'all') {
    list = list.filter(e => e.status.toLowerCase() === activeFilter.value)
  }

  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter(e =>
      e.name.toLowerCase().includes(q) ||
      e.number.toLowerCase().includes(q) ||
      e.status.toLowerCase().includes(q)
    )
  }

  return list
})

// ============ Derived State ============

const parsedActiveCalls = computed(() => activeCalls.value)
const parsedQueues = computed(() => queues.value)

const lastUpdatedText = computed(() => {
  if (!lastUpdated.value) return 'Never updated'
  const diff = Math.floor((Date.now() - lastUpdated.value.getTime()) / 1000)
  if (diff < 5) return 'Just now'
  if (diff < 60) return `${diff}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  return `${Math.floor(diff / 3600)}h ago`
})

const isStale = computed(() => {
  if (!lastUpdated.value) return true
  return (Date.now() - lastUpdated.value.getTime()) > 30000 // 30s
})

// ============ Actions ============

const hangupCall = async (uuid) => {
  if (!uuid) return
  if (!confirm('Hang up this call?')) return
  try {
    await liveAPI.hangupCall(uuid)
    toast?.success('Hangup command sent')
    await fetchData()
  } catch (err) {
    toast?.error(err.message, 'Failed to hangup')
  }
}

const openDialModal = (ext) => {
  dialForm.value = { fromExtension: ext.number, toNumber: '' }
  dialModalOpen.value = true
}

const executeDial = async () => {
  try {
    await liveAPI.originate(dialForm.value.fromExtension, dialForm.value.toNumber)
    toast?.success('Dial command sent')
    dialModalOpen.value = false
    dialForm.value.toNumber = ''
  } catch (err) {
    toast?.error(err.message, 'Failed to dial')
  }
}

// ============ Lifecycle ============

let intervalId = null

onMounted(() => {
  fetchData()
  connectWebSocket()
  intervalId = setInterval(() => {
    if (!ws.value) {
      fetchData()
    }
  }, 5000)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
  if (ws.value) {
    ws.value.close()
    ws.value = null
  }
})
</script>

<style scoped>
.operator-panel {
  padding: 0;
}

/* Header */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.last-updated {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
  transition: color 0.3s;
}
.last-updated.stale {
  color: var(--status-warn);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--status-good-subtle);
  color: var(--status-good);
  padding: 4px 12px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 600;
}

.dot {
  width: 8px;
  height: 8px;
  background-color: var(--status-good);
  border-radius: 50%;
}

.pulse {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

/* Section Cards */
.section-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-lg);
  overflow: hidden;
}

.section-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--bg-hover);
  border-bottom: 1px solid var(--border-color);
}

.section-header h3 {
  margin: 0;
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.badge-count {
  background: var(--primary-light);
  color: var(--primary-text);
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 99px;
  margin-left: 4px;
}

.icon {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.icon-sm {
  width: 14px;
  height: 14px;
}

.icon-xs {
  width: 12px;
  height: 12px;
}

.icon-lg {
  width: 32px;
  height: 32px;
  color: var(--text-muted);
}

/* Active Calls Table */
.calls-table-wrapper {
  overflow-x: auto;
}

.calls-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.calls-table th {
  text-align: left;
  padding: 10px 16px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
}

.calls-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  vertical-align: middle;
}

.calls-table tr:hover {
  background: var(--bg-hover);
}

.direction-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}

.direction-badge.inbound {
  background: var(--status-info-bg);
  color: var(--status-info);
}

.direction-badge.outbound {
  background: var(--secondary-light);
  color: var(--secondary-color);
}

.caller-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.caller-info strong {
  font-size: 13px;
  color: var(--text-primary);
}

.caller-info .mono {
  font-size: 11px;
  color: var(--text-muted);
}

.duration-cell {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text-primary);
}

.actions-col {
  width: 60px;
  text-align: right;
  white-space: nowrap;
}

/* Queues Grid */
.queues-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
}

.queue-stat-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  transition: all 0.2s;
}

.queue-stat-card.has-waiting {
  border-left: 3px solid var(--status-warn);
}

.queue-stat-header {
  margin-bottom: var(--spacing-md);
}

.queue-stat-header h4 {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
}

.queue-ext {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.queue-stat-body {
  display: flex;
  gap: var(--spacing-lg);
}

.queue-stat-item {
  flex: 1;
  text-align: center;
}

.queue-stat-value {
  display: block;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.queue-stat-value.alert {
  color: var(--status-warn);
}

.queue-stat-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
}

/* Extensions Header */
.extensions-header {
  background: transparent;
  border: none;
  padding: var(--spacing-md) 0;
  flex-wrap: wrap;
}

.extensions-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.search-bar {
  position: relative;
  flex: 1;
  min-width: 240px;
  max-width: 400px;
}

.search-input {
  width: 100%;
  padding: 8px 32px 8px 36px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 13px;
  background: white;
  box-sizing: border-box;
}

.search-input:focus {
  border-color: var(--primary-color);
  outline: none;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.clear-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
}

.filter-pills {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.filter-pill {
  padding: 4px 12px;
  border-radius: 99px;
  border: 1px solid var(--border-color);
  background: white;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s;
}

.filter-pill.active {
  background: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

/* Extensions Grid */
.panel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.ext-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  background: white;
  transition: all 0.2s;
  position: relative;
}

.ext-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.ext-card.busy {
  border-color: var(--status-bad);
  background: var(--status-bad-subtle);
}

.ext-card.ringing {
  border-color: var(--status-warn);
  background: var(--status-warn-subtle);
  animation: border-pulse 1s infinite alternate;
}

.ext-card.offline {
  opacity: 0.6;
  background: var(--bg-hover);
}

.ext-card.has-active-call {
  border-left: 3px solid var(--status-info);
}

.ext-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.ext-number {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.ext-name {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ext-call-info {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-main);
  background: rgba(255,255,255,0.6);
  padding: 4px 6px;
  border-radius: 4px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.call-detail {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.call-duration {
  font-family: var(--font-mono);
  font-weight: 600;
}

.ext-actions {
  display: flex;
  gap: 6px;
  justify-content: flex-end;
}

.btn-icon-sm {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-muted);
  transition: all 0.15s;
}

.btn-icon-sm:hover {
  color: var(--primary-color);
  border-color: var(--primary-color);
}

.btn-icon-sm.danger:hover {
  color: var(--status-bad);
  border-color: var(--status-bad);
  background: var(--status-bad-subtle);
}

@keyframes border-pulse {
  from { border-color: var(--status-warn); }
  to { border-color: #f97316; }
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  color: var(--text-muted);
  gap: var(--spacing-sm);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: var(--z-modal);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-card {
  background: white;
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 420px;
  box-shadow: var(--shadow-xl);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
}

.close-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: var(--bg-hover);
  border-radius: 6px;
  font-size: 18px;
  cursor: pointer;
  color: var(--text-muted);
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.input-field {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  box-sizing: border-box;
  font-family: var(--font-mono);
}

.input-field:focus {
  border-color: var(--primary-color);
  outline: none;
}

.input-field[readonly] {
  background: var(--bg-hover);
}

.btn-primary, .btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}

.btn-primary {
  background: var(--primary-color);
  color: white;
}

.btn-primary:hover {
  background: var(--primary-hover);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  color: var(--text-main);
}

.mono {
  font-family: var(--font-mono);
}

/* Responsive */
@media (max-width: 768px) {
  .view-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .extensions-header {
    flex-direction: column;
    align-items: stretch;
  }

  .search-bar {
    max-width: 100%;
  }

  .panel-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }

  .queues-grid {
    grid-template-columns: 1fr;
  }

  .calls-table th,
  .calls-table td {
    padding: 8px 10px;
  }
}
</style>
