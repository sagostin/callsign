<template>
  <div class="settings-page">
    <div class="view-header">
      <div class="header-content">
        <h2>My Settings</h2>
        <p class="text-muted text-sm">Manage your profile, preferences, and account settings.</p>
      </div>
    </div>

    <div class="settings-layout">
      <!-- Sidebar Navigation -->
      <nav class="settings-nav">
        <button 
          v-for="section in sections" 
          :key="section.id" 
          class="nav-btn"
          :class="{ active: activeSection === section.id }"
          @click="activeSection = section.id"
        >
          <component :is="section.icon" class="nav-icon" />
          <span>{{ section.label }}</span>
        </button>
      </nav>

      <!-- Settings Content -->
      <div class="settings-content">
        
        <!-- PROFILE SECTION -->
        <div v-if="activeSection === 'profile'" class="section">
          <h3>Profile Information</h3>
          
          <div class="profile-header">
            <div class="avatar-upload">
              <div class="avatar-preview">
                <img v-if="profile.avatar" :src="profile.avatar" alt="Avatar">
                <span v-else class="avatar-initials">{{ userInitials }}</span>
              </div>
              <button class="upload-btn">
                <CameraIcon class="icon-sm" />
                Change Photo
              </button>
            </div>
            <div class="profile-status">
              <span class="status-label">Status:</span>
              <select v-model="profile.status" class="status-select">
                <option value="available">🟢 Available</option>
                <option value="away">🟡 Away</option>
                <option value="dnd">🔴 Do Not Disturb</option>
                <option value="invisible">⚫ Invisible</option>
              </select>
            </div>
          </div>

          <div class="form-grid">
            <div class="form-group">
              <label>First Name</label>
              <input v-model="profile.firstName" class="input-field">
            </div>
            <div class="form-group">
              <label>Last Name</label>
              <input v-model="profile.lastName" class="input-field">
            </div>
            <div class="form-group">
              <label>Email Address</label>
              <input v-model="profile.email" type="email" class="input-field">
            </div>
            <div class="form-group">
              <label>Mobile Phone</label>
              <input v-model="profile.mobile" class="input-field" placeholder="(555) 555-5555">
            </div>
            <div class="form-group full-width">
              <label>Extension</label>
              <input :value="profile.extension" class="input-field" disabled>
              <span class="input-hint">Contact your administrator to change your extension.</span>
            </div>
          </div>

          <div class="section-actions">
            <button class="btn-primary" @click="saveProfile">Save Changes</button>
          </div>
        </div>

        <!-- CALL HANDLING SECTION -->
        <div v-else-if="activeSection === 'call-handling'" class="section">
          <h3>Call Handling</h3>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Ring Strategy</span>
                <span class="setting-desc">Choose how incoming calls ring your devices</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="ring-strategies">
                <label class="strategy-option" :class="{ active: callHandling.strategy === 'simultaneous' }">
                  <input type="radio" v-model="callHandling.strategy" value="simultaneous">
                  <div class="strategy-icon">
                    <PhoneCallIcon class="icon" />
                  </div>
                  <div class="strategy-info">
                    <strong>Ring All Simultaneously</strong>
                    <span>All enabled devices ring at the same time</span>
                  </div>
                </label>
                
                <label class="strategy-option" :class="{ active: callHandling.strategy === 'sequential' }">
                  <input type="radio" v-model="callHandling.strategy" value="sequential">
                  <div class="strategy-icon">
                    <ListOrderedIcon class="icon" />
                  </div>
                  <div class="strategy-info">
                    <strong>Ring in Order</strong>
                    <span>Ring devices one at a time in the order below</span>
                  </div>
                </label>
              </div>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">My Devices</span>
                <span class="setting-desc">Manage devices and ring order</span>
              </div>
            </div>
            <div class="setting-body">
               <div class="device-ring-list">
                <div 
                  class="ring-device-item" 
                  v-for="(device, index) in callHandling.devices" 
                  :key="device.id"
                  draggable="true"
                  @dragstart="dragStart(index)"
                  @dragover.prevent
                  @drop="drop(index)"
                  :class="{ disabled: !device.enabled, dragging: dragIndex === index }"
                >
                  <div class="drag-handle">
                    <GripVerticalIcon class="icon-sm" />
                  </div>
                  <div class="device-order">{{ index + 1 }}</div>
                  <div class="device-icon-box" :class="device.type">
                    <MonitorIcon v-if="device.type === 'softphone'" class="icon-sm" />
                    <PhoneIcon v-else-if="device.type === 'desk'" class="icon-sm" />
                    <SmartphoneIcon v-else-if="device.type === 'mobile'" class="icon-sm" />
                    <HeadphonesIcon v-else class="icon-sm" />
                  </div>
                  <div class="device-info">
                    <span class="device-name">{{ device.name }}</span>
                    <span class="device-details">{{ device.details }}</span>
                  </div>
                  <div class="ring-duration" v-if="callHandling.strategy === 'sequential' && device.enabled">
                    <select v-model="device.ringTime" class="input-field small">
                      <option value="10">10s</option>
                      <option value="15">15s</option>
                      <option value="20">20s</option>
                      <option value="30">30s</option>
                    </select>
                  </div>
                  <label class="toggle">
                    <input type="checkbox" v-model="device.enabled">
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">No Answer Action</span>
                <span class="setting-desc">What happens if you don't answer</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="form-group">
                <select v-model="callHandling.noAnswerAction" class="input-field">
                  <option value="voicemail">Send to Voicemail</option>
                  <option value="forward">Forward to Number</option>
                  <option value="hangup">Hang Up</option>
                </select>
              </div>
              
              <div class="form-group" v-if="callHandling.noAnswerAction === 'forward'" style="margin-top: 12px;">
                <label>Forward To</label>
                <input v-model="callHandling.forwardNumber" class="input-field" placeholder="(555) 555-1234">
              </div>
            </div>
          </div>

          <div class="setting-card">
             <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Other Preferences</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="pref-row">
                 <div class="pref-info">
                    <div class="pref-label">Do Not Disturb</div>
                    <div class="pref-desc">Send all calls directly to voicemail</div>
                 </div>
                 <label class="toggle">
                    <input type="checkbox" v-model="phoneSettings.dndEnabled">
                    <span class="toggle-slider"></span>
                  </label>
              </div>
              <div class="divider"></div>
               <div class="pref-row">
                 <div class="pref-info">
                    <div class="pref-label">Call Waiting</div>
                    <div class="pref-desc">Receive incoming calls while busy</div>
                 </div>
                 <label class="toggle">
                    <input type="checkbox" v-model="phoneSettings.callWaiting">
                    <span class="toggle-slider"></span>
                  </label>
              </div>
              <div class="divider"></div>
              <div class="form-group" style="margin-top: 12px;">
                <label>Outbound Caller ID Name</label>
                <input v-model="phoneSettings.callerIdName" class="input-field">
              </div>
            </div>
          </div>

          <div class="section-actions">
            <button class="btn-primary" @click="saveCallHandling">Save Changes</button>
          </div>
        </div>

        <!-- VOICEMAIL SECTION -->
        <div v-else-if="activeSection === 'voicemail'" class="section">
          <h3>Voicemail Settings</h3>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Voicemail PIN</span>
                <span class="setting-desc">Used to access voicemail by phone</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="form-row">
                <input v-model="voicemailSettings.pin" type="password" class="input-field" style="max-width: 150px;">
                <button class="btn-secondary" @click="showPin = !showPin">{{ showPin ? 'Hide' : 'Show' }}</button>
              </div>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Voicemail to Email</span>
                <span class="setting-desc">Receive voicemails as email attachments</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="voicemailSettings.emailEnabled">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Voicemail Transcription</span>
                <span class="setting-desc">Get text transcripts of voicemails</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="voicemailSettings.transcription">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Greeting</span>
                <span class="setting-desc">Your voicemail greeting message</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="greeting-options">
                <label class="radio-option">
                  <input type="radio" v-model="voicemailSettings.greetingType" value="default">
                  <span>Default System Greeting</span>
                </label>
                <label class="radio-option">
                  <input type="radio" v-model="voicemailSettings.greetingType" value="name">
                  <span>Name Only ("You've reached John Smith...")</span>
                </label>
                <label class="radio-option">
                  <input type="radio" v-model="voicemailSettings.greetingType" value="custom">
                  <span>Custom Recording</span>
                </label>
              </div>
              <div class="recording-controls" v-if="voicemailSettings.greetingType === 'custom'">
                <button class="btn-secondary"><MicIcon class="btn-icon" /> Record New</button>
                <button class="btn-secondary"><UploadIcon class="btn-icon" /> Upload</button>
                <button class="btn-secondary" v-if="voicemailSettings.hasCustomGreeting"><PlayIcon class="btn-icon" /> Play Current</button>
              </div>
            </div>
          </div>
        </div>

        <!-- NOTIFICATIONS SECTION -->
        <div v-else-if="activeSection === 'notifications'" class="section">
          <h3>Notification Preferences</h3>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Missed Call Alerts</span>
                <span class="setting-desc">Get notified of missed calls</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="notifications.missedCalls">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">New Voicemail Alerts</span>
                <span class="setting-desc">Get notified of new voicemails</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="notifications.voicemail">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">New Fax Alerts</span>
                <span class="setting-desc">Get notified of received faxes</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="notifications.fax">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Desktop Notifications</span>
                <span class="setting-desc">Show browser notifications</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="notifications.desktop">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Sound Alerts</span>
                <span class="setting-desc">Play sounds for notifications</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="notifications.sounds">
                <span class="toggle-slider"></span>
              </label>
            </div>
            <div class="setting-body" v-if="notifications.sounds">
              <div class="form-group">
                <label>Ringtone</label>
                <select v-model="notifications.ringtone" class="input-field">
                  <option value="default">Default</option>
                  <option value="classic">Classic Ring</option>
                  <option value="digital">Digital</option>
                  <option value="soft">Soft Chime</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- SECURITY SECTION -->
        <div v-else-if="activeSection === 'security'" class="section">
          <h3>Security Settings</h3>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Change Password</span>
                <span class="setting-desc">Update your account password</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="form-group">
                <label>Current Password</label>
                <input type="password" class="input-field">
              </div>
              <div class="form-group">
                <label>New Password</label>
                <input type="password" class="input-field">
              </div>
              <div class="form-group">
                <label>Confirm New Password</label>
                <input type="password" class="input-field">
              </div>
              <button class="btn-primary">Update Password</button>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Two-Factor Authentication</span>
                <span class="setting-desc">Add an extra layer of security</span>
              </div>
              <label class="toggle">
                <input type="checkbox" v-model="security.twoFactorEnabled">
                <span class="toggle-slider"></span>
              </label>
            </div>
          </div>

          <div class="setting-card">
            <div class="setting-header">
              <div class="setting-info">
                <span class="setting-title">Active Sessions</span>
                <span class="setting-desc">Manage your logged-in devices</span>
              </div>
            </div>
            <div class="setting-body">
              <div class="session-list">
                <div class="session-item">
                  <MonitorIcon class="session-icon" />
                  <div class="session-info">
                    <span class="session-name">Chrome on macOS</span>
                    <span class="session-meta">Current session • San Francisco, CA</span>
                  </div>
                  <span class="session-badge current">Current</span>
                </div>
                <div class="session-item">
                  <SmartphoneIcon class="session-icon" />
                  <div class="session-info">
                    <span class="session-name">Mobile App on iPhone</span>
                    <span class="session-meta">Last active 2 hours ago • San Francisco, CA</span>
                  </div>
                  <button class="btn-link text-danger">Sign Out</button>
                </div>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { 
  User as UserIcon, Phone as PhoneIcon, Voicemail as VoicemailIcon,
  Bell as BellIcon, Shield as ShieldIcon, Camera as CameraIcon,
  Mic as MicIcon, Upload as UploadIcon, Play as PlayIcon,
  Monitor as MonitorIcon, Smartphone as SmartphoneIcon,
  ListOrdered as ListOrderedIcon, GripVertical as GripVerticalIcon,
  PhoneCall as PhoneCallIcon, Headphones as HeadphonesIcon,
  Check as CheckIcon, X as XIcon
} from 'lucide-vue-next'

const activeSection = ref('profile')
const showPin = ref(false)

const sections = [
  { id: 'profile', label: 'Profile', icon: UserIcon },
  { id: 'call-handling', label: 'Call Handling', icon: PhoneIcon },
  { id: 'voicemail', label: 'Voicemail', icon: VoicemailIcon },
  { id: 'notifications', label: 'Notifications', icon: BellIcon },
  { id: 'security', label: 'Security', icon: ShieldIcon },
]

const profile = ref({
  firstName: 'John',
  lastName: 'Smith',
  email: 'john.smith@company.com',
  mobile: '(415) 555-1234',
  extension: '101',
  avatar: null,
  status: 'available'
})

const userInitials = computed(() => {
  return profile.value.firstName.charAt(0) + profile.value.lastName.charAt(0)
})

const phoneSettings = ref({
  dndEnabled: false,
  callWaiting: true,
  callerIdName: 'John Smith'
})

// Call Handling
const dragIndex = ref(null)
const callHandling = ref({
  strategy: 'simultaneous',
  noAnswerAction: 'voicemail',
  forwardNumber: '',
  devices: [
    { id: 1, type: 'softphone', name: 'Web Softphone', details: 'Browser / Desktop App', enabled: true, ringTime: '20' },
    { id: 2, type: 'desk', name: 'Desk Phone', details: 'Yealink T54W - Office', enabled: true, ringTime: '20' },
    { id: 3, type: 'mobile', name: 'Mobile App', details: 'iPhone 13', enabled: true, ringTime: '20' },
  ]
})

const dragStart = (index) => { dragIndex.value = index }
const drop = (index) => {
  const items = callHandling.value.devices
  const item = items.splice(dragIndex.value, 1)[0]
  items.splice(index, 0, item)
  dragIndex.value = null
}

const saveCallHandling = () => alert('Call handling settings saved!')

const voicemailSettings = ref({
  pin: '1234',
  emailEnabled: true,
  transcription: true,
  greetingType: 'name',
  hasCustomGreeting: false
})

const notifications = ref({
  missedCalls: true,
  voicemail: true,
  fax: true,
  desktop: true,
  sounds: true,
  ringtone: 'default'
})

const security = ref({
  twoFactorEnabled: false
})

const saveProfile = () => alert('Profile saved!')
</script>

<style scoped>
.settings-page { padding: 0; }

.view-header { margin-bottom: var(--spacing-lg); }
.header-content h2 { margin: 0 0 4px; }

.settings-layout { display: flex; gap: 24px; }

/* Settings Nav */
.settings-nav {
  width: 200px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  text-align: left;
  transition: all 0.15s;
}
.nav-btn:hover { background: var(--bg-app); color: var(--text-primary); }
.nav-btn.active { background: var(--primary-light); color: var(--primary-color); }
.nav-icon { width: 18px; height: 18px; }

/* Settings Content */
.settings-content {
  flex: 1;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
}

.section h3 { margin: 0 0 24px; font-size: 18px; font-weight: 700; }

/* Profile Header */
.profile-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; padding-bottom: 24px; border-bottom: 1px solid var(--border-color); }

.avatar-upload { display: flex; align-items: center; gap: 16px; }
.avatar-preview { width: 80px; height: 80px; border-radius: 50%; background: linear-gradient(135deg, var(--primary-color), #818cf8); display: flex; align-items: center; justify-content: center; overflow: hidden; }
.avatar-preview img { width: 100%; height: 100%; object-fit: cover; }
.avatar-initials { color: white; font-size: 28px; font-weight: 700; }
.upload-btn { display: flex; align-items: center; gap: 6px; padding: 8px 14px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); background: white; cursor: pointer; font-size: 12px; font-weight: 500; }

.profile-status { display: flex; align-items: center; gap: 8px; }
.status-label { font-size: 12px; color: var(--text-muted); }
.status-select { padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 13px; }

/* Form Grid */
.form-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full-width { grid-column: 1 / -1; }
.form-group label { font-size: 11px; font-weight: 700; text-transform: uppercase; color: var(--text-muted); }
.input-field { padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: 14px; }
.input-field:disabled { background: var(--bg-app); color: var(--text-muted); }
.input-hint { font-size: 11px; color: var(--text-muted); margin-top: 4px; }

.form-row { display: flex; gap: 12px; align-items: center; }

.section-actions { margin-top: 24px; padding-top: 24px; border-top: 1px solid var(--border-color); }

/* Setting Cards */
.setting-card { border: 1px solid var(--border-color); border-radius: var(--radius-sm); margin-bottom: 16px; overflow: hidden; }
.setting-header { display: flex; justify-content: space-between; align-items: center; padding: 16px; }
.setting-info { display: flex; flex-direction: column; }
.setting-title { font-weight: 600; font-size: 14px; }
.setting-desc { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
.setting-body { padding: 16px; padding-top: 0; }

/* Toggle */
.toggle { position: relative; display: inline-block; width: 44px; height: 24px; }
.toggle input { opacity: 0; width: 0; height: 0; }
.toggle-slider { position: absolute; cursor: pointer; inset: 0; background: #e2e8f0; border-radius: 24px; transition: 0.3s; }
.toggle-slider:before { content: ''; position: absolute; width: 18px; height: 18px; left: 3px; bottom: 3px; background: white; border-radius: 50%; transition: 0.3s; }
.toggle input:checked + .toggle-slider { background: var(--primary-color); }
.toggle input:checked + .toggle-slider:before { transform: translateX(20px); }

/* Radio Options */
.greeting-options { display: flex; flex-direction: column; gap: 10px; margin-bottom: 16px; }
.radio-option { display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer; }
.radio-option input { accent-color: var(--primary-color); }

.recording-controls { display: flex; gap: 12px; }

/* Session List */
.session-list { display: flex; flex-direction: column; gap: 12px; }
.session-item { display: flex; align-items: center; gap: 12px; padding: 12px; background: var(--bg-app); border-radius: var(--radius-sm); }
.session-icon { width: 24px; height: 24px; color: var(--text-muted); }
.session-info { flex: 1; display: flex; flex-direction: column; }
.session-name { font-weight: 600; font-size: 13px; }
.session-meta { font-size: 11px; color: var(--text-muted); }
.session-badge { font-size: 10px; font-weight: 600; padding: 3px 8px; border-radius: 4px; }
.session-badge.current { background: #dcfce7; color: #16a34a; }

/* Buttons */
.btn-primary { display: flex; align-items: center; gap: 6px; background-color: var(--primary-color); color: white; border: none; padding: 10px 20px; border-radius: var(--radius-sm); font-weight: 600; cursor: pointer; }
.btn-secondary { display: flex; align-items: center; gap: 6px; background: white; border: 1px solid var(--border-color); padding: 10px 16px; border-radius: var(--radius-sm); font-weight: 500; cursor: pointer; font-size: 13px; }
.btn-link { background: none; border: none; color: var(--primary-color); font-weight: 600; cursor: pointer; font-size: 12px; }
.btn-link.text-danger { color: #dc2626; }
.btn-icon { width: 16px; height: 16px; }

/* Call Handling */
.ring-strategies { display: flex; gap: 16px; margin-bottom: 8px; }
.strategy-option {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.15s;
}
.strategy-option:hover { border-color: var(--primary-color); }
.strategy-option.active { border-color: var(--primary-color); background: var(--primary-light); }
.strategy-option input { display: none; }
.strategy-icon { width: 40px; height: 40px; border-radius: 10px; background: var(--bg-app); display: flex; align-items: center; justify-content: center; }
.strategy-option.active .strategy-icon { background: var(--primary-color); color: white; }
.strategy-info { display: flex; flex-direction: column; gap: 2px; }
.strategy-info strong { font-size: 14px; }
.strategy-info span { font-size: 12px; color: var(--text-muted); }

.device-ring-list { display: flex; flex-direction: column; gap: 8px; }
.ring-device-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  transition: all 0.15s;
}
.ring-device-item:hover { background: var(--bg-app); }
.ring-device-item.disabled { opacity: 0.6; }
.ring-device-item.dragging { opacity: 0.5; background: var(--primary-light); }

.drag-handle { cursor: grab; color: var(--text-muted); margin-right: 4px; }
.drag-handle:active { cursor: grabbing; }
.device-order { width: 24px; height: 24px; background: var(--bg-app); border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; color: var(--text-muted); }
.device-icon-box { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; }
.device-icon-box.softphone { background: #dbeafe; color: #2563eb; }
.device-icon-box.desk { background: #dcfce7; color: #16a34a; }
.device-icon-box.mobile { background: #f3e8ff; color: #7c3aed; }
.device-info { flex: 1; display: flex; flex-direction: column; }
.device-name { font-weight: 600; font-size: 14px; }
.device-details { font-size: 12px; color: var(--text-muted); }
.ring-duration select { width: 80px; }

/* Preferences Row */
.pref-row { display: flex; justify-content: space-between; align-items: center; padding: 4px 0; }
.pref-info { display: flex; flex-direction: column; }
.pref-label { font-size: 14px; font-weight: 600; }
.pref-desc { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
.divider { height: 1px; background: var(--border-color); margin: 16px 0; }

.icon { width: 20px; height: 20px; }
.icon-sm { width: 16px; height: 16px; }

/* Responsive */
@media (max-width: 768px) {
  .settings-layout { flex-direction: column; }
  .settings-nav { width: 100%; flex-direction: row; overflow-x: auto; gap: 4px; }
  .nav-btn { white-space: nowrap; }
  .form-grid { grid-template-columns: 1fr; }
  .profile-header { flex-direction: column; gap: 16px; }
}
</style>
