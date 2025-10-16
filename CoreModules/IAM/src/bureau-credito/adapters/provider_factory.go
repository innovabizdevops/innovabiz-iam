/**
 * @file provider_factory.go
 * @description Fábrica de adaptadores para provedores de crédito
 * @author InnovaBiz DevOps Team
 * @copyright InnovaBiz 2025
 */

package adapters

import (
	"errors"
	"fmt"
	"sync"
)

// DefaultCreditProviderFactory é a implementação padrão da fábrica de provedores
type DefaultCreditProviderFactory struct {
	registeredProviders map[string]func() CreditProvider
	mutex               sync.RWMutex
}

// NewCreditProviderFactory cria uma nova instância da fábrica de provedores
func NewCreditProviderFactory() *DefaultCreditProviderFactory {
	factory := &DefaultCreditProviderFactory{
		registeredProviders: make(map[string]func() CreditProvider),
	}
	
	// Registrar provedores padrão
	factory.registerDefaultProviders()
	
	return factory
}

// RegisterProvider registra um novo tipo de provedor na fábrica
func (f *DefaultCreditProviderFactory) RegisterProvider(
	providerType string,
	createFn func() CreditProvider,
) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if _, exists := f.registeredProviders[providerType]; exists {
		return fmt.Errorf("provedor do tipo '%s' já está registrado", providerType)
	}
	
	f.registeredProviders[providerType] = createFn
	return nil
}

// CreateProvider cria uma nova instância de um provedor com as configurações fornecidas
func (f *DefaultCreditProviderFactory) CreateProvider(
	providerType string,
	config CreditProviderConfig,
) (CreditProvider, error) {
	f.mutex.RLock()
	createFn, exists := f.registeredProviders[providerType]
	f.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("provedor do tipo '%s' não encontrado", providerType)
	}
	
	provider := createFn()
	if err := provider.Initialize(config); err != nil {
		return nil, fmt.Errorf("erro ao inicializar provedor '%s': %w", providerType, err)
	}
	
	return provider, nil
}

// ListAvailableProviders lista todos os tipos de provedores disponíveis
func (f *DefaultCreditProviderFactory) ListAvailableProviders() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	providers := make([]string, 0, len(f.registeredProviders))
	for providerType := range f.registeredProviders {
		providers = append(providers, providerType)
	}
	
	return providers
}

// registerDefaultProviders registra os provedores padrão do sistema
func (f *DefaultCreditProviderFactory) registerDefaultProviders() {
	// Registrar Serasa
	f.registeredProviders["serasa"] = func() CreditProvider {
		return NewSerasaAdapter()
	}
	
	// Aqui seriam registrados outros provedores como SPC, Boa Vista, etc.
	// Exemplo:
	// f.registeredProviders["spc"] = func() CreditProvider {
	//     return NewSPCAdapter()
	// }
}

// CreateProviderWithConfig cria e configura um provedor em uma única chamada
func CreateProviderWithConfig(providerType string, config CreditProviderConfig) (CreditProvider, error) {
	factory := NewCreditProviderFactory()
	return factory.CreateProvider(providerType, config)
}