#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Validador Documental para Moçambique

Este módulo implementa validações específicas para documentos moçambicanos
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
logger = logging.getLogger("mocambique_document_validator")

# Importações de módulos de integração
try:
    from src.api.services.integration.banco_mocambique_connector import BancoMocambiqueConnector
    from src.api.services.document_validation.base_document_validator import BaseDocumentValidator
    from src.api.services.document_validation.document_validation_result import DocumentValidationResult
except ImportError:
    logger.warning("Importações não encontradas. Executando em modo isolado.")
    # Classes mock para testes isolados
    class BancoMocambiqueConnector:
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


class MocambiqueDocumentValidator(BaseDocumentValidator):
    """
    Validador especializado para documentos moçambicanos, implementando regras
    específicas para BI, NUIT, Passaporte e outros documentos moçambicanos.
    
    Esta classe é responsável por validar documentos moçambicanos de acordo com
    regras sintáticas, formatos e verificações externas com sistemas como Banco de Moçambique.
    """
    
    def __init__(self):
        """Inicializa o validador com configurações específicas para Moçambique."""
        super().__init__()
        self.country_code = "MZ"
        self.name = "Moçambique Document Validator"
        self.banco_mz_connector = BancoMocambiqueConnector()
        
        # Carregar configurações regionais específicas
        self.load_country_specific_rules()
        
    def load_country_specific_rules(self):
        """
        Carrega regras específicas para validação de documentos moçambicanos.
        Inclui expressões regulares, regras de formato e configurações de validação.
        """
        self.rules = {
            "bi": {
                "pattern": r"^\d{10}[A-Z]\d$",  # Formato típico do BI moçambicano
                "length": 12,
                "expiry_years": 10,
                "requires_external_verification": True,
                "high_risk_without_verification": True,
            },
            "nuit": {  # Número Único de Identificação Tributária
                "pattern": r"^\d{9}$",
                "length": 9,
                "expiry_years": None,  # Não expira
                "requires_external_verification": True,
                "high_risk_without_verification": False,
            },
            "passport": {
                "pattern": r"^[A-Z]{1,2}\d{7}$",  # Formato típico do passaporte moçambicano
                "length": [8, 9],
                "expiry_years": 10,
                "requires_external_verification": True,
                "high_risk_without_verification": True,
            },
            "dire": {  # Documento de Identificação de Residentes Estrangeiros
                "pattern": r"^[A-Z]\d{9}$",
                "length": 10,
                "expiry_years": 5,  # Validade padrão
                "requires_external_verification": True,
                "high_risk_without_verification": True,
            },
            "voter_card": {  # Cartão de Eleitor
                "pattern": r"^\d{8,12}$",  # Varia por região
                "length": [8, 9, 10, 11, 12],  # Diferentes formatos por região
                "expiry_years": None,  # Não expira
                "requires_external_verification": False,
                "high_risk_without_verification": False,
            }
        }
        
        # Regras para verificação de dígitos
        self.verification_digit_rules = {
            "nuit": {
                "modulus": 11,
                "weights": [2, 3, 4, 5, 6, 7, 8, 9]
            },
            "bi": {
                "modulus": 10,
                "weights": [2, 3, 4, 5, 6, 7, 8, 9, 10]
            }
        }
        
        # Mapeamento de códigos de província para o BI moçambicano
        self.province_codes = {
            "01": "Maputo Cidade",
            "02": "Maputo Província",
            "03": "Gaza",
            "04": "Inhambane",
            "05": "Sofala",
            "06": "Manica",
            "07": "Tete",
            "08": "Zambézia",
            "09": "Nampula",
            "10": "Niassa",
            "11": "Cabo Delgado"
        }
        
    def validate(self, document_data: Dict[str, Any]) -> DocumentValidationResult:
        """
        Valida um documento moçambicano com base nos critérios específicos do país.
        
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
            
            # Remover caracteres não alfanuméricos para documentos NUIT
            if doc_type == "nuit":
                doc_number = re.sub(r'[^0-9]', '', doc_number)
            
            # Inicializar resultado da validação
            validation_steps = []
            errors = []
            warnings = []
            metadata = {
                "country": "Moçambique",
                "document_type": doc_type,
                "verification_timestamp": datetime.now().isoformat()
            }
            
            # Verificar se o tipo de documento é suportado
            if doc_type not in self.rules:
                errors.append(f"Tipo de documento '{doc_type}' não é suportado para Moçambique")
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
            
            # 3. Validação de código de província para BI
            province_valid = True
            if doc_type == "bi" and len(doc_number) >= 12:
                province_code = doc_number[:2]
                province_valid = province_code in self.province_codes
                
                validation_steps.append({
                    "name": "codigo_provincia",
                    "description": "Validação do código de província no BI",
                    "result": province_valid,
                    "details": f"Código de província {'válido' if province_valid else 'inválido'}: {province_code}"
                })
                
                if not province_valid:
                    warnings.append(f"Código de província não reconhecido: {province_code}")
                else:
                    metadata["province"] = self.province_codes[province_code]
            
            # 4. Validação de dígito verificador (se aplicável)
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
            
            # 6. Verificação externa com Banco de Moçambique (se disponível e necessário)
            external_verification = {"verified": False, "score": 0.0}
            if rule["requires_external_verification"]:
                try:
                    external_verification = self.banco_mz_connector.verify_document(
                        doc_type=doc_type,
                        doc_number=doc_number,
                        name=holder_name
                    )
                    
                    external_valid = external_verification.get("verified", False)
                    external_score = external_verification.get("score", 0.0)
                    
                    validation_steps.append({
                        "name": "verificacao_banco_mz",
                        "description": "Verificação externa com Banco de Moçambique",
                        "result": external_valid,
                        "details": f"Verificação {'bem-sucedida' if external_valid else 'falhou'} com score {external_score:.2f}"
                    })
                    
                    if not external_valid:
                        errors.append(f"Documento não verificado pelo Banco de Moçambique: {external_verification.get('message', '')}")
                except Exception as e:
                    warnings.append(f"Erro na verificação com Banco de Moçambique: {str(e)}")
                    validation_steps.append({
                        "name": "verificacao_banco_mz",
                        "description": "Verificação externa com Banco de Moçambique",
                        "result": False,
                        "details": f"Erro na comunicação: {str(e)}"
                    })
                    if rule["high_risk_without_verification"]:
                        errors.append("Falha na verificação externa para documento de alto risco")
            
            # Calcular pontuação de confiança baseada nas validações
            base_validations = [pattern_valid, length_valid, province_valid, checkdigit_valid, expiry_valid]
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
            
            # Adicionar recomendações específicas para o contexto moçambicano
            if doc_type == "nuit" and is_valid:
                metadata["recommendations"] = [
                    "Recomenda-se a verificação adicional da Certidão de Quitação fiscal para transações de alto valor"
                ]
            elif doc_type == "bi" and is_valid:
                metadata["recommendations"] = [
                    "Para transações financeiras acima de 50.000 MZN, recomenda-se validação adicional"
                ]
                
            return DocumentValidationResult(
                is_valid=is_valid,
                confidence_score=confidence_score,
                validation_steps=validation_steps,
                errors=errors,
                warnings=warnings,
                metadata=metadata
            )
            
        except Exception as e:
            logger.error(f"Erro durante validação do documento moçambicano: {str(e)}")
            return DocumentValidationResult(
                is_valid=False,
                confidence_score=0.0,
                errors=[f"Erro interno durante validação: {str(e)}"],
                metadata={"country": "Moçambique", "error": True}
            )
    
    def _validate_check_digit(self, doc_number: str, doc_type: str) -> bool:
        """
        Valida o dígito verificador de acordo com as regras específicas do documento.
        
        Args:
            doc_number: Número do documento a validar
            doc_type: Tipo de documento (ex: "nuit", "bi")
            
        Returns:
            bool: True se o dígito verificador for válido, False caso contrário
        """
        if doc_type not in self.verification_digit_rules:
            return True  # Não há regra específica para este tipo
        
        if doc_type == "nuit":
            # Algoritmo de validação do NUIT moçambicano
            rule = self.verification_digit_rules["nuit"]
            digits = [int(d) for d in doc_number]
            check_digit = digits[-1]  # Último dígito
            
            # Cálculo do dígito verificador
            weighted_sum = sum(d * rule["weights"][i] for i, d in enumerate(digits[:-1]))
            remainder = weighted_sum % rule["modulus"]
            expected_digit = (rule["modulus"] - remainder) % rule["modulus"]
            
            # Validação especial: se o resto for 1, o dígito verificador deve ser 0
            if remainder == 1:
                expected_digit = 0
                
            return check_digit == expected_digit
            
        elif doc_type == "bi":
            # Algoritmo de validação do BI moçambicano
            rule = self.verification_digit_rules["bi"]
            
            # Para o BI moçambicano, o penúltimo caractere é uma letra de controle
            # Convertemos a letra para um valor numérico (A=0, B=1, etc.)
            control_char = doc_number[-2]
            if not control_char.isalpha():
                return False
                
            control_value = ord(control_char) - ord('A')
            
            # Cálculo baseado nos primeiros 10 dígitos
            numeric_part = doc_number[:10]
            if not numeric_part.isdigit():
                return False
                
            digits = [int(d) for d in numeric_part]
            
            # Cálculo do valor de controle esperado
            weighted_sum = sum(d * rule["weights"][i] for i, d in enumerate(digits))
            expected_control = weighted_sum % 26  # 26 letras no alfabeto
            
            # Validar a letra de controle
            return control_value == expected_control
            
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
            "bi": "identidade_nacional",
            "nuit": "identificacao_fiscal",
            "passport": "viagem_internacional",
            "dire": "residencia_estrangeiro",
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
        # Documentos considerados de alto risco para Moçambique
        high_risk_docs = ["passport", "dire"]
        medium_risk_docs = ["bi"]
        
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
    validator = MocambiqueDocumentValidator()
    
    # Teste com NUIT moçambicano (fictício)
    test_nuit = {
        "type": "nuit",
        "number": "123456789",
        "holder_name": "António José Machava",
        "holder_birth_date": "1982-04-15"
    }
    
    # Teste com BI moçambicano (fictício)
    test_bi = {
        "type": "bi",
        "number": "0512345678A1",
        "issue_date": "2018-07-20",
        "expiry_date": "2028-07-20",
        "issuer": "Direção de Identificação Civil de Moçambique",
        "holder_name": "Maria Luísa Tembe",
        "holder_birth_date": "1990-10-25"
    }
    
    # Validar NUIT
    nuit_result = validator.validate(test_nuit)
    print("=== Resultado da validação de NUIT ===")
    print(f"Válido: {nuit_result.is_valid}")
    print(f"Pontuação de confiança: {nuit_result.confidence_score:.2f}")
    print("Erros:", nuit_result.errors)
    print("Avisos:", nuit_result.warnings)
    
    print("\nEtapas de validação:")
    for step in nuit_result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")
    
    # Validar BI
    bi_result = validator.validate(test_bi)
    print("\n\n=== Resultado da validação de BI ===")
    print(f"Válido: {bi_result.is_valid}")
    print(f"Pontuação de confiança: {bi_result.confidence_score:.2f}")
    print("Erros:", bi_result.errors)
    print("Avisos:", bi_result.warnings)
    
    print("\nEtapas de validação:")
    for step in bi_result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")


if __name__ == "__main__":
    # Configurar logging para testes
    logging.basicConfig(level=logging.INFO)
    
    # Executar teste
    test_validator()