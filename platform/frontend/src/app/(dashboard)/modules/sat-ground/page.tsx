"use client";
import React from "react";
export default function SatGroundPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">Sat Ground</h1><p className="text-sm text-muted-foreground">Sat Ground • iBOS Enterprise</p></div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 rounded-lg border text-sm hover:bg-accent">Exportar</button>
                    <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">+ Novo</button>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {[{l:'Total',v:'1,247',t:'+12.5%',up:true},{l:'Ativos',v:'834',t:'+8.2%',up:true},{l:'Pendentes',v:'89',t:'-3.1%',up:false},{l:'Score',v:'94.7%',t:'+2.1%',up:true}].map((k,i) => (
                    <div key={i} className="rounded-xl border bg-card p-5">
                        <div className="text-xs text-muted-foreground uppercase tracking-wider">{k.l}</div>
                        <div className="flex items-end justify-between mt-2">
                            <span className="text-2xl font-bold">{k.v}</span>
                            <span className={`text-xs ${k.up?'text-emerald-500':'text-rose-500'}`}>{k.t}</span>
                        </div>
                    </div>
                ))}
            </div>
            <div className="rounded-xl border bg-card p-6"><h3 className="font-semibold mb-4">Visão Geral</h3><div className="h-48 flex items-end gap-2 px-2">{Array.from({length:12},(_,i)=>{const h=Math.floor(Math.random()*70+20);return(<div key={i} className="flex-1 flex flex-col items-center gap-1"><div className="w-full rounded-t" style={{height:h+'%',background:'linear-gradient(to top, #3b82f6, #8b5cf6)'}}/><span className="text-[10px] text-muted-foreground">{['J','F','M','A','M','J','J','A','S','O','N','D'][i]}</span></div>)})}</div></div>
        </div>
    );
}
