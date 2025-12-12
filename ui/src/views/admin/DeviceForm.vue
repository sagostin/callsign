<template>
  <div class="view-header">
    <div class="header-content">
       <button class="back-link" @click="$router.push('/admin/devices')">← Back to Devices</button>
      <h2>{{ isNew ? 'New Device' : 'Edit Device' }}</h2>
      <p class="text-muted text-sm" v-if="!isNew">MAC: {{ device.mac }}</p>
    </div>
  </div>

  <div class="form-grid">
    <div class="form-group">
      <label>MAC Address</label>
      <input type="text" v-model="device.mac" class="input-field code" placeholder="00:15:65:..." :disabled="!isNew">
      <span class="help-text" v-if="device.model === 'Generic SIP'">Optional for Generic SIP devices</span>
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
         </optgroup>
         <optgroup label="Polycom">
           <option value="Polycom VVX 450">Polycom VVX 450</option>
         </optgroup>
         <optgroup label="Grandstream">
           <option value="Grandstream GXP2170">Grandstream GXP2170</option>
         </optgroup>
       </select>
    </div>

    <div class="form-group">
       <label>Location</label>
       <select v-model="device.location" class="input-field">
          <option value="default">Default Location</option>
          <option value="hq">HQ - San Francisco</option>
          <option value="warehouse">Warehouse</option>
       </select>
    </div>

    <div class="form-group">
       <label>Assigned Extension</label>
       <select v-model="device.extension" class="input-field">
          <option value="">Unassigned</option>
          <option value="101">101 - Alice Smith</option>
          <option value="102">102 - Bob Jones</option>
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
                <input v-model="device.sipUsername" class="input-field code" placeholder="e.g. 101 or user@domain.com">
                <span class="help-text">The username used for SIP REGISTER requests</span>
             </div>
             <div class="form-group flex-1">
                <label>Auth Username</label>
                <input v-model="device.sipAuthUser" class="input-field code" placeholder="Same as above if empty">
                <span class="help-text">Optional, if different from registration username</span>
             </div>
           </div>

           <div class="form-row">
             <div class="form-group flex-1">
                <label>SIP Password</label>
                <div class="password-field">
                  <input :type="showPassword ? 'text' : 'password'" v-model="device.sipPassword" class="input-field" placeholder="Enter SIP password">
                  <button type="button" class="toggle-password" @click="showPassword = !showPassword">
                    <EyeIcon v-if="!showPassword" class="icon-sm" />
                    <EyeOffIcon v-else class="icon-sm" />
                  </button>
                </div>
             </div>
             <div class="form-group flex-1">
                <label>Display Name (Caller ID)</label>
                <input v-model="device.displayName" class="input-field" placeholder="e.g. John Smith">
             </div>
           </div>

           <div class="divider-sm"></div>

           <div class="form-row">
             <div class="form-group flex-1">
                <label>SIP Domain / Realm</label>
                <input v-model="device.sipDomain" class="input-field" placeholder="e.g. sip.company.com">
             </div>
             <div class="form-group flex-1">
                <label>Outbound Proxy</label>
                <input v-model="device.sipProxy" class="input-field" placeholder="e.g. proxy.company.com (optional)">
             </div>
           </div>

           <div class="form-row">
             <div class="form-group flex-1">
                <label>SIP Profile</label>
                <select v-model="device.sipProfile" class="input-field">
                   <option value="internal">internal (5060)</option>
                   <option value="external">external (5080)</option>
                   <option value="custom">Custom...</option>
                </select>
             </div>
             <div class="form-group flex-1">
                <label>Transport</label>
                <select v-model="device.transport" class="input-field">
                   <option value="udp">UDP</option>
                   <option value="tcp">TCP</option>
                   <option value="tls">TLS (Encrypted)</option>
                </select>
             </div>
             <div class="form-group flex-1">
                <label>Port</label>
                <input type="number" v-model="device.sipPort" class="input-field" placeholder="5060">
             </div>
           </div>

           <div class="divider-sm"></div>

           <div class="advanced-settings">
             <button class="toggle-advanced" @click="showAdvanced = !showAdvanced">
               <ChevronRightIcon :class="{ 'rotated': showAdvanced }" class="icon-sm" />
               Advanced Options
             </button>
             
             <div v-if="showAdvanced" class="advanced-fields">
               <div class="form-row">
                 <div class="form-group flex-1">
                    <label>Expiry (seconds)</label>
                    <input type="number" v-model="device.sipExpiry" class="input-field" placeholder="3600">
                 </div>
                 <div class="form-group flex-1">
                    <label>Codec Preference</label>
                    <input v-model="device.codecs" class="input-field code" placeholder="PCMU,PCMA,G722,opus">
                 </div>
               </div>
               
               <div class="form-row">
                 <label class="checkbox-row">
                   <input type="checkbox" v-model="device.natTraversal">
                   <span>Enable NAT Traversal (STUN/ICE)</span>
                 </label>
                 <label class="checkbox-row">
                   <input type="checkbox" v-model="device.allowRefer">
                   <span>Allow SIP REFER (Transfers)</span>
                 </label>
               </div>
             </div>
           </div>
         </div>
       </div>
    </div>

    <!-- Standard SIP Profile Selector (non-Generic devices) -->
    <div v-else class="form-group full-width">
       <div class="divider"></div>
       <label>Network / SIP Settings</label>
       <div class="form-row">
          <div class="form-group flex-1">
             <label>SIP Profile</label>
             <select v-model="device.sipProfile" class="input-field">
                <option value="internal">internal (5060)</option>
                <option value="external">external (5080)</option>
                <option value="custom">Custom Profile...</option>
             </select>
             <span class="help-text">Socket to register against.</span>
          </div>
          <div class="form-group flex-1">
             <label>Transport</label>
             <select v-model="device.transport" class="input-field">
                <option value="udp">UDP</option>
                <option value="tcp">TCP</option>
                <option value="tls">TLS</option>
             </select>
          </div>
       </div>
    </div>

    <div class="form-actions full-width">
       <button class="btn-danger-outline" v-if="!isNew" @click="deleteDevice">Delete Device</button>
       <div style="flex:1"></div>
       <button class="btn-secondary" @click="$router.push('/admin/devices')">Cancel</button>
       <button class="btn-primary" @click="saveDevice">Save Device</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Server as ServerIcon, Eye as EyeIcon, EyeOff as EyeOffIcon, ChevronRight as ChevronRightIcon } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.params.id === 'new')
const showPassword = ref(false)
const showAdvanced = ref(false)

const device = ref({
  mac: '',
  model: '',
  location: 'default',
  extension: '',
  sipProfile: 'internal',
  transport: 'udp',
  // Generic SIP fields
  sipUsername: '',
  sipAuthUser: '',
  sipPassword: '',
  displayName: '',
  sipDomain: '',
  sipProxy: '',
  sipPort: 5060,
  sipExpiry: 3600,
  codecs: 'PCMU,PCMA,G722',
  natTraversal: true,
  allowRefer: true
})

if (!isNew.value) {
  // Mock load
  device.value = { 
    mac: '00:15:65:12:34:56', 
    model: 'Yealink T54W', 
    location: 'hq', 
    extension: '101',
    sipProfile: 'internal',
    transport: 'udp',
    sipUsername: '',
    sipAuthUser: '',
    sipPassword: '',
    displayName: '',
    sipDomain: '',
    sipProxy: '',
    sipPort: 5060,
    sipExpiry: 3600,
    codecs: 'PCMU,PCMA,G722',
    natTraversal: true,
    allowRefer: true
  }
}

const saveDevice = () => {
   alert('Device Saved')
   router.push('/admin/devices')
}

const deleteDevice = () => {
   if(confirm('Delete this device permanently?')) {
      alert('Device deleted')
      router.push('/admin/devices')
   }
}
</script>

<style scoped>
.view-header {
  margin-bottom: var(--spacing-lg);
}
.back-link {
  background: none; border: none; color: var(--text-muted); padding: 0; font-size: 11px; cursor: pointer; margin-bottom: 8px;
}
.back-link:hover { color: var(--primary-color); text-decoration: underline; }

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-lg);
  max-width: 800px;
  background: white;
  padding: var(--spacing-xl);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}

.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full-width { grid-column: span 2; }
.form-row { display: flex; gap: 16px; }
.flex-1 { flex: 1; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }

.input-field {
  padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px;
}
.input-field.code { font-family: monospace; }
.divider { height: 1px; background: var(--border-color); margin-bottom: 12px; }
.divider-sm { height: 1px; background: var(--border-color); margin: 16px 0; }
.help-text { font-size: 11px; color: var(--text-muted); }

.form-actions { display: flex; gap: 12px; margin-top: 12px; }

.btn-primary {
  background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer;
}
.btn-secondary {
  background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer;
}
.btn-danger-outline {
  background: white; border: 1px solid var(--status-bad); color: var(--status-bad); padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer;
}

/* Generic SIP Section */
.generic-sip-section {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #bae6fd;
  border-radius: var(--radius-md);
  padding: 20px;
  margin-top: 8px;
}

.section-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 20px;
}

.section-icon {
  width: 24px;
  height: 24px;
  color: #0369a1;
  flex-shrink: 0;
  margin-top: 2px;
}

.section-header h3 {
  font-size: 15px;
  font-weight: 600;
  color: #0c4a6e;
  margin: 0 0 4px 0;
}

.sip-fields {
  background: white;
  border: 1px solid #e0f2fe;
  border-radius: var(--radius-sm);
  padding: 16px;
}

.sip-fields .form-group {
  margin-bottom: 12px;
}

.password-field {
  position: relative;
  display: flex;
}
.password-field .input-field {
  flex: 1;
  padding-right: 40px;
}
.toggle-password {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
}
.toggle-password:hover { color: var(--text-main); }

.advanced-settings {
  margin-top: 8px;
}

.toggle-advanced {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  padding: 4px 0;
}
.toggle-advanced .icon-sm {
  transition: transform 0.2s;
}
.toggle-advanced .rotated {
  transform: rotate(90deg);
}

.advanced-fields {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed var(--border-color);
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: normal;
  text-transform: none;
  color: var(--text-main);
  cursor: pointer;
}
.checkbox-row input {
  width: 16px;
  height: 16px;
}

.icon-sm { width: 16px; height: 16px; }
</style>
