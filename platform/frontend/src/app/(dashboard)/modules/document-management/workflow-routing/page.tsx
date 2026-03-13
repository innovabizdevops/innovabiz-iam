"use client";

import React from "react";

const STEPS = [
    { name: 'Início', desc: 'Solicitação recebida', status: 'done' },
    { name: 'Análise', desc: 'Verificação de dados', status: 'done' },
    { name: 'Aprovação', desc: 'Revisão por gestor', status: 'current' },
    { name: 'Execução', desc: 'Processamento', status: 'pending' },
    { name: 'Conclusão', desc: 'Finalização e notificação', status: 'pending' },
];

export default function WorkflowRoutingPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold">Workflow Routing</h1>
                    <p className="text-sm text-muted-foreground">Document Management • Workflow / Fluxo</p>
                </div>
                <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">Iniciar Novo Fluxo</button>
            </div>
            <div className="rounded-xl border bg-card p-8">
                <div className="flex items-center justify-between">
                    {STEPS.map((step, i) => (
                        <React.Fragment key={i}>
                            <div className="flex flex-col items-center gap-2 flex-1">
                                <div className={`w-10 h-10 rounded-full flex items-center justify-center text-sm font-bold ${
                                    step.status==='done'?'bg-emerald-500 text-white':
                                    step.status==='current'?'bg-primary text-primary-foreground ring-4 ring-primary/20':
                                    'bg-muted text-muted-foreground'
                                }`}>{step.status==='done'?'✓':i+1}</div>
                                <span className="text-sm font-medium text-center">{step.name}</span>
                                <span className="text-xs text-muted-foreground text-center">{step.desc}</span>
                            </div>
                            {i < STEPS.length - 1 && (
                                <div className={`flex-1 h-0.5 mx-2 ${step.status==='done'?'bg-emerald-500':'bg-muted'}`} />
                            )}
                        </React.Fragment>
                    ))}
                </div>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="rounded-xl border bg-card p-6">
                    <h3 className="font-semibold mb-4">Execuções Recentes</h3>
                    <div className="space-y-3">
                        {['#FL-001 Completo','#FL-002 Em progresso','#FL-003 Aguardando','#FL-004 Completo','#FL-005 Rejeitado'].map((f,i)=>(
                            <div key={i} className="flex items-center justify-between p-3 rounded-lg border hover:bg-muted/30 cursor-pointer">
                                <span className="text-sm font-medium">{f.split(' ')[0]}</span>
                                <span className={`px-2 py-0.5 rounded-full text-xs ${
                                    f.includes('Completo')?'bg-emerald-100 text-emerald-700':
                                    f.includes('progresso')?'bg-blue-100 text-blue-700':
                                    f.includes('Rejeitado')?'bg-red-100 text-red-700':
                                    'bg-amber-100 text-amber-700'
                                }`}>{f.split(' ').slice(1).join(' ')}</span>
                            </div>
                        ))}
                    </div>
                </div>
                <div className="rounded-xl border bg-card p-6">
                    <h3 className="font-semibold mb-4">Métricas do Fluxo</h3>
                    <div className="grid grid-cols-2 gap-4">
                        {[{l:'Tempo Médio',v:'2.4 dias'},{l:'Taxa Aprovação',v:'87%'},{l:'Em Andamento',v:'14'},{l:'Concluídos/Mês',v:'156'}].map((m,i)=>(
                            <div key={i} className="p-4 rounded-lg bg-muted/30">
                                <div className="text-xs text-muted-foreground">{m.l}</div>
                                <div className="text-xl font-bold mt-1">{m.v}</div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
