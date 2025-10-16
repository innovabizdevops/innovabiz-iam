/**
 * @file main.go
 * @description Ponto de entrada principal para a API do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package api

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	
	"innovabiz/iam/src/bureau-credito/orchestration"
	"innovabiz/iam/src/bureau-credito/orchestration/cache"
	"innovabiz/iam/src/bureau-credito/orchestration/registry"
)

// Variáveis de configuração a partir de flags
var (
	configPath  = flag.String("config", "", "Caminho para arquivo de configuração JSON")
	port        = flag.Int("port", 8080, "Porta do servidor")
	logLevel    = flag.String("log-level", "info", "Nível de log (debug, info, warn, error)")
	enableGraph = flag.Bool("graphql", true, "Habilitar API GraphQL")
	enableCORS  = flag.Bool("cors", true, "Habilitar CORS")
)

// StartAPI inicia o servidor API do Bureau de Crédito
func StartAPI(ctx context.Context, orchestratorInstance *orchestration.BureauOrchestrator) error {
	// Processar flags
	flag.Parse()
	
	// Carregar configuração
	config := loadConfig(*configPath)
	
	// Sobrescrever configuração com flags, se fornecidas
	if *port != 0 {
		config.Port = *port
	}
	if *logLevel != "" {
		config.LogLevel = *logLevel
	}
	if *enableGraph {
		config.EnableGraphQL = *enableGraph
	}
	if *enableCORS {
		config.EnableCORS = *enableCORS
	}
	
	// Criar servidor
	server := NewAPIServer(orchestratorInstance, config)
	
	// Configurar rotas
	server.SetupRoutes()
	
	// Iniciar servidor
	return server.Start()
}

// loadConfig carrega a configuração a partir de um arquivo
func loadConfig(path string) *Config {
	// Se caminho não fornecido, retornar configuração padrão
	if path == "" {
		return DefaultConfig()
	}
	
	// Verificar se arquivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Warn().Str("path", path).Msg("Arquivo de configuração não encontrado, usando padrão")
		return DefaultConfig()
	}
	
	// Abrir arquivo
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Erro ao abrir arquivo de configuração")
		return DefaultConfig()
	}
	defer file.Close()
	
	// Decodificar JSON
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Erro ao decodificar arquivo de configuração")
		return DefaultConfig()
	}
	
	return &config
}

// BuildDefaultOrchestrator cria uma instância padrão do orquestrador
func BuildDefaultOrchestrator() (*orchestration.BureauOrchestrator, error) {
	// Criar registro de provedores de crédito
	creditRegistry := registry.NewProviderRegistry()
	
	// Registrar provedores padrão (mock)
	creditRegistry.RegisterProvider("MOCK", registry.NewMockProvider())
	
	// Criar cache com TTL de 1 hora
	assessmentCache := cache.NewInMemoryCache(1 * time.Hour)
	
	// Criar orquestrador com configuração básica
	orchestratorOptions := &orchestration.OrchestratorOptions{
		ConcurrencyLimit: 100,
		DefaultTimeout:   30 * time.Second,
		EnableCache:      true,
		CacheTTL:         1 * time.Hour,
	}
	
	return orchestration.NewBureauOrchestrator(creditRegistry, assessmentCache, orchestratorOptions)
}

// GetApplicationPath retorna o diretório da aplicação
func GetApplicationPath() string {
	// Obter diretório do executável
	exePath, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("Não foi possível determinar o caminho do executável")
		return "."
	}
	
	// Retornar diretório pai
	return filepath.Dir(exePath)
}