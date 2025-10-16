// Package repositories implementa os repositórios de acesso a dados
// para o sistema IAM da plataforma INNOVABIZ.
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/innovabiz/iam/logging"
	"github.com/innovabiz/iam/services/elevation"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// PostgresElevationRepository implementa ElevationRepository para PostgreSQL
type PostgresElevationRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
	tracer trace.Tracer
}

// NewPostgresElevationRepository cria uma nova instância de PostgresElevationRepository
func NewPostgresElevationRepository(db *sqlx.DB) *PostgresElevationRepository {
	return &PostgresElevationRepository{
		db:     db,
		logger: logging.GetLogger().Named("elevation-repository"),
		tracer: otel.Tracer("innovabiz/iam/repositories/elevation"),
	}
}

// dbToken é a representação do token na base de dados
type dbToken struct {
	ID            string         `db:"id"`
	UserID        string         `db:"user_id"`
	TenantID      string         `db:"tenant_id"`
	Market        string         `db:"market"`
	Scopes        pq.StringArray `db:"scopes"`
	Status        string         `db:"status"`
	Justification string         `db:"justification"`
	CreatedAt     time.Time      `db:"created_at"`
	ExpiresAt     time.Time      `db:"expires_at"`
	ApprovedBy    sql.NullString `db:"approved_by"`
	ApprovedAt    sql.NullTime   `db:"approved_at"`
	DeniedBy      sql.NullString `db:"denied_by"`
	DeniedAt      sql.NullTime   `db:"denied_at"`
	RevokedBy     sql.NullString `db:"revoked_by"`
	RevokedAt     sql.NullTime   `db:"revoked_at"`
	Emergency     bool           `db:"emergency"`
	RevokeReason  sql.NullString `db:"revoke_reason"`
	DenyReason    sql.NullString `db:"deny_reason"`
	UsageCount    int            `db:"usage_count"`
	LastUsedAt    sql.NullTime   `db:"last_used_at"`
}

// toServiceToken converte um dbToken para um elevation.Token
func (dt *dbToken) toServiceToken() *elevation.Token {
	token := &elevation.Token{
		ID:            dt.ID,
		UserID:        dt.UserID,
		TenantID:      dt.TenantID,
		Market:        dt.Market,
		Scopes:        []string(dt.Scopes),
		Status:        elevation.Status(dt.Status),
		Justification: dt.Justification,
		CreatedAt:     dt.CreatedAt,
		ExpiresAt:     dt.ExpiresAt,
		Emergency:     dt.Emergency,
	}

	if dt.ApprovedBy.Valid {
		token.ApprovedBy = dt.ApprovedBy.String
	}

	if dt.ApprovedAt.Valid {
		token.ApprovedAt = dt.ApprovedAt.Time
	}

	if dt.DeniedBy.Valid {
		token.DeniedBy = dt.DeniedBy.String
	}

	if dt.DeniedAt.Valid {
		token.DeniedAt = dt.DeniedAt.Time
	}

	if dt.RevokedBy.Valid {
		token.RevokedBy = dt.RevokedBy.String
	}

	if dt.RevokedAt.Valid {
		token.RevokedAt = dt.RevokedAt.Time
	}

	if dt.RevokeReason.Valid {
		token.RevokeReason = dt.RevokeReason.String
	}

	if dt.DenyReason.Valid {
		token.DenyReason = dt.DenyReason.String
	}

	return token
}

// fromServiceToken converte um elevation.Token para um dbToken
func fromServiceToken(t *elevation.Token) *dbToken {
	dt := &dbToken{
		ID:            t.ID,
		UserID:        t.UserID,
		TenantID:      t.TenantID,
		Market:        t.Market,
		Scopes:        pq.StringArray(t.Scopes),
		Status:        string(t.Status),
		Justification: t.Justification,
		CreatedAt:     t.CreatedAt,
		ExpiresAt:     t.ExpiresAt,
		Emergency:     t.Emergency,
	}

	if t.ApprovedBy != "" {
		dt.ApprovedBy = sql.NullString{String: t.ApprovedBy, Valid: true}
	}

	if !t.ApprovedAt.IsZero() {
		dt.ApprovedAt = sql.NullTime{Time: t.ApprovedAt, Valid: true}
	}

	if t.DeniedBy != "" {
		dt.DeniedBy = sql.NullString{String: t.DeniedBy, Valid: true}
	}

	if !t.DeniedAt.IsZero() {
		dt.DeniedAt = sql.NullTime{Time: t.DeniedAt, Valid: true}
	}

	if t.RevokedBy != "" {
		dt.RevokedBy = sql.NullString{String: t.RevokedBy, Valid: true}
	}

	if !t.RevokedAt.IsZero() {
		dt.RevokedAt = sql.NullTime{Time: t.RevokedAt, Valid: true}
	}

	if t.RevokeReason != "" {
		dt.RevokeReason = sql.NullString{String: t.RevokeReason, Valid: true}
	}

	if t.DenyReason != "" {
		dt.DenyReason = sql.NullString{String: t.DenyReason, Valid: true}
	}

	return dt
}

// SaveToken persiste um novo token de elevação
func (r *PostgresElevationRepository) SaveToken(ctx context.Context, token *elevation.Token) error {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.SaveToken",
		trace.WithAttributes(
			attribute.String("elevation_id", token.ID),
			attribute.String("user_id", token.UserID),
			attribute.String("tenant_id", token.TenantID),
			attribute.String("market", token.Market),
		))
	defer span.End()

	dt := fromServiceToken(token)

	query := `
		INSERT INTO elevation_tokens (
			id, user_id, tenant_id, market, scopes, status, 
			justification, created_at, expires_at, approved_by, 
			approved_at, denied_by, denied_at, revoked_by, 
			revoked_at, emergency, revoke_reason, deny_reason, 
			usage_count, last_used_at
		) VALUES (
			:id, :user_id, :tenant_id, :market, :scopes, :status, 
			:justification, :created_at, :expires_at, :approved_by, 
			:approved_at, :denied_by, :denied_at, :revoked_by, 
			:revoked_at, :emergency, :revoke_reason, :deny_reason, 
			0, NULL
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, dt)
	if err != nil {
		r.logger.Error("Erro ao salvar token de elevação",
			zap.String("elevation_id", token.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save elevation token: %w", err)
	}

	// Adiciona entrada no histórico
	historyEntry := TokenHistoryEntry{
		TokenID:    token.ID,
		Timestamp:  time.Now(),
		Action:     "created",
		ActorID:    token.UserID,
		PrevStatus: "",
		NewStatus:  token.Status,
		Metadata: map[string]interface{}{
			"emergency": token.Emergency,
			"market":    token.Market,
		},
	}

	if err := r.saveHistoryEntry(ctx, &historyEntry); err != nil {
		// Não falha a operação principal, mas loga o erro
		r.logger.Warn("Erro ao salvar histórico do token",
			zap.String("elevation_id", token.ID),
			zap.Error(err))
	}

	return nil
}

// UpdateToken atualiza um token de elevação existente
func (r *PostgresElevationRepository) UpdateToken(ctx context.Context, token *elevation.Token) error {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.UpdateToken",
		trace.WithAttributes(
			attribute.String("elevation_id", token.ID),
			attribute.String("status", string(token.Status)),
		))
	defer span.End()

	// Recupera o estado anterior para o histórico
	var prevStatus string
	err := r.db.GetContext(ctx, &prevStatus, 
		"SELECT status FROM elevation_tokens WHERE id = $1", token.ID)
	if err != nil {
		r.logger.Error("Erro ao recuperar status anterior do token",
			zap.String("elevation_id", token.ID),
			zap.Error(err))
		// Continuamos mesmo com erro, pois o histórico não é crítico
	}

	dt := fromServiceToken(token)

	query := `
		UPDATE elevation_tokens SET
			status = :status,
			approved_by = :approved_by,
			approved_at = :approved_at,
			denied_by = :denied_by,
			denied_at = :denied_at,
			revoked_by = :revoked_by,
			revoked_at = :revoked_at,
			revoke_reason = :revoke_reason,
			deny_reason = :deny_reason
		WHERE id = :id
	`

	_, err = r.db.NamedExecContext(ctx, query, dt)
	if err != nil {
		r.logger.Error("Erro ao atualizar token de elevação",
			zap.String("elevation_id", token.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update elevation token: %w", err)
	}

	// Determina a ação e ator para o histórico
	var action, actorID string
	var reason string
	switch token.Status {
	case elevation.StatusActive:
		if prevStatus == string(elevation.StatusPendingApproval) {
			action = "approved"
			actorID = token.ApprovedBy
		}
	case elevation.StatusDenied:
		action = "denied"
		actorID = token.DeniedBy
		reason = token.DenyReason
	case elevation.StatusRevoked:
		action = "revoked"
		actorID = token.RevokedBy
		reason = token.RevokeReason
	case elevation.StatusExpired:
		action = "expired"
		// Sistema como ator para expiração automática
		actorID = "system"
	}

	// Adiciona entrada no histórico se houver mudança significativa
	if action != "" {
		historyEntry := TokenHistoryEntry{
			TokenID:    token.ID,
			Timestamp:  time.Now(),
			Action:     action,
			ActorID:    actorID,
			PrevStatus: elevation.Status(prevStatus),
			NewStatus:  token.Status,
			Reason:     reason,
		}

		if err := r.saveHistoryEntry(ctx, &historyEntry); err != nil {
			// Não falha a operação principal, mas loga o erro
			r.logger.Warn("Erro ao salvar histórico do token",
				zap.String("elevation_id", token.ID),
				zap.Error(err))
		}
	}

	return nil
}

// GetToken recupera um token de elevação pelo ID
func (r *PostgresElevationRepository) GetToken(ctx context.Context, tokenID string) (*elevation.Token, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.GetToken",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
		))
	defer span.End()

	var dt dbToken
	query := `SELECT * FROM elevation_tokens WHERE id = $1`
	err := r.db.GetContext(ctx, &dt, query, tokenID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("elevation token not found: %s", tokenID)
		}
		r.logger.Error("Erro ao recuperar token de elevação",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get elevation token: %w", err)
	}

	return dt.toServiceToken(), nil
}

// GetUserTokens recupera todos os tokens de um usuário
func (r *PostgresElevationRepository) GetUserTokens(ctx context.Context, userID, tenantID string) ([]*elevation.Token, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.GetUserTokens",
		trace.WithAttributes(
			attribute.String("user_id", userID),
			attribute.String("tenant_id", tenantID),
		))
	defer span.End()

	var dbTokens []dbToken
	query := `
		SELECT * FROM elevation_tokens 
		WHERE user_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
	`
	
	err := r.db.SelectContext(ctx, &dbTokens, query, userID, tenantID)
	if err != nil {
		r.logger.Error("Erro ao recuperar tokens do usuário",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	tokens := make([]*elevation.Token, len(dbTokens))
	for i, dt := range dbTokens {
		tokens[i] = dt.toServiceToken()
	}

	return tokens, nil
}

// CountActiveTokensByUser conta tokens ativos de um usuário
func (r *PostgresElevationRepository) CountActiveTokensByUser(ctx context.Context, userID, tenantID string) (int, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.CountActiveTokensByUser",
		trace.WithAttributes(
			attribute.String("user_id", userID),
			attribute.String("tenant_id", tenantID),
		))
	defer span.End()

	var count int
	query := `
		SELECT COUNT(*) FROM elevation_tokens 
		WHERE user_id = $1 AND tenant_id = $2 AND status = 'active' AND expires_at > NOW()
	`
	
	err := r.db.GetContext(ctx, &count, query, userID, tenantID)
	if err != nil {
		r.logger.Error("Erro ao contar tokens ativos",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count active tokens: %w", err)
	}

	return count, nil
}

// ListPendingTokens recupera tokens pendentes de aprovação para um mercado e tenant
func (r *PostgresElevationRepository) ListPendingTokens(ctx context.Context, tenantID, market string) ([]*elevation.Token, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.ListPendingTokens",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	var dbTokens []dbToken
	query := `
		SELECT * FROM elevation_tokens 
		WHERE tenant_id = $1 AND market = $2 AND status = 'pending_approval'
		ORDER BY created_at ASC
	`
	
	err := r.db.SelectContext(ctx, &dbTokens, query, tenantID, market)
	if err != nil {
		r.logger.Error("Erro ao listar tokens pendentes",
			zap.String("tenant_id", tenantID),
			zap.String("market", market),
			zap.Error(err))
		return nil, fmt.Errorf("failed to list pending tokens: %w", err)
	}

	tokens := make([]*elevation.Token, len(dbTokens))
	for i, dt := range dbTokens {
		tokens[i] = dt.toServiceToken()
	}

	return tokens, nil
}

// ListTokensByStatus recupera tokens por status
func (r *PostgresElevationRepository) ListTokensByStatus(ctx context.Context, tenantID, market string, status elevation.Status) ([]*elevation.Token, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.ListTokensByStatus",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
			attribute.String("status", string(status)),
		))
	defer span.End()

	var dbTokens []dbToken
	query := `
		SELECT * FROM elevation_tokens 
		WHERE tenant_id = $1 AND market = $2 AND status = $3
		ORDER BY created_at DESC
	`
	
	err := r.db.SelectContext(ctx, &dbTokens, query, tenantID, market, string(status))
	if err != nil {
		r.logger.Error("Erro ao listar tokens por status",
			zap.String("tenant_id", tenantID),
			zap.String("market", market),
			zap.String("status", string(status)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to list tokens by status: %w", err)
	}

	tokens := make([]*elevation.Token, len(dbTokens))
	for i, dt := range dbTokens {
		tokens[i] = dt.toServiceToken()
	}

	return tokens, nil
}

// ListExpiredTokens recupera tokens expirados que ainda não foram marcados como tal
func (r *PostgresElevationRepository) ListExpiredTokens(ctx context.Context) ([]*elevation.Token, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.ListExpiredTokens")
	defer span.End()

	var dbTokens []dbToken
	query := `
		SELECT * FROM elevation_tokens 
		WHERE status = 'active' AND expires_at < NOW()
	`
	
	err := r.db.SelectContext(ctx, &dbTokens, query)
	if err != nil {
		r.logger.Error("Erro ao listar tokens expirados",
			zap.Error(err))
		return nil, fmt.Errorf("failed to list expired tokens: %w", err)
	}

	tokens := make([]*elevation.Token, len(dbTokens))
	for i, dt := range dbTokens {
		tokens[i] = dt.toServiceToken()
	}

	return tokens, nil
}

// GetTokenHistory recupera o histórico de alterações de um token
func (r *PostgresElevationRepository) GetTokenHistory(ctx context.Context, tokenID string) ([]*TokenHistoryEntry, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.GetTokenHistory",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
		))
	defer span.End()

	type dbHistoryEntry struct {
		TokenID    string         `db:"token_id"`
		Timestamp  time.Time      `db:"timestamp"`
		Action     string         `db:"action"`
		ActorID    string         `db:"actor_id"`
		PrevStatus string         `db:"prev_status"`
		NewStatus  string         `db:"new_status"`
		Reason     sql.NullString `db:"reason"`
		Metadata   []byte         `db:"metadata"`
	}

	var dbEntries []dbHistoryEntry
	query := `
		SELECT * FROM elevation_token_history 
		WHERE token_id = $1
		ORDER BY timestamp ASC
	`
	
	err := r.db.SelectContext(ctx, &dbEntries, query, tokenID)
	if err != nil {
		r.logger.Error("Erro ao recuperar histórico do token",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get token history: %w", err)
	}

	entries := make([]*TokenHistoryEntry, len(dbEntries))
	for i, dbe := range dbEntries {
		entry := &TokenHistoryEntry{
			TokenID:    dbe.TokenID,
			Timestamp:  dbe.Timestamp,
			Action:     dbe.Action,
			ActorID:    dbe.ActorID,
			PrevStatus: elevation.Status(dbe.PrevStatus),
			NewStatus:  elevation.Status(dbe.NewStatus),
		}

		if dbe.Reason.Valid {
			entry.Reason = dbe.Reason.String
		}

		if len(dbe.Metadata) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(dbe.Metadata, &metadata); err == nil {
				entry.Metadata = metadata
			}
		}

		entries[i] = entry
	}

	return entries, nil
}

// saveHistoryEntry salva uma entrada no histórico do token
func (r *PostgresElevationRepository) saveHistoryEntry(ctx context.Context, entry *TokenHistoryEntry) error {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.saveHistoryEntry",
		trace.WithAttributes(
			attribute.String("elevation_id", entry.TokenID),
			attribute.String("action", entry.Action),
		))
	defer span.End()

	metadataBytes, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO elevation_token_history (
			token_id, timestamp, action, actor_id, 
			prev_status, new_status, reason, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		entry.TokenID,
		entry.Timestamp,
		entry.Action,
		entry.ActorID,
		string(entry.PrevStatus),
		string(entry.NewStatus),
		sql.NullString{String: entry.Reason, Valid: entry.Reason != ""},
		metadataBytes,
	)

	if err != nil {
		return fmt.Errorf("failed to save token history entry: %w", err)
	}

	return nil
}

// GetTokenStats retorna estatísticas de tokens por mercado e tenant
func (r *PostgresElevationRepository) GetTokenStats(ctx context.Context, tenantID, market string, since time.Time) (*TokenStats, error) {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.GetTokenStats",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("market", market),
		))
	defer span.End()

	// SQL para buscar estatísticas
	query := `
		WITH token_stats AS (
			SELECT
				COUNT(*) AS total_requested,
				SUM(CASE WHEN status = 'active' OR status = 'expired' OR status = 'revoked' THEN 1 ELSE 0 END) AS total_approved,
				SUM(CASE WHEN status = 'denied' THEN 1 ELSE 0 END) AS total_denied,
				SUM(CASE WHEN status = 'revoked' THEN 1 ELSE 0 END) AS total_revoked,
				SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END) AS total_expired,
				SUM(CASE WHEN emergency = true THEN 1 ELSE 0 END) AS total_emergency,
				SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS current_active,
				SUM(CASE WHEN status = 'pending_approval' THEN 1 ELSE 0 END) AS current_pending,
				AVG(CASE WHEN status != 'pending_approval' AND approved_at IS NOT NULL 
					THEN EXTRACT(EPOCH FROM (approved_at - created_at)) / 60 
					ELSE NULL END) AS avg_approval_time,
				AVG(usage_count) AS avg_usage
			FROM elevation_tokens
			WHERE tenant_id = $1 AND market = $2 AND created_at >= $3
		)
		SELECT
			COALESCE(total_requested, 0) AS total_requested,
			COALESCE(total_approved, 0) AS total_approved,
			COALESCE(total_denied, 0) AS total_denied,
			COALESCE(total_revoked, 0) AS total_revoked,
			COALESCE(total_expired, 0) AS total_expired,
			COALESCE(total_emergency, 0) AS total_emergency,
			COALESCE(current_active, 0) AS current_active_tokens,
			COALESCE(current_pending, 0) AS current_pending_tokens,
			COALESCE(ROUND(avg_approval_time::numeric), 0)::int AS avg_approval_time_minutes,
			COALESCE(ROUND(avg_usage::numeric), 0)::int AS avg_token_usage_count
		FROM token_stats
	`

	// Estrutura para armazenar resultados
	var stats struct {
		TotalRequested        int `db:"total_requested"`
		TotalApproved         int `db:"total_approved"`
		TotalDenied           int `db:"total_denied"`
		TotalRevoked          int `db:"total_revoked"`
		TotalExpired          int `db:"total_expired"`
		TotalEmergency        int `db:"total_emergency"`
		CurrentActiveTokens   int `db:"current_active_tokens"`
		CurrentPendingTokens  int `db:"current_pending_tokens"`
		AvgApprovalTimeMinutes int `db:"avg_approval_time_minutes"`
		AvgTokenUsageCount    int `db:"avg_token_usage_count"`
	}

	// Executa a consulta
	err := r.db.GetContext(ctx, &stats, query, tenantID, market, since)
	if err != nil {
		r.logger.Error("Erro ao recuperar estatísticas de tokens",
			zap.String("tenant_id", tenantID),
			zap.String("market", market),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get token statistics: %w", err)
	}

	// Converte para o modelo de resposta
	return &TokenStats{
		TotalRequested:        stats.TotalRequested,
		TotalApproved:         stats.TotalApproved,
		TotalDenied:           stats.TotalDenied,
		TotalRevoked:          stats.TotalRevoked,
		TotalExpired:          stats.TotalExpired,
		TotalEmergency:        stats.TotalEmergency,
		AvgApprovalTimeMinutes: stats.AvgApprovalTimeMinutes,
		AvgTokenUsageCount:    stats.AvgTokenUsageCount,
		CurrentActiveTokens:   stats.CurrentActiveTokens,
		CurrentPendingTokens:  stats.CurrentPendingTokens,
	}, nil
}

// DeleteToken exclui permanentemente um token (apenas para testes)
func (r *PostgresElevationRepository) DeleteToken(ctx context.Context, tokenID string) error {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.DeleteToken",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
		))
	defer span.End()

	// Primeiro, exclui o histórico para evitar violação de FK
	_, err := r.db.ExecContext(ctx, 
		"DELETE FROM elevation_token_history WHERE token_id = $1", tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete token history: %w", err)
	}

	// Depois, exclui o token
	_, err = r.db.ExecContext(ctx, 
		"DELETE FROM elevation_tokens WHERE id = $1", tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// IncrementTokenUsage incrementa o contador de uso de um token
func (r *PostgresElevationRepository) IncrementTokenUsage(ctx context.Context, tokenID string) error {
	ctx, span := r.tracer.Start(ctx, "PostgresElevationRepository.IncrementTokenUsage",
		trace.WithAttributes(
			attribute.String("elevation_id", tokenID),
		))
	defer span.End()

	query := `
		UPDATE elevation_tokens 
		SET usage_count = usage_count + 1, last_used_at = NOW()
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		r.logger.Error("Erro ao incrementar uso do token",
			zap.String("elevation_id", tokenID),
			zap.Error(err))
		return fmt.Errorf("failed to increment token usage: %w", err)
	}

	return nil
}