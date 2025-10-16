/**
 * @file main.go
 * @description Ponto de entrada para o serviço de API do Bureau de Crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"innovabiz/iam/src/bureau-credito/api"
	"innovabiz/iam/src/bureau-credito/orchestration"
	"innovabiz/iam/src/bureau-credito/orchestration/providers"
	"innovabiz/iam/src/bureau-credito/orchestration/registry"
)

func main() {
	// Configurar logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()

	// Log inicial
	log.Info().Msg("Iniciando serviço API do Bureau de Crédito")

	// Criar contexto cancelável
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configurar orquestrador
	orchestrator, err := buildOrchestrator()
	if err != nil {
		log.Fatal().Err(err).Msg("Falha ao configurar orquestrador")
	}

	// Configurar encerramento gracioso
	setupGracefulShutdown(cancel)

	// Iniciar API
	if err := api.StartAPI(ctx, orchestrator); err != nil {
		log.Fatal().Err(err).Msg("Falha ao iniciar API")
	}

	log.Info().Msg("Serviço API do Bureau de Crédito encerrado")
}

// buildOrchestrator configura o orquestrador com provedores reais
func buildOrchestrator() (*orchestration.BureauOrchestrator, error) {
	// Obter diretório de configuração
	configDir := getConfigDir()
	log.Info().Str("configDir", configDir).Msg("Diretório de configuração")

	// Criar registro de provedores
	creditRegistry := registry.NewProviderRegistry()

	// Registrar provedores reais
	// Serasa
	serasaConfig := &providers.SerasaConfig{
		APIKey:      os.Getenv("SERASA_API_KEY"),
		APIEndpoint: os.Getenv("SERASA_API_ENDPOINT"),
		Timeout:     30,
	}
	if serasaConfig.APIEndpoint == "" {
		serasaConfig.APIEndpoint = "https://api.serasaexperian.com.br/v1"
	}
	creditRegistry.RegisterProvider("SERASA", providers.NewSerasaProvider(serasaConfig))

	// SPC Brasil
	spcConfig := &providers.SPCConfig{
		Username:    os.Getenv("SPC_USERNAME"),
		Password:    os.Getenv("SPC_PASSWORD"),
		APIEndpoint: os.Getenv("SPC_API_ENDPOINT"),
		Timeout:     30,
	}
	if spcConfig.APIEndpoint == "" {
		spcConfig.APIEndpoint = "https://api.spcbrasil.org/v1"
	}
	creditRegistry.RegisterProvider("SPC", providers.NewSPCProvider(spcConfig))

	// Outros provedores...

	// Provider mock para desenvolvimento e testes
	if os.Getenv("ENABLE_MOCK_PROVIDERS") == "true" {
		creditRegistry.RegisterProvider("MOCK", providers.NewMockProvider())
		log.Info().Msg("Provedor MOCK habilitado")
	}

	// Criar orquestrador usando configuração padrão
	return api.BuildDefaultOrchestrator()
}

// setupGracefulShutdown configura o encerramento gracioso do serviço
func setupGracefulShutdown(cancelFunc context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info().Msg("Sinal de término recebido, iniciando encerramento gracioso...")
		cancelFunc()
	}()
}

// getConfigDir retorna o diretório de configuração
func getConfigDir() string {
	// Verificar variável de ambiente
	configDir := os.Getenv("BUREAU_CONFIG_DIR")
	if configDir != "" {
		return configDir
	}

	// Usar diretório do executável
	exePath, err := os.Executable()
	if err != nil {
		log.Warn().Err(err).Msg("Não foi possível determinar o caminho do executável")
		return "config"
	}

	// Retornar subdiretório config no diretório do executável
	return filepath.Join(filepath.Dir(exePath), "config")
}