"""
Validador de Conformidade para Brasil - BRICS

Este módulo implementa o validador de conformidade específico para o Brasil
dentro do contexto BRICS, considerando LGPD, regulamentações do BACEN, CVM e COAF.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, Any, List

class BrazilComplianceValidator:
    """
    Validador de conformidade para regulamentações específicas do Brasil.
    """
    
    def __init__(self, logger=None, metrics=None):
        """
        Inicializa o validador de conformidade para o Brasil.
        
        Args:
            logger: Logger para registro de eventos
            metrics: Serviço de métricas
        """
        self.logger = logger
        self.metrics = metrics
        self.requisitos = self._carregar_requisitos()
    
    def validar(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações específicas do Brasil.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        dados_pessoais = contexto.get("dados_pessoais", {})
        dados_financeiros = contexto.get("dados_financeiros", {})
        
        if self.logger:
            self.logger.info("Iniciando validação de conformidade para Brasil")
            
        if self.metrics:
            self.metrics.incrementCounter("trustguard.compliance.brics.brazil.validation")
        
        # Validar LGPD
        self._validar_lgpd(resultado, contexto, dados_pessoais)
        
        # Validar regulamentações do BACEN
        self._validar_bacen(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações da CVM
        self._validar_cvm(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações do COAF
        self._validar_coaf(resultado, contexto, dados_financeiros)
        
        return resultado
    
    def _carregar_requisitos(self) -> Dict[str, Any]:
        """
        Carrega os requisitos de conformidade específicos para o Brasil.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "lgpd": {
                "nome": "Lei Geral de Proteção de Dados",
                "codigo": "Lei nº 13.709/2018",
                "idade_minima_consentimento": 16,
                "requer_consentimento_explicito": True,
                "direitos_titulares": ["acesso", "retificacao", "exclusao", "portabilidade", "informacao"]
            },
            "bacen": {
                "nome": "Regulamentações do Banco Central do Brasil",
                "requer_kyc": True,
                "limites_transacao": {
                    "pix": 100000.0,  # BRL
                    "pix_noturno": 1000.0,  # BRL
                    "ted": 1000000.0,  # BRL
                    "internacional": 10000.0  # USD
                }
            },
            "cvm": {
                "nome": "Comissão de Valores Mobiliários",
                "requer_suitability": True
            },
            "coaf": {
                "nome": "Conselho de Controle de Atividades Financeiras",
                "limite_reportavel": 50000.0  # BRL
            }
        }
    
    def _validar_lgpd(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_pessoais: Dict[str, Any]) -> None:
        """
        Valida conformidade com LGPD.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_pessoais: Dados pessoais do contexto
        """
        if not dados_pessoais:
            return
            
        lgpd = self.requisitos["lgpd"]
        
        # Verificar idade para consentimento
        if "idade" in dados_pessoais and dados_pessoais["idade"] < lgpd["idade_minima_consentimento"]:
            if not dados_pessoais.get("consentimento_responsavel", False):
                self._registrar_violacao(
                    resultado,
                    "LGPD_FALTA_CONSENTIMENTO_RESPONSAVEL",
                    f"LGPD requer consentimento do responsável legal para menores de {lgpd['idade_minima_consentimento']} anos",
                    "alta",
                    "bloqueante",
                    [lgpd["codigo"]]
                )
        
        # Verificar consentimento explícito para dados sensíveis
        if dados_pessoais.get("dados_sensiveis", False) and not dados_pessoais.get("consentimento_explicito", False):
            self._registrar_violacao(
                resultado,
                "LGPD_FALTA_CONSENTIMENTO_EXPLICITO",
                "LGPD requer consentimento explícito para tratamento de dados sensíveis",
                "alta",
                "bloqueante",
                [lgpd["codigo"]]
            )
        
        # Verificar finalidade específica
        if not dados_pessoais.get("finalidade_especifica", False):
            self._registrar_violacao(
                resultado,
                "LGPD_FALTA_FINALIDADE",
                "LGPD requer finalidade específica para tratamento de dados pessoais",
                "alta",
                "aviso",
                [lgpd["codigo"]]
            )
        
        # Verificar mecanismo para exercício de direitos
        if not contexto.get("mecanismo_exercicio_direitos", False):
            self._registrar_violacao(
                resultado,
                "LGPD_FALTA_MECANISMO_DIREITOS",
                "LGPD requer mecanismo para exercício de direitos dos titulares",
                "media",
                "aviso",
                [lgpd["codigo"]]
            )
    
    def _validar_bacen(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações do BACEN.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        bacen = self.requisitos["bacen"]
        
        # Verificar KYC
        if bacen["requer_kyc"] and not dados_financeiros.get("kyc_completo", False):
            self._registrar_violacao(
                resultado,
                "BACEN_KYC_INCOMPLETO",
                "Regulamentações do BACEN requerem processo KYC completo",
                "alta",
                "bloqueante"
            )
        
        # Verificar limites de transação
        if "tipo_transferencia" in dados_financeiros and "valor_transacao" in dados_financeiros:
            tipo = dados_financeiros["tipo_transferencia"]
            try:
                valor = float(dados_financeiros["valor_transacao"])
                
                if tipo in bacen["limites_transacao"] and valor > bacen["limites_transacao"][tipo]:
                    self._registrar_violacao(
                        resultado,
                        "BACEN_LIMITE_EXCEDIDO",
                        f"Valor excede o limite de {bacen['limites_transacao'][tipo]} para transferências {tipo}",
                        "alta",
                        "bloqueante"
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar requisitos para transações internacionais
        if dados_financeiros.get("tipo_transferencia") == "internacional":
            if not dados_financeiros.get("registro_bacen", False):
                self._registrar_violacao(
                    resultado,
                    "BACEN_FALTA_REGISTRO",
                    "Transferências internacionais requerem registro no BACEN",
                    "alta",
                    "bloqueante"
                )
            
            if not dados_financeiros.get("justificativa_economica", False):
                self._registrar_violacao(
                    resultado,
                    "BACEN_FALTA_JUSTIFICATIVA",
                    "Transferências internacionais requerem justificativa econômica",
                    "alta",
                    "bloqueante"
                )
    
    def _validar_cvm(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações da CVM.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros or "tipo_operacao" not in dados_financeiros:
            return
            
        # Verificar se a operação é de investimento
        if "investimento" not in dados_financeiros["tipo_operacao"].lower():
            return
            
        cvm = self.requisitos["cvm"]
        
        # Verificar suitability
        if cvm["requer_suitability"] and not dados_financeiros.get("suitability_realizado", False):
            self._registrar_violacao(
                resultado,
                "CVM_FALTA_SUITABILITY",
                "CVM requer análise de suitability para operações de investimento",
                "alta",
                "bloqueante",
                [cvm["nome"]]
            )
        
        # Verificar informações do investidor
        if not dados_financeiros.get("perfil_investidor", False):
            self._registrar_violacao(
                resultado,
                "CVM_FALTA_PERFIL",
                "CVM requer perfil de investidor definido",
                "alta",
                "aviso",
                [cvm["nome"]]
            )
        
        # Verificar adequação do produto ao perfil
        if dados_financeiros.get("perfil_investidor") and dados_financeiros.get("risco_produto"):
            perfil = dados_financeiros["perfil_investidor"].lower()
            risco = dados_financeiros["risco_produto"].lower()
            
            if (perfil == "conservador" and risco in ["moderado", "arrojado", "agressivo"]) or \
               (perfil == "moderado" and risco in ["arrojado", "agressivo"]):
                if not dados_financeiros.get("termo_ciencia_risco", False):
                    self._registrar_violacao(
                        resultado,
                        "CVM_FALTA_TERMO_CIENCIA",
                        "CVM requer termo de ciência de risco quando o produto não é adequado ao perfil do investidor",
                        "alta",
                        "bloqueante",
                        [cvm["nome"]]
                    )
    
    def _validar_coaf(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações do COAF.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        coaf = self.requisitos["coaf"]
        
        # Verificar transações reportáveis
        if "valor_transacao" in dados_financeiros and dados_financeiros.get("moeda") == "BRL":
            try:
                valor = float(dados_financeiros["valor_transacao"])
                if valor >= coaf["limite_reportavel"] and not dados_financeiros.get("reportado_coaf", False):
                    self._registrar_violacao(
                        resultado,
                        "COAF_NAO_REPORTADO",
                        f"Transações acima de {coaf['limite_reportavel']} BRL devem ser reportadas ao COAF",
                        "alta",
                        "aviso",
                        [coaf["nome"]]
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar operações suspeitas
        if dados_financeiros.get("indicadores_suspeita", []):
            if not dados_financeiros.get("analise_compliance", False):
                self._registrar_violacao(
                    resultado,
                    "COAF_FALTA_ANALISE_COMPLIANCE",
                    "Operações com indicadores de suspeita requerem análise de compliance",
                    "alta",
                    "bloqueante",
                    [coaf["nome"]]
                )
        
        # Verificar cliente PEP (Pessoa Politicamente Exposta)
        if contexto.get("pessoa_pep", False) and not dados_financeiros.get("monitoramento_reforçado_pep", False):
            self._registrar_violacao(
                resultado,
                "COAF_FALTA_MONITORAMENTO_PEP",
                "Clientes PEP requerem monitoramento reforçado conforme diretrizes do COAF",
                "alta",
                "bloqueante",
                [coaf["nome"]]
            )
    
    def _registrar_violacao(self, resultado: Dict[str, Any], codigo: str, descricao: str, 
                           severidade: str, impacto: str, normas: List[str] = None) -> None:
        """
        Registra uma violação de conformidade.
        
        Args:
            resultado: Resultado da validação para atualizar
            codigo: Código da violação
            descricao: Descrição da violação
            severidade: Severidade da violação (alta, media, baixa)
            impacto: Impacto da violação (bloqueante, aviso)
            normas: Lista de normas relacionadas à violação
        """
        # Adicionar violação
        violacao = {
            "codigo": codigo,
            "descricao": descricao,
            "severidade": severidade,
            "impacto": impacto
        }
        
        if normas:
            violacao["normas"] = normas
        
        # Adicionar ao resultado
        if impacto == "bloqueante":
            resultado["valido"] = False
            resultado["violacoes"].append(violacao)
            
            # Reduzir score de conformidade
            redução = 0.0
            if severidade == "alta":
                redução = 0.3
            elif severidade == "media":
                redução = 0.2
            else:
                redução = 0.1
                
            resultado["score_conformidade"] = max(0.0, resultado["score_conformidade"] - redução)
        else:
            resultado["avisos"].append(violacao)
        
        # Log
        if self.logger:
            self.logger.warning(
                f"Violação de conformidade Brasil: {codigo} - {descricao} "
                f"(severidade: {severidade}, impacto: {impacto})"
            )
        
        # Métricas
        if self.metrics:
            self.metrics.incrementCounter(
                "trustguard.compliance.brics.brazil.violation", 
                {"code": codigo, "severity": severidade, "impact": impacto}
            )
