"""
Validador de Conformidade Unificado

Este módulo implementa o integrador unificado para todos os validadores de conformidade
regionais, permitindo validação contextual baseada na localização ou contexto da operação.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional, Type
import logging
import importlib

from .base_compliance_validator import BaseComplianceValidator
from ...observability.core.multi_layer_monitor import MultiLayerMonitor

# Importar validadores regionais específicos
from .brics_compliance_validator import BRICSComplianceValidator
# Outros validadores regionais serão importados conforme necessário

class UnifiedComplianceValidator(BaseComplianceValidator):
    """
    Validador de Conformidade Unificado que integra todos os validadores regionais.
    
    Este validador serve como ponto de entrada para validação de conformidade,
    encaminhando requisições para os validadores específicos de cada região
    com base no contexto da operação.
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o validador de conformidade unificado.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
        super().__init__(observability_monitor)
        
        # Mapeamento de regiões/blocos econômicos para seus validadores
        self.validadores_por_regiao = {
            "BRICS": BRICSComplianceValidator(observability_monitor),
            # Outros validadores serão adicionados aqui conforme implementados
            # "SADC": SADCComplianceValidator(observability_monitor),
            # "EU": EUComplianceValidator(observability_monitor),
            # "US": USComplianceValidator(observability_monitor),
        }
        
        # Mapeamento de países para regiões/blocos econômicos
        self.regioes_por_pais = self._carregar_mapeamento_regioes_paises()
        
        # Carregar configurações de prioridade de validação
        self.prioridades_validacao = self._carregar_prioridades_validacao()
        
        self.logger.info("UnifiedComplianceValidator inicializado com sucesso")
    
    def validar_conformidade(self, contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida a conformidade com base no contexto da operação, encaminhando
        para os validadores regionais adequados.
        
        Args:
            contexto: Contexto para validação de conformidade
            
        Returns:
            Dict[str, Any]: Resultado da validação de conformidade
        """
        try:
            # Iniciar com resultado padrão
            resultado = self._resultado_padrao.copy()
            
            # Identificar país e regiões aplicáveis
            pais = self._obter_pais_contexto(contexto)
            regioes_aplicaveis = self._obter_regioes_aplicaveis(pais, contexto)
            
            # Registrar início da validação
            self.logger.info(
                f"Iniciando validação de conformidade unificada para "
                f"país: {pais or 'não especificado'}, "
                f"regiões: {', '.join(regioes_aplicaveis) if regioes_aplicaveis else 'não especificadas'}"
            )
            
            self.metrics.incrementCounter(
                "trustguard.compliance.unified.validation",
                {"country": pais or "unspecified"}
            )
            
            # Aplicar validadores específicos para cada região na ordem de prioridade
            if regioes_aplicaveis:
                for regiao in sorted(regioes_aplicaveis, 
                                     key=lambda r: self.prioridades_validacao.get(r, 999)):
                    if regiao in self.validadores_por_regiao:
                        self.logger.info(f"Aplicando validador de conformidade para região: {regiao}")
                        
                        validador_regional = self.validadores_por_regiao[regiao]
                        resultado_regional = validador_regional.validar_conformidade(contexto)
                        
                        # Mesclar resultados
                        self._mesclar_resultados(resultado, resultado_regional)
                        
                        self.metrics.incrementCounter(
                            "trustguard.compliance.unified.regional_validation",
                            {"region": regiao}
                        )
            else:
                # Sem regiões identificadas, aplicar validação genérica
                self.logger.warning(
                    "Nenhuma região específica identificada para validação de conformidade. "
                    "Aplicando validação genérica."
                )
                
                self._aplicar_validacao_generica(resultado, contexto)
            
            # Verificar conformidade cross-regional
            self._validar_cross_regional(resultado, contexto, pais, regioes_aplicaveis)
            
            # Registrar conclusão da validação
            self.logger.info(
                f"Validação de conformidade unificada concluída: "
                f"país={pais or 'não especificado'}, "
                f"válido={resultado['valido']}, "
                f"score={resultado['score_conformidade']}"
            )
            
            self.metrics.recordValue(
                "trustguard.compliance.unified.score", 
                resultado["score_conformidade"],
                {"country": pais or "unspecified"}
            )
            
            return resultado
            
        except Exception as e:
            self.logger.error(f"Erro durante validação de conformidade unificada: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.compliance.unified.error", 
                {"error_type": type(e).__name__}
            )
            
            # Retornar resultado de erro
            return {
                "valido": False,
                "violacoes": [
                    {
                        "codigo": "ERRO_VALIDACAO_UNIFICADA",
                        "descricao": f"Erro durante validação de conformidade unificada: {str(e)}",
                        "severidade": "alta",
                        "impacto": "bloqueante"
                    }
                ],
                "avisos": [],
                "score_conformidade": 0,
                "erro": True
            }
    
    def _mesclar_resultados(self, resultado_principal: Dict[str, Any], 
                           resultado_regional: Dict[str, Any]) -> None:
        """
        Mescla resultados de validação regional no resultado principal.
        
        Args:
            resultado_principal: Resultado principal a ser atualizado
            resultado_regional: Resultado regional a ser mesclado
        """
        # Atualizar status de validade (se qualquer um for inválido, o resultado é inválido)
        if not resultado_regional["valido"]:
            resultado_principal["valido"] = False
        
        # Mesclar violações
        for violacao in resultado_regional.get("violacoes", []):
            resultado_principal["violacoes"].append(violacao)
        
        # Mesclar avisos
        for aviso in resultado_regional.get("avisos", []):
            resultado_principal["avisos"].append(aviso)
        
        # Ajustar score de conformidade (média ponderada)
        score_atual = resultado_principal["score_conformidade"]
        score_regional = resultado_regional.get("score_conformidade", 1.0)
        
        # Usar o menor score como base (mais restritivo)
        resultado_principal["score_conformidade"] = min(score_atual, score_regional)
    
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
    
    def _obter_regioes_aplicaveis(self, pais: Optional[str], 
                                 contexto: Dict[str, Any]) -> List[str]:
        """
        Determina as regiões aplicáveis com base no país e no contexto.
        
        Args:
            pais: Código do país (ISO 3166-1 alpha-2) ou None
            contexto: Contexto da operação
            
        Returns:
            List[str]: Lista de regiões aplicáveis
        """
        regioes = set()
        
        # Adicionar regiões baseadas no país
        if pais and pais in self.regioes_por_pais:
            regioes.update(self.regioes_por_pais[pais])
        
        # Adicionar regiões explicitamente especificadas no contexto
        if "regioes_aplicaveis" in contexto:
            regioes_ctx = contexto["regioes_aplicaveis"]
            if isinstance(regioes_ctx, list):
                regioes.update([r.upper() for r in regioes_ctx])
            elif isinstance(regioes_ctx, str):
                regioes.add(regioes_ctx.upper())
        
        # Adicionar regiões com base no tipo de operação
        if "tipo_operacao" in contexto:
            tipo_op = contexto["tipo_operacao"]
            
            # Exemplo: Se é uma operação internacional, considerar regiões de origem e destino
            if tipo_op == "transferencia_internacional" and "pais_destino" in contexto:
                pais_destino = contexto["pais_destino"].upper()
                if pais_destino in self.regioes_por_pais:
                    regioes.update(self.regioes_por_pais[pais_destino])
        
        return list(regioes)
    
    def _aplicar_validacao_generica(self, resultado: Dict[str, Any], contexto: Dict[str, Any]) -> None:
        """
        Aplica validação genérica quando nenhuma região específica é identificada.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
        """
        # Aplicar verificações básicas de conformidade genéricas
        
        # Verificar se há consentimento para processamento de dados
        if "dados_pessoais" in contexto and not contexto.get("consentimento_dados_pessoais", False):
            self.registrar_violacao(
                resultado,
                "FALTA_CONSENTIMENTO_DADOS",
                "Falta consentimento para processamento de dados pessoais",
                "alta",
                "bloqueante"
            )
        
        # Verificar KYC básico
        if "tipo_operacao" in contexto and contexto["tipo_operacao"] in ["financeira", "transferencia", "pagamento"]:
            if not contexto.get("kyc_completo", False):
                self.registrar_violacao(
                    resultado,
                    "KYC_INCOMPLETO",
                    "KYC incompleto para operação financeira",
                    "alta",
                    "bloqueante"
                )
    
    def _validar_cross_regional(self, resultado: Dict[str, Any], contexto: Dict[str, Any],
                              pais: Optional[str], regioes: List[str]) -> None:
        """
        Valida conformidade para regras que cruzam várias regiões.
        
        Args:
            resultado: Resultado da validação
            contexto: Contexto da operação
            pais: Código do país
            regioes: Lista de regiões aplicáveis
        """
        # Verificar operações internacionais entre regiões diferentes
        if "tipo_operacao" in contexto and contexto["tipo_operacao"] == "transferencia_internacional":
            if "pais_origem" in contexto and "pais_destino" in contexto:
                pais_origem = contexto["pais_origem"].upper()
                pais_destino = contexto["pais_destino"].upper()
                
                # Verificar se os países estão em regiões diferentes
                regioes_origem = set(self.regioes_por_pais.get(pais_origem, []))
                regioes_destino = set(self.regioes_por_pais.get(pais_destino, []))
                
                regioes_diferentes = regioes_origem.symmetric_difference(regioes_destino)
                if regioes_diferentes:
                    # Verificar se há autorização para transferência entre regiões diferentes
                    if not contexto.get("autorizacao_transferencia_cross_regional", False):
                        self.registrar_violacao(
                            resultado,
                            "TRANSFERENCIA_CROSS_REGIONAL_SEM_AUTORIZACAO",
                            f"Transferência entre regiões diferentes ({', '.join(regioes_diferentes)}) "
                            f"requer autorização específica",
                            "alta",
                            "bloqueante"
                        )
    
    def _carregar_mapeamento_regioes_paises(self) -> Dict[str, List[str]]:
        """
        Carrega o mapeamento de países para regiões econômicas.
        
        Returns:
            Dict[str, List[str]]: Mapeamento de códigos de países para listas de regiões
        """
        return {
            # BRICS
            "BR": ["BRICS", "MERCOSUL", "LATAM"],
            "RU": ["BRICS", "EAEU"],
            "IN": ["BRICS", "SAARC"],
            "CN": ["BRICS", "RCEP"],
            "ZA": ["BRICS", "SADC", "AU"],
            "EG": ["BRICS", "AU", "COMESA"],
            "ET": ["BRICS", "AU", "COMESA"],
            "IR": ["BRICS", "ECO"],
            "AE": ["BRICS", "GCC"],
            "SA": ["BRICS", "GCC"],
            
            # União Europeia
            "AT": ["EU", "SEPA", "GDPR"],
            "BE": ["EU", "SEPA", "GDPR"],
            "BG": ["EU", "SEPA", "GDPR"],
            "HR": ["EU", "SEPA", "GDPR"],
            "CY": ["EU", "SEPA", "GDPR"],
            "CZ": ["EU", "SEPA", "GDPR"],
            "DK": ["EU", "SEPA", "GDPR"],
            "EE": ["EU", "SEPA", "GDPR"],
            "FI": ["EU", "SEPA", "GDPR"],
            "FR": ["EU", "SEPA", "GDPR"],
            "DE": ["EU", "SEPA", "GDPR"],
            "GR": ["EU", "SEPA", "GDPR"],
            "HU": ["EU", "SEPA", "GDPR"],
            "IE": ["EU", "SEPA", "GDPR"],
            "IT": ["EU", "SEPA", "GDPR"],
            "LV": ["EU", "SEPA", "GDPR"],
            "LT": ["EU", "SEPA", "GDPR"],
            "LU": ["EU", "SEPA", "GDPR"],
            "MT": ["EU", "SEPA", "GDPR"],
            "NL": ["EU", "SEPA", "GDPR"],
            "PL": ["EU", "SEPA", "GDPR"],
            "PT": ["EU", "SEPA", "GDPR", "CPLP"],
            "RO": ["EU", "SEPA", "GDPR"],
            "SK": ["EU", "SEPA", "GDPR"],
            "SI": ["EU", "SEPA", "GDPR"],
            "ES": ["EU", "SEPA", "GDPR"],
            "SE": ["EU", "SEPA", "GDPR"],
            
            # Estados Unidos e Américas
            "US": ["US", "NAFTA"],
            "CA": ["NAFTA"],
            "MX": ["NAFTA", "LATAM"],
            "AR": ["MERCOSUL", "LATAM"],
            "PY": ["MERCOSUL", "LATAM"],
            "UY": ["MERCOSUL", "LATAM"],
            
            # África
            "AO": ["SADC", "AU", "PALOP", "CPLP"],
            "MZ": ["SADC", "AU", "PALOP", "CPLP"],
            "NA": ["SADC", "AU"],
            "BW": ["SADC", "AU"],
            "ZW": ["SADC", "AU"],
            "MU": ["SADC", "AU"],
            "MG": ["SADC", "AU"],
            
            # CPLP (Comunidade dos Países de Língua Portuguesa)
            "CV": ["CPLP", "PALOP", "AU"],
            "GW": ["CPLP", "PALOP", "AU", "ECOWAS"],
            "ST": ["CPLP", "PALOP", "AU"],
            "TL": ["CPLP", "ASEAN"],
            
            # Ásia-Pacífico
            "JP": ["RCEP", "APAC"],
            "KR": ["RCEP", "APAC"],
            "AU": ["RCEP", "APAC"],
            "NZ": ["RCEP", "APAC"],
            "SG": ["RCEP", "APAC", "ASEAN"],
            "MY": ["RCEP", "APAC", "ASEAN"],
            "ID": ["RCEP", "APAC", "ASEAN"],
            "TH": ["RCEP", "APAC", "ASEAN"],
            "PH": ["RCEP", "APAC", "ASEAN"],
            "VN": ["RCEP", "APAC", "ASEAN"]
        }
    
    def _carregar_prioridades_validacao(self) -> Dict[str, int]:
        """
        Carrega as prioridades de validação para cada região.
        Quanto menor o número, maior a prioridade.
        
        Returns:
            Dict[str, int]: Mapeamento de regiões para prioridades
        """
        return {
            # Prioridades primárias (mais restritivas)
            "US": 10,      # Regulações dos EUA têm alta prioridade devido ao alcance extraterritorial
            "EU": 20,      # GDPR e outras regulações europeias têm alta prioridade
            "GDPR": 30,    # Específico para proteção de dados
            
            # Prioridades secundárias (blocos econômicos regionais)
            "BRICS": 100,
            "SADC": 110,
            "MERCOSUL": 120,
            "NAFTA": 130,
            "ASEAN": 140,
            "GCC": 150,
            "EAEU": 160,
            
            # Prioridades terciárias
            "CPLP": 200,
            "PALOP": 210,
            "LATAM": 220,
            "APAC": 230,
            "AU": 240,
            
            # Prioridades específicas funcionais
            "SEPA": 300,   # Pagamentos europeus
            "PCI": 310     # Segurança de cartões de pagamento
        }
