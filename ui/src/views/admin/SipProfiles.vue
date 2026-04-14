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
      <button class="btn-link" @click="restartProfile(row)">Restart</button>
      <button class="btn-link text-bad" @click="stopProfile(row)">Stop</button>
    </template>
  </DataTable>

  <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
      <h3 class="text-lg font-bold mb-4">{{ isEditing ? 'Edit Profile' : 'New SIP Profile' }}</h3>
      
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Profile Name</label>
          <input v-model="activeProfile.profile_name" class="w-full border p-2 rounded text-sm" placeholder="e.g. internal-tls" />
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
import { ref, onMounted } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api.js'

const columns = [
  { key: 'profile_name', label: 'Profile Name' },
  { key: 'ip', label: 'IP Address' },
  { key: 'port', label: 'Port' },
  { key: 'context', label: 'Context' },
  { key: 'status', label: 'Status' }
]

const profiles = ref([])
const showModal = ref(false)
const isEditing = ref(false)
const activeProfile = ref({ profile_name: '', ip: '', port: '5060', context: 'public', tls: false, cert: '' })

// Guard clause: toast may not be defined in some contexts
const toast = typeof window !== 'undefined' ? window.__toast : null

onMounted(async () => {
  await loadProfiles()
})

async function loadProfiles() {
  try {
    const response = await systemAPI.listSIPProfiles()
    profiles.value = response.data || []
  } catch (error) {
    toast?.error('Failed to load SIP profiles', error.message)
    profiles.value = []
  }
}

const openCreate = () => {
  activeProfile.value = { profile_name: '', ip: '', port: '5060', context: 'public', tls: false, cert: '' }
  isEditing.value = false
  showModal.value = true
}

const openEdit = (row) => {
  activeProfile.value = { ...row }
  isEditing.value = true
  showModal.value = true
}

const restartProfile = async (row) => {
  try {
    await systemAPI.restartSofiaProfile(row.profile_name)
    toast?.success(`Profile ${row.profile_name} restart command sent`)
    await loadProfiles()
  } catch (error) {
    toast?.error('Failed to restart profile', error.message)
  }
}

const stopProfile = async (row) => {
  if (!confirm(`Stop profile ${row.profile_name}? Active calls may be dropped.`)) return
  try {
    await systemAPI.stopSofiaProfile(row.profile_name)
    toast?.success(`Profile ${row.profile_name} stop command sent`)
    await loadProfiles()
  } catch (error) {
    toast?.error('Failed to stop profile', error.message)
  }
}

const saveProfile = async () => {
  try {
    if (isEditing.value) {
      await systemAPI.updateSIPProfile(activeProfile.value.id, activeProfile.value)
      toast?.success(`Profile "${activeProfile.value.profile_name}" updated`)
    } else {
      await systemAPI.createSIPProfile(activeProfile.value)
      toast?.success(`Profile "${activeProfile.value.profile_name}" created`)
    }
    showModal.value = false
    await loadProfiles()
  } catch (error) {
    toast?.error('Failed to save profile', error.message)
  }
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
