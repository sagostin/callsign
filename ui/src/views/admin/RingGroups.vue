<template>
  <div class="ring-groups-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Ring Groups</h2>
        <p class="text-muted text-sm">Manage ring groups for call distribution across multiple extensions.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="$router.push('/admin/ring-groups/new')">
          <PlusIcon class="btn-icon" /> New Ring Group
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ ringGroups.length }}</div>
        <div class="stat-label">Total Groups</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ totalMembers }}</div>
        <div class="stat-label">Total Members</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ activeGroups }}</div>
        <div class="stat-label">Active</div>
      </div>
    </div>

    <!-- Filter & Search -->
    <div class="toolbar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" placeholder="Search ring groups..." class="search-input" />
      </div>
    </div>

    <!-- Ring Groups Table -->
    <div class="table-container">
      <DataTable :columns="columns" :data="filteredGroups" actions>
        <template #name="{ value, row }">
          <div class="name-cell">
            <strong>{{ value }}</strong>
            <span v-if="row.protected" class="protected-badge">System</span>
          </div>
        </template>

        <template #extension="{ value }">
          <code class="ext-badge">{{ value || '—' }}</code>
        </template>

        <template #strategy="{ value }">
          <span class="strategy-pill">{{ formatStrategy(value) }}</span>
        </template>

        <template #members="{ value }">
          <span class="member-count">{{ value }} members</span>
        </template>

        <template #status="{ value }">
          <span class="status-badge" :class="value ? 'enabled' : 'disabled'">
            {{ value ? 'Active' : 'Inactive' }}
          </span>
        </template>

        <template #actions="{ row }">
          <button class="btn-link" @click="$router.push(`/admin/ring-groups/${row.id}`)">Edit</button>
          <button class="btn-link danger" @click="confirmDelete(row)" :disabled="row.protected">Delete</button>
        </template>
      </DataTable>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-card modal-sm">
        <div class="modal-header danger">
          <h3>Delete Ring Group</h3>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete <strong>{{ deleteTarget?.name }}</strong>?</p>
          <p class="text-muted">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger" @click="deleteGroup" :disabled="deleting">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import {
  Plus as PlusIcon,
  Search as SearchIcon
} from 'lucide-vue-next'
import DataTable from '../../components/common/DataTable.vue'
import { ringGroupsAPI } from '../../services/api'

const toast = inject('toast')

const ringGroups = ref([])
const loading = ref(false)
const searchQuery = ref('')
const showDeleteModal = ref(false)
const deleting = ref(false)
const deleteTarget = ref(null)

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'extension', label: 'Extension', width: '120px' },
  { key: 'strategy', label: 'Strategy', width: '140px' },
  { key: 'members', label: 'Members', width: '120px' },
  { key: 'status', label: 'Status', width: '100px' },
]

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const response = await ringGroupsAPI.list()
    ringGroups.value = (response.data || []).map(g => ({
      id: g.id,
      name: g.name,
      extension: g.extension,
      strategy: g.strategy || 'simultaneous',
      members: g.destination_count || (g.members?.length || 0),
      status: g.enabled !== false,
      protected: g.is_system
    }))
  } catch (error) {
    toast?.error(error.message || 'Failed to load ring groups')
    ringGroups.value = []
  } finally {
    loading.value = false
  }
}

const filteredGroups = computed(() => {
  if (!searchQuery.value) return ringGroups.value
  const q = searchQuery.value.toLowerCase()
  return ringGroups.value.filter(g =>
    g.name.toLowerCase().includes(q) ||
    (g.extension && g.extension.includes(q))
  )
})

const totalMembers = computed(() => ringGroups.value.reduce((sum, g) => sum + g.members, 0))
const activeGroups = computed(() => ringGroups.value.filter(g => g.status).length)

function formatStrategy(strategy) {
  const labels = {
    simultaneous: 'Ring All',
    sequence: 'Sequential',
    enterprise: 'Enterprise',
    rollover: 'Rollover',
    random: 'Random'
  }
  return labels[strategy] || strategy
}

function confirmDelete(row) {
  if (row.protected) return
  deleteTarget.value = row
  showDeleteModal.value = true
}

async function deleteGroup() {
  deleting.value = true
  try {
    await ringGroupsAPI.delete(deleteTarget.value.id)
    toast?.success('Ring group deleted')
    showDeleteModal.value = false
    await loadData()
  } catch (error) {
    toast?.error(error.message || 'Failed to delete ring group')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.ring-groups-page { padding: 0; }

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }

.btn-primary, .btn-secondary, .btn-danger {
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
.btn-danger { background: #dc2626; color: white; }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.toolbar { display: flex; gap: 12px; margin-bottom: 16px; }
.search-box { position: relative; flex: 1; max-width: 300px; }
.search-icon { position: absolute; left: 10px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 8px 10px 8px 36px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; }

.table-container { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }

.name-cell { display: flex; align-items: center; gap: 8px; }
.protected-badge { font-size: 9px; background: #f1f5f9; color: #64748b; padding: 2px 6px; border-radius: 3px; }
.ext-badge { background: #1e293b; color: #fff; padding: 3px 8px; border-radius: 4px; font-family: monospace; font-size: 12px; font-weight: 600; }
.strategy-pill { font-size: 11px; background: #f1f5f9; padding: 3px 8px; border-radius: 4px; }
.member-count { font-size: 12px; color: var(--text-muted); }
.status-badge { padding: 4px 8px; border-radius: 4px; font-size: 10px; font-weight: 700; }
.status-badge.enabled { background: #dcfce7; color: #16a34a; }
.status-badge.disabled { background: #f1f5f9; color: #94a3b8; }

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  margin-right: 8px;
}
.btn-link.danger { color: #dc2626; }
.btn-link:disabled { color: #cbd5e1; cursor: not-allowed; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-card.modal-sm { max-width: 400px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header.danger { background: #fef2f2; border-bottom-color: #fecaca; }
.modal-header.danger h3 { color: #dc2626; }
.modal-header h3 { margin: 0; font-size: 16px; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.text-muted { color: var(--text-muted); }
</style>
