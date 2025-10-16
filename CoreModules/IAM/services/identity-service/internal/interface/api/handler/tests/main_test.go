package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TestMain é a função principal executada pelo framework de testes
// Ela configura o ambiente de teste antes da execução e realiza limpeza depois
func TestMain(m *testing.M) {
	// Configuração inicial para todos os testes
	setup()
	
	// Executa todos os testes
	exitVal := m.Run()
	
	// Limpeza após testes
	teardown()
	
	// Encerra com o código retornado pelos testes
	os.Exit(exitVal)
}

// setup configura o ambiente de teste
func setup() {
	// Configurar o logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("TEST_LOG_LEVEL") == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// No modo normal de testes, suprimir logs
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	// Formato de timestamp consistente para logs
	zerolog.TimeFieldFormat = time.RFC3339Nano
	
	// Se estiver executando em CI, usar formato JSON para logs
	if os.Getenv("CI") != "" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Em ambiente local, usar formato mais legível
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}).
			With().Timestamp().Logger()
	}
	
	log.Info().Msg("Configurando ambiente de teste para RoleHandler")
	
	// Outras configurações podem ser adicionadas aqui conforme necessário
}

// teardown limpa o ambiente após os testes
func teardown() {
	log.Info().Msg("Limpando ambiente de teste para RoleHandler")
	
	// Cleanup de recursos, se necessário
}

// TestPackageDocumentation é um teste artificial que documenta o pacote
// Serve como ponto de entrada para documentação e metadados
func TestPackageDocumentation(t *testing.T) {
	doc := `
Package tests fornece testes unitários abrangentes para o RoleHandler da API HTTP do IAM INNOVABIZ.

Conformidade e Padrões:
- ISO/IEC 27001: Segurança da Informação
- TOGAF: Arquitetura Empresarial
- COBIT: Governança de TI
- PCI DSS: Segurança de Dados de Pagamento
- NIST Cybersecurity Framework

Os testes cobrem:
1. Operações CRUD básicas do RoleHandler
2. Gerenciamento de permissões
3. Gerenciamento de hierarquia de funções
4. Gerenciamento de associações usuário-função
5. Integração com middlewares:
   - Autenticação JWT
   - Autorização baseada em políticas ABAC (OPA)
   - CORS e segurança de cabeçalhos
6. Validação de entrada e tratamento de erros
7. Escopo de tenant e isolamento multitenancy
8. Auditoria e rastreamento de operações

Execução:
Para executar todos os testes:
  go test -v ./...

Para executar com logs detalhados:
  TEST_LOG_LEVEL=debug go test -v ./...

Criado como parte do módulo IAM da plataforma INNOVABIZ.
`
	fmt.Println(doc)
	// Teste sempre passa, é apenas informativo
}