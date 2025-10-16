/**
 * INNOVABIZ IAM - Registro de Resolvers de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Este arquivo é responsável pelo registro dos resolvers de grupos no schema GraphQL
 * principal da plataforma INNOVABIZ. Implementa a integração total com o módulo IAM
 * seguindo a arquitetura multi-dimensional, multi-tenant e multi-contextual.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR/PDPA/CCPA (Proteção de dados)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/domain/role"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/infrastructure/authorization"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/interfaces/graphql/resolvers"
	"go.opentelemetry.io/otel/trace"
)

// RegisterGroupResolvers registra todos os resolvers relacionados a grupos no schema GraphQL
// e carrega as definições do arquivo group.graphql
func RegisterGroupResolvers(
	schema *graphql.Schema,
	groupService group.Service,
	userService user.Service,
	roleService role.Service,
	authz authorization.Service,
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer trace.Tracer,
) error {
	// Criar o resolver para operações de grupo
	groupResolver := resolvers.NewGroupResolver(
		groupService,
		userService,
		roleService,
		authz,
		logger,
		metrics,
		tracer,
	)

	// Registrar os resolvers no schema
	resolvers.RegisterGroupResolvers(schema, groupResolver)

	logger.Info(nil, "Resolvers de grupo registrados com sucesso no schema GraphQL", logging.Fields{
		"component": "GraphQL",
		"module":    "IAM",
		"service":   "identity",
	})

	return nil
}