<template>
  <div class="view-header">
     <div class="header-content">
       <button class="back-link" @click="$router.push('/system/tenants')">← Back to Tenants</button>
       <h2>{{ isNew ? 'Create New Tenant' : 'Edit Tenant' }}</h2>
     </div>
  </div>

  <div class="settings-container">
    <div class="settings-content">
       <h3 class="card-title">Basic Information</h3>
       <div class="form-group">
          <label>Tenant Name</label>
          <input v-model="form.name" class="input-field" placeholder="Company Name">
       </div>
       
       <div class="form-row">
          <div class="form-group">
             <label>SIP Domain</label>
             <input v-model="form.domain" class="input-field" placeholder="tenant.callsign.io">
          </div>
          <div class="form-group">
             <label>Admin Email</label>
             <input v-model="form.admin_email" class="input-field" placeholder="admin@company.com">
          </div>
       </div>

       <div class="divider"></div>
       <h3 class="card-title">Service Configuration</h3>
       
       <div class="form-row">
          <div class="form-group">
             <label>Service Profile</label>
             <select v-model="form.profile_id" class="input-field">
                <option :value="null">-- Select Profile --</option>
                <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }}</option>
             </select>
          </div>
          <div class="form-group">
             <label>Status</label>
             <select v-model="form.enabled" class="input-field">
                <option :value="true">Active</option>
                <option :value="false">Suspended</option>
             </select>
          </div>
       </div>

       <div class="divider"></div>
       <h3 class="card-title">General Settings</h3>
       
       <div class="form-row">
          <div class="form-group">
             <label>Timezone</label>
             <select v-model="form.settings.timezone" class="input-field">
                <option value="America/New_York">Eastern Time (US & Canada)</option>
                <option value="America/Chicago">Central Time (US & Canada)</option>
                <option value="America/Denver">Mountain Time (US & Canada)</option>
                <option value="America/Los_Angeles">Pacific Time (US & Canada)</option>
                <option value="UTC">UTC</option>
             </select>
          </div>
          <div class="form-group">
             <label>Operator Extension</label>
             <input v-model="form.settings.operator_extension" class="input-field" placeholder="0">
          </div>
       </div>

       <div class="form-row">
          <div class="form-group">
             <label>Fallback Caller ID Name</label>
             <input v-model="form.settings.caller_id_name" class="input-field" placeholder="Company Name">
          </div>
          <div class="form-group">
             <label>Fallback Caller ID Number</label>
             <input v-model="form.settings.caller_id_number" class="input-field" placeholder="+1234567890">
          </div>
       </div>

       <div class="divider"></div>
       <h3 class="card-title">Limit Overrides (Optional)</h3>
       
       <div class="form-row">
          <div class="form-group">
             <label>VM Message Limit</label>
             <input v-model.number="form.settings.vm_limit" type="number" class="input-field" placeholder="e.g. 100">
             <span class="text-xs text-muted">Leave empty to use profile default</span>
          </div>
          <div class="form-group">
             <label>Fax Retention (Days)</label>
             <input v-model.number="form.settings.fax_retention_days" type="number" class="input-field" placeholder="e.g. 30">
          </div>
       </div>
    </div>
  </div>

  <div class="form-actions">
     <button class="btn-secondary" @click="$router.push('/system/tenants')">Cancel</button>
     <button class="btn-primary" @click="save" :disabled="saving">{{ saving ? 'Saving...' : 'Save Tenant' }}</button>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { systemAPI } from '../../services/api'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)
const tenantId = computed(() => route.params.id)
const loading = ref(false)
const saving = ref(false)
const profiles = ref([])

const form = ref({
   name: '',
   domain: '',
   admin_email: '',
   profile_id: null,
   enabled: true,
   settings: {
     timezone: 'America/Los_Angeles',
     operator_extension: '',
     caller_id_name: '',
     caller_id_number: '',
     vm_limit: null,
     fax_retention_days: null
   }
})

const loadProfiles = async () => {
  try {
    const response = await systemAPI.listProfiles()
    profiles.value = response.data.data || response.data || []
  } catch (e) {
    console.error('Failed to load profiles:', e)
  }
}

const loadTenant = async () => {
  if (isNew.value) return
  loading.value = true
  try {
    const response = await systemAPI.getTenant(tenantId.value)
    const t = response.data.data || response.data
    
    let parsedSettings = {}
    try {
        parsedSettings = typeof t.settings === 'string' ? JSON.parse(t.settings) : (t.settings || {})
    } catch {
        parsedSettings = {}
    }

    form.value = {
      name: t.name || '',
      domain: t.domain || '',
      admin_email: t.admin_email || '',
      profile_id: t.profile_id || null,
      enabled: t.enabled !== false,
      settings: {
        timezone: parsedSettings.timezone || 'America/Los_Angeles',
        operator_extension: parsedSettings.operator_extension || '',
        caller_id_name: parsedSettings.caller_id_name || '',
        caller_id_number: parsedSettings.caller_id_number || '',
        vm_limit: parsedSettings.vm_limit || null,
        fax_retention_days: parsedSettings.fax_retention_days || null
      }
    }
  } catch (e) {
    console.error('Failed to load tenant:', e)
    alert('Failed to load tenant')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadProfiles()
  await loadTenant()
})

const save = async () => {
  saving.value = true
  try {
    // Prepare payload
    const payload = {
        ...form.value,
        settings: JSON.stringify(form.value.settings)
    }

    if (isNew.value) {
      await systemAPI.createTenant(payload)
    } else {
      await systemAPI.updateTenant(tenantId.value, payload)
    }
    router.push('/system/tenants')
  } catch (e) {
    console.error('Failed to save tenant:', e)
    alert('Failed to save tenant: ' + (e.response?.data?.error || e.message))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.view-header { margin-bottom: 24px; }
.back-link { background: none; border: none; color: var(--text-muted); padding: 0; font-size: 11px; cursor: pointer; }
.back-link:hover { color: var(--primary-color); text-decoration: underline; }

.settings-container { display: flex; gap: var(--spacing-xl); align-items: flex-start; }
.settings-content { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 24px; max-width: 600px; }

.card-title { font-size: 14px; font-weight: 700; color: var(--text-primary); margin-bottom: 16px; border-bottom: 1px solid var(--border-color); padding-bottom: 8px; }

.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }

.divider { height: 1px; background: var(--border-color); margin: 24px 0; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }

.form-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px; }
</style>
