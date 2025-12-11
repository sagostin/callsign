<template>
  <div class="view-header">
    <div class="header-content">
      <h2>Music on Hold</h2>
      <p class="text-muted text-sm">Manage hold music classes and streams.</p>
    </div>
    <div class="header-actions">
      <button class="btn-primary" @click="$router.push('/admin/music-on-hold/new')">+ Add Stream</button>
    </div>
  </div>

  <div class="tabs">
    <button 
        v-for="rate in rates" 
        :key="rate"
        class="tab" 
        :class="{ active: activeTab === rate }" 
        @click="activeTab = rate"
    >
        {{ rate }} Hz
    </button>
  </div>

  <div class="tab-content">
      <div class="action-bar" style="margin-bottom: 16px;">
          <p class="text-xs text-muted">Files in <code>/usr/share/freeswitch/sounds/music/{{ activeTab }}</code></p>
          <button class="btn-primary small" @click="showUploadModal = true">
             <UploadIcon class="icon-small" style="width:12px; margin-right:4px;" /> Upload Music
          </button>
      </div>

      <DataTable :columns="columns" :data="currentFiles" actions>
        <template #name="{ item }">
            <span class="font-mono text-sm">{{ item.name }}</span>
            <span v-if="item.isOverride" class="override-badge">Override</span>
        </template>
        <template #size="{ value }">
            <span class="font-mono text-xs">{{ (value / 1024).toFixed(1) }} KB</span>
        </template>
        <template #actions="{ item }">
          <button v-if="isTenant && item.isOverride" class="btn-link text-bad" @click="deleteFile(item)">Revert</button>
        </template>
      </DataTable>
  </div>

  <!-- Upload Modal -->
  <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Upload Music ({{ activeTab }} Hz)</h3>
          <button class="close-btn" @click="showUploadModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Audio File (WAV/MP3)</label>
            <input type="file" @change="handleFileUpload" accept=".wav,.mp3,.ogg" class="input-field">
          </div>
        </div>
        <div class="modal-footer">
           <button class="btn-secondary" @click="showUploadModal = false">Cancel</button>
           <button class="btn-primary" @click="submitUpload">Upload</button>
        </div>
      </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { systemAPI, tenantMediaAPI } from '../services/api'
import DataTable from '../components/common/DataTable.vue'
import { Upload as UploadIcon, Download as DownloadIcon } from 'lucide-vue-next'

const activeTab = ref('8000') // Default rate
const musicData = ref([])
const isLoading = ref(false)
const showUploadModal = ref(false)
const uploadForm = ref({ file: null })

// Computed rates available in the FS
const rates = computed(() => {
    if (!musicData.value.length) return ['8000', '16000', '32000', '48000']
    
    // Extract folder names that are numbers
    const folders = musicData.value
        .filter(n => n.type === 'directory' && !isNaN(parseInt(n.name)))
        .map(n => n.name)
        .sort((a, b) => parseInt(a) - parseInt(b))
        
    return folders.length ? folders : ['8000', '16000', '32000', '48000']
})

const isTenant = computed(() => !!localStorage.getItem('tenantId'))

// Get files for active tab
const currentFiles = computed(() => {
    const rateNode = musicData.value.find(n => n.name === activeTab.value)
    if (!rateNode || !rateNode.children) return []
    
    return rateNode.children.filter(n => n.type === 'file').map(f => ({
        name: f.name,
        path: f.path,
        size: f.size,
        rate: activeTab.value,
        isOverride: f.is_override
    }))
})

const columns = [
  { key: 'name', label: 'Filename', width: '300px' },
  { key: 'size', label: 'Size (bytes)' },
  { key: 'path', label: 'Full Path' }
]

const loadMusic = async () => {
    isLoading.value = true
    try {
        const apiCall = isTenant.value ? tenantMediaAPI.listMusic : systemAPI.listMusic
        const response = await apiCall()
        musicData.value = response.data.data
        // Set active tab to first available if not set or invalid
        if (rates.value.length && !rates.value.includes(activeTab.value)) {
            activeTab.value = rates.value[0]
        }
    } catch (e) {
        console.error('Failed to load music', e)
    } finally {
        isLoading.value = false
    }
}

const handleFileUpload = (event) => {
    uploadForm.value.file = event.target.files[0]
}

const submitUpload = async () => {
    if (!uploadForm.value.file) return
    
    const formData = new FormData()
    formData.append('file', uploadForm.value.file)
    formData.append('rate', activeTab.value)
    
    try {
        if (isTenant.value) {
            await tenantMediaAPI.uploadMusic(formData)
        } else {
            await systemAPI.uploadMusic(formData)
        }
        showUploadModal.value = false
        loadMusic()
    } catch (e) {
        console.error('Upload failed', e)
        alert('Upload failed: ' + (e.response?.data?.error || e.message))
    }
}

const deleteFile = async (file) => {
    if (isTenant.value) {
        if (!confirm(`Delete/Revert ${file.name}?`)) return
        try {
            await tenantMediaAPI.deleteMusic(file.path)
            loadMusic()
        } catch (e) {
            console.error('Delete failed', e)
            alert('Failed to delete file')
        }
    } else {
        // System admin delete not implemented yet/safe
        console.log('System delete restricted for safety', file)
    }
}

onMounted(() => {
    loadMusic()
})
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
.btn-primary.small { padding: 6px 12px; font-size: 12px; }

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

.font-mono { font-family: monospace; }
.text-xs { font-size: 11px; }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { padding: 8px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 24px; border-radius: 0 0 4px 4px; }

/* Playlists */
.action-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.playlist-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 24px; }

.playlist-card { 
   border: 1px solid var(--border-color); border-radius: 8px; 
   background: white; overflow: hidden; display: flex; flex-direction: column;
}

.playlist-header { padding: 16px; border-bottom: 1px solid var(--border-color); display: flex; justify-content: space-between; align-items: flex-start; background: #f8fafc; }
.pl-name { font-weight: 600; font-size: 14px; display: block; color: var(--text-primary); }
.pl-count { font-size: 11px; color: var(--text-muted); }
.pl-more { color: var(--text-muted); cursor: pointer; font-weight: bold; letter-spacing: 1px; }

.track-list { padding: 16px; flex: 1; }
.track { display: flex; gap: 12px; align-items: center; margin-bottom: 12px; }
.track:last-child { margin-bottom: 0; }
.track-icon { width: 24px; height: 24px; background: var(--bg-secondary); border-radius: 4px; display: flex; align-items: center; justify-content: center; font-size: 10px; }
.track-details { flex: 1; min-width: 0; }
.track-name { font-size: 12px; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.track-dur { font-size: 10px; color: var(--text-muted); }

.pl-footer { padding: 12px; border-top: 1px solid var(--border-color); background: #f8fafc; }
.btn-secondary { background: white; border: 1px solid var(--border-color); color: var(--text-main); font-size: 12px; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-weight: 500; }
.full-width { width: 100%; }

/* Modal Styles */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 400px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }

.override-badge { 
    font-size: 10px; font-weight: 700; color: #d97706; background: #fef3c7; 
    padding: 2px 6px; border-radius: 4px; border: 1px solid #fcd34d; margin-left: 8px;
    text-transform: uppercase;
}
</style>
