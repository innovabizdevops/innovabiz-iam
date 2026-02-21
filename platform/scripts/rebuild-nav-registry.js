/**
 * REBUILD NAV-REGISTRY.TS — Complete reconstruction from filesystem + existing data
 * 
 * Strategy:
 * 1. Parse existing file to extract module data (id, title, icon, href, accent, items)
 * 2. Scan filesystem module dirs to find all modules  
 * 3. For any module missing data, generate sensible defaults
 * 4. Output a clean, properly structured nav-registry.ts
 */
const fs = require('fs');
const path = require('path');

const NAV_FILE = path.join(__dirname, '..', 'frontend', 'src', 'components', 'layout', 'nav-registry.ts');
const MOD_DIR = path.join(__dirname, '..', 'frontend', 'src', 'app', '(dashboard)', 'modules');

// ═══════════════════════════════════════════════════════════════
// STEP 1: Parse existing file to extract module data
// ═══════════════════════════════════════════════════════════════
const content = fs.readFileSync(NAV_FILE, 'utf8');

// Extract the header (imports + types + start of NAV_ITEMS)
const headerEnd = content.indexOf('export const NAV_ITEMS: NavEntry[] = [');
const headerWithArrayStart = content.substring(0, headerEnd + 'export const NAV_ITEMS: NavEntry[] = ['.length);

// Extract all module blocks using regex — find id, title, icon component, href, accent, items
const moduleDataMap = new Map();

// Pattern to find module entries with all their data
// We'll use a line-by-line approach to be more robust
const lines = content.split('\n');

let currentModule = null;
let collectingItems = false;
let itemsBuffer = [];
let braceDepth = 0;

for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();

    // Detect module ID line
    const idMatch = line.match(/id:\s*"([^"]+)",\s*title:\s*"([^"]+)"/);
    if (idMatch) {
        // Save previous module if exists
        if (currentModule && currentModule.id) {
            if (!moduleDataMap.has(currentModule.id)) {
                moduleDataMap.set(currentModule.id, { ...currentModule });
            } else {
                // Merge: prefer existing data but fill gaps
                const existing = moduleDataMap.get(currentModule.id);
                if (!existing.accent && currentModule.accent) existing.accent = currentModule.accent;
                if ((!existing.items || existing.items.length === 0) && currentModule.items && currentModule.items.length > 0) {
                    existing.items = currentModule.items;
                }
                if (!existing.iconComponent && currentModule.iconComponent) existing.iconComponent = currentModule.iconComponent;
            }
        }

        currentModule = {
            id: idMatch[1],
            title: idMatch[2],
            iconComponent: null,
            accent: null,
            items: [],
        };
        collectingItems = false;
        itemsBuffer = [];
    }

    // Detect icon line
    if (currentModule) {
        const iconMatch = line.match(/icon:\s*React\.createElement\((\w+)/);
        if (iconMatch && !currentModule.iconComponent) {
            currentModule.iconComponent = iconMatch[1];
        }

        // Detect accent
        const accentMatch = line.match(/accent:\s*"(#[0-9a-fA-F]{6})"/);
        if (accentMatch && !currentModule.accent) {
            currentModule.accent = accentMatch[1];
        }

        // Detect items array
        if (line.includes('items: [') || line === 'items: [') {
            collectingItems = true;
            itemsBuffer = [];
        }

        if (collectingItems) {
            // Match sub-item entries
            const subItemMatch = line.match(/\{\s*title:\s*"([^"]+)",\s*href:\s*"([^"]+)"(?:,\s*tag:\s*"([^"]+)")?\s*\}/);
            if (subItemMatch) {
                const item = { title: subItemMatch[1], href: subItemMatch[2] };
                if (subItemMatch[3]) item.tag = subItemMatch[3];
                itemsBuffer.push(item);
            }

            if (line.includes('],')) {
                collectingItems = false;
                if (itemsBuffer.length > 0 && currentModule) {
                    // Only assign if this items set belongs to the current module (check hrefs)
                    const moduleSlug = currentModule.id;
                    const belongsToModule = itemsBuffer.some(it => it.href.includes(`/modules/${moduleSlug}`));
                    if (belongsToModule) {
                        currentModule.items = [...itemsBuffer];
                    }
                }
            }
        }
    }
}

// Save last module
if (currentModule && currentModule.id) {
    if (!moduleDataMap.has(currentModule.id)) {
        moduleDataMap.set(currentModule.id, { ...currentModule });
    }
}

console.log(`Parsed ${moduleDataMap.size} unique modules from existing file`);

// ═══════════════════════════════════════════════════════════════
// STEP 2: Scan filesystem for ALL module directories + sub-pages
// ═══════════════════════════════════════════════════════════════
const fsDirs = fs.readdirSync(MOD_DIR, { withFileTypes: true })
    .filter(d => d.isDirectory())
    .map(d => d.name)
    .sort();

console.log(`Found ${fsDirs.length} module directories in filesystem`);

function getSubPages(moduleSlug) {
    const modPath = path.join(MOD_DIR, moduleSlug);
    const subs = [];
    try {
        const entries = fs.readdirSync(modPath, { withFileTypes: true });
        for (const e of entries) {
            if (e.isDirectory()) {
                const pagePath = path.join(modPath, e.name, 'page.tsx');
                if (fs.existsSync(pagePath)) {
                    subs.push(e.name);
                }
            }
        }
    } catch (e) { /* ignore */ }
    return subs.sort();
}

// ═══════════════════════════════════════════════════════════════
// STEP 3: Icon and accent defaults
// ═══════════════════════════════════════════════════════════════
const ICONS = [
    'Box', 'Layers', 'Package', 'Cog', 'Wrench', 'Database', 'Server', 'Monitor',
    'Globe', 'Zap', 'Activity', 'TrendingUp', 'BarChart3', 'PieChart', 'Target',
    'Award', 'Star', 'Shield', 'Lock', 'Key', 'Settings', 'Sliders', 'Filter',
    'Folder', 'FileText', 'ClipboardList', 'Puzzle', 'Cpu', 'Wifi', 'Radio',
    'Gauge', 'ScanLine', 'Heart', 'Users', 'Building2', 'Factory', 'Hammer',
    'Truck', 'Briefcase', 'BookOpen', 'GraduationCap', 'Stethoscope', 'Leaf',
    'Droplets', 'Flame', 'Wind', 'Mountain', 'Anchor', 'Compass', 'Map',
];

const ACCENTS = [
    '#6366f1', '#8b5cf6', '#7c3aed', '#ec4899', '#f43f5e', '#ef4444', '#f97316',
    '#f59e0b', '#84cc16', '#22c55e', '#10b981', '#14b8a6', '#06b6d4', '#0891b2',
    '#0ea5e9', '#3b82f6', '#2563eb', '#1e40af', '#4f46e5', '#9333ea', '#c026d3',
    '#db2777', '#be123c', '#b91c1c', '#c2410c', '#a16207', '#4d7c0f', '#15803d',
    '#047857', '#0f766e', '#0e7490', '#0284c7', '#4338ca', '#a855f7',
    '#d946ef', '#f472b6', '#fb923c', '#facc15', '#a3e635', '#34d399', '#2dd4bf',
    '#22d3ee', '#38bdf8', '#818cf8', '#c084fc', '#e879f9', '#fb7185',
    '#475569', '#78350f', '#b45309', '#059669', '#dc2626',
];

function titleCase(slug) {
    return slug.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
}

// ═══════════════════════════════════════════════════════════════
// STEP 4: Build complete module list
// ═══════════════════════════════════════════════════════════════
const allModuleIds = new Set([...moduleDataMap.keys(), ...fsDirs]);
const sortedIds = [...allModuleIds].sort();

// Separate "home" — it goes first
const homeModule = moduleDataMap.get('home');

const moduleEntries = [];
let iconIdx = 0;
let accentIdx = 0;

for (const id of sortedIds) {
    if (id === 'home') continue; // handled separately

    const existing = moduleDataMap.get(id);
    const fsSubPages = getSubPages(id);

    const title = existing?.title || titleCase(id);
    const iconComponent = existing?.iconComponent || ICONS[iconIdx++ % ICONS.length];
    const accent = existing?.accent || ACCENTS[accentIdx++ % ACCENTS.length];

    // Build items: prefer existing parsed items, fall back to filesystem
    let items = [];
    if (existing?.items && existing.items.length > 0) {
        items = existing.items;
    } else {
        // Generate from filesystem
        items.push({ title: 'Dashboard', href: `/modules/${id}` });
        for (const sub of fsSubPages) {
            items.push({ title: titleCase(sub), href: `/modules/${id}/${sub}` });
        }
    }

    // Ensure Dashboard is always first
    if (items.length === 0 || !items[0].href.endsWith(`/modules/${id}`)) {
        items.unshift({ title: 'Dashboard', href: `/modules/${id}` });
    }

    moduleEntries.push({ id, title, iconComponent, accent, items });
}

console.log(`Total modules to write: ${moduleEntries.length} + home`);

// ═══════════════════════════════════════════════════════════════
// STEP 5: Generate clean nav-registry.ts
// ═══════════════════════════════════════════════════════════════

function formatModule(mod, indent = '    ') {
    let itemsStr = mod.items.map(item => {
        let s = `${indent}        { title: "${item.title}", href: "${item.href}"`;
        if (item.tag) s += `, tag: "${item.tag}"`;
        s += ' },';
        return s;
    }).join('\n');

    return `${indent}{
${indent}    id: "${mod.id}", title: "${mod.title}",
${indent}    icon: React.createElement(${mod.iconComponent}, { className: "h-4 w-4" }), href: "/modules/${mod.id}",
${indent}    accent: "${mod.accent}",
${indent}    items: [
${itemsStr}
${indent}    ],
${indent}},`;
}

function formatHome(mod) {
    return `    {
        id: "home", title: "Home",
        icon: React.createElement(Home, { className: "h-4 w-4" }), href: "/",
        accent: "#6366f1",
        items: [
            { title: "Dashboard", href: "/" },
            { title: "Command Center", href: "/command-center" },
        ],
    },`;
}

// Build the full file
let output = headerWithArrayStart + '\n';
output += '    /* ————————————— 🏠 HOME ————————————— */\n';
output += formatHome(homeModule) + '\n';

// Group modules alphabetically
output += '    /* ————————————— 📦 MODULES (A-Z) ————————————— */\n';

for (const mod of moduleEntries) {
    output += formatModule(mod) + '\n';
}

output += `];

/** Find the active module based on the current pathname */
export function findActiveModule(pathname: string): {
    module: NavItem; items: NavSubItem[];
} | null {
    for (const entry of NAV_ITEMS) {
        if ("groupLabel" in entry) continue;
        const navItem = entry as NavItem;
        if (pathname.startsWith(navItem.href)) {
            return { module: navItem, items: navItem.items || [] };
        }
    }
    return null;
}
`;

fs.writeFileSync(NAV_FILE, output);

// ═══════════════════════════════════════════════════════════════
// STEP 6: Verify
// ═══════════════════════════════════════════════════════════════
const finalContent = fs.readFileSync(NAV_FILE, 'utf8');
const finalLines = finalContent.split('\n');
let parens = 0, brackets = 0, braces = 0;
for (const ch of finalContent) {
    if (ch === '(') parens++;
    if (ch === ')') parens--;
    if (ch === '[') brackets++;
    if (ch === ']') brackets--;
    if (ch === '{') braces++;
    if (ch === '}') braces--;
}

const finalIds = (finalContent.match(/id:\s*"([^"]+)"/g) || []).map(m => m.match(/"([^"]+)"/)[1]);
const finalUnique = new Set(finalIds);

console.log('\n═══════════════════════════════════════');
console.log('✅ REBUILD COMPLETE');
console.log('═══════════════════════════════════════');
console.log(`Lines: ${finalLines.length}`);
console.log(`Modules: ${finalUnique.size}`);
console.log(`Parens balance: ${parens} ${parens === 0 ? '✅' : '❌'}`);
console.log(`Brackets balance: ${brackets} ${brackets === 0 ? '✅' : '❌'}`);
console.log(`Braces balance: ${braces} ${braces === 0 ? '✅' : '❌'}`);
console.log(`Duplicates: ${finalIds.length - finalUnique.size}`);

// Check all filesystem dirs are covered
const registeredSet = new Set(finalIds);
const fsMissing = fsDirs.filter(d => !registeredSet.has(d));
console.log(`FS dirs missing from nav: ${fsMissing.length}`);
if (fsMissing.length > 0) {
    console.log('  Missing:', fsMissing.slice(0, 10).join(', '), fsMissing.length > 10 ? `... +${fsMissing.length - 10} more` : '');
}
