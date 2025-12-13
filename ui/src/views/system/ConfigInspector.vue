<template>
  <div class="view-container">
    <div class="view-header">
      <h2>System Configuration Inspector</h2>
      <p class="text-muted">Debug and inspect generated FreeSWITCH XML configuration.</p>
    </div>

    <!-- Tab Navigation -->
    <div class="tab-nav">
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'debug' }" 
        @click="activeTab = 'debug'"
      >
        <WrenchIcon class="tab-icon" />
        XML Debugger
      </button>
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'files' }" 
        @click="activeTab = 'files'; loadDirectory()"
      >
        <FolderIcon class="tab-icon" />
        Config Files
      </button>
    </div>

    <!-- Debug Tab -->
    <div v-if="activeTab === 'debug'" class="tab-content">
      <ConfigDebugger mode="system" />
    </div>

    <!-- Files Tab -->
    <div v-if="activeTab === 'files'" class="tab-content">
      <div class="file-browser">
        <!-- Breadcrumb -->
        <div class="breadcrumb">
          <button class="crumb" @click="navigateTo('')">/etc/freeswitch</button>
          <template v-for="(segment, idx) in pathSegments" :key="idx">
            <ChevronRightIcon class="crumb-sep" />
            <button class="crumb" @click="navigateTo(getPathUpTo(idx))">{{ segment }}</button>
          </template>
        </div>

        <div class="browser-layout">
          <!-- Directory Listing -->
          <div class="file-list" :class="{ narrow: selectedFile }">
            <div class="list-header">
              <span>Name</span>
              <span>Size</span>
            </div>
            
            <div v-if="loadingFiles" class="loading-state">
              <LoaderIcon class="spin" /> Loading...
            </div>

            <div v-else-if="files.length === 0" class="empty-state">
              Directory is empty
            </div>

            <div 
              v-for="file in files" 
              :key="file.path"
              class="file-item"
              :class="{ selected: selectedFile?.path === file.path, directory: file.is_dir }"
              @click="handleFileClick(file)"
            >
              <div class="file-name">
                <FolderIcon v-if="file.is_dir" class="file-icon folder" />
                <FileTextIcon v-else class="file-icon" />
                <span>{{ file.name }}</span>
              </div>
              <span class="file-size" v-if="!file.is_dir">{{ formatSize(file.size) }}</span>
            </div>
          </div>

          <!-- File Content Viewer -->
          <div class="file-viewer" v-if="selectedFile && !selectedFile.is_dir">
            <div class="viewer-header">
              <h4>{{ selectedFile.name }}</h4>
              <div class="viewer-actions">
                <button class="btn-icon" @click="copyFileContent" title="Copy">
                  <CopyIcon class="icon-sm" />
                </button>
                <button class="btn-icon" @click="closeFile" title="Close">
                  <XIcon class="icon-sm" />
                </button>
              </div>
            </div>
            <div class="viewer-content" v-if="loadingContent">
              <LoaderIcon class="spin" /> Loading file...
            </div>
            <div class="viewer-content" v-else>
              <pre><code>{{ fileContent }}</code></pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { 
  Wrench as WrenchIcon, Folder as FolderIcon, FileText as FileTextIcon,
  ChevronRight as ChevronRightIcon, Copy as CopyIcon, X as XIcon, Loader as LoaderIcon
} from 'lucide-vue-next'
import ConfigDebugger from '@/components/ConfigDebugger.vue'
import { systemAPI } from '@/services/api'

const toast = inject('toast')
const activeTab = ref('debug')
const currentPath = ref('')
const files = ref([])
const loadingFiles = ref(false)
const selectedFile = ref(null)
const fileContent = ref('')
const loadingContent = ref(false)

const pathSegments = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/').filter(Boolean)
})

const getPathUpTo = (idx) => {
  return pathSegments.value.slice(0, idx + 1).join('/')
}

const navigateTo = (path) => {
  currentPath.value = path
  selectedFile.value = null
  fileContent.value = ''
  loadDirectory()
}

const loadDirectory = async () => {
  loadingFiles.value = true
  try {
    const response = await systemAPI.listConfigFiles(currentPath.value)
    files.value = response.data?.files || []
  } catch (e) {
    console.error('Failed to load directory:', e)
    toast?.error('Failed to load directory')
    files.value = []
  } finally {
    loadingFiles.value = false
  }
}

const handleFileClick = async (file) => {
  if (file.is_dir) {
    // Navigate into directory
    currentPath.value = file.path
    selectedFile.value = null
    fileContent.value = ''
    loadDirectory()
  } else {
    // Select file and load content
    selectedFile.value = file
    loadFileContent(file.path)
  }
}

const loadFileContent = async (path) => {
  loadingContent.value = true
  try {
    const response = await systemAPI.readConfigFile(path)
    fileContent.value = response.data?.content || ''
  } catch (e) {
    console.error('Failed to read file:', e)
    toast?.error('Failed to read file')
    fileContent.value = 'Failed to load file content'
  } finally {
    loadingContent.value = false
  }
}

const closeFile = () => {
  selectedFile.value = null
  fileContent.value = ''
}

const copyFileContent = () => {
  navigator.clipboard.writeText(fileContent.value)
  toast?.success('File content copied')
}

const formatSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<style scoped>
.view-container {
  max-width: 1400px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 16px;
}
.view-header h2 { margin: 0 0 4px; }
.text-muted { color: var(--text-muted); font-size: 13px; }

/* Tabs */
.tab-nav {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 20px;
}
.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border: none;
  background: none;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: all 0.2s;
}
.tab-btn:hover { color: var(--text-main); }
.tab-btn.active {
  color: var(--primary-color);
  border-bottom-color: var(--primary-color);
}
.tab-icon { width: 16px; height: 16px; }

.tab-content {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

/* File Browser */
.file-browser {
  padding: 16px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 10px 12px;
  background: var(--bg-app);
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
}
.crumb {
  background: none;
  border: none;
  color: var(--primary-color);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
}
.crumb:hover { background: #e0e7ff; }
.crumb-sep { width: 12px; height: 12px; color: var(--text-muted); }

.browser-layout {
  display: flex;
  gap: 16px;
}

.file-list {
  flex: 1;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  overflow: hidden;
  max-height: 500px;
  overflow-y: auto;
}
.file-list.narrow { max-width: 300px; flex: none; }

.list-header {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-app);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid var(--border-light);
  font-size: 13px;
  transition: background 0.15s;
}
.file-item:hover { background: var(--bg-app); }
.file-item.selected { background: #e0e7ff; }
.file-item.directory { font-weight: 500; }

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.file-name span { 
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.file-icon { width: 16px; height: 16px; color: var(--text-muted); flex-shrink: 0; }
.file-icon.folder { color: #f59e0b; }

.file-size { font-size: 11px; color: var(--text-muted); flex-shrink: 0; }

.loading-state, .empty-state {
  padding: 24px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* File Viewer */
.file-viewer {
  flex: 2;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 500px;
}

.viewer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #252526;
  border-bottom: 1px solid #333;
}
.viewer-header h4 {
  margin: 0;
  font-size: 12px;
  font-weight: 500;
  color: #fff;
}
.viewer-actions { display: flex; gap: 4px; }

.viewer-content {
  flex: 1;
  overflow: auto;
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
  line-height: 1.6;
}
.viewer-content pre { margin: 0; white-space: pre-wrap; word-break: break-all; }

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: #aaa;
  padding: 4px;
}
.btn-icon:hover { color: #fff; }
.icon-sm { width: 14px; height: 14px; }
</style>
