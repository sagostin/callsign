<template>
  <div class="view-header">
    <div class="header-left">
      <button class="back-link" @click="$router.push('/admin/devices/templates')">← Back to Templates</button>
      <h2>{{ isNew ? 'New Master Template' : 'Edit Template: Yealink T54W' }}</h2>
    </div>
    <div class="header-actions">
      <button class="btn-primary">Save Changes</button>
    </div>
  </div>

  <div class="template-editor">
    <div class="editor-sidebar">
      <div class="section-title">Configuration</div>
      
      <div class="form-group">
        <label>Template Name</label>
        <input type="text" class="input-field" value="Standard Yealink T54W">
      </div>

      <div class="form-group">
         <label>Scope</label>
         <div class="radio-group">
           <label class="radio-label">
             <input type="radio" name="scope" checked> Global (Master)
           </label>
           <label class="radio-label">
             <input type="radio" name="scope"> Tenant Specific
           </label>
         </div>
      </div>

       <div class="form-group">
        <label>Firmware Version</label>
        <input type="text" class="input-field" value="96.85.0.5">
      </div>
      
      <div class="divider"></div>
      
      <div class="toggle-row">
        <span>Raw Config Mode</span>
        <label class="switch">
          <input type="checkbox" v-model="rawMode">
          <span class="slider round"></span>
        </label>
      </div>

    </div>

    <div class="editor-main">
      <div class="code-editor" v-if="rawMode">
        <div class="editor-header">yealink_common.cfg</div>
        <textarea class="code-area" spellcheck="false">
#!version:1.0.0.1
#Enable or disable the phone to save the local call log; 0-Disabled, 1-Enabled (default);
features.save_call_log = 1

# [INHERITED] Keypad lock settings inherited from Master
# details: features.keypad_lock = 0

#Configure the return code when the refuse a call; 404, 480, 486 (default);
features.call_refuse_code = 486

auto_provision.pnp_enable = 1
auto_provision.custom.protect = 1

account.1.enable = 1
account.1.label = %NULL%
account.1.display_name = %NULL%
account.1.auth_name = %NULL%
account.1.password = %NULL%
        </textarea>
      </div>

      <div class="visual-editor" v-else>
         <div class="visual-section">
            <h3>Programmable Keys (BLF / Line)</h3>
            <p class="text-muted mb-md">Click a key to configure its function.</p>
            
            <div class="phone-faceplate">
               <div class="screen-area">
                  <div class="screen-content">
                     <div class="status-bar">12:30 PM</div>
                     <div class="idle-text">CallSign</div>
                  </div>
               </div>
               
               <div class="keys-grid left">
                  <button 
                    v-for="i in 6" :key="'l'+i" 
                    class="line-key" 
                    :class="{ configured: keys['l'+i] }"
                    @click="editKey('l'+i)"
                  >
                    <div class="key-led" :class="{ 'on': keys['l'+i] }"></div>
                    <span class="key-label">{{ getKeyLabel('l'+i) }}</span>
                  </button>
               </div>
               
               <div class="keys-grid right">
                  <button 
                    v-for="i in 6" :key="'r'+i" 
                    class="line-key" 
                    :class="{ configured: keys['r'+i] }"
                    @click="editKey('r'+i)"
                  >
                    <span class="key-label">{{ getKeyLabel('r'+i) }}</span>
                    <div class="key-led" :class="{ 'on': keys['r'+i] }"></div>
                  </button>
               </div>
            </div>
         </div>

         <div class="config-panel" v-if="selectedKey">
            <h4>Configure Key: {{ selectedKey }}</h4>
            <div class="form-group">
               <label>Type</label>
               <select v-model="keys[selectedKey].type" class="input-field">
                  <option value="line">Line</option>
                  <option value="blf">BLF (Busy Lamp Field)</option>
                  <option value="speed_dial">Speed Dial</option>
                  <option value="park">Call Park</option>
               </select>
            </div>
            <div class="form-group">
               <label>Label</label>
               <input type="text" v-model="keys[selectedKey].label" class="input-field">
            </div>
            <div class="form-group" v-if="keys[selectedKey].type !== 'line'">
               <label>Value / Extension</label>
               <input type="text" v-model="keys[selectedKey].value" class="input-field">
            </div>
            <button class="btn-secondary small" @click="selectedKey = null">Done</button>
         </div>

         <div class="visual-section mt-lg">
            <h3>Common Settings</h3>
            <div class="form-grid">
               <div class="form-group">
                 <label>Screen Saver Wait Time (min)</label>
                 <input type="number" class="input-field" value="5">
               </div>
               <div class="form-group">
                 <label>Backlight Active Level</label>
                 <input type="range" class="range-input" min="1" max="10" value="8">
               </div>
               <div class="form-group">
                 <label>Time Format</label>
                 <select class="input-field">
                    <option>12 Hour</option>
                    <option>24 Hour</option>
                 </select>
               </div>
            </div>
         </div>
         <div class="visual-section mt-lg">
            <h3>Advanced Parameters</h3>
            <p class="text-muted mb-md">Add custom provisioning parameters not covered above.</p>
            <div class="custom-params">
               <div class="param-row header">
                  <span class="col-key">Parameter</span>
                  <span class="col-val">Value</span>
                  <span class="col-act"></span>
               </div>
               <div class="param-row" v-for="(val, key) in customParams" :key="key">
                  <input class="input-field small" :value="key" readonly>
                  <input class="input-field small" :value="val" @input="updateParam(key, $event.target.value)">
                  <button class="btn-icon text-bad" @click="removeParam(key)">×</button>
               </div>
               <div class="param-row new">
                  <input class="input-field small" placeholder="e.g. features.bluetooth_enable" v-model="newParamKey">
                  <input class="input-field small" placeholder="1" v-model="newParamVal">
                  <button class="btn-secondary small" @click="addParam">Add</button>
               </div>
            </div>
         </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const isNew = computed(() => route.params.id === 'new')

const rawMode = ref(false) // Default to Visual
const scope = ref('global')

const selectedKey = ref(null)
const keys = ref({
   'l1': { type: 'line', label: 'Line 1', value: '1' },
   'l2': { type: 'line', label: 'Line 2', value: '2' },
   'r1': { type: 'blf', label: 'Bob J.', value: '1002' },
})

const editKey = (keyId) => {
   if (!keys.value[keyId]) {
      keys.value[keyId] = { type: 'speed_dial', label: 'New Key', value: '' }
   }
   selectedKey.value = keyId
}

const getKeyLabel = (keyId) => {
   return keys.value[keyId]?.label || 'Empty'
}

const customParams = ref({
   'features.bluetooth_enable': '1',
   'network.pc_port.enable': '0'
})
const newParamKey = ref('')
const newParamVal = ref('')

const addParam = () => {
   if (newParamKey.value) {
      customParams.value[newParamKey.value] = newParamVal.value
      newParamKey.value = ''
      newParamVal.value = ''
   }
}
const removeParam = (key) => delete customParams.value[key]
const updateParam = (key, val) => customParams.value[key] = val
</script>

<style scoped>
.header-left {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.back-link {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  font-size: var(--text-xs);
  text-align: left;
}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.view-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.template-editor {
  display: grid;
  grid-template-columns: 300px 1fr;
  gap: var(--spacing-lg);
  height: calc(100vh - 200px);
}

.editor-sidebar {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.editor-main {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.section-title { font-weight: 700; color: var(--text-primary); font-size: var(--text-sm); }

.form-group { display: flex; flex-direction: column; gap: 6px; }

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
}

.divider { height: 1px; background: var(--border-color); }

.toggle-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--text-sm);
  font-weight: 600;
}

/* Switch */
.switch { position: relative; display: inline-block; width: 40px; height: 24px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0;
  background-color: var(--border-color); transition: .4s;
}
.slider:before {
  position: absolute; content: ""; height: 16px; width: 16px; left: 4px; bottom: 4px;
  background-color: white; transition: .4s;
}
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(16px); }
.slider.round { border-radius: 34px; }
.slider.round:before { border-radius: 50%; }

/* Code Editor */
.code-editor { flex: 1; display: flex; flex-direction: column; }
.editor-header {
  background: var(--bg-app);
  padding: 8px 16px;
  font-family: monospace;
  font-size: 12px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-muted);
}
.code-area {
  flex: 1;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Fira Code', monospace;
  font-size: 13px;
  line-height: 1.5;
  padding: 16px;
  border: none;
  resize: none;
  outline: none;
  overflow-y: auto;
}

.visual-editor { padding: 24px; overflow-y: auto; height: 100%; }
.visual-group { margin-top: 24px; border: 1px solid var(--border-color); padding: 16px; border-radius: var(--radius-sm); }
.radio-group { display: flex; gap: 12px; }
.radio-label { display: flex; align-items: center; gap: 4px; font-size: 13px; font-weight: 500; text-transform: none; color: var(--text-main); }

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  cursor: pointer;
}

.inheritance-info {
  background: #EFF6FF; border: 1px solid #BFDBFE; color: #1E40AF;
  padding: 8px; border-radius: 4px; display: flex; gap: 8px; align-items: start;
}
.text-xs { font-size: 11px; margin: 0; }

/* Visual Editor Styles */
.visual-section { margin-bottom: 32px; }
.mb-md { margin-bottom: 16px; }
.mt-lg { margin-top: 32px; }

.phone-faceplate {
  background: #262626;
  border-radius: 12px;
  padding: 20px;
  width: 320px;
  position: relative;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  display: flex;
  justify-content: space-between;
}

.screen-area {
  position: absolute; left: 60px; right: 60px; top: 20px; bottom: 20px;
  background: #1a1a1a;
  border: 2px solid #444;
  border-radius: 4px;
  display: flex; align-items: center; justify-content: center;
}
.screen-content { color: #555; text-align: center; }
.status-bar { font-size: 10px; margin-bottom: 4px; }
.idle-text { font-weight: 700; color: #777; }

.keys-grid {
   display: flex; flex-direction: column; gap: 12px; z-index: 10;
}

.line-key {
   display: flex; align-items: center; gap: 6px;
   background: #333; border: 1px solid #444;
   padding: 4px 8px; border-radius: 4px;
   cursor: pointer;
   color: #aaa;
   font-size: 10px;
   width: 50px;
   height: 32px;
   transition: all 0.2s;
}
.line-key:hover { background: #444; color: white; }
.line-key.configured { border-color: var(--primary-color); color: white; }

.key-led { width: 6px; height: 6px; border-radius: 50%; background: #111; border: 1px solid #555; }
.key-led.on { background: #10b981; border-color: #059669; box-shadow: 0 0 4px #10b981; }

.key-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 38px; }

.config-panel {
   background: #f8fafc;
   border: 1px solid var(--border-color);
   padding: 16px;
   border-radius: var(--radius-sm);
   margin-top: 16px;
   max-width: 400px;
}
.config-panel h4 { font-size: 14px; font-weight: 600; margin-bottom: 12px; }

.btn-secondary.small { padding: 4px 12px; font-size: 12px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.custom-params { display: flex; flex-direction: column; gap: 8px; max-width: 600px; }
.param-row { display: grid; grid-template-columns: 1fr 1fr 40px; gap: 8px; align-items: center; }
.param-row.header { font-size: 11px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; margin-bottom: 4px; }
.input-field.small { padding: 6px 8px; font-size: 12px; }
.btn-icon { background: none; border: none; cursor: pointer; font-size: 16px; display: flex; align-items: center; justify-content: center; }
.text-bad { color: var(--status-bad); }
</style>
