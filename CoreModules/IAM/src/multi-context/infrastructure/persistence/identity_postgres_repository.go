/**
 * @file identity_postgres_repository.go
 * @description Implementação do repositório de identidades usando PostgreSQL
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

// IdentityPostgresRepository implementa a interface IdentityRepository usando PostgreSQL
type IdentityPostgresRepository struct {
	db *sqlx.DB
}

// dbIdentity é a estrutura para mapeamento ORM da tabela de identidades
type dbIdentity struct {
	ID             uuid.UUID      `db:"id"`
	PrimaryKeyType string         `db:"primary_key_type"`
	PrimaryKeyValue string        `db:"primary_key_value"`
	MasterPersonID sql.NullString `db:"master_person_id"`
	Status         string         `db:"status"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	Metadata       []byte         `db:"metadata"`
}

// NewIdentityPostgresRepository cria uma nova instância do repositório
func NewIdentityPostgresRepository(db *sqlx.DB) *IdentityPostgresRepository {
	return &IdentityPostgresRepository{
		db: db,
	}
}

// Create implementa a criação de uma identidade
func (r *IdentityPostgresRepository) Create(ctx context.Context, identity *models.Identity) error {
	// Serializar metadados para JSON
	metadataJSON, err := json.Marshal(identity.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}
	
	// Preparar valor para master person ID (pode ser nulo)
	var masterPersonID sql.NullString
	if identity.MasterPersonID != nil {
		masterPersonID.String = identity.MasterPersonID.String()
		masterPersonID.Valid = true
	}
	
	// Inserir no banco de dados
	query := `
		INSERT INTO identity_multi_context (
			id, 
			primary_key_type, 
			primary_key_value, 
			master_person_id,
			status, 
			created_at, 
			updated_at, 
			metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err = r.db.ExecContext(
		ctx,
		query,
		identity.ID,
		identity.PrimaryKeyType,
		identity.PrimaryKeyValue,
		masterPersonID,
		identity.Status,
		identity.CreatedAt,
		identity.UpdatedAt,
		metadataJSON,
	)
	
	if err != nil {
		// Verificar se é erro de duplicidade
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return models.ErrDuplicateIdentity
		}
		return fmt.Errorf("erro ao inserir identidade: %w", err)
	}
	
	return nil
}

// Update implementa a atualização de uma identidade
func (r *IdentityPostgresRepository) Update(ctx context.Context, identity *models.Identity) error {
	// Serializar metadados para JSON
	metadataJSON, err := json.Marshal(identity.Metadata)
	if err != nil {
		return fmt.Errorf("erro ao serializar metadados: %w", err)
	}
	
	// Preparar valor para master person ID (pode ser nulo)
	var masterPersonID sql.NullString
	if identity.MasterPersonID != nil {
		masterPersonID.String = identity.MasterPersonID.String()
		masterPersonID.Valid = true
	}
	
	// Atualizar no banco de dados
	query := `
		UPDATE identity_multi_context
		SET 
			primary_key_type = $2,
			primary_key_value = $3,
			master_person_id = $4,
			status = $5,
			updated_at = $6,
			metadata = $7
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		identity.ID,
		identity.PrimaryKeyType,
		identity.PrimaryKeyValue,
		masterPersonID,
		identity.Status,
		time.Now().UTC(),
		metadataJSON,
	)
	
	if err != nil {
		return fmt.Errorf("erro ao atualizar identidade: %w", err)
	}
	
	// Verificar se identidade existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrIdentityNotFound
	}
	
	return nil
}

// GetByID implementa a recuperação de uma identidade por ID
func (r *IdentityPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Identity, error) {
	query := `
		SELECT 
			id, 
			primary_key_type, 
			primary_key_value, 
			master_person_id,
			status, 
			created_at, 
			updated_at, 
			metadata
		FROM identity_multi_context
		WHERE id = $1
	`
	
	var dbIdent dbIdentity
	if err := r.db.GetContext(ctx, &dbIdent, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrIdentityNotFound
		}
		return nil, fmt.Errorf("erro ao buscar identidade: %w", err)
	}
	
	// Converter para modelo de domínio
	return r.mapToModel(&dbIdent)
}

// GetByPrimaryKey implementa a recuperação de uma identidade por chave primária
func (r *IdentityPostgresRepository) GetByPrimaryKey(
	ctx context.Context, 
	keyType models.PrimaryKeyType, 
	keyValue string,
) (*models.Identity, error) {
	query := `
		SELECT 
			id, 
			primary_key_type, 
			primary_key_value, 
			master_person_id,
			status, 
			created_at, 
			updated_at, 
			metadata
		FROM identity_multi_context
		WHERE primary_key_type = $1 AND primary_key_value = $2
	`
	
	var dbIdent dbIdentity
	if err := r.db.GetContext(ctx, &dbIdent, query, keyType, keyValue); err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrIdentityNotFound
		}
		return nil, fmt.Errorf("erro ao buscar identidade: %w", err)
	}
	
	// Converter para modelo de domínio
	return r.mapToModel(&dbIdent)
}

// Delete implementa a remoção lógica de uma identidade
func (r *IdentityPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE identity_multi_context
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
		models.StatusDeleted,
		time.Now().UTC(),
	)
	
	if err != nil {
		return fmt.Errorf("erro ao excluir identidade: %w", err)
	}
	
	// Verificar se identidade existe
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao obter linhas afetadas: %w", err)
	}
	
	if rowsAffected == 0 {
		return models.ErrIdentityNotFound
	}
	
	return nil
}

// List implementa a listagem de identidades com filtros e paginação
func (r *IdentityPostgresRepository) List(
	ctx context.Context, 
	filter repositories.IdentityFilter, 
	page, pageSize int,
) ([]*models.Identity, int, error) {
	// Construir query com filtros dinâmicos
	baseQuery := `
		SELECT 
			id, 
			primary_key_type, 
			primary_key_value, 
			master_person_id,
			status, 
			created_at, 
			updated_at, 
			metadata
		FROM identity_multi_context
	`
	
	countQuery := `
		SELECT COUNT(*) 
		FROM identity_multi_context
	`
	
	// Construir condições WHERE
	var conditions []string
	var params []interface{}
	paramIdx := 1
	
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
	
	// Filtro por tipo de chave primária
	if len(filter.PrimaryKeyType) > 0 {
		var keyTypeValues []string
		for _, kt := range filter.PrimaryKeyType {
			keyTypeValues = append(keyTypeValues, string(kt))
		}
		conditions = append(conditions, fmt.Sprintf("primary_key_type = ANY($%d)", paramIdx))
		params = append(params, pq.Array(keyTypeValues))
		paramIdx++
	}
	
	// Filtro por texto de busca
	if filter.SearchText != "" {
		conditions = append(conditions, fmt.Sprintf("(primary_key_value ILIKE $%d OR metadata::text ILIKE $%d)", paramIdx, paramIdx))
		params = append(params, "%"+filter.SearchText+"%")
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
			"primary_key_value": true,
			"status":            true,
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
		return nil, 0, fmt.Errorf("erro ao contar identidades: %w", err)
	}
	
	// Se não há resultados, retornar lista vazia
	if totalCount == 0 {
		return []*models.Identity{}, 0, nil
	}
	
	// Executar query de listagem
	var dbIdents []dbIdentity
	if err := r.db.SelectContext(ctx, &dbIdents, listQuery, listParams...); err != nil {
		return nil, 0, fmt.Errorf("erro ao listar identidades: %w", err)
	}
	
	// Converter para modelos de domínio
	identities := make([]*models.Identity, 0, len(dbIdents))
	for _, dbIdent := range dbIdents {
		identity, err := r.mapToModel(&dbIdent)
		if err != nil {
			return nil, 0, err
		}
		identities = append(identities, identity)
	}
	
	return identities, totalCount, nil
}

// LoadContexts carrega os contextos para uma identidade específica
func (r *IdentityPostgresRepository) LoadContexts(ctx context.Context, identity *models.Identity) error {
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
	`
	
	rows, err := r.db.QueryxContext(ctx, query, identity.ID)
	if err != nil {
		return fmt.Errorf("erro ao carregar contextos: %w", err)
	}
	defer rows.Close()
	
	contexts := make([]*models.IdentityContext, 0)
	
	for rows.Next() {
		var contextID, identityID uuid.UUID
		var contextType, status, verificationLevel string
		var trustScore float64
		var createdAt, updatedAt time.Time
		var metadataJSON []byte
		
		if err := rows.Scan(
			&contextID,
			&identityID,
			&contextType,
			&status,
			&verificationLevel,
			&trustScore,
			&createdAt,
			&updatedAt,
			&metadataJSON,
		); err != nil {
			return fmt.Errorf("erro ao escanear contexto: %w", err)
		}
		
		// Desserializar metadados
		var metadata map[string]interface{}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return fmt.Errorf("erro ao desserializar metadados: %w", err)
			}
		}
		
		// Criar contexto de domínio
		context := &models.IdentityContext{
			ID:               contextID,
			IdentityID:       identityID,
			ContextType:      models.ContextType(contextType),
			Status:           models.ContextStatus(status),
			VerificationLevel: models.VerificationLevel(verificationLevel),
			TrustScore:       trustScore,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			Metadata:         metadata,
			Attributes:       make([]*models.ContextAttribute, 0),
		}
		
		contexts = append(contexts, context)
	}
	
	if err := rows.Err(); err != nil {
		return fmt.Errorf("erro ao iterar contextos: %w", err)
	}
	
	// Atribuir contextos à identidade
	identity.Contexts = contexts
	
	return nil
}

// mapToModel converte uma estrutura de banco de dados para um modelo de domínio
func (r *IdentityPostgresRepository) mapToModel(dbIdent *dbIdentity) (*models.Identity, error) {
	// Desserializar metadados
	var metadata map[string]interface{}
	if len(dbIdent.Metadata) > 0 {
		if err := json.Unmarshal(dbIdent.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("erro ao desserializar metadados: %w", err)
		}
	}
	
	// Criar identidade
	identity := &models.Identity{
		ID:              dbIdent.ID,
		PrimaryKeyType:  models.PrimaryKeyType(dbIdent.PrimaryKeyType),
		PrimaryKeyValue: dbIdent.PrimaryKeyValue,
		Status:          models.IdentityStatus(dbIdent.Status),
		CreatedAt:       dbIdent.CreatedAt,
		UpdatedAt:       dbIdent.UpdatedAt,
		Metadata:        metadata,
		Contexts:        make([]*models.IdentityContext, 0),
	}
	
	// Converter MasterPersonID se existir
	if dbIdent.MasterPersonID.Valid {
		masterPersonID, err := uuid.Parse(dbIdent.MasterPersonID.String)
		if err == nil {
			identity.MasterPersonID = &masterPersonID
		}
	}
	
	return identity, nil
}