"""
Validador de Conformidade para África do Sul - BRICS

Este módulo implementa o validador de conformidade específico para a África do Sul
dentro do contexto BRICS, considerando POPIA, FICA, SARB e regulamentações
específicas do mercado financeiro sul-africano.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, Any, List

class SouthAfricaComplianceValidator:
    """
    Validador de conformidade para regulamentações específicas da África do Sul.
    """
    
    def __init__(self, logger=None, metrics=None):
        """
        Inicializa o validador de conformidade para a África do Sul.
        
        Args:
            logger: Logger para registro de eventos
            metrics: Serviço de métricas
        """
        self.logger = logger
        self.metrics = metrics
        self.requisitos = self._carregar_requisitos()
    
    def validar(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações específicas da África do Sul.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        dados_pessoais = contexto.get("dados_pessoais", {})
        dados_financeiros = contexto.get("dados_financeiros", {})
        
        if self.logger:
            self.logger.info("Iniciando validação de conformidade para África do Sul")
            
        if self.metrics:
            self.metrics.incrementCounter("trustguard.compliance.brics.southafrica.validation")
        
        # Validar Lei de Proteção de Informações Pessoais (POPIA)
        self._validar_popia(resultado, contexto, dados_pessoais)
        
        # Validar Lei de Centro de Inteligência Financeira (FICA)
        self._validar_fica(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações do Banco de Reserva da África do Sul
        self._validar_sarb(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações da Autoridade de Conduta do Setor Financeiro
        self._validar_fsca(resultado, contexto, dados_financeiros)
        
        return resultado
    
    def _carregar_requisitos(self) -> Dict[str, Any]:
        """
        Carrega os requisitos de conformidade específicos para a África do Sul.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "popia": {
                "nome": "Lei de Proteção de Informações Pessoais",
                "codigo": "POPIA",
                "consentimento_explicito": True,
                "direito_acesso": True,
                "direito_correcao": True,
                "direito_exclusao": True,
                "restricao_transferencia_transfronteirica": True,
                "notificacao_violacao": True
            },
            "fica": {
                "nome": "Lei do Centro de Inteligência Financeira",
                "codigo": "FICA",
                "verificacao_identidade": True,
                "limite_reportavel": 25000,  # ZAR
                "identificacao_beneficiario_final": True,
                "verificacao_reforçada": ["pep", "alto_risco", "transacao_grande"],
                "atualizacao_periodica": True
            },
            "sarb": {
                "nome": "Banco de Reserva da África do Sul",
                "codigo": "SARB",
                "limites_cambio": {
                    "individual": 1000000,  # ZAR por ano
                    "empresarial": 1000000000  # ZAR por ano
                },
                "declaracao_obrigatoria": True,
                "sistema_pagamentos": ["rtgs", "eft", "sepa", "swift"]
            },
            "fsca": {
                "nome": "Autoridade de Conduta do Setor Financeiro",
                "codigo": "FSCA",
                "licenca_obrigatoria": True,
                "tratamento_justo_clientes": True,
                "transparencia_custos": True,
                "resolucao_disputas": True
            }
        }
    
    def _validar_popia(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_pessoais: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei de Proteção de Informações Pessoais.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_pessoais: Dados pessoais do contexto
        """
        if not dados_pessoais:
            return
            
        popia = self.requisitos["popia"]
        
        # Verificar consentimento explícito
        if popia["consentimento_explicito"] and not dados_pessoais.get("consentimento_explicito_za", False):
            self._registrar_violacao(
                resultado,
                "ZA_POPIA_FALTA_CONSENTIMENTO",
                "POPIA requer consentimento explícito para processamento de dados pessoais",
                "alta",
                "bloqueante",
                [popia["codigo"]]
            )
        
        # Verificar mecanismos para direitos do titular
        direitos = ["direito_acesso", "direito_correcao", "direito_exclusao"]
        for direito in direitos:
            if popia[direito] and not contexto.get(f"mecanismo_{direito}_za", False):
                nome_direito = direito.replace("direito_", "").capitalize()
                self._registrar_violacao(
                    resultado,
                    f"ZA_POPIA_FALTA_MECANISMO_{direito.upper()}",
                    f"POPIA requer mecanismo para {nome_direito} de dados pelos titulares",
                    "media",
                    "aviso",
                    [popia["codigo"]]
                )
        
        # Verificar transferência transfronteiriça
        if popia["restricao_transferencia_transfronteirica"] and contexto.get("transferencia_internacional_dados", False):
            if not contexto.get("base_legal_transferencia_za", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_POPIA_TRANSFERENCIA_TRANSFRONTEIRICA",
                    "POPIA requer base legal para transferência transfronteiriça de dados",
                    "alta",
                    "bloqueante",
                    [popia["codigo"]]
                )
        
        # Verificar notificação de violação
        if popia["notificacao_violacao"] and contexto.get("violacao_dados", False):
            if not contexto.get("notificacao_inforegulator_za", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_POPIA_FALTA_NOTIFICACAO_VIOLACAO",
                    "POPIA exige notificação de violações de dados ao Information Regulator",
                    "alta",
                    "bloqueante",
                    [popia["codigo"]]
                )
    
    def _validar_fica(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei do Centro de Inteligência Financeira.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        fica = self.requisitos["fica"]
        
        # Verificar verificação de identidade
        if fica["verificacao_identidade"] and not contexto.get("verificacao_identidade_za", False):
            self._registrar_violacao(
                resultado,
                "ZA_FICA_FALTA_VERIFICACAO_IDENTIDADE",
                "FICA exige verificação de identidade para clientes",
                "alta",
                "bloqueante",
                [fica["codigo"]]
            )
        
        # Verificar identificação de beneficiário final
        if fica["identificacao_beneficiario_final"] and contexto.get("tipo_cliente") == "pessoa_juridica":
            if not contexto.get("identificacao_beneficiario_final_za", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_FICA_FALTA_BENEFICIARIO_FINAL",
                    "FICA exige identificação de beneficiários finais para pessoas jurídicas",
                    "alta",
                    "bloqueante",
                    [fica["codigo"]]
                )
        
        # Verificar limite reportável
        if "valor_transacao" in dados_financeiros and dados_financeiros.get("moeda") == "ZAR":
            try:
                valor = float(dados_financeiros["valor_transacao"])
                
                if valor >= fica["limite_reportavel"] and not dados_financeiros.get("reportado_fic_za", False):
                    self._registrar_violacao(
                        resultado,
                        "ZA_FICA_TRANSACAO_NAO_REPORTADA",
                        f"Transações acima de {fica['limite_reportavel']} ZAR devem ser reportadas ao FIC",
                        "alta",
                        "bloqueante",
                        [fica["codigo"]]
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar verificação reforçada
        for categoria in fica["verificacao_reforçada"]:
            if contexto.get(categoria, False) and not contexto.get(f"verificacao_reforcada_{categoria}_za", False):
                descricao_categoria = {
                    "pep": "Pessoas Politicamente Expostas",
                    "alto_risco": "clientes de alto risco",
                    "transacao_grande": "transações de grande valor"
                }.get(categoria, categoria)
                
                self._registrar_violacao(
                    resultado,
                    f"ZA_FICA_FALTA_VERIFICACAO_REFORCADA_{categoria.upper()}",
                    f"FICA exige verificação reforçada para {descricao_categoria}",
                    "alta",
                    "bloqueante",
                    [fica["codigo"]]
                )
    
    def _validar_sarb(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações do Banco de Reserva da África do Sul.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        sarb = self.requisitos["sarb"]
        
        # Verificar limites de câmbio
        if dados_financeiros.get("tipo_transacao") == "cambio" and "valor_transacao" in dados_financeiros:
            try:
                valor = float(dados_financeiros["valor_transacao"])
                tipo_cliente = contexto.get("tipo_cliente", "individual")
                valor_acumulado = float(dados_financeiros.get("valor_acumulado_anual", 0))
                
                limite_aplicavel = sarb["limites_cambio"]["individual"]
                if tipo_cliente == "pessoa_juridica" or tipo_cliente == "empresarial":
                    limite_aplicavel = sarb["limites_cambio"]["empresarial"]
                
                if valor_acumulado + valor > limite_aplicavel and not dados_financeiros.get("autorizacao_especial_sarb", False):
                    self._registrar_violacao(
                        resultado,
                        "ZA_SARB_LIMITE_CAMBIO_EXCEDIDO",
                        f"Valor excede o limite de {limite_aplicavel} ZAR para operações de câmbio anuais",
                        "alta",
                        "bloqueante",
                        [sarb["codigo"]]
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar declaração obrigatória
        if sarb["declaracao_obrigatoria"] and dados_financeiros.get("tipo_transacao") == "internacional":
            if not dados_financeiros.get("declaracao_sarb", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_SARB_FALTA_DECLARACAO",
                    "SARB exige declaração para transferências internacionais",
                    "alta",
                    "bloqueante",
                    [sarb["codigo"]]
                )
        
        # Verificar uso de sistemas de pagamento aprovados
        if "sistema_pagamento" in dados_financeiros:
            sistema = dados_financeiros["sistema_pagamento"].lower()
            if sistema not in sarb["sistema_pagamentos"]:
                self._registrar_violacao(
                    resultado,
                    "ZA_SARB_SISTEMA_PAGAMENTO_NAO_APROVADO",
                    f"Sistema de pagamento '{sistema}' não é aprovado pelo SARB",
                    "alta",
                    "bloqueante",
                    [sarb["codigo"]]
                )
    
    def _validar_fsca(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações da Autoridade de Conduta do Setor Financeiro.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        fsca = self.requisitos["fsca"]
        
        # Verificar licença obrigatória para provedores de serviços financeiros
        if fsca["licenca_obrigatoria"] and contexto.get("tipo_entidade") == "provedor_servico_financeiro":
            if not contexto.get("licenca_fsca", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_FSCA_FALTA_LICENCA",
                    "FSCA exige licença para provedores de serviços financeiros",
                    "alta",
                    "bloqueante",
                    [fsca["codigo"]]
                )
        
        # Verificar tratamento justo de clientes
        if fsca["tratamento_justo_clientes"]:
            if not contexto.get("tcf_implementado_za", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_FSCA_FALTA_TCF",
                    "FSCA exige implementação dos princípios de Tratamento Justo de Clientes (TCF)",
                    "media",
                    "aviso",
                    [fsca["codigo"]]
                )
        
        # Verificar transparência de custos
        if fsca["transparencia_custos"] and "produto_financeiro" in dados_financeiros:
            if not dados_financeiros.get("divulgacao_custos_completa", False):
                self._registrar_violacao(
                    resultado,
                    "ZA_FSCA_FALTA_TRANSPARENCIA_CUSTOS",
                    "FSCA exige transparência total dos custos para produtos financeiros",
                    "media",
                    "aviso",
                    [fsca["codigo"]]
                )
        
        # Verificar mecanismo de resolução de disputas
        if fsca["resolucao_disputas"] and not contexto.get("mecanismo_resolucao_disputas_za", False):
            self._registrar_violacao(
                resultado,
                "ZA_FSCA_FALTA_RESOLUCAO_DISPUTAS",
                "FSCA exige mecanismo de resolução de disputas para clientes",
                "media",
                "aviso",
                [fsca["codigo"]]
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
                f"Violação de conformidade África do Sul: {codigo} - {descricao} "
                f"(severidade: {severidade}, impacto: {impacto})"
            )
        
        # Métricas
        if self.metrics:
            self.metrics.incrementCounter(
                "trustguard.compliance.brics.southafrica.violation", 
                {"code": codigo, "severity": severidade, "impact": impacto}
            )
