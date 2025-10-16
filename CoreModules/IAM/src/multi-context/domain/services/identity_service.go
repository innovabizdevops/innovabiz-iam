/**
 * @file identity_service.go
 * @description Serviço de domínio para gerenciamento de identidades multi-contexto
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/repositories"
)

// TrustGuardClient define interface para integração com TrustGuard
type TrustGuardClient interface {
	// EvaluateIdentity avalia a identidade e retorna pontuação de confiança
	EvaluateIdentity(ctx context.Context, identityID string, contextType string) (float64, error)
	
	// VerifyAttribute verifica um atributo específico
	VerifyAttribute(ctx context.Context, attributeKey, attributeValue string, contextType string) (bool, string, error)
}

// IdentityService encapsula a lógica de domínio para identidades multi-contexto
type IdentityService struct {
	identityRepo    repositories.IdentityRepository
	contextRepo     repositories.ContextRepository
	attributeRepo   repositories.AttributeRepository
	trustGuardClient TrustGuardClient
}

// NewIdentityService cria uma nova instância do serviço de identidade
func NewIdentityService(
	identityRepo repositories.IdentityRepository,
	contextRepo repositories.ContextRepository,
	attributeRepo repositories.AttributeRepository,
	trustGuardClient TrustGuardClient,
) *IdentityService {
	return &IdentityService{
		identityRepo:     identityRepo,
		contextRepo:      contextRepo,
		attributeRepo:    attributeRepo,
		trustGuardClient: trustGuardClient,
	}
}

// CreateIdentity cria uma nova identidade
func (s *IdentityService) CreateIdentity(
	ctx context.Context, 
	primaryKeyType models.PrimaryKeyType, 
	primaryKeyValue string,
	masterPersonID *uuid.UUID,
	metadata map[string]interface{},
) (*models.Identity, error) {
	// Verificar se identidade já existe
	existing, err := s.identityRepo.GetByPrimaryKey(ctx, primaryKeyType, primaryKeyValue)
	if err == nil && existing != nil {
		return nil, models.ErrDuplicateIdentity
	}
	
	// Criar nova identidade
	identity, err := models.NewIdentity(primaryKeyType, primaryKeyValue)
	if err != nil {
		return nil, err
	}
	
	// Configurar dados adicionais
	if masterPersonID != nil {
		identity.MasterPersonID = masterPersonID
	}
	
	if metadata != nil {
		for key, value := range metadata {
			identity.SetMetadata(key, value)
		}
	}
	
	// Persistir identidade
	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, fmt.Errorf("erro ao persistir identidade: %w", err)
	}
	
	return identity, nil
}

// GetIdentity recupera uma identidade por ID
func (s *IdentityService) GetIdentity(ctx context.Context, identityID uuid.UUID, loadContexts bool) (*models.Identity, error) {
	identity, err := s.identityRepo.GetByID(ctx, identityID)
	if err != nil {
		return nil, err
	}
	
	if loadContexts {
		if err := s.identityRepo.LoadContexts(ctx, identity); err != nil {
			return nil, fmt.Errorf("erro ao carregar contextos: %w", err)
		}
		
		// Carregar atributos para cada contexto
		for _, context := range identity.Contexts {
			if err := s.contextRepo.LoadAttributes(ctx, context); err != nil {
				return nil, fmt.Errorf("erro ao carregar atributos para contexto %s: %w", context.ID, err)
			}
		}
	}
	
	return identity, nil
}

// GetIdentityByPrimaryKey recupera uma identidade por chave primária
func (s *IdentityService) GetIdentityByPrimaryKey(
	ctx context.Context, 
	keyType models.PrimaryKeyType, 
	keyValue string,
	loadContexts bool,
) (*models.Identity, error) {
	identity, err := s.identityRepo.GetByPrimaryKey(ctx, keyType, keyValue)
	if err != nil {
		return nil, err
	}
	
	if loadContexts {
		if err := s.identityRepo.LoadContexts(ctx, identity); err != nil {
			return nil, fmt.Errorf("erro ao carregar contextos: %w", err)
		}
		
		// Carregar atributos para cada contexto
		for _, context := range identity.Contexts {
			if err := s.contextRepo.LoadAttributes(ctx, context); err != nil {
				return nil, fmt.Errorf("erro ao carregar atributos para contexto %s: %w", context.ID, err)
			}
		}
	}
	
	return identity, nil
}

// AddContext adiciona um novo contexto a uma identidade
func (s *IdentityService) AddContext(
	ctx context.Context, 
	identityID uuid.UUID,
	contextType models.ContextType,
) (*models.IdentityContext, error) {
	// Verificar se identidade existe
	identity, err := s.identityRepo.GetByID(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("identidade não encontrada: %w", err)
	}
	
	// Verificar se contexto já existe
	existing, err := s.contextRepo.GetByIdentityAndType(ctx, identityID, contextType)
	if err == nil && existing != nil {
		return nil, models.ErrDuplicateContext
	}
	
	// Criar novo contexto
	identityContext, err := models.NewIdentityContext(identityID, contextType)
	if err != nil {
		return nil, err
	}
	
	// Persistir contexto
	if err := s.contextRepo.Create(ctx, identityContext); err != nil {
		return nil, fmt.Errorf("erro ao persistir contexto: %w", err)
	}
	
	// Adicionar contexto à identidade na memória
	identity.AddContext(identityContext)
	
	// Iniciar avaliação de confiança em background se TrustGuard estiver disponível
	if s.trustGuardClient != nil {
		go func() {
			bgCtx := context.Background()
			trustScore, err := s.trustGuardClient.EvaluateIdentity(
				bgCtx, 
				identity.ID.String(), 
				string(contextType),
			)
			if err == nil {
				// Atualizar pontuação de confiança no contexto
				s.contextRepo.UpdateTrustScore(bgCtx, identityContext.ID, trustScore)
			}
		}()
	}
	
	return identityContext, nil
}

// AddContextAttribute adiciona um atributo a um contexto
func (s *IdentityService) AddContextAttribute(
	ctx context.Context,
	contextID uuid.UUID,
	key string,
	value string,
	sensitivityLevel models.SensitivityLevel,
	verifyAttribute bool,
) (*models.ContextAttribute, error) {
	// Verificar se contexto existe
	identityContext, err := s.contextRepo.GetByID(ctx, contextID)
	if err != nil {
		return nil, fmt.Errorf("contexto não encontrado: %w", err)
	}
	
	// Verificar se atributo já existe
	existingAttr, err := s.attributeRepo.GetByContextAndKey(ctx, contextID, key)
	if err == nil && existingAttr != nil {
		// Atualizar valor do atributo existente
		existingAttr.UpdateValue(value)
		if err := s.attributeRepo.Update(ctx, existingAttr); err != nil {
			return nil, fmt.Errorf("erro ao atualizar atributo: %w", err)
		}
		
		if verifyAttribute && s.trustGuardClient != nil {
			// Verificar atributo com TrustGuard
			go s.verifyAttributeWithTrustGuard(existingAttr, identityContext.ContextType)
		}
		
		return existingAttr, nil
	}
	
	// Criar novo atributo
	attribute, err := models.NewContextAttribute(
		contextID,
		key,
		value,
		sensitivityLevel,
	)
	if err != nil {
		return nil, err
	}
	
	// Persistir atributo
	if err := s.attributeRepo.Create(ctx, attribute); err != nil {
		return nil, fmt.Errorf("erro ao persistir atributo: %w", err)
	}
	
	// Adicionar atributo ao contexto na memória
	identityContext.AddAttribute(attribute)
	
	// Iniciar verificação do atributo em background
	if verifyAttribute && s.trustGuardClient != nil {
		go s.verifyAttributeWithTrustGuard(attribute, identityContext.ContextType)
	}
	
	return attribute, nil
}

// verifyAttributeWithTrustGuard verifica um atributo usando o TrustGuard
func (s *IdentityService) verifyAttributeWithTrustGuard(
	attribute *models.ContextAttribute,
	contextType models.ContextType,
) {
	bgCtx := context.Background()
	isValid, source, err := s.trustGuardClient.VerifyAttribute(
		bgCtx,
		attribute.AttributeKey,
		attribute.AttributeValue,
		string(contextType),
	)
	
	if err != nil {
		return
	}
	
	var status models.VerificationStatus
	if isValid {
		status = models.VerificationStatusVerified
	} else {
		status = models.VerificationStatusFailed
	}
	
	// Atualizar status de verificação
	s.attributeRepo.UpdateVerification(bgCtx, attribute.ID, status, source)
}

// UpdateIdentityStatus atualiza o status de uma identidade
func (s *IdentityService) UpdateIdentityStatus(
	ctx context.Context, 
	identityID uuid.UUID, 
	status models.IdentityStatus,
) error {
	identity, err := s.identityRepo.GetByID(ctx, identityID)
	if err != nil {
		return err
	}
	
	identity.UpdateStatus(status)
	return s.identityRepo.Update(ctx, identity)
}

// ListIdentitiesByFilter lista identidades com base em filtros
func (s *IdentityService) ListIdentitiesByFilter(
	ctx context.Context,
	filter repositories.IdentityFilter,
	page, pageSize int,
) ([]*models.Identity, int, error) {
	return s.identityRepo.List(ctx, filter, page, pageSize)
}