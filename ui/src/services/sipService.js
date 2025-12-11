/**
 * SIP.js Service for WebRTC calling
 * This provides the groundwork for integrating SIP.js with the Callsign dialer
 * 
 * Installation: npm install sip.js
 * Docs: https://sipjs.com/
 */

import { ref, reactive, computed } from 'vue'

// SIP Connection States
export const SipState = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  REGISTERED: 'registered',
  FAILED: 'failed'
}

// Call States
export const CallState = {
  IDLE: 'idle',
  DIALING: 'dialing',
  RINGING: 'ringing',
  EARLY_MEDIA: 'early_media',
  ESTABLISHED: 'established',
  HOLDING: 'holding',
  HELD: 'held',
  TERMINATED: 'terminated'
}

// Audio Mode - determines which device handles audio
export const AudioMode = {
  SOFTPHONE: 'softphone',       // Browser WebRTC
  DESK_PHONE: 'desk_phone',     // Bound to physical desk phone
  MOBILE: 'mobile',              // Mobile device
  CALL_ME: 'call_me'             // Click-to-call - system calls you
}

// Create a reactive SIP service instance
export function useSipService() {
  // Connection state
  const state = reactive({
    connectionState: SipState.DISCONNECTED,
    callState: CallState.IDLE,
    audioMode: AudioMode.SOFTPHONE,
    boundDevice: null,
    registeredDevices: [],
    error: null
  })

  // Current call info
  const currentCall = reactive({
    id: null,
    remoteNumber: '',
    remoteName: '',
    direction: null, // 'inbound' | 'outbound'
    startTime: null,
    duration: 0,
    muted: false,
    onHold: false
  })

  // Configuration (would come from server/user settings)
  const config = reactive({
    server: '',      // e.g., 'wss://sip.callsign.io:7443/ws'
    domain: '',      // e.g., 'sip.callsign.io'
    extension: '',
    password: '',
    displayName: '',
    stunServers: ['stun:stun.l.google.com:19302'],
    turnServers: []
  })

  // Audio elements for media handling
  let localAudio = null
  let remoteAudio = null
  let userAgent = null
  let session = null
  let durationInterval = null

  // ============================================
  // INITIALIZATION
  // ============================================

  /**
   * Initialize the SIP service with user credentials
   */
  async function initialize(options) {
    console.log('[SIP] Initializing SIP service...', options)
    
    config.server = options.server || config.server
    config.domain = options.domain || config.domain
    config.extension = options.extension || config.extension
    config.password = options.password || config.password
    config.displayName = options.displayName || config.displayName

    // Create audio elements for media
    if (state.audioMode === AudioMode.SOFTPHONE) {
      setupAudioElements()
    }

    // Fetch available devices for binding
    await fetchRegisteredDevices()
    
    return true
  }

  /**
   * Create audio elements for WebRTC media
   */
  function setupAudioElements() {
    if (!localAudio) {
      localAudio = new Audio()
      localAudio.autoplay = true
      localAudio.muted = true // Local audio is muted (you don't want to hear yourself)
    }
    if (!remoteAudio) {
      remoteAudio = new Audio()
      remoteAudio.autoplay = true
    }
  }

  /**
   * Fetch devices registered to this extension for binding
   */
  async function fetchRegisteredDevices() {
    // This would be an API call to get registered devices
    // For now, mock data
    state.registeredDevices = [
      { id: 'softphone', type: 'softphone', name: 'Browser Softphone', status: 'available', mac: null },
      { id: 'yealink-t46u', type: 'desk_phone', name: 'Yealink T46U (Desk)', status: 'registered', mac: '80:5E:C0:12:34:56' },
      { id: 'poly-vvx450', type: 'desk_phone', name: 'Polycom VVX 450 (Conf Room)', status: 'registered', mac: '64:16:7F:78:90:AB' },
      { id: 'mobile-app', type: 'mobile', name: 'Mobile App (iPhone)', status: 'available', mac: null },
    ]
  }

  // ============================================
  // CONNECTION MANAGEMENT
  // ============================================

  /**
   * Connect to SIP server and register
   */
  async function connect() {
    if (state.connectionState === SipState.CONNECTED || state.connectionState === SipState.REGISTERED) {
      console.log('[SIP] Already connected')
      return true
    }

    state.connectionState = SipState.CONNECTING
    state.error = null

    try {
      console.log('[SIP] Connecting to', config.server)
      
      // ===== SIP.js Integration Point =====
      // When sip.js is installed, uncomment this:
      /*
      const { UserAgent, Registerer, Inviter, SessionState } = await import('sip.js')
      
      const uri = UserAgent.makeURI(`sip:${config.extension}@${config.domain}`)
      
      userAgent = new UserAgent({
        uri: uri,
        authorizationUsername: config.extension,
        authorizationPassword: config.password,
        transportOptions: {
          server: config.server
        },
        sessionDescriptionHandlerFactoryOptions: {
          peerConnectionConfiguration: {
            iceServers: [
              { urls: config.stunServers },
              ...config.turnServers
            ]
          }
        },
        displayName: config.displayName
      })

      userAgent.delegate = {
        onInvite: handleIncomingCall
      }

      await userAgent.start()
      
      const registerer = new Registerer(userAgent)
      await registerer.register()
      */
      // =====================================

      // Mock successful connection for now
      await new Promise(r => setTimeout(r, 500))
      state.connectionState = SipState.REGISTERED
      console.log('[SIP] Connected and registered')
      return true

    } catch (error) {
      console.error('[SIP] Connection failed:', error)
      state.connectionState = SipState.FAILED
      state.error = error.message
      return false
    }
  }

  /**
   * Disconnect from SIP server
   */
  async function disconnect() {
    if (session) {
      await hangup()
    }
    
    if (userAgent) {
      await userAgent.stop()
      userAgent = null
    }
    
    state.connectionState = SipState.DISCONNECTED
    console.log('[SIP] Disconnected')
  }

  // ============================================
  // DEVICE BINDING
  // ============================================

  /**
   * Bind to a specific device for audio
   */
  function bindToDevice(device) {
    console.log('[SIP] Binding to device:', device)
    
    state.boundDevice = device
    state.audioMode = device.type === 'softphone' ? AudioMode.SOFTPHONE :
                      device.type === 'desk_phone' ? AudioMode.DESK_PHONE :
                      device.type === 'mobile' ? AudioMode.MOBILE : AudioMode.SOFTPHONE
    
    // If binding to physical device, we need to tell server to route audio there
    if (device.type !== 'softphone') {
      console.log('[SIP] Audio will be routed to:', device.name)
      // API call to set device binding would go here
    }
    
    return true
  }

  /**
   * Unbind from device (return to softphone)
   */
  function unbindDevice() {
    const softphone = state.registeredDevices.find(d => d.type === 'softphone')
    if (softphone) {
      bindToDevice(softphone)
    }
  }

  // ============================================
  // CALL MANAGEMENT
  // ============================================

  /**
   * Make an outbound call
   */
  async function call(number) {
    if (state.callState !== CallState.IDLE) {
      console.warn('[SIP] Already on a call')
      return false
    }

    console.log('[SIP] Calling:', number)
    state.callState = CallState.DIALING
    
    currentCall.id = Date.now().toString()
    currentCall.remoteNumber = number
    currentCall.direction = 'outbound'
    currentCall.startTime = null
    currentCall.duration = 0

    try {
      // If using bound device, initiate call via server
      if (state.audioMode !== AudioMode.SOFTPHONE) {
        console.log('[SIP] Initiating call through bound device:', state.boundDevice?.name)
        // Server API would handle this - ring the bound device first, then connect to destination
        // POST /api/click-to-call { extension, destination, device }
      }

      // ===== SIP.js Integration Point =====
      /*
      const { Inviter } = await import('sip.js')
      const target = UserAgent.makeURI(`sip:${number}@${config.domain}`)
      
      session = new Inviter(userAgent, target, {
        sessionDescriptionHandlerOptions: {
          constraints: { audio: true, video: false }
        }
      })
      
      session.stateChange.addListener((newState) => {
        handleSessionStateChange(newState)
      })
      
      await session.invite({
        requestDelegate: {
          onProgress: () => {
            state.callState = CallState.RINGING
          }
        }
      })
      */
      // =====================================

      // Mock call progress
      await new Promise(r => setTimeout(r, 1000))
      state.callState = CallState.RINGING
      
      await new Promise(r => setTimeout(r, 2000))
      state.callState = CallState.ESTABLISHED
      currentCall.startTime = new Date()
      startDurationTimer()
      
      return true

    } catch (error) {
      console.error('[SIP] Call failed:', error)
      state.callState = CallState.IDLE
      state.error = error.message
      return false
    }
  }

  /**
   * Answer an incoming call
   */
  async function answer() {
    if (state.callState !== CallState.RINGING || currentCall.direction !== 'inbound') {
      return false
    }

    console.log('[SIP] Answering call')
    
    // ===== SIP.js Integration Point =====
    /*
    if (session) {
      await session.accept({
        sessionDescriptionHandlerOptions: {
          constraints: { audio: true, video: false }
        }
      })
    }
    */
    // =====================================

    state.callState = CallState.ESTABLISHED
    currentCall.startTime = new Date()
    startDurationTimer()
    
    return true
  }

  /**
   * Hang up current call
   */
  async function hangup() {
    if (state.callState === CallState.IDLE) return false
    
    console.log('[SIP] Hanging up')
    
    // ===== SIP.js Integration Point =====
    /*
    if (session) {
      session.bye()
      session = null
    }
    */
    // =====================================

    stopDurationTimer()
    state.callState = CallState.TERMINATED
    
    setTimeout(() => {
      state.callState = CallState.IDLE
      resetCallInfo()
    }, 1000)
    
    return true
  }

  /**
   * Toggle mute
   */
  function toggleMute() {
    currentCall.muted = !currentCall.muted
    console.log('[SIP] Mute:', currentCall.muted)
    
    // ===== SIP.js Integration Point =====
    /*
    if (session) {
      const pc = session.sessionDescriptionHandler.peerConnection
      pc.getSenders().forEach(sender => {
        if (sender.track?.kind === 'audio') {
          sender.track.enabled = !currentCall.muted
        }
      })
    }
    */
    // =====================================
    
    return currentCall.muted
  }

  /**
   * Toggle hold
   */
  async function toggleHold() {
    if (state.callState !== CallState.ESTABLISHED && state.callState !== CallState.HOLDING) {
      return false
    }

    // ===== SIP.js Integration Point =====
    /*
    if (session) {
      if (currentCall.onHold) {
        await session.unhold()
      } else {
        await session.hold()
      }
    }
    */
    // =====================================

    currentCall.onHold = !currentCall.onHold
    state.callState = currentCall.onHold ? CallState.HOLDING : CallState.ESTABLISHED
    console.log('[SIP] Hold:', currentCall.onHold)
    
    return currentCall.onHold
  }

  /**
   * Send DTMF tone
   */
  function sendDtmf(tone) {
    if (state.callState !== CallState.ESTABLISHED) return false
    
    console.log('[SIP] DTMF:', tone)
    
    // ===== SIP.js Integration Point =====
    /*
    if (session) {
      session.sessionDescriptionHandler.sendDtmf(tone)
    }
    */
    // =====================================
    
    return true
  }

  /**
   * Transfer call (blind transfer)
   */
  async function transfer(targetNumber) {
    if (state.callState !== CallState.ESTABLISHED) return false
    
    console.log('[SIP] Transferring to:', targetNumber)
    
    // ===== SIP.js Integration Point =====
    /*
    if (session) {
      const target = UserAgent.makeURI(`sip:${targetNumber}@${config.domain}`)
      await session.refer(target)
    }
    */
    // =====================================
    
    return true
  }

  // ============================================
  // HELPER FUNCTIONS
  // ============================================

  function handleIncomingCall(invitation) {
    console.log('[SIP] Incoming call')
    session = invitation
    state.callState = CallState.RINGING
    
    currentCall.id = Date.now().toString()
    currentCall.remoteNumber = invitation.remoteIdentity?.uri?.user || 'Unknown'
    currentCall.remoteName = invitation.remoteIdentity?.displayName || ''
    currentCall.direction = 'inbound'
  }

  function handleSessionStateChange(newState) {
    console.log('[SIP] Session state:', newState)
    // Map SIP.js session states to our states
  }

  function startDurationTimer() {
    durationInterval = setInterval(() => {
      if (currentCall.startTime) {
        currentCall.duration = Math.floor((Date.now() - currentCall.startTime.getTime()) / 1000)
      }
    }, 1000)
  }

  function stopDurationTimer() {
    if (durationInterval) {
      clearInterval(durationInterval)
      durationInterval = null
    }
  }

  function resetCallInfo() {
    currentCall.id = null
    currentCall.remoteNumber = ''
    currentCall.remoteName = ''
    currentCall.direction = null
    currentCall.startTime = null
    currentCall.duration = 0
    currentCall.muted = false
    currentCall.onHold = false
  }

  function formatDuration(seconds) {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  // ============================================
  // COMPUTED / GETTERS
  // ============================================

  const isConnected = computed(() => 
    state.connectionState === SipState.CONNECTED || 
    state.connectionState === SipState.REGISTERED
  )

  const isOnCall = computed(() => 
    state.callState !== CallState.IDLE && 
    state.callState !== CallState.TERMINATED
  )

  const canCall = computed(() => 
    isConnected.value && 
    state.callState === CallState.IDLE
  )

  const formattedDuration = computed(() => formatDuration(currentCall.duration))

  // ============================================
  // EXPORT
  // ============================================

  return {
    // State
    state,
    currentCall,
    config,
    
    // Computed
    isConnected,
    isOnCall,
    canCall,
    formattedDuration,
    
    // Methods
    initialize,
    connect,
    disconnect,
    bindToDevice,
    unbindDevice,
    call,
    answer,
    hangup,
    toggleMute,
    toggleHold,
    sendDtmf,
    transfer,
    formatDuration
  }
}

// Singleton instance for global use
let sipServiceInstance = null

export function getSipService() {
  if (!sipServiceInstance) {
    sipServiceInstance = useSipService()
  }
  return sipServiceInstance
}

export default {
  useSipService,
  getSipService,
  SipState,
  CallState,
  AudioMode
}
