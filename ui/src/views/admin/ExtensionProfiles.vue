<template>
  <div class="profiles-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Extension Profiles</h2>
        <p class="text-muted text-sm">Manage pre-configured extension settings templates.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="showCreateModal = true">
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
      <div class="stat-card">
        <div class="stat-value">{{ activeProfiles }}</div>
        <div class="stat-label">Active</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ assignedExtensions }}</div>
        <div class="stat-label">Extensions Using</div>
      </div>
    </div>

    <!-- Profiles Grid -->
    <div class="profiles-grid">
      <div v-for="profile in profiles" :key="profile.id" class="profile-card" :class="{ default: profile.isDefault }">
        <div class="profile-header">
          <div class="profile-info">
            <h4>{{ profile.name }}</h4>
            <span class="profile-meta">{{ profile.extensionsCount }} extensions</span>
          </div>
          <span v-if="profile.isDefault" class="default-badge">Default</span>
        </div>
        
        <div class="profile-settings">
          <div class="setting-item">
            <span class="setting-label">Caller ID</span>
            <span class="setting-value">{{ profile.callerId || 'Extension CID' }}</span>
          </div>
          <div class="setting-item">
            <span class="setting-label">Recording</span>
            <span class="setting-value" :class="profile.recording ? 'yes' : 'no'">
              {{ profile.recording ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
          <div class="setting-item">
            <span class="setting-label">Voicemail</span>
            <span class="setting-value">{{ profile.voicemailEnabled ? 'Enabled' : 'Disabled' }}</span>
          </div>
          <div class="setting-item">
            <span class="setting-label">Ring Timeout</span>
            <span class="setting-value">{{ profile.ringTimeout }}s</span>
          </div>
        </div>
        
        <div class="profile-features">
          <span v-if="profile.callWaiting" class="feature-badge">Call Waiting</span>
          <span v-if="profile.dnd" class="feature-badge">DND</span>
          <span v-if="profile.callForward" class="feature-badge">Call Forward</span>
          <span v-if="profile.intercom" class="feature-badge">Intercom</span>
        </div>

        <div class="profile-footer">
          <button class="btn-link" @click="editProfile(profile)">Edit</button>
          <button class="btn-link" @click="duplicateProfile(profile)">Duplicate</button>
          <button class="btn-link danger" @click="deleteProfile(profile)" :disabled="profile.isDefault">Delete</button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div class="modal-overlay" v-if="showCreateModal" @click.self="closeModal">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>{{ editingProfile ? 'Edit Profile' : 'New Extension Profile' }}</h3>
          <button class="close-btn" @click="closeModal">×</button>
        </div>
        <div class="modal-body">
          <div class="form-section">
            <h4>Basic Information</h4>
            <div class="form-row">
              <div class="form-group">
                <label>Profile Name</label>
                <input v-model="form.name" class="input-field" placeholder="Standard User">
              </div>
              <div class="form-group">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="form.isDefault">
                  Set as Default Profile
                </label>
              </div>
            </div>
            <div class="form-group">
              <label>Description</label>
              <input v-model="form.description" class="input-field" placeholder="Profile for standard office users">
            </div>
          </div>

          <div class="form-section">
            <h4>Caller ID Settings</h4>
            <div class="form-row">
              <div class="form-group">
                <label>Outbound Caller ID</label>
                <select v-model="form.callerIdType" class="input-field">
                  <option value="extension">Use Extension CID</option>
                  <option value="main">Use Main Number</option>
                  <option value="custom">Custom Number</option>
                </select>
              </div>
              <div class="form-group" v-if="form.callerIdType === 'custom'">
                <label>Custom CID Number</label>
                <input v-model="form.customCid" class="input-field" placeholder="+1 555-123-4567">
              </div>
            </div>
          </div>

          <div class="form-section">
            <h4>Call Settings</h4>
            <div class="form-row">
              <div class="form-group">
                <label>Ring Timeout (seconds)</label>
                <input type="number" v-model.number="form.ringTimeout" class="input-field" min="10" max="120">
              </div>
              <div class="form-group">
                <label>Max Concurrent Calls</label>
                <input type="number" v-model.number="form.maxCalls" class="input-field" min="1" max="10">
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="checkbox-label"><input type="checkbox" v-model="form.callWaiting"> Call Waiting</label>
              </div>
              <div class="form-group">
                <label class="checkbox-label"><input type="checkbox" v-model="form.intercom"> Intercom/Paging</label>
              </div>
            </div>
          </div>

          <div class="form-section">
            <h4>Recording & Voicemail</h4>
            <div class="form-row">
              <div class="form-group">
                <label>Call Recording</label>
                <select v-model="form.recording" class="input-field">
                  <option value="">Disabled</option>
                  <option value="all">Record All Calls</option>
                  <option value="inbound">Inbound Only</option>
                  <option value="outbound">Outbound Only</option>
                </select>
              </div>
              <div class="form-group">
                <label class="checkbox-label"><input type="checkbox" v-model="form.voicemailEnabled"> Voicemail Enabled</label>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveProfile">Save Profile</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Plus as PlusIcon } from 'lucide-vue-next'

const showCreateModal = ref(false)
const editingProfile = ref(null)

const profiles = ref([
  { 
    id: 1, name: 'Standard Office', isDefault: true, extensionsCount: 45,
    callerId: 'Extension CID', recording: false, voicemailEnabled: true, ringTimeout: 30,
    callWaiting: true, dnd: false, callForward: true, intercom: true
  },
  { 
    id: 2, name: 'Executive', isDefault: false, extensionsCount: 5,
    callerId: 'Main Number', recording: true, voicemailEnabled: true, ringTimeout: 20,
    callWaiting: true, dnd: true, callForward: true, intercom: false
  },
  { 
    id: 3, name: 'Call Center Agent', isDefault: false, extensionsCount: 25,
    callerId: 'Main Number', recording: true, voicemailEnabled: false, ringTimeout: 15,
    callWaiting: false, dnd: false, callForward: false, intercom: true
  },
  { 
    id: 4, name: 'Lobby Phone', isDefault: false, extensionsCount: 3,
    callerId: 'Extension CID', recording: false, voicemailEnabled: false, ringTimeout: 60,
    callWaiting: false, dnd: false, callForward: false, intercom: true
  },
])

const form = ref({
  name: '',
  description: '',
  isDefault: false,
  callerIdType: 'extension',
  customCid: '',
  ringTimeout: 30,
  maxCalls: 2,
  callWaiting: true,
  intercom: true,
  recording: '',
  voicemailEnabled: true
})

const activeProfiles = computed(() => profiles.value.filter(p => p.extensionsCount > 0).length)
const assignedExtensions = computed(() => profiles.value.reduce((sum, p) => sum + p.extensionsCount, 0))

const editProfile = (profile) => {
  editingProfile.value = profile
  form.value = { ...profile }
  showCreateModal.value = true
}

const duplicateProfile = (profile) => {
  const newProfile = { ...profile, id: Date.now(), name: profile.name + ' (Copy)', isDefault: false, extensionsCount: 0 }
  profiles.value.push(newProfile)
}

const deleteProfile = (profile) => {
  if (profile.isDefault) return
  if (confirm(`Delete profile "${profile.name}"?`)) {
    profiles.value = profiles.value.filter(p => p.id !== profile.id)
  }
}

const saveProfile = () => {
  if (editingProfile.value) {
    const idx = profiles.value.findIndex(p => p.id === editingProfile.value.id)
    if (idx !== -1) profiles.value[idx] = { ...form.value, id: editingProfile.value.id }
  } else {
    profiles.value.push({ ...form.value, id: Date.now(), extensionsCount: 0 })
  }
  closeModal()
}

const closeModal = () => {
  showCreateModal.value = false
  editingProfile.value = null
  form.value = { name: '', description: '', isDefault: false, callerIdType: 'extension', customCid: '', ringTimeout: 30, maxCalls: 2, callWaiting: true, intercom: true, recording: '', voicemailEnabled: true }
}
</script>

<style scoped>
.profiles-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.btn-primary { display: flex; align-items: center; gap: 6px; background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.profiles-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }

.profile-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}
.profile-card.default { border-left: 3px solid var(--primary-color); }
.profile-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }

.profile-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 14px;
  background: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}
.profile-info h4 { margin: 0 0 4px; font-size: 14px; }
.profile-meta { font-size: 11px; color: var(--text-muted); }
.default-badge { font-size: 9px; background: var(--primary-color); color: white; padding: 2px 8px; border-radius: 4px; text-transform: uppercase; font-weight: 600; }

.profile-settings { padding: 14px; }
.setting-item { display: flex; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid #f1f5f9; }
.setting-item:last-child { border-bottom: none; }
.setting-label { font-size: 11px; color: var(--text-muted); }
.setting-value { font-size: 12px; font-weight: 500; }
.setting-value.yes { color: #22c55e; }
.setting-value.no { color: #94a3b8; }

.profile-features { display: flex; flex-wrap: wrap; gap: 4px; padding: 0 14px 14px; }
.feature-badge { font-size: 9px; background: #eff6ff; color: #3b82f6; padding: 3px 8px; border-radius: 4px; font-weight: 500; }

.profile-footer {
  display: flex;
  gap: 12px;
  padding: 10px 14px;
  background: #f8fafc;
  border-top: 1px solid var(--border-color);
}
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 500; cursor: pointer; }
.btn-link.danger { color: #ef4444; }
.btn-link:disabled { color: #cbd5e1; cursor: not-allowed; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 500px; max-height: 85vh; overflow: hidden; display: flex; flex-direction: column; }
.modal-card.large { max-width: 600px; }
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
.checkbox-label { display: flex !important; align-items: center; gap: 8px; font-size: 12px !important; cursor: pointer; text-transform: none !important; }
</style>
