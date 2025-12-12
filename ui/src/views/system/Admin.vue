<template>
  <div class="view-header">
    <div class="header-content">
      <h2>System Dashboard</h2>
      <p class="text-muted text-sm">Cluster overview and status.</p>
    </div>
    <div class="header-actions">
       <div class="status-indicator">
          <span class="dot online"></span>
          <span class="label">System Online</span>
       </div>
    </div>
  </div>

  <div class="dashboard-grid">
     <!-- Stats Cards -->
     <div class="stat-card">
        <div class="icon-box bg-indigo-100 text-indigo-600">
           <Building class="w-6 h-6" />
        </div>
        <div class="stat-info">
           <span class="label">Total Tenants</span>
           <span class="value">{{ stats.tenants }}</span>
        </div>
     </div>
     <div class="stat-card">
        <div class="icon-box bg-emerald-100 text-emerald-600">
           <Users class="w-6 h-6" />
        </div>
        <div class="stat-info">
           <span class="label">Active Users</span>
           <span class="value">{{ stats.users.toLocaleString() }}</span>
        </div>
     </div>
     <div class="stat-card">
        <div class="icon-box bg-amber-100 text-amber-600">
           <Phone class="w-6 h-6" />
        </div>
        <div class="stat-info">
           <span class="label">Active Channels</span>
           <span class="value">{{ stats.active_channels }}</span>
        </div>
     </div>
     <div class="stat-card">
        <div class="icon-box bg-rose-100 text-rose-600">
           <AlertTriangle class="w-6 h-6" />
        </div>
        <div class="stat-info">
           <span class="label">System Alerts</span>
           <span class="value">{{ stats.alerts }}</span>
        </div>
     </div>
  </div>

  <!-- Registration Stats -->
  <div class="section mt-8">
     <h3 class="section-title">Device Registrations</h3>
     <div class="registration-grid">
        <div class="reg-card">
           <div class="reg-header">
              <MonitorIcon class="reg-icon" />
              <span class="reg-title">Desk Phones</span>
           </div>
           <div class="reg-stats">
              <div class="stat-num">{{ stats.devices?.desk_phones?.total || 0 }}</div>
              <div class="stat-details">
                 <span class="online"><span class="dot"></span> {{ stats.devices?.desk_phones?.online || 0 }} online</span>
                 <span class="offline"><span class="dot"></span> {{ (stats.devices?.desk_phones?.total || 0) - (stats.devices?.desk_phones?.online || 0) }} offline</span>
              </div>
           </div>
        </div>
        <div class="reg-card">
           <div class="reg-header">
              <SmartphoneIcon class="reg-icon" />
              <span class="reg-title">Softphones</span>
           </div>
           <div class="reg-stats">
              <div class="stat-num">{{ stats.devices?.softphones?.total || 0 }}</div>
              <div class="stat-details">
                 <span class="online"><span class="dot"></span> {{ stats.devices?.softphones?.online || 0 }} online</span>
                 <span class="offline"><span class="dot"></span> {{ (stats.devices?.softphones?.total || 0) - (stats.devices?.softphones?.online || 0) }} offline</span>
              </div>
           </div>
        </div>
        <div class="reg-card">
           <div class="reg-header">
              <PhoneIcon class="reg-icon" />
              <span class="reg-title">Mobile Apps</span>
           </div>
           <div class="reg-stats">
              <div class="stat-num">{{ stats.devices?.mobile?.total || 0 }}</div>
              <div class="stat-details">
                 <span class="online"><span class="dot"></span> {{ stats.devices?.mobile?.online || 0 }} online</span>
                 <span class="offline"><span class="dot"></span> {{ (stats.devices?.mobile?.total || 0) - (stats.devices?.mobile?.online || 0) }} offline</span>
              </div>
           </div>
        </div>
        <div class="reg-card">
           <div class="reg-header">
              <ServerIcon class="reg-icon" />
              <span class="reg-title">SIP Trunks</span>
           </div>
           <div class="reg-stats">
              <div class="stat-num">{{ stats.devices?.trunks?.total || 0 }}</div>
              <div class="stat-details">
                 <span class="online"><span class="dot"></span> {{ stats.devices?.trunks?.online || 0 }} online</span>
                 <span class="offline"><span class="dot"></span> {{ (stats.devices?.trunks?.total || 0) - (stats.devices?.trunks?.online || 0) }} offline</span>
              </div>
           </div>
        </div>
     </div>
  </div>

  <div class="section mt-8">
     <h3 class="section-title">Quick Actions</h3>
     <div class="quick-actions">
        <button class="action-btn" @click="$router.push('/system/tenants')">
           <span class="icon">+</span>
           Create Tenant
        </button>
        <button class="action-btn" @click="$router.push('/admin/gateways/new')">
           <span class="icon">+</span>
           Add Gateway
        </button>
         <button class="action-btn" @click="$router.push('/system/logs')">
           <span class="icon">></span>
           View Logs
        </button>
     </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Building, Users, Phone, AlertTriangle, Monitor as MonitorIcon, Smartphone as SmartphoneIcon, Phone as PhoneIcon, Server as ServerIcon } from 'lucide-vue-next'
import { systemAPI } from '../services/api'

// Stats data
const stats = ref({
  tenants: 0,
  users: 0,
  active_channels: 0,
  alerts: 0,
  devices: {
    desk_phones: { total: 0, online: 0 },
    softphones: { total: 0, online: 0 },
    mobile: { total: 0, online: 0 },
    trunks: { total: 0, online: 0 }
  }
})

const loading = ref(true)
const error = ref(null)

onMounted(async () => {
  try {
    const response = await systemAPI.getStats()
    stats.value = response.data
  } catch (e) {
    error.value = e.message
    console.error('Failed to load system stats:', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xl);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  background: white;
  padding: 6px 12px;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  font-size: 13px;
  font-weight: 500;
}
.dot { width: 8px; height: 8px; border-radius: 50%; }
.dot.online { background: #22c55e; box-shadow: 0 0 0 2px #dcfce7; }

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 24px;
}

.stat-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: var(--shadow-sm);
}

.icon-box {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-info .label { font-size: 13px; color: var(--text-muted); font-weight: 500; }
.stat-info .value { font-size: 24px; font-weight: 700; color: var(--text-primary); line-height: 1.2; }

.section-title { font-size: 16px; font-weight: 600; margin-bottom: 16px; color: var(--text-primary); }

.quick-actions {
  display: flex;
  gap: 16px;
}

.action-btn {
  background: white;
  border: 1px solid var(--border-color);
  padding: 12px 20px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}
.action-btn:hover { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn .icon { font-weight: bold; font-size: 16px; }

.mt-8 { margin-top: 32px; }

/* Registration Grid */
.registration-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.reg-card {
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
}

.reg-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.reg-icon {
  width: 20px;
  height: 20px;
  color: var(--primary-color);
}

.reg-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
}

.reg-stats .stat-num {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: 8px;
}

.stat-details {
  display: flex;
  gap: 12px;
  font-size: 12px;
}

.stat-details .online { color: #22c55e; display: flex; align-items: center; gap: 4px; }
.stat-details .offline { color: #94a3b8; display: flex; align-items: center; gap: 4px; }
.stat-details .dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; }
</style>
