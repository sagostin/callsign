<template>
  <div class="templates-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Device Templates</h2>
        <p class="text-muted text-sm">Manage master provisioning templates organized by manufacturer and model.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showVarsModal = true">
          <CodeIcon class="btn-icon" /> Variables
        </button>
        <button class="btn-primary" @click="showCreateModal = true">
          <PlusIcon class="btn-icon" /> New Template
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ manufacturers.length }}</div>
        <div class="stat-label">Manufacturers</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ allTemplates.length }}</div>
        <div class="stat-label">Templates</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ totalModels }}</div>
        <div class="stat-label">Device Models</div>
      </div>
    </div>

    <!-- Manufacturer Tabs -->
    <div class="manufacturer-tabs">
      <button 
        class="mfg-tab" 
        :class="{ active: selectedManufacturer === 'all' }" 
        @click="selectedManufacturer = 'all'"
      >
        All Templates
      </button>
      <button 
        v-for="mfg in manufacturers" 
        :key="mfg.id"
        class="mfg-tab"
        :class="{ active: selectedManufacturer === mfg.id }"
        @click="selectedManufacturer = mfg.id"
      >
        <img v-if="mfg.logo" :src="mfg.logo" :alt="mfg.name" class="mfg-logo">
        <span>{{ mfg.name }}</span>
        <span class="mfg-count">{{ getTemplateCount(mfg.id) }}</span>
      </button>
    </div>

    <!-- Templates Content -->
    <div class="templates-content">
      <!-- Manufacturer Groups (when All selected) -->
      <div v-if="selectedManufacturer === 'all'" class="manufacturer-groups">
        <div v-for="mfg in manufacturersWithTemplates" :key="mfg.id" class="manufacturer-group">
          <div class="group-header" @click="toggleManufacturer(mfg.id)">
            <div class="group-info">
              <component :is="getManufacturerIcon(mfg.id)" class="mfg-icon" />
              <h3>{{ mfg.name }}</h3>
              <span class="template-count">{{ getTemplateCount(mfg.id) }} templates</span>
            </div>
            <ChevronDownIcon class="expand-icon" :class="{ expanded: expandedMfgs.includes(mfg.id) }" />
          </div>
          
          <transition name="slide">
            <div v-if="expandedMfgs.includes(mfg.id)" class="group-content">
              <div class="model-grid">
                <div v-for="template in getTemplatesForMfg(mfg.id)" :key="template.id" class="template-card">
                  <div class="card-header">
                    <div class="model-info">
                      <h4>{{ template.model }}</h4>
                      <span class="template-name">{{ template.name }}</span>
                    </div>
                    <span class="version-badge">{{ template.version }}</span>
                  </div>
                  <div class="card-meta">
                    <span><TenantsIcon /> {{ template.tenants }} tenants</span>
                    <span><CalendarIcon /> {{ template.updated }}</span>
                  </div>
                  <div class="card-actions">
                    <button class="btn-sm" @click="editTemplate(template)"><EditIcon /> Edit</button>
                    <button class="btn-sm" @click="duplicateTemplate(template)"><CopyIcon /> Clone</button>
                    <button class="btn-sm danger" @click="deleteTemplate(template)"><TrashIcon /></button>
                  </div>
                </div>
              </div>
            </div>
          </transition>
        </div>
      </div>

      <!-- Single Manufacturer View -->
      <div v-else class="single-manufacturer">
        <div class="search-bar">
          <SearchIcon class="search-icon" />
          <input v-model="searchQuery" placeholder="Search models..." class="search-input">
        </div>
        
        <div class="model-grid large">
          <div v-for="template in filteredTemplates" :key="template.id" class="template-card">
            <div class="card-header">
              <div class="model-info">
                <h4>{{ template.model }}</h4>
                <span class="template-name">{{ template.name }}</span>
              </div>
              <span class="version-badge">{{ template.version }}</span>
            </div>
            <div class="card-meta">
              <span><TenantsIcon /> {{ template.tenants }} tenants</span>
              <span><CalendarIcon /> {{ template.updated }}</span>
            </div>
            <div class="card-actions">
              <button class="btn-sm" @click="editTemplate(template)"><EditIcon /> Edit</button>
              <button class="btn-sm" @click="duplicateTemplate(template)"><CopyIcon /> Clone</button>
              <button class="btn-sm danger" @click="deleteTemplate(template)"><TrashIcon /></button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>{{ editingTemplate ? 'Edit Template' : 'New Device Template' }}</h3>
          <button class="close-btn" @click="showCreateModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="editor-layout">
            <div class="editor-sidebar">
              <div class="form-group">
                <label>Manufacturer</label>
                <select v-model="form.manufacturer" class="input-field">
                  <option v-for="mfg in manufacturers" :key="mfg.id" :value="mfg.id">{{ mfg.name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label>Model</label>
                <input v-model="form.model" class="input-field" placeholder="T54W, T57W">
              </div>
              <div class="form-group">
                <label>Template Name</label>
                <input v-model="form.name" class="input-field" placeholder="Yealink T5 Series Default">
              </div>
              <div class="form-group">
                <label>Version</label>
                <input v-model="form.version" class="input-field code" placeholder="v1.0">
              </div>
              
              <div class="divider"></div>
              
              <div class="variables-hint">
                <h5>Variables</h5>
                <div class="var-list">
                  <code>{$mac_address}</code>
                  <code>{$display_name_1}</code>
                  <code>{$user_id_1}</code>
                  <code>{$password_1}</code>
                </div>
              </div>
            </div>
            <div class="editor-main">
              <div class="code-header">
                <span>Configuration Template</span>
                <select v-model="form.format" class="format-select">
                  <option value="cfg">CFG</option>
                  <option value="xml">XML</option>
                  <option value="ini">INI</option>
                </select>
              </div>
              <textarea v-model="form.content" class="code-editor" spellcheck="false" placeholder="# Enter configuration..."></textarea>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showCreateModal = false">Cancel</button>
          <button class="btn-primary" @click="saveTemplate" :disabled="saving || !form.name">
            {{ saving ? 'Saving...' : (editingTemplate ? 'Update' : 'Create') }} Template
          </button>
        </div>
      </div>
    </div>

    <!-- Variables Modal -->
    <div v-if="showVarsModal" class="modal-overlay" @click.self="showVarsModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Provisioning Variables</h3>
          <button class="close-btn" @click="showVarsModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="vars-grid">
            <div v-for="cat in variableCategories" :key="cat.name" class="var-category">
              <h4>{{ cat.name }}</h4>
              <div class="var-items">
                <div v-for="v in cat.vars" :key="v.name" class="var-item">
                  <code>{{ v.name }}</code>
                  <span>{{ v.desc }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { 
  Plus as PlusIcon, Code as CodeIcon, Search as SearchIcon,
  Edit as EditIcon, Copy as CopyIcon, Trash2 as TrashIcon,
  ChevronDown as ChevronDownIcon, Users as TenantsIcon, Calendar as CalendarIcon,
  Monitor, Smartphone, Server as ServerIcon
} from 'lucide-vue-next'
import { systemAPI } from '@/services/api'

const toast = inject('toast')
const loading = ref(true)
const saving = ref(false)

const selectedManufacturer = ref('all')
const expandedMfgs = ref([])
const searchQuery = ref('')
const showCreateModal = ref(false)
const showVarsModal = ref(false)
const editingTemplate = ref(null)

const form = ref({
  manufacturer: 'yealink',
  model: '',
  name: '',
  version: 'v1.0',
  format: 'cfg',
  content: ''
})

const manufacturers = ref([])

// Default manufacturers if API returns empty
const defaultManufacturers = [
  { id: 'yealink', code: 'yealink', name: 'Yealink', logo_url: null },
  { id: 'poly', code: 'poly', name: 'Poly', logo_url: null },
  { id: 'cisco', code: 'cisco', name: 'Cisco', logo_url: null },
  { id: 'grandstream', code: 'grandstream', name: 'Grandstream', logo_url: null },
  { id: 'fanvil', code: 'fanvil', name: 'Fanvil', logo_url: null },
  { id: 'generic', code: 'generic', name: 'Generic SIP', logo_url: null },
]

const allTemplates = ref([])

// Load templates from API
const loadTemplates = async () => {
  loading.value = true
  try {
    // Load manufacturers first
    try {
      const mfgRes = await systemAPI.listDeviceManufacturers()
      const mfgList = mfgRes.data?.data || mfgRes.data || []
      if (mfgList.length > 0) {
        manufacturers.value = mfgList.map(m => ({
          id: m.code,
          name: m.name,
          logo: m.logo_url
        }))
      } else {
        manufacturers.value = defaultManufacturers
      }
    } catch (e) {
      console.log('Using default manufacturers')
      manufacturers.value = defaultManufacturers
    }

    const response = await systemAPI.listDeviceTemplates()
    const templates = response.data?.data || response.data || []
    allTemplates.value = templates.map(t => ({
      id: t.id,
      uuid: t.uuid,
      manufacturer: t.manufacturer?.toLowerCase() || 'generic',
      model: t.model || t.device_model || 'Unknown',
      name: t.name,
      version: t.version || 'v1.0',
      tenants: t.tenant_count || 0,
      updated: t.updated_at?.split('T')[0] || '',
      format: t.file_type || 'cfg',
      content: t.content || ''
    }))
    // Auto-expand first manufacturer with templates
    if (allTemplates.value.length > 0 && expandedMfgs.value.length === 0) {
      expandedMfgs.value = [allTemplates.value[0].manufacturer]
    }
  } catch (e) {
    console.error('Failed to load templates:', e)
    toast?.error('Failed to load templates')
  } finally {
    loading.value = false
  }
}

onMounted(loadTemplates)

const variableCategories = ref([
  { name: 'Device', vars: [
    { name: '{$mac_address}', desc: 'Device MAC address' },
    { name: '{$device_name}', desc: 'Friendly device name' },
  ]},
  { name: 'Line 1', vars: [
    { name: '{$display_name_1}', desc: 'Line 1 display name' },
    { name: '{$user_id_1}', desc: 'Line 1 user ID' },
    { name: '{$auth_id_1}', desc: 'Line 1 auth ID' },
    { name: '{$password_1}', desc: 'Line 1 password' },
    { name: '{$realm_1}', desc: 'Line 1 SIP realm' },
  ]},
  { name: 'Server', vars: [
    { name: '{$sip_server}', desc: 'Primary SIP server' },
    { name: '{$outbound_proxy}', desc: 'Outbound proxy' },
    { name: '{$ntp_server}', desc: 'NTP server address' },
  ]},
])

const manufacturersWithTemplates = computed(() => {
  return manufacturers.value.filter(m => allTemplates.value.some(t => t.manufacturer === m.id))
})

const totalModels = computed(() => {
  return [...new Set(allTemplates.value.map(t => t.model))].length
})

const filteredTemplates = computed(() => {
  let templates = allTemplates.value
  if (selectedManufacturer.value !== 'all') {
    templates = templates.filter(t => t.manufacturer === selectedManufacturer.value)
  }
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    templates = templates.filter(t => t.model.toLowerCase().includes(q) || t.name.toLowerCase().includes(q))
  }
  return templates
})

const getTemplateCount = (mfgId) => allTemplates.value.filter(t => t.manufacturer === mfgId).length
const getTemplatesForMfg = (mfgId) => allTemplates.value.filter(t => t.manufacturer === mfgId)

const toggleManufacturer = (mfgId) => {
  const idx = expandedMfgs.value.indexOf(mfgId)
  if (idx === -1) expandedMfgs.value.push(mfgId)
  else expandedMfgs.value.splice(idx, 1)
}

const getManufacturerIcon = (mfgId) => {
  if (mfgId === 'generic') return ServerIcon
  return Monitor
}

const editTemplate = (template) => {
  editingTemplate.value = template
  form.value = { ...template }
  showCreateModal.value = true
}

const duplicateTemplate = async (template) => {
  try {
    const payload = {
      vendor: template.manufacturer,
      model: template.model,
      name: template.name + ' (Copy)',
      file_type: template.format,
      content: template.content
    }
    await systemAPI.createDeviceTemplate(payload)
    toast?.success('Template duplicated')
    await loadTemplates()
  } catch (e) {
    toast?.error('Failed to duplicate template')
  }
}

const saveTemplate = async () => {
  saving.value = true
  try {
    const payload = {
      vendor: form.value.manufacturer,
      model: form.value.model,
      name: form.value.name,
      file_type: form.value.format,
      content: form.value.content
    }
    
    if (editingTemplate.value?.id) {
      await systemAPI.updateDeviceTemplate(editingTemplate.value.id, payload)
      toast?.success('Template updated')
    } else {
      await systemAPI.createDeviceTemplate(payload)
      toast?.success('Template created')
    }
    showCreateModal.value = false
    editingTemplate.value = null
    form.value = { manufacturer: 'yealink', model: '', name: '', version: 'v1.0', format: 'cfg', content: '' }
    await loadTemplates()
  } catch (e) {
    toast?.error('Failed to save template', e.message)
  } finally {
    saving.value = false
  }
}

const deleteTemplate = async (template) => {
  if (!confirm(`Delete template "${template.name}"?`)) return
  try {
    await systemAPI.deleteDeviceTemplate(template.id)
    toast?.success('Template deleted')
    await loadTemplates()
  } catch (e) {
    toast?.error('Failed to delete template', e.message)
  }
}
</script>

<style scoped>
.templates-page { padding: 0; }
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

/* Manufacturer Tabs */
.manufacturer-tabs { display: flex; gap: 4px; margin-bottom: 20px; overflow-x: auto; padding-bottom: 4px; }
.mfg-tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: white; border: 1px solid var(--border-color); border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; white-space: nowrap; transition: all 0.2s; }
.mfg-tab:hover { border-color: var(--primary-color); }
.mfg-tab.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.mfg-logo { width: 16px; height: 16px; object-fit: contain; }
.mfg-count { font-size: 10px; background: rgba(0,0,0,0.1); padding: 2px 6px; border-radius: 10px; }
.mfg-tab.active .mfg-count { background: rgba(255,255,255,0.2); }

/* Templates Content */
.templates-content { background: white; border: 1px solid var(--border-color); border-radius: 8px; }

/* Manufacturer Groups */
.manufacturer-group { border-bottom: 1px solid var(--border-color); }
.manufacturer-group:last-child { border-bottom: none; }
.group-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; cursor: pointer; transition: background 0.2s; }
.group-header:hover { background: #f8fafc; }
.group-info { display: flex; align-items: center; gap: 12px; }
.mfg-icon { width: 24px; height: 24px; color: var(--text-muted); }
.group-info h3 { margin: 0; font-size: 15px; }
.template-count { font-size: 12px; color: var(--text-muted); background: #f1f5f9; padding: 2px 8px; border-radius: 4px; }
.expand-icon { width: 20px; height: 20px; color: var(--text-muted); transition: transform 0.2s; }
.expand-icon.expanded { transform: rotate(180deg); }
.group-content { padding: 0 20px 20px; }

/* Model Grid */
.model-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }
.model-grid.large { grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); }

.template-card { background: #f8fafc; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.template-card:hover { border-color: var(--primary-color); }
.card-header { display: flex; justify-content: space-between; align-items: flex-start; padding: 14px; background: white; border-bottom: 1px solid var(--border-color); }
.model-info h4 { margin: 0 0 4px; font-size: 14px; }
.template-name { font-size: 11px; color: var(--text-muted); }
.version-badge { font-size: 10px; font-family: monospace; background: #dbeafe; color: #2563eb; padding: 2px 8px; border-radius: 4px; }
.card-meta { display: flex; gap: 16px; padding: 10px 14px; font-size: 11px; color: var(--text-muted); }
.card-meta span { display: flex; align-items: center; gap: 4px; }
.card-meta svg { width: 12px; height: 12px; }
.card-actions { display: flex; gap: 8px; padding: 10px 14px; background: white; border-top: 1px solid var(--border-color); }
.btn-sm { display: flex; align-items: center; gap: 4px; padding: 6px 10px; background: white; border: 1px solid var(--border-color); border-radius: 4px; font-size: 11px; cursor: pointer; }
.btn-sm:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-sm.danger:hover { border-color: #ef4444; color: #ef4444; }
.btn-sm svg { width: 12px; height: 12px; }

/* Single Manufacturer View */
.single-manufacturer { padding: 20px; }
.search-bar { display: flex; align-items: center; gap: 8px; background: #f8fafc; padding: 10px 14px; border-radius: 8px; margin-bottom: 20px; }
.search-icon { width: 16px; height: 16px; color: var(--text-muted); }
.search-input { border: none; background: none; flex: 1; font-size: 13px; outline: none; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 500px; max-height: 85vh; overflow: hidden; display: flex; flex-direction: column; }
.modal-card.large { max-width: 900px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Editor Layout */
.editor-layout { display: grid; grid-template-columns: 240px 1fr; gap: 20px; min-height: 400px; }
.editor-sidebar { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; }
.input-field.code { font-family: monospace; }
.divider { border-top: 1px solid var(--border-color); margin: 8px 0; }
.variables-hint h5 { font-size: 11px; text-transform: uppercase; color: var(--text-muted); margin: 0 0 8px; }
.var-list { display: flex; flex-direction: column; gap: 4px; }
.var-list code { font-size: 11px; background: #f1f5f9; padding: 4px 8px; border-radius: 4px; }
.editor-main { display: flex; flex-direction: column; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.code-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #1e293b; color: white; font-size: 12px; }
.format-select { background: #334155; color: white; border: none; border-radius: 4px; padding: 4px 8px; font-size: 11px; }
.code-editor { flex: 1; padding: 12px; font-family: monospace; font-size: 12px; border: none; resize: none; background: #1e293b; color: #e2e8f0; min-height: 300px; }
.code-editor::placeholder { color: #64748b; }

/* Variables Modal */
.vars-grid { display: grid; gap: 20px; }
.var-category h4 { font-size: 12px; text-transform: uppercase; color: var(--text-muted); margin: 0 0 12px; padding-bottom: 8px; border-bottom: 1px solid var(--border-color); }
.var-items { display: flex; flex-direction: column; gap: 8px; }
.var-item { display: flex; align-items: center; gap: 12px; }
.var-item code { font-size: 11px; background: #f1f5f9; padding: 4px 8px; border-radius: 4px; min-width: 140px; }
.var-item span { font-size: 12px; color: var(--text-muted); }

/* Transitions */
.slide-enter-active, .slide-leave-active { transition: all 0.3s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; max-height: 0; padding-top: 0; padding-bottom: 0; }
</style>
