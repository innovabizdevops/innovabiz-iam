#!/usr/bin/env node
/**
 * ABSOLUTE FINAL — Phases 351-356 → SURPASS 3500+ ABSOLUTELY
 * 6 phases × 2 modules × 10 subs = 120 new → ~3500+
 */
const fs = require("fs");
const path = require("path");
const ROOT = path.resolve(__dirname, "..");
const NAV = path.join(ROOT, "frontend/src/components/layout/nav-registry.ts");
const PAGES = path.join(ROOT, "frontend/src/app/(dashboard)/modules");

const PHASES = {
    351: {
        label: "Funeral Insurance & Pre-Need", modules: [
            { id: "funeral-insurance", title: "Funeral Insurance", accent: "#78350f", subs: ["pre-need-plan", "premium-collect", "beneficiary-reg", "claim-process-fun", "underwriting-fun", "agent-network-fun", "commission-fun", "policy-admin-fun", "actuarial-fun", "compliance-fun"] },
            { id: "memorial-services", title: "Memorial Services", accent: "#92400e", subs: ["ceremony-plan", "floral-arrange", "catering-memorial", "music-selection", "funeral-live-stream", "grief-counseling", "legacy-plan", "monument-engrave", "urn-selection", "obituary-mgmt"] },
        ]
    },
    352: {
        label: "Esports & Competitive Gaming", modules: [
            { id: "esports-org", title: "Esports Organization", accent: "#7c3aed", subs: ["team-roster", "player-contract", "tournament-register", "scrimmage-schedule", "coach-staff", "performance-analytics-esport", "sponsor-deal", "content-creation-esport", "fan-base", "brand-collab"] },
            { id: "esports-league", title: "Esports League", accent: "#6d28d9", subs: ["league-structure", "match-schedule", "bracket-system", "referee-assign", "anti-cheat", "broadcast-produce", "prize-distribute", "transfer-window", "fines-discipline", "integrity-unit"] },
        ]
    },
    353: {
        label: "Renewable Energy Projects", modules: [
            { id: "solar-project", title: "Solar Project", accent: "#f59e0b", subs: ["site-assess-solar", "panel-procure", "inverter-select", "mounting-system", "grid-connect", "commissioning-solar", "performance-monitor-solar", "cleaning-schedule-solar", "warranty-track-solar", "decommission-solar"] },
            { id: "wind-project", title: "Wind Project", accent: "#d97706", subs: ["wind-assess", "turbine-procure", "foundation-design", "crane-ops", "cable-install", "grid-connect-wind", "scada-wind", "blade-inspect", "gearbox-maintain", "environmental-wind"] },
        ]
    },
    354: {
        label: "Biomedical Equipment Mgmt", modules: [
            { id: "biomed-mgmt", title: "Biomedical Equipment", accent: "#ef4444", subs: ["asset-register-biomed", "pm-schedule", "calibration-biomed", "repair-track", "vendor-service-biomed", "safety-test", "utilization-biomed", "lifecycle-biomed", "parts-inventory-biomed", "recall-alert"] },
            { id: "clinical-engineering", title: "Clinical Engineering", accent: "#dc2626", subs: ["technology-plan-clin", "acquisition-eval", "installation-val", "user-training", "incident-report-biomed", "hazard-alert", "interoperability", "cybersecurity-device", "replacement-plan", "roi-analysis-clin"] },
        ]
    },
    355: {
        label: "Election & Voting Systems", modules: [
            { id: "election-mgmt", title: "Election Mgmt", accent: "#4338ca", subs: ["voter-register", "polling-station", "ballot-design", "candidate-filing", "campaign-finance", "election-staff", "early-voting", "absentee-ballot", "result-tabulate", "audit-recount"] },
            { id: "civic-engagement", title: "Civic Engagement", accent: "#3730a3", subs: ["petition-mgmt", "public-hearing", "budget-participatory", "citizen-feedback", "transparency-portal", "foia-request", "lobbyist-register", "community-board", "ward-mgmt", "constituent-service"] },
        ]
    },
    356: {
        label: "Cruise & Maritime Tourism", modules: [
            { id: "cruise-ops", title: "Cruise Operations", accent: "#0891b2", subs: ["itinerary-plan", "port-call", "passenger-embark", "cabin-management", "dining-cruise", "entertainment-cruise", "shore-excursion", "medical-cruise", "safety-drill", "environmental-cruise"] },
            { id: "maritime-tourism", title: "Maritime Tourism", accent: "#0e7490", subs: ["charter-yacht", "dive-operation", "marine-safari", "island-hop", "boat-rental", "fishing-charter", "whale-watch", "kayak-tour", "snorkel-trip", "marina-service"] },
        ]
    },
};

function richPage(fn, t, m) { return `"use client";\nimport React from "react";\nimport { BarChart3,TrendingUp,Clock,CheckCircle2,Search,Filter,Download,Plus,MoreHorizontal,ArrowUpRight,ArrowDownRight } from "lucide-react";\n\nconst PAGE="${t}";\nconst KPIS=[{l:"Total ${t}",v:"2,847",d:"+12.3%",u:true},{l:"${t} Ativos",v:"1,923",d:"+5.1%",u:true},{l:"Pendentes",v:"124",d:"-8.4%",u:false},{l:"Score (%)",v:"94.2%",d:"+2.1%",u:true}];\nconst COLS=["ID","Nome","Categoria","Responsável","Data","Estado"];\nconst ROWS=[{id:"001",c:["${t}-001","Alpha","A","Ana Costa","2026-02-20","Ativo"]},{id:"002",c:["${t}-002","Beta","B","João Silva","2026-02-19","Em Revisão"]},{id:"003",c:["${t}-003","Gamma","A","Maria Santos","2026-02-18","Concluído"]},{id:"004",c:["${t}-004","Delta","C","Pedro Oliveira","2026-02-17","Pendente"]},{id:"005",c:["${t}-005","Epsilon","B","Carla Mendes","2026-02-16","Ativo"]}];\nconst CHART=[65,78,52,90,85,72,91,68,83,76,89,94];\nconst ACTIVITY=[{t:"Novo em ${t}",ts:"2min",a:"Sistema"},{t:"Aprovação #247",ts:"15min",a:"Ana"},{t:"Relatório",ts:"1h",a:"Auto"},{t:"Update",ts:"3h",a:"João"}];\n\nexport default function ${fn}(){return(<div className="p-6 space-y-6 max-w-[1600px] mx-auto"><div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4"><div><h1 className="text-2xl font-bold tracking-tight">{PAGE}</h1><p className="text-sm text-muted-foreground">${m} — iBOS</p></div><div className="flex gap-2"><button className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-sm hover:bg-accent"><Download className="h-3.5 w-3.5"/> Exportar</button><button className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary text-primary-foreground text-sm hover:bg-primary/90"><Plus className="h-3.5 w-3.5"/> Novo</button></div></div><div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">{KPIS.map((k,i)=>(<div key={i} className="rounded-xl border bg-card p-4 hover:shadow-md transition-shadow"><div className="flex items-center justify-between"><p className="text-xs font-medium text-muted-foreground">{k.l}</p>{i===0?<BarChart3 className="h-4 w-4 text-muted-foreground"/>:i===1?<TrendingUp className="h-4 w-4 text-muted-foreground"/>:i===2?<Clock className="h-4 w-4 text-muted-foreground"/>:<CheckCircle2 className="h-4 w-4 text-muted-foreground"/>}</div><p className="text-2xl font-bold mt-2">{k.v}</p><div className="flex items-center gap-1 mt-1">{k.u?<ArrowUpRight className="h-3 w-3 text-emerald-500"/>:<ArrowDownRight className="h-3 w-3 text-red-500"/>}<span className={\`text-xs font-medium \${k.u?"text-emerald-500":"text-red-500"}\`}>{k.d}</span></div></div>))}</div><div className="grid grid-cols-1 lg:grid-cols-3 gap-4"><div className="lg:col-span-2 rounded-xl border bg-card p-4"><h3 className="text-sm font-semibold mb-4">Evolução Mensal</h3><div className="flex items-end gap-1.5 h-32">{CHART.map((v,i)=>(<div key={i} className="flex-1 rounded-t hover:opacity-80" style={{height:\`\${v}%\`,background:\`hsl(\${200+i*10},70%,50%)\`}}/>))}</div></div><div className="rounded-xl border bg-card p-4"><h3 className="text-sm font-semibold mb-3">Atividade</h3><div className="space-y-3">{ACTIVITY.map((a,i)=>(<div key={i} className="flex gap-3 items-start"><div className="h-2 w-2 rounded-full bg-primary mt-1.5 shrink-0"/><div><p className="text-xs">{a.t}</p><p className="text-[10px] text-muted-foreground">{a.a} · {a.ts}</p></div></div>))}</div></div></div><div className="rounded-xl border bg-card"><div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b"><h3 className="text-sm font-semibold">Registos de {PAGE}</h3><div className="flex gap-2"><div className="relative"><Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground"/><input placeholder="Pesquisar..." className="pl-8 pr-3 py-1.5 rounded-lg border text-sm w-48 bg-transparent"/></div><button className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg border text-sm hover:bg-accent"><Filter className="h-3.5 w-3.5"/> Filtrar</button></div></div><div className="overflow-x-auto"><table className="w-full"><thead><tr className="border-b bg-muted/30">{COLS.map(c=>(<th key={c} className="text-left text-xs font-medium text-muted-foreground px-4 py-2.5">{c}</th>))}<th className="w-10"/></tr></thead><tbody>{ROWS.map(r=>(<tr key={r.id} className="border-b last:border-0 hover:bg-muted/20">{r.c.map((cell,j)=>(<td key={j} className="px-4 py-2.5 text-sm">{j===r.c.length-1?(<span className={\`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium \${cell==="Ativo"?"bg-emerald-500/10 text-emerald-500":cell==="Concluído"?"bg-blue-500/10 text-blue-500":cell==="Pendente"?"bg-amber-500/10 text-amber-500":"bg-purple-500/10 text-purple-500"}\`}>{cell}</span>):cell}</td>))}<td className="px-2"><button className="p-1 hover:bg-accent rounded"><MoreHorizontal className="h-4 w-4 text-muted-foreground"/></button></td></tr>))}</tbody></table></div><div className="p-3 border-t flex items-center justify-between text-xs text-muted-foreground"><span>1-5 de {ROWS.length}</span><div className="flex gap-1"><button className="px-2 py-1 rounded border hover:bg-accent">Ant</button><button className="px-2 py-1 rounded border bg-primary text-primary-foreground">1</button><button className="px-2 py-1 rounded border hover:bg-accent">Seg</button></div></div></div></div>);}\n`; }

let navSrc = fs.readFileSync(NAV, "utf8");
const ids = new Set(); let mx; const re = /id:\s*"([^"]+)"/g;
while ((mx = re.exec(navSrc)) !== null) ids.add(mx[1]);
console.log("[1] Nav: " + ids.size + " modules");
const ICONS = ["Shield", "Gamepad2", "Sun", "Wind", "Stethoscope", "Vote", "Ship", "Flower", "Swords", "MapPin",
    "Heart", "Monitor", "Zap", "Anchor", "Microscope", "Scale", "Sailboat", "Star", "Flag", "Compass"];
let n = 0, d = 0, p = 0; const cb = navSrc.lastIndexOf("];");
for (const [ph, data] of Object.entries(PHASES)) {
    const f = parseInt(ph); process.stdout.write("F" + f + ": " + data.label);
    for (const mod of data.modules) {
        if (ids.has(mod.id)) { d += mod.subs.length; continue; }
        const ic = ICONS[n % ICONS.length];
        const items = mod.subs.map(s => `            { title: "${s.replace(/-/g, " ").replace(/\b\w/g, c => c.toUpperCase())}", href: "/modules/${mod.id}/${s}" },`).join("\n");
        navSrc = navSrc.substring(0, cb) + `\n    /* ── F${f}: ${mod.title} ── */\n    { id: "${mod.id}", title: "${mod.title}", icon: React.createElement(${ic}, { className: "h-4 w-4" }), href: "/modules/${mod.id}", accent: "${mod.accent}", items: [\n            { title: "${mod.title} Dashboard", href: "/modules/${mod.id}" },\n${items}\n        ] },\n` + navSrc.substring(cb);
        ids.add(mod.id); n++;
        const md2 = path.join(PAGES, mod.id); if (!fs.existsSync(md2)) fs.mkdirSync(md2, { recursive: true });
        const dp = path.join(md2, "page.tsx");
        if (!fs.existsSync(dp)) { fs.writeFileSync(dp, richPage(mod.id.replace(/-./g, x => x[1].toUpperCase()).replace(/^./, c => c.toUpperCase()) + "Page", mod.title, mod.title)); p++; }
        for (const sub of mod.subs) {
            const sd = path.join(md2, sub); if (!fs.existsSync(sd)) fs.mkdirSync(sd, { recursive: true });
            const sp = path.join(sd, "page.tsx");
            if (!fs.existsSync(sp)) { fs.writeFileSync(sp, richPage(sub.replace(/-./g, x => x[1].toUpperCase()).replace(/^./, c => c.toUpperCase()) + "Page", sub.replace(/-/g, " ").replace(/\b\w/g, c => c.toUpperCase()), mod.title)); p++; }
        }
    }
    console.log(" ✓");
}
fs.writeFileSync(NAV, navSrc);
console.log("\n╔════════════════════════════════════════╗");
console.log("║  🎯 Total: " + ids.size + " modules");
console.log("║  New: " + n + " | Dupes: " + d + " | Pages: " + p);
console.log("║  META 3500+ → " + (ids.size >= 3500 ? "✅ ALCANÇADA!" : "❌ Faltam " + (3500 - ids.size)));
console.log("╚════════════════════════════════════════╝");
