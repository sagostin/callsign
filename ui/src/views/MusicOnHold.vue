<template>
  <div class="moh-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Music on Hold</h2>
        <p class="text-muted text-sm">Manage hold music by category and sample rate.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showCreateFolderModal = true">
          <FolderPlusIcon class="btn-icon" /> New Folder
        </button>
        <button class="btn-primary" @click="showUploadModal = true">
          <UploadIcon class="btn-icon" /> Upload Music
        </button>
      </div>
    </div>

    <!-- Rate Selector -->
    <div class="rate-selector">
      <label>Sample Rate:</label>
      <div class="rate-btns">
        <button 
          v-for="rate in rates" 
          :key="rate"
          class="rate-btn" 
          :class="{ active: activeRate === rate }"
          @click="activeRate = rate"
        >
          {{ rate }} Hz
        </button>
      </div>
    </div>

    <!-- Folders / Categories -->
    <div class="folders-section">
      <div class="folder-grid">
        <!-- Default / Root files -->
        <div class="folder-card" :class="{ expanded: expandedFolder === '__root__' }" @click="toggleFolder('__root__')">
          <div class="folder-header">
            <div class="folder-info">
              <MusicIcon class="folder-icon" />
              <div>
                <span class="folder-name">Default</span>
                <span class="folder-count">{{ getRootFiles().length }} tracks</span>
              </div>
            </div>
            <ChevronRightIcon class="chevron" :class="{ open: expandedFolder === '__root__' }" />
          </div>
        </div>
        
        <!-- Custom folders/genres -->
        <div 
          v-for="folder in folders" 
          :key="folder.name"
          class="folder-card"
          :class="{ expanded: expandedFolder === folder.name }"
          @click="toggleFolder(folder.name)"
        >
          <div class="folder-header">
            <div class="folder-info">
              <FolderIcon class="folder-icon" />
              <div>
                <span class="folder-name">{{ folder.name }}</span>
                <span class="folder-count">{{ getFilesForFolder(folder.name).length }} tracks</span>
              </div>
            </div>
            <ChevronRightIcon class="chevron" :class="{ open: expandedFolder === folder.name }" />
          </div>
        </div>
      </div>
    </div>

    <!-- Expanded folder contents -->
    <transition name="slide">
      <div v-if="expandedFolder" class="folder-contents">
        <div class="contents-header">
          <h3>{{ expandedFolder === '__root__' ? 'Default' : expandedFolder }}</h3>
          <button class="btn-sm" @click="expandedFolder = null">
            <XIcon class="icon-sm" /> Close
          </button>
        </div>
        <table class="tracks-table">
          <thead>
            <tr>
              <th style="width: 50px;"></th>
              <th>Filename</th>
              <th>Size</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="file in currentFolderFiles" :key="file.path" :class="{ playing: currentlyPlaying === file.path }">
              <td>
                <button class="play-btn" @click.stop="togglePlay(file)">
                  <StopCircleIcon v-if="currentlyPlaying === file.path" />
                  <PlayCircleIcon v-else />
                </button>
              </td>
              <td>
                <span class="filename">{{ file.name }}</span>
                <span v-if="file.isOverride" class="override-badge">Override</span>
              </td>
              <td class="size-cell">{{ formatSize(file.size) }}</td>
              <td>
                <button v-if="isTenant && file.isOverride" class="btn-link text-bad" @click.stop="deleteFile(file)">
                  Revert
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </transition>

    <!-- Audio Player Bar -->
    <transition name="slide-up">
      <div v-if="currentSound" class="audio-player-bar">
        <div class="player-info">
          <div class="player-icon">
            <MusicIcon class="icon" />
          </div>
          <div class="player-details">
            <span class="player-title">{{ currentSound.name }}</span>
            <span class="player-category">{{ activeRate }} Hz</span>
          </div>
        </div>
        <div class="player-controls">
          <button class="ctrl-btn" @click="stopSound" title="Stop">
            <SquareIcon class="ctrl-icon" />
          </button>
          <div class="progress-bar" @click="seekAudio" title="Click to seek">
            <div class="progress-fill" :style="{ width: audioProgress + '%' }"></div>
          </div>
          <span class="time-display">{{ formatTime(audioCurrentTime) }} / {{ formatTime(audioDuration) }}</span>
        </div>
        <button class="close-player" @click="stopSound" title="Close">
          <XIcon />
        </button>
      </div>
    </transition>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Upload Music ({{ activeRate }} Hz)</h3>
          <button class="close-btn" @click="showUploadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Folder/Category</label>
            <select v-model="uploadForm.folder" class="input-field">
              <option value="">Default (Root)</option>
              <option v-for="f in folders" :key="f.name" :value="f.name">{{ f.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>Audio File (WAV/MP3/OGG)</label>
            <div class="file-upload">
              <input type="file" id="music-file" @change="handleFileUpload" accept=".wav,.mp3,.ogg">
              <label for="music-file" class="file-label">
                <UploadIcon class="upload-icon" />
                <span v-if="uploadForm.file">{{ uploadForm.file.name }}</span>
                <span v-else>Choose file or drag & drop</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
          <button class="btn-primary" @click="submitUpload">Upload</button>
        </div>
      </div>
    </div>

    <!-- Create Folder Modal -->
    <div v-if="showCreateFolderModal" class="modal-overlay" @click.self="showCreateFolderModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Create Music Folder</h3>
          <button class="close-btn" @click="showCreateFolderModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Folder Name (Genre/Category)</label>
            <input v-model="newFolderName" class="input-field" placeholder="e.g., jazz, classical, rock">
          </div>
          <p class="text-xs text-muted">Folders help organize music into categories. They'll appear under each sample rate.</p>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showCreateFolderModal = false">Cancel</button>
          <button class="btn-primary" @click="createFolder">Create</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { systemAPI, tenantMediaAPI } from '../services/api'
import { 
  Upload as UploadIcon, Music as MusicIcon, Folder as FolderIcon, FolderPlus as FolderPlusIcon,
  ChevronRight as ChevronRightIcon, X as XIcon, PlayCircle as PlayCircleIcon, StopCircle as StopCircleIcon,
  Square as SquareIcon
} from 'lucide-vue-next'

const activeRate = ref('8000')
const expandedFolder = ref(null)
const musicData = ref([])
const isLoading = ref(false)
const showUploadModal = ref(false)
const showCreateFolderModal = ref(false)
const newFolderName = ref('')
const uploadForm = ref({ file: null, folder: '' })

// Audio player state
const currentlyPlaying = ref(null)
const currentSound = ref(null)
const audioPlayer = ref(null)
const audioProgress = ref(0)
const audioCurrentTime = ref(0)
const audioDuration = ref(0)

const rates = ['8000', '16000', '32000', '48000']
const isTenant = computed(() => !!localStorage.getItem('tenantId'))

// Get rate node from data
const rateNode = computed(() => {
  return musicData.value.find(n => n.name === activeRate.value)
})

// Get folders (subdirectories) within the rate
const folders = computed(() => {
  if (!rateNode.value?.children) return []
  return rateNode.value.children.filter(n => n.type === 'directory')
})

// Get root-level files (not in a subfolder)
const getRootFiles = () => {
  if (!rateNode.value?.children) return []
  return rateNode.value.children.filter(n => n.type === 'file').map(f => ({
    name: f.name,
    path: f.path,
    size: f.size,
    isOverride: f.is_override
  }))
}

// Get files for a specific folder
const getFilesForFolder = (folderName) => {
  const folder = folders.value.find(f => f.name === folderName)
  if (!folder?.children) return []
  return folder.children.filter(n => n.type === 'file').map(f => ({
    name: f.name,
    path: f.path,
    size: f.size,
    isOverride: f.is_override
  }))
}

const currentFolderFiles = computed(() => {
  if (expandedFolder.value === '__root__') return getRootFiles()
  return getFilesForFolder(expandedFolder.value)
})

const toggleFolder = (folderName) => {
  expandedFolder.value = expandedFolder.value === folderName ? null : folderName
}

const formatSize = (bytes) => {
  if (!bytes) return '—'
  return (bytes / 1024).toFixed(1) + ' KB'
}

// Audio player functions
const stopSound = () => {
  if (audioPlayer.value) {
    audioPlayer.value.pause()
    audioPlayer.value.src = ''
    audioPlayer.value = null
  }
  currentlyPlaying.value = null
  currentSound.value = null
  audioProgress.value = 0
  audioCurrentTime.value = 0
  audioDuration.value = 0
}

const togglePlay = (file) => {
  if (currentlyPlaying.value === file.path) {
    stopSound()
    return
  }
  
  stopSound()
  
  const token = localStorage.getItem('token')
  const url = `/api/system/media/music/stream?path=${encodeURIComponent(file.path)}&token=${encodeURIComponent(token)}`
  
  audioPlayer.value = new Audio(url)
  currentSound.value = file
  
  audioPlayer.value.onloadedmetadata = () => {
    audioDuration.value = audioPlayer.value.duration
  }
  
  audioPlayer.value.ontimeupdate = () => {
    if (audioPlayer.value) {
      audioCurrentTime.value = audioPlayer.value.currentTime
      audioDuration.value = audioPlayer.value.duration
      audioProgress.value = (audioPlayer.value.currentTime / audioPlayer.value.duration) * 100
    }
  }
  
  audioPlayer.value.onended = () => stopSound()
  audioPlayer.value.onerror = () => stopSound()
  
  audioPlayer.value.play()
    .then(() => { currentlyPlaying.value = file.path })
    .catch(err => { console.error('Playback failed:', err); stopSound() })
}

const seekAudio = (event) => {
  if (!audioPlayer.value || !audioDuration.value) return
  const rect = event.currentTarget.getBoundingClientRect()
  const percent = (event.clientX - rect.left) / rect.width
  audioPlayer.value.currentTime = percent * audioDuration.value
}

const formatTime = (seconds) => {
  if (!seconds || isNaN(seconds)) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// Data operations
const loadMusic = async () => {
  isLoading.value = true
  try {
    const apiCall = isTenant.value ? tenantMediaAPI.listMusic : systemAPI.listMusic
    const response = await apiCall()
    musicData.value = response.data.data || []
  } catch (e) {
    console.error('Failed to load music', e)
  } finally {
    isLoading.value = false
  }
}

const handleFileUpload = (event) => {
  uploadForm.value.file = event.target.files[0]
}

const submitUpload = async () => {
  if (!uploadForm.value.file) return
  
  const formData = new FormData()
  formData.append('file', uploadForm.value.file)
  formData.append('rate', activeRate.value)
  if (uploadForm.value.folder) {
    formData.append('folder', uploadForm.value.folder)
  }
  
  try {
    if (isTenant.value) {
      await tenantMediaAPI.uploadMusic(formData)
    } else {
      await systemAPI.uploadMusic(formData)
    }
    showUploadModal.value = false
    uploadForm.value = { file: null, folder: '' }
    loadMusic()
  } catch (e) {
    console.error('Upload failed', e)
    alert('Upload failed: ' + (e.message || 'Unknown error'))
  }
}

const createFolder = async () => {
  if (!newFolderName.value.trim()) return
  // TODO: API call to create folder - for now just show alert
  alert(`Folder "${newFolderName.value}" would be created. Upload a file to this folder to auto-create it.`)
  showCreateFolderModal.value = false
  newFolderName.value = ''
}

const deleteFile = async (file) => {
  if (!confirm(`Delete/Revert ${file.name}?`)) return
  try {
    await tenantMediaAPI.deleteMusic(file.path)
    loadMusic()
  } catch (e) {
    console.error('Delete failed', e)
    alert('Failed to delete file')
  }
}

onMounted(() => loadMusic())
onUnmounted(() => stopSound())
</script>

<style scoped>
.moh-page { padding-bottom: 80px; }
.view-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

.rate-selector { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.rate-selector label { font-size: 12px; font-weight: 600; color: var(--text-muted); }
.rate-btns { display: flex; gap: 4px; }
.rate-btn { padding: 8px 16px; border: 1px solid var(--border-color); background: white; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; color: var(--text-muted); }
.rate-btn.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }

.folders-section { margin-bottom: 24px; }
.folder-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px; }
.folder-card { background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; cursor: pointer; transition: all 0.15s; }
.folder-card:hover { border-color: var(--primary-color); }
.folder-card.expanded { border-color: var(--primary-color); background: #f0f9ff; }
.folder-header { display: flex; justify-content: space-between; align-items: center; }
.folder-info { display: flex; align-items: center; gap: 12px; }
.folder-icon { width: 24px; height: 24px; color: var(--primary-color); }
.folder-name { font-size: 14px; font-weight: 600; display: block; }
.folder-count { font-size: 11px; color: var(--text-muted); }
.chevron { width: 16px; height: 16px; color: var(--text-muted); transition: transform 0.2s; }
.chevron.open { transform: rotate(90deg); color: var(--primary-color); }

.folder-contents { background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 20px; }
.contents-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.contents-header h3 { margin: 0; font-size: 16px; }
.btn-sm { display: flex; align-items: center; gap: 4px; padding: 6px 12px; font-size: 12px; background: white; border: 1px solid var(--border-color); border-radius: 4px; cursor: pointer; }
.icon-sm { width: 12px; height: 12px; }

.tracks-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.tracks-table th { text-align: left; padding: 10px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); }
.tracks-table td { padding: 12px 10px; border-bottom: 1px solid var(--border-color); }
.tracks-table tr:hover { background: #f8fafc; }
.tracks-table tr.playing { background: #f0f9ff; }
.filename { font-family: monospace; font-weight: 500; }
.size-cell { color: var(--text-muted); font-size: 12px; }
.play-btn { width: 32px; height: 32px; background: white; border: 1px solid var(--border-color); border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); }
.play-btn:hover { color: var(--primary-color); border-color: var(--primary-color); }
.play-btn svg { width: 18px; height: 18px; }

.override-badge { font-size: 10px; font-weight: 700; color: #d97706; background: #fef3c7; padding: 2px 6px; border-radius: 4px; margin-left: 8px; }
.btn-link { background: none; border: none; color: var(--primary-color); cursor: pointer; font-size: 12px; }
.text-bad { color: #dc2626; }

/* Audio Player Bar */
.audio-player-bar {
  position: fixed; bottom: 0; left: 240px; right: 0; height: 64px;
  background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
  display: flex; align-items: center; padding: 0 24px; gap: 20px; z-index: 100;
  box-shadow: 0 -4px 20px rgba(0,0,0,0.3);
}
.player-info { display: flex; align-items: center; gap: 12px; min-width: 200px; }
.player-icon { width: 40px; height: 40px; background: rgba(255,255,255,0.15); border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.player-icon .icon { width: 20px; height: 20px; color: #fff; }
.player-details { display: flex; flex-direction: column; }
.player-title { color: #fff; font-size: 13px; font-weight: 600; }
.player-category { color: rgba(255,255,255,0.6); font-size: 11px; }
.player-controls { display: flex; align-items: center; gap: 12px; flex: 1; }
.ctrl-btn { width: 36px; height: 36px; background: rgba(255,255,255,0.15); border: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; }
.ctrl-btn:hover { background: rgba(255,255,255,0.25); }
.ctrl-icon { width: 14px; height: 14px; color: #fff; }
.progress-bar { flex: 1; height: 8px; background: rgba(255,255,255,0.2); border-radius: 4px; cursor: pointer; overflow: hidden; }
.progress-bar:hover { background: rgba(255,255,255,0.3); }
.progress-fill { height: 100%; background: linear-gradient(90deg, #3b82f6, #60a5fa); border-radius: 4px; transition: width 0.1s linear; }
.time-display { color: rgba(255,255,255,0.7); font-size: 11px; font-family: monospace; min-width: 80px; }
.close-player { width: 28px; height: 28px; background: transparent; border: none; color: rgba(255,255,255,0.5); cursor: pointer; display: flex; align-items: center; justify-content: center; }
.close-player:hover { color: #fff; }
.close-player svg { width: 18px; height: 18px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 420px; }
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

/* Transitions */
.slide-enter-active, .slide-leave-active { transition: all 0.3s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; transform: translateY(-10px); }
.slide-up-enter-active, .slide-up-leave-active { transition: transform 0.3s ease, opacity 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); opacity: 0; }

@media (max-width: 1024px) { .audio-player-bar { left: 60px; } }
@media (max-width: 768px) { .audio-player-bar { left: 0; } .player-details { display: none; } }
</style>
