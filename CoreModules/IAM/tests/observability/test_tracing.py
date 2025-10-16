"""
Testes unitários para o módulo de rastreamento distribuído (tracing) do framework de observabilidade.

Estes testes verificam a integração com OpenTelemetry, propagação de contexto multi-tenant
e comportamento dos decoradores de instrumentação.
"""

import unittest
from unittest import mock
import asyncio
import json
import contextlib
from datetime import datetime

import pytest
from fastapi import FastAPI, Request, Response
from fastapi.testclient import TestClient
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider, Span
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from opentelemetry.sdk.trace.export.in_memory_span_exporter import InMemorySpanExporter

from src.observability.config import TracingConfig
from src.observability.tracing import (
    TracingIntegration,
    traced,
    traced_audit_event,
    traced_compliance_check,
    traced_retention_policy,
    SpanContextExtractor,
    TenantRegionInjector
)


class MockSpanProcessor:
    """Processador de span para testes."""
    
    def __init__(self):
        self.spans = []
    
    def on_start(self, span, parent_context=None):
        pass
    
    def on_end(self, span):
        self.spans.append(span)


@pytest.fixture
def memory_exporter():
    """Configura um exportador de spans em memória para testes."""
    return InMemorySpanExporter()


@pytest.fixture
def tracer_provider(memory_exporter):
    """Configura um provider de tracer para testes."""
    provider = TracerProvider()
    processor = SimpleSpanProcessor(memory_exporter)
    provider.add_span_processor(processor)
    return provider


@pytest.fixture
def tracing_integration():
    """Configura uma integração de tracing para testes."""
    config = TracingConfig(
        enabled=True,
        service_name="test-service",
        sample_ratio=1.0,
        otlp_endpoint="localhost:4317",
        resource_attributes={
            "service.name": "test-service",
            "service.version": "1.0.0",
            "deployment.environment": "test"
        }
    )
    
    tracing = TracingIntegration(config=config)
    tracing.initialize = mock.AsyncMock()
    tracing.shutdown = mock.AsyncMock()
    tracing._setup_tracer_provider = mock.MagicMock()
    
    return tracing


@pytest.mark.asyncio
class TestTracingDecorators:
    """Testes para os decoradores de tracing."""
    
    @pytest.fixture
    async def setup_tracer(self, tracer_provider, memory_exporter):
        """Configura o tracer global para testes."""
        # Salva o provider original para restaurar depois
        original_provider = trace.get_tracer_provider()
        
        # Define o provider de teste como global
        trace.set_tracer_provider(tracer_provider)
        
        yield memory_exporter
        
        # Restaura o provider original após o teste
        trace.set_tracer_provider(original_provider)
    
    @pytest.mark.asyncio
    async def test_traced_decorator(self, setup_tracer):
        """Testa o decorador traced para funções genéricas."""
        exporter = setup_tracer
        
        # Define uma função decorada para teste
        @traced(name="test.function")
        async def test_function(param1, param2):
            await asyncio.sleep(0.01)
            return f"{param1}-{param2}"
        
        # Executa a função
        result = await test_function("value1", "value2")
        assert result == "value1-value2"
        
        # Verifica se a span foi gerada corretamente
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert span.name == "test.function"
        
        # Limpa os spans para o próximo teste
        exporter.clear()
        
        # Testa com atributos personalizados
        @traced(
            name="test.with_attributes",
            attributes={"custom_attr": "custom_value"}
        )
        async def test_function_with_attrs(param):
            await asyncio.sleep(0.01)
            return param.upper()
        
        # Executa a função
        result = await test_function_with_attrs("hello")
        assert result == "HELLO"
        
        # Verifica se a span foi gerada com os atributos corretos
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert span.name == "test.with_attributes"
        assert span.attributes.get("custom_attr") == "custom_value"
    
    @pytest.mark.asyncio
    async def test_traced_audit_event(self, setup_tracer):
        """Testa o decorador traced_audit_event."""
        exporter = setup_tracer
        
        # Define uma classe de evento para teste
        class TestEvent:
            def __init__(self, event_id, event_type, tenant_id, user_id, details=None):
                self.event_id = event_id
                self.event_type = event_type
                self.tenant_id = tenant_id
                self.user_id = user_id
                self.details = details or {}
        
        # Define uma função decorada para teste
        @traced_audit_event()
        async def process_login_event(event, tenant, region):
            await asyncio.sleep(0.01)
            if event.details.get("error"):
                raise ValueError("Evento com erro")
            return {"status": "processed", "event_id": event.event_id}
        
        # Cria um evento de teste
        event = TestEvent(
            event_id="123",
            event_type="login",
            tenant_id="test_tenant",
            user_id="test_user",
            details={"source": "test"}
        )
        
        # Executa a função
        result = await process_login_event(event, "test_tenant", "test_region")
        assert result["status"] == "processed"
        
        # Verifica se a span foi gerada corretamente
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert "audit" in span.name.lower()
        assert "event_type" in span.attributes
        assert span.attributes["event_type"] == "login"
        assert span.attributes["tenant"] == "test_tenant"
        assert span.attributes["region"] == "test_region"
        
        # Limpa os spans para o próximo teste
        exporter.clear()
        
        # Testa com evento que gera erro
        event_error = TestEvent(
            event_id="456",
            event_type="login",
            tenant_id="test_tenant",
            user_id="test_user",
            details={"error": True}
        )
        
        # Executa a função com erro
        with pytest.raises(ValueError):
            await process_login_event(event_error, "test_tenant", "test_region")
        
        # Verifica se a span foi gerada com informações de erro
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert span.status.is_error
    
    @pytest.mark.asyncio
    async def test_traced_compliance_check(self, setup_tracer):
        """Testa o decorador traced_compliance_check."""
        exporter = setup_tracer
        
        # Define uma função decorada para teste
        @traced_compliance_check(compliance_type="gdpr")
        async def check_gdpr_compliance(record_id, tenant, region):
            await asyncio.sleep(0.01)
            if record_id == "invalid":
                return {"compliant": False, "reason": "Invalid record"}
            return {"compliant": True}
        
        # Executa a função
        result = await check_gdpr_compliance("record1", "test_tenant", "eu_west")
        assert result["compliant"] is True
        
        # Verifica se a span foi gerada corretamente
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert "compliance" in span.name.lower()
        assert span.attributes["compliance_type"] == "gdpr"
        assert span.attributes["tenant"] == "test_tenant"
        assert span.attributes["region"] == "eu_west"
        
        # Limpa os spans para o próximo teste
        exporter.clear()
        
        # Testa com registro não-conforme
        result = await check_gdpr_compliance("invalid", "test_tenant", "eu_west")
        assert result["compliant"] is False
        
        # Verifica se a span foi gerada com informações de não-conformidade
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        # O status não deve ser erro, pois não-conformidade não é um erro técnico
        assert not span.status.is_error
        assert span.attributes.get("compliant") is False
    
    @pytest.mark.asyncio
    async def test_traced_retention_policy(self, setup_tracer):
        """Testa o decorador traced_retention_policy."""
        exporter = setup_tracer
        
        # Define uma função decorada para teste
        @traced_retention_policy(policy_name="gdpr-30days")
        async def apply_retention_policy(tenant, region, days):
            await asyncio.sleep(0.01)
            if days <= 0:
                raise ValueError("Dias inválidos")
            return {
                "policy": "gdpr-30days",
                "affected_records": 42,
                "tenant": tenant,
                "region": region
            }
        
        # Executa a função
        result = await apply_retention_policy("test_tenant", "eu_west", 30)
        assert result["affected_records"] == 42
        
        # Verifica se a span foi gerada corretamente
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert "retention" in span.name.lower()
        assert span.attributes["policy_name"] == "gdpr-30days"
        assert span.attributes["tenant"] == "test_tenant"
        assert span.attributes["region"] == "eu_west"
        assert span.attributes["days"] == 30
        
        # Limpa os spans para o próximo teste
        exporter.clear()
        
        # Testa com erro
        with pytest.raises(ValueError):
            await apply_retention_policy("test_tenant", "eu_west", -1)
        
        # Verifica se a span foi gerada com informações de erro
        spans = exporter.get_finished_spans()
        assert len(spans) >= 1
        
        span = spans[-1]
        assert span.status.is_error


class TestTracingIntegration:
    """Testes para a integração de tracing com FastAPI."""
    
    @pytest.fixture
    def app_with_tracing(self, tracing_integration):
        """Configura uma aplicação FastAPI com tracing."""
        app = FastAPI()
        
        # Adiciona middlewares e rotas de teste
        tracing_integration.instrument_app(app)
        
        @app.get("/test")
        async def test_endpoint():
            return {"status": "ok"}
        
        @app.get("/test/{item_id}")
        async def test_with_param(item_id: str):
            return {"item_id": item_id}
        
        @app.get("/error")
        async def error_endpoint():
            raise ValueError("Test error")
        
        return TestClient(app)
    
    def test_middleware_headers(self, app_with_tracing):
        """Testa se o middleware propaga corretamente os headers de contexto."""
        # Faz uma requisição com headers personalizados
        response = app_with_tracing.get(
            "/test",
            headers={
                "X-Tenant-ID": "custom_tenant",
                "X-Region": "custom_region",
                "X-Environment": "custom_env"
            }
        )
        
        assert response.status_code == 200
        
        # Verifica se os headers foram propagados para a resposta
        assert response.headers.get("X-Tenant-ID") == "custom_tenant"
        assert response.headers.get("X-Region") == "custom_region"
    
    def test_path_normalization(self, tracing_integration):
        """Testa a normalização de path para evitar cardinalidade alta."""
        # Mock do método de normalização
        tracing_integration._normalize_path = TracingIntegration._normalize_path
        
        # Testa diferentes cenários
        assert tracing_integration._normalize_path("/users/123") == "/users/:id"
        assert tracing_integration._normalize_path("/orders/abc-123-def") == "/orders/:id"
        assert tracing_integration._normalize_path("/products/1/reviews") == "/products/:id/reviews"
        assert tracing_integration._normalize_path("/api/v1/items") == "/api/v1/items"
        assert tracing_integration._normalize_path("/tenant/123e4567-e89b-12d3-a456-426614174000/resources") == "/tenant/:id/resources"


class TestSpanContextExtractor:
    """Testes para o extrator de contexto de span."""
    
    def test_extract_tenant_region(self):
        """Testa a extração de tenant e região de uma requisição."""
        extractor = SpanContextExtractor()
        
        # Cria um mock de request
        request_mock = mock.MagicMock()
        request_mock.headers = {
            "X-Tenant-ID": "custom_tenant",
            "X-Region": "custom_region",
            "X-Environment": "custom_env"
        }
        
        # Extrai o contexto
        attributes = extractor.extract_context_attributes(request_mock)
        
        # Verifica os atributos extraídos
        assert attributes["tenant"] == "custom_tenant"
        assert attributes["region"] == "custom_region"
        assert attributes["environment"] == "custom_env"
    
    def test_extract_with_defaults(self):
        """Testa a extração com valores padrão quando headers não estão presentes."""
        extractor = SpanContextExtractor(
            default_tenant="default_tenant",
            default_region="default_region",
            default_environment="default_env"
        )
        
        # Cria um mock de request sem headers
        request_mock = mock.MagicMock()
        request_mock.headers = {}
        
        # Extrai o contexto
        attributes = extractor.extract_context_attributes(request_mock)
        
        # Verifica os atributos padrão
        assert attributes["tenant"] == "default_tenant"
        assert attributes["region"] == "default_region"
        assert attributes["environment"] == "default_env"


class TestTenantRegionInjector:
    """Testes para o injetor de tenant e região em spans."""
    
    def test_inject_context_to_span(self):
        """Testa a injeção de contexto em uma span."""
        injector = TenantRegionInjector()
        
        # Cria um mock de span
        span_mock = mock.MagicMock(spec=Span)
        
        # Injeta contexto
        injector.inject_context_to_span(
            span_mock,
            tenant="test_tenant",
            region="test_region",
            environment="test_env"
        )
        
        # Verifica se os atributos foram adicionados à span
        span_mock.set_attribute.assert_any_call("tenant", "test_tenant")
        span_mock.set_attribute.assert_any_call("region", "test_region")
        span_mock.set_attribute.assert_any_call("environment", "test_env")


if __name__ == "__main__":
    unittest.main()