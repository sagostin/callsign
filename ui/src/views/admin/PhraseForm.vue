<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Phrase' : 'Edit Phrase' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Phrase Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. ivr_welcome">
        <span class="help-text">Unique identifier for this phrase.</span>
      </div>

       <div class="form-group">
        <label>Primary Language</label>
        <select v-model="form.language" class="input-field">
          <option value="en-us">English (US)</option>
          <option value="es-mx">Spanish (MX)</option>
          <option value="fr-ca">French (CA)</option>
        </select>
      </div>

       <div class="form-group">
        <label>Description</label>
        <textarea v-model="form.description" class="input-field" rows="2"></textarea>
      </div>

      <div class="divider"></div>
      
      <div class="section-title">Phrase Resources (Files)</div>
      
      <div class="file-list">
        <div v-for="(file, idx) in form.files" :key="idx" class="file-item">
           <div class="file-order">{{ idx + 1 }}</div>
             <div class="file-row">
                 <select v-model="file.recording_id" class="input-field" style="flex: 1;">
                     <option value="">Select Recording...</option>
                     <option value="rec_1">welcome.wav</option>
                     <option value="rec_2">options.wav</option>
                 </select>
                 <button class="btn-icon-danger" @click="removeFile(idx)">×</button>
             </div>
        </div>
        
        <button class="btn-dashed" @click="addFile">+ Add File Resource</button>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Save Phrase</button>
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
  name: '',
  language: 'en-us',
  description: '',
  files: [{ recording_id: '' }]
})

const addFile = () => {
  form.value.files.push({ recording_id: '' })
}

const removeFile = (idx) => {
    form.value.files.splice(idx, 1)
}

const save = () => {
    router.back()
}
</script>

<style scoped>
.form-container { max-width: 600px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field.small { padding: 6px; font-size: 13px; }
.input-field:focus { border-color: var(--primary-color); }
.help-text { font-size: 11px; color: var(--text-muted); }

.divider { height: 1px; background: var(--border-color); margin: 24px 0; }
.section-title { font-weight: 600; margin-bottom: 12px; font-size: 14px; }

.file-list { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; }
.file-item { display: flex; align-items: center; gap: 12px; }
.file-order {
  width: 24px; height: 24px; background: var(--bg-app); border-radius: 50%; 
  display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; color: var(--text-muted);
}
.file-info { flex: 1; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary.small { font-size: 12px; padding: 6px 12px; align-self: flex-start; }
.btn-icon { background: none; border: none; cursor: pointer; font-size: 14px; }

.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; }
.text-bad { color: var(--status-bad); }
</style>
