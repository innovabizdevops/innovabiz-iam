"""
Analisador de Comportamento do Usuário

Este módulo implementa o analisador de comportamento de usuário para o TrustGuard,
permitindo detecção de anomalias em padrões de transações, logins, dispositivos e localização.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from typing import Dict, List, Any, Optional, Set, Tuple
import datetime
import json
import math
import uuid
from dataclasses import dataclass

from ...observability.core.multi_layer_monitor import MultiLayerMonitor
from ..contextual.multi_tenant_context_provider import ContextoAnalise


class UserBehaviorAnalyzer:
    """
    Analisador de comportamento do usuário.
    
    Detecta anomalias nos padrões de uso e transações do usuário,
    considerando contexto histórico, localização, dispositivo e horários de uso.
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o analisador de comportamento.
        
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
        
        # Inicializar repositório de comportamento (mock, deve ser substituído por implementação real)
        self.behavior_repository = None
        
        self.logger.info("UserBehaviorAnalyzer inicializado")

    def analyze_user_behavior(self, user_id: str, tenant_id: str, contexto: ContextoAnalise) -> Dict[str, Any]:
        """
        Analisa o comportamento do usuário para detectar anomalias.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            contexto: Contexto atual da operação
            
        Returns:
            Dict[str, Any]: Resultado da análise comportamental
        """
        try:
            # Registrar início da análise
            self.logger.info(f"Iniciando análise comportamental para usuário {user_id} do tenant {tenant_id}")
            self.metrics.incrementCounter("trustguard.behavior_analyzer.analysis", 
                                        {"tenant_id": tenant_id, "user_id": user_id})
            
            # Recuperar histórico comportamental do usuário (implementação mock)
            user_history = self._get_user_history(user_id, tenant_id)
            
            # Verificar padrões de localização
            location_score = self._analyze_location_pattern(user_history, contexto)
            
            # Verificar padrões de dispositivo
            device_score = self._analyze_device_pattern(user_history, contexto)
            
            # Verificar padrões de horário
            time_score = self._analyze_time_pattern(user_history, contexto)
            
            # Verificar padrões de transação (se aplicável)
            transaction_score = 0
            if contexto.tipo_transacao and contexto.valor_transacao:
                transaction_score = self._analyze_transaction_pattern(user_history, contexto)
            
            # Calcular pontuação geral e anomalias
            anomalies = []
            anomaly_weights = {
                "location": 0.3,
                "device": 0.2,
                "time": 0.1,
                "transaction": 0.4
            }
            
            # Identificar anomalias específicas
            if location_score > 50:
                anomalies.append({
                    "tipo": "localização_incomum",
                    "severidade": self._map_score_to_severity(location_score),
                    "pontuacao": location_score,
                    "contribuicao": location_score * anomaly_weights["location"]
                })
                
            if device_score > 50:
                anomalies.append({
                    "tipo": "dispositivo_desconhecido",
                    "severidade": self._map_score_to_severity(device_score),
                    "pontuacao": device_score,
                    "contribuicao": device_score * anomaly_weights["device"]
                })
                
            if time_score > 50:
                anomalies.append({
                    "tipo": "horario_incomum",
                    "severidade": self._map_score_to_severity(time_score),
                    "pontuacao": time_score,
                    "contribuicao": time_score * anomaly_weights["time"]
                })
                
            if transaction_score > 50:
                anomalies.append({
                    "tipo": "transacao_incomum",
                    "severidade": self._map_score_to_severity(transaction_score),
                    "pontuacao": transaction_score,
                    "contribuicao": transaction_score * anomaly_weights["transaction"]
                })
            
            # Calcular pontuação final ponderada
            risk_score = (
                location_score * anomaly_weights["location"] +
                device_score * anomaly_weights["device"] +
                time_score * anomaly_weights["time"]
            )
            
            if transaction_score > 0:
                risk_score += transaction_score * anomaly_weights["transaction"]
            else:
                # Normalizar para caso de não ter transação
                risk_score = risk_score / (1 - anomaly_weights["transaction"])
            
            # Criar resultado da análise
            resultado = {
                "pontuacao_risco": min(100, risk_score),
                "anomalias": anomalies,
                "perfil_comportamental": {
                    "localizacoes_frequentes": self._get_frequent_locations(user_history),
                    "dispositivos_conhecidos": self._get_known_devices(user_history),
                    "horarios_comuns": self._get_common_times(user_history),
                    "padrao_transacional": self._get_transaction_pattern(user_history)
                },
                "confianca_analise": self._calculate_confidence(user_history)
            }
            
            # Registrar resultado
            self.logger.info(
                f"Análise comportamental concluída: usuario={user_id}, "
                f"pontuacao_risco={resultado['pontuacao_risco']:.2f}, "
                f"anomalias={len(anomalies)}"
            )
            
            self.metrics.recordValue(
                "trustguard.behavior_analyzer.risk_score",
                resultado["pontuacao_risco"],
                {"tenant_id": tenant_id, "user_id": user_id}
            )
            
            return resultado
            
        except Exception as e:
            self.logger.error(f"Erro durante análise comportamental: {str(e)}")
            self.metrics.incrementCounter(
                "trustguard.behavior_analyzer.error", 
                {"tenant_id": tenant_id, "user_id": user_id, "error_type": type(e).__name__}
            )
            
            # Retornar resultado padrão em caso de erro
            return {
                "pontuacao_risco": 50,  # Risco médio como fallback
                "anomalias": [
                    {
                        "tipo": "erro_analise",
                        "severidade": "medio",
                        "pontuacao": 50,
                        "contribuicao": 50
                    }
                ],
                "confianca_analise": 0.5
            }
    
    def _get_user_history(self, user_id: str, tenant_id: str) -> Dict[str, Any]:
        """
        Obtém o histórico comportamental do usuário.
        
        Nota: Esta é uma implementação mock para fins de desenvolvimento.
        Na implementação real, esta função consultaria um repositório de dados.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            
        Returns:
            Dict[str, Any]: Histórico comportamental do usuário
        """
        # Simular histórico básico para desenvolvimento
        # Em produção, isso seria substituído por uma consulta ao banco de dados
        
        # Gerar um histórico pseudoaleatório mas consistente para o mesmo usuário
        seed = sum(ord(c) for c in user_id)
        
        # Lista de dispositivos conhecidos
        devices = [
            {"hash": f"device_{user_id[-4:]}_{i}", "nome": f"Dispositivo {i+1}", "confianca": 0.9 - (i * 0.2)} 
            for i in range(min(3, seed % 4 + 1))
        ]
        
        # Lista de localizações conhecidas
        paises = ["AO", "PT", "BR", "MZ", "US"]
        cidades = {
            "AO": ["Luanda", "Benguela", "Huambo", "Lubango"],
            "PT": ["Lisboa", "Porto", "Faro"],
            "BR": ["São Paulo", "Rio de Janeiro", "Brasília"],
            "MZ": ["Maputo", "Beira", "Nampula"],
            "US": ["New York", "Miami", "Los Angeles"]
        }
        
        user_countries = [paises[seed % len(paises)], paises[(seed + 1) % len(paises)]]
        locations = []
        
        for pais in user_countries:
            for i in range(min(2, (seed + ord(pais[0])) % 3 + 1)):
                cidade = cidades[pais][i % len(cidades[pais])]
                locations.append({
                    "pais": pais,
                    "cidade": cidade,
                    "frequencia": 0.8 - (i * 0.3),
                    "ultima_vez": (datetime.datetime.now() - datetime.timedelta(days=i*3)).isoformat()
                })
        
        # Padrões de horário (UTC)
        time_patterns = []
        base_hour = (seed % 12) + 8  # Hora base entre 8 e 19
        
        time_patterns.append({
            "hora_inicio": base_hour,
            "hora_fim": base_hour + 2,
            "dias_semana": [1, 2, 3, 4, 5],  # Segunda a sexta
            "frequencia": 0.8
        })
        
        time_patterns.append({
            "hora_inicio": (base_hour + 6) % 24,
            "hora_fim": (base_hour + 8) % 24,
            "dias_semana": [6, 7],  # Fim de semana
            "frequencia": 0.5
        })
        
        # Padrões transacionais
        transaction_patterns = {
            "valor_medio": 100 * (seed % 10 + 1),
            "valor_maximo": 500 * (seed % 10 + 1),
            "frequencia_diaria": (seed % 5) + 1,
            "categorias_comuns": ["supermercado", "restaurante", "transporte", "entretenimento"][:2 + (seed % 3)],
            "destinos_comuns": [f"dest_{i}" for i in range((seed % 3) + 1)]
        }
        
        return {
            "usuario_id": user_id,
            "tenant_id": tenant_id,
            "dispositivos": devices,
            "localizacoes": locations,
            "padroes_horario": time_patterns,
            "padroes_transacao": transaction_patterns,
            "primeiro_acesso": (datetime.datetime.now() - datetime.timedelta(days=90 + (seed % 90))).isoformat(),
            "ultimo_acesso": (datetime.datetime.now() - datetime.timedelta(hours=(seed % 24))).isoformat(),
            "total_acessos": 50 + (seed % 200),
            "total_transacoes": 20 + (seed % 100)
        }
    
    def _analyze_location_pattern(self, history: Dict[str, Any], contexto: ContextoAnalise) -> float:
        """
        Analisa padrões de localização do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            contexto: Contexto atual da operação
            
        Returns:
            float: Pontuação de anomalia de localização (0-100, onde 100 é mais anômalo)
        """
        # Se não houver informação de localização no contexto atual, retorna pontuação média
        if not contexto.localizacao or "pais" not in contexto.localizacao:
            return 50
        
        # Verificar se o país atual está entre os países conhecidos
        current_country = contexto.localizacao.get("pais")
        current_city = contexto.localizacao.get("cidade", "")
        
        known_locations = history.get("localizacoes", [])
        
        # Se não houver histórico de localizações, retorna pontuação alta
        if not known_locations:
            return 80
        
        # Verificar correspondência exata (país e cidade)
        for location in known_locations:
            if (location["pais"] == current_country and 
                location.get("cidade", "") == current_city and
                location.get("cidade", "")):
                # Localização conhecida e exata
                return 10
        
        # Verificar só correspondência de país
        for location in known_locations:
            if location["pais"] == current_country:
                # País conhecido, cidade nova
                return 40
        
        # País desconhecido - verificar se é um país vizinho/relacionado
        # Esta é uma simplificação. Uma implementação real usaria uma API de geolocalização
        
        # Mapeamento simples de países "relacionados" para fins de demonstração
        related_countries = {
            "AO": ["NA", "CD", "CG", "ZM"],  # Angola -> países vizinhos
            "MZ": ["ZA", "ZW", "TZ", "MW", "SZ"],  # Moçambique -> países vizinhos
            "BR": ["AR", "UY", "PY", "BO", "PE", "CO"],  # Brasil -> países vizinhos
            "PT": ["ES", "FR"],  # Portugal -> países vizinhos
            # etc.
        }
        
        # Verificar se o país atual está relacionado com algum país conhecido
        for location in known_locations:
            known_country = location["pais"]
            if known_country in related_countries and current_country in related_countries[known_country]:
                # País relacionado a um país conhecido
                return 65
        
        # Localização completamente nova e não relacionada
        return 90
    
    def _analyze_device_pattern(self, history: Dict[str, Any], contexto: ContextoAnalise) -> float:
        """
        Analisa padrões de dispositivo do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            contexto: Contexto atual da operação
            
        Returns:
            float: Pontuação de anomalia de dispositivo (0-100, onde 100 é mais anômalo)
        """
        # Se não houver hash do dispositivo no contexto atual, retorna pontuação média
        if not contexto.dispositivo_hash:
            return 50
        
        known_devices = history.get("dispositivos", [])
        
        # Se não houver histórico de dispositivos, retorna pontuação alta
        if not known_devices:
            return 80
        
        # Verificar se o dispositivo atual está entre os dispositivos conhecidos
        for device in known_devices:
            if device["hash"] == contexto.dispositivo_hash:
                # Dispositivo conhecido
                return 10 * (1 - device.get("confianca", 0.5))  # Menor pontuação para dispositivos mais confiáveis
        
        # Dispositivo desconhecido
        return 85
    
    def _analyze_time_pattern(self, history: Dict[str, Any], contexto: ContextoAnalise) -> float:
        """
        Analisa padrões de horário de acesso do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            contexto: Contexto atual da operação
            
        Returns:
            float: Pontuação de anomalia de horário (0-100, onde 100 é mais anômalo)
        """
        current_time = contexto.timestamp
        current_hour = current_time.hour
        current_day = current_time.weekday() + 1  # 1 = Segunda, 7 = Domingo
        
        time_patterns = history.get("padroes_horario", [])
        
        # Se não houver padrões de horário conhecidos, retorna pontuação média
        if not time_patterns:
            return 50
        
        # Verificar se o horário atual está dentro de algum padrão conhecido
        for pattern in time_patterns:
            hora_inicio = pattern.get("hora_inicio", 0)
            hora_fim = pattern.get("hora_fim", 24)
            dias = pattern.get("dias_semana", list(range(1, 8)))
            
            if (hora_inicio <= current_hour < hora_fim and current_day in dias):
                # Horário dentro de um padrão conhecido
                return 10 * (1 - pattern.get("frequencia", 0.5))
        
        # Verificar se o horário está próximo de um padrão conhecido
        for pattern in time_patterns:
            hora_inicio = pattern.get("hora_inicio", 0)
            hora_fim = pattern.get("hora_fim", 24)
            dias = pattern.get("dias_semana", list(range(1, 8)))
            
            # Horário próximo (até 2h de diferença)
            if ((abs(hora_inicio - current_hour) <= 2 or abs(hora_fim - current_hour) <= 2) 
                    and current_day in dias):
                return 40
            
            # Dia certo, horário errado
            if current_day in dias:
                return 60
        
        # Horário completamente fora do padrão
        return 80
    
    def _analyze_transaction_pattern(self, history: Dict[str, Any], contexto: ContextoAnalise) -> float:
        """
        Analisa padrões de transação do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            contexto: Contexto atual da operação
            
        Returns:
            float: Pontuação de anomalia de transação (0-100, onde 100 é mais anômalo)
        """
        # Se não houver informações de transação no contexto, retorna 0 (não aplicável)
        if not contexto.tipo_transacao or contexto.valor_transacao is None:
            return 0
        
        transaction_patterns = history.get("padroes_transacao", {})
        
        # Se não houver padrões de transação conhecidos, retorna pontuação alta
        if not transaction_patterns:
            return 75
        
        current_value = contexto.valor_transacao
        avg_value = transaction_patterns.get("valor_medio", 100)
        max_value = transaction_patterns.get("valor_maximo", 500)
        
        # Verificar o valor da transação
        if current_value > max_value * 3:
            # Valor extremamente alto comparado ao histórico
            return 95
        elif current_value > max_value:
            # Valor acima do máximo histórico
            return 80
        elif current_value > avg_value * 2:
            # Valor significativamente acima da média
            return 60
        elif current_value > avg_value * 1.5:
            # Valor moderadamente acima da média
            return 40
        else:
            # Valor dentro do padrão normal
            return 20
    
    def _map_score_to_severity(self, score: float) -> str:
        """
        Mapeia uma pontuação numérica para um nível de severidade.
        
        Args:
            score: Pontuação numérica (0-100)
            
        Returns:
            str: Nível de severidade (baixo, medio, alto, critico)
        """
        if score < 30:
            return "baixo"
        elif score < 60:
            return "medio"
        elif score < 85:
            return "alto"
        else:
            return "critico"
    
    def _get_frequent_locations(self, history: Dict[str, Any]) -> List[Dict[str, Any]]:
        """
        Extrai as localizações mais frequentes do histórico do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            
        Returns:
            List[Dict[str, Any]]: Lista de localizações frequentes
        """
        locations = history.get("localizacoes", [])
        return sorted(locations, key=lambda x: x.get("frequencia", 0), reverse=True)
    
    def _get_known_devices(self, history: Dict[str, Any]) -> List[Dict[str, Any]]:
        """
        Extrai os dispositivos conhecidos do histórico do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            
        Returns:
            List[Dict[str, Any]]: Lista de dispositivos conhecidos
        """
        return history.get("dispositivos", [])
    
    def _get_common_times(self, history: Dict[str, Any]) -> List[Dict[str, Any]]:
        """
        Extrai os horários comuns de acesso do histórico do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            
        Returns:
            List[Dict[str, Any]]: Lista de padrões de horário
        """
        return history.get("padroes_horario", [])
    
    def _get_transaction_pattern(self, history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Extrai o padrão transacional do histórico do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            
        Returns:
            Dict[str, Any]: Padrão transacional
        """
        return history.get("padroes_transacao", {})
    
    def _calculate_confidence(self, history: Dict[str, Any]) -> float:
        """
        Calcula o nível de confiança da análise com base no histórico do usuário.
        
        Args:
            history: Histórico comportamental do usuário
            
        Returns:
            float: Nível de confiança (0-1)
        """
        # Verificar quantidade de acessos e transações
        total_acessos = history.get("total_acessos", 0)
        total_transacoes = history.get("total_transacoes", 0)
        
        # Calcular idade do perfil em dias
        try:
            primeiro_acesso = datetime.datetime.fromisoformat(history.get("primeiro_acesso", ""))
            idade_perfil = (datetime.datetime.now() - primeiro_acesso).days
        except:
            idade_perfil = 0
        
        # Fator de confiança baseado no volume de dados históricos
        confianca_volume = min(1.0, (total_acessos + total_transacoes) / 100)
        
        # Fator de confiança baseado na idade do perfil
        confianca_idade = min(1.0, idade_perfil / 30)
        
        # Fator de confiança baseado na diversidade de dados
        diversidade = (
            min(1.0, len(history.get("dispositivos", [])) / 2) * 0.3 +
            min(1.0, len(history.get("localizacoes", [])) / 2) * 0.3 +
            min(1.0, len(history.get("padroes_horario", [])) / 2) * 0.2 +
            (0.2 if history.get("padroes_transacao") else 0)
        )
        
        # Calcular confiança final
        confianca_final = (
            confianca_volume * 0.4 +
            confianca_idade * 0.3 +
            diversidade * 0.3
        )
        
        return confianca_final
