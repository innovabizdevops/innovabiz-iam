// Package hooks implementa a integração entre o serviço de elevação
// e os diferentes hooks MCP (Docker, Desktop Commander, GitHub, Figma)
// da plataforma INNOVABIZ IAM.
package hooks

import (
	"context"
	"fmt"
	"sync"

	"github.com/innovabiz/iam/config"
	"github.com/innovabiz/iam/logging"
	"go.uber.org/zap"
)

// HookRegistry mantém o registro de todos os hooks MCP disponíveis
type HookRegistry struct {
	hooks  map[MCPHookType]MCPHook
	mutex  sync.RWMutex
	logger *zap.Logger
	config *config.ElevationConfig
}

// NewHookRegistry cria um novo registro de hooks
func NewHookRegistry(cfg *config.ElevationConfig) *HookRegistry {
	registry := &HookRegistry{
		hooks:  make(map[MCPHookType]MCPHook),
		logger: logging.GetLogger().Named("hook-registry"),
		config: cfg,
	}

	// Inicializa hooks padrão se configurações disponíveis
	if cfg != nil {
		registry.initDefaultHooks()
	}

	return registry
}

// initDefaultHooks inicializa hooks padrão
func (r *HookRegistry) initDefaultHooks() {
	// Registra o hook Docker se configurado
	if r.config.Docker != nil {
		dockerHook := NewDockerHook(r.config.Docker)
		r.RegisterHook(dockerHook)
		r.logger.Info("Hook Docker registrado")
	}

	// Registra o hook GitHub se configurado
	if r.config.GitHub != nil {
		githubHook := NewGitHubHook(r.config.GitHub)
		r.RegisterHook(githubHook)
		r.logger.Info("Hook GitHub registrado")
	}

	// Registra o hook Desktop Commander se configurado
	if r.config.DesktopCommander != nil {
		desktopHook := NewDesktopCommanderHook(r.config.DesktopCommander)
		r.RegisterHook(desktopHook)
		r.logger.Info("Hook Desktop Commander registrado")
	}

	// Registra o hook Figma se configurado
	if r.config.Figma != nil {
		figmaHook := NewFigmaHook(r.config.Figma)
		r.RegisterHook(figmaHook)
		r.logger.Info("Hook Figma registrado")
	}
}

// RegisterHook registra um hook no registro
func (r *HookRegistry) RegisterHook(hook MCPHook) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	hookType := hook.HookType()
	r.hooks[hookType] = hook
	r.logger.Info("Hook registrado com sucesso", 
		zap.String("type", string(hookType)))
}

// GetHook retorna um hook pelo tipo
func (r *HookRegistry) GetHook(hookType MCPHookType) (MCPHook, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	hook, exists := r.hooks[hookType]
	if !exists {
		return nil, fmt.Errorf("hook do tipo %s não encontrado", hookType)
	}

	return hook, nil
}

// GetAllHooks retorna todos os hooks registrados
func (r *HookRegistry) GetAllHooks() map[MCPHookType]MCPHook {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Cria uma cópia do mapa para evitar condições de corrida
	result := make(map[MCPHookType]MCPHook, len(r.hooks))
	for hookType, hook := range r.hooks {
		result[hookType] = hook
	}

	return result
}

// ValidateScopesForAllHooks valida escopos para todos os hooks
func (r *HookRegistry) ValidateScopesForAllHooks(
	ctx context.Context, 
	scopes []string, 
	tenantID string, 
	market string,
) (map[MCPHookType][]string, error) {
	result := make(map[MCPHookType][]string)

	for _, scope := range scopes {
		// Determina o tipo de hook baseado no prefixo do escopo
		var hookType MCPHookType
		switch {
		case IsDockerScope(scope):
			hookType = Docker
		case IsGitHubScope(scope):
			hookType = GitHub
		case IsDesktopCommanderScope(scope):
			hookType = DesktopCommander
		case IsFigmaScope(scope):
			hookType = Figma
		default:
			return nil, fmt.Errorf("tipo de escopo desconhecido: %s", scope)
		}

		hook, err := r.GetHook(hookType)
		if err != nil {
			return nil, fmt.Errorf("hook não disponível para escopo %s: %w", scope, err)
		}

		// Valida o escopo com o hook apropriado
		_, err = hook.ValidateScope(ctx, scope, tenantID, market)
		if err != nil {
			return nil, fmt.Errorf("validação de escopo falhou: %w", err)
		}

		// Adiciona escopo ao resultado
		if _, exists := result[hookType]; !exists {
			result[hookType] = []string{scope}
		} else {
			result[hookType] = append(result[hookType], scope)
		}
	}

	return result, nil
}

// IsDockerScope verifica se o escopo é para Docker
func IsDockerScope(scope string) bool {
	return scope == ScopeDockerAdmin || 
		scope == ScopeDockerExec || 
		scope == ScopeDockerBuild ||
		scope == ScopeDockerPull ||
		scope == ScopeDockerPush ||
		scope == ScopeDockerRun ||
		scope == ScopeDockerNetwork ||
		scope == ScopeDockerVolume
}

// IsGitHubScope verifica se o escopo é para GitHub
func IsGitHubScope(scope string) bool {
	return scope == ScopeGitHubAdmin || 
		scope == ScopeGitHubRead || 
		scope == ScopeGitHubWrite ||
		scope == ScopeGitHubDelete ||
		scope == ScopeGitHubPR ||
		scope == ScopeGitHubIssues ||
		scope == ScopeGitHubSecurity
}

// IsDesktopCommanderScope verifica se o escopo é para Desktop Commander
func IsDesktopCommanderScope(scope string) bool {
	return scope == ScopeDesktopAdmin || 
		scope == ScopeDesktopFS || 
		scope == ScopeDesktopCmd ||
		scope == ScopeDesktopProcess ||
		scope == ScopeDesktopConfig ||
		scope == ScopeDesktopSearch ||
		scope == ScopeDesktopEdit
}

// IsFigmaScope verifica se o escopo é para Figma
func IsFigmaScope(scope string) bool {
	return scope == ScopeFigmaAdmin || 
		scope == ScopeFigmaView || 
		scope == ScopeFigmaEdit ||
		scope == ScopeFigmaComment ||
		scope == ScopeFigmaExport ||
		scope == ScopeFigmaLibrary ||
		scope == ScopeFigmaTeam
}