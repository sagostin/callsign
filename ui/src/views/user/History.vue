<template>
  <div class="history-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Call History</h2>
        <p class="text-muted text-sm">View your recent inbound and outbound calls.</p>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon total"><PhoneIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ callHistory.length }}</span>
          <span class="stat-label">Total Calls</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon inbound"><PhoneIncomingIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ callHistory.filter(c => c.type === 'inbound').length }}</span>
          <span class="stat-label">Inbound</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon outbound"><PhoneOutgoingIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ callHistory.filter(c => c.type === 'outbound').length }}</span>
          <span class="stat-label">Outbound</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon missed"><PhoneMissedIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ callHistory.filter(c => c.type === 'missed').length }}</span>
          <span class="stat-label">Missed</span>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" class="search-input" placeholder="Search by name or number...">
      </div>
      <div class="filter-group">
        <button class="filter-btn" :class="{ active: activeFilter === 'all' }" @click="activeFilter = 'all'">All</button>
        <button class="filter-btn" :class="{ active: activeFilter === 'inbound' }" @click="activeFilter = 'inbound'">Inbound</button>
        <button class="filter-btn" :class="{ active: activeFilter === 'outbound' }" @click="activeFilter = 'outbound'">Outbound</button>
        <button class="filter-btn" :class="{ active: activeFilter === 'missed' }" @click="activeFilter = 'missed'">Missed</button>
      </div>
      <select v-model="dateFilter" class="filter-select">
        <option value="">All Time</option>
        <option value="today">Today</option>
        <option value="week">This Week</option>
        <option value="month">This Month</option>
      </select>
    </div>

    <!-- Call List -->
    <div class="call-list">
      <!-- Date Groups -->
      <div class="date-group" v-for="group in groupedCalls" :key="group.date">
        <div class="date-header">{{ group.label }}</div>
        
        <div class="call-item" v-for="call in group.calls" :key="call.id" @click="selectCall(call)">
          <div class="call-type-icon" :class="call.type">
            <PhoneIncomingIcon v-if="call.type === 'inbound'" class="icon-sm" />
            <PhoneOutgoingIcon v-else-if="call.type === 'outbound'" class="icon-sm" />
            <PhoneMissedIcon v-else class="icon-sm" />
          </div>
          
          <div class="call-main">
            <div class="call-party">
              <span class="party-name" v-if="call.name">{{ call.name }}</span>
              <span class="party-number" :class="{ primary: !call.name }">{{ call.number }}</span>
            </div>
            <div class="call-time">{{ call.time }}</div>
          </div>

          <div class="call-duration">
            <span v-if="call.duration">{{ call.duration }}</span>
            <span v-else class="text-muted">—</span>
          </div>

          <div class="call-actions">
            <button class="action-btn" @click.stop="makeCall(call.number)" title="Call">
              <PhoneIcon class="icon-sm" />
            </button>
            <button class="action-btn" @click.stop="sendMessage(call.number)" title="Message">
              <MessageSquareIcon class="icon-sm" />
            </button>
            <button class="action-btn" @click.stop="addToContacts(call)" title="Add Contact" v-if="!call.name">
              <UserPlusIcon class="icon-sm" />
            </button>
          </div>
        </div>
      </div>

      <div class="empty-state" v-if="filteredCalls.length === 0">
        <PhoneOffIcon class="empty-icon" />
        <p>No calls found</p>
        <span class="text-muted text-sm">Try adjusting your filters</span>
      </div>
    </div>

    <!-- Call Detail Panel -->
    <div class="detail-panel" v-if="selectedCall" @click.self="selectedCall = null">
      <div class="detail-card">
        <button class="close-btn" @click="selectedCall = null"><XIcon class="icon-sm" /></button>
        
        <div class="detail-header">
          <div class="detail-avatar" :class="selectedCall.type">
            {{ selectedCall.name ? selectedCall.name.charAt(0).toUpperCase() : '#' }}
          </div>
          <div class="detail-info">
            <span class="detail-name">{{ selectedCall.name || 'Unknown Caller' }}</span>
            <span class="detail-number">{{ selectedCall.number }}</span>
          </div>
        </div>

        <div class="detail-row">
          <span class="detail-label">Type</span>
          <span class="type-badge" :class="selectedCall.type">{{ selectedCall.type }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Date & Time</span>
          <span>{{ selectedCall.fullDate }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">Duration</span>
          <span>{{ selectedCall.duration || 'N/A' }}</span>
        </div>
        <div class="detail-row" v-if="selectedCall.recording">
          <span class="detail-label">Recording</span>
          <button class="btn-link"><PlayIcon class="icon-xs" /> Play Recording</button>
        </div>

        <div class="detail-actions">
          <button class="btn-primary" @click="makeCall(selectedCall.number)">
            <PhoneIcon class="btn-icon" /> Call Back
          </button>
          <button class="btn-secondary" @click="sendMessage(selectedCall.number)">
            <MessageSquareIcon class="btn-icon" /> Message
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Phone as PhoneIcon, PhoneIncoming as PhoneIncomingIcon, 
  PhoneOutgoing as PhoneOutgoingIcon, PhoneMissed as PhoneMissedIcon,
  PhoneOff as PhoneOffIcon, Search as SearchIcon,
  MessageSquare as MessageSquareIcon, UserPlus as UserPlusIcon,
  X as XIcon, Play as PlayIcon
} from 'lucide-vue-next'

const searchQuery = ref('')
const activeFilter = ref('all')
const dateFilter = ref('')
const selectedCall = ref(null)

const callHistory = ref([
  { id: 1, type: 'missed', name: null, number: '(415) 555-9999', time: '10:45 AM', duration: null, date: 'today', fullDate: 'Dec 9, 2024 10:45 AM', recording: false },
  { id: 2, type: 'inbound', name: 'Alice Smith', number: '(415) 555-1234', time: '9:30 AM', duration: '5m 12s', date: 'today', fullDate: 'Dec 9, 2024 9:30 AM', recording: true },
  { id: 3, type: 'outbound', name: 'Bob Jones', number: 'Ext 102', time: '9:00 AM', duration: '2m 45s', date: 'today', fullDate: 'Dec 9, 2024 9:00 AM', recording: false },
  { id: 4, type: 'inbound', name: null, number: '(212) 555-8888', time: '4:15 PM', duration: '8m 33s', date: 'yesterday', fullDate: 'Dec 8, 2024 4:15 PM', recording: true },
  { id: 5, type: 'outbound', name: 'Support Team', number: '(800) 555-0100', time: '2:00 PM', duration: '15m 22s', date: 'yesterday', fullDate: 'Dec 8, 2024 2:00 PM', recording: false },
  { id: 6, type: 'missed', name: 'Jane Doe', number: '(310) 555-4567', time: '11:30 AM', duration: null, date: 'yesterday', fullDate: 'Dec 8, 2024 11:30 AM', recording: false },
  { id: 7, type: 'inbound', name: 'Client ABC', number: '(555) 123-4567', time: '3:45 PM', duration: '12m 10s', date: 'week', fullDate: 'Dec 5, 2024 3:45 PM', recording: true },
  { id: 8, type: 'outbound', name: null, number: '(555) 987-6543', time: '10:00 AM', duration: '1m 30s', date: 'week', fullDate: 'Dec 4, 2024 10:00 AM', recording: false },
])

const filteredCalls = computed(() => {
  return callHistory.value.filter(call => {
    const matchesSearch = !searchQuery.value || 
      call.number.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (call.name && call.name.toLowerCase().includes(searchQuery.value.toLowerCase()))
    
    const matchesFilter = activeFilter.value === 'all' || call.type === activeFilter.value
    
    const matchesDate = !dateFilter.value || call.date === dateFilter.value ||
      (dateFilter.value === 'week' && ['today', 'yesterday', 'week'].includes(call.date)) ||
      (dateFilter.value === 'month' && true)
    
    return matchesSearch && matchesFilter && matchesDate
  })
})

const groupedCalls = computed(() => {
  const groups = {}
  filteredCalls.value.forEach(call => {
    if (!groups[call.date]) {
      groups[call.date] = {
        date: call.date,
        label: call.date === 'today' ? 'Today' : call.date === 'yesterday' ? 'Yesterday' : 'Earlier This Week',
        calls: []
      }
    }
    groups[call.date].calls.push(call)
  })
  return Object.values(groups)
})

const selectCall = (call) => { selectedCall.value = call }
const makeCall = (number) => alert(`Calling ${number}...`)
const sendMessage = (number) => alert(`Messaging ${number}...`)
const addToContacts = (call) => alert(`Add ${call.number} to contacts`)
</script>

<style scoped>
.history-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.total { background: #f3e8ff; color: #7c3aed; }
.stat-icon.inbound { background: #dcfce7; color: #16a34a; }
.stat-icon.outbound { background: #dbeafe; color: #2563eb; }
.stat-icon.missed { background: #fee2e2; color: #dc2626; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Filters */
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.search-box { position: relative; flex: 1; min-width: 200px; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 10px 12px 10px 38px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }

.filter-group { display: flex; gap: 2px; background: var(--bg-app); padding: 3px; border-radius: 8px; }
.filter-btn { padding: 8px 14px; border: none; background: transparent; border-radius: 6px; font-size: 12px; font-weight: 500; cursor: pointer; color: var(--text-muted); }
.filter-btn.active { background: white; color: var(--text-primary); box-shadow: 0 1px 2px rgba(0,0,0,0.05); }

.filter-select { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }

/* Call List */
.call-list { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }

.date-header { padding: 10px 20px; background: var(--bg-app); font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); }

.call-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background 0.15s;
}
.call-item:last-child { border-bottom: none; }
.call-item:hover { background: var(--bg-app); }

.call-type-icon { width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
.call-type-icon.inbound { background: #dcfce7; color: #16a34a; }
.call-type-icon.outbound { background: #dbeafe; color: #2563eb; }
.call-type-icon.missed { background: #fee2e2; color: #dc2626; }

.call-main { flex: 1; }
.call-party { display: flex; flex-direction: column; }
.party-name { font-weight: 600; font-size: 14px; }
.party-number { font-size: 12px; color: var(--text-muted); font-family: monospace; }
.party-number.primary { font-size: 14px; font-weight: 500; color: var(--text-primary); }
.call-time { font-size: 11px; color: var(--text-muted); margin-top: 2px; }

.call-duration { font-size: 13px; color: var(--text-muted); min-width: 70px; text-align: right; }

.call-actions { display: flex; gap: 4px; opacity: 0; transition: opacity 0.15s; }
.call-item:hover .call-actions { opacity: 1; }

.action-btn { width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }

.empty-state { text-align: center; padding: 48px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 16px; }

/* Detail Panel */
.detail-panel { position: fixed; inset: 0; z-index: 100; background: rgba(0,0,0,0.3); display: flex; align-items: center; justify-content: flex-end; }
.detail-card { width: 360px; height: 100%; background: white; padding: 24px; box-shadow: -4px 0 20px rgba(0,0,0,0.1); position: relative; }

.close-btn { position: absolute; top: 16px; right: 16px; background: none; border: none; cursor: pointer; color: var(--text-muted); }

.detail-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; padding-bottom: 24px; border-bottom: 1px solid var(--border-color); }
.detail-avatar { width: 56px; height: 56px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 700; color: white; }
.detail-avatar.inbound { background: linear-gradient(135deg, #22c55e, #16a34a); }
.detail-avatar.outbound { background: linear-gradient(135deg, #3b82f6, #2563eb); }
.detail-avatar.missed { background: linear-gradient(135deg, #ef4444, #dc2626); }
.detail-info { display: flex; flex-direction: column; }
.detail-name { font-size: 18px; font-weight: 700; }
.detail-number { font-size: 14px; color: var(--text-muted); font-family: monospace; }

.detail-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--border-color); }
.detail-label { font-size: 12px; color: var(--text-muted); text-transform: uppercase; font-weight: 600; }
.type-badge { font-size: 11px; font-weight: 700; padding: 3px 10px; border-radius: 4px; text-transform: capitalize; }
.type-badge.inbound { background: #dcfce7; color: #16a34a; }
.type-badge.outbound { background: #dbeafe; color: #2563eb; }
.type-badge.missed { background: #fee2e2; color: #dc2626; }

.detail-actions { display: flex; gap: 12px; margin-top: 24px; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; flex: 1; justify-content: center; background-color: var(--primary-color); color: white; border: none; padding: 12px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { display: flex; align-items: center; gap: 6px; flex: 1; justify-content: center; background: white; border: 1px solid var(--border-color); padding: 12px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-link { display: flex; align-items: center; gap: 4px; background: none; border: none; color: var(--primary-color); font-weight: 600; cursor: pointer; font-size: 13px; }
.btn-icon { width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 14px; height: 14px; }
.icon { width: 20px; height: 20px; }

/* Responsive */
@media (max-width: 768px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .filter-bar { flex-direction: column; }
  .search-box { max-width: none; }
  .detail-card { width: 100%; }
  .call-actions { opacity: 1; }
}
</style>
