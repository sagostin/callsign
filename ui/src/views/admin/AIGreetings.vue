<template>
  <div class="view-header">
    <div class="header-content">
      <h2>AI Greetings</h2>
      <p class="text-muted text-sm">Generate and manage AI-powered audio greetings for IVRs, voicemail, queues, and announcements.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="showCreateModal = true">
        <PlusIcon class="icon-sm" /> New Script
      </button>
    </div>
  </div>

  <!-- Category Filter Tabs -->
  <div class="tab-bar">
    <button v-for="cat in categories" :key="cat.value" 
      class="tab-btn" :class="{ active: filterCategory === cat.value }" 
      @click="filterCategory = cat.value">
      {{ cat.label }}
    </button>
  </div>

  <!-- Search -->
  <div class="filter-bar">
    <div class="search-box">
      <SearchIcon class="search-icon" />
      <input type="text" v-model="searchQuery" placeholder="Search scripts..." class="search-input">
    </div>
    <select v-model="filterStatus" class="filter-select">
      <option value="">All Status</option>
      <option value="draft">Draft</option>
      <option value="ready">Ready</option>
      <option value="generating">Generating</option>
      <option value="error">Error</option>
    </select>
  </div>

  <!-- Scripts Table -->
  <div class="audio-list">
    <DataTable :columns="columns" :data="filteredScripts" actions>
      <template #name="{ value, row }">
        <div class="file-info">
          <div class="file-icon" :class="getCategoryClass(row.category)">
            <MicIcon class="icon-sm" />
          </div>
          <div>
            <span class="file-name">{{ value }}</span>
            <span class="file-category">{{ row.category }}</span>
          </div>
        </div>
      </template>
      <template #status="{ value }">
        <span class="status-badge" :class="'status-' + value">{{ value }}</span>
      </template>
      <template #provider="{ value }">
        <span class="provider-badge">{{ value }}</span>
      </template>
      <template #duration="{ value }">
        <span class="mono-text" v-if="value">{{ formatDuration(value) }}</span>
        <span class="text-muted" v-else>—</span>
      </template>
      <template #generated_at="{ value }">
        <span v-if="value">{{ formatDate(value) }}</span>
        <span class="text-muted" v-else>Not generated</span>
      </template>
      <template #actions="{ row }">
        <div class="action-buttons">
          <button class="btn-icon" title="Play" @click="playAudio(row)" :disabled="row.status !== 'ready'">
            <PlayIcon v-if="!isPlaying(row)" class="icon-sm" />
            <PauseIcon v-else class="icon-sm text-primary" />
          </button>
          <button class="btn-icon" title="Regenerate" @click="regenerateScript(row)" :disabled="regenerating === row.id">
            <RefreshCwIcon class="icon-sm" :class="{ 'spin': regenerating === row.id }" />
          </button>
          <button class="btn-icon" title="Edit" @click="editScript(row)">
            <EditIcon class="icon-sm" />
          </button>
          <button class="btn-icon" title="Copy Path" @click="copyPath(row)" :disabled="!row.file_path">
            <CopyIcon class="icon-sm" />
          </button>
          <button class="btn-icon text-bad" title="Delete" @click="deleteScript(row)">
            <TrashIcon class="icon-sm" />
          </button>
        </div>
      </template>
    </DataTable>

    <!-- Audio Player Bar -->
    <div v-if="currentAudio" class="audio-player-bar">
      <div class="player-info">
        <MicIcon class="icon-sm" />
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

  <!-- Create/Edit Modal -->
  <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModal">
    <div class="modal-card modal-lg">
      <div class="modal-header">
        <h3>{{ showEditModal ? 'Edit Script' : 'New Greeting Script' }}</h3>
        <button class="close-btn" @click="closeModal">×</button>
      </div>
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-1">
            <label>Name</label>
            <input v-model="form.name" class="input-field" placeholder="E.g., Main IVR Welcome">
          </div>
          <div class="form-group" style="width: 160px;">
            <label>Category</label>
            <select v-model="form.category" class="input-field">
              <option value="ivr">IVR</option>
              <option value="voicemail">Voicemail</option>
              <option value="queue">Queue</option>
              <option value="announcement">Announcement</option>
              <option value="custom">Custom</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>Description</label>
          <input v-model="form.description" class="input-field" placeholder="Optional description">
        </div>

        <div class="form-group">
          <label>Script Text</label>
          <div class="script-editor">
            <div class="line-numbers">
              <div v-for="n in lineCount" :key="n" class="line-num">{{ n }}</div>
            </div>
            <textarea v-model="form.script_text" class="script-textarea" 
              placeholder="Thank you for calling Example Company.&#10;For sales, press 1.&#10;For support, press 2.&#10;To speak with an operator, press 0."
              rows="8" @input="updateLineCount"></textarea>
          </div>
          <span class="help-text">{{ charCount }} characters · {{ lineCount }} lines</span>
        </div>

        <div class="divider"></div>
        <div class="section-title">Voice Settings</div>

        <div class="form-row">
          <div class="form-group flex-1">
            <label>Provider</label>
            <select v-model="form.provider" class="input-field" @change="onProviderChange">
              <option v-for="p in availableProviders" :key="p.id" :value="p.id" :disabled="!p.available">
                {{ p.name }} {{ !p.available ? '(not configured)' : '' }}
              </option>
            </select>
          </div>
          <div class="form-group flex-1">
            <label>Voice</label>
            <select v-model="form.voice_id" class="input-field">
              <option v-for="v in filteredVoices" :key="v.voice_id" :value="v.voice_id">
                {{ v.name }}
              </option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group flex-1">
            <label>Speed ({{ form.speed.toFixed(1) }}x)</label>
            <input type="range" v-model.number="form.speed" min="0.5" max="2.0" step="0.1" class="slider">
          </div>
          <div class="form-group" style="width: 160px;">
            <label>Language</label>
            <select v-model="form.language" class="input-field">
              <option value="en-US">English (US)</option>
              <option value="en-GB">English (UK)</option>
              <option value="es-MX">Spanish (MX)</option>
              <option value="fr-CA">French (CA)</option>
            </select>
          </div>
        </div>

        <!-- Preview Player (for generated scripts being edited) -->
        <div v-if="showEditModal && editForm?.status === 'ready'" class="preview-section">
          <div class="section-title">Current Audio</div>
          <div class="preview-bar">
            <button class="btn-sm btn-outline" @click="playAudio(editForm)">
              <PlayIcon class="icon-sm" /> Preview
            </button>
            <span class="text-muted text-sm">{{ formatDuration(editForm.duration) }} · {{ formatSize(editForm.file_size) }}</span>
          </div>
        </div>
      </div>
      <div class="modal-footer">
        <button class="btn-secondary" @click="closeModal">Cancel</button>
        <button class="btn-secondary" @click="saveScript(false)" :disabled="isSaving">
          {{ isSaving ? 'Saving...' : 'Save Draft' }}
        </button>
        <button class="btn-primary" @click="saveScript(true)" :disabled="isSaving">
          <RefreshCwIcon class="icon-sm" /> {{ isSaving ? 'Generating...' : 'Save & Generate' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import { 
  Search as SearchIcon, Plus as PlusIcon, Mic as MicIcon,
  Play as PlayIcon, Pause as PauseIcon, Edit as EditIcon,
  Trash2 as TrashIcon, X as XIcon, RefreshCw as RefreshCwIcon,
  Copy as CopyIcon
} from 'lucide-vue-next'
import { greetingsAPI } from '../../services/api'

const searchQuery = ref('')
const filterCategory = ref('')
const filterStatus = ref('')
const showCreateModal = ref(false)
const showEditModal = ref(false)
const isSaving = ref(false)
const regenerating = ref(null)
const scripts = ref([])
const voices = ref([])
const availableProviders = ref([])

const categories = [
  { value: '', label: 'All' },
  { value: 'ivr', label: 'IVR' },
  { value: 'voicemail', label: 'Voicemail' },
  { value: 'queue', label: 'Queue' },
  { value: 'announcement', label: 'Announcement' },
  { value: 'custom', label: 'Custom' },
]

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'provider', label: 'Provider', width: '110px' },
  { key: 'status', label: 'Status', width: '100px' },
  { key: 'duration', label: 'Duration', width: '90px' },
  { key: 'generated_at', label: 'Generated', width: '130px' },
]

const defaultForm = () => ({
  name: '', description: '', script_text: '', category: 'ivr',
  provider: 'flite', voice_id: 'default', speed: 1.0, pitch: 1.0, language: 'en-US'
})

const form = ref(defaultForm())
const editForm = ref(null)

const lineCount = computed(() => {
  const text = form.value.script_text || ''
  return Math.max(text.split('\n').length, 1)
})

const charCount = computed(() => (form.value.script_text || '').length)

const filteredScripts = computed(() => {
  return scripts.value.filter(s => {
    const matchesSearch = !searchQuery.value || 
      s.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (s.description || '').toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesCategory = !filterCategory.value || s.category === filterCategory.value
    const matchesStatus = !filterStatus.value || s.status === filterStatus.value
    return matchesSearch && matchesCategory && matchesStatus
  })
})

const filteredVoices = computed(() => {
  return voices.value.filter(v => v.provider === form.value.provider)
})

const loadScripts = async () => {
  try {
    const res = await greetingsAPI.listScripts()
    scripts.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load scripts', e)
  }
}

const loadVoices = async () => {
  try {
    const res = await greetingsAPI.listVoices()
    voices.value = res.data.data || []
    availableProviders.value = res.data.providers || []
  } catch (e) {
    console.error('Failed to load voices', e)
  }
}

onMounted(() => {
  loadScripts()
  loadVoices()
})

const onProviderChange = () => {
  const available = filteredVoices.value
  if (available.length > 0) {
    form.value.voice_id = available[0].voice_id
  }
}

const saveScript = async (generate = false) => {
  if (!form.value.name || !form.value.script_text) {
    alert('Name and script text are required')
    return
  }
  isSaving.value = true
  try {
    if (showEditModal.value && editForm.value) {
      await greetingsAPI.updateScript(editForm.value.id, form.value)
      if (generate) {
        await greetingsAPI.generateAudio(editForm.value.id)
      }
    } else {
      await greetingsAPI.createScript({ ...form.value, generate })
    }
    closeModal()
    loadScripts()
  } catch (e) {
    console.error('Save failed', e)
    alert('Failed to save: ' + (e.message || 'Unknown error'))
  } finally {
    isSaving.value = false
  }
}

const editScript = (row) => {
  editForm.value = row
  form.value = {
    name: row.name, description: row.description || '', script_text: row.script_text,
    category: row.category, provider: row.provider, voice_id: row.voice_id,
    speed: row.speed || 1.0, pitch: row.pitch || 1.0, language: row.language || 'en-US'
  }
  showEditModal.value = true
}

const deleteScript = async (row) => {
  if (!confirm(`Delete "${row.name}"? This will also remove any generated audio.`)) return
  try {
    await greetingsAPI.deleteScript(row.id)
    loadScripts()
  } catch (e) {
    console.error('Delete failed', e)
    alert('Failed to delete')
  }
}

const regenerateScript = async (row) => {
  regenerating.value = row.id
  try {
    await greetingsAPI.generateAudio(row.id)
    loadScripts()
  } catch (e) {
    console.error('Regeneration failed', e)
    alert('Generation failed: ' + (e.message || 'Unknown error'))
  } finally {
    regenerating.value = null
  }
}

const copyPath = (row) => {
  if (row.file_path) {
    navigator.clipboard.writeText(row.file_path)
  }
}

const closeModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editForm.value = null
  form.value = defaultForm()
}

// Audio Player
const currentAudio = ref(null)
const playing = ref(false)
const audioObj = ref(null)
const progress = ref(0)
const currentTime = ref(0)
const duration = ref(0)

const isPlaying = (row) => currentAudio.value?.id === row.id && playing.value

const playAudio = async (row) => {
  if (currentAudio.value?.id === row.id) { togglePlay(); return }
  stopAudio()
  currentAudio.value = row
  try {
    const token = localStorage.getItem('token')
    const tenantId = localStorage.getItem('tenantId')
    const headers = { 'Authorization': `Bearer ${token}` }
    if (tenantId) headers['X-Tenant-ID'] = tenantId
    const res = await fetch(`/api/greetings/scripts/${row.id}/stream`, { headers })
    if (!res.ok) throw new Error(`Failed: ${res.status}`)
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    audioObj.value = new Audio(url)
    audioObj.value.addEventListener('timeupdate', () => {
      currentTime.value = audioObj.value.currentTime
      duration.value = audioObj.value.duration || 0
      progress.value = duration.value ? (currentTime.value / duration.value) * 100 : 0
    })
    audioObj.value.addEventListener('ended', () => { playing.value = false; progress.value = 0; URL.revokeObjectURL(url) })
    audioObj.value.play()
    playing.value = true
  } catch (e) {
    console.error('Playback failed', e)
    alert('Failed to play: ' + e.message)
  }
}

const togglePlay = () => {
  if (!audioObj.value) return
  if (playing.value) audioObj.value.pause()
  else audioObj.value.play()
  playing.value = !playing.value
}

const stopAudio = () => {
  if (audioObj.value) { audioObj.value.pause(); audioObj.value = null }
  currentAudio.value = null; playing.value = false; progress.value = 0; currentTime.value = 0; duration.value = 0
}

// Formatters
const formatDuration = (secs) => {
  if (!secs) return '0:00'
  const m = Math.floor(secs / 60), s = Math.floor(secs % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}
const formatTime = formatDuration
const formatDate = (d) => d ? new Date(d).toLocaleDateString() : '—'
const formatSize = (b) => {
  if (!b) return '0 B'
  const k = 1024
  if (b < k) return b + ' B'
  const s = ['KB', 'MB']; const i = Math.floor(Math.log(b) / Math.log(k))
  return parseFloat((b / Math.pow(k, i)).toFixed(1)) + ' ' + s[i-1]
}
const getCategoryClass = (cat) => cat || 'custom'
const updateLineCount = () => {}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.text-sm { font-size: 13px; }
.text-muted { color: var(--text-muted); }

/* Tab Bar */
.tab-bar { display: flex; gap: 2px; margin-bottom: 16px; background: var(--bg-app); padding: 4px; border-radius: var(--radius-sm); }
.tab-btn {
  padding: 8px 16px; border: none; background: transparent; font-size: 13px; font-weight: 500;
  color: var(--text-muted); cursor: pointer; border-radius: var(--radius-sm); transition: all var(--transition-fast);
}
.tab-btn.active { background: white; color: var(--primary-color); box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.tab-btn:hover:not(.active) { color: var(--text-main); }

/* Filter */
.filter-bar { display: flex; gap: 12px; padding: 12px 0; }
.search-box { position: relative; flex: 1; max-width: 280px; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 12px 8px 34px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; min-width: 130px; }

/* File Info */
.file-info { display: flex; align-items: center; gap: 10px; }
.file-icon { width: 32px; height: 32px; border-radius: 6px; display: flex; align-items: center; justify-content: center; background: #e0e7ff; color: #4f46e5; }
.file-icon.ivr { background: #dbeafe; color: #2563eb; }
.file-icon.voicemail { background: #fce7f3; color: #db2777; }
.file-icon.queue { background: #d1fae5; color: #059669; }
.file-icon.announcement { background: #fef3c7; color: #d97706; }
.file-name { font-weight: 500; display: block; }
.file-category { font-size: 11px; color: var(--text-muted); text-transform: capitalize; }

/* Status Badges */
.status-badge { padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 700; text-transform: uppercase; }
.status-draft { background: #f1f5f9; color: #64748b; }
.status-ready { background: #d1fae5; color: #047857; }
.status-generating { background: #dbeafe; color: #2563eb; }
.status-error { background: #fee2e2; color: #dc2626; }
.provider-badge { padding: 2px 8px; border-radius: 4px; font-size: 10px; font-weight: 600; background: #f1f5f9; color: #475569; text-transform: capitalize; }
.mono-text { font-family: monospace; font-size: 12px; color: var(--text-muted); }

/* Buttons */
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: 13px; font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-sm { padding: 6px 12px; font-size: 12px; }
.btn-outline { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); cursor: pointer; display: flex; align-items: center; gap: 4px; }
.btn-icon { background: none; border: none; padding: 6px; border-radius: var(--radius-sm); cursor: pointer; color: var(--text-muted); transition: all var(--transition-fast); }
.btn-icon:hover { background: var(--bg-app); color: var(--text-main); }
.btn-icon.text-bad:hover { background: #fee2e2; color: var(--status-bad); }
.btn-icon:disabled { opacity: 0.4; cursor: not-allowed; }
.text-primary { color: var(--primary-color); }
.text-bad { color: var(--status-bad); }
.action-buttons { display: flex; gap: 4px; }
.icon-sm { width: 16px; height: 16px; }

/* Audio Player */
.audio-player-bar { display: flex; align-items: center; justify-content: space-between; padding: 12px 16px; background: linear-gradient(135deg, #4f46e5, #6366f1); border-radius: 0 0 var(--radius-md) var(--radius-md); color: white; }
.player-info { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 500; }
.player-controls { display: flex; align-items: center; gap: 12px; }
.player-controls .btn-icon { color: white; }
.player-controls .btn-icon:hover { background: rgba(255,255,255,0.2); }
.progress-bar { width: 200px; height: 4px; background: rgba(255,255,255,0.3); border-radius: 2px; overflow: hidden; }
.progress-fill { height: 100%; background: white; transition: width 0.3s; }
.time-display { font-size: 11px; font-family: monospace; opacity: 0.9; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 640px; max-height: 90vh; overflow-y: auto; }
.modal-lg { max-width: 680px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); position: sticky; top: 0; background: white; z-index: 1; }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); position: sticky; bottom: 0; background: white; }

/* Form */
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.form-row { display: flex; gap: 16px; }
.flex-1 { flex: 1; }
.input-field { padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; width: 100%; }
.input-field:focus { border-color: var(--primary-color); outline: none; }
.help-text { font-size: 11px; color: var(--text-muted); }
.divider { height: 1px; background: var(--border-color); margin: 20px 0; }
.section-title { font-weight: 600; margin-bottom: 12px; font-size: 14px; }

/* Script Editor */
.script-editor { display: flex; border: 1px solid var(--border-color); border-radius: 6px; overflow: hidden; }
.line-numbers { padding: 10px 8px; background: #f8fafc; border-right: 1px solid var(--border-color); user-select: none; min-width: 36px; }
.line-num { font-family: monospace; font-size: 12px; color: #94a3b8; line-height: 1.6; text-align: right; }
.script-textarea { flex: 1; padding: 10px; border: none; font-family: monospace; font-size: 13px; line-height: 1.6; resize: vertical; min-height: 140px; outline: none; }

/* Slider */
.slider { width: 100%; -webkit-appearance: none; height: 6px; border-radius: 3px; background: #e2e8f0; outline: none; }
.slider::-webkit-slider-thumb { -webkit-appearance: none; width: 18px; height: 18px; border-radius: 50%; background: var(--primary-color); cursor: pointer; }

/* Preview */
.preview-section { margin-top: 16px; padding: 12px; background: #f8fafc; border-radius: 8px; }
.preview-bar { display: flex; align-items: center; gap: 12px; }

/* Spin animation */
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.spin { animation: spin 1s linear infinite; }
</style>
