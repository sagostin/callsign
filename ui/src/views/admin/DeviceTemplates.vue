<template>
  <div class="view-header">
    <div class="header-content">
      <button class="back-link" @click="$router.push('/admin/devices')">← Back to Devices</button>
      <h2>Device Templates</h2>
      <p class="text-muted text-sm">Manage provisioning templates for different device models.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="showCreateModal = true">+ New Template</button>
    </div>
  </div>

  <div class="templates-grid">
    <div class="template-card" v-for="template in templates" :key="template.id">
      <div class="card-header">
        <div class="icon-box">
          <FileCode class="icon" />
        </div>
        <div class="card-actions">
          <button class="btn-icon" @click.stop="toggleDropdown(template.id)">
            <MoreVertical class="icon-sm" />
          </button>
          <div v-if="activeDropdown === template.id" class="dropdown-menu">
            <button @click="cloneTemplate(template)">Clone</button>
            <button @click="exportTemplate(template)">Export CFG</button>
            <button class="text-bad" @click="deleteTemplate(template)">Delete</button>
          </div>
        </div>
      </div>
      <div class="card-body">
        <h3>{{ template.name }}</h3>
        <p class="model">{{ template.model }}</p>
        <div class="tags">
          <span class="tag">
            <span class="tag-label">Firmware</span>
            <span class="tag-value">{{ template.firmware }}</span>
          </span>
          <span class="tag devices">
            <PhoneIcon class="tag-icon" />
            {{ template.provisions }}
          </span>
        </div>
        <div class="master-ref" v-if="template.masterTemplate">
          <span class="ref-label">Based on:</span>
          <span class="ref-value">{{ template.masterTemplate }}</span>
        </div>
      </div>
      <div class="card-footer">
        <button class="btn-secondary full-width" @click="$router.push(`/admin/devices/templates/${template.id}`)">
          <SettingsIcon class="btn-icon-left" />
          Edit Config
        </button>
      </div>
    </div>
  </div>

  <!-- Create Template Modal -->
  <div class="modal-overlay" v-if="showCreateModal" @click.self="showCreateModal = false">
    <div class="modal-card">
      <div class="modal-header">
        <h3>Create New Template</h3>
        <button class="btn-icon" @click="showCreateModal = false"><XIcon class="icon-sm" /></button>
      </div>
      
      <div class="modal-body">
        <div class="form-group">
          <label>Base Master Template</label>
          <select class="input-field" v-model="newTemplateBase">
            <option value="" disabled>Select Manufacturer / Model...</option>
            <optgroup label="Yealink">
              <option value="yealink_t54">Yealink T54W (Master)</option>
              <option value="yealink_t57">Yealink T57W (Master)</option>
              <option value="yealink_w60">Yealink W60B DECT (Master)</option>
            </optgroup>
            <optgroup label="Poly">
              <option value="poly_ccx">Poly CCX 500 (Master)</option>
              <option value="poly_vvx">Poly VVX 400/500 (Master)</option>
            </optgroup>
            <optgroup label="Grandstream">
              <option value="grandstream_gxp">Grandstream GXP2100 (Master)</option>
            </optgroup>
          </select>
          <span class="help-text">Master templates are managed in the System Admin panel.</span>
        </div>
        
        <div class="form-group">
          <label>Template Name</label>
          <input type="text" class="input-field" placeholder="e.g. Sales Department Yealink" v-model="newTemplateName">
        </div>
        
        <div class="form-group">
          <label>Description (Optional)</label>
          <textarea class="input-field" rows="2" placeholder="Brief description of this template's purpose..." v-model="newTemplateDesc"></textarea>
        </div>
      </div>
      
      <div class="modal-actions">
        <button class="btn-secondary" @click="showCreateModal = false">Cancel</button>
        <button class="btn-primary" @click="confirmCreate" :disabled="!newTemplateBase || !newTemplateName">Create Template</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, inject } from 'vue'
import { FileCode, MoreVertical, Phone as PhoneIcon, Settings as SettingsIcon, X as XIcon } from 'lucide-vue-next'
import { deviceTemplatesAPI } from '@/services/api'

const toast = inject('toast')
const templates = ref([])
const isLoading = ref(false)

onMounted(async () => {
  await fetchTemplates()
  document.addEventListener('click', closeDropdown)
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown)
})

const closeDropdown = () => {
  activeDropdown.value = null
}

async function fetchTemplates() {
  isLoading.value = true
  try {
    const response = await deviceTemplatesAPI.list()
    templates.value = (response.data?.data || response.data || []).map(t => ({
      id: t.id,
      name: t.name,
      model: t.model || t.device_model || 'Unknown',
      firmware: t.firmware_version || 'Latest',
      provisions: t.device_count || 0,
      masterTemplate: t.parent_template_name || t.master_template_name || null,
      description: t.description
    }))
  } catch (error) {
    toast?.error(error.message, 'Failed to load templates')
    templates.value = []
  } finally {
    isLoading.value = false
  }
}

const showCreateModal = ref(false)
const newTemplateBase = ref('')
const newTemplateName = ref('')
const newTemplateDesc = ref('')

const modelMap = {
  'yealink_t54': { model: 'Yealink T54W', master: 'Yealink T5 Series Master' },
  'yealink_t57': { model: 'Yealink T57W', master: 'Yealink T5 Series Master' },
  'yealink_w60': { model: 'Yealink W60B', master: null },
  'poly_ccx': { model: 'Poly CCX 500', master: 'Polycom VVX Generic' },
  'poly_vvx': { model: 'Poly VVX 450', master: 'Polycom VVX Generic' },
  'grandstream_gxp': { model: 'Grandstream GXP2100', master: 'Grandstream GXP' }
}

const confirmCreate = async () => {
  if (!newTemplateBase.value || !newTemplateName.value) return
  
  const selected = modelMap[newTemplateBase.value]
  try {
    await deviceTemplatesAPI.create({
      name: newTemplateName.value,
      device_model: selected.model,
      description: newTemplateDesc.value,
      master_template_key: newTemplateBase.value
    })
    toast?.success('Template created successfully')
    await fetchTemplates()
    showCreateModal.value = false
    newTemplateBase.value = ''
    newTemplateName.value = ''
    newTemplateDesc.value = ''
  } catch (error) {
    toast?.error(error.message, 'Failed to create template')
  }
}

const activeDropdown = ref(null)

const toggleDropdown = (id) => {
  activeDropdown.value = activeDropdown.value === id ? null : id
}

const deleteTemplate = async (t) => {
  activeDropdown.value = null
  if (confirm(`Delete "${t.name}"?`)) {
    try {
      await deviceTemplatesAPI.delete(t.id)
      toast?.success(`Template "${t.name}" deleted`)
      await fetchTemplates()
    } catch (error) {
      toast?.error(error.message, 'Failed to delete template')
    }
  }
}

const cloneTemplate = async (t) => {
  activeDropdown.value = null
  try {
    await deviceTemplatesAPI.create({
      name: `${t.name} (Copy)`,
      device_model: t.model,
      description: t.description,
      clone_from_id: t.id
    })
    toast?.success('Template cloned successfully')
    await fetchTemplates()
  } catch (error) {
    toast?.error(error.message, 'Failed to clone template')
  }
}

const exportTemplate = (t) => {
  activeDropdown.value = null
  // TODO: Implement template export endpoint
  toast?.info(`Exporting ${t.name} configuration...`)
}
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-lg);
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.back-link {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  font-size: var(--text-xs);
  text-align: left;
  margin-bottom: 4px;
}
.back-link:hover { text-decoration: underline; color: var(--primary-color); }

.templates-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-lg);
}

.template-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.template-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-md);
  background: var(--bg-app);
  border-bottom: 1px solid var(--border-color);
}

.icon-box {
  width: 40px;
  height: 40px;
  background: white;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-color);
  border: 1px solid var(--border-color);
}

.card-actions {
  position: relative;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-muted);
  padding: 4px;
  border-radius: 4px;
}
.btn-icon:hover { 
  background: white; 
  color: var(--text-primary); 
}

.dropdown-menu {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 4px;
  background: white;
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-md);
  border-radius: var(--radius-sm);
  min-width: 120px;
  z-index: 20;
  overflow: hidden;
}

.dropdown-menu button {
  display: block;
  width: 100%;
  text-align: left;
  padding: 8px 12px;
  border: none;
  background: none;
  font-size: 12px;
  cursor: pointer;
  color: var(--text-main);
}
.dropdown-menu button:hover { background: var(--bg-app); }
.dropdown-menu button.text-bad { color: var(--status-bad); }

.card-body {
  padding: var(--spacing-md);
}

.card-body h3 {
  font-size: var(--text-md);
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--text-primary);
}

.model {
  font-size: var(--text-sm);
  color: var(--text-muted);
  margin-bottom: 12px;
}

.tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.tag {
  font-size: 11px;
  background: var(--bg-app);
  padding: 4px 8px;
  border-radius: 4px;
  color: var(--text-main);
  border: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 4px;
}

.tag-label {
  color: var(--text-muted);
}
.tag-value {
  font-weight: 600;
  font-family: monospace;
}

.tag.devices {
  background: #eef2ff;
  border-color: #c7d2fe;
  color: #4338ca;
}

.tag-icon {
  width: 12px;
  height: 12px;
}

.master-ref {
  display: flex;
  gap: 6px;
  font-size: 11px;
  padding: 8px;
  background: #fefce8;
  border: 1px solid #fde68a;
  border-radius: 4px;
}
.ref-label { color: #92400e; }
.ref-value { color: #78350f; font-weight: 500; }

.card-footer {
  padding: var(--spacing-md);
  border-top: 1px solid var(--border-color);
}

.btn-primary {
  background-color: var(--primary-color);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  font-size: var(--text-sm);
  cursor: pointer;
}
.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  border: 1px solid var(--border-color);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.btn-secondary:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.full-width { width: 100%; }

.btn-icon-left {
  width: 14px;
  height: 14px;
}

.icon { width: 20px; height: 20px; }
.icon-sm { width: 16px; height: 16px; }

/* Modal Styles */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 24px;
}

.modal-card {
  background: white;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 480px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 700;
  margin: 0;
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
}

.input-field {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 14px;
}
.input-field:focus {
  outline: none;
  border-color: var(--primary-color);
}

textarea.input-field {
  resize: vertical;
  min-height: 60px;
}

.help-text {
  font-size: 11px;
  color: var(--text-muted);
}
</style>
