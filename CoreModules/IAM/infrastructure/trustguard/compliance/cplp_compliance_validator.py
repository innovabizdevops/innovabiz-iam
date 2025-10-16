"""
Validador de Conformidade para CPLP

Este módulo implementa o validador de conformidade para países da CPLP (Comunidade dos Países 
de Língua Portuguesa), considerando regulamentações comuns e específicas de cada país membro.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional
import re
import datetime

from .base_compliance_validator import BaseComplianceValidator
from ....observability.core.multi_layer_monitor import MultiLayerMonitor


class CPLPComplianceValidator(BaseComplianceValidator):
    """
    Validador de conformidade para países da CPLP.
    
    Implementa verificações de conformidade com regulamentações dos países da CPLP:
    - Angola, Brasil, Cabo Verde, Guiné-Bissau, Guiné Equatorial, Moçambique,
      Portugal, São Tomé e Príncipe e Timor-Leste
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o validador de conformidade para CPLP.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
        super().__init__(observability_monitor)
        
        # Países membros da CPLP
        self.paises_cplp = {
            "AO": "Angola",
            "BR": "Brasil",
            "CV": "Cabo Verde",
            "GW": "Guiné-Bissau",
            "GQ": "Guiné Equatorial",
            "MZ": "Moçambique",
            "PT": "Portugal",
            "ST": "São Tomé e Príncipe",
            "TL": "Timor-Leste"
        }
        
        # Mapeamento de país para formato de documento
        self.formatos_documentos = {
            "AO": {"bi": r"^\d{9}[A-Z]{2}\d{3}$", "nif": r"^\d{10}$"},
            "BR": {"cpf": r"^\d{3}\.\d{3}\.\d{3}-\d{2}$", "cnpj": r"^\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}$"},
            "PT": {"cc": r"^\d{8}-\d{1}$", "nif": r"^\d{9}$"},
            "MZ": {"bi": r"^\d{12}[A-Z]$", "nuit": r"^\d{9}$"},
            "CV": {"nif": r"^\d{9}$"},
            "GW": {"bi": r"^\d{10}$"},
            "ST": {"bi": r"^\d{8}$"},
            "TL": {"bi": r"^[A-Z]{2}\d{6}$"},
            "GQ": {"dni": r"^\d{8}[A-Z]$"}
        }
        
        # Requisitos de conformidade para CPLP
        self.requisitos = self._carregar_requisitos()
        
        self.logger.info("CPLPComplianceValidator inicializado com sucesso")
    
    def validar_conformidade(self, contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida a conformidade com as regulamentações dos países da CPLP.
        
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
            if not pais or pais not in self.paises_cplp:
                self.registrar_violacao(
                    resultado,
                    "PAIS_NAO_CPLP",
                    f"País não pertence à CPLP ou não foi especificado: {pais}",
                    "media",
                    "aviso"
                )
                # Continuar validação com regras gerais da CPLP
            
            # Registrar início da validação
            self.logger.info(f"Iniciando validação de conformidade CPLP para país: {pais or 'não especificado'}")
            self.metrics.incrementCounter("trustguard.compliance.cplp.validation", 
                                        {"country": pais or "unspecified"})
            
            # 1. Validar proteção de dados (GDPR para Portugal, LGPD para Brasil, etc)
            self._validar_protecao_dados(resultado, contexto, pais)
            
            # 2. Validar documentos de identificação específicos do país
            self._validar_documentos_identificacao(resultado, contexto, pais)
            
            # 3. Validar conformidade financeira
            if self._e_operacao_financeira(contexto):
                self._validar_conformidade_financeira(resultado, contexto, pais)
            
            # 4. Validar regulamentações específicas do país
            self._validar_regulamentacoes_especificas(resultado, contexto, pais)
            
            # Registrar conclusão da validação
            self.logger.info(
                f"Validação de conformidade CPLP concluída: "
                f"país={pais or 'não especificado'}, "
                f"válido={resultado['valido']}, "
                f"score={resultado['score_conformidade']}"
            )
            
            self.metrics.recordValue(
                "trustguard.compliance.cplp.score", 
                resultado["score_conformidade"],
                {"country": pais or "unspecified"}
            )
            
            return resultado
            
        except Exception as e:
            self.logger.error(f"Erro durante validação de conformidade CPLP: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.compliance.cplp.error", 
                {"error_type": type(e).__name__}
            )
            
            # Retornar resultado de erro
            return {
                "valido": False,
                "violacoes": [
                    {
                        "codigo": "ERRO_VALIDACAO_CPLP",
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
        Obtém os requisitos de conformidade específicos para CPLP.
        
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
        Carrega os requisitos de conformidade específicos para CPLP.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "protecao_dados": {
                "PT": {
                    "nome": "Regulamento Geral de Proteção de Dados",
                    "codigo": "RGPD/GDPR",
                    "idade_consentimento": 16,
                    "campos_sensiveis": ["raca", "religiao", "orientacao_politica", "saude", "orientacao_sexual", "dados_biometricos"]
                },
                "BR": {
                    "nome": "Lei Geral de Proteção de Dados",
                    "codigo": "LGPD (Lei nº 13.709/2018)",
                    "idade_consentimento": 12,
                    "campos_sensiveis": ["raca", "religiao", "opiniao_politica", "saude", "vida_sexual", "dados_biometricos", "dados_geneticos"]
                },
                "AO": {
                    "nome": "Lei de Proteção de Dados",
                    "codigo": "Lei n.º 22/11",
                    "idade_consentimento": 18,
                    "campos_sensiveis": ["raca", "religiao", "orientacao_politica", "saude"]
                },
                "padrao": {
                    "nome": "Requisitos gerais de proteção de dados CPLP",
                    "codigo": "CPLP-PDP",
                    "idade_consentimento": 18,
                    "campos_sensiveis": ["raca", "religiao", "opiniao_politica", "saude", "dados_biometricos"]
                }
            },
            "financeiro": {
                "PT": {
                    "moeda": "EUR",
                    "limite_transacao": 10000.0,
                    "banco_central": "Banco de Portugal",
                    "requer_identificacao_forte": True
                },
                "BR": {
                    "moeda": "BRL",
                    "limite_transacao": 10000.0,
                    "banco_central": "Banco Central do Brasil",
                    "requer_identificacao_forte": True
                },
                "AO": {
                    "moeda": "AOA",
                    "limite_transacao": 500000.0,
                    "banco_central": "Banco Nacional de Angola",
                    "requer_identificacao_forte": True
                },
                "MZ": {
                    "moeda": "MZN",
                    "limite_transacao": 100000.0,
                    "banco_central": "Banco de Moçambique",
                    "requer_identificacao_forte": True
                },
                "CV": {
                    "moeda": "CVE",
                    "limite_transacao": 100000.0,
                    "banco_central": "Banco de Cabo Verde",
                    "requer_identificacao_forte": True
                },
                "padrao": {
                    "requer_identificacao_forte": True,
                    "limite_transacao": 5000.0,  # Em USD como referência
                    "requer_origem_fundos": True
                }
            },
            "documentos": {
                "tipos_por_pais": {
                    "AO": ["bi", "nif"],
                    "BR": ["cpf", "cnpj", "rg"],
                    "PT": ["cc", "nif"],
                    "MZ": ["bi", "nuit"],
                    "CV": ["bi", "nif"],
                    "GW": ["bi"],
                    "ST": ["bi"],
                    "TL": ["bi"],
                    "GQ": ["dni"]
                }
            },
            "especificos": {
                "BR": {
                    "marco_civil": {
                        "nome": "Marco Civil da Internet",
                        "codigo": "Lei nº 12.965/2014",
                        "retencao_logs": 6  # meses
                    }
                },
                "PT": {
                    "servicos_sociedade_informacao": {
                        "nome": "Lei dos Serviços da Sociedade de Informação",
                        "codigo": "Decreto-Lei n.º 7/2004"
                    }
                },
                "AO": {
                    "comunicacoes_eletronicas": {
                        "nome": "Lei das Comunicações Eletrônicas",
                        "codigo": "Lei n.º 23/11"
                    }
                }
            }
        }
    
    def _validar_protecao_dados(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida conformidade com leis de proteção de dados.
        
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
        
        # Obter requisitos específicos do país ou usar padrão
        req_protecao = self.requisitos["protecao_dados"]
        req_pais = req_protecao.get(pais, req_protecao["padrao"]) if pais else req_protecao["padrao"]
        
        # Validar idade para consentimento
        if "idade" in dados_pessoais:
            try:
                idade = int(dados_pessoais["idade"])
                if idade < req_pais["idade_consentimento"]:
                    self.registrar_violacao(
                        resultado,
                        "IDADE_INSUFICIENTE",
                        f"Idade mínima para consentimento em {self.paises_cplp.get(pais, 'CPLP')} é {req_pais['idade_consentimento']} anos",
                        "alta",
                        "bloqueante",
                        [req_pais["codigo"]]
                    )
            except (ValueError, TypeError):
                self.registrar_violacao(
                    resultado,
                    "FORMATO_IDADE_INVALIDO",
                    "Formato de idade inválido",
                    "media",
                    "aviso"
                )
        
        # Validar dados sensíveis
        for campo_sensivel in req_pais["campos_sensiveis"]:
            if campo_sensivel in dados_pessoais and dados_pessoais.get(campo_sensivel):
                # Verificar se há consentimento explícito para dados sensíveis
                if not dados_pessoais.get("consentimento_explicito", False):
                    self.registrar_violacao(
                        resultado,
                        f"FALTA_CONSENTIMENTO_DADOS_SENSIVEIS",
                        f"Dados sensíveis requerem consentimento explícito segundo a lei de {self.paises_cplp.get(pais, 'CPLP')}",
                        "alta",
                        "bloqueante",
                        [req_pais["codigo"]]
                    )
                    break  # Uma violação é suficiente
        
        # Verificar finalidade
        if not contexto.get("finalidade_uso_dados"):
            self.registrar_violacao(
                resultado,
                "FALTA_FINALIDADE",
                "A finalidade do uso dos dados pessoais deve ser especificada",
                "alta",
                "bloqueante",
                [req_pais["codigo"]]
            )
        
        return resultado
    
    def _validar_documentos_identificacao(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida documentos de identificação específicos do país.
        
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
        
        # Se o país não for especificado ou não for da CPLP, não validar documentos
        if not pais or pais not in self.paises_cplp:
            return resultado
        
        # Verificar se pelo menos um documento válido para o país foi fornecido
        documentos_pais = self.requisitos["documentos"]["tipos_por_pais"].get(pais, [])
        tem_documento_valido = False
        
        for tipo_doc in documentos_pais:
            if tipo_doc in dados_pessoais and dados_pessoais[tipo_doc]:
                tem_documento_valido = True
                
                # Validar formato do documento, se houver regex definido
                if pais in self.formatos_documentos and tipo_doc in self.formatos_documentos[pais]:
                    padrao = self.formatos_documentos[pais][tipo_doc]
                    if not re.match(padrao, dados_pessoais[tipo_doc]):
                        self.registrar_violacao(
                            resultado,
                            f"FORMATO_INVALIDO_{tipo_doc.upper()}",
                            f"Formato de {tipo_doc.upper()} inválido para {self.paises_cplp[pais]}",
                            "media",
                            "aviso"
                        )
        
        # Se nenhum documento válido foi fornecido, registrar violação
        if not tem_documento_valido and documentos_pais:
            self.registrar_violacao(
                resultado,
                "FALTA_DOCUMENTO_IDENTIFICACAO",
                f"É necessário fornecer pelo menos um documento de identificação válido para {self.paises_cplp[pais]}: {', '.join(documentos_pais)}",
                "alta",
                "bloqueante"
            )
        
        return resultado
    
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
    
    def _validar_conformidade_financeira(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações financeiras.
        
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
        
        # Obter requisitos financeiros específicos do país ou usar padrão
        req_financeiro = self.requisitos["financeiro"]
        req_pais = req_financeiro.get(pais, req_financeiro["padrao"]) if pais else req_financeiro["padrao"]
        
        # Validar valor da transação
        if "valor_transacao" in dados_financeiros and "moeda" in dados_financeiros:
            try:
                valor = float(dados_financeiros["valor_transacao"])
                moeda = dados_financeiros["moeda"]
                
                # Se a moeda do país for diferente da transação, considerar como internacional
                moeda_pais = req_pais.get("moeda")
                if moeda_pais and moeda != moeda_pais:
                    if not dados_financeiros.get("declaracao_cambio", False):
                        self.registrar_violacao(
                            resultado,
                            "FALTA_DECLARACAO_CAMBIO",
                            f"Transação em moeda estrangeira ({moeda}) requer declaração cambial",
                            "media",
                            "aviso"
                        )
                
                # Verificar se excede limite
                if "limite_transacao" in req_pais:
                    # Se for a moeda do país, comparar diretamente
                    if moeda == req_pais.get("moeda"):
                        if valor > req_pais["limite_transacao"] and not dados_financeiros.get("verificacao_adicional", False):
                            self.registrar_violacao(
                                resultado,
                                "TRANSACAO_ACIMA_LIMITE",
                                f"Transação acima do limite permitido sem verificação adicional ({req_pais['limite_transacao']} {moeda})",
                                "alta",
                                "bloqueante"
                            )
            except (ValueError, TypeError):
                self.registrar_violacao(
                    resultado,
                    "FORMATO_VALOR_INVALIDO",
                    "Formato de valor de transação inválido",
                    "media",
                    "aviso"
                )
        
        # Verificar identificação forte para transações
        if req_pais.get("requer_identificacao_forte", False):
            if not dados_financeiros.get("autenticacao_forte", False):
                self.registrar_violacao(
                    resultado,
                    "FALTA_AUTENTICACAO_FORTE",
                    f"Transações financeiras em {self.paises_cplp.get(pais, 'CPLP')} requerem autenticação forte (2FA/MFA)",
                    "alta",
                    "bloqueante"
                )
        
        # Verificar origem de fundos para valores altos
        if req_pais.get("requer_origem_fundos", False):
            try:
                if "valor_transacao" in dados_financeiros:
                    valor = float(dados_financeiros["valor_transacao"])
                    limite = req_pais.get("limite_transacao", 5000.0)
                    
                    if valor > limite and not dados_financeiros.get("origem_fundos_declarada", False):
                        self.registrar_violacao(
                            resultado,
                            "FALTA_ORIGEM_FUNDOS",
                            f"Transações acima de {limite} requerem declaração de origem de fundos",
                            "alta",
                            "bloqueante"
                        )
            except (ValueError, TypeError):
                pass
        
        return resultado
    
    def _validar_regulamentacoes_especificas(self, resultado: Dict[str, Any], contexto: Dict[str, Any], pais: Optional[str]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações específicas do país.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            pais: Código do país
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        # Se o país não for especificado ou não tiver regulamentações específicas, retornar
        if not pais or pais not in self.requisitos["especificos"]:
            return resultado
        
        req_especificos = self.requisitos["especificos"][pais]
        
        # Verificações específicas por país
        if pais == "BR":
            # Marco Civil da Internet (Brasil)
            if contexto.get("tipo_operacao") == "servico_online" and not contexto.get("retencao_logs_configurada"):
                self.registrar_violacao(
                    resultado,
                    "FALTA_RETENCAO_LOGS",
                    f"Marco Civil da Internet requer retenção de logs por {req_especificos['marco_civil']['retencao_logs']} meses",
                    "media",
                    "aviso",
                    [req_especificos['marco_civil']['codigo']]
                )
        
        elif pais == "PT":
            # Lei dos Serviços da Sociedade de Informação (Portugal)
            if contexto.get("tipo_operacao") == "comercio_eletronico" and not contexto.get("termos_servico_apresentados"):
                self.registrar_violacao(
                    resultado,
                    "FALTA_TERMOS_SERVICO",
                    "Lei dos Serviços da Sociedade de Informação requer apresentação clara dos termos de serviço",
                    "media",
                    "aviso",
                    [req_especificos['servicos_sociedade_informacao']['codigo']]
                )
        
        elif pais == "AO":
            # Lei das Comunicações Eletrônicas (Angola)
            if contexto.get("tipo_operacao") == "telecomunicacao" and not contexto.get("registro_regulador"):
                self.registrar_violacao(
                    resultado,
                    "FALTA_REGISTRO_REGULADOR",
                    "Lei das Comunicações Eletrônicas requer registro junto ao regulador INACOM",
                    "media",
                    "aviso",
                    [req_especificos['comunicacoes_eletronicas']['codigo']]
                )
        
        return resultado
