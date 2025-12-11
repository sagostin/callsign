<template>
  <div class="feature-codes-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Feature Codes</h2>
        <p class="text-muted text-sm">Manage star codes for system actions, telephony features, and custom integrations.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showImportModal = true">
          <UploadIcon class="btn-icon" /> Import
        </button>
        <button class="btn-primary" @click="showAddModal = true">
          <PlusIcon class="btn-icon" /> Add Feature Code
        </button>
      </div>
    </div>

    <!-- Speed Dial Prefix Settings -->
    <div class="speed-dial-settings">
      <div class="settings-card">
        <div class="settings-header">
          <div class="settings-title">
            <ZapIcon class="settings-icon" />
            <div>
              <h4>Speed Dial Settings</h4>
              <p>Configure the prefix pattern for speed dial functionality</p>
            </div>
          </div>
          <button class="btn-link" @click="showSpeedDialHelp = !showSpeedDialHelp">
            <HelpCircleIcon class="icon-sm" />
          </button>
        </div>
        <div class="settings-body">
          <div class="settings-form">
            <div class="form-group compact">
              <label>Speed Dial Prefix</label>
              <div class="prefix-input-group">
                <input v-model="speedDialSettings.prefix" class="input-field mono" placeholder="*0" style="width: 80px;" />
                <span class="help-text">+ digit (0-9) dials the speed dial entry</span>
              </div>
            </div>
            <div class="form-group compact">
              <label>Digit Range</label>
              <div class="range-inputs">
                <input v-model="speedDialSettings.startDigit" type="number" min="0" max="9" class="input-field mono" style="width: 60px;" />
                <span>to</span>
                <input v-model="speedDialSettings.endDigit" type="number" min="0" max="9" class="input-field mono" style="width: 60px;" />
              </div>
            </div>
            <div class="form-group compact">
              <label class="checkbox-label">
                <input type="checkbox" v-model="speedDialSettings.enableUserDefined" />
                <span>Allow users to define personal speed dials</span>
              </label>
            </div>
          </div>
          <div class="settings-preview">
            <span class="preview-label">Preview:</span>
            <code class="preview-code">{{ speedDialSettings.prefix }}1</code> → <code class="preview-code">{{ speedDialSettings.prefix }}{{ speedDialSettings.endDigit }}</code>
            <span class="preview-count">({{ speedDialSettings.endDigit - speedDialSettings.startDigit + 1 }} slots available)</span>
          </div>
        </div>
        <div class="help-panel" v-if="showSpeedDialHelp">
          <p><strong>How Speed Dials Work:</strong></p>
          <ul>
            <li>Users dial <code>{{ speedDialSettings.prefix }}X</code> where X is a digit (e.g., <code>{{ speedDialSettings.prefix }}1</code>)</li>
            <li>The system looks up the configured destination for that speed dial slot</li>
            <li>Speed dials can be configured globally by admins or per-user if allowed</li>
            <li>Manage speed dial entries in <router-link to="/admin/speed-dials">Speed Dials</router-link></li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Category Tabs -->
    <div class="category-tabs">
      <button 
        class="tab" 
        :class="{ active: activeCategory === 'all' }" 
        @click="activeCategory = 'all'"
      >
        All Codes
        <span class="tab-count">{{ featureCodes.length }}</span>
      </button>
      <button 
        class="tab" 
        :class="{ active: activeCategory === 'system' }" 
        @click="activeCategory = 'system'"
      >
        <SettingsIcon class="tab-icon" /> System
      </button>
      <button 
        class="tab" 
        :class="{ active: activeCategory === 'call-handling' }" 
        @click="activeCategory = 'call-handling'"
      >
        <PhoneIcon class="tab-icon" /> Call Handling
      </button>
      <button 
        class="tab" 
        :class="{ active: activeCategory === 'voicemail' }" 
        @click="activeCategory = 'voicemail'"
      >
        <VoicemailIcon class="tab-icon" /> Voicemail
      </button>
      <button 
        class="tab" 
        :class="{ active: activeCategory === 'queue' }" 
        @click="activeCategory = 'queue'"
      >
        <UsersIcon class="tab-icon" /> Queue/Agent
      </button>
      <button 
        class="tab" 
        :class="{ active: activeCategory === 'custom' }" 
        @click="activeCategory = 'custom'"
      >
        <CodeIcon class="tab-icon" /> Custom
      </button>
    </div>

    <!-- Search and Filter -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" class="search-input" placeholder="Search by code or name...">
      </div>
      <div class="filter-group">
        <select v-model="statusFilter" class="filter-select">
          <option value="">All Status</option>
          <option value="active">Active</option>
          <option value="disabled">Disabled</option>
        </select>
      </div>
    </div>

    <!-- Feature Codes Grid -->
    <div class="codes-grid">
      <div 
        class="code-card" 
        v-for="code in filteredCodes" 
        :key="code.id"
        :class="{ disabled: code.status === 'disabled' }"
      >
        <div class="code-header">
          <div class="code-badge">{{ code.code }}</div>
          <div class="code-status">
            <label class="toggle-sm">
              <input type="checkbox" :checked="code.status === 'active'" @change="toggleStatus(code)">
              <span class="toggle-slider-sm"></span>
            </label>
          </div>
        </div>
        
        <div class="code-body">
          <h4 class="code-name">{{ code.name }}</h4>
          <p class="code-description" v-if="code.description">{{ code.description }}</p>
          
          <div class="code-type">
            <component :is="getTypeIcon(code.type)" class="type-icon" />
            <span class="type-label" :class="code.type.toLowerCase()">{{ code.type }}</span>
          </div>

          <div class="code-target" v-if="code.target">
            <span class="target-label">Target:</span>
            <code class="target-value">{{ code.target }}</code>
          </div>
        </div>

        <div class="code-footer">
          <span class="usage-count" v-if="code.usageCount">
            <ActivityIcon class="usage-icon" /> {{ code.usageCount }} uses this month
          </span>
          <div class="code-actions">
            <button class="action-btn" @click="testCode(code)" title="Test">
              <PlayIcon class="icon-sm" />
            </button>
            <button class="action-btn" @click="editCode(code)" title="Edit">
              <EditIcon class="icon-sm" />
            </button>
            <button class="action-btn danger" @click="deleteCode(code)" title="Delete" v-if="!code.system">
              <TrashIcon class="icon-sm" />
            </button>
          </div>
        </div>
      </div>

      <!-- Add New Card -->
      <div class="code-card add-new" @click="showAddModal = true">
        <PlusCircleIcon class="add-icon" />
        <span>Add Feature Code</span>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ editingCode ? 'Edit Feature Code' : 'Add Feature Code' }}</h3>
          <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group" style="width: 120px;">
              <label>Code *</label>
              <input v-model="codeForm.code" class="input-field mono" placeholder="*XX">
            </div>
            <div class="form-group" style="flex: 1;">
              <label>Name *</label>
              <input v-model="codeForm.name" class="input-field" placeholder="Feature name">
            </div>
          </div>

          <div class="form-group">
            <label>Description</label>
            <input v-model="codeForm.description" class="input-field" placeholder="Brief description of what this code does">
          </div>

          <div class="form-group">
            <label>Category</label>
            <select v-model="codeForm.category" class="input-field">
              <option value="call-handling">Call Handling</option>
              <option value="voicemail">Voicemail</option>
              <option value="queue">Queue/Agent</option>
              <option value="system">System</option>
              <option value="custom">Custom</option>
            </select>
          </div>

          <div class="form-group">
            <label>Action Type *</label>
            <select v-model="codeForm.type" class="input-field">
              <option value="Internal">Internal (FreeSWITCH App)</option>
              <option value="Transfer">Transfer to Extension/Number</option>
              <option value="CURL">CURL/Webhook</option>
              <option value="Sound">Play Sound</option>
              <option value="Lua">Lua Script</option>
              <option value="API">Internal API Call</option>
            </select>
          </div>

          <div class="form-group" v-if="codeForm.type === 'Internal'">
            <label>Application</label>
            <select v-model="codeForm.target" class="input-field">
              <option value="app::voicemail">Access Voicemail</option>
              <option value="app::directory">Company Directory</option>
              <option value="app::agent_login">Agent Login</option>
              <option value="app::agent_logout">Agent Logout</option>
              <option value="app::park">Park Call</option>
              <option value="app::pickup">Pickup Parked Call</option>
              <option value="app::intercom">Intercom</option>
              <option value="app::page">Page Group</option>
              <option value="app::recording_start">Start Recording</option>
              <option value="app::recording_stop">Stop Recording</option>
              <option value="app::dnd_on">Enable DND</option>
              <option value="app::dnd_off">Disable DND</option>
              <option value="app::cf_set">Set Call Forward</option>
              <option value="app::cf_cancel">Cancel Call Forward</option>
            </select>
          </div>

          <div class="form-group" v-else-if="codeForm.type === 'Transfer'">
            <label>Destination</label>
            <input v-model="codeForm.target" class="input-field" placeholder="Extension or phone number">
          </div>

          <div class="form-group" v-else-if="codeForm.type === 'CURL'">
            <label>Webhook URL</label>
            <input v-model="codeForm.target" class="input-field" placeholder="https://api.example.com/webhook">
          </div>

          <div class="form-group" v-else-if="codeForm.type === 'Sound'">
            <label>Sound File</label>
            <select v-model="codeForm.target" class="input-field">
              <option value="local_stream://default">Default Hold Music</option>
              <option value="tone_stream://%(1000,0,350,440)">Dial Tone</option>
              <option value="ivr/ivr-menu.wav">IVR Menu Prompt</option>
            </select>
          </div>

          <div class="form-group" v-else-if="codeForm.type === 'Lua'">
            <label>Script Path</label>
            <input v-model="codeForm.target" class="input-field" placeholder="/scripts/custom.lua">
          </div>

          <div class="form-group" v-else>
            <label>Target Value</label>
            <input v-model="codeForm.target" class="input-field" placeholder="Target or parameters">
          </div>

          <div class="form-divider"></div>

          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="codeForm.enabled">
              <span>Enable this feature code</span>
            </label>
          </div>

          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="codeForm.playConfirmation">
              <span>Play confirmation tone after execution</span>
            </label>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveCode" :disabled="!canSave">
            {{ editingCode ? 'Save Changes' : 'Create Feature Code' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Plus as PlusIcon, PlusCircle as PlusCircleIcon, Upload as UploadIcon,
  Settings as SettingsIcon, Phone as PhoneIcon, Voicemail as VoicemailIcon,
  Users as UsersIcon, Code as CodeIcon, Search as SearchIcon,
  Edit as EditIcon, Trash2 as TrashIcon, Play as PlayIcon, X as XIcon,
  Activity as ActivityIcon, Globe as GlobeIcon, Volume2 as SoundIcon,
  FileCode as ScriptIcon, ArrowRightLeft as TransferIcon, Server as InternalIcon,
  Zap as ZapIcon, HelpCircle as HelpCircleIcon
} from 'lucide-vue-next'

const router = useRouter()

const activeCategory = ref('all')
const searchQuery = ref('')
const statusFilter = ref('')
const showAddModal = ref(false)
const showImportModal = ref(false)
const editingCode = ref(null)
const showSpeedDialHelp = ref(false)

const speedDialSettings = ref({
  prefix: '*0',
  startDigit: 1,
  endDigit: 9,
  enableUserDefined: true
})

const codeForm = ref({
  code: '',
  name: '',
  description: '',
  category: 'call-handling',
  type: 'Internal',
  target: '',
  enabled: true,
  playConfirmation: true
})

const featureCodes = ref([
  // System Codes
  { id: 1, code: '*97', name: 'Check Voicemail', description: 'Access voicemail for your extension', type: 'Internal', target: 'app::voicemail', category: 'voicemail', status: 'active', system: true, usageCount: 245 },
  { id: 2, code: '*98', name: 'Check Any Voicemail', description: 'Access voicemail with extension prompt', type: 'Internal', target: 'app::voicemail_any', category: 'voicemail', status: 'active', system: true, usageCount: 45 },
  
  // Call Handling
  { id: 3, code: '*72', name: 'Call Forward Enable', description: 'Forward calls to another number', type: 'Internal', target: 'app::cf_set', category: 'call-handling', status: 'active', system: true, usageCount: 128 },
  { id: 4, code: '*73', name: 'Call Forward Cancel', description: 'Cancel call forwarding', type: 'Internal', target: 'app::cf_cancel', category: 'call-handling', status: 'active', system: true, usageCount: 89 },
  { id: 5, code: '*78', name: 'Do Not Disturb On', description: 'Enable DND - calls go to voicemail', type: 'Internal', target: 'app::dnd_on', category: 'call-handling', status: 'active', system: true, usageCount: 156 },
  { id: 6, code: '*79', name: 'Do Not Disturb Off', description: 'Disable DND', type: 'Internal', target: 'app::dnd_off', category: 'call-handling', status: 'active', system: true, usageCount: 134 },
  { id: 7, code: '*70', name: 'Park Call', description: 'Park current call in lot', type: 'Internal', target: 'app::park', category: 'call-handling', status: 'active', system: true, usageCount: 67 },
  { id: 8, code: '*71', name: 'Pickup Parked Call', description: 'Retrieve parked call', type: 'Internal', target: 'app::pickup', category: 'call-handling', status: 'active', system: true, usageCount: 65 },
  
  // Queue/Agent
  { id: 9, code: '*50', name: 'Agent Login', description: 'Log into call queues', type: 'Internal', target: 'app::agent_login', category: 'queue', status: 'active', system: true, usageCount: 89 },
  { id: 10, code: '*51', name: 'Agent Logout', description: 'Log out from call queues', type: 'Internal', target: 'app::agent_logout', category: 'queue', status: 'active', system: true, usageCount: 87 },
  { id: 11, code: '*52', name: 'Agent Pause', description: 'Pause to take a break', type: 'Internal', target: 'app::agent_pause', category: 'queue', status: 'active', system: true, usageCount: 156 },
  
  // Custom
  { id: 12, code: '*77', name: 'Mark Room Clean', description: 'Hospitality: mark room as cleaned', type: 'CURL', target: 'https://api.hotel-pms.com/rooms/clean?room=${caller_id}', category: 'custom', status: 'active', system: false, usageCount: 234 },
  { id: 13, code: '*99', name: 'Sound Test', description: 'Play test audio', type: 'Sound', target: 'local_stream://default', category: 'system', status: 'active', system: false, usageCount: 12 },
  { id: 14, code: '*88', name: 'Intercom', description: 'Intercom to extension', type: 'Internal', target: 'app::intercom', category: 'call-handling', status: 'disabled', system: true, usageCount: 0 },
])

const filteredCodes = computed(() => {
  return featureCodes.value.filter(code => {
    const matchesSearch = !searchQuery.value || 
      code.code.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      code.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesCategory = activeCategory.value === 'all' || code.category === activeCategory.value
    const matchesStatus = !statusFilter.value || code.status === statusFilter.value
    
    return matchesSearch && matchesCategory && matchesStatus
  })
})

const canSave = computed(() => codeForm.value.code && codeForm.value.name && codeForm.value.target)

const getTypeIcon = (type) => {
  const icons = {
    'Internal': InternalIcon,
    'Transfer': TransferIcon,
    'CURL': GlobeIcon,
    'Sound': SoundIcon,
    'Lua': ScriptIcon,
    'API': InternalIcon
  }
  return icons[type] || CodeIcon
}

const toggleStatus = (code) => {
  code.status = code.status === 'active' ? 'disabled' : 'active'
}

const editCode = (code) => {
  editingCode.value = code
  codeForm.value = { ...code, enabled: code.status === 'active' }
  showAddModal.value = true
}

const deleteCode = (code) => {
  if (confirm(`Delete feature code ${code.code}?`)) {
    featureCodes.value = featureCodes.value.filter(c => c.id !== code.id)
  }
}

const testCode = (code) => {
  alert(`Testing ${code.code}: ${code.name}\n\nThis would dial ${code.code} from your phone.`)
}

const saveCode = () => {
  if (editingCode.value) {
    Object.assign(editingCode.value, codeForm.value, { status: codeForm.value.enabled ? 'active' : 'disabled' })
  } else {
    featureCodes.value.push({
      id: Date.now(),
      ...codeForm.value,
      status: codeForm.value.enabled ? 'active' : 'disabled',
      system: false,
      usageCount: 0
    })
  }
  closeModal()
}

const closeModal = () => {
  showAddModal.value = false
  editingCode.value = null
  codeForm.value = { code: '', name: '', description: '', category: 'call-handling', type: 'Internal', target: '', enabled: true, playConfirmation: true }
}
</script>

<style scoped>
.feature-codes-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

/* Category Tabs */
.category-tabs { display: flex; gap: 4px; margin-bottom: 16px; border-bottom: 1px solid var(--border-color); padding-bottom: 0; }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: none; border-bottom: 2px solid transparent; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); margin-bottom: -1px; }
.tab:hover { color: var(--text-primary); }
.tab.active { color: var(--primary-color); border-bottom-color: var(--primary-color); }
.tab-icon { width: 14px; height: 14px; }
.tab-count { background: var(--bg-app); padding: 2px 8px; border-radius: 10px; font-size: 11px; }

/* Filter Bar */
.filter-bar { display: flex; gap: 12px; margin-bottom: 20px; }
.search-box { position: relative; flex: 1; max-width: 300px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 10px 12px 10px 38px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.filter-select { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }

/* Codes Grid */
.codes-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }

.code-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: all 0.2s;
}
.code-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.code-card.disabled { opacity: 0.6; }

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
}

.code-badge {
  font-family: monospace;
  font-size: 20px;
  font-weight: 700;
  color: var(--primary-color);
  background: white;
  padding: 6px 14px;
  border-radius: 8px;
  border: 2px solid var(--primary-color);
}

.code-body { padding: 16px; }
.code-name { margin: 0 0 4px; font-size: 15px; font-weight: 600; }
.code-description { margin: 0 0 12px; font-size: 12px; color: var(--text-muted); line-height: 1.4; }

.code-type { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.type-icon { width: 14px; height: 14px; }
.type-label { font-size: 11px; font-weight: 600; padding: 3px 10px; border-radius: 4px; text-transform: uppercase; }
.type-label.internal { background: #dcfce7; color: #16a34a; }
.type-label.transfer { background: #dbeafe; color: #2563eb; }
.type-label.curl { background: #fef3c7; color: #b45309; }
.type-label.sound { background: #fce7f3; color: #be185d; }
.type-label.lua { background: #f3e8ff; color: #7c3aed; }
.type-label.api { background: #e0e7ff; color: #4338ca; }

.code-target { margin-top: 10px; padding: 8px; background: var(--bg-app); border-radius: 6px; }
.target-label { font-size: 10px; font-weight: 600; color: var(--text-muted); display: block; margin-bottom: 4px; }
.target-value { font-size: 11px; font-family: monospace; word-break: break-all; }

.code-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-app);
}

.usage-count { display: flex; align-items: center; gap: 4px; font-size: 11px; color: var(--text-muted); }
.usage-icon { width: 12px; height: 12px; }

.code-actions { display: flex; gap: 4px; }
.action-btn { width: 28px; height: 28px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }

.code-card.add-new {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  border: 2px dashed var(--border-color);
  cursor: pointer;
  color: var(--text-muted);
  gap: 12px;
}
.code-card.add-new:hover { border-color: var(--primary-color); color: var(--primary-color); }
.add-icon { width: 40px; height: 40px; }

/* Toggle */
.toggle-sm { position: relative; display: inline-block; width: 36px; height: 20px; }
.toggle-sm input { opacity: 0; width: 0; height: 0; }
.toggle-slider-sm { position: absolute; cursor: pointer; inset: 0; background: #e2e8f0; border-radius: 20px; transition: 0.3s; }
.toggle-slider-sm:before { content: ''; position: absolute; width: 16px; height: 16px; left: 2px; bottom: 2px; background: white; border-radius: 50%; transition: 0.3s; }
.toggle-sm input:checked + .toggle-slider-sm { background: #22c55e; }
.toggle-sm input:checked + .toggle-slider-sm:before { transform: translateX(16px); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 520px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-row { display: flex; gap: 12px; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.mono { font-family: monospace; font-size: 16px; font-weight: 600; }
.form-divider { height: 1px; background: var(--border-color); margin: 8px 0 16px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { width: 16px; height: 16px; }

.icon-sm { width: 14px; height: 14px; }

/* Responsive */
@media (max-width: 768px) {
  .category-tabs { overflow-x: auto; }
  .codes-grid { grid-template-columns: 1fr; }
}

/* Speed Dial Settings */
.speed-dial-settings { margin-bottom: 24px; }
.settings-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.settings-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; background: linear-gradient(135deg, #fef3c7, #fef9c3); border-bottom: 1px solid #fde68a; }
.settings-title { display: flex; align-items: center; gap: 12px; }
.settings-icon { width: 24px; height: 24px; color: #b45309; }
.settings-title h4 { margin: 0; font-size: 14px; font-weight: 600; color: #92400e; }
.settings-title p { margin: 2px 0 0; font-size: 12px; color: #a16207; }
.settings-body { padding: 20px; display: flex; justify-content: space-between; align-items: flex-start; gap: 24px; flex-wrap: wrap; }
.settings-form { display: flex; gap: 24px; flex-wrap: wrap; align-items: flex-end; }
.form-group.compact { margin-bottom: 0; }
.form-group.compact label { margin-bottom: 4px; }
.prefix-input-group { display: flex; align-items: center; gap: 12px; }
.prefix-input-group .help-text { font-size: 12px; color: var(--text-muted); }
.range-inputs { display: flex; align-items: center; gap: 8px; }
.range-inputs span { color: var(--text-muted); font-size: 12px; }
.checkbox-label { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; text-transform: none; font-weight: 500; color: var(--text-main); }
.checkbox-label input { width: 16px; height: 16px; }
.settings-preview { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text-muted); }
.preview-label { font-weight: 600; }
.preview-code { background: #f1f5f9; padding: 4px 8px; border-radius: 4px; font-family: monospace; font-weight: 600; color: #b45309; }
.preview-count { font-size: 11px; }
.help-panel { padding: 16px 20px; background: #fffbeb; border-top: 1px solid #fde68a; }
.help-panel p { margin: 0 0 8px; font-size: 13px; }
.help-panel ul { margin: 0; padding-left: 20px; font-size: 12px; color: var(--text-main); }
.help-panel li { margin-bottom: 4px; }
.help-panel code { background: white; padding: 2px 6px; border-radius: 4px; font-size: 11px; }
.help-panel a { color: var(--primary-color); font-weight: 500; }
.btn-link { background: none; border: none; padding: 4px; cursor: pointer; color: #b45309; border-radius: 4px; }
.btn-link:hover { background: rgba(180, 83, 9, 0.1); }
</style>
