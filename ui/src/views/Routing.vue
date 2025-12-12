<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Call Routing</h2>
      <p class="text-muted text-sm">Manage inbound/outbound routing and dialplan logic (FreeSWITCH-compatible).</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" v-if="activeTab === 'numbers'" @click="showAddNumberModal = true">+ Add Number</button>
      <button class="btn-primary" v-if="activeTab === 'inbound'" @click="showInboundModal = true">+ Inbound Route</button>
      <button class="btn-primary" v-if="activeTab === 'outbound'" @click="showOutboundModal = true">+ Outbound Route</button>
    </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'numbers' }" @click="activeTab = 'numbers'">Phone Numbers</button>
    <button class="tab" :class="{ active: activeTab === 'inbound' }" @click="activeTab = 'inbound'">Inbound Routes</button>
    <button class="tab" :class="{ active: activeTab === 'outbound' }" @click="activeTab = 'outbound'">Outbound Routes</button>
    <button class="tab" :class="{ active: activeTab === 'settings' }" @click="activeTab = 'settings'">Settings</button>
  </div>

  <!-- NUMBERS TAB -->
  <div class="tab-content" v-if="activeTab === 'numbers'">
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input type="text" v-model="numberSearch" placeholder="Search numbers..." class="search-input">
      </div>
      <select v-model="numberFilter" class="filter-select">
        <option value="">All Types</option>
        <option value="Voice">Voice</option>
        <option value="Fax">Fax</option>
        <option value="SMS">SMS</option>
      </select>
    </div>

    <DataTable :columns="numberColumns" :data="filteredNumbers" actions>
      <template #number="{ value }">
        <span class="font-mono font-semibold">{{ formatPhoneNumber(value) }}</span>
      </template>
      <template #capabilities="{ row }">
        <div class="cap-badges">
          <span class="cap-badge voice" v-if="row.voice">Voice</span>
          <span class="cap-badge sms" v-if="row.sms">SMS</span>
          <span class="cap-badge fax" v-if="row.fax">Fax</span>
        </div>
      </template>
      <template #routing="{ value, row }">
        <div class="route-display">
          <span class="route-target">{{ value }}</span>
          <span class="route-context" v-if="row.context">ctx: {{ row.context }}</span>
        </div>
      </template>
      <template #smsProvider="{ value }">
        <span v-if="value" class="sms-provider-badge">{{ value }}</span>
        <span v-else class="sms-na">N/A</span>
      </template>
      <template #status="{ value }">
        <StatusBadge :status="value" />
      </template>
      <template #actions="{ row }">
        <button class="btn-link" @click="editNumber(row)">Edit</button>
        <button class="btn-link" @click="viewNumberStats(row)">Stats</button>
      </template>
    </DataTable>
  </div>

  <!-- INBOUND ROUTES TAB -->
  <div class="tab-content" v-else-if="activeTab === 'inbound'">
    <div class="route-help">
      <InfoIcon class="help-icon" />
      <span>Inbound routes process incoming calls in order. First matching route wins. Drag to reorder.</span>
    </div>

    <div class="routes-list">
      <div class="route-card" v-for="(route, idx) in inboundRoutes" :key="route.id" 
        :class="{ disabled: !route.enabled, dragging: dragIndex === idx, dragover: dragOverIndex === idx }"
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
              <span class="badge time" v-if="route.timeCondition">
                <ClockIcon class="badge-icon" /> {{ route.timeCondition }}
              </span>
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
      <span>Outbound routes match dialed numbers and determine which gateway/trunk to use.</span>
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
      <h3>Dialing Preferences</h3>
      <div class="settings-grid">
        <div class="form-group">
          <label>Primary Region</label>
          <select class="input-field" v-model="settings.region">
            <option value="nanp">North America (NANP)</option>
            <option value="uk">United Kingdom</option>
            <option value="eu">Europe (General)</option>
            <option value="au">Australia</option>
          </select>
          <span class="help-text">Defines 10/11 digit interpretation.</span>
        </div>
        <div class="form-group">
          <label>Outbound Format</label>
          <select class="input-field" v-model="settings.format">
            <option value="e164">E.164 (Global Standard, +1...)</option>
            <option value="national">National (10-digit)</option>
            <option value="passthrough">Passthrough (As Dialed)</option>
          </select>
          <span class="help-text">Convert dialed digits for carrier.</span>
        </div>
      </div>
    </div>

    <div class="settings-section">
      <h3>Default Contexts</h3>
      <div class="settings-grid">
        <div class="form-group">
          <label>Default Inbound Context</label>
          <input type="text" class="input-field code" v-model="settings.inboundContext" placeholder="public">
        </div>
        <div class="form-group">
          <label>Default Outbound Context</label>
          <input type="text" class="input-field code" v-model="settings.outboundContext" placeholder="default">
        </div>
      </div>
    </div>

    <div class="settings-section">
      <h3>International Dialing</h3>
      <div class="checkbox-group">
        <label class="checkbox-row">
          <input type="checkbox" v-model="settings.support011">
          <span>Support 011 prefix (convert to E.164)</span>
        </label>
        <label class="checkbox-row">
          <input type="checkbox" v-model="settings.blockIntlByDefault">
          <span>Block international by default (require profile permission)</span>
        </label>
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
        <h3>{{ editingInbound ? 'Edit Inbound Route' : 'New Inbound Route' }}</h3>
        <button class="btn-icon" @click="showInboundModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Route Name</label>
            <input v-model="inboundForm.name" class="input-field" placeholder="e.g. Main Line to IVR">
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
          <p class="help-text">Match incoming calls based on FreeSWITCH channel variables.</p>
          
          <div class="conditions-editor">
            <div class="condition-row" v-for="(cond, i) in inboundForm.conditions" :key="i">
              <select v-model="cond.variable" class="input-field">
                <option value="destination_number">destination_number</option>
                <option value="caller_id_number">caller_id_number</option>
                <option value="caller_id_name">caller_id_name</option>
                <option value="sip_to_user">sip_to_user</option>
                <option value="sip_from_host">sip_from_host</option>
                <option value="network_addr">network_addr</option>
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

        <div class="form-section">
          <div class="section-header">
            <h4>Time Condition (Optional)</h4>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Active Hours</label>
              <select v-model="inboundForm.timeCondition" class="input-field">
                <option value="">Always Active</option>
                <option value="business">Business Hours (Mon-Fri 9-5)</option>
                <option value="afterhours">After Hours</option>
                <option value="weekends">Weekends Only</option>
                <option value="custom">Custom Schedule...</option>
              </select>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Actions</h4>
            <button class="btn-small" @click="addInboundAction">+ Add Action</button>
          </div>
          <p class="help-text">Execute these FreeSWITCH actions in order.</p>
          
          <div class="actions-editor">
            <div class="action-row" v-for="(act, i) in inboundForm.actions" :key="i">
              <select v-model="act.app" class="input-field">
                <optgroup label="Routing">
                  <option value="transfer">transfer</option>
                  <option value="bridge">bridge</option>
                </optgroup>
                <optgroup label="Media">
                  <option value="answer">answer</option>
                  <option value="playback">playback</option>
                  <option value="ivr">ivr (menu)</option>
                </optgroup>
                <optgroup label="Variables">
                  <option value="set">set</option>
                  <option value="export">export</option>
                </optgroup>
                <optgroup label="Call Control">
                  <option value="hangup">hangup</option>
                  <option value="voicemail">voicemail</option>
                  <option value="ring_ready">ring_ready</option>
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
        <h3>{{ editingOutbound ? 'Edit Outbound Route' : 'New Outbound Route' }}</h3>
        <button class="btn-icon" @click="showOutboundModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Route Name</label>
          <input v-model="outboundForm.name" class="input-field" placeholder="e.g. US Local Calls">
        </div>

        <div class="form-group">
          <label>Pattern (Regex)</label>
          <input v-model="outboundForm.pattern" class="input-field code" placeholder="^\\+?1?([2-9]\\d{9})$">
          <span class="help-text">Match dialed digits. Use capture groups for substitution.</span>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Strip Digits</label>
            <input type="number" v-model="outboundForm.strip" class="input-field" placeholder="0">
          </div>
          <div class="form-group">
            <label>Prepend</label>
            <input v-model="outboundForm.prepend" class="input-field code" placeholder="+1">
          </div>
        </div>

        <div class="form-group">
          <label>Gateway / Trunk</label>
          <select v-model="outboundForm.gateway" class="input-field">
            <option value="">Select Gateway...</option>
            <option value="flowroute">Flowroute Primary</option>
            <option value="twilio">Twilio Elastic SIP</option>
            <option value="telnyx">Telnyx Backup</option>
            <option value="local_pri">Local PRI Gateway</option>
          </select>
        </div>

        <div class="checkbox-group">
          <label class="checkbox-row">
            <input type="checkbox" v-model="outboundForm.international">
            <span>This route handles international calls</span>
          </label>
          <label class="checkbox-row">
            <input type="checkbox" v-model="outboundForm.continue">
            <span>Continue to next route on failure</span>
          </label>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showOutboundModal = false">Cancel</button>
        <button class="btn-primary" @click="saveOutboundRoute" :disabled="!outboundForm.name || !outboundForm.pattern">Save Route</button>
      </div>
    </div>
  </div>

  <!-- ADD NUMBER MODAL -->
  <div v-if="showAddNumberModal" class="modal-overlay" @click.self="showAddNumberModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>Add Phone Number</h3>
        <button class="btn-icon" @click="showAddNumberModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Phone Number (DID)</label>
          <input v-model="newNumber.number" class="input-field code" placeholder="+14155551234">
        </div>
        
        <div class="form-group">
          <label>Capabilities</label>
          <div class="checkbox-group horizontal">
            <label class="checkbox-row"><input type="checkbox" v-model="newNumber.voice"> Voice</label>
            <label class="checkbox-row"><input type="checkbox" v-model="newNumber.sms"> SMS</label>
            <label class="checkbox-row"><input type="checkbox" v-model="newNumber.fax"> Fax</label>
          </div>
        </div>

        <div class="form-group">
          <label>Inbound Context</label>
          <input v-model="newNumber.context" class="input-field code" placeholder="public">
        </div>

        <div class="form-group">
          <label>Default Routing</label>
          <select v-model="newNumber.routeType" class="input-field">
            <option value="extension">Extension</option>
            <option value="ivr">IVR / Auto Attendant</option>
            <option value="ring_group">Ring Group</option>
            <option value="queue">Call Queue</option>
            <option value="voicemail">Voicemail</option>
            <option value="custom">Custom Dialplan</option>
          </select>
        </div>

        <div class="form-group" v-if="newNumber.routeType !== 'custom'">
          <label>Target</label>
          <input v-model="newNumber.target" class="input-field" placeholder="101 or ivr:main_menu">
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showAddNumberModal = false">Cancel</button>
        <button class="btn-primary" @click="saveNewNumber" :disabled="!newNumber.number">Add Number</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Search as SearchIcon, Info as InfoIcon, GripVertical as GripVerticalIcon,
  Clock as ClockIcon, ArrowRight as ArrowRightIcon, Edit as EditIcon,
  Trash2 as TrashIcon, X as XIcon
} from 'lucide-vue-next'
import DataTable from '../components/common/DataTable.vue'
import StatusBadge from '../components/common/StatusBadge.vue'

const activeTab = ref('inbound')

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

// Inbound Routes
const showInboundModal = ref(false)
const editingInbound = ref(false)
const inboundForm = ref({
  name: '',
  context: 'public',
  conditions: [{ variable: 'destination_number', operator: '=~', value: '' }],
  actions: [{ app: 'transfer', data: '' }]
})

const inboundRoutes = ref([])

const loadInboundRoutes = async () => {
  try {
    const response = await routingAPI.listInbound()
    inboundRoutes.value = response.data.data || []
  } catch (e) {
    console.error('Failed to load inbound routes', e)
  }
}

// Numbers
import { formatPhoneNumber } from '../utils/formatters'
import { numbersAPI } from '../services/api'

const numberSearch = ref('')
const numberFilter = ref('')
const showAddNumberModal = ref(false)
const newNumber = ref({ number: '', voice: true, sms: false, fax: false, context: 'public', routeType: 'extension', target: '' })
const numbers = ref([])

const numberColumns = [
  { key: 'destination_number', label: 'Number', width: '150px' },
  { key: 'capabilities', label: 'Capabilities', width: '140px' },
  { key: 'routing', label: 'Voice Routing', width: '160px' },
  { key: 'smsProvider', label: 'SMS Provider', width: '130px' },
  { key: 'carrier', label: 'Carrier', width: '100px' },
  { key: 'status', label: 'Status', width: '90px' }
]

const loadNumbers = async () => {
    try {
        const response = await numbersAPI.list()
        numbers.value = response.data.data || []
    } catch (e) {
        console.error('Failed to load numbers', e)
    }
}

const filteredNumbers = computed(() => {
  return numbers.value.filter(n => {
    const numStr = n.destination_number || ''
    const matchesSearch = numStr.includes(numberSearch.value)
    const matchesFilter = !numberFilter.value || 
      (numberFilter.value === 'Voice' && n.enabled) // Basic filter assumption
    return matchesSearch && matchesFilter
  })
})

const editNumber = (row) => alert(`Editing not fully implemented yet for ${row.destination_number}`)
const viewNumberStats = (row) => alert(`Stats for ${row.destination_number}`)

const saveNewNumber = async () => {
  try {
      const payload = {
        destination_number: newNumber.value.number,
        destination_type: newNumber.value.routeType,
        destination_action: newNumber.value.target,
        context: newNumber.value.context,
        enabled: true, // Auto enable
        // Other fields like voice/sms flags would need to be in model if we want to store them separately
        // For now, assuming standard setup
      }
      await numbersAPI.create(payload)
      await loadNumbers()
      showAddNumberModal.value = false
      newNumber.value = { number: '', voice: true, sms: false, fax: false, context: 'public', routeType: 'extension', target: '' }
  } catch (e) {
      console.error(e)
      alert('Failed to create number')
  }
}

const outboundRoutes = ref([])

const loadOutboundRoutes = async () => {
    try {
        const response = await routingAPI.listOutbound()
        outboundRoutes.value = response.data.data || []
    } catch (e) {
        console.error(e)
    }
}

const outboundForm = ref({ name: '', pattern: '', strip: 0, prepend: '', gateway: '', continue: true, international: false, enabled: true })

const editOutboundRoute = (route) => {
    // Map existing structure back to simple form if possible, or warn
    // Simplified logic: Just copy known fields. Complex fields might be lost in simple editor.
    outboundForm.value = { ...route }
    editingOutbound.value = true
    showOutboundModal.value = true
}

const deleteOutboundRoute = async (route) => {
    if (confirm(`Delete route "${route.name}"?`)) {
        try {
            await dialPlansAPI.delete(route.id)
            await loadOutboundRoutes()
        } catch (e) {
            console.error(e)
            alert('Failed to delete route')
        }
    }
}

const saveOutboundRoute = async () => {
    try {
        // We need to construct the dialplan model based on the simple UI
        // Pattern -> Condition destination_number
        // Strip/Prepend/Gateway -> Bridge action with manipulation
        
        // This logic is complex to do purely frontend if we want to match backend's model structure rigidly.
        // Ideally backend has a "SimpleOutbound" endpoint, OR we construct the detailed Dialplan object here.
        // Going with constructing the Dialplan object for the generic Create/Update endpoint.
        
        const details = []
        let order = 10
        // Condition
        details.push({ detail_type: 'condition', condition_field: 'destination_number', condition_expression: outboundForm.value.pattern, condition_break: 'on-false', detail_order: order++ })
        
        // Action: Bridge
        // Logic: sofia/gateway/{gateway}/{prepend}{destination_number:strip}
        // This is pseudo-code logic for FreeSWITCH.
        // Actually simplest is: bridge sofia/gateway/${gateway}/${prepend}${destination_number:${strip}}
        // But for regex pattern match $1, usually we bridge to $1.
        
        // Let's assume the user regex has a capture group for the number part we want.
        // bridge data: sofia/gateway/${gateway_name}/${prepend}$1
        
        const bridgeData = `sofia/gateway/${outboundForm.value.gateway}/${outboundForm.value.prepend}$1`
        details.push({ detail_type: 'action', action_application: 'bridge', action_data: bridgeData, detail_order: order++ })

        const payload = {
            dialplan_name: outboundForm.value.name,
            dialplan_context: 'default',
            enabled: outboundForm.value.enabled,
            continue: outboundForm.value.continue,
            details: details
        }
        
        if (editingOutbound.value) {
            await dialPlansAPI.update(outboundForm.value.id, payload)
        } else {
            await routingAPI.createOutbound(payload)
        }
        
        await loadOutboundRoutes()
        showOutboundModal.value = false
        editingOutbound.value = false
        outboundForm.value = { name: '', pattern: '', strip: 0, prepend: '', gateway: '', continue: true, international: false, enabled: true }
    } catch (e) {
        console.error(e)
        alert('Failed to save outbound route')
    }
}

const createDefaultRoutes = async () => {
    if (confirm('Create standard US/CA (10/11 digit) and Emergency routes? This may duplicate existing routes.')) {
        try {
            await routingAPI.createDefaultOutbound()
            await loadOutboundRoutes()
        } catch (e) {
            console.error(e)
            alert('Failed to create default routes')
        }
    }
}

// Initial Load
onMounted(() => {
    loadNumbers()
    loadInboundRoutes()
    loadOutboundRoutes()
})

const addInboundCondition = () => inboundForm.value.conditions.push({ variable: 'destination_number', operator: '=~', value: '' })
const removeInboundCondition = (i) => inboundForm.value.conditions.splice(i, 1)
const addInboundAction = () => inboundForm.value.actions.push({ app: 'log', data: 'INFO' })
const removeInboundAction = (i) => inboundForm.value.actions.splice(i, 1)

const editInboundRoute = (route) => {
  inboundForm.value = JSON.parse(JSON.stringify(route))
  editingInbound.value = true
  showInboundModal.value = true
}

const deleteInboundRoute = async (route) => {
  if (confirm(`Delete inbound route "${route.name}"?`)) {
    try {
      await dialPlansAPI.delete(route.id)
      await loadInboundRoutes()
    } catch (e) {
      console.error(e)
      alert('Failed to delete route')
    }
  }
}

const saveInboundRoute = async () => {
  try {
    if (editingInbound.value) {
      await dialPlansAPI.update(inboundForm.value.id, inboundForm.value)
    } else {
      await routingAPI.createInbound(inboundForm.value)
    }
    await loadInboundRoutes()
    showInboundModal.value = false
    editingInbound.value = false
    inboundForm.value = { name: '', context: 'public', conditions: [{ variable: 'destination_number', operator: '=~', value: '' }], actions: [{ app: 'transfer', data: '' }] }
  } catch (e) {
    console.error(e)
    alert('Failed to save route')
  }
}

// Settings
const settings = ref({
  region: 'nanp',
  format: 'e164',
  inboundContext: 'public',
  outboundContext: 'default',
  support011: true,
  blockIntlByDefault: true
})
const saveSettings = () => alert('Settings saved!')
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); }

/* Filter Bar */
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-box { position: relative; flex: 1; max-width: 280px; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 12px 8px 34px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); background: white; }

/* Capability Badges */
.cap-badges { display: flex; gap: 4px; }
.cap-badge { font-size: 10px; padding: 2px 6px; border-radius: 3px; font-weight: 600; }
.cap-badge.voice { background: #dbeafe; color: #1d4ed8; }
.cap-badge.sms { background: #dcfce7; color: #16a34a; }
.cap-badge.fax { background: #fef3c7; color: #b45309; }

.route-display { display: flex; flex-direction: column; }
.route-target { font-weight: 500; }
.route-context { font-size: 10px; color: var(--text-muted); font-family: monospace; }

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
.badge.time { background: #fef3c7; color: #b45309; }
.badge.gateway { background: #dbeafe; color: #1d4ed8; }
.badge.intl { background: #fee2e2; color: #dc2626; }
.badge-icon { width: 10px; height: 10px; }

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

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-small { font-size: 11px; padding: 4px 8px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: var(--text-xs); margin-right: 8px; cursor: pointer; font-weight: 500; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 16px; height: 16px; }
.text-bad { color: var(--status-bad); }

/* Form Elements */
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
.checkbox-group.horizontal { flex-direction: row; gap: 16px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

.form-section { margin-bottom: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-header h4 { font-size: 13px; font-weight: 600; margin: 0; }

.conditions-editor, .actions-editor { display: flex; flex-direction: column; gap: 8px; }
.condition-row, .action-row { display: flex; gap: 8px; align-items: center; }
.condition-row .input-field, .action-row .input-field { flex: 1; }

.form-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 20px; }

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
.switch.small { width: 32px; height: 18px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: var(--border-color); transition: .3s; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 3px; bottom: 3px; background-color: white; transition: .3s; }
.switch.small .slider:before { height: 12px; width: 12px; left: 3px; bottom: 3px; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(16px); }
.switch.small input:checked + .slider:before { transform: translateX(14px); }
.slider.round { border-radius: 20px; }
.slider.round:before { border-radius: 50%; }

.font-mono { font-family: monospace; }
.font-semibold { font-weight: 600; }

.sms-provider-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
  background: #dbeafe;
  color: #2563eb;
}
.sms-na {
  font-size: 11px;
  color: var(--text-muted);
}
</style>
