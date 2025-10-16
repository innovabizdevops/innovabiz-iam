/**
 * INNOVABIZ IAM - Conversores de Modelos para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos conversores entre modelos de domínio e modelos GraphQL
 * para o módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - SOX (Rastreabilidade e auditoria)
 */

package converters

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	
	"github.com/innovabiz/iam/internal/domain/entities"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
	"github.com/innovabiz/iam/internal/interfaces/graphql/scalars"
)

// DomainGroupToGraphQL converte um grupo de domínio para modelo GraphQL
func DomainGroupToGraphQL(domainGroup *entities.Group) *model.Group {
	if domainGroup == nil {
		return nil
	}

	// Converter atributos e metadados para JSONObject
	var attributes scalars.JSONObject
	if domainGroup.Attributes != nil {
		attributes = scalars.JSONObject{}
		for k, v := range domainGroup.Attributes {
			attributes[k] = v
		}
	}

	var metadata scalars.JSONObject
	if domainGroup.Metadata != nil {
		metadata = scalars.JSONObject{}
		for k, v := range domainGroup.Metadata {
			metadata[k] = v
		}
	}

	// Converter status para enumeração GraphQL
	var status model.GroupStatus
	switch domainGroup.Status {
	case entities.GroupStatusActive:
		status = model.GroupStatusActive
	case entities.GroupStatusInactive:
		status = model.GroupStatusInactive
	case entities.GroupStatusLocked:
		status = model.GroupStatusLocked
	default:
		status = model.GroupStatusInactive
	}

	// Converter ID de grupo pai se existir
	var parentGroupID *string
	if domainGroup.ParentGroupID != uuid.Nil {
		parentIDStr := domainGroup.ParentGroupID.String()
		parentGroupID = &parentIDStr
	}

	// Converter timestamps
	createdAt := scalars.DateTime{Time: domainGroup.CreatedAt}
	var updatedAt *scalars.DateTime
	if !domainGroup.UpdatedAt.IsZero() {
		updatedAt = &scalars.DateTime{Time: domainGroup.UpdatedAt}
	}

	// Converter IDs de usuário
	var createdBy *string
	if domainGroup.CreatedBy != uuid.Nil {
		createdByStr := domainGroup.CreatedBy.String()
		createdBy = &createdByStr
	}

	var updatedBy *string
	if domainGroup.UpdatedBy != uuid.Nil {
		updatedByStr := domainGroup.UpdatedBy.String()
		updatedBy = &updatedByStr
	}

	// Montar e retornar o modelo GraphQL
	return &model.Group{
		ID:            domainGroup.ID.String(),
		Code:          domainGroup.Code,
		Name:          domainGroup.Name,
		Description:   domainGroup.Description,
		Status:        status,
		Type:          domainGroup.Type,
		ParentGroupID: parentGroupID,
		Attributes:    attributes,
		Metadata:      metadata,
		TenantID:      domainGroup.TenantID.String(),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		CreatedBy:     createdBy,
		UpdatedBy:     updatedBy,
		Path:          domainGroup.Path,
		Level:         &domainGroup.Level,
	}
}