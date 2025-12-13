<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/admin/time-conditions')">← Back to Time Conditions</button>
      <h2>{{ isNew ? 'New Time Condition' : `Editing: ${form.name}` }}</h2>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="testCondition">
        <PlayIcon class="btn-icon-left" />
        Test Now
      </button>
      <button class="btn-primary" @click="saveCondition">Save Condition</button>
    </div>
  </div>

  <div class="form-layout">
    <div class="form-main">
      <!-- Basic Info -->
      <div class="form-section">
        <h3>Basic Information</h3>
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Condition Name</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. Business Hours">
          </div>
          <div class="form-group">
            <label>Priority</label>
            <input type="number" v-model="form.priority" class="input-field" min="1" max="100">
            <span class="help-text">Lower numbers evaluated first</span>
          </div>
        </div>
      </div>

      <!-- Schedule Editor -->
      <div class="form-section">
        <div class="section-header">
          <h3>Schedule Rules</h3>
          <button class="btn-secondary small" @click="addRule">+ Add Rule</button>
        </div>
        <p class="text-muted text-sm mb-md">Define when this condition should match.</p>

        <div class="schedule-editor">
          <div class="rule-card" v-for="(rule, i) in form.rules" :key="i">
            <div class="rule-header">
              <span class="rule-number">Rule {{ i + 1 }}</span>
              <button class="btn-icon text-bad" @click="removeRule(i)" v-if="form.rules.length > 1">
                <TrashIcon class="icon-sm" />
              </button>
            </div>

            <div class="days-picker">
              <label class="day-btn" v-for="d in allDays" :key="d">
                <input type="checkbox" :value="d" v-model="rule.days">
                <span>{{ d }}</span>
              </label>
            </div>

            <div class="time-range">
              <div class="time-field">
                <label>Start Time</label>
                <input type="time" v-model="rule.startTime" class="input-field">
              </div>
              <span class="time-arrow">→</span>
              <div class="time-field">
                <label>End Time</label>
                <input type="time" v-model="rule.endTime" class="input-field">
              </div>
            </div>

            <!-- Visual preview -->
            <div class="rule-preview">
              <div class="hour-labels">
                <span v-for="h in [0,6,12,18,24]" :key="h">{{ h }}:00</span>
              </div>
              <div class="hour-bar">
                <div class="hour-fill" :style="getTimeBlockStyle(rule)"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Templates -->
        <div class="quick-templates">
          <span class="template-label">Quick:</span>
          <button class="template-btn" @click="applyTemplate('business')">9-5 Weekdays</button>
          <button class="template-btn" @click="applyTemplate('extended')">8-8 Weekdays</button>
          <button class="template-btn" @click="applyTemplate('24x7')">24/7</button>
          <button class="template-btn" @click="applyTemplate('weekend')">Weekend Only</button>
        </div>
      </div>

      <!-- Holiday Override -->
      <div class="form-section">
        <h3>Holiday Override</h3>
        <p class="text-muted text-sm mb-md">Select a holiday list that should override this condition.</p>
        
        <div class="form-row">
          <div class="form-group flex-1">
            <label>Holiday List</label>
            <select v-model="form.holiday_list_id" class="input-field">
              <option value="">-- None --</option>
              <option v-for="list in holidayLists" :key="list.id" :value="list.id">
                {{ list.name }}
              </option>
            </select>
          </div>
          <div class="form-group flex-1" v-if="form.holiday_list_id">
            <label>Holiday Destination</label>
            <div class="dest-form">
              <select v-model="form.holidayDestType" class="input-field">
                <option value="voicemail">Voicemail</option>
                <option value="extension">Extension</option>
                <option value="ivr">IVR Menu</option>
              </select>
              <input v-model="form.holidayDestValue" class="input-field" placeholder="e.g. 5900">
            </div>
          </div>
        </div>
      </div>

      <!-- Destinations -->
      <div class="form-section">
        <h3>Routing Destinations</h3>
        
        <div class="dest-config match">
          <div class="dest-indicator match">
            <CheckCircleIcon class="icon-sm" />
            When Matched
          </div>
          <div class="dest-form">
            <div class="form-group">
              <label>Destination Type</label>
              <select v-model="form.matchType" class="input-field">
                <option value="ivr">IVR Menu</option>
                <option value="extension">Extension</option>
                <option value="ring_group">Ring Group</option>
                <option value="queue">Queue</option>
                <option value="voicemail">Voicemail</option>
                <option value="external">External Number</option>
                <option value="continue">Continue to Next Condition</option>
              </select>
            </div>
            <div class="form-group flex-1" v-if="form.matchType !== 'continue'">
              <label>Target</label>
              <input v-model="form.matchTarget" class="input-field" :placeholder="getPlaceholder(form.matchType)">
            </div>
          </div>
        </div>

        <div class="dest-config nomatch">
          <div class="dest-indicator nomatch">
            <XCircleIcon class="icon-sm" />
            When NOT Matched
          </div>
          <div class="dest-form">
            <div class="form-group">
              <label>Destination Type</label>
              <select v-model="form.noMatchType" class="input-field">
                <option value="ivr">IVR Menu</option>
                <option value="extension">Extension</option>
                <option value="ring_group">Ring Group</option>
                <option value="queue">Queue</option>
                <option value="voicemail">Voicemail</option>
                <option value="external">External Number</option>
                <option value="continue">Continue to Next Condition</option>
              </select>
            </div>
            <div class="form-group flex-1" v-if="form.noMatchType !== 'continue'">
              <label>Target</label>
              <input v-model="form.noMatchTarget" class="input-field" :placeholder="getPlaceholder(form.noMatchType)">
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sidebar Preview -->
    <div class="form-sidebar">
      <div class="preview-card">
        <h4>Preview</h4>
        <div class="preview-status" :class="{ active: testResult }">
          <ClockIcon class="preview-icon" />
          <span>{{ testResult ? 'Would Match Now' : 'Not Matching Now' }}</span>
        </div>

        <div class="weekly-preview">
          <div class="preview-day" v-for="day in allDays" :key="day">
            <span class="preview-day-label">{{ day.slice(0,1) }}</span>
            <div class="preview-day-bar">
              <div 
                v-for="(block, bi) in getDayBlocks(day)" 
                :key="bi" 
                class="preview-block"
                :style="block"
              ></div>
            </div>
          </div>
        </div>

        <div class="preview-legend">
          <div class="legend-item"><span class="legend-color match"></span> Active Period</div>
        </div>
      </div>

      <div class="help-card">
        <h4>Tips</h4>
        <ul>
          <li>Multiple rules are combined with OR logic</li>
          <li>Lower priority conditions are checked first</li>
          <li>Holiday lists override normal schedules</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Clock as ClockIcon, CheckCircle as CheckCircleIcon, XCircle as XCircleIcon,
  Play as PlayIcon, Trash2 as TrashIcon
} from 'lucide-vue-next'
import { timeConditionsAPI, holidaysAPI } from '../../services/api'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id || route.params.id === 'new')
const loading = ref(false)
const error = ref(null)

const allDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
const dayToNumber = { 'Sun': 0, 'Mon': 1, 'Tue': 2, 'Wed': 3, 'Thu': 4, 'Fri': 5, 'Sat': 6 }
const numberToDay = { 0: 'Sun', 1: 'Mon', 2: 'Tue', 3: 'Wed', 4: 'Thu', 5: 'Fri', 6: 'Sat' }

const form = ref({
  name: '',
  priority: 10,
  rules: [
    { days: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'], startTime: '09:00', endTime: '17:00' }
  ],
  holiday_list_id: '',
  holidayDestType: 'voicemail',
  holidayDestValue: '',
  matchType: 'ivr',
  matchTarget: '',
  noMatchType: 'voicemail',
  noMatchTarget: '',
  timezone: 'America/New_York',
  extension: '',
  enabled: true
})

const holidayLists = ref([])
const testResult = ref(true)

onMounted(async () => {
  await loadHolidayLists()
  if (!isNew.value && route.params.id) {
    await loadCondition(route.params.id)
  }
})

const loadHolidayLists = async () => {
  try {
    const response = await holidaysAPI.list()
    holidayLists.value = (response.data?.data || []).map(h => ({
      id: h.id,
      name: h.name,
      count: (h.dates || []).length
    }))
  } catch (e) {
    console.error('Failed to load holiday lists', e)
  }
}

const loadCondition = async (id) => {
  loading.value = true
  try {
    const response = await timeConditionsAPI.get(id)
    const tc = response.data?.data || response.data
    
    // Map backend model to form structure
    // Backend: weekdays (int array), start_time, end_time strings
    const weekdays = (tc.weekdays || []).map(d => numberToDay[d])
    
    form.value = {
      name: tc.name || '',
      priority: 10,
      rules: [{
        days: weekdays,
        startTime: tc.start_time || '09:00',
        endTime: tc.end_time || '17:00'
      }],
      holiday_list_id: tc.holiday_list_id || '',
      holidayDestType: tc.holiday_dest_type || 'voicemail',
      holidayDestValue: tc.holiday_dest_value || '',
      matchType: tc.match_dest_type || 'ivr',
      matchTarget: tc.match_dest_value || '',
      noMatchType: tc.nomatch_dest_type || 'voicemail',
      noMatchTarget: tc.nomatch_dest_value || '',
      timezone: tc.timezone || 'America/New_York',
      extension: tc.extension || '',
      enabled: tc.enabled !== false
    }
  } catch (e) {
    error.value = 'Failed to load time condition'
    console.error(e)
  } finally {
    loading.value = false
  }
}

const addRule = () => {
  form.value.rules.push({ days: [], startTime: '09:00', endTime: '17:00' })
}

const removeRule = (i) => {
  form.value.rules.splice(i, 1)
}

const getTimeBlockStyle = (rule) => {
  const start = parseInt(rule.startTime.split(':')[0]) + parseInt(rule.startTime.split(':')[1]) / 60
  const end = parseInt(rule.endTime.split(':')[0]) + parseInt(rule.endTime.split(':')[1]) / 60
  const left = (start / 24) * 100
  const width = ((end - start) / 24) * 100
  return { left: `${left}%`, width: `${Math.max(width, 2)}%` }
}

const getDayBlocks = (day) => {
  return form.value.rules
    .filter(r => r.days.includes(day))
    .map(r => getTimeBlockStyle(r))
}

const applyTemplate = (type) => {
  switch (type) {
    case 'business':
      form.value.rules = [{ days: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'], startTime: '09:00', endTime: '17:00' }]
      break
    case 'extended':
      form.value.rules = [{ days: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'], startTime: '08:00', endTime: '20:00' }]
      break
    case '24x7':
      form.value.rules = [{ days: allDays, startTime: '00:00', endTime: '23:59' }]
      break
    case 'weekend':
      form.value.rules = [{ days: ['Sat', 'Sun'], startTime: '09:00', endTime: '17:00' }]
      break
  }
}

const getPlaceholder = (type) => {
  const placeholders = {
    ivr: 'Select IVR menu...',
    extension: 'e.g. 101',
    ring_group: 'Select ring group...',
    queue: 'Select queue...',
    voicemail: 'Select voicemail box...',
    external: '+1 555-123-4567'
  }
  return placeholders[type] || ''
}

const testCondition = () => {
  const now = new Date()
  const dayName = allDays[now.getDay() === 0 ? 6 : now.getDay() - 1]
  const currentTime = `${now.getHours().toString().padStart(2,'0')}:${now.getMinutes().toString().padStart(2,'0')}`
  
  testResult.value = form.value.rules.some(rule => {
    return rule.days.includes(dayName) && 
           currentTime >= rule.startTime && 
           currentTime <= rule.endTime
  })
}

const saveCondition = async () => {
  error.value = null
  loading.value = true
  
  try {
    // Collect all weekdays from all rules (flatten)
    const allWeekdays = new Set()
    form.value.rules.forEach(rule => {
      rule.days.forEach(d => allWeekdays.add(dayToNumber[d]))
    })
    
    // Use first rule's times (simplified - full implementation would store multiple rules)
    const firstRule = form.value.rules[0] || { startTime: '09:00', endTime: '17:00' }
    
    const payload = {
      name: form.value.name,
      extension: form.value.extension,
      timezone: form.value.timezone,
      weekdays: Array.from(allWeekdays),
      start_time: firstRule.startTime,
      end_time: firstRule.endTime,
      holiday_list_id: form.value.holiday_list_id ? parseInt(form.value.holiday_list_id) : null,
      holiday_dest_type: form.value.holiday_list_id ? form.value.holidayDestType : '',
      holiday_dest_value: form.value.holiday_list_id ? form.value.holidayDestValue : '',
      match_dest_type: form.value.matchType,
      match_dest_value: form.value.matchTarget,
      nomatch_dest_type: form.value.noMatchType,
      nomatch_dest_value: form.value.noMatchTarget,
      enabled: form.value.enabled
    }
    
    if (isNew.value) {
      await timeConditionsAPI.create(payload)
    } else {
      await timeConditionsAPI.update(route.params.id, payload)
    }
    
    router.push('/admin/time-conditions')
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to save time condition'
    console.error('Save error:', e)
  } finally {
    loading.value = false
  }
}
</script>


<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: var(--spacing-lg); }
.header-left { display: flex; flex-direction: column; gap: 8px; }
.header-actions { display: flex; gap: 8px; }
.back-link { background: none; border: none; color: var(--text-muted); cursor: pointer; padding: 0; font-size: 12px; text-align: left; }
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.form-layout { display: grid; grid-template-columns: 1fr 300px; gap: 24px; }
.form-main { display: flex; flex-direction: column; gap: 20px; }
.form-sidebar { display: flex; flex-direction: column; gap: 16px; }

.form-section { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 20px; }
.form-section h3 { font-size: 14px; font-weight: 600; margin: 0 0 12px 0; }
.section-header { display: flex; justify-content: space-between; align-items: center; }

.form-row { display: flex; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.flex-1 { flex: 1; }
.flex-2 { flex: 2; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); }
.text-muted { color: var(--text-muted); }
.text-sm { font-size: 12px; }
.mb-md { margin-bottom: 12px; }

/* Schedule Editor */
.schedule-editor { display: flex; flex-direction: column; gap: 16px; }
.rule-card { background: var(--bg-app); border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 16px; }
.rule-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.rule-number { font-size: 12px; font-weight: 600; color: var(--text-muted); }

.days-picker { display: flex; gap: 6px; margin-bottom: 12px; }
.day-btn { display: flex; }
.day-btn input { display: none; }
.day-btn span { padding: 8px 12px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 12px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
.day-btn input:checked + span { background: var(--primary-color); color: white; border-color: var(--primary-color); }

.time-range { display: flex; align-items: flex-end; gap: 12px; margin-bottom: 12px; }
.time-field { display: flex; flex-direction: column; gap: 6px; }
.time-arrow { color: var(--text-muted); font-weight: bold; margin-bottom: 10px; }

.rule-preview { background: #e2e8f0; border-radius: 4px; overflow: hidden; }
.hour-labels { display: flex; justify-content: space-between; padding: 4px 8px; font-size: 9px; color: var(--text-muted); }
.hour-bar { height: 12px; background: #cbd5e1; position: relative; }
.hour-fill { position: absolute; top: 2px; bottom: 2px; background: linear-gradient(90deg, #6366f1, #8b5cf6); border-radius: 2px; }

.quick-templates { display: flex; align-items: center; gap: 8px; margin-top: 16px; padding-top: 16px; border-top: 1px dashed var(--border-color); }
.template-label { font-size: 12px; color: var(--text-muted); font-weight: 500; }
.template-btn { padding: 4px 10px; background: white; border: 1px solid var(--border-color); border-radius: 99px; font-size: 11px; cursor: pointer; transition: all 0.2s; }
.template-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }

/* Holiday Checkboxes */
.holiday-checkboxes { display: flex; flex-direction: column; gap: 8px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }
.checkbox-row input { width: 16px; height: 16px; }
.holiday-count { margin-left: auto; font-size: 11px; color: var(--text-muted); }

/* Destinations */
.dest-config { background: var(--bg-app); border-radius: var(--radius-sm); padding: 16px; margin-bottom: 12px; }
.dest-config.match { border-left: 3px solid #16a34a; }
.dest-config.nomatch { border-left: 3px solid #dc2626; }
.dest-indicator { display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 600; margin-bottom: 12px; }
.dest-indicator.match { color: #16a34a; }
.dest-indicator.nomatch { color: #dc2626; }
.dest-form { display: flex; gap: 12px; }

/* Preview Card */
.preview-card, .help-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; }
.preview-card h4, .help-card h4 { font-size: 12px; font-weight: 600; margin: 0 0 12px 0; text-transform: uppercase; color: var(--text-muted); }

.preview-status { display: flex; align-items: center; gap: 8px; padding: 10px; background: #fee2e2; color: #dc2626; border-radius: var(--radius-sm); font-size: 13px; font-weight: 500; margin-bottom: 16px; }
.preview-status.active { background: #dcfce7; color: #16a34a; }
.preview-icon { width: 16px; height: 16px; }

.weekly-preview { display: flex; flex-direction: column; gap: 4px; }
.preview-day { display: flex; align-items: center; gap: 8px; }
.preview-day-label { width: 16px; font-size: 10px; font-weight: 600; color: var(--text-muted); text-align: center; }
.preview-day-bar { flex: 1; height: 12px; background: #e2e8f0; border-radius: 2px; position: relative; }
.preview-block { position: absolute; top: 2px; bottom: 2px; background: linear-gradient(90deg, #6366f1, #8b5cf6); border-radius: 2px; }

.preview-legend { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--border-color); }
.legend-item { display: flex; align-items: center; gap: 6px; font-size: 11px; color: var(--text-muted); }
.legend-color { width: 12px; height: 8px; border-radius: 2px; }
.legend-color.match { background: linear-gradient(90deg, #6366f1, #8b5cf6); }

.help-card ul { margin: 0; padding-left: 16px; font-size: 12px; color: var(--text-muted); }
.help-card li { margin-bottom: 6px; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary.small { padding: 6px 10px; font-size: 12px; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; display: flex; }
.btn-icon:hover { color: var(--text-primary); }
.btn-icon-left { width: 14px; height: 14px; }
.icon-sm { width: 16px; height: 16px; }
.text-bad { color: var(--status-bad); }
</style>
