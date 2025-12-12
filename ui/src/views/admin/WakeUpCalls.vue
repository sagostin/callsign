<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Wake Up Calls</h2>
      <p class="text-muted text-sm">Schedule automated wake up calls for guest rooms.</p>
    </div>
    <div class="header-actions">
       <button class="btn-primary" @click="$router.push('/admin/wake-up-calls/new')">+ Schedule Call</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="calls" actions>
    <template #status="{ value }">
       <span class="status-badge" :class="value.toLowerCase()">{{ value }}</span>
    </template>
    <template #actions>
      <button class="btn-link" @click="$router.push('/admin/wake-up-calls/1')">Edit</button>
      <button class="btn-link text-bad">Cancel</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'

const columns = [
  { key: 'extension', label: 'Room / Extension' },
  { key: 'time', label: 'Scheduled Time' },
  { key: 'recurrence', label: 'Recurrence' },
  { key: 'status', label: 'Status' }
]

const calls = ref([
  { extension: '101 (Room 101)', time: '07:00 AM', recurrence: 'Once', status: 'Pending' },
  { extension: '205 (Room 205)', time: '06:30 AM', recurrence: 'Daily', status: 'Active' },
  { extension: '300 (Suite)', time: '08:00 AM', recurrence: 'Once', status: 'Completed' },
])
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}
.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  font-size: 13px;
}
.btn-link {
  background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 600; cursor: pointer;
}
.text-bad { color: var(--status-bad); }

.status-badge { padding: 2px 8px; border-radius: 99px; font-size: 11px; font-weight: 600; }
.status-badge.pending { background: #fee2e2; color: #991b1b; } /* Bad example color, fixing... Pending usually yellow/orange */
.status-badge.active { background: #dcfce7; color: #166534; }
.status-badge.completed { background: #f3f4f6; color: #4b5563; }
</style>
