<template>
  <div class="notification-center">
    <!-- Trigger Button -->
    <button class="notification-btn" @click="isOpen = !isOpen" :class="{ 'has-unread': unreadCount > 0 }">
      <BellIcon class="bell-icon" />
      <span v-if="unreadCount > 0" class="badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>

    <!-- Dropdown Panel -->
    <Transition name="panel">
      <div v-if="isOpen" class="notification-panel" @click.stop>
        <div class="panel-header">
          <h4>Notifications</h4>
          <div class="header-actions">
            <button v-if="unreadCount > 0" class="mark-all-btn" @click="markAllAsRead">
              Mark all read
            </button>
            <button class="close-btn" @click="isOpen = false">
              <XIcon class="icon-sm" />
            </button>
          </div>
        </div>

        <div class="notification-list" v-if="notifications.length > 0">
          <div
            v-for="notification in notifications"
            :key="notification.id"
            class="notification-item"
            :class="{ unread: !notification.read, [notification.type]: true }"
            @click="handleClick(notification)"
          >
            <div class="notification-icon">
              <PhoneIncoming v-if="notification.type === 'call'" />
              <Voicemail v-else-if="notification.type === 'voicemail'" />
              <MessageSquare v-else-if="notification.type === 'message'" />
              <AlertCircle v-else-if="notification.type === 'system'" />
              <CheckCircle v-else-if="notification.type === 'success'" />
              <XCircle v-else-if="notification.type === 'error'" />
              <AlertTriangle v-else-if="notification.type === 'warning'" />
              <Info v-else />
            </div>
            <div class="notification-content">
              <div class="notification-title">{{ notification.title }}</div>
              <div class="notification-message">{{ notification.message }}</div>
              <div class="notification-time">{{ formatTime(notification.timestamp) }}</div>
            </div>
            <button class="dismiss-btn" @click.stop="removeNotification(notification.id)">
              <XIcon class="icon-xs" />
            </button>
          </div>
        </div>

        <div v-else class="empty-state">
          <BellOff class="empty-icon" />
          <p>No notifications</p>
        </div>

        <div v-if="notifications.length > 0" class="panel-footer">
          <button class="clear-btn" @click="clearNotifications">Clear All</button>
        </div>
      </div>
    </Transition>

    <!-- Click outside to close -->
    <div v-if="isOpen" class="overlay" @click="isOpen = false"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { 
  Bell as BellIcon, 
  BellOff, 
  X as XIcon,
  PhoneIncoming,
  Voicemail,
  MessageSquare,
  AlertCircle,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Info
} from 'lucide-vue-next'
import { useNotifications } from '@/services/notifications'

const { 
  notifications, 
  unreadCount, 
  markAsRead, 
  markAllAsRead, 
  clearNotifications,
  removeNotification 
} = useNotifications()

const isOpen = ref(false)

function handleClick(notification) {
  markAsRead(notification.id)
  
  // Handle action if defined
  if (notification.action) {
    notification.action()
    isOpen.value = false
  }
}

function formatTime(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  
  // Less than 1 minute
  if (diff < 60000) return 'Just now'
  
  // Less than 1 hour
  if (diff < 3600000) {
    const mins = Math.floor(diff / 60000)
    return `${mins}m ago`
  }
  
  // Less than 24 hours
  if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours}h ago`
  }
  
  // Same year
  if (date.getFullYear() === now.getFullYear()) {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  }
  
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

// Close on escape key
function handleKeydown(e) {
  if (e.key === 'Escape' && isOpen.value) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.notification-center {
  position: relative;
}

.notification-btn {
  position: relative;
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  color: var(--text-muted);
  transition: all 0.2s;
}

.notification-btn:hover {
  background: var(--bg-app);
  color: var(--text-primary);
}

.notification-btn.has-unread {
  color: var(--primary-color);
}

.bell-icon {
  width: 20px;
  height: 20px;
}

.badge {
  position: absolute;
  top: 2px;
  right: 2px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  font-size: 10px;
  font-weight: 700;
  color: white;
  background: #ef4444;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.overlay {
  position: fixed;
  inset: 0;
  z-index: 99;
}

.notification-panel {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  width: 380px;
  max-height: 480px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  z-index: 100;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-app);
}

.panel-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mark-all-btn {
  font-size: 11px;
  color: var(--primary-color);
  background: none;
  border: none;
  cursor: pointer;
  font-weight: 500;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
  display: flex;
}

.notification-list {
  flex: 1;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background 0.2s;
}

.notification-item:hover {
  background: var(--bg-app);
}

.notification-item.unread {
  background: #f0f9ff;
}

.notification-item.unread:hover {
  background: #e0f2fe;
}

.notification-icon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #e2e8f0;
  color: #64748b;
}

.notification-icon svg {
  width: 16px;
  height: 16px;
}

.notification-item.call .notification-icon { background: #dcfce7; color: #16a34a; }
.notification-item.voicemail .notification-icon { background: #fef3c7; color: #b45309; }
.notification-item.message .notification-icon { background: #dbeafe; color: #2563eb; }
.notification-item.system .notification-icon { background: #fae8ff; color: #a21caf; }
.notification-item.success .notification-icon { background: #dcfce7; color: #16a34a; }
.notification-item.error .notification-icon { background: #fee2e2; color: #dc2626; }
.notification-item.warning .notification-icon { background: #fef3c7; color: #b45309; }

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.notification-message {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notification-time {
  font-size: 10px;
  color: var(--text-muted);
  margin-top: 4px;
}

.dismiss-btn {
  flex-shrink: 0;
  opacity: 0;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
}

.notification-item:hover .dismiss-btn {
  opacity: 1;
}

.dismiss-btn:hover {
  color: #ef4444;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-muted);
}

.empty-icon {
  width: 40px;
  height: 40px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 13px;
}

.panel-footer {
  padding: 10px 16px;
  border-top: 1px solid var(--border-color);
  text-align: center;
}

.clear-btn {
  font-size: 12px;
  color: var(--text-muted);
  background: none;
  border: none;
  cursor: pointer;
  font-weight: 500;
}

.clear-btn:hover {
  color: #ef4444;
}

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 14px; height: 14px; }

/* Transitions */
.panel-enter-active,
.panel-leave-active {
  transition: all 0.2s ease;
}

.panel-enter-from,
.panel-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
