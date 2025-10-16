/**
 * INNOVABIZ IAM - Interface do Repositório de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Interface de repositório para persistência de dados de grupos no módulo core IAM,
 * seguindo a arquitetura multi-dimensional, multi-tenant e observabilidade total
 * da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7)
 * - LGPD/GDPR/PDPA (Proteção de dados)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package group

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para operações de persistência de grupos
type Repository interface {
	// GetByID busca um grupo pelo ID
	GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Group, error)
	
	// GetByCode busca um grupo pelo código
	GetByCode(ctx context.Context, code string, tenantID uuid.UUID) (*Group, error)
	
	// Create cria um novo grupo
	Create(ctx context.Context, g *Group) error
	
	// Update atualiza um grupo existente
	Update(ctx context.Context, g *Group) error
	
	// List lista grupos com filtros e paginação
	List(ctx context.Context, tenantID uuid.UUID, filter GroupFilter) (*GroupListResult, error)
	
	// AddUserToGroup adiciona um usuário a um grupo
	AddUserToGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error
	
	// RemoveUserFromGroup remove um usuário de um grupo
	RemoveUserFromGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) error
	
	// IsUserInGroup verifica se um usuário pertence a um grupo
	IsUserInGroup(ctx context.Context, groupID, userID, tenantID uuid.UUID) (bool, error)
	
	// ListGroupMembers lista os usuários membros de um grupo
	ListGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool, page, pageSize int, filter map[string]interface{}) (*UserListResult, error)
	
	// CountGroupMembers conta os membros de um grupo
	CountGroupMembers(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool) (int, error)
	
	// CountSubgroups conta os subgrupos de um grupo
	CountSubgroups(ctx context.Context, groupID, tenantID uuid.UUID, recursive bool) (int, error)
	
	// GetGroupHierarchy busca a hierarquia de grupos ancestrais
	GetGroupHierarchy(ctx context.Context, groupID, tenantID uuid.UUID) ([]*Group, error)
	
	// GetGroupStats obtém estatísticas detalhadas de um grupo
	GetGroupStats(ctx context.Context, groupID, tenantID uuid.UUID) (*GroupStats, error)
	
	// CheckGroupCircularReference verifica se existe referência circular na hierarquia de grupos
	CheckGroupCircularReference(ctx context.Context, groupID, parentGroupID, tenantID uuid.UUID) (bool, error)
	
	// GetGroupLevel obtém o nível hierárquico de um grupo
	GetGroupLevel(ctx context.Context, groupID, tenantID uuid.UUID) (int, error)
	
	// GetGroupsByUser busca grupos aos quais um usuário pertence
	GetGroupsByUser(ctx context.Context, userID, tenantID uuid.UUID) ([]*Group, error)
	
	// GetGroupsByRole busca grupos que têm um determinado papel
	GetGroupsByRole(ctx context.Context, roleID, tenantID uuid.UUID) ([]*Group, error)
	
	// BeginTx inicia uma transação
	BeginTx(ctx context.Context) (context.Context, error)
	
	// CommitTx confirma uma transação
	CommitTx(ctx context.Context) error
	
	// RollbackTx reverte uma transação
	RollbackTx(ctx context.Context) error
}