/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Interface de repositório para gerenciamento de permissões.
 * Define operações para persistir, recuperar e gerenciar permissões no banco de dados.
 */

package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/innovabiz/iam/services/identity-service/internal/domain/model"
)

// PermissionRepository define a interface para operações de persistência de permissões
type PermissionRepository interface {
	// Create cria uma nova permissão no banco de dados
	Create(ctx context.Context, permission *model.Permission) error

	// GetByID recupera uma permissão pelo seu ID
	GetByID(ctx context.Context, tenantID, permissionID uuid.UUID) (*model.Permission, error)

	// GetByCode recupera uma permissão pelo seu código
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Permission, error)

	// List lista permissões com filtros e paginação
	List(ctx context.Context, filter PermissionFilter) ([]*model.Permission, int, error)

	// Update atualiza uma permissão existente
	Update(ctx context.Context, permission *model.Permission) error

	// Delete exclui uma permissão
	Delete(ctx context.Context, tenantID, permissionID uuid.UUID) error

	// IsAssociatedWithRoles verifica se uma permissão está associada a alguma função
	IsAssociatedWithRoles(ctx context.Context, tenantID, permissionID uuid.UUID) (bool, error)

	// GetUserDirectPermissions recupera as permissões atribuídas diretamente a um usuário
	GetUserDirectPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Permission, error)

	// UserHasDirectPermission verifica se um usuário tem uma permissão específica atribuída diretamente
	UserHasDirectPermission(ctx context.Context, tenantID, userID uuid.UUID, permissionCode string) (bool, error)

	// UserHasPermissionViaRole verifica se um usuário tem uma permissão específica via função
	UserHasPermissionViaRole(ctx context.Context, tenantID, userID uuid.UUID, permissionCode string) (bool, error)
	
	// GetByIDs recupera múltiplas permissões pelos seus IDs
	GetByIDs(ctx context.Context, tenantID uuid.UUID, permissionIDs []uuid.UUID) ([]*model.Permission, error)
	
	// GetByModule recupera permissões por módulo
	GetByModule(ctx context.Context, tenantID uuid.UUID, module string) ([]*model.Permission, error)
	
	// GetByResource recupera permissões por recurso
	GetByResource(ctx context.Context, tenantID uuid.UUID, resource string) ([]*model.Permission, error)
	
	// GetByAction recupera permissões por ação
	GetByAction(ctx context.Context, tenantID uuid.UUID, action string) ([]*model.Permission, error)
}

// PermissionFilter representa os filtros para listar permissões
type PermissionFilter struct {
	TenantID   uuid.UUID
	Page       int
	PageSize   int
	Code       string
	Module     string
	Resource   string
	Action     string
	IsActive   *bool
	SearchTerm string
	OrderBy    string
	Order      string
}