<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Trunks</h2>
      <p class="text-muted text-sm">Manage SIP trunks and PSTN gateways for this tenant.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/trunks/new')">+ Add Trunk</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="gateways" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    
    <template #type="{ value }">
        <span class="badge" :class="value.toLowerCase()">{{ value }}</span>
    </template>

    <template #actions="{ row }">
      <button class="btn-link" @click="$router.push(`/admin/trunks/${row.id}`)">Edit</button>
      <button class="btn-link text-bad" @click="deleteTrunk(row)">Delete</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import DataTable from '../components/common/DataTable.vue'
import StatusBadge from '../components/common/StatusBadge.vue'
import { systemAPI } from '../services/api'

const columns = [
  { key: 'gateway_name', label: 'Trunk Name' },
  { key: 'proxy', label: 'Hostname / IP' },
  { key: 'gateway_type', label: 'Type' },
  { key: 'status', label: 'Status' }
]

const gateways = ref([])
const loading = ref(true)

const loadTrunks = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listGateways()
    const data = response.data?.data || response.data || []
    gateways.value = data.map(g => ({
      ...g,
      status: g.enabled ? (g.register ? 'Registered' : 'Active') : 'Disabled',
      gateway_type: g.gateway_type || 'Public'
    }))
  } catch (e) {
    console.error('Failed to load trunks', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadTrunks)

const deleteTrunk = async (trunk) => {
  if (confirm(`Delete trunk "${trunk.gateway_name}"?`)) {
    try {
      await systemAPI.deleteGateway(trunk.id)
      await loadTrunks()
    } catch (e) {
      alert('Failed to delete trunk: ' + e.message)
    }
  }
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  cursor: pointer;
  font-weight: 500;
}

.text-bad { color: var(--status-bad); }

.badge { padding: 2px 6px; border-radius: 4px; font-size: 11px; font-weight: 600; background: var(--bg-secondary); }
.badge.public { background: #e0f2fe; color: #0369a1; }
.badge.local { background: #f3e8ff; color: #7e22ce; }
</style>
