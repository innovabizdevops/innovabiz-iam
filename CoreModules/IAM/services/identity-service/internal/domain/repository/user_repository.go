/*
 * INNOVABIZ IAM - Identity Service
 * Copyright (c) 2025 INNOVABIZ
 *
 * Interface de repositório para o domínio de Usuário no sistema IAM.
 * Implementa princípios de inversão de dependência (SOLID).
 */

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/innovabiz/iam/services/identity-service/internal/domain/model"
)

// UserRepository define a interface para operações de persistência de usuários
// seguindo o princípio de inversão de dependência do SOLID
type UserRepository interface {
	// Create persiste um novo usuário no sistema
	Create(ctx context.Context, user *model.User) error
	
	// GetByID obtém um usuário pelo seu ID
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*model.User, error)
	
	// GetByUsername obtém um usuário pelo seu nome de usuário
	GetByUsername(ctx context.Context, tenantID uuid.UUID, username string) (*model.User, error)
	
	// GetByEmail obtém um usuário pelo seu email
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*model.User, error)
	
	// Update atualiza os dados de um usuário existente
	Update(ctx context.Context, user *model.User) error
	
	// Delete marca um usuário como excluído (soft delete)
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error
	
	// HardDelete remove permanentemente um usuário do sistema
	HardDelete(ctx context.Context, tenantID, userID uuid.UUID) error
	
	// List obtém uma lista paginada de usuários
	List(ctx context.Context, tenantID uuid.UUID, filter UserFilter) ([]*model.User, int64, error)
	
	// FindByExternalID encontra um usuário por ID de provedor externo
	FindByExternalID(ctx context.Context, tenantID uuid.UUID, provider model.AuthProvider, externalID string) (*model.User, error)
	
	// UpdateStatus atualiza apenas o status de um usuário
	UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status model.UserStatus) error
	
	// UpdateCredentials atualiza as credenciais de um usuário
	UpdateCredentials(ctx context.Context, cred *model.UserCredential) error
	
	// GetCredentials obtém as credenciais de um usuário
	GetCredentials(ctx context.Context, tenantID, userID uuid.UUID) (*model.UserCredential, error)
	
	// UpdateMFASettings atualiza as configurações de MFA de um usuário
	UpdateMFASettings(ctx context.Context, userID uuid.UUID, mfa *model.MFASettings) error
	
	// GetMFASettings obtém as configurações de MFA de um usuário
	GetMFASettings(ctx context.Context, userID uuid.UUID) (*model.MFASettings, error)
	
	// AddAddress adiciona um novo endereço para um usuário
	AddAddress(ctx context.Context, address *model.Address) error
	
	// UpdateAddress atualiza um endereço existente
	UpdateAddress(ctx context.Context, address *model.Address) error
	
	// DeleteAddress remove um endereço
	DeleteAddress(ctx context.Context, tenantID, userID, addressID uuid.UUID) error
	
	// GetAddresses obtém todos os endereços de um usuário
	GetAddresses(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Address, error)
	
	// AddContact adiciona um novo contato para um usuário
	AddContact(ctx context.Context, contact *model.Contact) error
	
	// UpdateContact atualiza um contato existente
	UpdateContact(ctx context.Context, contact *model.Contact) error
	
	// DeleteContact remove um contato
	DeleteContact(ctx context.Context, tenantID, userID, contactID uuid.UUID) error
	
	// GetContacts obtém todos os contatos de um usuário
	GetContacts(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.Contact, error)
	
	// CreateSession cria uma nova sessão para um usuário
	CreateSession(ctx context.Context, session *model.UserSession) error
	
	// GetSession obtém uma sessão pelo ID
	GetSession(ctx context.Context, sessionID uuid.UUID) (*model.UserSession, error)
	
	// UpdateSession atualiza uma sessão existente
	UpdateSession(ctx context.Context, session *model.UserSession) error
	
	// DeleteSession encerra uma sessão
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	
	// DeleteAllSessions encerra todas as sessões de um usuário
	DeleteAllSessions(ctx context.Context, tenantID, userID uuid.UUID) error
	
	// ListSessions lista as sessões ativas de um usuário
	ListSessions(ctx context.Context, tenantID, userID uuid.UUID) ([]*model.UserSession, error)
}

// UserFilter define os critérios para filtragem e paginação na busca de usuários
type UserFilter struct {
	// Campos para filtragem
	Status      []model.UserStatus
	EmailVerified *bool
	PhoneVerified *bool
	SearchTerm  string // Busca por nome, email ou username
	
	// Campos para ordenação
	SortBy      string // Campo para ordenação
	SortOrder   string // "asc" ou "desc"
	
	// Campos para paginação
	Offset      int
	Limit       int
}