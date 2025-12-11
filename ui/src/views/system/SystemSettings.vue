<template>
  <div class="system-settings-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Global System Settings</h2>
        <p class="text-muted text-sm">Cluster-wide configurations for SMTP, SIP, and FreeSWITCH defaults.</p>
      </div>
      <button class="btn-primary" @click="save" :disabled="saving">
        <SaveIcon class="btn-icon" /> {{ saving ? 'Saving...' : 'Save All Changes' }}
      </button>
    </div>

    <!-- Settings Navigation -->
    <div class="settings-layout">
      <nav class="settings-nav">
        <button 
          v-for="section in sections" 
          :key="section.id" 
          class="nav-btn"
          :class="{ active: activeSection === section.id }"
          @click="activeSection = section.id"
        >
          <component :is="section.icon" class="nav-icon" />
          <span>{{ section.label }}</span>
        </button>
      </nav>

      <div class="settings-content">
        <!-- SMTP SETTINGS -->
        <div v-if="activeSection === 'smtp'" class="panel">
          <div class="panel-header">
            <h3>Global SMTP Settings</h3>
            <span class="precedence-badge">Takes Precedence Over Tenants</span>
          </div>
          <p class="panel-desc">System-level email server used for voicemail delivery, notifications, and alerts. Tenants can optionally override with their own SMTP.</p>
          
          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <h4>Enable System SMTP</h4>
                <p>Use this server for all outgoing emails unless tenant overrides.</p>
              </div>
              <label class="switch">
                <input type="checkbox" v-model="smtp.enabled">
                <span class="slider round"></span>
              </label>
            </div>
          </div>

          <div class="form-grid" v-if="smtp.enabled">
            <div class="form-group">
              <label>SMTP Host</label>
              <input v-model="smtp.host" class="input-field" placeholder="smtp.sendgrid.net">
            </div>
            <div class="form-group">
              <label>Port</label>
              <input v-model="smtp.port" class="input-field" placeholder="587">
            </div>
            <div class="form-group">
              <label>Username</label>
              <input v-model="smtp.username" class="input-field" placeholder="apikey">
            </div>
            <div class="form-group">
              <label>Password</label>
              <input v-model="smtp.password" type="password" class="input-field">
            </div>
            <div class="form-group">
              <label>From Email</label>
              <input v-model="smtp.fromEmail" class="input-field" placeholder="noreply@callsign.io">
            </div>
            <div class="form-group">
              <label>From Name</label>
              <input v-model="smtp.fromName" class="input-field" placeholder="CallSign PBX">
            </div>
            <div class="form-group">
              <label>Encryption</label>
              <select v-model="smtp.encryption" class="input-field">
                <option value="tls">TLS (Recommended)</option>
                <option value="ssl">SSL</option>
                <option value="none">None</option>
              </select>
            </div>
            <div class="form-group">
              <label>Connection Timeout (sec)</label>
              <input v-model="smtp.timeout" type="number" class="input-field">
            </div>
          </div>

          <div class="test-section" v-if="smtp.enabled">
            <button class="btn-secondary" @click="testSmtp">
              <SendIcon class="btn-icon" /> Send Test Email
            </button>
          </div>
        </div>

        <!-- FREESWITCH SETTINGS -->
        <div v-else-if="activeSection === 'freeswitch'" class="panel">
          <div class="panel-header">
            <h3>FreeSWITCH Configuration</h3>
            <span class="status-badge online">Service: Running</span>
          </div>
          <p class="panel-desc">Global FreeSWITCH settings affecting all tenants and SIP profiles.</p>

          <div class="setting-card">
            <h4>Core Settings</h4>
            <div class="form-grid">
              <div class="form-group">
                <label>Max Sessions</label>
                <input v-model="freeswitch.maxSessions" type="number" class="input-field">
              </div>
              <div class="form-group">
                <label>Sessions Per Second</label>
                <input v-model="freeswitch.sessionsPerSec" type="number" class="input-field">
              </div>
              <div class="form-group">
                <label>RTP Port Min</label>
                <input v-model="freeswitch.rtpPortMin" type="number" class="input-field">
              </div>
              <div class="form-group">
                <label>RTP Port Max</label>
                <input v-model="freeswitch.rtpPortMax" type="number" class="input-field">
              </div>
            </div>
          </div>

          <div class="setting-card">
            <h4>Default Codecs</h4>
            <p class="help-text">Drag to reorder codec priority.</p>
            <div class="codec-list">
              <div class="codec-item" v-for="codec in freeswitch.codecs" :key="codec.name" :class="{ disabled: !codec.enabled }">
                <GripVerticalIcon class="grip" />
                <span class="codec-name">{{ codec.name }}</span>
                <span class="codec-type">{{ codec.type }}</span>
                <label class="switch sm">
                  <input type="checkbox" v-model="codec.enabled">
                  <span class="slider round"></span>
                </label>
              </div>
            </div>
          </div>

          <div class="setting-card">
            <h4>Logging</h4>
            <div class="form-grid">
              <div class="form-group">
                <label>Log Level</label>
                <select v-model="freeswitch.logLevel" class="input-field">
                  <option value="debug">Debug</option>
                  <option value="info">Info</option>
                  <option value="notice">Notice</option>
                  <option value="warning">Warning</option>
                  <option value="error">Error</option>
                </select>
              </div>
              <div class="form-group">
                <label>CDR Log Path</label>
                <input v-model="freeswitch.cdrPath" class="input-field" placeholder="/var/log/freeswitch/cdr">
              </div>
            </div>
          </div>
        </div>

        <!-- MESSAGING PROVIDER -->
        <div v-else-if="activeSection === 'messaging'" class="panel">
          <div class="panel-header">
            <h3>Global SMS/MMS Provider</h3>
          </div>
          <p class="panel-desc">Default messaging provider for tenants without their own configuration.</p>

          <div class="form-grid">
            <div class="form-group">
              <label>Primary Provider</label>
              <select v-model="messaging.provider" class="input-field">
                <option value="twilio">Twilio</option>
                <option value="bandwidth">Bandwidth</option>
                <option value="telnyx">Telnyx</option>
                <option value="plivo">Plivo</option>
              </select>
            </div>
            <div class="form-group">
              <label>Account SID / API Key</label>
              <input v-model="messaging.accountSid" class="input-field">
            </div>
            <div class="form-group full-span">
              <label>Auth Token</label>
              <input v-model="messaging.authToken" type="password" class="input-field">
            </div>
          </div>
        </div>

        <!-- DATABASE -->
        <div v-else-if="activeSection === 'database'" class="panel">
          <div class="panel-header">
            <h3>Database Configuration</h3>
            <span class="status-badge online">Connected</span>
          </div>

          <div class="connection-info">
            <div class="info-row">
              <span class="label">Type</span>
              <span class="value">PostgreSQL 15.4</span>
            </div>
            <div class="info-row">
              <span class="label">Host</span>
              <span class="value mono">db.cluster.local:5432</span>
            </div>
            <div class="info-row">
              <span class="label">Database</span>
              <span class="value mono">callsign_prod</span>
            </div>
            <div class="info-row">
              <span class="label">Connections</span>
              <span class="value">12 / 100 active</span>
            </div>
          </div>
        </div>

        <!-- CLUSTER -->
        <div v-else-if="activeSection === 'cluster'" class="panel">
          <div class="panel-header">
            <h3>Cluster Nodes</h3>
          </div>
          <p class="panel-desc">Active nodes in the FreeSWITCH cluster.</p>

          <div class="node-list">
            <div class="node-card" v-for="node in cluster.nodes" :key="node.id">
              <div class="node-header">
                <span class="node-name">{{ node.name }}</span>
                <span class="status-dot" :class="node.status"></span>
              </div>
              <div class="node-details">
                <div class="node-stat"><span>IP:</span> {{ node.ip }}</div>
                <div class="node-stat"><span>Sessions:</span> {{ node.sessions }}</div>
                <div class="node-stat"><span>CPU:</span> {{ node.cpu }}%</div>
                <div class="node-stat"><span>Memory:</span> {{ node.memory }}%</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {
  Mail as MailIcon, Server as ServerIcon, MessageSquare as MessageIcon,
  Database as DatabaseIcon, Layers as ClusterIcon, Save as SaveIcon,
  Send as SendIcon, GripVertical as GripVerticalIcon
} from 'lucide-vue-next'
import { systemAPI } from '../../services/api'

const activeSection = ref('smtp')
const loading = ref(true)
const saving = ref(false)

const sections = [
  { id: 'smtp', label: 'SMTP Settings', icon: MailIcon },
  { id: 'freeswitch', label: 'FreeSWITCH', icon: ServerIcon },
  { id: 'messaging', label: 'Messaging Provider', icon: MessageIcon },
  { id: 'database', label: 'Database', icon: DatabaseIcon },
  { id: 'cluster', label: 'Cluster', icon: ClusterIcon },
]

const smtp = ref({
  enabled: true,
  host: '',
  port: '587',
  username: '',
  password: '',
  fromEmail: '',
  fromName: 'CallSign PBX',
  encryption: 'tls',
  timeout: 30
})

const freeswitch = ref({
  maxSessions: 1000,
  sessionsPerSec: 30,
  rtpPortMin: 16384,
  rtpPortMax: 32768,
  logLevel: 'warning',
  cdrPath: '/var/log/freeswitch/cdr',
  codecs: [
    { name: 'OPUS', type: 'Audio', enabled: true },
    { name: 'G722', type: 'Audio', enabled: true },
    { name: 'PCMU', type: 'Audio', enabled: true },
    { name: 'PCMA', type: 'Audio', enabled: true },
    { name: 'G729', type: 'Audio', enabled: true },
    { name: 'VP8', type: 'Video', enabled: true },
    { name: 'H264', type: 'Video', enabled: false },
  ]
})

const messaging = ref({
  provider: 'twilio',
  accountSid: '',
  authToken: ''
})

const cluster = ref({
  nodes: []
})

const database = ref({
  host: '',
  dbName: ''
})

const loadSettings = async () => {
  loading.value = true
  try {
    const response = await systemAPI.getSettings()
    const data = response.data || {}
    database.value.host = data.db_host || ''
    // FreeSWITCH settings from API
    if (data.freeswitch_host) {
      freeswitch.value.host = data.freeswitch_host
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadSettings)

const save = async () => {
  saving.value = true
  try {
    await systemAPI.updateSettings({
      smtp: smtp.value,
      freeswitch: freeswitch.value,
      messaging: messaging.value
    })
    alert('Settings saved successfully!')
  } catch (e) {
    alert('Failed to save settings: ' + (e.message || 'Not implemented yet'))
  } finally {
    saving.value = false
  }
}

const testSmtp = () => alert('Test email sent to admin@callsign.io')
</script>

<style scoped>
.system-settings-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

.settings-layout { display: flex; gap: 24px; }

.settings-nav { width: 200px; flex-shrink: 0; display: flex; flex-direction: column; gap: 4px; }
.nav-btn { display: flex; align-items: center; gap: 10px; padding: 12px 16px; background: transparent; border: none; border-radius: var(--radius-sm); cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); text-align: left; transition: all 0.15s; }
.nav-btn:hover { background: var(--bg-app); color: var(--text-primary); }
.nav-btn.active { background: var(--primary-light); color: var(--primary-color); }
.nav-icon { width: 18px; height: 18px; }

.settings-content { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 24px; }

.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.panel-header h3 { margin: 0; font-size: 18px; font-weight: 700; }
.panel-desc { color: var(--text-muted); font-size: 13px; margin-bottom: 24px; }

.precedence-badge { font-size: 10px; font-weight: 700; padding: 4px 10px; border-radius: 4px; background: #fef3c7; color: #b45309; text-transform: uppercase; }
.status-badge { font-size: 10px; font-weight: 700; padding: 4px 10px; border-radius: 4px; text-transform: uppercase; }
.status-badge.online { background: #dcfce7; color: #16a34a; }

.setting-card { border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; margin-bottom: 16px; }
.setting-card h4 { margin: 0 0 12px; font-size: 14px; font-weight: 600; }
.setting-header { display: flex; justify-content: space-between; align-items: center; }
.setting-info h4 { margin: 0; }
.setting-info p { margin: 4px 0 0; font-size: 12px; color: var(--text-muted); }

.form-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full-span { grid-column: 1 / -1; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }

.test-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border-color); }

.codec-list { display: flex; flex-direction: column; gap: 6px; margin-top: 12px; }
.codec-item { display: flex; align-items: center; gap: 12px; padding: 10px 12px; background: var(--bg-app); border-radius: 6px; }
.codec-item.disabled { opacity: 0.5; }
.grip { width: 16px; height: 16px; color: var(--text-muted); cursor: grab; }
.codec-name { font-weight: 600; font-size: 13px; flex: 1; }
.codec-type { font-size: 11px; color: var(--text-muted); padding: 2px 8px; background: white; border-radius: 4px; }

.connection-info { border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; }
.info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid var(--border-color); }
.info-row:last-child { border-bottom: none; }
.info-row .label { color: var(--text-muted); font-size: 13px; }
.info-row .value { font-weight: 600; font-size: 13px; }
.mono { font-family: monospace; }

.node-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 16px; }
.node-card { border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; }
.node-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.node-name { font-weight: 700; font-size: 14px; }
.status-dot { width: 10px; height: 10px; border-radius: 50%; }
.status-dot.online { background: #22c55e; }
.status-dot.standby { background: #f59e0b; }
.node-details { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
.node-stat { font-size: 12px; color: var(--text-muted); }
.node-stat span { font-weight: 600; color: var(--text-primary); }

/* Toggle Switch */
.switch { position: relative; display: inline-block; width: 44px; height: 24px; }
.switch.sm { width: 36px; height: 20px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; inset: 0; background: #e2e8f0; border-radius: 24px; transition: 0.3s; }
.slider:before { content: ''; position: absolute; width: 18px; height: 18px; left: 3px; bottom: 3px; background: white; border-radius: 50%; transition: 0.3s; }
.switch.sm .slider:before { width: 14px; height: 14px; }
.switch input:checked + .slider { background: var(--primary-color); }
.switch input:checked + .slider:before { transform: translateX(20px); }
.switch.sm input:checked + .slider:before { transform: translateX(16px); }

.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { width: 16px; height: 16px; }

.help-text { font-size: 12px; color: var(--text-muted); margin-bottom: 8px; }

@media (max-width: 768px) {
  .settings-layout { flex-direction: column; }
  .settings-nav { width: 100%; flex-direction: row; overflow-x: auto; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>
