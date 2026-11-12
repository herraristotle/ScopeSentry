package constants

// Role definitions
const (
	RoleSuperAdmin = "super_admin" // Super administrator
	RoleAdmin      = "admin"       // Administrator
	RoleUser       = "user"        // Regular user
	RoleGuest      = "guest"       // Guest
)

// Permission prefix definitions
const (
	PrefixUser    = "user"    // User management
	PrefixRole    = "role"    // Role management
	PrefixAsset   = "asset"   // Asset management
	PrefixDomain  = "domain"  // Subdomain management
	PrefixApp     = "app"     // Application management
	PrefixTask    = "task"    // Task management
	PrefixProject = "project" // Project management
	PrefixSystem  = "system"  // System management
)

// Action type definitions
const (
	ActionCreate = "create" // Create
	ActionRead   = "read"   // Read
	ActionUpdate = "update" // Update
	ActionDelete = "delete" // Delete
	ActionList   = "list"   // List
	ActionExport = "export" // Export
	ActionImport = "import" // Import
)

// Permission definitions
const (
	// User management permissions
	PermissionUserCreate = PrefixUser + ":" + ActionCreate
	PermissionUserRead   = PrefixUser + ":" + ActionRead
	PermissionUserUpdate = PrefixUser + ":" + ActionUpdate
	PermissionUserDelete = PrefixUser + ":" + ActionDelete
	PermissionUserList   = PrefixUser + ":" + ActionList

	// Role management permissions
	PermissionRoleCreate = PrefixRole + ":" + ActionCreate
	PermissionRoleRead   = PrefixRole + ":" + ActionRead
	PermissionRoleUpdate = PrefixRole + ":" + ActionUpdate
	PermissionRoleDelete = PrefixRole + ":" + ActionDelete
	PermissionRoleList   = PrefixRole + ":" + ActionList

	// Asset management permissions
	PermissionAssetCreate = PrefixAsset + ":" + ActionCreate
	PermissionAssetRead   = PrefixAsset + ":" + ActionRead
	PermissionAssetUpdate = PrefixAsset + ":" + ActionUpdate
	PermissionAssetDelete = PrefixAsset + ":" + ActionDelete
	PermissionAssetList   = PrefixAsset + ":" + ActionList
	PermissionAssetExport = PrefixAsset + ":" + ActionExport
	PermissionAssetImport = PrefixAsset + ":" + ActionImport

	// Subdomain management permissions
	PermissionDomainCreate = PrefixDomain + ":" + ActionCreate
	PermissionDomainRead   = PrefixDomain + ":" + ActionRead
	PermissionDomainUpdate = PrefixDomain + ":" + ActionUpdate
	PermissionDomainDelete = PrefixDomain + ":" + ActionDelete
	PermissionDomainList   = PrefixDomain + ":" + ActionList
	PermissionDomainExport = PrefixDomain + ":" + ActionExport
	PermissionDomainImport = PrefixDomain + ":" + ActionImport

	// Application management permissions
	PermissionAppCreate = PrefixApp + ":" + ActionCreate
	PermissionAppRead   = PrefixApp + ":" + ActionRead
	PermissionAppUpdate = PrefixApp + ":" + ActionUpdate
	PermissionAppDelete = PrefixApp + ":" + ActionDelete
	PermissionAppList   = PrefixApp + ":" + ActionList

	// Task management permissions
	PermissionTaskCreate = PrefixTask + ":" + ActionCreate
	PermissionTaskRead   = PrefixTask + ":" + ActionRead
	PermissionTaskUpdate = PrefixTask + ":" + ActionUpdate
	PermissionTaskDelete = PrefixTask + ":" + ActionDelete
	PermissionTaskList   = PrefixTask + ":" + ActionList
	PermissionTaskExport = PrefixTask + ":" + ActionExport

	// Project management permissions
	PermissionProjectCreate = PrefixProject + ":" + ActionCreate
	PermissionProjectRead   = PrefixProject + ":" + ActionRead
	PermissionProjectUpdate = PrefixProject + ":" + ActionUpdate
	PermissionProjectDelete = PrefixProject + ":" + ActionDelete
	PermissionProjectList   = PrefixProject + ":" + ActionList

	// System management permissions
	PermissionSystemConfig = PrefixSystem + ":config"
	PermissionSystemLog    = PrefixSystem + ":log"
	PermissionSystemBackup = PrefixSystem + ":backup"
)

// Role to permission mapping
var RolePermissions = map[string][]string{
	RoleSuperAdmin: {
		// User management
		PermissionUserCreate,
		PermissionUserRead,
		PermissionUserUpdate,
		PermissionUserDelete,
		PermissionUserList,

		// Role management
		PermissionRoleCreate,
		PermissionRoleRead,
		PermissionRoleUpdate,
		PermissionRoleDelete,
		PermissionRoleList,

		// Asset management
		PermissionAssetCreate,
		PermissionAssetRead,
		PermissionAssetUpdate,
		PermissionAssetDelete,
		PermissionAssetList,
		PermissionAssetExport,
		PermissionAssetImport,

		// Subdomain management
		PermissionDomainCreate,
		PermissionDomainRead,
		PermissionDomainUpdate,
		PermissionDomainDelete,
		PermissionDomainList,
		PermissionDomainExport,
		PermissionDomainImport,

		// Application management
		PermissionAppCreate,
		PermissionAppRead,
		PermissionAppUpdate,
		PermissionAppDelete,
		PermissionAppList,

		// Task management
		PermissionTaskCreate,
		PermissionTaskRead,
		PermissionTaskUpdate,
		PermissionTaskDelete,
		PermissionTaskList,
		PermissionTaskExport,

		// Project management
		PermissionProjectCreate,
		PermissionProjectRead,
		PermissionProjectUpdate,
		PermissionProjectDelete,
		PermissionProjectList,

		// System management
		PermissionSystemConfig,
		PermissionSystemLog,
		PermissionSystemBackup,
	},
	RoleAdmin: {
		// User management
		PermissionUserRead,
		PermissionUserList,

		// Asset management
		PermissionAssetRead,
		PermissionAssetList,
		PermissionAssetExport,

		// Subdomain management
		PermissionDomainRead,
		PermissionDomainList,
		PermissionDomainExport,

		// Application management
		PermissionAppRead,
		PermissionAppList,

		// Task management
		PermissionTaskCreate,
		PermissionTaskRead,
		PermissionTaskList,
		PermissionTaskExport,

		// Project management
		PermissionProjectCreate,
		PermissionProjectRead,
		PermissionProjectUpdate,
		PermissionProjectList,
	},
	RoleUser: {
		// Asset management
		PermissionAssetRead,
		PermissionAssetList,

		// Subdomain management
		PermissionDomainRead,
		PermissionDomainList,

		// Application management
		PermissionAppRead,
		PermissionAppList,

		// Task management
		PermissionTaskRead,
		PermissionTaskList,

		// Project management
		PermissionProjectRead,
		PermissionProjectList,
	},
	RoleGuest: {
		// Read-only permissions
		PermissionAssetRead,
		PermissionDomainRead,
		PermissionAppRead,
		PermissionTaskRead,
		PermissionProjectRead,
	},
} 