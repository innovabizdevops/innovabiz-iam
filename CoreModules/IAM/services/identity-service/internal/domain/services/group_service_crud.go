/**
 * INNOVABIZ IAM - Operações CRUD do Serviço de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações CRUD do serviço de domínio para grupos
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
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/domain/validation"
	"github.com/innovabiz/iam/internal/infrastructure/events"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// GetByID recupera um grupo pelo seu ID
func (s *groupService) GetByID(ctx context.Context, groupID, tenantID uuid.UUID) (*group.Group, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := s.metrics.Timer("service.group.getById.duration")
	defer timer.ObserveDuration()

	s.logger.Debug(ctx, "Buscando grupo por ID", logging.Fields{
		"groupId":  groupID.String(),
		"tenantId": tenantID.String(),
	})

	// Chamar o repositório para buscar o grupo
	g, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Debug(ctx, "Grupo não encontrado", logging.Fields{
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getById.notFound").Inc(1)
			return nil, group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao buscar grupo por ID", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.getById.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo: %w", err)
	}

	s.metrics.Counter("service.group.getById.success").Inc(1)
	return g, nil
}

// GetByCode recupera um grupo pelo seu código
func (s *groupService) GetByCode(ctx context.Context, code string, tenantID uuid.UUID) (*group.Group, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetByCode")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.code", code),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := s.metrics.Timer("service.group.getByCode.duration")
	defer timer.ObserveDuration()

	s.logger.Debug(ctx, "Buscando grupo por código", logging.Fields{
		"groupCode": code,
		"tenantId":  tenantID.String(),
	})

	// Validar código
	if code == "" {
		s.logger.Error(ctx, "Código do grupo não pode ser vazio", nil)
		s.metrics.Counter("service.group.getByCode.invalidCode").Inc(1)
		return nil, validation.ErrInvalidInput
	}

	// Chamar o repositório para buscar o grupo
	g, err := s.groupRepo.GetByCode(ctx, code, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Debug(ctx, "Grupo não encontrado pelo código", logging.Fields{
				"groupCode": code,
				"tenantId":  tenantID.String(),
			})
			s.metrics.Counter("service.group.getByCode.notFound").Inc(1)
			return nil, group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao buscar grupo por código", logging.Fields{
			"error":     err.Error(),
			"groupCode": code,
			"tenantId":  tenantID.String(),
		})
		s.metrics.Counter("service.group.getByCode.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo por código: %w", err)
	}

	s.metrics.Counter("service.group.getByCode.success").Inc(1)
	return g, nil
}

// Create cria um novo grupo
func (s *groupService) Create(ctx context.Context, g *group.Group) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.code", g.Code),
		attribute.String("group.name", g.Name),
		attribute.String("tenant.id", g.TenantID.String()),
	)

	timer := s.metrics.Timer("service.group.create.duration")
	defer timer.ObserveDuration()

	s.logger.Info(ctx, "Criando novo grupo", logging.Fields{
		"groupCode": g.Code,
		"groupName": g.Name,
		"tenantId":  g.TenantID.String(),
	})

	// Validar dados do grupo
	if err := s.validateGroupData(ctx, g); err != nil {
		s.metrics.Counter("service.group.create.validationError").Inc(1)
		return err
	}

	// Se não for fornecido um ID, gerar um novo
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}

	// Definir timestamps de criação
	g.CreatedAt = time.Now().UTC()

	// Chamar o repositório para criar o grupo
	if err := s.groupRepo.Create(ctx, g); err != nil {
		s.logger.Error(ctx, "Erro ao criar grupo", logging.Fields{
			"error":     err.Error(),
			"groupCode": g.Code,
			"groupName": g.Name,
			"tenantId":  g.TenantID.String(),
		})
		s.metrics.Counter("service.group.create.error").Inc(1)
		return fmt.Errorf("erro ao criar grupo: %w", err)
	}

	// Publicar evento de grupo criado
	event := events.GroupCreatedEvent{
		GroupID:   g.ID,
		TenantID:  g.TenantID,
		Code:      g.Code,
		Name:      g.Name,
		Status:    g.Status,
		CreatedBy: g.CreatedBy,
		Timestamp: time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishGroupCreated(ctx, event); err != nil {
		s.logger.Error(ctx, "Erro ao publicar evento de grupo criado", logging.Fields{
			"error":   err.Error(),
			"groupId": g.ID.String(),
		})
		// Não retornar erro aqui, pois o grupo já foi criado com sucesso
		s.metrics.Counter("service.group.create.eventError").Inc(1)
	}

	s.logger.Info(ctx, "Grupo criado com sucesso", logging.Fields{
		"groupId":  g.ID.String(),
		"tenantId": g.TenantID.String(),
	})

	s.metrics.Counter("service.group.create.success").Inc(1)
	return nil
}

// Update atualiza um grupo existente
func (s *groupService) Update(ctx context.Context, g *group.Group) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.Update")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", g.ID.String()),
		attribute.String("group.code", g.Code),
		attribute.String("tenant.id", g.TenantID.String()),
	)

	timer := s.metrics.Timer("service.group.update.duration")
	defer timer.ObserveDuration()

	s.logger.Info(ctx, "Atualizando grupo", logging.Fields{
		"groupId":   g.ID.String(),
		"groupCode": g.Code,
		"tenantId":  g.TenantID.String(),
	})

	// Verificar se o grupo existe
	existingGroup, err := s.groupRepo.GetByID(ctx, g.ID, g.TenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.logger.Error(ctx, "Grupo não encontrado para atualização", logging.Fields{
				"groupId":  g.ID.String(),
				"tenantId": g.TenantID.String(),
			})
			s.metrics.Counter("service.group.update.notFound").Inc(1)
			return group.ErrGroupNotFound
		}
		s.logger.Error(ctx, "Erro ao buscar grupo para atualização", logging.Fields{
			"error":    err.Error(),
			"groupId":  g.ID.String(),
			"tenantId": g.TenantID.String(),
		})
		s.metrics.Counter("service.group.update.error").Inc(1)
		return fmt.Errorf("erro ao buscar grupo para atualização: %w", err)
	}

	// Validar dados do grupo
	if err := s.validateGroupData(ctx, g); err != nil {
		s.metrics.Counter("service.group.update.validationError").Inc(1)
		return err
	}

	// Preservar dados imutáveis
	g.CreatedAt = existingGroup.CreatedAt
	g.CreatedBy = existingGroup.CreatedBy

	// Chamar o repositório para atualizar o grupo
	if err := s.groupRepo.Update(ctx, g); err != nil {
		s.logger.Error(ctx, "Erro ao atualizar grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  g.ID.String(),
			"tenantId": g.TenantID.String(),
		})
		s.metrics.Counter("service.group.update.error").Inc(1)
		return fmt.Errorf("erro ao atualizar grupo: %w", err)
	}

	// Publicar evento de grupo atualizado
	event := events.GroupUpdatedEvent{
		GroupID:   g.ID,
		TenantID:  g.TenantID,
		Code:      g.Code,
		Name:      g.Name,
		Status:    g.Status,
		UpdatedBy: g.UpdatedBy,
		Timestamp: time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishGroupUpdated(ctx, event); err != nil {
		s.logger.Error(ctx, "Erro ao publicar evento de grupo atualizado", logging.Fields{
			"error":   err.Error(),
			"groupId": g.ID.String(),
		})
		// Não retornar erro aqui, pois o grupo já foi atualizado com sucesso
		s.metrics.Counter("service.group.update.eventError").Inc(1)
	}

	s.logger.Info(ctx, "Grupo atualizado com sucesso", logging.Fields{
		"groupId":  g.ID.String(),
		"tenantId": g.TenantID.String(),
	})

	s.metrics.Counter("service.group.update.success").Inc(1)
	return nil
}

// ChangeStatus altera o status de um grupo
func (s *groupService) ChangeStatus(ctx context.Context, groupID, tenantID uuid.UUID, status string, updatedBy *uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.ChangeStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.String("status", status),
	)

	timer := s.metrics.Timer("service.group.changeStatus.duration")
	defer timer.ObserveDuration()

	s.logger.Info(ctx, "Alterando status do grupo", logging.Fields{
		"groupId":  groupID.String(),
		"tenantId": tenantID.String(),
		"status":   status,
	})

	// Validar status
	validStatus := map[string]bool{
		group.StatusActive:   true,
		group.StatusInactive: true,
		group.StatusLocked:   true,
	}

	if !validStatus[status] {
		s.logger.Error(ctx, "Status inválido", logging.Fields{
			"status":   status,
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.changeStatus.invalidStatus").Inc(1)
		return group.ErrInvalidStatus
	}

	// Verificar se o grupo existe
	existingGroup, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.metrics.Counter("service.group.changeStatus.notFound").Inc(1)
			return group.ErrGroupNotFound
		}
		s.metrics.Counter("service.group.changeStatus.error").Inc(1)
		return err
	}

	// Se o status já é o mesmo, retornar sem erro
	if existingGroup.Status == status {
		s.logger.Debug(ctx, "Grupo já está com o status solicitado", logging.Fields{
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
			"status":   status,
		})
		s.metrics.Counter("service.group.changeStatus.unchanged").Inc(1)
		return nil
	}

	// Chamar o repositório para alterar o status
	if err := s.groupRepo.ChangeStatus(ctx, groupID, tenantID, status, updatedBy); err != nil {
		s.logger.Error(ctx, "Erro ao alterar status do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
			"status":   status,
		})
		s.metrics.Counter("service.group.changeStatus.error").Inc(1)
		return fmt.Errorf("erro ao alterar status: %w", err)
	}

	// Publicar evento de status alterado
	event := events.GroupStatusChangedEvent{
		GroupID:   groupID,
		TenantID:  tenantID,
		Status:    status,
		UpdatedBy: updatedBy,
		Timestamp: time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishGroupStatusChanged(ctx, event); err != nil {
		s.logger.Error(ctx, "Erro ao publicar evento de status alterado", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
			"status":  status,
		})
		// Não retornar erro aqui, pois o status já foi alterado com sucesso
		s.metrics.Counter("service.group.changeStatus.eventError").Inc(1)
	}

	s.logger.Info(ctx, "Status do grupo alterado com sucesso", logging.Fields{
		"groupId":  groupID.String(),
		"tenantId": tenantID.String(),
		"status":   status,
	})

	s.metrics.Counter("service.group.changeStatus.success").Inc(1)
	return nil
}

// Delete realiza a exclusão lógica de um grupo
func (s *groupService) Delete(ctx context.Context, groupID, tenantID uuid.UUID, deletedBy *uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "GroupService.Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)

	timer := s.metrics.Timer("service.group.delete.duration")
	defer timer.ObserveDuration()

	s.logger.Info(ctx, "Excluindo grupo", logging.Fields{
		"groupId":  groupID.String(),
		"tenantId": tenantID.String(),
	})

	// Verificar se o grupo existe
	_, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			s.metrics.Counter("service.group.delete.notFound").Inc(1)
			return group.ErrGroupNotFound
		}
		s.metrics.Counter("service.group.delete.error").Inc(1)
		return err
	}

	// Iniciar transação para operações de exclusão
	tx, err := s.groupRepo.BeginTx(ctx)
	if err != nil {
		s.logger.Error(ctx, "Erro ao iniciar transação para exclusão do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		s.metrics.Counter("service.group.delete.transactionError").Inc(1)
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}

	// Criar contexto com a transação
	txCtx := tx.WithContext(ctx)

	// Executar a exclusão lógica
	if err := s.groupRepo.SoftDelete(txCtx, groupID, tenantID, deletedBy); err != nil {
		s.logger.Error(ctx, "Erro ao excluir grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})

		// Rollback da transação em caso de erro
		if rbErr := s.groupRepo.RollbackTx(txCtx); rbErr != nil {
			s.logger.Error(ctx, "Erro ao fazer rollback da transação", logging.Fields{
				"error": rbErr.Error(),
			})
		}

		s.metrics.Counter("service.group.delete.error").Inc(1)
		return fmt.Errorf("erro ao excluir grupo: %w", err)
	}

	// Commit da transação
	if err := s.groupRepo.CommitTx(txCtx); err != nil {
		s.logger.Error(ctx, "Erro ao fazer commit da transação de exclusão", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		s.metrics.Counter("service.group.delete.commitError").Inc(1)
		return fmt.Errorf("erro ao fazer commit da transação: %w", err)
	}

	// Publicar evento de grupo excluído
	event := events.GroupDeletedEvent{
		GroupID:   groupID,
		TenantID:  tenantID,
		DeletedBy: deletedBy,
		Timestamp: time.Now().UTC(),
	}

	if err := s.eventPublisher.PublishGroupDeleted(ctx, event); err != nil {
		s.logger.Error(ctx, "Erro ao publicar evento de grupo excluído", logging.Fields{
			"error":   err.Error(),
			"groupId": groupID.String(),
		})
		// Não retornar erro aqui, pois o grupo já foi excluído com sucesso
		s.metrics.Counter("service.group.delete.eventError").Inc(1)
	}

	s.logger.Info(ctx, "Grupo excluído com sucesso", logging.Fields{
		"groupId":  groupID.String(),
		"tenantId": tenantID.String(),
	})

	s.metrics.Counter("service.group.delete.success").Inc(1)
	return nil
}

// validateGroupData valida os dados do grupo antes de criação/atualização
func (s *groupService) validateGroupData(ctx context.Context, g *group.Group) error {
	// Validar campos obrigatórios
	if g.Code == "" {
		s.logger.Error(ctx, "Código do grupo é obrigatório", nil)
		return validation.NewValidationError("O código do grupo é obrigatório")
	}

	if g.Name == "" {
		s.logger.Error(ctx, "Nome do grupo é obrigatório", nil)
		return validation.NewValidationError("O nome do grupo é obrigatório")
	}

	if g.TenantID == uuid.Nil {
		s.logger.Error(ctx, "ID do tenant é obrigatório", nil)
		return validation.NewValidationError("O ID do tenant é obrigatório")
	}

	// Validar formato do código (apenas caracteres alfanuméricos e hífen)
	if !validation.IsValidGroupCode(g.Code) {
		s.logger.Error(ctx, "Código do grupo contém caracteres inválidos", logging.Fields{
			"code": g.Code,
		})
		return validation.NewValidationError("O código do grupo deve conter apenas letras, números e hífen")
	}

	// Validar status
	if g.Status == "" {
		// Se não for fornecido, usar o padrão
		g.Status = group.StatusActive
	} else {
		// Se for fornecido, validar
		validStatus := map[string]bool{
			group.StatusActive:   true,
			group.StatusInactive: true,
			group.StatusLocked:   true,
		}

		if !validStatus[g.Status] {
			s.logger.Error(ctx, "Status do grupo inválido", logging.Fields{
				"status": g.Status,
			})
			return validation.NewValidationError("Status inválido. Deve ser ACTIVE, INACTIVE ou LOCKED")
		}
	}

	// Se tiver grupo pai, validar
	if g.ParentGroupID != nil {
		// Verificar se o pai existe
		_, err := s.groupRepo.GetByID(ctx, *g.ParentGroupID, g.TenantID)
		if err != nil {
			if errors.Is(err, group.ErrGroupNotFound) {
				s.logger.Error(ctx, "Grupo pai não encontrado", logging.Fields{
					"parentId": g.ParentGroupID.String(),
				})
				return validation.NewValidationError("Grupo pai não encontrado")
			}
			s.logger.Error(ctx, "Erro ao validar grupo pai", logging.Fields{
				"error":    err.Error(),
				"parentId": g.ParentGroupID.String(),
			})
			return fmt.Errorf("erro ao validar grupo pai: %w", err)
		}

		// Verificar referência circular se o grupo já tiver ID
		if g.ID != uuid.Nil {
			hasCircular, err := s.groupRepo.CheckGroupCircularReference(ctx, g.ID, *g.ParentGroupID, g.TenantID)
			if err != nil {
				s.logger.Error(ctx, "Erro ao verificar referência circular", logging.Fields{
					"error":    err.Error(),
					"groupId":  g.ID.String(),
					"parentId": g.ParentGroupID.String(),
				})
				return fmt.Errorf("erro ao verificar referência circular: %w", err)
			}

			if hasCircular {
				s.logger.Error(ctx, "Detectada referência circular na hierarquia", logging.Fields{
					"groupId":  g.ID.String(),
					"parentId": g.ParentGroupID.String(),
				})
				return group.ErrGroupCircularReference
			}
		}
	}

	return nil
}