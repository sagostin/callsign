<template>
  <div class="view-header">
     <div class="header-content">
       <button class="back-link" @click="$router.push('/admin/queues')">← Back to Groups</button>
       <h2>{{ isNew ? 'New Ring Group' : 'Edit Ring Group' }}</h2>
     </div>
  </div>

  <div class="form-card">
     <div class="form-group">
        <label>Group Name</label>
        <input v-model="form.name" class="input-field" placeholder="Sales Team">
     </div>
     
      <div class="form-row">
         <div class="form-group">
            <label>Extension</label>
            <input v-model="form.extension" class="input-field" placeholder="500">
         </div>
         <div class="form-group">
            <label>Ring Strategy</label>
            <select v-model="form.strategy" class="input-field">
               <option value="simultaneous">Ring All</option>
               <option value="sequence">Sequential</option>
               <option value="enterprise">Enterprise</option>
               <option value="rollover">Rollover</option>
               <option value="random">Random</option>
            </select>
         </div>
      </div>

      <div class="form-row">
         <div class="form-group">
            <label>Ring Timeout (seconds per member)</label>
            <input type="number" v-model="form.timeout" class="input-field" min="5" max="120" />
            <p class="hint">How long to ring each member before trying next</p>
         </div>
         <div class="form-group">
            <label>Ringback Tone</label>
            <select v-model="ringGroupSettings.ringback" class="input-field">
               <option value="default">Default</option>
               <option value="music">Music</option>
               <option value="silent">Silent</option>
            </select>
         </div>
      </div>

      <div class="form-group">
         <label class="checkbox-label">
            <input type="checkbox" v-model="ringGroupSettings.skipBusyMembers" />
            Skip members who are busy
         </label>
         <p class="hint">When enabled, busy members are skipped in the ring order</p>
      </div>

      <div class="divider"></div>

      <div class="form-group">
         <label>On No Answer - Destination Type</label>
         <select v-model="ringGroupSettings.timeoutDestType" class="input-field">
            <option value="">Select action...</option>
            <option value="voicemail">Voicemail</option>
            <option value="extension">Extension</option>
            <option value="queue">Queue</option>
            <option value="ring_group">Ring Group</option>
            <option value="ivr">IVR Menu</option>
            <option value="hangup">Hang Up</option>
         </select>
      </div>

      <div class="form-group" v-if="ringGroupSettings.timeoutDestType && ringGroupSettings.timeoutDestType !== 'hangup'">
         <label>Destination</label>
         <input type="text" v-model="ringGroupSettings.timeoutDestValue" class="input-field" placeholder="Extension or number" />
      </div>
     
     <div class="divider"></div>
     
     <div class="form-group">
        <div class="flex justify-between items-center mb-2">
           <label>Group Members (Destinations)</label>
           <button class="btn-secondary small" @click="addMember">+ Add Member</button>
        </div>
        
        <div class="members-list">
           <div v-for="(mem, idx) in form.members" :key="idx" class="member-row">
              <select v-model="mem.type" class="input-field small" style="width: 100px">
                 <option value="user">User</option>
                 <option value="external">External</option>
                 <option value="device">Device</option>
              </select>
              <input v-model="mem.target" class="input-field small flex-1" placeholder="Extension or Number">
              <button class="text-red-500 hover:text-red-700 px-2" @click="removeMember(idx)">×</button>
           </div>
           <div v-if="form.members.length === 0" class="text-xs text-slate-400 italic p-2 text-center bg-slate-50 rounded">
              No members added.
           </div>
        </div>
     </div>

     <div class="form-actions border-t border-slate-200 pt-4 mt-6">
        <button class="btn-secondary" @click="$router.push('/admin/queues')">Cancel</button>
        <button class="btn-primary" @click="save">Save Group</button>
     </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ringGroupsAPI } from '../../services/api'

const toast = inject('toast')
const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)

const form = ref({
   name: '',
   extension: '',
   strategy: 'simultaneous',
   timeout: 30,
   members: []
})

const ringGroupSettings = ref({
   skipBusyMembers: true,
   timeoutDestType: '',
   timeoutDestValue: '',
   ringback: 'default',
})

onMounted(async () => {
  if (!isNew.value) {
    try {
      const res = await ringGroupsAPI.get(route.params.id)
      const d = res.data
      form.value = {
        name: d.name || '',
        extension: d.extension || '',
        strategy: d.strategy || 'simultaneous',
        timeout: d.timeout || 30,
        members: (d.members || []).map(m => ({
          type: m.type || 'user',
          target: m.extension || m.target || ''
        }))
      }
      ringGroupSettings.value = {
        skipBusyMembers: d.skipBusyMembers ?? true,
        timeoutDestType: d.timeoutDestType || '',
        timeoutDestValue: d.timeoutDestValue || '',
        ringback: d.ringback || 'default',
      }
    } catch (err) {
      toast?.error(err.message, 'Failed to load ring group')
    }
  }
})

const addMember = () => form.value.members.push({ type: 'user', target: '' })
const removeMember = (idx) => form.value.members.splice(idx, 1)

const save = async () => {
  try {
    const payload = {
      name: form.value.name,
      extension: form.value.extension,
      strategy: form.value.strategy,
      timeout: form.value.timeout,
      members: form.value.members.filter(m => m.target),
      skip_busy_members: ringGroupSettings.value.skipBusyMembers,
      timeout_dest_type: ringGroupSettings.value.timeoutDestType,
      timeout_dest_value: ringGroupSettings.value.timeoutDestValue,
      ringback: ringGroupSettings.value.ringback,
    }
    if (isNew.value) {
      await ringGroupsAPI.create(payload)
      toast?.success('Ring group created')
    } else {
      await ringGroupsAPI.update(route.params.id, payload)
      toast?.success('Ring group updated')
    }
    router.push('/admin/queues')
  } catch (err) {
    toast?.error(err.message, 'Failed to save ring group')
  }
}
</script>

<style scoped>
.view-header { margin-bottom: 24px; }
.back-link { background: none; border: none; color: var(--text-muted); padding: 0; font-size: 11px; cursor: pointer; }
.back-link:hover { color: var(--primary-color); text-decoration: underline; }

.form-card { background: white; padding: 24px; border-radius: var(--radius-md); border: 1px solid var(--border-color); max-width: 600px; }
.form-group { margin-bottom: 16px; display: flex; flex-direction: column; gap: 6px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 8px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.small { padding: 6px; font-size: 13px; }

.btn-primary { background: var(--primary-color); color: white; border: none; padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); padding: 10px 24px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary.small { padding: 4px 10px; font-size: 11px; }

.form-actions { display: flex; justify-content: flex-end; gap: 12px; }
.divider { height: 1px; background: var(--border-color); margin: 12px 0; }

.members-list { display: flex; flex-direction: column; gap: 8px; }
.member-row { display: flex; gap: 8px; align-items: center; }
.flex-1 { flex: 1; }
.hint { font-size: 11px; color: var(--text-muted); margin: 0; }
.checkbox-label { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.checkbox-label input { width: 16px; height: 16px; }
</style>
