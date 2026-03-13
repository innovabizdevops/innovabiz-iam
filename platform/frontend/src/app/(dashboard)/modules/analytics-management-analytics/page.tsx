"use client";

import React from "react";

const DATA = Array.from({length:12}, (_,i) => ({
    month: ['Jan','Fev','Mar','Abr','Mai','Jun','Jul','Ago','Set','Out','Nov','Dez'][i],
    v1: Math.floor(Math.random()*80+20),
    v2: Math.floor(Math.random()*60+15),
}));

export default function AnalyticsManagementAnalyticsPage() {
    const max = Math.max(...DATA.map(d => d.v1));
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold">Analytics Management Analytics</h1>
                    <p className="text-sm text-muted-foreground">Analytics Management Analytics • Análise Visual</p>
                </div>
                <div className="flex gap-2">
                    {['7D','30D','90D','1Y','All'].map((p,i) => (
                        <button key={i} className={`px-3 py-1.5 rounded-lg text-xs ${i===3?'bg-primary text-primary-foreground':'border hover:bg-accent'}`}>{p}</button>
                    ))}
                </div>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="rounded-xl border bg-card p-6">
                    <h3 className="font-semibold mb-4">Tendência Mensal</h3>
                    <div className="h-64 flex items-end gap-1.5 px-2">
                        {DATA.map((d, i) => (
                            <div key={i} className="flex-1 flex flex-col items-center gap-1">
                                <div className="w-full rounded-t-md" style={{height: (d.v1/max*100)+'%', background:'linear-gradient(to top, #3b82f6, #8b5cf6)'}} />
                                <span className="text-[10px] text-muted-foreground">{d.month}</span>
                            </div>
                        ))}
                    </div>
                </div>
                <div className="rounded-xl border bg-card p-6">
                    <h3 className="font-semibold mb-4">Distribuição</h3>
                    <div className="h-64 flex items-center justify-center">
                        <div className="relative w-48 h-48">
                            <svg viewBox="0 0 100 100" className="w-full h-full -rotate-90">
                                {[
                                    {pct:35,color:'#3b82f6',offset:0},
                                    {pct:25,color:'#10b981',offset:35},
                                    {pct:20,color:'#f59e0b',offset:60},
                                    {pct:20,color:'#8b5cf6',offset:80},
                                ].map((s,i) => (
                                    <circle key={i} cx="50" cy="50" r="40" fill="none" stroke={s.color}
                                        strokeWidth="20" strokeDasharray={`${s.pct*2.51} ${251-s.pct*2.51}`}
                                        strokeDashoffset={`${-s.offset*2.51}`} className="transition-all" />
                                ))}
                            </svg>
                            <div className="absolute inset-0 flex flex-col items-center justify-center">
                                <span className="text-2xl font-bold">100%</span>
                                <span className="text-xs text-muted-foreground">Total</span>
                            </div>
                        </div>
                    </div>
                    <div className="flex flex-wrap gap-3 mt-4 justify-center">
                        {[{l:'Tipo A',c:'#3b82f6'},{l:'Tipo B',c:'#10b981'},{l:'Tipo C',c:'#f59e0b'},{l:'Tipo D',c:'#8b5cf6'}].map((item,i)=>(
                            <div key={i} className="flex items-center gap-1.5 text-xs">
                                <div className="w-2.5 h-2.5 rounded-full" style={{background:item.c}} />
                                <span>{item.l}</span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
            <div className="rounded-xl border bg-card p-6">
                <h3 className="font-semibold mb-4">Comparativo</h3>
                <div className="space-y-3">
                    {['Região Norte','Região Sul','Região Centro','Região Leste','Região Oeste'].map((r,i)=>{
                        const val = Math.floor(Math.random()*80+20);
                        return (
                            <div key={i} className="flex items-center gap-3">
                                <span className="text-sm w-28 shrink-0">{r}</span>
                                <div className="flex-1 h-6 rounded-full bg-muted overflow-hidden">
                                    <div className="h-full rounded-full transition-all" style={{width:val+'%', background:'linear-gradient(to right, #3b82f6, #8b5cf6)'}} />
                                </div>
                                <span className="text-sm font-mono w-12 text-right">{val}%</span>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
}
