<template>
  <div class="page-content">
    <!-- View Header -->
    <div class="view-header">
      <div class="view-header-content">
        <h1 class="view-header-title">System Overview</h1>
        <p class="view-header-subtitle">Real-time system status and metrics</p>
      </div>
      <div class="view-header-actions">
        <button class="btn btn-secondary" @click="refreshData">
          <RefreshCwIcon class="btn-icon" :class="{ 'animate-spin': isRefreshing }" />
          <span class="hide-mobile">Refresh</span>
        </button>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <StatCard 
        label="Active Calls" 
        :value="stats.activeCalls" 
        :subtext="stats.activeCallsSub"
        iconName="calls"
        variant="info"
      />
      <StatCard 
        label="Registrations" 
        :value="stats.registrations" 
        :subtext="stats.registrationsSub"
        iconName="users"
        variant="success"
      />
      <StatCard 
        label="Failed Calls (1h)" 
        :value="stats.failedCalls" 
        :subtext="stats.failedCallsSub"
        iconName="alert"
        :variant="failedCallsVariant"
      />
      <StatCard 
        label="System Health" 
        :value="stats.health" 
        :subtext="stats.healthSub"
        iconName="activity"
        variant="success"
      />
    </div>

    <!-- Dashboard Panels -->
    <div class="dashboard-panels">
      <!-- Recent Alerts Panel -->
      <div class="panel">
        <div class="panel-header">
          <div class="panel-header-content">
            <AlertTriangleIcon class="panel-icon" />
            <h3 class="panel-title">Recent Alerts</h3>
          </div>
          <button class="btn btn-sm btn-ghost" @click="$router.push('/admin/audit-log')">
            View All
            <ArrowRightIcon class="btn-icon-sm" />
          </button>
        </div>
        <div class="panel-body">
          <div class="alert-list">
            <div 
              v-for="alert in recentAlerts" 
              :key="alert.id" 
              class="alert-item"
              :class="alert.severity"
            >
              <div class="alert-icon">
                <AlertCircleIcon v-if="alert.severity === 'bad'" />
                <AlertTriangleIcon v-else />
              </div>
              <div class="alert-content">
                <div class="alert-message">{{ alert.message }}</div>
                <div class="alert-time">{{ alert.time }}</div>
              </div>
            </div>
            <div v-if="recentAlerts.length === 0" class="empty-state-compact">
              <CheckCircleIcon class="empty-icon-small" />
              <span>No recent alerts</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Live Calls Panel -->
      <div class="panel">
        <div class="panel-header">
          <div class="panel-header-content">
            <PhoneIcon class="panel-icon" />
            <h3 class="panel-title">Live Calls</h3>
          </div>
          <span class="live-indicator">
            <span class="live-dot"></span>
            Live
          </span>
        </div>
        <div class="panel-body">
          <div v-if="liveCalls.length > 0" class="live-calls-list">
            <div 
              v-for="call in liveCalls" 
              :key="call.id" 
              class="live-call-item"
            >
              <div class="call-parties">
                <div class="call-party">
                  <span class="party-label">From</span>
                  <span class="party-number">{{ call.from }}</span>
                </div>
                <ArrowRightIcon class="call-arrow" />
                <div class="call-party">
                  <span class="party-label">To</span>
                  <span class="party-number">{{ call.to }}</span>
                </div>
              </div>
              <div class="call-duration">{{ call.duration }}</div>
            </div>
          </div>
          <div v-else class="empty-state-compact">
            <PhoneOffIcon class="empty-icon-small" />
            <span>No active calls</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import StatCard from '../../components/features/StatCard.vue'
import { systemAPI, auditLogAPI } from '../../services/api'
import { 
  RefreshCw as RefreshCwIcon,
  ArrowRight as ArrowRightIcon,
  AlertTriangle as AlertTriangleIcon,
  AlertCircle as AlertCircleIcon,
  CheckCircle as CheckCircleIcon,
  Phone as PhoneIcon,
  PhoneOff as PhoneOffIcon
} from 'lucide-vue-next'

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

const isRefreshing = ref(false)
const recentAlerts = ref([])
const liveCalls = ref([])
let refreshTimer = null

const failedCallsVariant = computed(() => {
  const failed = parseInt(stats.failedCalls) || 0
  if (failed === 0) return 'success'
  if (failed < 5) return 'warning'
  return 'error'
})

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
    liveCalls.value = []
  }
}

const formatDuration = (sec) => {
  const m = Math.floor(sec / 60)
  const s = String(sec % 60).padStart(2, '0')
  return `${m}:${s}`
}

const refreshData = async () => {
  isRefreshing.value = true
  await Promise.all([fetchOverview(), fetchAlerts(), fetchLiveCalls()])
  setTimeout(() => { isRefreshing.value = false }, 500)
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
/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--spacing-5);
  margin-bottom: var(--spacing-6);
}

/* Dashboard Panels */
.dashboard-panels {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-6);
}

/* Panel Customization */
.panel {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-4) var(--spacing-5);
  background: var(--bg-hover);
  border-bottom: 1px solid var(--border-color);
}

.panel-header-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
}

.panel-icon {
  width: 18px;
  height: 18px;
  color: var(--text-secondary);
}

.panel-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0;
}

.panel-body {
  padding: var(--spacing-4);
}

/* Live Indicator */
.live-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-1-5);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--status-good);
  text-transform: uppercase;
  letter-spacing: var(--tracking-wider);
}

.live-dot {
  width: 8px;
  height: 8px;
  background: var(--status-good);
  border-radius: var(--radius-full);
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Alert List */
.alert-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.alert-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-3);
  padding: var(--spacing-3);
  background: var(--bg-hover);
  border-radius: var(--radius-md);
  transition: transform var(--transition-fast);
}

.alert-item:hover {
  transform: translateX(2px);
}

.alert-item.bad {
  background: var(--status-bad-subtle);
  border-left: 3px solid var(--status-bad);
}

.alert-item.warn {
  background: var(--status-warn-subtle);
  border-left: 3px solid var(--status-warn);
}

.alert-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  margin-top: 2px;
}

.alert-item.bad .alert-icon {
  color: var(--status-bad);
}

.alert-item.warn .alert-icon {
  color: var(--status-warn);
}

.alert-content {
  flex: 1;
  min-width: 0;
}

.alert-message {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--text-primary);
  line-height: var(--leading-snug);
}

.alert-time {
  font-size: var(--text-xs);
  color: var(--text-muted);
  margin-top: var(--spacing-0-5);
}

/* Live Calls List */
.live-calls-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
}

.live-call-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-3);
  background: var(--bg-hover);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.live-call-item:hover {
  background: var(--border-light);
}

.call-parties {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  flex: 1;
  min-width: 0;
}

.call-party {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.party-label {
  font-size: var(--text-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: var(--tracking-wider);
}

.party-number {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.call-arrow {
  width: 14px;
  height: 14px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.call-duration {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  padding-left: var(--spacing-3);
  border-left: 1px solid var(--border-color);
  margin-left: var(--spacing-3);
}

/* Empty State Compact */
.empty-state-compact {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-2);
  padding: var(--spacing-8);
  color: var(--text-muted);
  font-size: var(--text-sm);
}

.empty-icon-small {
  width: 20px;
  height: 20px;
}

/* Button Icons */
.btn-icon {
  width: 16px;
  height: 16px;
}

.btn-icon-sm {
  width: 14px;
  height: 14px;
}

/* Mobile Responsive */
@media (max-width: 1024px) {
  .dashboard-panels {
    gap: var(--spacing-4);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-4);
  }
  
  .dashboard-panels {
    grid-template-columns: 1fr;
  }
  
  .panel-header {
    padding: var(--spacing-3) var(--spacing-4);
  }
  
  .panel-body {
    padding: var(--spacing-3);
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .call-parties {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-2);
  }
  
  .call-arrow {
    transform: rotate(90deg);
  }
  
  .call-duration {
    border-left: none;
    border-top: 1px solid var(--border-color);
    padding-left: 0;
    padding-top: var(--spacing-2);
    margin-left: 0;
    margin-top: var(--spacing-2);
  }
  
  .live-call-item {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
