<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Tenant Settings</h2>
      <p class="text-muted text-sm">Configure tenant environment, features, and integrations.</p>
    </div>
    <button class="btn-primary" @click="saveSettings">Save All Changes</button>
  </div>

  <div class="settings-layout">
    <div class="settings-nav">
      <div class="nav-group">
        <span class="nav-group-title">General</span>
        <div class="nav-item" :class="{ active: activeSection === 'general' }" @click="activeSection = 'general'">
          <SettingsIcon class="nav-icon" />
          <span>General</span>
        </div>

        <div class="nav-item" :class="{ active: activeSection === 'locations' }" @click="activeSection = 'locations'">
          <MapPinIcon class="nav-icon" />
          <span>Locations (E911)</span>
        </div>
      </div>

      <div class="nav-group">
        <span class="nav-group-title">Integrations</span>
        <div class="nav-item" :class="{ active: activeSection === 'messaging' }" @click="activeSection = 'messaging'">
          <MessageSquareIcon class="nav-icon" />
          <span>Messaging API</span>
        </div>
        <div class="nav-item" :class="{ active: activeSection === 'smtp' }" @click="activeSection = 'smtp'">
          <MailIcon class="nav-icon" />
          <span>SMTP / Email</span>
        </div>

        <div class="nav-item" :class="{ active: activeSection === 'hospitality' }" @click="activeSection = 'hospitality'">
          <HotelIcon class="nav-icon" />
          <span>Hospitality</span>
        </div>
      </div>

      <div class="nav-group">
        <span class="nav-group-title">Appearance</span>
        <div class="nav-item" :class="{ active: activeSection === 'whitelabel' }" @click="activeSection = 'whitelabel'">
          <PaletteIcon class="nav-icon" />
          <span>White Label</span>
        </div>
      </div>

      <div class="nav-group">
        <span class="nav-group-title">Security</span>
        <div class="nav-item" :class="{ active: activeSection === 'security' }" @click="activeSection = 'security'">
          <ShieldIcon class="nav-icon" />
          <span>SSL & Security</span>
        </div>
        <div class="nav-item" :class="{ active: activeSection === 'limits' }" @click="activeSection = 'limits'">
          <GaugeIcon class="nav-icon" />
          <span>Usage & Limits</span>
        </div>
      </div>
    </div>

    <div class="settings-content">
      
      <!-- GENERAL -->
      <div v-if="activeSection === 'general'" class="settings-panel">
        <div class="panel-header">
          <h3>General Configuration</h3>
        </div>
        
        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>SIP Domain</h4>
              <p>Used for device registration and internal routing.</p>
            </div>
            <input type="text" class="input-field" v-model="settings.sipDomain" placeholder="sip.tenant.com">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Portal Domain</h4>
              <p>Custom domain for user portal access.</p>
            </div>
            <input type="text" class="input-field" v-model="settings.portalDomain" placeholder="portal.tenant.com">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Timezone</h4>
              <p>Used for schedules and call logs.</p>
            </div>
            <select class="input-field" v-model="settings.timezone">
              <option value="America/Los_Angeles">America/Los_Angeles (PST)</option>
              <option value="America/New_York">America/New_York (EST)</option>
              <option value="America/Chicago">America/Chicago (CST)</option>
              <option value="UTC">UTC</option>
            </select>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Operator Extension</h4>
              <p>Extension for operator/receptionist.</p>
            </div>
            <input type="text" class="input-field code" v-model="settings.operatorExt" style="width: 80px">
          </div>
        </div>

        <div class="divider"></div>

        <div class="panel-header">
          <h3>Emergency & E911</h3>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Fallback Caller ID</h4>
              <p>Used when E911 location cannot be determined.</p>
            </div>
            <input type="text" class="input-field" v-model="settings.fallbackCallerId" placeholder="+14155559111">
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Enable Panic Button (999)</h4>
              <p>Allow 999 to trigger emergency alert system.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="settings.panicEnabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>
      </div>



      <!-- LOCATIONS -->
      <div v-if="activeSection === 'locations'" class="settings-panel">
        <LocationManager />
      </div>

      <!-- MESSAGING -->
      <div v-if="activeSection === 'messaging'" class="settings-panel">
        <div class="panel-header">
          <h3>Messaging API Configuration</h3>
        </div>
        <p class="panel-desc">Configure your upstream provider for SMS/MMS messaging.</p>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Provider</h4>
              <p>Choose your messaging gateway.</p>
            </div>
            <select class="input-field" v-model="messaging.provider">
              <option value="twilio">Twilio</option>
              <option value="bandwidth">Bandwidth</option>
              <option value="telnyx">Telnyx</option>
            </select>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Account SID / Username</h4>
              <p>Your provider account identifier.</p>
            </div>
            <input type="text" class="input-field" v-model="messaging.accountSid" placeholder="ACxxx...">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Auth Token / Secret</h4>
              <p>Your provider authentication secret.</p>
            </div>
            <input type="password" class="input-field" v-model="messaging.authToken" placeholder="••••••••">
          </div>
        </div>

        <div class="info-card">
          <div class="info-label">Webhook URL</div>
          <div class="info-value">
            <code>https://api.callsign.io/webhooks/sms/tenant_123</code>
            <button class="btn-small" @click="copyWebhook">Copy</button>
          </div>
          <p class="info-desc">Configure this URL in your provider's webhook settings.</p>
        </div>
      </div>

      <!-- SMTP -->
      <div v-if="activeSection === 'smtp'" class="settings-panel">
        <div class="panel-header">
          <h3>SMTP / Email Settings</h3>
        </div>
        <p class="panel-desc">Configure tenant-specific email settings. Leave disabled to use system defaults.</p>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Override System SMTP</h4>
              <p>Use custom SMTP settings instead of system defaults.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="smtpSettings.override">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div v-if="smtpSettings.override">
          <div class="setting-card">
            <div class="setting-row">
              <div class="setting-info">
                <h4>SMTP Host</h4>
                <p>Your outgoing mail server.</p>
              </div>
              <input type="text" class="input-field" v-model="smtpSettings.host" placeholder="smtp.example.com">
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-row">
              <div class="setting-info">
                <h4>Port</h4>
                <p>Usually 587 for TLS or 465 for SSL.</p>
              </div>
              <input type="text" class="input-field small" v-model="smtpSettings.port" placeholder="587">
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-row">
              <div class="setting-info">
                <h4>Username</h4>
                <p>SMTP authentication username.</p>
              </div>
              <input type="text" class="input-field" v-model="smtpSettings.username">
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-row">
              <div class="setting-info">
                <h4>Password</h4>
                <p>SMTP authentication password.</p>
              </div>
              <input type="password" class="input-field" v-model="smtpSettings.password">
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-row">
              <div class="setting-info">
                <h4>From Email</h4>
                <p>Sender email address for this tenant.</p>
              </div>
              <input type="email" class="input-field" v-model="smtpSettings.fromEmail" placeholder="voicemail@yourdomain.com">
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-row">
              <div class="setting-info">
                <h4>Encryption</h4>
                <p>Connection security method.</p>
              </div>
              <select class="input-field" v-model="smtpSettings.encryption">
                <option value="tls">TLS (Recommended)</option>
                <option value="ssl">SSL</option>
                <option value="none">None</option>
              </select>
            </div>
          </div>

          <button class="btn-secondary" @click="testSmtp">Send Test Email</button>
        </div>

        <div v-else class="info-card">
          <div class="info-label">Using System Defaults</div>
          <p class="info-desc">This tenant is using the global SMTP server configured in System Settings. Enable override above to use custom settings.</p>
        </div>
      </div>



      <!-- HOSPITALITY -->
      <div v-if="activeSection === 'hospitality'" class="settings-panel">
        <div class="panel-header">
          <h3>Hospitality Module</h3>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Enable Hospitality Features</h4>
              <p>Activate hotel/property management features.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="hospitality.enabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="info-card" v-if="hospitality.enabled">
          <p>PMS integration, room codes, and housekeeping status are managed in the dedicated Hospitality dashboard.</p>
          <button class="btn-secondary" @click="$router.push('/admin/hospitality')">
            <ExternalLinkIcon class="btn-icon-left" />
            Open Hospitality Dashboard
          </button>
        </div>
      </div>

      <!-- WHITE LABEL -->
      <div v-if="activeSection === 'whitelabel'" class="settings-panel">
        <div class="panel-header">
          <h3>Brand Customization</h3>
        </div>
        <p class="panel-desc">Customize the look and feel of the user portal.</p>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Brand Name</h4>
              <p>Displayed in portal header and emails.</p>
            </div>
            <input type="text" class="input-field" v-model="branding.name" placeholder="Company Name">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Logo URL</h4>
              <p>Company logo for portal and emails.</p>
            </div>
            <div class="input-group">
              <input type="text" class="input-field" v-model="branding.logoUrl" placeholder="https://...">
              <button class="btn-secondary small">Upload</button>
            </div>
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Primary Color</h4>
              <p>Accent color for buttons and links.</p>
            </div>
            <div class="color-picker">
              <input type="color" v-model="branding.primaryColor">
              <input type="text" class="input-field code" v-model="branding.primaryColor" style="width: 90px">
            </div>
          </div>
        </div>

        <div class="brand-preview">
          <span class="preview-label">Preview</span>
          <div class="preview-box" :style="{ '--preview-color': branding.primaryColor }">
            <div class="preview-header">
              <img v-if="branding.logoUrl" :src="branding.logoUrl" class="preview-logo">
              <span v-else class="preview-brand">{{ branding.name }}</span>
            </div>
            <button class="preview-btn">Sample Button</button>
          </div>
        </div>
      </div>

      <!-- SECURITY -->
      <div v-if="activeSection === 'security'" class="settings-panel">
        <div class="panel-header">
          <h3>SSL & Security</h3>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Enable SSL / HTTPS</h4>
              <p>Secure web access with TLS encryption.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="security.sslEnabled">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="setting-card toggle-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Force HTTPS Redirect</h4>
              <p>Redirect all HTTP traffic to HTTPS.</p>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="security.forceHttps">
              <span class="slider round"></span>
            </label>
          </div>
        </div>

        <div class="divider"></div>

        <div class="panel-header">
          <h3>Certificate</h3>
        </div>

        <div class="cert-card valid">
          <div class="cert-icon">🔒</div>
          <div class="cert-info">
            <span class="cert-issuer">Let's Encrypt Authority X3</span>
            <span class="cert-domain">*.callsign.io (Wildcard)</span>
            <span class="cert-expiry">Expires: Dec 12, 2025</span>
          </div>
          <div class="cert-actions">
            <button class="btn-secondary small">Renew</button>
            <button class="btn-secondary small">Replace</button>
          </div>
        </div>
      </div>

      <!-- LIMITS -->
      <div v-if="activeSection === 'limits'" class="settings-panel">
        <div class="panel-header">
          <h3>Usage & Quotas</h3>
        </div>
        <p class="panel-desc">View current usage against tenant limits (read-only).</p>

        <div class="usage-grid">
          <div class="usage-card">
            <div class="usage-header">
              <span class="usage-label">Extensions</span>
              <span class="usage-value">12 / 50</span>
            </div>
            <div class="usage-bar">
              <div class="usage-fill" style="width: 24%; background: #22c55e"></div>
            </div>
          </div>

          <div class="usage-card">
            <div class="usage-header">
              <span class="usage-label">Disk Storage</span>
              <span class="usage-value">4.2 GB / 10 GB</span>
            </div>
            <div class="usage-bar">
              <div class="usage-fill" style="width: 42%; background: #6366f1"></div>
            </div>
          </div>

          <div class="usage-card">
            <div class="usage-header">
              <span class="usage-label">Recordings</span>
              <span class="usage-value">1.2 GB / 5 GB</span>
            </div>
            <div class="usage-bar">
              <div class="usage-fill" style="width: 24%; background: #f59e0b"></div>
            </div>
          </div>

          <div class="usage-card">
            <div class="usage-header">
              <span class="usage-label">Fax Pages</span>
              <span class="usage-value">450 / 1000</span>
            </div>
            <div class="usage-bar">
              <div class="usage-fill" style="width: 45%; background: #6366f1"></div>
            </div>
          </div>
        </div>

        <div class="divider"></div>

        <div class="panel-header">
          <h3>Default User Limits</h3>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Voicemail Message Limit</h4>
              <p>Maximum messages per mailbox.</p>
            </div>
            <input type="number" class="input-field" v-model="limits.vmLimit" style="width: 100px">
          </div>
        </div>

        <div class="setting-card">
          <div class="setting-row">
            <div class="setting-info">
              <h4>Fax Retention</h4>
              <p>How long to keep fax documents.</p>
            </div>
            <select class="input-field" v-model="limits.faxRetention">
              <option value="30">30 Days</option>
              <option value="60">60 Days</option>
              <option value="90">90 Days</option>
              <option value="0">Forever</option>
            </select>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { 
  Settings as SettingsIcon, Hash as HashIcon, MapPin as MapPinIcon,
  MessageSquare as MessageSquareIcon, Server as ServerIcon, Building as HotelIcon,
  Palette as PaletteIcon, Shield as ShieldIcon, Gauge as GaugeIcon,
  ExternalLink as ExternalLinkIcon, Mail as MailIcon
} from 'lucide-vue-next'
import LocationManager from './LocationManager.vue'

const activeSection = ref('general')

const settings = ref({
  sipDomain: 'sip.acme.com',
  portalDomain: 'acme.callsign.io',
  timezone: 'America/Los_Angeles',
  operatorExt: '0',
  fallbackCallerId: '(415) 555-9111',
  panicEnabled: true
})

const smtpSettings = ref({
  override: false,
  host: '',
  port: '587',
  username: '',
  password: '',
  fromEmail: '',
  encryption: 'tls'
})

const testSmtp = () => alert('Test email sent!')

const featureCodes = ref([
  { key: 'vm', name: 'Voicemail Access', description: 'Check voicemail', value: '*97' },
  { key: 'park', name: 'Call Park', description: 'Park a call', value: '*700' },
  { key: 'pickup', name: 'Call Pickup', description: 'Pick up ringing call', value: '*8' },
  { key: 'block', name: 'Block Last', description: 'Block last caller', value: '*69' },
  { key: 'login', name: 'Agent Login', description: 'Log into queue', value: '*22' },
  { key: 'logout', name: 'Agent Logout', description: 'Log out of queue', value: '*23' },
  { key: 'fwdon', name: 'Forward Enable', description: 'Turn on forwarding', value: '*72' },
  { key: 'fwdoff', name: 'Forward Disable', description: 'Turn off forwarding', value: '*73' },
  { key: 'dndon', name: 'DND Enable', description: 'Turn on Do Not Disturb', value: '*78' },
  { key: 'dndoff', name: 'DND Disable', description: 'Turn off Do Not Disturb', value: '*79' },
])

const messaging = ref({
  provider: 'twilio',
  accountSid: 'ACsaf234...',
  authToken: ''
})

const provisioning = ref({
  protocol: 'https',
  auth: 'mac',
  secret: 'k9s8d7f6g5h4j3k2'
})

const hospitality = ref({ enabled: true })

const branding = ref({
  name: 'CallSign',
  logoUrl: '',
  primaryColor: '#6366f1'
})

const security = ref({
  sslEnabled: true,
  forceHttps: true
})

const limits = ref({
  vmLimit: 100,
  faxRetention: '30'
})

const saveSettings = () => alert('Settings saved!')
const resetCodes = () => alert('Feature codes reset to defaults')
const copyWebhook = () => alert('Copied to clipboard')
const copyProvUrl = () => alert('Copied to clipboard')
const regenerateSecret = () => {
  provisioning.value.secret = Math.random().toString(36).substring(2, 18)
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }

.settings-layout { display: flex; gap: 24px; align-items: flex-start; }

/* Navigation */
.settings-nav { width: 200px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 12px; position: sticky; top: 20px; }
.nav-group { margin-bottom: 16px; }
.nav-group:last-child { margin-bottom: 0; }
.nav-group-title { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); padding: 8px 12px 4px; display: block; }

.nav-item { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border-radius: var(--radius-sm); cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-main); }
.nav-item:hover { background: var(--bg-app); }
.nav-item.active { background: var(--primary-light); color: var(--primary-color); }
.nav-icon { width: 16px; height: 16px; opacity: 0.7; }
.nav-item.active .nav-icon { opacity: 1; }

/* Content */
.settings-content { flex: 1; max-width: 700px; }
.settings-panel { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 24px; }

.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.panel-header h3 { font-size: 16px; font-weight: 600; margin: 0; }
.panel-desc { font-size: 13px; color: var(--text-muted); margin-bottom: 20px; }

/* Setting Cards */
.setting-card { padding: 16px; background: var(--bg-app); border-radius: var(--radius-sm); margin-bottom: 12px; }
.setting-card.toggle-card { background: white; border: 1px solid var(--border-color); }

.setting-row { display: flex; justify-content: space-between; align-items: center; gap: 16px; }
.setting-info { flex: 1; }
.setting-info h4 { font-size: 14px; font-weight: 600; margin: 0 0 2px; }
.setting-info p { font-size: 12px; color: var(--text-muted); margin: 0; }

/* Codes Grid */
.codes-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.code-card { display: flex; justify-content: space-between; align-items: center; padding: 12px; background: var(--bg-app); border-radius: var(--radius-sm); }
.code-info { display: flex; flex-direction: column; }
.code-name { font-size: 13px; font-weight: 600; }
.code-desc { font-size: 11px; color: var(--text-muted); }

/* Info Cards */
.info-card { padding: 16px; background: #eff6ff; border: 1px solid #bfdbfe; border-radius: var(--radius-sm); margin-top: 16px; }
.info-label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: #1e40af; margin-bottom: 6px; }
.info-value { display: flex; align-items: center; gap: 8px; }
.info-value code { font-size: 12px; background: white; padding: 6px 10px; border-radius: 4px; flex: 1; }
.info-desc { font-size: 11px; color: #1e40af; margin-top: 8px; }

/* Usage Grid */
.usage-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.usage-card { padding: 16px; background: var(--bg-app); border-radius: var(--radius-sm); }
.usage-header { display: flex; justify-content: space-between; margin-bottom: 8px; }
.usage-label { font-size: 12px; color: var(--text-muted); }
.usage-value { font-size: 14px; font-weight: 600; }
.usage-bar { height: 6px; background: #e5e7eb; border-radius: 3px; overflow: hidden; }
.usage-fill { height: 100%; border-radius: 3px; }

/* Cert Card */
.cert-card { display: flex; align-items: center; gap: 16px; padding: 16px; background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: var(--radius-sm); }
.cert-icon { font-size: 24px; }
.cert-info { flex: 1; display: flex; flex-direction: column; }
.cert-issuer { font-size: 14px; font-weight: 600; color: #15803d; }
.cert-domain { font-size: 12px; color: #166534; }
.cert-expiry { font-size: 11px; color: var(--text-muted); }
.cert-actions { display: flex; gap: 8px; }

/* Brand Preview */
.brand-preview { margin-top: 20px; }
.preview-label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); display: block; margin-bottom: 8px; }
.preview-box { padding: 16px; background: #f8fafc; border: 1px dashed var(--border-color); border-radius: var(--radius-sm); }
.preview-header { margin-bottom: 12px; }
.preview-logo { height: 32px; }
.preview-brand { font-size: 18px; font-weight: 700; color: var(--preview-color, var(--primary-color)); }
.preview-btn { padding: 8px 16px; background: var(--preview-color, var(--primary-color)); color: white; border: none; border-radius: 4px; font-weight: 500; }

/* Color Picker */
.color-picker { display: flex; align-items: center; gap: 8px; }
.color-picker input[type="color"] { width: 40px; height: 40px; border: none; padding: 0; cursor: pointer; }

/* Form Elements */
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; min-width: 200px; }
.input-field.code { font-family: monospace; background: white; }
.input-field:focus { outline: none; border-color: var(--primary-color); }
.input-group { display: flex; gap: 8px; }
.divider { height: 1px; background: var(--border-color); margin: 24px 0; }

/* Buttons */
.btn-primary { background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); font-weight: 500; color: var(--text-main); cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-secondary.small { padding: 6px 10px; font-size: 12px; }
.btn-small { padding: 4px 8px; font-size: 11px; border: 1px solid var(--border-color); background: white; border-radius: 4px; cursor: pointer; }
.btn-link { background: none; border: none; color: var(--primary-color); cursor: pointer; font-size: 12px; font-weight: 500; }
.btn-icon-left { width: 14px; height: 14px; }
.mt-sm { margin-top: 8px; }

/* Switch */
.switch { position: relative; display: inline-block; width: 44px; height: 24px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #d1d5db; transition: .3s; }
.slider:before { position: absolute; content: ""; height: 18px; width: 18px; left: 3px; bottom: 3px; background-color: white; transition: .3s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(20px); }
.slider.round { border-radius: 24px; }
.slider.round:before { border-radius: 50%; }
</style>
