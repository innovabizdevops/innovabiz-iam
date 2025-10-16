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

// IdentityResolver resolve consultas e mutações relacionadas às identidades multi-contexto
type IdentityResolver struct {
	identityService *services.IdentityService
	contextService  *services.ContextService
	logger          *logging.Logger
}

// NewIdentityResolver cria uma nova instância do resolver de identidades
func NewIdentityResolver(
	identityService *services.IdentityService,
	contextService *services.ContextService,
	logger *logging.Logger,
) *IdentityResolver {
	return &IdentityResolver{
		identityService: identityService,
		contextService:  contextService,
		logger:          logger,
	}
}

// Identity resolve a consulta para obter uma identidade pelo ID
func (r *IdentityResolver) Identity(ctx context.Context, args struct {
	IdentityID graphql.ID
}) (*IdentityResolver, error) {
	identityID, err := uuid.Parse(string(args.IdentityID))
	if err != nil {
		r.logger.Error("invalid identity ID format", "error", err, "id", args.IdentityID)
		return nil, errors.New("ID de identidade inválido")
	}

	identity, err := r.identityService.GetIdentityByID(ctx, identityID)
	if err != nil {
		r.logger.Error("failed to get identity", "error", err, "id", identityID)
		return nil, err
	}

	if identity == nil {
		return nil, nil // Retorna null se não encontrar a identidade
	}

	return &IdentityResolver{
		identityService: r.identityService,
		contextService:  r.contextService,
		logger:          r.logger,
	}, nil
}

// Identities resolve a consulta para listar identidades com filtros
func (r *IdentityResolver) Identities(ctx context.Context, args struct {
	Filter     *IdentityFilterInput
	Pagination *PaginationInput
}) (*IdentitiesResultResolver, error) {
	filter := models.IdentityFilter{}
	if args.Filter != nil {
		// Mapeia os filtros do GraphQL para o modelo de domínio
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

		if args.Filter.PrimaryKeyType != nil {
			filter.PrimaryKeyType = models.PrimaryKeyType(*args.Filter.PrimaryKeyType)
		}

		if args.Filter.PrimaryKeyValue != nil {
			filter.PrimaryKeyValue = *args.Filter.PrimaryKeyValue
		}

		if args.Filter.Status != nil {
			filter.Status = models.IdentityStatus(*args.Filter.Status)
		}

		if args.Filter.TrustLevel != nil {
			filter.TrustLevel = models.TrustLevel(*args.Filter.TrustLevel)
		}

		if args.Filter.MasterPersonId != nil {
			masterPersonID, err := uuid.Parse(string(*args.Filter.MasterPersonId))
			if err != nil {
				r.logger.Error("invalid master person ID in filter", "error", err, "id", *args.Filter.MasterPersonId)
				return nil, errors.New("ID de pessoa mestra inválido no filtro")
			}
			filter.MasterPersonID = &masterPersonID
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
			pagination.Page = *args.Pagination.Page
		}
		if args.Pagination.PageSize != nil {
			pagination.PageSize = *args.Pagination.PageSize
		}
	}

	// Busca as identidades
	result, err := r.identityService.ListIdentities(ctx, filter, pagination)
	if err != nil {
		r.logger.Error("failed to list identities", "error", err)
		return nil, err
	}

	// Converte o resultado para o formato do GraphQL
	return &IdentitiesResultResolver{
		items:      result.Items,
		totalCount: result.TotalCount,
		hasMore:    result.HasMore,
		resolver:   r,
	}, nil
}

// IdentityByPrimaryKey resolve a consulta para obter uma identidade pelo tipo e valor da chave primária
func (r *IdentityResolver) IdentityByPrimaryKey(ctx context.Context, args struct {
	Type  string
	Value string
}) (*IdentityResolver, error) {
	keyType := models.PrimaryKeyType(args.Type)
	if !keyType.IsValid() {
		r.logger.Error("invalid primary key type", "type", args.Type)
		return nil, errors.New("tipo de chave primária inválido")
	}

	identity, err := r.identityService.GetIdentityByPrimaryKey(ctx, keyType, args.Value)
	if err != nil {
		r.logger.Error("failed to get identity by primary key", "error", err, "type", args.Type, "value", args.Value)
		return nil, err
	}

	if identity == nil {
		return nil, nil // Retorna null se não encontrar a identidade
	}

	return &IdentityResolver{
		identityService: r.identityService,
		contextService:  r.contextService,
		logger:          r.logger,
	}, nil
}

// VerifyIdentityInContext verifica se uma identidade existe em um contexto específico
func (r *IdentityResolver) VerifyIdentityInContext(ctx context.Context, args struct {
	PrimaryKeyType  string
	PrimaryKeyValue string
	ContextType     string
}) (bool, error) {
	keyType := models.PrimaryKeyType(args.PrimaryKeyType)
	if !keyType.IsValid() {
		r.logger.Error("invalid primary key type", "type", args.PrimaryKeyType)
		return false, errors.New("tipo de chave primária inválido")
	}

	contextType := models.ContextType(args.ContextType)
	if !contextType.IsValid() {
		r.logger.Error("invalid context type", "type", args.ContextType)
		return false, errors.New("tipo de contexto inválido")
	}

	exists, err := r.identityService.VerifyIdentityInContext(ctx, keyType, args.PrimaryKeyValue, contextType)
	if err != nil {
		r.logger.Error("failed to verify identity in context", "error", err, "keyType", args.PrimaryKeyType, "keyValue", args.PrimaryKeyValue, "contextType", args.ContextType)
		return false, err
	}

	return exists, nil
}

// CreateIdentity resolve a mutação para criar uma nova identidade
func (r *IdentityResolver) CreateIdentity(ctx context.Context, args struct {
	Input CreateIdentityInput
}) (*IdentityResolver, error) {
	// Valida os dados de entrada
	if args.Input.PrimaryKeyType == "" || args.Input.PrimaryKeyValue == "" {
		return nil, errors.New("tipo e valor da chave primária são obrigatórios")
	}

	// Cria o modelo de domínio
	identity := &models.Identity{
		PrimaryKeyType:  models.PrimaryKeyType(args.Input.PrimaryKeyType),
		PrimaryKeyValue: args.Input.PrimaryKeyValue,
		Status:          models.IdentityStatus(args.Input.Status),
		TrustLevel:      models.TrustLevel(args.Input.TrustLevel),
	}

	if args.Input.MasterPersonId != nil {
		masterPersonID, err := uuid.Parse(string(*args.Input.MasterPersonId))
		if err != nil {
			r.logger.Error("invalid master person ID", "error", err, "id", *args.Input.MasterPersonId)
			return nil, errors.New("ID de pessoa mestra inválido")
		}
		identity.MasterPersonID = &masterPersonID
	}

	if args.Input.Metadata != nil {
		identity.Metadata = *args.Input.Metadata
	}

	// Cria a identidade
	createdIdentity, err := r.identityService.CreateIdentity(ctx, identity)
	if err != nil {
		r.logger.Error("failed to create identity", "error", err)
		return nil, err
	}

	return &IdentityResolver{
		identityService: r.identityService,
		contextService:  r.contextService,
		logger:          r.logger,
	}, nil
}

// UpdateIdentity resolve a mutação para atualizar uma identidade existente
func (r *IdentityResolver) UpdateIdentity(ctx context.Context, args struct {
	Input UpdateIdentityInput
}) (*IdentityResolver, error) {
	identityID, err := uuid.Parse(string(args.Input.IdentityId))
	if err != nil {
		r.logger.Error("invalid identity ID", "error", err, "id", args.Input.IdentityId)
		return nil, errors.New("ID de identidade inválido")
	}

	// Obtém a identidade atual
	existingIdentity, err := r.identityService.GetIdentityByID(ctx, identityID)
	if err != nil {
		r.logger.Error("failed to get identity for update", "error", err, "id", identityID)
		return nil, err
	}

	if existingIdentity == nil {
		return nil, errors.New("identidade não encontrada")
	}

	// Atualiza os campos conforme os inputs
	if args.Input.Status != nil {
		existingIdentity.Status = models.IdentityStatus(*args.Input.Status)
	}

	if args.Input.TrustLevel != nil {
		existingIdentity.TrustLevel = models.TrustLevel(*args.Input.TrustLevel)
	}

	if args.Input.MasterPersonId != nil {
		masterPersonID, err := uuid.Parse(string(*args.Input.MasterPersonId))
		if err != nil {
			r.logger.Error("invalid master person ID", "error", err, "id", *args.Input.MasterPersonId)
			return nil, errors.New("ID de pessoa mestra inválido")
		}
		existingIdentity.MasterPersonID = &masterPersonID
	}

	if args.Input.Metadata != nil {
		existingIdentity.Metadata = *args.Input.Metadata
	}

	// Atualiza a identidade
	updatedIdentity, err := r.identityService.UpdateIdentity(ctx, existingIdentity)
	if err != nil {
		r.logger.Error("failed to update identity", "error", err, "id", identityID)
		return nil, err
	}

	return &IdentityResolver{
		identityService: r.identityService,
		contextService:  r.contextService,
		logger:          r.logger,
	}, nil
}

// IdentitiesResultResolver resolve o resultado paginado de identidades
type IdentitiesResultResolver struct {
	items      []models.Identity
	totalCount int
	hasMore    bool
	resolver   *IdentityResolver
}

// Items retorna os itens da página atual
func (r *IdentitiesResultResolver) Items() []*IdentityResolver {
	resolvers := make([]*IdentityResolver, len(r.items))
	for i := range r.items {
		resolvers[i] = &IdentityResolver{
			identityService: r.resolver.identityService,
			contextService:  r.resolver.contextService,
			logger:          r.resolver.logger,
		}
	}
	return resolvers
}

// TotalCount retorna o número total de itens
func (r *IdentitiesResultResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

// HasMore indica se existem mais itens além da página atual
func (r *IdentitiesResultResolver) HasMore() bool {
	return r.hasMore
}

// IdentityFilterInput representa os filtros para busca de identidades
type IdentityFilterInput struct {
	IdentityIds     []graphql.ID
	PrimaryKeyType  *string
	PrimaryKeyValue *string
	Status          *string
	TrustLevel      *string
	MasterPersonId  *graphql.ID
	CreatedAfter    *graphql.Time
	CreatedBefore   *graphql.Time
}

// PaginationInput representa as opções de paginação
type PaginationInput struct {
	Page     *int32
	PageSize *int32
}

// CreateIdentityInput representa os dados de entrada para criar uma identidade
type CreateIdentityInput struct {
	PrimaryKeyType  string
	PrimaryKeyValue string
	MasterPersonId  *graphql.ID
	Status          string
	TrustLevel      string
	Metadata        *map[string]interface{}
}

// UpdateIdentityInput representa os dados de entrada para atualizar uma identidade
type UpdateIdentityInput struct {
	IdentityId     graphql.ID
	Status         *string
	TrustLevel     *string
	MasterPersonId *graphql.ID
	Metadata       *map[string]interface{}
}