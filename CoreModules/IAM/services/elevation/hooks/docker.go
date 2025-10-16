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

// DockerHook implementa a interface MCPHook para Docker
type DockerHook struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *config.DockerElevationConfig
}

// DockerElevationConfig contém configurações específicas para o hook Docker
type DockerElevationConfig struct {
	// Mapeamento de comandos para escopos necessários
	CommandScopeMap map[string]string
	
	// Configurações específicas por mercado
	MarketConfigs map[string]*MarketDockerConfig
	
	// Valores padrão quando mercado específico não configurado
	DefaultConfig *MarketDockerConfig
}

// MarketDockerConfig contém configurações específicas por mercado
type MarketDockerConfig struct {
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
	
	// SensitiveCommands requer aprovação adicional
	SensitiveCommands []string
	
	// ForbiddenCommands são comandos proibidos
	ForbiddenCommands []string
	
	// MFARequiredCommands exigem MFA independente da configuração padrão
	MFARequiredCommands []string
}

// NewDockerHook cria uma nova instância do hook Docker
func NewDockerHook(config *config.DockerElevationConfig) *DockerHook {
	return &DockerHook{
		logger: logging.GetLogger().Named("docker-hook"),
		tracer: otel.Tracer("innovabiz/iam/elevation/hooks/docker"),
		config: config,
	}
}

// HookType retorna o tipo do hook
func (h *DockerHook) HookType() MCPHookType {
	return Docker
}

// ValidateScope valida um escopo para Docker
func (h *DockerHook) ValidateScope(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (*ScopeDetails, error) {
	ctx, span := h.tracer.Start(ctx, "DockerHook.ValidateScope",
		trace.WithAttributes(
			attribute.String("scope", scope),
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	// Verifica se o escopo começa com o prefixo docker:
	if !strings.HasPrefix(scope, "docker:") {
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: Docker,
			Reason:   "escopo deve começar com 'docker:'",
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
	case ScopeDockerAdmin:
		details.DisplayName = "Administração Docker"
		details.Description = "Acesso completo a todas operações Docker"
		details.SensitivityLevel = 5
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	case ScopeDockerExec:
		details.DisplayName = "Execução de Comandos"
		details.Description = "Execução de comandos em containers Docker"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 60 // 1 hora

	case ScopeDockerPush:
		details.DisplayName = "Push de Imagens"
		details.Description = "Envio de imagens para registries"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 120 // 2 horas

	case ScopeDockerPull:
		details.DisplayName = "Pull de Imagens"
		details.Description = "Download de imagens de registries"
		details.SensitivityLevel = 2
		details.RequiresApproval = false
		details.RequiredMFA = elevation.MFANone
		details.AllowedInEmergency = true
		details.MaxDuration = 240 // 4 horas

	case ScopeDockerBuild:
		details.DisplayName = "Build de Imagens"
		details.Description = "Construção de imagens Docker"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFANone
		details.AllowedInEmergency = true
		details.MaxDuration = 180 // 3 horas

	case ScopeDockerVolume:
		details.DisplayName = "Gerenciamento de Volumes"
		details.Description = "Gerenciamento de volumes Docker"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 60 // 1 hora

	case ScopeDockerNetwork:
		details.DisplayName = "Gerenciamento de Rede"
		details.Description = "Gerenciamento de redes Docker"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 60 // 1 hora

	case ScopeDockerSystem:
		details.DisplayName = "Operações de Sistema"
		details.Description = "Operações a nível de sistema Docker"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	default:
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: Docker,
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
func (h *DockerHook) GetRequiredMFA(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (elevation.MFALevel, error) {
	ctx, span := h.tracer.Start(ctx, "DockerHook.GetRequiredMFA")
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
func (h *DockerHook) GetRequireApproval(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (bool, error) {
	ctx, span := h.tracer.Start(ctx, "DockerHook.GetRequireApproval")
	defer span.End()

	details, err := h.ValidateScope(ctx, scope, tenantID, market)
	if err != nil {
		return false, err
	}

	return details.RequiresApproval, nil
}

// ValidateRequest valida uma solicitação de elevação
func (h *DockerHook) ValidateRequest(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) error {
	ctx, span := h.tracer.Start(ctx, "DockerHook.ValidateRequest")
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
		_, err := h.ValidateScope(ctx, scope, request.TenantID, request.Market)
		if err != nil {
			return err
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
				HookType:      Docker,
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

	return nil
}

// GetApprovers retorna a lista de usuários que podem aprovar a solicitação
func (h *DockerHook) GetApprovers(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) ([]string, error) {
	// Nota: Na implementação real, isso consultaria o serviço de tenant 
	// para obter aprovadores específicos baseados em papéis e políticas
	ctx, span := h.tracer.Start(ctx, "DockerHook.GetApprovers")
	defer span.End()

	// Simulação para testes
	return []string{"admin1", "admin2", "securityOfficer1"}, nil
}

// ValidateElevationUse valida o uso de um token
func (h *DockerHook) ValidateElevationUse(
	ctx context.Context, 
	tokenID string, 
	scope string, 
	metadata map[string]interface{},
) error {
	ctx, span := h.tracer.Start(ctx, "DockerHook.ValidateElevationUse",
		trace.WithAttributes(
			attribute.String("token_id", tokenID),
			attribute.String("scope", scope),
		))
	defer span.End()

	// Validação específica para comandos Docker
	if command, ok := metadata["command"].(string); ok {
		// Verifica comandos proibidos globalmente
		if h.isCommandForbidden(command, metadata["market"].(string)) {
			return fmt.Errorf("comando Docker '%s' é proibido por política de segurança", command)
		}

		// Verifica escopo necessário para o comando
		requiredScope, err := h.getScopeForCommand(command)
		if err != nil {
			h.logger.Warn("Comando sem mapeamento de escopo definido",
				zap.String("command", command),
				zap.String("token_id", tokenID))
		} else if requiredScope != scope && requiredScope != ScopeDockerAdmin {
			return fmt.Errorf("escopo %s insuficiente para comando %s, requer %s", 
				scope, command, requiredScope)
		}
	}

	h.logger.Info("Uso de token Docker validado com sucesso", 
		zap.String("token_id", tokenID),
		zap.String("scope", scope))

	return nil
}

// GetPolicyLimits retorna os limites da política
func (h *DockerHook) GetPolicyLimits(
	ctx context.Context, 
	tenantID string, 
	market string,
) (*PolicyLimits, error) {
	ctx, span := h.tracer.Start(ctx, "DockerHook.GetPolicyLimits",
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
		},
		AllowedScopes: []string{
			ScopeDockerAdmin,
			ScopeDockerExec,
			ScopeDockerPush,
			ScopeDockerPull,
			ScopeDockerBuild,
			ScopeDockerVolume,
			ScopeDockerNetwork,
			ScopeDockerSystem,
		},
		RoleBasedRestrictions: map[string]interface{}{
			"developer": []string{
				ScopeDockerPull,
				ScopeDockerBuild,
				ScopeDockerExec,
			},
			"operator": []string{
				ScopeDockerPull,
				ScopeDockerExec,
				ScopeDockerVolume,
				ScopeDockerNetwork,
			},
			"admin": []string{
				ScopeDockerAdmin,
				ScopeDockerSystem,
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
func (h *DockerHook) GetAuditMetadata(
	ctx context.Context, 
	tokenID string, 
	scope string,
) (map[string]interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "DockerHook.GetAuditMetadata")
	defer span.End()

	return map[string]interface{}{
		"hook_type":     "docker",
		"scope":         scope,
		"timestamp":     time.Now().Format(time.RFC3339),
		"token_id":      tokenID,
	}, nil
}

// Métodos auxiliares

// getMarketConfig retorna configuração específica do mercado ou padrão
func (h *DockerHook) getMarketConfig(market string) *MarketDockerConfig {
	if h.config == nil || h.config.MarketConfigs == nil {
		// Configuração padrão em caso de erro
		return &MarketDockerConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     60,
			RequireApproval:        true,
			RequireMFA:             true,
			DefaultMFALevel:        elevation.MFABasic,
			AllowEmergencyMode:     true,
		}
	}

	config, exists := h.config.MarketConfigs[market]
	if !exists {
		if h.config.DefaultConfig != nil {
			return h.config.DefaultConfig
		}
		
		// Configuração padrão em caso de erro
		return &MarketDockerConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     60,
			RequireApproval:        true,
			RequireMFA:             true,
			DefaultMFALevel:        elevation.MFABasic,
			AllowEmergencyMode:     true,
		}
	}
	
	return config
}

// isCommandForbidden verifica se um comando é proibido
func (h *DockerHook) isCommandForbidden(command, market string) bool {
	config := h.getMarketConfig(market)
	
	// Verifica comandos proibidos
	for _, forbidden := range config.ForbiddenCommands {
		if strings.Contains(command, forbidden) {
			return true
		}
	}
	
	return false
}

// getScopeForCommand retorna o escopo necessário para um comando
func (h *DockerHook) getScopeForCommand(command string) (string, error) {
	if h.config == nil || h.config.CommandScopeMap == nil {
		// Mapeamento padrão
		commandMap := map[string]string{
			"exec":    ScopeDockerExec,
			"run":     ScopeDockerExec,
			"push":    ScopeDockerPush,
			"pull":    ScopeDockerPull,
			"build":   ScopeDockerBuild,
			"volume":  ScopeDockerVolume,
			"network": ScopeDockerNetwork,
			"system":  ScopeDockerSystem,
		}
		
		for cmd, scope := range commandMap {
			if strings.HasPrefix(command, "docker "+cmd) {
				return scope, nil
			}
		}
		
		return "", fmt.Errorf("comando sem mapeamento de escopo")
	}
	
	for cmd, scope := range h.config.CommandScopeMap {
		if strings.HasPrefix(command, cmd) {
			return scope, nil
		}
	}
	
	return "", fmt.Errorf("comando sem mapeamento de escopo")
}