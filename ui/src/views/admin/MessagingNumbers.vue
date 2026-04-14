<template>
  <div class="messaging-numbers-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Messaging Numbers</h2>
        <p class="text-muted text-sm">Manage SMS-enabled numbers, configure settings, and assign to users.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="refreshNumbers" :disabled="loading">
          <RefreshCwIcon class="btn-icon" :class="{ spinning: loading }" /> Refresh
        </button>
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="error" class="error-banner">
      <AlertCircleIcon class="icon-sm" />
      <span>{{ error }}</span>
      <button @click="error = null" class="btn-icon"><XIcon class="icon-sm" /></button>
    </div>

    <!-- Numbers Table -->
    <div class="panel">
      <div v-if="loading && numbers.length === 0" class="loading-state">
        <RefreshCwIcon class="spinning icon-lg" />
        <span>Loading numbers...</span>
      </div>

      <div v-else-if="numbers.length === 0" class="empty-state">
        <MessageSquareIcon class="icon-xl muted" />
        <h3>No SMS Numbers</h3>
        <p class="text-muted">No SMS-enabled numbers found. Configure numbers in the system admin.</p>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>Phone Number</th>
            <th>Provider</th>
            <th>SMS</th>
            <th>MMS</th>
            <th>Assigned To</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="num in numbers" :key="num.id">
            <td>
              <div class="number-cell">
                <span class="phone-number mono">{{ num.phone_number }}</span>
                <span v-if="num.friendly_name" class="friendly-name">{{ num.friendly_name }}</span>
              </div>
            </td>
            <td>
              <span class="provider-badge" :class="num.provider_type">{{ num.provider_name || 'Unknown' }}</span>
            </td>
            <td>
              <span class="status-toggle" :class="num.sms_enabled ? 'enabled' : 'disabled'">
                <CheckIcon v-if="num.sms_enabled" class="icon-xs" />
                <XIcon v-else class="icon-xs" />
                SMS
              </span>
            </td>
            <td>
              <span class="status-toggle" :class="num.mms_enabled ? 'enabled' : 'disabled'">
                <CheckIcon v-if="num.mms_enabled" class="icon-xs" />
                <XIcon v-else class="icon-xs" />
                MMS
              </span>
            </td>
            <td>
              <div v-if="num.assignments && num.assignments.length > 0" class="assignments-cell">
                <span 
                  v-for="assign in num.assignments" 
                  :key="assign.id" 
                  class="assignment-tag"
                >
                  {{ assign.extension_number || assign.user_name || 'Assigned' }}
                  <button 
                    class="btn-unassign" 
                    @click="unassignNumber(num.id, assign.id)"
                    title="Unassign"
                  >
                    <XIcon class="icon-xs" />
                  </button>
                </span>
              </div>
              <span v-else class="text-muted">—</span>
            </td>
            <td>
              <div class="action-buttons">
                <button 
                  class="action-btn" 
                  @click="openSmsConfig(num)"
                  title="Configure SMS"
                >
                  <SettingsIcon class="icon-sm" />
                </button>
                <button 
                  class="action-btn" 
                  @click="openAssignmentModal(num)"
                  title="Assign Number"
                >
                  <UserPlusIcon class="icon-sm" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- SMS Configuration Modal -->
    <div v-if="showSmsModal" class="modal-overlay" @click.self="closeSmsModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Configure SMS: {{ selectedNumber?.phone_number }}</h3>
          <button class="btn-icon" @click="closeSmsModal"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Voice Handler (IVR/Hunting)</label>
            <select v-model="smsForm.voice_handler_id" class="input-field">
              <option :value="null">None</option>
              <option v-for="iv in ivrMenus" :key="iv.id" :value="iv.id">{{ iv.name }}</option>
            </select>
            <span class="help-text">Where to send inbound calls to this number (for voice, not SMS)</span>
          </div>
          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="smsForm.reply_to_sms_enabled" />
              <span>Enable SMS Replies</span>
            </label>
            <span class="help-text">Allow recipients to reply directly to incoming SMS</span>
          </div>
          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="smsForm.conversation_on_reply" />
              <span>Start New Conversation on Reply</span>
            </label>
            <span class="help-text">Create a new conversation thread when someone replies to a message</span>
          </div>
          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="smsForm.webhook_enabled" />
              <span>Forward to Webhook</span>
            </label>
          </div>
          <div v-if="smsForm.webhook_enabled" class="form-group">
            <label>Webhook URL</label>
            <input v-model="smsForm.webhook_url" class="input-field" placeholder="https://..." />
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="closeSmsModal">Cancel</button>
          <button class="btn-primary" @click="saveSmsConfig" :disabled="saving">
            {{ saving ? 'Saving...' : 'Save Configuration' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Assignment Modal -->
    <div v-if="showAssignmentModal" class="modal-overlay" @click.self="closeAssignmentModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Assign: {{ selectedNumber?.phone_number }}</h3>
          <button class="btn-icon" @click="closeAssignmentModal"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Assignment Type</label>
            <select v-model="assignmentForm.type" class="input-field">
              <option value="extension">Extension</option>
              <option value="user">User</option>
              <option value="queue">Queue</option>
            </select>
          </div>
          <div class="form-group">
            <label>Target {{ assignmentForm.type === 'extension' ? 'Extension' : assignmentForm.type === 'user' ? 'User' : 'Queue' }} ID</label>
            <input 
              v-model="assignmentForm.target_id" 
              type="number" 
              class="input-field" 
              :placeholder="assignmentForm.type === 'extension' ? 'Extension number' : assignmentForm.type === 'user' ? 'User ID' : 'Queue ID'" 
            />
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="closeAssignmentModal">Cancel</button>
          <button class="btn-primary" @click="assignNumber" :disabled="!assignmentForm.target_id || saving">
            {{ saving ? 'Assigning...' : 'Assign' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {
  RefreshCw as RefreshCwIcon,
  MessageSquare as MessageSquareIcon,
  AlertCircle as AlertCircleIcon,
  X as XIcon,
  Check as CheckIcon,
  Settings as SettingsIcon,
  UserPlus as UserPlusIcon
} from 'lucide-vue-next'
import { messagingNumbersAPI } from '../../services/api'

// State
const loading = ref(false)
const saving = ref(false)
const error = ref(null)
const numbers = ref([])
const ivrMenus = ref([])

// Modal state
const showSmsModal = ref(false)
const showAssignmentModal = ref(false)
const selectedNumber = ref(null)

// Forms
const smsForm = ref({
  voice_handler_id: null,
  reply_to_sms_enabled: false,
  conversation_on_reply: false,
  webhook_enabled: false,
  webhook_url: ''
})

const assignmentForm = ref({
  type: 'extension',
  target_id: ''
})

// Load all SMS numbers with their assignments
const loadNumbers = async () => {
  loading.value = true
  error.value = null
  
  try {
    const response = await messagingNumbersAPI.list()
    const data = response.data || []
    
    // Load assignments for each number in parallel
    const numbersWithAssignments = await Promise.all(
      data.map(async (num) => {
        try {
          const assignResponse = await messagingNumbersAPI.listAssignments(num.id)
          return {
            ...num,
            assignments: assignResponse.data || []
          }
        } catch {
          return { ...num, assignments: [] }
        }
      })
    )
    
    numbers.value = numbersWithAssignments
  } catch (e) {
    error.value = 'Failed to load messaging numbers: ' + (e.message || 'Unknown error')
    numbers.value = []
  } finally {
    loading.value = false
  }
}

// Refresh handler
const refreshNumbers = () => loadNumbers()

// SMS Configuration
const openSmsConfig = async (num) => {
  selectedNumber.value = num
  
  // Pre-fill form with existing config or defaults
  smsForm.value = {
    voice_handler_id: num.voice_handler_id || null,
    reply_to_sms_enabled: num.reply_to_sms_enabled ?? true,
    conversation_on_reply: num.conversation_on_reply ?? false,
    webhook_enabled: num.webhook_enabled ?? false,
    webhook_url: num.webhook_url || ''
  }
  
  showSmsModal.value = true
}

const closeSmsModal = () => {
  showSmsModal.value = false
  selectedNumber.value = null
}

const saveSmsConfig = async () => {
  if (!selectedNumber.value) return
  
  saving.value = true
  error.value = null
  
  try {
    await messagingNumbersAPI.configureSms(selectedNumber.value.id, {
      voice_handler_id: smsForm.value.voice_handler_id,
      reply_to_sms_enabled: smsForm.value.reply_to_sms_enabled,
      conversation_on_reply: smsForm.value.conversation_on_reply,
      webhook_enabled: smsForm.value.webhook_enabled,
      webhook_url: smsForm.value.webhook_url
    })
    
    closeSmsModal()
    await loadNumbers()
  } catch (e) {
    error.value = 'Failed to save SMS configuration: ' + (e.message || 'Unknown error')
  } finally {
    saving.value = false
  }
}

// Assignment Management
const openAssignmentModal = (num) => {
  selectedNumber.value = num
  assignmentForm.value = {
    type: 'extension',
    target_id: ''
  }
  showAssignmentModal.value = true
}

const closeAssignmentModal = () => {
  showAssignmentModal.value = false
  selectedNumber.value = null
}

const assignNumber = async () => {
  if (!selectedNumber.value || !assignmentForm.value.target_id) return
  
  saving.value = true
  error.value = null
  
  try {
    await messagingNumbersAPI.assignNumber(selectedNumber.value.id, {
      type: assignmentForm.value.type,
      target_id: parseInt(assignmentForm.value.target_id, 10)
    })
    
    closeAssignmentModal()
    await loadNumbers()
  } catch (e) {
    error.value = 'Failed to assign number: ' + (e.message || 'Unknown error')
  } finally {
    saving.value = false
  }
}

const unassignNumber = async (numberId, assignmentId) => {
  if (!confirm('Unassign this number?')) return
  
  error.value = null
  
  try {
    await messagingNumbersAPI.unassignNumber(numberId, assignmentId)
    await loadNumbers()
  } catch (e) {
    error.value = 'Failed to unassign number: ' + (e.message || 'Unknown error')
  }
}

// Initial load
onMounted(loadNumbers)
</script>

<style scoped>
.messaging-numbers-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

.error-banner { display: flex; align-items: center; gap: 8px; padding: 12px 16px; background: #fef2f2; border: 1px solid #ef4444; border-radius: var(--radius-sm); margin-bottom: 16px; color: #dc2626; font-size: 13px; }
.error-banner .btn-icon { margin-left: auto; }

.panel { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }

.loading-state, .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 60px 20px; text-align: center; gap: 12px; }
.empty-state h3 { margin: 0; font-size: 16px; }
.empty-state p { margin: 0; font-size: 13px; }
.muted { color: var(--text-muted); }
.icon-xl { width: 48px; height: 48px; }
.icon-lg { width: 24px; height: 24px; }

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* Table */
.data-table { width: 100%; border-collapse: collapse; }
.data-table th { text-align: left; padding: 12px 16px; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); background: var(--bg-app); }
.data-table td { padding: 12px 16px; font-size: 13px; border-bottom: 1px solid var(--border-color); vertical-align: middle; }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr:hover td { background: #fafafa; }

.number-cell { display: flex; flex-direction: column; gap: 2px; }
.phone-number { font-weight: 600; font-family: monospace; }
.friendly-name { font-size: 11px; color: var(--text-muted); }

.provider-badge { font-size: 11px; font-weight: 600; padding: 4px 10px; border-radius: 4px; }
.provider-badge.twilio { background: #fee2e2; color: #dc2626; }
.provider-badge.bandwidth { background: #dbeafe; color: #2563eb; }
.provider-badge.telnyx { background: #dcfce7; color: #16a34a; }
.provider-badge.plivo { background: #dcfce7; color: #16a34a; }

.status-toggle { display: inline-flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 4px; }
.status-toggle.enabled { background: #dcfce7; color: #16a34a; }
.status-toggle.disabled { background: #f1f5f9; color: #94a3b8; }
.icon-xs { width: 12px; height: 12px; }

.assignments-cell { display: flex; flex-wrap: wrap; gap: 6px; }
.assignment-tag { display: inline-flex; align-items: center; gap: 4px; background: var(--primary-light); color: var(--primary-color); font-size: 11px; font-weight: 600; padding: 3px 8px; border-radius: 4px; }
.btn-unassign { display: inline-flex; align-items: center; background: none; border: none; cursor: pointer; padding: 0; color: inherit; opacity: 0.6; }
.btn-unassign:hover { opacity: 1; }

.action-buttons { display: flex; gap: 6px; }
.action-btn { width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }

.icon-sm { width: 16px; height: 16px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 480px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.form-group:last-child { margin-bottom: 0; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.help-text { font-size: 11px; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; text-transform: none; font-weight: 500; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { width: 16px; height: 16px; }
</style>
