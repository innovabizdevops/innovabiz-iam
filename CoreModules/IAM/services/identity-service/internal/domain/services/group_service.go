/**
 * INNOVABIZ IAM - Serviço de Domínio para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Interface e implementação do serviço de domínio para gestão de grupos
 * no módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package services

import (
	"context"

	"github.com/google/uuid"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/infrastructure/events"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
)

// GroupService define a interface do serviço de domínio para grupos
type GroupService interface {
	// Operações CRUD básicas
	GetByID(ctx context.Context, groupID, tenantID uuid.UUID) (*group.Group, error)
	GetByCode(ctx context.Context, code string, tenantID uuid.UUID) (*group.Group, error)
	Create(ctx context.Context, group *group.Group) error
	Update(ctx context.Context, group *group.Group) error
	ChangeStatus(ctx context.Context, groupID, tenantID uuid.UUID, status string, updatedBy *uuid.UUID) error
	Delete(ctx context.Context, groupID, tenantID uuid.UUID, deletedBy *uuid.UUID) error

	// Operações de listagem e consulta
	List(ctx context.Context, tenantID uuid.UUID, filter group.GroupFilter, page, pageSize int) (*group.GroupListResult, error)
	FindGroupsByUserID(ctx context.Context, userID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*group.GroupListResult, error)
	
	// Operações de hierarquia
	GetParentGroup(ctx context.Context, groupID, tenantID uuid.UUID) (*group.Group, error)
	GetChildGroups(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*group.GroupListResult, error)
	
	// Operações de membros
	AddUserToGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID, addedBy *uuid.UUID) error
	RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID, removedBy *uuid.UUID) error
	IsUserInGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) (bool, error)
	ListGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, filter user.UserFilter, page, pageSize int) (*user.UserListResult, error)
	
	// Operações de estatísticas
	GetGroupsStatistics(ctx context.Context, tenantID uuid.UUID, groupID *uuid.UUID) (*group.GroupStatistics, error)
	GetGroupUserCount(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool) (int, error)
	
	// Operações de verificação
	CheckGroupCircularReference(ctx context.Context, groupID, parentGroupID, tenantID uuid.UUID) (bool, error)
}

// groupService implementa a interface GroupService
type groupService struct {
	groupRepo      group.Repository
	userRepo       user.Repository
	eventPublisher events.Publisher
	logger         logging.Logger
	metrics        metrics.MetricsClient
	tracer         tracing.Tracer
}

// NewGroupService cria uma nova instância do serviço de grupos
func NewGroupService(
	groupRepo group.Repository,
	userRepo user.Repository,
	eventPublisher events.Publisher,
	logger logging.Logger,
	metrics metrics.MetricsClient,
	tracer tracing.Tracer,
) GroupService {
	return &groupService{
		groupRepo:      groupRepo,
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
		metrics:        metrics,
		tracer:         tracer,
	}
}