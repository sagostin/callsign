<template>
  <div class="debugger-container">
    <div class="debug-form">
      <h3>Configuration Inspector</h3>
      <p class="text-muted text-sm mb-4">Simulate FreeSWITCH XML CURL requests to view generated configuration.</p>

      <div class="form-grid">
        <div class="form-group full-width" v-if="mode === 'system'">
          <label>Section</label>
          <select v-model="form.section" class="input-field">
            <option value="dialplan">Dialplan (Routing)</option>
            <option value="directory">Directory (Users/Auth)</option>
            <option value="configuration">Configuration (Modules)</option>
          </select>
        </div>

        <!-- Common Fields -->
        <div class="form-group" v-if="mode === 'system'">
          <label>Domain</label>
          <input v-model="form.domain" class="input-field" placeholder="customer.domain.com" />
        </div>

        <!-- Dialplan Specific -->
        <template v-if="form.section === 'dialplan'">
          <div class="form-group">
            <label>Context</label>
            <select v-model="form.context" class="input-field">
              <option value="public">Public (Inbound)</option>
              <option value="default">Default (Internal/Outbound)</option>
              <option v-if="form.domain" :value="form.domain">Domain Specific ({{ form.domain }})</option>
            </select>
          </div>
          <div class="form-group">
            <label>Destination Number</label>
            <div class="input-group">
              <input v-model="form.destination_number" class="input-field" placeholder="e.g. 1001" @keyup.enter="runDebug" />
            </div>
          </div>
        </template>

        <!-- Directory Specific -->
        <template v-if="form.section === 'directory'">
           <div class="form-group">
             <label>User (Extension)</label>
             <input v-model="form.user" class="input-field" placeholder="e.g. 1001" @keyup.enter="runDebug" />
           </div>
           <div class="form-group">
             <label>Action</label>
             <select v-model="form.action" class="input-field">
               <option value="sip_auth">SIP Auth</option>
               <option value="message-count">Message Count</option>
               <option value="user_call">User Call</option>
             </select>
           </div>
        </template>

        <!-- Configuration Specific -->
        <template v-if="form.section === 'configuration'">
           <div class="form-group full-width">
             <label>Config Name</label>
             <select v-model="form.config_name" class="input-field">
               <option value="sofia.conf">Sofia (SIP)</option>
               <option value="acl.conf">ACL (Access Control)</option>
               <option value="post_load_modules.conf">Modules</option>
               <option value="event_socket.conf">Event Socket</option>
             </select>
           </div>
        </template>

        <div class="form-group full-width">
          <button class="btn-primary full-width" @click="runDebug" :disabled="loading">
             {{ loading ? 'Generating XML...' : 'Inspect Configuration' }}
          </button>
        </div>
      </div>
    </div>

    <div class="results-panel" v-if="result">
       <div class="panel-header">
         <h4>Generated XML</h4>
         <button class="btn-icon" @click="copyXML" title="Copy XML">
           <CopyIcon class="icon-sm" />
         </button>
       </div>
       <div class="xml-viewer">
         <pre><code>{{ result }}</code></pre>
       </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import { Copy as CopyIcon } from 'lucide-vue-next'
import api from '@/services/api' 

const props = defineProps({
  mode: {
    type: String,
    default: 'tenant' // 'tenant' or 'system'
  }
})

const toast = inject('toast')
const loading = ref(false)
const result = ref('')

const form = ref({
  section: 'dialplan', // Default
  domain: '',
  context: 'public', 
  destination_number: '',
  user: '',
  action: 'sip_auth',
  config_name: 'sofia.conf'
})

// If tenant mode, FORCE dialplan section
if (props.mode === 'tenant') {
  form.value.section = 'dialplan'
}

async function runDebug() {
  loading.value = true
  result.value = ''
  
  try {
    let url = ''
    let params = { ...form.value }

    if (props.mode === 'system') {
      url = '/system/xml/debug'
    } else {
      url = '/tenant/routing/debug'
      // Tenant endpoint only supports dialplan params, handled by backend
    }

    const res = await api.get(url, { params })
    result.value = formatXml(res.data.xml || '<no-result/>')
    
  } catch (e) {
    toast.error(e.response?.data?.error || 'Failed to debug config')
  } finally {
    loading.value = false
  }
}

function formatXml(xml) {
  return xml.replace(/>\s*</g, '>\n<')
}

function copyXML() {
  navigator.clipboard.writeText(result.value)
  toast.success('XML Copied to clipboard')
}
</script>

<style scoped>
.debugger-container { display: flex; flex-direction: column; gap: 24px; }
.debug-form { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px; align-items: flex-end; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full-width { grid-column: 1 / -1; } /* Span all columns */
.results-panel { background: #1e1e1e; color: #d4d4d4; border-radius: var(--radius-md); overflow: hidden; border: 1px solid var(--border-color); }
.panel-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: #252526; border-bottom: 1px solid #333; }
.panel-header h4 { margin: 0; font-size: 13px; font-weight: 600; color: #fff; text-transform: uppercase; }
.xml-viewer { padding: 16px; overflow-x: auto; font-family: 'Fira Code', 'Roboto Mono', monospace; font-size: 13px; line-height: 1.5; }
pre { margin: 0; }
.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; white-space: nowrap; }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.btn-icon { background: none; border: none; cursor: pointer; color: #aaa; padding: 4px; }
.btn-icon:hover { color: white; }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); width: 100%; }
.text-muted { color: var(--text-muted); }
.text-sm { font-size: 13px; }
.mb-4 { margin-bottom: 16px; }
.icon-sm { width: 16px; height: 16px; }
.full-width { width: 100%; }
</style>
