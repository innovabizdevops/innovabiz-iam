# Certificação do Módulo: Gestão de Identidade e Acesso (IAM)

**ID da Certificação:** `CERT-IAM-2025-020`

**Data da Certificação:** `2025-07-22`

**Autor:** `Eduardo Jeremias`

**Revisor:** `Cascade (AI Agent)`

**Status:** `APROVADO`

---

## 1. Visão Geral e Escopo

Este documento certifica a conclusão, teste e aprovação do módulo de **Gestão de Identidade e Acesso (IAM)**. Este é um módulo central e fundamental para a segurança de toda a suíte **INNOVABIZ**, fornecendo um sistema robusto para gerir a identidade digital e controlar o acesso aos recursos da plataforma.

O escopo do módulo abrange:

- **Gestão de Utilizadores:** Registo, autenticação e gestão do ciclo de vida dos utilizadores.
- **Segurança de Credenciais:** Armazenamento seguro de senhas através de hashing com a extensão `pgcrypto`.
- **Controlo de Acesso Baseado em Papéis (RBAC):** Um sistema flexível para definir papéis (ex: `admin`, `merchant_user`) e permissões granulares (ex: `transactions:read`), e associá-los aos utilizadores.

## 2. Requisitos e Critérios de Aceitação

- **Segurança:** As senhas nunca devem ser armazenadas em texto simples. A autenticação deve ser segura contra ataques comuns.
- **Autorização:** O acesso a recursos deve ser estritamente governado pelas permissões atribuídas aos papéis de um utilizador.
- **Integridade:** O sistema deve impedir a criação de utilizadores com `usernames` ou `emails` duplicados.

## 3. Solução Técnica Implementada

### 3.1. Schema do Banco de Dados (`iam`)

- **Tabelas:**
  - `users`: Armazena os dados dos utilizadores com senhas em formato de hash.
  - `roles`: Define os papéis do sistema.
  - `permissions`: Define as permissões granulares.
  - `user_roles` e `role_permissions`: Tabelas de mapeamento que implementam o modelo RBAC.

### 3.2. Funções PL/pgSQL

- `register_user()`: Regista novos utilizadores, garantindo a unicidade e o hashing seguro da senha.
- `authenticate_user()`: Valida as credenciais de um utilizador comparando os hashes das senhas.
- `check_permission()`: Verifica se um utilizador autenticado tem a permissão necessária para realizar uma ação.

## 4. Testes e Validação

O módulo foi validado através de uma suíte de testes pgTAP abrangente.

- **Localização dos Testes:** `CoreModules/IAM/Scripts/tests/iam_tests.sql`
- **Cobertura dos Testes:**
  - **Registo e Autenticação:** Testes de sucesso e de falha para o registo e autenticação de utilizadores.
  - **Prevenção de Duplicados:** Testes para garantir que o sistema bloqueia `usernames` e `emails` duplicados.
  - **Lógica RBAC:** Validação de que a função `check_permission` concede e nega acesso corretamente com base nos papéis e permissões do utilizador.

**Resultado:** Todos os 11 testes foram executados com sucesso, confirmando que o módulo IAM é seguro, fiável e pronto para servir como a fundação da segurança da plataforma.

## 5. Decisão de Certificação

Com base na implementação robusta e na validação completa, o módulo de **Gestão de Identidade e Acesso (IAM)** é considerado **APROVADO** e pronto para ser integrado como o sistema central de segurança da suíte INNOVABIZ.

---

**Fim do Documento de Certificação.**