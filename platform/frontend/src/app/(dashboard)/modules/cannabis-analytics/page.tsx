"use client";
import React from "react";
import {
    BarChart3, Users, DollarSign, Activity, TrendingUp, TrendingDown,
    Package, LayoutDashboard,
    ArrowUpRight, PlusCircle, Download, Filter, RefreshCcw,
    Clock, CheckCircle, AlertTriangle, type LucideIcon,
} from "lucide-react";

/* ═══ Cannabis Analytics — Dashboard ═══ */

const ACCENT = "#0d3b1e";

const KPIS = [
    { label: "Total Registos", value: "7603", trend: "-4%", trendDirection: "down" as const, icon: BarChart3 },
    { label: "Ativos", value: "3703", trend: "+23%", trendDirection: "up" as const, icon: Users },
    { label: "Receita Mensal", value: "€391K", trend: "+17%", trendDirection: "up" as const, icon: DollarSign },
    { label: "Eficiência", value: "72%", trend: "-11%", trendDirection: "down" as const, icon: Activity },
];

const TABLE_COLS = ["Referência","Descrição","Data","Valor"];

const TABLE_ROWS = [
    { id: "1431131703", cells: ["CAN-6703", "Cannabis Analytics Item 1", "16/4/2026", "€31 803"], status: "active" as const },
    { id: "1431135274", cells: ["CAN-1274", "Cannabis Analytics Item 2", "3/11/2026", "€35 374"], status: "pending" as const },
    { id: "1431138845", cells: ["CAN-4845", "Cannabis Analytics Item 3", "18/6/2026", "€38 945"], status: "completed" as const },
    { id: "1431142416", cells: ["CAN-8416", "Cannabis Analytics Item 4", "5/1/2026", "€42 516"], status: "warning" as const },
    { id: "1431145987", cells: ["CAN-2987", "Cannabis Analytics Item 5", "20/8/2026", "€46 087"], status: "active" as const },
];

const CHART = [{ label: "Jan", value: 43 }, { label: "Fev", value: 52 }, { label: "Mar", value: 61 }, { label: "Abr", value: 70 }, { label: "Mai", value: 79 }, { label: "Jun", value: 88 }];

const STATUS_CLS: Record<string, string> = {
    active: "bg-emerald-500/15 text-emerald-400 border-emerald-500/30",
    completed: "bg-blue-500/15 text-blue-400 border-blue-500/30",
    pending: "bg-amber-500/15 text-amber-400 border-amber-500/30",
    warning: "bg-orange-500/15 text-orange-400 border-orange-500/30",
    error: "bg-red-500/15 text-red-400 border-red-500/30",
};
const STATUS_LBL: Record<string, string> = {
    active: "Ativo", completed: "Concluído", pending: "Pendente", warning: "Atenção", error: "Erro",
};

export default function CannabisAnalyticsPage() {
    const mx = Math.max(...CHART.map(c => c.value), 1);
    return (
        <div className="min-h-screen bg-[var(--surface-primary,#0a0e1a)] text-[var(--text-primary,#e2e8f0)] p-6 space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between flex-wrap gap-4">
                <div className="flex items-center gap-4">
                    <div className="p-3 rounded-xl" style={{ background: ACCENT + "15" }}>
                        <LayoutDashboard className="w-7 h-7" style={{ color: ACCENT }} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold font-[Outfit] tracking-tight flex items-center">Cannabis Analytics</h1>
                        <p className="text-sm text-[var(--text-tertiary,#64748b)] mt-0.5">InnovaBiz • Cannabis Analytics</p>
                    </div>
                </div>
                <div className="flex items-center gap-2 flex-wrap">
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white shadow-lg hover:brightness-110" style={{ background: ACCENT }}>
                        <PlusCircle className="w-4 h-4" /> Novo
                    </button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-white/5 border border-white/10 text-[var(--text-secondary,#94a3b8)] hover:bg-white/10">
                        <Download className="w-4 h-4" /> Exportar
                    </button>
                    <button className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-white/5 border border-white/10 text-[var(--text-secondary,#94a3b8)] hover:bg-white/10">
                        <Filter className="w-4 h-4" /> Filtrar
                    </button>
                </div>
            </div>

            {/* KPIs */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {KPIS.map((kpi) => {
                    const KIcon = kpi.icon;
                    const isUp = kpi.trendDirection === "up";
                    const TIcon = isUp ? TrendingUp : TrendingDown;
                    return (
                        <div key={kpi.label} className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5 hover:border-white/20 transition-all">
                            <div className="flex items-center justify-between mb-3">
                                <div className="p-2 rounded-lg" style={{ background: ACCENT + "15" }}>
                                    <KIcon className="w-4 h-4" style={{ color: ACCENT }} />
                                </div>
                                <div className={`flex items-center gap-1 text-xs font-semibold ${isUp ? "text-emerald-400" : "text-red-400"}`}>
                                    <TIcon className="w-3.5 h-3.5" />
                                    <span>{kpi.trend}</span>
                                </div>
                            </div>
                            <p className="text-2xl font-bold font-[Outfit]">{kpi.value}</p>
                            <p className="text-xs text-[var(--text-tertiary,#64748b)] mt-1 uppercase tracking-wider">{kpi.label}</p>
                        </div>
                    );
                })}
            </div>

            {/* Main Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* Table */}
                <div className="lg:col-span-2 rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm overflow-hidden">
                    <div className="px-5 py-4 border-b border-white/10 flex items-center justify-between">
                        <h2 className="text-base font-semibold font-[Outfit]">Registos Recentes</h2>
                        <button className="text-xs flex items-center gap-1 hover:underline" style={{ color: ACCENT }}>
                            Ver todos <ArrowUpRight className="w-3 h-3" />
                        </button>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b border-white/5">
                                    {TABLE_COLS.map((c) => (
                                        <th key={c} className="text-left px-5 py-3 text-xs uppercase tracking-wider text-[var(--text-tertiary,#64748b)] font-medium">{c}</th>
                                    ))}
                                    <th className="text-left px-5 py-3 text-xs uppercase tracking-wider text-[var(--text-tertiary,#64748b)] font-medium">Estado</th>
                                </tr>
                            </thead>
                            <tbody>
                                {TABLE_ROWS.map((row) => (
                                    <tr key={row.id} className="border-b border-white/5 hover:bg-white/[0.03] transition-colors cursor-pointer">
                                        {row.cells.map((cell, ci) => (
                                            <td key={ci} className={`px-5 py-3 ${ci === 0 ? "font-medium" : "text-[var(--text-secondary,#94a3b8)]"}`}>{cell}</td>
                                        ))}
                                        <td className="px-5 py-3">
                                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${STATUS_CLS[row.status]}`}>
                                                {STATUS_LBL[row.status]}
                                            </span>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Right Panel */}
                <div className="space-y-6">
                    {/* Chart */}
                    <div className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5">
                        <h3 className="text-sm font-semibold font-[Outfit] mb-4">Evolução Mensal</h3>
                        <div className="flex items-end gap-2 h-32">
                            {CHART.map((bar) => (
                                <div key={bar.label} className="flex-1 flex flex-col items-center gap-1">
                                    <div className="w-full rounded-t-md transition-all hover:brightness-125"
                                        style={{ height: `${(bar.value / mx) * 100}%`, background: `linear-gradient(to top, ${ACCENT}40, ${ACCENT})` }} />
                                    <span className="text-[10px] text-[var(--text-tertiary,#64748b)]">{bar.label}</span>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Activity */}
                    <div className="rounded-xl border border-white/10 bg-white/5 backdrop-blur-sm p-5">
                        <h3 className="text-sm font-semibold font-[Outfit] mb-3">Atividade Recente</h3>
                        <div className="space-y-3">
                            <div className="flex items-start gap-3 text-sm">
                                <CheckCircle className="w-4 h-4 mt-0.5 text-emerald-400 flex-shrink-0" />
                                <div><p className="text-[var(--text-secondary,#94a3b8)]">Novo registo criado</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3" /> 2m atrás</p></div>
                            </div>
                            <div className="flex items-start gap-3 text-sm">
                                <AlertTriangle className="w-4 h-4 mt-0.5 text-amber-400 flex-shrink-0" />
                                <div><p className="text-[var(--text-secondary,#94a3b8)]">Aprovação pendente</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3" /> 15m atrás</p></div>
                            </div>
                            <div className="flex items-start gap-3 text-sm">
                                <Activity className="w-4 h-4 mt-0.5 text-blue-400 flex-shrink-0" />
                                <div><p className="text-[var(--text-secondary,#94a3b8)]">Tarefa concluída</p><p className="text-[10px] text-[var(--text-tertiary,#64748b)] mt-0.5 flex items-center gap-1"><Clock className="w-3 h-3" /> 1h atrás</p></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
