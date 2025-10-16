# Dockerfile para CoreModules/IAM
FROM python:3.11-slim

WORKDIR /app

# Copia o código-fonte do IAM
COPY . /app

# Instala dependências (ajuste requirements.txt se necessário)
RUN pip install --no-cache-dir -r requirements.txt

# Exemplo de variáveis de ambiente (ajuste conforme o módulo)
ENV IAM_ENV=production
ENV IAM_DB_HOST=postgres
ENV IAM_DB_USER=iam_user
ENV IAM_DB_PASS=iam_pass

# Exponha a porta padrão do serviço IAM (ajuste se necessário)
EXPOSE 8080

# Comando de inicialização (ajuste conforme o entrypoint real)
CMD ["python", "main.py"]
