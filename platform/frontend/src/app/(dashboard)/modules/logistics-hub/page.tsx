"use client";

import React from "react";

export default function LogisticsHubPage() {
    return (
        <div className="p-6 space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold">Logistics Hub</h1>
                    <p className="text-sm text-muted-foreground">Logistics Hub • iBOS Enterprise Module</p>
                </div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 rounded-lg border text-sm hover:bg-accent transition-colors">Export</button>
                    <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm hover:opacity-90 transition-opacity">+ New</button>
                </div>
            </div>

            {/* KPI Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                    <div key="0" className="rounded-xl border bg-card p-5 hover:shadow-md transition-shadow">
                        <div className="flex items-center justify-between">
                            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Deliveries/Day</span>
                            <span className="text-xs text-emerald-500">↑ 12.7%</span>
                        </div>
                        <div className="text-2xl font-bold mt-2">805</div>
                        <div className="h-1 mt-3 rounded-full bg-muted overflow-hidden">
                            <div className="h-full rounded-full" style={{ width: '72%', background: '#3b82f6' }} />
                        </div>
                    </div>
                    <div key="1" className="rounded-xl border bg-card p-5 hover:shadow-md transition-shadow">
                        <div className="flex items-center justify-between">
                            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">On-Time %</span>
                            <span className="text-xs text-emerald-500">↑ 11.3%</span>
                        </div>
                        <div className="text-2xl font-bold mt-2">185</div>
                        <div className="h-1 mt-3 rounded-full bg-muted overflow-hidden">
                            <div className="h-full rounded-full" style={{ width: '43%', background: '#10b981' }} />
                        </div>
                    </div>
                    <div key="2" className="rounded-xl border bg-card p-5 hover:shadow-md transition-shadow">
                        <div className="flex items-center justify-between">
                            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Cost/Shipment</span>
                            <span className="text-xs text-emerald-500">↑ 10.5%</span>
                        </div>
                        <div className="text-2xl font-bold mt-2">454</div>
                        <div className="h-1 mt-3 rounded-full bg-muted overflow-hidden">
                            <div className="h-full rounded-full" style={{ width: '77%', background: '#f59e0b' }} />
                        </div>
                    </div>
                    <div key="3" className="rounded-xl border bg-card p-5 hover:shadow-md transition-shadow">
                        <div className="flex items-center justify-between">
                            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Claims</span>
                            <span className="text-xs text-rose-500">↓ 0.5%</span>
                        </div>
                        <div className="text-2xl font-bold mt-2">660</div>
                        <div className="h-1 mt-3 rounded-full bg-muted overflow-hidden">
                            <div className="h-full rounded-full" style={{ width: '72%', background: '#ef4444' }} />
                        </div>
                    </div>
            </div>

                <div className="flex flex-wrap gap-2">
                    {[{"title":"Shipment Tracking","href":"/modules/logistics-hub/tracking"},{"title":"Cross-Docking","href":"/modules/logistics-hub/cross-docking"},{"title":"3PL Management","href":"/modules/logistics-hub/3pl"},{"title":"Returns & Reverse","href":"/modules/logistics-hub/returns"},{"title":"Customs & Trade","href":"/modules/logistics-hub/customs"}].map((item, i) => (
                        <a key={i} href={item.href} className="px-3 py-1.5 rounded-full text-xs font-medium border hover:bg-accent transition-colors">
                            {item.title}
                        </a>
                    ))}
                </div>
            {/* Main Content */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2 rounded-xl border bg-card p-6">
                    <h3 className="font-semibold mb-4">Tendência (Barras)</h3>
                    <div className="h-64 flex items-end gap-2 px-4">
                        {['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'].map((m, i) => {
                            const h = Math.floor(Math.random() * 60 + 20);
                            return (
                                <div key={i} className="flex-1 flex flex-col items-center gap-1">
                                    <div 
                                        className="w-full rounded-t-md transition-all hover:opacity-80" 
                                        style={{ height: h + '%', background: 'linear-gradient(to top, #3b82f6, #8b5cf6)' }}
                                    />
                                    <span className="text-[10px] text-muted-foreground">{m}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
                <div className="rounded-xl border bg-card p-6">
                    <h3 className="font-semibold mb-4">Atividade Recente</h3>
                    <div className="space-y-3">
                        {['Registro atualizado', 'Novo item criado', 'Aprovação pendente', 'Relatório gerado', 'Alerta resolvido'].map((act, i) => (
                            <div key={i} className="flex items-center gap-3 p-2 rounded-lg hover:bg-muted/50 transition-colors">
                                <div className="w-2 h-2 rounded-full" style={{ background: '#3b82f6' }} />
                                <div className="flex-1">
                                    <div className="text-sm">{act}</div>
                                    <div className="text-xs text-muted-foreground">{i + 1}h atrás</div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
