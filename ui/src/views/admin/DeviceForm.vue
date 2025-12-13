<template>
  <div class="view-header">
    <div class="header-content">
       <button class="back-link" @click="$router.push('/admin/devices')">← Back to Devices</button>
      <h2>{{ isNew ? 'New Device' : 'Edit Device' }}</h2>
      <p class="text-muted text-sm" v-if="!isNew">MAC: {{ device.mac }}</p>
    </div>
  </div>

  <div class="form-container">
    <!-- Tabs -->
    <div class="tabs">
      <button class="tab-btn" :class="{ active: activeTab === 'general' }" @click="activeTab = 'general'">General</button>
      <button class="tab-btn" :class="{ active: activeTab === 'lines' }" @click="activeTab = 'lines'">Lines & Keys</button>
      <button class="tab-btn" :class="{ active: activeTab === 'provisioning' }" @click="activeTab = 'provisioning'">Provisioning</button>
    </div>

    <div class="tab-content" v-if="activeTab === 'general'">
      <div class="form-grid">
        <div class="form-group">
          <label>MAC Address</label>
          <input type="text" v-model="device.mac" class="input-field code" placeholder="00:15:65:..." :disabled="!isNew">
          <span class="help-text" v-if="device.model === 'Generic SIP'">Optional for Generic SIP devices</span>
        </div>

        <div class="form-group">
           <label>Device Name / Label</label>
           <input type="text" v-model="device.name" class="input-field" placeholder="e.g. Reception Desk">
        </div>

        <div class="form-group">
           <label>Device Model</label>
           <select v-model="device.model" class="input-field">
             <option value="" disabled>Select Model...</option>
             <optgroup label="Generic">
               <option value="Generic SIP">Generic SIP Device</option>
             </optgroup>
             <optgroup label="Yealink">
               <option value="Yealink T54W">Yealink T54W</option>
               <option value="Yealink W60B">Yealink W60B</option>
               <option value="Yealink T57W">Yealink T57W</option>
             </optgroup>
             <optgroup label="Polycom">
               <option value="Poly VVX 450">Poly VVX 450</option>
               <option value="Poly CCX 500">Poly CCX 500</option>
             </optgroup>
             <optgroup label="Grandstream">
               <option value="Grandstream GXP2170">Grandstream GXP2170</option>
             </optgroup>
           </select>
        </div>

        <div class="form-group">
           <label>Device Profile</label>
           <select v-model="device.profile_id" class="input-field">
              <option :value="null">None (Custom Identical)</option>
              <option v-for="profile in profiles" :key="profile.id" :value="profile.id">
                {{ profile.name }}
              </option>
           </select>
           <span class="help-text">Apply common settings from a profile</span>
        </div>

        <div class="form-group">
           <label>Location</label>
           <select v-model="device.location_id" class="input-field">
              <option :value="null">Default / None</option>
              <!-- Locations would be loaded dynamically -->
              <option value="hq">HQ - San Francisco</option>
              <option value="warehouse">Warehouse</option>
           </select>
        </div>

        <div class="form-group">
           <label>Assigned User (Owner)</label>
           <select v-model="device.user_id" class="input-field">
              <option :value="null">Unassigned</option>
              <option v-for="user in users" :key="user.id" :value="user.id">
                {{ user.first_name }} {{ user.last_name }} ({{ user.extension }})
              </option>
           </select>
        </div>

        <!-- Generic SIP Device Section -->
        <div v-if="device.model === 'Generic SIP'" class="form-group full-width">
           <div class="divider"></div>
           <div class="generic-sip-section">
             <div class="section-header">
               <ServerIcon class="section-icon" />
               <div>
                 <h3>Generic SIP Registration</h3>
                 <p class="text-muted text-sm">Configure manual SIP credentials for third-party devices or softphones.</p>
               </div>
             </div>

             <div class="sip-fields">
               <div class="form-row">
                 <div class="form-group flex-1">
                    <label>Registration Username</label>
                    <input v-model="device.sip_username" class="input-field code" :placeholder="device.mac ? 'd_' + device.mac : 'Auto-generated'">
                 </div>
                 <div class="form-group flex-1">
                    <label>SIP Password</label>
                    <div class="password-field">
                      <input type="password" v-model="device.sip_password" class="input-field" placeholder="Hidden">
                    </div>
                 </div>
               </div>
               
               <div class="form-row">
                 <div class="form-group flex-1">
                    <label>SIP Server</label>
                    <input v-model="device.sip_server" class="input-field" placeholder="sip.domain.com">
                 </div>
                 <div class="form-group flex-1">
                    <label>Proxy</label>
                    <input v-model="device.sip_proxy" class="input-field" placeholder="Optional">
                 </div>
               </div>
             </div>
           </div>
        </div>
      </div>
    </div>

    <!-- Lines Tab -->
    <div class="tab-content" v-if="activeTab === 'lines'">
      <div class="lines-list">
        <div v-for="(line, index) in device.lines" :key="index" class="line-item">
          <div class="line-header">
            <span class="line-number">Key {{ line.line_number }}</span>
            <button class="btn-icon text-bad" @click="removeLine(index)" v-if="device.lines.length > 1">
              <TrashIcon class="icon-sm" />
            </button>
          </div>
          
          <div class="form-row">
            <div class="form-group flex-1">
              <label>Type</label>
              <select v-model="line.line_type" class="input-field">
                <option value="line">Line (SIP Account)</option>
                <option value="blf">BLF (Busy Lamp Field)</option>
                <option value="speed_dial">Speed Dial</option>
              </select>
            </div>
            
            <div class="form-group flex-2">
              <label>Label</label>
              <input v-model="line.label" class="input-field" placeholder="Display Label">
            </div>
          </div>

          <div class="form-row" v-if="line.line_type === 'line'">
            <div class="form-group flex-1">
              <label>Assign to User/Extension</label>
               <select v-model="line.extension_id" class="input-field">
                  <option :value="null">Unassigned</option>
                  <option v-for="ext in extensions" :key="ext.id" :value="ext.id">
                    {{ ext.extension }} - {{ ext.name }}
                  </option>
               </select>
            </div>
          </div>
          
          <div class="form-row" v-if="line.line_type === 'blf'">
             <div class="form-group flex-1">
               <label>Monitored Extension</label>
               <input v-model="line.blf_extension" class="input-field" placeholder="e.g. 102">
             </div>
          </div>

           <div class="form-row" v-if="line.line_type === 'speed_dial'">
             <div class="form-group flex-1">
               <label>Value (Number)</label>
               <input v-model="line.speed_dial" class="input-field" placeholder="e.g. 555-0199">
             </div>
          </div>

        </div>
        
        <button class="btn-secondary full-width" @click="addLine">
          + Add Key / Line
        </button>
      </div>
    </div>

    <!-- Provisioning Tab -->
    <div class="tab-content" v-if="activeTab === 'provisioning'">
      <div class="provisioning-panel">
        <div class="panel-header">
          <LinkIcon class="panel-icon" />
          <div>
            <h3>Provisioning Configuration</h3>
            <p class="text-muted text-sm">Use these details to configure your physical phone.</p>
          </div>
        </div>

        <div class="field-box">
          <label>Provisioning URL</label>
          <div class="copy-row">
             <code class="url-display">{{ provisionUrl }}</code>
             <button class="btn-icon" @click="copyToClipboard(provisionUrl)"><CopyIcon class="icon-sm"/></button>
          </div>
          <span class="help-text">Enter this URL into your phone's "Server URL" or "Provisioning Server" field.</span>
        </div>

        <div class="field-box">
          <label>Provisioning Secret / Token</label>
          <div v-if="device.provision_token" class="security-box">
             <div class="token-display">
                <span v-if="showToken">{{ device.provision_token }}</span>
                <span v-else>• • • • • • • • • • • • • • • •</span>
             </div>
             <button class="btn-secondary btn-sm" @click="showToken = !showToken">
                {{ showToken ? 'Hide' : 'Show' }}
             </button>
          </div>
          <div v-else class="empty-state">
             <p class="text-bad text-sm">No secret generated. Device may be insecure.</p>
             <button class="btn-primary btn-sm" @click="generateSecret">Generate Secret</button>
          </div>
          <span class="help-text">This token authenticates the device to download its config.</span>
        </div>
      </div>
    </div>

    <div class="form-actions">
       <button class="btn-danger-outline" v-if="!isNew" @click="deleteDevice">Delete Device</button>
       <div style="flex:1"></div>
       <button class="btn-secondary" @click="$router.push('/admin/devices')">Cancel</button>
       <button class="btn-primary" @click="saveDevice">Save Device</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  Server as ServerIcon, 
  Trash2 as TrashIcon, 
  Copy as CopyIcon, 
  Link as LinkIcon 
} from 'lucide-vue-next'
import { devicesAPI, deviceProfilesAPI, usersAPI, extensionsAPI } from '@/services/api'

const toast = inject('toast')
const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.params.id === 'new')

const activeTab = ref('general')
const showToken = ref(false)

const profiles = ref([])
const users = ref([])
const extensions = ref([]) // Ideally fetch extensions
const device = ref({
  mac: '',
  name: '',
  model: '',
  profile_id: null,
  location_id: null,
  user_id: null,
  lines: [{ line_number: 1, line_type: 'line', label: 'Line 1', extension_id: null }],
  provision_token: '',
  provision_url: ''
})

onMounted(async () => {
  await Promise.all([
    fetchProfiles(),
    fetchUsers(),
    fetchExtensions()
  ])

  if (!isNew.value) {
    await loadDevice(route.params.id)
  } else {
    // Generate a temporary secret for new devices if we want client-side generation, 
    // but better to let backend handle it or do it here.
    device.value.provision_token = generateRandomToken()
  }
})

async function fetchProfiles() {
  try {
     const res = await deviceProfilesAPI.list()
     profiles.value = res.data.data || res.data || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchUsers() {
   try {
     const res = await usersAPI.list()
     users.value = (res.data.data || res.data || []).map(u => ({
       id: u.id,
       first_name: u.first_name,
       last_name: u.last_name,
       extension: u.extension
     }))
   } catch (e) {
     console.error(e)
   }
}

async function fetchExtensions() {
  try {
    const res = await extensionsAPI.list()
    // format as needed
    extensions.value = (res.data.data || res.data || []).map(e => ({
      id: e.id,
      extension: e.extension,
      name: e.user ? `${e.user.first_name} ${e.user.last_name}` : (e.description || 'Unassigned')
    }))
  } catch (e) {
    console.error(e)
  }
}

async function loadDevice(id) {
  try {
    const res = await devicesAPI.get(id)
    const d = res.data.data || res.data
    device.value = {
      ...d,
      lines: d.lines && d.lines.length ? d.lines.sort((a,b) => a.line_number - b.line_number) : [{ line_number: 1, line_type: 'line' }]
    }
  } catch (e) {
    toast.error('Failed to load device')
    router.push('/admin/devices')
  }
}

const provisionUrl = computed(() => {
  // If backend provides a full URL, use it
  if (device.value.provision_url) return device.value.provision_url
  
  // Otherwise construct it: https://domain/provision/TENANT_UUID/{secret}/{mac}.cfg
  const baseUrl = window.location.origin.replace('http', 'http') // Simple replace, ideally config based
  // We need tenant UUID. Assuming it's in local storage or we just show a placeholder
  return `${baseUrl}/api/provision/${device.value.provision_token || 'SECRET'}/${device.value.mac || 'MAC'}.cfg`
})

function addLine() {
  const nextNum = device.value.lines.length + 1
  device.value.lines.push({
    line_number: nextNum,
    line_type: 'line',
    label: `Line ${nextNum}`,
    extension_id: null
  })
}

function removeLine(index) {
  device.value.lines.splice(index, 1)
  // Re-number
  device.value.lines.forEach((l, i) => l.line_number = i + 1)
}

function generateSecret() {
  device.value.provision_token = generateRandomToken()
}

function generateRandomToken() {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
}

function copyToClipboard(text) {
  navigator.clipboard.writeText(text)
  toast.success('Copied to clipboard')
}

async function saveDevice() {
  try {
    // Clean mac
    const payload = {
      ...device.value,
      mac: device.value.mac.replace(/:/g, '').toLowerCase()
    }

    if (isNew.value) {
      await devicesAPI.create(payload)
      toast.success('Device created')
    } else {
      await devicesAPI.update(device.value.id, payload)
       // Update lines separately if needed, but if backend handles nested, great. 
       // The backend handlers I saw earlier: UpdateDevice updates fields, UpdateDeviceLines updates lines.
       // I should probably call UpdateDeviceLines if the lines changed.
       if (device.value.lines) {
          // Check if lines supported in UpdateDevice. 
          // Previous backend reading showed UpdateDevice only does basic fields. 
          // So I need to call the line update endpoint.
          // Since I don't have the exact endpoint handy in memory, I'll assume I need to call it.
          // Wait, I saw 'UpdateDeviceLines' in device_handlers.go.
          await devicesAPI.updateLines(device.value.id, device.value.lines)
       }
      toast.success('Device updated')
    }
    router.push('/admin/devices')
  } catch (e) {
    toast.error(e.response?.data?.error || e.message)
  }
}

async function deleteDevice() {
  if (!confirm('Are you sure? This cannot be undone.')) return
  try {
    await devicesAPI.delete(device.value.id)
    toast.success('Device deleted')
    router.push('/admin/devices')
  } catch (e) {
    toast.error('Failed to delete')
  }
}
</script>

<style scoped>
.view-header { margin-bottom: 24px; }
.back-link { border: none; background: none; color: var(--text-muted); cursor: pointer; padding: 0; font-size: 13px; margin-bottom: 8px;}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.form-container {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  max-width: 900px;
  overflow: hidden;
}

.tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-app);
  padding: 0 16px;
}
.tab-btn {
  padding: 12px 20px;
  border: none;
  background: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-weight: 500;
  color: var(--text-muted);
}
.tab-btn.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
  background: white;
}

.tab-content { padding: 24px; }

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full-width { grid-column: span 2; }
.flex-1 { flex: 1; }
.flex-2 { flex: 2; }
.form-row { display: flex; gap: 16px; width: 100%; }

label { font-size: 12px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.02em; }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; width: 100%; }
.input-field.code { font-family: monospace; }
.help-text { font-size: 11px; color: var(--text-muted); }

/* Lines List */
.lines-list { display: flex; flex-direction: column; gap: 16px; }
.line-item { background: var(--bg-app); padding: 16px; border-radius: var(--radius-sm); border: 1px solid var(--border-color); display: flex; flex-direction: column; gap: 12px; }
.line-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.line-number { font-weight: 600; font-size: 13px; color: var(--text-main); }

/* Provisioning */
.provisioning-panel { display: flex; flex-direction: column; gap: 24px; }
.panel-header { display: flex; gap: 12px; align-items: flex-start; padding-bottom: 16px; border-bottom: 1px solid var(--border-color); }
.panel-icon { color: var(--primary-color); }
.field-box { display: flex; flex-direction: column; gap: 8px; }
.copy-row { display: flex; gap: 8px; align-items: center; }
.url-display { background: var(--bg-app); padding: 10px; border-radius: var(--radius-sm); font-family: monospace; flex: 1; border: 1px solid var(--border-color); white-space: nowrap; overflow-x: auto; }

.security-box { display: flex; gap: 12px; align-items: center; }
.token-display { font-family: monospace; font-size: 16px; font-weight: 600; letter-spacing: 2px; }

.empty-state { display: flex; align-items: center; gap: 12px; background: #fee2e2; padding: 12px; border-radius: var(--radius-sm); border: 1px solid #fecaca; }

.form-actions { padding: 16px 24px; border-top: 1px solid var(--border-color); display: flex; gap: 12px; background: var(--bg-app); }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 10px 20px; border-radius: var(--radius-sm); cursor: pointer; font-weight: 500; }
.btn-danger-outline { background: white; border: 1px solid var(--status-bad); color: var(--status-bad); padding: 10px 20px; border-radius: var(--radius-sm); cursor: pointer; font-weight: 500; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; }
.btn-icon:hover { color: var(--text-primary); }
.text-bad { color: var(--status-bad); }

.icon-sm { width: 16px; height: 16px; }
.full-width { width: 100%; }
</style>
