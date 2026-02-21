const fs = require('fs');
const path = require('path');

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
const content = fs.readFileSync(NAV_FILE, 'utf8');

// Find all icons used in React.createElement calls
const usedIcons = new Set();
const createRe = /React\.createElement\((\w+)/g;
let m;
while ((m = createRe.exec(content)) !== null) {
    usedIcons.add(m[1]);
}

// Find all icons in the import block
const importMatch = content.match(/import\s*\{([\s\S]*?)\}\s*from\s*["']lucide-react["']/);
const importedRaw = importMatch[1].split(',').map(s => s.trim()).filter(Boolean);
const importedIcons = new Set();
const aliasMap = {};
for (const entry of importedRaw) {
    const parts = entry.split(/\s+as\s+/);
    importedIcons.add(parts[0].trim());
    if (parts[1]) {
        aliasMap[parts[0].trim()] = parts[1].trim();
    }
}

// Find used icons not in the import
const missing = [...usedIcons].filter(icon => {
    if (importedIcons.has(icon)) return false;
    // Check aliases
    for (const [orig, alias] of Object.entries(aliasMap)) {
        if (alias === icon) return false;
    }
    return true;
}).sort();

console.log('Used icons:', usedIcons.size);
console.log('Imported icons:', importedIcons.size);
console.log('Missing from imports:', missing.length);
missing.forEach(i => console.log('  +', i));

// Check which icons likely DON'T exist in lucide-react
// Known problematic icons from the error
const knownBad = ['Palm', 'Blast', 'Cup', 'Popcorn', 'ArmchairIcon', 'Theater', 'Bandage', 'Drumstick', 'CakeSlice', 'GlassWater', 'Croissant', 'Palmtree', 'Ambulance'];
const badInFile = knownBad.filter(i => usedIcons.has(i) || importedIcons.has(i));
console.log('\nPotentially non-existent icons used:', badInFile.length);
badInFile.forEach(i => console.log('  ⚠️', i));
