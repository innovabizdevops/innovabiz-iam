/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Interface de repositório para gerenciamento de funções (roles).
 * Define operações para persistir, recuperar e gerenciar funções no banco de dados.
 */

package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/innovabiz/iam/services/identity-service/internal/domain/model"
)

// RoleRepository define a interface para operações de persistência de funções
type RoleRepository interface {
	// Create cria uma nova função no banco de dados
	Create(ctx context.Context, role *model.Role) error

	// GetByID recupera uma função pelo seu ID
	GetByID(ctx context.Context, tenantID, roleID uuid.UUID) (*model.Role, error)

	// GetByCode recupera uma função pelo seu código
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Role, error)

	// List lista funções com filtros e paginação
	List(ctx context.Context, filter RoleFilter) ([]*model.Role, int, error)

	// Update atualiza uma função existente
	Update(ctx context.Context, role *model.Role) error

	// Delete exclui uma função
	Delete(ctx context.Context, tenantID, roleID uuid.UUID) error

	// AssignPermissions atribui permissões a uma função
	AssignPermissions(ctx context.Context, tenantID, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// RevokePermissions revoga permissões de uma função
	RevokePermissions(ctx context.Context, tenantID, roleID uuid.UUID, permissionIDs []uuid.UUID) error

	// GetPermissions recupera as permissões de uma função
	GetPermissions(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Permission, error)

	// GetUserRoles recupera as funções de um usuário
	GetUserRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Role, error)

	// AssignRolesToUser atribui funções a um usuário
	AssignRolesToUser(ctx context.Context, tenantID, userID uuid.UUID, roleIDs []uuid.UUID) error

	// RevokeRolesFromUser revoga funções de um usuário
	RevokeRolesFromUser(ctx context.Context, tenantID, userID uuid.UUID, roleIDs []uuid.UUID) error
	
	// IsAssociatedWithUsers verifica se uma função está associada a algum usuário
	IsAssociatedWithUsers(ctx context.Context, tenantID, roleID uuid.UUID) (bool, error)
	
	// GetByIDs recupera múltiplas funções pelos seus IDs
	GetByIDs(ctx context.Context, tenantID uuid.UUID, roleIDs []uuid.UUID) ([]*model.Role, error)
	
	// GetRolesWithPermission recupera todas as funções que têm uma permissão específica
	GetRolesWithPermission(ctx context.Context, tenantID uuid.UUID, permissionID uuid.UUID) ([]*model.Role, error)
	
	// UserHasRole verifica se um usuário tem uma função específica
	UserHasRole(ctx context.Context, tenantID, userID uuid.UUID, roleCode string) (bool, error)
}

// RoleFilter representa os filtros para listar funções
type RoleFilter struct {
	TenantID   uuid.UUID
	Page       int
	PageSize   int
	Code       string
	Name       string
	IsActive   *bool
	Type       string
	SearchTerm string
	OrderBy    string
	Order      string
}