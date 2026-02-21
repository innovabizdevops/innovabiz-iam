const fs = require('fs');
const path = require('path');

const navFile = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
const content = fs.readFileSync(navFile, 'utf8');

// Count id: "..." entries (each module has one)
const ids = content.match(/id:\s*"[^"]+"/g) || [];
console.log('Registered modules in nav-registry:', ids.length);

// Count unique ids
const uniqueIds = new Set(ids.map(m => m.match(/"([^"]+)"/)[1]));
console.log('Unique module IDs:', uniqueIds.size);

// Check for duplicates
if (ids.length !== uniqueIds.size) {
    const seen = {};
    ids.forEach(m => {
        const id = m.match(/"([^"]+)"/)[1];
        seen[id] = (seen[id] || 0) + 1;
    });
    const dups = Object.entries(seen).filter(([, c]) => c > 1);
    console.log('Duplicate IDs found:', dups.length);
    dups.slice(0, 10).forEach(([id, count]) => console.log('  -', id, '×', count));
}

// Count module directories
const modDir = path.join(__dirname, '..', 'frontend', 'src', 'app', '(dashboard)', 'modules');
const dirs = fs.readdirSync(modDir, { withFileTypes: true }).filter(d => d.isDirectory());
console.log('\nModule page directories:', dirs.length);

// Check missing registrations
const registeredIds = [...uniqueIds];
const dirNames = dirs.map(d => d.name);
const missingInNav = dirNames.filter(d => !registeredIds.includes(d));
const missingInFS = registeredIds.filter(id => !dirNames.includes(id));

console.log('Dirs missing from nav-registry:', missingInNav.length);
console.log('Nav entries missing from filesystem:', missingInFS.length);
