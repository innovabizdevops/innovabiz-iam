"""
Testes de integração para o sistema de regras dinâmicas.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/Regras Dinâmicas
Data: 21/08/2025
"""

import os
import json
import unittest
import asyncio
from unittest import mock
from datetime import datetime, timedelta

import pytest
import httpx
from pydantic import BaseModel, Field

from observability.rules.rules_observability_config import ObservabilityConfig, RulesObservabilityConfigurator
from observability.rules.rules_observability_monitor import RulesObservabilityMonitor


# Modelos para os testes
class User(BaseModel):
    user_id: str
    name: str
    document_id: str
    risk_profile: str = "MEDIUM"
    creation_date: str = Field(default_factory=lambda: datetime.now().isoformat())
    last_login: str = Field(default_factory=lambda: datetime.now().isoformat())
    attributes: dict = Field(default_factory=dict)


class Transaction(BaseModel):
    transaction_id: str
    user_id: str
    amount: float
    currency: str = "BRL"
    timestamp: str = Field(default_factory=lambda: datetime.now().isoformat())
    transaction_type: str
    source_account: str
    destination_account: str
    status: str = "PENDING"
    risk_score: float = 0.0
    metadata: dict = Field(default_factory=dict)


class Rule(BaseModel):
    id: str
    name: str
    description: str
    condition: str
    risk_level: str = "MEDIUM"
    enabled: bool = True
    priority: int = 1
    tags: list = Field(default_factory=list)
    actions: list = Field(default_factory=list)


class RuleSet(BaseModel):
    id: str
    name: str
    description: str
    rules: list[Rule]
    enabled: bool = True
    version: str = "1.0.0"
    created_at: str = Field(default_factory=lambda: datetime.now().isoformat())
    updated_at: str = Field(default_factory=lambda: datetime.now().isoformat())


class RuleEvaluationResult(BaseModel):
    rule_id: str
    triggered: bool
    risk_level: str
    confidence: float = 1.0
    timestamp: str = Field(default_factory=lambda: datetime.now().isoformat())
    metadata: dict = Field(default_factory=dict)


class Event(BaseModel):
    id: str
    type: str
    user_id: str
    timestamp: str = Field(default_factory=lambda: datetime.now().isoformat())
    data: dict
    context: dict = Field(default_factory=dict)


class DynamicRulesEngine:
    """Mock da engine de regras dinâmicas para testes."""
    
    def __init__(self, monitor=None):
        self.monitor = monitor
        self.rules = {}
        self.bureau_connector = None
        self.trustguard_connector = None
        self._setup_default_rules()
    
    def _setup_default_rules(self):
        """Configura regras padrão para testes."""
        self.rules = {
            "ruleset1": RuleSet(
                id="ruleset1",
                name="Regras de Transação",
                description="Regras para avaliação de transações financeiras",
                rules=[
                    Rule(
                        id="rule1",
                        name="Transação de alto valor",
                        description="Detecta transações com valores acima de R$ 10.000",
                        condition="data.amount > 10000",
                        risk_level="HIGH",
                        priority=1,
                        tags=["transaction", "high-value"],
                    ),
                    Rule(
                        id="rule2",
                        name="Transação de novo destinatário",
                        description="Detecta transações para contas que nunca receberam transferências do usuário",
                        condition="data.destination_account not in context.known_accounts",
                        risk_level="MEDIUM",
                        priority=2,
                        tags=["transaction", "new-account"],
                    ),
                    Rule(
                        id="rule3",
                        name="Transação noturna",
                        description="Detecta transações realizadas durante a noite",
                        condition="'23:00' < data.timestamp[-8:-3] < '06:00'",
                        risk_level="MEDIUM",
                        priority=3,
                        tags=["transaction", "time-based"],
                    ),
                ]
            )
        }
    
    def set_bureau_connector(self, connector):
        """Define o conector do Bureau de Créditos."""
        self.bureau_connector = connector
    
    def set_trustguard_connector(self, connector):
        """Define o conector do TrustGuard."""
        self.trustguard_connector = connector
    
    async def evaluate_rules(self, ruleset_id: str, event: Event) -> dict[str, RuleEvaluationResult]:
        """
        Avalia um evento contra um conjunto de regras.
        
        Args:
            ruleset_id: ID do conjunto de regras
            event: Evento a ser avaliado
            
        Returns:
            dict[str, RuleEvaluationResult]: Resultados da avaliação por ID de regra
        """
        if self.monitor:
            # Se temos um monitor, usar o decorador de tracing
            return await self._evaluate_rules_with_monitoring(ruleset_id, event)
        
        return await self._evaluate_rules_internal(ruleset_id, event)
    
    async def _evaluate_rules_with_monitoring(self, ruleset_id: str, event: Event) -> dict[str, RuleEvaluationResult]:
        """Versão com monitoramento da avaliação de regras."""
        # Este método seria decorado pelo monitor.trace_rule_evaluation na implementação real
        start_time = datetime.now()
        
        try:
            result = await self._evaluate_rules_internal(ruleset_id, event)
            
            if self.monitor and hasattr(self.monitor, "rule_evaluation_counter"):
                self.monitor.rule_evaluation_counter.add(1, {"ruleset_id": ruleset_id})
                
                # Registrar regras acionadas
                triggered_rules = sum(1 for r in result.values() if r.triggered)
                self.monitor.rule_triggered_counter.add(
                    triggered_rules, {"ruleset_id": ruleset_id}
                )
            
            return result
            
        except Exception as e:
            if self.monitor and hasattr(self.monitor, "logger"):
                self.monitor.logger.error(
                    f"Erro na avaliação de regras: ruleset_id={ruleset_id}, "
                    f"event_type={event.type}, erro={str(e)}"
                )
            raise
    
    async def _evaluate_rules_internal(self, ruleset_id: str, event: Event) -> dict[str, RuleEvaluationResult]:
        """
        Implementação interna da avaliação de regras.
        
        Args:
            ruleset_id: ID do conjunto de regras
            event: Evento a ser avaliado
            
        Returns:
            dict[str, RuleEvaluationResult]: Resultados da avaliação por ID de regra
        """
        if ruleset_id not in self.rules:
            raise ValueError(f"Conjunto de regras não encontrado: {ruleset_id}")
        
        ruleset = self.rules[ruleset_id]
        results = {}
        
        # Se temos um conector de Bureau de Créditos, enriquece o evento
        if self.bureau_connector and hasattr(event.data, "user_id"):
            try:
                bureau_data = await self.bureau_connector.get_user_financial_data(
                    event.data.user_id
                )
                event.context["bureau_data"] = bureau_data
            except Exception as e:
                if self.monitor and hasattr(self.monitor, "logger"):
                    self.monitor.logger.warning(
                        f"Erro ao obter dados do Bureau: user_id={event.data.user_id}, erro={str(e)}"
                    )
        
        # Avalia cada regra
        for rule in ruleset.rules:
            if not rule.enabled:
                continue
                
            try:
                # Avaliação mock simplificada - na implementação real, seria uma avaliação dinâmica da condição
                triggered = False
                
                # Simula algumas condições comuns
                data = event.data
                context = event.context
                
                # Regra de transação de alto valor
                if rule.id == "rule1" and hasattr(data, "amount") and data.amount > 10000:
                    triggered = True
                
                # Regra de novo destinatário
                elif (rule.id == "rule2" and hasattr(data, "destination_account") and 
                      context.get("known_accounts") and 
                      data.destination_account not in context["known_accounts"]):
                    triggered = True
                
                # Regra de horário
                elif rule.id == "rule3":
                    try:
                        timestamp = data.timestamp
                        hour = int(timestamp.split("T")[1].split(":")[0])
                        if 23 <= hour or hour <= 6:
                            triggered = True
                    except (IndexError, ValueError, AttributeError):
                        pass
                
                # Criar resultado
                result = RuleEvaluationResult(
                    rule_id=rule.id,
                    triggered=triggered,
                    risk_level=rule.risk_level if triggered else "LOW",
                    confidence=0.95,
                    metadata={
                        "rule_name": rule.name,
                        "evaluation_type": "mock",
                    }
                )
                results[rule.id] = result
                
            except Exception as e:
                if self.monitor and hasattr(self.monitor, "logger"):
                    self.monitor.logger.error(
                        f"Erro ao avaliar regra: ruleset_id={ruleset_id}, "
                        f"rule_id={rule.id}, erro={str(e)}"
                    )
                # Regra não acionada em caso de erro
                results[rule.id] = RuleEvaluationResult(
                    rule_id=rule.id,
                    triggered=False,
                    risk_level="ERROR",
                    confidence=0.0,
                    metadata={
                        "rule_name": rule.name,
                        "error": str(e),
                    }
                )
        
        return results


class TestDynamicRulesIntegration:
    """Testes de integração para o sistema de regras dinâmicas."""
    
    @pytest.fixture
    def configurator(self):
        """Fixture para criar configurador de observabilidade."""
        config = ObservabilityConfig(
            service_name="test-rules-service",
            environment="test",
            version="1.0.0",
            tenant_id="test-tenant",
            log_level="INFO",
            tracing_enabled=True,
            metrics_enabled=True,
        )
        return RulesObservabilityConfigurator(config, "test.rules")
    
    @pytest.fixture
    def monitor(self, configurator):
        """Fixture para criar monitor de observabilidade."""
        return RulesObservabilityMonitor(configurator, "rules-engine")
    
    @pytest.fixture
    def rules_engine(self, monitor):
        """Fixture para criar engine de regras."""
        engine = DynamicRulesEngine(monitor)
        return engine
    
    @pytest.fixture
    def mock_bureau_connector(self):
        """Fixture para criar mock do conector do Bureau de Créditos."""
        connector = mock.AsyncMock()
        connector.get_user_financial_data = mock.AsyncMock(return_value={
            "credit_score": 750,
            "last_check": datetime.now().isoformat(),
            "risk_level": "LOW",
            "payment_history": "GOOD",
            "debt_level": "LOW",
        })
        return connector
    
    @pytest.fixture
    def mock_trustguard_connector(self):
        """Fixture para criar mock do conector do TrustGuard."""
        connector = mock.AsyncMock()
        connector.evaluate_access = mock.AsyncMock(return_value=mock.MagicMock(
            decision="ALLOW",
            risk_level="LOW",
        ))
        return connector
    
    @pytest.mark.asyncio
    async def test_rules_evaluation_basic(self, rules_engine):
        """Testa avaliação básica de regras."""
        # Criar evento de transação
        transaction = Transaction(
            transaction_id="tx123",
            user_id="user456",
            amount=5000.0,
            transaction_type="TRANSFER",
            source_account="123456",
            destination_account="789012",
        )
        
        event = Event(
            id="evt789",
            type="TRANSACTION",
            user_id="user456",
            data=transaction,
            context={
                "known_accounts": ["789012"],  # Conta conhecida
            },
        )
        
        # Avaliar regras
        results = await rules_engine.evaluate_rules("ruleset1", event)
        
        # Verificar resultados
        assert len(results) == 3
        assert not results["rule1"].triggered  # Não deve acionar (valor abaixo do limite)
        assert not results["rule2"].triggered  # Não deve acionar (conta conhecida)
    
    @pytest.mark.asyncio
    async def test_rules_evaluation_high_value(self, rules_engine):
        """Testa avaliação de regras para transações de alto valor."""
        # Criar evento de transação de alto valor
        transaction = Transaction(
            transaction_id="tx456",
            user_id="user789",
            amount=15000.0,  # Acima do limite
            transaction_type="TRANSFER",
            source_account="123456",
            destination_account="789012",
        )
        
        event = Event(
            id="evt012",
            type="TRANSACTION",
            user_id="user789",
            data=transaction,
            context={
                "known_accounts": ["789012"],
            },
        )
        
        # Avaliar regras
        results = await rules_engine.evaluate_rules("ruleset1", event)
        
        # Verificar resultados
        assert results["rule1"].triggered  # Deve acionar (valor acima do limite)
        assert results["rule1"].risk_level == "HIGH"
        assert not results["rule2"].triggered  # Não deve acionar (conta conhecida)
    
    @pytest.mark.asyncio
    async def test_rules_evaluation_new_account(self, rules_engine):
        """Testa avaliação de regras para transações para novas contas."""
        # Criar evento de transação para nova conta
        transaction = Transaction(
            transaction_id="tx789",
            user_id="user012",
            amount=3000.0,
            transaction_type="TRANSFER",
            source_account="123456",
            destination_account="345678",  # Nova conta
        )
        
        event = Event(
            id="evt345",
            type="TRANSACTION",
            user_id="user012",
            data=transaction,
            context={
                "known_accounts": ["789012", "901234"],  # Contas conhecidas
            },
        )
        
        # Avaliar regras
        results = await rules_engine.evaluate_rules("ruleset1", event)
        
        # Verificar resultados
        assert not results["rule1"].triggered  # Não deve acionar (valor abaixo do limite)
        assert results["rule2"].triggered  # Deve acionar (nova conta)
        assert results["rule2"].risk_level == "MEDIUM"
    
    @pytest.mark.asyncio
    async def test_rules_with_bureau_data(self, rules_engine, mock_bureau_connector):
        """Testa avaliação de regras com enriquecimento de dados do Bureau."""
        # Configurar conector
        rules_engine.set_bureau_connector(mock_bureau_connector)
        
        # Criar evento de transação
        transaction = Transaction(
            transaction_id="tx012",
            user_id="user345",
            amount=12000.0,  # Acima do limite
            transaction_type="TRANSFER",
            source_account="123456",
            destination_account="345678",  # Nova conta
        )
        
        event = Event(
            id="evt678",
            type="TRANSACTION",
            user_id="user345",
            data=transaction,
            context={
                "known_accounts": ["789012"],
            },
        )
        
        # Avaliar regras
        results = await rules_engine.evaluate_rules("ruleset1", event)
        
        # Verificar chamada ao Bureau
        mock_bureau_connector.get_user_financial_data.assert_called_once_with("user345")
        
        # Verificar resultados
        assert results["rule1"].triggered  # Deve acionar (valor acima do limite)
        assert results["rule1"].risk_level == "HIGH"
        assert results["rule2"].triggered  # Deve acionar (nova conta)
        assert results["rule2"].risk_level == "MEDIUM"
    
    @pytest.mark.asyncio
    async def test_rules_with_monitoring(self, rules_engine, monitor):
        """Testa monitoramento da avaliação de regras."""
        # Criar mocks para métricas
        monitor.rule_evaluation_counter = mock.MagicMock()
        monitor.rule_triggered_counter = mock.MagicMock()
        monitor.rule_evaluation_time = mock.MagicMock()
        monitor.logger = mock.MagicMock()
        
        # Criar evento de transação de alto valor
        transaction = Transaction(
            transaction_id="tx345",
            user_id="user678",
            amount=20000.0,  # Acima do limite
            transaction_type="TRANSFER",
            source_account="123456",
            destination_account="789012",
        )
        
        event = Event(
            id="evt901",
            type="TRANSACTION",
            user_id="user678",
            data=transaction,
            context={
                "known_accounts": ["789012"],
            },
        )
        
        # Avaliar regras
        results = await rules_engine.evaluate_rules("ruleset1", event)
        
        # Verificar chamadas às métricas
        monitor.rule_evaluation_counter.add.assert_called_once()
        monitor.rule_triggered_counter.add.assert_called_once()
        
        # Verificar resultados
        assert results["rule1"].triggered  # Deve acionar (valor acima do limite)
        assert results["rule1"].risk_level == "HIGH"
        assert not results["rule2"].triggered  # Não deve acionar (conta conhecida)


if __name__ == "__main__":
    pytest.main()