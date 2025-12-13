<template>
  <div class="view-container">
    <div class="view-header">
      <div class="header-content">
        <h2>Speed Dials</h2>
        <p class="text-muted text-sm">Manage speed dial prefixes and global contact lists.</p>
      </div>
      <button class="btn-primary" @click="openCreateModal">
        <PlusIcon class="icon-sm" /> New Group
      </button>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && groups.length === 0" class="empty-state">
      <ZapIcon class="empty-icon" />
      <h3>No Speed Dials Defined</h3>
      <p>Create a speed dial group to assign short codes (e.g. *01) to frequently called numbers.</p>
      <button class="btn-primary" @click="openCreateModal">Create First Group</button>
    </div>

    <!-- Groups Grid -->
    <div v-else-if="!loading" class="groups-grid">
      <div v-for="group in groups" :key="group.id" class="group-card">
        <div class="group-header">
          <div class="group-info">
            <h3>{{ group.name }}</h3>
            <div class="group-meta">
              <span class="prefix-badge">Prefix: {{ group.prefix }}</span>
              <span class="entry-count">{{ group.entries?.length || 0 }} entries</span>
            </div>
          </div>
          <div class="group-actions">
            <button class="btn-icon" @click="editGroup(group)" title="Edit"><EditIcon class="icon-sm" /></button>
            <button class="btn-icon text-bad" @click="deleteGroup(group)" title="Delete"><TrashIcon class="icon-sm" /></button>
          </div>
        </div>
        
        <div class="entries-list">
          <template v-if="group.entries?.length > 0">
            <div 
              v-for="(entry, idx) in group.entries" 
              :key="idx" 
              class="entry-row"
              draggable="true"
              @dragstart="dragStart(group, idx)"
              @dragover.prevent
              @drop="drop(group, idx)"
            >
              <GripVertical class="drag-handle" />
              <span class="entry-code">{{ group.prefix }}{{ entry.slot || (idx + 1) }}</span>
              <span class="entry-label">{{ entry.label }}</span>
              <span class="entry-dest">
                <PhoneIcon class="icon-xs" /> {{ entry.destination }}
              </span>
            </div>
          </template>
          <div v-else class="no-entries">No numbers added yet.</div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else class="loading-state">
      <RefreshCw class="spin" /> Loading...
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Speed Dial Group' : 'New Speed Dial Group' }}</h3>
          <button class="btn-icon" @click="closeModal"><X class="icon-sm" /></button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Group Name</label>
            <input v-model="form.name" class="input-field" placeholder="e.g. Executive Directory">
          </div>
          <div class="form-group">
            <label>Prefix</label>
            <input v-model="form.prefix" class="input-field" placeholder="e.g. *0">
            <span class="input-hint">Users dial prefix + slot number (e.g. *01, *02)</span>
          </div>
          <div class="form-group">
            <label>Description</label>
            <input v-model="form.description" class="input-field" placeholder="Optional description">
          </div>

          <div class="divider"></div>

          <div class="entries-section">
            <div class="section-header">
              <h4>Speed Dial Entries</h4>
              <button class="btn-secondary small" @click="addEntry">+ Add Entry</button>
            </div>
            <div class="entry-form-list">
              <div v-for="(entry, idx) in form.entries" :key="idx" class="entry-form-row">
                <input v-model.number="entry.slot" type="number" class="input-field slot-input" placeholder="#" min="1" max="99">
                <input v-model="entry.label" class="input-field label-input" placeholder="Label">
                <input v-model="entry.destination" class="input-field dest-input" placeholder="Phone number">
                <button class="btn-icon text-bad" @click="removeEntry(idx)"><TrashIcon class="icon-sm" /></button>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="closeModal">Cancel</button>
          <button class="btn-primary" @click="saveGroup" :disabled="!form.name || !form.prefix">
            {{ isEditing ? 'Save Changes' : 'Create Group' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { PlusIcon, ZapIcon, EditIcon, TrashIcon, PhoneIcon, X, RefreshCw, GripVertical } from 'lucide-vue-next'
import { speedDialsAPI } from '../../services/api'

const groups = ref([])
const loading = ref(false)
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const form = ref({
  name: '',
  prefix: '',
  description: '',
  entries: []
})

// Drag state
let dragGroupId = null
let dragIndex = null

onMounted(() => loadGroups())

async function loadGroups() {
  loading.value = true
  try {
    const response = await speedDialsAPI.list()
    groups.value = response.data?.data || []
  } catch (error) {
    console.error('Failed to load speed dial groups:', error)
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  form.value = { name: '', prefix: '', description: '', entries: [] }
  isEditing.value = false
  editingId.value = null
  showModal.value = true
}

function editGroup(group) {
  form.value = {
    name: group.name,
    prefix: group.prefix,
    description: group.description || '',
    entries: [...(group.entries || [])]
  }
  isEditing.value = true
  editingId.value = group.id
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

function addEntry() {
  const nextSlot = form.value.entries.length + 1
  form.value.entries.push({ slot: nextSlot, label: '', destination: '' })
}

function removeEntry(idx) {
  form.value.entries.splice(idx, 1)
}

async function saveGroup() {
  try {
    const payload = {
      name: form.value.name,
      prefix: form.value.prefix,
      description: form.value.description,
      enabled: true,
      entries: form.value.entries.filter(e => e.label && e.destination)
    }

    if (isEditing.value) {
      await speedDialsAPI.update(editingId.value, payload)
    } else {
      await speedDialsAPI.create(payload)
    }
    
    await loadGroups()
    closeModal()
  } catch (error) {
    console.error('Failed to save speed dial group:', error)
    alert('Failed to save speed dial group')
  }
}

async function deleteGroup(group) {
  if (!confirm(`Delete speed dial group "${group.name}"?`)) return
  try {
    await speedDialsAPI.delete(group.id)
    await loadGroups()
  } catch (error) {
    console.error('Failed to delete speed dial group:', error)
    alert('Failed to delete speed dial group')
  }
}

// Drag and drop for reordering entries
function dragStart(group, idx) {
  dragGroupId = group.id
  dragIndex = idx
}

async function drop(group, targetIdx) {
  if (dragGroupId !== group.id || dragIndex === targetIdx) return
  
  const entries = [...group.entries]
  const [moved] = entries.splice(dragIndex, 1)
  entries.splice(targetIdx, 0, moved)
  
  // Update slot numbers based on new order
  entries.forEach((e, i) => e.slot = i + 1)
  
  try {
    await speedDialsAPI.update(group.id, { ...group, entries })
    await loadGroups()
  } catch (error) {
    console.error('Failed to reorder entries:', error)
  }
  
  dragGroupId = null
  dragIndex = null
}
</script>

<style scoped>
.view-container { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.view-header h2 { margin: 0 0 4px; }

.btn-primary { display: flex; align-items: center; gap: 6px; background: var(--primary-color); color: white; border: none; padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary { background: white; border: 1px solid var(--border-color); padding: 8px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-secondary.small { padding: 6px 12px; font-size: 12px; }
.btn-icon { background: none; border: none; cursor: pointer; padding: 6px; color: var(--text-muted); border-radius: 4px; }
.btn-icon:hover { background: var(--bg-app); color: var(--text-primary); }

.empty-state { text-align: center; padding: 60px 20px; background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); }
.empty-icon { width: 48px; height: 48px; margin-bottom: 16px; color: var(--text-muted); }
.empty-state h3 { margin: 0 0 8px; }
.empty-state p { color: var(--text-muted); max-width: 320px; margin: 0 auto 24px; }

.loading-state { text-align: center; padding: 60px; color: var(--text-muted); }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.groups-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(360px, 1fr)); gap: 20px; }
.group-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }
.group-header { display: flex; justify-content: space-between; align-items: flex-start; padding: 16px; border-bottom: 1px solid var(--border-color); background: var(--bg-app); }
.group-info h3 { margin: 0 0 6px; font-size: 15px; font-weight: 600; }
.group-meta { display: flex; gap: 10px; align-items: center; }
.prefix-badge { font-size: 11px; font-family: monospace; background: #fef3c7; color: #b45309; padding: 2px 8px; border-radius: 4px; font-weight: 600; }
.entry-count { font-size: 12px; color: var(--text-muted); }
.group-actions { display: flex; gap: 4px; }

.entries-list { padding: 8px; max-height: 200px; overflow-y: auto; background: #f8fafc; }
.entry-row { display: flex; align-items: center; gap: 10px; padding: 8px 10px; background: white; border-radius: 6px; margin-bottom: 4px; cursor: grab; }
.entry-row:hover { background: #f1f5f9; }
.drag-handle { width: 14px; height: 14px; color: var(--text-muted); cursor: grab; }
.entry-code { font-family: monospace; font-size: 12px; font-weight: 600; color: var(--text-muted); min-width: 48px; }
.entry-label { flex: 1; font-size: 13px; font-weight: 500; }
.entry-dest { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--text-muted); }
.no-entries { text-align: center; padding: 24px; color: var(--text-muted); font-size: 13px; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); padding: 24px; }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 560px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.modal-body { padding: 20px; overflow-y: auto; flex: 1; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 12px; font-weight: 600; color: var(--text-muted); margin-bottom: 6px; text-transform: uppercase; }
.input-field { width: 100%; padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-hint { font-size: 11px; color: var(--text-muted); margin-top: 4px; display: block; }

.divider { height: 1px; background: var(--border-color); margin: 20px 0; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.section-header h4 { margin: 0; font-size: 14px; }

.entry-form-list { display: flex; flex-direction: column; gap: 8px; }
.entry-form-row { display: flex; gap: 8px; align-items: center; }
.slot-input { width: 60px; flex-shrink: 0; text-align: center; }
.label-input { flex: 1; }
.dest-input { flex: 1; }

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 12px; height: 12px; }
.text-bad { color: #dc2626 !important; }
</style>
