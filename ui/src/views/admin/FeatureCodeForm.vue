<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Feature Code' : 'Edit Feature Code' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
       <div class="form-row">
        <div class="form-group">
          <label>Code</label>
          <input v-model="form.code" type="text" class="input-field" placeholder="e.g. *72">
        </div>
        <div class="form-group">
          <label>Type</label>
          <select v-model="form.type" class="input-field">
             <option value="forwarding">Call Forwarding</option>
             <option value="dnd">Do Not Disturb</option>
             <option value="voicemail">Voicemail Access</option>
             <option value="custom">Custom</option>
          </select>
        </div>
      </div>

      <div class="form-group">
        <label>Description / Label</label>
        <input v-model="form.description" type="text" class="input-field" placeholder="e.g. Enable Call Forwarding">
      </div>
      
       <div class="form-group">
        <label>Status</label>
        <select v-model="form.enabled" class="input-field">
           <option :value="true">Enabled</option>
           <option :value="false">Disabled</option>
        </select>
      </div>

      <div class="form-group" v-if="form.type === 'forwarding'">
         <div class="help-box">
            <strong>Usage:</strong> Dial <code>{{ form.code || '*72' }}</code> followed by the destination number to enable forwarding.
         </div>
      </div>
      <div class="form-group" v-if="form.type === 'dnd'">
         <div class="help-box">
            <strong>Usage:</strong> Dial <code>{{ form.code || '*78' }}</code> to toggle Do Not Disturb mode.
         </div>
      </div>
      
      <div v-if="form.type === 'custom'" class="form-row mt-4">
         <div class="form-group">
            <label>Action / Application</label>
            <input v-model="form.action" type="text" class="input-field" placeholder="e.g. lua, bridge, playback">
         </div>
         <div class="form-group">
            <label>Argument / Data</label>
            <input v-model="form.argument" type="text" class="input-field" placeholder="e.g. script.lua, user/1000">
         </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Save Feature Code</button>
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

const form = ref({
  code: '',
  type: 'custom',
  description: '',
  description: '',
  enabled: true,
  action: '',
  argument: ''
})

const save = () => {
  console.log('Saving feature code:', form.value)
  router.back()
}
</script>

<style scoped>
.form-container { max-width: 600px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; }
.help-box { background: #eff6ff; border: 1px solid #bfdbfe; color: #1e40af; padding: 12px; border-radius: 6px; font-size: 13px; margin-top: 16px; }
code { background: rgba(255,255,255,0.5); padding: 2px 4px; border-radius: 4px; font-family: monospace; font-weight: bold; }
</style>
