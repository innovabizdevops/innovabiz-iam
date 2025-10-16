/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Implementação do serviço de permissões.
 * Fornece lógica de negócios para gerenciar permissões e verificar autorização.
 */

package impl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/services/identity-service/internal/application"
	"github.com/innovabiz/iam/services/identity-service/internal/config"
	"github.com/innovabiz/iam/services/identity-service/internal/domain/model"
	"github.com/innovabiz/iam/services/identity-service/internal/domain/repository"
)

// PermissionServiceImpl implementa a interface PermissionService
type PermissionServiceImpl struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	userRepo       repository.UserRepository
	tracer         trace.Tracer
}

// NewPermissionService cria uma nova instância de PermissionServiceImpl
func NewPermissionService(
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
	tracer trace.Tracer,
) *PermissionServiceImpl {
	return &PermissionServiceImpl{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		userRepo:       userRepo,
		tracer:         tracer,
	}
}

// CreatePermission cria uma nova permissão
func (s *PermissionServiceImpl) CreatePermission(ctx context.Context, req application.CreatePermissionRequest) (*application.PermissionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.CreatePermission", 
		trace.WithAttributes(attribute.String("tenant_id", req.TenantID.String())),
	)
	defer span.End()

	// Validar se o código da permissão já existe
	existingPerm, err := s.permissionRepo.GetByCode(ctx, req.TenantID, req.Code)
	if err == nil && existingPerm != nil {
		return nil, application.NewAppError(
			"permission_already_exists",
			fmt.Sprintf("Permissão com código '%s' já existe", req.Code),
			nil,
			application.ErrorConflict,
		)
	} else if err != nil && !repository.IsNotFoundError(err) {
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("code", req.Code).
			Msg("Erro ao verificar permissão existente")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao verificar permissão existente",
			err,
			application.ErrorInternal,
		)
	}

	// Criar nova permissão
	permission := &model.Permission{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Module:      req.Module,
		Resource:    req.Resource,
		Action:      req.Action,
		IsActive:    req.IsActive,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Persistir no repositório
	err = s.permissionRepo.Create(ctx, permission)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("code", req.Code).
			Msg("Erro ao criar permissão")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao criar permissão",
			err,
			application.ErrorInternal,
		)
	}

	// Retornar resposta
	return s.mapPermissionToResponse(permission), nil
}

// GetPermissionByID recupera uma permissão pelo seu ID
func (s *PermissionServiceImpl) GetPermissionByID(ctx context.Context, tenantID, permissionID uuid.UUID) (*application.PermissionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.GetPermissionByID", 
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("permission_id", permissionID.String()),
		),
	)
	defer span.End()

	// Buscar permissão no repositório
	permission, err := s.permissionRepo.GetByID(ctx, tenantID, permissionID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return nil, application.NewAppError(
				"permission_not_found",
				"Permissão não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", permissionID.String()).
			Msg("Erro ao buscar permissão por ID")
			
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar permissão",
			err,
			application.ErrorInternal,
		)
	}

	// Retornar resposta
	return s.mapPermissionToResponse(permission), nil
}

// GetPermissionByCode recupera uma permissão pelo seu código
func (s *PermissionServiceImpl) GetPermissionByCode(ctx context.Context, tenantID uuid.UUID, code string) (*application.PermissionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.GetPermissionByCode", 
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("code", code),
		),
	)
	defer span.End()

	// Buscar permissão no repositório
	permission, err := s.permissionRepo.GetByCode(ctx, tenantID, code)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return nil, application.NewAppError(
				"permission_not_found",
				"Permissão não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("code", code).
			Msg("Erro ao buscar permissão por código")
			
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar permissão",
			err,
			application.ErrorInternal,
		)
	}

	// Retornar resposta
	return s.mapPermissionToResponse(permission), nil
}

// ListPermissions lista permissões com filtros e paginação
func (s *PermissionServiceImpl) ListPermissions(ctx context.Context, filter application.PermissionFilter) (*application.PaginatedPermissionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.ListPermissions", 
		trace.WithAttributes(
			attribute.String("tenant_id", filter.TenantID.String()),
			attribute.Int("page", filter.Page),
			attribute.Int("page_size", filter.PageSize),
		),
	)
	defer span.End()

	// Converter filtro para o modelo de domínio
	domainFilter := repository.PermissionFilter{
		TenantID:   filter.TenantID,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Code:       filter.Code,
		Module:     filter.Module,
		Resource:   filter.Resource,
		Action:     filter.Action,
		IsActive:   filter.IsActive,
		SearchTerm: filter.SearchTerm,
		OrderBy:    filter.OrderBy,
		Order:      filter.Order,
	}

	// Buscar permissões no repositório
	permissions, total, err := s.permissionRepo.List(ctx, domainFilter)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", filter.TenantID.String()).
			Msg("Erro ao listar permissões")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao listar permissões",
			err,
			application.ErrorInternal,
		)
	}

	// Mapear para resposta
	items := make([]application.PermissionResponse, len(permissions))
	for i, perm := range permissions {
		items[i] = *s.mapPermissionToResponse(perm)
	}

	// Calcular total de páginas
	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	// Retornar resposta paginada
	return &application.PaginatedPermissionResponse{
		Items:      items,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

// UpdatePermission atualiza uma permissão existente
func (s *PermissionServiceImpl) UpdatePermission(ctx context.Context, req application.UpdatePermissionRequest) (*application.PermissionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.UpdatePermission", 
		trace.WithAttributes(
			attribute.String("permission_id", req.ID.String()),
		),
	)
	defer span.End()

	// Buscar permissão existente
	permission, err := s.permissionRepo.GetByID(ctx, req.TenantID, req.ID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return nil, application.NewAppError(
				"permission_not_found",
				"Permissão não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("permission_id", req.ID.String()).
			Msg("Erro ao buscar permissão para atualização")
			
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar permissão",
			err,
			application.ErrorInternal,
		)
	}

	// Atualizar campos da permissão
	permission.Name = req.Name
	permission.Description = req.Description
	permission.IsActive = req.IsActive
	permission.Metadata = req.Metadata
	permission.UpdatedAt = time.Now().UTC()

	// Persistir alterações
	err = s.permissionRepo.Update(ctx, permission)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("permission_id", req.ID.String()).
			Msg("Erro ao atualizar permissão")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao atualizar permissão",
			err,
			application.ErrorInternal,
		)
	}

	// Retornar resposta
	return s.mapPermissionToResponse(permission), nil
}

// DeletePermission exclui uma permissão
func (s *PermissionServiceImpl) DeletePermission(ctx context.Context, tenantID, permissionID uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "PermissionService.DeletePermission", 
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("permission_id", permissionID.String()),
		),
	)
	defer span.End()

	// Verificar se a permissão existe
	permission, err := s.permissionRepo.GetByID(ctx, tenantID, permissionID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return application.NewAppError(
				"permission_not_found",
				"Permissão não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", permissionID.String()).
			Msg("Erro ao buscar permissão para exclusão")
			
		return application.NewAppError(
			"database_error",
			"Erro ao buscar permissão",
			err,
			application.ErrorInternal,
		)
	}

	// Verificar se a permissão está associada a funções
	isAssociated, err := s.permissionRepo.IsAssociatedWithRoles(ctx, tenantID, permissionID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", permissionID.String()).
			Msg("Erro ao verificar associações da permissão")
		return application.NewAppError(
			"database_error",
			"Erro ao verificar associações da permissão",
			err,
			application.ErrorInternal,
		)
	}

	if isAssociated {
		return application.NewAppError(
			"permission_in_use",
			"A permissão está em uso por uma ou mais funções e não pode ser excluída",
			nil,
			application.ErrorConflict,
		)
	}

	// Excluir permissão
	if err := s.permissionRepo.Delete(ctx, tenantID, permissionID); err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", permissionID.String()).
			Msg("Erro ao excluir permissão")
		return application.NewAppError(
			"database_error",
			"Erro ao excluir permissão",
			err,
			application.ErrorInternal,
		)
	}

	return nil
}

// AssignPermissionsToRole atribui permissões a uma função
func (s *PermissionServiceImpl) AssignPermissionsToRole(ctx context.Context, req application.AssignPermissionsToRoleRequest) error {
	ctx, span := s.tracer.Start(ctx, "PermissionService.AssignPermissionsToRole", 
		trace.WithAttributes(
			attribute.String("tenant_id", req.TenantID.String()),
			attribute.String("role_id", req.RoleID.String()),
			attribute.Int("permission_count", len(req.PermissionIDs)),
		),
	)
	defer span.End()

	// Verificar se a função existe
	role, err := s.roleRepo.GetByID(ctx, req.TenantID, req.RoleID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return application.NewAppError(
				"role_not_found",
				"Função não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("role_id", req.RoleID.String()).
			Msg("Erro ao buscar função para atribuição de permissões")
			
		return application.NewAppError(
			"database_error",
			"Erro ao buscar função",
			err,
			application.ErrorInternal,
		)
	}

	// Verificar se todas as permissões existem
	for _, permID := range req.PermissionIDs {
		_, err := s.permissionRepo.GetByID(ctx, req.TenantID, permID)
		if err != nil {
			if repository.IsNotFoundError(err) {
				return application.NewAppError(
					"permission_not_found",
					fmt.Sprintf("Permissão com ID %s não encontrada", permID),
					err,
					application.ErrorNotFound,
				)
			}
			
			log.Error().Err(err).
				Str("tenant_id", req.TenantID.String()).
				Str("permission_id", permID.String()).
				Msg("Erro ao verificar permissão para atribuição a função")
				
			return application.NewAppError(
				"database_error",
				"Erro ao verificar permissão",
				err,
				application.ErrorInternal,
			)
		}
	}

	// Atribuir permissões à função
	if err := s.roleRepo.AssignPermissions(ctx, req.TenantID, req.RoleID, req.PermissionIDs); err != nil {
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("role_id", req.RoleID.String()).
			Msg("Erro ao atribuir permissões à função")
		return application.NewAppError(
			"database_error",
			"Erro ao atribuir permissões à função",
			err,
			application.ErrorInternal,
		)
	}

	return nil
}

// RevokePermissionsFromRole revoga permissões de uma função
func (s *PermissionServiceImpl) RevokePermissionsFromRole(ctx context.Context, req application.RevokePermissionsFromRoleRequest) error {
	ctx, span := s.tracer.Start(ctx, "PermissionService.RevokePermissionsFromRole", 
		trace.WithAttributes(
			attribute.String("tenant_id", req.TenantID.String()),
			attribute.String("role_id", req.RoleID.String()),
			attribute.Int("permission_count", len(req.PermissionIDs)),
		),
	)
	defer span.End()

	// Verificar se a função existe
	role, err := s.roleRepo.GetByID(ctx, req.TenantID, req.RoleID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return application.NewAppError(
				"role_not_found",
				"Função não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("role_id", req.RoleID.String()).
			Msg("Erro ao buscar função para revogação de permissões")
			
		return application.NewAppError(
			"database_error",
			"Erro ao buscar função",
			err,
			application.ErrorInternal,
		)
	}

	// Revogar permissões da função
	if err := s.roleRepo.RevokePermissions(ctx, req.TenantID, req.RoleID, req.PermissionIDs); err != nil {
		log.Error().Err(err).
			Str("tenant_id", req.TenantID.String()).
			Str("role_id", req.RoleID.String()).
			Msg("Erro ao revogar permissões da função")
		return application.NewAppError(
			"database_error",
			"Erro ao revogar permissões da função",
			err,
			application.ErrorInternal,
		)
	}

	return nil
}

// GetRolePermissions recupera todas as permissões de uma função
func (s *PermissionServiceImpl) GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) (*application.RolePermissionsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.GetRolePermissions", 
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("role_id", roleID.String()),
		),
	)
	defer span.End()

	// Verificar se a função existe
	role, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return nil, application.NewAppError(
				"role_not_found",
				"Função não encontrada",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("role_id", roleID.String()).
			Msg("Erro ao buscar função para listar permissões")
			
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar função",
			err,
			application.ErrorInternal,
		)
	}

	// Buscar permissões da função
	permissions, err := s.roleRepo.GetPermissions(ctx, tenantID, roleID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("role_id", roleID.String()).
			Msg("Erro ao buscar permissões da função")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar permissões da função",
			err,
			application.ErrorInternal,
		)
	}

	// Mapear permissões para resposta
	permResponses := make([]application.PermissionResponse, len(permissions))
	for i, perm := range permissions {
		permResponses[i] = *s.mapPermissionToResponse(perm)
	}

	// Retornar resposta
	return &application.RolePermissionsResponse{
		RoleID:      role.ID,
		RoleName:    role.Name,
		Permissions: permResponses,
	}, nil
}

// GetUserPermissions recupera todas as permissões de um usuário (diretas + por função)
func (s *PermissionServiceImpl) GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) (*application.UserPermissionsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.GetUserPermissions", 
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("user_id", userID.String()),
		),
	)
	defer span.End()

	// Verificar se o usuário existe
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		if repository.IsNotFoundError(err) {
			return nil, application.NewAppError(
				"user_not_found",
				"Usuário não encontrado",
				err,
				application.ErrorNotFound,
			)
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Erro ao buscar usuário para listar permissões")
			
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar usuário",
			err,
			application.ErrorInternal,
		)
	}

	// Inicializar resposta
	response := &application.UserPermissionsResponse{
		UserID: userID,
	}

	// Buscar permissões diretas do usuário (se houver implementação)
	directPermissions, err := s.permissionRepo.GetUserDirectPermissions(ctx, tenantID, userID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Erro ao buscar permissões diretas do usuário")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar permissões diretas do usuário",
			err,
			application.ErrorInternal,
		)
	}

	// Mapear permissões diretas
	directPermResponses := make([]application.PermissionResponse, len(directPermissions))
	for i, perm := range directPermissions {
		directPermResponses[i] = *s.mapPermissionToResponse(perm)
	}
	response.DirectPermissions = directPermResponses

	// Buscar funções do usuário
	roles, err := s.roleRepo.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Erro ao buscar funções do usuário")
		return nil, application.NewAppError(
			"database_error",
			"Erro ao buscar funções do usuário",
			err,
			application.ErrorInternal,
		)
	}

	// Buscar permissões por função
	rolePermMap := make(map[uuid.UUID][]application.PermissionResponse)
	allPermMap := make(map[string]application.PermissionResponse) // Para evitar duplicatas

	// Adicionar permissões diretas ao mapa de todas as permissões
	for _, perm := range directPermResponses {
		allPermMap[perm.Code] = perm
	}

	// Adicionar permissões por função
	rolePermResponses := make([]application.RolePermissionsResponse, len(roles))
	for i, role := range roles {
		// Buscar permissões da função
		rolePerms, err := s.roleRepo.GetPermissions(ctx, tenantID, role.ID)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_id", role.ID.String()).
				Msg("Erro ao buscar permissões da função do usuário")
			return nil, application.NewAppError(
				"database_error",
				"Erro ao buscar permissões da função",
				err,
				application.ErrorInternal,
			)
		}

		// Mapear permissões da função
		permResponses := make([]application.PermissionResponse, len(rolePerms))
		for j, perm := range rolePerms {
			permResponse := s.mapPermissionToResponse(perm)
			permResponses[j] = *permResponse
			
			// Adicionar ao mapa de todas as permissões
			if perm.IsActive {
				allPermMap[perm.Code] = *permResponse
			}
		}

		// Adicionar à resposta
		rolePermResponses[i] = application.RolePermissionsResponse{
			RoleID:      role.ID,
			RoleName:    role.Name,
			Permissions: permResponses,
		}
	}
	response.RolePermissions = rolePermResponses

	// Converter mapa para slice
	allPermissions := make([]application.PermissionResponse, 0, len(allPermMap))
	for _, perm := range allPermMap {
		allPermissions = append(allPermissions, perm)
	}
	response.AllPermissions = allPermissions

	return response, nil
}

// CheckUserPermission verifica se um usuário tem uma permissão específica
func (s *PermissionServiceImpl) CheckUserPermission(ctx context.Context, tenantID, userID uuid.UUID, permissionCode string) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "PermissionService.CheckUserPermission", 
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID.String()),
			attribute.String("user_id", userID.String()),
			attribute.String("permission_code", permissionCode),
		),
	)
	defer span.End()

	// Verificar permissões diretas
	hasDirectPerm, err := s.permissionRepo.UserHasDirectPermission(ctx, tenantID, userID, permissionCode)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Str("permission_code", permissionCode).
			Msg("Erro ao verificar permissão direta do usuário")
		return false, application.NewAppError(
			"database_error",
			"Erro ao verificar permissão direta do usuário",
			err,
			application.ErrorInternal,
		)
	}

	// Se tem permissão direta, retornar true
	if hasDirectPerm {
		return true, nil
	}

	// Verificar permissões por função
	hasRolePerm, err := s.permissionRepo.UserHasPermissionViaRole(ctx, tenantID, userID, permissionCode)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Str("permission_code", permissionCode).
			Msg("Erro ao verificar permissão do usuário via função")
		return false, application.NewAppError(
			"database_error",
			"Erro ao verificar permissão do usuário via função",
			err,
			application.ErrorInternal,
		)
	}

	return hasRolePerm, nil
}

// mapPermissionToResponse converte um objeto de domínio Permission para a estrutura de resposta PermissionResponse
func (s *PermissionServiceImpl) mapPermissionToResponse(permission *model.Permission) *application.PermissionResponse {
	return &application.PermissionResponse{
		ID:          permission.ID,
		TenantID:    permission.TenantID,
		Code:        permission.Code,
		Name:        permission.Name,
		Description: permission.Description,
		Module:      permission.Module,
		Resource:    permission.Resource,
		Action:      permission.Action,
		IsActive:    permission.IsActive,
		Metadata:    permission.Metadata,
		CreatedAt:   permission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   permission.UpdatedAt.Format(time.RFC3339),
	}
}