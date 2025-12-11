<template>
  <div class="view-header">
     <div class="header-content">
       <button class="back-link" @click="$router.push('/admin/queues')">← Back to Queues</button>
       <h2>{{ isNew ? 'New Call Queue' : 'Edit Queue' }}</h2>
     </div>
  </div>

  <div class="form-card">
     <div class="form-group">
        <label>Queue Name</label>
        <input v-model="form.name" class="input-field" placeholder="Support Tier 1">
     </div>
     
     <div class="form-row">
        <div class="form-group">
           <label>Extension</label>
           <input v-model="form.extension" class="input-field" placeholder="8000">
        </div>
        <div class="form-group">
           <label>Agent Strategy</label>
           <select v-model="form.strategy" class="input-field">
              <option value="longest-idle-agent">Longest Idle Agent</option>
              <option value="round-robin">Round Robin</option>
              <option value="top-down">Top Down</option>
              <option value="ring-all">Ring All</option>
           </select>
        </div>
     </div>
     
     <div class="form-group">
        <label>Hold Music</label>
        <select v-model="form.moh" class="input-field">
           <option value="default">Default Music</option>
           <option value="classical">Classical</option>
           <option value="pop">Pop Mix</option>
        </select>
     </div>

     <div class="divider"></div>
     
     <div class="form-group">
        <div class="flex justify-between items-center mb-2">
           <label>Static Agents</label>
           <button class="btn-secondary small" @click="addAgent">+ Add Agent</button>
        </div>
        
        <div class="members-list">
           <div v-for="(agent, idx) in form.agents" :key="idx" class="member-row">
              <select v-model="agent.tier" class="input-field small" style="width: 80px">
                 <option value="1">Tier 1</option>
                 <option value="2">Tier 2</option>
              </select>
              <input v-model="agent.user" class="input-field small flex-1" placeholder="User or Extension">
              <button class="text-red-500 hover:text-red-700 px-2" @click="removeAgent(idx)">×</button>
           </div>
           <div v-if="form.agents.length === 0" class="text-xs text-slate-400 italic p-2 text-center bg-slate-50 rounded">
              No static agents. Agents can login dynamically.
           </div>
        </div>
     </div>

     <div class="form-actions border-t border-slate-200 pt-4 mt-6">
        <button class="btn-secondary" @click="$router.push('/admin/queues')">Cancel</button>
        <button class="btn-primary" @click="save">Save Queue</button>
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
   strategy: 'longest-idle-agent',
   moh: 'default',
   agents: [
      { tier: '1', user: '1001' }
   ]
})

if (!isNew.value) {
   form.value = {
      name: 'Support Tier 1',
      extension: '8001',
      strategy: 'longest-idle-agent',
      moh: 'default',
      agents: [
          { tier: '1', user: '101 - Alice' },
          { tier: '1', user: '102 - Bob' },
          { tier: '2', user: '105 - Manager' }
      ]
   }
}

const addAgent = () => form.value.agents.push({ tier: '1', user: '' })
const removeAgent = (idx) => form.value.agents.splice(idx, 1)

const save = () => {
   alert('Queue Saved')
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
