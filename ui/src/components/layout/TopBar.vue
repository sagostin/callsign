<template>
  <header class="topbar" :class="[mode, { impersonating: isImpersonating }]" role="banner">
    <!-- Impersonation Banner -->
    <div v-if="isImpersonating" class="impersonate-banner">
      <EyeIcon class="banner-icon" />
      <span>Viewing as <strong>{{ impersonatedTenantName }}</strong></span>
      <button class="exit-btn" @click="exitImpersonation">Exit</button>
    </div>
    
    <!-- Left Section: Search -->
    <div class="topbar-left">
      <div class="search-bar" :class="{ expanded: searchExpanded }">
        <SearchIcon class="search-icon" />
        <input 
          type="text" 
          v-model="searchQuery" 
          :placeholder="searchPlaceholder"
          @focus="searchExpanded = true"
          @blur="searchExpanded = false"
          aria-label="Search"
        >
        <span class="search-shortcut" v-if="!searchExpanded">⌘K</span>
      </div>
    </div>

    <!-- Center Section: Tenant Selector -->
    <div class="topbar-center" v-if="auth.permissions.isSystemAdmin() || auth.permissions.isTenantAdmin()">
      <div class="tenant-selector">
        <div class="tenant-badge" :class="{ system: selectedContext === 'system' }">
          <GlobeIcon v-if="selectedContext === 'system'" class="tenant-icon" />
          <BuildingIcon v-else class="tenant-icon" />
        </div>
        <select v-model="selectedContext" @change="handleContextChange" class="tenant-select" aria-label="Select context">
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

    <!-- Right Section: Actions & User -->
    <div class="topbar-right">
      <!-- Quick Actions -->
      <div class="quick-actions">
        <button class="action-btn" @click="showHelp" title="Help & Docs" aria-label="Help and documentation">
          <HelpCircleIcon class="action-icon" />
        </button>
        
        <NotificationCenter />

        <button 
          class="action-btn" 
          v-if="auth.permissions.isSystemAdmin() || auth.permissions.isTenantAdmin()" 
          @click="showQuickAdd" 
          title="Quick Add"
          aria-label="Quick add"
        >
          <PlusCircleIcon class="action-icon" />
        </button>
      </div>

      <!-- Portal Links -->
      <div class="portal-links">
        <button 
          v-if="!auth.permissions.isSystemAdmin()"
          class="portal-btn" 
          :class="{ active: mode === 'user' }"
          @click="$router.push('/')"
          title="User Portal"
          aria-label="User portal"
        >
          <PhoneIcon class="portal-icon" />
          <span class="portal-label">User</span>
        </button>
        <button 
          class="portal-btn" 
          :class="{ active: mode === 'admin', disabled: auth.permissions.isSystemAdmin() && !auth.state.currentTenantId }"
          @click="navigateToTenantAdmin"
          title="Tenant Admin"
          aria-label="Tenant admin"
        >
          <LayoutDashboardIcon class="portal-icon" />
          <span class="portal-label">Admin</span>
        </button>
        <button 
          v-if="auth.permissions.isSystemAdmin()"
          class="portal-btn system" 
          :class="{ active: mode === 'system' }"
          @click="$router.push('/system')"
          title="System Admin"
          aria-label="System admin"
        >
          <ServerCogIcon class="portal-icon" />
          <span class="portal-label">System</span>
        </button>
      </div>

      <!-- User Menu -->
      <div class="user-menu" @click="showUserDropdown = !showUserDropdown" role="button" tabindex="0" aria-haspopup="true" :aria-expanded="showUserDropdown">
        <div class="user-avatar">
          <span>{{ userInitials }}</span>
          <span class="status-indicator online"></span>
        </div>
        <div class="user-info">
          <span class="user-name">{{ userName }}</span>
          <span class="user-role">{{ userRole }}</span>
        </div>
        <ChevronDownIcon class="menu-chevron" :class="{ open: showUserDropdown }" />
      </div>

      <!-- User Dropdown -->
      <Transition name="dropdown">
        <div class="user-dropdown" v-if="showUserDropdown" @click.stop v-click-outside="() => showUserDropdown = false">
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
          <router-link to="/settings" class="dropdown-item" @click="showUserDropdown = false">
            <UserIcon class="dropdown-icon" />
            <span>My Account</span>
          </router-link>
          <router-link to="/settings" class="dropdown-item" @click="showUserDropdown = false">
            <SettingsIcon class="dropdown-icon" />
            <span>Preferences</span>
          </router-link>
          <div class="dropdown-divider"></div>
          <button class="dropdown-item logout" @click="logout">
            <LogOutIcon class="dropdown-icon" />
            <span>Sign Out</span>
          </button>
        </div>
      </Transition>
    </div>

    <!-- Backdrop for dropdown -->
    <div class="dropdown-backdrop" v-if="showUserDropdown" @click="showUserDropdown = false"></div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
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

// Click outside directive
const vClickOutside = {
  mounted(el, binding) {
    el._clickOutside = (event) => {
      if (!(el === event.target || el.contains(event.target))) {
        binding.value()
      }
    }
    document.addEventListener('click', el._clickOutside)
  },
  unmounted(el) {
    document.removeEventListener('click', el._clickOutside)
  }
}

const route = useRoute()
const router = useRouter()
const auth = useAuth()

const searchQuery = ref('')
const searchExpanded = ref(false)
const showUserDropdown = ref(false)

// Keyboard shortcut for search
const handleKeydown = (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    document.querySelector('.search-bar input')?.focus()
  }
  if (e.key === 'Escape') {
    showUserDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

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
  const isAdmin = auth.hasRole(['system_admin', 'tenant_admin'])
  await auth.logout()
  router.push(isAdmin ? '/admin/login' : '/login')
}
</script>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  padding: 0 var(--spacing-5);
  background: var(--bg-header);
  border-bottom: 1px solid var(--border-color);
  position: relative;
  gap: var(--spacing-4);
}

.topbar.system { 
  background: linear-gradient(90deg, var(--system-admin-bg) 0%, var(--bg-header) 50%); 
}

.topbar.admin { 
  background: linear-gradient(90deg, var(--tenant-admin-bg) 0%, var(--bg-header) 50%); 
}

.topbar.impersonating { 
  padding-top: 32px; 
}

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
  gap: var(--spacing-2);
  background: linear-gradient(90deg, #7c3aed, #6366f1);
  color: white;
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
}

.banner-icon { 
  width: 14px; 
  height: 14px; 
}

.exit-btn {
  padding: var(--spacing-0-5) var(--spacing-2);
  background: rgba(255,255,255,0.2);
  border: 1px solid rgba(255,255,255,0.3);
  border-radius: var(--radius-sm);
  color: white;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  cursor: pointer;
  margin-left: var(--spacing-2);
  transition: all var(--transition-fast);
}

.exit-btn:hover { 
  background: rgba(255,255,255,0.3); 
}

/* Left Section */
.topbar-left { 
  display: flex; 
  align-items: center; 
  flex: 1;
  min-width: 0;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  background: var(--bg-hover);
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  padding: var(--spacing-2) var(--spacing-3);
  width: 260px;
  max-width: 100%;
  transition: all var(--transition-fast);
}

.search-bar:focus-within,
.search-bar.expanded {
  background: var(--bg-card);
  border-color: var(--border-color);
  width: 320px;
  box-shadow: var(--shadow-sm);
}

.search-icon { 
  width: 16px; 
  height: 16px; 
  color: var(--text-muted); 
  flex-shrink: 0; 
}

.search-bar input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: var(--text-sm);
  outline: none;
  min-width: 0;
  color: var(--text-primary);
}

.search-bar input::placeholder { 
  color: var(--text-muted); 
}

.search-shortcut {
  font-size: var(--text-2xs);
  padding: var(--spacing-0-5) var(--spacing-1-5);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-family: var(--font-mono);
  flex-shrink: 0;
}

/* Center Section */
.topbar-center { 
  display: flex; 
  align-items: center; 
  flex-shrink: 0;
}

.tenant-selector {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) var(--spacing-3) var(--spacing-1) var(--spacing-2);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  position: relative;
  min-width: 160px;
  transition: all var(--transition-fast);
}

.tenant-selector:hover { 
  border-color: var(--border-hover); 
}

.tenant-badge {
  width: 28px;
  height: 28px;
  background: var(--tenant-admin-bg);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--tenant-admin-color);
  flex-shrink: 0;
}

.tenant-badge.system { 
  background: var(--system-admin-bg); 
  color: var(--system-admin-color); 
}

.tenant-icon { 
  width: 14px; 
  height: 14px; 
}

.tenant-select {
  flex: 1;
  appearance: none;
  background: transparent;
  border: none;
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  outline: none;
  cursor: pointer;
  padding-right: var(--spacing-5);
  min-width: 0;
}

.select-chevron {
  position: absolute;
  right: var(--spacing-2);
  width: 14px;
  height: 14px;
  color: var(--text-muted);
  pointer-events: none;
}

/* Right Section */
.topbar-right { 
  display: flex; 
  align-items: center; 
  gap: var(--spacing-3);
  flex-shrink: 0;
}

/* Quick Actions */
.quick-actions { 
  display: flex; 
  align-items: center; 
  gap: var(--spacing-0-5); 
}

.action-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.action-btn:hover { 
  background: var(--bg-hover); 
  color: var(--text-primary); 
}

.action-btn:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}

.action-icon { 
  width: 18px; 
  height: 18px; 
}

/* Portal Links */
.portal-links { 
  display: flex; 
  gap: var(--spacing-0-5); 
  background: var(--bg-hover); 
  border-radius: var(--radius-lg); 
  padding: var(--spacing-0-5); 
}

.portal-btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-1-5);
  padding: var(--spacing-1) var(--spacing-2);
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
}

.portal-btn:hover { 
  color: var(--text-primary); 
}

.portal-btn.active { 
  background: var(--bg-card); 
  color: var(--primary-color); 
  box-shadow: var(--shadow-sm); 
}

.portal-btn.system:hover { 
  color: var(--system-admin-color); 
}

.portal-btn.system.active { 
  color: var(--system-admin-color); 
}

.portal-btn.disabled { 
  opacity: 0.4; 
  cursor: not-allowed; 
}

.portal-icon { 
  width: 16px; 
  height: 16px; 
}

.portal-label {
  display: inline;
}

/* User Menu */
.user-menu {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  padding: var(--spacing-1) var(--spacing-2);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.user-menu:hover { 
  background: var(--bg-hover); 
}

.user-menu:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}

.user-avatar {
  width: 34px;
  height: 34px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-xs);
  font-weight: var(--font-bold);
  position: relative;
  flex-shrink: 0;
}

.user-avatar.large { 
  width: 48px; 
  height: 48px; 
  font-size: var(--text-base); 
  border-radius: var(--radius-lg); 
}

.status-indicator {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 10px;
  height: 10px;
  border: 2px solid var(--bg-card);
  border-radius: var(--radius-full);
}

.status-indicator.online { 
  background: var(--status-good); 
}

.status-indicator.away { 
  background: var(--status-warn); 
}

.status-indicator.busy { 
  background: var(--status-bad); 
}

.user-info { 
  display: flex; 
  flex-direction: column; 
  min-width: 0; 
}

.user-name { 
  font-size: var(--text-sm); 
  font-weight: var(--font-semibold); 
  color: var(--text-primary); 
  line-height: var(--leading-tight);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role { 
  font-size: var(--text-xs); 
  color: var(--text-muted); 
}

.menu-chevron { 
  width: 14px; 
  height: 14px; 
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}

.menu-chevron.open {
  transform: rotate(180deg);
}

/* User Dropdown */
.user-dropdown {
  position: absolute;
  top: calc(100% + var(--spacing-2));
  right: var(--spacing-4);
  width: 280px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  z-index: var(--z-dropdown);
  overflow: hidden;
}

.dropdown-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-4);
  background: var(--bg-hover);
}

.dropdown-user-info { 
  display: flex; 
  flex-direction: column;
  min-width: 0;
}

.dropdown-user-info .name { 
  font-size: var(--text-sm); 
  font-weight: var(--font-semibold);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-user-info .email { 
  font-size: var(--text-xs); 
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-divider { 
  height: 1px; 
  background: var(--border-color); 
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-2-5);
  padding: var(--spacing-2-5) var(--spacing-4);
  font-size: var(--text-sm);
  color: var(--text-main);
  text-decoration: none;
  cursor: pointer;
  border: none;
  background: transparent;
  width: 100%;
  text-align: left;
  transition: background var(--transition-fast);
}

.dropdown-item:hover { 
  background: var(--bg-hover); 
}

.dropdown-item.logout { 
  color: var(--status-bad); 
}

.dropdown-icon { 
  width: 16px; 
  height: 16px; 
  flex-shrink: 0;
}

.dropdown-backdrop {
  position: fixed;
  inset: 0;
  z-index: calc(var(--z-dropdown) - 1);
}

/* Dropdown Animation */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all var(--transition-fast);
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Mobile Responsiveness */
@media (max-width: 1024px) {
  .topbar {
    padding: 0 var(--spacing-4);
    gap: var(--spacing-3);
  }
  
  .search-bar {
    width: 200px;
  }
  
  .search-bar:focus-within,
  .search-bar.expanded {
    width: 240px;
  }
  
  .portal-label {
    display: none;
  }
  
  .portal-btn {
    padding: var(--spacing-1);
  }
}

@media (max-width: 768px) {
  .topbar {
    padding: 0 var(--spacing-3);
    gap: var(--spacing-2);
  }
  
  .search-bar {
    width: 140px;
    padding: var(--spacing-1-5) var(--spacing-2);
  }
  
  .search-bar:focus-within,
  .search-bar.expanded {
    width: 180px;
  }
  
  .search-shortcut { 
    display: none; 
  }
  
  .tenant-selector {
    min-width: 140px;
    padding: var(--spacing-1) var(--spacing-2) var(--spacing-1) var(--spacing-1);
  }
  
  .tenant-select { 
    font-size: var(--text-xs); 
  }
  
  .quick-actions .action-btn:not(:last-child) { 
    display: none; 
  }
  
  .user-info { 
    display: none; 
  }
  
  .menu-chevron { 
    display: none; 
  }
  
  .user-dropdown { 
    right: var(--spacing-2);
    width: 260px;
  }
}

@media (max-width: 480px) {
  .topbar-center { 
    display: none; 
  }
  
  .search-bar { 
    display: none; 
  }
  
  .portal-links { 
    gap: var(--spacing-0-5);
  }
  
  .portal-btn {
    padding: var(--spacing-1);
  }
  
  .quick-actions {
    display: none;
  }
  
  .user-menu {
    padding: var(--spacing-0-5);
  }
  
  .user-avatar {
    width: 32px;
    height: 32px;
  }
}
</style>
