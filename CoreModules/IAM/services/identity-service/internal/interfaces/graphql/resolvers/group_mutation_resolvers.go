/**
 * INNOVABIZ IAM - Resolvers GraphQL para Mutações de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos resolvers GraphQL para mutações relacionadas a grupos
 * no módulo IAM da plataforma INNOVABIZ, seguindo os princípios de
 * multi-tenant, multi-dimensional, multi-contextual e observabilidade total.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR/PDPA/CCPA (Proteção de dados e controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética para instituições financeiras)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package resolvers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
)

// CreateGroup resolve uma mutação para criar um novo grupo
func (r *GroupResolver) CreateGroup(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.CreateGroup")
	defer span.End()

	// Extrair parâmetros
	inputMap, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("input inválido ou não fornecido")
	}

	// Mapear input para struct CreateGroupInput
	input := group.CreateGroupInput{}

	// Campos obrigatórios
	if code, ok := inputMap["code"].(string); ok {
		input.Code = code
	} else {
		return nil, errors.New("código do grupo é obrigatório")
	}

	if name, ok := inputMap["name"].(string); ok {
		input.Name = name
	} else {
		return nil, errors.New("nome do grupo é obrigatório")
	}

	if tenantIDStr, ok := inputMap["tenantId"].(string); ok {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de tenant inválido")
		}
		input.TenantID = tenantID
	} else {
		return nil, errors.New("tenantId é obrigatório")
	}

	// Campos opcionais
	if description, ok := inputMap["description"].(string); ok {
		input.Description = description
	}

	if groupType, ok := inputMap["groupType"].(string); ok {
		input.GroupType = groupType
	}

	if statusStr, ok := inputMap["status"].(string); ok {
		input.Status = group.Status(statusStr)
	} else {
		input.Status = group.StatusActive
	}

	if regionCode, ok := inputMap["regionCode"].(string); ok {
		input.RegionCode = regionCode
	}

	if parentGroupIDStr, ok := inputMap["parentGroupId"].(string); ok {
		parentGroupID, err := uuid.Parse(parentGroupIDStr)
		if err == nil {
			input.ParentGroupID = &parentGroupID
		}
	}

	if metadata, ok := inputMap["metadata"].(map[string]interface{}); ok {
		input.Metadata = metadata
	}

	// Adicionar ID do usuário atual como criador
	currentUserID := r.authz.GetCurrentUserID(ctx)
	input.CreatedBy = currentUserID

	// Criar grupo
	createdGroup, err := r.groupService.CreateGroup(ctx, input)
	if err != nil {
		r.logger.Error(ctx, "Erro ao criar grupo", logging.Fields{
			"error":    err.Error(),
			"tenantId": input.TenantID.String(),
			"code":     input.Code,
		})

		// Mapear erro para o formato de resposta GraphQL
		result := map[string]interface{}{
			"success":  false,
			"message":  "Falha ao criar grupo: " + err.Error(),
			"group":    nil,
			"errorCode": nil,
		}

		if errors.Is(err, group.ErrGroupAlreadyExists) {
			result["errorCode"] = "GROUP_ALREADY_EXISTS"
		} else if errors.Is(err, group.ErrGroupHierarchyTooDeep) {
			result["errorCode"] = "GROUP_HIERARCHY_TOO_DEEP"
		} else if errors.Is(err, group.ErrGroupCircularReference) {
			result["errorCode"] = "GROUP_CIRCULAR_REFERENCE"
		} else if errors.Is(err, group.ErrInvalidTenant) {
			result["errorCode"] = "INVALID_TENANT"
		} else if errors.Is(err, group.ErrUnauthorizedOperation) {
			result["errorCode"] = "UNAUTHORIZED_OPERATION"
		}

		return result, nil
	}

	// Resultado de sucesso
	return map[string]interface{}{
		"success": true,
		"message": "Grupo criado com sucesso",
		"group":   createdGroup,
	}, nil
}

// UpdateGroup resolve uma mutação para atualizar um grupo existente
func (r *GroupResolver) UpdateGroup(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.UpdateGroup")
	defer span.End()

	// Extrair parâmetros
	inputMap, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("input inválido ou não fornecido")
	}

	// Mapear input para struct UpdateGroupInput
	input := group.UpdateGroupInput{}

	// Campos obrigatórios
	if idStr, ok := inputMap["id"].(string); ok {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, errors.New("formato de ID de grupo inválido")
		}
		input.ID = id
	} else {
		return nil, errors.New("ID do grupo é obrigatório")
	}

	// Campos opcionais
	if name, ok := inputMap["name"].(string); ok {
		input.Name = &name
	}

	if description, ok := inputMap["description"].(string); ok {
		input.Description = &description
	}

	if groupType, ok := inputMap["groupType"].(string); ok {
		input.GroupType = &groupType
	}

	if parentGroupIDStr, ok := inputMap["parentGroupId"].(string); ok {
		parentGroupID, err := uuid.Parse(parentGroupIDStr)
		if err == nil {
			input.ParentGroupID = &parentGroupID
		}
	}

	if metadata, ok := inputMap["metadata"].(map[string]interface{}); ok {
		input.Metadata = metadata
	}

	// Adicionar ID do usuário atual como atualizador
	currentUserID := r.authz.GetCurrentUserID(ctx)
	input.UpdatedBy = &currentUserID

	// Atualizar grupo
	updatedGroup, err := r.groupService.UpdateGroup(ctx, input)
	if err != nil {
		r.logger.Error(ctx, "Erro ao atualizar grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": input.ID.String(),
		})

		// Mapear erro para o formato de resposta GraphQL
		result := map[string]interface{}{
			"success":  false,
			"message":  "Falha ao atualizar grupo: " + err.Error(),
			"group":    nil,
			"errorCode": nil,
		}

		if errors.Is(err, group.ErrGroupNotFound) {
			result["errorCode"] = "GROUP_NOT_FOUND"
		} else if errors.Is(err, group.ErrGroupHierarchyTooDeep) {
			result["errorCode"] = "GROUP_HIERARCHY_TOO_DEEP"
		} else if errors.Is(err, group.ErrGroupCircularReference) {
			result["errorCode"] = "GROUP_CIRCULAR_REFERENCE"
		} else if errors.Is(err, group.ErrCannotModifySystemGroup) {
			result["errorCode"] = "CANNOT_MODIFY_SYSTEM_GROUP"
		} else if errors.Is(err, group.ErrUnauthorizedOperation) {
			result["errorCode"] = "UNAUTHORIZED_OPERATION"
		}

		return result, nil
	}

	// Resultado de sucesso
	return map[string]interface{}{
		"success": true,
		"message": "Grupo atualizado com sucesso",
		"group":   updatedGroup,
	}, nil
}

// ChangeGroupStatus resolve uma mutação para alterar o status de um grupo
func (r *GroupResolver) ChangeGroupStatus(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.ChangeGroupStatus")
	defer span.End()

	// Extrair parâmetros
	inputMap, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("input inválido ou não fornecido")
	}

	// Mapear input para struct ChangeGroupStatusInput
	input := group.ChangeGroupStatusInput{}

	// Campos obrigatórios
	if idStr, ok := inputMap["id"].(string); ok {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, errors.New("formato de ID de grupo inválido")
		}
		input.ID = id
	} else {
		return nil, errors.New("ID do grupo é obrigatório")
	}

	if statusStr, ok := inputMap["status"].(string); ok {
		input.Status = group.Status(statusStr)
	} else {
		return nil, errors.New("status é obrigatório")
	}

	// Campo opcional
	if reason, ok := inputMap["reason"].(string); ok {
		input.Reason = reason
	}

	// Adicionar ID do usuário atual como atualizador
	currentUserID := r.authz.GetCurrentUserID(ctx)
	input.UpdatedBy = &currentUserID

	// Alterar status do grupo
	updatedGroup, err := r.groupService.ChangeGroupStatus(ctx, input)
	if err != nil {
		r.logger.Error(ctx, "Erro ao alterar status do grupo", logging.Fields{
			"error":   err.Error(),
			"groupId": input.ID.String(),
			"status":  string(input.Status),
		})

		// Mapear erro para o formato de resposta GraphQL
		result := map[string]interface{}{
			"success":  false,
			"message":  "Falha ao alterar status do grupo: " + err.Error(),
			"group":    nil,
			"errorCode": nil,
		}

		if errors.Is(err, group.ErrGroupNotFound) {
			result["errorCode"] = "GROUP_NOT_FOUND"
		} else if errors.Is(err, group.ErrInvalidStatus) {
			result["errorCode"] = "INVALID_STATUS"
		} else if errors.Is(err, group.ErrCannotModifySystemGroup) {
			result["errorCode"] = "CANNOT_MODIFY_SYSTEM_GROUP"
		} else if errors.Is(err, group.ErrUnauthorizedOperation) {
			result["errorCode"] = "UNAUTHORIZED_OPERATION"
		}

		return result, nil
	}

	// Resultado de sucesso
	return map[string]interface{}{
		"success": true,
		"message": "Status do grupo alterado com sucesso",
		"group":   updatedGroup,
	}, nil
}

// AddUserToGroup resolve uma mutação para adicionar um usuário a um grupo
func (r *GroupResolver) AddUserToGroup(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.AddUserToGroup")
	defer span.End()

	// Extrair parâmetros
	inputMap, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("input inválido ou não fornecido")
	}

	// Extrair campos obrigatórios
	var groupID, userID, tenantID uuid.UUID
	var err error

	if groupIDStr, ok := inputMap["groupId"].(string); ok {
		groupID, err = uuid.Parse(groupIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de grupo inválido")
		}
	} else {
		return nil, errors.New("groupId é obrigatório")
	}

	if userIDStr, ok := inputMap["userId"].(string); ok {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de usuário inválido")
		}
	} else {
		return nil, errors.New("userId é obrigatório")
	}

	if tenantIDStr, ok := inputMap["tenantId"].(string); ok {
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de tenant inválido")
		}
	} else {
		return nil, errors.New("tenantId é obrigatório")
	}

	// Adicionar usuário ao grupo
	err = r.groupService.AddUserToGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao adicionar usuário ao grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})

		// Mapear erro para o formato de resposta GraphQL
		result := map[string]interface{}{
			"success":  false,
			"message":  "Falha ao adicionar usuário ao grupo: " + err.Error(),
			"group":    nil,
			"user":     nil,
		}

		return result, nil
	}

	// Buscar grupo e usuário para incluir na resposta
	group, _ := r.groupService.GetGroupByID(ctx, groupID, tenantID)
	user, _ := r.userService.GetUserByID(ctx, userID, tenantID)

	// Resultado de sucesso
	return map[string]interface{}{
		"success": true,
		"message": "Usuário adicionado ao grupo com sucesso",
		"group":   group,
		"user":    user,
	}, nil
}

// RemoveUserFromGroup resolve uma mutação para remover um usuário de um grupo
func (r *GroupResolver) RemoveUserFromGroup(ctx context.Context, params graphql.ResolveParams) (interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "GroupResolver.RemoveUserFromGroup")
	defer span.End()

	// Extrair parâmetros
	inputMap, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, errors.New("input inválido ou não fornecido")
	}

	// Extrair campos obrigatórios
	var groupID, userID, tenantID uuid.UUID
	var err error

	if groupIDStr, ok := inputMap["groupId"].(string); ok {
		groupID, err = uuid.Parse(groupIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de grupo inválido")
		}
	} else {
		return nil, errors.New("groupId é obrigatório")
	}

	if userIDStr, ok := inputMap["userId"].(string); ok {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de usuário inválido")
		}
	} else {
		return nil, errors.New("userId é obrigatório")
	}

	if tenantIDStr, ok := inputMap["tenantId"].(string); ok {
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, errors.New("formato de ID de tenant inválido")
		}
	} else {
		return nil, errors.New("tenantId é obrigatório")
	}

	// Buscar grupo e usuário antes da remoção para incluir na resposta
	group, _ := r.groupService.GetGroupByID(ctx, groupID, tenantID)
	user, _ := r.userService.GetUserByID(ctx, userID, tenantID)

	// Remover usuário do grupo
	err = r.groupService.RemoveUserFromGroup(ctx, groupID, userID, tenantID)
	if err != nil {
		r.logger.Error(ctx, "Erro ao remover usuário do grupo", logging.Fields{
			"error":    err.Error(),
			"groupId":  groupID.String(),
			"userId":   userID.String(),
			"tenantId": tenantID.String(),
		})

		// Mapear erro para o formato de resposta GraphQL
		result := map[string]interface{}{
			"success":  false,
			"message":  "Falha ao remover usuário do grupo: " + err.Error(),
			"group":    nil,
			"user":     nil,
		}

		return result, nil
	}

	// Resultado de sucesso
	return map[string]interface{}{
		"success": true,
		"message": "Usuário removido do grupo com sucesso",
		"group":   group,
		"user":    user,
	}, nil
}