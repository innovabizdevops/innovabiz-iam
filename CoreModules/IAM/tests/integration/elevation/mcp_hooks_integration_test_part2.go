package integration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/innovabizdevops/innovabiz-iam/authorization/elevation"
	"github.com/innovabizdevops/innovabiz-iam/hooks/mcp"
)

// testDesktopCommanderIsolation testa o isolamento multi-tenant no hook Desktop Commander
func testDesktopCommanderIsolation(ctx context.Context, desktopCommanderHook *mcp.DesktopCommanderHook, 
	elevationService *elevation.PrivilegeElevationService) func(t *testing.T) {
	return func(t *testing.T) {
		// Definir IDs para o teste
		userID := "user:admin:e2e-test-123"

		// Habilitar isolamento multi-tenant
		desktopCommanderHook.EnableMultiTenantIsolation(true)

		// Configurar mapeamento de tenants para diretórios
		desktopCommanderHook.ConfigureTenantDirectories(map[string][]string{
			"tenant_angola_1": {"/data/angola/", "C:\\data\\angola\\"},
			"tenant_brazil_1": {"/data/brazil/", "C:\\data\\brazil\\"},
			"tenant_mozambique_1": {"/data/mozambique/", "C:\\data\\mozambique\\"},
		})

		// 1. Solicitar elevação para tenant Angola
		angolaTenantCtx := mcp.WithTenantID(ctx, "tenant_angola_1")

		elevationRequest := &mcp.ElevationRequestPayload{
			UserID:          userID,
			RequestedScopes: []string{"desktop:file:write"},
			Justification:   "Atualização de configuração Angola",
			Duration:        "15m",
			EmergencyAccess: false,
			Context: map[string]interface{}{
				"tenant_id":     "tenant_angola_1",
				"market":        "angola",
				"business_unit": "operations",
				"regulatory_framework": "Angola Financial Services Authority",
			},
		}

		requestJSON, err := json.Marshal(elevationRequest)
		require.NoError(t, err, "Erro ao serializar requisição de elevação")

		resultJSON, err := desktopCommanderHook.RequestDesktopElevation(angolaTenantCtx, requestJSON)
		require.NoError(t, err, "Solicitação de elevação não deveria falhar")

		var angolaTenantResponse mcp.ElevationResponse
		err = json.Unmarshal(resultJSON, &angolaTenantResponse)
		require.NoError(t, err, "Erro ao deserializar resposta de elevação")

		// 2. Tentar acessar diretório do próprio tenant (permitido)
		angolaFilePayload := map[string]interface{}{
			"path":    "/data/angola/config.json",
			"content": "{ \"config\": \"angola\" }",
			"mode":    "rewrite",
		}

		angolaJSON, err := json.Marshal(angolaFilePayload)
		require.NoError(t, err, "Erro ao serializar payload Angola")

		elevatedAngolaCtx := mcp.WithElevationToken(angolaTenantCtx, angolaTenantResponse.ElevationToken)

		allowed, reason, err := desktopCommanderHook.AuthorizeDesktopCommand(elevatedAngolaCtx, "write_file", angolaJSON)
		require.NoError(t, err, "A autorização não deveria falhar")
		
		assert.True(t, allowed, "Acesso ao próprio tenant deveria ser permitido")
		assert.Contains(t, reason, "Elevação verificada com sucesso", "Razão incorreta")

		// 3. Tentar acessar diretório de outro tenant (negado)
		brazilFilePayload := map[string]interface{}{
			"path":    "/data/brazil/config.json",
			"content": "{ \"config\": \"cross-tenant-access\" }",
			"mode":    "rewrite",
		}

		brazilJSON, err := json.Marshal(brazilFilePayload)
		require.NoError(t, err, "Erro ao serializar payload Brazil")

		allowed, reason, err = desktopCommanderHook.AuthorizeDesktopCommand(elevatedAngolaCtx, "write_file", brazilJSON)
		require.NoError(t, err, "A autorização não deveria falhar tecnicamente")
		
		assert.False(t, allowed, "Acesso cross-tenant deveria ser negado")
		assert.Contains(t, reason, "Acesso negado: isolamento multi-tenant", "Razão de negação incorreta")

		// 4. Verificar log de auditoria para tentativa de acesso cross-tenant
		auditEvents, err := elevationService.QueryAuditEvents(ctx, map[string]interface{}{
			"elevation_id": angolaTenantResponse.ElevationID,
			"event_type":   "access_denied",
			"reason":       "cross_tenant_access",
		})

		require.NoError(t, err, "Consulta de auditoria não deveria falhar")
		assert.NotEmpty(t, auditEvents, "Deveria haver registro de auditoria para tentativa de acesso cross-tenant")
		
		// 5. Verificar conformidade com regulações específicas do mercado
		auditTrail, err := elevationService.GetComplianceAuditTrail(ctx, angolaTenantResponse.ElevationID)
		require.NoError(t, err, "Obtenção de trilha de auditoria de conformidade não deveria falhar")
		
		assert.Contains(t, auditTrail.AppliedRegulations, "Angola Financial Services Authority", 
			"Regulação específica de Angola deveria ser aplicada")
		assert.Contains(t, auditTrail.ComplianceChecks, "multi_tenant_isolation", 
			"Verificação de isolamento multi-tenant deveria ser registrada")
	}
}

// testGitHubMFAIntegration testa a integração MFA com o hook GitHub
func testGitHubMFAIntegration(ctx context.Context, githubHook *mcp.GitHubHook, 
	elevationService *elevation.PrivilegeElevationService) func(t *testing.T) {
	return func(t *testing.T) {
		// Definir IDs para o teste
		userID := "user:admin:e2e-test-123"
		tenantID := "tenant_angola_1"

		// Configurar MFA para o hook GitHub
		githubHook.RequireMFAForProtectedBranches(true)
		
		// Preparar payload para push em branch protegido
		pushPayload := map[string]interface{}{
			"owner":  "innovabizdevops",
			"repo":   "innovabiz-iam",
			"branch": "production", // Branch protegido
			"files": []map[string]interface{}{
				{
					"path":    "config/security.yaml",
					"content": "security_level: high",
				},
			},
			"message": "Update security config",
		}

		pushJSON, err := json.Marshal(pushPayload)
		require.NoError(t, err, "Erro ao serializar payload de push")

		// 1. Solicitar elevação para GitHub
		githubCtx := mcp.WithTenantID(ctx, tenantID)

		elevationRequest := &mcp.ElevationRequestPayload{
			UserID:          userID,
			RequestedScopes: []string{"github:repo:push"},
			Justification:   "Atualização de segurança urgente",
			Duration:        "15m",
			EmergencyAccess: true,
			Context: map[string]interface{}{
				"tenant_id":     tenantID,
				"market":        "angola",
				"business_unit": "development",
				"repository":    "innovabizdevops/innovabiz-iam",
				"branch":        "production",
				"risk_level":    "high",
			},
		}

		requestJSON, err := json.Marshal(elevationRequest)
		require.NoError(t, err, "Erro ao serializar requisição de elevação")

		resultJSON, err := githubHook.RequestGitHubElevation(githubCtx, requestJSON)
		require.NoError(t, err, "Solicitação de elevação não deveria falhar")

		var githubResponse mcp.ElevationResponse
		err = json.Unmarshal(resultJSON, &githubResponse)
		require.NoError(t, err, "Erro ao deserializar resposta de elevação")

		// 2. Simular desafio MFA
		mfaChallenge, err := elevationService.GenerateMFAChallenge(ctx, githubResponse.ElevationID, "totp")
		require.NoError(t, err, "Geração de desafio MFA não deveria falhar")
		assert.NotEmpty(t, mfaChallenge.ChallengeID, "ID de desafio MFA não deveria estar vazio")

		// 3. Simular verificação MFA bem-sucedida
		err = elevationService.VerifyMFAChallenge(ctx, mfaChallenge.ChallengeID, "123456") // Código simulado
		require.NoError(t, err, "Verificação MFA não deveria falhar")

		// 4. Autorizar operação após MFA
		elevatedGithubCtx := mcp.WithElevationToken(githubCtx, githubResponse.ElevationToken)

		allowed, reason, err := githubHook.AuthorizeGitHubCommand(elevatedGithubCtx, "push_files", pushJSON)
		require.NoError(t, err, "A autorização não deveria falhar")

		// Validar que o comando foi autorizado após MFA
		assert.True(t, allowed, "Comando deveria ser permitido após MFA")
		assert.Contains(t, reason, "Elevação verificada com sucesso", "Razão incorreta")

		// 5. Verificar registros de auditoria
		auditEvents, err := elevationService.QueryAuditEvents(ctx, map[string]interface{}{
			"elevation_id": githubResponse.ElevationID,
			"event_type":   "mfa_verification",
			"success":      true,
		})

		require.NoError(t, err, "Consulta de auditoria não deveria falhar")
		assert.NotEmpty(t, auditEvents, "Deveria haver registro de auditoria para verificação MFA")
		
		// 6. Verificar registro detalhado de branch protegido
		usageEvents, err := elevationService.QueryAuditEvents(ctx, map[string]interface{}{
			"elevation_id": githubResponse.ElevationID,
			"event_type":   "elevation_usage",
			"metadata.protected_branch": "production",
		})

		require.NoError(t, err, "Consulta de auditoria não deveria falhar")
		assert.NotEmpty(t, usageEvents, "Deveria haver registro de auditoria para branch protegido")
	}
}

// testProtectedBranchAuditing testa a auditoria detalhada de branches protegidos
func testProtectedBranchAuditing(ctx context.Context, githubHook *mcp.GitHubHook,
	elevationService *elevation.PrivilegeElevationService) func(t *testing.T) {
	return func(t *testing.T) {
		// Implementação do teste para auditoria de branches protegidos
		// Este teste verificaria em detalhes os logs de auditoria para operações em branches protegidos
		// incluindo rastreabilidade, conformidade regulatória, e relatórios de conformidade
		t.Skip("Implementação pendente para teste de auditoria detalhada")
	}
}

// testFigmaTraceability testa a rastreabilidade de operações no hook Figma
func testFigmaTraceability(ctx context.Context, figmaHook *mcp.FigmaHook,
	elevationService *elevation.PrivilegeElevationService) func(t *testing.T) {
	return func(t *testing.T) {
		// Implementação do teste para rastreabilidade de operações no Figma
		// Este teste verificaria a rastreabilidade completa das operações no Figma
		// incluindo quem fez o quê, quando, por que, e com qual aprovação
		t.Skip("Implementação pendente para teste de rastreabilidade do Figma")
	}
}