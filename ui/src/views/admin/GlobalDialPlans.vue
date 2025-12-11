<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Global Dial Plan</h2>
      <p class="text-muted text-sm">System-wide routing rules (Context: public/default).</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="openCreate">+ New Global Rule</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="rules" actions>
    <template #context="{ value }">
       <span class="badge">{{ value }}</span>
    </template>
    
    <template #actions="{ row }">
      <button class="btn-link" @click="editRule(row)">Edit</button>
      <button class="btn-link text-bad" @click="deleteRule(row)">Delete</button>
    </template>
  </DataTable>

  <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
      <h3 class="text-lg font-bold mb-4">{{ isEditing ? 'Edit Rule' : 'New Global Rule' }}</h3>
      
      <div class="space-y-4">
        <div class="grid grid-cols-3 gap-4">
           <div class="col-span-1">
             <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Order</label>
             <input v-model="activeRule.order" class="w-full border p-2 rounded text-sm" placeholder="000" />
           </div>
           <div class="col-span-2">
             <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Context</label>
             <select v-model="activeRule.context" class="w-full border p-2 rounded text-sm bg-white">
                <option value="public">public</option>
                <option value="default">default</option>
             </select>
           </div>
        </div>

        <div>
          <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Rule Name</label>
          <input v-model="activeRule.name" class="w-full border p-2 rounded text-sm" placeholder="e.g. intercept_all" />
        </div>

        <div>
           <label class="block text-xs font-bold text-gray-500 uppercase mb-1">First Action</label>
           <input v-model="activeRule.action" class="w-full border p-2 rounded text-sm" placeholder="application:data" />
        </div>
      </div>

      <div class="flex justify-end gap-2 mt-6">
        <button class="btn-link" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveRule">Save Rule</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from '../../components/common/DataTable.vue'

const columns = [
  { key: 'order', label: 'Order', width: '80px' },
  { key: 'name', label: 'Name' },
  { key: 'context', label: 'Context' },
  { key: 'action', label: 'First Action' },
]

const rules = ref([
  { id: 1, order: '005', name: 'debug_dump', context: 'public', action: 'info' },
  { id: 2, order: '010', name: 'global_intercept', context: 'default', action: 'answer' },
])

const showModal = ref(false)
const isEditing = ref(false)
const activeRule = ref({ order: '', name: '', context: 'public', action: '' })

const openCreate = () => {
  activeRule.value = { order: '', name: '', context: 'public', action: '' }
  isEditing.value = false
  showModal.value = true
}

const editRule = (row) => {
  activeRule.value = { ...row }
  isEditing.value = true
  showModal.value = true
}

const saveRule = () => {
  if (isEditing.value) {
    const idx = rules.value.findIndex(r => r.id === activeRule.value.id)
    if (idx !== -1) rules.value[idx] = { ...activeRule.value }
  } else {
    rules.value.push({ ...activeRule.value, id: Date.now() })
  }
  showModal.value = false
}

const deleteRule = (row) => confirm(`Delete rule ${row.name}?`) && alert('Deleted')
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary-color);
  font-size: var(--text-xs);
  margin-left: 8px;
  cursor: pointer;
  font-weight: 500;
}

.text-bad { color: var(--status-bad); }

.badge { background: #f3f4f6; padding: 2px 8px; border-radius: 99px; font-size: 11px; }
</style>
