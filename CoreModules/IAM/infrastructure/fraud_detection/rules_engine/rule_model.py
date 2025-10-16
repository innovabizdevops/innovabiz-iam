#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Modelo de regras dinâmicas para detecção de anomalias comportamentais

Este módulo define as estruturas de dados para o modelo de regras dinâmicas
utilizadas na detecção de anomalias comportamentais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from enum import Enum, auto
from typing import Dict, Any, List, Optional, Union, Tuple, Callable
from dataclasses import dataclass, field
import datetime
import uuid
import json


class RuleOperator(str, Enum):
    """Operadores para condições de regras."""
    
    EQUAL = "eq"  # ==
    NOT_EQUAL = "ne"  # !=
    GREATER_THAN = "gt"  # >
    GREATER_EQUAL = "ge"  # >=
    LESS_THAN = "lt"  # <
    LESS_EQUAL = "le"  # <=
    IN = "in"  # in (lista)
    NOT_IN = "not_in"  # not in (lista)
    CONTAINS = "contains"  # contém substring
    STARTS_WITH = "starts_with"  # começa com substring
    ENDS_WITH = "ends_with"  # termina com substring
    REGEX = "regex"  # expressão regular
    EXISTS = "exists"  # campo existe
    NOT_EXISTS = "not_exists"  # campo não existe
    BETWEEN = "between"  # entre dois valores (inclusive)
    NOT_BETWEEN = "not_between"  # não entre dois valores
    ALL = "all"  # todos os valores atendem
    ANY = "any"  # qualquer valor atende
    NONE = "none"  # nenhum valor atende


class RuleValueType(str, Enum):
    """Tipos de valores para condições de regras."""
    
    STRING = "string"
    NUMBER = "number"
    BOOLEAN = "boolean"
    DATETIME = "datetime"
    LIST = "list"
    DICT = "dict"
    NULL = "null"


class RuleLogicalOperator(str, Enum):
    """Operadores lógicos para combinação de condições."""
    
    AND = "and"
    OR = "or"
    NOT = "not"


class RuleSeverity(str, Enum):
    """Severidade da regra."""
    
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class RuleCategory(str, Enum):
    """Categoria da regra."""
    
    AUTHENTICATION = "authentication"  # Regra relacionada a autenticação
    TRANSACTION = "transaction"  # Regra relacionada a transações
    SESSION = "session"  # Regra relacionada a sessões
    DEVICE = "device"  # Regra relacionada a dispositivos
    LOCATION = "location"  # Regra relacionada a localização
    PROFILE = "profile"  # Regra relacionada ao perfil do usuário
    COMBINED = "combined"  # Regra que combina múltiplos aspectos


class RuleAction(str, Enum):
    """Ação a ser tomada quando a regra é ativada."""
    
    LOG = "log"  # Apenas registrar em log
    ALERT = "alert"  # Gerar alerta
    BLOCK = "block"  # Bloquear ação
    CHALLENGE = "challenge"  # Solicitar verificação adicional
    NOTIFY = "notify"  # Notificar usuário/administrador
    ESCALATE = "escalate"  # Escalar para aprovação
    CUSTOM = "custom"  # Ação personalizada


@dataclass
class RuleCondition:
    """Condição de uma regra."""
    
    field: str  # Campo a ser avaliado (caminho usando notação de ponto)
    operator: RuleOperator  # Operador da condição
    value: Any  # Valor para comparação
    value_type: Optional[RuleValueType] = None  # Tipo do valor
    
    def __post_init__(self):
        """Inicializa o tipo de valor com base no valor fornecido."""
        if self.value_type is None:
            if isinstance(self.value, str):
                self.value_type = RuleValueType.STRING
            elif isinstance(self.value, (int, float)):
                self.value_type = RuleValueType.NUMBER
            elif isinstance(self.value, bool):
                self.value_type = RuleValueType.BOOLEAN
            elif isinstance(self.value, (list, tuple)):
                self.value_type = RuleValueType.LIST
            elif isinstance(self.value, dict):
                self.value_type = RuleValueType.DICT
            elif self.value is None:
                self.value_type = RuleValueType.NULL
            elif isinstance(self.value, (datetime.datetime, datetime.date)):
                self.value_type = RuleValueType.DATETIME
    
    def to_dict(self) -> Dict[str, Any]:
        """
        Converte a condição para dicionário.
        
        Returns:
            Dicionário com dados da condição
        """
        return {
            "field": self.field,
            "operator": self.operator.value if isinstance(self.operator, Enum) else self.operator,
            "value": self.value,
            "value_type": self.value_type.value if isinstance(self.value_type, Enum) else self.value_type,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "RuleCondition":
        """
        Cria uma condição a partir de um dicionário.
        
        Args:
            data: Dicionário com dados da condição
            
        Returns:
            Instância de RuleCondition
        """
        return cls(
            field=data["field"],
            operator=RuleOperator(data["operator"]) if "operator" in data else RuleOperator.EQUAL,
            value=data["value"],
            value_type=RuleValueType(data["value_type"]) if "value_type" in data else None,
        )


@dataclass
class RuleGroup:
    """Grupo de condições com operador lógico."""
    
    operator: RuleLogicalOperator  # Operador lógico para combinar condições/grupos
    conditions: List[Union[RuleCondition, "RuleGroup"]] = field(default_factory=list)  # Lista de condições ou grupos
    
    def to_dict(self) -> Dict[str, Any]:
        """
        Converte o grupo para dicionário.
        
        Returns:
            Dicionário com dados do grupo
        """
        return {
            "operator": self.operator.value if isinstance(self.operator, Enum) else self.operator,
            "conditions": [
                cond.to_dict() if hasattr(cond, "to_dict") else cond
                for cond in self.conditions
            ],
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "RuleGroup":
        """
        Cria um grupo a partir de um dicionário.
        
        Args:
            data: Dicionário com dados do grupo
            
        Returns:
            Instância de RuleGroup
        """
        group = cls(
            operator=RuleLogicalOperator(data["operator"]) if "operator" in data else RuleLogicalOperator.AND,
        )
        
        # Processar condições/grupos
        for cond_data in data.get("conditions", []):
            if "operator" in cond_data and cond_data["operator"] in [op.value for op in RuleLogicalOperator]:
                # É um grupo
                group.conditions.append(RuleGroup.from_dict(cond_data))
            else:
                # É uma condição
                group.conditions.append(RuleCondition.from_dict(cond_data))
        
        return group


@dataclass
class Rule:
    """Regra dinâmica para detecção de anomalias comportamentais."""
    
    id: str  # ID único da regra
    name: str  # Nome da regra
    description: str  # Descrição da regra
    version: str  # Versão da regra
    severity: RuleSeverity  # Severidade da regra
    category: RuleCategory  # Categoria da regra
    condition: Union[RuleCondition, RuleGroup]  # Condição ou grupo de condições
    actions: List[RuleAction]  # Ações a serem executadas quando a regra é ativada
    region: Optional[str] = None  # Código da região (opcional)
    enabled: bool = True  # Status da regra (habilitada/desabilitada)
    tags: List[str] = field(default_factory=list)  # Tags para categorização
    score_multiplier: float = 1.0  # Multiplicador para o score de anomalia
    created_at: datetime.datetime = field(default_factory=datetime.datetime.now)  # Data de criação
    updated_at: Optional[datetime.datetime] = None  # Data da última atualização
    metadata: Dict[str, Any] = field(default_factory=dict)  # Metadados adicionais
    custom_action: Optional[str] = None  # Nome da ação personalizada
    
    def __post_init__(self):
        """Inicializa valores padrão se necessário."""
        if not self.id:
            self.id = str(uuid.uuid4())
            
        if not self.created_at:
            self.created_at = datetime.datetime.now()
    
    def to_dict(self) -> Dict[str, Any]:
        """
        Converte a regra para dicionário.
        
        Returns:
            Dicionário com dados da regra
        """
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "version": self.version,
            "severity": self.severity.value if isinstance(self.severity, Enum) else self.severity,
            "category": self.category.value if isinstance(self.category, Enum) else self.category,
            "condition": self.condition.to_dict() if hasattr(self.condition, "to_dict") else self.condition,
            "actions": [action.value if isinstance(action, Enum) else action for action in self.actions],
            "region": self.region,
            "enabled": self.enabled,
            "tags": self.tags,
            "score_multiplier": self.score_multiplier,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "metadata": self.metadata,
            "custom_action": self.custom_action,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Rule":
        """
        Cria uma regra a partir de um dicionário.
        
        Args:
            data: Dicionário com dados da regra
            
        Returns:
            Instância de Rule
        """
        # Processar condição/grupo
        condition_data = data.get("condition", {})
        if "operator" in condition_data and condition_data["operator"] in [op.value for op in RuleLogicalOperator]:
            condition = RuleGroup.from_dict(condition_data)
        else:
            condition = RuleCondition.from_dict(condition_data)
        
        # Processar datas
        created_at = data.get("created_at")
        if created_at and isinstance(created_at, str):
            created_at = datetime.datetime.fromisoformat(created_at)
            
        updated_at = data.get("updated_at")
        if updated_at and isinstance(updated_at, str):
            updated_at = datetime.datetime.fromisoformat(updated_at)
        
        return cls(
            id=data.get("id", str(uuid.uuid4())),
            name=data.get("name", ""),
            description=data.get("description", ""),
            version=data.get("version", "1.0.0"),
            severity=RuleSeverity(data["severity"]) if "severity" in data else RuleSeverity.MEDIUM,
            category=RuleCategory(data["category"]) if "category" in data else RuleCategory.COMBINED,
            condition=condition,
            actions=[RuleAction(action) for action in data.get("actions", [])],
            region=data.get("region"),
            enabled=data.get("enabled", True),
            tags=data.get("tags", []),
            score_multiplier=data.get("score_multiplier", 1.0),
            created_at=created_at,
            updated_at=updated_at,
            metadata=data.get("metadata", {}),
            custom_action=data.get("custom_action"),
        )


@dataclass
class RuleSet:
    """Conjunto de regras."""
    
    id: str  # ID único do conjunto de regras
    name: str  # Nome do conjunto de regras
    description: str  # Descrição do conjunto de regras
    version: str  # Versão do conjunto de regras
    rules: List[Rule] = field(default_factory=list)  # Lista de regras
    region: Optional[str] = None  # Código da região (opcional)
    enabled: bool = True  # Status do conjunto (habilitado/desabilitado)
    tags: List[str] = field(default_factory=list)  # Tags para categorização
    created_at: datetime.datetime = field(default_factory=datetime.datetime.now)  # Data de criação
    updated_at: Optional[datetime.datetime] = None  # Data da última atualização
    metadata: Dict[str, Any] = field(default_factory=dict)  # Metadados adicionais
    
    def __post_init__(self):
        """Inicializa valores padrão se necessário."""
        if not self.id:
            self.id = str(uuid.uuid4())
            
        if not self.created_at:
            self.created_at = datetime.datetime.now()
    
    def add_rule(self, rule: Rule) -> None:
        """
        Adiciona uma regra ao conjunto.
        
        Args:
            rule: Regra a ser adicionada
        """
        self.rules.append(rule)
        self.updated_at = datetime.datetime.now()
    
    def remove_rule(self, rule_id: str) -> bool:
        """
        Remove uma regra do conjunto.
        
        Args:
            rule_id: ID da regra a ser removida
            
        Returns:
            True se a regra foi removida, False caso contrário
        """
        for i, rule in enumerate(self.rules):
            if rule.id == rule_id:
                self.rules.pop(i)
                self.updated_at = datetime.datetime.now()
                return True
        return False
    
    def update_rule(self, rule: Rule) -> bool:
        """
        Atualiza uma regra no conjunto.
        
        Args:
            rule: Regra atualizada
            
        Returns:
            True se a regra foi atualizada, False caso contrário
        """
        for i, existing_rule in enumerate(self.rules):
            if existing_rule.id == rule.id:
                self.rules[i] = rule
                self.updated_at = datetime.datetime.now()
                return True
        return False
    
    def get_rule(self, rule_id: str) -> Optional[Rule]:
        """
        Obtém uma regra do conjunto pelo ID.
        
        Args:
            rule_id: ID da regra
            
        Returns:
            Regra encontrada ou None
        """
        for rule in self.rules:
            if rule.id == rule_id:
                return rule
        return None
    
    def to_dict(self) -> Dict[str, Any]:
        """
        Converte o conjunto de regras para dicionário.
        
        Returns:
            Dicionário com dados do conjunto
        """
        return {
            "id": self.id,
            "name": self.name,
            "description": self.description,
            "version": self.version,
            "rules": [rule.to_dict() for rule in self.rules],
            "region": self.region,
            "enabled": self.enabled,
            "tags": self.tags,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "metadata": self.metadata,
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "RuleSet":
        """
        Cria um conjunto de regras a partir de um dicionário.
        
        Args:
            data: Dicionário com dados do conjunto
            
        Returns:
            Instância de RuleSet
        """
        # Processar datas
        created_at = data.get("created_at")
        if created_at and isinstance(created_at, str):
            created_at = datetime.datetime.fromisoformat(created_at)
            
        updated_at = data.get("updated_at")
        if updated_at and isinstance(updated_at, str):
            updated_at = datetime.datetime.fromisoformat(updated_at)
        
        # Criar conjunto
        ruleset = cls(
            id=data.get("id", str(uuid.uuid4())),
            name=data.get("name", ""),
            description=data.get("description", ""),
            version=data.get("version", "1.0.0"),
            region=data.get("region"),
            enabled=data.get("enabled", True),
            tags=data.get("tags", []),
            created_at=created_at,
            updated_at=updated_at,
            metadata=data.get("metadata", {}),
        )
        
        # Adicionar regras
        for rule_data in data.get("rules", []):
            ruleset.add_rule(Rule.from_dict(rule_data))
        
        return ruleset