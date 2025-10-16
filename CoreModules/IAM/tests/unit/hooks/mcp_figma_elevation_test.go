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

// TestMCPFigmaElevationHook testa a integração do hook MCP Figma com o sistema de elevação de privilégios
func TestMCPFigmaElevationHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_figma_elevation_hook_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Figma
	figmaHook := mcp.NewFigmaHook(mockElevationManager)

	// Configurar observabilidade
	figmaHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Configurar mapeamento de comandos Figma para escopos de elevação
	figmaHook.ConfigureScopeMappings(map[string][]string{
		"add_figma_file":  {"figma:file:add"},
		"post_comment":    {"figma:comment:create"},
		"reply_to_comment": {"figma:comment:reply"},
		"read_comments":   {"figma:comment:read"},
		"view_node":       {"figma:node:view"},
		"modify_design":   {"figma:design:modify"},
		"export_design":   {"figma:design:export"},
		"share_design":    {"figma:design:share"},
		"delete_design":   {"figma:design:delete"},
	})

	// Configurar arquivos Figma sensíveis (protegidos)
	figmaHook.ConfigureProtectedFiles([]string{
		"KpgCKLp83FhGZ3v2hFXoUG",  // Arquivo de design principal do IAM
		"L4qDdTT92JkMp1w3zWqRxH",  // Arquivo de design principal do Payment Gateway
		"NrSxYY67QzBvP8w4mDcZlJ",  // Arquivo de design principal do Mobile Money
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de verificação bem-sucedida para alguns dos testes
	validVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-figma-001",
		UserID:         "user:designer:123",
		ElevatedRoles:  []string{"figma_admin"},
		ElevatedScopes: []string{"figma:design:modify", "figma:comment:create", "figma:design:share"},
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
		ElevationID:    "elev-figma-999",
		UserID:         "user:designer:123",
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
			name:           "Comentar em arquivo protegido com elevação válida",
			command:        "post_comment",
			elevationToken: "valid-figma-token-123",
			requestPayload: map[string]interface{}{
				"file_key": "KpgCKLp83FhGZ3v2hFXoUG", // Arquivo protegido
				"message":  "Precisamos ajustar este componente para conformidade com GDPR",
				"node_id":  "35:42",
				"x":        120.5,
				"y":        250.3,
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-figma-token-123",
					[]string{"figma:comment:create"},
				).Return(validVerificationResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Modificar design em arquivo protegido sem elevação",
			command:        "modify_design",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"file_key":  "L4qDdTT92JkMp1w3zWqRxH", // Arquivo protegido
				"node_id":   "42:57",
				"operations": []map[string]interface{}{
					{
						"type": "UPDATE_STYLE",
						"properties": map[string]interface{}{
							"fill": "#3366FF",
						},
					},
				},
			},
			mockSetup: func() {
				// Sem token, não deve chamar VerifyElevation
			},
			expectedAllowed: false,
			expectedReason:  "Elevação de privilégios requerida para figma:design:modify",
		},
		{
			name:           "Compartilhar design protegido com elevação inválida",
			command:        "share_design",
			elevationToken: "invalid-figma-token-999",
			requestPayload: map[string]interface{}{
				"file_key": "NrSxYY67QzBvP8w4mDcZlJ", // Arquivo protegido
				"access_level": "view",
				"emails": []string{
					"parceiro@exemplo.com",
					"externo@exemplo.com",
				},
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação inválida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"invalid-figma-token-999",
					[]string{"figma:design:share"},
				).Return(invalidVerificationResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Token expirado",
		},
		{
			name:           "Visualizar nó em arquivo não protegido",
			command:        "view_node",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"file_key": "RqsTZZ41JdLnV2r5mFbAcK", // Arquivo não protegido
				"node_id":  "15:23",
			},
			mockSetup: func() {
				// Não é arquivo protegido, não deve chamar VerifyElevation
			},
			expectedAllowed: true,
			expectedReason:  "Comando não requer elevação de privilégios",
		},
		{
			name:           "Compartilhar design protegido com elevação válida",
			command:        "share_design",
			elevationToken: "valid-figma-token-456",
			requestPayload: map[string]interface{}{
				"file_key": "KpgCKLp83FhGZ3v2hFXoUG", // Arquivo protegido
				"access_level": "edit",
				"emails": []string{
					"designer@innovabiz.com",
					"developer@innovabiz.com",
				},
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-figma-token-456",
					[]string{"figma:design:share"},
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
			allowed, reason, err := figmaHook.AuthorizeFigmaCommand(hookCtx, tc.command, requestJSON)
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

// TestMCPFigmaElevationRequest testa o fluxo de solicitação de elevação através do hook MCP Figma
func TestMCPFigmaElevationRequest(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_figma_elevation_request_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Figma
	figmaHook := mcp.NewFigmaHook(mockElevationManager)

	// Configurar observabilidade
	figmaHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de solicitação de elevação para os testes
	elevationResult := &elevation.ElevationResult{
		ElevationID:    "elev-figma-request-001",
		UserID:         "user:designer:789",
		ElevatedRoles:  []string{"figma_admin"},
		ElevatedScopes: []string{"figma:design:modify", "figma:design:share"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		ElevationToken: "new-figma-token-123",
		ApprovedBy:     "user:manager:456",
		ApprovalTime:   baseTime,
		AuditMetadata: map[string]interface{}{
			"tenant_id":   "tenant_angola_1",
			"market":      "angola",
			"request_ip":  "192.168.1.100",
			"user_agent":  "MCP Figma Client/1.0",
		},
	}

	// Configurar mock para a solicitação de elevação
	mockElevationManager.On(
		"RequestElevation",
		mock.Anything,
		mock.MatchedBy(func(req *elevation.ElevationRequest) bool {
			return req.UserID == "user:designer:789" &&
				len(req.RequestedScopes) > 0 &&
				req.Duration.Minutes() <= 60 // Máximo 1 hora
		}),
	).Return(elevationResult, nil).Once()

	// Preparar requisição de elevação
	elevationRequest := &mcp.ElevationRequestPayload{
		UserID:          "user:designer:789",
		RequestedScopes: []string{"figma:design:modify", "figma:design:share"},
		Justification:   "Atualização urgente de design para lançamento do produto",
		Duration:        "30m",
		EmergencyAccess: true,
		Context: map[string]interface{}{
			"tenant_id":    "tenant_angola_1",
			"market":       "angola",
			"business_unit": "design",
			"client_ip":    "192.168.1.100",
			"user_agent":   "MCP Figma Client/1.0",
			"file_key":     "KpgCKLp83FhGZ3v2hFXoUG",
		},
	}

	// Serializar requisição
	requestJSON, err := json.Marshal(elevationRequest)
	require.NoError(t, err, "Erro ao serializar requisição de elevação")

	// Executar o teste
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "FigmaElevationRequest")

	// Chamar a função de solicitação de elevação
	resultJSON, err := figmaHook.RequestFigmaElevation(testCtx, requestJSON)
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
	obs.RecordTestEnd(testCtx, "FigmaElevationRequest", err == nil, time.Since(time.Now()))
}

// TestMCPFigmaMultiTenantIsolation testa o isolamento multi-tenant no hook MCP Figma
func TestMCPFigmaMultiTenantIsolation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_figma_multi_tenant_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Figma com isolamento multi-tenant
	figmaHook := mcp.NewFigmaHook(mockElevationManager)
	figmaHook.ConfigureObservability(obs.Logger(), obs.Tracer())
	figmaHook.EnableMultiTenantIsolation(true)

	// Configurar mapeamento de tenants para arquivos Figma permitidos
	figmaHook.ConfigureTenantFiles(map[string][]string{
		"tenant_angola_1": {
			"KpgCKLp83FhGZ3v2hFXoUG", // Design principal IAM - Angola
			"RqsTZZ41JdLnV2r5mFbAcK", // Design secundário - Angola
			"XqW3BB67KpLsY9m4zVdZaJ", // Design compartilhado global
		},
		"tenant_brazil_1": {
			"L4qDdTT92JkMp1w3zWqRxH", // Design principal Payment Gateway - Brasil
			"NrSxYY67QzBvP8w4mDcZlJ", // Design principal Mobile Money - Brasil
			"XqW3BB67KpLsY9m4zVdZaJ", // Design compartilhado global
		},
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultados de verificação para diferentes tenants
	angolaTenantResult := &elevation.VerificationResult{
		ElevationID:    "elev-figma-angola-001",
		UserID:         "user:designer:123",
		ElevatedRoles:  []string{"figma_admin"},
		ElevatedScopes: []string{"figma:design:modify"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		IsValid:        true,
		AuditMetadata: map[string]interface{}{
			"tenant_id": "tenant_angola_1",
			"market":    "angola",
		},
	}

	brazilTenantResult := &elevation.VerificationResult{
		ElevationID:    "elev-figma-brazil-001",
		UserID:         "user:designer:456",
		ElevatedRoles:  []string{"figma_admin"},
		ElevatedScopes: []string{"figma:design:modify"},
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
		fileKey        string
		mockSetup      func()
		expectedAllowed bool
		expectedReason  string
	}{
		{
			name:           "Acesso ao arquivo do tenant correto - Angola",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			fileKey:        "KpgCKLp83FhGZ3v2hFXoUG",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"figma:design:modify"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso cruzado entre tenants - Angola tentando acessar Brasil",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			fileKey:        "L4qDdTT92JkMp1w3zWqRxH",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"figma:design:modify"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Acesso negado: isolamento multi-tenant",
		},
		{
			name:           "Acesso ao arquivo global a partir de Angola",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			fileKey:        "XqW3BB67KpLsY9m4zVdZaJ",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"figma:design:modify"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso ao arquivo do tenant correto - Brasil",
			elevationToken: "brazil-tenant-token-456",
			tenantID:       "tenant_brazil_1",
			fileKey:        "L4qDdTT92JkMp1w3zWqRxH",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"brazil-tenant-token-456",
					[]string{"figma:design:modify"},
				).Return(brazilTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso cruzado entre tenants - Brasil tentando acessar Angola",
			elevationToken: "brazil-tenant-token-456",
			tenantID:       "tenant_brazil_1",
			fileKey:        "KpgCKLp83FhGZ3v2hFXoUG",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"brazil-tenant-token-456",
					[]string{"figma:design:modify"},
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
				"file_key":  tc.fileKey,
				"node_id":   "42:57",
				"operations": []map[string]interface{}{
					{
						"type": "UPDATE_STYLE",
						"properties": map[string]interface{}{
							"fill": "#3366FF",
						},
					},
				},
			}

			requestJSON, err := json.Marshal(requestPayload)
			require.NoError(t, err, "Erro ao serializar payload de requisição")

			// Configurar contexto com token de elevação e tenant
			hookCtx := mcp.WithElevationToken(testCtx, tc.elevationToken)
			hookCtx = mcp.WithTenantID(hookCtx, tc.tenantID)

			// Executar o hook
			allowed, reason, err := figmaHook.AuthorizeFigmaCommand(hookCtx, "modify_design", requestJSON)
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