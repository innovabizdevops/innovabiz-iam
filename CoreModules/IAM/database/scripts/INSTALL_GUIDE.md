# Guia de Instalação do Módulo IAM

## Visão Geral
Este guia fornece instruções detalhadas para instalar o módulo de Identity and Access Management (IAM) no banco de dados PostgreSQL.

## Pré-requisitos

- PostgreSQL 13 ou superior
- Acesso de administrador ao banco de dados
- PowerShell 5.1 ou superior (para execução do script de instalação)

## 1. Preparação do Ambiente

1. Certifique-se de que o PostgreSQL está instalado e em execução
2. Verifique se o banco de dados `innovabiz_iam` existe:
   ```sql
   SELECT 1 FROM pg_database WHERE datname = 'innovabiz_iam';
   ```
3. Se o banco de dados não existir, crie-o:
   ```sql
   CREATE DATABASE innovabiz_iam 
   WITH ENCODING='UTF8' 
   LC_COLLATE='pt_BR.UTF-8' 
   LC_CTYPE='pt_BR.UTF-8' 
   TEMPLATE=template0;
   ```

## 2. Instalação do Módulo

### Método 1: Usando o script PowerShell (Recomendado)

1. Navegue até o diretório `CoreModules/IAM/database/scripts/` do seu projeto.
   Por exemplo:
   ```powershell
   cd C:\Path\To\YourProject\InnovaBiz\CoreModules\IAM\database\scripts
   ```

2. Execute o script de instalação a partir deste diretório:
   ```powershell
   .\install_iam_module.ps1
   ```

3. O script irá:
   - Verificar a conexão com o banco de dados
   - Executar todos os scripts SQL na ordem correta
   - Verificar se a instalação foi bem-sucedida

### Método 2: Execução Manual dos Scripts SQL

Se preferir executar os scripts manualmente, siga esta ordem:

1. Conecte-se ao banco de dados `innovabiz_iam`
2. Execute os scripts da subpasta `install/` na seguinte ordem:

   ```
   install/00_install_extensions.sql
   install/00_install_iam_module.sql
   install/01_install_iam_core.sql
   install/02_install_iam_auth.sql
   install/03_install_iam_authz.sql
   ```

   **Nota:** O script `install/01_install_iam_core.sql` por sua vez executará os scripts da pasta `core/` (`01_schema_iam_core.sql`, `02_views_iam_core.sql`, etc.).

## 3. Pós-Instalação

Após a instalação, verifique se todas as tabelas foram criadas corretamente:

```sql
\dt iam.*
```

Você deve ver uma lista de tabelas do módulo IAM, incluindo:
- `iam.organizations`
- `iam.users`
- `iam.roles`
- `iam.role_permissions`
- `iam.permissions`
- `iam.sessions`
- E outras tabelas relacionadas

## 4. Configuração Inicial

1. Crie uma organização inicial:
   ```sql
   INSERT INTO iam.organizations (name, code, industry, sector, country_code, region_code)
   VALUES ('Organização de Exemplo', 'ORG001', 'Tecnologia', 'Software', 'BR', 'Sudeste');
   ```

2. Crie um usuário administrador:
   ```sql
   -- Substitua 'senha123' por uma senha forte
   INSERT INTO iam.users (organization_id, username, email, full_name, password_hash)
   VALUES (
       (SELECT id FROM iam.organizations WHERE code = 'ORG001'),
       'admin',
       'admin@exemplo.com',
       'Administrador do Sistema',
       crypt('senha123', gen_salt('bf'))
   );
   ```

## 5. Solução de Problemas

### Erro de Conexão
- Verifique se o PostgreSQL está em execução
- Confirme as credenciais de acesso
- Verifique se o banco de dados existe

### Erros de Permissão
- Certifique-se de que o usuário tem permissões suficientes
- Se necessário, execute os scripts como superusuário

### Problemas com Scripts
- Verifique se todos os arquivos SQL estão no local correto
- Verifique se não há erros de sintaxe nos scripts
- Consulte o log de instalação para mensagens de erro detalhadas

## 6. Atualização

Para atualizar o módulo IAM para uma versão mais recente, siga estas etapas:

1. Faça backup do banco de dados
2. Execute os scripts de atualização na pasta `updates/` em ordem numérica
3. Verifique se todas as atualizações foram aplicadas com sucesso

## Suporte

Para obter ajuda, entre em contato com a equipe de suporte:
- E-mail: innovabizdevops@gmail.com
- Responsável: Eduardo Jeremias

---

**Nota de Direitos Autorais**  
© 2025 InnovaBiz. Todos os direitos reservados.
