"use client";
import React from "react";
import { BarChart3, TrendingUp, Clock, CheckCircle2, Search, Filter, Download, Plus, MoreHorizontal, ArrowUpRight, ArrowDownRight } from "lucide-react";

const PAGE = "DevOps Platform";
const KPIS = [
  { l: "Total DevOps Platform", v: "2,847", d: "+12.3%", u: true },
  { l: "DevOps Platform Ativos", v: "1,923", d: "+5.1%", u: true },
  { l: "Pendentes", v: "124", d: "-8.4%", u: false },
  { l: "Score (%)", v: "94.2%", d: "+2.1%", u: true },
];
const COLS = ["ID", "Nome", "Categoria", "Responsável", "Data", "Estado"];
const ROWS = [
  { id: "001", c: ["DevOps Platform-001", "DevOps Platform Alpha", "Tipo A", "Ana Costa", "2026-02-20", "Ativo"] },
  { id: "002", c: ["DevOps Platform-002", "DevOps Platform Beta", "Tipo B", "João Silva", "2026-02-19", "Em Revisão"] },
  { id: "003", c: ["DevOps Platform-003", "DevOps Platform Gamma", "Tipo A", "Maria Santos", "2026-02-18", "Concluído"] },
  { id: "004", c: ["DevOps Platform-004", "DevOps Platform Delta", "Tipo C", "Pedro Oliveira", "2026-02-17", "Pendente"] },
  { id: "005", c: ["DevOps Platform-005", "DevOps Platform Epsilon", "Tipo B", "Carla Mendes", "2026-02-16", "Ativo"] },
];
const CHART = [65, 78, 52, 90, 85, 72, 91, 68, 83, 76, 89, 94];
const ACTIVITY = [
  { t: "Novo registo criado em DevOps Platform", ts: "Há 2min", a: "Sistema" },
  { t: "Aprovação pendente para item #247", ts: "Há 15min", a: "Ana Costa" },
  { t: "Relatório mensal gerado", ts: "Há 1h", a: "Auto" },
  { t: "Atualização de estado em lote", ts: "Há 3h", a: "João Silva" },
];

export default function DevopsPlatformPage() {
  return (
    <div className="p-6 space-y-6 max-w-[1600px] mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{PAGE}</h1>
          <p className="text-sm text-muted-foreground">DevOps Platform — iBOS Enterprise</p>
        </div>
        <div className="flex gap-2">
          <button className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-sm hover:bg-accent transition-colors"><Download className="h-3.5 w-3.5" /> Exportar</button>
          <button className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary text-primary-foreground text-sm hover:bg-primary/90 transition-colors"><Plus className="h-3.5 w-3.5" /> Novo</button>
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {KPIS.map((k, i) => (
          <div key={i} className="rounded-xl border bg-card p-4 hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium text-muted-foreground">{k.l}</p>
              {i === 0 ? <BarChart3 className="h-4 w-4 text-muted-foreground" /> : i === 1 ? <TrendingUp className="h-4 w-4 text-muted-foreground" /> : i === 2 ? <Clock className="h-4 w-4 text-muted-foreground" /> : <CheckCircle2 className="h-4 w-4 text-muted-foreground" />}
            </div>
            <p className="text-2xl font-bold mt-2">{k.v}</p>
            <div className="flex items-center gap-1 mt-1">
              {k.u ? <ArrowUpRight className="h-3 w-3 text-emerald-500" /> : <ArrowDownRight className="h-3 w-3 text-red-500" />}
              <span className={`text-xs font-medium ${k.u ? "text-emerald-500" : "text-red-500"}`}>{k.d}</span>
              <span className="text-xs text-muted-foreground ml-1">vs mês anterior</span>
            </div>
          </div>
        ))}
      </div>

      {/* Chart + Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 rounded-xl border bg-card p-4">
          <h3 className="text-sm font-semibold mb-4">Evolução Mensal</h3>
          <div className="flex items-end gap-1.5 h-32">
            {CHART.map((v, i) => (<div key={i} className="flex-1 rounded-t transition-all hover:opacity-80" style={{ height: `${v}%`, background: `hsl(${200 + i * 10}, 70%, 50%)` }} title={`${v}%`} />))}
          </div>
          <div className="flex justify-between mt-2 text-[10px] text-muted-foreground">
            {["Jan","Fev","Mar","Abr","Mai","Jun","Jul","Ago","Set","Out","Nov","Dez"].map(m => (<span key={m}>{m}</span>))}
          </div>
        </div>
        <div className="rounded-xl border bg-card p-4">
          <h3 className="text-sm font-semibold mb-3">Atividade Recente</h3>
          <div className="space-y-3">
            {ACTIVITY.map((a, i) => (
              <div key={i} className="flex gap-3 items-start">
                <div className="h-2 w-2 rounded-full bg-primary mt-1.5 shrink-0" />
                <div className="min-w-0">
                  <p className="text-xs leading-snug">{a.t}</p>
                  <p className="text-[10px] text-muted-foreground mt-0.5">{a.a} · {a.ts}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border bg-card">
        <div className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-b">
          <h3 className="text-sm font-semibold">Registos de {PAGE}</h3>
          <div className="flex gap-2">
            <div className="relative"><Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" /><input placeholder="Pesquisar..." className="pl-8 pr-3 py-1.5 rounded-lg border text-sm w-48 bg-transparent" /></div>
            <button className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg border text-sm hover:bg-accent"><Filter className="h-3.5 w-3.5" /> Filtrar</button>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead><tr className="border-b bg-muted/30">{COLS.map(c => (<th key={c} className="text-left text-xs font-medium text-muted-foreground px-4 py-2.5">{c}</th>))}<th className="w-10" /></tr></thead>
            <tbody>
              {ROWS.map(r => (
                <tr key={r.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                  {r.c.map((cell, j) => (<td key={j} className="px-4 py-2.5 text-sm">{j === r.c.length - 1 ? (<span className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium ${cell === "Ativo" ? "bg-emerald-500/10 text-emerald-500" : cell === "Concluído" ? "bg-blue-500/10 text-blue-500" : cell === "Pendente" ? "bg-amber-500/10 text-amber-500" : "bg-purple-500/10 text-purple-500"}`}>{cell}</span>) : cell}</td>))}
                  <td className="px-2"><button className="p-1 hover:bg-accent rounded"><MoreHorizontal className="h-4 w-4 text-muted-foreground" /></button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="p-3 border-t flex items-center justify-between text-xs text-muted-foreground">
          <span>A mostrar 1-5 de {ROWS.length} registos</span>
          <div className="flex gap-1">
            <button className="px-2 py-1 rounded border hover:bg-accent">Anterior</button>
            <button className="px-2 py-1 rounded border bg-primary text-primary-foreground">1</button>
            <button className="px-2 py-1 rounded border hover:bg-accent">Seguinte</button>
          </div>
        </div>
      </div>
    </div>
  );
}
