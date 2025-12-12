<template>
  <div class="view-container">
    <div class="page-header">
      <h2>Voicemail</h2>
      <div class="filters">
        <button class="filter-btn" :class="{ active: filter === 'new' }" @click="filter = 'new'">
          New ({{ newCount }})
        </button>
        <button class="filter-btn" :class="{ active: filter === 'saved' }" @click="filter = 'saved'">
          Saved ({{ savedCount }})
        </button>
        <button class="filter-btn" :class="{ active: filter === 'all' }" @click="filter = 'all'">
          All
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-state">
      <RefreshCw class="spin" /> Loading messages...
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredMessages.length === 0" class="empty-state">
      <Voicemail class="empty-icon" />
      <p>No {{ filter }} voicemail messages</p>
    </div>

    <!-- Message List -->
    <div v-else class="vm-list">
      <div 
        v-for="msg in filteredMessages" 
        :key="msg.id" 
        class="vm-item" 
        :class="{ unread: msg.is_new }"
      >
        <div class="vm-icon" :class="{ read: !msg.is_new }">
          <Voicemail class="icon" />
        </div>
        <div class="vm-details">
          <div class="vm-row">
            <span class="caller">{{ msg.caller_name || msg.caller_id || 'Unknown' }}</span>
            <span v-if="msg.is_new" class="new-tag">Unread</span>
          </div>
          <div class="vm-meta">
            {{ formatDate(msg.created_at) }} • {{ formatDuration(msg.duration) }}
          </div>
        </div>
        <div class="vm-actions">
          <button 
            class="btn-icon circle" 
            @click="playMessage(msg)"
            :class="{ playing: currentPlaying === msg.id }"
          >
            <StopCircle v-if="currentPlaying === msg.id" class="icon-sm" />
            <Play v-else class="icon-sm" />
          </button>
          <button v-if="msg.is_new" class="btn-secondary small" @click="markRead(msg)">
            Mark Read
          </button>
          <button class="btn-icon" @click="deleteMessage(msg)" title="Delete">
            <Trash class="icon-sm" />
          </button>
        </div>
      </div>
    </div>

    <!-- Audio Player Bar -->
    <transition name="slide-up">
      <div v-if="currentMessage" class="audio-player-bar">
        <div class="player-info">
          <div class="player-icon">
            <Voicemail class="icon" />
          </div>
          <div class="player-details">
            <span class="player-title">{{ currentMessage.caller_name || currentMessage.caller_id }}</span>
            <span class="player-meta">{{ formatDate(currentMessage.created_at) }}</span>
          </div>
        </div>
        <div class="player-controls">
          <button class="ctrl-btn" @click="togglePause">
            <Play v-if="isPaused" class="ctrl-icon" />
            <Pause v-else class="ctrl-icon" />
          </button>
          <div class="progress-container">
            <span class="time-display">{{ formatTime(currentTime) }}</span>
            <div class="progress-bar" @click="seek">
              <div class="progress-fill" :style="{ width: progress + '%' }"></div>
            </div>
            <span class="time-display">{{ formatTime(duration) }}</span>
          </div>
        </div>
        <button class="close-player" @click="stopPlayback">
          <X />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Voicemail, Play, Pause, StopCircle, Trash, RefreshCw, X } from 'lucide-vue-next'
import { voicemailAPI, authAPI } from '../../services/api'

const messages = ref([])
const loading = ref(false)
const filter = ref('new')
const currentPlaying = ref(null)
const currentMessage = ref(null)
const isPaused = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const progress = ref(0)
let audioPlayer = null

const newCount = computed(() => messages.value.filter(m => m.is_new).length)
const savedCount = computed(() => messages.value.filter(m => !m.is_new).length)

const filteredMessages = computed(() => {
  if (filter.value === 'new') return messages.value.filter(m => m.is_new)
  if (filter.value === 'saved') return messages.value.filter(m => !m.is_new)
  return messages.value
})

const loadMessages = async () => {
  loading.value = true
  try {
    // Get current user's extension
    const profileRes = await authAPI.getProfile()
    const extension = profileRes.data?.data?.extension || profileRes.data?.extension
    if (!extension) {
      console.warn('No extension found for user')
      return
    }
    
    const response = await voicemailAPI.listMessages(extension)
    messages.value = response.data?.data || []
  } catch (error) {
    console.error('Failed to load voicemail messages:', error)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('en-US', { 
    month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' 
  })
}

const formatDuration = (seconds) => {
  if (!seconds) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const formatTime = (seconds) => formatDuration(seconds)

const playMessage = (msg) => {
  if (currentPlaying.value === msg.id) {
    stopPlayback()
    return
  }
  
  stopPlayback()
  
  const token = localStorage.getItem('token')
  const url = voicemailAPI.streamUrl(msg.id) + `?token=${encodeURIComponent(token)}`
  
  audioPlayer = new Audio(url)
  currentMessage.value = msg
  
  audioPlayer.onloadedmetadata = () => {
    duration.value = audioPlayer.duration
  }
  
  audioPlayer.ontimeupdate = () => {
    currentTime.value = audioPlayer.currentTime
    progress.value = (audioPlayer.currentTime / audioPlayer.duration) * 100
  }
  
  audioPlayer.onended = () => {
    stopPlayback()
    // Auto mark as read if was new
    if (msg.is_new) {
      markRead(msg)
    }
  }
  
  audioPlayer.play()
    .then(() => {
      currentPlaying.value = msg.id
      isPaused.value = false
    })
    .catch(err => {
      console.error('Playback failed:', err)
      stopPlayback()
    })
}

const stopPlayback = () => {
  if (audioPlayer) {
    audioPlayer.pause()
    audioPlayer.src = ''
    audioPlayer = null
  }
  currentPlaying.value = null
  currentMessage.value = null
  currentTime.value = 0
  duration.value = 0
  progress.value = 0
  isPaused.value = false
}

const togglePause = () => {
  if (!audioPlayer) return
  if (isPaused.value) {
    audioPlayer.play()
    isPaused.value = false
  } else {
    audioPlayer.pause()
    isPaused.value = true
  }
}

const seek = (event) => {
  if (!audioPlayer || !duration.value) return
  const bar = event.currentTarget
  const rect = bar.getBoundingClientRect()
  const percent = (event.clientX - rect.left) / rect.width
  audioPlayer.currentTime = percent * duration.value
}

const markRead = async (msg) => {
  try {
    await voicemailAPI.markRead(msg.id)
    msg.is_new = false
  } catch (error) {
    console.error('Failed to mark as read:', error)
  }
}

const deleteMessage = async (msg) => {
  if (!confirm('Delete this voicemail message?')) return
  try {
    await voicemailAPI.deleteMessage(msg.id)
    messages.value = messages.value.filter(m => m.id !== msg.id)
    if (currentPlaying.value === msg.id) {
      stopPlayback()
    }
  } catch (error) {
    console.error('Failed to delete message:', error)
    alert('Failed to delete message')
  }
}

onMounted(() => {
  loadMessages()
})

onUnmounted(() => {
  stopPlayback()
})
</script>

<style scoped>
.view-container { padding: 8px 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h2 { font-size: 20px; font-weight: 700; color: var(--text-primary); }

.filters { display: flex; gap: 8px; background: #f1f5f9; padding: 4px; border-radius: 8px; }
.filter-btn { padding: 6px 12px; border: none; background: none; font-size: 13px; font-weight: 600; color: #64748b; cursor: pointer; border-radius: 6px; }
.filter-btn.active { background: white; color: var(--primary-color); box-shadow: 0 1px 2px rgba(0,0,0,0.1); }

.loading-state, .empty-state { text-align: center; padding: 60px 20px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; margin-bottom: 12px; opacity: 0.5; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.vm-list { display: flex; flex-direction: column; gap: 12px; }
.vm-item {
  display: flex; align-items: center; gap: 16px;
  background: white; border: 1px solid var(--border-color);
  padding: 16px; border-radius: var(--radius-md);
  transition: all 0.2s;
}
.vm-item.unread { border-left: 3px solid var(--primary-color); background: #F8FAFC; }
.vm-icon {
  width: 40px; height: 40px; background: #e2e8f0; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #64748b;
}
.vm-item.unread .vm-icon { background: var(--primary-light); color: var(--primary-color); }

.vm-details { flex: 1; }
.vm-row { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.caller { font-weight: 600; color: var(--text-primary); }
.new-tag { font-size: 10px; background: var(--primary-color); color: white; padding: 2px 6px; border-radius: 99px; font-weight: 700; }
.vm-meta { font-size: 12px; color: var(--text-muted); }

.vm-actions { display: flex; align-items: center; gap: 8px; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 8px; border-radius: 6px; }
.btn-icon:hover { background: #f1f5f9; color: var(--text-primary); }
.btn-icon.circle { border: 1px solid var(--border-color); border-radius: 50%; width: 36px; height: 36px; padding: 0; display: flex; align-items: center; justify-content: center; }
.btn-icon.circle.playing { background: var(--primary-color); border-color: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 6px 12px; border-radius: 6px; font-size: 12px; font-weight: 600; cursor: pointer; }
.btn-secondary:hover { border-color: var(--primary-color); color: var(--primary-color); }
.icon { width: 20px; height: 20px; }
.icon-sm { width: 16px; height: 16px; }

/* Audio Player Bar */
.audio-player-bar {
  position: fixed; bottom: 0; left: 0; right: 0;
  height: 64px; background: #1e293b;
  display: flex; align-items: center; padding: 0 24px; gap: 16px;
  z-index: 50; box-shadow: 0 -2px 10px rgba(0,0,0,0.2);
}
.player-info { display: flex; align-items: center; gap: 12px; min-width: 180px; }
.player-icon { width: 36px; height: 36px; background: rgba(255,255,255,0.1); border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; }
.player-details { display: flex; flex-direction: column; }
.player-title { color: white; font-size: 13px; font-weight: 600; }
.player-meta { color: rgba(255,255,255,0.6); font-size: 11px; }

.player-controls { display: flex; align-items: center; gap: 12px; flex: 1; }
.ctrl-btn { width: 36px; height: 36px; background: rgba(255,255,255,0.1); border: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; }
.ctrl-btn:hover { background: rgba(255,255,255,0.2); }
.ctrl-icon { width: 16px; height: 16px; color: white; }

.progress-container { display: flex; align-items: center; gap: 10px; flex: 1; }
.time-display { color: rgba(255,255,255,0.7); font-size: 11px; font-family: monospace; min-width: 40px; }
.progress-bar { flex: 1; height: 6px; background: rgba(255,255,255,0.2); border-radius: 3px; cursor: pointer; }
.progress-fill { height: 100%; background: var(--primary-color); border-radius: 3px; }

.close-player { background: none; border: none; color: rgba(255,255,255,0.6); cursor: pointer; padding: 8px; }
.close-player:hover { color: white; }

/* Transitions */
.slide-up-enter-active, .slide-up-leave-active { transition: transform 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(100%); }
</style>
