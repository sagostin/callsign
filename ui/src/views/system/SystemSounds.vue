<template>
  <div class="sounds-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Sounds</h2>
        <p class="text-muted text-sm">Manage internal system sounds, tones, and prompts across all tenants.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showImportModal = true">
          <DownloadIcon class="btn-icon" /> Import Pack
        </button>
        <button class="btn-primary" @click="showUploadModal = true">
          <UploadIcon class="btn-icon" /> Upload Sound
        </button>
      </div>
    </div>

    <!-- File Browser Layout -->
    <div class="file-browser">
      <!-- Folder Tree Sidebar -->
      <aside class="folder-sidebar">
        <div class="sidebar-header">
          <FolderIcon class="header-icon" />
          <span>Sound Folders</span>
        </div>
        <div class="folder-tree">
          <div v-if="isLoading" class="loading-state">
            <div class="loader"></div>
            <span>Loading...</span>
          </div>
          <template v-else>
            <FolderNode 
              v-for="node in rawData" 
              :key="node.path" 
              :node="node"
              :selected-path="selectedFolderPath"
              :expanded-paths="expandedPaths"
              @select="selectFolder"
              @toggle="toggleFolder"
            />
          </template>
        </div>
      </aside>

      <!-- File List Area -->
      <main class="file-list-area">
        <!-- Breadcrumb -->
        <div class="breadcrumb-bar">
          <div class="breadcrumb">
            <button class="crumb" @click="selectFolder('')">
              <HomeIcon class="crumb-icon" />
            </button>
            <template v-for="(part, idx) in breadcrumbParts" :key="idx">
              <ChevronRightIcon class="crumb-separator" />
              <button class="crumb" @click="navigateToBreadcrumb(idx)">
                {{ part }}
              </button>
            </template>
          </div>
          <div class="view-toggle">
            <button 
              :class="{ active: viewMode === 'list' }" 
              @click="viewMode = 'list'"
              title="List View"
            >
              <ListIcon />
            </button>
            <button 
              :class="{ active: viewMode === 'grid' }" 
              @click="viewMode = 'grid'"
              title="Grid View"
            >
              <GridIcon />
            </button>
          </div>
        </div>

        <!-- Files Content -->
        <div v-if="currentFolderContents.length === 0" class="empty-state">
          <FolderOpenIcon class="empty-icon" />
          <p>{{ selectedFolderPath ? 'This folder is empty' : 'Select a folder to view sounds' }}</p>
        </div>

        <!-- List View -->
        <div v-else-if="viewMode === 'list'" class="files-list">
          <table class="files-table">
            <thead>
              <tr>
                <th class="col-name">Name</th>
                <th class="col-type">Type</th>
                <th class="col-size">Size</th>
                <th class="col-actions">Actions</th>
              </tr>
            </thead>
            <tbody>
              <!-- Folders first -->
              <tr 
                v-for="item in currentFolderContents.filter(i => i.type === 'directory')" 
                :key="item.path"
                class="folder-row"
                @dblclick="selectFolder(item.path)"
              >
                <td class="col-name">
                  <FolderIcon class="item-icon folder" />
                  <span class="item-name">{{ item.name }}</span>
                </td>
                <td class="col-type">Folder</td>
                <td class="col-size">{{ item.children?.length || 0 }} items</td>
                <td class="col-actions">
                  <button class="btn-icon-sm" @click="selectFolder(item.path)" title="Open">
                    <ChevronRightIcon />
                  </button>
                </td>
              </tr>
              <!-- Files -->
              <tr 
                v-for="item in currentFolderContents.filter(i => i.type === 'file')" 
                :key="item.path"
                class="file-row"
                :class="{ playing: currentlyPlaying === item.path }"
              >
                <td class="col-name">
                  <FileAudioIcon class="item-icon audio" />
                  <span class="item-name">{{ item.name }}</span>
                  <span v-if="item.is_override" class="override-badge">Override</span>
                </td>
                <td class="col-type">{{ getFileType(item.name) }}</td>
                <td class="col-size">{{ formatFileSize(item.size) }}</td>
                <td class="col-actions">
                  <button 
                    class="btn-icon-sm play-btn" 
                    :class="{ 'is-playing': currentlyPlaying === item.path }"
                    @click="togglePlaySound(item)" 
                    :title="currentlyPlaying === item.path ? 'Stop' : 'Play'"
                  >
                    <StopCircleIcon v-if="currentlyPlaying === item.path" />
                    <PlayCircleIcon v-else />
                  </button>
                  <button v-if="item.is_override" class="btn-icon-sm" @click="deleteOverride(item)" title="Revert">
                    <XIcon class="text-bad" />
                  </button>
                  <button v-else class="btn-icon-sm" @click="uploadForPath(item)" :title="isTenant ? 'Override' : 'Replace'">
                    <UploadIcon />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Grid View -->
        <div v-else class="files-grid">
          <!-- Folders -->
          <div 
            v-for="item in currentFolderContents.filter(i => i.type === 'directory')" 
            :key="item.path"
            class="grid-item folder"
            @dblclick="selectFolder(item.path)"
          >
            <div class="grid-icon">
              <FolderIcon />
            </div>
            <span class="grid-name">{{ item.name }}</span>
            <span class="grid-meta">{{ item.children?.length || 0 }} items</span>
          </div>
          <!-- Files -->
          <div 
            v-for="item in currentFolderContents.filter(i => i.type === 'file')" 
            :key="item.path"
            class="grid-item file"
            :class="{ playing: currentlyPlaying === item.path }"
            @click="togglePlaySound(item)"
          >
            <div class="grid-icon" :class="{ 'is-playing': currentlyPlaying === item.path }">
              <StopCircleIcon v-if="currentlyPlaying === item.path" />
              <FileAudioIcon v-else />
            </div>
            <span class="grid-name">{{ item.name }}</span>
            <span class="grid-meta">{{ getFileType(item.name) }}</span>
            <span v-if="item.is_override" class="override-badge">Override</span>
          </div>
        </div>

        <!-- Stats Bar -->
        <div class="stats-bar">
          <span>{{ folderStats.folders }} folders • {{ folderStats.files }} files</span>
        </div>
      </main>
    </div>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Upload System Sound</h3>
          <button class="close-btn" @click="showUploadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Target Folder</label>
            <input v-model="uploadForm.path" class="input-field" :placeholder="selectedFolderPath || 'Select a folder first'">
          </div>
          <div class="form-group">
            <label>Source</label>
            <div class="tabs-small">
               <button :class="{ active: uploadSource === 'file' }" @click="uploadSource = 'file'">Upload File</button>
               <button :class="{ active: uploadSource === 'record' }" @click="uploadSource = 'record'">Record Audio</button>
            </div>
          </div>
          
          <div v-if="uploadSource === 'file'" class="form-group">
            <label>Audio File</label>
            <div class="file-upload-zone">
              <input type="file" id="sound-file" accept=".wav,.mp3,.ogg" @change="handleFileUpload">
              <label for="sound-file" class="file-label">
                <UploadIcon class="upload-icon" />
                <span v-if="uploadForm.file">{{ uploadForm.file.name }}</span>
                <span v-else>Choose file or drag & drop</span>
              </label>
            </div>
          </div>
          
          <div v-else class="form-group">
             <label>Record New Sound</label>
           <AudioRecorder @record-complete="handleRecordingComplete" />
         </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
          <button class="btn-primary" @click="submitUpload">Upload</button>
        </div>
      </div>
    </div>

    <!-- Persistent Audio Player Bar -->
    <transition name="slide-up">
      <div v-if="currentSound" class="audio-player-bar">
        <div class="player-info">
          <div class="player-icon" :class="{ 'is-playing': !isPaused }">
            <Volume2 class="icon" />
          </div>
          <div class="player-details">
            <span class="player-title">{{ currentSound.name }}</span>
            <span class="player-path">{{ currentSound.path }}</span>
          </div>
        </div>
        
        <div class="player-controls">
          <button class="ctrl-btn" @click="togglePause" :title="isPaused ? 'Play' : 'Pause'">
            <PlayIcon v-if="isPaused" class="ctrl-icon" />
            <PauseIcon v-else class="ctrl-icon" />
          </button>
          <button class="ctrl-btn" @click="stopSound" title="Stop">
            <SquareIcon class="ctrl-icon" />
          </button>
          
          <div class="progress-container">
            <span class="time-display time-current">{{ formatTime(audioCurrentTime) }}</span>
            <div 
              class="progress-bar" 
              @mousedown="startScrub"
              @mousemove="scrubbing && scrubMove($event)"
              @mouseup="endScrub"
              @mouseleave="scrubbing && endScrub()"
              @click="seekAudio"
              title="Click to seek or drag to scrub"
            >
              <div class="progress-fill" :style="{ width: audioProgress + '%' }"></div>
              <div class="progress-thumb" :style="{ left: audioProgress + '%' }"></div>
            </div>
            <span class="time-display time-duration">{{ formatTime(audioDuration) }}</span>
          </div>
          
          <div class="volume-control">
            <button class="ctrl-btn" @click="toggleMute" :title="isMuted ? 'Unmute' : 'Mute'">
              <VolumeXIcon v-if="isMuted" class="ctrl-icon" />
              <Volume2Icon v-else class="ctrl-icon" />
            </button>
            <input 
              type="range" 
              class="volume-slider" 
              min="0" 
              max="100" 
              v-model="audioVolume"
              @input="updateVolume"
            />
          </div>
        </div>
        
        <button class="close-player" @click="stopSound" title="Close">
          <XIcon />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, defineComponent, h } from 'vue'
import {
  Upload as UploadIcon, Download as DownloadIcon, ChevronDown as ChevronDownIcon, ChevronRight as ChevronRightIcon,
  PlayCircle as PlayCircleIcon, StopCircle as StopCircleIcon, Check as CheckIcon, X as XIcon,
  Volume2, Square as SquareIcon, Play as PlayIcon, Pause as PauseIcon, VolumeX as VolumeXIcon, Volume2 as Volume2Icon,
  Folder as FolderIcon, FolderOpen as FolderOpenIcon, FileAudio as FileAudioIcon, Home as HomeIcon,
  List as ListIcon, LayoutGrid as GridIcon
} from 'lucide-vue-next'
import { systemAPI, tenantMediaAPI } from '../../services/api'
import AudioRecorder from '../../components/common/AudioRecorder.vue'

// Recursive FolderNode component
const FolderNode = defineComponent({
  name: 'FolderNode',
  props: ['node', 'selectedPath', 'expandedPaths', 'depth'],
  emits: ['select', 'toggle'],
  setup(props, { emit }) {
    const isExpanded = computed(() => props.expandedPaths.includes(props.node.path))
    const isSelected = computed(() => props.selectedPath === props.node.path)
    const hasChildren = computed(() => props.node.children?.some(c => c.type === 'directory'))
    const depth = props.depth || 0
    
    return () => {
      if (props.node.type !== 'directory') return null
      
      const children = []
      
      // Folder item
      children.push(
        h('div', {
          class: ['folder-item', { selected: isSelected.value }],
          style: { paddingLeft: `${12 + depth * 16}px` },
          onClick: () => emit('select', props.node.path)
        }, [
          hasChildren.value
            ? h('button', {
                class: ['expand-btn', { expanded: isExpanded.value }],
                onClick: (e) => { e.stopPropagation(); emit('toggle', props.node.path) }
              }, [h(ChevronRightIcon, { class: 'expand-icon' })])
            : h('span', { class: 'expand-spacer' }),
          h(isExpanded.value ? FolderOpenIcon : FolderIcon, { class: 'folder-icon' }),
          h('span', { class: 'folder-name' }, props.node.name)
        ])
      )
      
      // Children (if expanded)
      if (isExpanded.value && props.node.children) {
        const folderChildren = props.node.children.filter(c => c.type === 'directory')
        folderChildren.forEach(child => {
          children.push(
            h(FolderNode, {
              node: child,
              selectedPath: props.selectedPath,
              expandedPaths: props.expandedPaths,
              depth: depth + 1,
              onSelect: (path) => emit('select', path),
              onToggle: (path) => emit('toggle', path)
            })
          )
        })
      }
      
      return h('div', { class: 'folder-node' }, children)
    }
  }
})

const isLoading = ref(false)
const rawData = ref([])
const selectedFolderPath = ref('')
const expandedPaths = ref([])
const viewMode = ref('list')

const showUploadModal = ref(false)
const showImportModal = ref(false)
const uploadForm = ref({ path: '', file: null, blob: null })
const uploadSource = ref('file')

// Audio player state
const currentlyPlaying = ref(null)
const currentSound = ref(null)
const audioPlayer = ref(null)
const audioProgress = ref(0)
const audioCurrentTime = ref(0)
const audioDuration = ref(0)
const isPaused = ref(false)
const audioVolume = ref(80)
const isMuted = ref(false)
const scrubbing = ref(false)

const isTenant = computed(() => !!localStorage.getItem('tenantId'))

// Get current folder contents
const currentFolderContents = computed(() => {
  if (!selectedFolderPath.value || !rawData.value.length) {
    // Show top-level items
    return rawData.value
  }
  
  const parts = selectedFolderPath.value.split('/')
  let current = rawData.value
  
  for (const part of parts) {
    const found = current.find(n => n.name === part)
    if (!found || !found.children) return []
    current = found.children
  }
  
  return current
})

// Breadcrumb parts
const breadcrumbParts = computed(() => {
  if (!selectedFolderPath.value) return []
  return selectedFolderPath.value.split('/')
})

// Folder stats
const folderStats = computed(() => {
  const items = currentFolderContents.value
  return {
    folders: items.filter(i => i.type === 'directory').length,
    files: items.filter(i => i.type === 'file').length
  }
})

// Folder navigation
const selectFolder = (path) => {
  selectedFolderPath.value = path
  uploadForm.value.path = path
  
  // Auto-expand parent paths
  if (path) {
    const parts = path.split('/')
    let currentPath = ''
    for (const part of parts) {
      currentPath = currentPath ? `${currentPath}/${part}` : part
      if (!expandedPaths.value.includes(currentPath)) {
        expandedPaths.value.push(currentPath)
      }
    }
  }
}

const toggleFolder = (path) => {
  const idx = expandedPaths.value.indexOf(path)
  if (idx === -1) {
    expandedPaths.value.push(path)
  } else {
    expandedPaths.value.splice(idx, 1)
  }
}

const navigateToBreadcrumb = (idx) => {
  const parts = breadcrumbParts.value.slice(0, idx + 1)
  selectFolder(parts.join('/'))
}

// File utilities
const getFileType = (filename) => {
  const ext = filename.split('.').pop().toLowerCase()
  const types = { wav: 'WAV Audio', mp3: 'MP3 Audio', ogg: 'OGG Audio', gsm: 'GSM Audio' }
  return types[ext] || 'Audio File'
}

const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
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
  isPaused.value = false
  scrubbing.value = false
}

const togglePlaySound = (item) => {
  if (item.type === 'directory') return
  
  if (currentlyPlaying.value === item.path) {
    stopSound()
    return
  }
  
  stopSound()
  
  const token = localStorage.getItem('token')
  const baseUrl = '/api/system/media/sounds/stream'
  const url = `${baseUrl}?path=${encodeURIComponent(item.path)}&token=${encodeURIComponent(token)}`
  
  audioPlayer.value = new Audio(url)
  currentSound.value = item
  
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
  
  audioPlayer.value.onended = () => {
    stopSound()
  }
  
  audioPlayer.value.onerror = (e) => {
    console.error('Audio playback error:', e)
    stopSound()
  }
  
  audioPlayer.value.play()
    .then(() => {
      currentlyPlaying.value = item.path
    })
    .catch(err => {
      console.error('Failed to play:', err)
      stopSound()
    })
}

const seekAudio = (event) => {
  if (!audioPlayer.value || !audioDuration.value) return
  const progressBar = event.currentTarget
  const rect = progressBar.getBoundingClientRect()
  const clickX = event.clientX - rect.left
  const percent = clickX / rect.width
  audioPlayer.value.currentTime = percent * audioDuration.value
}

const togglePause = () => {
  if (!audioPlayer.value) return
  if (isPaused.value) {
    audioPlayer.value.play()
    isPaused.value = false
  } else {
    audioPlayer.value.pause()
    isPaused.value = true
  }
}

const toggleMute = () => {
  if (!audioPlayer.value) return
  isMuted.value = !isMuted.value
  audioPlayer.value.muted = isMuted.value
}

const updateVolume = () => {
  if (!audioPlayer.value) return
  audioPlayer.value.volume = audioVolume.value / 100
  if (audioVolume.value === 0) {
    isMuted.value = true
  } else if (isMuted.value && audioVolume.value > 0) {
    isMuted.value = false
    audioPlayer.value.muted = false
  }
}

const startScrub = (event) => {
  scrubbing.value = true
  seekAudio(event)
}

const scrubMove = (event) => {
  if (scrubbing.value) {
    seekAudio(event)
  }
}

const endScrub = () => {
  scrubbing.value = false
}

const formatTime = (seconds) => {
  if (!seconds || isNaN(seconds)) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// Upload functions
const uploadForPath = (item) => {
  // Set path to parent folder
  const parts = item.path.split('/')
  parts.pop()
  uploadForm.value.path = parts.join('/')
  uploadForm.value.file = null
  uploadForm.value.blob = null
  uploadSource.value = 'file'
  showUploadModal.value = true
}

const handleFileUpload = (event) => {
  uploadForm.value.file = event.target.files[0]
}

const handleRecordingComplete = (blob) => {
  uploadForm.value.blob = blob
}

const submitUpload = async () => {
  let fileToUpload = null
  
  if (uploadSource.value === 'file') {
    fileToUpload = uploadForm.value.file
  } else {
    if (!uploadForm.value.blob) return
    fileToUpload = new File([uploadForm.value.blob], 'recording.wav', { type: 'audio/wav' })
  }
  
  if (!fileToUpload) return
  
  const formData = new FormData()
  formData.append('file', fileToUpload)
  formData.append('path', uploadForm.value.path)
  
  try {
    if (isTenant.value) {
      await tenantMediaAPI.uploadSound(formData)
    } else {
      await systemAPI.uploadSound(formData)
    }
    showUploadModal.value = false
    loadSounds()
  } catch (e) {
    console.error('Upload failed', e)
    alert('Upload failed: ' + (e.response?.data?.error || e.message))
  }
}

const deleteOverride = async (item) => {
  if (!confirm(`Revert override for ${item.name}?`)) return
  
  try {
    await tenantMediaAPI.deleteSound(item.path)
    loadSounds()
  } catch (e) {
    console.error('Delete failed', e)
    alert('Failed to delete override')
  }
}

const loadSounds = async () => {
  isLoading.value = true
  try {
    const apiCall = isTenant.value ? tenantMediaAPI.listSounds : systemAPI.listSounds
    const response = await apiCall()
    rawData.value = response.data.data
  } catch (e) {
    console.error('Failed to load sounds', e)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  loadSounds()
})

onUnmounted(() => {
  stopSound()
})
</script>

<style scoped>
.sounds-page { padding: 0; display: flex; flex-direction: column; height: 100%; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; flex-shrink: 0; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

/* File Browser Layout */
.file-browser {
  display: flex;
  flex: 1;
  gap: 0;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
  min-height: 0;
}

/* Folder Sidebar */
.folder-sidebar {
  width: 280px;
  min-width: 280px;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  background: #f8fafc;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
}

.sidebar-header .header-icon { width: 16px; height: 16px; }

.folder-tree {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 20px;
  color: var(--text-muted);
  font-size: 13px;
}

.loader {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

/* Folder Node Styles */
:deep(.folder-node) { display: contents; }
:deep(.folder-item) {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-main);
  transition: background 0.15s;
}

:deep(.folder-item:hover) { background: #e2e8f0; }
:deep(.folder-item.selected) { background: var(--primary-light); color: var(--primary-color); font-weight: 600; }

:deep(.expand-btn) {
  width: 18px; height: 18px;
  padding: 0; border: none; background: none;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; color: var(--text-muted);
  transition: transform 0.15s;
}
:deep(.expand-btn.expanded) { transform: rotate(90deg); }
:deep(.expand-icon) { width: 14px; height: 14px; }
:deep(.expand-spacer) { width: 18px; }
:deep(.folder-icon) { width: 18px; height: 18px; color: #f59e0b; }
:deep(.folder-name) { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* File List Area */
.file-list-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.breadcrumb-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-color);
  background: #fafafa;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
}

.crumb {
  padding: 4px 8px;
  border: none;
  background: none;
  font-size: 13px;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.crumb:hover { background: var(--border-color); color: var(--text-primary); }
.crumb-icon { width: 16px; height: 16px; }
.crumb-separator { width: 14px; height: 14px; color: var(--text-muted); }

.view-toggle {
  display: flex;
  gap: 2px;
  background: var(--border-color);
  border-radius: 6px;
  padding: 2px;
}

.view-toggle button {
  padding: 6px;
  border: none;
  background: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  color: var(--text-muted);
}

.view-toggle button.active { background: white; color: var(--text-primary); box-shadow: 0 1px 2px rgba(0,0,0,0.1); }
.view-toggle button svg { width: 16px; height: 16px; }

/* Empty State */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  gap: 12px;
}

.empty-icon { width: 48px; height: 48px; opacity: 0.4; }

/* Files Table */
.files-list {
  flex: 1;
  overflow-y: auto;
}

.files-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.files-table th {
  text-align: left;
  padding: 10px 16px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
  background: #f8fafc;
  position: sticky;
  top: 0;
}

.files-table td {
  padding: 10px 16px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}

.files-table tr:hover { background: #f8fafc; }
.files-table tr.folder-row { cursor: pointer; }
.files-table tr.playing { background: var(--primary-light); }

.col-name { min-width: 200px; }
.col-type { width: 120px; color: var(--text-muted); }
.col-size { width: 100px; color: var(--text-muted); }
.col-actions { width: 120px; }

.col-name { display: flex; align-items: center; gap: 10px; }
.col-actions { display: flex; gap: 4px; }

.item-icon { width: 20px; height: 20px; flex-shrink: 0; }
.item-icon.folder { color: #f59e0b; }
.item-icon.audio { color: var(--primary-color); }
.item-name { font-weight: 500; }

.btn-icon-sm {
  width: 28px; height: 28px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-muted);
  transition: all 0.15s ease;
}

.btn-icon-sm:hover { color: var(--primary-color); border-color: var(--primary-color); }
.btn-icon-sm svg { width: 16px; height: 16px; }
.btn-icon-sm.play-btn.is-playing { background: var(--primary-color); border-color: var(--primary-color); color: white; }

.override-badge { 
  font-size: 10px; font-weight: 700; color: #d97706; background: #fef3c7; 
  padding: 2px 6px; border-radius: 4px; border: 1px solid #fcd34d;
  text-transform: uppercase; margin-left: 8px;
}

/* Files Grid */
.files-grid {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 12px;
  align-content: start;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 12px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s;
  position: relative;
  text-align: center;
}

.grid-item:hover { border-color: var(--primary-color); background: var(--primary-light); }
.grid-item.playing { border-color: var(--primary-color); background: var(--primary-light); }

.grid-icon {
  width: 48px; height: 48px;
  display: flex; align-items: center; justify-content: center;
  background: #f1f5f9;
  border-radius: 12px;
  margin-bottom: 10px;
}

.grid-item.folder .grid-icon { background: #fef3c7; }
.grid-item.folder .grid-icon svg { color: #f59e0b; width: 28px; height: 28px; }
.grid-item.file .grid-icon svg { color: var(--primary-color); width: 24px; height: 24px; }
.grid-icon.is-playing { background: var(--primary-color); }
.grid-icon.is-playing svg { color: white; }

.grid-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.grid-meta {
  font-size: 10px;
  color: var(--text-muted);
  margin-top: 2px;
}

.grid-item .override-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  font-size: 8px;
  padding: 2px 4px;
}

/* Stats Bar */
.stats-bar {
  padding: 10px 16px;
  border-top: 1px solid var(--border-color);
  font-size: 12px;
  color: var(--text-muted);
  background: #f8fafc;
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
.tabs-small { display: flex; gap: 4px; margin-bottom: 8px; }
.tabs-small button { padding: 4px 8px; font-size: 12px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.tabs-small button.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.file-upload-zone { border: 2px dashed var(--border-color); border-radius: 8px; padding: 24px; text-align: center; }
.file-upload-zone input { display: none; }
.file-label { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.upload-icon { width: 24px; height: 24px; }

/* Audio Player Bar */
.audio-player-bar {
  position: fixed;
  bottom: 0;
  left: 240px;
  right: 0;
  height: 64px;
  background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 20px;
  z-index: 100;
  box-shadow: 0 -4px 20px rgba(0,0,0,0.3);
}

.player-info { display: flex; align-items: center; gap: 12px; min-width: 200px; }
.player-icon { 
  width: 40px; height: 40px; 
  background: rgba(255,255,255,0.15); 
  border-radius: 8px; 
  display: flex; align-items: center; justify-content: center;
  transition: background 0.2s;
}
.player-icon.is-playing { background: rgba(59, 130, 246, 0.5); animation: pulse 2s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.7; } }
.player-icon .icon { width: 20px; height: 20px; color: #fff; }
.player-details { display: flex; flex-direction: column; }
.player-title { color: #fff; font-size: 13px; font-weight: 600; }
.player-path { color: rgba(255,255,255,0.5); font-size: 10px; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.player-controls { display: flex; align-items: center; gap: 12px; flex: 1; }
.ctrl-btn { 
  width: 36px; height: 36px; 
  background: rgba(255,255,255,0.15); 
  border: none; 
  border-radius: 50%; 
  display: flex; align-items: center; justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
}
.ctrl-btn:hover { background: rgba(255,255,255,0.25); }
.ctrl-icon { width: 14px; height: 14px; color: #fff; }

.progress-container { display: flex; align-items: center; gap: 10px; flex: 1; }
.time-display { color: rgba(255,255,255,0.7); font-size: 11px; font-family: monospace; min-width: 36px; }
.time-current { text-align: right; }
.time-duration { text-align: left; }

.progress-bar { 
  flex: 1; 
  height: 8px; 
  background: rgba(255,255,255,0.2); 
  border-radius: 4px; 
  cursor: pointer;
  position: relative;
}
.progress-bar:hover { background: rgba(255,255,255,0.3); }
.progress-bar:hover .progress-thumb { opacity: 1; transform: translateY(-50%) scale(1); }
.progress-fill { 
  height: 100%; 
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  border-radius: 4px;
  transition: width 0.05s linear;
}
.progress-thumb {
  position: absolute;
  top: 50%;
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  box-shadow: 0 1px 4px rgba(0,0,0,0.3);
  transform: translateY(-50%) scale(0);
  opacity: 0;
  transition: opacity 0.15s, transform 0.15s;
  margin-left: -7px;
}

.volume-control { display: flex; align-items: center; gap: 8px; min-width: 120px; }
.volume-slider {
  width: 80px;
  height: 4px;
  accent-color: #60a5fa;
  cursor: pointer;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(255,255,255,0.2);
  border-radius: 2px;
  outline: none;
}
.volume-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 12px;
  height: 12px;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0,0,0,0.3);
}

.close-player { 
  width: 28px; height: 28px; 
  background: transparent; 
  border: none; 
  color: rgba(255,255,255,0.5);
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
}
.close-player:hover { color: #fff; }
.close-player svg { width: 18px; height: 18px; }

/* Slide-up animation for player bar */
.slide-up-enter-active, .slide-up-leave-active { transition: transform 0.3s ease, opacity 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); opacity: 0; }

/* Adjust for smaller sidebar */
@media (max-width: 1024px) {
  .audio-player-bar { left: 60px; }
  .folder-sidebar { width: 200px; min-width: 200px; }
}
@media (max-width: 768px) {
  .audio-player-bar { left: 0; }
  .player-info { min-width: auto; }
  .player-details { display: none; }
  .folder-sidebar { display: none; }
}
</style>
