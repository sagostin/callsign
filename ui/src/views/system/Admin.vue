<template>
  <div class="page-content">
    <!-- View Header -->
    <div class="view-header">
      <div class="view-header-content">
        <h1 class="view-header-title">System Dashboard</h1>
        <p class="view-header-subtitle">Cluster overview and infrastructure status</p>
      </div>
      <div class="view-header-actions">
        <div class="status-indicator">
          <span class="status-dot online"></span>
          <span class="status-text">System Online</span>
        </div>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <StatCard 
        label="Total Tenants"
        :value="stats.tenants"
        iconName="users"
        variant="info"
      />
      <StatCard 
        label="Active Users"
        :value="formattedUsers"
        iconName="users"
        variant="success"
      />
      <StatCard 
        label="Active Channels"
        :value="stats.active_channels"
        iconName="calls"
        variant="warning"
      />
      <StatCard 
        label="System Alerts"
        :value="stats.alerts"
        iconName="alert"
        :variant="stats.alerts > 0 ? 'error' : 'success'"
      />
    </div>

    <!-- Registration Stats -->
    <div class="section">
      <div class="section-header">
        <h2 class="section-title">Device Registrations</h2>
        <span class="section-subtitle">Real-time device status across all tenants</span>
      </div>
      
      <div class="registration-grid">
        <div class="reg-card">
          <div class="reg-header">
            <div class="reg-icon-wrapper bg-blue-100">
              <MonitorIcon class="reg-icon text-blue-600" />
            </div>
            <span class="reg-title">Desk Phones</span>
          </div>
          <div class="reg-stats">
            <div class="stat-num">{{ stats.devices?.desk_phones?.total || 0 }}</div>
            <div class="stat-details">
              <span class="online">
                <span class="dot"></span>
                {{ stats.devices?.desk_phones?.online || 0 }} online
              </span>
              <span class="offline">
                <span class="dot"></span>
                {{ (stats.devices?.desk_phones?.total || 0) - (stats.devices?.desk_phones?.online || 0) }} offline
              </span>
            </div>
          </div>
        </div>

        <div class="reg-card">
          <div class="reg-header">
            <div class="reg-icon-wrapper bg-emerald-100">
              <SmartphoneIcon class="reg-icon text-emerald-600" />
            </div>
            <span class="reg-title">Softphones</span>
          </div>
          <div class="reg-stats">
            <div class="stat-num">{{ stats.devices?.softphones?.total || 0 }}</div>
            <div class="stat-details">
              <span class="online">
                <span class="dot"></span>
                {{ stats.devices?.softphones?.online || 0 }} online
              </span>
              <span class="offline">
                <span class="dot"></span>
                {{ (stats.devices?.softphones?.total || 0) - (stats.devices?.softphones?.online || 0) }} offline
              </span>
            </div>
          </div>
        </div>

        <div class="reg-card">
          <div class="reg-header">
            <div class="reg-icon-wrapper bg-amber-100">
              <PhoneIcon class="reg-icon text-amber-600" />
            </div>
            <span class="reg-title">Mobile Apps</span>
          </div>
          <div class="reg-stats">
            <div class="stat-num">{{ stats.devices?.mobile?.total || 0 }}</div>
            <div class="stat-details">
              <span class="online">
                <span class="dot"></span>
                {{ stats.devices?.mobile?.online || 0 }} online
              </span>
              <span class="offline">
                <span class="dot"></span>
                {{ (stats.devices?.mobile?.total || 0) - (stats.devices?.mobile?.online || 0) }} offline
              </span>
            </div>
          </div>
        </div>

        <div class="reg-card">
          <div class="reg-header">
            <div class="reg-icon-wrapper bg-purple-100">
              <ServerIcon class="reg-icon text-purple-600" />
            </div>
            <span class="reg-title">SIP Trunks</span>
          </div>
          <div class="reg-stats">
            <div class="stat-num">{{ stats.devices?.trunks?.total || 0 }}</div>
            <div class="stat-details">
              <span class="online">
                <span class="dot"></span>
                {{ stats.devices?.trunks?.online || 0 }} online
              </span>
              <span class="offline">
                <span class="dot"></span>
                {{ (stats.devices?.trunks?.total || 0) - (stats.devices?.trunks?.online || 0) }} offline
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="section">
      <div class="section-header">
        <h2 class="section-title">Quick Actions</h2>
        <span class="section-subtitle">Common system management tasks</span>
      </div>
      
      <div class="quick-actions">
        <button class="action-card" @click="$router.push('/system/tenants')">
          <div class="action-icon-wrapper">
            <PlusIcon class="action-icon" />
          </div>
          <div class="action-content">
            <span class="action-title">Create Tenant</span>
            <span class="action-desc">Add a new tenant organization</span>
          </div>
          <ArrowRightIcon class="action-arrow" />
        </button>

        <button class="action-card" @click="$router.push('/system/trunks')">
          <div class="action-icon-wrapper">
            <GlobeIcon class="action-icon" />
          </div>
          <div class="action-content">
            <span class="action-title">Add Gateway</span>
            <span class="action-desc">Configure a new SIP gateway</span>
          </div>
          <ArrowRightIcon class="action-arrow" />
        </button>

        <button class="action-card" @click="$router.push('/system/logs')">
          <div class="action-icon-wrapper">
            <FileTextIcon class="action-icon" />
          </div>
          <div class="action-content">
            <span class="action-title">View Logs</span>
            <span class="action-desc">Check system logs & events</span>
          </div>
          <ArrowRightIcon class="action-arrow" />
        </button>

        <button class="action-card" @click="$router.push('/system/settings')">
          <div class="action-icon-wrapper">
            <SettingsIcon class="action-icon" />
          </div>
          <div class="action-content">
            <span class="action-title">Global Settings</span>
            <span class="action-desc">Manage system configuration</span>
          </div>
          <ArrowRightIcon class="action-arrow" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import StatCard from '../../components/features/StatCard.vue'
import { systemAPI } from '../../services/api'
import { 
  Monitor as MonitorIcon, 
  Smartphone as SmartphoneIcon, 
  Phone as PhoneIcon, 
  Server as ServerIcon,
  Plus as PlusIcon,
  Globe as GlobeIcon,
  FileText as FileTextIcon,
  Settings as SettingsIcon,
  ArrowRight as ArrowRightIcon
} from 'lucide-vue-next'

// Stats data
const stats = ref({
  tenants: 0,
  users: 0,
  active_channels: 0,
  alerts: 0,
  esl_connected: false,
  registrations: 0,
  devices: {
    desk_phones: { total: 0, online: 0 },
    softphones: { total: 0, online: 0 },
    mobile: { total: 0, online: 0 },
    trunks: { total: 0, online: 0 }
  }
})

const formattedUsers = computed(() => {
  return stats.value.users.toLocaleString()
})

const loading = ref(true)
const error = ref(null)
let refreshInterval = null

async function loadStats() {
  try {
    const response = await systemAPI.getStats()
    stats.value = { ...stats.value, ...response.data }
    error.value = null
  } catch (e) {
    error.value = e.message
    console.error('Failed to load system stats:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadStats()
  // Auto-refresh every 30 seconds
  refreshInterval = setInterval(loadStats, 30000)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
</script>

<style scoped>
/* Status Indicator */
.status-indicator {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  background: var(--bg-card);
  padding: var(--spacing-2) var(--spacing-4);
  border-radius: var(--radius-full);
  border: 1px solid var(--border-color);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
}

.status-dot.online {
  background: var(--status-good);
  box-shadow: 0 0 0 2px var(--status-good-subtle);
}

.status-text {
  color: var(--text-primary);
}

/* Section Styles */
.section {
  margin-top: var(--spacing-8);
}

.section-header {
  margin-bottom: var(--spacing-5);
}

.section-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin: 0 0 var(--spacing-1) 0;
}

.section-subtitle {
  font-size: var(--text-sm);
  color: var(--text-secondary);
}

/* Registration Grid */
.registration-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--spacing-4);
}

.reg-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-5);
  transition: box-shadow var(--transition-fast);
}

.reg-card:hover {
  box-shadow: var(--shadow-md);
}

.reg-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  margin-bottom: var(--spacing-4);
}

.reg-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.reg-icon {
  width: 20px;
  height: 20px;
}

.reg-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.reg-stats {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-3);
}

.stat-num {
  font-size: var(--text-3xl);
  font-weight: var(--font-bold);
  color: var(--text-primary);
  line-height: var(--leading-none);
}

.stat-details {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-3);
  font-size: var(--text-xs);
}

.stat-details .online {
  color: var(--status-good);
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.stat-details .offline {
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: var(--spacing-1);
}

.stat-details .dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
  background: currentColor;
}

/* Quick Actions */
.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--spacing-4);
}

.action-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  padding: var(--spacing-4);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: left;
  width: 100%;
}

.action-card:hover {
  border-color: var(--primary-color);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.action-card:hover .action-arrow {
  transform: translateX(4px);
  color: var(--primary-color);
}

.action-icon-wrapper {
  width: 44px;
  height: 44px;
  background: var(--primary-subtle);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.action-icon {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
}

.action-content {
  flex: 1;
  min-width: 0;
}

.action-title {
  display: block;
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-0-5);
}

.action-desc {
  display: block;
  font-size: var(--text-xs);
  color: var(--text-secondary);
}

.action-arrow {
  width: 16px;
  height: 16px;
  color: var(--text-muted);
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

/* Mobile Responsive */
@media (max-width: 1024px) {
  .registration-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .quick-actions {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .section {
    margin-top: var(--spacing-6);
  }
  
  .registration-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-3);
  }
  
  .reg-card {
    padding: var(--spacing-4);
  }
  
  .stat-num {
    font-size: var(--text-2xl);
  }
  
  .quick-actions {
    grid-template-columns: 1fr;
  }
  
  .action-card {
    padding: var(--spacing-3);
  }
}

@media (max-width: 480px) {
  .registration-grid {
    grid-template-columns: 1fr;
  }
  
  .reg-header {
    margin-bottom: var(--spacing-3);
  }
  
  .stat-details {
    flex-direction: column;
    gap: var(--spacing-1);
  }
  
  .action-icon-wrapper {
    width: 40px;
    height: 40px;
  }
}
</style>
