// ==============================================================================
// Nome: ecommerce_connector.go
// Descrição: Conector de integração entre IAM e módulo E-Commerce
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

package ecommerce

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/innovabiz/iam/common/errors"
	"github.com/innovabiz/iam/common/logging"
	"github.com/innovabiz/iam/common/metrics"
	"github.com/innovabiz/iam/common/security"
	"github.com/innovabiz/iam/common/tracing"
	"github.com/innovabiz/iam/integration"
	"github.com/innovabiz/iam/models"

	"github.com/innovabiz/datacore/client"
	ec "github.com/innovabiz/ecommerce/client"
)

// Constantes para o conector
const (
	ModuleName         = "e-commerce"
	IntegrationType    = "ecommerce"
	DefaultTokenExpiry = 24 * time.Hour
)

// ECommerceConnector implementa a integração entre IAM e E-Commerce
type ECommerceConnector struct {
	config      *ECommerceConfig
	ecClient    ec.Client
	dataClient  client.DataCoreClient
	logger      logging.Logger
	metrics     metrics.MetricsCollector
	tracer      tracing.Tracer
	tokenService security.TokenService
}

// ECommerceConfig contém configurações para conexão com E-Commerce
type ECommerceConfig struct {
	BaseURL             string
	APIVersion          string
	Timeout             time.Duration
	MaxRetries          int
	EnableCache         bool
	CacheExpiration     time.Duration
	IntegrationEndpoint string
	ClientID            string
	ClientSecret        string
	TokenEndpoint       string
}

// NewECommerceConnector cria uma nova instância do conector de E-Commerce
func NewECommerceConnector(
	config *ECommerceConfig,
	ecClient ec.Client,
	dataClient client.DataCoreClient,
	logger logging.Logger,
	metrics metrics.MetricsCollector,
	tracer tracing.Tracer,
	tokenService security.TokenService,
) *ECommerceConnector {
	return &ECommerceConnector{
		config:      config,
		ecClient:    ecClient,
		dataClient:  dataClient,
		logger:      logger.WithField("module", ModuleName),
		metrics:     metrics,
		tracer:      tracer,
		tokenService: tokenService,
	}
}

// VincularUsuario cria/atualiza vínculo de usuário IAM com E-Commerce
func (c *ECommerceConnector) VincularUsuario(
	ctx context.Context,
	usuarioID string,
	tenantID string,
	perfilTipo string,
	nivelAcesso string,
	detalhes map[string]interface{},
) (*models.IntegrationIdentity, error) {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.VincularUsuario")
	defer span.End()

	c.metrics.CountEvent("ecommerce_vincular_usuario_attempt")

	// Validar dados de entrada
	if usuarioID == "" || tenantID == "" || perfilTipo == "" {
		return nil, errors.NewValidationError("usuarioID, tenantID e perfilTipo são obrigatórios")
	}

	// Verificar se usuário existe no IAM
	usuario, err := c.dataClient.GetUsuario(ctx, usuarioID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar usuário no IAM")
		c.metrics.CountEvent("ecommerce_vincular_usuario_error")
		return nil, fmt.Errorf("falha ao buscar usuário: %w", err)
	}

	// Preparar dados para envio ao E-Commerce
	ecUsuarioRequest := ec.UsuarioRequest{
		UsuarioID:    usuarioID,
		TenantID:     tenantID,
		Email:        usuario.Email,
		Nome:         usuario.NomeCompleto,
		Documento:    usuario.DocumentoPrincipal,
		TipoDocumento: usuario.TipoDocumento,
		Celular:      usuario.Celular,
		PerfilTipo:   perfilTipo,
		NivelAcesso:  nivelAcesso,
		Detalhes:     detalhes,
	}

	// Registrar usuário no E-Commerce
	ecUsuario, err := c.ecClient.RegisterUsuario(ctx, ecUsuarioRequest)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao registrar usuário no E-Commerce")
		c.metrics.CountEvent("ecommerce_vincular_usuario_error")
		return nil, fmt.Errorf("falha ao registrar usuário no E-Commerce: %w", err)
	}

	// Armazenar o vínculo no IAM via DataCore
	detalhesJSON, _ := json.Marshal(detalhes)
	identity := &models.IntegrationIdentity{
		UsuarioID:      usuarioID,
		TenantID:       tenantID,
		IntegrationType: IntegrationType,
		ExternalID:     ecUsuario.ID,
		ExternalTenantID: ecUsuario.TenantID,
		ProfileType:    perfilTipo,
		AccessLevel:    nivelAcesso,
		Status:         "ativo",
		Details:        string(detalhesJSON),
	}

	// Persistir no banco de dados via DataCore
	err = c.dataClient.CreateIntegrationIdentity(ctx, identity)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao persistir vínculo IAM-ECommerce")
		c.metrics.CountEvent("ecommerce_vincular_usuario_error")
		return nil, fmt.Errorf("falha ao persistir vínculo: %w", err)
	}

	c.metrics.CountEvent("ecommerce_vincular_usuario_success")
	return identity, nil
}

// GerarToken gera um token de acesso para o E-Commerce
func (c *ECommerceConnector) GerarToken(
	ctx context.Context,
	identityID string,
	escopo string,
	dispositivoID string,
	duracaoMinutos int,
) (*models.IntegrationToken, error) {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.GerarToken")
	defer span.End()

	c.metrics.CountEvent("ecommerce_gerar_token_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("ecommerce_gerar_token_error")
		return nil, fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	if identity.Status != "ativo" {
		c.metrics.CountEvent("ecommerce_gerar_token_error")
		return nil, errors.NewAuthorizationError("identidade de integração não está ativa")
	}

	// Definir expiração padrão se não especificada
	if duracaoMinutos <= 0 {
		duracaoMinutos = 60 // 1 hora padrão
	}

	// Preparar e assinar token JWT
	claims := map[string]interface{}{
		"sub":            identity.UsuarioID,
		"tenant_id":      identity.TenantID,
		"ec_id":          identity.ExternalID,
		"ec_tenant_id":   identity.ExternalTenantID,
		"profile_type":   identity.ProfileType,
		"access_level":   identity.AccessLevel,
		"scope":          escopo,
		"device_id":      dispositivoID,
	}

	// Gerar token JWT com assinatura
	expiry := time.Duration(duracaoMinutos) * time.Minute
	token, err := c.tokenService.GenerateToken(ctx, claims, expiry)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao gerar token JWT")
		c.metrics.CountEvent("ecommerce_gerar_token_error")
		return nil, fmt.Errorf("falha ao gerar token: %w", err)
	}

	// Gerar refresh token
	refreshClaims := map[string]interface{}{
		"token_type": "refresh",
		"token_id":   token.ID,
		"user_id":    identity.UsuarioID,
	}
	refreshToken, err := c.tokenService.GenerateToken(ctx, refreshClaims, 30*24*time.Hour) // 30 dias
	if err != nil {
		c.logger.WithError(err).Error("Falha ao gerar refresh token")
		c.metrics.CountEvent("ecommerce_gerar_token_error")
		return nil, fmt.Errorf("falha ao gerar refresh token: %w", err)
	}

	// Criar registro do token na base de dados
	integrationToken := &models.IntegrationToken{
		IdentityID:     identityID,
		TokenHash:      token.Hash,
		RefreshTokenHash: refreshToken.Hash,
		Escopo:         escopo,
		DataExpiracao:  time.Now().Add(expiry),
		DataCriacao:    time.Now(),
		DispositivoID:  dispositivoID,
	}

	// Persistir token
	err = c.dataClient.CreateIntegrationToken(ctx, integrationToken)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao persistir token de integração")
		c.metrics.CountEvent("ecommerce_gerar_token_error")
		return nil, fmt.Errorf("falha ao persistir token: %w", err)
	}

	// Adicionar tokens à resposta
	integrationToken.Token = token.Token
	integrationToken.RefreshToken = refreshToken.Token

	c.metrics.CountEvent("ecommerce_gerar_token_success")
	return integrationToken, nil
}

// SincronizarEndereco sincroniza um endereço do usuário com o E-Commerce
func (c *ECommerceConnector) SincronizarEndereco(
	ctx context.Context,
	identityID string,
	endereco *models.Endereco,
) error {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.SincronizarEndereco")
	defer span.End()

	c.metrics.CountEvent("ecommerce_sincronizar_endereco_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("ecommerce_sincronizar_endereco_error")
		return fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	// Converter para o formato do E-Commerce
	ecEndereco := ec.EnderecoRequest{
		UsuarioID:    identity.ExternalID,
		TenantID:     identity.ExternalTenantID,
		Tipo:         endereco.Tipo,
		Nome:         endereco.Nome,
		Logradouro:   endereco.Logradouro,
		Numero:       endereco.Numero,
		Complemento:  endereco.Complemento,
		Bairro:       endereco.Bairro,
		Cidade:       endereco.Cidade,
		Estado:       endereco.Estado,
		Pais:         endereco.Pais,
		CEP:          endereco.CEP,
		Principal:    endereco.Principal,
	}

	// Enviar para o E-Commerce
	resultado, err := c.ecClient.RegisterEndereco(ctx, ecEndereco)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao registrar endereço no E-Commerce")
		c.metrics.CountEvent("ecommerce_sincronizar_endereco_error")
		return fmt.Errorf("falha ao registrar endereço no E-Commerce: %w", err)
	}

	// Atualizar o ID externo do endereço
	endereco.ExternalID = resultado.ID
	endereco.LastSynced = time.Now()

	// Atualizar no banco de dados
	err = c.dataClient.UpdateEndereco(ctx, endereco)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao atualizar endereço no IAM")
		c.metrics.CountEvent("ecommerce_sincronizar_endereco_error")
		return fmt.Errorf("falha ao atualizar endereço: %w", err)
	}

	c.metrics.CountEvent("ecommerce_sincronizar_endereco_success")
	return nil
}

// AdicionarConsentimento registra um consentimento do usuário para E-Commerce
func (c *ECommerceConnector) AdicionarConsentimento(
	ctx context.Context,
	identityID string,
	tipoConsentimento string,
	descricao string,
	ipOrigem string,
	userAgent string,
	diasValidade int,
	documentoURL string,
) error {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.AdicionarConsentimento")
	defer span.End()

	c.metrics.CountEvent("ecommerce_adicionar_consentimento_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("ecommerce_adicionar_consentimento_error")
		return fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	// Registrar consentimento
	consentimento := &models.Consentimento{
		UsuarioID:        identity.UsuarioID,
		TenantID:         identity.TenantID,
		IdentityID:       identityID,
		TipoConsentimento: tipoConsentimento,
		Descricao:        descricao,
		IPOrigem:         ipOrigem,
		UserAgent:        userAgent,
		DataConsentimento: time.Now(),
		DataExpiracao:    time.Now().AddDate(0, 0, diasValidade),
		DocumentoURL:     documentoURL,
	}

	// Calcular hash do documento de consentimento se disponível
	if documentoURL != "" {
		hash, err := c.calcularHashDocumento(ctx, documentoURL)
		if err != nil {
			c.logger.WithError(err).Warn("Falha ao calcular hash do documento de consentimento")
			// Não falhar a operação, apenas registrar o aviso
		} else {
			consentimento.HashDocumento = hash
		}
	}

	// Persistir consentimento
	err = c.dataClient.CreateConsentimento(ctx, consentimento)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao persistir consentimento")
		c.metrics.CountEvent("ecommerce_adicionar_consentimento_error")
		return fmt.Errorf("falha ao persistir consentimento: %w", err)
	}

	// Registrar no E-Commerce
	ecConsentimento := ec.ConsentimentoRequest{
		UsuarioID:        identity.ExternalID,
		TenantID:         identity.ExternalTenantID,
		TipoConsentimento: tipoConsentimento,
		DataConsentimento: consentimento.DataConsentimento,
		DataExpiracao:    consentimento.DataExpiracao,
		IPOrigem:         ipOrigem,
	}

	_, err = c.ecClient.RegisterConsentimento(ctx, ecConsentimento)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao registrar consentimento no E-Commerce")
		// Não falhar completamente, já que o consentimento está registrado no IAM
		c.metrics.CountEvent("ecommerce_adicionar_consentimento_partial")
	} else {
		c.metrics.CountEvent("ecommerce_adicionar_consentimento_success")
	}

	return nil
}

// calcularHashDocumento calcula o hash SHA-256 de um documento a partir da URL
func (c *ECommerceConnector) calcularHashDocumento(ctx context.Context, documentoURL string) (string, error) {
	// Implementação omitida para brevidade
	// Deveria baixar o documento da URL e calcular o hash
	return "", nil
}

// ValidarToken valida um token de integração E-Commerce
func (c *ECommerceConnector) ValidarToken(
	ctx context.Context,
	token string,
) (*models.IntegrationIdentity, error) {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.ValidarToken")
	defer span.End()

	c.metrics.CountEvent("ecommerce_validar_token_attempt")

	// Validar token JWT
	claims, err := c.tokenService.ValidateToken(ctx, token)
	if err != nil {
		c.logger.WithError(err).Error("Token inválido")
		c.metrics.CountEvent("ecommerce_validar_token_error")
		return nil, errors.NewAuthenticationError("token inválido")
	}

	// Extrair identidade a partir das claims
	usuarioID, ok := claims["sub"].(string)
	if !ok {
		c.metrics.CountEvent("ecommerce_validar_token_error")
		return nil, errors.NewAuthenticationError("token malformado: usuário não identificado")
	}

	// Verificar se token expirou
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			c.metrics.CountEvent("ecommerce_validar_token_error")
			return nil, errors.NewAuthenticationError("token expirado")
		}
	}

	// Buscar identidade no banco de dados
	integrationID := claims["token_id"].(string)
	identity, err := c.dataClient.GetIntegrationIdentityByToken(ctx, integrationID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade pelo token")
		c.metrics.CountEvent("ecommerce_validar_token_error")
		return nil, fmt.Errorf("falha ao validar token: %w", err)
	}

	// Atualizar último uso do token
	err = c.dataClient.UpdateTokenLastUse(ctx, integrationID)
	if err != nil {
		c.logger.WithError(err).Warn("Falha ao atualizar último uso do token")
		// Não falhar a operação, apenas registrar o aviso
	}

	c.metrics.CountEvent("ecommerce_validar_token_success")
	return identity, nil
}

// ListarPreferencias lista as preferências do usuário no E-Commerce
func (c *ECommerceConnector) ListarPreferencias(
	ctx context.Context,
	identityID string,
) (map[string]interface{}, error) {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.ListarPreferencias")
	defer span.End()

	c.metrics.CountEvent("ecommerce_listar_preferencias_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("ecommerce_listar_preferencias_error")
		return nil, fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	// Buscar preferências no E-Commerce
	prefs, err := c.ecClient.GetPreferencias(ctx, identity.ExternalID, identity.ExternalTenantID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar preferências no E-Commerce")
		c.metrics.CountEvent("ecommerce_listar_preferencias_error")
		return nil, fmt.Errorf("falha ao buscar preferências: %w", err)
	}

	c.metrics.CountEvent("ecommerce_listar_preferencias_success")
	return prefs, nil
}

// AtualizarPreferencias atualiza as preferências do usuário no E-Commerce
func (c *ECommerceConnector) AtualizarPreferencias(
	ctx context.Context,
	identityID string,
	preferencias map[string]interface{},
) error {
	ctx, span := c.tracer.StartSpan(ctx, "ECommerceConnector.AtualizarPreferencias")
	defer span.End()

	c.metrics.CountEvent("ecommerce_atualizar_preferencias_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("ecommerce_atualizar_preferencias_error")
		return fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	// Atualizar preferências no E-Commerce
	err = c.ecClient.UpdatePreferencias(ctx, identity.ExternalID, identity.ExternalTenantID, preferencias)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao atualizar preferências no E-Commerce")
		c.metrics.CountEvent("ecommerce_atualizar_preferencias_error")
		return fmt.Errorf("falha ao atualizar preferências: %w", err)
	}

	c.metrics.CountEvent("ecommerce_atualizar_preferencias_success")
	return nil
}