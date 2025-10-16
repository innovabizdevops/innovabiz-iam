/**
 * INNOVABIZ IAM - Testes do Repositório PostgreSQL de Grupos
 * Copyright (c) 2025 INNOVABIZ
 * 
 * Implementação dos testes unitários para o repositório PostgreSQL de grupos,
 * seguindo a arquitetura multi-dimensional, multi-tenant e com observabilidade
 * total da plataforma INNOVABIZ.
 * 
 * Compliance:
 * - ISO/IEC 27001:2022 (A.5.15 - Controle de acesso)
 * - PCI DSS v4.0 (Requisito 7.2.4 - Gestão de grupos)
 * - LGPD/GDPR/PDPA (Controle de acesso)
 * - BNA Instrução 7/2021 (Segurança cibernética)
 * - SOX (Rastreabilidade e auditoria)
 * - NIST CSF (Proteção de identidade)
 */

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabiz/iam/internal/domain/group"
	"github.com/innovabiz/iam/internal/infrastructure/logging"
	"github.com/innovabiz/iam/internal/infrastructure/metrics"
	"github.com/innovabiz/iam/internal/infrastructure/tracing"
)

// setupMockDB configura um banco de dados mock para testes
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *GroupRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dbx := sqlx.NewDb(db, "sqlmock")
	
	logger := logging.NewMockLogger()
	metricClient := metrics.NewMockMetricsClient()
	tracer := tracing.NewMockTracer()

	repo := NewGroupRepository(dbx, logger, metricClient, tracer)
	return db, mock, repo
}

func TestGetByID(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()
	updatedBy := uuid.New()
	parentGroupID := uuid.New()
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()

	// Definir metadata para testes
	metadata := map[string]interface{}{
		"description": "Metadados de teste",
		"attributes": map[string]interface{}{
			"category": "admin",
			"priority": 1,
		},
	}

	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)
	metadataStr := string(metadataBytes)

	// Definir colunas esperadas na consulta
	columns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	// Configurar expectativas mock
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(groupID, tenantID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				groupID, "TEST-GROUP", "Grupo de Teste", "Descrição do grupo de teste",
				tenantID, "AO", "SYSTEM", "ACTIVE", "TEST-GROUP", 1,
				parentGroupID, createdAt, createdBy, updatedAt, updatedBy, metadataStr,
			))

	// Executar o método a ser testado
	g, err := repo.GetByID(ctx, groupID, tenantID)

	// Verificações
	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, groupID, g.ID)
	assert.Equal(t, "TEST-GROUP", g.Code)
	assert.Equal(t, "Grupo de Teste", g.Name)
	assert.Equal(t, "Descrição do grupo de teste", g.Description)
	assert.Equal(t, tenantID, g.TenantID)
	assert.Equal(t, "AO", g.RegionCode)
	assert.Equal(t, "SYSTEM", g.GroupType)
	assert.Equal(t, "ACTIVE", g.Status)
	assert.Equal(t, "TEST-GROUP", g.Path)
	assert.Equal(t, 1, g.Level)
	assert.Equal(t, parentGroupID, *g.ParentGroupID)
	assert.Equal(t, createdAt, g.CreatedAt)
	assert.Equal(t, createdBy, *g.CreatedBy)
	assert.Equal(t, updatedAt, g.UpdatedAt)
	assert.Equal(t, updatedBy, *g.UpdatedBy)
	assert.Equal(t, metadata["description"], g.Metadata["description"])

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByIDNotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	tenantID := uuid.New()

	// Configurar expectativas mock para não encontrar o grupo
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(groupID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{}))

	// Executar o método a ser testado
	g, err := repo.GetByID(ctx, groupID, tenantID)

	// Verificações
	assert.Error(t, err)
	assert.Nil(t, g)
	assert.Equal(t, group.ErrGroupNotFound, err)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByCode(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()
	parentGroupID := uuid.New()
	createdAt := time.Now().UTC()

	// Definir colunas esperadas na consulta
	columns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	// Configurar expectativas mock
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs("FINANCE", tenantID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				groupID, "FINANCE", "Finanças", "Grupo de finanças",
				tenantID, "AO", "BUSINESS", "ACTIVE", "FINANCE", 1,
				parentGroupID, createdAt, createdBy, nil, nil, nil,
			))

	// Executar o método a ser testado
	g, err := repo.GetByCode(ctx, "FINANCE", tenantID)

	// Verificações
	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, groupID, g.ID)
	assert.Equal(t, "FINANCE", g.Code)
	assert.Equal(t, "Finanças", g.Name)
	assert.Equal(t, "BUSINESS", g.GroupType)
	assert.Equal(t, "ACTIVE", g.Status)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()

	// Criar grupo de teste
	testGroup := &group.Group{
		ID:          groupID,
		Code:        "HR-TEAM",
		Name:        "Equipe de RH",
		Description: "Grupo de recursos humanos",
		TenantID:    tenantID,
		RegionCode:  "AO",
		GroupType:   "BUSINESS",
		Status:      "ACTIVE",
		CreatedBy:   &createdBy,
	}

	// Mock para verificação de código duplicado
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM iam_groups").
		WithArgs("HR-TEAM", tenantID, nil).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock para inserção
	mock.ExpectExec("INSERT INTO iam_groups").
		WithArgs(
			sqlmock.AnyArg(), "HR-TEAM", "Equipe de RH", "Grupo de recursos humanos",
			tenantID, "AO", "BUSINESS", "ACTIVE", "HR-TEAM", 1,
			nil, sqlmock.AnyArg(), createdBy, nil, nil, sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Executar o método a ser testado
	err := repo.Create(ctx, testGroup)

	// Verificações
	assert.NoError(t, err)
	assert.Equal(t, "HR-TEAM", testGroup.Path)
	assert.Equal(t, 1, testGroup.Level)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithParent(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	parentID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()

	// Criar grupo de teste
	testGroup := &group.Group{
		ID:            groupID,
		Code:          "FINANCE-AP",
		Name:          "Contas a Pagar",
		Description:   "Equipe de contas a pagar",
		TenantID:      tenantID,
		RegionCode:    "AO",
		GroupType:     "BUSINESS",
		Status:        "ACTIVE",
		ParentGroupID: &parentID,
		CreatedBy:     &createdBy,
	}

	// Mock para verificação de código duplicado
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM iam_groups").
		WithArgs("FINANCE-AP", tenantID, nil).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock para buscar grupo pai
	columns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}
	
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(parentID, tenantID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				parentID, "FINANCE", "Finanças", "Departamento financeiro",
				tenantID, "AO", "BUSINESS", "ACTIVE", "FINANCE", 1,
				nil, time.Now(), createdBy, nil, nil, nil,
			))

	// Mock para verificação de referência circular
	mock.ExpectQuery("WITH RECURSIVE group_hierarchy").
		WithArgs(groupID, parentID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock para inserção
	mock.ExpectExec("INSERT INTO iam_groups").
		WithArgs(
			sqlmock.AnyArg(), "FINANCE-AP", "Contas a Pagar", "Equipe de contas a pagar",
			tenantID, "AO", "BUSINESS", "ACTIVE", "FINANCE.FINANCE-AP", 2,
			parentID, sqlmock.AnyArg(), createdBy, nil, nil, sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Executar o método a ser testado
	err := repo.Create(ctx, testGroup)

	// Verificações
	assert.NoError(t, err)
	assert.Equal(t, "FINANCE.FINANCE-AP", testGroup.Path)
	assert.Equal(t, 2, testGroup.Level)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	tenantID := uuid.New()
	updatedBy := uuid.New()
	createdBy := uuid.New()
	createdAt := time.Now().UTC().Add(-24 * time.Hour) // criado ontem

	// Criar grupo existente para atualização
	existingGroup := &group.Group{
		ID:          groupID,
		Code:        "SALES",
		Name:        "Vendas",
		Description: "Equipe de vendas",
		TenantID:    tenantID,
		RegionCode:  "AO",
		GroupType:   "BUSINESS",
		Status:      "ACTIVE",
		Path:        "SALES",
		Level:       1,
		CreatedAt:   createdAt,
		CreatedBy:   &createdBy,
	}

	// Versão atualizada do grupo
	updatedGroup := &group.Group{
		ID:          groupID,
		Code:        "SALES",
		Name:        "Vendas Corporativas",
		Description: "Equipe de vendas para clientes corporativos",
		TenantID:    tenantID,
		RegionCode:  "AO",
		GroupType:   "BUSINESS",
		Status:      "ACTIVE",
		Path:        "SALES",
		Level:       1,
		CreatedAt:   createdAt,
		CreatedBy:   &createdBy,
		UpdatedBy:   &updatedBy,
	}

	// Definir colunas esperadas na consulta de busca do grupo existente
	columns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	// Mock para buscar o grupo existente
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(groupID, tenantID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				groupID, "SALES", "Vendas", "Equipe de vendas",
				tenantID, "AO", "BUSINESS", "ACTIVE", "SALES", 1,
				nil, createdAt, createdBy, nil, nil, nil,
			))

	// Mock para atualização
	mock.ExpectExec("UPDATE iam_groups SET").
		WithArgs(
			"SALES", "Vendas Corporativas", "Equipe de vendas para clientes corporativos",
			"AO", "BUSINESS", "ACTIVE", "SALES", 1,
			nil, sqlmock.AnyArg(), updatedBy, sqlmock.AnyArg(),
			groupID, tenantID,
		).WillReturnResult(sqlmock.NewResult(0, 1))

	// Executar o método a ser testado
	err := repo.Update(ctx, updatedGroup)

	// Verificações
	assert.NoError(t, err)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUserToGroup(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()
	addedBy := uuid.New()

	// Mock para verificação se o grupo existe
	groupColumns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(groupID, tenantID).
		WillReturnRows(sqlmock.NewRows(groupColumns).
			AddRow(
				groupID, "ADMIN", "Administradores", "Grupo de administradores",
				tenantID, "AO", "SYSTEM", "ACTIVE", "ADMIN", 1,
				nil, time.Now(), addedBy, nil, nil, nil,
			))

	// Mock para verificar se o usuário já está no grupo
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM iam_group_members").
		WithArgs(groupID, userID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock para adicionar o usuário ao grupo
	mock.ExpectExec("INSERT INTO iam_group_members").
		WithArgs(
			sqlmock.AnyArg(), groupID, userID, tenantID,
			sqlmock.AnyArg(), addedBy, nil, nil, nil, nil,
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Executar o método a ser testado
	err := repo.AddUserToGroup(ctx, groupID, userID, tenantID, &addedBy)

	// Verificações
	assert.NoError(t, err)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRemoveUserFromGroup(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()
	removedBy := uuid.New()

	// Mock para verificação se o grupo existe
	groupColumns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(groupID, tenantID).
		WillReturnRows(sqlmock.NewRows(groupColumns).
			AddRow(
				groupID, "ADMIN", "Administradores", "Grupo de administradores",
				tenantID, "AO", "SYSTEM", "ACTIVE", "ADMIN", 1,
				nil, time.Now(), removedBy, nil, nil, nil,
			))

	// Mock para verificar se o usuário está no grupo
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM iam_group_members").
		WithArgs(groupID, userID, tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock para remover o usuário do grupo
	mock.ExpectExec("UPDATE iam_group_members").
		WithArgs(
			sqlmock.AnyArg(), removedBy, groupID, userID, tenantID,
		).WillReturnResult(sqlmock.NewResult(0, 1))

	// Executar o método a ser testado
	err := repo.RemoveUserFromGroup(ctx, groupID, userID, tenantID, &removedBy)

	// Verificações
	assert.NoError(t, err)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := uuid.New()
	groupID1 := uuid.New()
	groupID2 := uuid.New()
	createdBy := uuid.New()

	// Filtro para teste
	filter := group.GroupFilter{
		SearchTerm:    "Admin",
		Statuses:      []string{"ACTIVE"},
		SortBy:        "name",
		SortDirection: "ASC",
	}

	// Mock para contar total de registros
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs(tenantID, "%Admin%", "ACTIVE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Definir colunas esperadas na consulta de listagem
	columns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	// Mock para listar grupos
	mock.ExpectQuery("SELECT g.id, g.code, g.name").
		WithArgs(tenantID, "%Admin%", "ACTIVE", 10, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				groupID1, "ADMIN", "Administradores", "Grupo de administradores",
				tenantID, "AO", "SYSTEM", "ACTIVE", "ADMIN", 1,
				nil, time.Now(), createdBy, nil, nil, nil,
			).
			AddRow(
				groupID2, "USER-ADMIN", "Administradores de Usuários", "Gestores de usuários",
				tenantID, "AO", "BUSINESS", "ACTIVE", "USER-ADMIN", 1,
				nil, time.Now(), createdBy, nil, nil, nil,
			))

	// Executar o método a ser testado
	result, err := repo.List(ctx, tenantID, filter, 1, 10)

	// Verificações
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 2, len(result.Groups))
	assert.Equal(t, "ADMIN", result.Groups[0].Code)
	assert.Equal(t, "USER-ADMIN", result.Groups[1].Code)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetParentGroup(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	ctx := context.Background()
	groupID := uuid.New()
	parentID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()

	// Definir colunas esperadas
	columns := []string{
		"id", "code", "name", "description", "tenant_id",
		"region_code", "group_type", "status", "path", "level",
		"parent_group_id", "created_at", "created_by",
		"updated_at", "updated_by", "metadata",
	}

	// Mock para buscar o grupo filho
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(groupID, tenantID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				groupID, "FINANCE-AP", "Contas a Pagar", "Equipe de contas a pagar",
				tenantID, "AO", "BUSINESS", "ACTIVE", "FINANCE.FINANCE-AP", 2,
				parentID, time.Now(), createdBy, nil, nil, nil,
			))

	// Mock para buscar o grupo pai
	mock.ExpectQuery("SELECT (.+) FROM iam_groups").
		WithArgs(parentID, tenantID).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(
				parentID, "FINANCE", "Finanças", "Departamento financeiro",
				tenantID, "AO", "BUSINESS", "ACTIVE", "FINANCE", 1,
				nil, time.Now(), createdBy, nil, nil, nil,
			))

	// Executar o método a ser testado
	parent, err := repo.GetParentGroup(ctx, groupID, tenantID)

	// Verificações
	require.NoError(t, err)
	assert.NotNil(t, parent)
	assert.Equal(t, parentID, parent.ID)
	assert.Equal(t, "FINANCE", parent.Code)
	assert.Equal(t, "Finanças", parent.Name)
	assert.Equal(t, 1, parent.Level)

	// Verificar se todas as expectativas foram atendidas
	assert.NoError(t, mock.ExpectationsWereMet())
}