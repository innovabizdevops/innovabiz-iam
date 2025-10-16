# Makefile para o serviço IAM da plataforma INNOVABIZ
# Versão: 1.0
# Copyright © 2025 INNOVABIZ

# Configurações de ambiente
GO_CMD=go
GOTEST=$(GO_CMD) test
GOVET=$(GO_CMD) vet
GOLINT=golangci-lint
GOFMT=$(GO_CMD) fmt

# Diretórios
API_DIR=./cmd/api
INTERNAL_DIR=./internal
INTERFACE_DIR=$(INTERNAL_DIR)/interface
HANDLER_DIR=$(INTERFACE_DIR)/api/handler
MIDDLEWARE_DIR=$(INTERFACE_DIR)/middleware
TEST_DIR=$(HANDLER_DIR)/tests
POLICY_DIR=./policies

# Configuração de cobertura
COVERAGE_PROFILE=coverage.out
COVERAGE_HTML=coverage.html
COVERAGE_THRESHOLD=80

# Configuração de logs para testes
TEST_VERBOSE=-v
TEST_DEBUG_ENV=TEST_LOG_LEVEL=debug

# Alvos padrão
.PHONY: all test clean fmt lint vet coverage run help test-handler test-middleware test-integration security-scan

# Alvo padrão
all: lint test coverage

# Executar todos os testes
test:
	@echo "Executando todos os testes..."
	@$(GOTEST) $(TEST_VERBOSE) ./...

# Executar testes com logs detalhados
test-debug:
	@echo "Executando testes com logs detalhados..."
	@$(TEST_DEBUG_ENV) $(GOTEST) $(TEST_VERBOSE) ./...

# Executar apenas testes do RoleHandler
test-handler:
	@echo "Executando testes do RoleHandler..."
	@$(GOTEST) $(TEST_VERBOSE) $(TEST_DIR)

# Executar testes de middleware
test-middleware:
	@echo "Executando testes dos middlewares..."
	@$(GOTEST) $(TEST_VERBOSE) $(MIDDLEWARE_DIR)

# Executar testes de integração (simulando requisições completas)
test-integration:
	@echo "Executando testes de integração..."
	@$(GOTEST) $(TEST_VERBOSE) -tags=integration ./...

# Verificar cobertura de testes
coverage:
	@echo "Gerando relatório de cobertura..."
	@$(GOTEST) -coverprofile=$(COVERAGE_PROFILE) ./...
	@$(GO_CMD) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	@$(GO_CMD) tool cover -func=$(COVERAGE_PROFILE) | grep total | awk '{print $$3}' | sed 's/%//' > coverage_percent.txt
	@if [ `cat coverage_percent.txt` -lt $(COVERAGE_THRESHOLD) ]; then \
		echo "Cobertura de código abaixo do limite! Atual: `cat coverage_percent.txt`%, Esperado: $(COVERAGE_THRESHOLD)%"; \
		rm coverage_percent.txt; \
		exit 1; \
	else \
		echo "Cobertura de código: `cat coverage_percent.txt`% (limite: $(COVERAGE_THRESHOLD)%)"; \
		rm coverage_percent.txt; \
	fi

# Verificar formatação de código
fmt:
	@echo "Verificando formatação de código..."
	@$(GOFMT) ./...

# Executar linter para verificar qualidade de código
lint:
	@echo "Verificando qualidade de código..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint não instalado. Instalando..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
		$(GOLINT) run ./...; \
	fi

# Executar go vet para encontrar possíveis erros
vet:
	@echo "Executando go vet..."
	@$(GOVET) ./...

# Verificar políticas OPA (requer opa CLI)
check-policies:
	@echo "Verificando políticas OPA..."
	@if command -v opa >/dev/null 2>&1; then \
		opa check $(POLICY_DIR)/*.rego; \
	else \
		echo "OPA CLI não instalado. Para instalar, visite: https://www.openpolicyagent.org/docs/latest/get-started/"; \
	fi

# Análise de segurança
security-scan:
	@echo "Executando análise de segurança..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -exclude-generated ./...; \
	else \
		echo "gosec não instalado. Instalando..."; \
		$(GO_CMD) install github.com/securego/gosec/v2/cmd/gosec@latest; \
		gosec -exclude-generated ./...; \
	fi

# Executar o servidor da API
run:
	@echo "Iniciando servidor API..."
	@cd $(API_DIR) && $(GO_CMD) run main.go

# Limpar arquivos temporários
clean:
	@echo "Limpando arquivos temporários..."
	@rm -f $(COVERAGE_PROFILE) $(COVERAGE_HTML)
	@find . -type f -name "*.out" -delete

# Exibir ajuda
help:
	@echo "Alvos disponíveis:"
	@echo "  all            : Executa lint e testes com cobertura"
	@echo "  test           : Executa todos os testes"
	@echo "  test-debug     : Executa testes com logs detalhados"
	@echo "  test-handler   : Executa apenas testes do RoleHandler"
	@echo "  test-middleware: Executa testes dos middlewares"
	@echo "  test-integration: Executa testes de integração"
	@echo "  coverage       : Gera relatório de cobertura de código"
	@echo "  fmt            : Verifica formatação de código"
	@echo "  lint           : Executa linter para verificar qualidade de código"
	@echo "  vet            : Executa go vet para encontrar possíveis erros"
	@echo "  check-policies : Verifica políticas OPA (requer CLI OPA)"
	@echo "  security-scan  : Executa análise de segurança (requer gosec)"
	@echo "  run            : Executa o servidor da API"
	@echo "  clean          : Remove arquivos temporários"
	@echo "  help           : Exibe esta mensagem de ajuda"