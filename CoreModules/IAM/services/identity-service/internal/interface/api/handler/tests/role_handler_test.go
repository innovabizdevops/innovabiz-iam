package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
	"innovabiz/iam/identity-service/internal/interface/api/handler"
)

// Mock do RoleService
type MockRoleService struct {
	mock.Mock
}

// Implementação de todos os métodos necessários para a interface RoleService
func (m *MockRoleService) CreateRole(ctx context.Context, req application.CreateRoleRequest) (*model.Role, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) GetRole(ctx context.Context, req application.GetRoleRequest) (*model.Role, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) ListRoles(ctx context.Context, req application.ListRolesRequest) (*application.ListRolesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.ListRolesResponse), args.Error(1)
}

func (m *MockRoleService) UpdateRole(ctx context.Context, req application.UpdateRoleRequest) (*model.Role, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) DeleteRole(ctx context.Context, req application.DeleteRoleRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) HardDeleteRole(ctx context.Context, req application.DeleteRoleRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) CloneRole(ctx context.Context, req application.CloneRoleRequest) (*model.Role, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleService) SyncSystemRoles(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Implementações necessárias para gerenciamento de permissões
func (m *MockRoleService) GetRolePermissions(ctx context.Context, req application.GetRolePermissionsRequest) (*application.GetRolePermissionsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.GetRolePermissionsResponse), args.Error(1)
}

func (m *MockRoleService) GetAllRolePermissions(ctx context.Context, req application.GetRolePermissionsRequest) (*application.GetRolePermissionsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.GetRolePermissionsResponse), args.Error(1)
}

func (m *MockRoleService) AssignPermissionToRole(ctx context.Context, req application.AssignPermissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) RevokePermissionFromRole(ctx context.Context, req application.RevokePermissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) CheckRoleHasPermission(ctx context.Context, req application.CheckRolePermissionRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

// Implementações necessárias para gerenciamento de hierarquia
func (m *MockRoleService) GetChildRoles(ctx context.Context, req application.GetRoleRelationsRequest) (*application.RoleRelationsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.RoleRelationsResponse), args.Error(1)
}

func (m *MockRoleService) GetParentRoles(ctx context.Context, req application.GetRoleRelationsRequest) (*application.RoleRelationsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.RoleRelationsResponse), args.Error(1)
}

func (m *MockRoleService) GetDescendantRoles(ctx context.Context, req application.GetRoleRelationsRequest) (*application.RoleRelationsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.RoleRelationsResponse), args.Error(1)
}

func (m *MockRoleService) GetAncestorRoles(ctx context.Context, req application.GetRoleRelationsRequest) (*application.RoleRelationsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.RoleRelationsResponse), args.Error(1)
}

func (m *MockRoleService) AssignChildRole(ctx context.Context, req application.AssignChildRoleRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) RemoveChildRole(ctx context.Context, req application.RemoveChildRoleRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// Implementações necessárias para gerenciamento de usuários
func (m *MockRoleService) GetRoleUsers(ctx context.Context, req application.GetRoleUsersRequest) (*application.GetRoleUsersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.GetRoleUsersResponse), args.Error(1)
}

func (m *MockRoleService) GetUserRoles(ctx context.Context, req application.GetUserRolesRequest) (*application.GetUserRolesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.GetUserRolesResponse), args.Error(1)
}

func (m *MockRoleService) AssignRoleToUser(ctx context.Context, req application.AssignRoleToUserRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) UpdateUserRoleExpiration(ctx context.Context, req application.UpdateUserRoleExpirationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) RemoveRoleFromUser(ctx context.Context, req application.RemoveRoleFromUserRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRoleService) CheckUserHasRole(ctx context.Context, req application.CheckUserRoleRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

// Mock extrator de contexto
func mockTenantAndUserExtractor(tenantID, userID uuid.UUID) func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	return func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
		return tenantID, userID, nil
	}
}

// Configuração inicial para testes
func setupTest() (*MockRoleService, *handler.RoleHandler, *mux.Router, zerolog.Logger, trace.Tracer) {
	// Configurar logger para testes
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if os.Getenv("TEST_LOG_LEVEL") != "debug" {
		logger = zerolog.Nop() // Desabilitar logs nos testes
	}

	// Configurar tracer para testes
	tracer := noop.NewTracerProvider().Tracer("")

	// Criar mock do serviço
	mockService := new(MockRoleService)

	// Criar handler
	roleHandler := handler.NewRoleHandler(mockService, logger, tracer)

	// Mock para extrator de tenant e usuário
	tenantID := uuid.New()
	userID := uuid.New()
	roleHandler.TenantAndUserExtractor = mockTenantAndUserExtractor(tenantID, userID)

	// Configurar router
	router := mux.NewRouter()
	roleHandler.RegisterRoutes(router)

	return mockService, roleHandler, router, logger, tracer
}

// TestCreateRole testa a criação de uma função
func TestCreateRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	role := &model.Role{
		ID:          roleID,
		TenantID:    tenantID,
		Name:        "TestRole",
		Description: "Test role description",
		Type:        model.RoleTypeCustom,
		Status:      model.StatusActive,
		CreatedAt:   now,
		CreatedBy:   userID,
		UpdatedAt:   now,
		UpdatedBy:   userID,
	}

	// Preparar request
	reqBody := map[string]interface{}{
		"name":        "TestRole",
		"description": "Test role description",
		"type":        "CUSTOM",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("CreateRole", mock.Anything, mock.MatchedBy(func(req application.CreateRoleRequest) bool {
		return req.Name == "TestRole" && req.Description == "Test role description" && req.Type == model.RoleTypeCustom
	})).Return(role, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusCreated, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar conteúdo da resposta
	assert.Equal(t, role.ID.String(), responseBody["id"])
	assert.Equal(t, role.Name, responseBody["name"])
	assert.Equal(t, role.Description, responseBody["description"])
	assert.Equal(t, string(role.Type), responseBody["type"])
	assert.Equal(t, string(role.Status), responseBody["status"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestGetRole testa a obtenção de uma função por ID
func TestGetRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	role := &model.Role{
		ID:          roleID,
		TenantID:    tenantID,
		Name:        "TestRole",
		Description: "Test role description",
		Type:        model.RoleTypeCustom,
		Status:      model.StatusActive,
		CreatedAt:   now,
		CreatedBy:   userID,
		UpdatedAt:   now,
		UpdatedBy:   userID,
	}

	// Expectativas do mock
	mockService.On("GetRole", mock.Anything, mock.MatchedBy(func(req application.GetRoleRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID
	})).Return(role, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s", roleID), nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar conteúdo da resposta
	assert.Equal(t, role.ID.String(), responseBody["id"])
	assert.Equal(t, role.Name, responseBody["name"])
	assert.Equal(t, role.Description, responseBody["description"])
	assert.Equal(t, string(role.Type), responseBody["type"])
	assert.Equal(t, string(role.Status), responseBody["status"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestListRoles testa a listagem de funções
func TestListRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	roles := []*model.Role{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Role1",
			Description: "Description 1",
			Type:        model.RoleTypeCustom,
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Role2",
			Description: "Description 2",
			Type:        model.RoleTypeSystem,
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	listResponse := &application.ListRolesResponse{
		Roles:      roles,
		TotalCount: 2,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("ListRoles", mock.Anything, mock.MatchedBy(func(req application.ListRolesRequest) bool {
		return req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(listResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, "/api/v1/roles?page=1&limit=10", nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar conteúdo da resposta
	rolesData, ok := responseBody["roles"].([]interface{})
	require.True(t, ok)
	assert.Len(t, rolesData, 2)
	
	assert.Equal(t, float64(2), responseBody["total_count"])
	assert.Equal(t, float64(1), responseBody["page"])
	assert.Equal(t, float64(10), responseBody["limit"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestUpdateRole testa a atualização de uma função
func TestUpdateRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	updatedRole := &model.Role{
		ID:          roleID,
		TenantID:    tenantID,
		Name:        "UpdatedRole",
		Description: "Updated description",
		Type:        model.RoleTypeCustom,
		Status:      model.StatusActive,
		CreatedAt:   now,
		CreatedBy:   userID,
		UpdatedAt:   now,
		UpdatedBy:   userID,
	}

	// Preparar request
	reqBody := map[string]interface{}{
		"name":        "UpdatedRole",
		"description": "Updated description",
		"status":      "ACTIVE",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("UpdateRole", mock.Anything, mock.MatchedBy(func(req application.UpdateRoleRequest) bool {
		return req.RoleID == roleID && 
               req.TenantID == tenantID && 
               req.Name == "UpdatedRole" && 
               req.Description == "Updated description" &&
               req.Status == model.StatusActive
	})).Return(updatedRole, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/roles/%s", roleID), bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar conteúdo da resposta
	assert.Equal(t, updatedRole.ID.String(), responseBody["id"])
	assert.Equal(t, updatedRole.Name, responseBody["name"])
	assert.Equal(t, updatedRole.Description, responseBody["description"])
	assert.Equal(t, string(updatedRole.Status), responseBody["status"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestDeleteRole testa a exclusão lógica de uma função
func TestDeleteRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Expectativas do mock
	mockService.On("DeleteRole", mock.Anything, mock.MatchedBy(func(req application.DeleteRoleRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s", roleID), nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusNoContent, rr.Code)

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestHardDeleteRole testa a exclusão permanente de uma função
func TestHardDeleteRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Expectativas do mock
	mockService.On("HardDeleteRole", mock.Anything, mock.MatchedBy(func(req application.DeleteRoleRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/hard", roleID), nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusNoContent, rr.Code)

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestCloneRole testa a clonagem de uma função
func TestCloneRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	newRoleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	newRole := &model.Role{
		ID:          newRoleID,
		TenantID:    tenantID,
		Name:        "ClonedRole",
		Description: "Cloned description",
		Type:        model.RoleTypeCustom,
		Status:      model.StatusActive,
		CreatedAt:   now,
		CreatedBy:   userID,
		UpdatedAt:   now,
		UpdatedBy:   userID,
	}

	// Preparar request
	reqBody := map[string]interface{}{
		"new_name":        "ClonedRole",
		"new_description": "Cloned description",
		"include_permissions": true,
		"include_hierarchy":   true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("CloneRole", mock.Anything, mock.MatchedBy(func(req application.CloneRoleRequest) bool {
		return req.RoleID == roleID && 
               req.TenantID == tenantID && 
               req.NewName == "ClonedRole" && 
               req.NewDescription == "Cloned description" &&
               req.IncludePermissions == true &&
               req.IncludeHierarchy == true
	})).Return(newRole, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/clone", roleID), bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusCreated, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar conteúdo da resposta
	assert.Equal(t, newRole.ID.String(), responseBody["id"])
	assert.Equal(t, newRole.Name, responseBody["name"])
	assert.Equal(t, newRole.Description, responseBody["description"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestSyncSystemRoles testa a sincronização de funções do sistema
func TestSyncSystemRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	tenantID := uuid.New()
	userID := uuid.New()

	// Expectativas do mock
	mockService.On("SyncSystemRoles", mock.Anything).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, "/api/v1/roles/sync", nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar conteúdo da resposta
	assert.Equal(t, "Funções do sistema sincronizadas com sucesso", responseBody["message"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}