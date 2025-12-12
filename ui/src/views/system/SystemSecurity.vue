<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Security</h2>
      <p class="text-muted text-sm">Monitor and manage security threats, banned IPs, and fail2ban status.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="loadBannedIPs">
        <RefreshCwIcon class="btn-icon" /> Refresh
      </button>
    </div>
  </div>

  <!-- Stats Row -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-icon">
        <ShieldAlertIcon class="icon-lg text-bad" />
      </div>
      <div class="stat-info">
        <div class="stat-value">{{ bannedCount }}</div>
        <div class="stat-label">Currently Banned</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon">
        <ShieldCheckIcon class="icon-lg text-good" />
      </div>
      <div class="stat-info">
        <div class="stat-value">{{ unbannedCount }}</div>
        <div class="stat-label">Unbanned Today</div>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon">
        <ActivityIcon class="icon-lg text-warning" />
      </div>
      <div class="stat-info">
        <div class="stat-value">{{ totalBans }}</div>
        <div class="stat-label">Total Bans</div>
      </div>
    </div>
  </div>

  <!-- Filter Tabs -->
  <div class="tabs">
    <button class="tab" :class="{ active: statusFilter === 'banned' }" @click="statusFilter = 'banned'">
      Currently Banned
    </button>
    <button class="tab" :class="{ active: statusFilter === 'all' }" @click="statusFilter = 'all'">
      All History
    </button>
  </div>

  <!-- Banned IPs Table -->
  <div class="table-container">
    <DataTable :columns="columns" :data="filteredIPs" :loading="loading">
      <template #ip="{ value }">
        <span class="ip-badge font-mono">{{ value }}</span>
      </template>
      <template #source="{ value }">
        <span class="source-badge">{{ value || 'fail2ban' }}</span>
      </template>
      <template #failures="{ value }">
        <span class="failures-count">{{ value || 0 }}</span>
      </template>
      <template #banned_at="{ value }">
        <span class="text-muted">{{ formatDate(value) }}</span>
      </template>
      <template #status="{ value }">
        <StatusBadge :status="value === 'banned' ? 'Banned' : 'Unbanned'" />
      </template>
      <template #actions="{ row }">
        <button 
          v-if="row.status === 'banned'" 
          class="btn-sm btn-danger" 
          @click="unbanIP(row)"
        >
          <UnlockIcon class="icon-xs" /> Unban
        </button>
        <span v-else class="text-muted text-sm">—</span>
      </template>
    </DataTable>

    <div v-if="!loading && filteredIPs.length === 0" class="empty-state">
      <ShieldCheckIcon class="empty-icon" />
      <h3>No Banned IPs</h3>
      <p>Your system is clean. No threats have been detected.</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  RefreshCw as RefreshCwIcon, 
  ShieldAlert as ShieldAlertIcon,
  ShieldCheck as ShieldCheckIcon,
  Activity as ActivityIcon,
  Unlock as UnlockIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api'

const loading = ref(false)
const bannedIPs = ref([])
const statusFilter = ref('banned')

const columns = [
  { key: 'ip', label: 'IP Address', width: '150px' },
  { key: 'source', label: 'Source', width: '120px' },
  { key: 'reason', label: 'Reason', width: '200px' },
  { key: 'failures', label: 'Failures', width: '80px' },
  { key: 'banned_at', label: 'Banned At', width: '180px' },
  { key: 'status', label: 'Status', width: '100px' },
  { key: 'actions', label: '', width: '100px' }
]

const filteredIPs = computed(() => {
  if (statusFilter.value === 'all') return bannedIPs.value
  return bannedIPs.value.filter(ip => ip.status === statusFilter.value)
})

const bannedCount = computed(() => 
  bannedIPs.value.filter(ip => ip.status === 'banned').length
)

const unbannedCount = computed(() => 
  bannedIPs.value.filter(ip => ip.status === 'unbanned').length
)

const totalBans = computed(() => bannedIPs.value.length)

const loadBannedIPs = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listBannedIPs({ status: 'all' })
    bannedIPs.value = response.data.data || []
  } catch (error) {
    console.error('Failed to load banned IPs:', error)
  } finally {
    loading.value = false
  }
}

const unbanIP = async (ip) => {
  if (!confirm(`Unban IP ${ip.ip}? This will allow traffic from this IP again.`)) return
  
  try {
    await systemAPI.unbanIP(ip.ip)
    ip.status = 'unbanned'
  } catch (error) {
    console.error('Failed to unban IP:', error)
    alert('Failed to unban IP: ' + error.message)
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  return date.toLocaleString()
}

onMounted(() => {
  loadBannedIPs()
})
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}
.header-content h2 { margin: 0 0 4px 0; font-size: 24px; font-weight: 600; }
.header-content p { margin: 0; }

.stats-row {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}
.stat-card {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 16px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
}
.stat-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  background: var(--bg-secondary);
}
.icon-lg { width: 24px; height: 24px; }
.text-bad { color: #ef4444; }
.text-good { color: #22c55e; }
.text-warning { color: #f59e0b; }
.stat-value { font-size: 28px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

.tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 12px;
}
.tab {
  padding: 8px 16px;
  border: none;
  background: none;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 6px;
}
.tab:hover { background: var(--bg-secondary); }
.tab.active { background: var(--primary-color); color: white; }

.table-container {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
}

.ip-badge {
  background: #fef2f2;
  color: #dc2626;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 13px;
}
.source-badge {
  background: var(--bg-secondary);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.failures-count {
  font-weight: 600;
  color: #ef4444;
}

.btn-secondary {
  display: flex;
  align-items: center;
  gap: 6px;
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.btn-secondary:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-icon { width: 14px; height: 14px; }

.btn-sm {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
  border: none;
  cursor: pointer;
}
.btn-danger { background: #fef2f2; color: #dc2626; }
.btn-danger:hover { background: #fee2e2; }
.icon-xs { width: 12px; height: 12px; }

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-muted);
}
.empty-icon { width: 48px; height: 48px; margin-bottom: 16px; color: #22c55e; }
.empty-state h3 { margin: 0 0 8px 0; color: var(--text-primary); }
.empty-state p { margin: 0; }

@media (max-width: 768px) {
  .stats-row { flex-direction: column; }
  .view-header { flex-direction: column; gap: 12px; }
}
</style>
