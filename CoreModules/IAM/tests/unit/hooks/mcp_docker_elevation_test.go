// Package hooks_test contém testes unitários para hooks de integração MCP com elevação de privilégios
package hooks_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/innovabizdevops/innovabiz-iam/authorization/elevation"
	"github.com/innovabizdevops/innovabiz-iam/hooks/mcp"
	"github.com/innovabizdevops/innovabiz-iam/observability"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// MockPrivilegeElevationManager é um mock do gerenciador de elevação de privilégios
type MockPrivilegeElevationManager struct {
	mock.Mock
}

func (m *MockPrivilegeElevationManager) RequestElevation(ctx context.Context, request *elevation.ElevationRequest) (*elevation.ElevationResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*elevation.ElevationResult), args.Error(1)
}

func (m *MockPrivilegeElevationManager) VerifyElevation(ctx context.Context, token string, requiredScopes []string) (*elevation.VerificationResult, error) {
	args := m.Called(ctx, token, requiredScopes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*elevation.VerificationResult), args.Error(1)
}

func (m *MockPrivilegeElevationManager) RevokeElevation(ctx context.Context, elevationID string, revokerID string, reason string) error {
	args := m.Called(ctx, elevationID, revokerID, reason)
	return args.Error(0)
}

// TestMCPDockerElevationHook testa a integração do hook MCP Docker com o sistema de elevação de privilégios
func TestMCPDockerElevationHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_docker_elevation_hook_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Docker
	dockerHook := mcp.NewDockerHook(mockElevationManager)

	// Configurar observabilidade
	dockerHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Configurar mapeamento de comandos Docker para escopos de elevação
	dockerHook.ConfigureScopeMappings(map[string][]string{
		"docker exec":  {"docker:containers:exec"},
		"docker run":   {"docker:containers:run"},
		"docker rm":    {"docker:containers:remove"},
		"docker stop":  {"docker:containers:stop"},
		"docker build": {"docker:images:build"},
		"docker push":  {"docker:images:push"},
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de verificação bem-sucedida para alguns dos testes
	validVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-docker-001",
		UserID:         "user:devops:123",
		ElevatedRoles:  []string{"docker_admin"},
		ElevatedScopes: []string{"docker:containers:exec", "docker:containers:run"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		IsValid:        true,
		AuditMetadata: map[string]interface{}{
			"approved_by": "user:manager:456",
			"tenant_id":   "tenant_angola_1",
			"market":      "angola",
		},
	}

	// Configurar resultado de verificação inválida para alguns dos testes
	invalidVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-docker-999",
		UserID:         "user:devops:123",
		ElevatedRoles:  []string{},
		ElevatedScopes: []string{},
		ExpirationTime: baseTime.Add(-10 * time.Minute), // Expirado
		IsValid:        false,
		AuditMetadata: map[string]interface{}{
			"error_reason": "Token expirado",
		},
	}

	// Casos de teste
	testCases := []struct {
		name            string
		command         string
		elevationToken  string
		requestPayload  map[string]interface{}
		mockSetup       func()
		expectedAllowed bool
		expectedReason  string
	}{
		{
			name:           "Comando Docker exec com elevação válida",
			command:        "docker exec",
			elevationToken: "valid-docker-token-123",
			requestPayload: map[string]interface{}{
				"args": []string{"docker", "exec", "-it", "container-123", "/bin/bash"},
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-docker-token-123",
					[]string{"docker:containers:exec"},
				).Return(validVerificationResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Comando Docker rm sem elevação",
			command:        "docker rm",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"args": []string{"docker", "rm", "container-456"},
			},
			mockSetup: func() {
				// Sem token, não deve chamar VerifyElevation
			},
			expectedAllowed: false,
			expectedReason:  "Elevação de privilégios requerida para docker:containers:remove",
		},
		{
			name:           "Comando Docker run com elevação inválida",
			command:        "docker run",
			elevationToken: "invalid-docker-token-999",
			requestPayload: map[string]interface{}{
				"args": []string{"docker", "run", "-d", "nginx:latest"},
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação inválida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"invalid-docker-token-999",
					[]string{"docker:containers:run"},
				).Return(invalidVerificationResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Token expirado",
		},
		{
			name:           "Comando Docker sem necessidade de elevação",
			command:        "docker images",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"args": []string{"docker", "images", "--all"},
			},
			mockSetup: func() {
				// Não requer elevação, não deve chamar VerifyElevation
			},
			expectedAllowed: true,
			expectedReason:  "Comando não requer elevação de privilégios",
		},
		{
			name:           "Comando Docker stop com elevação válida",
			command:        "docker stop",
			elevationToken: "valid-docker-token-456",
			requestPayload: map[string]interface{}{
				"args": []string{"docker", "stop", "container-789"},
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-docker-token-456",
					[]string{"docker:containers:stop"},
				).Return(validVerificationResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)

			// Configurar mocks específicos para este caso
			tc.mockSetup()

			// Criar payload de requisição para o hook
			requestJSON, err := json.Marshal(tc.requestPayload)
			require.NoError(t, err, "Erro ao serializar payload de requisição")

			// Configurar contexto com token de elevação, se houver
			var hookCtx context.Context
			if tc.elevationToken != "" {
				hookCtx = mcp.WithElevationToken(testCtx, tc.elevationToken)
			} else {
				hookCtx = testCtx
			}

			// Executar o hook
			allowed, reason, err := dockerHook.AuthorizeDockerCommand(hookCtx, tc.command, requestJSON)
			require.NoError(t, err, "O hook não deveria falhar")

			// Verificar resultados
			assert.Equal(t, tc.expectedAllowed, allowed, "Resultado de autorização incorreto")
			assert.Contains(t, reason, tc.expectedReason, "Razão de autorização incorreta")

			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}

	// Verificar que todos os mocks foram chamados como esperado
	mockElevationManager.AssertExpectations(t)
}

// TestMCPDockerElevationRequest testa o fluxo de solicitação de elevação através do hook MCP Docker
func TestMCPDockerElevationRequest(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_docker_elevation_request_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Docker
	dockerHook := mcp.NewDockerHook(mockElevationManager)

	// Configurar observabilidade
	dockerHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de solicitação de elevação para os testes
	elevationResult := &elevation.ElevationResult{
		ElevationID:    "elev-docker-request-001",
		UserID:         "user:devops:789",
		ElevatedRoles:  []string{"docker_admin"},
		ElevatedScopes: []string{"docker:containers:exec", "docker:containers:run"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		ElevationToken: "new-docker-token-123",
		ApprovedBy:     "user:manager:456",
		ApprovalTime:   baseTime,
		AuditMetadata: map[string]interface{}{
			"tenant_id":   "tenant_angola_1",
			"market":      "angola",
			"request_ip":  "192.168.1.100",
			"user_agent":  "MCP Docker Client/1.0",
		},
	}

	// Configurar mock para a solicitação de elevação
	mockElevationManager.On(
		"RequestElevation",
		mock.Anything,
		mock.MatchedBy(func(req *elevation.ElevationRequest) bool {
			return req.UserID == "user:devops:789" &&
				len(req.RequestedScopes) > 0 &&
				req.Duration.Minutes() <= 60 // Máximo 1 hora
		}),
	).Return(elevationResult, nil).Once()

	// Preparar requisição de elevação
	elevationRequest := &mcp.ElevationRequestPayload{
		UserID:          "user:devops:789",
		RequestedScopes: []string{"docker:containers:exec", "docker:containers:run"},
		Justification:   "Manutenção emergencial de contêineres Docker",
		Duration:        "30m",
		EmergencyAccess: true,
		Context: map[string]interface{}{
			"tenant_id":    "tenant_angola_1",
			"market":       "angola",
			"business_unit": "operations",
			"client_ip":    "192.168.1.100",
			"user_agent":   "MCP Docker Client/1.0",
		},
	}

	// Serializar requisição
	requestJSON, err := json.Marshal(elevationRequest)
	require.NoError(t, err, "Erro ao serializar requisição de elevação")

	// Executar o teste
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "DockerElevationRequest")

	// Chamar a função de solicitação de elevação
	resultJSON, err := dockerHook.RequestDockerElevation(testCtx, requestJSON)
	require.NoError(t, err, "Solicitação de elevação não deveria falhar")

	// Deserializar resposta
	var elevationResponse mcp.ElevationResponse
	err = json.Unmarshal(resultJSON, &elevationResponse)
	require.NoError(t, err, "Erro ao deserializar resposta de elevação")

	// Verificar resultados
	assert.Equal(t, elevationResult.ElevationID, elevationResponse.ElevationID, "ID de elevação incorreto")
	assert.Equal(t, elevationResult.ElevationToken, elevationResponse.ElevationToken, "Token de elevação incorreto")
	assert.Equal(t, elevationResult.ExpirationTime.UTC().Format(time.RFC3339), elevationResponse.ExpirationTime, "Tempo de expiração incorreto")
	assert.Equal(t, elevationResult.ElevatedScopes, elevationResponse.ElevatedScopes, "Escopos elevados incorretos")
	assert.Equal(t, elevationResult.ApprovedBy, elevationResponse.ApprovedBy, "Aprovador incorreto")

	// Verificar que todos os mocks foram chamados como esperado
	mockElevationManager.AssertExpectations(t)

	// Registrar conclusão do teste
	obs.RecordTestEnd(testCtx, "DockerElevationRequest", err == nil, time.Since(time.Now()))
}

// TestMCPDockerElevationUsage testa o registro de uso da elevação através do hook MCP Docker
func TestMCPDockerElevationUsage(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_docker_elevation_usage_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Docker
	dockerHook := mcp.NewDockerHook(mockElevationManager)

	// Configurar observabilidade
	dockerHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de verificação para o teste
	validVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-docker-usage-001",
		UserID:         "user:devops:123",
		ElevatedRoles:  []string{"docker_admin"},
		ElevatedScopes: []string{"docker:containers:exec"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		IsValid:        true,
		AuditMetadata: map[string]interface{}{
			"approved_by": "user:manager:456",
			"tenant_id":   "tenant_angola_1",
			"market":      "angola",
		},
	}

	// Configurar mock para verificação de elevação
	mockElevationManager.On(
		"VerifyElevation",
		mock.Anything,
		"docker-usage-token-123",
		[]string{"docker:containers:exec"},
	).Return(validVerificationResult, nil).Once()

	// Preparar requisição para comando Docker com elevação
	requestPayload := map[string]interface{}{
		"args": []string{"docker", "exec", "-it", "container-123", "/bin/bash"},
		"command_id": "cmd-docker-123",
		"resource": "container-123",
		"operation": "exec",
	}

	requestJSON, err := json.Marshal(requestPayload)
	require.NoError(t, err, "Erro ao serializar payload de requisição")

	// Executar o teste
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "DockerElevationUsage")

	// Configurar contexto com token de elevação
	hookCtx := mcp.WithElevationToken(testCtx, "docker-usage-token-123")

	// Chamar o hook para autorizar o comando com elevação
	allowed, reason, err := dockerHook.AuthorizeDockerCommand(hookCtx, "docker exec", requestJSON)
	require.NoError(t, err, "A autorização não deveria falhar")

	// Verificar resultados
	assert.True(t, allowed, "O comando deveria ser autorizado")
	assert.Contains(t, reason, "Elevação verificada com sucesso", "Razão de autorização incorreta")

	// Verificar que o token e escopo foram verificados
	mockElevationManager.AssertExpectations(t)

	// Registrar conclusão do teste
	obs.RecordTestEnd(testCtx, "DockerElevationUsage", err == nil, time.Since(time.Now()))
}

// MockElevationTracer é um mock para o tracer de observabilidade de elevação
type MockElevationTracer struct {
	mock.Mock
}

func (m *MockElevationTracer) StartSpan(ctx context.Context, operationName string) (context.Context, observability.Span) {
	args := m.Called(ctx, operationName)
	if args.Get(0) == nil {
		return ctx, nil
	}
	return args.Get(0).(context.Context), args.Get(1).(observability.Span)
}

// TestMCPDockerElevationObservability testa a observabilidade do hook MCP Docker
func TestMCPDockerElevationObservability(t *testing.T) {
	// Configurar mock para o tracer
	mockTracer := new(MockElevationTracer)
	mockSpan := new(MockSpan)

	// Configurar mocks para o gerenciador de elevação
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Configurar retorno do mock do tracer
	mockTracer.On("StartSpan", mock.Anything, "AuthorizeDockerCommand").
		Return(context.Background(), mockSpan)

	// Configurar comportamento do span
	mockSpan.On("SetAttribute", mock.Anything, mock.Anything).Return()
	mockSpan.On("End").Return()

	// Criar o hook MCP Docker
	dockerHook := mcp.NewDockerHook(mockElevationManager)

	// Configurar tracer mock
	dockerHook.ConfigureObservability(nil, mockTracer)

	// Preparar requisição para comando Docker
	requestPayload := map[string]interface{}{
		"args": []string{"docker", "images", "--all"},
	}

	requestJSON, err := json.Marshal(requestPayload)
	require.NoError(t, err, "Erro ao serializar payload de requisição")

	// Executar o hook
	ctx := context.Background()
	_, _, err = dockerHook.AuthorizeDockerCommand(ctx, "docker images", requestJSON)
	require.NoError(t, err, "O hook não deveria falhar")

	// Verificar que o tracer foi usado corretamente
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}