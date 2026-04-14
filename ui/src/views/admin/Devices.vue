<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Devices</h2>
      <p class="text-muted text-sm">Provisioned hard phones and softphone clients.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="$router.push('/admin/devices/templates')">
        <FileCodeIcon class="btn-icon-left" />
        Templates
      </button>
      <button class="btn-primary" @click="showAddModal = true">+ Add Device</button>
    </div>
  </div>

  <!-- Stats Cards -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-icon online"><CheckCircleIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ registeredCount }}</span>
        <span class="stat-label">Registered</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon offline"><XCircleIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ offlineCount }}</span>
        <span class="stat-label">Offline</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon ringing"><PhoneIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ inCallCount }}</span>
        <span class="stat-label">In Call</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon total"><MonitorIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ devices.length }}</span>
        <span class="stat-label">Total Devices</span>
      </div>
    </div>
  </div>

  <!-- Filter Bar -->
  <div class="filter-bar">
    <div class="search-box">
      <SearchIcon class="search-icon" />
      <input type="text" v-model="searchQuery" placeholder="Search by MAC, model, extension..." class="search-input">
    </div>
    <select v-model="filterStatus" class="filter-select">
      <option value="">All Statuses</option>
      <option value="Registered">Registered</option>
      <option value="Offline">Offline</option>
      <option value="In Call">In Call</option>
    </select>
    <select v-model="filterLocation" class="filter-select">
      <option value="">All Locations</option>
      <option v-for="loc in locations" :key="loc" :value="loc">{{ loc }}</option>
    </select>
    <select v-model="filterModel" class="filter-select">
      <option value="">All Models</option>
      <option v-for="m in models" :key="m" :value="m">{{ m }}</option>
    </select>
    <select v-model="filterProfile" class="filter-select">
      <option value="">All Profiles</option>
      <option v-for="p in deviceProfiles" :key="p.id" :value="p.id">{{ p.name }}</option>
    </select>
  </div>

  <!-- Devices Table -->
  <div class="table-container">
    <DataTable :columns="columns" :data="filteredDevices" actions>
      <template #mac="{ value, row }">
        <div class="mac-cell">
          <div class="device-indicator" :class="row.status.toLowerCase().replace(' ', '-')"></div>
          <span class="font-mono">{{ value }}</span>
        </div>
      </template>
      
      <template #model="{ value, row }">
        <div class="model-cell">
          <span class="model-name">{{ value }}</span>
          <span class="template-badge" v-if="row.template">{{ row.template }}</span>
        </div>
      </template>

      <template #ext="{ value, row }">
        <div class="ext-cell" v-if="row.userName">
          <span class="ext-number">{{ value || '—' }}</span>
          <span class="ext-name">{{ row.userName }}</span>
        </div>
        <span class="unassigned" v-else>Unassigned</span>
      </template>

      <template #profile="{ value }">
        <span class="profile-badge" v-if="value" :style="{ background: getProfileColor(value) }">
          {{ getProfileName(value) }}
        </span>
        <span class="text-muted text-xs" v-else>—</span>
      </template>

      <template #status="{ value }">
        <StatusBadge :status="value" />
      </template>
      
      <template #lastSeen="{ value }">
        <span class="text-muted text-xs">{{ value }}</span>
      </template>

      <template #actions="{ row }">
        <button class="btn-link" @click="$router.push(`/admin/devices/${row.mac.replace(/:/g, '')}`)">Edit</button>
        <button class="btn-link" @click="reprovision(row)">Reprovision</button>
        <div class="dropdown-container">
          <button class="btn-icon" @click.stop="toggleDropdown(row.mac)">
            <MoreVerticalIcon class="icon-sm" />
          </button>
          <div v-if="activeDropdown === row.mac" class="dropdown-menu">
            <button @click="rebootDevice(row)"><RefreshCwIcon class="menu-icon" /> Reboot</button>
            <button @click="viewLogs(row)"><FileTextIcon class="menu-icon" /> View Logs</button>
            <button class="text-bad" @click="deleteDevice(row)"><TrashIcon class="menu-icon" /> Delete</button>
          </div>
        </div>
      </template>
    </DataTable>
  </div>

  <!-- Add Device Modal -->
  <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>Add New Device</h3>
        <button class="btn-icon" @click="showAddModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="add-method-tabs">
          <button class="method-tab" :class="{ active: addMethod === 'manual' }" @click="addMethod = 'manual'">Manual Entry</button>
          <button class="method-tab" :class="{ active: addMethod === 'scan' }" @click="addMethod = 'scan'">Scan Network</button>
        </div>

        <div v-if="addMethod === 'manual'" class="add-form">
          <div class="form-group">
            <label>MAC Address</label>
            <input v-model="newDevice.mac" class="input-field code" placeholder="00:15:65:XX:XX:XX">
          </div>
          
          <div class="form-group">
            <label>Device Model</label>
            <select v-model="newDevice.model" class="input-field">
              <option value="">Select Model...</option>
              <option v-for="model in deviceModels" :key="model.id || model.name" :value="model.name">
                {{ model.name }} {{ model.vendor ? `(${model.vendor})` : '' }}
              </option>
            </select>
          </div>

          <!-- Generic SIP Configuration -->
          <div v-if="newDevice.model === 'Generic SIP'" class="generic-sip-config">
            <div class="config-header">
              <ServerIcon class="config-icon" />
              <span>Device Registration Settings</span>
            </div>
            <p class="help-text">Configure credentials and media settings for this generic device.</p>
            
            <div class="form-group">
              <label>Registration Username</label>
              <input v-model="newDevice.sipUsername" class="input-field code" placeholder="Auto-generated from MAC">
            </div>
            
            <div class="form-group">
              <label>Password</label>
              <input type="password" v-model="newDevice.sipPassword" class="input-field" placeholder="Enter device password">
            </div>

            <div class="form-group">
              <label>Supported Codecs</label>
              <div class="codec-checkboxes">
                <label class="codec-check" v-for="codec in codecOptions" :key="codec">
                  <input type="checkbox" :value="codec" v-model="newDevice.codecs">
                  <span>{{ codec }}</span>
                </label>
              </div>
            </div>

            <div class="form-row">
              <label class="toggle-row">
                <input type="checkbox" v-model="newDevice.earlyMedia">
                <span>Early Media (183)</span>
              </label>
              <label class="toggle-row">
                <input type="checkbox" v-model="newDevice.encryption">
                <span>Encryption (SRTP)</span>
              </label>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Transport</label>
                <select v-model="newDevice.sipTransport" class="input-field">
                  <option value="udp">UDP</option>
                  <option value="tcp">TCP</option>
                  <option value="tls">TLS</option>
                </select>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label>Template</label>
            <select v-model="newDevice.template_id" class="input-field">
              <option :value="null">Default for Model</option>
              <option v-for="t in addModalTemplates" :key="t.id" :value="t.id">{{ t.name }}</option>
            </select>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Assign to Extension</label>
              <select v-model="newDevice.extension_id" class="input-field">
                <option :value="null">Unassigned</option>
                <option v-for="ext in addModalExtensions" :key="ext.id" :value="ext.id">
                  {{ ext.extension }} - {{ ext.name }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>Device Profile</label>
              <select v-model="newDevice.profile_id" class="input-field">
                <option :value="null">None</option>
                <option v-for="p in deviceProfiles" :key="p.id" :value="p.id">{{ p.name }}</option>
              </select>
            </div>
          </div>
        </div>

        <div v-else class="scan-panel">
          <div class="scan-info">
            <WifiIcon class="scan-icon" />
            <p>Scan your network for unprovisioned SIP devices.</p>
          </div>
          <button class="btn-secondary full-width" @click="startScan">
            <SearchIcon class="btn-icon-left" /> Scan Network
          </button>
          <div class="scan-results" v-if="scanResults.length">
            <div class="scan-result" v-for="d in scanResults" :key="d.mac">
              <input type="checkbox" v-model="d.selected">
              <span class="font-mono">{{ d.mac }}</span>
              <span class="model-detect">{{ d.model }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showAddModal = false">Cancel</button>
        <button class="btn-primary" @click="addDevice" :disabled="!canAddDevice">Add Device</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, inject } from 'vue'
import { 
  Search as SearchIcon, 
  MoreVertical as MoreVerticalIcon, 
  FileCode as FileCodeIcon,
  CheckCircle as CheckCircleIcon,
  XCircle as XCircleIcon,
  Phone as PhoneIcon,
  Monitor as MonitorIcon,
  RefreshCw as RefreshCwIcon,
  FileText as FileTextIcon,
  Trash2 as TrashIcon,
  X as XIcon,
  Wifi as WifiIcon,
  Server as ServerIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { devicesAPI, deviceProfilesAPI, deviceTemplatesAPI, extensionsAPI, systemAPI } from '@/services/api'

const toast = inject('toast')
const isLoading = ref(false)

const columns = [
  { key: 'mac', label: 'MAC Address', width: '150px' },
  { key: 'model', label: 'Model / Template' },
  { key: 'ext', label: 'User / Extension', width: '150px' },
  { key: 'profile', label: 'Profile', width: '120px' },
  { key: 'location', label: 'Location', width: '100px' },
  { key: 'ip', label: 'IP Address', width: '120px' },
  { key: 'status', label: 'Status', width: '90px' },
  { key: 'lastSeen', label: 'Last Seen', width: '90px' }
]

const devices = ref([])
const deviceProfiles = ref([])
const deviceModels = ref([])
const searchQuery = ref('')
const filterStatus = ref('')
const filterLocation = ref('')
const filterModel = ref('')
const filterProfile = ref('')
const activeDropdown = ref(null)

onMounted(async () => {
  await Promise.all([fetchDevices(), fetchDeviceProfiles(), fetchAddModalData(), loadDeviceModels()])
  document.addEventListener('click', closeDropdown)
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown)
})

const closeDropdown = () => {
  activeDropdown.value = null
}

async function fetchDevices() {
  isLoading.value = true
  try {
    const response = await devicesAPI.list()
    const data = response.data || []
    devices.value = data.map(d => ({
      id: d.id,
      mac: formatMac(d.mac),
      model: d.model || 'Unknown',
      manufacturer: d.manufacturer || null,
      template: d.template_name || null,
      ext: d.lines?.[0]?.extension?.extension || null,
      userName: d.user?.first_name ? `${d.user.first_name} ${d.user.last_name || ''}`.trim() : null,
      userId: d.user_id || null,
      profileId: d.profile_id || null,
      profile: d.profile_id || null,
      location: d.location || null,
      ip: d.ip_address || '—',
      status: getDeviceStatus(d),
      lastSeen: formatLastSeen(d.last_seen)
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load devices')
    devices.value = []
  } finally {
    isLoading.value = false
  }
}

async function fetchDeviceProfiles() {
  try {
    const response = await deviceProfilesAPI.list()
    deviceProfiles.value = (response.data || []).map(p => ({
      id: p.id,
      name: p.name,
      color: p.color || '#6366f1'
    }))
  } catch (error) {
    console.error('Failed to load device profiles', error)
    deviceProfiles.value = []
  }
}

// Data for add modal dropdowns
const addModalTemplates = ref([])
const addModalExtensions = ref([])
const codecOptions = ['PCMU', 'PCMA', 'G722', 'G729', 'opus']

async function fetchAddModalData() {
  try {
    const [tmplRes, extRes] = await Promise.all([
      deviceTemplatesAPI.list(),
      extensionsAPI.list()
    ])
    addModalTemplates.value = (tmplRes.data || []).map(t => ({ id: t.id, name: t.name }))
    addModalExtensions.value = (extRes.data || []).map(e => ({
      id: e.id,
      extension: e.extension,
      name: e.user ? `${e.user.first_name} ${e.user.last_name || ''}`.trim() : (e.description || 'Unassigned')
    }))
  } catch (error) {
    console.error('Failed to load add-modal data', error)
  }
}

async function loadDeviceModels() {
  try {
    // Try device manufacturers endpoint which provides model listings
    const response = await systemAPI.listDeviceManufacturers()
    // Flatten manufacturers into model options
    const manufacturers = response.data || []
    const models = []
    for (const mfr of manufacturers) {
      if (mfr.models && Array.isArray(mfr.models)) {
        for (const model of mfr.models) {
          models.push({
            id: model.id || model.name,
            name: model.name,
            vendor: mfr.name
          })
        }
      }
    }
    deviceModels.value = models
  } catch (err) {
    // API for device models not available - this is an architectural gap
    console.error('Failed to load device models (no API available):', err)
    deviceModels.value = []
  }
}

const getProfileName = (id) => deviceProfiles.value.find(p => p.id === id)?.name || 'Unknown'
const getProfileColor = (id) => deviceProfiles.value.find(p => p.id === id)?.color || '#94a3b8'

function formatMac(mac) {
  if (!mac) return '—'
  return mac.toUpperCase().replace(/(.{2})/g, '$1:').slice(0, -1)
}

function getDeviceStatus(d) {
  if (d.in_call) return 'In Call'
  if (d.registered) return 'Registered'
  return 'Offline'
}

function formatLastSeen(dateStr) {
  if (!dateStr) return 'Never'
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  if (diffMins < 1) return 'Now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMs / 3600000)
  if (diffHours < 24) return `${diffHours}h ago`
  return `${Math.floor(diffMs / 86400000)}d ago`
}

const locations = computed(() => [...new Set(devices.value.map(d => d.location).filter(Boolean))])
const models = computed(() => [...new Set(devices.value.map(d => d.model))])

const registeredCount = computed(() => devices.value.filter(d => d.status === 'Registered').length)
const offlineCount = computed(() => devices.value.filter(d => d.status === 'Offline').length)
const inCallCount = computed(() => devices.value.filter(d => d.status === 'In Call').length)

const filteredDevices = computed(() => {
  return devices.value.filter(d => {
    const matchesSearch = d.mac.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          d.model.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          (d.ext && d.ext.includes(searchQuery.value)) ||
                          (d.userName && d.userName.toLowerCase().includes(searchQuery.value.toLowerCase()))
    const matchesStatus = !filterStatus.value || d.status === filterStatus.value
    const matchesLocation = !filterLocation.value || d.location === filterLocation.value
    const matchesModel = !filterModel.value || d.model === filterModel.value
    const matchesProfile = !filterProfile.value || d.profileId === parseInt(filterProfile.value)
    return matchesSearch && matchesStatus && matchesLocation && matchesModel && matchesProfile
  })
})

const toggleDropdown = (id) => {
  activeDropdown.value = activeDropdown.value === id ? null : id
}

const rebootDevice = async (row) => {
  activeDropdown.value = null
  if (!confirm(`Reboot device ${row.mac}?`)) return

  try {
    const macClean = row.mac.replace(/:/g, '')
    await devicesAPI.reboot(macClean)
    toast?.success(`Reboot command sent to ${row.mac}`)
  } catch (error) {
    toast?.error(error.message, 'Failed to reboot device')
  }
}

const viewLogs = (row) => {
  activeDropdown.value = null
  toast?.info(`Opening logs for ${row.mac}`)
}

const deleteDevice = async (row) => {
  activeDropdown.value = null
  if (confirm(`Delete device ${row.mac}?`)) {
    try {
      const macClean = row.mac.replace(/:/g, '')
      await devicesAPI.delete(macClean)
      toast?.success(`Device ${row.mac} deleted`)
      await fetchDevices()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete device')
    }
  }
}

const reprovision = async (row) => {
  try {
    const macClean = row.mac.replace(/:/g, '')
    await devicesAPI.reprovision(macClean)
    toast?.success(`Provisioning config sent to ${row.mac}`)
  } catch (error) {
    toast?.error(error.message, 'Failed to reprovision device')
  }
}

// Add Device Modal
const showAddModal = ref(false)
const addMethod = ref('manual')
const newDevice = ref({
  mac: '',
  model: '',
  template_id: null,
  extension_id: null,
  profile_id: null,
  sipUsername: '',
  sipPassword: '',
  sipTransport: 'udp',
  earlyMedia: false,
  encryption: false,
  codecs: ['PCMU', 'PCMA']
})

const canAddDevice = computed(() => {
  if (addMethod.value === 'manual') {
    return newDevice.value.mac && newDevice.value.model
  }
  return scanResults.value.some(r => r.selected)
})

const scanResults = ref([])

const startScan = () => {
  // Simulate network scan - in production this would call an API
  scanResults.value = [
    { mac: '00:15:65:AA:11:22', model: 'Yealink T54W', selected: false },
    { mac: '00:15:65:BB:33:44', model: 'Yealink T54W', selected: false },
  ]
  toast?.info('Network scan started...')
}

const addDevice = async () => {
  try {
    if (addMethod.value === 'manual') {
      const payload = {
        mac: newDevice.value.mac.replace(/:/g, '').toLowerCase(),
        model: newDevice.value.model,
        template_id: newDevice.value.template_id || undefined,
        profile_id: newDevice.value.profile_id || undefined,
        device_type: newDevice.value.model === 'Generic SIP' ? 'generic_sip' : 'provisioned',
        registration_pass: newDevice.value.sipPassword || undefined,
        registration_user: newDevice.value.sipUsername || undefined,
        sip_transport: newDevice.value.sipTransport,
        early_media: newDevice.value.earlyMedia,
        encryption_enabled: newDevice.value.encryption,
        supported_codecs: newDevice.value.codecs.join(','),
      }
      // If an extension is selected, we'll assign via lines after creation
      const res = await devicesAPI.create(payload)
      const createdDevice = res.data || res

      // Assign extension to line 1 if selected
      if (newDevice.value.extension_id && createdDevice.id) {
        await devicesAPI.updateLines(createdDevice.id, [{
          line_number: 1,
          line_type: 'line',
          extension_id: newDevice.value.extension_id,
          enabled: true
        }])
      }

      toast?.success('Device added successfully')
    } else {
      for (const r of scanResults.value.filter(r => r.selected)) {
        await devicesAPI.create({
          mac: r.mac.replace(/:/g, '').toLowerCase(),
          model: r.model,
        })
      }
      toast?.success(`${scanResults.value.filter(r => r.selected).length} devices added`)
    }
    await fetchDevices()
    showAddModal.value = false
    newDevice.value = { mac: '', model: '', template_id: null, extension_id: null, profile_id: null, sipUsername: '', sipPassword: '', sipTransport: 'udp', earlyMedia: false, encryption: false, codecs: ['PCMU', 'PCMA'] }
    scanResults.value = []
  } catch (error) {
    toast?.error(error.message, 'Failed to add device')
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

.header-actions {
  display: flex;
  gap: 8px;
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: var(--spacing-lg);
}

.stat-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.online { background: #dcfce7; color: #16a34a; }
.stat-icon.offline { background: #fee2e2; color: #dc2626; }
.stat-icon.ringing { background: #dbeafe; color: #2563eb; }
.stat-icon.total { background: #f3f4f6; color: #4b5563; }

.stat-info {
  display: flex;
  flex-direction: column;
}
.stat-value { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: var(--spacing-md);
}

.search-box {
  position: relative;
  flex: 1;
  max-width: 300px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--text-muted);
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  background: white;
}

/* Table Container */
.table-container {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

/* Cell Styles */
.mac-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.device-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.device-indicator.registered { background: #22c55e; }
.device-indicator.offline { background: #ef4444; }
.device-indicator.in-call { background: #3b82f6; animation: pulse 1.5s infinite; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.font-mono { font-family: monospace; font-size: 12px; }

.model-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.model-name { font-weight: 500; }
.template-badge {
  font-size: 10px;
  color: var(--text-muted);
  background: var(--bg-app);
  padding: 1px 6px;
  border-radius: 3px;
  display: inline-block;
  width: fit-content;
}

.ext-cell {
  display: flex;
  flex-direction: column;
}
.ext-number { font-weight: 600; font-family: monospace; }
.ext-name { font-size: 11px; color: var(--text-muted); }

.unassigned { color: var(--text-muted); font-style: italic; font-size: 12px; }

.profile-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 10px;
  font-weight: 600;
  color: white;
  white-space: nowrap;
}

/* Buttons */
.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
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
  display: flex;
  align-items: center;
  gap: 6px;
}
.btn-secondary:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-secondary.full-width { width: 100%; justify-content: center; }

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-right: 8px;
  cursor: pointer;
  font-weight: 500;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
}
.btn-icon:hover { color: var(--text-primary); }

.btn-icon-left { width: 14px; height: 14px; }
.icon-sm { width: 16px; height: 16px; }

/* Dropdown */
.dropdown-container {
  position: relative;
  display: inline-block;
}

.dropdown-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  background: white;
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-md);
  border-radius: var(--radius-sm);
  min-width: 140px;
  z-index: 20;
  overflow: hidden;
}

.dropdown-menu button {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  border: none;
  background: none;
  font-size: 12px;
  cursor: pointer;
  color: var(--text-main);
}
.dropdown-menu button:hover { background: var(--bg-app); }
.dropdown-menu button.text-bad { color: var(--status-bad); }

.menu-icon { width: 14px; height: 14px; }

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

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }

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

.add-method-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 20px;
  background: var(--bg-app);
  padding: 4px;
  border-radius: var(--radius-sm);
}

.method-tab {
  flex: 1;
  padding: 8px;
  border: none;
  background: transparent;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  color: var(--text-muted);
}
.method-tab.active {
  background: white;
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
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
.input-field.code { font-family: monospace; }
.input-field:focus { outline: none; border-color: var(--primary-color); }

.scan-panel {
  text-align: center;
}

.scan-info {
  padding: 24px;
  color: var(--text-muted);
}

.scan-icon {
  width: 48px;
  height: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.scan-results {
  margin-top: 16px;
  text-align: left;
}

.scan-result {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  margin-bottom: 8px;
}

.model-detect {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-muted);
}

.text-muted { color: var(--text-muted); }
.text-xs { font-size: 11px; }

/* Generic SIP Configuration */
.generic-sip-config {
  margin-top: 16px;
  padding: 16px;
  background: linear-gradient(to bottom right, #f0f9ff, #e0f2fe);
  border: 1px solid #bae6fd;
  border-radius: var(--radius-sm);
}
.generic-sip-config .config-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #0369a1;
  margin-bottom: 4px;
}
.generic-sip-config .config-icon {
  width: 18px;
  height: 18px;
}
.generic-sip-config .help-text {
  font-size: 12px;
  color: #0c4a6e;
  margin-bottom: 16px;
}
.generic-sip-config .form-group {
  margin-bottom: 12px;
}
.generic-sip-config .form-row {
  margin-bottom: 12px;
}

.codec-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.codec-check {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  text-transform: none;
  color: var(--text-main);
  cursor: pointer;
  background: white;
  transition: border-color 0.15s, background 0.15s;
}
.codec-check:has(input:checked) {
  border-color: var(--primary-color);
  background: rgba(99, 102, 241, 0.06);
  color: var(--primary-color);
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  text-transform: none;
  color: var(--text-main);
  cursor: pointer;
  padding: 4px 0;
}
</style>
