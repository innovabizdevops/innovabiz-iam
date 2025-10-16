package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/domain/model"
)

// AssignChildRole adiciona uma função filha a uma função pai
func (r *RoleRepository) AssignChildRole(ctx context.Context, tenantID, parentRoleID, childRoleID, assignedBy uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.AssignChildRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("parent_role.id", parentRoleID.String()),
		attribute.String("child_role.id", childRoleID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função pai existe
		parentExists, err := r.roleExists(ctx, tx, tenantID, parentRoleID)
		if err != nil {
			return err
		}
		if !parentExists {
			return model.NewParentRoleNotFoundError(parentRoleID)
		}

		// Verificar se a função filha existe
		childExists, err := r.roleExists(ctx, tx, tenantID, childRoleID)
		if err != nil {
			return err
		}
		if !childExists {
			return model.NewChildRoleNotFoundError(childRoleID)
		}

		// Verificar se a relação já existe
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM role_hierarchy
				WHERE parent_role_id = $1 AND child_role_id = $2 AND tenant_id = $3
			)
		`, parentRoleID, childRoleID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar relação hierárquica existente: %w", err)
		}

		if relationExists {
			return model.NewChildRoleAlreadyAssignedError(parentRoleID, childRoleID)
		}

		// Verificar se geraria ciclo na hierarquia
		wouldCreateCycle, err := r.wouldCreateCycle(ctx, tx, tenantID, childRoleID, parentRoleID)
		if err != nil {
			return err
		}

		if wouldCreateCycle {
			return model.NewCyclicRoleHierarchyError(parentRoleID, childRoleID)
		}

		// Verificar tipo compatível
		err = r.verifyRoleTypeCompatibility(ctx, tx, tenantID, parentRoleID, childRoleID)
		if err != nil {
			return err
		}

		// Inserir na hierarquia
		_, err = tx.Exec(ctx, `
			INSERT INTO role_hierarchy (
				tenant_id, parent_role_id, child_role_id, 
				created_at, created_by
			) VALUES (
				$1, $2, $3, NOW(), $4
			)
		`, tenantID, parentRoleID, childRoleID, assignedBy)

		if err != nil {
			return fmt.Errorf("erro ao atribuir função filha: %w", err)
		}

		return nil
	})
}

// RemoveChildRole remove uma função filha de uma função pai
func (r *RoleRepository) RemoveChildRole(ctx context.Context, tenantID, parentRoleID, childRoleID, removedBy uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.RemoveChildRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("parent_role.id", parentRoleID.String()),
		attribute.String("child_role.id", childRoleID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função pai existe
		parentExists, err := r.roleExists(ctx, tx, tenantID, parentRoleID)
		if err != nil {
			return err
		}
		if !parentExists {
			return model.NewParentRoleNotFoundError(parentRoleID)
		}

		// Verificar se a função filha existe
		childExists, err := r.roleExists(ctx, tx, tenantID, childRoleID)
		if err != nil {
			return err
		}
		if !childExists {
			return model.NewChildRoleNotFoundError(childRoleID)
		}

		// Verificar se a relação existe
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM role_hierarchy
				WHERE parent_role_id = $1 AND child_role_id = $2 AND tenant_id = $3
			)
		`, parentRoleID, childRoleID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar relação hierárquica existente: %w", err)
		}

		if !relationExists {
			return model.NewChildRoleNotAssignedError(parentRoleID, childRoleID)
		}

		// Remover da hierarquia
		_, err = tx.Exec(ctx, `
			DELETE FROM role_hierarchy
			WHERE parent_role_id = $1 AND child_role_id = $2 AND tenant_id = $3
		`, parentRoleID, childRoleID, tenantID)

		if err != nil {
			return fmt.Errorf("erro ao remover função filha: %w", err)
		}

		// Registrar o evento de remoção na tabela de auditoria
		_, err = tx.Exec(ctx, `
			INSERT INTO role_hierarchy_audit (
				tenant_id, parent_role_id, child_role_id,
				action, action_at, action_by
			) VALUES (
				$1, $2, $3, 'REMOVE', NOW(), $4
			)
		`, tenantID, parentRoleID, childRoleID, removedBy)

		if err != nil {
			return fmt.Errorf("erro ao registrar auditoria de remoção de função filha: %w", err)
		}

		return nil
	})
}

// GetChildRoles retorna as funções filhas de uma função
func (r *RoleRepository) GetChildRoles(ctx context.Context, tenantID, roleID uuid.UUID, pagination model.Pagination) ([]*model.Role, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetChildRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	// Consulta base para contagem e obtenção de funções filhas
	baseQuery := `
		FROM roles r
		JOIN role_hierarchy rh ON r.id = rh.child_role_id
		WHERE rh.parent_role_id = $1
		AND rh.tenant_id = $2
		AND r.tenant_id = $2
		AND r.deleted_at IS NULL
	`

	// Consulta de contagem
	countQuery := "SELECT COUNT(r.id) " + baseQuery

	// Consulta principal com ordenação e paginação
	query := `
		SELECT 
			r.id, r.tenant_id, r.code, r.name, r.description, 
			r.type, r.is_system, r.is_active, r.metadata, 
			r.created_at, r.created_by, r.updated_at, r.updated_by,
			r.deleted_at, r.deleted_by, r.version
	` + baseQuery + `
		ORDER BY r.name ASC
		LIMIT $3 OFFSET $4
	`

	var roles []*model.Role
	var totalCount int64

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função pai existe
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
			return fmt.Errorf("erro ao contar funções filhas: %w", err)
		}

		// Executar consulta principal com paginação
		rows, err := tx.Query(ctx, query, roleID, tenantID, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
		if err != nil {
			return fmt.Errorf("erro ao listar funções filhas: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, err := r.scanRoleFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função filha: %w", err)
			}
			roles = append(roles, role)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar funções filhas: %w", rows.Err())
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

// GetParentRoles retorna as funções pais de uma função
func (r *RoleRepository) GetParentRoles(ctx context.Context, tenantID, roleID uuid.UUID, pagination model.Pagination) ([]*model.Role, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetParentRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	// Consulta base para contagem e obtenção de funções pais
	baseQuery := `
		FROM roles r
		JOIN role_hierarchy rh ON r.id = rh.parent_role_id
		WHERE rh.child_role_id = $1
		AND rh.tenant_id = $2
		AND r.tenant_id = $2
		AND r.deleted_at IS NULL
	`

	// Consulta de contagem
	countQuery := "SELECT COUNT(r.id) " + baseQuery

	// Consulta principal com ordenação e paginação
	query := `
		SELECT 
			r.id, r.tenant_id, r.code, r.name, r.description, 
			r.type, r.is_system, r.is_active, r.metadata, 
			r.created_at, r.created_by, r.updated_at, r.updated_by,
			r.deleted_at, r.deleted_by, r.version
	` + baseQuery + `
		ORDER BY r.name ASC
		LIMIT $3 OFFSET $4
	`

	var roles []*model.Role
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
			return fmt.Errorf("erro ao contar funções pais: %w", err)
		}

		// Executar consulta principal com paginação
		rows, err := tx.Query(ctx, query, roleID, tenantID, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
		if err != nil {
			return fmt.Errorf("erro ao listar funções pais: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, err := r.scanRoleFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função pai: %w", err)
			}
			roles = append(roles, role)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar funções pais: %w", rows.Err())
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

// GetAncestorRoles retorna todas as funções ancestrais (recursivamente) de uma função
func (r *RoleRepository) GetAncestorRoles(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetAncestorRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Consulta recursiva para obter todos os ancestrais
	query := `
		WITH RECURSIVE role_ancestors AS (
			-- Base case: direct parents
			SELECT 
				r.id, r.tenant_id, r.code, r.name, r.description, 
				r.type, r.is_system, r.is_active, r.metadata, 
				r.created_at, r.created_by, r.updated_at, r.updated_by,
				r.deleted_at, r.deleted_by, r.version
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			WHERE rh.child_role_id = $1
			AND rh.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
			
			UNION
			
			-- Recursive case: parents of parents
			SELECT 
				r.id, r.tenant_id, r.code, r.name, r.description, 
				r.type, r.is_system, r.is_active, r.metadata, 
				r.created_at, r.created_by, r.updated_at, r.updated_by,
				r.deleted_at, r.deleted_by, r.version
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			JOIN role_ancestors ra ON rh.child_role_id = ra.id
			WHERE rh.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
		)
		SELECT * FROM role_ancestors
		ORDER BY name ASC
	`

	var roles []*model.Role

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		exists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !exists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Executar consulta recursiva
		rows, err := tx.Query(ctx, query, roleID, tenantID)
		if err != nil {
			return fmt.Errorf("erro ao obter funções ancestrais: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, err := r.scanRoleFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função ancestral: %w", err)
			}
			roles = append(roles, role)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar funções ancestrais: %w", rows.Err())
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

// GetDescendantRoles retorna todas as funções descendentes (recursivamente) de uma função
func (r *RoleRepository) GetDescendantRoles(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Role, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetDescendantRoles")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Consulta recursiva para obter todos os descendentes
	query := `
		WITH RECURSIVE role_descendants AS (
			-- Base case: direct children
			SELECT 
				r.id, r.tenant_id, r.code, r.name, r.description, 
				r.type, r.is_system, r.is_active, r.metadata, 
				r.created_at, r.created_by, r.updated_at, r.updated_by,
				r.deleted_at, r.deleted_by, r.version
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.child_role_id
			WHERE rh.parent_role_id = $1
			AND rh.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
			
			UNION
			
			-- Recursive case: children of children
			SELECT 
				r.id, r.tenant_id, r.code, r.name, r.description, 
				r.type, r.is_system, r.is_active, r.metadata, 
				r.created_at, r.created_by, r.updated_at, r.updated_by,
				r.deleted_at, r.deleted_by, r.version
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.child_role_id
			JOIN role_descendants rd ON rh.parent_role_id = rd.id
			WHERE rh.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
		)
		SELECT * FROM role_descendants
		ORDER BY name ASC
	`

	var roles []*model.Role

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		exists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !exists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Executar consulta recursiva
		rows, err := tx.Query(ctx, query, roleID, tenantID)
		if err != nil {
			return fmt.Errorf("erro ao obter funções descendentes: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		roles = make([]*model.Role, 0)
		for rows.Next() {
			role, err := r.scanRoleFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar função descendente: %w", err)
			}
			roles = append(roles, role)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar funções descendentes: %w", rows.Err())
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

// Métodos auxiliares

// roleExists verifica se uma função existe
func (r *RoleRepository) roleExists(ctx context.Context, tx pgx.Tx, tenantID, roleID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM roles
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`, roleID, tenantID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("erro ao verificar existência de função: %w", err)
	}

	return exists, nil
}

// wouldCreateCycle verifica se adicionar childID como filho de parentID criaria um ciclo
func (r *RoleRepository) wouldCreateCycle(ctx context.Context, tx pgx.Tx, tenantID, childID, parentID uuid.UUID) (bool, error) {
	// Verificar se childID é igual a parentID (auto-referência)
	if childID == parentID {
		return true, nil
	}

	// Verificar se parentID já é descendente de childID (o que criaria um ciclo)
	query := `
		WITH RECURSIVE role_descendants AS (
			-- Base case: direct children of childID
			SELECT child_role_id
			FROM role_hierarchy
			WHERE parent_role_id = $1 AND tenant_id = $2
			
			UNION
			
			-- Recursive case: children of children
			SELECT rh.child_role_id
			FROM role_hierarchy rh
			JOIN role_descendants rd ON rh.parent_role_id = rd.child_role_id
			WHERE rh.tenant_id = $2
		)
		SELECT EXISTS (
			SELECT 1 FROM role_descendants
			WHERE child_role_id = $3
		)
	`

	var wouldCreateCycle bool
	err := tx.QueryRow(ctx, query, childID, tenantID, parentID).Scan(&wouldCreateCycle)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar ciclo na hierarquia: %w", err)
	}

	return wouldCreateCycle, nil
}

// verifyRoleTypeCompatibility verifica se os tipos de função são compatíveis para hierarquia
func (r *RoleRepository) verifyRoleTypeCompatibility(ctx context.Context, tx pgx.Tx, tenantID, parentRoleID, childRoleID uuid.UUID) error {
	var parentType, childType string
	
	// Obter tipo da função pai
	err := tx.QueryRow(ctx, `
		SELECT type FROM roles
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, parentRoleID, tenantID).Scan(&parentType)
	
	if err != nil {
		return fmt.Errorf("erro ao obter tipo da função pai: %w", err)
	}
	
	// Obter tipo da função filha
	err = tx.QueryRow(ctx, `
		SELECT type FROM roles
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, childRoleID, tenantID).Scan(&childType)
	
	if err != nil {
		return fmt.Errorf("erro ao obter tipo da função filha: %w", err)
	}
	
	// Verificar compatibilidade dos tipos
	// Regras de compatibilidade para hierarquia:
	// - Funções do mesmo tipo podem ter relação hierárquica
	// - Funções do tipo "SYSTEM" podem ser pais de funções do tipo "USER"
	// - Outras combinações não são permitidas
	if parentType != childType {
		if parentType != model.RoleTypeSystem && childType != model.RoleTypeUser {
			return model.NewRolesTypeMismatchError(parentRoleID, parentType, childRoleID, childType)
		}
	}
	
	return nil
}