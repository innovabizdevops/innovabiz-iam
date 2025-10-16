/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Implementação do repositório de permissões (PermissionRepository) para PostgreSQL.
 * Gerencia persistência, consulta e relacionamentos para o domínio de permissões.
 * Implementa Row-Level Security (RLS) para isolamento multi-tenant.
 * Segue princípios Clean Architecture, SOLID, e Domain-Driven Design.
 */

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/internal/domain/model"
	"github.com/innovabiz/iam/internal/domain/repository"
)

// PermissionRepository implementação PostgreSQL do repositório de permissões
type PermissionRepository struct {
	db *pgxpool.Pool
}

// NewPermissionRepository cria uma nova instância do repositório de permissões PostgreSQL
func NewPermissionRepository(db *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{
		db: db,
	}
}

// permissionEntity representa a estrutura da tabela de permissões no banco de dados
type permissionEntity struct {
	ID           uuid.UUID              `db:"id"`
	TenantID     uuid.UUID              `db:"tenant_id"`
	Code         string                 `db:"code"`
	Name         string                 `db:"name"`
	Description  string                 `db:"description"`
	ResourceType string                 `db:"resource_type"`
	ResourceID   *string                `db:"resource_id"`
	Action       string                 `db:"action"`
	IsActive     bool                   `db:"is_active"`
	IsSystem     bool                   `db:"is_system"`
	CreatedAt    time.Time              `db:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at"`
	CreatedBy    uuid.UUID              `db:"created_by"`
	UpdatedBy    *uuid.UUID             `db:"updated_by"`
	DeletedAt    *time.Time             `db:"deleted_at"`
	DeletedBy    *uuid.UUID             `db:"deleted_by"`
	Metadata     map[string]interface{} `db:"metadata"`
}

// mapToEntity converte um modelo de domínio Permission para uma entidade do banco de dados
func mapPermissionToEntity(permission *model.Permission) permissionEntity {
	return permissionEntity{
		ID:           permission.ID(),
		TenantID:     permission.TenantID(),
		Code:         permission.Code(),
		Name:         permission.Name(),
		Description:  permission.Description(),
		ResourceType: permission.ResourceType(),
		ResourceID:   permission.ResourceID(),
		Action:       permission.Action(),
		IsActive:     permission.IsActive(),
		IsSystem:     permission.IsSystem(),
		CreatedAt:    permission.CreatedAt(),
		UpdatedAt:    permission.UpdatedAt(),
		CreatedBy:    permission.CreatedBy(),
		UpdatedBy:    permission.UpdatedBy(),
		DeletedAt:    permission.DeletedAt(),
		DeletedBy:    permission.DeletedBy(),
		Metadata:     permission.Metadata(),
	}
}

// mapToDomain converte uma entidade do banco de dados para um modelo de domínio Permission
func mapPermissionToDomain(entity permissionEntity) (*model.Permission, error) {
	// Nota: Em um cenário real, usaríamos o construtor do modelo de domínio
	// Aqui simulamos um construtor simplificado para o modelo
	permission := &model.Permission{
		ID_:           entity.ID,
		TenantID_:     entity.TenantID,
		Code_:         entity.Code,
		Name_:         entity.Name,
		Description_:  entity.Description,
		ResourceType_: entity.ResourceType,
		ResourceID_:   entity.ResourceID,
		Action_:       entity.Action,
		IsActive_:     entity.IsActive,
		IsSystem_:     entity.IsSystem,
		CreatedAt_:    entity.CreatedAt,
		UpdatedAt_:    entity.UpdatedAt,
		CreatedBy_:    entity.CreatedBy,
		UpdatedBy_:    entity.UpdatedBy,
		DeletedAt_:    entity.DeletedAt,
		DeletedBy_:    entity.DeletedBy,
		Metadata_:     entity.Metadata,
	}

	return permission, nil
}// Create implementa a criação de uma nova permissão no banco de dados
func (r *PermissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	ctx, span := tracer.Start(ctx, "PermissionRepository.Create", trace.WithAttributes(
		attribute.String("tenant_id", permission.TenantID().String()),
		attribute.String("permission_id", permission.ID().String()),
		attribute.String("permission_code", permission.Code()),
	))
	defer span.End()

	// Mapear modelo de domínio para entidade de banco de dados
	entity := mapPermissionToEntity(permission)

	// Consulta SQL para inserir nova permissão
	query := `
		INSERT INTO iam.permissions (
			id, tenant_id, code, name, description, resource_type, resource_id, 
			action, is_active, is_system, created_at, updated_at, 
			created_by, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, 
			$8, $9, $10, $11, $12, 
			$13, $14
		)
	`

	// Executar consulta com os parâmetros da entidade
	_, err := r.db.Exec(ctx, query,
		entity.ID, entity.TenantID, entity.Code, entity.Name, entity.Description, 
		entity.ResourceType, entity.ResourceID, entity.Action,
		entity.IsActive, entity.IsSystem, entity.CreatedAt, entity.UpdatedAt,
		entity.CreatedBy, entity.Metadata,
	)

	if err != nil {
		// Verificar erros específicos do PostgreSQL para tratamento adequado
		if isPgErrorCode(err, uniqueViolationCode) {
			return repository.ErrDuplicateCode
		}
		
		log.Error().Err(err).
			Str("tenant_id", permission.TenantID().String()).
			Str("permission_id", permission.ID().String()).
			Str("permission_code", permission.Code()).
			Msg("Erro ao criar permissão no banco de dados")
		
		return fmt.Errorf("falha ao criar permissão: %w", err)
	}

	return nil
}

// FindByID recupera uma permissão pelo seu ID
func (r *PermissionRepository) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.FindByID", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_id", id.String()),
	))
	defer span.End()

	query := `
		SELECT id, tenant_id, code, name, description, resource_type, resource_id, 
		       action, is_active, is_system, created_at, updated_at, 
		       created_by, updated_by, deleted_at, deleted_by, metadata
		FROM iam.permissions
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	var entity permissionEntity
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
		&entity.ResourceType, &entity.ResourceID, &entity.Action,
		&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
		&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPermissionNotFound
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", id.String()).
			Msg("Erro ao buscar permissão por ID")
		
		return nil, fmt.Errorf("falha ao buscar permissão: %w", err)
	}

	return mapPermissionToDomain(entity)
}

// FindByCode recupera uma permissão pelo seu código
func (r *PermissionRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.FindByCode", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_code", code),
	))
	defer span.End()

	query := `
		SELECT id, tenant_id, code, name, description, resource_type, resource_id, 
		       action, is_active, is_system, created_at, updated_at, 
		       created_by, updated_by, deleted_at, deleted_by, metadata
		FROM iam.permissions
		WHERE code = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	var entity permissionEntity
	err := r.db.QueryRow(ctx, query, code, tenantID).Scan(
		&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
		&entity.ResourceType, &entity.ResourceID, &entity.Action,
		&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
		&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPermissionNotFound
		}
		
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_code", code).
			Msg("Erro ao buscar permissão por código")
		
		return nil, fmt.Errorf("falha ao buscar permissão por código: %w", err)
	}

	return mapPermissionToDomain(entity)
}

// Update atualiza os dados de uma permissão existente
func (r *PermissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	ctx, span := tracer.Start(ctx, "PermissionRepository.Update", trace.WithAttributes(
		attribute.String("tenant_id", permission.TenantID().String()),
		attribute.String("permission_id", permission.ID().String()),
	))
	defer span.End()

	// Mapear modelo de domínio para entidade de banco de dados
	entity := mapPermissionToEntity(permission)

	// Consulta SQL para atualizar permissão
	query := `
		UPDATE iam.permissions
		SET name = $1, 
		    description = $2, 
		    resource_type = $3,
		    resource_id = $4,
		    action = $5,
		    is_active = $6, 
		    updated_at = $7, 
		    updated_by = $8, 
		    metadata = $9
		WHERE id = $10 AND tenant_id = $11 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		entity.Name, entity.Description, entity.ResourceType, entity.ResourceID, entity.Action,
		entity.IsActive, entity.UpdatedAt, entity.UpdatedBy, entity.Metadata, 
		entity.ID, entity.TenantID,
	)

	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", permission.TenantID().String()).
			Str("permission_id", permission.ID().String()).
			Msg("Erro ao atualizar permissão no banco de dados")
		
		return fmt.Errorf("falha ao atualizar permissão: %w", err)
	}

	// Verificar se alguma linha foi afetada
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrPermissionNotFound
	}

	return nil
}// SoftDelete realiza uma exclusão lógica da permissão
func (r *PermissionRepository) SoftDelete(ctx context.Context, tenantID, id, deletedBy uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "PermissionRepository.SoftDelete", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_id", id.String()),
		attribute.String("deleted_by", deletedBy.String()),
	))
	defer span.End()

	// Consulta SQL para exclusão lógica
	query := `
		UPDATE iam.permissions
		SET deleted_at = $1, deleted_by = $2, is_active = false
		WHERE id = $3 AND tenant_id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		time.Now().UTC(), deletedBy, id, tenantID,
	)

	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", id.String()).
			Msg("Erro ao excluir logicamente permissão")
		
		return fmt.Errorf("falha ao excluir logicamente permissão: %w", err)
	}

	// Verificar se alguma linha foi afetada
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrPermissionNotFound
	}

	return nil
}

// HardDelete realiza a exclusão permanente da permissão
func (r *PermissionRepository) HardDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "PermissionRepository.HardDelete", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("permission_id", id.String()),
	))
	defer span.End()

	// Iniciar uma transação para garantir atomicidade na remoção de registros relacionados
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Remover associações com funções
	deleteRoleQuery := `DELETE FROM iam.role_permissions WHERE permission_id = $1 AND tenant_id = $2`
	_, err = tx.Exec(ctx, deleteRoleQuery, id, tenantID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", id.String()).
			Msg("Erro ao remover associações com funções durante exclusão permanente")
		
		return fmt.Errorf("falha ao remover associações com funções: %w", err)
	}

	// Remover a própria permissão
	deletePermQuery := `DELETE FROM iam.permissions WHERE id = $1 AND tenant_id = $2`
	result, err := tx.Exec(ctx, deletePermQuery, id, tenantID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("permission_id", id.String()).
			Msg("Erro ao excluir permanentemente permissão")
		
		return fmt.Errorf("falha ao excluir permanentemente permissão: %w", err)
	}

	// Verificar se alguma linha foi afetada
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrPermissionNotFound
	}

	// Confirmar a transação
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("falha ao confirmar exclusão permanente da permissão: %w", err)
	}

	return nil
}

// FindAll recupera permissões com base em filtros e paginação
func (r *PermissionRepository) FindAll(ctx context.Context, tenantID uuid.UUID, filter repository.PermissionFilter, pagination repository.Pagination) ([]*model.Permission, int64, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.FindAll", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
	))
	defer span.End()

	// Construir consulta base
	baseQuery := `
		FROM iam.permissions
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`

	// Construir cláusula de contagem
	countQuery := `SELECT COUNT(*) ` + baseQuery

	// Construir consulta para dados
	dataQuery := `
		SELECT id, tenant_id, code, name, description, resource_type, resource_id, 
		       action, is_active, is_system, created_at, updated_at, 
		       created_by, updated_by, deleted_at, deleted_by, metadata
	` + baseQuery

	// Parâmetros para consulta
	params := []interface{}{tenantID}
	paramIndex := 2 // O índice começa em 2 porque $1 já é tenantID

	// Aplicar filtros
	if filter.NameOrCodeContains != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d)", paramIndex, paramIndex)
		params = append(params, "%"+filter.NameOrCodeContains+"%")
		paramIndex++
	}

	if filter.ResourceType != "" {
		baseQuery += fmt.Sprintf(" AND resource_type = $%d", paramIndex)
		params = append(params, filter.ResourceType)
		paramIndex++
	}

	if filter.Action != "" {
		baseQuery += fmt.Sprintf(" AND action = $%d", paramIndex)
		params = append(params, filter.Action)
		paramIndex++
	}

	if filter.ResourceID != nil {
		baseQuery += fmt.Sprintf(" AND resource_id = $%d", paramIndex)
		params = append(params, *filter.ResourceID)
		paramIndex++
	}

	if filter.IsActive != nil {
		baseQuery += fmt.Sprintf(" AND is_active = $%d", paramIndex)
		params = append(params, *filter.IsActive)
		paramIndex++
	}

	if filter.IsSystem != nil {
		baseQuery += fmt.Sprintf(" AND is_system = $%d", paramIndex)
		params = append(params, *filter.IsSystem)
	}

	// Aplicar contagem total
	var total int64
	countQueryFull := countQuery + baseQuery
	err := r.db.QueryRow(ctx, countQueryFull, params...).Scan(&total)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao contar total de permissões")
		return nil, 0, fmt.Errorf("falha ao contar permissões: %w", err)
	}

	// Aplicar ordenação e paginação
	baseQuery += " ORDER BY code ASC LIMIT $" + fmt.Sprintf("%d", len(params)+1) + 
		" OFFSET $" + fmt.Sprintf("%d", len(params)+2)
	
	params = append(params, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
	
	// Consulta final para dados
	dataQueryFull := dataQuery + baseQuery

	// Executar consulta de dados
	rows, err := r.db.Query(ctx, dataQueryFull, params...)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao listar permissões")
		return nil, 0, fmt.Errorf("falha ao listar permissões: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	permissions := make([]*model.Permission, 0)
	for rows.Next() {
		var entity permissionEntity
		err := rows.Scan(
			&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
			&entity.ResourceType, &entity.ResourceID, &entity.Action,
			&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
		)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Msg("Erro ao escanear dados de permissão")
			return nil, 0, fmt.Errorf("falha ao escanear dados de permissão: %w", err)
		}

		permission, err := mapPermissionToDomain(entity)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("permission_id", entity.ID.String()).
				Msg("Erro ao mapear entidade para modelo de domínio")
			return nil, 0, fmt.Errorf("falha ao mapear entidade para modelo de domínio: %w", err)
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao iterar resultados de permissões")
		return nil, 0, fmt.Errorf("falha ao iterar resultados de permissões: %w", err)
	}

	return permissions, total, nil
}// FindByResourceType recupera permissões por tipo de recurso
func (r *PermissionRepository) FindByResourceType(ctx context.Context, tenantID uuid.UUID, resourceType string, pagination repository.Pagination) ([]*model.Permission, int64, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.FindByResourceType", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("resource_type", resourceType),
	))
	defer span.End()

	// Construir consulta base
	baseQuery := `
		FROM iam.permissions
		WHERE tenant_id = $1 AND resource_type = $2 AND deleted_at IS NULL
	`

	// Construir cláusula de contagem
	countQuery := `SELECT COUNT(*) ` + baseQuery

	// Construir consulta para dados
	dataQuery := `
		SELECT id, tenant_id, code, name, description, resource_type, resource_id, 
		       action, is_active, is_system, created_at, updated_at, 
		       created_by, updated_by, deleted_at, deleted_by, metadata
	` + baseQuery

	// Aplicar contagem total
	var total int64
	err := r.db.QueryRow(ctx, countQuery, tenantID, resourceType).Scan(&total)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("resource_type", resourceType).
			Msg("Erro ao contar permissões por tipo de recurso")
		return nil, 0, fmt.Errorf("falha ao contar permissões por tipo de recurso: %w", err)
	}

	// Aplicar ordenação e paginação
	dataQuery += ` ORDER BY code ASC LIMIT $3 OFFSET $4`

	// Executar consulta de dados
	rows, err := r.db.Query(ctx, dataQuery, tenantID, resourceType, 
		pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("resource_type", resourceType).
			Msg("Erro ao listar permissões por tipo de recurso")
		return nil, 0, fmt.Errorf("falha ao listar permissões por tipo de recurso: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	permissions := make([]*model.Permission, 0)
	for rows.Next() {
		var entity permissionEntity
		err := rows.Scan(
			&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
			&entity.ResourceType, &entity.ResourceID, &entity.Action,
			&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
		)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("resource_type", resourceType).
				Msg("Erro ao escanear dados de permissão")
			return nil, 0, fmt.Errorf("falha ao escanear dados de permissão: %w", err)
		}

		permission, err := mapPermissionToDomain(entity)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("permission_id", entity.ID.String()).
				Msg("Erro ao mapear entidade para modelo de domínio")
			return nil, 0, fmt.Errorf("falha ao mapear entidade para modelo de domínio: %w", err)
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("resource_type", resourceType).
			Msg("Erro ao iterar resultados de permissões")
		return nil, 0, fmt.Errorf("falha ao iterar resultados de permissões: %w", err)
	}

	return permissions, total, nil
}

// FindByIDs recupera múltiplas permissões pelos seus IDs
func (r *PermissionRepository) FindByIDs(ctx context.Context, tenantID uuid.UUID, ids []uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.FindByIDs", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.Int("permission_count", len(ids)),
	))
	defer span.End()

	if len(ids) == 0 {
		return []*model.Permission{}, nil
	}

	// Consulta para recuperar permissões por IDs
	query := `
		SELECT id, tenant_id, code, name, description, resource_type, resource_id, 
		       action, is_active, is_system, created_at, updated_at, 
		       created_by, updated_by, deleted_at, deleted_by, metadata
		FROM iam.permissions
		WHERE tenant_id = $1 AND id = ANY($2) AND deleted_at IS NULL
		ORDER BY code ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID, ids)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao buscar permissões por IDs")
		return nil, fmt.Errorf("falha ao buscar permissões por IDs: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	permissions := make([]*model.Permission, 0)
	for rows.Next() {
		var entity permissionEntity
		err := rows.Scan(
			&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
			&entity.ResourceType, &entity.ResourceID, &entity.Action,
			&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
		)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Msg("Erro ao escanear dados de permissão")
			return nil, fmt.Errorf("falha ao escanear dados de permissão: %w", err)
		}

		permission, err := mapPermissionToDomain(entity)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("permission_id", entity.ID.String()).
				Msg("Erro ao mapear entidade para modelo de domínio")
			return nil, fmt.Errorf("falha ao mapear entidade para modelo de domínio: %w", err)
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao iterar resultados de permissões")
		return nil, fmt.Errorf("falha ao iterar resultados de permissões: %w", err)
	}

	return permissions, nil
}

// GetUserDirectPermissions obtém permissões atribuídas diretamente a um usuário
func (r *PermissionRepository) GetUserDirectPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.GetUserDirectPermissions", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	))
	defer span.End()

	// Verificar se o usuário existe
	userExistsQuery := `
		SELECT EXISTS(
			SELECT 1 FROM iam.users 
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`
	var userExists bool
	err := r.db.QueryRow(ctx, userExistsQuery, userID, tenantID).Scan(&userExists)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar existência do usuário: %w", err)
	}
	if !userExists {
		return nil, repository.ErrUserNotFound
	}

	// Em uma implementação inicial, podemos não ter permissões diretas
	// Este é um placeholder para futura implementação de permissões diretas por usuário
	// Retornamos uma lista vazia por enquanto
	return []*model.Permission{}, nil
}

// GetUserRolePermissions obtém permissões atribuídas a um usuário através de funções
func (r *PermissionRepository) GetUserRolePermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.GetUserRolePermissions", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	))
	defer span.End()

	// Verificar se o usuário existe
	userExistsQuery := `
		SELECT EXISTS(
			SELECT 1 FROM iam.users 
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`
	var userExists bool
	err := r.db.QueryRow(ctx, userExistsQuery, userID, tenantID).Scan(&userExists)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar existência do usuário: %w", err)
	}
	if !userExists {
		return nil, repository.ErrUserNotFound
	}

	now := time.Now().UTC()

	// Consulta para buscar permissões através das funções atribuídas ao usuário
	query := `
		SELECT DISTINCT p.id, p.tenant_id, p.code, p.name, p.description, 
		               p.resource_type, p.resource_id, p.action,
		               p.is_active, p.is_system, p.created_at, p.updated_at, 
		               p.created_by, p.updated_by, p.deleted_at, p.deleted_by, p.metadata
		FROM iam.permissions p
		JOIN iam.role_permissions rp ON p.id = rp.permission_id AND p.tenant_id = rp.tenant_id
		JOIN iam.roles r ON rp.role_id = r.id AND rp.tenant_id = r.tenant_id
		JOIN iam.user_roles ur ON r.id = ur.role_id AND r.tenant_id = ur.tenant_id
		WHERE ur.user_id = $1 
		  AND ur.tenant_id = $2 
		  AND p.is_active = true 
		  AND r.is_active = true 
		  AND p.deleted_at IS NULL
		  AND r.deleted_at IS NULL
		  AND (ur.activates_at IS NULL OR ur.activates_at <= $3)
		  AND (ur.expires_at IS NULL OR ur.expires_at > $3)
		ORDER BY p.code ASC
	`

	rows, err := r.db.Query(ctx, query, userID, tenantID, now)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Erro ao buscar permissões de usuário via funções")
		return nil, fmt.Errorf("falha ao buscar permissões de usuário via funções: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	permissions := make([]*model.Permission, 0)
	for rows.Next() {
		var entity permissionEntity
		err := rows.Scan(
			&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
			&entity.ResourceType, &entity.ResourceID, &entity.Action,
			&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
		)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("user_id", userID.String()).
				Msg("Erro ao escanear dados de permissão")
			return nil, fmt.Errorf("falha ao escanear dados de permissão: %w", err)
		}

		permission, err := mapPermissionToDomain(entity)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("permission_id", entity.ID.String()).
				Msg("Erro ao mapear entidade para modelo de domínio")
			return nil, fmt.Errorf("falha ao mapear entidade para modelo de domínio: %w", err)
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Msg("Erro ao iterar resultados de permissões")
		return nil, fmt.Errorf("falha ao iterar resultados de permissões: %w", err)
	}

	return permissions, nil
}// GetUserAllPermissions obtém todas as permissões de um usuário (diretas + via funções)
func (r *PermissionRepository) GetUserAllPermissions(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.GetUserAllPermissions", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
	))
	defer span.End()

	// Obter permissões via funções
	rolePermissions, err := r.GetUserRolePermissions(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Obter permissões diretas
	directPermissions, err := r.GetUserDirectPermissions(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	// Combinar os dois conjuntos de permissões sem duplicatas
	permissionMap := make(map[uuid.UUID]*model.Permission)
	
	// Adicionar permissões de funções
	for _, perm := range rolePermissions {
		permissionMap[perm.ID()] = perm
	}
	
	// Adicionar permissões diretas (substituindo qualquer duplicata das permissões de funções)
	for _, perm := range directPermissions {
		permissionMap[perm.ID()] = perm
	}

	// Converter mapa de volta para slice
	allPermissions := make([]*model.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		allPermissions = append(allPermissions, perm)
	}

	// Ordenar resultado por código para consistência
	sort.Slice(allPermissions, func(i, j int) bool {
		return allPermissions[i].Code() < allPermissions[j].Code()
	})

	return allPermissions, nil
}

// CheckUserPermission verifica se um usuário tem uma permissão específica
func (r *PermissionRepository) CheckUserPermission(ctx context.Context, tenantID, userID uuid.UUID, permissionCode string) (bool, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.CheckUserPermission", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("permission_code", permissionCode),
	))
	defer span.End()

	// Verificar se o usuário existe
	userExistsQuery := `
		SELECT EXISTS(
			SELECT 1 FROM iam.users 
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`
	var userExists bool
	err := r.db.QueryRow(ctx, userExistsQuery, userID, tenantID).Scan(&userExists)
	if err != nil {
		return false, fmt.Errorf("falha ao verificar existência do usuário: %w", err)
	}
	if !userExists {
		return false, repository.ErrUserNotFound
	}

	now := time.Now().UTC()

	// Verificar permissões através de funções
	roleQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM iam.permissions p
			JOIN iam.role_permissions rp ON p.id = rp.permission_id AND p.tenant_id = rp.tenant_id
			JOIN iam.roles r ON rp.role_id = r.id AND rp.tenant_id = r.tenant_id
			JOIN iam.user_roles ur ON r.id = ur.role_id AND r.tenant_id = ur.tenant_id
			WHERE ur.user_id = $1 
			  AND ur.tenant_id = $2 
			  AND p.code = $3
			  AND p.is_active = true 
			  AND r.is_active = true 
			  AND p.deleted_at IS NULL
			  AND r.deleted_at IS NULL
			  AND (ur.activates_at IS NULL OR ur.activates_at <= $4)
			  AND (ur.expires_at IS NULL OR ur.expires_at > $4)
		)
	`

	var hasPermission bool
	err = r.db.QueryRow(ctx, roleQuery, userID, tenantID, permissionCode, now).Scan(&hasPermission)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Str("permission_code", permissionCode).
			Msg("Erro ao verificar permissão de usuário via funções")
		return false, fmt.Errorf("falha ao verificar permissão de usuário: %w", err)
	}

	// Se já encontrou a permissão via funções, retorna positivo
	if hasPermission {
		return true, nil
	}

	// Verificar permissões diretas do usuário (implementação futura)
	// Este é um placeholder para futura implementação

	return false, nil
}

// CheckUserResourcePermission verifica se um usuário tem uma permissão para um recurso específico
func (r *PermissionRepository) CheckUserResourcePermission(ctx context.Context, tenantID, userID uuid.UUID, resourceType, action string, resourceID *string) (bool, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.CheckUserResourcePermission", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("resource_type", resourceType),
		attribute.String("action", action),
		attribute.String("resource_id", stringValue(resourceID)),
	))
	defer span.End()

	// Verificar se o usuário existe
	userExistsQuery := `
		SELECT EXISTS(
			SELECT 1 FROM iam.users 
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`
	var userExists bool
	err := r.db.QueryRow(ctx, userExistsQuery, userID, tenantID).Scan(&userExists)
	if err != nil {
		return false, fmt.Errorf("falha ao verificar existência do usuário: %w", err)
	}
	if !userExists {
		return false, repository.ErrUserNotFound
	}

	now := time.Now().UTC()

	// Verificar permissões específicas para o recurso através de funções
	// Primeiro verificamos permissões específicas para o ID do recurso
	var hasPermission bool
	if resourceID != nil {
		specificQuery := `
			SELECT EXISTS (
				SELECT 1 
				FROM iam.permissions p
				JOIN iam.role_permissions rp ON p.id = rp.permission_id AND p.tenant_id = rp.tenant_id
				JOIN iam.roles r ON rp.role_id = r.id AND rp.tenant_id = r.tenant_id
				JOIN iam.user_roles ur ON r.id = ur.role_id AND r.tenant_id = ur.tenant_id
				WHERE ur.user_id = $1 
				  AND ur.tenant_id = $2 
				  AND p.resource_type = $3
				  AND p.action = $4
				  AND p.resource_id = $5
				  AND p.is_active = true 
				  AND r.is_active = true 
				  AND p.deleted_at IS NULL
				  AND r.deleted_at IS NULL
				  AND (ur.activates_at IS NULL OR ur.activates_at <= $6)
				  AND (ur.expires_at IS NULL OR ur.expires_at > $6)
			)
		`

		err = r.db.QueryRow(ctx, specificQuery, userID, tenantID, resourceType, action, resourceID, now).Scan(&hasPermission)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("user_id", userID.String()).
				Str("resource_type", resourceType).
				Str("action", action).
				Str("resource_id", *resourceID).
				Msg("Erro ao verificar permissão específica para recurso")
			return false, fmt.Errorf("falha ao verificar permissão específica para recurso: %w", err)
		}

		// Se já encontrou a permissão específica, retorna positivo
		if hasPermission {
			return true, nil
		}
	}

	// Se não encontrou permissão específica (ou se resourceID é nil),
	// verificamos permissões gerais para o tipo de recurso (com resource_id = NULL)
	generalQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM iam.permissions p
			JOIN iam.role_permissions rp ON p.id = rp.permission_id AND p.tenant_id = rp.tenant_id
			JOIN iam.roles r ON rp.role_id = r.id AND rp.tenant_id = r.tenant_id
			JOIN iam.user_roles ur ON r.id = ur.role_id AND r.tenant_id = ur.tenant_id
			WHERE ur.user_id = $1 
			  AND ur.tenant_id = $2 
			  AND p.resource_type = $3
			  AND p.action = $4
			  AND p.resource_id IS NULL
			  AND p.is_active = true 
			  AND r.is_active = true 
			  AND p.deleted_at IS NULL
			  AND r.deleted_at IS NULL
			  AND (ur.activates_at IS NULL OR ur.activates_at <= $5)
			  AND (ur.expires_at IS NULL OR ur.expires_at > $5)
		)
	`

	err = r.db.QueryRow(ctx, generalQuery, userID, tenantID, resourceType, action, now).Scan(&hasPermission)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("user_id", userID.String()).
			Str("resource_type", resourceType).
			Str("action", action).
			Msg("Erro ao verificar permissão geral para tipo de recurso")
		return false, fmt.Errorf("falha ao verificar permissão geral para tipo de recurso: %w", err)
	}

	return hasPermission, nil
}

// GetSystemPermissions obtém todas as permissões de sistema
func (r *PermissionRepository) GetSystemPermissions(ctx context.Context, tenantID uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "PermissionRepository.GetSystemPermissions", trace.WithAttributes(
		attribute.String("tenant_id", tenantID.String()),
	))
	defer span.End()

	query := `
		SELECT id, tenant_id, code, name, description, resource_type, resource_id, 
		       action, is_active, is_system, created_at, updated_at, 
		       created_by, updated_by, deleted_at, deleted_by, metadata
		FROM iam.permissions
		WHERE tenant_id = $1 AND is_system = true AND deleted_at IS NULL
		ORDER BY code ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao buscar permissões de sistema")
		return nil, fmt.Errorf("falha ao buscar permissões de sistema: %w", err)
	}
	defer rows.Close()

	// Processar resultados
	permissions := make([]*model.Permission, 0)
	for rows.Next() {
		var entity permissionEntity
		err := rows.Scan(
			&entity.ID, &entity.TenantID, &entity.Code, &entity.Name, &entity.Description, 
			&entity.ResourceType, &entity.ResourceID, &entity.Action,
			&entity.IsActive, &entity.IsSystem, &entity.CreatedAt, &entity.UpdatedAt,
			&entity.CreatedBy, &entity.UpdatedBy, &entity.DeletedAt, &entity.DeletedBy, &entity.Metadata,
		)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Msg("Erro ao escanear dados de permissão de sistema")
			return nil, fmt.Errorf("falha ao escanear dados de permissão de sistema: %w", err)
		}

		permission, err := mapPermissionToDomain(entity)
		if err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID.String()).
				Str("permission_id", entity.ID.String()).
				Msg("Erro ao mapear entidade para modelo de domínio")
			return nil, fmt.Errorf("falha ao mapear entidade para modelo de domínio: %w", err)
		}

		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Erro ao iterar resultados de permissões de sistema")
		return nil, fmt.Errorf("falha ao iterar resultados de permissões de sistema: %w", err)
	}

	return permissions, nil
}

// Função auxiliar para exibir valor de ponteiro string nos logs
func stringValue(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}