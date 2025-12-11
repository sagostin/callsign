<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Dial Plan Manager</h2>
      <p class="text-muted text-sm">Manage inbound and outbound routing logic.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/dial-plans/new')">+ New Dial Plan</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="plans" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    
    <template #actions>
      <button class="btn-link" @click="$router.push('/admin/dial-plans/1')">Edit</button>
      <button class="btn-link text-bad">Delete</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../components/common/DataTable.vue'

const columns = [
  { key: 'priority', label: 'Order', width: '80px' },
  { key: 'name', label: 'Name' },
  { key: 'condition', label: 'Condition / RegEx' },
  { key: 'action', label: 'Action' },
  { key: 'enabled', label: 'Global' }
]

const rules = ref([
  { priority: '010', name: 'Internal Calls', condition: '^(\\d{3})$', action: 'bridge:user/$1', enabled: 'Yes' },
  { priority: '100', name: 'Local 10-digit', condition: '^\\+?1?(\\d{10})$', action: 'bridge:gateway/flowroute/$1', enabled: 'No' },
  { priority: '999', name: 'Catch All', condition: '.*', action: 'hangup', enabled: 'No' },
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

.btn-secondary {
  background: white; border: 1px solid var(--border-color); color: var(--text-main);
  padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500;
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

.font-mono { font-family: monospace; color: var(--text-muted); }
.condition-code { 
  background: var(--bg-secondary); 
  padding: 2px 6px; 
  border-radius: 4px; 
  font-family: monospace; 
  font-size: 11px; 
}
</style>
