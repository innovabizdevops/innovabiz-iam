/**
 * INNOVABIZ IAM - Operações de Listagem e Consulta do Serviço de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das operações de listagem e consulta do serviço de domínio 
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
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/domain/validation"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// List retorna uma lista paginada de grupos com base em filtros
func (s *groupService) List(ctx context.Context, tenantID uuid.UUID, filter group.GroupFilter, page, pageSize int) (*group.GroupListResult, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant.id", tenantID.String()),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)
	
	if filter.SearchTerm != nil {
		span.SetAttributes(attribute.String("filter.searchTerm", *filter.SearchTerm))
	}
	if filter.Status != nil {
		span.SetAttributes(attribute.String("filter.status", *filter.Status))
	}
	
	timer := s.metrics.Timer("service.group.list.duration")
	defer timer.ObserveDuration()

	// Validar parâmetros de paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Tamanho padrão da página
	}

	// Aplicar validação aos filtros
	if err := s.validateListFilter(ctx, &filter); err != nil {
		s.metrics.Counter("service.group.list.validationError").Inc(1)
		return nil, err
	}

	s.logger.Debug(ctx, "Listando grupos", logging.Fields{
		"tenantId": tenantID.String(),
		"page":     page,
		"pageSize": pageSize,
		"filter":   fmt.Sprintf("%+v", filter),
	})

	// Chamar o repositório para listar os grupos
	result, err := s.groupRepo.List(ctx, tenantID, filter, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "Erro ao listar grupos", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID.String(),
			"page":     page,
			"pageSize": pageSize,
		})
		s.metrics.Counter("service.group.list.error").Inc(1)
		return nil, fmt.Errorf("erro ao listar grupos: %w", err)
	}

	// Registrar métricas de resultado
	s.metrics.Counter("service.group.list.success").Inc(1)
	s.metrics.Gauge("service.group.list.totalItems").Set(float64(result.TotalCount))
	s.metrics.Gauge("service.group.list.resultCount").Set(float64(len(result.Groups)))

	s.logger.Debug(ctx, "Grupos listados com sucesso", logging.Fields{
		"tenantId":   tenantID.String(),
		"page":       page,
		"pageSize":   pageSize,
		"totalCount": result.TotalCount,
		"itemCount":  len(result.Groups),
	})

	return result, nil
}

// FindGroupsByUserID encontra grupos dos quais um usuário é membro
func (s *groupService) FindGroupsByUserID(ctx context.Context, userID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*group.GroupListResult, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.FindGroupsByUserID")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)
	
	timer := s.metrics.Timer("service.group.findGroupsByUserId.duration")
	defer timer.ObserveDuration()

	// Validar parâmetros
	if userID == uuid.Nil {
		s.logger.Error(ctx, "ID de usuário inválido", nil)
		s.metrics.Counter("service.group.findGroupsByUserId.invalidUserId").Inc(1)
		return nil, validation.ErrInvalidInput
	}

	// Validar parâmetros de paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Tamanho padrão da página
	}

	s.logger.Debug(ctx, "Buscando grupos do usuário", logging.Fields{
		"userId":    userID.String(),
		"tenantId":  tenantID.String(),
		"recursive": recursive,
		"page":      page,
		"pageSize":  pageSize,
	})

	// Verificar se o usuário existe
	// Esta verificação pode ser opcional, dependendo dos requisitos
	userExists, err := s.userRepo.Exists(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error(ctx, "Erro ao verificar existência do usuário", logging.Fields{
			"error":   err.Error(),
			"userId":  userID.String(),
			"tenant":  tenantID.String(),
		})
		s.metrics.Counter("service.group.findGroupsByUserId.userCheckError").Inc(1)
		return nil, fmt.Errorf("erro ao verificar existência do usuário: %w", err)
	}
	
	if !userExists {
		s.logger.Error(ctx, "Usuário não encontrado", logging.Fields{
			"userId":  userID.String(),
			"tenant":  tenantID.String(),
		})
		s.metrics.Counter("service.group.findGroupsByUserId.userNotFound").Inc(1)
		return nil, user.ErrUserNotFound
	}

	// Chamar o repositório para encontrar os grupos do usuário
	result, err := s.groupRepo.FindGroupsByUserID(ctx, userID, tenantID, recursive, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "Erro ao buscar grupos do usuário", logging.Fields{
			"error":     err.Error(),
			"userId":    userID.String(),
			"tenantId":  tenantID.String(),
			"recursive": recursive,
		})
		s.metrics.Counter("service.group.findGroupsByUserId.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupos do usuário: %w", err)
	}

	// Registrar métricas de resultado
	s.metrics.Counter("service.group.findGroupsByUserId.success").Inc(1)
	s.metrics.Gauge("service.group.findGroupsByUserId.totalItems").Set(float64(result.TotalCount))
	s.metrics.Gauge("service.group.findGroupsByUserId.resultCount").Set(float64(len(result.Groups)))

	s.logger.Debug(ctx, "Grupos do usuário encontrados com sucesso", logging.Fields{
		"userId":     userID.String(),
		"tenantId":   tenantID.String(),
		"recursive":  recursive,
		"totalCount": result.TotalCount,
		"itemCount":  len(result.Groups),
	})

	return result, nil
}

// GetParentGroup retorna o grupo pai de um grupo especificado
func (s *groupService) GetParentGroup(ctx context.Context, groupID, tenantID uuid.UUID) (*group.Group, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetParentGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
	)
	
	timer := s.metrics.Timer("service.group.getParentGroup.duration")
	defer timer.ObserveDuration()

	// Validar parâmetros
	if groupID == uuid.Nil {
		s.logger.Error(ctx, "ID de grupo inválido", nil)
		s.metrics.Counter("service.group.getParentGroup.invalidGroupId").Inc(1)
		return nil, validation.ErrInvalidInput
	}

	s.logger.Debug(ctx, "Buscando grupo pai", logging.Fields{
		"groupId":  groupID.String(),
		"tenantId": tenantID.String(),
	})

	// Chamar o repositório para buscar o grupo pai
	parent, err := s.groupRepo.GetParentGroup(ctx, groupID, tenantID)
	if err != nil {
		if err == group.ErrGroupNotFound {
			s.logger.Debug(ctx, "Grupo não possui pai ou não foi encontrado", logging.Fields{
				"groupId":  groupID.String(),
				"tenantId": tenantID.String(),
			})
			s.metrics.Counter("service.group.getParentGroup.notFound").Inc(1)
			return nil, group.ErrGroupNotFound
		}
		
		s.logger.Error(ctx, "Erro ao buscar grupo pai", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.getParentGroup.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo pai: %w", err)
	}

	s.metrics.Counter("service.group.getParentGroup.success").Inc(1)
	return parent, nil
}

// GetChildGroups retorna os grupos filhos de um grupo especificado
func (s *groupService) GetChildGroups(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, page, pageSize int) (*group.GroupListResult, error) {
	ctx, span := s.tracer.Start(ctx, "GroupService.GetChildGroups")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", groupID.String()),
		attribute.String("tenant.id", tenantID.String()),
		attribute.Bool("recursive", recursive),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
	)
	
	timer := s.metrics.Timer("service.group.getChildGroups.duration")
	defer timer.ObserveDuration()

	// Validar parâmetros
	if groupID == uuid.Nil {
		s.logger.Error(ctx, "ID de grupo inválido", nil)
		s.metrics.Counter("service.group.getChildGroups.invalidGroupId").Inc(1)
		return nil, validation.ErrInvalidInput
	}

	// Validar parâmetros de paginação
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Tamanho padrão da página
	}

	s.logger.Debug(ctx, "Buscando grupos filhos", logging.Fields{
		"groupId":   groupID.String(),
		"tenantId":  tenantID.String(),
		"recursive": recursive,
		"page":      page,
		"pageSize":  pageSize,
	})

	// Verificar se o grupo pai existe
	_, err := s.groupRepo.GetByID(ctx, groupID, tenantID)
	if err != nil {
		if err == group.ErrGroupNotFound {
			s.metrics.Counter("service.group.getChildGroups.parentNotFound").Inc(1)
			return nil, group.ErrGroupNotFound
		}
		s.metrics.Counter("service.group.getChildGroups.error").Inc(1)
		return nil, err
	}

	// Construir filtro para grupos filhos
	filter := group.GroupFilter{
		ParentGroupID: &groupID,
	}

	// Chamar o repositório para listar os grupos filhos
	result, err := s.groupRepo.List(ctx, tenantID, filter, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "Erro ao buscar grupos filhos", logging.Fields{
			"error":     err.Error(),
			"groupId":   groupID.String(),
			"tenantId":  tenantID.String(),
			"recursive": recursive,
		})
		s.metrics.Counter("service.group.getChildGroups.error").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupos filhos: %w", err)
	}

	// Se não for recursivo, retornar apenas os filhos diretos
	if !recursive {
		s.metrics.Counter("service.group.getChildGroups.success").Inc(1)
		s.metrics.Gauge("service.group.getChildGroups.directCount").Set(float64(len(result.Groups)))
		return result, nil
	}

	// Para implementação recursiva, podemos precisar de várias consultas ao banco
	// dependendo da estrutura do banco de dados. Para bancos que suportam consultas
	// recursivas (como PostgreSQL), seria melhor implementar isso no próprio repositório.

	// Este é um exemplo simples para ilustrar a recursividade:
	// Nota: Em uma implementação real, isso seria feito de forma mais eficiente
	// no nível do banco de dados usando CTE (Common Table Expressions) recursivas
	// ou outras técnicas de consulta hierárquica.

	// Aqui estamos usando a implementação do repositório que deve suportar recursividade
	result, err = s.groupRepo.GetChildGroups(ctx, groupID, tenantID, true, page, pageSize)
	if err != nil {
		s.logger.Error(ctx, "Erro ao buscar grupos filhos recursivamente", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"tenantId": tenantID.String(),
		})
		s.metrics.Counter("service.group.getChildGroups.recursiveError").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupos filhos recursivamente: %w", err)
	}

	s.metrics.Counter("service.group.getChildGroups.success").Inc(1)
	s.metrics.Gauge("service.group.getChildGroups.recursiveCount").Set(float64(len(result.Groups)))
	
	s.logger.Debug(ctx, "Grupos filhos encontrados com sucesso", logging.Fields{
		"groupId":    groupID.String(),
		"tenantId":   tenantID.String(),
		"recursive":  recursive,
		"totalCount": result.TotalCount,
		"itemCount":  len(result.Groups),
	})

	return result, nil
}

// validateListFilter valida os filtros de listagem
func (s *groupService) validateListFilter(ctx context.Context, filter *group.GroupFilter) error {
	// Validar status se fornecido
	if filter.Status != nil && *filter.Status != "" {
		validStatus := map[string]bool{
			group.StatusActive:   true,
			group.StatusInactive: true,
			group.StatusLocked:   true,
			group.StatusDeleted:  true,
		}

		if !validStatus[*filter.Status] {
			s.logger.Error(ctx, "Status de filtro inválido", logging.Fields{
				"status": *filter.Status,
			})
			return validation.NewValidationError("Status de filtro inválido")
		}
	}

	// Validar datas se fornecidas
	if filter.CreatedAfter != nil && filter.CreatedBefore != nil {
		if filter.CreatedAfter.After(*filter.CreatedBefore) {
			s.logger.Error(ctx, "Data inicial não pode ser posterior à data final", nil)
			return validation.NewValidationError("Data inicial não pode ser posterior à data final")
		}
	}

	// Validar nível hierárquico
	if filter.Level != nil && (*filter.Level < 1 || *filter.Level > 10) {
		s.logger.Error(ctx, "Nível hierárquico inválido", logging.Fields{
			"level": *filter.Level,
		})
		return validation.NewValidationError("Nível hierárquico deve ser entre 1 e 10")
	}

	// Validar grupo pai se fornecido
	if filter.ParentGroupID != nil && *filter.ParentGroupID == uuid.Nil {
		s.logger.Error(ctx, "ID do grupo pai inválido", nil)
		return validation.NewValidationError("ID do grupo pai inválido")
	}

	// Validar usuário membro se fornecido
	if filter.UserID != nil && *filter.UserID == uuid.Nil {
		s.logger.Error(ctx, "ID de usuário membro inválido", nil)
		return validation.NewValidationError("ID de usuário membro inválido")
	}

	// Validar ordenação
	if filter.SortBy != nil {
		validSortFields := map[string]bool{
			"name":       true,
			"code":       true,
			"status":     true,
			"created_at": true,
			"updated_at": true,
			"level":      true,
		}

		if !validSortFields[*filter.SortBy] {
			s.logger.Error(ctx, "Campo de ordenação inválido", logging.Fields{
				"sortBy": *filter.SortBy,
			})
			return validation.NewValidationError("Campo de ordenação inválido")
		}
	}

	if filter.SortDirection != nil {
		*filter.SortDirection = strings.ToUpper(*filter.SortDirection)
		if *filter.SortDirection != "ASC" && *filter.SortDirection != "DESC" {
			s.logger.Error(ctx, "Direção de ordenação inválida", logging.Fields{
				"sortDirection": *filter.SortDirection,
			})
			return validation.NewValidationError("Direção de ordenação deve ser ASC ou DESC")
		}
	}

	return nil
}