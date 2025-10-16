/**
 * @file base_identity.go
 * @description Modelo base de identidade para o microserviço de gestão de identidade multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package models

import (
	"time"
	"encoding/json"

	"github.com/google/uuid"
)

// IdentityStatus define os possíveis estados de uma identidade
type IdentityStatus string

const (
	// StatusActive indica que a identidade está ativa e pode ser utilizada
	StatusActive IdentityStatus = "active"
	
	// StatusInactive indica que a identidade está inativa temporariamente
	StatusInactive IdentityStatus = "inactive"
	
	// StatusSuspended indica que a identidade está suspensa por questões de segurança
	StatusSuspended IdentityStatus = "suspended"
	
	// StatusPendingVerification indica que a identidade ainda requer verificação
	StatusPendingVerification IdentityStatus = "pending_verification"
	
	// StatusRevoked indica que a identidade foi revogada e não pode mais ser utilizada
	StatusRevoked IdentityStatus = "revoked"
)

// PrimaryKeyType define os tipos de chaves primárias suportadas
type PrimaryKeyType string

const (
	// PrimaryKeyCPF para CPF brasileiro
	PrimaryKeyCPF PrimaryKeyType = "cpf"
	
	// PrimaryKeyPassport para passaporte
	PrimaryKeyPassport PrimaryKeyType = "passport"
	
	// PrimaryKeyBI para Bilhete de Identidade angolano
	PrimaryKeyBI PrimaryKeyType = "bi"
	
	// PrimaryKeyNIF para Número de Identificação Fiscal
	PrimaryKeyNIF PrimaryKeyType = "nif"
	
	// PrimaryKeyEmail para email
	PrimaryKeyEmail PrimaryKeyType = "email"
	
	// PrimaryKeyPhone para número de telefone
	PrimaryKeyPhone PrimaryKeyType = "phone"
	
	// PrimaryKeyCustom para tipo personalizado
	PrimaryKeyCustom PrimaryKeyType = "custom"
)

// Identity representa a entidade base de identidade no sistema
type Identity struct {
	// ID único da identidade
	ID uuid.UUID `json:"identity_id"`
	
	// Tipo da chave primária utilizada
	PrimaryKeyType PrimaryKeyType `json:"primary_key_type"`
	
	// Valor da chave primária
	PrimaryKeyValue string `json:"primary_key_value"`
	
	// ID opcional de pessoa master (para associação em sistemas externos)
	MasterPersonID *uuid.UUID `json:"master_person_id,omitempty"`
	
	// Status atual da identidade
	Status IdentityStatus `json:"status"`
	
	// Data de criação
	CreatedAt time.Time `json:"created_at"`
	
	// Data da última atualização
	UpdatedAt time.Time `json:"updated_at"`
	
	// Metadados adicionais da identidade (flexível)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	
	// Contextos associados a esta identidade (carregados sob demanda)
	Contexts []*IdentityContext `json:"contexts,omitempty"`
}

// NewIdentity cria uma nova instância de identidade
func NewIdentity(primaryKeyType PrimaryKeyType, primaryKeyValue string) (*Identity, error) {
	if primaryKeyValue == "" {
		return nil, ErrInvalidPrimaryKey
	}
	
	now := time.Now().UTC()
	
	return &Identity{
		ID:             uuid.New(),
		PrimaryKeyType: primaryKeyType,
		PrimaryKeyValue: primaryKeyValue,
		Status:         StatusPendingVerification,
		CreatedAt:      now,
		UpdatedAt:      now,
		Metadata:       make(map[string]interface{}),
	}, nil
}

// UpdateStatus atualiza o status da identidade
func (i *Identity) UpdateStatus(status IdentityStatus) {
	i.Status = status
	i.UpdatedAt = time.Now().UTC()
}

// SetMetadata define um valor nos metadados da identidade
func (i *Identity) SetMetadata(key string, value interface{}) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]interface{})
	}
	
	i.Metadata[key] = value
	i.UpdatedAt = time.Now().UTC()
}

// GetMetadata obtém um valor dos metadados da identidade
func (i *Identity) GetMetadata(key string) (interface{}, bool) {
	if i.Metadata == nil {
		return nil, false
	}
	
	value, exists := i.Metadata[key]
	return value, exists
}

// AddContext adiciona um novo contexto à identidade
func (i *Identity) AddContext(context *IdentityContext) {
	// Verificar se o contexto já existe
	for _, existing := range i.Contexts {
		if existing.ContextType == context.ContextType {
			// Contexto do mesmo tipo já existe, não adicionar duplicado
			return
		}
	}
	
	i.Contexts = append(i.Contexts, context)
	i.UpdatedAt = time.Now().UTC()
}

// GetContext obtém um contexto específico por tipo
func (i *Identity) GetContext(contextType ContextType) *IdentityContext {
	for _, context := range i.Contexts {
		if context.ContextType == contextType {
			return context
		}
	}
	
	return nil
}

// IsVerified retorna se a identidade está completamente verificada
func (i *Identity) IsVerified() bool {
	return i.Status == StatusActive
}

// MarshalJSON customiza a serialização JSON da identidade
func (i *Identity) MarshalJSON() ([]byte, error) {
	type Alias Identity
	
	return json.Marshal(&struct {
		*Alias
		ID string `json:"identity_id"`
		MasterPersonID *string `json:"master_person_id,omitempty"`
	}{
		Alias: (*Alias)(i),
		ID: i.ID.String(),
		MasterPersonID: stringPtrFromUUIDPtr(i.MasterPersonID),
	})
}

// Helper para converter UUID pointer para string pointer
func stringPtrFromUUIDPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}