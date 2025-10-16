"""
Validador de Conformidade para Angola

Este módulo implementa o validador de conformidade específico para o mercado angolano,
considerando regulamentações locais como a Lei de Proteção de Dados (Lei n.º 22/11),
regulamentações bancárias do BNA e normas de telecomunicações da INACOM.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional, Set, Tuple
import datetime
import re
import json

from .base_compliance_validator import BaseComplianceValidator
from ....observability.core.multi_layer_monitor import MultiLayerMonitor


class AngolaComplianceValidator(BaseComplianceValidator):
    """
    Validador de conformidade específico para o mercado angolano.
    
    Implementa verificações de conformidade com regulamentações angolanas relevantes:
    - Lei de Proteção de Dados (Lei n.º 22/11)
    - Regulamentações bancárias do BNA (Banco Nacional de Angola)
    - Regulamentações de telecomunicações da INACOM
    - Normas anti-lavagem de dinheiro e combate ao financiamento do terrorismo
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o validador de conformidade para Angola.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
        super().__init__(observability_monitor)
        
        # Requisitos de conformidade específicos para Angola
        self.requisitos = self._carregar_requisitos()
        
        # Expressões regulares para validação de documentos angolanos
        self.validadores = {
            "bi": re.compile(r"^\d{9}[A-Z]{2}\d{3}$"),  # Formato simplificado BI Angola
            "nif": re.compile(r"^\d{10}$"),             # NIF angolano (10 dígitos)
            "celular": re.compile(r"^(9[1-5]|9[9])\d{7}$")  # Números de celular Angola
        }
        
        self.logger.info("AngolaComplianceValidator inicializado com sucesso")

    def validar_conformidade(self, contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida a conformidade com as regulamentações angolanas.
        
        Args:
            contexto: Contexto para validação de conformidade
            
        Returns:
            Dict[str, Any]: Resultado da validação de conformidade
        """
        try:
            # Iniciar com resultado padrão
            resultado = self._resultado_padrao.copy()
            
            # Registrar início da validação
            self.logger.info(f"Iniciando validação de conformidade angolana para operação: {contexto.get('tipo_operacao', 'desconhecida')}")
            self.metrics.incrementCounter("trustguard.compliance.angola.validation", 
                                        {"operation_type": contexto.get('tipo_operacao', 'unknown')})
            
            # Obter detalhes relevantes do contexto
            tipo_operacao = contexto.get('tipo_operacao', '').lower()
            dados_pessoais = contexto.get('dados_pessoais', {})
            dados_financeiros = contexto.get('dados_financeiros', {})
            
            # 1. Validações de Proteção de Dados (Lei n.º 22/11)
            if dados_pessoais:
                self._validar_protecao_dados(resultado, dados_pessoais)
            
            # 2. Validações bancárias (se aplicável)
            if 'transacao' in tipo_operacao or dados_financeiros:
                self._validar_regulamentacao_bancaria(resultado, contexto)
            
            # 3. Validações de telecomunicações (se aplicável)
            if 'telecomunicacao' in tipo_operacao or 'celular' in dados_pessoais:
                self._validar_regulamentacao_telecom(resultado, contexto)
            
            # 4. Validações AML/CFT (para todas as operações financeiras)
            if 'transacao' in tipo_operacao or dados_financeiros:
                self._validar_aml_cft(resultado, contexto)
            
            # Registrar conclusão da validação
            self.logger.info(
                f"Validação de conformidade angolana concluída: "
                f"válido={resultado['valido']}, "
                f"score={resultado['score_conformidade']}, "
                f"violações={len(resultado.get('violacoes', []))}, "
                f"avisos={len(resultado.get('avisos', []))}"
            )
            
            self.metrics.recordValue(
                "trustguard.compliance.angola.score", 
                resultado["score_conformidade"]
            )
            
            return resultado
            
        except Exception as e:
            self.logger.error(f"Erro durante validação de conformidade angolana: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.compliance.angola.error", 
                {"error_type": type(e).__name__}
            )
            
            # Retornar resultado de erro
            return {
                "valido": False,
                "violacoes": [
                    {
                        "codigo": "ERRO_VALIDACAO_ANGOLA",
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
        Obtém os requisitos de conformidade específicos para Angola.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return self.requisitos
    
    def _carregar_requisitos(self) -> Dict[str, Any]:
        """
        Carrega os requisitos de conformidade específicos para Angola.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        return {
            "protecao_dados": {
                "nome": "Lei de Proteção de Dados",
                "codigo": "Lei n.º 22/11",
                "descricao": "Lei angolana que regula a proteção de dados pessoais",
                "campos_obrigatorios": ["nome", "bi_ou_nif"],
                "campos_sensiveis": ["raca", "religiao", "orientacao_politica", "saude"],
                "idade_minima_consentimento": 18,
                "periodo_retencao_padrao": 5  # anos
            },
            "regulamentacao_bancaria": {
                "nome": "Regulamentações Bancárias do BNA",
                "codigo": "Instrutivo n.º 07/2018 e Avisos do BNA",
                "descricao": "Normas que regulam operações bancárias em Angola",
                "limite_transacao_sem_verificacao": 100000,  # AOA
                "limite_transacao_diaria": 500000,  # AOA
                "verificacao_identidade_obrigatoria": True,
                "validacao_origem_fundos": ["transacao_internacional", "valor_elevado"]
            },
            "telecomunicacoes": {
                "nome": "Regulamentações de Telecomunicações da INACOM",
                "codigo": "Lei n.º 23/11",
                "descricao": "Normas que regulam serviços de telecomunicações em Angola",
                "registro_obrigatorio": True,
                "formatos_celular": ["9XXXXXXXX"],
                "operadoras": ["91", "92", "93", "94", "95", "99"]
            },
            "aml_cft": {
                "nome": "Combate à Lavagem de Dinheiro e Financiamento do Terrorismo",
                "codigo": "Lei n.º 34/11",
                "descricao": "Lei que estabelece medidas de prevenção e combate à lavagem de dinheiro e financiamento ao terrorismo",
                "operacoes_reportaveis": ["transacao_internacional", "valor_elevado"],
                "limite_reportavel": 1000000,  # AOA
                "paises_alto_risco": ["IR", "KP", "MM", "SY"],
                "documentos_obrigatorios": ["comprovante_identidade", "comprovante_residencia"]
            }
        }
    
    def _validar_protecao_dados(self, resultado: Dict[str, Any], dados_pessoais: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com a Lei de Proteção de Dados angolana.
        
        Args:
            resultado: Resultado da validação
            dados_pessoais: Dados pessoais a serem validados
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        requisitos = self.requisitos["protecao_dados"]
        
        # Verificar campos obrigatórios
        for campo in requisitos["campos_obrigatorios"]:
            if campo == "bi_ou_nif":
                # Verificar se pelo menos um documento está presente
                if "bi" not in dados_pessoais and "nif" not in dados_pessoais:
                    self.registrar_violacao(
                        resultado,
                        "FALTA_DOCUMENTO_IDENTIFICACAO",
                        "É necessário fornecer pelo menos um documento de identificação (BI ou NIF)",
                        "alta",
                        "bloqueante",
                        ["Lei n.º 22/11", "Artigo 14"]
                    )
            else:
                # Verificar outros campos obrigatórios
                resultado = self.validar_campo_requerido(
                    resultado, 
                    dados_pessoais, 
                    campo, 
                    f"Campo '{campo}' é obrigatório pela Lei de Proteção de Dados angolana"
                )
        
        # Verificar formato de documentos, se fornecidos
        if "bi" in dados_pessoais:
            resultado = self.validar_formato_campo(
                resultado,
                dados_pessoais,
                "bi",
                lambda x: bool(self.validadores["bi"].match(x)),
                "Formato de BI angolano inválido. Deve seguir o padrão: 9 dígitos + 2 letras + 3 dígitos",
                "alta",
                "bloqueante"
            )
        
        if "nif" in dados_pessoais:
            resultado = self.validar_formato_campo(
                resultado,
                dados_pessoais,
                "nif",
                lambda x: bool(self.validadores["nif"].match(x)),
                "Formato de NIF angolano inválido. Deve conter exatamente 10 dígitos numéricos",
                "alta",
                "bloqueante"
            )
        
        # Verificar dados sensíveis
        for campo_sensivel in requisitos["campos_sensiveis"]:
            if campo_sensivel in dados_pessoais and dados_pessoais.get(campo_sensivel):
                # Verificar se há consentimento explícito para dados sensíveis
                if not dados_pessoais.get("consentimento_explicito_" + campo_sensivel, False):
                    self.registrar_violacao(
                        resultado,
                        f"FALTA_CONSENTIMENTO_{campo_sensivel.upper()}",
                        f"Dados sensíveis de {campo_sensivel} requerem consentimento explícito segundo a Lei angolana",
                        "alta",
                        "bloqueante",
                        ["Lei n.º 22/11", "Artigo 6"]
                    )
        
        # Verificar idade para consentimento
        if "idade" in dados_pessoais:
            try:
                idade = int(dados_pessoais["idade"])
                if idade < requisitos["idade_minima_consentimento"]:
                    self.registrar_violacao(
                        resultado,
                        "IDADE_INSUFICIENTE",
                        f"Idade mínima para consentimento em Angola é {requisitos['idade_minima_consentimento']} anos",
                        "alta",
                        "bloqueante",
                        ["Lei n.º 22/11", "Artigo 5"]
                    )
            except (ValueError, TypeError):
                self.registrar_violacao(
                    resultado,
                    "FORMATO_IDADE_INVALIDO",
                    "Formato de idade inválido",
                    "media",
                    "aviso"
                )
        
        return resultado
    
    def _validar_regulamentacao_bancaria(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações bancárias do BNA.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto completo da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        requisitos = self.requisitos["regulamentacao_bancaria"]
        dados_financeiros = contexto.get("dados_financeiros", {})
        
        # Verificar limites de transação
        if "valor_transacao" in dados_financeiros:
            try:
                valor = float(dados_financeiros["valor_transacao"])
                
                # Verificar se excede limite sem verificação adicional
                if valor > requisitos["limite_transacao_sem_verificacao"]:
                    # Verificar se há verificação adicional
                    if not dados_financeiros.get("verificacao_adicional", False):
                        self.registrar_violacao(
                            resultado,
                            "FALTA_VERIFICACAO_VALOR_ALTO",
                            f"Transações acima de {requisitos['limite_transacao_sem_verificacao']} AOA requerem verificação adicional",
                            "alta",
                            "bloqueante",
                            ["Instrutivo n.º 07/2018 do BNA"]
                        )
                
                # Verificar limite diário
                limite_diario = requisitos["limite_transacao_diaria"]
                total_diario = dados_financeiros.get("total_transacoes_diarias", 0) + valor
                
                if total_diario > limite_diario and not dados_financeiros.get("aprovacao_limite_excedido", False):
                    self.registrar_violacao(
                        resultado,
                        "LIMITE_DIARIO_EXCEDIDO",
                        f"Limite diário de transações excedido ({limite_diario} AOA)",
                        "alta",
                        "bloqueante",
                        ["Avisos do BNA sobre limites operacionais"]
                    )
            
            except (ValueError, TypeError):
                self.registrar_violacao(
                    resultado,
                    "FORMATO_VALOR_INVALIDO",
                    "Formato de valor de transação inválido",
                    "media",
                    "aviso"
                )
        
        # Verificar se é transação internacional
        if dados_financeiros.get("tipo_transacao") == "internacional":
            # Verificar se origem de fundos foi validada
            if not dados_financeiros.get("origem_fundos_validada", False):
                self.registrar_violacao(
                    resultado,
                    "FALTA_VALIDACAO_ORIGEM_FUNDOS",
                    "Transações internacionais requerem validação da origem dos fundos",
                    "alta",
                    "bloqueante",
                    ["Lei n.º 34/11", "Artigo 12"]
                )
            
            # Verificar país de destino (sanções)
            pais_destino = dados_financeiros.get("pais_destino", "")
            if pais_destino in self.requisitos["aml_cft"]["paises_alto_risco"]:
                self.registrar_violacao(
                    resultado,
                    "TRANSACAO_PAIS_ALTO_RISCO",
                    f"Transação para país de alto risco: {pais_destino}",
                    "alta",
                    "bloqueante",
                    ["Circular do BNA sobre Países de Alto Risco"]
                )
        
        return resultado
    
    def _validar_regulamentacao_telecom(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com regulamentações de telecomunicações da INACOM.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto completo da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        requisitos = self.requisitos["telecomunicacoes"]
        dados_pessoais = contexto.get("dados_pessoais", {})
        
        # Validar formato de número de celular angolano
        if "celular" in dados_pessoais:
            resultado = self.validar_formato_campo(
                resultado,
                dados_pessoais,
                "celular",
                lambda x: bool(self.validadores["celular"].match(x)),
                "Formato de número de celular angolano inválido. Deve começar com 91-95 ou 99 seguido de 7 dígitos",
                "media",
                "aviso"
            )
            
            # Verificar operadora válida
            if dados_pessoais.get("celular", ""):
                celular = dados_pessoais["celular"]
                if len(celular) >= 2:
                    operadora = celular[:2]
                    if operadora not in requisitos["operadoras"]:
                        self.registrar_violacao(
                            resultado,
                            "OPERADORA_INVALIDA",
                            f"Operadora de celular inválida. Operadoras válidas em Angola: {', '.join(requisitos['operadoras'])}",
                            "baixa",
                            "informativo"
                        )
        
        return resultado
    
    def _validar_aml_cft(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida conformidade com normas anti-lavagem de dinheiro e combate ao financiamento do terrorismo.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto completo da operação
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        requisitos = self.requisitos["aml_cft"]
        dados_financeiros = contexto.get("dados_financeiros", {})
        
        # Verificar se a transação é reportável
        e_reportavel = False
        
        # Verificar tipo de operação
        tipo_transacao = dados_financeiros.get("tipo_transacao", "")
        if tipo_transacao in requisitos["operacoes_reportaveis"]:
            e_reportavel = True
        
        # Verificar valor da transação
        if "valor_transacao" in dados_financeiros:
            try:
                valor = float(dados_financeiros["valor_transacao"])
                if valor >= requisitos["limite_reportavel"]:
                    e_reportavel = True
            except (ValueError, TypeError):
                pass
        
        # Se for reportável, verificar documentação obrigatória
        if e_reportavel:
            documentos = dados_financeiros.get("documentos_fornecidos", [])
            for doc_obrigatorio in requisitos["documentos_obrigatorios"]:
                if doc_obrigatorio not in documentos:
                    self.registrar_violacao(
                        resultado,
                        f"FALTA_DOCUMENTO_{doc_obrigatorio.upper()}",
                        f"Documento obrigatório não fornecido: {doc_obrigatorio}",
                        "alta",
                        "bloqueante",
                        ["Lei n.º 34/11", "Artigo 5"]
                    )
            
            # Verificar se a operação foi marcada para reportar
            if not dados_financeiros.get("marcada_para_reporte", False):
                self.registrar_violacao(
                    resultado,
                    "FALTA_MARCACAO_REPORTE",
                    "Operação deve ser marcada para reporte às autoridades competentes",
                    "media",
                    "aviso",
                    ["Lei n.º 34/11", "Artigo 13"]
                )
        
        return resultado
