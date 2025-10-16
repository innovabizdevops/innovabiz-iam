#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Matriz de Autorização e Aprovação para Alertas Comportamentais

Este módulo implementa um sistema de matriz de autorização para aprovação ou 
rejeição de alertas gerados pelo sistema de análise comportamental, com base
em níveis hierárquicos, papéis, tipos de alerta e limites de valores.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import uuid
import json
import logging
import datetime
from enum import Enum
from typing import Dict, Any, List, Optional, Set, Tuple, Union

# Configuração do logger
logger = logging.getLogger("iam.trustguard.authorization.approval_matrix")


class ApprovalAction(Enum):
    """Ações possíveis para um alerta comportamental."""
    APPROVE = "approve"  # Aprovar transação/operação
    REJECT = "reject"    # Rejeitar transação/operação
    ESCALATE = "escalate"  # Encaminhar para nível superior
    INVESTIGATE = "investigate"  # Enviar para investigação manual
    CHALLENGE = "challenge"  # Solicitar verificação adicional
    BLOCK = "block"  # Bloquear temporariamente a conta/dispositivo
    RESTRICT = "restrict"  # Restringir algumas funcionalidades
    MONITOR = "monitor"  # Aprovar mas com monitoramento intensificado


class AlertSeverity(Enum):
    """Níveis de severidade para alertas comportamentais."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class AlertCategory(Enum):
    """Categorias de alertas comportamentais."""
    AUTHENTICATION = "authentication"
    TRANSACTION = "transaction"
    SESSION = "session"
    DEVICE = "device"
    LOCATION = "location"
    PROFILE = "profile"
    COMBINED = "combined"


class ApprovalLevel(Enum):
    """Níveis de autoridade para aprovação."""
    L1_AGENT = "l1_agent"  # Analista L1
    L2_SPECIALIST = "l2_specialist"  # Especialista L2
    L3_SUPERVISOR = "l3_supervisor"  # Supervisor L3
    L4_MANAGER = "l4_manager"  # Gerente L4
    L5_DIRECTOR = "l5_director"  # Diretor L5
    AUTO_SYSTEM = "auto_system"  # Sistema automatizado


class ApprovalCriteria:
    """
    Critérios para autorização e aprovação de alertas comportamentais.
    
    Define as regras e limites para diferentes níveis de aprovação
    baseados em severidade, valor da transação, categoria e região.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa os critérios de aprovação.
        
        Args:
            config: Configurações para os critérios de aprovação
        """
        self.config = config
        self.threshold_matrix = config.get("threshold_matrix", {})
        self.role_permissions = config.get("role_permissions", {})
        self.region_overrides = config.get("region_overrides", {})
        self.customer_segment_rules = config.get("customer_segment_rules", {})
        self.auto_approval_rules = config.get("auto_approval_rules", {})
        self.escalation_rules = config.get("escalation_rules", {})
        self.default_max_amount = config.get("default_max_amount", 1000.0)
        
        logger.info("Critérios de aprovação inicializados")
    
    def get_required_approval_level(self, 
                                  alert_data: Dict[str, Any]) -> ApprovalLevel:
        """
        Determina o nível de aprovação necessário para um alerta.
        
        Args:
            alert_data: Dados do alerta comportamental
            
        Returns:
            Nível de aprovação requerido
        """
        try:
            # Extrair dados relevantes do alerta
            severity = AlertSeverity(alert_data.get("severity", "medium"))
            category = AlertCategory(alert_data.get("category", "authentication"))
            transaction_amount = float(alert_data.get("transaction_amount", 0.0))
            risk_score = float(alert_data.get("risk_score", 0.5))
            region = alert_data.get("region", "global")
            customer_segment = alert_data.get("customer_segment", "default")
            
            # Verificar se há override específico para a região
            if region in self.region_overrides:
                regional_matrix = self.region_overrides[region]
                if severity.value in regional_matrix and category.value in regional_matrix[severity.value]:
                    return ApprovalLevel(regional_matrix[severity.value][category.value])
            
            # Verificar regras específicas por segmento de cliente
            if customer_segment in self.customer_segment_rules:
                segment_rules = self.customer_segment_rules[customer_segment]
                if severity.value in segment_rules:
                    # Encontrar o limite de valor aplicável
                    for limit, level in sorted(segment_rules[severity.value].items(), key=lambda x: float(x[0])):
                        if transaction_amount <= float(limit):
                            return ApprovalLevel(level)
            
            # Aplicar matriz padrão de thresholds
            if severity.value in self.threshold_matrix:
                severity_matrix = self.threshold_matrix[severity.value]
                
                if category.value in severity_matrix:
                    category_matrix = severity_matrix[category.value]
                    
                    # Encontrar o nível de aprovação baseado no valor da transação
                    for amount_str, level in sorted(category_matrix.items(), key=lambda x: float(x[0])):
                        if transaction_amount <= float(amount_str):
                            return ApprovalLevel(level)
            
            # Regra padrão baseada apenas na severidade, se não encontrou regra específica
            severity_defaults = {
                AlertSeverity.LOW: ApprovalLevel.L1_AGENT,
                AlertSeverity.MEDIUM: ApprovalLevel.L2_SPECIALIST,
                AlertSeverity.HIGH: ApprovalLevel.L3_SUPERVISOR,
                AlertSeverity.CRITICAL: ApprovalLevel.L4_MANAGER
            }
            
            return severity_defaults.get(severity, ApprovalLevel.L3_SUPERVISOR)
            
        except Exception as e:
            logger.error(f"Erro ao determinar nível de aprovação: {str(e)}")
            # Em caso de erro, usar um nível conservador
            return ApprovalLevel.L3_SUPERVISOR
    
    def can_auto_approve(self, alert_data: Dict[str, Any]) -> bool:
        """
        Verifica se um alerta pode ser aprovado automaticamente.
        
        Args:
            alert_data: Dados do alerta comportamental
            
        Returns:
            True se puder ser aprovado automaticamente
        """
        try:
            # Extrair dados relevantes
            severity = AlertSeverity(alert_data.get("severity", "medium"))
            category = AlertCategory(alert_data.get("category", "authentication"))
            transaction_amount = float(alert_data.get("transaction_amount", 0.0))
            risk_score = float(alert_data.get("risk_score", 0.5))
            region = alert_data.get("region", "global")
            customer_segment = alert_data.get("customer_segment", "default")
            
            # Verificar regras de auto-aprovação
            auto_rules = self.auto_approval_rules
            
            # Regra 1: Verificar limites por severidade
            if severity.value in auto_rules.get("severity_limits", {}):
                max_amount = auto_rules["severity_limits"][severity.value]
                if transaction_amount > max_amount:
                    return False
            
            # Regra 2: Verificar limites por categoria
            if category.value in auto_rules.get("category_limits", {}):
                max_amount = auto_rules["category_limits"][category.value]
                if transaction_amount > max_amount:
                    return False
            
            # Regra 3: Verificar score de risco máximo para auto-aprovação
            if risk_score > auto_rules.get("max_risk_score", 0.3):
                return False
            
            # Regra 4: Verificar limites específicos por região
            if region in auto_rules.get("region_limits", {}):
                max_amount = auto_rules["region_limits"][region]
                if transaction_amount > max_amount:
                    return False
            
            # Regra 5: Verificar limites específicos por segmento de cliente
            if customer_segment in auto_rules.get("segment_limits", {}):
                max_amount = auto_rules["segment_limits"][customer_segment]
                if transaction_amount > max_amount:
                    return False
            
            # Se passou por todas as verificações, pode auto-aprovar
            return True
            
        except Exception as e:
            logger.error(f"Erro ao verificar auto-aprovação: {str(e)}")
            # Em caso de erro, não permitir auto-aprovação
            return False
    
    def get_allowed_actions(self, 
                          approval_level: ApprovalLevel, 
                          alert_data: Dict[str, Any]) -> Set[ApprovalAction]:
        """
        Obtém as ações permitidas para um nível de aprovação e alerta.
        
        Args:
            approval_level: Nível de aprovação do usuário
            alert_data: Dados do alerta comportamental
            
        Returns:
            Conjunto de ações permitidas
        """
        try:
            severity = AlertSeverity(alert_data.get("severity", "medium"))
            
            # Obter permissões baseadas no papel
            if approval_level.value in self.role_permissions:
                role_perms = self.role_permissions[approval_level.value]
                
                # Verificar permissões para esta severidade
                if severity.value in role_perms:
                    return {ApprovalAction(action) for action in role_perms[severity.value]}
                
                # Se não houver permissão específica para esta severidade, usar padrão do papel
                if "default" in role_perms:
                    return {ApprovalAction(action) for action in role_perms["default"]}
            
            # Permissões padrão por nível
            default_permissions = {
                ApprovalLevel.L1_AGENT: {
                    AlertSeverity.LOW: {ApprovalAction.APPROVE, ApprovalAction.ESCALATE, ApprovalAction.MONITOR},
                    AlertSeverity.MEDIUM: {ApprovalAction.ESCALATE, ApprovalAction.MONITOR},
                    AlertSeverity.HIGH: {ApprovalAction.ESCALATE},
                    AlertSeverity.CRITICAL: {ApprovalAction.ESCALATE}
                },
                ApprovalLevel.L2_SPECIALIST: {
                    AlertSeverity.LOW: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.MONITOR},
                    AlertSeverity.MEDIUM: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.ESCALATE, ApprovalAction.MONITOR},
                    AlertSeverity.HIGH: {ApprovalAction.ESCALATE, ApprovalAction.INVESTIGATE},
                    AlertSeverity.CRITICAL: {ApprovalAction.ESCALATE}
                },
                ApprovalLevel.L3_SUPERVISOR: {
                    AlertSeverity.LOW: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.MONITOR},
                    AlertSeverity.MEDIUM: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.MONITOR},
                    AlertSeverity.HIGH: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.ESCALATE, ApprovalAction.INVESTIGATE},
                    AlertSeverity.CRITICAL: {ApprovalAction.ESCALATE, ApprovalAction.INVESTIGATE, ApprovalAction.BLOCK}
                },
                ApprovalLevel.L4_MANAGER: {
                    AlertSeverity.LOW: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.MONITOR},
                    AlertSeverity.MEDIUM: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.MONITOR},
                    AlertSeverity.HIGH: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.INVESTIGATE, ApprovalAction.BLOCK, ApprovalAction.RESTRICT},
                    AlertSeverity.CRITICAL: {ApprovalAction.APPROVE, ApprovalAction.REJECT, ApprovalAction.INVESTIGATE, ApprovalAction.BLOCK, ApprovalAction.RESTRICT}
                },
                ApprovalLevel.L5_DIRECTOR: {
                    AlertSeverity.LOW: {action for action in ApprovalAction},
                    AlertSeverity.MEDIUM: {action for action in ApprovalAction},
                    AlertSeverity.HIGH: {action for action in ApprovalAction},
                    AlertSeverity.CRITICAL: {action for action in ApprovalAction}
                },
                ApprovalLevel.AUTO_SYSTEM: {
                    AlertSeverity.LOW: {ApprovalAction.APPROVE, ApprovalAction.MONITOR},
                    AlertSeverity.MEDIUM: {ApprovalAction.MONITOR, ApprovalAction.CHALLENGE},
                    AlertSeverity.HIGH: set(),
                    AlertSeverity.CRITICAL: set()
                }
            }
            
            if approval_level in default_permissions and severity in default_permissions[approval_level]:
                return default_permissions[approval_level][severity]
            
            # Se não encontrou permissões específicas
            return {ApprovalAction.ESCALATE}
            
        except Exception as e:
            logger.error(f"Erro ao obter ações permitidas: {str(e)}")
            # Em caso de erro, permitir apenas escalação
            return {ApprovalAction.ESCALATE}


class ApprovalMatrix:
    """
    Matriz de Autorização e Aprovação para Alertas Comportamentais.
    
    Gerencia o fluxo de aprovação de alertas comportamentais, determinando
    os níveis de aprovação necessários e rastreando aprovações.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa a matriz de aprovação.
        
        Args:
            config: Configurações para a matriz de aprovação
        """
        self.config = config
        self.approval_criteria = ApprovalCriteria(config.get("approval_criteria", {}))
        self.max_escalation_levels = config.get("max_escalation_levels", 3)
        self.auto_approval_enabled = config.get("auto_approval_enabled", True)
        self.require_comments_for_reject = config.get("require_comments_for_reject", True)
        self.store_audit_trail = config.get("store_audit_trail", True)
        self.pending_approvals = {}  # Armazenamento em memória (deve ser persistido em produção)
        
        logger.info("Matriz de aprovação inicializada")
    
    def create_approval_request(self, alert_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Cria uma nova solicitação de aprovação para um alerta.
        
        Args:
            alert_data: Dados do alerta comportamental
            
        Returns:
            Solicitação de aprovação criada
        """
        try:
            alert_id = alert_data.get("alert_id", str(uuid.uuid4()))
            required_level = self.approval_criteria.get_required_approval_level(alert_data)
            
            # Verificar se pode ser auto-aprovado
            can_auto_approve = False
            if self.auto_approval_enabled:
                can_auto_approve = self.approval_criteria.can_auto_approve(alert_data)
            
            # Criar solicitação
            approval_request = {
                "request_id": str(uuid.uuid4()),
                "alert_id": alert_id,
                "created_at": datetime.datetime.now().isoformat(),
                "updated_at": datetime.datetime.now().isoformat(),
                "status": "auto_approved" if can_auto_approve else "pending",
                "required_approval_level": required_level.value,
                "current_approval_level": required_level.value,
                "escalation_count": 0,
                "can_auto_approve": can_auto_approve,
                "alert_data": alert_data,
                "approvals": [],
                "comments": []
            }
            
            # Auto-aprovar se possível
            if can_auto_approve:
                approval_request["approvals"].append({
                    "user_id": "system",
                    "approval_level": ApprovalLevel.AUTO_SYSTEM.value,
                    "action": ApprovalAction.APPROVE.value,
                    "timestamp": datetime.datetime.now().isoformat(),
                    "comments": "Aprovação automática conforme regras de negócio"
                })
                
                logger.info(f"Alerta {alert_id} aprovado automaticamente")
            else:
                # Armazenar solicitação pendente
                self.pending_approvals[approval_request["request_id"]] = approval_request
                
                logger.info(
                    f"Solicitação de aprovação criada para alerta {alert_id}. "
                    f"Nível requerido: {required_level.value}"
                )
            
            return approval_request
            
        except Exception as e:
            logger.error(f"Erro ao criar solicitação de aprovação: {str(e)}")
            # Retornar solicitação de erro
            return {
                "request_id": str(uuid.uuid4()),
                "alert_id": alert_data.get("alert_id", "unknown"),
                "status": "error",
                "error": str(e)
            }
    
    def process_approval_action(self, 
                             request_id: str, 
                             user_data: Dict[str, Any],
                             action: ApprovalAction,
                             comments: str = "") -> Dict[str, Any]:
        """
        Processa uma ação de aprovação/rejeição de um alerta.
        
        Args:
            request_id: ID da solicitação de aprovação
            user_data: Dados do usuário que está tomando a ação
            action: Ação de aprovação
            comments: Comentários opcionais
            
        Returns:
            Resultado do processamento
        """
        try:
            # Verificar se a solicitação existe
            if request_id not in self.pending_approvals:
                logger.warning(f"Solicitação de aprovação não encontrada: {request_id}")
                return {
                    "success": False,
                    "error": "Solicitação de aprovação não encontrada",
                    "request_id": request_id
                }
            
            # Obter solicitação
            request = self.pending_approvals[request_id]
            
            # Verificar se a solicitação já foi finalizada
            if request["status"] in ["approved", "rejected"]:
                logger.warning(f"Solicitação de aprovação já finalizada: {request_id}")
                return {
                    "success": False,
                    "error": f"Solicitação já {request['status']}",
                    "request_id": request_id,
                    "status": request["status"]
                }
            
            # Obter nível de aprovação do usuário
            user_approval_level = ApprovalLevel(user_data.get("approval_level", ApprovalLevel.L1_AGENT.value))
            
            # Verificar se o usuário tem nível suficiente
            required_level = ApprovalLevel(request["required_approval_level"])
            if self._get_level_rank(user_approval_level) < self._get_level_rank(required_level):
                logger.warning(
                    f"Usuário {user_data.get('user_id')} com nível {user_approval_level.value} "
                    f"não tem autorização para {action.value} solicitação que requer {required_level.value}"
                )
                return {
                    "success": False,
                    "error": f"Nível de aprovação insuficiente. Requerido: {required_level.value}",
                    "request_id": request_id
                }
            
            # Verificar se o usuário tem permissão para realizar esta ação
            allowed_actions = self.approval_criteria.get_allowed_actions(
                user_approval_level, request["alert_data"]
            )
            
            if action not in allowed_actions:
                logger.warning(
                    f"Usuário {user_data.get('user_id')} com nível {user_approval_level.value} "
                    f"não tem permissão para ação {action.value}"
                )
                return {
                    "success": False,
                    "error": f"Ação {action.value} não permitida para seu nível de autorização",
                    "request_id": request_id,
                    "allowed_actions": [a.value for a in allowed_actions]
                }
            
            # Se rejeição e comentários são obrigatórios, verificar
            if action == ApprovalAction.REJECT and self.require_comments_for_reject and not comments:
                return {
                    "success": False,
                    "error": "Comentários são obrigatórios para rejeição",
                    "request_id": request_id
                }
            
            # Registrar aprovação/ação
            approval_entry = {
                "user_id": user_data.get("user_id", "unknown"),
                "user_name": user_data.get("name", "Unknown User"),
                "approval_level": user_approval_level.value,
                "action": action.value,
                "timestamp": datetime.datetime.now().isoformat(),
                "comments": comments
            }
            
            request["approvals"].append(approval_entry)
            
            # Adicionar comentários se fornecidos
            if comments:
                request["comments"].append({
                    "user_id": user_data.get("user_id", "unknown"),
                    "user_name": user_data.get("name", "Unknown User"),
                    "timestamp": datetime.datetime.now().isoformat(),
                    "comment": comments
                })
            
            # Atualizar status baseado na ação
            if action == ApprovalAction.APPROVE:
                request["status"] = "approved"
                request["updated_at"] = datetime.datetime.now().isoformat()
                logger.info(f"Solicitação {request_id} aprovada por {user_data.get('user_id')}")
                
            elif action == ApprovalAction.REJECT:
                request["status"] = "rejected"
                request["updated_at"] = datetime.datetime.now().isoformat()
                logger.info(f"Solicitação {request_id} rejeitada por {user_data.get('user_id')}")
                
            elif action == ApprovalAction.ESCALATE:
                # Verificar se pode escalar mais
                if request["escalation_count"] >= self.max_escalation_levels:
                    logger.warning(f"Solicitação {request_id} atingiu limite máximo de escalação")
                    return {
                        "success": False,
                        "error": f"Limite máximo de escalações atingido ({self.max_escalation_levels})",
                        "request_id": request_id
                    }
                
                # Escalar para próximo nível
                next_level = self._get_next_approval_level(ApprovalLevel(request["current_approval_level"]))
                request["current_approval_level"] = next_level.value
                request["escalation_count"] += 1
                request["updated_at"] = datetime.datetime.now().isoformat()
                logger.info(
                    f"Solicitação {request_id} escalada para {next_level.value} "
                    f"por {user_data.get('user_id')}"
                )
            
            else:
                # Outras ações (investigar, bloquear, etc.)
                request["status"] = f"{action.value}_in_progress"
                request["updated_at"] = datetime.datetime.now().isoformat()
                logger.info(
                    f"Solicitação {request_id} em processamento: {action.value} "
                    f"por {user_data.get('user_id')}"
                )
            
            # Armazenar trilha de auditoria se configurado
            if self.store_audit_trail:
                self._store_audit_entry(request_id, approval_entry, action)
            
            # Retornar resultado
            return {
                "success": True,
                "request_id": request_id,
                "status": request["status"],
                "action": action.value,
                "updated_at": request["updated_at"],
                "current_approval_level": request["current_approval_level"] 
                    if request["status"] == "pending" else None
            }
            
        except Exception as e:
            logger.error(f"Erro ao processar ação de aprovação: {str(e)}")
            return {
                "success": False,
                "error": str(e),
                "request_id": request_id
            }
    
    def get_approval_request(self, request_id: str) -> Optional[Dict[str, Any]]:
        """
        Obtém uma solicitação de aprovação pelo ID.
        
        Args:
            request_id: ID da solicitação de aprovação
            
        Returns:
            Solicitação de aprovação ou None se não encontrada
        """
        return self.pending_approvals.get(request_id)
    
    def get_pending_approvals(self, 
                           approval_level: Optional[ApprovalLevel] = None,
                           limit: int = 100, 
                           offset: int = 0) -> List[Dict[str, Any]]:
        """
        Obtém solicitações de aprovação pendentes.
        
        Args:
            approval_level: Filtrar por nível de aprovação requerido
            limit: Limite de resultados
            offset: Offset para paginação
            
        Returns:
            Lista de solicitações pendentes
        """
        pending = [req for req in self.pending_approvals.values() 
                 if req["status"] == "pending"]
        
        # Filtrar por nível de aprovação
        if approval_level:
            pending = [req for req in pending 
                     if req["current_approval_level"] == approval_level.value]
        
        # Ordenar por data de criação (mais antigo primeiro)
        pending.sort(key=lambda x: x["created_at"])
        
        # Aplicar paginação
        return pending[offset:offset + limit]
    
    def _get_level_rank(self, level: ApprovalLevel) -> int:
        """
        Obtém o rank numérico de um nível de aprovação.
        
        Args:
            level: Nível de aprovação
            
        Returns:
            Rank numérico do nível
        """
        ranks = {
            ApprovalLevel.AUTO_SYSTEM: 0,
            ApprovalLevel.L1_AGENT: 1,
            ApprovalLevel.L2_SPECIALIST: 2,
            ApprovalLevel.L3_SUPERVISOR: 3,
            ApprovalLevel.L4_MANAGER: 4,
            ApprovalLevel.L5_DIRECTOR: 5
        }
        
        return ranks.get(level, 0)
    
    def _get_next_approval_level(self, current_level: ApprovalLevel) -> ApprovalLevel:
        """
        Obtém o próximo nível de aprovação para escalação.
        
        Args:
            current_level: Nível atual
            
        Returns:
            Próximo nível de aprovação
        """
        rank = self._get_level_rank(current_level)
        next_ranks = {
            0: ApprovalLevel.L1_AGENT,
            1: ApprovalLevel.L2_SPECIALIST,
            2: ApprovalLevel.L3_SUPERVISOR,
            3: ApprovalLevel.L4_MANAGER,
            4: ApprovalLevel.L5_DIRECTOR,
            5: ApprovalLevel.L5_DIRECTOR  # Não há nível acima do diretor
        }
        
        return next_ranks.get(rank, ApprovalLevel.L5_DIRECTOR)
    
    def _store_audit_entry(self, 
                        request_id: str, 
                        approval_entry: Dict[str, Any],
                        action: ApprovalAction) -> None:
        """
        Armazena entrada de auditoria para uma ação de aprovação.
        
        Args:
            request_id: ID da solicitação
            approval_entry: Dados da aprovação
            action: Ação realizada
        """
        # Implementação simplificada - em produção, deve persistir em banco de dados
        audit_entry = {
            "timestamp": datetime.datetime.now().isoformat(),
            "request_id": request_id,
            "action": action.value,
            "approval_entry": approval_entry,
            "ip_address": "N/A",  # Em produção, coletar IP real
            "user_agent": "N/A"  # Em produção, coletar User-Agent real
        }
        
        # Log da auditoria
        logger.info(f"Audit trail: {json.dumps(audit_entry)}")


class ApprovalMatrixFactory:
    """
    Factory para criar instâncias da matriz de aprovação.
    """
    
    @staticmethod
    def create(config: Dict[str, Any]) -> ApprovalMatrix:
        """
        Cria uma instância da matriz de aprovação com as configurações fornecidas.
        
        Args:
            config: Configurações para a matriz
            
        Returns:
            Instância da ApprovalMatrix
        """
        return ApprovalMatrix(config)