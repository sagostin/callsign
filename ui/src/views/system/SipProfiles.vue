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
      <button class="btn-primary" @click="showModal = true">+ New Profile</button>
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
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ isEditing ? 'Edit SIP Profile' : 'New SIP Profile' }}</h3>
        <button class="btn-icon" @click="showModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
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
            <input v-model="profileSettings['sip-ip']" class="input-field" placeholder="$${local_ip_v4}" />
          </div>
          <div class="form-group">
            <label>SIP Port</label>
            <input v-model="profileSettings['sip-port']" class="input-field" placeholder="5060" />
          </div>
          <div class="form-group">
            <label>RTP IP</label>
            <input v-model="profileSettings['rtp-ip']" class="input-field" placeholder="$${local_ip_v4}" />
          </div>
        </div>

        <div class="form-group">
          <label>Context</label>
          <input v-model="profileSettings['context']" class="input-field" placeholder="public" />
        </div>

        <div class="form-group">
          <label>Codec Preferences</label>
          <input v-model="profileSettings['codec-prefs']" class="input-field" placeholder="OPUS,G722,PCMU,PCMA" />
        </div>

        <div class="form-row">
          <label class="checkbox-label">
            <input type="checkbox" v-model="profileSettings['auth-calls']" />
            <span>Authenticate Calls</span>
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="profileSettings['manage-presence']" />
            <span>Manage Presence</span>
          </label>
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
import { ref, onMounted, inject } from 'vue'
import { 
  RefreshCw as RefreshCwIcon, Server as ServerIcon, X as XIcon,
  Edit as EditIcon, StopCircle as StopCircleIcon, PlayCircle as PlayCircleIcon
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
const profileSettings = ref({
  'sip-ip': '$${local_ip_v4}',
  'sip-port': '5060',
  'rtp-ip': '$${local_ip_v4}',
  'context': 'public',
  'codec-prefs': 'OPUS,G722,PCMU,PCMA',
  'auth-calls': true,
  'manage-presence': true
})

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
  profileSettings.value = {}
  profile.settings?.forEach(s => {
    profileSettings.value[s.setting_name] = s.setting_value
  })
  isEditing.value = true
  showModal.value = true
}

async function saveProfile() {
  try {
    const payload = {
      ...activeProfile.value,
      settings: Object.entries(profileSettings.value).map(([name, value]) => ({
        setting_name: name,
        setting_value: String(value),
        enabled: true
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
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

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
