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

    <!-- Stats & Language Selector -->
    <div class="top-bar">
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-value">{{ categories.length }}</div>
          <div class="stat-label">Categories</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ totalSounds }}</div>
          <div class="stat-label">Total Sounds</div>
        </div>
        <div class="stat-card highlight">
          <div class="stat-value">{{ languages.length }}</div>
          <div class="stat-label">Languages</div>
        </div>
      </div>
      <div class="language-selector">
        <label>Language:</label>
        <select v-model="selectedLanguage" class="lang-select">
          <option v-for="lang in languages" :key="lang.code" :value="lang.code">
            {{ lang.flag }} {{ lang.name }}
          </option>
        </select>
      </div>
    </div>

    <!-- Categories -->
    <div class="categories-grid">
      <div v-for="category in categories" :key="category.id" class="category-card">
        <div class="category-header" @click="toggleCategory(category.id)">
          <div class="category-info">
            <component :is="getCategoryIcon(category.icon)" class="category-icon" />
            <div>
              <h4>{{ category.name }}</h4>
              <span class="category-meta">{{ getSoundsForCategory(category.id).length }} sounds</span>
            </div>
          </div>
          <div class="category-actions">
            <span class="lang-coverage" :class="getCoverageClass(category.id)">
              {{ getLanguageCoverage(category.id) }}
            </span>
            <ChevronDownIcon class="expand-icon" :class="{ expanded: expandedCategories.includes(category.id) }" />
          </div>
        </div>
        
        <transition name="slide">
          <div v-if="expandedCategories.includes(category.id)" class="category-content">
            <table class="sounds-table">
              <thead>
                <tr>
                  <th>Sound Name</th>
                  <th>Description</th>
                  <th>{{ selectedLanguageName }}</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="sound in getSoundsForCategory(category.id)" :key="sound.id">
                  <td>
                    <span class="sound-name">{{ sound.name }}</span>
                  </td>
                  <td class="desc-cell">{{ sound.description }}</td>
                  <td>
                    <div class="lang-status">
                      <span v-if="hasLanguage(sound, selectedLanguage)" class="has-lang">
                        <CheckIcon /> Available
                      </span>
                      <span v-else class="missing-lang">
                        <XIcon /> Missing
                      </span>
                    </div>
                  </td>
                  <td class="actions-cell">
                    <button class="btn-icon" @click="playSound(sound)" title="Play">
                      <PlayIcon v-if="currentlyPlaying !== sound.id" />
                      <PauseIcon v-else class="playing" />
                    </button>
                    <button class="btn-icon" @click="editSound(sound)" title="Edit">
                      <EditIcon />
                    </button>
                    <button class="btn-icon" @click="uploadForLang(sound)" title="Upload for this language">
                      <UploadIcon />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </transition>
      </div>
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
            <label>Sound Name</label>
            <input v-model="uploadForm.name" class="input-field" placeholder="ivr_option_invalid">
          </div>
          <div class="form-group">
            <label>Description</label>
            <input v-model="uploadForm.description" class="input-field" placeholder="Invalid option selected">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Category</label>
              <select v-model="uploadForm.category" class="input-field">
                <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>Language</label>
              <select v-model="uploadForm.language" class="input-field">
                <option v-for="lang in languages" :key="lang.code" :value="lang.code">{{ lang.name }}</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>Audio File</label>
            <div class="file-upload">
              <input type="file" id="sound-file" accept=".wav,.mp3,.ogg">
              <label for="sound-file" class="file-label">
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
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Upload as UploadIcon, Download as DownloadIcon, ChevronDown as ChevronDownIcon,
  Play as PlayIcon, Pause as PauseIcon, Edit as EditIcon, Check as CheckIcon, X as XIcon,
  Phone, Bell, VolumeX, Volume2, Music, MessageSquare, AlertCircle, Clock
} from 'lucide-vue-next'

const selectedLanguage = ref('en-us')
const expandedCategories = ref(['ivr', 'voicemail'])
const showUploadModal = ref(false)
const showImportModal = ref(false)
const currentlyPlaying = ref(null)

const uploadForm = ref({ name: '', description: '', category: 'ivr', language: 'en-us' })

const languages = ref([
  { code: 'en-us', name: 'English (US)', flag: '🇺🇸' },
  { code: 'en-gb', name: 'English (UK)', flag: '🇬🇧' },
  { code: 'es-mx', name: 'Spanish (MX)', flag: '🇲🇽' },
  { code: 'es-es', name: 'Spanish (ES)', flag: '🇪🇸' },
  { code: 'fr-ca', name: 'French (CA)', flag: '🇨🇦' },
  { code: 'fr-fr', name: 'French (FR)', flag: '🇫🇷' },
  { code: 'de-de', name: 'German', flag: '🇩🇪' },
  { code: 'pt-br', name: 'Portuguese (BR)', flag: '🇧🇷' },
])

const categories = ref([
  { id: 'ivr', name: 'IVR Prompts', icon: 'Phone' },
  { id: 'voicemail', name: 'Voicemail', icon: 'MessageSquare' },
  { id: 'tones', name: 'Tones & Signals', icon: 'Bell' },
  { id: 'errors', name: 'Error Messages', icon: 'AlertCircle' },
  { id: 'queue', name: 'Queue Announcements', icon: 'Clock' },
  { id: 'conference', name: 'Conference', icon: 'Volume2' },
])

const sounds = ref([
  { id: 1, category: 'ivr', name: 'ivr_welcome', description: 'Welcome greeting', languages: ['en-us', 'es-mx', 'fr-ca'] },
  { id: 2, category: 'ivr', name: 'ivr_goodbye', description: 'Goodbye message', languages: ['en-us', 'es-mx'] },
  { id: 3, category: 'ivr', name: 'ivr_invalid_option', description: 'Invalid option selected', languages: ['en-us'] },
  { id: 4, category: 'ivr', name: 'ivr_timeout', description: 'Timeout - no input', languages: ['en-us', 'es-mx', 'fr-ca', 'de-de'] },
  { id: 5, category: 'voicemail', name: 'vm_greeting', description: 'Default voicemail greeting', languages: ['en-us', 'es-mx'] },
  { id: 6, category: 'voicemail', name: 'vm_unavailable', description: 'User unavailable', languages: ['en-us'] },
  { id: 7, category: 'voicemail', name: 'vm_mailbox_full', description: 'Mailbox full notification', languages: ['en-us', 'es-mx', 'fr-ca'] },
  { id: 8, category: 'tones', name: 'tone_beep', description: 'Standard beep tone', languages: ['en-us'] },
  { id: 9, category: 'tones', name: 'tone_busy', description: 'Busy signal', languages: ['en-us'] },
  { id: 10, category: 'errors', name: 'err_number_invalid', description: 'Number not in service', languages: ['en-us', 'es-mx'] },
  { id: 11, category: 'queue', name: 'queue_position', description: 'Queue position announcement', languages: ['en-us', 'es-mx', 'fr-ca'] },
  { id: 12, category: 'queue', name: 'queue_estimated_wait', description: 'Estimated wait time', languages: ['en-us'] },
])

const selectedLanguageName = computed(() => languages.value.find(l => l.code === selectedLanguage.value)?.name || '')
const totalSounds = computed(() => sounds.value.length)

const getSoundsForCategory = (catId) => sounds.value.filter(s => s.category === catId)
const hasLanguage = (sound, langCode) => sound.languages.includes(langCode)

const getLanguageCoverage = (catId) => {
  const catSounds = getSoundsForCategory(catId)
  const covered = catSounds.filter(s => hasLanguage(s, selectedLanguage.value)).length
  return `${covered}/${catSounds.length}`
}

const getCoverageClass = (catId) => {
  const catSounds = getSoundsForCategory(catId)
  const covered = catSounds.filter(s => hasLanguage(s, selectedLanguage.value)).length
  const ratio = covered / catSounds.length
  if (ratio === 1) return 'full'
  if (ratio >= 0.5) return 'partial'
  return 'low'
}

const getCategoryIcon = (icon) => {
  const icons = { Phone, MessageSquare, Bell, AlertCircle, Clock, Volume2 }
  return icons[icon] || Volume2
}

const toggleCategory = (catId) => {
  const idx = expandedCategories.value.indexOf(catId)
  if (idx === -1) expandedCategories.value.push(catId)
  else expandedCategories.value.splice(idx, 1)
}

const playSound = (sound) => {
  currentlyPlaying.value = currentlyPlaying.value === sound.id ? null : sound.id
}

const editSound = (sound) => { console.log('Edit', sound.name) }
const uploadForLang = (sound) => { 
  uploadForm.value.name = sound.name
  uploadForm.value.description = sound.description
  uploadForm.value.category = sound.category
  showUploadModal.value = true
}
</script>

<style scoped>
.sounds-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

.top-bar { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 20px; gap: 20px; }
.stats-row { display: flex; gap: 16px; flex: 1; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.language-selector { display: flex; align-items: center; gap: 8px; background: white; padding: 12px 16px; border-radius: 8px; border: 1px solid var(--border-color); }
.language-selector label { font-size: 12px; font-weight: 600; color: var(--text-muted); }
.lang-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; min-width: 160px; }

.categories-grid { display: flex; flex-direction: column; gap: 12px; }

.category-card { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.category-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; cursor: pointer; }
.category-header:hover { background: #f8fafc; }
.category-info { display: flex; align-items: center; gap: 12px; }
.category-icon { width: 24px; height: 24px; color: var(--primary-color); }
.category-info h4 { margin: 0; font-size: 14px; }
.category-meta { font-size: 11px; color: var(--text-muted); }
.category-actions { display: flex; align-items: center; gap: 12px; }
.lang-coverage { font-size: 11px; font-weight: 600; padding: 4px 10px; border-radius: 4px; }
.lang-coverage.full { background: #dcfce7; color: #16a34a; }
.lang-coverage.partial { background: #fef3c7; color: #d97706; }
.lang-coverage.low { background: #fef2f2; color: #dc2626; }
.expand-icon { width: 20px; height: 20px; color: var(--text-muted); transition: transform 0.2s; }
.expand-icon.expanded { transform: rotate(180deg); }
.category-content { padding: 0 20px 20px; }

.sounds-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.sounds-table th { text-align: left; padding: 10px 12px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); background: #f8fafc; }
.sounds-table td { padding: 12px; border-bottom: 1px solid var(--border-color); }
.sounds-table tr:hover { background: #f8fafc; }
.sound-name { font-family: monospace; font-weight: 600; }
.desc-cell { color: var(--text-muted); font-size: 12px; }
.lang-status { display: flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 500; }
.has-lang { color: #16a34a; display: flex; align-items: center; gap: 4px; }
.missing-lang { color: #dc2626; display: flex; align-items: center; gap: 4px; }
.lang-status svg { width: 12px; height: 12px; }
.actions-cell { display: flex; gap: 4px; }
.actions-cell .btn-icon { width: 28px; height: 28px; background: white; border: 1px solid var(--border-color); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); }
.actions-cell .btn-icon:hover { color: var(--primary-color); border-color: var(--primary-color); }
.actions-cell .btn-icon svg { width: 14px; height: 14px; }
.actions-cell .btn-icon .playing { color: var(--primary-color); }

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
.file-upload { border: 2px dashed var(--border-color); border-radius: 8px; padding: 24px; text-align: center; }
.file-upload input { display: none; }
.file-label { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.upload-icon { width: 24px; height: 24px; }

/* Transitions */
.slide-enter-active, .slide-leave-active { transition: all 0.3s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; max-height: 0; }
</style>
