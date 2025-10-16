package impl

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"github.com/rs/zerolog/log"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/event"
	"innovabiz/iam/identity-service/internal/domain/model"
	"innovabiz/iam/identity-service/internal/domain/repository"
)

var tracer = otel.Tracer("innovabiz.iam.application.role")

// RoleServiceImpl implementa a interface RoleService
type RoleServiceImpl struct {
	roleRepository       repository.RoleRepository
	permissionRepository repository.PermissionRepository
	eventPublisher       event.Publisher
}

// NewRoleService cria uma nova instância de RoleService
func NewRoleService(
	roleRepository repository.RoleRepository,
	permissionRepository repository.PermissionRepository,
	eventPublisher event.Publisher,
) application.RoleService {
	return &RoleServiceImpl{
		roleRepository:       roleRepository,
		permissionRepository: permissionRepository,
		eventPublisher:       eventPublisher,
	}
}

// CreateRole cria uma nova função no sistema
func (r *RoleServiceImpl) CreateRole(ctx context.Context, req application.CreateRoleRequest) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.CreateRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("code", req.Code),
		attribute.String("name", req.Name),
		attribute.String("type", req.Type),
	))
	defer span.End()

	// Verificar se já existe uma função com o mesmo código
	existingRole, err := r.roleRepository.FindByCode(ctx, req.TenantID, req.Code)
	if err == nil && existingRole != nil {
		return nil, application.ErrRoleCodeAlreadyExists
	} else if err != nil && err != repository.ErrRoleNotFound {
		return nil, fmt.Errorf("erro ao verificar existência da função: %w", err)
	}

	// Criar uma nova instância de função
	role, err := model.NewRole(
		uuid.New(),
		req.TenantID,
		req.Code,
		req.Name,
		req.Description,
		req.Type,
		req.CreatedBy,
		req.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar modelo de função: %w", err)
	}

	// Persistir no repositório
	err = r.roleRepository.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("erro ao persistir função: %w", err)
	}

	// Se a função for marcada como do sistema, sincronizar permissões do sistema
	if req.IsSystem {
		role.MarkAsSystem()
		err = r.roleRepository.Update(ctx, role)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", req.TenantID.String()).
				Str("role_id", role.ID().String()).
				Str("role_code", role.Code()).
				Msg("Erro ao marcar função como do sistema após criação")
			// Não retornamos o erro aqui, pois a função já foi criada com sucesso
		}

		// Sincronizar permissões para funções do sistema
		if req.SyncSystemPermissions {
			err = r.syncRolePermissions(ctx, role, req.PermissionCodes)
			if err != nil {
				log.Error().Err(err).
					Str("tenant_id", req.TenantID.String()).
					Str("role_id", role.ID().String()).
					Str("role_code", role.Code()).
					Msg("Erro ao sincronizar permissões da função do sistema")
				// Não retornamos o erro aqui, pois a função já foi criada com sucesso
			}
		}
	}

	// Publicar evento de criação de função
	r.publishRoleCreatedEvent(role)

	return role, nil
}

// UpdateRole atualiza uma função existente
func (r *RoleServiceImpl) UpdateRole(ctx context.Context, req application.UpdateRoleRequest) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.UpdateRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("role_id", req.ID.String()),
	))
	defer span.End()

	// Buscar função existente
	role, err := r.roleRepository.FindByID(ctx, req.TenantID, req.ID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função para atualização: %w", err)
	}

	// Verificar se o código está sendo alterado e se já existe
	if req.Code != "" && req.Code != role.Code() {
		existingRole, err := r.roleRepository.FindByCode(ctx, req.TenantID, req.Code)
		if err == nil && existingRole != nil && !existingRole.ID().String().Equals(role.ID()) {
			return nil, application.ErrRoleCodeAlreadyExists
		} else if err != nil && err != repository.ErrRoleNotFound {
			return nil, fmt.Errorf("erro ao verificar existência do código: %w", err)
		}
	}

	// Verificar quais campos precisam ser atualizados
	updated := false

	if req.Name != "" && req.Name != role.Name() {
		role.UpdateName(req.Name)
		updated = true
	}

	if req.Code != "" && req.Code != role.Code() {
		role.UpdateCode(req.Code)
		updated = true
	}

	if req.Description != nil && *req.Description != role.Description() {
		role.UpdateDescription(*req.Description)
		updated = true
	}

	if req.Type != "" && req.Type != role.Type() {
		err = role.UpdateType(req.Type)
		if err != nil {
			return nil, fmt.Errorf("erro ao atualizar tipo da função: %w", err)
		}
		updated = true
	}

	if req.IsActive != nil && *req.IsActive != role.IsActive() {
		if *req.IsActive {
			role.Activate()
		} else {
			role.Deactivate()
		}
		updated = true
	}

	if req.Metadata != nil {
		err = role.UpdateMetadata(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("erro ao atualizar metadados da função: %w", err)
		}
		updated = true
	}

	// Se o campo IsSystem estiver definido, atualizar
	if req.IsSystem != nil && *req.IsSystem != role.IsSystem() {
		if *req.IsSystem {
			role.MarkAsSystem()
		} else {
			role.UnmarkAsSystem()
		}
		updated = true
	}

	// Atualizar quem modificou
	if req.UpdatedBy != uuid.Nil {
		role.SetUpdatedBy(req.UpdatedBy)
		updated = true
	}

	// Se houve alterações, persistir no repositório
	if updated {
		err = r.roleRepository.Update(ctx, role)
		if err != nil {
			return nil, fmt.Errorf("erro ao persistir atualização da função: %w", err)
		}

		// Publicar evento de atualização de função
		r.publishRoleUpdatedEvent(role)
	}

	// Se solicitado, sincronizar permissões da função
	if req.SyncPermissions && req.PermissionCodes != nil {
		err = r.syncRolePermissions(ctx, role, req.PermissionCodes)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", req.TenantID.String()).
				Str("role_id", role.ID().String()).
				Str("role_code", role.Code()).
				Msg("Erro ao sincronizar permissões da função")
			// Não retornamos o erro aqui, pois a função já foi atualizada com sucesso
		}
	}

	return role, nil
}

// DeleteRole exclui uma função do sistema
func (r *RoleServiceImpl) DeleteRole(ctx context.Context, req application.DeleteRoleRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.DeleteRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("role_id", req.ID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	role, err := r.roleRepository.FindByID(ctx, req.TenantID, req.ID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função para exclusão: %w", err)
	}

	// Verificar se é uma função do sistema e a exclusão não é forçada
	if role.IsSystem() && !req.Force {
		return application.ErrCannotDeleteSystemRole
	}

	// Verificar se a função tem funções filhas e a exclusão não é forçada
	hasChildren, err := r.roleRepository.HasChildRoles(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("erro ao verificar funções filhas: %w", err)
	}
	if hasChildren && !req.Force {
		return application.ErrRoleHasChildren
	}

	// Verificar se existem usuários com esta função e a exclusão não é forçada
	hasUsers, err := r.roleRepository.HasUsers(ctx, req.TenantID, req.ID)
	if err != nil {
		return fmt.Errorf("erro ao verificar usuários associados: %w", err)
	}
	if hasUsers && !req.Force {
		return application.ErrRoleHasUsers
	}

	// Executar a exclusão com base no tipo solicitado
	if req.HardDelete {
		err = r.roleRepository.HardDelete(ctx, req.TenantID, req.ID)
	} else {
		err = r.roleRepository.SoftDelete(ctx, req.TenantID, req.ID, req.DeletedBy)
	}

	if err != nil {
		return fmt.Errorf("erro ao excluir função: %w", err)
	}

	// Publicar evento de exclusão de função
	r.publishRoleDeletedEvent(role, req.HardDelete)

	return nil
}

// GetRole recupera uma função pelo ID
func (r *RoleServiceImpl) GetRole(ctx context.Context, tenantID, id uuid.UUID) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetRole", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", id.String()),
	))
	defer span.End()

	role, err := r.roleRepository.FindByID(ctx, tenantID, id)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função: %w", err)
	}

	return role, nil
}

// GetRoleByCode recupera uma função pelo código
func (r *RoleServiceImpl) GetRoleByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetRoleByCode", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_code", code),
	))
	defer span.End()

	role, err := r.roleRepository.FindByCode(ctx, tenantID, code)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função por código: %w", err)
	}

	return role, nil
}

// ListRoles lista funções com filtros e paginação
func (r *RoleServiceImpl) ListRoles(
	ctx context.Context, 
	tenantID uuid.UUID, 
	filter application.RoleFilter,
	pagination application.Pagination,
) ([]*model.Role, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.ListRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.Int("page", pagination.Page),
		attribute.Int("page_size", pagination.PageSize),
	))
	defer span.End()

	// Mapear filtro de aplicação para filtro de repositório
	repoFilter := repository.RoleFilter{
		NameOrCodeContains: filter.NameOrCodeContains,
		Types:              filter.Types,
		IsActive:           filter.IsActive,
		IsSystem:           filter.IsSystem,
	}

	repoPagination := repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}

	roles, total, err := r.roleRepository.FindAll(ctx, tenantID, repoFilter, repoPagination)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao listar funções: %w", err)
	}

	return roles, total, nil
}// GetRolePermissions obtém as permissões atribuídas a uma função
func (r *RoleServiceImpl) GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetRolePermissions", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	_, err := r.roleRepository.FindByID(ctx, tenantID, roleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Buscar permissões da função
	permissions, err := r.roleRepository.GetPermissions(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar permissões da função: %w", err)
	}

	return permissions, nil
}

// AssignPermission atribui uma permissão a uma função
func (r *RoleServiceImpl) AssignPermission(ctx context.Context, req application.AssignPermissionRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.AssignPermission", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("role_id", req.RoleID.String()),
		attribute.String("permission_id", req.PermissionID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	role, err := r.roleRepository.FindByID(ctx, req.TenantID, req.RoleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Verificar se a permissão existe
	permission, err := r.permissionRepository.FindByID(ctx, req.TenantID, req.PermissionID)
	if err != nil {
		if err == repository.ErrPermissionNotFound {
			return application.ErrPermissionNotFound
		}
		return fmt.Errorf("erro ao buscar permissão: %w", err)
	}

	// Verificar se a permissão já está atribuída
	hasPermission, err := r.roleRepository.HasPermission(ctx, req.TenantID, req.RoleID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("erro ao verificar existência da permissão na função: %w", err)
	}

	if hasPermission {
		return application.ErrPermissionAlreadyAssigned
	}

	// Atribuir a permissão
	err = r.roleRepository.AssignPermission(ctx, req.TenantID, req.RoleID, req.PermissionID, req.CreatedBy)
	if err != nil {
		return fmt.Errorf("erro ao atribuir permissão à função: %w", err)
	}

	// Publicar evento de atribuição de permissão
	r.publishPermissionAssignedEvent(role, permission, req.CreatedBy)

	return nil
}

// RevokePermission remove uma permissão de uma função
func (r *RoleServiceImpl) RevokePermission(ctx context.Context, req application.RevokePermissionRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.RevokePermission", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("role_id", req.RoleID.String()),
		attribute.String("permission_id", req.PermissionID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	role, err := r.roleRepository.FindByID(ctx, req.TenantID, req.RoleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Verificar se a permissão existe
	permission, err := r.permissionRepository.FindByID(ctx, req.TenantID, req.PermissionID)
	if err != nil {
		if err == repository.ErrPermissionNotFound {
			return application.ErrPermissionNotFound
		}
		return fmt.Errorf("erro ao buscar permissão: %w", err)
	}

	// Verificar se a permissão está atribuída
	hasPermission, err := r.roleRepository.HasPermission(ctx, req.TenantID, req.RoleID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("erro ao verificar existência da permissão na função: %w", err)
	}

	if !hasPermission {
		return application.ErrPermissionNotAssigned
	}

	// Revogar a permissão
	err = r.roleRepository.RevokePermission(ctx, req.TenantID, req.RoleID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("erro ao revogar permissão da função: %w", err)
	}

	// Publicar evento de revogação de permissão
	r.publishPermissionRevokedEvent(role, permission)

	return nil
}

// SyncRolePermissions sincroniza as permissões de uma função
func (r *RoleServiceImpl) syncRolePermissions(ctx context.Context, role *model.Role, permissionCodes []string) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.syncRolePermissions", trace.WithAttributes(
		attribute.String("tenant_id", role.TenantID().String()),
		attribute.String("role_id", role.ID().String()),
		attribute.Int("permission_count", len(permissionCodes)),
	))
	defer span.End()

	if len(permissionCodes) == 0 {
		return nil
	}

	// Obter todas as permissões atuais da função
	currentPermissions, err := r.roleRepository.GetPermissions(ctx, role.TenantID(), role.ID())
	if err != nil {
		return fmt.Errorf("erro ao buscar permissões atuais da função: %w", err)
	}

	// Mapear permissões atuais por código
	currentPermissionMap := make(map[string]*model.Permission)
	for _, perm := range currentPermissions {
		currentPermissionMap[perm.Code()] = perm
	}

	// Mapear códigos desejados
	desiredPermissionCodes := make(map[string]bool)
	for _, code := range permissionCodes {
		desiredPermissionCodes[code] = true
	}

	// Identificar permissões a adicionar (estão nos códigos desejados mas não nas atuais)
	permissionsToAdd := make([]string, 0)
	for _, code := range permissionCodes {
		if _, exists := currentPermissionMap[code]; !exists {
			permissionsToAdd = append(permissionsToAdd, code)
		}
	}

	// Identificar permissões a remover (estão nas atuais mas não nos códigos desejados)
	permissionsToRemove := make([]*model.Permission, 0)
	for code, perm := range currentPermissionMap {
		if _, exists := desiredPermissionCodes[code]; !exists {
			permissionsToRemove = append(permissionsToRemove, perm)
		}
	}

	// Buscar permissões a adicionar por código
	if len(permissionsToAdd) > 0 {
		for _, code := range permissionsToAdd {
			permission, err := r.permissionRepository.FindByCode(ctx, role.TenantID(), code)
			if err != nil {
				log.Error().Err(err).
					Str("tenant_id", role.TenantID().String()).
					Str("role_id", role.ID().String()).
					Str("permission_code", code).
					Msg("Erro ao buscar permissão por código durante sincronização")
				continue
			}

			// Atribuir permissão
			err = r.roleRepository.AssignPermission(ctx, role.TenantID(), role.ID(), permission.ID(), role.CreatedBy())
			if err != nil {
				log.Error().Err(err).
					Str("tenant_id", role.TenantID().String()).
					Str("role_id", role.ID().String()).
					Str("permission_id", permission.ID().String()).
					Str("permission_code", permission.Code()).
					Msg("Erro ao atribuir permissão durante sincronização")
				continue
			}

			// Publicar evento
			r.publishPermissionAssignedEvent(role, permission, role.CreatedBy())
		}
	}

	// Remover permissões que não estão na lista
	if len(permissionsToRemove) > 0 {
		for _, perm := range permissionsToRemove {
			err = r.roleRepository.RevokePermission(ctx, role.TenantID(), role.ID(), perm.ID())
			if err != nil {
				log.Error().Err(err).
					Str("tenant_id", role.TenantID().String()).
					Str("role_id", role.ID().String()).
					Str("permission_id", perm.ID().String()).
					Str("permission_code", perm.Code()).
					Msg("Erro ao revogar permissão durante sincronização")
				continue
			}

			// Publicar evento
			r.publishPermissionRevokedEvent(role, perm)
		}
	}

	return nil
}

// AssignChildRole adiciona uma função filha a uma função pai
func (r *RoleServiceImpl) AssignChildRole(ctx context.Context, req application.AssignChildRoleRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.AssignChildRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("parent_id", req.ParentID.String()),
		attribute.String("child_id", req.ChildID.String()),
	))
	defer span.End()

	// Verificar se a função pai existe
	parentRole, err := r.roleRepository.FindByID(ctx, req.TenantID, req.ParentID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrParentRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função pai: %w", err)
	}

	// Verificar se a função filha existe
	childRole, err := r.roleRepository.FindByID(ctx, req.TenantID, req.ChildID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrChildRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função filha: %w", err)
	}

	// Verificar se as funções são do mesmo tipo
	if parentRole.Type() != childRole.Type() {
		return application.ErrRolesTypeMismatch
	}

	// Verificar se já existe a relação
	isChild, err := r.roleRepository.IsChildOf(ctx, req.TenantID, req.ParentID, req.ChildID)
	if err != nil {
		return fmt.Errorf("erro ao verificar hierarquia de funções: %w", err)
	}
	if isChild {
		return application.ErrChildRoleAlreadyAssigned
	}

	// Verificar se a adição criaria um ciclo
	wouldCreateCycle, err := r.roleRepository.IsChildOf(ctx, req.TenantID, req.ChildID, req.ParentID)
	if err != nil {
		return fmt.Errorf("erro ao verificar potencial ciclo na hierarquia: %w", err)
	}
	if wouldCreateCycle {
		return application.ErrCyclicRoleHierarchy
	}

	// Adicionar relação
	err = r.roleRepository.AddChildRole(ctx, req.TenantID, req.ParentID, req.ChildID, req.CreatedBy)
	if err != nil {
		return fmt.Errorf("erro ao adicionar função filha: %w", err)
	}

	// Publicar evento de adição de função filha
	r.publishChildRoleAssignedEvent(parentRole, childRole, req.CreatedBy)

	return nil
}

// RemoveChildRole remove uma função filha de uma função pai
func (r *RoleServiceImpl) RemoveChildRole(ctx context.Context, req application.RemoveChildRoleRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.RemoveChildRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("parent_id", req.ParentID.String()),
		attribute.String("child_id", req.ChildID.String()),
	))
	defer span.End()

	// Verificar se a função pai existe
	parentRole, err := r.roleRepository.FindByID(ctx, req.TenantID, req.ParentID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrParentRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função pai: %w", err)
	}

	// Verificar se a função filha existe
	childRole, err := r.roleRepository.FindByID(ctx, req.TenantID, req.ChildID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrChildRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função filha: %w", err)
	}

	// Verificar se existe a relação
	isChild, err := r.roleRepository.IsChildOf(ctx, req.TenantID, req.ParentID, req.ChildID)
	if err != nil {
		return fmt.Errorf("erro ao verificar hierarquia de funções: %w", err)
	}
	if !isChild {
		return application.ErrChildRoleNotAssigned
	}

	// Remover relação
	err = r.roleRepository.RemoveChildRole(ctx, req.TenantID, req.ParentID, req.ChildID)
	if err != nil {
		return fmt.Errorf("erro ao remover função filha: %w", err)
	}

	// Publicar evento de remoção de função filha
	r.publishChildRoleRemovedEvent(parentRole, childRole)

	return nil
}// GetChildRoles recupera as funções filhas de uma função
func (r *RoleServiceImpl) GetChildRoles(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetChildRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	_, err := r.roleRepository.FindByID(ctx, tenantID, roleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Buscar funções filhas
	childRoles, err := r.roleRepository.GetChildRoles(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções filhas: %w", err)
	}

	return childRoles, nil
}

// GetParentRoles recupera as funções pai de uma função
func (r *RoleServiceImpl) GetParentRoles(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetParentRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	_, err := r.roleRepository.FindByID(ctx, tenantID, roleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Buscar funções pai
	parentRoles, err := r.roleRepository.GetParentRoles(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções pai: %w", err)
	}

	return parentRoles, nil
}

// GetAncestorRoles recupera todas as funções ancestrais de uma função
func (r *RoleServiceImpl) GetAncestorRoles(ctx context.Context, tenantID, roleID uuid.UUID, maxDepth int) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetAncestorRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
		attribute.Int("max_depth", maxDepth),
	))
	defer span.End()

	// Verificar se a função existe
	_, err := r.roleRepository.FindByID(ctx, tenantID, roleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Definir profundidade máxima padrão
	if maxDepth <= 0 {
		maxDepth = 10 // valor padrão razoável
	}

	// Buscar funções ancestrais
	ancestorRoles, err := r.roleRepository.GetAncestorRoles(ctx, tenantID, roleID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções ancestrais: %w", err)
	}

	return ancestorRoles, nil
}

// GetDescendantRoles recupera todas as funções descendentes de uma função
func (r *RoleServiceImpl) GetDescendantRoles(ctx context.Context, tenantID, roleID uuid.UUID, maxDepth int) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetDescendantRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
		attribute.Int("max_depth", maxDepth),
	))
	defer span.End()

	// Verificar se a função existe
	_, err := r.roleRepository.FindByID(ctx, tenantID, roleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Definir profundidade máxima padrão
	if maxDepth <= 0 {
		maxDepth = 10 // valor padrão razoável
	}

	// Buscar funções descendentes
	descendantRoles, err := r.roleRepository.GetDescendantRoles(ctx, tenantID, roleID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções descendentes: %w", err)
	}

	return descendantRoles, nil
}

// AssignUserToRole atribui um usuário a uma função
func (r *RoleServiceImpl) AssignUserToRole(ctx context.Context, req application.AssignUserToRoleRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.AssignUserToRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("role_id", req.RoleID.String()),
		attribute.String("user_id", req.UserID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	role, err := r.roleRepository.FindByID(ctx, req.TenantID, req.RoleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Verificar se o usuário já está atribuído
	hasUser, err := r.roleRepository.HasUser(ctx, req.TenantID, req.RoleID, req.UserID)
	if err != nil {
		return fmt.Errorf("erro ao verificar associação usuário-função: %w", err)
	}
	if hasUser {
		return application.ErrUserAlreadyAssigned
	}

	// Preparar datas de ativação e expiração
	activatesAt := req.ActivatesAt
	if activatesAt.IsZero() {
		activatesAt = time.Now().UTC()
	}

	// Atribuir o usuário à função
	err = r.roleRepository.AssignUser(ctx, req.TenantID, req.RoleID, req.UserID, activatesAt, req.ExpiresAt, req.CreatedBy)
	if err != nil {
		return fmt.Errorf("erro ao atribuir usuário à função: %w", err)
	}

	// Publicar evento de atribuição de usuário
	r.publishUserAssignedEvent(role, req.UserID, activatesAt, req.ExpiresAt, req.CreatedBy)

	return nil
}

// RevokeUserFromRole remove um usuário de uma função
func (r *RoleServiceImpl) RevokeUserFromRole(ctx context.Context, req application.RevokeUserFromRoleRequest) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.RevokeUserFromRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("role_id", req.RoleID.String()),
		attribute.String("user_id", req.UserID.String()),
	))
	defer span.End()

	// Verificar se a função existe
	role, err := r.roleRepository.FindByID(ctx, req.TenantID, req.RoleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return application.ErrRoleNotFound
		}
		return fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Verificar se o usuário está atribuído
	hasUser, err := r.roleRepository.HasUser(ctx, req.TenantID, req.RoleID, req.UserID)
	if err != nil {
		return fmt.Errorf("erro ao verificar associação usuário-função: %w", err)
	}
	if !hasUser {
		return application.ErrUserNotAssigned
	}

	// Revogar o usuário da função
	err = r.roleRepository.RevokeUser(ctx, req.TenantID, req.RoleID, req.UserID)
	if err != nil {
		return fmt.Errorf("erro ao revogar usuário da função: %w", err)
	}

	// Publicar evento de revogação de usuário
	r.publishUserRevokedEvent(role, req.UserID)

	return nil
}

// GetRoleUsers recupera usuários atribuídos a uma função
func (r *RoleServiceImpl) GetRoleUsers(ctx context.Context, tenantID, roleID uuid.UUID, pagination application.Pagination) ([]*model.UserRoleAssignment, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetRoleUsers", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_id", roleID.String()),
		attribute.Int("page", pagination.Page),
		attribute.Int("page_size", pagination.PageSize),
	))
	defer span.End()

	// Verificar se a função existe
	_, err := r.roleRepository.FindByID(ctx, tenantID, roleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, 0, application.ErrRoleNotFound
		}
		return nil, 0, fmt.Errorf("erro ao buscar função: %w", err)
	}

	// Mapear paginação
	repoPagination := repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}

	// Buscar usuários da função
	users, total, err := r.roleRepository.GetUsers(ctx, tenantID, roleID, repoPagination)
	if err != nil {
		return nil, 0, fmt.Errorf("erro ao buscar usuários da função: %w", err)
	}

	return users, total, nil
}

// SyncSystemRoles sincroniza funções de sistema no banco de dados
func (r *RoleServiceImpl) SyncSystemRoles(ctx context.Context, tenantID uuid.UUID, systemRoles []application.SystemRoleDefinition) error {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.SyncSystemRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.Int("role_count", len(systemRoles)),
	))
	defer span.End()

	log.Info().
		Str("tenant_id", tenantID.String()).
		Int("role_count", len(systemRoles)).
		Msg("Iniciando sincronização de funções de sistema")

	// Para cada função do sistema na definição
	for _, sysRoleDef := range systemRoles {
		// Verificar se a função já existe
		existingRole, err := r.roleRepository.FindByCode(ctx, tenantID, sysRoleDef.Code)
		
		if err != nil && err != repository.ErrRoleNotFound {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("role_code", sysRoleDef.Code).
				Msg("Erro ao verificar existência da função de sistema")
			continue
		}

		// Se a função não existe, criar
		if err == repository.ErrRoleNotFound {
			// Criar nova função
			role, err := model.NewRole(
				uuid.New(),
				tenantID,
				sysRoleDef.Code,
				sysRoleDef.Name,
				sysRoleDef.Description,
				sysRoleDef.Type,
				sysRoleDef.CreatedBy,
				sysRoleDef.Metadata,
			)
			if err != nil {
				log.Error().Err(err).
					Str("tenant_id", tenantID.String()).
					Str("role_code", sysRoleDef.Code).
					Msg("Erro ao criar modelo de função de sistema")
				continue
			}

			// Marcar como função do sistema
			role.MarkAsSystem()

			// Persistir no repositório
			err = r.roleRepository.Create(ctx, role)
			if err != nil {
				log.Error().Err(err).
					Str("tenant_id", tenantID.String()).
					Str("role_code", sysRoleDef.Code).
					Msg("Erro ao persistir função de sistema")
				continue
			}

			// Publicar evento
			r.publishRoleCreatedEvent(role)

			// Sincronizar permissões da função
			if len(sysRoleDef.PermissionCodes) > 0 {
				err = r.syncRolePermissions(ctx, role, sysRoleDef.PermissionCodes)
				if err != nil {
					log.Error().Err(err).
						Str("tenant_id", tenantID.String()).
						Str("role_id", role.ID().String()).
						Str("role_code", role.Code()).
						Msg("Erro ao sincronizar permissões de função de sistema recém-criada")
				}
			}
		} else {
			// A função já existe, verificar se precisa ser atualizada
			updated := false

			// Verificar nome
			if existingRole.Name() != sysRoleDef.Name {
				existingRole.UpdateName(sysRoleDef.Name)
				updated = true
			}

			// Verificar descrição
			if existingRole.Description() != sysRoleDef.Description && sysRoleDef.Description != "" {
				existingRole.UpdateDescription(sysRoleDef.Description)
				updated = true
			}

			// Verificar tipo
			if existingRole.Type() != sysRoleDef.Type && sysRoleDef.Type != "" {
				err = existingRole.UpdateType(sysRoleDef.Type)
				if err == nil {
					updated = true
				} else {
					log.Error().Err(err).
						Str("tenant_id", tenantID.String()).
						Str("role_id", existingRole.ID().String()).
						Str("role_code", existingRole.Code()).
						Str("current_type", existingRole.Type()).
						Str("new_type", sysRoleDef.Type).
						Msg("Erro ao atualizar tipo da função de sistema")
				}
			}

			// Verificar metadados
			if sysRoleDef.Metadata != nil && len(sysRoleDef.Metadata) > 0 {
				err = existingRole.UpdateMetadata(sysRoleDef.Metadata)
				if err == nil {
					updated = true
				} else {
					log.Error().Err(err).
						Str("tenant_id", tenantID.String()).
						Str("role_id", existingRole.ID().String()).
						Str("role_code", existingRole.Code()).
						Msg("Erro ao atualizar metadados da função de sistema")
				}
			}

			// Garantir que esteja marcada como sistema
			if !existingRole.IsSystem() {
				existingRole.MarkAsSystem()
				updated = true
			}

			// Ativar a função se estiver inativa
			if !existingRole.IsActive() {
				existingRole.Activate()
				updated = true
			}

			// Atualizar se houve mudanças
			if updated {
				// Atualizar quem modificou
				if sysRoleDef.CreatedBy != uuid.Nil {
					existingRole.SetUpdatedBy(sysRoleDef.CreatedBy)
				}

				// Persistir atualizações
				err = r.roleRepository.Update(ctx, existingRole)
				if err != nil {
					log.Error().Err(err).
						Str("tenant_id", tenantID.String()).
						Str("role_id", existingRole.ID().String()).
						Str("role_code", existingRole.Code()).
						Msg("Erro ao atualizar função de sistema")
					continue
				}

				// Publicar evento de atualização
				r.publishRoleUpdatedEvent(existingRole)
			}

			// Sincronizar permissões se necessário
			if len(sysRoleDef.PermissionCodes) > 0 {
				err = r.syncRolePermissions(ctx, existingRole, sysRoleDef.PermissionCodes)
				if err != nil {
					log.Error().Err(err).
						Str("tenant_id", tenantID.String()).
						Str("role_id", existingRole.ID().String()).
						Str("role_code", existingRole.Code()).
						Msg("Erro ao sincronizar permissões de função de sistema existente")
				}
			}
		}

		// Processar hierarquia de funções, se definida
		if len(sysRoleDef.ParentCodes) > 0 {
			for _, parentCode := range sysRoleDef.ParentCodes {
				err = r.syncRoleHierarchy(ctx, tenantID, sysRoleDef.Code, parentCode, sysRoleDef.CreatedBy)
				if err != nil {
					log.Error().Err(err).
						Str("tenant_id", tenantID.String()).
						Str("role_code", sysRoleDef.Code).
						Str("parent_code", parentCode).
						Msg("Erro ao sincronizar hierarquia de função de sistema")
				}
			}
		}
	}

	log.Info().
		Str("tenant_id", tenantID.String()).
		Int("role_count", len(systemRoles)).
		Msg("Sincronização de funções de sistema concluída")

	return nil
}

// Método auxiliar para sincronizar a hierarquia de funções
func (r *RoleServiceImpl) syncRoleHierarchy(ctx context.Context, tenantID uuid.UUID, childCode, parentCode string, createdBy uuid.UUID) error {
	// Buscar função pai pelo código
	parentRole, err := r.roleRepository.FindByCode(ctx, tenantID, parentCode)
	if err != nil {
		return fmt.Errorf("função pai não encontrada: %w", err)
	}

	// Buscar função filha pelo código
	childRole, err := r.roleRepository.FindByCode(ctx, tenantID, childCode)
	if err != nil {
		return fmt.Errorf("função filha não encontrada: %w", err)
	}

	// Verificar se já existe a relação
	isChild, err := r.roleRepository.IsChildOf(ctx, tenantID, parentRole.ID(), childRole.ID())
	if err != nil {
		return fmt.Errorf("erro ao verificar hierarquia existente: %w", err)
	}

	// Se a relação já existe, não fazer nada
	if isChild {
		return nil
	}

	// Verificar se a adição criaria um ciclo
	wouldCreateCycle, err := r.roleRepository.IsChildOf(ctx, tenantID, childRole.ID(), parentRole.ID())
	if err != nil {
		return fmt.Errorf("erro ao verificar potencial ciclo: %w", err)
	}
	if wouldCreateCycle {
		return application.ErrCyclicRoleHierarchy
	}

	// Adicionar relação
	err = r.roleRepository.AddChildRole(ctx, tenantID, parentRole.ID(), childRole.ID(), createdBy)
	if err != nil {
		return fmt.Errorf("erro ao adicionar relação hierárquica: %w", err)
	}

	// Publicar evento
	r.publishChildRoleAssignedEvent(parentRole, childRole, createdBy)

	return nil
}// Métodos para publicação de eventos

// publishRoleCreatedEvent publica evento de criação de função
func (r *RoleServiceImpl) publishRoleCreatedEvent(role *model.Role) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRoleCreatedEvent(role)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Msg("Erro ao publicar evento de criação de função")
	}
}

// publishRoleUpdatedEvent publica evento de atualização de função
func (r *RoleServiceImpl) publishRoleUpdatedEvent(role *model.Role) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRoleUpdatedEvent(role)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Msg("Erro ao publicar evento de atualização de função")
	}
}

// publishRoleDeletedEvent publica evento de exclusão de função
func (r *RoleServiceImpl) publishRoleDeletedEvent(role *model.Role, hardDelete bool) {
	if r.eventPublisher == nil {
		return
	}
	
	var evt event.Event
	if hardDelete {
		evt = event.NewRoleHardDeletedEvent(role)
	} else {
		evt = event.NewRoleSoftDeletedEvent(role)
	}
	
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Str("delete_type", map[bool]string{true: "hard", false: "soft"}[hardDelete]).
			Msg("Erro ao publicar evento de exclusão de função")
	}
}

// publishPermissionAssignedEvent publica evento de atribuição de permissão
func (r *RoleServiceImpl) publishPermissionAssignedEvent(role *model.Role, permission *model.Permission, assignedBy uuid.UUID) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRolePermissionAssignedEvent(role, permission, assignedBy)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Str("permission_id", permission.ID().String()).
			Str("permission_code", permission.Code()).
			Msg("Erro ao publicar evento de atribuição de permissão")
	}
}

// publishPermissionRevokedEvent publica evento de revogação de permissão
func (r *RoleServiceImpl) publishPermissionRevokedEvent(role *model.Role, permission *model.Permission) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRolePermissionRevokedEvent(role, permission)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Str("permission_id", permission.ID().String()).
			Str("permission_code", permission.Code()).
			Msg("Erro ao publicar evento de revogação de permissão")
	}
}

// publishChildRoleAssignedEvent publica evento de atribuição de função filha
func (r *RoleServiceImpl) publishChildRoleAssignedEvent(parentRole, childRole *model.Role, assignedBy uuid.UUID) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRoleChildAssignedEvent(parentRole, childRole, assignedBy)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", parentRole.TenantID().String()).
			Str("parent_id", parentRole.ID().String()).
			Str("parent_code", parentRole.Code()).
			Str("child_id", childRole.ID().String()).
			Str("child_code", childRole.Code()).
			Msg("Erro ao publicar evento de atribuição de função filha")
	}
}

// publishChildRoleRemovedEvent publica evento de remoção de função filha
func (r *RoleServiceImpl) publishChildRoleRemovedEvent(parentRole, childRole *model.Role) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRoleChildRevokedEvent(parentRole, childRole)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", parentRole.TenantID().String()).
			Str("parent_id", parentRole.ID().String()).
			Str("parent_code", parentRole.Code()).
			Str("child_id", childRole.ID().String()).
			Str("child_code", childRole.Code()).
			Msg("Erro ao publicar evento de remoção de função filha")
	}
}

// publishUserAssignedEvent publica evento de atribuição de usuário
func (r *RoleServiceImpl) publishUserAssignedEvent(role *model.Role, userID uuid.UUID, activatesAt, expiresAt time.Time, assignedBy uuid.UUID) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRoleUserAssignedEvent(role, userID, activatesAt, expiresAt, assignedBy)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Str("user_id", userID.String()).
			Msg("Erro ao publicar evento de atribuição de usuário à função")
	}
}

// publishUserRevokedEvent publica evento de revogação de usuário
func (r *RoleServiceImpl) publishUserRevokedEvent(role *model.Role, userID uuid.UUID) {
	if r.eventPublisher == nil {
		return
	}
	
	evt := event.NewRoleUserRevokedEvent(role, userID)
	err := r.eventPublisher.Publish(context.Background(), evt)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", role.TenantID().String()).
			Str("role_id", role.ID().String()).
			Str("role_code", role.Code()).
			Str("user_id", userID.String()).
			Msg("Erro ao publicar evento de revogação de usuário da função")
	}
}

// GetUserRoles recupera as funções atribuídas a um usuário específico
func (r *RoleServiceImpl) GetUserRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.UserRoleAssignment, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetUserRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	))
	defer span.End()

	// Buscar funções do usuário
	userRoles, err := r.roleRepository.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções do usuário: %w", err)
	}

	return userRoles, nil
}

// GetUserActiveRoles recupera as funções ativamente atribuídas a um usuário específico
func (r *RoleServiceImpl) GetUserActiveRoles(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.UserRoleAssignment, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetUserActiveRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	))
	defer span.End()

	// Buscar funções ativas do usuário (considerando data de ativação e expiração)
	now := time.Now().UTC()
	userRoles, err := r.roleRepository.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções do usuário: %w", err)
	}

	// Filtrar apenas as funções ativas no momento atual
	activeRoles := make([]*model.UserRoleAssignment, 0, len(userRoles))
	for _, ur := range userRoles {
		// Verificar se a função está ativa
		if ur.Role.IsActive() && 
		   // Verificar se a atribuição está dentro do período válido
		   (ur.ActivatesAt.IsZero() || ur.ActivatesAt.Before(now) || ur.ActivatesAt.Equal(now)) &&
		   (ur.ExpiresAt.IsZero() || ur.ExpiresAt.After(now)) {
			activeRoles = append(activeRoles, ur)
		}
	}

	return activeRoles, nil
}

// GetSystemDefaultRoles recupera as funções padrão do sistema para um tipo específico
func (r *RoleServiceImpl) GetSystemDefaultRoles(ctx context.Context, tenantID uuid.UUID, roleType string) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.GetSystemDefaultRoles", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("role_type", roleType),
	))
	defer span.End()

	// Preparar filtro para funções do sistema e do tipo específico
	filter := repository.RoleFilter{
		Types:    []string{roleType},
		IsSystem: model.BoolPointer(true),
		IsActive: model.BoolPointer(true),
	}

	// Não aplicar paginação para obter todas as funções do sistema
	pagination := repository.Pagination{
		Page:     1,
		PageSize: 1000, // Valor alto para obter todas
	}

	// Buscar funções
	roles, _, err := r.roleRepository.FindAll(ctx, tenantID, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funções padrão do sistema: %w", err)
	}

	// Ordenar por nome para garantir consistência
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name() < roles[j].Name()
	})

	return roles, nil
}

// CloneRole cria uma cópia de uma função existente
func (r *RoleServiceImpl) CloneRole(ctx context.Context, req application.CloneRoleRequest) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleServiceImpl.CloneRole", trace.WithAttributes(
		attribute.String("tenant_id", req.TenantID.String()),
		attribute.String("source_role_id", req.SourceRoleID.String()),
	))
	defer span.End()

	// Buscar função de origem
	sourceRole, err := r.roleRepository.FindByID(ctx, req.TenantID, req.SourceRoleID)
	if err != nil {
		if err == repository.ErrRoleNotFound {
			return nil, application.ErrRoleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar função de origem: %w", err)
	}

	// Verificar se o código de destino já existe
	if req.TargetCode != "" {
		existingRole, err := r.roleRepository.FindByCode(ctx, req.TenantID, req.TargetCode)
		if err == nil && existingRole != nil {
			return nil, application.ErrRoleCodeAlreadyExists
		} else if err != nil && err != repository.ErrRoleNotFound {
			return nil, fmt.Errorf("erro ao verificar existência do código de destino: %w", err)
		}
	}

	// Criar clone da função
	cloneCode := req.TargetCode
	if cloneCode == "" {
		// Gerar código automaticamente com sufixo
		cloneCode = sourceRole.Code() + "_clone_" + time.Now().Format("20060102150405")
	}

	cloneName := req.TargetName
	if cloneName == "" {
		// Usar nome da função original com prefixo
		cloneName = "Cópia de " + sourceRole.Name()
	}

	// Criar nova função clonada
	clonedRole, err := sourceRole.Clone(
		uuid.New(),
		req.TenantID,
		cloneCode,
		cloneName,
		req.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao clonar função: %w", err)
	}

	// Persistir função clonada
	err = r.roleRepository.Create(ctx, clonedRole)
	if err != nil {
		return nil, fmt.Errorf("erro ao persistir função clonada: %w", err)
	}

	// Se solicitado, copiar permissões
	if req.CopyPermissions {
		// Buscar permissões da função original
		permissions, err := r.roleRepository.GetPermissions(ctx, req.TenantID, req.SourceRoleID)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", req.TenantID.String()).
				Str("source_role_id", req.SourceRoleID.String()).
				Str("cloned_role_id", clonedRole.ID().String()).
				Msg("Erro ao buscar permissões da função original")
		} else {
			// Atribuir cada permissão à nova função
			for _, perm := range permissions {
				err = r.roleRepository.AssignPermission(ctx, req.TenantID, clonedRole.ID(), perm.ID(), req.CreatedBy)
				if err != nil {
					log.Error().Err(err).
						Str("tenant_id", req.TenantID.String()).
						Str("cloned_role_id", clonedRole.ID().String()).
						Str("permission_id", perm.ID().String()).
						Msg("Erro ao atribuir permissão à função clonada")
					continue
				}
				
				// Publicar evento de atribuição
				r.publishPermissionAssignedEvent(clonedRole, perm, req.CreatedBy)
			}
		}
	}

	// Se solicitado, estabelecer a mesma hierarquia
	if req.CopyHierarchy {
		// Buscar funções pai
		parentRoles, err := r.roleRepository.GetParentRoles(ctx, req.TenantID, req.SourceRoleID)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", req.TenantID.String()).
				Str("source_role_id", req.SourceRoleID.String()).
				Str("cloned_role_id", clonedRole.ID().String()).
				Msg("Erro ao buscar funções pai para clonar hierarquia")
		} else {
			// Estabelecer relações com as funções pai
			for _, parentRole := range parentRoles {
				err = r.roleRepository.AddChildRole(ctx, req.TenantID, parentRole.ID(), clonedRole.ID(), req.CreatedBy)
				if err != nil {
					log.Error().Err(err).
						Str("tenant_id", req.TenantID.String()).
						Str("cloned_role_id", clonedRole.ID().String()).
						Str("parent_role_id", parentRole.ID().String()).
						Msg("Erro ao adicionar relação hierárquica para função clonada")
					continue
				}
				
				// Publicar evento
				r.publishChildRoleAssignedEvent(parentRole, clonedRole, req.CreatedBy)
			}
		}
	}

	// Publicar evento de criação
	r.publishRoleCreatedEvent(clonedRole)

	return clonedRole, nil
}