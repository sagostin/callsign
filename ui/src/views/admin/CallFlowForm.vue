<template>
  <div class="form-container">
    <div class="form-header">
      <div class="header-left">
        <button class="back-link" @click="$router.push('/admin/call-flows')">&larr; Back to Call Flows</button>
        <h2>{{ isNew ? 'New Call Flow' : 'Edit Call Flow' }}</h2>
      </div>
      <button class="btn-secondary" @click="$router.push('/admin/call-flows')">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Call Flow Name <span class="required">*</span></label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Day/Night Mode" />
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Extension <span class="required">*</span></label>
          <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 30" />
        </div>
        <div class="form-group">
          <label>Feature Code</label>
          <input v-model="form.feature_code" type="text" class="input-field" placeholder="e.g. *30" />
        </div>
      </div>

      <div class="form-group">
        <label>Description</label>
        <textarea v-model="form.description" rows="2" class="input-field"></textarea>
      </div>

      <div class="form-group">
        <label class="checkbox-label">
          <input type="checkbox" v-model="form.enabled" />
          Enabled
        </label>
      </div>

      <hr class="divider" />

      <div class="destinations-section">
        <div class="section-header">
          <h4 class="section-title">Toggle States (Minimum 2)</h4>
          <button type="button" class="btn-sm" @click="addDestination">
            <PlusIcon class="btn-icon-sm" /> Add State
          </button>
        </div>

        <div class="destinations-list">
          <div
            class="destination-card"
            v-for="(dest, i) in form.destinations"
            :key="i"
          >
            <div class="dest-header">
              <span class="dest-index">State {{ i + 1 }}</span>
              <button
                v-if="form.destinations.length > 2"
                type="button"
                class="btn-icon remove-btn"
                @click="removeDestination(i)"
                title="Remove state"
              >
                <TrashIcon class="icon-sm text-bad" />
              </button>
            </div>

            <div class="form-row">
              <div class="form-group flex-1">
                <label>Label</label>
                <input v-model="dest.label" class="input-field" :placeholder="i === 0 ? 'Day Mode' : 'Night Mode'" />
              </div>
              <div class="form-group flex-1">
                <label>Sound (optional)</label>
                <select v-model="dest.sound" class="input-field">
                  <option value="">None</option>
                  <option value="day_mode.wav">Day Mode Activated</option>
                  <option value="night_mode.wav">Night Mode Activated</option>
                  <option value="holiday_mode.wav">Holiday Mode Activated</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label>Destination</label>
              <div class="dest-selector">
                <select v-model="dest.dest_type" class="input-field">
                  <option value="ivr">IVR Menu</option>
                  <option value="time_condition">Time Condition</option>
                  <option value="extension">Extension</option>
                  <option value="ring_group">Ring Group</option>
                  <option value="queue">Queue</option>
                  <option value="voicemail">Voicemail</option>
                  <option value="external">External Number</option>
                </select>
                <input v-model="dest.dest_value" class="input-field flex-1" placeholder="Destination value" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" :disabled="saving || !isValid" @click="save">
          {{ saving ? 'Saving...' : (isNew ? 'Create Call Flow' : 'Update Call Flow') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Plus as PlusIcon, Trash2 as TrashIcon } from 'lucide-vue-next'
import { callFlowsAPI } from '../../services/api'

const toast = inject('toast')
const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)

const saving = ref(false)
const loading = ref(false)

const defaultDestinations = () => [
  { label: 'Day Mode', dest_type: 'ivr', dest_value: '', sound: '' },
  { label: 'Night Mode', dest_type: 'voicemail', dest_value: '', sound: '' }
]

const form = ref({
  name: '',
  extension: '',
  feature_code: '',
  description: '',
  enabled: true,
  destinations: defaultDestinations()
})

const isValid = computed(() => {
  return form.value.name && form.value.extension && form.value.destinations.length >= 2
})

const addDestination = () => {
  const num = form.value.destinations.length + 1
  form.value.destinations.push({
    label: `State ${num}`,
    dest_type: 'extension',
    dest_value: '',
    sound: ''
  })
}

const removeDestination = (index) => {
  if (form.value.destinations.length > 2) {
    form.value.destinations.splice(index, 1)
  }
}

const save = async () => {
  if (!form.value.name) {
    toast?.warning('Call flow name is required.')
    return
  }
  if (!form.value.extension) {
    toast?.warning('Extension is required.')
    return
  }
  if (form.value.destinations.length < 2) {
    toast?.warning('At least 2 toggle states are required.')
    return
  }

  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      extension: form.value.extension,
      feature_code: form.value.feature_code,
      description: form.value.description,
      enabled: form.value.enabled,
      destinations: form.value.destinations
    }

    if (isNew.value) {
      await callFlowsAPI.create(payload)
      toast?.success(`Call flow "${form.value.name}" created`)
    } else {
      await callFlowsAPI.update(route.params.id, payload)
      toast?.success(`Call flow "${form.value.name}" updated`)
    }

    router.push('/admin/call-flows')
  } catch (err) {
    console.error('Failed to save call flow:', err)
    toast?.error(err.message || 'Failed to save call flow')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  if (!isNew.value && route.params.id) {
    loading.value = true
    try {
      const res = await callFlowsAPI.get(route.params.id)
      const d = res.data
      form.value = {
        name: d.name || '',
        extension: d.extension || '',
        feature_code: d.feature_code || '',
        description: d.description || '',
        enabled: d.enabled !== false,
        destinations: (d.destinations || []).map(dest => ({
          label: dest.label || '',
          dest_type: dest.dest_type || 'extension',
          dest_value: dest.dest_value || '',
          sound: dest.sound || ''
        }))
      }
      if (form.value.destinations.length < 2) {
        form.value.destinations = defaultDestinations()
      }
    } catch (err) {
      console.error('Failed to load call flow:', err)
      toast?.error(err.message || 'Failed to load call flow')
    } finally {
      loading.value = false
    }
  }
})
</script>

<style scoped>
.form-container { max-width: 720px; margin: 0 auto; padding: 0 20px; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.header-left { display: flex; flex-direction: column; gap: 8px; }
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

.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.required { color: #ef4444; }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--text-sm);
  color: var(--text-primary);
  text-transform: none;
  letter-spacing: normal;
  font-weight: 500;
  cursor: pointer;
}
.checkbox-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.divider { border: none; border-top: 1px dashed var(--border-color); margin: 16px 0; }
.section-title { font-size: 12px; font-weight: 700; color: var(--text-primary); margin: 0 0 12px 0; }

.destinations-section { margin-top: 8px; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }

.destinations-list { display: flex; flex-direction: column; gap: 12px; }
.destination-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 16px;
  background: var(--bg-app);
}
.dest-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.dest-index {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
}

.dest-selector {
  display: flex;
  gap: 8px;
}
.dest-selector select { width: 140px; }
.dest-selector input { flex: 1; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-sm {
  background: white;
  border: 1px solid var(--border-color);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
  display: flex;
}
.remove-btn:hover { color: var(--status-bad); }
.text-bad { color: var(--status-bad); }
.icon-sm { width: 14px; height: 14px; }
.btn-icon-sm { width: 14px; height: 14px; }

.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; }
.flex-1 { flex: 1; }
</style>
