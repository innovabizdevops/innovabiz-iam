const fs = require('fs');
const path = require('path');

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
let c = fs.readFileSync(NAV_FILE, 'utf8');

// Count occurrences before fixing
const be = (c.match(/BuildingEstate/g) || []).length;
const sp = (c.match(/\bSpray\b/g) || []).length;
console.log('BuildingEstate occurrences:', be);
console.log('Spray occurrences:', sp);

// Replace in createElement calls
c = c.replace(/React\.createElement\(BuildingEstate/g, 'React.createElement(Building2');
c = c.replace(/React\.createElement\(Spray/g, 'React.createElement(Droplets');

// Fix the import: remove "Building2 as BuildingEstate" alias and standalone Spray
c = c.replace(/,?\s*Building2\s+as\s+BuildingEstate/g, '');
c = c.replace(/,?\s*Spray\b/g, '');

fs.writeFileSync(NAV_FILE, c);

// Verify balance
let b = 0;
for (const ch of c) {
    if (ch === '{') b++;
    if (ch === '}') b--;
}

const remaining_be = (c.match(/BuildingEstate/g) || []).length;
const remaining_sp = (c.match(/\bSpray\b/g) || []).length;

console.log('\n✅ Done!');
console.log('BuildingEstate remaining:', remaining_be);
console.log('Spray remaining:', remaining_sp);
console.log('Brace balance:', b, b === 0 ? '✅' : '❌');
console.log('Lines:', c.split('\n').length);
