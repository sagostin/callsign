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
        <button class="btn-link text-indigo-600">Impersonate</button>
     </template>
  </DataTable>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'

const columns = [
   { key: 'name', label: 'Tenant Name' },
   { key: 'domain', label: 'SIP Domain' },
   { key: 'usage', label: 'Resource Usage' },
   { key: 'status', label: 'Status', width: '100px' }
]

const tenants = ref([
   { id: 1, name: 'Acme Corp', domain: 'acme.callsign.io', extensions: 12, limit_ext: 50, disk: '4.2GB', limit_disk: '10GB', status: 'Active' },
   { id: 2, name: 'Hotel California', domain: 'hcal.callsign.io', extensions: 8, limit_ext: 200, disk: '1.1GB', limit_disk: '50GB', status: 'Active' },
   { id: 3, name: 'StartUp Inc', domain: 'startup.io', extensions: 4, limit_ext: 5, disk: '0.2GB', limit_disk: '1GB', status: 'Suspended' },
])
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 600; cursor: pointer; margin-right: 8px; }
</style>
