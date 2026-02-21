/**
 * Fix non-existent lucide-react icons in nav-registry.ts
 * Replaces problematic icons with valid alternatives and fixes imports.
 */
const fs = require('fs');
const path = require('path');

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
let content = fs.readFileSync(NAV_FILE, 'utf8');

// ═══════════════════════════════════════════════════════════
// Icon replacements: non-existent → valid alternative
// ═══════════════════════════════════════════════════════════
const REPLACEMENTS = {
    'Palm': 'TreePine',          // Nature/tree variant
    'Blast': 'Zap',              // Explosion → Energy
    'Cup': 'Coffee',             // Cup → Coffee cup  
    'Popcorn': 'Cookie',         // Snack → Cookie
    'ArmchairIcon': 'Armchair',  // Remove Icon suffix
    'Theater': 'Drama',          // Theater → Drama masks
    'Bandage': 'Heart',          // Medical bandage → Heart
    'Drumstick': 'Beef',         // Food item
    'CakeSlice': 'Cookie',       // Food → Cookie
    'GlassWater': 'Droplets',    // Water glass → Droplets
    'Croissant': 'Wheat',        // Pastry → Wheat
    'Palmtree': 'TreePine',      // Palm tree → Pine tree
    'Ambulance': 'Siren',        // Emergency vehicle → Siren
    'Sliders': 'SlidersHorizontal',  // Alias fix
    'Filter': 'SlidersVertical',      // Alias fix
    'Folder': 'FolderOpen',           // Alias fix
};

// Apply replacements in React.createElement calls
let replacedCount = 0;
for (const [bad, good] of Object.entries(REPLACEMENTS)) {
    const re = new RegExp(`React\\.createElement\\(${bad}\\b`, 'g');
    const matches = content.match(re);
    if (matches) {
        replacedCount += matches.length;
        content = content.replace(re, `React.createElement(${good}`);
        console.log(`  Replaced ${bad} → ${good} (${matches.length} occurrences)`);
    }
}

// Also replace in import block if the bad icons are imported
for (const [bad, good] of Object.entries(REPLACEMENTS)) {
    // Remove from import if it's there and the good one is already imported
    const importRe = new RegExp(`\\b${bad}\\b,?\\s*`, 'g');
    // Simple approach: just ensure we're not importing non-existent ones
}

// ═══════════════════════════════════════════════════════════
// Now fix the import block to match actual usage
// ═══════════════════════════════════════════════════════════

// Collect all icons actually used
const usedIcons = new Set();
const createRe = /React\.createElement\((\w+)/g;
let m;
while ((m = createRe.exec(content)) !== null) {
    usedIcons.add(m[1]);
}

// Add Home since it's used for the Home icon
usedIcons.add('Home');

// Build proper import statement
const sortedIcons = [...usedIcons].sort();

// Format into readable lines of ~6-8 icons each
const iconLines = [];
let currentLine = [];
for (const icon of sortedIcons) {
    currentLine.push(icon);
    if (currentLine.length >= 8) {
        iconLines.push('    ' + currentLine.join(', ') + ',');
        currentLine = [];
    }
}
if (currentLine.length > 0) {
    iconLines.push('    ' + currentLine.join(', ') + ',');
}

const newImportBlock = `import {\n${iconLines.join('\n')}\n} from "lucide-react";`;

// Replace old import block
content = content.replace(
    /import\s*\{[\s\S]*?\}\s*from\s*["']lucide-react["'];/,
    newImportBlock
);

fs.writeFileSync(NAV_FILE, content);

// ═══════════════════════════════════════════════════════════
// Verify
// ═══════════════════════════════════════════════════════════
const final = fs.readFileSync(NAV_FILE, 'utf8');
const finalUsed = new Set();
const re2 = /React\.createElement\((\w+)/g;
while ((m = re2.exec(final)) !== null) finalUsed.add(m[1]);

const importMatch = final.match(/import\s*\{([\s\S]*?)\}\s*from\s*["']lucide-react["']/);
const importedNames = importMatch[1].split(',').map(s => s.trim()).filter(Boolean);
const importedSet = new Set(importedNames);

const stillMissing = [...finalUsed].filter(i => !importedSet.has(i));

console.log(`\n✅ Fixed ${replacedCount} icon references`);
console.log(`Icons used: ${finalUsed.size}`);
console.log(`Icons imported: ${importedSet.size}`);
console.log(`Still missing: ${stillMissing.length}`);
if (stillMissing.length > 0) {
    stillMissing.forEach(i => console.log('  ❌', i));
}

// Brace balance
let b = 0;
for (const ch of final) {
    if (ch === '{') b++;
    if (ch === '}') b--;
}
console.log(`Brace balance: ${b} ${b === 0 ? '✅' : '❌'}`);
console.log(`Total lines: ${final.split('\n').length}`);
