/**
 * INNOVABIZ IAM - Operações de Listagem e Consulta de Grupos em PostgreSQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de listagem, busca e consulta para grupos
 * usando PostgreSQL, seguindo a arquitetura multi-dimensional, multi-tenant 
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"go.opentelemetry.io/otel/attribute"
)

// List lista grupos com filtros, ordenação e paginação
func (r *GroupRepository) List(ctx context.Context, tenantID uuid.UUID, filter group.GroupFilter, page, pageSize int) (*group.GroupListResult, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)

	timer := r.metrics.Timer("repository.group.list.duration")
	defer timer.ObserveDuration()

	// Ajustar paginação (página inicia em 1)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Preparar consulta base
	baseQuery := `
		SELECT g.id, g.code, g.name, g.description, 
		       g.tenant_id, g.region_code, g.group_type, g.status, 
		       g.path, g.level, g.parent_group_id,
		       g.created_at, g.created_by, g.updated_at, g.updated_by,
		       g.metadata
		FROM iam_groups g
		WHERE g.tenant_id = $1 AND g.deleted_at IS NULL
	`

	countQuery := `
		SELECT COUNT(*)
		FROM iam_groups g
		WHERE g.tenant_id = $1 AND g.deleted_at IS NULL
	`

	// Inicializar argumentos com valores base
	args := []interface{}{tenantID}
	argIndex := 2 // Próximo índice de argumento

	// Construir cláusulas WHERE adicionais com base no filtro
	whereFilters := []string{}

	// Filtro por termo de busca (código, nome ou descrição)
	if filter.SearchTerm != "" {
		whereFilters = append(whereFilters, fmt.Sprintf("(g.code ILIKE $%d OR g.name ILIKE $%d OR g.description ILIKE $%d)", 
			argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.SearchTerm+"%")
		argIndex++
	}

	// Filtro por status
	if len(filter.Statuses) > 0 {
		placeholders := make([]string, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, status)
			argIndex++
		}
		whereFilters = append(whereFilters, fmt.Sprintf("g.status IN (%s)", 
			strings.Join(placeholders, ",")))
	}

	// Filtro por tipo de grupo
	if len(filter.Types) > 0 {
		placeholders := make([]string, 0, len(filter.Types))
		for _, groupType := range filter.Types {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, groupType)
			argIndex++
		}
		whereFilters = append(whereFilters, fmt.Sprintf("g.group_type IN (%s)", 
			strings.Join(placeholders, ",")))
	}

	// Filtro por região
	if len(filter.RegionCodes) > 0 {
		placeholders := make([]string, 0, len(filter.RegionCodes))
		for _, regionCode := range filter.RegionCodes {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
			args = append(args, regionCode)
			argIndex++
		}
		whereFilters = append(whereFilters, fmt.Sprintf("g.region_code IN (%s)", 
			strings.Join(placeholders, ",")))
	}

	// Filtro por grupo pai
	if filter.ParentGroupID != nil {
		if *filter.ParentGroupID == uuid.Nil {
			// Grupos raiz (sem pai)
			whereFilters = append(whereFilters, "g.parent_group_id IS NULL")
		} else {
			// Subgrupos de um grupo específico
			whereFilters = append(whereFilters, fmt.Sprintf("g.parent_group_id = $%d", argIndex))
			args = append(args, *filter.ParentGroupID)
			argIndex++
		}
	}

	// Filtro por usuário (grupos dos quais o usuário é membro)
	if filter.UserID != nil {
		whereFilters = append(whereFilters, fmt.Sprintf(`
			g.id IN (
				SELECT gm.group_id 
				FROM iam_group_members gm 
				WHERE gm.user_id = $%d AND gm.tenant_id = $1
			)
		`, argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	// Filtro por papel (grupos que possuem determinado papel)
	if filter.RoleID != nil {
		whereFilters = append(whereFilters, fmt.Sprintf(`
			g.id IN (
				SELECT gr.group_id 
				FROM iam_group_roles gr 
				WHERE gr.role_id = $%d AND gr.tenant_id = $1
			)
		`, argIndex))
		args = append(args, *filter.RoleID)
		argIndex++
	}

	// Filtro por nível hierárquico
	if filter.Level > 0 {
		whereFilters = append(whereFilters, fmt.Sprintf("g.level = $%d", argIndex))
		args = append(args, filter.Level)
		argIndex++
	}

	// Filtro por data de criação (intervalo)
	if !filter.CreatedAfter.IsZero() {
		whereFilters = append(whereFilters, fmt.Sprintf("g.created_at >= $%d", argIndex))
		args = append(args, filter.CreatedAfter)
		argIndex++
	}

	if !filter.CreatedBefore.IsZero() {
		whereFilters = append(whereFilters, fmt.Sprintf("g.created_at <= $%d", argIndex))
		args = append(args, filter.CreatedBefore)
		argIndex++
	}

	// Adicionar cláusulas WHERE adicionais, se existirem
	if len(whereFilters) > 0 {
		whereClause := " AND " + strings.Join(whereFilters, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Ordenação
	sortBy := "g.name"
	sortDirection := "ASC"

	// Validar campo de ordenação
	validFields := map[string]string{
		"name":        "g.name",
		"code":        "g.code",
		"status":      "g.status",
		"level":       "g.level",
		"createdAt":   "g.created_at",
		"updatedAt":   "g.updated_at",
		"groupType":   "g.group_type",
		"regionCode":  "g.region_code",
	}
	
	if validField, ok := validFields[filter.SortBy]; ok && filter.SortBy != "" {
		sortBy = validField
	}
	
	if strings.ToUpper(filter.SortDirection) == "DESC" {
		sortDirection = "DESC"
	}

	// Finalizar query com ordenação e paginação
	baseQuery += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", 
		sortBy, sortDirection, argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Extrair transação do contexto
	tx := persistence.GetTxFromContext(ctx)

	// Executar consulta de contagem
	var totalCount int
	if tx != nil {
		err := tx.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar grupos", logging.Fields{
				"error":    err.Error(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.list.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar grupos: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar grupos", logging.Fields{
				"error":    err.Error(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.list.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar grupos: %w", err)
		}
	}

	// Calcular total de páginas
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// Se não houver registros, retornar resultado vazio
	if totalCount == 0 {
		return &group.GroupListResult{
			Groups:     []*group.Group{},
			TotalCount: 0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}, nil
	}

	// Executar consulta principal
	var rows *sql.Rows
	var queryErr error

	if tx != nil {
		rows, queryErr = tx.QueryContext(ctx, baseQuery, args...)
	} else {
		rows, queryErr = r.db.QueryContext(ctx, baseQuery, args...)
	}

	if queryErr != nil {
		r.logger.Error(ctx, "Erro ao listar grupos", logging.Fields{
			"error":    queryErr.Error(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.list.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar grupos: %w", queryErr)
	}
	defer rows.Close()

	// Processar resultados
	groups := []*group.Group{}

	for rows.Next() {
		var g group.Group
		var metadata sql.NullString
		var parentGroupID sql.NullString
		var description, regionCode, groupType sql.NullString
		var updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&g.ID,
			&g.Code,
			&g.Name,
			&description,
			&g.TenantID,
			&regionCode,
			&groupType,
			&g.Status,
			&g.Path,
			&g.Level,
			&parentGroupID,
			&g.CreatedAt,
			&g.CreatedBy,
			&updatedAt,
			&updatedBy,
			&metadata,
		)

		if err != nil {
			r.logger.Error(ctx, "Erro ao processar registro de grupo", logging.Fields{
				"error":    err.Error(),
				"tenantId": tenantID.String(),
			})
			continue
		}

		// Processar campos opcionais
		if description.Valid {
			g.Description = description.String
		}
		
		if regionCode.Valid {
			g.RegionCode = regionCode.String
		}
		
		if groupType.Valid {
			g.GroupType = groupType.String
		}

		if parentGroupID.Valid {
			parentUUID, err := uuid.Parse(parentGroupID.String)
			if err == nil {
				g.ParentGroupID = &parentUUID
			}
		}

		if updatedAt.Valid {
			g.UpdatedAt = updatedAt.Time
		}

		if updatedBy.Valid {
			updatedByUUID, err := uuid.Parse(updatedBy.String)
			if err == nil {
				g.UpdatedBy = &updatedByUUID
			}
		}

		// Processar metadados (JSON)
		if metadata.Valid && metadata.String != "" {
			g.Metadata = make(map[string]interface{})
			if err := json.Unmarshal([]byte(metadata.String), &g.Metadata); err != nil {
				r.logger.Warn(ctx, "Erro ao processar metadados do grupo", logging.Fields{
					"error":   err.Error(),
					"groupId": g.ID.String(),
				})
			}
		}

		groups = append(groups, &g)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(ctx, "Erro ao iterar resultados de grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.list.error").Inc(1)
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	r.metrics.Counter("repository.group.list.success").Inc(1)

	// Retornar resultado
	return &group.GroupListResult{
		Groups:     groups,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// FindGroupsByUserID busca grupos aos quais um usuário pertence
func (r *GroupRepository) FindGroupsByUserID(ctx context.Context, userID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*group.GroupListResult, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.FindGroupsByUserID")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)

	timer := r.metrics.Timer("repository.group.findGroupsByUserID.duration")
	defer timer.ObserveDuration()

	// Ajustar paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var baseQuery string
	var countQuery string
	var args []interface{}

	if recursive {
		// Busca recursiva inclui grupos ancestrais
		baseQuery = `
			WITH RECURSIVE user_groups AS (
				-- Grupos diretos do usuário
				SELECT g.id, g.code, g.name, g.description, 
					  g.tenant_id, g.region_code, g.group_type, g.status, 
					  g.path, g.level, g.parent_group_id,
					  g.created_at, g.created_by, g.updated_at, g.updated_by,
					  g.metadata, 1 AS membership_level
				FROM iam_groups g
				JOIN iam_group_members gm ON g.id = gm.group_id
				WHERE gm.user_id = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
				
				UNION
				
				-- Grupos ancestrais (hierarquia para cima)
				SELECT parent.id, parent.code, parent.name, parent.description, 
					  parent.tenant_id, parent.region_code, parent.group_type, parent.status, 
					  parent.path, parent.level, parent.parent_group_id,
					  parent.created_at, parent.created_by, parent.updated_at, parent.updated_by,
					  parent.metadata, ug.membership_level + 1
				FROM user_groups ug
				JOIN iam_groups parent ON parent.id = ug.parent_group_id
				WHERE parent.tenant_id = $2 AND parent.deleted_at IS NULL
			)
			SELECT id, code, name, description, tenant_id, region_code, group_type, status, 
				   path, level, parent_group_id, created_at, created_by, updated_at, updated_by,
				   metadata, membership_level
			FROM user_groups
			ORDER BY membership_level ASC, name ASC
			LIMIT $3 OFFSET $4
		`
		
		countQuery = `
			WITH RECURSIVE user_groups AS (
				-- Grupos diretos do usuário
				SELECT g.id, 1 AS membership_level
				FROM iam_groups g
				JOIN iam_group_members gm ON g.id = gm.group_id
				WHERE gm.user_id = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
				
				UNION
				
				-- Grupos ancestrais (hierarquia para cima)
				SELECT parent.id, ug.membership_level + 1
				FROM user_groups ug
				JOIN iam_groups parent ON parent.id = ug.parent_group_id
				WHERE parent.tenant_id = $2 AND parent.deleted_at IS NULL
			)
			SELECT COUNT(*)
			FROM user_groups
		`
	} else {
		// Busca direta apenas para grupos aos quais o usuário pertence diretamente
		baseQuery = `
			SELECT g.id, g.code, g.name, g.description, 
				  g.tenant_id, g.region_code, g.group_type, g.status, 
				  g.path, g.level, g.parent_group_id,
				  g.created_at, g.created_by, g.updated_at, g.updated_by,
				  g.metadata, 1 AS membership_level
			FROM iam_groups g
			JOIN iam_group_members gm ON g.id = gm.group_id
			WHERE gm.user_id = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
			ORDER BY g.name ASC
			LIMIT $3 OFFSET $4
		`
		
		countQuery = `
			SELECT COUNT(*)
			FROM iam_groups g
			JOIN iam_group_members gm ON g.id = gm.group_id
			WHERE gm.user_id = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
		`
	}

	args = []interface{}{userID, tenantID, pageSize, offset}

	// Extrair transação do contexto
	tx := persistence.GetTxFromContext(ctx)

	// Executar consulta de contagem
	var totalCount int
	if tx != nil {
		err := tx.QueryRowContext(ctx, countQuery, userID, tenantID).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar grupos do usuário", logging.Fields{
				"error":    err.Error(),
				"userId":   userID.String(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.findGroupsByUserID.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar grupos do usuário: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery, userID, tenantID).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar grupos do usuário", logging.Fields{
				"error":    err.Error(),
				"userId":   userID.String(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.findGroupsByUserID.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar grupos do usuário: %w", err)
		}
	}

	// Calcular total de páginas
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// Se não houver registros, retornar resultado vazio
	if totalCount == 0 {
		return &group.GroupListResult{
			Groups:     []*group.Group{},
			TotalCount: 0,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: 0,
		}, nil
	}

	// Executar consulta principal
	var rows *sql.Rows
	var queryErr error

	if tx != nil {
		rows, queryErr = tx.QueryContext(ctx, baseQuery, args...)
	} else {
		rows, queryErr = r.db.QueryContext(ctx, baseQuery, args...)
	}

	if queryErr != nil {
		r.logger.Error(ctx, "Erro ao listar grupos do usuário", logging.Fields{
			"error":    queryErr.Error(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.findGroupsByUserID.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar grupos do usuário: %w", queryErr)
	}
	defer rows.Close()

	// Processar resultados
	groups := []*group.Group{}

	for rows.Next() {
		var g group.Group
		var metadata sql.NullString
		var membershipLevel int
		var parentGroupID sql.NullString
		var description, regionCode, groupType sql.NullString
		var updatedAt sql.NullTime
		var updatedBy sql.NullString

		err := rows.Scan(
			&g.ID,
			&g.Code,
			&g.Name,
			&description,
			&g.TenantID,
			&regionCode,
			&groupType,
			&g.Status,
			&g.Path,
			&g.Level,
			&parentGroupID,
			&g.CreatedAt,
			&g.CreatedBy,
			&updatedAt,
			&updatedBy,
			&metadata,
			&membershipLevel,
		)

		if err != nil {
			r.logger.Error(ctx, "Erro ao processar registro de grupo", logging.Fields{
				"error":    err.Error(),
				"userId":   userID.String(),
				"tenantId": tenantID.String(),
			})
			continue
		}

		// Processar campos opcionais
		if description.Valid {
			g.Description = description.String
		}
		
		if regionCode.Valid {
			g.RegionCode = regionCode.String
		}
		
		if groupType.Valid {
			g.GroupType = groupType.String
		}

		if parentGroupID.Valid {
			parentUUID, err := uuid.Parse(parentGroupID.String)
			if err == nil {
				g.ParentGroupID = &parentUUID
			}
		}

		if updatedAt.Valid {
			g.UpdatedAt = updatedAt.Time
		}

		if updatedBy.Valid {
			updatedByUUID, err := uuid.Parse(updatedBy.String)
			if err == nil {
				g.UpdatedBy = &updatedByUUID
			}
		}

		// Processar metadados (JSON)
		if metadata.Valid && metadata.String != "" {
			g.Metadata = make(map[string]interface{})
			if err := json.Unmarshal([]byte(metadata.String), &g.Metadata); err != nil {
				r.logger.Warn(ctx, "Erro ao processar metadados do grupo", logging.Fields{
					"error":   err.Error(),
					"groupId": g.ID.String(),
				})
			}
		}

		groups = append(groups, &g)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(ctx, "Erro ao iterar resultados de grupos", logging.Fields{
			"error":    err.Error(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.findGroupsByUserID.error").Inc(1)
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	r.metrics.Counter("repository.group.findGroupsByUserID.success").Inc(1)

	// Retornar resultado
	return &group.GroupListResult{
		Groups:     groups,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}