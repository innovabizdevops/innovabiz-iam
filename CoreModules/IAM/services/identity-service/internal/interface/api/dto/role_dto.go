package dto

import (
	"time"
)

// CreateRoleRequest representa a requisição para criar uma nova função
type CreateRoleRequest struct {
	Code                  string                 `json:"code"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description,omitempty"`
	Type                  string                 `json:"type"`
	IsSystem              bool                   `json:"is_system,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	SyncSystemPermissions bool                   `json:"sync_system_permissions,omitempty"`
	PermissionCodes       []string               `json:"permission_codes,omitempty"`
}

// UpdateRoleRequest representa a requisição para atualizar uma função existente
type UpdateRoleRequest struct {
	Code            string                 `json:"code,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Type            string                 `json:"type,omitempty"`
	IsActive        *bool                  `json:"is_active,omitempty"`
	IsSystem        *bool                  `json:"is_system,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	SyncPermissions bool                   `json:"sync_permissions,omitempty"`
	PermissionCodes []string               `json:"permission_codes,omitempty"`
}

// RoleResponse representa a resposta de uma função
type RoleResponse struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	IsActive    bool                   `json:"is_active"`
	IsSystem    bool                   `json:"is_system"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
	DeletedBy   string                 `json:"deleted_by,omitempty"`
}

// RoleListResponse representa a resposta de uma listagem paginada de funções
type RoleListResponse struct {
	Items      []RoleResponse `json:"items"`
	TotalItems int64          `json:"total_items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// RoleFilterRequest representa filtros para busca de funções
type RoleFilterRequest struct {
	NameOrCodeContains string   `json:"name_or_code_contains,omitempty"`
	Types              []string `json:"types,omitempty"`
	IsActive           *bool    `json:"is_active,omitempty"`
	IsSystem           *bool    `json:"is_system,omitempty"`
	Page               int      `json:"page,omitempty"`
	PageSize           int      `json:"page_size,omitempty"`
}

// AssignPermissionRequest representa a requisição para atribuir uma permissão a uma função
type AssignPermissionRequest struct {
	// O permission_id virá do path param
}

// RevokePermissionRequest representa a requisição para revogar uma permissão de uma função
type RevokePermissionRequest struct {
	// O permission_id virá do path param
}

// PermissionResponse representa a resposta de uma permissão
type PermissionResponse struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ResourceType string                `json:"resource_type"`
	Action      string                 `json:"action"`
	Effect      string                 `json:"effect"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsActive    bool                   `json:"is_active"`
	IsSystem    bool                   `json:"is_system"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
}

// AssignChildRoleRequest representa a requisição para atribuir uma função filha
type AssignChildRoleRequest struct {
	// Os role_ids virão dos path params
}

// RemoveChildRoleRequest representa a requisição para remover uma função filha
type RemoveChildRoleRequest struct {
	// Os role_ids virão dos path params
}

// AssignUserToRoleRequest representa a requisição para atribuir um usuário a uma função
type AssignUserToRoleRequest struct {
	ActivatesAt time.Time `json:"activates_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	// O user_id virá do path param
}

// RevokeUserFromRoleRequest representa a requisição para revogar um usuário de uma função
type RevokeUserFromRoleRequest struct {
	// O user_id virá do path param
}

// UserRoleResponse representa a resposta de uma atribuição de usuário a função
type UserRoleResponse struct {
	RoleID      string     `json:"role_id"`
	RoleCode    string     `json:"role_code"`
	RoleName    string     `json:"role_name"`
	Type        string     `json:"type"`
	IsActive    bool       `json:"is_active"`
	ActivatesAt time.Time  `json:"activates_at"`
	ExpiresAt   time.Time  `json:"expires_at,omitempty"`
	AssignedAt  time.Time  `json:"assigned_at"`
	AssignedBy  string     `json:"assigned_by"`
}

// UserRoleListResponse representa a resposta de uma listagem paginada de usuários em uma função
type UserRoleListResponse struct {
	Items      []UserRoleDetailResponse `json:"items"`
	TotalItems int64                    `json:"total_items"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// UserRoleDetailResponse representa detalhes de um usuário atribuído a uma função
type UserRoleDetailResponse struct {
	UserID      string     `json:"user_id"`
	ActivatesAt time.Time  `json:"activates_at"`
	ExpiresAt   time.Time  `json:"expires_at,omitempty"`
	AssignedAt  time.Time  `json:"assigned_at"`
	AssignedBy  string     `json:"assigned_by"`
}

// CloneRoleRequest representa a requisição para clonar uma função
type CloneRoleRequest struct {
	TargetCode     string `json:"target_code,omitempty"`
	TargetName     string `json:"target_name,omitempty"`
	CopyPermissions bool   `json:"copy_permissions,omitempty"`
	CopyHierarchy  bool   `json:"copy_hierarchy,omitempty"`
}

// SystemRoleDefinition representa a definição de uma função do sistema para sincronização
type SystemRoleDefinition struct {
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Type            string                 `json:"type"`
	PermissionCodes []string               `json:"permission_codes,omitempty"`
	ParentCodes     []string               `json:"parent_codes,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SyncSystemRolesRequest representa a requisição para sincronizar funções do sistema
type SyncSystemRolesRequest struct {
	SystemRoles []SystemRoleDefinition `json:"system_roles"`
}

// DeleteRoleRequest representa a requisição para excluir uma função
type DeleteRoleRequest struct {
	HardDelete bool `json:"hard_delete,omitempty"`
	Force      bool `json:"force,omitempty"`
}