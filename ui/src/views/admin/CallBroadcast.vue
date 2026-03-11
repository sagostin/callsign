<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Call Broadcast</h2>
      <p class="text-muted text-sm">Manage voice campaigns and massive blasting.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/call-broadcast/new')">+ New Campaign</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="campaigns" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    
    <template #progress="{ value }">
      <span class="text-xs">{{ value }}%</span>
    </template>

    <template #actions="{ row }">
      <button class="btn-link" @click="$router.push(`/admin/call-broadcast/${row.id}`)">Edit</button>
      <button class="btn-link text-bad" @click="deleteCampaign(row)">Delete</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { broadcastAPI } from '../../services/api'

const toast = inject('toast')

const columns = [
  { key: 'name', label: 'Campaign Name' },
  { key: 'status', label: 'Status' },
  { key: 'target_count', label: 'Targets' },
  { key: 'progress', label: 'Completion' },
]

const campaigns = ref([])

onMounted(async () => {
  try {
    const res = await broadcastAPI.list()
    campaigns.value = res.data || []
  } catch (err) {
    console.error('Failed to load campaigns:', err)
    campaigns.value = []
  }
})

const deleteCampaign = async (row) => {
  if (!confirm(`Delete campaign "${row.name}"?`)) return
  try {
    await broadcastAPI.delete(row.id)
    campaigns.value = campaigns.value.filter(c => c.id !== row.id)
    toast?.success('Campaign deleted')
  } catch (err) {
    toast?.error(err.message, 'Failed to delete campaign')
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
.text-xs { font-size: 11px; }
</style>
