# ADR-001: Adoção de GraphQL como API Principal para o Módulo IAM

**Status:** Aprovado  
**Data:** 2025-08-06  
**Autor:** Equipa INNOVABIZ DevSecOps  
**Stakeholders:** Arquitetura, Desenvolvimento, Segurança, Compliance  

## Contexto

O módulo IAM da plataforma INNOVABIZ precisa fornecer APIs robustas, flexíveis e de alto desempenho para autenticação, autorização e gestão de identidades que serão consumidas por múltiplos módulos internos e aplicações externas. A escolha da tecnologia de API é crucial para garantir a flexibilidade, escalabilidade e segurança em linha com a arquitetura multi-dimensional e multi-tenant da plataforma.

## Opções Consideradas

1. **REST API Tradicional**
   * Tecnologia estabelecida e amplamente compreendida
   * Simplicidade de implementação
   * Bom suporte em ferramentas de documentação (Swagger/OpenAPI)
   * Limitações em termos de overfetching e underfetching

2. **GraphQL como API Principal**
   * Consultas flexíveis e específicas
   * Tipagem forte e esquema bem definido
   * Federação para distribuição entre serviços
   * Introspection para autodocumentação
   * Maior complexidade inicial de implementação

3. **gRPC como API Principal**
   * Alto desempenho com Protocol Buffers
   * Bom para comunicação entre serviços internos
   * Suporte limitado para clientes web
   * Menos flexibilidade para clientes externos

4. **Abordagem Híbrida (GraphQL + REST)**
   * GraphQL para operações principais
   * REST para compatibilidade legada e operações simples
   * Maior complexidade de manutenção

## Decisão

**Adotar GraphQL como API principal para o módulo IAM, com APIs REST auxiliares para casos específicos.**

Esta decisão se baseia nos seguintes fatores:

1. **Alinhamento com Multi-dimensionalidade**: O GraphQL permite modelar e consultar estruturas complexas como hierarquias de grupos, permissões em múltiplos níveis e contextos variáveis de autenticação de forma natural.

2. **Eficiência de Rede**: Reduz significativamente o overfetching e underfetching, permitindo que os clientes solicitem exatamente os dados necessários, crucial em ambientes mobile com banda limitada nos mercados-alvo como Angola e outros PALOP.

3. **Tipagem Forte**: O esquema GraphQL fornece tipagem forte e documentação integrada, alinhando-se aos requisitos de qualidade e governança da plataforma INNOVABIZ.

4. **Federation**: Permite a distribuição do esquema GraphQL entre múltiplos serviços, possibilitando uma arquitetura mais modular e escalável.

5. **Flexibilidade para Integrações**: Facilita a integração com outros módulos core (Payment Gateway, Mobile Money, Marketplace) permitindo que cada cliente consuma exatamente os campos necessários.

6. **Redução de Versões de API**: Minimiza a necessidade de versionamento explícito da API, permitindo evolução gradual do esquema.

7. **Observabilidade**: Oferece métricas detalhadas sobre cada campo e tipo, permitindo otimizações focadas em pontos de gargalo reais.

## Consequências

### Positivas

* Maior flexibilidade para clientes com diferentes requisitos de dados
* Redução do volume de dados transferidos em redes com limitação de banda (importante para mercados africanos)
* Autodocumentação e tipagem forte melhoram a experiência do desenvolvedor
* Suporte nativo para subscrições em tempo real (importante para notificações de segurança)
* Facilidade para federação entre múltiplos serviços IAM distribuídos geograficamente

### Negativas

* Curva de aprendizado inicial para equipes mais acostumadas com REST
* Necessidade de controles adicionais para limitar complexidade de queries (prevenção de DoS)
* Desafios adicionais de cache comparado com REST
* Necessidade de ferramentas específicas para monitoramento e teste

### Neutras

* Manutenção de algumas APIs REST para compatibilidade com sistemas legados e operações simples
* Necessidade de revisão mais rigorosa de segurança nas queries GraphQL

## Conformidade e Governança

Esta decisão está em conformidade com:

* **ISO/IEC 25010:2011** (Qualidade de Software) - Melhora a manutenibilidade e eficiência
* **TOGAF 10** - Alinha-se com os princípios de arquitetura modular e flexível
* **OWASP ASVS 4.0** - Requer implementação de controles específicos para GraphQL
* **PCI-DSS v4.0** - Compatível com exigências de segurança em ambientes de pagamento

## Alternativas Não Selecionadas

* **REST puro** foi rejeitado devido às limitações para modelar relações complexas de identidade e alto overhead de rede.
* **gRPC puro** foi rejeitado pela limitação de suporte em clients web e menor flexibilidade para evolução de API.

## Implementação

A implementação será realizada usando:

* **gqlgen** para Go - Geração de código typesafe e eficiente
* **Apollo Server** para Node.js - Serviços GraphQL com suporte a federação
* **Apollo Federation** - Para distribuição do esquema entre serviços

## Verificação

O sucesso desta decisão será avaliado por:

* Redução de 40% no volume de dados transferidos em comparação com REST
* Redução de 30% no tempo de desenvolvimento de novas integrações
* Satisfação de desenvolvedores >85% em pesquisas internas
* Zero incidentes de segurança relacionados à implementação GraphQL em 12 meses

## Referências

* [GraphQL Best Practices - Apollo](https://www.apollographql.com/docs/apollo-server/security/overview/)
* [Securing GraphQL - OWASP](https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html)
* [GraphQL Federation - Apollo](https://www.apollographql.com/docs/federation/)