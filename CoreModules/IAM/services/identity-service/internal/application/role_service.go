package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"innovabiz/iam/identity-service/internal/domain/model"
)

// Erros específicos para o serviço de função
var (
	ErrRoleNotFound            = model.ErrRoleNotFound
	ErrRoleCodeAlreadyExists   = model.ErrRoleCodeAlreadyExists
	ErrParentRoleNotFound      = model.ErrParentRoleNotFound
	ErrChildRoleNotFound       = model.ErrChildRoleNotFound
	ErrPermissionNotFound      = model.ErrPermissionNotFound
	ErrPermissionAlreadyAssigned = model.ErrPermissionAlreadyAssigned
	ErrPermissionNotAssigned   = model.ErrPermissionNotAssigned
	ErrUserAlreadyAssigned     = model.ErrUserAlreadyAssigned
	ErrUserNotAssigned         = model.ErrUserNotAssigned
	ErrChildRoleAlreadyAssigned = model.ErrChildRoleAlreadyAssigned
	ErrChildRoleNotAssigned    = model.ErrChildRoleNotAssigned
	ErrCyclicRoleHierarchy     = model.ErrCyclicRoleHierarchy
	ErrRolesTypeMismatch       = model.ErrRolesTypeMismatch
	ErrCannotDeleteSystemRole  = model.ErrCannotDeleteSystemRole
	ErrRoleHasChildren         = model.ErrRoleHasChildren
	ErrRoleHasUsers            = model.ErrRoleHasUsers
)

// Pagination representa opções de paginação
type Pagination struct {
	Page     int
	PageSize int
}

// RoleFilter representa filtros para busca de funções
type RoleFilter struct {
	NameOrCodeContains string
	Types              []string
	IsActive           *bool
	IsSystem           *bool
}

// CreateRoleRequest representa a requisição para criar uma nova função
type CreateRoleRequest struct {
	TenantID              uuid.UUID
	Code                  string
	Name                  string
	Description           string
	Type                  string
	CreatedBy             uuid.UUID
	IsSystem              bool
	Metadata              map[string]interface{}
	SyncSystemPermissions bool
	PermissionCodes       []string
}

// UpdateRoleRequest representa a requisição para atualizar uma função
type UpdateRoleRequest struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Code            string
	Name            string
	Description     *string
	Type            string
	IsActive        *bool
	IsSystem        *bool
	UpdatedBy       uuid.UUID
	Metadata        map[string]interface{}
	SyncPermissions bool
	PermissionCodes []string
}

// DeleteRoleRequest representa a requisição para excluir uma função
type DeleteRoleRequest struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	DeletedBy  uuid.UUID
	HardDelete bool
	Force      bool
}

// CloneRoleRequest representa a requisição para clonar uma função
type CloneRoleRequest struct {
	TenantID       uuid.UUID
	SourceRoleID   uuid.UUID
	TargetCode     string
	TargetName     string
	CopyPermissions bool
	CopyHierarchy  bool
	CreatedBy      uuid.UUID
}

// UserRoleAssignment representa a atribuição de um usuário a uma função
type UserRoleAssignment struct {
	UserID      uuid.UUID
	Role        *model.Role
	ActivatesAt time.Time
	ExpiresAt   *time.Time
	AssignedAt  time.Time
	AssignedBy  uuid.UUID
}

// UserRoleDetail representa os detalhes de um usuário atribuído a uma função
type UserRoleDetail struct {
	UserID      uuid.UUID
	ActivatesAt time.Time
	ExpiresAt   *time.Time
	AssignedAt  time.Time
	AssignedBy  uuid.UUID
}

// RoleService define a interface de serviço para gerenciamento de funções
type RoleService interface {
	// Operações básicas de CRUD
	CreateRole(ctx context.Context, req CreateRoleRequest) (*model.Role, error)
	GetRole(ctx context.Context, tenantID, roleID uuid.UUID) (*model.Role, error)
	GetRoleByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Role, error)
	UpdateRole(ctx context.Context, req UpdateRoleRequest) (*model.Role, error)
	DeleteRole(ctx context.Context, req DeleteRoleRequest) error
	ListRoles(ctx context.Context, tenantID uuid.UUID, filter RoleFilter, pagination Pagination) ([]*model.Role, int64, error)

	// Operações de gerenciamento de permissões
	GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID, pagination Pagination) ([]*model.Permission, int64, error)
	AssignPermission(ctx context.Context, tenantID, roleID, permissionID, assignedBy uuid.UUID) error
	RevokePermission(ctx context.Context, tenantID, roleID, permissionID, revokedBy uuid.UUID) error

	// Operações de gerenciamento de hierarquia
	GetChildRoles(ctx context.Context, tenantID, roleID uuid.UUID, pagination Pagination) ([]*model.Role, int64, error)
	GetParentRoles(ctx context.Context, tenantID, roleID uuid.UUID, pagination Pagination) ([]*model.Role, int64, error)
	GetAncestorRoles(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Role, error)
	GetDescendantRoles(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Role, error)
	AssignChildRole(ctx context.Context, tenantID, parentRoleID, childRoleID, assignedBy uuid.UUID) error
	RemoveChildRole(ctx context.Context, tenantID, parentRoleID, childRoleID, removedBy uuid.UUID) error

	// Operações de gerenciamento de usuários
	AssignUserToRole(ctx context.Context, tenantID, roleID, userID uuid.UUID, activatesAt time.Time, expiresAt *time.Time, assignedBy uuid.UUID) error
	RevokeUserFromRole(ctx context.Context, tenantID, roleID, userID, revokedBy uuid.UUID) error
	GetRoleUsers(ctx context.Context, tenantID, roleID uuid.UUID, activeOnly bool, pagination Pagination) ([]UserRoleDetail, int64, error)
	GetUserRoles(ctx context.Context, tenantID, userID uuid.UUID, pagination Pagination) ([]UserRoleAssignment, int64, error)
	GetUserActiveRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]UserRoleAssignment, error)

	// Operações de funções do sistema
	SyncSystemRoles(ctx context.Context, tenantID uuid.UUID, systemRoles []model.SystemRoleDefinition, syncedBy uuid.UUID) (int, int, error)
	
	// Operação de clonagem
	CloneRole(ctx context.Context, req CloneRoleRequest) (*model.Role, error)
}