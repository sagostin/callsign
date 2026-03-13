<template>
  <div class="view-header">
     <div class="header-content">
       <button class="back-link" @click="$router.push('/system/tenants')">← Back to Tenants</button>
       <h2>{{ isNew ? 'Create New Tenant' : tenantName }}</h2>
     </div>
     <button class="btn-primary" @click="save" :disabled="saving">{{ saving ? 'Saving...' : 'Save Changes' }}</button>
  </div>

  <div class="settings-layout">
    <!-- Sidebar Navigation -->
    <div class="settings-nav">
      <div class="nav-group">
        <span class="nav-group-title">General</span>
        <div class="nav-item" :class="{ active: activeSection === 'basics' }" @click="activeSection = 'basics'">
          <BuildingIcon class="nav-icon" />
          <span>Basics</span>
        </div>
        <div class="nav-item" :class="{ active: activeSection === 'domains' }" @click="activeSection = 'domains'">
          <GlobeIcon class="nav-icon" />
          <span>Domains & URLs</span>
        </div>
        <div class="nav-item" :class="{ active: activeSection === 'settings' }" @click="activeSection = 'settings'">
          <SettingsIcon class="nav-icon" />
          <span>Settings</span>
        </div>
      </div>

      <div class="nav-group">
        <span class="nav-group-title">Features</span>
        <div class="nav-item" :class="{ active: activeSection === 'features' }" @click="activeSection = 'features'">
          <PackageIcon class="nav-icon" />
          <span>Enabled Features</span>
        </div>
        <div class="nav-item" :class="{ active: activeSection === 'limits' }" @click="activeSection = 'limits'">
          <GaugeIcon class="nav-icon" />
          <span>Limits & Quotas</span>
        </div>
      </div>

      <div class="nav-group">
        <span class="nav-group-title">Branding</span>
        <div class="nav-item" :class="{ active: activeSection === 'branding' }" @click="activeSection = 'branding'">
          <PaletteIcon class="nav-icon" />
          <span>White Label</span>
        </div>
      </div>

      <div class="nav-group" v-if="!isNew">
        <span class="nav-group-title">Admin</span>
        <div class="nav-item" :class="{ active: activeSection === 'users' }" @click="activeSection = 'users'">
          <UsersIcon class="nav-icon" />
          <span>Users</span>
        </div>
      </div>

      <div class="nav-group" v-if="!isNew">
        <span class="nav-group-title">Danger</span>
        <div class="nav-item nav-item-danger" :class="{ active: activeSection === 'danger' }" @click="activeSection = 'danger'">
          <Trash2Icon class="nav-icon" />
          <span>Delete Tenant</span>
        </div>
      </div>
    </div>

    <!-- Content Panel -->
    <div class="settings-content">
      <!-- BASICS -->
      <div v-if="activeSection === 'basics'" class="settings-panel">
        <div class="panel-header">
          <h3>Basic Information</h3>
        </div>
        
        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Tenant Name</h4>
              <p>Company or organization name.</p>
            </div>
            <input v-model="form.name" class="input-field" placeholder="Company Name">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Admin Email</h4>
              <p>Primary contact email.</p>
            </div>
            <input v-model="form.admin_email" class="input-field" placeholder="admin@company.com">
          </div>
        </div>

        <div class="divider"></div>

        <div class="panel-header">
          <h3>Service Configuration</h3>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Service Profile</h4>
              <p>Defines features and limits for this tenant.</p>
            </div>
            <select v-model="form.profile_id" class="input-field">
               <option :value="null">-- Select Profile --</option>
               <option v-for="p in profiles" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Tenant Status</h4>
              <p>Suspended tenants cannot make/receive calls.</p>
            </div>
            <select v-model="form.enabled" class="input-field" style="width: 150px">
               <option :value="true">Active</option>
               <option :value="false">Suspended</option>
            </select>
          </div>
        </div>
      </div>

      <!-- DOMAINS & URLs -->
      <div v-if="activeSection === 'domains'" class="settings-panel">
        <div class="panel-header">
          <h3>Domains & URLs</h3>
        </div>
        <p class="panel-desc">Configure SIP and web portal access domains.</p>
        
        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>SIP Domain</h4>
              <p>Used for device registration and call routing.</p>
            </div>
            <input v-model="form.domain" class="input-field" placeholder="sip.tenant.com">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Portal Domain</h4>
              <p>Custom domain for user portal access.</p>
            </div>
            <input v-model="form.settings.portal_domain" class="input-field" placeholder="portal.tenant.com">
          </div>
        </div>

        <div class="info-card">
          <div class="info-label">Internal URLs</div>
          <div class="info-grid">
            <span class="info-key">SIP Registrar:</span>
            <code>{{ form.domain || 'sip.tenant.com' }}:5060</code>
            <span class="info-key">Web Portal:</span>
            <code>https://{{ form.settings.portal_domain || form.domain || 'tenant.callsign.io' }}</code>
          </div>
        </div>
      </div>

      <!-- SETTINGS -->
      <div v-if="activeSection === 'settings'" class="settings-panel">
        <div class="panel-header">
          <h3>General Settings</h3>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Timezone</h4>
              <p>Used for schedules and call logs.</p>
            </div>
            <select v-model="form.settings.timezone" class="input-field">
               <option value="America/New_York">Eastern Time (US & Canada)</option>
               <option value="America/Chicago">Central Time (US & Canada)</option>
               <option value="America/Denver">Mountain Time (US & Canada)</option>
               <option value="America/Los_Angeles">Pacific Time (US & Canada)</option>
               <option value="UTC">UTC</option>
            </select>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Operator Extension</h4>
              <p>Extension for operator/receptionist dial 0.</p>
            </div>
            <input v-model="form.settings.operator_extension" class="input-field code" style="width: 80px" placeholder="0">
          </div>
        </div>

        <div class="divider"></div>

        <div class="panel-header">
          <h3>Emergency & E911</h3>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Fallback Caller ID Name</h4>
              <p>Used when E911 location cannot be determined.</p>
            </div>
            <input v-model="form.settings.caller_id_name" class="input-field" placeholder="Company Name">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Fallback Caller ID Number</h4>
              <p>E.164 format phone number.</p>
            </div>
            <input v-model="form.settings.caller_id_number" class="input-field" placeholder="+14155559111">
          </div>
        </div>
      </div>

      <!-- FEATURES -->
      <div v-if="activeSection === 'features'" class="settings-panel">
        <div class="panel-header">
          <h3>Enabled Features</h3>
        </div>
        <p class="panel-desc">Toggle features for this tenant. Some features may be restricted by the service profile.</p>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Call Recording</h4>
              <p>Enable call recording for this tenant.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="form.settings.recording_enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Fax Support</h4>
              <p>Enable fax to email functionality.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="form.settings.fax_enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Hospitality Mode</h4>
              <p>Enable hotel/property management features.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="form.settings.hospitality_enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>SMS/MMS Messaging</h4>
              <p>Enable text messaging features.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="form.settings.messaging_enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>White Label / Custom Branding</h4>
              <p>Allow tenant to customize portal appearance.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="form.settings.whitelabel_enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>
      </div>

      <!-- LIMITS -->
      <div v-if="activeSection === 'limits'" class="settings-panel">
        <div class="panel-header">
          <h3>Limit Overrides</h3>
        </div>
        <p class="panel-desc">Override default limits from the service profile. Leave empty to use profile defaults.</p>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Max Extensions</h4>
              <p>Maximum number of extensions allowed.</p>
            </div>
            <input v-model.number="form.settings.max_extensions" type="number" class="input-field" style="width: 100px" placeholder="From profile">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Max Concurrent Channels</h4>
              <p>Maximum simultaneous calls.</p>
            </div>
            <input v-model.number="form.settings.max_channels" type="number" class="input-field" style="width: 100px" placeholder="From profile">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Recording Storage (GB)</h4>
              <p>Maximum storage for call recordings.</p>
            </div>
            <input v-model.number="form.settings.recording_storage_gb" type="number" class="input-field" style="width: 100px" placeholder="From profile">
          </div>
        </div>

        <div class="divider"></div>

        <div class="panel-header">
          <h3>User Limits</h3>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Voicemail Message Limit</h4>
              <p>Max messages per mailbox.</p>
            </div>
            <input v-model.number="form.settings.vm_limit" type="number" class="input-field" style="width: 100px" placeholder="100">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Fax Retention (Days)</h4>
              <p>How long to keep fax documents.</p>
            </div>
            <input v-model.number="form.settings.fax_retention_days" type="number" class="input-field" style="width: 100px" placeholder="30">
          </div>
        </div>
      </div>

      <!-- BRANDING -->
      <div v-if="activeSection === 'branding'" class="settings-panel">
        <div class="panel-header">
          <h3>White Label Settings</h3>
        </div>
        <p class="panel-desc">Customize the appearance of the user portal for this tenant.</p>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Brand Name</h4>
              <p>Displayed in portal header and emails.</p>
            </div>
            <input v-model="form.settings.brand_name" class="input-field" placeholder="Company Name">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Logo URL</h4>
              <p>Company logo for portal and emails.</p>
            </div>
            <input v-model="form.settings.logo_url" class="input-field" placeholder="https://...">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Primary Color</h4>
              <p>Accent color for buttons and links.</p>
            </div>
            <div class="color-picker">
              <input type="color" v-model="form.settings.primary_color">
              <input type="text" class="input-field code" v-model="form.settings.primary_color" style="width: 90px">
            </div>
          </div>
        </div>
      </div>

      <!-- USERS -->
      <div v-if="activeSection === 'users'" class="settings-panel">
        <div class="panel-header">
          <h3>Tenant Users</h3>
          <button class="btn-secondary small" @click="impersonateTenant">Impersonate Tenant</button>
        </div>
        <p class="panel-desc">Users with admin access to this tenant.</p>

        <div class="info-card">
          <p>Tenant user management is available when impersonating this tenant. Click "Impersonate Tenant" to access the admin dashboard.</p>
        </div>
      </div>

      <!-- DANGER ZONE -->
      <div v-if="activeSection === 'danger'" class="settings-panel danger-panel">
        <div class="panel-header">
          <h3 class="danger-title">⚠️ Danger Zone</h3>
        </div>
        <p class="panel-desc">Irreversible and destructive actions for this tenant.</p>

        <div class="danger-card">
          <div class="danger-card-content">
            <div class="danger-info">
              <h4>Delete this Tenant</h4>
              <p>Permanently delete <strong>{{ form.name || 'this tenant' }}</strong> and <strong>all associated data</strong> including extensions, users, devices, call recordings, voicemail, CDR records, and all configuration. This action <strong>cannot be undone</strong>.</p>
            </div>
            <button class="btn-danger" @click="startDeletion" :disabled="deletionLoading">
              {{ deletionLoading ? 'Loading...' : 'Delete Tenant' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Deletion Confirmation Modal -->
      <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
        <div class="modal-panel danger-modal">
          <div class="modal-header danger-header">
            <h3>⚠️ Confirm Tenant Deletion</h3>
          </div>
          <div class="modal-body">
            <p class="modal-warning">You are about to <strong>permanently delete</strong> the tenant <strong>"{{ form.name }}"</strong>. This will destroy:</p>

            <div class="resource-summary" v-if="deletionPreview">
              <div v-for="item in resourceList" :key="item.key" class="resource-row" v-show="item.count > 0">
                <span class="resource-label">{{ item.label }}</span>
                <span class="resource-count">{{ item.count }}</span>
              </div>
              <div v-if="totalResources === 0" class="resource-row">
                <span class="resource-label" style="color: var(--text-muted)">No associated resources found</span>
              </div>
            </div>

            <div class="confirm-input-section">
              <p>To confirm, type <strong>{{ form.name }}</strong> below:</p>
              <input
                v-model="deleteConfirmText"
                class="input-field confirm-input"
                :placeholder="form.name"
                autocomplete="off"
              >
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-secondary" @click="showDeleteModal = false">Cancel</button>
            <button
              class="btn-danger"
              @click="confirmDeletion"
              :disabled="deleteConfirmText !== form.name || deleting"
            >
              {{ deleting ? 'Deleting...' : 'Permanently Delete Tenant' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Building as BuildingIcon, Globe as GlobeIcon, Settings as SettingsIcon,
  Package as PackageIcon, Gauge as GaugeIcon, Palette as PaletteIcon,
  Users as UsersIcon, Trash2 as Trash2Icon
} from 'lucide-vue-next'
import { systemAPI } from '../../services/api'
import { useAuth } from '../../services/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const toast = inject('toast')

const isNew = computed(() => !route.params.id)
const tenantId = computed(() => route.params.id)
const tenantName = computed(() => form.value.name || 'Edit Tenant')
const loading = ref(false)
const saving = ref(false)
const profiles = ref([])
const showDeleteModal = ref(false)
const deleteConfirmText = ref('')
const deletionPreview = ref(null)
const deletionLoading = ref(false)
const deleting = ref(false)
const activeSection = ref('basics')

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
     portal_domain: '',
     // Features
     recording_enabled: false,
     fax_enabled: false,
     hospitality_enabled: false,
     messaging_enabled: false,
     whitelabel_enabled: false,
     // Limits
     max_extensions: null,
     max_channels: null,
     recording_storage_gb: null,
     vm_limit: null,
     fax_retention_days: null,
     // Branding
     brand_name: '',
     logo_url: '',
     primary_color: '#6366f1'
   }
})

const loadProfiles = async () => {
  try {
    const response = await systemAPI.listProfiles()
    profiles.value = response.data || []
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
        portal_domain: parsedSettings.portal_domain || '',
        recording_enabled: parsedSettings.recording_enabled || false,
        fax_enabled: parsedSettings.fax_enabled || false,
        hospitality_enabled: parsedSettings.hospitality_enabled || false,
        messaging_enabled: parsedSettings.messaging_enabled || false,
        whitelabel_enabled: parsedSettings.whitelabel_enabled || false,
        max_extensions: parsedSettings.max_extensions || null,
        max_channels: parsedSettings.max_channels || null,
        recording_storage_gb: parsedSettings.recording_storage_gb || null,
        vm_limit: parsedSettings.vm_limit || null,
        fax_retention_days: parsedSettings.fax_retention_days || null,
        brand_name: parsedSettings.brand_name || '',
        logo_url: parsedSettings.logo_url || '',
        primary_color: parsedSettings.primary_color || '#6366f1'
      }
    }
  } catch (e) {
    console.error('Failed to load tenant:', e)
    toast?.error('Failed to load tenant')
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
    // Build clean payload, removing null values that backend might reject
    const payload = {
      name: form.value.name,
      domain: form.value.domain,
      enabled: form.value.enabled,
      settings: JSON.stringify(form.value.settings)
    }
    
    // Only include profile_id if set
    if (form.value.profile_id) {
      payload.profile_id = form.value.profile_id
    }
    
    // Only include admin_email if set
    if (form.value.admin_email) {
      payload.admin_email = form.value.admin_email
    }

    if (isNew.value) {
      await systemAPI.createTenant(payload)
      toast?.success('Tenant created successfully')
    } else {
      await systemAPI.updateTenant(tenantId.value, payload)
      toast?.success('Tenant updated successfully')
    }
    router.push('/system/tenants')
  } catch (e) {
    console.error('Failed to save tenant:', e)
    toast?.error(e.response?.data?.error || e.message, 'Failed to save tenant')
  } finally {
    saving.value = false
  }
}

const impersonateTenant = () => {
  if (tenantId.value) {
    auth.impersonate(tenantId.value)
    router.push('/admin')
  }
}

// --- Deletion flow ---
const resourceLabels = {
  extensions: 'Extensions',
  users: 'Users',
  devices: 'Devices',
  voicemail_boxes: 'Voicemail Boxes',
  voicemail_messages: 'Voicemail Messages',
  recordings: 'Audio Recordings',
  call_recordings: 'Call Recordings',
  ivr_menus: 'IVR Menus',
  queues: 'Call Queues',
  ring_groups: 'Ring Groups',
  conferences: 'Conferences',
  call_flows: 'Call Flows',
  time_conditions: 'Time Conditions',
  feature_codes: 'Feature Codes',
  dialplans: 'Dial Plans',
  page_groups: 'Page Groups',
  speed_dial_groups: 'Speed Dial Groups',
  contacts: 'Contacts',
  conversations: 'SMS Conversations',
  messages: 'SMS Messages',
  broadcasts: 'Broadcast Campaigns',
  fax_boxes: 'Fax Boxes',
  fax_jobs: 'Fax Jobs',
  hotel_rooms: 'Hotel Rooms',
  chat_threads: 'Chat Threads',
  cdr_records: 'CDR Records',
  call_blocks: 'Call Blocks',
  holiday_lists: 'Holiday Lists',
  locations: 'Locations',
  media_files: 'Media Files',
  audit_logs: 'Audit Logs',
  client_registrations: 'Client Registrations'
}

const resourceList = computed(() => {
  if (!deletionPreview.value) return []
  return Object.entries(deletionPreview.value)
    .map(([key, count]) => ({ key, label: resourceLabels[key] || key, count }))
    .filter(item => item.count > 0)
    .sort((a, b) => b.count - a.count)
})

const totalResources = computed(() => {
  if (!deletionPreview.value) return 0
  return Object.values(deletionPreview.value).reduce((sum, v) => sum + v, 0)
})

const startDeletion = async () => {
  deletionLoading.value = true
  try {
    const response = await systemAPI.previewTenantDeletion(tenantId.value)
    deletionPreview.value = response.data?.data || response.data
    deleteConfirmText.value = ''
    showDeleteModal.value = true
  } catch (e) {
    console.error('Failed to load deletion preview:', e)
    toast?.error('Failed to load resource summary')
  } finally {
    deletionLoading.value = false
  }
}

const confirmDeletion = async () => {
  if (deleteConfirmText.value !== form.value.name) return
  deleting.value = true
  try {
    await systemAPI.deleteTenant(tenantId.value)
    toast?.success(`Tenant "${form.value.name}" has been permanently deleted`)
    router.push('/system/tenants')
  } catch (e) {
    console.error('Failed to delete tenant:', e)
    toast?.error(e.response?.data?.error || e.message, 'Failed to delete tenant')
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.back-link { background: none; border: none; color: var(--text-muted); padding: 0; font-size: 11px; cursor: pointer; }
.back-link:hover { color: var(--primary-color); text-decoration: underline; }

.settings-layout { display: flex; gap: 24px; align-items: flex-start; }

/* Navigation */
.settings-nav { width: 180px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 12px; position: sticky; top: 20px; }
.nav-group { margin-bottom: 16px; }
.nav-group:last-child { margin-bottom: 0; }
.nav-group-title { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); padding: 8px 12px 4px; display: block; }

.nav-item { display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-radius: var(--radius-sm); cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-main); }
.nav-item:hover { background: var(--bg-app); }
.nav-item.active { background: var(--primary-light); color: var(--primary-color); }
.nav-icon { width: 16px; height: 16px; opacity: 0.7; }
.nav-item.active .nav-icon { opacity: 1; }

/* Content */
.settings-content { flex: 1; max-width: 600px; }
.settings-panel { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 24px; }

.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.panel-header h3 { font-size: 16px; font-weight: 600; margin: 0; }
.panel-desc { font-size: 13px; color: var(--text-muted); margin-bottom: 20px; }

/* Setting Cards */
.setting-card { padding: 14px; background: var(--bg-app); border-radius: var(--radius-sm); margin-bottom: 10px; }
.setting-card.toggle-card { background: white; border: 1px solid var(--border-color); }

.setting-row { display: flex; justify-content: space-between; align-items: center; gap: 16px; }
.setting-info { flex: 1; }
.setting-info h4 { font-size: 14px; font-weight: 600; margin: 0 0 2px; }
.setting-info p { font-size: 12px; color: var(--text-muted); margin: 0; }

/* Info Card */
.info-card { padding: 16px; background: #eff6ff; border: 1px solid #bfdbfe; border-radius: var(--radius-sm); margin-top: 16px; }
.info-label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: #1e40af; margin-bottom: 8px; }
.info-grid { display: grid; grid-template-columns: auto 1fr; gap: 6px 12px; font-size: 12px; }
.info-key { color: #1e40af; font-weight: 600; }
.info-grid code { font-size: 11px; background: white; padding: 4px 8px; border-radius: 4px; }

/* Form Elements */
.input-field { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; min-width: 200px; }
.input-field.code { font-family: monospace; background: white; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.divider { height: 1px; background: var(--border-color); margin: 20px 0; }

/* Toggle Switch */
.switch { position: relative; display: inline-block; width: 42px; height: 24px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .3s; }
.slider:before { position: absolute; content: ""; height: 18px; width: 18px; left: 3px; bottom: 3px; background-color: white; transition: .3s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(18px); }
.slider.round { border-radius: 24px; }
.slider.round:before { border-radius: 50%; }

/* Color Picker */
.color-picker { display: flex; align-items: center; gap: 8px; }
.color-picker input[type="color"] { width: 36px; height: 36px; border: 1px solid var(--border-color); padding: 0; cursor: pointer; border-radius: 6px; }

/* Buttons */
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary.small { padding: 6px 12px; font-size: 12px; }

/* Danger Zone Nav */
.nav-item-danger { color: #dc2626 !important; }
.nav-item-danger:hover { background: #fef2f2 !important; }
.nav-item-danger.active { background: #fee2e2 !important; color: #dc2626 !important; }
.nav-item-danger .nav-icon { opacity: 1; color: #dc2626; }

/* Danger Zone Panel */
.danger-panel { border-color: #fecaca; }
.danger-title { color: #dc2626; }

.danger-card {
  border: 2px solid #fecaca;
  border-radius: var(--radius-md);
  padding: 20px;
  background: #fef2f2;
}

.danger-card-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
}

.danger-info h4 { font-size: 15px; font-weight: 600; margin: 0 0 6px; color: #991b1b; }
.danger-info p { font-size: 13px; color: #7f1d1d; margin: 0; line-height: 1.5; }

.btn-danger {
  background: #dc2626;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  font-size: 13px;
  transition: background 0.2s;
}
.btn-danger:hover { background: #b91c1c; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }

/* Deletion Confirmation Modal */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-panel {
  background: white;
  border-radius: var(--radius-lg, 12px);
  width: 520px;
  max-width: 90vw;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.danger-modal { border: 2px solid #fecaca; }

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
}
.modal-header h3 { margin: 0; font-size: 18px; }
.danger-header { background: #fef2f2; }
.danger-header h3 { color: #991b1b; }

.modal-body { padding: 20px 24px; }

.modal-warning {
  font-size: 14px;
  line-height: 1.6;
  margin: 0 0 16px;
  color: #374151;
}

.resource-summary {
  background: #f9fafb;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
  margin-bottom: 20px;
  max-height: 240px;
  overflow-y: auto;
}

.resource-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  border-bottom: 1px solid #f3f4f6;
  font-size: 13px;
}
.resource-row:last-child { border-bottom: none; }

.resource-label { color: #374151; }
.resource-count { font-weight: 700; color: #dc2626; font-family: var(--font-mono, monospace); }

.confirm-input-section {
  border-top: 1px solid var(--border-color);
  padding-top: 16px;
}
.confirm-input-section p { font-size: 13px; margin: 0 0 8px; color: #374151; }

.confirm-input {
  width: 100%;
  padding: 10px 12px;
  border: 2px solid #fecaca;
  border-radius: var(--radius-sm);
  font-size: 14px;
  transition: border-color 0.2s;
}
.confirm-input:focus { outline: none; border-color: #dc2626; }

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  background: #f9fafb;
  border-radius: 0 0 12px 12px;
}
</style>
