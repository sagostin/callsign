<template>
  <div class="location-manager">
    <div class="header-actions">
      <h3>Physical Locations</h3>
      <button class="btn-secondary small" @click="openAddModal">+ Add Location</button>
    </div>
    <p class="text-muted text-sm mb-lg">Manage site-specific settings for E911 and Caller ID.</p>
    
    <div v-for="(loc, index) in locations" :key="index" class="location-card">
       <div class="card-header">
         <span class="loc-name">{{ loc.name }}</span>
         <div class="actions">
            <button class="btn-link" @click="openEditModal(loc)">Edit</button>
            <button class="btn-link text-bad" @click="removeLocation(loc)">Remove</button>
         </div>
       </div>
       <div class="card-body">
          <div class="info-row">
             <span class="label">Main Number:</span>
             <span class="value">{{ loc.mainNumber }}</span>
          </div>
          <div class="info-row">
             <span class="label">Caller ID Name:</span>
             <span class="value">{{ loc.cidName }}</span>
          </div>
          <div class="info-row">
             <span class="label">Emergency Fallback:</span>
             <span class="value text-warn">{{ loc.fallback }}</span>
          </div>
          <div class="info-row">
             <span class="label">Address:</span>
             <span class="value">{{ loc.address }}</span>
          </div>
       </div>
    </div>

    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Location' : 'Add Location' }}</h3>
          <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label>Location Name</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. HQ Office" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Main Number (DID)</label>
              <input v-model="form.mainNumber" class="input-field" placeholder="+15551234567" />
            </div>
            <div class="form-group">
              <label>Emergency Fallback</label>
              <input v-model="form.fallback" class="input-field" placeholder="+15559876543" />
            </div>
          </div>

          <div class="form-group">
            <label>Caller ID Name</label>
            <input v-model="form.cidName" class="input-field" placeholder="ACME Corp" />
          </div>

          <div class="form-group">
            <label>Address</label>
            <input v-model="form.address" class="input-field" placeholder="123 Main St, City, State ZIP" />
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveLocation">{{ isEditing ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'
import { tenantSettingsAPI } from '../../services/api'

const toast = inject('toast')

// State
const locations = ref([])
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref(null)

// Form defaults - atomic predictable shape
const defaultForm = () => ({
  name: '',
  mainNumber: '',
  cidName: '',
  fallback: '',
  address: ''
})

const form = ref(defaultForm())

// Fetch all locations on mount
onMounted(fetchLocations)

// Load locations from API
async function fetchLocations() {
  try {
    const res = await tenantSettingsAPI.listLocations()
    const data = res.data?.locations || res.data || []
    locations.value = (Array.isArray(data) ? data : []).map(parseLocation)
  } catch (err) {
    console.error('Failed to load locations:', err)
    locations.value = []
  }
}

// Parse API location to component shape
function parseLocation(loc) {
  return {
    id: loc.id,
    name: loc.name,
    mainNumber: loc.main_number || loc.did || '',
    cidName: loc.caller_id_name || loc.cid_name || '',
    fallback: loc.fallback_number || loc.failover_did || '',
    address: loc.address || ''
  }
}

// Open modal for adding new location
function openAddModal() {
  form.value = defaultForm()
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

// Open modal for editing existing location
function openEditModal(loc) {
  form.value = {
    name: loc.name,
    mainNumber: loc.mainNumber,
    cidName: loc.cidName,
    fallback: loc.fallback,
    address: loc.address
  }
  isEditing.value = true
  editingId.value = loc.id
  showModal.value = true
}

// Close modal and reset state
function closeModal() {
  showModal.value = false
  isEditing.value = false
  editingId.value = null
}

// Save location - handles both create and update
async function saveLocation() {
  if (!form.value.name.trim()) {
    toast.error('Location name is required')
    return
  }

  const payload = {
    name: form.value.name.trim(),
    main_number: form.value.mainNumber.trim(),
    caller_id_name: form.value.cidName.trim(),
    fallback_number: form.value.fallback.trim(),
    address: form.value.address.trim()
  }

  try {
    if (isEditing.value) {
      await tenantSettingsAPI.updateLocation(editingId.value, payload)
    } else {
      await tenantSettingsAPI.createLocation(payload)
    }
    closeModal()
    await fetchLocations()
  } catch (err) {
    console.error('Failed to save location:', err)
    toast.error('Failed to save location. Please try again.')
  }
}

// Remove location with confirmation
async function removeLocation(loc) {
  if (!confirm(`Delete location "${loc.name}"? This cannot be undone.`)) {
    return
  }

  try {
    await tenantSettingsAPI.deleteLocation(loc.id)
    await fetchLocations()
  } catch (err) {
    console.error('Failed to remove location:', err)
    toast.error('Failed to remove location. Please try again.')
  }
}
</script>

<style scoped>
.location-manager { padding-top: 8px; }

.header-actions { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.header-actions h3 { margin: 0; font-size: 16px; font-weight: 600; }

.mb-lg { margin-bottom: 24px; }

.location-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 16px;
  margin-bottom: 16px;
  background: white;
}

.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; border-bottom: 1px dashed var(--border-color); padding-bottom: 8px; }
.loc-name { font-weight: 600; font-size: 14px; color: var(--text-primary); }

.card-body { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }

.info-row { display: flex; flex-direction: column; gap: 2px; }
.label { font-size: 10px; text-transform: uppercase; color: var(--text-muted); font-weight: 700; }
.value { font-size: 13px; color: var(--text-main); }
.text-warn { color: #d97706; }

.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 4px 12px; border-radius: 4px; font-size: 12px; cursor: pointer; }
.btn-primary { background: var(--primary-color); color: white; border: 1px solid var(--primary-color); padding: 4px 12px; border-radius: 4px; font-size: 12px; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 600; cursor: pointer; margin-left: 8px; }
.text-bad { color: var(--status-bad); }

.modal-overlay { position: fixed; inset: 0; z-index: 50; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); }
.modal-card { background: white; border-radius: 8px; width: 100%; max-width: 500px; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1); }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; font-weight: 600; }
.modal-body { padding: 20px; display: flex; flex-direction: column; gap: 16px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 8px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group label { font-size: 12px; font-weight: 600; color: var(--text-primary); }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.input-field { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 4px; font-size: 14px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }

.btn-icon { background: none; border: none; cursor: pointer; padding: 4px; display: flex; }
</style>