/**
 * Script para corrigir union type errors nos spec files
 * Os CRUD methods agora retornam Prisma types | { error }, criando union types
 * Este script adiciona `as any` casts para resolver TS2339
 */
const fs = require('fs');
const path = require('path');

const MODULES_DIR = path.join(__dirname, '..', 'src', 'modules');
let fixed = 0;

function findSpecs(dir) {
    const results = [];
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const e of entries) {
        const full = path.join(dir, e.name);
        if (e.isDirectory()) results.push(...findSpecs(full));
        else if (e.name.endsWith('.spec.ts') && !e.name.endsWith('.e2e-spec.ts')) results.push(full);
    }
    return results;
}

const specs = findSpecs(MODULES_DIR);

for (const specPath of specs) {
    let content = fs.readFileSync(specPath, 'utf-8');
    let changed = false;

    // Pattern 1: (await service.findOne(id)).error → ((await service.findOne(id)) as any).error
    // Pattern 2: (await service.findOne(id)).name → ((await service.findOne(id)) as any).name
    // Pattern 3: (await service.findOne(c.id)).n → ((await service.findOne(c.id)) as any).n

    // Fix findOne property access without as any
    const findOneRegex = /\(await service\.findOne\(([^)]+)\)\)\.(\w+)/g;
    if (findOneRegex.test(content) && !content.includes('(await service.findOne') || true) {
        const newContent = content.replace(
            /\(await service\.findOne\(([^)]+)\)\)\.(\w+)/g,
            (match) => {
                if (match.includes('as any')) return match;
                return match.replace(
                    /\(await service\.findOne\(([^)]+)\)\)/,
                    '((await service.findOne($1)) as any)'
                );
            }
        );
        if (newContent !== content) {
            content = newContent;
            changed = true;
        }
    }

    // Fix update property access
    const updateRegex = /\(await service\.update\(([^)]+(?:\([^)]*\))?[^)]*)\)\)\.(\w+)/g;
    if (updateRegex.test(content)) {
        content = content.replace(updateRegex, (match) => {
            if (match.includes('as any')) return match;
            return match.replace(
                /\(await service\.update\(([^)]+)\)\)/,
                '((await service.update($1)) as any)'
            );
        });
        changed = true;
    }

    if (changed) {
        fs.writeFileSync(specPath, content, 'utf-8');
        console.log(`✅ FIXED: ${path.relative(MODULES_DIR, specPath)}`);
        fixed++;
    }
}

console.log(`\n📊 Fixed: ${fixed} spec files`);
