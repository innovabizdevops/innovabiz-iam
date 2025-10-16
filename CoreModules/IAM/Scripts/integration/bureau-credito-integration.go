// Bureau de Crédito (Central de Risco) - Integração com MCP-IAM Observability
// Desenvolvido para INNOVABIZ - Módulo Core
// Copyright © 2025 INNOVABIZ. Todos os direitos reservados.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/innovabizdevops/innovabiz-iam/observability/adapter"
	"github.com/innovabizdevops/innovabiz-iam/observability/constants"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TipoConsulta define os diferentes tipos de consulta ao Bureau de Crédito
type TipoConsulta string

const (
	// Tipos de consulta
	ConsultaCompleta       TipoConsulta = "completa"
	ConsultaScore          TipoConsulta = "score"
	ConsultaBasica         TipoConsulta = "basica"
	ConsultaRestricoes     TipoConsulta = "restricoes"
	ConsultaHistorico      TipoConsulta = "historico"
	ConsultaRelacionamento TipoConsulta = "relacionamento"
)

// FinalidadeConsulta define a finalidade da consulta ao Bureau de Crédito
type FinalidadeConsulta string

const (
	// Finalidades de consulta
	FinalidadeConcessaoCredito   FinalidadeConsulta = "concessao_credito"
	FinalidadeRevisaoLimites     FinalidadeConsulta = "revisao_limites"
	FinalidadeGerencialRegulador FinalidadeConsulta = "gerencial_regulador"
	FinalidadeVerificacaoCliente FinalidadeConsulta = "verificacao_cliente"
	FinalidadeAberturaConta      FinalidadeConsulta = "abertura_conta"
	FinalidadePrevencaoFraude    FinalidadeConsulta = "prevencao_fraude"
)

// OrigemRegistro define as possíveis origens dos registros de crédito
type OrigemRegistro string

const (
	// Origens de registro
	OrigemBancosCaixas      OrigemRegistro = "bancos_caixas"
	OrigemFinanceiras       OrigemRegistro = "financeiras"
	OrigemComercio          OrigemRegistro = "comercio"
	OrigemPrestadorServicos OrigemRegistro = "prestadores_servicos"
	OrigemTelecom           OrigemRegistro = "telecom"
	OrigemUtilidades        OrigemRegistro = "utilidades"
	OrigemMicroFinanceiras  OrigemRegistro = "microfinanceiras"
)

// TipoRegistro define os diferentes tipos de registros de crédito
type TipoRegistro string

const (
	// Tipos de registro
	RegistroInadimplencia      TipoRegistro = "inadimplencia"
	RegistroFraude             TipoRegistro = "fraude"
	RegistroProtesto           TipoRegistro = "protesto"
	RegistroHistoricoCredito   TipoRegistro = "historico_credito"
	RegistroScoreCredito       TipoRegistro = "score_credito"
	RegistroDividaAtiva        TipoRegistro = "divida_ativa"
	RegistroRecuperacaoCredito TipoRegistro = "recuperacao_credito"
)

// ConsultaCredito representa uma consulta ao Bureau de Crédito
type ConsultaCredito struct {
	ConsultaID       string             `json:"consultaId"`
	TipoConsulta     TipoConsulta       `json:"tipoConsulta"`
	Finalidade       FinalidadeConsulta `json:"finalidade"`
	EntidadeID       string             `json:"entidadeId"`
	TipoEntidade     string             `json:"tipoEntidade"` // PF ou PJ
	DocumentoCliente string             `json:"documentoCliente"`
	NomeCliente      string             `json:"nomeCliente"`
	UsuarioID        string             `json:"usuarioId"`
	DataConsulta     time.Time          `json:"dataConsulta"`
	ConsentimentoID  string             `json:"consentimentoId,omitempty"`
	SolicitanteID    string             `json:"solicitanteId"`
	MarketContext    adapter.MarketContext `json:"marketContext"`
	MFALevel         string             `json:"mfaLevel"`
	Parametros       map[string]interface{} `json:"parametros,omitempty"`
}

// RegistroCredito representa um registro no Bureau de Crédito
type RegistroCredito struct {
	RegistroID       string         `json:"registroId"`
	EntidadeID       string         `json:"entidadeId"`
	TipoEntidade     string         `json:"tipoEntidade"` // PF ou PJ
	DocumentoCliente string         `json:"documentoCliente"`
	Valor            float64        `json:"valor"`
	DataOcorrencia   time.Time      `json:"dataOcorrencia"`
	DataInclusao     time.Time      `json:"dataInclusao"`
	DataExpiracao    *time.Time     `json:"dataExpiracao,omitempty"`
	TipoRegistro     TipoRegistro   `json:"tipoRegistro"`
	OrigemRegistro   OrigemRegistro `json:"origemRegistro"`
	FonteID          string         `json:"fonteId"`
	FonteNome        string         `json:"fonteNome"`
	Detalhes         map[string]interface{} `json:"detalhes,omitempty"`
	MarketContext    adapter.MarketContext `json:"marketContext"`
	DocumentosRelacionados []string      `json:"documentosRelacionados,omitempty"`
}

// ResultadoConsulta representa o resultado de uma consulta ao Bureau de Crédito
type ResultadoConsulta struct {
	ConsultaID       string         `json:"consultaId"`
	DataResposta     time.Time      `json:"dataResposta"`
	ScoreCredito     *int           `json:"scoreCredito,omitempty"`
	FaixaRisco       *string        `json:"faixaRisco,omitempty"`
	RegistrosCredito []RegistroCredito `json:"registrosCredito,omitempty"`
	RestricoesList   []RegistroCredito `json:"restricoesList,omitempty"`
	DataAnalise      time.Time      `json:"dataAnalise"`
	RecomendacaoList []string       `json:"recomendacaoList,omitempty"`
	RelatorioURL     *string        `json:"relatorioUrl,omitempty"`
	MetadadosConsulta map[string]interface{} `json:"metadadosConsulta,omitempty"`
	TempoProcessamento int64         `json:"tempoProcessamento"`
}

// RegrasCompliance define as regras de compliance para operações do Bureau de Crédito
type RegrasCompliance struct {
	ID            string   `json:"id"`
	Market        string   `json:"market"`
	Description   string   `json:"description"`
	Framework     []string `json:"framework"`
	MandatoryFor  []string `json:"mandatoryFor"` // Tipos de consulta para os quais a regra é mandatória
	Validate      func(*ConsultaCredito) (bool, string, error)
}

// RegraAcesso define as regras de acesso aos dados do Bureau de Crédito
type RegraAcesso struct {
	ID           string   `json:"id"`
	Market       string   `json:"market"`
	Description  string   `json:"description"`
	TipoEntidade []string `json:"tipoEntidade"` // PF, PJ ou ambos
	TipoConsulta []string `json:"tipoConsulta"` // Tipos de consulta permitidos
	MFAMinimo    string   `json:"mfaMinimo"`    // Nível mínimo de MFA necessário
	Validate     func(*ConsultaCredito) (bool, string, error)
}

// BureauCreditoConfig representa a configuração para o Bureau de Crédito
type BureauCreditoConfig struct {
	Market                    string             `json:"market"`
	TenantType                string             `json:"tenantType"`
	Environment               string             `json:"environment"`
	TempoRetencaoHistorico    map[string]int     `json:"tempoRetencaoHistorico"` // Em meses, por tipo de registro
	FontesDadosAtivas         []string           `json:"fontesDadosAtivas"`
	LimiteConsultasDiarias    int                `json:"limiteConsultasDiarias"`
	ConsentimentoObrigatorio  map[string]bool    `json:"consentimentoObrigatorio"` // Por mercado
	ValidadeConsentimento     map[string]int     `json:"validadeConsentimento"`    // Em dias, por finalidade
	NotificacaoObrigatoria    map[string]bool    `json:"notificacaoObrigatoria"`   // Por mercado
	CamposObrigatorios        map[string][]string `json:"camposObrigatorios"`      // Por tipo de consulta
}

// BureauCredito representa o serviço de Bureau de Crédito
type BureauCredito struct {
	config              BureauCreditoConfig
	observability       adapter.IAMObservability
	logger              *zap.Logger
	regrasCompliance    []RegrasCompliance
	regrasAcesso        []RegraAcesso
	consultasDiarias    map[string]int // Contador de consultas diárias por entidade
	mutex               sync.RWMutex
	shutdown            chan struct{}
	wg                  sync.WaitGroup
}

// NewBureauCredito cria uma nova instância do Bureau de Crédito
func NewBureauCredito(config BureauCreditoConfig, obs adapter.IAMObservability, logger *zap.Logger) *BureauCredito {
	return &BureauCredito{
		config:           config,
		observability:    obs,
		logger:           logger,
		regrasCompliance: []RegrasCompliance{},
		regrasAcesso:     []RegraAcesso{},
		consultasDiarias: make(map[string]int),
		shutdown:         make(chan struct{}),
	}
}// RealizarConsulta processa uma consulta ao Bureau de Crédito
func (bc *BureauCredito) RealizarConsulta(ctx context.Context, consulta ConsultaCredito) (*ResultadoConsulta, error) {
	// Iniciar rastreamento com OpenTelemetry
	ctx, span := bc.observability.Tracer().Start(ctx, "bureau_credito_consulta",
		trace.WithAttributes(
			attribute.String("consulta_id", consulta.ConsultaID),
			attribute.String("tipo_consulta", string(consulta.TipoConsulta)),
			attribute.String("finalidade", string(consulta.Finalidade)),
			attribute.String("entidade_id", consulta.EntidadeID),
			attribute.String("documento_cliente", consulta.DocumentoCliente),
			attribute.String("market", consulta.MarketContext.Market),
		),
	)
	defer span.End()

	// Registrar tentativa de consulta nos logs
	bc.logger.Info("Iniciando consulta ao Bureau de Crédito",
		zap.String("consulta_id", consulta.ConsultaID),
		zap.String("tipo_consulta", string(consulta.TipoConsulta)),
		zap.String("finalidade", string(consulta.Finalidade)),
		zap.String("entidade_id", consulta.EntidadeID),
		zap.String("documento_cliente", consulta.DocumentoCliente),
		zap.String("market", consulta.MarketContext.Market))

	// Registrar evento de auditoria para a consulta
	bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
		"bureau_credito_consulta_iniciada",
		fmt.Sprintf("Consulta %s iniciada para documento %s (tipo: %s, finalidade: %s)",
			consulta.ConsultaID, consulta.DocumentoCliente, consulta.TipoConsulta, consulta.Finalidade))

	// Registrar métrica de tentativa de consulta
	bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_consultas_total", 
		string(consulta.TipoConsulta), 1)

	// Verificar autenticação do usuário
	authenticated, err := bc.verificarAutenticacao(ctx, consulta)
	if err != nil {
		bc.logger.Error("Falha na verificação de autenticação",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("usuario_id", consulta.UsuarioID),
			zap.Error(err))
		
		// Registrar evento de segurança para falha de autenticação
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityHigh, "bureau_credito_auth_failure",
			fmt.Sprintf("Falha de autenticação na consulta %s: %v", consulta.ConsultaID, err))
		
		return nil, fmt.Errorf("falha na verificação de autenticação: %w", err)
	}

	// Verificar autorização do usuário
	authorized, err := bc.verificarAutorizacao(ctx, consulta)
	if err != nil {
		bc.logger.Error("Falha na verificação de autorização",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("usuario_id", consulta.UsuarioID),
			zap.Error(err))
		
		// Registrar evento de segurança para falha de autorização
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityHigh, "bureau_credito_auth_failure",
			fmt.Sprintf("Falha de autorização na consulta %s: %v", consulta.ConsultaID, err))
		
		return nil, fmt.Errorf("falha na verificação de autorização: %w", err)
	}

	// Verificar regras de acesso específicas
	if err := bc.verificarRegrasAcesso(ctx, consulta); err != nil {
		bc.logger.Error("Falha na verificação de regras de acesso",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("usuario_id", consulta.UsuarioID),
			zap.Error(err))
		
		// Registrar evento de segurança para falha nas regras de acesso
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityMedium, "bureau_credito_access_denied",
			fmt.Sprintf("Acesso negado na consulta %s: %v", consulta.ConsultaID, err))
		
		return nil, fmt.Errorf("acesso negado: %w", err)
	}

	// Verificar limite diário de consultas
	if err := bc.verificarLimiteConsultas(ctx, consulta); err != nil {
		bc.logger.Warn("Limite de consultas excedido",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("entidade_id", consulta.EntidadeID),
			zap.Error(err))
		
		// Registrar evento de segurança para limite excedido
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityMedium, "bureau_credito_limit_exceeded",
			fmt.Sprintf("Limite de consultas excedido para entidade %s: %v", 
				consulta.EntidadeID, err))
		
		return nil, fmt.Errorf("limite de consultas excedido: %w", err)
	}

	// Verificar consentimento quando necessário
	if err := bc.verificarConsentimento(ctx, consulta); err != nil {
		bc.logger.Error("Falha na verificação de consentimento",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("documento_cliente", consulta.DocumentoCliente),
			zap.Error(err))
		
		// Registrar evento de segurança para falha de consentimento
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityHigh, "bureau_credito_consent_failure",
			fmt.Sprintf("Consentimento ausente ou inválido na consulta %s: %v", 
				consulta.ConsultaID, err))
		
		return nil, fmt.Errorf("consentimento ausente ou inválido: %w", err)
	}

	// Verificar regras de compliance específicas por mercado
	if err := bc.verificarCompliance(ctx, consulta); err != nil {
		bc.logger.Error("Falha na verificação de compliance",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("market", consulta.MarketContext.Market),
			zap.Error(err))
		
		// Registrar evento de segurança para falha de compliance
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityHigh, "bureau_credito_compliance_failure",
			fmt.Sprintf("Falha de compliance na consulta %s: %v", consulta.ConsultaID, err))
		
		return nil, fmt.Errorf("falha de compliance: %w", err)
	}

	// Iniciar tempo de processamento
	startTime := time.Now()

	// Processar a consulta (simulada para este exemplo)
	resultado, err := bc.processarConsulta(ctx, consulta)
	if err != nil {
		bc.logger.Error("Erro no processamento da consulta",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.Error(err))
		
		return nil, fmt.Errorf("erro no processamento da consulta: %w", err)
	}

	// Calcular tempo de processamento
	processTime := time.Since(startTime).Milliseconds()
	resultado.TempoProcessamento = processTime

	// Incrementar contador de consultas diárias
	bc.incrementarConsultasDiarias(consulta.EntidadeID)

	// Registrar evento de auditoria para consulta bem-sucedida
	bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
		"bureau_credito_consulta_concluida",
		fmt.Sprintf("Consulta %s concluída para documento %s (tipo: %s, tempo: %dms)",
			consulta.ConsultaID, consulta.DocumentoCliente, consulta.TipoConsulta, processTime))

	// Registrar métricas de sucesso e tempo de processamento
	bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_consultas_sucesso", 
		string(consulta.TipoConsulta), 1)
	bc.observability.RecordHistogram(consulta.MarketContext, "bureau_credito_tempo_processamento", 
		float64(processTime), string(consulta.TipoConsulta))

	// Verificar notificações obrigatórias por regulador
	bc.processarNotificacoes(ctx, consulta, resultado)

	return resultado, nil
}

// verificarAutenticacao verifica a autenticação do usuário
func (bc *BureauCredito) verificarAutenticacao(ctx context.Context, consulta ConsultaCredito) (bool, error) {
	ctx, span := bc.observability.Tracer().Start(ctx, "verificar_autenticacao")
	defer span.End()

	// Obter metadados de compliance para o mercado
	metadata, exists := bc.observability.GetComplianceMetadata(consulta.MarketContext.Market)
	if !exists {
		metadata, _ = bc.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Determinar nível MFA necessário com base no tipo de consulta
	var requiredMFALevel string
	
	// Consultas completas ou com dados sensíveis exigem MFA mais forte
	if consulta.TipoConsulta == ConsultaCompleta || consulta.TipoConsulta == ConsultaHistorico {
		requiredMFALevel = "high"
	} else {
		// Caso contrário, usar requisito padrão do mercado
		requiredMFALevel = metadata.RequiredMFALevel
	}
	
	// Verificar se o nível MFA fornecido é suficiente
	mfaResult, err := bc.observability.ValidateMFA(ctx, consulta.MarketContext, consulta.UsuarioID, consulta.MFALevel)
	if err != nil {
		return false, err
	}

	if !mfaResult {
		return false, fmt.Errorf("nível MFA insuficiente para consulta ao Bureau de Crédito no mercado %s: requer %s, fornecido %s",
			consulta.MarketContext.Market, requiredMFALevel, consulta.MFALevel)
	}

	// Registrar evento de auditoria
	bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID, "authentication_verified",
		fmt.Sprintf("Autenticação verificada com MFA nível %s para consulta ao Bureau de Crédito", consulta.MFALevel))

	return true, nil
}

// verificarAutorizacao verifica a autorização do usuário para a consulta
func (bc *BureauCredito) verificarAutorizacao(ctx context.Context, consulta ConsultaCredito) (bool, error) {
	ctx, span := bc.observability.Tracer().Start(ctx, "verificar_autorizacao")
	defer span.End()

	// Verificar escopo para consulta ao Bureau de Crédito
	scope := fmt.Sprintf("bureau_credito:%s", consulta.TipoConsulta)
	scopeResult, err := bc.observability.ValidateScope(ctx, consulta.MarketContext, consulta.UsuarioID, scope)
	if err != nil {
		return false, err
	}

	if !scopeResult {
		return false, fmt.Errorf("usuário não tem escopo para consultas de tipo %s", consulta.TipoConsulta)
	}

	// Verificar escopo adicional para finalidade específica
	finalidadeScope := fmt.Sprintf("bureau_credito:finalidade:%s", consulta.Finalidade)
	finalidadeScopeResult, err := bc.observability.ValidateScope(ctx, consulta.MarketContext, consulta.UsuarioID, finalidadeScope)
	if err != nil || !finalidadeScopeResult {
		return false, fmt.Errorf("usuário não tem escopo para finalidade %s", consulta.Finalidade)
	}

	// Verificar requisitos específicos por mercado
	switch consulta.MarketContext.Market {
	case constants.MarketAngola:
		// BNA exige verificação adicional para consultas completas
		if consulta.TipoConsulta == ConsultaCompleta {
			additionalScope, err := bc.observability.ValidateScope(ctx, consulta.MarketContext, consulta.UsuarioID, "bureau_credito:bna:completa")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BNA para consultas completas")
			}
		}
		
	case constants.MarketBrazil:
		// BACEN exige escopo especial para consultas de histórico
		if consulta.TipoConsulta == ConsultaHistorico {
			additionalScope, err := bc.observability.ValidateScope(ctx, consulta.MarketContext, consulta.UsuarioID, "bureau_credito:bacen:historico")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo BACEN para consultas de histórico")
			}
		}

	case constants.MarketEU:
		// GDPR exige escopo especial para finalidades regulatórias
		if consulta.Finalidade == FinalidadeGerencialRegulador {
			additionalScope, err := bc.observability.ValidateScope(ctx, consulta.MarketContext, consulta.UsuarioID, "bureau_credito:gdpr:regulatory")
			if err != nil || !additionalScope {
				return false, fmt.Errorf("usuário não tem escopo GDPR para finalidades regulatórias")
			}
		}
	}

	// Registrar evento de auditoria
	bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID, "authorization_verified",
		fmt.Sprintf("Autorização verificada para consulta ao Bureau de Crédito %s", consulta.ConsultaID))

	return true, nil
}

// verificarRegrasAcesso verifica as regras de acesso específicas
func (bc *BureauCredito) verificarRegrasAcesso(ctx context.Context, consulta ConsultaCredito) error {
	ctx, span := bc.observability.Tracer().Start(ctx, "verificar_regras_acesso")
	defer span.End()

	// Verificar regras de acesso aplicáveis
	for _, regra := range bc.regrasAcesso {
		// Verificar se a regra se aplica ao mercado atual ou é global
		if regra.Market == constants.MarketGlobal || regra.Market == consulta.MarketContext.Market {
			// Verificar se a regra se aplica ao tipo de entidade
			aplicaTipoEntidade := false
			for _, tipoEntidade := range regra.TipoEntidade {
				if tipoEntidade == consulta.TipoEntidade {
					aplicaTipoEntidade = true
					break
				}
			}

			// Verificar se a regra se aplica ao tipo de consulta
			aplicaTipoConsulta := false
			for _, tipoConsulta := range regra.TipoConsulta {
				if tipoConsulta == string(consulta.TipoConsulta) {
					aplicaTipoConsulta = true
					break
				}
			}

			if aplicaTipoEntidade && aplicaTipoConsulta {
				// Aplicar regra de acesso
				acessoPermitido, mensagem, err := regra.Validate(&consulta)
				if err != nil {
					bc.logger.Error("Erro ao validar regra de acesso",
						zap.String("regra_id", regra.ID),
						zap.String("consulta_id", consulta.ConsultaID),
						zap.Error(err))
					return fmt.Errorf("erro ao validar regra de acesso %s: %w", regra.ID, err)
				}

				if !acessoPermitido {
					// Registrar evento de segurança para acesso negado
					bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
						constants.SecurityEventSeverityMedium, "bureau_credito_access_denied",
						fmt.Sprintf("Regra de acesso %s negou consulta %s: %s", 
							regra.ID, consulta.ConsultaID, mensagem))
					
					return fmt.Errorf("acesso negado - %s: %s", regra.ID, mensagem)
				}

				// Registrar evento de auditoria para acesso permitido
				bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
					fmt.Sprintf("access_rule_%s_verified", regra.ID),
					fmt.Sprintf("Regra de acesso %s verificada: %s", regra.ID, mensagem))
			}
		}
	}

	return nil
}// verificarLimiteConsultas verifica se o limite diário de consultas foi atingido
func (bc *BureauCredito) verificarLimiteConsultas(ctx context.Context, consulta ConsultaCredito) error {
	ctx, span := bc.observability.Tracer().Start(ctx, "verificar_limite_consultas")
	defer span.End()

	bc.mutex.RLock()
	consultas, existe := bc.consultasDiarias[consulta.EntidadeID]
	bc.mutex.RUnlock()

	if !existe {
		return nil // Primeira consulta do dia
	}

	// Verificar limite configurado
	if consultas >= bc.config.LimiteConsultasDiarias {
		// Registrar evento de segurança para limite excedido
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityMedium, "bureau_credito_limit_exceeded",
			fmt.Sprintf("Limite de consultas diárias excedido para entidade %s: %d/%d", 
				consulta.EntidadeID, consultas, bc.config.LimiteConsultasDiarias))
		
		// Registrar métrica de limite excedido
		bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_limite_excedido", 
			string(consulta.TipoConsulta), 1)
		
		return fmt.Errorf("limite diário de %d consultas excedido para entidade %s", 
			bc.config.LimiteConsultasDiarias, consulta.EntidadeID)
	}

	return nil
}

// verificarConsentimento verifica se o consentimento é necessário e válido
func (bc *BureauCredito) verificarConsentimento(ctx context.Context, consulta ConsultaCredito) error {
	ctx, span := bc.observability.Tracer().Start(ctx, "verificar_consentimento")
	defer span.End()

	// Verificar se o consentimento é obrigatório para este mercado
	consentimentoObrigatorio, existe := bc.config.ConsentimentoObrigatorio[consulta.MarketContext.Market]
	if !existe {
		// Se não houver configuração específica, usar global
		consentimentoObrigatorio = bc.config.ConsentimentoObrigatorio[constants.MarketGlobal]
	}

	// Se o consentimento não for obrigatório, retornar sucesso
	if !consentimentoObrigatorio {
		return nil
	}

	// Consentimento é obrigatório, mas não foi fornecido
	if consulta.ConsentimentoID == "" {
		return fmt.Errorf("consentimento obrigatório para consultas no mercado %s", consulta.MarketContext.Market)
	}

	// Verificar validade do consentimento
	valido, err := bc.observability.ValidateConsent(ctx, consulta.MarketContext, consulta.DocumentoCliente, consulta.ConsentimentoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar consentimento: %w", err)
	}

	if !valido {
		// Registrar evento de segurança para consentimento inválido
		bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			constants.SecurityEventSeverityHigh, "bureau_credito_invalid_consent",
			fmt.Sprintf("Consentimento inválido na consulta %s para documento %s", 
				consulta.ConsultaID, consulta.DocumentoCliente))
		
		return fmt.Errorf("consentimento inválido ou expirado")
	}

	// Registrar evento de auditoria para consentimento válido
	bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
		"consent_verified",
		fmt.Sprintf("Consentimento verificado para consulta %s, documento %s", 
			consulta.ConsultaID, consulta.DocumentoCliente))

	return nil
}

// verificarCompliance verifica as regras de compliance específicas
func (bc *BureauCredito) verificarCompliance(ctx context.Context, consulta ConsultaCredito) error {
	ctx, span := bc.observability.Tracer().Start(ctx, "verificar_compliance")
	defer span.End()

	// Verificar regras de compliance aplicáveis
	for _, regra := range bc.regrasCompliance {
		// Verificar se a regra se aplica ao mercado atual ou é global
		if regra.Market == constants.MarketGlobal || regra.Market == consulta.MarketContext.Market {
			// Verificar se a regra é obrigatória para este tipo de consulta
			aplicaTipoConsulta := false
			for _, tipoConsulta := range regra.MandatoryFor {
				if tipoConsulta == string(consulta.TipoConsulta) {
					aplicaTipoConsulta = true
					break
				}
			}

			if aplicaTipoConsulta {
				// Aplicar regra de compliance
				conforme, mensagem, err := regra.Validate(&consulta)
				if err != nil {
					bc.logger.Error("Erro ao validar regra de compliance",
						zap.String("regra_id", regra.ID),
						zap.String("consulta_id", consulta.ConsultaID),
						zap.Error(err))
					return fmt.Errorf("erro ao validar regra de compliance %s: %w", regra.ID, err)
				}

				if !conforme {
					// Registrar evento de segurança para compliance violado
					bc.observability.TraceSecurityEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
						constants.SecurityEventSeverityHigh, "bureau_credito_compliance_violation",
						fmt.Sprintf("Regra de compliance %s violada na consulta %s: %s", 
							regra.ID, consulta.ConsultaID, mensagem))
					
					return fmt.Errorf("violação de compliance - %s: %s", regra.ID, mensagem)
				}

				// Registrar evento de auditoria para compliance verificado
				bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
					fmt.Sprintf("compliance_rule_%s_verified", regra.ID),
					fmt.Sprintf("Regra de compliance %s verificada: %s", regra.ID, mensagem))
			}
		}
	}

	return nil
}

// processarConsulta simula o processamento real de uma consulta
func (bc *BureauCredito) processarConsulta(ctx context.Context, consulta ConsultaCredito) (*ResultadoConsulta, error) {
	ctx, span := bc.observability.Tracer().Start(ctx, "processar_consulta")
	defer span.End()

	// Registrar evento de início do processamento
	bc.logger.Info("Processando consulta",
		zap.String("consulta_id", consulta.ConsultaID),
		zap.String("tipo", string(consulta.TipoConsulta)),
		zap.String("documento_cliente", consulta.DocumentoCliente))
	
	// Simular tempo de processamento
	time.Sleep(100 * time.Millisecond)

	// Criar resultado simulado baseado no tipo de consulta
	resultado := &ResultadoConsulta{
		ConsultaID:   consulta.ConsultaID,
		DataResposta: time.Now(),
		DataAnalise:  time.Now(),
		MetadadosConsulta: map[string]interface{}{
			"market":       consulta.MarketContext.Market,
			"finalidade":   string(consulta.Finalidade),
			"tipoConsulta": string(consulta.TipoConsulta),
		},
		RegistrosCredito: []RegistroCredito{},
		RestricoesList:   []RegistroCredito{},
	}

	// Simulação específica para cada tipo de consulta
	switch consulta.TipoConsulta {
	case ConsultaCompleta:
		// Gerar score simulado
		score := 650 + (consulta.ConsultaID[0] % 30) * 10
		resultado.ScoreCredito = &score
		
		// Determinar faixa de risco com base no score
		var faixaRisco string
		switch {
		case score >= 800:
			faixaRisco = "Risco Muito Baixo"
		case score >= 700:
			faixaRisco = "Risco Baixo"
		case score >= 600:
			faixaRisco = "Risco Médio"
		case score >= 500:
			faixaRisco = "Risco Alto"
		default:
			faixaRisco = "Risco Muito Alto"
		}
		resultado.FaixaRisco = &faixaRisco
		
		// Adicionar registros simulados
		resultado.RegistrosCredito = bc.gerarRegistrosSimulados(consulta, 5)
		resultado.RestricoesList = bc.gerarRestricoesSimuladas(consulta, 2)
		
		// Adicionar recomendações
		resultado.RecomendacaoList = []string{
			"Verificar documentação adicional",
			"Solicitar garantias complementares",
			"Considerar limite reduzido para primeiras operações",
		}
		
		// URL para relatório detalhado
		relatorioURL := fmt.Sprintf("https://bureau.innovabiz.com/reports/%s", consulta.ConsultaID)
		resultado.RelatorioURL = &relatorioURL
		
	case ConsultaScore:
		// Gerar score simulado
		score := 650 + (consulta.ConsultaID[0] % 30) * 10
		resultado.ScoreCredito = &score
		
		// Determinar faixa de risco
		var faixaRisco string
		switch {
		case score >= 800:
			faixaRisco = "Risco Muito Baixo"
		case score >= 700:
			faixaRisco = "Risco Baixo"
		case score >= 600:
			faixaRisco = "Risco Médio"
		case score >= 500:
			faixaRisco = "Risco Alto"
		default:
			faixaRisco = "Risco Muito Alto"
		}
		resultado.FaixaRisco = &faixaRisco
		
	case ConsultaBasica:
		// Apenas informação básica
		resultado.RestricoesList = bc.gerarRestricoesSimuladas(consulta, 1)
		
	case ConsultaRestricoes:
		// Apenas restrições
		resultado.RestricoesList = bc.gerarRestricoesSimuladas(consulta, 3)
		
	case ConsultaHistorico:
		// Apenas histórico de crédito
		resultado.RegistrosCredito = bc.gerarRegistrosSimulados(consulta, 8)
		
	case ConsultaRelacionamento:
		// Dados de relacionamento com instituições
		resultado.RegistrosCredito = bc.gerarRegistrosSimulados(consulta, 3)
		// Adicionar metadados específicos
		resultado.MetadadosConsulta["quantidadeInstituicoes"] = 3
		resultado.MetadadosConsulta["tempoMedioRelacionamento"] = "2.5 anos"
	}

	// Registrar métricas específicas
	if resultado.ScoreCredito != nil {
		bc.observability.RecordHistogram(consulta.MarketContext, "bureau_credito_score", 
			float64(*resultado.ScoreCredito), string(consulta.TipoConsulta))
	}
	
	bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_registros_retornados", 
		string(consulta.TipoConsulta), float64(len(resultado.RegistrosCredito) + len(resultado.RestricoesList)))
	
	return resultado, nil
}

// gerarRegistrosSimulados gera registros simulados para teste
func (bc *BureauCredito) gerarRegistrosSimulados(consulta ConsultaCredito, quantidade int) []RegistroCredito {
	registros := make([]RegistroCredito, 0, quantidade)
	
	origensRegistro := []OrigemRegistro{
		OrigemBancosCaixas,
		OrigemFinanceiras,
		OrigemComercio,
		OrigemPrestadorServicos,
		OrigemTelecom,
	}
	
	tiposRegistro := []TipoRegistro{
		RegistroHistoricoCredito,
		RegistroScoreCredito,
	}
	
	for i := 0; i < quantidade; i++ {
		// Obter valores simulados com base na iteração
		origemIndex := i % len(origensRegistro)
		tipoIndex := i % len(tiposRegistro)
		
		// Gerar datas simuladas
		dataOcorrencia := time.Now().AddDate(0, -(i+1), 0)
		dataInclusao := dataOcorrencia.AddDate(0, 0, 2)
		var dataExpiracao *time.Time
		
		// Adicionar data de expiração para alguns registros
		if i%2 == 0 {
			expiracao := time.Now().AddDate(5, 0, 0)
			dataExpiracao = &expiracao
		}
		
		// Criar registro
		registro := RegistroCredito{
			RegistroID:       fmt.Sprintf("REG%d%s", i, consulta.ConsultaID[:5]),
			EntidadeID:       consulta.EntidadeID,
			TipoEntidade:     consulta.TipoEntidade,
			DocumentoCliente: consulta.DocumentoCliente,
			Valor:            1000.0 + float64(i*500),
			DataOcorrencia:   dataOcorrencia,
			DataInclusao:     dataInclusao,
			DataExpiracao:    dataExpiracao,
			TipoRegistro:     tiposRegistro[tipoIndex],
			OrigemRegistro:   origensRegistro[origemIndex],
			FonteID:          fmt.Sprintf("FONTE%d", i+1),
			FonteNome:        fmt.Sprintf("Instituição Simulada %d", i+1),
			Detalhes: map[string]interface{}{
				"tipoOperacao":   "Simulada",
				"situacaoAtual":  "Regular",
				"numContrato":    fmt.Sprintf("CONT%d", 10000+i),
			},
			MarketContext: consulta.MarketContext,
			DocumentosRelacionados: []string{
				fmt.Sprintf("DOC%d", 1000+i),
			},
		}
		
		registros = append(registros, registro)
	}
	
	return registros
}

// gerarRestricoesSimuladas gera restrições simuladas para teste
func (bc *BureauCredito) gerarRestricoesSimuladas(consulta ConsultaCredito, quantidade int) []RegistroCredito {
	restricoes := make([]RegistroCredito, 0, quantidade)
	
	origensRegistro := []OrigemRegistro{
		OrigemComercio,
		OrigemTelecom,
		OrigemUtilidades,
		OrigemBancosCaixas,
	}
	
	tiposRegistro := []TipoRegistro{
		RegistroInadimplencia,
		RegistroProtesto,
		RegistroDividaAtiva,
	}
	
	for i := 0; i < quantidade; i++ {
		// Obter valores simulados com base na iteração
		origemIndex := i % len(origensRegistro)
		tipoIndex := i % len(tiposRegistro)
		
		// Gerar datas simuladas
		dataOcorrencia := time.Now().AddDate(0, -(i+3), 0)
		dataInclusao := dataOcorrencia.AddDate(0, 0, 5)
		
		// Criar registro de restrição
		restricao := RegistroCredito{
			RegistroID:       fmt.Sprintf("RES%d%s", i, consulta.ConsultaID[:5]),
			EntidadeID:       consulta.EntidadeID,
			TipoEntidade:     consulta.TipoEntidade,
			DocumentoCliente: consulta.DocumentoCliente,
			Valor:            500.0 + float64(i*200),
			DataOcorrencia:   dataOcorrencia,
			DataInclusao:     dataInclusao,
			TipoRegistro:     tiposRegistro[tipoIndex],
			OrigemRegistro:   origensRegistro[origemIndex],
			FonteID:          fmt.Sprintf("FONTE%d", i+20),
			FonteNome:        fmt.Sprintf("Credor Simulado %d", i+1),
			Detalhes: map[string]interface{}{
				"motivoRestricao": "Simulação",
				"situacaoAtual":   "Pendente",
				"diasAtraso":      30 + i*15,
			},
			MarketContext: consulta.MarketContext,
		}
		
		restricoes = append(restricoes, restricao)
	}
	
	return restricoes
}// processarNotificacoes processa notificações obrigatórias por mercado
func (bc *BureauCredito) processarNotificacoes(ctx context.Context, consulta ConsultaCredito, resultado *ResultadoConsulta) {
	ctx, span := bc.observability.Tracer().Start(ctx, "processar_notificacoes")
	defer span.End()

	// Verificar se notificação é obrigatória para este mercado
	notificacaoObrigatoria, existe := bc.config.NotificacaoObrigatoria[consulta.MarketContext.Market]
	if !existe {
		// Se não houver configuração específica, usar global
		notificacaoObrigatoria = bc.config.NotificacaoObrigatoria[constants.MarketGlobal]
	}

	// Se a notificação não for obrigatória, sair
	if !notificacaoObrigatoria {
		return
	}

	// Obter metadados de compliance para o mercado
	metadata, existe := bc.observability.GetComplianceMetadata(consulta.MarketContext.Market)
	if !existe {
		metadata, _ = bc.observability.GetComplianceMetadata(constants.MarketGlobal)
	}

	// Processar notificações específicas por mercado
	switch consulta.MarketContext.Market {
	case constants.MarketAngola:
		// BNA exige notificação ao cliente para qualquer consulta completa
		if consulta.TipoConsulta == ConsultaCompleta {
			// Enviar notificação (simulado)
			bc.logger.Info("Enviando notificação BNA para consulta completa",
				zap.String("consulta_id", consulta.ConsultaID),
				zap.String("documento_cliente", consulta.DocumentoCliente))
			
			// Registrar evento de auditoria para notificação
			bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
				"bna_notificacao_enviada",
				fmt.Sprintf("Notificação BNA enviada para documento %s referente à consulta %s", 
					consulta.DocumentoCliente, consulta.ConsultaID))
			
			// Registrar métrica de notificação
			bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_notificacoes", 
				"bna", 1)
		}
		
	case constants.MarketBrazil:
		// LGPD/BACEN exige notificação ao cliente para qualquer consulta que retorne restrições
		if len(resultado.RestricoesList) > 0 {
			// Enviar notificação (simulado)
			bc.logger.Info("Enviando notificação LGPD/BACEN para consulta com restrições",
				zap.String("consulta_id", consulta.ConsultaID),
				zap.String("documento_cliente", consulta.DocumentoCliente),
				zap.Int("qtd_restricoes", len(resultado.RestricoesList)))
			
			// Registrar evento de auditoria para notificação
			bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
				"lgpd_bacen_notificacao_enviada",
				fmt.Sprintf("Notificação LGPD/BACEN enviada para documento %s referente a %d restrições", 
					consulta.DocumentoCliente, len(resultado.RestricoesList)))
			
			// Registrar métrica de notificação
			bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_notificacoes", 
				"lgpd_bacen", 1)
		}
		
	case constants.MarketEU:
		// GDPR exige notificação ao cliente para qualquer consulta
		// Enviar notificação (simulado)
		bc.logger.Info("Enviando notificação GDPR para consulta",
			zap.String("consulta_id", consulta.ConsultaID),
			zap.String("documento_cliente", consulta.DocumentoCliente))
		
		// Registrar evento de auditoria para notificação
		bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
			"gdpr_notificacao_enviada",
			fmt.Sprintf("Notificação GDPR enviada para documento %s referente à consulta %s", 
				consulta.DocumentoCliente, consulta.ConsultaID))
		
		// Registrar métrica de notificação
		bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_notificacoes", 
			"gdpr", 1)

	case constants.MarketUSA:
		// FCRA (Fair Credit Reporting Act) exige notificação ao cliente para qualquer consulta que impacte negativamente
		if resultado.FaixaRisco != nil && (*resultado.FaixaRisco == "Risco Alto" || *resultado.FaixaRisco == "Risco Muito Alto") {
			// Enviar notificação (simulado)
			bc.logger.Info("Enviando notificação FCRA para consulta com risco elevado",
				zap.String("consulta_id", consulta.ConsultaID),
				zap.String("documento_cliente", consulta.DocumentoCliente),
				zap.String("faixa_risco", *resultado.FaixaRisco))
			
			// Registrar evento de auditoria para notificação
			bc.observability.TraceAuditEvent(ctx, consulta.MarketContext, consulta.UsuarioID,
				"fcra_notificacao_enviada",
				fmt.Sprintf("Notificação FCRA enviada para documento %s referente à faixa de risco %s", 
					consulta.DocumentoCliente, *resultado.FaixaRisco))
			
			// Registrar métrica de notificação
			bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_notificacoes", 
				"fcra", 1)
		}
	}

	// Para todos os reguladores, registrar evento de notificação
	if metadata.RegulatoryBodies != nil && len(metadata.RegulatoryBodies) > 0 {
		for _, regulador := range metadata.RegulatoryBodies {
			bc.logger.Debug("Registrando consulta para regulador",
				zap.String("regulador", regulador),
				zap.String("consulta_id", consulta.ConsultaID))
			
			// Registrar métrica de notificação por regulador
			bc.observability.RecordMetric(consulta.MarketContext, "bureau_credito_registros_regulador", 
				regulador, 1)
		}
	}
}

// incrementarConsultasDiarias incrementa o contador de consultas diárias
func (bc *BureauCredito) incrementarConsultasDiarias(entidadeID string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if _, existe := bc.consultasDiarias[entidadeID]; !existe {
		bc.consultasDiarias[entidadeID] = 0
	}

	bc.consultasDiarias[entidadeID]++
}

// resetConsultasDiarias reseta o contador de consultas diárias
func (bc *BureauCredito) resetConsultasDiarias() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.consultasDiarias = make(map[string]int)
}

// RegistrarRegraCompliance adiciona uma nova regra de compliance
func (bc *BureauCredito) RegistrarRegraCompliance(regra RegrasCompliance) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.regrasCompliance = append(bc.regrasCompliance, regra)
}

// RegistrarRegraAcesso adiciona uma nova regra de acesso
func (bc *BureauCredito) RegistrarRegraAcesso(regra RegraAcesso) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.regrasAcesso = append(bc.regrasAcesso, regra)
}

// Start inicia o serviço Bureau de Crédito
func (bc *BureauCredito) Start() error {
	// Iniciar worker para reset diário
	bc.wg.Add(1)
	go bc.startDailyResetWorker()

	// Registrar métrica de início do serviço
	marketContext := adapter.MarketContext{
		Market:     bc.config.Market,
		TenantType: bc.config.TenantType,
	}
	bc.observability.RecordMetric(marketContext, "bureau_credito_service_started", "start", 1)

	bc.logger.Info("Serviço Bureau de Crédito iniciado",
		zap.String("market", bc.config.Market),
		zap.String("environment", bc.config.Environment))

	return nil
}

// startDailyResetWorker inicia o worker para reset diário de consultas
func (bc *BureauCredito) startDailyResetWorker() {
	defer bc.wg.Done()

	bc.logger.Info("Worker de reset diário iniciado")
	
	// Configurar ticker para execução diária à meia-noite
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	initialDelay := nextMidnight.Sub(now)
	
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	// Esperar até a meia-noite para o primeiro reset
	timer := time.NewTimer(initialDelay)
	defer timer.Stop()
	
	marketContext := adapter.MarketContext{
		Market:     bc.config.Market,
		TenantType: bc.config.TenantType,
	}

	for {
		select {
		case <-bc.shutdown:
			bc.logger.Info("Worker de reset diário encerrado")
			return
			
		case <-timer.C:
			// Realizar o primeiro reset à meia-noite
			bc.logger.Info("Realizando reset diário inicial de consultas")
			bc.resetConsultasDiarias()
			bc.observability.RecordMetric(marketContext, "bureau_credito_daily_reset", "reset", 1)
			
			// Agora usar o ticker para resets diários
			for {
				select {
				case <-bc.shutdown:
					bc.logger.Info("Worker de reset diário encerrado")
					return
					
				case <-ticker.C:
					bc.logger.Info("Realizando reset diário de consultas")
					bc.resetConsultasDiarias()
					bc.observability.RecordMetric(marketContext, "bureau_credito_daily_reset", "reset", 1)
				}
			}
		}
	}
}

// Stop encerra o serviço Bureau de Crédito
func (bc *BureauCredito) Stop() {
	bc.logger.Info("Parando serviço Bureau de Crédito")
	
	// Notificar workers para encerrar
	close(bc.shutdown)
	
	// Aguardar todos os workers encerrarem
	bc.wg.Wait()
	
	// Registrar métrica de encerramento do serviço
	marketContext := adapter.MarketContext{
		Market:     bc.config.Market,
		TenantType: bc.config.TenantType,
	}
	bc.observability.RecordMetric(marketContext, "bureau_credito_service_stopped", "stop", 1)
	
	bc.logger.Info("Serviço Bureau de Crédito encerrado")
}

// ConfigurarRegrasCompliancePadrao configura regras de compliance padrão para diversos mercados
func (bc *BureauCredito) ConfigurarRegrasCompliancePadrao() {
	// Regra global para validação de campos obrigatórios
	bc.RegistrarRegraCompliance(RegrasCompliance{
		ID:           "COMPLIANCE-G-001",
		Market:       constants.MarketGlobal,
		Description:  "Validação de campos obrigatórios por tipo de consulta",
		Framework:    []string{"ISO 27001", "TOGAF", "PCI DSS"},
		MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaScore), string(ConsultaBasica), 
			string(ConsultaRestricoes), string(ConsultaHistorico), string(ConsultaRelacionamento)},
		Validate: func(consulta *ConsultaCredito) (bool, string, error) {
			campos, existe := bc.config.CamposObrigatorios[string(consulta.TipoConsulta)]
			if !existe {
				return true, "Sem requisitos específicos de campos para este tipo de consulta", nil
			}
			
			for _, campo := range campos {
				switch campo {
				case "documentoCliente":
					if consulta.DocumentoCliente == "" {
						return false, "Documento do cliente é obrigatório", nil
					}
				case "nomeCliente":
					if consulta.NomeCliente == "" {
						return false, "Nome do cliente é obrigatório", nil
					}
				case "finalidade":
					if consulta.Finalidade == "" {
						return false, "Finalidade é obrigatória", nil
					}
				case "solicitanteID":
					if consulta.SolicitanteID == "" {
						return false, "ID do solicitante é obrigatório", nil
					}
				case "consentimentoID":
					if consulta.ConsentimentoID == "" {
						return false, "ID de consentimento é obrigatório", nil
					}
				}
			}
			
			return true, "Todos os campos obrigatórios estão presentes", nil
		},
	})

	// Regra Angola - BNA (Banco Nacional de Angola)
	bc.RegistrarRegraCompliance(RegrasCompliance{
		ID:           "COMPLIANCE-AO-001",
		Market:       constants.MarketAngola,
		Description:  "Validação de regras BNA para consultas completas",
		Framework:    []string{"BNA", "ISO 27001"},
		MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico)},
		Validate: func(consulta *ConsultaCredito) (bool, string, error) {
			// Verificar requisitos específicos do BNA
			if consulta.Finalidade == FinalidadeConcessaoCredito && consulta.ConsentimentoID == "" {
				return false, "BNA exige consentimento explícito para concessão de crédito", nil
			}
			
			// Verificação adicional de MFA para consultas completas
			if consulta.TipoConsulta == ConsultaCompleta && consulta.MFALevel != "high" {
				return false, "BNA exige MFA de nível alto para consultas completas", nil
			}
			
			return true, "Requisitos BNA atendidos", nil
		},
	})

	// Regra Brasil - BACEN e LGPD
	bc.RegistrarRegraCompliance(RegrasCompliance{
		ID:           "COMPLIANCE-BR-001",
		Market:       constants.MarketBrazil,
		Description:  "Validação de regras BACEN e LGPD",
		Framework:    []string{"BACEN", "LGPD", "SCR"},
		MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico)},
		Validate: func(consulta *ConsultaCredito) (bool, string, error) {
			// Verificar requisitos específicos do BACEN/LGPD
			if consulta.ConsentimentoID == "" {
				return false, "LGPD exige consentimento explícito para qualquer consulta de dados pessoais", nil
			}
			
			// Verificar finalidade específica
			if consulta.Finalidade == "" {
				return false, "BACEN exige especificação clara da finalidade da consulta", nil
			}
			
			return true, "Requisitos BACEN/LGPD atendidos", nil
		},
	})

	// Regra União Europeia - GDPR
	bc.RegistrarRegraCompliance(RegrasCompliance{
		ID:           "COMPLIANCE-EU-001",
		Market:       constants.MarketEU,
		Description:  "Validação de regras GDPR para consultas de crédito",
		Framework:    []string{"GDPR", "PSD2"},
		MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico), string(ConsultaScore)},
		Validate: func(consulta *ConsultaCredito) (bool, string, error) {
			// Verificar requisitos específicos do GDPR
			if consulta.ConsentimentoID == "" {
				return false, "GDPR exige consentimento explícito para processamento de dados pessoais", nil
			}
			
			// Verificar finalidade específica e limitada
			if consulta.Finalidade == "" {
				return false, "GDPR exige finalidade específica e limitada para processamento de dados", nil
			}
			
			// Verificar se há parâmetros de minimização de dados
			if consulta.Parametros == nil || len(consulta.Parametros) == 0 {
				return false, "GDPR exige parâmetros para minimização de dados", nil
			}
			
			// Verificar parâmetro específico de minimização
			_, temMinimizacao := consulta.Parametros["minimizacaoDados"]
			if !temMinimizacao {
				return false, "GDPR exige especificação de minimização de dados", nil
			}
			
			return true, "Requisitos GDPR atendidos", nil
		},
	})

	// Regra EUA - FCRA (Fair Credit Reporting Act)
	bc.RegistrarRegraCompliance(RegrasCompliance{
		ID:           "COMPLIANCE-US-001",
		Market:       constants.MarketUSA,
		Description:  "Validação de regras FCRA para consultas de crédito",
		Framework:    []string{"FCRA", "GLBA"},
		MandatoryFor: []string{string(ConsultaCompleta), string(ConsultaHistorico), string(ConsultaScore)},
		Validate: func(consulta *ConsultaCredito) (bool, string, error) {
			// Verificar requisitos específicos do FCRA
			if consulta.Finalidade == "" {
				return false, "FCRA exige finalidade permissível para consulta de crédito", nil
			}
			
			// Verificar finalidades permissíveis conforme FCRA
			finalidadesPermissiveis := []FinalidadeConsulta{
				FinalidadeConcessaoCredito, 
				FinalidadeRevisaoLimites,
				FinalidadeVerificacaoCliente,
				FinalidadeAberturaConta,
			}
			
			finalidadePermitida := false
			for _, fp := range finalidadesPermissiveis {
				if consulta.Finalidade == fp {
					finalidadePermitida = true
					break
				}
			}
			
			if !finalidadePermitida {
				return false, "FCRA exige finalidade permissível específica para consulta", nil
			}
			
			return true, "Requisitos FCRA atendidos", nil
		},
	})
}

// main é o ponto de entrada do programa
func main() {
	// Configurar logger
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalf("Erro ao criar logger: %v", err)
	}
	defer logger.Sync()

	// Obter variáveis de ambiente
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	market := os.Getenv("MARKET")
	if market == "" {
		market = constants.MarketGlobal
	}

	tenantType := os.Getenv("TENANT_TYPE")
	if tenantType == "" {
		tenantType = "default"
	}

	// Configurar observabilidade
	observability, err := adapter.NewIAMObservability(adapter.ObservabilityConfig{
		ServiceName:      "bureau-credito",
		ServiceVersion:   os.Getenv("SERVICE_VERSION"),
		Environment:      environment,
		DefaultMarket:    market,
		DefaultTenantID:  "global",
		DefaultTenantType: tenantType,
	})
	if err != nil {
		logger.Fatal("Falha ao inicializar observabilidade", zap.Error(err))
	}
	defer observability.Shutdown(context.Background())

	// Configurar Bureau de Crédito
	config := BureauCreditoConfig{
		Market:                 market,
		TenantType:             tenantType,
		Environment:            environment,
		TempoRetencaoHistorico: map[string]int{
			constants.MarketAngola: 60,  // 5 anos
			constants.MarketBrazil: 60,  // 5 anos
			constants.MarketEU:     24,  // 2 anos (GDPR)
			constants.MarketUSA:    84,  // 7 anos
			constants.MarketGlobal: 60,  // Padrão global
		},
		FontesDadosAtivas: []string{
			"bancos", "telecom", "varejo", "servicos", "utilidades",
		},
		LimiteConsultasDiarias: 100,
		ConsentimentoObrigatorio: map[string]bool{
			constants.MarketAngola: true,
			constants.MarketBrazil: true,
			constants.MarketEU:     true,
			constants.MarketUSA:    false,
			constants.MarketGlobal: true,
		},
		ValidadeConsentimento: map[string]int{
			string(FinalidadeConcessaoCredito):   90,  // 90 dias
			string(FinalidadeRevisaoLimites):     30,  // 30 dias
			string(FinalidadeGerencialRegulador): 365, // 1 ano
			string(FinalidadeVerificacaoCliente): 7,   // 7 dias
		},
		NotificacaoObrigatoria: map[string]bool{
			constants.MarketAngola: true,
			constants.MarketBrazil: true,
			constants.MarketEU:     true,
			constants.MarketUSA:    true,
			constants.MarketGlobal: true,
		},
		CamposObrigatorios: map[string][]string{
			string(ConsultaCompleta): {"documentoCliente", "nomeCliente", "finalidade", "solicitanteID", "consentimentoID"},
			string(ConsultaScore):    {"documentoCliente", "finalidade", "solicitanteID"},
			string(ConsultaBasica):   {"documentoCliente", "finalidade"},
		},
	}

	// Criar instância do Bureau de Crédito
	bureau := NewBureauCredito(config, observability, logger)

	// Configurar regras de compliance padrão
	bureau.ConfigurarRegrasCompliancePadrao()

	// Iniciar o serviço
	if err := bureau.Start(); err != nil {
		logger.Fatal("Falha ao iniciar serviço Bureau de Crédito", zap.Error(err))
	}

	// Configurar canal para sinais de término
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Aguardar sinal para encerrar
	sig := <-sigChan
	logger.Info("Sinal de encerramento recebido", zap.String("signal", sig.String()))

	// Encerrar o serviço
	bureau.Stop()
}