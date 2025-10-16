# Apache Kafka Integration with INNOVABIZ IAM Module

## Overview

This document describes the architecture for integrating Apache Kafka with the Identity and Access Management (IAM) module of the INNOVABIZ platform. Apache Kafka serves as the backbone for asynchronous communication between the various components of the IAM module, enabling scalability, resilience, and real-time observability.

## Event Architecture

The INNOVABIZ IAM module adopts an Event-Driven Architecture (EDA) to:

1. **Service Decoupling**: Allow microservices to operate independently
2. **Horizontal Scalability**: Facilitate scaling of individual components
3. **Resilience**: Ensure that failures in one component do not affect the entire system
4. **Complete Audit Trail**: Maintain an immutable record of all authentication and authorization operations
5. **Regional Adaptation**: Meet the specific requirements of each implementation region

### Kafka Infrastructure Components

The IAM module's Kafka infrastructure includes:

- **Kafka Brokers**: Responsible for storing and distributing events
- **Zookeeper**: Manages coordination between brokers
- **Schema Registry**: Ensures consistency of event schemas
- **Kafka Connect**: Facilitates integration with external systems
- **KSQLDB**: Enables real-time event stream processing
- **Kafka UI**: Management and monitoring interface

## Topics and Event Domains

Kafka topics are organized by functional domains, as described below:

### Authentication Domain

| Topic | Description | Partitions | Retention |
|-------|-------------|------------|-----------|
| `iam-auth-events` | Authentication events (login, logout) | 12 | 7 days |
| `iam-token-events` | Token lifecycle | 12 | 12 hours |
| `iam-mfa-challenges` | Multi-factor authentication challenges | 6 | 24 hours |
| `iam-risk-scores` | Risk scores for adaptive authentication | 6 | 7 days |

### Users and Tenants Domain

| Topic | Description | Partitions | Retention |
|-------|-------------|------------|-----------|
| `iam-user-events` | User operations | 6 | 14 days |
| `iam-tenant-events` | Tenant operations | 3 | 30 days |
| `iam-sessions` | Session management | 12 | 24 hours |

### Configuration and Management Domain

| Topic | Description | Partitions | Retention |
|-------|-------------|------------|-----------|
| `iam-method-updates` | Authentication method updates | 3 | Compacted |
| `iam-auth-configurations` | Authentication configurations | 3 | Compacted |

### Security and Compliance Domain

| Topic | Description | Partitions | Retention |
|-------|-------------|------------|-----------|
| `iam-security-alerts` | Security alerts | 6 | 30 days |
| `iam-audit-logs` | Audit logs | 12 | 90 days |
| `iam-user-deletion-requests` | Deletion requests (GDPR, LGPD) | 3 | 365 days |
| `iam-data-subject-requests` | Data subject requests | 3 | 365 days |
| `iam-consent-events` | Consent events | 6 | 730 days |

### Operations Domain

| Topic | Description | Partitions | Retention |
|-------|-------------|------------|-----------|
| `iam-notification-events` | Events for notifications | 6 | 3 days |
| `iam-healthchecks` | Health checks | 3 | 24 hours |
| `iam-deadletter` | Unprocessed messages | 3 | 90 days |

### Region/Sector-Specific Domains

| Topic | Description | Partitions | Retention | Active Regions |
|-------|-------------|------------|-----------|----------------|
| `iam-offline-auth-events` | Offline authentication | 6 | 90 days | AO |
| `iam-healthcare-auth-events` | Healthcare-specific events | 6 | 365 days | US, EU |

## Event Flows and Use Cases

### 1. Authentication Flow with MFA

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Initial   │────▶│    Risk     │────▶│     MFA     │────▶│     MFA     │
│    Login    │     │ Assessment  │     │  Challenge  │     │ Verification │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│iam-auth-eve.│────▶│iam-risk-sco.│────▶│iam-mfa-chal.│────▶│iam-auth-eve.│
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                                                    │
                                                                    ▼
                                                            ┌─────────────┐
                                                            │iam-token-ev.│
                                                            └─────────────┘
```

### 2. Audit and Compliance Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│     IAM     │────▶│ Aggregation │────▶│    GDPR     │
│    Events   │     │ and Normaliz│     │   Reports   │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │
       ▼                   ▼
┌─────────────┐     ┌─────────────┐
│iam-*-events │────▶│iam-audit-lo.│
└─────────────┘     └─────────────┘
```

### 3. Adaptive Authentication Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Initial   │────▶│  Contextual │────▶│   Adaptive  │────▶│Authentication│
│    Login    │     │   Analysis  │     │   Decision  │     │    Flow     │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│iam-auth-eve.│────▶│iam-risk-sco.│────▶│iam-auth-con.│────▶│iam-auth-eve.│
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
```

## Regional Adaptations

The IAM module's Kafka infrastructure is adapted to meet the specific requirements of each implementation region:

### European Union/Portugal (EU)

- **GDPR**: Enabling personal data masking in events
- **Retention**: Shorter retention policies for personal data
- **Security**: Enhanced encryption and strict schema validation
- **Auditing**: Extensive data access logs

### Brazil (BR)

- **LGPD**: Compliance with the General Data Protection Law
- **ICP-Brasil**: Specific topics for certificate validation
- **Retention**: Policies according to local regulatory requirements

### Angola (AO)

- **Intermittent Connectivity**: Support for offline authentication
- **PNDSB**: Compliance with local regulations
- **Extended Storage**: Longer retention time for events

### United States (US)

- **Sector-Specific**: Dedicated topics for healthcare (HIPAA) and finance
- **High Performance**: Greater number of partitions for increased parallelism
- **Extended Retention**: Compliance with audit requirements

## Security and Encryption

The security of the IAM Kafka infrastructure includes:

1. **SASL/SSL Authentication**: Ensures that only authorized clients can connect
2. **ACL Authorization**: Controls which clients can produce/consume from which topics
3. **TLS Encryption**: Protects data in transit
4. **Data Encryption**: Protects sensitive data in events
5. **PII Masking**: Does not expose personally identifiable information in events

## Observability and Monitoring

The IAM Kafka system is monitored through:

1. **JMX Metrics**: For broker and client performance
2. **Audit Logs**: For tracking sensitive activities
3. **Anomaly Alerts**: For detecting unexpected behaviors
4. **Operational Dashboards**: For real-time visualization
5. **OpenTelemetry Integration**: For distributed tracing

## Operational Procedures

### Topic Management

```bash
# Create a new topic
kafka-topics --bootstrap-server iam-kafka:9092 --create --topic iam-auth-events --partitions 12 --replication-factor 3

# List topics
kafka-topics --bootstrap-server iam-kafka:9092 --list

# Describe a topic
kafka-topics --bootstrap-server iam-kafka:9092 --describe --topic iam-auth-events
```

### Consumer Management

```bash
# List consumer groups
kafka-consumer-groups --bootstrap-server iam-kafka:9092 --list

# Describe a consumer group
kafka-consumer-groups --bootstrap-server iam-kafka:9092 --describe --group iam-auth-consumer
```

### Backup and Recovery

```bash
# Topic backup
kafka-mirror-maker --consumer.config consumer.properties --producer.config producer.properties --whitelist iam-audit-logs

# Topic restoration
kafka-console-consumer --bootstrap-server iam-kafka:9092 --topic backup.iam-audit-logs --from-beginning | \
kafka-console-producer --bootstrap-server iam-kafka:9092 --topic iam-audit-logs
```

## Integration with Other Systems

### GraphQL and REST APIs

Kafka events are integrated with the INNOVABIZ GraphQL gateway and REST APIs through:

1. **Kafka Connect**: For ingestion/exposure of data to external systems
2. **KSQLDB**: For real-time event transformation
3. **Schema Registry**: To ensure data compatibility

### MCP (Model Context Protocol)

The integration with MCP allows:

1. **Event Contextualization**: Enrichment of events with contextual information
2. **Intelligent Distribution**: Routing of events based on metadata
3. **Traceability**: Correlation of events through complete flows

## Best Practices

1. **Development**:
   - Use the official Kafka client libraries
   - Implement error handling and retries for producers
   - Clearly define processing semantics (at-least-once, exactly-once)

2. **Operation**:
   - Monitor consumer lag in real-time
   - Regularly back up critical topics
   - Keep broker and client versions updated

3. **Security**:
   - Review ACLs periodically
   - Monitor unauthorized access attempts
   - Update certificates before expiration

## Compliance and Governance

The Kafka implementation in the IAM module meets the following compliance requirements:

- **ISO/IEC 27001**: Information security
- **GDPR**: Data protection in the European Union
- **LGPD**: Data protection in Brazil
- **HIPAA**: For healthcare data in the US
- **PCI DSS**: For payment data
- **SOC 2**: Security and availability controls

## References and Additional Documentation

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Confluent Platform Documentation](https://docs.confluent.io/platform/current/overview.html)
- [Schema Registry Documentation](https://docs.confluent.io/platform/current/schema-registry/index.html)
- [KSQLDB Documentation](https://docs.ksqldb.io/)
- [Kafka Connect Documentation](https://docs.confluent.io/platform/current/connect/index.html)
