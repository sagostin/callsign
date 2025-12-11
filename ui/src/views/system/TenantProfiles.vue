<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Tenant Profiles</h2>
      <p class="text-muted text-sm">Manage service plans and default limits for new tenants.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="showModal = true">+ New Profile</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="profiles" actions>
    <template #tenantCount="{ value }">
      <span class="badge tenant">{{ value }} tenants</span>
    </template>
    <template #limits="{ row }">
      <div class="limits-cell">
        <span><strong>{{ row.limits.extensions }}</strong> ext</span>
        <span class="divider-dot">•</span>
        <span><strong>{{ row.limits.disk }}</strong> GB</span>
        <span class="divider-dot">•</span>
        <span><strong>{{ row.limits.channels }}</strong> ch</span>
      </div>
    </template>
    <template #features="{ row }">
      <div class="feature-tags">
        <span v-if="row.features.hospitality" class="tag tag-green">Hospitality</span>
        <span v-if="row.features.recording" class="tag tag-blue">Recording</span>
        <span v-if="row.features.fax" class="tag tag-purple">Fax</span>
        <span v-if="!row.features.hospitality && !row.features.recording && !row.features.fax" class="tag tag-muted">Basic</span>
      </div>
    </template>
    <template #actions="{ row }">
      <button class="btn-link" @click="duplicateProfile(row)">Duplicate</button>
      <button class="btn-link" @click="editProfile(row)">Edit</button>
      <button class="btn-link text-bad" @click="deleteProfile(row)">Delete</button>
    </template>
  </DataTable>

  <!-- Modal for Create/Edit -->
  <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
    <div class="modal-card">
      <h3>{{ isEditing ? 'Edit Profile' : 'New Tenant Profile' }}</h3>
      
      <div class="form-section">
        <div class="form-row">
          <div class="form-group">
            <label>Profile Name</label>
            <input v-model="activeProfile.name" class="input-field" placeholder="e.g. Enterprise" />
          </div>
          <div class="form-group">
            <label>Profile Code</label>
            <input v-model="activeProfile.code" class="input-field code" placeholder="e.g. enterprise" />
          </div>
        </div>
      </div>

      <div class="divider"></div>

      <div class="form-section">
        <h4>Resource Limits</h4>
        <div class="form-row three">
          <div class="form-group">
            <label>Max Extensions</label>
            <input v-model.number="activeProfile.limits.extensions" type="number" class="input-field" />
          </div>
          <div class="form-group">
            <label>Disk Space (GB)</label>
            <input v-model.number="activeProfile.limits.disk" type="number" class="input-field" />
          </div>
          <div class="form-group">
            <label>Concurrent Channels</label>
            <input v-model.number="activeProfile.limits.channels" type="number" class="input-field" />
          </div>
        </div>
      </div>

      <div class="divider"></div>

      <div class="form-section">
        <h4>Features Included</h4>
        <div class="feature-checkboxes">
          <label class="checkbox-label">
            <input type="checkbox" v-model="activeProfile.features.hospitality" />
            <span>Hospitality Mode</span>
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="activeProfile.features.recording" />
            <span>Call Recording</span>
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="activeProfile.features.fax" />
            <span>Fax Support</span>
          </label>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn-secondary" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveProfile">{{ isEditing ? 'Save Changes' : 'Create Profile' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'

const columns = [
  { key: 'name', label: 'Profile Name' },
  { key: 'code', label: 'Code', width: '120px' },
  { key: 'tenantCount', label: 'Usage' },
  { key: 'limits', label: 'Limits' },
  { key: 'features', label: 'Features' }
]

const profiles = ref([
  { 
    id: 1, 
    name: 'Basic (SMB)', 
    code: 'basic', 
    tenantCount: 15,
    limits: { extensions: 10, disk: 5, channels: 5 },
    features: { hospitality: false, recording: false, fax: false }
  },
  { 
    id: 2, 
    name: 'Hospitality Standard', 
    code: 'hotel', 
    tenantCount: 4,
    limits: { extensions: 200, disk: 50, channels: 50 },
    features: { hospitality: true, recording: false, fax: false }
  },
  { 
    id: 3, 
    name: 'Enterprise', 
    code: 'enterprise', 
    tenantCount: 2,
    limits: { extensions: 1000, disk: 500, channels: 200 },
    features: { hospitality: true, recording: true, fax: true }
  },
])

const showModal = ref(false)
const isEditing = ref(false)
const activeProfile = ref({
  name: '',
  code: '',
  limits: { extensions: 10, disk: 5, channels: 5 },
  features: { hospitality: false, recording: false, fax: false }
})

const editProfile = (profile) => {
  activeProfile.value = JSON.parse(JSON.stringify(profile))
  isEditing.value = true
  showModal.value = true
}

const duplicateProfile = (profile) => {
  activeProfile.value = {
    ...JSON.parse(JSON.stringify(profile)),
    id: null,
    name: `${profile.name} (Copy)`,
    code: `${profile.code}_copy`,
    tenantCount: 0
  }
  isEditing.value = false
  showModal.value = true
}

const deleteProfile = (profile) => {
  if (confirm(`Delete profile "${profile.name}"?`)) {
    profiles.value = profiles.value.filter(p => p.id !== profile.id)
  }
}

const saveProfile = () => {
  if (isEditing.value) {
    const idx = profiles.value.findIndex(p => p.id === activeProfile.value.id)
    if (idx !== -1) profiles.value[idx] = { ...activeProfile.value }
  } else {
    profiles.value.push({
      ...activeProfile.value,
      id: Date.now(),
      tenantCount: 0
    })
  }
  showModal.value = false
  activeProfile.value = {
    name: '',
    code: '',
    limits: { extensions: 10, disk: 5, channels: 5 },
    features: { hospitality: false, recording: false, fax: false }
  }
  isEditing.value = false
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
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

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  cursor: pointer;
  font-weight: 500;
}

.text-bad { color: var(--status-bad); }

/* Limits cell styling */
.limits-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: var(--text-xs);
  color: var(--text-main);
}
.limits-cell strong {
  font-weight: 700;
  color: var(--text-primary);
}
.divider-dot {
  color: var(--text-muted);
  font-size: 8px;
}

/* Feature tags */
.feature-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
.tag {
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}
.tag-green { background: #dcfce7; color: #166534; }
.tag-blue { background: #dbeafe; color: #1e40af; }
.tag-purple { background: #f3e8ff; color: #7c3aed; }
.tag-muted { background: var(--bg-secondary); color: var(--text-muted); }

.badge {
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 600;
}
.badge.tenant {
  background: #eef2ff;
  color: #4f46e5;
}

/* Modal Overlay */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(4px);
  padding: 24px;
}

.modal-card {
  background: white;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 520px;
  padding: 24px;
}

.modal-card h3 {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 20px;
  color: var(--text-primary);
}

.modal-card h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.form-section {
  margin-bottom: 12px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.form-row.three {
  grid-template-columns: 1fr 1fr 1fr;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 14px;
}
.input-field.code {
  font-family: monospace;
}
.input-field:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.divider {
  height: 1px;
  background: var(--border-color);
  margin: 16px 0;
}

.feature-checkboxes {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 400;
  color: var(--text-main);
  cursor: pointer;
  text-transform: none;
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--primary-color);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}
</style>
