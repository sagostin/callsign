<template>
  <div class="conferences-page">
    <div class="view-header">
      <div class="header-content">
        <h2>My Conferences</h2>
        <p class="text-muted text-sm">Create, manage, and join your conference rooms and meetings.</p>
      </div>
      <button class="btn-primary" @click="showCreateModal = true">+ New Conference</button>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon rooms"><UsersIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ conferences.length }}</span>
          <span class="stat-label">My Rooms</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon active"><PhoneCallIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ conferences.filter(c => c.status === 'active').length }}</span>
          <span class="stat-label">Active Now</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon scheduled"><CalendarIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ scheduledMeetings.length }}</span>
          <span class="stat-label">Scheduled</span>
        </div>
      </div>
    </div>

    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'rooms' }" @click="activeTab = 'rooms'">My Rooms</button>
      <button class="tab" :class="{ active: activeTab === 'scheduled' }" @click="activeTab = 'scheduled'">Scheduled</button>
      <button class="tab" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">History</button>
    </div>

    <!-- MY ROOMS TAB -->
    <div class="tab-content" v-if="activeTab === 'rooms'">
      <div class="rooms-grid">
        <div class="room-card" v-for="conf in conferences" :key="conf.id" :class="{ active: conf.status === 'active' }">
          <div class="room-header">
            <div class="room-icon" :class="{ active: conf.status === 'active' }">
              <UsersIcon class="icon-md" />
            </div>
            <div class="room-info">
              <h4>{{ conf.name }}</h4>
              <div class="room-meta">
                <span class="room-ext">Ext: {{ conf.extension }}</span>
                <span class="meta-dot"></span>
                <span class="room-pin" v-if="conf.pin">PIN: {{ conf.pin }}</span>
                <span class="room-pin" v-else>No PIN</span>
              </div>
            </div>
            <div class="room-status" :class="conf.status">
              <span class="status-dot"></span>
              {{ conf.status === 'active' ? `${conf.participants} Participants` : 'Idle' }}
            </div>
          </div>

          <div class="room-details" v-if="conf.status === 'active'">
            <div class="participants-preview">
              <div class="participant-avatar" v-for="i in Math.min(conf.participants, 4)" :key="i">{{ i }}</div>
              <span class="more-participants" v-if="conf.participants > 4">+{{ conf.participants - 4 }}</span>
            </div>
            <span class="call-duration">{{ conf.duration }}</span>
          </div>

          <div class="room-actions">
            <button class="btn-action join" @click="joinConference(conf)">
              <PhoneCallIcon class="icon-sm" /> Join
            </button>
            <button class="btn-action share" @click="shareConference(conf)">
              <ShareIcon class="icon-sm" /> Share
            </button>
            <button class="btn-action settings" @click="editConference(conf)">
              <SettingsIcon class="icon-sm" />
            </button>
          </div>
        </div>

        <!-- Quick Join Card -->
        <div class="room-card add-card" @click="showCreateModal = true">
          <div class="add-icon"><PlusIcon class="icon-lg" /></div>
          <span>Create New Room</span>
        </div>
      </div>
    </div>

    <!-- SCHEDULED TAB -->
    <div class="tab-content" v-else-if="activeTab === 'scheduled'">
      <div class="scheduled-list">
        <div class="scheduled-card" v-for="meeting in scheduledMeetings" :key="meeting.id">
          <div class="meeting-date">
            <span class="date-day">{{ meeting.day }}</span>
            <span class="date-month">{{ meeting.month }}</span>
          </div>
          <div class="meeting-info">
            <h4>{{ meeting.title }}</h4>
            <div class="meeting-meta">
              <ClockIcon class="icon-xs" />
              <span>{{ meeting.time }} ({{ meeting.duration }})</span>
            </div>
            <div class="meeting-participants">
              <span class="participant-count">{{ meeting.invitees.length }} invited</span>
              <div class="participant-avatars">
                <div class="avatar" v-for="(inv, i) in meeting.invitees.slice(0, 3)" :key="i">{{ inv.charAt(0) }}</div>
              </div>
            </div>
          </div>
          <div class="meeting-actions">
            <button class="btn-secondary small" @click="startMeeting(meeting)">Start Now</button>
            <button class="btn-link" @click="editMeeting(meeting)">Edit</button>
            <button class="btn-link text-bad" @click="cancelMeeting(meeting)">Cancel</button>
          </div>
        </div>

        <div class="empty-state" v-if="scheduledMeetings.length === 0">
          <CalendarIcon class="empty-icon" />
          <p>No scheduled meetings</p>
          <button class="btn-link" @click="showScheduleModal = true">Schedule one now</button>
        </div>
      </div>

      <button class="btn-fab" @click="showScheduleModal = true">
        <PlusIcon class="icon-sm" />
      </button>
    </div>

    <!-- HISTORY TAB -->
    <div class="tab-content" v-else-if="activeTab === 'history'">
      <DataTable :columns="historyColumns" :data="meetingHistory" actions>
        <template #duration="{ value }">
          <span class="mono">{{ value }}</span>
        </template>
        <template #participants="{ value }">
          <span class="participant-badge">{{ value }} people</span>
        </template>
        <template #actions="{ row }">
          <button class="btn-link" v-if="row.recording" @click="playRecording(row)">Recording</button>
          <button class="btn-link" @click="viewDetails(row)">Details</button>
        </template>
      </DataTable>
    </div>

    <!-- CREATE ROOM MODAL -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Create Conference Room</h3>
          <button class="btn-icon" @click="showCreateModal = false"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label>Room Name</label>
            <input v-model="newRoom.name" class="input-field" placeholder="e.g. Team Standup">
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Extension</label>
              <input v-model="newRoom.extension" class="input-field code" placeholder="3050">
            </div>
            <div class="form-group">
              <label>PIN (Optional)</label>
              <input v-model="newRoom.pin" class="input-field code" placeholder="1234">
            </div>
          </div>

          <div class="form-group">
            <label>Max Participants</label>
            <select v-model="newRoom.maxParticipants" class="input-field">
              <option value="10">10 participants</option>
              <option value="25">25 participants</option>
              <option value="50">50 participants</option>
              <option value="100">100 participants</option>
            </select>
          </div>

          <div class="form-section">
            <h4>Options</h4>
            <div class="checkbox-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="newRoom.recordCalls">
                <span>Record all calls</span>
              </label>
              <label class="checkbox-row">
                <input type="checkbox" v-model="newRoom.announceJoin">
                <span>Announce when participants join/leave</span>
              </label>
              <label class="checkbox-row">
                <input type="checkbox" v-model="newRoom.muteOnEntry">
                <span>Mute participants on entry</span>
              </label>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="showCreateModal = false">Cancel</button>
          <button class="btn-primary" @click="createRoom" :disabled="!newRoom.name">Create Room</button>
        </div>
      </div>
    </div>

    <!-- SHARE MODAL -->
    <div v-if="showShareModal" class="modal-overlay" @click.self="showShareModal = false">
      <div class="modal-card small">
        <div class="modal-header">
          <h3>Share Conference</h3>
          <button class="btn-icon" @click="showShareModal = false"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body">
          <div class="share-info">
            <div class="share-item">
              <span class="share-label">Dial-In Number</span>
              <div class="share-value">
                <span>+1 (415) 555-3000</span>
                <button class="btn-icon" @click="copy('+14155553000')"><CopyIcon class="icon-xs" /></button>
              </div>
            </div>
            <div class="share-item">
              <span class="share-label">Extension</span>
              <div class="share-value">
                <span>{{ sharingConference?.extension }}</span>
                <button class="btn-icon" @click="copy(sharingConference?.extension)"><CopyIcon class="icon-xs" /></button>
              </div>
            </div>
            <div class="share-item" v-if="sharingConference?.pin">
              <span class="share-label">PIN</span>
              <div class="share-value">
                <span>{{ sharingConference?.pin }}</span>
                <button class="btn-icon" @click="copy(sharingConference?.pin)"><CopyIcon class="icon-xs" /></button>
              </div>
            </div>
          </div>

          <div class="share-link">
            <input class="input-field" :value="`https://meet.callsign.io/${sharingConference?.extension}`" readonly>
            <button class="btn-secondary small" @click="copyLink">Copy Link</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { 
  Users as UsersIcon, PhoneCall as PhoneCallIcon, Calendar as CalendarIcon,
  Plus as PlusIcon, Share2 as ShareIcon, Settings as SettingsIcon,
  Clock as ClockIcon, X as XIcon, Copy as CopyIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'

const activeTab = ref('rooms')
const showCreateModal = ref(false)
const showScheduleModal = ref(false)
const showShareModal = ref(false)
const sharingConference = ref(null)

const newRoom = ref({
  name: '',
  extension: '',
  pin: '',
  maxParticipants: '25',
  recordCalls: false,
  announceJoin: true,
  muteOnEntry: false
})

const conferences = ref([
  { id: 1, name: 'Personal Meeting Room', extension: '3050', pin: '1234', status: 'idle', participants: 0, duration: '' },
  { id: 2, name: 'Project Team Sync', extension: '3051', pin: '', status: 'active', participants: 5, duration: '00:23:45' },
  { id: 3, name: 'Client Calls', extension: '3052', pin: '9876', status: 'idle', participants: 0, duration: '' },
])

const scheduledMeetings = ref([
  { id: 1, title: 'Weekly Standup', day: '12', month: 'Dec', time: '10:00 AM', duration: '30 min', invitees: ['John', 'Sarah', 'Mike', 'Lisa'] },
  { id: 2, title: 'Client Review', day: '13', month: 'Dec', time: '2:00 PM', duration: '1 hour', invitees: ['Client Team', 'Sales'] },
])

const historyColumns = [
  { key: 'date', label: 'Date', width: '120px' },
  { key: 'name', label: 'Conference' },
  { key: 'duration', label: 'Duration', width: '100px' },
  { key: 'participants', label: 'Participants', width: '120px' },
  { key: 'recording', label: 'Recording', width: '100px' }
]

const meetingHistory = ref([
  { date: 'Dec 10, 2024', name: 'Team Sync', duration: '00:45:12', participants: 6, recording: true },
  { date: 'Dec 9, 2024', name: 'Client Call', duration: '01:02:33', participants: 3, recording: true },
  { date: 'Dec 8, 2024', name: 'Personal Room', duration: '00:15:00', participants: 2, recording: false },
])

const joinConference = (conf) => alert(`Joining ${conf.name}...`)
const shareConference = (conf) => {
  sharingConference.value = conf
  showShareModal.value = true
}
const editConference = (conf) => alert(`Edit ${conf.name}`)
const startMeeting = (meeting) => alert(`Starting ${meeting.title}`)
const editMeeting = (meeting) => alert(`Edit ${meeting.title}`)
const cancelMeeting = (meeting) => {
  if (confirm(`Cancel "${meeting.title}"?`)) {
    scheduledMeetings.value = scheduledMeetings.value.filter(m => m.id !== meeting.id)
  }
}
const playRecording = (row) => alert(`Playing recording for ${row.name}`)
const viewDetails = (row) => alert(`Details for ${row.name}`)
const createRoom = () => {
  conferences.value.push({
    id: Date.now(),
    ...newRoom.value,
    status: 'idle',
    participants: 0,
    duration: ''
  })
  showCreateModal.value = false
  newRoom.value = { name: '', extension: '', pin: '', maxParticipants: '25', recordCalls: false, announceJoin: true, muteOnEntry: false }
}
const copy = (text) => {
  navigator.clipboard.writeText(text)
  alert('Copied!')
}
const copyLink = () => copy(`https://meet.callsign.io/${sharingConference.value?.extension}`)
</script>

<style scoped>
.conferences-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.rooms { background: #dbeafe; color: #2563eb; }
.stat-icon.active { background: #dcfce7; color: #16a34a; }
.stat-icon.scheduled { background: #fef3c7; color: #b45309; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); }

/* Rooms Grid */
.rooms-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }

.room-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.room-card.active { border-color: #22c55e; background: #f0fdf4; }

.room-card.add-card { border-style: dashed; align-items: center; justify-content: center; min-height: 180px; cursor: pointer; color: var(--text-muted); }
.room-card.add-card:hover { border-color: var(--primary-color); color: var(--primary-color); }
.add-icon { width: 48px; height: 48px; background: var(--bg-app); border-radius: 50%; display: flex; align-items: center; justify-content: center; margin-bottom: 8px; }

.room-header { display: flex; align-items: flex-start; gap: 12px; }
.room-icon { width: 40px; height: 40px; background: var(--primary-light); border-radius: 10px; display: flex; align-items: center; justify-content: center; color: var(--primary-color); }
.room-icon.active { background: #dcfce7; color: #16a34a; }
.room-info { flex: 1; }
.room-info h4 { font-size: 14px; font-weight: 600; margin: 0 0 4px; }
.room-meta { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-muted); font-family: monospace; }
.meta-dot { width: 3px; height: 3px; background: var(--text-muted); border-radius: 50%; }

.room-status { font-size: 11px; font-weight: 600; padding: 4px 8px; border-radius: 4px; background: var(--bg-app); color: var(--text-muted); display: flex; align-items: center; gap: 6px; }
.room-status.active { background: #dcfce7; color: #16a34a; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
.room-status.active .status-dot { animation: pulse 1.5s infinite; }

.room-details { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-top: 1px solid var(--border-color); }
.participants-preview { display: flex; align-items: center; }
.participant-avatar { width: 28px; height: 28px; background: var(--primary-light); color: var(--primary-color); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 700; margin-left: -8px; border: 2px solid white; }
.participant-avatar:first-child { margin-left: 0; }
.more-participants { font-size: 11px; color: var(--text-muted); margin-left: 8px; }
.call-duration { font-family: monospace; font-size: 12px; color: #16a34a; font-weight: 600; }

.room-actions { display: flex; gap: 8px; }
.btn-action { flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px; padding: 8px 12px; border-radius: 6px; font-size: 12px; font-weight: 600; border: 1px solid var(--border-color); background: white; cursor: pointer; }
.btn-action.join { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.btn-action.share { background: var(--bg-app); }
.btn-action.settings { flex: none; width: 36px; }

/* Scheduled */
.scheduled-list { display: flex; flex-direction: column; gap: 12px; }
.scheduled-card { display: flex; gap: 16px; padding: 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }

.meeting-date { display: flex; flex-direction: column; align-items: center; justify-content: center; width: 50px; padding: 8px; background: var(--primary-light); border-radius: 8px; }
.date-day { font-size: 20px; font-weight: 700; color: var(--primary-color); }
.date-month { font-size: 11px; color: var(--text-muted); text-transform: uppercase; }

.meeting-info { flex: 1; }
.meeting-info h4 { font-size: 14px; font-weight: 600; margin: 0 0 6px; }
.meeting-meta { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-muted); margin-bottom: 6px; }
.meeting-participants { display: flex; align-items: center; gap: 8px; }
.participant-count { font-size: 11px; color: var(--text-muted); }
.participant-avatars { display: flex; }
.avatar { width: 24px; height: 24px; background: var(--bg-app); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 600; margin-left: -6px; border: 2px solid white; }
.avatar:first-child { margin-left: 0; }

.meeting-actions { display: flex; flex-direction: column; gap: 4px; align-items: flex-end; }

.empty-state { text-align: center; padding: 48px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 16px; }

.btn-fab { position: fixed; bottom: 24px; right: 24px; width: 48px; height: 48px; border-radius: 50%; background: var(--primary-color); color: white; border: none; display: flex; align-items: center; justify-content: center; box-shadow: var(--shadow-lg); cursor: pointer; }

/* Share */
.share-info { display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px; }
.share-item { display: flex; justify-content: space-between; align-items: center; padding: 12px; background: var(--bg-app); border-radius: 6px; }
.share-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; font-weight: 600; }
.share-value { display: flex; align-items: center; gap: 8px; font-family: monospace; }
.share-link { display: flex; gap: 8px; }
.share-link .input-field { flex: 1; font-size: 12px; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-secondary.small { padding: 6px 12px; font-size: 12px; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 500; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.text-bad { color: var(--status-bad); }

/* Form */
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.form-section { margin-top: 16px; }
.form-section h4 { font-size: 13px; font-weight: 600; margin: 0 0 8px; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.code { font-family: monospace; }
.checkbox-group { display: flex; flex-direction: column; gap: 8px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); }
.modal-card { background: white; border-radius: var(--radius-md); box-shadow: var(--shadow-lg); width: 100%; max-width: 480px; }
.modal-card.small { max-width: 400px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.icon-sm { width: 16px; height: 16px; }
.icon-md { width: 20px; height: 20px; }
.icon-lg { width: 24px; height: 24px; }
.icon-xs { width: 14px; height: 14px; }
.mono { font-family: monospace; }
.participant-badge { background: var(--bg-app); padding: 2px 8px; border-radius: 4px; font-size: 12px; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
