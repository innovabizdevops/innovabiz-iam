# 🔄 Fluxos de Trabalho de Autenticação - Parte 2

## 📖 Visão Geral

Este documento continua a definição dos fluxos de trabalho para implementação dos métodos de autenticação no módulo IAM da plataforma INNOVABIZ, seguindo os padrões ISO/IEC 27001, NIST 800-63, FIDO, OpenID Connect e OAuth 2.0, alinhados com os princípios de Arquitetura de Integração Total e Governança Aumentada.

## 📱 Fluxos de Trabalho de Autenticação por Dispositivo

### Autenticação por Certificado em Dispositivo

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant IAM as IAM Core
    participant PKI as Infraestrutura PKI
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    U->>App: Inicia autenticação por certificado
    App->>App: Acessa certificado no keystore do dispositivo
    App->>IAM: Inicia TLS mútuo com certificado do cliente
    IAM->>PKI: Valida cadeia de certificados
    PKI->>PKI: Verifica status de revogação (CRL/OCSP)
    PKI->>IAM: Retorna status de validação do certificado
    
    alt Certificado válido
        IAM->>IAM: Extrai identidade do subject do certificado
        IAM->>IAM: Mapeia certificado para usuário
        IAM->>AS: Solicita tokens de autenticação
        AS->>IAM: Gera tokens de sessão (Access/ID/Refresh)
    else Certificado inválido/revogado
        IAM->>App: Retorna erro de autenticação
    end
    
    IAM->>AuditLog: Registra evento de autenticação por certificado
    IAM->>App: Retorna resultado e tokens (se sucesso)
    App->>U: Apresenta resultado da autenticação
```

### Autenticação por Token de Hardware FIDO2

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant RP as Relying Party (Backend)
    participant IAM as IAM Core
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    U->>App: Inicia login com token de segurança
    App->>RP: Solicita desafio de autenticação
    RP->>App: Envia desafio WebAuthn
    App->>App: Invoca navigator.credentials.get() com desafio
    App->>U: Solicita inserção/toque no token de segurança
    U->>App: Insere/toca o token de segurança
    App->>App: Token gera assinatura criptográfica
    App->>RP: Envia resposta de asserção WebAuthn
    RP->>IAM: Valida asserção WebAuthn
    IAM->>IAM: Verifica assinatura, desafio e contador
    IAM->>AS: Solicita geração de tokens
    AS->>IAM: Retorna tokens de autenticação
    IAM->>AuditLog: Registra autenticação FIDO2
    IAM->>App: Transmite resultado e tokens
    App->>U: Confirma autenticação bem-sucedida
```

### Autenticação por Dispositivo de IoT

```mermaid
sequenceDiagram
    participant D as Dispositivo IoT
    participant GW as Gateway IoT
    participant IAM as IAM Core
    participant IDP as Provedor de Identidade
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    D->>GW: Inicia conexão com certificado de dispositivo
    GW->>IAM: Solicita autenticação de dispositivo
    IAM->>IAM: Valida certificado/credencial do dispositivo
    IAM->>IDP: Verifica identidade do dispositivo
    IDP->>IAM: Confirma identidade do dispositivo
    IAM->>IAM: Avalia atributos de segurança do dispositivo
    
    alt Dispositivo confiável
        IAM->>AS: Solicita tokens de acesso limitado
        AS->>IAM: Gera tokens com escopo específico
    else Dispositivo não confiável
        IAM->>GW: Nega acesso ao dispositivo
    end
    
    IAM->>AuditLog: Registra autenticação de dispositivo IoT
    IAM->>GW: Retorna resultado de autenticação
    GW->>D: Confirma/rejeita conexão
```

## 🔄 Fluxos de Autenticação Contínua

### Autenticação Contínua Comportamental

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant BC as Coletor Comportamental
    participant IAM as IAM Core
    participant ML as Motor de Aprendizado de Máquina
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    Note over U,AuditLog: Autenticação inicial já realizada
    loop Monitoramento Contínuo
        BC->>BC: Coleta padrões comportamentais passivos
        Note over BC: Tipificação, movimento do mouse, padrões de uso
        BC->>IAM: Envia telemetria comportamental
        IAM->>ML: Processa telemetria comportamental
        ML->>ML: Compara com perfil comportamental do usuário
        ML->>IAM: Calcula score de confiança contínuo
        
        alt Score abaixo do limiar crítico
            IAM->>App: Solicita reautenticação imediata
            App->>U: Exibe solicitação de reautenticação
            U->>App: Fornece fator de autenticação
            App->>IAM: Valida fator de autenticação
        else Score abaixo do limiar de atenção
            IAM->>App: Solicita fator adicional em segundo plano
            App->>U: Solicita verificação adicional
            U->>App: Fornece fator adicional
            App->>IAM: Valida fator adicional
        else Score dentro da faixa aceitável
            IAM->>AS: Mantém sessão ativa
        end
        
        IAM->>AuditLog: Registra avaliação de autenticação contínua
    end
```

### Autenticação Step-Up para Transação de Alto Risco

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant API as API Gateway
    participant IAM as IAM Core
    participant Risk as Motor de Risco
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    Note over U,AuditLog: Sessão autenticada já estabelecida
    U->>App: Inicia transação sensível
    App->>API: Solicita autorização para transação
    API->>IAM: Verifica nível de autenticação atual
    IAM->>Risk: Avalia risco da transação específica
    Risk->>Risk: Calcula score de risco da transação
    Risk->>IAM: Retorna nível de autenticação necessário
    
    alt Nível atual suficiente
        IAM->>API: Autoriza transação com nível atual
    else Necessita step-up
        IAM->>App: Solicita autenticação adicional
        App->>U: Solicita fator adicional
        U->>App: Fornece fator adicional
        App->>IAM: Valida fator adicional
        IAM->>AS: Atualiza contexto de autenticação
        AS->>IAM: Retorna tokens com novo nível
        IAM->>API: Autoriza transação com nível elevado
    end
    
    API->>App: Retorna resultado da transação
    IAM->>AuditLog: Registra decisão de step-up e transação
    App->>U: Apresenta resultado da transação
```

## 🌐 Fluxos de Trabalho de Federação e SSO

### Autenticação via SSO Empresarial

```mermaid
sequenceDiagram
    participant U as Usuário
    participant SP as Service Provider (INNOVABIZ)
    participant IAM as IAM Core
    participant IDP as Identity Provider Externo
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    U->>SP: Acessa aplicação e inicia login
    SP->>IAM: Verifica opções de autenticação
    IAM->>IAM: Identifica domínio empresarial
    IAM->>SP: Redireciona para fluxo SSO
    SP->>IDP: Redireciona com SAML Request/OIDC Auth Request
    IDP->>U: Solicita credenciais corporativas
    
    alt Já autenticado no IDP
        IDP->>IDP: Verifica sessão SSO existente
    else Não autenticado
        U->>IDP: Fornece credenciais corporativas
        IDP->>IDP: Valida credenciais e políticas
        IDP->>IDP: Aplica MFA conforme política corporativa
        U->>IDP: Completa MFA se necessário
    end
    
    IDP->>SP: Retorna SAML Response/OIDC Token
    SP->>IAM: Valida asserção/token e extrai claims
    IAM->>IAM: Mapeia atributos externos para perfil interno
    IAM->>AS: Gera tokens de sessão INNOVABIZ
    IAM->>AuditLog: Registra login federado
    IAM->>SP: Estabelece sessão autenticada
    SP->>U: Apresenta interface autenticada
```

### Autenticação Social com Step-Up

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant IAM as IAM Core
    participant Social as Provedor Social
    participant AS as Servidor de Autenticação
    participant AuditLog as Sistema de Auditoria

    U->>App: Seleciona login social
    App->>IAM: Inicia fluxo OAuth para provedor social
    IAM->>Social: Redireciona para autorização
    Social->>U: Solicita autorização
    U->>Social: Aprova autorização
    Social->>IAM: Retorna código de autorização
    IAM->>Social: Troca código por access_token
    IAM->>Social: Obtém informações do perfil
    IAM->>IAM: Vincula ou cria conta local
    IAM->>IAM: Avalia nível de confiança da identidade social
    
    alt Nível suficiente para acesso básico
        IAM->>AS: Gera tokens com nível básico
    else Necessário step-up para acesso completo
        IAM->>App: Solicita verificação adicional
        App->>U: Solicita segundo fator
        U->>App: Fornece segundo fator
        App->>IAM: Valida segundo fator
        IAM->>AS: Gera tokens com acesso completo
    end
    
    IAM->>AuditLog: Registra autenticação social
    IAM->>App: Retorna tokens e nível de acesso
    App->>U: Apresenta interface com acesso apropriado
```

## 🛡️ Fluxos de Recuperação e Delegação

### Recuperação de Acesso Sem Senha

```mermaid
sequenceDiagram
    participant U as Usuário
    participant App as Aplicação Cliente
    participant IAM as IAM Core
    participant Recovery as Serviço de Recuperação
    participant Comms as Serviços de Comunicação
    participant Verify as Verificação de Identidade
    participant AuditLog as Sistema de Auditoria

    U->>App: Solicita recuperação de acesso
    App->>IAM: Inicia fluxo de recuperação
    IAM->>Recovery: Avalia opções de recuperação disponíveis
    Recovery->>IAM: Retorna métodos possíveis para o usuário
    IAM->>App: Apresenta opções de recuperação
    U->>App: Seleciona método de recuperação
    
    alt Recuperação por Email
        App->>IAM: Solicita recuperação por email
        IAM->>Comms: Envia email de recuperação
        Comms->>U: Entrega email com link de recuperação
        U->>App: Clica no link de recuperação
    else Recuperação por SMS
        App->>IAM: Solicita recuperação por SMS
        IAM->>Comms: Envia código por SMS
        Comms->>U: Entrega código via SMS
        U->>App: Insere código de recuperação
    else Recuperação por Perguntas de Segurança
        App->>IAM: Solicita desafio de perguntas
        IAM->>App: Envia perguntas de segurança
        U->>App: Responde perguntas de segurança
        App->>IAM: Submete respostas para validação
    end
    
    IAM->>Verify: Valida provas de identidade
    Verify->>IAM: Confirma identidade do usuário
    
    IAM->>App: Autoriza estabelecimento de novas credenciais
    App->>U: Solicita configuração de novas credenciais
    U->>App: Configura novas credenciais
    App->>IAM: Registra novas credenciais
    IAM->>AuditLog: Registra evento de recuperação
    IAM->>App: Confirma recuperação completa
    App->>U: Notifica recuperação bem-sucedida
```

### Delegação de Autenticação (Break-Glass)

```mermaid
sequenceDiagram
    participant U1 as Usuário Solicitante
    participant U2 as Aprovador Autorizado
    participant App as Aplicação Cliente
    participant IAM as IAM Core
    participant Delegate as Serviço de Delegação
    participant Notify as Serviço de Notificação
    participant AuditLog as Sistema de Auditoria

    U1->>App: Solicita acesso emergencial
    App->>IAM: Submete solicitação de acesso emergencial
    IAM->>Delegate: Avalia solicitação de emergência
    Delegate->>Delegate: Identifica aprovadores autorizados
    Delegate->>Notify: Envia notificações de aprovação
    Notify->>U2: Entrega solicitação de aprovação
    
    U2->>Notify: Revisa e aprova solicitação
    Notify->>Delegate: Transmite aprovação
    Delegate->>Delegate: Valida aprovação e autorização
    Delegate->>IAM: Autoriza acesso emergencial temporário
    IAM->>IAM: Configura credenciais temporárias
    IAM->>IAM: Define limites de tempo e escopo reduzido
    IAM->>App: Retorna credenciais temporárias
    IAM->>AuditLog: Registra acesso emergencial com detalhes
    App->>U1: Concede acesso temporário
    
    Note over AuditLog: Monitoramento intensivo durante sessão emergencial
    
    IAM->>Notify: Agenda notificações de expiração
    Notify->>U1: Notifica sobre tempo restante
    Notify->>U2: Notifica sobre uso do acesso emergencial
```