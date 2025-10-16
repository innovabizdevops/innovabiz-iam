/**
 * @file context_postgres_repository.go
 * @description Implementação do repositório de contextos de identidade usando PostgreSQL
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	
	"innovabiz/iam/src/multi-context/domain/models"
	"innovabiz/iam/src/multi-context/domain/repositories"
)

// ContextPostgresRepository implementa a interface ContextRepository usando PostgreSQL
type ContextPostgresRepository struct {
	db *sqlx.DB
}

// dbContext é a estrutura para mapeamento ORM da tabela de contextos
type dbContext struct {
	ID               uuid.UUID `db:"id"`
	IdentityID       uuid.UUID `db:"identity_id"`
	ContextType      string    `db:"context_type"`
	Status           string    `db:"status"`
	VerificationLevel string    `db:"verification_level"`
	TrustScore       float64   `db:"trust_score"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
	Metadata         []byte    `db:"metadata"`
}

// NewContextPostgresRepository cria uma nova instância do repositório
func NewContextPostgresRepository(db *sqlx.DB) *ContextPostgresRepository {
	return &ContextPostgresRepository{
		db: db,
	}
}

// Create implementa a criação de um contexto
func (r *ContextPostgresRepository) Create(ctx context.Context, identityContext *models.IdentityContext) error {
	// Serializar metadados para JSON
	metadataJSON, err := json.Marshal(identityContext.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}
	
	// Inserir no banco de dados
	query := `
		INSERT INTO identity_context (
			id,
			identity_id,
			context_type,
			status,
			verification_level,
			trust_score,
			created_at,
			updated_at,
			metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err = r.db.ExecContext(
		ctx,
		query,
		identityContext.ID,
		identityContext.IdentityID,
		identityContext.ContextType,
		identityContext.Status,
		identityContext.VerificationLevel,
		identityContext.TrustScore,
		identityContext.CreatedAt,
		identityContext.UpdatedAt,
		metadataJSON,
	)
	
	if err != nil {
		// Verificar se é erro de duplicidade
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return models.ErrDuplicateContext
		}
		
		// Verificar se a identidade existe (violação de chave estrangeira)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return models.ErrIdentityNotFound
		}
		
		return fmt.Errorf("erro ao inserir contexto: %w", err)
	}
	
	return nil
}

// Update implementa a atualização de um contexto
func (r *ContextPostgresRepository) Update(ctx context.Context, identityContext *models.IdentityContext) error {
	// Serializar metadados para JSON
	metadataJSON, err := json.Marshal(identityContext.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}
	
	// Atualizar no banco de dados
	query := `
		UPDATE identity_context
		SET 
			context_type = $2,
			status = $3,
			verification_level = $4,
			trust_score = $5,
			updated_at = $6,
			metadata = $7
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		identityContext.ID,
		identityContext.ContextType,
		identityContext.Status,
		identityContext.VerificationLevel,
		identityContext.TrustScore,
		time.Now().UTC(),
		metadataJSON,
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar contexto: %w", err)
	}
	
	// Verificar se contexto existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrContextNotFound
	}
	
	return nil
}

// GetByID implementa a recuperação de um contexto por ID
func (r *ContextPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.IdentityContext, error) {
	query := `
		SELECT 
			id,
			identity_id,
			context_type,
			status,
			verification_level,
			trust_score,
			created_at,
			updated_at,
			metadata
		FROM identity_context
		WHERE id = $1
	`
	
	var dbCtx dbContext
	if err := r.db.GetContext(ctx, &dbCtx, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrContextNotFound
		}
		return nil, fmt.Errorf("erro ao buscar contexto: %w", err)
	}
	
	// Converter para modelo de domínio
	return r.mapToModel(&dbCtx)
}

// GetByIdentityAndType implementa a recuperação de um contexto por identidade e tipo
func (r *ContextPostgresRepository) GetByIdentityAndType(
	ctx context.Context,
	identityID uuid.UUID,
	contextType models.ContextType,
) (*models.IdentityContext, error) {
	query := `
		SELECT 
			id,
			identity_id,
			context_type,
			status,
			verification_level,
			trust_score,
			created_at,
			updated_at,
			metadata
		FROM identity_context
		WHERE identity_id = $1 AND context_type = $2
	`
	
	var dbCtx dbContext
	if err := r.db.GetContext(ctx, &dbCtx, query, identityID, contextType); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrContextNotFound
		}
		return nil, fmt.Errorf("erro ao buscar contexto: %w", err)
	}
	
	// Converter para modelo de domínio
	return r.mapToModel(&dbCtx)
}

// ListByIdentity implementa a listagem de contextos de uma identidade
func (r *ContextPostgresRepository) ListByIdentity(
	ctx context.Context,
	identityID uuid.UUID,
) ([]*models.IdentityContext, error) {
	query := `
		SELECT 
			id,
			identity_id,
			context_type,
			status,
			verification_level,
			trust_score,
			created_at,
			updated_at,
			metadata
		FROM identity_context
		WHERE identity_id = $1
		ORDER BY created_at
	`
	
	var dbCtxs []dbContext
	if err := r.db.SelectContext(ctx, &dbCtxs, query, identityID); err != nil {
		return nil, fmt.Errorf("erro ao listar contextos: %w", err)
	}
	
	// Converter para modelos de domínio
	contexts := make([]*models.IdentityContext, 0, len(dbCtxs))
	for _, dbCtx := range dbCtxs {
		context, err := r.mapToModel(&dbCtx)
		if err != nil {
			return nil, err
		}
		contexts = append(contexts, context)
	}
	
	return contexts, nil
}

// Delete implementa a remoção lógica de um contexto
func (r *ContextPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE identity_context
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
		models.ContextStatusInactive,
		time.Now().UTC(),
	)
	
	if err != nil {
		return fmt.Errorf("erro ao excluir contexto: %w", err)
	}
	
	// Verificar se contexto existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrContextNotFound
	}
	
	return nil
}

// LoadAttributes carrega os atributos para um contexto específico
func (r *ContextPostgresRepository) LoadAttributes(
	ctx context.Context,
	identityContext *models.IdentityContext,
) error {
	query := `
		SELECT 
			id,
			context_id,
			attribute_key,
			attribute_value,
			sensitivity_level,
			verification_status,
			verification_source,
			created_at,
			updated_at,
			metadata
		FROM context_attribute
		WHERE context_id = $1
	`
	
	rows, err := r.db.QueryxContext(ctx, query, identityContext.ID)
	if err != nil {
		return fmt.Errorf("erro ao carregar atributos: %w", err)
	}
	defer rows.Close()
	
	attributes := make([]*models.ContextAttribute, 0)
	
	for rows.Next() {
		var attributeID, contextID uuid.UUID
		var attributeKey, attributeValue, sensitivityLevel, verificationStatus string
		var verificationSource sql.NullString
		var createdAt, updatedAt time.Time
		var metadataJSON []byte
		
		if err := rows.Scan(
			&attributeID,
			&contextID,
			&attributeKey,
			&attributeValue,
			&sensitivityLevel,
			&verificationStatus,
			&verificationSource,
			&createdAt,
			&updatedAt,
			&metadataJSON,
		); err != nil {
			return fmt.Errorf("erro ao escanear atributo: %w", err)
		}
		
		// Desserializar metadados
		var metadata map[string]interface{}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return fmt.Errorf("erro ao desserializar metadados do atributo: %w", err)
			}
		}
		
		// Criar atributo de domínio
		attribute := &models.ContextAttribute{
			ID:                attributeID,
			ContextID:         contextID,
			AttributeKey:      attributeKey,
			AttributeValue:    attributeValue,
			SensitivityLevel:  models.SensitivityLevel(sensitivityLevel),
			VerificationStatus: models.VerificationStatus(verificationStatus),
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
			Metadata:          metadata,
		}
		
		// Adicionar fonte de verificação se existir
		if verificationSource.Valid {
			attribute.VerificationSource = verificationSource.String
		}
		
		attributes = append(attributes, attribute)
	}
	
	if err := rows.Err(); err != nil {
		return fmt.Errorf("erro ao iterar atributos: %w", err)
	}
	
	// Atribuir atributos ao contexto
	identityContext.Attributes = attributes
	
	return nil
}

// List implementa a listagem de contextos com filtros e paginação
func (r *ContextPostgresRepository) List(
	ctx context.Context, 
	filter repositories.ContextFilter, 
	page, pageSize int,
) ([]*models.IdentityContext, int, error) {
	// Construir query com filtros dinâmicos
	baseQuery := `
		SELECT 
			id,
			identity_id,
			context_type,
			status,
			verification_level,
			trust_score,
			created_at,
			updated_at,
			metadata
		FROM identity_context
	`
	
	countQuery := `
		SELECT COUNT(*) 
		FROM identity_context
	`
	
	// Construir condições WHERE
	var conditions []string
	var params []interface{}
	paramIdx := 1
	
	// Filtro por identidade
	if filter.IdentityID != nil {
		conditions = append(conditions, fmt.Sprintf("identity_id = $%d", paramIdx))
		params = append(params, *filter.IdentityID)
		paramIdx++
	}
	
	// Filtro por tipos de contexto
	if len(filter.ContextTypes) > 0 {
		var contextTypeValues []string
		for _, ct := range filter.ContextTypes {
			contextTypeValues = append(contextTypeValues, string(ct))
		}
		conditions = append(conditions, fmt.Sprintf("context_type = ANY($%d)", paramIdx))
		params = append(params, pq.Array(contextTypeValues))
		paramIdx++
	}
	
	// Filtro por status
	if len(filter.Status) > 0 {
		var statusValues []string
		for _, s := range filter.Status {
			statusValues = append(statusValues, string(s))
		}
		conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", paramIdx))
		params = append(params, pq.Array(statusValues))
		paramIdx++
	}
	
	// Filtro por nível de verificação mínimo
	if filter.MinVerificationLevel != nil {
		// Mapeamento de níveis de verificação para valores numéricos para comparação
		verificationLevels := map[models.VerificationLevel]int{
			models.VerificationNone:     0,
			models.VerificationBasic:    1,
			models.VerificationStandard: 2,
			models.VerificationEnhanced: 3,
			models.VerificationComplete: 4,
		}
		
		levelValue := verificationLevels[*filter.MinVerificationLevel]
		
		// Construir condição IN com níveis iguais ou superiores
		var validLevels []string
		for level, value := range verificationLevels {
			if value >= levelValue {
				validLevels = append(validLevels, string(level))
			}
		}
		
		conditions = append(conditions, fmt.Sprintf("verification_level = ANY($%d)", paramIdx))
		params = append(params, pq.Array(validLevels))
		paramIdx++
	}
	
	// Filtro por pontuação de confiança mínima
	if filter.MinTrustScore != nil {
		conditions = append(conditions, fmt.Sprintf("trust_score >= $%d", paramIdx))
		params = append(params, *filter.MinTrustScore)
		paramIdx++
	}
	
	// Filtro por data de criação (início)
	if filter.CreatedStart != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramIdx))
		params = append(params, *filter.CreatedStart)
		paramIdx++
	}
	
	// Filtro por data de criação (fim)
	if filter.CreatedEnd != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramIdx))
		params = append(params, *filter.CreatedEnd)
		paramIdx++
	}
	
	// Adicionar condições WHERE se houver
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	// Adicionar ordenação
	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		// Validar campos permitidos para ordenação
		allowedFields := map[string]bool{
			"created_at":        true,
			"updated_at":        true,
			"context_type":      true,
			"status":            true,
			"trust_score":       true,
			"verification_level": true,
		}
		
		if allowedFields[filter.OrderBy] {
			orderDir := "DESC"
			if filter.OrderDirection == "asc" {
				orderDir = "ASC"
			}
			orderBy = fmt.Sprintf("%s %s", filter.OrderBy, orderDir)
		}
	}
	
	// Construir query completa
	listQuery := fmt.Sprintf("%s %s ORDER BY %s LIMIT $%d OFFSET $%d", 
		baseQuery, whereClause, orderBy, paramIdx, paramIdx+1)
	
	// Adicionar parâmetros de paginação
	listParams := append(params, pageSize, (page-1)*pageSize)
	
	// Executar query de contagem
	countQueryFull := countQuery + " " + whereClause
	var totalCount int
	if err := r.db.GetContext(ctx, &totalCount, countQueryFull, params...); err != nil {
		return nil, 0, fmt.Errorf("erro ao contar contextos: %w", err)
	}
	
	// Se não há resultados, retornar lista vazia
	if totalCount == 0 {
		return []*models.IdentityContext{}, 0, nil
	}
	
	// Executar query de listagem
	var dbCtxs []dbContext
	if err := r.db.SelectContext(ctx, &dbCtxs, listQuery, listParams...); err != nil {
		return nil, 0, fmt.Errorf("erro ao listar contextos: %w", err)
	}
	
	// Converter para modelos de domínio
	contexts := make([]*models.IdentityContext, 0, len(dbCtxs))
	for _, dbCtx := range dbCtxs {
		context, err := r.mapToModel(&dbCtx)
		if err != nil {
			return nil, 0, err
		}
		contexts = append(contexts, context)
	}
	
	return contexts, totalCount, nil
}

// UpdateTrustScore implementa a atualização da pontuação de confiança de um contexto
func (r *ContextPostgresRepository) UpdateTrustScore(
	ctx context.Context,
	contextID uuid.UUID, 
	score float64,
) error {
	query := `
		UPDATE identity_context
		SET trust_score = $2, updated_at = $3
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		contextID,
		score,
		time.Now().UTC(),
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar pontuação de confiança: %w", err)
	}
	
	// Verificar se contexto existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrContextNotFound
	}
	
	return nil
}

// UpdateVerificationLevel implementa a atualização do nível de verificação de um contexto
func (r *ContextPostgresRepository) UpdateVerificationLevel(
	ctx context.Context,
	contextID uuid.UUID, 
	level models.VerificationLevel,
) error {
	query := `
		UPDATE identity_context
		SET verification_level = $2, updated_at = $3
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		contextID,
		level,
		time.Now().UTC(),
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar nível de verificação: %w", err)
	}
	
	// Verificar se contexto existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrContextNotFound
	}
	
	return nil
}

// mapToModel converte uma estrutura de banco de dados para um modelo de domínio
func (r *ContextPostgresRepository) mapToModel(dbCtx *dbContext) (*models.IdentityContext, error) {
	// Desserializar metadados
	var metadata map[string]interface{}
	if len(dbCtx.Metadata) > 0 {
		if err := json.Unmarshal(dbCtx.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("erro ao desserializar metadados: %w", err)
		}
	}
	
	// Criar contexto
	context := &models.IdentityContext{
		ID:                dbCtx.ID,
		IdentityID:        dbCtx.IdentityID,
		ContextType:       models.ContextType(dbCtx.ContextType),
		Status:            models.ContextStatus(dbCtx.Status),
		VerificationLevel: models.VerificationLevel(dbCtx.VerificationLevel),
		TrustScore:        dbCtx.TrustScore,
		CreatedAt:         dbCtx.CreatedAt,
		UpdatedAt:         dbCtx.UpdatedAt,
		Metadata:          metadata,
		Attributes:        make([]*models.ContextAttribute, 0),
	}
	
	return context, nil
}