/**
 * INNOVABIZ IAM - Middleware de Logging GraphQL
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação do middleware de logging para GraphQL no módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.12.4 - Logging e monitoramento)
 * - PCI DSS v4.0 (Requisito 10.2 - Registro de atividades)
 * - LGPD/GDPR/PDPA (Arts. 46, 50 - Registros de segurança)
 * - SOX (Sec. 404 - Controles de auditoria)
 * - NIST CSF (DE.CM - Monitoramento contínuo)
 * - OWASP ASVS 4.0 (Requisito 7 - Logging)
 */

package logging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/innovabiz/iam/internal/infrastructure/auth"
)

// Logger define a interface para logging estruturado
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

// GraphQLMiddleware é um middleware para logging de operações GraphQL
type GraphQLMiddleware struct {
	logger Logger
}

// NewGraphQLMiddleware cria uma nova instância do middleware de logging para GraphQL
func NewGraphQLMiddleware(logger Logger) *GraphQLMiddleware {
	return &GraphQLMiddleware{
		logger: logger,
	}
}

// ExtensionName retorna o nome da extensão para gqlgen
func (m *GraphQLMiddleware) ExtensionName() string {
	return "InnovaBizLoggingMiddleware"
}

// Validate valida a extensão
func (m *GraphQLMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

// InterceptOperation intercepta uma operação GraphQL para adicionar logging
func (m *GraphQLMiddleware) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// Obter o contexto da operação
	oc := graphql.GetOperationContext(ctx)
	
	// Gerar ID de correlação para a operação se não existir
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = uuid.New().String()
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}
	
	// Obter usuário do contexto (se autenticado)
	var userID, username, tenantID string
	user, err := auth.GetUserFromContext(ctx)
	if err == nil && user != nil {
		userID = user.ID.String()
		username = user.Username
		if user.TenantID != uuid.Nil {
			tenantID = user.TenantID.String()
		}
	}
	
	// Preparar campos para logging
	fields := map[string]interface{}{
		"correlation_id":  correlationID,
		"operation_id":    oc.OperationID,
		"operation_name":  oc.OperationName,
		"operation_type":  oc.Operation.Operation,
		"user_id":         userID,
		"username":        username,
		"tenant_id":       tenantID,
		"client_name":     oc.Stats.GetExtension("ClientName"),
		"client_version":  oc.Stats.GetExtension("ClientVersion"),
		"request_source":  ctx.Value("request_source"),
	}
	
	// Sanitizar as variáveis para remover informações sensíveis
	sanitizedVars := sanitizeVariables(oc.Variables)
	if len(sanitizedVars) > 0 {
		fields["variables"] = sanitizedVars
	}
	
	// Logar início da operação
	startTime := time.Now()
	m.logger.WithFields(fields).Info("GraphQL operation started")
	
	// Executar a operação
	rh := next(ctx)
	
	return func(ctx context.Context) *graphql.Response {
		// Obter resposta
		resp := rh(ctx)
		
		// Calcular duração
		duration := time.Since(startTime)
		
		// Adicionar informações sobre a resposta
		fields["duration_ms"] = duration.Milliseconds()
		fields["errors"] = len(resp.Errors)
		fields["has_data"] = resp.Data != nil
		fields["extension_keys"] = getKeysFromMap(resp.Extensions)
		
		if len(resp.Errors) > 0 {
			fields["error_messages"] = getErrorMessages(resp.Errors)
			m.logger.WithFields(fields).Warn("GraphQL operation completed with errors")
		} else {
			m.logger.WithFields(fields).Info("GraphQL operation completed successfully")
		}
		
		// Adicionar complexidade da query aos campos de logging se disponível
		if complexity, ok := resp.Extensions["complexity"].(extension.ComplexityStats); ok {
			fields["complexity_score"] = complexity.ActualComplexity
			fields["complexity_limit"] = complexity.MaximumComplexity
		}
		
		return resp
	}
}

// sanitizeVariables remove informações sensíveis das variáveis
func sanitizeVariables(variables map[string]interface{}) map[string]interface{} {
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
			sanitized[k] = sanitizeVariables(m)
		} else {
			sanitized[k] = v
		}
	}
	
	return sanitized
}

// getKeysFromMap extrai as chaves de um mapa
func getKeysFromMap(m map[string]interface{}) []string {
	if len(m) == 0 {
		return nil
	}
	
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	
	return keys
}

// getErrorMessages extrai mensagens de erro
func getErrorMessages(errors []*graphql.Error) []string {
	messages := make([]string, 0, len(errors))
	
	for _, err := range errors {
		messages = append(messages, err.Message)
	}
	
	return messages
}