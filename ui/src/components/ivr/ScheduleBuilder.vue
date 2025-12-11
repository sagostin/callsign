<template>
  <div class="scheduler">
    <div class="scheduler-header">
      <div class="header-row">
        <input type="text" class="schedule-name-input" value="Business Hours" placeholder="Schedule Name">
        <select class="timezone-select">
          <option>America/Los_Angeles (PST)</option>
          <option>America/New_York (EST)</option>
        </select>
      </div>
    </div>

    <div class="schedule-grid">
      <div class="day-row" v-for="(day, index) in weekDays" :key="day">
        <div class="day-label">{{ day }}</div>
        <div class="hours-config">
          <div v-if="schedule[index].closed" class="closed-badge">Closed</div>
          <div v-else class="time-ranges">
            <span class="status-open">Open</span>
            <div class="range">9:00 AM - 5:00 PM</div>
            <button class="btn-icon"><Edit2 class="icon-xs" /></button>
          </div>
        </div>
        <div class="day-toggle">
          <label class="switch small">
            <input type="checkbox" :checked="!schedule[index].closed" @change="schedule[index].closed = !schedule[index].closed">
            <span class="slider round"></span>
          </label>
        </div>
      </div>
    </div>
    
    <!-- Closed Routing -->
    <div class="closed-routing">
       <div class="section-title">When Closed (Default Route)</div>
       <div class="routing-config">
          <span class="text-sm">If no schedule matches or status is Closed, route to:</span>
          <div class="input-group">
            <select class="input-field small">
               <option>Voicemail</option>
               <option>IVR</option>
               <option>Disconnect</option>
            </select>
            <select class="input-field">
               <option>General Box</option>
               <option>After Hours IVR</option>
            </select>
          </div>
       </div>
    </div>

    <div class="holidays-section">
      <h3>Holiday List</h3>
      <div class="holiday-config">
        <p class="text-sm text-muted">Select a holiday list to apply to this schedule.</p>
        <div class="holiday-selector">
          <CalendarDays class="icon-sm" />
          <select class="input-field full-width">
            <option value="">No Holidays</option>
            <option value="us-federal" selected>US Federal 2024 (External)</option>
            <option value="office-closures">Office Closures (Manual)</option>
          </select>
        </div>
        <button class="btn-link">View/Edit List</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Edit2, CalendarDays } from 'lucide-vue-next'

const weekDays = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']

const schedule = ref([
  { closed: false }, { closed: false }, { closed: false }, { closed: false }, { closed: false },
  { closed: true }, { closed: true }
])
</script>

<style scoped>
.scheduler {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.scheduler-header {
  padding: 16px;
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.schedule-name-input {
  border: none;
  background: transparent;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  outline: none;
  width: 200px;
}
.schedule-name-input:focus { border-bottom: 1px solid var(--primary-color); }

.timezone-select {
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  font-size: 12px;
}

.schedule-grid {
  padding: 16px;
}

.day-row {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color);
}
.day-row:last-child { border-bottom: none; }

.day-label { width: 100px; font-weight: 600; font-size: 13px; color: var(--text-primary); }

.hours-config { flex: 1; }

.closed-badge {
  display: inline-block;
  background: var(--bg-app);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 700;
  padding: 4px 8px;
  border-radius: 4px;
  text-transform: uppercase;
  border: 1px solid var(--border-color);
}

.status-open {
  font-size: 11px;
  font-weight: 700;
  color: #166534;
  background: #dcfce7;
  padding: 2px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  margin-right: 8px;
}

.time-ranges { display: flex; gap: 8px; align-items: center; }
.range { font-family: monospace; font-size: 13px; background: #E0F2FE; color: #0284C7; padding: 4px 8px; border-radius: 4px; }

.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--primary-color); }

.btn-icon:hover { color: var(--primary-color); }

.closed-routing {
  padding: 16px;
  background: #F8FAFC;
  border-top: 1px solid var(--border-color);
}
.routing-config { display: flex; align-items: center; gap: 12px; margin-top: 8px; }
.input-group { display: flex; gap: 8px; }
.input-field.small { width: 120px; }

.switch { position: relative; display: inline-block; width: 32px; height: 18px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .4s; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 2px; bottom: 2px; background-color: white; transition: .4s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(14px); }
.slider.round { border-radius: 34px; }
.slider.round:before { border-radius: 50%; }

.holidays-section {
  padding: 16px;
  background: var(--bg-app);
  border-top: 1px solid var(--border-color);
}

.holidays-section h3 { font-size: 12px; text-transform: uppercase; color: var(--text-muted); margin-bottom: 12px; font-weight: 700; }

.holiday-config {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.holiday-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--border-color);
  padding: 8px;
  border-radius: var(--radius-sm);
  background: white;
}

.full-width { width: 100%; border: none; outline: none; background: transparent; font-size: 13px; }
.text-sm { font-size: 12px; }

.btn-link { 
  background: none; border: none; color: var(--primary-color); cursor: pointer; font-size: 12px; font-weight: 600; text-align: left; width: fit-content; padding: 0;
}


.icon-xs { width: 14px; height: 14px; }
</style>
