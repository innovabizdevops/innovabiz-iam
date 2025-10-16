#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Repositório de regras dinâmicas para detecção de anomalias comportamentais

Este módulo implementa o repositório para armazenar e gerenciar regras dinâmicas
e conjuntos de regras. Em produção, este módulo seria integrado com um banco de dados.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
import json
import logging
import datetime
import uuid
from typing import Dict, Any, List, Optional, Union, Tuple
from concurrent.futures import ThreadPoolExecutor

from .rule_model import Rule, RuleSet, RuleCondition, RuleGroup

# Configuração do logger
logger = logging.getLogger("iam.trustguard.rules_engine.repository")


class RuleRepository:
    """
    Repositório de regras dinâmicas.
    
    Esta implementação utiliza memória para simular um repositório.
    Em produção, seria integrado com um banco de dados.
    """
    
    def __init__(self, data_dir: Optional[str] = None):
        """
        Inicializa o repositório de regras.
        
        Args:
            data_dir: Diretório para armazenar dados (opcional)
        """
        self.data_dir = data_dir
        
        # Armazenamento em memória
        self.rules: Dict[str, Rule] = {}
        self.rulesets: Dict[str, RuleSet] = {}
        
        # Cache de regras por região
        self.rules_by_region: Dict[str, List[str]] = {}
        self.rulesets_by_region: Dict[str, List[str]] = {}
        
        # Executor para operações assíncronas
        self.executor = ThreadPoolExecutor(max_workers=2)
        
        # Carregar regras e conjuntos predefinidos
        self._load_default_rules()
        
        logger.info("RuleRepository inicializado")
    
    def _load_default_rules(self) -> None:
        """Carrega regras e conjuntos predefinidos para teste."""
        try:
            # Criar algumas regras de exemplo
            rule1 = Rule(
                id="rule-001",
                name="Login de Local Suspeito",
                description="Detecta tentativas de login de locais incomuns para o usuário",
                version="1.0.0",
                severity="high",
                category="authentication",
                condition=RuleCondition(
                    field="authentication.location.risk_score",
                    operator="gt",
                    value=0.8
                ),
                actions=["log", "alert", "challenge"],
                region="global",
                tags=["location", "authentication", "security"],
                score_multiplier=1.5
            )
            
            rule2 = Rule(
                id="rule-002",
                name="Múltiplas Falhas de Autenticação",
                description="Detecta múltiplas falhas de autenticação em curto período",
                version="1.0.0",
                severity="medium",
                category="authentication",
                condition=RuleGroup(
                    operator="and",
                    conditions=[
                        RuleCondition(
                            field="authentication.failures_count",
                            operator="gt",
                            value=3
                        ),
                        RuleCondition(
                            field="authentication.failures_window_minutes",
                            operator="lt",
                            value=15
                        )
                    ]
                ),
                actions=["log", "alert"],
                region="global",
                tags=["authentication", "security", "brute-force"],
                score_multiplier=1.2
            )
            
            rule3 = Rule(
                id="rule-003",
                name="Transação de Alto Valor",
                description="Detecta transações com valor acima do normal para o usuário",
                version="1.0.0",
                severity="high",
                category="transaction",
                condition=RuleCondition(
                    field="transaction.amount",
                    operator="gt",
                    value=10000
                ),
                actions=["log", "alert", "escalate"],
                region="BR",
                tags=["transaction", "amount", "financial"],
                score_multiplier=1.8
            )
            
            # Adicionar regras ao repositório
            self.add_rule(rule1)
            self.add_rule(rule2)
            self.add_rule(rule3)
            
            # Criar conjunto de regras de exemplo
            ruleset = RuleSet(
                id="ruleset-001",
                name="Regras de Segurança Básicas",
                description="Conjunto de regras de segurança básicas para detecção de anomalias",
                version="1.0.0",
                rules=[rule1, rule2],
                region="global",
                tags=["security", "basic", "authentication"]
            )
            
            # Adicionar conjunto ao repositório
            self.add_ruleset(ruleset)
            
        except Exception as e:
            logger.error(f"Erro ao carregar regras predefinidas: {str(e)}")
    
    def add_rule(self, rule: Rule) -> str:
        """
        Adiciona uma regra ao repositório.
        
        Args:
            rule: Regra a ser adicionada
            
        Returns:
            ID da regra adicionada
        """
        try:
            # Verificar se já existe uma regra com o mesmo ID
            if rule.id in self.rules:
                logger.warning(f"Regra com ID {rule.id} já existe, atualizando")
            
            # Adicionar ao dicionário
            self.rules[rule.id] = rule
            
            # Atualizar cache de região
            region = rule.region or "global"
            if region not in self.rules_by_region:
                self.rules_by_region[region] = []
            
            if rule.id not in self.rules_by_region[region]:
                self.rules_by_region[region].append(rule.id)
            
            logger.info(f"Regra adicionada/atualizada: {rule.id}")
            
            # Persistir regras (simulado)
            self._persist_rules_async()
            
            return rule.id
            
        except Exception as e:
            logger.error(f"Erro ao adicionar regra: {str(e)}")
            raise
    
    def get_rule(self, rule_id: str) -> Optional[Rule]:
        """
        Obtém uma regra pelo ID.
        
        Args:
            rule_id: ID da regra
            
        Returns:
            Regra encontrada ou None
        """
        return self.rules.get(rule_id)
    
    def update_rule(self, rule: Rule) -> bool:
        """
        Atualiza uma regra existente.
        
        Args:
            rule: Regra com dados atualizados
            
        Returns:
            True se a regra foi atualizada, False se não existe
        """
        if rule.id not in self.rules:
            logger.warning(f"Regra com ID {rule.id} não existe")
            return False
        
        # Atualizar regra
        old_rule = self.rules[rule.id]
        old_region = old_rule.region or "global"
        
        # Atualizar campos
        rule.updated_at = datetime.datetime.now()
        
        # Adicionar ao dicionário
        self.rules[rule.id] = rule
        
        # Atualizar cache de região se mudou
        new_region = rule.region or "global"
        if old_region != new_region:
            if old_region in self.rules_by_region and rule.id in self.rules_by_region[old_region]:
                self.rules_by_region[old_region].remove(rule.id)
            
            if new_region not in self.rules_by_region:
                self.rules_by_region[new_region] = []
            
            if rule.id not in self.rules_by_region[new_region]:
                self.rules_by_region[new_region].append(rule.id)
        
        logger.info(f"Regra atualizada: {rule.id}")
        
        # Persistir regras (simulado)
        self._persist_rules_async()
        
        # Atualizar regra em conjuntos
        self._update_rule_in_rulesets(rule)
        
        return True
    
    def delete_rule(self, rule_id: str) -> bool:
        """
        Exclui uma regra pelo ID.
        
        Args:
            rule_id: ID da regra
            
        Returns:
            True se a regra foi excluída, False se não existe
        """
        if rule_id not in self.rules:
            logger.warning(f"Regra com ID {rule_id} não existe")
            return False
        
        # Obter regra
        rule = self.rules[rule_id]
        region = rule.region or "global"
        
        # Remover do dicionário
        del self.rules[rule_id]
        
        # Remover do cache de região
        if region in self.rules_by_region and rule_id in self.rules_by_region[region]:
            self.rules_by_region[region].remove(rule_id)
        
        logger.info(f"Regra excluída: {rule_id}")
        
        # Persistir regras (simulado)
        self._persist_rules_async()
        
        # Remover regra de conjuntos
        self._remove_rule_from_rulesets(rule_id)
        
        return True
    
    def get_rules(self, region: Optional[str] = None, tags: Optional[List[str]] = None) -> List[Rule]:
        """
        Obtém regras com filtros opcionais.
        
        Args:
            region: Código da região (opcional)
            tags: Lista de tags (opcional)
            
        Returns:
            Lista de regras
        """
        # Filtrar por região
        if region:
            # Obter regras da região específica e globais
            rule_ids = set()
            
            # Adicionar regras globais
            if "global" in self.rules_by_region:
                rule_ids.update(self.rules_by_region["global"])
            
            # Adicionar regras da região
            if region in self.rules_by_region:
                rule_ids.update(self.rules_by_region[region])
            
            rules = [self.rules[rule_id] for rule_id in rule_ids if rule_id in self.rules]
        else:
            # Todas as regras
            rules = list(self.rules.values())
        
        # Filtrar por tags
        if tags:
            filtered_rules = []
            for rule in rules:
                # Verificar se a regra tem todas as tags
                if all(tag in rule.tags for tag in tags):
                    filtered_rules.append(rule)
            rules = filtered_rules
        
        return rules
    
    def add_ruleset(self, ruleset: RuleSet) -> str:
        """
        Adiciona um conjunto de regras ao repositório.
        
        Args:
            ruleset: Conjunto de regras a ser adicionado
            
        Returns:
            ID do conjunto adicionado
        """
        try:
            # Verificar se já existe um conjunto com o mesmo ID
            if ruleset.id in self.rulesets:
                logger.warning(f"Conjunto de regras com ID {ruleset.id} já existe, atualizando")
            
            # Adicionar ao dicionário
            self.rulesets[ruleset.id] = ruleset
            
            # Atualizar cache de região
            region = ruleset.region or "global"
            if region not in self.rulesets_by_region:
                self.rulesets_by_region[region] = []
            
            if ruleset.id not in self.rulesets_by_region[region]:
                self.rulesets_by_region[region].append(ruleset.id)
            
            logger.info(f"Conjunto de regras adicionado/atualizado: {ruleset.id}")
            
            # Persistir conjuntos (simulado)
            self._persist_rulesets_async()
            
            return ruleset.id
            
        except Exception as e:
            logger.error(f"Erro ao adicionar conjunto de regras: {str(e)}")
            raise
    
    def get_ruleset(self, ruleset_id: str) -> Optional[RuleSet]:
        """
        Obtém um conjunto de regras pelo ID.
        
        Args:
            ruleset_id: ID do conjunto
            
        Returns:
            Conjunto de regras encontrado ou None
        """
        return self.rulesets.get(ruleset_id)
    
    def update_ruleset(self, ruleset: RuleSet) -> bool:
        """
        Atualiza um conjunto de regras existente.
        
        Args:
            ruleset: Conjunto com dados atualizados
            
        Returns:
            True se o conjunto foi atualizado, False se não existe
        """
        if ruleset.id not in self.rulesets:
            logger.warning(f"Conjunto de regras com ID {ruleset.id} não existe")
            return False
        
        # Atualizar conjunto
        old_ruleset = self.rulesets[ruleset.id]
        old_region = old_ruleset.region or "global"
        
        # Atualizar campos
        ruleset.updated_at = datetime.datetime.now()
        
        # Adicionar ao dicionário
        self.rulesets[ruleset.id] = ruleset
        
        # Atualizar cache de região se mudou
        new_region = ruleset.region or "global"
        if old_region != new_region:
            if old_region in self.rulesets_by_region and ruleset.id in self.rulesets_by_region[old_region]:
                self.rulesets_by_region[old_region].remove(ruleset.id)
            
            if new_region not in self.rulesets_by_region:
                self.rulesets_by_region[new_region] = []
            
            if ruleset.id not in self.rulesets_by_region[new_region]:
                self.rulesets_by_region[new_region].append(ruleset.id)
        
        logger.info(f"Conjunto de regras atualizado: {ruleset.id}")
        
        # Persistir conjuntos (simulado)
        self._persist_rulesets_async()
        
        return True
    
    def delete_ruleset(self, ruleset_id: str) -> bool:
        """
        Exclui um conjunto de regras pelo ID.
        
        Args:
            ruleset_id: ID do conjunto
            
        Returns:
            True se o conjunto foi excluído, False se não existe
        """
        if ruleset_id not in self.rulesets:
            logger.warning(f"Conjunto de regras com ID {ruleset_id} não existe")
            return False
        
        # Obter conjunto
        ruleset = self.rulesets[ruleset_id]
        region = ruleset.region or "global"
        
        # Remover do dicionário
        del self.rulesets[ruleset_id]
        
        # Remover do cache de região
        if region in self.rulesets_by_region and ruleset_id in self.rulesets_by_region[region]:
            self.rulesets_by_region[region].remove(ruleset_id)
        
        logger.info(f"Conjunto de regras excluído: {ruleset_id}")
        
        # Persistir conjuntos (simulado)
        self._persist_rulesets_async()
        
        return True
    
    def get_rulesets(self, region: Optional[str] = None, tags: Optional[List[str]] = None) -> List[RuleSet]:
        """
        Obtém conjuntos de regras com filtros opcionais.
        
        Args:
            region: Código da região (opcional)
            tags: Lista de tags (opcional)
            
        Returns:
            Lista de conjuntos de regras
        """
        # Filtrar por região
        if region:
            # Obter conjuntos da região específica e globais
            ruleset_ids = set()
            
            # Adicionar conjuntos globais
            if "global" in self.rulesets_by_region:
                ruleset_ids.update(self.rulesets_by_region["global"])
            
            # Adicionar conjuntos da região
            if region in self.rulesets_by_region:
                ruleset_ids.update(self.rulesets_by_region[region])
            
            rulesets = [self.rulesets[ruleset_id] for ruleset_id in ruleset_ids if ruleset_id in self.rulesets]
        else:
            # Todos os conjuntos
            rulesets = list(self.rulesets.values())
        
        # Filtrar por tags
        if tags:
            filtered_rulesets = []
            for ruleset in rulesets:
                # Verificar se o conjunto tem todas as tags
                if all(tag in ruleset.tags for tag in tags):
                    filtered_rulesets.append(ruleset)
            rulesets = filtered_rulesets
        
        return rulesets
    
    def _update_rule_in_rulesets(self, rule: Rule) -> None:
        """
        Atualiza uma regra em todos os conjuntos que a contêm.
        
        Args:
            rule: Regra atualizada
        """
        for ruleset in self.rulesets.values():
            # Verificar se a regra está no conjunto
            for i, existing_rule in enumerate(ruleset.rules):
                if existing_rule.id == rule.id:
                    # Atualizar regra no conjunto
                    ruleset.rules[i] = rule
                    ruleset.updated_at = datetime.datetime.now()
                    logger.debug(f"Regra {rule.id} atualizada no conjunto {ruleset.id}")
                    
                    # Persistir conjuntos (simulado)
                    self._persist_rulesets_async()
                    break
    
    def _remove_rule_from_rulesets(self, rule_id: str) -> None:
        """
        Remove uma regra de todos os conjuntos que a contêm.
        
        Args:
            rule_id: ID da regra
        """
        for ruleset in self.rulesets.values():
            # Verificar se a regra está no conjunto
            for i, existing_rule in enumerate(ruleset.rules):
                if existing_rule.id == rule_id:
                    # Remover regra do conjunto
                    ruleset.rules.pop(i)
                    ruleset.updated_at = datetime.datetime.now()
                    logger.debug(f"Regra {rule_id} removida do conjunto {ruleset.id}")
                    
                    # Persistir conjuntos (simulado)
                    self._persist_rulesets_async()
                    break
    
    def _persist_rules_async(self) -> None:
        """Persiste regras de forma assíncrona (simulado)."""
        if self.data_dir:
            self.executor.submit(self._persist_rules)
    
    def _persist_rulesets_async(self) -> None:
        """Persiste conjuntos de regras de forma assíncrona (simulado)."""
        if self.data_dir:
            self.executor.submit(self._persist_rulesets)
    
    def _persist_rules(self) -> None:
        """Persiste regras em arquivo (simulado)."""
        if not self.data_dir:
            return
            
        try:
            # Criar diretório se não existir
            os.makedirs(self.data_dir, exist_ok=True)
            
            # Caminho do arquivo
            rules_file = os.path.join(self.data_dir, "rules.json")
            
            # Converter regras para dicionários
            rules_data = {rule_id: rule.to_dict() for rule_id, rule in self.rules.items()}
            
            # Salvar em arquivo
            with open(rules_file, "w", encoding="utf-8") as f:
                json.dump(rules_data, f, indent=2)
                
            logger.debug("Regras persistidas em arquivo")
            
        except Exception as e:
            logger.error(f"Erro ao persistir regras: {str(e)}")
    
    def _persist_rulesets(self) -> None:
        """Persiste conjuntos de regras em arquivo (simulado)."""
        if not self.data_dir:
            return
            
        try:
            # Criar diretório se não existir
            os.makedirs(self.data_dir, exist_ok=True)
            
            # Caminho do arquivo
            rulesets_file = os.path.join(self.data_dir, "rulesets.json")
            
            # Converter conjuntos para dicionários
            rulesets_data = {ruleset_id: ruleset.to_dict() for ruleset_id, ruleset in self.rulesets.items()}
            
            # Salvar em arquivo
            with open(rulesets_file, "w", encoding="utf-8") as f:
                json.dump(rulesets_data, f, indent=2)
                
            logger.debug("Conjuntos de regras persistidos em arquivo")
            
        except Exception as e:
            logger.error(f"Erro ao persistir conjuntos de regras: {str(e)}")
    
    def _load_rules(self) -> None:
        """Carrega regras de arquivo (simulado)."""
        if not self.data_dir:
            return
            
        try:
            # Caminho do arquivo
            rules_file = os.path.join(self.data_dir, "rules.json")
            
            # Verificar se o arquivo existe
            if not os.path.isfile(rules_file):
                logger.debug("Arquivo de regras não encontrado")
                return
            
            # Carregar do arquivo
            with open(rules_file, "r", encoding="utf-8") as f:
                rules_data = json.load(f)
            
            # Converter para objetos Rule
            for rule_id, rule_data in rules_data.items():
                rule = Rule.from_dict(rule_data)
                self.rules[rule_id] = rule
                
                # Atualizar cache de região
                region = rule.region or "global"
                if region not in self.rules_by_region:
                    self.rules_by_region[region] = []
                
                if rule.id not in self.rules_by_region[region]:
                    self.rules_by_region[region].append(rule.id)
            
            logger.debug(f"Carregadas {len(rules_data)} regras de arquivo")
            
        except Exception as e:
            logger.error(f"Erro ao carregar regras: {str(e)}")
    
    def _load_rulesets(self) -> None:
        """Carrega conjuntos de regras de arquivo (simulado)."""
        if not self.data_dir:
            return
            
        try:
            # Caminho do arquivo
            rulesets_file = os.path.join(self.data_dir, "rulesets.json")
            
            # Verificar se o arquivo existe
            if not os.path.isfile(rulesets_file):
                logger.debug("Arquivo de conjuntos de regras não encontrado")
                return
            
            # Carregar do arquivo
            with open(rulesets_file, "r", encoding="utf-8") as f:
                rulesets_data = json.load(f)
            
            # Converter para objetos RuleSet
            for ruleset_id, ruleset_data in rulesets_data.items():
                ruleset = RuleSet.from_dict(ruleset_data)
                self.rulesets[ruleset_id] = ruleset
                
                # Atualizar cache de região
                region = ruleset.region or "global"
                if region not in self.rulesets_by_region:
                    self.rulesets_by_region[region] = []
                
                if ruleset.id not in self.rulesets_by_region[region]:
                    self.rulesets_by_region[region].append(ruleset.id)
            
            logger.debug(f"Carregados {len(rulesets_data)} conjuntos de regras de arquivo")
            
        except Exception as e:
            logger.error(f"Erro ao carregar conjuntos de regras: {str(e)}")