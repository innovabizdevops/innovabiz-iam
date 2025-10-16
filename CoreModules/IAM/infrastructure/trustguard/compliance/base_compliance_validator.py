"""
Validador de Conformidade Base

Este módulo define a classe base para validadores de conformidade específicos de mercado.
Fornece a estrutura comum e métodos para validação de conformidade regulatória.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from abc import ABC, abstractmethod
from typing import Dict, List, Any, Optional, Set, Tuple
import datetime
import json

from ....observability.core.multi_layer_monitor import MultiLayerMonitor


class BaseComplianceValidator(ABC):
    """
    Classe base para validadores de conformidade específicos de mercado.
    
    Define a interface e comportamentos comuns para todos os validadores de conformidade.
    """
    
    def __init__(self, observability_monitor: Optional[MultiLayerMonitor] = None):
        """
        Inicializa o validador de conformidade base.
        
        Args:
            observability_monitor: Monitor de observabilidade multi-camada
        """
        self.observability = observability_monitor
        
        # Configurar logger e métricas
        if observability_monitor:
            self.logger = observability_monitor.get_logger()
            self.metrics = observability_monitor.get_metrics_collector()
        else:
            from ....observability.logging.hook_logger import Logger
            from ....observability.metrics.hook_metrics import MetricsCollector
            
            self.logger = Logger()
            self.metrics = MetricsCollector()
        
        # Inicializar resultado de validação padrão
        self._resultado_padrao = {
            "valido": True,
            "violacoes": [],
            "avisos": [],
            "score_conformidade": 100,
            "detalhes": {}
        }
        
        self.logger.info(f"Validador de conformidade {self.__class__.__name__} inicializado")
    
    @abstractmethod
    def validar_conformidade(self, contexto: Dict[str, Any]) -> Dict[str, Any]:
        """
        Valida a conformidade com base no contexto fornecido.
        
        Args:
            contexto: Contexto para validação de conformidade
            
        Returns:
            Dict[str, Any]: Resultado da validação de conformidade
        """
        pass
    
    @abstractmethod
    def obter_requisitos_conformidade(self) -> Dict[str, Any]:
        """
        Obtém os requisitos de conformidade específicos deste validador.
        
        Returns:
            Dict[str, Any]: Requisitos de conformidade
        """
        pass
    
    def registrar_violacao(self, resultado: Dict[str, Any], codigo: str, descricao: str, 
                           severidade: str = "alta", impacto: str = "bloqueante",
                           referencias: List[str] = None, dados_contexto: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Registra uma violação de conformidade no resultado.
        
        Args:
            resultado: Resultado da validação onde a violação será registrada
            codigo: Código da violação
            descricao: Descrição da violação
            severidade: Severidade da violação (baixa, media, alta, critica)
            impacto: Impacto da violação (informativo, aviso, restritivo, bloqueante)
            referencias: Referências normativas/regulatórias
            dados_contexto: Dados de contexto relacionados à violação
            
        Returns:
            Dict[str, Any]: Resultado atualizado com a violação registrada
        """
        # Garantir que as listas de violações e avisos existam
        if "violacoes" not in resultado:
            resultado["violacoes"] = []
        
        if "avisos" not in resultado:
            resultado["avisos"] = []
        
        # Criar objeto de violação
        violacao = {
            "codigo": codigo,
            "descricao": descricao,
            "severidade": severidade,
            "impacto": impacto,
            "timestamp": datetime.datetime.now().isoformat(),
            "referencias": referencias or []
        }
        
        # Adicionar dados de contexto se fornecidos
        if dados_contexto:
            violacao["contexto"] = dados_contexto
        
        # Adicionar à lista apropriada
        if impacto in ["bloqueante", "restritivo"]:
            resultado["violacoes"].append(violacao)
            # Marcar como inválido se for bloqueante
            if impacto == "bloqueante":
                resultado["valido"] = False
        else:
            resultado["avisos"].append(violacao)
        
        # Ajustar score de conformidade
        if severidade == "baixa":
            reducao = 5
        elif severidade == "media":
            reducao = 10
        elif severidade == "alta":
            reducao = 20
        else:  # critica
            reducao = 30
        
        if "score_conformidade" not in resultado:
            resultado["score_conformidade"] = 100
        
        resultado["score_conformidade"] = max(0, resultado["score_conformidade"] - reducao)
        
        # Registrar métrica
        self.metrics.incrementCounter(
            "trustguard.compliance.violation", 
            {"validator": self.__class__.__name__, "severity": severidade, "impact": impacto}
        )
        
        return resultado
    
    def validar_campo_requerido(self, resultado: Dict[str, Any], dados: Dict[str, Any], 
                               campo: str, mensagem: str = None, 
                               severidade: str = "alta", impacto: str = "bloqueante") -> Dict[str, Any]:
        """
        Valida se um campo requerido está presente e não é vazio.
        
        Args:
            resultado: Resultado da validação
            dados: Dados a serem validados
            campo: Nome do campo a ser validado
            mensagem: Mensagem personalizada (opcional)
            severidade: Severidade da violação se o campo não for válido
            impacto: Impacto da violação se o campo não for válido
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        if campo not in dados or dados[campo] is None or (isinstance(dados[campo], str) and not dados[campo].strip()):
            msg = mensagem or f"Campo obrigatório '{campo}' não fornecido ou vazio"
            return self.registrar_violacao(
                resultado, 
                f"CAMPO_REQUERIDO_{campo.upper()}", 
                msg, 
                severidade, 
                impacto
            )
        
        return resultado
    
    def validar_formato_campo(self, resultado: Dict[str, Any], dados: Dict[str, Any], 
                             campo: str, validacao_fn, mensagem: str = None,
                             severidade: str = "media", impacto: str = "aviso") -> Dict[str, Any]:
        """
        Valida se um campo tem o formato correto usando uma função de validação.
        
        Args:
            resultado: Resultado da validação
            dados: Dados a serem validados
            campo: Nome do campo a ser validado
            validacao_fn: Função que recebe o valor do campo e retorna bool
            mensagem: Mensagem personalizada (opcional)
            severidade: Severidade da violação se o campo não for válido
            impacto: Impacto da violação se o campo não for válido
            
        Returns:
            Dict[str, Any]: Resultado atualizado
        """
        if campo in dados and dados[campo] is not None:
            try:
                if not validacao_fn(dados[campo]):
                    msg = mensagem or f"Campo '{campo}' tem formato inválido"
                    return self.registrar_violacao(
                        resultado, 
                        f"FORMATO_INVALIDO_{campo.upper()}", 
                        msg, 
                        severidade, 
                        impacto,
                        dados_contexto={"valor": str(dados[campo])}
                    )
            except Exception as e:
                msg = f"Erro ao validar formato do campo '{campo}': {str(e)}"
                return self.registrar_violacao(
                    resultado, 
                    f"ERRO_VALIDACAO_{campo.upper()}", 
                    msg, 
                    severidade, 
                    impacto
                )
        
        return resultado
