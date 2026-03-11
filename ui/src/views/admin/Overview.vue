<template>
  <div class="view-header">
    <h2>System Overview</h2>
    <p class="text-muted text-sm">Real-time system status and metrics.</p>
  </div>

  <div class="dashboard-grid">
    <StatCard 
      label="Active Calls" 
      :value="stats.activeCalls" 
      :subtext="stats.activeCallsSub"
      iconName="calls"
    />
    <StatCard 
      label="Registrations" 
      :value="stats.registrations" 
      :subtext="stats.registrationsSub"
      iconName="server"
    />
    <StatCard 
      label="Failed Calls (1h)" 
      :value="stats.failedCalls" 
      :subtext="stats.failedCallsSub"
      iconName="alert"
    />
    <StatCard 
      label="System Health" 
      :value="stats.health" 
      :subtext="stats.healthSub"
      iconName="activity"
    />
  </div>

  <div class="layout-split">
    <!-- Recent Alerts -->
    <div class="panel">
      <div class="panel-header">
        <h3>Recent Alerts</h3>
        <button class="btn-link" @click="$router.push('/admin/audit-log')">View All</button>
      </div>
      <div class="alert-list">
        <div class="alert-item" v-for="alert in recentAlerts" :key="alert.id">
          <span class="badge" :class="alert.severity">{{ alert.severity.toUpperCase() }}</span>
          <div class="alert-content">
            <div class="alert-msg">{{ alert.message }}</div>
            <div class="alert-time">{{ alert.time }}</div>
          </div>
        </div>
        <div v-if="recentAlerts.length === 0" class="empty-text">No recent alerts</div>
      </div>
    </div>

    <!-- Active Calls Table (Mini) -->
    <div class="panel">
      <div class="panel-header">
        <h3>Live Calls</h3>
      </div>
      <table class="simple-table" v-if="liveCalls.length > 0">
        <thead>
          <tr>
            <th>From</th>
            <th>To</th>
            <th>Duration</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="call in liveCalls" :key="call.id">
            <td>{{ call.from }}</td>
            <td>{{ call.to }}</td>
            <td>{{ call.duration }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-text">No active calls</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import StatCard from '../../components/features/StatCard.vue'
import { systemAPI, auditLogAPI } from '../../services/api'

const stats = reactive({
  activeCalls: '—',
  activeCallsSub: '',
  registrations: '—',
  registrationsSub: '',
  failedCalls: '—',
  failedCallsSub: '',
  health: '—',
  healthSub: ''
})

const recentAlerts = ref([])
const liveCalls = ref([])
let refreshTimer = null

const fetchOverview = async () => {
  try {
    const res = await systemAPI.getStats()
    const d = res.data || {}
    stats.activeCalls = String(d.active_calls ?? d.channels ?? '0')
    stats.activeCallsSub = d.peak_calls ? `Peak: ${d.peak_calls} today` : ''
    stats.registrations = String(d.registrations ?? d.registered_devices ?? '0')
    const total = d.total_devices || d.registrations || 0
    const pct = total > 0 ? Math.round((d.registrations || 0) / total * 100) : 0
    stats.registrationsSub = total > 0 ? `${pct}% Online` : ''
    stats.failedCalls = String(d.failed_calls_1h ?? d.failed_calls ?? '0')
    stats.failedCallsSub = d.failed_reason || ''
    stats.health = d.health_status || 'Operational'
    stats.healthSub = d.uptime ? `Uptime: ${d.uptime}` : ''
  } catch (err) {
    console.error('Failed to load overview stats:', err)
  }
}

const fetchAlerts = async () => {
  try {
    const res = await auditLogAPI.list({ limit: 5, severity: 'warning,critical' })
    const items = res.data?.data || res.data || []
    recentAlerts.value = items.slice(0, 4).map(a => {
      const dt = new Date(a.created_at)
      const diffMin = Math.round((Date.now() - dt) / 60000)
      return {
        id: a.id,
        severity: a.severity === 'critical' ? 'bad' : 'warn',
        message: a.action || a.description || 'Alert',
        time: diffMin < 60 ? `${diffMin} mins ago` : `${Math.round(diffMin / 60)}h ago`
      }
    })
  } catch (err) {
    console.error('Failed to load alerts:', err)
  }
}

const fetchLiveCalls = async () => {
  try {
    const res = await systemAPI.getChannels()
    const calls = res.data?.channels || res.data || []
    liveCalls.value = calls.slice(0, 5).map((c, i) => ({
      id: c.uuid || i,
      from: c.caller_id_number || c.cid_num || '—',
      to: c.destination || c.dest || '—',
      duration: formatDuration(c.elapsed || c.duration || 0)
    }))
  } catch (err) {
    // Live calls endpoint may not exist yet — that's fine
    liveCalls.value = []
  }
}

const formatDuration = (sec) => {
  const m = Math.floor(sec / 60)
  const s = String(sec % 60).padStart(2, '0')
  return `${m}:${s}`
}

onMounted(() => {
  fetchOverview()
  fetchAlerts()
  fetchLiveCalls()
  refreshTimer = setInterval(() => {
    fetchOverview()
    fetchLiveCalls()
  }, 15000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--spacing-lg);
  margin: var(--spacing-lg) 0 var(--spacing-xl);
}

.layout-split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-lg);
}

.panel {
  background: white;
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: var(--spacing-sm);
}

.panel-header h3 {
  font-size: var(--text-sm);
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  font-weight: 600;
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.alert-item {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
  padding: 12px;
  background-color: var(--bg-app);
  border-radius: var(--radius-md);
  transition: transform var(--transition-fast);
}

.alert-item:hover {
  transform: translateX(2px);
}

.alert-content {
  flex: 1;
}

.alert-msg {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-primary);
}

.alert-time {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 2px;
}

.simple-table {
  width: 100%;
}
.simple-table th {
  text-align: left;
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 700;
  text-transform: uppercase;
  padding-bottom: 12px;
}
.simple-table td {
  font-size: var(--text-sm);
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-main);
}
.simple-table tr:last-child td {
  border-bottom: none;
}
.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  cursor: pointer;
  font-weight: 600;
}
</style>
