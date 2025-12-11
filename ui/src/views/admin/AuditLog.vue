<template>
  <div class="audit-log-page">
    <div class="view-header">
      <div class="header-content">
        <h2>Audit Log</h2>
        <p class="text-muted text-sm">Track all administrative actions and system changes.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="exportLogs">
          <DownloadIcon class="btn-icon" /> Export
        </button>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon total"><ActivityIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ logs.length }}</span>
          <span class="stat-label">Total Events</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon today"><ClockIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ todayCount }}</span>
          <span class="stat-label">Today</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon security"><ShieldIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ securityCount }}</span>
          <span class="stat-label">Security Events</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon config"><SettingsIcon class="icon" /></div>
        <div class="stat-info">
          <span class="stat-value">{{ configCount }}</span>
          <span class="stat-label">Config Changes</span>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <div class="search-box">
        <SearchIcon class="search-icon" />
        <input v-model="searchQuery" class="search-input" placeholder="Search by user, action, or resource...">
      </div>
      <select v-model="categoryFilter" class="filter-select">
        <option value="">All Categories</option>
        <option value="security">Security</option>
        <option value="configuration">Configuration</option>
        <option value="user">User Management</option>
        <option value="telephony">Telephony</option>
      </select>
      <select v-model="severityFilter" class="filter-select">
        <option value="">All Severity</option>
        <option value="info">Info</option>
        <option value="warning">Warning</option>
        <option value="critical">Critical</option>
      </select>
      <select v-model="dateFilter" class="filter-select">
        <option value="today">Today</option>
        <option value="week">This Week</option>
        <option value="month">This Month</option>
        <option value="all">All Time</option>
      </select>
    </div>

    <!-- Logs List -->
    <div class="logs-container">
      <div class="logs-timeline">
        <div 
          class="log-entry" 
          v-for="log in filteredLogs" 
          :key="log.id"
          :class="log.severity"
        >
          <div class="log-icon" :class="log.category">
            <UserIcon v-if="log.category === 'user'" class="icon-sm" />
            <ShieldIcon v-else-if="log.category === 'security'" class="icon-sm" />
            <SettingsIcon v-else-if="log.category === 'configuration'" class="icon-sm" />
            <PhoneIcon v-else-if="log.category === 'telephony'" class="icon-sm" />
            <ActivityIcon v-else class="icon-sm" />
          </div>
          
          <div class="log-content">
            <div class="log-header">
              <span class="log-action">{{ log.action }}</span>
              <span class="log-time">{{ log.time }}</span>
            </div>
            <div class="log-details">
              <span class="log-user"><UserIcon class="detail-icon" /> {{ log.user }}</span>
              <span class="log-resource" v-if="log.resource">
                <FolderIcon class="detail-icon" /> {{ log.resource }}
              </span>
              <span class="log-ip" v-if="log.ip">
                <GlobeIcon class="detail-icon" /> {{ log.ip }}
              </span>
            </div>
            <div class="log-description" v-if="log.description">
              {{ log.description }}
            </div>
            <div class="log-changes" v-if="log.changes">
              <button class="expand-btn" @click="log.expanded = !log.expanded">
                <ChevronDownIcon class="icon-xs" :class="{ rotated: log.expanded }" />
                View Changes
              </button>
              <div class="changes-content" v-if="log.expanded">
                <div class="change-item" v-for="(change, key) in log.changes" :key="key">
                  <span class="change-key">{{ key }}:</span>
                  <span class="change-old">{{ change.old }}</span>
                  <ArrowRightIcon class="change-arrow" />
                  <span class="change-new">{{ change.new }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="log-severity">
            <span class="severity-badge" :class="log.severity">{{ log.severity }}</span>
          </div>
        </div>
      </div>

      <div class="empty-state" v-if="filteredLogs.length === 0">
        <ActivityIcon class="empty-icon" />
        <p>No audit logs found</p>
      </div>
    </div>

    <!-- Pagination -->
    <div class="pagination" v-if="filteredLogs.length > 0">
      <span class="page-info">Showing {{ filteredLogs.length }} of {{ logs.length }} entries</span>
      <div class="page-controls">
        <button class="page-btn" disabled><ChevronLeftIcon class="icon-sm" /></button>
        <span class="page-number">1</span>
        <button class="page-btn"><ChevronRightIcon class="icon-sm" /></button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  Activity as ActivityIcon, Clock as ClockIcon, Shield as ShieldIcon, 
  Settings as SettingsIcon, Search as SearchIcon, Download as DownloadIcon,
  User as UserIcon, Phone as PhoneIcon, Folder as FolderIcon, Globe as GlobeIcon,
  ChevronDown as ChevronDownIcon, ChevronLeft as ChevronLeftIcon, 
  ChevronRight as ChevronRightIcon, ArrowRight as ArrowRightIcon
} from 'lucide-vue-next'

const searchQuery = ref('')
const categoryFilter = ref('')
const severityFilter = ref('')
const dateFilter = ref('week')

const logs = ref([
  { id: 1, action: 'User Login', user: 'admin@company.com', category: 'security', severity: 'info', time: '10:45 AM', date: 'today', ip: '192.168.1.100', description: 'Successful login from admin portal' },
  { id: 2, action: 'Extension Modified', user: 'admin@company.com', category: 'configuration', severity: 'info', time: '10:32 AM', date: 'today', resource: 'Extension 101', changes: { 'Forward to': { old: 'Disabled', new: '(415) 555-1234' }, 'Ring timeout': { old: '20', new: '30' } } },
  { id: 3, action: 'Failed Login Attempt', user: 'unknown@test.com', category: 'security', severity: 'warning', time: '10:15 AM', date: 'today', ip: '203.0.113.50', description: 'Invalid credentials - account locked after 3 attempts' },
  { id: 4, action: 'User Created', user: 'admin@company.com', category: 'user', severity: 'info', time: '9:30 AM', date: 'today', resource: 'jane.doe@company.com', description: 'New tenant admin user created' },
  { id: 5, action: 'Gateway Configuration Changed', user: 'system@company.com', category: 'configuration', severity: 'warning', time: '9:00 AM', date: 'today', resource: 'Primary SIP Gateway', changes: { 'Codec priority': { old: 'G711,G729', new: 'G729,G711,OPUS' } } },
  { id: 6, action: 'Dial Plan Modified', user: 'admin@company.com', category: 'telephony', severity: 'info', time: 'Yesterday, 4:30 PM', date: 'week', resource: 'International Outbound', description: 'Updated international dial plan routing' },
  { id: 7, action: 'Bulk Extension Import', user: 'admin@company.com', category: 'configuration', severity: 'info', time: 'Yesterday, 2:00 PM', date: 'week', description: '15 extensions imported from CSV' },
  { id: 8, action: 'System Backup Completed', user: 'system', category: 'configuration', severity: 'info', time: 'Dec 8, 11:00 PM', date: 'week', description: 'Automated nightly backup completed successfully' },
  { id: 9, action: 'API Key Generated', user: 'admin@company.com', category: 'security', severity: 'critical', time: 'Dec 8, 10:00 AM', date: 'week', description: 'New API key generated for CRM integration' },
])

const todayCount = computed(() => logs.value.filter(l => l.date === 'today').length)
const securityCount = computed(() => logs.value.filter(l => l.category === 'security').length)
const configCount = computed(() => logs.value.filter(l => l.category === 'configuration').length)

const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    const matchesSearch = !searchQuery.value || 
      log.action.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      log.user.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      (log.resource && log.resource.toLowerCase().includes(searchQuery.value.toLowerCase()))
    
    const matchesCategory = !categoryFilter.value || log.category === categoryFilter.value
    const matchesSeverity = !severityFilter.value || log.severity === severityFilter.value
    const matchesDate = dateFilter.value === 'all' || 
      (dateFilter.value === 'today' && log.date === 'today') ||
      (dateFilter.value === 'week' && ['today', 'week'].includes(log.date)) ||
      (dateFilter.value === 'month')
    
    return matchesSearch && matchesCategory && matchesSeverity && matchesDate
  })
})

const exportLogs = () => alert('Exporting audit logs...')
</script>

<style scoped>
.audit-log-page { padding: 0; }

.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }
.header-actions { display: flex; gap: 12px; }

/* Stats */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: var(--spacing-lg); }
.stat-card { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 16px; display: flex; align-items: center; gap: 12px; }
.stat-icon { width: 40px; height: 40px; border-radius: 10px; display: flex; align-items: center; justify-content: center; }
.stat-icon .icon { width: 20px; height: 20px; }
.stat-icon.total { background: #f3e8ff; color: #7c3aed; }
.stat-icon.today { background: #dbeafe; color: #2563eb; }
.stat-icon.security { background: #fee2e2; color: #dc2626; }
.stat-icon.config { background: #dcfce7; color: #16a34a; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 20px; font-weight: 700; }
.stat-label { font-size: 12px; color: var(--text-muted); }

/* Filters */
.filter-bar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.search-box { position: relative; flex: 1; min-width: 200px; max-width: 320px; }
.search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.search-input { width: 100%; padding: 10px 12px 10px 38px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }
.filter-select { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; background: white; }

/* Logs Container */
.logs-container { background: white; border: 1px solid var(--border-color); border-radius: var(--radius-md); overflow: hidden; }

.logs-timeline { display: flex; flex-direction: column; }

.log-entry {
  display: flex;
  gap: 16px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  border-left: 3px solid transparent;
  transition: background 0.15s;
}
.log-entry:last-child { border-bottom: none; }
.log-entry:hover { background: var(--bg-app); }
.log-entry.info { border-left-color: #3b82f6; }
.log-entry.warning { border-left-color: #f59e0b; }
.log-entry.critical { border-left-color: #ef4444; }

.log-icon { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.log-icon.security { background: #fee2e2; color: #dc2626; }
.log-icon.configuration { background: #dcfce7; color: #16a34a; }
.log-icon.user { background: #dbeafe; color: #2563eb; }
.log-icon.telephony { background: #f3e8ff; color: #7c3aed; }

.log-content { flex: 1; min-width: 0; }
.log-header { display: flex; justify-content: space-between; margin-bottom: 4px; }
.log-action { font-weight: 600; font-size: 14px; }
.log-time { font-size: 12px; color: var(--text-muted); }

.log-details { display: flex; gap: 16px; margin-bottom: 6px; flex-wrap: wrap; }
.log-details > span { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--text-muted); }
.detail-icon { width: 12px; height: 12px; }

.log-description { font-size: 13px; color: var(--text-muted); }

.log-changes { margin-top: 8px; }
.expand-btn { display: flex; align-items: center; gap: 4px; background: none; border: none; color: var(--primary-color); font-size: 12px; font-weight: 500; cursor: pointer; padding: 4px 0; }
.expand-btn .icon-xs { transition: transform 0.2s; }
.expand-btn .rotated { transform: rotate(180deg); }

.changes-content { margin-top: 8px; padding: 12px; background: var(--bg-app); border-radius: 6px; }
.change-item { display: flex; align-items: center; gap: 8px; font-size: 12px; margin-bottom: 4px; }
.change-item:last-child { margin-bottom: 0; }
.change-key { font-weight: 600; min-width: 120px; }
.change-old { color: #dc2626; text-decoration: line-through; }
.change-arrow { width: 12px; height: 12px; color: var(--text-muted); }
.change-new { color: #16a34a; font-weight: 500; }

.log-severity { flex-shrink: 0; }
.severity-badge { font-size: 10px; font-weight: 700; padding: 4px 10px; border-radius: 4px; text-transform: uppercase; }
.severity-badge.info { background: #dbeafe; color: #2563eb; }
.severity-badge.warning { background: #fef3c7; color: #b45309; }
.severity-badge.critical { background: #fee2e2; color: #dc2626; }

.empty-state { text-align: center; padding: 48px; color: var(--text-muted); }
.empty-icon { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 16px; }

/* Pagination */
.pagination { display: flex; justify-content: space-between; align-items: center; padding: 16px; background: white; border: 1px solid var(--border-color); border-top: none; border-radius: 0 0 var(--radius-md) var(--radius-md); }
.page-info { font-size: 12px; color: var(--text-muted); }
.page-controls { display: flex; align-items: center; gap: 8px; }
.page-btn { width: 32px; height: 32px; border: 1px solid var(--border-color); background: white; border-radius: 6px; cursor: pointer; display: flex; align-items: center; justify-content: center; }
.page-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.page-number { font-size: 13px; font-weight: 600; padding: 0 8px; }

/* Buttons */
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; }
.btn-icon { width: 16px; height: 16px; }

.icon-sm { width: 16px; height: 16px; }
.icon-xs { width: 12px; height: 12px; }
.icon { width: 20px; height: 20px; }

@media (max-width: 768px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .filter-bar { flex-direction: column; }
  .search-box { max-width: none; }
}
</style>
