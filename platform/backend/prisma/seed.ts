/**
 * Prisma Seed Script — iBOS Platform
 * Popula a base de dados com dados demo para desenvolvimento.
 * 
 * Standards: APQC Process Classification Framework, ISIC Rev.4
 * Compliance: GDPR Art. 25 (Data Protection by Design), LGPD
 * 
 * Uso: npx prisma db seed
 *      npm run seed
 */
import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

async function main() {
    console.log('🌱 Seeding iBOS database...\n');

    // --- Wave 1: Core Business ---

    const grcFrameworks = await Promise.all([
        prisma.grcFramework.upsert({
            where: { tenantId_name_version: { tenantId: 'default', name: 'COBIT 2019', version: '2019' } },
            update: {},
            create: { name: 'COBIT 2019', type: 'COBIT', version: '2019', status: 'ACTIVE', scope: ['IT Governance', 'Enterprise Governance'], controls: { total: 40, domains: ['EDM', 'APO', 'BAI', 'DSS', 'MEA'] }, tenantId: 'default' },
        }),
        prisma.grcFramework.upsert({
            where: { tenantId_name_version: { tenantId: 'default', name: 'NIST CSF 2.0', version: '2.0' } },
            update: {},
            create: { name: 'NIST CSF 2.0', type: 'NIST', version: '2.0', status: 'ACTIVE', scope: ['Cybersecurity'], controls: { total: 108, functions: ['Govern', 'Identify', 'Protect', 'Detect', 'Respond', 'Recover'] }, tenantId: 'default' },
        }),
        prisma.grcFramework.upsert({
            where: { tenantId_name_version: { tenantId: 'default', name: 'ISO 27001:2022', version: '2022' } },
            update: {},
            create: { name: 'ISO 27001:2022', type: 'ISO27001', version: '2022', status: 'ACTIVE', scope: ['Information Security'], tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ GRC Frameworks: ${grcFrameworks.length}`);

    const complianceReqs = await Promise.all([
        prisma.complianceRequirement.create({
            data: { regulation: 'GDPR', article: 'Art. 5', description: 'Principles relating to processing of personal data', status: 'COMPLIANT', riskLevel: 'high', tenantId: 'default' },
        }),
        prisma.complianceRequirement.create({
            data: { regulation: 'LGPD', article: 'Art. 6', description: 'Princípios de tratamento de dados pessoais', status: 'COMPLIANT', riskLevel: 'high', tenantId: 'default' },
        }),
        prisma.complianceRequirement.create({
            data: { regulation: 'SOX', article: 'Sec. 302', description: 'Corporate Responsibility for Financial Reports', status: 'PENDING', riskLevel: 'critical', owner: 'CFO', tenantId: 'default' },
        }),
        prisma.complianceRequirement.create({
            data: { regulation: 'EU-AI-Act', article: 'Art. 9', description: 'Risk Management System for High-Risk AI', status: 'PENDING', riskLevel: 'high', owner: 'Chief AI Officer', tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ Compliance Requirements: ${complianceReqs.length}`);

    const risks = await Promise.all([
        prisma.riskRegister.create({
            data: { title: 'Cybersecurity Breach', category: 'CYBER', likelihood: 3, impact: 5, riskScore: 15, status: 'ASSESSED', owner: 'CISO', tenantId: 'default' },
        }),
        prisma.riskRegister.create({
            data: { title: 'Regulatory Non-Compliance (GDPR)', category: 'COMPLIANCE', likelihood: 2, impact: 5, riskScore: 10, status: 'MITIGATED', owner: 'DPO', tenantId: 'default' },
        }),
        prisma.riskRegister.create({
            data: { title: 'AI Model Bias', category: 'OPERATIONAL', likelihood: 3, impact: 4, riskScore: 12, status: 'IDENTIFIED', owner: 'Chief AI Officer', tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ Risk Register: ${risks.length}`);

    const contracts = await Promise.all([
        prisma.contract.create({
            data: { title: 'Cloud Infrastructure SLA', type: 'SLA', status: 'ACTIVE', counterparty: 'AWS', value: 500000, currency: 'EUR', startDate: new Date('2026-01-01'), endDate: new Date('2027-12-31'), renewalType: 'AUTO', tenantId: 'default' },
        }),
        prisma.contract.create({
            data: { title: 'Data Processing Agreement', type: 'MSA', status: 'ACTIVE', counterparty: 'DataCorp EU', value: 120000, currency: 'EUR', tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ Contracts: ${contracts.length}`);

    const employees = await Promise.all([
        prisma.hcmEmployee.create({
            data: { employeeId: 'EMP-001', firstName: 'Eduardo', lastName: 'Jeremias', email: 'eduardo@ibos.ai', department: 'Executive', position: 'CEO', level: 'C-Level', status: 'ACTIVE', skills: ['Strategy', 'AI', 'Enterprise Architecture'], tenantId: 'default' },
        }),
        prisma.hcmEmployee.create({
            data: { employeeId: 'EMP-002', firstName: 'Ana', lastName: 'Santos', email: 'ana.santos@ibos.ai', department: 'Engineering', position: 'CTO', level: 'C-Level', status: 'ACTIVE', skills: ['NestJS', 'React', 'Kubernetes', 'AI/ML'], tenantId: 'default' },
        }),
        prisma.hcmEmployee.create({
            data: { employeeId: 'EMP-003', firstName: 'Carlos', lastName: 'Mendes', email: 'carlos.mendes@ibos.ai', department: 'Finance', position: 'CFO', level: 'C-Level', status: 'ACTIVE', skills: ['IFRS', 'SOX', 'Treasury'], tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ HCM Employees: ${employees.length}`);

    const campaigns = await Promise.all([
        prisma.marketingCampaign.create({
            data: { name: 'iBOS Platform Launch', type: 'CONTENT', status: 'ACTIVE', channel: ['LinkedIn', 'Email', 'Webinar'], budget: 50000, currency: 'EUR', startDate: new Date('2026-03-01'), endDate: new Date('2026-06-30'), tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ Marketing Campaigns: ${campaigns.length}`);

    const partners = await Promise.all([
        prisma.partner.create({
            data: { name: 'Accenture Portugal', type: 'CONSULTING', status: 'ACTIVE', tier: 'PLATINUM', region: 'EU-PT', industry: 'Technology', tenantId: 'default' },
        }),
        prisma.partner.create({
            data: { name: 'AWS Partner Network', type: 'TECHNOLOGY', status: 'ACTIVE', tier: 'GOLD', region: 'GLOBAL', industry: 'Cloud', tenantId: 'default' },
        }),
    ]);
    console.log(`  ✅ Partners: ${partners.length}`);

    // --- Wave 2: Operations ---

    const workflows = await prisma.processWorkflow.create({
        data: { name: 'Employee Onboarding', category: 'ONBOARDING', status: 'ACTIVE', version: 1, steps: [{ stepId: 1, name: 'Document Collection', type: 'FORM' }, { stepId: 2, name: 'IT Setup', type: 'AUTOMATION' }, { stepId: 3, name: 'Training', type: 'TASK' }], tenantId: 'default' },
    });
    console.log(`  ✅ Process Workflows: 1`);

    const articles = await prisma.knowledgeArticle.create({
        data: { title: 'iBOS Platform Architecture Guide', category: 'REFERENCE', status: 'PUBLISHED', content: 'The iBOS platform uses a Hexagonal Architecture with Multi-Tenant Native design...', author: 'Architecture Team', tags: ['architecture', 'hexagonal', 'multi-tenant'], language: 'en', tenantId: 'default' },
    });
    console.log(`  ✅ Knowledge Articles: 1`);

    const integrations = await prisma.integrationConfig.create({
        data: { name: 'SAP ERP Connector', type: 'REST', provider: 'SAP', status: 'ACTIVE', direction: 'BIDIRECTIONAL', endpoint: 'https://sap.example.com/api/v1', authType: 'OAUTH2', tenantId: 'default' },
    });
    console.log(`  ✅ Integration Configs: 1`);

    // --- Wave 3: Open Ecosystem ---

    const bankingApis = await prisma.openBankingApi.create({
        data: { name: 'PSD2 Account Information', version: '3.1', standard: 'PSD2', status: 'ACTIVE', endpoint: '/api/v3/accounts', scope: ['accounts', 'balances', 'transactions'], provider: 'Euro Banking API', tenantId: 'default' },
    });
    console.log(`  ✅ Open Banking APIs: 1`);

    console.log('\n🎉 Seed completed successfully!');
}

main()
    .catch((e) => {
        console.error('❌ Seed error:', e);
        process.exit(1);
    })
    .finally(async () => {
        await prisma.$disconnect();
    });
