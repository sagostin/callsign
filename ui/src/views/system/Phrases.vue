<template>
  <div class="phrases-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Phrases</h2>
        <p class="text-muted text-sm">Manage global system recordings and IVR prompts across all tenants.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showRecordModal = true">
          <MicIcon class="btn-icon" /> Record New
        </button>
        <button class="btn-primary" @click="showUploadModal = true">
          <UploadIcon class="btn-icon" /> Upload Phrase
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ phrases.length }}</div>
        <div class="stat-label">Total Phrases</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ languages.length }}</div>
        <div class="stat-label">Languages</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ customCount }}</div>
        <div class="stat-label">Custom</div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input type="text" v-model="searchQuery" placeholder="Search phrases..." class="search-input">
      </div>
      <select v-model="filterLanguage" class="filter-select">
        <option value="">All Languages</option>
        <option v-for="lang in languages" :key="lang" :value="lang">{{ getLanguageName(lang) }}</option>
      </select>
      <select v-model="filterCategory" class="filter-select">
        <option value="">All Categories</option>
        <option value="ivr">IVR Prompts</option>
        <option value="voicemail">Voicemail</option>
        <option value="system">System Messages</option>
        <option value="custom">Custom</option>
      </select>
    </div>

    <!-- Phrases Grid -->
    <div class="phrases-grid">
      <div v-for="phrase in filteredPhrases" :key="phrase.id" class="phrase-card">
        <div class="phrase-header">
          <div class="phrase-info">
            <h4>{{ phrase.name }}</h4>
            <span class="phrase-desc">{{ phrase.description }}</span>
          </div>
          <span class="lang-badge">{{ phrase.language }}</span>
        </div>
        <div class="phrase-meta">
          <span class="category-tag">{{ phrase.category }}</span>
          <span class="file-count">{{ phrase.file_count }} file(s)</span>
        </div>
        <div class="phrase-actions">
          <button class="btn-icon" @click="playPhrase(phrase)" title="Play">
            <PlayIcon v-if="currentlyPlaying !== phrase.id" />
            <PauseIcon v-else class="text-primary" />
          </button>
          <button class="btn-icon" @click="editPhrase(phrase)" title="Edit">
            <EditIcon />
          </button>
          <button class="btn-icon" @click="downloadPhrase(phrase)" title="Download">
            <DownloadIcon />
          </button>
          <button class="btn-icon danger" @click="deletePhrase(phrase)" :disabled="phrase.isSystem" title="Delete">
            <TrashIcon />
          </button>
        </div>
      </div>
    </div>

    <!-- Upload Modal -->
    <div class="modal-overlay" v-if="showUploadModal" @click.self="showUploadModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Upload Phrase</h3>
          <button class="close-btn" @click="showUploadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Phrase Name</label>
            <input v-model="uploadForm.name" class="input-field" placeholder="welcome_message">
          </div>
          <div class="form-group">
            <label>Description</label>
            <input v-model="uploadForm.description" class="input-field" placeholder="Main welcome greeting">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Language</label>
              <select v-model="uploadForm.language" class="input-field">
                <option value="en-us">English (US)</option>
                <option value="en-gb">English (UK)</option>
                <option value="es-mx">Spanish (MX)</option>
                <option value="fr-ca">French (CA)</option>
              </select>
            </div>
            <div class="form-group">
              <label>Category</label>
              <select v-model="uploadForm.category" class="input-field">
                <option value="ivr">IVR Prompts</option>
                <option value="voicemail">Voicemail</option>
                <option value="system">System Messages</option>
                <option value="custom">Custom</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>Audio File</label>
            <div class="file-upload">
              <input type="file" id="phrase-file" accept=".wav,.mp3,.ogg">
              <label for="phrase-file" class="file-label">
                <UploadIcon class="upload-icon" />
                <span>Choose file or drag & drop</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
          <button class="btn-primary">Upload</button>
        </div>
      </div>
    </div>

    <!-- Record Modal -->
    <div class="modal-overlay" v-if="showRecordModal" @click.self="showRecordModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Record New Phrase</h3>
          <button class="close-btn" @click="showRecordModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Phrase Name</label>
            <input v-model="recordForm.name" class="input-field" placeholder="custom_greeting">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Language</label>
              <select v-model="recordForm.language" class="input-field">
                <option value="en-us">English (US)</option>
                <option value="es-mx">Spanish (MX)</option>
              </select>
            </div>
            <div class="form-group">
              <label>Category</label>
              <select v-model="recordForm.category" class="input-field">
                <option value="custom">Custom</option>
                <option value="ivr">IVR Prompts</option>
              </select>
            </div>
          </div>
          <div class="recorder">
            <div class="recorder-controls">
              <button class="record-btn" :class="{ recording: isRecording }" @click="toggleRecording">
                <MicIcon v-if="!isRecording" />
                <StopIcon v-else />
              </button>
              <span class="record-time">{{ recordTime }}</span>
            </div>
            <div class="waveform" v-if="isRecording">
              <div class="wave-bar" v-for="i in 10" :key="i" :style="{ height: Math.random() * 30 + 10 + 'px' }"></div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showRecordModal = false">Cancel</button>
          <button class="btn-primary" :disabled="!hasRecording">Save Recording</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Search as SearchIcon, Upload as UploadIcon, Mic as MicIcon, 
  Play as PlayIcon, Pause as PauseIcon, Edit as EditIcon, 
  Download as DownloadIcon, Trash2 as TrashIcon, Square as StopIcon
} from 'lucide-vue-next'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const searchQuery = ref('')
const filterLanguage = ref('')
const filterCategory = ref('')
const showUploadModal = ref(false)
const showRecordModal = ref(false)
const currentlyPlaying = ref(null)
const isRecording = ref(false)
const recordTime = ref('00:00')
const hasRecording = ref(false)

const uploadForm = ref({ name: '', description: '', language: 'en-us', category: 'custom' })
const recordForm = ref({ name: '', language: 'en-us', category: 'custom' })

const phrases = ref([
  { id: 1, name: 'dnd_activated', description: 'Do Not Disturb Enabled', language: 'en-us', category: 'system', file_count: 1, isSystem: true },
  { id: 2, name: 'voicemail_greeting', description: 'Default Voicemail Greeting', language: 'en-us', category: 'voicemail', file_count: 1, isSystem: true },
  { id: 3, name: 'ivr_welcome', description: 'Main IVR Welcome', language: 'en-us', category: 'ivr', file_count: 2, isSystem: false },
  { id: 4, name: 'ivr_welcome', description: 'Main IVR Welcome (Spanish)', language: 'es-mx', category: 'ivr', file_count: 1, isSystem: false },
  { id: 5, name: 'transfer_connecting', description: 'Transfer connecting message', language: 'en-us', category: 'system', file_count: 1, isSystem: true },
  { id: 6, name: 'queue_position', description: 'Queue position announcement', language: 'en-us', category: 'system', file_count: 1, isSystem: true },
])

const languages = computed(() => [...new Set(phrases.value.map(p => p.language))])
const customCount = computed(() => phrases.value.filter(p => !p.isSystem).length)

const filteredPhrases = computed(() => {
  return phrases.value.filter(p => {
    const matchesSearch = !searchQuery.value || p.name.includes(searchQuery.value) || p.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesLang = !filterLanguage.value || p.language === filterLanguage.value
    const matchesCat = !filterCategory.value || p.category === filterCategory.value
    return matchesSearch && matchesLang && matchesCat
  })
})

const getLanguageName = (code) => {
  const names = { 'en-us': 'English (US)', 'en-gb': 'English (UK)', 'es-mx': 'Spanish (MX)', 'fr-ca': 'French (CA)' }
  return names[code] || code
}

const playPhrase = (phrase) => {
  currentlyPlaying.value = currentlyPlaying.value === phrase.id ? null : phrase.id
}

const editPhrase = (phrase) => {
  const basePath = route.path.startsWith('/system') ? '/system' : '/admin'
  router.push(`${basePath}/phrases/${phrase.id}`)
}

const downloadPhrase = (phrase) => { console.log('Download', phrase.name) }
const deletePhrase = (phrase) => { if (!phrase.isSystem && confirm(`Delete phrase "${phrase.name}"?`)) phrases.value = phrases.value.filter(p => p.id !== phrase.id) }

const toggleRecording = () => {
  isRecording.value = !isRecording.value
  if (!isRecording.value) hasRecording.value = true
}
</script>

<style scoped>
.phrases-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; background: white; padding: 12px; border-radius: 8px; border: 1px solid var(--border-color); }
.search-box { flex: 1; display: flex; align-items: center; gap: 8px; background: #f8fafc; padding: 8px 12px; border-radius: 6px; }
.search-icon { width: 16px; height: 16px; color: var(--text-muted); }
.search-input { border: none; background: none; flex: 1; font-size: 13px; outline: none; }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; background: white; }

.phrases-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }

.phrase-card { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.phrase-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.phrase-header { display: flex; justify-content: space-between; align-items: flex-start; padding: 14px; background: #f8fafc; border-bottom: 1px solid var(--border-color); }
.phrase-info h4 { margin: 0 0 4px; font-size: 13px; font-family: monospace; }
.phrase-desc { font-size: 11px; color: var(--text-muted); }
.lang-badge { font-size: 9px; background: #dbeafe; color: #2563eb; padding: 2px 6px; border-radius: 3px; font-weight: 600; text-transform: uppercase; }
.phrase-meta { display: flex; justify-content: space-between; padding: 10px 14px; font-size: 11px; color: var(--text-muted); }
.category-tag { background: #f1f5f9; padding: 2px 8px; border-radius: 4px; text-transform: capitalize; }
.phrase-actions { display: flex; gap: 4px; padding: 10px 14px; background: #f8fafc; border-top: 1px solid var(--border-color); }
.phrase-actions .btn-icon { width: 28px; height: 28px; background: white; border: 1px solid var(--border-color); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); }
.phrase-actions .btn-icon:hover { color: var(--primary-color); border-color: var(--primary-color); }
.phrase-actions .btn-icon.danger:hover { color: #ef4444; border-color: #ef4444; }
.phrase-actions .btn-icon:disabled { opacity: 0.4; cursor: not-allowed; }
.phrase-actions .btn-icon svg { width: 14px; height: 14px; }
.text-primary { color: var(--primary-color); }

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
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.file-upload { border: 2px dashed var(--border-color); border-radius: 8px; padding: 32px; text-align: center; }
.file-upload input { display: none; }
.file-label { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.upload-icon { width: 24px; height: 24px; }

.recorder { padding: 16px; background: #f8fafc; border-radius: 8px; text-align: center; }
.recorder-controls { display: flex; align-items: center; justify-content: center; gap: 16px; margin-bottom: 16px; }
.record-btn { width: 56px; height: 56px; border-radius: 50%; border: none; background: #ef4444; color: white; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.record-btn.recording { background: #dc2626; animation: pulse 1s infinite; }
.record-btn svg { width: 24px; height: 24px; }
.record-time { font-family: monospace; font-size: 16px; }
.waveform { display: flex; align-items: center; justify-content: center; gap: 3px; height: 40px; }
.wave-bar { width: 4px; background: var(--primary-color); border-radius: 2px; animation: wave 0.5s ease-in-out infinite alternate; }
@keyframes pulse { 0%, 100% { transform: scale(1); } 50% { transform: scale(1.05); } }
@keyframes wave { to { height: 30px; } }
</style>
