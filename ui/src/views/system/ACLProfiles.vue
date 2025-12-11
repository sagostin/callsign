<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Access Control Lists</h2>
      <p class="text-muted text-sm">Network-based access control for SIP profiles and trunks (FreeSWITCH ACLs).</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="openCreate">+ New ACL</button>
    </div>
  </div>

  <div class="route-help">
    <InfoIcon class="help-icon" />
    <span>ACLs control which IP addresses can register or send calls through SIP profiles. Use for trunk authentication and security.</span>
  </div>

  <div class="acl-grid" v-if="acls.length > 0">
    <div class="acl-card" v-for="acl in acls" :key="acl.id" :class="{ disabled: !acl.enabled }">
      <div class="acl-header">
        <div class="acl-info">
          <h4>{{ acl.name }}</h4>
          <span class="acl-desc">{{ acl.description }}</span>
        </div>
        <div class="acl-badges">
          <span class="badge" :class="acl.default === 'allow' ? 'allow' : 'deny'">
            Default: {{ acl.default }}
          </span>
        </div>
      </div>
      
      <div class="nodes-list" v-if="acl.nodes && acl.nodes.length > 0">
        <div class="node-item" v-for="node in acl.nodes.slice(0, 5)" :key="node.id" :class="node.type">
          <span class="node-type">{{ node.type }}</span>
          <code class="node-cidr">{{ node.cidr || node.domain }}</code>
          <span class="node-desc" v-if="node.description">{{ node.description }}</span>
        </div>
        <div class="node-more" v-if="acl.nodes.length > 5">
          +{{ acl.nodes.length - 5 }} more entries...
        </div>
      </div>
      <div class="nodes-empty" v-else>
        No entries configured
      </div>

      <div class="acl-controls">
        <label class="switch small">
          <input type="checkbox" v-model="acl.enabled" @change="toggleACL(acl)">
          <span class="slider round"></span>
        </label>
        <button class="btn-icon" @click="editACL(acl)"><EditIcon class="icon-sm" /></button>
        <button class="btn-icon" @click="deleteACL(acl)"><TrashIcon class="icon-sm text-bad" /></button>
      </div>
    </div>
  </div>

  <div v-else class="empty-state">
    <ShieldIcon class="empty-icon" />
    <p>No ACLs configured yet.</p>
    <button class="btn-secondary" @click="openCreate">Create your first ACL</button>
  </div>

  <!-- ACL MODAL -->
  <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
    <div class="modal-card large">
      <div class="modal-header">
        <h3>{{ isEditing ? 'Edit ACL' : 'New Access Control List' }}</h3>
        <button class="btn-icon" @click="showModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-2">
            <label>Name</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. trunks, domains">
          </div>
          <div class="form-group">
            <label>Default Action</label>
            <select v-model="form.default" class="input-field">
              <option value="deny">Deny</option>
              <option value="allow">Allow</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>Description</label>
          <input v-model="form.description" class="input-field" placeholder="What is this ACL for?">
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <div class="section-header">
            <h4>Nodes (IP/CIDR Rules)</h4>
            <button class="btn-small" @click="addNode">+ Add Entry</button>
          </div>
          <p class="help-text">Define IP addresses or CIDR blocks. First matching rule wins.</p>
          
          <div class="nodes-editor">
            <div class="node-row" v-for="(node, i) in formNodes" :key="i">
              <select v-model="node.type" class="input-field small">
                <option value="allow">Allow</option>
                <option value="deny">Deny</option>
              </select>
              <input v-model="node.cidr" class="input-field code flex-2" placeholder="192.168.1.0/24 or IP">
              <input v-model="node.description" class="input-field" placeholder="Description (optional)">
              <input v-model.number="node.priority" type="number" class="input-field tiny" placeholder="100" title="Priority">
              <button class="btn-icon" @click="removeNode(i)"><XIcon class="icon-sm" /></button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveACL" :disabled="saving || !form.name">
          {{ saving ? 'Saving...' : 'Save ACL' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { 
  Info as InfoIcon, Shield as ShieldIcon, Edit as EditIcon,
  Trash2 as TrashIcon, X as XIcon
} from 'lucide-vue-next'
import { systemAPI } from '../../services/api'

const acls = ref([])
const loading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)

const defaultForm = () => ({
  name: '',
  description: '',
  default: 'deny',
  enabled: true
})

const form = ref(defaultForm())
const formNodes = ref([])

const loadACLs = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listACLs()
    acls.value = response.data.data || response.data || []
  } catch (e) {
    console.error('Failed to load ACLs:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadACLs)

const openCreate = () => {
  form.value = defaultForm()
  formNodes.value = [{ type: 'allow', cidr: '', description: '', priority: 100 }]
  isEditing.value = false
  showModal.value = true
}

const editACL = (acl) => {
  form.value = { ...acl }
  formNodes.value = (acl.nodes || []).map(n => ({ ...n }))
  if (formNodes.value.length === 0) {
    formNodes.value = [{ type: 'allow', cidr: '', description: '', priority: 100 }]
  }
  isEditing.value = true
  showModal.value = true
}

const saveACL = async () => {
  saving.value = true
  try {
    // Save the ACL first
    let savedACL
    if (isEditing.value && form.value.id) {
      const resp = await systemAPI.updateACL(form.value.id, form.value)
      savedACL = resp.data
      
      // Delete existing nodes and recreate
      const existingNodes = acls.value.find(a => a.id === form.value.id)?.nodes || []
      for (const node of existingNodes) {
        try { await systemAPI.deleteACLNode(form.value.id, node.id) } catch(e) {}
      }
      // Add new nodes
      for (const node of formNodes.value.filter(n => n.cidr)) {
        await systemAPI.createACLNode(form.value.id, node)
      }
    } else {
      const resp = await systemAPI.createACL(form.value)
      savedACL = resp.data
      // Add nodes
      for (const node of formNodes.value.filter(n => n.cidr)) {
        await systemAPI.createACLNode(savedACL.id, node)
      }
    }
    await loadACLs()
    showModal.value = false
  } catch (e) {
    alert('Failed to save ACL: ' + e.message)
  } finally {
    saving.value = false
  }
}

const deleteACL = async (acl) => {
  if (!confirm(`Delete ACL "${acl.name}"? This cannot be undone.`)) return
  try {
    await systemAPI.deleteACL(acl.id)
    await loadACLs()
  } catch (e) {
    alert('Failed to delete ACL: ' + e.message)
  }
}

const toggleACL = async (acl) => {
  try {
    await systemAPI.updateACL(acl.id, { enabled: acl.enabled })
  } catch (e) {
    alert('Failed to update ACL: ' + e.message)
  }
}

const addNode = () => formNodes.value.push({ type: 'allow', cidr: '', description: '', priority: 100 })
const removeNode = (i) => formNodes.value.splice(i, 1)
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

.route-help { display: flex; align-items: center; gap: 8px; padding: 12px; background: #eff6ff; border-radius: var(--radius-sm); margin-bottom: 16px; color: #1e40af; font-size: 13px; }
.help-icon { width: 16px; height: 16px; }

.acl-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 16px; }

.acl-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.acl-card.disabled { opacity: 0.5; }

.acl-header { display: flex; justify-content: space-between; align-items: flex-start; }
.acl-info h4 { font-size: 15px; font-weight: 600; margin: 0 0 2px; }
.acl-desc { font-size: 12px; color: var(--text-muted); }

.acl-badges { display: flex; gap: 6px; }
.badge { font-size: 10px; padding: 3px 8px; border-radius: 3px; font-weight: 600; text-transform: uppercase; }
.badge.allow { background: #dcfce7; color: #16a34a; }
.badge.deny { background: #fee2e2; color: #dc2626; }

.nodes-list { display: flex; flex-direction: column; gap: 6px; }
.node-item { display: flex; align-items: center; gap: 8px; font-size: 12px; background: var(--bg-app); padding: 6px 10px; border-radius: 4px; }
.node-item.allow { border-left: 3px solid #16a34a; }
.node-item.deny { border-left: 3px solid #dc2626; }
.node-type { font-weight: 600; text-transform: uppercase; font-size: 10px; width: 40px; }
.node-item.allow .node-type { color: #16a34a; }
.node-item.deny .node-type { color: #dc2626; }
.node-cidr { background: #1e293b; color: #22d3ee; padding: 2px 6px; border-radius: 3px; font-size: 11px; }
.node-desc { color: var(--text-muted); flex: 1; text-align: right; }
.node-more { font-size: 11px; color: var(--text-muted); font-style: italic; padding: 4px 10px; }

.nodes-empty { font-size: 12px; color: var(--text-muted); font-style: italic; padding: 8px 0; }

.acl-controls { display: flex; align-items: center; gap: 8px; justify-content: flex-end; padding-top: 8px; border-top: 1px solid var(--border-color); }

.empty-state { text-align: center; padding: 60px; color: var(--text-muted); background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }
.empty-icon { width: 48px; height: 48px; margin-bottom: 16px; opacity: 0.5; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); }
.modal-card { background: white; border-radius: var(--radius-md); box-shadow: var(--shadow-lg); width: 100%; max-width: 520px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-card.large { max-width: 680px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Form Elements */
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.flex-2 { flex: 2; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.code { font-family: monospace; background: #f8fafc; }
.input-field.small { width: 90px; }
.input-field.tiny { width: 60px; text-align: center; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); margin-bottom: 8px; }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.form-section { margin-bottom: 16px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-header h4 { font-size: 13px; font-weight: 600; margin: 0; }

.nodes-editor { display: flex; flex-direction: column; gap: 8px; }
.node-row { display: flex; gap: 8px; align-items: center; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; }
.btn-small { font-size: 11px; padding: 4px 8px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.icon-sm { width: 16px; height: 16px; }
.text-bad { color: var(--status-bad); }

/* Toggle Switch */
.switch { position: relative; display: inline-block; width: 34px; height: 20px; }
.switch.small { width: 30px; height: 18px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; border-radius: 34px; transition: 0.3s; }
.slider:before { position: absolute; content: ""; height: 14px; width: 14px; left: 2px; bottom: 2px; background-color: white; border-radius: 50%; transition: 0.3s; }
.switch.small .slider:before { height: 12px; width: 12px; }
.switch input:checked + .slider { background-color: var(--primary-color); }
.switch input:checked + .slider:before { transform: translateX(14px); }
.switch.small input:checked + .slider:before { transform: translateX(12px); }
</style>
