<template>
  <div class="recordings-page">
    <div class="view-header">
      <div class="header-content">
        <h2>My Recordings</h2>
        <p class="text-muted text-sm">Listen to and manage your call recordings.</p>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon total"><MicIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ recordings.length }}</span>
          <span class="stat-label">Total Recordings</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon personal"><UserIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ recordings.filter(r => !r.group).length }}</span>
          <span class="stat-label">Personal</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon group"><UsersIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ recordings.filter(r => r.group).length }}</span>
          <span class="stat-label">Group/Queue</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon storage"><HardDriveIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ totalStorage }}</span>
          <span class="stat-label">Storage Used</span>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" class="search-input" placeholder="Search by name, number, or date...">
      </div>
      <select v-model="filterType" class="filter-select">
        <option value="">All Recordings</option>
        <option value="personal">Personal Only</option>
        <option value="group">Group/Queue Only</option>
      </select>
      <select v-model="filterDateRange" class="filter-select">
        <option value="">All Time</option>
        <option value="today">Today</option>
        <option value="week">This Week</option>
        <option value="month">This Month</option>
      </select>
    </div>

    <!-- Audio Player -->
    <div class="audio-player" v-if="currentlyPlaying">
      <div class="player-info">
        <div class="player-icon"><PlayCircleIcon class="icon" /></div>
        <div class="player-details">
          <span class="player-title">{{ currentlyPlaying.from }} → {{ currentlyPlaying.to }}</span>
          <span class="player-meta">{{ currentlyPlaying.date }} at {{ currentlyPlaying.time }}</span>
        </div>
      </div>
      <div class="player-controls">
        <div class="waveform">
          <div class="wave-bar" v-for="i in 30" :key="i" :style="{ height: Math.random() * 100 + '%' }"></div>
        </div>
        <div class="time-display">
          <span>{{ playbackTime }}</span>
          <span>/</span>
          <span>{{ currentlyPlaying.duration }}</span>
        </div>
      </div>
      <div class="player-actions">
        <button class="player-btn" @click="seekBackward"><SkipBackIcon class="icon-sm" /></button>
        <button class="player-btn play" @click="togglePlayback">
          <PauseIcon v-if="isPlaying" class="icon" />
          <PlayIcon v-else class="icon" />
        </button>
        <button class="player-btn" @click="seekForward"><SkipForwardIcon class="icon-sm" /></button>
        <button class="player-btn" @click="stopPlayback"><XIcon class="icon-sm" /></button>
      </div>
    </div>

    <!-- Recordings List -->
    <div class="recordings-list">
      <div class="recording-item" 
           v-for="rec in filteredRecordings" 
           :key="rec.id"
           :class="{ playing: currentlyPlaying?.id === rec.id }">
        <div class="recording-icon" :class="rec.direction">
          <PhoneIncomingIcon v-if="rec.direction === 'inbound'" class="icon-sm" />
          <PhoneOutgoingIcon v-else class="icon-sm" />
        </div>
        
        <div class="recording-main">
          <div class="recording-parties">
            <span class="party from">{{ rec.from }}</span>
            <ArrowRightIcon class="arrow-icon" />
            <span class="party to">{{ rec.to }}</span>
          </div>
          <div class="recording-meta">
            <span class="meta-item"><CalendarIcon class="meta-icon" /> {{ rec.date }}</span>
            <span class="meta-item"><ClockIcon class="meta-icon" /> {{ rec.time }}</span>
            <span class="meta-item"><TimerIcon class="meta-icon" /> {{ rec.duration }}</span>
          </div>
        </div>

        <div class="recording-context">
          <span class="context-badge" :class="rec.group ? 'group' : 'personal'">
            {{ rec.group ? rec.group : 'Personal' }}
          </span>
        </div>

        <div class="recording-actions">
          <button class="action-btn play" @click="playRecording(rec)" :title="isPlayingThis(rec) ? 'Pause' : 'Play'">
            <PauseIcon v-if="isPlayingThis(rec)" class="icon-sm" />
            <PlayIcon v-else class="icon-sm" />
          </button>
          <button class="action-btn" @click="downloadRecording(rec)" title="Download">
            <DownloadIcon class="icon-sm" />
          </button>
          <button class="action-btn" @click="showTranscript(rec)" title="Transcript" v-if="rec.hasTranscript">
            <FileTextIcon class="icon-sm" />
          </button>
          <button class="action-btn danger" @click="deleteRecording(rec)" title="Delete">
            <TrashIcon class="icon-sm" />
          </button>
        </div>
      </div>

      <div class="empty-state" v-if="filteredRecordings.length === 0">
        <MicOffIcon class="empty-icon" />
        <p>No recordings found</p>
        <span class="text-muted text-sm">Try adjusting your filters</span>
      </div>
    </div>

    <!-- Transcript Modal -->
    <div v-if="showingTranscript" class="modal-overlay" @click.self="showingTranscript = null">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Call Transcript</h3>
          <button class="btn-icon" @click="showingTranscript = null"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="transcript-meta">
            <span>{{ showingTranscript.from }} → {{ showingTranscript.to }}</span>
            <span>{{ showingTranscript.date }} at {{ showingTranscript.time }}</span>
          </div>
          <div class="transcript-content">
            <div class="transcript-line" v-for="(line, i) in sampleTranscript" :key="i">
              <span class="speaker" :class="line.speaker">{{ line.speaker === 'caller' ? 'Caller' : 'Agent' }}</span>
              <span class="timestamp">{{ line.time }}</span>
              <p class="text">{{ line.text }}</p>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="copyTranscript">Copy Text</button>
          <button class="btn-primary" @click="downloadTranscript">Download</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Mic as MicIcon, MicOff as MicOffIcon, User as UserIcon, Users as UsersIcon,
  HardDrive as HardDriveIcon, Search as SearchIcon, 
  PhoneIncoming as PhoneIncomingIcon, PhoneOutgoing as PhoneOutgoingIcon,
  ArrowRight as ArrowRightIcon, Calendar as CalendarIcon, Clock as ClockIcon, Timer as TimerIcon,
  Play as PlayIcon, Pause as PauseIcon, PlayCircle as PlayCircleIcon,
  SkipBack as SkipBackIcon, SkipForward as SkipForwardIcon,
  Download as DownloadIcon, FileText as FileTextIcon, Trash2 as TrashIcon, X as XIcon
} from 'lucide-vue-next'

const searchQuery = ref('')
const filterType = ref('')
const filterDateRange = ref('')
const currentlyPlaying = ref(null)
const isPlaying = ref(false)
const playbackTime = ref('00:00')
const showingTranscript = ref(null)

const recordings = ref([
  { id: 1, date: 'Dec 9, 2024', time: '2:30 PM', from: '(415) 555-0199', to: 'Ext 101', duration: '04:12', direction: 'inbound', group: null, hasTranscript: true, size: '2.1 MB' },
  { id: 2, date: 'Dec 9, 2024', time: '11:15 AM', from: '(415) 555-0200', to: 'Sales Queue', duration: '12:05', direction: 'inbound', group: 'Sales', hasTranscript: true, size: '5.8 MB' },
  { id: 3, date: 'Dec 8, 2024', time: '4:45 PM', from: 'Ext 101', to: '(415) 555-0300', duration: '08:33', direction: 'outbound', group: null, hasTranscript: false, size: '4.2 MB' },
  { id: 4, date: 'Dec 8, 2024', time: '10:30 AM', from: '(212) 555-9876', to: 'Support', duration: '15:22', direction: 'inbound', group: 'Support', hasTranscript: true, size: '7.3 MB' },
  { id: 5, date: 'Dec 7, 2024', time: '3:00 PM', from: 'Ext 101', to: '(310) 555-4567', duration: '02:45', direction: 'outbound', group: null, hasTranscript: false, size: '1.4 MB' },
])

const sampleTranscript = [
  { speaker: 'caller', time: '00:00', text: 'Hi, I was calling about my recent order.' },
  { speaker: 'agent', time: '00:05', text: 'Of course! I\'d be happy to help. Can I get your order number?' },
  { speaker: 'caller', time: '00:12', text: 'Sure, it\'s 12345-ABC.' },
  { speaker: 'agent', time: '00:18', text: 'Thank you. I can see your order here. What seems to be the issue?' },
]

const totalStorage = computed(() => {
  const total = recordings.value.reduce((sum, r) => sum + parseFloat(r.size), 0)
  return total.toFixed(1) + ' MB'
})

const filteredRecordings = computed(() => {
  return recordings.value.filter(r => {
    const matchesSearch = !searchQuery.value || 
      r.from.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      r.to.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      r.date.toLowerCase().includes(searchQuery.value.toLowerCase())
    
    const matchesType = !filterType.value ||
      (filterType.value === 'personal' && !r.group) ||
      (filterType.value === 'group' && r.group)
    
    return matchesSearch && matchesType
  })
})

const isPlayingThis = (rec) => currentlyPlaying.value?.id === rec.id && isPlaying.value

const playRecording = (rec) => {
  if (currentlyPlaying.value?.id === rec.id) {
    isPlaying.value = !isPlaying.value
  } else {
    currentlyPlaying.value = rec
    isPlaying.value = true
    playbackTime.value = '00:00'
  }
}

const togglePlayback = () => { isPlaying.value = !isPlaying.value }
const stopPlayback = () => { currentlyPlaying.value = null; isPlaying.value = false }
const seekBackward = () => console.log('Seek -10s')
const seekForward = () => console.log('Seek +10s')

const downloadRecording = (rec) => alert(`Downloading ${rec.from} - ${rec.to}`)
const deleteRecording = (rec) => {
  if (confirm('Delete this recording?')) {
    recordings.value = recordings.value.filter(r => r.id !== rec.id)
  }
}

const showTranscript = (rec) => { showingTranscript.value = rec }
const copyTranscript = () => alert('Transcript copied!')
const downloadTranscript = () => alert('Downloading transcript...')
</script>

<style scoped>
.recordings-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.total { background: #dbeafe; color: #2563eb; }
.stat-icon.personal { background: #dcfce7; color: #16a34a; }
.stat-icon.group { background: #fef3c7; color: #b45309; }
.stat-icon.storage { background: #f3e8ff; color: #7c3aed; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Filters */
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-box { position: relative; flex: 1; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 10px 12px 10px 38px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.filter-select { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }

/* Audio Player */
.audio-player {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 16px 20px;
  background: linear-gradient(135deg, var(--primary-color), #818cf8);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
  color: white;
}

.player-info { display: flex; align-items: center; gap: 12px; }
.player-icon { opacity: 0.8; }
.player-icon .icon { width: 32px; height: 32px; }
.player-details { display: flex; flex-direction: column; }
.player-title { font-weight: 600; font-size: 14px; }
.player-meta { font-size: 12px; opacity: 0.8; }

.player-controls { flex: 1; display: flex; align-items: center; gap: 16px; }
.waveform { display: flex; align-items: center; gap: 2px; height: 32px; flex: 1; }
.wave-bar { width: 3px; background: rgba(255,255,255,0.5); border-radius: 2px; min-height: 4px; }
.time-display { font-family: monospace; font-size: 12px; opacity: 0.9; }

.player-actions { display: flex; gap: 8px; }
.player-btn { width: 36px; height: 36px; border-radius: 50%; border: none; background: rgba(255,255,255,0.2); color: white; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.player-btn:hover { background: rgba(255,255,255,0.3); }
.player-btn.play { width: 44px; height: 44px; background: white; color: var(--primary-color); }

/* Recordings List */
.recordings-list { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }

.recording-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.15s;
}
.recording-item:last-child { border-bottom: none; }
.recording-item:hover { background: var(--bg-app); }
.recording-item.playing { background: var(--primary-light); }

.recording-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.recording-icon.inbound { background: #dcfce7; color: #16a34a; }
.recording-icon.outbound { background: #dbeafe; color: #2563eb; }

.recording-main { flex: 1; }
.recording-parties { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.party { font-weight: 600; font-size: 14px; }
.arrow-icon { width: 14px; height: 14px; color: var(--text-muted); }

.recording-meta { display: flex; gap: 16px; }
.meta-item { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--text-muted); }
.meta-icon { width: 12px; height: 12px; }

.recording-context { min-width: 100px; }
.context-badge { font-size: 11px; font-weight: 600; padding: 4px 10px; border-radius: 20px; text-transform: uppercase; }
.context-badge.personal { background: #f1f5f9; color: #64748b; }
.context-badge.group { background: #dbeafe; color: #2563eb; }

.recording-actions { display: flex; gap: 4px; }
.action-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  transition: all 0.15s;
}
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.play { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }

.empty-state { text-align: center; padding: 48px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 16px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 560px; max-height: 80vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.transcript-meta { display: flex; justify-content: space-between; font-size: 12px; color: var(--text-muted); margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid var(--border-color); }

.transcript-content { display: flex; flex-direction: column; gap: 12px; }
.transcript-line { padding: 12px; background: var(--bg-app); border-radius: 8px; }
.transcript-line .speaker { font-size: 11px; font-weight: 700; text-transform: uppercase; }
.transcript-line .speaker.caller { color: #2563eb; }
.transcript-line .speaker.agent { color: #16a34a; }
.transcript-line .timestamp { font-size: 10px; color: var(--text-muted); margin-left: 8px; }
.transcript-line .text { margin: 6px 0 0; font-size: 13px; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }

.icon-sm { width: 16px; height: 16px; }
.icon { width: 20px; height: 20px; }
</style>
