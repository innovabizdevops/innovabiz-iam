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

// ContextResolver resolve consultas e mutações relacionadas aos contextos de identidade
type ContextResolver struct {
	contextService   *services.ContextService
	attributeService *services.AttributeService
	logger           *logging.Logger
	context          *models.IdentityContext
}

// NewContextResolver cria uma nova instância do resolver de contextos
func NewContextResolver(
	contextService *services.ContextService,
	attributeService *services.AttributeService,
	logger *logging.Logger,
) *ContextResolver {
	return &ContextResolver{
		contextService:   contextService,
		attributeService: attributeService,
		logger:           logger,
	}
}

// IdentityContext resolve a consulta para obter um contexto pelo ID
func (r *ContextResolver) IdentityContext(ctx context.Context, args struct {
	ContextID graphql.ID
}) (*ContextResolver, error) {
	contextID, err := uuid.Parse(string(args.ContextID))
	if err != nil {
		r.logger.Error("invalid context ID format", "error", err, "id", args.ContextID)
		return nil, errors.New("ID de contexto inválido")
	}

	identityContext, err := r.contextService.GetContextByID(ctx, contextID)
	if err != nil {
		r.logger.Error("failed to get context", "error", err, "id", contextID)
		return nil, err
	}

	if identityContext == nil {
		return nil, nil // Retorna null se não encontrar o contexto
	}

	return &ContextResolver{
		contextService:   r.contextService,
		attributeService: r.attributeService,
		logger:           r.logger,
		context:          identityContext,
	}, nil
}

// IdentityContexts resolve a consulta para listar contextos com filtros
func (r *ContextResolver) IdentityContexts(ctx context.Context, args struct {
	Filter     *ContextFilterInput
	Pagination *PaginationInput
}) (*ContextsResultResolver, error) {
	filter := models.ContextFilter{}
	if args.Filter != nil {
		// Mapeia os filtros do GraphQL para o modelo de domínio
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

		if args.Filter.IdentityIds != nil {
			for _, id := range args.Filter.IdentityIds {
				parsedID, err := uuid.Parse(string(id))
				if err != nil {
					r.logger.Error("invalid identity ID in filter", "error", err, "id", id)
					return nil, errors.New("ID de identidade inválido no filtro")
				}
				filter.IdentityIDs = append(filter.IdentityIDs, parsedID)
			}
		}

		if args.Filter.ContextType != nil {
			filter.ContextType = models.ContextType(*args.Filter.ContextType)
		}

		if args.Filter.ContextSubtype != nil {
			filter.ContextSubtype = args.Filter.ContextSubtype
		}

		if args.Filter.ContextStatus != nil {
			filter.ContextStatus = models.ContextStatus(*args.Filter.ContextStatus)
		}

		if args.Filter.VerificationLevel != nil {
			filter.VerificationLevel = models.VerificationLevel(*args.Filter.VerificationLevel)
		}

		if args.Filter.RiskLevel != nil {
			filter.RiskLevel = models.RiskLevel(*args.Filter.RiskLevel)
		}

		if args.Filter.RegionCode != nil {
			filter.RegionCode = args.Filter.RegionCode
		}

		if args.Filter.CountryCode != nil {
			filter.CountryCode = args.Filter.CountryCode
		}

		if args.Filter.LegalFramework != nil {
			filter.LegalFramework = args.Filter.LegalFramework
		}

		if args.Filter.Issuer != nil {
			filter.Issuer = args.Filter.Issuer
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

	// Busca os contextos
	result, err := r.contextService.ListContexts(ctx, filter, pagination)
	if err != nil {
		r.logger.Error("failed to list contexts", "error", err)
		return nil, err
	}

	// Converte o resultado para o formato do GraphQL
	return &ContextsResultResolver{
		items:      result.Items,
		totalCount: result.TotalCount,
		hasMore:    result.HasMore,
		resolver:   r,
	}, nil
}

// CreateIdentityContext resolve a mutação para criar um novo contexto de identidade
func (r *ContextResolver) CreateIdentityContext(ctx context.Context, args struct {
	Input CreateIdentityContextInput
}) (*ContextResolver, error) {
	// Valida os dados de entrada
	identityID, err := uuid.Parse(string(args.Input.IdentityId))
	if err != nil {
		r.logger.Error("invalid identity ID", "error", err, "id", args.Input.IdentityId)
		return nil, errors.New("ID de identidade inválido")
	}

	contextType := models.ContextType(args.Input.ContextType)
	if !contextType.IsValid() {
		r.logger.Error("invalid context type", "type", args.Input.ContextType)
		return nil, errors.New("tipo de contexto inválido")
	}

	// Cria o modelo de domínio
	identityContext := &models.IdentityContext{
		IdentityID:       identityID,
		ContextType:      contextType,
		ContextStatus:    models.ContextStatus(args.Input.ContextStatus),
		VerificationLevel: models.VerificationLevel(args.Input.VerificationLevel),
		RiskLevel:        models.RiskLevel(args.Input.RiskLevel),
	}

	if args.Input.ContextSubtype != nil {
		identityContext.ContextSubtype = *args.Input.ContextSubtype
	}

	if args.Input.TrustScore != nil {
		trustScore := float64(*args.Input.TrustScore)
		identityContext.TrustScore = &trustScore
	}

	if args.Input.RegionCode != nil {
		identityContext.RegionCode = *args.Input.RegionCode
	}

	if args.Input.CountryCode != nil {
		identityContext.CountryCode = *args.Input.CountryCode
	}

	if args.Input.LegalFramework != nil {
		identityContext.LegalFramework = *args.Input.LegalFramework
	}

	if args.Input.Issuer != nil {
		identityContext.Issuer = *args.Input.Issuer
	}

	if args.Input.ExpiresAt != nil {
		identityContext.ExpiresAt = &args.Input.ExpiresAt.Time
	}

	if args.Input.Metadata != nil {
		identityContext.Metadata = *args.Input.Metadata
	}

	// Cria o contexto
	createdContext, err := r.contextService.CreateContext(ctx, identityContext)
	if err != nil {
		r.logger.Error("failed to create context", "error", err)
		return nil, err
	}

	return &ContextResolver{
		contextService:   r.contextService,
		attributeService: r.attributeService,
		logger:           r.logger,
		context:          createdContext,
	}, nil
}

// UpdateIdentityContext resolve a mutação para atualizar um contexto existente
func (r *ContextResolver) UpdateIdentityContext(ctx context.Context, args struct {
	Input UpdateIdentityContextInput
}) (*ContextResolver, error) {
	contextID, err := uuid.Parse(string(args.Input.ContextId))
	if err != nil {
		r.logger.Error("invalid context ID", "error", err, "id", args.Input.ContextId)
		return nil, errors.New("ID de contexto inválido")
	}

	// Obtém o contexto atual
	existingContext, err := r.contextService.GetContextByID(ctx, contextID)
	if err != nil {
		r.logger.Error("failed to get context for update", "error", err, "id", contextID)
		return nil, err
	}

	if existingContext == nil {
		return nil, errors.New("contexto não encontrado")
	}

	// Atualiza os campos conforme os inputs
	if args.Input.ContextStatus != nil {
		existingContext.ContextStatus = models.ContextStatus(*args.Input.ContextStatus)
	}

	if args.Input.TrustScore != nil {
		trustScore := float64(*args.Input.TrustScore)
		existingContext.TrustScore = &trustScore
	}

	if args.Input.VerificationLevel != nil {
		existingContext.VerificationLevel = models.VerificationLevel(*args.Input.VerificationLevel)
	}

	if args.Input.RiskLevel != nil {
		existingContext.RiskLevel = models.RiskLevel(*args.Input.RiskLevel)
	}

	if args.Input.ExpiresAt != nil {
		existingContext.ExpiresAt = &args.Input.ExpiresAt.Time
	}

	if args.Input.LastVerifiedAt != nil {
		existingContext.LastVerifiedAt = &args.Input.LastVerifiedAt.Time
	}

	if args.Input.Metadata != nil {
		existingContext.Metadata = *args.Input.Metadata
	}

	// Atualiza o contexto
	updatedContext, err := r.contextService.UpdateContext(ctx, existingContext)
	if err != nil {
		r.logger.Error("failed to update context", "error", err, "id", contextID)
		return nil, err
	}

	return &ContextResolver{
		contextService:   r.contextService,
		attributeService: r.attributeService,
		logger:           r.logger,
		context:          updatedContext,
	}, nil
}

// RemoveIdentityContext resolve a mutação para remover um contexto de uma identidade
func (r *ContextResolver) RemoveIdentityContext(ctx context.Context, args struct {
	ContextID graphql.ID
}) (*OperationResultResolver, error) {
	contextID, err := uuid.Parse(string(args.ContextID))
	if err != nil {
		r.logger.Error("invalid context ID", "error", err, "id", args.ContextID)
		return nil, errors.New("ID de contexto inválido")
	}

	// Remove o contexto
	err = r.contextService.RemoveContext(ctx, contextID)
	if err != nil {
		r.logger.Error("failed to remove context", "error", err, "id", contextID)
		return nil, err
	}

	return &OperationResultResolver{
		success: true,
		message: "Contexto removido com sucesso",
		code:    "CONTEXT_REMOVED",
	}, nil
}

// VerifyIdentityWithTrustGuard resolve a mutação para verificar uma identidade usando o TrustGuard
func (r *ContextResolver) VerifyIdentityWithTrustGuard(ctx context.Context, args struct {
	IdentityID       graphql.ID
	ContextType      string
	VerificationLevel string
}) (*IdentityVerificationResultResolver, error) {
	identityID, err := uuid.Parse(string(args.IdentityID))
	if err != nil {
		r.logger.Error("invalid identity ID", "error", err, "id", args.IdentityID)
		return nil, errors.New("ID de identidade inválido")
	}

	contextType := models.ContextType(args.ContextType)
	if !contextType.IsValid() {
		r.logger.Error("invalid context type", "type", args.ContextType)
		return nil, errors.New("tipo de contexto inválido")
	}

	verificationLevel := models.VerificationLevel(args.VerificationLevel)
	if !verificationLevel.IsValid() {
		r.logger.Error("invalid verification level", "level", args.VerificationLevel)
		return nil, errors.New("nível de verificação inválido")
	}

	// Verifica a identidade
	verificationResult, err := r.contextService.VerifyIdentityWithTrustGuard(ctx, identityID, contextType, verificationLevel)
	if err != nil {
		r.logger.Error("failed to verify identity with TrustGuard", "error", err, "id", identityID)
		return nil, err
	}

	return &IdentityVerificationResultResolver{
		identityID:         graphql.ID(verificationResult.IdentityID.String()),
		verificationID:     graphql.ID(verificationResult.VerificationID.String()),
		verificationStatus: string(verificationResult.VerificationStatus),
		trustScore:         float64(verificationResult.TrustScore),
		verificationDetails: verificationResult.VerificationDetails,
		timestamp:          graphql.Time{Time: verificationResult.Timestamp},
	}, nil
}

// ContextsResultResolver resolve o resultado paginado de contextos
type ContextsResultResolver struct {
	items      []models.IdentityContext
	totalCount int
	hasMore    bool
	resolver   *ContextResolver
}

// Items retorna os itens da página atual
func (r *ContextsResultResolver) Items() []*ContextResolver {
	resolvers := make([]*ContextResolver, len(r.items))
	for i := range r.items {
		resolvers[i] = &ContextResolver{
			contextService:   r.resolver.contextService,
			attributeService: r.resolver.attributeService,
			logger:           r.resolver.logger,
			context:          &r.items[i],
		}
	}
	return resolvers
}

// TotalCount retorna o número total de itens
func (r *ContextsResultResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

// HasMore indica se existem mais itens além da página atual
func (r *ContextsResultResolver) HasMore() bool {
	return r.hasMore
}

// OperationResultResolver resolve o resultado de operações
type OperationResultResolver struct {
	success bool
	message string
	code    string
}

// Success retorna se a operação foi bem-sucedida
func (r *OperationResultResolver) Success() bool {
	return r.success
}

// Message retorna a mensagem da operação
func (r *OperationResultResolver) Message() *string {
	if r.message == "" {
		return nil
	}
	return &r.message
}

// Code retorna o código da operação
func (r *OperationResultResolver) Code() *string {
	if r.code == "" {
		return nil
	}
	return &r.code
}

// IdentityVerificationResultResolver resolve o resultado de verificação de identidade
type IdentityVerificationResultResolver struct {
	identityID         graphql.ID
	verificationID     graphql.ID
	verificationStatus string
	trustScore         float64
	verificationDetails map[string]interface{}
	timestamp          graphql.Time
}

// IdentityID retorna o ID da identidade verificada
func (r *IdentityVerificationResultResolver) IdentityID() graphql.ID {
	return r.identityID
}

// VerificationID retorna o ID da verificação
func (r *IdentityVerificationResultResolver) VerificationID() graphql.ID {
	return r.verificationID
}

// VerificationStatus retorna o status da verificação
func (r *IdentityVerificationResultResolver) VerificationStatus() string {
	return r.verificationStatus
}

// TrustScore retorna a pontuação de confiabilidade
func (r *IdentityVerificationResultResolver) TrustScore() float64 {
	return r.trustScore
}

// VerificationDetails retorna os detalhes da verificação
func (r *IdentityVerificationResultResolver) VerificationDetails() map[string]interface{} {
	return r.verificationDetails
}

// Timestamp retorna o timestamp da verificação
func (r *IdentityVerificationResultResolver) Timestamp() graphql.Time {
	return r.timestamp
}

// ContextFilterInput representa os filtros para busca de contextos
type ContextFilterInput struct {
	ContextIds       []graphql.ID
	IdentityIds      []graphql.ID
	ContextType      *string
	ContextSubtype   *string
	ContextStatus    *string
	VerificationLevel *string
	RiskLevel        *string
	RegionCode       *string
	CountryCode      *string
	LegalFramework   *string
	Issuer           *string
	CreatedAfter     *graphql.Time
	CreatedBefore    *graphql.Time
}

// CreateIdentityContextInput representa os dados de entrada para criar um contexto de identidade
type CreateIdentityContextInput struct {
	IdentityId        graphql.ID
	ContextType       string
	ContextSubtype    *string
	ContextStatus     string
	TrustScore        *float64
	VerificationLevel string
	RiskLevel         string
	RegionCode        *string
	CountryCode       *string
	LegalFramework    *string
	Issuer            *string
	ExpiresAt         *graphql.Time
	Metadata          *map[string]interface{}
}

// UpdateIdentityContextInput representa os dados de entrada para atualizar um contexto de identidade
type UpdateIdentityContextInput struct {
	ContextId        graphql.ID
	ContextStatus    *string
	TrustScore       *float64
	VerificationLevel *string
	RiskLevel        *string
	ExpiresAt        *graphql.Time
	LastVerifiedAt   *graphql.Time
	Metadata         *map[string]interface{}
}