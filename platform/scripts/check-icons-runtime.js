/**
 * Check which lucide-react icons are actually available at runtime.
 * This script requires the actual lucide-react package and checks each imported icon.
 */
const fs = require('fs');
const path = require('path');

// Load lucide-react from the project's node_modules
const lucidePath = path.join(__dirname, '..', 'frontend', 'node_modules', 'lucide-react');
let lucide;
try {
    lucide = require(lucidePath);
} catch (e) {
    console.error('Cannot load lucide-react:', e.message);
    process.exit(1);
}

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
const content = fs.readFileSync(NAV_FILE, 'utf8');

// Find all icons used in React.createElement calls
const usedIcons = new Set();
const createRe = /React\.createElement\((\w+)/g;
let m;
while ((m = createRe.exec(content)) !== null) {
    usedIcons.add(m[1]);
}

// Check each used icon against lucide-react exports
const undefinedIcons = [];
const validIcons = [];
for (const icon of [...usedIcons].sort()) {
    if (lucide[icon]) {
        validIcons.push(icon);
    } else {
        undefinedIcons.push(icon);
    }
}

console.log(`Total icons used: ${usedIcons.size}`);
console.log(`Valid icons: ${validIcons.length}`);
console.log(`UNDEFINED icons: ${undefinedIcons.length}`);

if (undefinedIcons.length > 0) {
    console.log('\n❌ These icons do NOT exist in lucide-react:');
    undefinedIcons.forEach(i => console.log(`  - ${i}`));

    // Suggest replacements from available icons
    console.log('\n📋 Available icon count:', Object.keys(lucide).filter(k => typeof lucide[k] === 'object' || typeof lucide[k] === 'function').length);
}

// Also list all available lucide exports that look like components (PascalCase)
const available = Object.keys(lucide).filter(k => /^[A-Z]/.test(k) && k !== 'default' && k !== 'icons' && k !== 'createLucideIcon');
console.log(`\nTotal available lucide-react icons: ${available.length}`);
