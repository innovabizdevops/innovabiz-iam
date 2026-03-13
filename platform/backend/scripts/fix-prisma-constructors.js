/**
 * Script v2 — Fix missing constructors in services that the first script missed
 * Some Open Ecosystem services don't have 'private readonly logger' pattern
 */
const fs = require('fs');
const path = require('path');

const MODULES_DIR = path.join(__dirname, '..', 'src', 'modules');

// Find all service files that have PrismaService import but no constructor
const modules = fs.readdirSync(MODULES_DIR, { withFileTypes: true })
    .filter(d => d.isDirectory())
    .map(d => d.name);

let fixed = 0;

for (const mod of modules) {
    const modDir = path.join(MODULES_DIR, mod);
    const serviceFiles = fs.readdirSync(modDir)
        .filter(f => f.endsWith('Service.ts') && !f.endsWith('.spec.ts'));

    for (const sf of serviceFiles) {
        const filePath = path.join(modDir, sf);
        let content = fs.readFileSync(filePath, 'utf-8');

        // Only fix files with PrismaService import but no constructor
        if (content.includes('PrismaService') && !content.includes('constructor(')) {
            const className = sf.replace('.ts', '');

            // Find the class body opening and add constructor + logger
            const classMatch = content.match(new RegExp(`export class ${className}\\s*\\{`));
            if (classMatch) {
                const idx = content.indexOf(classMatch[0]);
                const insertPoint = idx + classMatch[0].length;
                
                const injection = `
    private readonly logger = new Logger(${className}.name);

    constructor(private readonly prisma: PrismaService) {}
`;
                content = content.slice(0, insertPoint) + injection + content.slice(insertPoint);
                
                // Ensure Logger is imported
                if (!content.includes('Logger')) {
                    content = content.replace(
                        "import { Injectable } from '@nestjs/common';",
                        "import { Injectable, Logger } from '@nestjs/common';"
                    );
                }

                fs.writeFileSync(filePath, content, 'utf-8');
                console.log(`✅ FIXED: ${mod}/${sf}`);
                fixed++;
            }
        }
    }
}

console.log(`\n📊 Fixed: ${fixed} services`);
