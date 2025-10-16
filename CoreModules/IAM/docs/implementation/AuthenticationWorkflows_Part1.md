# 🔄 Fluxos de Trabalho de Autenticação - Parte 1

## 📖 Visão Geral

Este documento técnico define os fluxos de trabalho detalhados para implementação dos métodos de autenticação no módulo IAM da plataforma INNOVABIZ. Os workflows apresentados seguem os padrões da indústria (ISO/IEC 27001, NIST 800-63, FIDO, OpenID Connect, OAuth 2.0), incorporando controles de segurança robustos e promovendo uma experiência do usuário otimizada.

## 🔐 Fluxos de Trabalho de Autenticação Biométrica

### Autenticação por Impressão Digital

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant BiometricAPI as API Biométrica
    participant IAM as IAM Core
    participant AS as Servidor de Autenticação
    participant TEE as Trusted Execution Environment
    participant AuditLog as Sistema de Auditoria

    U->>App: Inicia autenticação
    App->>App: Verifica capacidade biométrica
    App->>BiometricAPI: Solicita verificação por impressão digital
    BiometricAPI->>TEE: Solicita verificação em ambiente seguro
    TEE->>TEE: Realiza verificação da impressão digital
    TEE->>TEE: Executa detecção de vivacidade
    TEE->>BiometricAPI: Retorna resultado da verificação (sem dados biométricos)
    BiometricAPI->>IAM: Envia resultado + metadados de verificação
    IAM->>AS: Solicita validação de sessão
    AS->>AS: Aplica regras de política de acesso
    AS->>IAM: Retorna token de autenticação
    IAM->>App: Envia resposta de autenticação
    IAM->>AuditLog: Registra evento de autenticação (sem dados biométricos)
    App->>U: Apresenta resultado da autenticação
```

### Autenticação Facial

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant BiometricAPI as API Biométrica
    participant IAM as IAM Core
    participant AS as Servidor de Autenticação
    participant PAD as Sistema de Detecção de Ataque
    participant AuditLog as Sistema de Auditoria

    U->>App: Inicia autenticação facial
    App->>BiometricAPI: Solicita autenticação facial
    BiometricAPI->>BiometricAPI: Captura dados faciais
    BiometricAPI->>PAD: Executa detecção de ataque de apresentação
    PAD->>PAD: Verifica vivacidade e anti-spoofing
    PAD->>BiometricAPI: Retorna resultado da verificação de vivacidade
    BiometricAPI->>BiometricAPI: Processa template facial seguro
    BiometricAPI->>IAM: Envia resultado e metadados de verificação
    IAM->>IAM: Valida score biométrico contra limiar
    IAM->>IAM: Aplica regras adaptativas de política
    IAM->>AS: Solicita token de autenticação
    AS->>IAM: Retorna token/sessão
    IAM->>AuditLog: Registra evento (compliance com GDPR/LGPD)
    IAM->>App: Envia resultado da autenticação
    App->>U: Apresenta resultado ao usuário
```

## 🔑 Fluxos de Trabalho de Autenticação Sem Senha

### Autenticação com Passkeys

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant RP as Relying Party (Backend)
    participant IAM as IAM Core
    participant Auth as Servidor de Autenticação
    participant Audit as Sistema de Auditoria

    U->>App: Inicia login
    App->>RP: Solicita opções de autenticação
    RP->>App: Retorna opções WebAuthn/Passkey
    App->>App: Invoca API WebAuthn navigator.credentials.get()
    Note over App: Plataforma busca credenciais disponíveis
    App->>U: Solicita desbloqueio da Passkey (biometria/PIN)
    U->>App: Fornece verificação (impressão digital/face/PIN)
    Note over App: WebAuthn gera asserção criptográfica
    App->>RP: Envia asserção para validação
    RP->>IAM: Valida assinatura criptográfica
    IAM->>IAM: Verifica desafio e assinatura
    IAM->>Auth: Gera tokens (access/id/refresh)
    IAM->>Audit: Registra autenticação bem-sucedida
    RP->>App: Retorna tokens e status de autenticação
    App->>U: Confirma login bem-sucedido
```

### Autenticação por Magic Link

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant IAM as IAM Core
    participant Email as Serviço de Email
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    U->>App: Insere email e solicita login
    App->>IAM: Envia solicitação de magic link
    IAM->>IAM: Gera token único de uso único
    IAM->>IAM: Associa token ao usuário + dispositivo + timeout
    IAM->>Email: Solicita envio de email com link seguro
    Email->>U: Envia email com magic link
    U->>U: Abre email e clica no magic link
    Note over IAM: Link contém token assinado e criptografado
    IAM->>IAM: Valida token (expiração, uso único, dispositivo)
    IAM->>AS: Solicita criação de sessão autenticada
    AS->>IAM: Retorna tokens de sessão
    IAM->>AuditLog: Registra evento de autenticação bem-sucedida
    IAM->>App: Redireciona com tokens de sessão
    App->>U: Apresenta interface autenticada
```

## 🛡️ Fluxos de Trabalho de Autenticação Multi-Fator

### Autenticação Multi-Fator Baseada em Risco

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant IAM as IAM Core
    participant RiskEngine as Motor de Risco
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    U->>App: Insere credenciais primárias (senha/biometria)
    App->>IAM: Valida credencial primária
    IAM->>RiskEngine: Solicita avaliação de risco contextual
    RiskEngine->>RiskEngine: Analisa fatores de risco
    Note over RiskEngine: Avalia localização, dispositivo, comportamento, hora, rede
    RiskEngine->>IAM: Retorna score de risco
    
    alt Score de risco baixo
        IAM->>AS: Solicita autenticação simplificada
        AS->>IAM: Retorna token com nível de confiança
    else Score de risco médio
        IAM->>App: Solicita segundo fator
        App->>U: Solicita segundo fator ao usuário
        U->>App: Fornece segundo fator
        App->>IAM: Submete segundo fator para validação
        IAM->>AS: Valida segundo fator
    else Score de risco alto
        IAM->>App: Solicita múltiplos fatores adicionais
        App->>U: Solicita fatores adicionais
        U->>App: Fornece fatores adicionais
        App->>IAM: Submete fatores para validação
        IAM->>AS: Valida múltiplos fatores
    end
    
    AS->>IAM: Retorna tokens de autenticação
    IAM->>AuditLog: Registra evento de autenticação com nível de risco
    IAM->>App: Transmite tokens e resultado de autenticação
    App->>U: Apresenta interface autenticada
```

### Autenticação Push com Aprovação Explícita

```mermaid
sequenceDiagram
    participant U as Usuário
    participant WebApp as Aplicação Web
    participant MobileApp as Aplicação Móvel
    participant Push as Serviço de Notificações
    participant IAM as IAM Core
    participant Auth as Servidor de Autenticação
    participant Audit as Sistema de Auditoria

    U->>WebApp: Insere identificador de usuário
    WebApp->>IAM: Solicita autenticação push
    IAM->>IAM: Gera transação de autenticação
    IAM->>Push: Envia notificação push para dispositivo registrado
    Push->>MobileApp: Entrega notificação push
    MobileApp->>MobileApp: Verifica origem da transação
    MobileApp->>U: Apresenta detalhes da solicitação de login
    MobileApp->>U: Solicita verificação (PIN/biometria)
    U->>MobileApp: Fornece verificação e aprova
    MobileApp->>MobileApp: Assina token de aprovação
    MobileApp->>IAM: Envia confirmação assinada
    IAM->>IAM: Valida assinatura de aprovação
    IAM->>Auth: Solicita criação de sessão autenticada
    Auth->>IAM: Retorna tokens de sessão
    IAM->>Audit: Registra evento de autenticação push
    IAM->>WebApp: Transmite tokens de sessão
    WebApp->>U: Completa login automaticamente
```

## 🌐 Fluxos de Trabalho de Autenticação Contextual

### Autenticação Adaptativa Baseada em Contexto

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant Context as Coletor de Contexto
    participant IAM as IAM Core
    participant Policy as Motor de Políticas
    participant Auth as Servidor de Autenticação
    participant Audit as Sistema de Auditoria

    U->>App: Inicia processo de login
    App->>Context: Coleta informações contextuais
    Note over Context: Coleta localização, dispositivo, horário, IP, rede
    Context->>IAM: Envia pacote de contexto
    IAM->>Policy: Solicita requisitos de autenticação
    Policy->>Policy: Avalia matriz de requisitos x contexto
    
    alt Contexto familiar e baixo risco
        Policy->>IAM: Requer autenticação simplificada
    else Contexto parcialmente conhecido
        Policy->>IAM: Requer autenticação padrão
    else Contexto não familiar ou de alto risco
        Policy->>IAM: Requer autenticação reforçada
    end
    
    IAM->>App: Solicita fatores de autenticação requeridos
    App->>U: Solicita fatores específicos ao contexto
    U->>App: Fornece fatores de autenticação
    App->>IAM: Envia fatores para validação
    IAM->>Auth: Valida fatores e constrói sessão
    Auth->>IAM: Retorna tokens com contexto associado
    IAM->>Audit: Registra decisão baseada em contexto
    IAM->>App: Transmite resultado de autenticação
    App->>U: Acesso concedido com nível apropriado
```