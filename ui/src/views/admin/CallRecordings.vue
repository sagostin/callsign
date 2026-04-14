<template>
  <div class="call-recordings-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Call Recordings</h2>
        <p class="text-muted text-sm">Browse and manage recorded calls.</p>
      </div>
      <div class="header-actions">
        <input type="text" v-model="search" class="search-input" placeholder="Search by name, number, or ext...">
        <button class="btn-secondary" @click="showFilters = !showFilters">
          <FilterIcon class="btn-icon" /> Filters
        </button>
        <button class="btn-secondary" @click="exportRecordings">
          <DownloadIcon class="btn-icon" /> Export
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <span class="stat-value">{{ totalRecordings.toLocaleString() }}</span>
        <span class="stat-label">Total Recordings</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ todayRecordings }}</span>
        <span class="stat-label">Today</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ storageUsed }}</span>
        <span class="stat-label">Storage Used</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">00:02:34</span>
        <span class="stat-label">Avg Duration</span>
      </div>
    </div>

    <!-- Filters Panel -->
    <div class="filters-panel" v-if="showFilters">
      <div class="filter-group">
        <label>Date Range</label>
        <select v-model="filters.dateRange" class="input-field">
          <option value="today">Today</option>
          <option value="week">This Week</option>
          <option value="month">This Month</option>
          <option value="custom">Custom Range</option>
        </select>
      </div>
      <div class="filter-group">
        <label>Direction</label>
        <select v-model="filters.direction" class="input-field">
          <option value="">All</option>
          <option value="inbound">Inbound</option>
          <option value="outbound">Outbound</option>
          <option value="internal">Internal</option>
        </select>
      </div>
      <div class="filter-group">
        <label>Extension</label>
        <select v-model="filters.extension" class="input-field">
          <option value="">All Extensions</option>
          <option v-for="ext in extensionsList" :key="ext.extension" :value="ext.extension">
            {{ ext.extension }} - {{ ext.name || ext.first_name || 'Extension' }}
          </option>
        </select>
      </div>
      <div class="filter-group">
        <label>Min Duration</label>
        <select v-model="filters.minDuration" class="input-field">
          <option value="0">Any</option>
          <option value="30">30s+</option>
          <option value="60">1 min+</option>
          <option value="300">5 min+</option>
        </select>
      </div>
    </div>

    <!-- Recordings Table -->
    <div class="recordings-table">
      <table>
        <thead>
          <tr>
            <th style="width: 40px;"><input type="checkbox" @change="selectAll"></th>
            <th>Date / Time</th>
            <th>Direction</th>
            <th>From</th>
            <th>To</th>
            <th>Duration</th>
            <th>Size</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rec in filteredRecordings" :key="rec.id" @click="selectRecording(rec)" :class="{ selected: selectedIds.includes(rec.id) }">
            <td @click.stop><input type="checkbox" :checked="selectedIds.includes(rec.id)" @change="toggleSelect(rec.id)"></td>
            <td>
              <div class="datetime-cell">
                <span class="date">{{ rec.date }}</span>
                <span class="time">{{ rec.time }}</span>
              </div>
            </td>
            <td>
              <span class="direction-badge" :class="rec.direction">
                <PhoneIncomingIcon v-if="rec.direction === 'inbound'" class="icon-xs" />
                <PhoneOutgoingIcon v-else-if="rec.direction === 'outbound'" class="icon-xs" />
                <ArrowRightLeft v-else class="icon-xs" />
                {{ rec.direction }}
              </span>
            </td>
            <td>
              <div class="party-cell">
                <span class="number">{{ rec.from }}</span>
                <span class="name" v-if="rec.fromName">{{ rec.fromName }}</span>
              </div>
            </td>
            <td>
              <div class="party-cell">
                <span class="number">{{ rec.to }}</span>
                <span class="name" v-if="rec.toName">{{ rec.toName }}</span>
              </div>
            </td>
            <td class="mono">{{ rec.duration }}</td>
            <td class="mono">{{ rec.size }}</td>
            <td class="actions-cell" @click.stop>
              <button class="action-btn" @click="playRecording(rec)" title="Play"><PlayIcon class="icon-sm" /></button>
              <button class="action-btn" @click="downloadRecording(rec)" title="Download"><DownloadIcon class="icon-sm" /></button>
              <button class="action-btn danger" @click="deleteRecording(rec)" title="Delete"><TrashIcon class="icon-sm" /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="pagination">
      <span class="page-info">Showing 1-20 of 1,247 recordings</span>
      <div class="page-controls">
        <button class="page-btn" disabled><ChevronLeftIcon class="icon-sm" /></button>
        <button class="page-btn active">1</button>
        <button class="page-btn">2</button>
        <button class="page-btn">3</button>
        <span class="dots">...</span>
        <button class="page-btn">63</button>
        <button class="page-btn"><ChevronRightIcon class="icon-sm" /></button>
      </div>
    </div>

    <!-- Player Modal -->
    <div v-if="showPlayer" class="modal-overlay" @click.self="showPlayer = false">
      <div class="player-modal">
        <div class="player-header">
          <h3>Call Recording</h3>
          <button class="btn-icon" @click="showPlayer = false"><XIcon class="icon-sm" /></button>
        </div>
        <div class="player-info">
          <div class="call-parties">
            <span>{{ currentRecording?.from }}</span>
            <ArrowRightIcon class="arrow" />
            <span>{{ currentRecording?.to }}</span>
          </div>
          <div class="call-meta">
            {{ currentRecording?.date }} {{ currentRecording?.time }} · {{ currentRecording?.duration }}
          </div>
        </div>
        <div class="audio-player">
          <div class="waveform"></div>
          <div class="player-controls">
            <button class="play-btn"><PlayIcon class="icon-lg" /></button>
            <div class="time-display">
              <span>00:00</span> / <span>{{ currentRecording?.duration }}</span>
            </div>
            <div class="volume-control">
              <VolumeIcon class="icon-sm" />
              <input type="range" min="0" max="100" value="80">
            </div>
          </div>
        </div>
        <div class="player-actions">
          <button class="btn-secondary" @click="downloadRecording(currentRecording)">
            <DownloadIcon class="btn-icon" /> Download
          </button>
          <button class="btn-secondary">
            <MailIcon class="btn-icon" /> Email
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import {
  Filter as FilterIcon, Download as DownloadIcon, Play as PlayIcon,
  Trash2 as TrashIcon, PhoneIncoming as PhoneIncomingIcon,
  PhoneOutgoing as PhoneOutgoingIcon, ArrowRightLeft,
  ChevronLeft as ChevronLeftIcon, ChevronRight as ChevronRightIcon,
  X as XIcon, ArrowRight as ArrowRightIcon, Volume2 as VolumeIcon,
  Mail as MailIcon
} from 'lucide-vue-next'
import { recordingsAPI, extensionsAPI } from '../../services/api'

const toast = inject('toast')

const search = ref('')
const showFilters = ref(false)
const showPlayer = ref(false)
const currentRecording = ref(null)
const selectedIds = ref([])

const totalRecordings = ref(0)
const todayRecordings = ref(0)
const storageUsed = ref('0 MB')
const extensionsList = ref([])

const filters = ref({
  dateRange: 'week',
  direction: '',
  extension: '',
  minDuration: '0'
})

const recordings = ref([])

const formatRecording = (r) => {
  const dt = new Date(r.created_at || r.start_time || r.date)
  const durationSec = r.duration || r.billsec || 0
  const mins = Math.floor(durationSec / 60)
  const secs = String(durationSec % 60).padStart(2, '0')
  const sizeKB = r.file_size || 0
  const sizeMB = (sizeKB / (1024 * 1024)).toFixed(1)

  return {
    id: r.id,
    date: dt.toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' }),
    time: dt.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' }),
    direction: r.direction || 'inbound',
    from: r.caller_id_number || r.src || 'Unknown',
    fromName: r.caller_id_name || null,
    to: r.destination_number || r.dst || 'Unknown',
    toName: r.destination_name || null,
    duration: `${mins}:${secs}`,
    size: sizeKB > 0 ? `${sizeMB} MB` : '0 MB',
    rawDate: dt.toISOString(),
  }
}

const loadStats = async () => {
  try {
    const response = await recordingsAPI.list()
    const allRecordings = response.data?.data || response.data?.recordings || []
    totalRecordings.value = allRecordings.length
    
    const today = new Date().toISOString().split('T')[0]
    todayRecordings.value = allRecordings.filter(r => {
      const recDate = new Date(r.created_at || r.start_time || r.date).toISOString().split('T')[0]
      return recDate === today
    }).length
    
    const totalMB = allRecordings.length * 1
    storageUsed.value = totalMB > 1000 ? `${(totalMB / 1000).toFixed(1)} GB` : `${totalMB} MB`
  } catch (err) {
    console.error('Failed to load recording stats:', err)
  }
}

const loadExtensions = async () => {
  try {
    const response = await extensionsAPI.list()
    extensionsList.value = response.data?.data || []
  } catch (err) {
    console.error('Failed to load extensions:', err)
  }
}

const loadRecordings = async () => {
  try {
    const response = await recordingsAPI.list()
    const rawRecordings = response.data?.data || response.data?.recordings || []
    recordings.value = rawRecordings.map(formatRecording)
  } catch (err) {
    console.error('Failed to load recordings:', err)
  }
}

onMounted(() => {
  loadStats()
  loadExtensions()
  loadRecordings()
})

const filteredRecordings = computed(() => {
  let result = recordings.value
  
  if (search.value) {
    const q = search.value.toLowerCase()
    result = result.filter(r =>
      r.from.toLowerCase().includes(q) || r.to.toLowerCase().includes(q) ||
      (r.fromName && r.fromName.toLowerCase().includes(q)) || (r.toName && r.toName.toLowerCase().includes(q))
    )
  }
  
  if (filters.value.extension) {
    result = result.filter(r => r.from?.includes(`Ext ${filters.value.extension}`) || r.to?.includes(`Ext ${filters.value.extension}`))
  }
  
  return result
})

const selectAll = (e) => {
  if (e.target.checked) {
    selectedIds.value = recordings.value.map(r => r.id)
  } else {
    selectedIds.value = []
  }
}

const toggleSelect = (id) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx > -1) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

const selectRecording = (rec) => {
  currentRecording.value = rec
  showPlayer.value = true
}

const playRecording = (rec) => {
  currentRecording.value = rec
  showPlayer.value = true
}

const downloadRecording = async (rec) => {
  if (!rec?.id) {
    console.error('Invalid recording: missing id')
    return
  }

  try {
    const response = await recordingsAPI.download(rec.id)
    const blob = response.data
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `recording-${rec.id}.bin`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Download failed:', err)
    toast.error('Failed to download recording. Please try again.')
  }
}

const exportRecordings = () => {
  const dataToExport = selectedIds.value.length > 0
    ? recordings.value.filter(r => selectedIds.value.includes(r.id))
    : recordings.value

  if (dataToExport.length === 0) {
    toast.error('No recordings to export.')
    return
  }

  const headers = ['Date', 'Time', 'Direction', 'From', 'From Name', 'To', 'To Name', 'Duration', 'Size']
  const rows = dataToExport.map(rec => [
    rec.date,
    rec.time,
    rec.direction,
    rec.from,
    rec.fromName || '',
    rec.to,
    rec.toName || '',
    rec.duration,
    rec.size,
  ])

  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
  ].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `call-recordings-${new Date().toISOString().split('T')[0]}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
</script>

<style scoped>
.call-recordings-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); flex-wrap: wrap; gap: 16px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; align-items: center; }

.search-input { padding: 8px 14px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); min-width: 240px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 24px; }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; text-align: center; }
.stat-value { display: block; font-size: 24px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Filters */
.filters-panel { display: flex; gap: 16px; padding: 16px; background: var(--bg-app); border-radius: var(--radius-md); margin-bottom: 24px; }
.filter-group { display: flex; flex-direction: column; gap: 4px; }
.filter-group label { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; min-width: 140px; }

/* Table */
.recordings-table { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.recordings-table table { width: 100%; border-collapse: collapse; }
.recordings-table th { text-align: left; padding: 12px 16px; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); background: var(--bg-app); border-bottom: 1px solid var(--border-color); }
.recordings-table td { padding: 12px 16px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.recordings-table tr:hover { background: #f8fafc; cursor: pointer; }
.recordings-table tr.selected { background: #eef2ff; }

.datetime-cell { display: flex; flex-direction: column; }
.datetime-cell .date { font-weight: 600; }
.datetime-cell .time { font-size: 11px; color: var(--text-muted); }

.direction-badge { display: inline-flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 4px; text-transform: capitalize; }
.direction-badge.inbound { background: #dcfce7; color: #16a34a; }
.direction-badge.outbound { background: #dbeafe; color: #2563eb; }
.direction-badge.internal { background: #f3e8ff; color: #7c3aed; }

.party-cell { display: flex; flex-direction: column; }
.party-cell .number { font-weight: 500; font-family: monospace; }
.party-cell .name { font-size: 11px; color: var(--text-muted); }

.mono { font-family: monospace; font-size: 12px; }

.actions-cell { display: flex; gap: 4px; }
.action-btn { width: 28px; height: 28px; border-radius: 4px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }

/* Pagination */
.pagination { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; padding: 12px 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }
.page-info { font-size: 12px; color: var(--text-muted); }
.page-controls { display: flex; align-items: center; gap: 4px; }
.page-btn { width: 32px; height: 32px; border: 1px solid var(--border-color); background: white; border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; font-size: 12px; font-weight: 500; }
.page-btn:hover { border-color: var(--primary-color); }
.page-btn.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.page-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.dots { color: var(--text-muted); }

/* Player Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.player-modal { background: white; border-radius: var(--radius-md); width: 100%; max-width: 500px; }
.player-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.player-header h3 { margin: 0; font-size: 16px; }

.player-info { padding: 16px 20px; background: var(--bg-app); }
.call-parties { display: flex; align-items: center; gap: 8px; font-weight: 600; font-size: 14px; margin-bottom: 4px; }
.call-parties .arrow { width: 16px; height: 16px; color: var(--text-muted); }
.call-meta { font-size: 12px; color: var(--text-muted); }

.audio-player { padding: 20px; }
.waveform { height: 60px; background: linear-gradient(90deg, var(--primary-light), var(--primary-color), var(--primary-light)); border-radius: 4px; margin-bottom: 16px; }
.player-controls { display: flex; align-items: center; gap: 16px; }
.play-btn { width: 48px; height: 48px; border-radius: 50%; background: var(--primary-color); color: white; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.time-display { font-family: monospace; font-size: 13px; flex: 1; }
.volume-control { display: flex; align-items: center; gap: 8px; }
.volume-control input { width: 80px; }

.player-actions { display: flex; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 8px 14px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; font-size: 13px; }
.btn-secondary:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-icon { width: 16px; height: 16px; }

.icon-xs { width: 12px; height: 12px; }
.icon-sm { width: 16px; height: 16px; }
.icon-lg { width: 24px; height: 24px; }

@media (max-width: 768px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .filters-panel { flex-wrap: wrap; }
}
</style>
