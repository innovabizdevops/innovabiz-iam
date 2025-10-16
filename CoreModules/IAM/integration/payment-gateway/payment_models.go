package paymentgateway

import (
	"time"

	cv "github.com/innovabizdevops/innovabiz-iam/integration/cross-verification"
)

// PaymentRequest contém os dados necessários para processar um pagamento
type PaymentRequest struct {
	RequestID         string                 `json:"request_id"`
	TransactionID     string                 `json:"transaction_id,omitempty"`
	UserID            string                 `json:"user_id"`
	TenantID          string                 `json:"tenant_id"`
	RegionCode        string                 `json:"region_code"`
	Amount            float64                `json:"amount"`
	Currency          string                 `json:"currency"`
	PaymentMethod     string                 `json:"payment_method"`
	PaymentType       string                 `json:"payment_type"`
	MerchantID        string                 `json:"merchant_id"`
	MerchantCategory  string                 `json:"merchant_category"`
	Description       string                 `json:"description"`
	UserData          UserData               `json:"user_data"`
	FinancialData     FinancialData          `json:"financial_data"`
	DeviceInfo        DeviceInfo             `json:"device_info"`
	FinancialProducts []string               `json:"financial_products"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	RecurringPayment  bool                   `json:"recurring_payment"`
	InternationalPayment bool                `json:"international_payment"`
	HighRiskCategory  bool                   `json:"high_risk_category"`
	PaymentReference  string                 `json:"payment_reference,omitempty"`
	CallbackURL       string                 `json:"callback_url,omitempty"`
	Timestamp         time.Time              `json:"timestamp"`
}

// PaymentResponse contém o resultado do processamento do pagamento
type PaymentResponse struct {
	RequestID             string                 `json:"request_id"`
	TransactionID         string                 `json:"transaction_id"`
	Status                string                 `json:"status"`
	ProcessingTimeMs      int64                  `json:"processing_time_ms"`
	TrustScore            int                    `json:"trust_score,omitempty"`
	TrustLevel            string                 `json:"trust_level,omitempty"`
	StatusDescription     string                 `json:"status_description"`
	StatusCode            string                 `json:"status_code"`
	ApprovalCode          string                 `json:"approval_code,omitempty"`
	AuthorizationID       string                 `json:"authorization_id,omitempty"`
	ChallengeRequired     bool                   `json:"challenge_required"`
	ChallengeDetails      *ChallengeDetails      `json:"challenge_details,omitempty"`
	DetectedAnomalies     []cv.Anomaly           `json:"detected_anomalies,omitempty"`
	RiskLevel             string                 `json:"risk_level,omitempty"`
	VerificationDetails   VerificationDetails    `json:"verification_details,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	Timestamp             time.Time              `json:"timestamp"`
}

// UserData contém informações do usuário para verificação
type UserData struct {
	FullName          string    `json:"full_name"`
	DocumentType      string    `json:"document_type"`
	DocumentNumber    string    `json:"document_number"`
	DocumentIssueDate time.Time `json:"document_issue_date"`
	DocumentExpiry    time.Time `json:"document_expiry"`
	DateOfBirth       time.Time `json:"date_of_birth"`
	PhoneNumber       string    `json:"phone_number"`
	EmailAddress      string    `json:"email_address"`
	Address           Address   `json:"address"`
	AccountCreated    time.Time `json:"account_created"`
	LastLogin         time.Time `json:"last_login"`
	VerificationLevel string    `json:"verification_level"`
}

// Address contém detalhes de endereço
type Address struct {
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"coordinates,omitempty"`
	Verified bool `json:"verified"`
}

// FinancialData contém informações financeiras do usuário
type FinancialData struct {
	CreditScore           int                        `json:"credit_score"`
	AccountDetails        []AccountDetail            `json:"account_details"`
	FinancialProfile      FinancialProfile           `json:"financial_profile"`
	PaymentHistory        []PaymentHistoryEntry      `json:"payment_history"`
	IncomeVerification    IncomeVerification         `json:"income_verification"`
	CreditHistory         []CreditHistoryEvent       `json:"credit_history"`
	TotalTransactionsLast30Days int                  `json:"total_transactions_last_30_days"`
	ValueTransactionsLast30Days float64              `json:"value_transactions_last_30_days"`
	LastTransactionDate   time.Time                  `json:"last_transaction_date"`
}

// AccountDetail contém detalhes de uma conta financeira
type AccountDetail struct {
	AccountID       string    `json:"account_id"`
	AccountType     string    `json:"account_type"`
	AccountStatus   string    `json:"account_status"`
	Balance         float64   `json:"balance"`
	Currency        string    `json:"currency"`
	AccountAgeDays  int       `json:"account_age_days"`
	LastActivity    time.Time `json:"last_activity"`
	InstitutionID   string    `json:"institution_id"`
	InstitutionName string    `json:"institution_name"`
}

// FinancialProfile contém o perfil financeiro do usuário
type FinancialProfile struct {
	MonthlyIncome      float64                `json:"monthly_income"`
	DebtToIncomeRatio  float64                `json:"debt_to_income_ratio"`
	SpendingCategories map[string]float64     `json:"spending_categories"`
	AvgMonthlySpending float64                `json:"avg_monthly_spending"`
	CreditUtilization  float64                `json:"credit_utilization"`
}

// PaymentHistoryEntry contém uma entrada do histórico de pagamentos
type PaymentHistoryEntry struct {
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Date          time.Time `json:"date"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	Category      string    `json:"category"`
}

// IncomeVerification contém dados de verificação de renda
type IncomeVerification struct {
	Verified          bool    `json:"verified"`
	DeclaredIncome    float64 `json:"declared_income"`
	VerifiedIncome    float64 `json:"verified_income"`
	VerificationDate  time.Time `json:"verification_date"`
	VerificationMethod string  `json:"verification_method"`
}

// CreditHistoryEvent contém um evento do histórico de crédito
type CreditHistoryEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	EventDate   time.Time `json:"event_date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount,omitempty"`
	Provider    string    `json:"provider,omitempty"`
}

// DeviceInfo contém informações sobre o dispositivo usado
type DeviceInfo struct {
	DeviceID           string                 `json:"device_id"`
	DeviceType         string                 `json:"device_type"`
	DeviceModel        string                 `json:"device_model"`
	OperatingSystem    string                 `json:"operating_system"`
	OSVersion          string                 `json:"os_version"`
	IPAddress          string                 `json:"ip_address"`
	LocationInfo       LocationInfo           `json:"location_info"`
	Browser            string                 `json:"browser,omitempty"`
	BrowserVersion     string                 `json:"browser_version,omitempty"`
	UserAgent          string                 `json:"user_agent,omitempty"`
	KnownDevice        bool                   `json:"known_device"`
	DeviceFingerprint  string                 `json:"device_fingerprint"`
	FirstSeen          time.Time              `json:"first_seen"`
	LastSeen           time.Time              `json:"last_seen"`
	DeviceAttributes   map[string]interface{} `json:"device_attributes,omitempty"`
}

// LocationInfo contém informações de localização
type LocationInfo struct {
	Country        string  `json:"country"`
	City           string  `json:"city"`
	Region         string  `json:"region"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	AccuracyRadius int     `json:"accuracy_radius"`
	ISP            string  `json:"isp"`
	TimeZone       string  `json:"time_zone"`
	ASN            string  `json:"asn"`
}

// LimitCheckResult contém o resultado da verificação de limites
type LimitCheckResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
	Details string `json:"details,omitempty"`
}

// ChallengeDetails contém detalhes de um desafio para o usuário
type ChallengeDetails struct {
	ChallengeID       string                 `json:"challenge_id"`
	ChallengeType     string                 `json:"challenge_type"`
	ChallengeMethod   string                 `json:"challenge_method"`
	Instructions      string                 `json:"instructions"`
	ExpirationTime    time.Time              `json:"expiration_time"`
	RetryCount        int                    `json:"retry_count"`
	MaxRetries        int                    `json:"max_retries"`
	ChallengeMetadata map[string]interface{} `json:"challenge_metadata,omitempty"`
	VerificationURL   string                 `json:"verification_url,omitempty"`
}

// VerificationDetails contém detalhes do processo de verificação
type VerificationDetails struct {
	VerificationID      string    `json:"verification_id"`
	VerificationLevel   string    `json:"verification_level"`
	VerifiedFields      []string  `json:"verified_fields"`
	FailedFields        []string  `json:"failed_fields"`
	VerificationTime    time.Time `json:"verification_time"`
	TrustAssessmentID   string    `json:"trust_assessment_id,omitempty"`
	EnhancedVerification bool      `json:"enhanced_verification"`
	RelevantFactors     []string  `json:"relevant_factors,omitempty"`
}

// ChallengeManagerConfig contém configurações para o gerenciador de desafios
type ChallengeManagerConfig struct {
	DefaultExpirationMinutes  int                          `json:"default_expiration_minutes"`
	MaxRetries                int                          `json:"max_retries"`
	EnableChallengeProgression bool                        `json:"enable_challenge_progression"`
	ChallengeMethodMap        map[string][]string         `json:"challenge_method_map"`
}

// ChallengeManager gerencia desafios para transações
type ChallengeManager struct {
	config            ChallengeManagerConfig
	activeChallenges  sync.Map
}

// NewChallengeManager cria uma nova instância do gerenciador de desafios
func NewChallengeManager(config ChallengeManagerConfig) *ChallengeManager {
	// Configurar valores padrão
	if config.DefaultExpirationMinutes == 0 {
		config.DefaultExpirationMinutes = 15
	}
	
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	
	return &ChallengeManager{
		config: config,
	}
}

// Funções auxiliares para mapeamento de dados

// mapToIdentityData converte UserData para IdentityData do cross-verification
func mapToIdentityData(userData UserData) cv.IdentityData {
	return cv.IdentityData{
		FullName:        userData.FullName,
		DocumentType:    userData.DocumentType,
		DocumentNumber:  userData.DocumentNumber,
		DateOfBirth:     userData.DateOfBirth,
		Address: cv.AddressData{
			Street:     userData.Address.Street,
			City:       userData.Address.City,
			State:      userData.Address.State,
			Country:    userData.Address.Country,
			PostalCode: userData.Address.PostalCode,
		},
		ContactInfo: cv.ContactInfo{
			PhoneNumber:  userData.PhoneNumber,
			EmailAddress: userData.EmailAddress,
		},
	}
}

// mapToFinancialData converte FinancialData para FinancialData do cross-verification
func mapToFinancialData(data FinancialData) cv.FinancialData {
	cvFinancialData := cv.FinancialData{
		CreditScore: data.CreditScore,
	}
	
	// Mapear outros campos conforme necessário...
	// Este é um mapeamento simplificado, na implementação real seria completo
	
	return cvFinancialData
}

// mapToDeviceData converte DeviceInfo para DeviceData do cross-verification
func mapToDeviceData(deviceInfo DeviceInfo) cv.DeviceData {
	return cv.DeviceData{
		DeviceID:        deviceInfo.DeviceID,
		DeviceType:      deviceInfo.DeviceType,
		IPAddress:       deviceInfo.IPAddress,
		OperatingSystem: deviceInfo.OperatingSystem,
		KnownDevice:     deviceInfo.KnownDevice,
		LocationInfo: cv.GeoLocation{
			Country:   deviceInfo.LocationInfo.Country,
			City:      deviceInfo.LocationInfo.City,
			Latitude:  deviceInfo.LocationInfo.Latitude,
			Longitude: deviceInfo.LocationInfo.Longitude,
		},
	}
}