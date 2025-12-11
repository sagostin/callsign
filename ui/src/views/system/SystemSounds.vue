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
                    <span v-if="sound.isOverride" class="override-badge">Override</span>
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
                    <!-- Show revert if override, otherwise show upload for replacement -->
                    <button v-if="sound.isOverride" class="btn-icon" @click="deleteOverride(sound)" title="Revert to System Default">
                       <XIcon class="text-bad" />
                    </button>
                    <button v-else class="btn-icon" @click="uploadForLang(sound)" :title="isTenant ? 'Override this sound' : 'Replace sound'">
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
            <label>Source</label>
            <div class="tabs-small">
               <button :class="{ active: uploadSource === 'file' }" @click="uploadSource = 'file'">Upload File</button>
               <button :class="{ active: uploadSource === 'record' }" @click="uploadSource = 'record'">Record Audio</button>
            </div>
          </div>
          
          <div v-if="uploadSource === 'file'" class="form-group">
            <label>Audio File</label>
            <div class="file-upload">
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
             <p class="text-xs text-muted mt-2" v-if="uploadForm.blob">Recording captured ready for upload.</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
          <button class="btn-primary" @click="submitUpload">Upload</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  Upload as UploadIcon, Download as DownloadIcon, ChevronDown as ChevronDownIcon,
  Play as PlayIcon, Pause as PauseIcon, Edit as EditIcon, Check as CheckIcon, X as XIcon,
  Phone, Bell, VolumeX, Volume2, Music, MessageSquare, AlertCircle, Clock
} from 'lucide-vue-next'
import { systemAPI, tenantMediaAPI } from '../../services/api'
import AudioRecorder from '../../components/common/AudioRecorder.vue'

const selectedLanguage = ref('en/us') // changed to path format
const expandedCategories = ref([])
const showUploadModal = ref(false)
const showImportModal = ref(false)
const currentlyPlaying = ref(null)
const isLoading = ref(false)
const rawData = ref([]) // The full tree from API

const uploadForm = ref({ name: '', description: '', category: 'ivr', language: 'en/us', file: null, blob: null })
const uploadSource = ref('file')

// Computed languages based on available folders in rawData
const languages = computed(() => {
  if (!rawData.value.length) return []
  const langs = []
  
  // Expecting structure: lang/region
  rawData.value.forEach(langNode => {
    if (langNode.type !== 'directory') return
    const langCode = langNode.name
    
    if (langNode.children) {
      langNode.children.forEach(regionNode => {
        if (regionNode.type !== 'directory') return
        const regionCode = regionNode.name
        const code = `${langCode}/${regionCode}`
        
        // Form human readable name
        let name = code
        let flag = '🏳️'
        if (code === 'en/us') { name = 'English (US)'; flag = '🇺🇸' }
        else if (code === 'en/gb') { name = 'English (UK)'; flag = '🇬🇧' }
        else if (code === 'fr/ca') { name = 'French (CA)'; flag = '🇨🇦' }
        else if (code === 'es/mx') { name = 'Spanish (MX)'; flag = '🇲🇽' }
        
        langs.push({ code, name, flag })
      })
    }
  })
  
  return langs.length ? langs : [{ code: 'en/us', name: 'English (US)', flag: '🇺🇸' }]
})

// Dynamically generate categories based on available files
// For strict mode we might want predefined categories, but dynamic is fine here.
const categories = computed(() => {
  const cats = new Set()
  // Default categories to ensure order/existence of common ones
  const defaults = [
    { id: 'ivr', name: 'IVR Prompts', icon: 'Phone' },
    { id: 'voicemail', name: 'Voicemail', icon: 'MessageSquare' },
    { id: 'conference', name: 'Conference', icon: 'Volume2' },
    { id: 'digits', name: 'Digits', icon: 'Clock' },
    { id: 'misc', name: 'Misc', icon: 'Bell' },
  ]
  
  // Scan sounds to find other categories
  sounds.value.forEach(s => cats.add(s.category))
  
  const result = defaults.filter(d => cats.has(d.id))
  cats.forEach(c => {
    if (!result.find(r => r.id === c)) {
      result.push({ id: c, name: c.charAt(0).toUpperCase() + c.slice(1), icon: 'Volume2' })
    }
  })
  return result
})

const sounds = computed(() => {
  if (!selectedLanguage.value || !rawData.value.length) return []
  
  const parts = selectedLanguage.value.split('/')
  if (parts.length !== 2) return []
  
  const langNode = rawData.value.find(n => n.name === parts[0])
  if (!langNode || !langNode.children) return []
  
  const regionNode = langNode.children.find(n => n.name === parts[1])
  if (!regionNode || !regionNode.children) return []
  
  // Ideally there is a 'voice' folder next, e.g. 'callie'
  // We will aggregate sounds from all voices or just pick the first one?
  // Let's flatten all sounds found under this language/region
  const flatSounds = []
  
  const traverse = (node, pathStr, category) => {
    if (node.type === 'file') {
      flatSounds.push({
        id: node.path,
        name: node.name,
        path: node.path,
        category: category || 'misc',
        description: node.path,
        languages: [selectedLanguage.value], // Simplification
        isOverride: node.is_override
      })
    } else if (node.children) {
      // If we are deep enough, the folder name is the category
      // structure: lang/region/voice/category/file
      // path parts: [0]=lang, [1]=region, [2]=voice, [3]=category
      const currentParts = node.path.split('/')
      // Remove root prefix if present (api returns relative to root usually, but let's check node.path)
      // Node.path is relative to sounds root. e.g. "en/us/callie/ivr"
      // parts: "en", "us", "callie", "ivr"
      
      let nextCat = category
      if (!category && currentParts.length >= 4) {
        nextCat = currentParts[3]
      }
      
      node.children.forEach(child => traverse(child, child.path, nextCat))
    }
  }
  
  regionNode.children.forEach(voiceNode => {
      if (voiceNode.children) {
          voiceNode.children.forEach(child => traverse(child, child.path, null))
      }
  })
  
  return flatSounds
})

const selectedLanguageName = computed(() => languages.value.find(l => l.code === selectedLanguage.value)?.name || '')
const totalSounds = computed(() => sounds.value.length)

const getSoundsForCategory = (catId) => sounds.value.filter(s => s.category === catId)
const hasLanguage = (sound, langCode) => sound.languages.includes(langCode) // logic simplified for FS files

const getLanguageCoverage = (catId) => {
    return `${getSoundsForCategory(catId).length} files`
}

const getCoverageClass = (catId) => 'full'

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
    // Requires a playback endpoint, not implemented yet.
    // Ideally we'd have a /system/media/play?path=...
    console.log('Play not implemented yet for', sound.path)
    currentlyPlaying.value = currentlyPlaying.value === sound.id ? null : sound.id
}

const editSound = (sound) => { console.log('Edit', sound.name) }

const isTenant = computed(() => !!localStorage.getItem('tenantId'))

const uploadForLang = (sound) => {
  uploadForm.value.name = sound.name
  uploadForm.value.description = ''
  uploadForm.value.category = sound.category
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
        // Create file from blob
        fileToUpload = new File([uploadForm.value.blob], uploadForm.value.name || 'recording.wav', { type: 'audio/wav' })
    }
    
    if (!fileToUpload) return
    
    // Construct path: lang/region/voice/category
    // We need to know the voice. For now, default to first voice found or 'callie'?
    // Let's assume 'callie' for en/us if not found.
    // Actually, we can just put it in the category folder under the current language's first voice.
    
    // Find voice folder
    const parts = uploadForm.value.language.split('/')
    const langNode = rawData.value.find(n => n.name === parts[0])
    const regionNode = langNode?.children?.find(n => n.name === parts[1])
    let voice = 'callie' // default
    if (regionNode?.children?.length) {
        voice = regionNode.children[0].name
    }
    
    const targetPath = `${uploadForm.value.language}/${voice}/${uploadForm.value.category}`
    
    const formData = new FormData()
    formData.append('file', fileToUpload)
    formData.append('path', targetPath)
    
    try {
        if (isTenant.value) {
            await tenantMediaAPI.uploadSound(formData)
        } else {
            await systemAPI.uploadSound(formData)
        }
        showUploadModal.value = false
        loadSounds() // Refresh
    } catch (e) {
        console.error('Upload failed', e)
        alert('Upload failed: ' + (e.response?.data?.error || e.message))
    }
}

const deleteOverride = async (sound) => {
    if (!confirm(`Revert override for ${sound.name}?`)) return
    
    try {
        await tenantMediaAPI.deleteSound(sound.path)
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
.btn-link { background: none; border: none; color: var(--primary-color); cursor: pointer; font-size: 11px; padding: 0; }

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

.override-badge { 
    font-size: 10px; font-weight: 700; color: #d97706; background: #fef3c7; 
    padding: 2px 6px; border-radius: 4px; border: 1px solid #fcd34d; margin-left: 8px;
    text-transform: uppercase;
}

.tabs-small { display: flex; gap: 4px; margin-bottom: 8px; }
.tabs-small button { padding: 4px 8px; font-size: 12px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.tabs-small button.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.mt-2 { margin-top: 8px; }

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
