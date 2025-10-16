/**
 * @file context_service.go
 * @description Serviço de domínio para gerenciamento de contextos de identidade
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/repositories"
)

// ContextMappingProvider define interface para mapeamento entre contextos
type ContextMappingProvider interface {
	// MapAttributes mapeia atributos entre diferentes contextos
	MapAttributes(ctx context.Context, sourceContextType, targetContextType models.ContextType, attributes []*models.ContextAttribute) ([]*models.ContextAttribute, error)
	
	// GetCompatibleContexts retorna tipos de contextos compatíveis com um determinado contexto
	GetCompatibleContexts(contextType models.ContextType) []models.ContextType
	
	// GetMappingPolicies obtém políticas de mapeamento entre dois contextos
	GetMappingPolicies(sourceType, targetType models.ContextType) (map[string]string, error)
}

// ContextService encapsula a lógica de domínio para contextos de identidade
type ContextService struct {
	contextRepo     repositories.ContextRepository
	attributeRepo   repositories.AttributeRepository
	identityRepo    repositories.IdentityRepository
	mappingProvider ContextMappingProvider
	trustGuardClient TrustGuardClient
}

// NewContextService cria uma nova instância do serviço de contexto
func NewContextService(
	contextRepo repositories.ContextRepository,
	attributeRepo repositories.AttributeRepository,
	identityRepo repositories.IdentityRepository,
	mappingProvider ContextMappingProvider,
	trustGuardClient TrustGuardClient,
) *ContextService {
	return &ContextService{
		contextRepo:     contextRepo,
		attributeRepo:   attributeRepo,
		identityRepo:    identityRepo,
		mappingProvider: mappingProvider,
		trustGuardClient: trustGuardClient,
	}
}

// GetContext obtém um contexto específico com seus atributos
func (s *ContextService) GetContext(ctx context.Context, contextID uuid.UUID) (*models.IdentityContext, error) {
	identityContext, err := s.contextRepo.GetByID(ctx, contextID)
	if err != nil {
		return nil, err
	}
	
	// Carregar atributos
	if err := s.contextRepo.LoadAttributes(ctx, identityContext); err != nil {
		return nil, fmt.Errorf("erro ao carregar atributos para contexto %s: %w", contextID, err)
	}
	
	return identityContext, nil
}

// UpdateContextStatus atualiza o status de um contexto
func (s *ContextService) UpdateContextStatus(ctx context.Context, contextID uuid.UUID, status models.ContextStatus) error {
	identityContext, err := s.contextRepo.GetByID(ctx, contextID)
	if err != nil {
		return err
	}
	
	identityContext.UpdateStatus(status)
	return s.contextRepo.Update(ctx, identityContext)
}

// UpdateVerificationLevel atualiza o nível de verificação de um contexto
func (s *ContextService) UpdateVerificationLevel(ctx context.Context, contextID uuid.UUID, level models.VerificationLevel) error {
	identityContext, err := s.contextRepo.GetByID(ctx, contextID)
	if err != nil {
		return err
	}
	
	identityContext.UpdateVerificationLevel(level)
	return s.contextRepo.Update(ctx, identityContext)
}

// MapContexts cria um novo contexto para uma identidade baseado em um contexto existente
func (s *ContextService) MapContexts(
	ctx context.Context, 
	identityID uuid.UUID,
	sourceContextType models.ContextType,
	targetContextType models.ContextType,
) (*models.IdentityContext, error) {
	// Verificar se mapeador está disponível
	if s.mappingProvider == nil {
		return nil, fmt.Errorf("provedor de mapeamento não disponível")
	}
	
	// Verificar se os tipos de contexto são compatíveis
	compatibleContexts := s.mappingProvider.GetCompatibleContexts(sourceContextType)
	isCompatible := false
	for _, compatibleType := range compatibleContexts {
		if compatibleType == targetContextType {
			isCompatible = true
			break
		}
	}
	
	if !isCompatible {
		return nil, fmt.Errorf("tipos de contexto incompatíveis: %s -> %s", sourceContextType, targetContextType)
	}
	
	// Obter contexto de origem
	sourceContext, err := s.contextRepo.GetByIdentityAndType(ctx, identityID, sourceContextType)
	if err != nil {
		return nil, fmt.Errorf("contexto de origem não encontrado: %w", err)
	}
	
	// Carregar atributos do contexto de origem
	if err := s.contextRepo.LoadAttributes(ctx, sourceContext); err != nil {
		return nil, fmt.Errorf("erro ao carregar atributos do contexto de origem: %w", err)
	}
	
	// Verificar se contexto de destino já existe
	existingTargetContext, err := s.contextRepo.GetByIdentityAndType(ctx, identityID, targetContextType)
	if err == nil && existingTargetContext != nil {
		return nil, models.ErrDuplicateContext
	}
	
	// Criar novo contexto de destino
	targetContext, err := models.NewIdentityContext(identityID, targetContextType)
	if err != nil {
		return nil, err
	}
	
	// Persistir contexto de destino
	if err := s.contextRepo.Create(ctx, targetContext); err != nil {
		return nil, fmt.Errorf("erro ao persistir contexto de destino: %w", err)
	}
	
	// Mapear atributos entre contextos
	mappedAttributes, err := s.mappingProvider.MapAttributes(
		ctx,
		sourceContextType,
		targetContextType,
		sourceContext.Attributes,
	)
	if err != nil {
		// Não falhar completamente se o mapeamento tiver problemas parciais
		// Apenas registrar erro e continuar com os atributos que foram mapeados
		fmt.Printf("Aviso: Mapeamento parcial de atributos: %v\n", err)
	}
	
	// Persistir atributos mapeados
	if len(mappedAttributes) > 0 {
		for _, attr := range mappedAttributes {
			attr.ContextID = targetContext.ID
			if err := s.attributeRepo.Create(ctx, attr); err != nil {
				fmt.Printf("Erro ao persistir atributo mapeado %s: %v\n", attr.AttributeKey, err)
			} else {
				targetContext.AddAttribute(attr)
			}
		}
	}
	
	// Avaliar confiança do novo contexto
	if s.trustGuardClient != nil {
		go func() {
			bgCtx := context.Background()
			trustScore, err := s.trustGuardClient.EvaluateIdentity(
				bgCtx, 
				identityID.String(), 
				string(targetContextType),
			)
			if err == nil {
				// Atualizar pontuação de confiança no contexto
				s.contextRepo.UpdateTrustScore(bgCtx, targetContext.ID, trustScore)
			}
		}()
	}
	
	return targetContext, nil
}

// VerifyAllAttributesInContext verifica todos os atributos em um contexto
func (s *ContextService) VerifyAllAttributesInContext(ctx context.Context, contextID uuid.UUID) (int, error) {
	// Verificar se contexto existe
	identityContext, err := s.contextRepo.GetByID(ctx, contextID)
	if err != nil {
		return 0, fmt.Errorf("contexto não encontrado: %w", err)
	}
	
	// Carregar atributos
	if err := s.contextRepo.LoadAttributes(ctx, identityContext); err != nil {
		return 0, fmt.Errorf("erro ao carregar atributos para contexto %s: %w", contextID, err)
	}
	
	// Verificar se TrustGuard está disponível
	if s.trustGuardClient == nil {
		return 0, fmt.Errorf("cliente TrustGuard não disponível para verificação")
	}
	
	// Contador de atributos verificados com sucesso
	verifiedCount := 0
	
	// Verificar cada atributo
	for _, attr := range identityContext.Attributes {
		// Ignorar atributos já verificados
		if attr.VerificationStatus == models.VerificationStatusVerified {
			verifiedCount++
			continue
		}
		
		// Verificar atributo
		isValid, source, err := s.trustGuardClient.VerifyAttribute(
			ctx,
			attr.AttributeKey,
			attr.AttributeValue,
			string(identityContext.ContextType),
		)
		
		if err != nil {
			// Registrar erro mas continuar com outros atributos
			fmt.Printf("Erro ao verificar atributo %s: %v\n", attr.AttributeKey, err)
			continue
		}
		
		var status models.VerificationStatus
		if isValid {
			status = models.VerificationStatusVerified
			verifiedCount++
		} else {
			status = models.VerificationStatusFailed
		}
		
		// Atualizar status de verificação
		if err := s.attributeRepo.UpdateVerification(ctx, attr.ID, status, source); err != nil {
			fmt.Printf("Erro ao atualizar verificação do atributo %s: %v\n", attr.AttributeKey, err)
		} else {
			// Atualizar em memória
			attr.UpdateVerification(status, source)
		}
	}
	
	// Atualizar nível de verificação do contexto com base nos atributos verificados
	s.updateContextVerificationLevel(ctx, identityContext, verifiedCount)
	
	return verifiedCount, nil
}

// updateContextVerificationLevel atualiza o nível de verificação do contexto
// com base nos atributos verificados
func (s *ContextService) updateContextVerificationLevel(
	ctx context.Context, 
	identityContext *models.IdentityContext, 
	verifiedCount int,
) {
	totalAttributes := len(identityContext.Attributes)
	if totalAttributes == 0 {
		return
	}
	
	verificationRatio := float64(verifiedCount) / float64(totalAttributes)
	
	var newLevel models.VerificationLevel
	
	switch {
	case verificationRatio >= 0.9:
		newLevel = models.VerificationComplete
	case verificationRatio >= 0.7:
		newLevel = models.VerificationEnhanced
	case verificationRatio >= 0.5:
		newLevel = models.VerificationStandard
	case verificationRatio > 0:
		newLevel = models.VerificationBasic
	default:
		newLevel = models.VerificationNone
	}
	
	// Atualizar nível de verificação se mudou
	if identityContext.VerificationLevel != newLevel {
		identityContext.UpdateVerificationLevel(newLevel)
		s.contextRepo.Update(ctx, identityContext)
	}
}

// ListContextsByIdentity lista todos os contextos de uma identidade
func (s *ContextService) ListContextsByIdentity(ctx context.Context, identityID uuid.UUID) ([]*models.IdentityContext, error) {
	// Verificar se identidade existe
	_, err := s.identityRepo.GetByID(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("identidade não encontrada: %w", err)
	}
	
	return s.contextRepo.ListByIdentity(ctx, identityID)
}

// ListContextsByFilter lista contextos com base em filtros
func (s *ContextService) ListContextsByFilter(
	ctx context.Context,
	filter repositories.ContextFilter,
	page, pageSize int,
) ([]*models.IdentityContext, int, error) {
	return s.contextRepo.List(ctx, filter, page, pageSize)
}

// SearchAttributesByValue busca atributos por valor em todos os contextos
func (s *ContextService) SearchAttributesByValue(
	ctx context.Context, 
	searchValue string,
	sensitivityLevels []models.SensitivityLevel,
) ([]*models.ContextAttribute, error) {
	// Limitar busca a níveis de sensibilidade permitidos
	if len(sensitivityLevels) == 0 {
		sensitivityLevels = []models.SensitivityLevel{
			models.SensitivityPublic,
			models.SensitivityLow,
		}
	}
	
	return s.attributeRepo.SearchAttributes(ctx, searchValue, sensitivityLevels)
}