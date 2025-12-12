<template>
  <div class="feature-codes-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Feature Codes</h2>
        <p class="text-muted text-sm">Configure feature codes for call forwarding, voicemail, parking, and more.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="refreshData" :disabled="loading">
          <RefreshCwIcon class="btn-icon" :class="{ 'spin': loading }" /> Refresh
        </button>
        <button class="btn-primary" @click="openCreateModal">
          <PlusIcon class="btn-icon" /> Add Feature Code
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ featureCodes.length }}</div>
        <div class="stat-label">Total Codes</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ customCodes.length }}</div>
        <div class="stat-label">Custom</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ enabledCodes.length }}</div>
        <div class="stat-label">Enabled</div>
      </div>
    </div>

    <!-- Filter & Search -->
    <div class="toolbar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" placeholder="Search codes..." class="search-input" />
      </div>
      <select v-model="filterAction" class="filter-select">
        <option value="">All Actions</option>
        <option v-for="action in actionTypes" :key="action.value" :value="action.value">
          {{ action.label }}
        </option>
      </select>
    </div>

    <!-- Feature Codes Table -->
    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>Code</th>
            <th>Name</th>
            <th>Action</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="code in filteredCodes" :key="code.id" :class="{ disabled: !code.enabled }">
            <td>
              <code class="code-badge">{{ code.code }}</code>
              <span v-if="code.code_regex" class="regex-hint" :title="code.code_regex">
                (regex)
              </span>
            </td>
            <td>{{ code.name }}</td>
            <td>
              <span class="action-badge" :class="getActionClass(code.action)">
                <component :is="getActionIcon(code.action)" class="action-icon" />
                {{ formatAction(code.action) }}
              </span>
            </td>
            <td>
              <span class="status-badge" :class="code.enabled ? 'enabled' : 'disabled'">
                {{ code.enabled ? 'Enabled' : 'Disabled' }}
              </span>
            </td>
            <td class="actions-cell">
              <button class="btn-icon" @click="openEditModal(code)" title="Edit">
                <EditIcon />
              </button>
              <button class="btn-icon" @click="toggleEnabled(code)" :title="code.enabled ? 'Disable' : 'Enable'">
                <ToggleRightIcon v-if="code.enabled" class="text-good" />
                <ToggleLeftIcon v-else class="text-muted" />
              </button>
              <button class="btn-icon danger" @click="confirmDelete(code)" title="Delete">
                <TrashIcon />
              </button>
            </td>
          </tr>
          <tr v-if="filteredCodes.length === 0">
            <td colspan="5" class="empty-row">
              <PhoneOffIcon class="empty-icon" />
              <p>No feature codes found</p>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card modal-lg">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Feature Code' : 'Create Feature Code' }}</h3>
          <button class="close-btn" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group">
              <label>Feature Code *</label>
              <input v-model="form.code" class="input-field" placeholder="*72" />
              <span class="help-text">E.g., *72, *98, #123</span>
            </div>
            <div class="form-group">
              <label>Regex Pattern (optional)</label>
              <input v-model="form.code_regex" class="input-field mono" placeholder="^\*72(\d+)$" />
              <span class="help-text">For codes with variable parts</span>
            </div>
          </div>

          <div class="form-group">
            <label>Name *</label>
            <input v-model="form.name" class="input-field" placeholder="Call Forward Enable" />
          </div>

          <div class="form-group">
            <label>Description</label>
            <input v-model="form.description" class="input-field" placeholder="Enable call forwarding to a destination" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Action *</label>
              <select v-model="form.action" class="input-field">
                <option v-for="action in actionTypes" :key="action.value" :value="action.value">
                  {{ action.label }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>Order</label>
              <input v-model.number="form.order" type="number" class="input-field" placeholder="100" />
            </div>
          </div>

          <!-- Action-specific fields -->
          <div v-if="showActionData" class="form-group">
            <label>Action Data</label>
            <input v-model="form.action_data" class="input-field" :placeholder="getActionDataPlaceholder()" />
            <span class="help-text">{{ getActionDataHelp() }}</span>
          </div>

          <!-- Park-specific fields -->
          <div v-if="form.action === 'park' || form.action === 'park_slot'" class="form-row">
            <div class="form-group">
              <label>Park Lot Name</label>
              <input v-model="form.park_lot_name" class="input-field" placeholder="default" />
            </div>
            <div class="form-group">
              <label>Park Timeout (seconds)</label>
              <input v-model.number="form.park_timeout" type="number" class="input-field" placeholder="120" />
            </div>
          </div>

          <!-- Webhook fields -->
          <div v-if="form.action === 'webhook'" class="form-row">
            <div class="form-group flex-2">
              <label>Webhook URL</label>
              <input v-model="form.webhook_url" class="input-field" placeholder="https://example.com/hook" />
            </div>
            <div class="form-group">
              <label>Method</label>
              <select v-model="form.webhook_method" class="input-field">
                <option value="POST">POST</option>
                <option value="GET">GET</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group checkbox-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="form.enabled" />
                <span>Enabled</span>
              </label>
            </div>
            <div class="form-group checkbox-group" v-if="form.action === 'park'">
              <label class="checkbox-label">
                <input type="checkbox" v-model="form.park_announce" />
                <span>Announce Slot Number</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveCode" :disabled="saving">
            {{ saving ? 'Saving...' : (isEditing ? 'Update' : 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-card modal-sm">
        <div class="modal-header danger">
          <h3>Delete Feature Code</h3>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete <strong>{{ deleteTarget?.name }}</strong> ({{ deleteTarget?.code }})?</p>
          <p class="text-muted">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger" @click="deleteCode" :disabled="deleting">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  Plus as PlusIcon, RefreshCw as RefreshCwIcon, Search as SearchIcon,
  Edit as EditIcon, Trash as TrashIcon, ToggleLeft as ToggleLeftIcon,
  ToggleRight as ToggleRightIcon, PhoneOff as PhoneOffIcon,
  Phone, Voicemail, BellOff, ArrowRightLeft, ParkingCircle, PhoneIncoming,
  Mic, Users, Webhook, Code, Settings
} from 'lucide-vue-next'
import { featureCodesAPI } from '../../services/api'

// Simple toast-like notification helper
const notify = {
  success: (msg) => console.log('✅', msg),
  error: (msg) => { console.error('❌', msg); alert(msg) }
}

const featureCodes = ref([])
const loading = ref(false)
const searchQuery = ref('')
const filterAction = ref('')
const showModal = ref(false)
const showDeleteModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleting = ref(false)
const deleteTarget = ref(null)

const defaultForm = {
  code: '',
  code_regex: '',
  name: '',
  description: '',
  action: 'voicemail',
  action_data: '',
  order: 100,
  enabled: true,
  is_global: true,
  park_lot_name: 'default',
  park_timeout: 120,
  park_announce: true,
  webhook_url: '',
  webhook_method: 'POST'
}

const form = ref({ ...defaultForm })
const editingId = ref(null)

const actionTypes = [
  { value: 'voicemail', label: 'Voicemail', icon: 'Voicemail' },
  { value: 'call_forward', label: 'Call Forward', icon: 'ArrowRightLeft' },
  { value: 'dnd', label: 'Do Not Disturb', icon: 'BellOff' },
  { value: 'call_flow_toggle', label: 'Call Flow Toggle', icon: 'Settings' },
  { value: 'park', label: 'Valet Park (Auto)', icon: 'ParkingCircle' },
  { value: 'park_slot', label: 'Park to Slot', icon: 'ParkingCircle' },
  { value: 'park_retrieve', label: 'Retrieve from Slot', icon: 'ParkingCircle' },
  { value: 'pickup', label: 'Call Pickup', icon: 'PhoneIncoming' },
  { value: 'intercom', label: 'Intercom', icon: 'Mic' },
  { value: 'page_group', label: 'Page Group', icon: 'Users' },
  { value: 'transfer', label: 'Transfer', icon: 'Phone' },
  { value: 'record', label: 'Recording', icon: 'Mic' },
  { value: 'webhook', label: 'Webhook', icon: 'Webhook' },
  { value: 'lua', label: 'Lua Script', icon: 'Code' },
  { value: 'custom', label: 'Custom', icon: 'Settings' }
]

const customCodes = computed(() => featureCodes.value.filter(c => c.action === 'custom' || c.action === 'webhook'))
const enabledCodes = computed(() => featureCodes.value.filter(c => c.enabled))

const filteredCodes = computed(() => {
  let codes = featureCodes.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    codes = codes.filter(c => 
      c.code.toLowerCase().includes(q) ||
      c.name.toLowerCase().includes(q) ||
      c.action.toLowerCase().includes(q)
    )
  }
  if (filterAction.value) {
    codes = codes.filter(c => c.action === filterAction.value)
  }
  return codes.sort((a, b) => a.order - b.order)
})

const showActionData = computed(() => {
  return ['call_forward', 'transfer', 'lua', 'custom', 'pickup'].includes(form.value.action)
})

const loadData = async () => {
  loading.value = true
  try {
    const response = await featureCodesAPI.list()
    featureCodes.value = response.data.data || response.data || []
  } catch (error) {
    notify.error('Failed to load feature codes')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const refreshData = () => loadData()

const formatAction = (action) => {
  const type = actionTypes.find(t => t.value === action)
  return type?.label || action
}

const getActionClass = (action) => {
  const classes = {
    voicemail: 'action-voicemail',
    call_forward: 'action-forward',
    dnd: 'action-dnd',
    park: 'action-park',
    park_slot: 'action-park',
    park_retrieve: 'action-park',
    pickup: 'action-pickup',
    intercom: 'action-intercom'
  }
  return classes[action] || 'action-default'
}

const getActionIcon = (action) => {
  const icons = {
    voicemail: Voicemail,
    call_forward: ArrowRightLeft,
    dnd: BellOff,
    call_flow_toggle: Settings,
    park: ParkingCircle,
    park_slot: ParkingCircle,
    park_retrieve: ParkingCircle,
    pickup: PhoneIncoming,
    intercom: Mic,
    page_group: Users,
    transfer: Phone,
    record: Mic,
    webhook: Webhook,
    lua: Code,
    custom: Settings
  }
  return icons[action] || Phone
}

const getActionDataPlaceholder = () => {
  const placeholders = {
    call_forward: 'enable or disable',
    transfer: '1001 XML default',
    lua: 'script_name.lua',
    custom: 'action1,action2',
    pickup: 'group or directed'
  }
  return placeholders[form.value.action] || ''
}

const getActionDataHelp = () => {
  const help = {
    call_forward: 'Set to "enable" or "disable" for toggle behavior',
    transfer: 'Extension Context Profile format',
    lua: 'Name of Lua script to execute',
    pickup: '"group" for group pickup, or leave empty for directed'
  }
  return help[form.value.action] || ''
}

const openCreateModal = () => {
  form.value = { ...defaultForm }
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

const openEditModal = (code) => {
  form.value = { ...code }
  isEditing.value = true
  editingId.value = code.id
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  form.value = { ...defaultForm }
}

const saveCode = async () => {
  if (!form.value.code || !form.value.name || !form.value.action) {
    notify.error('Please fill in required fields')
    return
  }

  saving.value = true
  try {
    if (isEditing.value) {
      await featureCodesAPI.update(editingId.value, form.value)
      notify.success('Feature code updated')
    } else {
      await featureCodesAPI.create(form.value)
      notify.success('Feature code created')
    }
    closeModal()
    loadData()
  } catch (error) {
    notify.error(error.message || 'Failed to save feature code')
  } finally {
    saving.value = false
  }
}

const toggleEnabled = async (code) => {
  try {
    await featureCodesAPI.update(code.id, { ...code, enabled: !code.enabled })
    code.enabled = !code.enabled
    notify.success(`Feature code ${code.enabled ? 'enabled' : 'disabled'}`)
  } catch (error) {
    notify.error('Failed to update feature code')
  }
}

const confirmDelete = (code) => {
  deleteTarget.value = code
  showDeleteModal.value = true
}

const deleteCode = async () => {
  deleting.value = true
  try {
    await featureCodesAPI.delete(deleteTarget.value.id)
    notify.success('Feature code deleted')
    showDeleteModal.value = false
    loadData()
  } catch (error) {
    notify.error('Failed to delete feature code')
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.feature-codes-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary, .btn-danger { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-danger { background: #dc2626; color: white; }
.btn-icon { width: 14px; height: 14px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-box { position: relative; flex: 1; max-width: 300px; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 10px 8px 36px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; }
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; min-width: 150px; }

.table-container { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { text-align: left; padding: 12px 16px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); background: #f8fafc; }
.data-table td { padding: 12px 16px; border-bottom: 1px solid var(--border-color); }
.data-table tr:hover { background: #f8fafc; }
.data-table tr.disabled { opacity: 0.6; }

.code-badge { background: #1e293b; color: #fff; padding: 4px 10px; border-radius: 4px; font-family: monospace; font-size: 13px; font-weight: 600; }
.regex-hint { font-size: 10px; color: var(--text-muted); margin-left: 6px; }

.action-badge { display: inline-flex; align-items: center; gap: 6px; padding: 4px 10px; border-radius: 4px; font-size: 11px; font-weight: 600; }
.action-icon { width: 12px; height: 12px; }
.action-voicemail { background: #dbeafe; color: #1d4ed8; }
.action-forward { background: #dcfce7; color: #16a34a; }
.action-dnd { background: #fee2e2; color: #dc2626; }
.action-park { background: #fef3c7; color: #d97706; }
.action-pickup { background: #e9d5ff; color: #9333ea; }
.action-intercom { background: #cffafe; color: #0891b2; }
.action-default { background: #f1f5f9; color: #475569; }

.scope-badge { padding: 4px 8px; border-radius: 4px; font-size: 10px; font-weight: 700; text-transform: uppercase; }
.scope-badge.global { background: #dbeafe; color: #1d4ed8; }
.scope-badge.tenant { background: #f1f5f9; color: #475569; }

.status-badge { padding: 4px 8px; border-radius: 4px; font-size: 10px; font-weight: 700; }
.status-badge.enabled { background: #dcfce7; color: #16a34a; }
.status-badge.disabled { background: #f1f5f9; color: #94a3b8; }

.actions-cell { display: flex; gap: 4px; }
.actions-cell .btn-icon { width: 32px; height: 32px; background: white; border: 1px solid var(--border-color); border-radius: 6px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); transition: all 0.15s ease; }
.actions-cell .btn-icon:hover { color: var(--primary-color); border-color: var(--primary-color); background: #f0f9ff; }
.actions-cell .btn-icon.danger:hover { color: #dc2626; border-color: #dc2626; background: #fef2f2; }
.actions-cell .btn-icon svg { width: 16px; height: 16px; }
.text-good { color: #16a34a; }
.text-muted { color: #94a3b8; }

.empty-row { text-align: center; padding: 40px !important; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; margin-bottom: 12px; opacity: 0.5; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-card.modal-lg { max-width: 640px; }
.modal-card.modal-sm { max-width: 400px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header.danger { background: #fef2f2; border-bottom-color: #fecaca; }
.modal-header.danger h3 { color: #dc2626; }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; max-height: 60vh; overflow-y: auto; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.form-row .form-group.flex-2 { flex: 2; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.input-field.mono { font-family: monospace; }
.help-text { font-size: 11px; color: var(--text-muted); margin-top: 4px; display: block; }
.checkbox-group { display: flex; align-items: center; }
.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 13px; }
.checkbox-label input { width: 16px; height: 16px; }
</style>
