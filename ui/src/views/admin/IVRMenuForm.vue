<template>
  <div class="ivr-editor">
    <!-- Header -->
    <div class="editor-header">
      <div class="header-left">
        <button class="back-link" @click="$router.push('/admin/ivr')">← Back to IVR</button>
        <div class="menu-info">
          <input v-model="form.name" class="menu-name-input" placeholder="Menu Name">
          <div class="menu-meta">
            <span class="meta-label">Ext:</span>
            <input v-model="form.extension" class="ext-input" placeholder="8000">
            <label class="enabled-toggle">
              <input type="checkbox" v-model="form.enabled">
              <span>Enabled</span>
            </label>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showSettings = true">
          <SettingsIcon class="btn-icon-left" />
          Settings
        </button>
        <button class="btn-secondary" @click="testIVR">
          <PlayIcon class="btn-icon-left" />
          Test
        </button>
        <button class="btn-primary" @click="saveMenu">Save Menu</button>
      </div>
    </div>

    <!-- Visual Editor Layout -->
    <div class="visual-layout">
      <!-- Node Palette -->
      <NodePalette />
      
      <!-- Canvas Area -->
      <div class="canvas-area">
        <div class="canvas-toolbar">
          <div class="zoom-controls">
            <button class="tool-btn" @click="zoomOut" title="Zoom Out">−</button>
            <span class="zoom-level">{{ Math.round(zoom * 100) }}%</span>
            <button class="tool-btn" @click="zoomIn" title="Zoom In">+</button>
            <button class="tool-btn" @click="zoom = 1" title="Reset Zoom">⟲</button>
          </div>
          <div class="toolbar-spacer"></div>
          <button class="tool-btn" @click="clearCanvas" title="Clear All">🗑</button>
        </div>
        <FlowCanvas v-model="flowData" :zoom="zoom" />
      </div>

      <!-- Properties Panel -->
      <div class="properties-panel">
        <h4>IVR Properties</h4>
        
        <div class="prop-section">
          <label>Greet Long</label>
          <select v-model="form.greetLong" class="prop-select">
            <option value="">Select audio...</option>
            <option v-for="r in recordings" :key="r.id" :value="r.file">{{ r.name }}</option>
          </select>
          <span class="help-text">Played when entering the menu</span>
        </div>

        <div class="prop-section">
          <label>Greet Short</label>
          <select v-model="form.greetShort" class="prop-select">
            <option value="">Same as Greet Long</option>
            <option v-for="r in recordings" :key="r.id" :value="r.file">{{ r.name }}</option>
          </select>
          <span class="help-text">Played on menu repeat</span>
        </div>

        <div class="prop-section">
          <label>Language</label>
          <select v-model="form.language" class="prop-select">
            <option v-for="lang in FLOW_ENUMS.languages" :key="lang.value" :value="lang.value">{{ lang.label }}</option>
          </select>
        </div>

        <div class="prop-section">
          <label>Description</label>
          <textarea v-model="form.description" class="prop-textarea" rows="2" placeholder="Optional notes..."></textarea>
        </div>

        <div class="prop-divider"></div>
        <h5>Advanced</h5>

        <div class="prop-section">
          <label>Caller ID Prefix</label>
          <input v-model="form.cidPrefix" class="prop-input" placeholder="IVR:">
          <span class="help-text">Added when transferring calls</span>
        </div>

        <div class="prop-section">
          <label>Ring Back Tone</label>
          <select v-model="form.ringBack" class="prop-select">
            <option v-for="tone in FLOW_ENUMS.ringbackTones" :key="tone.value" :value="tone.value">{{ tone.label }}</option>
          </select>
        </div>

        <div class="prop-section">
          <label class="checkbox-label">
            <input type="checkbox" v-model="form.pinProtected">
            PIN Protection
          </label>
          <input v-if="form.pinProtected" v-model="form.pin" type="password" class="prop-input" placeholder="Enter PIN">
        </div>
      </div>
    </div>

    <!-- Settings Modal -->
    <div class="modal-overlay" v-if="showSettings" @click.self="showSettings = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Advanced Settings</h3>
          <button class="close-btn" @click="showSettings = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group">
              <label>Language</label>
              <select v-model="form.language" class="input-field">
                <option v-for="lang in FLOW_ENUMS.languages" :key="lang.value" :value="lang.value">{{ lang.label }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>Ring Back</label>
              <select v-model="form.ringBack" class="input-field">
                <option v-for="tone in FLOW_ENUMS.ringbackTones" :key="tone.value" :value="tone.value">{{ tone.label }}</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>Caller ID Prefix</label>
            <input v-model="form.cidPrefix" class="input-field" placeholder="IVR:">
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>PIN Protection</label>
              <input v-model="form.pin" class="input-field" type="password" placeholder="Leave empty for none">
            </div>
            <div class="form-group">
              <label>Confirm Key</label>
              <select v-model="form.confirmKey" class="input-field">
                <option v-for="term in FLOW_ENUMS.terminators" :key="term.value" :value="term.value">{{ term.label }}</option>
              </select>
            </div>
          </div>

          <h4>Text-to-Speech</h4>
          <div class="form-row">
            <div class="form-group">
              <label>TTS Engine</label>
              <select v-model="form.ttsEngine" class="input-field">
                <option v-for="engine in FLOW_ENUMS.ttsEngines" :key="engine.value" :value="engine.value">{{ engine.label }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>TTS Voice</label>
              <select v-model="form.ttsVoice" class="input-field">
                <option v-for="voice in FLOW_ENUMS.ttsVoices" :key="voice.value" :value="voice.value">{{ voice.label }}</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>Description</label>
            <textarea v-model="form.description" class="input-field" rows="2" placeholder="Optional description"></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showSettings = false">Cancel</button>
          <button class="btn-primary" @click="showSettings = false">Apply</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div class="modal-overlay" v-if="showDeleteModal">
      <div class="modal-card small">
        <h3>Delete IVR Menu?</h3>
        <p>This may break call flows using this menu.</p>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger" @click="confirmDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, provide, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Play as PlayIcon, Settings as SettingsIcon } from 'lucide-vue-next'
import { ivrAPI, recordingsAPI, queuesAPI, extensionsAPI, ringGroupsAPI, voicemailAPI } from '../../services/api.js'
import NodePalette from '../../components/flow/NodePalette.vue'
import FlowCanvas from '../../components/flow/FlowCanvas.vue'

const toast = inject('toast')

// Flow node enums - ideally from API config
const FLOW_ENUMS = {
  languages: [
    { value: 'en-US', label: 'English (US)' },
    { value: 'en-GB', label: 'English (UK)' },
    { value: 'es-ES', label: 'Spanish (Spain)' },
    { value: 'fr-FR', label: 'French (France)' },
    { value: 'de-DE', label: 'German (Germany)' },
  ],
  ringbackTones: [
    { value: 'default', label: 'System Default' },
    { value: 'us-ring', label: 'US Ring' },
    { value: 'uk-ring', label: 'UK Ring' },
    { value: 'music', label: 'Music' },
  ],
  ttsEngines: [
    { value: 'flite', label: 'Flite (Local)' },
    { value: 'google', label: 'Google Cloud TTS' },
    { value: 'aws', label: 'AWS Polly' },
    { value: 'azure', label: 'Azure Speech' },
  ],
  ttsVoices: [
    { value: 'default', label: 'Default' },
    { value: 'male', label: 'Male' },
    { value: 'female', label: 'Female' },
  ],
  terminators: [
    { value: '#', label: '# key' },
    { value: '*', label: '* key' },
    { value: '', label: 'None (timeout only)' },
  ]
}

const route = useRoute()
const router = useRouter()

const isNew = computed(() => route.params.id === 'new')
const loading = ref(false)
const saving = ref(false)

const zoom = ref(1)
const showSettings = ref(false)
const showDeleteModal = ref(false)

const form = ref({
  name: '',
  extension: '',
  enabled: true,
  language: 'en-US',
  description: '',
  greetLong: '',
  greetShort: '',
  directDial: false,
  digitLength: 4,
  interDigitTimeout: 2000,
  timeout: 3000,
  maxTimeouts: 3,
  maxFailures: 3,
  exitAction: { type: 'hangup', value: '' },
  invalidSound: '',
  exitSound: '',
  ringBack: 'default',
  cidPrefix: '',
  pin: '',
  pinProtected: false,
  confirmKey: '#',
  ttsEngine: 'flite',
  ttsVoice: 'default'
})

const flowData = ref({
  nodes: [
    { id: 'start', type: 'ivr_start', label: 'IVR Start', x: 50, y: 50, config: {} }
  ],
  connections: []
})

// Dynamic data for dropdowns (loaded from API)
const recordings = ref([])
const queuesList = ref([])
const extensionsList = ref([])
const ivrMenusList = ref([])
const ringGroupsList = ref([])
const voicemailBoxesList = ref([])

// Provide dynamic data to child flow components
provide('flowExtensions', extensionsList)
provide('flowQueues', queuesList)
provide('flowIVRMenus', ivrMenusList)
provide('flowRingGroups', ringGroupsList)
provide('flowRecordings', recordings)
provide('flowVoicemailBoxes', voicemailBoxesList)

const zoomIn = () => { zoom.value = Math.min(2, zoom.value + 0.1) }
const zoomOut = () => { zoom.value = Math.max(0.5, zoom.value - 0.1) }

const clearCanvas = () => {
  if (confirm('Clear all nodes?')) {
    flowData.value = { nodes: [], connections: [] }
  }
}

const testIVR = async () => {
  if (!form.value.extension) {
    toast?.warning('Please set an extension number before testing')
    return
  }
  try {
    const response = await ivrAPI.testMenu ? 
      ivrAPI.testMenu(route.params.id) :
      ivrAPI.callMenu ? 
        ivrAPI.callMenu({ extension: form.value.extension }) :
        Promise.reject(new Error('Test not available'))
    toast?.success(`Testing IVR: ${form.value.name || 'Untitled'} - calling extension ${form.value.extension}`)
  } catch (e) {
    console.error('Failed to test IVR:', e)
    toast?.error(e.message || 'Failed to test IVR')
  }
}

// Build the API payload from form + flow data
const buildPayload = () => ({
  name: form.value.name,
  extension: form.value.extension,
  enabled: form.value.enabled,
  greet_long: form.value.greetLong,
  greet_short: form.value.greetShort,
  invalid_sound: form.value.invalidSound,
  exit_sound: form.value.exitSound,
  transfer_sound: '',
  timeout: Math.floor(form.value.timeout / 1000) || 10,
  max_failures: form.value.maxFailures,
  max_timeouts: form.value.maxTimeouts,
  digit_len: form.value.digitLength,
  inter_digit_time: form.value.interDigitTimeout,
  direct_dial: form.value.directDial,
  ringback: form.value.ringBack,
  caller_id_prefix: form.value.cidPrefix,
  flow_data: flowData.value
})

// Save menu to backend
const saveMenu = async () => {
  if (!form.value.name) {
    toast?.warning('Please enter a menu name.')
    return
  }
  saving.value = true
  try {
    const payload = buildPayload()
    if (isNew.value) {
      await ivrAPI.createMenu(payload)
    } else {
      await ivrAPI.updateMenu(route.params.id, payload)
    }
    router.push('/admin/ivr')
  } catch (err) {
    console.error('Failed to save IVR menu:', err)
    toast?.error(err.message || 'Failed to save IVR menu')
  } finally {
    saving.value = false
  }
}

// Load existing menu for editing
const loadMenu = async () => {
  if (isNew.value) return
  loading.value = true
  try {
    const { data } = await ivrAPI.getMenu(route.params.id)
    const menu = data.data
    form.value.name = menu.name || ''
    form.value.extension = menu.extension || ''
    form.value.enabled = menu.enabled !== false
    form.value.greetLong = menu.greet_long || ''
    form.value.greetShort = menu.greet_short || ''
    form.value.invalidSound = menu.invalid_sound || ''
    form.value.exitSound = menu.exit_sound || ''
    form.value.timeout = (menu.timeout || 10) * 1000
    form.value.maxFailures = menu.max_failures || 3
    form.value.maxTimeouts = menu.max_timeouts || 3
    form.value.digitLength = menu.digit_len || 4
    form.value.interDigitTimeout = menu.inter_digit_time || 2000
    form.value.directDial = menu.direct_dial || false
    form.value.ringBack = menu.ringback || 'default'
    form.value.cidPrefix = menu.caller_id_prefix || ''

    // Load flow data if present
    if (menu.flow_data && menu.flow_data.nodes && menu.flow_data.nodes.length > 0) {
      flowData.value = menu.flow_data
    }
  } catch (err) {
    console.error('Failed to load IVR menu:', err)
  } finally {
    loading.value = false
  }
}

// Load dynamic data for dropdowns
const loadDropdownData = async () => {
  try {
    const [recRes, queueRes, extRes, ivrRes, rgRes, vmRes] = await Promise.allSettled([
      recordingsAPI.list(),
      queuesAPI.list(),
      extensionsAPI.list(),
      ivrAPI.listMenus(),
      ringGroupsAPI.list(),
      voicemailAPI.listBoxes()
    ])
    if (recRes.status === 'fulfilled') {
      recordings.value = (recRes.value.data.data || []).map(r => ({
        id: r.id, name: r.name || r.file_name, file: r.file_path || r.file_name
      }))
    }
    if (queueRes.status === 'fulfilled') {
      queuesList.value = queueRes.value.data.data || []
    }
    if (extRes.status === 'fulfilled') {
      extensionsList.value = extRes.value.data.data || []
    }
    if (ivrRes.status === 'fulfilled') {
      ivrMenusList.value = (ivrRes.value.data.data || []).filter(m => String(m.id) !== route.params.id)
    }
    if (rgRes.status === 'fulfilled') {
      ringGroupsList.value = rgRes.value.data.data || []
    }
    if (vmRes.status === 'fulfilled') {
      voicemailBoxesList.value = vmRes.value.data.data || []
    }
  } catch (err) {
    console.error('Failed to load dropdown data:', err)
  }
}

const confirmDelete = async () => {
  try {
    if (!isNew.value) {
      await ivrAPI.deleteMenu(route.params.id)
    }
  } catch (err) {
    console.error('Delete failed:', err)
  }
  showDeleteModal.value = false
  router.push('/admin/ivr')
}

onMounted(() => {
  loadMenu()
  loadDropdownData()
})

// Expose dropdown data for FlowNode components to consume via provide/inject later
</script>

<style scoped>
.ivr-editor {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
  background: #f8fafc;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: white;
  border-bottom: 1px solid var(--border-color);
}

.header-left { display: flex; align-items: center; gap: 16px; }
.back-link { background: none; border: none; color: var(--primary-color); font-size: 13px; cursor: pointer; font-weight: 500; }

.menu-info { display: flex; flex-direction: column; gap: 4px; }
.menu-name-input { font-size: 16px; font-weight: 600; border: none; padding: 4px 8px; border-radius: 4px; background: transparent; width: 200px; }
.menu-name-input:hover { background: #f1f5f9; }
.menu-name-input:focus { background: white; outline: none; box-shadow: 0 0 0 2px var(--primary-color); }

.menu-meta { display: flex; align-items: center; gap: 8px; font-size: 12px; }
.meta-label { color: var(--text-muted); }
.ext-input { width: 60px; border: 1px solid transparent; padding: 2px 6px; border-radius: 4px; font-family: monospace; background: #f1f5f9; }
.ext-input:focus { border-color: var(--primary-color); outline: none; }

.enabled-toggle { display: flex; align-items: center; gap: 4px; font-size: 11px; color: var(--text-muted); cursor: pointer; }
.enabled-toggle input { margin: 0; }

.header-actions { display: flex; gap: 8px; }

/* Visual Layout */
.visual-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.canvas-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f1f5f9;
  overflow: hidden;
}

.canvas-toolbar {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: white;
  border-bottom: 1px solid var(--border-color);
  gap: 8px;
}

.zoom-controls { display: flex; align-items: center; gap: 4px; }
.zoom-level { font-size: 11px; font-weight: 600; color: var(--text-muted); min-width: 40px; text-align: center; }
.toolbar-spacer { flex: 1; }

.tool-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.tool-btn:hover { background: #f1f5f9; }

/* Properties Panel */
.properties-panel {
  width: 220px;
  background: white;
  border-left: 1px solid var(--border-color);
  padding: 12px;
  overflow-y: auto;
}

.properties-panel h4 {
  font-size: 12px;
  font-weight: 700;
  margin: 0 0 12px 0;
  color: var(--text-primary);
}

.prop-section { margin-bottom: 10px; }
.prop-section label { display: block; font-size: 10px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 4px; }
.prop-input, .prop-select, .prop-textarea {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 12px;
  box-sizing: border-box;
}
.prop-textarea { resize: vertical; font-family: inherit; min-height: 50px; }
.prop-input:focus, .prop-select:focus, .prop-textarea:focus { border-color: var(--primary-color); outline: none; }

.prop-divider { border-top: 1px solid var(--border-color); margin: 12px 0; }
.properties-panel h5 { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); margin: 0 0 10px; }
.help-text { font-size: 9px; color: var(--text-muted); margin-top: 2px; display: block; }

.checkbox-label { display: flex !important; align-items: center; gap: 6px; font-size: 11px !important; text-transform: none !important; cursor: pointer; }
.checkbox-label input { margin: 0; }

/* Buttons */
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: 6px; font-weight: 500; font-size: 13px; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-danger { background: #ef4444; color: white; border: none; padding: 8px 16px; border-radius: 6px; font-weight: 500; cursor: pointer; }
.btn-icon-left { width: 14px; height: 14px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 500px; max-height: 80vh; overflow: hidden; display: flex; flex-direction: column; }
.modal-card.small { max-width: 360px; padding: 20px; text-align: center; }
.modal-card.small h3 { margin: 0 0 8px; }
.modal-card.small p { margin: 0 0 20px; color: var(--text-muted); font-size: 13px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-row { display: flex; gap: 12px; }
.form-group { display: flex; flex-direction: column; gap: 4px; margin-bottom: 12px; flex: 1; }
.form-group label { font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; }
.input-field:focus { border-color: var(--primary-color); outline: none; }
.modal-body h4 { font-size: 13px; font-weight: 600; margin: 16px 0 12px; color: var(--text-primary); border-top: 1px solid var(--border-color); padding-top: 16px; }
</style>
