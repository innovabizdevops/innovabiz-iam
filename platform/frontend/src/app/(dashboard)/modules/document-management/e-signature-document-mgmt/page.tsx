"use client";
import React from "react";
export default function ESignatureDocumentMgmtPage() {
    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div><h1 className="text-2xl font-bold">E Signature Document Mgmt</h1><p className="text-sm text-muted-foreground">Document Management • Detalhe</p></div>
                <div className="flex gap-2">
                    <button className="px-4 py-2 rounded-lg border text-sm">Editar</button>
                    <button className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm">Ações</button>
                </div>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2 space-y-6">
                    <div className="rounded-xl border bg-card p-6">
                        <h3 className="font-semibold mb-4">Informações</h3>
                        <div className="grid grid-cols-2 gap-y-3 gap-x-8">
                            {['Identificador','Tipo','Categoria','Status','Responsável','Data Criação','Última Atualização','Prioridade'].map((f,i) => (
                                <div key={i}><span className="text-xs text-muted-foreground">{f}</span><p className="text-sm font-medium mt-0.5">{['#REF-001','Padrão','Categoria A','Ativo','Admin','01/01/2026','19/02/2026','Alta'][i]}</p></div>
                            ))}
                        </div>
                    </div>
                    <div className="rounded-xl border bg-card p-6">
                        <h3 className="font-semibold mb-4">Timeline</h3>
                        <div className="space-y-4 border-l-2 border-muted pl-4 ml-2">
                            {['Criado por Admin','Atualizado: status → Ativo','Aprovado por Gestor','Revisão concluída','Publicado'].map((ev,i) => (
                                <div key={i} className="relative">
                                    <div className="absolute -left-[1.35rem] top-1 w-3 h-3 rounded-full bg-primary border-2 border-background" />
                                    <p className="text-sm">{ev}</p>
                                    <p className="text-xs text-muted-foreground">{i+1} dias atrás</p>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
                <div className="space-y-6">
                    <div className="rounded-xl border bg-card p-6">
                        <h3 className="font-semibold mb-3">Resumo</h3>
                        <div className="space-y-3">
                            {[{l:'Score',v:'87/100',c:'text-emerald-500'},{l:'Risco',v:'Baixo',c:'text-emerald-500'},{l:'SLA',v:'99.2%',c:'text-blue-500'},{l:'Valor',v:'€12,450',c:'text-primary'}].map((m,i) => (
                                <div key={i} className="flex justify-between items-center">
                                    <span className="text-sm text-muted-foreground">{m.l}</span>
                                    <span className={`text-sm font-bold ${m.c}`}>{m.v}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                    <div className="rounded-xl border bg-card p-6">
                        <h3 className="font-semibold mb-3">Tags</h3>
                        <div className="flex flex-wrap gap-1.5">
                            {['Ativo','Prioritário','Auditado','Conforme','v2.1'].map((tag,i) => (
                                <span key={i} className="px-2 py-0.5 rounded-full text-xs bg-primary/10 text-primary">{tag}</span>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
