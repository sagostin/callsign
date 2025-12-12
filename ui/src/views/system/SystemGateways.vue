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
      <button class="btn-primary" @click="showModal = true">+ Add Trunk</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="gateways" actions>
    <template #status="{ value }">
      <StatusBadge :status="value" />
    </template>
    
    <template #type="{ value }">
      <span class="badge" :class="value.toLowerCase()">{{ value }}</span>
    </template>

    <template #tenants="{ value }">
      <span class="tenant-badge">{{ value }} tenants</span>
    </template>

    <template #actions="{ row }">
      <button class="btn-link" @click="editGateway(row)">Edit</button>
      <button class="btn-link" @click="restartGateway(row)">Restart</button>
      <button class="btn-link text-bad" @click="deleteGateway(row)">Delete</button>
    </template>
  </DataTable>

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
          <h4>Dial Format</h4>
          <p class="help-text" style="margin-bottom: 12px;">Configure how numbers are formatted when sent to this trunk.</p>
          
          <div class="form-row">
            <label class="checkbox-label">
              <input type="checkbox" v-model="form.allow_10_digit">
              <span>Allow 10-digit dialing (NANPA)</span>
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="form.allow_11_digit">
              <span>Allow 11-digit dialing (1+10)</span>
            </label>
          </div>

          <div class="form-row" style="margin-top: 12px;">
            <div class="form-group">
              <label>International Prefix</label>
              <input type="text" v-model="form.international_prefix" class="input-field" placeholder="011">
            </div>
            <div class="form-group">
              <label>Outbound Format</label>
              <select v-model="form.dial_format" class="input-field">
                <option value="e164">E.164 (+1XXXXXXXXXX)</option>
                <option value="11d">11-digit (1XXXXXXXXXX)</option>
                <option value="10d">10-digit (XXXXXXXXXX)</option>
                <option value="custom">Custom</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Strip Prefix</label>
              <input type="text" v-model="form.strip_prefix" class="input-field" placeholder="None">
            </div>
            <div class="form-group">
              <label>Prepend Prefix</label>
              <input type="text" v-model="form.prepend_prefix" class="input-field" placeholder="None">
            </div>
          </div>

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
import { ref, onMounted, onUnmounted } from 'vue'
import { X as XIcon, RefreshCw as RefreshCwIcon } from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { systemAPI } from '../../services/api'

const columns = [
  { key: 'name', label: 'Gateway Name' },
  { key: 'hostname', label: 'Hostname / IP' },
  { key: 'type', label: 'Type', width: '120px' },
  { key: 'tenants', label: 'Usage', width: '100px' },
  { key: 'status', label: 'Status', width: '120px' }
]

const gateways = ref([])
const loading = ref(true)
const refreshing = ref(false)
const eslConnected = ref(false)
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
let statusInterval = null

const defaultForm = () => ({
  id: null,
  name: '',
  hostname: '',
  protocol: 'udp',
  port: 5060,
  type: 'Public',
  register: false,
  username: '',
  password: '',
  realm: '',
  enabled: true,
  access: 'all',
  // Dial format fields
  dial_format: 'e164',
  allow_10_digit: true,
  allow_11_digit: true,
  international_prefix: '011',
  strip_prefix: '',
  prepend_prefix: '',
  priority: 0,
  weight: 100
})

const form = ref(defaultForm())

const loadGateways = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listGateways()
    const data = response.data.data || response.data || []
    gateways.value = data.map(g => ({
      id: g.id,
      name: g.name,
      hostname: g.proxy || g.hostname,
      type: g.gateway_type || 'Public',
      tenants: g.tenant_count || 0,
      status: g.enabled ? (g.register ? 'Registered' : 'Active') : 'Disabled',
      protocol: g.protocol || 'udp',
      port: g.port || 5060,
      register: g.register || false,
      username: g.username || '',
      realm: g.realm || '',
      enabled: g.enabled !== false
    }))
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

onMounted(async () => {
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
      register: form.value.register,
      username: form.value.username,
      password: form.value.password,
      realm: form.value.realm,
      enabled: form.value.enabled,
      // Dial format fields
      dial_format: form.value.dial_format,
      allow_10_digit: form.value.allow_10_digit,
      allow_11_digit: form.value.allow_11_digit,
      international_prefix: form.value.international_prefix,
      strip_prefix: form.value.strip_prefix,
      prepend_prefix: form.value.prepend_prefix,
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

const restartGateway = (gw) => {
  alert(`Restarting gateway: ${gw.name} - Not implemented yet`)
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
</style>
