# CLI de Testes de Compliance com Remediação Automática

## Visão Geral

Esta ferramenta de linha de comando permite verificar a conformidade de políticas OPA (Open Policy Agent) com requisitos regulatórios de diferentes jurisdições. A ferramenta agora inclui capacidade de remediação automática, permitindo não apenas detectar mas também corrigir automaticamente problemas de conformidade encontrados.

## Funcionalidades

- **Teste de Conformidade:** Execute testes de conformidade contra políticas OPA definidas
- **Suporte Multi-regional:** Suporte a diferentes regiões e jurisdições
- **Filtragem por Framework:** Execute testes apenas para frameworks específicos
- **Relatórios Detalhados:** Gere relatórios em formatos JSON e HTML
- **Remediação Automática:** Aplique correções automáticas para problemas de compliance identificados
- **Controles de Segurança:** Modo dry-run, aprovação do usuário e backups automáticos

## Opções de Configuração

### Opções Básicas

- `--opa <caminho>`: Caminho para as políticas OPA (padrão: "./policies")
- `--tests <caminho>`: Caminho para os testes de compliance (padrão: "./tests/opa-compliance")
- `--output <caminho>`: Diretório para relatórios (padrão: "./reports")
- `--regions <regiões>`: Regiões a testar, separadas por vírgula (padrão: "AO")
- `--frameworks <frameworks>`: Frameworks específicos para testar, separados por vírgula
- `--tags <tags>`: Tags para filtrar testes, separadas por vírgula
- `--format <formato>`: Formato do relatório no console: table, json, html (padrão: "table")
- `--verbose`: Modo verboso com logs detalhados
- `--summary`: Exibir sumário no console (padrão: true)
- `--json`: Gerar relatório JSON (padrão: true)
- `--html`: Gerar relatório HTML (padrão: true)

### Opções de Remediação

- `--remediate`: Ativa o modo de remediação automática (padrão: false)
- `--dry-run`: Executa remediação sem aplicar mudanças reais (padrão: true)
- `--max-severity <nível>`: Severidade máxima para remediação: baixa, media, alta (padrão: "alta")
- `--min-severity <nível>`: Severidade mínima para remediação: baixa, media, alta (padrão: "baixa")
- `--ignore-types <tipos>`: Tipos de violação a ignorar, separados por vírgula
- `--rules-path <caminho>`: Caminho para regras de remediação (padrão: "remediator/rules")
- `--backup-dir <caminho>`: Diretório para backups de arquivos (padrão: "remediator/backups")
- `--no-approval`: Não solicitar aprovação antes de aplicar remediações (padrão: false)
- `--max-remed-per-policy <número>`: Número máximo de remediações por arquivo de política (padrão: 5)

## Funcionamento da Remediação Automática

O processo de remediação automática funciona da seguinte forma:

1. Os testes de compliance são executados normalmente para a região configurada.
2. Se habilitado (`--remediate`), o sistema carrega regras de remediação específicas para a região.
3. Para cada violação de compliance detectada:
   - É determinado o tipo de violação (ex: configuração incorreta, implementação incompleta)
   - São filtradas as violações com base na severidade e no tipo
   - São encontradas regras de remediação aplicáveis à violação
   - As regras são aplicadas às políticas OPA em modo simulação ou real
   - Backups são criados antes de qualquer modificação real
4. Um resumo das remediações é exibido e incluído nos relatórios.

## Arquivos de Regras de Remediação

Os arquivos de regras de remediação são arquivos JSON que definem como corrigir problemas específicos. Cada regra inclui:

- **ID**: Identificador único da regra
- **Nome e Descrição**: Detalhes sobre o que a regra faz
- **Tipo de Violação**: O tipo de problema que a regra resolve
- **ID do Teste**: O teste específico que esta regra resolve
- **Frameworks**: Frameworks para os quais esta regra é aplicável
- **Padrões de Arquivo**: Padrões para identificar quais arquivos a regra pode modificar
- **Severidade**: Nível de importância da regra
- **Código de Remediação**: O código Rego que substitui ou complementa o código existente
- **Padrão de Correspondência**: Expressão regular para encontrar a seção específica a ser modificada

## Exemplo de Uso

Para executar testes de compliance para Angola com remediação automática em modo simulação:

```bash
./compliance-test --regions AO --frameworks BNA,FINANCEIRO --remediate --dry-run
```

Para aplicar correções reais (modo perigoso, requer aprovação):

```bash
./compliance-test --regions AO --frameworks BNA --remediate --dry-run=false
```

Para aplicar remediações apenas para violações de alta severidade:

```bash
./compliance-test --regions AO --remediate --min-severity alta --max-severity alta
```

## Desenvolvimento e Extensão

Para adicionar suporte a novas regiões:

1. Crie uma matriz de conformidade em `tests/opa-compliance/regions/<CÓDIGO-REGIÃO>/compliance_matrix.json`
2. Adicione casos de teste em `testcases/<CÓDIGO-REGIÃO>/<framework>_testcases.json`
3. Crie regras de remediação em `remediator/rules/<código-região>_remediation_rules.json`

## Observações de Segurança

- Use sempre o modo `--dry-run` primeiro para verificar as alterações que serão feitas
- A ferramenta sempre criará backups antes de modificar arquivos (quando não estiver em modo dry-run)
- Por padrão, a aprovação do usuário é solicitada antes de aplicar modificações reais
- Recomenda-se revisar os arquivos alterados após a remediação automática
- Utilize controle de versão para suas políticas para facilitar a reversão de alterações quando necessário