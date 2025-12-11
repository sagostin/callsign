<template>
  <div class="view-header">
     <div class="header-content">
       <button class="back-link" @click="$router.push('/system/tenants')">← Back to Tenants</button>
       <h2>{{ beginsWithNew ? 'Create New Tenant' : 'Edit Tenant' }}</h2>
     </div>
  </div>

  <div class="settings-container">
    <div class="settings-nav">
      <div class="nav-item" :class="{ active: activeTab === 'general' }" @click="activeTab = 'general'">
        General
      </div>
      <div class="nav-item" :class="{ active: activeTab === 'users' }" @click="activeTab = 'users'">
        Users
      </div>
      <div class="nav-item" :class="{ active: activeTab === 'devices' }" @click="activeTab = 'devices'">
        Devices
      </div>
    </div>

    <div class="settings-content">
       <!-- GENERAL TAB -->
       <div v-if="activeTab === 'general'">
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
               <input v-model="form.email" class="input-field" placeholder="admin@company.com">
            </div>
         </div>

         <div class="divider"></div>
         <h3 class="card-title">System Configuration</h3>
         
         <div class="form-group">
            <label>Service Plan / Package</label>
            <select v-model="form.plan" class="input-field">
               <option value="basic">Basic (SMB)</option>
               <option value="hotel">Hospitality</option>
               <option value="enterprise">Enterprise</option>
            </select>
         </div>

         <div class="form-group">
            <label>Default Dial Plan</label>
            <select v-model="form.dialplan" class="input-field">
               <option value="us-std">US Standard</option>
               <option value="uk-std">UK Standard</option>
            </select>
         </div>
         
         <div class="form-group">
            <label>Media Server (Cluster)</label>
            <select v-model="form.media" class="input-field">
               <option value="us-east">US East (Virginia)</option>
               <option value="us-west">US West (Oregon)</option>
            </select>
         </div>
       </div>

       <!-- PROFILES TAB -->
       <div v-if="activeTab === 'profiles'">
          <h3 class="card-title">Allocations & Limits</h3>
          <div class="grid grid-cols-2 gap-4 mb-6">
             <div class="form-group">
                <label>Max Extensions</label>
                <input v-model="form.maxExt" type="number" class="input-field">
             </div>
             <div class="form-group">
                <label>Max Disk Space (GB)</label>
                <input v-model="form.maxDisk" type="number" class="input-field">
             </div>
             <div class="form-group">
                <label>Concurrent Calls (Channels)</label>
                <input v-model="form.maxChannels" type="number" class="input-field">
             </div>
          </div>

          <div class="divider"></div>
          <h3 class="card-title">Main Tenant Overrides</h3>
          <p class="text-xs text-muted mb-4">Enable or disable specific features for this tenant.</p>

          <div class="form-group bg-panel">
             <div class="toggle-row">
                <div class="toggle-info">
                   <span class="label">Enable Hospitality Features</span>
                   <span class="desc">Modules for Hotel / PMS integration.</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.overrides.hospitality">
                  <span class="slider round"></span>
                </label>
             </div>
          </div>

          <div class="form-group bg-panel">
             <div class="toggle-row">
                <div class="toggle-info">
                   <span class="label">Enable Panic Button (E911)</span>
                   <span class="desc">Allow emergency panic button usage for this tenant.</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.overrides.panic">
                  <span class="slider round"></span>
                </label>
             </div>
          </div>

          <div class="form-group bg-panel">
             <div class="toggle-row">
                <div class="toggle-info">
                   <span class="label">Enable SSL / HTTPS</span>
                   <span class="desc">Provision SSL certificates for tenant domains.</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.overrides.ssl">
                  <span class="slider round"></span>
                </label>
             </div>
          </div>

          <div class="form-group bg-panel">
             <div class="toggle-row">
                <div class="toggle-info">
                   <span class="label">Force HTTPS Redirect</span>
                   <span class="desc">Redirect all HTTP traffic to HTTPS.</span>
                </div>
                <label class="switch">
                  <input type="checkbox" v-model="form.overrides.forceHttps">
                  <span class="slider round"></span>
                </label>
             </div>
          </div>
       </div>

       <!-- USERS TAB -->
       <div v-if="activeTab === 'users'">
          <div class="flex justify-between items-center mb-4 border-b border-slate-100 pb-4">
             <h3 class="card-title mb-0 border-none p-0">Tenant Users</h3>
             <button class="btn-primary small text-xs">+ Add User</button>
          </div>
          
          <table class="w-full text-left text-sm">
             <thead>
                <tr class="text-xs text-slate-500 uppercase border-b">
                   <th class="pb-2">Ext</th>
                   <th class="pb-2">Name</th>
                   <th class="pb-2">Email</th>
                   <th class="pb-2 text-right">Action</th>
                </tr>
             </thead>
             <tbody>
                <tr class="border-b border-slate-100">
                   <td class="py-3 font-mono">1001</td>
                   <td class="py-3 font-medium">Alice Admin</td>
                   <td class="py-3 text-slate-500">alice@tenant.com</td>
                   <td class="py-3 text-right text-indigo-600 cursor-pointer">Edit</td>
                </tr>
                <tr class="border-b border-slate-100">
                   <td class="py-3 font-mono">1002</td>
                   <td class="py-3 font-medium">Bob Staff</td>
                   <td class="py-3 text-slate-500">bob@tenant.com</td>
                   <td class="py-3 text-right text-indigo-600 cursor-pointer">Edit</td>
                </tr>
             </tbody>
          </table>
       </div>

       <!-- DEVICES TAB -->
       <div v-if="activeTab === 'devices'">
          <div class="flex justify-between items-center mb-4 border-b border-slate-100 pb-4">
             <h3 class="card-title mb-0 border-none p-0">Provisioned Devices</h3>
             <button class="btn-primary small text-xs">+ Add Device</button>
          </div>
          
          <table class="w-full text-left text-sm">
             <thead>
                <tr class="text-xs text-slate-500 uppercase border-b">
                   <th class="pb-2">MAC Address</th>
                   <th class="pb-2">Model</th>
                   <th class="pb-2">Assigned To</th>
                   <th class="pb-2">Status</th>
                </tr>
             </thead>
             <tbody>
                <tr class="border-b border-slate-100">
                   <td class="py-3 font-mono text-xs">80:5E:C0:D1:F2:A1</td>
                   <td class="py-3">Yealink T54W</td>
                   <td class="py-3">1001 - Alice</td>
                   <td class="py-3"><span class="px-2 py-0.5 rounded text-[10px] bg-green-100 text-green-700 font-bold uppercase">Online</span></td>
                </tr>
                <tr class="border-b border-slate-100">
                   <td class="py-3 font-mono text-xs">00:04:F2:12:34:56</td>
                   <td class="py-3">Poly CCX 500</td>
                   <td class="py-3">1002 - Bob</td>
                   <td class="py-3"><span class="px-2 py-0.5 rounded text-[10px] bg-slate-100 text-slate-500 font-bold uppercase">Offline</span></td>
                </tr>
             </tbody>
          </table>
       </div>
    </div>
  </div>

  <div class="form-actions mt-6">
     <button class="btn-secondary" @click="$router.push('/system/tenants')">Cancel</button>
     <button class="btn-primary" @click="save">Save Tenant</button>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const beginsWithNew = computed(() => !route.params.id)
const activeTab = ref('general')

const form = ref({
   name: '',
   domain: '',
   email: '',
   maxExt: 10,
   maxDisk: 5,
   maxChannels: 5,
   plan: 'basic',
   dialplan: 'us-std',
   media: 'us-west',
   overrides: {
     hospitality: false,
     panic: true,
     ssl: true,
     forceHttps: false
   }
})

const save = () => {
   alert('Tenant Saved')
   router.push('/system/tenants')
}
</script>

<style scoped>
.view-header { margin-bottom: 24px; }
.back-link { background: none; border: none; color: var(--text-muted); padding: 0; font-size: 11px; cursor: pointer; }
.back-link:hover { color: var(--primary-color); text-decoration: underline; }

.settings-container { display: flex; gap: var(--spacing-xl); align-items: flex-start; }
.settings-nav { width: 180px; display: flex; flex-direction: column; gap: 4px; }
.nav-item { padding: 10px 16px; border-radius: var(--radius-sm); font-size: var(--text-sm); color: var(--text-main); cursor: pointer; font-weight: 500; }
.nav-item:hover { background: var(--bg-app); }
.nav-item.active { background: var(--primary-light); color: var(--primary-color); font-weight: 600; }

.settings-content { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 24px; max-width: 800px; }

.card-title { font-size: 14px; font-weight: 700; color: var(--text-primary); margin-bottom: 16px; border-bottom: 1px solid var(--border-color); padding-bottom: 8px; }

.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }

.divider { height: 1px; background: var(--border-color); margin: 24px 0; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }

.form-actions { display: flex; justify-content: flex-end; gap: 12px; }

/* In-View Toggles */
.bg-panel { background: #f8fafc; padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border-color); margin-bottom: 12px; }
.toggle-row { display: flex; justify-content: space-between; align-items: center; }
.toggle-info { display: flex; flex-direction: column; }
.toggle-info .label { font-weight: 600; font-size: 13px; color: var(--text-main); text-transform: none; }
.toggle-info .desc { font-size: 11px; color: var(--text-muted); }

.switch { position: relative; display: inline-block; width: 34px; height: 20px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .4s; }
.slider:before { position: absolute; content: ""; height: 16px; width: 16px; left: 2px; bottom: 2px; background-color: white; transition: .4s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(14px); }
.slider.round { border-radius: 34px; }
.slider.round:before { border-radius: 50%; }
</style>
