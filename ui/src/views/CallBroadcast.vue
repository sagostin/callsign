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
      <button class="btn-link text-bad">Delete</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../components/common/DataTable.vue'
import StatusBadge from '../components/common/StatusBadge.vue'

const columns = [
  { key: 'name', label: 'Campaign Name' },
  { key: 'type', label: 'Type' },
  { key: 'target_count', label: 'Targets' },
  { key: 'progress', label: 'Completion' },
  { key: 'status', label: 'Status' }
]

const campaigns = ref([
  { name: 'Service Outage Alert', type: 'Voice', target_count: 540, progress: 100, status: 'Completed' },
  { name: 'Promo Q4', type: 'SMS', target_count: 1200, progress: 45, status: 'Active' },
  { name: 'Appointment Reminders', type: 'Voice', target_count: 12, progress: 0, status: 'Scheduled' },
])
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
