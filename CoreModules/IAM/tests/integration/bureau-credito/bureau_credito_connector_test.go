// ==============================================================================
// Nome: bureau_credito_connector_test.go
// Descrição: Testes de integração para o conector do Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

package bureautests

import (
	"context"
	"testing"
	"time"

	"github.com/innovabiz/iam/common/logging"
	"github.com/innovabiz/iam/common/metrics"
	"github.com/innovabiz/iam/common/security"
	"github.com/innovabiz/iam/common/tracing"
	"github.com/innovabiz/iam/integration/bureau-credito"
	"github.com/innovabiz/iam/models"
	"github.com/innovabiz/iam/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestVincularUsuario testa a funcionalidade de vinculação de usuário com Bureau de Créditos
func TestVincularUsuario(t *testing.T) {
	// Configurar mocks
	mockBCClient := new(mocks.MockBureauCreditoClient)
	mockDataClient := new(mocks.MockDataCoreClient)
	mockLogger := new(mocks.MockLogger)
	mockMetrics := new(mocks.MockMetricsCollector)
	mockTracer := new(mocks.MockTracer)
	mockTokenService := new(mocks.MockTokenService)
	mockSpan := new(mocks.MockSpan)

	// Configurar comportamento esperado dos mocks
	mockTracer.On("StartSpan", mock.Anything, "BureauCreditoConnector.VincularUsuario").Return(context.Background(), mockSpan)
	mockSpan.On("End").Return()
	mockMetrics.On("CountEvent", "bureau_credito_vincular_usuario_attempt").Return()

	// Mock para busca de usuário
	mockUsuario := &models.Usuario{
		ID:              "usuario-123",
		Email:           "teste@exemplo.com",
		NomeCompleto:    "Usuário Teste",
		DocumentoPrincipal: "12345678901",
		TipoDocumento:   "CPF",
	}
	mockDataClient.On("GetUsuario", mock.Anything, "usuario-123").Return(mockUsuario, nil)

	// Mock para registro no Bureau de Créditos
	mockBCUsuario := struct {
		ID       string
		TenantID string
	}{
		ID:       "bc-usuario-456",
		TenantID: "bc-tenant-789",
	}
	mockBCClient.On(
		"RegisterUsuario",
		mock.Anything,
		mock.MatchedBy(func(req interface{}) bool {
			return true // Simplificado para o teste
		}),
	).Return(mockBCUsuario, nil)

	// Mock para criação de identidade
	mockDataClient.On(
		"CreateIntegrationIdentity",
		mock.Anything,
		mock.MatchedBy(func(identity *models.IntegrationIdentity) bool {
			return identity.UsuarioID == "usuario-123" &&
				identity.TenantID == "tenant-123" &&
				identity.IntegrationType == "bureau_credito"
		}),
	).Return(nil)

	mockLogger.On("WithField", "module", mock.Anything).Return(mockLogger)
	mockMetrics.On("CountEvent", "bureau_credito_vincular_usuario_success").Return()

	// Configurar conector com os mocks
	config := &bureaupkg.BureauCreditoConfig{
		BaseURL:         "https://api.bureaucredito.exemplo.com",
		APIVersion:      "v1",
		Timeout:         5 * time.Second,
		MaxRetries:      3,
		EnableCache:     false,
		CacheExpiration: 30 * time.Minute,
	}

	connector := bureaupkg.NewBureauCreditoConnector(
		config,
		mockBCClient,
		mockDataClient,
		mockLogger,
		mockMetrics,
		mockTracer,
		mockTokenService,
	)

	// Executar a função a ser testada
	ctx := context.Background()
	identity, err := connector.VincularUsuario(
		ctx,
		"usuario-123",
		"tenant-123",
		"CONSULTA",
		"BASICO",
		map[string]interface{}{
			"observacao": "Teste de integração",
		},
	)

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	assert.Equal(t, "usuario-123", identity.UsuarioID)
	assert.Equal(t, "tenant-123", identity.TenantID)
	assert.Equal(t, "bureau_credito", identity.IntegrationType)
	assert.Equal(t, "bc-usuario-456", identity.ExternalID)

	// Verificar se todos os mocks foram chamados conforme esperado
	mockBCClient.AssertExpectations(t)
	mockDataClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

// TestCriarAutorizacaoConsulta testa a funcionalidade de criação de autorização de consulta
func TestCriarAutorizacaoConsulta(t *testing.T) {
	// Configurar mocks
	mockBCClient := new(mocks.MockBureauCreditoClient)
	mockDataClient := new(mocks.MockDataCoreClient)
	mockLogger := new(mocks.MockLogger)
	mockMetrics := new(mocks.MockMetricsCollector)
	mockTracer := new(mocks.MockTracer)
	mockTokenService := new(mocks.MockTokenService)
	mockSpan := new(mocks.MockSpan)

	// Configurar comportamento esperado dos mocks
	mockTracer.On("StartSpan", mock.Anything, "BureauCreditoConnector.CriarAutorizacaoConsulta").Return(context.Background(), mockSpan)
	mockSpan.On("End").Return()
	mockMetrics.On("CountEvent", "bureau_credito_criar_autorizacao_attempt").Return()

	// Mock para busca de identidade
	mockIdentity := &models.IntegrationIdentity{
		ID:               "identity-123",
		UsuarioID:        "usuario-123",
		TenantID:         "tenant-123",
		IntegrationType:  "bureau_credito",
		ExternalID:       "bc-usuario-456",
		ExternalTenantID: "bc-tenant-789",
		ProfileType:      "CONSULTA",
		AccessLevel:      "BASICO",
		Status:           "ativo",
	}
	mockDataClient.On("GetIntegrationIdentity", mock.Anything, "identity-123").Return(mockIdentity, nil)

	// Mock para criação de autorização no Bureau de Créditos
	mockBCAutorizacao := struct {
		ID string
	}{
		ID: "bc-autorizacao-789",
	}
	mockBCClient.On(
		"CreateAutorizacao",
		mock.Anything,
		mock.MatchedBy(func(req interface{}) bool {
			return true // Simplificado para o teste
		}),
	).Return(mockBCAutorizacao, nil)

	// Mock para criação de autorização no banco de dados
	mockDataClient.On(
		"CreateBureauAutorizacao",
		mock.Anything,
		mock.MatchedBy(func(autorizacao *models.BureauAutorizacao) bool {
			return autorizacao.IdentityID == "identity-123" &&
				autorizacao.TipoConsulta == "SIMPLES" &&
				autorizacao.Finalidade == "Avaliação de crédito"
		}),
	).Return(nil)

	mockLogger.On("WithField", "module", mock.Anything).Return(mockLogger)
	mockMetrics.On("CountEvent", "bureau_credito_criar_autorizacao_success").Return()

	// Configurar conector com os mocks
	config := &bureaupkg.BureauCreditoConfig{
		BaseURL:         "https://api.bureaucredito.exemplo.com",
		APIVersion:      "v1",
		Timeout:         5 * time.Second,
		MaxRetries:      3,
		EnableCache:     false,
		CacheExpiration: 30 * time.Minute,
	}

	connector := bureaupkg.NewBureauCreditoConnector(
		config,
		mockBCClient,
		mockDataClient,
		mockLogger,
		mockMetrics,
		mockTracer,
		mockTokenService,
	)

	// Executar a função a ser testada
	ctx := context.Background()
	autorizacao, err := connector.CriarAutorizacaoConsulta(
		ctx,
		"identity-123",
		"SIMPLES",
		"Avaliação de crédito",
		"Solicitação de empréstimo pelo cliente",
		30,
		"operador-123",
	)

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, autorizacao)
	assert.Equal(t, "bc-autorizacao-789", autorizacao.ID)
	assert.Equal(t, "identity-123", autorizacao.IdentityID)
	assert.Equal(t, "SIMPLES", autorizacao.TipoConsulta)
	assert.Equal(t, "Avaliação de crédito", autorizacao.Finalidade)
	assert.Equal(t, "ativa", autorizacao.Status)

	// Verificar se todos os mocks foram chamados conforme esperado
	mockBCClient.AssertExpectations(t)
	mockDataClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}