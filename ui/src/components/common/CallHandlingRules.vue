<template>
  <div class="call-handling-section">
    <div class="section-header">
      <div class="header-content">
        <h4>Call Handling Rules</h4>
        <p class="text-muted text-sm">Define custom routing based on events, conditions, and actions. By default, all registered devices ring.</p>
      </div>
      <button class="btn-primary small" @click="showRuleModal = true">+ Add Rule</button>
    </div>

    <div class="rules-list" v-if="rules.length">
      <div 
        class="rule-card" 
        v-for="(rule, idx) in rules" 
        :key="rule.id"
        :class="{ disabled: !rule.enabled, dragging: dragIndex === idx, dragover: dragOverIndex === idx }"
        draggable="true"
        @dragstart="onDragStart($event, idx)"
        @dragover.prevent="onDragOver($event, idx)"
        @dragleave="onDragLeave"
        @drop="onDrop($event, idx)"
        @dragend="onDragEnd"
      >
        <div class="rule-handle">
          <GripVerticalIcon class="grip-icon" />
          <span class="rule-order">{{ idx + 1 }}</span>
        </div>

        <div class="rule-content">
          <div class="rule-header">
            <span class="rule-name">{{ rule.name }}</span>
            <div class="rule-badges">
              <span class="badge event" v-for="evt in getActiveEvents(rule)" :key="evt">{{ evt }}</span>
              <span class="badge action" :class="rule.action_type">{{ formatAction(rule) }}</span>
            </div>
          </div>
          <p class="rule-description" v-if="rule.description">{{ rule.description }}</p>
          <div class="rule-conditions" v-if="rule.conditions?.length">
            <span class="condition" v-for="(cond, i) in rule.conditions" :key="i">
              {{ formatCondition(cond) }}
            </span>
          </div>
        </div>

        <div class="rule-actions">
          <label class="switch small">
            <input type="checkbox" v-model="rule.enabled" @change="toggleRule(rule)">
            <span class="slider round"></span>
          </label>
          <button class="btn-icon" @click="editRule(rule)"><EditIcon class="icon-sm" /></button>
          <button class="btn-icon" @click="deleteRule(rule)"><TrashIcon class="icon-sm text-bad" /></button>
        </div>
      </div>
    </div>

    <div class="empty-state" v-else>
      <PhoneIcon class="empty-icon" />
      <p>No call handling rules defined</p>
      <p class="text-muted text-sm">All registered devices will ring for incoming calls.</p>
    </div>

    <!-- Rule Editor Modal -->
    <div v-if="showRuleModal" class="modal-overlay" @click.self="showRuleModal = false">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>{{ editingRule ? 'Edit Rule' : 'New Call Handling Rule' }}</h3>
          <button class="btn-icon" @click="showRuleModal = false"><XIcon class="icon-sm" /></button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>Rule Name</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. After Hours Forward">
          </div>

          <div class="form-group">
            <label>Description (optional)</label>
            <input v-model="form.description" class="input-field" placeholder="Describe what this rule does">
          </div>

          <div class="divider"></div>

          <!-- Events Section -->
          <div class="form-section">
            <h4>Events (When to Apply)</h4>
            <div class="checkbox-grid">
              <label class="checkbox-card" :class="{ checked: form.events.any_call }">
                <input type="checkbox" v-model="form.events.any_call">
                <PhoneIncomingIcon class="checkbox-icon" />
                <span>Any Incoming Call</span>
              </label>
              <label class="checkbox-card" :class="{ checked: form.events.on_phone }">
                <input type="checkbox" v-model="form.events.on_phone">
                <PhoneCallIcon class="checkbox-icon" />
                <span>On Another Call</span>
              </label>
              <label class="checkbox-card" :class="{ checked: form.events.no_answer }">
                <input type="checkbox" v-model="form.events.no_answer">
                <PhoneMissedIcon class="checkbox-icon" />
                <span>No Answer</span>
              </label>
              <label class="checkbox-card" :class="{ checked: form.events.unavailable }">
                <input type="checkbox" v-model="form.events.unavailable">
                <PhoneOffIcon class="checkbox-icon" />
                <span>Unavailable/DND</span>
              </label>
            </div>
          </div>

          <div class="divider"></div>

          <!-- Conditions Section -->
          <div class="form-section">
            <div class="section-header-inline">
              <h4>Conditions (Additional Filters)</h4>
              <button class="btn-small" @click="addCondition">+ Add Condition</button>
            </div>
            <p class="help-text">Optional: Only apply this rule when all conditions match.</p>

            <div class="conditions-list">
              <div class="condition-row" v-for="(cond, i) in form.conditions" :key="i">
                <select v-model="cond.type" class="input-field">
                  <option value="caller_id">Caller ID</option>
                  <option value="caller_name">Caller Name</option>
                  <option value="time_of_day">Time of Day</option>
                  <option value="day_of_week">Day of Week</option>
                  <option value="date_range">Date Range</option>
                  <option value="holiday_list">Holiday List</option>
                </select>
                <select v-model="cond.op" class="input-field small">
                  <option value="equals">equals</option>
                  <option value="contains">contains</option>
                  <option value="starts_with">starts with</option>
                  <option value="regex">matches regex</option>
                  <option value="in">is in list</option>
                  <option value="between">between</option>
                </select>
                <input v-model="cond.value" class="input-field flex-1" :placeholder="getConditionPlaceholder(cond.type)">
                <button class="btn-icon" @click="removeCondition(i)"><XIcon class="icon-sm" /></button>
              </div>
            </div>
          </div>

          <div class="divider"></div>

          <!-- Action Section -->
          <div class="form-section">
            <h4>Action (What To Do)</h4>
            <div class="action-selector">
              <label class="action-option" :class="{ selected: form.action_type === 'ring_devices' }">
                <input type="radio" v-model="form.action_type" value="ring_devices">
                <PhoneIcon class="action-icon" />
                <div class="action-text">
                  <span class="action-name">Ring Devices</span>
                  <span class="action-desc">Ring specific devices only</span>
                </div>
              </label>
              <label class="action-option" :class="{ selected: form.action_type === 'forward' }">
                <input type="radio" v-model="form.action_type" value="forward">
                <PhoneForwardedIcon class="action-icon" />
                <div class="action-text">
                  <span class="action-name">Forward To</span>
                  <span class="action-desc">Forward to another number</span>
                </div>
              </label>
              <label class="action-option" :class="{ selected: form.action_type === 'voicemail' }">
                <input type="radio" v-model="form.action_type" value="voicemail">
                <VoicemailIcon class="action-icon" />
                <div class="action-text">
                  <span class="action-name">Voicemail</span>
                  <span class="action-desc">Send to voicemail</span>
                </div>
              </label>
              <label class="action-option" :class="{ selected: form.action_type === 'find_me' }">
                <input type="radio" v-model="form.action_type" value="find_me">
                <SearchIcon class="action-icon" />
                <div class="action-text">
                  <span class="action-name">Find Me</span>
                  <span class="action-desc">Ring multiple numbers</span>
                </div>
              </label>
              <label class="action-option" :class="{ selected: form.action_type === 'reject' }">
                <input type="radio" v-model="form.action_type" value="reject">
                <PhoneOffIcon class="action-icon" />
                <div class="action-text">
                  <span class="action-name">Reject</span>
                  <span class="action-desc">Reject the call</span>
                </div>
              </label>
            </div>

            <!-- Action-specific fields -->
            <div class="action-params" v-if="form.action_type === 'forward'">
              <div class="form-group">
                <label>Forward To Number</label>
                <input v-model="form.action_target" class="input-field" placeholder="+14155551234 or extension 101">
              </div>
              <div class="form-group">
                <label>Ring Timeout (seconds)</label>
                <input type="number" v-model="form.action_params.ring_timeout" class="input-field small" placeholder="30">
              </div>
            </div>

            <div class="action-params" v-if="form.action_type === 'voicemail'">
              <div class="form-group">
                <label>Voicemail Greeting</label>
                <select v-model="form.action_params.greeting_id" class="input-field">
                  <option value="">Default Greeting</option>
                  <option value="busy">Busy Greeting</option>
                  <option value="unavailable">Unavailable Greeting</option>
                </select>
              </div>
            </div>

            <div class="action-params" v-if="form.action_type === 'find_me'">
              <div class="form-group">
                <label>Ring Mode</label>
                <select v-model="form.action_params.find_me_mode" class="input-field">
                  <option value="simultaneous">Simultaneous (ring all at once)</option>
                  <option value="sequential">Sequential (ring one after another)</option>
                </select>
              </div>
              <div class="form-group">
                <label>Numbers to Ring</label>
                <div class="find-me-numbers">
                  <div class="find-me-row" v-for="(num, i) in form.action_params.find_me_numbers" :key="i">
                    <input v-model="form.action_params.find_me_numbers[i]" class="input-field" placeholder="+14155551234">
                    <button class="btn-icon" @click="removeFindMeNumber(i)"><XIcon class="icon-sm" /></button>
                  </div>
                  <button class="btn-small" @click="addFindMeNumber">+ Add Number</button>
                </div>
              </div>
              <div class="form-group" v-if="form.action_params.find_me_mode === 'sequential'">
                <label>Delay Between Rings (seconds)</label>
                <input type="number" v-model="form.action_params.find_me_delay" class="input-field small" placeholder="15">
              </div>
            </div>

            <div class="action-params" v-if="form.action_type === 'ring_devices'">
              <p class="help-text">Override default device ringing:</p>
              <div class="checkbox-row">
                <label><input type="checkbox" v-model="form.action_params.ring_devices.softphone"> Softphone</label>
                <label><input type="checkbox" v-model="form.action_params.ring_devices.desk_phone"> Desk Phone</label>
                <label><input type="checkbox" v-model="form.action_params.ring_devices.mobile"> Mobile</label>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="showRuleModal = false">Cancel</button>
          <button class="btn-primary" @click="saveRule" :disabled="!form.name || !hasAnyEvent">Save Rule</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import {
  GripVertical as GripVerticalIcon,
  Phone as PhoneIcon,
  PhoneIncoming as PhoneIncomingIcon,
  PhoneCall as PhoneCallIcon,
  PhoneMissed as PhoneMissedIcon,
  PhoneOff as PhoneOffIcon,
  PhoneForwarded as PhoneForwardedIcon,
  Voicemail as VoicemailIcon,
  Search as SearchIcon,
  Edit as EditIcon,
  Trash2 as TrashIcon,
  X as XIcon
} from 'lucide-vue-next'

const props = defineProps({
  extensionId: { type: [Number, String], default: null },
  profileId: { type: [Number, String], default: null },
  api: { type: Object, required: true } // extensionsAPI or extensionProfilesAPI
})

const emit = defineEmits(['updated'])

const rules = ref([])
const showRuleModal = ref(false)
const editingRule = ref(null)

const defaultForm = () => ({
  name: '',
  description: '',
  enabled: true,
  events: { on_phone: false, no_answer: false, any_call: false, unavailable: false },
  conditions: [],
  action_type: 'forward',
  action_target: '',
  action_params: {
    ring_timeout: 30,
    greeting_id: '',
    record_option: 'always',
    find_me_mode: 'simultaneous',
    find_me_numbers: [''],
    find_me_delay: 15,
    ring_devices: { softphone: true, desk_phone: true, mobile: true },
    reject_reason: 'unavailable'
  }
})

const form = ref(defaultForm())

const hasAnyEvent = computed(() => 
  form.value.events.on_phone || form.value.events.no_answer || 
  form.value.events.any_call || form.value.events.unavailable
)

// Drag and drop state
const dragIndex = ref(null)
const dragOverIndex = ref(null)

onMounted(() => loadRules())

const getOwnerId = () => props.extensionId || props.profileId

const loadRules = async () => {
  try {
    const response = await props.api.listCallRules(getOwnerId())
    rules.value = response.data?.data || []
  } catch (e) {
    console.error('Failed to load call handling rules', e)
  }
}

const saveRule = async () => {
  try {
    if (editingRule.value) {
      await props.api.updateCallRule(getOwnerId(), editingRule.value.id, form.value)
    } else {
      await props.api.createCallRule(getOwnerId(), form.value)
    }
    await loadRules()
    showRuleModal.value = false
    form.value = defaultForm()
    editingRule.value = null
    emit('updated')
  } catch (e) {
    console.error('Failed to save rule', e)
  }
}

const editRule = (rule) => {
  editingRule.value = rule
  form.value = JSON.parse(JSON.stringify(rule))
  if (!form.value.action_params) form.value.action_params = defaultForm().action_params
  showRuleModal.value = true
}

const deleteRule = async (rule) => {
  if (!confirm(`Delete rule "${rule.name}"?`)) return
  try {
    await props.api.deleteCallRule(getOwnerId(), rule.id)
    await loadRules()
    emit('updated')
  } catch (e) {
    console.error('Failed to delete rule', e)
  }
}

const toggleRule = async (rule) => {
  try {
    await props.api.updateCallRule(getOwnerId(), rule.id, { enabled: rule.enabled })
  } catch (e) {
    console.error('Failed to toggle rule', e)
  }
}

// Drag and drop handlers
const onDragStart = (e, idx) => { dragIndex.value = idx; e.dataTransfer.effectAllowed = 'move' }
const onDragOver = (e, idx) => { dragOverIndex.value = idx; e.dataTransfer.dropEffect = 'move' }
const onDragLeave = () => { dragOverIndex.value = null }
const onDragEnd = () => { dragIndex.value = null; dragOverIndex.value = null }

const onDrop = async (e, idx) => {
  if (dragIndex.value === null || dragIndex.value === idx) return
  const item = rules.value.splice(dragIndex.value, 1)[0]
  rules.value.splice(idx, 0, item)
  dragIndex.value = null
  dragOverIndex.value = null
  
  // Save new order
  const ruleIds = rules.value.map(r => r.id)
  await props.api.reorderCallRules(getOwnerId(), ruleIds)
}

// Condition helpers
const addCondition = () => form.value.conditions.push({ type: 'caller_id', op: 'equals', value: '' })
const removeCondition = (i) => form.value.conditions.splice(i, 1)
const getConditionPlaceholder = (type) => {
  const placeholders = {
    caller_id: '+14155551234 or pattern',
    caller_name: 'Name to match',
    time_of_day: '09:00-17:00',
    day_of_week: 'Mon,Tue,Wed,Thu,Fri',
    date_range: '2025-01-01 to 2025-12-31',
    holiday_list: 'Select holiday list'
  }
  return placeholders[type] || ''
}

// Find Me helpers
const addFindMeNumber = () => form.value.action_params.find_me_numbers.push('')
const removeFindMeNumber = (i) => form.value.action_params.find_me_numbers.splice(i, 1)

// Display formatters
const getActiveEvents = (rule) => {
  const events = []
  if (rule.events?.any_call) events.push('Any Call')
  if (rule.events?.on_phone) events.push('Busy')
  if (rule.events?.no_answer) events.push('No Answer')
  if (rule.events?.unavailable) events.push('Unavailable')
  return events
}

const formatAction = (rule) => {
  const actions = { forward: 'Forward', voicemail: 'Voicemail', find_me: 'Find Me', reject: 'Reject', ring_devices: 'Ring Devices' }
  return actions[rule.action_type] || rule.action_type
}

const formatCondition = (cond) => {
  return `${cond.type} ${cond.op} ${cond.value}`
}

watch(() => props.extensionId, loadRules)
watch(() => props.profileId, loadRules)
</script>

<style scoped>
.call-handling-section { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 20px; }
.section-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.header-content h4 { font-size: 14px; font-weight: 600; margin: 0 0 4px 0; }
.header-content .text-muted { margin: 0; }

.rules-list { display: flex; flex-direction: column; gap: 12px; }
.rule-card { display: flex; gap: 12px; padding: 12px; background: var(--bg-app); border: 1px solid var(--border-color); border-radius: var(--radius-sm); align-items: flex-start; transition: all 0.2s; }
.rule-card.disabled { opacity: 0.5; }
.rule-card.dragging { opacity: 0.5; background: #e0e7ff; border-style: dashed; }
.rule-card.dragover { border-color: var(--primary-color); border-width: 2px; }

.rule-handle { display: flex; flex-direction: column; align-items: center; gap: 4px; cursor: grab; color: var(--text-muted); }
.rule-handle:active { cursor: grabbing; }
.grip-icon { width: 14px; height: 14px; }
.rule-order { font-size: 10px; font-weight: 700; background: white; padding: 2px 6px; border-radius: 3px; }

.rule-content { flex: 1; }
.rule-header { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }
.rule-name { font-size: 13px; font-weight: 600; }
.rule-badges { display: flex; gap: 4px; }
.badge { font-size: 9px; padding: 2px 6px; border-radius: 3px; font-weight: 600; }
.badge.event { background: #dbeafe; color: #1d4ed8; }
.badge.action { background: #f3e8ff; color: #7c3aed; }
.badge.action.voicemail { background: #fef3c7; color: #b45309; }
.badge.action.reject { background: #fee2e2; color: #dc2626; }

.rule-description { font-size: 12px; color: var(--text-muted); margin: 4px 0; }
.rule-conditions { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 6px; }
.condition { font-size: 10px; padding: 2px 6px; background: #f1f5f9; border-radius: 3px; color: var(--text-muted); }

.rule-actions { display: flex; align-items: center; gap: 8px; }

.empty-state { text-align: center; padding: 40px 20px; color: var(--text-muted); }
.empty-icon { width: 40px; height: 40px; margin-bottom: 12px; opacity: 0.3; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 640px; max-height: 90vh; overflow-y: auto; }
.modal-card.large { max-width: 700px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); position: sticky; top: 0; background: white; }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); position: sticky; bottom: 0; background: white; }

.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.form-section { margin-bottom: 16px; }
.form-section h4 { font-size: 13px; font-weight: 600; margin: 0 0 12px 0; }
.section-header-inline { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.input-field { width: 100%; padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.input-field.small { width: 120px; }
.input-field.flex-1 { flex: 1; }
.help-text { font-size: 11px; color: var(--text-muted); margin: 4px 0; }
.divider { height: 1px; background: var(--border-color); margin: 20px 0; }

/* Events Checkbox Grid */
.checkbox-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 10px; }
.checkbox-card { display: flex; align-items: center; gap: 10px; padding: 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); cursor: pointer; transition: all 0.2s; }
.checkbox-card:hover { border-color: var(--primary-color); }
.checkbox-card.checked { background: #eff6ff; border-color: var(--primary-color); }
.checkbox-card input { display: none; }
.checkbox-icon { width: 18px; height: 18px; color: var(--text-muted); }
.checkbox-card.checked .checkbox-icon { color: var(--primary-color); }

/* Conditions */
.conditions-list { display: flex; flex-direction: column; gap: 8px; margin-top: 10px; }
.condition-row { display: flex; gap: 8px; align-items: center; }

/* Action Selector */
.action-selector { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; }
.action-option { display: flex; align-items: center; gap: 12px; padding: 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); cursor: pointer; transition: all 0.2s; }
.action-option:hover { border-color: var(--primary-color); }
.action-option.selected { background: #eff6ff; border-color: var(--primary-color); }
.action-option input { display: none; }
.action-icon { width: 20px; height: 20px; color: var(--text-muted); }
.action-option.selected .action-icon { color: var(--primary-color); }
.action-text { display: flex; flex-direction: column; }
.action-name { font-size: 13px; font-weight: 600; }
.action-desc { font-size: 11px; color: var(--text-muted); }

.action-params { padding: 16px; background: var(--bg-app); border-radius: var(--radius-sm); margin-top: 12px; }
.find-me-numbers { display: flex; flex-direction: column; gap: 8px; }
.find-me-row { display: flex; gap: 8px; }
.checkbox-row { display: flex; gap: 16px; }
.checkbox-row label { display: flex; align-items: center; gap: 6px; font-size: 13px; cursor: pointer; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary.small { padding: 6px 12px; font-size: 12px; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; cursor: pointer; }
.btn-small { font-size: 11px; padding: 4px 8px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 14px; height: 14px; }
.text-bad { color: var(--status-bad); }
.text-muted { color: var(--text-muted); }
.text-sm { font-size: 12px; }

/* Switch */
.switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.switch.small { width: 32px; height: 18px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; inset: 0; background-color: #ccc; transition: 0.3s; }
.slider.round { border-radius: 20px; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 3px; bottom: 3px; background-color: white; transition: 0.3s; border-radius: 50%; }
.switch.small .slider:before { height: 12px; width: 12px; left: 3px; bottom: 3px; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(16px); }
.switch.small input:checked + .slider:before { transform: translateX(14px); }
</style>
