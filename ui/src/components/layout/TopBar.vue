<template>
  <header class="topbar" :class="[mode, { impersonating: isImpersonating }]">
    <!-- Impersonation Banner -->
    <div v-if="isImpersonating" class="impersonate-banner">
      <EyeIcon class="banner-icon" />
      <span>Viewing as <strong>{{ impersonatedTenantName }}</strong></span>
      <button class="exit-btn" @click="exitImpersonation">Exit</button>
    </div>
    
    <!-- Left Section: Breadcrumb / Context -->
    <div class="topbar-left">
      <!-- Search Bar -->
      <div class="search-bar" :class="{ expanded: searchExpanded }">
        <SearchIcon class="search-icon" />
        <input 
          type="text" 
          v-model="searchQuery" 
          :placeholder="searchPlaceholder"
          @focus="searchExpanded = true"
          @blur="searchExpanded = false"
        >
        <span class="search-shortcut" v-if="!searchExpanded">⌘K</span>
      </div>
    </div>

    <!-- Center Section: Tenant Selector (Admin/System only) -->
    <div class="topbar-center" v-if="auth.permissions.isSystemAdmin() || auth.permissions.isTenantAdmin()">
      <div class="tenant-selector">
        <div class="tenant-badge" :class="{ system: selectedContext === 'system' }">
          <GlobeIcon v-if="selectedContext === 'system'" class="tenant-icon" />
          <BuildingIcon v-else class="tenant-icon" />
        </div>
        <select v-model="selectedContext" @change="handleContextChange" class="tenant-select">
          <optgroup label="System" v-if="auth.permissions.isSystemAdmin()">
            <option value="system">System Admin (Global)</option>
          </optgroup>
          <optgroup label="Tenants">
            <option v-for="tenant in auth.state.tenants" :key="tenant.id" :value="tenant.id">
              {{ tenant.name }}
            </option>
          </optgroup>
        </select>
        <ChevronDownIcon class="select-chevron" />
      </div>
    </div>

    <!-- Right Section: Quick Actions & User -->
    <div class="topbar-right">
      <!-- Quick Actions -->
      <div class="quick-actions">
        <button class="action-btn" @click="showHelp" title="Help & Docs">
          <HelpCircleIcon class="action-icon" />
        </button>
        
        <!-- Notification Center -->
        <NotificationCenter />

      <button class="action-btn" v-if="auth.permissions.isSystemAdmin() || auth.permissions.isTenantAdmin()" @click="showQuickAdd" title="Quick Add">
          <PlusCircleIcon class="action-icon" />
        </button>
      </div>

      <div class="divider"></div>

      <!-- Portal Links -->
      <div class="portal-links">
        <button 
          v-if="!auth.permissions.isSystemAdmin()"
          class="portal-btn" 
          :class="{ active: mode === 'user' }"
          @click="$router.push('/')"
          title="User Portal"
        >
          <PhoneIcon class="portal-icon" />
        </button>
        <button 
          class="portal-btn" 
          :class="{ active: mode === 'admin', disabled: auth.permissions.isSystemAdmin() && !auth.state.currentTenantId }"
          @click="navigateToTenantAdmin"
          title="Tenant Admin"
        >
          <LayoutDashboardIcon class="portal-icon" />
        </button>
        <button 
          v-if="auth.permissions.isSystemAdmin()"
          class="portal-btn system" 
          :class="{ active: mode === 'system' }"
          @click="$router.push('/system')"
          title="System Admin"
        >
          <ServerCogIcon class="portal-icon" />
        </button>
      </div>

      <div class="divider"></div>

      <!-- User Menu -->
      <div class="user-menu" @click="showUserDropdown = !showUserDropdown">
        <div class="user-avatar">
          <span>{{ userInitials }}</span>
          <span class="status-indicator online"></span>
        </div>
        <div class="user-info">
          <span class="user-name">{{ userName }}</span>
          <span class="user-role">{{ userRole }}</span>
        </div>
        <ChevronDownIcon class="menu-chevron" />
      </div>

      <!-- User Dropdown -->
      <div class="user-dropdown" v-if="showUserDropdown" @click.stop>
        <div class="dropdown-header">
          <div class="user-avatar large">
            <span>{{ userInitials }}</span>
          </div>
          <div class="dropdown-user-info">
            <span class="name">{{ userName }}</span>
            <span class="email">{{ auth.state.user?.email }}</span>
          </div>
        </div>
        <div class="dropdown-divider"></div>
        <router-link to="/settings" class="dropdown-item">
          <UserIcon class="dropdown-icon" />
          <span>My Account</span>
        </router-link>
        <router-link to="/settings" class="dropdown-item">
          <SettingsIcon class="dropdown-icon" />
          <span>Preferences</span>
        </router-link>
        <div class="dropdown-divider"></div>
        <button class="dropdown-item logout" @click="logout">
          <LogOutIcon class="dropdown-icon" />
          <span>Sign Out</span>
        </button>
      </div>
    </div>

    <!-- Backdrop for dropdown -->
    <div class="dropdown-backdrop" v-if="showUserDropdown" @click="showUserDropdown = false"></div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '../../services/auth'
import { 
  Search as SearchIcon, Globe as GlobeIcon, Building2 as BuildingIcon,
  ChevronDown as ChevronDownIcon, HelpCircle as HelpCircleIcon,
  PlusCircle as PlusCircleIcon, Phone as PhoneIcon,
  LayoutDashboard as LayoutDashboardIcon, ServerCog as ServerCogIcon,
  User as UserIcon, Settings as SettingsIcon, LogOut as LogOutIcon,
  Eye as EyeIcon
} from 'lucide-vue-next'
import NotificationCenter from '@/components/NotificationCenter.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuth()

const searchQuery = ref('')
const searchExpanded = ref(false)
const showUserDropdown = ref(false)

// Initialize selectedContext based on current tenant or system
const selectedContext = ref('system')

// Update selectedContext based on auth state
onMounted(async () => {
    await auth.fetchAvailableTenants()
    
    const tenantId = localStorage.getItem('tenantId')
    if (tenantId) {
        selectedContext.value = tenantId
    } else {
        selectedContext.value = 'system'
    }
})

const mode = computed(() => {
  if (route.path.startsWith('/system')) return 'system'
  if (route.path.startsWith('/admin')) return 'admin'
  return 'user'
})

const searchPlaceholder = computed(() => {
  if (mode.value === 'system') return 'Search tenants, users, settings...'
  if (mode.value === 'admin') return 'Search extensions, devices, routes...'
  return 'Search contacts, messages...'
})

const userName = computed(() => auth.state.user?.name || auth.state.user?.username || 'User')
const userInitials = computed(() => {
    const name = userName.value
    return name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase()
})

const userRole = computed(() => {
  const role = auth.state.user?.role
  if (role === 'system_admin') return 'System Admin'
  if (role === 'tenant_admin') return 'Tenant Admin'
  return 'User'
})

// Impersonation state (system admin viewing as specific tenant)
const isImpersonating = computed(() => {
  return auth.permissions.isSystemAdmin() && !!auth.state.currentTenantId
})

const impersonatedTenantName = computed(() => {
  if (!isImpersonating.value) return ''
  const tenant = auth.state.tenants.find(t => t.id == auth.state.currentTenantId)
  return tenant?.name || localStorage.getItem('tenantName') || 'Tenant'
})

const exitImpersonation = () => {
  auth.switchTenant('system')
}

const handleContextChange = () => {
  auth.switchTenant(selectedContext.value)
}

const navigateToTenantAdmin = () => {
  if (auth.permissions.isSystemAdmin() && !auth.state.currentTenantId) return
  router.push('/admin')
}

const showHelp = () => alert('Help & Documentation')
const showQuickAdd = () => alert('Quick Add Menu')
const logout = async () => {
  showUserDropdown.value = false
  await auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  padding: 0 20px;
  background: white;
  border-bottom: 1px solid var(--border-color);
  position: relative;
}

.topbar.system { background: linear-gradient(90deg, #fffbeb 0%, white 50%); }
.topbar.admin { background: linear-gradient(90deg, #f0fdf4 0%, white 50%); }
.topbar.impersonating { padding-top: 36px; }

/* Impersonation Banner */
.impersonate-banner {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: linear-gradient(90deg, #7c3aed, #6366f1);
  color: white;
  font-size: 12px;
  font-weight: 500;
}
.banner-icon { width: 14px; height: 14px; }
.exit-btn {
  padding: 2px 8px;
  background: rgba(255,255,255,0.2);
  border: 1px solid rgba(255,255,255,0.3);
  border-radius: 4px;
  color: white;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  margin-left: 12px;
}
.exit-btn:hover { background: rgba(255,255,255,0.3); }

/* Left Section */
.topbar-left { display: flex; align-items: center; gap: 16px; flex: 1; }

.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-app);
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 8px 12px;
  width: 280px;
  transition: all 0.2s ease;
}

.search-bar:focus-within,
.search-bar.expanded {
  background: white;
  border-color: var(--border-color);
  width: 360px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.search-icon { width: 16px; height: 16px; color: var(--text-muted); }
.search-bar input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 13px;
  outline: none;
}
.search-bar input::placeholder { color: var(--text-muted); }
.search-shortcut {
  font-size: 10px;
  padding: 2px 6px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-muted);
  font-family: monospace;
}

/* Center Section */
.topbar-center { display: flex; align-items: center; }

.tenant-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  position: relative;
  min-width: 220px;
}

.tenant-badge {
  width: 28px;
  height: 28px;
  background: #dcfce7;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #16a34a;
}
.tenant-badge.system { background: #fef3c7; color: #b45309; }
.tenant-icon { width: 14px; height: 14px; }

.tenant-select {
  flex: 1;
  appearance: none;
  background: transparent;
  border: none;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  outline: none;
  cursor: pointer;
  padding-right: 20px;
}

.select-chevron {
  position: absolute;
  right: 10px;
  width: 14px;
  height: 14px;
  color: var(--text-muted);
  pointer-events: none;
}

/* Right Section */
.topbar-right { display: flex; align-items: center; gap: 8px; }

.quick-actions { display: flex; align-items: center; gap: 4px; }

.action-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-muted);
  position: relative;
  transition: all 0.15s ease;
}
.action-btn:hover { background: var(--bg-app); color: var(--text-primary); }
.action-icon { width: 18px; height: 18px; }

.badge-dot {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 8px;
  height: 8px;
  background: #ef4444;
  border: 2px solid white;
  border-radius: 50%;
}

.divider { width: 1px; height: 28px; background: var(--border-color); margin: 0 8px; }

/* Portal Links */
.portal-links { display: flex; gap: 2px; background: var(--bg-app); border-radius: 8px; padding: 3px; }

.portal-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-muted);
  transition: all 0.15s ease;
}
.portal-btn:hover { color: var(--text-primary); }
.portal-btn.active { background: white; color: var(--primary-color); box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.portal-btn.system:hover { color: #b45309; }
.portal-btn.system.active { color: #b45309; }
.portal-icon { width: 16px; height: 16px; }

/* User Menu */
.user-menu {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.user-menu:hover { background: var(--bg-app); }

.user-avatar {
  width: 34px;
  height: 34px;
  background: linear-gradient(135deg, var(--primary-color), #818cf8);
  color: white;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  position: relative;
}
.user-avatar.large { width: 48px; height: 48px; font-size: 16px; border-radius: 12px; }

.status-indicator {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 10px;
  height: 10px;
  border: 2px solid white;
  border-radius: 50%;
}
.status-indicator.online { background: #22c55e; }
.status-indicator.away { background: #f59e0b; }
.status-indicator.busy { background: #ef4444; }

.user-info { display: flex; flex-direction: column; }
.user-name { font-size: 13px; font-weight: 600; color: var(--text-primary); line-height: 1.2; }
.user-role { font-size: 10px; color: var(--text-muted); }
.menu-chevron { width: 14px; height: 14px; color: var(--text-muted); }

/* User Dropdown */
.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 20px;
  width: 260px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.12);
  z-index: 100;
  overflow: hidden;
}

.dropdown-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-app);
}

.dropdown-user-info { display: flex; flex-direction: column; }
.dropdown-user-info .name { font-size: 14px; font-weight: 600; }
.dropdown-user-info .email { font-size: 12px; color: var(--text-muted); }

.dropdown-divider { height: 1px; background: var(--border-color); }

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  font-size: 13px;
  color: var(--text-main);
  text-decoration: none;
  cursor: pointer;
  border: none;
  background: transparent;
  width: 100%;
  text-align: left;
}
.dropdown-item:hover { background: var(--bg-app); }
.dropdown-item.logout { color: #dc2626; }
.dropdown-icon { width: 16px; height: 16px; }

.dropdown-backdrop {
  position: fixed;
  inset: 0;
  z-index: 99;
}

/* Mobile Responsiveness */
@media (max-width: 768px) {
  .topbar {
    padding: 0 12px;
    gap: 8px;
  }
  
  .search-bar {
    width: 140px;
    padding: 6px 10px;
  }
  .search-bar:focus-within,
  .search-bar.expanded {
    width: 180px;
  }
  .search-shortcut { display: none; }
  
  .tenant-selector {
    min-width: 150px;
    padding: 4px 8px;
  }
  .tenant-select { font-size: 12px; }
  
  .quick-actions .action-btn { display: none; }
  .quick-actions .action-btn:first-child { display: flex; }
  
  .portal-links { gap: 4px; }
  .portal-btn { padding: 4px 8px; font-size: 10px; }
  .portal-btn span { display: none; }
  
  .divider { margin: 0 4px; }
  
  .user-menu { padding: 4px 6px; gap: 6px; }
  .user-info { display: none; }
  .menu-chevron { display: none; }
  
  .user-dropdown { right: 10px; }
}

@media (max-width: 480px) {
  .topbar-center { display: none; }
  .search-bar { display: none; }
  .portal-links { display: none; }
}
</style>
