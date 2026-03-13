"use client";
import React from "react";
export default function NuclearPowerPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">Nuclear Power</h1><p className="text-sm text-muted-foreground">Nuclear Power • iBOS</p></div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {[{l:'Total',v:'1,247',t:'+12%',up:true},{l:'Ativos',v:'834',t:'+8%',up:true},{l:'Pendentes',v:'89',t:'-3%',up:false},{l:'Score',v:'95%',t:'+2%',up:true}].map((k,i)=>(<div key={i} className="rounded-xl border bg-card p-5"><div className="text-xs text-muted-foreground">{k.l}</div><div className="flex items-end justify-between mt-2"><span className="text-2xl font-bold">{k.v}</span><span className={`text-xs ${k.up?'text-emerald-500':'text-rose-500'}`}>{k.t}</span></div></div>))}
            </div>
        </div>
    );
}
