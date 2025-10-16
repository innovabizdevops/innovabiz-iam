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

// AssignPermission associa uma permissão a uma função
func (r *RoleRepository) AssignPermission(ctx context.Context, tenantID, roleID, permissionID, assignedBy uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.AssignPermission")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
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

		// Verificar se a permissão existe
		permissionExists, err := r.permissionExists(ctx, tx, tenantID, permissionID)
		if err != nil {
			return err
		}
		if !permissionExists {
			return model.NewPermissionNotFoundError(permissionID)
		}

		// Verificar se a permissão já está atribuída
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM role_permissions
				WHERE role_id = $1 AND permission_id = $2 AND tenant_id = $3
			)
		`, roleID, permissionID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar atribuição de permissão existente: %w", err)
		}

		if relationExists {
			return model.NewPermissionAlreadyAssignedError(roleID, permissionID)
		}

		// Associar permissão à função
		_, err = tx.Exec(ctx, `
			INSERT INTO role_permissions (
				tenant_id, role_id, permission_id, 
				created_at, created_by
			) VALUES (
				$1, $2, $3, NOW(), $4
			)
		`, tenantID, roleID, permissionID, assignedBy)

		if err != nil {
			return fmt.Errorf("erro ao atribuir permissão: %w", err)
		}

		// Registrar evento de auditoria
		_, err = tx.Exec(ctx, `
			INSERT INTO role_permission_audit (
				tenant_id, role_id, permission_id,
				action, action_at, action_by
			) VALUES (
				$1, $2, $3, 'ASSIGN', NOW(), $4
			)
		`, tenantID, roleID, permissionID, assignedBy)

		if err != nil {
			return fmt.Errorf("erro ao registrar auditoria de atribuição de permissão: %w", err)
		}

		return nil
	})
}

// RevokePermission remove uma permissão de uma função
func (r *RoleRepository) RevokePermission(ctx context.Context, tenantID, roleID, permissionID, revokedBy uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.RevokePermission")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
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

		// Verificar se a permissão existe
		permissionExists, err := r.permissionExists(ctx, tx, tenantID, permissionID)
		if err != nil {
			return err
		}
		if !permissionExists {
			return model.NewPermissionNotFoundError(permissionID)
		}

		// Verificar se a permissão está atribuída
		var relationExists bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM role_permissions
				WHERE role_id = $1 AND permission_id = $2 AND tenant_id = $3
			)
		`, roleID, permissionID, tenantID).Scan(&relationExists)

		if err != nil {
			return fmt.Errorf("erro ao verificar atribuição de permissão: %w", err)
		}

		if !relationExists {
			return model.NewPermissionNotAssignedError(roleID, permissionID)
		}

		// Remover permissão da função
		_, err = tx.Exec(ctx, `
			DELETE FROM role_permissions
			WHERE role_id = $1 AND permission_id = $2 AND tenant_id = $3
		`, roleID, permissionID, tenantID)

		if err != nil {
			return fmt.Errorf("erro ao revogar permissão: %w", err)
		}

		// Registrar evento de auditoria
		_, err = tx.Exec(ctx, `
			INSERT INTO role_permission_audit (
				tenant_id, role_id, permission_id,
				action, action_at, action_by
			) VALUES (
				$1, $2, $3, 'REVOKE', NOW(), $4
			)
		`, tenantID, roleID, permissionID, revokedBy)

		if err != nil {
			return fmt.Errorf("erro ao registrar auditoria de revogação de permissão: %w", err)
		}

		return nil
	})
}

// GetRolePermissions retorna as permissões atribuídas a uma função
func (r *RoleRepository) GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID, pagination model.Pagination) ([]*model.Permission, int64, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetRolePermissions")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("pagination.page", pagination.Page),
		attribute.Int("pagination.pageSize", pagination.PageSize),
	)

	// Consulta base para contagem e obtenção de permissões
	baseQuery := `
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		AND rp.tenant_id = $2
		AND p.tenant_id = $2
		AND p.deleted_at IS NULL
	`

	// Consulta de contagem
	countQuery := "SELECT COUNT(p.id) " + baseQuery

	// Consulta principal com ordenação e paginação
	query := `
		SELECT 
			p.id, p.tenant_id, p.code, p.name, p.description, 
			p.category, p.is_system, p.is_active, p.metadata, 
			p.created_at, p.created_by, p.updated_at, p.updated_by,
			p.deleted_at, p.deleted_by, p.version
	` + baseQuery + `
		ORDER BY p.name ASC
		LIMIT $3 OFFSET $4
	`

	var permissions []*model.Permission
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
			return fmt.Errorf("erro ao contar permissões: %w", err)
		}

		// Executar consulta principal com paginação
		rows, err := tx.Query(ctx, query, roleID, tenantID, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)
		if err != nil {
			return fmt.Errorf("erro ao listar permissões: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		permissions = make([]*model.Permission, 0)
		for rows.Next() {
			permission, err := r.scanPermissionFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar permissão: %w", err)
			}
			permissions = append(permissions, permission)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar permissões: %w", rows.Err())
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, 0, err
	}

	return permissions, totalCount, nil
}

// GetAllPermissionsForRole retorna todas as permissões de uma função, incluindo as herdadas
func (r *RoleRepository) GetAllPermissionsForRole(ctx context.Context, tenantID, roleID uuid.UUID) ([]*model.Permission, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.GetAllPermissionsForRole")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Consulta recursiva para obter permissões diretas e herdadas
	query := `
		WITH RECURSIVE role_tree AS (
			-- Base case: the role itself
			SELECT id, id as origin_role_id
			FROM roles
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			
			UNION
			
			-- Recursive case: all ancestor roles
			SELECT r.id, rt.origin_role_id
			FROM roles r
			JOIN role_hierarchy rh ON r.id = rh.parent_role_id
			JOIN role_tree rt ON rh.child_role_id = rt.id
			WHERE rh.tenant_id = $2
			AND r.tenant_id = $2
			AND r.deleted_at IS NULL
		)
		SELECT DISTINCT
			p.id, p.tenant_id, p.code, p.name, p.description, 
			p.category, p.is_system, p.is_active, p.metadata, 
			p.created_at, p.created_by, p.updated_at, p.updated_by,
			p.deleted_at, p.deleted_by, p.version
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN role_tree rt ON rp.role_id = rt.id
		WHERE rp.tenant_id = $2
		AND p.tenant_id = $2
		AND p.deleted_at IS NULL
		ORDER BY p.name ASC
	`

	var permissions []*model.Permission

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
			return fmt.Errorf("erro ao obter permissões da função e ancestrais: %w", err)
		}
		defer rows.Close()

		// Processar resultados
		permissions = make([]*model.Permission, 0)
		for rows.Next() {
			permission, err := r.scanPermissionFromRows(ctx, rows)
			if err != nil {
				return fmt.Errorf("erro ao processar permissão: %w", err)
			}
			permissions = append(permissions, permission)
		}

		if rows.Err() != nil {
			return fmt.Errorf("erro ao iterar permissões: %w", rows.Err())
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return permissions, nil
}

// HasPermission verifica se uma função tem uma determinada permissão (direta ou herdada)
func (r *RoleRepository) HasPermission(ctx context.Context, tenantID, roleID, permissionID uuid.UUID) (bool, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.HasPermission")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	var hasPermission bool

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Verificar se a permissão existe
		permissionExists, err := r.permissionExists(ctx, tx, tenantID, permissionID)
		if err != nil {
			return err
		}
		if !permissionExists {
			return model.NewPermissionNotFoundError(permissionID)
		}

		// Consulta recursiva para verificar permissão direta ou herdada
		query := `
			WITH RECURSIVE role_tree AS (
				-- Base case: the role itself
				SELECT id FROM roles
				WHERE id = $1 AND tenant_id = $3 AND deleted_at IS NULL
				
				UNION
				
				-- Recursive case: all ancestor roles
				SELECT r.id
				FROM roles r
				JOIN role_hierarchy rh ON r.id = rh.parent_role_id
				JOIN role_tree rt ON rh.child_role_id = rt.id
				WHERE rh.tenant_id = $3
				AND r.tenant_id = $3
				AND r.deleted_at IS NULL
			)
			SELECT EXISTS (
				SELECT 1
				FROM role_permissions rp
				JOIN role_tree rt ON rp.role_id = rt.id
				WHERE rp.permission_id = $2
				AND rp.tenant_id = $3
			)
		`

		err = tx.QueryRow(ctx, query, roleID, permissionID, tenantID).Scan(&hasPermission)
		if err != nil {
			return fmt.Errorf("erro ao verificar permissão: %w", err)
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}

	return hasPermission, nil
}

// HasDirectPermission verifica se uma função tem uma permissão diretamente atribuída
func (r *RoleRepository) HasDirectPermission(ctx context.Context, tenantID, roleID, permissionID uuid.UUID) (bool, error) {
	ctx, span := tracer.Start(ctx, "RoleRepository.HasDirectPermission")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", roleID.String()),
		attribute.String("permission.id", permissionID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	var hasPermission bool

	err := r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Verificar se a função existe
		roleExists, err := r.roleExists(ctx, tx, tenantID, roleID)
		if err != nil {
			return err
		}
		if !roleExists {
			return model.NewRoleNotFoundError(roleID)
		}

		// Verificar se a permissão existe
		permissionExists, err := r.permissionExists(ctx, tx, tenantID, permissionID)
		if err != nil {
			return err
		}
		if !permissionExists {
			return model.NewPermissionNotFoundError(permissionID)
		}

		// Verificar permissão direta
		query := `
			SELECT EXISTS (
				SELECT 1 FROM role_permissions
				WHERE role_id = $1 AND permission_id = $2 AND tenant_id = $3
			)
		`

		err = tx.QueryRow(ctx, query, roleID, permissionID, tenantID).Scan(&hasPermission)
		if err != nil {
			return fmt.Errorf("erro ao verificar permissão direta: %w", err)
		}

		return nil
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return false, err
	}

	return hasPermission, nil
}

// permissionExists verifica se uma permissão existe
func (r *RoleRepository) permissionExists(ctx context.Context, tx pgx.Tx, tenantID, permissionID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM permissions
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		)
	`, permissionID, tenantID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("erro ao verificar existência de permissão: %w", err)
	}

	return exists, nil
}

// scanPermissionFromRows é uma função auxiliar para ler Permission a partir de resultados de consulta
func (r *RoleRepository) scanPermissionFromRows(ctx context.Context, rows pgx.Rows) (*model.Permission, error) {
	// Esta é uma implementação simplificada e precisaria ser expandida com base no modelo Permission
	// Seria similar ao scanRoleFromRows, mas adaptado para os campos da entidade Permission

	// Por enquanto, retornamos um erro para indicar que a implementação precisa ser completada
	return nil, fmt.Errorf("método scanPermissionFromRows precisa ser implementado")

	// A implementação completa depende da estrutura exata da tabela permissions
	// e da implementação do modelo Permission
}