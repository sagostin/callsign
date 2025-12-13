<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Extensions</h2>
      <p class="text-muted text-sm">Manage subscriber extensions, devices, and call handling.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/extensions/new')">+ New Extension</button>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">All Extensions</button>
    <button class="tab" :class="{ active: viewMode === 'profiles' }" @click="viewMode = 'profiles'">Extension Profiles</button>
  </div>

  <!-- EXTENSIONS LIST -->
  <div class="tab-content" v-if="viewMode === 'list'">
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input type="text" v-model="searchQuery" placeholder="Search extensions..." class="search-input">
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
    </div>

    <DataTable :columns="columns" :data="filteredExtensions" actions>
      <template #profile="{ row }">
        <span class="profile-badge" :style="{ background: getProfileColor(row.profileId) }">
          {{ getProfileName(row.profileId) }}
        </span>
      </template>
      <template #status="{ value }">
        <StatusBadge :status="value" />
      </template>
      <template #device="{ value }">
        <span class="font-mono text-xs">{{ value || '—' }}</span>
      </template>
      <template #actions="{ row }">
        <button class="btn-link" @click="$router.push(`/admin/extensions/${row.id}`)">Edit</button>
        <button class="btn-link" @click="quickChangeProfile(row)">Profile</button>
      </template>
    </DataTable>
  </div>
  
  <!-- EXTENSION PROFILES -->
  <div class="tab-content profiles-panel" v-else>
    <div class="panel-header">
      <div>
        <h3>Extension Profiles</h3>
        <p class="text-muted text-sm">Define permission sets and calling rules. Profile rules take precedence over global routing rules.</p>
      </div>
      <button class="btn-primary" @click="showProfileModal = true">+ New Profile</button>
    </div>

    <!-- Empty State -->
    <div v-if="profiles.length === 0" class="empty-state">
      <div class="empty-icon">👤</div>
      <h4>No Extension Profiles</h4>
      <p>Create extension profiles to define permission sets and calling rules that can be applied to multiple extensions.</p>
      <button class="btn-primary" @click="showProfileModal = true">Create First Profile</button>
    </div>

    <div v-else class="profiles-grid">
      <div class="profile-card" v-for="profile in profiles" :key="profile.id">
        <div class="profile-header">

          <div class="profile-icon" :style="{ background: profile.color }">
            {{ profile.name.charAt(0) }}
          </div>
          <div class="profile-info">
            <h4>{{ profile.name }}</h4>
            <span class="profile-count">{{ profile.extensionCount }} extensions</span>
          </div>
          <div class="profile-actions">
            <button class="btn-icon" @click="editProfile(profile)"><EditIcon class="icon-sm" /></button>
            <button class="btn-icon" @click="deleteProfile(profile)"><TrashIcon class="icon-sm text-bad" /></button>
          </div>
        </div>
        
        <div class="profile-permissions">
          <div class="perm-row">
            <span class="perm-label">Outbound Calls</span>
            <span class="perm-value" :class="{ enabled: profile.permissions.outbound }">
              {{ profile.permissions.outbound ? 'Allowed' : 'Blocked' }}
            </span>
          </div>
          <div class="perm-row highlight" v-if="profile.permissions.international">
            <span class="perm-label">International Dialing</span>
            <span class="perm-value enabled">Allowed</span>
          </div>
          <div class="perm-row" v-else>
            <span class="perm-label">International Dialing</span>
            <span class="perm-value blocked">Blocked</span>
          </div>
          <div class="perm-row">
            <span class="perm-label">Recording</span>
            <span class="perm-value" :class="{ enabled: profile.permissions.recording }">
              {{ profile.permissions.recording ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
          <div class="perm-row">
            <span class="perm-label">Portal Access</span>
            <span class="perm-value" :class="{ enabled: profile.permissions.portal }">
              {{ profile.permissions.portal ? 'Allowed' : 'Blocked' }}
            </span>
          </div>
        </div>

        <div class="profile-routing" v-if="profile.routingOverride">
          <div class="routing-label">
            <RouteIcon class="icon-xs" />
            <span>Custom Routing</span>
          </div>
          <p class="routing-desc">{{ profile.routingOverride }}</p>
        </div>

        <div class="profile-handling" v-if="profile.callHandling?.overrideStrategy">
          <div class="routing-label">
            <PhoneIcon class="icon-xs" />
            <span>Call Handling Override</span>
          </div>
          <p class="routing-desc">{{ profile.callHandling.strategy === 'simultaneous' ? 'Ring All' : 'Sequential' }}</p>
        </div>
      </div>
    </div>
  </div>

  <!-- Profile Modal -->
  <div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ isEditingProfile ? 'Edit Profile' : 'New Extension Profile' }}</h3>
        <button class="btn-icon" @click="showProfileModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Profile Name</label>
          <input v-model="profileForm.name" class="input-field" placeholder="e.g. Sales Team">
        </div>

        <div class="form-group">
          <label>Profile Color</label>
          <div class="color-picker">
            <button v-for="c in colorOptions" :key="c" 
              class="color-swatch" 
              :style="{ background: c }"
              :class="{ selected: profileForm.color === c }"
              @click="profileForm.color = c">
            </button>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Permissions</h4>
          <div class="perm-toggles">
            <label class="toggle-row">
              <input type="checkbox" v-model="profileForm.permissions.outbound">
              <span>Allow Outbound Calls</span>
            </label>
            <label class="toggle-row highlight-intl">
              <input type="checkbox" v-model="profileForm.permissions.international">
              <span>Allow International Calls</span>
              <span class="badge-intl">Premium</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" v-model="profileForm.permissions.recording">
              <span>Allow Call Recording</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" v-model="profileForm.permissions.portal">
              <span>Allow User Portal Access</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" v-model="profileForm.permissions.voicemail">
              <span>Allow Voicemail Config</span>
            </label>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Call Handling Override (Optional)</h4>
          <p class="help-text">Override call handling settings for extensions with this profile.</p>
          
          <label class="toggle-row">
            <input type="checkbox" v-model="profileForm.callHandling.overrideStrategy">
            <span>Override Ring Strategy</span>
          </label>
          
          <div v-if="profileForm.callHandling.overrideStrategy" class="override-options">
            <div class="strategy-radio">
              <label>
                <input type="radio" v-model="profileForm.callHandling.strategy" value="simultaneous">
                <span>Ring All Simultaneously</span>
              </label>
              <label>
                <input type="radio" v-model="profileForm.callHandling.strategy" value="sequential">
                <span>Ring Sequentially</span>
              </label>
            </div>
          </div>

          <label class="toggle-row" style="margin-top: 12px;">
            <input type="checkbox" v-model="profileForm.callHandling.overrideDevices">
            <span>Override Device Availability</span>
          </label>

          <div v-if="profileForm.callHandling.overrideDevices" class="override-options">
            <label class="toggle-row sub">
              <input type="checkbox" v-model="profileForm.callHandling.devices.softphone">
              <span>Allow Softphone</span>
            </label>
            <label class="toggle-row sub">
              <input type="checkbox" v-model="profileForm.callHandling.devices.deskPhone">
              <span>Allow Desk Phone</span>
            </label>
            <label class="toggle-row sub">
              <input type="checkbox" v-model="profileForm.callHandling.devices.mobile">
              <span>Allow Mobile Phone</span>
            </label>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Routing Override (Optional)</h4>
          <p class="help-text">If set, this overrides global routing rules for extensions with this profile.</p>
          <textarea v-model="profileForm.routingOverride" class="input-field" rows="2" 
            placeholder="e.g. Route through international gateway for all outbound"></textarea>
        </div>

        <!-- Profile-level Call Handling Rules (only when editing) -->
        <div v-if="isEditingProfile && profileForm.id" class="form-section">
          <div class="divider"></div>
          <CallHandlingRules 
            :profileId="profileForm.id" 
            :api="extensionProfilesAPI" 
          />
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showProfileModal = false">Cancel</button>
        <button class="btn-primary" @click="saveProfile" :disabled="!profileForm.name">
          {{ isEditingProfile ? 'Save Changes' : 'Create Profile' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Quick Profile Change Modal -->
  <div v-if="showQuickProfile" class="modal-overlay" @click.self="showQuickProfile = false">
    <div class="modal-card small">
      <div class="modal-header">
        <h3>Change Profile: {{ selectedExt?.name }}</h3>
        <button class="btn-icon" @click="showQuickProfile = false"><XIcon class="icon-sm" /></button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>Select Profile</label>
          <select v-model="selectedProfileId" class="input-field">
            <option value="">No Profile</option>
            <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn-secondary" @click="showQuickProfile = false">Cancel</button>
        <button class="btn-primary" @click="applyProfile">Apply</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { Search as SearchIcon, Edit as EditIcon, Trash2 as TrashIcon, X as XIcon, GitMerge as RouteIcon, Phone as PhoneIcon } from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import CallHandlingRules from '../../components/common/CallHandlingRules.vue'
import { extensionsAPI, extensionProfilesAPI } from '@/services/api'

// Toast notifications
const toast = inject('toast')

// State
const isLoading = ref(false)
const viewMode = ref('list')
const searchQuery = ref('')
const filterProfile = ref('')
const filterStatus = ref('')

const columns = [
  { key: 'extension', label: 'Ext', width: '80px' },
  { key: 'name', label: 'Name' },
  { key: 'profile', label: 'Profile', width: '140px' },
  { key: 'status', label: 'Status', width: '100px' },
  { key: 'device', label: 'Device', width: '140px' },
  { key: 'lastCall', label: 'Last Call', width: '100px' }
]

// Data from API
const extensions = ref([])
const profiles = ref([])

// Fetch extensions and profiles on mount
onMounted(async () => {
  await Promise.all([fetchExtensions(), fetchProfiles()])
})

async function fetchProfiles() {
  try {
    const response = await extensionProfilesAPI.list()
    profiles.value = (response.data.data || []).map(p => {
      const ch = p.call_handling || {}
      return {
        id: p.id,
        name: p.name,
        color: p.color,
        extensionCount: p.extension_count || 0,
        permissions: p.permissions || {},
        callHandling: {
          overrideStrategy: ch.override_strategy || false,
          strategy: ch.strategy || 'simultaneous',
          overrideDevices: ch.override_devices || false,
          devices: {
            softphone: ch.devices?.softphone !== false,
            deskPhone: ch.devices?.desk_phone !== false,
            mobile: ch.devices?.mobile !== false
          }
        },
        routingOverride: p.routing_override || ''
      }
    })
  } catch (error) {
    console.error('Failed to load profiles', error)
  }
}


async function fetchExtensions() {
  isLoading.value = true
  try {
    const response = await extensionsAPI.list()
    // Handle both {data: [...]} wrapper and direct array formats
    const data = response.data?.data || response.data || []
    extensions.value = data.map(ext => ({
      id: ext.id,
      ext: ext.extension, // for router link
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
    extensions.value = []
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
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)
  
  if (diffMins < 1) return 'Now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return `${diffDays}d ago`
}

const filteredExtensions = computed(() => {
  return extensions.value.filter(e => {
    const matchesSearch = e.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          e.extension.includes(searchQuery.value)
    const matchesProfile = !filterProfile.value || e.profileId === parseInt(filterProfile.value)
    const matchesStatus = !filterStatus.value || e.status === filterStatus.value
    return matchesSearch && matchesProfile && matchesStatus
  })
})

const getProfileName = (id) => profiles.value.find(p => p.id === id)?.name || 'None'
const getProfileColor = (id) => profiles.value.find(p => p.id === id)?.color || '#94a3b8'

// Profile Modal
const showProfileModal = ref(false)
const isEditingProfile = ref(false)
const profileForm = ref({
  id: null,
  name: '',
  color: '#6366f1',
  permissions: { outbound: true, international: false, recording: true, portal: true, voicemail: true },
  routingOverride: '',
  callHandling: {
    overrideStrategy: false,
    strategy: 'simultaneous',
    overrideDevices: false,
    devices: { softphone: true, deskPhone: true, mobile: true }
  }
})

const colorOptions = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#64748b']

const editProfile = (profile) => {
  profileForm.value = JSON.parse(JSON.stringify(profile))
  isEditingProfile.value = true
  showProfileModal.value = true
}

const deleteProfile = async (profile) => {
  if (confirm(`Delete profile "${profile.name}"? Extensions will be unassigned.`)) {
    try {
      await extensionProfilesAPI.delete(profile.id)
      await fetchProfiles()
      toast?.success(`Profile "${profile.name}" deleted`)
    } catch (error) {
      toast?.error(error.message, 'Failed to delete profile')
    }
  }
}

const saveProfile = async () => {
  try {
    const ch = profileForm.value.callHandling || {}
    const payload = {
      name: profileForm.value.name,
      color: profileForm.value.color,
      permissions: profileForm.value.permissions,
      call_handling: {
        override_strategy: ch.overrideStrategy || false,
        strategy: ch.strategy || 'simultaneous',
        override_devices: ch.overrideDevices || false,
        devices: {
          softphone: ch.devices?.softphone !== false,
          desk_phone: ch.devices?.deskPhone !== false,
          mobile: ch.devices?.mobile !== false
        }
      },
      routing_override: profileForm.value.routingOverride || ''
    }
    
    if (isEditingProfile.value) {
      await extensionProfilesAPI.update(profileForm.value.id, payload)
      toast?.success('Profile updated')
    } else {
      await extensionProfilesAPI.create(payload)
      toast?.success('Profile created')
    }
    await fetchProfiles()
    showProfileModal.value = false
    resetProfileForm()
  } catch (error) {
    toast?.error(error.message, 'Failed to save profile')
  }
}


const resetProfileForm = () => {
  profileForm.value = {
    id: null,
    name: '',
    color: '#6366f1',
    permissions: { outbound: true, international: false, recording: true, portal: true, voicemail: true },
    routingOverride: ''
  }
  isEditingProfile.value = false
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
      toast?.error(error.message, 'Failed to update profile')
    }
  }
  showQuickProfile.value = false
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
  cursor: pointer;
}
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  font-weight: 500;
  cursor: pointer;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
}
.btn-icon:hover { color: var(--text-primary); }

.tabs {
  display: flex;
  gap: 2px;
  border-bottom: 1px solid var(--border-color);
}

.tab {
  padding: 8px 16px;
  background: transparent;
  border: 1px solid transparent;
  border-bottom: none;
  cursor: pointer;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-muted);
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
}

.tab.active {
  background: white;
  border-color: var(--border-color);
  color: var(--primary-color);
  margin-bottom: -1px;
}

.tab-content {
  background: white;
  border: 1px solid var(--border-color);
  border-top: none;
  border-radius: 0 0 var(--radius-md) var(--radius-md);
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-app);
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 280px;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 34px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  background: white;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  background: white;
}

.profile-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 600;
  color: white;
}

/* Profiles Panel */
.profiles-panel {
  padding: var(--spacing-lg);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-lg);
}

.panel-header h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-lg);
}

.profile-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-app);
}

.profile-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 18px;
}

.profile-info {
  flex: 1;
}

.profile-info h4 {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}

.profile-count {
  font-size: 11px;
  color: var(--text-muted);
}

.profile-actions {
  display: flex;
  gap: 4px;
}

.profile-permissions {
  padding: 12px 16px;
}

.perm-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}

.perm-row.highlight {
  background: #fef3c7;
  margin: 4px -16px;
  padding: 6px 16px;
}

.perm-label {
  color: var(--text-muted);
}

.perm-value {
  font-weight: 500;
}
.perm-value.enabled { color: #16a34a; }
.perm-value.blocked { color: #dc2626; }

.profile-routing {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  background: #f0f9ff;
}

.routing-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  color: #0369a1;
  margin-bottom: 4px;
}

.routing-desc {
  font-size: 12px;
  color: #0c4a6e;
  margin: 0;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  padding: 24px;
}

.modal-card {
  background: white;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-card.small {
  max-width: 360px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 700;
  margin: 0;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.form-section {
  margin-bottom: 8px;
}

.form-section h4 {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--text-primary);
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 14px;
}

textarea.input-field {
  resize: vertical;
  min-height: 60px;
}

.color-picker {
  display: flex;
  gap: 8px;
}

.color-swatch {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 2px solid transparent;
  cursor: pointer;
}
.color-swatch.selected {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 2px white, 0 0 0 4px var(--text-primary);
}

.divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
}

.perm-toggles {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  cursor: pointer;
}

.toggle-row.highlight-intl {
  background: #fef3c7;
  padding: 8px;
  margin: -4px -8px;
  border-radius: 4px;
}

.badge-intl {
  font-size: 10px;
  background: #f59e0b;
  color: white;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
  margin-left: auto;
}

.help-text {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 12px; height: 12px; }
.text-bad { color: var(--status-bad); }
.font-mono { font-family: monospace; }
.text-xs { font-size: 11px; }

.profile-handling {
  margin-top: 12px;
  padding: 10px;
  background: #dbeafe;
  border-radius: var(--radius-sm);
}

.override-options {
  margin: 12px 0 0 24px;
  padding: 12px;
  background: var(--bg-app);
  border-radius: var(--radius-sm);
}

.strategy-radio {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.strategy-radio label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  cursor: pointer;
}

.toggle-row.sub {
  margin-left: 12px;
  font-size: 12px;
}
</style>
