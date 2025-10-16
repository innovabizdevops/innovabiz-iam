/**
 * INNOVABIZ IAM - Operações de Atualização de Grupos em PostgreSQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de atualização, deleção e status para grupos
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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"go.opentelemetry.io/otel/attribute"
)

// Update atualiza os dados de um grupo existente
func (r *GroupRepository) Update(ctx context.Context, g *group.Group) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", g.ID.String()),
		attribute.String("group.code", g.Code),
		attribute.String("tenant.id", g.TenantID.String()),
	)

	timer := r.metrics.Timer("repository.group.update.duration")
	defer timer.ObserveDuration()

	// Verificar se o grupo existe
	existingGroup, err := r.GetByID(ctx, g.ID, g.TenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			r.metrics.Counter("repository.group.update.notFound").Inc(1)
			return group.ErrGroupNotFound
		}
		r.metrics.Counter("repository.group.update.error").Inc(1)
		return err
	}

	// Verificar se o código foi alterado e se o novo código já existe
	if existingGroup.Code != g.Code {
		exists, err := r.checkCodeExists(ctx, g.Code, g.TenantID, &g.ID)
		if err != nil {
			return err
		}

		if exists {
			r.metrics.Counter("repository.group.update.codeExists").Inc(1)
			return group.ErrGroupAlreadyExists
		}
	}

	// Processar metadados para salvar como JSON
	var metadataJSON sql.NullString
	if g.Metadata != nil {
		metadataBytes, err := json.Marshal(g.Metadata)
		if err != nil {
			r.logger.Error(ctx, "Erro ao serializar metadados do grupo", logging.Fields{
				"error": err.Error(),
				"groupId": g.ID.String(),
			})
			r.metrics.Counter("repository.group.update.error").Inc(1)
			return fmt.Errorf("erro ao serializar metadados: %w", err)
		}
		metadataJSON = sql.NullString{
			String: string(metadataBytes),
			Valid:  true,
		}
	}

	// Verificar se houve mudança na hierarquia e atualizar path/level se necessário
	path := g.Path
	level := g.Level

	if g.ParentGroupID != existingGroup.ParentGroupID {
		// Hierarquia mudou, precisamos recalcular path e level
		
		// Se tem um novo grupo pai
		if g.ParentGroupID != nil {
			// Verificar se o novo pai existe
			parent, err := r.GetByID(ctx, *g.ParentGroupID, g.TenantID)
			if err != nil {
				r.logger.Error(ctx, "Erro ao buscar novo grupo pai", logging.Fields{
					"error": err.Error(),
					"parentId": g.ParentGroupID.String(),
					"groupId": g.ID.String(),
				})
				r.metrics.Counter("repository.group.update.error").Inc(1)
				return fmt.Errorf("erro ao buscar novo grupo pai: %w", err)
			}

			// Verificar referência circular com o novo pai
			hasCircular, err := r.CheckGroupCircularReference(ctx, g.ID, *g.ParentGroupID, g.TenantID)
			if err != nil {
				r.logger.Error(ctx, "Erro ao verificar referência circular", logging.Fields{
					"error": err.Error(),
					"groupId": g.ID.String(),
					"parentId": g.ParentGroupID.String(),
				})
				r.metrics.Counter("repository.group.update.error").Inc(1)
				return err
			}

			if hasCircular {
				r.metrics.Counter("repository.group.update.circularReference").Inc(1)
				return group.ErrGroupCircularReference
			}

			// Calcular novo path e level
			path = parent.Path + "." + g.Code
			level = parent.Level + 1

			// Verificar nível máximo de hierarquia
			if level > 10 { // Limitar a 10 níveis de hierarquia
				r.metrics.Counter("repository.group.update.hierarchyTooDeep").Inc(1)
				return group.ErrGroupHierarchyTooDeep
			}
		} else {
			// Se não tem grupo pai, é um grupo raiz
			path = g.Code
			level = 1
		}

		// Se o código mudou, precisamos também atualizar os paths de todos os subgrupos
		if existingGroup.Code != g.Code {
			if err := r.updateChildGroupsPaths(ctx, g.ID, g.TenantID); err != nil {
				r.logger.Error(ctx, "Erro ao atualizar paths dos subgrupos", logging.Fields{
					"error": err.Error(),
					"groupId": g.ID.String(),
				})
				r.metrics.Counter("repository.group.update.error").Inc(1)
				return fmt.Errorf("erro ao atualizar paths dos subgrupos: %w", err)
			}
		}
	} else if existingGroup.Code != g.Code {
		// Apenas o código mudou, precisamos atualizar o path deste grupo e de seus subgrupos
		
		// Se tem pai, recalcular path com o pai atual
		if g.ParentGroupID != nil {
			parent, err := r.GetByID(ctx, *g.ParentGroupID, g.TenantID)
			if err != nil {
				r.logger.Error(ctx, "Erro ao buscar grupo pai para atualizar path", logging.Fields{
					"error": err.Error(),
					"parentId": g.ParentGroupID.String(),
					"groupId": g.ID.String(),
				})
				r.metrics.Counter("repository.group.update.error").Inc(1)
				return fmt.Errorf("erro ao buscar grupo pai: %w", err)
			}
			path = parent.Path + "." + g.Code
		} else {
			// Se não tem pai, é um grupo raiz
			path = g.Code
		}

		// Atualizar paths dos subgrupos
		if err := r.updateChildGroupsPaths(ctx, g.ID, g.TenantID); err != nil {
			r.logger.Error(ctx, "Erro ao atualizar paths dos subgrupos", logging.Fields{
				"error": err.Error(),
				"groupId": g.ID.String(),
			})
			r.metrics.Counter("repository.group.update.error").Inc(1)
			return fmt.Errorf("erro ao atualizar paths dos subgrupos: %w", err)
		}
	}

	// Atualizar campos de auditoria
	g.UpdatedAt = time.Now().UTC()

	// Preparar query para atualização
	query := `
		UPDATE iam_groups
		SET 
			code = $1,
			name = $2,
			description = $3,
			region_code = $4,
			group_type = $5,
			status = $6,
			path = $7,
			level = $8,
			parent_group_id = $9,
			updated_at = $10,
			updated_by = $11,
			metadata = $12
		WHERE id = $13 AND tenant_id = $14 AND deleted_at IS NULL
	`

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var result sql.Result
	var execErr error

	if tx != nil {
		result, execErr = tx.ExecContext(ctx, query,
			g.Code, g.Name, g.Description, g.RegionCode,
			g.GroupType, g.Status, path, level, g.ParentGroupID,
			g.UpdatedAt, g.UpdatedBy, metadataJSON,
			g.ID, g.TenantID,
		)
	} else {
		result, execErr = r.db.ExecContext(ctx, query,
			g.Code, g.Name, g.Description, g.RegionCode,
			g.GroupType, g.Status, path, level, g.ParentGroupID,
			g.UpdatedAt, g.UpdatedBy, metadataJSON,
			g.ID, g.TenantID,
		)
	}

	if execErr != nil {
		r.logger.Error(ctx, "Erro ao atualizar grupo", logging.Fields{
			"error":   execErr.Error(),
			"groupId": g.ID.String(),
		})
		r.metrics.Counter("repository.group.update.error").Inc(1)
		return fmt.Errorf("erro ao atualizar grupo: %w", execErr)
	}

	// Verificar se a atualização foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar linhas afetadas na atualização de grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": g.ID.String(),
		})
		r.metrics.Counter("repository.group.update.error").Inc(1)
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Error(ctx, "Nenhuma linha afetada ao atualizar grupo", logging.Fields{
			"groupId": g.ID.String(),
		})
		r.metrics.Counter("repository.group.update.notFound").Inc(1)
		return group.ErrGroupNotFound
	}

	// Atualizar g com os novos valores de path e level
	g.Path = path
	g.Level = level

	r.metrics.Counter("repository.group.update.success").Inc(1)
	return nil
}

// updateChildGroupsPaths atualiza recursivamente os paths dos subgrupos após mudança de código ou hierarquia
func (r *GroupRepository) updateChildGroupsPaths(ctx context.Context, groupID, tenantID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.updateChildGroupsPaths")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := r.metrics.Timer("repository.group.updateChildPaths.duration")
	defer timer.ObserveDuration()

	// Obter o grupo atualizado para ter o path correto
	group, err := r.GetByID(ctx, groupID, tenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter grupo para atualização de paths", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		return err
	}

	// Buscar todos os filhos diretos para atualização
	query := `
		SELECT id, code FROM iam_groups 
		WHERE parent_group_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	tx := persistence.GetTxFromContext(ctx)

	var rows *sql.Rows
	var queryErr error

	if tx != nil {
		rows, queryErr = tx.QueryContext(ctx, query, groupID, tenantID)
	} else {
		rows, queryErr = r.db.QueryContext(ctx, query, groupID, tenantID)
	}

	if queryErr != nil {
		r.logger.Error(ctx, "Erro ao buscar subgrupos para atualização de paths", logging.Fields{
			"error":   queryErr.Error(),
			"groupId": groupID.String(),
		})
		return fmt.Errorf("erro ao buscar subgrupos: %w", queryErr)
	}
	defer rows.Close()

	// Processar cada subgrupo
	for rows.Next() {
		var childID uuid.UUID
		var childCode string

		if err := rows.Scan(&childID, &childCode); err != nil {
			r.logger.Error(ctx, "Erro ao processar subgrupo", logging.Fields{
				"error":   err.Error(),
				"groupId": groupID.String(),
			})
			continue
		}

		// Calcular novo path para o subgrupo
		newPath := group.Path + "." + childCode
		newLevel := group.Level + 1

		// Atualizar path do subgrupo
		updateQuery := `
			UPDATE iam_groups
			SET path = $1, level = $2, updated_at = $3
			WHERE id = $4 AND tenant_id = $5 AND deleted_at IS NULL
		`

		var updateResult sql.Result
		var updateErr error
		now := time.Now().UTC()

		if tx != nil {
			updateResult, updateErr = tx.ExecContext(ctx, updateQuery, newPath, newLevel, now, childID, tenantID)
		} else {
			updateResult, updateErr = r.db.ExecContext(ctx, updateQuery, newPath, newLevel, now, childID, tenantID)
		}

		if updateErr != nil {
			r.logger.Error(ctx, "Erro ao atualizar path do subgrupo", logging.Fields{
				"error":    updateErr.Error(),
				"groupId":  groupID.String(),
				"childId":  childID.String(),
				"childPath": newPath,
			})
			return fmt.Errorf("erro ao atualizar path do subgrupo: %w", updateErr)
		}

		// Verificar se a atualização foi bem-sucedida
		rowsAffected, err := updateResult.RowsAffected()
		if err != nil {
			r.logger.Error(ctx, "Erro ao verificar linhas afetadas na atualização de path", logging.Fields{
				"error":   err.Error(),
				"childId": childID.String(),
			})
			return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
		}

		if rowsAffected == 0 {
			r.logger.Warn(ctx, "Nenhuma linha afetada ao atualizar path do subgrupo", logging.Fields{
				"childId": childID.String(),
			})
		}

		// Recursivamente atualizar subgrupos deste grupo
		if err := r.updateChildGroupsPaths(ctx, childID, tenantID); err != nil {
			r.logger.Error(ctx, "Erro ao atualizar paths de subgrupos recursivamente", logging.Fields{
				"error":   err.Error(),
				"childId": childID.String(),
			})
			return err
		}
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(ctx, "Erro ao iterar subgrupos", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		return fmt.Errorf("erro ao iterar subgrupos: %w", err)
	}

	return nil
}

// ChangeStatus atualiza o status de um grupo
func (r *GroupRepository) ChangeStatus(ctx context.Context, groupID, tenantID uuid.UUID, status string, updatedBy *uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.ChangeStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("status", status),
	)

	timer := r.metrics.Timer("repository.group.changeStatus.duration")
	defer timer.ObserveDuration()

	// Verificar se o grupo existe
	_, err := r.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			r.metrics.Counter("repository.group.changeStatus.notFound").Inc(1)
			return group.ErrGroupNotFound
		}
		r.metrics.Counter("repository.group.changeStatus.error").Inc(1)
		return err
	}

	// Validar status
	validStatus := map[string]bool{
		group.StatusActive:   true,
		group.StatusInactive: true,
		group.StatusLocked:   true,
	}

	if !validStatus[status] {
		r.logger.Error(ctx, "Status inválido", logging.Fields{
			"status":  status,
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.changeStatus.invalidStatus").Inc(1)
		return group.ErrInvalidStatus
	}

	// Atualizar o status do grupo
	query := `
		UPDATE iam_groups
		SET status = $1, updated_at = $2, updated_by = $3
		WHERE id = $4 AND tenant_id = $5 AND deleted_at IS NULL
	`

	// Definir horário de atualização
	now := time.Now().UTC()

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var result sql.Result
	var execErr error

	if tx != nil {
		result, execErr = tx.ExecContext(ctx, query, status, now, updatedBy, groupID, tenantID)
	} else {
		result, execErr = r.db.ExecContext(ctx, query, status, now, updatedBy, groupID, tenantID)
	}

	if execErr != nil {
		r.logger.Error(ctx, "Erro ao atualizar status do grupo", logging.Fields{
			"error":   execErr.Error(),
			"groupId": groupID.String(),
			"status":  status,
		})
		r.metrics.Counter("repository.group.changeStatus.error").Inc(1)
		return fmt.Errorf("erro ao atualizar status do grupo: %w", execErr)
	}

	// Verificar se a atualização foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar linhas afetadas na atualização de status", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.changeStatus.error").Inc(1)
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Error(ctx, "Nenhuma linha afetada ao atualizar status do grupo", logging.Fields{
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.changeStatus.notFound").Inc(1)
		return group.ErrGroupNotFound
	}

	r.metrics.Counter("repository.group.changeStatus.success").Inc(1)
	return nil
}

// SoftDelete realiza uma exclusão lógica do grupo
func (r *GroupRepository) SoftDelete(ctx context.Context, groupID, tenantID uuid.UUID, deletedBy *uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.SoftDelete")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := r.metrics.Timer("repository.group.softDelete.duration")
	defer timer.ObserveDuration()

	// Verificar se o grupo existe
	existingGroup, err := r.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			r.metrics.Counter("repository.group.softDelete.notFound").Inc(1)
			return group.ErrGroupNotFound
		}
		r.metrics.Counter("repository.group.softDelete.error").Inc(1)
		return err
	}

	// Verificar se o grupo tem subgrupos
	childGroups, err := r.GetChildGroups(ctx, groupID, tenantID, false, 1, 1)
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar subgrupos", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.softDelete.error").Inc(1)
		return fmt.Errorf("erro ao verificar subgrupos: %w", err)
	}

	if childGroups.TotalCount > 0 {
		r.logger.Error(ctx, "Grupo possui subgrupos e não pode ser excluído", logging.Fields{
			"groupId":    groupID.String(),
			"childCount": childGroups.TotalCount,
		})
		r.metrics.Counter("repository.group.softDelete.hasChildren").Inc(1)
		return group.ErrGroupHasChildren
	}

	// Verificar se o grupo tem usuários associados
	userCount, err := r.GetGroupUserCount(ctx, groupID, tenantID, false)
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar usuários do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.softDelete.error").Inc(1)
		return fmt.Errorf("erro ao verificar usuários do grupo: %w", err)
	}

	if userCount > 0 {
		r.logger.Error(ctx, "Grupo possui usuários e não pode ser excluído", logging.Fields{
			"groupId":   groupID.String(),
			"userCount": userCount,
		})
		r.metrics.Counter("repository.group.softDelete.hasUsers").Inc(1)
		return group.ErrGroupHasUsers
	}

	// Realizar a exclusão lógica
	query := `
		UPDATE iam_groups
		SET deleted_at = $1, deleted_by = $2
		WHERE id = $3 AND tenant_id = $4 AND deleted_at IS NULL
	`

	// Definir horário de exclusão
	now := time.Now().UTC()

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var result sql.Result
	var execErr error

	if tx != nil {
		result, execErr = tx.ExecContext(ctx, query, now, deletedBy, groupID, tenantID)
	} else {
		result, execErr = r.db.ExecContext(ctx, query, now, deletedBy, groupID, tenantID)
	}

	if execErr != nil {
		r.logger.Error(ctx, "Erro ao excluir grupo logicamente", logging.Fields{
			"error":   execErr.Error(),
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.softDelete.error").Inc(1)
		return fmt.Errorf("erro ao excluir grupo logicamente: %w", execErr)
	}

	// Verificar se a exclusão foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar linhas afetadas na exclusão lógica", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.softDelete.error").Inc(1)
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Error(ctx, "Nenhuma linha afetada ao excluir grupo logicamente", logging.Fields{
			"groupId": groupID.String(),
		})
		r.metrics.Counter("repository.group.softDelete.notFound").Inc(1)
		return group.ErrGroupNotFound
	}

	r.metrics.Counter("repository.group.softDelete.success").Inc(1)
	return nil
}