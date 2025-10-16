# Documentação Técnica da API REST do Módulo IAM

## Visão Geral

A API REST do módulo IAM (Identity and Access Management) da INNOVABIZ fornece uma interface moderna e segura para gerenciamento de identidades, autenticação, autorização e conformidade regulatória, com foco especial em compatibilidade com regulamentações de saúde (HIPAA, GDPR, LGPD e PNDSB) e suporte para métodos avançados de autenticação, incluindo AR/VR.

## Arquitetura

A API segue uma arquitetura em camadas:

1. **Camada de Apresentação**: Endpoints REST com validação de entrada
2. **Camada de Serviços**: Lógica de negócios em Python
3. **Camada de Dados**: Funções SQL no banco de dados PostgreSQL
4. **Camada de Segurança**: Middleware para multi-tenancy e auditoria

### Tecnologias Utilizadas

- **Framework**: FastAPI 0.95+
- **Banco de Dados**: PostgreSQL 14+
- **Autenticação**: JWT com suporte a multi-fator
- **Documentação**: OpenAPI 3.0

## Componentes Principais

### 1. Sistema de Autenticação

Suporta diversos fluxos de autenticação:

- Autenticação por usuário/senha com suporte a multi-tenant
- Multi-fator (MFA) com diversos métodos:
  - TOTP (Google Authenticator, Microsoft Authenticator)
  - SMS
  - Email
  - Códigos de backup
  - Métodos espaciais AR/VR
- Autenticação federada via:
  - OpenID Connect (OIDC)
  - SAML 2.0
  - OAuth 2.0

### 2. Gerenciamento de Identidade

Funcionalidades completas para gerenciamento de identidades:

- CRUD de usuários com atributos customizáveis
- Gerenciamento de ciclo de vida (provisionamento, desprovisionamento)
- Federação de identidades com provedores externos
- Gestão de organizações e departamentos

### 3. Controle de Acesso

Sistema híbrido de controle de acesso:

- RBAC (Role-Based Access Control)
- ABAC (Attribute-Based Access Control)
- Políticas de acesso complexas
- Delegação de administrativo

### 4. Autenticação AR/VR

Métodos inovadores de autenticação para realidade aumentada e virtual:

- Gestos espaciais
- Padrões de olhar
- Senhas espaciais
- Autenticação contínua com pontuação de confiança

### 5. Compliance para Saúde

Validadores específicos para as principais regulamentações:

- HIPAA (EUA)
- GDPR para dados de saúde (Europa)
- LGPD para dados de saúde (Brasil)
- PNDSB (Política Nacional de Dados em Saúde do Brasil)

## Endpoints Principais

### Autenticação

- `POST /auth/login`: Inicia a autenticação
- `POST /auth/mfa/initiate`: Inicia verificação multi-fator
- `POST /auth/mfa/verify`: Verifica código MFA
- `POST /auth/logout`: Encerra sessão
- `POST /auth/refresh`: Renova token de acesso

### Usuários

- `GET /users`: Lista usuários
- `POST /users`: Cria usuário
- `GET /users/{user_id}`: Obtém detalhes do usuário
- `PUT /users/{user_id}`: Atualiza usuário
- `DELETE /users/{user_id}`: Remove usuário

### MFA

- `GET /mfa/methods`: Lista métodos MFA disponíveis
- `POST /mfa/methods`: Registra novo método MFA
- `DELETE /mfa/methods/{method_id}`: Remove método MFA

### AR/VR

- `POST /ar/methods`: Registra método AR/VR
- `GET /ar/methods`: Lista métodos AR/VR
- `POST /ar/continuous-auth/start`: Inicia autenticação contínua
- `PUT /ar/continuous-auth/update`: Atualiza confiança da sessão

### Healthcare Compliance

- `POST /healthcare/compliance/validate/{regulation}`: Executa validação de compliance
- `GET /healthcare/compliance/history`: Histórico de validações
- `GET /healthcare/compliance/requirements`: Lista requisitos de compliance

## Mecanismos de Segurança

### Multi-Tenancy

A API implementa isolamento completo entre tenants usando:

- Contexto de tenant em middleware
- Políticas de segurança em nível de linha (RLS) no banco de dados
- Verificação de tenant em todas as operações

### Auditoria

Sistema de auditoria abrangente que registra:

- Tentativas de login (bem-sucedidas e falhas)
- Operações CRUD em entidades sensíveis
- Alterações em papéis e permissões
- Validações de compliance
- Eventos de segurança

### Validação de Entrada

- Validação de todos os dados de entrada usando Pydantic
- Sanitização de parâmetros
- Proteção contra injeção SQL

## Considerações de Deployment

### Variáveis de Ambiente

As principais variáveis de ambiente incluem:

- `IAM_API_HOST`: Host para o servidor
- `IAM_API_PORT`: Porta para o servidor
- `IAM_API_SECRET_KEY`: Chave secreta para JWT
- `IAM_DATABASE_URL`: URL de conexão com o banco de dados
- `IAM_DATABASE_SCHEMA`: Schema do banco de dados
- `IAM_LOG_LEVEL`: Nível de log

### Requisitos de Sistema

- Python 3.9+
- PostgreSQL 14+
- Redis (opcional, para cache)
- 2 GB RAM mínimo
- 4 CPUs recomendados

## Desenvolvimento e Extensão

### Adição de Novos Endpoints

Para adicionar novos endpoints:

1. Crie um schema em `app/schemas/`
2. Implemente a lógica de negócios em `app/services/`
3. Crie o router em `app/routers/`
4. Registre o router em `app/main.py`

### Implementação de Novas Funcionalidades

Para implementar novas funcionalidades:

1. Crie as tabelas/funções necessárias no banco de dados
2. Implemente os serviços correspondentes
3. Crie os endpoints REST
4. Atualize a documentação

## Testes

A API inclui testes automatizados:

- Testes unitários para validação de lógica
- Testes de integração para fluxos completos
- Testes de performance
- Testes de segurança
