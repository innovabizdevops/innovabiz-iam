// Package elevation implementa o serviço core de elevação de privilégios para o sistema MCP-IAM
// da plataforma INNOVABIZ. Este serviço gerencia o ciclo de vida completo de tokens de elevação,
// incluindo solicitação, aprovação, uso e revogação, respeitando políticas por mercado e tenant.
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
	"github.com/innovabiz/iam/logging"
	"github.com/innovabiz/iam/metrics"
	"github.com/innovabiz/iam/models"
	"github.com/innovabiz/iam/repositories"
	"github.com/innovabiz/iam/services/tenant"
	"github.com/innovabiz/iam/services/notification"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Status define os possíveis estados de um token de elevação
type Status string

const (
	// StatusPendingApproval indica que o token está aguardando aprovação
	StatusPendingApproval Status = "pending_approval"
	
	// StatusActive indica que o token está ativo e pode ser usado
	StatusActive Status = "active"
	
	// StatusExpired indica que o token expirou
	StatusExpired Status = "expired"
	
	// StatusRevoked indica que o token foi revogado
	StatusRevoked Status = "revoked"
	
	// StatusDenied indica que a solicitação de elevação foi negada
	StatusDenied Status = "denied"
)

// ErrTokenExpired é retornado quando um token expirou
var ErrTokenExpired = errors.New("elevation token expired")

// ErrTokenRevoked é retornado quando um token foi revogado
var ErrTokenRevoked = errors.New("elevation token revoked")

// ErrTokenDenied é retornado quando um token foi negado
var ErrTokenDenied = errors.New("elevation token denied")

// ErrInvalidMarket é retornado quando o mercado é inválido ou não configurado
var ErrInvalidMarket = errors.New("invalid or unconfigured market")

// ErrMFARequired é retornado quando MFA é necessário mas não foi completado
var ErrMFARequired = errors.New("MFA verification required for elevation")

// ErrTenantMismatch é retornado quando há incompatibilidade de tenant
var ErrTenantMismatch = errors.New("tenant mismatch in elevation token")

// ErrScopeNotAllowed é retornado quando o escopo solicitado não é permitido
var ErrScopeNotAllowed = errors.New("requested scope not allowed for user/tenant")

// ErrInsufficientPermission é retornado quando o usuário não tem permissão
var ErrInsufficientPermission = errors.New("insufficient permission for elevation action")

// ElevationRequest contém os dados necessários para solicitar elevação
type ElevationRequest struct {
	UserID        string   `json:"user_id"`
	TenantID      string   `json:"tenant_id"`
	Market        string   `json:"market"`
	Scopes        []string `json:"scopes"`
	Justification string   `json:"justification"`
	Duration      int64    `json:"duration"` // Duração em minutos
	Emergency     bool     `json:"emergency"` // Indica solicitação de emergência
}

// Token representa um token de elevação de privilégios
type Token struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	TenantID      string    `json:"tenant_id"`
	Market        string    `json:"market"`
	Scopes        []string  `json:"scopes"`
	Status        Status    `json:"status"`
	Justification string    `json:"justification"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	ApprovedBy    string    `json:"approved_by,omitempty"`
	ApprovedAt    time.Time `json:"approved_at,omitempty"`
	DeniedBy      string    `json:"denied_by,omitempty"`
	DeniedAt      time.Time `json:"denied_at,omitempty"`
	RevokedBy     string    `json:"revoked_by,omitempty"`
	RevokedAt     time.Time `json:"revoked_at,omitempty"`
	Emergency     bool      `json:"emergency"`
	RevokeReason  string    `json:"revoke_reason,omitempty"`
	DenyReason    string    `json:"deny_reason,omitempty"`
}

// IsExpired verifica se o token expirou
func (t *Token) IsExpired() bool {
	return t.Status == StatusExpired || time.Now().After(t.ExpiresAt)
}

// IsRevoked verifica se o token foi revogado
func (t *Token) IsRevoked() bool {
	return t.Status == StatusRevoked
}

// IsDenied verifica se o token foi negado
func (t *Token) IsDenied() bool {
	return t.Status == StatusDenied
}

// IsActive verifica se o token está ativo
func (t *Token) IsActive() bool {
	if t.Status != StatusActive {
		return false
	}
	return !t.IsExpired()
}

// Service implementa o serviço de elevação de privilégios
type Service struct {
	config           *config.ElevationConfig
	repository       repositories.ElevationRepository
	auditService     audit.AuditService
	tenantService    tenant.Service
	notificationSvc  notification.Service
	logger           *zap.Logger
	tracer           trace.Tracer
	activeTokens     prometheus.Gauge
	elevationCounter prometheus.CounterVec
}

// NewService cria uma nova instância do serviço de elevação
func NewService(
	cfg *config.ElevationConfig,
	repo repositories.ElevationRepository,
	auditSvc audit.AuditService,
	tenantSvc tenant.Service,
	notifSvc notification.Service,
) *Service {
	logger := logging.GetLogger().Named("elevation-service")
	
	// Métricas para monitoramento
	activeTokens := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "innovabiz",
		Subsystem: "iam",
		Name:      "elevation_active_tokens",
		Help:      "Número de tokens de elevação ativos",
	})
	
	elevationCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "innovabiz",
			Subsystem: "iam",
			Name:      "elevation_operations_total",
			Help:      "Total de operações de elevação por tipo e status",
		},
		[]string{"operation", "status", "market"},
	)
	
	metrics.GetRegistry().MustRegister(activeTokens, elevationCounter)
	
	return &Service{
		config:           cfg,
		repository:       repo,
		auditService:     auditSvc,
		tenantService:    tenantSvc,
		notificationSvc:  notifSvc,
		logger:           logger,
		tracer:           otel.Tracer("innovabiz/iam/elevation"),
		activeTokens:     activeTokens,
		elevationCounter: elevationCounter,
	}
}

// GetAuditService retorna o serviço de auditoria (útil para testes)
func (s *Service) GetAuditService() audit.AuditService {
	return s.auditService
}

// RequestElevation processa uma solicitação de elevação de privilégios
func (s *Service) RequestElevation(ctx context.Context, request *ElevationRequest) (*Token, error) {
	ctx, span := s.tracer.Start(ctx, "ElevationService.RequestElevation",
		trace.WithAttributes(
			attribute.String("user_id", request.UserID),
			attribute.String("tenant_id", request.TenantID),
			attribute.String("market", request.Market),
			attribute.Int64("duration", request.Duration),
			attribute.Bool("emergency", request.Emergency),
		))
	defer span.End()
	
	// Recupera token de autenticação do contexto
	authToken := auth.TokenFromContext(ctx)
	if authToken == nil {
		s.logger.Error("Token de autenticação não encontrado no contexto")
		return nil, errors.New("authentication token not found in context")
	}
	
	// Verifica compatibilidade de tenant e mercado
	if authToken.TenantID != request.TenantID {
		s.logger.Error("Incompatibilidade de tenant",
			zap.String("token_tenant", authToken.TenantID),
			zap.String("request_tenant", request.TenantID))
		return nil, ErrTenantMismatch
	}
	
	// Verifica MFA conforme políticas do mercado
	if s.requiresMFA(ctx, request.Market, request.Scopes) && !authToken.MFACompleted {
		s.logger.Warn("Tentativa de elevação sem MFA",
			zap.String("user_id", request.UserID),
			zap.String("tenant_id", request.TenantID),
			zap.String("market", request.Market))
		
		s.elevationCounter.WithLabelValues("request", "mfa_required", request.Market).Inc()
		return nil, ErrMFARequired
	}
	
	// Valida scopes de elevação solicitados
	if err := s.validateScopes(ctx, request); err != nil {
		s.logger.Warn("Escopo de elevação inválido",
			zap.String("user_id", request.UserID),
			zap.Strings("scopes", request.Scopes),
			zap.Error(err))
		
		s.elevationCounter.WithLabelValues("request", "invalid_scope", request.Market).Inc()
		return nil, err
	}
	
	// Verifica limite de tokens ativos por usuário
	activeTokenCount, err := s.repository.CountActiveTokensByUser(ctx, request.UserID, request.TenantID)
	if err != nil {
		s.logger.Error("Erro ao contar tokens ativos",
			zap.String("user_id", request.UserID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to count active tokens: %w", err)
	}
	
	userLimit := s.getUserActiveTokenLimit(request.Market)
	if activeTokenCount >= userLimit {
		s.logger.Warn("Limite de tokens ativos excedido",
			zap.String("user_id", request.UserID),
			zap.Int("active_count", activeTokenCount),
			zap.Int("limit", userLimit))
		
		s.elevationCounter.WithLabelValues("request", "limit_exceeded", request.Market).Inc()
		return nil, fmt.Errorf("active token limit (%d) exceeded", userLimit)
	}
	
	// Define duração máxima baseada nas políticas do mercado
	maxDuration := s.getMaxTokenDuration(request.Market, request.Scopes, request.Emergency)
	if request.Duration > maxDuration {
		request.Duration = maxDuration
	}
	
	// Cria token de elevação
	token := &Token{
		ID:            uuid.New().String(),
		UserID:        request.UserID,
		TenantID:      request.TenantID,
		Market:        request.Market,
		Scopes:        request.Scopes,
		Status:        StatusPendingApproval,
		Justification: request.Justification,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(time.Duration(request.Duration) * time.Minute),
		Emergency:     request.Emergency,
	}
	
	// Verifica se é emergência (auto-aprovação)
	if request.Emergency {
		// Tokens de emergência são automaticamente aprovados, mas auditados especialmente
		token.Status = StatusActive
		token.ApprovedAt = time.Now()
		token.ApprovedBy = fmt.Sprintf("emergency-auto-%s", request.UserID)
		
		// Registra evento especial de aprovação automática de emergência
		s.auditService.LogEvent(ctx, &audit.Event{
			Type:      "elevation_emergency_auto_approved",
			UserID:    request.UserID,
			TenantID:  request.TenantID,
			Market:    request.Market,
			Timestamp: time.Now(),
			ClientIP:  authToken.ClientIP,
			Metadata: map[string]interface{}{
				"elevation_id": token.ID,
				"scopes":       request.Scopes,
				"justification": request.Justification,
				"duration":     request.Duration,
				"user_agent":   authToken.UserAgent,
			},
		})
		
		// Notifica administradores sobre elevação de emergência
		s.notifyEmergencyElevation(ctx, token)
	} else {
		// Verifica se requer aprovação conforme políticas do mercado
		if s.requiresApproval(ctx, request.Market, request.Scopes) {
			token.Status = StatusPendingApproval
			
			// Notifica aprovadores sobre nova solicitação
			s.notifyApprovers(ctx, token)
		} else {
			// Auto-aprovação para escopos não sensíveis
			token.Status = StatusActive
			token.ApprovedAt = time.Now()
			token.ApprovedBy = "auto-approved"
		}
	}
	
	// Persiste token no repositório
	if err := s.repository.SaveToken(ctx, token); err != nil {
		s.logger.Error("Erro ao salvar token de elevação",
			zap.String("user_id", request.UserID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to save elevation token: %w", err)
	}
	
	// Atualiza métricas
	if token.Status == StatusActive {
		s.activeTokens.Inc()
	}
	
	// Registra evento de auditoria
	s.auditService.LogEvent(ctx, &audit.Event{
		Type:      "elevation_requested",
		UserID:    request.UserID,
		TenantID:  request.TenantID,
		Market:    request.Market,
		Timestamp: time.Now(),
		ClientIP:  authToken.ClientIP,
		Metadata: map[string]interface{}{
			"elevation_id": token.ID,
			"scopes":       request.Scopes,
			"justification": request.Justification,
			"status":       string(token.Status),
			"duration":     request.Duration,
			"emergency":    request.Emergency,
		},
	})
	
	s.elevationCounter.WithLabelValues("request", string(token.Status), request.Market).Inc()
	
	s.logger.Info("Token de elevação criado",
		zap.String("elevation_id", token.ID),
		zap.String("user_id", request.UserID),
		zap.String("status", string(token.Status)),
		zap.Strings("scopes", request.Scopes))
	
	return token, nil
}

// ContextWithElevationToken adiciona um token de elevação ao contexto
func ContextWithElevationToken(ctx context.Context, token *Token) context.Context {
	return context.WithValue(ctx, constants.ElevationTokenContextKey, token)
}

// TokenFromContext recupera um token de elevação do contexto
func TokenFromContext(ctx context.Context) *Token {
	token, _ := ctx.Value(constants.ElevationTokenContextKey).(*Token)
	return token
}