<template>
  <div class="greetings-page">
    <div class="page-header">
      <h2>My Voicemail Greetings</h2>
      <p class="text-muted">Create AI-generated greetings for your voicemail. Type your message and we'll generate a professional audio greeting.</p>
    </div>

    <!-- Existing Greetings List -->
    <div class="greetings-list" v-if="greetings.length > 0">
      <div class="section-title">Your Greetings</div>
      <div v-for="g in greetings" :key="g.id" class="greeting-card" :class="{ active: isActive(g) }">
        <div class="greeting-main">
          <div class="greeting-icon" :class="g.status">
            <MicIcon class="icon-sm" />
          </div>
          <div class="greeting-info">
            <div class="greeting-name">
              {{ g.name }}
              <span v-if="isActive(g)" class="active-badge">Active</span>
            </div>
            <div class="greeting-text">{{ truncate(g.script_text, 100) }}</div>
            <div class="greeting-meta">
              <span>{{ g.provider }}</span>
              <span v-if="g.duration">· {{ formatDuration(g.duration) }}</span>
              <span v-if="g.generated_at">· Generated {{ formatDate(g.generated_at) }}</span>
            </div>
          </div>
        </div>
        <div class="greeting-actions">
          <button class="btn-sm btn-outline" @click="playGreeting(g)" :disabled="g.status !== 'ready'" title="Play">
            <PlayIcon v-if="!isPlaying(g)" class="icon-xs" />
            <PauseIcon v-else class="icon-xs" />
          </button>
          <button class="btn-sm btn-primary-sm" @click="activateGreeting(g)" 
            :disabled="g.status !== 'ready' || isActive(g)" title="Set Active">
            <CheckIcon class="icon-xs" />
          </button>
          <button class="btn-sm btn-outline" @click="editGreeting(g)" title="Edit">
            <EditIcon class="icon-xs" />
          </button>
          <button class="btn-sm btn-danger-sm" @click="deleteGreeting(g)" title="Delete">
            <TrashIcon class="icon-xs" />
          </button>
        </div>
      </div>
    </div>

    <!-- Create / Edit Section -->
    <div class="create-section">
      <div class="section-title">{{ editing ? 'Edit Greeting' : 'Create New Greeting' }}</div>
      <div class="create-card">
        <div class="form-group">
          <label>Greeting Name</label>
          <input v-model="form.name" class="input-field" placeholder="E.g., My Voicemail Greeting">
        </div>

        <div class="form-group">
          <label>What should your greeting say?</label>
          <textarea v-model="form.script_text" class="input-field script-input" rows="5"
            placeholder="Hi, you've reached [Your Name]. I'm unable to take your call right now. Please leave a message after the tone and I'll get back to you as soon as possible. Thank you!"></textarea>
          <span class="char-count">{{ (form.script_text || '').length }} characters</span>
        </div>

        <div class="voice-row">
          <div class="form-group flex-1">
            <label>Voice</label>
            <select v-model="form.voice_id" class="input-field">
              <option v-for="v in voices" :key="v.voice_id" :value="v.voice_id">{{ v.name }}</option>
            </select>
          </div>
          <div class="form-group" style="width: 140px;">
            <label>Speed ({{ (form.speed || 1.0).toFixed(1) }}x)</label>
            <input type="range" v-model.number="form.speed" min="0.5" max="2.0" step="0.1" class="slider">
          </div>
        </div>

        <!-- Preview Player -->
        <div v-if="previewAudio" class="preview-player">
          <div class="preview-info">
            <MicIcon class="icon-sm" />
            <span>Preview</span>
          </div>
          <div class="preview-controls">
            <button class="btn-icon" @click="togglePreview">
              <PlayIcon v-if="!previewPlaying" class="icon-sm" />
              <PauseIcon v-else class="icon-sm" />
            </button>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: previewProgress + '%' }"></div>
            </div>
            <button class="btn-icon" @click="stopPreview">
              <XIcon class="icon-sm" />
            </button>
          </div>
        </div>

        <div class="form-actions">
          <button v-if="editing" class="btn-secondary" @click="cancelEdit">Cancel</button>
          <button class="btn-primary" @click="saveGreeting" :disabled="isSaving || !form.script_text">
            <RefreshCwIcon v-if="isSaving" class="icon-sm spin" />
            {{ isSaving ? 'Generating...' : (editing ? 'Save & Regenerate' : 'Generate Greeting') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { 
  Mic as MicIcon, Play as PlayIcon, Pause as PauseIcon,
  Edit as EditIcon, Trash2 as TrashIcon, X as XIcon,
  Check as CheckIcon, RefreshCw as RefreshCwIcon
} from 'lucide-vue-next'
import { userGreetingsAPI, greetingsAPI } from '../../services/api'

const greetings = ref([])
const voices = ref([])
const editing = ref(null)
const isSaving = ref(false)
const activeGreetingPath = ref('')

const form = ref({
  name: 'My Voicemail Greeting',
  script_text: '',
  provider: 'flite',
  voice_id: 'default',
  speed: 1.0,
  language: 'en-US'
})

const loadGreetings = async () => {
  try {
    const res = await userGreetingsAPI.list()
    greetings.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load greetings', e)
  }
}

const loadVoices = async () => {
  try {
    const res = await greetingsAPI.listVoices()
    voices.value = res.data.data || []
  } catch (e) {
    // Default voice if API fails
    voices.value = [{ voice_id: 'default', name: 'Default', provider: 'flite' }]
  }
}

onMounted(() => {
  loadGreetings()
  loadVoices()
})

const saveGreeting = async () => {
  if (!form.value.script_text) return
  isSaving.value = true
  try {
    if (editing.value) {
      await userGreetingsAPI.update(editing.value.id, { ...form.value, regenerate: true })
    } else {
      await userGreetingsAPI.create(form.value)
    }
    cancelEdit()
    loadGreetings()
  } catch (e) {
    console.error('Save failed', e)
    alert('Failed to generate greeting: ' + (e.message || 'Unknown error'))
  } finally {
    isSaving.value = false
  }
}

const editGreeting = (g) => {
  editing.value = g
  form.value = {
    name: g.name, script_text: g.script_text,
    provider: g.provider, voice_id: g.voice_id,
    speed: g.speed || 1.0, language: g.language || 'en-US'
  }
}

const cancelEdit = () => {
  editing.value = null
  form.value = { name: 'My Voicemail Greeting', script_text: '', provider: 'flite', voice_id: 'default', speed: 1.0, language: 'en-US' }
}

const deleteGreeting = async (g) => {
  if (!confirm(`Delete "${g.name}"?`)) return
  try {
    await userGreetingsAPI.delete(g.id)
    loadGreetings()
  } catch (e) {
    alert('Failed to delete')
  }
}

const activateGreeting = async (g) => {
  try {
    await userGreetingsAPI.activate(g.id)
    activeGreetingPath.value = g.file_path
    alert('Greeting set as your active voicemail greeting!')
    loadGreetings()
  } catch (e) {
    alert('Failed to activate: ' + (e.message || 'Unknown error'))
  }
}

const isActive = (g) => activeGreetingPath.value && g.file_path === activeGreetingPath.value

const truncate = (text, len) => text && text.length > len ? text.slice(0, len) + '...' : text

// Audio playback
const currentPlayId = ref(null)
const audioObj = ref(null)
const previewAudio = ref(false)
const previewPlaying = ref(false)
const previewProgress = ref(0)

const isPlaying = (g) => currentPlayId.value === g.id && previewPlaying.value

const playGreeting = async (g) => {
  if (currentPlayId.value === g.id) { togglePreview(); return }
  stopPreview()
  currentPlayId.value = g.id
  try {
    const token = localStorage.getItem('token')
    const headers = { 'Authorization': `Bearer ${token}` }
    const res = await fetch(`/api/user/greetings/${g.id}/stream`, { headers })
    if (!res.ok) throw new Error(`Failed: ${res.status}`)
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    audioObj.value = new Audio(url)
    audioObj.value.addEventListener('timeupdate', () => {
      const d = audioObj.value.duration || 0
      previewProgress.value = d ? (audioObj.value.currentTime / d) * 100 : 0
    })
    audioObj.value.addEventListener('ended', () => { previewPlaying.value = false; previewProgress.value = 0; URL.revokeObjectURL(url) })
    audioObj.value.play()
    previewPlaying.value = true
    previewAudio.value = true
  } catch (e) {
    alert('Playback failed: ' + e.message)
  }
}

const togglePreview = () => {
  if (!audioObj.value) return
  if (previewPlaying.value) audioObj.value.pause()
  else audioObj.value.play()
  previewPlaying.value = !previewPlaying.value
}

const stopPreview = () => {
  if (audioObj.value) { audioObj.value.pause(); audioObj.value = null }
  currentPlayId.value = null; previewPlaying.value = false; previewProgress.value = 0; previewAudio.value = false
}

const formatDuration = (secs) => {
  if (!secs) return '0:00'
  const m = Math.floor(secs / 60), s = Math.floor(secs % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}
const formatDate = (d) => d ? new Date(d).toLocaleDateString() : ''
</script>

<style scoped>
.greetings-page { max-width: 700px; margin: 0 auto; }
.page-header { margin-bottom: 24px; }
.page-header h2 { margin-bottom: 4px; }
.text-muted { color: var(--text-muted); font-size: 13px; }

.section-title { font-weight: 600; font-size: 14px; margin-bottom: 12px; }

/* Greeting Cards */
.greetings-list { margin-bottom: 32px; }
.greeting-card {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 16px; border: 1px solid var(--border-color); border-radius: 10px;
  margin-bottom: 10px; transition: all var(--transition-fast); background: white;
}
.greeting-card.active { border-color: var(--primary-color); background: #f0f4ff; }
.greeting-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.06); }
.greeting-main { display: flex; align-items: center; gap: 12px; flex: 1; min-width: 0; }
.greeting-icon { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.greeting-icon.ready { background: #d1fae5; color: #047857; }
.greeting-icon.draft { background: #f1f5f9; color: #64748b; }
.greeting-icon.generating { background: #dbeafe; color: #2563eb; }
.greeting-icon.error { background: #fee2e2; color: #dc2626; }
.greeting-info { flex: 1; min-width: 0; }
.greeting-name { font-weight: 600; font-size: 14px; display: flex; align-items: center; gap: 8px; }
.active-badge { font-size: 10px; font-weight: 700; background: var(--primary-color); color: white; padding: 2px 8px; border-radius: 4px; text-transform: uppercase; }
.greeting-text { font-size: 12px; color: var(--text-muted); margin: 2px 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.greeting-meta { font-size: 11px; color: var(--text-muted); }
.greeting-actions { display: flex; gap: 6px; flex-shrink: 0; }

/* Create Section */
.create-section { margin-top: 8px; }
.create-card { background: white; border: 1px solid var(--border-color); border-radius: 12px; padding: 20px; }

/* Form */
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: 8px; font-size: 14px; box-sizing: border-box; width: 100%; }
.input-field:focus { border-color: var(--primary-color); outline: none; }
.script-input { font-family: inherit; resize: vertical; min-height: 100px; line-height: 1.6; }
.char-count { font-size: 11px; color: var(--text-muted); text-align: right; }
.voice-row { display: flex; gap: 16px; }
.flex-1 { flex: 1; }
.slider { width: 100%; -webkit-appearance: none; height: 6px; border-radius: 3px; background: #e2e8f0; outline: none; margin-top: 8px; }
.slider::-webkit-slider-thumb { -webkit-appearance: none; width: 18px; height: 18px; border-radius: 50%; background: var(--primary-color); cursor: pointer; }

/* Buttons */
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 6px; font-size: 14px; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 10px 20px; border-radius: var(--radius-sm); font-size: 14px; font-weight: 500; cursor: pointer; }
.btn-sm { background: none; border: 1px solid var(--border-color); padding: 6px 8px; border-radius: 6px; cursor: pointer; display: flex; align-items: center; }
.btn-sm:hover { background: var(--bg-app); }
.btn-sm:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-outline { background: transparent; }
.btn-primary-sm { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.btn-primary-sm:disabled { opacity: 0.4; }
.btn-danger-sm { color: var(--status-bad); }
.btn-danger-sm:hover { background: #fee2e2; }
.btn-icon { background: none; border: none; cursor: pointer; padding: 4px; color: var(--text-muted); }
.btn-icon:hover { color: var(--text-main); }
.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 14px; height: 14px; }
.form-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }

/* Preview Player */
.preview-player { display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; background: linear-gradient(135deg, #4f46e5, #6366f1); border-radius: 8px; color: white; margin-bottom: 16px; }
.preview-info { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.preview-controls { display: flex; align-items: center; gap: 10px; }
.preview-controls .btn-icon { color: white; }
.progress-bar { width: 140px; height: 4px; background: rgba(255,255,255,0.3); border-radius: 2px; overflow: hidden; }
.progress-fill { height: 100%; background: white; transition: width 0.3s; }

@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
.spin { animation: spin 1s linear infinite; }
</style>
