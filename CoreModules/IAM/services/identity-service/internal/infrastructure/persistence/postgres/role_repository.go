package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"innovabiz/iam/identity-service/internal/domain/model"
	"innovabiz/iam/identity-service/internal/domain/repository"
)

// RoleRepository implementa a interface repository.RoleRepository usando PostgreSQL
type RoleRepository struct {
	db *DB
}

// NewRoleRepository cria uma nova instância do RoleRepository
func NewRoleRepository(db *DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create insere uma nova função no banco de dados
func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", role.ID().String()),
		attribute.String("role.code", role.Code()),
		attribute.String("tenant.id", role.TenantID().String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Inserir na tabela principal de funções
		query := `
			INSERT INTO roles (
				id, tenant_id, code, name, description, 
				type, is_system, is_active, metadata, 
				created_at, created_by, updated_at, updated_by,
				version
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12, $13,
				$14
			)
		`

		_, err := tx.Exec(ctx,
			query,
			role.ID(),
			role.TenantID(),
			role.Code(),
			role.Name(),
			role.Description(),
			role.Type(),
			role.IsSystem(),
			role.IsActive(),
			role.Metadata(),
			role.CreatedAt(),
			role.CreatedBy(),
			role.UpdatedAt(),
			role.UpdatedBy(),
			role.Version(),
		)

		if err != nil {
			// Verificar se é um erro de violação de constraint
			if pgErr, ok := err.(*pgconn.PgError); ok {
				if pgErr.Code == pgerrcode.UniqueViolation {
					if strings.Contains(pgErr.Message, "roles_tenant_id_code_key") {
						return model.NewRoleCodeAlreadyExistsError(role.Code(), role.TenantID())
					}
				}
			}
			return fmt.Errorf("erro ao criar função no banco de dados: %w", err)
		}

		return nil
	})
}

// Update atualiza uma função existente no banco de dados
func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	ctx, span := tracer.Start(ctx, "RoleRepository.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("role.id", role.ID().String()),
		attribute.String("role.code", role.Code()),
		attribute.String("tenant.id", role.TenantID().String()),
	)

	return r.db.InTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Atualizar função com controle de versão otimista
		query := `
			UPDATE roles SET 
				code = $1,
				name = $2,
				description = $3,
				type = $4,
				is_system = $5,
				is_active = $6,
				metadata = $7,
				updated_at = $8,
				updated_by = $9,
				version = version + 1
			WHERE id = $10 AND tenant_id = $11 AND version = $12
		`

		result, err := tx.Exec(ctx,
			query,
			role.Code(),
			role.Name(),
			role.Description(),
			role.Type(),
			role.IsSystem(),
			role.IsActive(),
			role.Metadata(),
			role.UpdatedAt(),
			role.UpdatedBy(),
			role.ID(),
			role.TenantID(),
			role.Version(),
		)

		if err != nil {
			// Verificar se é um erro de violação de constraint
			if pgErr, ok := err.(*pgconn.PgError); ok {
				if pgErr.Code == pgerrcode.UniqueViolation {
					if strings.Contains(pgErr.Message, "roles_tenant_id_code_key") {
						return model.NewRoleCodeAlreadyExistsError(role.Code(), role.TenantID())
					}
				}
			}
			return fmt.Errorf("erro ao atualizar função no banco de dados: %w", err)
		}

		// Verificar se a função foi encontrada e corresponde à versão esperada
		if result.RowsAffected() == 0 {
			// Verificar se a função existe
			var exists bool
			err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1 AND tenant_id = $2)", role.ID(), role.TenantID()).Scan(&exists)
			
			if err != nil {
				return fmt.Errorf("erro ao verificar existência de função: %w", err)
			}
			
			if !exists {
				return model.NewRoleNotFoundError(role.ID())
			}
			
			// A função existe, mas a versão não corresponde
			return repository.ErrConcurrentModification
		}

		// Atualizar role.Version() com o novo valor de versão
		var newVersion int
		err = tx.QueryRow(ctx, "SELECT version FROM roles WHERE id = $1 AND tenant_id = $2", role.ID(), role.TenantID()).Scan(&newVersion)
		if err != nil {
			return fmt.Errorf("erro ao obter nova versão da função: %w", err)
		}
		
		// Atualizar o objeto do domínio com a nova versão
		role.SetVersion(newVersion)
		
		return nil
	})
}