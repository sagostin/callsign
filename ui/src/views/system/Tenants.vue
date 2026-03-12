<template>
  <div class="page-content">
    <!-- View Header -->
    <div class="view-header">
      <div class="view-header-content">
        <h1 class="view-header-title">Tenants</h1>
        <p class="view-header-subtitle">Manage multi-tenant environments, limits, and resource allocation</p>
      </div>
      <div class="view-header-actions">
        <button class="btn btn-primary" @click="$router.push('/system/tenants/new')">
          <PlusIcon class="btn-icon" />
          <span>Create Tenant</span>
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-box-icon" />
        <input 
          type="text" 
          v-model="searchQuery" 
          placeholder="Search tenants..." 
          class="search-box-input"
        >
      </div>
      <select v-model="filterStatus" class="filter-select">
        <option value="">All Statuses</option>
        <option value="active">Active</option>
        <option value="suspended">Suspended</option>
      </select>
    </div>

    <!-- Data Table -->
    <DataTable 
      :columns="columns" 
      :data="filteredTenants" 
      actions
      :pagination="true"
      :page-size="10"
    >
      <template #status="{ value }">
        <StatusBadge :status="value" />
      </template>
      
      <template #usage="{ row }">
        <div class="usage-cell">
          <div class="usage-row">
            <span class="usage-label">Extensions</span>
            <span class="usage-value">{{ row.extensions }} / {{ row.limit_ext }}</span>
          </div>
          <div class="usage-row">
            <span class="usage-label">Storage</span>
            <span class="usage-value">{{ row.disk }} / {{ row.limit_disk }}</span>
          </div>
        </div>
      </template>
      
      <template #actions="{ row }">
        <div class="action-buttons">
          <button class="btn-icon" @click="$router.push(`/system/tenants/${row.id}`)" title="Manage tenant">
            <SettingsIcon class="icon-sm" />
          </button>
          <button class="btn-icon" @click="impersonate(row)" title="Impersonate tenant">
            <EyeIcon class="icon-sm" />
          </button>
        </div>
      </template>
    </DataTable>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api'
import { 
  Plus as PlusIcon, 
  Search as SearchIcon,
  Settings as SettingsIcon,
  Eye as EyeIcon
} from 'lucide-vue-next'

const columns = [
  { key: 'name', label: 'Tenant Name', sortable: true },
  { key: 'domain', label: 'SIP Domain', sortable: true },
  { key: 'usage', label: 'Resource Usage' },
  { key: 'status', label: 'Status', width: '120px', sortable: true }
]

const tenants = ref([])
const searchQuery = ref('')
const filterStatus = ref('')
const loading = ref(true)

const filteredTenants = computed(() => {
  return tenants.value.filter(t => {
    const matchesSearch = !searchQuery.value || 
      t.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      t.domain.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesStatus = !filterStatus.value || 
      t.status.toLowerCase() === filterStatus.value.toLowerCase()
    
    return matchesSearch && matchesStatus
  })
})

const loadTenants = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listTenants()
    const data = response.data || []
    tenants.value = data.map(t => ({
      id: t.id,
      name: t.name,
      domain: t.domain || t.name.toLowerCase().replace(/\s+/g, '-') + '.callsign.io',
      extensions: t.extension_count || 0,
      limit_ext: t.profile?.max_extensions || 50,
      disk: t.disk_usage || '0GB',
      limit_disk: (t.profile?.max_disk_gb || 10) + 'GB',
      status: t.enabled !== false ? 'Active' : 'Suspended'
    }))
  } catch (e) {
    console.error('Failed to load tenants:', e)
  } finally {
    loading.value = false
  }
}

const impersonate = (tenant) => {
  localStorage.setItem('tenantId', tenant.id)
  localStorage.setItem('tenantName', tenant.name)
  window.location.href = '/admin'
}

onMounted(loadTenants)
</script>

<style scoped>
/* Filter Bar Customization */
.filter-bar {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-6);
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 240px;
  max-width: 360px;
}

.search-box-icon {
  position: absolute;
  left: var(--spacing-3);
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--text-muted);
  pointer-events: none;
}

.search-box-input {
  width: 100%;
  padding: var(--spacing-2) var(--spacing-3) var(--spacing-2) var(--spacing-8);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  background: var(--bg-card);
  transition: all var(--transition-fast);
}

.search-box-input:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-light);
  outline: none;
}

.filter-select {
  padding: var(--spacing-2) var(--spacing-3);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  background: var(--bg-card);
  min-width: 140px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-select:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-light);
  outline: none;
}

/* Usage Cell */
.usage-cell {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-1);
}

.usage-row {
  display: flex;
  justify-content: space-between;
  gap: var(--spacing-4);
  font-size: var(--text-xs);
}

.usage-label {
  color: var(--text-muted);
}

.usage-value {
  color: var(--text-primary);
  font-weight: var(--font-medium);
  font-family: var(--font-mono);
}

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: var(--spacing-1);
  justify-content: flex-end;
}

.icon-sm {
  width: 16px;
  height: 16px;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .filter-bar {
    flex-direction: column;
  }
  
  .search-box {
    max-width: none;
  }
  
  .filter-select {
    width: 100%;
  }
  
  .usage-cell {
    gap: var(--spacing-0-5);
  }
}

@media (max-width: 480px) {
  .action-buttons {
    flex-direction: column;
    gap: var(--spacing-1);
  }
  
  .action-buttons .btn-icon {
    width: 32px;
    height: 32px;
  }
}
</style>
