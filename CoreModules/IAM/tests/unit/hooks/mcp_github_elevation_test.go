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
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// TestMCPGitHubElevationHook testa a integração do hook MCP GitHub com o sistema de elevação de privilégios
func TestMCPGitHubElevationHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_github_elevation_hook_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP GitHub
	githubHook := mcp.NewGitHubHook(mockElevationManager)

	// Configurar observabilidade
	githubHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Configurar mapeamento de comandos GitHub para escopos de elevação
	githubHook.ConfigureScopeMappings(map[string][]string{
		"push_files":             {"github:repo:push"},
		"merge_pull_request":     {"github:repo:merge"},
		"create_or_update_file":  {"github:repo:file:write"},
		"delete_issue":           {"github:issues:delete"},
		"create_repository":      {"github:org:repo:create"},
		"delete_repository":      {"github:org:repo:delete"},
		"create_branch":          {"github:repo:branch:create"},
		"delete_branch":          {"github:repo:branch:delete"},
		"create_pull_request":    {"github:repo:pr:create"},
		"update_pull_request":    {"github:repo:pr:update"},
		"fork_repository":        {"github:repo:fork"},
	})

	// Configurar proteções específicas para repositórios
	githubHook.ConfigureProtectedRepositories([]string{
		"innovabizdevops/innovabiz-iam",
		"innovabizdevops/innovabiz-api-gateway",
		"innovabizdevops/innovabiz-core",
	})

	// Configurar proteções para branches específicas
	githubHook.ConfigureProtectedBranches(map[string][]string{
		"innovabizdevops/innovabiz-iam": {"main", "production", "release"},
		"innovabizdevops/innovabiz-api-gateway": {"main", "production"},
		"innovabizdevops/innovabiz-core": {"main", "production", "staging"},
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de verificação bem-sucedida para alguns dos testes
	validVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-github-001",
		UserID:         "user:developer:123",
		ElevatedRoles:  []string{"github_admin"},
		ElevatedScopes: []string{"github:repo:push", "github:repo:merge", "github:repo:branch:create"},
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
		ElevationID:    "elev-github-999",
		UserID:         "user:developer:123",
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
			name:           "Push para branch protegida com elevação válida",
			command:        "push_files",
			elevationToken: "valid-github-token-123",
			requestPayload: map[string]interface{}{
				"owner":  "innovabizdevops",
				"repo":   "innovabiz-iam",
				"branch": "main",
				"files": []map[string]interface{}{
					{
						"path":    "src/auth/auth.go",
						"content": "package auth\n\nfunc Authenticate() bool {\n\treturn true\n}\n",
					},
				},
				"message": "Update authentication logic",
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-github-token-123",
					[]string{"github:repo:push"},
				).Return(validVerificationResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Merge pull request para branch protegida sem elevação",
			command:        "merge_pull_request",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"owner":      "innovabizdevops",
				"repo":       "innovabiz-iam",
				"pull_number": 42,
				"merge_method": "merge",
			},
			mockSetup: func() {
				// Sem token, não deve chamar VerifyElevation
			},
			expectedAllowed: false,
			expectedReason:  "Elevação de privilégios requerida para github:repo:merge",
		},
		{
			name:           "Criar arquivo em repositório protegido com elevação inválida",
			command:        "create_or_update_file",
			elevationToken: "invalid-github-token-999",
			requestPayload: map[string]interface{}{
				"owner":   "innovabizdevops",
				"repo":    "innovabiz-core",
				"path":    "config/production.yaml",
				"content": "key: value",
				"message": "Update production config",
				"branch":  "production",
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação inválida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"invalid-github-token-999",
					[]string{"github:repo:file:write"},
				).Return(invalidVerificationResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Token expirado",
		},
		{
			name:           "Criar novo repositório não protegido",
			command:        "create_repository",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"name":        "new-project",
				"description": "A new project",
				"private":     true,
				"autoInit":    true,
			},
			mockSetup: func() {
				// Não é repositório protegido, não deve chamar VerifyElevation
			},
			expectedAllowed: true,
			expectedReason:  "Comando não requer elevação de privilégios",
		},
		{
			name:           "Criar branch em repositório protegido com elevação válida",
			command:        "create_branch",
			elevationToken: "valid-github-token-456",
			requestPayload: map[string]interface{}{
				"owner":       "innovabizdevops",
				"repo":        "innovabiz-api-gateway",
				"branch":      "feature/new-auth",
				"from_branch": "main",
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-github-token-456",
					[]string{"github:repo:branch:create"},
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
			allowed, reason, err := githubHook.AuthorizeGitHubCommand(hookCtx, tc.command, requestJSON)
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

// TestMCPGitHubElevationRequest testa o fluxo de solicitação de elevação através do hook MCP GitHub
func TestMCPGitHubElevationRequest(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_github_elevation_request_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP GitHub
	githubHook := mcp.NewGitHubHook(mockElevationManager)

	// Configurar observabilidade
	githubHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de solicitação de elevação para os testes
	elevationResult := &elevation.ElevationResult{
		ElevationID:    "elev-github-request-001",
		UserID:         "user:developer:789",
		ElevatedRoles:  []string{"github_admin"},
		ElevatedScopes: []string{"github:repo:push", "github:repo:merge"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		ElevationToken: "new-github-token-123",
		ApprovedBy:     "user:manager:456",
		ApprovalTime:   baseTime,
		AuditMetadata: map[string]interface{}{
			"tenant_id":   "tenant_angola_1",
			"market":      "angola",
			"request_ip":  "192.168.1.100",
			"user_agent":  "MCP GitHub Client/1.0",
		},
	}

	// Configurar mock para a solicitação de elevação
	mockElevationManager.On(
		"RequestElevation",
		mock.Anything,
		mock.MatchedBy(func(req *elevation.ElevationRequest) bool {
			return req.UserID == "user:developer:789" &&
				len(req.RequestedScopes) > 0 &&
				req.Duration.Minutes() <= 60 // Máximo 1 hora
		}),
	).Return(elevationResult, nil).Once()

	// Preparar requisição de elevação
	elevationRequest := &mcp.ElevationRequestPayload{
		UserID:          "user:developer:789",
		RequestedScopes: []string{"github:repo:push", "github:repo:merge"},
		Justification:   "Release emergencial para correção de segurança",
		Duration:        "30m",
		EmergencyAccess: true,
		Context: map[string]interface{}{
			"tenant_id":    "tenant_angola_1",
			"market":       "angola",
			"business_unit": "development",
			"client_ip":    "192.168.1.100",
			"user_agent":   "MCP GitHub Client/1.0",
			"repository":   "innovabizdevops/innovabiz-iam",
			"branch":       "main",
			"commit":       "abc123def456",
		},
	}

	// Serializar requisição
	requestJSON, err := json.Marshal(elevationRequest)
	require.NoError(t, err, "Erro ao serializar requisição de elevação")

	// Executar o teste
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "GitHubElevationRequest")

	// Chamar a função de solicitação de elevação
	resultJSON, err := githubHook.RequestGitHubElevation(testCtx, requestJSON)
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
	obs.RecordTestEnd(testCtx, "GitHubElevationRequest", err == nil, time.Since(time.Now()))
}

// TestMCPGitHubMultiTenantIsolation testa o isolamento multi-tenant no hook MCP GitHub
func TestMCPGitHubMultiTenantIsolation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_github_multi_tenant_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP GitHub com isolamento multi-tenant
	githubHook := mcp.NewGitHubHook(mockElevationManager)
	githubHook.ConfigureObservability(obs.Logger(), obs.Tracer())
	githubHook.EnableMultiTenantIsolation(true)

	// Configurar mapeamento de tenants para organizações/repositórios permitidos
	githubHook.ConfigureTenantRepositories(map[string][]string{
		"tenant_angola_1": {
			"innovabizdevops/innovabiz-angola",
			"innovabizdevops/innovabiz-global",
		},
		"tenant_brazil_1": {
			"innovabizdevops/innovabiz-brazil",
			"innovabizdevops/innovabiz-global",
		},
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultados de verificação para diferentes tenants
	angolaTenantResult := &elevation.VerificationResult{
		ElevationID:    "elev-github-angola-001",
		UserID:         "user:developer:123",
		ElevatedRoles:  []string{"github_admin"},
		ElevatedScopes: []string{"github:repo:push"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		IsValid:        true,
		AuditMetadata: map[string]interface{}{
			"tenant_id": "tenant_angola_1",
			"market":    "angola",
		},
	}

	brazilTenantResult := &elevation.VerificationResult{
		ElevationID:    "elev-github-brazil-001",
		UserID:         "user:developer:456",
		ElevatedRoles:  []string{"github_admin"},
		ElevatedScopes: []string{"github:repo:push"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		IsValid:        true,
		AuditMetadata: map[string]interface{}{
			"tenant_id": "tenant_brazil_1",
			"market":    "brazil",
		},
	}

	// Casos de teste
	testCases := []struct {
		name           string
		elevationToken string
		tenantID       string
		repoOwner      string
		repoName       string
		mockSetup      func()
		expectedAllowed bool
		expectedReason  string
	}{
		{
			name:           "Acesso ao repositório do tenant correto - Angola",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			repoOwner:      "innovabizdevops",
			repoName:       "innovabiz-angola",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"github:repo:push"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso cruzado entre tenants - Angola tentando acessar Brasil",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			repoOwner:      "innovabizdevops",
			repoName:       "innovabiz-brazil",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"github:repo:push"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Acesso negado: isolamento multi-tenant",
		},
		{
			name:           "Acesso ao repositório global a partir de Angola",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			repoOwner:      "innovabizdevops",
			repoName:       "innovabiz-global",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"github:repo:push"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso ao repositório do tenant correto - Brasil",
			elevationToken: "brazil-tenant-token-456",
			tenantID:       "tenant_brazil_1",
			repoOwner:      "innovabizdevops",
			repoName:       "innovabiz-brazil",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"brazil-tenant-token-456",
					[]string{"github:repo:push"},
				).Return(brazilTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso cruzado entre tenants - Brasil tentando acessar Angola",
			elevationToken: "brazil-tenant-token-456",
			tenantID:       "tenant_brazil_1",
			repoOwner:      "innovabizdevops",
			repoName:       "innovabiz-angola",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"brazil-tenant-token-456",
					[]string{"github:repo:push"},
				).Return(brazilTenantResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Acesso negado: isolamento multi-tenant",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)

			// Configurar mocks específicos para este caso
			tc.mockSetup()

			// Criar payload de requisição para o hook
			requestPayload := map[string]interface{}{
				"owner":  tc.repoOwner,
				"repo":   tc.repoName,
				"branch": "main",
				"files": []map[string]interface{}{
					{
						"path":    "README.md",
						"content": "# Test Repository",
					},
				},
				"message": "Update README",
			}

			requestJSON, err := json.Marshal(requestPayload)
			require.NoError(t, err, "Erro ao serializar payload de requisição")

			// Configurar contexto com token de elevação e tenant
			hookCtx := mcp.WithElevationToken(testCtx, tc.elevationToken)
			hookCtx = mcp.WithTenantID(hookCtx, tc.tenantID)

			// Executar o hook
			allowed, reason, err := githubHook.AuthorizeGitHubCommand(hookCtx, "push_files", requestJSON)
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