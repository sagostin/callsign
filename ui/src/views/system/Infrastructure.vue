<template>
  <div class="view-header">
    <h2>Infrastructure</h2>
    <p class="text-muted text-sm">Cluster health, node status, and capacity planning.</p>
  </div>

  <div v-if="loading" class="loading-state">Loading infrastructure status...</div>
  <div v-else-if="error" class="error-state">{{ error }}</div>

  <div class="node-grid" v-else>
    <div class="node-card active">
      <div class="node-header">
        <h3>{{ nodeName }}</h3>
        <StatusBadge :status="eslConnected ? 'good' : 'warning'" />
      </div>
      <div class="node-stats">
        <div class="stat-row">
          <span>ESL Connection</span>
          <span :class="eslConnected ? 'text-good' : 'text-bad'">{{ eslConnected ? 'Connected' : 'Disconnected' }}</span>
        </div>
        <div class="stat-row">
          <span>Registrations</span>
          <span>{{ registrations.toLocaleString() }}</span>
        </div>
        <div class="stat-row">
          <span>Active Channels</span>
          <span>{{ activeChannels.toLocaleString() }}</span>
        </div>
        <div class="stat-row">
          <span>Tenants</span>
          <span>{{ tenantCount }}</span>
        </div>
        <div class="stat-row">
          <span>Extensions</span>
          <span>{{ extensionCount.toLocaleString() }}</span>
        </div>
        <div class="stat-row">
          <span>Gateways</span>
          <span>{{ gatewayCount }}</span>
        </div>
      </div>
    </div>

    <!-- FreeSWITCH Status Panel -->
    <div class="node-card" :class="{ active: fsStatus }">
      <div class="node-header">
        <h3>FreeSWITCH</h3>
        <StatusBadge :status="eslConnected ? 'good' : 'warning'" />
      </div>
      <div v-if="fsStatus" class="node-stats">
        <div class="stat-row" v-for="(value, key) in fsStatus" :key="key">
          <span>{{ formatKey(key) }}</span>
          <span>{{ value }}</span>
        </div>
      </div>
      <div v-else class="node-stats">
        <div class="stat-row">
          <span>Status</span>
          <span class="text-muted">{{ eslConnected ? 'Querying...' : 'ESL not connected' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api'

const loading = ref(true)
const error = ref(null)
const eslConnected = ref(false)
const registrations = ref(0)
const activeChannels = ref(0)
const tenantCount = ref(0)
const extensionCount = ref(0)
const gatewayCount = ref(0)
const fsStatus = ref(null)
const nodeName = ref('Primary Node')
let refreshInterval = null

async function loadData() {
  try {
    // Load stats
    const statsResp = await systemAPI.getStats()
    const s = statsResp.data
    eslConnected.value = s.esl_connected || false
    registrations.value = s.registrations || 0
    activeChannels.value = s.active_channels || 0
    tenantCount.value = s.tenants || 0
    extensionCount.value = s.extensions || 0
    gatewayCount.value = s.gateways || 0

    // Try to load FS status
    try {
      const statusResp = await systemAPI.getStatus()
      const status = statusResp.data
      if (status.freeswitch) {
        fsStatus.value = status.freeswitch
      }
    } catch {
      // ESL not connected — just skip
    }

    error.value = null
  } catch (e) {
    error.value = e.message
    console.error('Failed to load infrastructure data:', e)
  } finally {
    loading.value = false
  }
}

function formatKey(key) {
  return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
}

onMounted(() => {
  loadData()
  refreshInterval = setInterval(loadData, 30000)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
</script>

<style scoped>
.node-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-lg);
  margin-top: var(--spacing-lg);
}

.node-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}

.node-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.node-header h3 {
  font-size: var(--text-lg);
  margin: 0;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px dashed var(--border-color);
  font-size: var(--text-sm);
}

.stat-row:last-child {
  border-bottom: none;
}

.text-good { color: var(--status-good); font-weight: 600; }
.text-bad { color: var(--status-bad); font-weight: 600; }
.text-muted { color: var(--text-muted); }
.loading-state, .error-state { padding: 40px; text-align: center; color: var(--text-muted); }
.error-state { color: var(--status-bad); }
</style>
