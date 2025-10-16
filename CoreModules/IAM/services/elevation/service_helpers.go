// Package elevation implementa o serviço core de elevação de privilégios para o sistema MCP-IAM
// da plataforma INNOVABIZ.
package elevation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/innovabiz/iam/audit"
	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/constants"
	"github.com/innovabiz/iam/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Funções auxiliares para o serviço de elevação

// requiresMFA verifica se a política do mercado exige MFA para os escopos solicitados
func (s *Service) requiresMFA(ctx context.Context, market string, scopes []string) bool {
	// Políticas específicas por mercado para requisito de MFA
	// Isto é configurado de acordo com requisitos regulatórios de cada mercado
	marketConfigs := map[string]struct {
		requireMFAForAll  bool                 // Exige MFA para todos os escopos
		requireMFAScopes  map[string]bool      // Escopos específicos que exigem MFA
		exemptMFAScopes   map[string]bool      // Escopos isentos de MFA mesmo quando requireMFAForAll=true
	}{
		"angola": {
			requireMFAForAll: true, // BNA exige MFA para todas as operações sensíveis
			exemptMFAScopes: map[string]bool{
				"read:logs":    true,
				"read:metrics": true,
			},
		},
		"brasil": {
			requireMFAForAll: false,
			requireMFAScopes: map[string]bool{
				"delete:repo":    true,
				"modify:design":  true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
			},
		},
		"mocambique": {
			requireMFAForAll: false,
			requireMFAScopes: map[string]bool{
				"delete:repo":    true,
				"admin:payment":  true,
				"transfer:money": true,
			},
		},
	}
	
	// Configuração padrão para mercados não especificados
	config, exists := marketConfigs[market]
	if !exists {
		// Por segurança, se o mercado não está configurado, exige MFA
		return true
	}
	
	if config.requireMFAForAll {
		// Para mercados que exigem MFA em todas as operações sensíveis
		// Verifica se todos os escopos solicitados estão na lista de isentos
		for _, scope := range scopes {
			if !config.exemptMFAScopes[scope] {
				return true
			}
		}
		return false
	}
	
	// Para mercados que exigem MFA apenas em operações específicas
	for _, scope := range scopes {
		if config.requireMFAScopes[scope] {
			return true
		}
	}
	
	return false
}

// validateScopes verifica se os escopos solicitados são válidos e permitidos
func (s *Service) validateScopes(ctx context.Context, request *ElevationRequest) error {
	ctx, span := s.tracer.Start(ctx, "ElevationService.validateScopes",
		trace.WithAttributes(
			attribute.String("user_id", request.UserID),
			attribute.String("market", request.Market),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto para verificar permissões
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		return errors.New("authentication token not found in context")
	}
	
	// Verifica políticas do mercado para escopos permitidos
	marketScopeRules, err := s.getMarketScopeRules(ctx, request.Market)
	if err != nil {
		return err
	}
	
	// Verifica cada escopo solicitado
	for _, scope := range request.Scopes {
		// Verifica se o escopo é permitido para o mercado
		if !marketScopeRules.IsAllowedScope(scope) {
			s.logger.Warn("Escopo não permitido para o mercado",
				zap.String("scope", scope),
				zap.String("market", request.Market))
			return fmt.Errorf("%w: scope %s not allowed for market %s", ErrScopeNotAllowed, scope, request.Market)
		}
		
		// Verifica se o usuário tem permissão para solicitar este escopo
		hasPermission, err := s.hasPermissionForScope(ctx, request.UserID, request.TenantID, scope)
		if err != nil {
			s.logger.Error("Erro ao verificar permissão para escopo",
				zap.String("user_id", request.UserID),
				zap.String("scope", scope),
				zap.Error(err))
			return fmt.Errorf("failed to check permission for scope %s: %w", scope, err)
		}
		
		if !hasPermission {
			s.logger.Warn("Usuário não tem permissão para o escopo solicitado",
				zap.String("user_id", request.UserID),
				zap.String("scope", scope))
			return ErrInsufficientPermission
		}
	}
	
	return nil
}

// getMarketScopeRules retorna as regras de escopos permitidos para um mercado
func (s *Service) getMarketScopeRules(ctx context.Context, market string) (*MarketScopeRules, error) {
	// Regras específicas por mercado para escopos permitidos
	marketRules := map[string]*MarketScopeRules{
		"angola": {
			AllowedScopes: map[string]bool{
				"delete:repo":    true,
				"create:repo":    true,
				"modify:design":  true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
				"manage:docker":  true,
				"system:command": true,
				"read:logs":      true,
				"read:metrics":   true,
			},
			RequiresApproval: map[string]bool{
				"delete:repo":    true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
				"system:command": true,
			},
			EmergencyAllowed: map[string]bool{
				"delete:repo":    true,
				"admin:payment":  true,
				"system:command": true,
			},
		},
		"brasil": {
			AllowedScopes: map[string]bool{
				"delete:repo":    true,
				"create:repo":    true,
				"modify:design":  true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
				"manage:docker":  true,
				"system:command": true,
				"read:logs":      true,
				"read:metrics":   true,
			},
			RequiresApproval: map[string]bool{
				"delete:repo":    true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
			},
			EmergencyAllowed: map[string]bool{
				"delete:repo":   true,
				"admin:payment": true,
			},
		},
		"mocambique": {
			AllowedScopes: map[string]bool{
				"delete:repo":    true,
				"create:repo":    true,
				"modify:design":  true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
				"manage:docker":  true,
				"system:command": true,
				"read:logs":      true,
				"read:metrics":   true,
			},
			RequiresApproval: map[string]bool{
				"delete:repo":    true,
				"access:pii":     true,
				"admin:payment":  true,
				"transfer:money": true,
			},
			EmergencyAllowed: map[string]bool{
				"delete:repo":   true,
				"admin:payment": true,
			},
		},
	}
	
	rules, exists := marketRules[market]
	if !exists {
		return nil, ErrInvalidMarket
	}
	
	return rules, nil
}

// MarketScopeRules define as regras para escopos por mercado
type MarketScopeRules struct {
	AllowedScopes    map[string]bool // Escopos permitidos no mercado
	RequiresApproval map[string]bool // Escopos que exigem aprovação
	EmergencyAllowed map[string]bool // Escopos permitidos em modo emergência
}

// IsAllowedScope verifica se um escopo é permitido no mercado
func (r *MarketScopeRules) IsAllowedScope(scope string) bool {
	return r.AllowedScopes[scope]
}

// RequiresApprovalForScope verifica se um escopo exige aprovação
func (r *MarketScopeRules) RequiresApprovalForScope(scope string) bool {
	return r.RequiresApproval[scope]
}

// IsEmergencyAllowed verifica se um escopo pode ser usado em modo emergência
func (r *MarketScopeRules) IsEmergencyAllowed(scope string) bool {
	return r.EmergencyAllowed[scope]
}

// hasPermissionForScope verifica se o usuário tem permissão para solicitar um escopo
func (s *Service) hasPermissionForScope(ctx context.Context, userID, tenantID, scope string) (bool, error) {
	// Esta implementação seria integrada com o sistema de autenticação/autorização
	// Para fins de exemplo, retornamos uma verificação básica
	
	// TODO: Integrar com o serviço de autorização real
	if strings.HasPrefix(scope, "admin:") {
		// Verificação de administrador via serviço de tenant
		isAdmin, err := s.tenantService.IsUserAdmin(ctx, userID, tenantID)
		if err != nil {
			return false, err
		}
		return isAdmin, nil
	}
	
	// Por padrão, permitimos o usuário solicitar escopos não-admin
	return true, nil
}

// getUserActiveTokenLimit retorna o limite de tokens ativos por usuário para um mercado
func (s *Service) getUserActiveTokenLimit(market string) int {
	limits := map[string]int{
		"angola":     2, // BNA exige limite mais restritivo
		"brasil":     3, // LGPD permite um pouco mais de flexibilidade
		"mocambique": 3,
	}
	
	limit, exists := limits[market]
	if !exists {
		// Limite padrão para mercados não configurados
		return 1
	}
	
	return limit
}

// getMaxTokenDuration retorna a duração máxima de um token para um mercado e escopos
func (s *Service) getMaxTokenDuration(market string, scopes []string, emergency bool) int64 {
	// Duração máxima em minutos para tokens de elevação
	// Regulado por políticas específicas de cada mercado
	
	// Durações para operações de emergência (sempre mais curtas)
	emergencyDurations := map[string]int64{
		"angola":     30,  // 30 minutos
		"brasil":     60,  // 1 hora
		"mocambique": 60,  // 1 hora
	}
	
	// Durações padrão para operações normais
	standardDurations := map[string]int64{
		"angola":     240, // 4 horas
		"brasil":     480, // 8 horas
		"mocambique": 360, // 6 horas
	}
	
	// Durações para operações sensíveis específicas (não-emergência)
	sensitiveScopeDurations := map[string]map[string]int64{
		"angola": {
			"delete:repo":    60,  // 1 hora
			"access:pii":     30,  // 30 minutos
			"admin:payment":  60,  // 1 hora
			"transfer:money": 30,  // 30 minutos
		},
		"brasil": {
			"delete:repo":    120, // 2 horas
			"access:pii":     60,  // 1 hora
			"admin:payment":  120, // 2 horas
			"transfer:money": 60,  // 1 hora
		},
		"mocambique": {
			"delete:repo":    120, // 2 horas
			"access:pii":     60,  // 1 hora
			"admin:payment":  120, // 2 horas
			"transfer:money": 60,  // 1 hora
		},
	}
	
	if emergency {
		duration, exists := emergencyDurations[market]
		if exists {
			return duration
		}
		// Padrão para emergência em mercados não configurados: 30 minutos
		return 30
	}
	
	// Verifica se há escopos sensíveis com duração específica
	marketSensitiveDurations, exists := sensitiveScopeDurations[market]
	if exists {
		// Encontra a menor duração para todos os escopos solicitados
		var minDuration int64 = 1440 // 24 horas (máximo absoluto)
		foundSensitiveScope := false
		
		for _, scope := range scopes {
			if duration, scopeExists := marketSensitiveDurations[scope]; scopeExists {
				foundSensitiveScope = true
				if duration < minDuration {
					minDuration = duration
				}
			}
		}
		
		if foundSensitiveScope {
			return minDuration
		}
	}
	
	// Usa a duração padrão para o mercado
	duration, exists := standardDurations[market]
	if exists {
		return duration
	}
	
	// Duração padrão para mercados não configurados: 60 minutos
	return 60
}

// requiresApproval verifica se os escopos solicitados requerem aprovação
func (s *Service) requiresApproval(ctx context.Context, market string, scopes []string) bool {
	rules, err := s.getMarketScopeRules(ctx, market)
	if err != nil {
		// Em caso de erro, por segurança, assume que requer aprovação
		return true
	}
	
	// Se qualquer escopo requer aprovação, todo o token requer aprovação
	for _, scope := range scopes {
		if rules.RequiresApprovalForScope(scope) {
			return true
		}
	}
	
	return false
}

// notifyApprovers envia notificação aos aprovadores sobre nova solicitação
func (s *Service) notifyApprovers(ctx context.Context, token *Token) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.notifyApprovers")
	defer span.End()
	
	// Obtém lista de aprovadores com base no tenant e mercado
	approvers, err := s.tenantService.GetElevationApprovers(ctx, token.TenantID, token.Market)
	if err != nil {
		s.logger.Error("Erro ao obter aprovadores",
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market),
			zap.Error(err))
		return
	}
	
	// Formata mensagem de notificação
	message := fmt.Sprintf(
		"Nova solicitação de elevação de privilégios:\n"+
		"ID: %s\n"+
		"Usuário: %s\n"+
		"Tenant: %s\n"+
		"Mercado: %s\n"+
		"Escopos: %s\n"+
		"Justificativa: %s\n"+
		"Expira em: %s",
		token.ID,
		token.UserID,
		token.TenantID,
		token.Market,
		strings.Join(token.Scopes, ", "),
		token.Justification,
		token.ExpiresAt.Format(time.RFC3339),
	)
	
	// Notifica cada aprovador
	for _, approver := range approvers {
		err := s.notificationSvc.SendNotification(ctx, &notification.Notification{
			UserID:      approver.UserID,
			Type:        "elevation_request",
			Title:       "Nova solicitação de elevação",
			Message:     message,
			Priority:    "high",
			ActionURL:   fmt.Sprintf("/iam/elevation/requests/%s", token.ID),
			Metadata: map[string]interface{}{
				"elevation_id": token.ID,
				"tenant_id":    token.TenantID,
				"market":       token.Market,
			},
		})
		
		if err != nil {
			s.logger.Error("Erro ao notificar aprovador",
				zap.String("approver_id", approver.UserID),
				zap.String("elevation_id", token.ID),
				zap.Error(err))
		}
	}
}

// notifyEmergencyElevation notifica administradores sobre elevação de emergência
func (s *Service) notifyEmergencyElevation(ctx context.Context, token *Token) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.notifyEmergencyElevation")
	defer span.End()
	
	// Obtém lista de administradores para notificação de emergência
	admins, err := s.tenantService.GetEmergencyNotificationRecipients(ctx, token.TenantID, token.Market)
	if err != nil {
		s.logger.Error("Erro ao obter administradores para notificação de emergência",
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market),
			zap.Error(err))
		return
	}
	
	// Formata mensagem de alerta
	message := fmt.Sprintf(
		"⚠️ ALERTA: Elevação de emergência de privilégios:\n"+
		"ID: %s\n"+
		"Usuário: %s\n"+
		"Tenant: %s\n"+
		"Mercado: %s\n"+
		"Escopos: %s\n"+
		"Justificativa: %s\n"+
		"Expira em: %s",
		token.ID,
		token.UserID,
		token.TenantID,
		token.Market,
		strings.Join(token.Scopes, ", "),
		token.Justification,
		token.ExpiresAt.Format(time.RFC3339),
	)
	
	// Notifica cada administrador
	for _, admin := range admins {
		err := s.notificationSvc.SendNotification(ctx, &notification.Notification{
			UserID:      admin.UserID,
			Type:        "elevation_emergency",
			Title:       "⚠️ ALERTA: Elevação de emergência",
			Message:     message,
			Priority:    "critical",
			ActionURL:   fmt.Sprintf("/iam/elevation/emergency/%s", token.ID),
			Metadata: map[string]interface{}{
				"elevation_id": token.ID,
				"tenant_id":    token.TenantID,
				"market":       token.Market,
				"emergency":    true,
			},
		})
		
		if err != nil {
			s.logger.Error("Erro ao notificar administrador sobre elevação de emergência",
				zap.String("admin_id", admin.UserID),
				zap.String("elevation_id", token.ID),
				zap.Error(err))
		}
	}
}