# 🔄 Estados do Módulo Core IAM

## 📊 **ESTADOS DO SISTEMA**

### **Estados de Identidade**
| Estado | Código | Descrição | Ações Permitidas | Próximos Estados |
|--------|--------|-----------|-----------------|------------------|
| **Pending** | IDENT_PENDING | Identidade criada, aguardando ativação | Verificar email, cancelar | Active, Cancelled |
| **Active** | IDENT_ACTIVE | Identidade ativa e operacional | Todas as ações | Suspended, Locked, Inactive |
| **Suspended** | IDENT_SUSPENDED | Temporariamente suspensa | Visualizar, reativar | Active, Terminated |
| **Locked** | IDENT_LOCKED | Bloqueada por segurança | Desbloquear com verificação | Active, Terminated |
| **Inactive** | IDENT_INACTIVE | Inativa por período | Reativar | Active, Archived |
| **Archived** | IDENT_ARCHIVED | Arquivada para histórico | Consultar logs | None |
| **Terminated** | IDENT_TERMINATED | Permanentemente desativada | Nenhuma | None |

### **Estados de Autenticação**
| Estado | Código | Descrição | Timeout | Ação em Falha |
|--------|--------|-----------|---------|---------------|
| **Initiating** | AUTH_INIT | Processo iniciado | 30s | Timeout error |
| **Challenging** | AUTH_CHALLENGE | Desafio enviado | 300s | Retry ou Block |
| **Verifying** | AUTH_VERIFY | Verificando credenciais | 10s | Error response |
| **MFA_Required** | AUTH_MFA | Aguardando segundo fator | 120s | Fallback method |
| **Success** | AUTH_SUCCESS | Autenticação bem-sucedida | N/A | Token issued |
| **Failed** | AUTH_FAILED | Falha na autenticação | N/A | Increment counter |
| **Blocked** | AUTH_BLOCKED | Bloqueado por tentativas | 3600s | Admin review |