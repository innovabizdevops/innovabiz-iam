# Modelo de Dados do Módulo IAM - INNOVABIZ

## Visão Geral

Este documento descreve o modelo de dados do módulo IAM (Identity and Access Management) da plataforma INNOVABIZ, com ênfase no suporte aos 70 métodos de autenticação conforme definido no plano de implementação. O modelo foi projetado para atender aos seguintes requisitos:

- **Multi-tenancy**: Suporte a múltiplos clientes (tenants) em uma única instância
- **Multi-regional**: Adaptações específicas para as regiões-alvo (UE/Portugal, Brasil, Angola, EUA)
- **Conformidade regulatória**: Aderência ao GDPR, LGPD, PNDSB e regulamentações dos EUA
- **Escalabilidade**: Estrutura otimizada para alto volume de autenticações
- **Flexibilidade**: Facilidade para adicionar novos métodos de autenticação
- **Segurança**: Conformidade com as melhores práticas de segurança (NIST, ISO 27001)

## Estrutura do Banco de Dados

O módulo IAM possui sua própria base de dados que se integra à base de dados principal da plataforma INNOVABIZ. Todas as tabelas são organizadas sob o esquema `iam` para isolamento e organização.

### Entidades Principais

#### 1. Tenants (`iam.tenants`)

Armazena informações sobre as organizações que utilizam a plataforma.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| tenant_id | UUID | Identificador único do tenant |
| tenant_code | VARCHAR(50) | Código do tenant (único) |
| nome | VARCHAR(200) | Nome do tenant |
| descricao | TEXT | Descrição do tenant |
| dominio | VARCHAR(255) | Domínio principal do tenant |
| regiao | VARCHAR(10) | Região principal (EU, BR, AO, US) |
| configuracoes | JSONB | Configurações específicas em formato JSON |
| plano | VARCHAR(50) | Plano de assinatura |
| status | VARCHAR(20) | Status do tenant (ativo, inativo, bloqueado, trial) |

#### 2. Usuários (`iam.usuarios`)

Armazena informações dos usuários da plataforma.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| usuario_id | UUID | Identificador único do usuário |
| tenant_id | UUID | Referência ao tenant |
| nome_usuario | VARCHAR(100) | Nome de usuário (login) |
| email | VARCHAR(255) | E-mail do usuário |
| telefone | VARCHAR(50) | Telefone do usuário |
| senha_hash | TEXT | Hash da senha (quando aplicável) |
| nome_completo | VARCHAR(200) | Nome completo do usuário |
| status | VARCHAR(20) | Status do usuário (ativo, inativo, bloqueado, pendente, excluído) |
| dados_verificados | BOOLEAN | Indica se os dados foram verificados |
| email_verificado | BOOLEAN | Indica se o e-mail foi verificado |
| telefone_verificado | BOOLEAN | Indica se o telefone foi verificado |
| mfa_obrigatorio | BOOLEAN | Indica se MFA é obrigatório |
| tentativas_falhas | INTEGER | Contador de tentativas de login falhas |
| dados_perfil | JSONB | Dados de perfil em formato JSON |

#### 3. Métodos de Autenticação (`iam.metodos_autenticacao`)

Catálogo dos métodos de autenticação disponíveis.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| metodo_id | VARCHAR(10) | Identificador único do método (ex: K01, P02) |
| codigo_metodo | VARCHAR(50) | Código interno do método |
| nome_pt | VARCHAR(100) | Nome em português |
| nome_en | VARCHAR(100) | Nome em inglês |
| categoria | VARCHAR(50) | Categoria (knowledge, possession, biometric, context) |
| fator | VARCHAR(20) | Fator de autenticação (knowledge, possession, inherence) |
| complexidade | VARCHAR(20) | Nível de complexidade de implementação |
| prioridade | INTEGER | Prioridade do método (0-100) |
| onda_implementacao | INTEGER | Onda de implementação (1-7) |
| nivel_seguranca | VARCHAR(20) | Nível de segurança oferecido |
| status | VARCHAR(20) | Status do método (planejado, desenvolvimento, ativo, desativado, depreciado) |
| adaptacoes_regionais | JSONB | Adaptações específicas por região |

#### 4. Métodos do Usuário (`iam.usuario_metodos`)

Associação dos métodos de autenticação a cada usuário.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| usuario_metodo_id | UUID | Identificador único da associação |
| usuario_id | UUID | Referência ao usuário |
| metodo_id | VARCHAR(10) | Referência ao método de autenticação |
| habilitado | BOOLEAN | Indica se o método está habilitado |
| verificado | BOOLEAN | Indica se o método foi verificado |
| preferencial | BOOLEAN | Indica se é o método preferencial |
| dados_autenticacao | JSONB | Dados específicos do método para o usuário |
| nome_dispositivo | VARCHAR(200) | Nome do dispositivo (quando aplicável) |

#### 5. Sessões (`iam.sessoes`)

Armazena informações sobre sessões ativas.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| sessao_id | UUID | Identificador único da sessão |
| usuario_id | UUID | Referência ao usuário |
| token_refresh | TEXT | Token de atualização |
| cliente_id | VARCHAR(100) | Identificador do cliente/aplicação |
| ip_address | VARCHAR(45) | Endereço IP de origem |
| dispositivo_id | VARCHAR(255) | Identificador do dispositivo |
| data_criacao | TIMESTAMP | Data de criação da sessão |
| data_expiracao | TIMESTAMP | Data de expiração da sessão |
| ativa | BOOLEAN | Indica se a sessão está ativa |
| fatores_autenticados | JSONB | Lista de fatores utilizados na autenticação |
| nivel_autenticacao | VARCHAR(20) | Nível de autenticação (single_factor, two_factor, multi_factor) |

#### 6. Aplicações (`iam.aplicacoes`)

Aplicações registradas por tenant.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| aplicacao_id | UUID | Identificador único da aplicação |
| tenant_id | UUID | Referência ao tenant |
| nome | VARCHAR(200) | Nome da aplicação |
| app_tipo | VARCHAR(50) | Tipo da aplicação (web, mobile, desktop, api) |
| cliente_id | VARCHAR(100) | ID do cliente OAuth |
| cliente_secret | TEXT | Secret do cliente OAuth |
| redirect_uris | TEXT[] | URIs de redirecionamento permitidas |
| status | VARCHAR(20) | Status da aplicação |

#### 7. Fluxos de Autenticação (`iam.fluxos_autenticacao`)

Definição de fluxos de autenticação configuráveis.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| fluxo_id | UUID | Identificador único do fluxo |
| tenant_id | UUID | Referência ao tenant |
| nome | VARCHAR(100) | Nome do fluxo |
| passos | JSONB | Passos do fluxo em formato JSON |
| adaptativo | BOOLEAN | Indica se o fluxo é adaptativo baseado em risco |
| nivel_seguranca | VARCHAR(20) | Nível de segurança do fluxo |
| status | VARCHAR(20) | Status do fluxo |

#### 8. Perfis de Risco (`iam.perfis_risco`)

Perfis de risco dos usuários para autenticação adaptativa.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| perfil_id | UUID | Identificador único do perfil |
| usuario_id | UUID | Referência ao usuário |
| score_risco | INTEGER | Pontuação de risco (0-100) |
| nivel_risco | VARCHAR(20) | Nível de risco (baixo, médio, alto) |
| localizacoes_comuns | JSONB | Localizações conhecidas do usuário |
| dispositivos_comuns | JSONB | Dispositivos conhecidos do usuário |
| padroes_tempo | JSONB | Padrões temporais de uso |
| padroes_comportamento | JSONB | Padrões comportamentais |
| anomalias_detectadas | JSONB | Registro de anomalias detectadas |

### Relacionamentos

O modelo de dados segue uma estrutura relacional com as seguintes conexões:

1. Um **tenant** pode ter muitos **usuários**
2. Um **usuário** pode ter múltiplos **métodos de autenticação** associados
3. Um **usuário** pode ter muitas **sessões** ativas
4. Um **tenant** pode definir múltiplos **fluxos de autenticação**
5. Um **usuário** possui um **perfil de risco**
6. Um **tenant** pode ter múltiplas **aplicações**
7. Um **método de autenticação** pode estar associado a múltiplos **tenants**

## Adaptações Regionais

### União Europeia (Portugal)

- Período máximo de retenção de senhas: 90 dias
- Complexidade mínima de senhas: 12 caracteres
- Histórico de senhas: 10 últimas senhas
- Privacidade por padrão ativada
- Termos e políticas específicas para GDPR

### Brasil

- Período máximo de retenção de senhas: 60 dias
- Complexidade mínima de senhas: 10 caracteres
- Histórico de senhas: 5 últimas senhas
- Integração com certificados ICP-Brasil
- Termos e políticas específicas para LGPD

### Angola

- Período máximo de retenção de senhas: 90 dias
- Complexidade mínima de senhas: 8 caracteres
- Histórico de senhas: 3 últimas senhas
- Suporte a métodos alternativos para áreas com conectividade limitada
- Termos e políticas específicas para PNDSB

### Estados Unidos

- Período máximo de retenção de senhas: 120 dias
- Complexidade mínima de senhas: 8 caracteres
- Histórico de senhas: 5 últimas senhas
- Configurações específicas para setores regulados (HIPAA, SOX, GLBA)
- Conformidade com NIST 800-63

## Armazenamento Seguro

- **Senhas**: Armazenadas usando Argon2id (algoritmo recomendado para 2025+)
- **Dados Sensíveis**: Criptografados individualmente
- **Tokens**: Utilizando mecanismos de assinatura seguros (com rotação de chaves)
- **Dados Biométricos**: Armazenados como templates seguros, nunca em formato bruto
- **Dados Pessoais**: Respeitando princípios de minimização de dados

## Considerações de Segurança

- Proteção completa contra vazamento de informações através de SQL Injection
- Prevenção de timing attacks em operações críticas como verificação de credenciais
- Monitoramento em tempo real de tentativas de autenticação suspeitas
- Auditoria completa de todas as operações de autenticação
- Índices otimizados para redução do tempo de resposta sem comprometer a segurança

## Extensibilidade

O modelo foi projetado para facilitar a adição de novos métodos de autenticação através de:

1. Cadastro do novo método na tabela `iam.metodos_autenticacao`
2. Habilitação do método para tenants específicos
3. Configuração das adaptações regionais
4. Integração com os fluxos de autenticação existentes

## Integração com a Plataforma INNOVABIZ

O módulo IAM se integra com outros módulos da plataforma INNOVABIZ através de:

1. APIs internas seguras
2. Eventos publicados no barramento de eventos
3. Autenticação e autorização centralizada via KrakenD API Gateway
4. Suporte ao Model Context Protocol (MCP) para comunicação entre módulos

---

© 2025 INNOVABIZ - Todos os direitos reservados
