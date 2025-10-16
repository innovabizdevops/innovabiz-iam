/**
 * INNOVABIZ IAM - Middleware de Tracing GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do middleware de tracing distribuído para GraphQL no módulo Core IAM,
 * baseado em OpenTelemetry, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.12.4 - Monitoramento)
 * - PCI DSS v4.0 (Requisito 10.2.1 - Rastreabilidade de acessos)
 * - LGPD/GDPR/PDPA (Arts. 37, 38 - Registros de acesso)
 * - SOX (Sec. 404 - Rastreabilidade de transações)
 * - NIST CSF (ID.AM - Rastreabilidade de eventos)
 * - OWASP ASVS 4.0 (Requisito 7.2 - Rastreabilidade de logs)
 */

package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/innovabiz/iam/internal/infrastructure/auth"
)

// Tracer define a interface para tracing distribuído
type Tracer interface {
	// Iniciar um novo span a partir do contexto
	StartFromContext(ctx context.Context, spanName string) (trace.Span, context.Context)
	
	// Injetar contexto de tracing em headers HTTP
	Inject(ctx context.Context, headers map[string]string)
	
	// Extrair contexto de tracing de headers HTTP
	Extract(ctx context.Context, headers map[string]string) context.Context
	
	// Iniciar um span para operações de banco de dados
	StartDBSpan(ctx context.Context, operation, statement string) (trace.Span, context.Context)
	
	// Iniciar um span para chamadas externas
	StartExternalCallSpan(ctx context.Context, service, operation string) (trace.Span, context.Context)
}

// GraphQLMiddleware é um middleware para tracing de operações GraphQL
type GraphQLMiddleware struct {
	tracer Tracer
}

// NewGraphQLMiddleware cria uma nova instância do middleware de tracing para GraphQL
func NewGraphQLMiddleware(tracer Tracer) *GraphQLMiddleware {
	return &GraphQLMiddleware{
		tracer: tracer,
	}
}

// ExtensionName retorna o nome da extensão para gqlgen
func (m *GraphQLMiddleware) ExtensionName() string {
	return "InnovaBizTracingMiddleware"
}

// Validate valida a extensão
func (m *GraphQLMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptOperation intercepta uma operação GraphQL para adicionar tracing
func (m *GraphQLMiddleware) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// Obter o contexto da operação
	oc := graphql.GetOperationContext(ctx)
	
	// Gerar ID de correlação para a operação se não existir
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = uuid.New().String()
		ctx = context.WithValue(ctx, "correlation_id", correlationID.(string))
	}
	
	// Criar nome para o span
	operationName := oc.OperationName
	if operationName == "" {
		operationName = "unnamed_operation"
	}
	spanName := fmt.Sprintf("GraphQL.%s.%s", strings.ToLower(string(oc.Operation.Operation)), operationName)
	
	// Iniciar span para a operação
	span, ctx := m.tracer.StartFromContext(ctx, spanName)
	defer span.End()
	
	// Adicionar atributos ao span
	span.SetAttributes(
		attribute.String("graphql.operation_id", oc.OperationID),
		attribute.String("graphql.operation_name", operationName),
		attribute.String("graphql.operation_type", string(oc.Operation.Operation)),
		attribute.String("correlation_id", correlationID.(string)),
	)
	
	// Adicionar informações sobre variáveis (sanitizadas)
	if len(oc.Variables) > 0 {
		sanitizedVars := sanitizeVariablesForTracing(oc.Variables)
		varJSON, err := json.Marshal(sanitizedVars)
		if err == nil {
			span.SetAttributes(attribute.String("graphql.variables", string(varJSON)))
		}
	}
	
	// Adicionar informações sobre o usuário se disponível
	user, err := auth.GetUserFromContext(ctx)
	if err == nil && user != nil {
		span.SetAttributes(
			attribute.String("user.id", user.ID.String()),
			attribute.String("user.username", user.Username),
			attribute.String("tenant.id", user.TenantID.String()),
		)
	}
	
	// Adicionar informações sobre a query
	queryString := oc.RawQuery
	if len(queryString) > 1000 {
		// Truncar queries muito grandes para evitar sobrecarga no span
		queryString = queryString[:1000] + "..."
	}
	span.SetAttributes(attribute.String("graphql.query", queryString))
	
	// Executar a operação
	rh := next(ctx)
	
	return func(ctx context.Context) *graphql.Response {
		// Obter resposta
		resp := rh(ctx)
		
		// Registrar erros se existirem
		if len(resp.Errors) > 0 {
			// Marcar o span como erro
			span.SetStatus(codes.Error, resp.Errors[0].Message)
			
			// Adicionar detalhes dos erros ao span
			errorMessages := make([]string, 0, len(resp.Errors))
			for _, err := range resp.Errors {
				errorMessages = append(errorMessages, err.Message)
				span.RecordError(fmt.Errorf("%s", err.Message))
			}
			
			span.SetAttributes(attribute.StringSlice("graphql.errors", errorMessages))
		} else {
			span.SetStatus(codes.Ok, "")
		}
		
		// Adicionar informações sobre o tamanho da resposta (estimativa)
		if resp.Data != nil {
			dataJSON, err := json.Marshal(resp.Data)
			if err == nil {
				span.SetAttributes(attribute.Int("graphql.response_size_bytes", len(dataJSON)))
			}
		}
		
		return resp
	}
}

// StartFromContext inicia um novo span a partir do contexto atual
// Este método é utilizado pelos resolvers para adicionar tracing específico
func (m *GraphQLMiddleware) StartFromContext(ctx context.Context, spanName string) (trace.Span, context.Context) {
	return m.tracer.StartFromContext(ctx, spanName)
}

// sanitizeVariablesForTracing remove informações sensíveis das variáveis para tracing
func sanitizeVariablesForTracing(variables map[string]interface{}) map[string]interface{} {
	if len(variables) == 0 {
		return nil
	}
	
	sanitized := make(map[string]interface{})
	
	// Lista de campos sensíveis para mascarar
	sensitiveFields := []string{
		"password", "senha", "secret", "token", "apiKey", "api_key", "apiSecret",
		"api_secret", "auth", "credentials", "credential", "pin", "cvv", "cvc",
	}
	
	for k, v := range variables {
		// Verificar se é um campo sensível
		isSensitive := false
		for _, field := range sensitiveFields {
			if k == field {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			sanitized[k] = "******"
		} else if m, ok := v.(map[string]interface{}); ok {
			// Recursivamente sanitizar maps
			sanitized[k] = sanitizeVariablesForTracing(m)
		} else {
			sanitized[k] = v
		}
	}
	
	return sanitized
}