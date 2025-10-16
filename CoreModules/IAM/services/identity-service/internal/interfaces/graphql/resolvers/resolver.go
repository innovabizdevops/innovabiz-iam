/**
 * INNOVABIZ IAM - Resolver GraphQL Principal
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação da interface principal do resolver GraphQL para o módulo Core IAM,
 * integrando todos os resolvers específicos de recursos (grupos, usuários, etc.)
 * 
 * Seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (PR.AC-4: Gerenciamento de identidades e credenciais)
 */

package resolvers

import (
	"github.com/innovabiz/iam/internal/domain/entities"
	"github.com/innovabiz/iam/internal/domain/services"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
	"github.com/innovabiz/iam/internal/interfaces/graphql/generated"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
)

// Resolver é a interface principal do resolver GraphQL que conecta os modelos GraphQL
// com os serviços de domínio subjacentes
type Resolver struct {
	Logger       logging.Logger
	Metrics      metrics.MetricsClient
	Tracer       tracing.Tracer
	GroupService services.GroupService
	UserService  services.UserService
}

// Query retorna o resolver para queries GraphQL
func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{
		Resolver: r,
	}
}

// Mutation retorna o resolver para mutations GraphQL
func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{
		Resolver: r,
	}
}

// Group retorna o resolver para o tipo Group
func (r *Resolver) Group() generated.GroupResolver {
	return &groupResolver{
		Resolver: r,
	}
}

// queryResolver implementa a interface QueryResolver gerada
type queryResolver struct {
	*Resolver
}

// mutationResolver implementa a interface MutationResolver gerada
type mutationResolver struct {
	*Resolver
}

// groupResolver implementa a interface GroupResolver gerada
type groupResolver struct {
	*Resolver
}

// Definição de tipos importantes compartilhados para referência em outros arquivos

// Importante: estas interfaces devem corresponder às entidades do domínio
// que são usadas nos resolvers

// As interfaces de entidade são definidas aqui apenas para referência e documentação
// O código real usa os pacotes importados

// Interface compartilhada de GroupService para facilitar mock em testes
// e garantir que todos os resolvers usem os mesmos métodos de serviço
type GroupServiceInterface interface {
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*entities.Group, error)
	GetByCode(ctx context.Context, code string, tenantID uuid.UUID) (*entities.Group, error)
	List(ctx context.Context, tenantID uuid.UUID, page, pageSize int, filter *entities.GroupFilter, sort *entities.SortOption) (*entities.GroupListResult, error)
	Create(ctx context.Context, group *entities.Group, createdBy uuid.UUID) (*entities.Group, error)
	Update(ctx context.Context, group *entities.Group, updatedBy uuid.UUID) (*entities.Group, error)
	ChangeStatus(ctx context.Context, id, tenantID uuid.UUID, status entities.GroupStatus, updatedBy uuid.UUID) (*entities.Group, error)
	Delete(ctx context.Context, id, tenantID uuid.UUID, deletedBy uuid.UUID) error
	AddUserToGroup(ctx context.Context, groupID, userID, tenantID, updatedBy uuid.UUID) (*entities.Group, error)
	RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID, updatedBy uuid.UUID) (*entities.Group, error)
	ListGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, page, pageSize int, filter *entities.UserFilter, sort *entities.SortOption) (*entities.UserListResult, error)
	ListUserGroups(ctx context.Context, userID, tenantID uuid.UUID, includeInheritedGroups bool, page, pageSize int, filter *entities.GroupFilter, sort *entities.SortOption) (*entities.GroupListResult, error)
	GetGroupHierarchy(ctx context.Context, groupID, tenantID uuid.UUID, maxDepth int) ([]*entities.Group, error)
	CheckGroupCircularReference(ctx context.Context, groupID, parentGroupID, tenantID uuid.UUID) (bool, error)
	GetGroupsStatistics(ctx context.Context, tenantID uuid.UUID, groupID uuid.UUID) (*entities.GroupStatistics, error)
}