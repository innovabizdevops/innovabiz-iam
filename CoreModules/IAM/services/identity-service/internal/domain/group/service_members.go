/**
 * INNOVABIZ IAM - Gestão de Membros de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de gestão de membros em grupos,
 * incluindo adição, remoção e consulta de usuários em grupos.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Seção 5.13: Controle de Acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR (Minimização de dados e controle de acesso)
 * - BNA Instrução 7/2021 (Requisitos de segurança para instituições financeiras)
 */

package group

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
)

// PaginatedUsers representa um resultado paginado de usuários
type PaginatedUsers struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	TotalPages int    `json:"totalPages"`
}

// User representa um usuário simplificado para o contexto do serviço de grupos
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	DisplayName  string    `json:"displayName"`
	Status       string    `json:"status"`
	TenantID     uuid.UUID `json:"tenantId"`
}

// AddUserToGroup adiciona um usuário a um grupo
func (s *ServiceImpl) AddUserToGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.AddUserToGroup")
	defer span.End()
	
	s.logger.Debug(ctx, "Adicionando usuário ao grupo", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})
	
	// Verificar autorização
	if err := s.authz.CheckPermission(ctx, "iam:groups:members:add", map[string]interface{}{
		"tenantId": tenantId.String(),
		"groupId":  groupID.String(),
		"userId":   userID.String(),
	}); err != nil {
		s.logger.Warn(ctx, "Acesso não autorizado para adicionar usuário ao grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
			"error":    err.Error(),
		})
		return ErrUnauthorizedOperation
	}
	
	// Iniciar transação
	tx, err := s.db.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Erro ao iniciar transação", logging.Fields{
			"error": err.Error(),
		})
		return err
	}
	defer tx.Rollback(ctx)
	
	// Verificar se o grupo existe
	group, err := s.GetGroupByID(ctx, groupID, tenantID)
	if err != nil {
		return err
	}
	
	// Verificar se o grupo está ativo
	if group.Status != StatusActive {
		s.logger.Warn(ctx, "Tentativa de adicionar usuário a grupo inativo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
			"status":   string(group.Status),
		})
		return fmt.Errorf("grupo não está ativo: %s", group.Status)
	}
	
	// Verificar se o usuário existe
	user, err := s.userService.GetUserByID(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Usuário não encontrado", logging.Fields{
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
			"error":    err.Error(),
		})
		return ErrUserNotFound
	}
	
	// Verificar se o usuário já pertence ao grupo
	existingMember, err := s.db.FindOne(ctx, "group_members", map[string]interface{}{
		"groupId":  groupID,
		"userId":   userID,
		"tenantId": tenantID,
	})
	
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar membro existente", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return err
	}
	
	if existingMember != nil {
		s.logger.Warn(ctx, "Usuário já é membro do grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return ErrUserAlreadyInGroup
	}
	
	// Criar registro de associação
	now := time.Now().UTC()
	membership := map[string]interface{}{
		"id":        uuid.New(),
		"groupId":   groupID,
		"userId":    userID,
		"tenantId":  tenantID,
		"createdAt": now,
		"addedBy":   s.authz.GetCurrentUserID(ctx),
	}
	
	// Inserir no banco de dados
	if err := tx.Insert(ctx, "group_members", membership); err != nil {
		s.logger.Error(ctx, "Erro ao adicionar usuário ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return err
	}
	
	// Publicar evento
	event := events.Event{
		Type:      "iam.group.member.added",
		TenantID:  tenantID.String(),
		Subject:   userID.String(),
		Object:    groupID.String(),
		Timestamp: now,
		Actor:     s.authz.GetCurrentUserID(ctx).String(),
		Data: map[string]interface{}{
			"groupName":      group.Name,
			"groupCode":      group.Code,
			"userEmail":      user.Email,
			"userFirstName":  user.FirstName,
			"userLastName":   user.LastName,
		},
		Metadata: map[string]string{
			"regionCode":    group.RegionCode,
			"groupType":     group.GroupType,
			"userStatus":    string(user.Status),
		},
	}
	
	if err := s.eventPublisher.Publish(ctx, "iam.events", event); err != nil {
		s.logger.Warn(ctx, "Erro ao publicar evento de adição de membro", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		// Não falha a operação se o evento não for publicado
	}
	
	// Registrar métricas
	s.tracer.RecordMetric(ctx, MetricNamespaceGroup, "group_member_added", 1, map[string]string{
		"tenantId": tenantID.String(),
		"groupId":  groupID.String(),
	})
	
	// Confirmar a transação
	if err := tx.Commit(ctx); err != nil {
		s.logger.Error(ctx, "Erro ao confirmar transação", logging.Fields{
			"error": err.Error(),
		})
		return err
	}
	
	s.logger.Info(ctx, "Usuário adicionado ao grupo com sucesso", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})
	
	return nil
}

// RemoveUserFromGroup remove um usuário de um grupo
func (s *ServiceImpl) RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.RemoveUserFromGroup")
	defer span.End()
	
	s.logger.Debug(ctx, "Removendo usuário do grupo", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})
	
	// Verificar autorização
	if err := s.authz.CheckPermission(ctx, "iam:groups:members:remove", map[string]interface{}{
		"tenantId": tenantId.String(),
		"groupId":  groupID.String(),
		"userId":   userID.String(),
	}); err != nil {
		s.logger.Warn(ctx, "Acesso não autorizado para remover usuário do grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
			"error":    err.Error(),
		})
		return ErrUnauthorizedOperation
	}
	
	// Verificar se o grupo existe
	group, err := s.GetGroupByID(ctx, groupID, tenantID)
	if err != nil {
		return err
	}
	
	// Verificar se o usuário existe
	user, err := s.userService.GetUserByID(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Usuário não encontrado", logging.Fields{
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
			"error":    err.Error(),
		})
		return ErrUserNotFound
	}
	
	// Verificar se o usuário pertence ao grupo
	existingMember, err := s.db.FindOne(ctx, "group_members", map[string]interface{}{
		"groupId":  groupID,
		"userId":   userID,
		"tenantId": tenantID,
	})
	
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar membro existente", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return err
	}
	
	if existingMember == nil {
		s.logger.Warn(ctx, "Usuário não é membro do grupo", logging.Fields{
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return ErrUserNotInGroup
	}
	
	// Iniciar transação
	tx, err := s.db.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Erro ao iniciar transação", logging.Fields{
			"error": err.Error(),
		})
		return err
	}
	defer tx.Rollback(ctx)
	
	// Remover do banco de dados
	if err := tx.Delete(ctx, "group_members", map[string]interface{}{
		"groupId":  groupID,
		"userId":   userID,
		"tenantId": tenantID,
	}); err != nil {
		s.logger.Error(ctx, "Erro ao remover usuário do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		return err
	}
	
	// Publicar evento
	now := time.Now().UTC()
	event := events.Event{
		Type:      "iam.group.member.removed",
		TenantID:  tenantID.String(),
		Subject:   userID.String(),
		Object:    groupID.String(),
		Timestamp: now,
		Actor:     s.authz.GetCurrentUserID(ctx).String(),
		Data: map[string]interface{}{
			"groupName":      group.Name,
			"groupCode":      group.Code,
			"userEmail":      user.Email,
			"userFirstName":  user.FirstName,
			"userLastName":   user.LastName,
		},
		Metadata: map[string]string{
			"regionCode":    group.RegionCode,
			"groupType":     group.GroupType,
		},
	}
	
	if err := s.eventPublisher.Publish(ctx, "iam.events", event); err != nil {
		s.logger.Warn(ctx, "Erro ao publicar evento de remoção de membro", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})
		// Não falha a operação se o evento não for publicado
	}
	
	// Registrar métricas
	s.tracer.RecordMetric(ctx, MetricNamespaceGroup, "group_member_removed", 1, map[string]string{
		"tenantId": tenantID.String(),
		"groupId":  groupID.String(),
	})
	
	// Confirmar a transação
	if err := tx.Commit(ctx); err != nil {
		s.logger.Error(ctx, "Erro ao confirmar transação", logging.Fields{
			"error": err.Error(),
		})
		return err
	}
	
	s.logger.Info(ctx, "Usuário removido do grupo com sucesso", logging.Fields{
		"groupId":  groupID.String(),
		"userId":   userID.String(),
		"tenantId": tenantID.String(),
	})
	
	return nil
}