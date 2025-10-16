"""
Validador de Conformidade para China - BRICS

Este módulo implementa o validador de conformidade específico para a China
dentro do contexto BRICS, considerando PIPL, CSL, CBDC e regulamentações do PBOC.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, Any, List

class ChinaComplianceValidator:
    """
    Validador de conformidade para regulamentações específicas da China.
    """
    
    def __init__(self, logger=None, metrics=None):
        """
        Inicializa o validador de conformidade para a China.
        
        Args:
            logger: Logger para registro de eventos
            metrics: Serviço de métricas
        """
        self.logger = logger
        self.metrics = metrics
        self.requisitos = self._carregar_requisitos()
    
    def validar(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações específicas da China.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        dados_pessoais = contexto.get("dados_pessoais", {})
        dados_financeiros = contexto.get("dados_financeiros", {})
        
        if self.logger:
            self.logger.info("Iniciando validação de conformidade para China")
            
        if self.metrics:
            self.metrics.incrementCounter("trustguard.compliance.brics.china.validation")
        
        # Validar Lei de Proteção de Informações Pessoais
        self._validar_pipl(resultado, contexto, dados_pessoais)
        
        # Validar Lei de Segurança Cibernética
        self._validar_csl(resultado, contexto)
        
        # Validar regulamentações de moeda digital
        self._validar_cbdc(resultado, contexto, dados_financeiros)
        
        # Validar regulamentações do PBOC
        self._validar_pboc(resultado, contexto, dados_financeiros)
        
        return resultado
    
    def _carregar_requisitos(self) -> Dict[str, Any]:
        """
        Carrega os requisitos de conformidade específicos para a China.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "pipl": {
                "nome": "Lei de Proteção de Informações Pessoais",
                "codigo": "PIPL",
                "restricao_transferencia_internacional": True,
                "requer_consentimento": True,
                "restricao_dados_sensiveis": True,
                "dados_sensiveis": [
                    "biometria", "crença_religiosa", "identidade_específica", 
                    "saude", "financeiro", "localizacao", "menores"
                ]
            },
            "csl": {
                "nome": "Lei de Segurança Cibernética",
                "codigo": "CSL",
                "requer_localizacao_dados": True,
                "requer_avaliacao_seguranca": True,
                "infraestrutura_critica": [
                    "financas", "energia", "transporte", "agua", 
                    "saude", "educacao", "telecomunicacoes"
                ]
            },
            "cbdc": {
                "nome": "Moeda Digital do Banco Central",
                "codigo": "e-CNY",
                "suporte_obrigatorio": True,
                "restricoes": {
                    "limite_transacao_individual": 50000.0,  # CNY
                    "limite_diario": 100000.0  # CNY
                }
            },
            "pboc": {
                "nome": "Banco Popular da China",
                "codigo": "PBOC",
                "limite_transferencia_internacional": 50000.0,  # USD por ano
                "requer_autorizacao": True,
                "reportar_transacoes_grandes": 100000.0,  # CNY
                "restricoes_cripto": True
            }
        }
    
    def _validar_pipl(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_pessoais: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei de Proteção de Informações Pessoais.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_pessoais: Dados pessoais do contexto
        """
        if not dados_pessoais:
            return
            
        pipl = self.requisitos["pipl"]
        
        # Verificar transferência internacional de dados
        if pipl["restricao_transferencia_internacional"] and contexto.get("transferencia_internacional_dados", False):
            if not contexto.get("avaliacao_seguranca_dados", False):
                self._registrar_violacao(
                    resultado,
                    "CN_FALTA_AVALIACAO_SEGURANCA",
                    "PIPL requer avaliação de segurança para transferência internacional de dados",
                    "alta",
                    "bloqueante",
                    [pipl["codigo"]]
                )
            
            if not contexto.get("contrato_padronizado_transferencia", False):
                self._registrar_violacao(
                    resultado,
                    "CN_FALTA_CONTRATO_TRANSFERENCIA",
                    "PIPL requer contrato padronizado para transferência internacional de dados",
                    "alta",
                    "bloqueante",
                    [pipl["codigo"]]
                )
        
        # Verificar consentimento
        if pipl["requer_consentimento"] and not dados_pessoais.get("consentimento_cn", False):
            self._registrar_violacao(
                resultado,
                "CN_FALTA_CONSENTIMENTO",
                "PIPL requer consentimento para coleta e processamento de dados pessoais",
                "alta",
                "bloqueante",
                [pipl["codigo"]]
            )
        
        # Verificar dados sensíveis
        if pipl["restricao_dados_sensiveis"]:
            for tipo_dado in pipl["dados_sensiveis"]:
                if dados_pessoais.get(tipo_dado, False) and not dados_pessoais.get(f"consentimento_especifico_{tipo_dado}", False):
                    self._registrar_violacao(
                        resultado,
                        "CN_FALTA_CONSENTIMENTO_ESPECIFICO",
                        f"PIPL requer consentimento específico para dados sensíveis: {tipo_dado}",
                        "alta",
                        "bloqueante",
                        [pipl["codigo"]]
                    )
        
        # Verificar informações sobre processamento
        if not contexto.get("informacoes_processamento_cn", False):
            self._registrar_violacao(
                resultado,
                "CN_FALTA_INFORMACOES_PROCESSAMENTO",
                "PIPL requer informações claras sobre o processamento de dados pessoais",
                "media",
                "aviso",
                [pipl["codigo"]]
            )
    
    def _validar_csl(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> None:
        """
        Valida conformidade com a Lei de Segurança Cibernética.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
        """
        csl = self.requisitos["csl"]
        
        # Verificar localização de dados para operadores de infraestrutura crítica
        if csl["requer_localizacao_dados"]:
            setor = contexto.get("setor", "").lower()
            if setor in csl["infraestrutura_critica"] and not contexto.get("dados_armazenados_china", False):
                self._registrar_violacao(
                    resultado,
                    "CN_CSL_LOCALIZACAO_DADOS",
                    f"CSL requer que dados de infraestrutura crítica ({setor}) sejam armazenados na China",
                    "alta",
                    "bloqueante",
                    [csl["codigo"]]
                )
        
        # Verificar avaliação de segurança
        if csl["requer_avaliacao_seguranca"]:
            setor = contexto.get("setor", "").lower()
            if setor in csl["infraestrutura_critica"] and not contexto.get("avaliacao_seguranca_csl", False):
                self._registrar_violacao(
                    resultado,
                    "CN_CSL_FALTA_AVALIACAO_SEGURANCA",
                    f"CSL requer avaliação de segurança para operadores de infraestrutura crítica ({setor})",
                    "alta",
                    "bloqueante",
                    [csl["codigo"]]
                )
        
        # Verificar medidas de proteção de dados
        if not contexto.get("medidas_protecao_dados_cn", False):
            self._registrar_violacao(
                resultado,
                "CN_CSL_FALTA_MEDIDAS_PROTECAO",
                "CSL requer medidas adequadas de proteção de dados",
                "media",
                "aviso",
                [csl["codigo"]]
            )
    
    def _validar_cbdc(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações de moeda digital.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        cbdc = self.requisitos["cbdc"]
        
        # Verificar suporte a e-CNY
        if cbdc["suporte_obrigatorio"] and contexto.get("setor", "").lower() == "financas" and not contexto.get("suporte_ecny", False):
            self._registrar_violacao(
                resultado,
                "CN_FALTA_SUPORTE_ECNY",
                "É obrigatório suporte ao e-CNY (moeda digital do Banco Central da China) para instituições financeiras",
                "alta",
                "aviso",
                [cbdc["codigo"]]
            )
        
        # Verificar limites de transação se estiver usando e-CNY
        if dados_financeiros.get("moeda") == "e-CNY" or dados_financeiros.get("moeda") == "CNY" and dados_financeiros.get("tipo_moeda") == "digital":
            try:
                valor = float(dados_financeiros.get("valor_transacao", 0))
                
                # Verificar limite individual
                if valor > cbdc["restricoes"]["limite_transacao_individual"]:
                    self._registrar_violacao(
                        resultado,
                        "CN_ECNY_LIMITE_EXCEDIDO",
                        f"Valor excede o limite individual de {cbdc['restricoes']['limite_transacao_individual']} CNY para transações e-CNY",
                        "alta",
                        "bloqueante",
                        [cbdc["codigo"]]
                    )
                
                # Verificar limite diário
                if dados_financeiros.get("total_diario", 0) > cbdc["restricoes"]["limite_diario"]:
                    self._registrar_violacao(
                        resultado,
                        "CN_ECNY_LIMITE_DIARIO_EXCEDIDO",
                        f"Total diário excede o limite de {cbdc['restricoes']['limite_diario']} CNY para transações e-CNY",
                        "alta",
                        "bloqueante",
                        [cbdc["codigo"]]
                    )
            except (ValueError, TypeError):
                pass
    
    def _validar_pboc(self, resultado: Dict[str, Any], contexto: Dict[str, Any], dados_financeiros: Dict[str, Any]) -> None:
        """
        Valida conformidade com regulamentações do PBOC.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            dados_financeiros: Dados financeiros do contexto
        """
        if not dados_financeiros:
            return
            
        pboc = self.requisitos["pboc"]
        
        # Verificar limites de transferência internacional
        if dados_financeiros.get("tipo_transacao") == "internacional" and "valor_transacao" in dados_financeiros:
            try:
                valor = float(dados_financeiros["valor_transacao"])
                if dados_financeiros.get("moeda") == "USD" and valor > pboc["limite_transferencia_internacional"]:
                    if not dados_financeiros.get("autorizacao_especial_pboc", False):
                        self._registrar_violacao(
                            resultado,
                            "CN_LIMITE_TRANSFERENCIA_EXCEDIDO",
                            f"Valor excede o limite anual de {pboc['limite_transferencia_internacional']} USD para transferências internacionais",
                            "alta",
                            "bloqueante",
                            [pboc["codigo"]]
                        )
            except (ValueError, TypeError):
                pass
        
        # Verificar autorização para transferências internacionais
        if pboc["requer_autorizacao"] and dados_financeiros.get("tipo_transacao") == "internacional":
            if not dados_financeiros.get("autorizacao_pboc", False):
                self._registrar_violacao(
                    resultado,
                    "CN_FALTA_AUTORIZACAO_PBOC",
                    "PBOC requer autorização para transferências internacionais",
                    "alta",
                    "bloqueante",
                    [pboc["codigo"]]
                )
        
        # Verificar reporte de transações grandes
        if "valor_transacao" in dados_financeiros and dados_financeiros.get("moeda") == "CNY":
            try:
                valor = float(dados_financeiros["valor_transacao"])
                if valor >= pboc["reportar_transacoes_grandes"] and not dados_financeiros.get("reportado_pboc", False):
                    self._registrar_violacao(
                        resultado,
                        "CN_FALTA_REPORTE_TRANSACAO",
                        f"Transações acima de {pboc['reportar_transacoes_grandes']} CNY devem ser reportadas ao PBOC",
                        "media",
                        "aviso",
                        [pboc["codigo"]]
                    )
            except (ValueError, TypeError):
                pass
        
        # Verificar restrições de criptomoedas
        if pboc["restricoes_cripto"] and dados_financeiros.get("tipo_ativo") == "criptomoeda":
            self._registrar_violacao(
                resultado,
                "CN_RESTRICAO_CRIPTO",
                "PBOC proíbe transações com criptomoedas",
                "alta",
                "bloqueante",
                [pboc["codigo"]]
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
                f"Violação de conformidade China: {codigo} - {descricao} "
                f"(severidade: {severidade}, impacto: {impacto})"
            )
        
        # Métricas
        if self.metrics:
            self.metrics.incrementCounter(
                "trustguard.compliance.brics.china.violation", 
                {"code": codigo, "severity": severidade, "impact": impacto}
            )
