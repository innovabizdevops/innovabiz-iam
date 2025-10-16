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

// TestMCPDesktopCommanderElevationHook testa a integração do hook MCP Desktop Commander com o sistema de elevação de privilégios
func TestMCPDesktopCommanderElevationHook(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_desktop_commander_elevation_hook_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Desktop Commander
	dcHook := mcp.NewDesktopCommanderHook(mockElevationManager)

	// Configurar observabilidade
	dcHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Configurar mapeamento de comandos Desktop Commander para escopos de elevação
	dcHook.ConfigureScopeMappings(map[string][]string{
		"start_process":       {"desktop:process:start"},
		"kill_process":        {"desktop:process:kill"},
		"edit_block":          {"desktop:file:edit"},
		"write_file":          {"desktop:file:write"},
		"set_config_value":    {"desktop:config:modify"},
		"execute_command":     {"desktop:command:execute"},
		"read_file":           {"desktop:file:read"},
		"read_multiple_files": {"desktop:file:read"},
		"search_code":         {"desktop:code:search"},
		"search_files":        {"desktop:file:search"},
	})

	// Configurar áreas sensíveis que requerem elevação específica
	dcHook.ConfigureSensitiveAreas(map[string][]string{
		"desktop:file:write": {
			"/etc/",
			"/var/",
			"/usr/bin/",
			"/usr/local/bin/",
			"C:\\Windows\\System32\\",
			"C:\\Program Files\\",
		},
		"desktop:process:kill": {
			"systemd",
			"sshd",
			"explorer.exe",
			"winlogon.exe",
		},
		"desktop:config:modify": {
			"allowedDirectories",
			"blockedCommands",
		},
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de verificação bem-sucedida para alguns dos testes
	validVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-dc-001",
		UserID:         "user:admin:123",
		ElevatedRoles:  []string{"desktop_admin"},
		ElevatedScopes: []string{"desktop:file:write", "desktop:config:modify", "desktop:process:kill"},
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
		ElevationID:    "elev-dc-999",
		UserID:         "user:admin:123",
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
			name:           "Modificar arquivo do sistema com elevação válida",
			command:        "write_file",
			elevationToken: "valid-dc-token-123",
			requestPayload: map[string]interface{}{
				"path":    "/etc/hosts",
				"content": "127.0.0.1 localhost\n",
				"mode":    "rewrite",
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-dc-token-123",
					[]string{"desktop:file:write"},
				).Return(validVerificationResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Modificar configuração sensível sem elevação",
			command:        "set_config_value",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"key":   "allowedDirectories",
				"value": "[]", // Array vazio é sensível
			},
			mockSetup: func() {
				// Sem token, não deve chamar VerifyElevation
			},
			expectedAllowed: false,
			expectedReason:  "Elevação de privilégios requerida para desktop:config:modify",
		},
		{
			name:           "Encerrar processo do sistema com elevação inválida",
			command:        "kill_process",
			elevationToken: "invalid-dc-token-999",
			requestPayload: map[string]interface{}{
				"pid": 1, // PID 1 é sensível
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação inválida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"invalid-dc-token-999",
					[]string{"desktop:process:kill"},
				).Return(invalidVerificationResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Token expirado",
		},
		{
			name:           "Ler arquivo sem necessidade de elevação",
			command:        "read_file",
			elevationToken: "",
			requestPayload: map[string]interface{}{
				"path":   "/home/user/document.txt",
				"offset": 0,
				"length": 100,
			},
			mockSetup: func() {
				// Não requer elevação para arquivo não sensível, não deve chamar VerifyElevation
			},
			expectedAllowed: true,
			expectedReason:  "Comando não requer elevação de privilégios",
		},
		{
			name:           "Modificar configurações do sistema com elevação válida",
			command:        "set_config_value",
			elevationToken: "valid-dc-token-456",
			requestPayload: map[string]interface{}{
				"key":   "blockedCommands",
				"value": `["rm -rf /", "format c:"]`,
			},
			mockSetup: func() {
				// Configurar mock para verificar elevação válida
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"valid-dc-token-456",
					[]string{"desktop:config:modify"},
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
			allowed, reason, err := dcHook.AuthorizeDesktopCommand(hookCtx, tc.command, requestJSON)
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

// TestMCPDesktopCommanderElevationRequest testa o fluxo de solicitação de elevação através do hook MCP Desktop Commander
func TestMCPDesktopCommanderElevationRequest(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_desktop_commander_elevation_request_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Desktop Commander
	dcHook := mcp.NewDesktopCommanderHook(mockElevationManager)

	// Configurar observabilidade
	dcHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de solicitação de elevação para os testes
	elevationResult := &elevation.ElevationResult{
		ElevationID:    "elev-dc-request-001",
		UserID:         "user:admin:789",
		ElevatedRoles:  []string{"desktop_admin"},
		ElevatedScopes: []string{"desktop:file:write", "desktop:config:modify"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		ElevationToken: "new-dc-token-123",
		ApprovedBy:     "user:manager:456",
		ApprovalTime:   baseTime,
		AuditMetadata: map[string]interface{}{
			"tenant_id":   "tenant_angola_1",
			"market":      "angola",
			"request_ip":  "192.168.1.100",
			"user_agent":  "MCP Desktop Commander Client/1.0",
		},
	}

	// Configurar mock para a solicitação de elevação
	mockElevationManager.On(
		"RequestElevation",
		mock.Anything,
		mock.MatchedBy(func(req *elevation.ElevationRequest) bool {
			return req.UserID == "user:admin:789" &&
				len(req.RequestedScopes) > 0 &&
				req.Duration.Minutes() <= 60 // Máximo 1 hora
		}),
	).Return(elevationResult, nil).Once()

	// Preparar requisição de elevação
	elevationRequest := &mcp.ElevationRequestPayload{
		UserID:          "user:admin:789",
		RequestedScopes: []string{"desktop:file:write", "desktop:config:modify"},
		Justification:   "Manutenção de configuração do sistema",
		Duration:        "30m",
		EmergencyAccess: true,
		Context: map[string]interface{}{
			"tenant_id":    "tenant_angola_1",
			"market":       "angola",
			"business_unit": "operations",
			"client_ip":    "192.168.1.100",
			"user_agent":   "MCP Desktop Commander Client/1.0",
		},
	}

	// Serializar requisição
	requestJSON, err := json.Marshal(elevationRequest)
	require.NoError(t, err, "Erro ao serializar requisição de elevação")

	// Executar o teste
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "DesktopCommanderElevationRequest")

	// Chamar a função de solicitação de elevação
	resultJSON, err := dcHook.RequestDesktopElevation(testCtx, requestJSON)
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
	obs.RecordTestEnd(testCtx, "DesktopCommanderElevationRequest", err == nil, time.Since(time.Now()))
}

// TestMCPDesktopCommanderElevationUsage testa o registro de uso da elevação através do hook MCP Desktop Commander
func TestMCPDesktopCommanderElevationUsage(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_desktop_commander_elevation_usage_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Desktop Commander
	dcHook := mcp.NewDesktopCommanderHook(mockElevationManager)

	// Configurar observabilidade
	dcHook.ConfigureObservability(obs.Logger(), obs.Tracer())

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultado de verificação para o teste
	validVerificationResult := &elevation.VerificationResult{
		ElevationID:    "elev-dc-usage-001",
		UserID:         "user:admin:123",
		ElevatedRoles:  []string{"desktop_admin"},
		ElevatedScopes: []string{"desktop:file:write"},
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
		"dc-usage-token-123",
		[]string{"desktop:file:write"},
	).Return(validVerificationResult, nil).Once()

	// Preparar requisição para comando Desktop Commander com elevação
	requestPayload := map[string]interface{}{
		"path":    "/etc/hosts",
		"content": "127.0.0.1 localhost\n192.168.1.10 server-1\n",
		"mode":    "rewrite",
	}

	requestJSON, err := json.Marshal(requestPayload)
	require.NoError(t, err, "Erro ao serializar payload de requisição")

	// Executar o teste
	ctx := context.Background()
	testCtx := obs.RecordTestStart(ctx, "DesktopCommanderElevationUsage")

	// Configurar contexto com token de elevação
	hookCtx := mcp.WithElevationToken(testCtx, "dc-usage-token-123")

	// Chamar o hook para autorizar o comando com elevação
	allowed, reason, err := dcHook.AuthorizeDesktopCommand(hookCtx, "write_file", requestJSON)
	require.NoError(t, err, "A autorização não deveria falhar")

	// Verificar resultados
	assert.True(t, allowed, "O comando deveria ser autorizado")
	assert.Contains(t, reason, "Elevação verificada com sucesso", "Razão de autorização incorreta")

	// Verificar que o token e escopo foram verificados
	mockElevationManager.AssertExpectations(t)

	// Registrar conclusão do teste
	obs.RecordTestEnd(testCtx, "DesktopCommanderElevationUsage", err == nil, time.Since(time.Now()))
}

// TestMCPDesktopCommanderMultiTenantIsolation testa o isolamento multi-tenant no hook MCP Desktop Commander
func TestMCPDesktopCommanderMultiTenantIsolation(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("mcp_desktop_commander_multi_tenant_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())

	// Configurar mocks
	mockElevationManager := new(MockPrivilegeElevationManager)

	// Criar o hook MCP Desktop Commander com isolamento multi-tenant
	dcHook := mcp.NewDesktopCommanderHook(mockElevationManager)
	dcHook.ConfigureObservability(obs.Logger(), obs.Tracer())
	dcHook.EnableMultiTenantIsolation(true)

	// Configurar mapeamento de tenants para diretórios permitidos
	dcHook.ConfigureTenantDirectories(map[string][]string{
		"tenant_angola_1": {"/data/angola/", "C:\\data\\angola\\"},
		"tenant_brazil_1": {"/data/brazil/", "C:\\data\\brazil\\"},
	})

	// Definir base de tempo para os testes
	baseTime := time.Now()

	// Configurar resultados de verificação para diferentes tenants
	angolaTenantResult := &elevation.VerificationResult{
		ElevationID:    "elev-dc-angola-001",
		UserID:         "user:admin:123",
		ElevatedRoles:  []string{"desktop_admin"},
		ElevatedScopes: []string{"desktop:file:write"},
		ExpirationTime: baseTime.Add(30 * time.Minute),
		IsValid:        true,
		AuditMetadata: map[string]interface{}{
			"tenant_id": "tenant_angola_1",
			"market":    "angola",
		},
	}

	brazilTenantResult := &elevation.VerificationResult{
		ElevationID:    "elev-dc-brazil-001",
		UserID:         "user:admin:456",
		ElevatedRoles:  []string{"desktop_admin"},
		ElevatedScopes: []string{"desktop:file:write"},
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
		filePath       string
		mockSetup      func()
		expectedAllowed bool
		expectedReason  string
	}{
		{
			name:           "Acesso ao diretório do tenant correto - Angola",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			filePath:       "/data/angola/config.json",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"desktop:file:write"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso cruzado entre tenants - Angola tentando acessar Brasil",
			elevationToken: "angola-tenant-token-123",
			tenantID:       "tenant_angola_1",
			filePath:       "/data/brazil/config.json",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"angola-tenant-token-123",
					[]string{"desktop:file:write"},
				).Return(angolaTenantResult, nil).Once()
			},
			expectedAllowed: false,
			expectedReason:  "Acesso negado: isolamento multi-tenant",
		},
		{
			name:           "Acesso ao diretório do tenant correto - Brasil",
			elevationToken: "brazil-tenant-token-456",
			tenantID:       "tenant_brazil_1",
			filePath:       "/data/brazil/config.json",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"brazil-tenant-token-456",
					[]string{"desktop:file:write"},
				).Return(brazilTenantResult, nil).Once()
			},
			expectedAllowed: true,
			expectedReason:  "Elevação verificada com sucesso",
		},
		{
			name:           "Acesso cruzado entre tenants - Brasil tentando acessar Angola",
			elevationToken: "brazil-tenant-token-456",
			tenantID:       "tenant_brazil_1",
			filePath:       "/data/angola/config.json",
			mockSetup: func() {
				mockElevationManager.On(
					"VerifyElevation",
					mock.Anything,
					"brazil-tenant-token-456",
					[]string{"desktop:file:write"},
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
				"path":    tc.filePath,
				"content": "test content",
				"mode":    "rewrite",
			}

			requestJSON, err := json.Marshal(requestPayload)
			require.NoError(t, err, "Erro ao serializar payload de requisição")

			// Configurar contexto com token de elevação e tenant
			hookCtx := mcp.WithElevationToken(testCtx, tc.elevationToken)
			hookCtx = mcp.WithTenantID(hookCtx, tc.tenantID)

			// Executar o hook
			allowed, reason, err := dcHook.AuthorizeDesktopCommand(hookCtx, "write_file", requestJSON)
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