<template>
  <div class="provisioning-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Provisioning</h2>
        <p class="text-muted text-sm">Device provisioning settings and template overrides for your organization.</p>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ devices.registered }}</div>
        <div class="stat-label">Registered Devices</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ devices.total }}</div>
        <div class="stat-label">Total Devices</div>
      </div>
    </div>

    <!-- Provisioning URL Section -->
    <div class="section-card">
      <div class="section-header">
        <div class="section-title">
          <LinkIcon class="section-icon" />
          <div>
            <h3>Provisioning URL</h3>
            <p class="text-muted text-sm">Use this URL to configure your IP phones for auto-provisioning.</p>
          </div>
        </div>
        <button 
          class="btn-secondary" 
          @click="regenerateSecret" 
          :disabled="regenerating || hasExistingSecret"
          :title="hasExistingSecret ? 'Secret already exists. Delete existing secret first.' : 'Generate provisioning secret'"
        >
          <RefreshCwIcon class="btn-icon" />
          {{ regenerating ? 'Generating...' : (hasExistingSecret ? 'Secret Exists' : 'Generate Secret') }}
        </button>
      </div>
      
      <div class="url-display">
        <div class="url-box">
          <code>{{ provisioningUrl }}</code>
          <button class="copy-btn" @click="copyUrl" v-tooltip="'Copy URL'">
            <CopyIcon class="icon-sm" />
          </button>
        </div>
        <p class="url-hint">
          Replace <code>{MAC}</code> with the device's MAC address (e.g., <code>001565ABCDEF</code>)
        </p>
      </div>

      <div class="url-examples">
        <h5>Example Configurations</h5>
        <div class="example-grid">
          <div class="example-item">
            <span class="brand">Yealink</span>
            <code>{{ provisioningUrl.replace('{MAC}', '$MAC') }}</code>
          </div>
          <div class="example-item">
            <span class="brand">Poly/Polycom</span>
            <code>{{ provisioningUrl.replace('{MAC}', '[PHONE_MAC_ADDRESS]') }}</code>
          </div>
          <div class="example-item">
            <span class="brand">Grandstream</span>
            <code>{{ provisioningUrl.replace('{MAC}', '&lt;MAC&gt;') }}</code>
          </div>
        </div>
      </div>
    </div>

    <!-- Syslog Configuration -->
    <div class="section-card">
      <div class="section-header">
        <div class="section-title">
          <FileTextIcon class="section-icon" />
          <div>
            <h3>Syslog Configuration</h3>
            <p class="text-muted text-sm">Configure where devices should send their logs.</p>
          </div>
        </div>
        <button class="btn-primary" @click="saveSyslogSettings" :disabled="savingSyslog">
          {{ savingSyslog ? 'Saving...' : 'Save Settings' }}
        </button>
      </div>
      
      <div class="form-grid">
        <div class="form-group">
          <label>Syslog Server</label>
          <input v-model="syslogSettings.server" class="input-field" placeholder="syslog.example.com">
        </div>
        <div class="form-group">
          <label>Port</label>
          <input v-model="syslogSettings.port" class="input-field" type="number" placeholder="514">
        </div>
        <div class="form-group">
          <label>Protocol</label>
          <select v-model="syslogSettings.protocol" class="input-field">
            <option value="udp">UDP</option>
            <option value="tcp">TCP</option>
            <option value="tls">TLS</option>
          </select>
        </div>
        <div class="form-group">
          <label>Log Level</label>
          <select v-model="syslogSettings.level" class="input-field">
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
            <option value="debug">Debug</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { 
  Link as LinkIcon, Copy as CopyIcon, RefreshCw as RefreshCwIcon,
  FileText as FileTextIcon
} from 'lucide-vue-next'
import { devicesAPI, tenantSettingsAPI } from '@/services/api'

const toast = inject('toast')
const loading = ref(false)
const savingSyslog = ref(false)
const regenerating = ref(false)

// Tenant info loaded from API
const tenantInfo = ref({
  uuid: '',
  secret: '',
  domain: window.location.hostname
})

// Provisioning URL - computed from tenant info
const provisioningUrl = computed(() => {
  const baseUrl = window.location.origin
  const uuid = tenantInfo.value.uuid || 'unknown'
  const secret = tenantInfo.value.secret || 'secret'
  return `${baseUrl}/api/provision/${uuid}/${secret}/{MAC}.cfg`
})

// Device stats
const devices = ref({
  registered: 0,
  total: 0
})

// Syslog settings
const syslogSettings = ref({
  server: '',
  port: '514',
  protocol: 'udp',
  level: 'info'
})

// Check if provisioning secret already exists
const hasExistingSecret = computed(() => {
  return tenantInfo.value.secret && tenantInfo.value.secret !== '' && tenantInfo.value.secret !== 'secret'
})

// Load data
const loadData = async () => {
  loading.value = true
  try {
    // Load tenant settings (includes UUID and provisioning secret)
    const settingsResponse = await tenantSettingsAPI.get()
    const settings = settingsResponse.data?.data || settingsResponse.data || {}
    tenantInfo.value = {
      uuid: settings.uuid || '',
      secret: settings.provisioning_secret || '',
      domain: settings.domain || window.location.hostname
    }
    
    // Load devices
    const deviceResponse = await devicesAPI.list()
    const deviceList = deviceResponse.data?.data || deviceResponse.data || []
    devices.value = {
      registered: deviceList.filter(d => d.status === 'Registered').length,
      total: deviceList.length
    }
    
    // Load syslog settings if available
    if (settings.syslog_server) {
      syslogSettings.value = {
        server: settings.syslog_server || '',
        port: settings.syslog_port || '514',
        protocol: settings.syslog_protocol || 'udp',
        level: settings.syslog_level || 'info'
      }
    }
  } catch (e) {
    console.error('Failed to load data:', e)
    toast?.error('Failed to load provisioning data')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)

// Copy URL
const copyUrl = async () => {
  try {
    await navigator.clipboard.writeText(provisioningUrl.value.replace('{MAC}', ''))
    toast?.success('Provisioning URL copied')
  } catch (e) {
    toast?.error('Failed to copy URL')
  }
}

// Generate secret (only if no existing secret)
const regenerateSecret = async () => {
  if (hasExistingSecret.value) {
    toast?.warning('Provisioning secret already exists')
    return
  }
  regenerating.value = true
  try {
    // TODO: Call API to generate tenant secret
    await new Promise(r => setTimeout(r, 1000))
    toast?.success('Secret generated. Configure your devices with the new URL.')
  } catch (e) {
    toast?.error('Failed to generate secret')
  } finally {
    regenerating.value = false
  }
}

// Save syslog settings
const saveSyslogSettings = async () => {
  savingSyslog.value = true
  try {
    // TODO: Call API to save syslog settings
    await new Promise(r => setTimeout(r, 500))
    toast?.success('Syslog settings saved')
  } catch (e) {
    toast?.error('Failed to save syslog settings')
  } finally {
    savingSyslog.value = false
  }
}
</script>

<style scoped>
.provisioning-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 24px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.section-card { background: white; border: 1px solid var(--border-color); border-radius: 12px; padding: 20px; margin-bottom: 24px; }
.section-header { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 20px; flex-wrap: wrap; }
.section-title { display: flex; gap: 12px; align-items: flex-start; }
.section-icon { width: 24px; height: 24px; color: var(--primary-color); flex-shrink: 0; margin-top: 2px; }
.section-title h3 { margin: 0 0 4px; font-size: 16px; }

.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); }
.btn-icon { width: 14px; height: 14px; }
.btn-sm { display: flex; align-items: center; gap: 4px; padding: 6px 10px; background: white; border: 1px solid var(--border-color); border-radius: 4px; font-size: 11px; cursor: pointer; }
.btn-sm:hover { border-color: var(--primary-color); color: var(--primary-color); }

/* URL Display */
.url-display { margin-bottom: 20px; }
.url-box { display: flex; align-items: center; gap: 8px; background: #f8fafc; border: 1px solid var(--border-color); border-radius: 8px; padding: 12px 16px; }
.url-box code { flex: 1; font-size: 13px; word-break: break-all; color: var(--text-primary); }
.copy-btn { background: white; border: 1px solid var(--border-color); border-radius: 4px; padding: 6px; cursor: pointer; }
.copy-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.url-hint { font-size: 12px; color: var(--text-muted); margin-top: 8px; }
.url-hint code { background: #e2e8f0; padding: 2px 6px; border-radius: 3px; font-size: 11px; }

.url-examples { background: #f8fafc; border-radius: 8px; padding: 16px; }
.url-examples h5 { font-size: 12px; text-transform: uppercase; color: var(--text-muted); margin: 0 0 12px; }
.example-grid { display: flex; flex-direction: column; gap: 8px; }
.example-item { display: flex; align-items: center; gap: 12px; font-size: 12px; }
.example-item .brand { min-width: 100px; font-weight: 600; color: var(--text-primary); }
.example-item code { background: white; padding: 6px 10px; border-radius: 4px; border: 1px solid var(--border-color); font-size: 11px; flex: 1; word-break: break-all; }

/* Form Grid */
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; }
.input-field:focus { outline: none; border-color: var(--primary-color); }

/* Template Grid */
.template-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }
.template-card { background: #f8fafc; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.template-card:hover { border-color: var(--primary-color); }
.template-header { display: flex; justify-content: space-between; padding: 14px; background: white; border-bottom: 1px solid var(--border-color); }
.template-info h4 { margin: 4px 0 2px; font-size: 14px; }
.template-actions { display: flex; gap: 4px; }
.btn-icon { background: none; border: none; cursor: pointer; padding: 4px; color: var(--text-muted); }
.btn-icon:hover { color: var(--text-primary); }
.btn-icon.danger:hover { color: #ef4444; }
.manufacturer-badge { font-size: 10px; font-weight: 600; text-transform: uppercase; background: var(--primary-light); color: var(--primary-color); padding: 2px 6px; border-radius: 3px; }
.template-meta { display: flex; gap: 16px; padding: 10px 14px; font-size: 11px; color: var(--text-muted); }
.template-meta span { display: flex; align-items: center; gap: 4px; }

/* System Templates Grid */
.filter-tabs { display: flex; gap: 4px; flex-wrap: wrap; }
.filter-tab { padding: 6px 12px; border: 1px solid var(--border-color); background: white; border-radius: 6px; font-size: 12px; cursor: pointer; }
.filter-tab:hover { border-color: var(--primary-color); }
.filter-tab.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }

.system-templates-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 12px; }
.system-template-card { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: #f8fafc; border: 1px solid var(--border-color); border-radius: 6px; }
.system-template-card:hover { border-color: var(--primary-color); }
.system-template-card h5 { margin: 0 0 2px; font-size: 13px; }

/* Empty State */
.empty-state { text-align: center; padding: 40px 20px; }
.empty-icon { width: 48px; height: 48px; color: var(--text-muted); margin-bottom: 16px; }
.empty-state h4 { margin: 0 0 8px; }
.empty-state p { color: var(--text-muted); margin-bottom: 16px; }

/* Loading */
.loading-state { display: flex; align-items: center; justify-content: center; gap: 12px; padding: 40px; color: var(--text-muted); }
.spinner { width: 20px; height: 20px; border: 2px solid var(--border-color); border-top-color: var(--primary-color); border-radius: 50%; animation: spin 0.8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; padding: 24px; }
.modal-card { background: white; border-radius: 12px; width: 100%; max-width: 500px; max-height: 90vh; overflow: hidden; display: flex; flex-direction: column; }
.modal-card.large { max-width: 900px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Editor Layout */
.editor-layout { display: grid; grid-template-columns: 240px 1fr; gap: 20px; min-height: 400px; }
.editor-sidebar { display: flex; flex-direction: column; gap: 16px; }
.divider { height: 1px; background: var(--border-color); margin: 8px 0; }
.variables-hint h5 { font-size: 11px; text-transform: uppercase; color: var(--text-muted); margin: 0 0 8px; }
.var-list { display: flex; flex-direction: column; gap: 4px; }
.var-list code { font-size: 11px; background: #f1f5f9; padding: 4px 8px; border-radius: 4px; }
.editor-main { display: flex; flex-direction: column; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.code-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #1e293b; color: white; font-size: 12px; }
.code-editor { flex: 1; padding: 12px; font-family: monospace; font-size: 12px; border: none; resize: none; background: #1e293b; color: #e2e8f0; min-height: 300px; }
.code-editor::placeholder { color: #64748b; }

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 12px; height: 12px; }
.text-muted { color: var(--text-muted); }
.text-sm { font-size: 13px; }
.text-xs { font-size: 11px; }

@media (max-width: 768px) {
  .editor-layout { grid-template-columns: 1fr; }
  .section-header { flex-direction: column; }
}
</style>
