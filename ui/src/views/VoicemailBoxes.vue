<template>
  <div class="voicemail-manager">
    <div class="view-header">
      <div class="header-content">
        <h2>Voicemail Manager</h2>
        <p class="text-muted text-sm">Manage voicemail boxes, quotas, delivery status, and authentication.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showSettingsModal = true">
          <SettingsIcon class="btn-icon" /> Settings
        </button>
        <button class="btn-primary" @click="showAddModal = true">
          <PlusIcon class="btn-icon" /> Add Voicemail Box
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon total"><InboxIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ boxes.length }}</span>
          <span class="stat-label">Total Boxes</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon messages"><MailIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ totalMessages }}</span>
          <span class="stat-label">Total Messages</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon delivered"><CheckCircleIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ deliveredCount }}</span>
          <span class="stat-label">Delivered Today</span>
        </div>
      </div>
      <div class="stat-card clickable" @click="activeTab = 'attempts'">
        <div class="stat-icon failed"><AlertCircleIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ failedCount }}</span>
          <span class="stat-label">Failed Deliveries</span>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'boxes' }" @click="activeTab = 'boxes'">
        <InboxIcon class="tab-icon" /> Voicemail Boxes
      </button>
      <button class="tab" :class="{ active: activeTab === 'attempts' }" @click="activeTab = 'attempts'">
        <SendIcon class="tab-icon" /> Delivery Attempts
        <span class="tab-badge" v-if="failedCount > 0">{{ failedCount }}</span>
      </button>
      <button class="tab" :class="{ active: activeTab === 'status' }" @click="activeTab = 'status'">
        <ActivityIcon class="tab-icon" /> System Status
      </button>
    </div>

    <!-- VOICEMAIL BOXES TAB -->
    <div class="tab-content" v-if="activeTab === 'boxes'">
      <div class="filter-bar">
        <div class="search-box">
          <SearchIcon class="search-icon" />
          <input v-model="searchQuery" class="search-input" placeholder="Search by extension or owner...">
        </div>
        <select v-model="typeFilter" class="filter-select">
          <option value="">All Types</option>
          <option value="Standard">Standard</option>
          <option value="Shared">Shared</option>
          <option value="Room">Room</option>
        </select>
      </div>

      <div class="boxes-table">
        <table>
          <thead>
            <tr>
              <th>Extension</th>
              <th>Type</th>
              <th>Owner / Group</th>
              <th>Notification Email</th>
              <th>Storage Usage</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="box in filteredBoxes" :key="box.ext" :class="{ 'high-usage': box.usage > 80 }">
              <td class="mono">{{ box.ext }}</td>
              <td><span class="type-badge" :class="box.type.toLowerCase()">{{ box.type }}</span></td>
              <td>{{ box.owner }}</td>
              <td class="email-cell">{{ box.email }}</td>
              <td>
                <div class="usage-bar">
                  <div class="bar-fill" :style="{ width: box.usage + '%' }" :class="{ high: box.usage > 80 }"></div>
                </div>
                <span class="usage-text">{{ box.count }} msgs / {{ box.usage }}%</span>
              </td>
              <td>
                <span class="status-badge" :class="box.status">{{ box.status }}</span>
              </td>
              <td class="actions-cell">
                <button class="action-btn" @click="editBox(box)" title="Edit"><EditIcon class="icon-sm" /></button>
                <button class="action-btn" @click="viewMessages(box)" title="View Messages"><MailIcon class="icon-sm" /></button>
                <button class="action-btn danger" @click="deleteBox(box)" title="Delete"><TrashIcon class="icon-sm" /></button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- DELIVERY ATTEMPTS TAB -->
    <div class="tab-content" v-else-if="activeTab === 'attempts'">
      <div class="filter-bar">
        <select v-model="attemptFilter" class="filter-select">
          <option value="">All Attempts</option>
          <option value="delivered">Delivered</option>
          <option value="pending">Pending</option>
          <option value="failed">Failed</option>
        </select>
        <select v-model="attemptDateFilter" class="filter-select">
          <option value="today">Today</option>
          <option value="week">This Week</option>
          <option value="month">This Month</option>
        </select>
        <button class="btn-secondary" @click="retryFailed" :disabled="failedCount === 0">
          <RefreshCwIcon class="btn-icon" /> Retry Failed
        </button>
      </div>

      <div class="attempts-list">
        <div 
          class="attempt-item" 
          v-for="attempt in filteredAttempts" 
          :key="attempt.id"
          :class="attempt.status"
        >
          <div class="attempt-icon" :class="attempt.status">
            <CheckCircleIcon v-if="attempt.status === 'delivered'" class="icon-sm" />
            <ClockIcon v-else-if="attempt.status === 'pending'" class="icon-sm" />
            <XCircleIcon v-else class="icon-sm" />
          </div>
          <div class="attempt-main">
            <div class="attempt-header">
              <span class="attempt-mailbox">Mailbox {{ attempt.mailbox }}</span>
              <span class="attempt-type">{{ attempt.type }}</span>
            </div>
            <div class="attempt-details">
              <span>To: {{ attempt.destination }}</span>
              <span class="attempt-time">{{ attempt.time }}</span>
            </div>
            <div class="attempt-error" v-if="attempt.error">
              <AlertCircleIcon class="error-icon" /> {{ attempt.error }}
            </div>
          </div>
          <div class="attempt-actions">
            <button class="btn-link" @click="viewAttemptDetails(attempt)">Details</button>
            <button class="btn-link" @click="retryAttempt(attempt)" v-if="attempt.status === 'failed'">Retry</button>
          </div>
        </div>
      </div>
    </div>

    <!-- SYSTEM STATUS TAB -->
    <div class="tab-content" v-else-if="activeTab === 'status'">
      <div class="status-grid">
        <div class="status-card">
          <div class="status-header">
            <h4>Email Delivery</h4>
            <span class="status-indicator online"></span>
          </div>
          <div class="status-details">
            <div class="status-row">
              <span>SMTP Server</span>
              <span class="mono">smtp.mailserver.com:587</span>
            </div>
            <div class="status-row">
              <span>Connection</span>
              <span class="status-good">Connected (TLS)</span>
            </div>
            <div class="status-row">
              <span>Last Delivery</span>
              <span>2 minutes ago</span>
            </div>
          </div>
        </div>

        <div class="status-card">
          <div class="status-header">
            <h4>Storage</h4>
            <span class="status-indicator online"></span>
          </div>
          <div class="status-details">
            <div class="storage-bar">
              <div class="storage-fill" style="width: 35%"></div>
            </div>
            <div class="status-row">
              <span>Used</span>
              <span>3.5 GB / 10 GB</span>
            </div>
            <div class="status-row">
              <span>Messages</span>
              <span>{{ totalMessages }} total</span>
            </div>
          </div>
        </div>

        <div class="status-card">
          <div class="status-header">
            <h4>Transcription Service</h4>
            <span class="status-indicator online"></span>
          </div>
          <div class="status-details">
            <div class="status-row">
              <span>Provider</span>
              <span>Google Cloud Speech</span>
            </div>
            <div class="status-row">
              <span>Transcriptions Today</span>
              <span>47</span>
            </div>
            <div class="status-row">
              <span>Avg. Processing Time</span>
              <span>2.3 seconds</span>
            </div>
          </div>
        </div>

        <div class="status-card">
          <div class="status-header">
            <h4>MWI (Message Waiting)</h4>
            <span class="status-indicator online"></span>
          </div>
          <div class="status-details">
            <div class="status-row">
              <span>Protocol</span>
              <span>SIP NOTIFY</span>
            </div>
            <div class="status-row">
              <span>Active Subscriptions</span>
              <span>24</span>
            </div>
            <div class="status-row">
              <span>Last Update</span>
              <span>Just now</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- SETTINGS MODAL -->
    <div v-if="showSettingsModal" class="modal-overlay" @click.self="showSettingsModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Voicemail Settings</h3>
          <button class="btn-icon" @click="showSettingsModal = false"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="settings-section">
            <h4>Default Settings</h4>
            <div class="form-group">
              <label>Max Message Length (seconds)</label>
              <input v-model="settings.maxLength" type="number" class="input-field">
            </div>
            <div class="form-group">
              <label>Default Storage Quota (MB)</label>
              <input v-model="settings.quota" type="number" class="input-field">
            </div>
            <div class="form-group">
              <label>Message Retention (days)</label>
              <input v-model="settings.retention" type="number" class="input-field">
            </div>
          </div>

          <div class="settings-section">
            <h4>Email Delivery</h4>
            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="settings.emailEnabled">
                <span>Enable email delivery</span>
              </label>
            </div>
            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="settings.attachAudio">
                <span>Attach audio file to email</span>
              </label>
            </div>
            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="settings.transcription">
                <span>Enable voicemail transcription</span>
              </label>
            </div>
          </div>

          <div class="settings-section">
            <h4>Security</h4>
            <div class="form-group">
              <label>Minimum PIN Length</label>
              <input v-model="settings.minPin" type="number" class="input-field" style="width: 100px;">
            </div>
            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="settings.requirePin">
                <span>Require PIN for internal access</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showSettingsModal = false">Cancel</button>
          <button class="btn-primary" @click="saveSettings">Save Settings</button>
        </div>
      </div>
    </div>

    <!-- ADD/EDIT BOX MODAL -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ editingBox ? 'Edit Voicemail Box' : 'Add Voicemail Box' }}</h3>
          <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group">
              <label>Extension *</label>
              <input v-model="boxForm.ext" class="input-field mono" placeholder="101">
            </div>
            <div class="form-group">
              <label>Type</label>
              <select v-model="boxForm.type" class="input-field">
                <option value="Standard">Standard</option>
                <option value="Shared">Shared</option>
                <option value="Room">Room</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>Owner / Group Name *</label>
            <input v-model="boxForm.owner" class="input-field" placeholder="John Smith">
          </div>
          <div class="form-group">
            <label>Notification Email</label>
            <input v-model="boxForm.email" type="email" class="input-field" placeholder="user@company.com">
          </div>
          <div class="form-group">
            <label>PIN</label>
            <input v-model="boxForm.pin" type="password" class="input-field" placeholder="••••">
          </div>
          <div class="form-group">
            <label>Storage Quota (MB)</label>
            <input v-model="boxForm.quota" type="number" class="input-field">
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveBox">{{ editingBox ? 'Save Changes' : 'Create Box' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import {
  Inbox as InboxIcon, Mail as MailIcon, CheckCircle as CheckCircleIcon, 
  AlertCircle as AlertCircleIcon, Send as SendIcon, Activity as ActivityIcon,
  Search as SearchIcon, Edit as EditIcon, Trash2 as TrashIcon, X as XIcon,
  Settings as SettingsIcon, Plus as PlusIcon, RefreshCw as RefreshCwIcon,
  Clock as ClockIcon, XCircle as XCircleIcon
} from 'lucide-vue-next'
import { voicemailAPI } from '@/services/api'

const toast = inject('toast')
const isLoading = ref(false)

const activeTab = ref('boxes')
const searchQuery = ref('')
const typeFilter = ref('')
const attemptFilter = ref('')
const attemptDateFilter = ref('today')
const showSettingsModal = ref(false)
const showAddModal = ref(false)
const editingBox = ref(null)

const settings = ref({
  maxLength: 180,
  quota: 50,
  retention: 30,
  emailEnabled: true,
  attachAudio: true,
  transcription: true,
  minPin: 4,
  requirePin: false
})

const boxForm = ref({
  ext: '',
  type: 'Standard',
  owner: '',
  email: '',
  pin: '',
  quota: 50
})

const boxes = ref([])
const deliveryAttempts = ref([])

onMounted(async () => {
  await fetchVoicemailBoxes()
})

async function fetchVoicemailBoxes() {
  isLoading.value = true
  try {
    const response = await voicemailAPI.list()
    boxes.value = (response.data || []).map(b => ({
      ext: b.extension,
      type: b.type || 'Standard',
      owner: b.owner_name || 'Unknown',
      email: b.notification_email || '',
      count: b.message_count || 0,
      usage: b.storage_percent || 0,
      status: b.storage_percent > 90 ? 'full' : 'active'
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load voicemail boxes')
    // Fallback to demo data
    boxes.value = [
      { ext: '101', type: 'Standard', owner: 'Alice Smith', email: 'alice@acme.com', count: 12, usage: 15, status: 'active' },
      { ext: '102', type: 'Standard', owner: 'Bob Jones', email: 'bob@acme.com', count: 0, usage: 0, status: 'active' },
      { ext: '8000', type: 'Shared', owner: 'Sales Queue (Shared)', email: 'sales@acme.com', count: 5, usage: 8, status: 'active' },
    ]
    deliveryAttempts.value = [
      { id: 1, mailbox: '101', type: 'Email', destination: 'alice@acme.com', status: 'delivered', time: '10:32 AM', error: null },
      { id: 2, mailbox: '105', type: 'Email', destination: 'david@acme.com', status: 'failed', time: '10:15 AM', error: 'SMTP timeout' },
    ]
  } finally {
    isLoading.value = false
  }
}

const totalMessages = computed(() => boxes.value.reduce((sum, b) => sum + b.count, 0))
const deliveredCount = computed(() => deliveryAttempts.value.filter(a => a.status === 'delivered').length)
const failedCount = computed(() => deliveryAttempts.value.filter(a => a.status === 'failed').length)

const filteredBoxes = computed(() => {
  return boxes.value.filter(box => {
    const matchesSearch = !searchQuery.value || 
      box.ext.includes(searchQuery.value) ||
      box.owner.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesType = !typeFilter.value || box.type === typeFilter.value
    return matchesSearch && matchesType
  })
})

const filteredAttempts = computed(() => {
  return deliveryAttempts.value.filter(a => !attemptFilter.value || a.status === attemptFilter.value)
})

const editBox = (box) => {
  editingBox.value = box
  boxForm.value = { ...box }
  showAddModal.value = true
}

const viewMessages = (box) => toast?.info(`Viewing messages for ${box.ext}`)

const deleteBox = async (box) => {
  if (confirm(`Delete voicemail box ${box.ext}?`)) {
    try {
      await voicemailAPI.delete(box.ext)
      toast?.success(`Voicemail box ${box.ext} deleted`)
      await fetchVoicemailBoxes()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete voicemail box')
    }
  }
}

const saveBox = async () => {
  try {
    const data = {
      extension: boxForm.value.ext,
      type: boxForm.value.type,
      owner_name: boxForm.value.owner,
      notification_email: boxForm.value.email,
      pin: boxForm.value.pin,
      quota_mb: boxForm.value.quota,
    }
    
    if (editingBox.value) {
      await voicemailAPI.update(boxForm.value.ext, data)
      toast?.success('Voicemail box updated')
    } else {
      await voicemailAPI.create(data)
      toast?.success('Voicemail box created')
    }
    
    await fetchVoicemailBoxes()
    closeModal()
  } catch (error) {
    toast?.error(error.message, 'Failed to save voicemail box')
  }
}

const closeModal = () => {
  showAddModal.value = false
  editingBox.value = null
  boxForm.value = { ext: '', type: 'Standard', owner: '', email: '', pin: '', quota: 50 }
}

const saveSettings = () => {
  showSettingsModal.value = false
  toast?.success('Settings saved')
}

const retryFailed = () => toast?.info('Retrying failed deliveries...')
const retryAttempt = (attempt) => toast?.info(`Retrying delivery to ${attempt.destination}`)
const viewAttemptDetails = (attempt) => toast?.info(`Attempt ID: ${attempt.id}, Status: ${attempt.status}`)
</script>

<style scoped>
.voicemail-manager { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-card.clickable { cursor: pointer; transition: all 0.15s; }
.stat-card.clickable:hover { border-color: var(--primary-color); }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.total { background: #dbeafe; color: #2563eb; }
.stat-icon.messages { background: #f3e8ff; color: #7c3aed; }
.stat-icon.delivered { background: #dcfce7; color: #16a34a; }
.stat-icon.failed { background: #fee2e2; color: #dc2626; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 16px; height: 16px; }
.tab-badge { background: #ef4444; color: white; font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 10px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); min-height: 300px; }

/* Filter Bar */
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-box { position: relative; flex: 1; max-width: 300px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 10px 12px 10px 38px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.filter-select { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }

/* Boxes Table */
.boxes-table { overflow-x: auto; }
.boxes-table table { width: 100%; border-collapse: collapse; }
.boxes-table th { text-align: left; padding: 12px; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 2px solid var(--border-color); }
.boxes-table td { padding: 12px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.boxes-table tr:hover { background: var(--bg-app); }
.boxes-table tr.high-usage { background: #fef2f2; }
.mono { font-family: monospace; font-weight: 600; }

.type-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; text-transform: uppercase; }
.type-badge.standard { background: #dbeafe; color: #2563eb; }
.type-badge.shared { background: #fef3c7; color: #b45309; }
.type-badge.room { background: #f3e8ff; color: #7c3aed; }

.email-cell { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.usage-bar { height: 6px; width: 100px; background: var(--bg-app); border-radius: 4px; overflow: hidden; margin-bottom: 4px; }
.bar-fill { height: 100%; background: var(--status-good); transition: width 0.3s; }
.bar-fill.high { background: var(--status-bad); }
.usage-text { font-size: 11px; color: var(--text-muted); }

.status-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; text-transform: uppercase; }
.status-badge.active { background: #dcfce7; color: #16a34a; }
.status-badge.full { background: #fee2e2; color: #dc2626; }

.actions-cell { display: flex; gap: 4px; }
.action-btn { width: 28px; height: 28px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }

/* Attempts List */
.attempts-list { display: flex; flex-direction: column; gap: 8px; }
.attempt-item { display: flex; align-items: flex-start; gap: 12px; padding: 14px; background: var(--bg-app); border-radius: var(--radius-sm); border-left: 3px solid transparent; }
.attempt-item.delivered { border-left-color: #22c55e; }
.attempt-item.pending { border-left-color: #f59e0b; }
.attempt-item.failed { border-left-color: #ef4444; background: #fef2f2; }

.attempt-icon { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
.attempt-icon.delivered { background: #dcfce7; color: #16a34a; }
.attempt-icon.pending { background: #fef3c7; color: #b45309; }
.attempt-icon.failed { background: #fee2e2; color: #dc2626; }

.attempt-main { flex: 1; }
.attempt-header { display: flex; gap: 12px; margin-bottom: 4px; }
.attempt-mailbox { font-weight: 600; }
.attempt-type { font-size: 11px; color: var(--text-muted); background: white; padding: 2px 8px; border-radius: 4px; }
.attempt-details { font-size: 12px; color: var(--text-muted); display: flex; gap: 16px; }
.attempt-error { margin-top: 8px; font-size: 12px; color: #dc2626; display: flex; align-items: center; gap: 6px; }
.error-icon { width: 14px; height: 14px; }

.attempt-actions { display: flex; gap: 8px; }

/* Status Grid */
.status-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.status-card { background: var(--bg-app); border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; }
.status-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.status-header h4 { margin: 0; font-size: 14px; }
.status-indicator { width: 10px; height: 10px; border-radius: 50%; }
.status-indicator.online { background: #22c55e; }
.status-indicator.offline { background: #ef4444; }

.status-details { display: flex; flex-direction: column; gap: 8px; }
.status-row { display: flex; justify-content: space-between; font-size: 12px; }
.status-row span:first-child { color: var(--text-muted); }
.status-good { color: #16a34a; font-weight: 600; }

.storage-bar { height: 8px; background: #e2e8f0; border-radius: 4px; overflow: hidden; margin-bottom: 12px; }
.storage-fill { height: 100%; background: var(--primary-color); }

/* Settings Section */
.settings-section { margin-bottom: 24px; }
.settings-section h4 { margin: 0 0 12px; font-size: 14px; padding-bottom: 8px; border-bottom: 1px solid var(--border-color); }

/* Form */
.form-row { display: flex; gap: 12px; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; flex: 1; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 480px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-link { background: none; border: none; color: var(--primary-color); font-weight: 600; cursor: pointer; font-size: 12px; }
.btn-icon { width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }
.icon { width: 20px; height: 20px; }

@media (max-width: 768px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .status-grid { grid-template-columns: 1fr; }
}
</style>
