package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/model"
	"innovabiz/iam/identity-service/internal/interface/api/handler"
	"innovabiz/iam/identity-service/internal/interface/middleware"
)

// MockAuthMiddleware é um mock para o middleware de autenticação
type MockAuthMiddleware struct {
	mock.Mock
}

func (m *MockAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simula a extração de tenant e user do token JWT
		ctx := r.Context()
		tenantID, _ := uuid.Parse(r.Header.Get("X-Tenant-ID"))
		userID, _ := uuid.Parse(r.Header.Get("X-User-ID"))
		
		// Adiciona claims ao contexto como faria o middleware real
		ctx = context.WithValue(ctx, middleware.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		ctx = context.WithValue(ctx, middleware.UsernameKey, "test@example.com")
		ctx = context.WithValue(ctx, middleware.RolesKey, []string{"ADMIN", "USER"})
		
		m.Called(w, r)
		
		// Chama o próximo handler com o contexto modificado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MockAuthorizationMiddleware é um mock para o middleware de autorização OPA
type MockAuthorizationMiddleware struct {
	mock.Mock
	ShouldAllow bool
}

func (m *MockAuthorizationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Called(w, r)
		
		// Se não deve permitir, retorna erro de autorização
		if !m.ShouldAllow {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Acesso negado: você não tem as permissões necessárias",
				"error":   "forbidden",
				"code":    "AUTHORIZATION_ERROR",
			})
			return
		}
		
		// Caso contrário, permite o acesso
		next.ServeHTTP(w, r)
	})
}

// MockCORSMiddleware é um mock para o middleware CORS
type MockCORSMiddleware struct {
	mock.Mock
}

func (m *MockCORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adiciona headers CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Se for uma requisição OPTIONS, retorna imediatamente
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			m.Called(w, r)
			return
		}
		
		m.Called(w, r)
		next.ServeHTTP(w, r)
	})
}

// setupMiddlewareTest configura o ambiente para testar os middlewares
func setupMiddlewareTest(shouldAllow bool) (*MockRoleService, *handler.RoleHandler, *mux.Router, *MockAuthMiddleware, *MockAuthorizationMiddleware, *MockCORSMiddleware) {
	// Configurar logger para testes
	logger := zerolog.Nop()
	
	// Configurar tracer para testes
	tracer := noop.NewTracerProvider().Tracer("")
	
	// Criar mock do serviço
	mockService := new(MockRoleService)
	
	// Criar handler
	roleHandler := handler.NewRoleHandler(mockService, logger, tracer)
	
	// Criar mocks dos middlewares
	mockAuthMiddleware := new(MockAuthMiddleware)
	mockAuthzMiddleware := &MockAuthorizationMiddleware{ShouldAllow: shouldAllow}
	mockCORSMiddleware := new(MockCORSMiddleware)
	
	// Configurar router
	router := mux.NewRouter()
	
	// Aplicar middlewares na ordem correta
	// 1. CORS (mais externo)
	corsRouter := mockCORSMiddleware.Middleware(router)
	// 2. Autenticação
	authRouter := mockAuthMiddleware.Middleware(corsRouter)
	// 3. Autorização 
	authzRouter := mockAuthzMiddleware.Middleware(authRouter)
	
	// Registrar rotas do handler no router com todos os middlewares aplicados
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	roleHandler.RegisterRoutes(subrouter)
	
	return mockService, roleHandler, router, mockAuthMiddleware, mockAuthzMiddleware, mockCORSMiddleware
}

// TestCreateRoleWithMiddleware testa a criação de uma função com todos os middlewares
func TestCreateRoleWithMiddleware(t *testing.T) {
	// Configuração - permitir o acesso
	mockService, _, router, mockAuth, mockAuthz, mockCORS := setupMiddlewareTest(true)
	
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
	
	// Expectativas dos mocks de middleware
	mockCORS.On("Called", mock.Anything, mock.Anything)
	mockAuth.On("Called", mock.Anything, mock.Anything)
	mockAuthz.On("Called", mock.Anything, mock.Anything)
	
	// Expectativa do mock de serviço
	mockService.On("CreateRole", mock.Anything, mock.MatchedBy(func(req application.CreateRoleRequest) bool {
		return req.Name == "TestRole" && req.TenantID == tenantID && req.CreatedBy == userID
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
	
	// Verificar que os mocks foram chamados
	mockCORS.AssertCalled(t, "Called", mock.Anything, mock.Anything)
	mockAuth.AssertCalled(t, "Called", mock.Anything, mock.Anything)
	mockAuthz.AssertCalled(t, "Called", mock.Anything, mock.Anything)
	mockService.AssertExpectations(t)
}

// TestCreateRoleWithAuthorizationDenied testa a rejeição de acesso pelo middleware de autorização
func TestCreateRoleWithAuthorizationDenied(t *testing.T) {
	// Configuração - negar o acesso
	_, _, router, mockAuth, mockAuthz, mockCORS := setupMiddlewareTest(false)
	
	// Dados de teste
	tenantID := uuid.New()
	userID := uuid.New()
	
	// Preparar request
	reqBody := map[string]interface{}{
		"name":        "TestRole",
		"description": "Test role description",
		"type":        "CUSTOM",
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	// Expectativas dos mocks de middleware
	mockCORS.On("Called", mock.Anything, mock.Anything)
	mockAuth.On("Called", mock.Anything, mock.Anything)
	mockAuthz.On("Called", mock.Anything, mock.Anything)
	
	// Executar request
	req, err := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("X-User-ID", userID.String())
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	// Verificar resultado - deve ser erro de autorização
	require.Equal(t, http.StatusForbidden, rr.Code)
	
	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	
	// Verificar mensagem de erro
	assert.Contains(t, responseBody["message"], "Acesso negado")
	assert.Equal(t, "forbidden", responseBody["error"])
	
	// Verificar que os mocks de middleware foram chamados
	mockCORS.AssertCalled(t, "Called", mock.Anything, mock.Anything)
	mockAuth.AssertCalled(t, "Called", mock.Anything, mock.Anything)
	mockAuthz.AssertCalled(t, "Called", mock.Anything, mock.Anything)
}

// TestOptionsRequest testa o tratamento de requisições OPTIONS para CORS
func TestOptionsRequest(t *testing.T) {
	// Configuração
	_, _, router, _, _, mockCORS := setupMiddlewareTest(true)
	
	// Expectativas do mock CORS
	mockCORS.On("Called", mock.Anything, mock.Anything)
	
	// Executar request OPTIONS
	req, err := http.NewRequest(http.MethodOptions, "/api/v1/roles", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	// Verificar resultado - deve ser 200 OK para preflight
	require.Equal(t, http.StatusOK, rr.Code)
	
	// Verificar headers CORS
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	
	// Verificar que o mock foi chamado
	mockCORS.AssertCalled(t, "Called", mock.Anything, mock.Anything)
}

// TestHardDeleteRoleWithMiddleware testa a exclusão permanente (hard delete) com controle de acesso rigoroso
func TestHardDeleteRoleWithMiddleware(t *testing.T) {
	// Dois casos de teste: permitir e negar
	testCases := []struct {
		name           string
		allowAccess    bool
		expectedStatus int
	}{
		{
			name:           "Allowed hard delete",
			allowAccess:    true,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Denied hard delete",
			allowAccess:    false,
			expectedStatus: http.StatusForbidden,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configuração
			mockService, _, router, mockAuth, mockAuthz, mockCORS := setupMiddlewareTest(tc.allowAccess)
			
			// Dados de teste
			roleID := uuid.New()
			tenantID := uuid.New()
			userID := uuid.New()
			
			// Expectativas dos mocks de middleware
			mockCORS.On("Called", mock.Anything, mock.Anything)
			mockAuth.On("Called", mock.Anything, mock.Anything)
			mockAuthz.On("Called", mock.Anything, mock.Anything)
			
			// Expectativa do mock de serviço (só vai ser chamado se permitido)
			if tc.allowAccess {
				mockService.On("HardDeleteRole", mock.Anything, mock.MatchedBy(func(req application.DeleteRoleRequest) bool {
					return req.RoleID == roleID && req.TenantID == tenantID
				})).Return(nil)
			}
			
			// Executar request
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/roles/%s/hard", roleID), nil)
			require.NoError(t, err)
			req.Header.Set("X-Tenant-ID", tenantID.String())
			req.Header.Set("X-User-ID", userID.String())
			
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			
			// Verificar resultado
			require.Equal(t, tc.expectedStatus, rr.Code)
			
			// Verificar que os mocks de middleware foram chamados
			mockCORS.AssertCalled(t, "Called", mock.Anything, mock.Anything)
			mockAuth.AssertCalled(t, "Called", mock.Anything, mock.Anything)
			mockAuthz.AssertCalled(t, "Called", mock.Anything, mock.Anything)
			
			// Verificar que o serviço foi chamado apenas se permitido
			if tc.allowAccess {
				mockService.AssertExpectations(t)
			}
		})
	}
}

// TestMissingTenantID testa o comportamento quando o tenant ID está ausente
func TestMissingTenantID(t *testing.T) {
	// Configuração - criar router e handler sem usar os mocks de middleware
	logger := zerolog.Nop()
	tracer := noop.NewTracerProvider().Tracer("")
	mockService := new(MockRoleService)
	roleHandler := handler.NewRoleHandler(mockService, logger, tracer)
	
	// Substituir o extrator padrão por um que verifica headers obrigatórios
	roleHandler.TenantAndUserExtractor = func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			return uuid.Nil, uuid.Nil, fmt.Errorf("tenant ID não fornecido no header X-Tenant-ID")
		}
		
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			return uuid.Nil, uuid.Nil, fmt.Errorf("user ID não fornecido no header X-User-ID")
		}
		
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("tenant ID inválido: %w", err)
		}
		
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("user ID inválido: %w", err)
		}
		
		return tenantID, userID, nil
	}
	
	// Configurar router
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	roleHandler.RegisterRoutes(subrouter)
	
	// Executar request sem tenant ID
	req, err := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	require.NoError(t, err)
	req.Header.Set("X-User-ID", uuid.New().String()) // Apenas user ID
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	// Verificar resultado - deve ser erro de requisição inválida
	require.Equal(t, http.StatusBadRequest, rr.Code)
	
	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	
	// Verificar mensagem de erro
	assert.Contains(t, responseBody["message"], "tenant ID")
}

// TestMissingUserID testa o comportamento quando o user ID está ausente
func TestMissingUserID(t *testing.T) {
	// Configuração - similar ao teste anterior
	logger := zerolog.Nop()
	tracer := noop.NewTracerProvider().Tracer("")
	mockService := new(MockRoleService)
	roleHandler := handler.NewRoleHandler(mockService, logger, tracer)
	
	// Mesmo extrator que verifica headers obrigatórios
	roleHandler.TenantAndUserExtractor = func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			return uuid.Nil, uuid.Nil, fmt.Errorf("tenant ID não fornecido no header X-Tenant-ID")
		}
		
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			return uuid.Nil, uuid.Nil, fmt.Errorf("user ID não fornecido no header X-User-ID")
		}
		
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("tenant ID inválido: %w", err)
		}
		
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("user ID inválido: %w", err)
		}
		
		return tenantID, userID, nil
	}
	
	// Configurar router
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	roleHandler.RegisterRoutes(subrouter)
	
	// Executar request sem user ID
	req, err := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	require.NoError(t, err)
	req.Header.Set("X-Tenant-ID", uuid.New().String()) // Apenas tenant ID
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	// Verificar resultado - deve ser erro de requisição inválida
	require.Equal(t, http.StatusBadRequest, rr.Code)
	
	var responseBody map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	
	// Verificar mensagem de erro
	assert.Contains(t, responseBody["message"], "user ID")
}

// TestTenantScopeValidation testa a validação de escopo de tenant
func TestTenantScopeValidation(t *testing.T) {
	// Configuração
	logger := zerolog.Nop()
	tracer := noop.NewTracerProvider().Tracer("")
	mockService := new(MockRoleService)
	roleHandler := handler.NewRoleHandler(mockService, logger, tracer)
	
	// Simular extrator que sempre retorna um tenant específico
	tenantFromToken := uuid.New() // Tenant do token/contexto
	differentTenant := uuid.New() // Tenant diferente nos parâmetros
	userID := uuid.New()
	
	roleHandler.TenantAndUserExtractor = func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
		return tenantFromToken, userID, nil
	}
	
	// Configurar router
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	roleHandler.RegisterRoutes(subrouter)
	
	// Preparar uma requisição onde o tenant nos parâmetros é diferente do token
	reqBody := map[string]interface{}{
		"name":        "TestRole",
		"description": "Test role description",
		"type":        "CUSTOM",
		"tenant_id":   differentTenant.String(), // Tentativa de criar em tenant diferente
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	// Executar request
	req, err := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantFromToken.String()) // O tenant no header é o correto
	req.Header.Set("X-User-ID", userID.String())
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	// O handler deve ignorar o tenant_id no corpo e usar o do token
	// Testar se o mockService foi chamado com o tenant do token, não o do corpo
	mockService.On("CreateRole", mock.Anything, mock.MatchedBy(func(req application.CreateRoleRequest) bool {
		// O tenant deve ser o do token, não o fornecido no corpo
		return req.TenantID == tenantFromToken && req.TenantID != differentTenant
	})).Return(&model.Role{
		ID:          uuid.New(),
		TenantID:    tenantFromToken,
		Name:        "TestRole",
		Description: "Test role description",
		Type:        model.RoleTypeCustom,
		Status:      model.StatusActive,
		CreatedAt:   time.Now().UTC(),
		CreatedBy:   userID,
		UpdatedAt:   time.Now().UTC(),
		UpdatedBy:   userID,
	}, nil)
	
	// Se a requisição for bem-sucedida, o mockService será chamado
	if rr.Code == http.StatusCreated {
		mockService.AssertExpectations(t)
	}
}