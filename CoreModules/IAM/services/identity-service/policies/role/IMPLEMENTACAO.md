# Guia de Implementação - Políticas OPA RoleService

Este documento fornece instruções detalhadas para implementação, configuração e utilização das políticas OPA (Open Policy Agent) para o serviço de gerenciamento de funções (RoleService) da plataforma INNOVABIZ.

## Índice
1. [Estrutura dos Arquivos e Diretórios](#estrutura-dos-arquivos-e-diretórios)
2. [Configuração do OPA](#configuração-do-opa)
   - [Instalação Local](#instalação-local)
   - [Implantação em Kubernetes](#implantação-em-kubernetes)
3. [Integração com Aplicação](#integração-com-aplicação)
   - [Integração via Sidecar](#integração-via-sidecar)
   - [Integração via SDK Go](#integração-via-sdk-go)
   - [Integração via Middleware HTTP](#integração-via-middleware-http)
4. [Modelo de Input para Decisões](#modelo-de-input-para-decisões)
5. [Exemplos de Uso](#exemplos-de-uso)
6. [Monitoramento e Governança](#monitoramento-e-governança)
7. [Resolução de Problemas](#resolução-de-problemas)

## Estrutura dos Arquivos e Diretórios

```
policies/role/
│
├── src/                      # Código fonte das políticas OPA
│   ├── main.rego             # Política principal e roteamento
│   ├── common.rego           # Funções comuns e utilitárias
│   ├── constants.rego        # Constantes globais
│   ├── crud.rego             # Operações CRUD para funções
│   ├── permissions.rego      # Gerenciamento de permissões
│   ├── hierarchy.rego        # Hierarquia de funções
│   ├── user_assignment.rego  # Atribuição de funções a usuários
│   └── audit.rego            # Regras de auditoria e logs
│
├── test/                     # Testes unitários e integrados
│   ├── crud_test.rego        # Testes para operações CRUD
│   ├── permissions_test.rego # Testes para gerenciamento de permissões
│   ├── hierarchy_test.rego   # Testes para hierarquia de funções
│   ├── user_assignment_test_part1.rego  # Testes para atribuição de funções (parte 1)
│   ├── user_assignment_test_part2.rego  # Testes para atribuição de funções (parte 2)
│   ├── user_assignment_test_part3.rego  # Testes para atribuição de funções (parte 3)
│   └── integration_test.rego # Testes integrados
│
├── CONFORMIDADE.md           # Análise de conformidade com normas
└── IMPLEMENTACAO.md          # Este documento
```

## Configuração do OPA

### Instalação Local

1. **Instalar OPA**:

   ```bash
   curl -L -o opa https://openpolicyagent.org/downloads/latest/opa_linux_amd64
   chmod 755 ./opa
   mv opa /usr/local/bin/opa
   ```

2. **Iniciar OPA localmente**:

   ```bash
   opa run --server --addr :8181 --log-level debug
   ```

3. **Carregar políticas**:

   ```bash
   opa bundle build policies/role/src/ -o bundle.tar.gz
   curl -X PUT http://localhost:8181/v1/policies/innovabiz/iam/role --data-binary @bundle.tar.gz
   ```

### Implantação em Kubernetes

1. **Criar ConfigMap para políticas**:

   ```bash
   kubectl create configmap role-policies --from-file=policies/role/src/
   ```

2. **Aplicar manifesto para deployment do OPA**:

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: opa-role-service
     namespace: innovabiz-iam
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: opa-role-service
     template:
       metadata:
         labels:
           app: opa-role-service
       spec:
         containers:
         - name: opa
           image: openpolicyagent/opa:latest
           args:
           - "run"
           - "--server"
           - "--addr=:8181"
           - "--log-level=info"
           - "--log-format=json"
           - "/policies"
           volumeMounts:
           - readOnly: true
             mountPath: /policies
             name: role-policies
         volumes:
         - name: role-policies
           configMap:
             name: role-policies
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: opa-role-service
     namespace: innovabiz-iam
   spec:
     selector:
       app: opa-role-service
     ports:
     - name: http
       port: 8181
       protocol: TCP
   ```

3. **Aplicar o manifesto**:

   ```bash
   kubectl apply -f opa-deployment.yaml
   ```

## Integração com Aplicação

### Integração via Sidecar

O padrão sidecar permite que o OPA seja executado como um contêiner separado no mesmo pod que seu serviço:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: identity-service
  namespace: innovabiz-iam
spec:
  template:
    spec:
      containers:
      - name: identity-service
        # configuração do contêiner principal
      - name: opa
        image: openpolicyagent/opa:latest
        args:
        - "run"
        - "--server"
        - "--addr=:8181"
        - "--log-level=info"
        - "/policies"
        volumeMounts:
        - readOnly: true
          mountPath: /policies
          name: role-policies
      volumes:
      - name: role-policies
        configMap:
          name: role-policies
```

### Integração via SDK Go

Para integração direta no código Go, use a biblioteca `github.com/open-policy-agent/opa/sdk`:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/open-policy-agent/opa/sdk"
)

func main() {
    ctx := context.Background()
    
    // Configurar o cliente OPA
    config := sdk.Config{
        Labels: map[string]string{
            "service": "identity-service",
        },
    }
    
    opa, err := sdk.New(ctx, sdk.Options{
        Config: config,
    })
    if err != nil {
        // Tratar erro
    }
    defer opa.Stop(ctx)
    
    // Preparar input para decisão
    input := map[string]interface{}{
        "user": map[string]interface{}{
            "id": "user123",
            "roles": []string{"admin"},
        },
        "tenant_id": "tenant456",
        // outros dados...
    }
    
    // Consultar OPA para decisão
    result, err := opa.Decision(ctx, sdk.DecisionOptions{
        Path: "innovabiz/iam/role/allow",
        Input: input,
    })
    if err != nil {
        // Tratar erro
    }
    
    // Verificar resultado
    if allowed, ok := result["result"].(bool); ok && allowed {
        fmt.Println("Acesso permitido")
    } else {
        fmt.Println("Acesso negado")
    }
}
```

### Integração via Middleware HTTP

Para aplicações web em Go, use o middleware implementado para integração com OPA:

```go
package main

import (
    "net/http"
    
    "github.com/innovabiz/iam/logging"
    "github.com/innovabiz/iam/middleware"
)

func main() {
    // Configurar o logger
    logConfig := logging.LoggerConfig{
        Level:          logging.InfoLevel,
        Format:         logging.JSONFormat,
        EnableConsole:  true,
        EnableFile:     true,
        FilePath:       "/var/log/innovabiz/identity-service/app.log",
        Metadata: logging.LogMetadata{
            ServiceName:    "identity-service",
            ServiceVersion: "1.0.0",
            Environment:    "production",
            HostName:       "app-server-01",
        },
    }
    
    logger, err := logging.NewZapLogger(logConfig)
    if err != nil {
        panic(err)
    }
    
    // Configurar o middleware OPA
    opaConfig := middleware.OPAConfig{
        OPAEndpoint:    "http://localhost:8181",
        TimeoutSeconds: 5,
        EnableCache:    true,
        CacheTTL:       300, // 5 minutos
        FailOpen:       false,
        VerboseLogging: true,
    }
    
    opaMiddleware := middleware.NewOPAMiddleware(opaConfig, logger, nil)
    
    // Criar router HTTP
    mux := http.NewServeMux()
    
    // Aplicar middleware OPA
    mux.Handle("/roles", opaMiddleware.Authorize("innovabiz.iam.role.list")(handleListRoles()))
    mux.Handle("/roles/create", opaMiddleware.Authorize("innovabiz.iam.role.create")(handleCreateRole()))
    mux.Handle("/roles/update", opaMiddleware.Authorize("innovabiz.iam.role.update")(handleUpdateRole()))
    mux.Handle("/roles/delete", opaMiddleware.Authorize("innovabiz.iam.role.delete")(handleDeleteRole()))
    
    // Iniciar servidor HTTP
    http.ListenAndServe(":8080", mux)
}

func handleListRoles() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Lógica para listar funções
    })
}

// Outros handlers...
```

## Modelo de Input para Decisões

Para decisões consistentes, utilize o seguinte formato de input ao consultar o OPA:

```json
{
  "input": {
    "http_method": "GET|POST|PUT|DELETE",
    "tenant_id": "tenant123",
    "user": {
      "id": "user456",
      "username": "john.doe",
      "email": "john.doe@example.com",
      "roles": ["admin", "user"],
      "permissions": ["role:read", "role:create"],
      "access_level": "HIGH",
      "session_id": "sess789",
      "mfa_verified": true,
      "attributes": {
        "department": "IT",
        "location": "HQ"
      }
    },
    "resource": {
      "path": "/roles",
      "id": "role123",
      "data": {
        // dados do recurso, quando aplicável
      }
    },
    "context": {
      "client_ip": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "request_id": "req987",
      "origin": "https://console.innovabiz.com",
      "session_id": "sess789",
      "device_info": "desktop"
    },
    "path": "innovabiz.iam.role.action"
  }
}
```

## Exemplos de Uso

### Verificar se um usuário pode criar uma função

```bash
curl -X POST http://localhost:8181/v1/data/innovabiz/iam/role/allow \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "http_method": "POST",
      "tenant_id": "tenant123",
      "user": {
        "id": "user456",
        "roles": ["tenant_admin"],
        "permissions": ["role:create"]
      },
      "resource": {
        "path": "/roles",
        "data": {
          "name": "new_role",
          "permissions": ["user:read"]
        }
      },
      "context": {
        "client_ip": "192.168.1.100",
        "request_id": "req987"
      },
      "path": "innovabiz.iam.role.create"
    }
  }'
```

### Verificar se um usuário pode atribuir uma função a outro usuário

```bash
curl -X POST http://localhost:8181/v1/data/innovabiz/iam/role/allow \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "http_method": "POST",
      "tenant_id": "tenant123",
      "user": {
        "id": "user456",
        "roles": ["iam_admin"],
        "permissions": ["role:assign"]
      },
      "resource": {
        "path": "/users/user789/roles",
        "user_id": "user789",
        "role_id": "role123",
        "data": {
          "expires_at": "2023-12-31T23:59:59Z",
          "justification": "Necessário para projeto X"
        }
      },
      "context": {
        "client_ip": "192.168.1.100",
        "request_id": "req987"
      },
      "path": "innovabiz.iam.role.assign"
    }
  }'
```

## Monitoramento e Governança

### Métricas do OPA

O OPA expõe métricas Prometheus no endpoint `/metrics`. Configure o Prometheus para coletar essas métricas:

```yaml
scrape_configs:
  - job_name: 'opa'
    scrape_interval: 15s
    static_configs:
    - targets: ['opa:8181']
```

### Logging e Auditoria

Os logs de auditoria são gerados automaticamente para todas as decisões. Configure o armazenamento de logs:

1. **Elasticsearch + Kibana**: Para análise e visualização avançada de logs
2. **Loki + Grafana**: Para visualização de logs em tempo real

Exemplo de configuração para o Loki:

```yaml
promtail:
  config:
    clients:
      - url: http://loki:3100/loki/api/v1/push
    scrape_configs:
      - job_name: opa_logs
        static_configs:
          - targets:
              - localhost
            labels:
              job: opa_role_service
              __path__: /var/log/opa/decision_logs/*.log
```

### Dashboards de Governança

Implementar dashboards para monitoramento e governança:

1. **Dashboard de Decisões**: Quantidade e tipos de decisões (permitidas/negadas)
2. **Dashboard de Conformidade**: Alinhamento com normas (ISO 27001, etc.)
3. **Dashboard de Exceções**: Decisões que utilizaram regras de exceção
4. **Dashboard de Performance**: Tempo de resposta para decisões

## Resolução de Problemas

### Problemas Comuns

1. **Decisões inesperadas**: 
   - Verifique se o input está correto
   - Use o endpoint `/v1/explain` para entender a decisão

2. **Performance baixa**:
   - Ative o cache no middleware
   - Considere otimizar as políticas Rego
   - Aumente os recursos do OPA

3. **Erros de conexão com OPA**:
   - Verifique se o OPA está em execução
   - Confirme as configurações de rede e firewall
   - Verifique os logs do OPA

### Comandos Úteis

```bash
# Verificar sintaxe das políticas
opa check policies/role/src/

# Testar políticas
opa test policies/role/src/ policies/role/test/

# Ver políticas carregadas
curl http://localhost:8181/v1/policies

# Ver documentação gerada
opa eval --format pretty data.innovabiz.iam.role

# Analisar desempenho
opa eval --bench 10 --format pretty --data policies/role/src/ data.innovabiz.iam.role.allow
```

---

## Contatos para Suporte

Para suporte relacionado às políticas OPA e integração:

- **Equipe de IAM**: innovabiz-iam@innovabiz.com
- **Equipe de Segurança**: security@innovabiz.com
- **Centro de Excelência em Governança**: governance@innovabiz.com

---

Documento criado e mantido pela equipe de Identidade e Acesso da INNOVABIZ.