# Estratégia de Segurança Multi-Tenant para IAM INNOVABIZ

## Visão Geral

Este documento define a estratégia de segurança multi-tenant para o módulo IAM (Identity and Access Management) da plataforma INNOVABIZ, garantindo o isolamento rigoroso, proteção de dados, controles de acesso e conformidade regulatória em ambientes compartilhados. A abordagem implementa conceitos avançados de segurança por design, alinhados com frameworks TOGAF 10.0, COBIT 2019, ISO/IEC 27001:2022, NIST SP 800-53 e requisitos específicos para instituições financeiras.

## Objetivos

1. **Isolamento Total**: Garantir separação completa entre tenants em todos os níveis de arquitetura
2. **Proteção de Dados**: Implementar controles de segurança para dados em repouso, em uso e em trânsito
3. **Escalabilidade Segura**: Permitir crescimento sem comprometer a segurança ou performance
4. **Compliance Multinível**: Atender requisitos regulatórios de múltiplos países/jurisdições
5. **Defesa em Profundidade**: Implementar múltiplas camadas de segurança para proteção abrangente
6. **Auditabilidade Granular**: Permitir rastreamento e verificação de todas as atividades por tenant
7. **Segurança por Design**: Incorporar controles de segurança desde os estágios iniciais de desenvolvimento

## Arquitetura Multi-Tenant

### Modelos de Isolamento

O IAM INNOVABIZ implementa um modelo híbrido de multi-tenancy com diferentes níveis de isolamento conforme requisitos de segurança, conformidade e performance:

| Nível | Modelo | Implementação | Casos de Uso |
|-------|--------|---------------|-------------|
| Dados | Pool Segregado | Esquema por tenant, criptografia por tenant | Dados sensíveis, tenants com requisitos regulatórios específicos |
| Dados | Pool Compartilhado | Discriminador de tenant com RBAC/ABAC | Dados comuns, configurações, metadados |
| Aplicação | Instâncias Dedicadas | Pods/containers específicos por tenant | Tenants premium, requisitos de compliance estritos |
| Aplicação | Instâncias Compartilhadas | Filtros de contexto em runtime | Tenants padrão, máxima eficiência de recursos |
| Rede | Segmentação Virtual | NSGs, Network Policies, Service Mesh | Todos os tenants |

### Hierarquia de Tenants

O sistema suporta uma estrutura hierárquica de tenants para modelar relações organizacionais complexas:

1. **Root Tenant**: Controle global da plataforma (INNOVABIZ)
2. **Master Tenants**: Organizações principais (ex: instituições financeiras, holdings)
3. **Sub-Tenants**: Divisões, subsidiárias, departamentos
4. **Micro-Tenants**: Unidades especializadas, projetos temporários

Cada nível herda políticas de segurança do nível superior, podendo adicionar restrições, mas nunca relaxá-las.

## Estratégias de Segurança por Camada

### 1. Camada de Dados

| Controle | Implementação | Padrões/Frameworks |
|----------|---------------|-------------------|
| Isolamento de Dados | Discriminador de tenant em todas as consultas | OWASP ASVS 4.0 - 4.1 |
| Criptografia em Repouso | AES-256 com chaves específicas por tenant | NIST FIPS 140-2, ISO 27001 A.10 |
| Criptografia em Trânsito | TLS 1.3, mTLS entre serviços | NIST SP 800-52r2 |
| Mascaramento de Dados | Tokenização de PII por contexto de tenant | GDPR Art. 32, ISO 27018 |
| Backup Segregado | Separação lógica de backups por tenant | ISO 27001 A.12.3 |
| Ciclo de Vida de Dados | Políticas de retenção/exclusão específicas por tenant | GDPR Art. 17, ISO 27701 |
| Prevenção de Vazamento | DLP com políticas por tenant | ISO 27001 A.8.2 |

#### Implementação PostgreSQL Multi-Tenant

```sql
-- Exemplo de função para impor isolamento de tenant a nível de banco
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
    IF NULLIF(current_setting('app.current_tenant_id', TRUE), '') IS NULL THEN
        RAISE EXCEPTION 'tenant_id must be set';
    END IF;
    
    NEW.tenant_id = current_setting('app.current_tenant_id');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicação em tabela de usuários
CREATE TRIGGER enforce_tenant_isolation_users
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE enforce_tenant_isolation();

-- Função para filtrar consultas por tenant atual
CREATE OR REPLACE FUNCTION tenant_isolation_policy(tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_id::text, false);
    -- Aplicar RLS (Row Level Security)
    -- Configurar search_path para esquema específico do tenant
END;
$$ LANGUAGE plpgsql;
```

### 2. Camada de Aplicação

| Controle | Implementação | Padrões/Frameworks |
|----------|---------------|-------------------|
| Contexto de Tenant | Propagação obrigatória em todos endpoints | TOGAF Security Architecture |
| Validação de Tenant | Verificação em toda requisição/operação | OWASP ASVS 4.0 - 4.1 |
| Cache Isolado | Particionamento de cache por tenant | SANS CWE Top 25 |
| Autorização Contextual | RBAC+ABAC com políticas por tenant | NIST SP 800-162 |
| Rate Limiting | Quotas e limites específicos por tenant | OWASP ASVS 4.0 - 4.13 |
| Sanitização de Entrada | Validação específica por tenant e contexto | OWASP ASVS 4.0 - 5.1 |
| Gestão de Sessão | Isolamento de sessões por tenant | OWASP ASVS 4.0 - 3.1 |

#### Implementação de Middleware Multi-Tenant

```go
// Middleware Go para validação de tenant
func TenantValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extrair tenant_id do token JWT ou cabeçalho
        tenantID := extractTenantID(c)
        
        // Validar se tenant existe e está ativo
        if !isValidTenant(tenantID) {
            c.AbortWithStatusJSON(403, gin.H{"error": "Tenant inválido ou inativo"})
            auditLog.Warning("Tentativa de acesso com tenant inválido", 
                map[string]interface{}{"tenant_id": tenantID, "ip": c.ClientIP()})
            return
        }
        
        // Configurar contexto com tenant_id
        ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
        c.Request = c.Request.WithContext(ctx)
        
        // Configurar conexão com banco de dados para isolamento de tenant
        db.SetTenantContext(tenantID)
        
        // Continuar processamento
        c.Next()
        
        // Limpar contexto após requisição
        db.ClearTenantContext()
    }
}
```

### 3. Camada de API e GraphQL

| Controle | Implementação | Padrões/Frameworks |
|----------|---------------|-------------------|
| Diretivas de Segurança | Validação de permissões por tenant em resolvers | GraphQL Security, OWASP API Security |
| Query Complexity | Limites específicos por tenant e tipo de usuário | GraphQL Security |
| Field-Level Security | Controle de acesso granular a campos específicos | NIST ABAC |
| Filtering Automático | Filtragem implícita por tenant em todos resultados | OWASP ASVS 4.0 |
| Validação de Entrada | Validadores específicos por operação e tenant | ISO 27001 A.14.2 |
| Normalização de Erros | Respostas consistentes sem vazamento de informação | OWASP API Security |
| Introspection Limitada | Restrição com base em permissões e tenant | GraphQL Security |

#### Implementação de Resolver Multi-Tenant

```go
// Exemplo de resolver GraphQL com segurança multi-tenant
func (r *queryResolver) Users(ctx context.Context, filter UserFilter) ([]*User, error) {
    // Extrair tenant_id do contexto de autenticação
    tenantID, err := auth.TenantFromContext(ctx)
    if err != nil {
        return nil, gqlerror.Forbidden("Acesso não autorizado")
    }
    
    // Iniciar trace com informações de tenant para observabilidade
    span := tracer.StartSpan("users.list", 
        trace.WithAttributes(attribute.String("tenant.id", tenantID.String())))
    defer span.End()
    
    // Aplicar filtro de tenant automaticamente
    filter.TenantID = tenantID
    
    // Verificar permissão específica no contexto do tenant atual
    if !auth.HasPermission(ctx, "users:list", tenantID) {
        auditLog.SecurityEvent("permission_denied", 
            map[string]interface{}{"tenant_id": tenantID, "operation": "users:list"})
        return nil, gqlerror.Forbidden("Permissão insuficiente")
    }
    
    // Executar query com contexto de tenant
    users, err := r.UserService.ListUsers(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // Log de auditoria da operação
    auditLog.Info("Consulta de usuários", 
        map[string]interface{}{"tenant_id": tenantID, "count": len(users)})
    
    return users, nil
}
```

### 4. Camada de Rede e Infraestrutura

| Controle | Implementação | Padrões/Frameworks |
|----------|---------------|-------------------|
| Segmentação de Rede | NSGs, Network Policies por tenant | NIST SP 800-125B |
| mTLS | Autenticação mútua entre serviços | NIST SP 800-52r2 |
| Isolamento de Pods | Anti-affinity, namespaces por tenant (premium) | Kubernetes Security |
| WAF Contextual | Regras específicas por tenant | OWASP Top 10 |
| API Gateway | Roteamento e proteção específica por tenant | NIST Zero Trust |
| DoS Protection | Mitigação por tenant | ISO 27001 A.17 |
| Secrets Management | Isolamento de segredos por tenant | NIST SP 800-57 |

#### Implementação Kubernetes Multi-Tenant

```yaml
# Exemplo de Network Policy para isolamento de tenant
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tenant-isolation-policy
  namespace: iam-system
spec:
  podSelector:
    matchLabels:
      tenant-id: "tenant-12345"
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          tenant-id: "tenant-12345"
    - namespaceSelector:
        matchLabels:
          tenant-isolation: "shared-services"
  egress:
  - to:
    - podSelector:
        matchLabels:
          tenant-id: "tenant-12345"
    - namespaceSelector:
        matchLabels:
          tenant-isolation: "shared-services"
```

## Gestão de Identidade e Acesso Multi-Tenant

### 1. Modelo de Autorização Multi-contexto

O IAM implementa um modelo híbrido combinando:

- **RBAC**: Papéis padrão por tenant (Admin, Operador, Usuário)
- **ABAC**: Atributos dinâmicos (localização, dispositivo, risco)
- **ReBAC**: Relacionamentos entre entidades (gestor, subordinado)
- **PBAC**: Políticas baseadas em tenant, contexto e conformidade

#### Estrutura de Políticas

```json
{
  "policy_id": "fin-ops-policy-123",
  "tenant_id": "tenant-12345",
  "version": "1.2.0",
  "effect": "allow",
  "subjects": ["role:finance-manager", "group:payment-admins"],
  "resources": ["payment:transaction", "payment:batch"],
  "actions": ["approve", "reject", "view"],
  "conditions": {
    "transaction.amount": {
      "lte": 50000
    },
    "user.risk_score": {
      "gte": 80
    },
    "request.time": {
      "between": ["08:00:00Z", "18:00:00Z"]
    },
    "user.location": {
      "in": ["BR", "PT", "AO"]
    },
    "context.compliance": {
      "requires": ["2FA", "ip_whitelist"]
    }
  }
}
```

### 2. Federação e SSO Multi-tenant

| Funcionalidade | Implementação | Benefício de Segurança |
|----------------|---------------|------------------------|
| IdP Customizado | SAML/OIDC por tenant | Conformidade com políticas corporativas |
| Mapeamento de Claims | Transformação específica por tenant | Normalização de atributos |
| Federação em Cascata | Delegação controlada de autenticação | Suporte a hierarquias complexas |
| Perímetro de Federação | Limitação de provedores externos por tenant | Redução de superfície de ataque |
| Políticas de Sessão | Configurações específicas por tenant e contexto | Controle granular de acesso |

### 3. Autenticação Multi-fator Adaptativa

O sistema implementa MFA adaptativo baseado em:

- Perfil de risco do tenant
- Classificação de dados/operação
- Contexto de acesso (localização, dispositivo, horário)
- Requisitos regulatórios aplicáveis

| Nível de Risco | Fatores Mínimos | Exemplos de Fatores |
|----------------|----------------|---------------------|
| Baixo | 1 | Senha |
| Médio | 2 | Senha + OTP |
| Alto | 3 | Senha + Biometria + Smart Card |
| Crítico | 3+ | Biometria + Smart Card + Aprovação adicional |

## Proteção de Dados Multi-tenant

### 1. Criptografia Hierárquica de Chaves

O sistema implementa um modelo hierárquico de gerenciamento de chaves:

1. **Master Keys**: Chave raiz para toda plataforma (HSM)
2. **Tenant Keys**: Derivadas da Master Key para cada tenant
3. **Data Category Keys**: Derivadas por categoria de dados dentro do tenant
4. **Data Keys**: Chaves específicas para conjuntos de dados

```
Master Key (HSM)
├── Tenant A Key (KMS)
│   ├── PII Key
│   ├── Payment Key
│   └── Audit Log Key
├── Tenant B Key (KMS)
│   ├── PII Key
│   ├── Payment Key
│   └── Audit Log Key
```

### 2. Tokenização e Pseudonimização

| Tipo de Dado | Método | Reversibilidade | Uso |
|--------------|--------|----------------|-----|
| PII Crítico | Tokenização Format-Preserving | Reversível com autorização | Dados KYC |
| PII Secundário | Pseudonimização | Semi-reversível | Analytics, BI |
| Dados Financeiros | Tokenização PCI-DSS | Reversível com controles estritos | Pagamentos |
| Dados Comportamentais | Hashing com sal por tenant | Irreversível | Análise de segurança |

### 3. Controle de Acesso a Dados

| Mecanismo | Implementação | Aplicação |
|-----------|---------------|-----------|
| Visibilidade de Campos | Mascaramento dinâmico baseado em contexto | Dados sensíveis em UI/relatórios |
| Row-Level Security | Filtros automáticos por tenant | Todos os acessos a banco |
| Column Encryption | Criptografia por coluna com chaves por tenant | Dados altamente sensíveis |
| Data Loss Prevention | Escaneamento e bloqueio de exposição | Exportações, relatórios |

## Auditoria e Monitoramento Multi-tenant

### 1. Logging Segregado

- Logs físicos separados por tenant para máxima isolação
- Retenção configurável por tenant conforme requisitos regulatórios
- Assinatura digital de logs para garantia de integridade
- Pipeline de processamento isolado por tenant

### 2. Detecção de Anomalias Contextual

O sistema implementa detecção de anomalias com contexto de tenant:

- Baselines específicas por tenant para comportamento normal
- Modelos de ML treinados com dados de tenant específico
- Thresholds adaptativos baseados em padrões de uso por tenant
- Correlação de eventos cross-tenant para detecção de ataques coordenados

### 3. Estrutura de Evento de Segurança

```json
{
  "event_id": "evt-1234567890",
  "timestamp": "2025-08-06T15:04:05.123Z",
  "tenant_id": "tenant-12345",
  "tenant_hierarchy": ["tenant-root", "tenant-parent", "tenant-12345"],
  "event_type": "security.auth.failed",
  "severity": "high",
  "source": {
    "service": "identity-service",
    "instance": "pod-iam-5678",
    "ip": "10.0.12.24",
    "component": "AuthenticationService"
  },
  "subject": {
    "user_id": "user-98765",
    "username": "joao.silva",
    "roles": ["finance-manager"],
    "ip": "192.168.1.50",
    "location": "BR-SP",
    "device_id": "dev-abc123"
  },
  "action": {
    "operation": "login",
    "resource_type": "user-account",
    "resource_id": "user-98765",
    "status": "failed"
  },
  "context": {
    "auth_method": "password",
    "failure_reason": "invalid_credentials",
    "attempt_count": 3,
    "risk_score": 75,
    "previous_success": "2025-08-05T10:24:15Z"
  },
  "compliance": {
    "pci_dss": ["10.2.4", "10.2.5"],
    "gdpr": ["Art. 32"],
    "iso27001": ["A.9.4.2"]
  },
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736"
}
```

## Resposta a Incidentes Multi-tenant

### 1. Classificação e Priorização

| Fator | Impacto na Priorização | Exemplo |
|-------|------------------------|---------|
| Classificação de dados | Crítico, Alto, Médio, Baixo | Dados financeiros > Dados analíticos |
| Impacto de tenant | Multiplicador de prioridade | Tenant financeiro > Tenant de teste |
| Cross-tenant | Aumento de severidade | Ataque que afeta múltiplos tenants |
| Regulamentação aplicável | Requisitos de notificação | PCI-DSS, GDPR, LGPD |

### 2. Playbooks Específicos por Tenant

- Templates base de resposta a incidentes
- Customizações específicas por tenant (contatos, procedimentos)
- Integração com canais de comunicação do tenant
- SLAs diferenciados por contrato/tenant

### 3. Isolamento e Contenção

| Técnica | Implementação | Escopo |
|---------|---------------|--------|
| Circuit Breaking | Degradação controlada de serviço | Tenant específico ou função |
| Quarentena de Tenant | Isolamento de network/dados | Tenant comprometido |
| Snapshot Forense | Captura de estado para investigação | Recursos afetados |
| Rollback Isolado | Restauração para estado conhecido | Tenant específico sem afetar outros |

## Gestão de Compliance Multi-tenant

### 1. Matriz de Requisitos Regulatórios

| Tenant | Jurisdições | Regulamentações | Controles Específicos |
|--------|-------------|-----------------|------------------------|
| Banco XYZ | BR, US | LGPD, GDPR, PCI-DSS, SOX | Criptografia forte, auditoria extendida |
| Seguradora ABC | AO, PT | PNDSB, GDPR | Classificação de dados de saúde |
| Fintech 123 | BR, MZ | LGPD, Regulamentos locais | E-KYC, prevenção de fraude |

### 2. Controles Dinâmicos

- Ativação automática de controles com base em metadados de tenant
- Verificação contínua de conformidade com dashboard por tenant
- Workflows de remediação com base em gaps identificados
- Atualização automatizada de controles quando regulamentações mudam

### 3. Relatórios de Compliance

- Geração automatizada de relatórios por tenant e regulamentação
- Evidências criptograficamente verificáveis de controles
- Dashboard de postura de segurança e compliance por tenant
- Histórico de conformidade para auditoria

## DevSecOps Multi-tenant

### 1. CI/CD Seguro

| Fase | Controles Multi-tenant | Ferramentas |
|------|------------------------|-------------|
| Código | SAST com regras específicas por tenant | SonarQube, Checkmarx |
| Build | Verificação de dependências por requisitos de tenant | OWASP Dependency Check |
| Teste | Casos de teste específicos para isolamento de tenant | Testify, Ginkgo |
| Deploy | Validação de configuração de segurança por tenant | Terraform Sentinel, OPA |
| Runtime | Scanning contínuo com contexto de tenant | Falco, Prisma Cloud |

### 2. IaC com Segurança por Tenant

```terraform
# Exemplo de configuração Terraform com segurança multi-tenant
module "tenant_namespace" {
  source = "./modules/tenant_namespace"
  
  for_each = var.tenants
  
  name                     = "tenant-${each.key}"
  tenant_id                = each.key
  security_level           = each.value.security_level
  compliance_requirements  = each.value.compliance_requirements
  network_isolation        = each.value.network_isolation
  enable_pod_security      = true
  encryption_key_rotation  = each.value.encryption_key_rotation
  
  resource_quotas = {
    cpu     = each.value.quota.cpu
    memory  = each.value.quota.memory
    pods    = each.value.quota.pods
    secrets = each.value.quota.secrets
  }
}
```

## Plano de Implementação

### Fase 1: Fundação (M1-M3)

1. Implementação de isolamento básico de tenant em todos os serviços
2. Contexto de tenant obrigatório em APIs e middlewares
3. Validação de tenant em todos os endpoints
4. Logging básico com separação por tenant

### Fase 2: Segurança Avançada (M4-M6)

1. Criptografia hierárquica de chaves por tenant
2. RBAC+ABAC com políticas específicas por tenant
3. Instrumentação de auditoria avançada
4. Federação e MFA adaptativo

### Fase 3: Compliance e Automação (M7-M9)

1. Matriz de requisitos regulatórios automatizada
2. Testes de penetração específicos por tenant
3. Dashboards de compliance por tenant
4. Detecção de anomalias contextual

### Fase 4: Otimização e Maturidade (M10-M12)

1. ML para análise de comportamento por tenant
2. Automação avançada de resposta a incidentes
3. Self-service de configuração de segurança por tenant
4. Verificação contínua de postura de segurança

## Métricas de Segurança Multi-tenant

| Métrica | Baseline | Meta 6 meses | Meta 12 meses |
|---------|----------|--------------|--------------|
| Índice de isolamento de tenant | 85% | 95% | 99.5% |
| Vazamentos cross-tenant | <5 por mês | <1 por mês | Zero |
| Cobertura de controles por tenant | 75% | 90% | 98% |
| Tempo médio de detecção (MTTD) | 120 min | 30 min | 5 min |
| Conformidade regulatória | 80% | 95% | 100% |
| Cobertura de testes de penetração | 60% | 85% | 100% |

## Referências

1. NIST SP 800-53 Rev. 5 - Security and Privacy Controls
2. ISO/IEC 27001:2022 - Information Security Management Systems
3. OWASP ASVS 4.0 - Application Security Verification Standard
4. CSA CCM 4.0 - Cloud Controls Matrix
5. CIS Controls v8 - Critical Security Controls
6. NIST Cybersecurity Framework v1.1
7. OWASP API Security Top 10 2023
8. GDPR, LGPD, PNDSB (Angola) - Regulamentos de Proteção de Dados
9. PCI DSS v4.0 - Payment Card Industry Data Security Standard
10. FAPI - Financial-grade API Security Profile

---

*Este documento está em conformidade com os padrões de documentação técnica da INNOVABIZ e deve ser revisado e atualizado regularmente conforme a evolução do sistema.*

*Última atualização: 06/08/2025*