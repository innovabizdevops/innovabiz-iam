package directives

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/innovabiz/iam/internal/infrastructure/auth"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RedactDirective implementa a diretiva @redact do GraphQL
// para mascarar dados sensíveis nas respostas.
type RedactDirective struct {
	tracer trace.Tracer
	logger observability.Logger
}

// NewRedactDirective cria uma nova instância da diretiva de redação
func NewRedactDirective(tracer trace.Tracer, logger observability.Logger) *RedactDirective {
	return &RedactDirective{
		tracer: tracer,
		logger: logger,
	}
}

// Directive é o ponto de entrada para a diretiva @redact e será chamada pelo
// runtime do GraphQL após executar o resolver do campo correspondente
func (d *RedactDirective) Directive() func(ctx context.Context, obj interface{}, next graphql.Resolver, replacement string) (interface{}, error) {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, replacement string) (interface{}, error) {
		// Iniciar span para observabilidade da redação
		redactCtx, span := d.tracer.Start(ctx, "redact.directive.apply")
		defer span.End()

		// Obter o campo atual que está sendo resolvido
		fieldContext := graphql.GetFieldContext(ctx)
		if fieldContext == nil {
			// Se não temos contexto do campo, simplesmente passamos para o próximo resolver
			return next(ctx)
		}

		// Extrair informações do contexto de autenticação
		authInfo := auth.GetAuthInfoFromContext(ctx)
		
		// Obter o valor original do campo
		val, err := next(ctx)
		if err != nil || val == nil {
			// Em caso de erro ou valor nulo, repassar para o cliente sem alteração
			return val, err
		}

		// Verificar se o usuário tem permissão para ver este campo não redatado
		shouldRedact := d.shouldRedactField(ctx, authInfo, fieldContext)
		
		if shouldRedact {
			// Aplicar redação baseada no tipo do campo
			redactedValue := d.applyRedaction(val, replacement)
			
			// Logging para auditoria de acesso a dados sensíveis
			d.logger.Debug(ctx, "Field redacted due to sensitivity",
				"field_path", fieldContext.Path().String(),
				"field_name", fieldContext.Field.Name,
				"user_id", getStringSafe(authInfo, "UserID"),
				"tenant_id", getStringSafe(authInfo, "TenantID"),
			)
			
			span.SetAttributes(attribute.Bool("redact.applied", true))
			span.SetAttributes(attribute.String("redact.field", fieldContext.Field.Name))
			
			return redactedValue, nil
		}
		
		// Se não precisa redatar, retornar o valor original
		span.SetAttributes(attribute.Bool("redact.applied", false))
		return val, nil
	}
}

// shouldRedactField decide se um campo deve ser redatado com base no contexto
func (d *RedactDirective) shouldRedactField(ctx context.Context, authInfo *auth.AuthInfo, fieldContext *graphql.FieldContext) bool {
	// Se não há contexto de autenticação, sempre redatar campos sensíveis
	if authInfo == nil {
		return true
	}

	// Verificar permissões especiais que permitem ver dados sensíveis
	if authInfo.HasPermission("IAM:ViewSensitiveData") {
		return false
	}

	// Verificar se é o próprio usuário acessando seus dados (princípio de self-service)
	if d.isSelfServiceAccess(ctx, authInfo, fieldContext) {
		return false
	}

	// Verificar regras de controle de acesso baseado no contexto
	// Por exemplo, conformidade com GDPR para residentes da UE
	if d.hasContextualAccessRestriction(ctx, authInfo, fieldContext) {
		return true
	}

	// Por padrão, redatar o campo se não houver regras específicas permitindo acesso
	return true
}

// isSelfServiceAccess verifica se é o próprio usuário acessando seus dados
func (d *RedactDirective) isSelfServiceAccess(ctx context.Context, authInfo *auth.AuthInfo, fieldContext *graphql.FieldContext) bool {
	// Analisar o caminho do campo para determinar se pertence ao usuário atual
	// Exemplo: verificar se está acessando "user(id: X)" onde X é o ID do próprio usuário
	
	// Este é um exemplo simplificado; a implementação real dependeria da estrutura do seu schema
	path := fieldContext.Path()
	
	// Verificar se estamos dentro de um objeto User
	if len(path) >= 2 && path[len(path)-2] == "User" {
		// Tentar encontrar o ID do usuário no objeto
		parentObj := fieldContext.Object
		if parentObj == "User" {
			// Verificar se o objeto pai tem um campo 'id'
			if parent, ok := fieldContext.Parent.Result.(map[string]interface{}); ok {
				if userID, exists := parent["id"].(string); exists {
					return userID == authInfo.UserID
				}
			}
		}
	}
	
	return false
}

// hasContextualAccessRestriction verifica restrições contextuais de acesso
func (d *RedactDirective) hasContextualAccessRestriction(ctx context.Context, authInfo *auth.AuthInfo, fieldContext *graphql.FieldContext) bool {
	// Implementação de regras de acesso baseadas em contexto
	// Exemplos:
	// - Restrições geográficas (ex: dados PII para usuários da UE sob GDPR)
	// - Restrições regulatórias (ex: dados financeiros sob PCI DSS)
	// - Restrições de classificação de dados
	
	// Obter contexto regulatório do usuário/tenant atual
	// regulatoryContext := d.getRegulatoryContext(ctx, authInfo)
	
	// Exemplo de verificação para GDPR
	// if regulatoryContext.AppliesGDPR && fieldContext.Field.Name == "email" {
	//     return true
	// }
	
	// Esta é uma implementação simplificada; em produção, usaria um serviço
	// dedicado para avaliar regras de conformidade baseado no contexto completo
	
	return false
}

// applyRedaction aplica a redação conforme o tipo de dado
func (d *RedactDirective) applyRedaction(val interface{}, replacement string) interface{} {
	if val == nil {
		return nil
	}
	
	// Tratar diferentes tipos de dados
	switch reflect.TypeOf(val).Kind() {
	case reflect.String:
		// Para strings, substituir pelo texto de redação
		return replacement
		
	case reflect.Slice, reflect.Array:
		// Para arrays, redatar cada elemento
		sliceValue := reflect.ValueOf(val)
		length := sliceValue.Len()
		
		// Se for um array vazio, retornar como está
		if length == 0 {
			return val
		}
		
		// Para slices de strings, redatar cada elemento
		if reflect.TypeOf(val).Elem().Kind() == reflect.String {
			result := make([]string, length)
			for i := 0; i < length; i++ {
				result[i] = replacement
			}
			return result
		}
		
		// Para outros tipos de slice, indicar que contém dados sensíveis
		return fmt.Sprintf("[%d %s]", length, replacement)
		
	case reflect.Map:
		// Para mapas, indicar que contém dados sensíveis
		return map[string]string{"_redacted": replacement}
		
	case reflect.Struct:
		// Para estruturas, indicar que contém dados sensíveis
		return replacement
		
	default:
		// Para números e booleanos, podemos usar um valor padrão ou indicar redação
		if reflect.TypeOf(val).Kind() == reflect.Int || 
		   reflect.TypeOf(val).Kind() == reflect.Int64 ||
		   reflect.TypeOf(val).Kind() == reflect.Float64 {
			return 0
		}
		
		if reflect.TypeOf(val).Kind() == reflect.Bool {
			return false
		}
		
		// Tipo desconhecido, usar texto de redação genérico
		return replacement
	}
}

// Função auxiliar para extrair strings de forma segura
func getStringSafe(obj interface{}, field string) string {
	// Se o objeto for nulo, retornar string vazia
	if obj == nil {
		return ""
	}
	
	// Usar reflection para extrair o campo de forma segura
	val := reflect.ValueOf(obj)
	
	// Se for um ponteiro, obter o valor apontado
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}
	
	// Verificar se é uma estrutura
	if val.Kind() != reflect.Struct {
		return ""
	}
	
	// Tentar obter o campo
	fieldVal := val.FieldByName(field)
	if !fieldVal.IsValid() {
		return ""
	}
	
	// Se o campo for uma string, retornar seu valor
	if fieldVal.Kind() == reflect.String {
		return fieldVal.String()
	}
	
	// Caso contrário, retornar string vazia
	return ""
}