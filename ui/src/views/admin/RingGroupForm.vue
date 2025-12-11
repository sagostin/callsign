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
              <option value="ring-all">Ring All</option>
              <option value="sequential">Sequential</option>
              <option value="enterprise">Enterprise</option>
              <option value="random">Random</option>
           </select>
        </div>
     </div>

     <div class="form-group">
        <label>Ring Timeout (Seconds)</label>
        <input v-model="form.timeout" type="number" class="input-field" value="30">
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
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => !route.params.id)

const form = ref({
   name: '',
   extension: '',
   strategy: 'ring-all',
   timeout: 30,
   members: [
      { type: 'user', target: '1001' },
      { type: 'user', target: '1002' }
   ]
})

if (!isNew.value) {
   // Mock load
   form.value = {
      name: 'Sales Team',
      extension: '500',
      strategy: 'enterprise',
      timeout: 30,
      members: [
          { type: 'user', target: '101' },
          { type: 'user', target: '102' },
          { type: 'external', target: '555-0011' }
      ]
   }
}

const addMember = () => form.value.members.push({ type: 'user', target: '' })
const removeMember = (idx) => form.value.members.splice(idx, 1)

const save = () => {
   alert('Ring Group Saved')
   router.push('/admin/queues')
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
</style>
