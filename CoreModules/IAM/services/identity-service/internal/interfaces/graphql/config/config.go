/**
 * INNOVABIZ IAM - Configuração GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação da configuração do servidor GraphQL para o módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 */

package config

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/innovabiz/iam/internal/domain/services"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
	"github.com/innovabiz/iam/internal/interfaces/graphql/generated"
	"github.com/innovabiz/iam/internal/interfaces/graphql/resolvers"
	"github.com/innovabiz/iam/internal/interfaces/graphql/scalars"
)

// Config representa a configuração para a API GraphQL
type Config struct {
	Logger       logging.Logger
	Metrics      metrics.MetricsClient
	Tracer       tracing.Tracer
	GroupService services.GroupService
	UserService  services.UserService
}

// NewGraphQLServer cria um novo servidor GraphQL
func NewGraphQLServer(cfg Config) *handler.Server {
	// Configurar os resolvers
	resolver := &resolvers.Resolver{
		Logger:       cfg.Logger,
		Metrics:      cfg.Metrics,
		Tracer:       cfg.Tracer,
		GroupService: cfg.GroupService,
		UserService:  cfg.UserService,
	}

	// Configurar o servidor GraphQL
	config := generated.Config{
		Resolvers: resolver,
	}

	// Adicionar scalars personalizados
	config.Directives.Auth = func(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
		// Placeholder para implementação futura de autenticação e autorização
		// Em uma implementação real, verificar as permissões do usuário aqui
		return next(ctx)
	}

	// Adicionar scalars personalizados
	config.Directives.ValidateInput = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		// Placeholder para implementação futura de validação de entrada
		// Em uma implementação real, validar entrada aqui
		return next(ctx)
	}

	// Registrar os scalars customizados
	config.Resolvers = resolver
	server := handler.NewDefaultServer(generated.NewExecutableSchema(config))

	// Adicionar middlewares para observabilidade
	server.Use(tracing.NewGraphQLMiddleware(cfg.Tracer))
	server.Use(metrics.NewGraphQLMiddleware(cfg.Metrics))
	server.Use(logging.NewGraphQLMiddleware(cfg.Logger))

	return server
}

// NewPlaygroundHandler cria um handler para o GraphQL Playground
func NewPlaygroundHandler(endpoint string) http.Handler {
	return playground.Handler("INNOVABIZ IAM GraphQL", endpoint)
}

// RegisterScalars registra os scalars personalizados com o GraphQL
func RegisterScalars(schema *graphql.ExecutableSchema) {
	// Registrar o scalar DateTime
	schema.MustParseSchema(schema.Schema(), map[string]interface{}{
		"DateTime":   scalars.DateTimeScalar,
		"JSONObject": scalars.JSONObjectScalar,
	})
}