<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Call Detail Records</h2>
      <p class="text-muted text-sm">Search and analyze call history with detailed records.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="exportCSV">
        <DownloadIcon class="btn-icon-left" />
        Export CSV
      </button>
    </div>
  </div>

  <!-- Stats Row -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-icon total"><PhoneIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ records.length }}</span>
        <span class="stat-label">Total Calls</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon answered"><PhoneCallIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ records.filter(r => r.status === 'Answered').length }}</span>
        <span class="stat-label">Answered</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon missed"><PhoneMissedIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ records.filter(r => r.status === 'Missed' || r.status === 'No Answer').length }}</span>
        <span class="stat-label">Missed</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon voicemail"><VoicemailIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ records.filter(r => r.status === 'Voicemail').length }}</span>
        <span class="stat-label">Voicemail</span>
      </div>
    </div>
  </div>

  <!-- Filters -->
  <div class="filter-bar">
    <div class="search-box">
      <SearchIcon class="search-icon" />
      <input type="text" v-model="searchQuery" placeholder="Search by name, number, extension..." class="search-input">
    </div>
    
    <div class="filter-group">
      <label>Date Range</label>
      <input type="date" v-model="filters.startDate" class="filter-input">
      <span class="filter-sep">to</span>
      <input type="date" v-model="filters.endDate" class="filter-input">
    </div>

    <select v-model="filters.extension" class="filter-select">
      <option value="">All Extensions</option>
      <option v-for="ext in extensions" :key="ext" :value="ext">{{ ext }}</option>
    </select>

    <select v-model="filters.status" class="filter-select">
      <option value="">All Statuses</option>
      <option value="Answered">Answered</option>
      <option value="Missed">Missed</option>
      <option value="No Answer">No Answer</option>
      <option value="Voicemail">Voicemail</option>
      <option value="Busy">Busy</option>
      <option value="Cancelled">Cancelled</option>
      <option value="Failed">Failed</option>
    </select>

    <select v-model="filters.direction" class="filter-select">
      <option value="">All Directions</option>
      <option value="inbound">Inbound</option>
      <option value="outbound">Outbound</option>
      <option value="local">Local</option>
    </select>

    <button class="btn-secondary small" @click="resetFilters">Clear</button>
  </div>

  <!-- Records Table -->
  <div class="table-container">
    <table class="data-table">
      <thead>
        <tr>
          <th @click="sortBy('direction')" class="sortable">Direction</th>
          <th @click="sortBy('extension')" class="sortable">Extension</th>
          <th @click="sortBy('callerName')" class="sortable">Caller Name</th>
          <th @click="sortBy('callerNumber')" class="sortable">Caller Number</th>
          <th>Destination</th>
          <th>Recording</th>
          <th @click="sortBy('dateTime')" class="sortable">Date & Time</th>
          <th>TTA</th>
          <th>PDD</th>
          <th>MOS</th>
          <th @click="sortBy('duration')" class="sortable">Duration</th>
          <th @click="sortBy('status')" class="sortable">Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="record in filteredRecords" :key="record.id" @click="viewDetails(record)">
          <td>
            <span class="direction-badge" :class="record.direction">
              <PhoneIncomingIcon v-if="record.direction === 'inbound'" class="icon-xs" />
              <PhoneOutgoingIcon v-else-if="record.direction === 'outbound'" class="icon-xs" />
              <PhoneIcon v-else class="icon-xs" />
              {{ record.direction }}
            </span>
          </td>
          <td class="mono">{{ record.extension }}</td>
          <td>{{ record.callerName }}</td>
          <td class="mono">{{ record.callerNumber }}</td>
          <td class="mono">{{ record.destination }}</td>
          <td>
            <button v-if="record.hasRecording" class="btn-icon small" @click.stop="playRecording(record)">
              <PlayIcon class="icon-xs" />
            </button>
            <span v-else class="no-recording">—</span>
          </td>
          <td class="meta">{{ record.dateTime }}</td>
          <td class="meta">{{ record.tta }}s</td>
          <td class="meta">{{ record.pdd }}ms</td>
          <td>
            <span class="mos-badge" :class="getMOSClass(record.mos)">{{ record.mos }}</span>
          </td>
          <td class="mono">{{ formatDuration(record.duration) }}</td>
          <td>
            <span class="status-badge" :class="record.status.toLowerCase().replace(' ', '-')">
              {{ record.status }}
            </span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Pagination -->
  <div class="pagination">
    <span class="page-info">Showing {{ filteredRecords.length }} of {{ records.length }} records</span>
    <div class="page-controls">
      <button class="page-btn" :disabled="currentPage === 1" @click="currentPage--">Previous</button>
      <span class="page-num">Page {{ currentPage }}</span>
      <button class="page-btn" @click="currentPage++">Next</button>
    </div>
  </div>

  <!-- Detail Modal -->
  <div class="modal-overlay" v-if="showDetailModal" @click.self="showDetailModal = false">
    <div class="modal-card detail-modal">
      <div class="modal-header">
        <h3>Call Details</h3>
        <button class="btn-icon" @click="showDetailModal = false"><XIcon class="icon-sm" /></button>
      </div>
      <div class="modal-body" v-if="selectedRecord">
        <div class="detail-grid">
          <div class="detail-item">
            <label>Direction</label>
            <span>{{ selectedRecord.direction }}</span>
          </div>
          <div class="detail-item">
            <label>Extension</label>
            <span>{{ selectedRecord.extension }}</span>
          </div>
          <div class="detail-item">
            <label>Caller Name</label>
            <span>{{ selectedRecord.callerName }}</span>
          </div>
          <div class="detail-item">
            <label>Caller Number</label>
            <span>{{ selectedRecord.callerNumber }}</span>
          </div>
          <div class="detail-item">
            <label>Destination</label>
            <span>{{ selectedRecord.destination }}</span>
          </div>
          <div class="detail-item">
            <label>Date & Time</label>
            <span>{{ selectedRecord.dateTime }}</span>
          </div>
          <div class="detail-item">
            <label>Duration</label>
            <span>{{ formatDuration(selectedRecord.duration) }}</span>
          </div>
          <div class="detail-item">
            <label>Status</label>
            <span class="status-badge" :class="selectedRecord.status.toLowerCase()">{{ selectedRecord.status }}</span>
          </div>
          <div class="detail-item">
            <label>Codec</label>
            <span>{{ selectedRecord.codec }}</span>
          </div>
          <div class="detail-item">
            <label>MOS Score</label>
            <span class="mos-badge" :class="getMOSClass(selectedRecord.mos)">{{ selectedRecord.mos }}</span>
          </div>
        </div>
        
        <div class="recording-player" v-if="selectedRecord.hasRecording">
          <label>Recording</label>
          <div class="audio-player">
            <button class="play-btn" @click="togglePlay">
              <PlayIcon v-if="!isPlaying" />
              <PauseIcon v-else />
            </button>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: '35%' }"></div>
            </div>
            <span class="time">0:35 / 1:42</span>
            <button class="btn-icon small" @click="downloadRecording">
              <DownloadIcon class="icon-xs" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { 
  Phone as PhoneIcon, PhoneCall as PhoneCallIcon, PhoneMissed as PhoneMissedIcon,
  PhoneIncoming as PhoneIncomingIcon, PhoneOutgoing as PhoneOutgoingIcon,
  Voicemail as VoicemailIcon, Search as SearchIcon, Download as DownloadIcon,
  Play as PlayIcon, Pause as PauseIcon, X as XIcon
} from 'lucide-vue-next'
import { cdrAPI } from '@/services/api'

const toast = inject('toast')
const isLoading = ref(false)

const searchQuery = ref('')
const currentPage = ref(1)
const showDetailModal = ref(false)
const selectedRecord = ref(null)
const isPlaying = ref(false)

const filters = ref({
  startDate: '',
  endDate: '',
  extension: '',
  status: '',
  direction: ''
})

const extensions = ['101', '102', '103', '104', '105', '200', '201']
const records = ref([])

onMounted(async () => {
  await fetchCDR()
})

async function fetchCDR() {
  isLoading.value = true
  try {
    const response = await cdrAPI.list({
      page: currentPage.value,
      extension: filters.value.extension,
      start_date: filters.value.startDate,
      end_date: filters.value.endDate,
      direction: filters.value.direction,
    })
    records.value = (response.data || []).map(r => ({
      id: r.id,
      direction: r.direction || 'inbound',
      extension: r.caller_id_number || '',
      callerName: r.caller_id_name || 'Unknown',
      callerNumber: r.caller_id_number || '',
      destination: r.destination_number || '',
      hasRecording: !!r.recording_path,
      dateTime: formatDateTime(r.start_stamp),
      codec: r.codec || 'PCMU/8000',
      tta: r.answer_seconds || 0,
      pdd: r.pdd_ms || 0,
      mos: r.mos || 0,
      duration: r.billsec || 0,
      status: mapHangupCause(r.hangup_cause)
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load CDR')
    // Fallback to demo data
    records.value = [
      { id: 1, direction: 'inbound', extension: '101', callerName: 'John Smith', callerNumber: '+1 555-123-4567', destination: '101', hasRecording: true, dateTime: 'Dec 10, 2024 08:45 AM', codec: 'PCMU/8000', tta: 3, pdd: 180, mos: 4.2, duration: 185, status: 'Answered' },
      { id: 2, direction: 'outbound', extension: '102', callerName: 'Jane Doe', callerNumber: '+1 555-987-6543', destination: '+1 555-111-2222', hasRecording: true, dateTime: 'Dec 10, 2024 08:30 AM', codec: 'G722/8000', tta: 5, pdd: 210, mos: 4.5, duration: 342, status: 'Answered' },
    ]
  } finally {
    isLoading.value = false
  }
}

function formatDateTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit', hour12: true })
}

function mapHangupCause(cause) {
  const map = {
    'NORMAL_CLEARING': 'Answered',
    'NO_ANSWER': 'No Answer',
    'USER_BUSY': 'Busy',
    'ORIGINATOR_CANCEL': 'Cancelled',
    'NO_USER_RESPONSE': 'Missed',
  }
  return map[cause] || 'Answered'
}

const filteredRecords = computed(() => {
  return records.value.filter(r => {
    const matchesSearch = !searchQuery.value || 
      r.callerName.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      r.callerNumber.includes(searchQuery.value) ||
      r.extension.includes(searchQuery.value)
    const matchesExt = !filters.value.extension || r.extension === filters.value.extension
    const matchesStatus = !filters.value.status || r.status === filters.value.status
    const matchesDir = !filters.value.direction || r.direction === filters.value.direction
    return matchesSearch && matchesExt && matchesStatus && matchesDir
  })
})

const formatDuration = (seconds) => {
  if (!seconds) return '0:00'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

const getMOSClass = (mos) => {
  if (mos >= 4) return 'good'
  if (mos >= 3) return 'ok'
  if (mos > 0) return 'bad'
  return ''
}

const resetFilters = () => {
  filters.value = { startDate: '', endDate: '', extension: '', status: '', direction: '' }
  searchQuery.value = ''
}

const viewDetails = (record) => {
  selectedRecord.value = record
  showDetailModal.value = true
}

const playRecording = (record) => {
  toast?.info(`Playing recording for call ${record.id}`)
}

const togglePlay = () => {
  isPlaying.value = !isPlaying.value
}

const downloadRecording = () => {
  toast?.info('Downloading recording...')
}

const exportCSV = async () => {
  try {
    await cdrAPI.export(filters.value)
    toast?.success('CDR exported to CSV')
  } catch (error) {
    toast?.error(error.message, 'Failed to export CDR')
  }
}

const sortBy = (field) => {
  // Implement sorting logic
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.total { background: #dbeafe; color: #2563eb; }
.stat-icon.answered { background: #dcfce7; color: #16a34a; }
.stat-icon.missed { background: #fee2e2; color: #dc2626; }
.stat-icon.voicemail { background: #fef3c7; color: #b45309; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-muted); }

.filter-bar { display: flex; gap: 12px; flex-wrap: wrap; align-items: flex-end; margin-bottom: var(--spacing-md); padding: 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }
.search-box { position: relative; flex: 1; min-width: 200px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 12px 8px 36px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); }
.filter-group { display: flex; align-items: center; gap: 8px; }
.filter-group label { font-size: 11px; font-weight: 600; color: var(--text-muted); }
.filter-input { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); }
.filter-sep { font-size: 12px; color: var(--text-muted); }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); background: white; }

.table-container { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { background: var(--bg-app); padding: 10px 12px; text-align: left; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); }
.data-table th.sortable { cursor: pointer; }
.data-table th.sortable:hover { color: var(--primary-color); }
.data-table td { padding: 10px 12px; border-bottom: 1px solid var(--border-color); }
.data-table tr:hover { background: #f8fafc; cursor: pointer; }
.data-table tr:last-child td { border-bottom: none; }
.mono { font-family: monospace; font-size: 12px; }
.meta { font-size: 11px; color: var(--text-muted); }

.direction-badge { display: inline-flex; align-items: center; gap: 4px; font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 99px; text-transform: capitalize; }
.direction-badge.inbound { background: #dcfce7; color: #16a34a; }
.direction-badge.outbound { background: #dbeafe; color: #2563eb; }
.direction-badge.local { background: #f3f4f6; color: #6b7280; }
.icon-xs { width: 12px; height: 12px; }

.status-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 99px; }
.status-badge.answered { background: #dcfce7; color: #16a34a; }
.status-badge.missed, .status-badge.no-answer { background: #fee2e2; color: #dc2626; }
.status-badge.voicemail { background: #fef3c7; color: #b45309; }
.status-badge.busy { background: #fce7f3; color: #be185d; }
.status-badge.cancelled { background: #f3f4f6; color: #6b7280; }
.status-badge.failed { background: #fef2f2; color: #dc2626; }

.mos-badge { font-size: 11px; font-weight: 600; font-family: monospace; }
.mos-badge.good { color: #16a34a; }
.mos-badge.ok { color: #b45309; }
.mos-badge.bad { color: #dc2626; }

.no-recording { color: var(--text-muted); }

.pagination { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; padding: 12px 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }
.page-info { font-size: 12px; color: var(--text-muted); }
.page-controls { display: flex; align-items: center; gap: 12px; }
.page-btn { padding: 6px 12px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 12px; cursor: pointer; }
.page-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.page-num { font-size: 12px; color: var(--text-muted); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 600px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; }

.detail-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; margin-bottom: 20px; }
.detail-item { display: flex; flex-direction: column; gap: 4px; }
.detail-item label { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.detail-item span { font-size: 13px; color: var(--text-primary); }

.recording-player { padding-top: 16px; border-top: 1px solid var(--border-color); }
.recording-player label { display: block; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); margin-bottom: 8px; }
.audio-player { display: flex; align-items: center; gap: 12px; padding: 12px; background: var(--bg-app); border-radius: var(--radius-sm); }
.play-btn { width: 36px; height: 36px; border-radius: 50%; background: var(--primary-color); color: white; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.progress-bar { flex: 1; height: 6px; background: #e2e8f0; border-radius: 3px; overflow: hidden; }
.progress-fill { height: 100%; background: var(--primary-color); }
.time { font-size: 11px; font-family: monospace; color: var(--text-muted); }

.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary.small { padding: 6px 12px; font-size: 12px; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; display: flex; }
.btn-icon.small { padding: 4px; }
.btn-icon-left { width: 14px; height: 14px; }
.icon-sm { width: 16px; height: 16px; }
</style>
