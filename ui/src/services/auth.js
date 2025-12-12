import { reactive, readonly } from 'vue'
import { authAPI, systemAPI } from './api'

// Auth state
// Auth state
const state = reactive({
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    token: localStorage.getItem('token') || null,
    isAuthenticated: !!localStorage.getItem('token'),
    currentTenantId: localStorage.getItem('tenantId') || null,
    tenants: [],
    isLoading: false,
    error: null,
})

// Computed permissions
const permissions = {
    isSystemAdmin: () => state.user?.role === 'system_admin',
    isTenantAdmin: () => ['system_admin', 'tenant_admin'].includes(state.user?.role),
    isUser: () => !!state.user,
    canManageExtensions: () => permissions.isTenantAdmin(),
    canManageDevices: () => permissions.isTenantAdmin(),
    canManageQueues: () => permissions.isTenantAdmin(),
    canManageIVR: () => permissions.isTenantAdmin(),
    canManageConferences: () => permissions.isTenantAdmin(),
    canViewCDR: () => permissions.isTenantAdmin(),
    canManageUsers: () => permissions.isTenantAdmin(),
    canManageTenants: () => permissions.isSystemAdmin(),
    canManageGateways: () => permissions.isSystemAdmin(),
    canViewSystemSettings: () => permissions.isSystemAdmin(),
}

// Actions
async function login(username, password) {
    state.isLoading = true
    state.error = null

    try {
        const response = await authAPI.login(username, password)
        const { token, user } = response.data

        setAuth(token, user)
        return { success: true, user }
    } catch (error) {
        state.error = error.message || 'Login failed'
        return { success: false, error: state.error }
    } finally {
        state.isLoading = false
    }
}

async function adminLogin(username, password) {
    state.isLoading = true
    state.error = null

    try {
        const response = await authAPI.adminLogin(username, password)
        const { token, user } = response.data

        setAuth(token, user)
        return { success: true, user }
    } catch (error) {
        state.error = error.message || 'Login failed'
        return { success: false, error: state.error }
    } finally {
        state.isLoading = false
    }
}

function setAuth(token, user) {
    state.token = token
    state.user = user
    state.isAuthenticated = true

    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    if (user.tenant_id) {
        localStorage.setItem('tenantId', user.tenant_id)
        state.currentTenantId = user.tenant_id
    }
}

async function logout() {
    try {
        await authAPI.logout()
    } catch (e) {
        // Ignore errors on logout
    }

    clearAuth()
}

function clearAuth() {
    state.token = null
    state.user = null
    state.isAuthenticated = false
    state.currentTenantId = null

    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('tenantId')
    localStorage.removeItem('refreshToken')
}

async function refreshProfile() {
    if (!state.isAuthenticated) return

    try {
        const response = await authAPI.getProfile()
        state.user = response.data
        localStorage.setItem('user', JSON.stringify(response.data))
    } catch (error) {
        if (error.status === 401) {
            clearAuth()
        }
    }
}

async function changePassword(currentPassword, newPassword) {
    state.isLoading = true

    try {
        await authAPI.changePassword(currentPassword, newPassword)
        return { success: true }
    } catch (error) {
        return { success: false, error: error.message }
    } finally {
        state.isLoading = false
    }
}

// Check if user has permission
function hasPermission(permission) {
    if (typeof permissions[permission] === 'function') {
        return permissions[permission]()
    }
    return false
}

// Check role
function hasRole(role) {
    if (Array.isArray(role)) {
        return role.includes(state.user?.role)
    }
    return state.user?.role === role
}

// Tenant Management
async function fetchAvailableTenants() {
    if (!permissions.isSystemAdmin() && !permissions.isTenantAdmin()) return

    try {
        // For system admins, fetch all tenants
        if (permissions.isSystemAdmin()) {
            const response = await systemAPI.listTenants()
            state.tenants = response.data.data || response.data
        }
        // For tenant admins, we might rely on a different endpoint or just the current user's tenant
        // Assuming systemAPI.listTenants handles context or returns appropriate list
        else {
            // If the API supports listing available tenants for the user context
            // state.tenants = [state.user.tenant] // Fallback
        }
    } catch (e) {
        console.error('Failed to fetch tenants', e)
    }
}

function switchTenant(tenantId) {
    if (tenantId === 'system') {
        localStorage.removeItem('tenantId')
        state.currentTenantId = null
        // Refresh page or trigger reload to clear headers
        window.location.href = '/system'
        return
    }

    const tenant = state.tenants.find(t => t.id === tenantId)
    if (tenant) {
        localStorage.setItem('tenantId', tenant.id)
        state.currentTenantId = tenant.id
        // Refresh page to apply new tenant header
        window.location.href = '/admin'
    }
}

// Export store
export const useAuth = () => ({
    state: readonly(state),
    permissions,
    login,
    adminLogin,
    logout,
    refreshProfile,
    changePassword,
    hasPermission,
    hasRole,
    clearAuth,
    fetchAvailableTenants,
    switchTenant,
})

export default useAuth
