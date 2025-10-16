# Kafka Integration Guide for IAM Module

**Author:** INNOVABIZ Dev Team  
**Version:** 1.0.0  
**Date:** May/2025  
**Status:** Production  
**Classification:** Internal  

## Summary

1. [Introduction](#introduction)
2. [Event Architecture](#event-architecture)
3. [Topic Structure](#topic-structure)
4. [Producers and Consumers](#producers-and-consumers)
5. [Regional Adaptation](#regional-adaptation)
6. [MCP Integration](#mcp-integration)
7. [Healthcare Sector Integration](#healthcare-sector-integration)
8. [Compliance and Security](#compliance-and-security)
9. [Observability](#observability)
10. [Troubleshooting](#troubleshooting)
11. [References](#references)

## Introduction

This document describes how to integrate with the Apache Kafka event infrastructure of the IAM module in the INNOVABIZ platform. Apache Kafka is used as an asynchronous messaging middleware to ensure reliable, distributed, and scalable communication between the various components of the IAM module and other platform modules.

### Objectives

- Provide a comprehensive guide for developers who need to integrate with IAM events
- Explain the topic structure and data schemas used
- Detail the specific considerations for each implementation region
- Describe the available components for event production and consumption

### Prerequisites

- Basic knowledge of Apache Kafka and event processing
- Access to INNOVABIZ environment configurations
- Understanding of the authentication and authorization flows of the IAM module

## Event Architecture

The IAM module adopts an Event-Driven Architecture that promotes:

- **Decoupling**: Services can evolve independently
- **Scalability**: Components can be scaled as needed
- **Resilience**: Failures are isolated and not propagated
- **Auditability**: All operations are recorded as immutable events
- **Regional Adaptation**: Compliance with specific regulations

### Event Flow

![IAM Event Flow](/docs/iam/03-Desenvolvimento/kafka/docs/images/event-flow-diagram.png)

The typical IAM event flow follows these steps:

1. A producer service generates an event (e.g., login attempt)
2. The event is serialized following the Avro schema registered in the Schema Registry
3. The event is published to the appropriate Kafka topic
4. Multiple consumers process the event according to their needs
5. Events are adapted to the MCP protocol for integration with other modules
6. Audit events are generated as a byproduct of processing

## Topic Structure

IAM module Kafka topics are organized by functional domains:

### Authentication Domain

| Topic | Description | Retention | Partitions |
|-------|-------------|-----------|------------|
| `iam-auth-events` | Authentication events (login, logout) | 7 days | 12 |
| `iam-token-events` | Token lifecycle | 12 hours | 12 |
| `iam-mfa-challenges` | Multi-factor authentication challenges | 24 hours | 6 |

### User Domain

| Topic | Description | Retention | Partitions |
|-------|-------------|-----------|------------|
| `iam-user-events` | User operations | 14 days | 6 |
| `iam-sessions` | Session management | 24 hours | 12 |

### Security Domain

| Topic | Description | Retention | Partitions |
|-------|-------------|-----------|------------|
| `iam-security-alerts` | Security alerts | 30 days | 6 |
| `iam-audit-logs` | Audit logs | 90 days | 12 |

### Region/Sector Specific Domains

| Topic | Description | Applicable Regions | Partitions |
|-------|-------------|-------------------|------------|
| `iam-offline-auth-events` | Offline authentication | AO | 6 |
| `iam-healthcare-auth-events` | Healthcare-specific events | US, EU, BR | 6 |

## Producers and Consumers

The IAM module provides components to simplify integration with the Kafka infrastructure:

### Producers

The main class for event production is `AuthEventProducer`, which offers:

- Automatic connection management
- Avro serialization with Schema Registry
- Regional masking of sensitive data
- Transaction management
- Support for event batches

**Usage example:**

```javascript
const { AuthEventProducer } = require('iam/auth-framework/kafka/auth-event-producer');

// Create instance with regional configuration
const producer = new AuthEventProducer({
  regionCode: 'BR',  // Execution region
  schemaRegistryUrl: 'http://iam-schema-registry:8081'
});

// Connect to Kafka
await producer.connect();

// Publish authentication event
const result = await producer.publishAuthEvent({
  event_type: 'LOGIN_SUCCESS',
  tenant_id: 'acme-corp',
  user_id: '123e4567-e89b-12d3-a456-426614174000',
  method_code: 'K01',
  status: 'SUCCESS',
  timestamp: Date.now()
});

console.log(`Event published: ${result.event_id}`);

// Disconnect when finished
await producer.disconnect();
```

### Consumers

The main class for event consumption is `AuthEventConsumer`, which offers:

- Consumer group management
- Automatic Avro event deserialization
- Configurable event handlers
- Region-adapted processing
- DLQ (Dead Letter Queue) support

**Usage example:**

```javascript
const { AuthEventConsumer } = require('iam/auth-framework/kafka/auth-event-consumer');

// Create instance with regional configuration
const consumer = new AuthEventConsumer({
  regionCode: 'EU',
  groupId: 'my-service-consumer-group',
  eventHandlers: {
    // Custom handler for LOGIN_SUCCESS
    LOGIN_SUCCESS: async (event, headers) => {
      console.log(`Processing successful login: ${event.user_id}`);
      // Custom logic here
      return { processed: true, action: 'update-cache' };
    }
  }
});

// Connect and start consumption
await consumer.connect();
await consumer.subscribe(['iam-auth-events']);
await consumer.run();

// For proper shutdown
process.on('SIGTERM', async () => {
  await consumer.shutdown();
});
```

## Regional Adaptation

The IAM Kafka infrastructure is adapted to meet the regulatory specificities of each region:

### European Union (EU)

- Masking of personal data in events (GDPR)
- Strict consent verification
- Retention limitation for authentication data
- eIDAS authentication validation

### Brazil (BR)

- LGPD compliance
- Support for ICP-Brasil certificate validation
- Regional backup and retention policies

### Angola (AO)

- Support for offline authentication
- Less stringent data masking requirements
- Policies adapted to PNDSB (National Data Policy)

### United States (US)

- HIPAA validation for healthcare events
- SOC 2 and PCI DSS compliance
- Sector-specific policies (healthcare, finance)

## MCP Integration

The Model Context Protocol (MCP) is used to integrate Kafka events with other services in the INNOVABIZ platform.

### MCP Adapter

The `MCPKafkaAdapter` class facilitates conversion between Kafka events and MCP messages:

- Automatic mapping between Kafka topics and MCP channels
- Context enrichment
- Message attribute-based routing
- Event traceability

**Usage example:**

```javascript
const { MCPKafkaAdapter } = require('iam/auth-framework/kafka/mcp/mcp-kafka-adapter');

// Create MCP adapter
const mcpAdapter = new MCPKafkaAdapter({
  regionCode: 'EU',
  contextEnrichment: true
});

// Connect to MCP broker
await mcpAdapter.connect();

// Publish Kafka event as MCP message
await mcpAdapter.publishToMCP(
  'iam-auth-events',
  authEvent,
  { tenant: 'acme-corp' }
);

// Subscribe to an MCP channel and convert messages to Kafka events
await mcpAdapter.subscribeMCP('auth.events.eu', async (kafkaEvent) => {
  console.log(`MCP event received: ${kafkaEvent.event_id}`);
  // Process the event
});
```

## Healthcare Sector Integration

The INNOVABIZ platform offers specialized adapters for integration with the healthcare sector, considering specific regulatory requirements.

### Healthcare Events Adapter

The `HealthcareAuthEventAdapter` class provides:

- Specific compliance validation (HIPAA, GDPR Healthcare, LGPD Healthcare)
- Masking of PHI (Protected Health Information) data
- Integration with healthcare systems through MCP
- Compliance reporting generation

**Usage example:**

```javascript
const { HealthcareAuthEventAdapter } = 
  require('iam/auth-framework/kafka/healthcare/healthcare-auth-event-adapter');

// Create adapter for healthcare events
const healthcareAdapter = new HealthcareAuthEventAdapter({
  regionCode: 'US',
  enableComplianceValidation: true,
  enableObservability: true
});

// Publish healthcare-specific event
await healthcareAdapter.publishHealthcareAuthEvent({
  event_type: 'LOGIN_SUCCESS',
  tenant_id: 'central-hospital',
  user_id: '123e4567-e89b-12d3-a456-426614174000',
  method_code: 'K05',  // OTP as second factor (required for HIPAA)
  status: 'SUCCESS',
  additional_context: {
    phi_access: true,
    department: 'radiology'
  }
});
```

## Compliance and Security

The IAM Kafka infrastructure implements various security measures:

### Authentication and Authorization

- SASL/SSL for client authentication
- ACLs for topic access control
- Mutual TLS authentication

### Data Protection

- In-transit encryption (TLS 1.3)
- Masking of sensitive data
- Sanitization of personal data according to regulations

### Auditing

- Detailed recording of all operations
- Complete event traceability
- Possibility of historical state reconstruction

## Observability

The platform offers detailed metrics for monitoring:

### Kafka Metrics

- Publication/consumption latency
- Throughput rate per topic
- Consumer lag
- Production/consumption errors

### Domain Metrics

- Authentication event rate
- Distribution of authentication methods
- Success/failure rate
- MFA attempts

### OpenTelemetry Integration

All Kafka components integrate with the INNOVABIZ platform's OpenTelemetry observability infrastructure.

## Troubleshooting

### Common Problems and Solutions

| Problem | Probable Cause | Solution |
|---------|----------------|----------|
| `SchemaRegistryError` | Schema incompatibility | Check if the event follows the registered schema |
| `KafkaConnectionError` | Network or credential problems | Check connectivity and SASL configurations |
| `DeserializationError` | Data format problem | Validate data types and encoding |
| `ConsumerGroupRebalance` | Addition/removal of consumers | Normal operation, check if all instances recover |

### Logs and Diagnostics

Kafka components use the standard INNOVABIZ platform logger. To enable more detailed logs:

```javascript
// Configure log level for Kafka components
logger.setLevel('kafka', 'DEBUG');
```

## References

- [Complete Schema Registry Documentation](/docs/iam/04-Infraestrutura/kafka/schemas/README.md)
- [Regional Compliance Guide](/docs/iam/05-Seguranca/compliance/regional-compliance-guide.md)
- [Kafka Capacity Planning](/docs/iam/04-Infraestrutura/kafka/capacity-planning.md)
- [MCP Integration Guide](/docs/iam/03-Desenvolvimento/mcp/integration-guide.md)
- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
