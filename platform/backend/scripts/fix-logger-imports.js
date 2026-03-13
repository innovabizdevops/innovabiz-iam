const fs = require('fs');
const path = require('path');
const dirs = ['open-data','open-education','open-health','open-innovation','open-insurance'];
const base = path.join(__dirname, '..', 'src', 'modules');
let fixed = 0;
for (const d of dirs) {
    const files = fs.readdirSync(path.join(base, d)).filter(f => f.endsWith('Service.ts') && !f.includes('.spec'));
    for (const f of files) {
        const fp = path.join(base, d, f);
        let c = fs.readFileSync(fp, 'utf-8');
        // Check if Logger is used but not in the import statement
        if (c.includes('new Logger(') && !c.match(/import\s*\{[^}]*Logger[^}]*\}/)) {
            c = c.replace(
                /import \{ Injectable \} from '@nestjs\/common';/,
                "import { Injectable, Logger } from '@nestjs/common';"
            );
            fs.writeFileSync(fp, c);
            console.log('FIXED:', d + '/' + f);
            fixed++;
        }
    }
}
console.log('Total fixed:', fixed);
