<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Global Dial Plans</h2>
      <p class="text-muted text-sm">System-wide routing rules applied to all tenants (Context: public/default).</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="openCreate">+ New Global Rule</button>
    </div>
  </div>

  <div class="route-help">
    <InfoIcon class="help-icon" />
    <span>Global dial plans are processed in order. First matching rule wins. Drag to reorder.</span>
  </div>

  <div class="routes-list" v-if="rules.length > 0">
    <div class="route-card" v-for="(rule, idx) in rules" :key="rule.id"
      :class="{ disabled: !rule.enabled }"
      draggable="true"
      @dragstart="onDragStart($event, idx)"
      @dragover.prevent="onDragOver($event, idx)"
      @dragleave="onDragLeave"
      @drop="onDrop($event, idx)"
      @dragend="onDragEnd">
      <div class="route-handle">
        <GripVerticalIcon class="grip-icon" />
        <span class="route-order">{{ rule.dialplan_order || idx + 1 }}</span>
      </div>
      
      <div class="route-main">
        <div class="route-name-row">
          <h4>{{ rule.dialplan_name }}</h4>
          <div class="route-badges">
            <span class="badge context">{{ rule.dialplan_context || 'public' }}</span>
            <span class="badge continue" v-if="rule.continue">Continue</span>
          </div>
        </div>
        
        <div class="route-conditions" v-if="rule.details && rule.details.length > 0">
          <div class="condition" v-for="(detail, i) in rule.details.filter(d => d.detail_type === 'condition').slice(0, 2)" :key="i">
            <span class="cond-var">{{ detail.condition_field }}</span>
            <span class="cond-op">=~</span>
            <span class="cond-val font-mono">{{ detail.condition_expression }}</span>
          </div>
        </div>
        
        <div class="route-description" v-if="rule.description">
          {{ rule.description }}
        </div>
        
        <div class="route-actions-display" v-if="rule.details && rule.details.length > 0">
          <ArrowRightIcon class="arrow-icon" />
          <div class="action-chain">
            <span class="action-item" v-for="(act, i) in rule.details.filter(d => d.detail_type === 'action').slice(0, 3)" :key="i">
              <span class="action-app">{{ act.action_application }}</span>
              <span class="action-data" v-if="act.action_data">{{ act.action_data }}</span>
            </span>
          </div>
        </div>
      </div>

      <div class="route-controls">
        <label class="switch small">
          <input type="checkbox" v-model="rule.enabled" @change="toggleRule(rule)">
          <span class="slider round"></span>
        </label>
        <button class="btn-icon" @click="editRule(rule)"><EditIcon class="icon-sm" /></button>
        <button class="btn-icon" @click="deleteRule(rule)"><TrashIcon class="icon-sm text-bad" /></button>
      </div>
    </div>
  </div>

  <div v-else class="empty-state">
    <p>No global dial plans configured yet.</p>
    <button class="btn-secondary" @click="openCreate">Create your first rule</button>
  </div>

  <!-- DIALPLAN MODAL -->
  <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>{{ isEditing ? 'Edit Global Rule' : 'New Global Rule' }}</h3>
        <button class="btn-icon" @click="showModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Rule Name</label>
            <input v-model="activeRule.dialplan_name" class="input-field" placeholder="e.g. intercept_all">
          </div>
          <div class="form-group">
            <label>Context</label>
            <select v-model="activeRule.dialplan_context" class="input-field">
              <option value="public">public</option>
              <option value="default">default</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Order / Priority</label>
            <input v-model.number="activeRule.dialplan_order" type="number" class="input-field" placeholder="100">
          </div>
          <div class="form-group">
            <label class="checkbox-row" style="margin-top: 24px;">
              <input type="checkbox" v-model="activeRule.continue">
              <span>Continue (Fallthrough)</span>
            </label>
          </div>
        </div>

        <div class="form-group">
          <label>Description</label>
          <input v-model="activeRule.description" class="input-field" placeholder="What does this rule do?">
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Conditions</h4>
            <button class="btn-small" @click="addCondition">+ Add Condition</button>
          </div>
          <p class="help-text">Match incoming calls based on FreeSWITCH channel variables.</p>
          
          <div class="conditions-editor">
            <div class="condition-row" v-for="(cond, i) in formConditions" :key="i">
              <select v-model="cond.condition_field" class="input-field">
                <option value="destination_number">destination_number</option>
                <option value="caller_id_number">caller_id_number</option>
                <option value="caller_id_name">caller_id_name</option>
                <option value="sip_to_user">sip_to_user</option>
                <option value="network_addr">network_addr</option>
              </select>
              <input v-model="cond.condition_expression" class="input-field code flex-2" placeholder="^\\+?1?(\\d{10})$">
              <button class="btn-icon" @click="removeCondition(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Actions</h4>
            <button class="btn-small" @click="addAction">+ Add Action</button>
          </div>
          <p class="help-text">Execute these FreeSWITCH actions in order.</p>
          
          <div class="actions-editor">
            <div class="action-row" v-for="(act, i) in formActions" :key="i">
              <select v-model="act.action_application" class="input-field">
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
              <input v-model="act.action_data" class="input-field flex-2" placeholder="application data">
              <button class="btn-icon" @click="removeAction(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveRule" :disabled="saving || !activeRule.dialplan_name">
          {{ saving ? 'Saving...' : 'Save Rule' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Info as InfoIcon, GripVertical as GripVerticalIcon,
  ArrowRight as ArrowRightIcon, Edit as EditIcon,
  Trash2 as TrashIcon, X as XIcon
} from 'lucide-vue-next'
import { systemAPI } from '../../services/api'

const rules = ref([])
const loading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)

const defaultRule = () => ({
  dialplan_order: 100,
  dialplan_name: '',
  dialplan_context: 'public',
  description: '',
  enabled: true,
  continue: false,
  details: []
})

const activeRule = ref(defaultRule())
const formConditions = ref([{ condition_field: 'destination_number', condition_expression: '', detail_type: 'condition', detail_order: 10 }])
const formActions = ref([{ action_application: 'transfer', action_data: '', detail_type: 'action', detail_order: 10 }])

// Drag state
const dragIndex = ref(null)
const dragOverIndex = ref(null)

const loadDialplans = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listDialplans()
    rules.value = response.data.data || response.data || []
  } catch (e) {
    console.error('Failed to load dial plans:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadDialplans)

const openCreate = () => {
  activeRule.value = defaultRule()
  formConditions.value = [{ condition_field: 'destination_number', condition_expression: '', detail_type: 'condition', detail_order: 10 }]
  formActions.value = [{ action_application: 'transfer', action_data: '', detail_type: 'action', detail_order: 10 }]
  isEditing.value = false
  showModal.value = true
}

const editRule = (rule) => {
  activeRule.value = { ...rule }
  formConditions.value = (rule.details || []).filter(d => d.detail_type === 'condition').map(d => ({ ...d }))
  formActions.value = (rule.details || []).filter(d => d.detail_type === 'action').map(d => ({ ...d }))
  if (formConditions.value.length === 0) formConditions.value = [{ condition_field: 'destination_number', condition_expression: '', detail_type: 'condition', detail_order: 10 }]
  if (formActions.value.length === 0) formActions.value = [{ action_application: 'transfer', action_data: '', detail_type: 'action', detail_order: 10 }]
  isEditing.value = true
  showModal.value = true
}

const saveRule = async () => {
  saving.value = true
  try {
    // Combine conditions and actions into details
    const details = [
      ...formConditions.value.filter(c => c.condition_expression).map((c, i) => ({ ...c, detail_order: (i + 1) * 10 })),
      ...formActions.value.filter(a => a.action_application).map((a, i) => ({ ...a, detail_order: (i + 1) * 10 + 100 }))
    ]
    const payload = { ...activeRule.value, details }
    
    if (isEditing.value && activeRule.value.id) {
      await systemAPI.updateDialplan(activeRule.value.id, payload)
    } else {
      await systemAPI.createDialplan(payload)
    }
    await loadDialplans()
    showModal.value = false
  } catch (e) {
    alert('Failed to save dial plan: ' + e.message)
  } finally {
    saving.value = false
  }
}

const deleteRule = async (rule) => {
  if (!confirm(`Delete dial plan "${rule.dialplan_name}"?`)) return
  try {
    await systemAPI.deleteDialplan(rule.id)
    await loadDialplans()
  } catch (e) {
    alert('Failed to delete dial plan: ' + e.message)
  }
}

const toggleRule = async (rule) => {
  try {
    await systemAPI.updateDialplan(rule.id, { enabled: rule.enabled })
  } catch (e) {
    alert('Failed to update dial plan: ' + e.message)
  }
}

const addCondition = () => formConditions.value.push({ condition_field: 'destination_number', condition_expression: '', detail_type: 'condition', detail_order: 10 })
const removeCondition = (i) => formConditions.value.splice(i, 1)
const addAction = () => formActions.value.push({ action_application: 'hangup', action_data: '', detail_type: 'action', detail_order: 10 })
const removeAction = (i) => formActions.value.splice(i, 1)

// Drag and drop
const onDragStart = (e, idx) => { dragIndex.value = idx; e.dataTransfer.effectAllowed = 'move' }
const onDragOver = (e, idx) => { dragOverIndex.value = idx; e.dataTransfer.dropEffect = 'move' }
const onDragLeave = () => { dragOverIndex.value = null }
const onDrop = async (e, idx) => {
  if (dragIndex.value === null || dragIndex.value === idx) return
  const item = rules.value.splice(dragIndex.value, 1)[0]
  rules.value.splice(idx, 0, item)
  // TODO: Update order in backend
  dragIndex.value = null
  dragOverIndex.value = null
}
const onDragEnd = () => { dragIndex.value = null; dragOverIndex.value = null }
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

/* Route Help */
.route-help { display: flex; align-items: center; gap: 8px; padding: 12px; background: #eff6ff; border-radius: var(--radius-sm); margin-bottom: 16px; color: #1e40af; font-size: 13px; }
.help-icon { width: 16px; height: 16px; }

/* Routes List */
.routes-list { display: flex; flex-direction: column; gap: 12px; }
.route-card { display: flex; gap: 16px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; align-items: flex-start; cursor: grab; }
.route-card.disabled { opacity: 0.5; background: var(--bg-app); }
.route-card:active { cursor: grabbing; }

.route-handle { display: flex; flex-direction: column; align-items: center; gap: 4px; color: var(--text-muted); }
.grip-icon { width: 16px; height: 16px; }
.route-order { font-size: 11px; font-weight: 700; background: var(--bg-app); padding: 2px 6px; border-radius: 3px; }

.route-main { flex: 1; }
.route-name-row { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.route-name-row h4 { font-size: 14px; font-weight: 600; margin: 0; }

.route-badges { display: flex; gap: 6px; }
.badge { font-size: 10px; padding: 2px 6px; border-radius: 3px; font-weight: 600; }
.badge.context { background: #f3e8ff; color: #7c3aed; }
.badge.continue { background: #dcfce7; color: #16a34a; }

.route-conditions { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 8px; }
.condition { display: flex; align-items: center; gap: 4px; font-size: 12px; background: var(--bg-app); padding: 4px 8px; border-radius: 4px; }
.cond-var { color: #7c3aed; font-weight: 500; }
.cond-op { color: var(--text-muted); }
.cond-val { color: #059669; }
.font-mono { font-family: monospace; }

.route-description { font-size: 12px; color: var(--text-muted); margin-bottom: 8px; }

.route-actions-display { display: flex; align-items: center; gap: 8px; }
.arrow-icon { width: 16px; height: 16px; color: var(--text-muted); }
.action-chain { display: flex; flex-wrap: wrap; gap: 6px; }
.action-item { display: flex; gap: 4px; font-size: 12px; background: #ecfdf5; padding: 4px 8px; border-radius: 4px; }
.action-app { color: #059669; font-weight: 600; }
.action-data { color: var(--text-main); font-family: monospace; font-size: 11px; }

.route-controls { display: flex; align-items: center; gap: 8px; }

.empty-state { text-align: center; padding: 40px; color: var(--text-muted); background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); }
.modal-card { background: white; border-radius: var(--radius-md); box-shadow: var(--shadow-lg); width: 100%; max-width: 520px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-card.large { max-width: 680px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Form Elements */
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.flex-2 { flex: 2; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.code { font-family: monospace; background: #f8fafc; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); margin-bottom: 8px; }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; text-transform: none; font-weight: 500; color: var(--text-main); }

.form-section { margin-bottom: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-header h4 { font-size: 13px; font-weight: 600; margin: 0; }

.conditions-editor, .actions-editor { display: flex; flex-direction: column; gap: 8px; }
.condition-row, .action-row { display: flex; gap: 8px; align-items: center; }
.condition-row .input-field, .action-row .input-field { flex: 1; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-small { font-size: 11px; padding: 4px 8px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 16px; height: 16px; }
.text-bad { color: var(--status-bad); }

/* Toggle Switch */
.switch { position: relative; display: inline-block; width: 34px; height: 20px; }
.switch.small { width: 30px; height: 18px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; border-radius: 34px; transition: 0.3s; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 2px; bottom: 2px; background-color: white; border-radius: 50%; transition: 0.3s; }
.switch.small .slider:before { height: 12px; width: 12px; }
.switch input:checked + .slider { background-color: var(--primary-color); }
.switch input:checked + .slider:before { transform: translateX(14px); }
.switch.small input:checked + .slider:before { transform: translateX(12px); }
</style>
