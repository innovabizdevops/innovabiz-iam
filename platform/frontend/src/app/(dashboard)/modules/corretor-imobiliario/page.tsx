"use client";
import React from "react";
export default function CorretorImobiliarioPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">Corretor Imobiliario</h1><p className="text-sm text-muted-foreground">Corretor Imobiliario • iBOS Enterprise</p></div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 rounded-lg border text-sm hover:bg-accent">Exportar</button>
                    <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">+ Novo</button>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {[{l:'Total',v:'3,847',t:'+16.2%',up:true},{l:'Ativos',v:'2,531',t:'+11.1%',up:true},{l:'Pendentes',v:'142',t:'-6.3%',up:false},{l:'Score',v:'97.8%',t:'+2.9%',up:true}].map((k,i) => (
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
