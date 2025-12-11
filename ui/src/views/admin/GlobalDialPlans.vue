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
    <template #dialplan_context="{ value }">
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
             <input v-model.number="activeRule.dialplan_order" type="number" class="w-full border p-2 rounded text-sm" placeholder="100" />
           </div>
           <div class="col-span-2">
             <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Context</label>
             <select v-model="activeRule.dialplan_context" class="w-full border p-2 rounded text-sm bg-white">
                <option value="public">public</option>
                <option value="default">default</option>
             </select>
           </div>
        </div>

        <div>
          <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Rule Name</label>
          <input v-model="activeRule.dialplan_name" class="w-full border p-2 rounded text-sm" placeholder="e.g. intercept_all" />
        </div>

        <div>
           <label class="block text-xs font-bold text-gray-500 uppercase mb-1">Description</label>
           <input v-model="activeRule.description" class="w-full border p-2 rounded text-sm" placeholder="Rule description" />
        </div>
      </div>

      <div class="flex justify-end gap-2 mt-6">
        <button class="btn-link" @click="showModal = false">Cancel</button>
        <button class="btn-primary" @click="saveRule" :disabled="saving">{{ saving ? 'Saving...' : 'Save Rule' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import { systemAPI } from '../../services/api'

const columns = [
  { key: 'dialplan_order', label: 'Order', width: '80px' },
  { key: 'dialplan_name', label: 'Name' },
  { key: 'dialplan_context', label: 'Context' },
  { key: 'description', label: 'Description' },
]

const rules = ref([])
const loading = ref(true)
const showModal = ref(false)
const isEditing = ref(false)
const saving = ref(false)

const activeRule = ref({ 
  dialplan_order: 100, 
  dialplan_name: '', 
  dialplan_context: 'public', 
  description: '',
  enabled: true
})

const loadDialplans = async () => {
  loading.value = true
  try {
    const response = await systemAPI.listDialplans()
    rules.value = response.data.data || response.data || []
  } catch (e) {
    console.error('Failed to load dial plans:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadDialplans)

const openCreate = () => {
  activeRule.value = { dialplan_order: 100, dialplan_name: '', dialplan_context: 'public', description: '', enabled: true }
  isEditing.value = false
  showModal.value = true
}

const editRule = (row) => {
  activeRule.value = { ...row }
  isEditing.value = true
  showModal.value = true
}

const saveRule = async () => {
  saving.value = true
  try {
    if (isEditing.value && activeRule.value.id) {
      await systemAPI.updateDialplan(activeRule.value.id, activeRule.value)
    } else {
      await systemAPI.createDialplan(activeRule.value)
    }
    await loadDialplans()
    showModal.value = false
  } catch (e) {
    alert('Failed to save dial plan: ' + e.message)
  } finally {
    saving.value = false
  }
}

const deleteRule = async (row) => {
  if (!confirm(`Delete dial plan "${row.dialplan_name}"?`)) return
  try {
    await systemAPI.deleteDialplan(row.id)
    await loadDialplans()
  } catch (e) {
    alert('Failed to delete dial plan: ' + e.message)
  }
}
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
