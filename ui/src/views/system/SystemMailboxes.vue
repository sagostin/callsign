<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Mailboxes</h2>
      <p class="text-muted text-sm">Manage voicemail boxes across all tenants.</p>
    </div>
  </div>

  <div class="filter-bar">
    <input type="text" placeholder="Search mailboxes..." class="search-input">
    <select class="filter-select">
       <option>All Tenants</option>
       <option>Acme Corp</option>
       <option>Globex Inc</option>
    </select>
  </div>

  <DataTable :columns="columns" :data="mailboxes" actions>
    <template #tenant="{ value }">
      <span class="badge tenant">{{ value }}</span>
    </template>
    
    <template #usage="{ value }">
       <div class="usage-bar">
          <div class="fill" :style="{ width: value + '%' }"></div>
       </div>
       <span class="usage-text">{{ value }}%</span>
    </template>

    <template #actions>
      <button class="btn-link">View</button>
      <button class="btn-link text-bad">Reset</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'

const columns = [
  { key: 'mailbox', label: 'Mailbox', width: '100px' },
  { key: 'user', label: 'User / Name' },
  { key: 'tenant', label: 'Tenant' },
  { key: 'messages', label: 'Msgs', width: '80px' },
  { key: 'usage', label: 'Storage' }
]

const mailboxes = ref([
  { mailbox: '1001', user: 'Alice Smith', tenant: 'Acme Corp', messages: 12, usage: 45 },
  { mailbox: '1002', user: 'Bob Jones', tenant: 'Acme Corp', messages: 0, usage: 2 },
  { mailbox: '5001', user: 'Support Queue', tenant: 'Globex Inc', messages: 142, usage: 92 },
  { mailbox: '2023', user: 'Front Desk', tenant: 'Hampton Hotel', messages: 5, usage: 10 },
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

.usage-bar {
  width: 60px;
  height: 6px;
  background: var(--bg-app);
  border-radius: 3px;
  display: inline-block;
  margin-right: 8px;
  overflow: hidden;
}

.usage-bar .fill {
  height: 100%;
  background: var(--primary-color);
}

.usage-text {
  font-size: 11px;
  color: var(--text-muted);
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
