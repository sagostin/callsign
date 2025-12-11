<template>
  <div class="fax-page">
    <div class="view-header">
      <div class="header-content">
        <h2>My Faxes</h2>
        <p class="text-muted text-sm">Send and receive faxes from your assigned fax number.</p>
      </div>
      <button class="btn-primary" @click="showSendModal = true">
        <SendIcon class="btn-icon" />
        Send New Fax
      </button>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon inbox"><InboxIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ inboxData.length }}</span>
          <span class="stat-label">Inbox</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon sent"><SendIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ sentData.filter(f => f.status === 'Sent').length }}</span>
          <span class="stat-label">Sent</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon pending"><ClockIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ pendingFaxes.length }}</span>
          <span class="stat-label">Pending</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon fax-number"><PhoneIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value mono">(415) 555-3299</span>
          <span class="stat-label">Your Fax Number</span>
        </div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'inbox' }" @click="activeTab = 'inbox'">
        <InboxIcon class="tab-icon" /> Inbox
        <span class="tab-badge" v-if="unreadCount">{{ unreadCount }}</span>
      </button>
      <button class="tab" :class="{ active: activeTab === 'sent' }" @click="activeTab = 'sent'">
        <SendIcon class="tab-icon" /> Sent
      </button>
      <button class="tab" :class="{ active: activeTab === 'pending' }" @click="activeTab = 'pending'">
        <ClockIcon class="tab-icon" /> Pending
        <span class="tab-badge" v-if="pendingFaxes.length">{{ pendingFaxes.length }}</span>
      </button>
    </div>

    <!-- INBOX TAB -->
    <div class="tab-content" v-if="activeTab === 'inbox'">
      <div class="fax-list">
        <div class="fax-item" v-for="fax in inboxData" :key="fax.id" :class="{ unread: !fax.read }">
          <div class="fax-icon incoming">
            <FileDownIcon class="icon-sm" />
          </div>
          <div class="fax-main">
            <div class="fax-header">
              <span class="fax-from">{{ fax.from }}</span>
              <span class="fax-date">{{ fax.date }}</span>
            </div>
            <div class="fax-meta">
              <span class="meta-item"><FileIcon class="meta-icon" /> {{ fax.pages }} pages</span>
              <span class="meta-item"><InboxIcon class="meta-icon" /> {{ fax.box }}</span>
            </div>
          </div>
          <div class="fax-actions">
            <button class="action-btn" @click="viewFax(fax)" title="View"><EyeIcon class="icon-sm" /></button>
            <button class="action-btn" @click="downloadFax(fax)" title="Download"><DownloadIcon class="icon-sm" /></button>
            <button class="action-btn" @click="forwardFax(fax)" title="Forward"><ForwardIcon class="icon-sm" /></button>
            <button class="action-btn danger" @click="deleteFax(fax)" title="Delete"><TrashIcon class="icon-sm" /></button>
          </div>
        </div>
      </div>
    </div>

    <!-- SENT TAB -->
    <div class="tab-content" v-else-if="activeTab === 'sent'">
      <div class="fax-list">
        <div class="fax-item" v-for="fax in sentData" :key="fax.id">
          <div class="fax-icon outgoing">
            <FileUpIcon class="icon-sm" />
          </div>
          <div class="fax-main">
            <div class="fax-header">
              <span class="fax-to">To: {{ fax.to }}</span>
              <span class="fax-date">{{ fax.date }}</span>
            </div>
            <div class="fax-meta">
              <span class="meta-item"><FileIcon class="meta-icon" /> {{ fax.pages }} pages</span>
              <span class="status-badge" :class="fax.status.toLowerCase()">{{ fax.status }}</span>
            </div>
          </div>
          <div class="fax-actions">
            <button class="action-btn" @click="viewFax(fax)" title="View"><EyeIcon class="icon-sm" /></button>
            <button class="action-btn" @click="resendFax(fax)" title="Resend" v-if="fax.status === 'Failed'"><RefreshCwIcon class="icon-sm" /></button>
            <button class="action-btn danger" @click="deleteFax(fax)" title="Delete"><TrashIcon class="icon-sm" /></button>
          </div>
        </div>
      </div>
    </div>

    <!-- PENDING TAB -->
    <div class="tab-content" v-else-if="activeTab === 'pending'">
      <div class="fax-list" v-if="pendingFaxes.length">
        <div class="fax-item pending" v-for="fax in pendingFaxes" :key="fax.id">
          <div class="fax-icon processing">
            <LoaderIcon class="icon-sm spinning" />
          </div>
          <div class="fax-main">
            <div class="fax-header">
              <span class="fax-to">To: {{ fax.to }}</span>
              <span class="fax-status">{{ fax.statusText }}</span>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: fax.progress + '%' }"></div>
            </div>
            <div class="fax-meta">
              <span class="meta-item"><FileIcon class="meta-icon" /> {{ fax.pages }} pages</span>
              <span class="meta-item">{{ fax.progress }}% complete</span>
            </div>
          </div>
          <div class="fax-actions">
            <button class="action-btn danger" @click="cancelFax(fax)" title="Cancel"><XIcon class="icon-sm" /></button>
          </div>
        </div>
      </div>
      <div class="empty-state" v-else>
        <ClockIcon class="empty-icon" />
        <p>No pending faxes</p>
      </div>
    </div>

    <!-- SEND FAX MODAL -->
    <div v-if="showSendModal" class="modal-overlay" @click.self="closeSendModal">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>Send New Fax</h3>
          <button class="btn-icon" @click="closeSendModal"><XIcon class="icon-sm" /></button>
        </div>
        
        <div class="modal-body send-fax-body">
          <!-- Left: Form -->
          <div class="send-form">
            <div class="form-group">
              <label>Recipient Fax Number *</label>
              <input v-model="sendForm.to" class="input-field" placeholder="(555) 555-1234">
            </div>

            <div class="form-group">
              <label>From (Caller ID)</label>
              <select v-model="sendForm.from" class="input-field">
                <option value="(415) 555-3299">(415) 555-3299 - My Fax</option>
                <option value="(415) 555-0100">(415) 555-0100 - Main Office</option>
              </select>
            </div>

            <div class="form-group">
              <label>Cover Page</label>
              <select v-model="sendForm.coverPage" class="input-field">
                <option value="none">No Cover Page</option>
                <option value="standard">Standard Cover</option>
                <option value="confidential">Confidential</option>
                <option value="urgent">Urgent</option>
              </select>
            </div>

            <div class="form-group" v-if="sendForm.coverPage !== 'none'">
              <label>Cover Page Message</label>
              <textarea v-model="sendForm.coverMessage" class="input-field textarea" placeholder="Optional message for cover page..." rows="3"></textarea>
            </div>

            <div class="form-group">
              <label>Documents *</label>
              <div 
                class="file-drop-zone"
                :class="{ dragover: isDragging, 'has-files': uploadedFiles.length > 0 }"
                @dragover.prevent="isDragging = true"
                @dragleave="isDragging = false"
                @drop.prevent="handleFileDrop"
                @click="triggerFileInput"
              >
                <input type="file" ref="fileInput" @change="handleFileSelect" multiple accept=".pdf,.doc,.docx,.tiff,.tif,.png,.jpg,.jpeg" hidden>
                
                <div v-if="uploadedFiles.length === 0" class="drop-placeholder">
                  <UploadCloudIcon class="upload-icon" />
                  <span>Drop files here or click to browse</span>
                  <span class="text-muted text-xs">PDF, DOC, DOCX, TIFF, PNG, JPG</span>
                </div>

                <div v-else class="file-list">
                  <div class="file-item" v-for="(file, i) in uploadedFiles" :key="i">
                    <FileIcon class="file-icon" />
                    <div class="file-info">
                      <span class="file-name">{{ file.name }}</span>
                      <span class="file-size">{{ formatFileSize(file.size) }}</span>
                    </div>
                    <button class="remove-file" @click.stop="removeFile(i)"><XIcon class="icon-xs" /></button>
                  </div>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label class="checkbox-row">
                <input type="checkbox" v-model="sendForm.notifyComplete">
                <span>Email me when fax is delivered</span>
              </label>
            </div>
          </div>

          <!-- Right: Preview -->
          <div class="preview-panel">
            <div class="preview-header">
              <span>Preview</span>
              <span class="preview-pages" v-if="previewState === 'ready'">{{ previewPages }} pages</span>
            </div>
            
            <div class="preview-content">
              <!-- Processing State -->
              <div class="preview-loading" v-if="previewState === 'processing'">
                <LoaderIcon class="spinning preview-spinner" />
                <span>Processing document...</span>
                <div class="processing-progress">
                  <div class="progress-bar small">
                    <div class="progress-fill" :style="{ width: processingProgress + '%' }"></div>
                  </div>
                  <span class="text-xs">{{ processingProgress }}%</span>
                </div>
              </div>

              <!-- Ready State -->
              <div class="preview-ready" v-else-if="previewState === 'ready'">
                <div class="preview-page">
                  <img :src="previewImage" alt="Fax Preview" class="preview-image">
                </div>
                <div class="preview-nav" v-if="previewPages > 1">
                  <button @click="prevPage" :disabled="currentPreviewPage === 1"><ChevronLeftIcon class="icon-sm" /></button>
                  <span>Page {{ currentPreviewPage }} of {{ previewPages }}</span>
                  <button @click="nextPage" :disabled="currentPreviewPage === previewPages"><ChevronRightIcon class="icon-sm" /></button>
                </div>
              </div>

              <!-- Empty State -->
              <div class="preview-empty" v-else>
                <FileTextIcon class="preview-empty-icon" />
                <span>Upload a document to see preview</span>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="closeSendModal">Cancel</button>
          <button class="btn-primary" @click="sendFax" :disabled="!canSend">
            <SendIcon class="btn-icon" />
            Send Fax
          </button>
        </div>
      </div>
    </div>

    <!-- VIEW FAX MODAL -->
    <div v-if="viewingFax" class="modal-overlay" @click.self="viewingFax = null">
      <div class="modal-card large">
        <div class="modal-header">
          <h3>Fax from {{ viewingFax.from || viewingFax.to }}</h3>
          <button class="btn-icon" @click="viewingFax = null"><XIcon class="icon-sm" /></button>
        </div>
        <div class="modal-body fax-viewer">
          <div class="viewer-page">
            <img src="https://placehold.co/600x800/f8fafc/64748b?text=Fax+Preview" alt="Fax Page" class="viewer-image">
          </div>
          <div class="viewer-nav">
            <button><ChevronLeftIcon class="icon-sm" /></button>
            <span>Page 1 of {{ viewingFax.pages }}</span>
            <button><ChevronRightIcon class="icon-sm" /></button>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="downloadFax(viewingFax)"><DownloadIcon class="btn-icon" /> Download PDF</button>
          <button class="btn-secondary" @click="forwardFax(viewingFax)"><ForwardIcon class="btn-icon" /> Forward</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { 
  Send as SendIcon, Inbox as InboxIcon, Clock as ClockIcon, Phone as PhoneIcon,
  FileDown as FileDownIcon, FileUp as FileUpIcon, File as FileIcon, FileText as FileTextIcon,
  Eye as EyeIcon, Download as DownloadIcon, Forward as ForwardIcon, Trash2 as TrashIcon,
  RefreshCw as RefreshCwIcon, Loader as LoaderIcon, X as XIcon,
  UploadCloud as UploadCloudIcon, ChevronLeft as ChevronLeftIcon, ChevronRight as ChevronRightIcon
} from 'lucide-vue-next'

const activeTab = ref('inbox')
const showSendModal = ref(false)
const viewingFax = ref(null)
const fileInput = ref(null)
const isDragging = ref(false)

// Send Form
const sendForm = ref({
  to: '',
  from: '(415) 555-3299',
  coverPage: 'none',
  coverMessage: '',
  notifyComplete: true
})

const uploadedFiles = ref([])
const previewState = ref('empty') // empty, processing, ready
const processingProgress = ref(0)
const previewPages = ref(0)
const currentPreviewPage = ref(1)
const previewImage = ref('')

// Data
const inboxData = ref([
  { id: 1, date: 'Dec 9, 2:30 PM', from: '(555) 987-6543', pages: 3, box: 'My Fax', read: false },
  { id: 2, date: 'Dec 9, 10:15 AM', from: '(555) 111-2222', pages: 1, box: 'My Fax', read: true },
  { id: 3, date: 'Dec 8, 4:45 PM', from: '(212) 555-9876', pages: 5, box: 'Sales Fax', read: true },
])

const sentData = ref([
  { id: 4, date: 'Dec 9, 11:00 AM', to: '(555) 555-1212', pages: 2, status: 'Sent' },
  { id: 5, date: 'Dec 8, 3:00 PM', to: '(555) 555-9999', pages: 5, status: 'Failed' },
  { id: 6, date: 'Dec 7, 9:30 AM', to: '(310) 555-4567', pages: 1, status: 'Sent' },
])

const pendingFaxes = ref([
  { id: 7, to: '(415) 555-8888', pages: 3, progress: 65, statusText: 'Transmitting page 2 of 3...' }
])

const unreadCount = computed(() => inboxData.value.filter(f => !f.read).length)
const canSend = computed(() => sendForm.value.to && uploadedFiles.value.length > 0 && previewState.value === 'ready')

// File Handling
const triggerFileInput = () => fileInput.value?.click()

const handleFileSelect = (e) => {
  const files = Array.from(e.target.files)
  addFiles(files)
}

const handleFileDrop = (e) => {
  isDragging.value = false
  const files = Array.from(e.dataTransfer.files)
  addFiles(files)
}

const addFiles = (files) => {
  uploadedFiles.value.push(...files)
  processDocuments()
}

const removeFile = (index) => {
  uploadedFiles.value.splice(index, 1)
  if (uploadedFiles.value.length === 0) {
    previewState.value = 'empty'
  }
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

// Document Processing (Mock)
const processDocuments = async () => {
  previewState.value = 'processing'
  processingProgress.value = 0
  
  // Simulate server processing
  for (let i = 0; i <= 100; i += 10) {
    await new Promise(r => setTimeout(r, 150))
    processingProgress.value = i
  }
  
  // Mock preview ready
  previewPages.value = uploadedFiles.value.length + (sendForm.value.coverPage !== 'none' ? 1 : 0)
  currentPreviewPage.value = 1
  previewImage.value = 'https://placehold.co/400x520/f8fafc/64748b?text=Page+1'
  previewState.value = 'ready'
}

const prevPage = () => {
  if (currentPreviewPage.value > 1) {
    currentPreviewPage.value--
    previewImage.value = `https://placehold.co/400x520/f8fafc/64748b?text=Page+${currentPreviewPage.value}`
  }
}

const nextPage = () => {
  if (currentPreviewPage.value < previewPages.value) {
    currentPreviewPage.value++
    previewImage.value = `https://placehold.co/400x520/f8fafc/64748b?text=Page+${currentPreviewPage.value}`
  }
}

watch(() => sendForm.value.coverPage, () => {
  if (uploadedFiles.value.length > 0) {
    previewPages.value = uploadedFiles.value.length + (sendForm.value.coverPage !== 'none' ? 1 : 0)
  }
})

// Actions
const sendFax = () => {
  const newFax = {
    id: Date.now(),
    to: sendForm.value.to,
    pages: previewPages.value,
    progress: 0,
    statusText: 'Connecting...'
  }
  pendingFaxes.value.push(newFax)
  closeSendModal()
  activeTab.value = 'pending'
  
  // Simulate sending
  simulateFaxSend(newFax)
}

const simulateFaxSend = async (fax) => {
  for (let i = 0; i <= 100; i += 5) {
    await new Promise(r => setTimeout(r, 200))
    const idx = pendingFaxes.value.findIndex(f => f.id === fax.id)
    if (idx !== -1) {
      pendingFaxes.value[idx].progress = i
      pendingFaxes.value[idx].statusText = i < 100 ? `Transmitting page ${Math.ceil(i / 33)} of ${fax.pages}...` : 'Completing...'
    }
  }
  
  // Move to sent
  pendingFaxes.value = pendingFaxes.value.filter(f => f.id !== fax.id)
  sentData.value.unshift({
    id: fax.id,
    date: 'Just now',
    to: fax.to,
    pages: fax.pages,
    status: 'Sent'
  })
}

const closeSendModal = () => {
  showSendModal.value = false
  sendForm.value = { to: '', from: '(415) 555-3299', coverPage: 'none', coverMessage: '', notifyComplete: true }
  uploadedFiles.value = []
  previewState.value = 'empty'
}

const viewFax = (fax) => { 
  viewingFax.value = fax
  if (!fax.read) fax.read = true
}
const downloadFax = (fax) => alert(`Downloading fax from ${fax.from || fax.to}`)
const forwardFax = (fax) => alert(`Forward fax to...`)
const deleteFax = (fax) => {
  if (confirm('Delete this fax?')) {
    inboxData.value = inboxData.value.filter(f => f.id !== fax.id)
    sentData.value = sentData.value.filter(f => f.id !== fax.id)
  }
}
const resendFax = (fax) => alert(`Resending fax to ${fax.to}`)
const cancelFax = (fax) => {
  pendingFaxes.value = pendingFaxes.value.filter(f => f.id !== fax.id)
}
</script>

<style scoped>
.fax-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.inbox { background: #dbeafe; color: #2563eb; }
.stat-icon.sent { background: #dcfce7; color: #16a34a; }
.stat-icon.pending { background: #fef3c7; color: #b45309; }
.stat-icon.fax-number { background: #f3e8ff; color: #7c3aed; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 18px; font-weight: 700; }
.stat-value.mono { font-family: monospace; font-size: 14px; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Tabs */
.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 4px 4px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 16px; height: 16px; }
.tab-badge { background: #ef4444; color: white; font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 10px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 var(--radius-md) var(--radius-md); min-height: 300px; }

/* Fax List */
.fax-list { display: flex; flex-direction: column; gap: 8px; }

.fax-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--bg-app);
  border-radius: var(--radius-sm);
  transition: all 0.15s;
}
.fax-item:hover { background: #f1f5f9; }
.fax-item.unread { background: #eff6ff; border-left: 3px solid var(--primary-color); }
.fax-item.pending { background: #fef3c7; }

.fax-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.fax-icon.incoming { background: #dcfce7; color: #16a34a; }
.fax-icon.outgoing { background: #dbeafe; color: #2563eb; }
.fax-icon.processing { background: #fef3c7; color: #b45309; }

.fax-main { flex: 1; }
.fax-header { display: flex; justify-content: space-between; margin-bottom: 4px; }
.fax-from, .fax-to { font-weight: 600; font-size: 14px; }
.fax-date { font-size: 12px; color: var(--text-muted); }
.fax-status { font-size: 12px; color: var(--text-muted); font-style: italic; }

.fax-meta { display: flex; gap: 16px; }
.meta-item { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--text-muted); }
.meta-icon { width: 12px; height: 12px; }

.status-badge { font-size: 10px; font-weight: 700; padding: 3px 8px; border-radius: 4px; text-transform: uppercase; }
.status-badge.sent { background: #dcfce7; color: #16a34a; }
.status-badge.failed { background: #fee2e2; color: #dc2626; }

.progress-bar { height: 4px; background: rgba(0,0,0,0.1); border-radius: 2px; margin: 8px 0; overflow: hidden; }
.progress-bar.small { height: 3px; margin: 4px 0; }
.progress-fill { height: 100%; background: var(--primary-color); border-radius: 2px; transition: width 0.3s; }

.fax-actions { display: flex; gap: 4px; }
.action-btn { width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); transition: all 0.15s; }
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn.danger:hover { border-color: #ef4444; color: #ef4444; }

.empty-state { text-align: center; padding: 48px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 16px; }

/* Send Fax Modal */
.send-fax-body { display: flex; gap: 24px; }
.send-form { flex: 1; }

.preview-panel { width: 280px; background: var(--bg-app); border-radius: var(--radius-sm); display: flex; flex-direction: column; }
.preview-header { display: flex; justify-content: space-between; padding: 12px; border-bottom: 1px solid var(--border-color); font-size: 12px; font-weight: 600; }
.preview-pages { color: var(--text-muted); }
.preview-content { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 16px; min-height: 300px; }

.preview-loading, .preview-empty { display: flex; flex-direction: column; align-items: center; gap: 12px; color: var(--text-muted); font-size: 12px; text-align: center; }
.preview-spinner { width: 32px; height: 32px; }
.preview-empty-icon { width: 48px; height: 48px; opacity: 0.3; }
.processing-progress { display: flex; align-items: center; gap: 8px; width: 100%; }

.preview-ready { width: 100%; }
.preview-page { background: white; border: 1px solid var(--border-color); border-radius: 4px; overflow: hidden; margin-bottom: 12px; }
.preview-image { width: 100%; display: block; }
.preview-nav { display: flex; align-items: center; justify-content: center; gap: 12px; font-size: 11px; color: var(--text-muted); }
.preview-nav button { width: 28px; height: 28px; border-radius: 4px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.preview-nav button:disabled { opacity: 0.5; cursor: not-allowed; }

/* File Drop Zone */
.file-drop-zone {
  border: 2px dashed var(--border-color);
  border-radius: var(--radius-sm);
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}
.file-drop-zone:hover, .file-drop-zone.dragover { border-color: var(--primary-color); background: var(--primary-light); }
.file-drop-zone.has-files { border-style: solid; padding: 12px; }

.drop-placeholder { display: flex; flex-direction: column; align-items: center; gap: 8px; color: var(--text-muted); }
.upload-icon { width: 32px; height: 32px; opacity: 0.5; }

.file-list { display: flex; flex-direction: column; gap: 8px; }
.file-item { display: flex; align-items: center; gap: 10px; padding: 8px; background: white; border-radius: 4px; }
.file-icon { width: 20px; height: 20px; color: var(--primary-color); }
.file-info { flex: 1; text-align: left; }
.file-name { font-size: 13px; font-weight: 500; display: block; }
.file-size { font-size: 11px; color: var(--text-muted); }
.remove-file { width: 20px; height: 20px; border: none; background: none; cursor: pointer; color: var(--text-muted); display: flex; align-items: center; justify-content: center; }
.remove-file:hover { color: #ef4444; }

/* Fax Viewer */
.fax-viewer { display: flex; flex-direction: column; align-items: center; }
.viewer-page { background: white; border: 1px solid var(--border-color); border-radius: 4px; margin-bottom: 16px; overflow: hidden; max-height: 60vh; overflow-y: auto; }
.viewer-image { width: 100%; max-width: 600px; display: block; }
.viewer-nav { display: flex; align-items: center; gap: 16px; }
.viewer-nav button { width: 32px; height: 32px; border-radius: 4px; border: 1px solid var(--border-color); background: white; cursor: pointer; display: flex; align-items: center; justify-content: center; }

/* Form */
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field.textarea { resize: vertical; min-height: 60px; }
.checkbox-row { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }

/* Modal */
.modal-overlay { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.modal-card { background: white; border-radius: var(--radius-md); width: 100%; max-width: 480px; max-height: 90vh; display: flex; flex-direction: column; }
.modal-card.large { max-width: 720px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { font-size: 16px; font-weight: 700; margin: 0; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-actions { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 20px; border-top: 1px solid var(--border-color); }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { background: none; border: none; cursor: pointer; color: var(--text-muted); padding: 4px; width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 12px; height: 12px; }
.icon { width: 20px; height: 20px; }

.spinning { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
