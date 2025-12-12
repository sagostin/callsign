<template>
  <div class="queues-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Queues & Groups</h2>
        <p class="text-muted text-sm">Manage call queues and ring groups for agent distribution.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="activeTab = 'queues'; $router.push('/admin/queues/new')">
          <PlusIcon class="btn-icon" /> New Queue
        </button>
        <button class="btn-primary" @click="activeTab = 'groups'; showGroupModal = true">
          <UsersIcon class="btn-icon" /> New Ring Group
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ queues.length }}</div>
        <div class="stat-label">Call Queues</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ groups.length }}</div>
        <div class="stat-label">Ring Groups</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ totalAgents }}</div>
        <div class="stat-label">Total Agents</div>
      </div>
      <div class="stat-card" :class="{ 'alert': waitingCalls > 0 }">
        <div class="stat-value">{{ waitingCalls }}</div>
        <div class="stat-label">Waiting Calls</div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'queues' }" @click="activeTab = 'queues'">
        <PhoneIncomingIcon class="tab-icon" /> Call Queues
      </button>
      <button class="tab" :class="{ active: activeTab === 'groups' }" @click="activeTab = 'groups'">
        <UsersIcon class="tab-icon" /> Ring Groups
      </button>
    </div>

    <!-- Queues Tab -->
    <div class="tab-content" v-if="activeTab === 'queues'">
      <div class="queue-grid">
        <div v-for="queue in queues" :key="queue.id" class="queue-card" :class="{ 'has-waiting': queue.waiting > 0 }">
          <div class="queue-header">
            <div class="queue-info">
              <h4>{{ queue.name }}</h4>
              <span class="queue-ext">Ext. {{ queue.extension }}</span>
            </div>
            <div class="queue-status" :class="queue.status.toLowerCase()">{{ queue.status }}</div>
          </div>
          <div class="queue-stats">
            <div class="queue-stat">
              <span class="stat-num">{{ queue.agents }}</span>
              <span class="stat-lbl">Agents</span>
            </div>
            <div class="queue-stat">
              <span class="stat-num">{{ queue.waiting }}</span>
              <span class="stat-lbl">Waiting</span>
            </div>
            <div class="queue-stat">
              <span class="stat-num">{{ queue.avgWait }}</span>
              <span class="stat-lbl">Avg Wait</span>
            </div>
          </div>
          <div class="queue-footer">
            <span class="strategy-badge">{{ queue.strategy }}</span>
            <div class="queue-actions">
              <button class="btn-icon-sm" @click="$router.push(`/admin/queues/${queue.id}`)" title="Edit">
                <EditIcon />
              </button>
              <button class="btn-icon-sm" title="View Stats">
                <BarChartIcon />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Ring Groups Tab -->
    <div class="tab-content" v-if="activeTab === 'groups'">
      <table class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Extension</th>
            <th>Strategy</th>
            <th>Members</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="group in groups" :key="group.id" :class="{ 'protected-row': group.protected }">
            <td class="name-cell">
              <strong>{{ group.name }}</strong>
              <span v-if="group.protected" class="protected-badge">System</span>
            </td>
            <td class="mono">{{ group.extension }}</td>
            <td><span class="strategy-pill">{{ group.strategy }}</span></td>
            <td>
              <span class="member-count">{{ group.members }} members</span>
            </td>
            <td>
              <span class="status-dot" :class="group.enabled ? 'active' : 'inactive'"></span>
              {{ group.enabled ? 'Active' : 'Inactive' }}
            </td>
            <td class="actions-cell">
              <button class="btn-link" @click="editGroup(group)">Edit</button>
              <button class="btn-link danger" @click="deleteGroup(group)" :disabled="group.protected">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Ring Group Modal -->
    <div class="modal-overlay" v-if="showGroupModal" @click.self="showGroupModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ editingGroup ? 'Edit Ring Group' : 'New Ring Group' }}</h3>
          <button class="close-btn" @click="showGroupModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Name</label>
            <input v-model="groupForm.name" class="input-field" placeholder="Sales Team">
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Extension</label>
              <input v-model="groupForm.extension" class="input-field" placeholder="500">
            </div>
            <div class="form-group">
              <label>Strategy</label>
              <select v-model="groupForm.strategy" class="input-field">
                <option value="Simultaneous">Simultaneous (Ring All)</option>
                <option value="Sequential">Sequential</option>
                <option value="Enterprise">Enterprise</option>
                <option value="Rollover">Rollover</option>
                <option value="Random">Random</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>Members</label>
            <textarea v-model="groupForm.membersList" class="input-field" rows="3" placeholder="Enter extensions, one per line..."></textarea>
            <span class="help-text">One extension per line</span>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="groupForm.enabled">
              Enabled
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showGroupModal = false">Cancel</button>
          <button class="btn-primary" @click="saveGroup">Save Ring Group</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { 
  Plus as PlusIcon, 
  Users as UsersIcon, 
  PhoneIncoming as PhoneIncomingIcon,
  Edit as EditIcon,
  BarChart2 as BarChartIcon
} from 'lucide-vue-next'
import { queuesAPI, ringGroupsAPI } from '@/services/api'

const toast = inject('toast')

const isLoading = ref(false)
const activeTab = ref('queues')
const showGroupModal = ref(false)
const editingGroup = ref(null)

const queues = ref([])
const groups = ref([])

const groupForm = ref({
  name: '',
  extension: '',
  strategy: 'Simultaneous',
  membersList: '',
  enabled: true
})

onMounted(async () => {
  await Promise.all([fetchQueues(), fetchRingGroups()])
})

async function fetchQueues() {
  isLoading.value = true
  try {
    const response = await queuesAPI.list()
    queues.value = (response.data || []).map(q => ({
      id: q.id,
      name: q.name,
      extension: q.extension,
      agents: q.agent_count || 0,
      waiting: q.waiting_calls || 0,
      avgWait: formatWaitTime(q.avg_wait_time),
      strategy: q.strategy || 'Ring All',
      status: q.waiting_calls > 0 ? 'Active' : (q.agent_count > 0 ? 'Idle' : 'Inactive')
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load queues')
    // Fallback to demo data
    queues.value = [
      { id: 1, name: 'Sales Main', extension: '8000', agents: 5, waiting: 2, avgWait: '0:45', strategy: 'Longest Idle', status: 'Active' },
      { id: 2, name: 'Support Tier 1', extension: '8001', agents: 12, waiting: 0, avgWait: '1:20', strategy: 'Ring All', status: 'Idle' },
    ]
  } finally {
    isLoading.value = false
  }
}

async function fetchRingGroups() {
  try {
    const response = await ringGroupsAPI.list()
    groups.value = (response.data || []).map(g => ({
      id: g.id,
      name: g.name,
      extension: g.extension,
      strategy: g.strategy || 'Simultaneous',
      members: g.destination_count || 0,
      enabled: g.enabled,
      protected: g.is_system
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load ring groups')
    // Fallback to demo data
    groups.value = [
      { id: 1, name: 'Sales Team', extension: '500', strategy: 'Simultaneous', members: 5, enabled: true },
      { id: 2, name: 'Support Team', extension: '501', strategy: 'Sequential', members: 8, enabled: true },
    ]
  }
}

function formatWaitTime(seconds) {
  if (!seconds) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const totalAgents = computed(() => {
  return queues.value.reduce((sum, q) => sum + q.agents, 0)
})

const waitingCalls = computed(() => {
  return queues.value.reduce((sum, q) => sum + q.waiting, 0)
})

const editGroup = (group) => {
  editingGroup.value = group
  groupForm.value = { ...group, membersList: '' }
  showGroupModal.value = true
}

const saveGroup = async () => {
  try {
    const data = {
      name: groupForm.value.name,
      extension: groupForm.value.extension,
      strategy: groupForm.value.strategy,
      enabled: groupForm.value.enabled,
    }
    
    if (editingGroup.value) {
      await ringGroupsAPI.update(editingGroup.value.id, data)
      toast?.success('Ring group updated')
    } else {
      await ringGroupsAPI.create(data)
      toast?.success('Ring group created')
    }
    
    await fetchRingGroups()
    showGroupModal.value = false
    editingGroup.value = null
    groupForm.value = { name: '', extension: '', strategy: 'Simultaneous', membersList: '', enabled: true }
  } catch (error) {
    toast?.error(error.message, 'Failed to save ring group')
  }
}

const deleteGroup = async (group) => {
  if (group.protected) return
  if (confirm(`Delete ring group "${group.name}"?`)) {
    try {
      await ringGroupsAPI.delete(group.id)
      toast?.success(`Ring group "${group.name}" deleted`)
      await fetchRingGroups()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete ring group')
    }
  }
}
</script>

<style scoped>
.queues-page { padding: 0; }

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

/* Stats */
.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card {
  flex: 1;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-card.alert { border-color: #f59e0b; background: #fffbeb; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

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

/* Queue Grid */
.queue-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; }

.queue-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s;
}
.queue-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
.queue-card.has-waiting { border-left: 3px solid #f59e0b; }

.queue-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 14px;
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.queue-info h4 { margin: 0 0 4px; font-size: 14px; }
.queue-ext { font-size: 11px; color: var(--text-muted); font-family: monospace; }
.queue-status {
  font-size: 10px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}
.queue-status.active { background: #dcfce7; color: #16a34a; }
.queue-status.idle { background: #f1f5f9; color: #64748b; }

.queue-stats {
  display: flex;
  padding: 14px;
  gap: 20px;
}
.queue-stat { text-align: center; flex: 1; }
.stat-num { display: block; font-size: 20px; font-weight: 700; color: var(--text-primary); }
.stat-lbl { font-size: 10px; color: var(--text-muted); text-transform: uppercase; }

.queue-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: #f8fafc;
  border-top: 1px solid var(--border-color);
}
.strategy-badge {
  font-size: 10px;
  background: #eff6ff;
  color: #3b82f6;
  padding: 3px 8px;
  border-radius: 4px;
  font-weight: 500;
}
.queue-actions { display: flex; gap: 4px; }
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
.btn-icon-sm svg { width: 14px; height: 14px; }

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
.protected-row { opacity: 0.7; }

.name-cell { display: flex; align-items: center; gap: 8px; }
.protected-badge { font-size: 9px; background: #f1f5f9; color: #64748b; padding: 2px 6px; border-radius: 3px; }
.mono { font-family: monospace; }
.strategy-pill { font-size: 11px; background: #f1f5f9; padding: 3px 8px; border-radius: 4px; }
.member-count { font-size: 12px; color: var(--text-muted); }
.status-dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; margin-right: 6px; }
.status-dot.active { background: #22c55e; }
.status-dot.inactive { background: #94a3b8; }

.actions-cell { display: flex; gap: 8px; }
.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
}
.btn-link.danger { color: #ef4444; }
.btn-link:disabled { color: #cbd5e1; cursor: not-allowed; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 500px; }
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
.help-text { font-size: 10px; color: var(--text-muted); margin-top: 4px; }
.checkbox-label { display: flex; align-items: center; gap: 8px; font-size: 12px; cursor: pointer; text-transform: none !important; }
</style>
