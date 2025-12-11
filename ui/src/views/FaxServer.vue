<template>
  <div class="fax-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Fax Server</h2>
        <p class="text-muted text-sm">Manage virtual fax machines and view fax logs.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="showComposeModal = true">
          <SendIcon class="btn-icon" /> Send Fax
        </button>
        <button class="btn-primary" @click="showNewServerModal = true">
          <PlusIcon class="btn-icon" /> New Fax Box
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ servers.length }}</div>
        <div class="stat-label">Fax Boxes</div>
      </div>
      <div class="stat-card highlight">
        <div class="stat-value">{{ totalInbox }}</div>
        <div class="stat-label">Inbox</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ totalSent }}</div>
        <div class="stat-label">Sent Today</div>
      </div>
      <div class="stat-card" :class="{ alert: pendingCount > 0 }">
        <div class="stat-value">{{ pendingCount }}</div>
        <div class="stat-label">Pending</div>
      </div>
    </div>

    <!-- Layout -->
    <div class="fax-layout">
      <!-- Sidebar -->
      <div class="fax-sidebar">
        <div class="sidebar-header">Fax Boxes</div>
        <div 
          v-for="server in servers" 
          :key="server.id"
          class="server-item"
          :class="{ active: activeServer.id === server.id }"
          @click="activeServer = server"
        >
          <div class="server-icon">
            <PrinterIcon />
          </div>
          <div class="server-info">
            <div class="server-name">{{ server.name }}</div>
            <div class="server-did">{{ server.did }}</div>
          </div>
          <div class="server-badge" v-if="server.unread > 0">{{ server.unread }}</div>
        </div>
      </div>

      <!-- Main Area -->
      <div class="fax-main">
        <div class="server-header">
          <div class="server-details">
            <h3>{{ activeServer.name }}</h3>
            <div class="server-meta">
              <span><PhoneIcon /> Ext. {{ activeServer.ext }}</span>
              <span><MailIcon /> {{ activeServer.email }}</span>
            </div>
          </div>
          <button class="btn-icon-sm" @click="editServer(activeServer)" title="Edit">
            <SettingsIcon />
          </button>
        </div>

        <!-- Tabs -->
        <div class="tabs">
          <button class="tab" :class="{ active: activeTab === 'inbox' }" @click="activeTab = 'inbox'">
            <InboxIcon class="tab-icon" /> Inbox
          </button>
          <button class="tab" :class="{ active: activeTab === 'sent' }" @click="activeTab = 'sent'">
            <SendIcon class="tab-icon" /> Sent
          </button>
          <button class="tab" :class="{ active: activeTab === 'settings' }" @click="activeTab = 'settings'">
            <SettingsIcon class="tab-icon" /> Settings
          </button>
        </div>

        <!-- Inbox Tab -->
        <div class="tab-content" v-if="activeTab === 'inbox'">
          <table class="fax-table">
            <thead>
              <tr>
                <th></th>
                <th>From</th>
                <th>Date/Time</th>
                <th>Pages</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="fax in inboxData" :key="fax.id" :class="{ unread: fax.unread }">
                <td><input type="checkbox"></td>
                <td><span class="caller-id">{{ fax.from }}</span></td>
                <td>{{ fax.date }}</td>
                <td>{{ fax.pages }} pages</td>
                <td class="actions-cell">
                  <button class="btn-link"><EyeIcon /> View</button>
                  <button class="btn-link"><DownloadIcon /> PDF</button>
                  <button class="btn-link danger"><TrashIcon /></button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Sent Tab -->
        <div class="tab-content" v-if="activeTab === 'sent'">
          <table class="fax-table">
            <thead>
              <tr>
                <th>To</th>
                <th>Date/Time</th>
                <th>Pages</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="fax in sentData" :key="fax.id">
                <td><span class="caller-id">{{ fax.to }}</span></td>
                <td>{{ fax.date }}</td>
                <td>{{ fax.pages }} pages</td>
                <td><span class="status-pill" :class="fax.status.toLowerCase()">{{ fax.status }}</span></td>
                <td class="actions-cell">
                  <button class="btn-link"><RefreshIcon /> Retry</button>
                  <button class="btn-link danger"><TrashIcon /></button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Settings Tab -->
        <div class="tab-content" v-if="activeTab === 'settings'">
          <div class="settings-grid">
            <div class="form-group">
              <label>Fax Box Name</label>
              <input v-model="activeServer.name" class="input-field">
            </div>
            <div class="form-group">
              <label>Assigned DID</label>
              <input v-model="activeServer.did" class="input-field">
            </div>
            <div class="form-group">
              <label>Extension</label>
              <input v-model="activeServer.ext" class="input-field">
            </div>
            <div class="form-group">
              <label>Email Notification</label>
              <input v-model="activeServer.email" class="input-field">
            </div>
            <div class="form-group">
              <label>Retention Days</label>
              <input type="number" v-model.number="activeServer.retentionDays" class="input-field">
            </div>
            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="activeServer.emailOnReceive">
                Email on Receive
              </label>
            </div>
          </div>
          <div class="settings-footer">
            <button class="btn-primary">Save Settings</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Compose Modal -->
    <div class="modal-overlay" v-if="showComposeModal" @click.self="showComposeModal = false">
      <div class="modal-card">
        <div class="modal-header">
          <h3>Send Fax</h3>
          <button class="close-btn" @click="showComposeModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>From Fax Box</label>
            <select class="input-field">
              <option v-for="s in servers" :key="s.id" :value="s.id">{{ s.name }} ({{ s.did }})</option>
            </select>
          </div>
          <div class="form-group">
            <label>To Number</label>
            <input class="input-field" placeholder="+1 555-123-4567">
          </div>
          <div class="form-group">
            <label>Subject (Cover Page)</label>
            <input class="input-field" placeholder="Invoice #12345">
          </div>
          <div class="form-group">
            <label>Attach PDF</label>
            <div class="file-upload">
              <input type="file" id="fax-file" accept=".pdf">
              <label for="fax-file" class="file-label">Choose file or drag & drop</label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showComposeModal = false">Cancel</button>
          <button class="btn-primary">Send Fax</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  Plus as PlusIcon, Send as SendIcon, Printer as PrinterIcon, Phone as PhoneIcon,
  Mail as MailIcon, Settings as SettingsIcon, Inbox as InboxIcon, Eye as EyeIcon,
  Download as DownloadIcon, Trash2 as TrashIcon, RefreshCw as RefreshIcon
} from 'lucide-vue-next'

const activeTab = ref('inbox')
const showComposeModal = ref(false)
const showNewServerModal = ref(false)

const servers = ref([
  { id: 1, name: 'Sales Fax', did: '(415) 555-0199', ext: '8801', email: 'sales@fax.acme.com', unread: 3, retentionDays: 90, emailOnReceive: true },
  { id: 2, name: 'HR Confidential', did: '(415) 555-0299', ext: '8802', email: 'hr@fax.acme.com', unread: 0, retentionDays: 365, emailOnReceive: true },
  { id: 3, name: 'General', did: '(415) 555-0300', ext: '8800', email: 'fax@acme.com', unread: 1, retentionDays: 30, emailOnReceive: false },
])

const activeServer = ref(servers.value[0])

const inboxData = ref([
  { id: 1, from: '(555) 987-6543', date: '2024-03-20 14:30', pages: 2, unread: true },
  { id: 2, from: '(555) 111-2222', date: '2024-03-19 09:15', pages: 5, unread: false },
  { id: 3, from: '(555) 333-4444', date: '2024-03-18 16:45', pages: 1, unread: true },
])

const sentData = ref([
  { id: 1, to: '(555) 555-5555', date: '2024-03-20 10:00', pages: 3, status: 'Sent' },
  { id: 2, to: '(555) 444-4444', date: '2024-03-20 10:05', pages: 2, status: 'Failed' },
  { id: 3, to: '(555) 222-2222', date: '2024-03-19 15:30', pages: 1, status: 'Pending' },
])

const totalInbox = computed(() => servers.value.reduce((sum, s) => sum + s.unread, 0) + 2)
const totalSent = computed(() => sentData.value.filter(f => f.status === 'Sent').length)
const pendingCount = computed(() => sentData.value.filter(f => f.status === 'Pending').length)

const editServer = (server) => { /* open edit modal */ }
</script>

<style scoped>
.fax-page { padding: 0; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 8px; }
.btn-primary, .btn-secondary { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-primary { background: var(--primary-color); color: white; }
.btn-secondary { background: white; border: 1px solid var(--border-color); }
.btn-icon { width: 14px; height: 14px; }

.stats-row { display: flex; gap: 16px; margin-bottom: 20px; }
.stat-card { flex: 1; background: white; border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; text-align: center; }
.stat-card.highlight { border-color: var(--primary-color); background: #f0f9ff; }
.stat-card.alert { border-color: #f59e0b; background: #fffbeb; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-label { font-size: 11px; color: var(--text-muted); text-transform: uppercase; margin-top: 4px; }

.fax-layout { display: grid; grid-template-columns: 240px 1fr; gap: 20px; min-height: 400px; }

.fax-sidebar { background: white; border: 1px solid var(--border-color); border-radius: 8px; overflow: hidden; }
.sidebar-header { background: #f8fafc; padding: 12px 16px; font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); }
.server-item { display: flex; align-items: center; gap: 12px; padding: 12px 16px; cursor: pointer; border-bottom: 1px solid var(--border-color); transition: all 0.2s; }
.server-item:hover { background: #f8fafc; }
.server-item.active { background: #eff6ff; border-left: 3px solid var(--primary-color); }
.server-icon { width: 32px; height: 32px; background: #f1f5f9; border-radius: 6px; display: flex; align-items: center; justify-content: center; color: var(--text-muted); }
.server-icon svg { width: 16px; height: 16px; }
.server-info { flex: 1; }
.server-name { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.server-did { font-size: 11px; color: var(--text-muted); }
.server-badge { background: var(--primary-color); color: white; font-size: 10px; font-weight: 600; padding: 2px 6px; border-radius: 10px; }

.fax-main { display: flex; flex-direction: column; }
.server-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.server-details h3 { margin: 0 0 4px; font-size: 16px; }
.server-meta { display: flex; gap: 16px; font-size: 12px; color: var(--text-muted); }
.server-meta span { display: flex; align-items: center; gap: 4px; }
.server-meta svg { width: 12px; height: 12px; }
.btn-icon-sm { width: 32px; height: 32px; background: white; border: 1px solid var(--border-color); border-radius: 6px; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); }
.btn-icon-sm:hover { color: var(--primary-color); border-color: var(--primary-color); }
.btn-icon-sm svg { width: 14px; height: 14px; }

.tabs { display: flex; gap: 2px; border-bottom: 1px solid var(--border-color); }
.tab { display: flex; align-items: center; gap: 6px; padding: 10px 16px; background: transparent; border: 1px solid transparent; border-bottom: none; cursor: pointer; font-size: 13px; font-weight: 500; color: var(--text-muted); border-radius: 6px 6px 0 0; }
.tab.active { background: white; border-color: var(--border-color); color: var(--primary-color); margin-bottom: -1px; }
.tab-icon { width: 14px; height: 14px; }
.tab-content { background: white; border: 1px solid var(--border-color); border-top: none; padding: 20px; border-radius: 0 0 8px 8px; flex: 1; }

.fax-table { width: 100%; border-collapse: collapse; }
.fax-table th { text-align: left; padding: 10px 12px; font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); border-bottom: 1px solid var(--border-color); }
.fax-table td { padding: 12px; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.fax-table tr:hover { background: #f8fafc; }
.fax-table tr.unread { font-weight: 600; background: #fefce8; }
.caller-id { font-family: monospace; }
.actions-cell { display: flex; gap: 8px; }
.btn-link { display: flex; align-items: center; gap: 4px; background: none; border: none; color: var(--primary-color); font-size: 11px; font-weight: 500; cursor: pointer; }
.btn-link svg { width: 12px; height: 12px; }
.btn-link.danger { color: #ef4444; }
.status-pill { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; }
.status-pill.sent { background: #dcfce7; color: #16a34a; }
.status-pill.failed { background: #fef2f2; color: #ef4444; }
.status-pill.pending { background: #fef3c7; color: #d97706; }

.settings-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; max-width: 600px; }
.form-group { margin-bottom: 0; }
.form-group label { display: block; font-size: 11px; font-weight: 600; text-transform: uppercase; color: var(--text-muted); margin-bottom: 6px; }
.input-field { width: 100%; padding: 8px 10px; border: 1px solid var(--border-color); border-radius: 6px; font-size: 13px; box-sizing: border-box; }
.checkbox-label { display: flex !important; align-items: center; gap: 8px; font-size: 12px !important; cursor: pointer; text-transform: none !important; }
.settings-footer { margin-top: 20px; padding-top: 16px; border-top: 1px solid var(--border-color); }

/* Modal */
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal-card { background: white; border-radius: 12px; width: 90%; max-width: 480px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; border-bottom: 1px solid var(--border-color); }
.modal-header h3 { margin: 0; font-size: 16px; }
.close-btn { width: 28px; height: 28px; border: none; background: #f1f5f9; border-radius: 6px; font-size: 18px; cursor: pointer; }
.modal-body { padding: 20px; }
.modal-body .form-group { margin-bottom: 16px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 16px 20px; border-top: 1px solid var(--border-color); }
.file-upload { border: 2px dashed var(--border-color); border-radius: 8px; padding: 24px; text-align: center; }
.file-upload input { display: none; }
.file-label { color: var(--text-muted); cursor: pointer; font-size: 13px; }
</style>
