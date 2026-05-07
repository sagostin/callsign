<template>
  <div class="page-content">
    <!-- View Header -->
    <div class="view-header">
      <div class="view-header-content">
        <h1 class="view-header-title">Extensions</h1>
        <p class="view-header-subtitle">Manage subscriber extensions, devices, and call handling</p>
      </div>
      <div class="view-header-actions">
        <button class="btn btn-primary" @click="$router.push('/admin/extensions/new')">
          <PlusIcon class="btn-icon" />
          <span>New Extension</span>
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
          placeholder="Search extensions..." 
          class="search-box-input"
        >
      </div>
      <select v-model="filterProfile" class="filter-select">
        <option value="">All Profiles</option>
        <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
      <select v-model="filterStatus" class="filter-select">
        <option value="">All Statuses</option>
        <option value="Idle">Idle</option>
        <option value="In Call">In Call</option>
        <option value="Ringing">Ringing</option>
        <option value="Offline">Offline</option>
      </select>
      <button class="btn btn-secondary" @click="$router.push('/admin/extension-profiles')">
        <UsersIcon class="btn-icon" />
        <span>Manage Profiles</span>
      </button>
    </div>

    <!-- Data Table -->
    <DataTable 
      :columns="columns" 
      :data="filteredExtensions" 
      actions
      :pagination="true"
      :page-size="15"
    >
      <template #profile="{ row }">
        <span 
          class="profile-badge" 
          :style="{ background: getProfileColor(row.profileId) }"
        >
          {{ getProfileName(row.profileId) }}
        </span>
      </template>
      
      <template #status="{ value }">
        <StatusBadge :status="value" />
      </template>
      
      <template #device="{ value }">
        <span class="device-text">{{ value || '—' }}</span>
      </template>
      
      <template #actions="{ row }">
        <div class="action-buttons">
          <button 
            class="btn btn-sm btn-ghost" 
            @click="$router.push(`/admin/extensions/${row.id}`)"
          >
            Edit
          </button>
          <button 
            class="btn btn-sm btn-secondary" 
            @click="quickChangeProfile(row)"
          >
            Profile
          </button>
        </div>
      </template>
    </DataTable>

    <!-- Quick Profile Change Modal -->
    <Teleport to="body">
      <div v-if="showQuickProfile" class="modal-overlay" @click.self="showQuickProfile = false">
        <div class="modal modal-sm">
          <div class="modal-header">
            <h3 class="modal-title">Change Profile: {{ selectedExt?.name }}</h3>
            <button class="btn-icon" @click="showQuickProfile = false">
              <XIcon class="icon-sm" />
            </button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Select Profile</label>
              <select v-model="selectedProfileId" class="input">
                <option value="">No Profile</option>
                <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showQuickProfile = false">
              Cancel
            </button>
            <button class="btn btn-primary" @click="applyProfile">
              Apply
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { 
  Search as SearchIcon, 
  X as XIcon, 
  Plus as PlusIcon,
  Users as UsersIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { extensionsAPI, extensionProfilesAPI } from '@/services/api'

const toast = inject('toast')

// State
const isLoading = ref(false)
const searchQuery = ref('')
const filterProfile = ref('')
const filterStatus = ref('')

const columns = [
  { key: 'extension', label: 'Ext', width: '80px', sortable: true },
  { key: 'name', label: 'Name', sortable: true },
  { key: 'profile', label: 'Profile', width: '140px' },
  { key: 'status', label: 'Status', width: '110px', sortable: true },
  { key: 'device', label: 'Device', width: '140px' },
  { key: 'lastCall', label: 'Last Call', width: '100px' }
]

// Data
const extensions = ref([])
const profiles = ref([])

const filteredExtensions = computed(() => {
  return extensions.value.filter(e => {
    const matchesSearch = !searchQuery.value || 
      e.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      e.extension.includes(searchQuery.value)
    const matchesProfile = !filterProfile.value || e.profileId === parseInt(filterProfile.value)
    const matchesStatus = !filterStatus.value || e.status === filterStatus.value
    return matchesSearch && matchesProfile && matchesStatus
  })
})

const getProfileName = (id) => profiles.value.find(p => p.id === id)?.name || 'None'
const getProfileColor = (id) => profiles.value.find(p => p.id === id)?.color || '#94a3b8'

// Fetch data
async function fetchProfiles() {
  try {
    const response = await extensionProfilesAPI.list()
    profiles.value = (response.data || []).map(p => ({
      id: p.id,
      name: p.name,
      color: p.color,
      extensionCount: p.extension_count || 0,
      permissions: p.permissions || {}
    }))
  } catch (error) {
    console.error('Failed to load profiles', error)
  }
}

async function fetchExtensions() {
  isLoading.value = true
  try {
    const response = await extensionsAPI.list()
    const data = response.data || []
    extensions.value = data.map(ext => ({
      id: ext.id,
      ext: ext.extension,
      extension: ext.extension,
      name: ext.effective_caller_id_name || ext.display_name || `Ext ${ext.extension}`,
      profileId: ext.profile_id || null,
      status: getStatusLabel(ext),
      device: ext.device_name || null,
      lastCall: formatLastCall(ext.last_call_at)
    }))
  } catch (error) {
    console.error('Failed to load extensions:', error)
    toast?.error('Failed to load extensions')
  } finally {
    isLoading.value = false
  }
}

function getStatusLabel(ext) {
  if (!ext.registered) return 'Offline'
  if (ext.in_call) return 'In Call'
  if (ext.ringing) return 'Ringing'
  return 'Idle'
}

function formatLastCall(dateStr) {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  const now = new Date()
  const diffMin = Math.floor((now - date) / 60000)
  const diffHours = Math.floor((now - date) / 3600000)
  const diffDays = Math.floor((now - date) / 86400000)
  
  if (diffMin < 1) return 'Now'
  if (diffMin < 60) return `${diffMin}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return `${diffDays}d ago`
}

// Quick Profile Change
const showQuickProfile = ref(false)
const selectedExt = ref(null)
const selectedProfileId = ref('')

const quickChangeProfile = (ext) => {
  selectedExt.value = ext
  selectedProfileId.value = ext.profileId || ''
  showQuickProfile.value = true
}

const applyProfile = async () => {
  if (selectedExt.value) {
    try {
      const newProfileId = selectedProfileId.value ? parseInt(selectedProfileId.value) : null
      await extensionsAPI.update(selectedExt.value.id, { profile_id: newProfileId })
      selectedExt.value.profileId = newProfileId
      toast?.success(`Profile updated for extension ${selectedExt.value.extension}`)
    } catch (error) {
      toast?.error('Failed to update profile')
    }
  }
  showQuickProfile.value = false
}

onMounted(() => {
  Promise.all([fetchExtensions(), fetchProfiles()])
})
</script>

<style scoped>
/* Filter Bar */
.filter-bar {
  display: flex;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-6);
  flex-wrap: wrap;
  align-items: center;
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
}

.search-box-input {
  width: 100%;
  padding: var(--spacing-2) var(--spacing-3) var(--spacing-2) var(--spacing-8);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  background: var(--bg-card);
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
}

/* Profile Badge */
.profile-badge {
  display: inline-block;
  padding: var(--spacing-0-5) var(--spacing-2);
  border-radius: var(--radius-full);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: white;
}

.device-text {
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  color: var(--text-secondary);
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
    align-items: stretch;
  }
  
  .search-box {
    max-width: none;
  }
  
  .filter-select {
    width: 100%;
  }
}
</style>
