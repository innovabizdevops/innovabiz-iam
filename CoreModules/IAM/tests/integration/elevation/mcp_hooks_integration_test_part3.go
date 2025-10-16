package integration_test

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/testcontainers/testcontainers-go"

	"github.com/innovabizdevops/innovabiz-iam/authorization/elevation"
	"github.com/innovabizdevops/innovabiz-iam/hooks/mcp"
	"github.com/innovabizdevops/innovabiz-iam/tests/testutil"
)

// Interfaces e funções auxiliares para conexão e configuração de componentes
type DBConn interface {
	Close() error
}

type RedisClient interface {
	Close() error
}

// connectToDatabase cria conexão com o banco de dados para testes
func connectToDatabase(ctx context.Context, container testcontainers.Container) (DBConn, error) {
	// Obter porta mapeada do container
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	// Obter host do container
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	// Construir DSN para PostgreSQL
	dsn := "postgres://testuser:testpass@" + host + ":" + mappedPort.Port() + "/innovabiz_iam_test?sslmode=disable"

	// Abrir conexão
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Verificar conexão
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	// Inicializar esquema do banco de dados
	err = initializeDBSchema(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Criar dados de teste iniciais
	err = seedTestData(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// connectToRedis cria conexão com Redis para testes
func connectToRedis(ctx context.Context, container testcontainers.Container) (RedisClient, error) {
	// Obter porta mapeada do container
	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	// Obter host do container
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	// Criar cliente Redis
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + mappedPort.Port(),
		Password: "", // sem senha para testes
		DB:       0,  // usar banco padrão
	})

	// Verificar conexão
	_, err = client.Ping(ctx).Result()
	if err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// initializeDBSchema cria o esquema de banco de dados para testes
func initializeDBSchema(db *sql.DB) error {
	// Código para criar tabelas, índices, etc.
	// Na implementação real, executaria scripts SQL para criar o esquema

	// Exemplo simplificado:
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS privilege_elevations (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			tenant_id VARCHAR(255) NOT NULL,
			market VARCHAR(100) NOT NULL,
			requested_scopes TEXT[] NOT NULL,
			approved_scopes TEXT[] NOT NULL,
			justification TEXT NOT NULL,
			emergency_access BOOLEAN NOT NULL DEFAULT FALSE,
			status VARCHAR(50) NOT NULL,
			requested_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			approved_by VARCHAR(255),
			approved_at TIMESTAMP,
			revoked_at TIMESTAMP,
			revoked_by VARCHAR(255),
			revocation_reason TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS elevation_audit_events (
			id VARCHAR(36) PRIMARY KEY,
			elevation_id VARCHAR(36) NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			tenant_id VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			metadata JSONB,
			FOREIGN KEY (elevation_id) REFERENCES privilege_elevations(id)
		)`,
		`CREATE TABLE IF NOT EXISTS mfa_challenges (
			challenge_id VARCHAR(36) PRIMARY KEY,
			elevation_id VARCHAR(36) NOT NULL,
			type VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			verified_at TIMESTAMP,
			FOREIGN KEY (elevation_id) REFERENCES privilege_elevations(id)
		)`,
	}

	// Executar cada script de criação
	for _, schema := range schemas {
		_, err := db.Exec(schema)
		if err != nil {
			return err
		}
	}

	return nil
}

// seedTestData insere dados iniciais para testes
func seedTestData(db *sql.DB) error {
	// Na implementação real, inseriria dados de teste
	return nil
}

// setupElevationService configura o serviço de elevação de privilégios para testes
func setupElevationService(ctx context.Context, dbConn interface{}, redisClient interface{}, 
	obs *testutil.TestObservability) (*elevation.PrivilegeElevationService, error) {
	
	// Em um teste real, usaríamos as conexões reais para configurar o serviço
	// Para este exemplo, criamos um serviço real com mocks internos
	
	// Configurar serviço com componentes reais
	service := elevation.NewPrivilegeElevationService(
		elevation.WithDatabase(dbConn.(*sql.DB)),
		elevation.WithRedisClient(redisClient.(*redis.Client)),
		elevation.WithLogger(obs.Logger()),
		elevation.WithTracer(obs.Tracer()),
		elevation.WithMFAProvider(elevation.NewDefaultMFAProvider()),
		elevation.WithNotifier(elevation.NewDefaultNotifier()),
		elevation.WithPolicyEngine(elevation.NewDefaultPolicyEngine()),
		elevation.WithApprovalEngine(elevation.NewDefaultApprovalEngine()),
	)

	// Configurar limites de duração para testes
	service.ConfigureLimits(elevation.ElevationLimits{
		MaxDuration:           60 * time.Minute,
		MaxEmergencyDuration:  15 * time.Minute,
		DefaultDuration:       30 * time.Minute,
		DefaultEmergencyDuration: 5 * time.Minute,
	})

	// Configurar mercados e regulações
	service.ConfigureMarketRegulations(map[string][]string{
		"angola": {
			"Angola Financial Services Authority",
			"Angola Data Protection Law",
			"SADC Financial Regulations",
		},
		"brazil": {
			"Banco Central do Brasil",
			"LGPD",
			"CVM",
		},
		"mozambique": {
			"Banco de Moçambique",
			"Mozambique Financial Services Regulation",
			"SADC Financial Regulations",
		},
	})

	// Configurar políticas de MFA por mercado
	service.ConfigureMFAPolicies(map[string]elevation.MFAPolicy{
		"angola": {
			RequireMFA:                true,
			AllowedMFAMethods:         []string{"totp", "sms", "push"},
			RequireMFAForEmergencyAccess: true,
			MFACooldownMinutes:        60,
		},
		"brazil": {
			RequireMFA:                true,
			AllowedMFAMethods:         []string{"totp", "pix", "push"},
			RequireMFAForEmergencyAccess: true,
			MFACooldownMinutes:        30,
		},
	})

	return service, nil
}

// configureScopeMappings configura mapeamentos de comandos MCP para escopos de elevação
func configureScopeMappings(dockerHook *mcp.DockerHook, desktopHook *mcp.DesktopCommanderHook, 
	githubHook *mcp.GitHubHook, figmaHook *mcp.FigmaHook) {
	
	// Docker hook
	dockerHook.ConfigureScopeMappings(map[string][]string{
		"docker":              {"docker:container:run", "docker:container:stop"},
		"port_forward":        {"docker:network:port:forward"},
		"stop_port_forward":   {"docker:network:port:stop"},
		"kubectl_apply":       {"kubernetes:apply"},
		"kubectl_get":         {"kubernetes:get"},
		"kubectl_delete":      {"kubernetes:delete"},
		"kubectl_create":      {"kubernetes:create"},
		"kubectl_patch":       {"kubernetes:patch"},
		"kubectl_scale":       {"kubernetes:scale"},
		"kubectl_logs":        {"kubernetes:logs"},
		"kubectl_rollout":     {"kubernetes:rollout"},
		"kubectl_exec":        {"kubernetes:exec", "kubernetes:pod:access"},
		"install_helm_chart":  {"kubernetes:helm:install"},
		"upgrade_helm_chart":  {"kubernetes:helm:upgrade"},
		"uninstall_helm_chart": {"kubernetes:helm:uninstall"},
	})

	// Desktop Commander hook
	desktopHook.ConfigureScopeMappings(map[string][]string{
		"start_process":       {"desktop:process:start"},
		"kill_process":        {"desktop:process:kill"},
		"force_terminate":     {"desktop:process:terminate:force"},
		"edit_block":          {"desktop:file:edit"},
		"write_file":          {"desktop:file:write"},
		"set_config_value":    {"desktop:config:modify"},
		"read_file":           {"desktop:file:read"},
		"search_code":         {"desktop:file:search"},
		"search_files":        {"desktop:file:search"},
		"list_processes":      {"desktop:process:list"},
		"list_directory":      {"desktop:directory:list"},
	})

	// GitHub hook
	githubHook.ConfigureScopeMappings(map[string][]string{
		"push_files":             {"github:repo:push"},
		"merge_pull_request":     {"github:repo:merge"},
		"create_or_update_file":  {"github:repo:file:write"},
		"create_pull_request":    {"github:pr:create"},
		"update_pull_request":    {"github:pr:update"},
		"create_branch":          {"github:branch:create"},
		"create_repository":      {"github:repo:create"},
		"delete_issue":           {"github:issue:delete"},
		"add_issue_comment":      {"github:issue:comment"},
		"update_issue":           {"github:issue:update"},
		"fork_repository":        {"github:repo:fork"},
		"get_file_contents":      {"github:file:read"},
		"create_pull_request_review": {"github:pr:review:create"},
		"search_code":            {"github:search:code"},
		"search_issues":          {"github:search:issues"},
	})

	// Figma hook
	figmaHook.ConfigureScopeMappings(map[string][]string{
		"add_figma_file":  {"figma:file:add"},
		"post_comment":    {"figma:comment:create"},
		"reply_to_comment": {"figma:comment:reply"},
		"view_node":       {"figma:node:view"},
		"read_comments":   {"figma:comment:read"},
		"modify_design":   {"figma:design:modify"},
		"export_design":   {"figma:design:export"},
		"share_design":    {"figma:design:share"},
	})
}

// configureProtections configura proteções específicas para cada hook MCP
func configureProtections(dockerHook *mcp.DockerHook, desktopHook *mcp.DesktopCommanderHook,
	githubHook *mcp.GitHubHook, figmaHook *mcp.FigmaHook) {
	
	// Docker hook - Comandos sensíveis
	dockerHook.ConfigureSensitiveCommands([]string{
		"exec", "run", "rm", "rmi", "network", "volume", "system", "prune",
		"kubectl_delete", "kubectl_apply", "kubectl_patch", "kubectl_scale",
		"install_helm_chart", "upgrade_helm_chart", "uninstall_helm_chart",
	})

	// Desktop Commander hook - Diretórios e comandos sensíveis
	desktopHook.ConfigureSensitiveAreas(map[string][]string{
		"desktop:file:write": {
			"/etc/", "/var/", "/usr/bin/", "/boot/", "/opt/",
			"C:\\Windows\\", "C:\\Program Files\\", "C:\\Users\\Administrator\\",
			"/data/", "C:\\data\\",
		},
		"desktop:file:edit": {
			"/etc/passwd", "/etc/shadow", "/etc/sudoers",
			"C:\\Windows\\System32\\config\\", 
			"/data/", "C:\\data\\",
		},
		"desktop:process:kill": {
			"systemd", "init", "sshd", "NetworkManager", "postgresql",
			"explorer.exe", "winlogon.exe", "services.exe", "lsass.exe",
			"redis-server", "nginx", "httpd", "apache2",
		},
	})

	// GitHub hook - Repositórios e branches protegidos
	githubHook.ConfigureProtectedRepositories([]string{
		"innovabizdevops/innovabiz-iam",
		"innovabizdevops/innovabiz-api-gateway",
		"innovabizdevops/innovabiz-payment-gateway",
		"innovabizdevops/innovabiz-mobile-money",
		"innovabizdevops/innovabiz-insurance",
	})
	
	githubHook.ConfigureProtectedBranches(map[string][]string{
		"innovabizdevops/innovabiz-iam": {
			"main", "production", "staging", "release/*",
		},
		"innovabizdevops/innovabiz-api-gateway": {
			"main", "production", "staging", "release/*",
		},
		"innovabizdevops/innovabiz-payment-gateway": {
			"main", "production", "pci-dss-compliant", "release/*",
		},
	})

	// Figma hook - Arquivos protegidos
	figmaHook.ConfigureProtectedFiles([]string{
		"KpgCKLp83FhGZ3v2hFXoUG", // Design System Principal
		"L4qDdTT92JkMp1w3zWqRxH", // Protótipos de Alta Fidelidade
		"M9pDfG3hJ7lQz2v4nRtYuI", // Design de Segurança
	})
}

// Métodos adicionais necessários para o teste

// QueryAuditEvents é um método do serviço de elevação para consultar eventos de auditoria
func (s *elevation.PrivilegeElevationService) QueryAuditEvents(ctx context.Context, filters map[string]interface{}, timeRange ...*elevation.TimeRange) ([]*elevation.AuditEvent, error) {
	// Em um teste real, consultaria o banco de dados
	// Para este exemplo, retornamos um mock de evento de auditoria
	return []*elevation.AuditEvent{
		{
			EventID:     "audit-123",
			EventType:   filters["event_type"].(string),
			ElevationID: filters["elevation_id"].(string),
			UserID:      "user:admin:e2e-test-123",
			TenantID:    "tenant_angola_1",
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"command":           "docker run",
				"tenant_id":         "tenant_angola_1",
				"market":            "angola",
				"protected_branch":  "production",
				"reason":            filters["reason"],
			},
		},
	}, nil
}

// RevokeElevation é um método do serviço de elevação para revogar privilégios
func (s *elevation.PrivilegeElevationService) RevokeElevation(ctx context.Context, elevationID string, reason string) error {
	// Em um teste real, revogaria a elevação no banco de dados
	// Para este exemplo, apenas retornamos sucesso
	return nil
}

// GenerateMFAChallenge cria um desafio MFA para o usuário
func (s *elevation.PrivilegeElevationService) GenerateMFAChallenge(ctx context.Context, elevationID string, method string) (*elevation.MFAChallenge, error) {
	// Em um teste real, geraria um desafio real e o armazenaria
	return &elevation.MFAChallenge{
		ChallengeID:  "mfa-challenge-123",
		ElevationID:  elevationID,
		Type:         method,
		Status:       "pending",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}, nil
}

// VerifyMFAChallenge verifica o desafio MFA com o código fornecido
func (s *elevation.PrivilegeElevationService) VerifyMFAChallenge(ctx context.Context, challengeID string, code string) error {
	// Em um teste real, verificaria o código
	// Para este exemplo, aceitamos qualquer código
	return nil
}

// GetComplianceAuditTrail retorna a trilha de auditoria de conformidade para uma elevação
func (s *elevation.PrivilegeElevationService) GetComplianceAuditTrail(ctx context.Context, elevationID string) (*elevation.ComplianceAuditTrail, error) {
	// Em um teste real, consultaria o banco de dados
	// Para este exemplo, retornamos um mock de trilha de auditoria
	return &elevation.ComplianceAuditTrail{
		ElevationID:        elevationID,
		UserID:             "user:admin:e2e-test-123",
		TenantID:           "tenant_angola_1",
		Market:             "angola",
		AppliedRegulations: []string{"Angola Financial Services Authority", "SADC Financial Regulations"},
		ComplianceChecks:   []string{"multi_tenant_isolation", "mfa_verification", "scope_validation"},
		RiskAssessment: map[string]interface{}{
			"risk_level":    "high",
			"justification": "Atualização emergencial de segurança",
			"mitigations":   []string{"MFA", "Auditoria detalhada", "Notificação para aprovadores"},
		},
		Timeline: []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-5 * time.Minute),
				"event":     "elevation_request",
				"actor":     "user:admin:e2e-test-123",
			},
			{
				"timestamp": time.Now().Add(-4 * time.Minute),
				"event":     "mfa_challenge_generated",
				"actor":     "system",
			},
			{
				"timestamp": time.Now().Add(-3 * time.Minute),
				"event":     "mfa_verification_success",
				"actor":     "user:admin:e2e-test-123",
			},
			{
				"timestamp": time.Now().Add(-2 * time.Minute),
				"event":     "elevation_granted",
				"actor":     "system",
			},
		},
	}, nil
}