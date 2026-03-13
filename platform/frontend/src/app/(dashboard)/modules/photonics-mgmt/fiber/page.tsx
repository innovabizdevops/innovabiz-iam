"use client";
import React from "react";
import { BarChart3, Users, Activity, Shield, TrendingUp, TrendingDown, PlusCircle, Download, Filter, RefreshCcw, Search, ArrowUpRight, Eye, Settings, CheckCircle, Clock, AlertTriangle, Bell } from "lucide-react";

const ACCENT = "#b45309";
const MODULE = "Photonics Management";
const PAGE = "Telecom Fiber";
const KPIS = [{"l":"Total Photonics Mgmt","v":"8.043","t":"+9.3%","d":"up"},{"l":"Telecom Fiber Ativos","v":"5.692","t":"-1.0%","d":"down"},{"l":"Eficiência (%)","v":"75.6%","t":"+5.2%","d":"up"},{"l":"Score Performance","v":"83","t":"+5.0%","d":"up"}];
const COLS = ["ID","Telecom Fiber","Categoria","Responsável","Data","Estado"];
const ROWS = [{"id":"r0","cells":["Photonics Mgmt #001","Photonics Mgmt #002","8.351","8.844","2026-03-02","Ativo"],"status":"active"},{"id":"r1","cells":["Photonics Mgmt #002","Photonics Mgmt #003","3.891","3.360","2026-02-02","Pendente"],"status":"pending"},{"id":"r2","cells":["Photonics Mgmt #003","Photonics Mgmt #004","4.820","6.183","2026-02-05","Concluído"],"status":"done"},{"id":"r3","cells":["Photonics Mgmt #004","Photonics Mgmt #005","6.081","4.050","2026-03-21","Em Análise"],"status":"risk"},{"id":"r4","cells":["Photonics Mgmt #005","Photonics Mgmt #001","7.659","6.101","2026-01-20","Aprovado"],"status":"active"}];
const CHART = [{"label":"Jan","value":434},{"label":"Fev","value":508},{"label":"Mar","value":571},{"label":"Abr","value":383},{"label":"Mai","value":432},{"label":"Jun","value":495},{"label":"Jul","value":481},{"label":"Ago","value":697},{"label":"Set","value":655},{"label":"Out","value":434},{"label":"Nov","value":571},{"label":"Dez","value":691}];
const MX = 697;
const ICONS = [BarChart3, Users, Activity, Shield];
const ST_C: Record<string,string> = {active:"bg-emerald-500/10 text-emerald-400 border-emerald-500/20",pending:"bg-amber-500/10 text-amber-400 border-amber-500/20",done:"bg-blue-500/10 text-blue-400 border-blue-500/20",risk:"bg-red-500/10 text-red-400 border-red-500/20"};
const ST_L: Record<string,string> = {active:"Ativo",pending:"Pendente",done:"Concluído",risk:"Risco"};

export default function FiberPage() {
    return (
        <div className="p-6 space-y-6 min-h-screen">
            {/* Header */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                    <p className="text-xs text-[var(--text-tertiary,#64748b)] uppercase tracking-widest mb-1">{MODULE}</p>
                    <h1 className="text-2xl font-bold font-[Outfit] flex items-center">{PAGE}</h1>
                </div>
                <div className="flex items-center gap-2 flex-wrap">
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white shadow-lg hover:brightness-110" style={{background:ACCENT}}><PlusCircle className="w-4 h-4"/>Novo</button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-white/5 border border-white/10 hover:bg-white/10"><Download className="w-4 h-4"/>Exportar</button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-white/5 border border-white/10 hover:bg-white/10"><Filter className="w-4 h-4"/>Filtrar</button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-white/5 border border-white/10 hover:bg-white/10"><RefreshCcw className="w-4 h-4"/>Atualizar</button>
                </div>
            </div>
            {/* Search */}
            <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-tertiary,#64748b)]"/>
                <input type="text" placeholder={`Pesquisar em ${PAGE}...`} className="w-full pl-10 pr-4 py-2.5 rounded-lg border border-white/10 bg-white/5 text-sm focus:outline-none focus:border-white/20"/>
            </div>
            {/* KPIs */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {KPIS.map((k: any,i: number) => {const I=ICONS[i];const isUp=k.d==="up";const T=isUp?TrendingUp:TrendingDown;return(
                    <div key={k.l} className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5 hover:border-white/20 transition-all">
                        <div className="flex items-center justify-between mb-3">
                            <div className="p-2 rounded-lg" style={{background:ACCENT+"15"}}><I className="w-4 h-4" style={{color:ACCENT}}/></div>
                            <div className={`flex items-center gap-1 text-xs font-semibold ${isUp?"text-emerald-400":"text-red-400"}`}><T className="w-3.5 h-3.5"/><span>{k.t}</span></div>
                        </div>
                        <p className="text-2xl font-bold font-[Outfit]">{k.v}</p>
                        <p className="text-xs text-[var(--text-tertiary,#64748b)] mt-1 uppercase tracking-wider">{k.l}</p>
                    </div>
                );})}
            </div>
            {/* Main Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Table */}
                <div className="lg:col-span-2 rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm overflow-hidden">
                    <div className="px-5 py-4 border-b border-white/10 flex items-center justify-between">
                        <h2 className="text-base font-semibold font-[Outfit]">Registos — {PAGE}</h2>
                        <button className="text-xs flex items-center gap-1 hover:underline" style={{color:ACCENT}}>Ver todos <ArrowUpRight className="w-3 h-3"/></button>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead><tr className="border-b border-white/5">{COLS.map((c: string)=>(<th key={c} className="text-left px-5 py-3 text-xs uppercase tracking-wider text-[var(--text-tertiary,#64748b)] font-medium">{c}</th>))}<th className="text-left px-5 py-3 text-xs uppercase tracking-wider text-[var(--text-tertiary,#64748b)] font-medium">Estado</th></tr></thead>
                            <tbody>{ROWS.map((r: any)=>(<tr key={r.id} className="border-b border-white/5 hover:bg-white/[0.03] transition-colors cursor-pointer">{r.cells.map((c: string,ci: number)=>(<td key={ci} className={`px-5 py-3 ${ci===0?"font-medium":"text-[var(--text-secondary,#94a3b8)]"}`}>{c}</td>))}<td className="px-5 py-3"><span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${ST_C[r.status]}`}>{ST_L[r.status]}</span></td></tr>))}</tbody>
                        </table>
                    </div>
                    <div className="px-5 py-3 border-t border-white/5 flex items-center justify-between text-xs text-[var(--text-tertiary,#64748b)]">
                        <span>1-5 de 100</span>
                        <div className="flex gap-1">
                            <button className="px-3 py-1 rounded border border-white/10 hover:bg-white/5">Ant.</button>
                            <button className="px-3 py-1 rounded text-white" style={{background:ACCENT}}>1</button>
                            <button className="px-3 py-1 rounded border border-white/10 hover:bg-white/5">2</button>
                            <button className="px-3 py-1 rounded border border-white/10 hover:bg-white/5">Seg.</button>
                        </div>
                    </div>
                </div>
                {/* Right */}
                <div className="space-y-6">
                    <div className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5">
                        <h3 className="text-sm font-semibold font-[Outfit] mb-4">Evolução — {PAGE}</h3>
                        <div className="flex items-end gap-1.5 h-36">{CHART.map((b: any)=>(<div key={b.label} className="flex-1 flex flex-col items-center gap-1"><div className="w-full rounded-t-md transition-all hover:brightness-125" style={{height:`${(b.value/MX)*100}%`,background:`linear-gradient(to top,${ACCENT}40,${ACCENT})`}}/><span className="text-[9px] text-[var(--text-tertiary,#64748b)]">{b.label}</span></div>))}</div>
                    </div>
                    <div className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5">
                        <h3 className="text-sm font-semibold font-[Outfit] mb-3">Ações Rápidas</h3>
                        <div className="grid grid-cols-2 gap-2">
                            <button className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs bg-white/5 border border-white/10 hover:bg-white/10"><PlusCircle className="w-3.5 h-3.5" style={{color:ACCENT}}/>Criar</button>
                            <button className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs bg-white/5 border border-white/10 hover:bg-white/10"><Download className="w-3.5 h-3.5" style={{color:ACCENT}}/>Exportar</button>
                            <button className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs bg-white/5 border border-white/10 hover:bg-white/10"><Eye className="w-3.5 h-3.5" style={{color:ACCENT}}/>Ver</button>
                            <button className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs bg-white/5 border border-white/10 hover:bg-white/10"><Settings className="w-3.5 h-3.5" style={{color:ACCENT}}/>Config</button>
                        </div>
                    </div>
                    <div className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5">
                        <h3 className="text-sm font-semibold font-[Outfit] mb-3">Atividade</h3>
                        <div className="space-y-3">
                            <div className="flex items-start gap-3 text-sm"><CheckCircle className="w-4 h-4 mt-0.5 text-emerald-400 flex-shrink-0"/><div><p className="text-[var(--text-secondary,#94a3b8)]">Novo {PAGE} criado</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3"/>2m</p></div></div>
                            <div className="flex items-start gap-3 text-sm"><AlertTriangle className="w-4 h-4 mt-0.5 text-amber-400 flex-shrink-0"/><div><p className="text-[var(--text-secondary,#94a3b8)]">Aprovação pendente</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3"/>15m</p></div></div>
                            <div className="flex items-start gap-3 text-sm"><Activity className="w-4 h-4 mt-0.5 text-blue-400 flex-shrink-0"/><div><p className="text-[var(--text-secondary,#94a3b8)]">Tarefa concluída</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3"/>1h</p></div></div>
                            <div className="flex items-start gap-3 text-sm"><Bell className="w-4 h-4 mt-0.5 text-purple-400 flex-shrink-0"/><div><p className="text-[var(--text-secondary,#94a3b8)]">Alerta em {MODULE}</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3"/>3h</p></div></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
