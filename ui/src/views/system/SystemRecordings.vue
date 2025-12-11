<template>
  <div class="recordings-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Recordings</h2>
        <p class="text-muted text-sm">Manage global audio files for IVR, voicemail, and system prompts.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showRecordModal = true">
          <MicIcon class="btn-icon" /> Record
        </button>
        <button class="btn-primary" @click="showUploadModal = true">
          <UploadIcon class="btn-icon" /> Upload
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ recordings.length }}</div>
        <div class="stat-label">Total Recordings</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ categories.length }}</div>
        <div class="stat-label">Categories</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ totalSize }}</div>
        <div class="stat-label">Storage Used</div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" placeholder="Search recordings..." class="search-input">
      </div>
      <select v-model="filterCategory" class="filter-select">
        <option value="">All Categories</option>
        <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
      </select>
    </div>

    <!-- Recordings Table -->
    <div class="recordings-table">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Category</th>
            <th>Duration</th>
            <th>Format</th>
            <th>Description</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rec in filteredRecordings" :key="rec.id">
            <td class="name-cell">
              <FileAudioIcon class="file-icon" />
              <span class="file-name">{{ rec.name }}</span>
            </td>
            <td><span class="category-badge">{{ rec.category }}</span></td>
            <td class="mono">{{ rec.duration }}</td>
            <td class="mono">{{ rec.format }}</td>
            <td class="desc-cell">{{ rec.description }}</td>
            <td class="actions-cell">
              <button class="btn-icon" @click="playRecording(rec)" :title="currentlyPlaying === rec.id ? 'Stop' : 'Play'">
                <PlayIcon v-if="currentlyPlaying !== rec.id" />
                <PauseIcon v-else class="playing" />
              </button>
              <button class="btn-icon" @click="downloadRecording(rec)" title="Download">
                <DownloadIcon />
              </button>
              <button class="btn-icon" @click="editRecording(rec)" title="Edit">
                <EditIcon />
              </button>
              <button class="btn-icon danger" @click="deleteRecording(rec)" title="Delete">
                <TrashIcon />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
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
            <label>File Name</label>
            <input v-model="uploadForm.name" class="input-field" placeholder="welcome_greeting_en">
            <span class="help-text">Use naming convention: description_language (e.g., welcome_en-us)</span>
          </div>
          <div class="form-group">
            <label>Category</label>
            <select v-model="uploadForm.category" class="input-field">
              <option value="IVR">IVR</option>
              <option value="Voicemail">Voicemail</option>
              <option value="System">System</option>
              <option value="Queue">Queue</option>
              <option value="Custom">Custom</option>
            </select>
          </div>
          <div class="form-group">
            <label>Description</label>
            <input v-model="uploadForm.description" class="input-field" placeholder="Welcome greeting for main IVR">
          </div>
          <div class="form-group">
            <label>Audio File</label>
            <div class="file-upload">
              <input type="file" id="upload-file" accept=".wav,.mp3,.ogg">
              <label for="upload-file" class="file-label">
                <UploadIcon class="upload-icon" />
                <span>Choose WAV, MP3, or OGG file</span>
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
    <div v-if="showRecordModal" class="modal-overlay" @click.self="showRecordModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Record via Phone</h3>
          <button class="close-btn" @click="showRecordModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="record-info">
            <div class="record-step">
              <span class="step-num">1</span>
              <span>Dial <code>*732</code> from any extension</span>
            </div>
            <div class="record-step">
              <span class="step-num">2</span>
              <span>Enter PIN followed by <code>#</code></span>
            </div>
            <div class="record-step">
              <span class="step-num">3</span>
              <span>Enter a 3+ digit recording ID (e.g., 100)</span>
            </div>
            <div class="record-step">
              <span class="step-num">4</span>
              <span>Record after the beep, press <code>#</code> to finish</span>
            </div>
            <div class="record-step">
              <span class="step-num">5</span>
              <span>Press <code>1</code> to save or <code>2</code> to re-record</span>
            </div>
          </div>
          <div class="record-note">
            Recording will appear as <code>recording[ID].wav</code> and can be renamed after.
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showRecordModal = false">Close</button>
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
            <label>File Name</label>
            <input v-model="editForm.name" class="input-field">
          </div>
          <div class="form-group">
            <label>Category</label>
            <select v-model="editForm.category" class="input-field">
              <option value="IVR">IVR</option>
              <option value="Voicemail">Voicemail</option>
              <option value="System">System</option>
              <option value="Queue">Queue</option>
              <option value="Custom">Custom</option>
            </select>
          </div>
          <div class="form-group">
            <label>Description</label>
            <input v-model="editForm.description" class="input-field">
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showEditModal = false">Cancel</button>
          <button class="btn-primary" @click="saveEdit">Save</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Upload as UploadIcon, Mic as MicIcon, Search as SearchIcon,
  FileAudio as FileAudioIcon, Play as PlayIcon, Pause as PauseIcon,
  Download as DownloadIcon, Edit as EditIcon, Trash2 as TrashIcon
} from 'lucide-vue-next'

const searchQuery = ref('')
const filterCategory = ref('')
const showUploadModal = ref(false)
const showRecordModal = ref(false)
const showEditModal = ref(false)
const currentlyPlaying = ref(null)

const uploadForm = ref({ name: '', category: 'IVR', description: '' })
const editForm = ref({ id: null, name: '', category: '', description: '' })

const categories = ['IVR', 'Voicemail', 'System', 'Queue', 'Custom']

const recordings = ref([
  { id: 1, name: 'ivr_welcome_en-us', category: 'IVR', duration: '0:08', format: 'WAV', description: 'Main IVR welcome greeting (English)' },
  { id: 2, name: 'ivr_welcome_es-mx', category: 'IVR', duration: '0:09', format: 'WAV', description: 'Main IVR welcome greeting (Spanish)' },
  { id: 3, name: 'ivr_options_en-us', category: 'IVR', duration: '0:15', format: 'WAV', description: 'Press 1 for sales, press 2 for support...' },
  { id: 4, name: 'vm_unavailable_en-us', category: 'Voicemail', duration: '0:05', format: 'WAV', description: 'User is unavailable' },
  { id: 5, name: 'vm_mailbox_full_en-us', category: 'Voicemail', duration: '0:04', format: 'WAV', description: 'Mailbox full notification' },
  { id: 6, name: 'queue_hold_msg_en-us', category: 'Queue', duration: '0:12', format: 'WAV', description: 'Thank you for waiting message' },
  { id: 7, name: 'tone_beep', category: 'System', duration: '0:01', format: 'WAV', description: 'Standard beep tone' },
  { id: 8, name: 'silence_1s', category: 'System', duration: '0:01', format: 'WAV', description: '1 second silence' },
])

const totalSize = computed(() => '24.5 MB')

const filteredRecordings = computed(() => {
  return recordings.value.filter(r => {
    const matchesSearch = !searchQuery.value || r.name.toLowerCase().includes(searchQuery.value.toLowerCase()) || r.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesCat = !filterCategory.value || r.category === filterCategory.value
    return matchesSearch && matchesCat
  })
})

const playRecording = (rec) => {
  currentlyPlaying.value = currentlyPlaying.value === rec.id ? null : rec.id
}

const downloadRecording = (rec) => console.log('Download', rec.name)

const editRecording = (rec) => {
  editForm.value = { ...rec }
  showEditModal.value = true
}

const saveEdit = () => {
  const idx = recordings.value.findIndex(r => r.id === editForm.value.id)
  if (idx !== -1) recordings.value[idx] = { ...editForm.value }
  showEditModal.value = false
}

const deleteRecording = (rec) => {
  if (confirm(`Delete recording "${rec.name}"?`)) {
    recordings.value = recordings.value.filter(r => r.id !== rec.id)
  }
}
</script>

<style scoped>
.recordings-page { padding: 0; }
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

.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-box { flex: 1; display: flex; align-items: center; gap: 8px; background: white; padding: 10px 14px; border-radius: 8px; border: 1px solid var(--border-color); }
.search-icon { width: 16px; height: 16px; color: var(--text-muted); }
.search-input { border: none; background: none; flex: 1; font-size: 13px; outline: none; }
.filter-select { padding: 10px 14px; border: 1px solid var(--border-color); border-radius: 8px; font-size: 13px; background: white; min-width: 150px; }

.recordings-table { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.recordings-table table { width: 100%; border-collapse: collapse; }
.recordings-table th { text-align: left; padding: 12px 16px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); background: #f8fafc; border-bottom: 1px solid var(--border-color); }
.recordings-table td { padding: 12px 16px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.recordings-table tr:hover { background: #f8fafc; }
.recordings-table tr:last-child td { border-bottom: none; }
.name-cell { display: flex; align-items: center; gap: 8px; }
.file-icon { width: 18px; height: 18px; color: var(--primary-color); }
.file-name { font-family: monospace; font-weight: 600; }
.category-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; background: #f1f5f9; }
.mono { font-family: monospace; font-size: 12px; }
.desc-cell { color: var(--text-muted); font-size: 12px; max-width: 200px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.actions-cell { display: flex; gap: 4px; }
.actions-cell .btn-icon { width: 28px; height: 28px; background: white; border: 1px solid var(--border-color); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); }
.actions-cell .btn-icon:hover { color: var(--primary-color); border-color: var(--primary-color); }
.actions-cell .btn-icon.danger:hover { color: #ef4444; border-color: #ef4444; }
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
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.help-text { font-size: 10px; color: var(--text-muted); margin-top: 4px; display: block; }
.file-upload { border: 2px dashed var(--border-color); border-radius: 8px; padding: 24px; text-align: center; }
.file-upload input { display: none; }
.file-label { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.upload-icon { width: 24px; height: 24px; }

/* Record Modal */
.record-info { display: flex; flex-direction: column; gap: 12px; }
.record-step { display: flex; align-items: center; gap: 12px; padding: 10px; background: #f8fafc; border-radius: 6px; }
.step-num { width: 24px; height: 24px; background: var(--primary-color); color: white; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; }
.record-step code { background: #e2e8f0; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
.record-note { margin-top: 16px; padding: 12px; background: #fef3c7; border-radius: 6px; font-size: 12px; color: #92400e; }
.record-note code { background: rgba(0,0,0,0.1); padding: 2px 4px; border-radius: 3px; }
</style>
