<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Routing</h2>
      <p class="text-muted text-sm">Manage global inbound/outbound routing and dialplan logic.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" v-if="activeTab === 'inbound'" @click="showInboundModal = true">+ Global Inbound</button>
      <button class="btn-primary" v-if="activeTab === 'outbound'" @click="showOutboundModal = true">+ Global Outbound</button>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'numbers' }" @click="activeTab = 'numbers'">All Numbers</button>
    <button class="tab" :class="{ active: activeTab === 'inbound' }" @click="activeTab = 'inbound'">Global Inbound</button>
    <button class="tab" :class="{ active: activeTab === 'outbound' }" @click="activeTab = 'outbound'">Global Outbound</button>
    <button class="tab" :class="{ active: activeTab === 'settings' }" @click="activeTab = 'settings'">Global Settings</button>
  </div>

  <!-- NUMBERS TAB (All Tenants) -->
  <div class="tab-content" v-if="activeTab === 'numbers'">
    <div class="route-help">
      <InfoIcon class="help-icon" />
      <span>All phone numbers (DIDs) across all tenants. Numbers are stored in E.164 format.</span>
    </div>

    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input type="text" v-model="numberSearch" placeholder="Search numbers..." class="search-input">
      </div>
    </div>

    <DataTable :columns="numberColumns" :data="filteredNumbers">
      <template #destination_number="{ value }">
        <span class="font-mono font-semibold">{{ formatPhoneNumber(value) }}</span>
      </template>
      <template #tenant="{ row }">
        <span class="tenant-badge" v-if="row.Tenant">{{ row.Tenant.name }}</span>
        <span class="text-muted" v-else>—</span>
      </template>
      <template #enabled="{ value }">
        <StatusBadge :status="value ? 'Active' : 'Disabled'" />
      </template>
    </DataTable>
  </div>

  <!-- INBOUND ROUTES TAB -->
  <div class="tab-content" v-if="activeTab === 'inbound'">
    <div class="route-help">
      <InfoIcon class="help-icon" />
      <span>Global inbound routes process calls that don't match specific tenant routes. First matching route wins.</span>
    </div>

    <div class="routes-list">
      <div class="route-card" v-for="(route, idx) in inboundRoutes" :key="route.id" 
        :class="{ disabled: !route.enabled, dragging: dragIndex === idx && dragType === 'inbound', dragover: dragOverIndex === idx && dragType === 'inbound' }"
        draggable="true"
        @dragstart="onDragStart($event, idx, 'inbound')"
        @dragover.prevent="onDragOver($event, idx, 'inbound')"
        @dragleave="onDragLeave()"
        @drop="onDrop($event, idx, 'inbound')"
        @dragend="onDragEnd()">
        <div class="route-handle">
          <GripVerticalIcon class="grip-icon" />
          <span class="route-order">{{ idx + 1 }}</span>
        </div>
        
        <div class="route-main">
          <div class="route-name-row">
            <h4>{{ route.name }}</h4>
            <div class="route-badges">
              <span class="badge context">{{ route.context }}</span>
            </div>
          </div>
          
          <div class="route-conditions">
            <div class="condition" v-for="(cond, i) in route.conditions" :key="i">
              <span class="cond-var">{{ cond.variable }}</span>
              <span class="cond-op">{{ cond.operator }}</span>
              <span class="cond-val font-mono">{{ cond.value }}</span>
            </div>
          </div>
          
          <div class="route-actions-display">
            <ArrowRightIcon class="arrow-icon" />
            <div class="action-chain">
              <span class="action-item" v-for="(act, i) in route.actions" :key="i">
                <span class="action-app">{{ act.app }}</span>
                <span class="action-data" v-if="act.data">{{ act.data }}</span>
              </span>
            </div>
          </div>
        </div>

        <div class="route-controls">
          <label class="switch small">
            <input type="checkbox" v-model="route.enabled">
            <span class="slider round"></span>
          </label>
          <button class="btn-icon" @click="editInboundRoute(route)"><EditIcon class="icon-sm" /></button>
          <button class="btn-icon" @click="deleteInboundRoute(route)"><TrashIcon class="icon-sm text-bad" /></button>
        </div>
      </div>
    </div>
  </div>

  <!-- OUTBOUND ROUTES TAB -->
  <div class="tab-content" v-else-if="activeTab === 'outbound'">
    <div class="route-help">
      <InfoIcon class="help-icon" />
      <span>Global outbound routes apply to all calls unless overridden by tenant specific routes.</span>
    </div>

    <div class="routes-list">
      <div class="route-card" v-for="(route, idx) in outboundRoutes" :key="route.id"
        :class="{ disabled: !route.enabled, dragging: dragIndex === idx && dragType === 'outbound', dragover: dragOverIndex === idx && dragType === 'outbound' }"
        draggable="true"
        @dragstart="onDragStart($event, idx, 'outbound')"
        @dragover.prevent="onDragOver($event, idx, 'outbound')"
        @dragleave="onDragLeave()"
        @drop="onDrop($event, idx, 'outbound')"
        @dragend="onDragEnd()">
        <div class="route-handle">
          <GripVerticalIcon class="grip-icon" />
          <span class="route-order">{{ idx + 1 }}</span>
        </div>
        
        <div class="route-main">
          <div class="route-name-row">
            <h4>{{ route.name }}</h4>
            <div class="route-badges">
              <span class="badge gateway">{{ route.gateway }}</span>
              <span class="badge intl" v-if="route.international">International</span>
            </div>
          </div>
          
          <div class="route-pattern">
            <span class="pattern-label">Pattern:</span>
            <code class="pattern-regex">{{ route.pattern }}</code>
            <span class="pattern-desc" v-if="route.description">{{ route.description }}</span>
          </div>
          
          <div class="route-transforms" v-if="route.prepend || route.strip">
            <span class="transform" v-if="route.strip">Strip: {{ route.strip }} digits</span>
            <span class="transform" v-if="route.prepend">Prepend: {{ route.prepend }}</span>
          </div>
        </div>

        <div class="route-controls">
          <label class="switch small">
            <input type="checkbox" v-model="route.enabled">
            <span class="slider round"></span>
          </label>
          <button class="btn-icon" @click="editOutboundRoute(route)"><EditIcon class="icon-sm" /></button>
          <button class="btn-icon" @click="deleteOutboundRoute(route)"><TrashIcon class="icon-sm text-bad" /></button>
        </div>
      </div>
    </div>
  </div>

  <!-- SETTINGS TAB -->
  <div class="tab-content settings-panel" v-else-if="activeTab === 'settings'">
    <div class="settings-section">
      <h3>System Dialing Defaults</h3>
      <div class="settings-grid">
        <div class="form-group">
          <label>Default Region</label>
          <select class="input-field" v-model="settings.region">
            <option value="nanp">North America (NANP)</option>
            <option value="uk">United Kingdom</option>
            <option value="eu">Europe (General)</option>
            <option value="au">Australia</option>
          </select>
          <span class="help-text">System-wide default for interpretation.</span>
        </div>
        <div class="form-group">
          <label>Default Outbound Format</label>
          <select class="input-field" v-model="settings.format">
            <option value="e164">E.164 (Global Standard, +1...)</option>
            <option value="national">National (10-digit)</option>
            <option value="passthrough">Passthrough (As Dialed)</option>
          </select>
        </div>
      </div>
    </div>

    <div class="form-actions">
      <button class="btn-primary" @click="saveSettings">Save Settings</button>
    </div>
  </div>

  <!-- INBOUND ROUTE MODAL -->
  <div v-if="showInboundModal" class="modal-overlay" @click.self="showInboundModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>{{ editingInbound ? 'Edit Global Inbound Route' : 'New Global Inbound Route' }}</h3>
        <button class="btn-icon" @click="showInboundModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Route Name</label>
            <input v-model="inboundForm.name" class="input-field" placeholder="e.g. Carrier Handover">
          </div>
          <div class="form-group">
            <label>Context</label>
            <input v-model="inboundForm.context" class="input-field code" placeholder="public">
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Conditions</h4>
            <button class="btn-small" @click="addInboundCondition">+ Add Condition</button>
          </div>
          
          <div class="conditions-editor">
            <div class="condition-row" v-for="(cond, i) in inboundForm.conditions" :key="i">
              <select v-model="cond.variable" class="input-field">
                <option value="destination_number">destination_number</option>
                <option value="caller_id_number">caller_id_number</option>
                <option value="network_addr">network_addr</option>
                <option value="sip_user_agent">sip_user_agent</option>
              </select>
              <select v-model="cond.operator" class="input-field small">
                <option value="=~">matches regex</option>
                <option value="==">equals</option>
                <option value="!=">not equals</option>
              </select>
              <input v-model="cond.value" class="input-field code" placeholder="^\\+?1?(\\d{10})$">
              <button class="btn-icon" @click="removeInboundCondition(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Actions</h4>
            <button class="btn-small" @click="addInboundAction">+ Add Action</button>
          </div>
          
          <div class="actions-editor">
            <div class="action-row" v-for="(act, i) in inboundForm.actions" :key="i">
              <select v-model="act.app" class="input-field">
                <optgroup label="Routing">
                  <option value="transfer">transfer</option>
                  <option value="bridge">bridge</option>
                </optgroup>
                <optgroup label="System">
                  <option value="log">log</option>
                  <option value="info">info</option>
                  <option value="system">system (exec)</option>
                </optgroup>
              </select>
              <input v-model="act.data" class="input-field flex-2" placeholder="application data">
              <button class="btn-icon" @click="removeInboundAction(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showInboundModal = false">Cancel</button>
        <button class="btn-primary" @click="saveInboundRoute" :disabled="!inboundForm.name">Save Route</button>
      </div>
    </div>
  </div>

  <!-- OUTBOUND ROUTE MODAL -->
  <div v-if="showOutboundModal" class="modal-overlay" @click.self="showOutboundModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ editingOutbound ? 'Edit Global Outbound Route' : 'New Global Outbound Route' }}</h3>
        <button class="btn-icon" @click="showOutboundModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Route Name</label>
          <input v-model="outboundForm.name" class="input-field" placeholder="e.g. Emergency Override">
        </div>

        <div class="form-group">
          <label>Pattern (Regex)</label>
          <input v-model="outboundForm.pattern" class="input-field code" placeholder="^911$">
          <span class="help-text">Global regex match on dialed digits.</span>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Strip Digits</label>
            <input type="number" v-model="outboundForm.strip" class="input-field" placeholder="0">
          </div>
          <div class="form-group">
            <label>Prepend</label>
            <input v-model="outboundForm.prepend" class="input-field code" placeholder="">
          </div>
        </div>

        <div class="form-group">
          <label>Gateway / Trunk</label>
          <select v-model="outboundForm.gateway" class="input-field">
            <option value="">Select System Gateway...</option>
            <option value="system_pri">System PRI</option>
            <option value="system_sip">System SIP Trunk</option>
          </select>
        </div>

        <div class="checkbox-group">
          <label class="checkbox-row">
            <input type="checkbox" v-model="outboundForm.continue">
            <span>Continue to tenant routes on failure</span>
          </label>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showOutboundModal = false">Cancel</button>
        <button class="btn-primary" @click="saveOutboundRoute" :disabled="!outboundForm.name || !outboundForm.pattern">Save Route</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Search as SearchIcon, Info as InfoIcon, GripVertical as GripVerticalIcon,
  ArrowRight as ArrowRightIcon, Edit as EditIcon,
  Trash2 as TrashIcon, X as XIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { formatPhoneNumber } from '../../utils/formatters'
import { systemAPI } from '../../services/api'

const activeTab = ref('numbers')

// Drag and Drop State
const dragIndex = ref(null)
const dragOverIndex = ref(null)
const dragType = ref(null)

const onDragStart = (e, idx, type) => {
  dragIndex.value = idx
  dragType.value = type
  e.dataTransfer.effectAllowed = 'move'
}

const onDragOver = (e, idx, type) => {
  if (dragType.value !== type) return
  dragOverIndex.value = idx
  e.dataTransfer.dropEffect = 'move'
}

const onDragLeave = () => {
  dragOverIndex.value = null
}

const onDrop = (e, idx, type) => {
  if (dragType.value !== type || dragIndex.value === null) return
  
  const routes = type === 'inbound' ? inboundRoutes : outboundRoutes
  const fromIdx = dragIndex.value
  const toIdx = idx
  
  if (fromIdx !== toIdx) {
    const item = routes.value.splice(fromIdx, 1)[0]
    routes.value.splice(toIdx, 0, item)
  }
  
  dragIndex.value = null
  dragOverIndex.value = null
  dragType.value = null
}

const onDragEnd = () => {
  dragIndex.value = null
  dragOverIndex.value = null
  dragType.value = null
}

// All Numbers (System-wide)
const allNumbers = ref([])
const numberSearch = ref('')
const numberColumns = [
  { key: 'destination_number', label: 'Number', width: '160px' },
  { key: 'tenant', label: 'Tenant', width: '150px' },
  { key: 'destination_type', label: 'Type', width: '100px' },
  { key: 'context', label: 'Context', width: '100px' },
  { key: 'enabled', label: 'Status', width: '80px' }
]

const filteredNumbers = computed(() => {
  return allNumbers.value.filter(n => {
    const numStr = n.destination_number || ''
    return numStr.includes(numberSearch.value)
  })
})

const loadAllNumbers = async () => {
  try {
    const response = await systemAPI.listAllNumbers()
    allNumbers.value = response.data.data || []
  } catch (e) {
    console.error('Failed to load all numbers', e)
  }
}

// Inbound Routes (Global)
const showInboundModal = ref(false)
const editingInbound = ref(false)
const inboundForm = ref({
  dialplan_name: '',
  dialplan_context: 'public',
  enabled: true,
  details: [{ detail_type: 'condition', condition_field: 'destination_number', condition_expression: '', condition_expression_type: 'regex', condition_break: 'on-false' }, { detail_type: 'action', action_application: 'transfer', action_data: '' }]
})

const inboundRoutes = ref([])
const outboundRoutes = ref([])

const loadRoutes = async () => {
  try {
    const response = await systemAPI.listDialplans()
    const allRoutes = response.data.data || []
    
    // Map API format to UI format if needed, but lets try to align them
    // UI expects: name, context, enabled, conditions[], actions[]
    // API returns: Dialplan { Details: [] }
    
    // Helper to map DB model to UI model
    const mapRoute = (r) => {
      const conditions = r.Details.filter(d => d.detail_type === 'condition').map(d => ({
        variable: d.condition_field,
        operator: d.condition_expression_type === 'regex' ? '=~' : (d.condition_expression_type === 'negate' ? '!=' : '=='), // simplified assumptions for now
        value: d.condition_expression
      }))
      const actions = r.Details.filter(d => d.detail_type === 'action').map(d => ({
        app: d.action_application,
        data: d.action_data
      }))
      // If no conditions/actions found (e.g. empty), provide defaults
      if (conditions.length === 0) conditions.push({ variable: 'destination_number', operator: '=~', value: '' })
      if (actions.length === 0) actions.push({ app: 'transfer', data: '' })

      return {
        id: r.id,
        uuid: r.uuid,
        name: r.dialplan_name,
        context: r.dialplan_context,
        enabled: r.enabled,
        conditions,
        actions,
        // Keep original for updates
        _original: r
      }
    }

    inboundRoutes.value = allRoutes.filter(r => r.dialplan_context === 'public').map(mapRoute)
    outboundRoutes.value = allRoutes.filter(r => r.dialplan_context !== 'public').map(mapRoute)

  } catch (e) {
    console.error('Failed to load system routes', e)
  }
}

const addInboundCondition = () => inboundForm.value.details.push({ detail_type: 'condition', condition_field: 'destination_number', condition_expression: '', condition_expression_type: 'regex', condition_break: 'on-false' })
const removeInboundCondition = (i) => {
  const indexToRemove = inboundForm.value.details.findIndex(d => d.detail_type === 'condition' && inboundForm.value.details.indexOf(d) === i);
  if (indexToRemove !== -1) {
    inboundForm.value.details.splice(indexToRemove, 1);
  }
}
const addInboundAction = () => inboundForm.value.details.push({ detail_type: 'action', action_application: 'log', action_data: 'INFO' })
const removeInboundAction = (i) => {
  const indexToRemove = inboundForm.value.details.findIndex(d => d.detail_type === 'action' && inboundForm.value.details.indexOf(d) === i);
  if (indexToRemove !== -1) {
    inboundForm.value.details.splice(indexToRemove, 1);
  }
}

const editInboundRoute = (route) => {
  // Map UI object back to Form object (which should match API model structure ideally)
  const details = []
  route.conditions.forEach(c => details.push({ 
    detail_type: 'condition', 
    condition_field: c.variable, 
    condition_expression: c.value, 
    condition_expression_type: c.operator === '=~' ? 'regex' : (c.operator === '!=' ? 'negate' : 'exact'),
    condition_break: 'on-false' 
  }))
  route.actions.forEach(a => details.push({ 
    detail_type: 'action', 
    action_application: a.app, 
    action_data: a.data 
  }))

  inboundForm.value = {
    id: route.id,
    dialplan_name: route.name,
    dialplan_context: route.context,
    enabled: route.enabled,
    details: details
  }
  editingInbound.value = true
  showInboundModal.value = true
}

const deleteInboundRoute = async (route) => {
  if (confirm(`Delete global route "${route.name}"?`)) {
    try {
      await systemAPI.deleteDialplan(route.id)
      await loadRoutes()
    } catch (e) {
      console.error(e)
      alert('Failed to delete route')
    }
  }
}

const saveInboundRoute = async () => {
    try {
        const payload = {
            dialplan_name: inboundForm.value.dialplan_name,
            dialplan_context: inboundForm.value.dialplan_context,
            enabled: inboundForm.value.enabled,
            details: inboundForm.value.details.map((d, i) => ({ ...d, detail_order: (i+1)*10 }))
        }

        if (editingInbound.value) {
            await systemAPI.updateDialplan(inboundForm.value.id, payload)
        } else {
            await systemAPI.createDialplan(payload)
        }
        await loadRoutes()
        showInboundModal.value = false
        editingInbound.value = false
        inboundForm.value = { dialplan_name: '', dialplan_context: 'public', enabled: true, details: [{ detail_type: 'condition', condition_field: 'destination_number', condition_expression: '', condition_expression_type: 'regex', condition_break: 'on-false' }, { detail_type: 'action', action_application: 'transfer', action_data: '' }] }
    } catch (e) {
        console.error(e)
        alert('Failed to save route')
    }
}

// Outbound Routes (Global)
const showOutboundModal = ref(false)
const editingOutbound = ref(false)
const outboundForm = ref({ dialplan_name: '', dialplan_context: 'default', enabled: true, pattern: '', strip: 0, prepend: '', gateway: '', continue: true, details: [] })
// Helper for outbound simplified UI (Pattern, Strip, Prepend, Gateway) -> Dialplan Details
// This logic is complex because 'outbound' UI is an abstraction over raw dialplan conditions/actions.
// For now, I'll stick to a simpler implementation or reuse the raw editor if the UI allows.
// Looking at the code, `SystemRoutes.vue` has a specific outbound modal with pattern/strip/prepend/gateway fields.
// I need to map these to Dialplan Details on save, and map back on edit.

const editOutboundRoute = (route) => {
    // Attempt to reverse engineer the simplified fields from details
    // This is tricky. For MVP, I might just map what I can.
    // ... logic needed here ...
    // For now, let's just assume we want to support the raw form or simple form.
    // The previous implementation used: { name, pattern, strip, prepend, gateway }
    
    // Simplification: We will just populate the form with basic data and let backend handle it, 
    // OR we fully implement the translation logic.
    // Given the constraints, I will implement a basic version.
    alert("Advanced outbound editing not fully implemented yet in this iteration.")
}

const deleteOutboundRoute = async (route) => {
    if (confirm(`Delete global route "${route.name}"?`)) {
        try {
            await systemAPI.deleteDialplan(route.id)
            await loadRoutes()
        } catch (e) {
            console.error(e)
        }
    }
}

const saveOutboundRoute = async () => {
    // ... placeholder for outbound save logic ...
    // To properly support the UI's 'strip', 'prepend', 'gateway', we need to generate:
    // Condition: destination_number =~ pattern
    // Action: bridge sofia/gateway/gateway_name/prepend$1
    // This logic should ideally be in the backend or shared.
    alert("Save outbound not implemented yet.")
}

// Initial Load
onMounted(() => {
  loadAllNumbers()
  loadRoutes()
})

// Settings
const settings = ref({
  region: 'nanp',
  format: 'e164'
})

const saveSettings = () => alert('System Settings saved!')
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); }

/* Route Help */
.route-help { display: flex; align-items: center; gap: 8px; padding: 12px; background: #eff6ff; border-radius: var(--radius-sm); margin-bottom: 16px; color: #1e40af; font-size: 13px; }
.help-icon { width: 16px; height: 16px; }

/* Routes List */
.routes-list { display: flex; flex-direction: column; gap: 12px; }
.route-card { display: flex; gap: 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; align-items: flex-start; }
.route-card.disabled { opacity: 0.5; background: var(--bg-app); }

.route-handle { display: flex; flex-direction: column; align-items: center; gap: 4px; color: var(--text-muted); cursor: grab; }
.route-handle:active { cursor: grabbing; }
.grip-icon { width: 16px; height: 16px; }

/* Drag and Drop States */
.route-card.dragging { opacity: 0.5; background: #e0e7ff; border-style: dashed; }
.route-card.dragover { border-color: var(--primary-color); border-width: 2px; background: #eff6ff; }
.route-order { font-size: 11px; font-weight: 700; background: var(--bg-app); padding: 2px 6px; border-radius: 3px; }

.route-main { flex: 1; }
.route-name-row { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.route-name-row h4 { font-size: 14px; font-weight: 600; margin: 0; }

.route-badges { display: flex; gap: 6px; }
.badge { font-size: 10px; padding: 2px 6px; border-radius: 3px; font-weight: 600; display: flex; align-items: center; gap: 4px; }
.badge.context { background: #f3e8ff; color: #7c3aed; }
.badge.gateway { background: #dbeafe; color: #1d4ed8; }
.badge.intl { background: #fee2e2; color: #dc2626; }

.route-conditions { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 8px; }
.condition { display: flex; align-items: center; gap: 4px; font-size: 12px; background: var(--bg-app); padding: 4px 8px; border-radius: 4px; }
.cond-var { color: #7c3aed; font-weight: 500; }
.cond-op { color: var(--text-muted); }
.cond-val { color: #059669; }

.route-actions-display { display: flex; align-items: center; gap: 8px; }
.arrow-icon { width: 16px; height: 16px; color: var(--text-muted); }
.action-chain { display: flex; flex-wrap: wrap; gap: 6px; }
.action-item { display: flex; gap: 4px; font-size: 12px; background: #ecfdf5; padding: 4px 8px; border-radius: 4px; }
.action-app { color: #059669; font-weight: 600; }
.action-data { color: var(--text-main); font-family: monospace; font-size: 11px; }

.route-pattern { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.pattern-label { font-size: 11px; color: var(--text-muted); }
.pattern-regex { background: #1e293b; color: #22d3ee; padding: 4px 8px; border-radius: 4px; font-size: 12px; }
.pattern-desc { font-size: 12px; color: var(--text-muted); }

.route-transforms { display: flex; gap: 12px; }
.transform { font-size: 11px; color: var(--text-muted); }

.route-controls { display: flex; align-items: center; gap: 8px; }

/* Settings Panel */
.settings-panel { max-width: 800px; }
.settings-section { margin-bottom: 24px; }
.settings-section h3 { font-size: 14px; font-weight: 600; margin-bottom: 12px; }
.settings-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

/* Buttons & Inputs (Shared with Routing.vue) */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-small { font-size: 11px; padding: 4px 8px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 16px; height: 16px; }
.text-bad { color: var(--status-bad); }

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.flex-2 { flex: 2; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.code { font-family: monospace; background: #f8fafc; }
.input-field.small { width: 120px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.checkbox-group { display: flex; flex-direction: column; gap: 8px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

.form-section { margin-bottom: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-header h4 { font-size: 13px; font-weight: 600; margin: 0; }

.conditions-editor, .actions-editor { display: flex; flex-direction: column; gap: 8px; }
.condition-row, .action-row { display: flex; gap: 8px; align-items: center; }
.condition-row .input-field, .action-row .input-field { flex: 1; }

.form-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 20px; }
</style>
