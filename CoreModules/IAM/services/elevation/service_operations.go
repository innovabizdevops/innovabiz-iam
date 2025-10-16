// Package elevation implementa o serviço core de elevação de privilégios para o sistema MCP-IAM
// da plataforma INNOVABIZ.
package elevation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/innovabiz/iam/audit"
	"github.com/innovabiz/iam/auth"
	"github.com/innovabiz/iam/config"
	"github.com/innovabiz/iam/constants"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ApproveElevation aprova uma solicitação de elevação
func (s *Service) ApproveElevation(ctx context.Context, tokenID, approverID string) (*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.ApproveElevation",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
			attribute.String("approver_id", approverID),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		s.logger.Error("Token de autenticação não encontrado no contexto")
		return nil, errors.New("authentication token not found in context")
	}
	
	// Recupera token de elevação do repositório
	token, err := s.repository.GetToken(ctx, tokenID)
	if err != nil {
		s.logger.Error("Erro ao recuperar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve elevation token: %w", err)
	}
	
	// Verifica se o token já está aprovado, expirado, negado ou revogado
	if token.Status != StatusPendingApproval {
		s.logger.Warn("Tentativa de aprovar token com status inválido",
			zap.String("elevation_id", tokenID),
			zap.String("status", string(token.Status)))
		
		s.elevationCounter.WithLabelValues("approve", "invalid_status", token.Market).Inc()
		return nil, fmt.Errorf("cannot approve token with status %s", token.Status)
	}
	
	// Verifica se o token já expirou
	if token.IsExpired() {
		token.Status = StatusExpired
		_ = s.repository.UpdateToken(ctx, token)
		
		s.logger.Warn("Tentativa de aprovar token expirado",
			zap.String("elevation_id", tokenID))
			
		s.elevationCounter.WithLabelValues("approve", "expired", token.Market).Inc()
		return nil, ErrTokenExpired
	}
	
	// Verifica se o aprovador tem permissão
	isApprover, err := s.tenantService.IsElevationApprover(ctx, approverID, token.TenantID, token.Market)
	if err != nil {
		s.logger.Error("Erro ao verificar permissão do aprovador",
			zap.String("approver_id", approverID),
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market),
			zap.Error(err))
		return nil, fmt.Errorf("failed to check approver permission: %w", err)
	}
	
	if !isApprover {
		s.logger.Warn("Usuário não autorizado para aprovar elevação",
			zap.String("approver_id", approverID),
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market))
			
		s.elevationCounter.WithLabelValues("approve", "unauthorized", token.Market).Inc()
		return nil, ErrInsufficientPermission
	}
	
	// Atualiza token para aprovado
	token.Status = StatusActive
	token.ApprovedAt = time.Now()
	token.ApprovedBy = approverID
	
	if err := s.repository.UpdateToken(ctx, token); err != nil {
		s.logger.Error("Erro ao atualizar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update elevation token: %w", err)
	}
	
	// Atualiza métricas
	s.activeTokens.Inc()
	
	// Registra evento de auditoria
	s.auditService.LogEvent(ctx, &audit.Event{
		Type:      "elevation_approved",
		UserID:    approverID,
		TenantID:  token.TenantID,
		Market:    token.Market,
		Timestamp: time.Now(),
		ClientIP:  authToken.ClientIP,
		Metadata: map[string]interface{}{
			"elevation_id":   token.ID,
			"requester_id":   token.UserID,
			"scopes":         token.Scopes,
			"justification":  token.Justification,
			"request_time":   token.CreatedAt,
			"expiration_time": token.ExpiresAt,
		},
	})
	
	s.elevationCounter.WithLabelValues("approve", "success", token.Market).Inc()
	
	// Notifica o solicitante sobre a aprovação
	s.notifySolicitante(ctx, token, "approved")
	
	s.logger.Info("Token de elevação aprovado",
		zap.String("elevation_id", token.ID),
		zap.String("approver_id", approverID),
		zap.String("user_id", token.UserID))
	
	return token, nil
}

// DenyElevation nega uma solicitação de elevação
func (s *Service) DenyElevation(ctx context.Context, tokenID, denyingUserID, reason string) (*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.DenyElevation",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
			attribute.String("denying_user_id", denyingUserID),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		s.logger.Error("Token de autenticação não encontrado no contexto")
		return nil, errors.New("authentication token not found in context")
	}
	
	// Recupera token de elevação do repositório
	token, err := s.repository.GetToken(ctx, tokenID)
	if err != nil {
		s.logger.Error("Erro ao recuperar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve elevation token: %w", err)
	}
	
	// Verifica se o token está pendente de aprovação
	if token.Status != StatusPendingApproval {
		s.logger.Warn("Tentativa de negar token com status inválido",
			zap.String("elevation_id", tokenID),
			zap.String("status", string(token.Status)))
			
		s.elevationCounter.WithLabelValues("deny", "invalid_status", token.Market).Inc()
		return nil, fmt.Errorf("cannot deny token with status %s", token.Status)
	}
	
	// Verifica se o usuário tem permissão para negar
	isApprover, err := s.tenantService.IsElevationApprover(ctx, denyingUserID, token.TenantID, token.Market)
	if err != nil {
		s.logger.Error("Erro ao verificar permissão do usuário",
			zap.String("user_id", denyingUserID),
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market),
			zap.Error(err))
		return nil, fmt.Errorf("failed to check user permission: %w", err)
	}
	
	if !isApprover {
		s.logger.Warn("Usuário não autorizado para negar elevação",
			zap.String("user_id", denyingUserID),
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market))
			
		s.elevationCounter.WithLabelValues("deny", "unauthorized", token.Market).Inc()
		return nil, ErrInsufficientPermission
	}
	
	// Atualiza token para negado
	token.Status = StatusDenied
	token.DeniedAt = time.Now()
	token.DeniedBy = denyingUserID
	token.DenyReason = reason
	
	if err := s.repository.UpdateToken(ctx, token); err != nil {
		s.logger.Error("Erro ao atualizar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update elevation token: %w", err)
	}
	
	// Registra evento de auditoria
	s.auditService.LogEvent(ctx, &audit.Event{
		Type:      "elevation_denied",
		UserID:    denyingUserID,
		TenantID:  token.TenantID,
		Market:    token.Market,
		Timestamp: time.Now(),
		ClientIP:  authToken.ClientIP,
		Metadata: map[string]interface{}{
			"elevation_id":  token.ID,
			"requester_id":  token.UserID,
			"scopes":        token.Scopes,
			"justification": token.Justification,
			"request_time":  token.CreatedAt,
			"deny_reason":   reason,
		},
	})
	
	s.elevationCounter.WithLabelValues("deny", "success", token.Market).Inc()
	
	// Notifica o solicitante sobre a negação
	s.notifySolicitante(ctx, token, "denied")
	
	s.logger.Info("Token de elevação negado",
		zap.String("elevation_id", token.ID),
		zap.String("denying_user_id", denyingUserID),
		zap.String("user_id", token.UserID),
		zap.String("reason", reason))
	
	return token, nil
}

// RevokeElevation revoga um token de elevação ativo
func (s *Service) RevokeElevation(ctx context.Context, tokenID, revokingUserID, reason string) (*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.RevokeElevation",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
			attribute.String("revoking_user_id", revokingUserID),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		s.logger.Error("Token de autenticação não encontrado no contexto")
		return nil, errors.New("authentication token not found in context")
	}
	
	// Recupera token de elevação do repositório
	token, err := s.repository.GetToken(ctx, tokenID)
	if err != nil {
		s.logger.Error("Erro ao recuperar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve elevation token: %w", err)
	}
	
	// Verifica se o token está ativo
	if token.Status != StatusActive {
		s.logger.Warn("Tentativa de revogar token com status inválido",
			zap.String("elevation_id", tokenID),
			zap.String("status", string(token.Status)))
			
		s.elevationCounter.WithLabelValues("revoke", "invalid_status", token.Market).Inc()
		return nil, fmt.Errorf("cannot revoke token with status %s", token.Status)
	}
	
	// Verifica se o usuário tem permissão para revogar
	// Permissão concedida se for: o próprio usuário, um aprovador ou um administrador
	isSelfRevoke := revokingUserID == token.UserID
	isApprover, err := s.tenantService.IsElevationApprover(ctx, revokingUserID, token.TenantID, token.Market)
	if err != nil {
		s.logger.Error("Erro ao verificar permissão do usuário",
			zap.String("user_id", revokingUserID),
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market),
			zap.Error(err))
		return nil, fmt.Errorf("failed to check user permission: %w", err)
	}
	
	isAdmin, err := s.tenantService.IsUserAdmin(ctx, revokingUserID, token.TenantID)
	if err != nil {
		s.logger.Error("Erro ao verificar status de administrador",
			zap.String("user_id", revokingUserID),
			zap.String("tenant_id", token.TenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to check admin status: %w", err)
	}
	
	if !isSelfRevoke && !isApprover && !isAdmin {
		s.logger.Warn("Usuário não autorizado para revogar elevação",
			zap.String("user_id", revokingUserID),
			zap.String("tenant_id", token.TenantID),
			zap.String("market", token.Market))
			
		s.elevationCounter.WithLabelValues("revoke", "unauthorized", token.Market).Inc()
		return nil, ErrInsufficientPermission
	}
	
	// Atualiza token para revogado
	token.Status = StatusRevoked
	token.RevokedAt = time.Now()
	token.RevokedBy = revokingUserID
	token.RevokeReason = reason
	
	if err := s.repository.UpdateToken(ctx, token); err != nil {
		s.logger.Error("Erro ao atualizar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update elevation token: %w", err)
	}
	
	// Atualiza métricas
	s.activeTokens.Dec()
	
	// Registra evento de auditoria
	s.auditService.LogEvent(ctx, &audit.Event{
		Type:      "elevation_revoked",
		UserID:    revokingUserID,
		TenantID:  token.TenantID,
		Market:    token.Market,
		Timestamp: time.Now(),
		ClientIP:  authToken.ClientIP,
		Metadata: map[string]interface{}{
			"elevation_id":   token.ID,
			"requester_id":   token.UserID,
			"scopes":         token.Scopes,
			"revoke_reason":  reason,
			"self_revoke":    isSelfRevoke,
			"admin_revoke":   isAdmin,
			"approver_revoke": isApprover && !isAdmin,
			"request_time":   token.CreatedAt,
			"approved_time":  token.ApprovedAt,
			"expiration_time": token.ExpiresAt,
		},
	})
	
	s.elevationCounter.WithLabelValues("revoke", "success", token.Market).Inc()
	
	// Se não for auto-revogação, notifica o solicitante
	if !isSelfRevoke {
		s.notifySolicitante(ctx, token, "revoked")
	}
	
	s.logger.Info("Token de elevação revogado",
		zap.String("elevation_id", token.ID),
		zap.String("revoking_user_id", revokingUserID),
		zap.String("user_id", token.UserID),
		zap.String("reason", reason))
	
	return token, nil
}

// ValidateTokenForScope valida se um token permite o uso de um determinado escopo
func (s *Service) ValidateTokenForScope(ctx context.Context, tokenID, scope, hookType, operation string) (*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.ValidateTokenForScope",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
			attribute.String("scope", scope),
			attribute.String("hook_type", hookType),
			attribute.String("operation", operation),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		s.logger.Error("Token de autenticação não encontrado no contexto")
		return nil, errors.New("authentication token not found in context")
	}
	
	// Recupera token de elevação do repositório
	token, err := s.repository.GetToken(ctx, tokenID)
	if err != nil {
		s.logger.Error("Erro ao recuperar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve elevation token: %w", err)
	}
	
	// Verifica se o token está ativo
	if token.Status != StatusActive {
		s.logger.Warn("Token de elevação não está ativo",
			zap.String("elevation_id", tokenID),
			zap.String("status", string(token.Status)))
			
		s.elevationCounter.WithLabelValues("validate", "inactive", token.Market).Inc()
		
		if token.Status == StatusRevoked {
			return nil, ErrTokenRevoked
		} else if token.Status == StatusDenied {
			return nil, ErrTokenDenied
		} else if token.Status == StatusExpired || token.IsExpired() {
			// Atualiza status para expirado se ainda não foi feito
			if token.Status != StatusExpired {
				token.Status = StatusExpired
				_ = s.repository.UpdateToken(ctx, token)
			}
			return nil, ErrTokenExpired
		} else {
			return nil, fmt.Errorf("token is not active, current status: %s", token.Status)
		}
	}
	
	// Verifica se o token expirou
	if token.IsExpired() {
		token.Status = StatusExpired
		_ = s.repository.UpdateToken(ctx, token)
		
		s.logger.Warn("Token de elevação expirado",
			zap.String("elevation_id", tokenID))
			
		s.elevationCounter.WithLabelValues("validate", "expired", token.Market).Inc()
		return nil, ErrTokenExpired
	}
	
	// Verifica compatibilidade de tenant
	if authToken.TenantID != token.TenantID {
		s.logger.Warn("Incompatibilidade de tenant na validação do token",
			zap.String("token_tenant", token.TenantID),
			zap.String("auth_tenant", authToken.TenantID))
			
		s.elevationCounter.WithLabelValues("validate", "tenant_mismatch", token.Market).Inc()
		return nil, ErrTenantMismatch
	}
	
	// Verifica compatibilidade de usuário
	if authToken.UserID != token.UserID {
		s.logger.Warn("Incompatibilidade de usuário na validação do token",
			zap.String("token_user", token.UserID),
			zap.String("auth_user", authToken.UserID))
			
		s.elevationCounter.WithLabelValues("validate", "user_mismatch", token.Market).Inc()
		return nil, errors.New("user mismatch in elevation token")
	}
	
	// Verifica se o escopo está incluído no token
	hasScope := false
	for _, tokenScope := range token.Scopes {
		if tokenScope == scope {
			hasScope = true
			break
		}
	}
	
	if !hasScope {
		s.logger.Warn("Escopo não incluído no token de elevação",
			zap.String("elevation_id", tokenID),
			zap.String("requested_scope", scope),
			zap.Strings("token_scopes", token.Scopes))
			
		s.elevationCounter.WithLabelValues("validate", "scope_not_included", token.Market).Inc()
		return nil, fmt.Errorf("scope %s not included in elevation token", scope)
	}
	
	// Registra evento de uso do token de elevação
	s.auditService.LogEvent(ctx, &audit.Event{
		Type:      "elevation_used",
		UserID:    token.UserID,
		TenantID:  token.TenantID,
		Market:    token.Market,
		Timestamp: time.Now(),
		ClientIP:  authToken.ClientIP,
		Metadata: map[string]interface{}{
			"elevation_id": token.ID,
			"scope":        scope,
			"hook_type":    hookType,
			"operation":    operation,
			"user_agent":   authToken.UserAgent,
		},
	})
	
	s.elevationCounter.WithLabelValues("validate", "success", token.Market).Inc()
	
	s.logger.Info("Token de elevação validado com sucesso",
		zap.String("elevation_id", token.ID),
		zap.String("user_id", token.UserID),
		zap.String("scope", scope),
		zap.String("hook_type", hookType),
		zap.String("operation", operation))
	
	return token, nil
}

// GetToken recupera um token de elevação pelo ID
func (s *Service) GetToken(ctx context.Context, tokenID string) (*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.GetToken",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
		))
	defer span.End()
	
	token, err := s.repository.GetToken(ctx, tokenID)
	if err != nil {
		s.logger.Error("Erro ao recuperar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve elevation token: %w", err)
	}
	
	// Verifica se o token expirou mas não foi marcado como expirado
	if token.Status == StatusActive && token.IsExpired() {
		token.Status = StatusExpired
		_ = s.repository.UpdateToken(ctx, token)
	}
	
	return token, nil
}

// ListUserTokens lista todos os tokens de um usuário
func (s *Service) ListUserTokens(ctx context.Context, userID, tenantID string) ([]*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.ListUserTokens",
		trace.WithAttributes(
			attribute.String("user_id", userID),
			attribute.String("tenant_id", tenantID),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		s.logger.Error("Token de autenticação não encontrado no contexto")
		return nil, errors.New("authentication token not found in context")
	}
	
	// Verifica se é o próprio usuário ou um administrador
	isSelfCheck := userID == authToken.UserID
	isAdmin := false
	
	if !isSelfCheck {
		var err error
		isAdmin, err = s.tenantService.IsUserAdmin(ctx, authToken.UserID, tenantID)
		if err != nil {
			s.logger.Error("Erro ao verificar status de administrador",
				zap.String("user_id", authToken.UserID),
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to check admin status: %w", err)
		}
		
		if !isAdmin {
			s.logger.Warn("Usuário não autorizado a listar tokens de outro usuário",
				zap.String("requesting_user", authToken.UserID),
				zap.String("target_user", userID))
			return nil, ErrInsufficientPermission
		}
	}
	
	// Recupera tokens do repositório
	tokens, err := s.repository.GetUserTokens(ctx, userID, tenantID)
	if err != nil {
		s.logger.Error("Erro ao recuperar tokens do usuário",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve user tokens: %w", err)
	}
	
	// Atualiza status de tokens expirados
	for _, token := range tokens {
		if token.Status == StatusActive && token.IsExpired() {
			token.Status = StatusExpired
			_ = s.repository.UpdateToken(ctx, token)
		}
	}
	
	return tokens, nil
}

// notifySolicitante notifica o solicitante sobre mudanças no token
func (s *Service) notifySolicitante(ctx context.Context, token *Token, action string) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.notifySolicitante",
		trace.WithAttributes(
			attribute.String("action", action),
			attribute.String("elevation_id", token.ID),
		))
	defer span.End()
	
	var title, message, notificationType string
	
	switch action {
	case "approved":
		title = "Solicitação de elevação aprovada"
		message = fmt.Sprintf(
			"Sua solicitação de elevação de privilégios foi aprovada:\n"+
			"ID: %s\n"+
			"Escopos: %s\n"+
			"Expira em: %s\n"+
			"Aprovado por: %s",
			token.ID,
			strings.Join(token.Scopes, ", "),
			token.ExpiresAt.Format(time.RFC3339),
			token.ApprovedBy,
		)
		notificationType = "elevation_approved"
	case "denied":
		title = "Solicitação de elevação negada"
		message = fmt.Sprintf(
			"Sua solicitação de elevação de privilégios foi negada:\n"+
			"ID: %s\n"+
			"Escopos: %s\n"+
			"Negado por: %s\n"+
			"Motivo: %s",
			token.ID,
			strings.Join(token.Scopes, ", "),
			token.DeniedBy,
			token.DenyReason,
		)
		notificationType = "elevation_denied"
	case "revoked":
		title = "Token de elevação revogado"
		message = fmt.Sprintf(
			"Seu token de elevação de privilégios foi revogado:\n"+
			"ID: %s\n"+
			"Escopos: %s\n"+
			"Revogado por: %s\n"+
			"Motivo: %s",
			token.ID,
			strings.Join(token.Scopes, ", "),
			token.RevokedBy,
			token.RevokeReason,
		)
		notificationType = "elevation_revoked"
	default:
		s.logger.Error("Tipo de notificação desconhecido", zap.String("action", action))
		return
	}
	
	err := s.notificationSvc.SendNotification(ctx, &notification.Notification{
		UserID:  token.UserID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Priority: action == "revoked" ? "high" : "normal",
		ActionURL: fmt.Sprintf("/iam/elevation/tokens/%s", token.ID),
		Metadata: map[string]interface{}{
			"elevation_id": token.ID,
			"tenant_id":    token.TenantID,
			"market":       token.Market,
			"action":       action,
		},
	})
	
	if err != nil {
		s.logger.Error("Erro ao enviar notificação ao solicitante",
			zap.String("user_id", token.UserID),
			zap.String("elevation_id", token.ID),
			zap.String("action", action),
			zap.Error(err))
	}
}