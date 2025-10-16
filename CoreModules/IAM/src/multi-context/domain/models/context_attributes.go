/**
 * @file context_attributes.go
 * @description Modelo de atributos contextuais para o microserviço de gestão de identidade multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package models

import (
	"time"
	"encoding/json"

	"github.com/google/uuid"
)

// SensitivityLevel define os níveis de sensibilidade de um atributo
type SensitivityLevel string

const (
	// SensitivityPublic para atributos públicos
	SensitivityPublic SensitivityLevel = "public"
	
	// SensitivityLow para atributos de baixa sensibilidade
	SensitivityLow SensitivityLevel = "low"
	
	// SensitivityMedium para atributos de média sensibilidade
	SensitivityMedium SensitivityLevel = "medium"
	
	// SensitivityHigh para atributos de alta sensibilidade
	SensitivityHigh SensitivityLevel = "high"
	
	// SensitivityCritical para atributos críticos
	SensitivityCritical SensitivityLevel = "critical"
)

// VerificationStatus define os status de verificação de um atributo
type VerificationStatus string

const (
	// VerificationStatusUnverified indica que o atributo não foi verificado
	VerificationStatusUnverified VerificationStatus = "unverified"
	
	// VerificationStatusPending indica verificação em andamento
	VerificationStatusPending VerificationStatus = "pending"
	
	// VerificationStatusVerified indica atributo verificado
	VerificationStatusVerified VerificationStatus = "verified"
	
	// VerificationStatusFailed indica falha na verificação
	VerificationStatusFailed VerificationStatus = "failed"
	
	// VerificationStatusExpired indica verificação expirada
	VerificationStatusExpired VerificationStatus = "expired"
)

// ContextAttribute representa um atributo específico de um contexto
type ContextAttribute struct {
	// ID único do atributo
	ID uuid.UUID `json:"attribute_id"`
	
	// ID do contexto ao qual este atributo pertence
	ContextID uuid.UUID `json:"context_id"`
	
	// Chave do atributo
	AttributeKey string `json:"attribute_key"`
	
	// Valor do atributo
	AttributeValue string `json:"attribute_value"`
	
	// Nível de sensibilidade do atributo
	SensitivityLevel SensitivityLevel `json:"sensitivity_level"`
	
	// Status de verificação do atributo
	VerificationStatus VerificationStatus `json:"verification_status"`
	
	// Fonte da verificação (se aplicável)
	VerificationSource string `json:"verification_source,omitempty"`
	
	// Data de criação
	CreatedAt time.Time `json:"created_at"`
	
	// Data da última atualização
	UpdatedAt time.Time `json:"updated_at"`
	
	// Metadados adicionais do atributo (flexível)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewContextAttribute cria um novo atributo de contexto
func NewContextAttribute(
	contextID uuid.UUID,
	key string,
	value string,
	sensitivity SensitivityLevel,
) (*ContextAttribute, error) {
	if contextID == uuid.Nil || key == "" {
		return nil, ErrInvalidAttribute
	}
	
	now := time.Now().UTC()
	
	return &ContextAttribute{
		ID:                uuid.New(),
		ContextID:         contextID,
		AttributeKey:      key,
		AttributeValue:    value,
		SensitivityLevel:  sensitivity,
		VerificationStatus: VerificationStatusUnverified,
		CreatedAt:         now,
		UpdatedAt:         now,
		Metadata:          make(map[string]interface{}),
	}, nil
}

// UpdateValue atualiza o valor do atributo
func (a *ContextAttribute) UpdateValue(value string) {
	a.AttributeValue = value
	a.UpdatedAt = time.Now().UTC()
	
	// Resetar verificação se o valor foi alterado
	if a.VerificationStatus == VerificationStatusVerified {
		a.VerificationStatus = VerificationStatusUnverified
	}
}

// UpdateVerification atualiza o status e fonte de verificação
func (a *ContextAttribute) UpdateVerification(status VerificationStatus, source string) {
	a.VerificationStatus = status
	a.VerificationSource = source
	a.UpdatedAt = time.Now().UTC()
	
	// Adicionar timestamp de verificação nos metadados
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	
	if status == VerificationStatusVerified {
		a.Metadata["verified_at"] = time.Now().UTC().Format(time.RFC3339)
	}
}

// SetMetadata define um valor nos metadados do atributo
func (a *ContextAttribute) SetMetadata(key string, value interface{}) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	
	a.Metadata[key] = value
	a.UpdatedAt = time.Now().UTC()
}

// GetMetadata obtém um valor dos metadados do atributo
func (a *ContextAttribute) GetMetadata(key string) (interface{}, bool) {
	if a.Metadata == nil {
		return nil, false
	}
	
	value, exists := a.Metadata[key]
	return value, exists
}

// IsVerified retorna se o atributo está verificado
func (a *ContextAttribute) IsVerified() bool {
	return a.VerificationStatus == VerificationStatusVerified
}

// MarshalJSON customiza a serialização JSON do atributo
func (a *ContextAttribute) MarshalJSON() ([]byte, error) {
	type Alias ContextAttribute
	
	return json.Marshal(&struct {
		*Alias
		ID string `json:"attribute_id"`
		ContextID string `json:"context_id"`
	}{
		Alias: (*Alias)(a),
		ID: a.ID.String(),
		ContextID: a.ContextID.String(),
	})
}