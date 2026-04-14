<template>
  <div class="chat-page">
    <!-- Header -->
    <div class="view-header">
      <div class="header-content">
        <h2>Chat Management</h2>
        <p class="text-muted text-sm">Manage chat threads, rooms, and queues.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showRoomModal = true">
          <PlusIcon class="btn-icon" /> New Room
        </button>
        <button class="btn-secondary" @click="showQueueModal = true">
          <PlusIcon class="btn-icon" /> New Queue
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'threads' }" @click="activeTab = 'threads'">
        <MessageSquareIcon class="tab-icon" /> Threads
      </button>
      <button class="tab" :class="{ active: activeTab === 'rooms' }" @click="activeTab = 'rooms'">
        <DoorOpenIcon class="tab-icon" /> Rooms
      </button>
      <button class="tab" :class="{ active: activeTab === 'queues' }" @click="activeTab = 'queues'">
        <HeadphonesIcon class="tab-icon" /> Queues
      </button>
    </div>

    <!-- Threads Tab -->
    <div class="tab-content" v-if="activeTab === 'threads'">
      <div class="threads-layout">
        <!-- Thread List -->
        <div class="thread-list-panel">
          <div class="panel-header">
            <h3>Threads</h3>
            <button class="btn-icon-sm" @click="showThreadModal = true" title="New Thread">
              <PlusIcon class="icon-xs" />
            </button>
          </div>
          <div class="search-box">
            <Search class="icon-xs muted" />
            <input type="text" placeholder="Search threads..." v-model="threadSearch" class="search-input">
          </div>
          <div class="thread-list">
            <div v-if="filteredThreads.length === 0" class="empty-state">
              No threads found
            </div>
            <div
              v-for="thread in filteredThreads"
              :key="thread.id"
              class="thread-item"
              :class="{ active: selectedThread?.id === thread.id }"
              @click="selectThread(thread)"
            >
              <div class="thread-avatar">{{ thread.initials }}</div>
              <div class="thread-info">
                <div class="thread-top">
                  <span class="thread-name">{{ thread.name }}</span>
                  <span class="thread-time">{{ thread.lastTime }}</span>
                </div>
                <div class="thread-preview">{{ thread.preview }}</div>
              </div>
              <div class="unread-badge" v-if="thread.unread">{{ thread.unread }}</div>
            </div>
          </div>
        </div>

        <!-- Thread Detail -->
        <div class="thread-detail-panel">
          <template v-if="selectedThread">
            <div class="detail-header">
              <div class="user-details">
                <h3>{{ selectedThread.name }}</h3>
                <span class="status-text">{{ selectedThread.status || 'Active' }}</span>
              </div>
              <div class="header-actions">
                <button class="btn-icon" title="More">
                  <MoreVerticalIcon class="icon-sm" />
                </button>
              </div>
            </div>

            <div class="messages-area">
              <div v-if="isLoadingMessages" class="loading-state">
                Loading messages...
              </div>
              <div v-else-if="threadMessages.length === 0" class="empty-state">
                No messages yet. Start the conversation!
              </div>
              <div
                v-for="msg in threadMessages"
                :key="msg.id"
                class="message"
                :class="msg.direction"
              >
                <div class="message-bubble">{{ msg.body }}</div>
                <div class="message-meta">
                  <span class="message-sender">{{ msg.sender }}</span>
                  <span class="message-time">{{ msg.time }}</span>
                </div>
              </div>
            </div>

            <div class="compose-area">
              <input
                type="text"
                class="compose-input"
                placeholder="Type a message..."
                v-model="newMessage"
                @keydown.enter="sendMessage"
              >
              <button class="btn-primary send-btn" @click="sendMessage" :disabled="!newMessage.trim()">
                <SendIcon class="icon-sm" />
              </button>
            </div>
          </template>
          <div v-else class="no-selection">
            <MessageSquareIcon class="no-selection-icon" />
            <p>Select a thread to view messages</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Rooms Tab -->
    <div class="tab-content" v-if="activeTab === 'rooms'">
      <div class="rooms-grid">
        <div v-if="rooms.length === 0" class="empty-state">
          No rooms found. Create one to get started.
        </div>
        <div v-for="room in rooms" :key="room.id" class="room-card">
          <div class="room-header">
            <div class="room-info">
              <h4>{{ room.name }}</h4>
              <span class="room-topic">{{ room.topic || 'No topic' }}</span>
            </div>
            <span class="room-status" :class="room.status?.toLowerCase() || 'active'">
              {{ room.status || 'Active' }}
            </span>
          </div>
          <div class="room-stats">
            <div class="room-stat">
              <UsersIcon class="stat-icon" />
              <span>{{ room.members || 0 }} members</span>
            </div>
          </div>
          <div class="room-footer">
            <button class="btn-secondary btn-sm" @click="joinRoom(room)">
              Join Room
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Queues Tab -->
    <div class="tab-content" v-if="activeTab === 'queues'">
      <table class="data-table" v-if="chatQueues.length > 0">
        <thead>
          <tr>
            <th>Name</th>
            <th>Extension</th>
            <th>Strategy</th>
            <th>Agents</th>
            <th>Waiting</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="queue in chatQueues" :key="queue.id">
            <td><strong>{{ queue.name }}</strong></td>
            <td class="mono">{{ queue.extension }}</td>
            <td><span class="strategy-pill">{{ queue.strategy || 'Round Robin' }}</span></td>
            <td>{{ queue.agent_count || 0 }}</td>
            <td>{{ queue.waiting || 0 }}</td>
            <td>
              <span class="status-dot" :class="queue.status === 'active' ? 'active' : 'inactive'"></span>
              {{ queue.status || 'Active' }}
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">
        No chat queues found. Create one to manage customer chats.
      </div>
    </div>

    <!-- New Thread Modal -->
    <div class="modal-overlay" v-if="showThreadModal" @click.self="showThreadModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>New Thread</h3>
          <button class="close-btn" @click="showThreadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Subject</label>
            <input v-model="threadForm.subject" class="input-field" placeholder="Enter thread subject">
          </div>
          <div class="form-group">
            <label>Participant</label>
            <input v-model="threadForm.participant" class="input-field" placeholder="Enter participant ID or name">
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showThreadModal = false">Cancel</button>
          <button class="btn-primary" @click="createThread">Create Thread</button>
        </div>
      </div>
    </div>

    <!-- New Room Modal -->
    <div class="modal-overlay" v-if="showRoomModal" @click.self="showRoomModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>New Room</h3>
          <button class="close-btn" @click="showRoomModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Name</label>
            <input v-model="roomForm.name" class="input-field" placeholder="Room name">
          </div>
          <div class="form-group">
            <label>Topic</label>
            <input v-model="roomForm.topic" class="input-field" placeholder="Room topic (optional)">
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showRoomModal = false">Cancel</button>
          <button class="btn-primary" @click="createRoom">Create Room</button>
        </div>
      </div>
    </div>

    <!-- New Queue Modal -->
    <div class="modal-overlay" v-if="showQueueModal" @click.self="showQueueModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>New Chat Queue</h3>
          <button class="close-btn" @click="showQueueModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Name</label>
            <input v-model="queueForm.name" class="input-field" placeholder="Queue name">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Extension</label>
              <input v-model="queueForm.extension" class="input-field" placeholder="600">
            </div>
            <div class="form-group">
              <label>Strategy</label>
              <select v-model="queueForm.strategy" class="input-field">
                <option value="round_robin">Round Robin</option>
                <option value="least_busy">Least Busy</option>
                <option value="fewest_calls">Fewest Calls</option>
              </select>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showQueueModal = false">Cancel</button>
          <button class="btn-primary" @click="createQueue">Create Queue</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import {
  Plus as PlusIcon,
  MessageSquare as MessageSquareIcon,
  MessageSquarePlus as MessageSquarePlusIcon,
  DoorOpen as DoorOpenIcon,
  Headphones as HeadphonesIcon,
  Users as UsersIcon,
  Search,
  Send as SendIcon,
  MoreVertical as MoreVerticalIcon
} from 'lucide-vue-next'
import { chatAPI } from '@/services/api'

const toast = inject('toast')

// State
const activeTab = ref('threads')
const isLoading = ref(false)
const isLoadingMessages = ref(false)

// Thread state
const threads = ref([])
const selectedThread = ref(null)
const threadMessages = ref([])
const threadSearch = ref('')
const newMessage = ref('')

// Room state
const rooms = ref([])

// Queue state
const chatQueues = ref([])

// Modals
const showThreadModal = ref(false)
const showRoomModal = ref(false)
const showQueueModal = ref(false)

// Forms
const threadForm = ref({ subject: '', participant: '' })
const roomForm = ref({ name: '', topic: '' })
const queueForm = ref({ name: '', extension: '', strategy: 'round_robin' })

// Computed
const filteredThreads = computed(() => {
  if (!threadSearch.value) return threads.value
  const q = threadSearch.value.toLowerCase()
  return threads.value.filter(t =>
    t.name.toLowerCase().includes(q) ||
    t.preview.toLowerCase().includes(q)
  )
})

// Helpers
const formatTime = (ts) => {
  if (!ts) return ''
  const dt = new Date(ts)
  const diffMin = Math.round((Date.now() - dt) / 60000)
  if (diffMin < 1) return 'Just now'
  if (diffMin < 60) return `${diffMin}m ago`
  if (diffMin < 1440) return `${Math.round(diffMin / 60)}h ago`
  return dt.toLocaleDateString()
}

const getInitials = (name) => {
  if (!name) return '??'
  return name.slice(0, 2).toUpperCase()
}

// Data Fetching
async function fetchThreads() {
  isLoading.value = true
  try {
    const response = await chatAPI.listThreads()
    threads.value = (response.data || []).map(t => ({
      id: t.id,
      name: t.subject || t.name || 'Unnamed Thread',
      initials: getInitials(t.subject || t.name),
      preview: t.last_message || t.preview || '',
      lastTime: formatTime(t.updated_at || t.last_message_at),
      unread: t.unread_count || 0,
      status: t.status || 'Active'
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load threads')
    threads.value = []
  } finally {
    isLoading.value = false
  }
}

async function fetchRooms() {
  try {
    const response = await chatAPI.listRooms()
    rooms.value = (response.data || []).map(r => ({
      id: r.id,
      name: r.name,
      topic: r.topic,
      members: r.member_count || 0,
      status: r.status || 'Active'
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load rooms')
    rooms.value = []
  }
}

async function fetchQueues() {
  try {
    const response = await chatAPI.listQueues()
    chatQueues.value = (response.data || []).map(q => ({
      id: q.id,
      name: q.name,
      extension: q.extension,
      strategy: q.strategy || 'round_robin',
      agent_count: q.agent_count || 0,
      waiting: q.waiting || 0,
      status: q.status || 'active'
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load queues')
    chatQueues.value = []
  }
}

async function selectThread(thread) {
  selectedThread.value = thread
  await fetchThreadMessages(thread.id)
}

async function fetchThreadMessages(threadId) {
  isLoadingMessages.value = true
  try {
    const response = await chatAPI.getThread(threadId)
    const data = response.data
    const messages = data?.messages || data?.data || []

    threadMessages.value = messages.map(m => ({
      id: m.id,
      body: m.body || m.content || '',
      direction: m.direction === 'outbound' || m.is_outgoing ? 'outgoing' : 'incoming',
      sender: m.sender_name || m.sender || 'Unknown',
      time: new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load messages')
    threadMessages.value = []
  } finally {
    isLoadingMessages.value = false
  }
}

// Actions
async function createThread() {
  const subject = threadForm.value.subject.trim()
  if (!subject) {
    toast?.error('Please enter a subject', 'Validation Error')
    return
  }

  try {
    await chatAPI.createThread({
      subject,
      participant: threadForm.value.participant
    })
    toast?.success('Thread created successfully')
    showThreadModal.value = false
    threadForm.value = { subject: '', participant: '' }
    await fetchThreads()
  } catch (error) {
    toast?.error(error.message, 'Failed to create thread')
  }
}

async function sendMessage() {
  const body = newMessage.value.trim()
  if (!body || !selectedThread.value) return

  const tempMessage = {
    id: Date.now(),
    body,
    direction: 'outgoing',
    sender: 'You',
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  threadMessages.value.push(tempMessage)
  newMessage.value = ''

  try {
    await chatAPI.sendMessage(selectedThread.value.id, { body })
  } catch (error) {
    toast?.error(error.message, 'Failed to send message')
    threadMessages.value = threadMessages.value.filter(m => m.id !== tempMessage.id)
  }
}

async function createRoom() {
  const name = roomForm.value.name.trim()
  if (!name) {
    toast?.error('Please enter a room name', 'Validation Error')
    return
  }

  try {
    await chatAPI.createRoom({
      name,
      topic: roomForm.value.topic
    })
    toast?.success('Room created successfully')
    showRoomModal.value = false
    roomForm.value = { name: '', topic: '' }
    await fetchRooms()
  } catch (error) {
    toast?.error(error.message, 'Failed to create room')
  }
}

async function joinRoom(room) {
  try {
    await chatAPI.joinRoom(room.id)
    toast?.success(`Joined room "${room.name}"`)
  } catch (error) {
    toast?.error(error.message, 'Failed to join room')
  }
}

async function createQueue() {
  const name = queueForm.value.name.trim()
  if (!name) {
    toast?.error('Please enter a queue name', 'Validation Error')
    return
  }

  try {
    await chatAPI.createQueue({
      name: queueForm.value.name,
      extension: queueForm.value.extension,
      strategy: queueForm.value.strategy
    })
    toast?.success('Queue created successfully')
    showQueueModal.value = false
    queueForm.value = { name: '', extension: '', strategy: 'round_robin' }
    await fetchQueues()
  } catch (error) {
    toast?.error(error.message, 'Failed to create queue')
  }
}

// Lifecycle
onMounted(async () => {
  await Promise.all([
    fetchThreads(),
    fetchRooms(),
    fetchQueues()
  ])
})
</script>

<style scoped>
.chat-page { padding: 0; }

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }

.btn-primary, .btn-secondary {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); margin-bottom: 0; }
.tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: transparent;
  border: 1px solid transparent;
  border-bottom: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  border-radius: 6px 6px 0 0;
}
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 14px; height: 14px; }

.tab-content {
  background: white;
  border: 1px solid var(--border-color);
  border-top: none;
  padding: 20px;
  border-radius: 0 0 8px 8px;
}

/* Threads Layout */
.threads-layout {
  display: flex;
  gap: 20px;
  height: calc(100vh - 280px);
  min-height: 400px;
}

.thread-list-panel {
  width: 320px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.panel-header h3 { margin: 0; font-size: 14px; font-weight: 600; }

.search-box {
  padding: 12px 16px;
  background: var(--bg-app);
  display: flex;
  align-items: center;
  gap: 8px;
}
.search-input {
  border: none;
  background: transparent;
  width: 100%;
  font-size: 13px;
  outline: none;
}

.thread-list {
  flex: 1;
  overflow-y: auto;
}

.thread-item {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.15s;
  position: relative;
}
.thread-item:hover { background: #f8fafc; }
.thread-item.active { background: #eff6ff; border-left: 3px solid var(--primary-color); }

.thread-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  flex-shrink: 0;
}

.thread-info { flex: 1; min-width: 0; }
.thread-top { display: flex; justify-content: space-between; margin-bottom: 4px; }
.thread-name { font-weight: 600; font-size: 13px; color: var(--text-primary); }
.thread-time { font-size: 10px; color: var(--text-muted); }
.thread-preview {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.unread-badge {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  background: var(--primary-color);
  color: white;
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

/* Thread Detail Panel */
.thread-detail-panel {
  flex: 1;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  background: #f8fafc;
}
.user-details h3 { margin: 0 0 4px; font-size: 15px; font-weight: 600; }
.status-text { font-size: 11px; color: var(--text-muted); }
.header-actions { display: flex; gap: 8px; }

.messages-area {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background: #f8fafc;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message { display: flex; flex-direction: column; max-width: 70%; }
.message.incoming { align-self: flex-start; }
.message.outgoing { align-self: flex-end; align-items: flex-end; }

.message-bubble {
  padding: 10px 16px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.5;
}
.message.incoming .message-bubble {
  background: white;
  border: 1px solid var(--border-color);
  border-bottom-left-radius: 2px;
}
.message.outgoing .message-bubble {
  background: var(--primary-color);
  color: white;
  border-bottom-right-radius: 2px;
}

.message-meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  padding: 0 4px;
}
.message-sender { font-size: 10px; font-weight: 600; color: var(--text-muted); }
.message-time { font-size: 10px; color: var(--text-muted); }

.compose-area {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-top: 1px solid var(--border-color);
  background: white;
}
.compose-input {
  flex: 1;
  padding: 10px 16px;
  border: 1px solid var(--border-color);
  border-radius: 24px;
  font-size: 13px;
  outline: none;
}
.compose-input:focus { border-color: var(--primary-color); }

.send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-color);
  border: none;
  color: white;
  cursor: pointer;
}
.send-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* No Selection */
.no-selection {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  gap: 12px;
}
.no-selection-icon { width: 48px; height: 48px; opacity: 0.4; }

/* Loading & Empty States */
.loading-state, .empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--text-muted);
  font-size: 13px;
}

/* Rooms Grid */
.rooms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.room-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: box-shadow 0.2s;
}
.room-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }

.room-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 14px;
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.room-info h4 { margin: 0 0 4px; font-size: 14px; }
.room-topic { font-size: 11px; color: var(--text-muted); }

.room-status {
  font-size: 10px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}
.room-status.active { background: #dcfce7; color: #16a34a; }
.room-status.inactive { background: #f1f5f9; color: #64748b; }

.room-stats { padding: 14px; }
.room-stat { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--text-muted); }
.stat-icon { width: 14px; height: 14px; }

.room-footer { padding: 10px 14px; background: #f8fafc; border-top: 1px solid var(--border-color); }

.btn-sm { padding: 6px 12px; font-size: 12px; }

/* Data Table */
.data-table {
  width: 100%;
  border-collapse: collapse;
}
.data-table th {
  text-align: left;
  padding: 10px 12px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
}
.data-table td {
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
  font-size: 13px;
}
.data-table tr:hover { background: #f8fafc; }

.mono { font-family: monospace; }
.strategy-pill { font-size: 11px; background: #f1f5f9; padding: 3px 8px; border-radius: 4px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-right: 6px; }
.status-dot.active { background: #22c55e; }
.status-dot.inactive { background: #94a3b8; }

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: #f1f5f9;
  border-radius: 6px;
  font-size: 18px;
  cursor: pointer;
}
.modal-body { padding: 20px; }
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.form-group { margin-bottom: 16px; }
.form-group label {
  display: block;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 6px;
}
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.input-field {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 13px;
  box-sizing: border-box;
}
.input-field:focus { border-color: var(--primary-color); outline: none; }

/* Icons */
.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  color: var(--text-muted);
}
.btn-icon:hover { background: var(--bg-app); color: var(--text-primary); }
.btn-icon-sm {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  cursor: pointer;
  color: var(--text-muted);
}
.btn-icon-sm:hover { color: var(--primary-color); border-color: var(--primary-color); }
.icon-xs { width: 14px; height: 14px; }
.icon-sm { width: 18px; height: 18px; }
.muted { color: var(--text-muted); }
</style>
