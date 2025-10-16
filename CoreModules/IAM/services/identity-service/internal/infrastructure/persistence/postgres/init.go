/**
 * INNOVABIZ IAM - Inicialização do Repositório PostgreSQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Inicialização e registro de todos os repositórios PostgreSQL para o
 * módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant 
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
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/container"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
)

// RegisterRepositories registra todos os repositórios PostgreSQL no container de dependências
func RegisterRepositories(ctx context.Context, c *container.Container, db *sqlx.DB) error {
	logger := c.GetLogger()
	metricsClient := c.GetMetricsClient()
	tracer := c.GetTracer()

	// Registrar o repositório de grupos
	if err := registerGroupRepository(ctx, c, db, logger, metricsClient, tracer); err != nil {
		return err
	}

	// Aqui seriam registrados outros repositórios PostgreSQL para o módulo IAM:
	// - Usuários
	// - Papéis
	// - Permissões
	// - Tenants
	// - Regiões
	// etc.

	return nil
}

// registerGroupRepository registra o repositório de grupos no container
func registerGroupRepository(
	ctx context.Context,
	c *container.Container,
	db *sqlx.DB,
	logger logging.Logger,
	metricsClient metrics.MetricsClient,
	tracer tracing.Tracer,
) error {
	ctx, span := tracer.Start(ctx, "postgres.registerGroupRepository")
	defer span.End()

	startTime := time.Now()
	defer func() {
		metricsClient.Timer("repository.init.group.duration").Observe(time.Since(startTime))
	}()

	logger.Info(ctx, "Registrando repositório PostgreSQL para grupos", logging.Fields{
		"component": "infrastructure.postgres",
	})

	// Inicializar o repositório
	repo := NewGroupRepository(db, logger, metricsClient, tracer)

	// Registrar o repositório no container
	if err := c.RegisterGroupRepository(repo); err != nil {
		logger.Error(ctx, "Falha ao registrar repositório de grupos", logging.Fields{
			"error": err.Error(),
		})
		span.SetAttributes(attribute.String("error", err.Error()))
		metricsClient.Counter("repository.init.group.error").Inc(1)
		return fmt.Errorf("falha ao registrar repositório de grupos: %w", err)
	}

	// Verificar conexão e tabelas (opcional)
	if err := testGroupRepositoryConnection(ctx, repo); err != nil {
		logger.Error(ctx, "Falha ao testar conexão do repositório de grupos", logging.Fields{
			"error": err.Error(),
		})
		span.SetAttributes(attribute.String("error", err.Error()))
		metricsClient.Counter("repository.init.group.connection.error").Inc(1)
		return fmt.Errorf("falha ao testar conexão do repositório de grupos: %w", err)
	}

	logger.Info(ctx, "Repositório PostgreSQL para grupos registrado com sucesso", logging.Fields{
		"component": "infrastructure.postgres",
	})
	
	metricsClient.Counter("repository.init.group.success").Inc(1)
	return nil
}

// testGroupRepositoryConnection testa a conexão com o banco de dados para o repositório de grupos
func testGroupRepositoryConnection(ctx context.Context, repo group.Repository) error {
	ctx, span := tracing.GetTracerFromContext(ctx).Start(ctx, "postgres.testGroupRepositoryConnection")
	defer span.End()

	// Realizar uma operação simples para testar a conexão
	// Por exemplo, contar o número total de grupos (limitado a 1 para performance)
	filter := group.GroupFilter{
		// Filtro vazio para contar todos os grupos
	}

	// Este teste verifica apenas se o repositório pode se comunicar com o banco
	// Não importa qual tenant usamos aqui, pois estamos apenas testando a conexão
	// Usaremos o tenant nulo (00000000-0000-0000-0000-000000000000)
	tenantID := group.SystemTenantID // UUID zero para tenant de sistema

	// Tentar listar grupos com limite 1
	_, err := repo.List(ctx, tenantID, filter, 1, 1)
	if err != nil {
		return fmt.Errorf("falha ao testar conexão: %w", err)
	}

	return nil
}