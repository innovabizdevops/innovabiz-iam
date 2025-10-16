/**
 * INNOVABIZ IAM - Conversores para Estatísticas e Filtros de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos conversores para estatísticas e filtros relacionados a grupos
 * para o módulo Core IAM, seguindo a arquitetura multi-dimensional, multi-tenant
 * e com observabilidade total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 */

package converters

import (
	"encoding/json"

	"github.com/google/uuid"
	
	"github.com/innovabiz/iam/internal/domain/entities"
	"github.com/innovabiz/iam/internal/interfaces/graphql/model"
	"github.com/innovabiz/iam/internal/interfaces/graphql/scalars"
)

// GraphQLFilterToDomain converte um filtro GraphQL para filtro de domínio
func GraphQLFilterToDomain(filter *model.GroupFilter) (*entities.GroupFilter, error) {
	if filter == nil {
		return nil, nil
	}

	domainFilter := &entities.GroupFilter{}

	// Converter IDs
	if len(filter.IDs) > 0 {
		domainFilter.IDs = make([]uuid.UUID, 0, len(filter.IDs))
		for _, idStr := range filter.IDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return nil, err
			}
			domainFilter.IDs = append(domainFilter.IDs, id)
		}
	}

	// Converter códigos
	if len(filter.Codes) > 0 {
		domainFilter.Codes = filter.Codes
	}

	// Converter filtros de texto
	if filter.NameContains != nil {
		domainFilter.NameContains = *filter.NameContains
	}

	if filter.DescriptionContains != nil {
		domainFilter.DescriptionContains = *filter.DescriptionContains
	}

	// Converter status
	if len(filter.Statuses) > 0 {
		domainFilter.Statuses = make([]entities.GroupStatus, 0, len(filter.Statuses))
		for _, status := range filter.Statuses {
			switch status {
			case model.GroupStatusActive:
				domainFilter.Statuses = append(domainFilter.Statuses, entities.GroupStatusActive)
			case model.GroupStatusInactive:
				domainFilter.Statuses = append(domainFilter.Statuses, entities.GroupStatusInactive)
			case model.GroupStatusLocked:
				domainFilter.Statuses = append(domainFilter.Statuses, entities.GroupStatusLocked)
			}
		}
	}

	// Converter tipos
	if len(filter.Types) > 0 {
		domainFilter.Types = filter.Types
	}

	// Converter ID de grupo pai
	if filter.ParentGroupID != nil {
		parentID, err := uuid.Parse(*filter.ParentGroupID)
		if err != nil {
			return nil, err
		}
		domainFilter.ParentGroupID = parentID
	}

	// Converter datas
	if filter.CreatedAtStart != nil {
		domainFilter.CreatedAtStart = filter.CreatedAtStart.Time
	}

	if filter.CreatedAtEnd != nil {
		domainFilter.CreatedAtEnd = filter.CreatedAtEnd.Time
	}

	if filter.UpdatedAtStart != nil {
		domainFilter.UpdatedAtStart = filter.UpdatedAtStart.Time
	}

	if filter.UpdatedAtEnd != nil {
		domainFilter.UpdatedAtEnd = filter.UpdatedAtEnd.Time
	}

	// Converter IDs de usuário
	if filter.CreatedBy != nil {
		createdBy, err := uuid.Parse(*filter.CreatedBy)
		if err != nil {
			return nil, err
		}
		domainFilter.CreatedBy = createdBy
	}

	if filter.UpdatedBy != nil {
		updatedBy, err := uuid.Parse(*filter.UpdatedBy)
		if err != nil {
			return nil, err
		}
		domainFilter.UpdatedBy = updatedBy
	}

	// Converter flag hasParent
	if filter.HasParent != nil {
		domainFilter.HasParent = *filter.HasParent
	}

	return domainFilter, nil
}

// DomainGroupStatisticsToGraphQL converte estatísticas de grupo de domínio para modelo GraphQL
func DomainGroupStatisticsToGraphQL(stats *entities.GroupStatistics) *model.GroupStatistics {
	if stats == nil {
		return nil
	}

	// Converter ID de grupo se existir
	var groupID *string
	if stats.GroupID != uuid.Nil {
		groupIDStr := stats.GroupID.String()
		groupID = &groupIDStr
	}

	// Converter distribuições para JSONObject
	var distributionByType scalars.JSONObject
	if stats.DistributionByType != nil {
		byteData, err := json.Marshal(stats.DistributionByType)
		if err == nil {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(byteData, &rawMap); err == nil {
				distributionByType = scalars.JSONObject(rawMap)
			}
		}
	}

	var distributionByLevel scalars.JSONObject
	if stats.DistributionByLevel != nil {
		byteData, err := json.Marshal(stats.DistributionByLevel)
		if err == nil {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(byteData, &rawMap); err == nil {
				distributionByLevel = scalars.JSONObject(rawMap)
			}
		}
	}

	return &model.GroupStatistics{
		TenantID:           stats.TenantID.String(),
		GroupID:            groupID,
		TimestampGenerated: scalars.DateTime{Time: stats.TimestampGenerated},
		TotalGroups:        stats.TotalGroups,
		ActiveGroups:       stats.ActiveGroups,
		InactiveGroups:     stats.InactiveGroups,
		LockedGroups:       stats.LockedGroups,
		DirectUsers:        stats.DirectUsers,
		TotalUsers:         stats.TotalUsers,
		DirectChildGroups:  stats.DirectChildGroups,
		TotalChildGroups:   stats.TotalChildGroups,
		MaxHierarchyDepth:  stats.MaxHierarchyDepth,
		DistributionByType: distributionByType,
		DistributionByLevel: distributionByLevel,
	}
}

// GraphQLSortToDomain converte ordem de classificação GraphQL para domínio
func GraphQLSortToDomain(field string, direction model.SortDirection) entities.SortOption {
	var dir entities.SortDirection
	if direction == model.SortDirectionDesc {
		dir = entities.SortDirectionDesc
	} else {
		dir = entities.SortDirectionAsc
	}

	return entities.SortOption{
		Field:     field,
		Direction: dir,
	}
}