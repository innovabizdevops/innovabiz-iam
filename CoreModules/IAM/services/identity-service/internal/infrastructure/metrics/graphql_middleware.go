/**
 * INNOVABIZ IAM - Middleware de Métricas GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do middleware de métricas para GraphQL no módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.12.4 - Monitoramento)
 * - PCI DSS v4.0 (Requisito 10.2 - Monitoramento de eventos)
 * - SOX (Sec. 404 - Controles de performance)
 * - NIST CSF (DE.AE - Análise de eventos)
 * - ISO/IEC 20000-1:2018 (Gestão de performance)
 */

package metrics

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/google/uuid"

	"github.com/innovabiz/iam/internal/infrastructure/auth"
)

// Define constantes para métricas do GraphQL
const (
	MetricGraphQLOperations      = "graphql_operations_total"
	MetricGraphQLErrors          = "graphql_errors_total"
	MetricGraphQLDuration        = "graphql_operation_duration_ms"
	MetricGraphQLComplexity      = "graphql_operation_complexity"
	MetricGraphQLDepth           = "graphql_operation_depth"
	MetricAuthSuccess            = "auth_success_total"
	MetricAuthErrors             = "auth_errors_total" 
	MetricValidationSuccess      = "validation_success_total"
	MetricValidationErrors       = "validation_errors_total"
)

// MetricsClient define a interface para métricas
type MetricsClient interface {
	// Contadores
	IncCounter(name string, labels map[string]string, value float64)
	
	// Histogramas
	ObserveHistogram(name string, labels map[string]string, value float64)
	
	// Gauges
	SetGauge(name string, labels map[string]string, value float64)
	
	// Utilidades
	TimeFn(name string, labels map[string]string, fn func() error) error
}

// GraphQLMiddleware é um middleware para coletar métricas de operações GraphQL
type GraphQLMiddleware struct {
	metrics MetricsClient
}

// NewGraphQLMiddleware cria uma nova instância do middleware de métricas para GraphQL
func NewGraphQLMiddleware(metrics MetricsClient) *GraphQLMiddleware {
	return &GraphQLMiddleware{
		metrics: metrics,
	}
}

// ExtensionName retorna o nome da extensão para gqlgen
func (m *GraphQLMiddleware) ExtensionName() string {
	return "InnovaBizMetricsMiddleware"
}

// Validate valida a extensão
func (m *GraphQLMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptOperation intercepta uma operação GraphQL para adicionar métricas
func (m *GraphQLMiddleware) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// Obter o contexto da operação
	oc := graphql.GetOperationContext(ctx)
	
	// Obter usuário do contexto (se autenticado)
	var userID, tenantID string
	user, err := auth.GetUserFromContext(ctx)
	if err == nil && user != nil {
		userID = user.ID.String()
		if user.TenantID != uuid.Nil {
			tenantID = user.TenantID.String()
		}
	}
	
	// Preparar labels para métricas
	labels := map[string]string{
		"operation_type": string(oc.Operation.Operation),
		"operation_name": oc.OperationName,
		"tenant_id":      tenantID,
		"has_user":       boolToString(userID != ""),
	}
	
	// Registrar início da operação
	startTime := time.Now()
	
	// Executar a operação
	rh := next(ctx)
	
	return func(ctx context.Context) *graphql.Response {
		// Obter resposta
		resp := rh(ctx)
		
		// Calcular duração
		duration := time.Since(startTime)
		
		// Incrementar contador de operações
		m.metrics.IncCounter(MetricGraphQLOperations, labels, 1)
		
		// Registrar duração da operação
		m.metrics.ObserveHistogram(MetricGraphQLDuration, labels, float64(duration.Milliseconds()))
		
		// Registrar métricas de erros
		if len(resp.Errors) > 0 {
			errorLabels := copyLabels(labels)
			errorLabels["error_code"] = classifyErrors(resp.Errors)
			m.metrics.IncCounter(MetricGraphQLErrors, errorLabels, float64(len(resp.Errors)))
		}
		
		// Adicionar métricas de complexidade se disponíveis
		if complexity, ok := resp.Extensions["complexity"].(extension.ComplexityStats); ok {
			complexityLabels := copyLabels(labels)
			m.metrics.SetGauge(MetricGraphQLComplexity, complexityLabels, float64(complexity.ActualComplexity))
		}
		
		// Adicionar métricas de profundidade da query (estimativa baseada na análise da query)
		queryDepth := estimateQueryDepth(oc.Operation)
		m.metrics.SetGauge(MetricGraphQLDepth, labels, float64(queryDepth))
		
		return resp
	}
}

// Funções auxiliares para incrementar métricas específicas
// Estas funções são utilizadas pelos resolvers e diretivas

// IncAuthSuccess incrementa métricas de autenticação bem-sucedidas
func IncAuthSuccess(operationName string) {
	// Implementação real integraria com o cliente de métricas
	// Esta é uma versão simplificada para ilustração
}

// IncAuthErrors incrementa métricas de erros de autenticação
func IncAuthErrors(errorType, operationName string) {
	// Implementação real integraria com o cliente de métricas
	// Esta é uma versão simplificada para ilustração
}

// IncValidationSuccess incrementa métricas de validação bem-sucedidas
func IncValidationSuccess(inputName string) {
	// Implementação real integraria com o cliente de métricas
	// Esta é uma versão simplificada para ilustração
}

// IncValidationErrors incrementa métricas de erros de validação
func IncValidationErrors(errorType, inputName string) {
	// Implementação real integraria com o cliente de métricas
	// Esta é uma versão simplificada para ilustração
}

// Funções auxiliares internas

// boolToString converte um valor booleano para string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// copyLabels cria uma cópia de um mapa de labels
func copyLabels(labels map[string]string) map[string]string {
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return copy
}

// classifyErrors classifica erros em categorias para métricas
func classifyErrors(errors []*graphql.Error) string {
	// Simplificado - em produção teria análise mais sofisticada
	if len(errors) == 0 {
		return "none"
	}
	
	// Verificar o primeiro erro para classificação
	err := errors[0]
	
	if err.Path != nil && len(err.Path) > 0 {
		return "resolver_error"
	}
	
	if err.Extensions != nil {
		if _, ok := err.Extensions["code"]; ok {
			code, _ := err.Extensions["code"].(string)
			return code
		}
	}
	
	return "unknown"
}

// estimateQueryDepth estima a profundidade de uma query GraphQL
func estimateQueryDepth(op *graphql.OperationDefinition) int {
	if op == nil || len(op.SelectionSet) == 0 {
		return 0
	}
	
	// Uma implementação real faria uma análise detalhada da estrutura da query
	// Esta é uma versão simplificada para ilustração
	maxDepth := 1
	
	// Análise recursiva para identificar a profundidade máxima
	// da hierarquia de seleções
	
	return maxDepth
}