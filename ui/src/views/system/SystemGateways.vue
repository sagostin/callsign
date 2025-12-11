<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Gateways</h2>
      <p class="text-muted text-sm">Manage global SIP trunks and PSTN gateways available to all tenants.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="showModal = true">+ Add Gateway</button>
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
        <h3>{{ isEditing ? 'Edit Gateway' : 'Add System Gateway' }}</h3>
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
          <p class="help-text">Control which tenants can route calls through this gateway.</p>
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
import { ref, onMounted } from 'vue'
import { X as XIcon } from 'lucide-vue-next'
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
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)

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
  enabled: true
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

onMounted(loadGateways)

const editGateway = (gw) => {
  form.value = { ...gw, password: '' }
  isEditing.value = true
  showModal.value = true
}

const saveGateway = async () => {
  saving.value = true
  try {
    const data = {
      name: form.value.name,
      proxy: form.value.hostname,
      protocol: form.value.protocol,
      port: parseInt(form.value.port) || 5060,
      gateway_type: form.value.type,
      register: form.value.register,
      username: form.value.username,
      password: form.value.password,
      realm: form.value.realm,
      enabled: form.value.enabled
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
</style>
