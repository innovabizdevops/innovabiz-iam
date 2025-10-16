/**
 * @file mutation_resolvers.go
 * @description Resolvers GraphQL para mutations de contextos e atributos
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package resolvers

import (
	"context"
	"time"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"innovabiz/iam/src/multi-context/application/commands"
	"innovabiz/iam/src/multi-context/domain/models"
)

// MutationResolver implementa os resolvers para mutations
type MutationResolver struct {
	createAttributeHandler             *commands.CreateAttributeHandler
	updateAttributeHandler             *commands.UpdateAttributeHandler
	verifyAttributeHandler             *commands.VerifyAttributeHandler
	updateContextVerificationLevelHandler *commands.UpdateContextVerificationLevelHandler
	updateContextTrustScoreHandler     *commands.UpdateContextTrustScoreHandler
	contextService                     *services.ContextService
	auditLogger                        services.AuditLogger
}

// NewMutationResolver cria uma nova instância do resolver para mutations
func NewMutationResolver(
	createAttributeHandler *commands.CreateAttributeHandler,
	updateAttributeHandler *commands.UpdateAttributeHandler,
	verifyAttributeHandler *commands.VerifyAttributeHandler,
	updateContextVerificationLevelHandler *commands.UpdateContextVerificationLevelHandler,
	updateContextTrustScoreHandler *commands.UpdateContextTrustScoreHandler,
	contextService *services.ContextService,
	auditLogger services.AuditLogger,
) *MutationResolver {
	return &MutationResolver{
		createAttributeHandler:             createAttributeHandler,
		updateAttributeHandler:             updateAttributeHandler,
		verifyAttributeHandler:             verifyAttributeHandler,
		updateContextVerificationLevelHandler: updateContextVerificationLevelHandler,
		updateContextTrustScoreHandler:     updateContextTrustScoreHandler,
		contextService:                     contextService,
		auditLogger:                        auditLogger,
	}
}

// GraphQLCreateAttributeInput representa a entrada para criar um atributo
type GraphQLCreateAttributeInput struct {
	ContextID         graphql.ID `json:"contextId"`
	AttributeKey      string     `json:"attributeKey"`
	AttributeValue    string     `json:"attributeValue"`
	SensitivityLevel  *string    `json:"sensitivityLevel"`
	VerificationStatus *string   `json:"verificationStatus"`
	VerificationSource *string   `json:"verificationSource"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// GraphQLUpdateAttributeInput representa a entrada para atualizar um atributo
type GraphQLUpdateAttributeInput struct {
	AttributeID       graphql.ID `json:"attributeId"`
	AttributeValue    *string    `json:"attributeValue"`
	SensitivityLevel  *string    `json:"sensitivityLevel"`
	VerificationStatus *string   `json:"verificationStatus"`
	VerificationSource *string   `json:"verificationSource"`
	Metadata          map[string]interface{} `json:"metadata"`
	Reason            string     `json:"reason"`
}

// GraphQLVerifyAttributeInput representa a entrada para verificar um atributo
type GraphQLVerifyAttributeInput struct {
	AttributeID       graphql.ID `json:"attributeId"`
	VerificationStatus string     `json:"verificationStatus"`
	VerificationSource string     `json:"verificationSource"`
	Notes             *string    `json:"notes"`
	EvidenceMetadata  map[string]interface{} `json:"evidenceMetadata"`
}

// GraphQLUpdateContextVerificationLevelInput representa a entrada para atualizar nível de verificação
type GraphQLUpdateContextVerificationLevelInput struct {
	ContextID        graphql.ID `json:"contextId"`
	VerificationLevel string     `json:"verificationLevel"`
	Reason           string     `json:"reason"`
	VerificationSource string    `json:"verificationSource"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// GraphQLUpdateContextTrustScoreInput representa a entrada para atualizar pontuação de confiança
type GraphQLUpdateContextTrustScoreInput struct {
	ContextID   graphql.ID `json:"contextId"`
	TrustScore  float64    `json:"trustScore"`
	Reason      string     `json:"reason"`
	Source      string     `json:"source"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CreateAttribute resolve a mutation para criar um novo atributo contextual
func (r *MutationResolver) CreateAttribute(ctx context.Context, args struct {
	Input GraphQLCreateAttributeInput
}) (*GraphQLAttribute, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	contextID, err := uuid.Parse(string(args.Input.ContextID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CREATE_ATTRIBUTE_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID de contexto inválido",
				"context_id": string(args.Input.ContextID),
			},
		})
		
		return nil, err
	}
	
	// Preparar o comando
	cmd := commands.CreateAttributeCommand{
		RequestedBy:    userInfo.UserID,
		ContextID:      contextID,
		Key:            args.Input.AttributeKey,
		Value:          args.Input.AttributeValue,
		Metadata:       args.Input.Metadata,
	}
	
	// Aplicar valores opcionais
	if args.Input.SensitivityLevel != nil {
		sensLevel := mapGraphQLSensitivityLevelToModel(*args.Input.SensitivityLevel)
		cmd.SensitivityLevel = sensLevel
	} else {
		cmd.SensitivityLevel = models.SensitivityLevelMedium
	}
	
	if args.Input.VerificationStatus != nil {
		verStatus := mapGraphQLVerificationStatusToModel(*args.Input.VerificationStatus)
		cmd.VerificationStatus = verStatus
	} else {
		cmd.VerificationStatus = models.VerificationStatusUnverified
	}
	
	if args.Input.VerificationSource != nil {
		cmd.VerificationSource = *args.Input.VerificationSource
	}
	
	// Executar o comando
	attribute, err := r.createAttributeHandler.Handle(ctx, cmd)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_CREATE_ATTRIBUTE_FAILED",
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"context_id": contextID.String(),
				"attribute_key": args.Input.AttributeKey,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	result := mapAttributeToGraphQL(attribute)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_CREATE_ATTRIBUTE_SUCCEEDED",
		ResourceID:  attribute.ID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"context_id": contextID.String(),
			"attribute_key": args.Input.AttributeKey,
		},
	})
	
	return result, nil
}

// UpdateAttribute resolve a mutation para atualizar um atributo existente
func (r *MutationResolver) UpdateAttribute(ctx context.Context, args struct {
	Input GraphQLUpdateAttributeInput
}) (*GraphQLAttribute, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	attributeID, err := uuid.Parse(string(args.Input.AttributeID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  string(args.Input.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID de atributo inválido",
			},
		})
		
		return nil, err
	}
	
	// Preparar o comando
	cmd := commands.UpdateAttributeCommand{
		RequestedBy:    userInfo.UserID,
		AttributeID:    attributeID,
		Reason:         args.Input.Reason,
		Metadata:       args.Input.Metadata,
	}
	
	// Aplicar valores opcionais
	if args.Input.AttributeValue != nil {
		cmd.Value = args.Input.AttributeValue
	}
	
	if args.Input.SensitivityLevel != nil {
		sensLevel := mapGraphQLSensitivityLevelToModel(*args.Input.SensitivityLevel)
		cmd.SensitivityLevel = &sensLevel
	}
	
	if args.Input.VerificationStatus != nil {
		verStatus := mapGraphQLVerificationStatusToModel(*args.Input.VerificationStatus)
		cmd.VerificationStatus = &verStatus
	}
	
	if args.Input.VerificationSource != nil {
		cmd.VerificationSource = args.Input.VerificationSource
	}
	
	// Executar o comando
	attribute, err := r.updateAttributeHandler.Handle(ctx, cmd)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_ATTRIBUTE_FAILED",
			ResourceID:  attributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"reason": args.Input.Reason,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	result := mapAttributeToGraphQL(attribute)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_UPDATE_ATTRIBUTE_SUCCEEDED",
		ResourceID:  attributeID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"attribute_key": attribute.Key,
			"reason": args.Input.Reason,
		},
	})
	
	return result, nil
}

// VerifyAttribute resolve a mutation para verificar um atributo
func (r *MutationResolver) VerifyAttribute(ctx context.Context, args struct {
	Input GraphQLVerifyAttributeInput
}) (*GraphQLAttribute, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Verificar permissões de verificador
	if !userHasRole(ctx, "verifier") {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_VERIFY_ATTRIBUTE_ACCESS_DENIED",
			ResourceID:  string(args.Input.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "Permissão de verificador necessária",
			},
		})
		
		return nil, fmt.Errorf("acesso negado: permissão de verificador necessária")
	}
	
	// Converter o ID do formato GraphQL para UUID
	attributeID, err := uuid.Parse(string(args.Input.AttributeID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  string(args.Input.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID de atributo inválido",
			},
		})
		
		return nil, err
	}
	
	// Preparar o comando
	cmd := commands.VerifyAttributeCommand{
		RequestedBy:         userInfo.UserID,
		AttributeID:         attributeID,
		VerificationStatus:  mapGraphQLVerificationStatusToModel(args.Input.VerificationStatus),
		VerificationSource:  args.Input.VerificationSource,
		EvidenceMetadata:    args.Input.EvidenceMetadata,
	}
	
	if args.Input.Notes != nil {
		cmd.Notes = *args.Input.Notes
	}
	
	// Executar o comando
	attribute, err := r.verifyAttributeHandler.Handle(ctx, cmd)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_VERIFY_ATTRIBUTE_FAILED",
			ResourceID:  attributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"status": args.Input.VerificationStatus,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	result := mapAttributeToGraphQL(attribute)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_VERIFY_ATTRIBUTE_SUCCEEDED",
		ResourceID:  attributeID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"attribute_key": attribute.Key,
			"status": args.Input.VerificationStatus,
			"source": args.Input.VerificationSource,
		},
	})
	
	return result, nil
}

// UpdateContextVerificationLevel resolve a mutation para atualizar o nível de verificação de um contexto
func (r *MutationResolver) UpdateContextVerificationLevel(ctx context.Context, args struct {
	Input GraphQLUpdateContextVerificationLevelInput
}) (*GraphQLContext, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Verificar permissões de verificador
	if !userHasRole(ctx, "verifier") {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_VERIFICATION_ACCESS_DENIED",
			ResourceID:  string(args.Input.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "Permissão de verificador necessária",
			},
		})
		
		return nil, fmt.Errorf("acesso negado: permissão de verificador necessária")
	}
	
	// Converter o ID do formato GraphQL para UUID
	contextID, err := uuid.Parse(string(args.Input.ContextID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_VERIFICATION_FAILED",
			ResourceID:  string(args.Input.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID de contexto inválido",
			},
		})
		
		return nil, err
	}
	
	// Preparar o comando
	cmd := commands.UpdateContextVerificationLevelCommand{
		RequestedBy:        userInfo.UserID,
		ContextID:          contextID,
		VerificationLevel:  mapGraphQLVerificationLevelToModel(args.Input.VerificationLevel),
		Reason:             args.Input.Reason,
		VerificationSource: args.Input.VerificationSource,
		Metadata:           args.Input.Metadata,
	}
	
	// Executar o comando
	context, err := r.updateContextVerificationLevelHandler.Handle(ctx, cmd)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_VERIFICATION_FAILED",
			ResourceID:  contextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"level": args.Input.VerificationLevel,
				"reason": args.Input.Reason,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	result := mapIdentityContextToGraphQL(context)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_UPDATE_CONTEXT_VERIFICATION_SUCCEEDED",
		ResourceID:  contextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"level": args.Input.VerificationLevel,
			"reason": args.Input.Reason,
		},
	})
	
	return result, nil
}

// UpdateContextTrustScore resolve a mutation para atualizar a pontuação de confiança de um contexto
func (r *MutationResolver) UpdateContextTrustScore(ctx context.Context, args struct {
	Input GraphQLUpdateContextTrustScoreInput
}) (*GraphQLContext, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Verificar permissões de avaliador de confiança
	if !userHasRole(ctx, "trust_evaluator") {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_TRUST_SCORE_ACCESS_DENIED",
			ResourceID:  string(args.Input.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "Permissão de avaliador de confiança necessária",
			},
		})
		
		return nil, fmt.Errorf("acesso negado: permissão de avaliador de confiança necessária")
	}
	
	// Converter o ID do formato GraphQL para UUID
	contextID, err := uuid.Parse(string(args.Input.ContextID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_TRUST_SCORE_FAILED",
			ResourceID:  string(args.Input.ContextID),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID de contexto inválido",
			},
		})
		
		return nil, err
	}
	
	// Validar a pontuação de confiança (entre 0 e 1)
	if args.Input.TrustScore < 0 || args.Input.TrustScore > 1 {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_TRUST_SCORE_FAILED",
			ResourceID:  contextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "Pontuação de confiança deve estar entre 0 e 1",
				"trust_score": args.Input.TrustScore,
			},
		})
		
		return nil, fmt.Errorf("pontuação de confiança deve estar entre 0 e 1")
	}
	
	// Preparar o comando
	cmd := commands.UpdateContextTrustScoreCommand{
		RequestedBy:  userInfo.UserID,
		ContextID:    contextID,
		TrustScore:   args.Input.TrustScore,
		Reason:       args.Input.Reason,
		Source:       args.Input.Source,
		Metadata:     args.Input.Metadata,
	}
	
	// Executar o comando
	context, err := r.updateContextTrustScoreHandler.Handle(ctx, cmd)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_UPDATE_CONTEXT_TRUST_SCORE_FAILED",
			ResourceID:  contextID.String(),
			ResourceType: "IDENTITY_CONTEXT",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"trust_score": args.Input.TrustScore,
				"reason": args.Input.Reason,
			},
		})
		
		return nil, err
	}
	
	// Converter para o tipo GraphQL
	result := mapIdentityContextToGraphQL(context)
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_UPDATE_CONTEXT_TRUST_SCORE_SUCCEEDED",
		ResourceID:  contextID.String(),
		ResourceType: "IDENTITY_CONTEXT",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"trust_score": args.Input.TrustScore,
			"reason": args.Input.Reason,
			"source": args.Input.Source,
		},
	})
	
	return result, nil
}

// DeleteAttribute resolve a mutation para excluir (marcar como excluído) um atributo
func (r *MutationResolver) DeleteAttribute(ctx context.Context, args struct {
	AttributeID graphql.ID
	Reason      string
}) (bool, error) {
	// Extrair informações de autenticação do contexto da requisição
	userInfo := extractUserInfo(ctx)
	
	// Converter o ID do formato GraphQL para UUID
	attributeID, err := uuid.Parse(string(args.AttributeID))
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_DELETE_ATTRIBUTE_FAILED",
			ResourceID:  string(args.AttributeID),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": "ID de atributo inválido",
			},
		})
		
		return false, err
	}
	
	// Em um sistema real, aqui seria chamado um handler específico para exclusão lógica
	// Para este exemplo, simplificaremos usando o serviço diretamente
	err = r.contextService.DeleteAttribute(ctx, attributeID, userInfo.UserID, args.Reason)
	if err != nil {
		r.auditLogger.LogEvent(ctx, services.AuditEvent{
			EventType:   "GRAPHQL_DELETE_ATTRIBUTE_FAILED",
			ResourceID:  attributeID.String(),
			ResourceType: "CONTEXT_ATTRIBUTE",
			UserID:      userInfo.UserID,
			Timestamp:   time.Now(),
			Details: map[string]interface{}{
				"error": err.Error(),
				"reason": args.Reason,
			},
		})
		
		return false, err
	}
	
	r.auditLogger.LogEvent(ctx, services.AuditEvent{
		EventType:   "GRAPHQL_DELETE_ATTRIBUTE_SUCCEEDED",
		ResourceID:  attributeID.String(),
		ResourceType: "CONTEXT_ATTRIBUTE",
		UserID:      userInfo.UserID,
		Timestamp:   time.Now(),
		Details: map[string]interface{}{
			"reason": args.Reason,
		},
	})
	
	return true, nil
}

// userHasRole verifica se o usuário possui o papel especificado
func userHasRole(ctx context.Context, role string) bool {
	// Em um sistema real, isso seria implementado para verificar os papéis
	// do usuário a partir do token JWT ou outra fonte de autorização
	
	// Implementação temporária para demonstração
	userRoles := ctx.Value("user_roles")
	if userRoles == nil {
		return false
	}
	
	roles, ok := userRoles.([]string)
	if !ok {
		return false
	}
	
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	
	return false
}