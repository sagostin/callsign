<template>
  <div class="view-header">
    <div class="header-content">
      <h2>SIP Profiles</h2>
      <p class="text-muted text-sm">Manage SIP sockets, ports, and contexts.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="openCreate">+ New Profile</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="profiles" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    <template #actions="{ row }">
      <button class="btn-link" @click="openEdit(row)">Edit</button>
      <button class="btn-link">Restart</button>
      <button class="btn-link text-bad">Stop</button>
    </template>
  </DataTable>

  <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
      <h3 class="text-lg font-bold mb-4">{{ isEditing ? 'Edit Profile' : 'New SIP Profile' }}</h3>
      
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Profile Name</label>
          <input v-model="activeProfile.name" class="w-full border p-2 rounded text-sm" placeholder="e.g. internal-tls" />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-bold text-gray-500 uppercase mb-1">IP Address</label>
            <input v-model="activeProfile.ip" class="w-full border p-2 rounded text-sm" placeholder="1.2.3.4" />
          </div>
          <div>
            <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Port</label>
            <input v-model="activeProfile.port" class="w-full border p-2 rounded text-sm" placeholder="5060" />
          </div>
        </div>
        
        <div class="bg-gray-50 p-3 rounded border border-gray-200">
          <div class="flex items-center gap-2 mb-2">
             <input type="checkbox" v-model="activeProfile.tls" id="useTls" class="rounded text-indigo-600" />
             <label for="useTls" class="text-sm font-medium">Enable TLS (SIP over SSL)</label>
          </div>
          <div v-if="activeProfile.tls">
             <label class="block text-xs font-bold text-gray-500 uppercase mb-1">SSL Certificate Path</label>
             <input v-model="activeProfile.cert" class="w-full border p-2 rounded text-sm" placeholder="/etc/ssl/..." />
             <p class="text-xs text-gray-400 mt-1">Tenant SSL certs can be managed in Tenant Settings.</p>
          </div>
        </div>
      </div>

      <div class="flex justify-end gap-2 mt-6">
        <button class="btn-link" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveProfile">Save Profile</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'

const columns = [
  { key: 'name', label: 'Profile Name' },
  { key: 'ip', label: 'IP Address' },
  { key: 'port', label: 'Port' },
  { key: 'context', label: 'Context' },
  { key: 'status', label: 'Status' }
]

const profiles = ref([
  { id: 1, name: 'internal', ip: '10.0.0.5', port: '5060', context: 'public', status: 'Running', tls: false },
  { id: 2, name: 'external', ip: '10.0.0.5', port: '5080', context: 'public', status: 'Running', tls: false },
  { id: 3, name: 'internal-ipv6', ip: '::1', port: '5060', context: 'public', status: 'Stopped', tls: false },
  { id: 4, name: 'internal-tls', ip: '10.0.0.5', port: '5061', context: 'public', status: 'Running', tls: true, cert: '/etc/ssl/certs/sip_cert.pem' },
])

const showModal = ref(false)
const isEditing = ref(false)
const activeProfile = ref({ name: '', ip: '', port: '5060', context: 'public', tls: false, cert: '' })

const openCreate = () => {
  activeProfile.value = { name: '', ip: '', port: '5060', context: 'public', tls: false, cert: '' }
  isEditing.value = false
  showModal.value = true
}

const openEdit = (row) => {
  activeProfile.value = { ...row }
  isEditing.value = true
  showModal.value = true
}

const saveProfile = () => {
  if (isEditing.value) {
    const idx = profiles.value.findIndex(p => p.id === activeProfile.value.id)
    if (idx !== -1) profiles.value[idx] = { ...activeProfile.value }
  } else {
    profiles.value.push({ ...activeProfile.value, id: Date.now(), status: 'Stopped' })
  }
  showModal.value = false
}
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
</style>
