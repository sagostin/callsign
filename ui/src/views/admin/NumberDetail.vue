<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/numbers')">← Back to Numbers</button>
      <h2>Edit Number: {{ $route.params.id }}</h2>
    </div>
      <div class="header-actions">
      <button class="btn-danger" @click="releaseNumber">Release Number</button>
    </div>
  </div>

  <div class="form-container">
    <div v-if="loading" class="loading-state">Loading number details...</div>
    <div v-else-if="error" class="error-state">{{ error }}</div>

    <template v-else>
      <div class="form-section">
        <h3>Routing Configuration</h3>
        <div class="form-group">
          <label>Destination Type</label>
          <select v-model="form.destinationType" class="input-field">
            <option value="extension">Extension</option>
            <option value="ivr">IVR Menu</option>
            <option value="queue">Queue</option>
            <option value="voicemail">Voicemail Box</option>
            <option value="ring_group">Ring Group</option>
          </select>
        </div>

        <div class="form-group">
          <label>Destination Target</label>
          <select v-model="form.destinationTarget" class="input-field">
            <optgroup v-if="extensions.length" label="Extensions">
              <option v-for="ext in extensions" :key="ext.id" :value="String(ext.id)">
                {{ ext.extension }} - {{ ext.name || 'Unnamed' }}
              </option>
            </optgroup>
            <optgroup v-if="queues.length" label="Queues">
              <option v-for="q in queues" :key="q.id" :value="String(q.id)">
                {{ q.name }}
              </option>
            </optgroup>
            <optgroup v-if="ringGroups.length" label="Ring Groups">
              <option v-for="rg in ringGroups" :key="rg.id" :value="String(rg.id)">
                {{ rg.name }}
              </option>
            </optgroup>
            <optgroup v-if="ivrMenus.length" label="IVR Menus">
              <option v-for="menu in ivrMenus" :key="menu.id" :value="String(menu.id)">
                {{ menu.name }}
              </option>
            </optgroup>
            <optgroup v-if="voicemailBoxes.length" label="Voicemail Boxes">
              <option v-for="vm in voicemailBoxes" :key="vm.id" :value="String(vm.id)">
                {{ vm.name || vm.extension }}
              </option>
            </optgroup>
          </select>
        </div>
        
        <div class="form-group">
          <label>Caller ID Prefix (Optional)</label>
          <input v-model="form.callerIdPrefix" type="text" class="input-field" placeholder="e.g. Sales: ">
        </div>

        <div class="form-group checkbox-row">
          <div class="check-item">
            <input v-model="form.supportsSms" type="checkbox" id="sms"> 
            <label for="sms" class="inline">Supports SMS</label>
          </div>
          <div class="check-item">
            <input v-model="form.supportsMms" type="checkbox" id="mms"> 
            <label for="mms" class="inline">Supports MMS</label>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" :disabled="saving" @click="saveChanges">
          {{ saving ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { numbersAPI, extensionsAPI, queuesAPI, ringGroupsAPI, ivrAPI, voicemailAPI } from '../../services/api.js'

const router = useRouter()
const route = useRoute()

// State with atomic predictability - explicit initial state
const loading = ref(true)
const saving = ref(false)
const error = ref(null)

const form = ref({
  destinationType: 'extension',
  destinationTarget: '',
  callerIdPrefix: '',
  supportsSms: false,
  supportsMms: false,
})

// Destination options
const extensions = ref([])
const queues = ref([])
const ringGroups = ref([])
const ivrMenus = ref([])
const voicemailBoxes = ref([])

/**
 * Fetches all destination options in parallel for atomic predictability.
 * Early exit on critical failure.
 */
const fetchDestinationOptions = async () => {
  try {
    const [extensionsRes, queuesRes, ringGroupsRes, ivrRes, voicemailRes] = await Promise.all([
      extensionsAPI.list(),
      queuesAPI.list(),
      ringGroupsAPI.list(),
      ivrAPI.listMenus(),
      voicemailAPI.listBoxes(),
    ])

    extensions.value = extensionsRes.data || []
    queues.value = queuesRes.data || []
    ringGroups.value = ringGroupsRes.data || []
    ivrMenus.value = ivrRes.data || []
    voicemailBoxes.value = voicemailRes.data || []
  } catch (err) {
    // Fail loud - log but don't halt since these are supplementary
    console.error('Failed to load destination options:', err)
  }
}

/**
 * Fetches the number being edited using route param ID.
 * Sets error state on failure for Fail Loud compliance.
 */
const fetchNumberDetails = async () => {
  const numberId = route.params.id
  
  if (!numberId) {
    error.value = 'No number ID provided'
    loading.value = false
    return
  }

  try {
    const response = await numbersAPI.get(numberId)
    const number = response.data

    // Parse number data into form - parse at boundary, trust internally
    form.value.destinationType = number.destination_type || 'extension'
    form.value.destinationTarget = String(number.destination_target || '')
    form.value.callerIdPrefix = number.caller_id_prefix || ''
    form.value.supportsSms = Boolean(number.supports_sms)
    form.value.supportsMms = Boolean(number.supports_mms)
  } catch (err) {
    error.value = `Failed to load number details: ${err.message || 'Unknown error'}`
  } finally {
    loading.value = false
  }
}

/**
 * Releases the number by deleting it via numbersAPI.delete.
 * Early exit on confirmation denied - fail loud on API error.
 */
const releaseNumber = async () => {
  const numberId = route.params.id

  if (!numberId) {
    alert('Error: No number ID provided')
    return
  }

  const confirmed = confirm(`Are you sure you want to release number ${numberId}? This action cannot be undone.`)
  if (!confirmed) return

  try {
    await numbersAPI.delete(numberId)
    router.push('/numbers')
  } catch (err) {
    alert(`Failed to release number: ${err.message || 'Unknown error'}`)
  }
}

/**
 * Saves changes using numbersAPI.update with proper error handling.
 * Returns boolean for caller feedback - atomic predictability.
 */
const saveChanges = async () => {
  const numberId = route.params.id
  
  if (!numberId) {
    alert('Error: No number ID provided')
    return
  }

  // Fail loud if no destination selected
  if (!form.value.destinationTarget) {
    alert('Please select a destination target')
    return
  }

  saving.value = true

  try {
    const payload = {
      destination_type: form.value.destinationType,
      destination_target: form.value.destinationTarget,
      caller_id_prefix: form.value.callerIdPrefix,
      supports_sms: form.value.supportsSms,
      supports_mms: form.value.supportsMms,
    }

    await numbersAPI.update(numberId, payload)
    
    router.push('/numbers')
  } catch (err) {
    alert(`Failed to save changes: ${err.message || 'Unknown error'}`)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  // Fetch number details and destination options concurrently
  await Promise.all([
    fetchNumberDetails(),
    fetchDestinationOptions(),
  ])
})
</script>

<style scoped>
/* Reusing Form Styles */
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: var(--spacing-xl);
}
.view-header {
  display: flex;
  justify-content: space-between;
}
.back-link {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  font-size: var(--text-xs);
  text-align: left;
}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.form-container {
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.form-section h3 {
  font-size: var(--text-md);
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 8px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: var(--spacing-md);
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
  font-size: var(--text-sm);
  color: var(--text-primary);
  outline: none;
  background: white;
  transition: border-color var(--transition-fast);
}
.input-field:focus { border-color: var(--primary-color); }

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  width: 100%;
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-danger {
  background: white;
  color: var(--status-bad);
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  cursor: pointer;
}

.checkbox-row {
  flex-direction: row; gap: 24px; margin-top: 8px;
}
.check-item { display: flex; align-items: center; gap: 8px; }
label.inline { text-transform: none; font-size: var(--text-sm); font-weight: 500; cursor: pointer; color: var(--text-primary); }

.loading-state,
.error-state {
  padding: var(--spacing-lg);
  text-align: center;
  color: var(--text-muted);
}
.error-state {
  color: var(--status-bad);
}
</style>
