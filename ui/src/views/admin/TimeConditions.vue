<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Time Conditions</h2>
      <p class="text-muted text-sm">Manage time-based routing schedules, business hours, and holiday lists.</p>
    </div>
    <div class="header-actions">
      <button class="btn-secondary" @click="showHolidayModal = true">
        <CalendarIcon class="btn-icon-left" />
        Holidays
      </button>
      <button class="btn-primary" @click="$router.push('/admin/time-conditions/new')">+ New Time Condition</button>
    </div>
  </div>

  <!-- Stats Row -->
  <div class="stats-row">
    <div class="stat-card">
      <div class="stat-icon active"><CheckCircleIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ activeCount }}</span>
        <span class="stat-label">Active Now</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon inactive"><ClockIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ timeConditions.filter(t => !t.currentMatch).length }}</span>
        <span class="stat-label">Inactive</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon total"><ListIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ timeConditions.length }}</span>
        <span class="stat-label">Total Conditions</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon holidays"><CalendarIcon class="icon" /></div>
      <div class="stat-info">
        <span class="stat-value">{{ holidayLists.length }}</span>
        <span class="stat-label">Holiday Lists</span>
      </div>
    </div>
  </div>

  <!-- Tabs -->
  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'conditions' }" @click="activeTab = 'conditions'">Time Conditions</button>
    <button class="tab" :class="{ active: activeTab === 'holidays' }" @click="activeTab = 'holidays'">Holiday Lists</button>
  </div>

  <!-- TIME CONDITIONS TAB -->
  <div class="tab-content" v-if="activeTab === 'conditions'">
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input type="text" v-model="searchQuery" placeholder="Search time conditions..." class="search-input">
      </div>
      <select v-model="filterStatus" class="filter-select">
        <option value="">All Statuses</option>
        <option value="active">Currently Active</option>
        <option value="inactive">Not Active</option>
      </select>
    </div>

    <div class="conditions-list">
      <div class="condition-card" v-for="tc in filteredConditions" :key="tc.id" :class="{ active: tc.currentMatch }">
        <div class="condition-header">
          <div class="condition-icon" :class="{ matching: tc.currentMatch }">
            <ClockIcon class="icon-md" />
            <div v-if="tc.currentMatch" class="active-pulse"></div>
          </div>
          <div class="condition-info">
            <h4>{{ tc.name }}</h4>
            <div class="condition-meta">
              <span class="ext-badge">Ext. {{ tc.extension }}</span>
              <span class="condition-status" :class="{ matching: tc.currentMatch }">
                {{ tc.currentMatch ? '● Currently Active' : '○ Not Active' }}
              </span>
            </div>
          </div>
          <label class="switch">
            <input type="checkbox" v-model="tc.enabled">
            <span class="slider round"></span>
          </label>
        </div>

        <div class="condition-schedule">
          <div class="schedule-visual">
            <div class="day-column" v-for="day in ['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" :key="day">
              <span class="day-label">{{ day }}</span>
              <div class="day-bar">
                <div 
                  v-for="(rule, ri) in getRulesForDay(tc, day)" 
                  :key="ri" 
                  class="time-block"
                  :style="getTimeBlockStyle(rule)"
                  :title="`${rule.startTime} - ${rule.endTime}`"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="condition-destinations">
          <div class="dest-row match">
            <CheckCircleIcon class="dest-icon" />
            <span class="dest-label">Match →</span>
            <span class="dest-target">{{ tc.matchDestination }}</span>
          </div>
          <div class="dest-row nomatch">
            <XCircleIcon class="dest-icon" />
            <span class="dest-label">No Match →</span>
            <span class="dest-target">{{ tc.noMatchDestination }}</span>
          </div>
        </div>

        <div class="condition-actions">
          <button class="btn-link" @click="$router.push(`/admin/time-conditions/${tc.id}`)">Edit</button>
          <button class="btn-link" @click="duplicateCondition(tc)">Duplicate</button>
          <button class="btn-link text-bad" @click="deleteCondition(tc)">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <!-- HOLIDAYS TAB -->
  <div class="tab-content" v-else-if="activeTab === 'holidays'">
    <div class="holidays-header">
      <p class="text-muted">Define holiday dates that override normal time conditions.</p>
      <button class="btn-primary" @click="showHolidayModal = true">+ New Holiday List</button>
    </div>

    <div class="holidays-grid">
      <div class="holiday-card" v-for="list in holidayLists" :key="list.id">
        <div class="holiday-header">
          <div class="holiday-icon"><CalendarIcon class="icon-md" /></div>
          <div class="holiday-info">
            <h4>{{ list.name }}</h4>
            <span class="holiday-meta">{{ list.count }} dates</span>
          </div>
          <span class="source-badge">{{ list.source }}</span>
        </div>
        
        <div class="holiday-preview">
          <div class="preview-date" v-for="date in list.upcoming" :key="date.date">
            <span class="date-label">{{ date.name }}</span>
            <span class="date-value">{{ date.date }}</span>
          </div>
        </div>
        
        <div class="holiday-actions">
          <button class="btn-link" @click="editHolidayList(list)">Edit</button>
          <button class="btn-link" @click="syncHolidayList(list)" v-if="list.source === 'External URL'">
            <RefreshCwIcon class="btn-icon-sm" /> Sync
          </button>
          <button class="btn-link text-bad" @click="deleteHolidayList(list)">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Holiday Modal -->
  <div v-if="showHolidayModal" class="modal-overlay" @click.self="showHolidayModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ editingHoliday ? 'Edit Holiday List' : 'New Holiday List' }}</h3>
        <button class="btn-icon" @click="showHolidayModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>List Name</label>
          <input v-model="holidayForm.name" class="input-field" placeholder="e.g. US Federal Holidays 2025">
        </div>

        <div class="form-group">
          <label>Source</label>
          <div class="source-toggle">
            <button class="toggle-btn" :class="{ active: holidayForm.source === 'manual' }" @click="holidayForm.source = 'manual'">Manual Entry</button>
            <button class="toggle-btn" :class="{ active: holidayForm.source === 'url' }" @click="holidayForm.source = 'url'">External URL</button>
          </div>
        </div>

        <div v-if="holidayForm.source === 'url'" class="form-group">
          <label>ICS/URL</label>
          <input v-model="holidayForm.url" class="input-field" placeholder="https://example.com/holidays.ics">
          <span class="help-text">Supports iCal (.ics) format</span>
        </div>

        <div v-else class="dates-editor">
          <div class="date-row" v-for="(date, i) in holidayForm.dates" :key="i">
            <input type="date" v-model="date.date" class="input-field">
            <input type="text" v-model="date.name" class="input-field flex-1" placeholder="Holiday name">
            <button class="btn-icon" @click="removeHolidayDate(i)"><XIcon class="icon-sm" /></button>
          </div>
          <button class="btn-secondary small" @click="addHolidayDate">+ Add Date</button>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showHolidayModal = false">Cancel</button>
        <button class="btn-primary" @click="saveHolidayList" :disabled="!holidayForm.name">Save List</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Clock as ClockIcon, CheckCircle as CheckCircleIcon, XCircle as XCircleIcon,
  Calendar as CalendarIcon, List as ListIcon, Search as SearchIcon,
  RefreshCw as RefreshCwIcon, X as XIcon
} from 'lucide-vue-next'

const activeTab = ref('conditions')
const searchQuery = ref('')
const filterStatus = ref('')

// Time Conditions Data
const timeConditions = ref([
  {
    id: 1, name: 'Business Hours', extension: '40', enabled: true, currentMatch: true,
    rules: [
      { days: ['Mon','Tue','Wed','Thu','Fri'], startTime: '09:00', endTime: '17:00' }
    ],
    matchDestination: 'IVR: Main Menu (8000)',
    noMatchDestination: 'IVR: After Hours (8001)'
  },
  {
    id: 2, name: 'Weekend Support', extension: '41', enabled: true, currentMatch: false,
    rules: [
      { days: ['Sat'], startTime: '10:00', endTime: '14:00' }
    ],
    matchDestination: 'Ring Group: Weekend On-Call',
    noMatchDestination: 'Voicemail: General'
  },
  {
    id: 3, name: 'Lunch Break Override', extension: '42', enabled: true, currentMatch: false,
    rules: [
      { days: ['Mon','Tue','Wed','Thu','Fri'], startTime: '12:00', endTime: '13:00' }
    ],
    matchDestination: 'IVR: Lunch Menu',
    noMatchDestination: 'Continue to Next'
  }
])

const activeCount = computed(() => timeConditions.value.filter(t => t.currentMatch && t.enabled).length)

const filteredConditions = computed(() => {
  return timeConditions.value.filter(tc => {
    const matchesSearch = tc.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesStatus = !filterStatus.value || 
      (filterStatus.value === 'active' && tc.currentMatch) ||
      (filterStatus.value === 'inactive' && !tc.currentMatch)
    return matchesSearch && matchesStatus
  })
})

const getRulesForDay = (tc, day) => {
  return tc.rules.filter(r => r.days.includes(day))
}

const getTimeBlockStyle = (rule) => {
  const start = parseInt(rule.startTime.split(':')[0])
  const end = parseInt(rule.endTime.split(':')[0])
  const left = (start / 24) * 100
  const width = ((end - start) / 24) * 100
  return { left: `${left}%`, width: `${width}%` }
}

const duplicateCondition = (tc) => {
  const copy = { ...tc, id: Date.now(), name: `${tc.name} (Copy)` }
  timeConditions.value.push(copy)
}

const deleteCondition = (tc) => {
  if (confirm(`Delete "${tc.name}"?`)) {
    timeConditions.value = timeConditions.value.filter(t => t.id !== tc.id)
  }
}

// Holiday Lists
const holidayLists = ref([
  { 
    id: 1, name: 'US Federal 2025', count: 11, source: 'External URL',
    upcoming: [
      { name: 'New Year', date: 'Jan 1' },
      { name: 'MLK Day', date: 'Jan 20' },
      { name: 'Presidents Day', date: 'Feb 17' }
    ]
  },
  { 
    id: 2, name: 'Office Closures', count: 3, source: 'Manual',
    upcoming: [
      { name: 'Company Holiday', date: 'Dec 26' },
      { name: 'Team Building', date: 'Mar 15' }
    ]
  }
])

const showHolidayModal = ref(false)
const editingHoliday = ref(false)
const holidayForm = ref({
  name: '',
  source: 'manual',
  url: '',
  dates: [{ date: '', name: '' }]
})

const addHolidayDate = () => {
  holidayForm.value.dates.push({ date: '', name: '' })
}

const removeHolidayDate = (i) => {
  holidayForm.value.dates.splice(i, 1)
}

const editHolidayList = (list) => {
  editingHoliday.value = true
  showHolidayModal.value = true
}

const syncHolidayList = (list) => {
  alert(`Syncing ${list.name} from external URL...`)
}

const saveHolidayList = () => {
  showHolidayModal.value = false
  editingHoliday.value = false
}

const deleteHolidayList = (list) => {
  if (confirm(`Delete "${list.name}"?`)) {
    holidayLists.value = holidayLists.value.filter(l => l.id !== list.id)
  }
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

/* Stats Row */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.active { background: #dcfce7; color: #16a34a; }
.stat-icon.inactive { background: #fef3c7; color: #b45309; }
.stat-icon.total { background: #dbeafe; color: #2563eb; }
.stat-icon.holidays { background: #fce7f3; color: #db2777; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); }

/* Filter Bar */
.filter-bar { display: flex; gap: 12px; margin-bottom: var(--spacing-lg); }
.search-box { position: relative; flex: 1; max-width: 300px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 12px 8px 36px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--text-sm); background: white; }

/* Conditions List */
.conditions-list { display: flex; flex-direction: column; gap: 16px; }
.condition-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; transition: all 0.2s; }
.condition-card.active { border-color: #22c55e; background: linear-gradient(to right, #f0fdf4, white); }

.condition-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.condition-icon { position: relative; width: 44px; height: 44px; background: #fef3c7; border-radius: 10px; display: flex; align-items: center; justify-content: center; color: #b45309; }
.condition-icon.matching { background: #dcfce7; color: #16a34a; }
.active-pulse { position: absolute; inset: -3px; border: 2px solid #22c55e; border-radius: 12px; animation: pulse 2s infinite; }
@keyframes pulse { 0%, 100% { opacity: 0.3; transform: scale(1); } 50% { opacity: 0.6; transform: scale(1.05); } }

.condition-info { flex: 1; }
.condition-info h4 { font-size: 15px; font-weight: 600; margin: 0; }
.condition-meta { display: flex; align-items: center; gap: 8px; margin-top: 2px; }
.ext-badge { font-size: 10px; font-family: monospace; background: #f0fdf4; color: #16a34a; padding: 2px 6px; border-radius: 4px; font-weight: 600; }
.condition-status { font-size: 11px; color: var(--text-muted); }
.condition-status.matching { color: #16a34a; font-weight: 600; }

/* Visual Schedule */
.schedule-visual { display: flex; gap: 4px; margin-bottom: 16px; padding: 12px; background: var(--bg-app); border-radius: var(--radius-sm); }
.day-column { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.day-label { font-size: 10px; font-weight: 600; text-align: center; color: var(--text-muted); }
.day-bar { height: 20px; background: #e2e8f0; border-radius: 2px; position: relative; overflow: hidden; }
.time-block { position: absolute; top: 2px; bottom: 2px; background: linear-gradient(90deg, #6366f1, #8b5cf6); border-radius: 2px; min-width: 4px; }

/* Destinations */
.condition-destinations { display: flex; gap: 16px; margin-bottom: 12px; }
.dest-row { display: flex; align-items: center; gap: 6px; font-size: 13px; padding: 6px 12px; background: var(--bg-app); border-radius: var(--radius-sm); flex: 1; }
.dest-icon { width: 14px; height: 14px; }
.dest-row.match .dest-icon { color: #16a34a; }
.dest-row.nomatch .dest-icon { color: #dc2626; }
.dest-label { font-weight: 500; color: var(--text-muted); font-size: 11px; }
.dest-target { color: var(--text-main); font-weight: 500; }

.condition-actions { display: flex; gap: 8px; justify-content: flex-end; padding-top: 12px; border-top: 1px solid var(--border-color); }

/* Holidays */
.holidays-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.holidays-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.holiday-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; }

.holiday-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.holiday-icon { width: 40px; height: 40px; background: #fce7f3; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #db2777; }
.holiday-info { flex: 1; }
.holiday-info h4 { font-size: 14px; font-weight: 600; margin: 0; }
.holiday-meta { font-size: 11px; color: var(--text-muted); }
.source-badge { font-size: 10px; background: #f3f4f6; color: #6b7280; padding: 2px 8px; border-radius: 99px; font-weight: 600; }

.holiday-preview { display: flex; flex-direction: column; gap: 4px; margin-bottom: 12px; }
.preview-date { display: flex; justify-content: space-between; padding: 4px 8px; background: var(--bg-app); border-radius: 4px; font-size: 12px; }
.date-label { font-weight: 500; }
.date-value { color: var(--text-muted); }

.holiday-actions { display: flex; gap: 8px; justify-content: flex-end; padding-top: 12px; border-top: 1px solid var(--border-color); }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary.small { padding: 6px 10px; font-size: 12px; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: var(--text-xs); cursor: pointer; font-weight: 500; display: flex; align-items: center; gap: 4px; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; display: flex; }
.btn-icon:hover { color: var(--text-primary); }
.btn-icon-left { width: 14px; height: 14px; }
.btn-icon-sm { width: 12px; height: 12px; }
.icon-sm { width: 16px; height: 16px; }
.icon-md { width: 20px; height: 20px; }
.text-bad { color: var(--status-bad); }

/* Switch */
.switch { position: relative; display: inline-block; width: 36px; height: 20px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: var(--border-color); transition: .3s; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 3px; bottom: 3px; background-color: white; transition: .3s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(16px); }
.slider.round { border-radius: 20px; }
.slider.round:before { border-radius: 50%; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); padding: 24px; }
.modal-card { background: white; border-radius: var(--radius-md); box-shadow: var(--shadow-lg); width: 100%; max-width: 500px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); }
.flex-1 { flex: 1; }

.source-toggle { display: flex; gap: 4px; background: var(--bg-app); padding: 4px; border-radius: var(--radius-sm); }
.toggle-btn { flex: 1; padding: 8px; border: none; background: transparent; border-radius: 4px; font-size: 13px; font-weight: 500; cursor: pointer; color: var(--text-muted); }
.toggle-btn.active { background: white; color: var(--primary-color); box-shadow: var(--shadow-sm); }

.dates-editor { display: flex; flex-direction: column; gap: 8px; }
.date-row { display: flex; gap: 8px; align-items: center; }
</style>
