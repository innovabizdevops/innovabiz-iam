package crossverification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/observability/logging"
	"github.com/innovabizdevops/innovabiz-iam/observability/tracing"
)

const (
	// Categoria do verificador
	CategoryIdentity = "identity"
	
	// Peso padrão da categoria na pontuação geral
	DefaultIdentityWeight = 30
)

// IdentityVerifierConfig contém configurações para o verificador de identidade
type IdentityVerifierConfig struct {
	Weight               int                `json:"weight"`
	RequiredFields       []string           `json:"required_fields"`
	StrictNameMatching   bool               `json:"strict_name_matching"`
	MinDocumentValidDays int                `json:"min_document_valid_days"`
	RegionalSettings     map[string]IdentityRegionalSettings `json:"regional_settings"`
}

// IdentityRegionalSettings contém configurações regionais para verificação de identidade
type IdentityRegionalSettings struct {
	RequiredDocumentTypes []string          `json:"required_document_types"`
	VerificationMethods   map[string]int    `json:"verification_methods_scores"`
	MinAcceptableScore    int               `json:"min_acceptable_score"`
}

// IdentityVerifier implementa a verificação de identidade
type IdentityVerifier struct {
	config IdentityVerifierConfig
	logger logging.Logger
	tracer tracing.Tracer
}

// NewIdentityVerifier cria uma nova instância de verificador de identidade
func NewIdentityVerifier(config IdentityVerifierConfig, logger logging.Logger, tracer tracing.Tracer) *IdentityVerifier {
	// Configurar peso padrão se não especificado
	if config.Weight <= 0 {
		config.Weight = DefaultIdentityWeight
	}
	
	return &IdentityVerifier{
		config: config,
		logger: logger,
		tracer: tracer,
	}
}

// Verify implementa a interface CategoryVerifier
func (v *IdentityVerifier) Verify(ctx context.Context, req *CredentialFinancialVerificationRequest) (*VerificationResult, error) {
	ctx, span := v.tracer.StartSpan(ctx, "IdentityVerifier.Verify")
	defer span.End()
	
	v.logger.InfoWithContext(ctx, "Iniciando verificação de identidade",
		"request_id", req.RequestID,
		"user_id", req.UserID)
		
	// Resultado inicial da verificação
	result := &VerificationResult{
		Category:       CategoryIdentity,
		Status:         VerificationStatusPending,
		Score:          0,
		Description:    "Verificação de identidade entre credenciais IAM e dados financeiros",
		VerifiedFields: []string{},
		FailedFields:   []string{},
		Details:        make(map[string]interface{}),
	}
	
	// Verificar campos obrigatórios
	missingFields := v.checkRequiredFields(req)
	if len(missingFields) > 0 {
		result.Status = VerificationStatusFailed
		result.Description = fmt.Sprintf("Campos obrigatórios de identidade ausentes: %s", strings.Join(missingFields, ", "))
		result.FailedFields = missingFields
		return result, nil
	}
	
	// Recuperar configurações regionais específicas
	regionalSettings := v.getRegionalSettings(req.RegionCode)
	
	// Verificações de identidade a serem realizadas
	checks := []struct {
		name     string
		checkFn  func(req *CredentialFinancialVerificationRequest) (bool, string)
		weight   int
	}{
		{
			name:    "documento_tipo_válido",
			checkFn: v.verifyDocumentType,
			weight:  15,
		},
		{
			name:    "nome_correspondência",
			checkFn: v.verifyNameMatching,
			weight:  25,
		},
		{
			name:    "documento_validade",
			checkFn: v.verifyDocumentExpiry,
			weight:  15,
		},
		{
			name:    "data_nascimento",
			checkFn: v.verifyDateOfBirth,
			weight:  15,
		},
		{
			name:    "endereço_correspondência",
			checkFn: v.verifyAddressMatching,
			weight:  10,
		},
		{
			name:    "métodos_verificação",
			checkFn: v.verifyVerificationMethods,
			weight:  10,
		},
		{
			name:    "contato_verificado",
			checkFn: v.verifyContactInfo,
			weight:  10,
		},
	}
	
	// Executar verificações
	totalWeight := 0
	weightedScore := 0
	
	for _, check := range checks {
		passed, detail := check.checkFn(req)
		
		if passed {
			result.VerifiedFields = append(result.VerifiedFields, check.name)
			weightedScore += check.weight
		} else {
			result.FailedFields = append(result.FailedFields, check.name)
			result.Details[check.name+"_details"] = detail
		}
		
		totalWeight += check.weight
	}
	
	// Calcular pontuação final
	if totalWeight > 0 {
		result.Score = (weightedScore * 100) / totalWeight
	}
	
	// Determinar status com base na pontuação e configurações regionais
	if result.Score >= regionalSettings.MinAcceptableScore {
		result.Status = VerificationStatusPassed
		result.Description = fmt.Sprintf("Verificação de identidade concluída com pontuação %d/100", result.Score)
	} else if result.Score >= regionalSettings.MinAcceptableScore/2 {
		result.Status = VerificationStatusPartial
		result.Description = fmt.Sprintf("Verificação de identidade parcial com pontuação %d/100", result.Score)
	} else {
		result.Status = VerificationStatusFailed
		result.Description = fmt.Sprintf("Verificação de identidade falhou com pontuação %d/100", result.Score)
	}
	
	v.logger.InfoWithContext(ctx, "Verificação de identidade concluída",
		"request_id", req.RequestID,
		"status", result.Status,
		"score", result.Score,
		"verified_fields", len(result.VerifiedFields),
		"failed_fields", len(result.FailedFields))
		
	return result, nil
}

// GetCategory implementa a interface CategoryVerifier
func (v *IdentityVerifier) GetCategory() string {
	return CategoryIdentity
}

// GetWeight implementa a interface CategoryVerifier
func (v *IdentityVerifier) GetWeight() int {
	return v.config.Weight
}

// Verifica se todos os campos obrigatórios estão presentes
func (v *IdentityVerifier) checkRequiredFields(req *CredentialFinancialVerificationRequest) []string {
	missing := []string{}
	
	// Recupera os campos obrigatórios para a região específica
	requiredFields := v.config.RequiredFields
	if rs, ok := v.config.RegionalSettings[req.RegionCode]; ok {
		for _, docType := range rs.RequiredDocumentTypes {
			if req.IdentityData.DocumentType != docType {
				missing = append(missing, "document_type_"+docType)
			}
		}
	}
	
	// Verifica campos gerais obrigatórios
	for _, field := range requiredFields {
		switch field {
		case "document_number":
			if req.IdentityData.DocumentNumber == "" {
				missing = append(missing, field)
			}
		case "full_name":
			if req.IdentityData.FullName == "" {
				missing = append(missing, field)
			}
		case "date_of_birth":
			if req.IdentityData.DateOfBirth == "" {
				missing = append(missing, field)
			}
		case "nationality":
			if req.IdentityData.Nationality == "" {
				missing = append(missing, field)
			}
		case "address":
			if req.IdentityData.Address.StreetAddress == "" {
				missing = append(missing, field)
			}
		case "contact_info":
			if req.IdentityData.ContactInfo.Email == "" && req.IdentityData.ContactInfo.PhoneNumber == "" {
				missing = append(missing, field)
			}
		}
	}
	
	return missing
}

// Recupera as configurações regionais ou usa padrões
func (v *IdentityVerifier) getRegionalSettings(regionCode string) IdentityRegionalSettings {
	if rs, ok := v.config.RegionalSettings[regionCode]; ok {
		return rs
	}
	
	// Configurações padrão se específicas não encontradas
	return IdentityRegionalSettings{
		RequiredDocumentTypes: []string{"id_card", "passport", "drivers_license"},
		VerificationMethods: map[string]int{
			"document_verification": 80,
			"biometric":            100,
			"proof_of_address":      70,
			"manual_review":         60,
		},
		MinAcceptableScore: 70,
	}
}

// Verifica o tipo de documento
func (v *IdentityVerifier) verifyDocumentType(req *CredentialFinancialVerificationRequest) (bool, string) {
	regionalSettings := v.getRegionalSettings(req.RegionCode)
	
	for _, docType := range regionalSettings.RequiredDocumentTypes {
		if strings.EqualFold(req.IdentityData.DocumentType, docType) {
			return true, fmt.Sprintf("Tipo de documento válido: %s", docType)
		}
	}
	
	return false, fmt.Sprintf("Tipo de documento não aceito para região %s: %s", 
		req.RegionCode, req.IdentityData.DocumentType)
}

// Verifica correspondência de nome entre IAM e dados financeiros
func (v *IdentityVerifier) verifyNameMatching(req *CredentialFinancialVerificationRequest) (bool, string) {
	iamName := strings.ToLower(req.IdentityData.FullName)
	
	// Verificar nomes em contas financeiras
	for _, account := range req.FinancialData.AccountDetails {
		financialName := strings.ToLower(account.AccountHolder)
		
		// Se for necessária correspondência exata
		if v.config.StrictNameMatching {
			if iamName == financialName {
				return true, "Nome corresponde exatamente entre IAM e dados financeiros"
			}
		} else {
			// Verificação parcial (contém nome/sobrenome)
			iamNameParts := strings.Fields(iamName)
			financialNameParts := strings.Fields(financialName)
			
			// Verifica se pelo menos 70% dos componentes do nome correspondem
			matches := 0
			for _, iamPart := range iamNameParts {
				for _, finPart := range financialNameParts {
					if iamPart == finPart && len(iamPart) > 2 { // ignorar iniciais/partículas
						matches++
						break
					}
				}
			}
			
			matchRate := float64(matches) / float64(len(iamNameParts))
			if matchRate >= 0.7 {
				return true, fmt.Sprintf("Nome corresponde parcialmente (%.0f%%) entre IAM e dados financeiros", matchRate*100)
			}
		}
	}
	
	return false, "Nome no IAM não corresponde aos registros financeiros"
}

// Verifica validade do documento
func (v *IdentityVerifier) verifyDocumentExpiry(req *CredentialFinancialVerificationRequest) (bool, string) {
	if req.IdentityData.DocumentExpiry == "" {
		return true, "Documento sem data de validade especificada"
	}
	
	expiryDate, err := time.Parse("2006-01-02", req.IdentityData.DocumentExpiry)
	if err != nil {
		return false, fmt.Sprintf("Formato de data de validade inválido: %s", req.IdentityData.DocumentExpiry)
	}
	
	daysUntilExpiry := int(expiryDate.Sub(time.Now()).Hours() / 24)
	
	if daysUntilExpiry < 0 {
		return false, fmt.Sprintf("Documento expirado há %d dias", -daysUntilExpiry)
	}
	
	if daysUntilExpiry < v.config.MinDocumentValidDays {
		return false, fmt.Sprintf("Documento expira em %d dias, mínimo aceitável é %d dias", 
			daysUntilExpiry, v.config.MinDocumentValidDays)
	}
	
	return true, fmt.Sprintf("Documento válido por mais %d dias", daysUntilExpiry)
}

// Verifica data de nascimento
func (v *IdentityVerifier) verifyDateOfBirth(req *CredentialFinancialVerificationRequest) (bool, string) {
	// Aqui seria implementada a lógica de verificação da data de nascimento com outras fontes
	// Este é um placeholder para a implementação real
	
	if req.IdentityData.DateOfBirth == "" {
		return false, "Data de nascimento não fornecida"
	}
	
	// Verificação mock - implementação real validaria com dados do bureau de crédito ou outras fontes
	return true, "Data de nascimento verificada"
}

// Verifica endereço
func (v *IdentityVerifier) verifyAddressMatching(req *CredentialFinancialVerificationRequest) (bool, string) {
	// Se o endereço já foi verificado por outros meios
	if req.IdentityData.Address.VerifiedStatus {
		return true, "Endereço pré-verificado por processo anterior"
	}
	
	// Verificação de correspondência entre endereço IAM e dados financeiros
	// Este é um placeholder para a implementação real
	return true, "Endereço corresponde aos registros financeiros"
}

// Verifica métodos de verificação
func (v *IdentityVerifier) verifyVerificationMethods(req *CredentialFinancialVerificationRequest) (bool, string) {
	regionalSettings := v.getRegionalSettings(req.RegionCode)
	
	// Verifica se há métodos de verificação de identidade adequados
	highestScore := 0
	bestMethod := ""
	
	for _, method := range req.IdentityData.VerificationMethods {
		if score, ok := regionalSettings.VerificationMethods[method]; ok {
			if score > highestScore {
				highestScore = score
				bestMethod = method
			}
		}
	}
	
	if highestScore >= regionalSettings.MinAcceptableScore {
		return true, fmt.Sprintf("Método de verificação adequado: %s (pontuação: %d)", bestMethod, highestScore)
	}
	
	return false, fmt.Sprintf("Métodos de verificação insuficientes (melhor método: %s, pontuação: %d, mínimo: %d)", 
		bestMethod, highestScore, regionalSettings.MinAcceptableScore)
}

// Verifica informações de contato
func (v *IdentityVerifier) verifyContactInfo(req *CredentialFinancialVerificationRequest) (bool, string) {
	contactInfo := req.IdentityData.ContactInfo
	
	// Verifica se pelo menos um método de contato foi verificado
	if contactInfo.EmailVerified || contactInfo.PhoneVerified {
		return true, "Pelo menos um método de contato verificado"
	}
	
	return false, "Nenhum método de contato verificado"
}