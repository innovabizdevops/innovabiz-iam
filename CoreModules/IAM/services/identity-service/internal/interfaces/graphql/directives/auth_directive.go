/**
 * INNOVABIZ IAM - Diretiva de Autenticação GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação da diretiva de autenticação e autorização para GraphQL
 * no módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15, A.8.3 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4, 8.2 - Gestão de identidade)
 * - LGPD/GDPR/PDPA (Arts. 46-48 - Segurança e controle de acesso)
 * - BNA Instrução 7/2021 (Art. 9 - Controle de acesso e autorização)
 * - SOX (Sec. 404 - Controles internos)
 * - NIST CSF (PR.AC - Gerenciamento de identidades e credenciais)
 */

package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/innovabiz/iam/internal/domain/errors"
	"github.com/innovabiz/iam/internal/infrastructure/auth"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
)

// AuthDirective implementa a diretiva @auth para autorização de operações GraphQL
func AuthDirective(
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer tracing.Tracer,
) func(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
		span, ctx := tracer.StartFromContext(ctx, "AuthDirective")
		defer span.End()

		// Adicionar atributos de telemetria
		span.SetAttributes(attribute.StringSlice("required_roles", roles))

		operationName := graphql.GetOperationContext(ctx).OperationName
		if operationName == "" {
			operationName = "unknown"
		}

		logger := logger.WithContext(ctx).
			WithField("directive", "auth").
			WithField("operation", operationName).
			WithField("required_roles", roles)

		logger.Info("Verificando autorização")

		// Obter usuário autenticado do contexto
		user, err := auth.GetUserFromContext(ctx)
		if err != nil {
			logger.WithError(err).Warn("Falha na autenticação: usuário não encontrado no contexto")
			span.RecordError(err)
			metrics.IncAuthErrors("authentication_failed", operationName)
			return nil, errors.NewAuthenticationError("Usuário não autenticado")
		}

		// Verificar se o usuário tem as permissões necessárias
		hasRoles, err := auth.UserHasRoles(ctx, user.ID, roles)
		if err != nil {
			logger.WithError(err).Error("Erro ao verificar permissões do usuário")
			span.RecordError(err)
			metrics.IncAuthErrors("role_check_failed", operationName)
			return nil, fmt.Errorf("erro ao verificar permissões: %w", err)
		}

		if !hasRoles {
			logger.WithField("user_id", user.ID.String()).
				WithField("required_roles", roles).
				Warn("Acesso negado: usuário não possui permissões necessárias")
			metrics.IncAuthErrors("authorization_failed", operationName)
			return nil, errors.NewAuthorizationError("Permissão insuficiente")
		}

		// Adicionar informações do usuário ao contexto para logging e auditoria
		ctx = auth.EnrichContextWithUser(ctx, user)
		
		// Registrar acesso autorizado em métricas
		metrics.IncAuthSuccess(operationName)
		
		logger.WithField("user_id", user.ID.String()).Info("Acesso autorizado")

		// Continuar com a resolução da query/mutation
		return next(ctx)
	}
}

// MockAuthDirective é uma implementação de diretiva de autenticação para ambientes de desenvolvimento e testes
func MockAuthDirective(
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer tracing.Tracer,
) func(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error) {
		span, ctx := tracer.StartFromContext(ctx, "MockAuthDirective")
		defer span.End()

		operationName := graphql.GetOperationContext(ctx).OperationName
		if operationName == "" {
			operationName = "unknown"
		}

		logger := logger.WithContext(ctx).
			WithField("directive", "mock_auth").
			WithField("operation", operationName)

		logger.Info("Usando diretiva de autenticação simulada para ambiente de desenvolvimento")

		// Criar usuário simulado para testes
		mockUser := &auth.User{
			ID:       uuid.New(),
			Username: "test_user",
			Email:    "test@innovabiz.com",
			Roles:    []string{"ADMIN", "USER"},
		}

		// Adicionar usuário simulado ao contexto
		ctx = auth.EnrichContextWithUser(ctx, mockUser)
		
		metrics.IncAuthSuccess(operationName)

		// Continuar com a resolução da query/mutation
		return next(ctx)
	}
}