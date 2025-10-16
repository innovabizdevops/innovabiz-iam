# Templates de Integração para Observabilidade INNOVABIZ

## Visão Geral

Este documento fornece templates padronizados para integração de novos módulos e serviços com o Framework de Observabilidade INNOVABIZ. Os templates seguem as melhores práticas, padrões e requisitos estabelecidos nas quatro partes do Framework de Integração, garantindo consistência, qualidade e conformidade em toda a plataforma.

## Índice

1. [Template para Instrumentação OpenTelemetry](#template-para-instrumentação-opentelemetry)
2. [Template para Configuração do OpenTelemetry Collector](#template-para-configuração-do-opentelemetry-collector)
3. [Template para Dashboards Grafana](#template-para-dashboards-grafana)
4. [Template para Alertas Prometheus](#template-para-alertas-prometheus)
5. [Template para Documentação de Observabilidade](#template-para-documentação-de-observabilidade)
6. [Template para Runbooks Operacionais](#template-para-runbooks-operacionais)

## Template para Instrumentação OpenTelemetry

### Visão Geral da Instrumentação

Este template fornece a estrutura básica para instrumentar aplicações com OpenTelemetry, garantindo consistência nas métricas, logs e traces coletados em toda a plataforma INNOVABIZ.

### Node.js (Express)

```javascript
// Configuração do OpenTelemetry para Node.js (Express)
const { NodeSDK } = require('@opentelemetry/sdk-node');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');
const { ConsoleSpanExporter } = require('@opentelemetry/sdk-trace-node');
const { metrics } = require('@opentelemetry/api');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-http');
const { ExpressInstrumentation } = require('@opentelemetry/instrumentation-express');
const { HttpInstrumentation } = require('@opentelemetry/instrumentation-http');
const { MongoDBInstrumentation } = require('@opentelemetry/instrumentation-mongodb');
const { PgInstrumentation } = require('@opentelemetry/instrumentation-pg');
const { RedisInstrumentation } = require('@opentelemetry/instrumentation-redis');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics');
const { DnsInstrumentation } = require('@opentelemetry/instrumentation-dns');
const { W3CTraceContextPropagator } = require('@opentelemetry/core');
const { B3Propagator } = require('@opentelemetry/propagator-b3');
const { JaegerPropagator } = require('@opentelemetry/propagator-jaeger');

// Configuração do contexto multi-dimensional INNOVABIZ
const { InnovabizContextPropagator } = require('@innovabiz/observability-sdk');

// Função de inicialização da telemetria
function initializeOpenTelemetry(serviceName, moduleId, serviceVersion) {
  // Informações multi-contexto INNOVABIZ
  const resourceAttributes = {
    [SemanticResourceAttributes.SERVICE_NAME]: serviceName,
    [SemanticResourceAttributes.SERVICE_VERSION]: serviceVersion,
    'innovabiz.module.id': moduleId,
    'innovabiz.deployment.environment': process.env.ENVIRONMENT || 'development',
    'innovabiz.tenant.id': process.env.TENANT_ID || 'default',
    'innovabiz.region.id': process.env.REGION_ID || 'default',
  };

  // Configuração do coletor
  const collectorOptions = {
    url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318/v1/traces',
    headers: {
      'x-innovabiz-tenant-id': process.env.TENANT_ID || 'default',
      'x-innovabiz-region-id': process.env.REGION_ID || 'default',
    },
  };

  // Configuração do pipeline de métricas
  const metricReaders = [
    new PeriodicExportingMetricReader({
      exporter: new OTLPMetricExporter({
        url: process.env.OTEL_EXPORTER_OTLP_METRICS_ENDPOINT || 'http://localhost:4318/v1/metrics',
        headers: {
          'x-innovabiz-tenant-id': process.env.TENANT_ID || 'default',
          'x-innovabiz-region-id': process.env.REGION_ID || 'default',
        },
      }),
      exportIntervalMillis: 15000,
    }),
  ];

  // Configuração de instrumentação automática
  const instrumentations = [
    new HttpInstrumentation({
      ignoreIncomingPaths: ['/health', '/metrics', '/ready'],
    }),
    new ExpressInstrumentation(),
    new MongoDBInstrumentation(),
    new PgInstrumentation(),
    new RedisInstrumentation(),
    new DnsInstrumentation(),
  ];

  // Configuração de propagadores de contexto
  const propagators = {
    registeredPropagators: [
      new W3CTraceContextPropagator(),
      new B3Propagator(),
      new JaegerPropagator(),
      new InnovabizContextPropagator(), // Propagador personalizado INNOVABIZ
    ],
  };

  // Inicializa o SDK
  const sdk = new NodeSDK({
    resource: new Resource(resourceAttributes),
    traceExporter: new OTLPTraceExporter(collectorOptions),
    metricReaders: metricReaders,
    instrumentations: instrumentations,
    propagators: propagators,
  });

  // Inicializa e registra manipuladores de erro/desligamento
  sdk.start()
    .then(() => console.log('Telemetria iniciada com sucesso'))
    .catch((error) => console.error('Erro ao iniciar telemetria:', error));

  // Manipulador de desligamento
  process.on('SIGTERM', () => {
    sdk.shutdown()
      .then(() => console.log('Telemetria encerrada com sucesso'))
      .catch((error) => console.error('Erro ao encerrar telemetria:', error))
      .finally(() => process.exit(0));
  });

  return sdk;
}

// Exemplos de uso de métricas customizadas
function registerCustomMetrics() {
  const meter = metrics.getMeter('innovabiz-custom-metrics');
  
  // Counter para transações
  const transactionCounter = meter.createCounter('transactions.count', {
    description: 'Contador de transações processadas',
    unit: '1',
  });
  
  // Histogram para latência de transações
  const transactionDuration = meter.createHistogram('transaction.duration', {
    description: 'Duração das transações',
    unit: 'ms',
  });
  
  // Up/Down Counter para usuários ativos
  const activeUsers = meter.createUpDownCounter('users.active', {
    description: 'Usuários ativos no momento',
    unit: '1',
  });

  // Gauge para utilização de recursos
  const cpuUsage = meter.createObservableGauge('system.cpu.usage', {
    description: 'Uso de CPU',
    unit: '%',
    callback: (result) => {
      // Lógica para obter uso de CPU
      const usage = process.cpuUsage();
      result.observe(((usage.user + usage.system) / 1000000) * 100, {
        'innovabiz.resource.type': 'cpu',
      });
    },
  });

  return {
    transactionCounter,
    transactionDuration,
    activeUsers,
    cpuUsage,
  };
}

// Exporta funções
module.exports = {
  initializeOpenTelemetry,
  registerCustomMetrics,
};
```

### Python (FastAPI)

```python
# Configuração do OpenTelemetry para Python (FastAPI)
import os
from typing import Dict, Any
import logging
from opentelemetry import trace, metrics, context
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentation
from opentelemetry.instrumentation.requests import RequestsInstrumentation
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentation
from opentelemetry.instrumentation.redis import RedisInstrumentation
from opentelemetry.instrumentation.logging import LoggingInstrumentation
from opentelemetry.instrumentation.kafka import KafkaInstrumentation
from opentelemetry.propagate import set_global_textmap
from opentelemetry.propagators.composite import CompositePropagator
from opentelemetry.propagators.b3 import B3Format
from opentelemetry.propagators.jaeger import JaegerPropagator

# Propagador de Contexto INNOVABIZ personalizado
from innovabiz.observability.propagation import InnovabizContextPropagator

logger = logging.getLogger(__name__)

def setup_opentelemetry(service_name: str, module_id: str, service_version: str) -> None:
    """
    Configura OpenTelemetry para a aplicação
    
    Args:
        service_name: Nome do serviço
        module_id: ID do módulo INNOVABIZ
        service_version: Versão do serviço
    """
    # Informações multi-contexto INNOVABIZ
    resource = Resource.create({
        ResourceAttributes.SERVICE_NAME: service_name,
        ResourceAttributes.SERVICE_VERSION: service_version,
        "innovabiz.module.id": module_id,
        "innovabiz.deployment.environment": os.getenv("ENVIRONMENT", "development"),
        "innovabiz.tenant.id": os.getenv("TENANT_ID", "default"),
        "innovabiz.region.id": os.getenv("REGION_ID", "default"),
    })

    # Configuração do propagador composto
    set_global_textmap(CompositePropagator([
        InnovabizContextPropagator(),  # Propagador personalizado INNOVABIZ
        B3Format(),
        JaegerPropagator(),
    ]))

    # Configuração do tracer provider
    tracer_provider = TracerProvider(resource=resource)
    
    # Exportador OTLP para traces
    otlp_trace_exporter = OTLPSpanExporter(
        endpoint=os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317"),
        headers={
            "x-innovabiz-tenant-id": os.getenv("TENANT_ID", "default"),
            "x-innovabiz-region-id": os.getenv("REGION_ID", "default"),
        }
    )
    
    # Adiciona processador de spans
    tracer_provider.add_span_processor(BatchSpanProcessor(otlp_trace_exporter))
    
    # Se em ambiente de desenvolvimento, adiciona console exporter para debugging
    if os.getenv("ENVIRONMENT") == "development":
        tracer_provider.add_span_processor(BatchSpanProcessor(ConsoleSpanExporter()))
    
    # Define o tracer provider global
    trace.set_tracer_provider(tracer_provider)
    
    # Exportador OTLP para métricas
    otlp_metric_exporter = OTLPMetricExporter(
        endpoint=os.getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4317"),
        headers={
            "x-innovabiz-tenant-id": os.getenv("TENANT_ID", "default"),
            "x-innovabiz-region-id": os.getenv("REGION_ID", "default"),
        }
    )
    
    # Configuração do leitor de métricas
    metric_reader = PeriodicExportingMetricReader(
        exporter=otlp_metric_exporter,
        export_interval_millis=15000
    )
    
    # Configuração do meter provider
    metrics.set_meter_provider(MeterProvider(resource=resource, metric_readers=[metric_reader]))
    
    # Instrumentação automática
    FastAPIInstrumentation().instrument()
    RequestsInstrumentation().instrument()
    SQLAlchemyInstrumentation().instrument()
    RedisInstrumentation().instrument()
    LoggingInstrumentation().instrument()
    KafkaInstrumentation().instrument()
    
    logger.info(f"OpenTelemetry inicializado para {service_name} v{service_version}")

def create_custom_metrics():
    """
    Cria e registra métricas customizadas para o serviço
    """
    meter = metrics.get_meter("innovabiz-custom-metrics")
    
    # Counter para transações processadas
    transaction_counter = meter.create_counter(
        name="transactions.count",
        description="Contador de transações processadas",
        unit="1",
    )
    
    # Histogram para latência de transações
    transaction_duration = meter.create_histogram(
        name="transaction.duration",
        description="Duração das transações",
        unit="ms",
    )
    
    # Up/Down Counter para usuários ativos
    active_users = meter.create_up_down_counter(
        name="users.active",
        description="Usuários ativos no momento",
        unit="1",
    )
    
    # Observer para métricas do sistema
    system_memory = meter.create_observable_gauge(
        name="system.memory.usage",
        description="Uso de memória",
        unit="bytes",
        callbacks=[_get_memory_usage]
    )
    
    return {
        "transaction_counter": transaction_counter,
        "transaction_duration": transaction_duration,
        "active_users": active_users,
        "system_memory": system_memory
    }

def _get_memory_usage(observer):
    """
    Callback para observar uso de memória
    """
    import psutil
    memory = psutil.virtual_memory()
    observer.observe(memory.used, {"innovabiz.resource.type": "memory"})
```

### Java (Spring Boot)

```java
// Configuração do OpenTelemetry para Java (Spring Boot)
package com.innovabiz.observability.config;

import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.metrics.Meter;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.context.propagation.TextMapPropagator;
import io.opentelemetry.exporter.otlp.metrics.OtlpGrpcMetricExporter;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;
import io.opentelemetry.propagation.B3Propagator;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.metrics.SdkMeterProvider;
import io.opentelemetry.sdk.metrics.export.PeriodicMetricReader;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.time.Duration;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

import com.innovabiz.observability.propagation.InnovabizContextPropagator;

@Configuration
public class OpenTelemetryConfig {

    @Value("${spring.application.name}")
    private String serviceName;
    
    @Value("${innovabiz.module.id}")
    private String moduleId;
    
    @Value("${innovabiz.service.version}")
    private String serviceVersion;
    
    @Value("${innovabiz.tenant.id:default}")
    private String tenantId;
    
    @Value("${innovabiz.region.id:default}")
    private String regionId;
    
    @Value("${otel.exporter.otlp.endpoint:http://localhost:4317}")
    private String otlpEndpoint;
    
    @Value("${innovabiz.environment:development}")
    private String environment;

    @Bean
    public OpenTelemetrySdk openTelemetrySdk() {
        // Recurso com atributos multi-contexto INNOVABIZ
        Resource resource = Resource.getDefault()
                .merge(Resource.create(Attributes.builder()
                        .put(ResourceAttributes.SERVICE_NAME, serviceName)
                        .put(ResourceAttributes.SERVICE_VERSION, serviceVersion)
                        .put("innovabiz.module.id", moduleId)
                        .put("innovabiz.tenant.id", tenantId)
                        .put("innovabiz.region.id", regionId)
                        .put("innovabiz.deployment.environment", environment)
                        .build()));

        // Cabeçalhos para exportadores
        Map<String, String> headers = new HashMap<>();
        headers.put("x-innovabiz-tenant-id", tenantId);
        headers.put("x-innovabiz-region-id", regionId);

        // Configuração do exportador de spans
        OtlpGrpcSpanExporter spanExporter = OtlpGrpcSpanExporter.builder()
                .setEndpoint(otlpEndpoint)
                .setTimeout(10, TimeUnit.SECONDS)
                .setHeaders(headers)
                .build();

        // Configuração do provider de tracer
        SdkTracerProvider tracerProvider = SdkTracerProvider.builder()
                .setResource(resource)
                .addSpanProcessor(BatchSpanProcessor.builder(spanExporter)
                        .setMaxQueueSize(2048)
                        .setMaxExportBatchSize(512)
                        .setExporterTimeout(Duration.ofSeconds(30))
                        .build())
                .build();

        // Configuração do exportador de métricas
        OtlpGrpcMetricExporter metricExporter = OtlpGrpcMetricExporter.builder()
                .setEndpoint(otlpEndpoint)
                .setTimeout(10, TimeUnit.SECONDS)
                .setHeaders(headers)
                .build();

        // Configuração do provider de métricas
        SdkMeterProvider meterProvider = SdkMeterProvider.builder()
                .setResource(resource)
                .registerMetricReader(PeriodicMetricReader.builder(metricExporter)
                        .setInterval(Duration.ofMillis(15000))
                        .build())
                .build();

        // Configuração dos propagadores (incluindo o propagador personalizado INNOVABIZ)
        ContextPropagators propagators = ContextPropagators.create(
                TextMapPropagator.composite(
                        B3Propagator.injectingMultiHeaders(),
                        new InnovabizContextPropagator()));

        // Construção do SDK
        return OpenTelemetrySdk.builder()
                .setTracerProvider(tracerProvider)
                .setMeterProvider(meterProvider)
                .setPropagators(propagators)
                .build();
    }

    @Bean
    public Tracer tracer(OpenTelemetrySdk openTelemetrySdk) {
        return openTelemetrySdk.getTracer(serviceName, serviceVersion);
    }

    @Bean
    public Meter meter(OpenTelemetrySdk openTelemetrySdk) {
        return openTelemetrySdk.getMeter(serviceName);
    }
}
```

### Go

```go
// Configuração do OpenTelemetry para Go
package observability

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	
	// Importações para o propagador personalizado INNOVABIZ
	"github.com/innovabiz/observability-go/propagation"
)

// Configuração da inicialização da telemetria
func InitOpenTelemetry(serviceName, moduleID, serviceVersion string) (func(context.Context) error, error) {
	// Obter variáveis de ambiente
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		tenantID = "default"
	}
	
	regionID := os.Getenv("REGION_ID")
	if regionID == "" {
		regionID = "default"
	}
	
	// Configuração do endpoint OTLP
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4317"
	}
	
	// Criar recurso com atributos multi-contexto INNOVABIZ
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("innovabiz.module.id", moduleID),
			attribute.String("innovabiz.tenant.id", tenantID),
			attribute.String("innovabiz.region.id", regionID),
			attribute.String("innovabiz.deployment.environment", environment),
		),
	)
	if err != nil {
		return nil, err
	}
	
	// Headers para os exportadores
	headers := map[string]string{
		"x-innovabiz-tenant-id": tenantID,
		"x-innovabiz-region-id": regionID,
	}
	
	// Configuração do exportador de traces
	traceExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithTimeout(time.Second*10),
	)
	if err != nil {
		return nil, err
	}
	
	// Configuração do exportador de métricas
	metricExporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithEndpoint(otlpEndpoint),
		otlpmetricgrpc.WithHeaders(headers),
		otlpmetricgrpc.WithTimeout(time.Second*10),
	)
	if err != nil {
		return nil, err
	}
	
	// Configuração do processador de traces
	bsp := trace.NewBatchSpanProcessor(traceExporter)
	
	// Configuração do provider de traces
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)
	
	// Configuração do leitor de métricas
	metricReader := metric.NewPeriodicReader(
		metricExporter,
		metric.WithInterval(15*time.Second),
	)
	
	// Configuração do provider de métricas
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metricReader),
	)
	
	// Definir providers globais
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	
	// Configuração do propagador (incluindo o propagador personalizado INNOVABIZ)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		propagation.B3{},
		propagation.NewInnovabizContextPropagator(), // Propagador personalizado INNOVABIZ
	))
	
	// Retornar função de shutdown
	return func(ctx context.Context) error {
		log.Println("Shutting down OpenTelemetry...")
		
		// Desligar tracer provider
		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
		
		// Desligar meter provider
		if err := meterProvider.Shutdown(ctx); err != nil {
			return err
		}
		
		return nil
	}, nil
}
```