package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/domain/model"
)

// GetByID recupera uma função pelo seu ID
func (r *RoleRepository) GetByID(ctx context.Context, tenantID, roleID uuid.UUID) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	query := `
		SELECT 
			id, tenant_id, code, name, description, 
			type, is_system, is_active, metadata, 
			created_at, created_by, updated_at, updated_by,
			deleted_at, deleted_by, version
		FROM roles 
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	var role *model.Role
	var err error

	err = r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		role, err = r.scanRole(ctx, tx.QueryRow(ctx, query, roleID, tenantID))
		if err != nil {
			if err == pgx.ErrNoRows {
				return model.NewRoleNotFoundError(roleID)
			}
			return fmt.Errorf("erro ao consultar função por ID: %w", err)
		}
		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return role, nil
}

// GetByCode recupera uma função pelo seu código
func (r *RoleRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetByCode")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.code", code),
		attribute.String("tenant.id", tenantID.String()),
	)

	query := `
		SELECT 
			id, tenant_id, code, name, description, 
			type, is_system, is_active, metadata, 
			created_at, created_by, updated_at, updated_by,
			deleted_at, deleted_by, version
		FROM roles 
		WHERE code = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	var role *model.Role
	var err error

	err = r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		role, err = r.scanRole(ctx, tx.QueryRow(ctx, query, code, tenantID))
		if err != nil {
			if err == pgx.ErrNoRows {
				return model.NewRoleCodeNotFoundError(code, tenantID)
			}
			return fmt.Errorf("erro ao consultar função por código: %w", err)
		}
		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return role, nil
}

// List recupera funções com filtros e paginação
func (r *RoleRepository) List(ctx context.Context, tenantID uuid.UUID, filter model.RoleFilter, pagination model.Pagination) ([]*model.Role, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	// Construir consulta base
	baseQuery := `
		FROM roles r
		WHERE r.tenant_id = $1 AND r.deleted_at IS NULL
	`

	// Parâmetros da consulta
	params := []interface{}{tenantID}
	paramIndex := 2 // Próximo índice de parâmetro
	
	// Adicionar condições de filtro
	filterConditions := []string{}

	// Filtro por nome ou código
	if filter.NameOrCodeContains != "" {
		filterConditions = append(filterConditions, 
			fmt.Sprintf("(r.name ILIKE $%d OR r.code ILIKE $%d)", paramIndex, paramIndex))
		params = append(params, "%"+filter.NameOrCodeContains+"%")
		paramIndex++
	}
	
	// Filtro por tipos
	if len(filter.Types) > 0 {
		typesPlaceholders := make([]string, len(filter.Types))
		for i, _ := range filter.Types {
			typesPlaceholders[i] = fmt.Sprintf("$%d", paramIndex)
			params = append(params, filter.Types[i])
			paramIndex++
		}
		filterConditions = append(filterConditions, 
			fmt.Sprintf("r.type IN (%s)", strings.Join(typesPlaceholders, ", ")))
	}

	// Filtro por status ativo/inativo
	if filter.IsActive != nil {
		filterConditions = append(filterConditions, 
			fmt.Sprintf("r.is_active = $%d", paramIndex))
		params = append(params, *filter.IsActive)
		paramIndex++
	}

	// Filtro por função de sistema ou customizada
	if filter.IsSystem != nil {
		filterConditions = append(filterConditions, 
			fmt.Sprintf("r.is_system = $%d", paramIndex))
		params = append(params, *filter.IsSystem)
		paramIndex++
	}

	// Adicionar condições à consulta
	if len(filterConditions) > 0 {
		baseQuery += " AND " + strings.Join(filterConditions, " AND ")
	}

	// Consulta de contagem
	countQuery := "SELECT COUNT(*) " + baseQuery

	// Consulta principal com paginação
	query := `
		SELECT 
			r.id, r.tenant_id, r.code, r.name, r.description, 
			r.type, r.is_system, r.is_active, r.metadata, 
			r.created_at, r.created_by, r.updated_at, r.updated_by,
			r.deleted_at, r.deleted_by, r.version
	` + baseQuery + `
		ORDER BY r.name ASC
		LIMIT $` + fmt.Sprintf("%d", paramIndex) + ` OFFSET $` + fmt.Sprintf("%d", paramIndex+1)
	
	// Parâmetros de paginação
	params = append(params, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)

	var roles []*model.Role
	var totalCount int64

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Executar consulta de contagem
		err := tx.QueryRow(ctx, countQuery, params[:len(params)-2]...).Scan(&totalCount)
		if err != nil {
			return fmt.Errorf("erro ao contar funções: %w", err)
		}

		// Executar consulta principal
		rows, err := tx.Query(ctx, query, params...)
		if err != nil {
			return fmt.Errorf("erro ao listar funções: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, err := r.scanRoleFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função: %w", err)
			}
			roles = append(roles, role)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar funções: %w", rows.Err())
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, 0, err
	}

	return roles, totalCount, nil
}

// Delete marca uma função como excluída ou a remove permanentemente
func (r *RoleRepository) Delete(ctx context.Context, tenantID, roleID, deletedBy uuid.UUID, hardDelete bool) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("hard_delete", hardDelete),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se existem usuários atribuídos à função
		var userCount int64
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM user_roles 
			WHERE role_id = $1 AND tenant_id = $2
		`, roleID, tenantID).Scan(&userCount)
		
		if err != nil {
			return fmt.Errorf("erro ao verificar usuários associados à função: %w", err)
		}
		
		if userCount > 0 {
			return model.NewRoleHasUsersError(roleID)
		}
		
		// Verificar se existem funções filhas
		var childCount int64
		err = tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM role_hierarchy 
			WHERE parent_role_id = $1 AND tenant_id = $2
		`, roleID, tenantID).Scan(&childCount)
		
		if err != nil {
			return fmt.Errorf("erro ao verificar funções filhas: %w", err)
		}
		
		if childCount > 0 {
			return model.NewRoleHasChildrenError(roleID)
		}

		// Verificar se é uma função de sistema
		var isSystem bool
		err = tx.QueryRow(ctx, `
			SELECT is_system FROM roles
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		`, roleID, tenantID).Scan(&isSystem)
		
		if err != nil {
			if err == pgx.ErrNoRows {
				return model.NewRoleNotFoundError(roleID)
			}
			return fmt.Errorf("erro ao verificar se é função de sistema: %w", err)
		}
		
		if isSystem {
			return model.NewCannotDeleteSystemRoleError(roleID)
		}

		if hardDelete {
			// Hard delete - remover registros relacionados
			
			// Remover permissões da função
			_, err = tx.Exec(ctx, `
				DELETE FROM role_permissions
				WHERE role_id = $1 AND tenant_id = $2
			`, roleID, tenantID)
			
			if err != nil {
				return fmt.Errorf("erro ao remover permissões da função: %w", err)
			}
			
			// Remover da hierarquia como filho
			_, err = tx.Exec(ctx, `
				DELETE FROM role_hierarchy
				WHERE child_role_id = $1 AND tenant_id = $2
			`, roleID, tenantID)
			
			if err != nil {
				return fmt.Errorf("erro ao remover função da hierarquia como filho: %w", err)
			}
			
			// Remover a função
			_, err = tx.Exec(ctx, `
				DELETE FROM roles
				WHERE id = $1 AND tenant_id = $2
			`, roleID, tenantID)
			
		} else {
			// Soft delete - marcar como excluído
			_, err = tx.Exec(ctx, `
				UPDATE roles
				SET deleted_at = NOW(), deleted_by = $3
				WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			`, roleID, tenantID, deletedBy)
		}
		
		if err != nil {
			return fmt.Errorf("erro ao excluir função: %w", err)
		}

		return nil
	})
}

// scanRole lê uma linha de resultado e constrói um objeto Role do domínio
func (r *RoleRepository) scanRole(ctx context.Context, row pgx.Row) (*model.Role, error) {
	var (
		id          uuid.UUID
		tenantID    uuid.UUID
		code        string
		name        string
		description string
		roleType    string
		isSystem    bool
		isActive    bool
		metadataJSON []byte
		createdAt   time.Time
		createdBy   uuid.UUID
		updatedAt   time.Time
		updatedBy   uuid.UUID
		deletedAt   sql.NullTime
		deletedBy   uuid.NullString
		version     int
	)

	err := row.Scan(
		&id, &tenantID, &code, &name, &description,
		&roleType, &isSystem, &isActive, &metadataJSON,
		&createdAt, &createdBy, &updatedAt, &updatedBy,
		&deletedAt, &deletedBy, &version,
	)
	
	if err != nil {
		return nil, err
	}

	// Converter JSON de metadados para map
	metadata := make(map[string]interface{})
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			log.Ctx(ctx).Warn().Err(err).
				Str("role_id", id.String()).
				Msg("Erro ao deserializar metadados da função")
			// Continuar com metadata vazio em caso de erro
		}
	}

	// Construir e configurar o objeto do domínio
	role, err := model.ReconstructRole(
		id, tenantID, code, name, description, 
		roleType, isSystem, isActive, metadata,
		createdAt, createdBy, updatedAt, updatedBy,
		version,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao reconstruir objeto Role: %w", err)
	}

	// Configurar campos de exclusão se aplicável
	if deletedAt.Valid {
		var deletedByUUID uuid.UUID
		if deletedBy.Valid {
			var parseErr error
			deletedByUUID, parseErr = uuid.Parse(deletedBy.String)
			if parseErr != nil {
				log.Ctx(ctx).Warn().Err(parseErr).
					Str("role_id", id.String()).
					Msg("UUID de deletedBy inválido")
				// Continuar com UUID zero em caso de erro
			}
		}
		role.MarkAsDeleted(deletedAt.Time, deletedByUUID)
	}

	return role, nil
}

// scanRoleFromRows é uma função auxiliar para ler Role a partir de resultados de consulta
func (r *RoleRepository) scanRoleFromRows(ctx context.Context, rows pgx.Rows) (*model.Role, error) {
	var (
		id          uuid.UUID
		tenantID    uuid.UUID
		code        string
		name        string
		description string
		roleType    string
		isSystem    bool
		isActive    bool
		metadataJSON []byte
		createdAt   time.Time
		createdBy   uuid.UUID
		updatedAt   time.Time
		updatedBy   uuid.UUID
		deletedAt   sql.NullTime
		deletedBy   uuid.NullString
		version     int
	)

	err := rows.Scan(
		&id, &tenantID, &code, &name, &description,
		&roleType, &isSystem, &isActive, &metadataJSON,
		&createdAt, &createdBy, &updatedAt, &updatedBy,
		&deletedAt, &deletedBy, &version,
	)
	
	if err != nil {
		return nil, err
	}

	// Converter JSON de metadados para map
	metadata := make(map[string]interface{})
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			log.Ctx(ctx).Warn().Err(err).
				Str("role_id", id.String()).
				Msg("Erro ao deserializar metadados da função")
			// Continuar com metadata vazio em caso de erro
		}
	}

	// Construir e configurar o objeto do domínio
	role, err := model.ReconstructRole(
		id, tenantID, code, name, description, 
		roleType, isSystem, isActive, metadata,
		createdAt, createdBy, updatedAt, updatedBy,
		version,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao reconstruir objeto Role: %w", err)
	}

	// Configurar campos de exclusão se aplicável
	if deletedAt.Valid {
		var deletedByUUID uuid.UUID
		if deletedBy.Valid {
			var parseErr error
			deletedByUUID, parseErr = uuid.Parse(deletedBy.String)
			if parseErr != nil {
				log.Ctx(ctx).Warn().Err(parseErr).
					Str("role_id", id.String()).
					Msg("UUID de deletedBy inválido")
				// Continuar com UUID zero em caso de erro
			}
		}
		role.MarkAsDeleted(deletedAt.Time, deletedByUUID)
	}

	return role, nil
}