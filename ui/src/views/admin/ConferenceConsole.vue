<template>
  <div class="console-container">
    <div class="console-header">
      <div class="flex items-center gap-3">
         <div class="status-indicator live"></div>
         <div>
            <h1 class="text-xl font-bold text-white">Weekly Sales <span class="text-slate-400 font-normal">#3001</span></h1>
            <p class="text-xs text-slate-400">00:12:43 Elapsed • 4 Participants</p>
         </div>
      </div>
      <div class="flex gap-2">
         <button class="btn-console red"><SquareIcon class="w-4 h-4 mr-2" /> Stop Rec</button>
         <button class="btn-console"><LockIcon class="w-4 h-4 mr-2" /> Lock Room</button>
         <button class="btn-console"><MicOffIcon class="w-4 h-4 mr-2" /> Mute All</button>
      </div>
    </div>

    <div class="console-grid">
       <div class="participant-card" v-for="p in participants" :key="p.id">
          <div class="p-avatar">
             {{ p.name.charAt(0) }}
          </div>
          <div class="p-info">
             <div class="p-name">{{ p.name }}</div>
             <div class="p-number">{{ p.number }}</div>
          </div>
          <div class="p-status">
             <MicIcon v-if="p.talking" class="w-4 h-4 text-green-500 animate-pulse" />
             <MicOffIcon v-if="p.muted" class="w-4 h-4 text-red-500" />
          </div>
          <div class="p-actions">
             <button class="action-btn" :class="{ active: p.muted }" @click="p.muted = !p.muted"><MicOffIcon class="w-4 h-4" /></button>
             <button class="action-btn text-red-500"><UserMinusIcon class="w-4 h-4" /></button>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { MicIcon, MicOffIcon, LockIcon, SquareIcon, UserMinusIcon } from 'lucide-vue-next'

const participants = ref([
   { id: 1, name: 'Alice Smith', number: '101', talking: true, muted: false },
   { id: 2, name: 'Bob Jones', number: '102', talking: false, muted: false },
   { id: 3, name: 'External Caller', number: '+15550009999', talking: false, muted: true },
   { id: 4, name: 'Dave Wilson', number: '104', talking: false, muted: false },
])
</script>

<style scoped>
.console-container {
  background: #0f172a;
  min-height: 100vh;
  color: white;
  display: flex;
  flex-direction: column;
}

.console-header {
  padding: 20px 32px;
  background: #1e293b;
  border-bottom: 1px solid #334155;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-indicator { width: 12px; height: 12px; border-radius: 50%; background: #64748b; }
.status-indicator.live { background: #22c55e; box-shadow: 0 0 10px #22c55e; }

.btn-console {
  background: #334155;
  border: 1px solid #475569;
  color: #e2e8f0;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s;
}
.btn-console:hover { background: #475569; }
.btn-console.red { color: #fca5a5; border-color: #7f1d1d; background: #450a0a; }

.console-grid {
  padding: 32px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
}

.participant-card {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  position: relative;
}

.p-avatar {
  width: 48px; height: 48px; background: #3b82f6; color: white; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 18px;
}

.p-info { flex: 1; }
.p-name { font-weight: 600; font-size: 14px; }
.p-number { font-size: 12px; color: #94a3b8; font-mono: true; }

.p-actions {
  display: flex; gap: 8px;
}

.action-btn {
  width: 32px; height: 32px; border-radius: 6px; border: none; background: #334155; color: #cbd5e1;
  display: flex; align-items: center; justify-content: center; cursor: pointer;
}
.action-btn:hover { background: #475569; }
.action-btn.active { background: #ef4444; color: white; }
.text-red-500 { color: #f87171; }
</style>
