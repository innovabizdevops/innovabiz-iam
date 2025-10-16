package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"innovabiz/iam/identity-service/internal/application"
	"innovabiz/iam/identity-service/internal/domain/event"
	"innovabiz/iam/identity-service/internal/domain/model"
	"innovabiz/iam/identity-service/internal/domain/repository"
)

var tracer = otel.Tracer("innovabiz.iam.application.impl")

// ServiceFactory é responsável por criar e configurar todos os serviços de aplicação
type ServiceFactory struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	eventPublisher event.Publisher
	config         ServiceConfig
}

// ServiceConfig contém as configurações para todos os serviços
type ServiceConfig struct {
	// Configurações do serviço de função (role)
	Role struct {
		// Tempo máximo de cache para hierarquia de funções em segundos
		HierarchyCacheTTL time.Duration
		
		// Configuração para sincronização de funções do sistema
		SystemRoleSync struct {
			// Intervalo entre sincronizações automáticas
			Interval time.Duration
			
			// Indica se a sincronização automática está habilitada
			Enabled bool
			
			// ID do usuário para sincronizações automáticas
			ServiceUserID string
		}
	}
}

// DefaultServiceConfig retorna uma configuração padrão para os serviços
func DefaultServiceConfig() ServiceConfig {
	config := ServiceConfig{}
	
	// Configurações padrão para o serviço de função (role)
	config.Role.HierarchyCacheTTL = 5 * time.Minute
	config.Role.SystemRoleSync.Interval = 24 * time.Hour
	config.Role.SystemRoleSync.Enabled = true
	config.Role.SystemRoleSync.ServiceUserID = "00000000-0000-0000-0000-000000000000" // Sistema
	
	return config
}

// NewServiceFactory cria uma nova fábrica de serviços com as dependências necessárias
func NewServiceFactory(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	eventPublisher event.Publisher,
	config ServiceConfig,
) *ServiceFactory {
	return &ServiceFactory{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		eventPublisher: eventPublisher,
		config:         config,
	}
}

// CreateRoleService cria e configura o serviço de função (role)
func (f *ServiceFactory) CreateRoleService(ctx context.Context) application.RoleService {
	ctx, span := tracer.Start(ctx, "ServiceFactory.CreateRoleService")
	defer span.End()

	log.Info().
		Float64("hierarchyCacheTTL", f.config.Role.HierarchyCacheTTL.Seconds()).
		Bool("systemRoleSyncEnabled", f.config.Role.SystemRoleSync.Enabled).
		Dur("systemRoleSyncInterval", f.config.Role.SystemRoleSync.Interval).
		Msg("Criando serviço de função (role)")

	span.SetAttributes(
		attribute.Float64("hierarchyCacheTTL", f.config.Role.HierarchyCacheTTL.Seconds()),
		attribute.Bool("systemRoleSyncEnabled", f.config.Role.SystemRoleSync.Enabled),
	)

	service := NewRoleService(f.roleRepo, f.permissionRepo, f.eventPublisher)
	
	// Configurar serviço com as opções específicas
	service.SetHierarchyCacheTTL(f.config.Role.HierarchyCacheTTL)
	
	// Configurar sincronização automática de funções do sistema
	if f.config.Role.SystemRoleSync.Enabled {
		go f.startSystemRoleSyncScheduler(service)
	}

	return service
}

// startSystemRoleSyncScheduler inicia um agendador para sincronização periódica de funções do sistema
func (f *ServiceFactory) startSystemRoleSyncScheduler(service *RoleServiceImpl) {
	log.Info().
		Dur("interval", f.config.Role.SystemRoleSync.Interval).
		Msg("Iniciando agendador de sincronização de funções do sistema")

	ticker := time.NewTicker(f.config.Role.SystemRoleSync.Interval)
	
	go func() {
		for range ticker.C {
			ctx := context.Background()
			ctx, span := tracer.Start(ctx, "SystemRoleSync.ScheduledSync")
			
			log.Info().Msg("Executando sincronização automática de funções do sistema")
			
			// Obter definições das funções do sistema da fonte definitiva
			systemRoles := getSystemRoleDefinitions()
			
			// Sincronizar com todas as tenants ativas
			tenants, err := f.getTenantList(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Erro ao obter lista de tenants para sincronização de funções do sistema")
				span.RecordError(err)
				span.End()
				continue
			}
			
			// Converter string UUID para UUID real
			// serviceUserID, _ := uuid.Parse(f.config.Role.SystemRoleSync.ServiceUserID)
			
			for _, tenant := range tenants {
				tenantCtx, tenantSpan := tracer.Start(ctx, "SystemRoleSync.TenantSync", 
					trace.WithAttributes(attribute.String("tenant_id", tenant.String())))
				
				// Tentativa de sincronizar funções do sistema para esta tenant
				// created, updated, err := service.SyncSystemRoles(
				//	tenantCtx, tenant, systemRoles, serviceUserID)
				
				// TODO: Implementar sincronização real quando as dependências estiverem disponíveis
				// Por enquanto, apenas registramos a intenção
				log.Info().
					Str("tenant_id", tenant.String()).
					Msg("Sincronização de funções do sistema agendada para tenant")
				
				tenantSpan.End()
			}
			
			span.End()
		}
	}()
}

// getSystemRoleDefinitions obtém as definições atualizadas das funções do sistema
// Este é um placeholder e deve ser implementado com uma fonte autoritativa real
func getSystemRoleDefinitions() []model.SystemRoleDefinition {
	// TODO: Implementar obtenção de definições das funções do sistema
	// Isso poderia vir de um arquivo de configuração, banco de dados, ou API
	return []model.SystemRoleDefinition{}
}

// getTenantList obtém a lista de IDs de tenant ativos
// Este é um placeholder e deve ser implementado com acesso real à lista de tenants
func (f *ServiceFactory) getTenantList(ctx context.Context) ([]uuid.UUID, error) {
	// TODO: Implementar obtenção da lista de tenants
	// Isso poderia vir de um serviço de tenant ou banco de dados
	return []uuid.UUID{}, nil
}