<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Trunks</h2>
      <p class="text-muted text-sm">Manage SIP trunks and PSTN gateways. Configure dial formats for outbound calling.</p>
    </div>
    <div class="header-actions">
      <span v-if="eslConnected" class="status-pill connected">ESL Connected</span>
      <span v-else class="status-pill disconnected">ESL Offline</span>
      <button class="btn-secondary" @click="refreshStatus" :disabled="refreshing">
        <RefreshCwIcon class="btn-icon" :class="{ spinning: refreshing }" />
        Refresh
      </button>
      <button class="btn-secondary" v-if="orderChanged" @click="saveOrder">
        Save Order
      </button>
      <button class="btn-primary" @click="showModal = true">+ Add Trunk</button>
    </div>
  </div>

  <div class="gw-list">
    <div class="gw-list-header">
      <span class="gw-col handle-col"></span>
      <span class="gw-col name-col">Gateway Name</span>
      <span class="gw-col host-col">Hostname / IP</span>
      <span class="gw-col profile-col">Profile</span>
      <span class="gw-col type-col">Type</span>
      <span class="gw-col priority-col">Priority</span>
      <span class="gw-col status-col">Status</span>
      <span class="gw-col actions-col">Actions</span>
    </div>
    <div 
      v-for="(gw, idx) in gateways" :key="gw.id"
      class="gw-row"
      :class="{ dragging: dragIndex === idx, dragover: dragOverIndex === idx }"
      draggable="true"
      @dragstart="onDragStart($event, idx)"
      @dragover.prevent="onDragOver($event, idx)"
      @dragleave="onDragLeave"
      @drop.prevent="onDrop($event, idx)"
      @dragend="onDragEnd"
    >
      <span class="gw-col handle-col">
        <GripVerticalIcon class="grip-icon" />
      </span>
      <span class="gw-col name-col font-semibold">{{ gw.name }}</span>
      <span class="gw-col host-col text-muted">{{ gw.hostname }}</span>
      <span class="gw-col profile-col">
        <span class="badge profile">{{ gw.profile_name || 'external' }}</span>
      </span>
      <span class="gw-col type-col">
        <span class="badge" :class="(gw.type || '').toLowerCase()">{{ gw.type }}</span>
      </span>
      <span class="gw-col priority-col">
        <span class="priority-pill">P{{ gw.priority || 0 }}</span>
        <span class="weight-pill">W{{ gw.weight || 100 }}</span>
      </span>
      <span class="gw-col status-col">
        <StatusBadge :status="gw.status" />
      </span>
      <span class="gw-col actions-col">
        <button class="btn-link" @click="editGateway(gw)">Edit</button>
        <button class="btn-link" @click="restartGateway(gw)">Restart</button>
        <button class="btn-link text-bad" @click="deleteGateway(gw)">Delete</button>
      </span>
    </div>
  </div>

  <!-- Gateway Modal -->
  <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ isEditing ? 'Edit Trunk' : 'Add Trunk' }}</h3>
        <button class="btn-icon" @click="showModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-section">
          <h4>Gateway Connectivity</h4>
          <div class="form-group">
            <label>Gateway Name</label>
            <input type="text" v-model="form.name" class="input-field" placeholder="e.g. Flowroute Primary">
          </div>
          
          <div class="form-group">
            <label>Proxy Address / Domain</label>
            <input type="text" v-model="form.hostname" class="input-field" placeholder="sip.provider.com">
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Protocol</label>
              <select v-model="form.protocol" class="input-field">
                <option value="udp">UDP</option>
                <option value="tcp">TCP</option>
                <option value="tls">TLS</option>
              </select>
            </div>
            <div class="form-group">
              <label>Port</label>
              <input type="text" v-model="form.port" class="input-field" placeholder="5060">
            </div>
          </div>

          <div class="form-group">
            <label>Type</label>
            <select v-model="form.type" class="input-field">
              <option value="Public">Public (Cloud Provider)</option>
              <option value="Local">Local (On-Premise PRI/FXO)</option>
              <option value="Peering">Peering (Direct Interconnect)</option>
            </select>
          </div>

          <div class="form-group">
            <label>SIP Profile</label>
            <select v-model="form.profile_name" class="input-field">
              <option v-for="p in sipProfiles" :key="p.profile_name" :value="p.profile_name">
                {{ p.profile_name }}{{ p.description ? ` — ${p.description}` : '' }}
              </option>
            </select>
            <p class="help-text">Which SIP profile (interface) this trunk registers on. Trunks typically use the <strong>external</strong> profile.</p>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Authentication</h4>
          <div class="form-row">
            <div class="form-check">
              <input type="checkbox" id="registerCheck" v-model="form.register">
              <label for="registerCheck" class="check-label">Register with provider</label>
            </div>
          </div>
          
          <div v-if="form.register" class="auth-fields">
            <div class="form-group">
              <label>Username</label>
              <input type="text" v-model="form.username" class="input-field">
            </div>
            <div class="form-group">
              <label>Password</label>
              <input type="password" v-model="form.password" class="input-field">
            </div>
            <div class="form-group">
              <label>Realm (Optional)</label>
              <input type="text" v-model="form.realm" class="input-field" placeholder="Leave blank to auto-detect">
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Tenant Access</h4>
          <div class="form-group">
            <label>Available To</label>
            <select v-model="form.access" class="input-field">
              <option value="all">All Tenants</option>
              <option value="selected">Selected Tenants Only</option>
              <option value="none">System Use Only</option>
            </select>
          </div>
          <p class="help-text">Control which tenants can route calls through this trunk.</p>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Routing Priority</h4>
          <p class="help-text" style="margin-bottom: 12px;">Dial format and number transformations are now configured on <strong>Number Group routing rules</strong> in System Routing.</p>

          <div class="form-row">
            <div class="form-group">
              <label>Priority (LCR)</label>
              <input type="number" v-model="form.priority" class="input-field" placeholder="0">
            </div>
            <div class="form-group">
              <label>Weight</label>
              <input type="number" v-model="form.weight" class="input-field" placeholder="100">
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveGateway" :disabled="!form.name || !form.hostname">
          {{ isEditing ? 'Save Changes' : 'Add Gateway' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, inject } from 'vue'
import { X as XIcon, RefreshCw as RefreshCwIcon, GripVertical as GripVerticalIcon } from 'lucide-vue-next'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api'

const toast = inject('toast')

const gateways = ref([])
const loading = ref(true)
const refreshing = ref(false)
const eslConnected = ref(false)
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const restarting = ref(null)
let statusInterval = null

const defaultForm = () => ({
  id: null,
  name: '',
  hostname: '',
  protocol: 'udp',
  port: 5060,
  type: 'Public',
  profile_name: 'external',
  register: false,
  username: '',
  password: '',
  realm: '',
  enabled: true,
  access: 'all',
  priority: 0,
  weight: 100
})

const sipProfiles = ref([])

const form = ref(defaultForm())

const loadGateways = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listGateways()
    const data = response.data || []
    gateways.value = data.map(g => ({
      id: g.id,
      name: g.gateway_name || g.name || '',
      hostname: g.proxy || g.hostname || '',
      type: g.gateway_type || 'Public',
      profile_name: g.profile_name || 'external',
      tenants: g.tenant_count || 0,
      status: g.enabled ? (g.register ? 'Registered' : 'Active') : 'Disabled',
      protocol: g.transport || g.protocol || 'udp',
      port: g.port || 5060,
      register: g.register || false,
      username: g.username || '',
      realm: g.realm || '',
      priority: g.priority || 0,
      weight: g.weight || 100
    })).sort((a, b) => a.priority - b.priority)
    originalOrder.value = gateways.value.map(g => ({ id: g.id, priority: g.priority, weight: g.weight }))
  } catch (e) {
    console.error('Failed to load gateways:', e)
  } finally {
    loading.value = false
  }
}

// Load live status from FreeSWITCH
const loadGatewayStatus = async () => {
  try {
    const response = await systemAPI.getGatewayStatus()
    eslConnected.value = true
    const statusData = response.data?.data || {}
    
    // Update gateway statuses
    gateways.value.forEach(gw => {
      const status = statusData[gw.name] || statusData[gw.id]
      if (status) {
        gw.liveStatus = status.state // REGED, NOREG, TRYING, etc.
        gw.status = status.state === 'REGED' ? 'Registered' : 
                    status.state === 'NOREG' ? 'Not Registered' :
                    status.state === 'TRYING' ? 'Connecting' : gw.status
      }
    })
  } catch (e) {
    eslConnected.value = false
    console.log('Could not fetch live gateway status')
  }
}

const refreshStatus = async () => {
  refreshing.value = true
  await loadGateways()
  await loadGatewayStatus()
  refreshing.value = false
}

const loadSipProfiles = async () => {
  try {
    const response = await systemAPI.listSIPProfiles()
    sipProfiles.value = (response.data?.data || response.data || []).filter(p => p.enabled !== false)
  } catch (e) {
    // Fallback to common defaults
    sipProfiles.value = [
      { profile_name: 'external', description: 'External/Trunk interface' },
      { profile_name: 'internal', description: 'Internal/Registration interface' }
    ]
  }
}

onMounted(async () => {
  await loadSipProfiles()
  await loadGateways()
  await loadGatewayStatus()
  // Auto-refresh every 30 seconds
  statusInterval = setInterval(loadGatewayStatus, 30000)
})

onUnmounted(() => {
  if (statusInterval) clearInterval(statusInterval)
})

const editGateway = (gw) => {
  form.value = { ...gw, password: '' }
  isEditing.value = true
  showModal.value = true
}

const saveGateway = async () => {
  saving.value = true
  try {
    const data = {
      gateway_name: form.value.name,
      proxy: form.value.hostname,
      transport: form.value.protocol,
      gateway_type: form.value.type,
      profile_name: form.value.profile_name || 'external',
      register: form.value.register,
      username: form.value.username,
      password: form.value.password,
      realm: form.value.realm,
      enabled: form.value.enabled,
      priority: parseInt(form.value.priority) || 0,
      weight: parseInt(form.value.weight) || 100
    }
    if (isEditing.value && form.value.id) {
      await systemAPI.updateGateway(form.value.id, data)
    } else {
      await systemAPI.createGateway(data)
    }
    await loadGateways()
    showModal.value = false
    isEditing.value = false
    form.value = defaultForm()
  } catch (e) {
    alert('Failed to save gateway: ' + (e.response?.data?.error || e.message))
  } finally {
    saving.value = false
  }
}

const deleteGateway = async (gw) => {
  if (!confirm(`Delete gateway "${gw.name}"?`)) return
  try {
    await systemAPI.deleteGateway(gw.id)
    await loadGateways()
  } catch (e) {
    alert('Failed to delete gateway: ' + e.message)
  }
}

const restartGateway = async (gw) => {
  if (!confirm(`Restart gateway "${gw.name}"? This will temporarily interrupt connections.`)) return
  restarting.value = gw.id
  try {
    await systemAPI.restartGateway(gw.id)
    toast?.success(`Restarting gateway: ${gw.name}`)
    await loadGatewayStatus()
  } catch (e) {
    console.error('Failed to restart gateway:', e)
    toast?.error(e.message || 'Failed to restart gateway')
  } finally {
    restarting.value = null
  }
}

// Drag-to-reorder
const dragIndex = ref(null)
const dragOverIndex = ref(null)
const orderChanged = ref(false)
const originalOrder = ref([])

const onDragStart = (e, idx) => {
  dragIndex.value = idx
  e.dataTransfer.effectAllowed = 'move'
}

const onDragOver = (e, idx) => {
  dragOverIndex.value = idx
  e.dataTransfer.dropEffect = 'move'
}

const onDragLeave = () => {
  dragOverIndex.value = null
}

const onDrop = (e, idx) => {
  if (dragIndex.value === null || dragIndex.value === idx) return
  const item = gateways.value.splice(dragIndex.value, 1)[0]
  gateways.value.splice(idx, 0, item)
  // Recalculate priorities based on new order
  gateways.value.forEach((gw, i) => {
    gw.priority = (i + 1) * 10
  })
  orderChanged.value = true
  dragIndex.value = null
  dragOverIndex.value = null
}

const onDragEnd = () => {
  dragIndex.value = null
  dragOverIndex.value = null
}

const saveOrder = async () => {
  try {
    const items = gateways.value.map(gw => ({
      id: gw.id,
      priority: gw.priority,
      weight: gw.weight
    }))
    await systemAPI.reorderGateways(items)
    orderChanged.value = false
    originalOrder.value = items.map(i => ({ ...i }))
  } catch (e) {
    alert('Failed to save gateway order: ' + (e.message || 'Unknown error'))
  }
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
  cursor: pointer;
}
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  cursor: pointer;
  font-weight: 500;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
}

.text-bad { color: var(--status-bad); }

.badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; background: var(--bg-secondary); }
.badge.public { background: #e0f2fe; color: #0369a1; }
.badge.local { background: #f3e8ff; color: #7e22ce; }
.badge.peering { background: #fef3c7; color: #92400e; }
.badge.profile { background: #ecfdf5; color: #065f46; }

.tenant-badge {
  font-size: 11px;
  color: var(--text-muted);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  padding: 24px;
}

.modal-card {
  background: white;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 700;
  margin: 0;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.form-section {
  margin-bottom: 8px;
}

.form-section h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 14px;
}
.input-field:focus {
  outline: none;
  border-color: var(--primary-color);
}

.divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
}

.form-check {
  display: flex;
  align-items: center;
  gap: 8px;
}

.check-label {
  font-size: 14px;
  font-weight: 500;
  text-transform: none;
  color: var(--text-main);
}

.auth-fields {
  margin-top: 12px;
}

.help-text {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
}

.icon-sm { width: 16px; height: 16px; }

/* Status and Actions */
.status-pill { font-size: 11px; padding: 4px 10px; border-radius: 99px; font-weight: 600; }
.status-pill.connected { background: #dcfce7; color: #16a34a; }
.status-pill.disconnected { background: #fef2f2; color: #dc2626; }

.btn-icon { width: 14px; height: 14px; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* Gateway draggable list */
.gw-list { border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.gw-list-header {
  display: grid;
  grid-template-columns: 36px 1.5fr 1.5fr 90px 100px 120px 120px 180px;
  gap: 8px;
  padding: 10px 16px;
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}
.gw-row {
  display: grid;
  grid-template-columns: 36px 1.5fr 1.5fr 90px 100px 120px 120px 180px;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
  align-items: center;
  background: white;
  cursor: grab;
  transition: all 0.15s ease;
}
.gw-row:last-child { border-bottom: none; }
.gw-row:hover { background: #f8fafc; }
.gw-row.dragging { opacity: 0.4; background: #e0e7ff; border-style: dashed; }
.gw-row.dragover { border-top: 2px solid var(--primary-color); background: #eff6ff; }
.gw-row:active { cursor: grabbing; }

.gw-col { font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.handle-col { display: flex; align-items: center; justify-content: center; }
.grip-icon { width: 16px; height: 16px; color: var(--text-muted); }

.priority-pill, .weight-pill {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
  margin-right: 4px;
}
.priority-pill { background: #dbeafe; color: #1d4ed8; }
.weight-pill { background: #f3e8ff; color: #7c3aed; }

.font-semibold { font-weight: 600; }
.text-muted { color: var(--text-muted); }

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  text-transform: none;
  color: var(--text-main);
  cursor: pointer;
}
</style>
