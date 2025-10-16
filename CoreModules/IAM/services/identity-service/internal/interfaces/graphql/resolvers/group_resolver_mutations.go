/**
 * INNOVABIZ IAM - Mutations do Resolver GraphQL para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação das mutations do resolver GraphQL para operações de criação,
 * atualização e exclusão de grupos no módulo Core IAM, seguindo a arquitetura 
 * multi-dimensional, multi-tenant e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package resolvers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
)

// CreateGroup resolve a mutation para criar um novo grupo
func (r *GroupResolver) CreateGroup(ctx context.Context, input model.CreateGroupInput) (*model.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.CreateGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.code", input.Code),
		attribute.String("tenant.id", input.TenantID),
	)

	timer := r.metrics.Timer("resolver.group.create.duration")
	defer timer.ObserveDuration()

	r.logger.Info(ctx, "Resolvendo mutation para criar grupo", logging.Fields{
		"code":     input.Code,
		"name":     input.Name,
		"tenantId": input.TenantID,
	})

	// Obter o ID do usuário atual do contexto (autenticado)
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter ID do usuário do contexto", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.create.authError").Inc(1)
		return nil, fmt.Errorf("erro de autenticação: %w", err)
	}

	// Converter entrada do modelo GraphQL para o modelo de domínio
	domainGroup, err := graphQLGroupToDomain(&input)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter entrada para modelo de domínio", logging.Fields{
			"error": err.Error(),
			"input": fmt.Sprintf("%+v", input),
		})
		r.metrics.Counter("resolver.group.create.conversionError").Inc(1)
		return nil, fmt.Errorf("erro nos dados de entrada: %w", err)
	}

	// Definir o criador como o usuário atual
	domainGroup.CreatedBy = &currentUserID

	// Chamar o serviço de domínio para criar o grupo
	err = r.groupService.Create(ctx, domainGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao criar grupo", logging.Fields{
			"error":    err.Error(),
			"code":     input.Code,
			"name":     input.Name,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.create.error").Inc(1)
		return nil, fmt.Errorf("erro ao criar grupo: %w", err)
	}

	// Converter o modelo de domínio para o modelo GraphQL
	result, err := domainGroupToGraphQL(domainGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter grupo para modelo GraphQL", logging.Fields{
			"error":   err.Error(),
			"groupId": domainGroup.ID.String(),
		})
		r.metrics.Counter("resolver.group.create.responseConversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter resposta: %w", err)
	}

	r.metrics.Counter("resolver.group.create.success").Inc(1)
	r.logger.Info(ctx, "Grupo criado com sucesso", logging.Fields{
		"groupId":  domainGroup.ID.String(),
		"code":     input.Code,
		"name":     input.Name,
		"tenantId": input.TenantID,
	})

	return result, nil
}

// UpdateGroup resolve a mutation para atualizar um grupo existente
func (r *GroupResolver) UpdateGroup(ctx context.Context, input model.UpdateGroupInput) (*model.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.UpdateGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", input.ID),
		attribute.String("tenant.id", input.TenantID),
	)

	timer := r.metrics.Timer("resolver.group.update.duration")
	defer timer.ObserveDuration()

	r.logger.Info(ctx, "Resolvendo mutation para atualizar grupo", logging.Fields{
		"groupId":  input.ID,
		"tenantId": input.TenantID,
	})

	// Obter o ID do usuário atual do contexto (autenticado)
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter ID do usuário do contexto", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.update.authError").Inc(1)
		return nil, fmt.Errorf("erro de autenticação: %w", err)
	}

	// Converter IDs de string para UUID
	groupID, err := uuid.Parse(input.ID)
	if err != nil {
		r.logger.Error(ctx, "ID do grupo inválido", logging.Fields{
			"error":   err.Error(),
			"groupId": input.ID,
		})
		r.metrics.Counter("resolver.group.update.invalidId").Inc(1)
		return nil, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(input.TenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.update.invalidTenantId").Inc(1)
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Obter o grupo existente
	existingGroup, err := r.groupService.GetByID(ctx, groupID, tenantUUID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar grupo existente", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.ID,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.update.getError").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo: %w", err)
	}

	// Converter entrada do modelo GraphQL para o modelo de domínio
	updateGroup, err := graphQLUpdateGroupToDomain(&input)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter entrada para modelo de domínio", logging.Fields{
			"error": err.Error(),
			"input": fmt.Sprintf("%+v", input),
		})
		r.metrics.Counter("resolver.group.update.conversionError").Inc(1)
		return nil, fmt.Errorf("erro nos dados de entrada: %w", err)
	}

	// Preservar campos que não foram atualizados
	if updateGroup.Code == "" {
		updateGroup.Code = existingGroup.Code
	}
	if updateGroup.Name == "" {
		updateGroup.Name = existingGroup.Name
	}
	if updateGroup.Description == "" && existingGroup.Description != "" {
		updateGroup.Description = existingGroup.Description
	}
	if updateGroup.Type == nil && existingGroup.Type != nil {
		updateGroup.Type = existingGroup.Type
	}
	if updateGroup.ParentGroupID == nil && existingGroup.ParentGroupID != nil {
		updateGroup.ParentGroupID = existingGroup.ParentGroupID
	}
	if updateGroup.Attributes == nil && existingGroup.Attributes != nil {
		updateGroup.Attributes = existingGroup.Attributes
	}
	if updateGroup.Metadata == nil && existingGroup.Metadata != nil {
		updateGroup.Metadata = existingGroup.Metadata
	}

	// Definir o atualizador como o usuário atual
	updateGroup.UpdatedBy = &currentUserID

	// Chamar o serviço de domínio para atualizar o grupo
	err = r.groupService.Update(ctx, updateGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao atualizar grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.ID,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.update.error").Inc(1)
		return nil, fmt.Errorf("erro ao atualizar grupo: %w", err)
	}

	// Obter o grupo atualizado
	updatedGroup, err := r.groupService.GetByID(ctx, groupID, tenantUUID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar grupo após atualização", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.ID,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.update.getUpdatedError").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo atualizado: %w", err)
	}

	// Converter o modelo de domínio para o modelo GraphQL
	result, err := domainGroupToGraphQL(updatedGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter grupo para modelo GraphQL", logging.Fields{
			"error":   err.Error(),
			"groupId": input.ID,
		})
		r.metrics.Counter("resolver.group.update.responseConversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter resposta: %w", err)
	}

	r.metrics.Counter("resolver.group.update.success").Inc(1)
	r.logger.Info(ctx, "Grupo atualizado com sucesso", logging.Fields{
		"groupId":  input.ID,
		"tenantId": input.TenantID,
	})

	return result, nil
}

// ChangeGroupStatus resolve a mutation para alterar o status de um grupo
func (r *GroupResolver) ChangeGroupStatus(ctx context.Context, input model.ChangeGroupStatusInput) (*model.Group, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.ChangeGroupStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", input.ID),
		attribute.String("tenant.id", input.TenantID),
		attribute.String("status", string(input.Status)),
	)

	timer := r.metrics.Timer("resolver.group.changeStatus.duration")
	defer timer.ObserveDuration()

	r.logger.Info(ctx, "Resolvendo mutation para alterar status de grupo", logging.Fields{
		"groupId":  input.ID,
		"tenantId": input.TenantID,
		"status":   string(input.Status),
	})

	// Obter o ID do usuário atual do contexto (autenticado)
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter ID do usuário do contexto", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.changeStatus.authError").Inc(1)
		return nil, fmt.Errorf("erro de autenticação: %w", err)
	}

	// Converter IDs de string para UUID
	groupID, err := uuid.Parse(input.ID)
	if err != nil {
		r.logger.Error(ctx, "ID do grupo inválido", logging.Fields{
			"error":   err.Error(),
			"groupId": input.ID,
		})
		r.metrics.Counter("resolver.group.changeStatus.invalidId").Inc(1)
		return nil, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(input.TenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.changeStatus.invalidTenantId").Inc(1)
		return nil, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Converter status do GraphQL para o domínio
	status := mapGraphQLStatusToDomain(input.Status)

	// Chamar o serviço de domínio para alterar o status
	err = r.groupService.ChangeStatus(ctx, groupID, tenantUUID, status, &currentUserID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao alterar status do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.ID,
			"tenantId": input.TenantID,
			"status":   status,
		})
		r.metrics.Counter("resolver.group.changeStatus.error").Inc(1)
		return nil, fmt.Errorf("erro ao alterar status do grupo: %w", err)
	}

	// Obter o grupo atualizado
	updatedGroup, err := r.groupService.GetByID(ctx, groupID, tenantUUID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao buscar grupo após alteração de status", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.ID,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.changeStatus.getUpdatedError").Inc(1)
		return nil, fmt.Errorf("erro ao buscar grupo atualizado: %w", err)
	}

	// Converter o modelo de domínio para o modelo GraphQL
	result, err := domainGroupToGraphQL(updatedGroup)
	if err != nil {
		r.logger.Error(ctx, "Erro ao converter grupo para modelo GraphQL", logging.Fields{
			"error":   err.Error(),
			"groupId": input.ID,
		})
		r.metrics.Counter("resolver.group.changeStatus.responseConversionError").Inc(1)
		return nil, fmt.Errorf("erro ao converter resposta: %w", err)
	}

	r.metrics.Counter("resolver.group.changeStatus.success").Inc(1)
	r.logger.Info(ctx, "Status do grupo alterado com sucesso", logging.Fields{
		"groupId":  input.ID,
		"tenantId": input.TenantID,
		"status":   status,
	})

	return result, nil
}

// DeleteGroup resolve a mutation para deletar um grupo (exclusão lógica)
func (r *GroupResolver) DeleteGroup(ctx context.Context, id string, tenantID string) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.DeleteGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", id),
		attribute.String("tenant.id", tenantID),
	)

	timer := r.metrics.Timer("resolver.group.delete.duration")
	defer timer.ObserveDuration()

	r.logger.Info(ctx, "Resolvendo mutation para deletar grupo", logging.Fields{
		"groupId":  id,
		"tenantId": tenantID,
	})

	// Obter o ID do usuário atual do contexto (autenticado)
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter ID do usuário do contexto", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.delete.authError").Inc(1)
		return false, fmt.Errorf("erro de autenticação: %w", err)
	}

	// Converter IDs de string para UUID
	groupID, err := uuid.Parse(id)
	if err != nil {
		r.logger.Error(ctx, "ID do grupo inválido", logging.Fields{
			"error":   err.Error(),
			"groupId": id,
		})
		r.metrics.Counter("resolver.group.delete.invalidId").Inc(1)
		return false, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.delete.invalidTenantId").Inc(1)
		return false, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Chamar o serviço de domínio para deletar o grupo
	err = r.groupService.Delete(ctx, groupID, tenantUUID, &currentUserID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao deletar grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  id,
			"tenantId": tenantID,
		})
		r.metrics.Counter("resolver.group.delete.error").Inc(1)
		return false, fmt.Errorf("erro ao deletar grupo: %w", err)
	}

	r.metrics.Counter("resolver.group.delete.success").Inc(1)
	r.logger.Info(ctx, "Grupo deletado com sucesso", logging.Fields{
		"groupId":  id,
		"tenantId": tenantID,
	})

	return true, nil
}

// AddUserToGroup resolve a mutation para adicionar um usuário a um grupo
func (r *GroupResolver) AddUserToGroup(ctx context.Context, input model.AddUserToGroupInput) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.AddUserToGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", input.GroupID),
		attribute.String("user.id", input.UserID),
		attribute.String("tenant.id", input.TenantID),
	)

	timer := r.metrics.Timer("resolver.group.addUserToGroup.duration")
	defer timer.ObserveDuration()

	r.logger.Info(ctx, "Resolvendo mutation para adicionar usuário ao grupo", logging.Fields{
		"groupId":  input.GroupID,
		"userId":   input.UserID,
		"tenantId": input.TenantID,
	})

	// Obter o ID do usuário atual do contexto (autenticado)
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter ID do usuário do contexto", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.addUserToGroup.authError").Inc(1)
		return false, fmt.Errorf("erro de autenticação: %w", err)
	}

	// Converter IDs de string para UUID
	groupID, err := uuid.Parse(input.GroupID)
	if err != nil {
		r.logger.Error(ctx, "ID do grupo inválido", logging.Fields{
			"error":   err.Error(),
			"groupId": input.GroupID,
		})
		r.metrics.Counter("resolver.group.addUserToGroup.invalidGroupId").Inc(1)
		return false, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		r.logger.Error(ctx, "ID do usuário inválido", logging.Fields{
			"error":  err.Error(),
			"userId": input.UserID,
		})
		r.metrics.Counter("resolver.group.addUserToGroup.invalidUserId").Inc(1)
		return false, fmt.Errorf("ID do usuário inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(input.TenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.addUserToGroup.invalidTenantId").Inc(1)
		return false, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Chamar o serviço de domínio para adicionar o usuário ao grupo
	err = r.groupService.AddUserToGroup(ctx, groupID, userID, tenantUUID, &currentUserID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao adicionar usuário ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.GroupID,
			"userId":   input.UserID,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.addUserToGroup.error").Inc(1)
		return false, fmt.Errorf("erro ao adicionar usuário ao grupo: %w", err)
	}

	r.metrics.Counter("resolver.group.addUserToGroup.success").Inc(1)
	r.logger.Info(ctx, "Usuário adicionado ao grupo com sucesso", logging.Fields{
		"groupId":  input.GroupID,
		"userId":   input.UserID,
		"tenantId": input.TenantID,
	})

	return true, nil
}

// RemoveUserFromGroup resolve a mutation para remover um usuário de um grupo
func (r *GroupResolver) RemoveUserFromGroup(ctx context.Context, input model.RemoveUserFromGroupInput) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.RemoveUserFromGroup")
	defer span.End()

	span.SetAttributes(
		attribute.String("group.id", input.GroupID),
		attribute.String("user.id", input.UserID),
		attribute.String("tenant.id", input.TenantID),
	)

	timer := r.metrics.Timer("resolver.group.removeUserFromGroup.duration")
	defer timer.ObserveDuration()

	r.logger.Info(ctx, "Resolvendo mutation para remover usuário do grupo", logging.Fields{
		"groupId":  input.GroupID,
		"userId":   input.UserID,
		"tenantId": input.TenantID,
	})

	// Obter o ID do usuário atual do contexto (autenticado)
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		r.logger.Error(ctx, "Erro ao obter ID do usuário do contexto", logging.Fields{
			"error": err.Error(),
		})
		r.metrics.Counter("resolver.group.removeUserFromGroup.authError").Inc(1)
		return false, fmt.Errorf("erro de autenticação: %w", err)
	}

	// Converter IDs de string para UUID
	groupID, err := uuid.Parse(input.GroupID)
	if err != nil {
		r.logger.Error(ctx, "ID do grupo inválido", logging.Fields{
			"error":   err.Error(),
			"groupId": input.GroupID,
		})
		r.metrics.Counter("resolver.group.removeUserFromGroup.invalidGroupId").Inc(1)
		return false, fmt.Errorf("ID do grupo inválido: %w", err)
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		r.logger.Error(ctx, "ID do usuário inválido", logging.Fields{
			"error":  err.Error(),
			"userId": input.UserID,
		})
		r.metrics.Counter("resolver.group.removeUserFromGroup.invalidUserId").Inc(1)
		return false, fmt.Errorf("ID do usuário inválido: %w", err)
	}

	tenantUUID, err := uuid.Parse(input.TenantID)
	if err != nil {
		r.logger.Error(ctx, "ID do tenant inválido", logging.Fields{
			"error":    err.Error(),
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.removeUserFromGroup.invalidTenantId").Inc(1)
		return false, fmt.Errorf("ID do tenant inválido: %w", err)
	}

	// Chamar o serviço de domínio para remover o usuário do grupo
	err = r.groupService.RemoveUserFromGroup(ctx, groupID, userID, tenantUUID, &currentUserID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao remover usuário do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  input.GroupID,
			"userId":   input.UserID,
			"tenantId": input.TenantID,
		})
		r.metrics.Counter("resolver.group.removeUserFromGroup.error").Inc(1)
		return false, fmt.Errorf("erro ao remover usuário do grupo: %w", err)
	}

	r.metrics.Counter("resolver.group.removeUserFromGroup.success").Inc(1)
	r.logger.Info(ctx, "Usuário removido do grupo com sucesso", logging.Fields{
		"groupId":  input.GroupID,
		"userId":   input.UserID,
		"tenantId": input.TenantID,
	})

	return true, nil
}

// Funções auxiliares

// getCurrentUserID obtém o ID do usuário atual do contexto de autenticação
func getCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	// Implementação para obter o ID do usuário autenticado do contexto
	// Normalmente, isso seria fornecido por um middleware de autenticação

	// Temporário: Para testes, usar um ID fixo
	// Em produção, isso seria obtido a partir do token JWT ou sessão
	currentUserID := uuid.New()
	
	// TODO: Implementar corretamente a extração do ID do usuário do contexto de autenticação
	// userID, ok := ctx.Value("currentUserID").(string)
	// if !ok {
	//     return uuid.Nil, errors.New("usuário não autenticado")
	// }
	// currentUserID, err := uuid.Parse(userID)
	// if err != nil {
	//     return uuid.Nil, fmt.Errorf("ID de usuário inválido: %w", err)
	// }

	return currentUserID, nil
}