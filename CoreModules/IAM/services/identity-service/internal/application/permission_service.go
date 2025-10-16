/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Interface de serviço para gerenciamento de permissões.
 * Define operações para criar, atualizar, excluir e consultar permissões.
 */

package application

import (
	"context"

	"github.com/google/uuid"
)

// PermissionService define a interface para gerenciar permissões no sistema
type PermissionService interface {
	// CreatePermission cria uma nova permissão
	CreatePermission(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error)

	// GetPermissionByID recupera uma permissão pelo seu ID
	GetPermissionByID(ctx context.Context, tenantID, permissionID uuid.UUID) (*PermissionResponse, error)

	// GetPermissionByCode recupera uma permissão pelo seu código
	GetPermissionByCode(ctx context.Context, tenantID uuid.UUID, code string) (*PermissionResponse, error)

	// ListPermissions lista permissões com filtros e paginação
	ListPermissions(ctx context.Context, filter PermissionFilter) (*PaginatedPermissionResponse, error)

	// UpdatePermission atualiza uma permissão existente
	UpdatePermission(ctx context.Context, req UpdatePermissionRequest) (*PermissionResponse, error)

	// DeletePermission exclui uma permissão
	DeletePermission(ctx context.Context, tenantID, permissionID uuid.UUID) error

	// AssignPermissionsToRole atribui permissões a uma função
	AssignPermissionsToRole(ctx context.Context, req AssignPermissionsToRoleRequest) error

	// RevokePermissionsFromRole revoga permissões de uma função
	RevokePermissionsFromRole(ctx context.Context, req RevokePermissionsFromRoleRequest) error

	// GetRolePermissions recupera todas as permissões de uma função
	GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) (*RolePermissionsResponse, error)

	// GetUserPermissions recupera todas as permissões de um usuário (diretas + por função)
	GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) (*UserPermissionsResponse, error)

	// CheckUserPermission verifica se um usuário tem uma permissão específica
	CheckUserPermission(ctx context.Context, tenantID, userID uuid.UUID, permissionCode string) (bool, error)
}

// CreatePermissionRequest representa os dados para criação de uma permissão
type CreatePermissionRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	Code        string    `json:"code" validate:"required,min=3,max=100"`
	Name        string    `json:"name" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	Module      string    `json:"module" validate:"required,min=3,max=100"`
	Resource    string    `json:"resource" validate:"required,min=3,max=100"`
	Action      string    `json:"action" validate:"required,min=1,max=100"`
	IsActive    bool      `json:"is_active" validate:"boolean"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePermissionRequest representa os dados para atualização de uma permissão
type UpdatePermissionRequest struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	TenantID    uuid.UUID `json:"-"`
	Name        string    `json:"name" validate:"required,min=3,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	IsActive    bool      `json:"is_active" validate:"boolean"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionFilter representa os filtros para listar permissões
type PermissionFilter struct {
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	Page        int       `json:"page" validate:"min=1"`
	PageSize    int       `json:"page_size" validate:"min=1,max=100"`
	Code        string    `json:"code,omitempty"`
	Module      string    `json:"module,omitempty"`
	Resource    string    `json:"resource,omitempty"`
	Action      string    `json:"action,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
	SearchTerm  string    `json:"search_term,omitempty"`
	OrderBy     string    `json:"order_by,omitempty"`
	Order       string    `json:"order,omitempty"`
}

// PermissionResponse representa a resposta para operações com permissões
type PermissionResponse struct {
	ID          uuid.UUID                `json:"id"`
	TenantID    uuid.UUID                `json:"tenant_id"`
	Code        string                   `json:"code"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Module      string                   `json:"module"`
	Resource    string                   `json:"resource"`
	Action      string                   `json:"action"`
	IsActive    bool                     `json:"is_active"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
	CreatedAt   string                   `json:"created_at"`
	UpdatedAt   string                   `json:"updated_at"`
}

// PaginatedPermissionResponse representa a resposta paginada para listagem de permissões
type PaginatedPermissionResponse struct {
	Items      []PermissionResponse `json:"items"`
	TotalItems int                  `json:"total_items"`
	TotalPages int                  `json:"total_pages"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
}

// AssignPermissionsToRoleRequest representa os dados para atribuir permissões a uma função
type AssignPermissionsToRoleRequest struct {
	TenantID      uuid.UUID   `json:"tenant_id" validate:"required"`
	RoleID        uuid.UUID   `json:"role_id" validate:"required"`
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required,min=1"`
}

// RevokePermissionsFromRoleRequest representa os dados para revogar permissões de uma função
type RevokePermissionsFromRoleRequest struct {
	TenantID      uuid.UUID   `json:"tenant_id" validate:"required"`
	RoleID        uuid.UUID   `json:"role_id" validate:"required"`
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required,min=1"`
}

// RolePermissionsResponse representa a resposta com as permissões de uma função
type RolePermissionsResponse struct {
	RoleID      uuid.UUID           `json:"role_id"`
	RoleName    string              `json:"role_name"`
	Permissions []PermissionResponse `json:"permissions"`
}

// UserPermissionsResponse representa a resposta com as permissões de um usuário
type UserPermissionsResponse struct {
	UserID         uuid.UUID           `json:"user_id"`
	DirectPermissions []PermissionResponse `json:"direct_permissions"`
	RolePermissions   []RolePermissionsResponse `json:"role_permissions"`
	AllPermissions    []PermissionResponse `json:"all_permissions"`
}