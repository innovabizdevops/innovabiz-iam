"""
Motor de Regras para TrustGuard

Este módulo implementa o motor de regras para execução de regras de negócio
e segurança específicas para cada tenant e contexto de avaliação.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional, Callable, Union
import datetime
import json
import re
from dataclasses import dataclass

from ...observability.core.multi_layer_monitor import MultiLayerMonitor


@dataclass
class Rule:
    """Definição de uma regra de avaliação."""
    id: str
    name: str
    description: str
    condition: str  # Expressão de condição (será avaliada dinamicamente)
    risk_contribution: float  # Contribuição para a pontuação de risco (0-100)
    tenant_id: Optional[str] = None  # None significa regra global
    market: Optional[str] = None  # Mercado específico ou None para todos
    tags: List[str] = None
    enabled: bool = True


class RuleEngine:
    """
    Motor de regras para avaliação de condições de segurança.
    
    Executa regras de negócio baseadas em expressões condicionais 
    sobre o contexto de avaliação.
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o motor de regras.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
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
        
        # Registro de funções auxiliares disponíveis para regras
        self.helper_functions = {
            "is_in": lambda value, array: value in array,
            "contains": lambda array, value: value in array,
            "starts_with": lambda s, prefix: s.startswith(prefix) if isinstance(s, str) else False,
            "ends_with": lambda s, suffix: s.endswith(suffix) if isinstance(s, str) else False,
            "matches_pattern": lambda s, pattern: bool(re.match(pattern, s)) if isinstance(s, str) else False,
            "is_business_hours": self._is_business_hours,
            "is_weekend": self._is_weekend,
            "is_high_risk_country": self._is_high_risk_country,
            "calculate_time_diff": self._calculate_time_diff,
            "get_value_or_default": lambda obj, key, default: obj.get(key, default) if isinstance(obj, dict) else default,
            "safe_divide": lambda a, b, default=0: a / b if b != 0 else default,
            "max_value": max,
            "min_value": min,
        }
        
        self.logger.info("RuleEngine inicializado")
    
    def executar_regras(self, regras_config: Dict[str, Any], contexto_avaliacao: Dict[str, Any]) -> Dict[str, Any]:
        """
        Executa as regras especificadas sobre o contexto de avaliação.
        
        Args:
            regras_config: Configuração das regras a serem executadas
            contexto_avaliacao: Contexto para avaliação das regras
            
        Returns:
            Dict[str, Any]: Resultado da execução das regras
        """
        try:
            inicio = datetime.datetime.now()
            
            self.logger.info("Iniciando execução de regras")
            
            # Extrair lista de regras da configuração
            regras = regras_config.get("regras", [])
            
            # Verificar se há regras para executar
            if not regras:
                self.logger.warn("Nenhuma regra encontrada para execução")
                return {
                    "pontuacao_risco": 0,
                    "regras_acionadas": [],
                    "total_regras": 0,
                    "total_acionadas": 0
                }
            
            # Preparar ambiente para execução das regras
            ambiente_execucao = self._preparar_ambiente_execucao(contexto_avaliacao)
            
            # Executar cada regra
            regras_acionadas = []
            pontuacao_total = 0
            
            for regra in regras:
                if not self._validar_regra(regra):
                    self.logger.warn(f"Regra inválida ignorada: {regra.get('id', 'desconhecida')}")
                    continue
                
                # Verificar se a regra está habilitada
                if not regra.get("enabled", True):
                    continue
                
                # Verificar se a regra é aplicável ao mercado atual
                mercado_regra = regra.get("market")
                mercado_contexto = contexto_avaliacao.get("contexto", {}).get("localizacao", {}).get("pais")
                if mercado_regra and mercado_contexto and mercado_regra != mercado_contexto:
                    continue
                
                # Avaliar condição da regra
                condicao = regra.get("condition", "False")
                
                try:
                    # Avaliar a condição no ambiente de execução
                    resultado = self._avaliar_expressao(condicao, ambiente_execucao)
                    
                    # Se a condição for verdadeira, a regra foi acionada
                    if resultado:
                        contribuicao = float(regra.get("risk_contribution", 0))
                        regras_acionadas.append({
                            "id": regra.get("id"),
                            "nome": regra.get("name", "Regra sem nome"),
                            "descricao": regra.get("description", ""),
                            "contribuicao": contribuicao,
                        })
                        pontuacao_total += contribuicao
                        
                        self.logger.info(f"Regra acionada: {regra.get('id')} - {regra.get('name')}")
                
                except Exception as e:
                    self.logger.error(f"Erro ao avaliar regra {regra.get('id')}: {str(e)}")
            
            # Normalizar pontuação para o intervalo 0-100
            pontuacao_normalizada = min(100, pontuacao_total)
            
            # Criar resultado
            resultado = {
                "pontuacao_risco": pontuacao_normalizada,
                "regras_acionadas": regras_acionadas,
                "total_regras": len(regras),
                "total_acionadas": len(regras_acionadas)
            }
            
            # Registrar métricas
            duracao = (datetime.datetime.now() - inicio).total_seconds() * 1000
            self.metrics.recordValue(
                "trustguard.rule_engine.duration_ms", 
                duracao, 
                {"tenant_id": contexto_avaliacao.get("contexto", {}).get("tenant_id", "desconhecido")}
            )
            
            self.metrics.recordValue(
                "trustguard.rule_engine.rules_triggered", 
                len(regras_acionadas),
                {"tenant_id": contexto_avaliacao.get("contexto", {}).get("tenant_id", "desconhecido")}
            )
            
            self.logger.info(
                f"Execução de regras concluída: {len(regras_acionadas)}/{len(regras)} regras acionadas, "
                f"pontuação: {pontuacao_normalizada:.2f}, duração: {duracao:.2f}ms"
            )
            
            return resultado
            
        except Exception as e:
            self.logger.error(f"Erro durante execução de regras: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.rule_engine.error", 
                {"error_type": type(e).__name__}
            )
            
            # Retornar resultado padrão em caso de erro
            return {
                "pontuacao_risco": 50,
                "regras_acionadas": [
                    {
                        "id": "erro",
                        "nome": "Erro na execução do motor de regras",
                        "descricao": str(e),
                        "contribuicao": 50
                    }
                ],
                "total_regras": 0,
                "total_acionadas": 0,
                "erro": True
            }
    
    def _validar_regra(self, regra: Dict[str, Any]) -> bool:
        """
        Valida se uma regra possui todos os campos obrigatórios.
        
        Args:
            regra: Definição da regra
            
        Returns:
            bool: True se a regra é válida
        """
        # Verificar campos obrigatórios
        if not regra.get("id"):
            return False
        if not regra.get("name"):
            return False
        if "condition" not in regra:
            return False
        if "risk_contribution" not in regra:
            return False
        
        return True
    
    def _preparar_ambiente_execucao(self, contexto_avaliacao: Dict[str, Any]) -> Dict[str, Any]:
        """
        Prepara o ambiente de execução para avaliação das regras.
        
        Args:
            contexto_avaliacao: Contexto para avaliação das regras
            
        Returns:
            Dict[str, Any]: Ambiente de execução
        """
        # Criar cópia do contexto para não modificar o original
        ambiente = {}
        
        # Adicionar cada chave do contexto ao ambiente
        for key, value in contexto_avaliacao.items():
            ambiente[key] = value
        
        # Adicionar funções auxiliares
        for func_name, func in self.helper_functions.items():
            ambiente[func_name] = func
        
        return ambiente
    
    def _avaliar_expressao(self, expressao: str, ambiente: Dict[str, Any]) -> bool:
        """
        Avalia uma expressão condicional no ambiente de execução.
        
        Args:
            expressao: Expressão condicional a ser avaliada
            ambiente: Ambiente de execução
            
        Returns:
            bool: Resultado da avaliação
        """
        # Avaliar a expressão no ambiente fornecido
        try:
            # Usar eval para avaliar a expressão no ambiente fornecido
            resultado = eval(expressao, {"__builtins__": {}}, ambiente)
            return bool(resultado)
        except Exception as e:
            self.logger.error(f"Erro ao avaliar expressão '{expressao}': {str(e)}")
            return False
    
    # Funções auxiliares para avaliação de regras
    
    def _is_business_hours(self, timestamp=None, start_hour=9, end_hour=18, business_days=(0, 1, 2, 3, 4)) -> bool:
        """
        Verifica se o timestamp está dentro do horário comercial.
        
        Args:
            timestamp: Timestamp a ser verificado (None para usar o timestamp atual)
            start_hour: Hora de início do horário comercial (padrão: 9)
            end_hour: Hora de fim do horário comercial (padrão: 18)
            business_days: Dias úteis (0=Segunda, 6=Domingo, padrão: 0-4)
            
        Returns:
            bool: True se estiver dentro do horário comercial
        """
        if timestamp is None:
            dt = datetime.datetime.now()
        elif isinstance(timestamp, str):
            try:
                dt = datetime.datetime.fromisoformat(timestamp)
            except:
                return False
        elif isinstance(timestamp, datetime.datetime):
            dt = timestamp
        else:
            return False
        
        # Verificar se é um dia útil
        if dt.weekday() not in business_days:
            return False
        
        # Verificar se está dentro do horário comercial
        return start_hour <= dt.hour < end_hour
    
    def _is_weekend(self, timestamp=None) -> bool:
        """
        Verifica se o timestamp é um fim de semana.
        
        Args:
            timestamp: Timestamp a ser verificado (None para usar o timestamp atual)
            
        Returns:
            bool: True se for fim de semana (sábado ou domingo)
        """
        if timestamp is None:
            dt = datetime.datetime.now()
        elif isinstance(timestamp, str):
            try:
                dt = datetime.datetime.fromisoformat(timestamp)
            except:
                return False
        elif isinstance(timestamp, datetime.datetime):
            dt = timestamp
        else:
            return False
        
        # 5=Sábado, 6=Domingo
        return dt.weekday() >= 5
    
    def _is_high_risk_country(self, country_code: str, threshold: int = 60) -> bool:
        """
        Verifica se um país é considerado de alto risco.
        
        Args:
            country_code: Código do país (ISO 3166-1 alpha-2)
            threshold: Limite de pontuação para considerar alto risco (padrão: 60)
            
        Returns:
            bool: True se o país for considerado de alto risco
        """
        # Importar constantes de risco de país
        from ...constants.trustguard_constants import RISCO_PAIS
        
        # Verificar se o país está na lista e seu risco é acima do threshold
        return RISCO_PAIS.get(country_code, 50) >= threshold
    
    def _calculate_time_diff(self, time1: Union[str, datetime.datetime], time2: Union[str, datetime.datetime] = None) -> int:
        """
        Calcula a diferença de tempo em minutos entre dois timestamps.
        
        Args:
            time1: Primeiro timestamp
            time2: Segundo timestamp (None para usar o timestamp atual)
            
        Returns:
            int: Diferença em minutos
        """
        # Converter time1 para datetime
        if isinstance(time1, str):
            try:
                dt1 = datetime.datetime.fromisoformat(time1)
            except:
                return 0
        elif isinstance(time1, datetime.datetime):
            dt1 = time1
        else:
            return 0
        
        # Converter time2 para datetime
        if time2 is None:
            dt2 = datetime.datetime.now()
        elif isinstance(time2, str):
            try:
                dt2 = datetime.datetime.fromisoformat(time2)
            except:
                return 0
        elif isinstance(time2, datetime.datetime):
            dt2 = time2
        else:
            return 0
        
        # Calcular diferença em minutos
        diff = dt2 - dt1
        return int(diff.total_seconds() / 60)
