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

    <!-- Tabs -->
    <div class="tabs-container">
      <div class="tabs">
        <button 
          class="tab" 
          :class="{ active: viewMode === 'list' }" 
          @click="viewMode = 'list'"
        >
          <ListIcon class="tab-icon" />
          <span>All Extensions</span>
        </button>
        <button 
          class="tab" 
          :class="{ active: viewMode === 'profiles' }" 
          @click="viewMode = 'profiles'"
        >
          <UsersIcon class="tab-icon" />
          <span>Extension Profiles</span>
        </button>
      </div>
    </div>

    <!-- Extensions List Tab -->
    <div v-if="viewMode === 'list'" class="tab-content">
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
    </div>
    
    <!-- Extension Profiles Tab -->
    <div v-else class="tab-content profiles-panel">
      <div class="panel-header">
        <div class="panel-header-content">
          <h3 class="panel-title">Extension Profiles</h3>
          <p class="panel-subtitle">Define permission sets and calling rules for extensions</p>
        </div>
        <button class="btn btn-primary" @click="showProfileModal = true">
          <PlusIcon class="btn-icon" />
          <span>New Profile</span>
        </button>
      </div>

      <!-- Empty State -->
      <div v-if="profiles.length === 0" class="empty-state">
        <div class="empty-icon-wrapper">
          <UsersIcon class="empty-icon" />
        </div>
        <h4 class="empty-title">No Extension Profiles</h4>
        <p class="empty-text">Create extension profiles to define permission sets and calling rules that can be applied to multiple extensions.</p>
        <button class="btn btn-primary" @click="showProfileModal = true">
          Create First Profile
        </button>
      </div>

      <!-- Profiles Grid -->
      <div v-else class="profiles-grid">
        <div 
          v-for="profile in profiles" 
          :key="profile.id" 
          class="profile-card"
        >
          <div class="profile-header">
            <div class="profile-icon" :style="{ background: profile.color }">
              {{ profile.name.charAt(0) }}
            </div>
            <div class="profile-info">
              <h4 class="profile-name">{{ profile.name }}</h4>
              <span class="profile-count">{{ profile.extensionCount }} extensions</span>
            </div>
            <div class="profile-actions">
              <button class="btn-icon" @click="editProfile(profile)" title="Edit profile">
                <EditIcon class="icon-sm" />
              </button>
              <button class="btn-icon text-bad" @click="deleteProfile(profile)" title="Delete profile">
                <TrashIcon class="icon-sm" />
              </button>
            </div>
          </div>
          
          <div class="profile-permissions">
            <div 
              v-for="(perm, key) in permissionLabels" 
              :key="key"
              class="perm-row"
              :class="{ highlight: key === 'international' && profile.permissions[key] }"
            >
              <span class="perm-label">{{ perm.label }}</span>
              <span 
                class="perm-value" 
                :class="{ enabled: profile.permissions[key], blocked: !profile.permissions[key] }"
              >
                {{ profile.permissions[key] ? perm.enabled : perm.disabled }}
              </span>
            </div>
          </div>

          <div v-if="profile.routingOverride" class="profile-routing">
            <div class="routing-label">
              <RouteIcon class="icon-xs" />
              <span>Custom Routing</span>
            </div>
            <p class="routing-desc">{{ profile.routingOverride }}</p>
          </div>

          <div v-if="profile.callHandling?.overrideStrategy" class="profile-handling">
            <div class="handling-label">
              <PhoneIcon class="icon-xs" />
              <span>Call Handling Override</span>
            </div>
            <p class="handling-desc">
              {{ profile.callHandling.strategy === 'simultaneous' ? 'Ring All' : 'Sequential' }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Profile Modal -->
    <Teleport to="body">
      <div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h3 class="modal-title">
              {{ isEditingProfile ? 'Edit Profile' : 'New Extension Profile' }}
            </h3>
            <button class="btn-icon" @click="showProfileModal = false">
              <XIcon class="icon-sm" />
            </button>
          </div>
          
          <div class="modal-body">
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">Profile Name</label>
                <input 
                  v-model="profileForm.name" 
                  class="input" 
                  placeholder="e.g. Sales Team"
                >
              </div>

              <div class="form-group">
                <label class="form-label">Profile Color</label>
                <div class="color-picker">
                  <button 
                    v-for="c in colorOptions" 
                    :key="c" 
                    class="color-swatch" 
                    :style="{ background: c }"
                    :class="{ selected: profileForm.color === c }"
                    @click="profileForm.color = c"
                  >
                    <CheckIcon v-if="profileForm.color === c" class="check-icon" />
                  </button>
                </div>
              </div>

              <div class="form-group full-width">
                <div class="divider"></div>
              </div>

              <div class="form-group full-width">
                <h4 class="section-subtitle">Permissions</h4>
                <div class="perm-toggles">
                  <label 
                    v-for="(perm, key) in permissionLabels" 
                    :key="key"
                    class="toggle-row"
                    :class="{ 'highlight-intl': key === 'international' }"
                  >
                    <input 
                      type="checkbox" 
                      v-model="profileForm.permissions[key]"
                    >
                    <span>{{ perm.label }}</span>
                    <span v-if="key === 'international'" class="badge-premium">Premium</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showProfileModal = false">
              Cancel
            </button>
            <button 
              class="btn btn-primary" 
              @click="saveProfile" 
              :disabled="!profileForm.name"
            >
              {{ isEditingProfile ? 'Save Changes' : 'Create Profile' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

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
  Edit as EditIcon, 
  Trash2 as TrashIcon, 
  X as XIcon, 
  GitMerge as RouteIcon, 
  Phone as PhoneIcon,
  Plus as PlusIcon,
  List as ListIcon,
  Users as UsersIcon,
  Check as CheckIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { extensionsAPI, extensionProfilesAPI } from '@/services/api'

const toast = inject('toast')

// State
const isLoading = ref(false)
const viewMode = ref('list')
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

const permissionLabels = {
  outbound: { label: 'Outbound Calls', enabled: 'Allowed', disabled: 'Blocked' },
  international: { label: 'International Dialing', enabled: 'Allowed', disabled: 'Blocked' },
  recording: { label: 'Call Recording', enabled: 'Enabled', disabled: 'Disabled' },
  portal: { label: 'Portal Access', enabled: 'Allowed', disabled: 'Blocked' },
  voicemail: { label: 'Voicemail Config', enabled: 'Allowed', disabled: 'Blocked' }
}

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

// Profile Modal
const showProfileModal = ref(false)
const isEditingProfile = ref(false)
const profileForm = ref({
  id: null,
  name: '',
  color: '#6366f1',
  permissions: { outbound: true, international: false, recording: true, portal: true, voicemail: true }
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
      toast?.error('Failed to delete profile')
    }
  }
}

const saveProfile = async () => {
  try {
    const payload = {
      name: profileForm.value.name,
      color: profileForm.value.color,
      permissions: profileForm.value.permissions
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
    toast?.error('Failed to save profile')
  }
}

const resetProfileForm = () => {
  profileForm.value = {
    id: null,
    name: '',
    color: '#6366f1',
    permissions: { outbound: true, international: false, recording: true, portal: true, voicemail: true }
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
/* Tabs Container */
.tabs-container {
  margin-bottom: var(--spacing-6);
}

.tabs {
  display: flex;
  gap: var(--spacing-0-5);
  border-bottom: 1px solid var(--border-color);
}

.tab {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-3) var(--spacing-4);
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
  margin-bottom: -1px;
}

.tab:hover {
  color: var(--text-primary);
}

.tab.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}

.tab-icon {
  width: 16px;
  height: 16px;
}

/* Filter Bar */
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

/* Profiles Panel */
.profiles-panel {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-6);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-6);
  flex-wrap: wrap;
  gap: var(--spacing-4);
}

.panel-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0 0 var(--spacing-1) 0;
}

.panel-subtitle {
  font-size: var(--text-sm);
  color: var(--text-secondary);
}

/* Profiles Grid */
.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-5);
}

.profile-card {
  background: var(--bg-hover);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: box-shadow var(--transition-fast);
}

.profile-card:hover {
  box-shadow: var(--shadow-md);
}

.profile-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-4);
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
}

.profile-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: var(--font-bold);
  font-size: var(--text-lg);
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.profile-count {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.profile-actions {
  display: flex;
  gap: var(--spacing-1);
}

.profile-permissions {
  padding: var(--spacing-3) var(--spacing-4);
}

.perm-row {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-2) 0;
  font-size: var(--text-sm);
  border-bottom: 1px solid var(--border-light);
}

.perm-row:last-child {
  border-bottom: none;
}

.perm-row.highlight {
  background: var(--status-warn-subtle);
  margin: var(--spacing-1) calc(-1 * var(--spacing-4));
  padding: var(--spacing-2) var(--spacing-4);
  border-bottom: none;
  border-radius: var(--radius-sm);
}

.perm-label {
  color: var(--text-secondary);
}

.perm-value {
  font-weight: var(--font-medium);
}

.perm-value.enabled { color: var(--status-good); }
.perm-value.blocked { color: var(--text-muted); }

.profile-routing,
.profile-handling {
  padding: var(--spacing-3) var(--spacing-4);
  background: var(--primary-subtle);
  border-top: 1px solid var(--border-color);
}

.profile-handling {
  background: var(--secondary-light);
}

.routing-label,
.handling-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-1-5);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--primary-text);
  margin-bottom: var(--spacing-1);
}

.handling-label {
  color: var(--secondary-color);
}

.routing-desc,
.handling-desc {
  font-size: var(--text-sm);
  color: var(--text-main);
  margin: 0;
}

/* Color Picker */
.color-picker {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}

.color-swatch {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  border: 2px solid transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.color-swatch:hover {
  transform: scale(1.1);
}

.color-swatch.selected {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 2px white, 0 0 0 4px var(--text-primary);
}

.check-icon {
  width: 16px;
  height: 16px;
  color: white;
  filter: drop-shadow(0 1px 2px rgba(0,0,0,0.3));
}

/* Permission Toggles */
.perm-toggles {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  font-size: var(--text-sm);
  cursor: pointer;
  padding: var(--spacing-2);
  border-radius: var(--radius-md);
  transition: background var(--transition-fast);
}

.toggle-row:hover {
  background: var(--bg-hover);
}

.toggle-row.highlight-intl {
  background: var(--status-warn-subtle);
}

.toggle-row input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: var(--primary-color);
}

.badge-premium {
  margin-left: auto;
  font-size: var(--text-2xs);
  background: var(--status-warn);
  color: white;
  padding: var(--spacing-0-5) var(--spacing-1-5);
  border-radius: var(--radius-sm);
  font-weight: var(--font-semibold);
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

.icon-xs {
  width: 12px;
  height: 12px;
}

.text-bad {
  color: var(--status-bad);
}

/* Mobile Responsive */
@media (max-width: 1024px) {
  .profiles-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

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
  
  .profiles-panel {
    padding: var(--spacing-4);
  }
  
  .profiles-grid {
    grid-template-columns: 1fr;
  }
  
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 480px) {
  .tab span {
    display: none;
  }
  
  .profiles-grid {
    grid-template-columns: 1fr;
  }
  
  .profile-card {
    margin: 0 calc(-1 * var(--spacing-3));
    border-radius: 0;
    border-left: none;
    border-right: none;
  }
}
</style>
