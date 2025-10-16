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

// FigmaHook implementa a interface MCPHook para Figma
type FigmaHook struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *config.FigmaElevationConfig
}

// FigmaElevationConfig contém configurações específicas para o hook Figma
type FigmaElevationConfig struct {
	// Mapeamento de operações para escopos necessários
	OperationScopeMap map[string]string
	
	// Configurações específicas por mercado
	MarketConfigs map[string]*MarketFigmaConfig
	
	// Valores padrão quando mercado específico não configurado
	DefaultConfig *MarketFigmaConfig
}

// MarketFigmaConfig contém configurações específicas por mercado
type MarketFigmaConfig struct {
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
	
	// ProtectedProjects requerem aprovação adicional
	ProtectedProjects []string
	
	// SensitiveLibraries requerem aprovação adicional
	SensitiveLibraries []string
}

// NewFigmaHook cria uma nova instância do hook Figma
func NewFigmaHook(config *config.FigmaElevationConfig) *FigmaHook {
	return &FigmaHook{
		logger: logging.GetLogger().Named("figma-hook"),
		tracer: otel.Tracer("innovabiz/iam/elevation/hooks/figma"),
		config: config,
	}
}

// HookType retorna o tipo do hook
func (h *FigmaHook) HookType() MCPHookType {
	return Figma
}

// ValidateScope valida um escopo para Figma
func (h *FigmaHook) ValidateScope(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (*ScopeDetails, error) {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.ValidateScope",
		trace.WithAttributes(
			attribute.String("scope", scope),
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	// Verifica se o escopo começa com o prefixo figma:
	if !strings.HasPrefix(scope, "figma:") {
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: Figma,
			Reason:   "escopo deve começar com 'figma:'",
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
	case ScopeFigmaAdmin:
		details.DisplayName = "Administração Figma"
		details.Description = "Acesso completo a todas operações Figma"
		details.SensitivityLevel = 5
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	case ScopeFigmaEdit:
		details.DisplayName = "Edição de Designs"
		details.Description = "Edição e modificação de designs"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 120 // 2 horas

	case ScopeFigmaView:
		details.DisplayName = "Visualização de Designs"
		details.Description = "Visualização somente leitura de designs"
		details.SensitivityLevel = 1
		details.RequiresApproval = false
		details.RequiredMFA = elevation.MFANone
		details.AllowedInEmergency = true
		details.MaxDuration = 240 // 4 horas

	case ScopeFigmaComment:
		details.DisplayName = "Comentários em Designs"
		details.Description = "Adicionar comentários em designs"
		details.SensitivityLevel = 2
		details.RequiresApproval = false
		details.RequiredMFA = elevation.MFANone
		details.AllowedInEmergency = true
		details.MaxDuration = 180 // 3 horas

	case ScopeFigmaExport:
		details.DisplayName = "Exportação de Designs"
		details.Description = "Exportar designs para vários formatos"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 60 // 1 hora

	case ScopeFigmaLibrary:
		details.DisplayName = "Gerenciamento de Bibliotecas"
		details.Description = "Gerenciar bibliotecas de componentes"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = false
		details.MaxDuration = 60 // 1 hora

	case ScopeFigmaTeam:
		details.DisplayName = "Gerenciamento de Equipes"
		details.Description = "Gerenciar membros e permissões de equipes"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	default:
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: Figma,
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
		// Brasil (LGPD) tem requisitos específicos para dados pessoais
		if details.SensitivityLevel >= 3 {
			details.RequiredMFA = elevation.MFABasic
		}
	}

	return details, nil
}

// GetRequiredMFA retorna o nível de MFA necessário para o escopo
func (h *FigmaHook) GetRequiredMFA(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (elevation.MFALevel, error) {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.GetRequiredMFA")
	defer span.End()

	details, err := h.ValidateScope(ctx, scope, tenantID, market)
	if err != nil {
		return elevation.MFANone, err
	}

	// Para Angola e Moçambique, garantimos pelo menos MFA básico para operações sensíveis
	if market == "angola" || market == "mozambique" {
		if details.SensitivityLevel >= 3 && details.RequiredMFA < elevation.MFABasic {
			return elevation.MFABasic, nil
		}
	}

	return details.RequiredMFA, nil
}

// GetRequireApproval determina se aprovação é necessária
func (h *FigmaHook) GetRequireApproval(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (bool, error) {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.GetRequireApproval")
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
func (h *FigmaHook) ValidateRequest(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) error {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.ValidateRequest")
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
				HookType:      Figma,
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

	// Verificações específicas para metadados do Figma
	if metadata, ok := request.Metadata.(map[string]interface{}); ok {
		// Verifica projetos protegidos
		if projectID, ok := metadata["project_id"].(string); ok {
			if h.isProtectedProject(projectID, request.Market) && !request.Emergency {
				return fmt.Errorf("operação em projeto protegido '%s' requer aprovação", projectID)
			}
		}

		// Verifica bibliotecas sensíveis
		if libraryID, ok := metadata["library_id"].(string); ok {
			if h.isSensitiveLibrary(libraryID, request.Market) {
				for _, scope := range request.Scopes {
					if scope != ScopeFigmaAdmin && scope != ScopeFigmaLibrary {
						return fmt.Errorf("operações em biblioteca sensível '%s' requerem escopo apropriado", libraryID)
					}
				}

				if request.Emergency {
					return fmt.Errorf("modo de emergência não permitido para biblioteca sensível %s", libraryID)
				}

				if request.MFALevel < elevation.MFABasic {
					return fmt.Errorf("biblioteca sensível %s requer pelo menos MFA básico", libraryID)
				}
			}
		}

		// Verifica operações sensíveis
		if operation, ok := metadata["operation"].(string); ok {
			if h.isSensitiveOperation(operation, request.Market) {
				if request.MFALevel < elevation.MFABasic {
					return fmt.Errorf("operação sensível '%s' requer pelo menos MFA básico", operation)
				}
			}
		}
	}

	return nil
}

// GetApprovers retorna a lista de usuários que podem aprovar a solicitação
func (h *FigmaHook) GetApprovers(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) ([]string, error) {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.GetApprovers")
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

	// Verifica se é operação em bibliotecas sensíveis
	if metadata, ok := request.Metadata.(map[string]interface{}); ok {
		if libraryID, ok := metadata["library_id"].(string); ok {
			if h.isSensitiveLibrary(libraryID, request.Market) {
				highSensitivity = true
			}
		}
	}

	// Se é alta sensibilidade, inclui aprovadores de design
	if highSensitivity {
		return []string{"designLead1", "designLead2", "brandManager1"}, nil
	}

	// Operações normais
	return []string{"designApprover1", "designApprover2", "teamLead1"}, nil
}

// ValidateElevationUse valida o uso de um token
func (h *FigmaHook) ValidateElevationUse(
	ctx context.Context, 
	tokenID string, 
	scope string, 
	metadata map[string]interface{},
) error {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.ValidateElevationUse",
		trace.WithAttributes(
			attribute.String("token_id", tokenID),
			attribute.String("scope", scope),
		))
	defer span.End()

	// Obtém o mercado dos metadados
	market, _ := metadata["market"].(string)
	if market == "" {
		market = "default"
	}

	// Validação específica para operações Figma
	if operation, ok := metadata["operation"].(string); ok {
		// Verifica operações proibidas
		if h.isOperationForbidden(operation, market) {
			return fmt.Errorf("operação '%s' é proibida por política de segurança", operation)
		}

		// Verifica escopo necessário para a operação
		requiredScope, err := h.getScopeForOperation(operation)
		if err != nil {
			h.logger.Warn("Operação sem mapeamento de escopo definido",
				zap.String("operation", operation),
				zap.String("token_id", tokenID))
		} else if requiredScope != scope && scope != ScopeFigmaAdmin {
			return fmt.Errorf("escopo %s insuficiente para operação %s, requer %s", 
				scope, operation, requiredScope)
		}

		// Verificações para projetos protegidos
		if projectID, ok := metadata["project_id"].(string); ok {
			if h.isProtectedProject(projectID, market) && scope != ScopeFigmaAdmin && scope != ScopeFigmaEdit {
				return fmt.Errorf("operação em projeto protegido '%s' requer escopo admin ou edit", projectID)
			}
		}

		// Verificações para bibliotecas sensíveis
		if libraryID, ok := metadata["library_id"].(string); ok {
			if h.isSensitiveLibrary(libraryID, market) && scope != ScopeFigmaAdmin && scope != ScopeFigmaLibrary {
				return fmt.Errorf("operação em biblioteca sensível '%s' requer escopo admin ou library", libraryID)
			}
		}
	}

	h.logger.Info("Uso de token Figma validado com sucesso", 
		zap.String("token_id", tokenID),
		zap.String("scope", scope))

	return nil
}

// GetPolicyLimits retorna os limites da política
func (h *FigmaHook) GetPolicyLimits(
	ctx context.Context, 
	tenantID string, 
	market string,
) (*PolicyLimits, error) {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.GetPolicyLimits",
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
				ScopeFigmaAdmin, 
				ScopeFigmaLibrary,
				ScopeFigmaTeam,
			},
		},
		AllowedScopes: []string{
			ScopeFigmaAdmin,
			ScopeFigmaView,
			ScopeFigmaEdit,
			ScopeFigmaComment,
			ScopeFigmaExport,
			ScopeFigmaLibrary,
			ScopeFigmaTeam,
		},
		RoleBasedRestrictions: map[string]interface{}{
			"designer": []string{
				ScopeFigmaView,
				ScopeFigmaEdit,
				ScopeFigmaComment,
				ScopeFigmaExport,
			},
			"viewer": []string{
				ScopeFigmaView,
				ScopeFigmaComment,
			},
			"design_lead": []string{
				ScopeFigmaView,
				ScopeFigmaEdit,
				ScopeFigmaComment,
				ScopeFigmaExport,
				ScopeFigmaLibrary,
			},
			"design_admin": []string{
				ScopeFigmaAdmin,
				ScopeFigmaTeam,
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
func (h *FigmaHook) GetAuditMetadata(
	ctx context.Context, 
	tokenID string, 
	scope string,
) (map[string]interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "FigmaHook.GetAuditMetadata")
	defer span.End()

	return map[string]interface{}{
		"hook_type":     "figma",
		"scope":         scope,
		"timestamp":     time.Now().Format(time.RFC3339),
		"token_id":      tokenID,
		"audit_source":  "innovabiz-iam-elevation",
	}, nil
}

// Métodos auxiliares

// getMarketConfig retorna configuração específica do mercado ou padrão
func (h *FigmaHook) getMarketConfig(market string) *MarketFigmaConfig {
	if h.config == nil || h.config.MarketConfigs == nil {
		// Configuração padrão em caso de erro
		return &MarketFigmaConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     120,
			RequireApproval:        false,
			RequireMFA:             false,
			DefaultMFALevel:        elevation.MFANone,
			AllowEmergencyMode:     true,
			SensitiveOperations:    []string{"delete_library", "delete_project", "transfer_ownership"},
			ForbiddenOperations:    []string{"delete_team", "delete_organization"},
		}
	}

	config, exists := h.config.MarketConfigs[market]
	if !exists {
		if h.config.DefaultConfig != nil {
			return h.config.DefaultConfig
		}
		
		// Configuração padrão em caso de erro
		return &MarketFigmaConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     120,
			RequireApproval:        false,
			RequireMFA:             false,
			DefaultMFALevel:        elevation.MFANone,
			AllowEmergencyMode:     true,
			SensitiveOperations:    []string{"delete_library", "delete_project", "transfer_ownership"},
			ForbiddenOperations:    []string{"delete_team", "delete_organization"},
		}
	}
	
	return config
}

// isProtectedProject verifica se um projeto é protegido
func (h *FigmaHook) isProtectedProject(projectID string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, protectedProject := range config.ProtectedProjects {
		if strings.HasPrefix(projectID, protectedProject) {
			return true
		}
	}
	
	return false
}

// isSensitiveLibrary verifica se uma biblioteca é sensível
func (h *FigmaHook) isSensitiveLibrary(libraryID string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, sensitiveLibrary := range config.SensitiveLibraries {
		if strings.HasPrefix(libraryID, sensitiveLibrary) {
			return true
		}
	}
	
	return false
}

// isSensitiveOperation verifica se uma operação é sensível
func (h *FigmaHook) isSensitiveOperation(operation string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, sensitiveOp := range config.SensitiveOperations {
		if strings.Contains(operation, sensitiveOp) {
			return true
		}
	}
	
	return false
}

// isOperationForbidden verifica se uma operação é proibida
func (h *FigmaHook) isOperationForbidden(operation string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, forbidden := range config.ForbiddenOperations {
		if strings.Contains(operation, forbidden) {
			return true
		}
	}
	
	return false
}

// getScopeForOperation retorna o escopo necessário para uma operação
func (h *FigmaHook) getScopeForOperation(operation string) (string, error) {
	if h.config == nil || h.config.OperationScopeMap == nil {
		// Mapeamento padrão
		operationMap := map[string]string{
			"view_file":          ScopeFigmaView,
			"view_node":          ScopeFigmaView,
			"export_file":        ScopeFigmaExport,
			"export_node":        ScopeFigmaExport,
			"add_comment":        ScopeFigmaComment,
			"post_comment":       ScopeFigmaComment,
			"reply_to_comment":   ScopeFigmaComment,
			"edit_file":          ScopeFigmaEdit,
			"update_node":        ScopeFigmaEdit,
			"create_component":   ScopeFigmaEdit,
			"create_library":     ScopeFigmaLibrary,
			"update_library":     ScopeFigmaLibrary,
			"delete_library":     ScopeFigmaLibrary,
			"manage_team":        ScopeFigmaTeam,
			"add_team_member":    ScopeFigmaTeam,
			"remove_team_member": ScopeFigmaTeam,
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