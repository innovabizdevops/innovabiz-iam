/**
 * INNOVABIZ IAM - Operações de Membros do Serviço de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de gestão de membros do serviço de domínio 
 * para grupos no módulo Core IAM, seguindo a arquitetura multi-dimensional, 
 * multi-tenant e com observabilidade total da plataforma INNOVABIZ.
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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/domain/validation"
	"github.com/innovabiz/iam/internal/infrastructure/events"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// AddUserToGroup adiciona um usuário a um grupo
func (s *groupService) AddUserToGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID, addedBy *uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.AddUserToGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)
	
	if addedBy != nil {
		span.SetAttributes(attribute.String("added_by", addedBy.String()))
	}
	
	timer := s.metrics.Timer("service.group.addUserToGroup.duration")
	defer timer.ObserveDuration()

	s.logger.Info(ctx, "Adicionando usuário ao grupo", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})

	// Validar parâmetros
	if groupID == uuid.Nil || userID == uuid.Nil || tenantID == uuid.Nil {
		s.logger.Error(ctx, "Parâmetros inválidos para adicionar usuário ao grupo", nil)
		s.metrics.Counter("service.group.addUserToGroup.invalidParams").Inc(1)
		return validation.ErrInvalidInput
	}

	// Verificar se o grupo existe
	g, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Error(ctx, "Grupo não encontrado", logging.Fields{
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.addUserToGroup.groupNotFound").Inc(1)
			return group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao verificar existência do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.error").Inc(1)
		return fmt.Errorf("erro ao verificar grupo: %w", err)
	}

	// Verificar se o grupo está ativo
	if g.Status != group.StatusActive {
		s.logger.Error(ctx, "Grupo não está ativo", logging.Fields{
			"groupId":  groupID.String(),
			"status":   g.Status,
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.groupNotActive").Inc(1)
		return group.ErrGroupNotActive
	}

	// Verificar se o usuário existe
	userExists, err := s.userRepo.Exists(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar existência do usuário", logging.Fields{
			"error":    err.Error(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.userCheckError").Inc(1)
		return fmt.Errorf("erro ao verificar usuário: %w", err)
	}
	
	if !userExists {
		s.logger.Error(ctx, "Usuário não encontrado", logging.Fields{
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.userNotFound").Inc(1)
		return user.ErrUserNotFound
	}

	// Verificar se o usuário já está no grupo (idempotência)
	isMember, err := s.groupRepo.IsUserInGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar pertencimento ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.memberCheckError").Inc(1)
		return fmt.Errorf("erro ao verificar pertencimento ao grupo: %w", err)
	}
	
	if isMember {
		s.logger.Info(ctx, "Usuário já é membro do grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.alreadyMember").Inc(1)
		return nil // Retornar sucesso por idempotência
	}

	// Adicionar o usuário ao grupo
	if err := s.groupRepo.AddUserToGroup(ctx, groupID, userID, tenantID, addedBy); err != nil {
		s.logger.Error(ctx, "Erro ao adicionar usuário ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.addUserToGroup.error").Inc(1)
		return fmt.Errorf("erro ao adicionar usuário ao grupo: %w", err)
	}

	// Publicar evento de usuário adicionado ao grupo
	event := events.UserAddedToGroupEvent{
		GroupID:   groupID,
		UserID:    userID,
		TenantID:  tenantID,
		AddedBy:   addedBy,
		Timestamp: time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishUserAddedToGroup(ctx, event); err != nil {
		s.logger.Error(ctx, "Erro ao publicar evento de usuário adicionado ao grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
			"userId":  userID.String(),
		})
		// Não retornar erro aqui, pois o usuário já foi adicionado com sucesso
		s.metrics.Counter("service.group.addUserToGroup.eventError").Inc(1)
	}

	s.logger.Info(ctx, "Usuário adicionado ao grupo com sucesso", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})

	s.metrics.Counter("service.group.addUserToGroup.success").Inc(1)
	return nil
}

// RemoveUserFromGroup remove um usuário de um grupo
func (s *groupService) RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID, removedBy *uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.RemoveUserFromGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)
	
	if removedBy != nil {
		span.SetAttributes(attribute.String("removed_by", removedBy.String()))
	}
	
	timer := s.metrics.Timer("service.group.removeUserFromGroup.duration")
	defer timer.ObserveDuration()

	s.logger.Info(ctx, "Removendo usuário do grupo", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})

	// Validar parâmetros
	if groupID == uuid.Nil || userID == uuid.Nil || tenantID == uuid.Nil {
		s.logger.Error(ctx, "Parâmetros inválidos para remover usuário do grupo", nil)
		s.metrics.Counter("service.group.removeUserFromGroup.invalidParams").Inc(1)
		return validation.ErrInvalidInput
	}

	// Verificar se o grupo existe
	_, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Error(ctx, "Grupo não encontrado", logging.Fields{
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.removeUserFromGroup.groupNotFound").Inc(1)
			return group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao verificar existência do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.removeUserFromGroup.error").Inc(1)
		return fmt.Errorf("erro ao verificar grupo: %w", err)
	}

	// Verificar se o usuário existe
	userExists, err := s.userRepo.Exists(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar existência do usuário", logging.Fields{
			"error":    err.Error(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.removeUserFromGroup.userCheckError").Inc(1)
		return fmt.Errorf("erro ao verificar usuário: %w", err)
	}
	
	if !userExists {
		s.logger.Error(ctx, "Usuário não encontrado", logging.Fields{
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.removeUserFromGroup.userNotFound").Inc(1)
		return user.ErrUserNotFound
	}

	// Verificar se o usuário está no grupo
	isMember, err := s.groupRepo.IsUserInGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar pertencimento ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.removeUserFromGroup.memberCheckError").Inc(1)
		return fmt.Errorf("erro ao verificar pertencimento ao grupo: %w", err)
	}
	
	if !isMember {
		s.logger.Info(ctx, "Usuário não é membro do grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.removeUserFromGroup.notMember").Inc(1)
		return nil // Retornar sucesso por idempotência
	}

	// Remover o usuário do grupo
	if err := s.groupRepo.RemoveUserFromGroup(ctx, groupID, userID, tenantID, removedBy); err != nil {
		s.logger.Error(ctx, "Erro ao remover usuário do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.removeUserFromGroup.error").Inc(1)
		return fmt.Errorf("erro ao remover usuário do grupo: %w", err)
	}

	// Publicar evento de usuário removido do grupo
	event := events.UserRemovedFromGroupEvent{
		GroupID:   groupID,
		UserID:    userID,
		TenantID:  tenantID,
		RemovedBy: removedBy,
		Timestamp: time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishUserRemovedFromGroup(ctx, event); err != nil {
		s.logger.Error(ctx, "Erro ao publicar evento de usuário removido do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
			"userId":  userID.String(),
		})
		// Não retornar erro aqui, pois o usuário já foi removido com sucesso
		s.metrics.Counter("service.group.removeUserFromGroup.eventError").Inc(1)
	}

	s.logger.Info(ctx, "Usuário removido do grupo com sucesso", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})

	s.metrics.Counter("service.group.removeUserFromGroup.success").Inc(1)
	return nil
}

// IsUserInGroup verifica se um usuário é membro de um grupo
func (s *groupService) IsUserInGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.IsUserInGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)
	
	timer := s.metrics.Timer("service.group.isUserInGroup.duration")
	defer timer.ObserveDuration()

	s.logger.Debug(ctx, "Verificando se usuário é membro do grupo", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})

	// Validar parâmetros
	if groupID == uuid.Nil || userID == uuid.Nil || tenantID == uuid.Nil {
		s.logger.Error(ctx, "Parâmetros inválidos para verificar pertencimento ao grupo", nil)
		s.metrics.Counter("service.group.isUserInGroup.invalidParams").Inc(1)
		return false, validation.ErrInvalidInput
	}

	// Chamar o repositório para verificar o pertencimento
	isMember, err := s.groupRepo.IsUserInGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar pertencimento ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.isUserInGroup.error").Inc(1)
		return false, fmt.Errorf("erro ao verificar pertencimento ao grupo: %w", err)
	}

	s.metrics.Counter("service.group.isUserInGroup.success").Inc(1)
	if isMember {
		s.metrics.Counter("service.group.isUserInGroup.isMember").Inc(1)
	} else {
		s.metrics.Counter("service.group.isUserInGroup.isNotMember").Inc(1)
	}

	return isMember, nil
}

// ListGroupMembers lista os membros de um grupo
func (s *groupService) ListGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, filter user.UserFilter, page, pageSize int) (*user.UserListResult, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.ListGroupMembers")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)
	
	timer := s.metrics.Timer("service.group.listGroupMembers.duration")
	defer timer.ObserveDuration()

	// Validar parâmetros
	if groupID == uuid.Nil || tenantID == uuid.Nil {
		s.logger.Error(ctx, "Parâmetros inválidos para listar membros do grupo", nil)
		s.metrics.Counter("service.group.listGroupMembers.invalidParams").Inc(1)
		return nil, validation.ErrInvalidInput
	}

	// Validar parâmetros de paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Tamanho padrão da página
	}

	s.logger.Debug(ctx, "Listando membros do grupo", logging.Fields{
		"groupId":   groupID.String(),
		"tenantId":  tenantID.String(),
		"recursive": recursive,
		"page":      page,
		"pageSize":  pageSize,
	})

	// Verificar se o grupo existe
	_, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Error(ctx, "Grupo não encontrado", logging.Fields{
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.listGroupMembers.groupNotFound").Inc(1)
			return nil, group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao verificar existência do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.listGroupMembers.error").Inc(1)
		return nil, fmt.Errorf("erro ao verificar grupo: %w", err)
	}

	// Chamar o repositório para listar os membros
	result, err := s.groupRepo.ListGroupMembers(ctx, groupID, tenantID, recursive, filter, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "Erro ao listar membros do grupo", logging.Fields{
			"error":     err.Error(),
			"groupId":   groupID.String(),
			"tenantId":  tenantID.String(),
			"recursive": recursive,
		})
		s.metrics.Counter("service.group.listGroupMembers.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar membros do grupo: %w", err)
	}

	// Registrar métricas de resultado
	s.metrics.Counter("service.group.listGroupMembers.success").Inc(1)
	s.metrics.Gauge("service.group.listGroupMembers.totalItems").Set(float64(result.TotalCount))
	s.metrics.Gauge("service.group.listGroupMembers.resultCount").Set(float64(len(result.Users)))

	s.logger.Debug(ctx, "Membros do grupo listados com sucesso", logging.Fields{
		"groupId":    groupID.String(),
		"tenantId":   tenantID.String(),
		"recursive":  recursive,
		"totalCount": result.TotalCount,
		"itemCount":  len(result.Users),
	})

	return result, nil
}

// GetGroupUserCount retorna o número de usuários em um grupo
func (s *groupService) GetGroupUserCount(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool) (int, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetGroupUserCount")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
	)
	
	timer := s.metrics.Timer("service.group.getGroupUserCount.duration")
	defer timer.ObserveDuration()

	s.logger.Debug(ctx, "Contando usuários do grupo", logging.Fields{
		"groupId":   groupID.String(),
		"tenantId":  tenantID.String(),
		"recursive": recursive,
	})

	// Validar parâmetros
	if groupID == uuid.Nil || tenantID == uuid.Nil {
		s.logger.Error(ctx, "Parâmetros inválidos para contar usuários do grupo", nil)
		s.metrics.Counter("service.group.getGroupUserCount.invalidParams").Inc(1)
		return 0, validation.ErrInvalidInput
	}

	// Verificar se o grupo existe
	_, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Error(ctx, "Grupo não encontrado", logging.Fields{
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getGroupUserCount.groupNotFound").Inc(1)
			return 0, group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao verificar existência do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.getGroupUserCount.error").Inc(1)
		return 0, fmt.Errorf("erro ao verificar grupo: %w", err)
	}

	// Chamar o repositório para contar os membros
	count, err := s.groupRepo.GetGroupUserCount(ctx, groupID, tenantID, recursive)
	if err != nil {
		s.logger.Error(ctx, "Erro ao contar usuários do grupo", logging.Fields{
			"error":     err.Error(),
			"groupId":   groupID.String(),
			"tenantId":  tenantID.String(),
			"recursive": recursive,
		})
		s.metrics.Counter("service.group.getGroupUserCount.error").Inc(1)
		return 0, fmt.Errorf("erro ao contar usuários do grupo: %w", err)
	}

	s.metrics.Counter("service.group.getGroupUserCount.success").Inc(1)
	s.metrics.Gauge("service.group.getGroupUserCount.count").Set(float64(count))

	s.logger.Debug(ctx, "Usuários do grupo contados com sucesso", logging.Fields{
		"groupId":   groupID.String(),
		"tenantId":  tenantID.String(),
		"recursive": recursive,
		"count":     count,
	})

	return count, nil
}