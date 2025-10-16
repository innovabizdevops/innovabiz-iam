/**
 * @file testcontainers.setup.ts
 * @description Configuração de TestContainers para testes de integração e end-to-end
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { GenericContainer, StartedTestContainer, Network } from 'testcontainers';
import { PostgreSqlContainer } from 'testcontainers/modules/postgresql';
import { RedisContainer } from 'testcontainers/modules/redis';
import * as path from 'path';
import * as fs from 'fs';

/**
 * Classe para configuração e gerenciamento de containers para testes
 * Implementa setup de PostgreSQL, Redis e outros serviços necessários
 */
export class TestContainersSetup {
  private static postgresContainer: StartedTestContainer;
  private static redisContainer: StartedTestContainer;
  private static mockServicesContainer: StartedTestContainer;
  private static network: Network;

  /**
   * Configuração e inicialização de todos os containers necessários para testes
   * @returns Promise que resolve quando todos os containers estão prontos
   */
  static async setupContainers(): Promise<void> {
    console.log('Inicializando TestContainers para testes...');
    
    // Criar rede compartilhada para comunicação entre containers
    this.network = await new Network().start();
    
    // Iniciar containers em paralelo para otimizar tempo de setup
    const [postgres, redis, mockServices] = await Promise.all([
      this.setupPostgresContainer(),
      this.setupRedisContainer(),
      this.setupMockServicesContainer(),
    ]);
    
    this.postgresContainer = postgres;
    this.redisContainer = redis;
    this.mockServicesContainer = mockServices;
    
    // Configurar variáveis de ambiente para testes
    process.env.DATABASE_URL = this.getPostgresConnectionString();
    process.env.REDIS_URL = this.getRedisConnectionString();
    process.env.RISK_MANAGEMENT_API_URL = `http://${this.mockServicesContainer.getHost()}:8080/risk-management`;
    process.env.PAYMENT_GATEWAY_API_URL = `http://${this.mockServicesContainer.getHost()}:8080/payment-gateway`;
    
    console.log('TestContainers inicializados com sucesso');
    console.log(`PostgreSQL: ${process.env.DATABASE_URL}`);
    console.log(`Redis: ${process.env.REDIS_URL}`);
    console.log(`Mock Services: ${this.mockServicesContainer.getHost()}:8080`);
  }

  /**
   * Encerra todos os containers iniciados
   * @returns Promise que resolve quando todos os containers forem encerrados
   */
  static async teardownContainers(): Promise<void> {
    console.log('Encerrando TestContainers...');
    
    const containers = [
      this.postgresContainer, 
      this.redisContainer, 
      this.mockServicesContainer
    ].filter(Boolean);
    
    await Promise.all(containers.map(container => container.stop()));
    
    if (this.network) {
      await this.network.stop();
    }
    
    console.log('TestContainers encerrados com sucesso');
  }

  /**
   * Configura e inicia container do PostgreSQL com schema e dados de teste
   * @returns Promise com container iniciado
   */
  private static async setupPostgresContainer(): Promise<StartedTestContainer> {
    // Caminho para scripts de inicialização do banco
    const initScriptsPath = path.resolve(__dirname, '../../..', 'scripts/database');
    
    // Verificar se diretório existe
    if (!fs.existsSync(initScriptsPath)) {
      throw new Error(`Diretório de scripts não encontrado: ${initScriptsPath}`);
    }
    
    // Iniciar container do PostgreSQL
    const postgresContainer = await new PostgreSqlContainer('postgres:15-alpine')
      .withNetwork(this.network)
      .withNetworkAliases('postgres')
      .withDatabase('innovabiz_iam_test')
      .withUsername('test_user')
      .withPassword('test_password')
      .withCopyFilesToContainer([
        {
          source: path.join(initScriptsPath, 'iam_core.sql'),
          target: '/docker-entrypoint-initdb.d/01_iam_core.sql',
        },
        {
          source: path.join(initScriptsPath, '01_webauthn_schema.sql'),
          target: '/docker-entrypoint-initdb.d/02_webauthn_schema.sql',
        },
        {
          source: path.join(initScriptsPath, '02_webauthn_performance_optimization.sql'),
          target: '/docker-entrypoint-initdb.d/03_webauthn_performance_optimization.sql',
        },
        {
          source: path.join(initScriptsPath, '03_webauthn_initial_data.sql'),
          target: '/docker-entrypoint-initdb.d/04_webauthn_initial_data.sql',
        },
        {
          source: path.join(__dirname, 'test_data.sql'),
          target: '/docker-entrypoint-initdb.d/05_test_data.sql',
        },
      ])
      .withExposedPorts(5432)
      .start();
    
    console.log('Container PostgreSQL inicializado');
    return postgresContainer;
  }

  /**
   * Configura e inicia container do Redis
   * @returns Promise com container iniciado
   */
  private static async setupRedisContainer(): Promise<StartedTestContainer> {
    const redisContainer = await new RedisContainer('redis:7-alpine')
      .withNetwork(this.network)
      .withNetworkAliases('redis')
      .withExposedPorts(6379)
      .start();
    
    console.log('Container Redis inicializado');
    return redisContainer;
  }

  /**
   * Configura e inicia container com serviços mock para testes
   * @returns Promise com container iniciado
   */
  private static async setupMockServicesContainer(): Promise<StartedTestContainer> {
    // Caminho para dockerfile de serviços mock
    const mockServicesPath = path.resolve(__dirname, 'mock-services');
    
    // Iniciar container com serviços mock
    const mockServicesContainer = await GenericContainer.fromDockerfile(mockServicesPath)
      .build()
      .withNetwork(this.network)
      .withNetworkAliases('mock-services')
      .withExposedPorts(8080)
      .start();
    
    console.log('Container de serviços mock inicializado');
    return mockServicesContainer;
  }

  /**
   * Retorna string de conexão para o PostgreSQL
   * @returns String de conexão
   */
  static getPostgresConnectionString(): string {
    if (!this.postgresContainer) {
      throw new Error('Container PostgreSQL não está inicializado');
    }

    const host = this.postgresContainer.getHost();
    const port = this.postgresContainer.getMappedPort(5432);
    
    return `postgres://test_user:test_password@${host}:${port}/innovabiz_iam_test`;
  }

  /**
   * Retorna string de conexão para o Redis
   * @returns String de conexão
   */
  static getRedisConnectionString(): string {
    if (!this.redisContainer) {
      throw new Error('Container Redis não está inicializado');
    }
    
    const host = this.redisContainer.getHost();
    const port = this.redisContainer.getMappedPort(6379);
    
    return `redis://${host}:${port}`;
  }

  /**
   * Retorna URL base para os serviços mock
   * @returns URL base
   */
  static getMockServicesBaseUrl(): string {
    if (!this.mockServicesContainer) {
      throw new Error('Container de serviços mock não está inicializado');
    }
    
    const host = this.mockServicesContainer.getHost();
    const port = this.mockServicesContainer.getMappedPort(8080);
    
    return `http://${host}:${port}`;
  }
}