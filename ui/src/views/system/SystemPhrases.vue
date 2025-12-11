<template>
  <div class="phrases-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Phrases</h2>
        <p class="text-muted text-sm">Create phrases from audio recordings played in sequence.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="createPhrase">
          <PlusIcon class="btn-icon" /> New Phrase
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ phrases.length }}</div>
        <div class="stat-label">Total Phrases</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ enabledCount }}</div>
        <div class="stat-label">Enabled</div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" placeholder="Search phrases..." class="search-input">
      </div>
      <select v-model="filterLanguage" class="filter-select">
        <option value="">All Languages</option>
        <option value="en-us">English (US)</option>
        <option value="es-mx">Spanish (MX)</option>
        <option value="fr-ca">French (CA)</option>
      </select>
    </div>

    <!-- Phrases List -->
    <div class="phrases-list">
      <div v-for="phrase in filteredPhrases" :key="phrase.id" class="phrase-card" :class="{ disabled: !phrase.enabled }">
        <div class="phrase-header">
          <div class="phrase-info">
            <h4>{{ phrase.name }}</h4>
            <span class="phrase-desc">{{ phrase.description }}</span>
          </div>
          <div class="phrase-badges">
            <span class="lang-badge">{{ phrase.language }}</span>
            <span class="status-badge" :class="phrase.enabled ? 'enabled' : 'disabled'">
              {{ phrase.enabled ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
        </div>
        
        <div class="phrase-structure">
          <div class="structure-label">Structure:</div>
          <div class="structure-items">
            <div v-for="(item, idx) in phrase.structure" :key="idx" class="structure-item">
              <span class="item-order">{{ item.order }}</span>
              <span class="item-function">{{ item.function }}</span>
              <span class="item-action">{{ item.action }}</span>
            </div>
          </div>
        </div>
        
        <div class="phrase-actions">
          <button class="btn-link" @click="editPhrase(phrase)"><EditIcon /> Edit</button>
          <button class="btn-link" @click="duplicatePhrase(phrase)"><CopyIcon /> Duplicate</button>
          <button class="btn-link danger" @click="deletePhrase(phrase)"><TrashIcon /> Delete</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>{{ editingPhrase?.id ? 'Edit Phrase' : 'New Phrase' }}</h3>
          <button class="close-btn" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group">
              <label>Name</label>
              <input v-model="editForm.name" class="input-field" placeholder="main">
              <span class="help-text">Example: xyz_audio</span>
            </div>
            <div class="form-group">
              <label>Language</label>
              <input v-model="editForm.language" class="input-field" placeholder="en-us">
            </div>
          </div>
          
          <div class="form-group">
            <label>Description</label>
            <input v-model="editForm.description" class="input-field" placeholder="Welcome greeting">
          </div>

          <div class="structure-section">
            <div class="section-header">
              <label>Structure</label>
              <span class="section-hint">Define the various components that make up the phrase.</span>
            </div>
            
            <div class="structure-table">
              <div class="structure-header">
                <span>Function</span>
                <span>Action</span>
                <span>Order</span>
                <span></span>
              </div>
              <div v-for="(item, idx) in editForm.structure" :key="idx" class="structure-row">
                <select v-model="item.function" class="input-field">
                  <option value="Play">Play</option>
                  <option value="Pause">Pause</option>
                  <option value="TTS">TTS</option>
                </select>
                <div class="action-input">
                  <select v-if="item.function === 'Play'" v-model="item.action" class="input-field">
                    <option value="">Select recording...</option>
                    <option v-for="rec in availableRecordings" :key="rec" :value="rec">{{ rec }}</option>
                  </select>
                  <select v-else-if="item.function === 'Pause'" v-model="item.action" class="input-field">
                    <option value="0.5s">0.5 seconds</option>
                    <option value="1s">1 second</option>
                    <option value="2s">2 seconds</option>
                  </select>
                  <input v-else v-model="item.action" class="input-field" placeholder="Text to speak...">
                </div>
                <input v-model="item.order" class="input-field order-input" type="number">
                <button class="btn-icon danger" @click="removeStructureItem(idx)"><TrashIcon /></button>
              </div>
              
              <div class="add-row">
                <select v-model="newItem.function" class="input-field">
                  <option value="Play">Play</option>
                  <option value="Pause">Pause</option>
                  <option value="TTS">TTS</option>
                </select>
                <select v-model="newItem.action" class="input-field">
                  <option value="">Select...</option>
                  <option v-for="rec in availableRecordings" :key="rec" :value="rec">{{ rec }}</option>
                </select>
                <input v-model="newItem.order" class="input-field order-input" type="number" placeholder="000">
                <button class="btn-add" @click="addStructureItem"><PlusIcon /> Add</button>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label class="toggle-label">
              <span>Enabled</span>
              <label class="toggle-switch">
                <input type="checkbox" v-model="editForm.enabled">
                <span class="toggle-slider"></span>
              </label>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-danger" v-if="editingPhrase?.id" @click="deletePhrase(editingPhrase)">Delete</button>
          <button class="btn-primary" @click="savePhrase">Save</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Plus as PlusIcon, Search as SearchIcon, Edit as EditIcon, 
  Copy as CopyIcon, Trash2 as TrashIcon
} from 'lucide-vue-next'

const searchQuery = ref('')
const filterLanguage = ref('')
const showEditModal = ref(false)
const editingPhrase = ref(null)

const editForm = ref({
  name: '',
  language: 'en-us',
  description: '',
  enabled: true,
  structure: []
})

const newItem = ref({ function: 'Play', action: '', order: '000' })

const availableRecordings = [
  'ivr_welcome_en-us.wav',
  'ivr_options_en-us.wav',
  'ivr_thank_you_for_holding.wav',
  'silence_1s',
  'tone_beep'
]

const phrases = ref([
  { 
    id: 1, 
    name: 'main', 
    language: 'en-us', 
    description: 'Welcome greeting',
    enabled: true,
    structure: [
      { function: 'Play', action: 'ivr_welcome_en-us.wav', order: '010' },
      { function: 'Pause', action: '1s', order: '020' },
      { function: 'Play', action: 'ivr_options_en-us.wav', order: '030' }
    ]
  },
  { 
    id: 2, 
    name: 'hold_message', 
    language: 'en-us', 
    description: 'Queue hold message',
    enabled: true,
    structure: [
      { function: 'Play', action: 'ivr_thank_you_for_holding.wav', order: '010' }
    ]
  },
  { 
    id: 3, 
    name: 'main', 
    language: 'es-mx', 
    description: 'Bienvenida principal',
    enabled: true,
    structure: [
      { function: 'Play', action: 'ivr_welcome_es-mx.wav', order: '010' }
    ]
  },
])

const enabledCount = computed(() => phrases.value.filter(p => p.enabled).length)

const filteredPhrases = computed(() => {
  return phrases.value.filter(p => {
    const matchesSearch = !searchQuery.value || p.name.includes(searchQuery.value) || p.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesLang = !filterLanguage.value || p.language === filterLanguage.value
    return matchesSearch && matchesLang
  })
})

const createPhrase = () => {
  editingPhrase.value = null
  editForm.value = { name: '', language: 'en-us', description: '', enabled: true, structure: [] }
  showEditModal.value = true
}

const editPhrase = (phrase) => {
  editingPhrase.value = phrase
  editForm.value = { ...phrase, structure: [...phrase.structure.map(s => ({ ...s }))] }
  showEditModal.value = true
}

const duplicatePhrase = (phrase) => {
  const newPhrase = { ...phrase, id: Date.now(), name: phrase.name + '_copy', structure: [...phrase.structure] }
  phrases.value.push(newPhrase)
}

const deletePhrase = (phrase) => {
  if (confirm(`Delete phrase "${phrase.name}"?`)) {
    phrases.value = phrases.value.filter(p => p.id !== phrase.id)
    closeModal()
  }
}

const addStructureItem = () => {
  editForm.value.structure.push({ ...newItem.value })
  newItem.value = { function: 'Play', action: '', order: String(parseInt(newItem.value.order || 0) + 10).padStart(3, '0') }
}

const removeStructureItem = (idx) => {
  editForm.value.structure.splice(idx, 1)
}

const savePhrase = () => {
  if (editingPhrase.value) {
    const idx = phrases.value.findIndex(p => p.id === editingPhrase.value.id)
    if (idx !== -1) phrases.value[idx] = { ...editForm.value, id: editingPhrase.value.id }
  } else {
    phrases.value.push({ ...editForm.value, id: Date.now() })
  }
  closeModal()
}

const closeModal = () => {
  showEditModal.value = false
  editingPhrase.value = null
}
</script>

<style scoped>
.phrases-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.btn-primary, .btn-secondary, .btn-danger { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-danger { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; max-width: 200px; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-box { flex: 1; display: flex; align-items: center; gap: 8px; background: white; padding: 10px 14px; border-radius: 8px; border: 1px solid var(--border-color); }
.search-icon { width: 16px; height: 16px; color: var(--text-muted); }
.search-input { border: none; background: none; flex: 1; font-size: 13px; outline: none; }
.filter-select { padding: 10px 14px; border: 1px solid var(--border-color); border-radius: 8px; font-size: 13px; background: white; }

.phrases-list { display: flex; flex-direction: column; gap: 12px; }

.phrase-card { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.phrase-card.disabled { opacity: 0.6; }
.phrase-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.phrase-header { display: flex; justify-content: space-between; align-items: flex-start; padding: 14px 16px; background: #f8fafc; border-bottom: 1px solid var(--border-color); }
.phrase-info h4 { margin: 0 0 4px; font-size: 14px; font-family: monospace; }
.phrase-desc { font-size: 12px; color: var(--text-muted); }
.phrase-badges { display: flex; gap: 8px; }
.lang-badge { font-size: 10px; background: #dbeafe; color: #2563eb; padding: 3px 8px; border-radius: 4px; font-weight: 600; }
.status-badge { font-size: 10px; padding: 3px 8px; border-radius: 4px; font-weight: 600; }
.status-badge.enabled { background: #dcfce7; color: #16a34a; }
.status-badge.disabled { background: #f1f5f9; color: #64748b; }

.phrase-structure { padding: 12px 16px; }
.structure-label { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); margin-bottom: 8px; }
.structure-items { display: flex; flex-wrap: wrap; gap: 8px; }
.structure-item { display: flex; align-items: center; gap: 6px; padding: 6px 10px; background: #f8fafc; border-radius: 4px; font-size: 12px; }
.item-order { font-family: monospace; color: var(--text-muted); font-size: 10px; }
.item-function { font-weight: 600; color: var(--primary-color); }
.item-action { font-family: monospace; }

.phrase-actions { display: flex; gap: 12px; padding: 10px 16px; background: #f8fafc; border-top: 1px solid var(--border-color); }
.btn-link { display: flex; align-items: center; gap: 4px; background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 500; cursor: pointer; }
.btn-link.danger { color: #dc2626; }
.btn-link svg { width: 12px; height: 12px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 600px; max-height: 90vh; overflow: hidden; display: flex; flex-direction: column; }
.modal-card.large { max-width: 700px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.help-text { font-size: 10px; color: var(--text-muted); margin-top: 4px; }

.structure-section { margin-bottom: 16px; }
.section-header { margin-bottom: 12px; }
.section-header label { margin-bottom: 4px; }
.section-hint { font-size: 11px; color: var(--text-muted); }
.structure-table { border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.structure-header { display: grid; grid-template-columns: 100px 1fr 60px 32px; gap: 8px; padding: 10px 12px; background: #f8fafc; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.structure-row { display: grid; grid-template-columns: 100px 1fr 60px 32px; gap: 8px; padding: 10px 12px; border-bottom: 1px solid var(--border-color); align-items: center; }
.structure-row:last-child { border-bottom: none; }
.action-input { width: 100%; }
.order-input { width: 60px !important; text-align: center; font-family: monospace; }
.add-row { display: grid; grid-template-columns: 100px 1fr 60px auto; gap: 8px; padding: 10px 12px; background: #f8fafc; align-items: center; }
.btn-add { display: flex; align-items: center; gap: 4px; padding: 6px 12px; background: var(--primary-color); color: white; border: none; border-radius: 6px; font-size: 12px; font-weight: 500; cursor: pointer; }
.btn-add svg { width: 12px; height: 12px; }

.toggle-label { display: flex !important; justify-content: space-between; align-items: center; text-transform: none !important; font-size: 13px !important; cursor: pointer; }
.toggle-switch { position: relative; width: 44px; height: 24px; }
.toggle-switch input { display: none; }
.toggle-slider { position: absolute; inset: 0; background: #cbd5e1; border-radius: 12px; transition: all 0.2s; }
.toggle-slider:before { content: ''; position: absolute; width: 18px; height: 18px; background: white; border-radius: 50%; top: 3px; left: 3px; transition: all 0.2s; }
.toggle-switch input:checked + .toggle-slider { background: var(--primary-color); }
.toggle-switch input:checked + .toggle-slider:before { left: 23px; }
</style>
