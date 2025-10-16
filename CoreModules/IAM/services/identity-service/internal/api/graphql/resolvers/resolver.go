package resolvers

import (
	"github.com/innovabiz/iam/internal/application/services"
	"github.com/innovabiz/iam/internal/domain/repositories"
	"github.com/innovabiz/iam/internal/infrastructure/auth"
	"github.com/innovabiz/iam/internal/infrastructure/observability"
	"go.opentelemetry.io/otel/trace"
)

// Este arquivo contém a estrutura base do resolver GraphQL para o módulo IAM
// e será gerado o código complementar pelo gqlgen

// Resolver é a estrutura base para todos os resolvers do GraphQL
type Resolver struct {
	// Serviços da camada de aplicação
	userService        services.UserService
	groupService       services.GroupService
	roleService        services.RoleService
	tenantService      services.TenantService
	authService        services.AuthService
	permissionService  services.PermissionService
	securityService    services.SecurityService
	
	// Repositórios para acesso direto quando necessário
	userRepository     repositories.UserRepository
	groupRepository    repositories.GroupRepository
	roleRepository     repositories.RoleRepository
	tenantRepository   repositories.TenantRepository
	permissionRepository repositories.PermissionRepository
	
	// Componentes de infraestrutura
	tracer             trace.Tracer
	logger             observability.Logger
	authManager        *auth.Manager
	
	// Configuração
	config             *ResolverConfig
}

// ResolverConfig contém as configurações do resolver
type ResolverConfig struct {
	EnableIntrospection      bool
	MaxQueryComplexity       int
	MaxQueryDepth            int
	MaxPageSize              int
	DefaultPageSize          int
	EnableSubscriptions      bool
	EnableCrossTenantAccess  bool
	EnableAuditLogging       bool
	DataClassificationLevels map[string]int // Níveis de classificação de dados para controle de acesso
}

// NewResolver cria uma nova instância do resolver principal
func NewResolver(
	userService services.UserService,
	groupService services.GroupService,
	roleService services.RoleService,
	tenantService services.TenantService,
	authService services.AuthService,
	permissionService services.PermissionService,
	securityService services.SecurityService,
	userRepository repositories.UserRepository,
	groupRepository repositories.GroupRepository,
	roleRepository repositories.RoleRepository,
	tenantRepository repositories.TenantRepository,
	permissionRepository repositories.PermissionRepository,
	tracer trace.Tracer,
	logger observability.Logger,
	authManager *auth.Manager,
	config *ResolverConfig,
) *Resolver {
	return &Resolver{
		userService:        userService,
		groupService:       groupService,
		roleService:        roleService,
		tenantService:      tenantService,
		authService:        authService,
		permissionService:  permissionService,
		securityService:    securityService,
		userRepository:     userRepository,
		groupRepository:    groupRepository,
		roleRepository:     roleRepository,
		tenantRepository:   tenantRepository,
		permissionRepository: permissionRepository,
		tracer:             tracer,
		logger:             logger,
		authManager:        authManager,
		config:             config,
	}
}

// DefaultConfig retorna uma configuração padrão para o resolver
func DefaultConfig() *ResolverConfig {
	return &ResolverConfig{
		EnableIntrospection:     true,
		MaxQueryComplexity:      200,
		MaxQueryDepth:           10,
		MaxPageSize:             100,
		DefaultPageSize:         20,
		EnableSubscriptions:     true,
		EnableCrossTenantAccess: false, // Por padrão, desabilitado por segurança
		EnableAuditLogging:      true,
		DataClassificationLevels: map[string]int{
			"PUBLIC":       1,
			"INTERNAL":     2,
			"CONFIDENTIAL": 3,
			"RESTRICTED":   4,
		},
	}
}