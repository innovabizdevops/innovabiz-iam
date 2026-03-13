"use client";
import React from "react";
export default function CreditServicesCreditServicesPage() {
    const connectors = [
        { name: 'SAP', status: 'Conectado', color: 'emerald' },
        { name: 'Salesforce', status: 'Conectado', color: 'emerald' },
        { name: 'Slack', status: 'Pendente', color: 'amber' },
        { name: 'AWS', status: 'Conectado', color: 'emerald' },
        { name: 'Jira', status: 'Desligado', color: 'red' },
        { name: 'HubSpot', status: 'Conectado', color: 'emerald' },
    ];
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">Credit Services Credit Services</h1><p className="text-sm text-muted-foreground">Credit Services Credit Services • Integração</p></div>
                <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">+ Novo Conector</button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {connectors.map((c,i) => (
                    <div key={i} className="rounded-xl border bg-card p-5 hover:shadow-md transition-shadow">
                        <div className="flex items-center justify-between mb-3">
                            <span className="font-semibold">{c.name}</span>
                            <span className={`px-2 py-0.5 rounded-full text-xs bg-${c.color}-100 text-${c.color}-700 dark:bg-${c.color}-900/30 dark:text-${c.color}-400`}>{c.status}</span>
                        </div>
                        <div className="text-xs text-muted-foreground">Última sync: {i+1}h atrás</div>
                        <div className="flex gap-2 mt-3">
                            <button className="px-3 py-1 rounded-lg border text-xs hover:bg-accent">Config</button>
                            <button className="px-3 py-1 rounded-lg border text-xs hover:bg-accent">Logs</button>
                            <button className="px-3 py-1 rounded-lg border text-xs hover:bg-accent">Test</button>
                        </div>
                    </div>
                ))}
            </div>
            <div className="rounded-xl border bg-card p-6">
                <h3 className="font-semibold mb-4">Sync Log</h3>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                    {Array.from({length:8},(_,i) => (
                        <div key={i} className="flex items-center gap-3 text-xs font-mono p-2 rounded bg-muted/30">
                            <span className="text-muted-foreground">{new Date(Date.now()-i*3600000).toLocaleTimeString('pt-BR')}</span>
                            <span className={i%3===2?'text-amber-500':'text-emerald-500'}>{i%3===2?'WARN':'OK'}</span>
                            <span>Sync {connectors[i%connectors.length].name} - {i%3===2?'Retry':'Success'}</span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
