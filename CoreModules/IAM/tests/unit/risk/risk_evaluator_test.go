// Package risk_test contém testes unitários para o componente de avaliação de risco do MCP-IAM
package risk_test

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	
	"github.com/innovabizdevops/innovabiz-iam/authorization/risk"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// MockRiskDataProvider simula o fornecedor de dados para avaliação de risco
type MockRiskDataProvider struct {
	mock.Mock
}

func (m *MockRiskDataProvider) GetUserRiskProfile(ctx context.Context, userID string) (*risk.UserRiskProfile, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*risk.UserRiskProfile), args.Error(1)
}

func (m *MockRiskDataProvider) GetResourceSensitivity(ctx context.Context, resourceID string) (*risk.ResourceSensitivity, error) {
	args := m.Called(ctx, resourceID)
	return args.Get(0).(*risk.ResourceSensitivity), args.Error(1)
}

func (m *MockRiskDataProvider) GetLocationTrustScore(ctx context.Context, location string) (float64, error) {
	args := m.Called(ctx, location)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRiskDataProvider) GetIPTrustScore(ctx context.Context, ipAddress string) (float64, error) {
	args := m.Called(ctx, ipAddress)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRiskDataProvider) GetDeviceTrustScore(ctx context.Context, deviceID string) (float64, error) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRiskDataProvider) IsBusinessHours(ctx context.Context, accessTime time.Time, location string) (bool, error) {
	args := m.Called(ctx, accessTime, location)
	return args.Get(0).(bool), args.Error(1)
}

// TestAdaptiveRiskEvaluator_EvaluateRisk verifica se o avaliador de risco calcula
// corretamente a pontuação de risco com base em múltiplos fatores contextuais
func TestAdaptiveRiskEvaluator_EvaluateRisk(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("risk_evaluator_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para diferentes cenários de risco
	testCases := []struct {
		name               string
		userID             string
		resourceID         string
		action             string
		location           string
		ipAddress          string
		deviceID           string
		accessTime         time.Time
		isRecognizedDevice bool
		userRiskProfile    *risk.UserRiskProfile
		resourceSensitivity *risk.ResourceSensitivity
		locationTrustScore float64
		ipTrustScore       float64
		deviceTrustScore   float64
		isBusinessHours    bool
		expectedRiskScore  float64
		expectedRiskLevel  risk.RiskLevel
	}{
		{
			name:               "Baixo Risco - Acesso Comum",
			userID:             "user:operator:123",
			resourceID:         "crm:customer:list",
			action:             "read",
			location:           "Angola",
			ipAddress:          "192.168.1.100", // IP interno confiável
			deviceID:           "device:registered:456",
			accessTime:         time.Date(2025, 8, 5, 14, 30, 0, 0, time.UTC), // Horário comercial
			isRecognizedDevice: true,
			userRiskProfile: &risk.UserRiskProfile{
				ID:                "user:operator:123",
				BaseRiskScore:     0.1, // Usuário confiável
				AnomalyThreshold:  0.3,
				AuthFailures:      0,   // Sem falhas recentes
				LastAuthenticated: time.Now().Add(-2 * time.Hour),
				AccessPatterns: []*risk.AccessPattern{
					{
						ResourceType: "crm:customer",
						Action:       "read",
						Frequency:    risk.FrequencyHigh, // Acesso comum
					},
				},
			},
			resourceSensitivity: &risk.ResourceSensitivity{
				ID:              "crm:customer:list",
				SensitivityLevel: risk.SensitivityLow,  // Baixa sensibilidade
				Classification:  "public",
				RequiresMFA:     false,
				ContextChecks:   []string{"business_hours"},
			},
			locationTrustScore: 0.9,  // Local confiável
			ipTrustScore:       0.95, // IP confiável
			deviceTrustScore:   0.9,  // Dispositivo confiável
			isBusinessHours:    true,
			expectedRiskScore:  0.15, // Risco muito baixo
			expectedRiskLevel:  risk.RiskLevelLow,
		},
		{
			name:               "Risco Médio - Horário Incomum",
			userID:             "user:operator:123",
			resourceID:         "payment_gateway:transaction:list",
			action:             "read",
			location:           "Angola",
			ipAddress:          "192.168.1.100",
			deviceID:           "device:registered:456",
			accessTime:         time.Date(2025, 8, 5, 22, 30, 0, 0, time.UTC), // Fora do horário comercial
			isRecognizedDevice: true,
			userRiskProfile: &risk.UserRiskProfile{
				ID:                "user:operator:123",
				BaseRiskScore:     0.1,
				AnomalyThreshold:  0.3,
				AuthFailures:      1, // Uma falha recente
				LastAuthenticated: time.Now().Add(-8 * time.Hour),
				AccessPatterns: []*risk.AccessPattern{
					{
						ResourceType: "payment_gateway:transaction",
						Action:       "read",
						Frequency:    risk.FrequencyMedium, // Acesso ocasional
					},
				},
			},
			resourceSensitivity: &risk.ResourceSensitivity{
				ID:              "payment_gateway:transaction:list",
				SensitivityLevel: risk.SensitivityMedium, // Sensibilidade média
				Classification:  "internal",
				RequiresMFA:     true,
				ContextChecks:   []string{"business_hours", "location_check"},
			},
			locationTrustScore: 0.9,
			ipTrustScore:       0.8,
			deviceTrustScore:   0.9,
			isBusinessHours:    false, // Fora do horário comercial
			expectedRiskScore:  0.45,  // Risco médio
			expectedRiskLevel:  risk.RiskLevelMedium,
		},
		{
			name:               "Alto Risco - Acesso Sensível de Local Não Confiável",
			userID:             "user:manager:456",
			resourceID:         "payment_gateway:config:update",
			action:             "write",
			location:           "Desconhecida", // Local não confiável
			ipAddress:          "203.0.113.42", // IP externo não confiável
			deviceID:           "device:new:789",
			accessTime:         time.Date(2025, 8, 5, 2, 15, 0, 0, time.UTC), // Madrugada
			isRecognizedDevice: false,
			userRiskProfile: &risk.UserRiskProfile{
				ID:                "user:manager:456",
				BaseRiskScore:     0.2,
				AnomalyThreshold:  0.3,
				AuthFailures:      2, // Múltiplas falhas recentes
				LastAuthenticated: time.Now().Add(-24 * time.Hour),
				AccessPatterns: []*risk.AccessPattern{
					{
						ResourceType: "payment_gateway:config",
						Action:       "write",
						Frequency:    risk.FrequencyLow, // Acesso raro
					},
				},
			},
			resourceSensitivity: &risk.ResourceSensitivity{
				ID:              "payment_gateway:config:update",
				SensitivityLevel: risk.SensitivityHigh, // Alta sensibilidade
				Classification:  "restricted",
				RequiresMFA:     true,
				ContextChecks:   []string{"business_hours", "location_check", "device_check"},
			},
			locationTrustScore: 0.3,  // Local não confiável
			ipTrustScore:       0.4,  // IP não confiável
			deviceTrustScore:   0.3,  // Dispositivo não confiável
			isBusinessHours:    false, // Fora do horário comercial
			expectedRiskScore:  0.85,  // Risco muito alto
			expectedRiskLevel:  risk.RiskLevelHigh,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mock do RiskDataProvider
			mockDataProvider := new(MockRiskDataProvider)
			mockDataProvider.On("GetUserRiskProfile", mock.Anything, tc.userID).
				Return(tc.userRiskProfile, nil)
				
			mockDataProvider.On("GetResourceSensitivity", mock.Anything, tc.resourceID).
				Return(tc.resourceSensitivity, nil)
				
			mockDataProvider.On("GetLocationTrustScore", mock.Anything, tc.location).
				Return(tc.locationTrustScore, nil)
				
			mockDataProvider.On("GetIPTrustScore", mock.Anything, tc.ipAddress).
				Return(tc.ipTrustScore, nil)
				
			mockDataProvider.On("GetDeviceTrustScore", mock.Anything, tc.deviceID).
				Return(tc.deviceTrustScore, nil)
				
			mockDataProvider.On("IsBusinessHours", mock.Anything, tc.accessTime, tc.location).
				Return(tc.isBusinessHours, nil)
			
			// Criar o avaliador de risco
			evaluator := risk.NewAdaptiveRiskEvaluator(mockDataProvider)
			
			// Criar requisição de avaliação de risco
			request := &risk.RiskEvaluationRequest{
				UserID:             tc.userID,
				ResourceID:         tc.resourceID,
				Action:             tc.action,
				Location:           tc.location,
				IPAddress:          tc.ipAddress,
				DeviceID:           tc.deviceID,
				AccessTime:         tc.accessTime,
				IsRecognizedDevice: tc.isRecognizedDevice,
			}
			
			// Executar avaliação de risco
			result, err := evaluator.EvaluateRisk(testCtx, request)
			
			// Verificar resultados
			require.NoError(t, err, "Avaliação de risco não deveria falhar")
			assert.InDelta(t, tc.expectedRiskScore, result.RiskScore, 0.05, "Pontuação de risco calculada incorretamente")
			assert.Equal(t, tc.expectedRiskLevel, result.RiskLevel, "Nível de risco incorreto")
			
			// Verificar se todos os fatores de risco foram analisados
			assert.NotEmpty(t, result.RiskFactors, "Deveria incluir fatores de risco na análise")
			assert.NotEmpty(t, result.MitigationFactors, "Deveria incluir fatores de mitigação na análise")
			
			// Verificar se recomendações de segurança estão presentes para risco médio ou alto
			if tc.expectedRiskLevel >= risk.RiskLevelMedium {
				assert.NotEmpty(t, result.SecurityRecommendations, "Deveria incluir recomendações de segurança para risco médio/alto")
			}
			
			// Verificar chamadas ao mock
			mockDataProvider.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
}

// TestAdaptiveRiskEvaluator_AnomalyDetection verifica a detecção de anomalias
// no comportamento do usuário e padrões de acesso incomuns
func TestAdaptiveRiskEvaluator_AnomalyDetection(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("risk_anomaly_detection_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Definir cenários para detecção de anomalias
	testCases := []struct {
		name                string
		accessPatterns      []*risk.AccessPattern
		currentResourceType string
		currentAction       string
		expectedAnomaly     bool
		anomalyScore        float64
	}{
		{
			name: "Padrão Normal - Sem Anomalia",
			accessPatterns: []*risk.AccessPattern{
				{
					ResourceType: "crm:customer",
					Action:       "read",
					Frequency:    risk.FrequencyHigh,
				},
			},
			currentResourceType: "crm:customer",
			currentAction:       "read",
			expectedAnomaly:     false,
			anomalyScore:        0.1,
		},
		{
			name: "Padrão Incomum - Anomalia Detectada",
			accessPatterns: []*risk.AccessPattern{
				{
					ResourceType: "crm:customer",
					Action:       "read",
					Frequency:    risk.FrequencyHigh,
				},
				{
					ResourceType: "payment_gateway:config",
					Action:       "write",
					Frequency:    risk.FrequencyVeryLow,
				},
			},
			currentResourceType: "payment_gateway:config",
			currentAction:       "write",
			expectedAnomaly:     true,
			anomalyScore:        0.7,
		},
		{
			name: "Nova Ação em Recurso Familiar - Anomalia Leve",
			accessPatterns: []*risk.AccessPattern{
				{
					ResourceType: "crm:customer",
					Action:       "read",
					Frequency:    risk.FrequencyHigh,
				},
			},
			currentResourceType: "crm:customer",
			currentAction:       "write", // Nova ação
			expectedAnomaly:     true,
			anomalyScore:        0.4,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar avaliador de anomalias
			anomalyEvaluator := risk.NewAnomalyDetector()
			
			// Criar perfil de usuário com padrões de acesso
			userProfile := &risk.UserRiskProfile{
				ID:               "user:test:123",
				AccessPatterns:   tc.accessPatterns,
				AnomalyThreshold: 0.3,
			}
			
			// Avaliar anomalia
			isAnomaly, score := anomalyEvaluator.EvaluateAnomaly(
				testCtx,
				userProfile,
				tc.currentResourceType,
				tc.currentAction,
			)
			
			// Verificar resultados
			assert.Equal(t, tc.expectedAnomaly, isAnomaly, "Detecção de anomalia incorreta")
			assert.InDelta(t, tc.anomalyScore, score, 0.1, "Pontuação de anomalia incorreta")
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, true, time.Since(time.Now()))
		})
	}
}

// TestAdaptiveRiskEvaluator_MarketSpecificRules testa se o avaliador de risco
// aplica regras específicas para diferentes mercados geográficos
func TestAdaptiveRiskEvaluator_MarketSpecificRules(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("risk_market_rules_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para diferentes mercados
	testCases := []struct {
		name          string
		market        string
		resourceType  string
		action        string
		amount        float64
		expectedRules []string
	}{
		{
			name:         "Angola - Regras Específicas SADC",
			market:       "Angola",
			resourceType: "payment_gateway:transaction",
			action:       "create",
			amount:       50000,
			expectedRules: []string{
				"aml_angola",
				"sadc_transaction_limits",
				"angola_central_bank_reporting",
			},
		},
		{
			name:         "Brasil - Regras Específicas BACEN",
			market:       "Brasil",
			resourceType: "payment_gateway:transaction",
			action:       "create",
			amount:       50000,
			expectedRules: []string{
				"aml_brasil",
				"bacen_transaction_limits",
				"lgpd_data_protection",
			},
		},
		{
			name:         "União Europeia - Regras GDPR",
			market:       "União Europeia",
			resourceType: "crm:customer:data",
			action:       "read",
			amount:       0,
			expectedRules: []string{
				"gdpr_consent_verification",
				"eu_data_residency",
				"eu_right_to_access",
			},
		},
		{
			name:         "China - Regras Específicas",
			market:       "China",
			resourceType: "payment_gateway:transaction",
			action:       "create",
			amount:       50000,
			expectedRules: []string{
				"china_currency_controls",
				"china_transaction_reporting",
				"china_data_localization",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mock do RiskDataProvider
			mockDataProvider := new(MockRiskDataProvider)
			
			// Configurar UserRiskProfile básico
			mockDataProvider.On("GetUserRiskProfile", mock.Anything, mock.Anything).
				Return(&risk.UserRiskProfile{
					ID:           "user:test:123",
					BaseRiskScore: 0.2,
				}, nil)
				
			// Configurar ResourceSensitivity básico
			mockDataProvider.On("GetResourceSensitivity", mock.Anything, mock.Anything).
				Return(&risk.ResourceSensitivity{
					ID:               mock.Anything,
					SensitivityLevel: risk.SensitivityMedium,
				}, nil)
				
			// Outros mocks necessários
			mockDataProvider.On("GetLocationTrustScore", mock.Anything, mock.Anything).
				Return(0.8, nil)
			mockDataProvider.On("GetIPTrustScore", mock.Anything, mock.Anything).
				Return(0.8, nil)
			mockDataProvider.On("GetDeviceTrustScore", mock.Anything, mock.Anything).
				Return(0.8, nil)
			mockDataProvider.On("IsBusinessHours", mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
			
			// Criar o avaliador de risco
			evaluator := risk.NewAdaptiveRiskEvaluator(mockDataProvider)
			
			// Criar requisição de avaliação de risco com mercado específico
			request := &risk.RiskEvaluationRequest{
				UserID:     "user:test:123",
				ResourceID: tc.resourceType,
				Action:     tc.action,
				Market:     tc.market,
				Attributes: map[string]interface{}{
					"amount": tc.amount,
				},
				IsRecognizedDevice: true,
			}
			
			// Executar avaliação de risco
			result, err := evaluator.EvaluateRisk(testCtx, request)
			
			// Verificar resultados
			require.NoError(t, err, "Avaliação de risco não deveria falhar")
			
			// Verificar se as regras específicas do mercado foram aplicadas
			appliedRules := result.AppliedRules
			for _, expectedRule := range tc.expectedRules {
				assert.Contains(t, appliedRules, expectedRule, 
					"Regra específica do mercado %s não foi aplicada: %s", tc.market, expectedRule)
			}
			
			// Verificar chamadas ao mock
			mockDataProvider.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
}

// TestRiskEvaluator_ComplianceIntegration verifica se a avaliação de risco
// integra corretamente os requisitos de conformidade regulatória
func TestRiskEvaluator_ComplianceIntegration(t *testing.T) {
	// Configurar observabilidade para o teste
	obs, err := testutil.NewTestObservability("risk_compliance_test")
	require.NoError(t, err, "Falha ao configurar observabilidade para o teste")
	defer obs.Shutdown(context.Background())
	
	// Casos de teste para verificação de conformidade
	testCases := []struct {
		name                string
		resourceType        string
		action              string
		market              string
		dataClassification  string
		expectedComplianceChecks []string
	}{
		{
			name:               "Dados PII - GDPR/LGPD",
			resourceType:       "crm:customer:personal_data",
			action:             "read",
			market:             "União Europeia",
			dataClassification: "pii",
			expectedComplianceChecks: []string{
				"gdpr_data_access",
				"data_minimization",
				"purpose_limitation",
				"consent_verification",
			},
		},
		{
			name:               "Transação Financeira - PCI DSS",
			resourceType:       "payment_gateway:card_transaction",
			action:             "create",
			market:             "Global",
			dataClassification: "payment_card",
			expectedComplianceChecks: []string{
				"pci_dss_encryption",
				"pci_dss_access_control",
				"transaction_logging",
				"sensitive_data_handling",
			},
		},
		{
			name:               "Open Banking - Open Finance",
			resourceType:       "open_banking:account:data",
			action:             "read",
			market:             "Brasil",
			dataClassification: "financial",
			expectedComplianceChecks: []string{
				"open_banking_consent",
				"api_security",
				"access_token_validation",
				"scope_verification",
				"bacen_reporting",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			testCtx := obs.RecordTestStart(ctx, tc.name)
			
			// Configurar mock do RiskDataProvider
			mockDataProvider := new(MockRiskDataProvider)
			
			// Configurar ResourceSensitivity com classificação de dados
			mockDataProvider.On("GetResourceSensitivity", mock.Anything, mock.Anything).
				Return(&risk.ResourceSensitivity{
					ID:               tc.resourceType,
					SensitivityLevel: risk.SensitivityHigh,
					Classification:   tc.dataClassification,
					ComplianceRequirements: []string{
						"data_protection",
						"audit_logging",
					},
				}, nil)
			
			// Outros mocks necessários com implementações simplificadas
			mockDataProvider.On("GetUserRiskProfile", mock.Anything, mock.Anything).
				Return(&risk.UserRiskProfile{
					ID:           "user:test:123",
					BaseRiskScore: 0.2,
				}, nil)
			mockDataProvider.On("GetLocationTrustScore", mock.Anything, mock.Anything).
				Return(0.8, nil)
			mockDataProvider.On("GetIPTrustScore", mock.Anything, mock.Anything).
				Return(0.8, nil)
			mockDataProvider.On("GetDeviceTrustScore", mock.Anything, mock.Anything).
				Return(0.8, nil)
			mockDataProvider.On("IsBusinessHours", mock.Anything, mock.Anything, mock.Anything).
				Return(true, nil)
			
			// Criar o avaliador de risco com integração de compliance
			evaluator := risk.NewAdaptiveRiskEvaluator(mockDataProvider)
			
			// Criar requisição de avaliação de risco
			request := &risk.RiskEvaluationRequest{
				UserID:     "user:test:123",
				ResourceID: tc.resourceType,
				Action:     tc.action,
				Market:     tc.market,
				Attributes: map[string]interface{}{
					"data_classification": tc.dataClassification,
					"purpose":            "authorized_business_function",
				},
			}
			
			// Executar avaliação de risco
			result, err := evaluator.EvaluateRisk(testCtx, request)
			
			// Verificar resultados
			require.NoError(t, err, "Avaliação de risco não deveria falhar")
			
			// Verificar se as verificações de conformidade foram aplicadas
			complianceChecks := result.ComplianceChecks
			for _, expectedCheck := range tc.expectedComplianceChecks {
				assert.Contains(t, complianceChecks, expectedCheck, 
					"Verificação de conformidade não foi aplicada: %s", expectedCheck)
			}
			
			// Verificar se os metadados de auditoria para conformidade estão presentes
			assert.Contains(t, result.AuditMetadata, "compliance_framework",
				"Metadados de auditoria devem incluir o framework de conformidade")
			
			// Verificar chamadas ao mock
			mockDataProvider.AssertExpectations(t)
			
			// Registrar conclusão do teste
			obs.RecordTestEnd(testCtx, tc.name, err == nil, time.Since(time.Now()))
		})
	}
}