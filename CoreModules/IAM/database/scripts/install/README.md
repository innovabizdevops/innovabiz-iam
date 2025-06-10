# INNOVABIZ IAM – Automação, Instalação e Integração Contínua

Este diretório contém os scripts de instalação, automação e documentação do banco de dados IAM (`innovabiz_iam`).

## Estrutura dos Scripts
- `install_all_iam_modules.sql`: Script mestre que instala todos os domínios do IAM na ordem correta.
- `00_install_extensions.sql`, `00_install_iam_module.sql`: Scripts de preparação do ambiente e schema base.
- Scripts individuais para instalação modular (core, compliance, analytics, etc).

## Pipeline de Integração Contínua (CI/CD)

### Exemplo de Workflow com GitHub Actions

```yaml
name: CI - IAM Database (Identity and Access Management)

on:
  push:
    paths:
      - 'CoreModules/IAM/database/scripts/**'

jobs:
  build-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: innovabiz_iam
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - name: Wait for Postgres
        run: until pg_isready -h localhost -p 5432; do sleep 2; done
      - name: Instala dependências de SQL
        run: sudo apt-get install -y postgresql-client
      - name: Instalação completa IAM
        run: psql -h localhost -U postgres -d innovabiz_iam -f CoreModules/IAM/database/scripts/install/install_all_iam_modules.sql
        env:
          PGPASSWORD: postgres
      - name: Executa scripts de teste e validação
        run: |
          psql -h localhost -U postgres -d innovabiz_iam -f CoreModules/IAM/database/scripts/test/run_all_tests.sql
        env:
          PGPASSWORD: postgres
      - name: Gera documentação
        run: |
          psql -h localhost -U postgres -d innovabiz_iam -f CoreModules/IAM/database/scripts/metadata/06_generate_iam_schema_documentation.sql > iam_schema_doc.md
        env:
          PGPASSWORD: postgres
```

### Adaptação para Outras Plataformas
- **GitLab CI:** Utilize `.gitlab-ci.yml` com etapas equivalentes.
- **Azure DevOps:** Adapte para `azure-pipelines.yml`.
- **Jenkins:** Use pipeline declarativo e steps shell.

## Recomendações
- Sempre execute o script mestre de instalação em ambiente de staging antes de produção.
- Utilize variáveis de ambiente seguras para senhas e conexões.
- Armazene os relatórios de documentação e compliance gerados como artefatos do pipeline.
- Mantenha este README atualizado e documente scripts customizados ou integrações específicas.

## Suporte
Dúvidas ou sugestões? Consulte a documentação técnica ou entre em contato com o responsável pelo IAM.
