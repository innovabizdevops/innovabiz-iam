/**
 * Script para actualizar todos os spec files com PrismaService mock injection
 * Adiciona import do PRISMA_MOCK_PROVIDER e inclui no array de providers
 * 
 * Uso: node scripts/update-specs-prisma-mock.js
 */
const fs = require('fs');
const path = require('path');

const MODULES_DIR = path.join(__dirname, '..', 'src', 'modules');
let updated = 0;
let skipped = 0;

// Find all spec files in modules
function findSpecs(dir) {
    const results = [];
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const e of entries) {
        const full = path.join(dir, e.name);
        if (e.isDirectory()) {
            results.push(...findSpecs(full));
        } else if (e.name.endsWith('.spec.ts') && !e.name.endsWith('.e2e-spec.ts')) {
            results.push(full);
        }
    }
    return results;
}

const specs = findSpecs(MODULES_DIR);

for (const specPath of specs) {
    let content = fs.readFileSync(specPath, 'utf-8');
    
    // Skip if already has PrismaService mock
    if (content.includes('PRISMA_MOCK_PROVIDER') || content.includes('PrismaService')) {
        console.log(`⏭️  SKIP: ${path.relative(MODULES_DIR, specPath)} (already has mock)`);
        skipped++;
        continue;
    }
    
    // Skip if no TestingModule (not a NestJS test)
    if (!content.includes('createTestingModule')) {
        console.log(`⏭️  SKIP: ${path.relative(MODULES_DIR, specPath)} (not NestJS test)`);
        skipped++;
        continue;
    }

    // Calculate relative path from spec to test-utils
    const specDir = path.dirname(specPath);
    let relPath = path.relative(specDir, path.join(MODULES_DIR, '..', 'test-utils', 'prisma-mock'));
    relPath = relPath.replace(/\\/g, '/');
    if (!relPath.startsWith('.')) relPath = './' + relPath;

    // 1. Add import for PRISMA_MOCK_PROVIDER after the first import block
    const importLine = `import { PRISMA_MOCK_PROVIDER } from '${relPath}';\n`;
    
    // Insert after the last import statement
    const importRegex = /^(import .+;\n)+/m;
    const importMatch = content.match(importRegex);
    if (importMatch) {
        const insertIdx = importMatch.index + importMatch[0].length;
        content = content.slice(0, insertIdx) + importLine + content.slice(insertIdx);
    }

    // 2. Add PRISMA_MOCK_PROVIDER to providers array
    content = content.replace(
        /providers:\s*\[([^\]]+)\]/,
        (match, providers) => {
            const trimmed = providers.trim();
            return `providers: [${trimmed}, PRISMA_MOCK_PROVIDER]`;
        }
    );

    fs.writeFileSync(specPath, content, 'utf-8');
    console.log(`✅ UPDATED: ${path.relative(MODULES_DIR, specPath)}`);
    updated++;
}

console.log(`\n📊 Summary: ${updated} updated, ${skipped} skipped`);
