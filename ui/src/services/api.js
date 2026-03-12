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
    (response) => {
        // Backend wraps list/get responses in { data: [...] } via iris.Map.
        // Auto-unwrap so views can use response.data directly as the payload.
        const body = response.data
        if (body && typeof body === 'object' && !Array.isArray(body) && 'data' in body) {
            response.data = body.data
            // Preserve any sibling fields (e.g. "message", "box", "interval")
            const meta = { ...body }
            delete meta.data
            if (Object.keys(meta).length > 0) {
                response._meta = meta
            }
        }
        // Guard against null (Go nil slice serialises as JSON null)
        if (response.data === null || response.data === undefined) {
            response.data = []
        }
        return response
    },
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
    login: (username, password, domain) =>
        api.post('/auth/login', { username, password, domain }),

    adminLogin: (username, password, domain) =>
        api.post('/auth/admin/login', { username, password, domain }),

    extensionLogin: (extension, password, domain) =>
        api.post('/auth/extension/login', { extension, password, domain }),

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
    // Call Handling Rules
    listCallRules: (extId) => api.get(`/extensions/${extId}/call-rules`),
    createCallRule: (extId, data) => api.post(`/extensions/${extId}/call-rules`, data),
    updateCallRule: (extId, ruleId, data) => api.put(`/extensions/${extId}/call-rules/${ruleId}`, data),
    deleteCallRule: (extId, ruleId) => api.delete(`/extensions/${extId}/call-rules/${ruleId}`),
    reorderCallRules: (extId, ruleIds) => api.post(`/extensions/${extId}/call-rules/reorder`, { rule_ids: ruleIds }),
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
    // Call Handling Rules
    listCallRules: (profileId) => api.get(`/extension-profiles/${profileId}/call-rules`),
    createCallRule: (profileId, data) => api.post(`/extension-profiles/${profileId}/call-rules`, data),
    updateCallRule: (profileId, ruleId, data) => api.put(`/extension-profiles/${profileId}/call-rules/${ruleId}`, data),
    deleteCallRule: (profileId, ruleId) => api.delete(`/extension-profiles/${profileId}/call-rules/${ruleId}`),
    reorderCallRules: (profileId, ruleIds) => api.post(`/extension-profiles/${profileId}/call-rules/reorder`, { rule_ids: ruleIds }),
}


// =====================
// Speed Dials API
// =====================
export const speedDialsAPI = {
    list: () => api.get('/speed-dials'),
    get: (id) => api.get(`/speed-dials/${id}`),
    create: (data) => api.post('/speed-dials', data),
    update: (id, data) => api.put(`/speed-dials/${id}`, data),
    delete: (id) => api.delete(`/speed-dials/${id}`),
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
    // Agent management
    listAgents: (queueId) => api.get(`/queues/${queueId}/agents`),
    addAgent: (queueId, data) => api.post(`/queues/${queueId}/agents`, data),
    removeAgent: (queueId, agentId) => api.delete(`/queues/${queueId}/agents/${agentId}`),
    pauseAgent: (queueId, agentId) => api.post(`/queues/${queueId}/agents/${agentId}/pause`),
    unpauseAgent: (queueId, agentId) => api.post(`/queues/${queueId}/agents/${agentId}/unpause`),
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
    // Live conference control
    listLive: () => api.get('/conferences/live'),
    getLive: (name) => api.get(`/conferences/live/${name}`),
    muteMember: (name, memberId) => api.post(`/conferences/live/${name}/mute/${memberId}`),
    unmuteMember: (name, memberId) => api.post(`/conferences/live/${name}/unmute/${memberId}`),
    deafMember: (name, memberId) => api.post(`/conferences/live/${name}/deaf/${memberId}`),
    undeafMember: (name, memberId) => api.post(`/conferences/live/${name}/undeaf/${memberId}`),
    kickMember: (name, memberId) => api.post(`/conferences/live/${name}/kick/${memberId}`),
    lockConference: (name) => api.post(`/conferences/live/${name}/lock`),
    unlockConference: (name) => api.post(`/conferences/live/${name}/unlock`),
    startRecording: (name) => api.post(`/conferences/live/${name}/record/start`),
    stopRecording: (name) => api.post(`/conferences/live/${name}/record/stop`),
    muteAll: (name) => api.post(`/conferences/live/${name}/mute-all`),
    unmuteAll: (name) => api.post(`/conferences/live/${name}/unmute-all`),
    setFloor: (name, memberId) => api.post(`/conferences/live/${name}/floor/${memberId}`),
    // Conference stats & sessions
    getStats: (id) => api.get(`/conferences/${id}/stats`),
    getSessions: (id) => api.get(`/conferences/${id}/sessions`),
    getSessionParticipants: (sessionId) => api.get(`/conferences/sessions/${sessionId}/participants`),
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
// Toggles (Call Flows) API
// =====================
export const togglesAPI = {
    list: () => api.get('/call-flows'),
    get: (id) => api.get(`/call-flows/${id}`),
    create: (data) => api.post('/call-flows', data),
    update: (id, data) => api.put(`/call-flows/${id}`, data),
    delete: (id) => api.delete(`/call-flows/${id}`),
    toggle: (id) => api.post(`/call-flows/${id}/toggle`),
}

// =====================
// Dial Code Check Utility
// =====================
export const checkDialCode = (code, type, excludeId = 0) =>
    api.post('/check-dial-code', { code, type, exclude_id: excludeId })

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
    // Messages
    listMessages: (ext) => api.get(`/voicemail/boxes/${ext}/messages`),
    getMessage: (id) => api.get(`/voicemail/messages/${id}`),
    deleteMessage: (id) => api.delete(`/voicemail/messages/${id}`),
    markRead: (id) => api.post(`/voicemail/messages/${id}/read`),
    streamUrl: (id) => `/api/voicemail/messages/${id}/stream`,
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
// Audit Log API
// =====================
export const auditLogAPI = {
    list: (params) => api.get('/audit-logs', { params }),
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
// Fax API
// =====================
export const faxAPI = {
    listJobs: (params) => api.get('/fax/jobs', { params }),
    listInbox: (params) => api.get('/fax/jobs', { params: { direction: 'inbound', ...params } }),
    listSent: (params) => api.get('/fax/jobs', { params: { direction: 'outbound', ...params } }),
    listPending: () => api.get('/fax/active'),
    get: (id) => api.get(`/fax/jobs/${id}`),
    send: (formData) => api.post('/fax/send', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }),
    cancel: (id) => api.post(`/fax/jobs/${id}/cancel`),
    resend: (id) => api.post(`/fax/jobs/${id}/retry`),
    delete: (id) => api.delete(`/fax/jobs/${id}`),
    download: (id) => api.get(`/fax/jobs/${id}/download`, { responseType: 'blob' }),
    getStats: () => api.get('/fax/stats'),
    // Fax boxes
    listBoxes: () => api.get('/fax/boxes'),
    createBox: (data) => api.post('/fax/boxes', data),
    getBox: (id) => api.get(`/fax/boxes/${id}`),
    updateBox: (id, data) => api.put(`/fax/boxes/${id}`, data),
    deleteBox: (id) => api.delete(`/fax/boxes/${id}`),
    // Endpoints
    listEndpoints: () => api.get('/fax/endpoints'),
    createEndpoint: (data) => api.post('/fax/endpoints', data),
    updateEndpoint: (id, data) => api.put(`/fax/endpoints/${id}`, data),
    deleteEndpoint: (id) => api.delete(`/fax/endpoints/${id}`),
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
    // Location assignment (E911)
    assignLocation: (id, data) => api.post(`/numbers/${id}/location`, data),
    unassignLocation: (id) => api.delete(`/numbers/${id}/location`),
}

// =====================
// Routing API
// =====================
export const routingAPI = {
    // Inbound routes
    listInbound: (params) => api.get('/routing/inbound', { params }),
    getInbound: (id) => api.get(`/routing/inbound/${id}`),
    createInbound: (data) => api.post('/routing/inbound', data),
    updateInbound: (id, data) => api.put(`/routing/inbound/${id}`, data),
    deleteInbound: (id) => api.delete(`/routing/inbound/${id}`),
    reorderInbound: (items) => api.post('/routing/inbound/reorder', items),
    // Outbound routes
    listOutbound: (params) => api.get('/routing/outbound', { params }),
    getOutbound: (id) => api.get(`/routing/outbound/${id}`),
    createOutbound: (data) => api.post('/routing/outbound', data),
    updateOutbound: (id, data) => api.put(`/routing/outbound/${id}`, data),
    deleteOutbound: (id) => api.delete(`/routing/outbound/${id}`),
    reorderOutbound: (items) => api.post('/routing/outbound/reorder', items),
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
    // Stream & download
    streamUrl: (id) => `/api/recordings/${id}/stream`,
    downloadUrl: (id) => `/api/recordings/${id}/download`,
    stream: (id) => api.get(`/recordings/${id}/stream`, { responseType: 'blob' }),
    download: (id) => api.get(`/recordings/${id}/download`, { responseType: 'blob' }),
    // Notes & transcription
    updateNotes: (id, data) => api.put(`/recordings/${id}/notes`, data),
    getTranscription: (id) => api.get(`/recordings/${id}/transcription`),
    getConfig: () => api.get('/recordings/config'),
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
// Tenant Settings API (current tenant)
// =====================
export const tenantSettingsAPI = {
    // Get current tenant settings
    get: () => api.get('/tenant/settings'),
    update: (data) => api.put('/tenant/settings', data),

    // Branding
    getBranding: () => api.get('/tenant/branding'),
    updateBranding: (data) => api.put('/tenant/branding', data),

    // SMTP settings
    getSmtp: () => api.get('/tenant/smtp'),
    updateSmtp: (data) => api.put('/tenant/smtp', data),
    testSmtp: () => api.post('/tenant/smtp/test'),

    // Messaging (SMS/MMS)
    getMessaging: () => api.get('/tenant/messaging'),
    updateMessaging: (data) => api.put('/tenant/messaging', data),

    // Hospitality
    getHospitality: () => api.get('/tenant/hospitality'),
    updateHospitality: (data) => api.put('/tenant/hospitality', data),

    // Locations
    listLocations: () => api.get('/tenant/locations'),
    getLocation: (id) => api.get(`/tenant/locations/${id}`),
    createLocation: (data) => api.post('/tenant/locations', data),
    updateLocation: (id, data) => api.put(`/tenant/locations/${id}`, data),
    deleteLocation: (id) => api.delete(`/tenant/locations/${id}`),
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

    // System Numbers (centralized pool)
    listSystemNumbers: (params) => api.get('/system/numbers', { params }),
    createSystemNumber: (data) => api.post('/system/numbers', data),
    getSystemNumber: (id) => api.get(`/system/numbers/${id}`),
    updateSystemNumber: (id, data) => api.put(`/system/numbers/${id}`, data),
    deleteSystemNumber: (id) => api.delete(`/system/numbers/${id}`),
    assignNumber: (id, data) => api.post(`/system/numbers/${id}/assign`, data),
    unassignNumber: (id) => api.post(`/system/numbers/${id}/unassign`),

    // Number Groups (outbound routing groups)
    listNumberGroups: (params) => api.get('/system/number-groups', { params }),
    createNumberGroup: (data) => api.post('/system/number-groups', data),
    getNumberGroup: (id) => api.get(`/system/number-groups/${id}`),
    updateNumberGroup: (id, data) => api.put(`/system/number-groups/${id}`, data),
    deleteNumberGroup: (id) => api.delete(`/system/number-groups/${id}`),
    reorderGroupGateways: (id, data) => api.post(`/system/number-groups/${id}/reorder-gateways`, data),

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
    reorderGateways: (data) => api.post('/system/gateways/reorder', data),

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
    syncSIPProfiles: () => api.post('/system/sip-profiles/sync'), // Import from disk

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

    // Messaging Numbers
    listMessagingNumbers: (params) => api.get('/system/messaging-numbers', { params }),
    createMessagingNumber: (data) => api.post('/system/messaging-numbers', data),
    updateMessagingNumber: (id, data) => api.put(`/system/messaging-numbers/${id}`, data),
    deleteMessagingNumber: (id) => api.delete(`/system/messaging-numbers/${id}`),

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

    // Device Manufacturers (configurable groupings)
    listDeviceManufacturers: () => api.get('/system/device-manufacturers'),
    createDeviceManufacturer: (data) => api.post('/system/device-manufacturers', data),
    updateDeviceManufacturer: (id, data) => api.put(`/system/device-manufacturers/${id}`, data),
    deleteDeviceManufacturer: (id) => api.delete(`/system/device-manufacturers/${id}`),

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

    // FreeSWITCH Config Files (for Config Inspector)
    listConfigFiles: (path = '') => api.get('/system/config/files', { params: { path } }),
    readConfigFile: (path) => api.get('/system/config/file', { params: { path } }),

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

// =====================
// Extension Portal API (extension-scoped, preferred for web client / extension panel)
// =====================
export const extensionPortalAPI = {
    getDevices: () => api.get('/extension/portal/devices'),
    getCallHistory: (params) => api.get('/extension/portal/call-history', { params }),
    getVoicemail: () => api.get('/extension/portal/voicemail'),
    getSettings: () => api.get('/extension/portal/settings'),
    updateSettings: (data) => api.put('/extension/portal/settings', data),
    changePassword: (current, newPass) => api.put('/extension/portal/password', { current_password: current, new_password: newPass }),
    getContacts: () => api.get('/extension/portal/contacts'),
    createContact: (data) => api.post('/extension/portal/contacts', data),
}

// =====================
// Reports & Analytics API
// =====================
export const reportsAPI = {
    callVolume: (params) => api.get('/reports/call-volume', { params }),
    agentPerformance: (params) => api.get('/reports/agent-performance', { params }),
    queueStats: () => api.get('/reports/queue-stats'),
    extensionUsage: (params) => api.get('/reports/extension-usage', { params }),
    kpi: (params) => api.get('/reports/kpi', { params }),
    numberUsage: (params) => api.get('/reports/number-usage', { params }),
    export: (params) => api.get('/reports/export', { params, responseType: 'blob' }),
}

// =====================
// Hospitality API
// =====================
export const hospitalityAPI = {
    listRooms: (params) => api.get('/hospitality/rooms', { params }),
    createRoom: (data) => api.post('/hospitality/rooms', data),
    getRoom: (id) => api.get(`/hospitality/rooms/${id}`),
    updateRoom: (id, data) => api.put(`/hospitality/rooms/${id}`, data),
    deleteRoom: (id) => api.delete(`/hospitality/rooms/${id}`),
    checkIn: (id, data) => api.post(`/hospitality/rooms/${id}/checkin`, data),
    checkOut: (id) => api.post(`/hospitality/rooms/${id}/checkout`),
    setWakeup: (id, data) => api.post(`/hospitality/rooms/${id}/wakeup`, data),
}

// =====================
// Call Broadcast API
// =====================
export const broadcastAPI = {
    list: (params) => api.get('/broadcast', { params }),
    get: (id) => api.get(`/broadcast/${id}`),
    create: (data) => api.post('/broadcast', data),
    update: (id, data) => api.put(`/broadcast/${id}`, data),
    delete: (id) => api.delete(`/broadcast/${id}`),
    start: (id) => api.post(`/broadcast/${id}/start`),
    stop: (id) => api.post(`/broadcast/${id}/stop`),
    getStats: (id) => api.get(`/broadcast/${id}/stats`),
}

// =====================
// Operator Panel API
// =====================
export const operatorPanelAPI = {
    getData: () => api.get('/operator-panel'),
}

// =====================
// Live Operations API
// =====================
export const liveAPI = {
    startRecording: (uuid) => api.post('/live/recording/start', { uuid }),
    stopRecording: (uuid) => api.post('/live/recording/stop', { uuid }),
    getActiveCalls: () => api.get('/live/calls'),
    getQueueStats: () => api.get('/live/queue-stats'),
    scheduleWakeup: (data) => api.post('/live/wakeup/schedule', data),
}

// =====================
// E911 Location API
// =====================
export const locationAPI = {
    list: () => api.get('/tenant/locations'),
    get: (id) => api.get(`/tenant/locations/${id}`),
    create: (data) => api.post('/tenant/locations', data),
    update: (id, data) => api.put(`/tenant/locations/${id}`, data),
    delete: (id) => api.delete(`/tenant/locations/${id}`),
}

