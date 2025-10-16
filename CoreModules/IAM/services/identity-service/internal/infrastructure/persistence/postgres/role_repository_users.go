package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/domain/model"
)

// AssignUserToRole atribui um usuário a uma função
func (r *RoleRepository) AssignUserToRole(ctx context.Context, tenantID, roleID, userID, assignedBy uuid.UUID, expiresAt *time.Time) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.AssignUserToRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Verificar se o usuário existe
		userExists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !userExists {
			return model.NewUserNotFoundError(userID)
		}

		// Verificar se o usuário já está associado à função
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM user_roles
				WHERE role_id = $1 AND user_id = $2 AND tenant_id = $3
				AND (expires_at IS NULL OR expires_at > NOW())
			)
		`, roleID, userID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar atribuição de usuário existente: %w", err)
		}

		if relationExists {
			return model.NewUserAlreadyAssignedToRoleError(roleID, userID)
		}

		// Associar usuário à função
		if expiresAt != nil {
			_, err = tx.Exec(ctx, `
				INSERT INTO user_roles (
					tenant_id, role_id, user_id, 
					created_at, created_by, expires_at
				) VALUES (
					$1, $2, $3, NOW(), $4, $5
				)
			`, tenantID, roleID, userID, assignedBy, expiresAt)
		} else {
			_, err = tx.Exec(ctx, `
				INSERT INTO user_roles (
					tenant_id, role_id, user_id, 
					created_at, created_by
				) VALUES (
					$1, $2, $3, NOW(), $4
				)
			`, tenantID, roleID, userID, assignedBy)
		}

		if err != nil {
			return fmt.Errorf("erro ao atribuir usuário à função: %w", err)
		}

		// Registrar evento de auditoria
		auditAction := "ASSIGN"
		_, err = tx.Exec(ctx, `
			INSERT INTO user_role_audit (
				tenant_id, role_id, user_id,
				action, action_at, action_by, expires_at
			) VALUES (
				$1, $2, $3, $4, NOW(), $5, $6
			)
		`, tenantID, roleID, userID, auditAction, assignedBy, expiresAt)

		if err != nil {
			return fmt.Errorf("erro ao registrar auditoria de atribuição de usuário: %w", err)
		}

		return nil
	})
}

// RemoveUserFromRole remove um usuário de uma função
func (r *RoleRepository) RemoveUserFromRole(ctx context.Context, tenantID, roleID, userID, removedBy uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.RemoveUserFromRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Verificar se o usuário existe
		userExists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !userExists {
			return model.NewUserNotFoundError(userID)
		}

		// Verificar se o usuário está associado à função
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM user_roles
				WHERE role_id = $1 AND user_id = $2 AND tenant_id = $3
			)
		`, roleID, userID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar atribuição de usuário: %w", err)
		}

		if !relationExists {
			return model.NewUserNotAssignedToRoleError(roleID, userID)
		}

		// Remover usuário da função
		_, err = tx.Exec(ctx, `
			DELETE FROM user_roles
			WHERE role_id = $1 AND user_id = $2 AND tenant_id = $3
		`, roleID, userID, tenantID)

		if err != nil {
			return fmt.Errorf("erro ao remover usuário da função: %w", err)
		}

		// Registrar evento de auditoria
		auditAction := "REMOVE"
		_, err = tx.Exec(ctx, `
			INSERT INTO user_role_audit (
				tenant_id, role_id, user_id,
				action, action_at, action_by
			) VALUES (
				$1, $2, $3, $4, NOW(), $5
			)
		`, tenantID, roleID, userID, auditAction, removedBy)

		if err != nil {
			return fmt.Errorf("erro ao registrar auditoria de remoção de usuário: %w", err)
		}

		return nil
	})
}

// UpdateUserRoleExpiration atualiza a data de expiração da atribuição de um usuário a uma função
func (r *RoleRepository) UpdateUserRoleExpiration(ctx context.Context, tenantID, roleID, userID, updatedBy uuid.UUID, expiresAt *time.Time) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.UpdateUserRoleExpiration")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Verificar se o usuário existe
		userExists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !userExists {
			return model.NewUserNotFoundError(userID)
		}

		// Verificar se o usuário está associado à função
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM user_roles
				WHERE role_id = $1 AND user_id = $2 AND tenant_id = $3
			)
		`, roleID, userID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar atribuição de usuário: %w", err)
		}

		if !relationExists {
			return model.NewUserNotAssignedToRoleError(roleID, userID)
		}

		// Atualizar data de expiração
		_, err = tx.Exec(ctx, `
			UPDATE user_roles
			SET expires_at = $4, updated_at = NOW(), updated_by = $5
			WHERE role_id = $1 AND user_id = $2 AND tenant_id = $3
		`, roleID, userID, tenantID, expiresAt, updatedBy)

		if err != nil {
			return fmt.Errorf("erro ao atualizar expiração da atribuição: %w", err)
		}

		// Registrar evento de auditoria
		auditAction := "UPDATE_EXPIRATION"
		_, err = tx.Exec(ctx, `
			INSERT INTO user_role_audit (
				tenant_id, role_id, user_id,
				action, action_at, action_by, expires_at
			) VALUES (
				$1, $2, $3, $4, NOW(), $5, $6
			)
		`, tenantID, roleID, userID, auditAction, updatedBy, expiresAt)

		if err != nil {
			return fmt.Errorf("erro ao registrar auditoria de atualização de expiração: %w", err)
		}

		return nil
	})
}

// GetRoleUsers retorna os usuários associados a uma função
func (r *RoleRepository) GetRoleUsers(ctx context.Context, tenantID, roleID uuid.UUID, includeExpired bool, pagination model.Pagination) ([]*model.User, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetRoleUsers")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("include_expired", includeExpired),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	// Construção dinâmica da consulta com base em includeExpired
	var filterClause string
	if !includeExpired {
		filterClause = "AND (ur.expires_at IS NULL OR ur.expires_at > NOW())"
	}

	// Consulta base para contagem e obtenção de usuários
	baseQuery := fmt.Sprintf(`
		FROM users u
		JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1
		AND ur.tenant_id = $2
		AND u.tenant_id = $2
		AND u.deleted_at IS NULL
		%s
	`, filterClause)

	// Consulta de contagem
	countQuery := "SELECT COUNT(u.id) " + baseQuery

	// Consulta principal com ordenação e paginação
	query := fmt.Sprintf(`
		SELECT 
			u.id, u.tenant_id, u.username, u.email, u.first_name, u.last_name,
			u.is_active, u.is_verified, u.metadata, ur.expires_at,
			u.created_at, u.created_by, u.updated_at, u.updated_by,
			u.deleted_at, u.deleted_by, u.version
	%s
		ORDER BY u.username ASC
		LIMIT $3 OFFSET $4
	`, baseQuery)

	var users []*model.User
	var totalCount int64

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		exists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !exists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Executar consulta de contagem
		err = tx.QueryRow(ctx, countQuery, roleID, tenantID).Scan(&totalCount)
		if err != nil {
			return fmt.Errorf("erro ao contar usuários da função: %w", err)
		}

		// Executar consulta principal com paginação
		rows, err := tx.Query(ctx, query, roleID, tenantID, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
		if err != nil {
			return fmt.Errorf("erro ao listar usuários da função: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		users = make([]*model.User, 0)
		for rows.Next() {
			user, err := r.scanUserFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar usuário: %w", err)
			}
			users = append(users, user)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar usuários: %w", rows.Err())
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, 0, err
	}

	return users, totalCount, nil
}

// GetUserRoles retorna as funções associadas a um usuário
func (r *RoleRepository) GetUserRoles(ctx context.Context, tenantID, userID uuid.UUID, includeExpired bool, pagination model.Pagination) ([]*model.Role, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetUserRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("include_expired", includeExpired),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	// Construção dinâmica da consulta com base em includeExpired
	var filterClause string
	if !includeExpired {
		filterClause = "AND (ur.expires_at IS NULL OR ur.expires_at > NOW())"
	}

	// Consulta base para contagem e obtenção de funções
	baseQuery := fmt.Sprintf(`
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		AND ur.tenant_id = $2
		AND r.tenant_id = $2
		AND r.deleted_at IS NULL
		%s
	`, filterClause)

	// Consulta de contagem
	countQuery := "SELECT COUNT(r.id) " + baseQuery

	// Consulta principal com ordenação e paginação
	query := fmt.Sprintf(`
		SELECT 
			r.id, r.tenant_id, r.code, r.name, r.description, 
			r.type, r.is_system, r.is_active, r.metadata, 
			r.created_at, r.created_by, r.updated_at, r.updated_by,
			r.deleted_at, r.deleted_by, r.version,
			ur.expires_at
	%s
		ORDER BY r.name ASC
		LIMIT $3 OFFSET $4
	`, baseQuery)

	var roles []*model.Role
	var totalCount int64

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se o usuário existe
		exists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !exists {
			return model.NewUserNotFoundError(userID)
		}

		// Executar consulta de contagem
		err = tx.QueryRow(ctx, countQuery, userID, tenantID).Scan(&totalCount)
		if err != nil {
			return fmt.Errorf("erro ao contar funções do usuário: %w", err)
		}

		// Executar consulta principal com paginação
		rows, err := tx.Query(ctx, query, userID, tenantID, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
		if err != nil {
			return fmt.Errorf("erro ao listar funções do usuário: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, expiresAt, err := r.scanRoleWithExpirationFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função: %w", err)
			}
			
			// Adicionar informações de expiração aos metadados da função para fins de apresentação
			if expiresAt != nil {
				if role.Metadata == nil {
					role.Metadata = make(map[string]interface{})
				}
				role.Metadata["assignment_expires_at"] = expiresAt.Format(time.RFC3339)
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

// GetAllUserRoles retorna todas as funções associadas a um usuário, incluindo as herdadas
func (r *RoleRepository) GetAllUserRoles(ctx context.Context, tenantID, userID uuid.UUID, includeExpired bool) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetAllUserRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("include_expired", includeExpired),
	)

	// Construção dinâmica da consulta com base em includeExpired
	var filterClause string
	if !includeExpired {
		filterClause = "AND (ur.expires_at IS NULL OR ur.expires_at > NOW())"
	}

	// Consulta recursiva para obter funções diretas e ancestrais
	query := fmt.Sprintf(`
		WITH RECURSIVE user_role_tree AS (
			-- Base case: direct roles assigned to the user
			SELECT 
				r.id, r.tenant_id, r.code, r.name, r.description, 
				r.type, r.is_system, r.is_active, r.metadata, 
				r.created_at, r.created_by, r.updated_at, r.updated_by,
				r.deleted_at, r.deleted_by, r.version,
				ur.expires_at
			FROM roles r
			JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1
			AND ur.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
			%s
			
			UNION
			
			-- Recursive case: parent roles (through role hierarchy)
			SELECT 
				r.id, r.tenant_id, r.code, r.name, r.description, 
				r.type, r.is_system, r.is_active, r.metadata, 
				r.created_at, r.created_by, r.updated_at, r.updated_by,
				r.deleted_at, r.deleted_by, r.version,
				urt.expires_at
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			JOIN user_role_tree urt ON rh.child_role_id = urt.id
			WHERE rh.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
		)
		SELECT DISTINCT * FROM user_role_tree
		ORDER BY name ASC
	`, filterClause)

	var roles []*model.Role

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se o usuário existe
		exists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !exists {
			return model.NewUserNotFoundError(userID)
		}

		// Executar consulta recursiva
		rows, err := tx.Query(ctx, query, userID, tenantID)
		if err != nil {
			return fmt.Errorf("erro ao obter todas as funções do usuário: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, expiresAt, err := r.scanRoleWithExpirationFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função: %w", err)
			}
			
			// Adicionar informações de expiração aos metadados da função
			if expiresAt != nil {
				if role.Metadata == nil {
					role.Metadata = make(map[string]interface{})
				}
				role.Metadata["assignment_expires_at"] = expiresAt.Format(time.RFC3339)
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
		return nil, err
	}

	return roles, nil
}

// IsUserInRole verifica se um usuário está associado a uma função (direta ou indiretamente)
func (r *RoleRepository) IsUserInRole(ctx context.Context, tenantID, userID, roleID uuid.UUID, includeExpired bool) (bool, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.IsUserInRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("include_expired", includeExpired),
	)

	var isInRole bool

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se o usuário existe
		userExists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !userExists {
			return model.NewUserNotFoundError(userID)
		}

		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Construção dinâmica da consulta com base em includeExpired
		var filterClause string
		if !includeExpired {
			filterClause = "AND (ur.expires_at IS NULL OR ur.expires_at > NOW())"
		}

		// Consulta recursiva para verificar associação direta ou indireta
		query := fmt.Sprintf(`
			WITH RECURSIVE user_role_hierarchy AS (
				-- Base case: direct roles of the user
				SELECT r.id
				FROM roles r
				JOIN user_roles ur ON r.id = ur.role_id
				WHERE ur.user_id = $1
				AND ur.tenant_id = $3
				AND r.tenant_id = $3
				AND r.deleted_at IS NULL
				%s
				
				UNION
				
				-- Recursive case: parent roles
				SELECT r.id
				FROM roles r
				JOIN role_hierarchy rh ON r.id = rh.parent_role_id
				JOIN user_role_hierarchy urh ON rh.child_role_id = urh.id
				WHERE rh.tenant_id = $3
				AND r.tenant_id = $3
				AND r.deleted_at IS NULL
			)
			SELECT EXISTS (
				SELECT 1 FROM user_role_hierarchy
				WHERE id = $2
			)
		`, filterClause)

		err = tx.QueryRow(ctx, query, userID, roleID, tenantID).Scan(&isInRole)
		if err != nil {
			return fmt.Errorf("erro ao verificar se usuário está na função: %w", err)
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}

	return isInRole, nil
}

// HasUserDirectRole verifica se um usuário está diretamente associado a uma função
func (r *RoleRepository) HasUserDirectRole(ctx context.Context, tenantID, userID, roleID uuid.UUID, includeExpired bool) (bool, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.HasUserDirectRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("include_expired", includeExpired),
	)

	var hasDirectRole bool

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se o usuário existe
		userExists, err := r.userExists(ctx, tx, tenantID, userID)
		if err != nil {
			return err
		}
		if !userExists {
			return model.NewUserNotFoundError(userID)
		}

		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Verificar associação direta
		var query string
		if includeExpired {
			query = `
				SELECT EXISTS (
					SELECT 1 FROM user_roles
					WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3
				)
			`
		} else {
			query = `
				SELECT EXISTS (
					SELECT 1 FROM user_roles
					WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3
					AND (expires_at IS NULL OR expires_at > NOW())
				)
			`
		}

		err = tx.QueryRow(ctx, query, userID, roleID, tenantID).Scan(&hasDirectRole)
		if err != nil {
			return fmt.Errorf("erro ao verificar associação direta de função: %w", err)
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}

	return hasDirectRole, nil
}

// userExists verifica se um usuário existe
func (r *RoleRepository) userExists(ctx context.Context, tx pgx.Tx, tenantID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`, userID, tenantID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("erro ao verificar existência de usuário: %w", err)
	}

	return exists, nil
}

// scanUserFromRows é uma função auxiliar para ler User a partir de resultados de consulta
func (r *RoleRepository) scanUserFromRows(ctx context.Context, rows pgx.Rows) (*model.User, error) {
	// Esta é uma implementação simplificada e precisaria ser expandida com base no modelo User
	// Seria similar ao scanRoleFromRows, mas adaptado para os campos da entidade User

	// Por enquanto, retornamos um erro para indicar que a implementação precisa ser completada
	return nil, fmt.Errorf("método scanUserFromRows precisa ser implementado")
}

// scanRoleWithExpirationFromRows é uma função auxiliar para ler Role e data de expiração a partir de resultados de consulta
func (r *RoleRepository) scanRoleWithExpirationFromRows(ctx context.Context, rows pgx.Rows) (*model.Role, *time.Time, error) {
	// Esta função é similar a scanRoleFromRows, mas inclui a leitura da coluna de expiração
	// Por enquanto, retornamos um erro para indicar que a implementação precisa ser completada
	return nil, nil, fmt.Errorf("método scanRoleWithExpirationFromRows precisa ser implementado")
}