// Package hooks implementa a integração entre o serviço de elevação
// e os diferentes hooks MCP (Docker, Desktop Commander, GitHub, Figma)
// da plataforma INNOVABIZ IAM.
package hooks

import (
	"context"

	"github.com/innovabiz/iam/services/elevation"
)

// MCPHookType identifica o tipo de hook MCP
type MCPHookType string

const (
	// Docker representa o hook MCP para Docker
	Docker MCPHookType = "docker"
	
	// DesktopCommander representa o hook MCP para Desktop Commander
	DesktopCommander MCPHookType = "desktop-commander"
	
	// GitHub representa o hook MCP para GitHub
	GitHub MCPHookType = "github"
	
	// Figma representa o hook MCP para Figma
	Figma MCPHookType = "figma"
)

// MCPHook define a interface comum para todos os hooks MCP
type MCPHook interface {
	// HookType retorna o tipo do hook MCP
	HookType() MCPHookType
	
	// ValidateScope valida se o escopo requisitado é válido para este hook MCP
	// e retorna detalhes sobre o escopo requisitado
	ValidateScope(ctx context.Context, scope string, tenantID string, market string) (*ScopeDetails, error)
	
	// GetRequiredMFA retorna o nível de MFA necessário para o escopo
	GetRequiredMFA(ctx context.Context, scope string, tenantID string, market string) (elevation.MFALevel, error)
	
	// GetRequireApproval determina se o escopo necessita de aprovação manual
	GetRequireApproval(ctx context.Context, scope string, tenantID string, market string) (bool, error)
	
	// ValidateRequest valida uma solicitação de elevação para o hook
	ValidateRequest(ctx context.Context, request *elevation.ElevationRequest) error
	
	// GetApprovers retorna a lista de usuários que podem aprovar a solicitação
	GetApprovers(ctx context.Context, request *elevation.ElevationRequest) ([]string, error)
	
	// ValidateElevationUse valida o uso de um token de elevação para uma operação específica
	// Recebe o token, o escopo sendo acessado e metadados adicionais da operação
	ValidateElevationUse(ctx context.Context, tokenID string, scope string, metadata map[string]interface{}) error
	
	// GetPolicyLimits retorna os limites da política de elevação para o hook
	GetPolicyLimits(ctx context.Context, tenantID string, market string) (*PolicyLimits, error)
	
	// GetAuditMetadata retorna metadados adicionais para auditoria específicos do hook
	GetAuditMetadata(ctx context.Context, tokenID string, scope string) (map[string]interface{}, error)
}

// ScopeDetails contém informações sobre um escopo específico
type ScopeDetails struct {
	// Scope é o escopo original
	Scope string
	
	// DisplayName é o nome de exibição do escopo
	DisplayName string
	
	// Description é a descrição do escopo
	Description string
	
	// SensitivityLevel indica o nível de sensibilidade do escopo
	SensitivityLevel int
	
	// MaxDuration é a duração máxima permitida para este escopo
	MaxDuration int
	
	// RequiresApproval indica se o escopo requer aprovação
	RequiresApproval bool
	
	// RequiredMFA indica o nível de MFA necessário
	RequiredMFA elevation.MFALevel
	
	// AllowedInEmergency indica se o escopo é permitido em modo de emergência
	AllowedInEmergency bool
	
	// DenyMessage é a mensagem de negação caso o escopo não seja permitido
	DenyMessage string
	
	// Metadata são metadados adicionais específicos do hook
	Metadata map[string]interface{}
}

// PolicyLimits define os limites da política de elevação para um hook
type PolicyLimits struct {
	// MaxActiveTokensPerUser é o número máximo de tokens ativos permitidos por usuário
	MaxActiveTokensPerUser int
	
	// DefaultMaxDuration é a duração máxima padrão para tokens em minutos
	DefaultMaxDuration int
	
	// AbsoluteMaxDuration é a duração máxima absoluta permitida em minutos
	AbsoluteMaxDuration int
	
	// RequireJustification indica se é necessário fornecer justificativa
	RequireJustification bool
	
	// MinJustificationLength é o tamanho mínimo da justificativa
	MinJustificationLength int
	
	// RequireMFA indica se MFA é necessário por padrão
	RequireMFA bool
	
	// DefaultMFALevel é o nível de MFA padrão requerido
	DefaultMFALevel elevation.MFALevel
	
	// AllowEmergencyMode indica se o modo de emergência é permitido
	AllowEmergencyMode bool
	
	// EmergencyModeRestrictions são restrições adicionais para o modo de emergência
	EmergencyModeRestrictions map[string]interface{}
	
	// AllowedScopes são os escopos permitidos para este hook
	AllowedScopes []string
	
	// RoleBasedRestrictions são restrições baseadas em papel
	RoleBasedRestrictions map[string]interface{}
}

// ScopeNotAllowedError indica que um escopo não é permitido para o hook
type ScopeNotAllowedError struct {
	Scope       string
	HookType    MCPHookType
	Reason      string
	Market      string
	TenantID    string
}

func (e *ScopeNotAllowedError) Error() string {
	return "escopo não permitido para o hook: " + string(e.HookType) + ", escopo: " + e.Scope + ", motivo: " + e.Reason
}

// InsufficientMFAError indica que o nível de MFA é insuficiente
type InsufficientMFAError struct {
	Scope           string
	HookType        MCPHookType
	RequiredLevel   elevation.MFALevel
	ProvidedLevel   elevation.MFALevel
}

func (e *InsufficientMFAError) Error() string {
	return "nível de MFA insuficiente para o hook: " + string(e.HookType) + 
		", escopo: " + e.Scope + 
		", necessário: " + string(e.RequiredLevel) + 
		", fornecido: " + string(e.ProvidedLevel)
}

// PolicyLimitExceededError indica que um limite de política foi excedido
type PolicyLimitExceededError struct {
	Limit       string
	HookType    MCPHookType
	Current     int
	Maximum     int
}

func (e *PolicyLimitExceededError) Error() string {
	return "limite de política excedido para o hook: " + string(e.HookType) + 
		", limite: " + e.Limit + 
		", atual: " + string(rune(e.Current)) + 
		", máximo: " + string(rune(e.Maximum))
}

// HookRegistry mantém o registro de todos os hooks MCP disponíveis
type HookRegistry struct {
	hooks map[MCPHookType]MCPHook
}

// NewHookRegistry cria um novo registro de hooks
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[MCPHookType]MCPHook),
	}
}

// RegisterHook registra um novo hook no registro
func (r *HookRegistry) RegisterHook(hook MCPHook) {
	r.hooks[hook.HookType()] = hook
}

// GetHook retorna um hook pelo tipo
func (r *HookRegistry) GetHook(hookType MCPHookType) (MCPHook, bool) {
	hook, exists := r.hooks[hookType]
	return hook, exists
}

// GetAllHooks retorna todos os hooks registrados
func (r *HookRegistry) GetAllHooks() []MCPHook {
	hooks := make([]MCPHook, 0, len(r.hooks))
	for _, hook := range r.hooks {
		hooks = append(hooks, hook)
	}
	return hooks
}

// GetSupportedHookTypes retorna os tipos de hooks suportados
func (r *HookRegistry) GetSupportedHookTypes() []MCPHookType {
	types := make([]MCPHookType, 0, len(r.hooks))
	for hookType := range r.hooks {
		types = append(types, hookType)
	}
	return types
}

// Predefined scope constants for different hooks
const (
	// Docker scopes
	ScopeDockerAdmin       = "docker:admin"
	ScopeDockerExec        = "docker:exec"
	ScopeDockerPush        = "docker:push"
	ScopeDockerPull        = "docker:pull"
	ScopeDockerBuild       = "docker:build"
	ScopeDockerVolume      = "docker:volume"
	ScopeDockerNetwork     = "docker:network"
	ScopeDockerSystem      = "docker:system"
	
	// Desktop Commander scopes
	ScopeDesktopFS         = "desktop:filesystem"
	ScopeDesktopCmd        = "desktop:command"
	ScopeDesktopProcess    = "desktop:process"
	ScopeDesktopConfig     = "desktop:config"
	ScopeDesktopSearch     = "desktop:search"
	ScopeDesktopEdit       = "desktop:edit"
	ScopeDesktopAdmin      = "desktop:admin"
	
	// GitHub scopes
	ScopeGitHubPush        = "github:push"
	ScopeGitHubMerge       = "github:merge"
	ScopeGitHubSecrets     = "github:secrets"
	ScopeGitHubSettings    = "github:settings"
	ScopeGitHubRepo        = "github:repo"
	ScopeGitHubHooks       = "github:hooks"
	ScopeGitHubAdmin       = "github:admin"
	ScopeGitHubActions     = "github:actions"
	
	// Figma scopes
	ScopeFigmaView         = "figma:view"
	ScopeFigmaEdit         = "figma:edit"
	ScopeFigmaComment      = "figma:comment"
	ScopeFigmaExport       = "figma:export"
	ScopeFigmaAdmin        = "figma:admin"
	ScopeFigmaLibrary      = "figma:library"
	ScopeFigmaTeam         = "figma:team"
)