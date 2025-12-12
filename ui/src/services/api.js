import axios from 'axios'

// Create axios instance with base configuration
const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api',
    timeout: 30000,
    headers: {
        'Content-Type': 'application/json',
    },
})

// Request interceptor - adds auth token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }

        // Add tenant header if available
        const tenantId = localStorage.getItem('tenantId')
        if (tenantId) {
            config.headers['X-Tenant-ID'] = tenantId
        }

        return config
    },
    (error) => Promise.reject(error)
)

// Response interceptor - handles errors and token refresh
api.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config

        // Handle 401 - try to refresh token
        if (error.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true

            try {
                const refreshToken = localStorage.getItem('refreshToken')
                if (refreshToken) {
                    const response = await axios.post('/api/auth/refresh', {}, {
                        headers: { Authorization: `Bearer ${refreshToken}` }
                    })

                    const { token } = response.data
                    localStorage.setItem('token', token)
                    originalRequest.headers.Authorization = `Bearer ${token}`
                    return api(originalRequest)
                }
            } catch (refreshError) {
                // Refresh failed, logout
                localStorage.removeItem('token')
                localStorage.removeItem('refreshToken')
                localStorage.removeItem('user')
                window.location.href = '/login'
                return Promise.reject(refreshError)
            }
        }

        // Transform error for UI
        const errorMessage = error.response?.data?.error
            || error.response?.data?.message
            || error.message
            || 'An error occurred'

        return Promise.reject({
            status: error.response?.status,
            message: errorMessage,
            data: error.response?.data,
        })
    }
)

export default api

// =====================
// Auth API
// =====================
export const authAPI = {
    login: (username, password) =>
        api.post('/auth/login', { username, password }),

    adminLogin: (username, password) =>
        api.post('/auth/admin/login', { username, password }),

    logout: () => api.post('/auth/logout'),

    getProfile: () => api.get('/auth/me'),

    changePassword: (currentPassword, newPassword) =>
        api.put('/auth/password', { current_password: currentPassword, new_password: newPassword }),

    refreshToken: () => api.post('/auth/refresh'),
}

// =====================
// Extensions API
// =====================
export const extensionsAPI = {
    list: (params) => api.get('/extensions', { params }),
    get: (id) => api.get(`/extensions/${id}`),
    create: (data) => api.post('/extensions', data),
    update: (id, data) => api.put(`/extensions/${id}`, data),
    delete: (id) => api.delete(`/extensions/${id}`),
    getStatus: (id) => api.get(`/extensions/${id}/status`),
}

// =====================
// Extension Profiles API
// =====================
export const extensionProfilesAPI = {
    list: () => api.get('/extension-profiles'),
    get: (id) => api.get(`/extension-profiles/${id}`),
    create: (data) => api.post('/extension-profiles', data),
    update: (id, data) => api.put(`/extension-profiles/${id}`, data),
    delete: (id) => api.delete(`/extension-profiles/${id}`),
}

// =====================
// Devices API
// =====================
export const devicesAPI = {
    list: (params) => api.get('/devices', { params }),
    get: (id) => api.get(`/devices/${id}`),
    create: (data) => api.post('/devices', data),
    update: (id, data) => api.put(`/devices/${id}`, data),
    delete: (id) => api.delete(`/devices/${id}`),

    // Device actions
    reprovision: (id) => api.post(`/devices/${id}/reprovision`),
    assignUser: (id, userId) => api.post(`/devices/${id}/assign-user`, { user_id: userId }),
    assignProfile: (id, profileId) => api.post(`/devices/${id}/assign-profile`, { profile_id: profileId }),

    // Lines
    getLines: (id) => api.get(`/devices/${id}/lines`),
    updateLines: (id, lines) => api.put(`/devices/${id}/lines`, lines),

    // Call control
    hangup: (id) => api.post(`/devices/${id}/hangup`),
    transfer: (id, destination, type = 'blind') =>
        api.post(`/devices/${id}/transfer`, { destination, type }),
    hold: (id, hold = true) => api.post(`/devices/${id}/hold`, { hold }),
    dial: (id, number) => api.post(`/devices/${id}/dial`, { number }),
    callStatus: (id) => api.get(`/devices/${id}/call-status`),
}

// =====================
// Device Profiles API (tenant-level)
// =====================
export const deviceProfilesAPI = {
    list: (params) => api.get('/device-profiles', { params }),
    get: (id) => api.get(`/device-profiles/${id}`),
    create: (data) => api.post('/device-profiles', data),
    update: (id, data) => api.put(`/device-profiles/${id}`, data),
    delete: (id) => api.delete(`/device-profiles/${id}`),
}

// =====================
// Device Templates API (tenant-level, includes system templates)
// =====================
export const deviceTemplatesAPI = {
    list: (params) => api.get('/device-templates', { params }),
    get: (id) => api.get(`/device-templates/${id}`),
    create: (data) => api.post('/device-templates', data),
    update: (id, data) => api.put(`/device-templates/${id}`, data),
    delete: (id) => api.delete(`/device-templates/${id}`),
}

// =====================
// Queues API
// =====================
export const queuesAPI = {
    list: (params) => api.get('/queues', { params }),
    get: (id) => api.get(`/queues/${id}`),
    create: (data) => api.post('/queues', data),
    update: (id, data) => api.put(`/queues/${id}`, data),
    delete: (id) => api.delete(`/queues/${id}`),
}

// =====================
// Conferences API
// =====================
export const conferencesAPI = {
    list: (params) => api.get('/conferences', { params }),
    get: (id) => api.get(`/conferences/${id}`),
    create: (data) => api.post('/conferences', data),
    update: (id, data) => api.put(`/conferences/${id}`, data),
    delete: (id) => api.delete(`/conferences/${id}`),
}

// =====================
// IVR API
// =====================
export const ivrAPI = {
    listMenus: (params) => api.get('/ivr/menus', { params }),
    getMenu: (id) => api.get(`/ivr/menus/${id}`),
    createMenu: (data) => api.post('/ivr/menus', data),
    updateMenu: (id, data) => api.put(`/ivr/menus/${id}`, data),
    deleteMenu: (id) => api.delete(`/ivr/menus/${id}`),
}

// =====================
// Feature Codes API
// =====================
export const featureCodesAPI = {
    list: (params) => api.get('/feature-codes', { params }),
    listSystem: () => api.get('/feature-codes/system'),
    get: (id) => api.get(`/feature-codes/${id}`),
    create: (data) => api.post('/feature-codes', data),
    update: (id, data) => api.put(`/feature-codes/${id}`, data),
    delete: (id) => api.delete(`/feature-codes/${id}`),
}

// =====================
// Time Conditions API
// =====================
export const timeConditionsAPI = {
    list: (params) => api.get('/time-conditions', { params }),
    get: (id) => api.get(`/time-conditions/${id}`),
    create: (data) => api.post('/time-conditions', data),
    update: (id, data) => api.put(`/time-conditions/${id}`, data),
    delete: (id) => api.delete(`/time-conditions/${id}`),
}

// =====================
// Holidays API
// =====================
export const holidaysAPI = {
    list: (params) => api.get('/holidays', { params }),
    get: (id) => api.get(`/holidays/${id}`),
    create: (data) => api.post('/holidays', data),
    update: (id, data) => api.put(`/holidays/${id}`, data),
    delete: (id) => api.delete(`/holidays/${id}`),
    sync: (id) => api.post(`/holidays/${id}/sync`),
}

// =====================
// Call Flows API
// =====================
export const callFlowsAPI = {
    list: (params) => api.get('/call-flows', { params }),
    get: (id) => api.get(`/call-flows/${id}`),
    create: (data) => api.post('/call-flows', data),
    update: (id, data) => api.put(`/call-flows/${id}`, data),
    delete: (id) => api.delete(`/call-flows/${id}`),
    toggle: (id) => api.post(`/call-flows/${id}/toggle`),
}

// =====================
// Voicemail API
// =====================
export const voicemailAPI = {
    listBoxes: (params) => api.get('/voicemail/boxes', { params }),
    getBox: (ext) => api.get(`/voicemail/boxes/${ext}`),
    createBox: (data) => api.post('/voicemail/boxes', data),
    updateBox: (ext, data) => api.put(`/voicemail/boxes/${ext}`, data),
    deleteBox: (ext) => api.delete(`/voicemail/boxes/${ext}`),
}

// =====================
// CDR API
// =====================
export const cdrAPI = {
    list: (params) => api.get('/cdr', { params }),
    get: (id) => api.get(`/cdr/${id}`),
    export: (params) => api.get('/cdr/export', { params, responseType: 'blob' }),
}

// =====================
// Messaging API
// =====================
export const messagingAPI = {
    listConversations: (params) => api.get('/messaging/conversations', { params }),
    getConversation: (id) => api.get(`/messaging/conversations/${id}`),
    sendMessage: (data) => api.post('/messaging/send', data),
}

// =====================
// Contacts API
// =====================
export const contactsAPI = {
    list: (params) => api.get('/contacts', { params }),
    get: (id) => api.get(`/contacts/${id}`),
    create: (data) => api.post('/contacts', data),
    update: (id, data) => api.put(`/contacts/${id}`, data),
    delete: (id) => api.delete(`/contacts/${id}`),
    lookup: (phone) => api.get('/contacts/lookup', { params: { phone } }),
}

// =====================
// Paging API
// =====================
export const pagingAPI = {
    list: (params) => api.get('/page-groups', { params }),
    get: (id) => api.get(`/page-groups/${id}`),
    create: (data) => api.post('/page-groups', data),
    update: (id, data) => api.put(`/page-groups/${id}`, data),
    delete: (id) => api.delete(`/page-groups/${id}`),
}

// =====================
// Ring Groups API
// =====================
export const ringGroupsAPI = {
    list: (params) => api.get('/ring-groups', { params }),
    get: (id) => api.get(`/ring-groups/${id}`),
    create: (data) => api.post('/ring-groups', data),
    update: (id, data) => api.put(`/ring-groups/${id}`, data),
    delete: (id) => api.delete(`/ring-groups/${id}`),
}

// =====================
// Numbers/DIDs API
// =====================
export const numbersAPI = {
    list: (params) => api.get('/numbers', { params }),
    get: (id) => api.get(`/numbers/${id}`),
    create: (data) => api.post('/numbers', data),
    update: (id, data) => api.put(`/numbers/${id}`, data),
    delete: (id) => api.delete(`/numbers/${id}`),
}

// =====================
// Routing API
// =====================
export const routingAPI = {
    listInbound: (params) => api.get('/routing/inbound', { params }),
    createInbound: (data) => api.post('/routing/inbound', data),
    listOutbound: (params) => api.get('/routing/outbound', { params }),
    createOutbound: (data) => api.post('/routing/outbound', data),
    createDefaultOutbound: () => api.post('/routing/outbound/defaults'),

    // Call Blocks
    listBlocks: () => api.get('/routing/blocks'),
    createBlock: (data) => api.post('/routing/blocks', data),
    updateBlock: (id, data) => api.put(`/routing/blocks/${id}`, data),
    deleteBlock: (id) => api.delete(`/routing/blocks/${id}`),
}

// =====================
// Dial Plans API
// =====================
export const dialPlansAPI = {
    list: (params) => api.get('/dial-plans', { params }),
    get: (id) => api.get(`/dial-plans/${id}`),
    create: (data) => api.post('/dial-plans', data),
    update: (id, data) => api.put(`/dial-plans/${id}`, data),
    delete: (id) => api.delete(`/dial-plans/${id}`),
}

// =====================
// Audio Library API
// =====================
export const audioLibraryAPI = {
    list: (params) => api.get('/audio-library', { params }),
    get: (id) => api.get(`/audio-library/${id}`),
    upload: (formData) => api.post('/audio-library', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),
    update: (id, data) => api.put(`/audio-library/${id}`, data),
    delete: (id) => api.delete(`/audio-library/${id}`),
}

// =====================
// Music on Hold API
// =====================
export const mohAPI = {
    list: (params) => api.get('/music-on-hold', { params }),
    get: (id) => api.get(`/music-on-hold/${id}`),
    create: (data) => api.post('/music-on-hold', data),
    update: (id, data) => api.put(`/music-on-hold/${id}`, data),
    delete: (id) => api.delete(`/music-on-hold/${id}`),
}

// =====================
// Recordings API
// =====================
export const recordingsAPI = {
    list: (params) => api.get('/recordings', { params }),
    get: (id) => api.get(`/recordings/${id}`),
    delete: (id) => api.delete(`/recordings/${id}`),
}

// =====================
// Audit Logs API
// =====================
export const auditLogsAPI = {
    list: (params) => api.get('/audit-logs', { params }),
}

// =====================
// Provisioning API
// =====================
export const provisioningAPI = {
    listTemplates: (params) => api.get('/provisioning-templates', { params }),
    getTemplate: (id) => api.get(`/provisioning-templates/${id}`),
    createTemplate: (data) => api.post('/provisioning-templates', data),
    updateTemplate: (id, data) => api.put(`/provisioning-templates/${id}`, data),
    deleteTemplate: (id) => api.delete(`/provisioning-templates/${id}`),
}

// =====================
// System Admin API
// =====================
export const systemAPI = {
    // Tenants
    listTenants: (params) => api.get('/system/tenants', { params }),
    getTenant: (id) => api.get(`/system/tenants/${id}`),
    createTenant: (data) => api.post('/system/tenants', data),
    updateTenant: (id, data) => api.put(`/system/tenants/${id}`, data),
    deleteTenant: (id) => api.delete(`/system/tenants/${id}`),

    // System Numbers (All Tenants)
    listAllNumbers: () => api.get('/system/numbers'),

    // Tenant Profiles
    listProfiles: () => api.get('/system/tenant-profiles'),
    getProfile: (id) => api.get(`/system/tenant-profiles/${id}`),
    createProfile: (data) => api.post('/system/tenant-profiles', data),
    updateProfile: (id, data) => api.put(`/system/tenant-profiles/${id}`, data),
    deleteProfile: (id) => api.delete(`/system/tenant-profiles/${id}`),

    // Gateways
    listGateways: () => api.get('/system/gateways'),
    createGateway: (data) => api.post('/system/gateways', data),
    updateGateway: (id, data) => api.put(`/system/gateways/${id}`, data),
    deleteGateway: (id) => api.delete(`/system/gateways/${id}`),
    getGatewayStatus: () => api.get('/system/gateways/status'),

    // Bridges
    listBridges: () => api.get('/system/bridges'),
    createBridge: (data) => api.post('/system/bridges', data),
    updateBridge: (id, data) => api.put(`/system/bridges/${id}`, data),
    deleteBridge: (id) => api.delete(`/system/bridges/${id}`),

    // SIP Profiles
    listSIPProfiles: () => api.get('/system/sip-profiles'),
    getSIPProfile: (id) => api.get(`/system/sip-profiles/${id}`),
    createSIPProfile: (data) => api.post('/system/sip-profiles', data),
    updateSIPProfile: (id, data) => api.put(`/system/sip-profiles/${id}`, data),
    deleteSIPProfile: (id) => api.delete(`/system/sip-profiles/${id}`),

    // Sofia Control (live FreeSWITCH commands)
    getSofiaStatus: () => api.get('/system/sofia/status'),
    getSofiaProfileStatus: (name) => api.get(`/system/sofia/profiles/${name}/status`),
    getSofiaProfileRegistrations: (name) => api.get(`/system/sofia/profiles/${name}/registrations`),
    getSofiaProfileGateways: (name) => api.get(`/system/sofia/profiles/${name}/gateways`),
    restartSofiaProfile: (name) => api.post(`/system/sofia/profiles/${name}/restart`),
    startSofiaProfile: (name) => api.post(`/system/sofia/profiles/${name}/start`),
    stopSofiaProfile: (name) => api.post(`/system/sofia/profiles/${name}/stop`),
    reloadSofiaXML: () => api.post('/system/sofia/reload-xml'),

    // Messaging Providers
    listMessagingProviders: () => api.get('/system/messaging-providers'),
    getMessagingProvider: (id) => api.get(`/system/messaging-providers/${id}`),
    createMessagingProvider: (data) => api.post('/system/messaging-providers', data),
    updateMessagingProvider: (id, data) => api.put(`/system/messaging-providers/${id}`, data),
    deleteMessagingProvider: (id) => api.delete(`/system/messaging-providers/${id}`),

    // Global Dial Plans
    listDialplans: () => api.get('/system/dialplans'),
    getDialplan: (id) => api.get(`/system/dialplans/${id}`),
    createDialplan: (data) => api.post('/system/dialplans', data),
    updateDialplan: (id, data) => api.put(`/system/dialplans/${id}`, data),
    deleteDialplan: (id) => api.delete(`/system/dialplans/${id}`),

    // Access Control Lists (ACLs)
    listACLs: () => api.get('/system/acls'),
    getACL: (id) => api.get(`/system/acls/${id}`),
    createACL: (data) => api.post('/system/acls', data),
    updateACL: (id, data) => api.put(`/system/acls/${id}`, data),
    deleteACL: (id) => api.delete(`/system/acls/${id}`),
    // ACL Nodes
    createACLNode: (aclId, data) => api.post(`/system/acls/${aclId}/nodes`, data),
    updateACLNode: (aclId, nodeId, data) => api.put(`/system/acls/${aclId}/nodes/${nodeId}`, data),
    deleteACLNode: (aclId, nodeId) => api.delete(`/system/acls/${aclId}/nodes/${nodeId}`),

    // Settings
    getSettings: () => api.get('/system/settings'),
    updateSettings: (data) => api.put('/system/settings', data),

    // Status
    getStatus: () => api.get('/system/status'),
    getStats: () => api.get('/system/stats'),
    getLogs: (params) => api.get('/system/logs', { params }),

    // System Media
    listSounds: () => api.get('/system/media/sounds'),
    uploadSound: (formData) => api.post('/system/media/sounds', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),
    listMusic: () => api.get('/system/media/music'),
    uploadMusic: (formData) => api.post('/system/media/music', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),

    // Security - Banned IPs
    listBannedIPs: (params) => api.get('/system/security/banned-ips', { params }),
    unbanIP: (ip) => api.delete(`/system/security/banned-ips/${ip}`),

    // Device Templates (system-level master templates)
    listDeviceTemplates: (params) => api.get('/system/device-templates', { params }),
    getDeviceTemplate: (id) => api.get(`/system/device-templates/${id}`),
    createDeviceTemplate: (data) => api.post('/system/device-templates', data),
    updateDeviceTemplate: (id, data) => api.put(`/system/device-templates/${id}`, data),
    deleteDeviceTemplate: (id) => api.delete(`/system/device-templates/${id}`),

    // Firmware Management
    listFirmware: (params) => api.get('/system/firmware', { params }),
    getFirmware: (id) => api.get(`/system/firmware/${id}`),
    createFirmware: (data) => api.post('/system/firmware', data),
    updateFirmware: (id, data) => api.put(`/system/firmware/${id}`, data),
    deleteFirmware: (id) => api.delete(`/system/firmware/${id}`),
    uploadFirmwareFile: (id, formData) => api.post(`/system/firmware/${id}/upload`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),
    setDefaultFirmware: (id) => api.post(`/system/firmware/${id}/set-default`),
}

// =====================
// Users API
// =====================
export const usersAPI = {
    list: (params) => api.get('/users', { params }),
    get: (id) => api.get(`/users/${id}`),
    create: (data) => api.post('/users', data),
    update: (id, data) => api.put(`/users/${id}`, data),
    delete: (id) => api.delete(`/users/${id}`),
}

// =====================
// Tenant Media API
// =====================
export const tenantMediaAPI = {
    listSounds: () => api.get('/media/sounds'),
    uploadSound: (formData) => api.post('/media/sounds', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),
    deleteSound: (path) => api.delete('/media/sounds', { params: { path } }),

    listMusic: () => api.get('/media/music'),
    uploadMusic: (formData) => api.post('/media/music', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),
    deleteMusic: (path) => api.delete('/media/music', { params: { path } }),
}

// =====================
// User Portal API
// =====================
export const userPortalAPI = {
    getDevices: () => api.get('/user/devices'),
    getCallHistory: (params) => api.get('/user/call-history', { params }),
    getVoicemail: () => api.get('/user/voicemail'),
    getSettings: () => api.get('/user/settings'),
    updateSettings: (data) => api.put('/user/settings', data),
    getContacts: () => api.get('/user/contacts'),
    createContact: (data) => api.post('/user/contacts', data),
}
