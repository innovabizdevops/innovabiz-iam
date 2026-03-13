import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

describe('GRC & Compliance E2E', () => {
    let app: INestApplication;

    beforeAll(async () => {
        const moduleFixture: TestingModule = await Test.createTestingModule({
            imports: [AppModule],
        }).compile();

        app = moduleFixture.createNestApplication();
        await app.init();
    });

    afterAll(async () => {
        await app.close();
    });

    // --- Governance Risk Compliance ---
    describe('/governance-risk-compliance', () => {
        it('GET / should return hello message', () => {
            return request(app.getHttpServer())
                .get('/governance-risk-compliance')
                .expect(200)
                .expect((res) => {
                    expect(res.text).toContain('Governance Risk');
                });
        });

        it('GET /dashboard should return dashboard', () => {
            return request(app.getHttpServer())
                .get('/governance-risk-compliance/dashboard')
                .expect(200)
                .expect((res) => {
                    expect(res.body.module).toBe('GovernanceRiskCompliance');
                    expect(res.body.status).toBe('OPERATIONAL');
                });
        });

        it('GET /controls should return controls array', () => {
            return request(app.getHttpServer())
                .get('/governance-risk-compliance/controls')
                .expect(200)
                .expect((res) => {
                    expect(res.body).toBeInstanceOf(Array);
                    expect(res.body.length).toBeGreaterThan(0);
                });
        });

        it('GET /risk-matrix should return risk matrix', () => {
            return request(app.getHttpServer())
                .get('/governance-risk-compliance/risk-matrix')
                .expect(200)
                .expect((res) => {
                    expect(res.body).toBeInstanceOf(Array);
                });
        });

        it('GET /capabilities should return capabilities', () => {
            return request(app.getHttpServer())
                .get('/governance-risk-compliance/capabilities')
                .expect(200)
                .expect((res) => {
                    expect(res.body).toBeInstanceOf(Array);
                    expect(res.body.length).toBeGreaterThan(5);
                });
        });

        it('POST / should create a record', () => {
            return request(app.getHttpServer())
                .post('/governance-risk-compliance')
                .send({ name: 'E2E Test GRC', framework: 'COBIT' })
                .expect(201)
                .expect((res) => {
                    expect(res.body.status).toBe('SUCCESS');
                    expect(res.body.id).toBeDefined();
                });
        });
    });

    // --- Global Compliance ---
    describe('/global-compliance', () => {
        it('GET / should return hello message', () => {
            return request(app.getHttpServer())
                .get('/global-compliance')
                .expect(200);
        });

        it('GET /capabilities should return capabilities', () => {
            return request(app.getHttpServer())
                .get('/global-compliance/capabilities')
                .expect(200)
                .expect((res) => {
                    expect(res.body).toBeInstanceOf(Array);
                });
        });
    });

    // --- Audit Management ---
    describe('/audit-management', () => {
        it('GET / should return hello message', () => {
            return request(app.getHttpServer())
                .get('/audit-management')
                .expect(200);
        });
    });

    // --- Compliance Risk Management ---
    describe('/compliance-risk-management', () => {
        it('GET / should return hello message', () => {
            return request(app.getHttpServer())
                .get('/compliance-risk-management')
                .expect(200);
        });
    });

    // --- Contract Management ---
    describe('/contract-management', () => {
        it('GET / should return hello message', () => {
            return request(app.getHttpServer())
                .get('/contract-management')
                .expect(200);
        });
    });
});
