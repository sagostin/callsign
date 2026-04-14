<template>
  <div class="form-container">
    <div class="form-header">
      <h2>{{ isNew ? 'Schedule Wake Up Call' : 'Edit Wake Up Call' }}</h2>
      <button class="btn-secondary" @click="$router.back()">Cancel</button>
    </div>

    <div class="form-card">
      <div class="form-group">
        <label>Room / Extension</label>
        <input v-model="form.extension" type="text" class="input-field" placeholder="e.g. 101">
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Time</label>
          <input v-model="form.time" type="time" class="input-field">
        </div>
        <div class="form-group">
          <label>Date (Optional)</label>
          <input v-model="form.date" type="date" class="input-field">
        </div>
      </div>
      
       <div class="form-group">
        <label>Recurrence</label>
        <select v-model="form.recurrence" class="input-field">
           <option value="once">One Time</option>
           <option value="daily">Daily</option>
           <option value="weekdays">Weekdays (M-F)</option>
        </select>
      </div>

       <div class="form-group">
        <label>Status</label>
        <select v-model="form.status" class="input-field">
           <option value="pending">Pending</option>
           <option value="active">Active</option>
           <option value="cancelled">Cancelled</option>
        </select>
      </div>

      <div class="form-actions">
        <button class="btn-primary" :disabled="isSaving" @click="save">{{ isSaving ? 'Saving...' : 'Save Schedule' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { wakeupCallsAPI } from '../../services/api'

const route = useRoute()
const router = useRouter()
const toast = inject('toast')
const isNew = computed(() => !route.params.id)

const form = ref({
  extension: '',
  time: '07:00',
  date: '',
  recurrence: 'once',
  status: 'pending'
})

const isSaving = ref(false)

const save = async () => {
  // Guard clause: extension is required
  if (!form.value.extension?.trim()) {
    toast?.error('Room/Extension is required')
    return
  }

  isSaving.value = true
  try {
    const payload = {
      extension: form.value.extension.trim(),
      time: form.value.time,
      date: form.value.date || null,
      recurrence: form.value.recurrence,
      status: form.value.status
    }

    if (isNew.value) {
      await wakeupCallsAPI.create(payload)
      toast?.success('Wake up call scheduled')
    } else {
      await wakeupCallsAPI.update(route.params.id, payload)
      toast?.success('Wake up call updated')
    }

    router.push('/admin/wakeup-calls')
  } catch (err) {
    toast?.error(err.message || 'Failed to save wake up call')
  } finally {
    isSaving.value = false
  }
}
</script>

<style scoped>
.form-container { max-width: 500px; margin: 0 auto; }
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
