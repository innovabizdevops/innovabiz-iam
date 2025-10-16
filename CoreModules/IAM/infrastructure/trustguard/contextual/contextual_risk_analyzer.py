"""
Analisador de Risco Contextual para o TrustGuard

Este módulo implementa o analisador de risco contextual que avalia o nível
de risco de uma operação com base em múltiplos fatores contextuais, incluindo
localização, comportamento do usuário, histórico de transações e configurações
específicas por tenant e mercado.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional, Tuple
import datetime
import json
from enum import Enum

from ...observability.core.multi_layer_monitor import MultiLayerMonitor
from .multi_tenant_context_provider import MultiTenantContextProvider, ContextoAnalise
from ..behaviors.behavior_analyzer import UserBehaviorAnalyzer
from ..rules.rule_engine import RuleEngine
from ...constants.constants import NIVEIS_RISCO


class ResultadoAvaliacao(Enum):
    """Resultado da avaliação de risco."""
    PERMITIDO = "permitido"
    VERIFICACAO_ADICIONAL = "verificacao_adicional"
    BLOQUEADO = "bloqueado"


class AvaliacaoRisco:
    """Resultado da avaliação de risco contextual."""
    
    def __init__(
        self,
        nivel_risco: str,
        pontuacao: float,
        resultado: ResultadoAvaliacao,
        fatores_contribuintes: Dict[str, float],
        acoes_recomendadas: List[str],
        detalhes: Dict[str, Any]
    ):
        self.nivel_risco = nivel_risco
        self.pontuacao = pontuacao
        self.resultado = resultado
        self.fatores_contribuintes = fatores_contribuintes
        self.acoes_recomendadas = acoes_recomendadas
        self.detalhes = detalhes
        self.timestamp = datetime.datetime.now()


class ContextualRiskAnalyzer:
    """
    Analisador de risco contextual.
    
    Combina informações de contexto, comportamento do usuário e regras específicas
    do tenant para determinar o nível de risco de uma operação.
    """
    
    def __init__(
        self, 
        context_provider: MultiTenantContextProvider,
        behavior_analyzer: UserBehaviorAnalyzer,
        rule_engine: RuleEngine,
        observability_monitor: Optional[MultiLayerMonitor] = None
    ):
        """
        Inicializa o analisador de risco contextual.
        
        Args:
            context_provider: Provedor de contexto multi-tenant
            behavior_analyzer: Analisador de comportamento do usuário
            rule_engine: Motor de regras para avaliação
            observability_monitor: Monitor de observabilidade
        """
        self.context_provider = context_provider
        self.behavior_analyzer = behavior_analyzer
        self.rule_engine = rule_engine
        self.observability = observability_monitor
        
        # Configurar logger e métricas
        if observability_monitor:
            self.logger = observability_monitor.get_logger()
            self.metrics = observability_monitor.get_metrics_collector()
        else:
            from ...observability.logging.hook_logger import Logger
            from ...observability.metrics.hook_metrics import MetricsCollector
            
            self.logger = Logger()
            self.metrics = MetricsCollector()
        
        self.logger.info("ContextualRiskAnalyzer inicializado")
    
    def avaliar_risco(self, contexto: ContextoAnalise) -> AvaliacaoRisco:
        """
        Avalia o risco de uma operação com base no contexto fornecido.
        
        Args:
            contexto: Contexto da operação a ser avaliada
            
        Returns:
            AvaliacaoRisco: Resultado da avaliação de risco
        """
        inicio = datetime.datetime.now()
        
        # Registrar início da avaliação
        self.logger.info(f"Iniciando avaliação de risco para usuário {contexto.usuario_id} do tenant {contexto.tenant_id}")
        self.metrics.incrementCounter(
            "trustguard.risk_analyzer.evaluation", 
            {"tenant_id": contexto.tenant_id}
        )
        
        try:
            # Enriquecer contexto com informações específicas do tenant
            contexto_enriquecido = self.context_provider.enriquecer_contexto(contexto)
            
            # Obter configuração do tenant
            tenant_config = self.context_provider.obter_tenant(contexto.tenant_id)
            if not tenant_config:
                self.logger.error(f"Tenant {contexto.tenant_id} não encontrado. Usando configuração padrão de alto risco.")
                return self._criar_avaliacao_padrao_alto_risco(contexto)
            
            # Obter análise comportamental do usuário
            analise_comportamental = self.behavior_analyzer.analyze_user_behavior(
                contexto.usuario_id,
                contexto.tenant_id,
                contexto_enriquecido
            )
            
            # Obter regras específicas do tenant para o tipo de operação
            tipo_regra = "transacao" if contexto.tipo_transacao else "autenticacao"
            regras = self.context_provider.obter_regras_avaliacao(contexto.tenant_id, tipo_regra)
            
            # Executar motor de regras
            resultado_regras = self.rule_engine.executar_regras(
                regras,
                {
                    "contexto": contexto_enriquecido,
                    "comportamento": analise_comportamental,
                    "tenant_config": tenant_config
                }
            )
            
            # Calcular pontuação final de risco
            pontuacao_risco = self._calcular_pontuacao_risco(
                resultado_regras,
                analise_comportamental,
                contexto_enriquecido
            )
            
            # Determinar nível de risco com base na pontuação
            nivel_risco = self._determinar_nivel_risco(pontuacao_risco)
            
            # Determinar resultado da avaliação
            resultado = self._determinar_resultado_avaliacao(
                nivel_risco, 
                tenant_config.nivel_seguranca_padrao,
                contexto_enriquecido
            )
            
            # Gerar lista de ações recomendadas
            acoes_recomendadas = self._gerar_acoes_recomendadas(
                nivel_risco,
                resultado,
                contexto_enriquecido,
                tenant_config
            )
            
            # Registrar fatores que contribuíram para o risco
            fatores_contribuintes = self._identificar_fatores_contribuintes(
                resultado_regras,
                analise_comportamental
            )
            
            # Criar resultado da avaliação
            avaliacao = AvaliacaoRisco(
                nivel_risco=nivel_risco,
                pontuacao=pontuacao_risco,
                resultado=resultado,
                fatores_contribuintes=fatores_contribuintes,
                acoes_recomendadas=acoes_recomendadas,
                detalhes={
                    "regras_acionadas": resultado_regras.get("regras_acionadas", []),
                    "anomalias_comportamentais": analise_comportamental.get("anomalias", []),
                    "contexto_hash": self.context_provider.calcular_hash_contexto(contexto),
                }
            )
            
            # Registrar métricas de resultado
            self._registrar_metricas_resultado(avaliacao, contexto.tenant_id, inicio)
            
            return avaliacao
            
        except Exception as e:
            self.logger.error(f"Erro durante avaliação de risco: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.risk_analyzer.error", 
                {"tenant_id": contexto.tenant_id, "error_type": type(e).__name__}
            )
            return self._criar_avaliacao_padrao_alto_risco(contexto)
    
    def _calcular_pontuacao_risco(
        self,
        resultado_regras: Dict[str, Any],
        analise_comportamental: Dict[str, Any],
        contexto: ContextoAnalise
    ) -> float:
        """
        Calcula a pontuação de risco combinando resultado das regras e análise comportamental.
        
        Args:
            resultado_regras: Resultado da execução das regras
            analise_comportamental: Análise comportamental do usuário
            contexto: Contexto da operação
            
        Returns:
            float: Pontuação de risco (0-100, onde 100 é o maior risco)
        """
        # Extrair pontuação base das regras (0-100)
        pontuacao_regras = resultado_regras.get("pontuacao_risco", 0)
        
        # Extrair pontuação comportamental (0-100)
        pontuacao_comportamental = analise_comportamental.get("pontuacao_risco", 0)
        
        # Calcular pontuação contextual com base em fatores geográficos e temporais
        pontuacao_contextual = self._calcular_pontuacao_contextual(contexto)
        
        # Pesos para cada componente (customizáveis por implementação específica)
        peso_regras = 0.5
        peso_comportamental = 0.3
        peso_contextual = 0.2
        
        # Calcular pontuação final ponderada
        pontuacao_final = (
            pontuacao_regras * peso_regras +
            pontuacao_comportamental * peso_comportamental +
            pontuacao_contextual * peso_contextual
        )
        
        # Garantir que a pontuação esteja no intervalo 0-100
        return max(0, min(100, pontuacao_final))
    
    def _calcular_pontuacao_contextual(self, contexto: ContextoAnalise) -> float:
        """
        Calcula a pontuação de risco contextual com base em fatores geográficos e temporais.
        
        Args:
            contexto: Contexto da operação
            
        Returns:
            float: Pontuação contextual de risco (0-100)
        """
        pontuacao = 0
        
        # Verificar se há metadados de configuração específica para o país
        pais_config = contexto.metadados.get("pais_config", {}) if contexto.metadados else {}
        
        # Fator de risco base do país (0-100)
        fator_risco_pais = pais_config.get("fator_risco_base", 50)
        
        # Ajustar pontuação com base no país
        pontuacao += fator_risco_pais * 0.3  # Peso de 30% para o fator país
        
        # Verificar horário da transação (maior risco em horários não comerciais)
        hora_atual = contexto.timestamp.hour
        if hora_atual < 7 or hora_atual > 22:
            pontuacao += 20  # Acrescenta 20 pontos para horários de maior risco
        
        # Verificar canal (canais presenciais têm menor risco)
        if contexto.canal == "presencial":
            pontuacao -= 15
        elif contexto.canal == "mobile":
            pontuacao += 5
        elif contexto.canal == "web":
            pontuacao += 10
        elif contexto.canal == "api":
            pontuacao += 15
        
        # Verificar valor da transação se for uma operação financeira
        if contexto.tipo_transacao and "financeira" in contexto.tipo_transacao and contexto.valor_transacao:
            # Configurações específicas de limites do tenant, se disponíveis
            limites = contexto.metadados.get("transacao_config", {}).get("limites", {}) if contexto.metadados else {}
            
            # Limite padrão se não houver configuração específica
            limite_medio = limites.get("medio", 1000)
            limite_alto = limites.get("alto", 5000)
            
            # Ajustar pontuação com base no valor
            if contexto.valor_transacao > limite_alto:
                pontuacao += 30
            elif contexto.valor_transacao > limite_medio:
                pontuacao += 15
        
        # Garantir que a pontuação esteja no intervalo 0-100
        return max(0, min(100, pontuacao))
    
    def _determinar_nivel_risco(self, pontuacao: float) -> str:
        """
        Determina o nível de risco com base na pontuação.
        
        Args:
            pontuacao: Pontuação de risco (0-100)
            
        Returns:
            str: Nível de risco (baixo, medio, alto, critico)
        """
        if pontuacao < 20:
            return "baixo"
        elif pontuacao < 50:
            return "medio"
        elif pontuacao < 80:
            return "alto"
        else:
            return "critico"
    
    def _determinar_resultado_avaliacao(
        self, 
        nivel_risco: str, 
        nivel_seguranca_padrao: str,
        contexto: ContextoAnalise
    ) -> ResultadoAvaliacao:
        """
        Determina o resultado da avaliação com base no nível de risco.
        
        Args:
            nivel_risco: Nível de risco calculado
            nivel_seguranca_padrao: Nível de segurança padrão do tenant
            contexto: Contexto da operação
            
        Returns:
            ResultadoAvaliacao: Resultado da avaliação
        """
        # Mapeamento de níveis de risco para índices numéricos
        mapa_niveis = {nivel: idx for idx, nivel in enumerate(NIVEIS_RISCO)}
        
        # Determinar nível de segurança efetivo (considerar tipo de transação)
        nivel_seguranca_efetivo = nivel_seguranca_padrao
        if contexto.metadados and "transacao_config" in contexto.metadados:
            nivel_seguranca_efetivo = contexto.metadados["transacao_config"].get(
                "nivel_seguranca",
                nivel_seguranca_padrao
            )
        
        # Comparar níveis numéricos para facilitar a lógica
        idx_risco = mapa_niveis.get(nivel_risco, 0)
        idx_seguranca = mapa_niveis.get(nivel_seguranca_efetivo, 0)
        
        # Determinar resultado com base na comparação dos níveis
        if idx_risco <= idx_seguranca - 2:  # Risco muito menor que o limite de segurança
            return ResultadoAvaliacao.PERMITIDO
        elif idx_risco <= idx_seguranca:  # Risco dentro do limite aceitável
            return ResultadoAvaliacao.VERIFICACAO_ADICIONAL
        else:  # Risco maior que o limite aceitável
            return ResultadoAvaliacao.BLOQUEADO
    
    def _gerar_acoes_recomendadas(
        self,
        nivel_risco: str,
        resultado: ResultadoAvaliacao,
        contexto: ContextoAnalise,
        tenant_config: Any
    ) -> List[str]:
        """
        Gera lista de ações recomendadas com base no resultado da avaliação.
        
        Args:
            nivel_risco: Nível de risco calculado
            resultado: Resultado da avaliação
            contexto: Contexto da operação
            tenant_config: Configuração do tenant
            
        Returns:
            List[str]: Lista de ações recomendadas
        """
        acoes = []
        
        if resultado == ResultadoAvaliacao.PERMITIDO:
            acoes.append("permitir_operacao")
            
            # Em caso de operações financeiras de valor elevado, recomendar notificação
            if (contexto.tipo_transacao and 
                "financeira" in contexto.tipo_transacao and 
                contexto.valor_transacao and 
                contexto.valor_transacao > 1000):
                acoes.append("notificar_usuario")
                
        elif resultado == ResultadoAvaliacao.VERIFICACAO_ADICIONAL:
            # Determinar quais fatores de autenticação adicionais solicitar
            if nivel_risco == "medio":
                acoes.append("solicitar_segundo_fator")
                acoes.append("verificar_dispositivo")
            else:  # alto
                acoes.append("solicitar_multiplos_fatores")
                acoes.append("verificar_localizacao")
                acoes.append("confirmar_intencao")
            
            acoes.append("notificar_usuario")
            
        else:  # BLOQUEADO
            acoes.append("bloquear_operacao")
            acoes.append("notificar_usuario")
            acoes.append("registrar_tentativa_suspeita")
            
            if nivel_risco == "critico":
                acoes.append("escalar_para_analise_manual")
                acoes.append("notificar_seguranca")
        
        return acoes
    
    def _identificar_fatores_contribuintes(
        self,
        resultado_regras: Dict[str, Any],
        analise_comportamental: Dict[str, Any]
    ) -> Dict[str, float]:
        """
        Identifica os fatores que mais contribuíram para a pontuação de risco.
        
        Args:
            resultado_regras: Resultado da execução das regras
            analise_comportamental: Análise comportamental do usuário
            
        Returns:
            Dict[str, float]: Dicionário de fatores e suas contribuições para o risco
        """
        fatores = {}
        
        # Extrair fatores das regras
        for regra in resultado_regras.get("regras_acionadas", []):
            if "contribuicao" in regra and "nome" in regra:
                fatores[f"regra:{regra['nome']}"] = regra["contribuicao"]
        
        # Extrair fatores comportamentais
        for anomalia in analise_comportamental.get("anomalias", []):
            if "contribuicao" in anomalia and "tipo" in anomalia:
                fatores[f"comportamento:{anomalia['tipo']}"] = anomalia["contribuicao"]
        
        return fatores
    
    def _criar_avaliacao_padrao_alto_risco(self, contexto: ContextoAnalise) -> AvaliacaoRisco:
        """
        Cria uma avaliação padrão de alto risco para casos de falha.
        
        Args:
            contexto: Contexto da operação
            
        Returns:
            AvaliacaoRisco: Avaliação de alto risco
        """
        return AvaliacaoRisco(
            nivel_risco="alto",
            pontuacao=75.0,
            resultado=ResultadoAvaliacao.VERIFICACAO_ADICIONAL,
            fatores_contribuintes={"erro_sistema": 75.0},
            acoes_recomendadas=[
                "solicitar_multiplos_fatores", 
                "verificar_dispositivo", 
                "confirmar_intencao",
                "notificar_usuario"
            ],
            detalhes={
                "erro": "Falha na avaliação de risco",
                "modo_fallback": True,
                "contexto_hash": self.context_provider.calcular_hash_contexto(contexto),
            }
        )
    
    def _registrar_metricas_resultado(
        self,
        avaliacao: AvaliacaoRisco,
        tenant_id: str,
        inicio: datetime.datetime
    ) -> None:
        """
        Registra métricas sobre o resultado da avaliação.
        
        Args:
            avaliacao: Resultado da avaliação de risco
            tenant_id: ID do tenant
            inicio: Timestamp do início da avaliação
        """
        duracao = (datetime.datetime.now() - inicio).total_seconds() * 1000
        
        self.metrics.recordValue(
            "trustguard.risk_analyzer.duration_ms",
            duracao,
            {"tenant_id": tenant_id}
        )
        
        self.metrics.incrementCounter(
            "trustguard.risk_analyzer.result",
            {
                "tenant_id": tenant_id,
                "nivel_risco": avaliacao.nivel_risco,
                "resultado": avaliacao.resultado.value
            }
        )
        
        self.logger.info(
            f"Avaliação de risco concluída: tenant={tenant_id}, "
            f"nivel={avaliacao.nivel_risco}, resultado={avaliacao.resultado.value}, "
            f"duracao={duracao:.2f}ms"
        )
