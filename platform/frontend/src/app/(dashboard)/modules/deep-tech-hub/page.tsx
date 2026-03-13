"use client";
import React from "react";
export default function DeepTechHubPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">Deep Tech Hub</h1><p className="text-sm text-muted-foreground">Deep Tech Hub — iBOS Enterprise</p></div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 rounded-lg border text-sm hover:bg-accent">Exportar</button>
                    <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">+ Novo</button>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {[{l:"Total",v:"2,847",t:"+12.3%"},{l:"Ativos",v:"1,923",t:"+5.1%"},{l:"Pendentes",v:"124",t:"-8.4%"},{l:"Score",v:"94.2%",t:"+2.1%"}].map((k,i)=>(
                    <div key={i} className="rounded-xl border bg-card p-4"><p className="text-xs text-muted-foreground">{k.l}</p><p className="text-2xl font-bold mt-1">{k.v}</p><span className="text-xs text-emerald-500">{k.t}</span></div>
                ))}
            </div>
        </div>
    );
}
