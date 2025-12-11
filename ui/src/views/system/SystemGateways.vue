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
import { ref } from 'vue'
import { X as XIcon } from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'

const columns = [
  { key: 'name', label: 'Gateway Name' },
  { key: 'hostname', label: 'Hostname / IP' },
  { key: 'type', label: 'Type', width: '120px' },
  { key: 'tenants', label: 'Usage', width: '100px' },
  { key: 'status', label: 'Status', width: '120px' }
]

const gateways = ref([
  { id: 1, name: 'Flowroute Primary', hostname: 'sip.flowroute.com', type: 'Public', tenants: 8, status: 'Registered', protocol: 'udp', port: '5060', register: true, username: 'user1', access: 'all' },
  { id: 2, name: 'Twilio Elastic SIP', hostname: 'callsign.pstn.twilio.com', type: 'Public', tenants: 5, status: 'Registered', protocol: 'tls', port: '5061', register: true, username: 'twilio_user', access: 'all' },
  { id: 3, name: 'Telnyx Backup', hostname: 'sip.telnyx.com', type: 'Public', tenants: 3, status: 'Registered', protocol: 'udp', port: '5060', register: true, username: 'telnyx_acct', access: 'selected' },
  { id: 4, name: 'Local PRI Gateway', hostname: '192.168.1.200', type: 'Local', tenants: 2, status: 'Error', protocol: 'udp', port: '5060', register: false, access: 'all' },
])

const showModal = ref(false)
const isEditing = ref(false)
const form = ref({
  id: null,
  name: '',
  hostname: '',
  protocol: 'udp',
  port: '5060',
  type: 'Public',
  register: false,
  username: '',
  password: '',
  realm: '',
  access: 'all'
})

const resetForm = () => {
  form.value = {
    id: null,
    name: '',
    hostname: '',
    protocol: 'udp',
    port: '5060',
    type: 'Public',
    register: false,
    username: '',
    password: '',
    realm: '',
    access: 'all'
  }
}

const editGateway = (gw) => {
  form.value = { ...gw, password: '' }
  isEditing.value = true
  showModal.value = true
}

const saveGateway = () => {
  if (isEditing.value) {
    const idx = gateways.value.findIndex(g => g.id === form.value.id)
    if (idx !== -1) {
      gateways.value[idx] = { ...form.value, status: gateways.value[idx].status }
    }
  } else {
    gateways.value.push({
      ...form.value,
      id: Date.now(),
      status: 'Pending',
      tenants: 0
    })
  }
  showModal.value = false
  isEditing.value = false
  resetForm()
}

const deleteGateway = (gw) => {
  if (confirm(`Delete gateway "${gw.name}"?`)) {
    gateways.value = gateways.value.filter(g => g.id !== gw.id)
  }
}

const restartGateway = (gw) => {
  alert(`Restarting gateway: ${gw.name}`)
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
