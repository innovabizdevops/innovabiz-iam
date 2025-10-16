package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/innovabiz/iam/internal/domain/model/errors"
	"github.com/innovabiz/iam/internal/infrastructure/auth"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AuthDirective implementa a diretiva @auth do GraphQL para controle de acesso
// baseado em autenticação, permissões, roles e contexto de tenant.
type AuthDirective struct {
	tracer trace.Tracer
	logger observability.Logger
}

// NewAuthDirective cria uma nova instância da diretiva de autenticação
func NewAuthDirective(tracer trace.Tracer, logger observability.Logger) *AuthDirective {
	return &AuthDirective{
		tracer: tracer,
		logger: logger,
	}
}

// Directive é o ponto de entrada para a diretiva @auth e será chamada pelo
// runtime do GraphQL para verificar autorização antes de executar o campo
func (d *AuthDirective) Directive() func(ctx context.Context, obj interface{}, next graphql.Resolver, requires []string, allowSameUser bool, checkTenant bool) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, requires []string, allowSameUser bool, checkTenant bool) (interface{}, error) {
		// Iniciar span para observabilidade da verificação de autorização
		authCtx, span := d.tracer.Start(ctx, "auth.directive.authorize",
			trace.WithAttributes(
				attribute.StringSlice("auth.requires", requires),
				attribute.Bool("auth.allow_same_user", allowSameUser),
				attribute.Bool("auth.check_tenant", checkTenant),
			),
		)
		defer span.End()

		// Extrair informações do usuário autenticado do contexto
		authInfo := auth.GetAuthInfoFromContext(ctx)
		if authInfo == nil {
			d.logger.Warn(ctx, "Auth directive: no auth info in context")
			span.SetAttributes(attribute.Bool("auth.success", false))
			span.SetAttributes(attribute.String("auth.error", "not_authenticated"))
			return nil, errors.ErrUnauthorized
		}

		// Se não há usuário autenticado, negar acesso
		if authInfo.UserID == "" {
			d.logger.Warn(ctx, "Auth directive: no authenticated user")
			span.SetAttributes(attribute.Bool("auth.success", false))
			span.SetAttributes(attribute.String("auth.error", "no_user"))
			return nil, errors.ErrUnauthorized
		}

		// Adicionar atributos ao span para auditoria e observabilidade
		span.SetAttributes(attribute.String("user.id", authInfo.UserID))
		span.SetAttributes(attribute.String("tenant.id", authInfo.TenantID))

		// Se há permissões específicas requeridas, verificá-las
		if len(requires) > 0 {
			hasAllPermissions := true
			var missingPermissions []string

			for _, permission := range requires {
				if !authInfo.HasPermission(permission) {
					hasAllPermissions = false
					missingPermissions = append(missingPermissions, permission)
				}
			}

			// Se não tem todas as permissões necessárias, verificar exceções especiais
			if !hasAllPermissions {
				if allowSameUser {
					// Verificar se é o próprio usuário (self-service)
					// Essa lógica será implementada analisando o objeto e parâmetros da query
					isSameUser, err := d.checkIfSameUser(ctx, obj, authInfo)
					if err != nil {
						return nil, err
					}

					if isSameUser {
						// Permissão especial para operações no próprio usuário
						span.SetAttributes(attribute.Bool("auth.same_user", true))
						span.SetAttributes(attribute.Bool("auth.success", true))
						return next(ctx)
					}
				}

				// Logging e observabilidade para acesso negado
				d.logger.Warn(ctx, "Auth directive: permission denied",
					"userId", authInfo.UserID,
					"tenantId", authInfo.TenantID,
					"requiredPermissions", requires,
					"missingPermissions", missingPermissions,
				)

				span.SetAttributes(attribute.Bool("auth.success", false))
				span.SetAttributes(attribute.String("auth.error", "permission_denied"))
				span.SetAttributes(attribute.StringSlice("auth.missing_permissions", missingPermissions))

				return nil, errors.NewForbiddenError("insufficient_permissions", "Permissões insuficientes para esta operação")
			}
		}

		// Verificar contexto de tenant quando necessário
		if checkTenant {
			// Extração do tenant ID do objeto ou parâmetros é implementada aqui
			targetTenantID, err := d.extractTargetTenantID(ctx, obj)
			if err == nil && targetTenantID != "" && targetTenantID != authInfo.TenantID {
				// Verificar se é um caso especial de cross-tenant permitido
				if !d.isCrossTenantAllowed(ctx, authInfo, targetTenantID) {
					d.logger.Warn(ctx, "Auth directive: cross-tenant access denied",
						"userId", authInfo.UserID,
						"userTenantId", authInfo.TenantID,
						"targetTenantId", targetTenantID,
					)

					span.SetAttributes(attribute.Bool("auth.success", false))
					span.SetAttributes(attribute.String("auth.error", "tenant_mismatch"))
					span.SetAttributes(attribute.String("auth.target_tenant", targetTenantID))

					return nil, errors.NewForbiddenError("tenant_access_denied", "Acesso a outro tenant não permitido")
				}
			}
		}

		// Autorização bem-sucedida
		span.SetAttributes(attribute.Bool("auth.success", true))
		
		// Enriquecer o contexto para o resolver com informações adicionais de segurança
		enrichedCtx := auth.EnrichContext(authCtx, authInfo)
		
		// Chamar o próximo resolver com o contexto enriquecido
		return next(enrichedCtx)
	}
}

// checkIfSameUser verifica se a operação está sendo realizada pelo próprio usuário
func (d *AuthDirective) checkIfSameUser(ctx context.Context, obj interface{}, authInfo *auth.AuthInfo) (bool, error) {
	// Extrair ID do usuário alvo da operação
	// Isto depende da estrutura do campo e parâmetros
	targetUserID, err := d.extractTargetUserID(ctx, obj)
	if err != nil {
		return false, err
	}

	return targetUserID == authInfo.UserID, nil
}

// extractTargetUserID extrai o ID do usuário alvo da operação
// a partir do objeto GraphQL e do contexto da resolução
func (d *AuthDirective) extractTargetUserID(ctx context.Context, obj interface{}) (string, error) {
	// Implementação para extrair o ID do usuário do objeto ou argumentos
	// da query/mutation atual
	
	// Obter o campo atual que está sendo resolvido
	fieldContext := graphql.GetFieldContext(ctx)
	if fieldContext == nil {
		return "", nil
	}

	// Tentar extrair de argumentos comuns
	if id, exists := fieldContext.Args["id"].(string); exists {
		return id, nil
	}
	
	if id, exists := fieldContext.Args["userId"].(string); exists {
		return id, nil
	}

	// Para objetos do tipo User, extrair do próprio objeto
	if user, ok := obj.(map[string]interface{}); ok {
		if id, exists := user["id"].(string); exists {
			return id, nil
		}
	}

	// Não foi possível determinar o ID do usuário alvo
	return "", nil
}

// extractTargetTenantID extrai o ID do tenant alvo da operação
func (d *AuthDirective) extractTargetTenantID(ctx context.Context, obj interface{}) (string, error) {
	// Obter o campo atual que está sendo resolvido
	fieldContext := graphql.GetFieldContext(ctx)
	if fieldContext == nil {
		return "", nil
	}

	// Tentar extrair de argumentos comuns
	if id, exists := fieldContext.Args["tenantId"].(string); exists {
		return id, nil
	}
	
	// Para filtros que contêm tenantId
	if filter, exists := fieldContext.Args["filter"].(map[string]interface{}); exists {
		if id, exists := filter["tenantId"].(string); exists {
			return id, nil
		}
	}

	// Para objetos que têm tenantId
	if obj != nil {
		if objMap, ok := obj.(map[string]interface{}); ok {
			if id, exists := objMap["tenantId"].(string); exists {
				return id, nil
			}
		}
	}

	return "", nil
}

// isCrossTenantAllowed verifica se o acesso cross-tenant é permitido para o usuário
func (d *AuthDirective) isCrossTenantAllowed(ctx context.Context, authInfo *auth.AuthInfo, targetTenantID string) bool {
	// Verificar se o usuário tem permissão especial para acesso cross-tenant
	if authInfo.HasPermission("IAM:CrossTenantAccess") {
		d.logger.Info(ctx, "Cross-tenant access allowed due to special permission",
			"userId", authInfo.UserID,
			"userTenantId", authInfo.TenantID,
			"targetTenantId", targetTenantID,
		)
		return true
	}
	
	// Verificar se o tenant alvo é um subtenant do tenant do usuário
	isSubTenant := d.checkSubTenantRelation(ctx, authInfo.TenantID, targetTenantID)
	if isSubTenant {
		d.logger.Info(ctx, "Cross-tenant access allowed due to subtenant relationship",
			"userId", authInfo.UserID,
			"userTenantId", authInfo.TenantID,
			"targetTenantId", targetTenantID,
		)
		return true
	}

	// Por padrão, negar acesso cross-tenant
	return false
}

// checkSubTenantRelation verifica se existe relação hierárquica entre tenants
func (d *AuthDirective) checkSubTenantRelation(ctx context.Context, parentTenantID, childTenantID string) bool {
	// Implementação da lógica de verificação de hierarquia de tenants
	// Esta é uma implementação simplificada; na prática, isso exigiria
	// consulta à base de dados ou cache para verificar a relação
	
	// Exemplo de lógica:
	// 1. Consultar cache ou banco de dados para verificar relação
	// 2. Verificar se o tenant alvo está na árvore do tenant atual
	
	// Stub para implementação futura
	return false
}