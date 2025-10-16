/**
 * INNOVABIZ IAM - Operações de Hierarquia de Grupos em PostgreSQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações relacionadas à hierarquia de grupos usando PostgreSQL,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"go.opentelemetry.io/otel/attribute"
)

// GetParentGroup obtém o grupo pai de um grupo
func (r *GroupRepository) GetParentGroup(ctx context.Context, groupID, tenantID uuid.UUID) (*group.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.GetParentGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := r.metrics.Timer("repository.group.getParentGroup.duration")
	defer timer.ObserveDuration()

	query := `
		SELECT parent.id, parent.code, parent.name, parent.description, 
			   parent.tenant_id, parent.region_code, parent.group_type, parent.status, 
			   parent.path, parent.level, parent.parent_group_id,
			   parent.created_at, parent.created_by, parent.updated_at, parent.updated_by,
			   parent.metadata
		FROM iam_groups g
		JOIN iam_groups parent ON g.parent_group_id = parent.id
		WHERE g.id = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL 
			  AND parent.deleted_at IS NULL
	`

	tx := persistence.GetTxFromContext(ctx)

	var row *sqlx.Row
	if tx != nil {
		row = tx.QueryRowxContext(ctx, query, groupID, tenantID)
	} else {
		row = r.db.QueryRowxContext(ctx, query, groupID, tenantID)
	}

	var parent group.Group
	var metadata sql.NullString

	err := row.Scan(
		&parent.ID,
		&parent.Code,
		&parent.Name,
		&parent.Description,
		&parent.TenantID,
		&parent.RegionCode,
		&parent.GroupType,
		&parent.Status,
		&parent.Path,
		&parent.Level,
		&parent.ParentGroupID,
		&parent.CreatedAt,
		&parent.CreatedBy,
		&parent.UpdatedAt,
		&parent.UpdatedBy,
		&metadata,
	)

	if err != nil {
		r.metrics.Counter("repository.group.getParentGroup.error").Inc(1)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrParentGroupNotFound
		}

		r.logger.Error(ctx, "Erro ao buscar grupo pai", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		return nil, fmt.Errorf("erro ao buscar grupo pai: %w", err)
	}

	// Processar metadados (JSON)
	if metadata.Valid && metadata.String != "" {
		parent.Metadata = make(map[string]interface{})
		if err := json.Unmarshal([]byte(metadata.String), &parent.Metadata); err != nil {
			r.logger.Warn(ctx, "Erro ao processar metadados do grupo pai", logging.Fields{
				"error":   err.Error(),
				"groupId": parent.ID.String(),
			})
		}
	}

	r.metrics.Counter("repository.group.getParentGroup.success").Inc(1)
	return &parent, nil
}

// GetChildGroups obtém os grupos filhos de um grupo
func (r *GroupRepository) GetChildGroups(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*group.GroupListResult, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.GetChildGroups")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)

	timer := r.metrics.Timer("repository.group.getChildGroups.duration")
	defer timer.ObserveDuration()

	// Ajustar paginação (página inicia em 1)
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
		// Busca recursiva inclui todos os descendentes
		baseQuery = `
			WITH RECURSIVE group_hierarchy AS (
				-- Grupos filhos diretos
				SELECT child.id, child.code, child.name, child.description, 
					  child.tenant_id, child.region_code, child.group_type, child.status, 
					  child.path, child.level, child.parent_group_id,
					  child.created_at, child.created_by, child.updated_at, child.updated_by,
					  child.metadata, 1 AS depth
				FROM iam_groups parent
				JOIN iam_groups child ON child.parent_group_id = parent.id
				WHERE parent.id = $1 AND parent.tenant_id = $2 
					  AND parent.deleted_at IS NULL AND child.deleted_at IS NULL
				
				UNION ALL
				
				-- Descendentes recursivos
				SELECT child.id, child.code, child.name, child.description, 
					  child.tenant_id, child.region_code, child.group_type, child.status, 
					  child.path, child.level, child.parent_group_id,
					  child.created_at, child.created_by, child.updated_at, child.updated_by,
					  child.metadata, gh.depth + 1
				FROM group_hierarchy gh
				JOIN iam_groups child ON child.parent_group_id = gh.id
				WHERE child.tenant_id = $2 AND child.deleted_at IS NULL
			)
			SELECT id, code, name, description, tenant_id, region_code, group_type, status, 
				   path, level, parent_group_id, created_at, created_by, updated_at, updated_by,
				   metadata, depth
			FROM group_hierarchy
			ORDER BY depth ASC, name ASC
			LIMIT $3 OFFSET $4
		`
		
		countQuery = `
			WITH RECURSIVE group_hierarchy AS (
				-- Grupos filhos diretos
				SELECT child.id, 1 AS depth
				FROM iam_groups parent
				JOIN iam_groups child ON child.parent_group_id = parent.id
				WHERE parent.id = $1 AND parent.tenant_id = $2 
					  AND parent.deleted_at IS NULL AND child.deleted_at IS NULL
				
				UNION ALL
				
				-- Descendentes recursivos
				SELECT child.id, gh.depth + 1
				FROM group_hierarchy gh
				JOIN iam_groups child ON child.parent_group_id = gh.id
				WHERE child.tenant_id = $2 AND child.deleted_at IS NULL
			)
			SELECT COUNT(*)
			FROM group_hierarchy
		`
		
		args = []interface{}{groupID, tenantID, pageSize, offset}
	} else {
		// Busca direta apenas para filhos imediatos
		baseQuery = `
			SELECT child.id, child.code, child.name, child.description, 
				  child.tenant_id, child.region_code, child.group_type, child.status, 
				  child.path, child.level, child.parent_group_id,
				  child.created_at, child.created_by, child.updated_at, child.updated_by,
				  child.metadata, 1 AS depth
			FROM iam_groups parent
			JOIN iam_groups child ON child.parent_group_id = parent.id
			WHERE parent.id = $1 AND parent.tenant_id = $2 
				  AND parent.deleted_at IS NULL AND child.deleted_at IS NULL
			ORDER BY child.name ASC
			LIMIT $3 OFFSET $4
		`
		
		countQuery = `
			SELECT COUNT(*)
			FROM iam_groups parent
			JOIN iam_groups child ON child.parent_group_id = parent.id
			WHERE parent.id = $1 AND parent.tenant_id = $2 
				  AND parent.deleted_at IS NULL AND child.deleted_at IS NULL
		`
		
		args = []interface{}{groupID, tenantID, pageSize, offset}
	}

	tx := persistence.GetTxFromContext(ctx)

	// Executar consulta de contagem
	var totalCount int
	if tx != nil {
		err := tx.QueryRowContext(ctx, countQuery, groupID, tenantID).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar grupos filhos", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.getChildGroups.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar grupos filhos: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery, groupID, tenantID).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar grupos filhos", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.getChildGroups.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar grupos filhos: %w", err)
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
		r.logger.Error(ctx, "Erro ao listar grupos filhos", logging.Fields{
			"error":    queryErr.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.getChildGroups.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar grupos filhos: %w", queryErr)
	}
	defer rows.Close()

	// Processar resultados
	groups := []*group.Group{}

	for rows.Next() {
		var g group.Group
		var metadata sql.NullString
		var depth int

		err := rows.Scan(
			&g.ID,
			&g.Code,
			&g.Name,
			&g.Description,
			&g.TenantID,
			&g.RegionCode,
			&g.GroupType,
			&g.Status,
			&g.Path,
			&g.Level,
			&g.ParentGroupID,
			&g.CreatedAt,
			&g.CreatedBy,
			&g.UpdatedAt,
			&g.UpdatedBy,
			&metadata,
			&depth,
		)

		if err != nil {
			r.logger.Error(ctx, "Erro ao processar registro de grupo filho", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			continue
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
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.getChildGroups.error").Inc(1)
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	r.metrics.Counter("repository.group.getChildGroups.success").Inc(1)

	// Retornar resultado
	return &group.GroupListResult{
		Groups:     groups,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}