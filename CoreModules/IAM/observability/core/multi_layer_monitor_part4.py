"""
Monitor de Observabilidade Multi-Camada para a plataforma INNOVABIZ - Parte 4.

Implementação da camada de negócios e integrador de camadas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ - Sistema de Observabilidade Multi-Camada
Data: 20/08/2025
"""

import os
import time
import json
import inspect
import functools
import traceback
import asyncio
from typing import Dict, List, Optional, Union, Any, Callable, TypeVar, cast
from datetime import datetime
from enum import Enum

from .multi_layer_config import (
    MultiLayerObservabilityConfig,
    MultiLayerObservabilityConfigurator,
    ObservabilityLevel
)
from .multi_layer_monitor import (
    F, AsyncF, EventContext, EventSeverity, EventCategory,
    BUSINESS_TRANSACTION_COUNTER, BUSINESS_TRANSACTION_VALUE,
    BUSINESS_PROCESS_DURATION
)


class BusinessLayerMonitor:
    """
    Monitor específico para a camada de negócio da plataforma.
    
    Fornece decoradores e utilitários para monitorar processos de negócio,
    transações financeiras e fluxos operacionais.
    """
    
    def __init__(self, parent_monitor):
        """
        Inicializa o monitor de camada de negócio.
        
        Args:
            parent_monitor: Monitor principal de observabilidade
        """
        self.parent = parent_monitor
        self.config = parent_monitor.config
        self.logger = parent_monitor.logger
        self.tracer = parent_monitor.tracer
        self.meter = parent_monitor.meter
        
        # Inicializar contadores e histogramas específicos de negócio
        if self.meter:
            self.transaction_counter = self.meter.create_counter(
                BUSINESS_TRANSACTION_COUNTER,
                description="Número de transações de negócio",
                unit="1"
            )
            self.transaction_value = self.meter.create_histogram(
                BUSINESS_TRANSACTION_VALUE,
                description="Valor das transações de negócio",
                unit="currency"
            )
            self.process_duration = self.meter.create_histogram(
                BUSINESS_PROCESS_DURATION,
                description="Duração dos processos de negócio",
                unit="ms"
            )
    
    def trace_business_transaction(
        self,
        transaction_type: str,
        business_unit: Optional[str] = None,
        include_value: bool = True,
        mask_sensitive: bool = True
    ) -> Callable[[F], F]:
        """
        Decorador para rastrear transações de negócio.
        
        Args:
            transaction_type: Tipo da transação (payment, transfer, loan, etc.)
            business_unit: Unidade de negócio associada à transação
            include_value: Se deve incluir o valor da transação nas métricas
            mask_sensitive: Se deve mascarar dados sensíveis em logs e traces
            
        Returns:
            Callable: Decorador para função
        """
        def decorator(func: F) -> F:
            @functools.wraps(func)
            def wrapper(*args, **kwargs):
                # Determinar o nome do span
                name = f"business.transaction.{transaction_type}"
                
                # Extrair contexto da transação
                transaction_context = self._extract_transaction_context(args, kwargs)
                transaction_id = transaction_context.get("transaction.id")
                
                # Atributos base para telemetria
                attributes = {
                    "tenant_id": self.config.tenant_id,
                    "module_name": self.config.module_name,
                    "transaction.type": transaction_type,
                    "business.layer": "transaction"
                }
                
                if self.config.component_name:
                    attributes["component_name"] = self.config.component_name
                
                if business_unit:
                    attributes["business.unit"] = business_unit
                
                # Adicionar atributos do contexto da transação
                for key, value in transaction_context.items():
                    if key != "transaction.value" or include_value:
                        attributes[key] = value
                
                # Iniciar medição de tempo
                start_time = time.time()
                
                # Criar evento de contexto
                event_context = self.parent.create_event_context(
                    user_id=transaction_context.get("user.id"),
                    transaction_id=transaction_id,
                    additional_attributes={
                        "transaction.type": transaction_type,
                        **({"business.unit": business_unit} if business_unit else {})
                    }
                )
                
                # Se mascaramento estiver ativado, mascarar dados sensíveis
                if mask_sensitive:
                    self._mask_sensitive_transaction_data(transaction_context, event_context)
                
                # Criar e iniciar span se tracing estiver habilitado
                span = None
                if self.tracer and self.config.tracing_enabled:
                    span = self.tracer.start_span(name, attributes=attributes)
                
                try:
                    # Incrementar contador de transações
                    if self.meter and hasattr(self, "transaction_counter"):
                        self.transaction_counter.add(1, attributes)
                    
                    # Registrar valor da transação se disponível e configurado
                    transaction_value = transaction_context.get("transaction.value")
                    if transaction_value is not None and include_value and self.meter and hasattr(self, "transaction_value"):
                        try:
                            value = float(transaction_value)
                            self.transaction_value.record(value, attributes)
                            if span:
                                span.set_attribute("transaction.value", value)
                        except (ValueError, TypeError):
                            pass
                    
                    # Registrar início da transação
                    self.parent.log_event(
                        message=f"Iniciando transação de negócio: tipo={transaction_type}, ID={transaction_id}",
                        context=event_context,
                        severity=EventSeverity.INFO,
                        category=EventCategory.BUSINESS
                    )
                    
                    # Chamar função original
                    result = func(*args, **kwargs)
                    
                    # Extrair status do resultado
                    transaction_status = self._extract_transaction_status(result)
                    
                    if span:
                        span.set_attribute("transaction.status", transaction_status)
                    
                    # Registrar conclusão da transação
                    self.parent.log_event(
                        message=f"Transação de negócio concluída: tipo={transaction_type}, ID={transaction_id}, status={transaction_status}",
                        context=event_context,
                        severity=EventSeverity.INFO,
                        category=EventCategory.BUSINESS,
                        metadata={"transaction_status": transaction_status}
                    )
                    
                    return result
                
                except Exception as e:
                    # Registrar erro na transação
                    if span:
                        span.set_attribute("transaction.status", "ERROR")
                        span.set_attribute("error.type", e.__class__.__name__)
                        span.set_attribute("error.message", str(e))
                        span.record_exception(e)
                    
                    self.parent.log_event(
                        message=f"Erro na transação de negócio: tipo={transaction_type}, ID={transaction_id}, erro={str(e)}",
                        context=event_context,
                        severity=EventSeverity.ERROR,
                        category=EventCategory.BUSINESS,
                        exception=e
                    )
                    
                    # Re-lançar exceção
                    raise
                
                finally:
                    # Finalizar métricas de duração
                    duration_ms = (time.time() - start_time) * 1000
                    if self.meter and hasattr(self, "process_duration"):
                        self.process_duration.record(duration_ms, attributes)
                    
                    # Finalizar span
                    if span:
                        span.__exit__(None, None, None)
            
            # Suporte para funções assíncronas
            if asyncio.iscoroutinefunction(func):
                @functools.wraps(func)
                async def async_wrapper(*args, **kwargs):
                    # Implementação assíncrona similar à síncrona
                    name = f"business.transaction.{transaction_type}"
                    transaction_context = self._extract_transaction_context(args, kwargs)
                    transaction_id = transaction_context.get("transaction.id")
                    
                    attributes = {
                        "tenant_id": self.config.tenant_id,
                        "module_name": self.config.module_name,
                        "transaction.type": transaction_type,
                        "business.layer": "transaction"
                    }
                    
                    if self.config.component_name:
                        attributes["component_name"] = self.config.component_name
                    
                    if business_unit:
                        attributes["business.unit"] = business_unit
                    
                    for key, value in transaction_context.items():
                        if key != "transaction.value" or include_value:
                            attributes[key] = value
                    
                    start_time = time.time()
                    
                    event_context = self.parent.create_event_context(
                        user_id=transaction_context.get("user.id"),
                        transaction_id=transaction_id,
                        additional_attributes={
                            "transaction.type": transaction_type,
                            **({"business.unit": business_unit} if business_unit else {})
                        }
                    )
                    
                    if mask_sensitive:
                        self._mask_sensitive_transaction_data(transaction_context, event_context)
                    
                    span = None
                    if self.tracer and self.config.tracing_enabled:
                        span = self.tracer.start_span(name, attributes=attributes)
                    
                    try:
                        if self.meter and hasattr(self, "transaction_counter"):
                            self.transaction_counter.add(1, attributes)
                        
                        transaction_value = transaction_context.get("transaction.value")
                        if transaction_value is not None and include_value and self.meter and hasattr(self, "transaction_value"):
                            try:
                                value = float(transaction_value)
                                self.transaction_value.record(value, attributes)
                                if span:
                                    span.set_attribute("transaction.value", value)
                            except (ValueError, TypeError):
                                pass
                        
                        self.parent.log_event(
                            message=f"Iniciando transação de negócio assíncrona: tipo={transaction_type}, ID={transaction_id}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.BUSINESS
                        )
                        
                        result = await func(*args, **kwargs)
                        
                        transaction_status = self._extract_transaction_status(result)
                        
                        if span:
                            span.set_attribute("transaction.status", transaction_status)
                        
                        self.parent.log_event(
                            message=f"Transação de negócio assíncrona concluída: tipo={transaction_type}, ID={transaction_id}, status={transaction_status}",
                            context=event_context,
                            severity=EventSeverity.INFO,
                            category=EventCategory.BUSINESS,
                            metadata={"transaction_status": transaction_status}
                        )
                        
                        return result
                    
                    except Exception as e:
                        if span:
                            span.set_attribute("transaction.status", "ERROR")
                            span.set_attribute("error.type", e.__class__.__name__)
                            span.set_attribute("error.message", str(e))
                            span.record_exception(e)
                        
                        self.parent.log_event(
                            message=f"Erro na transação de negócio assíncrona: tipo={transaction_type}, ID={transaction_id}, erro={str(e)}",
                            context=event_context,
                            severity=EventSeverity.ERROR,
                            category=EventCategory.BUSINESS,
                            exception=e
                        )
                        
                        raise
                    
                    finally:
                        duration_ms = (time.time() - start_time) * 1000
                        if self.meter and hasattr(self, "process_duration"):
                            self.process_duration.record(duration_ms, attributes)
                        
                        if span:
                            span.__exit__(None, None, None)
                
                return cast(F, async_wrapper)
            
            return cast(F, wrapper)
        
        return decorator
    
    def _extract_transaction_context(self, args, kwargs) -> Dict[str, Any]:
        """
        Extrai contexto da transação dos argumentos da função.
        
        Returns:
            Dict[str, Any]: Contexto da transação
        """
        context = {}
        
        # Extrair de parâmetros nomeados
        transaction_id_keys = ["transaction_id", "transactionId", "id"]
        for key in transaction_id_keys:
            if key in kwargs:
                context["transaction.id"] = str(kwargs[key])
                break
        
        # Extrair valor da transação
        value_keys = ["amount", "value", "transaction_value", "transactionValue"]
        for key in value_keys:
            if key in kwargs:
                try:
                    context["transaction.value"] = float(kwargs[key])
                except (ValueError, TypeError):
                    pass
                break
        
        # Extrair usuário
        user_keys = ["user_id", "userId", "customer_id", "customerId"]
        for key in user_keys:
            if key in kwargs:
                context["user.id"] = str(kwargs[key])
                break
        
        # Extrair moeda
        currency_keys = ["currency", "currencyCode"]
        for key in currency_keys:
            if key in kwargs:
                context["transaction.currency"] = str(kwargs[key])
                break
        
        # Examinar objetos posicionais
        for arg in args:
            # Verificar se é um objeto de transação
            if hasattr(arg, "id") or hasattr(arg, "transaction_id") or hasattr(arg, "transactionId"):
                # Extrair ID
                for attr in ["id", "transaction_id", "transactionId"]:
                    if hasattr(arg, attr):
                        context["transaction.id"] = str(getattr(arg, attr))
                        break
                
                # Extrair valor
                for attr in ["amount", "value"]:
                    if hasattr(arg, attr):
                        try:
                            context["transaction.value"] = float(getattr(arg, attr))
                        except (ValueError, TypeError):
                            pass
                        break
                
                # Extrair usuário
                for attr in ["user_id", "userId", "customer_id", "customerId"]:
                    if hasattr(arg, attr):
                        context["user.id"] = str(getattr(arg, attr))
                        break
                
                # Extrair moeda
                for attr in ["currency", "currencyCode"]:
                    if hasattr(arg, attr):
                        context["transaction.currency"] = str(getattr(arg, attr))
                        break
        
        # Gerar ID de transação se não encontrado
        if "transaction.id" not in context:
            context["transaction.id"] = f"tx-{int(time.time())}"
        
        return context
    
    def _extract_transaction_status(self, result) -> str:
        """
        Extrai status da transação do resultado da função.
        
        Args:
            result: Resultado da função
            
        Returns:
            str: Status da transação (SUCCESS, PENDING, FAILED, etc.)
        """
        # Verificar atributos comuns para status
        if hasattr(result, "status"):
            return str(getattr(result, "status")).upper()
        elif hasattr(result, "transaction_status"):
            return str(getattr(result, "transaction_status")).upper()
        elif hasattr(result, "transactionStatus"):
            return str(getattr(result, "transactionStatus")).upper()
        elif isinstance(result, dict):
            for key in ["status", "transaction_status", "transactionStatus"]:
                if key in result:
                    return str(result[key]).upper()
        
        # Status padrão em caso de sucesso sem status específico
        return "SUCCESS"
    
    def _mask_sensitive_transaction_data(self, transaction_context: Dict[str, Any], event_context: EventContext) -> None:
        """
        Mascara dados sensíveis no contexto da transação para logs e traces.
        
        Args:
            transaction_context: Contexto extraído da transação
            event_context: Contexto do evento para logs
        """
        # Mascarar valor se presente no contexto
        if "transaction.value" in transaction_context:
            # Manter valor real para métricas, mas mascarar para logs
            masked_value = self._mask_currency_value(transaction_context["transaction.value"])
            event_context.add_attribute("transaction.value", masked_value)
        
        # Mascarar outros dados sensíveis se presentes
        sensitive_fields = ["account.number", "card.number", "document.id"]
        for field in sensitive_fields:
            if field in transaction_context:
                masked_value = self._mask_sensitive_field(transaction_context[field], field)
                event_context.add_attribute(field, masked_value)
    
    def _mask_currency_value(self, value) -> str:
        """
        Mascara valor monetário para exibição em logs.
        
        Args:
            value: Valor a ser mascarado
            
        Returns:
            str: Valor mascarado
        """
        try:
            # Converter para float se não for
            float_value = float(value)
            
            # Mascarar valores conforme regras de segurança da empresa
            if float_value >= 10000:
                return "***ALTO***"
            else:
                # Mascarar parcialmente (por exemplo, mostrar apenas a ordem de grandeza)
                order = 10 ** (len(str(int(float_value))) - 1)
                return f"{int(float_value / order)}X.XXX"
        except (ValueError, TypeError):
            return "***VALOR-INVÁLIDO***"
    
    def _mask_sensitive_field(self, value: str, field_type: str) -> str:
        """
        Mascara campo sensível conforme seu tipo.
        
        Args:
            value: Valor a ser mascarado
            field_type: Tipo do campo
            
        Returns:
            str: Valor mascarado
        """
        if field_type == "account.number":
            # Manter apenas os últimos 4 dígitos
            value_str = str(value)
            if len(value_str) > 4:
                return f"{'*' * (len(value_str) - 4)}{value_str[-4:]}"
            return "****"
        
        elif field_type == "card.number":
            # Formato padrão PCI DSS (primeiros 6, últimos 4)
            value_str = str(value).replace(" ", "").replace("-", "")
            if len(value_str) >= 13:  # Cartão de crédito válido
                return f"{value_str[:6]}******{value_str[-4:]}"
            return "************"
        
        elif field_type == "document.id":
            # Mascarar documentos (CPF, CNPJ, etc.)
            value_str = str(value)
            if len(value_str) > 6:
                visible_chars = min(3, len(value_str) // 3)
                return f"{value_str[:visible_chars]}{'*' * (len(value_str) - visible_chars*2)}{value_str[-visible_chars:]}"
            return "******"
        
        # Padrão para outros tipos
        return "********"


class UnifiedMonitor:
    """
    Monitor unificado que integra todas as camadas de observabilidade.
    
    Fornece acesso centralizado aos monitores de cada camada
    específica da plataforma.
    """
    
    def __init__(self, configurator: MultiLayerObservabilityConfigurator):
        """
        Inicializa o monitor unificado com o configurador fornecido.
        
        Args:
            configurator: Configurador de observabilidade multi-camada
        """
        self.configurator = configurator
        self.config = configurator.config
        self.logger = configurator.logger
        
        # Criar monitor base
        from .multi_layer_monitor_part2 import MultiLayerObservabilityMonitor
        self.base_monitor = MultiLayerObservabilityMonitor(configurator)
        
        # Inicializar monitores específicos de camada
        self.security = self._create_security_monitor()
        self.business = self._create_business_monitor()
        
        # Monitores adicionais serão inicializados sob demanda
        self._infrastructure_monitor = None
        self._user_monitor = None
    
    def _create_security_monitor(self):
        """
        Cria monitor para a camada de segurança.
        
        Returns:
            SecurityLayerMonitor: Monitor de segurança
        """
        from .multi_layer_monitor_part3 import SecurityLayerMonitor
        return SecurityLayerMonitor(self.base_monitor)
    
    def _create_business_monitor(self):
        """
        Cria monitor para a camada de negócio.
        
        Returns:
            BusinessLayerMonitor: Monitor de negócio
        """
        return BusinessLayerMonitor(self.base_monitor)
    
    @property
    def infrastructure(self):
        """
        Obtém o monitor para a camada de infraestrutura.
        
        Returns:
            InfrastructureLayerMonitor: Monitor de infraestrutura
        """
        # Inicializar sob demanda
        if not self._infrastructure_monitor:
            # A implementação real carregaria a classe apropriada
            # Por enquanto, retornamos o monitor base para placeholder
            self._infrastructure_monitor = self.base_monitor
        
        return self._infrastructure_monitor
    
    @property
    def user(self):
        """
        Obtém o monitor para a camada de usuário.
        
        Returns:
            UserLayerMonitor: Monitor de usuário
        """
        # Inicializar sob demanda
        if not self._user_monitor:
            # A implementação real carregaria a classe apropriada
            # Por enquanto, retornamos o monitor base para placeholder
            self._user_monitor = self.base_monitor
        
        return self._user_monitor
    
    def create_monitor_for_component(self, component_name: str):
        """
        Cria um novo monitor unificado para um componente específico.
        
        Args:
            component_name: Nome do componente
        
        Returns:
            UnifiedMonitor: Monitor unificado configurado para o componente
        """
        # Criar configurador atualizado com o nome do componente
        updated_config = self.config.copy()
        updated_config.component_name = component_name
        new_configurator = MultiLayerObservabilityConfigurator(updated_config, self.configurator.logger_name)
        
        # Retornar novo monitor unificado
        return UnifiedMonitor(new_configurator)
    
    @classmethod
    def from_env(cls, logger_name: str = "innovabiz.observability") -> 'UnifiedMonitor':
        """
        Cria um monitor unificado a partir de variáveis de ambiente.
        
        Args:
            logger_name: Nome base do logger
        
        Returns:
            UnifiedMonitor: Monitor unificado inicializado
        """
        # Criar configurador a partir de variáveis de ambiente
        configurator = MultiLayerObservabilityConfigurator.from_env(logger_name)
        
        # Criar monitor unificado
        return cls(configurator)