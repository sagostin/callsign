<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Mailboxes</h2>
      <p class="text-muted text-sm">Manage voicemail boxes across all tenants.</p>
    </div>
  </div>

  <div class="filter-bar">
    <input type="text" v-model="searchQuery" placeholder="Search mailboxes..." class="search-input">
    <select v-model="selectedTenant" class="filter-select">
       <option value="">All Tenants</option>
       <option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">
         {{ tenant.name }}
       </option>
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
import { ref, onMounted, watch } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api.js'
import { voicemailAPI } from '../../services/api.js'

const columns = [
  { key: 'mailbox', label: 'Mailbox', width: '100px' },
  { key: 'user', label: 'User / Name' },
  { key: 'tenant', label: 'Tenant' },
  { key: 'messages', label: 'Msgs', width: '80px' },
  { key: 'usage', label: 'Storage' }
]

const mailboxes = ref([])
const tenants = ref([])
const searchQuery = ref('')
const selectedTenant = ref('')

const loadTenants = async () => {
  const response = await systemAPI.listTenants()
  tenants.value = response.data
}

const loadMailboxes = async () => {
  const params = {}
  if (selectedTenant.value) {
    params.tenant_id = selectedTenant.value
  }
  const response = await voicemailAPI.listBoxes({ params })
  mailboxes.value = response.data
}

onMounted(() => {
  loadTenants()
  loadMailboxes()
})

watch(selectedTenant, () => {
  loadMailboxes()
})
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
