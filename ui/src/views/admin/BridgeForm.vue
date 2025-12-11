<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Bridge' : 'Edit Bridge' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Bridge Name</label>
        <input v-model="form.name" type="text" class="input-field" placeholder="e.g. Sales Conference">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Extension</label>
          <input v-model="form.extension" type="text" class="input-field" placeholder="*100">
        </div>
        <div class="form-group">
          <label>Bridge Type</label>
          <select v-model="form.type" class="input-field">
             <option value="conference">Conference Room</option>
             <option value="paging">Paging / Intercom</option>
             <option value="flow">Call Flow / IVR</option>
             <option value="application">Application / Script</option>
          </select>
        </div>
      </div>

      <div class="form-group" v-if="form.type === 'conference'">
         <label>Conference Profile</label>
         <select v-model="form.target" class="input-field">
             <option value="default">Default Profile</option>
             <option value="wideband">Wideband Audio</option>
             <option value="video">Video Conference</option>
         </select>
      </div>

      <div class="form-group" v-else-if="form.type === 'paging'">
          <label>Target Group / Users</label>
          <input v-model="form.target" type="text" class="input-field" placeholder="group:sales or 1001,1002">
      </div>

      <div class="form-group" v-else>
          <label>Target / Destination</label>
          <input v-model="form.target" type="text" class="input-field" placeholder="Destination data...">
      </div>

      <div class="form-group">
        <label>Description</label>
        <textarea v-model="form.description" class="input-field" rows="3"></textarea>
      </div>

      <div class="form-actions">
        <button class="btn-danger-outline" v-if="!isNew" @click="deleteBridge">Delete</button>
        <div style="flex:1"></div>
        <button class="btn-primary" @click="save">Save Bridge</button>
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
  type: 'conference',
  target: '',
  description: ''
})

if (!isNew.value) {
    // Mock Load
    form.value = {
        name: 'Conf Bridge A',
        extension: '7001',
        type: 'conference',
        target: 'default',
        description: 'Main sales bridge.'
    }
}

const save = () => {
  alert('Bridge saved successfully.')
  router.back()
}

const deleteBridge = () => {
   if(confirm('Are you sure you want to delete this bridge?')) {
      alert('Bridge deleted.')
      router.back()
   }
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
.btn-danger-outline { background: white; border: 1px solid var(--status-bad); color: var(--status-bad); padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.form-actions { margin-top: 24px; display: flex; gap: 12px; align-items: center; }
</style>
