<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Conference' : 'Edit Conference' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Conference Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Sales Weekly">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Extension</label>
          <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 3001">
        </div>
        <div class="form-group">
          <label>Profile</label>
          <select v-model="form.profile" class="input-field">
             <option value="default">Default</option>
             <option value="wideband">Wideband</option>
             <option value="video">Video</option>
          </select>
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>PIN Number (Optional)</label>
          <input v-model="form.pin" type="text" class="input-field" placeholder="e.g. 1234">
        </div>
         <div class="form-group">
          <label>Moderator PIN (Optional)</label>
          <input v-model="form.mod_pin" type="text" class="input-field" placeholder="e.g. 5678">
        </div>
      </div>

       <div class="form-group">
        <label>Description</label>
        <textarea v-model="form.description" class="input-field" rows="2"></textarea>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Save Conference</button>
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
  extension: '',
  profile: 'default',
  pin: '',
  mod_pin: '',
  description: ''
})

const save = () => {
  console.log('Saving conference:', form.value)
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
</style>
