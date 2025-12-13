<template>
  <div class="view-header">
     <div class="header-content">
       <h2>Tenants</h2>
       <p class="text-muted text-sm">Manage multi-tenant environments, limits, and resource allocation.</p>
     </div>
     <div class="header-actions">
        <button class="btn-primary" @click="$router.push('/system/tenants/new')">+ Create Tenant</button>
     </div>
  </div>

  <DataTable :columns="columns" :data="tenants" actions>
     <template #status="{ value }">
       <StatusBadge :status="value" />
     </template>
     <template #usage="{ row }">
        <div class="flex flex-col gap-1">
           <div class="text-xs font-mono">Ext: {{ row.extensions }} / {{ row.limit_ext }}</div>
           <div class="text-xs font-mono">Disk: {{ row.disk }} / {{ row.limit_disk }}</div>
        </div>
     </template>
     <template #actions="{ row }">
        <button class="btn-link" @click="$router.push(`/system/tenants/${row.id}`)">Manage</button>
        <button class="btn-link text-indigo-600" @click="impersonate(row)">Impersonate</button>
     </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api'

const columns = [
   { key: 'name', label: 'Tenant Name' },
   { key: 'domain', label: 'SIP Domain' },
   { key: 'usage', label: 'Resource Usage' },
   { key: 'status', label: 'Status', width: '100px' }
]

const tenants = ref([])
const loading = ref(true)

const loadTenants = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listTenants()
    const data = response.data.data || response.data || []
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
  // Set the tenant ID in localStorage (same key used by TopBar and auth service)
  localStorage.setItem('tenantId', tenant.id)
  
  // Also store tenant name for display purposes (optional, used in some UI components)
  localStorage.setItem('tenantName', tenant.name)
  
  // Navigate to the tenant admin panel
  window.location.href = '/admin'
}

onMounted(loadTenants)
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 600; cursor: pointer; margin-right: 8px; }
</style>
