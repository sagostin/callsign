<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Audio Library</h2>
      <p class="text-muted text-sm">Manage custom audio files for IVRs, greetings, and system prompts.</p>
    </div>
    <div class="header-actions">
       <button class="btn-primary" @click="$router.push('/admin/audio-library/new')">
         <UploadIcon class="icon-sm" /> Upload Recording
       </button>
    </div>
  </div>

  <!-- Search/Filter Bar -->
  <div class="filter-bar">
    <div class="search-box">
      <SearchIcon class="search-icon" />
      <input type="text" v-model="searchQuery" placeholder="Search by filename..." class="search-input">
    </div>
    <select v-model="filterType" class="filter-select">
      <option value="">All Types</option>
      <option value="audio/wav">WAV</option>
      <option value="audio/mp3">MP3</option>
      <option value="audio/ogg">OGG</option>
    </select>
    <select v-model="filterCategory" class="filter-select">
      <option value="">All Categories</option>
      <option value="greeting">IVR Greeting</option>
      <option value="hold">Hold Music</option>
      <option value="voicemail">Voicemail</option>
      <option value="announcement">Announcement</option>
    </select>
  </div>

  <div class="audio-list">
    <DataTable :columns="columns" :data="filteredRecordings" actions>
      <template #name="{ value, row }">
        <div class="file-info">
          <div class="file-icon" :class="getFileClass(row.type)">
            <FileAudioIcon class="icon-sm" />
          </div>
          <div>
            <span class="file-name">{{ value }}</span>
            <span class="file-category" v-if="row.category">{{ row.category }}</span>
          </div>
        </div>
      </template>
      <template #type="{ value }">
        <span class="type-badge" :class="getTypeClass(value)">{{ formatType(value) }}</span>
      </template>
      <template #size="{ value }">
        <span class="mono-text">{{ value }}</span>
      </template>
      <template #actions="{ row }">
        <div class="action-buttons">
          <button class="btn-icon" title="Play" @click="playAudio(row)">
            <PlayIcon v-if="!isPlaying(row)" class="icon-sm" />
            <PauseIcon v-else class="icon-sm text-primary" />
          </button>
          <button class="btn-icon" title="Download" @click="downloadAudio(row)">
            <DownloadIcon class="icon-sm" />
          </button>
          <button class="btn-icon" title="Edit" @click="editRecording(row)">
            <EditIcon class="icon-sm" />
          </button>
          <button class="btn-icon text-bad" title="Delete" @click="deleteRecording(row)">
            <TrashIcon class="icon-sm" />
          </button>
        </div>
      </template>
    </DataTable>
    
    <!-- Audio Player Bar -->
    <div v-if="currentAudio" class="audio-player-bar">
      <div class="player-info">
        <FileAudioIcon class="icon-sm" />
        <span>{{ currentAudio.name }}</span>
      </div>
      <div class="player-controls">
        <button class="btn-icon" @click="togglePlay">
          <PlayIcon v-if="!playing" class="icon-sm" />
          <PauseIcon v-else class="icon-sm" />
        </button>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: progress + '%' }"></div>
        </div>
        <span class="time-display">{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</span>
        <button class="btn-icon" @click="stopAudio">
          <XIcon class="icon-sm" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import DataTable from '../components/common/DataTable.vue'
import { 
  Search as SearchIcon, 
  Upload as UploadIcon,
  FileAudio as FileAudioIcon,
  Play as PlayIcon,
  Pause as PauseIcon,
  Download as DownloadIcon,
  Edit as EditIcon,
  Trash2 as TrashIcon,
  X as XIcon,
  AlertTriangle as AlertTriangleIcon
} from 'lucide-vue-next'

const activeTab = ref('manage')
const searchQuery = ref('')
const filterType = ref('')
const filterAccess = ref('')
const filterCategory = ref('')

const columns = [
  { key: 'name', label: 'Filename' },
  { key: 'type', label: 'Format', width: '100px' },
  { key: 'access', label: 'Access', width: '140px' },
  { key: 'size', label: 'Size', width: '100px' },
  { key: 'date', label: 'Uploaded', width: '120px' }
]

const recordings = ref([
  { id: 1, name: 'welcome_message.wav', type: 'audio/wav', access: 'Public', size: '1.2 MB', date: '2024-05-10', category: 'IVR Greeting' },
  { id: 2, name: 'holiday_greeting.wav', type: 'audio/wav', access: 'Public', size: '0.8 MB', date: '2024-11-20', category: 'Announcement' },
  { id: 3, name: 'after_hours.mp3', type: 'audio/mp3', access: 'Protected', size: '2.5 MB', date: '2024-01-15', category: 'IVR Greeting' },
  { id: 4, name: 'hold_music_jazz.mp3', type: 'audio/mp3', access: 'Public', size: '8.2 MB', date: '2024-12-01', category: 'Hold Music' },
  { id: 5, name: 'voicemail_unavailable.wav', type: 'audio/wav', access: 'Group', size: '0.5 MB', date: '2025-01-05', category: 'Voicemail' },
])

const filteredRecordings = computed(() => {
  return recordings.value.filter(r => {
    const matchesSearch = r.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesType = !filterType.value || r.type === filterType.value
    const matchesAccess = !filterAccess.value || r.access.includes(filterAccess.value)
    const matchesCategory = !filterCategory.value || r.category?.toLowerCase().includes(filterCategory.value)
    return matchesSearch && matchesType && matchesAccess && matchesCategory
  })
})

// Audio Player State
const currentAudio = ref(null)
const playing = ref(false)
const progress = ref(0)
const currentTime = ref(0)
const duration = ref(0)

const isPlaying = (row) => currentAudio.value?.id === row.id && playing.value

const playAudio = (row) => {
  if (currentAudio.value?.id === row.id) {
    togglePlay()
  } else {
    currentAudio.value = row
    playing.value = true
    duration.value = 45 // Mock duration
    progress.value = 0
    currentTime.value = 0
  }
}

const togglePlay = () => {
  playing.value = !playing.value
}

const stopAudio = () => {
  currentAudio.value = null
  playing.value = false
  progress.value = 0
}

const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const downloadAudio = (row) => {
  alert(`Downloading ${row.name}...`)
}

const editRecording = (row) => {
  alert(`Editing ${row.name}`)
}

const deleteRecording = (row) => {
  if (confirm(`Delete "${row.name}"? This cannot be undone.`)) {
    recordings.value = recordings.value.filter(r => r.id !== row.id)
  }
}

// Helper functions
const formatType = (type) => type?.split('/')[1]?.toUpperCase() || 'AUDIO'

const getFileClass = (type) => {
  if (type.includes('mp3')) return 'mp3'
  if (type.includes('ogg')) return 'ogg'
  return 'wav'
}

const getTypeClass = (type) => {
  if (type.includes('mp3')) return 'mp3'
  if (type.includes('ogg')) return 'ogg'
  return 'wav'
}

const getAccessClass = (access) => {
  if (access.includes('Public')) return 'public'
  if (access.includes('Protected')) return 'protected'
  return 'group'
}

// Settings
const settings = ref({
  groupPermissions: false,
  recordAll: false,
  onDemand: true,
  featureCode: '*1',
  retentionDays: 90
})
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
.btn-icon.text-bad:hover { background: #fee2e2; color: var(--status-bad); }
.text-primary { color: var(--primary-color); }
.text-bad { color: var(--status-bad); }

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
  min-width: 140px;
}

/* File Info Cell */
.file-info {
  display: flex;
  align-items: center;
  gap: 10px;
}
.file-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #e0e7ff;
  color: #4f46e5;
}
.file-icon.mp3 { background: #dcfce7; color: #16a34a; }
.file-icon.ogg { background: #fef3c7; color: #d97706; }
.file-name { font-weight: 500; display: block; }
.file-category { font-size: 11px; color: var(--text-muted); }

/* Badges */
.type-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  background: #e0e7ff;
  color: #4f46e5;
}
.type-badge.mp3 { background: #dcfce7; color: #16a34a; }
.type-badge.ogg { background: #fef3c7; color: #d97706; }

.access-badge {
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 500;
  background: #f1f5f9;
  color: #64748b;
}
.access-badge.public { background: #dcfce7; color: #16a34a; }
.access-badge.protected { background: #fef3c7; color: #d97706; }
.access-badge.group { background: #e0e7ff; color: #4f46e5; }

.mono-text { font-family: monospace; font-size: 12px; color: var(--text-muted); }

/* Action Buttons */
.action-buttons {
  display: flex;
  gap: 4px;
}

/* Audio Player Bar */
.audio-player-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #4f46e5, #6366f1);
  border-radius: 0 0 var(--radius-md) var(--radius-md);
  color: white;
}
.player-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
}
.player-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}
.player-controls .btn-icon { color: white; }
.player-controls .btn-icon:hover { background: rgba(255,255,255,0.2); }
.progress-bar {
  width: 200px;
  height: 4px;
  background: rgba(255,255,255,0.3);
  border-radius: 2px;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: white;
  transition: width 0.3s;
}
.time-display {
  font-size: 11px;
  font-family: monospace;
  opacity: 0.9;
}

/* Settings Panel */
.settings-panel {
  padding: var(--spacing-xl);
}
.settings-grid {
  display: grid;
  gap: 24px;
  max-width: 700px;
}
.settings-section {
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
}
.settings-section h3 {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 4px;
}
.help-text {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 16px;
}
.setting-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}
.setting-row:last-child { border-bottom: none; }
.setting-info { flex: 1; }
.setting-info label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
  text-transform: none;
  display: block;
}
.setting-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

/* Toggle Switch */
.toggle-switch {
  position: relative;
  width: 44px;
  height: 24px;
  flex-shrink: 0;
}
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background: #cbd5e1;
  border-radius: 24px;
  transition: 0.3s;
}
.toggle-slider::before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: 0.3s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-switch input:checked + .toggle-slider { background: var(--primary-color); }
.toggle-switch input:checked + .toggle-slider::before { transform: translateX(20px); }

.sub-setting {
  padding: 12px 0 0 0;
  margin-top: 8px;
  border-top: 1px dashed var(--border-color);
}
.mini-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 6px;
  display: block;
}
.input-field {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 14px;
  width: 100%;
  max-width: 300px;
}
.input-field.small { max-width: 100px; }

.form-group {
  margin-top: 12px;
}

.retention-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 10px 12px;
  background: #fef3c7;
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: #92400e;
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}

.icon-sm { width: 16px; height: 16px; }
</style>
