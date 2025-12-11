<template>
  <div class="system-streams-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Streams & Recordings</h2>
        <p class="text-muted text-sm">Manage music-on-hold streams and global audio resources.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="showAddModal = true">
          <PlusIcon class="btn-icon" /> Add Stream
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'moh' }" @click="activeTab = 'moh'">
        <MusicIcon class="tab-icon" /> Music On Hold
      </button>
      <button class="tab" :class="{ active: activeTab === 'recordings' }" @click="activeTab = 'recordings'">
        <MicIcon class="tab-icon" /> System Recordings
      </button>
      <button class="tab" :class="{ active: activeTab === 'tts' }" @click="activeTab = 'tts'">
        <VolumeIcon class="tab-icon" /> Text-to-Speech
      </button>
    </div>

    <!-- MUSIC ON HOLD TAB -->
    <div class="tab-content" v-if="activeTab === 'moh'">
      <p class="panel-desc">Configure music-on-hold streams that tenants can use.</p>

      <div class="streams-grid">
        <div class="stream-card" v-for="stream in mohStreams" :key="stream.id" :class="{ default: stream.isDefault }">
          <div class="stream-header">
            <div class="stream-icon" :class="stream.type">
              <MusicIcon v-if="stream.type === 'playlist'" class="icon" />
              <RadioIcon v-else-if="stream.type === 'shoutcast'" class="icon" />
              <FileAudioIcon v-else class="icon" />
            </div>
            <div class="stream-info">
              <h4>{{ stream.name }}</h4>
              <span class="stream-type">{{ stream.type }}</span>
            </div>
            <span class="default-badge" v-if="stream.isDefault">Default</span>
          </div>

          <div class="stream-details">
            <div class="detail-row" v-if="stream.type === 'shoutcast'">
              <span class="label">URL</span>
              <span class="value mono">{{ stream.url }}</span>
            </div>
            <div class="detail-row" v-else>
              <span class="label">Tracks</span>
              <span class="value">{{ stream.tracks }} files</span>
            </div>
            <div class="detail-row">
              <span class="label">Tenants Using</span>
              <span class="value">{{ stream.tenantCount }}</span>
            </div>
          </div>

          <div class="stream-actions">
            <button class="action-btn" @click="previewStream(stream)" title="Preview">
              <PlayIcon class="icon-sm" />
            </button>
            <button class="action-btn" @click="editStream(stream)" title="Edit">
              <EditIcon class="icon-sm" />
            </button>
            <button class="action-btn danger" @click="deleteStream(stream)" title="Delete" v-if="!stream.isDefault">
              <TrashIcon class="icon-sm" />
            </button>
          </div>
        </div>

        <div class="stream-card add-new" @click="showAddModal = true">
          <PlusCircleIcon class="add-icon" />
          <span>Add MOH Stream</span>
        </div>
      </div>
    </div>

    <!-- SYSTEM RECORDINGS TAB -->
    <div class="tab-content" v-else-if="activeTab === 'recordings'">
      <p class="panel-desc">Global system recordings available to all tenants.</p>

      <div class="recordings-table">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Category</th>
              <th>Duration</th>
              <th>Format</th>
              <th>Used By</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rec in recordings" :key="rec.id">
              <td class="name-cell">
                <FileAudioIcon class="file-icon" />
                <span>{{ rec.name }}</span>
              </td>
              <td><span class="category-badge">{{ rec.category }}</span></td>
              <td>{{ rec.duration }}</td>
              <td class="mono">{{ rec.format }}</td>
              <td>{{ rec.usedBy }} tenants</td>
              <td class="actions-cell">
                <button class="action-btn" @click="playRecording(rec)"><PlayIcon class="icon-sm" /></button>
                <button class="action-btn" @click="downloadRecording(rec)"><DownloadIcon class="icon-sm" /></button>
                <button class="action-btn danger" @click="deleteRecording(rec)"><TrashIcon class="icon-sm" /></button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <button class="btn-secondary" style="margin-top: 16px;" @click="showUploadModal = true">
        <UploadIcon class="btn-icon" /> Upload Recording
      </button>
    </div>

    <!-- TTS TAB -->
    <div class="tab-content" v-else-if="activeTab === 'tts'">
      <p class="panel-desc">Configure text-to-speech engine settings.</p>

      <div class="tts-config">
        <div class="setting-card">
          <h4>TTS Provider</h4>
          <select v-model="ttsConfig.provider" class="input-field">
            <option value="google">Google Cloud TTS</option>
            <option value="azure">Azure Cognitive Services</option>
            <option value="amazon">Amazon Polly</option>
            <option value="flite">FliteVoice (Local)</option>
          </select>
        </div>

        <div class="setting-card" v-if="ttsConfig.provider !== 'flite'">
          <h4>API Key</h4>
          <input v-model="ttsConfig.apiKey" type="password" class="input-field">
        </div>

        <div class="setting-card">
          <h4>Default Voice</h4>
          <select v-model="ttsConfig.voice" class="input-field">
            <option value="en-US-Wavenet-D">English (US) - Male</option>
            <option value="en-US-Wavenet-F">English (US) - Female</option>
            <option value="en-GB-Wavenet-B">English (UK) - Male</option>
            <option value="es-ES-Wavenet-B">Spanish - Male</option>
          </select>
        </div>

        <div class="setting-card">
          <h4>Test TTS</h4>
          <div class="tts-test">
            <input v-model="ttsTestText" class="input-field" placeholder="Enter text to test...">
            <button class="btn-secondary" @click="testTts">
              <PlayIcon class="btn-icon" /> Preview
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ADD STREAM MODAL -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Add MOH Stream</h3>
          <button class="btn-icon" @click="showAddModal = false"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Stream Name</label>
            <input v-model="streamForm.name" class="input-field" placeholder="Corporate Music">
          </div>
          <div class="form-group">
            <label>Type</label>
            <select v-model="streamForm.type" class="input-field">
              <option value="playlist">File Playlist</option>
              <option value="shoutcast">Shoutcast/Icecast URL</option>
              <option value="single">Single File (Loop)</option>
            </select>
          </div>
          <div class="form-group" v-if="streamForm.type === 'shoutcast'">
            <label>Stream URL</label>
            <input v-model="streamForm.url" class="input-field" placeholder="http://stream.example.com:8000/moh">
          </div>
          <div class="form-group" v-else>
            <label>Audio Files</label>
            <div class="file-upload-area">
              <UploadIcon class="upload-icon" />
              <span>Drop files here or click to upload</span>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showAddModal = false">Cancel</button>
          <button class="btn-primary" @click="saveStream">Add Stream</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import {
  Plus as PlusIcon, PlusCircle as PlusCircleIcon, Music as MusicIcon,
  Mic as MicIcon, Volume2 as VolumeIcon, Radio as RadioIcon,
  FileAudio as FileAudioIcon, Play as PlayIcon, Edit as EditIcon,
  Trash2 as TrashIcon, Upload as UploadIcon, Download as DownloadIcon,
  X as XIcon
} from 'lucide-vue-next'

const activeTab = ref('moh')
const showAddModal = ref(false)
const showUploadModal = ref(false)

const streamForm = ref({ name: '', type: 'playlist', url: '' })

const mohStreams = ref([
  { id: 1, name: 'Default Hold Music', type: 'playlist', tracks: 12, tenantCount: 15, isDefault: true },
  { id: 2, name: 'Jazz Collection', type: 'playlist', tracks: 8, tenantCount: 3, isDefault: false },
  { id: 3, name: 'Classical Radio', type: 'shoutcast', url: 'http://stream.example.com:8000/classical', tenantCount: 2, isDefault: false },
])

const recordings = ref([
  { id: 1, name: 'default-ivr-greeting', category: 'IVR', duration: '0:15', format: 'WAV', usedBy: 8 },
  { id: 2, name: 'voicemail-unavailable', category: 'Voicemail', duration: '0:08', format: 'WAV', usedBy: 12 },
  { id: 3, name: 'transfer-connecting', category: 'System', duration: '0:04', format: 'WAV', usedBy: 15 },
  { id: 4, name: 'queue-position-announcement', category: 'Queue', duration: '0:10', format: 'WAV', usedBy: 5 },
])

const ttsConfig = ref({
  provider: 'google',
  apiKey: '',
  voice: 'en-US-Wavenet-D'
})
const ttsTestText = ref('')

const previewStream = (stream) => alert(`Playing: ${stream.name}`)
const editStream = (stream) => alert(`Edit: ${stream.name}`)
const deleteStream = (stream) => { mohStreams.value = mohStreams.value.filter(s => s.id !== stream.id) }
const saveStream = () => { showAddModal.value = false }

const playRecording = (rec) => alert(`Playing: ${rec.name}`)
const downloadRecording = (rec) => alert(`Downloading: ${rec.name}`)
const deleteRecording = (rec) => { recordings.value = recordings.value.filter(r => r.id !== rec.id) }

const testTts = () => alert(`TTS Preview: "${ttsTestText.value}"`)
</script>

<style scoped>
.system-streams-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 16px; height: 16px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 24px; border-radius: 0 0 var(--radius-md) var(--radius-md); min-height: 300px; }

.panel-desc { color: var(--text-muted); font-size: 13px; margin-bottom: 20px; }

/* Streams Grid */
.streams-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }

.stream-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; transition: all 0.15s; }
.stream-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.stream-card.default { border-color: var(--primary-color); }

.stream-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.stream-icon { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.stream-icon.playlist { background: #dbeafe; color: #2563eb; }
.stream-icon.shoutcast { background: #f3e8ff; color: #7c3aed; }
.stream-icon.single { background: #dcfce7; color: #16a34a; }
.stream-info { flex: 1; }
.stream-info h4 { margin: 0 0 2px; font-size: 14px; }
.stream-type { font-size: 11px; color: var(--text-muted); text-transform: capitalize; }
.default-badge { font-size: 10px; font-weight: 700; padding: 3px 8px; border-radius: 4px; background: var(--primary-light); color: var(--primary-color); text-transform: uppercase; }

.stream-details { margin-bottom: 12px; }
.detail-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 12px; border-bottom: 1px solid var(--border-color); }
.detail-row:last-child { border-bottom: none; }
.detail-row .label { color: var(--text-muted); }
.detail-row .value { font-weight: 600; }
.mono { font-family: monospace; font-size: 11px; }

.stream-actions { display: flex; gap: 6px; }
.action-btn { width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }

.stream-card.add-new { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 180px; border: 2px dashed var(--border-color); cursor: pointer; color: var(--text-muted); gap: 12px; }
.stream-card.add-new:hover { border-color: var(--primary-color); color: var(--primary-color); }
.add-icon { width: 36px; height: 36px; }

/* Recordings Table */
.recordings-table { overflow-x: auto; }
.recordings-table table { width: 100%; border-collapse: collapse; }
.recordings-table th { text-align: left; padding: 12px; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 2px solid var(--border-color); }
.recordings-table td { padding: 12px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.recordings-table tr:hover { background: var(--bg-app); }

.name-cell { display: flex; align-items: center; gap: 8px; font-weight: 600; }
.file-icon { width: 18px; height: 18px; color: var(--primary-color); }
.category-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; background: var(--bg-app); color: var(--text-muted); }
.actions-cell { display: flex; gap: 4px; }

/* TTS Config */
.tts-config { display: flex; flex-direction: column; gap: 16px; max-width: 500px; }
.setting-card { padding: 16px; background: var(--bg-app); border-radius: var(--radius-md); }
.setting-card h4 { margin: 0 0 8px; font-size: 13px; font-weight: 600; }
.tts-test { display: flex; gap: 8px; }
.tts-test .input-field { flex: 1; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 480px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }

.file-upload-area { padding: 32px; border: 2px dashed var(--border-color); border-radius: var(--radius-md); display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.file-upload-area:hover { border-color: var(--primary-color); color: var(--primary-color); }
.upload-icon { width: 32px; height: 32px; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { width: 16px; height: 16px; }

.icon { width: 20px; height: 20px; }
.icon-sm { width: 16px; height: 16px; }

@media (max-width: 768px) {
  .streams-grid { grid-template-columns: 1fr; }
}
</style>
