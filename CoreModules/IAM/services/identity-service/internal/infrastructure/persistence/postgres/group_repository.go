/**
 * INNOVABIZ IAM - Repositório PostgreSQL para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do repositório de grupos usando PostgreSQL no módulo core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant, multi-contextual
 * e com observabilidade total da plataforma INNOVABIZ.
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
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/persistence"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GroupRepository implementa a interface group.Repository para PostgreSQL
type GroupRepository struct {
	db      *sqlx.DB
	logger  logging.Logger
	metrics metrics.MetricsClient
	tracer  trace.Tracer
}

// NewGroupRepository cria uma nova instância do repositório de grupos
func NewGroupRepository(
	db *sqlx.DB,
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer trace.Tracer,
) *GroupRepository {
	return &GroupRepository{
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// GetByID busca um grupo pelo ID
func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*group.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupRepository.GetByID")
	defer span.End()

	// Adicionar atributos ao span para rastreabilidade
	span.SetAttributes(
		attribute.String("group.id", id.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	// Iniciar temporizador para métrica
	timer := r.metrics.Timer("repository.group.getById.duration")
	defer timer.ObserveDuration()

	query := `
		SELECT g.id, g.code, g.name, g.description, g.tenant_id, g.region_code, 
		       g.group_type, g.status, g.path, g.level, g.parent_group_id,
		       g.created_at, g.created_by, g.updated_at, g.updated_by,
		       g.metadata
		FROM iam_groups g
		WHERE g.id = $1 AND g.tenant_id = $2 AND g.deleted_at IS NULL
	`

	// Extrair transação do contexto se existir
	tx := persistence.GetTxFromContext(ctx)

	var row *sqlx.Row
	if tx != nil {
		row = tx.QueryRowxContext(ctx, query, id, tenantID)
	} else {
		row = r.db.QueryRowxContext(ctx, query, id, tenantID)
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
		r.metrics.Counter("repository.group.getById.error").Inc(1)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrGroupNotFound
		}

		r.logger.Error(ctx, "Erro ao buscar grupo por ID", logging.Fields{
			"error":    err.Error(),
			"groupId":  id.String(),
			"tenantId": tenantID.String(),
		})
		return nil, fmt.Errorf("erro ao buscar grupo por ID: %w", err)
	}

	// Processar metadados (JSON)
	if metadata.Valid && metadata.String != "" {
		g.Metadata = make(map[string]interface{})
		// Implementar a conversão de JSON para map aqui
		// json.Unmarshal([]byte(metadata.String), &g.Metadata)
	}

	// Incrementar contador de sucesso
	r.metrics.Counter("repository.group.getById.success").Inc(1)

	return &g, nil
}

// tabela iam_groups estrutura sugerida:
// id uuid PRIMARY KEY,
// code VARCHAR(100) NOT NULL,
// name VARCHAR(200) NOT NULL,
// description TEXT,
// tenant_id uuid NOT NULL,
// region_code VARCHAR(50),
// group_type VARCHAR(50),
// status VARCHAR(20) NOT NULL,
// path VARCHAR(500),
// level INTEGER NOT NULL DEFAULT 0,
// parent_group_id uuid,
// metadata JSONB,
// created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
// created_by uuid,
// updated_at TIMESTAMP WITH TIME ZONE,
// updated_by uuid,
// deleted_at TIMESTAMP WITH TIME ZONE,
// deleted_by uuid,
// CONSTRAINT fk_parent_group FOREIGN KEY (parent_group_id) REFERENCES iam_groups (id),
// CONSTRAINT unique_group_code_tenant UNIQUE (code, tenant_id),
// CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES iam_tenants (id)

// tabela iam_group_members estrutura sugerida:
// group_id uuid NOT NULL,
// user_id uuid NOT NULL,
// tenant_id uuid NOT NULL,
// created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
// created_by uuid,
// PRIMARY KEY (group_id, user_id, tenant_id),
// CONSTRAINT fk_group FOREIGN KEY (group_id) REFERENCES iam_groups (id),
// CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES iam_users (id),
// CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES iam_tenants (id)