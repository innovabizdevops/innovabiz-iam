/**
 * @file attribute_postgres_repository.go
 * @description Implementação do repositório de atributos contextuais usando PostgreSQL
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

// AttributePostgresRepository implementa a interface AttributeRepository usando PostgreSQL
type AttributePostgresRepository struct {
	db *sqlx.DB
}

// dbAttribute é a estrutura para mapeamento ORM da tabela de atributos
type dbAttribute struct {
	ID                uuid.UUID      `db:"id"`
	ContextID         uuid.UUID      `db:"context_id"`
	AttributeKey      string         `db:"attribute_key"`
	AttributeValue    string         `db:"attribute_value"`
	SensitivityLevel  string         `db:"sensitivity_level"`
	VerificationStatus string         `db:"verification_status"`
	VerificationSource sql.NullString `db:"verification_source"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	Metadata          []byte         `db:"metadata"`
}

// NewAttributePostgresRepository cria uma nova instância do repositório
func NewAttributePostgresRepository(db *sqlx.DB) *AttributePostgresRepository {
	return &AttributePostgresRepository{
		db: db,
	}
}

// Create implementa a criação de um atributo
func (r *AttributePostgresRepository) Create(ctx context.Context, attribute *models.ContextAttribute) error {
	// Serializar metadados para JSON
	metadataJSON, err := json.Marshal(attribute.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}
	
	// Preparar valor para fonte de verificação (pode ser nulo)
	var verificationSource sql.NullString
	if attribute.VerificationSource != "" {
		verificationSource.String = attribute.VerificationSource
		verificationSource.Valid = true
	}
	
	// Inserir no banco de dados
	query := `
		INSERT INTO context_attribute (
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
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	_, err = r.db.ExecContext(
		ctx,
		query,
		attribute.ID,
		attribute.ContextID,
		attribute.AttributeKey,
		attribute.AttributeValue,
		attribute.SensitivityLevel,
		attribute.VerificationStatus,
		verificationSource,
		attribute.CreatedAt,
		attribute.UpdatedAt,
		metadataJSON,
	)
	
	if err != nil {
		// Verificar se é erro de duplicidade
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return models.ErrDuplicateAttribute
		}
		
		// Verificar se o contexto existe (violação de chave estrangeira)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return models.ErrContextNotFound
		}
		
		return fmt.Errorf("erro ao inserir atributo: %w", err)
	}
	
	return nil
}

// Update implementa a atualização de um atributo
func (r *AttributePostgresRepository) Update(ctx context.Context, attribute *models.ContextAttribute) error {
	// Serializar metadados para JSON
	metadataJSON, err := json.Marshal(attribute.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}
	
	// Preparar valor para fonte de verificação (pode ser nulo)
	var verificationSource sql.NullString
	if attribute.VerificationSource != "" {
		verificationSource.String = attribute.VerificationSource
		verificationSource.Valid = true
	}
	
	// Atualizar no banco de dados
	query := `
		UPDATE context_attribute
		SET 
			attribute_key = $2,
			attribute_value = $3,
			sensitivity_level = $4,
			verification_status = $5,
			verification_source = $6,
			updated_at = $7,
			metadata = $8
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		attribute.ID,
		attribute.AttributeKey,
		attribute.AttributeValue,
		attribute.SensitivityLevel,
		attribute.VerificationStatus,
		verificationSource,
		time.Now().UTC(),
		metadataJSON,
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar atributo: %w", err)
	}
	
	// Verificar se atributo existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrAttributeNotFound
	}
	
	return nil
}

// GetByID implementa a recuperação de um atributo por ID
func (r *AttributePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ContextAttribute, error) {
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
		WHERE id = $1
	`
	
	var dbAttr dbAttribute
	if err := r.db.GetContext(ctx, &dbAttr, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrAttributeNotFound
		}
		return nil, fmt.Errorf("erro ao buscar atributo: %w", err)
	}
	
	// Converter para modelo de domínio
	return r.mapToModel(&dbAttr)
}

// GetByContextAndKey implementa a recuperação de um atributo por contexto e chave
func (r *AttributePostgresRepository) GetByContextAndKey(
	ctx context.Context,
	contextID uuid.UUID,
	key string,
) (*models.ContextAttribute, error) {
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
		WHERE context_id = $1 AND attribute_key = $2
	`
	
	var dbAttr dbAttribute
	if err := r.db.GetContext(ctx, &dbAttr, query, contextID, key); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrAttributeNotFound
		}
		return nil, fmt.Errorf("erro ao buscar atributo: %w", err)
	}
	
	// Converter para modelo de domínio
	return r.mapToModel(&dbAttr)
}

// ListByContext implementa a listagem de atributos de um contexto
func (r *AttributePostgresRepository) ListByContext(
	ctx context.Context,
	contextID uuid.UUID,
) ([]*models.ContextAttribute, error) {
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
		ORDER BY created_at
	`
	
	var dbAttrs []dbAttribute
	if err := r.db.SelectContext(ctx, &dbAttrs, query, contextID); err != nil {
		return nil, fmt.Errorf("erro ao listar atributos: %w", err)
	}
	
	// Converter para modelos de domínio
	attributes := make([]*models.ContextAttribute, 0, len(dbAttrs))
	for _, dbAttr := range dbAttrs {
		attribute, err := r.mapToModel(&dbAttr)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attribute)
	}
	
	return attributes, nil
}

// Delete implementa a remoção de um atributo
func (r *AttributePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM context_attribute
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erro ao excluir atributo: %w", err)
	}
	
	// Verificar se atributo existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrAttributeNotFound
	}
	
	return nil
}

// BatchCreate implementa a criação de múltiplos atributos em uma única operação
func (r *AttributePostgresRepository) BatchCreate(ctx context.Context, attributes []*models.ContextAttribute) error {
	// Usar transação para garantir atomicidade
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	
	// Garantir rollback em caso de erro
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	
	// Preparar statement para reutilização
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO context_attribute (
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
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement: %w", err)
	}
	defer stmt.Close()
	
	// Inserir cada atributo
	for _, attribute := range attributes {
		// Serializar metadados para JSON
		metadataJSON, err := json.Marshal(attribute.Metadata)
		if err != nil {
			return fmt.Errorf("erro ao serializar metadados: %w", err)
		}
		
		// Preparar valor para fonte de verificação (pode ser nulo)
		var verificationSource sql.NullString
		if attribute.VerificationSource != "" {
			verificationSource.String = attribute.VerificationSource
			verificationSource.Valid = true
		}
		
		_, err = stmt.ExecContext(
			ctx,
			attribute.ID,
			attribute.ContextID,
			attribute.AttributeKey,
			attribute.AttributeValue,
			attribute.SensitivityLevel,
			attribute.VerificationStatus,
			verificationSource,
			attribute.CreatedAt,
			attribute.UpdatedAt,
			metadataJSON,
		)
		
		if err != nil {
			// Verificar erro de duplicidade individualmente
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("atributo duplicado: %s", attribute.AttributeKey)
			}
			
			return fmt.Errorf("erro ao inserir atributo em lote: %w", err)
		}
	}
	
	// Comitar transação
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao comitar transação: %w", err)
	}
	
	return nil
}

// List implementa a listagem de atributos com filtros e paginação
func (r *AttributePostgresRepository) List(
	ctx context.Context, 
	filter repositories.AttributeFilter, 
	page, pageSize int,
) ([]*models.ContextAttribute, int, error) {
	// Construir query com filtros dinâmicos
	baseQuery := `
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
	`
	
	countQuery := `
		SELECT COUNT(*) 
		FROM context_attribute
	`
	
	// Construir condições WHERE
	var conditions []string
	var params []interface{}
	paramIdx := 1
	
	// Filtro por contexto
	if filter.ContextID != nil {
		conditions = append(conditions, fmt.Sprintf("context_id = $%d", paramIdx))
		params = append(params, *filter.ContextID)
		paramIdx++
	}
	
	// Filtro por chaves de atributos
	if len(filter.AttributeKeys) > 0 {
		conditions = append(conditions, fmt.Sprintf("attribute_key = ANY($%d)", paramIdx))
		params = append(params, pq.Array(filter.AttributeKeys))
		paramIdx++
	}
	
	// Filtro por níveis de sensibilidade
	if len(filter.SensitivityLevels) > 0 {
		var sensitivityValues []string
		for _, s := range filter.SensitivityLevels {
			sensitivityValues = append(sensitivityValues, string(s))
		}
		conditions = append(conditions, fmt.Sprintf("sensitivity_level = ANY($%d)", paramIdx))
		params = append(params, pq.Array(sensitivityValues))
		paramIdx++
	}
	
	// Filtro por status de verificação
	if len(filter.VerificationStatus) > 0 {
		var verificationValues []string
		for _, v := range filter.VerificationStatus {
			verificationValues = append(verificationValues, string(v))
		}
		conditions = append(conditions, fmt.Sprintf("verification_status = ANY($%d)", paramIdx))
		params = append(params, pq.Array(verificationValues))
		paramIdx++
	}
	
	// Filtro por fontes de verificação
	if len(filter.VerificationSources) > 0 {
		conditions = append(conditions, fmt.Sprintf("verification_source = ANY($%d)", paramIdx))
		params = append(params, pq.Array(filter.VerificationSources))
		paramIdx++
	}
	
	// Filtro por texto de busca em valores
	if filter.SearchValue != "" {
		conditions = append(conditions, fmt.Sprintf("attribute_value ILIKE $%d", paramIdx))
		params = append(params, "%"+filter.SearchValue+"%")
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
			"created_at":         true,
			"updated_at":         true,
			"attribute_key":      true,
			"verification_status": true,
			"sensitivity_level":  true,
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
		return nil, 0, fmt.Errorf("erro ao contar atributos: %w", err)
	}
	
	// Se não há resultados, retornar lista vazia
	if totalCount == 0 {
		return []*models.ContextAttribute{}, 0, nil
	}
	
	// Executar query de listagem
	var dbAttrs []dbAttribute
	if err := r.db.SelectContext(ctx, &dbAttrs, listQuery, listParams...); err != nil {
		return nil, 0, fmt.Errorf("erro ao listar atributos: %w", err)
	}
	
	// Converter para modelos de domínio
	attributes := make([]*models.ContextAttribute, 0, len(dbAttrs))
	for _, dbAttr := range dbAttrs {
		attribute, err := r.mapToModel(&dbAttr)
		if err != nil {
			return nil, 0, err
		}
		attributes = append(attributes, attribute)
	}
	
	return attributes, totalCount, nil
}

// UpdateVerification implementa a atualização do status e fonte de verificação de um atributo
func (r *AttributePostgresRepository) UpdateVerification(
	ctx context.Context,
	attributeID uuid.UUID, 
	status models.VerificationStatus, 
	source string,
) error {
	// Preparar valor para fonte de verificação (pode ser nulo)
	var verificationSource sql.NullString
	if source != "" {
		verificationSource.String = source
		verificationSource.Valid = true
	}
	
	query := `
		UPDATE context_attribute
		SET 
			verification_status = $2, 
			verification_source = $3, 
			updated_at = $4
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		attributeID,
		status,
		verificationSource,
		time.Now().UTC(),
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar verificação: %w", err)
	}
	
	// Verificar se atributo existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrAttributeNotFound
	}
	
	return nil
}

// SearchAttributes implementa a busca de atributos por valor em todos os contextos
func (r *AttributePostgresRepository) SearchAttributes(
	ctx context.Context,
	searchValue string,
	sensitivityLevels []models.SensitivityLevel,
) ([]*models.ContextAttribute, error) {
	// Converter níveis de sensibilidade para strings
	var sensitivityValues []string
	for _, s := range sensitivityLevels {
		sensitivityValues = append(sensitivityValues, string(s))
	}
	
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
		WHERE attribute_value ILIKE $1
		AND sensitivity_level = ANY($2)
		ORDER BY created_at DESC
		LIMIT 100
	`
	
	var dbAttrs []dbAttribute
	if err := r.db.SelectContext(
		ctx, 
		&dbAttrs, 
		query, 
		"%"+searchValue+"%", 
		pq.Array(sensitivityValues),
	); err != nil {
		return nil, fmt.Errorf("erro ao buscar atributos: %w", err)
	}
	
	// Converter para modelos de domínio
	attributes := make([]*models.ContextAttribute, 0, len(dbAttrs))
	for _, dbAttr := range dbAttrs {
		attribute, err := r.mapToModel(&dbAttr)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, attribute)
	}
	
	return attributes, nil
}

// mapToModel converte uma estrutura de banco de dados para um modelo de domínio
func (r *AttributePostgresRepository) mapToModel(dbAttr *dbAttribute) (*models.ContextAttribute, error) {
	// Desserializar metadados
	var metadata map[string]interface{}
	if len(dbAttr.Metadata) > 0 {
		if err := json.Unmarshal(dbAttr.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("erro ao desserializar metadados: %w", err)
		}
	}
	
	// Criar atributo
	attribute := &models.ContextAttribute{
		ID:                dbAttr.ID,
		ContextID:         dbAttr.ContextID,
		AttributeKey:      dbAttr.AttributeKey,
		AttributeValue:    dbAttr.AttributeValue,
		SensitivityLevel:  models.SensitivityLevel(dbAttr.SensitivityLevel),
		VerificationStatus: models.VerificationStatus(dbAttr.VerificationStatus),
		CreatedAt:         dbAttr.CreatedAt,
		UpdatedAt:         dbAttr.UpdatedAt,
		Metadata:          metadata,
	}
	
	// Adicionar fonte de verificação se existir
	if dbAttr.VerificationSource.Valid {
		attribute.VerificationSource = dbAttr.VerificationSource.String
	}
	
	return attribute, nil
}