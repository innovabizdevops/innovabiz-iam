const fs = require('fs');
const path = require('path');

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
let c = fs.readFileSync(NAV_FILE, 'utf8');

// Replace undefined icons in createElement calls
c = c.replace(/React\.createElement\(BuildingEstate\b/g, 'React.createElement(Building2');
c = c.replace(/React\.createElement\(Spray\b/g, 'React.createElement(Droplets');

// Remove BuildingEstate and Spray from import block (they might be aliases)
// The import has "Building2 as BuildingEstate" — we need to remove the alias
c = c.replace(/Building2\s+as\s+BuildingEstate,?\s*/g, '');
c = c.replace(/\bSpray,?\s*/g, '');

// Also remove from import if standalone
c = c.replace(/\bBuildingEstate,?\s*/g, '');

fs.writeFileSync(NAV_FILE, c);
console.log('✅ Replaced BuildingEstate → Building2, Spray → Droplets');

// Verify
const lucide = require(path.join(__dirname, '..', 'frontend', 'node_modules', 'lucide-react'));
const final = fs.readFileSync(NAV_FILE, 'utf8');
const used = new Set();
const re = /React\.createElement\((\w+)/g;
let m;
while ((m = re.exec(final)) !== null) used.add(m[1]);

const bad = [...used].filter(i => !lucide[i]);
console.log('Icons used:', used.size);
console.log('Remaining undefined:', bad.length);
if (bad.length > 0) bad.forEach(i => console.log('  ❌', i));
else console.log('✅ All icons valid!');
