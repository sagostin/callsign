import { createRouter, createWebHistory } from 'vue-router'
import Overview from './views/admin/Overview.vue'
import TemplateDetail from './views/admin/TemplateDetail.vue'
import UserVoicemail from './views/user/Voicemail.vue'
import UserContacts from './views/user/Contacts.vue'
import UserHistory from './views/user/History.vue'
import Login from './views/auth/Login.vue'
import AdminLogin from './views/auth/AdminLogin.vue'

const routes = [
  // Auth Routes
  { path: '/login', component: Login, name: 'Login' },
  { path: '/admin/login', component: AdminLogin, name: 'AdminLogin' },

  // User Portal Layout (Root)
  {
    path: '',
    component: () => import('./layouts/UserLayout.vue'),
    children: [
      { path: '', redirect: '/dialer' },
      { path: 'dialer', component: () => import('./views/user/Softphone.vue'), name: 'PortalDialer' },
      { path: 'messages', component: () => import('./views/admin/Messaging.vue'), name: 'PortalMessages' },
      { path: 'voicemail', component: UserVoicemail, name: 'PortalVoicemail' },
      { path: 'conferences', component: () => import('./views/user/UserConferences.vue'), name: 'PortalConferences' },
      { path: 'fax', component: () => import('./views/user/UserFax.vue'), name: 'PortalFax' },
      { path: 'contacts', component: UserContacts, name: 'PortalContacts' },
      { path: 'recordings', component: () => import('./views/user/UserRecordings.vue'), name: 'PortalRecordings' },
      { path: 'history', component: UserHistory, name: 'PortalHistory' },
      { path: 'settings', component: () => import('./views/user/UserSettings.vue'), name: 'PortalSettings' },
    ]
  },

  // Tenant Admin Layout (Prefix /admin)
  {
    path: '/admin',
    component: () => import('./components/layout/LayoutShell.vue'),
    children: [
      { path: '', component: Overview, name: 'Overview' },
      { path: 'extensions', component: () => import('./views/admin/Extensions.vue'), name: 'Extensions' },
      { path: 'extensions/:id', component: () => import('./views/admin/ExtensionDetail.vue'), name: 'ExtensionDetail' },
      { path: 'conferences', component: () => import('./views/admin/Conferences.vue'), name: 'Conferences' },
      { path: 'conferences/new', component: () => import('./views/admin/ConferenceForm.vue'), name: 'ConferenceFormNew' },
      { path: 'conferences/:id', component: () => import('./views/admin/ConferenceForm.vue'), name: 'ConferenceFormEdit' },
      { path: 'conferences/console/live', component: () => import('./views/admin/ConferenceConsole.vue'), name: 'ConferenceConsole' },

      { path: 'hospitality', component: () => import('./views/admin/Hospitality.vue'), name: 'Hospitality' },
      { path: 'wake-up-calls', component: () => import('./views/admin/WakeUpCalls.vue'), name: 'WakeUpCalls' },
      { path: 'wake-up-calls/new', component: () => import('./views/admin/WakeUpCallForm.vue'), name: 'WakeUpCallFormNew' },
      { path: 'wake-up-calls/:id', component: () => import('./views/admin/WakeUpCallForm.vue'), name: 'WakeUpCallFormEdit' },

      { path: 'ivr', component: () => import('./views/admin/IVR.vue'), name: 'IVR' },
      { path: 'ivr/menus/new', component: () => import('./views/admin/IVRMenuForm.vue'), name: 'IVRMenuFormNew' },
      { path: 'ivr/menus/:id', component: () => import('./views/admin/IVRMenuForm.vue'), name: 'IVRMenuForm' },

      { path: 'call-flows', component: () => import('./views/admin/CallFlows.vue'), name: 'CallFlows' },

      { path: 'toggles', component: () => import('./views/admin/Toggles.vue'), name: 'Toggles' },

      { path: 'cdr', component: () => import('./views/admin/CDR.vue'), name: 'CDR' },

      { path: 'routing', component: () => import('./views/admin/Routing.vue'), name: 'Routing' },
      { path: 'feature-codes', component: () => import('./views/admin/FeatureCodes.vue'), name: 'TenantFeatureCodes' },
      // Legacy routes kept for form access
      { path: 'numbers/new', component: () => import('./views/admin/NumberForm.vue'), name: 'NumberForm' },
      { path: 'numbers/:id', component: () => import('./views/admin/NumberDetail.vue'), name: 'NumberDetail' },

      { path: 'dial-plans/new', component: () => import('./views/admin/DialPlanForm.vue'), name: 'DialPlanFormNew' },
      { path: 'dial-plans/:id', component: () => import('./views/admin/DialPlanForm.vue'), name: 'DialPlanFormEdit' },

      { path: 'devices', component: () => import('./views/admin/Devices.vue'), name: 'Devices' },
      { path: 'devices/templates', component: () => import('./views/admin/DeviceTemplates.vue'), name: 'DeviceTemplates' },
      { path: 'devices/templates/:id', component: TemplateDetail, name: 'TemplateDetail' },
      { path: 'devices/:id', component: () => import('./views/admin/DeviceForm.vue'), name: 'DeviceForm' },
      { path: 'device-profiles', component: () => import('./views/admin/DeviceProfiles.vue'), name: 'DeviceProfiles' },
      { path: 'provisioning', component: () => import('./views/admin/TenantProvisioning.vue'), name: 'TenantProvisioning' },
      { path: 'queues', component: () => import('./views/admin/Queues.vue'), name: 'Queues' },
      { path: 'queues/new', component: () => import('./views/admin/QueueForm.vue'), name: 'QueueFormNew' },
      { path: 'queues/:id', component: () => import('./views/admin/QueueForm.vue'), name: 'QueueFormEdit' },

      { path: 'bridges', component: () => import('./views/admin/Bridges.vue'), name: 'Bridges' },
      { path: 'bridges/new', component: () => import('./views/admin/BridgeForm.vue'), name: 'BridgeFormNew' },
      { path: 'bridges/:id', component: () => import('./views/admin/BridgeForm.vue'), name: 'BridgeFormEdit' },

      { path: 'trunks', component: () => import('./views/admin/Gateways.vue'), name: 'Trunks' },
      { path: 'trunks/new', component: () => import('./views/admin/GatewayForm.vue'), name: 'TrunkFormNew' },
      { path: 'trunks/:id', component: () => import('./views/admin/GatewayForm.vue'), name: 'TrunkFormEdit' },
      { path: 'gateways', redirect: '/admin/trunks' }, // legacy redirect

      { path: 'call-block', component: () => import('./views/admin/CallBlock.vue'), name: 'CallBlock' },
      { path: 'call-block/new', component: () => import('./views/admin/CallBlockForm.vue'), name: 'CallBlockFormNew' },
      { path: 'call-block/:id', component: () => import('./views/admin/CallBlockForm.vue'), name: 'CallBlockFormEdit' },

      { path: 'call-broadcast', component: () => import('./views/admin/CallBroadcast.vue'), name: 'CallBroadcast' },
      { path: 'call-broadcast/new', component: () => import('./views/admin/CallBroadcastForm.vue'), name: 'CallBroadcastFormNew' },
      { path: 'call-broadcast/:id', component: () => import('./views/admin/CallBroadcastForm.vue'), name: 'CallBroadcastFormEdit' },

      { path: 'speed-dials', component: () => import('./views/admin/SpeedDials.vue'), name: 'SpeedDials' },
      { path: 'extension-profiles', component: () => import('./views/admin/ExtensionProfiles.vue'), name: 'ExtensionProfiles' },

      { path: 'music-on-hold', component: () => import('./views/admin/MusicOnHold.vue'), name: 'MusicOnHold' },
      { path: 'music-on-hold/new', component: () => import('./views/admin/StreamForm.vue'), name: 'StreamFormNew' },
      { path: 'music-on-hold/:id', component: () => import('./views/admin/StreamForm.vue'), name: 'StreamFormEdit' },

      { path: 'operator-panel', component: () => import('./views/admin/OperatorPanel.vue'), name: 'OperatorPanel' },

      { path: 'call-recordings', component: () => import('./views/admin/CallRecordings.vue'), name: 'CallRecordings' },
      { path: 'audio-library', component: () => import('./views/admin/Recordings.vue'), name: 'AudioLibrary' },
      { path: 'audio-library/new', component: () => import('./views/admin/RecordingForm.vue'), name: 'RecordingFormNew' },

      { path: 'fax', component: () => import('./views/admin/FaxServer.vue'), name: 'FaxServer' },
      { path: 'fax/new', component: () => import('./views/admin/FaxBoxForm.vue'), name: 'FaxBoxFormNew' },
      { path: 'fax/:id', component: () => import('./views/admin/FaxBoxForm.vue'), name: 'FaxBoxFormEdit' },

      { path: 'voicemail-manager', component: () => import('./views/admin/VoicemailBoxes.vue'), name: 'VoicemailBoxes' },
      { path: 'voicemail-manager/new', component: () => import('./views/admin/VoicemailBoxForm.vue'), name: 'VoicemailBoxFormNew' },
      { path: 'voicemail-manager/:id', component: () => import('./views/admin/VoicemailBoxForm.vue'), name: 'VoicemailBoxFormEdit' },

      { path: 'messaging', component: () => import('./views/admin/Messaging.vue'), name: 'Messaging' },
      { path: 'reports', component: () => import('./views/admin/Reports.vue'), name: 'Reports' },
      { path: 'audit-log', component: () => import('./views/admin/AuditLog.vue'), name: 'AuditLog' },
      { path: 'settings', component: () => import('./views/admin/TenantSettings.vue'), name: 'TenantSettings' },

      { path: 'time-conditions', component: () => import('./views/admin/TimeConditions.vue'), name: 'TimeConditions' },
      { path: 'time-conditions/new', component: () => import('./views/admin/TimeConditionForm.vue'), name: 'TimeConditionFormNew' },
      { path: 'time-conditions/:id', component: () => import('./views/admin/TimeConditionForm.vue'), name: 'TimeConditionFormEdit' },
      // Legacy redirect
      { path: 'schedules', redirect: '/admin/time-conditions' },
    ]
  },

  // System Admin Layout (Prefix /system)
  {
    path: '/system',
    component: () => import('./components/layout/LayoutShell.vue'),
    children: [
      { path: '', component: () => import('./views/system/Admin.vue'), name: 'SystemDashboard' },
      { path: 'tenants', component: () => import('./views/system/Tenants.vue'), name: 'Tenants' },
      { path: 'tenants/new', component: () => import('./views/system/TenantForm.vue'), name: 'TenantFormNew' },
      { path: 'tenants/:id', component: () => import('./views/system/TenantForm.vue'), name: 'TenantFormEdit' },

      { path: 'profiles', component: () => import('./views/system/TenantProfiles.vue'), name: 'TenantProfiles' },

      { path: 'provisioning-templates', component: () => import('./views/system/ProvisioningTemplates.vue'), name: 'ProvisioningTemplates' },
      { path: 'firmware', component: () => import('./views/system/FirmwareUpdates.vue'), name: 'FirmwareUpdates' },
      { path: 'media', component: () => import('./views/system/SystemMedia.vue'), name: 'SystemMedia' },
      { path: 'sounds', redirect: '/system/media' }, // legacy redirect
      { path: 'moh', redirect: '/system/media' },    // legacy redirect
      { path: 'phrases', redirect: '/system/media' },// legacy redirect

      { path: 'infrastructure', component: () => import('./views/system/Infrastructure.vue'), name: 'Infrastructure' },
      { path: 'trunks', component: () => import('./views/system/SystemGateways.vue'), name: 'SystemTrunks' },
      { path: 'gateways', redirect: '/system/trunks' }, // legacy redirect
      { path: 'sip-profiles', component: () => import('./views/system/SipProfiles.vue'), name: 'SipProfiles' },
      { path: 'acls', component: () => import('./views/system/ACLProfiles.vue'), name: 'ACLProfiles' },
      { path: 'routing', component: () => import('./views/system/SystemRoutes.vue'), name: 'SystemRoutes' },
      { path: 'dial-plans', redirect: '/system/routing' }, // legacy redirect
      { path: 'phrases', component: () => import('./views/system/Phrases.vue'), name: 'SystemPhrases' },
      { path: 'phrases/new', component: () => import('./views/admin/PhraseForm.vue'), name: 'SystemPhraseFormNew' },
      { path: 'phrases/:id', component: () => import('./views/admin/PhraseForm.vue'), name: 'SystemPhraseFormEdit' },
      { path: 'logs', component: () => import('./views/system/SystemLogs.vue'), name: 'SystemLogs' },
      { path: 'messaging', component: () => import('./views/system/MessagingProviders.vue'), name: 'MessagingProviders' },
      { path: 'audit-log', component: () => import('./views/admin/AuditLog.vue'), name: 'SystemAuditLog' },
      { path: 'settings', component: () => import('./views/system/SystemSettings.vue'), name: 'SystemSettings' },
      { path: 'security', component: () => import('./views/system/SystemSecurity.vue'), name: 'SystemSecurity' },
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Auth guard - check authentication for protected routes
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null
  const tenantId = localStorage.getItem('tenantId')

  // Public routes that don't need auth
  const publicRoutes = ['Login', 'AdminLogin']
  if (publicRoutes.includes(to.name)) {
    // If already logged in, redirect based on role
    if (token && user) {
      if (user.role === 'system_admin') {
        return next('/system')
      }
      if (user.role === 'tenant_admin') {
        return next('/admin')
      }
      return next('/dialer')
    }
    return next()
  }

  // Check if authenticated
  if (!token) {
    // Redirect to appropriate login
    if (to.path.startsWith('/system') || to.path.startsWith('/admin')) {
      return next('/admin/login')
    }
    return next('/login')
  }

  // System admin restrictions
  if (user?.role === 'system_admin') {
    // Block system admin from user portal routes
    const userPortalRoutes = ['', '/', '/dialer', '/messages', '/voicemail', '/conferences', '/fax', '/contacts', '/recordings', '/history', '/settings']
    if (userPortalRoutes.includes(to.path) || to.path.match(/^\/(?!admin|system)/)) {
      return next('/system')
    }

    // System admin accessing /admin routes needs a tenant selected
    if (to.path.startsWith('/admin') && !tenantId) {
      console.warn('System admin accessing admin panel without tenant selected, redirecting to /system')
      return next('/system')
    }
  }

  // Check role permissions for admin/system routes
  if (to.path.startsWith('/system') && user?.role !== 'system_admin') {
    return next('/admin')
  }

  if (to.path.startsWith('/admin') && !['system_admin', 'tenant_admin'].includes(user?.role)) {
    return next('/dialer')
  }

  next()
})

export default router

