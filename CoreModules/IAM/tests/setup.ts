/**
 * @file setup.ts
 * @description Configuração global para testes do IAM INNOVABIZ
 * @author INNOVABIZ Development Team
 * @copyright 2025 INNOVABIZ
 * @version 1.0.0
 */

import { config } from 'dotenv';
import { resolve } from 'path';

// Carrega variáveis de ambiente para testes
config({ path: resolve(__dirname, '.env.test') });

// Configura tempo limite global para testes
jest.setTimeout(30000);

// Extensão global de matchers para testes
expect.extend({
  toBeWithinRange(received, floor, ceiling) {
    const pass = received >= floor && received <= ceiling;
    if (pass) {
      return {
        message: () => `esperado que ${received} não estivesse entre ${floor} e ${ceiling}`,
        pass: true,
      };
    } else {
      return {
        message: () => `esperado que ${received} estivesse entre ${floor} e ${ceiling}`,
        pass: false,
      };
    }
  },
});

// Tipos para extensão global de matchers
declare global {
  namespace jest {
    interface Matchers<R> {
      toBeWithinRange(floor: number, ceiling: number): R;
    }
  }
}

// Mock para API Fetch para testes
global.fetch = jest.fn() as jest.Mock;

// Configuração de mocks para WebAuthn API
Object.defineProperty(global.navigator, 'credentials', {
  value: {
    create: jest.fn(),
    get: jest.fn(),
  },
});

// Suprime logs durante testes, exceto erros críticos
const originalConsoleLog = console.log;
const originalConsoleInfo = console.info;
const originalConsoleWarn = console.warn;

// Desativa logs em testes, exceto se DEBUG=true
if (!process.env.DEBUG) {
  console.log = jest.fn();
  console.info = jest.fn();
  console.warn = jest.fn();
}

// Restaura console após testes
afterAll(() => {
  console.log = originalConsoleLog;
  console.info = originalConsoleInfo;
  console.warn = originalConsoleWarn;
});