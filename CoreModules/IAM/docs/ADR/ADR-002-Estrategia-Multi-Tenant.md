# ADR-002: Estratégia de Multi-Tenancy para o Módulo IAM

**Status:** Aprovado  
**Data:** 2025-08-06  
**Autor:** Equipa INNOVABIZ DevSecOps  
**Stakeholders:** Arquitetura, Desenvolvimento, DBA, Segurança, Compliance, Operações  

## Contexto

A plataforma INNOVABIZ foi projetada como uma solução SaaS multi-tenant para atender diversos mercados e setores, com foco particular em Angola, países da CPLP, SADC, PALOP e expansão para mercados globais. O módulo IAM é o componente central para gerenciar identidades e acessos em toda a plataforma, devendo suportar isolamento completo entre tenants enquanto mantém a escalabilidade, performance e conformidade regulatória em múltiplos contextos geográficos e regulatórios.

## Opções Consideradas

1. **Database por Tenant**
   * Isolamento completo no nível de banco de dados
   * Máxima segurança e conformidade
   * Personalização por tenant facilitada
   * Limitações de escalabilidade para grande número de tenants
   * Alto custo operacional

2. **Schema por Tenant**
   * Bom isolamento lógico
   * Esquemas separados dentro do mesmo banco de dados
   * Melhor relação entre isolamento e utilização de recursos
   * Limitações em número de schemas para alguns SGBDs

3. **Tenant ID como Discriminador**
   * Todos os tenants compartilham tabelas
   * Filtros em nível de aplicação para separação de dados
   * Alta densidade de tenants
   * Riscos maiores de vazamento de dados entre tenants

4. **Abordagem Híbrida**
   * Combinação de estratégias baseada em classificação de tenants
   * Tenants Premium com isolamento completo
   * Tenants Standard com compartilhamento de recursos

5. **Multi-tenant com Sharding Geográfico**
   * Combinação de multi-tenant com distribuição geográfica
   * Alinhamento com regulamentações regionais de dados
   * Otimização de latência por proximidade

## Decisão

**Adotar uma estratégia de Multi-Tenancy baseada em Schema por Tenant com Sharding Geográfico.**

Esta decisão se baseia nos seguintes fatores:

1. **Isolamento com Eficiência**: O modelo de schema por tenant oferece isolamento lógico suficiente para garantir segurança e conformidade, mantendo eficiência operacional superior à abordagem de database por tenant.

2. **Compliance Regional**: O sharding geográfico permite alocar tenants em regiões específicas conforme requisitos regulatórios (ex: dados de Angola permanecendo em data centers africanos, dados europeus conforme GDPR).

3. **Escalabilidade**: O PostgreSQL (escolhido como banco primário) suporta eficientemente milhares de schemas, permitindo crescimento significativo antes de limitações técnicas.

4. **Gestão de Recursos**: Melhor balanceamento de recursos compartilhados com capacidade de alocar recursos dedicados para tenants com maior demanda.

5. **Migração Facilitada**: Possibilidade de migrar tenants específicos para databases dedicados no futuro caso necessário, sem mudanças arquiteturais significativas.

6. **Soberania de Dados**: Atendimento aos requisitos de soberania de dados de diferentes jurisdições através do sharding geográfico.

## Implementação Técnica

1. **Schema Naming**: Schemas nomeados com formato padronizado `tenant_{tenant_id}` para facilitar gerenciamento e automação.

2. **Connection Pooling Otimizado**: Implementação de connection pooling por tenant para evitar contaminação entre conexões.

3. **Metadata Central**: Catálogo central de tenants e suas configurações em schema separado (`tenant_catalog`).

4. **Sharding Geográfico**:
   * África (Primary: Luanda, Secondary: Joanesburgo)
   * Europa (Primary: Lisboa, Secondary: Frankfurt)
   * América do Sul (Primary: São Paulo, Secondary: Rio de Janeiro)
   * América do Norte (Primary: Nova York, Secondary: São Francisco)
   * Ásia (Primary: Singapura, Secondary: Pequim)

5. **Migração e Rebalanceamento**: Procedimentos automatizados para migração de tenants entre regiões conforme necessário.

## Consequências

### Positivas

* Isolamento adequado para requisitos regulatórios e de segurança
* Utilização eficiente de recursos computacionais
* Capacidade de escalar para milhares de tenants
* Flexibilidade para atender requisitos específicos de localização de dados
* Separação clara dos dados para auditorias e conformidade regulatória

### Negativas

* Complexidade adicional na gestão de conexões de banco de dados
* Necessidade de ferramentas específicas para administração de múltiplos schemas
* Desafios em operações de backup/restore granulares
* Overhead de gerenciamento do catálogo central de tenants

### Mitigações

* Desenvolvimento de ferramentas automatizadas para gestão de schemas e migração
* Implementação de políticas de conexão com strict reset
* Monitoramento avançado para detecção de vazamentos entre tenants
* Framework centralizado para gestão de schemas e catálogo

## Conformidade e Governança

Esta estratégia de multi-tenancy está em conformidade com:

* **GDPR/LGPD**: Permite isolamento e localização de dados conforme exigido
* **BNA Instrução 7/2021**: Atende requisitos de armazenamento local para dados financeiros angolanos
* **SOX**: Suporta requisitos de auditoria e separação de dados
* **ISO/IEC 27001**: Implementa controles de segregação de ambientes
* **PCI DSS**: Permite isolamento adequado para dados relacionados a pagamentos

## Métricas e Monitoramento

Para validar o sucesso desta estratégia, serão monitoradas as seguintes métricas:

* Overhead de memória por tenant adicional (<5%)
* Tempo de criação/provisionamento de novo tenant (<30s)
* Impacto de performance entre tenants (variação <5%)
* Tempo médio para migração de tenant entre regiões (<2h)
* Incidentes de vazamento de dados entre tenants (meta: zero)

## Verificação

A efetividade desta decisão será verificada através de:

* Testes de isolamento automáticos em CI/CD
* Auditorias de segurança trimestrais
* Testes de performance com simulação de carga multi-tenant
* Validação de conformidade regulatória por região

## Referências

* [PostgreSQL Schema Documentation](https://www.postgresql.org/docs/current/ddl-schemas.html)
* [Multi-tenant Data Architecture - NIST SP 800-145](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-145.pdf)
* [Data Residency and Sovereignty Requirements - ISO/IEC 27701](https://www.iso.org/standard/71670.html)
* [Tenant Isolation Patterns in SaaS Applications - AWS Architecture](https://docs.aws.amazon.com/prescriptive-guidance/latest/saas-tenancy-patterns/)