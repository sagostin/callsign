<template>
  <div class="system-users-page">
    <div class="view-header">
      <div class="header-content">
        <h2>System Administrators</h2>
        <p class="text-muted text-sm">Manage users with system-level administrative access across all tenants.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="showAddModal = true">
          <PlusIcon class="btn-icon" /> Add System Admin
        </button>
      </div>
    </div>

    <!-- Info Banner -->
    <div class="info-banner">
      <InfoIcon class="info-icon" />
      <div class="info-content">
        <strong>System Administrators vs Tenant Admins</strong>
        <p>System administrators have access to all tenants and global system settings. Tenant admins are managed within each tenant's settings and only have access to their assigned tenant.</p>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon total"><UsersIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ users.length }}</span>
          <span class="stat-label">Total System Admins</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon active"><CheckCircleIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ users.filter(u => u.status === 'active').length }}</span>
          <span class="stat-label">Active</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon superadmin"><ShieldIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ users.filter(u => u.role === 'superadmin').length }}</span>
          <span class="stat-label">Super Admins</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon online"><ActivityIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ users.filter(u => u.online).length }}</span>
          <span class="stat-label">Online Now</span>
        </div>
      </div>
    </div>

    <!-- Users Table -->
    <div class="users-table-container">
      <div class="table-header">
        <div class="search-box">
          <SearchIcon class="search-icon" />
          <input v-model="searchQuery" class="search-input" placeholder="Search by name or email...">
        </div>
        <select v-model="roleFilter" class="filter-select">
          <option value="">All Roles</option>
          <option value="superadmin">Super Admin</option>
          <option value="admin">System Admin</option>
          <option value="readonly">Read Only</option>
        </select>
      </div>

      <table class="users-table">
        <thead>
          <tr>
            <th>User</th>
            <th>Role</th>
            <th>Tenant Access</th>
            <th>2FA</th>
            <th>Last Login</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in filteredUsers" :key="user.id">
            <td>
              <div class="user-cell">
                <div class="user-avatar" :class="{ online: user.online }">
                  {{ user.name.split(' ').map(n => n[0]).join('') }}
                </div>
                <div class="user-info">
                  <span class="user-name">{{ user.name }}</span>
                  <span class="user-email">{{ user.email }}</span>
                </div>
              </div>
            </td>
            <td>
              <span class="role-badge" :class="user.role">
                <ShieldIcon v-if="user.role === 'superadmin'" class="role-icon" />
                <KeyIcon v-else-if="user.role === 'admin'" class="role-icon" />
                <EyeIcon v-else class="role-icon" />
                {{ getRoleLabel(user.role) }}
              </span>
            </td>
            <td>
              <span class="tenant-access" v-if="user.role === 'superadmin'">All Tenants</span>
              <span class="tenant-access limited" v-else>{{ user.tenantAccess?.length || 0 }} Tenants</span>
            </td>
            <td>
              <span class="tfa-badge" :class="{ enabled: user.twoFactor }">
                {{ user.twoFactor ? 'Enabled' : 'Disabled' }}
              </span>
            </td>
            <td>
              <span class="last-login">{{ user.lastLogin }}</span>
            </td>
            <td>
              <span class="status-badge" :class="user.status">{{ user.status }}</span>
            </td>
            <td class="actions-cell">
              <button class="action-btn" @click="editUser(user)" title="Edit">
                <EditIcon class="icon-sm" />
              </button>
              <button class="action-btn" @click="resetPassword(user)" title="Reset Password">
                <KeyIcon class="icon-sm" />
              </button>
              <button class="action-btn" @click="viewActivity(user)" title="View Activity">
                <ActivityIcon class="icon-sm" />
              </button>
              <button class="action-btn danger" @click="deleteUser(user)" title="Delete" :disabled="user.role === 'superadmin' && users.filter(u => u.role === 'superadmin').length === 1">
                <TrashIcon class="icon-sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div class="empty-state" v-if="filteredUsers.length === 0">
        <UsersIcon class="empty-icon" />
        <p>No system administrators found</p>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ editingUser ? 'Edit System Admin' : 'Add System Admin' }}</h3>
          <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label>Full Name *</label>
            <input v-model="userForm.name" class="input-field" placeholder="John Smith">
          </div>
          <div class="form-group">
            <label>Email Address *</label>
            <input v-model="userForm.email" type="email" class="input-field" placeholder="admin@company.com">
          </div>
          <div class="form-group" v-if="!editingUser">
            <label>Password *</label>
            <input v-model="userForm.password" type="password" class="input-field" placeholder="••••••••">
          </div>
          <div class="form-group">
            <label>Role *</label>
            <select v-model="userForm.role" class="input-field">
              <option value="superadmin">Super Admin (Full Access)</option>
              <option value="admin">System Admin</option>
              <option value="readonly">Read Only</option>
            </select>
            <span class="input-hint">
              <template v-if="userForm.role === 'superadmin'">Full access to all tenants and system settings.</template>
              <template v-else-if="userForm.role === 'admin'">Can manage assigned tenants and most system settings.</template>
              <template v-else>View-only access to system dashboard and reports.</template>
            </span>
          </div>

          <div class="form-group" v-if="userForm.role !== 'superadmin'">
            <label>Tenant Access</label>
            <div class="tenant-checkboxes">
              <label class="checkbox-item" v-for="tenant in availableTenants" :key="tenant.id">
                <input type="checkbox" :value="tenant.id" v-model="userForm.tenantAccess">
                <span>{{ tenant.name }}</span>
              </label>
            </div>
          </div>

          <div class="form-divider"></div>

          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="userForm.requireTwoFactor">
              <span>Require Two-Factor Authentication</span>
            </label>
          </div>
          <div class="form-group">
            <label class="checkbox-row">
              <input type="checkbox" v-model="userForm.sendWelcomeEmail">
              <span>Send welcome email with login credentials</span>
            </label>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveUser" :disabled="!canSave">
            {{ editingUser ? 'Save Changes' : 'Create Admin' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  Users as UsersIcon, CheckCircle as CheckCircleIcon, Shield as ShieldIcon,
  Activity as ActivityIcon, Search as SearchIcon, Plus as PlusIcon,
  Edit as EditIcon, Trash2 as TrashIcon, Key as KeyIcon, Eye as EyeIcon,
  X as XIcon, Info as InfoIcon
} from 'lucide-vue-next'

const searchQuery = ref('')
const roleFilter = ref('')
const showAddModal = ref(false)
const editingUser = ref(null)

const userForm = ref({
  name: '',
  email: '',
  password: '',
  role: 'admin',
  tenantAccess: [],
  requireTwoFactor: true,
  sendWelcomeEmail: true
})

const users = ref([
  { id: 1, name: 'System Root', email: 'root@callsign.io', role: 'superadmin', status: 'active', twoFactor: true, lastLogin: 'Just now', online: true },
  { id: 2, name: 'Alice Thompson', email: 'alice.t@callsign.io', role: 'superadmin', status: 'active', twoFactor: true, lastLogin: '2 hours ago', online: true },
  { id: 3, name: 'Bob Martinez', email: 'bob.m@callsign.io', role: 'admin', status: 'active', twoFactor: true, lastLogin: 'Yesterday', online: false, tenantAccess: ['acme', 'globex'] },
  { id: 4, name: 'Carol Wilson', email: 'carol.w@callsign.io', role: 'admin', status: 'active', twoFactor: false, lastLogin: '3 days ago', online: false, tenantAccess: ['acme'] },
  { id: 5, name: 'David Chen', email: 'david.c@callsign.io', role: 'readonly', status: 'inactive', twoFactor: false, lastLogin: '2 weeks ago', online: false, tenantAccess: [] },
])

const availableTenants = ref([
  { id: 'acme', name: 'ACME Corporation' },
  { id: 'globex', name: 'Globex Industries' },
  { id: 'initech', name: 'Initech Solutions' },
  { id: 'umbrella', name: 'Umbrella Corp' },
])

const filteredUsers = computed(() => {
  return users.value.filter(user => {
    const matchesSearch = !searchQuery.value || 
      user.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesRole = !roleFilter.value || user.role === roleFilter.value
    return matchesSearch && matchesRole
  })
})

const canSave = computed(() => userForm.value.name && userForm.value.email && (editingUser.value || userForm.value.password))

const getRoleLabel = (role) => {
  const labels = { superadmin: 'Super Admin', admin: 'System Admin', readonly: 'Read Only' }
  return labels[role] || role
}

const editUser = (user) => {
  editingUser.value = user
  userForm.value = { ...user, password: '', sendWelcomeEmail: false }
  showAddModal.value = true
}

const resetPassword = (user) => alert(`Password reset email sent to ${user.email}`)
const viewActivity = (user) => alert(`Viewing activity log for ${user.name}`)
const deleteUser = (user) => {
  if (confirm(`Delete system admin ${user.name}?`)) {
    users.value = users.value.filter(u => u.id !== user.id)
  }
}

const saveUser = () => {
  if (editingUser.value) {
    Object.assign(editingUser.value, userForm.value)
  } else {
    users.value.push({
      id: Date.now(),
      ...userForm.value,
      status: 'active',
      twoFactor: userForm.value.requireTwoFactor,
      lastLogin: 'Never',
      online: false
    })
  }
  closeModal()
}

const closeModal = () => {
  showAddModal.value = false
  editingUser.value = null
  userForm.value = { name: '', email: '', password: '', role: 'admin', tenantAccess: [], requireTwoFactor: true, sendWelcomeEmail: true }
}
</script>

<style scoped>
.system-users-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

/* Info Banner */
.info-banner {
  display: flex;
  gap: 12px;
  padding: 16px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-lg);
}
.info-icon { width: 20px; height: 20px; color: #2563eb; flex-shrink: 0; margin-top: 2px; }
.info-content { font-size: 13px; }
.info-content strong { display: block; margin-bottom: 4px; color: #1e40af; }
.info-content p { margin: 0; color: #3b82f6; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.total { background: #dbeafe; color: #2563eb; }
.stat-icon.active { background: #dcfce7; color: #16a34a; }
.stat-icon.superadmin { background: #fef3c7; color: #b45309; }
.stat-icon.online { background: #f3e8ff; color: #7c3aed; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Table */
.users-table-container { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }

.table-header { display: flex; gap: 12px; padding: 16px; border-bottom: 1px solid var(--border-color); }
.search-box { position: relative; flex: 1; max-width: 300px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 10px 12px 10px 38px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.filter-select { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }

.users-table { width: 100%; border-collapse: collapse; }
.users-table th { text-align: left; padding: 12px 16px; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 2px solid var(--border-color); background: var(--bg-app); }
.users-table td { padding: 14px 16px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.users-table tr:hover { background: var(--bg-app); }

.user-cell { display: flex; align-items: center; gap: 12px; }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; background: linear-gradient(135deg, #6366f1, #818cf8); color: white; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; position: relative; }
.user-avatar.online::after { content: ''; position: absolute; bottom: 0; right: 0; width: 10px; height: 10px; background: #22c55e; border-radius: 50%; border: 2px solid white; }
.user-info { display: flex; flex-direction: column; }
.user-name { font-weight: 600; }
.user-email { font-size: 12px; color: var(--text-muted); }

.role-badge { display: inline-flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 600; padding: 4px 10px; border-radius: 4px; text-transform: uppercase; }
.role-icon { width: 12px; height: 12px; }
.role-badge.superadmin { background: #fef3c7; color: #b45309; }
.role-badge.admin { background: #dbeafe; color: #2563eb; }
.role-badge.readonly { background: #f1f5f9; color: #64748b; }

.tenant-access { font-size: 12px; }
.tenant-access.limited { color: var(--text-muted); }

.tfa-badge { font-size: 11px; font-weight: 600; }
.tfa-badge.enabled { color: #16a34a; }
.tfa-badge:not(.enabled) { color: #dc2626; }

.last-login { font-size: 12px; color: var(--text-muted); }

.status-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; text-transform: uppercase; }
.status-badge.active { background: #dcfce7; color: #16a34a; }
.status-badge.inactive { background: #f1f5f9; color: #64748b; }

.actions-cell { display: flex; gap: 4px; }
.action-btn { width: 28px; height: 28px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover:not(:disabled) { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover:not(:disabled) { border-color: #ef4444; color: #ef4444; }
.action-btn:disabled { opacity: 0.3; cursor: not-allowed; }

.empty-state { text-align: center; padding: 48px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 16px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 520px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-hint { font-size: 11px; color: var(--text-muted); margin-top: 4px; }
.form-divider { height: 1px; background: var(--border-color); margin: 8px 0 16px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

.tenant-checkboxes { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; margin-top: 8px; }
.checkbox-item { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; padding: 8px; background: var(--bg-app); border-radius: 6px; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }
.icon { width: 20px; height: 20px; }

@media (max-width: 768px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .tenant-checkboxes { grid-template-columns: 1fr; }
}
</style>
