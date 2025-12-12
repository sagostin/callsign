<template>
  <div class="view-header">
    <div class="header-content">
      <h2>SIP Profiles</h2>
      <p class="text-muted text-sm">Manage SIP endpoints, ports, and view live registrations.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="refreshAll" :disabled="refreshing">
        <RefreshCwIcon class="btn-icon" :class="{ spinning: refreshing }" />
        Refresh
      </button>
      <button class="btn-secondary" @click="reloadXML" :disabled="reloadingXML">
        <ServerIcon class="btn-icon" />
        Reload XML
      </button>
      <!-- Note: SIP profiles can only be edited, not created/deleted via UI.
           To add a new profile, place an XML file in sip_profiles/ directory. -->
    </div>
  </div>

  <!-- Live Status -->
  <div class="section-header">
    <h3>Live Status</h3>
    <span v-if="eslConnected" class="status-badge connected">ESL Connected</span>
    <span v-else class="status-badge disconnected">ESL Disconnected</span>
  </div>

  <div class="profiles-grid">
    <div 
      v-for="profile in profiles" 
      :key="profile.id" 
      class="profile-card"
      :class="{ selected: selectedProfile?.id === profile.id, running: profile.liveStatus === 'RUNNING' }"
      @click="selectProfile(profile)"
    >
      <div class="profile-header">
        <div class="profile-icon" :class="profile.liveStatus?.toLowerCase()">
          <ServerIcon class="icon" />
        </div>
        <div class="profile-info">
          <h4>{{ profile.profile_name }}</h4>
          <div class="profile-meta">
            <span class="ip-badge">{{ getProfileSetting(profile, 'sip-ip') || '0.0.0.0' }}</span>
            <span class="port-badge">:{{ getProfileSetting(profile, 'sip-port') || '5060' }}</span>
          </div>
        </div>
        <div class="profile-status">
          <span class="status-dot" :class="profile.liveStatus?.toLowerCase()"></span>
          <span class="status-text">{{ profile.liveStatus || 'Unknown' }}</span>
        </div>
      </div>

      <div class="profile-stats">
        <div class="stat">
          <span class="stat-value">{{ profile.registrations || 0 }}</span>
          <span class="stat-label">Registrations</span>
        </div>
        <div class="stat">
          <span class="stat-value">{{ profile.gateways || 0 }}</span>
          <span class="stat-label">Gateways</span>
        </div>
        <div class="stat">
          <span class="stat-value">{{ profile.calls || 0 }}</span>
          <span class="stat-label">Active Calls</span>
        </div>
      </div>

      <div class="profile-actions">
        <button class="btn-sm" @click.stop="restartProfile(profile)">
          <RefreshCwIcon class="btn-icon-sm" /> Restart
        </button>
        <button v-if="profile.liveStatus === 'RUNNING'" class="btn-sm btn-danger" @click.stop="stopProfile(profile)">
          <StopCircleIcon class="btn-icon-sm" /> Stop
        </button>
        <button v-else class="btn-sm btn-success" @click.stop="startProfile(profile)">
          <PlayCircleIcon class="btn-icon-sm" /> Start
        </button>
        <button class="btn-sm" @click.stop="editProfile(profile)">
          <EditIcon class="btn-icon-sm" /> Edit
        </button>
        <!-- Delete disabled - profiles are managed via backend files -->
      </div>
    </div>
  </div>

  <!-- Registrations Panel -->
  <div v-if="selectedProfile" class="registrations-panel">
    <div class="panel-header">
      <h3>{{ selectedProfile.profile_name }} - Registrations</h3>
      <button class="btn-link" @click="loadRegistrations(selectedProfile)">
        <RefreshCwIcon class="btn-icon-sm" /> Refresh
      </button>
    </div>
    
    <div v-if="loadingRegistrations" class="loading-state">Loading registrations...</div>
    <div v-else-if="registrations.length === 0" class="empty-state">
      <p>No active registrations for this profile.</p>
    </div>
    <pre v-else class="registrations-output">{{ registrations }}</pre>
  </div>

  <!-- Edit Modal -->
  <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>Edit SIP Profile</h3>
        <button class="btn-icon" @click="showModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <!-- Tabs -->
      <div class="modal-tabs">
        <button 
          v-for="tab in ['Basic', 'Settings', 'Domains']" 
          :key="tab" 
          class="tab-btn" 
          :class="{ active: activeTab === tab }"
          @click="activeTab = tab"
        >
          {{ tab }}
          <span v-if="tab === 'Settings'" class="tab-count">{{ allSettings.length }}</span>
        </button>
      </div>
      
      <div class="modal-body">
        <!-- Basic Tab -->
        <div v-if="activeTab === 'Basic'" class="tab-content">
          <div class="form-row">
            <div class="form-group">
              <label>Profile Name</label>
              <input v-model="activeProfile.profile_name" class="input-field" placeholder="e.g. internal-ipv4" />
            </div>
            <div class="form-group">
              <label>Description</label>
              <input v-model="activeProfile.description" class="input-field" placeholder="Internal SIP profile" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-group">
              <label>SIP IP</label>
              <input v-model="quickSettings['sip-ip']" class="input-field" placeholder="$${local_ip_v4}" />
            </div>
            <div class="form-group">
              <label>SIP Port</label>
              <input v-model="quickSettings['sip-port']" class="input-field" placeholder="5060" />
            </div>
            <div class="form-group">
              <label>RTP IP</label>
              <input v-model="quickSettings['rtp-ip']" class="input-field" placeholder="$${local_ip_v4}" />
            </div>
          </div>

          <div class="form-row three">
            <div class="form-group">
              <label>External SIP IP</label>
              <input v-model="quickSettings['ext-sip-ip']" class="input-field" placeholder="auto-nat" />
            </div>
            <div class="form-group">
              <label>External RTP IP</label>
              <input v-model="quickSettings['ext-rtp-ip']" class="input-field" placeholder="auto-nat" />
            </div>
            <div class="form-group">
              <label>Context</label>
              <input v-model="quickSettings['context']" class="input-field" placeholder="public" />
            </div>
          </div>

          <div class="form-group">
            <label>Inbound Codec Preferences</label>
            <input v-model="quickSettings['inbound-codec-prefs']" class="input-field" placeholder="OPUS,G722,PCMU,PCMA" />
          </div>
          <div class="form-group">
            <label>Outbound Codec Preferences</label>
            <input v-model="quickSettings['outbound-codec-prefs']" class="input-field" placeholder="OPUS,G722,PCMU,PCMA" />
          </div>
        </div>

        <!-- Settings Tab -->
        <div v-if="activeTab === 'Settings'" class="tab-content">
          <div class="settings-toolbar">
            <input 
              v-model="settingsSearch" 
              class="input-field search-input" 
              placeholder="Search settings..."
            />
            <button class="btn-sm" @click="addNewSetting">
              <PlusIcon class="btn-icon-sm" /> Add Setting
            </button>
          </div>

          <div class="settings-table-container">
            <table class="settings-table">
              <thead>
                <tr>
                  <th style="width: 30%">Name</th>
                  <th style="width: 35%">Value</th>
                  <th style="width: 15%">Enabled</th>
                  <th style="width: 15%">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(setting, idx) in filteredSettings" :key="idx" :class="{ disabled: !setting.enabled }">
                  <td>
                    <input 
                      v-model="setting.name" 
                      class="input-field setting-input"
                      :placeholder="setting.name || 'setting-name'"
                    />
                  </td>
                  <td>
                    <input 
                      v-model="setting.value" 
                      class="input-field setting-input"
                      :placeholder="setting.value || 'value'"
                    />
                  </td>
                  <td class="center">
                    <label class="toggle-switch">
                      <input type="checkbox" v-model="setting.enabled" />
                      <span class="toggle-slider"></span>
                    </label>
                  </td>
                  <td class="center">
                    <button class="btn-icon-only btn-delete-row" @click="removeSetting(idx)">
                      <TrashIcon class="icon-xs" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <p class="settings-hint">
            <strong>{{ allSettings.filter(s => s.enabled).length }}</strong> of <strong>{{ allSettings.length }}</strong> settings enabled. 
            Only enabled settings are sent to FreeSWITCH.
          </p>
        </div>

        <!-- Domains Tab -->
        <div v-if="activeTab === 'Domains'" class="tab-content">
          <div class="domains-toolbar">
            <button class="btn-sm" @click="addDomain">
              <PlusIcon class="btn-icon-sm" /> Add Domain
            </button>
          </div>

          <table class="settings-table" v-if="profileDomains.length > 0">
            <thead>
              <tr>
                <th>Domain Name</th>
                <th style="width: 15%">Alias</th>
                <th style="width: 15%">Parse</th>
                <th style="width: 15%">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(domain, idx) in profileDomains" :key="idx">
                <td>
                  <input v-model="domain.domain_name" class="input-field setting-input" placeholder="all" />
                </td>
                <td class="center">
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="domain.alias" />
                    <span class="toggle-slider"></span>
                  </label>
                </td>
                <td class="center">
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="domain.parse" />
                    <span class="toggle-slider"></span>
                  </label>
                </td>
                <td class="center">
                  <button class="btn-icon-only btn-delete-row" @click="removeDomain(idx)">
                    <TrashIcon class="icon-xs" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <div v-else class="empty-domains">
            <p>No domains configured. Click "Add Domain" to add one.</p>
            <p class="text-muted text-sm">Common options: "all" (matches all domains) or specific domain names.</p>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveProfile" :disabled="!activeProfile.profile_name">Save Profile</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { 
  RefreshCw as RefreshCwIcon, Server as ServerIcon, X as XIcon,
  Edit as EditIcon, StopCircle as StopCircleIcon, PlayCircle as PlayCircleIcon, 
  Trash2 as TrashIcon, Plus as PlusIcon
} from 'lucide-vue-next'
import { systemAPI } from '@/services/api'

const toast = inject('toast')
const profiles = ref([])
const selectedProfile = ref(null)
const registrations = ref('')
const eslConnected = ref(false)
const refreshing = ref(false)
const reloadingXML = ref(false)
const loadingRegistrations = ref(false)
const showModal = ref(false)
const isEditing = ref(false)
const activeProfile = ref({ profile_name: '', description: '', enabled: true })
const activeTab = ref('Basic')
const settingsSearch = ref('')
const profileDomains = ref([])

// Quick settings for Basic tab (most common ones)
const quickSettings = ref({
  'sip-ip': '$${local_ip_v4}',
  'sip-port': '5060',
  'rtp-ip': '$${local_ip_v4}',
  'ext-sip-ip': 'auto-nat',
  'ext-rtp-ip': 'auto-nat',
  'context': 'public',
  'inbound-codec-prefs': 'OPUS,G722,PCMU,PCMA',
  'outbound-codec-prefs': 'OPUS,G722,PCMU,PCMA',
})

// All settings array for Settings tab
const allSettings = ref([])

// Default SIP profile settings from FusionPBX
const defaultSettings = [
  { name: 'accept-blind-auth', value: 'true', enabled: false },
  { name: 'accept-blind-reg', value: 'true', enabled: false },
  { name: 'aggressive-nat-detection', value: 'true', enabled: false },
  { name: 'apply-inbound-acl', value: 'providers', enabled: true },
  { name: 'apply-nat-acl', value: 'nat.auto', enabled: true },
  { name: 'apply-register-acl', value: 'providers', enabled: false },
  { name: 'auth-all-packets', value: 'false', enabled: true },
  { name: 'auth-calls', value: 'true', enabled: true },
  { name: 'auth-subscriptions', value: 'true', enabled: true },
  { name: 'challenge-realm', value: 'auto_to', enabled: true },
  { name: 'context', value: 'public', enabled: true },
  { name: 'debug', value: '0', enabled: true },
  { name: 'dialplan', value: 'XML', enabled: true },
  { name: 'disable-naptr', value: 'false', enabled: false },
  { name: 'disable-register', value: 'true', enabled: false },
  { name: 'disable-rtp-auto-adjust', value: 'true', enabled: false },
  { name: 'disable-srv', value: 'false', enabled: false },
  { name: 'disable-transcoding', value: 'true', enabled: false },
  { name: 'dtmf-duration', value: '2000', enabled: true },
  { name: 'dtmf-type', value: 'rfc2833', enabled: true },
  { name: 'enable-timer', value: 'false', enabled: true },
  { name: 'ext-rtp-ip', value: 'auto-nat', enabled: true },
  { name: 'ext-sip-ip', value: 'auto-nat', enabled: true },
  { name: 'force-register-domain', value: '$${domain}', enabled: false },
  { name: 'force-register-db-domain', value: '$${domain}', enabled: false },
  { name: 'hold-music', value: 'local_stream://moh', enabled: true },
  { name: 'inbound-codec-prefs', value: 'OPUS,G722,PCMU,PCMA', enabled: true },
  { name: 'inbound-codec-negotiation', value: 'generous', enabled: true },
  { name: 'log-level', value: '0', enabled: true },
  { name: 'manage-presence', value: 'true', enabled: true },
  { name: 'manage-shared-appearance', value: 'true', enabled: true },
  { name: 'nonce-ttl', value: '60', enabled: true },
  { name: 'outbound-codec-prefs', value: 'OPUS,G722,PCMU,PCMA', enabled: true },
  { name: 'rfc2833-pt', value: '101', enabled: true },
  { name: 'rtp-hold-timeout-sec', value: '1800', enabled: true },
  { name: 'rtp-ip', value: '$${local_ip_v4}', enabled: true },
  { name: 'rtp-timeout-sec', value: '300', enabled: true },
  { name: 'rtp-timer-name', value: 'soft', enabled: true },
  { name: 'sip-ip', value: '$${local_ip_v4}', enabled: true },
  { name: 'sip-port', value: '5060', enabled: true },
  { name: 'sip-trace', value: 'false', enabled: false },
  { name: 'tls', value: 'true', enabled: false },
  { name: 'tls-bind-params', value: 'transport=tls', enabled: false },
  { name: 'tls-cert-dir', value: '/etc/freeswitch/tls', enabled: false },
  { name: 'tls-sip-port', value: '5061', enabled: false },
  { name: 'user-agent-string', value: 'Callsign', enabled: true },
]

// Computed: filtered settings based on search
const filteredSettings = computed(() => {
  if (!settingsSearch.value) return allSettings.value
  const q = settingsSearch.value.toLowerCase()
  return allSettings.value.filter(s => s.name.toLowerCase().includes(q) || s.value.toLowerCase().includes(q))
})

// Legacy profileSettings ref for compatibility
const profileSettings = ref({})

onMounted(async () => {
  await refreshAll()
})

async function refreshAll() {
  refreshing.value = true
  try {
    // Load profiles from DB
    const response = await systemAPI.listSIPProfiles()
    profiles.value = response.data?.data || response.data || []

    // Load live status from FreeSWITCH
    try {
      const statusResp = await systemAPI.getSofiaStatus()
      eslConnected.value = true
      parseSofiaStatus(statusResp.data?.data || '')
    } catch {
      eslConnected.value = false
    }
  } catch (error) {
    toast?.error('Failed to load profiles', error.message)
  } finally {
    refreshing.value = false
  }
}

function parseSofiaStatus(statusText) {
  // Parse FreeSWITCH sofia status output
  const lines = statusText.split('\n')
  profiles.value.forEach(p => {
    const line = lines.find(l => l.includes(p.profile_name))
    if (line) {
      p.liveStatus = line.includes('RUNNING') ? 'RUNNING' : 'STOPPED'
    }
  })
}

function getProfileSetting(profile, name) {
  const setting = profile.settings?.find(s => s.setting_name === name)
  return setting?.setting_value
}

async function selectProfile(profile) {
  selectedProfile.value = profile
  await loadRegistrations(profile)
}

async function loadRegistrations(profile) {
  loadingRegistrations.value = true
  try {
    const resp = await systemAPI.getSofiaProfileRegistrations(profile.profile_name)
    registrations.value = resp.data?.data || 'No registrations found'
  } catch (error) {
    registrations.value = 'Failed to load registrations: ' + error.message
  } finally {
    loadingRegistrations.value = false
  }
}

async function restartProfile(profile) {
  try {
    await systemAPI.restartSofiaProfile(profile.profile_name)
    toast?.success(`Profile ${profile.profile_name} restarted`)
    await refreshAll()
  } catch (error) {
    toast?.error('Failed to restart profile', error.message)
  }
}

async function startProfile(profile) {
  try {
    await systemAPI.startSofiaProfile(profile.profile_name)
    toast?.success(`Profile ${profile.profile_name} started`)
    await refreshAll()
  } catch (error) {
    toast?.error('Failed to start profile', error.message)
  }
}

async function stopProfile(profile) {
  if (!confirm(`Stop profile ${profile.profile_name}? Active calls may be dropped.`)) return
  try {
    await systemAPI.stopSofiaProfile(profile.profile_name)
    toast?.success(`Profile ${profile.profile_name} stopped`)
    await refreshAll()
  } catch (error) {
    toast?.error('Failed to stop profile', error.message)
  }
}

async function reloadXML() {
  reloadingXML.value = true
  try {
    await systemAPI.reloadSofiaXML()
    toast?.success('XML configuration reloaded')
  } catch (error) {
    toast?.error('Failed to reload XML', error.message)
  } finally {
    reloadingXML.value = false
  }
}

function editProfile(profile) {
  activeProfile.value = { ...profile }
  activeTab.value = 'Basic'
  settingsSearch.value = ''
  
  // Initialize allSettings from profile or defaults
  if (profile.settings?.length > 0) {
    allSettings.value = profile.settings.map(s => ({
      name: s.setting_name,
      value: s.setting_value,
      enabled: s.enabled !== false
    }))
  } else {
    allSettings.value = JSON.parse(JSON.stringify(defaultSettings))
  }
  
  // Initialize quickSettings from allSettings
  quickSettings.value = {}
  const quickKeys = ['sip-ip', 'sip-port', 'rtp-ip', 'ext-sip-ip', 'ext-rtp-ip', 'context', 'inbound-codec-prefs', 'outbound-codec-prefs']
  quickKeys.forEach(key => {
    const setting = allSettings.value.find(s => s.name === key)
    quickSettings.value[key] = setting?.value || ''
  })
  
  // Initialize domains
  profileDomains.value = profile.domains?.map(d => ({ ...d })) || []
  
  isEditing.value = true
  showModal.value = true
}

function openNewProfile() {
  activeProfile.value = { profile_name: '', description: '', enabled: true }
  activeTab.value = 'Basic'
  settingsSearch.value = ''
  allSettings.value = JSON.parse(JSON.stringify(defaultSettings))
  quickSettings.value = {
    'sip-ip': '$${local_ip_v4}',
    'sip-port': '5060',
    'rtp-ip': '$${local_ip_v4}',
    'ext-sip-ip': 'auto-nat',
    'ext-rtp-ip': 'auto-nat',
    'context': 'public',
    'inbound-codec-prefs': 'OPUS,G722,PCMU,PCMA',
    'outbound-codec-prefs': 'OPUS,G722,PCMU,PCMA',
  }
  profileDomains.value = [{ domain_name: 'all', alias: true, parse: true }]
  isEditing.value = false
  showModal.value = true
}

// Sync quick settings to allSettings before save
function syncQuickSettings() {
  for (const [name, value] of Object.entries(quickSettings.value)) {
    const existing = allSettings.value.find(s => s.name === name)
    if (existing) {
      existing.value = value
      existing.enabled = true
    } else if (value) {
      allSettings.value.push({ name, value, enabled: true })
    }
  }
}

async function saveProfile() {
  try {
    // Sync quick settings to allSettings
    syncQuickSettings()
    
    const payload = {
      ...activeProfile.value,
      settings: allSettings.value
        .filter(s => s.name.trim())
        .map(s => ({
          setting_name: s.name,
          setting_value: s.value,
          enabled: s.enabled
        })),
      domains: profileDomains.value
        .filter(d => d.domain_name.trim())
        .map(d => ({
          domain_name: d.domain_name,
          alias: d.alias,
          parse: d.parse
        }))
    }

    if (isEditing.value && activeProfile.value.id) {
      await systemAPI.updateSIPProfile(activeProfile.value.id, payload)
      toast?.success('Profile updated')
    } else {
      await systemAPI.createSIPProfile(payload)
      toast?.success('Profile created')
    }
    showModal.value = false
    await refreshAll()
  } catch (error) {
    toast?.error('Failed to save profile', error.message)
  }
}

// Settings CRUD
function addNewSetting() {
  allSettings.value.push({ name: '', value: '', enabled: true })
}

function removeSetting(idx) {
  allSettings.value.splice(idx, 1)
}

// Domains CRUD
function addDomain() {
  profileDomains.value.push({ domain_name: '', alias: true, parse: true })
}

function removeDomain(idx) {
  profileDomains.value.splice(idx, 1)
}

async function deleteProfile(profile) {
  // Check if profile has active registrations or is running
  if (profile.liveStatus === 'RUNNING') {
    toast?.error('Cannot delete', 'Profile is currently running. Stop it first.')
    return
  }
  
  if (profile.registrations > 0) {
    toast?.error('Cannot delete', 'Profile has active registrations.')
    return
  }
  
  if (!confirm(`Delete SIP profile "${profile.profile_name}"? This cannot be undone.`)) {
    return
  }
  
  try {
    await systemAPI.deleteSIPProfile(profile.id)
    toast?.success(`Profile ${profile.profile_name} deleted`)
    selectedProfile.value = null
    await refreshAll()
  } catch (error) {
    toast?.error('Failed to delete profile', error.message)
  }
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-actions { display: flex; gap: 8px; }
.section-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.section-header h3 { font-size: 14px; font-weight: 600; margin: 0; }

.status-badge { font-size: 11px; padding: 4px 10px; border-radius: 99px; font-weight: 600; }
.status-badge.connected { background: #dcfce7; color: #16a34a; }
.status-badge.disconnected { background: #fef2f2; color: #dc2626; }

.profiles-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; margin-bottom: 24px; }

.profile-card { background: white; border: 1px solid var(--border-color); border-radius: 12px; padding: 16px; cursor: pointer; transition: all 0.2s; }
.profile-card:hover { border-color: var(--primary-color); box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.profile-card.selected { border-color: var(--primary-color); box-shadow: 0 0 0 3px var(--primary-light); }
.profile-card.running { border-left: 3px solid #22c55e; }

.profile-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.profile-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; background: var(--bg-secondary); }
.profile-icon .icon { width: 20px; height: 20px; color: var(--text-muted); }
.profile-icon.running { background: #dcfce7; }
.profile-icon.running .icon { color: #16a34a; }
.profile-icon.stopped { background: #fef2f2; }
.profile-icon.stopped .icon { color: #dc2626; }

.profile-info { flex: 1; }
.profile-info h4 { font-size: 14px; font-weight: 600; margin: 0; }
.profile-meta { display: flex; gap: 4px; margin-top: 4px; }
.ip-badge, .port-badge { font-size: 10px; font-family: monospace; background: var(--bg-secondary); padding: 2px 6px; border-radius: 4px; }

.profile-status { display: flex; align-items: center; gap: 6px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-dot.running { background: #22c55e; }
.status-dot.stopped { background: #dc2626; }
.status-text { font-size: 11px; font-weight: 600; color: var(--text-muted); }

.profile-stats { display: flex; gap: 16px; padding: 12px 0; border-top: 1px solid var(--border-color); border-bottom: 1px solid var(--border-color); margin-bottom: 12px; }
.stat { display: flex; flex-direction: column; flex: 1; }
.stat-value { font-size: 18px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 10px; color: var(--text-muted); text-transform: uppercase; }

.profile-actions { display: flex; gap: 8px; }
.btn-sm { display: inline-flex; align-items: center; gap: 4px; padding: 6px 10px; font-size: 11px; font-weight: 500; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; }
.btn-sm:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-sm.btn-danger { border-color: #fecaca; color: #dc2626; }
.btn-sm.btn-success { border-color: #bbf7d0; color: #16a34a; }
.btn-sm.btn-delete { border-color: #fecaca; color: #dc2626; }
.btn-sm.btn-delete:hover { background: #fef2f2; }
.btn-icon-sm { width: 12px; height: 12px; }

.registrations-panel { background: white; border: 1px solid var(--border-color); border-radius: 12px; padding: 16px; }
.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.panel-header h3 { font-size: 14px; font-weight: 600; margin: 0; }
.registrations-output { background: #1e1e1e; color: #d4d4d4; padding: 16px; border-radius: 8px; font-size: 12px; font-family: monospace; overflow-x: auto; white-space: pre-wrap; max-height: 400px; }
.loading-state, .empty-state { padding: 40px; text-align: center; color: var(--text-muted); }

/* Buttons */
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: 6px; font-weight: 500; font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: 6px; font-weight: 500; font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; cursor: pointer; display: flex; align-items: center; gap: 4px; }
.btn-icon { width: 14px; height: 14px; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); padding: 24px; }
.modal-card { background: white; border-radius: 12px; box-shadow: 0 20px 40px rgba(0,0,0,0.2); width: 100%; max-width: 560px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-card.large { max-width: 900px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Modal Tabs */
.modal-tabs { display: flex; gap: 0; border-bottom: 1px solid var(--border-color); padding: 0 20px; }
.tab-btn { 
  padding: 12px 16px; 
  border: none; 
  background: none; 
  font-size: 13px; 
  font-weight: 600; 
  color: var(--text-muted); 
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.tab-btn:hover { color: var(--text-primary); }
.tab-btn.active { color: var(--primary-color); border-bottom-color: var(--primary-color); }
.tab-count { 
  background: var(--bg-secondary); 
  padding: 2px 6px; 
  border-radius: 10px; 
  font-size: 10px; 
  font-weight: 700; 
}
.tab-content { min-height: 300px; }

/* Settings Table */
.settings-toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-input { flex: 1; }
.settings-table-container { max-height: 400px; overflow-y: auto; border: 1px solid var(--border-color); border-radius: 8px; }
.settings-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.settings-table th { 
  text-align: left; 
  padding: 10px 12px; 
  background: var(--bg-secondary); 
  font-weight: 600; 
  font-size: 11px; 
  text-transform: uppercase; 
  color: var(--text-muted);
  position: sticky;
  top: 0;
}
.settings-table td { padding: 8px 12px; border-top: 1px solid var(--border-color); }
.settings-table tr.disabled { opacity: 0.5; }
.settings-table tr:hover { background: var(--bg-secondary); }
.setting-input { width: 100%; padding: 6px 10px; font-size: 12px; font-family: monospace; }
.center { text-align: center; }

/* Toggle Switch */
.toggle-switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background: #e5e7eb;
  border-radius: 20px;
  transition: 0.2s;
}
.toggle-slider::before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: 0.2s;
  box-shadow: 0 1px 2px rgba(0,0,0,0.2);
}
.toggle-switch input:checked + .toggle-slider { background: var(--primary-color); }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(16px); }

/* Delete button in table */
.btn-icon-only { 
  display: flex; 
  align-items: center; 
  justify-content: center; 
  width: 28px; 
  height: 28px; 
  border: none; 
  background: none; 
  cursor: pointer; 
  border-radius: 6px;
  color: var(--text-muted);
}
.btn-icon-only:hover { background: #fef2f2; color: #dc2626; }
.icon-xs { width: 14px; height: 14px; }

.settings-hint { font-size: 12px; color: var(--text-muted); margin-top: 12px; padding: 8px 12px; background: var(--bg-secondary); border-radius: 6px; }

/* Domains */
.domains-toolbar { margin-bottom: 16px; }
.empty-domains { padding: 40px; text-align: center; color: var(--text-muted); background: var(--bg-secondary); border-radius: 8px; }
.empty-domains p { margin: 0 0 8px; }

.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.form-row.three { grid-template-columns: 1fr 1fr 1fr; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 14px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.checkbox-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 400; text-transform: none; cursor: pointer; }
.checkbox-label input { width: 16px; height: 16px; accent-color: var(--primary-color); }
.icon-sm { width: 16px; height: 16px; }
</style>
