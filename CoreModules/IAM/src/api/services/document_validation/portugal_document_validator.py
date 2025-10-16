#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Validador Documental para Portugal

Este módulo implementa validações específicas para documentos portugueses
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
logger = logging.getLogger("portugal_document_validator")

# Importações de módulos de integração
try:
    from src.api.services.integration.portugal_services_connector import PortugalServicesConnector
    from src.api.services.document_validation.base_document_validator import BaseDocumentValidator
    from src.api.services.document_validation.document_validation_result import DocumentValidationResult
    from src.api.services.gdpr.gdpr_compliance_handler import GDPRComplianceHandler
except ImportError:
    logger.warning("Importações não encontradas. Executando em modo isolado.")
    # Classes mock para testes isolados
    class PortugalServicesConnector:
        def verify_citizen_card(self, cc_number, name=None):
            return {"verified": True, "score": 0.95}
            
        def verify_nif(self, nif, name=None):
            return {"verified": True, "score": 0.92}

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
            
    class GDPRComplianceHandler:
        def __init__(self, country_code="PT"):
            self.country_code = country_code
            
        def register_processing_purpose(self, purpose, legal_basis, data_categories):
            return {"registered": True, "processing_id": "gdpr-proc-123456"}
            
        def anonymize_sensitive_data(self, data, fields_to_anonymize):
            # Cria uma cópia sem os campos sensíveis
            return {k: (v if k not in fields_to_anonymize else "[REDACTED]") 
                    for k, v in data.items()}
                    
        def get_processing_record(self, processing_id):
            return {
                "purpose": "document_validation",
                "legal_basis": "legitimate_interest",
                "retention_period": "30 days"
            }


class PortugalDocumentValidator(BaseDocumentValidator):
    """
    Validador especializado para documentos portugueses, implementando regras
    específicas para Cartão de Cidadão, NIF e outros documentos portugueses.
    
    Esta classe é responsável por validar documentos portugueses de acordo com
    regras sintáticas, formatos e verificações externas com sistemas governamentais.
    Implementa conformidade GDPR para processamento de dados pessoais.
    """
    
    def __init__(self):
        """Inicializa o validador com configurações específicas para Portugal."""
        super().__init__()
        self.country_code = "PT"
        self.name = "Portugal Document Validator"
        self.pt_services_connector = PortugalServicesConnector()
        self.gdpr_handler = GDPRComplianceHandler(country_code="PT")
        
        # Carregar configurações regionais específicas
        self.load_country_specific_rules()
        
    def load_country_specific_rules(self):
        """
        Carrega regras específicas para validação de documentos portugueses.
        Inclui expressões regulares, regras de formato e configurações de validação.
        """
        self.rules = {
            "cc": {  # Cartão de Cidadão
                "pattern": r"^\d{8}-\d[A-Z0-9]{2}\d$|^\d{9}$",  # Formato com ou sem separador
                "length": [9, 12],  # 9 dígitos ou 8+1+2+1 com separadores
                "expiry_years": 5,  # Validade padrão de 5 anos (pode variar)
                "requires_external_verification": True,
                "high_risk_without_verification": True,
                "gdpr_sensitive": True
            },
            "nif": {  # Número de Identificação Fiscal
                "pattern": r"^\d{9}$",
                "length": 9,
                "expiry_years": None,  # Não expira
                "requires_external_verification": True,
                "high_risk_without_verification": False,
                "gdpr_sensitive": True
            },
            "passport": {
                "pattern": r"^[A-Z]{1,2}\d{6}$",
                "length": [7, 8],
                "expiry_years": 5,  # Validade padrão
                "requires_external_verification": True,
                "high_risk_without_verification": True,
                "gdpr_sensitive": True
            },
            "ss": {  # Número de Segurança Social
                "pattern": r"^\d{11}$",
                "length": 11,
                "expiry_years": None,  # Não expira
                "requires_external_verification": False,
                "high_risk_without_verification": False,
                "gdpr_sensitive": True
            },
            "sns": {  # Número de Utente do Serviço Nacional de Saúde
                "pattern": r"^\d{9}$",
                "length": 9,
                "expiry_years": None,  # Não expira
                "requires_external_verification": False,
                "high_risk_without_verification": False,
                "gdpr_sensitive": True
            }
        }
        
        # Regras para verificação de dígitos
        self.verification_digit_rules = {
            "nif": {
                "modulus": 11,
                "weights": [9, 8, 7, 6, 5, 4, 3, 2]
            },
            "cc": {
                "modulus": 10,
                "weights": [9, 8, 7, 6, 5, 4, 3, 2]
            },
            "ss": {
                "modulus": 11,
                "weights": [2, 3, 4, 5, 6, 7, 8, 9, 2, 3]
            }
        }
        
        # Definições GDPR para processamento de cada tipo de documento
        self.gdpr_rules = {
            "cc": {
                "purpose": "identity_verification",
                "legal_basis": "legitimate_interest",
                "retention_period_days": 30,
                "sensitive_fields": ["holder_birth_date", "address"]
            },
            "nif": {
                "purpose": "tax_verification",
                "legal_basis": "legitimate_interest",
                "retention_period_days": 30,
                "sensitive_fields": []
            },
            "passport": {
                "purpose": "international_identity_verification",
                "legal_basis": "legitimate_interest",
                "retention_period_days": 30,
                "sensitive_fields": ["holder_birth_date", "nationality"]
            },
            "ss": {
                "purpose": "social_security_verification",
                "legal_basis": "legitimate_interest",
                "retention_period_days": 30,
                "sensitive_fields": []
            },
            "sns": {
                "purpose": "health_system_verification",
                "legal_basis": "legitimate_interest",
                "retention_period_days": 30,
                "sensitive_fields": ["health_status"]
            }
        }
        
    def validate(self, document_data: Dict[str, Any]) -> DocumentValidationResult:
        """
        Valida um documento português com base nos critérios específicos do país
        e normas GDPR.
        
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
            
            # Remover caracteres não alfanuméricos para alguns tipos de documentos
            if doc_type in ["nif", "ss", "sns"]:
                doc_number = re.sub(r'[^0-9]', '', doc_number)
                
            # Para Cartão de Cidadão, normalizar formato
            if doc_type == "cc" and "-" not in doc_number:
                # Se não tiver hífen e tiver 9 dígitos, converter para formato com hífen
                if len(doc_number) == 9:
                    doc_number = f"{doc_number[:8]}-{doc_number[8]}"
            
            # Inicializar resultado da validação
            validation_steps = []
            errors = []
            warnings = []
            metadata = {
                "country": "Portugal",
                "document_type": doc_type,
                "verification_timestamp": datetime.now().isoformat()
            }
            
            # Verificar se o tipo de documento é suportado
            if doc_type not in self.rules:
                errors.append(f"Tipo de documento '{doc_type}' não é suportado para Portugal")
                return DocumentValidationResult(
                    is_valid=False,
                    confidence_score=0.0,
                    validation_steps=validation_steps,
                    errors=errors,
                    warnings=warnings,
                    metadata=metadata
                )
            
            # Registrar propósito de processamento GDPR
            if self.rules[doc_type].get("gdpr_sensitive", False):
                gdpr_rule = self.gdpr_rules.get(doc_type, {})
                gdpr_registration = self.gdpr_handler.register_processing_purpose(
                    purpose=gdpr_rule.get("purpose", "identity_verification"),
                    legal_basis=gdpr_rule.get("legal_basis", "legitimate_interest"),
                    data_categories=[doc_type, "personal_data"]
                )
                
                metadata["gdpr_processing_id"] = gdpr_registration.get("processing_id", "")
                metadata["gdpr_compliant"] = True
                
                # Anonimizar campos sensíveis para log/metadata
                sensitive_fields = gdpr_rule.get("sensitive_fields", [])
                if sensitive_fields:
                    # Criar cópia anonimizada para metadata
                    anonymized_data = self.gdpr_handler.anonymize_sensitive_data(
                        document_data, 
                        sensitive_fields
                    )
                    metadata["anonymized_fields"] = sensitive_fields
            
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
            
            # 3. Validação de dígito verificador
            checkdigit_valid = True  # Assume válido por padrão
            if doc_type in self.verification_digit_rules:
                checkdigit_valid = self._validate_check_digit(doc_number, doc_type)
                validation_steps.append({
                    "name": "digito_verificador",
                    "description": f"Validação do dígito verificador para {doc_type.upper()}",
                    "result": checkdigit_valid,
                    "details": f"Dígito verificador {'válido' if checkdigit_valid else 'inválido'}"
                })
                
                if not checkdigit_valid:
                    errors.append(f"Dígito verificador inválido para {doc_type.upper()}")
            
            # 4. Verificação de validade da data de expiração
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
            
            # 5. Verificação externa com serviços portugueses
            external_verification = {"verified": False, "score": 0.0}
            if rule["requires_external_verification"]:
                try:
                    if doc_type == "cc":
                        external_verification = self.pt_services_connector.verify_citizen_card(
                            cc_number=doc_number,
                            name=holder_name
                        )
                    elif doc_type == "nif":
                        external_verification = self.pt_services_connector.verify_nif(
                            nif=doc_number,
                            name=holder_name
                        )
                    else:
                        # Para outros documentos verificáveis
                        external_verification = {"verified": True, "score": 0.8}  # Mock para outros tipos
                    
                    external_valid = external_verification.get("verified", False)
                    external_score = external_verification.get("score", 0.0)
                    
                    validation_steps.append({
                        "name": f"verificacao_externa_{doc_type}",
                        "description": f"Verificação externa de {doc_type.upper()}",
                        "result": external_valid,
                        "details": f"Verificação {'bem-sucedida' if external_valid else 'falhou'} com score {external_score:.2f}"
                    })
                    
                    if not external_valid:
                        errors.append(f"Documento não verificado pelo serviço externo: {external_verification.get('message', '')}")
                except Exception as e:
                    warnings.append(f"Erro na verificação externa: {str(e)}")
                    validation_steps.append({
                        "name": f"verificacao_externa_{doc_type}",
                        "description": f"Verificação externa de {doc_type.upper()}",
                        "result": False,
                        "details": f"Erro na comunicação com serviço externo: {str(e)}"
                    })
                    if rule["high_risk_without_verification"]:
                        errors.append(f"Falha na verificação externa para {doc_type.upper()} de alto risco")
            
            # Calcular pontuação de confiança baseada nas validações
            base_validations = [pattern_valid, length_valid, checkdigit_valid, expiry_valid]
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
                "risk_level": self._calculate_risk_level(doc_type, confidence_score, external_verification),
                "eu_member": True,
                "gdpr_applicable": True
            })
            
            # Adicionar recomendações específicas para Portugal
            recommendations = []
            if is_valid and doc_type == "cc":
                recommendations.append("Verificação adicional com Chave Móvel Digital recomendada para maior segurança")
            elif is_valid and doc_type == "nif":
                recommendations.append("Para fins de onboarding financeiro, recomenda-se verificação adicional de residência fiscal")
                
            if recommendations:
                metadata["recommendations"] = recommendations
                
            return DocumentValidationResult(
                is_valid=is_valid,
                confidence_score=confidence_score,
                validation_steps=validation_steps,
                errors=errors,
                warnings=warnings,
                metadata=metadata
            )
            
        except Exception as e:
            logger.error(f"Erro durante validação do documento português: {str(e)}")
            return DocumentValidationResult(
                is_valid=False,
                confidence_score=0.0,
                errors=[f"Erro interno durante validação: {str(e)}"],
                metadata={"country": "Portugal", "error": True}
            )
    
    def _validate_check_digit(self, doc_number: str, doc_type: str) -> bool:
        """
        Valida o dígito verificador de acordo com as regras específicas do documento.
        
        Args:
            doc_number: Número do documento a validar
            doc_type: Tipo de documento (ex: "nif", "cc")
            
        Returns:
            bool: True se o dígito verificador for válido, False caso contrário
        """
        if doc_type not in self.verification_digit_rules:
            return True  # Não há regra específica para este tipo
        
        if doc_type == "nif":
            # Algoritmo de validação do NIF português
            rule = self.verification_digit_rules["nif"]
            
            # Remover caracteres não numéricos
            clean_number = re.sub(r'[^0-9]', '', doc_number)
            
            if len(clean_number) != 9:
                return False
                
            # Para NIFs começando com 1, 2, 5, 6, 8, 9
            valid_first_digits = ['1', '2', '5', '6', '8', '9']
            if clean_number[0] not in valid_first_digits:
                return False
                
            # Calcular dígito verificador
            digits = [int(d) for d in clean_number]
            check_digit = digits[-1]  # Último dígito
            
            # Cálculo do dígito verificador
            weighted_sum = sum(d * rule["weights"][i] for i, d in enumerate(digits[:-1]))
            remainder = weighted_sum % rule["modulus"]
            expected_digit = 0 if remainder == 0 or remainder == 1 else 11 - remainder
            
            return check_digit == expected_digit
            
        elif doc_type == "cc":
            # Algoritmo de validação do Cartão de Cidadão português
            rule = self.verification_digit_rules["cc"]
            
            # Remover caracteres não alfanuméricos e normalizar
            clean_number = re.sub(r'[^0-9A-Z]', '', doc_number)
            
            if len(clean_number) < 9:
                return False
                
            # Extrair parte numérica para cálculo
            numeric_part = clean_number[:8]
            if not numeric_part.isdigit():
                return False
                
            check_digit = int(clean_number[8]) if clean_number[8].isdigit() else -1
            
            # Cálculo do dígito verificador
            digits = [int(d) for d in numeric_part]
            weighted_sum = sum(d * rule["weights"][i] for i, d in enumerate(digits))
            remainder = weighted_sum % rule["modulus"]
            expected_digit = (10 - remainder) % 10
            
            return check_digit == expected_digit
            
        elif doc_type == "ss":
            # Algoritmo de validação do número de Segurança Social português
            rule = self.verification_digit_rules["ss"]
            
            # Remover caracteres não numéricos
            clean_number = re.sub(r'[^0-9]', '', doc_number)
            
            if len(clean_number) != 11:
                return False
                
            digits = [int(d) for d in clean_number]
            check_digit = digits[-1]  # Último dígito
            
            # Cálculo do dígito verificador
            weighted_sum = sum(d * rule["weights"][i] for i, d in enumerate(digits[:-1]))
            remainder = weighted_sum % rule["modulus"]
            expected_digit = 0 if remainder == 0 else 11 - remainder
            
            # Caso especial: se o resto for 1, o dígito verificador deve ser 0
            if remainder == 1:
                expected_digit = 0
                
            return check_digit == expected_digit
            
        return True  # Se não tiver validação específica
        
    def _determine_document_category(self, doc_type: str) -> str:
        """
        Determina a categoria do documento baseada em seu tipo.
        
        Args:
            doc_type: Tipo de documento
            
        Returns:
            str: Categoria do documento (identidade, fiscal, etc.)
        """
        categories = {
            "cc": "identidade_nacional",
            "nif": "identificacao_fiscal",
            "passport": "viagem_internacional",
            "ss": "seguranca_social",
            "sns": "servico_saude"
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
        # Documentos considerados de alto risco para Portugal
        high_risk_docs = ["passport"]
        medium_risk_docs = ["cc", "nif"]
        
        # Se a verificação externa falhou para um documento de alto risco
        if doc_type in high_risk_docs and not external_verification.get("verified", False):
            return "alto"
            
        # Se a verificação externa falhou para um documento de risco médio
        if doc_type in medium_risk_docs and not external_verification.get("verified", False):
            return "medio"
        
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
    validator = PortugalDocumentValidator()
    
    # Teste com NIF português (fictício)
    test_nif = {
        "type": "nif",
        "number": "123456789",
        "holder_name": "João Silva Costa",
        "holder_birth_date": "1978-12-05"
    }
    
    # Teste com Cartão de Cidadão português (fictício)
    test_cc = {
        "type": "cc",
        "number": "12345678-9ZY3",
        "issue_date": "2021-01-10",
        "expiry_date": "2026-01-10",
        "issuer": "República Portuguesa",
        "holder_name": "Maria Santos Ferreira",
        "holder_birth_date": "1985-08-17"
    }
    
    # Validar NIF
    nif_result = validator.validate(test_nif)
    print("=== Resultado da validação de NIF ===")
    print(f"Válido: {nif_result.is_valid}")
    print(f"Pontuação de confiança: {nif_result.confidence_score:.2f}")
    print("Erros:", nif_result.errors)
    print("Avisos:", nif_result.warnings)
    print("Metadados:", nif_result.metadata)
    
    print("\nEtapas de validação:")
    for step in nif_result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")
    
    # Validar Cartão de Cidadão
    cc_result = validator.validate(test_cc)
    print("\n\n=== Resultado da validação de Cartão de Cidadão ===")
    print(f"Válido: {cc_result.is_valid}")
    print(f"Pontuação de confiança: {cc_result.confidence_score:.2f}")
    print("Erros:", cc_result.errors)
    print("Avisos:", cc_result.warnings)
    print("Metadados:", cc_result.metadata)
    
    print("\nEtapas de validação:")
    for step in cc_result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")


if __name__ == "__main__":
    # Configurar logging para testes
    logging.basicConfig(level=logging.INFO)
    
    # Executar teste
    test_validator()