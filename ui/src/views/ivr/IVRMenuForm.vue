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
            <option value="en-US">English (US)</option>
            <option value="en-GB">English (UK)</option>
            <option value="es-ES">Spanish</option>
            <option value="fr-FR">French</option>
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
            <option value="default">Default Ring</option>
            <option value="us-ring">US Ring</option>
            <option value="uk-ring">UK Ring</option>
            <option value="music">Music on Hold</option>
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
                <option value="en-US">English (US)</option>
                <option value="en-GB">English (UK)</option>
                <option value="es-ES">Spanish</option>
                <option value="fr-FR">French</option>
              </select>
            </div>
            <div class="form-group">
              <label>Ring Back</label>
              <select v-model="form.ringBack" class="input-field">
                <option value="default">Default Ring</option>
                <option value="us-ring">US Ring</option>
                <option value="uk-ring">UK Ring</option>
                <option value="music">Music on Hold</option>
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
                <option value="#"># key</option>
                <option value="*">* key</option>
              </select>
            </div>
          </div>

          <h4>Text-to-Speech</h4>
          <div class="form-row">
            <div class="form-group">
              <label>TTS Engine</label>
              <select v-model="form.ttsEngine" class="input-field">
                <option value="flite">Flite (Built-in)</option>
                <option value="google">Google Cloud TTS</option>
                <option value="aws">Amazon Polly</option>
                <option value="azure">Azure Cognitive</option>
              </select>
            </div>
            <div class="form-group">
              <label>TTS Voice</label>
              <select v-model="form.ttsVoice" class="input-field">
                <option value="default">Default</option>
                <option value="male">Male</option>
                <option value="female">Female</option>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Play as PlayIcon, Settings as SettingsIcon } from 'lucide-vue-next'
import NodePalette from '../../components/flow/NodePalette.vue'
import FlowCanvas from '../../components/flow/FlowCanvas.vue'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.params.id === 'new')

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

const recordings = ref([
  { id: 1, name: 'main_greeting.wav', file: 'main_greeting.wav' },
  { id: 2, name: 'after_hours.wav', file: 'after_hours.wav' },
  { id: 3, name: 'invalid_option.wav', file: 'invalid_option.wav' },
  { id: 4, name: 'goodbye.wav', file: 'goodbye.wav' }
])

const zoomIn = () => { zoom.value = Math.min(2, zoom.value + 0.1) }
const zoomOut = () => { zoom.value = Math.max(0.5, zoom.value - 0.1) }

const clearCanvas = () => {
  if (confirm('Clear all nodes?')) {
    flowData.value = { nodes: [], connections: [] }
  }
}

const testIVR = () => {
  alert('Testing IVR: ' + (form.value.name || 'Untitled'))
}

const saveMenu = () => {
  console.log('Saving menu:', { form: form.value, flow: flowData.value })
  alert('Menu saved!')
  router.push('/admin/ivr')
}

const confirmDelete = () => {
  showDeleteModal.value = false
  router.push('/admin/ivr')
}
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
