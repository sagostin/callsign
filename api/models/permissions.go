package models

// Permission represents a specific permission that can be granted
type Permission string

// Permission constants for granular access control
const (
	// System Admin Permissions
	PermSystemManage     Permission = "system:manage"      // Full system access
	PermTenantCreate     Permission = "tenant:create"      // Create tenants
	PermTenantDelete     Permission = "tenant:delete"      // Delete tenants
	PermTenantManageAll  Permission = "tenant:manage:all"  // Manage all tenants
	PermGatewayManage    Permission = "gateway:manage"     // Manage SIP gateways
	PermSIPProfileManage Permission = "sip_profile:manage" // Manage SIP profiles
	PermSystemLogs       Permission = "system:logs"        // View system logs
	PermSystemSettings   Permission = "system:settings"    // Manage system settings

	// Tenant Admin Permissions
	PermTenantManage    Permission = "tenant:manage"     // Manage own tenant
	PermTenantSettings  Permission = "tenant:settings"   // Manage tenant settings
	PermUserCreate      Permission = "user:create"       // Create users
	PermUserDelete      Permission = "user:delete"       // Delete users
	PermUserManage      Permission = "user:manage"       // Manage users
	PermExtensionCreate Permission = "extension:create"  // Create extensions
	PermExtensionDelete Permission = "extension:delete"  // Delete extensions
	PermExtensionManage Permission = "extension:manage"  // Manage extensions
	PermDeviceCreate    Permission = "device:create"     // Create devices
	PermDeviceDelete    Permission = "device:delete"     // Delete devices
	PermDeviceManage    Permission = "device:manage"     // Manage devices
	PermIVRManage       Permission = "ivr:manage"        // Manage IVR menus
	PermQueueManage     Permission = "queue:manage"      // Manage call queues
	PermRingGroupManage Permission = "ring_group:manage" // Manage ring groups
	PermRecordingView   Permission = "recording:view"    // View recordings
	PermRecordingDelete Permission = "recording:delete"  // Delete recordings
	PermDialplanManage  Permission = "dialplan:manage"   // Manage dialplans
	PermNumberManage    Permission = "number:manage"     // Manage phone numbers
	PermReportsView     Permission = "reports:view"      // View reports/CDRs

	// User Permissions
	PermProfileView     Permission = "profile:view"      // View own profile
	PermProfileEdit     Permission = "profile:edit"      // Edit own profile
	PermVoicemailAccess Permission = "voicemail:access"  // Access own voicemail
	PermCallHistoryView Permission = "call_history:view" // View own call history
	PermContactsManage  Permission = "contacts:manage"   // Manage personal contacts
	PermSettingsEdit    Permission = "settings:edit"     // Edit personal settings
)

// RolePermissions maps roles to their default permissions
var RolePermissions = map[UserRole][]Permission{
	RoleSystemAdmin: {
		// System admins get all permissions
		PermSystemManage,
		PermTenantCreate,
		PermTenantDelete,
		PermTenantManageAll,
		PermGatewayManage,
		PermSIPProfileManage,
		PermSystemLogs,
		PermSystemSettings,
		PermTenantManage,
		PermTenantSettings,
		PermUserCreate,
		PermUserDelete,
		PermUserManage,
		PermExtensionCreate,
		PermExtensionDelete,
		PermExtensionManage,
		PermDeviceCreate,
		PermDeviceDelete,
		PermDeviceManage,
		PermIVRManage,
		PermQueueManage,
		PermRingGroupManage,
		PermRecordingView,
		PermRecordingDelete,
		PermDialplanManage,
		PermNumberManage,
		PermReportsView,
		PermProfileView,
		PermProfileEdit,
		PermVoicemailAccess,
		PermCallHistoryView,
		PermContactsManage,
		PermSettingsEdit,
	},
	RoleTenantAdmin: {
		// Tenant admins get tenant-level permissions
		PermTenantManage,
		PermTenantSettings,
		PermUserCreate,
		PermUserDelete,
		PermUserManage,
		PermExtensionCreate,
		PermExtensionDelete,
		PermExtensionManage,
		PermDeviceCreate,
		PermDeviceDelete,
		PermDeviceManage,
		PermIVRManage,
		PermQueueManage,
		PermRingGroupManage,
		PermRecordingView,
		PermRecordingDelete,
		PermDialplanManage,
		PermNumberManage,
		PermReportsView,
		PermProfileView,
		PermProfileEdit,
		PermVoicemailAccess,
		PermCallHistoryView,
		PermContactsManage,
		PermSettingsEdit,
	},
	RoleUser: {
		// Regular users get personal permissions only
		PermProfileView,
		PermProfileEdit,
		PermVoicemailAccess,
		PermCallHistoryView,
		PermContactsManage,
		PermSettingsEdit,
	},
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(perm Permission) bool {
	// System admins have all permissions
	if u.Role == RoleSystemAdmin {
		return true
	}

	// Check role-based permissions
	permissions, ok := RolePermissions[u.Role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the given permissions
func (u *User) HasAnyPermission(perms ...Permission) bool {
	for _, perm := range perms {
		if u.HasPermission(perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all of the given permissions
func (u *User) HasAllPermissions(perms ...Permission) bool {
	for _, perm := range perms {
		if !u.HasPermission(perm) {
			return false
		}
	}
	return true
}

// GetPermissions returns all permissions for the user's role
func (u *User) GetPermissions() []Permission {
	permissions, ok := RolePermissions[u.Role]
	if !ok {
		return []Permission{}
	}
	return permissions
}
