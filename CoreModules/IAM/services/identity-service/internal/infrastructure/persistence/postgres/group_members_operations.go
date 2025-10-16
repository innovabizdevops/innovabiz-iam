/**
 * INNOVABIZ IAM - Operações de Membros do Grupo em PostgreSQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações relacionadas a membros de grupos usando PostgreSQL,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15, A.5.16 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4, 7.2.5 - Gestão de grupos e acessos)
 * - LGPD/GDPR/PDPA (Minimização de dados e controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade e acesso)
 */

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"go.opentelemetry.io/otel/attribute"
)

// AddUserToGroup adiciona um usuário a um grupo
func (r *GroupRepository) AddUserToGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.AddUserToGroup")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.addUserToGroup.duration")
	defer timer.ObserveDuration()

	// Verificar se o usuário já pertence ao grupo
	isInGroup, err := r.IsUserInGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		return err
	}

	if isInGroup {
		r.logger.Info(ctx, "Usuário já pertence ao grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return nil // Não é um erro, o usuário já está no grupo
	}

	// Inserir o usuário no grupo
	query := `
		INSERT INTO iam_group_members (
			group_id, user_id, tenant_id, created_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`

	// Extrair transação e ID do usuário atual do contexto
	tx := persistence.GetTxFromContext(ctx)
	currentUserID := getCurrentUserID(ctx)

	var result sql.Result
	var execErr error

	// Registrar horário atual
	now := time.Now().UTC()

	if tx != nil {
		result, execErr = tx.ExecContext(ctx, query,
			groupID, userID, tenantID, now, currentUserID)
	} else {
		result, execErr = r.db.ExecContext(ctx, query,
			groupID, userID, tenantID, now, currentUserID)
	}

	if execErr != nil {
		r.logger.Error(ctx, "Erro ao adicionar usuário ao grupo", logging.Fields{
			"error":    execErr.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.addUserToGroup.error").Inc(1)
		return fmt.Errorf("erro ao adicionar usuário ao grupo: %w", execErr)
	}

	// Verificar se a inserção foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar linhas afetadas na adição de usuário ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.addUserToGroup.error").Inc(1)
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Error(ctx, "Nenhuma linha afetada ao adicionar usuário ao grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.addUserToGroup.error").Inc(1)
		return fmt.Errorf("nenhuma linha afetada ao adicionar usuário ao grupo")
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.addUserToGroup.success").Inc(1)

	return nil
}

// RemoveUserFromGroup remove um usuário de um grupo
func (r *GroupRepository) RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.RemoveUserFromGroup")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.removeUserFromGroup.duration")
	defer timer.ObserveDuration()

	// Verificar se o usuário pertence ao grupo
	isInGroup, err := r.IsUserInGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		return err
	}

	if !isInGroup {
		r.logger.Info(ctx, "Usuário não pertence ao grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return group.ErrUserNotInGroup
	}

	// Remover o usuário do grupo
	query := `
		DELETE FROM iam_group_members
		WHERE group_id = $1 AND user_id = $2 AND tenant_id = $3
	`

	// Extrair transação do contexto
	tx := persistence.GetTxFromContext(ctx)

	var result sql.Result
	var execErr error

	if tx != nil {
		result, execErr = tx.ExecContext(ctx, query, groupID, userID, tenantID)
	} else {
		result, execErr = r.db.ExecContext(ctx, query, groupID, userID, tenantID)
	}

	if execErr != nil {
		r.logger.Error(ctx, "Erro ao remover usuário do grupo", logging.Fields{
			"error":    execErr.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.removeUserFromGroup.error").Inc(1)
		return fmt.Errorf("erro ao remover usuário do grupo: %w", execErr)
	}

	// Verificar se a remoção foi bem-sucedida
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar linhas afetadas na remoção de usuário do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.removeUserFromGroup.error").Inc(1)
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Error(ctx, "Nenhuma linha afetada ao remover usuário do grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.removeUserFromGroup.error").Inc(1)
		return fmt.Errorf("nenhuma linha afetada ao remover usuário do grupo")
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.removeUserFromGroup.success").Inc(1)

	return nil
}

// IsUserInGroup verifica se um usuário pertence a um grupo
func (r *GroupRepository) IsUserInGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.IsUserInGroup")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.isUserInGroup.duration")
	defer timer.ObserveDuration()

	query := `
		SELECT EXISTS (
			SELECT 1 FROM iam_group_members
			WHERE group_id = $1 AND user_id = $2 AND tenant_id = $3
		)
	`

	// Extrair transação do contexto
	tx := persistence.GetTxFromContext(ctx)

	var exists bool
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, groupID, userID, tenantID).Scan(&exists)
	} else {
		err = r.db.QueryRowContext(ctx, query, groupID, userID, tenantID).Scan(&exists)
	}

	if err != nil {
		r.logger.Error(ctx, "Erro ao verificar se usuário pertence ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.isUserInGroup.error").Inc(1)
		return false, fmt.Errorf("erro ao verificar se usuário pertence ao grupo: %w", err)
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.isUserInGroup.success").Inc(1)

	return exists, nil
}

// ListGroupMembers lista os usuários membros de um grupo
func (r *GroupRepository) ListGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, page, pageSize int, filter map[string]interface{}) (*group.UserListResult, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.ListGroupMembers")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.listGroupMembers.duration")
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
	var baseQuery string
	var countQuery string
	var args []interface{}

	if recursive {
		// Busca recursiva inclui membros de subgrupos
		baseQuery = `
			WITH RECURSIVE group_hierarchy AS (
				-- Grupo base
				SELECT id, parent_group_id FROM iam_groups 
				WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
				
				UNION ALL
				
				-- Subgrupos recursivos
				SELECT g.id, g.parent_group_id FROM iam_groups g
				JOIN group_hierarchy gh ON g.parent_group_id = gh.id
				WHERE g.tenant_id = $2 AND g.deleted_at IS NULL
			)
			SELECT u.id, u.username, u.display_name, u.email, u.status,
				   u.tenant_id, u.metadata, u.created_at, u.created_by
			FROM iam_users u
			JOIN iam_group_members gm ON u.id = gm.user_id
			JOIN group_hierarchy gh ON gm.group_id = gh.id
			WHERE u.tenant_id = $2 AND u.deleted_at IS NULL
		`
		
		countQuery = `
			WITH RECURSIVE group_hierarchy AS (
				-- Grupo base
				SELECT id, parent_group_id FROM iam_groups 
				WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
				
				UNION ALL
				
				-- Subgrupos recursivos
				SELECT g.id, g.parent_group_id FROM iam_groups g
				JOIN group_hierarchy gh ON g.parent_group_id = gh.id
				WHERE g.tenant_id = $2 AND g.deleted_at IS NULL
			)
			SELECT COUNT(DISTINCT u.id)
			FROM iam_users u
			JOIN iam_group_members gm ON u.id = gm.user_id
			JOIN group_hierarchy gh ON gm.group_id = gh.id
			WHERE u.tenant_id = $2 AND u.deleted_at IS NULL
		`
	} else {
		// Busca direta apenas membros do grupo especificado
		baseQuery = `
			SELECT u.id, u.username, u.display_name, u.email, u.status,
				   u.tenant_id, u.metadata, u.created_at, u.created_by
			FROM iam_users u
			JOIN iam_group_members gm ON u.id = gm.user_id
			WHERE gm.group_id = $1 AND u.tenant_id = $2 AND u.deleted_at IS NULL
		`
		
		countQuery = `
			SELECT COUNT(*)
			FROM iam_users u
			JOIN iam_group_members gm ON u.id = gm.user_id
			WHERE gm.group_id = $1 AND u.tenant_id = $2 AND u.deleted_at IS NULL
		`
	}

	// Inicializar argumentos com valores base
	args = []interface{}{groupID, tenantID}

	// Aplicar filtros adicionais, se fornecidos
	whereFilters := []string{}
	argIndex := 3 // Próximo índice de argumento

	if filter != nil {
		// Filtro por termo de busca (nome de usuário, nome de exibição ou email)
		if searchTerm, ok := filter["searchTerm"].(string); ok && searchTerm != "" {
			whereFilters = append(whereFilters, fmt.Sprintf("(u.username ILIKE $%d OR u.display_name ILIKE $%d OR u.email ILIKE $%d)", 
				argIndex, argIndex, argIndex))
			args = append(args, "%"+searchTerm+"%")
			argIndex++
		}

		// Filtro por status
		if statusArray, ok := filter["status"].([]string); ok && len(statusArray) > 0 {
			placeholders := make([]string, 0, len(statusArray))
			for _, status := range statusArray {
				placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
				args = append(args, status)
				argIndex++
			}
			whereFilters = append(whereFilters, fmt.Sprintf("u.status IN (%s)", 
				strings.Join(placeholders, ",")))
		}
	}

	// Adicionar cláusulas WHERE adicionais, se existirem
	if len(whereFilters) > 0 {
		whereClause := " AND " + strings.Join(whereFilters, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Adicionar ordenação e paginação
	sortBy := "u.display_name"
	sortDirection := "ASC"

	if filter != nil {
		if sortByFilter, ok := filter["sortBy"].(string); ok && sortByFilter != "" {
			// Validar campo de ordenação para evitar injeção de SQL
			validFields := map[string]string{
				"username":    "u.username",
				"displayName": "u.display_name",
				"email":       "u.email",
				"status":      "u.status",
				"createdAt":   "u.created_at",
			}
			
			if validField, ok := validFields[sortByFilter]; ok {
				sortBy = validField
			}
		}
		
		if sortDirFilter, ok := filter["sortDirection"].(string); ok {
			// Validar direção de ordenação
			if strings.ToUpper(sortDirFilter) == "DESC" {
				sortDirection = "DESC"
			}
		}
	}

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
			r.logger.Error(ctx, "Erro ao contar membros do grupo", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.listGroupMembers.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar membros do grupo: %w", err)
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
		if err != nil {
			r.logger.Error(ctx, "Erro ao contar membros do grupo", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			r.metrics.Counter("repository.group.listGroupMembers.error").Inc(1)
			return nil, fmt.Errorf("erro ao contar membros do grupo: %w", err)
		}
	}

	// Calcular total de páginas
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// Se não houver registros, retornar resultado vazio
	if totalCount == 0 {
		return &group.UserListResult{
			Users:      []*group.UserDTO{},
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
		r.logger.Error(ctx, "Erro ao listar membros do grupo", logging.Fields{
			"error":    queryErr.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.listGroupMembers.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar membros do grupo: %w", queryErr)
	}
	defer rows.Close()

	// Processar resultados
	users := []*group.UserDTO{}

	for rows.Next() {
		var user group.UserDTO
		var metadata sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.DisplayName,
			&user.Email,
			&user.Status,
			&user.TenantID,
			&metadata,
			&user.CreatedAt,
			&user.CreatedBy,
		)

		if err != nil {
			r.logger.Error(ctx, "Erro ao processar registro de usuário", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			continue
		}

		// Processar metadados (JSON)
		if metadata.Valid && metadata.String != "" {
			user.Metadata = make(map[string]interface{})
			if err := json.Unmarshal([]byte(metadata.String), &user.Metadata); err != nil {
				r.logger.Warn(ctx, "Erro ao processar metadados do usuário", logging.Fields{
					"error":  err.Error(),
					"userId": user.ID.String(),
				})
			}
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(ctx, "Erro ao iterar resultados de usuários", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		r.metrics.Counter("repository.group.listGroupMembers.error").Inc(1)
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.listGroupMembers.success").Inc(1)

	// Retornar resultado
	return &group.UserListResult{
		Users:      users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Função auxiliar para obter o ID do usuário atual do contexto
func getCurrentUserID(ctx context.Context) *uuid.UUID {
	if userID, ok := ctx.Value("current_user_id").(uuid.UUID); ok {
		return &userID
	}
	return nil
}