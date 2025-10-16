/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Modelo de domínio para funções (roles).
 * Define a estrutura e comportamento das funções no sistema de controle de acesso.
 */

package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RoleType define os tipos de funções disponíveis
type RoleType string

const (
	// RoleTypeSystem representa funções do sistema (geralmente não podem ser modificadas)
	RoleTypeSystem RoleType = "SYSTEM"
	
	// RoleTypeCustom representa funções personalizadas criadas pelos administradores
	RoleTypeCustom RoleType = "CUSTOM"
	
	// RoleTypeDynamic representa funções atribuídas dinamicamente com base em regras
	RoleTypeDynamic RoleType = "DYNAMIC"
)

// Role representa uma função (papel) no sistema de controle de acesso
type Role struct {
	// ID único da função
	ID uuid.UUID `json:"id"`

	// TenantID identifica o tenant ao qual a função pertence
	TenantID uuid.UUID `json:"tenant_id"`

	// Code é o código único da função, usado para identificação em políticas de autorização
	Code string `json:"code"`

	// Name é o nome legível da função
	Name string `json:"name"`

	// Description é uma descrição detalhada da função
	Description string `json:"description"`

	// Type indica o tipo da função (sistema, personalizada, dinâmica)
	Type RoleType `json:"type"`

	// IsActive indica se a função está ativa e pode ser usada
	IsActive bool `json:"is_active"`

	// Priority define a prioridade da função em caso de conflitos de permissões (maior valor = maior prioridade)
	Priority int `json:"priority"`
	
	// ParentID referência opcional à função pai em hierarquia de funções
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	// Metadata armazena informações adicionais da função em formato chave-valor
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// CreatedAt registra quando a função foi criada
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt registra quando a função foi atualizada pela última vez
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRole cria uma nova instância de Role com valores padrão
func NewRole(tenantID uuid.UUID, code, name, description string, roleType RoleType) (*Role, error) {
	// Validação básica
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant ID não pode ser nulo")
	}
	
	if code == "" {
		return nil, fmt.Errorf("código da função não pode ser vazio")
	}
	
	if name == "" {
		return nil, fmt.Errorf("nome da função não pode ser vazio")
	}
	
	// Validar tipo de função
	if roleType != RoleTypeSystem && roleType != RoleTypeCustom && roleType != RoleTypeDynamic {
		roleType = RoleTypeCustom // Valor padrão se inválido
	}

	now := time.Now().UTC()
	
	return &Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		Name:        name,
		Description: description,
		Type:        roleType,
		IsActive:    true,
		Priority:    0, // Prioridade padrão
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate valida a função
func (r *Role) Validate() error {
	if r.ID == uuid.Nil {
		return fmt.Errorf("ID da função não pode ser nulo")
	}
	
	if r.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID não pode ser nulo")
	}
	
	if r.Code == "" {
		return fmt.Errorf("código da função não pode ser vazio")
	}
	
	if r.Name == "" {
		return fmt.Errorf("nome da função não pode ser vazio")
	}
	
	if r.Type != RoleTypeSystem && r.Type != RoleTypeCustom && r.Type != RoleTypeDynamic {
		return fmt.Errorf("tipo de função inválido: %s", r.Type)
	}
	
	return nil
}

// SetParentRole define a função pai na hierarquia
func (r *Role) SetParentRole(parentID uuid.UUID) {
	r.ParentID = &parentID
	r.UpdatedAt = time.Now().UTC()
}

// RemoveParentRole remove a referência à função pai
func (r *Role) RemoveParentRole() {
	r.ParentID = nil
	r.UpdatedAt = time.Now().UTC()
}

// UpdateMetadata adiciona ou atualiza um valor nos metadados da função
func (r *Role) UpdateMetadata(key string, value interface{}) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]interface{})
	}
	r.Metadata[key] = value
	r.UpdatedAt = time.Now().UTC()
}

// GetMetadata recupera um valor dos metadados da função
func (r *Role) GetMetadata(key string) (interface{}, bool) {
	if r.Metadata == nil {
		return nil, false
	}
	value, exists := r.Metadata[key]
	return value, exists
}

// Deactivate desativa a função
func (r *Role) Deactivate() {
	r.IsActive = false
	r.UpdatedAt = time.Now().UTC()
}

// Activate ativa a função
func (r *Role) Activate() {
	r.IsActive = true
	r.UpdatedAt = time.Now().UTC()
}

// SetPriority define a prioridade da função
func (r *Role) SetPriority(priority int) {
	r.Priority = priority
	r.UpdatedAt = time.Now().UTC()
}

// IsSystemRole verifica se a função é do tipo sistema
func (r *Role) IsSystemRole() bool {
	return r.Type == RoleTypeSystem
}

// Clone cria uma cópia independente da função
func (r *Role) Clone() *Role {
	metadata := make(map[string]interface{})
	for k, v := range r.Metadata {
		metadata[k] = v
	}
	
	clone := &Role{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Type:        r.Type,
		IsActive:    r.IsActive,
		Priority:    r.Priority,
		Metadata:    metadata,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	
	if r.ParentID != nil {
		parentID := *r.ParentID
		clone.ParentID = &parentID
	}
	
	return clone
}

// RoleAssignment representa a atribuição de uma função a um usuário
type RoleAssignment struct {
	// ID único da atribuição
	ID uuid.UUID `json:"id"`
	
	// TenantID identifica o tenant ao qual a atribuição pertence
	TenantID uuid.UUID `json:"tenant_id"`
	
	// UserID do usuário ao qual a função é atribuída
	UserID uuid.UUID `json:"user_id"`
	
	// RoleID da função atribuída
	RoleID uuid.UUID `json:"role_id"`
	
	// AssignedBy identifica quem fez a atribuição (opcional)
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty"`
	
	// ExpiresAt define quando a atribuição expira (opcional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	
	// IsActive indica se a atribuição está ativa
	IsActive bool `json:"is_active"`
	
	// Metadata armazena informações adicionais da atribuição
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	
	// CreatedAt registra quando a atribuição foi criada
	CreatedAt time.Time `json:"created_at"`
	
	// UpdatedAt registra quando a atribuição foi atualizada pela última vez
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRoleAssignment cria uma nova atribuição de função a um usuário
func NewRoleAssignment(tenantID, userID, roleID uuid.UUID, assignedBy *uuid.UUID) *RoleAssignment {
	now := time.Now().UTC()
	
	return &RoleAssignment{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
		IsActive:   true,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// SetExpiration define uma data de expiração para a atribuição
func (ra *RoleAssignment) SetExpiration(expiresAt time.Time) {
	ra.ExpiresAt = &expiresAt
	ra.UpdatedAt = time.Now().UTC()
}

// RemoveExpiration remove a data de expiração da atribuição
func (ra *RoleAssignment) RemoveExpiration() {
	ra.ExpiresAt = nil
	ra.UpdatedAt = time.Now().UTC()
}

// IsExpired verifica se a atribuição está expirada
func (ra *RoleAssignment) IsExpired() bool {
	if ra.ExpiresAt == nil {
		return false
	}
	return ra.ExpiresAt.Before(time.Now().UTC())
}

// Deactivate desativa a atribuição
func (ra *RoleAssignment) Deactivate() {
	ra.IsActive = false
	ra.UpdatedAt = time.Now().UTC()
}

// Activate ativa a atribuição
func (ra *RoleAssignment) Activate() {
	ra.IsActive = true
	ra.UpdatedAt = time.Now().UTC()
}

// UpdateMetadata adiciona ou atualiza um valor nos metadados da atribuição
func (ra *RoleAssignment) UpdateMetadata(key string, value interface{}) {
	if ra.Metadata == nil {
		ra.Metadata = make(map[string]interface{})
	}
	ra.Metadata[key] = value
	ra.UpdatedAt = time.Now().UTC()
}