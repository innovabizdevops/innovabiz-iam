/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Modelo de domínio para permissões.
 * Define a estrutura e comportamento das permissões no sistema.
 */

package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Permission representa uma permissão no sistema
type Permission struct {
	// ID único da permissão
	ID uuid.UUID `json:"id"`

	// TenantID identifica o tenant ao qual a permissão pertence
	TenantID uuid.UUID `json:"tenant_id"`

	// Code é o código único da permissão, usado para identificação em políticas de autorização
	// Formato recomendado: "{module}:{resource}:{action}" (ex: "users:profile:read")
	Code string `json:"code"`

	// Name é o nome legível da permissão
	Name string `json:"name"`

	// Description é uma descrição detalhada da permissão
	Description string `json:"description"`

	// Module é o módulo do sistema ao qual a permissão está associada
	Module string `json:"module"`

	// Resource é o recurso protegido pela permissão
	Resource string `json:"resource"`

	// Action é a ação permitida sobre o recurso
	Action string `json:"action"`

	// IsActive indica se a permissão está ativa e pode ser usada
	IsActive bool `json:"is_active"`

	// Metadata armazena informações adicionais da permissão em formato chave-valor
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// CreatedAt registra quando a permissão foi criada
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt registra quando a permissão foi atualizada pela última vez
	UpdatedAt time.Time `json:"updated_at"`
}

// NewPermission cria uma nova instância de Permission com valores padrão
func NewPermission(tenantID uuid.UUID, code, name, description, module, resource, action string) (*Permission, error) {
	// Validação básica
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant ID não pode ser nulo")
	}
	
	if code == "" {
		return nil, fmt.Errorf("código da permissão não pode ser vazio")
	}
	
	if name == "" {
		return nil, fmt.Errorf("nome da permissão não pode ser vazio")
	}
	
	if module == "" {
		return nil, fmt.Errorf("módulo da permissão não pode ser vazio")
	}
	
	if resource == "" {
		return nil, fmt.Errorf("recurso da permissão não pode ser vazio")
	}
	
	if action == "" {
		return nil, fmt.Errorf("ação da permissão não pode ser vazia")
	}

	now := time.Now().UTC()
	
	return &Permission{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		Name:        name,
		Description: description,
		Module:      module,
		Resource:    resource,
		Action:      action,
		IsActive:    true,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Validate valida a permissão
func (p *Permission) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("ID da permissão não pode ser nulo")
	}
	
	if p.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID não pode ser nulo")
	}
	
	if p.Code == "" {
		return fmt.Errorf("código da permissão não pode ser vazio")
	}
	
	if p.Name == "" {
		return fmt.Errorf("nome da permissão não pode ser vazio")
	}
	
	if p.Module == "" {
		return fmt.Errorf("módulo da permissão não pode ser vazio")
	}
	
	if p.Resource == "" {
		return fmt.Errorf("recurso da permissão não pode ser vazio")
	}
	
	if p.Action == "" {
		return fmt.Errorf("ação da permissão não pode ser vazia")
	}
	
	return nil
}

// StandardCode formata o código da permissão no padrão "{module}:{resource}:{action}"
func (p *Permission) StandardCode() string {
	return fmt.Sprintf("%s:%s:%s", p.Module, p.Resource, p.Action)
}

// UpdateMetadata adiciona ou atualiza um valor nos metadados da permissão
func (p *Permission) UpdateMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
	p.UpdatedAt = time.Now().UTC()
}

// GetMetadata recupera um valor dos metadados da permissão
func (p *Permission) GetMetadata(key string) (interface{}, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	value, exists := p.Metadata[key]
	return value, exists
}

// Deactivate desativa a permissão
func (p *Permission) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now().UTC()
}

// Activate ativa a permissão
func (p *Permission) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now().UTC()
}

// Clone cria uma cópia independente da permissão
func (p *Permission) Clone() *Permission {
	metadata := make(map[string]interface{})
	for k, v := range p.Metadata {
		metadata[k] = v
	}
	
	return &Permission{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Code:        p.Code,
		Name:        p.Name,
		Description: p.Description,
		Module:      p.Module,
		Resource:    p.Resource,
		Action:      p.Action,
		IsActive:    p.IsActive,
		Metadata:    metadata,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}