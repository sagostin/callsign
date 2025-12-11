<template>
  <div class="firmware-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Firmware Management</h2>
        <p class="text-muted text-sm">Manage device firmware versions and deployment policies.</p>
      </div>
      <div class="header-actions">
        <button class="btn-primary" @click="showUploadModal = true">
          <UploadIcon class="btn-icon" /> Upload Firmware
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ manufacturers.length }}</div>
        <div class="stat-label">Manufacturers</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ totalFirmware }}</div>
        <div class="stat-label">Firmware Files</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ activeDeployments }}</div>
        <div class="stat-label">Active Deployments</div>
      </div>
    </div>

    <!-- Manufacturer Tabs -->
    <div class="manufacturer-tabs">
      <button 
        v-for="mfg in manufacturers" 
        :key="mfg.id"
        class="mfg-tab"
        :class="{ active: selectedManufacturer === mfg.id }"
        @click="selectedManufacturer = mfg.id"
      >
        <MonitorIcon class="tab-icon" />
        <span>{{ mfg.name }}</span>
      </button>
    </div>

    <!-- Firmware Content -->
    <div class="firmware-content">
      <!-- Model Groups -->
      <div class="model-groups">
        <div v-for="model in modelsForManufacturer" :key="model.id" class="model-group">
          <div class="group-header" @click="toggleModel(model.id)">
            <div class="group-info">
              <SmartphoneIcon class="model-icon" />
              <div>
                <h4>{{ model.name }}</h4>
                <span class="model-meta">{{ getFirmwareCount(model.id) }} versions available</span>
              </div>
            </div>
            <div class="group-actions">
              <span v-if="model.recommended" class="recommended-badge">
                <CheckIcon /> v{{ model.recommended }}
              </span>
              <ChevronDownIcon class="expand-icon" :class="{ expanded: expandedModels.includes(model.id) }" />
            </div>
          </div>
          
          <transition name="slide">
            <div v-if="expandedModels.includes(model.id)" class="group-content">
              <!-- Version Table -->
              <table class="version-table">
                <thead>
                  <tr>
                    <th>Version</th>
                    <th>Release Date</th>
                    <th>Size</th>
                    <th>Status</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="fw in getFirmwareForModel(model.id)" :key="fw.id" :class="{ recommended: fw.isRecommended }">
                    <td>
                      <div class="version-cell">
                        <span class="version-num">{{ fw.version }}</span>
                        <span v-if="fw.isRecommended" class="rec-tag">Recommended</span>
                        <span v-if="fw.isBeta" class="beta-tag">Beta</span>
                      </div>
                    </td>
                    <td>{{ fw.releaseDate }}</td>
                    <td class="mono">{{ fw.size }}</td>
                    <td>
                      <span class="status-badge" :class="fw.status.toLowerCase()">{{ fw.status }}</span>
                    </td>
                    <td class="actions-cell">
                      <button class="btn-sm" @click="setRecommended(model.id, fw)">
                        <StarIcon :class="{ filled: fw.isRecommended }" /> Set Default
                      </button>
                      <button class="btn-sm" @click="viewReleaseNotes(fw)">
                        <FileTextIcon /> Notes
                      </button>
                      <button class="btn-sm danger" @click="deleteFirmware(fw)">
                        <TrashIcon />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
              
              <!-- Deployment Policy -->
              <div class="deployment-section">
                <h5>Deployment Policy</h5>
                <div class="policy-options">
                  <label class="policy-option">
                    <input type="radio" :name="'policy-' + model.id" value="manual" v-model="model.deployPolicy">
                    <div class="policy-content">
                      <strong>Manual</strong>
                      <span>Devices update only when manually triggered</span>
                    </div>
                  </label>
                  <label class="policy-option">
                    <input type="radio" :name="'policy-' + model.id" value="recommended" v-model="model.deployPolicy">
                    <div class="policy-content">
                      <strong>Auto-update to Recommended</strong>
                      <span>Devices automatically update to the recommended version</span>
                    </div>
                  </label>
                  <label class="policy-option">
                    <input type="radio" :name="'policy-' + model.id" value="latest" v-model="model.deployPolicy">
                    <div class="policy-content">
                      <strong>Auto-update to Latest</strong>
                      <span>Devices automatically update to the newest version (excluding beta)</span>
                    </div>
                  </label>
                </div>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>

    <!-- Upload Modal -->
    <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Upload Firmware</h3>
          <button class="close-btn" @click="showUploadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Manufacturer</label>
            <select v-model="uploadForm.manufacturer" class="input-field">
              <option v-for="mfg in manufacturers" :key="mfg.id" :value="mfg.id">{{ mfg.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>Model</label>
            <select v-model="uploadForm.model" class="input-field">
              <option v-for="m in getModelsForMfg(uploadForm.manufacturer)" :key="m.id" :value="m.id">{{ m.name }}</option>
            </select>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Version</label>
              <input v-model="uploadForm.version" class="input-field code" placeholder="96.3.0.5">
            </div>
            <div class="form-group">
              <label>Release Type</label>
              <select v-model="uploadForm.releaseType" class="input-field">
                <option value="stable">Stable</option>
                <option value="beta">Beta</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label>Release Notes</label>
            <textarea v-model="uploadForm.notes" class="input-field" rows="3" placeholder="Bug fixes and improvements..."></textarea>
          </div>
          <div class="form-group">
            <label>Firmware File</label>
            <div class="file-upload">
              <input type="file" id="firmware-file" accept=".rom,.bin,.fw,.zip">
              <label for="firmware-file" class="file-label">
                <UploadIcon class="upload-icon" />
                <span>Choose file or drag & drop</span>
                <span class="file-hint">.rom, .bin, .fw, .zip</span>
              </label>
            </div>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="uploadForm.setAsRecommended">
              Set as recommended version after upload
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
          <button class="btn-primary">Upload Firmware</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Upload as UploadIcon, Monitor as MonitorIcon, Smartphone as SmartphoneIcon,
  ChevronDown as ChevronDownIcon, Check as CheckIcon, Star as StarIcon,
  FileText as FileTextIcon, Trash2 as TrashIcon
} from 'lucide-vue-next'

const selectedManufacturer = ref('yealink')
const expandedModels = ref(['t5-series'])
const showUploadModal = ref(false)

const uploadForm = ref({
  manufacturer: 'yealink',
  model: '',
  version: '',
  releaseType: 'stable',
  notes: '',
  setAsRecommended: false
})

const manufacturers = ref([
  { id: 'yealink', name: 'Yealink' },
  { id: 'poly', name: 'Poly' },
  { id: 'cisco', name: 'Cisco' },
  { id: 'grandstream', name: 'Grandstream' },
  { id: 'fanvil', name: 'Fanvil' },
])

const models = ref([
  { id: 't5-series', manufacturer: 'yealink', name: 'T5 Series (T54W/T57W)', recommended: '96.3.0.5', deployPolicy: 'recommended' },
  { id: 't4-series', manufacturer: 'yealink', name: 'T4 Series (T46U/T48U)', recommended: '83.15.0.20', deployPolicy: 'manual' },
  { id: 'cp-series', manufacturer: 'yealink', name: 'CP Series (CP920/CP960)', recommended: '73.15.0.10', deployPolicy: 'recommended' },
  { id: 'vvx-series', manufacturer: 'poly', name: 'VVX Series (VVX350/VVX450)', recommended: '6.4.3.2', deployPolicy: 'recommended' },
  { id: 'ccx-series', manufacturer: 'poly', name: 'CCX Series (CCX500/CCX600)', recommended: '7.2.1.0', deployPolicy: 'manual' },
  { id: 'spa-series', manufacturer: 'cisco', name: 'SPA Series', recommended: '7.6.2c', deployPolicy: 'manual' },
])

const firmwareVersions = ref([
  { id: 1, modelId: 't5-series', version: '96.3.0.5', releaseDate: '2024-03-15', size: '48.2 MB', status: 'Active', isRecommended: true, isBeta: false },
  { id: 2, modelId: 't5-series', version: '96.3.0.4', releaseDate: '2024-02-20', size: '47.8 MB', status: 'Active', isRecommended: false, isBeta: false },
  { id: 3, modelId: 't5-series', version: '96.3.0.3', releaseDate: '2024-01-10', size: '47.5 MB', status: 'Archived', isRecommended: false, isBeta: false },
  { id: 4, modelId: 't5-series', version: '96.4.0.0-beta', releaseDate: '2024-03-20', size: '49.1 MB', status: 'Testing', isRecommended: false, isBeta: true },
  { id: 5, modelId: 't4-series', version: '83.15.0.20', releaseDate: '2024-03-01', size: '42.1 MB', status: 'Active', isRecommended: true, isBeta: false },
  { id: 6, modelId: 't4-series', version: '83.15.0.18', releaseDate: '2024-01-15', size: '41.8 MB', status: 'Active', isRecommended: false, isBeta: false },
  { id: 7, modelId: 'vvx-series', version: '6.4.3.2', releaseDate: '2024-02-28', size: '38.5 MB', status: 'Active', isRecommended: true, isBeta: false },
  { id: 8, modelId: 'vvx-series', version: '6.4.3.1', releaseDate: '2024-01-20', size: '38.2 MB', status: 'Active', isRecommended: false, isBeta: false },
])

const modelsForManufacturer = computed(() => {
  return models.value.filter(m => m.manufacturer === selectedManufacturer.value)
})

const totalFirmware = computed(() => firmwareVersions.value.length)
const activeDeployments = computed(() => models.value.filter(m => m.deployPolicy !== 'manual').length)

const getFirmwareCount = (modelId) => firmwareVersions.value.filter(f => f.modelId === modelId).length
const getFirmwareForModel = (modelId) => firmwareVersions.value.filter(f => f.modelId === modelId)
const getModelsForMfg = (mfgId) => models.value.filter(m => m.manufacturer === mfgId)

const toggleModel = (modelId) => {
  const idx = expandedModels.value.indexOf(modelId)
  if (idx === -1) expandedModels.value.push(modelId)
  else expandedModels.value.splice(idx, 1)
}

const setRecommended = (modelId, fw) => {
  firmwareVersions.value.filter(f => f.modelId === modelId).forEach(f => f.isRecommended = false)
  fw.isRecommended = true
  const model = models.value.find(m => m.id === modelId)
  if (model) model.recommended = fw.version
}

const viewReleaseNotes = (fw) => { console.log('View notes for', fw.version) }
const deleteFirmware = (fw) => {
  if (confirm(`Delete firmware version ${fw.version}?`)) {
    firmwareVersions.value = firmwareVersions.value.filter(f => f.id !== fw.id)
  }
}
</script>

<style scoped>
.firmware-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

/* Manufacturer Tabs */
.manufacturer-tabs { display: flex; gap: 4px; margin-bottom: 20px; }
.mfg-tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: white; border: 1px solid var(--border-color); border-radius: 8px; font-size: 13px; font-weight: 500; cursor: pointer; }
.mfg-tab:hover { border-color: var(--primary-color); }
.mfg-tab.active { background: var(--primary-color); color: white; border-color: var(--primary-color); }
.tab-icon { width: 16px; height: 16px; }

/* Firmware Content */
.firmware-content { background: white; border: 1px solid var(--border-color); border-radius: 8px; }

.model-group { border-bottom: 1px solid var(--border-color); }
.model-group:last-child { border-bottom: none; }
.group-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; cursor: pointer; }
.group-header:hover { background: #f8fafc; }
.group-info { display: flex; align-items: center; gap: 12px; }
.model-icon { width: 24px; height: 24px; color: var(--text-muted); }
.group-info h4 { margin: 0; font-size: 14px; }
.model-meta { font-size: 11px; color: var(--text-muted); }
.group-actions { display: flex; align-items: center; gap: 12px; }
.recommended-badge { display: flex; align-items: center; gap: 4px; font-size: 11px; font-weight: 600; color: #16a34a; background: #dcfce7; padding: 4px 10px; border-radius: 4px; }
.recommended-badge svg { width: 12px; height: 12px; }
.expand-icon { width: 20px; height: 20px; color: var(--text-muted); transition: transform 0.2s; }
.expand-icon.expanded { transform: rotate(180deg); }
.group-content { padding: 0 20px 20px; }

/* Version Table */
.version-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.version-table th { text-align: left; padding: 10px 12px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); background: #f8fafc; }
.version-table td { padding: 12px; border-bottom: 1px solid var(--border-color); }
.version-table tr:hover { background: #f8fafc; }
.version-table tr.recommended { background: #f0fdf4; }
.version-cell { display: flex; align-items: center; gap: 8px; }
.version-num { font-family: monospace; font-weight: 600; }
.rec-tag { font-size: 9px; font-weight: 600; background: #16a34a; color: white; padding: 2px 6px; border-radius: 3px; }
.beta-tag { font-size: 9px; font-weight: 600; background: #f59e0b; color: white; padding: 2px 6px; border-radius: 3px; }
.mono { font-family: monospace; }
.status-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; }
.status-badge.active { background: #dcfce7; color: #16a34a; }
.status-badge.testing { background: #fef3c7; color: #d97706; }
.status-badge.archived { background: #f1f5f9; color: #64748b; }
.actions-cell { display: flex; gap: 6px; }
.btn-sm { display: flex; align-items: center; gap: 4px; padding: 5px 8px; background: white; border: 1px solid var(--border-color); border-radius: 4px; font-size: 11px; cursor: pointer; }
.btn-sm:hover { border-color: var(--primary-color); color: var(--primary-color); }
.btn-sm.danger:hover { border-color: #ef4444; color: #ef4444; }
.btn-sm svg { width: 12px; height: 12px; }
.btn-sm svg.filled { fill: currentColor; }

/* Deployment Section */
.deployment-section { margin-top: 20px; padding: 16px; background: #f8fafc; border-radius: 8px; }
.deployment-section h5 { font-size: 12px; text-transform: uppercase; color: var(--text-muted); margin: 0 0 12px; }
.policy-options { display: flex; flex-direction: column; gap: 8px; }
.policy-option { display: flex; align-items: flex-start; gap: 10px; padding: 10px 12px; background: white; border: 1px solid var(--border-color); border-radius: 6px; cursor: pointer; }
.policy-option:has(input:checked) { border-color: var(--primary-color); background: #f0f9ff; }
.policy-option input { margin-top: 2px; }
.policy-content { display: flex; flex-direction: column; }
.policy-content strong { font-size: 13px; }
.policy-content span { font-size: 11px; color: var(--text-muted); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 500px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.form-row { display: flex; gap: 12px; }
.form-row .form-group { flex: 1; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.input-field.code { font-family: monospace; }
.file-upload { border: 2px dashed var(--border-color); border-radius: 8px; padding: 32px; text-align: center; }
.file-upload input { display: none; }
.file-label { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); cursor: pointer; }
.upload-icon { width: 24px; height: 24px; }
.file-hint { font-size: 10px; }
.checkbox-label { display: flex !important; align-items: center; gap: 8px; font-size: 12px !important; cursor: pointer; text-transform: none !important; }

/* Transitions */
.slide-enter-active, .slide-leave-active { transition: all 0.3s ease; }
.slide-enter-from, .slide-leave-to { opacity: 0; max-height: 0; }
</style>
