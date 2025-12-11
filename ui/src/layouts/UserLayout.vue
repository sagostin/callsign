<template>
  <div class="user-layout" @click="closeDropdowns">
    <aside class="mini-sidebar">
      <div class="brand-mark">C</div>
      <nav class="user-nav">
        <router-link to="/dialer" class="nav-icon" title="Dialer">
          <Phone class="icon" />
        </router-link>
        <router-link to="/messages" class="nav-icon" title="Messages">
          <MessageSquare class="icon" />
        </router-link>
        
        <router-link to="/voicemail" class="nav-icon" title="Voicemail">
          <Voicemail class="icon" />
        </router-link>

        <router-link to="/conferences" class="nav-icon" title="Conferences">
          <UsersRound class="icon" />
        </router-link>

        <router-link to="/fax" class="nav-icon" title="My Faxes">
          <Printer class="icon" />
        </router-link>

        <router-link to="/contacts" class="nav-icon" title="Contacts">
          <Users class="icon" />
        </router-link>

        <router-link to="/recordings" class="nav-icon" title="My Recordings">
          <MicIcon class="icon" />
        </router-link>

        <router-link to="/history" class="nav-icon" title="Call History">
          <Clock class="icon" />
        </router-link>

        <router-link to="/settings" class="nav-icon" title="Settings">
          <SettingsIcon class="icon" />
        </router-link>
      </nav>
      <div class="bottom-actions">
        <!-- PROFILE DROPDOWN TRIGGER -->
        <div class="user-avatar-container" @click.stop="toggleProfileMenu">
            <div class="user-avatar">JS</div>
            <div v-if="showProfileMenu" class="dropdown-menu profile-menu">
                <div class="menu-header">
                    <div class="user-name">John Smith</div>
                    <div class="user-ext">Ext. 101</div>
                </div>
                <div class="menu-divider"></div>
                <router-link to="/settings" class="menu-item">
                    <User class="icon-xs" /> Profile
                </router-link>
                <router-link to="/settings" class="menu-item">
                    <SettingsIcon class="icon-xs" /> Settings
                </router-link>
                <div class="menu-divider"></div>
                <div class="menu-item text-bad">
                    <LogOut class="icon-xs" /> Log Out
                </div>
            </div>
        </div>
      </div>
    </aside>

    <main class="main-content">
      <header class="user-header">
        
        <!-- STATUS SELECTOR -->
        <div class="status-container" @click.stop="toggleStatusMenu">
             <div class="status-pill" :class="currentStatus">
                <div class="status-dot"></div>
                <span class="status-text">{{ statusLabel }}</span>
                <ChevronDown class="icon-xs chevron" />
             </div>
             
             <div v-if="showStatusMenu" class="dropdown-menu status-menu">
                 <div class="menu-section-header">Set Status</div>
                 <div class="menu-item" @click="setStatus('available')" :class="{ active: currentStatus === 'available' }">
                     <div class="dot available"></div> Available
                     <CheckIcon v-if="currentStatus === 'available'" class="check-icon" />
                 </div>
                 <div class="menu-item" @click="setStatus('away')" :class="{ active: currentStatus === 'away' }">
                     <div class="dot away"></div> Away
                     <CheckIcon v-if="currentStatus === 'away'" class="check-icon" />
                 </div>
                 <div class="menu-item" @click="setStatus('dnd')" :class="{ active: currentStatus === 'dnd' }">
                     <div class="dot dnd"></div> Do Not Disturb
                     <CheckIcon v-if="currentStatus === 'dnd'" class="check-icon" />
                 </div>
                 <div class="menu-divider"></div>
                 <div class="menu-section-header">Custom Status</div>
                 <div class="custom-status-input">
                   <input v-model="customStatusText" placeholder="What's your status?" @keyup.enter="setCustomStatus">
                   <button class="custom-status-btn" @click="setCustomStatus" :disabled="!customStatusText">Set</button>
                 </div>
                 <div class="preset-statuses">
                   <button class="preset-btn" @click="setPresetStatus('In a meeting')">🗓️ In a meeting</button>
                   <button class="preset-btn" @click="setPresetStatus('On a call')">📞 On a call</button>
                   <button class="preset-btn" @click="setPresetStatus('Be right back')">⏰ Be right back</button>
                   <button class="preset-btn" @click="setPresetStatus('Out sick')">🤒 Out sick</button>
                   <button class="preset-btn" @click="setPresetStatus('Working remotely')">🏠 Working remotely</button>
                 </div>
                 <div v-if="currentStatus === 'custom'" class="clear-status">
                   <button class="btn-link" @click="clearCustomStatus">Clear custom status</button>
                 </div>
             </div>
        </div>

        <div class="search-bar">
          <Search class="icon-sm" />
          <input type="text" placeholder="Search contacts or dial..." />
        </div>
        
        <!-- QUEUE LOGIN DROPDOWN -->
        <div class="queue-login-container" @click.stop="toggleQueueMenu">
           <button class="queue-trigger" :class="{ 'has-active': loggedInQueues.length > 0 }">
             <Headphones class="queue-icon" />
             <span class="queue-label">
               <template v-if="loggedInQueues.length === 0">Queue Login</template>
               <template v-else>{{ loggedInQueues.length }} Queue{{ loggedInQueues.length > 1 ? 's' : '' }}</template>
             </span>
             <span class="queue-status-dot" v-if="loggedInQueues.length > 0" :class="agentStatus"></span>
             <ChevronDown class="icon-xs" />
           </button>

           <div v-if="showQueueMenu" class="dropdown-menu queue-menu">
             <!-- Agent Status -->
             <div class="agent-status-section">
               <div class="section-header">Agent Status</div>
               <div class="agent-status-buttons">
                 <button 
                   class="status-btn" 
                   :class="{ active: agentStatus === 'available' }"
                   @click="setAgentStatus('available')"
                 >
                   <div class="dot available"></div> Available
                 </button>
                 <button 
                   class="status-btn" 
                   :class="{ active: agentStatus === 'on-break' }"
                   @click="setAgentStatus('on-break')"
                 >
                   <div class="dot on-break"></div> On Break
                 </button>
                 <button 
                   class="status-btn" 
                   :class="{ active: agentStatus === 'wrap-up' }"
                   @click="setAgentStatus('wrap-up')"
                 >
                   <div class="dot wrap-up"></div> Wrap-Up
                 </button>
               </div>
             </div>

             <div class="menu-divider"></div>

             <!-- Queue List -->
             <div class="queue-list-section">
               <div class="section-header">Available Queues</div>
               <div class="queue-list">
                 <div 
                   class="queue-item" 
                   v-for="queue in availableQueues" 
                   :key="queue.id"
                   :class="{ active: isLoggedInto(queue.id) }"
                   @click="toggleQueueLogin(queue)"
                 >
                   <div class="queue-info">
                     <span class="queue-name">{{ queue.name }}</span>
                     <span class="queue-stats">
                       <span class="waiting" v-if="queue.waiting > 0">{{ queue.waiting }} waiting</span>
                       <span class="agents">{{ queue.agents }} agents</span>
                     </span>
                   </div>
                   <div class="queue-toggle">
                     <div class="toggle-switch" :class="{ on: isLoggedInto(queue.id) }">
                       <div class="toggle-handle"></div>
                     </div>
                   </div>
                 </div>
               </div>
             </div>

             <div class="menu-divider"></div>

             <!-- Quick Actions -->
             <div class="queue-actions">
               <button class="action-btn" @click="loginAllQueues" :disabled="loggedInQueues.length === availableQueues.length">
                 <LoginIcon class="icon-xs" /> Login All
               </button>
               <button class="action-btn" @click="logoutAllQueues" :disabled="loggedInQueues.length === 0">
                 <LogOut class="icon-xs" /> Logout All
               </button>
             </div>
           </div>
        </div>
      </header>
      <div class="content-body">
        <router-view></router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { 
  Phone, MessageSquare, Users, UsersRound, Clock, Search, Headphones, 
  ChevronDown, Voicemail, User, Settings as SettingsIcon, LogOut, Printer, 
  Mic as MicIcon, Check as CheckIcon, LogIn as LoginIcon
} from 'lucide-vue-next'

// --- STATUS LOGIC ---
const showStatusMenu = ref(false)
const currentStatus = ref('available') // available, away, dnd, custom
const customStatusText = ref('')

const statusLabel = computed(() => {
    switch(currentStatus.value) {
        case 'available': return 'Available'
        case 'away': return 'Away'
        case 'dnd': return 'Do Not Disturb'
        case 'custom': return customStatusText.value || 'Custom Status'
        default: return 'Available'
    }
})

const toggleStatusMenu = () => {
    showStatusMenu.value = !showStatusMenu.value
    showProfileMenu.value = false
    showQueueMenu.value = false
}

const setStatus = (status) => {
    currentStatus.value = status
    showStatusMenu.value = false
}

const setPresetStatus = (text) => {
    customStatusText.value = text
    currentStatus.value = 'custom'
    showStatusMenu.value = false
}

const setCustomStatus = () => {
    if (customStatusText.value) {
        currentStatus.value = 'custom'
        showStatusMenu.value = false
    }
}

const clearCustomStatus = () => {
    customStatusText.value = ''
    currentStatus.value = 'available'
}

// --- PROFILE LOGIC ---
const showProfileMenu = ref(false)

const toggleProfileMenu = () => {
    showProfileMenu.value = !showProfileMenu.value
    showStatusMenu.value = false
    showQueueMenu.value = false
}

// --- QUEUE LOGIN LOGIC ---
const showQueueMenu = ref(false)
const agentStatus = ref('available') // available, on-break, wrap-up
const loggedInQueues = ref([])

const availableQueues = ref([
  { id: 1, name: 'Sales Queue', waiting: 3, agents: 5 },
  { id: 2, name: 'Support Queue', waiting: 0, agents: 8 },
  { id: 3, name: 'Billing Queue', waiting: 1, agents: 3 },
  { id: 4, name: 'General Inquiries', waiting: 0, agents: 2 },
])

const toggleQueueMenu = () => {
    showQueueMenu.value = !showQueueMenu.value
    showStatusMenu.value = false
    showProfileMenu.value = false
}

const isLoggedInto = (queueId) => loggedInQueues.value.includes(queueId)

const toggleQueueLogin = (queue) => {
    if (isLoggedInto(queue.id)) {
        loggedInQueues.value = loggedInQueues.value.filter(id => id !== queue.id)
    } else {
        loggedInQueues.value.push(queue.id)
    }
}

const setAgentStatus = (status) => {
    agentStatus.value = status
}

const loginAllQueues = () => {
    loggedInQueues.value = availableQueues.value.map(q => q.id)
}

const logoutAllQueues = () => {
    loggedInQueues.value = []
}

const closeDropdowns = () => {
    showStatusMenu.value = false
    showProfileMenu.value = false
    showQueueMenu.value = false
}
</script>

<style scoped>
.user-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background: var(--bg-app);
}

.mini-sidebar {
  width: 64px;
  background: white;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 0;
  z-index: 20;
}

.brand-mark {
  width: 32px; height: 32px; background: var(--primary-color);
  color: white; border-radius: 6px; display: flex; align-items: center;
  justify-content: center; font-weight: bold; margin-bottom: 32px;
}

.user-nav { display: flex; flex-direction: column; gap: 8px; flex: 1; }

.nav-icon {
  width: 40px; height: 40px; display: flex; align-items: center; justify-content: center;
  border-radius: 8px; color: var(--text-muted); transition: all var(--transition-fast);
}
.nav-icon:hover { background: var(--bg-app); color: var(--text-primary); }
.nav-icon.router-link-active { background: var(--primary-light); color: var(--primary-color); }

.bottom-actions { margin-top: auto; position: relative; }

.user-avatar-container { position: relative; }
.user-avatar {
  width: 36px; height: 36px; background: linear-gradient(135deg, var(--primary-color), #818cf8); border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: 700; color: white; cursor: pointer;
}

.main-content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }

.user-header {
  height: 60px; background: white; border-bottom: 1px solid var(--border-color);
  display: flex; align-items: center; justify-content: space-between; padding: 0 24px;
}

/* STATUS SELECTOR */
.status-container { position: relative; }

.status-pill {
  display: flex; align-items: center; gap: 8px; padding: 8px 14px;
  border-radius: 99px; background: white; border: 1px solid var(--border-color);
  font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s;
}
.status-pill:hover { background: var(--bg-app); border-color: var(--primary-color); }

.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-pill.available .status-dot { background: #22c55e; }
.status-pill.away .status-dot { background: #f59e0b; }
.status-pill.dnd .status-dot { background: #ef4444; }
.status-pill.custom .status-dot { background: #8b5cf6; }
.status-text { max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chevron { color: var(--text-muted); }

.status-menu { width: 280px; }

.menu-section-header { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); padding: 8px 12px 4px; }

.menu-item.active { background: var(--primary-light); }
.check-icon { width: 14px; height: 14px; color: var(--primary-color); margin-left: auto; }

.custom-status-input { display: flex; gap: 8px; padding: 8px 12px; }
.custom-status-input input { 
  flex: 1; padding: 8px 10px; border: 1px solid var(--border-color); 
  border-radius: 6px; font-size: 12px; 
}
.custom-status-btn { 
  padding: 8px 12px; background: var(--primary-color); color: white; 
  border: none; border-radius: 6px; font-size: 11px; font-weight: 600; cursor: pointer; 
}
.custom-status-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.preset-statuses { display: flex; flex-wrap: wrap; gap: 6px; padding: 4px 12px 12px; }
.preset-btn { 
  padding: 5px 10px; background: var(--bg-app); border: 1px solid var(--border-color); 
  border-radius: 20px; font-size: 11px; cursor: pointer; 
}
.preset-btn:hover { border-color: var(--primary-color); background: var(--primary-light); }

.clear-status { padding: 8px 12px; text-align: center; }

/* QUEUE LOGIN */
.queue-login-container { position: relative; }

.queue-trigger {
  display: flex; align-items: center; gap: 8px;
  background: white; border: 1px solid var(--border-color);
  padding: 8px 14px; border-radius: 8px;
  font-size: 13px; font-weight: 600; color: var(--text-primary);
  cursor: pointer; transition: all 0.2s;
}
.queue-trigger:hover { background: var(--bg-app); border-color: var(--primary-color); }
.queue-trigger.has-active { border-color: #22c55e; background: #f0fdf4; }

.queue-icon { width: 16px; height: 16px; color: var(--text-muted); }
.queue-trigger.has-active .queue-icon { color: #22c55e; }

.queue-status-dot { width: 8px; height: 8px; border-radius: 50%; }
.queue-status-dot.available { background: #22c55e; }
.queue-status-dot.on-break { background: #f59e0b; }
.queue-status-dot.wrap-up { background: #8b5cf6; }

.queue-menu { width: 320px; right: 0; left: auto; }

.agent-status-section, .queue-list-section { padding: 12px; }
.section-header { font-size: 10px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); margin-bottom: 10px; }

.agent-status-buttons { display: flex; gap: 6px; }
.status-btn { 
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px;
  padding: 8px; border: 1px solid var(--border-color); border-radius: 6px;
  background: white; font-size: 11px; font-weight: 500; cursor: pointer;
}
.status-btn:hover { border-color: var(--primary-color); }
.status-btn.active { background: var(--primary-light); border-color: var(--primary-color); }

.dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dot.available { background: #22c55e; }
.dot.away { background: #f59e0b; }
.dot.dnd { background: #ef4444; }
.dot.on-break { background: #f59e0b; }
.dot.wrap-up { background: #8b5cf6; }
.dot.custom { background: #8b5cf6; }

.queue-list { display: flex; flex-direction: column; gap: 6px; }

.queue-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 12px; border: 1px solid var(--border-color); border-radius: 8px;
  cursor: pointer; transition: all 0.15s;
}
.queue-item:hover { background: var(--bg-app); }
.queue-item.active { background: #f0fdf4; border-color: #22c55e; }

.queue-info { display: flex; flex-direction: column; }
.queue-name { font-size: 13px; font-weight: 600; }
.queue-stats { display: flex; gap: 10px; font-size: 11px; color: var(--text-muted); }
.queue-stats .waiting { color: #f59e0b; font-weight: 600; }

.toggle-switch {
  width: 36px; height: 20px; background: #e2e8f0; border-radius: 20px;
  position: relative; transition: background 0.2s;
}
.toggle-switch.on { background: #22c55e; }
.toggle-handle {
  width: 16px; height: 16px; background: white; border-radius: 50%;
  position: absolute; top: 2px; left: 2px; transition: transform 0.2s;
  box-shadow: 0 1px 3px rgba(0,0,0,0.2);
}
.toggle-switch.on .toggle-handle { transform: translateX(16px); }

.queue-actions { display: flex; gap: 8px; padding: 12px; }
.action-btn {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px;
  padding: 8px; background: var(--bg-app); border: 1px solid var(--border-color);
  border-radius: 6px; font-size: 11px; font-weight: 600; cursor: pointer;
}
.action-btn:hover:not(:disabled) { border-color: var(--primary-color); color: var(--primary-color); }
.action-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* SEARCH BAR */
.search-bar {
  display: flex; align-items: center; gap: 8px; background: var(--bg-app);
  padding: 8px 16px; border-radius: 20px; width: 300px;
}
.search-bar input { border: none; background: transparent; outline: none; width: 100%; font-size: var(--text-sm); }

.content-body { flex: 1; padding: 24px; overflow-y: auto; }

/* DROPDOWN COMMON */
.dropdown-menu {
    position: absolute; top: calc(100% + 8px); left: 0;
    width: 200px; background: white; border: 1px solid var(--border-color);
    border-radius: var(--radius-md); box-shadow: var(--shadow-md);
    padding: 6px; z-index: 100;
}
.profile-menu { bottom: 10px; left: 100%; top: auto; margin-left: 10px; width: 220px; }

.menu-header { padding: 8px 10px; }
.user-name { font-weight: 600; font-size: 14px; }
.user-ext { font-size: 11px; color: var(--text-muted); }

.menu-divider { height: 1px; background: var(--border-color); margin: 6px 0; }

.menu-item {
    display: flex; align-items: center; gap: 10px;
    padding: 8px 10px; font-size: 13px; color: var(--text-main);
    border-radius: var(--radius-sm); cursor: pointer; text-decoration: none;
}
.menu-item:hover { background: var(--bg-app); }

.icon { width: 24px; height: 24px; }
.icon-sm { width: 16px; height: 16px; color: var(--text-muted); }
.icon-xs { width: 14px; height: 14px; color: var(--text-main); }
.text-bad { color: var(--status-bad); }
.btn-link { background: none; border: none; color: var(--primary-color); font-size: 12px; cursor: pointer; font-weight: 500; }
</style>
