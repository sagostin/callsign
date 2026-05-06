<template>
  <div class="paging-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Paging / Page Groups</h2>
        <p class="text-muted text-sm">Manage page groups for overhead paging and intercom broadcasting.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="openCreateModal">
          <PlusIcon class="btn-icon" /> New Page Group
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ pageGroups.length }}</div>
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
        <input v-model="searchQuery" placeholder="Search page groups..." class="search-input" />
      </div>
    </div>

    <!-- Page Groups Table -->
    <div class="table-container">
      <DataTable :columns="columns" :data="filteredGroups" actions>
        <template #name="{ value, row }">
          <div class="name-cell">
            <strong>{{ value }}</strong>
            <span v-if="row.description" class="description-text">{{ row.description }}</span>
          </div>
        </template>

        <template #extension="{ value }">
          <code class="ext-badge">{{ value || '—' }}</code>
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
          <button class="btn-link" @click="openEditModal(row)">Edit</button>
          <button class="btn-link danger" @click="confirmDelete(row)">Delete</button>
        </template>
      </DataTable>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Page Group' : 'New Page Group' }}</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Group Name *</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. Warehouse Paging">
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Extension *</label>
              <input v-model="form.extension" class="input-field" placeholder="e.g. 800">
            </div>
            <div class="form-group">
              <label>Status</label>
              <select v-model="form.enabled" class="input-field">
                <option :value="true">Active</option>
                <option :value="false">Inactive</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label>Description</label>
            <input v-model="form.description" class="input-field" placeholder="Optional description">
          </div>

          <div class="form-group">
            <label>Members</label>
            <textarea v-model="membersText" class="input-field" rows="4" placeholder="Enter extensions, one per line..."></textarea>
            <span class="help-text">One extension per line</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveGroup" :disabled="saving">
            {{ saving ? 'Saving...' : (isEditing ? 'Update' : 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-card modal-sm">
        <div class="modal-header danger">
          <h3>Delete Page Group</h3>
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
import { pagingAPI } from '../../services/api'

const toast = inject('toast')

const pageGroups = ref([])
const loading = ref(false)
const searchQuery = ref('')
const showModal = ref(false)
const showDeleteModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleting = ref(false)
const deleteTarget = ref(null)

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'extension', label: 'Extension', width: '120px' },
  { key: 'members', label: 'Members', width: '120px' },
  { key: 'status', label: 'Status', width: '100px' },
]

const defaultForm = {
  name: '',
  extension: '',
  description: '',
  enabled: true,
  members: []
}

const form = ref({ ...defaultForm })
const editingId = ref(null)
const membersText = ref('')

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const response = await pagingAPI.list()
    pageGroups.value = (response.data || []).map(g => ({
      id: g.id,
      name: g.name,
      extension: g.extension,
      description: g.description,
      members: g.destination_count || (g.members?.length || 0),
      status: g.enabled !== false,
      raw: g
    }))
  } catch (error) {
    toast?.error(error.message || 'Failed to load page groups')
    pageGroups.value = []
  } finally {
    loading.value = false
  }
}

const filteredGroups = computed(() => {
  if (!searchQuery.value) return pageGroups.value
  const q = searchQuery.value.toLowerCase()
  return pageGroups.value.filter(g =>
    g.name.toLowerCase().includes(q) ||
    (g.extension && g.extension.includes(q))
  )
})

const totalMembers = computed(() => pageGroups.value.reduce((sum, g) => sum + g.members, 0))
const activeGroups = computed(() => pageGroups.value.filter(g => g.status).length)

function openCreateModal() {
  form.value = { ...defaultForm }
  membersText.value = ''
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function openEditModal(row) {
  form.value = {
    name: row.raw.name || '',
    extension: row.raw.extension || '',
    description: row.raw.description || '',
    enabled: row.raw.enabled !== false,
    members: row.raw.members || []
  }
  membersText.value = (row.raw.members || []).map(m => m.extension || m.target || m).join('\n')
  isEditing.value = true
  editingId.value = row.id
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  form.value = { ...defaultForm }
  membersText.value = ''
}

async function saveGroup() {
  if (!form.value.name || !form.value.extension) {
    toast?.error('Please fill in required fields')
    return
  }

  const members = membersText.value
    .split('\n')
    .map(l => l.trim())
    .filter(Boolean)
    .map(ext => ({ type: 'user', target: ext }))

  const payload = {
    name: form.value.name,
    extension: form.value.extension,
    description: form.value.description,
    enabled: form.value.enabled,
    members
  }

  saving.value = true
  try {
    if (isEditing.value) {
      await pagingAPI.update(editingId.value, payload)
      toast?.success('Page group updated')
    } else {
      await pagingAPI.create(payload)
      toast?.success('Page group created')
    }
    closeModal()
    await loadData()
  } catch (error) {
    toast?.error(error.message || 'Failed to save page group')
  } finally {
    saving.value = false
  }
}

function confirmDelete(row) {
  deleteTarget.value = row
  showDeleteModal.value = true
}

async function deleteGroup() {
  deleting.value = true
  try {
    await pagingAPI.delete(deleteTarget.value.id)
    toast?.success('Page group deleted')
    showDeleteModal.value = false
    await loadData()
  } catch (error) {
    toast?.error(error.message || 'Failed to delete page group')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.paging-page { padding: 0; }

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

.name-cell { display: flex; flex-direction: column; gap: 2px; }
.description-text { font-size: 11px; color: var(--text-muted); }
.ext-badge { background: #1e293b; color: #fff; padding: 3px 8px; border-radius: 4px; font-family: monospace; font-size: 12px; font-weight: 600; }
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

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
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
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); margin-top: 4px; display: block; }
.text-muted { color: var(--text-muted); }
</style>
