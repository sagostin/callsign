<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/numbers')">← Back to Numbers</button>
      <h2>Add New Number</h2>
    </div>
  </div>

  <div class="form-container">
    <div class="form-section">
      <h3>1. Number Details</h3>
      <div class="form-grid">
        <div class="form-group">
          <label>Phone Number</label>
          <input 
            type="text" 
            v-model="form.destinationNumber" 
            class="input-field" 
            placeholder="+14155550101"
          >
        </div>

        <div class="form-group">
          <label>Description (optional)</label>
          <input 
            type="text" 
            v-model="form.description" 
            class="input-field" 
            placeholder="Main line, Support, etc."
          >
        </div>
      </div>
    </div>

    <div class="form-section" v-if="form.destinationNumber">
      <h3>2. Initial Routing</h3>
      <div class="form-group">
        <label>Destination Type</label>
        <select v-model="form.destinationType" class="input-field">
          <option value="extension">Extension</option>
          <option value="ivr">IVR Menu</option>
          <option value="queue">Queue</option>
          <option value="voicemail">Voicemail Box</option>
        </select>
      </div>

      <div class="form-group" v-if="destinationOptions.length">
        <label>Destination Target</label>
        <select v-model="form.destinationTarget" class="input-field">
          <option value="" disabled>Select Target</option>
          <option 
            v-for="opt in destinationOptions" 
            :key="opt.value" 
            :value="opt.value"
          >
            {{ opt.label }}
          </option>
        </select>
      </div>
    </div>

    <div class="form-actions">
      <button 
        class="btn-primary large" 
        :disabled="!isValid || isSaving" 
        @click="saveNumber"
      >
        <span v-if="isSaving">Creating...</span>
        <span v-else>Create Number</span>
      </button>
    </div>

    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { numbersAPI, extensionsAPI, queuesAPI, ivrAPI, voicemailAPI } from '@/services/api'

const router = useRouter()

const form = ref({
  destinationNumber: '',
  description: '',
  destinationType: 'extension',
  destinationTarget: ''
})

const isSaving = ref(false)
const errorMessage = ref('')
const availableExtensions = ref([])
const availableIVRs = ref([])
const availableQueues = ref([])
const availableVoicemails = ref([])

const isValid = computed(() => {
  return form.value.destinationNumber?.trim() && form.value.destinationTarget
})

const destinationOptions = computed(() => {
  switch (form.value.destinationType) {
    case 'extension':
      return availableExtensions.value
    case 'ivr':
      return availableIVRs.value
    case 'queue':
      return availableQueues.value
    case 'voicemail':
      return availableVoicemails.value
    default:
      return []
  }
})

const buildDestinationAction = () => {
  const target = form.value.destinationTarget
  switch (form.value.destinationType) {
    case 'extension':
      return `transfer ${target} XML default`
    case 'ivr':
      return `menu ${target} default`
    case 'queue':
      return `queue ${target} default`
    case 'voicemail':
      return `voicemail ${target}@default`
    default:
      return ''
  }
}

const loadDestinationOptions = async () => {
  try {
    const [extRes, ivrRes, queueRes, vmRes] = await Promise.all([
      extensionsAPI.list(),
      ivrAPI.listMenus(),
      queuesAPI.list(),
      voicemailAPI.listBoxes()
    ])

    availableExtensions.value = (extRes.data || []).map(ext => ({
      value: ext.extension,
      label: `${ext.extension} - ${ext.effective_caller_id_name || ext.display_name || 'Extension'}`
    }))

    availableIVRs.value = (ivrRes.data || []).map(ivr => ({
      value: String(ivr.id),
      label: ivr.name || `IVR ${ivr.id}`
    }))

    availableQueues.value = (queueRes.data || []).map(q => ({
      value: String(q.id),
      label: q.name || `Queue ${q.id}`
    }))

    availableVoicemails.value = (vmRes.data || []).map(vm => ({
      value: vm.extension,
      label: `${vm.extension} - Voicemail`
    }))
  } catch (err) {
    console.error('Failed to load destination options:', err)
  }
}

// Reset target when type changes
watch(() => form.value.destinationType, () => {
  form.value.destinationTarget = ''
})

const saveNumber = async () => {
  if (!isValid.value || isSaving.value) return

  isSaving.value = true
  errorMessage.value = ''

  try {
    const payload = {
      destination_number: form.value.destinationNumber.trim(),
      description: form.value.description.trim(),
      destination_type: form.value.destinationType,
      destination_action: buildDestinationAction(),
      enabled: true
    }

    await numbersAPI.create(payload)
    router.push('/numbers')
  } catch (err) {
    errorMessage.value = err.message || 'Failed to create number. Please try again.'
  } finally {
    isSaving.value = false
  }
}

// Load options on mount
loadDestinationOptions()
</script>

<style scoped>
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: var(--spacing-xl);
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

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-md);
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
  letter-spacing: 0.05em;
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

.input-field:focus {
  border-color: var(--primary-color);
}

.input-group {
  display: flex;
  gap: 8px;
}
.input-group .input-field { flex: 1; }

.number-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--spacing-sm);
}

.number-card {
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  text-align: center;
  font-size: var(--text-sm);
  cursor: pointer;
  background: white;
  transition: all var(--transition-fast);
}

.number-card:hover {
  border-color: var(--primary-color);
}

.number-card.selected {
  background-color: var(--primary-light);
  border-color: var(--primary-color);
  color: var(--primary-color);
  font-weight: 600;
}

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
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 0 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
}

.error-message {
  color: var(--error-color, #dc3545);
  padding: 12px;
  background: var(--error-bg, #f8d7da);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  margin-top: var(--spacing-md);
}
</style>
