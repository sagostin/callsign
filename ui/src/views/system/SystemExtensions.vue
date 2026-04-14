<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Extensions</h2>
      <p class="text-muted text-sm">Global registry of all extensions across tenants.</p>
    </div>
  </div>

  <div class="filter-bar">
    <input 
      v-model="searchQuery" 
      type="text" 
      placeholder="Search extensions..." 
      class="search-input"
    >
    <select v-model="selectedTenant" class="filter-select">
      <option value="">All Tenants</option>
      <option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">
        {{ tenant.name }}
      </option>
    </select>
  </div>

  <DataTable :columns="columns" :data="filteredExtensions" :loading="isLoading" actions>
    <template #tenant="{ value }">
      <span class="badge tenant">{{ value }}</span>
    </template>
    
    <template #status="{ value }">
       <StatusBadge :status="value" />
    </template>

    <template #actions="{ row }">
      <button class="btn-link" @click="handleDebug(row)">Debug</button>
      <button class="btn-link text-bad" @click="handleUnreg(row)">Unreg</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { extensionsAPI, systemAPI } from '../../services/api'

const toast = inject('toast')

// State
const isLoading = ref(false)
const searchQuery = ref('')
const selectedTenant = ref('')
const extensions = ref([])
const tenants = ref([])

const columns = [
  { key: 'ext', label: 'Extension', width: '100px' },
  { key: 'name', label: 'Name' },
  { key: 'tenant', label: 'Tenant' },
  { key: 'type', label: 'Device Type' },
  { key: 'status', label: 'Status' }
]

// Derived data
const filteredExtensions = computed(() => extensions.value.filter(ext => {
  if (!matchesSearch(ext)) return false
  if (selectedTenant.value && ext.tenantId !== selectedTenant.value) return false
  return true
}))

function matchesSearch(ext) {
  if (!searchQuery.value) return true
  const query = searchQuery.value.toLowerCase()
  return ext.ext.includes(query) || ext.name.toLowerCase().includes(query)
}

// Data fetching
async function loadExtensions() {
  isLoading.value = true
  try {
    const params = selectedTenant.value ? { tenant_id: selectedTenant.value } : {}
    const response = await extensionsAPI.list(params)
    const data = response.data?.data || response.data || []
    extensions.value = data.map(parseExtension)
  } catch (error) {
    console.error('Failed to load extensions:', error)
    toast?.error('Failed to load extensions')
  } finally {
    isLoading.value = false
  }
}

function parseExtension(ext) {
  return {
    id: ext.id,
    ext: ext.extension,
    name: ext.effective_caller_id_name || ext.display_name || `Ext ${ext.extension}`,
    tenant: ext.tenant_name || 'Unknown',
    tenantId: ext.tenant_id,
    type: ext.device_name || ext.device_type || 'Unknown',
    status: getStatusLabel(ext)
  }
}

function getStatusLabel(ext) {
  if (!ext.registered) return 'Offline'
  if (ext.in_call) return 'In Call'
  if (ext.ringing) return 'Ringing'
  return 'Idle'
}

// Tenant loading
async function loadTenants() {
  try {
    const response = await systemAPI.listTenants()
    tenants.value = response.data?.data || response.data || []
  } catch (error) {
    console.error('Failed to load tenants:', error)
  }
}

// Button handlers
async function handleDebug(row) {
  try {
    const response = await extensionsAPI.getStatus(row.id)
    const status = response.data
    alert(`Extension ${row.ext} Status:\n${JSON.stringify(status, null, 2)}`)
  } catch (error) {
    console.error('Debug failed:', error)
    toast?.error(`Failed to get status for ${row.ext}`)
  }
}

async function handleUnreg(row) {
  if (!confirm(`Unregister extension ${row.ext}?`)) return
  try {
    // Extension unregistration would call a DELETE or POST endpoint
    // Currently no API exists; report not supported
    toast?.error(`Unregister not supported for system extensions`)
  } catch (error) {
    console.error('Unregister failed:', error)
    toast?.error(`Failed to unregister ${row.ext}`)
  }
}

// Initialize
onMounted(() => {
  loadTenants()
  loadExtensions()
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
