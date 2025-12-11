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
              <span class="val">124</span>
           </div>
           <div class="stat-item">
              <span class="label">Clean</span>
              <span class="val text-green-600">85</span>
           </div>
           <div class="stat-item">
              <span class="label">Dirty</span>
              <span class="val text-red-600">32</span>
           </div>
            <div class="stat-item">
              <span class="label">Insp</span>
              <span class="val text-amber-600">7</span>
           </div>
        </div>
        <div class="flex gap-2">
           <select class="input-field">
              <option>All Floors</option>
              <option>Floor 1</option>
              <option>Floor 2</option>
           </select>
           <button class="btn-secondary small">Refresh</button>
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
        <button class="btn-primary">Save Configuration</button>
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
            <button class="btn-primary">Save Codes</button>
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
           <!-- Mock Wake Up Data -->
           <div class="room-card clean">
              <div class="room-header">
                 <span class="room-number">101</span>
                 <span class="room-icon">⏰</span>
              </div>
              <div class="room-guest">John Doe</div>
              <div class="room-status-badge">07:00 AM</div>
              <div class="room-actions">
                  <button class="btn-xs text-red-600">Cancel</button>
              </div>
           </div>
       </div>
  </div>

</template>

<script setup>
import { ref } from 'vue'

const activeTab = ref('dashboard')

const rooms = ref([
   { ext: '101', status: 'dirty', guest: 'John Doe' },
   { ext: '102', status: 'clean', guest: 'Vacant' },
   { ext: '103', status: 'inspect', guest: 'Vacant' },
   { ext: '104', status: 'clean', guest: 'Jane Smith' },
   { ext: '105', status: 'dirty', guest: 'Vacant' },
   { ext: '201', status: 'clean', guest: 'Vacant' },
   { ext: '202', status: 'dirty', guest: 'Guest' },
])

const toggleStatus = (room) => {
   room.status = 'clean'
}

const scheduleWakeUp = () => {
   const ext = prompt('Room Extension:')
   if(ext) alert(`Scheduled wake up for ${ext} at 07:00 AM (Mock)`)
}
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
