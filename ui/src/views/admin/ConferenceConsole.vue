<template>
  <div class="console-container">
    <div class="console-header">
      <div class="flex items-center gap-3">
         <div class="status-indicator live"></div>
         <div>
            <h1 class="text-xl font-bold text-white">{{ confTitle }} <span class="text-slate-400 font-normal">#{{ confNumber }}</span></h1>
            <p class="text-xs text-slate-400">{{ elapsed }} Elapsed • {{ participants.length }} Participants</p>
         </div>
      </div>
      <div class="flex gap-2">
         <button class="btn-console" :class="{ red: isRecording }" @click="toggleRecording"><SquareIcon class="w-4 h-4 mr-2" /> {{ isRecording ? 'Stop Rec' : 'Record' }}</button>
         <button class="btn-console" @click="toggleLock"><LockIcon class="w-4 h-4 mr-2" /> {{ isLocked ? 'Unlock' : 'Lock Room' }}</button>
         <button class="btn-console" @click="muteAll"><MicOffIcon class="w-4 h-4 mr-2" /> Mute All</button>
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
             <button class="action-btn" :class="{ active: p.muted }" @click="toggleMute(p)"><MicOffIcon class="w-4 h-4" /></button>
             <button class="action-btn text-red-500" @click="kickMember(p)"><UserMinusIcon class="w-4 h-4" /></button>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, inject } from 'vue'
import { useRoute } from 'vue-router'
import { MicIcon, MicOffIcon, LockIcon, SquareIcon, UserMinusIcon } from 'lucide-vue-next'
import { conferencesAPI } from '../../services/api'

const toast = inject('toast')
const route = useRoute()
const confName = ref(route.params.id || route.params.name || '')
const confTitle = ref('Conference')
const confNumber = ref('')
const elapsed = ref('00:00:00')
const participants = ref([])
const isRecording = ref(false)
const isLocked = ref(false)
let refreshInterval = null

const fetchLiveData = async () => {
  if (!confName.value) return
  try {
    const res = await conferencesAPI.getLive(confName.value)
    const data = res.data || {}
    confTitle.value = data.name || data.conference_name || confName.value
    confNumber.value = data.number || data.extension || ''
    elapsed.value = data.elapsed || data.run_time || '00:00:00'
    isRecording.value = data.recording || false
    isLocked.value = data.locked || false
    const members = data.members || data.participants || data || []
    participants.value = (Array.isArray(members) ? members : []).map(m => ({
      id: m.id || m.member_id,
      name: m.caller_id_name || m.name || 'Unknown',
      number: m.caller_id_number || m.number || '',
      talking: m.talking || false,
      muted: m.muted || false
    }))
  } catch (err) {
    console.error('Failed to load conference:', err)
    participants.value = []
  }
}

onMounted(() => {
  fetchLiveData()
  refreshInterval = setInterval(fetchLiveData, 3000) // Refresh every 3s
})
onUnmounted(() => { if (refreshInterval) clearInterval(refreshInterval) })

const toggleMute = async (p) => {
  try {
    if (p.muted) {
      await conferencesAPI.unmuteMember(confName.value, p.id)
    } else {
      await conferencesAPI.muteMember(confName.value, p.id)
    }
    p.muted = !p.muted
  } catch (err) {
    toast?.error(err.message, 'Failed to toggle mute')
  }
}

const kickMember = async (p) => {
  if (!confirm(`Kick ${p.name} from conference?`)) return
  try {
    await conferencesAPI.kickMember(confName.value, p.id)
    participants.value = participants.value.filter(m => m.id !== p.id)
    toast?.success(`${p.name} kicked`)
  } catch (err) {
    toast?.error(err.message, 'Failed to kick member')
  }
}

const toggleRecording = async () => {
  try {
    if (isRecording.value) {
      await conferencesAPI.stopRecording(confName.value)
      toast?.success('Recording stopped')
    } else {
      await conferencesAPI.startRecording(confName.value)
      toast?.success('Recording started')
    }
    isRecording.value = !isRecording.value
  } catch (err) {
    toast?.error(err.message, 'Failed to toggle recording')
  }
}

const toggleLock = async () => {
  try {
    if (isLocked.value) {
      await conferencesAPI.unlockConference(confName.value)
    } else {
      await conferencesAPI.lockConference(confName.value)
    }
    isLocked.value = !isLocked.value
  } catch (err) {
    toast?.error(err.message, 'Failed to toggle lock')
  }
}

const muteAll = async () => {
  try {
    await conferencesAPI.muteAll(confName.value)
    participants.value.forEach(p => p.muted = true)
    toast?.success('All participants muted')
  } catch (err) {
    toast?.error(err.message, 'Failed to mute all')
  }
}
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
