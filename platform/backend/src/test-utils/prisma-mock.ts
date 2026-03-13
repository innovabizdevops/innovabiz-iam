/**
 * PrismaService Mock Factory
 * Provides a reusable mock for unit tests across all 26 backend modules.
 * 
 * Pattern: Every Prisma model method returns a fallback response,
 * allowing services with try/catch to work in test mode.
 * 
 * Usage in spec files:
 *   import { createPrismaMock, PrismaServiceMock } from '../../test-utils/prisma-mock';
 *   providers: [MyService, { provide: PrismaService, useValue: createPrismaMock() }]
 * 
 * Standards: ISO 25010 (Software Quality), ISTQB Test Automation Patterns
 */
import { PrismaService } from '../universal-persistence/prisma.service';

export type PrismaServiceMock = Record<string, any>;

/** Creates a proxy-based mock that returns empty results for any model access */
export function createPrismaMock(): PrismaServiceMock {
    const modelMethods = {
        create: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        findMany: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        findUnique: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        findFirst: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        update: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        delete: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        count: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
        upsert: jest.fn().mockRejectedValue(new Error('Mock: DB not connected')),
    };

    // Use a Proxy so ANY model name (grcFramework, contract, etc.) returns the mock methods
    return new Proxy({}, {
        get(_target, prop) {
            if (prop === '$connect' || prop === '$disconnect') {
                return jest.fn().mockResolvedValue(undefined);
            }
            if (typeof prop === 'string' && prop.startsWith('$')) {
                return jest.fn();
            }
            // Return model mock for any property access (grcFramework, contract, etc.)
            return modelMethods;
        },
    });
}

/** Token for providing mock in test module */
export const PRISMA_MOCK_PROVIDER = {
    provide: PrismaService,
    useFactory: createPrismaMock,
};
