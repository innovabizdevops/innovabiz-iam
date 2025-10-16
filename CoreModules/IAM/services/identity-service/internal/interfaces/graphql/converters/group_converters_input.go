/**
 * INNOVABIZ IAM - Conversores de Inputs para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos conversores de entradas GraphQL para modelos de domínio
 * para o módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (PR.AC-4: Gerenciamento de identidades e credenciais)
 */

package converters

import (
	"time"

	"github.com/google/uuid"
	
	"github.com/innovabiz/iam/internal/domain/entities"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
)

// GraphQLCreateInputToDomain converte um input de criação GraphQL para modelo de domínio
func GraphQLCreateInputToDomain(input model.CreateGroupInput) (*entities.Group, error) {
	// Converter IDs de UUID
	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return nil, err
	}

	var parentGroupID uuid.UUID
	if input.ParentGroupID != nil {
		parentGroupID, err = uuid.Parse(*input.ParentGroupID)
		if err != nil {
			return nil, err
		}
	}

	// Converter atributos e metadados
	attributes := make(map[string]interface{})
	if input.Attributes != nil {
		for k, v := range *input.Attributes {
			attributes[k] = v
		}
	}

	metadata := make(map[string]interface{})
	if input.Metadata != nil {
		for k, v := range *input.Metadata {
			metadata[k] = v
		}
	}

	// Configurar tipo
	var groupType string
	if input.Type != nil {
		groupType = *input.Type
	}

	// Criar e retornar o modelo de domínio
	return &entities.Group{
		ID:            uuid.New(),
		Code:          input.Code,
		Name:          input.Name,
		Description:   input.Description,
		Status:        entities.GroupStatusActive, // Por padrão, grupos são criados ativos
		Type:          groupType,
		ParentGroupID: parentGroupID,
		Attributes:    attributes,
		Metadata:      metadata,
		TenantID:      tenantID,
		CreatedAt:     time.Now().UTC(),
	}, nil
}

// GraphQLUpdateInputToDomain converte um input de atualização GraphQL para modelo de domínio
func GraphQLUpdateInputToDomain(input model.UpdateGroupInput) (*entities.Group, error) {
	// Converter IDs de UUID
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return nil, err
	}

	var parentGroupID uuid.UUID
	if input.ParentGroupID != nil {
		parentGroupID, err = uuid.Parse(*input.ParentGroupID)
		if err != nil {
			return nil, err
		}
	}

	// Converter atributos e metadados
	attributes := make(map[string]interface{})
	if input.Attributes != nil {
		for k, v := range *input.Attributes {
			attributes[k] = v
		}
	}

	metadata := make(map[string]interface{})
	if input.Metadata != nil {
		for k, v := range *input.Metadata {
			metadata[k] = v
		}
	}

	// Criar modelo de domínio com campos obrigatórios
	group := &entities.Group{
		ID:            id,
		TenantID:      tenantID,
		ParentGroupID: parentGroupID,
		Attributes:    attributes,
		Metadata:      metadata,
		UpdatedAt:     time.Now().UTC(),
	}

	// Adicionar campos opcionais quando fornecidos
	if input.Code != nil {
		group.Code = *input.Code
	}

	if input.Name != nil {
		group.Name = *input.Name
	}

	if input.Description != nil {
		group.Description = *input.Description
	}

	if input.Type != nil {
		group.Type = *input.Type
	}

	return group, nil
}

// GraphQLChangeStatusInputToDomain converte um input de alteração de status para modelo de domínio
func GraphQLChangeStatusInputToDomain(input model.ChangeGroupStatusInput) (*entities.GroupStatusChange, error) {
	// Converter IDs de UUID
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return nil, err
	}

	// Converter status para enumeração de domínio
	var status entities.GroupStatus
	switch input.Status {
	case model.GroupStatusActive:
		status = entities.GroupStatusActive
	case model.GroupStatusInactive:
		status = entities.GroupStatusInactive
	case model.GroupStatusLocked:
		status = entities.GroupStatusLocked
	default:
		status = entities.GroupStatusInactive
	}

	return &entities.GroupStatusChange{
		ID:       id,
		TenantID: tenantID,
		Status:   status,
	}, nil
}

// GraphQLAddUserToGroupInputToDomain converte um input de adição de usuário para modelo de domínio
func GraphQLAddUserToGroupInputToDomain(input model.AddUserToGroupInput) (*entities.GroupUserRelation, error) {
	// Converter IDs de UUID
	groupID, err := uuid.Parse(input.GroupID)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return nil, err
	}

	return &entities.GroupUserRelation{
		GroupID:  groupID,
		UserID:   userID,
		TenantID: tenantID,
	}, nil
}

// GraphQLRemoveUserFromGroupInputToDomain converte um input de remoção de usuário para modelo de domínio
func GraphQLRemoveUserFromGroupInputToDomain(input model.RemoveUserFromGroupInput) (*entities.GroupUserRelation, error) {
	// Converter IDs de UUID
	groupID, err := uuid.Parse(input.GroupID)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return nil, err
	}

	return &entities.GroupUserRelation{
		GroupID:  groupID,
		UserID:   userID,
		TenantID: tenantID,
	}, nil
}