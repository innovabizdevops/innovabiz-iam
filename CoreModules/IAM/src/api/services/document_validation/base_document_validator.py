#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Classe Base para Validadores Documentais

Este módulo define a interface base para validadores documentais
no contexto do sistema IAM/TrustGuard da plataforma INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import abc
import logging
from typing import Dict, Any, List, Optional
from datetime import datetime

# Importando classe de resultado
from src.api.services.document_validation.document_validation_result import DocumentValidationResult

# Configuração de logging
logger = logging.getLogger("document_validator")


class BaseDocumentValidator(abc.ABC):
    """
    Classe abstrata base para validadores documentais específicos por região.
    Define a interface comum para todos os validadores regionais.
    """
    
    def __init__(self):
        """Inicializa o validador com configurações padrão."""
        self.name = "Base Document Validator"
        self.country_code = None
        self.rules = {}
        self.last_validation_timestamp = None
    
    @abc.abstractmethod
    def validate(self, document_data: Dict[str, Any]) -> DocumentValidationResult:
        """
        Método principal para validação de documentos.
        Deve ser implementado por todas as subclasses específicas por região.
        
        Args:
            document_data: Dicionário contendo os dados do documento a validar
            
        Returns:
            DocumentValidationResult: Resultado detalhado da validação
        """
        pass
    
    def load_country_specific_rules(self) -> None:
        """
        Carrega regras específicas para o país.
        Deve ser sobrescrito por implementações concretas.
        """
        pass
    
    def validate_document_format(self, doc_number: str, doc_type: str) -> Dict[str, Any]:
        """
        Valida o formato do número do documento usando as regras definidas.
        
        Args:
            doc_number: Número do documento a validar
            doc_type: Tipo de documento (ex: "bi", "passport", "nif")
            
        Returns:
            dict: Resultado da validação de formato com campos 'valid' e 'details'
        """
        if doc_type not in self.rules:
            return {
                "valid": False,
                "details": f"Tipo de documento '{doc_type}' não reconhecido"
            }
            
        rule = self.rules[doc_type]
        
        # Verificar se o padrão está definido
        if "pattern" not in rule:
            return {
                "valid": True,  # Assume válido se não houver padrão definido
                "details": "Sem regras de formato definidas para este tipo"
            }
            
        import re
        pattern_valid = bool(re.match(rule["pattern"], doc_number))
        
        return {
            "valid": pattern_valid,
            "details": f"Formato {'válido' if pattern_valid else 'inválido'} para {doc_type}"
        }
    
    def check_document_expiration(self, expiry_date: str, doc_type: str) -> Dict[str, Any]:
        """
        Verifica se o documento está dentro do prazo de validade.
        
        Args:
            expiry_date: Data de validade do documento (formato ISO: YYYY-MM-DD)
            doc_type: Tipo de documento
            
        Returns:
            dict: Resultado da verificação com campos 'valid' e 'details'
        """
        if not expiry_date:
            return {
                "valid": False,
                "details": "Data de expiração não fornecida"
            }
            
        try:
            from datetime import datetime
            expiry = datetime.fromisoformat(expiry_date) if isinstance(expiry_date, str) else expiry_date
            now = datetime.now()
            is_valid = expiry > now
            
            days_remaining = (expiry - now).days if is_valid else 0
            
            return {
                "valid": is_valid,
                "details": f"Documento {'válido' if is_valid else 'expirado'}. " + 
                          (f"Dias restantes: {days_remaining}" if is_valid else f"Expirou em {expiry_date}")
            }
        except (ValueError, TypeError) as e:
            return {
                "valid": False,
                "details": f"Erro ao verificar data de expiração: {str(e)}"
            }
    
    def _calculate_risk_level(self, is_valid: bool, confidence_score: float, 
                             doc_type: str, has_external_verification: bool) -> str:
        """
        Método comum para cálculo de nível de risco baseado nos resultados da validação.
        
        Args:
            is_valid: Se o documento foi considerado válido
            confidence_score: Pontuação de confiança
            doc_type: Tipo de documento
            has_external_verification: Se houve verificação externa bem-sucedida
            
        Returns:
            str: Nível de risco (baixo, medio, alto)
        """
        # Documentos considerados de alto risco por padrão
        high_risk_docs = ["passport", "id_card"]
        
        # Se o documento não é válido
        if not is_valid:
            return "alto"
            
        # Se é um documento de alto risco sem verificação externa
        if doc_type in high_risk_docs and not has_external_verification:
            return "alto"
            
        # Baseado na pontuação de confiança
        if confidence_score >= 0.9:
            return "baixo"
        elif confidence_score >= 0.7:
            return "medio"
        else:
            return "alto"
    
    def _log_validation_activity(self, document_data: Dict[str, Any], result: DocumentValidationResult) -> None:
        """
        Registra a atividade de validação para auditoria e monitoramento.
        
        Args:
            document_data: Dados do documento validado
            result: Resultado da validação
        """
        self.last_validation_timestamp = datetime.now()
        
        # Dados sanitizados para log (remove dados sensíveis)
        log_data = {
            "validator": self.name,
            "country_code": self.country_code,
            "doc_type": document_data.get("type", "unknown"),
            "result_valid": result.is_valid,
            "confidence_score": result.confidence_score,
            "risk_level": result.metadata.get("risk_level", "unknown"),
            "timestamp": self.last_validation_timestamp.isoformat(),
            "error_count": len(result.errors),
            "warning_count": len(result.warnings)
        }
        
        # Log para auditoria
        logger.info(f"Document validation: {log_data}")
        
        # Aqui poderia ser adicionado código para enviar logs para sistemas centralizados
        # como Elasticsearch, Splunk, etc.
    
    def get_validator_info(self) -> Dict[str, Any]:
        """
        Retorna informações sobre o validador.
        
        Returns:
            dict: Informações sobre o validador
        """
        return {
            "name": self.name,
            "country_code": self.country_code,
            "supported_document_types": list(self.rules.keys()),
            "last_validation": self.last_validation_timestamp.isoformat() if self.last_validation_timestamp else None
        }