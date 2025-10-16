# Makefile para validação e testes de políticas OPA do RoleService
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53, PCI DSS v4.0

# Variáveis
OPA_BIN := opa
POLICY_DIR := .
TEST_DIR := .
POLICY_FILES := $(wildcard $(POLICY_DIR)/*.rego)
TEST_FILES := $(wildcard $(TEST_DIR)/role_test.rego)
COVERAGE_DIR := ./coverage
BENCHMARK_DIR := ./benchmark

# Cores para saída no terminal
YELLOW := \033[1;33m
GREEN := \033[1;32m
RED := \033[1;31m
BLUE := \033[1;34m
RESET := \033[0m

# Configurações de formato
TIME_FORMAT := %Y-%m-%d_%H-%M-%S

# Alvos principais
.PHONY: all clean validate-policies test-policies test-policy policy-coverage benchmark-policies format-policies doc-policies check-vulnerabilities

all: validate-policies test-policies

# Validar sintaxe de todas as políticas
validate-policies:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Validando políticas Rego (OPA)...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@for policy in $(POLICY_FILES); do \
		echo "$(YELLOW)Validando $${policy}...$(RESET)"; \
		$(OPA_BIN) check $${policy} && echo "$(GREEN)OK: $${policy}$(RESET)" || { echo "$(RED)FALHA: $${policy}$(RESET)"; exit 1; }; \
	done

# Executar todos os testes unitários
test-policies:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Executando testes das políticas...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@$(OPA_BIN) test $(POLICY_DIR) --verbose

# Executar um teste específico
test-policy:
ifndef POLICY
	@echo "$(RED)Erro: Especifique o arquivo de política para testar com POLICY=nome_do_arquivo.rego$(RESET)"
else
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Executando testes para $(POLICY)...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@$(OPA_BIN) test $(POLICY_DIR) --run $(POLICY) --verbose
endif

# Gerar relatório de cobertura de testes
policy-coverage:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Gerando relatório de cobertura...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@timestamp=$$(date +$(TIME_FORMAT)); \
	$(OPA_BIN) test $(POLICY_DIR) --coverage --format=json > $(COVERAGE_DIR)/coverage_$${timestamp}.json && \
	echo "$(GREEN)Relatório de cobertura gerado em $(COVERAGE_DIR)/coverage_$${timestamp}.json$(RESET)"

# Benchmark de performance das políticas
benchmark-policies:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Executando benchmarks de performance...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@mkdir -p $(BENCHMARK_DIR)
	@timestamp=$$(date +$(TIME_FORMAT)); \
	$(OPA_BIN) test $(POLICY_DIR) --bench --count=10000 --format=json > $(BENCHMARK_DIR)/benchmark_$${timestamp}.json && \
	echo "$(GREEN)Relatório de benchmark gerado em $(BENCHMARK_DIR)/benchmark_$${timestamp}.json$(RESET)"

# Formatar código Rego
format-policies:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Formatando políticas Rego...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@for policy in $(POLICY_FILES); do \
		echo "$(YELLOW)Formatando $${policy}...$(RESET)"; \
		$(OPA_BIN) fmt -w $${policy}; \
	done
	@echo "$(GREEN)Todas as políticas foram formatadas.$(RESET)"

# Gerar documentação a partir das políticas
doc-policies:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Gerando documentação das políticas...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@mkdir -p ./docs
	@for policy in $(POLICY_FILES); do \
		filename=$$(basename $${policy} .rego); \
		echo "$(YELLOW)Gerando documentação para $${filename}...$(RESET)"; \
		$(OPA_BIN) inspect $${policy} --format=markdown > ./docs/$${filename}_doc.md; \
	done
	@echo "$(GREEN)Documentação gerada em ./docs/$(RESET)"

# Verificação de vulnerabilidades nas políticas
check-vulnerabilities:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Verificando vulnerabilidades nas políticas...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(YELLOW)Procurando por regras que permitam acesso irrestrito...$(RESET)"
	@grep -r "allow { true }" $(POLICY_DIR) || echo "$(GREEN)Nenhuma regra de permissão irrestrita encontrada.$(RESET)"
	@echo "$(YELLOW)Procurando por negações explícitas ausentes...$(RESET)"
	@for policy in $(POLICY_FILES); do \
		if ! grep -q "allow {" $$policy || grep -q "allow = false" $$policy || grep -q "deny {" $$policy; then \
			echo "$(GREEN)OK: $${policy} tem regras de negação adequadas.$(RESET)"; \
		else \
			echo "$(RED)ATENÇÃO: $${policy} pode não ter regras de negação explícitas.$(RESET)"; \
		fi \
	done

# Verificar conformidade com políticas corporativas
check-compliance:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Verificando conformidade com políticas corporativas...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(YELLOW)Verificando conformidade com ISO/IEC 27001:2022...$(RESET)"
	@grep -q "ISO/IEC 27001:2022" $(POLICY_FILES) && echo "$(GREEN)OK: Referência à ISO/IEC 27001:2022 encontrada.$(RESET)" || echo "$(RED)FALHA: Referência à ISO/IEC 27001:2022 não encontrada.$(RESET)"
	@echo "$(YELLOW)Verificando mecanismos de auditoria...$(RESET)"
	@grep -q "audit" $(POLICY_FILES) && echo "$(GREEN)OK: Referências a auditoria encontradas.$(RESET)" || echo "$(RED)FALHA: Referências a auditoria não encontradas.$(RESET)"
	@echo "$(YELLOW)Verificando mecanismos de multitenancy...$(RESET)"
	@grep -q "tenant" $(POLICY_FILES) && echo "$(GREEN)OK: Referências a tenant encontradas.$(RESET)" || echo "$(RED)FALHA: Referências a tenant não encontradas.$(RESET)"

# Simular decisões de autorização com dados de exemplo
simulate-decision:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Simulando decisões de autorização...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
ifndef POLICY
	@echo "$(RED)Erro: Especifique o arquivo de política para simular com POLICY=nome_do_arquivo.rego$(RESET)"
else ifndef INPUT
	@echo "$(RED)Erro: Especifique o arquivo de entrada JSON para simulação com INPUT=input.json$(RESET)"
else
	@echo "$(YELLOW)Simulando decisão para $(POLICY) com entrada $(INPUT)...$(RESET)"
	@$(OPA_BIN) eval --data $(POLICY_DIR) --input $(INPUT) "data.innovabiz.iam.role" --format=pretty
endif

# Limpar arquivos temporários e gerados
clean:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Limpando arquivos temporários...$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@rm -rf $(COVERAGE_DIR) $(BENCHMARK_DIR) ./docs
	@echo "$(GREEN)Arquivos temporários removidos.$(RESET)"

# Ajuda
help:
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(BLUE)Ajuda do Makefile para políticas OPA$(RESET)"
	@echo "$(BLUE)======================================$(RESET)"
	@echo "$(YELLOW)Comandos disponíveis:$(RESET)"
	@echo "  $(GREEN)make validate-policies$(RESET) - Valida a sintaxe de todas as políticas Rego"
	@echo "  $(GREEN)make test-policies$(RESET) - Executa todos os testes unitários"
	@echo "  $(GREEN)make test-policy POLICY=arquivo.rego$(RESET) - Executa testes de uma política específica"
	@echo "  $(GREEN)make policy-coverage$(RESET) - Gera relatório de cobertura de testes"
	@echo "  $(GREEN)make benchmark-policies$(RESET) - Executa benchmarks de performance"
	@echo "  $(GREEN)make format-policies$(RESET) - Formata o código Rego"
	@echo "  $(GREEN)make doc-policies$(RESET) - Gera documentação a partir das políticas"
	@echo "  $(GREEN)make check-vulnerabilities$(RESET) - Verifica vulnerabilidades nas políticas"
	@echo "  $(GREEN)make check-compliance$(RESET) - Verifica conformidade com políticas corporativas"
	@echo "  $(GREEN)make simulate-decision POLICY=arquivo.rego INPUT=input.json$(RESET) - Simula decisão de autorização"
	@echo "  $(GREEN)make clean$(RESET) - Remove arquivos temporários e gerados"
	@echo "$(BLUE)======================================$(RESET)"