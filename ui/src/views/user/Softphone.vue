<template>
  <div class="softphone-container">
    <!-- Device Binding Bar -->
    <div class="device-bar">
      <div class="device-selector">
        <span class="selector-label">Active Device:</span>
        <div class="device-dropdown" @click="showDeviceMenu = !showDeviceMenu">
          <div class="current-device">
            <component :is="getDeviceIcon(boundDevice)" class="device-icon" />
            <span class="device-name">{{ boundDevice?.name || 'Select Device' }}</span>
            <span class="device-status" :class="boundDevice?.status">{{ boundDevice?.status }}</span>
          </div>
          <ChevronDownIcon class="chevron" />
        </div>
        
        <!-- Device Menu -->
        <div class="device-menu" v-if="showDeviceMenu">
          <div class="menu-header">Your Devices</div>
          <div 
            class="device-option" 
            v-for="device in userDevices" 
            :key="device.id"
            :class="{ active: boundDevice?.id === device.id, disabled: device.status === 'offline' }"
            @click="bindDevice(device)"
          >
            <component :is="getDeviceIcon(device)" class="option-icon" />
            <div class="option-info">
              <span class="option-name">{{ device.name }}</span>
              <span class="option-meta">{{ device.meta }}</span>
            </div>
            <span class="option-status" :class="device.status"></span>
          </div>
        </div>
      </div>

      <!-- Active Call on Device Indicator -->
      <div class="active-call-indicator" v-if="deviceHasActiveCall">
        <div class="pulse-ring"></div>
        <PhoneCallIcon class="call-icon" />
        <span>Call on {{ boundDevice?.name }}</span>
        <button class="btn-take-control" @click="takeCallControl">Take Control</button>
      </div>
    </div>

    <div class="dialer-main">
      <!-- Dialer Panel -->
      <div class="dialer-panel">
        <!-- Connection Status -->
        <div class="connection-status" :class="sipState">
          <div class="status-dot"></div>
          <span>{{ sipStateLabel }}</span>
        </div>

        <!-- Call Display -->
        <div class="display">
          <div class="call-info" v-if="callState !== 'idle'">
            <span class="call-direction" :class="callDirection">{{ callDirection === 'inbound' ? 'Incoming' : 'Outgoing' }}</span>
            <span class="remote-name" v-if="remoteName">{{ remoteName }}</span>
            <span class="remote-number">{{ remoteNumber || number }}</span>
            <span class="call-status">{{ callStatusText }}</span>
            <span class="call-timer" v-if="callState === 'established'">{{ formattedDuration }}</span>
          </div>
          <input 
            v-else
            type="text" 
            v-model="number" 
            class="number-display" 
            placeholder="Enter number or extension..."
            @keyup.enter="makeCall"
          >
        </div>
        
        <!-- Keypad -->
        <div class="keypad" v-if="callState === 'idle' || callState === 'established'">
          <div class="key-row" v-for="row in keys" :key="row.join()">
            <button class="key-btn" v-for="key in row" :key="key.digit" @click="pressKey(key.digit)">
              <span class="digit">{{ key.digit }}</span>
              <span class="letters" v-if="key.letters">{{ key.letters }}</span>
            </button>
          </div>
        </div>

        <!-- In-Call Controls -->
        <div class="call-controls" v-if="callState !== 'idle'">
          <button class="control-btn" :class="{ active: isMuted }" @click="toggleMute" title="Mute">
            <MicOffIcon v-if="isMuted" class="control-icon" />
            <MicIcon v-else class="control-icon" />
            <span>{{ isMuted ? 'Unmute' : 'Mute' }}</span>
          </button>
          <button class="control-btn" :class="{ active: isOnHold }" @click="toggleHold" title="Hold">
            <PauseIcon class="control-icon" />
            <span>{{ isOnHold ? 'Resume' : 'Hold' }}</span>
          </button>
          <button class="control-btn" @click="showTransferModal = true" title="Transfer">
            <ArrowRightLeftIcon class="control-icon" />
            <span>Transfer</span>
          </button>
          <button class="control-btn" @click="showKeypad = !showKeypad" title="Keypad">
            <GridIcon class="control-icon" />
            <span>Keypad</span>
          </button>
        </div>

        <!-- Main Action Button -->
        <div class="actions">
          <button 
            v-if="callState === 'idle'"
            class="call-btn dial" 
            @click="makeCall"
            :disabled="!number || !canMakeCalls"
          >
            <PhoneIcon class="icon-lg" />
          </button>
          <button 
            v-else-if="callState === 'ringing' && callDirection === 'inbound'"
            class="call-btn answer"
            @click="answerCall"
          >
            <PhoneIcon class="icon-lg" />
          </button>
          <button 
            v-else
            class="call-btn hangup" 
            @click="hangupCall"
          >
            <PhoneOffIcon class="icon-lg" />
          </button>
        </div>

        <!-- Bound Device Info -->
        <div class="bound-device-info" v-if="boundDevice && boundDevice.type !== 'softphone'">
          <InfoIcon class="info-icon" />
          <span>Calls will ring on <strong>{{ boundDevice.name }}</strong></span>
        </div>
      </div>

      <!-- Recent Calls -->
      <div class="recent-calls">
        <div class="recent-header">
          <h3>Recent Calls</h3>
          <button class="btn-link" @click="$router.push('/history')">View All</button>
        </div>
        <div class="call-list">
          <div class="call-item" v-for="call in recentCalls" :key="call.id" @click="dialRecent(call)">
            <div class="call-icon" :class="call.type">
              <PhoneMissedIcon v-if="call.type === 'missed'" class="icon-sm" />
              <PhoneOutgoingIcon v-else-if="call.type === 'outgoing'" class="icon-sm" />
              <PhoneIncomingIcon v-else class="icon-sm" />
            </div>
            <div class="call-details">
              <span class="caller">{{ call.name || call.number }}</span>
              <span class="meta">{{ call.duration ? call.duration + ' - ' : '' }}{{ call.time }}</span>
            </div>
            <button class="call-back-btn" @click.stop="dialNumber(call.number)">
              <PhoneIcon class="icon-xs" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Transfer Modal -->
    <div v-if="showTransferModal" class="modal-overlay" @click.self="showTransferModal = false">
      <div class="modal-card small">
        <div class="modal-header">
          <h3>Transfer Call</h3>
          <button class="btn-icon" @click="showTransferModal = false"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="transfer-tabs">
            <button :class="{ active: transferType === 'blind' }" @click="transferType = 'blind'">Blind Transfer</button>
            <button :class="{ active: transferType === 'attended' }" @click="transferType = 'attended'">Attended</button>
          </div>
          <div class="form-group">
            <label>Transfer To</label>
            <input v-model="transferTarget" class="input-field" placeholder="Extension or number">
          </div>
          <div class="quick-transfer">
            <span class="quick-label">Quick:</span>
            <button class="quick-btn" @click="transferTarget = '100'">Reception</button>
            <button class="quick-btn" @click="transferTarget = '200'">Sales</button>
            <button class="quick-btn" @click="transferTarget = '*98'">Voicemail</button>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showTransferModal = false">Cancel</button>
          <button class="btn-primary" @click="executeTransfer" :disabled="!transferTarget">Transfer</button>
        </div>
      </div>
    </div>

    <!-- Device Menu Backdrop -->
    <div class="menu-backdrop" v-if="showDeviceMenu" @click="showDeviceMenu = false"></div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { 
  Phone as PhoneIcon, PhoneOff as PhoneOffIcon, PhoneCall as PhoneCallIcon,
  PhoneMissed as PhoneMissedIcon, PhoneOutgoing as PhoneOutgoingIcon, PhoneIncoming as PhoneIncomingIcon,
  Mic as MicIcon, MicOff as MicOffIcon, Pause as PauseIcon,
  ArrowRightLeft as ArrowRightLeftIcon, Grid3x3 as GridIcon,
  ChevronDown as ChevronDownIcon, Monitor as MonitorIcon, Smartphone as SmartphoneIcon,
  Headphones as HeadphonesIcon, X as XIcon, Info as InfoIcon
} from 'lucide-vue-next'

// ============================================
// DEVICE BINDING
// ============================================

const showDeviceMenu = ref(false)
const boundDevice = ref(null)
const deviceHasActiveCall = ref(false)

const userDevices = ref([
  { id: 'softphone', type: 'softphone', name: 'Browser Softphone', meta: 'WebRTC', status: 'online', icon: HeadphonesIcon },
  { id: 'yealink-desk', type: 'desk_phone', name: 'Yealink T46U (Desk)', meta: 'MAC: 80:5E:C0:12:34:56', status: 'online', icon: MonitorIcon },
  { id: 'poly-conf', type: 'desk_phone', name: 'Polycom VVX 450', meta: 'Conference Room B', status: 'online', icon: MonitorIcon },
  { id: 'mobile-app', type: 'mobile', name: 'Mobile App', meta: 'iPhone 15 Pro', status: 'online', icon: SmartphoneIcon },
  { id: 'home-phone', type: 'desk_phone', name: 'Home Office Phone', meta: 'Grandstream', status: 'offline', icon: MonitorIcon },
])

const getDeviceIcon = (device) => {
  if (!device) return MonitorIcon
  return device.icon || MonitorIcon
}

const bindDevice = (device) => {
  if (device.status === 'offline') return
  boundDevice.value = device
  showDeviceMenu.value = false
  console.log('[Dialer] Bound to device:', device.name)
  
  // If binding to physical device, subscribe to its call events
  if (device.type !== 'softphone') {
    subscribeToDeviceEvents(device)
  }
}

const subscribeToDeviceEvents = (device) => {
  // This would connect to a WebSocket/API to get real-time device state
  console.log('[Dialer] Subscribing to events for:', device.id)
  // Mock: simulate device has active call after 5s
  // In real implementation: WebSocket subscription to FreeSWITCH ESL or API
}

const takeCallControl = () => {
  console.log('[Dialer] Taking control of call on:', boundDevice.value?.name)
  // This would sync the call state from the device to the web UI
  callState.value = 'established'
  remoteNumber.value = '(415) 555-1234'
  remoteName.value = 'External Caller'
  callDirection.value = 'inbound'
  startTime.value = new Date(Date.now() - 45000) // Call started 45s ago
  deviceHasActiveCall.value = false
}

// ============================================
// SIP STATE (Groundwork for SIP.js)
// ============================================

const sipState = ref('disconnected') // disconnected, connecting, connected, registered
const sipStateLabel = computed(() => {
  const labels = {
    disconnected: 'Disconnected',
    connecting: 'Connecting...',
    connected: 'Connected',
    registered: 'Ready'
  }
  return labels[sipState.value] || 'Unknown'
})

const canMakeCalls = computed(() => {
  // For now, always allow (mock). With SIP.js, check registration state
  return sipState.value === 'registered' || sipState.value === 'disconnected' // Allow in mock mode
})

// SIP.js integration points (to be implemented)
const initializeSip = async () => {
  console.log('[SIP] Initializing...')
  sipState.value = 'connecting'
  
  // ===== SIP.js Integration Point =====
  // const { UserAgent, Registerer } = await import('sip.js')
  // userAgent = new UserAgent({ ... })
  // await userAgent.start()
  // const registerer = new Registerer(userAgent)
  // await registerer.register()
  // =====================================
  
  // Mock successful connection
  await new Promise(r => setTimeout(r, 500))
  sipState.value = 'registered'
}

// ============================================
// CALL STATE
// ============================================

const number = ref('')
const callState = ref('idle') // idle, dialing, ringing, established, holding, terminated
const callDirection = ref(null) // inbound, outbound
const remoteNumber = ref('')
const remoteName = ref('')
const startTime = ref(null)
const duration = ref(0)
const isMuted = ref(false)
const isOnHold = ref(false)

let durationTimer = null

const callStatusText = computed(() => {
  const statusMap = {
    idle: '',
    dialing: 'Dialing...',
    ringing: callDirection.value === 'inbound' ? 'Incoming Call' : 'Ringing...',
    established: isOnHold.value ? 'On Hold' : 'Connected',
    holding: 'On Hold',
    terminated: 'Call Ended'
  }
  return statusMap[callState.value] || ''
})

const formattedDuration = computed(() => {
  const mins = Math.floor(duration.value / 60)
  const secs = duration.value % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
})

// ============================================
// KEYPAD
// ============================================

const keys = [
  [{ digit: '1', letters: '' }, { digit: '2', letters: 'ABC' }, { digit: '3', letters: 'DEF' }],
  [{ digit: '4', letters: 'GHI' }, { digit: '5', letters: 'JKL' }, { digit: '6', letters: 'MNO' }],
  [{ digit: '7', letters: 'PQRS' }, { digit: '8', letters: 'TUV' }, { digit: '9', letters: 'WXYZ' }],
  [{ digit: '*', letters: '' }, { digit: '0', letters: '+' }, { digit: '#', letters: '' }]
]

const showKeypad = ref(false)

const pressKey = (key) => {
  if (callState.value === 'established') {
    sendDtmf(key)
  } else {
    number.value += key
  }
}

const sendDtmf = (tone) => {
  console.log('[Dialer] DTMF:', tone)
  // SIP.js: session.sessionDescriptionHandler.sendDtmf(tone)
}

// ============================================
// CALL ACTIONS
// ============================================

const makeCall = async () => {
  if (!number.value || callState.value !== 'idle') return
  
  console.log('[Dialer] Calling:', number.value)
  callState.value = 'dialing'
  callDirection.value = 'outbound'
  remoteNumber.value = number.value

  // If bound to physical device, trigger click-to-call via API
  if (boundDevice.value?.type !== 'softphone') {
    console.log('[Dialer] Click-to-call via:', boundDevice.value?.name)
    // API: POST /api/click-to-call { extension, destination, device_id }
  }

  // ===== SIP.js Integration Point =====
  // const { Inviter } = await import('sip.js')
  // session = new Inviter(userAgent, targetUri)
  // await session.invite()
  // =====================================

  // Mock call progress
  await new Promise(r => setTimeout(r, 1000))
  callState.value = 'ringing'
  await new Promise(r => setTimeout(r, 2000))
  callState.value = 'established'
  startTime.value = new Date()
  startDurationTimer()
}

const answerCall = async () => {
  if (callState.value !== 'ringing') return
  console.log('[Dialer] Answering call')
  
  // ===== SIP.js Integration Point =====
  // await session.accept()
  // =====================================
  
  callState.value = 'established'
  startTime.value = new Date()
  startDurationTimer()
}

const hangupCall = async () => {
  console.log('[Dialer] Hanging up')
  
  // ===== SIP.js Integration Point =====
  // session.bye()
  // =====================================
  
  stopDurationTimer()
  callState.value = 'terminated'
  
  setTimeout(() => {
    resetCall()
  }, 1500)
}

const toggleMute = () => {
  isMuted.value = !isMuted.value
  console.log('[Dialer] Mute:', isMuted.value)
  // SIP.js: toggle audio track enabled state
}

const toggleHold = () => {
  isOnHold.value = !isOnHold.value
  callState.value = isOnHold.value ? 'holding' : 'established'
  console.log('[Dialer] Hold:', isOnHold.value)
  // SIP.js: session.hold() / session.unhold()
}

const startDurationTimer = () => {
  duration.value = 0
  durationTimer = setInterval(() => {
    if (startTime.value) {
      duration.value = Math.floor((Date.now() - startTime.value.getTime()) / 1000)
    }
  }, 1000)
}

const stopDurationTimer = () => {
  if (durationTimer) {
    clearInterval(durationTimer)
    durationTimer = null
  }
}

const resetCall = () => {
  callState.value = 'idle'
  callDirection.value = null
  remoteNumber.value = ''
  remoteName.value = ''
  startTime.value = null
  duration.value = 0
  isMuted.value = false
  isOnHold.value = false
  number.value = ''
}

// ============================================
// TRANSFER
// ============================================

const showTransferModal = ref(false)
const transferType = ref('blind')
const transferTarget = ref('')

const executeTransfer = () => {
  console.log('[Dialer] Transfer:', transferType.value, 'to', transferTarget.value)
  // SIP.js: session.refer(targetUri)
  showTransferModal.value = false
  hangupCall()
}

// ============================================
// RECENT CALLS
// ============================================

const recentCalls = ref([
  { id: 1, number: '(415) 555-1234', name: null, type: 'missed', duration: null, time: '2m ago' },
  { id: 2, number: '(415) 555-5678', name: 'Alice Smith', type: 'outgoing', duration: '3:42', time: '1h ago' },
  { id: 3, number: '(212) 555-9999', name: 'Support Queue', type: 'incoming', duration: '12:10', time: 'Yesterday' },
  { id: 4, number: '101', name: 'John Doe', type: 'outgoing', duration: '0:45', time: 'Yesterday' },
])

const dialRecent = (call) => {
  number.value = call.number
}

const dialNumber = (num) => {
  number.value = num
  makeCall()
}

// ============================================
// LIFECYCLE
// ============================================

onMounted(() => {
  // Bind to softphone by default
  boundDevice.value = userDevices.value.find(d => d.type === 'softphone')
  initializeSip()
})

onUnmounted(() => {
  stopDurationTimer()
})
</script>

<style scoped>
.softphone-container { display: flex; flex-direction: column; height: 100%; }

/* Device Bar */
.device-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 16px;
  border-radius: var(--radius-md);
}

.device-selector { position: relative; }
.selector-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; font-weight: 600; margin-right: 12px; }

.device-dropdown {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
}
.device-dropdown:hover { border-color: var(--primary-color); }

.current-device { display: flex; align-items: center; gap: 8px; }
.device-icon { width: 18px; height: 18px; color: var(--text-muted); }
.device-name { font-size: 13px; font-weight: 600; }
.device-status { font-size: 10px; padding: 2px 6px; border-radius: 4px; text-transform: uppercase; font-weight: 600; }
.device-status.online { background: #dcfce7; color: #16a34a; }
.device-status.offline { background: #fee2e2; color: #dc2626; }
.chevron { width: 16px; height: 16px; color: var(--text-muted); }

.device-menu {
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: 4px;
  width: 280px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.12);
  z-index: 100;
  overflow: hidden;
}

.menu-header { padding: 12px 16px; font-size: 11px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; background: var(--bg-app); }

.device-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
}
.device-option:hover { background: var(--bg-app); }
.device-option.active { background: var(--primary-light); }
.device-option.disabled { opacity: 0.5; cursor: not-allowed; }

.option-icon { width: 20px; height: 20px; color: var(--text-muted); }
.option-info { flex: 1; display: flex; flex-direction: column; }
.option-name { font-size: 13px; font-weight: 600; }
.option-meta { font-size: 11px; color: var(--text-muted); }
.option-status { width: 8px; height: 8px; border-radius: 50%; }
.option-status.online { background: #22c55e; }
.option-status.offline { background: #d1d5db; }

.menu-backdrop { position: fixed; inset: 0; z-index: 99; }

/* Active Call Indicator */
.active-call-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  background: #dcfce7;
  border: 1px solid #22c55e;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #16a34a;
}
.call-icon { width: 16px; height: 16px; }
.pulse-ring { width: 8px; height: 8px; background: #22c55e; border-radius: 50%; animation: pulse 1.5s infinite; }
.btn-take-control {
  margin-left: auto;
  padding: 4px 12px;
  background: #22c55e;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

/* Dialer Main */
.dialer-main { display: flex; gap: 24px; flex: 1; max-width: 900px; margin: 0 auto; width: 100%; }

/* Dialer Panel */
.dialer-panel {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
  width: 340px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 16px;
  padding: 4px 12px;
  border-radius: 20px;
  background: var(--bg-app);
  color: var(--text-muted);
}
.connection-status.registered { background: #dcfce7; color: #16a34a; }
.connection-status.connecting { background: #fef3c7; color: #b45309; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }

.display { width: 100%; text-align: center; margin-bottom: 20px; min-height: 80px; }

.call-info { display: flex; flex-direction: column; align-items: center; gap: 4px; }
.call-direction { font-size: 10px; text-transform: uppercase; font-weight: 700; padding: 2px 8px; border-radius: 4px; }
.call-direction.inbound { background: #dcfce7; color: #16a34a; }
.call-direction.outbound { background: #dbeafe; color: #2563eb; }
.remote-name { font-size: 18px; font-weight: 700; color: var(--text-primary); }
.remote-number { font-size: 14px; color: var(--text-muted); font-family: monospace; }
.call-status { font-size: 12px; color: var(--text-muted); margin-top: 4px; }
.call-timer { font-size: 24px; font-weight: 700; color: var(--primary-color); font-family: monospace; margin-top: 8px; }

.number-display {
  width: 100%;
  border: none;
  font-size: 24px;
  font-weight: 600;
  text-align: center;
  outline: none;
  font-family: monospace;
}

/* Keypad */
.keypad { display: flex; flex-direction: column; gap: 12px; margin-bottom: 24px; }
.key-row { display: flex; gap: 16px; }

.key-btn {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: white;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}
.key-btn:hover { background: var(--bg-app); border-color: var(--primary-color); transform: scale(1.05); }
.key-btn:active { transform: scale(0.95); }

.digit { font-size: 24px; font-weight: 500; color: var(--text-primary); line-height: 1; }
.letters { font-size: 9px; color: var(--text-muted); letter-spacing: 1px; margin-top: 2px; }

/* Call Controls */
.call-controls { display: flex; gap: 8px; margin-bottom: 20px; flex-wrap: wrap; justify-content: center; }

.control-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 14px;
  background: var(--bg-app);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  font-size: 10px;
  font-weight: 600;
  color: var(--text-main);
}
.control-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.control-btn.active { background: var(--primary-light); border-color: var(--primary-color); color: var(--primary-color); }
.control-icon { width: 18px; height: 18px; }

/* Action Buttons */
.actions { margin-bottom: 16px; }

.call-btn {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  border: none;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  transition: all 0.15s ease;
}
.call-btn:hover { transform: scale(1.08); }
.call-btn:active { transform: scale(0.95); }
.call-btn:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

.call-btn.dial { background: linear-gradient(135deg, #22c55e, #16a34a); }
.call-btn.answer { background: linear-gradient(135deg, #22c55e, #16a34a); animation: pulse-ring 1.5s infinite; }
.call-btn.hangup { background: linear-gradient(135deg, #ef4444, #dc2626); }

@keyframes pulse-ring {
  0% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.5); }
  70% { box-shadow: 0 0 0 15px rgba(34, 197, 94, 0); }
  100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
}

.bound-device-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
  padding: 8px 12px;
  background: var(--bg-app);
  border-radius: 6px;
}
.info-icon { width: 14px; height: 14px; }

/* Recent Calls */
.recent-calls {
  flex: 1;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
}

.recent-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.recent-header h3 { font-size: 16px; font-weight: 600; margin: 0; }

.call-list { display: flex; flex-direction: column; gap: 8px; }

.call-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border-radius: 8px;
  cursor: pointer;
}
.call-item:hover { background: var(--bg-app); }

.call-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.call-icon.missed { background: #fee2e2; color: #ef4444; }
.call-icon.outgoing { background: #dbeafe; color: #2563eb; }
.call-icon.incoming { background: #dcfce7; color: #22c55e; }

.call-details { flex: 1; display: flex; flex-direction: column; }
.caller { font-weight: 600; font-size: 13px; color: var(--text-primary); }
.meta { font-size: 11px; color: var(--text-muted); }

.call-back-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-muted);
  opacity: 0;
  transition: all 0.15s;
}
.call-item:hover .call-back-btn { opacity: 1; }
.call-back-btn:hover { background: var(--primary-color); color: white; border-color: var(--primary-color); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 400px; }
.modal-card.small { max-width: 360px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.transfer-tabs { display: flex; gap: 4px; margin-bottom: 16px; background: var(--bg-app); padding: 4px; border-radius: 8px; }
.transfer-tabs button { flex: 1; padding: 8px; border: none; background: transparent; border-radius: 6px; font-size: 12px; font-weight: 600; cursor: pointer; }
.transfer-tabs button.active { background: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }

.quick-transfer { display: flex; align-items: center; gap: 8px; margin-top: 12px; flex-wrap: wrap; }
.quick-label { font-size: 11px; color: var(--text-muted); }
.quick-btn { padding: 4px 10px; border: 1px solid var(--border-color); background: white; border-radius: 4px; font-size: 11px; cursor: pointer; }
.quick-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 500; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }

.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }

.icon-lg { width: 28px; height: 28px; }
.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 14px; height: 14px; }

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.5); opacity: 0.5; }
}
</style>
