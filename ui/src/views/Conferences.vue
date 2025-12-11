<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Conferences</h2>
      <p class="text-muted text-sm">Manage audio conference rooms, profiles, and live sessions.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="$router.push('/admin/conferences/console/live')">
        <MonitorIcon class="icon-sm" /> Live Console
      </button>
      <button class="btn-primary" @click="$router.push('/admin/conferences/new')">
        <PlusIcon class="icon-sm" /> New Conference
      </button>
    </div>
  </div>

  <!-- Stats Cards -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-icon active">
        <UsersIcon class="icon-md" />
      </div>
      <div class="stat-info">
        <span class="stat-value">{{ stats.activeRooms }}</span>
        <span class="stat-label">Active Rooms</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon participants">
        <PhoneIcon class="icon-md" />
      </div>
      <div class="stat-info">
        <span class="stat-value">{{ stats.totalParticipants }}</span>
        <span class="stat-label">Total Participants</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon recording">
        <CircleIcon class="icon-md" />
      </div>
      <div class="stat-info">
        <span class="stat-value">{{ stats.recordingActive }}</span>
        <span class="stat-label">Recording</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon idle">
        <ClockIcon class="icon-md" />
      </div>
      <div class="stat-info">
        <span class="stat-value">{{ stats.idleRooms }}</span>
        <span class="stat-label">Idle Rooms</span>
      </div>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'rooms' }" @click="activeTab = 'rooms'">All Conferences</button>
    <button class="tab" :class="{ active: activeTab === 'profiles' }" @click="activeTab = 'profiles'">Conference Profiles</button>
  </div>

  <!-- CONFERENCES LIST -->
  <div class="tab-content" v-if="activeTab === 'rooms'">
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input type="text" v-model="searchQuery" placeholder="Search by room name..." class="search-input">
      </div>
      <select v-model="filterStatus" class="filter-select">
        <option value="">All Statuses</option>
        <option value="Active">Active</option>
        <option value="Idle">Idle</option>
      </select>
      <select v-model="filterProfile" class="filter-select">
        <option value="">All Profiles</option>
        <option v-for="p in profiles" :key="p.id" :value="p.name">{{ p.name }}</option>
      </select>
    </div>

    <DataTable :columns="columns" :data="filteredConferences" actions>
      <template #name="{ value, row }">
        <div class="room-name">
          <span class="room-title">{{ value }}</span>
          <span v-if="row.status === 'Active'" class="live-indicator">
            <span class="pulse-dot"></span> Live
          </span>
        </div>
      </template>
      <template #participants="{ value, row }">
        <div class="participant-badge" :class="{ active: value > 0 }">
          <UsersIcon class="icon-xs" />
          <span>{{ value }}</span>
        </div>
      </template>
      <template #status="{ value }">
        <StatusBadge :status="value" />
      </template>
      <template #profile="{ value }">
        <span class="profile-tag">{{ value }}</span>
      </template>
      <template #actions="{ row }">
        <div class="action-buttons">
          <button class="btn-icon" :class="{ 'text-primary': row.status === 'Active' }" 
                  title="Live Console" @click="openConsole(row)">
            <MonitorIcon class="icon-sm" />
          </button>
          <button class="btn-icon" title="Edit" @click="$router.push(`/admin/conferences/${row.id}`)">
            <EditIcon class="icon-sm" />
          </button>
          <button class="btn-icon text-bad" title="Delete" @click="deleteConference(row)">
            <TrashIcon class="icon-sm" />
          </button>
        </div>
      </template>
    </DataTable>
  </div>

  <!-- CONFERENCE PROFILES -->
  <div class="tab-content profiles-panel" v-else-if="activeTab === 'profiles'">
    <div class="panel-header">
      <div>
        <h3>Conference Profiles</h3>
        <p class="help-text">Define reusable settings for conference rooms including codecs, limits, and defaults.</p>
      </div>
      <button class="btn-primary" @click="showProfileModal = true">
        <PlusIcon class="icon-sm" /> New Profile
      </button>
    </div>

    <div class="profiles-grid">
      <div class="profile-card" v-for="profile in profiles" :key="profile.id">
        <div class="profile-header">
          <div class="profile-icon" :style="{ background: profile.color }">
            {{ profile.name.charAt(0) }}
          </div>
          <div class="profile-info">
            <h4>{{ profile.name }}</h4>
            <span class="profile-usage">{{ profile.roomCount }} rooms</span>
          </div>
          <div class="profile-actions">
            <button class="btn-icon" @click="editProfile(profile)"><EditIcon class="icon-sm" /></button>
            <button class="btn-icon" @click="deleteProfile(profile)"><TrashIcon class="icon-sm text-bad" /></button>
          </div>
        </div>
        
        <div class="profile-settings">
          <div class="setting-row">
            <span class="setting-label">Audio Codec</span>
            <span class="setting-value">{{ profile.codec }}</span>
          </div>
          <div class="setting-row">
            <span class="setting-label">Max Participants</span>
            <span class="setting-value">{{ profile.maxParticipants }}</span>
          </div>
          <div class="setting-row">
            <span class="setting-label">Auto-Record</span>
            <span class="setting-value" :class="profile.autoRecord ? 'enabled' : 'disabled'">
              {{ profile.autoRecord ? 'Yes' : 'No' }}
            </span>
          </div>
          <div class="setting-row">
            <span class="setting-label">Entry Tone</span>
            <span class="setting-value">{{ profile.entryTone ? 'Enabled' : 'Disabled' }}</span>
          </div>
        </div>

        <div class="profile-features" v-if="profile.features?.length">
          <span class="feature-tag" v-for="f in profile.features" :key="f">{{ f }}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Live Console Modal -->
  <div v-if="showConsole" class="modal-overlay" @click.self="showConsole = false">
    <div class="modal-card large">
      <div class="modal-header console-header">
        <div>
          <h3>{{ activeConference?.name }} <span class="text-muted">({{ activeConference?.ext }})</span></h3>
          <div class="live-status">
            <span class="pulse-dot"></span>
            <span>Live - {{ activeParticipants.length }} Participants</span>
          </div>
        </div>
        <div class="header-controls">
          <button class="btn-icon" title="Pop Out Window" @click="popOutConsole">
            <ExternalLinkIcon class="icon-sm" />
          </button>
          <button class="btn-icon" @click="showConsole = false">
            <XIcon class="icon-sm" />
          </button>
        </div>
      </div>
      
      <div class="console-toolbar">
        <button class="btn-tool danger" @click="muteAll">
          <MicOffIcon class="icon-sm" /> Mute All
        </button>
        <button class="btn-tool" @click="toggleLock">
          <LockIcon class="icon-sm" /> {{ activeConference?.locked ? 'Unlock Room' : 'Lock Room' }}
        </button>
        <button class="btn-tool">
          <CircleIcon class="icon-sm" /> {{ isRecording ? 'Stop Recording' : 'Start Recording' }}
        </button>
      </div>

      <div class="console-table">
        <table>
          <thead>
            <tr>
              <th>Participant</th>
              <th>Caller ID</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in activeParticipants" :key="p.id">
              <td>
                <div class="participant-info">
                  <span class="name">{{ p.name }}</span>
                  <span class="role">{{ p.role }}</span>
                </div>
              </td>
              <td class="mono">{{ p.number }}</td>
              <td>
                <span v-if="p.talking" class="talking-badge">
                  <MicIcon class="icon-xs" /> Talking
                </span>
                <span v-else class="silent-badge">Silent</span>
              </td>
              <td>
                <div class="action-buttons">
                  <button class="btn-icon" :class="{ 'text-bad': p.muted }" 
                          title="Mute" @click="toggleMute(p)">
                    <MicOffIcon class="icon-sm" />
                  </button>
                  <button class="btn-icon" title="Kick" @click="kickParticipant(p)">
                    <UserMinusIcon class="icon-sm" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- Profile Modal -->
  <div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ isEditingProfile ? 'Edit Profile' : 'New Conference Profile' }}</h3>
        <button class="btn-icon" @click="showProfileModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Profile Name</label>
          <input v-model="profileForm.name" class="input-field" placeholder="e.g. Wideband HD">
        </div>

        <div class="form-group">
          <label>Color</label>
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

        <div class="form-group">
          <label>Audio Codec</label>
          <select v-model="profileForm.codec" class="input-field">
            <option value="G.711">G.711 (Narrowband)</option>
            <option value="G.722">G.722 (Wideband HD)</option>
            <option value="Opus">Opus (Adaptive)</option>
          </select>
        </div>

        <div class="form-group">
          <label>Max Participants</label>
          <input type="number" v-model="profileForm.maxParticipants" class="input-field" min="2" max="500">
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Default Settings</h4>
          <label class="toggle-row">
            <input type="checkbox" v-model="profileForm.autoRecord">
            <span>Auto-Record Conferences</span>
          </label>
          <label class="toggle-row">
            <input type="checkbox" v-model="profileForm.entryTone">
            <span>Play Entry/Exit Tones</span>
          </label>
          <label class="toggle-row">
            <input type="checkbox" v-model="profileForm.announceCount">
            <span>Announce Participant Count</span>
          </label>
          <label class="toggle-row">
            <input type="checkbox" v-model="profileForm.waitForModerator">
            <span>Wait for Moderator</span>
          </label>
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
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import DataTable from '../components/common/DataTable.vue'
import StatusBadge from '../components/common/StatusBadge.vue'
import { 
  Search as SearchIcon,
  Plus as PlusIcon,
  Monitor as MonitorIcon,
  Users as UsersIcon,
  Phone as PhoneIcon,
  Circle as CircleIcon,
  Clock as ClockIcon,
  Edit as EditIcon,
  Trash2 as TrashIcon,
  X as XIcon,
  ExternalLink as ExternalLinkIcon,
  MicOff as MicOffIcon,
  Lock as LockIcon,
  Mic as MicIcon,
  UserMinus as UserMinusIcon
} from 'lucide-vue-next'
import { conferencesAPI } from '@/services/api'

const toast = inject('toast')
const isLoading = ref(false)

const activeTab = ref('rooms')
const searchQuery = ref('')
const filterStatus = ref('')
const filterProfile = ref('')
const showConsole = ref(false)
const activeConference = ref(null)
const isRecording = ref(true)
const showProfileModal = ref(false)
const isEditingProfile = ref(false)

const conferences = ref([])
const profiles = ref([])

onMounted(async () => {
  await fetchConferences()
})

async function fetchConferences() {
  isLoading.value = true
  try {
    const response = await conferencesAPI.list()
    conferences.value = (response.data || []).map(c => ({
      id: c.id,
      name: c.name,
      ext: c.extension,
      profile: c.profile_name || 'Default',
      participants: c.participant_count || 0,
      status: c.participant_count > 0 ? 'Active' : 'Idle'
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load conferences')
    // Fallback to demo data
    conferences.value = [
      { id: 1, name: 'Weekly Sales', ext: '3001', profile: 'Default', participants: 0, status: 'Idle' },
      { id: 2, name: 'Project Alpha', ext: '3005', profile: 'Wideband', participants: 4, status: 'Active' },
    ]
    profiles.value = [
      { id: 1, name: 'Default', color: '#6366f1', codec: 'G.711', maxParticipants: 50, autoRecord: false, entryTone: true, roomCount: 2 },
      { id: 2, name: 'Wideband', color: '#22c55e', codec: 'G.722', maxParticipants: 100, autoRecord: true, entryTone: true, roomCount: 1 },
    ]
  } finally {
    isLoading.value = false
  }
}

// Stats
const stats = computed(() => ({
  activeRooms: conferences.value.filter(c => c.status === 'Active').length,
  totalParticipants: conferences.value.reduce((sum, c) => sum + c.participants, 0),
  recordingActive: conferences.value.filter(c => c.status === 'Active').length,
  idleRooms: conferences.value.filter(c => c.status === 'Idle').length
}))

const columns = [
  { key: 'name', label: 'Room Name' },
  { key: 'ext', label: 'Extension', width: '100px' },
  { key: 'profile', label: 'Profile', width: '120px' },
  { key: 'participants', label: 'Participants', width: '120px' },
  { key: 'status', label: 'Status', width: '100px' }
]

const filteredConferences = computed(() => {
  return conferences.value.filter(c => {
    const matchesSearch = c.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesStatus = !filterStatus.value || c.status === filterStatus.value
    const matchesProfile = !filterProfile.value || c.profile === filterProfile.value
    return matchesSearch && matchesStatus && matchesProfile
  })
})

const colorOptions = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#64748b']

const profileForm = ref({
  name: '',
  color: '#6366f1',
  codec: 'G.711',
  maxParticipants: 50,
  autoRecord: false,
  entryTone: true,
  announceCount: false,
  waitForModerator: false
})

const editProfile = (profile) => {
  profileForm.value = { ...profile }
  isEditingProfile.value = true
  showProfileModal.value = true
}

const deleteProfile = async (profile) => {
  if (confirm(`Delete profile "${profile.name}"?`)) {
    try {
      await conferencesAPI.deleteProfile(profile.id)
      toast?.success(`Profile "${profile.name}" deleted`)
      profiles.value = profiles.value.filter(p => p.id !== profile.id)
    } catch (error) {
      toast?.error(error.message, 'Failed to delete profile')
    }
  }
}

const saveProfile = async () => {
  try {
    if (isEditingProfile.value) {
      toast?.success('Profile updated')
    } else {
      profiles.value.push({
        ...profileForm.value,
        id: Date.now(),
        roomCount: 0,
        features: []
      })
      toast?.success('Profile created')
    }
    showProfileModal.value = false
    isEditingProfile.value = false
  } catch (error) {
    toast?.error(error.message, 'Failed to save profile')
  }
}

// Console
const activeParticipants = ref([
  { id: 1, name: 'Alice Smith', role: 'Moderator', number: '101', talking: true, muted: false },
  { id: 2, name: 'Bob Jones', role: 'Member', number: '102', talking: false, muted: false },
])

const openConsole = (conf) => {
  activeConference.value = conf
  showConsole.value = true
}

const popOutConsole = () => {
  window.open('/admin/conferences/console/live', 'ConferenceConsole', 'width=800,height=600')
}

const muteAll = () => {
  activeParticipants.value.forEach(p => p.muted = true)
  toast?.info('All participants muted')
}

const toggleLock = () => {
  if (activeConference.value) {
    activeConference.value.locked = !activeConference.value.locked
    toast?.info(activeConference.value.locked ? 'Room locked' : 'Room unlocked')
  }
}

const toggleMute = (p) => {
  p.muted = !p.muted
}

const kickParticipant = (p) => {
  if (confirm(`Kick ${p.name}?`)) {
    activeParticipants.value = activeParticipants.value.filter(part => part.id !== p.id)
    toast?.success(`${p.name} kicked from conference`)
  }
}

const deleteConference = async (row) => {
  if (confirm(`Delete conference "${row.name}"?`)) {
    try {
      await conferencesAPI.delete(row.id)
      toast?.success(`Conference "${row.name}" deleted`)
      await fetchConferences()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete conference')
    }
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
  gap: 10px;
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
  display: flex;
  align-items: center;
  gap: 6px;
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
  display: flex;
  align-items: center;
  gap: 6px;
}
.btn-icon {
  background: none;
  border: none;
  padding: 6px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--text-muted);
  transition: all var(--transition-fast);
}
.btn-icon:hover { background: var(--bg-app); color: var(--text-main); }
.btn-icon.text-primary { color: var(--primary-color); }
.btn-icon.text-bad { color: var(--status-bad); }
.btn-icon.text-bad:hover { background: #fee2e2; }

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
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
  gap: 14px;
  box-shadow: var(--shadow-sm);
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon.active { background: #dcfce7; color: #16a34a; }
.stat-icon.participants { background: #e0e7ff; color: #4f46e5; }
.stat-icon.recording { background: #fee2e2; color: #dc2626; }
.stat-icon.idle { background: #f1f5f9; color: #64748b; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-main); }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Tabs */
.tabs {
  display: flex;
  gap: 2px;
  border-bottom: 1px solid var(--border-color);
}
.tab {
  padding: 10px 20px;
  background: transparent;
  border: 1px solid transparent;
  border-bottom: none;
  cursor: pointer;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-muted);
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  transition: all var(--transition-fast);
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
  box-shadow: var(--shadow-sm);
}

/* Filter Bar */
.filter-bar {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-app);
  flex-wrap: wrap;
}
.search-box {
  position: relative;
  flex: 1;
  min-width: 200px;
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
  min-width: 130px;
}

/* Room Name */
.room-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
.room-title { font-weight: 500; }
.live-indicator {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-weight: 600;
  color: #16a34a;
  background: #dcfce7;
  padding: 2px 8px;
  border-radius: 99px;
}
.pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #16a34a;
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* Participant Badge */
.participant-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 500;
  background: #f1f5f9;
  color: #64748b;
}
.participant-badge.active {
  background: #dcfce7;
  color: #16a34a;
}

.profile-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  background: #e0e7ff;
  color: #4f46e5;
}

.action-buttons {
  display: flex;
  gap: 4px;
}

/* Profiles Panel */
.profiles-panel {
  padding: var(--spacing-xl);
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
.help-text {
  font-size: 13px;
  color: var(--text-muted);
}

.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
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
.profile-info { flex: 1; }
.profile-info h4 { font-size: 14px; font-weight: 600; margin: 0; }
.profile-usage { font-size: 11px; color: var(--text-muted); }
.profile-actions { display: flex; gap: 4px; }

.profile-settings {
  padding: 12px 16px;
}
.setting-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}
.setting-label { color: var(--text-muted); }
.setting-value { font-weight: 500; }
.setting-value.enabled { color: #16a34a; }
.setting-value.disabled { color: #dc2626; }

.profile-features {
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.feature-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  background: #f1f5f9;
  color: #64748b;
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
  max-width: 480px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}
.modal-card.large {
  max-width: 700px;
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

/* Console Header */
.console-header {
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  color: white;
  border: none;
}
.console-header h3 { color: white; }
.console-header .text-muted { color: rgba(255,255,255,0.7); }
.live-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  margin-top: 4px;
  color: #86efac;
}
.header-controls { display: flex; gap: 8px; }
.header-controls .btn-icon { color: white; }
.header-controls .btn-icon:hover { background: rgba(255,255,255,0.2); }

.console-toolbar {
  display: flex;
  gap: 8px;
  padding: 12px 20px;
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.btn-tool {
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: white;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}
.btn-tool.danger {
  background: #fee2e2;
  border-color: #fecaca;
  color: #dc2626;
}

.console-table {
  overflow-y: auto;
  flex: 1;
}
.console-table table {
  width: 100%;
  text-align: left;
  border-collapse: collapse;
}
.console-table th {
  padding: 12px 16px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.console-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}
.participant-info .name { font-weight: 500; display: block; }
.participant-info .role { font-size: 11px; color: var(--text-muted); }
.mono { font-family: monospace; font-size: 12px; }

.talking-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #16a34a;
}
.silent-badge {
  font-size: 11px;
  color: var(--text-muted);
}

/* Form */
.form-group {
  margin-bottom: 16px;
}
.form-group label {
  display: block;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 6px;
}
.form-section h4 {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
}
.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 14px;
  width: 100%;
}
.divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
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
.toggle-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  font-weight: normal;
  text-transform: none;
  color: var(--text-main);
  margin-bottom: 8px;
  cursor: pointer;
}
.toggle-row input { width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }
.icon-md { width: 24px; height: 24px; }
.icon-xs { width: 12px; height: 12px; }
</style>
