#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Classe de Resultado de Validação Documental

Este módulo define a estrutura padrão para resultados de validação documental
no contexto do sistema IAM/TrustGuard da plataforma INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

from typing import Dict, Any, List, Optional
from dataclasses import dataclass, field
import json


@dataclass
class DocumentValidationResult:
    """
    Classe para representar o resultado de uma validação documental.
    
    Esta classe fornece uma estrutura padronizada para os resultados de validação,
    incluindo indicadores de validade, pontuações de confiança, etapas de validação,
    erros, avisos e metadados adicionais.
    """
    
    is_valid: bool = False
    confidence_score: float = 0.0
    validation_steps: List[Dict[str, Any]] = field(default_factory=list)
    errors: List[str] = field(default_factory=list)
    warnings: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o resultado para um dicionário."""
        return {
            "is_valid": self.is_valid,
            "confidence_score": self.confidence_score,
            "validation_steps": self.validation_steps,
            "errors": self.errors,
            "warnings": self.warnings,
            "metadata": self.metadata
        }
    
    def to_json(self) -> str:
        """Converte o resultado para uma string JSON."""
        return json.dumps(self.to_dict(), ensure_ascii=False, indent=2)
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'DocumentValidationResult':
        """Cria uma instância a partir de um dicionário."""
        return cls(
            is_valid=data.get("is_valid", False),
            confidence_score=data.get("confidence_score", 0.0),
            validation_steps=data.get("validation_steps", []),
            errors=data.get("errors", []),
            warnings=data.get("warnings", []),
            metadata=data.get("metadata", {})
        )
    
    def add_validation_step(self, name: str, description: str, result: bool, details: str = None) -> None:
        """
        Adiciona uma etapa de validação ao resultado.
        
        Args:
            name: Nome identificador da etapa de validação
            description: Descrição da etapa de validação
            result: Resultado da etapa (True para sucesso, False para falha)
            details: Detalhes adicionais sobre a etapa
        """
        self.validation_steps.append({
            "name": name,
            "description": description,
            "result": result,
            "details": details
        })
    
    def add_error(self, error_message: str) -> None:
        """Adiciona uma mensagem de erro ao resultado."""
        if error_message not in self.errors:
            self.errors.append(error_message)
            # Quando um erro é adicionado, o documento é considerado inválido
            self.is_valid = False
    
    def add_warning(self, warning_message: str) -> None:
        """Adiciona uma mensagem de aviso ao resultado."""
        if warning_message not in self.warnings:
            self.warnings.append(warning_message)
    
    def update_metadata(self, key: str, value: Any) -> None:
        """Atualiza um campo específico nos metadados."""
        self.metadata[key] = value
    
    def merge(self, other: 'DocumentValidationResult') -> 'DocumentValidationResult':
        """
        Combina este resultado com outro resultado de validação.
        Útil para consolidar resultados de múltiplos validadores.
        
        Args:
            other: Outro resultado de validação a ser combinado
            
        Returns:
            DocumentValidationResult: Um novo objeto com os resultados combinados
        """
        return DocumentValidationResult(
            is_valid=self.is_valid and other.is_valid,
            confidence_score=min(self.confidence_score, other.confidence_score),
            validation_steps=self.validation_steps + other.validation_steps,
            errors=list(set(self.errors + other.errors)),
            warnings=list(set(self.warnings + other.warnings)),
            metadata={**self.metadata, **other.metadata}
        )
    
    def get_risk_level(self) -> str:
        """
        Retorna o nível de risco associado à validação.
        
        Returns:
            str: Nível de risco (baixo, medio, alto)
        """
        if "risk_level" in self.metadata:
            return self.metadata["risk_level"]
        
        # Cálculo baseado na pontuação de confiança e presença de erros/avisos
        if not self.is_valid or self.confidence_score < 0.5:
            return "alto"
        elif self.warnings or self.confidence_score < 0.8:
            return "medio"
        else:
            return "baixo"
    
    def get_recommendation(self) -> str:
        """
        Retorna uma recomendação baseada no resultado da validação.
        
        Returns:
            str: Recomendação de ação
        """
        if not self.is_valid:
            return "Rejeitar documento - Falha na validação"
        elif self.get_risk_level() == "alto":
            return "Requerer validação manual adicional"
        elif self.get_risk_level() == "medio":
            return "Aceitar com verificação adicional"
        else:
            return "Aceitar documento - Validação bem-sucedida"