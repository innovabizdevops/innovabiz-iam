/**
 * INNOVABIZ IAM - Modelos GraphQL para Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos modelos GraphQL para operações relacionadas a grupos no módulo Core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade total
 * da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package model

import (
	"github.com/innovabiz/iam/internal/interfaces/graphql/scalars"
)

// Group representa um grupo no modelo GraphQL
type Group struct {
	ID              string                `json:"id"`
	Code            string                `json:"code"`
	Name            string                `json:"name"`
	Description     string                `json:"description,omitempty"`
	Status          GroupStatus           `json:"status"`
	Type            string                `json:"type,omitempty"`
	ParentGroupID   *string               `json:"parentGroupId,omitempty"`
	ParentGroup     *Group                `json:"parentGroup,omitempty"`
	Attributes      scalars.JSONObject    `json:"attributes,omitempty"`
	Metadata        scalars.JSONObject    `json:"metadata,omitempty"`
	TenantID        string                `json:"tenantId"`
	CreatedAt       scalars.DateTime      `json:"createdAt"`
	UpdatedAt       *scalars.DateTime     `json:"updatedAt,omitempty"`
	CreatedBy       *string               `json:"createdBy,omitempty"`
	UpdatedBy       *string               `json:"updatedBy,omitempty"`
	UserCount       *int                  `json:"userCount,omitempty"`
	ChildGroupsCount *int                 `json:"childGroupsCount,omitempty"`
	Path            *string               `json:"path,omitempty"`
	Level           *int                  `json:"level,omitempty"`
}

// GroupStatus representa o status de um grupo no modelo GraphQL
type GroupStatus string

// Definição das constantes para status de grupo
const (
	GroupStatusActive   GroupStatus = "ACTIVE"
	GroupStatusInactive GroupStatus = "INACTIVE"
	GroupStatusLocked   GroupStatus = "LOCKED"
)

// GroupListResult representa o resultado de uma consulta paginada de grupos
type GroupListResult struct {
	Items      []*Group   `json:"items"`
	TotalCount int        `json:"totalCount"`
	PageInfo   *PageInfo  `json:"pageInfo"`
}

// GroupStatistics representa estatísticas de grupos no modelo GraphQL
type GroupStatistics struct {
	TenantID           string              `json:"tenantId"`
	GroupID            *string             `json:"groupId,omitempty"`
	TimestampGenerated scalars.DateTime    `json:"timestampGenerated"`
	TotalGroups        int                 `json:"totalGroups"`
	ActiveGroups       int                 `json:"activeGroups"`
	InactiveGroups     int                 `json:"inactiveGroups"`
	LockedGroups       int                 `json:"lockedGroups"`
	DirectUsers        int                 `json:"directUsers"`
	TotalUsers         int                 `json:"totalUsers"`
	DirectChildGroups  int                 `json:"directChildGroups"`
	TotalChildGroups   int                 `json:"totalChildGroups"`
	MaxHierarchyDepth  int                 `json:"maxHierarchyDepth"`
	DistributionByType scalars.JSONObject  `json:"distributionByType,omitempty"`
	DistributionByLevel scalars.JSONObject `json:"distributionByLevel,omitempty"`
}

// CreateGroupInput representa os dados de entrada para criação de grupo
type CreateGroupInput struct {
	Code          string               `json:"code"`
	Name          string               `json:"name"`
	Description   string               `json:"description,omitempty"`
	Type          *string              `json:"type,omitempty"`
	ParentGroupID *string              `json:"parentGroupId,omitempty"`
	Attributes    *scalars.JSONObject  `json:"attributes,omitempty"`
	Metadata      *scalars.JSONObject  `json:"metadata,omitempty"`
	TenantID      string               `json:"tenantId"`
}

// UpdateGroupInput representa os dados de entrada para atualização de grupo
type UpdateGroupInput struct {
	ID            string               `json:"id"`
	Code          *string              `json:"code,omitempty"`
	Name          *string              `json:"name,omitempty"`
	Description   *string              `json:"description,omitempty"`
	Type          *string              `json:"type,omitempty"`
	ParentGroupID *string              `json:"parentGroupId,omitempty"`
	Attributes    *scalars.JSONObject  `json:"attributes,omitempty"`
	Metadata      *scalars.JSONObject  `json:"metadata,omitempty"`
	TenantID      string               `json:"tenantId"`
}

// GroupFilter representa os filtros para consulta de grupos
type GroupFilter struct {
	IDs                 []string            `json:"ids,omitempty"`
	Codes               []string            `json:"codes,omitempty"`
	NameContains        *string             `json:"nameContains,omitempty"`
	DescriptionContains *string             `json:"descriptionContains,omitempty"`
	Statuses            []GroupStatus       `json:"statuses,omitempty"`
	Types               []string            `json:"types,omitempty"`
	ParentGroupID       *string             `json:"parentGroupId,omitempty"`
	CreatedAtStart      *scalars.DateTime   `json:"createdAtStart,omitempty"`
	CreatedAtEnd        *scalars.DateTime   `json:"createdAtEnd,omitempty"`
	UpdatedAtStart      *scalars.DateTime   `json:"updatedAtStart,omitempty"`
	UpdatedAtEnd        *scalars.DateTime   `json:"updatedAtEnd,omitempty"`
	CreatedBy           *string             `json:"createdBy,omitempty"`
	UpdatedBy           *string             `json:"updatedBy,omitempty"`
	HasParent           *bool               `json:"hasParent,omitempty"`
}

// ChangeGroupStatusInput representa os dados de entrada para alteração de status de grupo
type ChangeGroupStatusInput struct {
	ID        string      `json:"id"`
	Status    GroupStatus `json:"status"`
	TenantID  string      `json:"tenantId"`
}

// AddUserToGroupInput representa os dados de entrada para adicionar usuário a grupo
type AddUserToGroupInput struct {
	GroupID   string     `json:"groupId"`
	UserID    string     `json:"userId"`
	TenantID  string     `json:"tenantId"`
}

// RemoveUserFromGroupInput representa os dados de entrada para remover usuário de grupo
type RemoveUserFromGroupInput struct {
	GroupID   string     `json:"groupId"`
	UserID    string     `json:"userId"`
	TenantID  string     `json:"tenantId"`
}

// PageInfo representa informações de paginação para consultas GraphQL
type PageInfo struct {
	CurrentPage      int  `json:"currentPage"`
	PageSize         int  `json:"pageSize"`
	TotalPages       int  `json:"totalPages"`
	HasNextPage      bool `json:"hasNextPage"`
	HasPreviousPage  bool `json:"hasPreviousPage"`
}

// SortDirection representa a direção de ordenação para consultas
type SortDirection string

// Definição das constantes para direção de ordenação
const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

// UserFilter representa os filtros para consulta de usuários
type UserFilter struct {
	IDs               []string            `json:"ids,omitempty"`
	UsernameContains  *string             `json:"usernameContains,omitempty"`
	EmailContains     *string             `json:"emailContains,omitempty"`
	Statuses          []string            `json:"statuses,omitempty"`
	Types             []string            `json:"types,omitempty"`
	CreatedAtStart    *scalars.DateTime   `json:"createdAtStart,omitempty"`
	CreatedAtEnd      *scalars.DateTime   `json:"createdAtEnd,omitempty"`
	UpdatedAtStart    *scalars.DateTime   `json:"updatedAtStart,omitempty"`
	UpdatedAtEnd      *scalars.DateTime   `json:"updatedAtEnd,omitempty"`
}

// UserListResult representa o resultado de uma consulta paginada de usuários
type UserListResult struct {
	Items      []*User    `json:"items"`
	TotalCount int        `json:"totalCount"`
	PageInfo   *PageInfo  `json:"pageInfo"`
}

// User representa um usuário no modelo GraphQL (referenciado em UserListResult)
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	// outros campos de usuário
}