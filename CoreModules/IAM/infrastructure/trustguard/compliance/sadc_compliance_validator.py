"""
Validador de Conformidade para SADC

Este módulo implementa o validador de conformidade para países da SADC 
(Comunidade de Desenvolvimento da África Austral), considerando regulamentações
comuns e específicas dos países membros.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional
import re
import datetime

from .base_compliance_validator import BaseComplianceValidator
from ....observability.core.multi_layer_monitor import MultiLayerMonitor


class SADCComplianceValidator(BaseComplianceValidator):
    """
    Validador de conformidade para países da SADC.
    
    Implementa verificações de conformidade com regulamentações dos países da SADC:
    - Angola, Botswana, Comores, RD Congo, Eswatini, Lesoto, Madagascar,
      Malawi, Maurícia, Moçambique, Namíbia, Seychelles, África do Sul,
      Tanzânia, Zâmbia e Zimbabwe
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o validador de conformidade para SADC.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
        super().__init__(observability_monitor)
        
        # Países membros da SADC
        self.paises_sadc = {
            "AO": "Angola",
            "BW": "Botswana",
            "KM": "Comores",
            "CD": "RD Congo",
            "SZ": "Eswatini",
            "LS": "Lesoto",
            "MG": "Madagascar",
            "MW": "Malawi",
            "MU": "Maurícia",
            "MZ": "Moçambique",
            "NA": "Namíbia",
            "SC": "Seychelles",
            "ZA": "África do Sul",
            "TZ": "Tanzânia",
            "ZM": "Zâmbia",
            "ZW": "Zimbabwe"
        }
        
        # Moedas por país
        self.moedas_por_pais = {
            "AO": "AOA",
            "BW": "BWP",
            "KM": "KMF",
            "CD": "CDF",
            "SZ": "SZL",
            "LS": "LSL",
            "MG": "MGA",
            "MW": "MWK",
            "MU": "MUR",
            "MZ": "MZN",
            "NA": "NAD",
            "SC": "SCR",
            "ZA": "ZAR",
            "TZ": "TZS",
            "ZM": "ZMW",
            "ZW": "USD"  # Zimbabwe usa várias moedas, mas USD é comum
        }
        
        # Requisitos de conformidade para SADC
        self.requisitos = self._carregar_requisitos()
        
        self.logger.info("SADCComplianceValidator inicializado com sucesso")
    
    def validar_conformidade(self, contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida a conformidade com as regulamentações dos países da SADC.
        
        Args:
            contexto: Contexto para validação de conformidade
            
        Returns:
            Dict[str, Any]: Resultado da validação de conformidade
        """
        try:
            # Iniciar com resultado padrão
            resultado = self._resultado_padrao.copy()
            
            # Obter país de contexto
            pais = self._obter_pais_contexto(contexto)
            if not pais or pais not in self.paises_sadc:
                self.registrar_violacao(
                    resultado,
                    "PAIS_NAO_SADC",
                    f"País não pertence à SADC ou não foi especificado: {pais}",
                    "media",
                    "aviso"
                )
                # Continuar validação com regras gerais da SADC
            
            # Registrar início da validação
            self.logger.info(f"Iniciando validação de conformidade SADC para país: {pais or 'não especificado'}")
            self.metrics.incrementCounter("trustguard.compliance.sadc.validation", 
                                        {"country": pais or "unspecified"})
            
            # 1. Validar conformidade com protocolos da SADC
            self._validar_protocolos_sadc(resultado, contexto, pais)
            
            # 2. Validar operações financeiras e cambiais
            if self._e_operacao_financeira(contexto):
                self._validar_operacoes_financeiras(resultado, contexto, pais)
            
            # 3. Validar documentos de identificação
            self._validar_documentos_identificacao(resultado, contexto, pais)
            
            # 4. Validar regras específicas do país
            self._validar_regras_especificas_pais(resultado, contexto, pais)
            
            # Registrar conclusão da validação
            self.logger.info(
                f"Validação de conformidade SADC concluída: "
                f"país={pais or 'não especificado'}, "
                f"válido={resultado['valido']}, "
                f"score={resultado['score_conformidade']}"
            )
            
            self.metrics.recordValue(
                "trustguard.compliance.sadc.score", 
                resultado["score_conformidade"],
                {"country": pais or "unspecified"}
            )
            
            return resultado
            
        except Exception as e:
            self.logger.error(f"Erro durante validação de conformidade SADC: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.compliance.sadc.error", 
                {"error_type": type(e).__name__}
            )
            
            # Retornar resultado de erro
            return {
                "valido": False,
                "violacoes": [
                    {
                        "codigo": "ERRO_VALIDACAO_SADC",
                        "descricao": f"Erro durante validação de conformidade: {str(e)}",
                        "severidade": "alta",
                        "impacto": "bloqueante"
                    }
                ],
                "avisos": [],
                "score_conformidade": 0,
                "erro": True
            }
    
    def obter_requisitos_conformidade(self) -> Dict[str, Any]:
        """
        Obtém os requisitos de conformidade específicos para SADC.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return self.requisitos
    
    def _obter_pais_contexto(self, contexto: Dict[str, Any]) -> Optional[str]:
        """
        Obtém o código do país a partir do contexto.
        
        Args:
            contexto: Contexto da operação
            
        Returns:
            Optional[str]: Código do país (ISO 3166-1 alpha-2) ou None
        """
        # Tentar obter de diferentes locais do contexto
        pais = None
        
        # Da localização
        if "localizacao" in contexto and "pais" in contexto["localizacao"]:
            pais = contexto["localizacao"]["pais"]
        
        # Dos dados pessoais
        elif "dados_pessoais" in contexto and "pais" in contexto["dados_pessoais"]:
            pais = contexto["dados_pessoais"]["pais"]
        
        # Dos dados financeiros
        elif "dados_financeiros" in contexto and "pais" in contexto["dados_financeiros"]:
            pais = contexto["dados_financeiros"]["pais"]
        
        # Normalizar para maiúsculas
        if pais:
            return pais.upper()
        
        return None
    
    def _carregar_requisitos(self) -> Dict[str, Any]:
        """
        Carrega os requisitos de conformidade específicos para SADC.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "protocolo_financas": {
                "nome": "Protocolo sobre Finanças e Investimentos da SADC",
                "codigo": "SADC-FIP",
                "controle_cambial": {
                    "limites_transferencia": {
                        "ZA": 1000000.0,  # ZAR
                        "AO": 10000000.0,  # AOA
                        "MZ": 500000.0,   # MZN
                        "padrao": 10000.0  # USD
                    },
                    "reportar_acima_limite": True,
                    "restricoes_moeda_estrangeira": ["ZW", "AO"]
                },
                "anti_corrupcao": {
                    "verificar_pep": True,
                    "verificar_origem_fundos": True,
                    "paises_alto_risco": ["CD", "ZW"]
                }
            },
            "protocolo_comercio": {
                "nome": "Protocolo sobre Comércio da SADC",
                "codigo": "SADC-TP",
                "requer_documentacao_origem": ["mercadoria", "servico_transfronteirico"],
                "paises_implementacao_completa": ["ZA", "BW", "MU", "NA", "SC", "SZ", "LS"]
            },
            "documentos_identificacao": {
                "tipos_por_pais": {
                    "AO": ["bi", "passaporte"],
                    "ZA": ["id", "passport"],
                    "MZ": ["bi", "nuit", "passaporte"],
                    "ZW": ["national_id", "passport"],
                    "NA": ["id", "passport"],
                    "BW": ["omang", "passport"],
                    "padrao": ["id_nacional", "passaporte"]
                }
            },
            "aml_cft": {
                "nome": "Combate à Lavagem de Dinheiro e Financiamento do Terrorismo",
                "requer_due_diligence": True,
                "verificacao_reforçada": {
                    "paises": ["CD", "ZW", "MG"],
                    "valor_limite": 5000.0  # USD
                }
            },
            "protecao_dados": {
                "paises_com_legislacao": ["AO", "ZA", "MU", "SC", "BW"],
                "transferencia_transfronteiriça": {
                    "requer_consentimento": True,
                    "paises_adequados": ["ZA", "MU", "SC"]
                }
            },
            "regras_especificas": {
                "ZA": {
                    "popi": {
                        "nome": "Lei de Proteção de Informações Pessoais",
                        "codigo": "POPI Act",
                        "requer_responsavel_protecao_dados": True,
                        "notificacao_violacao": True
                    },
                    "fica": {
                        "nome": "Lei de Inteligência Financeira",
                        "codigo": "FICA",
                        "limite_transacao_reportavel": 25000.0  # ZAR
                    }
                },
                "AO": {
                    "lei_protecao_dados": {
                        "nome": "Lei de Proteção de Dados",
                        "codigo": "Lei n.º 22/11"
                    },
                    "cambio": {
                        "nome": "Regulamentação cambial",
                        "codigo": "Lei de Cambios",
                        "limite_transferencia": 10000.0  # USD
                    }
                }
            }
        }
    
    def _e_operacao_financeira(self, contexto: Dict[str, Any]) -> bool:
        """
        Verifica se o contexto representa uma operação financeira.
        
        Args:
            contexto: Contexto da operação
            
        Returns:
            bool: True se for uma operação financeira
        """
        # Verificar tipo de operação
        if "tipo_operacao" in contexto and any(tipo in contexto["tipo_operacao"].lower() 
                                             for tipo in ["transacao", "pagamento", "financeira", "transferencia"]):
            return True
        
        # Verificar presença de dados financeiros
        return "dados_financeiros" in contexto and bool(contexto["dados_financeiros"])
    
    def _validar_protocolos_sadc(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida conformidade com os protocolos da SADC.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            pais: Código do país
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        tipo_operacao = contexto.get("tipo_operacao", "").lower()
        
        # Validar Protocolo de Comércio para operações comerciais
        if "comercio" in tipo_operacao or "mercadoria" in tipo_operacao:
            protocolo_comercio = self.requisitos["protocolo_comercio"]
            
            # Verificar se é uma operação que requer documentação de origem
            for tipo_requer_doc in protocolo_comercio["requer_documentacao_origem"]:
                if tipo_requer_doc in tipo_operacao and not contexto.get("documentacao_origem", False):
                    self.registrar_violacao(
                        resultado,
                        "FALTA_DOCUMENTACAO_ORIGEM",
                        f"O Protocolo de Comércio da SADC requer documentação de origem para operações de {tipo_requer_doc}",
                        "alta",
                        "bloqueante",
                        [protocolo_comercio["codigo"]]
                    )
        
        # Validar transferências de dados entre países da SADC
        if "dados_pessoais" in contexto and contexto.get("transferencia_transfronteirica", False):
            protecao_dados = self.requisitos["protecao_dados"]
            
            # Verificar consentimento para transferência transfronteiriça
            if protecao_dados["transferencia_transfronteiriça"]["requer_consentimento"]:
                if not contexto.get("consentimento_transferencia", False):
                    self.registrar_violacao(
                        resultado,
                        "FALTA_CONSENTIMENTO_TRANSFERENCIA",
                        "Transferências transfronteiriças de dados pessoais na região SADC requerem consentimento explícito",
                        "alta",
                        "bloqueante"
                    )
            
            # Verificar se o país destino tem proteção adequada
            pais_destino = contexto.get("pais_destino")
            if pais_destino and pais_destino not in protecao_dados["transferencia_transfronteiriça"]["paises_adequados"]:
                self.registrar_violacao(
                    resultado,
                    "PAIS_SEM_PROTECAO_ADEQUADA",
                    f"Transferência de dados pessoais para {pais_destino} requer salvaguardas adicionais",
                    "media",
                    "aviso"
                )
        
        return resultado
    
    def _validar_operacoes_financeiras(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida operações financeiras conforme protocolo financeiro da SADC.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            pais: Código do país
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        dados_financeiros = contexto.get("dados_financeiros", {})
        if not dados_financeiros:
            return resultado
        
        protocolo_financas = self.requisitos["protocolo_financas"]
        controle_cambial = protocolo_financas["controle_cambial"]
        
        # Verificar limites de transferência
        if "valor_transacao" in dados_financeiros and "moeda" in dados_financeiros:
            try:
                valor = float(dados_financeiros["valor_transacao"])
                moeda = dados_financeiros["moeda"]
                
                # Determinar limite aplicável
                limite = None
                
                if pais and pais in controle_cambial["limites_transferencia"]:
                    # Verificar se a moeda da transação corresponde à moeda do país
                    moeda_pais = self.moedas_por_pais.get(pais)
                    
                    if moeda == moeda_pais:
                        limite = controle_cambial["limites_transferencia"][pais]
                else:
                    # Usar limite padrão (em USD)
                    limite = controle_cambial["limites_transferencia"]["padrao"]
                
                if limite and valor > limite:
                    # Verificar se há aprovação especial
                    if not dados_financeiros.get("aprovacao_transferencia_alta", False):
                        self.registrar_violacao(
                            resultado,
                            "LIMITE_TRANSFERENCIA_EXCEDIDO",
                            f"Valor da transação excede o limite permitido ({limite} {moeda}) sem aprovação especial",
                            "alta",
                            "bloqueante",
                            [protocolo_financas["codigo"]]
                        )
                    
                    # Verificar se foi reportada conforme necessário
                    if controle_cambial["reportar_acima_limite"] and not dados_financeiros.get("reportada_autoridade", False):
                        self.registrar_violacao(
                            resultado,
                            "TRANSFERENCIA_NAO_REPORTADA",
                            "Transferências acima do limite devem ser reportadas às autoridades financeiras",
                            "media",
                            "aviso",
                            [protocolo_financas["codigo"]]
                        )
            except (ValueError, TypeError):
                self.registrar_violacao(
                    resultado,
                    "FORMATO_VALOR_INVALIDO",
                    "Formato de valor de transação inválido",
                    "media",
                    "aviso"
                )
        
        # Verificar restrições cambiais específicas do país
        if pais in controle_cambial["restricoes_moeda_estrangeira"]:
            if dados_financeiros.get("tipo_transacao") == "internacional" and not dados_financeiros.get("autorizacao_cambial", False):
                self.registrar_violacao(
                    resultado,
                    "FALTA_AUTORIZACAO_CAMBIAL",
                    f"Operações internacionais em {self.paises_sadc[pais]} requerem autorização cambial específica",
                    "alta",
                    "bloqueante"
                )
        
        # Verificar regras anti-corrupção
        if dados_financeiros.get("valor_transacao", 0) > 5000:  # Valor significativo
            anti_corrupcao = protocolo_financas["anti_corrupcao"]
            
            # Verificar se é PEP (Pessoa Politicamente Exposta)
            if anti_corrupcao["verificar_pep"] and contexto.get("pessoa_pep", False):
                if not dados_financeiros.get("verificacao_reforçada_pep", False):
                    self.registrar_violacao(
                        resultado,
                        "FALTA_VERIFICACAO_PEP",
                        "Transações envolvendo Pessoas Politicamente Expostas requerem verificação reforçada",
                        "alta",
                        "bloqueante"
                    )
            
            # Verificar origem dos fundos para valores altos
            if anti_corrupcao["verificar_origem_fundos"] and not dados_financeiros.get("origem_fundos_verificada", False):
                self.registrar_violacao(
                    resultado,
                    "FALTA_VERIFICACAO_ORIGEM_FUNDOS",
                    "Transações de valor significativo requerem verificação da origem dos fundos",
                    "alta",
                    "bloqueante"
                )
            
            # Verificação adicional para países de alto risco
            if pais in anti_corrupcao["paises_alto_risco"] and not dados_financeiros.get("due_diligence_reforçada", False):
                self.registrar_violacao(
                    resultado,
                    "FALTA_DUE_DILIGENCE_REFORÇADA",
                    f"Transações em {self.paises_sadc[pais]} requerem due diligence reforçada",
                    "alta",
                    "bloqueante"
                )
        
        return resultado
    
    def _validar_documentos_identificacao(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida documentos de identificação conforme exigido pelos países da SADC.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            pais: Código do país
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        dados_pessoais = contexto.get("dados_pessoais", {})
        if not dados_pessoais:
            return resultado
        
        # Determinar quais documentos são necessários
        documentos_requeridos = self.requisitos["documentos_identificacao"]["tipos_por_pais"].get(
            pais, 
            self.requisitos["documentos_identificacao"]["tipos_por_pais"]["padrao"]
        ) if pais else self.requisitos["documentos_identificacao"]["tipos_por_pais"]["padrao"]
        
        # Verificar se pelo menos um documento necessário foi fornecido
        tem_documento_valido = False
        for tipo_doc in documentos_requeridos:
            if tipo_doc in dados_pessoais and dados_pessoais[tipo_doc]:
                tem_documento_valido = True
                break
        
        if not tem_documento_valido:
            self.registrar_violacao(
                resultado,
                "FALTA_DOCUMENTO_IDENTIFICACAO",
                f"É necessário fornecer pelo menos um documento de identificação válido: {', '.join(documentos_requeridos)}",
                "alta",
                "bloqueante"
            )
        
        return resultado
    
    def _validar_regras_especificas_pais(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida regras específicas do país dentro da SADC.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            pais: Código do país
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        if not pais or pais not in self.requisitos["regras_especificas"]:
            return resultado
        
        regras_pais = self.requisitos["regras_especificas"][pais]
        
        # Regras específicas para África do Sul
        if pais == "ZA":
            # POPI Act (Proteção de Dados)
            if "dados_pessoais" in contexto:
                popi = regras_pais["popi"]
                
                # Verificar se houve violação de dados
                if contexto.get("violacao_dados", False) and not contexto.get("notificacao_violacao", False):
                    self.registrar_violacao(
                        resultado,
                        "FALTA_NOTIFICACAO_VIOLACAO",
                        "POPI Act requer notificação de violações de dados pessoais",
                        "alta",
                        "bloqueante",
                        [popi["codigo"]]
                    )
            
            # FICA (Inteligência Financeira)
            if "dados_financeiros" in contexto:
                fica = regras_pais["fica"]
                dados_financeiros = contexto["dados_financeiros"]
                
                # Verificar limite de transação reportável
                if dados_financeiros.get("moeda") == "ZAR" and "valor_transacao" in dados_financeiros:
                    try:
                        valor = float(dados_financeiros["valor_transacao"])
                        if valor >= fica["limite_transacao_reportavel"] and not dados_financeiros.get("fica_reportado", False):
                            self.registrar_violacao(
                                resultado,
                                "TRANSACAO_NAO_REPORTADA_FICA",
                                f"Transações acima de {fica['limite_transacao_reportavel']} ZAR devem ser reportadas conforme FICA",
                                "alta",
                                "bloqueante",
                                [fica["codigo"]]
                            )
                    except (ValueError, TypeError):
                        pass
        
        # Regras específicas para Angola
        elif pais == "AO":
            # Regulamentação cambial
            if "dados_financeiros" in contexto:
                cambio = regras_pais["cambio"]
                dados_financeiros = contexto["dados_financeiros"]
                
                # Verificar se é transação internacional
                if dados_financeiros.get("tipo_transacao") == "internacional":
                    # Verificar se tem autorização do BNA
                    if not dados_financeiros.get("autorizacao_bna", False):
                        self.registrar_violacao(
                            resultado,
                            "FALTA_AUTORIZACAO_BNA",
                            "Transferências internacionais em Angola requerem autorização do BNA",
                            "alta",
                            "bloqueante",
                            [cambio["codigo"]]
                        )
                    
                    # Verificar limite de transferência
                    if "valor_transacao" in dados_financeiros and dados_financeiros.get("moeda") == "USD":
                        try:
                            valor = float(dados_financeiros["valor_transacao"])
                            if valor > cambio["limite_transferencia"] and not dados_financeiros.get("autorizacao_especial", False):
                                self.registrar_violacao(
                                    resultado,
                                    "LIMITE_TRANSFERENCIA_EXCEDIDO_AO",
                                    f"Transferências acima de {cambio['limite_transferencia']} USD requerem autorização especial em Angola",
                                    "alta",
                                    "bloqueante",
                                    [cambio["codigo"]]
                                )
                        except (ValueError, TypeError):
                            pass
        
        return resultado
