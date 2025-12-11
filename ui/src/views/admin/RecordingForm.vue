<template>
  <div class="form-container">
    <div class="form-header">
      <h2>Upload Recording</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
       <div class="form-group">
        <label>Recording Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. holiday_greeting.wav">
      </div>
      
       <div class="form-group">
        <label>Description</label>
        <textarea v-model="form.description" class="input-field" rows="2"></textarea>
      </div>

       <div class="form-group">
        <label>Audio File</label>
        <div class="upload-zone">
           <div class="upload-icon">⬆️</div>
           <div class="upload-text">
              <strong>Click to upload</strong> or drag and drop<br>
              <span class="sub-text">WAV, MP3, or OGG (Max 10MB)</span>
           </div>
           <input type="file" class="file-input" accept="audio/*">
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Upload File</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const form = ref({
  name: '',
  description: ''
})

const save = () => {
  // Mock upload logic
  console.log('Uploading...', form.value)
  router.back()
}
</script>

<style scoped>
.form-container { max-width: 500px; margin: 0 auto; }
.form-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; outline: none; }
.input-field:focus { border-color: var(--primary-color); }

.upload-zone {
  border: 2px dashed var(--border-color);
  border-radius: var(--radius-md);
  padding: 40px;
  text-align: center;
  background: var(--bg-app);
  cursor: pointer;
  position: relative;
  transition: border-color 0.2s;
  overflow: hidden;
}
.upload-zone:hover { border-color: var(--primary-color); }

.upload-text { font-size: 13px; color: var(--text-muted); line-height: 1.5; margin-top: 12px; }
.upload-icon { font-size: 24px; }
.file-input { position: absolute; top: 0; left: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; }
</style>
