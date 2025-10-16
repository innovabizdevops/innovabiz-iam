/**
 * @file identity_context.go
 * @description Modelo de contexto de identidade para o microserviço de gestão de identidade multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package models

import (
	"time"
	"encoding/json"

	"github.com/google/uuid"
)

// ContextType define os tipos de contexto suportados
type ContextType string

const (
	// ContextFinancial para contexto financeiro
	ContextFinancial ContextType = "financial"
	
	// ContextHealth para contexto de saúde
	ContextHealth ContextType = "health"
	
	// ContextGovernment para contexto governamental
	ContextGovernment ContextType = "government"
	
	// ContextCommerce para contexto de comércio eletrônico
	ContextCommerce ContextType = "commerce"
	
	// ContextEducation para contexto educacional
	ContextEducation ContextType = "education"
	
	// ContextEmployment para contexto de emprego
	ContextEmployment ContextType = "employment"
	
	// ContextMobile para contexto de telefonia móvel
	ContextMobile ContextType = "mobile"
	
	// ContextGeneral para contexto geral
	ContextGeneral ContextType = "general"
)

// ContextStatus define os possíveis estados de um contexto
type ContextStatus string

const (
	// ContextStatusActive indica que o contexto está ativo e pode ser utilizado
	ContextStatusActive ContextStatus = "active"
	
	// ContextStatusInactive indica que o contexto está inativo temporariamente
	ContextStatusInactive ContextStatus = "inactive"
	
	// ContextStatusPendingVerification indica que o contexto ainda requer verificação
	ContextStatusPendingVerification ContextStatus = "pending_verification"
	
	// ContextStatusRevoked indica que o contexto foi revogado
	ContextStatusRevoked ContextStatus = "revoked"
)

// VerificationLevel define os níveis de verificação para um contexto
type VerificationLevel string

const (
	// VerificationNone indica nenhuma verificação realizada
	VerificationNone VerificationLevel = "none"
	
	// VerificationBasic indica verificação básica
	VerificationBasic VerificationLevel = "basic"
	
	// VerificationStandard indica verificação padrão
	VerificationStandard VerificationLevel = "standard"
	
	// VerificationEnhanced indica verificação aprimorada
	VerificationEnhanced VerificationLevel = "enhanced"
	
	// VerificationComplete indica verificação completa
	VerificationComplete VerificationLevel = "complete"
)

// IdentityContext representa um contexto específico associado a uma identidade
type IdentityContext struct {
	// ID único do contexto
	ID uuid.UUID `json:"context_id"`
	
	// ID da identidade associada
	IdentityID uuid.UUID `json:"identity_id"`
	
	// Tipo do contexto
	ContextType ContextType `json:"context_type"`
	
	// Status atual do contexto
	Status ContextStatus `json:"context_status"`
	
	// Pontuação de confiabilidade no contexto
	TrustScore *float64 `json:"trust_score,omitempty"`
	
	// Nível de verificação alcançado no contexto
	VerificationLevel VerificationLevel `json:"verification_level"`
	
	// Data de criação
	CreatedAt time.Time `json:"created_at"`
	
	// Data da última atualização
	UpdatedAt time.Time `json:"updated_at"`
	
	// Atributos específicos deste contexto
	Attributes []*ContextAttribute `json:"attributes,omitempty"`
}

// NewIdentityContext cria um novo contexto de identidade
func NewIdentityContext(identityID uuid.UUID, contextType ContextType) (*IdentityContext, error) {
	if identityID == uuid.Nil {
		return nil, ErrInvalidPrimaryKey
	}
	
	now := time.Now().UTC()
	
	return &IdentityContext{
		ID:               uuid.New(),
		IdentityID:       identityID,
		ContextType:      contextType,
		Status:           ContextStatusPendingVerification,
		VerificationLevel: VerificationNone,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// UpdateStatus atualiza o status do contexto
func (c *IdentityContext) UpdateStatus(status ContextStatus) {
	c.Status = status
	c.UpdatedAt = time.Now().UTC()
}

// UpdateTrustScore atualiza a pontuação de confiança
func (c *IdentityContext) UpdateTrustScore(score float64) {
	c.TrustScore = &score
	c.UpdatedAt = time.Now().UTC()
}

// UpdateVerificationLevel atualiza o nível de verificação
func (c *IdentityContext) UpdateVerificationLevel(level VerificationLevel) {
	c.VerificationLevel = level
	c.UpdatedAt = time.Now().UTC()
	
	// Se o nível de verificação for completo, atualizar status
	if level == VerificationComplete {
		c.Status = ContextStatusActive
	}
}

// AddAttribute adiciona um atributo ao contexto
func (c *IdentityContext) AddAttribute(attr *ContextAttribute) {
	// Verificar se o atributo já existe
	for i, existing := range c.Attributes {
		if existing.AttributeKey == attr.AttributeKey {
			// Substituir atributo existente
			c.Attributes[i] = attr
			c.UpdatedAt = time.Now().UTC()
			return
		}
	}
	
	// Atributo novo
	c.Attributes = append(c.Attributes, attr)
	c.UpdatedAt = time.Now().UTC()
}

// GetAttribute obtém um atributo específico por chave
func (c *IdentityContext) GetAttribute(key string) *ContextAttribute {
	for _, attr := range c.Attributes {
		if attr.AttributeKey == key {
			return attr
		}
	}
	
	return nil
}

// IsActive retorna se o contexto está ativo
func (c *IdentityContext) IsActive() bool {
	return c.Status == ContextStatusActive
}

// MarshalJSON customiza a serialização JSON do contexto
func (c *IdentityContext) MarshalJSON() ([]byte, error) {
	type Alias IdentityContext
	
	return json.Marshal(&struct {
		*Alias
		ID string `json:"context_id"`
		IdentityID string `json:"identity_id"`
	}{
		Alias: (*Alias)(c),
		ID: c.ID.String(),
		IdentityID: c.IdentityID.String(),
	})
}