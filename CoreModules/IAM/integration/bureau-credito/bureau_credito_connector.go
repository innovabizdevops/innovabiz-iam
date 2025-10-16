// ==============================================================================
// Nome: bureau_credito_connector.go
// Descrição: Conector de integração entre IAM e módulo Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

package bureaupkg

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
	bc "github.com/innovabiz/bureaucredito/client"
)

// Constantes para o conector
const (
	ModuleName         = "bureau-credito"
	IntegrationType    = "bureau_credito"
	DefaultTokenExpiry = 30 * time.Minute
)

// BureauCreditoConnector implementa a integração entre IAM e Bureau de Créditos
type BureauCreditoConnector struct {
	config      *BureauCreditoConfig
	bcClient    bc.Client
	dataClient  client.DataCoreClient
	logger      logging.Logger
	metrics     metrics.MetricsCollector
	tracer      tracing.Tracer
	tokenService security.TokenService
}

// BureauCreditoConfig contém configurações para conexão com Bureau de Créditos
type BureauCreditoConfig struct {
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
	AuditLevel          string // Nível de auditoria para operações de Bureau de Créditos
	CertificatePath     string // Caminho para certificado SSL para conexão segura
	PrivateKeyPath      string // Caminho para chave privada
}

// NewBureauCreditoConnector cria uma nova instância do conector de Bureau de Créditos
func NewBureauCreditoConnector(
	config *BureauCreditoConfig,
	bcClient bc.Client,
	dataClient client.DataCoreClient,
	logger logging.Logger,
	metrics metrics.MetricsCollector,
	tracer tracing.Tracer,
	tokenService security.TokenService,
) *BureauCreditoConnector {
	return &BureauCreditoConnector{
		config:      config,
		bcClient:    bcClient,
		dataClient:  dataClient,
		logger:      logger.WithField("module", ModuleName),
		metrics:     metrics,
		tracer:      tracer,
		tokenService: tokenService,
	}
}

// VincularUsuario cria/atualiza vínculo de usuário IAM com Bureau de Créditos
func (c *BureauCreditoConnector) VincularUsuario(
	ctx context.Context,
	usuarioID string,
	tenantID string,
	tipoVinculo string,
	nivelAcesso string,
	detalhesAutorizacao map[string]interface{},
) (*models.IntegrationIdentity, error) {
	ctx, span := c.tracer.StartSpan(ctx, "BureauCreditoConnector.VincularUsuario")
	defer span.End()

	c.metrics.CountEvent("bureau_credito_vincular_usuario_attempt")

	// Validar dados de entrada
	if usuarioID == "" || tenantID == "" || tipoVinculo == "" {
		return nil, errors.NewValidationError("usuarioID, tenantID e tipoVinculo são obrigatórios")
	}

	// Verificar se usuário existe no IAM
	usuario, err := c.dataClient.GetUsuario(ctx, usuarioID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar usuário no IAM")
		c.metrics.CountEvent("bureau_credito_vincular_usuario_error")
		return nil, fmt.Errorf("falha ao buscar usuário: %w", err)
	}

	// Preparar dados para envio ao Bureau de Créditos
	// Incluir apenas informações essenciais para o Bureau (sensibilidade de dados)
	bcUsuarioRequest := bc.UsuarioRequest{
		UsuarioID:    usuarioID,
		TenantID:     tenantID,
		Email:        usuario.Email,
		Nome:         usuario.NomeCompleto,
		Documento:    usuario.DocumentoPrincipal,
		TipoDocumento: usuario.TipoDocumento,
		TipoVinculo:  tipoVinculo,
		NivelAcesso:  nivelAcesso,
	}

	// Registrar usuário no Bureau de Créditos
	bcUsuario, err := c.bcClient.RegisterUsuario(ctx, bcUsuarioRequest)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao registrar usuário no Bureau de Créditos")
		c.metrics.CountEvent("bureau_credito_vincular_usuario_error")
		return nil, fmt.Errorf("falha ao registrar usuário no Bureau de Créditos: %w", err)
	}

	// Armazenar o vínculo no IAM via DataCore
	detalhesJSON, _ := json.Marshal(detalhesAutorizacao)
	identity := &models.IntegrationIdentity{
		UsuarioID:      usuarioID,
		TenantID:       tenantID,
		IntegrationType: IntegrationType,
		ExternalID:     bcUsuario.ID,
		ExternalTenantID: bcUsuario.TenantID,
		ProfileType:    tipoVinculo,
		AccessLevel:    nivelAcesso,
		Status:         "ativo",
		Details:        string(detalhesJSON),
	}

	// Persistir no banco de dados via DataCore
	err = c.dataClient.CreateIntegrationIdentity(ctx, identity)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao persistir vínculo IAM-BureauCredito")
		c.metrics.CountEvent("bureau_credito_vincular_usuario_error")
		return nil, fmt.Errorf("falha ao persistir vínculo: %w", err)
	}

	// Registrar evento de auditoria de alta severidade (requisito para operações em Bureau de Créditos)
	c.registrarEventoAuditoria(ctx, "VINCULAR_USUARIO", usuarioID, tenantID, bcUsuario.ID, "Vínculo de usuário com Bureau de Créditos", nil)

	c.metrics.CountEvent("bureau_credito_vincular_usuario_success")
	return identity, nil
}

// CriarAutorizacaoConsulta cria uma autorização para consulta ao Bureau de Créditos
func (c *BureauCreditoConnector) CriarAutorizacaoConsulta(
	ctx context.Context,
	identityID string,
	tipoConsulta string,
	finalidade string,
	justificativa string,
	duracaoDias int,
	autorizadoPor string,
) (*models.BureauAutorizacao, error) {
	ctx, span := c.tracer.StartSpan(ctx, "BureauCreditoConnector.CriarAutorizacaoConsulta")
	defer span.End()

	c.metrics.CountEvent("bureau_credito_criar_autorizacao_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("bureau_credito_criar_autorizacao_error")
		return nil, fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	if identity.Status != "ativo" {
		c.metrics.CountEvent("bureau_credito_criar_autorizacao_error")
		return nil, errors.NewAuthorizationError("identidade de integração não está ativa")
	}

	// Definir duração padrão se não especificada
	if duracaoDias <= 0 {
		duracaoDias = 30 // 30 dias padrão
	}

	// Preparar solicitação para o Bureau de Créditos
	bcAutorizacaoRequest := bc.AutorizacaoRequest{
		UsuarioID:     identity.ExternalID,
		TenantID:      identity.ExternalTenantID,
		TipoConsulta:  tipoConsulta,
		Finalidade:    finalidade,
		Justificativa: justificativa,
		DataValidade:  time.Now().AddDate(0, 0, duracaoDias),
		AutorizadoPor: autorizadoPor,
	}

	// Solicitar autorização ao Bureau de Créditos
	bcAutorizacao, err := c.bcClient.CreateAutorizacao(ctx, bcAutorizacaoRequest)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao criar autorização no Bureau de Créditos")
		c.metrics.CountEvent("bureau_credito_criar_autorizacao_error")
		return nil, fmt.Errorf("falha ao criar autorização no Bureau de Créditos: %w", err)
	}

	// Criar autorização no banco de dados local
	autorizacao := &models.BureauAutorizacao{
		ID:            bcAutorizacao.ID,
		IdentityID:    identityID,
		TipoConsulta:  tipoConsulta,
		Finalidade:    finalidade,
		Justificativa: justificativa,
		DataAutorizacao: time.Now(),
		DataValidade:  time.Now().AddDate(0, 0, duracaoDias),
		Status:        "ativa",
		AutorizadoPor: autorizadoPor,
	}

	// Persistir autorização
	err = c.dataClient.CreateBureauAutorizacao(ctx, autorizacao)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao persistir autorização")
		c.metrics.CountEvent("bureau_credito_criar_autorizacao_error")
		return nil, fmt.Errorf("falha ao persistir autorização: %w", err)
	}

	// Registrar evento de auditoria de alta severidade
	c.registrarEventoAuditoria(
		ctx, 
		"CRIAR_AUTORIZACAO_CONSULTA", 
		identity.UsuarioID, 
		identity.TenantID, 
		bcAutorizacao.ID, 
		fmt.Sprintf("Autorização de consulta ao Bureau de Créditos: %s", tipoConsulta),
		map[string]interface{}{
			"tipo_consulta": tipoConsulta,
			"finalidade": finalidade,
			"duracao_dias": duracaoDias,
		},
	)

	c.metrics.CountEvent("bureau_credito_criar_autorizacao_success")
	return autorizacao, nil
}// Continuação do arquivo bureau_credito_connector.go

// GerarTokenAcesso gera um token de acesso para acesso ao Bureau de Créditos
func (c *BureauCreditoConnector) GerarTokenAcesso(
	ctx context.Context, 
	identityID string,
	finalidade string,
	escopo []string,
) (*models.AccessToken, error) {
	ctx, span := c.tracer.StartSpan(ctx, "BureauCreditoConnector.GerarTokenAcesso")
	defer span.End()

	c.metrics.CountEvent("bureau_credito_gerar_token_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("bureau_credito_gerar_token_error")
		return nil, fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	// Verificar se a identidade está ativa
	if identity.Status != "ativo" {
		c.metrics.CountEvent("bureau_credito_gerar_token_error")
		return nil, errors.NewAuthorizationError("identidade de integração não está ativa")
	}

	// Verificar se há autorizações válidas para os escopos solicitados
	autorizacoes, err := c.dataClient.GetValidBureauAutorizacoes(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar autorizações do Bureau")
		c.metrics.CountEvent("bureau_credito_gerar_token_error")
		return nil, fmt.Errorf("falha ao buscar autorizações: %w", err)
	}

	// Validar se os escopos solicitados estão cobertos pelas autorizações
	var autorizacaoValida bool
	for _, escopo := range escopo {
		for _, autorizacao := range autorizacoes {
			if autorizacao.TipoConsulta == escopo && autorizacao.Status == "ativa" {
				autorizacaoValida = true
				break
			}
		}
		if !autorizacaoValida {
			c.metrics.CountEvent("bureau_credito_gerar_token_unauthorized")
			return nil, errors.NewAuthorizationError(fmt.Sprintf("não há autorização válida para o escopo: %s", escopo))
		}
	}

	// Gerar token no Bureau de Créditos
	bcTokenRequest := bc.TokenRequest{
		UsuarioID: identity.ExternalID,
		TenantID:  identity.ExternalTenantID,
		Escopos:   escopo,
		Finalidade: finalidade,
	}

	bcToken, err := c.bcClient.GetAccessToken(ctx, bcTokenRequest)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao gerar token no Bureau de Créditos")
		c.metrics.CountEvent("bureau_credito_gerar_token_error")
		return nil, fmt.Errorf("falha ao gerar token no Bureau de Créditos: %w", err)
	}

	// Criar token no serviço local
	token := &models.AccessToken{
		Token:        bcToken.Token,
		Type:         "bearer",
		ExpiresAt:    time.Now().Add(time.Duration(bcToken.ExpiresIn) * time.Second),
		RefreshToken: bcToken.RefreshToken,
		IdentityID:   identityID,
		Scope:        escopo,
		Finalidade:   finalidade,
	}

	// Salvar o token no banco de dados
	err = c.dataClient.SaveAccessToken(ctx, token)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao salvar token")
		c.metrics.CountEvent("bureau_credito_gerar_token_error")
		return nil, fmt.Errorf("falha ao salvar token: %w", err)
	}

	// Registrar evento de auditoria
	c.registrarEventoAuditoria(
		ctx,
		"GERAR_TOKEN_ACESSO",
		identity.UsuarioID,
		identity.TenantID,
		token.Token,
		"Geração de token de acesso para Bureau de Créditos",
		map[string]interface{}{
			"escopos": escopo,
			"finalidade": finalidade,
		},
	)

	c.metrics.CountEvent("bureau_credito_gerar_token_success")
	return token, nil
}

// RevogarVinculo revoga o vínculo entre o usuário IAM e o Bureau de Créditos
func (c *BureauCreditoConnector) RevogarVinculo(
	ctx context.Context,
	identityID string,
	motivo string,
) error {
	ctx, span := c.tracer.StartSpan(ctx, "BureauCreditoConnector.RevogarVinculo")
	defer span.End()

	c.metrics.CountEvent("bureau_credito_revogar_vinculo_attempt")

	// Buscar a identidade de integração
	identity, err := c.dataClient.GetIntegrationIdentity(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao buscar identidade de integração")
		c.metrics.CountEvent("bureau_credito_revogar_vinculo_error")
		return fmt.Errorf("falha ao buscar identidade: %w", err)
	}

	// Verificar se a identidade já está inativa
	if identity.Status == "inativo" {
		c.logger.Info("Vínculo já está inativo")
		return nil
	}

	// Revogar vínculo no Bureau de Créditos
	err = c.bcClient.RevokeUsuario(ctx, identity.ExternalID, identity.ExternalTenantID, motivo)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao revogar usuário no Bureau de Créditos")
		c.metrics.CountEvent("bureau_credito_revogar_vinculo_error")
		return fmt.Errorf("falha ao revogar usuário no Bureau de Créditos: %w", err)
	}

	// Atualizar status da identidade para inativo
	identity.Status = "inativo"
	identity.UpdatedAt = time.Now()
	identity.StatusReason = motivo

	// Persistir atualização
	err = c.dataClient.UpdateIntegrationIdentity(ctx, identity)
	if err != nil {
		c.logger.WithError(err).Error("Falha ao atualizar status da identidade")
		c.metrics.CountEvent("bureau_credito_revogar_vinculo_error")
		return fmt.Errorf("falha ao atualizar status da identidade: %w", err)
	}

	// Revogar todos os tokens ativos para esta identidade
	err = c.dataClient.RevokeAllTokens(ctx, identityID)
	if err != nil {
		c.logger.WithError(err).Warn("Falha ao revogar todos os tokens")
		// Não interrompe o fluxo, apenas registra o aviso
	}

	// Registrar evento de auditoria
	c.registrarEventoAuditoria(
		ctx,
		"REVOGAR_VINCULO",
		identity.UsuarioID,
		identity.TenantID,
		identity.ExternalID,
		"Revogação de vínculo com Bureau de Créditos",
		map[string]interface{}{
			"motivo": motivo,
		},
	)

	c.metrics.CountEvent("bureau_credito_revogar_vinculo_success")
	return nil
}

// registrarEventoAuditoria registra um evento de auditoria para operações do Bureau de Créditos
func (c *BureauCreditoConnector) registrarEventoAuditoria(
	ctx context.Context,
	tipoEvento string,
	usuarioID string,
	tenantID string,
	objetoID string,
	mensagem string,
	detalhes map[string]interface{},
) {
	// Registrar evento no sistema de auditoria
	evento := &models.AuditEvent{
		TipoEvento:   tipoEvento,
		ModuloOrigem: ModuleName,
		UsuarioID:    usuarioID,
		TenantID:     tenantID,
		ObjetoID:     objetoID,
		Mensagem:     mensagem,
		Data:         time.Now(),
		Severidade:   "alta", // Operações de Bureau de Créditos sempre têm alta severidade
	}

	if detalhes != nil {
		detalhesJSON, _ := json.Marshal(detalhes)
		evento.Detalhes = string(detalhesJSON)
	}

	// Enviar para sistema de auditoria assincronamente
	go func() {
		ctxBg := context.Background()
		err := c.dataClient.CreateAuditEvent(ctxBg, evento)
		if err != nil {
			c.logger.WithError(err).Error("Falha ao registrar evento de auditoria")
		}
	}()
}