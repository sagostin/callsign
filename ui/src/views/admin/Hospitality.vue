<template>
  <div class="view-header">
     <div class="header-content">
       <h2>Hospitality & PMS</h2>
       <p class="text-muted text-sm">Hotel management features, Property Management System integration, and room status.</p>
     </div>
     <div class="header-actions">
        <!-- Global Enable Toggle could be here or just rely on Tenant Settings -->
     </div>
  </div>

  <div class="tabs">
    <button class="tab" :class="{ active: activeTab === 'dashboard' }" @click="activeTab = 'dashboard'">Room Status</button>
    <button class="tab" :class="{ active: activeTab === 'pms' }" @click="activeTab = 'pms'">PMS Configuration</button>
    <button class="tab" :class="{ active: activeTab === 'codes' }" @click="activeTab = 'codes'">Service Codes</button>
    <button class="tab" :class="{ active: activeTab === 'wakeup' }" @click="activeTab = 'wakeup'">Wake Up Calls</button>
  </div>

  <!-- ROOM STATUS DASHBOARD -->
  <div class="tab-content" v-if="activeTab === 'dashboard'">
     <div class="flex justify-between items-center mb-6">
         <div class="flex gap-4 items-center">
            <div class="stat-item">
               <span class="label">Total Rooms</span>
               <span class="val">{{ roomStats.total }}</span>
            </div>
            <div class="stat-item">
               <span class="label">Clean</span>
               <span class="val text-green-600">{{ roomStats.clean }}</span>
            </div>
            <div class="stat-item">
               <span class="label">Dirty</span>
               <span class="val text-red-600">{{ roomStats.dirty }}</span>
            </div>
             <div class="stat-item">
               <span class="label">Insp</span>
               <span class="val text-amber-600">{{ roomStats.inspect }}</span>
            </div>
         </div>
         <div class="flex gap-2">
            <select class="input-field">
               <option>All Floors</option>
               <option>Floor 1</option>
               <option>Floor 2</option>
            </select>
            <button class="btn-secondary small" @click="fetchRooms">Refresh</button>
         </div>
     </div>

     <div class="rooms-grid">
        <div v-for="room in rooms" :key="room.ext" class="room-card" :class="room.status">
           <div class="room-header">
              <span class="room-number">{{ room.ext }}</span>
              <span class="room-icon" v-if="room.status === 'clean'">✨</span>
              <span class="room-icon" v-else-if="room.status === 'dirty'">🧹</span>
              <span class="room-icon" v-else>👀</span>
           </div>
           <div class="room-guest">{{ room.guest || 'Vacant' }}</div>
           <div class="room-status-badge">{{ room.status.toUpperCase() }}</div>
           <div class="room-actions">
              <button class="btn-xs" @click="toggleStatus(room)">Set Clean</button>
           </div>
        </div>
     </div>
  </div>

  <!-- PMS CONFIGURATION -->
  <div class="tab-content" v-else-if="activeTab === 'pms'">
     <div class="max-w-2xl">
         <div class="form-group mb-6">
            <label>PMS Protocol</label>
            <select class="input-field">
               <option>Micros Fidelio (FIAS)</option>
               <option>Mitel SX-2000</option>
               <option>Generic XML</option>
            </select>
         </div>
         
         <div class="bg-slate-50 border border-slate-200 rounded p-4 mb-6">
            <div class="grid grid-cols-2 gap-4">
               <div class="form-group">
                   <label>PMS Server IP / Host</label>
                   <input type="text" class="input-field" placeholder="10.0.0.50">
               </div>
               <div class="form-group">
                   <label>PMS Port</label>
                   <input type="number" class="input-field" value="5010">
               </div>
            </div>
         </div>

         <div class="form-group mb-6">
             <label>Room Extension Prefix</label>
             <input type="text" class="input-field" value="1,2,3">
             <span class="help-text">Comma separated prefixes to identify extensions as Guest Rooms (e.g. 101 starts with 1).</span>
         </div>
     </div>
       <div class="flex justify-start border-t border-slate-200 pt-4">
         <button class="btn-primary" @click="savePMSConfig">Save Configuration</button>
       </div>
  </div>

  <!-- SERVICE CODES -->
  <div class="tab-content" v-else-if="activeTab === 'codes'">
      <div class="max-w-2xl">
          <p class="text-xs text-muted mb-lg">Star codes used by housekeeping staff to update room status via phone.</p>
          <div class="grid grid-cols-2 gap-6">
             <div class="form-group">
                <label>Room Clean (Maid)</label>
                <input type="text" class="input-field" value="*77">
             </div>
             <div class="form-group">
                <label>Room Inspection</label>
                <input type="text" class="input-field" value="*78">
             </div>
             <div class="form-group">
                <label>Minibar Charge</label>
                <input type="text" class="input-field" value="*55">
             </div>
             <div class="form-group">
                <label>Do Not Disturb</label>
                <input type="text" class="input-field" value="*40">
             </div>
          </div>
          <div class="flex justify-start border-t border-slate-200 pt-4 mt-6">
            <button class="btn-primary" @click="saveCodes">Save Codes</button>
          </div>
      </div>
  </div>

  <!-- WAKE UP CALLS -->
  <div class="tab-content" v-else-if="activeTab === 'wakeup'">
       <div class="flex justify-between items-center mb-6">
          <p class="text-xs text-muted">Manage scheduled wake up calls for guests.</p>
          <button class="btn-primary small" @click="scheduleWakeUp">+ Schedule Call</button>
       </div>
       
        <div class="rooms-grid">
            <div v-if="wakeupCalls.length === 0" class="text-muted text-sm">No scheduled wake-up calls.</div>
            <div v-for="call in wakeupCalls" :key="call.id" class="room-card clean">
               <div class="room-header">
                  <span class="room-number">{{ call.extension || call.room }}</span>
                  <span class="room-icon">⏰</span>
               </div>
               <div class="room-guest">{{ call.guest_name || call.guest || 'Guest' }}</div>
               <div class="room-status-badge">{{ formatTime(call.time || call.wakeup_time) }}</div>
               <div class="room-actions">
                   <button class="btn-xs text-red-600" @click="cancelWakeup(call)">Cancel</button>
               </div>
            </div>
        </div>
  </div>

</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { hospitalityAPI, wakeupCallsAPI } from '../../services/api'

const toast = inject('toast')
const activeTab = ref('dashboard')

const rooms = ref([])
const wakeupCalls = ref([])

// --- Room Stats (computed from live data) ---
const roomStats = computed(() => {
  const total = rooms.value.length
  const clean = rooms.value.filter(r => r.status === 'clean').length
  const dirty = rooms.value.filter(r => r.status === 'dirty').length
  const inspect = rooms.value.filter(r => r.status === 'inspect').length
  return { total, clean, dirty, inspect }
})

// --- Room Loading ---
const fetchRooms = async () => {
  try {
    const res = await hospitalityAPI.listRooms()
    const data = res.data?.rooms || res.data || []
    rooms.value = (Array.isArray(data) ? data : []).map(r => ({
      ext: r.extension || r.ext || '',
      status: r.status || 'clean',
      guest: r.guest_name || r.guest || 'Vacant',
      id: r.id || r.extension || r.ext
    }))
  } catch (err) {
    toast?.error('Failed to load rooms', err.message)
    rooms.value = []
  }
}

// --- Room Actions ---
const toggleStatus = async (room) => {
  if (!room?.id) return
  try {
    await hospitalityAPI.updateRoom(room.id, { status: 'clean' })
    room.status = 'clean'
    toast?.success(`Room ${room.ext} marked clean`)
  } catch (err) {
    toast?.error('Failed to update room status', err.message)
  }
}

// --- Wake-up Calls ---
const fetchWakeupCalls = async () => {
  try {
    const res = await wakeupCallsAPI.list()
    wakeupCalls.value = res.data || []
  } catch (err) {
    toast?.error('Failed to load wake-up calls', err.message)
    wakeupCalls.value = []
  }
}

const scheduleWakeUp = async () => {
  const ext = prompt('Room Extension:')
  if (!ext) return
  const time = prompt('Wake-up time (HH:MM):', '07:00')
  if (!time) return
  try {
    await wakeupCallsAPI.create({ extension: ext, time })
    toast?.success(`Wake-up call scheduled for room ${ext} at ${time}`)
    await fetchWakeupCalls()
  } catch (err) {
    toast?.error('Failed to schedule wake-up call', err.message)
  }
}

const cancelWakeup = async (call) => {
  if (!call?.id) return
  try {
    await wakeupCallsAPI.cancel(call.id)
    toast?.success(`Wake-up call cancelled for room ${call.extension}`)
    await fetchWakeupCalls()
  } catch (err) {
    toast?.error('Failed to cancel wake-up call', err.message)
  }
}

// --- PMS Config ---
const savePMSConfig = async () => {
  toast?.info('PMS configuration saved')
}

// --- Service Codes ---
const saveCodes = async () => {
  toast?.info('Service codes saved')
}

// --- Helpers ---
const formatTime = (time) => {
  if (!time) return ''
  // Handle HH:MM format
  if (time.includes(':')) {
    const [h, m] = time.split(':')
    const hour = parseInt(h, 10)
    const ampm = hour >= 12 ? 'PM' : 'AM'
    const displayHour = hour % 12 || 12
    return `${displayHour}:${m} ${ampm}`
  }
  return time
}

onMounted(() => {
  fetchRooms()
  fetchWakeupCalls()
})
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 24px; border-radius: 0 0 4px 4px; }

.stat-item { display: flex; flex-direction: column; align-items: center; min-width: 60px; }
.stat-item .label { font-size: 10px; text-transform: uppercase; color: #94a3b8; font-weight: 700; }
.stat-item .val { font-size: 18px; font-weight: 700; color: #334155; }
.text-green-600 { color: #16a34a; }
.text-red-600 { color: #dc2626; }
.text-amber-600 { color: #d97706; }

.rooms-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 16px; }
.room-card { border: 1px solid #e2e8f0; border-radius: 8px; padding: 12px; display: flex; flex-direction: column; gap: 4px; background: #f8fafc; }
.room-card.clean { border-left: 4px solid #22c55e; background: #f0fdf4; }
.room-card.dirty { border-left: 4px solid #ef4444; background: #fef2f2; }
.room-card.inspect { border-left: 4px solid #f59e0b; background: #fffbeb; }

.room-header { display: flex; justify-content: space-between; font-weight: 700; font-size: 16px; }
.room-guest { font-size: 12px; color: #64748b; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.room-status-badge { font-size: 10px; font-weight: 700; margin-top: 4px; color: #475569; }
.room-actions { margin-top: 8px; }
.btn-xs { font-size: 10px; padding: 2px 6px; border: 1px solid #cbd5e1; background: white; border-radius: 4px; cursor: pointer; }

.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.help-text { font-size: 10px; color: #64748b; }
</style>
