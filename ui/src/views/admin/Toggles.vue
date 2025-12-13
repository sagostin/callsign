<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Toggles</h2>
      <p class="text-muted text-sm">Toggle between two routing modes by dialing an extension or feature code.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="showModal = true">+ New Toggle</button>
    </div>
  </div>


  <!-- Toggles List -->
  <div class="toggles-grid">
    <div class="toggle-card" v-for="toggle in mappedToggles" :key="toggle.id" :class="{ 'mode-a': toggle.status === 'A', 'mode-b': toggle.status === 'B' }">
      <div class="toggle-header">
        <div class="toggle-icon" :class="toggle.status === 'A' ? 'mode-a' : 'mode-b'">
          <ToggleRightIcon v-if="toggle.status === 'A'" class="icon-md" />
          <ToggleLeftIcon v-else class="icon-md" />
        </div>
        <div class="toggle-info">
          <h4>{{ toggle.name }}</h4>
          <div class="toggle-codes">
            <span class="code-badge ext">{{ toggle.extension }}</span>
          </div>
        </div>
        <div class="toggle-switch" @click="flipToggle(toggle)">
          <div class="switch-track" :class="{ 'active': toggle.currentState > 0 }">
            <span class="switch-label">{{ toggle.currentLabel }}</span>
            <span class="switch-state">{{ toggle.currentState + 1 }}/{{ toggle.stateCount }}</span>
          </div>
        </div>
      </div>


      <p class="toggle-desc">{{ toggle.description }}</p>

      <div class="toggle-modes">
        <div 
          class="mode-row" 
          v-for="(dest, i) in toggle.destinations" 
          :key="i"
          :class="{ active: toggle.currentState === i }"
        >
          <div class="mode-indicator">{{ String.fromCharCode(65 + i) }}</div>
          <div class="mode-content">
            <span class="mode-label">{{ dest.label || `State ${i + 1}` }}</span>
            <span class="mode-dest">→ {{ dest.dest_type }}: {{ dest.dest_value || '-' }}</span>
          </div>
          <VolumeIcon v-if="dest.sound" class="icon-sm sound-icon" />
        </div>
      </div>


      <div class="toggle-footer">
        <span class="toggle-meta">Last changed: {{ toggle.lastChanged }}</span>
        <div class="toggle-actions">
          <button class="btn-link" @click="editToggle(toggle)">Edit</button>
          <button class="btn-link text-bad" @click="deleteToggle(toggle)">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Add/Edit Modal -->
  <div class="modal-overlay" v-if="showModal" @click.self="closeModal">
    <div class="modal-card">
      <div class="modal-header">
        <h3>{{ editingToggle ? 'Edit Toggle' : 'New Toggle' }}</h3>
        <button class="btn-icon" @click="closeModal"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Toggle Name <span class="required">*</span></label>
          <input v-model="form.name" class="input-field" placeholder="Day/Night Mode">
        </div>

        <div class="form-group">
          <label>Dial Code <span class="required">*</span></label>
          <input v-model="form.extension" class="input-field" placeholder="30 or *30">
          <span class="help-text">Extension or feature code to dial to toggle (e.g., 30 or *30)</span>
        </div>


        <div class="form-group">
          <label>PIN (Optional)</label>
          <input v-model="form.pin" class="input-field" type="password" placeholder="Secure access">
        </div>

        <div class="form-group">
          <label>Description</label>
          <input v-model="form.description" class="input-field" placeholder="Toggle between day and night routing">
        </div>

        <hr class="divider">

        <!-- Dynamic Destinations -->
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
              :class="{ 'first': i === 0 }"
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
                  <input v-model="dest.label" class="input-field" :placeholder="i === 0 ? 'Day Mode' : 'Night Mode'">
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
                  <input v-model="dest.dest_value" class="input-field flex-1" :placeholder="i === 0 ? 'Main Menu' : 'After Hours VM'">
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>


      <div class="modal-actions">
        <button class="btn-secondary" @click="closeModal">Cancel</button>
        <button class="btn-primary" @click="saveToggle" :disabled="!form.name || !form.extension">Save Toggle</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { 
  ToggleLeft as ToggleLeftIcon, ToggleRight as ToggleRightIcon, X as XIcon,
  PhoneCall as PhoneCallIcon, PhoneForwarded as PhoneForwardedIcon, Volume2 as VolumeIcon, 
  RefreshCw, Plus as PlusIcon, Trash2 as TrashIcon
} from 'lucide-vue-next'
import { togglesAPI } from '../../services/api'

const showModal = ref(false)
const editingToggle = ref(null)
const loading = ref(false)
const error = ref(null)
const toggles = ref([])

// Default destinations (minimum 2)
const defaultDestinations = () => [
  { label: 'Day Mode', dest_type: 'ivr', dest_value: '', sound: '' },
  { label: 'Night Mode', dest_type: 'voicemail', dest_value: '', sound: '' }
]

const form = ref({
  name: '', extension: '', feature_code: '', description: '',
  destinations: defaultDestinations(),
  enabled: true
})

// Map backend CallFlow model to UI format
const mappedToggles = computed(() => toggles.value.map(t => {
  const dests = t.destinations || []
  const currentDest = dests[t.current_state] || dests[0] || {}
  return {
    ...t,
    status: t.current_state === 0 ? 'A' : String.fromCharCode(65 + t.current_state), // A, B, C...
    currentState: t.current_state || 0,
    featureCode: t.feature_code,
    currentLabel: currentDest.label || `State ${(t.current_state || 0) + 1}`,
    stateCount: dests.length,
    destinations: dests,
    lastChanged: t.updated_at ? new Date(t.updated_at).toLocaleDateString() : 'Unknown'
  }
}))

onMounted(() => loadToggles())

async function loadToggles() {
  loading.value = true
  error.value = null
  try {
    const response = await togglesAPI.list()
    toggles.value = response.data?.data || []
  } catch (e) {
    error.value = 'Failed to load toggles'
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function flipToggle(toggle) {
  try {
    await togglesAPI.toggle(toggle.id)
    await loadToggles()
  } catch (e) {
    error.value = 'Failed to toggle status'
    console.error(e)
  }
}

const closeModal = () => {
  showModal.value = false
  editingToggle.value = null
  resetForm()
}

const resetForm = () => {
  form.value = {
    name: '', extension: '', feature_code: '', description: '',
    destinations: defaultDestinations(),
    enabled: true
  }
}

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

const editToggle = (toggle) => {
  editingToggle.value = toggle
  // Clone destinations to avoid reactivity issues
  const dests = (toggle.destinations || []).map(d => ({ ...d }))
  form.value = {
    name: toggle.name,
    extension: toggle.extension,
    feature_code: toggle.feature_code || '',
    description: toggle.description || '',
    destinations: dests.length >= 2 ? dests : defaultDestinations(),
    enabled: toggle.enabled !== false
  }
  showModal.value = true
}

async function saveToggle() {
  error.value = null
  try {
    const payload = {
      name: form.value.name,
      extension: form.value.extension,
      feature_code: form.value.feature_code,
      description: form.value.description,
      destinations: form.value.destinations,
      enabled: form.value.enabled
    }

    if (editingToggle.value) {
      await togglesAPI.update(editingToggle.value.id, payload)
    } else {
      await togglesAPI.create(payload)
    }
    await loadToggles()
    closeModal()
  } catch (e) {
    error.value = e.response?.data?.error || 'Failed to save toggle'
    console.error(e)
  }
}

async function deleteToggle(toggle) {
  if (!confirm(`Delete "${toggle.name}"?`)) return
  error.value = null
  try {
    await togglesAPI.delete(toggle.id)
    await loadToggles()
  } catch (e) {
    error.value = 'Failed to delete toggle'
    console.error(e)
  }
}
</script>



<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-actions { display: flex; gap: 8px; }

.stats-row { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); max-width: 400px; }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.mode-a { background: #dbeafe; color: #2563eb; }
.stat-icon.mode-b { background: #fef3c7; color: #b45309; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 12px; color: var(--text-muted); }

.toggles-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 16px; }

.toggle-card { background: white; border: 2px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; transition: all 0.2s; }
.toggle-card.mode-a { border-color: #3b82f6; background: linear-gradient(135deg, #eff6ff 0%, white 50%); }
.toggle-card.mode-b { border-color: #f59e0b; background: linear-gradient(135deg, #fffbeb 0%, white 50%); }

.toggle-header { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.toggle-icon { width: 40px; height: 40px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.toggle-icon.mode-a { background: #dbeafe; color: #2563eb; }
.toggle-icon.mode-b { background: #fef3c7; color: #b45309; }
.icon-md { width: 20px; height: 20px; }
.toggle-info { flex: 1; }
.toggle-info h4 { font-size: 14px; font-weight: 600; margin: 0 0 4px 0; }
.toggle-codes { display: flex; gap: 6px; }
.code-badge { font-size: 10px; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-weight: 600; }
.code-badge.ext { background: #f0fdf4; color: #16a34a; }
.code-badge.feature { background: #faf5ff; color: #9333ea; }

.toggle-switch { cursor: pointer; }
.switch-track { min-width: 80px; height: 28px; background: linear-gradient(90deg, #3b82f6, #60a5fa); border-radius: 14px; position: relative; display: flex; align-items: center; justify-content: space-between; padding: 0 10px; transition: background 0.3s; }
.switch-track.active { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.switch-label { font-size: 10px; font-weight: 700; color: white; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 60px; }
.switch-state { font-size: 9px; font-weight: 600; color: rgba(255,255,255,0.8); background: rgba(0,0,0,0.15); padding: 2px 5px; border-radius: 8px; }


.toggle-desc { font-size: 12px; color: var(--text-muted); margin: 0 0 12px 0; }

.toggle-modes { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.mode-row { display: flex; align-items: center; gap: 10px; padding: 8px 10px; background: var(--bg-app); border-radius: var(--radius-sm); border: 1px solid transparent; transition: all 0.2s; }
.mode-row.active { background: white; border-color: var(--border-color); box-shadow: var(--shadow-sm); }
.mode-indicator { width: 22px; height: 22px; border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; background: #e2e8f0; color: #64748b; }
.mode-row.active .mode-indicator { background: #3b82f6; color: white; }
.mode-content { flex: 1; display: flex; flex-direction: column; }
.mode-label { font-size: 12px; font-weight: 600; color: var(--text-primary); }
.mode-dest { font-size: 11px; color: var(--text-muted); }
.sound-icon { color: #8b5cf6; }
.icon-sm { width: 14px; height: 14px; }

.toggle-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 12px; border-top: 1px solid var(--border-color); }
.toggle-meta { font-size: 10px; color: var(--text-muted); }
.toggle-actions { display: flex; gap: 8px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 540px; max-height: 90vh; overflow-y: auto; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); position: sticky; top: 0; background: white; }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); position: sticky; bottom: 0; background: white; }

.form-row { display: flex; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.flex-1 { flex: 1; }
label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.required { color: #ef4444; }
.input-field { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.help-text { font-size: 10px; color: var(--text-muted); }

.feature-code-input { display: flex; }
.feature-code-input .prefix { background: #f1f5f9; border: 1px solid var(--border-color); border-right: none; padding: 8px 12px; border-radius: var(--radius-sm) 0 0 var(--radius-sm); font-weight: 600; color: var(--text-muted); }
.feature-code-input .input-field { border-radius: 0 var(--radius-sm) var(--radius-sm) 0; width: 80px; }

.dest-selector { display: flex; gap: 8px; }
.dest-selector select { width: 140px; }
.dest-selector input { flex: 1; }

.divider { border: none; border-top: 1px dashed var(--border-color); margin: 16px 0; }
.section-title { font-size: 12px; font-weight: 700; color: var(--text-primary); margin: 0 0 12px 0; }

.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; font-size: var(--text-sm); cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 11px; cursor: pointer; font-weight: 500; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; display: flex; }
.text-bad { color: var(--status-bad); }
</style>
