<template>
  <aside class="sidebar" :class="{ collapsed: isCollapsed }">
    <div class="brand-area">
      <div class="logo-box">C</div>
      <h1 class="brand-title" v-if="!isCollapsed">Callsign</h1>
      <button class="collapse-btn" @click="isCollapsed = !isCollapsed">
        <ChevronLeftIcon v-if="!isCollapsed" class="icon-sm" />
        <ChevronRightIcon v-else class="icon-sm" />
      </button>
    </div>

    <!-- Mode Indicator -->
    <div class="mode-badge" :class="mode">
      <component :is="modeIcon" class="mode-icon" />
      <span v-if="!isCollapsed">{{ modeLabel }}</span>
    </div>
    
    <nav class="nav-menu">
      
      <!-- USER PORTAL MENU -->
      <template v-if="mode === 'user' && !auth.permissions.isSystemAdmin()">
        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">COMMUNICATION</div>
          <router-link to="/" class="nav-item" v-tooltip="isCollapsed ? 'Web Phone' : ''">
            <Phone class="nav-icon" />
            <span class="nav-label">Web Phone</span>
          </router-link>
          <router-link to="/messages" class="nav-item" v-tooltip="isCollapsed ? 'Messages' : ''">
            <MessageSquare class="nav-icon" />
            <span class="nav-label">Messages</span>
            <span class="nav-badge" v-if="!isCollapsed">3</span>
          </router-link>
          <router-link to="/voicemail" class="nav-item" v-tooltip="isCollapsed ? 'Voicemail' : ''">
            <Voicemail class="nav-icon" />
            <span class="nav-label">Voicemail</span>
            <span class="nav-badge" v-if="!isCollapsed">5</span>
          </router-link>
          <router-link to="/history" class="nav-item" v-tooltip="isCollapsed ? 'Call History' : ''">
            <ClockIcon class="nav-icon" />
            <span class="nav-label">Call History</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">CONFERENCING</div>
          <router-link to="/conferences" class="nav-item" v-tooltip="isCollapsed ? 'My Conferences' : ''">
            <UsersIcon class="nav-icon" />
            <span class="nav-label">My Conferences</span>
          </router-link>
          <router-link to="/conferences/schedule" class="nav-item" v-tooltip="isCollapsed ? 'Schedule' : ''">
            <CalendarIcon class="nav-icon" />
            <span class="nav-label">Schedule Meeting</span>
          </router-link>
          <router-link to="/conferences/recordings" class="nav-item" v-tooltip="isCollapsed ? 'Recordings' : ''">
            <VideoIcon class="nav-icon" />
            <span class="nav-label">Meeting Recordings</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">RESOURCES</div>
          <router-link to="/contacts" class="nav-item" v-tooltip="isCollapsed ? 'Contacts' : ''">
            <ContactIcon class="nav-icon" />
            <span class="nav-label">Contacts</span>
          </router-link>
          <router-link to="/recordings" class="nav-item" v-tooltip="isCollapsed ? 'My Recordings' : ''">
            <Mic class="nav-icon" />
            <span class="nav-label">My Recordings</span>
          </router-link>
          <router-link to="/fax" class="nav-item" v-tooltip="isCollapsed ? 'Fax' : ''">
            <PrinterIcon class="nav-icon" />
            <span class="nav-label">Fax</span>
          </router-link>
          <router-link to="/settings" class="nav-item" v-tooltip="isCollapsed ? 'My Settings' : ''">
            <Settings class="nav-icon" />
            <span class="nav-label">My Settings</span>
          </router-link>
        </div>

        <div class="nav-spacer"></div>
        <div class="nav-section bottom">
          <router-link to="/admin" class="nav-item portal-link" v-tooltip="isCollapsed ? 'Admin Portal' : ''">
            <LayoutDashboard class="nav-icon" />
            <span class="nav-label">Admin Portal</span>
            <ArrowRightIcon class="arrow-icon" v-if="!isCollapsed" />
          </router-link>
        </div>
      </template>

      <!-- TENANT ADMIN MENU -->
      <template v-else-if="mode === 'admin'">
        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">OVERVIEW</div>
          <router-link to="/admin" class="nav-item" v-tooltip="isCollapsed ? 'Dashboard' : ''">
            <LayoutDashboard class="nav-icon" />
            <span class="nav-label">Dashboard</span>
          </router-link>
          <router-link to="/admin/cdr" class="nav-item" v-tooltip="isCollapsed ? 'Call History' : ''">
            <PhoneCallIcon class="nav-icon" />
            <span class="nav-label">Call History</span>
          </router-link>
          <router-link to="/admin/reports" class="nav-item" v-tooltip="isCollapsed ? 'Reports' : ''">
            <BarChart3 class="nav-icon" />
            <span class="nav-label">Reports</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">USERS & DEVICES</div>
          <router-link to="/admin/extensions" class="nav-item" v-tooltip="isCollapsed ? 'Extensions' : ''">
            <Phone class="nav-icon" />
            <span class="nav-label">Extensions</span>
          </router-link>
          <router-link to="/admin/devices" class="nav-item" v-tooltip="isCollapsed ? 'Devices' : ''">

            <MonitorSmartphone class="nav-icon" />
            <span class="nav-label">Devices</span>
          </router-link>
          <router-link to="/admin/device-profiles" class="nav-item sub-item" v-tooltip="isCollapsed ? 'Device Profiles' : ''">
            <LayersIcon class="nav-icon" />
            <span class="nav-label">Device Profiles</span>
          </router-link>
          <router-link to="/admin/provisioning" class="nav-item sub-item" v-tooltip="isCollapsed ? 'Provisioning' : ''">
            <SettingsIcon class="nav-icon" />
            <span class="nav-label">Provisioning</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">CALL ROUTING</div>
          <router-link to="/admin/routing" class="nav-item" v-tooltip="isCollapsed ? 'DIDs & Routing' : ''">
            <GitMerge class="nav-icon" />
            <span class="nav-label">DIDs & Routing</span>
          </router-link>
          <router-link to="/admin/routing/debug" class="nav-item sub-item" v-tooltip="isCollapsed ? 'Routing Debugger' : ''">
            <SearchIcon class="nav-icon" />
            <span class="nav-label">Routing Debugger</span>
          </router-link>
          <router-link to="/admin/feature-codes" class="nav-item" v-tooltip="isCollapsed ? 'Feature Codes' : ''">
            <Hash class="nav-icon" />
            <span class="nav-label">Feature Codes</span>
          </router-link>
          <router-link to="/admin/call-flows" class="nav-item" v-tooltip="isCollapsed ? 'Call Flows' : ''">
            <GitMerge class="nav-icon" />
            <span class="nav-label">Call Flows</span>
          </router-link>
          <router-link to="/admin/ivr" class="nav-item" v-tooltip="isCollapsed ? 'IVR Menus' : ''">
            <MenuIcon class="nav-icon" />
            <span class="nav-label">IVR Menus</span>
          </router-link>
          <router-link to="/admin/queues" class="nav-item" v-tooltip="isCollapsed ? 'Queues' : ''">
            <GalleryVerticalEnd class="nav-icon" />
            <span class="nav-label">Queues & Groups</span>
          </router-link>
          <router-link to="/admin/time-conditions" class="nav-item" v-tooltip="isCollapsed ? 'Time Conditions' : ''">
            <ClockIcon class="nav-icon" />
            <span class="nav-label">Time Conditions</span>
          </router-link>
          <router-link to="/admin/toggles" class="nav-item" v-tooltip="isCollapsed ? 'Toggles' : ''">
            <ToggleIcon class="nav-icon" />
            <span class="nav-label">Toggles</span>
          </router-link>
          <router-link to="/admin/speed-dials" class="nav-item" v-tooltip="isCollapsed ? 'Speed Dials' : ''">
            <ZapIcon class="nav-icon" />
            <span class="nav-label">Speed Dials</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">COMMUNICATIONS</div>
          <router-link to="/admin/conferences" class="nav-item" v-tooltip="isCollapsed ? 'Conferences' : ''">
            <UsersIcon class="nav-icon" />
            <span class="nav-label">Conferences</span>
          </router-link>
          <router-link to="/admin/voicemail-manager" class="nav-item" v-tooltip="isCollapsed ? 'Voicemail' : ''">
            <Voicemail class="nav-icon" />
            <span class="nav-label">Voicemail</span>
          </router-link>
          <router-link to="/admin/fax" class="nav-item" v-tooltip="isCollapsed ? 'Fax' : ''">
            <PrinterIcon class="nav-icon" />
            <span class="nav-label">Fax Server</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">MEDIA & RECORDINGS</div>
          <router-link to="/admin/call-recordings" class="nav-item" v-tooltip="isCollapsed ? 'Call Recordings' : ''">
            <PlayCircle class="nav-icon" />
            <span class="nav-label">Call Recordings</span>
          </router-link>
          <router-link to="/admin/audio-library" class="nav-item" v-tooltip="isCollapsed ? 'Audio Library' : ''">
            <Mic class="nav-icon" />
            <span class="nav-label">Audio Library</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">SETTINGS</div>
          <router-link to="/admin/trunks" class="nav-item" v-tooltip="isCollapsed ? 'Trunks' : ''">
            <Server class="nav-icon" />
            <span class="nav-label">Trunks</span>
          </router-link>
          <router-link to="/admin/bridges" class="nav-item" v-tooltip="isCollapsed ? 'Bridges' : ''">
            <NetworkIcon class="nav-icon" />
            <span class="nav-label">Bridges</span>
          </router-link>
          <router-link to="/admin/call-block" class="nav-item" v-tooltip="isCollapsed ? 'Call Block' : ''">
            <Shield class="nav-icon" />
            <span class="nav-label">Call Block</span>
          </router-link>
          <router-link to="/admin/feature-codes" class="nav-item" v-tooltip="isCollapsed ? 'Feature Codes' : ''">
            <Hash class="nav-icon" />
            <span class="nav-label">Feature Codes</span>
          </router-link>
          <router-link to="/admin/hospitality" class="nav-item" v-tooltip="isCollapsed ? 'Hospitality' : ''">
            <Hotel class="nav-icon" />
            <span class="nav-label">Hospitality</span>
          </router-link>
          <router-link to="/admin/settings" class="nav-item" v-tooltip="isCollapsed ? 'Tenant Settings' : ''">
            <Settings class="nav-icon" />
            <span class="nav-label">Tenant Settings</span>
          </router-link>
        </div>

        <div class="nav-spacer"></div>
        <div class="nav-section bottom">
          <router-link to="/" class="nav-item portal-link" v-if="!auth.permissions.isSystemAdmin()" v-tooltip="isCollapsed ? 'User Portal' : ''">
            <ArrowLeftIcon class="nav-icon" />
            <span class="nav-label">User Portal</span>
          </router-link>
        </div>
      </template>

      <!-- SYSTEM ADMIN MENU -->
      <template v-else-if="mode === 'system'">
        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">OVERVIEW</div>
          <router-link to="/system" class="nav-item" v-tooltip="isCollapsed ? 'Dashboard' : ''">
            <LayoutDashboard class="nav-icon" />
            <span class="nav-label">Dashboard</span>
          </router-link>
          <router-link to="/system/logs" class="nav-item" v-tooltip="isCollapsed ? 'System Logs' : ''">
            <TerminalIcon class="nav-icon" />
            <span class="nav-label">System Logs</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">TENANTS</div>
          <router-link to="/system/tenants" class="nav-item" v-tooltip="isCollapsed ? 'Tenants' : ''">
            <Building class="nav-icon" />
            <span class="nav-label">Tenants</span>
          </router-link>
          <router-link to="/system/profiles" class="nav-item" v-tooltip="isCollapsed ? 'Tenant Profiles' : ''">
            <LayersIcon class="nav-icon" />
            <span class="nav-label">Tenant Profiles</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">MEDIA</div>
          <router-link to="/system/sounds" class="nav-item" v-tooltip="isCollapsed ? 'Sounds' : ''">
            <VolumeIcon class="nav-icon" />
            <span class="nav-label">Sounds</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">DEVICES</div>
          <router-link to="/system/provisioning-templates" class="nav-item" v-tooltip="isCollapsed ? 'Templates' : ''">
            <FileCodeIcon class="nav-icon" />
            <span class="nav-label">Device Templates</span>
          </router-link>
          <router-link to="/system/firmware" class="nav-item" v-tooltip="isCollapsed ? 'Firmware' : ''">
            <DownloadIcon class="nav-icon" />
            <span class="nav-label">Firmware Updates</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">INFRASTRUCTURE</div>
          <router-link to="/system/infrastructure" class="nav-item" v-tooltip="isCollapsed ? 'Servers' : ''">
            <Server class="nav-icon" />
            <span class="nav-label">Servers</span>
          </router-link>
          <router-link to="/system/sip-profiles" class="nav-item" v-tooltip="isCollapsed ? 'SIP Profiles' : ''">
            <NetworkIcon class="nav-icon" />
            <span class="nav-label">SIP Profiles</span>
          </router-link>
          <router-link to="/system/acls" class="nav-item" v-tooltip="isCollapsed ? 'Access Control' : ''">
            <Shield class="nav-icon" />
            <span class="nav-label">Access Control</span>
          </router-link>
          <router-link to="/system/trunks" class="nav-item" v-tooltip="isCollapsed ? 'Trunks' : ''">
            <GlobeIcon class="nav-icon" />
            <span class="nav-label">Trunks</span>
          </router-link>
          <router-link to="/system/routing" class="nav-item" v-tooltip="isCollapsed ? 'System Routing' : ''">
            <GitMerge class="nav-icon" />
            <span class="nav-label">System Routing</span>
          </router-link>
          <router-link to="/system/messaging" class="nav-item" v-tooltip="isCollapsed ? 'Messaging' : ''">
            <MessageSquare class="nav-icon" />
            <span class="nav-label">Messaging Providers</span>
          </router-link>
          <router-link to="/system/config-inspector" class="nav-item" v-tooltip="isCollapsed ? 'Config Inspector' : ''">
            <ContainerIcon class="nav-icon" />
            <span class="nav-label">Config Inspector</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-header" v-if="!isCollapsed">SYSTEM</div>
          <router-link to="/system/audit-log" class="nav-item" v-tooltip="isCollapsed ? 'Audit Log' : ''">
            <Shield class="nav-icon" />
            <span class="nav-label">Audit Log</span>
          </router-link>
          <router-link to="/system/settings" class="nav-item" v-tooltip="isCollapsed ? 'Global Settings' : ''">
            <Settings class="nav-icon" />
            <span class="nav-label">Global Settings</span>
          </router-link>
          <router-link to="/system/security" class="nav-item" v-tooltip="isCollapsed ? 'Security' : ''">
            <ShieldIcon class="nav-icon" />
            <span class="nav-label">Security</span>
          </router-link>
        </div>

        <div class="nav-spacer"></div>
        <!-- "Back to Admin" link removed as per user request to use Top Nav for context switching -->
      </template>

    </nav>
  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const emit = defineEmits(['navigated'])

const isCollapsed = ref(false)

// Check if a tenant is actively selected (for System Admins)
const hasSelectedTenant = computed(() => {
    return !!localStorage.getItem('tenantId')
})

router.afterEach(() => {
   emit('navigated')
})

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

// Simple tooltip directive (v-tooltip)
const vTooltip = {
  mounted(el, binding) {
    if (!binding.value) return
    el.setAttribute('title', binding.value)
  }
}
</script>

<style scoped>
.sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: white;
  border-right: 1px solid var(--border-color);
  width: 240px;
  transition: width 0.2s ease;
  overflow: hidden;
}

.sidebar.collapsed {
  width: 64px;
}

.sidebar.collapsed .nav-label,
.sidebar.collapsed .nav-header,
.sidebar.collapsed .nav-badge,
.sidebar.collapsed .arrow-icon { display: none; }

.brand-area {
  height: var(--header-height);
  display: flex;
  align-items: center;
  padding: 0 12px;
  gap: 10px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.logo-box {
  width: 36px;
  height: 36px;
  min-width: 36px;
  background: linear-gradient(135deg, var(--primary-color), #818cf8);
  color: white;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 18px;
}

.brand-title {
  color: var(--text-primary);
  font-size: 1.1rem;
  font-weight: 700;
  letter-spacing: -0.02em;
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
  background: var(--bg-app);
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-muted);
  flex-shrink: 0;
}
.collapse-btn:hover { background: var(--border-color); color: var(--text-primary); }

/* Mode Badge */
.mode-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 12px 12px 8px;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  background: var(--bg-app);
  color: var(--text-muted);
}
.mode-badge.user { background: #dbeafe; color: #1d4ed8; }
.mode-badge.admin { background: #dcfce7; color: #16a34a; }
.mode-badge.system { background: #fef3c7; color: #b45309; }
.mode-icon { width: 16px; height: 16px; }

.sidebar.collapsed .mode-badge {
  margin: 12px 8px 8px;
  padding: 8px;
  justify-content: center;
}

/* Nav Menu */
.nav-menu {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 8px;
  overflow-y: auto;
  overflow-x: hidden;
}

.nav-section { margin-bottom: 8px; }
.nav-section.bottom { margin-top: auto; margin-bottom: 0; padding-top: 8px; border-top: 1px solid var(--border-color); }

.nav-header {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-muted);
  padding: 12px 12px 6px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  color: var(--text-main);
  font-size: 13px;
  font-weight: 500;
  border-radius: 8px;
  transition: all 0.15s ease;
  text-decoration: none;
  white-space: nowrap;
}

.nav-item:hover {
  background-color: var(--bg-app);
  color: var(--text-primary);
}

.nav-item.router-link-active,
.nav-item.router-link-exact-active {
  background-color: var(--primary-light);
  color: var(--primary-color);
  font-weight: 600;
}

.nav-icon { width: 18px; height: 18px; min-width: 18px; stroke-width: 2px; }
.nav-label { flex: 1; }
.nav-badge {
  background: #ef4444;
  color: white;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

.portal-link {
  color: var(--text-muted);
  border: 1px dashed var(--border-color);
}
.portal-link:hover { border-color: var(--primary-color); color: var(--primary-color); }
.portal-link.system { border-style: solid; background: var(--bg-app); }

.arrow-icon { width: 14px; height: 14px; opacity: 0.5; }

.nav-spacer { flex: 1; }

.icon-sm { width: 16px; height: 16px; }
</style>
