<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'New Stream' : 'Edit Stream' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>
    
    <div class="form-card">
      <div class="form-group">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Sample Rate</label>
          <select v-model="form.rate" class="input-field">
            <option value="8000">8000 Hz</option>
            <option value="16000">16000 Hz</option>
            <option value="32000">32000 Hz</option>
            <option value="48000">48000 Hz</option>
          </select>
        </div>
        <div class="form-group">
          <label>Channels</label>
          <select v-model="form.channels" class="input-field">
            <option value="1">Mono (1)</option>
            <option value="2">Stereo (2)</option>
          </select>
        </div>
      </div>

       <div class="form-row">
        <div class="form-group">
          <label>Shuffle</label>
          <select v-model="form.shuffle" class="input-field">
             <option :value="true">True</option>
             <option :value="false">False</option>
          </select>
        </div>
        <div class="form-group">
          <label>Interval (Timer)</label>
           <input v-model="form.interval" type="number" class="input-field" placeholder="20">
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="save">Save Stream</button>
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
  path: '',
  rate: '48000',
  channels: '1',
  shuffle: true,
  interval: 20
})

const save = () => {
  console.log('Saving stream:', form.value)
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
.input-note { font-size: 11px; color: var(--text-muted); margin-bottom: 4px; }
.input-field:focus { border-color: var(--primary-color); }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.form-actions { margin-top: 24px; display: flex; justify-content: flex-end; }
</style>
