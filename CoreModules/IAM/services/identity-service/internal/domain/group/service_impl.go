/**
 * INNOVABIZ IAM - Implementação do Serviço de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações CRUD e gestão de grupos para o módulo IAM,
 * com suporte multi-tenant, multi-dimensional e observabilidade total.
 */

package group

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/tenant"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/infrastructure/database"
	"github.com/innovabiz/iam/internal/infrastructure/events"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"github.com/innovabiz/iam/internal/infrastructure/security"
)

// Service define a interface para o serviço de grupos
type Service interface {
	// Operações CRUD básicas
	GetGroupByID(ctx context.Context, id, tenantID uuid.UUID) (*Group, error)
	GetGroupByCode(ctx context.Context, code string, tenantID uuid.UUID) (*Group, error)
	ListGroups(ctx context.Context, tenantID uuid.UUID, filter GroupFilter) (*PaginatedGroups, error)
	CreateGroup(ctx context.Context, input CreateGroupInput) (*Group, error)
	UpdateGroup(ctx context.Context, input UpdateGroupInput) (*Group, error)
	ChangeGroupStatus(ctx context.Context, input ChangeGroupStatusInput) (*Group, error)
	
	// Operações de gestão de membros
	AddUserToGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error
	RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error
	ListGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*PaginatedUsers, error)
	
	// Operações de hierarquia
	GetGroupHierarchy(ctx context.Context, id, tenantID uuid.UUID) ([]Group, error)
	GetSubgroups(ctx context.Context, id, tenantID uuid.UUID, recursive bool) ([]Group, error)
	
	// Estatísticas
	GetGroupStats(ctx context.Context, id, tenantID uuid.UUID) (*GroupStats, error)
}

// ServiceImpl implementa a interface Service
type ServiceImpl struct {
	db             database.Repository
	userService    user.Service
	tenantService  tenant.Service
	eventPublisher events.Publisher
	logger         logging.Logger
	tracer         observability.Tracer
	validator      validation.Validator
	authz          security.Authorizer
}

// NewService cria uma nova instância do serviço de grupos
func NewService(
	db database.Repository,
	userService user.Service,
	tenantService tenant.Service,
	eventPublisher events.Publisher,
	logger logging.Logger,
	tracer observability.Tracer,
	validator validation.Validator,
	authz security.Authorizer,
) Service {
	return &ServiceImpl{
		db:             db,
		userService:    userService,
		tenantService:  tenantService,
		eventPublisher: eventPublisher,
		logger:         logger,
		tracer:         tracer,
		validator:      validator,
		authz:          authz,
	}
}

// GetGroupByID obtém um grupo pelo seu ID
func (s *ServiceImpl) GetGroupByID(ctx context.Context, id, tenantID uuid.UUID) (*Group, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetGroupByID")
	defer span.End()
	
	s.logger.Debug(ctx, "Buscando grupo por ID", logging.Fields{
		"groupId":  id.String(),
		"tenantId": tenantID.String(),
	})
	
	// Verificar autorização
	if err := s.authz.CheckPermission(ctx, "iam:groups:read", map[string]interface{}{
		"tenantId": tenantID.String(),
		"groupId":  id.String(),
	}); err != nil {
		s.logger.Warn(ctx, "Acesso não autorizado ao grupo", logging.Fields{
			"groupId":  id.String(),
			"tenantId": tenantID.String(),
			"error":    err.Error(),
		})
		return nil, ErrUnauthorizedOperation
	}
	
	// Buscar o grupo no banco de dados
	group, err := s.db.FindOne(ctx, "groups", map[string]interface{}{
		"id":       id,
		"tenantId": tenantID,
		"status":   map[string]interface{}{"$ne": StatusDeleted},
	})
	
	if err != nil {
		s.logger.Error(ctx, "Erro ao buscar grupo por ID", logging.Fields{
			"error":    err.Error(),
			"groupId":  id.String(),
			"tenantId": tenantID.String(),
		})
		return nil, err
	}
	
	if group == nil {
		return nil, ErrGroupNotFound
	}
	
	// Mapear resultado para entidade de domínio
	result := &Group{}
	if err := s.db.MapToEntity(group, result); err != nil {
		s.logger.Error(ctx, "Erro ao mapear grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  id.String(),
		})
		return nil, err
	}
	
	// Registrar métrica de acesso
	s.tracer.RecordMetric(ctx, MetricNamespaceGroup, "group_access_count", 1, map[string]string{
		"tenantId": tenantID.String(),
		"groupId":  id.String(),
	})
	
	return result, nil
}

// GetGroupByCode obtém um grupo pelo seu código único
func (s *ServiceImpl) GetGroupByCode(ctx context.Context, code string, tenantID uuid.UUID) (*Group, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetGroupByCode")
	defer span.End()
	
	s.logger.Debug(ctx, "Buscando grupo por código", logging.Fields{
		"code":     code,
		"tenantId": tenantID.String(),
	})
	
	// Verificar autorização
	if err := s.authz.CheckPermission(ctx, "iam:groups:read", map[string]interface{}{
		"tenantId": tenantID.String(),
		"code":     code,
	}); err != nil {
		s.logger.Warn(ctx, "Acesso não autorizado ao grupo", logging.Fields{
			"code":     code,
			"tenantId": tenantID.String(),
			"error":    err.Error(),
		})
		return nil, ErrUnauthorizedOperation
	}
	
	// Buscar o grupo no banco de dados
	group, err := s.db.FindOne(ctx, "groups", map[string]interface{}{
		"code":     code,
		"tenantId": tenantID,
		"status":   map[string]interface{}{"$ne": StatusDeleted},
	})
	
	if err != nil {
		s.logger.Error(ctx, "Erro ao buscar grupo por código", logging.Fields{
			"error":    err.Error(),
			"code":     code,
			"tenantId": tenantID.String(),
		})
		return nil, err
	}
	
	if group == nil {
		return nil, ErrGroupNotFound
	}
	
	// Mapear resultado para entidade de domínio
	result := &Group{}
	if err := s.db.MapToEntity(group, result); err != nil {
		s.logger.Error(ctx, "Erro ao mapear grupo", logging.Fields{
			"error": err.Error(),
			"code":  code,
		})
		return nil, err
	}
	
	return result, nil
}