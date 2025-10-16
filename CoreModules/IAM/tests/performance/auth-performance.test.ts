/**
 * @file auth-performance.test.ts
 * @description Testes de performance para autenticação do IAM INNOVABIZ
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import * as autocannon from 'autocannon';
import { spawn } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { promisify } from 'util';

// Caminho para relatórios
const REPORTS_DIR = path.join(__dirname, '..', '..', 'reports', 'performance');

// Garantir que o diretório de relatórios existe
if (!fs.existsSync(REPORTS_DIR)) {
  fs.mkdirSync(REPORTS_DIR, { recursive: true });
}

describe('Testes de Performance de Autenticação', () => {
  let server;
  const PORT = 4000;
  const BASE_URL = `http://localhost:${PORT}`;

  // Iniciar servidor para testes
  beforeAll(async () => {
    // Inicializa o servidor para testes em modo isolado
    server = spawn('node', ['dist/main.js'], {
      env: { ...process.env, NODE_ENV: 'test', PORT: PORT.toString() },
      stdio: 'pipe',
    });

    // Aguardar inicialização do servidor
    await new Promise((resolve) => {
      server.stdout.on('data', (data) => {
        if (data.toString().includes('Nest application successfully started')) {
          console.log('Servidor inicializado para testes de performance');
          resolve(null);
        }
      });
    });

    // Espera adicional para garantir que o servidor está completamente pronto
    await new Promise((resolve) => setTimeout(resolve, 3000));
  });

  afterAll(() => {
    // Encerra o servidor após os testes
    if (server) {
      server.kill('SIGTERM');
    }
  });

  /**
   * Teste de carga para o fluxo de autenticação básico
   * Verifica a capacidade do sistema em lidar com múltiplas requisições simultâneas
   */
  test('Teste de carga para fluxo de autenticação básico', async () => {
    const testName = 'basic-auth-load';
    const loginPayload = JSON.stringify({
      username: 'performance_test@example.com',
      password: 'Test123!',
      tenantId: 'performance-tenant',
    });

    // Configuração para teste de carga
    const autocannomPromise = promisify(autocannon);
    const result = await autocannomPromise({
      url: BASE_URL + '/api/v1/auth/login',
      connections: 100, // Conexões simultâneas
      duration: 20, // Duração em segundos
      headers: {
        'Content-Type': 'application/json',
      },
      method: 'POST',
      body: loginPayload,
      requests: [
        {
          path: '/api/v1/auth/login',
          method: 'POST',
          body: loginPayload,
          headers: { 'Content-Type': 'application/json' },
        },
      ],
    });

    // Salvar resultados
    fs.writeFileSync(
      path.join(REPORTS_DIR, `${testName}-result.json`),
      JSON.stringify(result, null, 2)
    );

    // Gerar relatório
    generateReport(result, testName);

    // Verificações básicas de performance
    expect(result.errors).toBeLessThan(5);
    expect(result.timeouts).toBe(0);
    expect(result.non2xx).toBeLessThan(5);
    
    // Requisições por segundo adequadas para ambiente de produção
    expect(result.requests.average).toBeGreaterThan(200);
    
    // Latência máxima adequada
    expect(result.latency.p99).toBeLessThan(500); // ms
  });

  /**
   * Teste de carga para o fluxo completo de WebAuthn/FIDO2
   * Verifica o desempenho do sistema com autenticação avançada
   */
  test('Teste de carga para WebAuthn/FIDO2', async () => {
    const testName = 'webauthn-load';
    
    // Simulação simplificada de fluxo WebAuthn/FIDO2
    // Em cenários reais, seria necessário um cliente WebAuthn completo
    const generateOptionsPayload = JSON.stringify({
      username: 'webauthn_test@example.com',
      tenantId: 'performance-tenant',
    });

    // Configuração para teste de carga
    const autocannomPromise = promisify(autocannon);
    const result = await autocannomPromise({
      url: BASE_URL + '/api/v1/auth/webauthn/generate-options',
      connections: 50, // Menos conexões devido à complexidade
      duration: 20, // Duração em segundos
      headers: {
        'Content-Type': 'application/json',
      },
      method: 'POST',
      body: generateOptionsPayload,
    });

    // Salvar resultados
    fs.writeFileSync(
      path.join(REPORTS_DIR, `${testName}-result.json`),
      JSON.stringify(result, null, 2)
    );

    // Gerar relatório
    generateReport(result, testName);

    // Verificações específicas para WebAuthn
    // Expectativas mais flexíveis devido à natureza complexa da operação
    expect(result.errors).toBeLessThan(10);
    expect(result.timeouts).toBeLessThan(5);
    
    // Requisições por segundo adequadas para WebAuthn
    expect(result.requests.average).toBeGreaterThan(50);
    
    // Latência máxima maior permitida para operações complexas
    expect(result.latency.p99).toBeLessThan(2000); // ms
  });

  /**
   * Teste de carga para operações multi-tenant
   * Verifica se o sistema mantém desempenho adequado com múltiplos tenants
   */
  test('Teste de carga para operações multi-tenant', async () => {
    const testName = 'multi-tenant-load';
    
    // Array de tenants para teste
    const tenants = [
      'tenant-1', 'tenant-2', 'tenant-3', 'tenant-4', 'tenant-5',
      'tenant-6', 'tenant-7', 'tenant-8', 'tenant-9', 'tenant-10'
    ];
    
    // Função para gerar payload com tenant aleatório
    const getRandomTenantPayload = () => {
      const randomTenant = tenants[Math.floor(Math.random() * tenants.length)];
      return JSON.stringify({
        username: `user@${randomTenant}.com`,
        password: 'Test123!',
        tenantId: randomTenant,
      });
    };
    
    // Criar múltiplas requisições com diferentes tenants
    const requests = tenants.map(tenant => ({
      path: '/api/v1/auth/login',
      method: 'POST',
      body: JSON.stringify({
        username: `user@${tenant}.com`,
        password: 'Test123!',
        tenantId: tenant,
      }),
      headers: { 'Content-Type': 'application/json' },
    }));

    // Configuração para teste de carga
    const autocannomPromise = promisify(autocannon);
    const result = await autocannomPromise({
      url: BASE_URL,
      connections: 200, // Alta concorrência para testar isolamento de tenants
      duration: 30, // Duração maior para avaliar consistência
      headers: {
        'Content-Type': 'application/json',
      },
      requests,
    });

    // Salvar resultados
    fs.writeFileSync(
      path.join(REPORTS_DIR, `${testName}-result.json`),
      JSON.stringify(result, null, 2)
    );

    // Gerar relatório
    generateReport(result, testName);

    // Verificações específicas para cenário multi-tenant
    expect(result.errors).toBeLessThan(20);
    expect(result.timeouts).toBeLessThan(10);
    
    // Manter throughput aceitável mesmo com múltiplos tenants
    expect(result.requests.average).toBeGreaterThan(150);
    
    // Latência deve se manter estável mesmo com múltiplos tenants
    expect(result.latency.p99).toBeLessThan(800); // ms
  });

  /**
   * Teste de carga para simulação de picos de acesso
   * Verifica a resposta do sistema a picos súbitos de tráfego
   */
  test('Teste de resistência a picos de acesso', async () => {
    const testName = 'peak-load';
    
    // Configuração para pico de acesso
    const autocannomPromise = promisify(autocannon);
    const result = await autocannomPromise({
      url: BASE_URL + '/api/v1/auth/login',
      connections: 500, // Pico alto de conexões
      duration: 15, // Curta duração para simular pico
      amount: 10000, // Total de requisições
      headers: {
        'Content-Type': 'application/json',
      },
      method: 'POST',
      body: JSON.stringify({
        username: 'stress_test@example.com',
        password: 'Test123!',
        tenantId: 'stress-tenant',
      }),
    });

    // Salvar resultados
    fs.writeFileSync(
      path.join(REPORTS_DIR, `${testName}-result.json`),
      JSON.stringify(result, null, 2)
    );

    // Gerar relatório
    generateReport(result, testName);

    // Em cenários de pico, toleramos mais erros e latência maior
    expect(result.errors).toBeLessThan(50);
    expect(result.timeouts).toBeLessThan(30);
    
    // Sistema deve degradar graciosamente, não colapsar completamente
    expect(result.requests.average).toBeGreaterThan(50);
    
    // Em picos, aceitamos latência maior, mas com limite razoável
    expect(result.latency.p99).toBeLessThan(5000); // ms
  });
});

/**
 * Função auxiliar para gerar relatório em HTML
 * @param result Resultado do teste de autocannon
 * @param testName Nome do teste para identificar o relatório
 */
function generateReport(result, testName) {
  // Template básico para relatório em HTML
  const htmlReport = `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Relatório de Performance - ${testName}</title>
      <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        .metric { margin-bottom: 20px; }
        .metric h2 { color: #555; }
        table { border-collapse: collapse; width: 100%; }
        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        .warning { color: orange; }
        .error { color: red; }
        .success { color: green; }
      </style>
    </head>
    <body>
      <h1>Relatório de Performance - ${testName}</h1>
      <div class="metric">
        <h2>Informações Gerais</h2>
        <table>
          <tr><th>Métrica</th><th>Valor</th></tr>
          <tr><td>URL</td><td>${result.url}</td></tr>
          <tr><td>Duração</td><td>${result.duration}s</td></tr>
          <tr><td>Conexões</td><td>${result.connections}</td></tr>
          <tr><td>Pipelining</td><td>${result.pipelining}</td></tr>
          <tr><td>Total de Erros</td><td class="${result.errors > 0 ? 'error' : 'success'}">${result.errors}</td></tr>
          <tr><td>Total de Timeouts</td><td class="${result.timeouts > 0 ? 'error' : 'success'}">${result.timeouts}</td></tr>
          <tr><td>Respostas não-2xx</td><td class="${result.non2xx > 0 ? 'warning' : 'success'}">${result.non2xx}</td></tr>
        </table>
      </div>
      
      <div class="metric">
        <h2>Requisições</h2>
        <table>
          <tr><th>Métrica</th><th>Valor</th></tr>
          <tr><td>Total</td><td>${result.requests.total}</td></tr>
          <tr><td>Média (req/sec)</td><td>${result.requests.average.toFixed(2)}</td></tr>
          <tr><td>Mínimo (req/sec)</td><td>${result.requests.min}</td></tr>
          <tr><td>Máximo (req/sec)</td><td>${result.requests.max}</td></tr>
        </table>
      </div>
      
      <div class="metric">
        <h2>Latência (ms)</h2>
        <table>
          <tr><th>Métrica</th><th>Valor</th></tr>
          <tr><td>Média</td><td>${result.latency.average.toFixed(2)}</td></tr>
          <tr><td>Mínima</td><td>${result.latency.min}</td></tr>
          <tr><td>Máxima</td><td>${result.latency.max}</td></tr>
          <tr><td>P50</td><td>${result.latency.p50.toFixed(2)}</td></tr>
          <tr><td>P90</td><td>${result.latency.p90.toFixed(2)}</td></tr>
          <tr><td>P99</td><td>${result.latency.p99.toFixed(2)}</td></tr>
        </table>
      </div>
      
      <div class="metric">
        <h2>Throughput (MB/sec)</h2>
        <table>
          <tr><th>Métrica</th><th>Valor</th></tr>
          <tr><td>Média</td><td>${(result.throughput.average / 1024 / 1024).toFixed(4)}</td></tr>
          <tr><td>Mínimo</td><td>${(result.throughput.min / 1024 / 1024).toFixed(4)}</td></tr>
          <tr><td>Máximo</td><td>${(result.throughput.max / 1024 / 1024).toFixed(4)}</td></tr>
          <tr><td>Total (MB)</td><td>${(result.throughput.total / 1024 / 1024).toFixed(4)}</td></tr>
        </table>
      </div>
      
      <div class="metric">
        <h2>Avaliação de Performance</h2>
        <table>
          <tr><th>Critério</th><th>Status</th></tr>
          <tr>
            <td>Requisições/segundo</td>
            <td class="${result.requests.average > 200 ? 'success' : (result.requests.average > 100 ? 'warning' : 'error')}">
              ${result.requests.average > 200 ? 'Excelente' : (result.requests.average > 100 ? 'Aceitável' : 'Insuficiente')}
            </td>
          </tr>
          <tr>
            <td>Latência P99</td>
            <td class="${result.latency.p99 < 500 ? 'success' : (result.latency.p99 < 1000 ? 'warning' : 'error')}">
              ${result.latency.p99 < 500 ? 'Excelente' : (result.latency.p99 < 1000 ? 'Aceitável' : 'Alta')}
            </td>
          </tr>
          <tr>
            <td>Taxa de erros</td>
            <td class="${result.errors === 0 ? 'success' : (result.errors < 10 ? 'warning' : 'error')}">
              ${result.errors === 0 ? 'Sem erros' : (result.errors < 10 ? 'Poucos erros' : 'Muitos erros')}
            </td>
          </tr>
        </table>
      </div>
    </body>
    </html>
  `;

  fs.writeFileSync(path.join(REPORTS_DIR, `${testName}-report.html`), htmlReport);
}