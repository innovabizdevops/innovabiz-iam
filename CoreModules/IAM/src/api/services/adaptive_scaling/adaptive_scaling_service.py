"""
Serviço de escalonamento adaptativo baseado em TrustScore.

Este serviço implementa os mecanismos de adaptação dinâmica dos níveis 
de segurança com base nas pontuações de confiança dos usuários.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import uuid
import logging
import asyncio
from typing import Dict, List, Any, Optional, Tuple, Set, Union
from datetime import datetime, timedelta
import asyncpg

from opentelemetry import trace
from opentelemetry.trace.status import Status, StatusCode

from ....app.repositories.trust_score_repository import TrustScoreRepository
from ....app.services.trust_score_query_service import TrustScoreQueryService
from ....app.trust_guard_models import TrustScoreResult, DetectedAnomaly

from .models import (
    ScalingDirection,
    SecurityLevel,
    SecurityMechanism,
    ScalingTrigger, 
    ScalingPolicy,
    ScalingEvent,
    SecurityAdjustment,
    AdaptiveConfig
)

from ..notification.notification_service import NotificationService


# Configuração de logging
logger = logging.getLogger(__name__)

# Tracer para OpenTelemetry
tracer = trace.get_tracer(__name__)


class AdaptiveScalingService:
    """
    Serviço que gerencia o escalonamento adaptativo de segurança
    baseado nas pontuações de confiança dos usuários.
    """
    
    def __init__(
        self,
        db_pool: asyncpg.Pool,
        trust_repository: TrustScoreRepository,
        trust_query_service: TrustScoreQueryService,
        notification_service: Optional[NotificationService] = None
    ):
        """
        Inicializa o serviço de escalonamento adaptativo.
        
        Args:
            db_pool: Pool de conexões de banco de dados
            trust_repository: Repositório para dados de confiança
            trust_query_service: Serviço de consulta otimizado
            notification_service: Serviço de notificação (opcional)
        """
        self.db_pool = db_pool
        self.trust_repository = trust_repository
        self.trust_query_service = trust_query_service
        self.notification_service = notification_service
        self.config = AdaptiveConfig()  # Configuração padrão
        self._policy_cache = {}
        self._trigger_cache = {}
        self._last_evaluation = {}  # Cache de última avaliação por usuário
    
    async def initialize(self) -> None:
        """
        Inicializa o serviço, carregando configurações e caches.
        """
        with tracer.start_as_current_span("adaptive_scaling.initialize") as span:
            try:
                # Carregar configuração
                config = await self._load_config()
                if config:
                    self.config = config
                
                # Carregar gatilhos e políticas em cache
                await self._refresh_cache()
                
                logger.info("Serviço de escalonamento adaptativo inicializado")
                span.set_status(Status(StatusCode.OK))
            except Exception as e:
                logger.error(f"Erro ao inicializar serviço de escalonamento adaptativo: {e}")
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                raise
    
    async def evaluate_trust_score(
        self,
        trust_score_result: TrustScoreResult
    ) -> Optional[ScalingEvent]:
        """
        Avalia uma pontuação de confiança e aplica escalonamento adaptativo se necessário.
        
        Args:
            trust_score_result: Resultado da avaliação de confiança
            
        Returns:
            Optional[ScalingEvent]: Evento de escalonamento gerado, se houver
        """
        with tracer.start_as_current_span("adaptive_scaling.evaluate_trust_score") as span:
            try:
                # Verificar se serviço está ativo
                if not self.config.enabled:
                    logger.debug("Serviço de escalonamento adaptativo desabilitado")
                    return None
                
                # Extrair informações do resultado
                user_id = trust_score_result.user_id
                tenant_id = trust_score_result.tenant_id
                context_id = trust_score_result.context_id
                region_code = trust_score_result.regional_context
                
                # Adicionar atributos ao span para rastreamento
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                if context_id:
                    span.set_attribute("context.id", context_id)
                if region_code:
                    span.set_attribute("region.code", region_code)
                
                # Verificar gatilhos aplicáveis
                triggered_policies = await self._check_triggers(trust_score_result)
                
                # Se não houver gatilhos acionados, não fazer nada
                if not triggered_policies:
                    logger.debug(f"Nenhum gatilho acionado para usuário {user_id}")
                    return None
                
                # Selecionar política de maior prioridade
                policy_id, trigger_id, direction = self._select_highest_priority_policy(triggered_policies)
                
                # Buscar política no cache
                policy = self._policy_cache.get(policy_id)
                if not policy:
                    logger.warning(f"Política {policy_id} não encontrada em cache")
                    return None
                
                # Buscar ajustes de segurança a serem aplicados
                adjustments = await self._determine_security_adjustments(
                    trust_score_result, policy, direction
                )
                
                # Se não houver ajustes, não fazer nada
                if not adjustments:
                    logger.debug(f"Nenhum ajuste necessário para usuário {user_id}")
                    return None
                
                # Criar e registrar evento de escalonamento
                event = await self._create_scaling_event(
                    trust_score_result=trust_score_result,
                    policy_id=policy_id,
                    trigger_id=trigger_id,
                    direction=direction,
                    adjustments=adjustments
                )
                
                # Aplicar ajustes de segurança
                await self._apply_security_adjustments(event)
                
                # Notificar usuário se configurado
                if self.config.notify_user_on_change and self.notification_service:
                    await self._notify_user_of_changes(event)
                
                logger.info(f"Escalonamento aplicado para usuário {user_id}: direção={direction.value}, gatilho={trigger_id}")
                return event
            except Exception as e:
                logger.error(f"Erro ao avaliar pontuação para escalonamento: {e}")
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                return None
    
    async def get_current_security_level(
        self,
        user_id: str,
        tenant_id: str,
        mechanism: SecurityMechanism,
        context_id: Optional[str] = None
    ) -> SecurityLevel:
        """
        Obtém o nível de segurança atual de um mecanismo para um usuário.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            mechanism: Mecanismo de segurança
            context_id: ID do contexto (opcional)
            
        Returns:
            SecurityLevel: Nível de segurança atual
        """
        with tracer.start_as_current_span("adaptive_scaling.get_current_security_level") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                span.set_attribute("security.mechanism", mechanism.value)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                async with self.db_pool.acquire() as conn:
                    query = """
                        SELECT level FROM security_levels
                        WHERE user_id = $1 AND tenant_id = $2 AND mechanism = $3
                        AND (context_id = $4 OR context_id IS NULL)
                        ORDER BY context_id NULLS LAST
                        LIMIT 1
                    """
                    
                    row = await conn.fetchrow(query, user_id, tenant_id, mechanism.value, context_id)
                    
                    if row:
                        return SecurityLevel(row['level'])
                    else:
                        # Retornar nível padrão
                        default_level = await self._get_default_security_level(tenant_id, mechanism, context_id)
                        return default_level
            except Exception as e:
                logger.error(f"Erro ao obter nível de segurança: {e}")
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                # Retornar nível padrão em caso de erro
                return SecurityLevel.STANDARD
    
    async def get_user_security_profile(
        self, 
        user_id: str, 
        tenant_id: str,
        context_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Obtém o perfil completo de segurança de um usuário.
        
        Args:
            user_id: ID do usuário
            tenant_id: ID do tenant
            context_id: ID do contexto (opcional)
            
        Returns:
            Dict: Perfil de segurança com níveis por mecanismo
        """
        with tracer.start_as_current_span("adaptive_scaling.get_user_security_profile") as span:
            try:
                # Adicionar atributos ao span
                span.set_attribute("user.id", user_id)
                span.set_attribute("tenant.id", tenant_id)
                if context_id:
                    span.set_attribute("context.id", context_id)
                
                # Buscar todos os níveis de segurança do usuário
                async with self.db_pool.acquire() as conn:
                    query = """
                        SELECT mechanism, level, context_id, updated_at, expires_at, metadata
                        FROM security_levels
                        WHERE user_id = $1 AND tenant_id = $2
                        AND (context_id = $3 OR context_id IS NULL)
                    """
                    
                    rows = await conn.fetch(query, user_id, tenant_id, context_id)
                    
                    # Organizar resultados
                    profile = {
                        "user_id": user_id,
                        "tenant_id": tenant_id,
                        "context_id": context_id,
                        "security_levels": {},
                        "last_modified": None,
                        "scaling_events": []
                    }
                    
                    for row in rows:
                        mechanism = row['mechanism']
                        level = row['level']
                        ctx = row['context_id'] or "default"
                        updated = row['updated_at']
                        expires = row['expires_at']
                        metadata = row['metadata'] if row['metadata'] else {}
                        
                        if ctx not in profile["security_levels"]:
                            profile["security_levels"][ctx] = {}
                        
                        profile["security_levels"][ctx][mechanism] = {
                            "level": level,
                            "updated_at": updated,
                            "expires_at": expires,
                            "metadata": metadata
                        }
                        
                        # Rastrear última modificação
                        if not profile["last_modified"] or updated > profile["last_modified"]:
                            profile["last_modified"] = updated
                    
                    # Buscar eventos de escalonamento recentes
                    query_events = """
                        SELECT id, trigger_id, policy_id, scaling_direction, event_time, expires_at
                        FROM scaling_events
                        WHERE user_id = $1 AND tenant_id = $2
                        AND (context_id = $3 OR context_id IS NULL)
                        ORDER BY event_time DESC
                        LIMIT 5
                    """
                    
                    event_rows = await conn.fetch(query_events, user_id, tenant_id, context_id)
                    profile["scaling_events"] = [dict(row) for row in event_rows]
                    
                    return profile
            except Exception as e:
                logger.error(f"Erro ao obter perfil de segurança: {e}")
                span.set_status(Status(StatusCode.ERROR, str(e)))
                span.record_exception(e)
                return {
                    "user_id": user_id,
                    "tenant_id": tenant_id,
                    "context_id": context_id,
                    "security_levels": {},
                    "error": str(e)
                }
                
    async def _load_config(self) -> Optional[AdaptiveConfig]:
        """Carrega a configuração do banco de dados."""
        try:
            async with self.db_pool.acquire() as conn:
                row = await conn.fetchrow("SELECT config FROM adaptive_scaling_config LIMIT 1")
                if row and row['config']:
                    return AdaptiveConfig(**json.loads(row['config']))
                return None
        except Exception as e:
            logger.error(f"Erro ao carregar configuração de escalonamento: {e}")
            return None
    
    async def _refresh_cache(self) -> None:
        """Atualiza os caches de gatilhos e políticas."""
        try:
            # Atualizar cache de gatilhos
            async with self.db_pool.acquire() as conn:
                trigger_rows = await conn.fetch("SELECT * FROM scaling_triggers WHERE enabled = true")
                
                self._trigger_cache = {}
                for row in trigger_rows:
                    trigger = ScalingTrigger(**dict(row))
                    self._trigger_cache[trigger.id] = trigger
                
                # Atualizar cache de políticas
                policy_rows = await conn.fetch("SELECT * FROM scaling_policies WHERE enabled = true")
                
                self._policy_cache = {}
                for row in policy_rows:
                    policy_dict = dict(row)
                    # Converter campos JSON
                    policy_dict['trigger_ids'] = json.loads(policy_dict['trigger_ids'])
                    policy_dict['adjustment_map'] = json.loads(policy_dict['adjustment_map'])
                    
                    policy = ScalingPolicy(**policy_dict)
                    self._policy_cache[policy.id] = policy
                
            logger.info(f"Cache atualizado: {len(self._trigger_cache)} gatilhos, {len(self._policy_cache)} políticas")
        except Exception as e:
            logger.error(f"Erro ao atualizar cache: {e}")
    
    async def _check_triggers(
        self, 
        trust_score_result: TrustScoreResult
    ) -> List[Tuple[str, str, ScalingDirection]]:
        """
        Verifica quais gatilhos foram acionados para uma pontuação de confiança.
        
        Args:
            trust_score_result: Resultado da avaliação de confiança
            
        Returns:
            List[Tuple[str, str, ScalingDirection]]: Lista de tuplas (policy_id, trigger_id, direction)
        """
        # Extrair dados relevantes
        user_id = trust_score_result.user_id
        tenant_id = trust_score_result.tenant_id
        context_id = trust_score_result.context_id
        region_code = trust_score_result.regional_context
        score = trust_score_result.overall_score
        dimension_scores = trust_score_result.dimension_scores
        anomalies = trust_score_result.anomalies or []
        
        # Chave para cache
        user_key = f"{user_id}:{tenant_id}:{context_id or 'default'}"
        
        # Verificar se estamos em período de cooldown
        last_eval = self._last_evaluation.get(user_key, {})
        if last_eval:
            cooldown_mins = self.config.default_cooldown_minutes
            last_time = last_eval.get("time")
            if last_time and (datetime.now() - last_time).total_seconds() < (cooldown_mins * 60):
                logger.debug(f"Usuário {user_id} em cooldown, pulando avaliação")
                return []
        
        # Lista para armazenar gatilhos acionados
        triggered = []
        
        # Verificar cada gatilho disponível
        for trigger_id, trigger in self._trigger_cache.items():
            # Ignorar gatilhos desabilitados
            if not trigger.enabled:
                continue
                
            # Verificar especificidade de tenant
            if trigger.tenant_specific and trigger.tenant_id != tenant_id:
                continue
                
            # Verificar especificidade de região
            if trigger.region_specific and trigger.region_code != region_code:
                continue
                
            # Verificar especificidade de contexto
            if trigger.context_specific and trigger.context_id != context_id:
                continue
            
            # Verificar condição
            triggered_value = None
            comparison_value = trigger.threshold_value
            
            # Verificar tipo de condição
            if trigger.condition_type == "threshold":
                # Condição baseada na pontuação geral ou de dimensão específica
                if trigger.dimension == "overall":
                    triggered_value = score
                else:
                    # Buscar pontuação da dimensão específica
                    triggered_value = dimension_scores.get(trigger.dimension.lower())
            
            elif trigger.condition_type == "delta":
                # Condição baseada na variação de pontuação
                prev_score = last_eval.get("scores", {}).get(trigger.dimension.lower() 
                    if trigger.dimension != "overall" else "overall")
                
                if prev_score is not None:
                    current = score if trigger.dimension == "overall" else dimension_scores.get(trigger.dimension.lower())
                    if current is not None:
                        triggered_value = current - prev_score
            
            elif trigger.condition_type == "anomaly":
                # Condição baseada na presença de anomalias
                anomaly_count = sum(1 for a in anomalies 
                                   if (not trigger.dimension or a.affected_dimensions and 
                                      trigger.dimension.lower() in [d.lower() for d in a.affected_dimensions]))
                triggered_value = anomaly_count
            
            # Se não temos valor para comparar, continuar para o próximo gatilho
            if triggered_value is None:
                continue
            
            # Aplicar comparação
            is_triggered = False
            
            if trigger.comparison == "lt":
                is_triggered = triggered_value < comparison_value
            elif trigger.comparison == "lte":
                is_triggered = triggered_value <= comparison_value
            elif trigger.comparison == "gt":
                is_triggered = triggered_value > comparison_value
            elif trigger.comparison == "gte":
                is_triggered = triggered_value >= comparison_value
            elif trigger.comparison == "eq":
                is_triggered = abs(triggered_value - comparison_value) < 0.001  # Para comparação de floats
                
            if is_triggered:
                # Encontrar políticas associadas a este gatilho
                for policy_id, policy in self._policy_cache.items():
                    # Verificar se o gatilho está associado à política
                    if trigger_id in policy.trigger_ids:
                        # Verificar especificidade de tenant na política
                        if policy.tenant_id and policy.tenant_id != tenant_id:
                            continue
                            
                        # Verificar especificidade de região na política
                        if policy.region_code and policy.region_code != region_code:
                            continue
                            
                        # Verificar especificidade de contexto na política
                        if policy.context_id and policy.context_id != context_id:
                            continue
                            
                        # Adicionar à lista de acionados
                        triggered.append((policy_id, trigger_id, trigger.scaling_direction))
                        
        # Atualizar cache de última avaliação
        self._last_evaluation[user_key] = {
            "time": datetime.now(),
            "scores": {
                "overall": score,
                **{k.lower(): v for k, v in dimension_scores.items()}
            },
            "anomalies": len(anomalies)
        }
        
        return triggered
    
    def _select_highest_priority_policy(
        self, 
        triggered_policies: List[Tuple[str, str, ScalingDirection]]
    ) -> Tuple[str, str, ScalingDirection]:
        """
        Seleciona a política de maior prioridade entre as acionadas.
        
        Args:
            triggered_policies: Lista de políticas acionadas
            
        Returns:
            Tuple[str, str, ScalingDirection]: Tupla com (policy_id, trigger_id, direction)
        """
        highest_policy = None
        highest_priority = -1
        highest_trigger = None
        highest_direction = None
        
        for policy_id, trigger_id, direction in triggered_policies:
            # Obter política do cache
            policy = self._policy_cache.get(policy_id)
            if not policy:
                continue
                
            # Verificar prioridade
            if policy.priority > highest_priority:
                highest_priority = policy.priority
                highest_policy = policy_id
                highest_trigger = trigger_id
                highest_direction = direction
        
        return (highest_policy, highest_trigger, highest_direction)
    
    async def _determine_security_adjustments(
        self,
        trust_score_result: TrustScoreResult,
        policy: ScalingPolicy,
        direction: ScalingDirection
    ) -> List[SecurityAdjustment]:
        """
        Determina os ajustes de segurança a serem aplicados com base na política.
        
        Args:
            trust_score_result: Resultado da avaliação de confiança
            policy: Política a ser aplicada
            direction: Direção do escalonamento
            
        Returns:
            List[SecurityAdjustment]: Lista de ajustes de segurança a serem aplicados
        """
        # Extrair dados relevantes
        user_id = trust_score_result.user_id
        tenant_id = trust_score_result.tenant_id
        context_id = trust_score_result.context_id
        
        adjustments = []
        
        # Obter mapeamento para a direção específica
        adjustment_map = policy.adjustment_map.get(direction.value, {})
        if not adjustment_map:
            logger.warning(f"Mapa de ajuste não encontrado para direção {direction.value} na política {policy.id}")
            return []
        
        # Para cada mecanismo no mapa
        for mechanism_str, new_level_str in adjustment_map.items():
            try:
                mechanism = SecurityMechanism(mechanism_str)
                new_level = SecurityLevel(new_level_str)
                
                # Obter nível atual
                current_level = await self.get_current_security_level(
                    user_id, tenant_id, mechanism, context_id
                )
                
                # Se o nível atual for diferente do novo nível, criar ajuste
                if current_level != new_level:
                    # Criar parâmetros específicos para o mecanismo
                    params = await self._create_mechanism_params(
                        mechanism, new_level, trust_score_result
                    )
                    
                    # Determinar período de expiração, se aplicável
                    expires_at = None
                    if direction == ScalingDirection.UP:
                        # Se estamos aumentando a segurança, configurar expiração
                        # apenas se permitido o downgrade automático
                        if self.config.allow_auto_downgrade:
                            expires_at = datetime.now() + timedelta(
                                minutes=self.config.downgrade_delay_minutes
                            )
                    
                    # Criar objeto de ajuste
                    adjustment = SecurityAdjustment(
                        mechanism=mechanism,
                        current_level=current_level,
                        new_level=new_level,
                        parameters=params,
                        reason=f"TrustScore: {trust_score_result.overall_score:.2f}",
                        expires_at=expires_at
                    )
                    
                    adjustments.append(adjustment)
            except Exception as e:
                logger.error(f"Erro ao determinar ajuste para mecanismo {mechanism_str}: {e}")
        
        return adjustments
        
    async def _create_mechanism_params(
        self,
        mechanism: SecurityMechanism,
        level: SecurityLevel,
        trust_score_result: TrustScoreResult
    ) -> Dict[str, Any]:
        """
        Cria parâmetros específicos para um mecanismo com base no nível.
        
        Args:
            mechanism: Mecanismo de segurança
            level: Nível de segurança
            trust_score_result: Resultado da avaliação de confiança
            
        Returns:
            Dict[str, Any]: Parâmetros específicos do mecanismo
        """
        params = {}
        
        # Definir parâmetros específicos por mecanismo e nível
        if mechanism == SecurityMechanism.AUTH_FACTORS:
            # Configura fatores de autenticação
            if level == SecurityLevel.MINIMAL:
                params["required_factors"] = 1
                params["allowed_factors"] = ["password"]
            elif level == SecurityLevel.LOW:
                params["required_factors"] = 1
                params["allowed_factors"] = ["password", "otp", "pin"]
            elif level == SecurityLevel.STANDARD:
                params["required_factors"] = 2
                params["allowed_factors"] = ["password", "otp", "pin", "device"]
            elif level == SecurityLevel.HIGH:
                params["required_factors"] = 2
                params["allowed_factors"] = ["password", "otp", "fingerprint", "face", "device"]
                params["biometric_preferred"] = True
            elif level == SecurityLevel.VERY_HIGH:
                params["required_factors"] = 3
                params["allowed_factors"] = ["password", "otp", "fingerprint", "face", "device", "card"]
                params["biometric_required"] = True
            elif level == SecurityLevel.MAXIMUM:
                params["required_factors"] = 3
                params["allowed_factors"] = ["password", "otp", "fingerprint", "face", "device", "card"]
                params["biometric_required"] = True
                params["location_verification"] = True
        
        elif mechanism == SecurityMechanism.SESSION_TIMEOUT:
            # Configura timeout de sessão
            timeouts = {
                SecurityLevel.MINIMAL: 240,  # 4 horas
                SecurityLevel.LOW: 120,      # 2 horas
                SecurityLevel.STANDARD: 60,  # 1 hora
                SecurityLevel.HIGH: 30,      # 30 minutos
                SecurityLevel.VERY_HIGH: 15, # 15 minutos
                SecurityLevel.MAXIMUM: 5     # 5 minutos
            }
            params["timeout_minutes"] = timeouts.get(level, 60)
            
        elif mechanism == SecurityMechanism.TRANSACTION_LIMITS:
            # Configura limites de transação
            score = trust_score_result.overall_score
            base_multiplier = {
                SecurityLevel.MINIMAL: 10.0,
                SecurityLevel.LOW: 5.0,
                SecurityLevel.STANDARD: 2.0,
                SecurityLevel.HIGH: 1.0,
                SecurityLevel.VERY_HIGH: 0.5,
                SecurityLevel.MAXIMUM: 0.25
            }.get(level, 1.0)
            
            # Ajustar com base na pontuação
            score_multiplier = max(0.1, min(2.0, score * 2))
            
            params["daily_limit_multiplier"] = base_multiplier * score_multiplier
            params["transaction_limit_multiplier"] = base_multiplier * score_multiplier
            params["approval_threshold"] = {
                SecurityLevel.MINIMAL: 1000,
                SecurityLevel.LOW: 500,
                SecurityLevel.STANDARD: 200,
                SecurityLevel.HIGH: 100,
                SecurityLevel.VERY_HIGH: 50,
                SecurityLevel.MAXIMUM: 10
            }.get(level, 200)
            
        # Adicionar outros mecanismos conforme necessário
        
        # Adicionar metadados gerais
        params["trust_score"] = trust_score_result.overall_score
        params["applied_at"] = datetime.now().isoformat()
        
        return params
        
    async def _create_scaling_event(
        self,
        trust_score_result: TrustScoreResult,
        policy_id: str,
        trigger_id: str,
        direction: ScalingDirection,
        adjustments: List[SecurityAdjustment]
    ) -> ScalingEvent:
        """
        Cria e registra um evento de escalonamento.
        
        Args:
            trust_score_result: Resultado da avaliação de confiança
            policy_id: ID da política aplicada
            trigger_id: ID do gatilho acionado
            direction: Direção do escalonamento
            adjustments: Lista de ajustes de segurança
            
        Returns:
            ScalingEvent: Evento de escalonamento criado
        """
        # Calcular tempo de expiração (o maior dos ajustes)
        expires_at = None
        for adj in adjustments:
            if adj.expires_at:
                if expires_at is None or adj.expires_at > expires_at:
                    expires_at = adj.expires_at
        
        # Criar evento
        event = ScalingEvent(
            id=str(uuid.uuid4()),
            user_id=trust_score_result.user_id,
            tenant_id=trust_score_result.tenant_id,
            context_id=trust_score_result.context_id,
            region_code=trust_score_result.regional_context,
            trigger_id=trigger_id,
            policy_id=policy_id,
            trust_score=trust_score_result.overall_score,
            dimension_scores=trust_score_result.dimension_scores,
            scaling_direction=direction,
            adjustments=adjustments,
            event_time=datetime.now(),
            expires_at=expires_at,
            metadata={
                "anomalies": [a.dict() for a in trust_score_result.anomalies] if trust_score_result.anomalies else []
            }
        )
        
        # Registrar evento no banco de dados
        try:
            async with self.db_pool.acquire() as conn:
                await conn.execute("""
                    INSERT INTO scaling_events 
                    (id, user_id, tenant_id, context_id, region_code, trigger_id, 
                     policy_id, trust_score, dimension_scores, scaling_direction, 
                     adjustments, event_time, expires_at, metadata)
                    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
                """, event.id, event.user_id, event.tenant_id, event.context_id,
                    event.region_code, event.trigger_id, event.policy_id, event.trust_score,
                    json.dumps(event.dimension_scores), event.scaling_direction.value,
                    json.dumps([adj.dict() for adj in event.adjustments]),
                    event.event_time, event.expires_at, json.dumps(event.metadata))
                
                logger.info(f"Evento de escalonamento {event.id} registrado")
        except Exception as e:
            logger.error(f"Erro ao registrar evento de escalonamento: {e}")
            
        return event
        
    async def _apply_security_adjustments(self, event: ScalingEvent) -> None:
        """
        Aplica ajustes de segurança ao usuário.
        
        Args:
            event: Evento de escalonamento com ajustes a serem aplicados
        """
        for adjustment in event.adjustments:
            try:
                async with self.db_pool.acquire() as conn:
                    # Verificar se já existe um registro para este usuário/mecanismo
                    existing = await conn.fetchrow("""
                        SELECT id FROM security_levels
                        WHERE user_id = $1 AND tenant_id = $2 
                        AND context_id = $3 AND mechanism = $4
                    """, event.user_id, event.tenant_id, event.context_id, adjustment.mechanism.value)
                    
                    if existing:
                        # Atualizar registro existente
                        await conn.execute("""
                            UPDATE security_levels
                            SET level = $1, parameters = $2, expires_at = $3,
                                updated_at = $4, updated_by = $5, metadata = $6
                            WHERE user_id = $7 AND tenant_id = $8 
                            AND context_id = $9 AND mechanism = $10
                        """, adjustment.new_level.value, json.dumps(adjustment.parameters),
                            adjustment.expires_at, datetime.now(), "adaptive_scaling",
                            json.dumps({"reason": adjustment.reason, "event_id": event.id}),
                            event.user_id, event.tenant_id, event.context_id, adjustment.mechanism.value)
                    else:
                        # Inserir novo registro
                        await conn.execute("""
                            INSERT INTO security_levels
                            (user_id, tenant_id, context_id, mechanism, level, 
                             parameters, expires_at, created_at, updated_at, created_by, updated_by, metadata)
                            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
                        """, event.user_id, event.tenant_id, event.context_id, adjustment.mechanism.value,
                            adjustment.new_level.value, json.dumps(adjustment.parameters),
                            adjustment.expires_at, datetime.now(), datetime.now(),
                            "adaptive_scaling", "adaptive_scaling",
                            json.dumps({"reason": adjustment.reason, "event_id": event.id}))
                    
                    logger.info(f"Ajuste aplicado: {adjustment.mechanism.value} -> {adjustment.new_level.value} para usuário {event.user_id}")
            except Exception as e:
                logger.error(f"Erro ao aplicar ajuste de segurança: {e}")
                
    async def _notify_user_of_changes(self, event: ScalingEvent) -> None:
        """
        Notifica o usuário sobre mudanças de segurança.
        
        Args:
            event: Evento de escalonamento com ajustes aplicados
        """
        if not self.notification_service:
            logger.warning("Serviço de notificação não disponível para enviar notificação")
            return
            
        try:
            # Criar mensagem de notificação
            direction_text = {
                ScalingDirection.UP: "aumentados",
                ScalingDirection.DOWN: "reduzidos",
                ScalingDirection.MAINTAIN: "mantidos"
            }.get(event.scaling_direction, "alterados")
            
            title = f"Níveis de segurança {direction_text}"
            
            # Criar descrição das mudanças
            changes = []
            for adj in event.adjustments:
                mechanism_name = {
                    SecurityMechanism.AUTH_FACTORS: "Fatores de autenticação",
                    SecurityMechanism.SESSION_TIMEOUT: "Timeout de sessão",
                    SecurityMechanism.TRANSACTION_LIMITS: "Limites de transação",
                    SecurityMechanism.DEVICE_VERIFICATION: "Verificação de dispositivo",
                    SecurityMechanism.LOCATION_VERIFICATION: "Verificação de localização",
                    SecurityMechanism.BIOMETRIC_REQUIREMENT: "Requisitos biométricos",
                    SecurityMechanism.BEHAVIORAL_ANALYSIS: "Análise comportamental",
                    SecurityMechanism.CONTEXTUAL_AWARENESS: "Consciência contextual",
                    SecurityMechanism.CREDENTIAL_COMPLEXITY: "Complexidade de credenciais",
                    SecurityMechanism.PRIVILEGED_ACCESS: "Acesso privilegiado"
                }.get(adj.mechanism, adj.mechanism.value)
                
                level_name = {
                    SecurityLevel.MINIMAL: "Mínimo",
                    SecurityLevel.LOW: "Baixo",
                    SecurityLevel.STANDARD: "Padrão",
                    SecurityLevel.HIGH: "Alto",
                    SecurityLevel.VERY_HIGH: "Muito alto",
                    SecurityLevel.MAXIMUM: "Máximo"
                }.get(adj.new_level, adj.new_level.value)
                
                changes.append(f"{mechanism_name}: {level_name}")
            
            # Criar corpo da mensagem
            body = f"Seus níveis de segurança foram {direction_text} com base na sua pontuação de confiança atual.\n\n"
            body += "Alterações:\n" + "\n".join(f"- {change}" for change in changes)
            
            if event.expires_at:
                expiry = event.expires_at.strftime("%d/%m/%Y às %H:%M")
                body += f"\n\nEstas alterações expiram em {expiry}."
                
            body += "\n\nSe você tiver dúvidas ou precisar de ajuda, entre em contato com o suporte."
            
            # Enviar notificação
            await self.notification_service.send_user_notification(
                user_id=event.user_id,
                tenant_id=event.tenant_id,
                title=title,
                body=body,
                category="security",
                priority="medium" if event.scaling_direction == ScalingDirection.UP else "low",
                metadata={
                    "event_id": event.id,
                    "trust_score": event.trust_score,
                    "scaling_direction": event.scaling_direction.value
                }
            )
            
            logger.info(f"Notificação enviada para usuário {event.user_id}")
        except Exception as e:
            logger.error(f"Erro ao notificar usuário sobre mudanças: {e}")
            
    async def _get_default_security_level(
        self,
        tenant_id: str,
        mechanism: SecurityMechanism,
        context_id: Optional[str] = None
    ) -> SecurityLevel:
        """
        Obtém o nível de segurança padrão para um mecanismo.
        
        Args:
            tenant_id: ID do tenant
            mechanism: Mecanismo de segurança
            context_id: ID do contexto (opcional)
            
        Returns:
            SecurityLevel: Nível de segurança padrão
        """
        try:
            async with self.db_pool.acquire() as conn:
                # Buscar padrão do contexto específico
                if context_id:
                    row = await conn.fetchrow("""
                        SELECT level FROM security_defaults
                        WHERE tenant_id = $1 AND mechanism = $2 AND context_id = $3
                    """, tenant_id, mechanism.value, context_id)
                    
                    if row:
                        return SecurityLevel(row['level'])
                
                # Buscar padrão do tenant
                row = await conn.fetchrow("""
                    SELECT level FROM security_defaults
                    WHERE tenant_id = $1 AND mechanism = $2 AND context_id IS NULL
                """, tenant_id, mechanism.value)
                
                if row:
                    return SecurityLevel(row['level'])
                
                # Buscar padrão global
                row = await conn.fetchrow("""
                    SELECT level FROM security_defaults
                    WHERE tenant_id IS NULL AND mechanism = $1 AND context_id IS NULL
                """, mechanism.value)
                
                if row:
                    return SecurityLevel(row['level'])
            
            # Padrão fallback
            return SecurityLevel.STANDARD
        except Exception as e:
            logger.error(f"Erro ao obter nível de segurança padrão: {e}")
            return SecurityLevel.STANDARD