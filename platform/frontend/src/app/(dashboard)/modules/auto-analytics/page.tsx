"use client";
import React from "react";
export default function AutoAnalyticsPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">Auto Analytics</h1><p className="text-sm text-muted-foreground">Auto Analytics • iBOS Enterprise</p></div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 rounded-lg border text-sm hover:bg-accent">Exportar</button>
                    <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">+ Novo</button>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {[{l:'Total',v:'2,841',t:'+14.3%',up:true},{l:'Ativos',v:'1,672',t:'+9.1%',up:true},{l:'Pendentes',v:'156',t:'-4.2%',up:false},{l:'Score',v:'96.2%',t:'+1.8%',up:true}].map((k,i) => (
                    <div key={i} className="rounded-xl border bg-card p-5">
                        <div className="text-xs text-muted-foreground uppercase tracking-wider">{k.l}</div>
                        <div className="flex items-end justify-between mt-2">
                            <span className="text-2xl font-bold">{k.v}</span>
                            <span className={`text-xs ${k.up?'text-emerald-500':'text-rose-500'}`}>{k.t}</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
