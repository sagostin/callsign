<template>
  <div class="callblock-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Call Block</h2>
        <p class="text-muted text-sm">Manage blocked numbers and patterns for inbound/outbound filtering.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showImportModal = true">
          <UploadIcon class="btn-icon" /> Import
        </button>
        <button class="btn-primary" @click="showCreateModal = true">
          <PlusIcon class="btn-icon" /> Block Number
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ blockedNumbers.length }}</div>
        <div class="stat-label">Total Blocked</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ inboundCount }}</div>
        <div class="stat-label">Inbound Rules</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ outboundCount }}</div>
        <div class="stat-label">Outbound Rules</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ totalHits }}</div>
        <div class="stat-label">Blocks Today</div>
      </div>
    </div>

    <!-- Filter Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: filter === 'all' }" @click="filter = 'all'">
        All Rules
      </button>
      <button class="tab" :class="{ active: filter === 'inbound' }" @click="filter = 'inbound'">
        <PhoneIncomingIcon class="tab-icon" /> Inbound
      </button>
      <button class="tab" :class="{ active: filter === 'outbound' }" @click="filter = 'outbound'">
        <PhoneOutgoingIcon class="tab-icon" /> Outbound
      </button>
    </div>

    <!-- Table -->
    <div class="tab-content">
      <table class="data-table">
        <thead>
          <tr>
            <th>Number / Pattern</th>
            <th>Name</th>
            <th>Type</th>
            <th>Action</th>
            <th>Hits</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in filteredRules" :key="rule.id">
            <td>
              <div class="number-cell">
                <span class="number-value">{{ rule.number }}</span>
                <span v-if="rule.number.includes('*')" class="pattern-badge">Pattern</span>
              </div>
            </td>
            <td>{{ rule.name }}</td>
            <td>
              <span class="type-badge" :class="rule.type.toLowerCase()">{{ rule.type }}</span>
            </td>
            <td>
              <span class="action-badge" :class="rule.action.toLowerCase()">
                <component :is="getActionIcon(rule.action)" class="action-icon" />
                {{ rule.action }}
              </span>
            </td>
            <td class="hits-cell">
              <span class="hits-value">{{ rule.hits }}</span>
              <span v-if="rule.recentHits" class="recent-badge">+{{ rule.recentHits }} today</span>
            </td>
            <td class="actions-cell">
              <button class="btn-icon-sm" @click="editRule(rule)" title="Edit"><EditIcon /></button>
              <button class="btn-icon-sm danger" @click="deleteRule(rule)" title="Delete"><TrashIcon /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <div class="modal-overlay" v-if="showCreateModal" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ editingRule ? 'Edit Block Rule' : 'Block Number' }}</h3>
          <button class="close-btn" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Number or Pattern</label>
            <input v-model="form.number" class="input-field" placeholder="+1555*, 8005551234">
            <span class="help-text">Use * as wildcard (e.g., +1555* blocks all starting with +1555)</span>
          </div>
          <div class="form-group">
            <label>Name / Reason</label>
            <input v-model="form.name" class="input-field" placeholder="Known telemarketer">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Type</label>
              <select v-model="form.type" class="input-field">
                <option value="Inbound">Inbound</option>
                <option value="Outbound">Outbound</option>
                <option value="Both">Both</option>
              </select>
            </div>
            <div class="form-group">
              <label>Action</label>
              <select v-model="form.action" class="input-field">
                <option value="Reject">Reject (silent)</option>
                <option value="Busy">Play Busy</option>
                <option value="Hangup">Hangup</option>
                <option value="Voicemail">Send to VM</option>
              </select>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveRule">{{ editingRule ? 'Update' : 'Block' }}</button>
        </div>
      </div>
    </div>

    <!-- Import Modal -->
    <div class="modal-overlay" v-if="showImportModal" @click.self="showImportModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Import Blocked Numbers</h3>
          <button class="close-btn" @click="showImportModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Paste Numbers (one per line)</label>
            <textarea v-model="importData" class="input-field" rows="8" placeholder="5551234567
8005551234
+1555*"></textarea>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Default Type</label>
              <select v-model="importType" class="input-field">
                <option value="Inbound">Inbound</option>
                <option value="Outbound">Outbound</option>
              </select>
            </div>
            <div class="form-group">
              <label>Default Action</label>
              <select v-model="importAction" class="input-field">
                <option value="Reject">Reject</option>
                <option value="Busy">Busy</option>
              </select>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showImportModal = false">Cancel</button>
          <button class="btn-primary" @click="processImport">Import</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { 
  Plus as PlusIcon, Upload as UploadIcon, 
  PhoneIncoming as PhoneIncomingIcon, PhoneOutgoing as PhoneOutgoingIcon,
  Edit as EditIcon, Trash2 as TrashIcon, Ban, Phone, Voicemail
} from 'lucide-vue-next'
import { routingAPI } from '../../services/api'

const filter = ref('all')
const showCreateModal = ref(false)
const showImportModal = ref(false)
const editingRule = ref(null)
const isLoading = ref(false)

const form = ref({ number: '', notes: '', match_type: 'exact', action: 'reject' })
const importData = ref('')
const importType = ref('exact')
const importAction = ref('reject')

const blockedNumbers = ref([])

// Computed stats
const inboundCount = computed(() => blockedNumbers.value.length) // All are inbound for now
const outboundCount = computed(() => 0) // Future feature
const totalHits = computed(() => 0) // CDR integration TODO

const filteredRules = computed(() => {
  return blockedNumbers.value // Filter by type when we add outbound support
})

const getActionIcon = (action) => {
  const icons = { reject: Ban, busy: Phone, hangup: Phone, voicemail: Voicemail }
  return icons[action] || Ban
}

// API Functions
const loadBlocks = async () => {
  isLoading.value = true
  try {
    const response = await routingAPI.listBlocks()
    blockedNumbers.value = response.data.data || []
  } catch (e) {
    console.error('Failed to load call blocks', e)
  } finally {
    isLoading.value = false
  }
}

const editRule = (rule) => {
  editingRule.value = rule
  form.value = {
    number: rule.number,
    notes: rule.notes || '',
    match_type: rule.match_type,
    action: rule.action
  }
  showCreateModal.value = true
}

const deleteRule = async (rule) => {
  if (!confirm(`Remove block for "${rule.number}"?`)) return
  try {
    await routingAPI.deleteBlock(rule.id)
    await loadBlocks()
  } catch (e) {
    console.error(e)
    alert('Failed to delete block')
  }
}

const saveRule = async () => {
  try {
    if (editingRule.value) {
      await routingAPI.updateBlock(editingRule.value.id, form.value)
    } else {
      await routingAPI.createBlock(form.value)
    }
    await loadBlocks()
    closeModal()
  } catch (e) {
    console.error(e)
    alert('Failed to save block rule')
  }
}

const closeModal = () => {
  showCreateModal.value = false
  editingRule.value = null
  form.value = { number: '', notes: '', match_type: 'exact', action: 'reject' }
}

const processImport = async () => {
  const lines = importData.value.split('\n').filter(l => l.trim())
  try {
    for (const line of lines) {
      await routingAPI.createBlock({
        number: line.trim(),
        match_type: importType.value,
        action: importAction.value,
        notes: 'Imported'
      })
    }
    await loadBlocks()
    showImportModal.value = false
    importData.value = ''
  } catch (e) {
    console.error(e)
    alert('Import failed')
  }
}

onMounted(() => {
  loadBlocks()
})
</script>

<style scoped>
.callblock-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: #ef4444; background: #fef2f2; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 6px 6px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 14px; height: 14px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 8px 8px; }

.data-table { width: 100%; border-collapse: collapse; }
.data-table th { text-align: left; padding: 10px 12px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); }
.data-table td { padding: 12px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.data-table tr:hover { background: #f8fafc; }

.number-cell { display: flex; align-items: center; gap: 8px; }
.number-value { font-family: monospace; font-weight: 600; }
.pattern-badge { font-size: 9px; background: #dbeafe; color: #2563eb; padding: 2px 6px; border-radius: 3px; font-weight: 500; }

.type-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; }
.type-badge.inbound { background: #dcfce7; color: #16a34a; }
.type-badge.outbound { background: #fef3c7; color: #d97706; }
.type-badge.both { background: #f3e8ff; color: #9333ea; }

.action-badge { display: inline-flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 500; padding: 3px 8px; border-radius: 4px; background: #fef2f2; color: #dc2626; }
.action-icon { width: 12px; height: 12px; }

.hits-cell { display: flex; align-items: center; gap: 8px; }
.hits-value { font-weight: 600; }
.recent-badge { font-size: 10px; color: #ef4444; font-weight: 500; }

.actions-cell { display: flex; gap: 4px; }
.btn-icon-sm { width: 28px; height: 28px; background: white; border: 1px solid var(--border-color); border-radius: 4px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); }
.btn-icon-sm:hover { color: var(--primary-color); border-color: var(--primary-color); }
.btn-icon-sm.danger:hover { color: #ef4444; border-color: #ef4444; }
.btn-icon-sm svg { width: 14px; height: 14px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.input-field:focus { border-color: var(--primary-color); outline: none; }
.help-text { font-size: 10px; color: var(--text-muted); margin-top: 4px; display: block; }
</style>
