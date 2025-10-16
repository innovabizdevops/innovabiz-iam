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

// AttributeResolver resolve consultas e mutações relacionadas aos atributos de contexto
type AttributeResolver struct {
	attributeService *services.AttributeService
	logger           *logging.Logger
	attribute        *models.ContextAttribute
}

// NewAttributeResolver cria uma nova instância do resolver de atributos
func NewAttributeResolver(
	attributeService *services.AttributeService,
	logger *logging.Logger,
) *AttributeResolver {
	return &AttributeResolver{
		attributeService: attributeService,
		logger:           logger,
	}
}

// ContextAttribute resolve a consulta para obter um atributo pelo ID
func (r *AttributeResolver) ContextAttribute(ctx context.Context, args struct {
	AttributeID graphql.ID
}) (*AttributeResolver, error) {
	attributeID, err := uuid.Parse(string(args.AttributeID))
	if err != nil {
		r.logger.Error("invalid attribute ID format", "error", err, "id", args.AttributeID)
		return nil, errors.New("ID de atributo inválido")
	}

	attribute, err := r.attributeService.GetAttributeByID(ctx, attributeID)
	if err != nil {
		r.logger.Error("failed to get attribute", "error", err, "id", attributeID)
		return nil, err
	}

	if attribute == nil {
		return nil, nil // Retorna null se não encontrar o atributo
	}

	return &AttributeResolver{
		attributeService: r.attributeService,
		logger:           r.logger,
		attribute:        attribute,
	}, nil
}

// ContextAttributes resolve a consulta para listar atributos com filtros
func (r *AttributeResolver) ContextAttributes(ctx context.Context, args struct {
	Filter     *AttributeFilterInput
	Pagination *PaginationInput
}) (*AttributesResultResolver, error) {
	filter := models.AttributeFilter{}
	if args.Filter != nil {
		// Mapeia os filtros do GraphQL para o modelo de domínio
		if args.Filter.AttributeIds != nil {
			for _, id := range args.Filter.AttributeIds {
				parsedID, err := uuid.Parse(string(id))
				if err != nil {
					r.logger.Error("invalid attribute ID in filter", "error", err, "id", id)
					return nil, errors.New("ID de atributo inválido no filtro")
				}
				filter.AttributeIDs = append(filter.AttributeIDs, parsedID)
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

		if args.Filter.AttributeKey != nil {
			filter.AttributeKey = *args.Filter.AttributeKey
		}

		if args.Filter.SensitivityLevel != nil {
			filter.SensitivityLevel = models.SensitivityLevel(*args.Filter.SensitivityLevel)
		}

		if args.Filter.VerificationStatus != nil {
			filter.VerificationStatus = models.VerificationStatus(*args.Filter.VerificationStatus)
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

	// Busca os atributos
	result, err := r.attributeService.ListAttributes(ctx, filter, pagination)
	if err != nil {
		r.logger.Error("failed to list attributes", "error", err)
		return nil, err
	}

	// Converte o resultado para o formato do GraphQL
	return &AttributesResultResolver{
		items:      result.Items,
		totalCount: result.TotalCount,
		hasMore:    result.HasMore,
		resolver:   r,
	}, nil
}

// CreateContextAttribute resolve a mutação para criar um novo atributo de contexto
func (r *AttributeResolver) CreateContextAttribute(ctx context.Context, args struct {
	Input CreateContextAttributeInput
}) (*AttributeResolver, error) {
	// Valida os dados de entrada
	contextID, err := uuid.Parse(string(args.Input.ContextId))
	if err != nil {
		r.logger.Error("invalid context ID", "error", err, "id", args.Input.ContextId)
		return nil, errors.New("ID de contexto inválido")
	}

	if args.Input.AttributeKey == "" {
		return nil, errors.New("chave de atributo é obrigatória")
	}

	// Cria o modelo de domínio
	attribute := &models.ContextAttribute{
		ContextID:         contextID,
		AttributeKey:      args.Input.AttributeKey,
		SensitivityLevel:  models.SensitivityLevel(args.Input.SensitivityLevel),
		VerificationStatus: models.VerificationStatus(args.Input.VerificationStatus),
		IsRequired:        args.Input.IsRequired,
		IsMutable:         args.Input.IsMutable,
	}

	if args.Input.AttributeValue != nil {
		attribute.AttributeValue = *args.Input.AttributeValue
	}

	if args.Input.VerificationSource != nil {
		attribute.VerificationSource = *args.Input.VerificationSource
	}

	if args.Input.VerificationTimestamp != nil {
		attribute.VerificationTimestamp = &args.Input.VerificationTimestamp.Time
	}

	if args.Input.ExpirationDate != nil {
		attribute.ExpirationDate = &args.Input.ExpirationDate.Time
	}

	if args.Input.Metadata != nil {
		attribute.Metadata = *args.Input.Metadata
	}

	// Cria o atributo
	createdAttribute, err := r.attributeService.CreateAttribute(ctx, attribute)
	if err != nil {
		r.logger.Error("failed to create attribute", "error", err)
		return nil, err
	}

	return &AttributeResolver{
		attributeService: r.attributeService,
		logger:           r.logger,
		attribute:        createdAttribute,
	}, nil
}

// UpdateContextAttribute resolve a mutação para atualizar um atributo existente
func (r *AttributeResolver) UpdateContextAttribute(ctx context.Context, args struct {
	Input UpdateContextAttributeInput
}) (*AttributeResolver, error) {
	attributeID, err := uuid.Parse(string(args.Input.AttributeId))
	if err != nil {
		r.logger.Error("invalid attribute ID", "error", err, "id", args.Input.AttributeId)
		return nil, errors.New("ID de atributo inválido")
	}

	// Obtém o atributo atual
	existingAttribute, err := r.attributeService.GetAttributeByID(ctx, attributeID)
	if err != nil {
		r.logger.Error("failed to get attribute for update", "error", err, "id", attributeID)
		return nil, err
	}

	if existingAttribute == nil {
		return nil, errors.New("atributo não encontrado")
	}

	// Atualiza os campos conforme os inputs
	if args.Input.AttributeValue != nil {
		existingAttribute.AttributeValue = *args.Input.AttributeValue
	}

	if args.Input.SensitivityLevel != nil {
		existingAttribute.SensitivityLevel = models.SensitivityLevel(*args.Input.SensitivityLevel)
	}

	if args.Input.VerificationStatus != nil {
		existingAttribute.VerificationStatus = models.VerificationStatus(*args.Input.VerificationStatus)
	}

	if args.Input.VerificationSource != nil {
		existingAttribute.VerificationSource = *args.Input.VerificationSource
	}

	if args.Input.VerificationTimestamp != nil {
		existingAttribute.VerificationTimestamp = &args.Input.VerificationTimestamp.Time
	}

	if args.Input.ExpirationDate != nil {
		existingAttribute.ExpirationDate = &args.Input.ExpirationDate.Time
	}

	if args.Input.IsRequired != nil {
		existingAttribute.IsRequired = *args.Input.IsRequired
	}

	if args.Input.IsMutable != nil {
		existingAttribute.IsMutable = *args.Input.IsMutable
	}

	if args.Input.VerifiedBy != nil {
		verifiedBy, err := uuid.Parse(string(*args.Input.VerifiedBy))
		if err != nil {
			r.logger.Error("invalid verified by ID", "error", err, "id", *args.Input.VerifiedBy)
			return nil, errors.New("ID de verificador inválido")
		}
		existingAttribute.VerifiedBy = &verifiedBy
	}

	if args.Input.Metadata != nil {
		existingAttribute.Metadata = *args.Input.Metadata
	}

	// Atualiza o atributo
	updatedAttribute, err := r.attributeService.UpdateAttribute(ctx, existingAttribute)
	if err != nil {
		r.logger.Error("failed to update attribute", "error", err, "id", attributeID)
		return nil, err
	}

	return &AttributeResolver{
		attributeService: r.attributeService,
		logger:           r.logger,
		attribute:        updatedAttribute,
	}, nil
}

// RemoveContextAttribute resolve a mutação para remover um atributo de contexto
func (r *AttributeResolver) RemoveContextAttribute(ctx context.Context, args struct {
	AttributeID graphql.ID
}) (*OperationResultResolver, error) {
	attributeID, err := uuid.Parse(string(args.AttributeID))
	if err != nil {
		r.logger.Error("invalid attribute ID", "error", err, "id", args.AttributeID)
		return nil, errors.New("ID de atributo inválido")
	}

	// Remove o atributo
	err = r.attributeService.RemoveAttribute(ctx, attributeID)
	if err != nil {
		r.logger.Error("failed to remove attribute", "error", err, "id", attributeID)
		return nil, err
	}

	return &OperationResultResolver{
		success: true,
		message: "Atributo removido com sucesso",
		code:    "ATTRIBUTE_REMOVED",
	}, nil
}

// CreateAttributeMapping resolve a mutação para criar um mapeamento entre atributos
func (r *AttributeResolver) CreateAttributeMapping(ctx context.Context, args struct {
	Input CreateAttributeMappingInput
}) (*AttributeMappingResolver, error) {
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

	if args.Input.SourceAttributeKey == "" || args.Input.TargetAttributeKey == "" {
		return nil, errors.New("chaves de atributo de origem e destino são obrigatórias")
	}

	// Cria o modelo de domínio
	mapping := &models.AttributeContextMapping{
		SourceContextID:    sourceContextID,
		TargetContextID:    targetContextID,
		SourceAttributeKey: args.Input.SourceAttributeKey,
		TargetAttributeKey: args.Input.TargetAttributeKey,
		MappingType:        models.MappingType(args.Input.MappingType),
		IsActive:           args.Input.IsActive,
	}

	if args.Input.TransformationRule != nil {
		mapping.TransformationRule = *args.Input.TransformationRule
	}

	// Cria o mapeamento
	createdMapping, err := r.attributeService.CreateAttributeMapping(ctx, mapping)
	if err != nil {
		r.logger.Error("failed to create attribute mapping", "error", err)
		return nil, err
	}

	return &AttributeMappingResolver{
		mapping: createdMapping,
	}, nil
}

// RemoveAttributeMapping resolve a mutação para remover um mapeamento de atributos
func (r *AttributeResolver) RemoveAttributeMapping(ctx context.Context, args struct {
	MappingID graphql.ID
}) (*OperationResultResolver, error) {
	mappingID, err := uuid.Parse(string(args.MappingID))
	if err != nil {
		r.logger.Error("invalid mapping ID", "error", err, "id", args.MappingID)
		return nil, errors.New("ID de mapeamento inválido")
	}

	// Remove o mapeamento
	err = r.attributeService.RemoveAttributeMapping(ctx, mappingID)
	if err != nil {
		r.logger.Error("failed to remove attribute mapping", "error", err, "id", mappingID)
		return nil, err
	}

	return &OperationResultResolver{
		success: true,
		message: "Mapeamento removido com sucesso",
		code:    "MAPPING_REMOVED",
	}, nil
}

// AttributesResultResolver resolve o resultado paginado de atributos
type AttributesResultResolver struct {
	items      []models.ContextAttribute
	totalCount int
	hasMore    bool
	resolver   *AttributeResolver
}

// Items retorna os itens da página atual
func (r *AttributesResultResolver) Items() []*AttributeResolver {
	resolvers := make([]*AttributeResolver, len(r.items))
	for i := range r.items {
		resolvers[i] = &AttributeResolver{
			attributeService: r.resolver.attributeService,
			logger:           r.resolver.logger,
			attribute:        &r.items[i],
		}
	}
	return resolvers
}

// TotalCount retorna o número total de itens
func (r *AttributesResultResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

// HasMore indica se existem mais itens além da página atual
func (r *AttributesResultResolver) HasMore() bool {
	return r.hasMore
}

// AttributeFilterInput representa os filtros para busca de atributos
type AttributeFilterInput struct {
	AttributeIds      []graphql.ID
	ContextIds        []graphql.ID
	AttributeKey      *string
	SensitivityLevel  *string
	VerificationStatus *string
	CreatedAfter      *graphql.Time
	CreatedBefore     *graphql.Time
}

// CreateContextAttributeInput representa os dados de entrada para criar um atributo de contexto
type CreateContextAttributeInput struct {
	ContextId            graphql.ID
	AttributeKey         string
	AttributeValue       *string
	SensitivityLevel     string
	VerificationStatus   string
	VerificationSource   *string
	VerificationTimestamp *graphql.Time
	ExpirationDate       *graphql.Time
	IsRequired           bool
	IsMutable            bool
	Metadata             *map[string]interface{}
}

// UpdateContextAttributeInput representa os dados de entrada para atualizar um atributo de contexto
type UpdateContextAttributeInput struct {
	AttributeId          graphql.ID
	AttributeValue       *string
	SensitivityLevel     *string
	VerificationStatus   *string
	VerificationSource   *string
	VerificationTimestamp *graphql.Time
	ExpirationDate       *graphql.Time
	IsRequired           *bool
	IsMutable            *bool
	VerifiedBy           *graphql.ID
	Metadata             *map[string]interface{}
}

// AttributeMappingResolver resolve consultas e mutações relacionadas aos mapeamentos de atributos
type AttributeMappingResolver struct {
	mapping *models.AttributeContextMapping
}

// MappingID retorna o ID do mapeamento
func (r *AttributeMappingResolver) MappingID() graphql.ID {
	return graphql.ID(r.mapping.MappingID.String())
}

// SourceContextID retorna o ID do contexto de origem
func (r *AttributeMappingResolver) SourceContextID() graphql.ID {
	return graphql.ID(r.mapping.SourceContextID.String())
}

// TargetContextID retorna o ID do contexto de destino
func (r *AttributeMappingResolver) TargetContextID() graphql.ID {
	return graphql.ID(r.mapping.TargetContextID.String())
}

// SourceAttributeKey retorna a chave do atributo de origem
func (r *AttributeMappingResolver) SourceAttributeKey() string {
	return r.mapping.SourceAttributeKey
}

// TargetAttributeKey retorna a chave do atributo de destino
func (r *AttributeMappingResolver) TargetAttributeKey() string {
	return r.mapping.TargetAttributeKey
}

// MappingType retorna o tipo de mapeamento
func (r *AttributeMappingResolver) MappingType() string {
	return string(r.mapping.MappingType)
}

// TransformationRule retorna a regra de transformação
func (r *AttributeMappingResolver) TransformationRule() *string {
	if r.mapping.TransformationRule == "" {
		return nil
	}
	return &r.mapping.TransformationRule
}

// IsActive retorna se o mapeamento está ativo
func (r *AttributeMappingResolver) IsActive() bool {
	return r.mapping.IsActive
}

// CreatedAt retorna a data de criação do mapeamento
func (r *AttributeMappingResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.mapping.CreatedAt}
}

// UpdatedAt retorna a data de atualização do mapeamento
func (r *AttributeMappingResolver) UpdatedAt() graphql.Time {
	return graphql.Time{Time: r.mapping.UpdatedAt}
}

// CreatedBy retorna quem criou o mapeamento
func (r *AttributeMappingResolver) CreatedBy() *graphql.ID {
	if r.mapping.CreatedBy == nil {
		return nil
	}
	id := graphql.ID(r.mapping.CreatedBy.String())
	return &id
}

// CreateAttributeMappingInput representa os dados de entrada para criar um mapeamento de atributos
type CreateAttributeMappingInput struct {
	SourceContextId    graphql.ID
	TargetContextId    graphql.ID
	SourceAttributeKey string
	TargetAttributeKey string
	MappingType        string
	TransformationRule *string
	IsActive           bool
}