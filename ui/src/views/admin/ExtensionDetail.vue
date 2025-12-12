<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/admin/extensions')">← Back to Extensions</button>
      <h2>{{ isNew ? 'New Extension' : `${extension.firstName} ${extension.lastName}` }}</h2>
      <span v-if="!isNew" class="ext-badge">Ext {{ extension.ext }}</span>
    </div>
    <div class="header-actions" v-if="!isNew">
      <button class="btn-danger" @click="showDeleteModal = true">Delete Extension</button>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'general' }" @click="activeTab = 'general'">General</button>
    <button class="tab" :class="{ active: activeTab === 'call-handling' }" @click="activeTab = 'call-handling'" v-if="!isNew">Call Handling</button>
    <button class="tab" :class="{ active: activeTab === 'voicemail' }" @click="activeTab = 'voicemail'" v-if="!isNew">Voicemail</button>
    <button class="tab" :class="{ active: activeTab === 'devices' }" @click="activeTab = 'devices'" v-if="!isNew">Devices</button>
  </div>

  <div class="tab-content">
    
    <!-- GENERAL TAB -->
    <div v-if="activeTab === 'general'" class="form-layout">
      <div class="form-main">
        <div class="card">
          <h3>Basic Information</h3>
          <div class="form-grid">
            <div class="form-group">
              <label>Extension Number</label>
              <input type="text" v-model="extension.ext" :disabled="!isNew" class="input-field" placeholder="1000">
            </div>
            
            <div class="form-group">
              <label>Email Address</label>
              <input type="email" v-model="extension.email" class="input-field" placeholder="user@company.com">
            </div>
            
            <div class="form-group">
              <label>First Name</label>
              <input type="text" v-model="extension.firstName" class="input-field" placeholder="Jane">
            </div>
            
            <div class="form-group">
              <label>Last Name</label>
              <input type="text" v-model="extension.lastName" class="input-field" placeholder="Doe">
            </div>
          </div>
        </div>

        <div class="card">
          <h3>Security & Authentication</h3>
          <div class="form-grid">
            <div class="form-group">
              <label>Web Portal Password</label>
              <div class="input-group">
                <input type="password" v-model="extension.webPassword" class="input-field" placeholder="••••••••">
                <button class="btn-secondary small" @click="resetWebPassword">Reset</button>
              </div>
            </div>
            
            <div class="form-group">
              <label>Voicemail PIN</label>
              <input type="text" v-model="extension.vmPin" class="input-field code" maxlength="6" placeholder="1234">
            </div>
            
            <div class="form-group full-span">
              <label>SIP Password</label>
              <div class="input-group">
                <input type="text" v-model="extension.sipPassword" readonly class="input-field code">
                <button class="btn-secondary small" @click="regenerateSipPassword">Regenerate</button>
              </div>
              <span class="help-text">Used for IP phone and softphone registration.</span>
            </div>
          </div>
        </div>
      </div>

      <div class="form-sidebar">
        <div class="card profile-card">
          <h3>Extension Profile</h3>
          <select v-model="extension.profileId" class="input-field">
            <option :value="null">No Profile (Manual)</option>
            <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          
          <div class="profile-preview" v-if="currentProfile">
            <div class="profile-icon" :style="{ background: currentProfile.color }">
              {{ currentProfile.name.charAt(0) }}
            </div>
            <div class="profile-perms">
              <div class="perm-item" :class="{ enabled: currentProfile.permissions.outbound }">
                <CheckIcon v-if="currentProfile.permissions.outbound" class="icon-xs" />
                <XIcon v-else class="icon-xs" />
                Outbound Calls
              </div>
              <div class="perm-item highlight" :class="{ enabled: currentProfile.permissions.international }">
                <CheckIcon v-if="currentProfile.permissions.international" class="icon-xs" />
                <XIcon v-else class="icon-xs" />
                International
              </div>
              <div class="perm-item" :class="{ enabled: currentProfile.permissions.recording }">
                <CheckIcon v-if="currentProfile.permissions.recording" class="icon-xs" />
                <XIcon v-else class="icon-xs" />
                Recording
              </div>
              <div class="perm-item" :class="{ enabled: currentProfile.permissions.portal }">
                <CheckIcon v-if="currentProfile.permissions.portal" class="icon-xs" />
                <XIcon v-else class="icon-xs" />
                Portal Access
              </div>
            </div>
          </div>
          
          <div class="manual-perms" v-else>
            <p class="help-text">Configure permissions manually:</p>
            <label class="perm-toggle">
              <input type="checkbox" v-model="extension.permissions.outbound">
              <span>Outbound Calls</span>
            </label>
            <label class="perm-toggle highlight-intl">
              <input type="checkbox" v-model="extension.permissions.international">
              <span>International Calls</span>
            </label>
            <label class="perm-toggle">
              <input type="checkbox" v-model="extension.permissions.recording">
              <span>Call Recording</span>
            </label>
            <label class="perm-toggle">
              <input type="checkbox" v-model="extension.permissions.portal">
              <span>Portal Access</span>
            </label>
          </div>
        </div>

        <div class="card status-card" v-if="!isNew">
          <h3>Status</h3>
          <div class="status-row">
            <span>Registration</span>
            <StatusBadge :status="extension.status" />
          </div>
          <div class="status-row">
            <span>Last Activity</span>
            <span class="value">{{ extension.lastCall }}</span>
          </div>
          <div class="status-row">
            <span>Device</span>
            <span class="value font-mono">{{ extension.device || 'None' }}</span>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-secondary" @click="$router.push('/admin/extensions')">Cancel</button>
        <button class="btn-primary" @click="saveExtension">{{ isNew ? 'Create Extension' : 'Save Changes' }}</button>
      </div>
    </div>

    <!-- CALL HANDLING TAB -->
    <div v-else-if="activeTab === 'call-handling'" class="form-single">
      <div class="card">
        <h3>Ring Strategy</h3>
        <p class="help-text" style="margin-bottom: 16px;">Choose how incoming calls ring your devices.</p>
        
        <div class="ring-strategies">
          <label class="strategy-option" :class="{ active: callHandling.strategy === 'simultaneous' }">
            <input type="radio" v-model="callHandling.strategy" value="simultaneous">
            <div class="strategy-icon">
              <PhoneCallIcon class="icon" />
            </div>
            <div class="strategy-info">
              <strong>Ring All Simultaneously</strong>
              <span>All enabled devices ring at the same time</span>
            </div>
          </label>
          
          <label class="strategy-option" :class="{ active: callHandling.strategy === 'sequential' }">
            <input type="radio" v-model="callHandling.strategy" value="sequential">
            <div class="strategy-icon">
              <ListOrderedIcon class="icon" />
            </div>
            <div class="strategy-info">
              <strong>Ring in Order</strong>
              <span>Ring devices one at a time in the order below</span>
            </div>
          </label>
        </div>
      </div>

      <div class="card">
        <h3>Device Ring Order</h3>
        <p class="help-text" style="margin-bottom: 16px;">Drag to reorder. Toggle to enable/disable ringing for each device.</p>
        
        <div class="device-ring-list">
          <div 
            class="ring-device-item" 
            v-for="(device, index) in callHandling.devices" 
            :key="device.id"
            draggable="true"
            @dragstart="dragStart(index)"
            @dragover.prevent
            @drop="drop(index)"
            :class="{ disabled: !device.enabled, dragging: dragIndex === index }"
          >
            <div class="drag-handle">
              <GripVerticalIcon class="icon-sm" />
            </div>
            <div class="device-order">{{ index + 1 }}</div>
            <div class="device-icon-box" :class="device.type">
              <MonitorIcon v-if="device.type === 'softphone'" class="icon-sm" />
              <PhoneIcon v-else-if="device.type === 'desk'" class="icon-sm" />
              <SmartphoneIcon v-else-if="device.type === 'mobile'" class="icon-sm" />
              <HeadphonesIcon v-else class="icon-sm" />
            </div>
            <div class="device-info">
              <span class="device-name">{{ device.name }}</span>
              <span class="device-details">{{ device.details }}</span>
            </div>
            <div class="ring-duration" v-if="callHandling.strategy === 'sequential'">
              <select v-model="device.ringTime" class="input-field small">
                <option value="10">10 sec</option>
                <option value="15">15 sec</option>
                <option value="20">20 sec</option>
                <option value="30">30 sec</option>
              </select>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="device.enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>
      </div>

      <div class="card">
        <h3>No Answer Action</h3>
        <p class="help-text" style="margin-bottom: 16px;">What happens after all devices have been tried.</p>
        
        <div class="form-group">
          <select v-model="callHandling.noAnswerAction" class="input-field">
            <option value="voicemail">Send to Voicemail</option>
            <option value="forward">Forward to Number</option>
            <option value="hangup">Hang Up</option>
            <option value="queue">Send to Queue</option>
          </select>
        </div>
        
        <div class="form-group" v-if="callHandling.noAnswerAction === 'forward'">
          <label>Forward To</label>
          <input v-model="callHandling.forwardNumber" class="input-field" placeholder="(555) 555-1234">
        </div>
      </div>
      
      <div class="form-actions">
        <button class="btn-primary" @click="saveCallHandling">Save Call Handling</button>
      </div>
    </div>

    <!-- VOICEMAIL TAB -->
    <div v-else-if="activeTab === 'voicemail'" class="form-single">
      <div class="card">
        <div class="card-header-row">
          <h3>Voicemail Settings</h3>
          <label class="switch">
            <input type="checkbox" v-model="vm.enabled">
            <span class="slider round"></span>
          </label>
        </div>
        
        <div class="form-grid" v-if="vm.enabled">
          <div class="form-group">
            <label>Voicemail PIN</label>
            <input type="text" v-model="vm.pin" class="input-field code" maxlength="6">
          </div>
          
          <div class="form-group">
            <label>Email Notification</label>
            <input type="email" v-model="vm.email" class="input-field" placeholder="user@company.com">
          </div>
          
          <div class="form-group full-span">
            <label>Greetings</label>
            <div class="greeting-box">
              <div class="greeting-row">
                <span>Unavailable Message</span>
                <div class="greeting-actions">
                  <button class="btn-small">Play</button>
                  <button class="btn-small">Upload</button>
                  <button class="btn-small text-bad">Reset</button>
                </div>
              </div>
              <div class="greeting-row">
                <span>Busy Message</span>
                <div class="greeting-actions">
                  <button class="btn-small">Play</button>
                  <button class="btn-small">Upload</button>
                  <button class="btn-small text-bad">Reset</button>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="disabled-notice" v-else>
          <p>Voicemail is disabled for this extension.</p>
        </div>
      </div>
      
      <div class="form-actions">
        <button class="btn-primary" @click="saveVoicemail">Save Voicemail Settings</button>
      </div>
    </div>

    <!-- FORWARDING TAB -->
    <div v-else-if="activeTab === 'forwarding'" class="form-single">
      <div class="card">
        <h3>Call Forwarding</h3>
        
        <div class="forward-option">
          <div class="forward-header">
            <label class="switch">
              <input type="checkbox" v-model="forwarding.always.enabled">
              <span class="slider round"></span>
            </label>
            <div class="forward-info">
              <strong>Always Forward</strong>
              <span class="help-text">Immediately forward all calls</span>
            </div>
          </div>
          <input v-if="forwarding.always.enabled" type="text" v-model="forwarding.always.number" 
            class="input-field" placeholder="Enter destination number">
        </div>

        <div class="forward-option">
          <div class="forward-header">
            <label class="switch">
              <input type="checkbox" v-model="forwarding.busy.enabled">
              <span class="slider round"></span>
            </label>
            <div class="forward-info">
              <strong>Forward on Busy</strong>
              <span class="help-text">When extension is on a call</span>
            </div>
          </div>
          <input v-if="forwarding.busy.enabled" type="text" v-model="forwarding.busy.number" 
            class="input-field" placeholder="Enter destination number">
        </div>

        <div class="forward-option">
          <div class="forward-header">
            <label class="switch">
              <input type="checkbox" v-model="forwarding.noAnswer.enabled">
              <span class="slider round"></span>
            </label>
            <div class="forward-info">
              <strong>Forward on No Answer</strong>
              <span class="help-text">After ringing for specified time</span>
            </div>
          </div>
          <div v-if="forwarding.noAnswer.enabled" class="forward-fields">
            <input type="text" v-model="forwarding.noAnswer.number" class="input-field" placeholder="Destination">
            <select v-model="forwarding.noAnswer.timeout" class="input-field small">
              <option value="15">15 sec</option>
              <option value="20">20 sec</option>
              <option value="30">30 sec</option>
              <option value="45">45 sec</option>
            </select>
          </div>
        </div>
      </div>
      
      <div class="form-actions">
        <button class="btn-primary" @click="saveForwarding">Save Forwarding Rules</button>
      </div>
    </div>

    <!-- DEVICES TAB -->
    <div v-else-if="activeTab === 'devices'" class="form-single">
      <div class="panel-header">
        <h3>Registered Devices</h3>
        <button class="btn-secondary small">+ Assign Device</button>
      </div>
      
      <DataTable :columns="deviceColumns" :data="devices" actions>
        <template #status="{ value }">
          <StatusBadge :status="value" />
        </template>
        <template #actions>
          <button class="btn-link text-bad">Unassign</button>
        </template>
      </DataTable>
    </div>

    <!-- DELETE MODAL -->
    <div class="modal-overlay" v-if="showDeleteModal">
      <div class="modal">
        <h3>Delete Extension?</h3>
        <p>Are you sure you want to delete <strong>{{ extension.firstName }} {{ extension.lastName }}</strong> (Ext {{ extension.ext }})? This cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger" @click="confirmDelete">Delete Forever</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Check as CheckIcon, X as XIcon, Phone as PhoneIcon, PhoneCall as PhoneCallIcon, ListOrdered as ListOrderedIcon, GripVertical as GripVerticalIcon, Monitor as MonitorIcon, Smartphone as SmartphoneIcon, Headphones as HeadphonesIcon } from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { extensionsAPI, extensionProfilesAPI, usersAPI } from '@/services/api'

const route = useRoute()
const router = useRouter()
const toast = inject('toast')
const isLoading = ref(false)
const isNew = computed(() => route.params.id === 'new')
const activeTab = ref('general')
const showDeleteModal = ref(false)

// Profiles from API
const profiles = ref([])

// Extension form data
const extension = ref({
  ext: '',
  firstName: '',
  lastName: '',
  email: '',
  webPassword: '',
  vmPin: '',
  sipPassword: '',
  profileId: null,
  permissions: { outbound: true, international: false, recording: true, portal: true },
  status: 'Offline',
  device: null,
  lastCall: '—'
})

const currentProfile = computed(() => {
  return profiles.value.find(p => p.id === extension.value.profileId)
})

const vm = ref({
  enabled: true,
  pin: '1234',
  email: ''
})

const forwarding = ref({
  always: { enabled: false, number: '' },
  busy: { enabled: false, number: '' },
  noAnswer: { enabled: true, number: '', timeout: '20' }
})

const devices = ref([])

// Call Handling
const dragIndex = ref(null)
const callHandling = ref({
  strategy: 'simultaneous',
  noAnswerAction: 'voicemail',
  forwardNumber: '',
  devices: []
})

const dragStart = (index) => { dragIndex.value = index }
const drop = (index) => {
  const items = callHandling.value.devices
  const item = items.splice(dragIndex.value, 1)[0]
  items.splice(index, 0, item)
  dragIndex.value = null
}

const deviceColumns = [
  { key: 'mac', label: 'MAC Address' },
  { key: 'model', label: 'Model' },
  { key: 'ip', label: 'IP Address' },
  { key: 'status', label: 'Status' }
]

// Fetch data on mount
onMounted(async () => {
  await fetchProfiles()
  if (!isNew.value) {
    await fetchExtension()
  } else {
    // Generate random SIP password for new extension
    extension.value.sipPassword = generatePassword(16)
  }
})

async function fetchProfiles() {
  try {
    const response = await extensionProfilesAPI.list()
    profiles.value = (response.data.data || []).map(p => ({
      id: p.id,
      name: p.name,
      color: p.color || '#6366f1',
      permissions: p.permissions || {}
    }))
  } catch (error) {
    console.error('Failed to load profiles', error)
  }
}

async function fetchExtension() {
  isLoading.value = true
  try {
    const response = await extensionsAPI.get(route.params.id)
    const ext = response.data
    extension.value = {
      id: ext.id,
      ext: ext.extension,
      firstName: ext.display_name?.split(' ')[0] || '',
      lastName: ext.display_name?.split(' ').slice(1).join(' ') || '',
      email: ext.email || '',
      webPassword: '',
      vmPin: ext.voicemail_pin || '',
      sipPassword: ext.password || '',
      profileId: ext.profile_id,
      permissions: {
        outbound: ext.outbound_caller_id_name !== '',
        international: ext.toll_allow?.includes('international') || false,
        recording: ext.call_recording_enabled || false,
        portal: true
      },
      status: ext.registered ? 'Online' : 'Offline',
      device: ext.user_agent || null,
      lastCall: formatLastCall(ext.last_call_at)
    }
    vm.value.email = ext.email || ''
    vm.value.pin = ext.voicemail_pin || '1234'
    vm.value.enabled = ext.voicemail_enabled !== false
  } catch (error) {
    toast?.error('Failed to load extension', error.message)
    router.push('/admin/extensions')
  } finally {
    isLoading.value = false
  }
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

function generatePassword(length = 16) {
  const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  return Array.from({ length }, () => chars.charAt(Math.floor(Math.random() * chars.length))).join('')
}

async function saveExtension() {
  isLoading.value = true
  try {
    const payload = {
      extension: extension.value.ext,
      display_name: `${extension.value.firstName} ${extension.value.lastName}`.trim(),
      email: extension.value.email,
      password: extension.value.sipPassword,
      voicemail_pin: extension.value.vmPin,
      profile_id: extension.value.profileId,
      enabled: true
    }
    
    if (isNew.value) {
      await extensionsAPI.create(payload)
      toast?.success('Extension created successfully')
    } else {
      await extensionsAPI.update(extension.value.id, payload)
      toast?.success('Extension updated successfully')
    }
    router.push('/admin/extensions')
  } catch (error) {
    toast?.error(error.response?.data?.error || error.message, 'Failed to save extension')
  } finally {
    isLoading.value = false
  }
}

const saveVoicemail = async () => {
  try {
    await extensionsAPI.update(extension.value.id, {
      voicemail_enabled: vm.value.enabled,
      voicemail_pin: vm.value.pin
    })
    toast?.success('Voicemail settings saved')
  } catch (error) {
    toast?.error(error.message, 'Failed to save voicemail settings')
  }
}

const saveForwarding = () => toast?.info('Forwarding rules would be saved here')
const saveCallHandling = () => toast?.info('Call handling would be saved here')

const resetWebPassword = () => toast?.info('Password reset email would be sent')
const regenerateSipPassword = () => {
  extension.value.sipPassword = generatePassword(16)
  toast?.success('SIP password regenerated - remember to save changes')
}

const confirmDelete = async () => {
  try {
    await extensionsAPI.delete(extension.value.id)
    toast?.success('Extension deleted')
    showDeleteModal.value = false
    router.push('/admin/extensions')
  } catch (error) {
    toast?.error(error.message, 'Failed to delete extension')
  }
}
</script>

<style scoped>
.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.back-link {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  font-size: var(--text-xs);
  text-align: left;
}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.ext-badge {
  display: inline-block;
  background: var(--bg-app);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  font-family: monospace;
}

.tabs {
  display: flex;
  gap: 2px;
  margin-top: var(--spacing-lg);
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
  padding: var(--spacing-xl);
  border-radius: 0 0 var(--radius-md) var(--radius-md);
}

/* Form Layouts */
.form-layout {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: var(--spacing-lg);
  grid-template-rows: auto 1fr auto;
}

.form-main {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.form-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.form-single {
  max-width: 700px;
}

.form-actions {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--border-color);
  margin-top: var(--spacing-lg);
}

.card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
}

.card h3 {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: var(--spacing-md);
  color: var(--text-primary);
}

.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-md);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group.full-span {
  grid-column: span 2;
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
  font-size: var(--text-sm);
}
.input-field.code { font-family: monospace; background: var(--bg-app); }
.input-field.small { width: 100px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }

.input-group {
  display: flex;
  gap: 8px;
}
.input-group .input-field { flex: 1; }

.help-text {
  font-size: 11px;
  color: var(--text-muted);
}

/* Profile Card */
.profile-card select {
  margin-bottom: var(--spacing-md);
}

.profile-preview {
  display: flex;
  gap: 12px;
  padding: 12px;
  background: var(--bg-app);
  border-radius: var(--radius-sm);
}

.profile-icon {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 16px;
}

.profile-perms {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.perm-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
}
.perm-item.enabled { color: #16a34a; }
.perm-item.highlight { font-weight: 600; }

.manual-perms {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.perm-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  cursor: pointer;
  text-transform: none;
  font-weight: 400;
  color: var(--text-main);
}

.perm-toggle.highlight-intl {
  background: #fef3c7;
  padding: 6px 8px;
  margin: 0 -8px;
  border-radius: 4px;
}

/* Status Card */
.status-card {
  background: var(--bg-app);
}

.status-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
}
.status-row:last-child { border-bottom: none; }
.status-row .value { font-weight: 500; }

/* Forwarding */
.forward-option {
  padding: 16px 0;
  border-bottom: 1px solid var(--border-color);
}
.forward-option:last-child { border-bottom: none; }

.forward-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.forward-info {
  display: flex;
  flex-direction: column;
}
.forward-info strong { font-size: 14px; }

.forward-fields {
  display: flex;
  gap: 12px;
}

/* Greetings */
.greeting-box {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}
.greeting-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
  font-size: 13px;
}
.greeting-row:last-child { border-bottom: none; }
.greeting-actions { display: flex; gap: 6px; }

.disabled-notice {
  padding: 24px;
  text-align: center;
  color: var(--text-muted);
}

/* Panel Header */
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

/* Buttons */
.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 10px 24px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
}

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
.btn-secondary.small { padding: 6px 12px; font-size: 12px; }

.btn-danger {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  cursor: pointer;
}

.btn-small {
  font-size: 11px;
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  background: white;
  border-radius: 4px;
  cursor: pointer;
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  cursor: pointer;
  font-weight: 600;
}

.text-bad { color: var(--status-bad); }
.font-mono { font-family: monospace; }
.icon-xs { width: 12px; height: 12px; }

/* Switch */
.switch { position: relative; display: inline-block; width: 40px; height: 22px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--border-color); transition: .3s;
}
.slider:before {
  position: absolute; content: ""; height: 16px; width: 16px; left: 3px; bottom: 3px;
  background-color: white; transition: .3s;
}
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(18px); }
.slider.round { border-radius: 22px; }
.slider.round:before { border-radius: 50%; }

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.modal {
  background: white;
  padding: 24px;
  border-radius: var(--radius-md);
  width: 400px;
  box-shadow: var(--shadow-lg);
}
.modal h3 { font-size: 16px; font-weight: 700; margin-bottom: 8px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 20px; }

/* Call Handling Tab */
.ring-strategies { display: flex; gap: 16px; }
.strategy-option {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.15s;
}
.strategy-option:hover { border-color: var(--primary-color); }
.strategy-option.active { border-color: var(--primary-color); background: var(--primary-light); }
.strategy-option input { display: none; }
.strategy-icon { width: 40px; height: 40px; border-radius: 10px; background: var(--bg-app); display: flex; align-items: center; justify-content: center; }
.strategy-option.active .strategy-icon { background: var(--primary-color); color: white; }
.strategy-info { display: flex; flex-direction: column; gap: 2px; }
.strategy-info strong { font-size: 14px; }
.strategy-info span { font-size: 12px; color: var(--text-muted); }

.device-ring-list { display: flex; flex-direction: column; gap: 8px; }
.ring-device-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  transition: all 0.15s;
}
.ring-device-item:hover { background: var(--bg-app); }
.ring-device-item.disabled { opacity: 0.5; }
.ring-device-item.dragging { opacity: 0.5; background: var(--primary-light); }

.drag-handle { cursor: grab; color: var(--text-muted); }
.drag-handle:active { cursor: grabbing; }
.device-order { width: 24px; height: 24px; background: var(--bg-app); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; }
.device-icon-box { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.device-icon-box.softphone { background: #dbeafe; color: #2563eb; }
.device-icon-box.desk { background: #dcfce7; color: #16a34a; }
.device-icon-box.mobile { background: #f3e8ff; color: #7c3aed; }
.device-info { flex: 1; display: flex; flex-direction: column; }
.device-name { font-weight: 600; font-size: 14px; }
.device-details { font-size: 12px; color: var(--text-muted); }
.ring-duration select { width: 90px; }

.icon { width: 20px; height: 20px; }
.icon-sm { width: 16px; height: 16px; }
</style>
