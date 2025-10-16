#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Extensão do Consumidor de Análise Comportamental com integração UniConnect

Este módulo estende o consumidor de análise comportamental para incluir
capacidades avançadas de notificação utilizando o módulo UniConnect,
permitindo alertas omnicanal sobre anomalias comportamentais detectadas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import logging
import json
from typing import Dict, Any, List, Optional, Tuple
from datetime import datetime

# Importar componentes base
from .behavioral_analysis_consumer import BehavioralAnalysisConsumer, ProcessingResult
from .integrations.uniconnect_notifier import (
    UniConnectNotifier, 
    NotificationChannel, 
    NotificationPriority,
    NotificationRecipient,
    create_uniconnect_notifier
)

# Configuração do logger
logger = logging.getLogger("iam.trustguard.behavioral.notifications")


class NotificationEnabledBehavioralConsumer(BehavioralAnalysisConsumer):
    """
    Versão estendida do consumidor de análise comportamental com capacidades
    avançadas de notificação via UniConnect para alertas omnicanal.
    
    Esta classe adiciona funcionalidades para:
    1. Enviar notificações em diferentes canais (email, SMS, push, etc.)
    2. Aplicar políticas de notificação baseadas em contexto e severidade
    3. Implementar estratégias de escalação para alertas críticos
    4. Integrar com fluxos de trabalho de gestão de fraudes
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o consumidor com notificações habilitadas.
        
        Args:
            config: Configurações do consumidor, incluindo:
                - Todas as configurações base do BehavioralAnalysisConsumer
                - uniconnect_config: Configurações específicas do UniConnect
                - notification_policies: Políticas para envio de notificações
                - escalation_matrix: Matriz de escalação para alertas
        """
        # Inicializar classe base
        super().__init__(config)
        
        # Configurações de notificação
        uniconnect_config = config.get("uniconnect_config", {})
        self.notification_policies = config.get("notification_policies", {})
        self.escalation_matrix = config.get("escalation_matrix", {})
        
        # Flag para habilitar/desabilitar notificações
        self.notifications_enabled = config.get("notifications_enabled", True)
        
        # Inicializar notificador UniConnect
        if self.notifications_enabled:
            self.notifier = create_uniconnect_notifier(uniconnect_config)
            logger.info("Notificações habilitadas via UniConnect")
        else:
            self.notifier = None
            logger.info("Notificações desabilitadas")
            
        # Configurar limites para notificação
        self.notification_thresholds = config.get("notification_thresholds", {
            "user": {
                "low": 0.65,
                "medium": 0.75,
                "high": 0.85,
                "critical": 0.95
            },
            "security": {
                "low": 0.60,
                "medium": 0.70,
                "high": 0.80,
                "critical": 0.90
            }
        })
        
        # Cache para evitar notificações duplicadas em curto período
        self.notification_cache = {}
        self.notification_cooldown = config.get("notification_cooldown", 600)  # 10 minutos em segundos
    
    def process_event(self, topic: str, event: Dict[str, Any]) -> ProcessingResult:
        """
        Processa um evento comportamental e envia notificações quando necessário.
        
        Sobrescreve o método da classe base para adicionar a lógica de notificação.
        
        Args:
            topic: Tópico do evento
            event: Dados do evento a ser processado
            
        Returns:
            Resultado do processamento com informações adicionais de notificação
        """
        # Processar evento usando a implementação base
        result = super().process_event(topic, event)
        
        # Se anomalias foram detectadas e notificações estão habilitadas, enviar alertas
        if result.anomalies_detected and self.notifications_enabled and self.notifier:
            notification_results = self._handle_anomaly_notifications(
                normalized_event=result.normalized_event,
                anomalies=result.anomalies,
                anomaly_score=result.anomaly_score,
                user_profile=result.user_profile
            )
            
            # Adicionar informações de notificação ao resultado
            result.notification_sent = notification_results.get("success", False)
            result.notification_ids = notification_results.get("notification_ids", [])
            
        return result
    
    def _handle_anomaly_notifications(self, 
                                    normalized_event: Dict[str, Any],
                                    anomalies: List[Dict[str, Any]],
                                    anomaly_score: float,
                                    user_profile: Dict[str, Any]) -> Dict[str, Any]:
        """
        Gerencia o envio de notificações baseadas em anomalias detectadas.
        
        Args:
            normalized_event: Evento normalizado
            anomalies: Lista de anomalias detectadas
            anomaly_score: Pontuação de anomalia calculada
            user_profile: Perfil do usuário
            
        Returns:
            Resultado das operações de notificação
        """
        # Verificar se já enviamos alertas recentemente para este usuário
        user_id = normalized_event.get("user_id")
        if user_id and self._is_notification_in_cooldown(user_id):
            logger.debug(f"Notificação em período de cooldown para usuário {user_id}")
            return {"success": False, "reason": "COOLDOWN"}
        
        # Determinar o tipo e severidade do alerta
        notification_type, priority = self._determine_notification_details(
            normalized_event, anomalies, anomaly_score
        )
        
        # Se não for necessário notificar, retornar
        if not notification_type:
            return {"success": False, "reason": "BELOW_THRESHOLD"}
        
        # Construir dados do alerta
        alert_data = self._build_alert_data(
            notification_type=notification_type,
            normalized_event=normalized_event,
            anomalies=anomalies,
            anomaly_score=anomaly_score,
            user_profile=user_profile
        )
        
        # Determinar destinatários com base nas políticas e matriz de escalação
        recipients = self._get_notification_recipients(
            notification_type=notification_type,
            priority=priority,
            user_id=user_id,
            context=normalized_event.get("context", {})
        )
        
        if not recipients:
            logger.warning(f"Nenhum destinatário encontrado para alerta {notification_type}")
            return {"success": False, "reason": "NO_RECIPIENTS"}
        
        # Registrar notificação no cache para controle de cooldown
        self._register_notification(user_id)
        
        # Obter código da região
        region_code = self._get_region_code(normalized_event, user_profile)
        
        # Enviar notificação via UniConnect
        notification_result = self.notifier.send_behavioral_alert(
            alert_data=alert_data,
            recipients=recipients,
            priority=priority,
            region_code=region_code
        )
        
        # Registrar envio no log
        if notification_result.get("success"):
            logger.info(f"Alerta comportamental enviado com sucesso: {notification_result['notification_ids']}")
        else:
            logger.error(f"Falha ao enviar alerta comportamental: {notification_result}")
        
        return notification_result
    
    def _determine_notification_details(self,
                                     normalized_event: Dict[str, Any],
                                     anomalies: List[Dict[str, Any]],
                                     anomaly_score: float) -> Tuple[Optional[str], NotificationPriority]:
        """
        Determina o tipo de notificação e prioridade com base nas anomalias.
        
        Args:
            normalized_event: Evento normalizado
            anomalies: Lista de anomalias detectadas
            anomaly_score: Pontuação de anomalia
            
        Returns:
            Tupla com tipo de notificação e prioridade, ou (None, None) se não precisar notificar
        """
        # Verificar se o score atinge algum limite
        thresholds = self.notification_thresholds.get("user", {})
        
        # Determinar tipo de notificação com base no contexto e anomalias
        notification_type = None
        priority = NotificationPriority.MEDIUM
        
        # Verificar anomalias críticas de segurança
        has_critical_security_anomaly = any(
            a.get("category") == "security" and a.get("severity", 0) >= 8
            for a in anomalies
        )
        
        # Verificar anomalias de localização
        has_location_anomaly = any(
            a.get("type") == "location_change" or a.get("type") == "impossible_travel"
            for a in anomalies
        )
        
        # Verificar anomalias de dispositivo
        has_device_anomaly = any(
            a.get("type") == "new_device" or a.get("type") == "device_change"
            for a in anomalies
        )
        
        # Verificar anomalias de transação
        has_transaction_anomaly = any(
            a.get("category") == "transaction"
            for a in anomalies
        )
        
        # Determinar tipo e prioridade com base em condições específicas
        if has_critical_security_anomaly:
            notification_type = "security_threat"
            priority = NotificationPriority.CRITICAL
            
        elif has_location_anomaly and anomaly_score >= thresholds.get("high", 0.85):
            notification_type = "location_anomaly"
            priority = NotificationPriority.HIGH
            
        elif has_transaction_anomaly and anomaly_score >= thresholds.get("medium", 0.75):
            notification_type = "transaction_anomaly"
            priority = NotificationPriority.HIGH if anomaly_score >= thresholds.get("high", 0.85) else NotificationPriority.MEDIUM
            
        elif has_device_anomaly and anomaly_score >= thresholds.get("medium", 0.75):
            notification_type = "device_anomaly"
            priority = NotificationPriority.MEDIUM
            
        elif anomaly_score >= thresholds.get("low", 0.65):
            notification_type = "behavioral_anomaly"
            priority = NotificationPriority.LOW
        
        return notification_type, priority
    
    def _build_alert_data(self,
                        notification_type: str,
                        normalized_event: Dict[str, Any],
                        anomalies: List[Dict[str, Any]],
                        anomaly_score: float,
                        user_profile: Dict[str, Any]) -> Dict[str, Any]:
        """
        Constrói os dados detalhados para o alerta comportamental.
        
        Args:
            notification_type: Tipo de notificação
            normalized_event: Evento normalizado
            anomalies: Lista de anomalias detectadas
            anomaly_score: Pontuação de anomalia
            user_profile: Perfil do usuário
            
        Returns:
            Dados formatados do alerta para notificação
        """
        # Gerar ID único para o alerta
        alert_id = f"beh-{normalized_event.get('event_id', 'unknown')}-{int(datetime.now().timestamp())}"
        
        # Extrair informações do dispositivo e localização
        device_info = normalized_event.get("device", {})
        location_info = normalized_event.get("location", {})
        
        # Formatar anomalias para exibição
        formatted_anomalies = []
        for anomaly in anomalies:
            formatted_anomalies.append({
                "type": anomaly.get("type", "unknown"),
                "description": anomaly.get("description", ""),
                "severity": anomaly.get("severity", 0),
                "details": anomaly.get("details", {})
            })
        
        # Construir dados do alerta
        alert_data = {
            "alert_id": alert_id,
            "user_id": normalized_event.get("user_id"),
            "event_type": normalized_event.get("event_type"),
            "notification_type": notification_type,
            "anomaly_score": anomaly_score,
            "timestamp": normalized_event.get("timestamp", datetime.now().isoformat()),
            "device": {
                "type": device_info.get("type", "unknown"),
                "os": device_info.get("os", "unknown"),
                "browser": device_info.get("browser", "unknown"),
                "fingerprint": device_info.get("fingerprint", "unknown"),
                "is_new": device_info.get("is_new", False)
            },
            "location": {
                "country": location_info.get("country", "unknown"),
                "city": location_info.get("city", "unknown"),
                "ip": location_info.get("ip", "unknown"),
                "coordinates": location_info.get("coordinates", {})
            },
            "anomalies": formatted_anomalies,
            "user_details": {
                "full_name": user_profile.get("full_name", ""),
                "account_type": user_profile.get("account_type", ""),
                "account_age_days": user_profile.get("account_age_days", 0),
                "risk_level": user_profile.get("risk_level", "medium")
            },
            "recommended_actions": self._get_recommended_actions(notification_type, anomalies),
            "additional_context": normalized_event.get("context", {})
        }
        
        return alert_data
    
    def _get_notification_recipients(self,
                                  notification_type: str,
                                  priority: NotificationPriority,
                                  user_id: str,
                                  context: Dict[str, Any]) -> List[NotificationRecipient]:
        """
        Determina os destinatários para uma notificação com base nas políticas e matriz de escalação.
        
        Args:
            notification_type: Tipo de notificação
            priority: Prioridade da notificação
            user_id: ID do usuário relacionado ao alerta
            context: Contexto adicional do evento
            
        Returns:
            Lista de destinatários para a notificação
        """
        recipients = []
        
        # Sempre notificar o próprio usuário para eventos de seu perfil,
        # exceto em casos de suspeita de comprometimento da conta
        user_compromised = any(a.get("type") == "account_compromise" 
                               for a in context.get("anomalies", []))
        
        if not user_compromised:
            recipients.append(NotificationRecipient(
                user_id=user_id,
                channels=[NotificationChannel.PUSH, NotificationChannel.EMAIL],
                region_code=context.get("region_code", "AO")
            ))
        
        # Para prioridades elevadas, adicionar equipe de segurança baseado na matriz de escalação
        if priority in [NotificationPriority.HIGH, NotificationPriority.CRITICAL, NotificationPriority.EMERGENCY]:
            # Obter equipe de segurança baseado na região
            region_code = context.get("region_code", "AO")
            escalation_level = 1 if priority == NotificationPriority.HIGH else 2
            
            # Obter responsáveis da matriz de escalação
            security_team = self._get_security_team_from_matrix(
                region_code=region_code,
                escalation_level=escalation_level,
                notification_type=notification_type
            )
            
            # Adicionar cada membro da equipe como destinatário
            for member in security_team:
                channels = [NotificationChannel.EMAIL]
                
                # Para emergências, adicionar mais canais
                if priority in [NotificationPriority.CRITICAL, NotificationPriority.EMERGENCY]:
                    channels.extend([NotificationChannel.SMS, NotificationChannel.PUSH])
                
                recipients.append(NotificationRecipient(
                    user_id=member["user_id"],
                    channels=channels,
                    region_code=region_code,
                    role=member.get("role"),
                    department=member.get("department", "security"),
                    escalation_level=escalation_level
                ))
        
        return recipients
    
    def _get_security_team_from_matrix(self,
                                    region_code: str,
                                    escalation_level: int,
                                    notification_type: str) -> List[Dict[str, Any]]:
        """
        Obtém a equipe de segurança baseado na matriz de escalação.
        
        Args:
            region_code: Código da região
            escalation_level: Nível de escalação
            notification_type: Tipo de notificação
            
        Returns:
            Lista de membros da equipe de segurança para notificar
        """
        # Por padrão, usar a matriz da região ou global se não existir específica
        matrix = self.escalation_matrix.get(region_code, self.escalation_matrix.get("global", {}))
        
        # Obter equipe para o nível de escalação e tipo de notificação
        team = matrix.get(f"level_{escalation_level}", {}).get(notification_type, [])
        
        # Se não houver equipe específica para o tipo, usar equipe padrão
        if not team:
            team = matrix.get(f"level_{escalation_level}", {}).get("default", [])
        
        return team
    
    def _get_recommended_actions(self,
                               notification_type: str,
                               anomalies: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """
        Gera recomendações de ações baseadas no tipo de notificação e anomalias.
        
        Args:
            notification_type: Tipo de notificação
            anomalies: Lista de anomalias detectadas
            
        Returns:
            Lista de ações recomendadas
        """
        actions = []
        
        # Recomendações baseadas no tipo de notificação
        if notification_type == "security_threat":
            actions.append({
                "action": "block_account",
                "description": "Bloquear conta temporariamente",
                "urgency": "immediate"
            })
            actions.append({
                "action": "contact_user",
                "description": "Contactar usuário por canal alternativo",
                "urgency": "high"
            })
            
        elif notification_type == "location_anomaly":
            actions.append({
                "action": "verify_identity",
                "description": "Solicitar verificação adicional de identidade",
                "urgency": "high"
            })
            
        elif notification_type == "transaction_anomaly":
            actions.append({
                "action": "review_transaction",
                "description": "Revisar detalhes da transação",
                "urgency": "medium"
            })
            
        elif notification_type == "device_anomaly":
            actions.append({
                "action": "verify_device",
                "description": "Confirmar novo dispositivo",
                "urgency": "medium"
            })
        
        # Recomendações baseadas em tipos específicos de anomalias
        for anomaly in anomalies:
            anomaly_type = anomaly.get("type")
            
            if anomaly_type == "account_compromise":
                if not any(a["action"] == "reset_password" for a in actions):
                    actions.append({
                        "action": "reset_password",
                        "description": "Solicitar redefinição de senha",
                        "urgency": "high"
                    })
                    
            elif anomaly_type == "impossible_travel":
                if not any(a["action"] == "verify_location" for a in actions):
                    actions.append({
                        "action": "verify_location",
                        "description": "Verificar localização atual do usuário",
                        "urgency": "high"
                    })
        
        return actions
    
    def _is_notification_in_cooldown(self, user_id: str) -> bool:
        """
        Verifica se o usuário está em período de cooldown para evitar spam de notificações.
        
        Args:
            user_id: ID do usuário
            
        Returns:
            True se estiver em cooldown, False caso contrário
        """
        if user_id not in self.notification_cache:
            return False
            
        last_notification_time = self.notification_cache[user_id]
        current_time = datetime.now().timestamp()
        
        return (current_time - last_notification_time) < self.notification_cooldown
    
    def _register_notification(self, user_id: str) -> None:
        """
        Registra timestamp de notificação para controle de cooldown.
        
        Args:
            user_id: ID do usuário
        """
        if not user_id:
            return
            
        self.notification_cache[user_id] = datetime.now().timestamp()
    
    def _get_region_code(self, 
                       normalized_event: Dict[str, Any], 
                       user_profile: Dict[str, Any]) -> str:
        """
        Determina o código da região para a notificação.
        
        Args:
            normalized_event: Evento normalizado
            user_profile: Perfil do usuário
            
        Returns:
            Código da região
        """
        # Tentar obter do evento
        region_code = normalized_event.get("context", {}).get("region_code")
        if region_code:
            return region_code
            
        # Tentar obter do perfil do usuário
        region_code = user_profile.get("region_code")
        if region_code:
            return region_code
            
        # Tentar inferir da localização
        location = normalized_event.get("location", {})
        country = location.get("country")
        
        if country:
            # Mapeamento básico de países para códigos de região
            country_to_region = {
                "Angola": "AO",
                "Brasil": "BR",
                "Brazil": "BR",
                "Portugal": "PT",
                "Moçambique": "MZ",
                "Mozambique": "MZ"
            }
            
            return country_to_region.get(country, "AO")
        
        # Valor padrão
        return "AO"