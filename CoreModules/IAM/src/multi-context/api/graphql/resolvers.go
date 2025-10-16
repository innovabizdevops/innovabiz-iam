/**
 * @file resolvers.go
 * @description Configuração principal dos resolvers GraphQL para o serviço multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package graphql

import (
	"io/ioutil"
	"path/filepath"
	
	"github.com/graph-gophers/graphql-go"
	
	"innovabiz/iam/src/multi-context/api/graphql/resolvers"
	"innovabiz/iam/src/multi-context/application/commands"
	"innovabiz/iam/src/multi-context/application/queries"
	"innovabiz/iam/src/multi-context/domain/services"
)

// NewSchema cria um novo schema GraphQL com todos os resolvers configurados
func NewSchema(
	// Queries
	listContextsHandler *queries.ListContextsHandler,
	listAttributesHandler *queries.ListAttributesHandler,
	searchAttributesHandler *queries.SearchAttributesHandler,
	
	// Commands
	createAttributeHandler *commands.CreateAttributeHandler,
	updateAttributeHandler *commands.UpdateAttributeHandler,
	verifyAttributeHandler *commands.VerifyAttributeHandler,
	updateContextVerificationLevelHandler *commands.UpdateContextVerificationLevelHandler,
	updateContextTrustScoreHandler *commands.UpdateContextTrustScoreHandler,
	
	// Services
	contextService *services.ContextService,
	attributeService *services.AttributeService,
	auditLogger services.AuditLogger,
) (*graphql.Schema, error) {
	// Carregar o schema GraphQL
	schemaBytes, err := ioutil.ReadFile(filepath.Join(".", "src", "multi-context", "api", "graphql", "schema.graphql"))
	if err != nil {
		return nil, err
	}
	schemaString := string(schemaBytes)
	
	// Criar os resolvers
	contextResolver := resolvers.NewContextsQueryResolver(
		listContextsHandler,
		contextService,
		auditLogger,
	)
	
	attributeResolver := resolvers.NewAttributesQueryResolver(
		listAttributesHandler,
		searchAttributesHandler,
		attributeService,
		contextService,
		auditLogger,
	)
	
	mutationResolver := resolvers.NewMutationResolver(
		createAttributeHandler,
		updateAttributeHandler,
		verifyAttributeHandler,
		updateContextVerificationLevelHandler,
		updateContextTrustScoreHandler,
		contextService,
		auditLogger,
	)
	
	// Criar o rootResolver que combina todos os resolvers
	rootResolver := &RootResolver{
		ContextsResolver: contextResolver,
		AttributesResolver: attributeResolver,
		MutationResolver: mutationResolver,
	}
	
	// Criar e retornar o schema com os resolvers configurados
	return graphql.ParseSchema(schemaString, rootResolver)
}

// RootResolver combina todos os resolvers para formar o resolver raiz
type RootResolver struct {
	ContextsResolver   *resolvers.ContextsQueryResolver
	AttributesResolver *resolvers.AttributesQueryResolver
	MutationResolver   *resolvers.MutationResolver
}

// Query endpoints

func (r *RootResolver) Context(args struct {
	ID graphql.ID
}) (*resolvers.GraphQLContext, error) {
	return r.ContextsResolver.Context(args)
}

func (r *RootResolver) Contexts(args struct {
	Filters          *resolvers.GraphQLContextFilters
	Pagination       *resolvers.GraphQLPagination
	Sorting          *resolvers.GraphQLSorting
	IncludeAttributes *bool
}) (*resolvers.GraphQLContextsResult, error) {
	return r.ContextsResolver.Contexts(args)
}

func (r *RootResolver) Attribute(args struct {
	ID graphql.ID
}) (*resolvers.GraphQLAttribute, error) {
	return r.AttributesResolver.Attribute(args)
}

func (r *RootResolver) Attributes(args struct {
	Filters    *resolvers.GraphQLAttributeFilters
	Pagination *resolvers.GraphQLPagination
	Sorting    *resolvers.GraphQLSorting
}) (*resolvers.GraphQLAttributesResult, error) {
	return r.AttributesResolver.Attributes(args)
}

func (r *RootResolver) SearchAttributes(args struct {
	Filters    resolvers.GraphQLAttributeSearchFilters
	Pagination *resolvers.GraphQLPagination
	Sorting    *resolvers.GraphQLSorting
}) (*resolvers.GraphQLAttributesResult, error) {
	return r.AttributesResolver.SearchAttributes(args)
}

func (r *RootResolver) AttributeVerificationHistory(args struct {
	AttributeID graphql.ID
}) ([]*resolvers.GraphQLVerificationHistory, error) {
	return r.AttributesResolver.AttributeVerificationHistory(args)
}

func (r *RootResolver) ContextVerificationHistory(args struct {
	ContextID graphql.ID
}) ([]*resolvers.GraphQLVerificationHistory, error) {
	return r.ContextsResolver.ContextVerificationHistory(args)
}

// Mutation endpoints

func (r *RootResolver) CreateAttribute(args struct {
	Input resolvers.GraphQLCreateAttributeInput
}) (*resolvers.GraphQLAttribute, error) {
	return r.MutationResolver.CreateAttribute(args)
}

func (r *RootResolver) UpdateAttribute(args struct {
	Input resolvers.GraphQLUpdateAttributeInput
}) (*resolvers.GraphQLAttribute, error) {
	return r.MutationResolver.UpdateAttribute(args)
}

func (r *RootResolver) VerifyAttribute(args struct {
	Input resolvers.GraphQLVerifyAttributeInput
}) (*resolvers.GraphQLAttribute, error) {
	return r.MutationResolver.VerifyAttribute(args)
}

func (r *RootResolver) UpdateContextVerificationLevel(args struct {
	Input resolvers.GraphQLUpdateContextVerificationLevelInput
}) (*resolvers.GraphQLContext, error) {
	return r.MutationResolver.UpdateContextVerificationLevel(args)
}

func (r *RootResolver) UpdateContextTrustScore(args struct {
	Input resolvers.GraphQLUpdateContextTrustScoreInput
}) (*resolvers.GraphQLContext, error) {
	return r.MutationResolver.UpdateContextTrustScore(args)
}

func (r *RootResolver) DeleteAttribute(args struct {
	AttributeID graphql.ID
	Reason      string
}) (bool, error) {
	return r.MutationResolver.DeleteAttribute(args)
}