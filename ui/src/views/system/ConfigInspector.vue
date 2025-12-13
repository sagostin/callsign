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
              <span>Type</span>
            </div>
            
            <div v-if="loadingFiles" class="loading-state">
              <LoaderIcon class="spin" /> Loading...
            </div>

            <!-- Dynamic Generated Files Section (only at root) -->
            <div v-if="!currentPath && !loadingFiles" class="dynamic-section">
              <div class="section-label">
                <ZapIcon class="section-icon" /> Dynamic/Generated (from XML CURL)
              </div>
              
              <div 
                v-for="gen in dynamicConfigs" 
                :key="gen.id"
                class="file-item dynamic"
                :class="{ selected: selectedFile?.id === gen.id }"
                @click="loadDynamicConfig(gen)"
              >
                <div class="file-name">
                  <CodeIcon class="file-icon dynamic-icon" />
                  <span>{{ gen.name }}</span>
                </div>
                <span class="file-badge">Generated</span>
              </div>
              
              <div class="section-divider"></div>
              <div class="section-label">
                <HardDriveIcon class="section-icon" /> Static Files (on disk)
              </div>
            </div>

            <div v-if="!loadingFiles && allFiles.length === 0 && currentPath" class="empty-state">
              Directory is empty
            </div>

            <div 
              v-for="file in allFiles" 
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

          <!-- File/Config Content Viewer -->
          <div class="file-viewer" v-if="selectedFile">
            <div class="viewer-header" :class="{ dynamic: selectedFile.isDynamic }">
              <div class="viewer-title">
                <ZapIcon v-if="selectedFile.isDynamic" class="title-icon dynamic" />
                <FileTextIcon v-else class="title-icon" />
                <h4>{{ selectedFile.name }}</h4>
              </div>
              <div class="viewer-actions">
                <button class="btn-icon" @click="copyFileContent" title="Copy">
                  <CopyIcon class="icon-sm" />
                </button>
                <button class="btn-icon" @click="refreshContent" title="Refresh" v-if="selectedFile.isDynamic">
                  <RefreshCwIcon class="icon-sm" />
                </button>
                <button class="btn-icon" @click="closeFile" title="Close">
                  <XIcon class="icon-sm" />
                </button>
              </div>
            </div>
            <div v-if="selectedFile.isDynamic" class="viewer-meta">
              <span class="meta-badge">Live Generated</span>
              <span class="meta-info">{{ selectedFile.description }}</span>
            </div>
            <div class="viewer-content" v-if="loadingContent">
              <LoaderIcon class="spin" /> Loading...
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
  ChevronRight as ChevronRightIcon, Copy as CopyIcon, X as XIcon, Loader as LoaderIcon,
  Zap as ZapIcon, Code as CodeIcon, HardDrive as HardDriveIcon, RefreshCw as RefreshCwIcon
} from 'lucide-vue-next'
import ConfigDebugger from '@/components/ConfigDebugger.vue'
import { systemAPI } from '@/services/api'
import api from '@/services/api'

const toast = inject('toast')
const activeTab = ref('debug')
const currentPath = ref('')
const files = ref([])
const loadingFiles = ref(false)
const selectedFile = ref(null)
const fileContent = ref('')
const loadingContent = ref(false)

// Dynamic/generated configuration endpoints
const dynamicConfigs = [
  // Dialplan Section
  { 
    id: 'dialplan-public', 
    name: 'dialplan/public.xml', 
    description: 'Inbound routing dialplan generated from database',
    section: 'dialplan',
    params: { context: 'public', destination_number: '1001' }
  },
  { 
    id: 'dialplan-default', 
    name: 'dialplan/default.xml', 
    description: 'Internal/outbound dialplan generated from database',
    section: 'dialplan',
    params: { context: 'default', destination_number: '1001' }
  },
  
  // Directory Section
  { 
    id: 'directory-users', 
    name: 'directory/users.xml', 
    description: 'User directory for SIP registration/auth',
    section: 'directory',
    params: { domain: 'example.com', action: 'sip_auth', user: '1001' }
  },
  { 
    id: 'directory-gateways', 
    name: 'directory/gateways.xml', 
    description: 'SIP trunk/gateway registrations',
    section: 'directory',
    params: { domain: 'example.com', purpose: 'gateways' }
  },
  
  // Configuration Section - Core modules
  { 
    id: 'config-acl', 
    name: 'configuration/acl.conf.xml', 
    description: 'Access control lists (network permissions)',
    section: 'configuration',
    params: { key_value: 'acl.conf' }
  },
  { 
    id: 'config-ivr', 
    name: 'configuration/ivr.conf.xml', 
    description: 'IVR menus and auto-attendants',
    section: 'configuration',
    params: { key_value: 'ivr.conf' }
  },
  { 
    id: 'config-conference', 
    name: 'configuration/conference.conf.xml', 
    description: 'Audio/video conference rooms',
    section: 'configuration',
    params: { key_value: 'conference.conf' }
  },
  { 
    id: 'config-local-stream', 
    name: 'configuration/local_stream.conf.xml', 
    description: 'Music on hold and local audio streams',
    section: 'configuration',
    params: { key_value: 'local_stream.conf' }
  },
  { 
    id: 'config-sofia', 
    name: 'configuration/sofia.conf.xml', 
    description: 'Sofia SIP stack (uses static file)',
    section: 'configuration',
    params: { key_value: 'sofia.conf' }
  },
]

const allFiles = computed(() => files.value)

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
    currentPath.value = file.path
    selectedFile.value = null
    fileContent.value = ''
    loadDirectory()
  } else {
    selectedFile.value = { ...file, isDynamic: false }
    loadFileContent(file.path)
  }
}

const loadDynamicConfig = async (config) => {
  selectedFile.value = {
    id: config.id,
    name: config.name,
    description: config.description,
    isDynamic: true,
    section: config.section,
    params: config.params
  }
  await loadDynamicContent(config)
}

const loadDynamicContent = async (config) => {
  loadingContent.value = true
  try {
    const params = {
      section: config.section,
      ...config.params
    }
    const response = await api.get('/system/xml/debug', { params })
    const xml = response.data?.xml || '<no-result/>'
    // Format the XML nicely
    fileContent.value = formatXml(xml)
  } catch (e) {
    console.error('Failed to load dynamic config:', e)
    toast?.error('Failed to generate configuration')
    fileContent.value = '<!-- Failed to generate configuration -->'
  } finally {
    loadingContent.value = false
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

const refreshContent = () => {
  if (selectedFile.value?.isDynamic) {
    loadDynamicContent(selectedFile.value)
  }
}

const closeFile = () => {
  selectedFile.value = null
  fileContent.value = ''
}

const copyFileContent = () => {
  navigator.clipboard.writeText(fileContent.value)
  toast?.success('Content copied to clipboard')
}

const formatXml = (xml) => {
  // Simple XML formatting
  let formatted = ''
  let indent = 0
  const lines = xml.replace(/>\s*</g, '>\n<').split('\n')
  
  for (const line of lines) {
    if (line.match(/^<\/\w/)) indent--
    formatted += '  '.repeat(Math.max(0, indent)) + line.trim() + '\n'
    if (line.match(/^<\w[^>]*[^/]>$/) && !line.match(/^<\?/)) indent++
  }
  
  return formatted.trim()
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
  max-height: 550px;
  overflow-y: auto;
}
.file-list.narrow { max-width: 320px; flex: none; }

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
  z-index: 1;
}

/* Dynamic Section */
.dynamic-section {
  background: #fefce8;
  border-bottom: 1px solid #fef08a;
}
.section-label {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: #92400e;
  background: #fef9c3;
}
.section-label:last-of-type {
  background: var(--bg-app);
  color: var(--text-muted);
  border-top: 1px solid var(--border-color);
}
.section-icon { width: 12px; height: 12px; }
.section-divider { height: 8px; background: white; }

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
.file-item.dynamic { background: #fffbeb; }
.file-item.dynamic:hover { background: #fef3c7; }
.file-item.dynamic.selected { background: #fde68a; }

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
.file-icon.dynamic-icon { color: #d97706; }

.file-size { font-size: 11px; color: var(--text-muted); flex-shrink: 0; }
.file-badge {
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  padding: 2px 6px;
  background: #fef08a;
  color: #92400e;
  border-radius: 4px;
}

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
  max-height: 550px;
}

.viewer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #252526;
  border-bottom: 1px solid #333;
}
.viewer-header.dynamic { background: #78350f; }

.viewer-title {
  display: flex;
  align-items: center;
  gap: 8px;
}
.title-icon { width: 14px; height: 14px; color: #aaa; }
.title-icon.dynamic { color: #fbbf24; }
.viewer-header h4 {
  margin: 0;
  font-size: 12px;
  font-weight: 500;
  color: #fff;
}
.viewer-actions { display: flex; gap: 4px; }

.viewer-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #fef9c3;
  border-bottom: 1px solid #fef08a;
  font-size: 11px;
}
.meta-badge {
  padding: 2px 6px;
  background: #fde68a;
  color: #92400e;
  border-radius: 4px;
  font-weight: 600;
}
.meta-info { color: #78350f; }

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
