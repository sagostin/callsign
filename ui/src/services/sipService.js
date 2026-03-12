/**
 * SIP.js Service for WebRTC calling
 * Integrates SIP.js with the Callsign dialer via FreeSWITCH's wss-binding.
 *
 * Architecture:
 *   Browser (SIP.js) → wss:// → FreeSWITCH mod_sofia → PBX dialplan
 *
 * The browser provisions itself as a web_client endpoint via /api/registrations/provision,
 * receives SIP credentials, then connects directly to FreeSWITCH over WebSocket.
 */

import { ref, reactive, computed } from 'vue'
import { UserAgent, Registerer, Inviter, SessionState, RegistererState } from 'sip.js'
import { extensionPortalAPI } from './api'

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

  // Configuration (populated from server + provisioning)
  const config = reactive({
    server: '',      // e.g., 'wss://sip.callsign.io:7443'
    domain: '',      // e.g., 'sip.callsign.io'
    extension: '',
    password: '',
    displayName: '',
    authToken: '',   // JWT for API calls
    userId: null,    // User ID for provisioning
    extensionId: null, // Extension ID for provisioning
    stunServers: ['stun:stun.l.google.com:19302'],
    turnServers: []
  })

  // SIP.js objects (not reactive — these are complex objects)
  let remoteAudio = null
  let ringtoneAudio = null
  let userAgent = null
  let registerer = null
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
    config.authToken = options.authToken || config.authToken
    config.userId = options.userId || config.userId
    config.extensionId = options.extensionId || config.extensionId
    if (options.stunServers) config.stunServers = options.stunServers
    if (options.turnServers) config.turnServers = options.turnServers

    // Create audio elements for media
    if (state.audioMode === AudioMode.SOFTPHONE) {
      setupAudioElements()
    }

    // Fetch available devices for binding
    await fetchRegisteredDevices()

    return true
  }

  /**
   * Fetch WebRTC configuration from the backend
   */
  async function fetchWebRTCConfig() {
    try {
      const res = await extensionPortalAPI.getWebRTCConfig()
      const data = res.data || res.data?.data || {}

      if (data.wss_url) config.server = data.wss_url
      if (data.sip_domain) config.domain = data.sip_domain
      if (data.stun_servers?.length) config.stunServers = data.stun_servers
      if (data.turn_servers?.length) config.turnServers = data.turn_servers

      console.log('[SIP] WebRTC config loaded:', { server: config.server, domain: config.domain })
      return data.enabled !== false
    } catch (error) {
      console.warn('[SIP] Failed to fetch WebRTC config:', error)
      return false
    }
  }

  /**
   * Create audio elements for WebRTC media
   */
  function setupAudioElements() {
    if (!remoteAudio) {
      remoteAudio = new Audio()
      remoteAudio.autoplay = true
    }
    if (!ringtoneAudio) {
      ringtoneAudio = new Audio()
      ringtoneAudio.loop = true
    }
  }

  /**
   * Set the remote audio element (from a <audio ref> in the component)
   */
  function setRemoteAudioElement(el) {
    if (el) {
      remoteAudio = el
      remoteAudio.autoplay = true
    }
  }

  /**
   * Fetch devices/endpoints registered to this extension from the API
   */
  async function fetchRegisteredDevices() {
    try {
      const extensionId = config.extensionId
      if (!extensionId) {
        state.registeredDevices = [
          { id: 'softphone', type: 'softphone', name: 'Browser Softphone', status: 'available', mac: null }
        ]
        return
      }

      const response = await fetch(`/api/registrations/extension/${extensionId}`, {
        headers: {
          'Authorization': `Bearer ${config.authToken}`,
          'Content-Type': 'application/json'
        }
      })

      if (response.ok) {
        const data = await response.json()
        const devices = [
          { id: 'softphone', type: 'softphone', name: 'Browser Softphone', status: 'available', mac: null }
        ]
        if (data.registrations) {
          for (const reg of data.registrations) {
            devices.push({
              id: reg.uuid,
              type: reg.endpoint_type === 'device' ? 'desk_phone' : reg.endpoint_type,
              name: reg.device_label || reg.registration_user,
              status: reg.status,
              mac: reg.device?.mac || null,
              registrationUser: reg.registration_user,
              endpointType: reg.endpoint_type
            })
          }
        }
        state.registeredDevices = devices
      } else {
        console.warn('[SIP] Failed to fetch registered devices, using defaults')
        state.registeredDevices = [
          { id: 'softphone', type: 'softphone', name: 'Browser Softphone', status: 'available', mac: null }
        ]
      }
    } catch (error) {
      console.warn('[SIP] Error fetching devices:', error)
      state.registeredDevices = [
        { id: 'softphone', type: 'softphone', name: 'Browser Softphone', status: 'available', mac: null }
      ]
    }
  }

  /**
   * Get or create a persistent instance ID for this browser
   */
  function getInstanceId() {
    let instanceId = localStorage.getItem('callsign_sip_instance_id')
    if (!instanceId) {
      instanceId = crypto.randomUUID ? crypto.randomUUID().substring(0, 8) : Math.random().toString(36).substring(2, 10)
      localStorage.setItem('callsign_sip_instance_id', instanceId)
    }
    return instanceId
  }

  /**
   * Provision this browser as a WebRTC client endpoint
   * Returns SIP credentials for independent registration with FreeSWITCH
   */
  async function provisionWebClient() {
    const instanceId = getInstanceId()

    try {
      const response = await fetch('/api/registrations/provision', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${config.authToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          endpoint_type: 'web_client',
          instance_id: instanceId,
          device_label: `Browser (${navigator.platform || 'Web'})`,
          os_info: navigator.userAgent.substring(0, 100),
          user_id: config.userId || undefined,
          extension_id: config.extensionId || undefined
        })
      })

      if (response.ok) {
        const data = await response.json()
        console.log('[SIP] Web client provisioned:', data.sip_user)
        return {
          sipUser: data.sip_user,
          sipPassword: data.sip_password,
          alreadyProvisioned: data.already_provisioned || false
        }
      } else {
        const err = await response.json()
        console.error('[SIP] Provision failed:', err.error)
        return null
      }
    } catch (error) {
      console.error('[SIP] Provision error:', error)
      return null
    }
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
      // Fetch WebRTC config from backend if not already set
      if (!config.server || !config.domain) {
        const enabled = await fetchWebRTCConfig()
        if (!enabled || !config.server || !config.domain) {
          console.warn('[SIP] WebRTC not configured on server')
          state.connectionState = SipState.FAILED
          state.error = 'WebRTC not configured. Check SIP_WSS_URL and SIP_DOMAIN in server settings.'
          return false
        }
      }

      console.log('[SIP] Connecting to', config.server)

      // Provision as independent web client if in softphone mode
      let sipUser = config.extension
      let sipPassword = config.password

      if (state.audioMode === AudioMode.SOFTPHONE && config.authToken) {
        const provisioned = await provisionWebClient()
        if (provisioned) {
          sipUser = provisioned.sipUser
          sipPassword = provisioned.sipPassword
          console.log('[SIP] Using provisioned web client identity:', sipUser)
        }
      }

      if (!sipUser || !sipPassword) {
        state.connectionState = SipState.FAILED
        state.error = 'No SIP credentials available'
        return false
      }

      // Build ICE server configuration
      const iceServers = []
      if (config.stunServers?.length) {
        iceServers.push({ urls: config.stunServers })
      }
      if (config.turnServers?.length) {
        for (const turn of config.turnServers) {
          if (typeof turn === 'string') {
            iceServers.push({ urls: turn })
          } else if (turn.urls) {
            iceServers.push(turn)
          }
        }
      }

      // Create SIP.js UserAgent
      const uri = UserAgent.makeURI(`sip:${sipUser}@${config.domain}`)
      if (!uri) {
        throw new Error(`Failed to create SIP URI for ${sipUser}@${config.domain}`)
      }

      userAgent = new UserAgent({
        uri: uri,
        authorizationUsername: sipUser,
        authorizationPassword: sipPassword,
        transportOptions: {
          server: config.server
        },
        sessionDescriptionHandlerFactoryOptions: {
          peerConnectionConfiguration: {
            iceServers: iceServers.length > 0 ? iceServers : undefined
          }
        },
        displayName: config.displayName || sipUser,
        logLevel: 'warn'
      })

      // Handle incoming calls
      userAgent.delegate = {
        onInvite: handleIncomingCall
      }

      // Listen for transport state changes
      userAgent.transport.onConnect = () => {
        console.log('[SIP] Transport connected')
        state.connectionState = SipState.CONNECTED
      }

      userAgent.transport.onDisconnect = (error) => {
        console.log('[SIP] Transport disconnected', error)
        if (state.connectionState !== SipState.CONNECTING) {
          state.connectionState = SipState.DISCONNECTED
        }
        // Attempt reconnect after delay
        if (error) {
          setTimeout(() => {
            if (state.connectionState === SipState.DISCONNECTED) {
              console.log('[SIP] Attempting reconnect...')
              connect()
            }
          }, 5000)
        }
      }

      // Start the UserAgent (establishes WebSocket connection)
      await userAgent.start()

      // Register with FreeSWITCH
      registerer = new Registerer(userAgent, {
        expires: 300,
        extraHeaders: []
      })

      registerer.stateChange.addListener((registererState) => {
        console.log('[SIP] Registerer state:', registererState)
        switch (registererState) {
          case RegistererState.Registered:
            state.connectionState = SipState.REGISTERED
            state.error = null
            break
          case RegistererState.Unregistered:
            if (state.connectionState === SipState.REGISTERED) {
              state.connectionState = SipState.CONNECTED
            }
            break
          case RegistererState.Terminated:
            state.connectionState = SipState.DISCONNECTED
            break
        }
      })

      await registerer.register()

      console.log('[SIP] Connected and registered')
      return true

    } catch (error) {
      console.error('[SIP] Connection failed:', error)
      state.connectionState = SipState.FAILED
      state.error = error.message || 'Connection failed'
      return false
    }
  }

  /**
   * Disconnect from SIP server
   */
  async function disconnect() {
    try {
      if (session) {
        await hangup()
      }

      if (registerer) {
        try {
          await registerer.unregister()
        } catch (e) {
          console.warn('[SIP] Unregister error (may be expected):', e.message)
        }
        registerer = null
      }

      if (userAgent) {
        try {
          await userAgent.stop()
        } catch (e) {
          console.warn('[SIP] UserAgent stop error:', e.message)
        }
        userAgent = null
      }
    } catch (e) {
      console.warn('[SIP] Disconnect error:', e.message)
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

    if (device.type !== 'softphone') {
      console.log('[SIP] Audio will be routed to:', device.name)
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
  // MEDIA STREAM HANDLING
  // ============================================

  /**
   * Extract and attach remote media stream from a session to the audio element
   */
  function attachRemoteMedia(currentSession) {
    if (!currentSession || !remoteAudio) return

    const sdh = currentSession.sessionDescriptionHandler
    if (!sdh || !sdh.peerConnection) {
      console.warn('[SIP] No session description handler or peer connection')
      return
    }

    const pc = sdh.peerConnection
    const receivers = pc.getReceivers()

    if (receivers.length > 0) {
      const remoteStream = new MediaStream()
      receivers.forEach(receiver => {
        if (receiver.track) {
          remoteStream.addTrack(receiver.track)
        }
      })
      remoteAudio.srcObject = remoteStream
      remoteAudio.play().catch(e => console.warn('[SIP] Audio play error:', e.message))
      console.log('[SIP] Remote media attached')
    }

    // Also listen for new tracks (in case they arrive late)
    pc.ontrack = (event) => {
      console.log('[SIP] New track received:', event.track.kind)
      if (remoteAudio.srcObject instanceof MediaStream) {
        remoteAudio.srcObject.addTrack(event.track)
      } else {
        const stream = new MediaStream([event.track])
        remoteAudio.srcObject = stream
        remoteAudio.play().catch(e => console.warn('[SIP] Audio play error:', e.message))
      }
    }
  }

  /**
   * Detach remote media from audio element
   */
  function detachRemoteMedia() {
    if (remoteAudio) {
      remoteAudio.srcObject = null
    }
  }

  // ============================================
  // CALL MANAGEMENT
  // ============================================

  /**
   * Setup session state change listener for a call session
   */
  function setupSessionStateListener(currentSession) {
    currentSession.stateChange.addListener((newState) => {
      console.log('[SIP] Session state:', newState)

      switch (newState) {
        case SessionState.Establishing:
          state.callState = CallState.RINGING
          break

        case SessionState.Established:
          state.callState = CallState.ESTABLISHED
          currentCall.startTime = new Date()
          startDurationTimer()
          attachRemoteMedia(currentSession)
          stopRingtone()
          break

        case SessionState.Terminating:
          // Transitional state — do nothing, wait for Terminated
          break

        case SessionState.Terminated:
          stopDurationTimer()
          detachRemoteMedia()
          stopRingtone()
          state.callState = CallState.TERMINATED
          session = null
          setTimeout(() => {
            state.callState = CallState.IDLE
            resetCallInfo()
          }, 1500)
          break
      }
    })
  }

  /**
   * Make an outbound call
   */
  async function call(number) {
    if (state.callState !== CallState.IDLE) {
      console.warn('[SIP] Already on a call')
      return false
    }

    if (!userAgent) {
      console.error('[SIP] UserAgent not initialized')
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
        // Server API would handle this — no SIP.js involved for hardware devices
        return false
      }

      // Create SIP INVITE via SIP.js
      const target = UserAgent.makeURI(`sip:${number}@${config.domain}`)
      if (!target) {
        throw new Error(`Invalid target URI: ${number}@${config.domain}`)
      }

      session = new Inviter(userAgent, target, {
        sessionDescriptionHandlerOptions: {
          constraints: { audio: true, video: false }
        }
      })

      // Listen for state changes
      setupSessionStateListener(session)

      // Send the INVITE
      await session.invite()

      return true

    } catch (error) {
      console.error('[SIP] Call failed:', error)
      state.callState = CallState.IDLE
      state.error = error.message
      session = null
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

    if (!session) {
      console.error('[SIP] No session to answer')
      return false
    }

    console.log('[SIP] Answering call')

    try {
      await session.accept({
        sessionDescriptionHandlerOptions: {
          constraints: { audio: true, video: false }
        }
      })
      return true
    } catch (error) {
      console.error('[SIP] Answer failed:', error)
      state.error = error.message
      return false
    }
  }

  /**
   * Reject an incoming call
   */
  async function reject() {
    if (state.callState !== CallState.RINGING || currentCall.direction !== 'inbound') {
      return false
    }

    if (!session) return false

    console.log('[SIP] Rejecting call')
    try {
      session.reject()
      stopRingtone()
      session = null
      state.callState = CallState.IDLE
      resetCallInfo()
      return true
    } catch (error) {
      console.error('[SIP] Reject failed:', error)
      return false
    }
  }

  /**
   * Hang up current call
   */
  async function hangup() {
    if (state.callState === CallState.IDLE) return false
    if (!session) {
      // No active session — just reset state
      stopDurationTimer()
      detachRemoteMedia()
      stopRingtone()
      state.callState = CallState.IDLE
      resetCallInfo()
      return true
    }

    console.log('[SIP] Hanging up')

    try {
      switch (session.state) {
        case SessionState.Initial:
        case SessionState.Establishing:
          // Outbound call not yet established
          if (currentCall.direction === 'outbound') {
            session.cancel()
          } else {
            session.reject()
          }
          break

        case SessionState.Established:
          session.bye()
          break

        default:
          // Already terminating/terminated
          break
      }
    } catch (error) {
      console.warn('[SIP] Hangup error:', error.message)
    }

    // State changes will be handled by the session state listener
    // but force cleanup if needed
    stopDurationTimer()
    detachRemoteMedia()
    stopRingtone()

    return true
  }

  /**
   * Toggle mute
   */
  function toggleMute() {
    if (!session) return currentCall.muted

    currentCall.muted = !currentCall.muted
    console.log('[SIP] Mute:', currentCall.muted)

    try {
      const sdh = session.sessionDescriptionHandler
      if (sdh && sdh.peerConnection) {
        const pc = sdh.peerConnection
        pc.getSenders().forEach(sender => {
          if (sender.track && sender.track.kind === 'audio') {
            sender.track.enabled = !currentCall.muted
          }
        })
      }
    } catch (error) {
      console.warn('[SIP] Mute toggle error:', error.message)
    }

    return currentCall.muted
  }

  /**
   * Toggle hold
   */
  async function toggleHold() {
    if (state.callState !== CallState.ESTABLISHED && state.callState !== CallState.HOLDING) {
      return false
    }

    if (!session) return false

    try {
      const sdh = session.sessionDescriptionHandler
      if (!sdh || !sdh.peerConnection) return false

      const pc = sdh.peerConnection

      if (currentCall.onHold) {
        // Unhold: set all transceivers back to sendrecv
        pc.getTransceivers().forEach(transceiver => {
          if (transceiver.direction === 'sendonly' || transceiver.direction === 'inactive') {
            transceiver.direction = 'sendrecv'
          }
        })

        // Send re-INVITE to notify remote end
        const options = {
          sessionDescriptionHandlerModifiers: [
            (description) => {
              // Replace sendonly/inactive with sendrecv in SDP
              description.sdp = description.sdp.replace(/a=sendonly/g, 'a=sendrecv')
              description.sdp = description.sdp.replace(/a=inactive/g, 'a=sendrecv')
              return Promise.resolve(description)
            }
          ]
        }
        await session.invite(options)

        currentCall.onHold = false
        state.callState = CallState.ESTABLISHED
        console.log('[SIP] Call resumed')
      } else {
        // Hold: set all transceivers to sendonly
        pc.getTransceivers().forEach(transceiver => {
          if (transceiver.direction === 'sendrecv') {
            transceiver.direction = 'sendonly'
          }
        })

        // Send re-INVITE to notify remote end
        const options = {
          sessionDescriptionHandlerModifiers: [
            (description) => {
              // Replace sendrecv with sendonly in SDP
              description.sdp = description.sdp.replace(/a=sendrecv/g, 'a=sendonly')
              return Promise.resolve(description)
            }
          ]
        }
        await session.invite(options)

        currentCall.onHold = true
        state.callState = CallState.HOLDING
        console.log('[SIP] Call on hold')
      }
    } catch (error) {
      console.warn('[SIP] Hold toggle error:', error.message)
    }

    return currentCall.onHold
  }

  /**
   * Send DTMF tone
   */
  function sendDtmf(tone) {
    if (state.callState !== CallState.ESTABLISHED) return false
    if (!session) return false

    console.log('[SIP] DTMF:', tone)

    try {
      // Use SIP INFO for DTMF (more reliable than RFC 2833 in WebRTC)
      const body = {
        contentDisposition: 'render',
        contentType: 'application/dtmf-relay',
        content: `Signal=${tone}\r\nDuration=250`
      }
      session.info({ requestOptions: { body } })
      return true
    } catch (error) {
      console.warn('[SIP] DTMF error:', error.message)
      return false
    }
  }

  /**
   * Transfer call (blind transfer via SIP REFER)
   */
  async function transfer(targetNumber) {
    if (state.callState !== CallState.ESTABLISHED) return false
    if (!session) return false

    console.log('[SIP] Transferring to:', targetNumber)

    try {
      const target = UserAgent.makeURI(`sip:${targetNumber}@${config.domain}`)
      if (!target) {
        throw new Error(`Invalid transfer target: ${targetNumber}`)
      }
      await session.refer(target)
      return true
    } catch (error) {
      console.error('[SIP] Transfer failed:', error)
      state.error = error.message
      return false
    }
  }

  // ============================================
  // INCOMING CALL HANDLING
  // ============================================

  function handleIncomingCall(invitation) {
    console.log('[SIP] Incoming call from:', invitation.remoteIdentity?.uri?.user)

    // If already on a call, reject the new one
    if (session && state.callState !== CallState.IDLE) {
      console.log('[SIP] Busy — rejecting incoming call')
      invitation.reject({ statusCode: 486, reasonPhrase: 'Busy Here' })
      return
    }

    session = invitation
    state.callState = CallState.RINGING

    currentCall.id = Date.now().toString()
    currentCall.remoteNumber = invitation.remoteIdentity?.uri?.user || 'Unknown'
    currentCall.remoteName = invitation.remoteIdentity?.displayName || ''
    currentCall.direction = 'inbound'

    // Listen for session state changes
    setupSessionStateListener(invitation)

    // Play ringtone
    playRingtone()
  }

  // ============================================
  // RINGTONE
  // ============================================

  function playRingtone() {
    if (!ringtoneAudio) return
    try {
      // Use a simple oscillator-based ringtone if no audio file is set
      // Alternatively, set ringtoneAudio.src = '/sounds/ringtone.wav'
      ringtoneAudio.src = '' // Will be silent — implement actual ringtone as needed
      // For now, rely on browser notification API instead
    } catch (e) {
      // Ringtone is best-effort
    }
  }

  function stopRingtone() {
    if (!ringtoneAudio) return
    try {
      ringtoneAudio.pause()
      ringtoneAudio.currentTime = 0
    } catch (e) {
      // Best-effort
    }
  }

  // ============================================
  // HELPER FUNCTIONS
  // ============================================

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
    fetchWebRTCConfig,
    connect,
    disconnect,
    bindToDevice,
    unbindDevice,
    provisionWebClient,
    setRemoteAudioElement,
    call,
    answer,
    reject,
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
