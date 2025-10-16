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

// TestGetChildRoles testa a obtenção das funções filhas diretas de uma função
func TestGetChildRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	childRoles := []*model.Role{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "ChildRole1",
			Description: "Child Role 1",
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
			Name:        "ChildRole2",
			Description: "Child Role 2",
			Type:        model.RoleTypeCustom,
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	relationResponse := &application.RoleRelationsResponse{
		Roles:      childRoles,
		TotalCount: 2,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("GetChildRoles", mock.Anything, mock.MatchedBy(func(req application.GetRoleRelationsRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(relationResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/children?page=1&limit=10", roleID), nil)
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

// TestGetParentRoles testa a obtenção das funções pais diretas de uma função
func TestGetParentRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	parentRoles := []*model.Role{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "ParentRole1",
			Description: "Parent Role 1",
			Type:        model.RoleTypeCustom,
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	relationResponse := &application.RoleRelationsResponse{
		Roles:      parentRoles,
		TotalCount: 1,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("GetParentRoles", mock.Anything, mock.MatchedBy(func(req application.GetRoleRelationsRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(relationResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/parents?page=1&limit=10", roleID), nil)
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
	assert.Len(t, rolesData, 1)

	assert.Equal(t, float64(1), responseBody["total_count"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestGetDescendantRoles testa a obtenção das funções descendentes (hierarquia completa) de uma função
func TestGetDescendantRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	// Criando uma hierarquia com vários níveis para testar descendência
	descendantRoles := []*model.Role{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "ChildRole",
			Description: "Child Role",
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
			Name:        "GrandchildRole",
			Description: "Grandchild Role",
			Type:        model.RoleTypeCustom,
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	relationResponse := &application.RoleRelationsResponse{
		Roles:      descendantRoles,
		TotalCount: 2,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("GetDescendantRoles", mock.Anything, mock.MatchedBy(func(req application.GetRoleRelationsRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(relationResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/descendants?page=1&limit=10", roleID), nil)
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

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestGetAncestorRoles testa a obtenção das funções ancestrais (hierarquia completa para cima) de uma função
func TestGetAncestorRoles(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	roleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	// Criando ancestrais para testar
	ancestorRoles := []*model.Role{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "ParentRole",
			Description: "Parent Role",
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
			Name:        "GrandparentRole",
			Description: "Grandparent Role",
			Type:        model.RoleTypeSystem,
			Status:      model.StatusActive,
			CreatedAt:   now,
			CreatedBy:   userID,
			UpdatedAt:   now,
			UpdatedBy:   userID,
		},
	}

	relationResponse := &application.RoleRelationsResponse{
		Roles:      ancestorRoles,
		TotalCount: 2,
		Page:       1,
		Limit:      10,
	}

	// Expectativas do mock
	mockService.On("GetAncestorRoles", mock.Anything, mock.MatchedBy(func(req application.GetRoleRelationsRequest) bool {
		return req.RoleID == roleID && req.TenantID == tenantID && req.Page == 1 && req.Limit == 10
	})).Return(relationResponse, nil)

	// Executar request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/roles/%s/ancestors?page=1&limit=10", roleID), nil)
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

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestAssignChildRole testa a atribuição de uma função filha a uma função pai
func TestAssignChildRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	parentRoleID := uuid.New()
	childRoleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Preparar request
	reqBody := map[string]interface{}{
		"child_role_id": childRoleID.String(),
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Expectativas do mock
	mockService.On("AssignChildRole", mock.Anything, mock.MatchedBy(func(req application.AssignChildRoleRequest) bool {
		return req.ParentRoleID == parentRoleID && 
               req.ChildRoleID == childRoleID && 
               req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/children", parentRoleID), bytes.NewBuffer(jsonBody))
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
	assert.Equal(t, "Função filha atribuída com sucesso", responseBody["message"])

	// Verificar que o mock foi chamado
	mockService.AssertExpectations(t)
}

// TestRemoveChildRole testa a remoção de uma função filha de uma função pai
func TestRemoveChildRole(t *testing.T) {
	// Configuração
	mockService, _, router, _, _ := setupTest()

	// Dados de teste
	parentRoleID := uuid.New()
	childRoleID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()

	// Expectativas do mock
	mockService.On("RemoveChildRole", mock.Anything, mock.MatchedBy(func(req application.RemoveChildRoleRequest) bool {
		return req.ParentRoleID == parentRoleID && 
               req.ChildRoleID == childRoleID && 
               req.TenantID == tenantID
	})).Return(nil)

	// Executar request
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/children/%s", parentRoleID, childRoleID), nil)
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

// TestHierarchyErrorHandling testa cenários de erro em operações de hierarquia
func TestHierarchyErrorHandling(t *testing.T) {
	// Testes para casos de erro na manipulação de hierarquia

	// 1. Teste para tentativa de criar ciclo na hierarquia
	t.Run("Prevent cycle in hierarchy", func(t *testing.T) {
		// Configuração
		mockService := new(MockRoleService)
		_, h, router, _, _ := setupTest()
		h.RoleService = mockService

		// Dados de teste
		parentRoleID := uuid.New()
		childRoleID := uuid.New()  // Este seria um ancestral do parentRoleID na hierarquia real
		tenantID := uuid.New()
		userID := uuid.New()

		// Simular erro de ciclo
		hierarchyError := fmt.Errorf("ciclo detectado na hierarquia de funções")

		// Preparar request
		reqBody := map[string]interface{}{
			"child_role_id": childRoleID.String(),
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Expectativas do mock - retornar erro
		mockService.On("AssignChildRole", mock.Anything, mock.Anything).Return(hierarchyError)

		// Executar request
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/children", parentRoleID), bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID.String())
		req.Header.Set("X-User-ID", userID.String())

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Verificar resultado - deve ser erro
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var responseBody map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		// Verificar mensagem de erro
		assert.Contains(t, responseBody["message"], "ciclo detectado")

		// Verificar que o mock foi chamado
		mockService.AssertExpectations(t)
	})

	// 2. Teste para tentativa de adicionar uma função filha inválida (UUID inválido)
	t.Run("Invalid child role UUID", func(t *testing.T) {
		// Configuração
		_, _, router, _, _ := setupTest()

		// Dados de teste
		parentRoleID := uuid.New()
		invalidChildID := "invalid-uuid"
		tenantID := uuid.New()
		userID := uuid.New()

		// Preparar request com ID de função filho inválido
		reqBody := map[string]interface{}{
			"child_role_id": invalidChildID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Executar request
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/children", parentRoleID), bytes.NewBuffer(jsonBody))
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
	})

	// 3. Teste para tentativa de adicionar uma função filha sem fornecer seu ID
	t.Run("Missing child role ID", func(t *testing.T) {
		// Configuração
		_, _, router, _, _ := setupTest()

		// Dados de teste
		parentRoleID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		// Preparar request com corpo vazio (sem child_role_id)
		reqBody := map[string]interface{}{}
		jsonBody, _ := json.Marshal(reqBody)

		// Executar request
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/roles/%s/children", parentRoleID), bytes.NewBuffer(jsonBody))
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
		assert.Contains(t, responseBody["message"], "child_role_id")
	})
}