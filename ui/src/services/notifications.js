import { ref, reactive } from 'vue'

// Notification types with specific icons and colors
export const NotificationType = {
    SUCCESS: 'success',
    ERROR: 'error',
    WARNING: 'warning',
    INFO: 'info',
    CALL: 'call',        // Incoming/missed call
    VOICEMAIL: 'voicemail', // New voicemail
    MESSAGE: 'message',  // New SMS/chat message
    SYSTEM: 'system'     // System alerts
}

// Notification store
const notifications = ref([])
const unreadCount = ref(0)
let notificationId = 0
let toastRef = null

// WebSocket connection
let ws = null
let reconnectAttempts = 0
const maxReconnectAttempts = 5
const reconnectDelay = 3000

/**
 * Initialize the notification service with toast reference
 */
export function initNotifications(toast) {
    toastRef = toast
}

/**
 * Add a notification
 */
export function addNotification(options) {
    const id = ++notificationId
    const notification = {
        id,
        type: options.type || NotificationType.INFO,
        title: options.title,
        message: options.message,
        timestamp: new Date(),
        read: false,
        data: options.data || null, // Extra data (e.g., call ID, voicemail ID)
        action: options.action || null, // Optional callback
        persistent: options.persistent ?? false // Whether to keep in history
    }

    notifications.value.unshift(notification)

    // Limit stored notifications
    if (notifications.value.length > 100) {
        notifications.value.pop()
    }

    // Update unread count
    unreadCount.value = notifications.value.filter(n => !n.read).length

    // Show toast for non-persistent notifications
    if (!options.silent && toastRef) {
        const toastType = ['call', 'voicemail', 'message', 'system'].includes(notification.type)
            ? 'info'
            : notification.type
        toastRef[toastType]?.(notification.message, notification.title)
    }

    return id
}

/**
 * Mark notification as read
 */
export function markAsRead(id) {
    const notification = notifications.value.find(n => n.id === id)
    if (notification && !notification.read) {
        notification.read = true
        unreadCount.value = notifications.value.filter(n => !n.read).length
    }
}

/**
 * Mark all notifications as read
 */
export function markAllAsRead() {
    notifications.value.forEach(n => n.read = true)
    unreadCount.value = 0
}

/**
 * Clear all notifications
 */
export function clearNotifications() {
    notifications.value = []
    unreadCount.value = 0
}

/**
 * Remove a specific notification
 */
export function removeNotification(id) {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
        const wasUnread = !notifications.value[index].read
        notifications.value.splice(index, 1)
        if (wasUnread) {
            unreadCount.value = Math.max(0, unreadCount.value - 1)
        }
    }
}

/**
 * Connect to WebSocket for real-time notifications
 */
export function connectWebSocket(token) {
    if (ws && ws.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/api/ws/notifications?token=${token}`

    try {
        ws = new WebSocket(wsUrl)

        ws.onopen = () => {
            console.log('Notification WebSocket connected')
            reconnectAttempts = 0
        }

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data)
                handleWebSocketMessage(data)
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e)
            }
        }

        ws.onclose = (event) => {
            console.log('Notification WebSocket closed:', event.code)
            ws = null

            // Attempt reconnect for non-intentional closures
            if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++
                setTimeout(() => connectWebSocket(token), reconnectDelay * reconnectAttempts)
            }
        }

        ws.onerror = (error) => {
            console.error('WebSocket error:', error)
        }

    } catch (e) {
        console.error('Failed to connect WebSocket:', e)
    }
}

/**
 * Disconnect WebSocket
 */
export function disconnectWebSocket() {
    if (ws) {
        ws.close(1000, 'User disconnect')
        ws = null
    }
    reconnectAttempts = maxReconnectAttempts // Prevent auto-reconnect
}

/**
 * Handle incoming WebSocket messages
 */
function handleWebSocketMessage(data) {
    switch (data.type) {
        case 'call_incoming':
            addNotification({
                type: NotificationType.CALL,
                title: 'Incoming Call',
                message: `From: ${data.caller_id || data.caller_number}`,
                data: { callId: data.call_id },
                persistent: true
            })
            break

        case 'call_missed':
            addNotification({
                type: NotificationType.CALL,
                title: 'Missed Call',
                message: `From: ${data.caller_id || data.caller_number}`,
                data: { callId: data.call_id },
                persistent: true
            })
            break

        case 'voicemail_new':
            addNotification({
                type: NotificationType.VOICEMAIL,
                title: 'New Voicemail',
                message: `From: ${data.caller_id} (${data.duration}s)`,
                data: { vmId: data.voicemail_id, boxId: data.box_id },
                persistent: true
            })
            break

        case 'message_new':
            addNotification({
                type: NotificationType.MESSAGE,
                title: 'New Message',
                message: data.preview || 'New message received',
                data: { threadId: data.thread_id },
                persistent: true
            })
            break

        case 'system_alert':
            addNotification({
                type: NotificationType.SYSTEM,
                title: data.title || 'System Alert',
                message: data.message,
                persistent: data.persistent ?? false
            })
            break

        default:
            // Generic notification
            if (data.message) {
                addNotification({
                    type: data.notification_type || NotificationType.INFO,
                    title: data.title,
                    message: data.message
                })
            }
    }
}

// Export reactive state for components
export function useNotifications() {
    return {
        notifications,
        unreadCount,
        addNotification,
        markAsRead,
        markAllAsRead,
        clearNotifications,
        removeNotification,
        connectWebSocket,
        disconnectWebSocket
    }
}
