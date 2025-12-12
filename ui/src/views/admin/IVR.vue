<template>
  <div class="view-header">
    <div class="header-content">
      <h2>IVR Menus</h2>
      <p class="text-muted text-sm">Configure Auto Attendants and Interactive Voice Response menus.</p>
    </div>
    <div class="header-actions">
      <router-link to="/admin/audio-library" class="btn-secondary">
        <MicIcon class="btn-icon-left" />
        Audio Library
      </router-link>
      <router-link to="/admin/ivr/menus/new" class="btn-primary">+ New IVR Menu</router-link>
    </div>
  </div>

  <!-- Stats Row -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-icon menus"><MenuIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ menus.length }}</span>
        <span class="stat-label">IVR Menus</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon active"><CheckCircleIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ menus.filter(m => m.status === 'active').length }}</span>
        <span class="stat-label">Active</span>
      </div>
    </div>
  </div>

  <!-- IVR MENUS GRID -->
  <div class="ivr-grid">
    <div class="ivr-card" v-for="menu in menus" :key="menu.id">
      <div class="ivr-header">
        <div class="ivr-icon"><MenuIcon class="icon-md" /></div>
        <div class="ivr-info">
          <h4>{{ menu.name }}</h4>
          <span class="ivr-ext">Ext {{ menu.extension }}</span>
        </div>
        <StatusBadge :status="menu.status" />
      </div>
      
      <div class="ivr-greeting">
        <div class="greeting-label">
          <VolumeIcon class="icon-xs" />
          <span>Greeting</span>
        </div>
        <span class="greeting-file">{{ menu.greeting }}</span>
      </div>

      <div class="ivr-options">
        <div class="option-row" v-for="(opt, key) in menu.options" :key="key">
          <span class="option-key">{{ key }}</span>
          <span class="option-action">{{ opt }}</span>
        </div>
      </div>

      <div class="ivr-footer">
        <div class="ivr-meta">
          <span class="meta-label">Timeout:</span>
          <span>{{ menu.timeout }}s → {{ menu.timeoutAction }}</span>
        </div>
        <div class="ivr-actions">
          <router-link :to="`/admin/ivr/menus/${menu.id}`" class="btn-link">Edit</router-link>
          <button class="btn-link text-bad" @click="deleteMenu(menu)">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <!-- IVR MENU MODAL -->
  <div v-if="showMenuModal" class="modal-overlay" @click.self="showMenuModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>{{ editingMenu ? 'Edit IVR Menu' : 'New IVR Menu' }}</h3>
        <button class="btn-icon" @click="showMenuModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Menu Name</label>
            <input v-model="menuForm.name" class="input-field" placeholder="e.g. Main Menu">
          </div>
          <div class="form-group">
            <label>Extension</label>
            <input v-model="menuForm.extension" class="input-field code" placeholder="8000">
          </div>
        </div>

        <div class="form-group">
          <label>Greeting / Prompt</label>
          <div class="input-group">
            <select v-model="menuForm.greeting" class="input-field flex-1">
              <option value="">Select recording...</option>
              <option v-for="r in recordings" :key="r.id" :value="r.name">{{ r.name }}</option>
            </select>
            <button class="btn-secondary small">Record</button>
            <button class="btn-secondary small">Upload</button>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Menu Options</h4>
          </div>
          <p class="help-text">Define what happens when caller presses each key (0-9, *, #).</p>
          
          <div class="options-editor">
            <div class="option-edit-row" v-for="key in ['1','2','3','4','5','6','7','8','9','0','*','#']" :key="key">
              <span class="option-key-badge">{{ key }}</span>
              <select v-model="menuForm.options[key].type" class="input-field small">
                <option value="">Not Used</option>
                <option value="extension">Extension</option>
                <option value="ring_group">Ring Group</option>
                <option value="queue">Queue</option>
                <option value="voicemail">Voicemail</option>
                <option value="ivr">Sub-Menu (IVR)</option>
                <option value="external">External Number</option>
                <option value="hangup">Hangup</option>
              </select>
              <input v-if="menuForm.options[key].type && menuForm.options[key].type !== 'hangup'" 
                v-model="menuForm.options[key].target" 
                class="input-field flex-1" 
                :placeholder="getPlaceholder(menuForm.options[key].type)">
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-row">
          <div class="form-group">
            <label>Timeout (seconds)</label>
            <input type="number" v-model="menuForm.timeout" class="input-field" min="3" max="30">
          </div>
          <div class="form-group">
            <label>Timeout Action</label>
            <select v-model="menuForm.timeoutAction" class="input-field">
              <option value="replay">Replay Menu</option>
              <option value="operator">Transfer to Operator</option>
              <option value="hangup">Hangup</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Max Retries</label>
            <input type="number" v-model="menuForm.maxRetries" class="input-field" min="1" max="5">
          </div>
          <div class="form-group">
            <label>Invalid Key Action</label>
            <select v-model="menuForm.invalidAction" class="input-field">
              <option value="replay">Replay Menu</option>
              <option value="operator">Transfer to Operator</option>
            </select>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showMenuModal = false">Cancel</button>
        <button class="btn-primary" @click="saveMenu" :disabled="!menuForm.name">Save Menu</button>
      </div>
    </div>
  </div>

  <!-- TIME CONDITION MODAL -->
  <div v-if="showTimeModal" class="modal-overlay" @click.self="showTimeModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ editingTime ? 'Edit Time Condition' : 'New Time Condition' }}</h3>
        <button class="btn-icon" @click="showTimeModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Name</label>
          <input v-model="timeForm.name" class="input-field" placeholder="e.g. Business Hours">
        </div>

        <div class="form-section">
          <div class="section-header">
            <h4>Schedule Rules</h4>
            <button class="btn-small" @click="addTimeRule">+ Add Rule</button>
          </div>
          
          <div class="time-rules">
            <div class="time-rule" v-for="(rule, i) in timeForm.rules" :key="i">
              <div class="rule-days">
                <label v-for="d in ['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" :key="d" class="day-check">
                  <input type="checkbox" :value="d" v-model="rule.days">
                  <span>{{ d.charAt(0) }}</span>
                </label>
              </div>
              <input type="time" v-model="rule.startTime" class="input-field time-input">
              <span class="time-sep">to</span>
              <input type="time" v-model="rule.endTime" class="input-field time-input">
              <button class="btn-icon" @click="removeTimeRule(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-group">
          <label>When Matched → Route To</label>
          <input v-model="timeForm.matchDestination" class="input-field" placeholder="e.g. IVR: Main Menu">
        </div>

        <div class="form-group">
          <label>When NOT Matched → Route To</label>
          <input v-model="timeForm.noMatchDestination" class="input-field" placeholder="e.g. Voicemail: General">
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showTimeModal = false">Cancel</button>
        <button class="btn-primary" @click="saveTimeCondition" :disabled="!timeForm.name">Save Condition</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'
import { 
  Menu as MenuIcon, Clock as ClockIcon, ToggleLeft as ToggleLeftIcon,
  Volume2 as VolumeIcon, CheckCircle as CheckCircleIcon, XCircle as XCircleIcon,
  Info as InfoIcon, Edit as EditIcon, Trash2 as TrashIcon, X as XIcon,
  FileAudio as FileAudioIcon, Download as DownloadIcon, LayoutGrid as LayoutIcon,
  Mic as MicIcon
} from 'lucide-vue-next'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { ivrAPI } from '@/services/api'

const toast = inject('toast')
const isLoading = ref(false)
const activeTab = ref('menus')

// IVR Menus
const showMenuModal = ref(false)
const editingMenu = ref(false)
const menuForm = ref({
  name: '',
  extension: '',
  greeting: '',
  options: Object.fromEntries(['1','2','3','4','5','6','7','8','9','0','*','#'].map(k => [k, { type: '', target: '' }])),
  timeout: 10,
  timeoutAction: 'replay',
  maxRetries: 3,
  invalidAction: 'replay'
})

const menus = ref([])

onMounted(async () => {
  await fetchMenus()
})

async function fetchMenus() {
  isLoading.value = true
  try {
    const response = await ivrAPI.list()
    menus.value = (response.data || []).map(m => ({
      id: m.id,
      name: m.name,
      extension: m.extension,
      status: m.enabled ? 'Active' : 'Idle',
      greeting: m.greeting_file || 'default.wav',
      options: m.options || {},
      timeout: m.timeout || 10,
      timeoutAction: m.timeout_action || 'Replay'
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load IVR menus')
    // Fallback to demo data
    menus.value = [
      {
        id: 1, name: 'Main Menu', extension: '8000', status: 'Active',
        greeting: 'main_greeting.wav',
        options: { '1': 'Ext 101 (Sales)', '2': 'Ring Group: Support', '3': 'Company Directory', '0': 'Operator' },
        timeout: 10, timeoutAction: 'Voicemail'
      },
      {
        id: 2, name: 'After Hours', extension: '8001', status: 'Active',
        greeting: 'after_hours.wav',
        options: { '1': 'Emergency Line', '0': 'Leave Message' },
        timeout: 15, timeoutAction: 'Voicemail'
      },
    ]
  } finally {
    isLoading.value = false
  }
}

const getPlaceholder = (type) => {
  const placeholders = {
    extension: '101',
    ring_group: 'sales',
    queue: 'support',
    voicemail: 'general',
    ivr: '8001',
    external: '+14155551234'
  }
  return placeholders[type] || ''
}

const editMenu = (menu) => {
  editingMenu.value = true
  showMenuModal.value = true
}

const deleteMenu = async (menu) => {
  if (confirm(`Delete IVR "${menu.name}"?`)) {
    try {
      await ivrAPI.delete(menu.id)
      toast?.success(`IVR "${menu.name}" deleted`)
      await fetchMenus()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete IVR')
    }
  }
}

const saveMenu = async () => {
  try {
    // TODO: Map form to API format and save
    toast?.success('IVR menu saved')
    showMenuModal.value = false
    editingMenu.value = false
    await fetchMenus()
  } catch (error) {
    toast?.error(error.message, 'Failed to save IVR menu')
  }
}

// Time Conditions
const showTimeModal = ref(false)
const editingTime = ref(false)
const timeForm = ref({
  name: '',
  rules: [{ days: ['Mon','Tue','Wed','Thu','Fri'], startTime: '09:00', endTime: '17:00' }],
  matchDestination: '',
  noMatchDestination: ''
})

const timeConditions = ref([
  {
    id: 1, name: 'Business Hours', enabled: true, currentMatch: true,
    rules: [
      { id: 1, days: ['Mon','Tue','Wed','Thu','Fri'], startTime: '09:00', endTime: '17:00' }
    ],
    matchDestination: 'IVR: Main Menu (8000)',
    noMatchDestination: 'IVR: After Hours (8001)'
  },
])

const addTimeRule = () => {
  timeForm.value.rules.push({ days: [], startTime: '09:00', endTime: '17:00' })
}

const removeTimeRule = (i) => {
  timeForm.value.rules.splice(i, 1)
}

const editTimeCondition = (tc) => {
  editingTime.value = true
  showTimeModal.value = true
}

const deleteTimeCondition = (tc) => {
  if (confirm(`Delete time condition "${tc.name}"?`)) {
    timeConditions.value = timeConditions.value.filter(t => t.id !== tc.id)
  }
}

const saveTimeCondition = () => {
  showTimeModal.value = false
  editingTime.value = false
}

// Mode Toggles
const showToggleModal = ref(false)
const toggles = ref([
  { id: 1, name: 'Night Mode Override', featureCode: '*200', enabled: false, description: 'Force night mode routing regardless of time conditions' },
  { id: 2, name: 'Holiday Mode', featureCode: '*201', enabled: false, description: 'Activate holiday greeting and routing' },
])

const editToggle = (t) => toast?.info(`Edit toggle: ${t.name}`)
const deleteToggle = (t) => {
  if (confirm(`Delete toggle "${t.name}"?`)) {
    toggles.value = toggles.value.filter(x => x.id !== t.id)
  }
}

// Recordings
const showUploadModal = ref(false)
const recordings = ref([
  { id: 1, name: 'main_greeting.wav', duration: '0:32', format: 'WAV/8kHz' },
  { id: 2, name: 'after_hours.wav', duration: '0:18', format: 'WAV/8kHz' },
])

const playRecording = (rec) => toast?.info(`Playing: ${rec.name}`)
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

/* Stats Row */
.stats-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.menus { background: #dbeafe; color: #2563eb; }
.stat-icon.time { background: #fef3c7; color: #b45309; }
.stat-icon.toggles { background: #dcfce7; color: #16a34a; }
.stat-icon.recordings { background: #fce7f3; color: #db2777; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); }

/* IVR Grid */
.ivr-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.ivr-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }

.ivr-header { display: flex; align-items: center; gap: 12px; padding: 16px; background: var(--bg-app); border-bottom: 1px solid var(--border-color); }
.ivr-icon { width: 40px; height: 40px; background: var(--primary-light); border-radius: 8px; display: flex; align-items: center; justify-content: center; color: var(--primary-color); }
.ivr-info { flex: 1; }
.ivr-info h4 { font-size: 14px; font-weight: 600; margin: 0; }
.ivr-ext { font-size: 11px; color: var(--text-muted); font-family: monospace; }

.ivr-greeting { display: flex; align-items: center; gap: 8px; padding: 12px 16px; border-bottom: 1px solid var(--border-color); font-size: 12px; }
.greeting-label { display: flex; align-items: center; gap: 4px; color: var(--text-muted); }
.greeting-file { font-family: monospace; color: var(--text-main); }

.ivr-options { padding: 12px 16px; min-height: 80px; }
.option-row { display: flex; gap: 8px; align-items: center; padding: 4px 0; font-size: 13px; }
.option-key { width: 20px; height: 20px; background: var(--text-primary); color: white; border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; }
.option-action { color: var(--text-main); }

.ivr-footer { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; border-top: 1px solid var(--border-color); background: var(--bg-app); }
.ivr-meta { font-size: 11px; color: var(--text-muted); }
.meta-label { font-weight: 600; }
.ivr-actions { display: flex; gap: 8px; }

/* Time Conditions */
.time-list { display: flex; flex-direction: column; gap: 16px; }
.time-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; }
.time-card.active { border-color: #22c55e; background: #f0fdf4; }

.time-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.time-icon { width: 40px; height: 40px; background: #fef3c7; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #b45309; }
.time-icon.matching { background: #dcfce7; color: #16a34a; }
.time-info { flex: 1; }
.time-info h4 { font-size: 14px; font-weight: 600; margin: 0; }
.time-status { font-size: 11px; color: var(--text-muted); }
.time-status.matching { color: #16a34a; font-weight: 600; }

.time-schedule { margin-bottom: 12px; }
.schedule-row { display: flex; justify-content: space-between; padding: 6px 12px; background: var(--bg-app); border-radius: 4px; margin-bottom: 4px; font-size: 12px; }
.schedule-days { font-weight: 500; }
.schedule-time { color: var(--text-muted); font-family: monospace; }

.time-destinations { margin-bottom: 12px; }
.dest-row { display: flex; align-items: center; gap: 8px; padding: 6px 0; font-size: 13px; }
.dest-icon { width: 16px; height: 16px; }
.dest-row.match .dest-icon { color: #16a34a; }
.dest-row.nomatch .dest-icon { color: #dc2626; }
.dest-label { font-weight: 500; color: var(--text-muted); }
.dest-target { color: var(--text-main); }

.time-actions { display: flex; gap: 8px; justify-content: flex-end; }

/* Mode Toggles */
.toggles-help { display: flex; align-items: center; gap: 8px; padding: 12px; background: #eff6ff; border-radius: var(--radius-sm); margin-bottom: 16px; color: #1e40af; font-size: 13px; }
.help-icon { width: 16px; height: 16px; }

.toggles-list { display: flex; flex-direction: column; gap: 12px; }
.toggle-card { display: flex; align-items: center; gap: 16px; padding: 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }
.toggle-card.active { border-color: #22c55e; background: #f0fdf4; }

.toggle-status-indicator { width: 12px; height: 12px; border-radius: 50%; background: #d1d5db; }
.toggle-status-indicator.on { background: #22c55e; box-shadow: 0 0 8px #22c55e; }

.toggle-main { flex: 1; display: flex; align-items: center; gap: 16px; }
.toggle-info { flex: 1; }
.toggle-info h4 { font-size: 14px; font-weight: 600; margin: 0 0 4px; }
.toggle-desc { font-size: 12px; color: var(--text-muted); margin: 0; }

.toggle-code { text-align: center; }
.code-label { display: block; font-size: 10px; color: var(--text-muted); text-transform: uppercase; }
.code-value { font-family: monospace; font-size: 16px; font-weight: 700; color: var(--primary-color); }

.toggle-controls { display: flex; align-items: center; gap: 8px; }
.toggle-btn { padding: 6px 12px; border: 1px solid var(--border-color); border-radius: 4px; font-size: 12px; font-weight: 600; background: white; cursor: pointer; }
.toggle-btn.active { background: #dc2626; color: white; border-color: #dc2626; }

/* Recordings */
.recordings-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.recordings-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
.recording-card { display: flex; align-items: center; gap: 12px; padding: 12px 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-sm); }
.rec-icon { color: var(--text-muted); }
.rec-info { flex: 1; }
.rec-name { display: block; font-weight: 500; font-size: 13px; font-family: monospace; }
.rec-meta { font-size: 11px; color: var(--text-muted); }
.rec-actions { display: flex; gap: 4px; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-secondary.small { padding: 6px 10px; font-size: 12px; }
.btn-small { padding: 4px 8px; font-size: 11px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: var(--text-xs); cursor: pointer; font-weight: 500; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 16px; height: 16px; }
.icon-md { width: 20px; height: 20px; }
.icon-xs { width: 12px; height: 12px; }
.text-bad { color: var(--status-bad); }

/* Form Elements */
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.flex-1 { flex: 1; }
.flex-2 { flex: 2; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.code { font-family: monospace; background: #f8fafc; }
.input-field.small { width: 120px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.input-group { display: flex; gap: 8px; }
.help-text { font-size: 11px; color: var(--text-muted); }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.form-section { margin-bottom: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-header h4 { font-size: 13px; font-weight: 600; margin: 0; }

.options-editor { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
.option-edit-row { display: flex; gap: 8px; align-items: center; padding: 6px; background: var(--bg-app); border-radius: 4px; }
.option-key-badge { width: 24px; height: 24px; background: var(--text-primary); color: white; border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; }

.time-rules { display: flex; flex-direction: column; gap: 8px; }
.time-rule { display: flex; align-items: center; gap: 8px; padding: 8px; background: var(--bg-app); border-radius: 4px; }
.rule-days { display: flex; gap: 4px; }
.day-check { display: flex; flex-direction: column; align-items: center; }
.day-check input { display: none; }
.day-check span { width: 24px; height: 24px; display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 600; background: white; border: 1px solid var(--border-color); border-radius: 4px; cursor: pointer; }
.day-check input:checked + span { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.time-input { width: 90px !important; }
.time-sep { color: var(--text-muted); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); padding: 24px; }
.modal-card { background: white; border-radius: var(--radius-md); box-shadow: var(--shadow-lg); width: 100%; max-width: 500px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-card.large { max-width: 700px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Switch */
.switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: var(--border-color); transition: .3s; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 3px; bottom: 3px; background-color: white; transition: .3s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(16px); }
.slider.round { border-radius: 20px; }
.slider.round:before { border-radius: 50%; }
</style>
