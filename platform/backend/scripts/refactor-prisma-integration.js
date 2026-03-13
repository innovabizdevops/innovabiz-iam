/**
 * Script para refactorar services backend com integração Prisma
 * Adiciona import do PrismaService e Constructor Injection
 * Substitui Map<string, any> por PrismaService com fallback
 * 
 * Uso: node scripts/refactor-prisma-integration.js
 */
const fs = require('fs');
const path = require('path');

// Mapeamento: directório do módulo → modelo Prisma
const MODULE_MAP = {
  'compliance-risk-management': { model: 'complianceRequirement', prefix: 'RISK', fields: { name: 'regulation', defaults: { regulation: 'CUSTOM', description: 'Auto-created', riskLevel: 'medium' } } },
  'audit-management': { model: 'auditFinding', prefix: 'AUD', fields: { name: 'title', defaults: { auditId: 'AUD-DEFAULT', title: 'Untitled Finding', type: 'OBSERVATION', description: 'Auto-created' } } },
  'contract-management': { model: 'contract', prefix: 'CTR', fields: { name: 'title', defaults: { title: 'Untitled Contract', type: 'SLA', counterparty: 'Unknown' } } },
  'human-capital-management': { model: 'hcmEmployee', prefix: 'HCM', fields: { name: 'firstName', defaults: { employeeId: 'EMP-' + Date.now(), firstName: 'Unknown', lastName: 'Employee', email: 'unknown@ibos.ai' } } },
  'marketing-management': { model: 'marketingCampaign', prefix: 'MKT', fields: { name: 'name', defaults: { name: 'Untitled Campaign', type: 'EMAIL' } } },
  'partner-management': { model: 'partner', prefix: 'PRT', fields: { name: 'name', defaults: { name: 'Unknown Partner', type: 'REFERRAL' } } },
  'process-management': { model: 'processWorkflow', prefix: 'PRC', fields: { name: 'name', defaults: { name: 'Untitled Workflow', category: 'CUSTOM', steps: '[]' } } },
  'quality-management': { model: 'qualityCheck', prefix: 'QAL', fields: { name: 'name', defaults: { name: 'Untitled Check', type: 'INSPECTION' } } },
  'knowledge-management': { model: 'knowledgeArticle', prefix: 'KNW', fields: { name: 'title', defaults: { title: 'Untitled Article', category: 'REFERENCE', content: 'Empty' } } },
  'vendor-management': { model: 'vendorRecord', prefix: 'VND', fields: { name: 'name', defaults: { name: 'Unknown Vendor', category: 'SERVICES' } } },
  'data-management': { model: 'dataAsset', prefix: 'DAT', fields: { name: 'name', defaults: { name: 'Untitled Asset', type: 'DATASET', classification: 'INTERNAL' } } },
  'device-management': { model: 'device', prefix: 'DEV', fields: { name: 'name', defaults: { name: 'Unknown Device', type: 'IOT' } } },
  'integration-services': { model: 'integrationConfig', prefix: 'INT', fields: { name: 'name', defaults: { name: 'Untitled Integration', type: 'REST' } } },
  'support-services': { model: 'supportTicket', prefix: 'SUP', fields: { name: 'subject', defaults: { ticketNumber: 'TKT-' + Date.now(), subject: 'Untitled Ticket', channel: 'PORTAL', description: 'Auto-created' } } },
  'document-management': { model: 'documentRecord', prefix: 'DOC', fields: { name: 'title', defaults: { title: 'Untitled Document', type: 'REPORT' } } },
  'notification-services': { model: 'notificationTemplate', prefix: 'NTF', fields: { name: 'name', defaults: { name: 'Untitled Template', channel: 'EMAIL', body: 'Empty' } } },
  'open-banking': { model: 'openBankingApi', prefix: 'OBK', fields: { name: 'name', defaults: { name: 'Untitled API', version: '1.0', standard: 'PSD2', endpoint: '/api/v1' } } },
  'open-finance': { model: 'openFinanceProduct', prefix: 'OFN', fields: { name: 'name', defaults: { name: 'Untitled Product', category: 'INVESTMENT', provider: 'Unknown' } } },
  'open-insurance': { model: 'openInsurancePolicy', prefix: 'OIN', fields: { name: 'policyNumber', defaults: { policyNumber: 'POL-' + Date.now(), type: 'AUTO', insurer: 'Unknown' } } },
  'open-data': { model: 'openDataDataset', prefix: 'ODT', fields: { name: 'name', defaults: { name: 'Untitled Dataset', category: 'GOVERNMENT', format: 'JSON' } } },
  'open-health': { model: 'openHealthResource', prefix: 'OHL', fields: { name: 'resourceType', defaults: { resourceType: 'Patient', data: '{}' } } },
  'open-education': { model: 'openEducationCourse', prefix: 'OED', fields: { name: 'title', defaults: { title: 'Untitled Course', provider: 'Unknown', format: 'MOOC' } } },
  'open-innovation': { model: 'openInnovationChallenge', prefix: 'OIV', fields: { name: 'title', defaults: { title: 'Untitled Challenge', type: 'HACKATHON', description: 'Auto-created' } } },
  'innovation-management': { model: 'innovationProject', prefix: 'INV', fields: { name: 'name', defaults: { name: 'Untitled Project', type: 'POC' } } },
};

const MODULES_DIR = path.join(__dirname, '..', 'src', 'modules');

let updated = 0;
let skipped = 0;
let errors = 0;

for (const [dir, config] of Object.entries(MODULE_MAP)) {
  const modulePath = path.join(MODULES_DIR, dir);
  
  if (!fs.existsSync(modulePath)) {
    console.log(`⚠️  SKIP: ${dir} (directory not found)`);
    skipped++;
    continue;
  }

  // Find the service file
  const files = fs.readdirSync(modulePath).filter(f => f.endsWith('Service.ts') && !f.endsWith('.spec.ts'));
  
  if (files.length === 0) {
    console.log(`⚠️  SKIP: ${dir} (no service file found)`);
    skipped++;
    continue;
  }

  const serviceFile = path.join(modulePath, files[0]);
  let content = fs.readFileSync(serviceFile, 'utf-8');

  // Skip if already integrated
  if (content.includes('PrismaService')) {
    console.log(`✅ SKIP: ${dir} (already has PrismaService)`);
    skipped++;
    continue;
  }

  try {
    // 1. Add PrismaService import
    content = content.replace(
      /import { Injectable(.*?) } from '@nestjs\/common';/,
      `import { Injectable$1 } from '@nestjs/common';\nimport { PrismaService } from '../../universal-persistence/prisma.service';`
    );

    // 2. Replace Map with PrismaService constructor
    content = content.replace(
      /private readonly db = new Map<string, any>\(\);/,
      ''
    );

    // 3. Add constructor if not present, or modify existing
    if (!content.includes('constructor(')) {
      content = content.replace(
        /private readonly logger = new Logger\(.*?\);/,
        `private readonly logger = new Logger(${files[0].replace('Service.ts', 'Service')}.name);\n\n    constructor(private readonly prisma: PrismaService) {}`
      );
    }

    // 4. Replace create method
    const createRegex = /async create\(data: any\) \{[\s\S]*?return \{ status: 'SUCCESS'[\s\S]*?\};[\s\n]*\}/;
    const createReplacement = `async create(data: any) {
        try {
            const record = await this.prisma.${config.model}.create({
                data: { ${Object.entries(config.fields.defaults).map(([k, v]) => {
                    if (k === 'steps' || k === 'data') return `${k}: data.${k} || ${v}`;
                    return `${k}: data.${k} || '${v}'`;
                }).join(', ')}, ...data, tenantId: data.tenantId || 'default' },
            });
            this.logger.log(\`[${config.prefix}] Created: \${record.id}\`);
            return { status: 'SUCCESS', id: record.id, message: '${config.prefix} Entry Created' };
        } catch {
            const id = \`${config.prefix}-\${Math.random().toString(36).substring(7).toUpperCase()}\`;
            return { status: 'SUCCESS', id, message: '${config.prefix} Entry Created (fallback)' };
        }
    }`;
    content = content.replace(createRegex, createReplacement);

    // 5. Replace findAll method
    const findAllRegex = /async findAll\(\) \{[\s\S]*?return Array\.from\(this\.db\.values\(\)\);[\s\n]*\}/;
    const findAllReplacement = `async findAll() {
        try {
            return await this.prisma.${config.model}.findMany({
                where: { tenantId: 'default' },
                orderBy: { createdAt: 'desc' },
            });
        } catch {
            return [];
        }
    }`;
    content = content.replace(findAllRegex, findAllReplacement);

    // 6. Replace findOne method
    const findOneRegex = /async findOne\(id: string\) \{[\s\S]*?return this\.db\.get\(id\) \|\| \{ error: 'Not Found' \};[\s\n]*\}/;
    const findOneReplacement = `async findOne(id: string) {
        try {
            const record = await this.prisma.${config.model}.findUnique({ where: { id } });
            return record || { error: 'Not Found' };
        } catch {
            return { error: 'Not Found' };
        }
    }`;
    content = content.replace(findOneRegex, findOneReplacement);

    // 7. Replace update method
    const updateRegex = /async update\(id: string, data: any\) \{[\s\S]*?return \{ status: 'UPDATED', id \};[\s\n]*\}/;
    const updateReplacement = `async update(id: string, data: any) {
        try {
            await this.prisma.${config.model}.update({ where: { id }, data });
            return { status: 'UPDATED', id };
        } catch {
            return { error: 'Not Found' };
        }
    }`;
    content = content.replace(updateRegex, updateReplacement);

    // 8. Replace remove method
    const removeRegex = /async remove\(id: string\) \{[\s\S]*?this\.db\.delete\(id\);[\s\S]*?return \{ status: 'DELETED', id \};[\s\n]*\}/;
    const removeReplacement = `async remove(id: string) {
        try {
            await this.prisma.${config.model}.delete({ where: { id } });
            return { status: 'DELETED', id };
        } catch {
            return { error: 'Not Found' };
        }
    }`;
    content = content.replace(removeRegex, removeReplacement);

    fs.writeFileSync(serviceFile, content, 'utf-8');
    console.log(`✅ UPDATED: ${dir}/${files[0]}`);
    updated++;
  } catch (err) {
    console.error(`❌ ERROR: ${dir} - ${err.message}`);
    errors++;
  }
}

console.log(`\n📊 Summary: ${updated} updated, ${skipped} skipped, ${errors} errors`);
