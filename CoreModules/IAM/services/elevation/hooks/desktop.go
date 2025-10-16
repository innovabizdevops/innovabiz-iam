// Package hooks implementa a integração entre o serviço de elevação
// e os diferentes hooks MCP (Docker, Desktop Commander, GitHub, Figma)
// da plataforma INNOVABIZ IAM.
package hooks

import (
	"context"
	"fmt"
	"path/filepath"
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

// DesktopCommanderHook implementa a interface MCPHook para Desktop Commander
type DesktopCommanderHook struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *config.DesktopElevationConfig
}

// DesktopElevationConfig contém configurações específicas para o hook Desktop Commander
type DesktopElevationConfig struct {
	// Mapeamento de comandos para escopos necessários
	CommandScopeMap map[string]string
	
	// Mapeamento de operações de arquivo para escopos necessários
	FileOpScopeMap map[string]string
	
	// Configurações específicas por mercado
	MarketConfigs map[string]*MarketDesktopConfig
	
	// Valores padrão quando mercado específico não configurado
	DefaultConfig *MarketDesktopConfig
	
	// Diretórios protegidos que requerem elevação especial
	ProtectedDirectories []string
}

// MarketDesktopConfig contém configurações específicas por mercado
type MarketDesktopConfig struct {
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
	
	// ForbiddenFilePatterns são padrões de arquivo proibidos
	ForbiddenFilePatterns []string
	
	// SensitiveFilePatterns são padrões de arquivos sensíveis
	SensitiveFilePatterns []string
}

// NewDesktopCommanderHook cria uma nova instância do hook Desktop Commander
func NewDesktopCommanderHook(config *config.DesktopElevationConfig) *DesktopCommanderHook {
	return &DesktopCommanderHook{
		logger: logging.GetLogger().Named("desktop-commander-hook"),
		tracer: otel.Tracer("innovabiz/iam/elevation/hooks/desktop"),
		config: config,
	}
}

// HookType retorna o tipo do hook
func (h *DesktopCommanderHook) HookType() MCPHookType {
	return DesktopCommander
}

// ValidateScope valida um escopo para Desktop Commander
func (h *DesktopCommanderHook) ValidateScope(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (*ScopeDetails, error) {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.ValidateScope",
		trace.WithAttributes(
			attribute.String("scope", scope),
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	// Verifica se o escopo começa com o prefixo desktop:
	if !strings.HasPrefix(scope, "desktop:") {
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: DesktopCommander,
			Reason:   "escopo deve começar com 'desktop:'",
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
	case ScopeDesktopAdmin:
		details.DisplayName = "Administração Desktop"
		details.Description = "Acesso completo a todas operações do sistema"
		details.SensitivityLevel = 5
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	case ScopeDesktopFS:
		details.DisplayName = "Acesso ao Sistema de Arquivos"
		details.Description = "Operações no sistema de arquivos (leitura/escrita)"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 60 // 1 hora

	case ScopeDesktopCmd:
		details.DisplayName = "Execução de Comandos"
		details.Description = "Execução de comandos no terminal"
		details.SensitivityLevel = 4
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 60 // 1 hora

	case ScopeDesktopProcess:
		details.DisplayName = "Gerenciamento de Processos"
		details.Description = "Iniciar, parar e monitorar processos"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 120 // 2 horas

	case ScopeDesktopConfig:
		details.DisplayName = "Configuração do Sistema"
		details.Description = "Alteração de configurações do sistema"
		details.SensitivityLevel = 5
		details.RequiresApproval = true
		details.RequiredMFA = elevation.MFAStrong
		details.AllowedInEmergency = false
		details.MaxDuration = 30 // 30 minutos

	case ScopeDesktopSearch:
		details.DisplayName = "Busca de Arquivos"
		details.Description = "Busca e indexação de arquivos"
		details.SensitivityLevel = 2
		details.RequiresApproval = false
		details.RequiredMFA = elevation.MFANone
		details.AllowedInEmergency = true
		details.MaxDuration = 180 // 3 horas

	case ScopeDesktopEdit:
		details.DisplayName = "Edição de Arquivos"
		details.Description = "Edição de arquivos no sistema"
		details.SensitivityLevel = 3
		details.RequiresApproval = marketConfig.RequireApproval
		details.RequiredMFA = elevation.MFABasic
		details.AllowedInEmergency = true
		details.MaxDuration = 120 // 2 horas

	default:
		return nil, &ScopeNotAllowedError{
			Scope:    scope,
			HookType: DesktopCommander,
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
			if details.RequiredMFA < elevation.MFABasic {
				details.RequiredMFA = elevation.MFABasic
			}
		}
	} else if market == "brasil" {
		// Brasil (LGPD) tem requisitos específicos para dados pessoais
		if scope == ScopeDesktopFS || scope == ScopeDesktopEdit {
			details.RequiresApproval = true
			details.RequiredMFA = elevation.MFABasic
		}
	}

	return details, nil
}

// GetRequiredMFA retorna o nível de MFA necessário para o escopo
func (h *DesktopCommanderHook) GetRequiredMFA(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (elevation.MFALevel, error) {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.GetRequiredMFA")
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
func (h *DesktopCommanderHook) GetRequireApproval(
	ctx context.Context, 
	scope string, 
	tenantID string, 
	market string,
) (bool, error) {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.GetRequireApproval")
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
func (h *DesktopCommanderHook) ValidateRequest(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) error {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.ValidateRequest")
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
				HookType:      DesktopCommander,
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

	// Verificações específicas para metadados do Desktop Commander
	if metadata, ok := request.Metadata.(map[string]interface{}); ok {
		// Verifica operações em caminhos protegidos
		if path, ok := metadata["path"].(string); ok {
			if h.isProtectedPath(path) && !request.Emergency {
				return fmt.Errorf("operação em caminho protegido '%s' requer aprovação", path)
			}
		}

		// Verifica comandos sensíveis
		if command, ok := metadata["command"].(string); ok {
			if h.isSensitiveCommand(command, request.Market) {
				if request.MFALevel < elevation.MFABasic {
					return fmt.Errorf("comando sensível '%s' requer pelo menos MFA básico", command)
				}
				
				if !request.Emergency && !h.hasScopeForCommand(command, request.Scopes) {
					return fmt.Errorf("escopo insuficiente para comando sensível '%s'", command)
				}
			}
		}
	}

	return nil
}
// GetApprovers retorna a lista de usuários que podem aprovar a solicitação
func (h *DesktopCommanderHook) GetApprovers(
	ctx context.Context, 
	request *elevation.ElevationRequest,
) ([]string, error) {
	// Nota: Na implementação real, isso consultaria o serviço de tenant 
	// para obter aprovadores específicos baseados em papéis e políticas
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.GetApprovers")
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

	// Verifica se é operação em arquivos sensíveis
	if metadata, ok := request.Metadata.(map[string]interface{}); ok {
		if path, ok := metadata["path"].(string); ok {
			if h.isSensitiveFile(path, request.Market) {
				highSensitivity = true
			}
		}
	}

	// Determina aprovadores com base na sensibilidade
	if highSensitivity {
		// Alta sensibilidade requer aprovadores de segurança
		return []string{"securityOfficer1", "securityOfficer2", "admin1"}, nil
	} else if request.Market == "angola" || request.Market == "mozambique" {
		// Mercados SADC/PALOP requerem aprovadores locais específicos
		return []string{"localCompliance1", "localAdmin1", "admin1"}, nil
	} else if request.Market == "brasil" {
		// Brasil (LGPD) requer aprovadores específicos
		return []string{"dataProtectionOfficer1", "localCompliance1", "admin1"}, nil
	}

	// Operações normais
	return []string{"teamLead1", "teamLead2", "admin1", "admin2"}, nil
}

// ValidateElevationUse valida o uso de um token
func (h *DesktopCommanderHook) ValidateElevationUse(
	ctx context.Context, 
	tokenID string, 
	scope string, 
	metadata map[string]interface{},
) error {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.ValidateElevationUse",
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

	// Validação específica para operações Desktop Commander
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
		} else if requiredScope != scope && scope != ScopeDesktopAdmin {
			return fmt.Errorf("escopo %s insuficiente para operação %s, requer %s", 
				scope, operation, requiredScope)
		}

		// Verificações para paths específicos
		if path, ok := metadata["path"].(string); ok {
			// Verifica caminhos protegidos
			if h.isProtectedPath(path) && scope != ScopeDesktopAdmin && scope != ScopeDesktopFS {
				return fmt.Errorf("operação em caminho protegido '%s' requer escopo admin ou filesystem", path)
			}

			// Verifica arquivos sensíveis
			if h.isSensitiveFile(path, market) {
				h.logger.Warn("Operação em arquivo sensível",
					zap.String("path", path),
					zap.String("scope", scope),
					zap.String("operation", operation))

				// Adiciona validações adicionais para arquivos sensíveis se necessário
				if scope != ScopeDesktopAdmin && scope != ScopeDesktopFS {
					return fmt.Errorf("operação em arquivo sensível '%s' requer escopo admin ou filesystem", path)
				}
			}

			// Verifica padrões de arquivo proibidos
			if h.isFilePatternForbidden(path, market) {
				return fmt.Errorf("operação em arquivo proibido '%s' não é permitida", path)
			}
		}

		// Verificações para comandos
		if command, ok := metadata["command"].(string); ok {
			if h.isCommandForbidden(command, market) {
				return fmt.Errorf("comando '%s' é proibido por política de segurança", command)
			}

			if h.isSensitiveCommand(command, market) && scope != ScopeDesktopAdmin && scope != ScopeDesktopCmd {
				return fmt.Errorf("comando sensível '%s' requer escopo admin ou command", command)
			}
		}
	}

	h.logger.Info("Uso de token Desktop Commander validado com sucesso", 
		zap.String("token_id", tokenID),
		zap.String("scope", scope))

	return nil
}

// GetPolicyLimits retorna os limites da política
func (h *DesktopCommanderHook) GetPolicyLimits(
	ctx context.Context, 
	tenantID string, 
	market string,
) (*PolicyLimits, error) {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.GetPolicyLimits",
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
				ScopeDesktopAdmin, 
				ScopeDesktopConfig,
			},
		},
		AllowedScopes: []string{
			ScopeDesktopAdmin,
			ScopeDesktopFS,
			ScopeDesktopCmd,
			ScopeDesktopProcess,
			ScopeDesktopConfig,
			ScopeDesktopSearch,
			ScopeDesktopEdit,
		},
		RoleBasedRestrictions: map[string]interface{}{
			"developer": []string{
				ScopeDesktopFS,
				ScopeDesktopCmd,
				ScopeDesktopProcess,
				ScopeDesktopSearch,
				ScopeDesktopEdit,
			},
			"operator": []string{
				ScopeDesktopFS,
				ScopeDesktopCmd,
				ScopeDesktopProcess,
			},
			"security": []string{
				ScopeDesktopFS,
				ScopeDesktopConfig,
				ScopeDesktopSearch,
			},
			"admin": []string{
				ScopeDesktopAdmin,
			},
		},
	}

	// Ajustes baseados no mercado (SADC, PALOP, BRICS, etc.)
	if market == "angola" || market == "mozambique" {
		limits.RequireMFA = true
		limits.DefaultMFALevel = elevation.MFABasic
		limits.EmergencyModeRestrictions["requires_secondary_approval"] = true
	} else if market == "brasil" {
		// Ajustes específicos para LGPD no Brasil
		limits.EmergencyModeRestrictions["data_access_logging"] = true
	}

	return limits, nil
}

// GetAuditMetadata retorna metadados para auditoria
func (h *DesktopCommanderHook) GetAuditMetadata(
	ctx context.Context, 
	tokenID string, 
	scope string,
) (map[string]interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "DesktopCommanderHook.GetAuditMetadata")
	defer span.End()

	return map[string]interface{}{
		"hook_type":     "desktop-commander",
		"scope":         scope,
		"timestamp":     time.Now().Format(time.RFC3339),
		"token_id":      tokenID,
		"audit_source":  "innovabiz-iam-elevation",
	}, nil
}

// Métodos auxiliares

// getMarketConfig retorna configuração específica do mercado ou padrão
func (h *DesktopCommanderHook) getMarketConfig(market string) *MarketDesktopConfig {
	if h.config == nil || h.config.MarketConfigs == nil {
		// Configuração padrão em caso de erro
		return &MarketDesktopConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     60,
			RequireApproval:        true,
			RequireMFA:             true,
			DefaultMFALevel:        elevation.MFABasic,
			AllowEmergencyMode:     true,
			SensitiveCommands:      []string{"rm", "chmod", "chown", "dd", "wget", "curl", "ssh"},
			ForbiddenCommands:      []string{"rm -rf /", "format", "fdisk", "mkfs"},
			ForbiddenFilePatterns:  []string{"/etc/shadow", "/etc/passwd", "id_rsa"},
			SensitiveFilePatterns:  []string{".env", "config.json", ".key", ".pem", ".sql"},
		}
	}

	config, exists := h.config.MarketConfigs[market]
	if !exists {
		if h.config.DefaultConfig != nil {
			return h.config.DefaultConfig
		}
		
		// Configuração padrão em caso de erro
		return &MarketDesktopConfig{
			MaxActiveTokensPerUser: 5,
			DefaultMaxDuration:     60,
			RequireApproval:        true,
			RequireMFA:             true,
			DefaultMFALevel:        elevation.MFABasic,
			AllowEmergencyMode:     true,
			SensitiveCommands:      []string{"rm", "chmod", "chown", "dd", "wget", "curl", "ssh"},
			ForbiddenCommands:      []string{"rm -rf /", "format", "fdisk", "mkfs"},
			ForbiddenFilePatterns:  []string{"/etc/shadow", "/etc/passwd", "id_rsa"},
			SensitiveFilePatterns:  []string{".env", "config.json", ".key", ".pem", ".sql"},
		}
	}
	
	return config
}

// isProtectedPath verifica se um caminho é protegido
func (h *DesktopCommanderHook) isProtectedPath(path string) bool {
	if h.config == nil || len(h.config.ProtectedDirectories) == 0 {
		// Diretórios protegidos padrão
		protectedDirs := []string{
			"/etc", 
			"/var/log", 
			"/boot", 
			"/usr/bin", 
			"/bin", 
			"/sbin", 
			"/usr/sbin",
			"C:\\Windows", 
			"C:\\Program Files", 
			"C:\\Program Files (x86)",
		}
		
		for _, dir := range protectedDirs {
			if strings.HasPrefix(path, dir) {
				return true
			}
		}
		
		return false
	}
	
	for _, dir := range h.config.ProtectedDirectories {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	
	return false
}

// isSensitiveFile verifica se um arquivo é sensível
func (h *DesktopCommanderHook) isSensitiveFile(path string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, pattern := range config.SensitiveFilePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
		
		// Verifica extensão
		if strings.HasSuffix(path, pattern) {
			return true
		}
		
		// Verifica correspondência de nome de arquivo
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err == nil && matched {
			return true
		}
	}
	
	return false
}

// isFilePatternForbidden verifica se um padrão de arquivo é proibido
func (h *DesktopCommanderHook) isFilePatternForbidden(path string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, pattern := range config.ForbiddenFilePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
		
		// Verifica correspondência de caminho completo
		if matched, err := filepath.Match(pattern, path); err == nil && matched {
			return true
		}
		
		// Verifica correspondência de nome de arquivo
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err == nil && matched {
			return true
		}
	}
	
	return false
}

// isSensitiveCommand verifica se um comando é sensível
func (h *DesktopCommanderHook) isSensitiveCommand(command string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, sensitiveCmd := range config.SensitiveCommands {
		// Verifica se o comando começa com o padrão sensível
		if strings.HasPrefix(command, sensitiveCmd+" ") || command == sensitiveCmd {
			return true
		}
	}
	
	return false
}

// isCommandForbidden verifica se um comando é proibido
func (h *DesktopCommanderHook) isCommandForbidden(command string, market string) bool {
	config := h.getMarketConfig(market)
	
	for _, forbidden := range config.ForbiddenCommands {
		if strings.Contains(command, forbidden) {
			return true
		}
	}
	
	return false
}

// isOperationForbidden verifica se uma operação é proibida
func (h *DesktopCommanderHook) isOperationForbidden(operation string, market string) bool {
	// Por enquanto, delegamos para isCommandForbidden
	return h.isCommandForbidden(operation, market)
}

// getScopeForOperation retorna o escopo necessário para uma operação
func (h *DesktopCommanderHook) getScopeForOperation(operation string) (string, error) {
	if h.config == nil || h.config.CommandScopeMap == nil {
		// Mapeamento padrão
		operationMap := map[string]string{
			"read_file":             ScopeDesktopFS,
			"write_file":            ScopeDesktopFS,
			"create_directory":      ScopeDesktopFS,
			"list_directory":        ScopeDesktopSearch,
			"move_file":             ScopeDesktopFS,
			"search_code":           ScopeDesktopSearch,
			"search_files":          ScopeDesktopSearch,
			"get_file_info":         ScopeDesktopFS,
			"edit_block":            ScopeDesktopEdit,
			"start_process":         ScopeDesktopProcess,
			"interact_with_process": ScopeDesktopProcess,
			"kill_process":          ScopeDesktopProcess,
			"execute_command":       ScopeDesktopCmd,
			"set_config":            ScopeDesktopConfig,
		}
		
		for op, scope := range operationMap {
			if strings.HasPrefix(operation, op) {
				return scope, nil
			}
		}
		
		return "", fmt.Errorf("operação sem mapeamento de escopo")
	}
	
	for op, scope := range h.config.CommandScopeMap {
		if strings.HasPrefix(operation, op) {
			return scope, nil
		}
	}
	
	// Se operação não encontrada, busca no mapeamento de operações de arquivo
	if h.config.FileOpScopeMap != nil {
		for op, scope := range h.config.FileOpScopeMap {
			if strings.HasPrefix(operation, op) {
				return scope, nil
			}
		}
	}
	
	return "", fmt.Errorf("operação sem mapeamento de escopo")
}

// hasScopeForCommand verifica se um conjunto de escopos contém o escopo necessário para um comando
func (h *DesktopCommanderHook) hasScopeForCommand(command string, scopes []string) bool {
	requiredScope, err := h.getScopeForOperation(command)
	if err != nil {
		return false
	}
	
	// Verifica se o escopo necessário está na lista ou se há escopo admin
	for _, scope := range scopes {
		if scope == requiredScope || scope == ScopeDesktopAdmin {
			return true
		}
	}
	
	return false
}