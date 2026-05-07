<template>
  <div class="profiles-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Extension Profiles</h2>
        <p class="text-muted text-sm">Manage pre-configured extension settings templates.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="openCreateModal">
          <PlusIcon class="btn-icon" /> New Profile
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ profiles.length }}</div>
        <div class="stat-label">Total Profiles</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ assignedExtensions }}</div>
        <div class="stat-label">Extensions Using</div>
      </div>
    </div>

    <!-- Profiles Grid -->
    <div v-if="loading" class="loading-state">
      <p class="text-muted">Loading profiles...</p>
    </div>
    <div v-else-if="profiles.length === 0" class="empty-state">
      <UsersIcon class="empty-icon" />
      <h4>No Extension Profiles</h4>
      <p class="text-muted text-sm">Create profiles to define permission sets and calling rules for extensions.</p>
      <button class="btn-primary" @click="openCreateModal">Create First Profile</button>
    </div>
    <div v-else class="profiles-grid">
      <div v-for="profile in profiles" :key="profile.id" class="profile-card">
        <div class="profile-header">
          <div class="profile-icon" :style="{ background: profile.color }">
            {{ profile.name.charAt(0) }}
          </div>
          <div class="profile-info">
            <h4>{{ profile.name }}</h4>
            <span class="profile-meta">{{ profile.extensionCount }} extensions</span>
          </div>
        </div>

        <div class="profile-permissions">
          <div v-for="(perm, key) in permissionLabels" :key="key" class="perm-row">
            <span class="perm-label">{{ perm.label }}</span>
            <span class="perm-value" :class="{ enabled: profile.permissions[key], blocked: !profile.permissions[key] }">
              {{ profile.permissions[key] ? perm.enabled : perm.disabled }}
            </span>
          </div>
        </div>

        <div class="profile-footer">
          <button class="btn-link" @click="editProfile(profile)">Edit</button>
          <button class="btn-link" @click="duplicateProfile(profile)">Duplicate</button>
          <button class="btn-link danger" @click="deleteProfile(profile)">Delete</button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div class="modal-overlay" v-if="showModal" @click.self="closeModal">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Profile' : 'New Extension Profile' }}</h3>
          <button class="close-btn" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <div class="form-section">
            <h4>Basic Information</h4>
            <div class="form-row">
              <div class="form-group">
                <label>Profile Name</label>
                <input v-model="form.name" class="input-field" placeholder="e.g. Sales Team">
              </div>
              <div class="form-group">
                <label>Profile Color</label>
                <div class="color-picker">
                  <button
                    v-for="c in colorOptions"
                    :key="c"
                    class="color-swatch"
                    :style="{ background: c }"
                    :class="{ selected: form.color === c }"
                    @click="form.color = c"
                  >
                    <CheckIcon v-if="form.color === c" class="check-icon" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="form-section">
            <h4>Permissions</h4>
            <div class="perm-toggles">
              <label v-for="(perm, key) in permissionLabels" :key="key" class="toggle-row">
                <input type="checkbox" v-model="form.permissions[key]">
                <span>{{ perm.label }}</span>
              </label>
            </div>
          </div>

          <!-- Call Handling Rules (only when editing) -->
          <div class="form-section" v-if="isEditing && form.id">
            <h4>Call Handling Rules</h4>
            <p class="help-text">Rules applied to all extensions using this profile.</p>
            <CallHandlingRules
              :profileId="form.id"
              :api="extensionProfilesAPI"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveProfile" :disabled="saving || !form.name">
            {{ saving ? 'Saving...' : 'Save Profile' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { Plus as PlusIcon, Users as UsersIcon, Check as CheckIcon } from 'lucide-vue-next'
import CallHandlingRules from '../../components/common/CallHandlingRules.vue'
import { extensionProfilesAPI } from '@/services/api'

const toast = inject('toast')

const profiles = ref([])
const loading = ref(false)
const saving = ref(false)
const showModal = ref(false)
const isEditing = ref(false)

const permissionLabels = {
  outbound: { label: 'Outbound Calls', enabled: 'Allowed', disabled: 'Blocked' },
  international: { label: 'International Dialing', enabled: 'Allowed', disabled: 'Blocked' },
  recording: { label: 'Call Recording', enabled: 'Enabled', disabled: 'Disabled' },
  portal: { label: 'Portal Access', enabled: 'Allowed', disabled: 'Blocked' },
  voicemail: { label: 'Voicemail Config', enabled: 'Allowed', disabled: 'Blocked' }
}

const colorOptions = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#64748b']

const defaultForm = () => ({
  id: null,
  name: '',
  color: '#6366f1',
  permissions: { outbound: true, international: false, recording: true, portal: true, voicemail: true }
})

const form = ref(defaultForm())

const assignedExtensions = computed(() => profiles.value.reduce((sum, p) => sum + (p.extensionCount || 0), 0))

onMounted(() => loadProfiles())

async function loadProfiles() {
  loading.value = true
  try {
    const response = await extensionProfilesAPI.list()
    profiles.value = (response.data || []).map(p => ({
      id: p.id,
      name: p.name,
      color: p.color || '#6366f1',
      extensionCount: p.extension_count || 0,
      permissions: p.permissions || {}
    }))
  } catch (error) {
    toast?.error(error.message || 'Failed to load profiles')
    profiles.value = []
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  form.value = defaultForm()
  isEditing.value = false
  showModal.value = true
}

const editProfile = (profile) => {
  form.value = {
    id: profile.id,
    name: profile.name,
    color: profile.color,
    permissions: { ...profile.permissions }
  }
  isEditing.value = true
  showModal.value = true
}

const duplicateProfile = async (profile) => {
  try {
    const payload = {
      name: profile.name + ' (Copy)',
      color: profile.color,
      permissions: { ...profile.permissions }
    }
    await extensionProfilesAPI.create(payload)
    toast?.success('Profile duplicated')
    await loadProfiles()
  } catch (error) {
    toast?.error(error.message || 'Failed to duplicate profile')
  }
}

const deleteProfile = async (profile) => {
  if (!confirm(`Delete profile "${profile.name}"? Extensions using it will be unassigned.`)) return
  try {
    await extensionProfilesAPI.delete(profile.id)
    toast?.success(`Profile "${profile.name}" deleted`)
    await loadProfiles()
  } catch (error) {
    toast?.error(error.message || 'Failed to delete profile')
  }
}

const saveProfile = async () => {
  if (!form.value.name) {
    toast?.error('Profile name is required')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      color: form.value.color,
      permissions: form.value.permissions
    }
    if (isEditing.value && form.value.id) {
      await extensionProfilesAPI.update(form.value.id, payload)
      toast?.success('Profile updated')
    } else {
      await extensionProfilesAPI.create(payload)
      toast?.success('Profile created')
    }
    closeModal()
    await loadProfiles()
  } catch (error) {
    toast?.error(error.message || 'Failed to save profile')
  } finally {
    saving.value = false
  }
}

const closeModal = () => {
  showModal.value = false
  isEditing.value = false
  form.value = defaultForm()
}
</script>

<style scoped>
.profiles-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.btn-primary { display: flex; align-items: center; gap: 6px; background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.loading-state, .empty-state { text-align: center; padding: 60px 20px; background: white; border: 1px solid var(--border-color); border-radius: 8px; }
.empty-state .empty-icon { width: 48px; height: 48px; color: var(--text-muted); margin-bottom: 16px; }
.empty-state h4 { margin: 0 0 8px; font-size: 16px; }
.empty-state p { margin: 0 0 16px; }

.profiles-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }

.profile-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}
.profile-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }

.profile-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.profile-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 16px;
  flex-shrink: 0;
}
.profile-info h4 { margin: 0 0 4px; font-size: 14px; }
.profile-meta { font-size: 11px; color: var(--text-muted); }

.profile-permissions { padding: 14px; }
.perm-row { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #f1f5f9; }
.perm-row:last-child { border-bottom: none; }
.perm-label { font-size: 11px; color: var(--text-muted); }
.perm-value { font-size: 12px; font-weight: 500; }
.perm-value.enabled { color: #22c55e; }
.perm-value.blocked { color: #94a3b8; }

.profile-footer {
  display: flex;
  gap: 12px;
  padding: 10px 14px;
  background: #f8fafc;
  border-top: 1px solid var(--border-color);
}
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 500; cursor: pointer; }
.btn-link.danger { color: #ef4444; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 500px; max-height: 85vh; overflow: hidden; display: flex; flex-direction: column; }
.modal-card.large { max-width: 700px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; }

.form-section { margin-bottom: 20px; }
.form-section h4 { font-size: 12px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; margin: 0 0 12px; padding-bottom: 8px; border-bottom: 1px solid #f1f5f9; }
.form-group { margin-bottom: 12px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.input-field:focus { border-color: var(--primary-color); outline: none; }
.help-text { font-size: 11px; color: var(--text-muted); margin: 4px 0; }

/* Color Picker */
.color-picker { display: flex; flex-wrap: wrap; gap: 8px; }
.color-swatch { width: 32px; height: 32px; border-radius: 6px; border: 2px solid transparent; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s; }
.color-swatch:hover { transform: scale(1.1); }
.color-swatch.selected { border-color: var(--text-primary); box-shadow: 0 0 0 2px white, 0 0 0 4px var(--text-primary); }
.check-icon { width: 16px; height: 16px; color: white; filter: drop-shadow(0 1px 2px rgba(0,0,0,0.3)); }

/* Permission Toggles */
.perm-toggles { display: flex; flex-direction: column; gap: 8px; }
.toggle-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; padding: 8px; border-radius: 6px; transition: background 0.15s; }
.toggle-row:hover { background: #f8fafc; }
.toggle-row input[type="checkbox"] { width: 18px; height: 18px; accent-color: var(--primary-color); }
</style>
