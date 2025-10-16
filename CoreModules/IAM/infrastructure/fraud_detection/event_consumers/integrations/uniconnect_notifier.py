#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Integração com o Módulo UniConnect para Notificações de Alertas Comportamentais

Este módulo fornece a integração entre o sistema de análise comportamental
e o módulo UniConnect para envio de notificações omnicanal sobre alertas 
de segurança e anomalias comportamentais detectadas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import json
import logging
import requests
import time
from datetime import datetime
from typing import Dict, Any, List, Optional, Union
from enum import Enum
import hmac
import hashlib
import base64
from dataclasses import dataclass

# Configuração do logger
logger = logging.getLogger("iam.trustguard.notifier")


class NotificationChannel(Enum):
    """Canais disponíveis para notificações via UniConnect."""
    
    EMAIL = "email"
    SMS = "sms"
    PUSH = "push"
    WHATSAPP = "whatsapp"
    IN_APP = "in_app"
    VOICE = "voice"
    TELEGRAM = "telegram"
    WEBSOCKET = "websocket"
    MOBILE_BANKING = "mobile_banking"
    ALL = "all"


class NotificationPriority(Enum):
    """Níveis de prioridade para notificações."""
    
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"
    EMERGENCY = "emergency"


@dataclass
class NotificationRecipient:
    """Representa um destinatário para notificações."""
    
    user_id: str
    channels: List[NotificationChannel]
    region_code: str
    role: Optional[str] = None
    department: Optional[str] = None
    escalation_level: Optional[int] = None
    device_tokens: Optional[List[str]] = None
    email: Optional[str] = None
    phone_number: Optional[str] = None
    language_preference: Optional[str] = None


class UniConnectNotifier:
    """
    Classe responsável pela integração com o módulo UniConnect para envio 
    de notificações relacionadas a alertas de análise comportamental.
    
    Esta classe gerencia o envio de alertas através de diversos canais,
    incluindo email, SMS, push notifications, WhatsApp e outros, seguindo
    as políticas de notificação e matrizes de autorização definidas.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o notificador com as configurações necessárias.
        
        Args:
            config: Configurações para conexão com o UniConnect, incluindo:
                - base_url: URL base da API UniConnect
                - api_key: Chave de API para autenticação
                - api_secret: Secret para assinatura das requisições
                - tenant_id: ID do tenant na plataforma (para multi-tenancy)
                - timeout: Timeout para requisições em segundos
                - default_channels: Canais padrão para notificações
                - template_path: Caminho para templates de notificação
                - retry_config: Configurações para tentativas de reenvio
        """
        self.base_url = config.get("base_url", "https://api.uniconnect.innovabiz.com")
        self.api_key = config["api_key"]
        self.api_secret = config["api_secret"]
        self.tenant_id = config["tenant_id"]
        self.timeout = config.get("timeout", 10)
        self.default_channels = config.get("default_channels", [NotificationChannel.EMAIL, 
                                                              NotificationChannel.PUSH])
        self.template_path = config.get("template_path", "templates/notifications")
        self.retry_config = config.get("retry_config", {
            "max_retries": 3,
            "initial_backoff": 1,  # em segundos
            "backoff_factor": 2,
            "jitter": 0.1
        })
        
        # Preparação dos headers
        self.headers = {
            "Content-Type": "application/json",
            "Accept": "application/json",
            "X-API-Key": self.api_key,
            "X-Tenant-ID": self.tenant_id
        }
        
        # Cache de templates de notificação
        self.template_cache = {}
        
        logger.info(f"UniConnectNotifier inicializado para tenant {self.tenant_id}")

    def send_behavioral_alert(self, 
                             alert_data: Dict[str, Any], 
                             recipients: List[NotificationRecipient],
                             priority: NotificationPriority = NotificationPriority.HIGH,
                             template_name: str = "behavioral_alert",
                             region_code: str = "AO") -> Dict[str, Any]:
        """
        Envia alerta de anomalia comportamental para os destinatários especificados.
        
        Args:
            alert_data: Dados do alerta a ser enviado
            recipients: Lista de destinatários para o alerta
            priority: Prioridade do alerta
            template_name: Nome do template a ser utilizado
            region_code: Código da região para adaptação regional
            
        Returns:
            Dicionário com resultado da operação, incluindo IDs das notificações geradas
        """
        logger.info(f"Enviando alerta comportamental para {len(recipients)} destinatários. Prioridade: {priority.value}")
        
        # Preparação dos dados para a notificação
        notification_data = self._prepare_notification_data(
            alert_data=alert_data,
            priority=priority,
            template_name=template_name,
            region_code=region_code
        )
        
        # Enviar para cada destinatário com seus canais específicos
        notification_results = []
        for recipient in recipients:
            channels = recipient.channels or self.default_channels
            
            # Enviar para cada canal selecionado
            for channel in channels:
                try:
                    result = self._send_to_channel(
                        notification_data=notification_data,
                        recipient=recipient,
                        channel=channel
                    )
                    notification_results.append(result)
                    
                    logger.debug(f"Notificação enviada com sucesso para usuário {recipient.user_id} via {channel.value}")
                    
                except Exception as e:
                    logger.error(f"Falha ao enviar notificação para usuário {recipient.user_id} "
                                 f"via {channel.value}: {str(e)}")
                    
                    # Tentar reenviar em caso de falha
                    self._retry_notification(
                        notification_data=notification_data,
                        recipient=recipient,
                        channel=channel
                    )
        
        # Retornar resultados consolidados
        return {
            "success": all(result.get("success", False) for result in notification_results),
            "notification_ids": [result.get("notification_id") for result in notification_results 
                                if result.get("notification_id")],
            "failed_count": sum(1 for result in notification_results if not result.get("success", False)),
            "total_count": len(notification_results)
        }

    def send_fraud_report(self, 
                         report_data: Dict[str, Any], 
                         recipients: List[NotificationRecipient],
                         include_attachments: bool = True,
                         region_code: str = "AO") -> Dict[str, Any]:
        """
        Envia relatório detalhado de fraude ou investigação para os destinatários autorizados.
        
        Args:
            report_data: Dados do relatório a ser enviado
            recipients: Lista de destinatários para o relatório
            include_attachments: Se deve incluir anexos
            region_code: Código da região para adaptação regional
            
        Returns:
            Dicionário com resultado da operação
        """
        logger.info(f"Enviando relatório de fraude para {len(recipients)} destinatários")
        
        # Preparação dos dados do relatório
        notification_data = self._prepare_notification_data(
            alert_data=report_data,
            priority=NotificationPriority.HIGH,
            template_name="fraud_report",
            region_code=region_code
        )
        
        # Adicionar anexos se necessário
        if include_attachments and "report_attachments" in report_data:
            notification_data["attachments"] = report_data["report_attachments"]
        
        # Enviar relatório para cada destinatário (apenas email e in-app para relatórios)
        allowed_channels = [NotificationChannel.EMAIL, NotificationChannel.IN_APP]
        report_results = []
        
        for recipient in recipients:
            channels = [ch for ch in recipient.channels if ch in allowed_channels] or [NotificationChannel.EMAIL]
            
            for channel in channels:
                try:
                    result = self._send_to_channel(
                        notification_data=notification_data,
                        recipient=recipient,
                        channel=channel
                    )
                    report_results.append(result)
                    
                except Exception as e:
                    logger.error(f"Falha ao enviar relatório para usuário {recipient.user_id} "
                                 f"via {channel.value}: {str(e)}")
        
        return {
            "success": all(result.get("success", False) for result in report_results),
            "report_ids": [result.get("notification_id") for result in report_results 
                           if result.get("notification_id")],
            "failed_count": sum(1 for result in report_results if not result.get("success", False))
        }

    def send_remediation_instructions(self,
                                    alert_id: str,
                                    user_id: str,
                                    instructions: List[Dict[str, Any]],
                                    channels: List[NotificationChannel] = None) -> Dict[str, Any]:
        """
        Envia instruções de remediação para o usuário após detecção de fraude ou anomalia.
        
        Args:
            alert_id: ID do alerta relacionado
            user_id: ID do usuário destinatário
            instructions: Lista de instruções e passos a serem seguidos
            channels: Canais a serem utilizados
            
        Returns:
            Dicionário com resultado da operação
        """
        logger.info(f"Enviando instruções de remediação para usuário {user_id}")
        
        # Configurar canais (prioriza canais interativos para instruções)
        if not channels:
            channels = [NotificationChannel.PUSH, NotificationChannel.SMS, NotificationChannel.EMAIL]
        
        # Preparar dados para notificação
        notification_data = {
            "alert_id": alert_id,
            "type": "remediation_instructions",
            "title": "Instruções de Segurança",
            "instructions": instructions,
            "timestamp": datetime.now().isoformat()
        }
        
        # Criar destinatário
        recipient = NotificationRecipient(
            user_id=user_id,
            channels=[ch for ch in channels],
            region_code="AO"  # Valor padrão, deve ser substituído pelo valor real
        )
        
        # Enviar para cada canal
        results = []
        for channel in channels:
            try:
                result = self._send_to_channel(
                    notification_data=notification_data,
                    recipient=recipient,
                    channel=channel,
                    template_name="remediation_instructions"
                )
                results.append(result)
                
            except Exception as e:
                logger.error(f"Falha ao enviar instruções de remediação para usuário {user_id} "
                             f"via {channel.value}: {str(e)}")
        
        return {
            "success": any(result.get("success", False) for result in results),
            "notification_ids": [result.get("notification_id") for result in results 
                               if result.get("notification_id")]
        }

    def _prepare_notification_data(self,
                                 alert_data: Dict[str, Any],
                                 priority: NotificationPriority,
                                 template_name: str,
                                 region_code: str) -> Dict[str, Any]:
        """
        Prepara os dados para envio de notificação.
        
        Args:
            alert_data: Dados do alerta
            priority: Prioridade da notificação
            template_name: Nome do template
            region_code: Código da região
            
        Returns:
            Dados formatados para notificação
        """
        # Assegurar que temos campos obrigatórios
        if "alert_id" not in alert_data:
            alert_data["alert_id"] = f"alert-{int(time.time())}"
            
        if "timestamp" not in alert_data:
            alert_data["timestamp"] = datetime.now().isoformat()
        
        notification_data = {
            "template": template_name,
            "priority": priority.value,
            "region_code": region_code,
            "data": alert_data,
            "metadata": {
                "source": "behavioral_analysis",
                "module": "trustguard",
                "alert_type": alert_data.get("alert_type", "behavioral_anomaly"),
                "tenant_id": self.tenant_id
            }
        }
        
        return notification_data

    def _send_to_channel(self,
                       notification_data: Dict[str, Any],
                       recipient: NotificationRecipient,
                       channel: NotificationChannel,
                       template_name: Optional[str] = None) -> Dict[str, Any]:
        """
        Envia notificação para um canal específico.
        
        Args:
            notification_data: Dados da notificação
            recipient: Destinatário
            channel: Canal para envio
            template_name: Nome do template (opcional, sobrescreve o template em notification_data)
            
        Returns:
            Resultado da operação de envio
        """
        if template_name:
            notification_data["template"] = template_name
            
        # Configurar dados específicos do canal
        channel_data = self._prepare_channel_specific_data(recipient, channel)
        
        # Preparar payload para API
        payload = {
            "channel": channel.value,
            "recipient": {
                "user_id": recipient.user_id,
                "region_code": recipient.region_code,
                **channel_data
            },
            "notification": notification_data,
            "tracking": {
                "source_system": "IAM_TrustGuard",
                "request_id": f"req-{int(time.time())}-{recipient.user_id[:8]}"
            }
        }
        
        # Adicionar campos opcionais se disponíveis
        for field in ["role", "department", "escalation_level", "language_preference"]:
            if hasattr(recipient, field) and getattr(recipient, field):
                payload["recipient"][field] = getattr(recipient, field)
        
        # Assinar payload
        timestamp = int(time.time())
        signature = self._generate_signature(payload, timestamp)
        
        # Adicionar headers de autenticação
        headers = {**self.headers}
        headers["X-Timestamp"] = str(timestamp)
        headers["X-Signature"] = signature
        
        # Realizar a requisição
        endpoint = f"{self.base_url}/api/v2/notifications/send"
        response = requests.post(
            endpoint,
            headers=headers,
            json=payload,
            timeout=self.timeout
        )
        
        # Processar resposta
        try:
            response_data = response.json()
            
            if response.status_code == 200 and response_data.get("success"):
                return {
                    "success": True,
                    "notification_id": response_data.get("notification_id"),
                    "channel": channel.value,
                    "recipient_id": recipient.user_id,
                    "delivery_status": response_data.get("delivery_status")
                }
            else:
                logger.warning(f"Falha ao enviar notificação: {response.status_code}, "
                             f"resposta: {response_data}")
                return {
                    "success": False,
                    "error": response_data.get("error", {
                        "code": "API_ERROR",
                        "message": f"Erro na API: {response.status_code}"
                    }),
                    "channel": channel.value,
                    "recipient_id": recipient.user_id
                }
                
        except Exception as e:
            logger.error(f"Erro ao processar resposta da API UniConnect: {str(e)}")
            return {
                "success": False,
                "error": {
                    "code": "PROCESSING_ERROR",
                    "message": str(e)
                },
                "channel": channel.value,
                "recipient_id": recipient.user_id
            }

    def _retry_notification(self,
                          notification_data: Dict[str, Any],
                          recipient: NotificationRecipient,
                          channel: NotificationChannel) -> None:
        """
        Implementa lógica de retry para notificações com backoff exponencial.
        
        Args:
            notification_data: Dados da notificação
            recipient: Destinatário
            channel: Canal para envio
        """
        retry_count = 0
        max_retries = self.retry_config["max_retries"]
        backoff = self.retry_config["initial_backoff"]
        
        while retry_count < max_retries:
            retry_count += 1
            
            try:
                # Adicionar delay exponencial com jitter
                jitter = self.retry_config["jitter"] * backoff * (0.5 - (time.time() % 1))
                retry_delay = backoff + jitter
                time.sleep(retry_delay)
                
                # Tentar novamente o envio
                logger.info(f"Tentativa {retry_count}/{max_retries} para usuário {recipient.user_id} "
                          f"via {channel.value}")
                
                result = self._send_to_channel(
                    notification_data=notification_data,
                    recipient=recipient,
                    channel=channel
                )
                
                if result.get("success"):
                    logger.info(f"Notificação reenviada com sucesso na tentativa {retry_count}")
                    return
                
                # Aumentar backoff para próxima tentativa
                backoff *= self.retry_config["backoff_factor"]
                
            except Exception as e:
                logger.error(f"Falha na tentativa {retry_count}/{max_retries}: {str(e)}")
                backoff *= self.retry_config["backoff_factor"]
        
        logger.error(f"Todas as tentativas de reenvio falharam para usuário {recipient.user_id}")

    def _prepare_channel_specific_data(self, 
                                    recipient: NotificationRecipient,
                                    channel: NotificationChannel) -> Dict[str, Any]:
        """
        Prepara dados específicos para cada canal de notificação.
        
        Args:
            recipient: Destinatário da notificação
            channel: Canal de envio
            
        Returns:
            Dados específicos do canal
        """
        channel_data = {}
        
        # Email
        if channel == NotificationChannel.EMAIL and recipient.email:
            channel_data["email"] = recipient.email
            
        # SMS ou Voice ou WhatsApp
        elif channel in [NotificationChannel.SMS, 
                         NotificationChannel.VOICE, 
                         NotificationChannel.WHATSAPP] and recipient.phone_number:
            channel_data["phone_number"] = recipient.phone_number
            
        # Push
        elif channel == NotificationChannel.PUSH and recipient.device_tokens:
            channel_data["device_tokens"] = recipient.device_tokens
            
        # Mobile Banking
        elif channel == NotificationChannel.MOBILE_BANKING:
            # Este canal usa apenas o user_id, não precisa de dados adicionais
            pass
        
        return channel_data

    def _generate_signature(self, payload: Dict[str, Any], timestamp: int) -> str:
        """
        Gera assinatura HMAC para autenticação.
        
        Args:
            payload: Dados a serem enviados
            timestamp: Timestamp atual em segundos
            
        Returns:
            Assinatura codificada em Base64
        """
        # Converter payload para string e concatenar com timestamp
        payload_str = json.dumps(payload, separators=(',', ':'), sort_keys=True)
        message = f"{payload_str}.{timestamp}.{self.tenant_id}"
        
        # Gerar HMAC-SHA256
        h = hmac.new(
            self.api_secret.encode('utf-8'),
            message.encode('utf-8'),
            hashlib.sha256
        )
        
        # Retornar assinatura em Base64
        return base64.b64encode(h.digest()).decode('utf-8')


def create_uniconnect_notifier(config: Dict[str, Any]) -> UniConnectNotifier:
    """
    Factory function para criar instâncias do notificador UniConnect.
    
    Args:
        config: Configurações do notificador
        
    Returns:
        Instância configurada do UniConnectNotifier
    """
    return UniConnectNotifier(config)