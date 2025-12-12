<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Audio Library</h2>
      <p class="text-muted text-sm">Manage custom audio files for IVRs, greetings, and system prompts.</p>
    </div>
    <div class="header-actions">
       <button class="btn-primary" @click="showUploadModal = true">
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

  <!-- Upload Modal -->
  <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>Upload Recording</h3>
        <button class="close-btn" @click="showUploadModal = false">×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>Name</label>
          <input v-model="uploadForm.name" class="input-field" placeholder="E.g., Welcome Greeting">
        </div>
        <div class="form-group">
          <label>Type / Category</label>
          <div style="display:flex; gap:10px;">
            <select v-model="uploadForm.type" class="input-field">
              <option value="custom">Custom</option>
              <option value="greeting">Greeting</option>
              <option value="music">Music</option>
            </select>
            <select v-model="uploadForm.category" class="input-field">
              <option value="">Select Category...</option>
              <option value="IVR">IVR</option>
              <option value="Voicemail">Voicemail</option>
              <option value="Hold Music">Hold Music</option>
              <option value="Announcement">Announcement</option>
            </select>
          </div>
        </div>
        <div class="form-group">
          <label>Description</label>
          <input v-model="uploadForm.description" class="input-field" placeholder="Optional description">
        </div>
        <div class="form-group">
          <label>File</label>
          <div class="file-upload" @click="$refs.fileInput.click()">
            <input ref="fileInput" type="file" accept=".wav,.mp3,.ogg" @change="handleFileUpload">
            <div class="file-label" v-if="!uploadForm.file">
              <UploadIcon class="upload-icon" />
              <span>Click to select audio file</span>
              <span class="text-xs text-muted">WAV, MP3, OGG supported</span>
            </div>
            <div v-else class="file-label">
              <FileAudioIcon class="upload-icon" />
              <span class="text-primary">{{ uploadForm.file.name }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
        <button class="btn-primary" @click="submitUpload" :disabled="isUploading">
          {{ isUploading ? 'Uploading...' : 'Upload' }}
        </button>
      </div>
    </div>
  </div>

  <!-- Edit Modal -->
  <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>Edit Recording</h3>
        <button class="close-btn" @click="showEditModal = false">×</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>Name</label>
          <input v-model="editForm.name" class="input-field">
        </div>
        <div class="form-group">
          <label>Category</label>
          <input v-model="editForm.category" class="input-field">
        </div>
        <div class="form-group">
          <label>Description</label>
          <input v-model="editForm.description" class="input-field">
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-secondary" @click="showEditModal = false">Cancel</button>
        <button class="btn-primary" @click="submitEdit">Save Changes</button>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
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
import { audioLibraryAPI } from '../services/api'

const activeTab = ref('manage')
const searchQuery = ref('')
const filterType = ref('')
const filterCategory = ref('')
const isLoading = ref(false)
const showUploadModal = ref(false)
const showEditModal = ref(false)
const isUploading = ref(false)

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'type', label: 'Type', width: '100px' },
  { key: 'category', label: 'Category', width: '120px' },
  { key: 'filename', label: 'Filename', width: '180px' },
  { key: 'size', label: 'Size', width: '100px' },
  { key: 'created_at', label: 'Uploaded', width: '120px' }
]

const recordings = ref([])

const uploadForm = ref({
  name: '',
  description: '',
  type: 'custom',
  category: '',
  file: null
})

const editForm = ref({
  id: null,
  name: '',
  description: '',
  type: 'custom',
  category: ''
})

const loadRecordings = async () => {
  isLoading.value = true
  try {
    const response = await audioLibraryAPI.list()
    recordings.value = response.data.data || []
  } catch (e) {
    console.error('Failed to load recordings', e)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadRecordings()
})

const filteredRecordings = computed(() => {
  return recordings.value.filter(r => {
    const matchesSearch = r.name.toLowerCase().includes(searchQuery.value.toLowerCase()) || 
                          r.filename.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesType = !filterType.value || r.type === filterType.value
    const matchesCategory = !filterCategory.value || r.category?.toLowerCase().includes(filterCategory.value.toLowerCase())
    return matchesSearch && matchesType && matchesCategory
  })
})

const handleFileUpload = (event) => {
  uploadForm.value.file = event.target.files[0]
}

const submitUpload = async () => {
  if (!uploadForm.value.file || !uploadForm.value.name) return
  
  isUploading.value = true
  const formData = new FormData()
  formData.append('file', uploadForm.value.file)
  formData.append('name', uploadForm.value.name)
  formData.append('description', uploadForm.value.description)
  formData.append('type', uploadForm.value.type)
  formData.append('category', uploadForm.value.category)

  try {
    await audioLibraryAPI.upload(formData)
    showUploadModal.value = false
    // Reset form
    uploadForm.value = { name: '', description: '', type: 'custom', category: '', file: null }
    loadRecordings()
  } catch (e) {
    console.error('Upload failed', e)
    alert('Upload failed: ' + (e.response?.data?.error || e.message))
  } finally {
    isUploading.value = false
  }
}

const editRecording = (row) => {
  editForm.value = { 
    id: row.id, 
    name: row.name, 
    description: row.description, 
    type: row.type, 
    category: row.category 
  }
  showEditModal.value = true
}

const submitEdit = async () => {
  try {
    await audioLibraryAPI.update(editForm.value.id, editForm.value)
    showEditModal.value = false
    loadRecordings()
  } catch (e) {
    console.error('Update failed', e)
    alert('Update failed')
  }
}

const deleteRecording = async (row) => {
  if (!confirm(`Delete "${row.name}"? This cannot be undone.`)) return
  
  try {
    await audioLibraryAPI.delete(row.id)
    loadRecordings()
  } catch (e) {
    console.error('Delete failed', e)
    alert('Failed to delete recording')
  }
}

// Audio Player State
const currentAudio = ref(null)
const playing = ref(false)
const audioObj = ref(null)
const progress = ref(0)
const currentTime = ref(0)
const duration = ref(0)

const isPlaying = (row) => currentAudio.value?.id === row.id && playing.value

const playAudio = (row) => {
  if (currentAudio.value?.id === row.id) {
    togglePlay()
  } else {
    stopAudio() // Stop previous
    currentAudio.value = row
    
    // Use the stream endpoint
    const url = `/api/audio-library/${row.id}/stream`
    
    audioObj.value = new Audio(url)
    audioObj.value.addEventListener('timeupdate', () => {
      currentTime.value = audioObj.value.currentTime
      duration.value = audioObj.value.duration || 0
      progress.value = duration.value ? (currentTime.value / duration.value) * 100 : 0
    })
    audioObj.value.addEventListener('ended', () => {
      playing.value = false
      progress.value = 0
    })
    audioObj.value.play()
    playing.value = true
  }
}

const togglePlay = () => {
  if (!audioObj.value) return
  if (playing.value) {
    audioObj.value.pause()
  } else {
    audioObj.value.play()
  }
  playing.value = !playing.value
}

const stopAudio = () => {
  if (audioObj.value) {
    audioObj.value.pause()
    audioObj.value.currentTime = 0
    audioObj.value = null
  }
  currentAudio.value = null
  playing.value = false
  progress.value = 0
  currentTime.value = 0
  duration.value = 0
}

const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const downloadAudio = (row) => {
  // Create a download link using the stream endpoint
  const url = `/api/audio-library/${row.id}/stream`
  const link = document.createElement('a')
  link.href = url
  link.download = row.filename || `${row.name}.wav`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// Helper functions
const formatType = (type) => type?.toUpperCase() || 'AUDIO'

const getFileClass = (type) => 'wav'
const getTypeClass = (type) => 'wav'

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString()
}
const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  if (bytes < k) return bytes + ' B'
  const sizes = ['KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i-1]
}

const getAccessClass = () => 'group'
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

.mono-text { font-family: monospace; font-size: 12px; color: var(--text-muted); }
.action-buttons { display: flex; gap: 4px; }

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

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.file-upload { border: 2px dashed var(--border-color); border-radius: 8px; padding: 24px; text-align: center; }
.file-upload input { display: none; }
.file-label { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.upload-icon { width: 24px; height: 24px; }
.icon-sm { width: 16px; height: 16px; }
</style>
