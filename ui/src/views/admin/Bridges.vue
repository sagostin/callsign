<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Bridges (Internal)</h2>
      <p class="text-muted text-sm">Internal extensions that bridge calls to other services or conference rooms.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/bridges/new')">+ Add Bridge Extension</button>
    </div>
  </div>

  <DataTable :columns="columns" :data="bridgeExtensions" actions>
    <template #type="{ value }">
       <span class="badge">{{ value }}</span>
    </template>
    <template #actions>
       <button class="btn-link" @click="$router.push('/admin/bridges/1')">Edit</button>
       <button class="btn-link text-bad" @click="deleteExtension">Delete</button>
    </template>
  </DataTable>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import DataTable from '../../components/common/DataTable.vue'
import { extensionsAPI } from '../../services/api'

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'extension', label: 'Extension' },
  { key: 'type', label: 'Bridge Type' },
  { key: 'target', label: 'Target' }
]

const bridgeExtensions = ref([])

const fetchBridges = async () => {
  try {
    const res = await extensionsAPI.listBridges ? await extensionsAPI.listBridges() : await extensionsAPI.list({ type: 'bridge' })
    bridgeExtensions.value = (res.data?.extensions || res.data || []).map(b => ({
      id: b.id,
      name: b.description || b.effective_caller_id_name || b.extension,
      extension: b.extension,
      type: b.bridge_type || b.type || 'Conference',
      target: b.bridge_target || b.destination || ''
    }))
  } catch (err) {
    console.error('Failed to load bridges:', err)
    bridgeExtensions.value = []
  }
}

onMounted(fetchBridges)

const deleteExtension = async (bridge) => {
  if (!confirm(`Delete bridge "${bridge.name}"?`)) return
  try {
    await extensionsAPI.delete(bridge.id)
    bridgeExtensions.value = bridgeExtensions.value.filter(b => b.id !== bridge.id)
  } catch (err) {
    console.error('Failed to delete bridge:', err)
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

.badge { padding: 2px 6px; background: var(--bg-secondary); border-radius: 4px; font-size: 11px; font-weight: 600; }
</style>
