/**
 * Re-register missing modules in nav-registry.ts
 * Finds modules with filesystem dirs but no nav-registry entry, then adds them.
 */
const fs = require('fs');
const path = require('path');

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
const MOD_DIR = path.join(__dirname, '..', 'frontend', 'src', 'app', '(dashboard)', 'modules');

// ── Parse existing registered module IDs ──
const navContent = fs.readFileSync(NAV_FILE, 'utf8');
const registeredIds = new Set((navContent.match(/id:\s*"([^"]+)"/g) || []).map(m => m.match(/"([^"]+)"/)[1]));

// ── Find filesystem module dirs ──
const fsDirs = fs.readdirSync(MOD_DIR, { withFileTypes: true })
    .filter(d => d.isDirectory())
    .map(d => d.name);

// ── Identify missing modules ──
const missing = fsDirs.filter(d => !registeredIds.has(d));
console.log(`Registered: ${registeredIds.size}, FS dirs: ${fsDirs.length}, Missing: ${missing.length}`);

// ── Icon rotation ──
const ICONS = [
    'Box', 'Layers', 'Package', 'Cog', 'Wrench', 'Database', 'Server', 'Monitor',
    'Globe', 'Zap', 'Activity', 'TrendingUp', 'BarChart3', 'PieChart', 'Target',
    'Award', 'Star', 'Shield', 'Lock', 'Key', 'Settings', 'Sliders', 'Filter',
    'Folder', 'FileText', 'ClipboardList', 'Puzzle', 'Cpu', 'Wifi', 'Radio',
    'Gauge', 'ScanLine', 'Heart', 'Users', 'Building2', 'Factory', 'Hammer',
    'Truck', 'Briefcase', 'BookOpen', 'GraduationCap', 'Stethoscope', 'Leaf',
    'Droplets', 'Flame', 'Wind', 'Mountain', 'Anchor', 'Compass', 'Map',
];

// ── Accent color rotation ──
const ACCENTS = [
    '#6366f1', '#8b5cf6', '#7c3aed', '#ec4899', '#f43f5e', '#ef4444', '#f97316',
    '#f59e0b', '#84cc16', '#22c55e', '#10b981', '#14b8a6', '#06b6d4', '#0891b2',
    '#0ea5e9', '#3b82f6', '#2563eb', '#1e40af', '#4f46e5', '#9333ea', '#c026d3',
    '#db2777', '#be123c', '#b91c1c', '#c2410c', '#a16207', '#4d7c0f', '#15803d',
    '#047857', '#0f766e', '#0e7490', '#0284c7', '#4338ca', '#7c3aed', '#a855f7',
    '#d946ef', '#f472b6', '#fb923c', '#facc15', '#a3e635', '#34d399', '#2dd4bf',
    '#22d3ee', '#38bdf8', '#818cf8', '#c084fc', '#e879f9', '#fb7185',
    '#475569', '#78350f', '#b45309', '#059669', '#dc2626',
];

function titleCase(slug) {
    return slug.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
}

// ── For each missing module, find its sub-pages from the filesystem ──
function getSubPages(moduleSlug) {
    const modPath = path.join(MOD_DIR, moduleSlug);
    const subs = [];
    try {
        const entries = fs.readdirSync(modPath, { withFileTypes: true });
        for (const e of entries) {
            if (e.isDirectory()) {
                // Check if it has a page.tsx
                const pagePath = path.join(modPath, e.name, 'page.tsx');
                if (fs.existsSync(pagePath)) {
                    subs.push(e.name);
                }
            }
        }
    } catch (e) { /* ignore */ }
    return subs;
}

// ── Build nav entries for missing modules ──
const entries = [];
missing.sort();

for (let i = 0; i < missing.length; i++) {
    const slug = missing[i];
    const title = titleCase(slug);
    const icon = ICONS[i % ICONS.length];
    const accent = ACCENTS[i % ACCENTS.length];
    const subPages = getSubPages(slug);

    let itemsStr = `\n                    { title: "Dashboard", href: "/modules/${slug}" },`;
    for (const sub of subPages.sort()) {
        itemsStr += `\n                    { title: "${titleCase(sub)}", href: "/modules/${slug}/${sub}" },`;
    }

    entries.push(`{
    id: "${slug}", title: "${title}",
        icon: React.createElement(${icon}, { className: "h-4 w-4" }), href: "/modules/${slug}",
            accent: "${accent}",
                items: [${itemsStr}
                ],
    },`);
}

// ── Insert before the closing ]; of NAV_ITEMS ──
const insertionPoint = navContent.lastIndexOf('];');
if (insertionPoint === -1) {
    console.error('ERROR: Could not find ]; in nav-registry.ts');
    process.exit(1);
}

const newContent = navContent.slice(0, insertionPoint) + entries.join('\n') + '\n' + navContent.slice(insertionPoint);

fs.writeFileSync(NAV_FILE, newContent);

// Verify
const finalContent = fs.readFileSync(NAV_FILE, 'utf8');
const finalIds = (finalContent.match(/id:\s*"([^"]+)"/g) || []).map(m => m.match(/"([^"]+)"/)[1]);
const finalUnique = new Set(finalIds);

// Brace check
let b = 0;
for (const ch of finalContent) {
    if (ch === '{') b++;
    if (ch === '}') b--;
}

console.log(`\n✅ Done! Added ${entries.length} modules.`);
console.log(`Total registered: ${finalUnique.size}`);
console.log(`Brace balance: ${b} ${b === 0 ? '✅' : '❌'}`);
console.log(`Final line count: ${finalContent.split('\n').length}`);
