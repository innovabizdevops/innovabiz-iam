#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Módulo de Integração com UniConnect para Notificações de Anomalias Comportamentais

Este módulo implementa a integração com o sistema UniConnect para envio de
notificações omnicanal (email, SMS, push, WhatsApp, in-app, etc.) relacionadas
a anomalias comportamentais e alertas de fraude.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import hmac
import json
import time
import base64
import hashlib
import logging
import datetime
from enum import Enum
from typing import Dict, Any, List, Optional, Union
from dataclasses import dataclass

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

# Configuração do logger
logger = logging.getLogger("iam.trustguard.notifications.uniconnect")


class NotificationType(Enum):
    """Tipos de notificações relacionadas a análise comportamental."""
    BEHAVIORAL_ALERT = "behavioral_alert"
    FRAUD_DETECTED = "fraud_detected"
    SUSPICIOUS_ACTIVITY = "suspicious_activity"
    ACCOUNT_LOCKDOWN = "account_lockdown"
    VERIFICATION_REQUIRED = "verification_required"
    LOCATION_CHANGE = "location_change"
    DEVICE_CHANGE = "device_change"
    REMEDIATION_INSTRUCTIONS = "remediation_instructions"
    FRAUD_REPORT = "fraud_report"


class NotificationPriority(Enum):
    """Níveis de prioridade para notificações."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class NotificationChannel(Enum):
    """Canais de entrega disponíveis para notificações."""
    EMAIL = "email"
    SMS = "sms"
    PUSH = "push"
    WHATSAPP = "whatsapp"
    IN_APP = "in_app"
    VOICE = "voice"
    TELEGRAM = "telegram"
    MOBILE_BANKING = "mobile_banking"
    WEBSOCKET = "websocket"


@dataclass
class NotificationRecipient:
    """
    Representa um destinatário de notificação com canais preferenciais.
    """
    recipient_id: str
    channels: List[NotificationChannel]
    name: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    language: str = "pt-PT"
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o objeto para dicionário para uso na API."""
        return {
            "recipient_id": self.recipient_id,
            "name": self.name,
            "email": self.email,
            "phone": self.phone,
            "language": self.language,
            "channels": [c.value for c in self.channels]
        }


class UniConnectNotifier:
    """
    Implementa a integração com o sistema UniConnect para envio de notificações
    omnicanal relacionadas a anomalias comportamentais e alertas de fraude.
    
    Esta classe encapsula a lógica de autenticação, geração de payload, envio de
    notificações e tratamento de respostas/erros para a API do UniConnect.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o notificador UniConnect.
        
        Args:
            config: Configurações para conexão com UniConnect, incluindo:
                - api_url: URL base da API
                - api_key: Chave de API
                - api_secret: Segredo de API para assinatura HMAC
                - default_timeout: Timeout padrão para requisições
                - retry_attempts: Número de tentativas de retry
                - notification_policy: Política de envio de notificações
        """
        self.api_url = config.get("api_url", "https://api.uniconnect.innovabiz.com/v1")
        self.api_key = config.get("api_key", "")
        self.api_secret = config.get("api_secret", "")
        self.timeout = config.get("default_timeout", 10)
        self.retry_attempts = config.get("retry_attempts", 3)
        self.notification_policy = config.get("notification_policy", {})
        self.notification_templates = config.get("notification_templates", {})
        
        # Validar configurações essenciais
        if not self.api_key or not self.api_secret:
            logger.error("Configuração incompleta: api_key e api_secret são obrigatórios")
            raise ValueError("Configuração incompleta: api_key e api_secret são obrigatórios")
        
        # Configurar sessão HTTP com retry automático
        self.session = requests.Session()
        retry_strategy = Retry(
            total=self.retry_attempts,
            status_forcelist=[429, 500, 502, 503, 504],
            allowed_methods=["HEAD", "GET", "POST"],
            backoff_factor=1  # Backoff exponencial
        )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("https://", adapter)
        self.session.mount("http://", adapter)
        
        logger.info(f"Notificador UniConnect inicializado para {self.api_url}")
    
    def _generate_signature(self, payload: Dict[str, Any]) -> str:
        """
        Gera assinatura HMAC para autenticação com o UniConnect.
        
        Args:
            payload: Conteúdo a ser assinado
            
        Returns:
            Assinatura codificada em base64
        """
        timestamp = str(int(time.time()))
        payload_str = json.dumps(payload, separators=(',', ':'))
        message = f"{timestamp}.{payload_str}"
        
        signature = hmac.new(
            self.api_secret.encode('utf-8'),
            message.encode('utf-8'),
            hashlib.sha256
        ).digest()
        
        return {
            "signature": base64.b64encode(signature).decode('utf-8'),
            "timestamp": timestamp
        }
    
    def _get_headers(self, payload: Dict[str, Any]) -> Dict[str, str]:
        """
        Prepara os headers da requisição, incluindo autenticação.
        
        Args:
            payload: Payload da requisição para assinatura
            
        Returns:
            Headers HTTP para a requisição
        """
        auth_data = self._generate_signature(payload)
        
        return {
            "Content-Type": "application/json",
            "X-API-Key": self.api_key,
            "X-Timestamp": auth_data["timestamp"],
            "X-Signature": auth_data["signature"],
            "X-Client-ID": "innovabiz-iam-trustguard",
            "User-Agent": "IAM-TrustGuard/1.0"
        }
    
    def send_notification(self, 
                       notification_type: NotificationType,
                       recipients: List[NotificationRecipient],
                       subject: str,
                       message: str,
                       priority: NotificationPriority = NotificationPriority.MEDIUM,
                       metadata: Optional[Dict[str, Any]] = None,
                       template_data: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Envia uma notificação via UniConnect para os destinatários especificados.
        
        Args:
            notification_type: Tipo da notificação
            recipients: Lista de destinatários
            subject: Assunto/título da notificação
            message: Corpo da mensagem
            priority: Prioridade da notificação
            metadata: Metadados adicionais para a notificação
            template_data: Dados para preenchimento de template
            
        Returns:
            Resposta da API UniConnect
        """
        payload = {
            "type": notification_type.value,
            "recipients": [r.to_dict() for r in recipients],
            "content": {
                "subject": subject,
                "message": message
            },
            "priority": priority.value,
            "metadata": metadata or {},
            "template_data": template_data or {}
        }
        
        # Adicionar ID do template, se disponível para o tipo de notificação
        if notification_type.value in self.notification_templates:
            template_id = self.notification_templates.get(notification_type.value)
            if template_id:
                payload["template_id"] = template_id
        
        # Adicionar política de escalonamento baseada na prioridade
        escalation_policy = self._get_escalation_policy(priority)
        if escalation_policy:
            payload["escalation_policy"] = escalation_policy
        
        try:
            headers = self._get_headers(payload)
            endpoint = f"{self.api_url}/notifications/send"
            
            response = self.session.post(
                endpoint,
                headers=headers,
                json=payload,
                timeout=self.timeout
            )
            
            if response.status_code == 200 or response.status_code == 201:
                logger.info(
                    f"Notificação {notification_type.value} enviada com sucesso para "
                    f"{len(recipients)} destinatários. ID: {response.json().get('notification_id')}"
                )
                return response.json()
            else:
                logger.error(
                    f"Erro ao enviar notificação: {response.status_code} - {response.text}"
                )
                return {
                    "success": False,
                    "error": {
                        "code": response.status_code,
                        "message": response.text
                    }
                }
        except Exception as e:
            logger.error(f"Exceção ao enviar notificação: {str(e)}")
            return {
                "success": False,
                "error": {
                    "message": str(e),
                    "exception": type(e).__name__
                }
            }
    
    def _get_escalation_policy(self, priority: NotificationPriority) -> Optional[Dict[str, Any]]:
        """
        Obtém política de escalonamento baseada na prioridade da notificação.
        
        Args:
            priority: Prioridade da notificação
            
        Returns:
            Política de escalonamento, se configurada
        """
        escalation_policies = self.notification_policy.get("escalation", {})
        return escalation_policies.get(priority.value)
    
    def send_behavioral_alert(self, 
                           user_id: str,
                           alert_data: Dict[str, Any],
                           recipients: List[NotificationRecipient]) -> Dict[str, Any]:
        """
        Envia um alerta comportamental via UniConnect.
        
        Args:
            user_id: ID do usuário associado ao alerta
            alert_data: Dados do alerta comportamental
            recipients: Lista de destinatários
            
        Returns:
            Resposta da API UniConnect
        """
        # Determinar prioridade com base no score de risco
        risk_score = alert_data.get("risk_score", 0.5)
        
        if risk_score >= 0.8:
            priority = NotificationPriority.CRITICAL
        elif risk_score >= 0.6:
            priority = NotificationPriority.HIGH
        elif risk_score >= 0.4:
            priority = NotificationPriority.MEDIUM
        else:
            priority = NotificationPriority.LOW
        
        # Preparar dados do template
        template_data = {
            "user_id": user_id,
            "risk_score": risk_score,
            "alert_type": alert_data.get("alert_type", "anomalia_comportamental"),
            "timestamp": datetime.datetime.now().isoformat(),
            "details": alert_data.get("details", {}),
            "recommendations": alert_data.get("recommendations", []),
            "device_info": alert_data.get("device_info", {}),
            "location_info": alert_data.get("location_info", {}),
            "action_required": alert_data.get("action_required", False)
        }
        
        # Determinar assunto e mensagem com base no tipo de alerta
        alert_type = alert_data.get("alert_type", "unknown")
        
        subject_templates = {
            "location_anomaly": "Alerta de Segurança: Localização Incomum Detectada",
            "device_anomaly": "Alerta de Segurança: Novo Dispositivo Detectado",
            "transaction_anomaly": "Alerta de Segurança: Transação Suspeita",
            "auth_anomaly": "Alerta de Segurança: Padrão de Autenticação Anômalo",
            "session_anomaly": "Alerta de Segurança: Comportamento de Sessão Incomum",
            "unknown": "Alerta de Segurança: Comportamento Suspeito Detectado"
        }
        
        subject = subject_templates.get(alert_type, subject_templates["unknown"])
        
        # Construir mensagem básica (será enriquecida pelo template no UniConnect)
        message = (
            f"Detectamos comportamento potencialmente suspeito "
            f"com nível de risco {int(risk_score * 100)}%. "
            f"Confira os detalhes no seu painel de segurança."
        )
        
        # Adicionar metadados para processamento e rastreabilidade
        metadata = {
            "user_id": user_id,
            "alert_id": alert_data.get("alert_id", ""),
            "risk_score": risk_score,
            "alert_type": alert_type,
            "source": "iam_trustguard_behavioral_consumer",
            "region": alert_data.get("region", "global")
        }
        
        return self.send_notification(
            notification_type=NotificationType.BEHAVIORAL_ALERT,
            recipients=recipients,
            subject=subject,
            message=message,
            priority=priority,
            metadata=metadata,
            template_data=template_data
        )
    
    def send_fraud_alert_to_security_team(self, 
                                       fraud_data: Dict[str, Any],
                                       security_team_recipients: List[NotificationRecipient]) -> Dict[str, Any]:
        """
        Envia alerta de fraude para a equipe de segurança.
        
        Args:
            fraud_data: Dados da fraude detectada
            security_team_recipients: Lista de destinatários da equipe de segurança
            
        Returns:
            Resposta da API UniConnect
        """
        # Sempre usar prioridade crítica para alertas de fraude para a equipe de segurança
        priority = NotificationPriority.CRITICAL
        
        # Preparar dados do template
        template_data = {
            "fraud_id": fraud_data.get("fraud_id", ""),
            "user_id": fraud_data.get("user_id", ""),
            "risk_score": fraud_data.get("risk_score", 1.0),
            "fraud_type": fraud_data.get("fraud_type", "unknown"),
            "detection_method": fraud_data.get("detection_method", "behavioral_analysis"),
            "timestamp": datetime.datetime.now().isoformat(),
            "details": fraud_data.get("details", {}),
            "evidence": fraud_data.get("evidence", {}),
            "recommended_actions": fraud_data.get("recommended_actions", []),
            "device_info": fraud_data.get("device_info", {}),
            "location_info": fraud_data.get("location_info", {}),
            "transaction_info": fraud_data.get("transaction_info", {})
        }
        
        # Construir assunto com informações críticas para visibilidade
        subject = (
            f"[CRÍTICO] Fraude Detectada - {fraud_data.get('fraud_type', 'Desconhecido')} - "
            f"Usuário: {fraud_data.get('user_id', '')}"
        )
        
        # Construir mensagem com detalhes essenciais
        message = (
            f"Fraude detectada pelo sistema de análise comportamental.\n"
            f"Tipo: {fraud_data.get('fraud_type', 'Desconhecido')}\n"
            f"Usuário: {fraud_data.get('user_id', '')}\n"
            f"Score de risco: {fraud_data.get('risk_score', 1.0)}\n"
            f"Região: {fraud_data.get('region', 'Global')}\n\n"
            f"Ação requerida imediatamente. Acesse o painel de segurança para detalhes."
        )
        
        # Adicionar metadados para processamento e rastreabilidade
        metadata = {
            "fraud_id": fraud_data.get("fraud_id", ""),
            "user_id": fraud_data.get("user_id", ""),
            "risk_score": fraud_data.get("risk_score", 1.0),
            "fraud_type": fraud_data.get("fraud_type", "unknown"),
            "source": "iam_trustguard_behavioral_consumer",
            "region": fraud_data.get("region", "global"),
            "requires_immediate_action": True
        }
        
        return self.send_notification(
            notification_type=NotificationType.FRAUD_DETECTED,
            recipients=security_team_recipients,
            subject=subject,
            message=message,
            priority=priority,
            metadata=metadata,
            template_data=template_data
        )


class UniConnectNotifierFactory:
    """
    Factory para criar instâncias do notificador UniConnect.
    """
    
    @staticmethod
    def create(config: Dict[str, Any]) -> UniConnectNotifier:
        """
        Cria uma instância do notificador UniConnect com as configurações fornecidas.
        
        Args:
            config: Configurações para o notificador
            
        Returns:
            Instância do UniConnectNotifier
        """
        return UniConnectNotifier(config)