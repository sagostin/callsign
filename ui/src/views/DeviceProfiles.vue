<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Device Profiles</h2>
      <p class="text-muted text-sm">Group devices with common settings and configuration overrides.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="showModal = true">+ New Profile</button>
    </div>
  </div>

  <div class="profiles-grid">
    <div class="profile-card" v-for="profile in profiles" :key="profile.id">
      <div class="profile-header">
        <div class="profile-icon" :style="{ background: profile.color }">
          {{ profile.name.charAt(0).toUpperCase() }}
        </div>
        <div class="profile-info">
          <h4>{{ profile.name }}</h4>
          <span class="device-count">{{ profile.deviceCount }} devices</span>
        </div>
        <div class="profile-actions">
          <button class="btn-icon" @click="editProfile(profile)"><EditIcon class="icon-sm" /></button>
          <button class="btn-icon" @click="deleteProfile(profile)"><TrashIcon class="icon-sm text-bad" /></button>
        </div>
      </div>
      
      <div class="profile-settings">
        <div class="setting-row">
          <span class="setting-label">Timezone</span>
          <span class="setting-value">{{ profile.timezone || 'Default' }}</span>
        </div>
        <div class="setting-row">
          <span class="setting-label">Language</span>
          <span class="setting-value">{{ profile.language || 'English' }}</span>
        </div>
        <div class="setting-row" v-if="profile.templateName">
          <span class="setting-label">Template</span>
          <span class="setting-value template-tag">{{ profile.templateName }}</span>
        </div>
      </div>

      <div class="profile-features">
        <span class="feature-tag" :class="{ enabled: profile.callWaiting }">
          <CheckIcon v-if="profile.callWaiting" class="icon-xs" />
          <XIcon v-else class="icon-xs" />
          Call Waiting
        </span>
        <span class="feature-tag" :class="{ enabled: profile.blf }">
          <CheckIcon v-if="profile.blf" class="icon-xs" />
          <XIcon v-else class="icon-xs" />
          BLF
        </span>
        <span class="feature-tag" :class="{ enabled: profile.directory }">
          <CheckIcon v-if="profile.directory" class="icon-xs" />
          <XIcon v-else class="icon-xs" />
          Directory
        </span>
      </div>
    </div>

    <div class="profile-card add-card" @click="showModal = true">
      <PlusIcon class="add-icon" />
      <span>Add Device Profile</span>
    </div>
  </div>

  <!-- Profile Modal -->
  <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ isEditing ? 'Edit Profile' : 'New Device Profile' }}</h3>
        <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-row">
          <div class="form-group flex-1">
            <label>Profile Name</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. Sales Floor Phones">
          </div>
          <div class="form-group">
            <label>Color</label>
            <div class="color-picker">
              <button v-for="c in colorOptions" :key="c" 
                class="color-swatch" 
                :style="{ background: c }"
                :class="{ selected: form.color === c }"
                @click="form.color = c">
              </button>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Regional Settings</h4>
          <div class="form-row">
            <div class="form-group">
              <label>Timezone</label>
              <select v-model="form.timezone" class="input-field">
                <option value="">Use Device Default</option>
                <option value="America/Los_Angeles">Pacific Time</option>
                <option value="America/Denver">Mountain Time</option>
                <option value="America/Chicago">Central Time</option>
                <option value="America/New_York">Eastern Time</option>
                <option value="Europe/London">London</option>
                <option value="Europe/Paris">Paris</option>
                <option value="Asia/Tokyo">Tokyo</option>
              </select>
            </div>
            <div class="form-group">
              <label>Language</label>
              <select v-model="form.language" class="input-field">
                <option value="">Device Default</option>
                <option value="en">English</option>
                <option value="es">Spanish</option>
                <option value="fr">French</option>
                <option value="de">German</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Date Format</label>
              <select v-model="form.dateFormat" class="input-field">
                <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                <option value="YYYY-MM-DD">YYYY-MM-DD</option>
              </select>
            </div>
            <div class="form-group">
              <label>Time Format</label>
              <select v-model="form.timeFormat" class="input-field">
                <option value="12h">12 Hour</option>
                <option value="24h">24 Hour</option>
              </select>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Audio Settings</h4>
          <div class="form-row">
            <div class="form-group">
              <label>Ring Volume</label>
              <input type="range" v-model="form.ringVolume" min="0" max="10" class="range-input">
            </div>
            <div class="form-group">
              <label>Handset Volume</label>
              <input type="range" v-model="form.handsetVolume" min="0" max="10" class="range-input">
            </div>
          </div>
          <label class="toggle-row">
            <input type="checkbox" v-model="form.vad">
            <span>Voice Activity Detection (VAD)</span>
          </label>
          <label class="toggle-row">
            <input type="checkbox" v-model="form.echoCancellation">
            <span>Echo Cancellation</span>
          </label>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Feature Toggles</h4>
          <div class="feature-toggles">
            <label class="toggle-row">
              <input type="checkbox" v-model="form.callWaiting">
              <span>Call Waiting</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" v-model="form.blf">
              <span>Busy Lamp Field (BLF)</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" v-model="form.directory">
              <span>Directory Access</span>
            </label>
            <label class="toggle-row">
              <input type="checkbox" v-model="form.dnd">
              <span>Do Not Disturb</span>
            </label>
          </div>
        </div>

        <div class="divider"></div>

        <div class="form-section">
          <h4>Template Override</h4>
          <div class="form-group">
            <label>Provisioning Template</label>
            <select v-model="form.templateId" class="input-field">
              <option :value="null">Use Default for Device Model</option>
              <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
            </select>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="closeModal">Cancel</button>
        <button class="btn-primary" @click="saveProfile" :disabled="!form.name">
          {{ isEditing ? 'Save Changes' : 'Create Profile' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'
import { Edit as EditIcon, Trash2 as TrashIcon, X as XIcon, Check as CheckIcon, Plus as PlusIcon } from 'lucide-vue-next'
import { deviceProfilesAPI, deviceTemplatesAPI } from '@/services/api'

const toast = inject('toast')
const profiles = ref([])
const templates = ref([])
const showModal = ref(false)
const isEditing = ref(false)

const colorOptions = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#64748b']

const defaultForm = {
  id: null,
  name: '',
  color: '#6366f1',
  timezone: '',
  language: '',
  dateFormat: 'MM/DD/YYYY',
  timeFormat: '12h',
  ringVolume: 7,
  handsetVolume: 7,
  vad: true,
  echoCancellation: true,
  callWaiting: true,
  blf: true,
  directory: true,
  dnd: false,
  templateId: null
}

const form = ref({ ...defaultForm })

onMounted(async () => {
  await Promise.all([fetchProfiles(), fetchTemplates()])
})

async function fetchProfiles() {
  try {
    const response = await deviceProfilesAPI.list()
    profiles.value = (response.data?.data || response.data || []).map(p => ({
      id: p.id,
      name: p.name,
      color: p.color || '#6366f1',
      timezone: p.timezone,
      language: p.language,
      dateFormat: p.date_format,
      timeFormat: p.time_format,
      ringVolume: p.ring_volume || 7,
      handsetVolume: p.handset_volume || 7,
      vad: p.vad !== false,
      echoCancellation: p.echo_cancellation !== false,
      callWaiting: p.call_waiting !== false,
      blf: p.blf !== false,
      directory: p.directory !== false,
      dnd: p.dnd || false,
      templateId: p.template_id,
      templateName: p.template_name,
      deviceCount: p.device_count || 0
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load device profiles')
  }
}

async function fetchTemplates() {
  try {
    const response = await deviceTemplatesAPI.list()
    templates.value = (response.data?.data || response.data || []).map(t => ({
      id: t.id,
      name: t.name
    }))
  } catch (error) {
    console.error('Failed to load templates', error)
  }
}

function editProfile(profile) {
  form.value = { ...profile }
  isEditing.value = true
  showModal.value = true
}

async function deleteProfile(profile) {
  if (confirm(`Delete profile "${profile.name}"? Devices will be unassigned from this profile.`)) {
    try {
      await deviceProfilesAPI.delete(profile.id)
      toast?.success(`Profile "${profile.name}" deleted`)
      await fetchProfiles()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete profile')
    }
  }
}

async function saveProfile() {
  try {
    const payload = {
      name: form.value.name,
      color: form.value.color,
      timezone: form.value.timezone,
      language: form.value.language,
      date_format: form.value.dateFormat,
      time_format: form.value.timeFormat,
      ring_volume: form.value.ringVolume,
      handset_volume: form.value.handsetVolume,
      vad: form.value.vad,
      echo_cancellation: form.value.echoCancellation,
      call_waiting: form.value.callWaiting,
      blf: form.value.blf,
      directory: form.value.directory,
      dnd: form.value.dnd,
      template_id: form.value.templateId
    }
    
    if (isEditing.value) {
      await deviceProfilesAPI.update(form.value.id, payload)
      toast?.success('Profile updated')
    } else {
      await deviceProfilesAPI.create(payload)
      toast?.success('Profile created')
    }
    await fetchProfiles()
    closeModal()
  } catch (error) {
    toast?.error(error.message, 'Failed to save profile')
  }
}

function closeModal() {
  showModal.value = false
  isEditing.value = false
  form.value = { ...defaultForm }
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.profiles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-lg);
}

.profile-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
}

.profile-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 18px;
}

.profile-info {
  flex: 1;
}

.profile-info h4 {
  font-size: 15px;
  font-weight: 600;
  margin: 0 0 2px;
}

.device-count {
  font-size: 12px;
  color: var(--text-muted);
}

.profile-actions {
  display: flex;
  gap: 4px;
}

.profile-settings {
  padding: 12px 16px;
}

.setting-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}

.setting-label {
  color: var(--text-muted);
}

.setting-value {
  font-weight: 500;
}

.template-tag {
  background: var(--primary-color);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
}

.profile-features {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 12px 16px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-app);
}

.feature-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  background: #fee2e2;
  color: #dc2626;
}

.feature-tag.enabled {
  background: #dcfce7;
  color: #16a34a;
}

.add-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  border: 2px dashed var(--border-color);
  background: transparent;
  cursor: pointer;
  color: var(--text-muted);
  transition: all 0.2s;
}

.add-card:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  background: rgba(99, 102, 241, 0.05);
}

.add-icon {
  width: 32px;
  height: 32px;
  margin-bottom: 8px;
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
  max-width: 560px;
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

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.form-group.flex-1 { flex: 1; }

.form-section h4 {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--text-primary);
}

label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
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

.color-picker {
  display: flex;
  gap: 6px;
}

.color-swatch {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 2px solid transparent;
  cursor: pointer;
}

.color-swatch.selected {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 2px white inset;
}

.toggle-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  cursor: pointer;
  padding: 6px 0;
  font-weight: 400;
  text-transform: none;
  color: var(--text-main);
}

.range-input {
  width: 100%;
  accent-color: var(--primary-color);
}

.feature-toggles {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
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

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
}
.btn-icon:hover { color: var(--text-primary); }

.text-bad { color: #dc2626; }
.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 12px; height: 12px; }
</style>
