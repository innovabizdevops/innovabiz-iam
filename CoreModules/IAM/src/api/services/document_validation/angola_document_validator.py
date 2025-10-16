#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Validador Documental para Angola

Este módulo implementa validações específicas para documentos angolanos
no contexto do sistema IAM/TrustGuard da plataforma INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import re
import json
import logging
import hashlib
from typing import Dict, Any, List, Tuple, Optional
from datetime import datetime

# Configuração de logging
logger = logging.getLogger("angola_document_validator")

# Importações de módulos de integração
try:
    from src.api.services.integration.circ_angola_connector import CIRCAngolaConnector
    from src.api.services.document_validation.base_document_validator import BaseDocumentValidator
    from src.api.services.document_validation.document_validation_result import DocumentValidationResult
except ImportError:
    logger.warning("Importações não encontradas. Executando em modo isolado.")
    # Classes mock para testes isolados
    class CIRCAngolaConnector:
        def verify_document(self, doc_type, doc_number, name=None):
            return {"verified": True, "score": 0.95}

    class BaseDocumentValidator:
        def __init__(self):
            self.name = "BaseValidator"
            
        def validate(self, document_data):
            return {"valid": True}

    class DocumentValidationResult:
        def __init__(self, is_valid=False, confidence_score=0.0, validation_steps=None, 
                     errors=None, warnings=None, metadata=None):
            self.is_valid = is_valid
            self.confidence_score = confidence_score
            self.validation_steps = validation_steps or []
            self.errors = errors or []
            self.warnings = warnings or []
            self.metadata = metadata or {}


class AngolaDocumentValidator(BaseDocumentValidator):
    """
    Validador especializado para documentos angolanos, implementando regras
    específicas para BI, Passaporte, NIF e outros documentos angolanos.
    
    Esta classe é responsável por validar documentos angolanos de acordo com
    regras sintáticas, formatos e verificações externas com sistemas como CIRC.
    """
    
    def __init__(self):
        """Inicializa o validador com configurações específicas para Angola."""
        super().__init__()
        self.country_code = "AO"
        self.name = "Angola Document Validator"
        self.circ_connector = CIRCAngolaConnector()
        
        # Carregar configurações regionais específicas
        self.load_country_specific_rules()
        
    def load_country_specific_rules(self):
        """
        Carrega regras específicas para validação de documentos angolanos.
        Inclui expressões regulares, regras de formato e configurações de validação.
        """
        self.rules = {
            "bi": {
                "pattern": r"^[0-9]{9}[A-Z]{2}[0-9]{3}$",
                "length": 14,
                "expiry_years": 5,
                "requires_external_verification": True,
                "high_risk_without_verification": True,
            },
            "passport": {
                "pattern": r"^[A-Z]{1,2}[0-9]{6,7}$",
                "length": [7, 8, 9],  # Passaportes angolanos podem ter diferentes comprimentos
                "expiry_years": 5,
                "requires_external_verification": True,
                "high_risk_without_verification": True,
            },
            "nif": {  # NIF Angolano (Número de Identificação Fiscal)
                "pattern": r"^[0-9]{10}$",
                "length": 10,
                "expiry_years": None,  # Não expira
                "requires_external_verification": True,
                "high_risk_without_verification": False,
            },
            "military_id": {
                "pattern": r"^[0-9]{10}[A-Z]$",
                "length": 11,
                "expiry_years": 5,
                "requires_external_verification": True,
                "high_risk_without_verification": True,
            },
            "voter_card": {
                "pattern": r"^[0-9]{12}$",
                "length": 12,
                "expiry_years": 10,
                "requires_external_verification": False,
                "high_risk_without_verification": False,
            }
        }
        
        # Regras para verificação de dígitos
        self.verification_digit_rules = {
            "nif": {
                "modulus": 11,
                "weights": [2, 3, 4, 5, 6, 7, 8, 9, 10]
            }
        }
        
        # Lista de prefixos válidos para BI por província
        self.valid_bi_prefixes = {
            "luanda": ["000", "001", "002"],
            "benguela": ["003", "004"],
            "huambo": ["005", "006"],
            "huila": ["007", "008"],
            # Outras províncias
            "default": ["999"]  # Prefixo genérico para casos não especificados
        }
        
    def validate(self, document_data: Dict[str, Any]) -> DocumentValidationResult:
        """
        Valida um documento angolano com base nos critérios específicos do país.
        
        Args:
            document_data: Dicionário contendo dados do documento a ser validado
            
        Returns:
            DocumentValidationResult: Resultado detalhado da validação
        """
        try:
            doc_type = document_data.get("type", "").lower()
            doc_number = document_data.get("number", "")
            issue_date = document_data.get("issue_date", "")
            expiry_date = document_data.get("expiry_date", "")
            issuer = document_data.get("issuer", "")
            holder_name = document_data.get("holder_name", "")
            holder_birth_date = document_data.get("holder_birth_date", "")
            additional_data = document_data.get("additional_data", {})
            
            # Inicializar resultado da validação
            validation_steps = []
            errors = []
            warnings = []
            metadata = {
                "country": "Angola",
                "document_type": doc_type,
                "verification_timestamp": datetime.now().isoformat()
            }
            
            # Verificar se o tipo de documento é suportado
            if doc_type not in self.rules:
                errors.append(f"Tipo de documento '{doc_type}' não é suportado para Angola")
                return DocumentValidationResult(
                    is_valid=False,
                    confidence_score=0.0,
                    validation_steps=validation_steps,
                    errors=errors,
                    warnings=warnings,
                    metadata=metadata
                )
            
            # Obter regras específicas para o tipo de documento
            rule = self.rules[doc_type]
            
            # 1. Validação de formato usando regex
            pattern_valid = bool(re.match(rule["pattern"], doc_number))
            validation_steps.append({
                "name": "formato_documento",
                "description": "Validação do formato do documento usando expressão regular",
                "result": pattern_valid,
                "details": f"Formato {'válido' if pattern_valid else 'inválido'} para {doc_type}"
            })
            
            if not pattern_valid:
                errors.append(f"Formato do número do documento {doc_type} inválido: {doc_number}")
                
            # 2. Validação de comprimento
            length_valid = False
            if isinstance(rule["length"], int):
                length_valid = len(doc_number) == rule["length"]
            else:  # lista de comprimentos possíveis
                length_valid = len(doc_number) in rule["length"]
                
            validation_steps.append({
                "name": "comprimento_documento",
                "description": "Validação do comprimento do número do documento",
                "result": length_valid,
                "details": f"Comprimento {'válido' if length_valid else 'inválido'} para {doc_type}"
            })
            
            if not length_valid:
                errors.append(f"Comprimento do número do documento {doc_type} inválido: {len(doc_number)}")
            
            # 3. Validação de dígito verificador (se aplicável)
            checkdigit_valid = True  # Assume válido por padrão
            if doc_type in self.verification_digit_rules:
                checkdigit_valid = self._validate_check_digit(doc_number, doc_type)
                validation_steps.append({
                    "name": "digito_verificador",
                    "description": "Validação do dígito verificador",
                    "result": checkdigit_valid,
                    "details": f"Dígito verificador {'válido' if checkdigit_valid else 'inválido'}"
                })
                
                if not checkdigit_valid:
                    errors.append("Dígito verificador inválido")
            
            # 4. Validação de prefixo para BI
            prefix_valid = True  # Assume válido por padrão
            if doc_type == "bi":
                prefix = doc_number[:3]
                valid_prefixes = []
                for region_prefixes in self.valid_bi_prefixes.values():
                    valid_prefixes.extend(region_prefixes)
                
                prefix_valid = prefix in valid_prefixes
                validation_steps.append({
                    "name": "prefixo_bi",
                    "description": "Validação do prefixo do BI por província",
                    "result": prefix_valid,
                    "details": f"Prefixo do BI {'válido' if prefix_valid else 'inválido'}: {prefix}"
                })
                
                if not prefix_valid:
                    warnings.append(f"Prefixo do BI não reconhecido: {prefix}")
                    
            # 5. Verificação de validade da data de expiração
            expiry_valid = True
            if rule["expiry_years"] and expiry_date:
                try:
                    expiry = datetime.fromisoformat(expiry_date) if isinstance(expiry_date, str) else expiry_date
                    now = datetime.now()
                    expiry_valid = expiry > now
                    
                    validation_steps.append({
                        "name": "data_validade",
                        "description": "Verificação da data de validade do documento",
                        "result": expiry_valid,
                        "details": f"Documento {'válido' if expiry_valid else 'expirado'}: {expiry_date}"
                    })
                    
                    if not expiry_valid:
                        errors.append(f"Documento expirado em: {expiry_date}")
                except (ValueError, TypeError):
                    warnings.append(f"Formato de data de expiração inválido: {expiry_date}")
                    expiry_valid = False
                    validation_steps.append({
                        "name": "data_validade",
                        "description": "Verificação da data de validade do documento",
                        "result": False,
                        "details": f"Formato de data inválido: {expiry_date}"
                    })
            
            # 6. Verificação externa com CIRC Angola (se disponível e necessário)
            external_verification = {"verified": False, "score": 0.0}
            if rule["requires_external_verification"]:
                try:
                    external_verification = self.circ_connector.verify_document(
                        doc_type=doc_type,
                        doc_number=doc_number,
                        name=holder_name
                    )
                    
                    external_valid = external_verification.get("verified", False)
                    external_score = external_verification.get("score", 0.0)
                    
                    validation_steps.append({
                        "name": "verificacao_circ",
                        "description": "Verificação externa com CIRC Angola",
                        "result": external_valid,
                        "details": f"Verificação CIRC {'bem-sucedida' if external_valid else 'falhou'} com score {external_score:.2f}"
                    })
                    
                    if not external_valid:
                        errors.append("Documento não verificado pelo CIRC Angola")
                except Exception as e:
                    warnings.append(f"Erro na verificação CIRC: {str(e)}")
                    validation_steps.append({
                        "name": "verificacao_circ",
                        "description": "Verificação externa com CIRC Angola",
                        "result": False,
                        "details": f"Erro na comunicação com CIRC: {str(e)}"
                    })
                    if rule["high_risk_without_verification"]:
                        errors.append("Falha na verificação CIRC para documento de alto risco")
            
            # Calcular pontuação de confiança baseada nas validações
            base_validations = [pattern_valid, length_valid, checkdigit_valid, prefix_valid, expiry_valid]
            base_score = sum(1 for v in base_validations if v) / len(base_validations)
            
            # Se houve verificação externa, combinar com a pontuação base
            if "score" in external_verification:
                confidence_score = (base_score * 0.6) + (external_verification["score"] * 0.4)
            else:
                confidence_score = base_score
                
            # O documento é válido se não houver erros e a pontuação de confiança for adequada
            is_valid = len(errors) == 0 and confidence_score >= 0.7
            
            # Adicionar metadados específicos
            metadata.update({
                "validation_score": confidence_score,
                "external_verification": bool(external_verification.get("verified", False)),
                "document_category": self._determine_document_category(doc_type),
                "risk_level": self._calculate_risk_level(doc_type, confidence_score, external_verification)
            })
            
            return DocumentValidationResult(
                is_valid=is_valid,
                confidence_score=confidence_score,
                validation_steps=validation_steps,
                errors=errors,
                warnings=warnings,
                metadata=metadata
            )
            
        except Exception as e:
            logger.error(f"Erro durante validação do documento angolano: {str(e)}")
            return DocumentValidationResult(
                is_valid=False,
                confidence_score=0.0,
                errors=[f"Erro interno durante validação: {str(e)}"],
                metadata={"country": "Angola", "error": True}
            )
    
    def _validate_check_digit(self, doc_number: str, doc_type: str) -> bool:
        """
        Valida o dígito verificador de acordo com as regras específicas do documento.
        
        Args:
            doc_number: Número do documento a validar
            doc_type: Tipo de documento (ex: "nif", "bi")
            
        Returns:
            bool: True se o dígito verificador for válido, False caso contrário
        """
        if doc_type not in self.verification_digit_rules:
            return True  # Não há regra específica para este tipo
        
        if doc_type == "nif":
            # Algoritmo específico para NIF angolano
            # Os primeiros 9 dígitos são multiplicados pelos pesos
            # O resultado da soma é dividido pelo módulo e subtraído
            rule = self.verification_digit_rules["nif"]
            digits = [int(d) for d in doc_number]
            check_digit = digits[-1]  # Último dígito
            
            # Cálculo do dígito verificador
            weighted_sum = sum(d * rule["weights"][i] for i, d in enumerate(digits[:-1]))
            remainder = weighted_sum % rule["modulus"]
            expected_digit = (rule["modulus"] - remainder) % rule["modulus"]
            
            return check_digit == expected_digit
        
        return True
    
    def _determine_document_category(self, doc_type: str) -> str:
        """
        Determina a categoria do documento baseada em seu tipo.
        
        Args:
            doc_type: Tipo de documento
            
        Returns:
            str: Categoria do documento (identidade, fiscal, etc.)
        """
        categories = {
            "bi": "identidade_primaria",
            "passport": "viagem",
            "nif": "fiscal",
            "military_id": "servico",
            "voter_card": "eleitoral"
        }
        
        return categories.get(doc_type, "outro")
    
    def _calculate_risk_level(self, doc_type: str, confidence_score: float, 
                              external_verification: Dict[str, Any]) -> str:
        """
        Calcula o nível de risco associado à validação do documento.
        
        Args:
            doc_type: Tipo de documento
            confidence_score: Pontuação de confiança
            external_verification: Resultados da verificação externa
            
        Returns:
            str: Nível de risco (baixo, medio, alto)
        """
        # Documentos considerados de alto risco para Angola
        high_risk_docs = ["bi", "passport"]
        
        # Se a verificação externa falhou para um documento de alto risco
        if doc_type in high_risk_docs and not external_verification.get("verified", False):
            return "alto"
        
        # Baseado na pontuação de confiança
        if confidence_score >= 0.9:
            return "baixo"
        elif confidence_score >= 0.7:
            return "medio"
        else:
            return "alto"


# Função utilitária para teste rápido
def test_validator():
    """Função para testes simples do validador"""
    validator = AngolaDocumentValidator()
    
    # Teste com BI angolano (fictício)
    test_bi = {
        "type": "bi",
        "number": "000123456AB123",
        "issue_date": "2020-01-15",
        "expiry_date": "2025-01-15",
        "issuer": "Governo de Angola",
        "holder_name": "João Manuel dos Santos",
        "holder_birth_date": "1985-07-23"
    }
    
    result = validator.validate(test_bi)
    
    print("Resultado da validação:")
    print(f"Válido: {result.is_valid}")
    print(f"Pontuação de confiança: {result.confidence_score:.2f}")
    print("Erros:", result.errors)
    print("Avisos:", result.warnings)
    print("Metadados:", result.metadata)
    print("\nEtapas de validação:")
    for step in result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")


if __name__ == "__main__":
    # Configurar logging para testes
    logging.basicConfig(level=logging.INFO)
    
    # Executar teste
    test_validator()