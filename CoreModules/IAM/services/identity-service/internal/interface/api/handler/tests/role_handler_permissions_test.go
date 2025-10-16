package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
)

// TestGetRolePermissions testa a obtenção das permissões diretas de uma função
func TestGetRolePermissions(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	permissions := []*model.Permission{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "perm1",
			Description: "Permission 1",
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "perm2",
			Description: "Permission 2",
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	permissionsResponse := &application.GetRolePermissionsResponse{
		Permissions: permissions,
		TotalCount:  2,
		Page:        1,
		Limit:       10,
	}

	// Expectativas do mock
	mockService.On("GetRolePermissions", mock.Anything, mock.MatchedBy(func(req application.GetRolePermissionsRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(permissionsResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/permissions?page=1&limit=10", roleID), nil)
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
	permData, ok := responseBody["permissions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, permData, 2)

	assert.Equal(t, float64(2), responseBody["total_count"])
	assert.Equal(t, float64(1), responseBody["page"])
	assert.Equal(t, float64(10), responseBody["limit"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestGetAllRolePermissions testa a obtenção de todas as permissões (diretas e herdadas) de uma função
func TestGetAllRolePermissions(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	// Criar uma mistura de permissões diretas e herdadas
	permissions := []*model.Permission{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "direct_perm",
			Description: "Direct Permission",
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "inherited_perm",
			Description: "Inherited Permission",
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	permissionsResponse := &application.GetRolePermissionsResponse{
		Permissions: permissions,
		TotalCount:  2,
		Page:        1,
		Limit:       10,
	}

	// Expectativas do mock
	mockService.On("GetAllRolePermissions", mock.Anything, mock.MatchedBy(func(req application.GetRolePermissionsRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(permissionsResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/permissions/all?page=1&limit=10", roleID), nil)
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
	permData, ok := responseBody["permissions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, permData, 2)

	assert.Equal(t, float64(2), responseBody["total_count"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestAssignPermissionToRole testa a atribuição de uma permissão a uma função
func TestAssignPermissionToRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	permissionID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Preparar request
	reqBody := map[string]interface{}{
		"permission_id": permissionID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("AssignPermissionToRole", mock.Anything, mock.MatchedBy(func(req application.AssignPermissionRequest) bool {
		return req.RoleID == roleID && req.PermissionID == permissionID && req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", roleID), bytes.NewBuffer(jsonBody))
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

	// Verificar mensagem de sucesso
	assert.Equal(t, "Permissão atribuída com sucesso", responseBody["message"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestRevokePermissionFromRole testa a revogação de uma permissão de uma função
func TestRevokePermissionFromRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	permissionID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Expectativas do mock
	mockService.On("RevokePermissionFromRole", mock.Anything, mock.MatchedBy(func(req application.RevokePermissionRequest) bool {
		return req.RoleID == roleID && req.PermissionID == permissionID && req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/permissions/%s", roleID, permissionID), nil)
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

// TestCheckRoleHasPermission testa a verificação se uma função tem uma permissão específica
func TestCheckRoleHasPermission(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	permissionID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Casos de teste
	testCases := []struct {
		name           string
		hasPermission  bool
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Has permission",
			hasPermission:  true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"has_permission": true,
			},
		},
		{
			name:           "Does not have permission",
			hasPermission:  false,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"has_permission": false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock
			mockService := new(MockRoleService)

			// Expectativas do mock
			mockService.On("CheckRoleHasPermission", mock.Anything, mock.MatchedBy(func(req application.CheckRolePermissionRequest) bool {
				return req.RoleID == roleID && req.PermissionID == permissionID && req.TenantID == tenantID
			})).Return(tc.hasPermission, nil)

			// Criar handler e router
			_, h, router, _, _ := setupTest()
			h.RoleService = mockService

			// Executar request
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/permissions/%s", roleID, permissionID), nil)
			require.NoError(t, err)
			req.Header.Set("X-Tenant-ID", tenantID.String())
			req.Header.Set("X-User-ID", userID.String())

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Verificar resultado
			require.Equal(t, tc.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			// Verificar conteúdo da resposta
			assert.Equal(t, tc.expectedBody["has_permission"], responseBody["has_permission"])

			// Verificar que o mock foi chamado
			mockService.AssertExpectations(t)
		})
	}
}

// TestErrorHandling testa o tratamento de erros para operações de permissão
func TestPermissionErrorHandling(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	invalidPermissionID := "invalid-uuid"
	tenantID := uuid.New()
	userID := uuid.New()

	// Preparar request com ID de permissão inválido
	reqBody := map[string]interface{}{
		"permission_id": invalidPermissionID,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", roleID), bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado (deve falhar com erro de validação)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar mensagem de erro
	assert.Contains(t, responseBody["message"], "UUID inválido")
}

// TestAssignPermissionMissingID testa a tentativa de atribuir uma permissão sem fornecer seu ID
func TestAssignPermissionMissingID(t *testing.T) {
	// Configuração
	_, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Preparar request com corpo vazio (sem permission_id)
	reqBody := map[string]interface{}{}
	jsonBody, _ := json.Marshal(reqBody)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/permissions", roleID), bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado (deve falhar com erro de validação)
	require.Equal(t, http.StatusBadRequest, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar mensagem de erro
	assert.Contains(t, responseBody["message"], "permission_id")
}