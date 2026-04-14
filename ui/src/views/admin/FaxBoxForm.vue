<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Fax Box' : 'Edit Fax Box' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Sales Fax">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Extension</label>
          <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 50">
        </div>
        <div class="form-group">
          <label>Email Destination</label>
          <input v-model="form.email" type="email" class="input-field" placeholder="fax@company.com">
        </div>
      </div>

       <div class="form-group">
        <label>Associated DID (Number)</label>
        <select v-model="form.did" class="input-field">
            <option value="">Select a Number...</option>
            <option v-for="num in availableDIDs" :key="num.id || num.number" :value="num.number || num.id">
              {{ num.formatted_number || num.number }} - {{ num.name || '' }}
            </option>
        </select>
      </div>

       <div class="form-group">
        <label>Caller ID Name</label>
        <input v-model="form.cid_name" type="text" class="input-field" placeholder="Use Default">
      </div>

       <div class="form-group full-width">
        <label>Access Control</label>
        <div class="radio-group">
          <label class="radio-label">
            <input type="radio" v-model="form.access" value="all">
            All Extensions
          </label>
          <label class="radio-label">
            <input type="radio" v-model="form.access" value="selected">
            Select Extensions
          </label>
        </div>
        
        <div class="extension-selector" v-if="form.access === 'selected'">
           <label class="sub-label">Allowed Extensions</label>
           <select multiple v-model="form.allowed_exts" class="input-field multiple-select">
              <option v-for="ext in availableExtensions" :key="ext.id || ext.extension" :value="ext.extension">
                {{ ext.extension }} - {{ ext.name || 'Extension' }}
              </option>
           </select>
           <p class="text-xs text-muted">Hold Cmd/Ctrl to select multiple.</p>
        </div>
      </div>

      <div class="form-actions">
        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
        <button class="btn-primary" :disabled="isSaving" @click="save">
          {{ isSaving ? 'Saving...' : 'Save Fax Box' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { numbersAPI, extensionsAPI, faxAPI } from '../../services/api'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)

const form = ref({
  name: '',
  extension: '',
  email: '',
  cid_name: '',
  access: 'all',
  allowed_exts: []
})

// Dropdown data loaded from API
const availableDIDs = ref([])
const availableExtensions = ref([])

const loadDropdownData = async () => {
  try {
    const [didRes, extRes] = await Promise.allSettled([
      numbersAPI.list(),
      extensionsAPI.list()
    ])
    if (didRes.status === 'fulfilled') {
      availableDIDs.value = didRes.value.data?.data || didRes.value.data || []
    }
    if (extRes.status === 'fulfilled') {
      availableExtensions.value = extRes.value.data?.data || extRes.value.data || []
    }
  } catch (err) {
    console.error('Failed to load dropdown data:', err)
  }
}

onMounted(() => {
  loadDropdownData()
})

const isSaving = ref(false)
const errorMessage = ref('')

const save = async () => {
  if (isSaving.value) return

  isSaving.value = true
  errorMessage.value = ''

  try {
    const payload = {
      name: form.value.name.trim(),
      extension: form.value.extension.trim(),
      email: form.value.email.trim(),
      cid_name: form.value.cid_name.trim(),
      did: form.value.did || null,
      access: form.value.access,
      allowed_exts: form.value.access === 'selected' ? form.value.allowed_exts : []
    }

    if (isNew.value) {
      await faxAPI.createBox(payload)
    } else {
      await faxAPI.updateBox(route.params.id, payload)
    }

    router.push('/admin/fax')
  } catch (err) {
    errorMessage.value = err.message || 'Failed to save fax box. Please try again.'
  } finally {
    isSaving.value = false
  }
}
</script>

<style scoped>
.form-container { max-width: 600px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; gap: 12px; align-items: center; }
.error-text { color: var(--error-color); font-size: 13px; margin: 0; }

/* Access Control Styles */
.radio-group { display: flex; gap: 16px; margin-bottom: 8px; }
.radio-label { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 500; text-transform: none; color: var(--text-main); cursor: pointer; }
.extension-selector { background: var(--bg-app); padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border-color); }
.multiple-select { height: 100px; width: 100%; margin-top: 4px; }
.text-xs { font-size: 11px; }
.sub-label { font-size: 11px; font-weight: 600; }
</style>
