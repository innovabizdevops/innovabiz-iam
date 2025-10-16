/**
 * INNOVABIZ IAM - Serviço de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Este serviço implementa operações de gerenciamento de grupos
 * para o módulo IAM da plataforma INNOVABIZ, incluindo:
 * - Criação, atualização e gestão de grupos
 * - Gestão de membros (usuários) nos grupos
 * - Hierarquia de grupos com suporte a múltiplos níveis
 * - Integração com políticas de autorização e observabilidade
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Gestão de acesso)
 * - PCI DSS v4.0 (Requisito 7: Controle de acesso)
 * - LGPD/GDPR/PDPA/CCPA (Minimização de dados e controle de acesso)
 * - BNA Instrução 7/2021 (Requisitos de segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package group

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/internal/domain/tenant"
	"github.com/innovabiz/iam/internal/domain/user"
	"github.com/innovabiz/iam/internal/infrastructure/database"
	"github.com/innovabiz/iam/internal/infrastructure/events"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"github.com/innovabiz/iam/internal/infrastructure/security"
	"github.com/innovabiz/iam/internal/infrastructure/validation"
	"github.com/innovabiz/platform/errors"
)

// Status representa os status possíveis para um grupo
type Status string

const (
	// StatusActive indica que o grupo está ativo e operacional
	StatusActive Status = "ACTIVE"

	// StatusInactive indica que o grupo está temporariamente inativo
	StatusInactive Status = "INACTIVE"

	// StatusDeleted indica que o grupo está marcado para exclusão
	StatusDeleted Status = "DELETED"

	// StatusPending indica que o grupo está em fase de aprovação
	StatusPending Status = "PENDING"
)

// MaxGroupHierarchyDepth define a profundidade máxima permitida na hierarquia de grupos
const MaxGroupHierarchyDepth = 10

// MetricNamespaceGroup define o namespace para métricas relacionadas a grupos
const MetricNamespaceGroup = "innovabiz:iam:group"

// Erros específicos do domínio de grupo
var (
	// ErrGroupNotFound indica que o grupo não foi encontrado
	ErrGroupNotFound = errors.New("grupo não encontrado")

	// ErrGroupAlreadyExists indica que já existe um grupo com o mesmo código
	ErrGroupAlreadyExists = errors.New("já existe um grupo com este código")

	// ErrGroupHierarchyTooDeep indica que a hierarquia de grupos excederia a profundidade máxima
	ErrGroupHierarchyTooDeep = errors.New("hierarquia de grupos excederia a profundidade máxima permitida")

	// ErrGroupCircularReference indica que a operação resultaria em uma referência circular
	ErrGroupCircularReference = errors.New("referência circular detectada na hierarquia de grupos")

	// ErrUserNotFound indica que o usuário não foi encontrado
	ErrUserNotFound = errors.New("usuário não encontrado")

	// ErrUserAlreadyInGroup indica que o usuário já é membro do grupo
	ErrUserAlreadyInGroup = errors.New("usuário já é membro deste grupo")

	// ErrUserNotInGroup indica que o usuário não é membro do grupo
	ErrUserNotInGroup = errors.New("usuário não é membro deste grupo")

	// ErrInvalidTenant indica que o tenant é inválido
	ErrInvalidTenant = errors.New("tenant inválido")

	// ErrUnauthorizedOperation indica que a operação não está autorizada
	ErrUnauthorizedOperation = errors.New("operação não autorizada")

	// ErrInvalidStatus indica que o status informado é inválido
	ErrInvalidStatus = errors.New("status inválido")

	// ErrCannotModifySystemGroup indica que não é possível modificar um grupo do sistema
	ErrCannotModifySystemGroup = errors.New("não é possível modificar um grupo do sistema")
)

// Group representa a entidade de grupo no domínio
type Group struct {
	ID           uuid.UUID          `json:"id"`
	Code         string             `json:"code"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Status       Status             `json:"status"`
	GroupType    string             `json:"groupType"`
	Path         string             `json:"path"`
	Level        int                `json:"level"`
	Metadata     map[string]interface{} `json:"metadata"`
	TenantID     uuid.UUID          `json:"tenantId"`
	RegionCode   string             `json:"regionCode"`
	ParentGroupID *uuid.UUID         `json:"parentGroupId"`
	IsSystem     bool               `json:"isSystem"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    *time.Time          `json:"updatedAt"`
	CreatedBy    *uuid.UUID          `json:"createdBy"`
	UpdatedBy    *uuid.UUID          `json:"updatedBy"`
}

// GroupStats representa estatísticas sobre um grupo
type GroupStats struct {
	TotalUsers      int `json:"totalUsers"`
	ActiveUsers     int `json:"activeUsers"`
	DirectUsers     int `json:"directUsers"`
	NestedUsers     int `json:"nestedUsers"`
	DirectSubgroups int `json:"directSubgroups"`
	TotalSubgroups  int `json:"totalSubgroups"`
	MaxDepth        int `json:"maxDepth"`
}

// GroupFilter representa os critérios para filtrar grupos
type GroupFilter struct {
	Status       []Status   `json:"status"`
	GroupType    string     `json:"groupType"`
	MaxLevel     *int        `json:"maxLevel"`
	ParentGroupID *uuid.UUID  `json:"parentGroupId"`
	MemberUserID  *uuid.UUID  `json:"memberUserId"`
	SearchTerm   string     `json:"searchTerm"`
	SortBy       string     `json:"sortBy"`
	SortDirection string    `json:"sortDirection"`
	Page         int        `json:"page"`
	PageSize     int        `json:"pageSize"`
}

// PaginatedGroups representa um resultado paginado de grupos
type PaginatedGroups struct {
	Groups     []Group `json:"groups"`
	TotalCount int     `json:"totalCount"`
	Page       int     `json:"page"`
	PageSize   int     `json:"pageSize"`
	TotalPages int     `json:"totalPages"`
}

// CreateGroupInput representa os dados para criar um novo grupo
type CreateGroupInput struct {
	Code         string             `json:"code" validate:"required,min=3,max=50"`
	Name         string             `json:"name" validate:"required,min=3,max=100"`
	Description  string             `json:"description" validate:"max=500"`
	GroupType    string             `json:"groupType" validate:"max=50"`
	Status       Status             `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE PENDING"`
	TenantID     uuid.UUID          `json:"tenantId" validate:"required"`
	RegionCode   string             `json:"regionCode" validate:"omitempty,max=10"`
	ParentGroupID *uuid.UUID         `json:"parentGroupId"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedBy    uuid.UUID          `json:"createdBy"`
}