/**
 * INNOVABIZ IAM - Interfaces GraphQL Geradas
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Definição das interfaces GraphQL geradas para o módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 */

package generated

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
)

// Este arquivo define as interfaces que devem ser implementadas pelos resolvers.
// Normalmente, essas interfaces seriam geradas automaticamente pelo gqlgen,
// mas aqui estamos definindo-as manualmente para referência e integração.

// ResolverRoot define os resolvers raiz para a API GraphQL
type ResolverRoot interface {
	Query() QueryResolver
	Mutation() MutationResolver
	Group() GroupResolver
}

// QueryResolver define os métodos para todas as queries GraphQL
type QueryResolver interface {
	// Queries para Grupos
	GetGroupByID(ctx context.Context, id string, tenantID string) (*model.Group, error)
	GetGroupByCode(ctx context.Context, code string, tenantID string) (*model.Group, error)
	ListGroups(ctx context.Context, tenantID string, page int, pageSize int, filter *model.GroupFilter, sortField *string, sortDirection *model.SortDirection) (*model.GroupListResult, error)
	
	// Queries para Hierarquia e Membros de Grupos
	GetGroupHierarchy(ctx context.Context, groupID string, tenantID string, maxDepth *int) ([]*model.Group, error)
	ListGroupMembers(ctx context.Context, groupID string, tenantID string, page int, pageSize int, filter *model.UserFilter, sortField *string, sortDirection *model.SortDirection) (*model.UserListResult, error)
	ListUserGroups(ctx context.Context, userID string, tenantID string, includeInheritedGroups *bool, page int, pageSize int, filter *model.GroupFilter, sortField *string, sortDirection *model.SortDirection) (*model.GroupListResult, error)
	
	// Queries para Estatísticas e Verificações de Grupos
	GetGroupStatistics(ctx context.Context, tenantID string, groupID *string) (*model.GroupStatistics, error)
	CheckGroupCircularReference(ctx context.Context, groupID string, parentGroupID string, tenantID string) (bool, error)
}

// MutationResolver define os métodos para todas as mutations GraphQL
type MutationResolver interface {
	// Mutations para CRUD de Grupos
	CreateGroup(ctx context.Context, input model.CreateGroupInput) (*model.Group, error)
	UpdateGroup(ctx context.Context, input model.UpdateGroupInput) (*model.Group, error)
	ChangeGroupStatus(ctx context.Context, input model.ChangeGroupStatusInput) (*model.Group, error)
	DeleteGroup(ctx context.Context, id string, tenantID string) (bool, error)
	
	// Mutations para Gerenciamento de Membros de Grupos
	AddUserToGroup(ctx context.Context, input model.AddUserToGroupInput) (*model.Group, error)
	RemoveUserFromGroup(ctx context.Context, input model.RemoveUserFromGroupInput) (*model.Group, error)
}

// GroupResolver define os métodos para resolver campos relacionados no tipo Group
type GroupResolver interface {
	ParentGroup(ctx context.Context, obj *model.Group) (*model.Group, error)
	UserCount(ctx context.Context, obj *model.Group) (int, error)
	ChildGroupsCount(ctx context.Context, obj *model.Group) (int, error)
	Path(ctx context.Context, obj *model.Group) (string, error)
	Level(ctx context.Context, obj *model.Group) (int, error)
}

// Config configura os resolvers para a API GraphQL
type Config struct {
	Resolvers  ResolverRoot
	Directives DirectiveRoot
}

// DirectiveRoot define as diretivas GraphQL disponíveis
type DirectiveRoot struct {
	Auth         func(ctx context.Context, obj interface{}, next graphql.Resolver, roles []string) (interface{}, error)
	ValidateInput func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error)
}

// NewExecutableSchema cria um novo esquema GraphQL executável
func NewExecutableSchema(config Config) graphql.ExecutableSchema {
	// Em uma implementação real, este método seria gerado pelo gqlgen
	// Aqui, estamos apenas definindo a interface para referência
	// Retorno placeholder
	return &executableSchema{config: config}
}

// executableSchema é uma implementação interna para o esquema executável
type executableSchema struct {
	config Config
}

// Schema retorna o esquema GraphQL como string
func (e *executableSchema) Schema() string {
	// Placeholder para o esquema GraphQL
	return ""
}

// Complexity retorna a complexidade do esquema GraphQL
func (e *executableSchema) Complexity(typeName, fieldName string, childComplexity int, args map[string]interface{}) (int, bool) {
	// Placeholder para cálculo de complexidade
	return 1, false
}

// Exec executa uma operação GraphQL
func (e *executableSchema) Exec(ctx context.Context) graphql.ResponseHandler {
	// Placeholder para execução de operações GraphQL
	return func(ctx context.Context) *graphql.Response {
		return &graphql.Response{}
	}
}