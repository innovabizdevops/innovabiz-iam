// Package repositories define as interfaces de acesso a dados para o sistema IAM
// da plataforma INNOVABIZ.
package repositories

import (
	"context"
	"time"

	"github.com/innovabiz/iam/services/elevation"
)

// ElevationRepository define a interface para persistência de tokens de elevação
type ElevationRepository interface {
	// SaveToken persiste um novo token de elevação
	SaveToken(ctx context.Context, token *elevation.Token) error

	// UpdateToken atualiza um token de elevação existente
	UpdateToken(ctx context.Context, token *elevation.Token) error

	// GetToken recupera um token de elevação pelo ID
	GetToken(ctx context.Context, tokenID string) (*elevation.Token, error)

	// GetUserTokens recupera todos os tokens de um usuário
	GetUserTokens(ctx context.Context, userID, tenantID string) ([]*elevation.Token, error)

	// CountActiveTokensByUser conta tokens ativos de um usuário
	CountActiveTokensByUser(ctx context.Context, userID, tenantID string) (int, error)

	// ListPendingTokens recupera tokens pendentes de aprovação para um mercado e tenant
	ListPendingTokens(ctx context.Context, tenantID, market string) ([]*elevation.Token, error)

	// ListTokensByStatus recupera tokens por status
	ListTokensByStatus(ctx context.Context, tenantID, market string, status elevation.Status) ([]*elevation.Token, error)

	// ListExpiredTokens recupera tokens expirados que ainda não foram marcados como tal
	ListExpiredTokens(ctx context.Context) ([]*elevation.Token, error)

	// GetTokenHistory recupera o histórico de alterações de um token
	GetTokenHistory(ctx context.Context, tokenID string) ([]*elevation.TokenHistoryEntry, error)

	// GetTokenStats retorna estatísticas de tokens por mercado e tenant
	GetTokenStats(ctx context.Context, tenantID, market string, since time.Time) (*elevation.TokenStats, error)

	// DeleteToken exclui permanentemente um token (apenas para testes)
	DeleteToken(ctx context.Context, tokenID string) error
}

// TokenHistoryEntry representa uma entrada no histórico de um token
type TokenHistoryEntry struct {
	TokenID    string             `json:"token_id"`
	Timestamp  time.Time          `json:"timestamp"`
	Action     string             `json:"action"`
	ActorID    string             `json:"actor_id"`
	PrevStatus elevation.Status   `json:"prev_status"`
	NewStatus  elevation.Status   `json:"new_status"`
	Reason     string             `json:"reason,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TokenStats contém estatísticas sobre tokens de elevação
type TokenStats struct {
	TotalRequested        int `json:"total_requested"`
	TotalApproved         int `json:"total_approved"`
	TotalDenied           int `json:"total_denied"`
	TotalRevoked          int `json:"total_revoked"`
	TotalExpired          int `json:"total_expired"`
	TotalEmergency        int `json:"total_emergency"`
	AvgApprovalTimeMinutes int `json:"avg_approval_time_minutes"`
	AvgTokenUsageCount    int `json:"avg_token_usage_count"`
	CurrentActiveTokens   int `json:"current_active_tokens"`
	CurrentPendingTokens  int `json:"current_pending_tokens"`
}