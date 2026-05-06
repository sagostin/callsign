<template>
  <div class="users-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Users</h2>
        <p class="text-muted text-sm">Manage tenant users, extensions, and access.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="openCreateModal">
          <PlusIcon class="btn-icon" /> New User
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ users.length }}</div>
        <div class="stat-label">Total Users</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ activeUsers }}</div>
        <div class="stat-label">Active</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ adminUsers }}</div>
        <div class="stat-label">Admins</div>
      </div>
    </div>

    <!-- Filter & Search -->
    <div class="toolbar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" placeholder="Search users..." class="search-input" />
      </div>
      <select v-model="roleFilter" class="filter-select">
        <option value="">All Roles</option>
        <option value="tenant_admin">Tenant Admin</option>
        <option value="user">User</option>
      </select>
    </div>

    <!-- Users Table -->
    <div class="table-container">
      <DataTable :columns="columns" :data="filteredUsers" actions>
        <template #name="{ value, row }">
          <div class="user-cell">
            <div class="user-avatar">{{ getInitials(value) }}</div>
            <div class="user-info">
              <span class="user-name">{{ value }}</span>
              <span class="user-email">{{ row.email }}</span>
            </div>
          </div>
        </template>

        <template #role="{ value }">
          <span class="role-badge" :class="value">{{ formatRole(value) }}</span>
        </template>

        <template #extension="{ value }">
          <code class="ext-badge" v-if="value">{{ value }}</code>
          <span class="text-muted text-xs" v-else>—</span>
        </template>

        <template #status="{ value }">
          <span class="status-badge" :class="value">{{ value }}</span>
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
          <h3>{{ isEditing ? 'Edit User' : 'New User' }}</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Full Name *</label>
            <input v-model="form.name" class="input-field" placeholder="John Smith">
          </div>

          <div class="form-group">
            <label>Email Address *</label>
            <input v-model="form.email" type="email" class="input-field" placeholder="user@company.com">
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Role</label>
              <select v-model="form.role" class="input-field">
                <option value="user">User</option>
                <option value="tenant_admin">Tenant Admin</option>
              </select>
            </div>
            <div class="form-group">
              <label>Status</label>
              <select v-model="form.status" class="input-field">
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>

          <div class="form-group" v-if="!isEditing">
            <label>Password *</label>
            <input v-model="form.password" type="password" class="input-field" placeholder="••••••••">
          </div>

          <div class="form-group">
            <label>Extension (optional)</label>
            <input v-model="form.extension" class="input-field" placeholder="e.g. 1001">
            <span class="help-text">Assign an existing extension to this user</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveUser" :disabled="saving">
            {{ saving ? 'Saving...' : (isEditing ? 'Update' : 'Create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-card modal-sm">
        <div class="modal-header danger">
          <h3>Delete User</h3>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete <strong>{{ deleteTarget?.name }}</strong>?</p>
          <p class="text-muted">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn-danger" @click="deleteUser" :disabled="deleting">
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
import { usersAPI } from '../../services/api'

const toast = inject('toast')

const users = ref([])
const loading = ref(false)
const searchQuery = ref('')
const roleFilter = ref('')
const showModal = ref(false)
const showDeleteModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)
const deleting = ref(false)
const deleteTarget = ref(null)

const columns = [
  { key: 'name', label: 'User' },
  { key: 'role', label: 'Role', width: '140px' },
  { key: 'extension', label: 'Extension', width: '120px' },
  { key: 'status', label: 'Status', width: '100px' },
]

const defaultForm = {
  name: '',
  email: '',
  role: 'user',
  status: 'active',
  password: '',
  extension: ''
}

const form = ref({ ...defaultForm })
const editingId = ref(null)

onMounted(() => loadData())

async function loadData() {
  loading.value = true
  try {
    const response = await usersAPI.list()
    users.value = (response.data || []).map(u => ({
      id: u.id,
      name: u.name || `${u.first_name || ''} ${u.last_name || ''}`.trim() || u.username || 'Unknown',
      email: u.email || '',
      role: u.role || 'user',
      extension: u.extension || (u.extensions?.[0]?.extension) || '',
      status: u.status || 'active'
    }))
  } catch (error) {
    toast?.error(error.message || 'Failed to load users')
    users.value = []
  } finally {
    loading.value = false
  }
}

const filteredUsers = computed(() => {
  let result = users.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(u =>
      u.name.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q)
    )
  }
  if (roleFilter.value) {
    result = result.filter(u => u.role === roleFilter.value)
  }
  return result
})

const activeUsers = computed(() => users.value.filter(u => u.status === 'active').length)
const adminUsers = computed(() => users.value.filter(u => u.role === 'tenant_admin').length)

function getInitials(name) {
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
}

function formatRole(role) {
  const labels = { tenant_admin: 'Tenant Admin', user: 'User', system_admin: 'System Admin' }
  return labels[role] || role
}

function openCreateModal() {
  form.value = { ...defaultForm }
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function openEditModal(row) {
  form.value = {
    name: row.name,
    email: row.email,
    role: row.role,
    status: row.status,
    password: '',
    extension: row.extension
  }
  isEditing.value = true
  editingId.value = row.id
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  form.value = { ...defaultForm }
}

async function saveUser() {
  if (!form.value.name || !form.value.email) {
    toast?.error('Please fill in required fields')
    return
  }
  if (!isEditing.value && !form.value.password) {
    toast?.error('Password is required for new users')
    return
  }

  const payload = {
    name: form.value.name,
    email: form.value.email,
    role: form.value.role,
    status: form.value.status,
    extension: form.value.extension || undefined
  }

  if (!isEditing.value) {
    payload.password = form.value.password
  }

  saving.value = true
  try {
    if (isEditing.value) {
      await usersAPI.update(editingId.value, payload)
      toast?.success('User updated')
    } else {
      await usersAPI.create(payload)
      toast?.success('User created')
    }
    closeModal()
    await loadData()
  } catch (error) {
    toast?.error(error.message || 'Failed to save user')
  } finally {
    saving.value = false
  }
}

function confirmDelete(row) {
  deleteTarget.value = row
  showDeleteModal.value = true
}

async function deleteUser() {
  deleting.value = true
  try {
    await usersAPI.delete(deleteTarget.value.id)
    toast?.success('User deleted')
    showDeleteModal.value = false
    await loadData()
  } catch (error) {
    toast?.error(error.message || 'Failed to delete user')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.users-page { padding: 0; }

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
.filter-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; min-width: 150px; background: white; }

.table-container { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }

.user-cell { display: flex; align-items: center; gap: 12px; }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; background: linear-gradient(135deg, #6366f1, #818cf8); color: white; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; }
.user-info { display: flex; flex-direction: column; }
.user-name { font-weight: 600; }
.user-email { font-size: 12px; color: var(--text-muted); }

.role-badge { display: inline-flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 600; padding: 4px 10px; border-radius: 4px; text-transform: uppercase; }
.role-badge.tenant_admin { background: #dbeafe; color: #2563eb; }
.role-badge.user { background: #f1f5f9; color: #64748b; }
.role-badge.system_admin { background: #fef3c7; color: #b45309; }

.ext-badge { background: #1e293b; color: #fff; padding: 3px 8px; border-radius: 4px; font-family: monospace; font-size: 12px; font-weight: 600; }
.status-badge { padding: 4px 8px; border-radius: 4px; font-size: 10px; font-weight: 700; text-transform: uppercase; }
.status-badge.active { background: #dcfce7; color: #16a34a; }
.status-badge.inactive { background: #f1f5f9; color: #94a3b8; }

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

.text-muted { color: var(--text-muted); }
.text-xs { font-size: 11px; }

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
</style>
