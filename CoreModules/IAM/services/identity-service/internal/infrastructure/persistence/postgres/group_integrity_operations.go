/**
 * INNOVABIZ IAM - Operações de Integridade de Grupos em PostgreSQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de verificação de integridade e transações para grupos
 * usando PostgreSQL, seguindo a arquitetura multi-dimensional, multi-tenant e com 
 * observabilidade total da plataforma INNOVABIZ.
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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"go.opentelemetry.io/otel/attribute"
)

// CheckGroupCircularReference verifica se existe uma referência circular na hierarquia de grupos
func (r *GroupRepository) CheckGroupCircularReference(ctx context.Context, groupID, parentID, tenantID uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.CheckGroupCircularReference")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("parent.id", parentID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := r.metrics.Timer("repository.group.checkCircularReference.duration")
	defer timer.ObserveDuration()

	// Se os IDs são iguais, já é uma referência circular
	if groupID == parentID {
		r.metrics.Counter("repository.group.checkCircularReference.detected").Inc(1)
		return true, nil
	}

	// Consulta recursiva para verificar se o grupo é ancestral dele mesmo
	query := `
		WITH RECURSIVE group_path AS (
			-- Grupo base
			SELECT id, parent_group_id, 1 AS level
			FROM iam_groups
			WHERE id = $1 AND tenant_id = $3 AND deleted_at IS NULL
			
			UNION ALL
			
			-- Ancestrais recursivos
			SELECT g.id, g.parent_group_id, gp.level + 1
			FROM group_path gp
			JOIN iam_groups g ON gp.parent_group_id = g.id
			WHERE g.tenant_id = $3 AND g.deleted_at IS NULL
		)
		SELECT EXISTS (
			SELECT 1 FROM group_path WHERE id = $2
		)
	`

	tx := persistence.GetTxFromContext(ctx)

	var circular bool
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, parentID, groupID, tenantID).Scan(&circular)
	} else {
		err = r.db.QueryRowContext(ctx, query, parentID, groupID, tenantID).Scan(&circular)
	}

	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar referência circular", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"parentId": parentID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.checkCircularReference.error").Inc(1)
		return false, fmt.Errorf("erro ao verificar referência circular: %w", err)
	}

	if circular {
		r.metrics.Counter("repository.group.checkCircularReference.detected").Inc(1)
	} else {
		r.metrics.Counter("repository.group.checkCircularReference.notDetected").Inc(1)
	}

	return circular, nil
}

// BeginTx inicia uma nova transação
func (r *GroupRepository) BeginTx(ctx context.Context) (context.Context, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.BeginTx")
	defer span.End()

	timer := r.metrics.Timer("repository.group.beginTx.duration")
	defer timer.ObserveDuration()

	// Verificar se já existe uma transação no contexto
	if tx := persistence.GetTxFromContext(ctx); tx != nil {
		r.logger.Warn(ctx, "Tentativa de iniciar transação quando já existe uma no contexto")
		r.metrics.Counter("repository.group.beginTx.alreadyExists").Inc(1)
		return ctx, nil
	}

	// Iniciar nova transação
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})

	if err != nil {
		r.logger.Error(ctx, "Erro ao iniciar transação", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("repository.group.beginTx.error").Inc(1)
		return ctx, fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	// Adicionar transação ao contexto
	ctx = persistence.SetTxToContext(ctx, tx)
	r.metrics.Counter("repository.group.beginTx.success").Inc(1)

	return ctx, nil
}

// CommitTx confirma uma transação
func (r *GroupRepository) CommitTx(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.CommitTx")
	defer span.End()

	timer := r.metrics.Timer("repository.group.commitTx.duration")
	defer timer.ObserveDuration()

	// Obter transação do contexto
	tx := persistence.GetTxFromContext(ctx)
	if tx == nil {
		r.logger.Error(ctx, "Tentativa de confirmar transação inexistente")
		r.metrics.Counter("repository.group.commitTx.noTransaction").Inc(1)
		return errors.New("nenhuma transação ativa para confirmar")
	}

	// Confirmar transação
	if err := tx.Commit(); err != nil {
		r.logger.Error(ctx, "Erro ao confirmar transação", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("repository.group.commitTx.error").Inc(1)
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	r.metrics.Counter("repository.group.commitTx.success").Inc(1)
	return nil
}

// RollbackTx reverte uma transação
func (r *GroupRepository) RollbackTx(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.RollbackTx")
	defer span.End()

	timer := r.metrics.Timer("repository.group.rollbackTx.duration")
	defer timer.ObserveDuration()

	// Obter transação do contexto
	tx := persistence.GetTxFromContext(ctx)
	if tx == nil {
		r.logger.Error(ctx, "Tentativa de reverter transação inexistente")
		r.metrics.Counter("repository.group.rollbackTx.noTransaction").Inc(1)
		return errors.New("nenhuma transação ativa para reverter")
	}

	// Reverter transação
	if err := tx.Rollback(); err != nil {
		r.logger.Error(ctx, "Erro ao reverter transação", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("repository.group.rollbackTx.error").Inc(1)
		return fmt.Errorf("erro ao reverter transação: %w", err)
	}

	r.metrics.Counter("repository.group.rollbackTx.success").Inc(1)
	return nil
}

// GetGroupsStatistics obtém estatísticas sobre um grupo específico ou todos os grupos
func (r *GroupRepository) GetGroupsStatistics(ctx context.Context, groupID *uuid.UUID, tenantID uuid.UUID) (*group.GroupStatistics, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.GetGroupsStatistics")
	defer span.End()

	if groupID != nil {
		span.SetAttributes(attribute.String("group.id", groupID.String()))
	}
	span.SetAttributes(attribute.String("tenant.id", tenantID.String()))

	timer := r.metrics.Timer("repository.group.getGroupsStatistics.duration")
	defer timer.ObserveDuration()

	var query string
	var args []interface{}

	if groupID != nil {
		// Estatísticas para um grupo específico
		query = `
			WITH RECURSIVE group_hierarchy AS (
				-- Grupo base
				SELECT id FROM iam_groups 
				WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
				
				UNION ALL
				
				-- Subgrupos recursivos
				SELECT g.id FROM iam_groups g
				JOIN group_hierarchy gh ON g.parent_group_id = gh.id
				WHERE g.tenant_id = $2 AND g.deleted_at IS NULL
			),
			user_counts AS (
				SELECT 
					COUNT(DISTINCT gm.user_id) as total_users,
					COUNT(DISTINCT gm.group_id) as groups_with_users
				FROM iam_group_members gm
				JOIN group_hierarchy gh ON gm.group_id = gh.id
				WHERE gm.tenant_id = $2
			),
			group_counts AS (
				SELECT 
					COUNT(*) as total_groups,
					SUM(CASE WHEN g.parent_group_id IS NULL THEN 1 ELSE 0 END) as root_groups,
					MAX(g.level) as max_depth
				FROM iam_groups g
				JOIN group_hierarchy gh ON g.id = gh.id
				WHERE g.tenant_id = $2 AND g.deleted_at IS NULL
			)
			SELECT 
				uc.total_users, 
				uc.groups_with_users, 
				gc.total_groups,
				gc.root_groups,
				gc.max_depth
			FROM user_counts uc, group_counts gc
		`
		args = []interface{}{*groupID, tenantID}
	} else {
		// Estatísticas globais para o tenant
		query = `
			WITH user_counts AS (
				SELECT 
					COUNT(DISTINCT gm.user_id) as total_users,
					COUNT(DISTINCT gm.group_id) as groups_with_users
				FROM iam_group_members gm
				WHERE gm.tenant_id = $1
			),
			group_counts AS (
				SELECT 
					COUNT(*) as total_groups,
					SUM(CASE WHEN g.parent_group_id IS NULL THEN 1 ELSE 0 END) as root_groups,
					MAX(g.level) as max_depth
				FROM iam_groups g
				WHERE g.tenant_id = $1 AND g.deleted_at IS NULL
			)
			SELECT 
				uc.total_users, 
				uc.groups_with_users, 
				gc.total_groups,
				gc.root_groups,
				gc.max_depth
			FROM user_counts uc, group_counts gc
		`
		args = []interface{}{tenantID}
	}

	tx := persistence.GetTxFromContext(ctx)

	var stats group.GroupStatistics
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, args...).Scan(
			&stats.TotalUsers,
			&stats.GroupsWithUsers,
			&stats.TotalGroups,
			&stats.RootGroups,
			&stats.MaxDepth,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, args...).Scan(
			&stats.TotalUsers,
			&stats.GroupsWithUsers,
			&stats.TotalGroups,
			&stats.RootGroups,
			&stats.MaxDepth,
		)
	}

	if err != nil {
		r.logger.Error(ctx, "Erro ao obter estatísticas de grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID.String(),
			"groupId":  groupID,
		})
		r.metrics.Counter("repository.group.getGroupsStatistics.error").Inc(1)
		return nil, fmt.Errorf("erro ao obter estatísticas de grupos: %w", err)
	}

	r.metrics.Counter("repository.group.getGroupsStatistics.success").Inc(1)
	return &stats, nil
}

// GetGroupUserCount obtém o número de usuários em um grupo específico
func (r *GroupRepository) GetGroupUserCount(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool) (int, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.GetGroupUserCount")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
	)

	timer := r.metrics.Timer("repository.group.getGroupUserCount.duration")
	defer timer.ObserveDuration()

	var query string
	var args []interface{}

	if recursive {
		// Contagem recursiva incluindo usuários de subgrupos
		query = `
			WITH RECURSIVE group_hierarchy AS (
				-- Grupo base
				SELECT id FROM iam_groups 
				WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
				
				UNION ALL
				
				-- Subgrupos recursivos
				SELECT g.id FROM iam_groups g
				JOIN group_hierarchy gh ON g.parent_group_id = gh.id
				WHERE g.tenant_id = $2 AND g.deleted_at IS NULL
			)
			SELECT COUNT(DISTINCT gm.user_id)
			FROM iam_group_members gm
			JOIN group_hierarchy gh ON gm.group_id = gh.id
			WHERE gm.tenant_id = $2
		`
	} else {
		// Contagem direta apenas para o grupo específico
		query = `
			SELECT COUNT(*)
			FROM iam_group_members gm
			WHERE gm.group_id = $1 AND gm.tenant_id = $2
		`
	}

	args = []interface{}{groupID, tenantID}

	tx := persistence.GetTxFromContext(ctx)

	var count int
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	} else {
		err = r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	}

	if err != nil {
		r.logger.Error(ctx, "Erro ao obter contagem de usuários do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
			"recursive": recursive,
		})
		r.metrics.Counter("repository.group.getGroupUserCount.error").Inc(1)
		return 0, fmt.Errorf("erro ao obter contagem de usuários: %w", err)
	}

	r.metrics.Counter("repository.group.getGroupUserCount.success").Inc(1)
	return count, nil
}