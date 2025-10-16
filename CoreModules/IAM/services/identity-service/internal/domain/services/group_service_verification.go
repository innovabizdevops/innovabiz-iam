/**
 * INNOVABIZ IAM - Operações de Verificação e Estatísticas do Serviço de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de verificação e estatísticas do serviço de domínio 
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
	"github.com/innovabiz/iam/internal/domain/validation"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// CheckGroupCircularReference verifica se existe referência circular na hierarquia de grupos
func (s *groupService) CheckGroupCircularReference(ctx context.Context, groupID, parentGroupID, tenantID uuid.UUID) (bool, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.CheckGroupCircularReference")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("parent.id", parentGroupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)
	
	timer := s.metrics.Timer("service.group.checkGroupCircularReference.duration")
	defer timer.ObserveDuration()

	s.logger.Debug(ctx, "Verificando referência circular na hierarquia", logging.Fields{
		"groupId":  groupID.String(),
		"parentId": parentGroupID.String(),
		"tenantId": tenantID.String(),
	})

	// Validar parâmetros
	if groupID == uuid.Nil || parentGroupID == uuid.Nil || tenantID == uuid.Nil {
		s.logger.Error(ctx, "Parâmetros inválidos para verificar referência circular", nil)
		s.metrics.Counter("service.group.checkGroupCircularReference.invalidParams").Inc(1)
		return false, validation.ErrInvalidInput
	}

	// Se os IDs são iguais, é uma referência circular direta
	if groupID == parentGroupID {
		s.logger.Warn(ctx, "Referência circular direta detectada", logging.Fields{
			"groupId":  groupID.String(),
			"parentId": parentGroupID.String(),
		})
		s.metrics.Counter("service.group.checkGroupCircularReference.directCircular").Inc(1)
		return true, nil
	}

	// Chamar o repositório para verificar referência circular
	hasCircular, err := s.groupRepo.CheckGroupCircularReference(ctx, groupID, parentGroupID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar referência circular", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"parentId": parentGroupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.checkGroupCircularReference.error").Inc(1)
		return false, fmt.Errorf("erro ao verificar referência circular: %w", err)
	}

	s.metrics.Counter("service.group.checkGroupCircularReference.success").Inc(1)
	if hasCircular {
		s.metrics.Counter("service.group.checkGroupCircularReference.circularDetected").Inc(1)
		s.logger.Warn(ctx, "Referência circular detectada na hierarquia", logging.Fields{
			"groupId":  groupID.String(),
			"parentId": parentGroupID.String(),
			"tenantId": tenantID.String(),
		})
	}

	return hasCircular, nil
}

// GetGroupsStatistics retorna estatísticas agregadas de grupos
func (s *groupService) GetGroupsStatistics(ctx context.Context, tenantID uuid.UUID, groupID *uuid.UUID) (*group.GroupStatistics, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetGroupsStatistics")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
	)
	
	if groupID != nil {
		span.SetAttributes(attribute.String("group.id", groupID.String()))
	}
	
	timer := s.metrics.Timer("service.group.getGroupsStatistics.duration")
	defer timer.ObserveDuration()

	logFields := logging.Fields{
		"tenantId": tenantID.String(),
	}
	
	if groupID != nil {
		logFields["groupId"] = groupID.String()
		s.logger.Debug(ctx, "Obtendo estatísticas para o grupo específico", logFields)
	} else {
		s.logger.Debug(ctx, "Obtendo estatísticas de todos os grupos", logFields)
	}

	// Validar parâmetros
	if tenantID == uuid.Nil {
		s.logger.Error(ctx, "ID do tenant inválido", nil)
		s.metrics.Counter("service.group.getGroupsStatistics.invalidTenant").Inc(1)
		return nil, validation.ErrInvalidInput
	}

	// Se um ID de grupo específico for fornecido, verificar se o grupo existe
	if groupID != nil && *groupID != uuid.Nil {
		_, err := s.groupRepo.GetByID(ctx, *groupID, tenantID)
		if err != nil {
			if errors.Is(err, group.ErrGroupNotFound) {
				s.logger.Error(ctx, "Grupo não encontrado", logging.Fields{
					"groupId":  groupID.String(),
					"tenantId": tenantID.String(),
				})
				s.metrics.Counter("service.group.getGroupsStatistics.groupNotFound").Inc(1)
				return nil, group.ErrGroupNotFound
			}
			s.logger.Error(ctx, "Erro ao verificar existência do grupo", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getGroupsStatistics.error").Inc(1)
			return nil, fmt.Errorf("erro ao verificar grupo: %w", err)
		}
	}

	// Inicializar resultado de estatísticas
	stats := &group.GroupStatistics{
		TenantID:           tenantID,
		GroupID:            groupID,
		TimestampGenerated: time.Now().UTC(),
	}

	// Calcular estatísticas - este processo pode ser complexo e envolver múltiplas consultas
	// ou uma única consulta agregada no banco de dados, dependendo da implementação

	// 1. Total de grupos e distribuição por status
	if groupID == nil {
		// Para estatísticas de todo o tenant
		var err error
		
		// Contar total de grupos
		stats.TotalGroups, err = s.countGroups(ctx, tenantID, nil)
		if err != nil {
			return nil, err
		}
		
		// Contar grupos por status
		statusFilter := group.StatusActive
		stats.ActiveGroups, err = s.countGroups(ctx, tenantID, &statusFilter)
		if err != nil {
			return nil, err
		}
		
		statusFilter = group.StatusInactive
		stats.InactiveGroups, err = s.countGroups(ctx, tenantID, &statusFilter)
		if err != nil {
			return nil, err
		}
		
		statusFilter = group.StatusLocked
		stats.LockedGroups, err = s.countGroups(ctx, tenantID, &statusFilter)
		if err != nil {
			return nil, err
		}
	} else {
		// Para estatísticas de um grupo específico e seus descendentes
		var err error
		
		// Contar subgrupos diretos
		filter := group.GroupFilter{
			ParentGroupID: groupID,
		}
		result, err := s.groupRepo.List(ctx, tenantID, filter, 1, 1)
		if err != nil {
			s.logger.Error(ctx, "Erro ao contar subgrupos diretos", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getGroupsStatistics.childCountError").Inc(1)
			return nil, fmt.Errorf("erro ao contar subgrupos diretos: %w", err)
		}
		stats.DirectChildGroups = result.TotalCount
		
		// Contar todos os subgrupos recursivamente
		stats.TotalChildGroups, err = s.countRecursiveChildren(ctx, *groupID, tenantID)
		if err != nil {
			return nil, err
		}
	}
	
	// 2. Contar usuários
	if groupID != nil {
		// Contar usuários diretos do grupo
		directUsers, err := s.groupRepo.GetGroupUserCount(ctx, *groupID, tenantID, false)
		if err != nil {
			s.logger.Error(ctx, "Erro ao contar usuários diretos do grupo", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getGroupsStatistics.userCountError").Inc(1)
			return nil, fmt.Errorf("erro ao contar usuários diretos: %w", err)
		}
		stats.DirectUsers = directUsers
		
		// Contar todos os usuários do grupo e subgrupos
		totalUsers, err := s.groupRepo.GetGroupUserCount(ctx, *groupID, tenantID, true)
		if err != nil {
			s.logger.Error(ctx, "Erro ao contar usuários totais do grupo", logging.Fields{
				"error":    err.Error(),
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getGroupsStatistics.totalUserCountError").Inc(1)
			return nil, fmt.Errorf("erro ao contar usuários totais: %w", err)
		}
		stats.TotalUsers = totalUsers
	} else {
		// Contar todos os usuários em grupos
		totalUsers, err := s.countUsersInGroups(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		stats.TotalUsers = totalUsers
	}

	// 3. Profundidade da árvore (apenas para estatísticas globais do tenant)
	if groupID == nil {
		maxDepth, err := s.groupRepo.GetMaxGroupDepth(ctx, tenantID)
		if err != nil {
			s.logger.Error(ctx, "Erro ao obter profundidade máxima da hierarquia", logging.Fields{
				"error":    err.Error(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getGroupsStatistics.maxDepthError").Inc(1)
			return nil, fmt.Errorf("erro ao obter profundidade máxima: %w", err)
		}
		stats.MaxHierarchyDepth = maxDepth
	}

	// Calcular as distribuições por nível e tipo se necessário
	// Isso pode envolver consultas adicionais ao banco de dados
	
	// Registrar conclusão bem-sucedida
	s.metrics.Counter("service.group.getGroupsStatistics.success").Inc(1)
	
	if groupID != nil {
		s.logger.Debug(ctx, "Estatísticas do grupo geradas com sucesso", logging.Fields{
			"groupId":      groupID.String(),
			"tenantId":     tenantId.String(),
			"directUsers":  stats.DirectUsers,
			"totalUsers":   stats.TotalUsers,
			"childGroups":  stats.DirectChildGroups,
			"totalGroups":  stats.TotalChildGroups,
		})
	} else {
		s.logger.Debug(ctx, "Estatísticas de grupos do tenant geradas com sucesso", logging.Fields{
			"tenantId":        tenantId.String(),
			"totalGroups":     stats.TotalGroups,
			"activeGroups":    stats.ActiveGroups,
			"inactiveGroups":  stats.InactiveGroups,
			"lockedGroups":    stats.LockedGroups,
			"totalUsers":      stats.TotalUsers,
			"maxDepth":        stats.MaxHierarchyDepth,
		})
	}

	return stats, nil
}

// Funções auxiliares para cálculo de estatísticas

// countGroups conta grupos por status (ou total se status for nil)
func (s *groupService) countGroups(ctx context.Context, tenantID uuid.UUID, status *string) (int, error) {
	filter := group.GroupFilter{}
	if status != nil {
		filter.Status = status
	}
	
	result, err := s.groupRepo.List(ctx, tenantID, filter, 1, 1)
	if err != nil {
		s.logger.Error(ctx, "Erro ao contar grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID.String(),
			"status":   fmt.Sprintf("%v", status),
		})
		s.metrics.Counter("service.group.countGroups.error").Inc(1)
		return 0, fmt.Errorf("erro ao contar grupos: %w", err)
	}
	
	return result.TotalCount, nil
}

// countRecursiveChildren conta todos os subgrupos de forma recursiva
func (s *groupService) countRecursiveChildren(ctx context.Context, groupID, tenantID uuid.UUID) (int, error) {
	// Usando a função GetChildGroups que já suporta recursividade
	result, err := s.groupRepo.GetChildGroups(ctx, groupID, tenantID, true, 1, 1)
	if err != nil {
		s.logger.Error(ctx, "Erro ao contar subgrupos recursivamente", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.countRecursiveChildren.error").Inc(1)
		return 0, fmt.Errorf("erro ao contar subgrupos recursivamente: %w", err)
	}
	
	return result.TotalCount, nil
}

// countUsersInGroups conta o total de usuários únicos em grupos
func (s *groupService) countUsersInGroups(ctx context.Context, tenantID uuid.UUID) (int, error) {
	// Esta operação precisa de uma função específica no repositório
	count, err := s.groupRepo.CountUniqueUsersInGroups(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao contar usuários em grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.countUsersInGroups.error").Inc(1)
		return 0, fmt.Errorf("erro ao contar usuários em grupos: %w", err)
	}
	
	return count, nil
}