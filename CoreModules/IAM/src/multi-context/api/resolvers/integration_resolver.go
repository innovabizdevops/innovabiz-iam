package resolvers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	
	"innovabiz/iam/multi-context/domain/models"
	"innovabiz/iam/multi-context/domain/services"
	"innovabiz/iam/multi-context/infrastructure/logging"
)

// IntegrationResolver resolve consultas e mutações relacionadas às integrações entre contextos
type IntegrationResolver struct {
	integrationService *services.IntegrationService
	contextService     *services.ContextService
	logger             *logging.Logger
	integration        *models.ContextIntegration
}

// NewIntegrationResolver cria uma nova instância do resolver de integrações
func NewIntegrationResolver(
	integrationService *services.IntegrationService,
	contextService *services.ContextService,
	logger *logging.Logger,
) *IntegrationResolver {
	return &IntegrationResolver{
		integrationService: integrationService,
		contextService:     contextService,
		logger:             logger,
	}
}

// ContextIntegration resolve a consulta para obter uma integração pelo ID
func (r *IntegrationResolver) ContextIntegration(ctx context.Context, args struct {
	IntegrationID graphql.ID
}) (*IntegrationResolver, error) {
	integrationID, err := uuid.Parse(string(args.IntegrationID))
	if err != nil {
		r.logger.Error("invalid integration ID format", "error", err, "id", args.IntegrationID)
		return nil, errors.New("ID de integração inválido")
	}

	integration, err := r.integrationService.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		r.logger.Error("failed to get integration", "error", err, "id", integrationID)
		return nil, err
	}

	if integration == nil {
		return nil, nil // Retorna null se não encontrar a integração
	}

	return &IntegrationResolver{
		integrationService: r.integrationService,
		contextService:     r.contextService,
		logger:             r.logger,
		integration:        integration,
	}, nil
}

// ContextIntegrations resolve a consulta para listar integrações com filtros
func (r *IntegrationResolver) ContextIntegrations(ctx context.Context, args struct {
	Filter     *IntegrationFilterInput
	Pagination *PaginationInput
}) (*IntegrationsResultResolver, error) {
	filter := models.IntegrationFilter{}
	if args.Filter != nil {
		// Mapeia os filtros do GraphQL para o modelo de domínio
		if args.Filter.IntegrationIds != nil {
			for _, id := range args.Filter.IntegrationIds {
				parsedID, err := uuid.Parse(string(id))
				if err != nil {
					r.logger.Error("invalid integration ID in filter", "error", err, "id", id)
					return nil, errors.New("ID de integração inválido no filtro")
				}
				filter.IntegrationIDs = append(filter.IntegrationIDs, parsedID)
			}
		}

		if args.Filter.ContextIds != nil {
			for _, id := range args.Filter.ContextIds {
				parsedID, err := uuid.Parse(string(id))
				if err != nil {
					r.logger.Error("invalid context ID in filter", "error", err, "id", id)
					return nil, errors.New("ID de contexto inválido no filtro")
				}
				filter.ContextIDs = append(filter.ContextIDs, parsedID)
			}
		}

		if args.Filter.IntegrationType != nil {
			filter.IntegrationType = models.IntegrationType(*args.Filter.IntegrationType)
		}

		if args.Filter.Status != nil {
			filter.Status = models.IntegrationStatus(*args.Filter.Status)
		}

		if args.Filter.CreatedAfter != nil {
			filter.CreatedAfter = &args.Filter.CreatedAfter.Time
		}

		if args.Filter.CreatedBefore != nil {
			filter.CreatedBefore = &args.Filter.CreatedBefore.Time
		}
	}

	// Configuração da paginação
	pagination := models.Pagination{
		Page:     1,
		PageSize: 20,
	}

	if args.Pagination != nil {
		if args.Pagination.Page != nil {
			pagination.Page = int(*args.Pagination.Page)
		}
		if args.Pagination.PageSize != nil {
			pagination.PageSize = int(*args.Pagination.PageSize)
		}
	}

	// Busca as integrações
	result, err := r.integrationService.ListIntegrations(ctx, filter, pagination)
	if err != nil {
		r.logger.Error("failed to list integrations", "error", err)
		return nil, err
	}

	// Converte o resultado para o formato do GraphQL
	return &IntegrationsResultResolver{
		items:      result.Items,
		totalCount: result.TotalCount,
		hasMore:    result.HasMore,
		resolver:   r,
	}, nil
}

// CreateContextIntegration resolve a mutação para criar uma nova integração entre contextos
func (r *IntegrationResolver) CreateContextIntegration(ctx context.Context, args struct {
	Input CreateContextIntegrationInput
}) (*IntegrationResolver, error) {
	// Valida os dados de entrada
	sourceContextID, err := uuid.Parse(string(args.Input.SourceContextId))
	if err != nil {
		r.logger.Error("invalid source context ID", "error", err, "id", args.Input.SourceContextId)
		return nil, errors.New("ID de contexto de origem inválido")
	}

	targetContextID, err := uuid.Parse(string(args.Input.TargetContextId))
	if err != nil {
		r.logger.Error("invalid target context ID", "error", err, "id", args.Input.TargetContextId)
		return nil, errors.New("ID de contexto de destino inválido")
	}

	if sourceContextID == targetContextID {
		return nil, errors.New("os contextos de origem e destino não podem ser iguais")
	}

	// Verifica se os contextos existem
	sourceContext, err := r.contextService.GetContextByID(ctx, sourceContextID)
	if err != nil || sourceContext == nil {
		r.logger.Error("source context not found", "error", err, "id", sourceContextID)
		return nil, errors.New("contexto de origem não encontrado")
	}

	targetContext, err := r.contextService.GetContextByID(ctx, targetContextID)
	if err != nil || targetContext == nil {
		r.logger.Error("target context not found", "error", err, "id", targetContextID)
		return nil, errors.New("contexto de destino não encontrado")
	}

	// Cria o modelo de domínio
	integration := &models.ContextIntegration{
		SourceContextID:     sourceContextID,
		TargetContextID:     targetContextID,
		IntegrationType:     models.IntegrationType(args.Input.IntegrationType),
		Status:              models.IntegrationStatus(args.Input.Status),
		SyncMode:            models.SyncMode(args.Input.SyncMode),
		SyncInterval:        args.Input.SyncInterval,
		AutoApprove:         args.Input.AutoApprove,
		RequiresUserConsent: args.Input.RequiresUserConsent,
	}

	if args.Input.SyncDirection != nil {
		integration.SyncDirection = models.SyncDirection(*args.Input.SyncDirection)
	} else {
		integration.SyncDirection = models.SyncDirectionBidirectional
	}

	if args.Input.AttributeMappings != nil {
		for _, mapping := range args.Input.AttributeMappings {
			integration.AttributeMappings = append(integration.AttributeMappings, models.AttributeMapping{
				SourceAttribute: mapping.SourceAttribute,
				TargetAttribute: mapping.TargetAttribute,
				Transformation:  mapping.Transformation,
			})
		}
	}

	if args.Input.SyncRules != nil {
		integration.SyncRules = *args.Input.SyncRules
	}

	if args.Input.ValidationRules != nil {
		integration.ValidationRules = *args.Input.ValidationRules
	}

	if args.Input.Description != nil {
		integration.Description = *args.Input.Description
	}

	// Cria a integração
	createdIntegration, err := r.integrationService.CreateIntegration(ctx, integration)
	if err != nil {
		r.logger.Error("failed to create integration", "error", err)
		return nil, err
	}

	return &IntegrationResolver{
		integrationService: r.integrationService,
		contextService:     r.contextService,
		logger:             r.logger,
		integration:        createdIntegration,
	}, nil
}

// UpdateContextIntegration resolve a mutação para atualizar uma integração existente
func (r *IntegrationResolver) UpdateContextIntegration(ctx context.Context, args struct {
	Input UpdateContextIntegrationInput
}) (*IntegrationResolver, error) {
	integrationID, err := uuid.Parse(string(args.Input.IntegrationId))
	if err != nil {
		r.logger.Error("invalid integration ID", "error", err, "id", args.Input.IntegrationId)
		return nil, errors.New("ID de integração inválido")
	}

	// Obtém a integração atual
	existingIntegration, err := r.integrationService.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		r.logger.Error("failed to get integration for update", "error", err, "id", integrationID)
		return nil, err
	}

	if existingIntegration == nil {
		return nil, errors.New("integração não encontrada")
	}

	// Atualiza os campos conforme os inputs
	if args.Input.Status != nil {
		existingIntegration.Status = models.IntegrationStatus(*args.Input.Status)
	}

	if args.Input.SyncMode != nil {
		existingIntegration.SyncMode = models.SyncMode(*args.Input.SyncMode)
	}

	if args.Input.SyncDirection != nil {
		existingIntegration.SyncDirection = models.SyncDirection(*args.Input.SyncDirection)
	}

	if args.Input.SyncInterval != nil {
		existingIntegration.SyncInterval = *args.Input.SyncInterval
	}

	if args.Input.AutoApprove != nil {
		existingIntegration.AutoApprove = *args.Input.AutoApprove
	}

	if args.Input.RequiresUserConsent != nil {
		existingIntegration.RequiresUserConsent = *args.Input.RequiresUserConsent
	}

	if args.Input.AttributeMappings != nil {
		existingIntegration.AttributeMappings = nil
		for _, mapping := range args.Input.AttributeMappings {
			existingIntegration.AttributeMappings = append(existingIntegration.AttributeMappings, models.AttributeMapping{
				SourceAttribute: mapping.SourceAttribute,
				TargetAttribute: mapping.TargetAttribute,
				Transformation:  mapping.Transformation,
			})
		}
	}

	if args.Input.SyncRules != nil {
		existingIntegration.SyncRules = *args.Input.SyncRules
	}

	if args.Input.ValidationRules != nil {
		existingIntegration.ValidationRules = *args.Input.ValidationRules
	}

	if args.Input.Description != nil {
		existingIntegration.Description = *args.Input.Description
	}

	// Atualiza a integração
	updatedIntegration, err := r.integrationService.UpdateIntegration(ctx, existingIntegration)
	if err != nil {
		r.logger.Error("failed to update integration", "error", err, "id", integrationID)
		return nil, err
	}

	return &IntegrationResolver{
		integrationService: r.integrationService,
		contextService:     r.contextService,
		logger:             r.logger,
		integration:        updatedIntegration,
	}, nil
}

// RemoveContextIntegration resolve a mutação para remover uma integração
func (r *IntegrationResolver) RemoveContextIntegration(ctx context.Context, args struct {
	IntegrationID graphql.ID
}) (*OperationResultResolver, error) {
	integrationID, err := uuid.Parse(string(args.IntegrationID))
	if err != nil {
		r.logger.Error("invalid integration ID", "error", err, "id", args.IntegrationID)
		return nil, errors.New("ID de integração inválido")
	}

	// Remove a integração
	err = r.integrationService.RemoveIntegration(ctx, integrationID)
	if err != nil {
		r.logger.Error("failed to remove integration", "error", err, "id", integrationID)
		return nil, err
	}

	return &OperationResultResolver{
		success: true,
		message: "Integração removida com sucesso",
		code:    "INTEGRATION_REMOVED",
	}, nil
}

// SynchronizeContexts resolve a mutação para sincronizar manualmente dois contextos
func (r *IntegrationResolver) SynchronizeContexts(ctx context.Context, args struct {
	IntegrationID graphql.ID
}) (*SyncResultResolver, error) {
	integrationID, err := uuid.Parse(string(args.IntegrationID))
	if err != nil {
		r.logger.Error("invalid integration ID", "error", err, "id", args.IntegrationID)
		return nil, errors.New("ID de integração inválido")
	}

	// Executa a sincronização
	syncResult, err := r.integrationService.SynchronizeContexts(ctx, integrationID)
	if err != nil {
		r.logger.Error("failed to synchronize contexts", "error", err, "id", integrationID)
		return nil, err
	}

	return &SyncResultResolver{
		success:            syncResult.Success,
		message:            syncResult.Message,
		syncedAttributes:   syncResult.SyncedAttributes,
		failedAttributes:   syncResult.FailedAttributes,
		conflictedAttributes: syncResult.ConflictedAttributes,
		syncID:             graphql.ID(syncResult.SyncID.String()),
		timestamp:          graphql.Time{Time: syncResult.Timestamp},
	}, nil
}

// ApproveSync resolve a mutação para aprovar uma sincronização que requer aprovação
func (r *IntegrationResolver) ApproveSync(ctx context.Context, args struct {
	SyncID graphql.ID
	ApprovedAttributes []string
}) (*SyncResultResolver, error) {
	syncID, err := uuid.Parse(string(args.SyncID))
	if err != nil {
		r.logger.Error("invalid sync ID", "error", err, "id", args.SyncID)
		return nil, errors.New("ID de sincronização inválido")
	}

	// Aprova a sincronização
	syncResult, err := r.integrationService.ApproveSync(ctx, syncID, args.ApprovedAttributes)
	if err != nil {
		r.logger.Error("failed to approve sync", "error", err, "id", syncID)
		return nil, err
	}

	return &SyncResultResolver{
		success:            syncResult.Success,
		message:            syncResult.Message,
		syncedAttributes:   syncResult.SyncedAttributes,
		failedAttributes:   syncResult.FailedAttributes,
		conflictedAttributes: syncResult.ConflictedAttributes,
		syncID:             graphql.ID(syncResult.SyncID.String()),
		timestamp:          graphql.Time{Time: syncResult.Timestamp},
	}, nil
}

// IntegrationsResultResolver resolve o resultado paginado de integrações
type IntegrationsResultResolver struct {
	items      []models.ContextIntegration
	totalCount int
	hasMore    bool
	resolver   *IntegrationResolver
}

// Items retorna os itens da página atual
func (r *IntegrationsResultResolver) Items() []*IntegrationResolver {
	resolvers := make([]*IntegrationResolver, len(r.items))
	for i := range r.items {
		resolvers[i] = &IntegrationResolver{
			integrationService: r.resolver.integrationService,
			contextService:     r.resolver.contextService,
			logger:             r.resolver.logger,
			integration:        &r.items[i],
		}
	}
	return resolvers
}

// TotalCount retorna o número total de itens
func (r *IntegrationsResultResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

// HasMore indica se existem mais itens além da página atual
func (r *IntegrationsResultResolver) HasMore() bool {
	return r.hasMore
}

// SyncResultResolver resolve o resultado de uma operação de sincronização
type SyncResultResolver struct {
	success            bool
	message            string
	syncedAttributes   []string
	failedAttributes   []string
	conflictedAttributes map[string]string
	syncID             graphql.ID
	timestamp          graphql.Time
}

// Success retorna se a sincronização foi bem-sucedida
func (r *SyncResultResolver) Success() bool {
	return r.success
}

// Message retorna a mensagem da sincronização
func (r *SyncResultResolver) Message() string {
	return r.message
}

// SyncedAttributes retorna os atributos sincronizados com sucesso
func (r *SyncResultResolver) SyncedAttributes() []string {
	return r.syncedAttributes
}

// FailedAttributes retorna os atributos que falharam na sincronização
func (r *SyncResultResolver) FailedAttributes() []string {
	return r.failedAttributes
}

// ConflictedAttributes retorna os atributos que tiveram conflitos durante a sincronização
func (r *SyncResultResolver) ConflictedAttributes() map[string]string {
	return r.conflictedAttributes
}

// SyncID retorna o ID da operação de sincronização
func (r *SyncResultResolver) SyncID() graphql.ID {
	return r.syncID
}

// Timestamp retorna o timestamp da sincronização
func (r *SyncResultResolver) Timestamp() graphql.Time {
	return r.timestamp
}

// IntegrationFilterInput representa os filtros para busca de integrações
type IntegrationFilterInput struct {
	IntegrationIds []graphql.ID
	ContextIds     []graphql.ID
	IntegrationType *string
	Status         *string
	CreatedAfter   *graphql.Time
	CreatedBefore  *graphql.Time
}

// AttributeMappingInput representa um mapeamento de atributos para integração
type AttributeMappingInput struct {
	SourceAttribute string
	TargetAttribute string
	Transformation  string
}

// CreateContextIntegrationInput representa os dados de entrada para criar uma integração
type CreateContextIntegrationInput struct {
	SourceContextId    graphql.ID
	TargetContextId    graphql.ID
	IntegrationType    string
	Status             string
	SyncMode           string
	SyncDirection      *string
	SyncInterval       int32
	AttributeMappings  []AttributeMappingInput
	SyncRules          *map[string]interface{}
	ValidationRules    *map[string]interface{}
	AutoApprove        bool
	RequiresUserConsent bool
	Description        *string
}

// UpdateContextIntegrationInput representa os dados de entrada para atualizar uma integração
type UpdateContextIntegrationInput struct {
	IntegrationId      graphql.ID
	Status             *string
	SyncMode           *string
	SyncDirection      *string
	SyncInterval       *int32
	AttributeMappings  []AttributeMappingInput
	SyncRules          *map[string]interface{}
	ValidationRules    *map[string]interface{}
	AutoApprove        *bool
	RequiresUserConsent *bool
	Description        *string
}