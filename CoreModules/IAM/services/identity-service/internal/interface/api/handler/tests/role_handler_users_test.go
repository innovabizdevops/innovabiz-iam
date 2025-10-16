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

// TestGetRoleUsers testa a obtenção de usuários atribuídos a uma função específica
func TestGetRoleUsers(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	expiration := now.Add(24 * time.Hour)

	// Simulando usuários atribuídos a uma função
	userRoles := []*model.UserRole{
		{
			UserID:      uuid.New(),
			RoleID:      roleID,
			TenantID:    tenantID,
			ExpiresAt:   &expiration,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
			Username:    "user1@example.com",
			DisplayName: "User One",
		},
		{
			UserID:      uuid.New(),
			RoleID:      roleID,
			TenantID:    tenantID,
			ExpiresAt:   nil, // Sem expiração
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
			Username:    "user2@example.com",
			DisplayName: "User Two",
		},
	}

	roleUsersResponse := &application.GetRoleUsersResponse{
		Users:      userRoles,
		TotalCount: 2,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("GetRoleUsers", mock.Anything, mock.MatchedBy(func(req application.GetRoleUsersRequest) bool {
		return req.RoleID == roleID && 
               req.TenantID == tenantID && 
               req.Page == 1 && 
               req.Limit == 10
	})).Return(roleUsersResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/users?page=1&limit=10", roleID), nil)
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
	usersData, ok := responseBody["users"].([]interface{})
	require.True(t, ok)
	assert.Len(t, usersData, 2)
	assert.Equal(t, float64(2), responseBody["total_count"])
	assert.Equal(t, float64(1), responseBody["page"])
	assert.Equal(t, float64(10), responseBody["limit"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestGetUserRoles testa a obtenção de funções atribuídas a um usuário específico
func TestGetUserRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	userIDParam := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	expiration := now.Add(24 * time.Hour)

	// Simulando funções atribuídas a um usuário
	userRoles := []*model.UserRole{
		{
			UserID:      userIDParam,
			RoleID:      uuid.New(),
			TenantID:    tenantID,
			ExpiresAt:   &expiration,
			CreatedAt:   now,
			CreatedBy:   currentUserID,
			UpdatedAt:   now,
			UpdatedBy:   currentUserID,
			RoleName:    "Admin",
			Description: "Administrator role",
		},
		{
			UserID:      userIDParam,
			RoleID:      uuid.New(),
			TenantID:    tenantID,
			ExpiresAt:   nil, // Sem expiração
			CreatedAt:   now,
			CreatedBy:   currentUserID,
			UpdatedAt:   now,
			UpdatedBy:   currentUserID,
			RoleName:    "User",
			Description: "Regular user role",
		},
	}

	userRolesResponse := &application.GetUserRolesResponse{
		Roles:      userRoles,
		TotalCount: 2,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("GetUserRoles", mock.Anything, mock.MatchedBy(func(req application.GetUserRolesRequest) bool {
		return req.UserID == userIDParam && 
               req.TenantID == tenantID && 
               req.Page == 1 && 
               req.Limit == 10
	})).Return(userRolesResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/users/%s?page=1&limit=10", userIDParam), nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", currentUserID.String())

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

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestAssignRoleToUser testa a atribuição de uma função a um usuário
func TestAssignRoleToUser(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	userIDToAssign := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()
	expiration := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second) // 30 dias no futuro

	// Preparar request
	reqBody := map[string]interface{}{
		"user_id":    userIDToAssign.String(),
		"expires_at": expiration.Format(time.RFC3339),
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("AssignRoleToUser", mock.Anything, mock.MatchedBy(func(req application.AssignRoleToUserRequest) bool {
		return req.RoleID == roleID && 
               req.UserID == userIDToAssign && 
               req.TenantID == tenantID &&
               req.ExpiresAt != nil &&
               req.ExpiresAt.Unix() == expiration.Unix()
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/users", roleID), bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", currentUserID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar mensagem de sucesso
	assert.Equal(t, "Função atribuída ao usuário com sucesso", responseBody["message"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestAssignRoleToUserWithoutExpiration testa a atribuição de uma função a um usuário sem data de expiração
func TestAssignRoleToUserWithoutExpiration(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	userIDToAssign := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()

	// Preparar request sem expiração
	reqBody := map[string]interface{}{
		"user_id": userIDToAssign.String(),
		// Sem expires_at
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock - deve passar nil para ExpiresAt
	mockService.On("AssignRoleToUser", mock.Anything, mock.MatchedBy(func(req application.AssignRoleToUserRequest) bool {
		return req.RoleID == roleID && 
               req.UserID == userIDToAssign && 
               req.TenantID == tenantID &&
               req.ExpiresAt == nil
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/users", roleID), bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", currentUserID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestUpdateUserRoleExpiration testa a atualização da data de expiração de uma função para um usuário
func TestUpdateUserRoleExpiration(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	userIDToUpdate := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()
	newExpiration := time.Now().UTC().Add(60 * 24 * time.Hour).Truncate(time.Second) // 60 dias no futuro

	// Preparar request
	reqBody := map[string]interface{}{
		"expires_at": newExpiration.Format(time.RFC3339),
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("UpdateUserRoleExpiration", mock.Anything, mock.MatchedBy(func(req application.UpdateUserRoleExpirationRequest) bool {
		return req.RoleID == roleID && 
               req.UserID == userIDToUpdate && 
               req.TenantID == tenantID &&
               req.ExpiresAt != nil &&
               req.ExpiresAt.Unix() == newExpiration.Unix()
	})).Return(nil)

	// Executar request
	reqPath := fmt.Sprintf("/api/v1/roles/%s/users/%s", roleID, userIDToUpdate)
	req, err := http.NewRequest(http.MethodPatch, reqPath, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", currentUserID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Verificar mensagem de sucesso
	assert.Equal(t, "Data de expiração atualizada com sucesso", responseBody["message"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestRemoveRoleExpiration testa a remoção da data de expiração (definindo como null)
func TestRemoveRoleExpiration(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	userIDToUpdate := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()

	// Preparar request para remover expiração (null)
	reqBody := map[string]interface{}{
		"expires_at": nil,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock - deve passar nil para ExpiresAt
	mockService.On("UpdateUserRoleExpiration", mock.Anything, mock.MatchedBy(func(req application.UpdateUserRoleExpirationRequest) bool {
		return req.RoleID == roleID && 
               req.UserID == userIDToUpdate && 
               req.TenantID == tenantID &&
               req.ExpiresAt == nil
	})).Return(nil)

	// Executar request
	reqPath := fmt.Sprintf("/api/v1/roles/%s/users/%s", roleID, userIDToUpdate)
	req, err := http.NewRequest(http.MethodPatch, reqPath, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", currentUserID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusOK, rr.Code)

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestRemoveRoleFromUser testa a remoção de uma função de um usuário
func TestRemoveRoleFromUser(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	userIDToRemove := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()

	// Expectativas do mock
	mockService.On("RemoveRoleFromUser", mock.Anything, mock.MatchedBy(func(req application.RemoveRoleFromUserRequest) bool {
		return req.RoleID == roleID && 
               req.UserID == userIDToRemove && 
               req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	reqPath := fmt.Sprintf("/api/v1/roles/%s/users/%s", roleID, userIDToRemove)
	req, err := http.NewRequest(http.MethodDelete, reqPath, nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", currentUserID.String())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Verificar resultado
	require.Equal(t, http.StatusNoContent, rr.Code)

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestCheckUserHasRole testa a verificação se um usuário tem uma função específica
func TestCheckUserHasRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	userIDToCheck := uuid.New()
	tenantID := uuid.New()
	currentUserID := uuid.New()

	// Casos de teste
	testCases := []struct {
		name           string
		hasRole        bool
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "User has role",
			hasRole:        true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"has_role": true,
			},
		},
		{
			name:           "User does not have role",
			hasRole:        false,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"has_role": false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock
			mockService := new(MockRoleService)

			// Expectativas do mock
			mockService.On("CheckUserHasRole", mock.Anything, mock.MatchedBy(func(req application.CheckUserRoleRequest) bool {
				return req.RoleID == roleID && 
                       req.UserID == userIDToCheck && 
                       req.TenantID == tenantID
			})).Return(tc.hasRole, nil)

			// Criar handler e router
			_, h, router, _, _ := setupTest()
			h.RoleService = mockService

			// Executar request
			reqPath := fmt.Sprintf("/api/v1/roles/%s/users/%s/check", roleID, userIDToCheck)
			req, err := http.NewRequest(http.MethodGet, reqPath, nil)
			require.NoError(t, err)
			req.Header.Set("X-Tenant-ID", tenantID.String())
			req.Header.Set("X-User-ID", currentUserID.String())

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Verificar resultado
			require.Equal(t, tc.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			// Verificar conteúdo da resposta
			assert.Equal(t, tc.expectedBody["has_role"], responseBody["has_role"])

			// Verificar que o mock foi chamado
			mockService.AssertExpectations(t)
		})
	}
}

// TestUserRoleErrorHandling testa cenários de erro em operações de atribuição de funções a usuários
func TestUserRoleErrorHandling(t *testing.T) {
	// 1. Teste para tentativa de atribuir função a um usuário com UUID inválido
	t.Run("Invalid user UUID", func(t *testing.T) {
		// Configuração
		_, _, router, _, _ := setupTest()

		// Dados de teste
		roleID := uuid.New()
		invalidUserID := "invalid-uuid"
		tenantID := uuid.New()
		currentUserID := uuid.New()

		// Preparar request com ID de usuário inválido
		reqBody := map[string]interface{}{
			"user_id": invalidUserID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Executar request
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/users", roleID), bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID.String())
		req.Header.Set("X-User-ID", currentUserID.String())

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Verificar resultado (deve falhar com erro de validação)
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var responseBody map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		// Verificar mensagem de erro
		assert.Contains(t, responseBody["message"], "UUID inválido")
	})

	// 2. Teste para tentativa de atribuir função a um usuário sem fornecer ID do usuário
	t.Run("Missing user ID", func(t *testing.T) {
		// Configuração
		_, _, router, _, _ := setupTest()

		// Dados de teste
		roleID := uuid.New()
		tenantID := uuid.New()
		currentUserID := uuid.New()

		// Preparar request com corpo vazio (sem user_id)
		reqBody := map[string]interface{}{}
		jsonBody, _ := json.Marshal(reqBody)

		// Executar request
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/users", roleID), bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID.String())
		req.Header.Set("X-User-ID", currentUserID.String())

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Verificar resultado (deve falhar com erro de validação)
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var responseBody map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		// Verificar mensagem de erro
		assert.Contains(t, responseBody["message"], "user_id")
	})

	// 3. Teste para formato de data de expiração inválido
	t.Run("Invalid expiration date format", func(t *testing.T) {
		// Configuração
		_, _, router, _, _ := setupTest()

		// Dados de teste
		roleID := uuid.New()
		userIDToAssign := uuid.New()
		tenantID := uuid.New()
		currentUserID := uuid.New()

		// Preparar request com formato de data inválido
		reqBody := map[string]interface{}{
			"user_id":    userIDToAssign.String(),
			"expires_at": "2023-13-32T25:61:61Z", // Data inválida
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Executar request
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/users", roleID), bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID.String())
		req.Header.Set("X-User-ID", currentUserID.String())

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Verificar resultado (deve falhar com erro de validação)
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var responseBody map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		// Verificar mensagem de erro
		assert.Contains(t, responseBody["message"], "formato de data")
	})
}