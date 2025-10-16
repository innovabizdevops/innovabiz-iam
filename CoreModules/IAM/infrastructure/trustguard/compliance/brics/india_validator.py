"""
Validador de Conformidade para Índia - BRICS

Este módulo implementa o validador de conformidade específico para a Índia
dentro do contexto BRICS, considerando PDPB, FEMA, RBI e regulamentações
específicas do mercado financeiro indiano.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, Any, List

class IndiaComplianceValidator:
    """
    Validador de conformidade para regulamentações específicas da Índia.
    """
    
    def __init__(self, logger=None, metrics=None):
        """
        Inicializa o validador de conformidade para a Índia.
        
        Args:
            logger: Logger para registro de eventos
            metrics: Serviço de métricas
        """
        self.logger = logger
        self.metrics = metrics
        self.requisitos = self._carregar_requisitos()
    
    def validar(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações específicas da Índia.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        dados_pessoais = contexto.get("dados_pessoais", {})
        dados_financeiros = contexto.get("dados_financeiros", {})
        
        if self.logger:
            self.logger.info("Iniciando validação de conformidade para Índia")
            
        if self.metrics:
            self.metrics.incrementCounter("trustguard.compliance.brics.india.validation")
        
        # Validar Lei de Proteção de Dados Pessoais
        self._validar_pdpb(resultado, contexto, dados_pessoais)
        
        # Validar FEMA (Lei de Gestão de Câmbio Estrangeiro)
        self._validar_fema(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações do Banco de Reserva da Índia
        self._validar_rbi(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações de prevenção à lavagem de dinheiro
        self._validar_pmla(resultado, contexto, dados_financeiros)
        
        return resultado
    
    def _carregar_requisitos(self) -> Dict[str, Any]:
        """
        Carrega os requisitos de conformidade específicos para a Índia.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "pdpb": {
                "nome": "Lei de Proteção de Dados Pessoais",
                "codigo": "PDPB",
                "consentimento_explicito": True,
                "direito_esquecimento": True,
                "restricao_dados_sensiveis": True,
                "localizacao_dados": True,
                "notificacao_violacao": True
            },
            "fema": {
                "nome": "Lei de Gestão de Câmbio Estrangeiro",
                "codigo": "FEMA 1999",
                "limite_remessa_internacional": 250000,  # USD por ano fiscal
                "requer_declaracao": True,
                "restricoes_por_pais": {
                    "SN": True,  # Países sob sanções
                    "IR": True,
                    "KP": True
                }
            },
            "rbi": {
                "nome": "Banco de Reserva da Índia",
                "codigo": "RBI",
                "limites_transacao": {
                    "upi": 100000,  # INR
                    "imps": 500000,  # INR
                    "neft": 1000000,  # INR
                    "rtgs": 200000  # INR mínimo
                },
                "kyc_obrigatorio": True,
                "monitoramento_transacao": True
            },
            "pmla": {
                "nome": "Lei de Prevenção à Lavagem de Dinheiro",
                "codigo": "PMLA 2002",
                "limite_reportavel": 1000000,  # INR
                "requer_verificacao_identidade": True,
                "monitoramento_pep": True,
                "manutencao_registros": True,
                "periodo_retencao": 5  # anos
            }
        }
    
    def _validar_pdpb(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_pessoais: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei de Proteção de Dados Pessoais da Índia.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_pessoais: Dados pessoais do contexto
        """
        if not dados_pessoais:
            return
            
        pdpb = self.requisitos["pdpb"]
        
        # Verificar consentimento explícito
        if pdpb["consentimento_explicito"] and not dados_pessoais.get("consentimento_explicito_in", False):
            self._registrar_violacao(
                resultado,
                "IN_PDPB_FALTA_CONSENTIMENTO",
                "PDPB requer consentimento explícito para processamento de dados pessoais",
                "alta",
                "bloqueante",
                [pdpb["codigo"]]
            )
        
        # Verificar tratamento de dados sensíveis
        if pdpb["restricao_dados_sensiveis"] and dados_pessoais.get("dados_sensiveis", False):
            if not dados_pessoais.get("consentimento_explicito_dados_sensiveis_in", False):
                self._registrar_violacao(
                    resultado,
                    "IN_PDPB_FALTA_CONSENTIMENTO_DADOS_SENSIVEIS",
                    "PDPB requer consentimento explícito para processamento de dados sensíveis",
                    "alta",
                    "bloqueante",
                    [pdpb["codigo"]]
                )
        
        # Verificar localização de dados
        if pdpb["localizacao_dados"] and not contexto.get("dados_armazenados_india", False):
            if dados_pessoais.get("dados_criticos", False):
                self._registrar_violacao(
                    resultado,
                    "IN_PDPB_DADOS_CRITICOS_FORA_TERRITORIO",
                    "PDPB exige que dados pessoais críticos sejam armazenados apenas na Índia",
                    "alta",
                    "bloqueante",
                    [pdpb["codigo"]]
                )
            elif not contexto.get("copia_dados_india", False):
                self._registrar_violacao(
                    resultado,
                    "IN_PDPB_FALTA_COPIA_LOCAL",
                    "PDPB exige que uma cópia dos dados pessoais seja mantida na Índia",
                    "alta",
                    "aviso",
                    [pdpb["codigo"]]
                )
        
        # Verificar direito ao esquecimento
        if pdpb["direito_esquecimento"] and not contexto.get("mecanismo_direito_esquecimento", False):
            self._registrar_violacao(
                resultado,
                "IN_PDPB_FALTA_MECANISMO_ESQUECIMENTO",
                "PDPB exige mecanismo para exercício do direito ao esquecimento",
                "media",
                "aviso",
                [pdpb["codigo"]]
            )
        
        # Verificar notificação de violação
        if pdpb["notificacao_violacao"] and contexto.get("violacao_dados", False):
            if not contexto.get("notificacao_dpa_in", False):
                self._registrar_violacao(
                    resultado,
                    "IN_PDPB_FALTA_NOTIFICACAO_VIOLACAO",
                    "PDPB exige notificação de violações de dados à Autoridade de Proteção de Dados",
                    "alta",
                    "bloqueante",
                    [pdpb["codigo"]]
                )
    
    def _validar_fema(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei de Gestão de Câmbio Estrangeiro.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        fema = self.requisitos["fema"]
        
        # Verificar limite de remessa internacional
        if dados_financeiros.get("tipo_transacao") == "remessa_internacional" and dados_financeiros.get("moeda") == "USD":
            try:
                valor = float(dados_financeiros.get("valor_transacao", 0))
                valor_acumulado_anual = float(dados_financeiros.get("valor_acumulado_anual", 0))
                
                if valor_acumulado_anual + valor > fema["limite_remessa_internacional"]:
                    self._registrar_violacao(
                        resultado,
                        "IN_FEMA_LIMITE_REMESSA_EXCEDIDO",
                        f"FEMA limita remessas internacionais a {fema['limite_remessa_internacional']} USD por ano fiscal",
                        "alta",
                        "bloqueante",
                        [fema["codigo"]]
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar declaração para remessas
        if fema["requer_declaracao"] and dados_financeiros.get("tipo_transacao") == "remessa_internacional":
            if not dados_financeiros.get("declaracao_fema_a2", False):
                self._registrar_violacao(
                    resultado,
                    "IN_FEMA_FALTA_DECLARACAO",
                    "FEMA exige declaração Formulário A2 para remessas internacionais",
                    "alta",
                    "bloqueante",
                    [fema["codigo"]]
                )
        
        # Verificar restrições por país
        if dados_financeiros.get("tipo_transacao") == "remessa_internacional" and "pais_destino" in dados_financeiros:
            codigo_pais = dados_financeiros["pais_destino"].upper()
            
            if codigo_pais in fema["restricoes_por_pais"] and fema["restricoes_por_pais"][codigo_pais]:
                if not dados_financeiros.get("autorizacao_especial_fema", False):
                    self._registrar_violacao(
                        resultado,
                        "IN_FEMA_PAIS_RESTRITO",
                        f"FEMA restringe transações com {codigo_pais} sem autorização especial",
                        "alta",
                        "bloqueante",
                        [fema["codigo"]]
                    )
    
    def _validar_rbi(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações do Banco de Reserva da Índia.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        rbi = self.requisitos["rbi"]
        
        # Verificar limites de transação
        if "sistema_pagamento" in dados_financeiros and "valor_transacao" in dados_financeiros and dados_financeiros.get("moeda") == "INR":
            sistema = dados_financeiros["sistema_pagamento"].lower()
            try:
                valor = float(dados_financeiros["valor_transacao"])
                
                if sistema in rbi["limites_transacao"]:
                    # Para RTGS, verificamos valor mínimo (não máximo)
                    if sistema == "rtgs" and valor < rbi["limites_transacao"][sistema]:
                        self._registrar_violacao(
                            resultado,
                            "IN_RBI_VALOR_MINIMO_NAO_ATINGIDO",
                            f"Transações RTGS devem ter valor mínimo de {rbi['limites_transacao'][sistema]} INR",
                            "alta",
                            "bloqueante",
                            [rbi["codigo"]]
                        )
                    # Para outros sistemas, verificamos valor máximo
                    elif sistema != "rtgs" and valor > rbi["limites_transacao"][sistema]:
                        self._registrar_violacao(
                            resultado,
                            "IN_RBI_LIMITE_EXCEDIDO",
                            f"Valor excede o limite de {rbi['limites_transacao'][sistema]} INR para transações {sistema.upper()}",
                            "alta",
                            "bloqueante",
                            [rbi["codigo"]]
                        )
            except (ValueError, TypeError):
                pass
        
        # Verificar KYC obrigatório
        if rbi["kyc_obrigatorio"] and not contexto.get("kyc_completo_in", False):
            self._registrar_violacao(
                resultado,
                "IN_RBI_FALTA_KYC",
                "RBI exige KYC completo para transações financeiras",
                "alta",
                "bloqueante",
                [rbi["codigo"]]
            )
        
        # Verificar monitoramento de transação
        if rbi["monitoramento_transacao"] and contexto.get("transacao_suspeita", False):
            if not contexto.get("transacao_monitorada", False):
                self._registrar_violacao(
                    resultado,
                    "IN_RBI_FALTA_MONITORAMENTO",
                    "RBI exige monitoramento especial para transações suspeitas",
                    "alta",
                    "bloqueante",
                    [rbi["codigo"]]
                )
    
    def _validar_pmla(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei de Prevenção à Lavagem de Dinheiro.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        pmla = self.requisitos["pmla"]
        
        # Verificar limite reportável
        if "valor_transacao" in dados_financeiros and dados_financeiros.get("moeda") == "INR":
            try:
                valor = float(dados_financeiros["valor_transacao"])
                
                if valor >= pmla["limite_reportavel"] and not dados_financeiros.get("reportado_fiu", False):
                    self._registrar_violacao(
                        resultado,
                        "IN_PMLA_TRANSACAO_NAO_REPORTADA",
                        f"Transações acima de {pmla['limite_reportavel']} INR devem ser reportadas à FIU-IND",
                        "alta",
                        "bloqueante",
                        [pmla["codigo"]]
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar verificação de identidade
        if pmla["requer_verificacao_identidade"] and not contexto.get("verificacao_identidade_pmla", False):
            self._registrar_violacao(
                resultado,
                "IN_PMLA_FALTA_VERIFICACAO_IDENTIDADE",
                "PMLA exige verificação de identidade para transações financeiras",
                "alta",
                "bloqueante",
                [pmla["codigo"]]
            )
        
        # Verificar monitoramento PEP
        if pmla["monitoramento_pep"] and contexto.get("pessoa_pep", False):
            if not contexto.get("monitoramento_pep_in", False):
                self._registrar_violacao(
                    resultado,
                    "IN_PMLA_FALTA_MONITORAMENTO_PEP",
                    "PMLA exige monitoramento reforçado para Pessoas Politicamente Expostas",
                    "alta",
                    "bloqueante",
                    [pmla["codigo"]]
                )
        
        # Verificar manutenção de registros
        if pmla["manutencao_registros"] and not contexto.get("manutencao_registros_pmla", False):
            self._registrar_violacao(
                resultado,
                "IN_PMLA_FALTA_MANUTENCAO_REGISTROS",
                f"PMLA exige manutenção de registros de transações por {pmla['periodo_retencao']} anos",
                "media",
                "aviso",
                [pmla["codigo"]]
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
                f"Violação de conformidade Índia: {codigo} - {descricao} "
                f"(severidade: {severidade}, impacto: {impacto})"
            )
        
        # Métricas
        if self.metrics:
            self.metrics.incrementCounter(
                "trustguard.compliance.brics.india.violation", 
                {"code": codigo, "severity": severidade, "impact": impacto}
            )
