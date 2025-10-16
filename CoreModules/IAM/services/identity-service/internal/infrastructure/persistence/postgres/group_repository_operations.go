/**
 * INNOVABIZ IAM - Operações do Repositório PostgreSQL para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações principais do repositório de grupos usando PostgreSQL,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR/PDPA (Proteção de dados)
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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"go.opentelemetry.io/otel/attribute"
)

// GetByCode busca um grupo pelo código
func (r *GroupRepository) GetByCode(ctx context.Context, code string, tenantID uuid.UUID) (*group.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.GetByCode")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.code", code),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.getByCode.duration")
	defer timer.ObserveDuration()

	query := `
		SELECT g.id, g.code, g.name, g.description, g.tenant_id, g.region_code, 
		       g.group_type, g.status, g.path, g.level, g.parent_group_id,
		       g.created_at, g.created_by, g.updated_at, g.updated_by,
		       g.metadata
		FROM iam_groups g
		WHERE g.code = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
	`

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var row *sqlx.Row
	if tx != nil {
		row = tx.QueryRowxContext(ctx, query, code, tenantID)
	} else {
		row = r.db.QueryRowxContext(ctx, query, code, tenantID)
	}

	var g group.Group
	var metadata sql.NullString

	err := row.Scan(
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
	)

	if err != nil {
		// Incrementar contador de erro
		r.metrics.Counter("repository.group.getByCode.error").Inc(1)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrGroupNotFound
		}

		r.logger.Error(ctx, "Erro ao buscar grupo por código", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantID.String(),
		})
		return nil, fmt.Errorf("erro ao buscar grupo por código: %w", err)
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

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.getByCode.success").Inc(1)

	return &g, nil
}

// Create cria um novo grupo
func (r *GroupRepository) Create(ctx context.Context, g *group.Group) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.Create")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.code", g.Code),
		attribute.String("tenant.id", g.TenantID.String()),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.create.duration")
	defer timer.ObserveDuration()

	// Verificar se já existe um grupo com o mesmo código no mesmo tenant
	exists, err := r.checkCodeExists(ctx, g.Code, g.TenantID, nil)
	if err != nil {
		return err
	}

	if exists {
		r.metrics.Counter("repository.group.create.error").Inc(1)
		return group.ErrGroupAlreadyExists
	}

	// Processar metadados para salvar como JSON
	var metadataJSON sql.NullString
	if g.Metadata != nil {
		metadataBytes, err := json.Marshal(g.Metadata)
		if err != nil {
			r.logger.Error(ctx, "Erro ao serializar metadados do grupo", logging.Fields{
				"error": err.Error(),
			})
			r.metrics.Counter("repository.group.create.error").Inc(1)
			return fmt.Errorf("erro ao serializar metadados: %w", err)
		}
		metadataJSON = sql.NullString{
			String: string(metadataBytes),
			Valid:  true,
		}
	}

	// Construir o path com base no grupo pai
	path := g.Code
	level := 1

	if g.ParentGroupID != nil {
		parent, err := r.GetByID(ctx, *g.ParentGroupID, g.TenantID)
		if err != nil {
			r.logger.Error(ctx, "Erro ao buscar grupo pai para construção do path", logging.Fields{
				"error":        err.Error(),
				"parentId":     g.ParentGroupID.String(),
				"tenantId":     g.TenantID.String(),
			})
			r.metrics.Counter("repository.group.create.error").Inc(1)
			return fmt.Errorf("erro ao buscar grupo pai: %w", err)
		}

		path = parent.Path + "." + g.Code
		level = parent.Level + 1

		// Verificar nível máximo de hierarquia
		if level > 10 { // Limitar a 10 níveis de hierarquia
			r.metrics.Counter("repository.group.create.error").Inc(1)
			return group.ErrGroupHierarchyTooDeep
		}

		// Verificar referência circular
		hasCircular, err := r.CheckGroupCircularReference(ctx, g.ID, *g.ParentGroupID, g.TenantID)
		if err != nil {
			r.logger.Error(ctx, "Erro ao verificar referência circular", logging.Fields{
				"error":    err.Error(),
				"groupId":  g.ID.String(),
				"parentId": g.ParentGroupID.String(),
			})
			r.metrics.Counter("repository.group.create.error").Inc(1)
			return err
		}

		if hasCircular {
			r.metrics.Counter("repository.group.create.error").Inc(1)
			return group.ErrGroupCircularReference
		}
	}

	g.Path = path
	g.Level = level

	// Definir horário de criação
	g.CreatedAt = time.Now().UTC()

	// Preparar query para inserção
	query := `
		INSERT INTO iam_groups (
			id, code, name, description, tenant_id, region_code,
			group_type, status, path, level, parent_group_id,
			created_at, created_by, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, $10, $11,
			$12, $13, $14
		)
	`

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var result sql.Result
	var execErr error

	if tx != nil {
		result, execErr = tx.ExecContext(ctx, query,
			g.ID, g.Code, g.Name, g.Description, g.TenantID, g.RegionCode,
			g.GroupType, g.Status, g.Path, g.Level, g.ParentGroupID,
			g.CreatedAt, g.CreatedBy, metadataJSON,
		)
	} else {
		result, execErr = r.db.ExecContext(ctx, query,
			g.ID, g.Code, g.Name, g.Description, g.TenantID, g.RegionCode,
			g.GroupType, g.Status, g.Path, g.Level, g.ParentGroupID,
			g.CreatedAt, g.CreatedBy, metadataJSON,
		)
	}

	if execErr != nil {
		r.logger.Error(ctx, "Erro ao criar grupo", logging.Fields{
			"error":    execErr.Error(),
			"groupId":  g.ID.String(),
			"tenantId": g.TenantID.String(),
		})
		r.metrics.Counter("repository.group.create.error").Inc(1)
		return fmt.Errorf("erro ao criar grupo: %w", execErr)
	}

	// Verificar se a inserção foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar linhas afetadas na criação de grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": g.ID.String(),
		})
		r.metrics.Counter("repository.group.create.error").Inc(1)
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Error(ctx, "Nenhuma linha afetada ao criar grupo", logging.Fields{
			"groupId": g.ID.String(),
		})
		r.metrics.Counter("repository.group.create.error").Inc(1)
		return fmt.Errorf("nenhuma linha afetada ao criar grupo")
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.create.success").Inc(1)

	return nil
}

// checkCodeExists verifica se já existe um grupo com o código fornecido
// excludeID permite excluir um grupo específico da verificação (útil para updates)
func (r *GroupRepository) checkCodeExists(ctx context.Context, code string, tenantID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.checkCodeExists")
	defer span.End()

	query := `
		SELECT EXISTS (
			SELECT 1 FROM iam_groups g 
			WHERE g.code = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
	`

	args := []interface{}{code, tenantID}

	// Se um ID for fornecido para exclusão, adiciona à cláusula WHERE
	if excludeID != nil {
		query += " AND g.id != $3"
		args = append(args, *excludeID)
	}

	query += ")"

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var exists bool
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&exists)
	} else {
		err = r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	}

	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar existência de código de grupo", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantID.String(),
		})
		return false, fmt.Errorf("erro ao verificar existência de código: %w", err)
	}

	return exists, nil
}