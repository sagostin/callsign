<template>
  <aside class="sidebar" :class="{ collapsed: isCollapsed }">
    <!-- Brand Area -->
    <div class="brand-area">
      <div class="logo-box">
        <span class="logo-text">C</span>
      </div>
      <h1 class="brand-title" v-if="!isCollapsed">Callsign</h1>
      <button 
        v-if="!isMobile"
        class="collapse-btn" 
        @click="$emit('toggle-collapse')"
        :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        aria-label="Toggle sidebar collapse"
      >
        <ChevronLeftIcon v-if="!isCollapsed" class="icon-sm" />
        <ChevronRightIcon v-else class="icon-sm" />
      </button>
    </div>

    <!-- Mode Indicator -->
    <div class="mode-badge" :class="mode" v-if="!isCollapsed">
      <component :is="modeIcon" class="mode-icon" />
      <span>{{ modeLabel }}</span>
    </div>
    <div class="mode-badge-icon" :class="mode" v-else>
      <component :is="modeIcon" class="mode-icon" />
    </div>
    
    <!-- Navigation Menu -->
    <nav class="nav-menu" aria-label="Main navigation">
      
      <!-- USER PORTAL MENU -->
      <template v-if="mode === 'user' && !auth.permissions.isSystemAdmin()">
        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Communication</div>
          <router-link to="/" class="nav-item" :title="isCollapsed ? 'Web Phone' : ''">
            <Phone class="nav-icon" />
            <span class="nav-label">Web Phone</span>
          </router-link>
          <router-link to="/messages" class="nav-item" :title="isCollapsed ? 'Messages' : ''">
            <MessageSquare class="nav-icon" />
            <span class="nav-label">Messages</span>
            <span class="nav-badge" v-if="!isCollapsed">3</span>
          </router-link>
          <router-link to="/voicemail" class="nav-item" :title="isCollapsed ? 'Voicemail' : ''">
            <Voicemail class="nav-icon" />
            <span class="nav-label">Voicemail</span>
            <span class="nav-badge" v-if="!isCollapsed">5</span>
          </router-link>
          <router-link to="/history" class="nav-item" :title="isCollapsed ? 'Call History' : ''">
            <ClockIcon class="nav-icon" />
            <span class="nav-label">Call History</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Conferencing</div>
          <router-link to="/conferences" class="nav-item" :title="isCollapsed ? 'My Conferences' : ''">
            <UsersIcon class="nav-icon" />
            <span class="nav-label">My Conferences</span>
          </router-link>
          <router-link to="/conferences/schedule" class="nav-item" :title="isCollapsed ? 'Schedule' : ''">
            <CalendarIcon class="nav-icon" />
            <span class="nav-label">Schedule Meeting</span>
          </router-link>
          <router-link to="/conferences/recordings" class="nav-item" :title="isCollapsed ? 'Recordings' : ''">
            <VideoIcon class="nav-icon" />
            <span class="nav-label">Meeting Recordings</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Resources</div>
          <router-link to="/contacts" class="nav-item" :title="isCollapsed ? 'Contacts' : ''">
            <ContactIcon class="nav-icon" />
            <span class="nav-label">Contacts</span>
          </router-link>
          <router-link to="/recordings" class="nav-item" :title="isCollapsed ? 'My Recordings' : ''">
            <Mic class="nav-icon" />
            <span class="nav-label">My Recordings</span>
          </router-link>
          <router-link to="/fax" class="nav-item" :title="isCollapsed ? 'Fax' : ''">
            <PrinterIcon class="nav-icon" />
            <span class="nav-label">Fax</span>
          </router-link>
          <router-link to="/settings" class="nav-item" :title="isCollapsed ? 'My Settings' : ''">
            <Settings class="nav-icon" />
            <span class="nav-label">My Settings</span>
          </router-link>
        </div>

        <div class="nav-spacer"></div>
        <div class="nav-section bottom">
          <router-link to="/admin" class="nav-item portal-link" :title="isCollapsed ? 'Admin Portal' : ''">
            <LayoutDashboard class="nav-icon" />
            <span class="nav-label">Admin Portal</span>
            <ArrowRightIcon class="arrow-icon" v-if="!isCollapsed" />
          </router-link>
        </div>
      </template>

      <!-- TENANT ADMIN MENU -->
      <template v-else-if="mode === 'admin'">
        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Overview</div>
          <router-link to="/admin" class="nav-item" :title="isCollapsed ? 'Dashboard' : ''">
            <LayoutDashboard class="nav-icon" />
            <span class="nav-label">Dashboard</span>
          </router-link>
          <router-link to="/admin/cdr" class="nav-item" :title="isCollapsed ? 'Call History' : ''">
            <PhoneCallIcon class="nav-icon" />
            <span class="nav-label">Call History</span>
          </router-link>
          <router-link to="/admin/reports" class="nav-item" :title="isCollapsed ? 'Reports' : ''">
            <BarChart3 class="nav-icon" />
            <span class="nav-label">Reports</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Users & Devices</div>
          <router-link to="/admin/extensions" class="nav-item" :title="isCollapsed ? 'Extensions' : ''">
            <Phone class="nav-icon" />
            <span class="nav-label">Extensions</span>
          </router-link>
          <router-link to="/admin/devices" class="nav-item" :title="isCollapsed ? 'Devices' : ''">
            <MonitorSmartphone class="nav-icon" />
            <span class="nav-label">Devices</span>
          </router-link>
          <router-link to="/admin/device-profiles" class="nav-item sub-item" :title="isCollapsed ? 'Device Profiles' : ''">
            <LayersIcon class="nav-icon" />
            <span class="nav-label">Device Profiles</span>
          </router-link>
          <router-link to="/admin/provisioning" class="nav-item sub-item" :title="isCollapsed ? 'Provisioning' : ''">
            <SettingsIcon class="nav-icon" />
            <span class="nav-label">Provisioning</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Call Routing</div>
          <router-link to="/admin/routing" class="nav-item" :title="isCollapsed ? 'DIDs & Routing' : ''">
            <GitMerge class="nav-icon" />
            <span class="nav-label">DIDs & Routing</span>
          </router-link>
          <router-link to="/admin/routing/debug" class="nav-item sub-item" :title="isCollapsed ? 'Routing Debugger' : ''">
            <SearchIcon class="nav-icon" />
            <span class="nav-label">Routing Debugger</span>
          </router-link>
          <router-link to="/admin/feature-codes" class="nav-item" :title="isCollapsed ? 'Feature Codes' : ''">
            <Hash class="nav-icon" />
            <span class="nav-label">Feature Codes</span>
          </router-link>
          <router-link to="/admin/call-flows" class="nav-item" :title="isCollapsed ? 'Call Flows' : ''">
            <GitMerge class="nav-icon" />
            <span class="nav-label">Call Flows</span>
          </router-link>
          <router-link to="/admin/ivr" class="nav-item" :title="isCollapsed ? 'IVR Menus' : ''">
            <MenuIcon class="nav-icon" />
            <span class="nav-label">IVR Menus</span>
          </router-link>
          <router-link to="/admin/queues" class="nav-item" :title="isCollapsed ? 'Queues' : ''">
            <GalleryVerticalEnd class="nav-icon" />
            <span class="nav-label">Queues & Groups</span>
          </router-link>
          <router-link to="/admin/time-conditions" class="nav-item" :title="isCollapsed ? 'Time Conditions' : ''">
            <ClockIcon class="nav-icon" />
            <span class="nav-label">Time Conditions</span>
          </router-link>
          <router-link to="/admin/toggles" class="nav-item" :title="isCollapsed ? 'Toggles' : ''">
            <ToggleIcon class="nav-icon" />
            <span class="nav-label">Toggles</span>
          </router-link>
          <router-link to="/admin/speed-dials" class="nav-item" :title="isCollapsed ? 'Speed Dials' : ''">
            <ZapIcon class="nav-icon" />
            <span class="nav-label">Speed Dials</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Communications</div>
          <router-link to="/admin/conferences" class="nav-item" :title="isCollapsed ? 'Conferences' : ''">
            <UsersIcon class="nav-icon" />
            <span class="nav-label">Conferences</span>
          </router-link>
          <router-link to="/admin/voicemail-manager" class="nav-item" :title="isCollapsed ? 'Voicemail' : ''">
            <Voicemail class="nav-icon" />
            <span class="nav-label">Voicemail</span>
          </router-link>
          <router-link to="/admin/fax" class="nav-item" :title="isCollapsed ? 'Fax' : ''">
            <PrinterIcon class="nav-icon" />
            <span class="nav-label">Fax Server</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Media & Recordings</div>
          <router-link to="/admin/call-recordings" class="nav-item" :title="isCollapsed ? 'Call Recordings' : ''">
            <PlayCircle class="nav-icon" />
            <span class="nav-label">Call Recordings</span>
          </router-link>
          <router-link to="/admin/audio-library" class="nav-item" :title="isCollapsed ? 'Audio Library' : ''">
            <Mic class="nav-icon" />
            <span class="nav-label">Audio Library</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Settings</div>
          <router-link to="/admin/trunks" class="nav-item" :title="isCollapsed ? 'Trunks' : ''">
            <Server class="nav-icon" />
            <span class="nav-label">Trunks</span>
          </router-link>
          <router-link to="/admin/bridges" class="nav-item" :title="isCollapsed ? 'Bridges' : ''">
            <NetworkIcon class="nav-icon" />
            <span class="nav-label">Bridges</span>
          </router-link>
          <router-link to="/admin/call-block" class="nav-item" :title="isCollapsed ? 'Call Block' : ''">
            <Shield class="nav-icon" />
            <span class="nav-label">Call Block</span>
          </router-link>
          <router-link to="/admin/feature-codes" class="nav-item" :title="isCollapsed ? 'Feature Codes' : ''">
            <Hash class="nav-icon" />
            <span class="nav-label">Feature Codes</span>
          </router-link>
          <router-link to="/admin/hospitality" class="nav-item" :title="isCollapsed ? 'Hospitality' : ''">
            <Hotel class="nav-icon" />
            <span class="nav-label">Hospitality</span>
          </router-link>
          <router-link to="/admin/settings" class="nav-item" :title="isCollapsed ? 'Tenant Settings' : ''">
            <Settings class="nav-icon" />
            <span class="nav-label">Tenant Settings</span>
          </router-link>
        </div>

        <div class="nav-spacer"></div>
        <div class="nav-section bottom">
          <router-link to="/" class="nav-item portal-link" v-if="!auth.permissions.isSystemAdmin()" :title="isCollapsed ? 'User Portal' : ''">
            <ArrowLeftIcon class="nav-icon" />
            <span class="nav-label">User Portal</span>
          </router-link>
        </div>
      </template>

      <!-- SYSTEM ADMIN MENU -->
      <template v-else-if="mode === 'system'">
        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Overview</div>
          <router-link to="/system" class="nav-item" :title="isCollapsed ? 'Dashboard' : ''">
            <LayoutDashboard class="nav-icon" />
            <span class="nav-label">Dashboard</span>
          </router-link>
          <router-link to="/system/logs" class="nav-item" :title="isCollapsed ? 'System Logs' : ''">
            <TerminalIcon class="nav-icon" />
            <span class="nav-label">System Logs</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Tenants</div>
          <router-link to="/system/tenants" class="nav-item" :title="isCollapsed ? 'Tenants' : ''">
            <Building class="nav-icon" />
            <span class="nav-label">Tenants</span>
          </router-link>
          <router-link to="/system/profiles" class="nav-item" :title="isCollapsed ? 'Tenant Profiles' : ''">
            <LayersIcon class="nav-icon" />
            <span class="nav-label">Tenant Profiles</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Media</div>
          <router-link to="/system/sounds" class="nav-item" :title="isCollapsed ? 'Sounds' : ''">
            <VolumeIcon class="nav-icon" />
            <span class="nav-label">Sounds</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Devices</div>
          <router-link to="/system/provisioning-templates" class="nav-item" :title="isCollapsed ? 'Templates' : ''">
            <FileCodeIcon class="nav-icon" />
            <span class="nav-label">Device Templates</span>
          </router-link>
          <router-link to="/system/firmware" class="nav-item" :title="isCollapsed ? 'Firmware' : ''">
            <DownloadIcon class="nav-icon" />
            <span class="nav-label">Firmware Updates</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">Infrastructure</div>
          <router-link to="/system/infrastructure" class="nav-item" :title="isCollapsed ? 'Servers' : ''">
            <Server class="nav-icon" />
            <span class="nav-label">Servers</span>
          </router-link>
          <router-link to="/system/sip-profiles" class="nav-item" :title="isCollapsed ? 'SIP Profiles' : ''">
            <NetworkIcon class="nav-icon" />
            <span class="nav-label">SIP Profiles</span>
          </router-link>
          <router-link to="/system/acls" class="nav-item" :title="isCollapsed ? 'Access Control' : ''">
            <Shield class="nav-icon" />
            <span class="nav-label">Access Control</span>
          </router-link>
          <router-link to="/system/trunks" class="nav-item" :title="isCollapsed ? 'Trunks' : ''">
            <GlobeIcon class="nav-icon" />
            <span class="nav-label">Trunks</span>
          </router-link>
          <router-link to="/system/routing" class="nav-item" :title="isCollapsed ? 'System Routing' : ''">
            <GitMerge class="nav-icon" />
            <span class="nav-label">System Routing</span>
          </router-link>
          <router-link to="/system/messaging" class="nav-item" :title="isCollapsed ? 'Messaging' : ''">
            <MessageSquare class="nav-icon" />
            <span class="nav-label">Messaging Providers</span>
          </router-link>
          <router-link to="/system/config-inspector" class="nav-item" :title="isCollapsed ? 'Config Inspector' : ''">
            <ContainerIcon class="nav-icon" />
            <span class="nav-label">Config Inspector</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">System</div>
          <router-link to="/system/audit-log" class="nav-item" :title="isCollapsed ? 'Audit Log' : ''">
            <Shield class="nav-icon" />
            <span class="nav-label">Audit Log</span>
          </router-link>
          <router-link to="/system/settings" class="nav-item" :title="isCollapsed ? 'Global Settings' : ''">
            <Settings class="nav-icon" />
            <span class="nav-label">Global Settings</span>
          </router-link>
          <router-link to="/system/security" class="nav-item" :title="isCollapsed ? 'Security' : ''">
            <ShieldIcon class="nav-icon" />
            <span class="nav-label">Security</span>
          </router-link>
        </div>

        <div class="nav-spacer"></div>
      </template>

    </nav>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuth } from '../../services/auth'
import { 
  LayoutDashboard, Phone, GitMerge, Hash, MonitorSmartphone, 
  Users, 
  Users as UsersIcon, MessageSquare, BarChart3, Settings, Server, Shield, Voicemail, Printer as PrinterIcon,
  Clock as ClockIcon, Music, Mic, GalleryVerticalEnd, Hotel, Building, PlayCircle,
  ChevronLeft as ChevronLeftIcon, ChevronRight as ChevronRightIcon,
  ArrowRight as ArrowRightIcon, ArrowLeft as ArrowLeftIcon,
  Calendar as CalendarIcon, Video as VideoIcon, Contact as ContactIcon,
  Menu as MenuIcon, Network as NetworkIcon, ServerCog as ServerCogIcon,
  Terminal as TerminalIcon, Layers as LayersIcon, FileCode as FileCodeIcon,
  Globe as GlobeIcon, Zap as ZapIcon, ToggleLeft as ToggleIcon, PhoneCall as PhoneCallIcon,
  Download as DownloadIcon, Volume2 as VolumeIcon, FileAudio as FileAudioIcon,
  Shield as ShieldIcon, Settings as SettingsIcon, Search as SearchIcon, Container as ContainerIcon
} from 'lucide-vue-next'

const props = defineProps({
  collapsed: {
    type: Boolean,
    default: false
  }
})

defineEmits(['navigated', 'toggle-collapse'])

const route = useRoute()
const auth = useAuth()

const isCollapsed = computed(() => props.collapsed)
const isMobile = computed(() => window.innerWidth <= 768)

const mode = computed(() => {
  if (route.path.startsWith('/system')) return 'system'
  if (route.path.startsWith('/admin')) return 'admin'
  return 'user'
})

const modeLabel = computed(() => {
  if (mode.value === 'system') return 'System Admin'
  if (mode.value === 'admin') return 'Tenant Admin'
  return 'User Portal'
})

const modeIcon = computed(() => {
  if (mode.value === 'system') return ServerCogIcon
  if (mode.value === 'admin') return LayoutDashboard
  return Phone
})
</script>

<style scoped>
.sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  width: 260px;
  transition: width var(--transition-slow);
  overflow: hidden;
}

.sidebar.collapsed {
  width: 64px;
}

.sidebar.collapsed .nav-label,
.sidebar.collapsed .nav-header,
.sidebar.collapsed .nav-badge,
.sidebar.collapsed .arrow-icon { 
  display: none; 
}

/* Brand Area */
.brand-area {
  height: var(--header-height);
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-3);
  gap: var(--spacing-2-5);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.logo-box {
  width: 36px;
  height: 36px;
  min-width: 36px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: white;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-bold);
  font-size: 18px;
  box-shadow: var(--shadow-sm);
}

.logo-text {
  background: linear-gradient(135deg, #fff 0%, #e0e7ff 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-title {
  color: var(--text-primary);
  font-size: var(--text-lg);
  font-weight: var(--font-bold);
  letter-spacing: var(--tracking-tight);
  white-space: nowrap;
}

.collapse-btn {
  margin-left: auto;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: var(--bg-hover);
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--text-muted);
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.collapse-btn:hover { 
  background: var(--border-color); 
  color: var(--text-primary); 
}

.collapse-btn:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}

/* Mode Badge */
.mode-badge {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  margin: var(--spacing-3);
  padding: var(--spacing-2) var(--spacing-3);
  border-radius: var(--radius-md);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  background: var(--bg-hover);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.mode-badge.user { 
  background: var(--user-portal-bg); 
  color: var(--user-portal-color); 
}

.mode-badge.admin { 
  background: var(--tenant-admin-bg); 
  color: var(--tenant-admin-color); 
}

.mode-badge.system { 
  background: var(--system-admin-bg); 
  color: var(--system-admin-color); 
}

.mode-icon { 
  width: 16px; 
  height: 16px; 
}

.mode-badge-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: var(--spacing-3) auto;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-lg);
  background: var(--bg-hover);
  color: var(--text-secondary);
}

.mode-badge-icon.user { background: var(--user-portal-bg); color: var(--user-portal-color); }
.mode-badge-icon.admin { background: var(--tenant-admin-bg); color: var(--tenant-admin-color); }
.mode-badge-icon.system { background: var(--system-admin-bg); color: var(--system-admin-color); }

/* Nav Menu */
.nav-menu {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: var(--spacing-2);
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: thin;
  scrollbar-color: var(--border-color) transparent;
}

.nav-menu::-webkit-scrollbar {
  width: 4px;
}

.nav-menu::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: var(--radius-full);
}

.nav-section { 
  margin-bottom: var(--spacing-1); 
}

.nav-section.bottom { 
  margin-top: auto; 
  margin-bottom: 0; 
  padding-top: var(--spacing-2); 
  border-top: 1px solid var(--border-color); 
}

.nav-header {
  font-size: var(--text-2xs);
  font-weight: var(--font-bold);
  color: var(--text-muted);
  padding: var(--spacing-3) var(--spacing-3) var(--spacing-1-5);
  letter-spacing: var(--tracking-widest);
  text-transform: uppercase;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-3);
  padding: var(--spacing-2) var(--spacing-3);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
  text-decoration: none;
  white-space: nowrap;
  margin-bottom: var(--spacing-0-5);
}

.nav-item:hover {
  background-color: var(--bg-hover);
  color: var(--text-primary);
}

.nav-item.router-link-active,
.nav-item.router-link-exact-active {
  background-color: var(--primary-subtle);
  color: var(--primary-color);
  font-weight: var(--font-semibold);
}

.nav-item.sub-item {
  padding-left: var(--spacing-8);
  font-size: var(--text-xs);
}

.nav-icon { 
  width: 18px; 
  height: 18px; 
  min-width: 18px; 
  stroke-width: 2px; 
}

.nav-label { 
  flex: 1; 
}

.nav-badge {
  background: var(--status-bad);
  color: white;
  font-size: var(--text-2xs);
  font-weight: var(--font-bold);
  padding: var(--spacing-0-5) var(--spacing-1-5);
  border-radius: var(--radius-full);
  min-width: 18px;
  text-align: center;
}

.portal-link {
  color: var(--text-muted);
  border: 1px dashed var(--border-color);
}

.portal-link:hover { 
  border-color: var(--primary-color); 
  color: var(--primary-color); 
}

.portal-link.system { 
  border-style: solid; 
  background: var(--bg-hover); 
}

.arrow-icon { 
  width: 14px; 
  height: 14px; 
  opacity: 0.5; 
}

.nav-spacer { 
  flex: 1; 
}

.icon-sm { 
  width: 16px; 
  height: 16px; 
}

/* Mobile Styles */
@media (max-width: 768px) {
  .sidebar {
    width: 280px;
  }
  
  .sidebar.collapsed {
    width: 280px;
  }
  
  .sidebar.collapsed .nav-label,
  .sidebar.collapsed .nav-header,
  .sidebar.collapsed .nav-badge,
  .sidebar.collapsed .arrow-icon {
    display: block;
  }
  
  .mode-badge-icon {
    display: none;
  }
  
  .mode-badge {
    display: flex;
    margin: var(--spacing-3);
  }
  
  .nav-item {
    padding: var(--spacing-3);
  }
}

@media (max-width: 480px) {
  .sidebar {
    width: 100%;
    max-width: 320px;
  }
  
  .nav-item {
    padding: var(--spacing-3) var(--spacing-4);
  }
  
  .nav-item.sub-item {
    padding-left: var(--spacing-10);
  }
}
</style>
