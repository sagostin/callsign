<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Extensions</h2>
      <p class="text-muted text-sm">Global registry of all extensions across tenants.</p>
    </div>
  </div>

  <div class="filter-bar">
    <input type="text" placeholder="Search extensions..." class="search-input">
    <select class="filter-select">
       <option>All Tenants</option>
       <option>Acme Corp</option>
       <option>Globex Inc</option>
    </select>
  </div>

  <DataTable :columns="columns" :data="extensions" actions>
    <template #tenant="{ value }">
      <span class="badge tenant">{{ value }}</span>
    </template>
    
    <template #status="{ value }">
       <StatusBadge :status="value" />
    </template>

    <template #actions>
      <button class="btn-link">Debug</button>
      <button class="btn-link text-bad">Unreg</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'

const columns = [
  { key: 'ext', label: 'Extension', width: '100px' },
  { key: 'name', label: 'Name' },
  { key: 'tenant', label: 'Tenant' },
  { key: 'type', label: 'Device Type' },
  { key: 'status', label: 'Status' }
]

const extensions = ref([
  { ext: '1001', name: 'Alice Smith', tenant: 'Acme Corp', type: 'Yealink T54W', status: 'Registered' },
  { ext: '1002', name: 'Bob Jones', tenant: 'Acme Corp', type: 'Softphone', status: 'Offline' },
  { ext: '5001', name: 'Support Phone', tenant: 'Globex Inc', type: 'Poly VVX', status: 'Registered' },
])
</script>

<style scoped>
.view-header {
  margin-bottom: var(--spacing-lg);
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: var(--spacing-lg);
}

.search-input {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  width: 300px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.badge.tenant {
  background: var(--bg-app);
  color: var(--text-muted);
  border: 1px solid var(--border-color);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
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
</style>
