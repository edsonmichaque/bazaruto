package authorization

import (
	"fmt"
	"strings"
)

// Role represents a user role
type Role string

const (
	// RoleAdmin represents an admin user
	RoleAdmin Role = "admin"

	// RoleAgent represents an insurance agent
	RoleAgent Role = "agent"

	// RoleCustomer represents a customer
	RoleCustomer Role = "customer"
)

// Permission represents a system permission
type Permission string

const (
	// Product permissions
	PermissionProductCreate Permission = "product:create"
	PermissionProductRead   Permission = "product:read"
	PermissionProductUpdate Permission = "product:update"
	PermissionProductDelete Permission = "product:delete"
	PermissionProductList   Permission = "product:list"

	// Quote permissions
	PermissionQuoteCreate Permission = "quote:create"
	PermissionQuoteRead   Permission = "quote:read"
	PermissionQuoteUpdate Permission = "quote:update"
	PermissionQuoteDelete Permission = "quote:delete"
	PermissionQuoteList   Permission = "quote:list"
	PermissionQuoteBulk   Permission = "quote:bulk"

	// Policy permissions
	PermissionPolicyCreate Permission = "policy:create"
	PermissionPolicyRead   Permission = "policy:read"
	PermissionPolicyUpdate Permission = "policy:update"
	PermissionPolicyDelete Permission = "policy:delete"
	PermissionPolicyList   Permission = "policy:list"
	PermissionPolicyRenew  Permission = "policy:renew"
	PermissionPolicyCancel Permission = "policy:cancel"

	// Claim permissions
	PermissionClaimCreate  Permission = "claim:create"
	PermissionClaimRead    Permission = "claim:read"
	PermissionClaimUpdate  Permission = "claim:update"
	PermissionClaimDelete  Permission = "claim:delete"
	PermissionClaimList    Permission = "claim:list"
	PermissionClaimApprove Permission = "claim:approve"
	PermissionClaimReject  Permission = "claim:reject"

	// User permissions
	PermissionUserCreate Permission = "user:create"
	PermissionUserRead   Permission = "user:read"
	PermissionUserUpdate Permission = "user:update"
	PermissionUserDelete Permission = "user:delete"
	PermissionUserList   Permission = "user:list"

	// Admin permissions
	PermissionAdminAccess Permission = "admin:access"
	PermissionAdminAudit  Permission = "admin:audit"
	PermissionAdminConfig Permission = "admin:config"

	// Document permissions
	PermissionDocumentUpload Permission = "document:upload"
	PermissionDocumentRead   Permission = "document:read"
	PermissionDocumentDelete Permission = "document:delete"
)

// RBACService handles role-based access control
type RBACService struct {
	rolePermissions map[Role][]Permission
}

// NewRBACService creates a new RBAC service
func NewRBACService() *RBACService {
	service := &RBACService{
		rolePermissions: make(map[Role][]Permission),
	}

	// Initialize role permissions
	service.initializeRolePermissions()

	return service
}

// initializeRolePermissions sets up the default role permissions
func (r *RBACService) initializeRolePermissions() {
	// Admin has all permissions
	r.rolePermissions[RoleAdmin] = []Permission{
		PermissionProductCreate, PermissionProductRead, PermissionProductUpdate, PermissionProductDelete, PermissionProductList,
		PermissionQuoteCreate, PermissionQuoteRead, PermissionQuoteUpdate, PermissionQuoteDelete, PermissionQuoteList, PermissionQuoteBulk,
		PermissionPolicyCreate, PermissionPolicyRead, PermissionPolicyUpdate, PermissionPolicyDelete, PermissionPolicyList, PermissionPolicyRenew, PermissionPolicyCancel,
		PermissionClaimCreate, PermissionClaimRead, PermissionClaimUpdate, PermissionClaimDelete, PermissionClaimList, PermissionClaimApprove, PermissionClaimReject,
		PermissionUserCreate, PermissionUserRead, PermissionUserUpdate, PermissionUserDelete, PermissionUserList,
		PermissionAdminAccess, PermissionAdminAudit, PermissionAdminConfig,
		PermissionDocumentUpload, PermissionDocumentRead, PermissionDocumentDelete,
	}

	// Agent has most permissions except admin and user management
	r.rolePermissions[RoleAgent] = []Permission{
		PermissionProductRead, PermissionProductList,
		PermissionQuoteCreate, PermissionQuoteRead, PermissionQuoteUpdate, PermissionQuoteList, PermissionQuoteBulk,
		PermissionPolicyCreate, PermissionPolicyRead, PermissionPolicyUpdate, PermissionPolicyList, PermissionPolicyRenew, PermissionPolicyCancel,
		PermissionClaimCreate, PermissionClaimRead, PermissionClaimUpdate, PermissionClaimList, PermissionClaimApprove, PermissionClaimReject,
		PermissionUserRead, PermissionUserList,
		PermissionDocumentUpload, PermissionDocumentRead, PermissionDocumentDelete,
	}

	// Customer has limited permissions
	r.rolePermissions[RoleCustomer] = []Permission{
		PermissionProductRead, PermissionProductList,
		PermissionQuoteCreate, PermissionQuoteRead, PermissionQuoteList,
		PermissionPolicyRead, PermissionPolicyList,
		PermissionClaimCreate, PermissionClaimRead, PermissionClaimList,
		PermissionUserRead, PermissionUserUpdate, // Can read and update own profile
		PermissionDocumentUpload, PermissionDocumentRead,
	}
}

// GetPermissions returns all permissions for a role
func (r *RBACService) GetPermissions(role Role) []Permission {
	permissions, exists := r.rolePermissions[role]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// HasPermission checks if a role has a specific permission
func (r *RBACService) HasPermission(role Role, permission Permission) bool {
	permissions := r.GetPermissions(role)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a role has any of the specified permissions
func (r *RBACService) HasAnyPermission(role Role, permissions []Permission) bool {
	for _, permission := range permissions {
		if r.HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the specified permissions
func (r *RBACService) HasAllPermissions(role Role, permissions []Permission) bool {
	for _, permission := range permissions {
		if !r.HasPermission(role, permission) {
			return false
		}
	}
	return true
}

// ValidateRole validates if a role is valid
func (r *RBACService) ValidateRole(role string) (Role, error) {
	normalizedRole := Role(strings.ToLower(role))

	switch normalizedRole {
	case RoleAdmin, RoleAgent, RoleCustomer:
		return normalizedRole, nil
	default:
		return "", fmt.Errorf("invalid role: %s", role)
	}
}

// GetValidRoles returns all valid roles
func (r *RBACService) GetValidRoles() []Role {
	return []Role{RoleAdmin, RoleAgent, RoleCustomer}
}

// IsAdmin checks if a role is admin
func (r *RBACService) IsAdmin(role Role) bool {
	return role == RoleAdmin
}

// IsAgent checks if a role is agent
func (r *RBACService) IsAgent(role Role) bool {
	return role == RoleAgent
}

// IsCustomer checks if a role is customer
func (r *RBACService) IsCustomer(role Role) bool {
	return role == RoleCustomer
}

// CanAccessResource checks if a role can access a specific resource
func (r *RBACService) CanAccessResource(role Role, resource, action string) bool {
	permission := Permission(fmt.Sprintf("%s:%s", resource, action))
	return r.HasPermission(role, permission)
}
