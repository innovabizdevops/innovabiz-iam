// Package hooks implementa a integração entre o serviço de elevação
// e os diferentes hooks MCP (Docker, Desktop Commander, GitHub, Figma)
// da plataforma INNOVABIZ IAM.
package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/innovabiz/iam/services/elevation"
	"github.com/innovabiz/iam/config"
	"github.com/innovabiz/iam/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GitHubHook implementa a interface MCPHook para GitHub
type GitHubHook struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *config.GitHubElevationConfig
}

// GitHubElevationConfig contém configurações específicas para o hook GitHub
type GitHubElevationConfig struct {
	// Mapeamento de operações para escopos necessários
	OperationScopeMap map[string]string
	
	// Configurações específicas por mercado
	MarketConfigs map[string]*MarketGitHubConfig
	
	// Valores padrão quando mercado específico não configurado
	DefaultConfig *MarketGitHubConfig
}

// MarketGitHubConfig contém configurações específicas por mercado
type MarketGitHubConfig struct {
	// MaxActiveTokensPerUser é o número máximo de tokens ativos por usuário
	MaxActiveTokensPerUser int
	
	// DefaultMaxDuration é a duração máxima em minutos
	DefaultMaxDuration int
	
	// RequireApproval indica se aprovação é necessária por padrão
	RequireApproval bool
	
	// RequireMFA indica se MFA é necessário por padrão
	RequireMFA bool
	
	// DefaultMFALevel é o nível de MFA padrão
	DefaultMFALevel elevation.MFALevel
	
	// AllowEmergencyMode permite modo de emergência (auto-aprovação)
	AllowEmergencyMode bool
	
	// SensitiveOperations requer aprovação adicional
	SensitiveOperations []string
	
	// ForbiddenOperations são operações proibidas
	ForbiddenOperations []string
	
	// ProtectedBranches requerem aprovação adicional
	ProtectedBranches []string
	
	// SensitiveRepositories requerem aprovação adicional
	SensitiveRepositories []string
}

// NewGitHubHook cria uma nova instância do hook GitHub
func NewGitHubHook(config *config.GitHubElevationConfig) *GitHubHook {
	return &GitHubHook{
		logger: logging.GetLogger().Named("github-hook"),
		tracer: otel.Tracer("innovabiz/iam/elevation/hooks/github"),
		config: config,
	}
}

// HookType retorna o tipo do hook
func (h *GitHubHook) HookType() MCPHookType {
	return GitHub
}

// ValidateScope valida um escopo para GitHub
func (h *GitHubHook) ValidateScope(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (*ScopeDetails, error) {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.ValidateScope",
		trace.WithAttributes(
			attribute.String("scope", scope),
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	// Verifica se o escopo começa com o prefixo github:
	if !strings.HasPrefix(scope, "github:") {
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: GitHub,
			Reason:   "escopo deve começar com 'github:'",
			Market:   market,
			TenantID: tenantID,
		}
	}

	// Obtém configuração específica do mercado ou usa a padrão
	marketConfig := h.getMarketConfig(market)

	// Informações de escopo baseadas no tipo
	details := &ScopeDetails{
		Scope:      scope,
		RequiredMFA: elevation.MFANone,
		Metadata:   make(map[string]interface{}),
	}

	// Configurações baseadas no tipo de escopo
	switch scope {
	case ScopeGitHubAdmin:
		details.DisplayName = "Administração GitHub"
		details.Description = "Acesso completo a todas operações GitHub"
		details.SensitivityLevel = 5
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	case ScopeGitHubPush:
		details.DisplayName = "Push de Código"
		details.Description = "Envio de código para repositórios"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 120 // 2 horas

	case ScopeGitHubMerge:
		details.DisplayName = "Merge de Pull Requests"
		details.Description = "Aprovação e merge de pull requests"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = false
		details.MaxDuration = 60 // 1 hora

	case ScopeGitHubSecrets:
		details.DisplayName = "Gerenciamento de Secrets"
		details.Description = "Gerenciamento de secrets do repositório"
		details.SensitivityLevel = 5
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	case ScopeGitHubSettings:
		details.DisplayName = "Configurações de Repositório"
		details.Description = "Alteração de configurações do repositório"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = false
		details.MaxDuration = 60 // 1 hora

	case ScopeGitHubRepo:
		details.DisplayName = "Gerenciamento de Repositório"
		details.Description = "Criação e configuração de repositórios"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 120 // 2 horas

	case ScopeGitHubHooks:
		details.DisplayName = "Gerenciamento de Hooks"
		details.Description = "Configuração de webhooks e integrações"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = false
		details.MaxDuration = 60 // 1 hora

	case ScopeGitHubActions:
		details.DisplayName = "GitHub Actions"
		details.Description = "Gerenciamento de fluxos de CI/CD"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = false
		details.MaxDuration = 60 // 1 hora

	default:
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: GitHub,
			Reason:   "escopo desconhecido ou não permitido",
			Market:   market,
			TenantID: tenantID,
		}
	}

	// Ajustes baseados no mercado específico (regras regulatórias)
	if market == "angola" || market == "mozambique" {
		// Mercados SADC/PALOP têm requisitos mais rigorosos
		if details.SensitivityLevel >= 3 {
			details.RequiresApproval = true
		}
	} else if market == "brasil" {
		// Brasil (LGPD) tem requisitos específicos
		details.RequiredMFA = elevation.MFABasic // Mínimo MFA básico para qualquer operação sensível
	}

	return details, nil
}

// GetRequiredMFA retorna o nível de MFA necessário para o escopo
func (h *GitHubHook) GetRequiredMFA(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (elevation.MFALevel, error) {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.GetRequiredMFA")
	defer span.End()

	details, err := h.ValidateScope(ctx, scope, tenantID, market)
	if err != nil {
		return elevation.MFANone, err
	}

	// Para Angola e Moçambique, garantimos pelo menos MFA básico
	if market == "angola" || market == "mozambique" {
		if details.RequiredMFA < elevation.MFABasic {
			return elevation.MFABasic, nil
		}
	}

	return details.RequiredMFA, nil
}

// GetRequireApproval determina se aprovação é necessária
func (h *GitHubHook) GetRequireApproval(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (bool, error) {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.GetRequireApproval")
	defer span.End()

	details, err := h.ValidateScope(ctx, scope, tenantID, market)
	if err != nil {
		return false, err
	}

	// Verificações adicionais específicas para mercados regulamentados
	if market == "angola" || market == "mozambique" || market == "brasil" {
		if details.SensitivityLevel >= 4 {
			return true, nil // Sempre requer aprovação para operações de alta sensibilidade
		}
	}

	return details.RequiresApproval, nil
}

// ValidateRequest valida uma solicitação de elevação
func (h *GitHubHook) ValidateRequest(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) error {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.ValidateRequest")
	defer span.End()

	marketConfig := h.getMarketConfig(request.Market)

	// Verifica justificativa
	if marketConfig.RequireApproval {
		minLength := 20
		if len(request.Justification) < minLength {
			return fmt.Errorf("justificativa deve ter pelo menos %d caracteres", minLength)
		}
	}

	// Verifica escopos solicitados
	for _, scope := range request.Scopes {
		details, err := h.ValidateScope(ctx, scope, request.TenantID, request.Market)
		if err != nil {
			return err
		}

		// Verifica aprovação necessária
		requiresApproval, err := h.GetRequireApproval(ctx, scope, request.TenantID, request.Market)
		if err != nil {
			return err
		}

		if requiresApproval && request.Emergency && !details.AllowedInEmergency {
			return fmt.Errorf("escopo %s requer aprovação e não pode ser usado em modo emergencial", scope)
		}
	}

	// Verifica nível de MFA
	for _, scope := range request.Scopes {
		requiredMFA, err := h.GetRequiredMFA(ctx, scope, request.TenantID, request.Market)
		if err != nil {
			return err
		}

		if requiredMFA > request.MFALevel {
			return &InsufficientMFAError{
				Scope:         scope,
				HookType:      GitHub,
				RequiredLevel: requiredMFA,
				ProvidedLevel: request.MFALevel,
			}
		}
	}

	// Verifica modo de emergência
	if request.Emergency {
		if !marketConfig.AllowEmergencyMode {
			return fmt.Errorf("modo de emergência não permitido para o mercado %s", request.Market)
		}

		// Verifica escopos não permitidos em emergência
		for _, scope := range request.Scopes {
			details, err := h.ValidateScope(ctx, scope, request.TenantID, request.Market)
			if err != nil {
				return err
			}

			if !details.AllowedInEmergency {
				return fmt.Errorf("escopo %s não permitido em modo de emergência", scope)
			}
		}
	}

	// Verificações específicas para metadados do GitHub
	if metadata, ok := request.Metadata.(map[string]interface{}); ok {
		// Verifica operação em branches protegidos
		if branch, ok := metadata["branch"].(string); ok {
			if h.isProtectedBranch(branch, request.Market) && !request.Emergency {
				return fmt.Errorf("operação em branch protegido '%s' requer aprovação", branch)
			}
		}

		// Verifica repositórios sensíveis
		if repo, ok := metadata["repository"].(string); ok {
			if h.isSensitiveRepository(repo, request.Market) {
				for _, scope := range request.Scopes {
					details, err := h.ValidateScope(ctx, scope, request.TenantID, request.Market)
					if err != nil {
						return err
					}

					// Aumenta requisitos para repositórios sensíveis
					if details.SensitivityLevel < 4 {
						h.logger.Warn("Aumentando nível de sensibilidade para repositório crítico",
							zap.String("repo", repo),
							zap.String("scope", scope))
						
						if request.Emergency {
							return fmt.Errorf("modo de emergência não permitido para repositório sensível %s", repo)
						}

						if request.MFALevel < elevation.MFABasic {
							return fmt.Errorf("repositório sensível %s requer pelo menos MFA básico", repo)
						}
					}
				}
			}
		}
	}

	return nil
}

// GetApprovers retorna a lista de usuários que podem aprovar a solicitação
func (h *GitHubHook) GetApprovers(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) ([]string, error) {
	// Nota: Na implementação real, isso consultaria o serviço de tenant 
	// para obter aprovadores específicos baseados em papéis e políticas
	ctx, span := h.tracer.Start(ctx, "GitHubHook.GetApprovers")
	defer span.End()

	// Detecta se algum escopo é de alta sensibilidade
	highSensitivity := false
	for _, scope := range request.Scopes {
		details, err := h.ValidateScope(ctx, scope, request.TenantID, request.Market)
		if err != nil {
			return nil, err
		}

		if details.SensitivityLevel >= 4 {
			highSensitivity = true
			break
		}
	}

	// Se é alta sensibilidade, inclui aprovadores de segurança
	if highSensitivity {
		return []string{"securityOfficer1", "securityOfficer2", "admin1"}, nil
	}

	// Operações normais
	return []string{"teamLead1", "teamLead2", "admin1", "admin2"}, nil
}

// ValidateElevationUse valida o uso de um token
func (h *GitHubHook) ValidateElevationUse(
	ctx context.Context, 
	tokenID string, 
	scope string, 
	metadata map[string]interface{},
) error {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.ValidateElevationUse",
		trace.WithAttributes(
			attribute.String("token_id", tokenID),
			attribute.String("scope", scope),
		))
	defer span.End()

	// Validação específica para operações GitHub
	if operation, ok := metadata["operation"].(string); ok {
		// Verifica operações proibidas globalmente
		if h.isOperationForbidden(operation, metadata["market"].(string)) {
			return fmt.Errorf("operação GitHub '%s' é proibida por política de segurança", operation)
		}

		// Verifica escopo necessário para a operação
		requiredScope, err := h.getScopeForOperation(operation)
		if err != nil {
			h.logger.Warn("Operação sem mapeamento de escopo definido",
				zap.String("operation", operation),
				zap.String("token_id", tokenID))
		} else if requiredScope != scope && scope != ScopeGitHubAdmin {
			return fmt.Errorf("escopo %s insuficiente para operação %s, requer %s", 
				scope, operation, requiredScope)
		}

		// Verificações adicionais para branches protegidos
		if branch, ok := metadata["branch"].(string); ok && h.isProtectedBranch(branch, metadata["market"].(string)) {
			if scope != ScopeGitHubAdmin && scope != ScopeGitHubMerge {
				return fmt.Errorf("operação em branch protegido '%s' requer escopo admin ou merge", branch)
			}
		}

		// Verificações para repositórios sensíveis
		if repo, ok := metadata["repository"].(string); ok && h.isSensitiveRepository(repo, metadata["market"].(string)) {
			if scope != ScopeGitHubAdmin && scope != ScopeGitHubSecrets {
				h.logger.Warn("Operação em repositório sensível com escopo limitado",
					zap.String("repository", repo),
					zap.String("scope", scope),
					zap.String("operation", operation))
				
				// Podemos adicionar restrições adicionais aqui se necessário
			}
		}
	}

	h.logger.Info("Uso de token GitHub validado com sucesso", 
		zap.String("token_id", tokenID),
		zap.String("scope", scope))

	return nil
}

// GetPolicyLimits retorna os limites da política
func (h *GitHubHook) GetPolicyLimits(
	ctx context.Context, 
	tenantID string, 
	market string,
) (*PolicyLimits, error) {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.GetPolicyLimits",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	marketConfig := h.getMarketConfig(market)

	limits := &PolicyLimits{
		MaxActiveTokensPerUser: marketConfig.MaxActiveTokensPerUser,
		DefaultMaxDuration:     marketConfig.DefaultMaxDuration,
		AbsoluteMaxDuration:    480, // 8 horas
		RequireJustification:   true,
		MinJustificationLength: 20,
		RequireMFA:             marketConfig.RequireMFA,
		DefaultMFALevel:        marketConfig.DefaultMFALevel,
		AllowEmergencyMode:     marketConfig.AllowEmergencyMode,
		EmergencyModeRestrictions: map[string]interface{}{
			"max_duration": 60, // 1 hora em emergência
			"requires_post_justification": true,
			"excluded_scopes": []string{
				ScopeGitHubAdmin, 
				ScopeGitHubSecrets,
				ScopeGitHubSettings,
			},
		},
		AllowedScopes: []string{
			ScopeGitHubAdmin,
			ScopeGitHubPush,
			ScopeGitHubMerge,
			ScopeGitHubSecrets,
			ScopeGitHubSettings,
			ScopeGitHubRepo,
			ScopeGitHubHooks,
			ScopeGitHubActions,
		},
		RoleBasedRestrictions: map[string]interface{}{
			"developer": []string{
				ScopeGitHubPush,
				ScopeGitHubRepo,
			},
			"tech_lead": []string{
				ScopeGitHubPush,
				ScopeGitHubMerge,
				ScopeGitHubRepo,
				ScopeGitHubActions,
			},
			"devops": []string{
				ScopeGitHubActions,
				ScopeGitHubHooks,
				ScopeGitHubRepo,
			},
			"security": []string{
				ScopeGitHubSecrets,
				ScopeGitHubSettings,
			},
			"admin": []string{
				ScopeGitHubAdmin,
			},
		},
	}

	// Ajustes baseados no mercado (SADC, PALOP, BRICS, etc.)
	if market == "angola" || market == "mozambique" {
		limits.RequireMFA = true
		limits.DefaultMFALevel = elevation.MFABasic
		limits.EmergencyModeRestrictions["requires_secondary_approval"] = true
	}

	return limits, nil
}

// GetAuditMetadata retorna metadados para auditoria
func (h *GitHubHook) GetAuditMetadata(
	ctx context.Context, 
	tokenID string, 
	scope string,
) (map[string]interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "GitHubHook.GetAuditMetadata")
	defer span.End()

	return map[string]interface{}{
		"hook_type":     "github",
		"scope":         scope,
		"timestamp":     time.Now().Format(time.RFC3339),
		"token_id":      tokenID,
	}, nil
}

// Métodos auxiliares

// getMarketConfig retorna configuração específica do mercado ou padrão
func (h *GitHubHook) getMarketConfig(market string) *MarketGitHubConfig {
	if h.config == nil || h.config.MarketConfigs == nil {
		// Configuração padrão em caso de erro
		return &MarketGitHubConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     120,
			RequireApproval:        true,
			RequireMFA:             true,
			DefaultMFALevel:        elevation.MFABasic,
			AllowEmergencyMode:     true,
			ProtectedBranches:      []string{"main", "master", "production", "staging"},
		}
	}

	config, exists := h.config.MarketConfigs[market]
	if !exists {
		if h.config.DefaultConfig != nil {
			return h.config.DefaultConfig
		}
		
		// Configuração padrão em caso de erro
		return &MarketGitHubConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     120,
			RequireApproval:        true,
			RequireMFA:             true,
			DefaultMFALevel:        elevation.MFABasic,
			AllowEmergencyMode:     true,
			ProtectedBranches:      []string{"main", "master", "production", "staging"},
		}
	}
	
	return config
}

// isOperationForbidden verifica se uma operação é proibida
func (h *GitHubHook) isOperationForbidden(operation, market string) bool {
	config := h.getMarketConfig(market)
	
	// Verifica operações proibidas
	for _, forbidden := range config.ForbiddenOperations {
		if strings.Contains(operation, forbidden) {
			return true
		}
	}
	
	return false
}

// getScopeForOperation retorna o escopo necessário para uma operação
func (h *GitHubHook) getScopeForOperation(operation string) (string, error) {
	if h.config == nil || h.config.OperationScopeMap == nil {
		// Mapeamento padrão
		operationMap := map[string]string{
			"push":            ScopeGitHubPush,
			"merge_pr":        ScopeGitHubMerge,
			"update_secret":   ScopeGitHubSecrets,
			"create_secret":   ScopeGitHubSecrets,
			"delete_secret":   ScopeGitHubSecrets,
			"update_setting":  ScopeGitHubSettings,
			"create_repo":     ScopeGitHubRepo,
			"delete_repo":     ScopeGitHubAdmin,
			"create_hook":     ScopeGitHubHooks,
			"update_hook":     ScopeGitHubHooks,
			"update_action":   ScopeGitHubActions,
			"create_action":   ScopeGitHubActions,
		}
		
		for op, scope := range operationMap {
			if strings.HasPrefix(operation, op) {
				return scope, nil
			}
		}
		
		return "", fmt.Errorf("operação sem mapeamento de escopo")
	}
	
	for op, scope := range h.config.OperationScopeMap {
		if strings.HasPrefix(operation, op) {
			return scope, nil
		}
	}
	
	return "", fmt.Errorf("operação sem mapeamento de escopo")
}

// isProtectedBranch verifica se um branch é protegido
func (h *GitHubHook) isProtectedBranch(branch string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, protectedBranch := range config.ProtectedBranches {
		if branch == protectedBranch {
			return true
		}
	}
	
	return false
}

// isSensitiveRepository verifica se um repositório é sensível
func (h *GitHubHook) isSensitiveRepository(repo string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, sensitiveRepo := range config.SensitiveRepositories {
		if strings.Contains(repo, sensitiveRepo) {
			return true
		}
	}
	
	return false
}