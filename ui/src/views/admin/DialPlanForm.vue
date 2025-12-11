<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Dial Plan' : 'Edit Dial Plan' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="flex justify-between items-center mb-6">
         <div class="form-group mb-0">
           <label>Name</label>
           <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Local Calls" style="width: 300px">
         </div>
         <div class="toggle-row">
            <span class="text-sm font-semibold mr-2 text-slate-600">Raw XML Mode</span>
            <label class="switch">
              <input type="checkbox" v-model="rawMode">
              <span class="slider round"></span>
            </label>
         </div>
      </div>
      
      <!-- RAW XML EDITOR -->
      <div v-if="rawMode" class="raw-editor">
         <textarea v-model="form.rawXml" class="code-area" spellcheck="false"></textarea>
      </div>

      <!-- VISUAL EDITOR -->
      <div v-else class="visual-editor">
         <div class="form-group mb-6">
            <label>Condition (Regex)</label>
            <div class="code-editor">
               <input v-model="form.condition" type="text" class="input-field code" placeholder="^(\d{10})$">
            </div>
            <span class="help-text">Regex pattern to match the dialed number.</span>
         </div>
         
         <div class="form-group mb-6">
            <div class="flex justify-between items-end mb-2">
               <label>Actions Sequence</label>
               <button class="btn-secondary small" @click="addAction">+ Add Action</button>
            </div>
            
            <div class="actions-list border rounded border-slate-200 overflow-hidden">
               <table class="w-full text-left">
                  <thead class="bg-slate-50 border-b border-slate-200">
                     <tr>
                        <th class="p-3 text-xs font-bold text-slate-500 w-10">#</th>
                        <th class="p-3 text-xs font-bold text-slate-500 w-40">Application</th>
                        <th class="p-3 text-xs font-bold text-slate-500">Data</th>
                        <th class="p-3 w-10"></th>
                     </tr>
                  </thead>
                  <tbody>
                     <tr v-for="(act, idx) in form.actions" :key="idx" class="border-b border-slate-100 last:border-0 hover:bg-slate-50">
                        <td class="p-3 text-xs font-mono text-slate-400">{{ idx + 1 }}</td>
                        <td class="p-3">
                           <select v-model="act.application" class="input-field w-full text-xs">
                              <option value="bridge">bridge</option>
                              <option value="transfer">transfer</option>
                              <option value="hangup">hangup</option>
                              <option value="set">set</option>
                              <option value="answer">answer</option>
                              <option value="playback">playback</option>
                              <option value="voicemail">voicemail</option>
                           </select>
                        </td>
                        <td class="p-3">
                           <input v-model="act.data" class="input-field w-full text-xs font-mono" placeholder="Arguments...">
                        </td>
                        <td class="p-3 text-right">
                           <button class="text-red-500 hover:text-red-700 font-bold px-2" @click="removeAction(idx)">×</button>
                        </td>
                     </tr>
                  </tbody>
               </table>
               <div v-if="form.actions.length === 0" class="p-4 text-center text-xs text-slate-400 italic">
                  No actions defined. Add one to route the call.
               </div>
            </div>
         </div>
         
         <div class="form-row">
            <div class="form-group">
               <label>Order / Priority</label>
               <input v-model="form.order" type="number" class="input-field" placeholder="100">
            </div>
            <div class="form-group flex items-center pt-6">
               <input type="checkbox" id="cont" v-model="form.continue" class="mr-2">
               <label for="cont" class="inline cursor-pointer">Continue (Fallthrough)</label>
            </div>
         </div>
      </div>

      <div class="form-actions border-t border-slate-200 pt-4 mt-4">
        <button class="btn-primary" @click="save">Save Dial Plan</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)
const rawMode = ref(false)

const form = ref({
  name: '',
  condition: '',
  order: 100,
  continue: false,
  actions: [
     { application: 'bridge', data: 'user/1000' }
  ],
  rawXml: `<extension name="Local_Calls">
  <condition field="destination_number" expression="^(\\d{10})$">
    <action application="set" data="hangup_after_bridge=true"/>
    <action application="bridge" data="sofia/gateway/default/$1"/>
  </condition>
</extension>`
})

const addAction = () => {
   form.value.actions.push({ application: 'bridge', data: '' })
}
const removeAction = (idx) => {
   form.value.actions.splice(idx, 1)
}

const save = () => {
  console.log('Saving dial plan:', form.value)
  router.back()
}
</script>

<style scoped>
.form-container { max-width: 700px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }
.input-field.code { font-family: monospace; }
.help-text { font-size: 11px; color: var(--text-muted); }

.code-editor { background: var(--bg-app); padding: 4px; border-radius: var(--radius-sm); }


.switch { position: relative; display: inline-block; width: 34px; height: 20px; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .4s; }
.slider:before { position: absolute; content: ""; height: 16px; width: 16px; left: 2px; bottom: 2px; background-color: white; transition: .4s; }
input:checked + .slider { background-color: var(--primary-color); }
input:checked + .slider:before { transform: translateX(14px); }
.slider.round { border-radius: 34px; }
.slider.round:before { border-radius: 50%; }

.code-area {
  width: 100%; height: 300px; background: #1e1e1e; color: #d4d4d4;
  font-family: 'Fira Code', monospace; padding: 16px; border-radius: 6px; resize: vertical;
}

.btn-secondary.small { padding: 4px 10px; font-size: 11px; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 0; display: flex; justify-content: flex-end; }
</style>
