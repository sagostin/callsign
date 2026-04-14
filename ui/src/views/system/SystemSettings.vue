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



        <!-- DATABASE -->
        <div v-else-if="activeSection === 'database'" class="panel">
          <div class="panel-header">
            <h3>Database Configuration</h3>
            <span class="status-badge" :class="database.connected ? 'online' : 'offline'">{{ database.connected ? 'Connected' : 'Disconnected' }}</span>
          </div>

          <div class="connection-info" v-if="database.loaded">
            <div class="info-row">
              <span class="label">Type</span>
              <span class="value">{{ database.type || 'PostgreSQL' }}</span>
            </div>
            <div class="info-row">
              <span class="label">Host</span>
              <span class="value mono">{{ database.host || '—' }}</span>
            </div>
            <div class="info-row">
              <span class="label">Database</span>
              <span class="value mono">{{ database.dbName || '—' }}</span>
            </div>
            <div class="info-row">
              <span class="label">Open Connections</span>
              <span class="value">{{ database.openConns }} / {{ database.maxConns }}</span>
            </div>
            <div class="info-row">
              <span class="label">In Use</span>
              <span class="value">{{ database.inUse }}</span>
            </div>
            <div class="info-row">
              <span class="label">Idle</span>
              <span class="value">{{ database.idle }}</span>
            </div>
          </div>
          <div v-else class="panel-desc" style="text-align:center; padding: 40px;">Loading database info...</div>
        </div>

        <!-- CLUSTER -->
        <div v-else-if="activeSection === 'cluster'" class="panel">
          <div class="panel-header">
            <h3>Cluster Nodes</h3>
            <span class="deferred-badge">Deferred</span>
          </div>
          <!-- TODO: Implement cluster management for multi-node FreeSWITCH orchestration
               - Node health monitoring and status
               - Failover configuration
               - Shared ESL connection pooling
               - Cluster-wide configuration sync
          -->
          <div class="deferred-notice">
            <ClusterIcon class="deferred-icon" />
            <p class="panel-desc">Cluster management is planned for future release.</p>
            <p class="panel-desc">This will enable multi-node FreeSWITCH cluster orchestration, node health monitoring, and failover configuration.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {
  Mail as MailIcon, Server as ServerIcon,
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


const cluster = ref({
  nodes: []
})

const database = ref({
  host: '',
  dbName: '',
  type: 'PostgreSQL',
  connected: false,
  openConns: 0,
  maxConns: 0,
  inUse: 0,
  idle: 0,
  loaded: false
})

const loadSettings = async () => {
  loading.value = true
  try {
    const [settingsResp, statusResp] = await Promise.all([
      systemAPI.getSettings(),
      systemAPI.getStatus()
    ])
    const settings = settingsResp.data || {}
    const status = statusResp.data || {}

    // FreeSWITCH settings
    if (settings.freeswitch) {
      freeswitch.value.maxSessions = settings.freeswitch.max_sessions ?? freeswitch.value.maxSessions
      freeswitch.value.sessionsPerSec = settings.freeswitch.sessions_per_second ?? freeswitch.value.sessionsPerSec
      freeswitch.value.rtpPortMin = settings.freeswitch.rtp_port_min ?? freeswitch.value.rtpPortMin
      freeswitch.value.rtpPortMax = settings.freeswitch.rtp_port_max ?? freeswitch.value.rtpPortMax
      freeswitch.value.logLevel = settings.freeswitch.log_level ?? freeswitch.value.logLevel
      freeswitch.value.cdrPath = settings.freeswitch.cdr_path ?? freeswitch.value.cdrPath
      if (settings.freeswitch.codecs?.length) {
        freeswitch.value.codecs = freeswitch.value.codecs.map((c, i) => ({
          ...c,
          enabled: settings.freeswitch.codecs[i]?.enabled ?? c.enabled
        }))
      }
    }

    // SMTP settings
    if (settings.smtp) {
      smtp.value.enabled = settings.smtp.enabled ?? smtp.value.enabled
      smtp.value.host = settings.smtp.host || smtp.value.host
      smtp.value.port = settings.smtp.port || smtp.value.port
      smtp.value.username = settings.smtp.username || smtp.value.username
      smtp.value.fromEmail = settings.smtp.from_email || smtp.value.fromEmail
      smtp.value.fromName = settings.smtp.from_name || smtp.value.fromName
      smtp.value.encryption = settings.smtp.encryption || smtp.value.encryption
      smtp.value.timeout = settings.smtp.timeout ?? smtp.value.timeout
    }

    // Database status
    if (status.database) {
      database.value.host = status.database.host || ''
      database.value.dbName = status.database.name || ''
      database.value.type = status.database.type || 'PostgreSQL'
      database.value.connected = status.database.connected !== false
      database.value.openConns = status.database.open_connections || 0
      database.value.maxConns = status.database.max_connections || 0
      database.value.inUse = status.database.in_use || 0
      database.value.idle = status.database.idle || 0
    } else {
      database.value.connected = true
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
  } finally {
    loading.value = false
    database.value.loaded = true
  }
}

onMounted(loadSettings)

const save = async () => {
  saving.value = true
  try {
    await systemAPI.updateSettings({
      smtp: smtp.value,
      freeswitch: freeswitch.value
    })
    alert('Settings saved successfully!')
  } catch (e) {
    console.error('Failed to save settings:', e)
    alert('Failed to save settings: ' + (e.response?.data?.error || e.message || 'Unknown error'))
  } finally {
    saving.value = false
  }
}

const testSmtp = async () => {
  try {
    const result = await systemAPI.testSmtp(smtp.value)
    if (result.data?.success) {
      alert('Test email sent successfully!')
    } else {
      alert('Failed to send test email: ' + (result.data?.error || 'Unknown error'))
    }
  } catch (e) {
    console.error('SMTP test failed:', e)
    alert('Failed to send test email: ' + (e.response?.data?.error || e.message || 'Unknown error'))
  }
}
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

.status-badge.offline { background: #fef2f2; color: #dc2626; }

.deferred-badge { font-size: 10px; font-weight: 700; padding: 4px 10px; border-radius: 4px; background: #e0e7ff; color: #4338ca; text-transform: uppercase; }
.deferred-notice { display: flex; flex-direction: column; align-items: center; justify-content: center; text-align: center; padding: 60px 40px; }
.deferred-icon { width: 48px; height: 48px; opacity: 0.2; margin-bottom: 12px; }

/* ============================================
   RESPONSIVE STYLES - System Settings
   ============================================ */

/* Tablet (max-width: 1024px) */
@media (max-width: 1024px) {
  .settings-layout {
    gap: 16px;
  }
  
  .settings-nav {
    width: 180px;
  }
}

/* Mobile (max-width: 768px) */
@media (max-width: 768px) {
  .view-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .view-header .btn-primary {
    width: 100%;
    justify-content: center;
  }
  
  .settings-layout {
    flex-direction: column;
  }
  
  .settings-nav {
    width: 100%;
    flex-direction: row;
    overflow-x: auto;
    gap: 4px;
    padding: 4px;
    background: var(--bg-app);
    border-radius: var(--radius-md);
  }
  
  .nav-btn {
    white-space: nowrap;
    padding: 10px 14px;
    font-size: 12px;
  }
  
  .settings-content {
    padding: 16px;
  }
  
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .panel-header h3 {
    font-size: 16px;
  }
  
  .form-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }
  
  .setting-card {
    padding: 12px;
  }
  
  .setting-card h4 {
    font-size: 13px;
  }
  
  .setting-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .codec-item {
    padding: 8px 10px;
  }
  
  .codec-name {
    font-size: 12px;
  }
  
  .codec-type {
    display: none;
  }
  
  .connection-info {
    padding: 12px;
  }
  
  .info-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    padding: 10px 0;
  }
  
  .info-row .label {
    font-size: 11px;
  }
  
  .info-row .value {
    font-size: 13px;
  }
  
  .test-section {
    margin-top: 12px;
    padding-top: 12px;
  }
  
  .btn-secondary {
    width: 100%;
    justify-content: center;
  }
}

/* Small Mobile (max-width: 480px) */
@media (max-width: 480px) {
  .view-header h2 {
    font-size: 18px;
  }
  
  .header-content p {
    font-size: 12px;
  }
  
  .settings-nav {
    padding: 4px;
  }
  
  .nav-btn {
    padding: 8px 10px;
    font-size: 11px;
    gap: 6px;
  }
  
  .nav-icon {
    width: 14px;
    height: 14px;
  }
  
  .settings-content {
    padding: 12px;
    border-radius: var(--radius-sm);
  }
  
  .panel-desc {
    font-size: 12px;
    margin-bottom: 16px;
  }
  
  .precedence-badge,
  .status-badge,
  .deferred-badge {
    font-size: 9px;
    padding: 3px 8px;
  }
  
  .setting-card {
    padding: 10px;
    margin-bottom: 12px;
  }
  
  .form-group label {
    font-size: 10px;
  }
  
  .input-field {
    padding: 8px 10px;
    font-size: 13px;
  }
  
  .codec-list {
    gap: 4px;
  }
  
  .codec-item {
    padding: 8px;
    gap: 8px;
  }
  
  .grip {
    width: 14px;
    height: 14px;
  }
  
  .switch.sm {
    width: 32px;
    height: 18px;
  }
  
  .switch.sm .slider:before {
    width: 12px;
    height: 12px;
    left: 3px;
    bottom: 3px;
  }
  
  .switch.sm input:checked + .slider:before {
    transform: translateX(14px);
  }
  
  .deferred-notice {
    padding: 40px 20px;
  }

  .deferred-notice p {
    font-size: 12px;
  }
}
</style>
