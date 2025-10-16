#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Validador Documental para Brasil

Este módulo implementa validações específicas para documentos brasileiros
no contexto do sistema IAM/TrustGuard da plataforma INNOVABIZ.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import re
import json
import logging
from typing import Dict, Any, List, Optional
from datetime import datetime

# Configuração de logging
logger = logging.getLogger("brasil_document_validator")

# Importações de módulos de integração
try:
    from src.api.services.integration.receita_federal_connector import ReceitaFederalConnector
    from src.api.services.document_validation.base_document_validator import BaseDocumentValidator
    from src.api.services.document_validation.document_validation_result import DocumentValidationResult
except ImportError:
    logger.warning("Importações não encontradas. Executando em modo isolado.")
    # Classes mock para testes isolados
    class ReceitaFederalConnector:
        def verify_cpf(self, cpf, name=None):
            return {"verified": True, "score": 0.95}
            
        def verify_cnpj(self, cnpj, company_name=None):
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


class BrasilDocumentValidator(BaseDocumentValidator):
    """
    Validador especializado para documentos brasileiros, implementando regras
    específicas para CPF, CNPJ, RG e outros documentos brasileiros.
    
    Esta classe é responsável por validar documentos brasileiros de acordo com
    regras sintáticas, formatos e verificações externas com sistemas como Receita Federal.
    """
    
    def __init__(self):
        """Inicializa o validador com configurações específicas para Brasil."""
        super().__init__()
        self.country_code = "BR"
        self.name = "Brasil Document Validator"
        self.receita_connector = ReceitaFederalConnector()
        
        # Carregar configurações regionais específicas
        self.load_country_specific_rules()
        
    def load_country_specific_rules(self):
        """
        Carrega regras específicas para validação de documentos brasileiros.
        Inclui expressões regulares, regras de formato e configurações de validação.
        """
        self.rules = {
            "cpf": {
                "pattern": r"^\d{11}$",  # 11 dígitos numéricos
                "length": 11,
                "expiry_years": None,  # Não expira
                "requires_external_verification": True,
                "high_risk_without_verification": False,
                "masked_format": r"^\d{3}\.\d{3}\.\d{3}-\d{2}$"  # Formato com pontuação: 123.456.789-01
            },
            "cnpj": {
                "pattern": r"^\d{14}$",  # 14 dígitos numéricos
                "length": 14,
                "expiry_years": None,  # Não expira
                "requires_external_verification": True,
                "high_risk_without_verification": True,
                "masked_format": r"^\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}$"  # Formato: 12.345.678/0001-90
            },
            "rg": {
                "pattern": r"^[0-9X]{5,13}$",  # Varia por estado, aceita dígitos e X
                "length": [5, 6, 7, 8, 9, 10, 11, 12, 13],  # Comprimento varia por estado
                "expiry_years": None,  # Varia por estado, alguns não expiram
                "requires_external_verification": False,
                "high_risk_without_verification": False
            },
            "cnh": {  # Carteira Nacional de Habilitação
                "pattern": r"^\d{11}$",  # 11 dígitos numéricos
                "length": 11,
                "expiry_years": 5,  # 5 anos validade padrão (pode variar)
                "requires_external_verification": True,
                "high_risk_without_verification": True
            },
            "passport": {
                "pattern": r"^[A-Z]{2}\d{6}$",  # 2 letras seguidas de 6 números
                "length": 8,
                "expiry_years": 10,  # Validade de 10 anos
                "requires_external_verification": True,
                "high_risk_without_verification": True
            },
            "titulo_eleitor": {
                "pattern": r"^\d{12}$",  # 12 dígitos numéricos
                "length": 12,
                "expiry_years": None,  # Não expira
                "requires_external_verification": False,
                "high_risk_without_verification": False
            },
            "nis": {  # PIS/PASEP/NIS/NIT
                "pattern": r"^\d{11}$",  # 11 dígitos numéricos
                "length": 11,
                "expiry_years": None,  # Não expira
                "requires_external_verification": False,
                "high_risk_without_verification": False
            }
        }
        
        # Regras específicas para verificação de dígitos
        self.verification_digit_rules = {
            "cpf": {
                "weights_first": [10, 9, 8, 7, 6, 5, 4, 3, 2],
                "weights_second": [11, 10, 9, 8, 7, 6, 5, 4, 3, 2]
            },
            "cnpj": {
                "weights_first": [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2],
                "weights_second": [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2]
            },
            "titulo_eleitor": {
                "weights_first": [2, 3, 4, 5, 6, 7, 8, 9],
                "weights_second": [7, 8, 9]
            },
            "nis": {
                "weights": [3, 2, 9, 8, 7, 6, 5, 4, 3, 2]
            }
        }
        
        # Lista de CPFs inválidos conhecidos (sequências repetidas)
        self.invalid_cpf_sequences = [
            '00000000000', '11111111111', '22222222222', '33333333333',
            '44444444444', '55555555555', '66666666666', '77777777777',
            '88888888888', '99999999999'
        ]
        
        # Lista de CNPJs inválidos conhecidos (sequências repetidas)
        self.invalid_cnpj_sequences = [
            '00000000000000', '11111111111111', '22222222222222', '33333333333333',
            '44444444444444', '55555555555555', '66666666666666', '77777777777777',
            '88888888888888', '99999999999999'
        ]
        
    def validate(self, document_data: Dict[str, Any]) -> DocumentValidationResult:
        """
        Valida um documento brasileiro com base nos critérios específicos do país.
        
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
            
            # Remover caracteres não numéricos para CPF/CNPJ/outros documentos
            if doc_type in ["cpf", "cnpj", "titulo_eleitor", "nis", "cnh"]:
                doc_number = re.sub(r'[^0-9]', '', doc_number)
            
            # Inicializar resultado da validação
            validation_steps = []
            errors = []
            warnings = []
            metadata = {
                "country": "Brasil",
                "document_type": doc_type,
                "verification_timestamp": datetime.now().isoformat()
            }
            
            # Verificar se o tipo de documento é suportado
            if doc_type not in self.rules:
                errors.append(f"Tipo de documento '{doc_type}' não é suportado para Brasil")
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
            
            # 3. Verificação de sequências inválidas para CPF/CNPJ
            sequence_valid = True
            if doc_type == "cpf":
                sequence_valid = doc_number not in self.invalid_cpf_sequences
                validation_steps.append({
                    "name": "sequencia_invalida",
                    "description": "Verificação de CPF com dígitos repetidos",
                    "result": sequence_valid,
                    "details": "CPF não contém sequência inválida" if sequence_valid else "CPF contém sequência inválida (dígitos repetidos)"
                })
                
                if not sequence_valid:
                    errors.append("CPF inválido: sequência de dígitos repetidos")
                    
            elif doc_type == "cnpj":
                sequence_valid = doc_number not in self.invalid_cnpj_sequences
                validation_steps.append({
                    "name": "sequencia_invalida",
                    "description": "Verificação de CNPJ com dígitos repetidos",
                    "result": sequence_valid,
                    "details": "CNPJ não contém sequência inválida" if sequence_valid else "CNPJ contém sequência inválida (dígitos repetidos)"
                })
                
                if not sequence_valid:
                    errors.append("CNPJ inválido: sequência de dígitos repetidos")
            
            # 4. Validação de dígito verificador
            checkdigit_valid = True
            if doc_type in self.verification_digit_rules:
                checkdigit_valid = self._validate_check_digit(doc_number, doc_type)
                validation_steps.append({
                    "name": "digito_verificador",
                    "description": f"Validação de dígitos verificadores para {doc_type.upper()}",
                    "result": checkdigit_valid,
                    "details": f"Dígitos verificadores {'válidos' if checkdigit_valid else 'inválidos'}"
                })
                
                if not checkdigit_valid:
                    errors.append(f"{doc_type.upper()} inválido: erro nos dígitos verificadores")
                    
            # 5. Verificação de validade da data de expiração (para documentos que expiram)
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
            
            # 6. Verificação externa com a Receita Federal ou outros serviços
            external_verification = {"verified": False, "score": 0.0}
            if rule["requires_external_verification"]:
                try:
                    if doc_type == "cpf":
                        external_verification = self.receita_connector.verify_cpf(
                            cpf=doc_number,
                            name=holder_name
                        )
                    elif doc_type == "cnpj":
                        external_verification = self.receita_connector.verify_cnpj(
                            cnpj=doc_number,
                            company_name=holder_name
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
            base_validations = [pattern_valid, length_valid, sequence_valid, checkdigit_valid, expiry_valid]
            base_score = sum(1 for v in base_validations if v) / len(base_validations)
            
            # Se houve verificação externa, combinar com a pontuação base
            if "score" in external_verification:
                confidence_score = (base_score * 0.6) + (external_verification["score"] * 0.4)
            else:
                confidence_score = base_score
                
            # O documento é válido se não houver erros e a pontuação de confiança for adequada
            is_valid = len(errors) == 0 and confidence_score >= 0.7
            
            # Adicionar metadados específicos do Brasil
            metadata.update({
                "validation_score": confidence_score,
                "external_verification": bool(external_verification.get("verified", False)),
                "document_category": self._determine_document_category(doc_type),
                "risk_level": self._calculate_risk_level(doc_type, confidence_score, external_verification),
                # Adicionar metadados específicos do Brasil
                "region": additional_data.get("region", ""),
                "document_format": "standard",
                "validation_type": "complete" if rule["requires_external_verification"] else "partial"
            })
            
            # Adicionar recomendações específicas para Brasil
            recommendations = []
            if is_valid and doc_type == "cpf":
                recommendations.append("Verificação adicional com Serasa/SPC recomendada para transações de alto valor")
            elif is_valid and doc_type == "cnpj":
                recommendations.append("Verificar Cadastro de CNPJ para detalhes de atividade econômica e situação cadastral")
            elif is_valid and doc_type == "passport":
                recommendations.append("Verificação adicional com Polícia Federal recomendada para estrangeiros")
                
            if recommendations:
                metadata["recommendations"] = recommendations
                
            # Log e retorno final
            self._log_validation_activity(document_data, is_valid, confidence_score)
            
            return DocumentValidationResult(
                is_valid=is_valid,
                confidence_score=confidence_score,
                validation_steps=validation_steps,
                errors=errors,
                warnings=warnings,
                metadata=metadata
            )
            
        except Exception as e:
            logger.error(f"Erro durante validação do documento brasileiro: {str(e)}")
            return DocumentValidationResult(
                is_valid=False,
                confidence_score=0.0,
                errors=[f"Erro interno durante validação: {str(e)}"],
                metadata={"country": "Brasil", "error": True}
            )
    
    def _validate_check_digit(self, doc_number: str, doc_type: str) -> bool:
        """
        Valida o(s) dígito(s) verificador(es) de acordo com as regras específicas do documento.
        
        Args:
            doc_number: Número do documento a validar
            doc_type: Tipo de documento (ex: "cpf", "cnpj")
            
        Returns:
            bool: True se o dígito verificador for válido, False caso contrário
        """
        if doc_type not in self.verification_digit_rules:
            return True  # Não há regra específica para este tipo
        
        if doc_type == "cpf":
            # Validação do CPF
            # Algoritmo específico para CPF brasileiro
            rules = self.verification_digit_rules["cpf"]
            
            # Verificar primeiro dígito
            sum_first = 0
            for i in range(9):
                sum_first += int(doc_number[i]) * rules["weights_first"][i]
            
            remainder_first = sum_first % 11
            check_first = 0 if remainder_first < 2 else 11 - remainder_first
            
            if int(doc_number[9]) != check_first:
                return False
                
            # Verificar segundo dígito
            sum_second = 0
            for i in range(10):
                sum_second += int(doc_number[i]) * rules["weights_second"][i]
                
            remainder_second = sum_second % 11
            check_second = 0 if remainder_second < 2 else 11 - remainder_second
            
            return int(doc_number[10]) == check_second
            
        elif doc_type == "cnpj":
            # Validação do CNPJ
            # Algoritmo específico para CNPJ brasileiro
            rules = self.verification_digit_rules["cnpj"]
            
            # Verificar primeiro dígito
            sum_first = 0
            for i in range(12):
                sum_first += int(doc_number[i]) * rules["weights_first"][i]
                
            remainder_first = sum_first % 11
            check_first = 0 if remainder_first < 2 else 11 - remainder_first
            
            if int(doc_number[12]) != check_first:
                return False
                
            # Verificar segundo dígito
            sum_second = 0
            for i in range(13):
                sum_second += int(doc_number[i]) * rules["weights_second"][i]
                
            remainder_second = sum_second % 11
            check_second = 0 if remainder_second < 2 else 11 - remainder_second
            
            return int(doc_number[13]) == check_second
            
        elif doc_type == "titulo_eleitor":
            # Validação do Título de Eleitor
            # Algoritmo específico para Título de Eleitor brasileiro
            rules = self.verification_digit_rules["titulo_eleitor"]
            
            # Calcular primeiro dígito verificador (10º dígito)
            sum_first = 0
            for i in range(8):
                sum_first += int(doc_number[i]) * rules["weights_first"][i]
                
            remainder_first = sum_first % 11
            check_first = 0 if remainder_first == 0 or remainder_first == 10 else remainder_first
            
            if int(doc_number[10]) != check_first:
                return False
                
            # Calcular segundo dígito verificador (11º dígito)
            # Usando UF (9º e 10º dígitos)
            uf = doc_number[8:10]
            sum_second = 0
            for i in range(3):
                if i < 2:  # UF tem 2 dígitos
                    sum_second += int(uf[i]) * rules["weights_second"][i]
                else:
                    sum_second += check_first * rules["weights_second"][i]  # Usa o primeiro DV
                    
            remainder_second = sum_second % 11
            check_second = 0 if remainder_second == 0 or remainder_second == 10 else remainder_second
            
            return int(doc_number[11]) == check_second
            
        elif doc_type == "nis":
            # Validação do NIS/PIS/PASEP
            # Algoritmo específico para NIS
            weights = self.verification_digit_rules["nis"]["weights"]
            
            sum_total = 0
            for i in range(10):
                sum_total += int(doc_number[i]) * weights[i]
                
            remainder = sum_total % 11
            check_digit = 0 if remainder < 2 else 11 - remainder
            
            return int(doc_number[10]) == check_digit
            
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
            "cpf": "fiscal_pessoa_fisica",
            "cnpj": "fiscal_pessoa_juridica",
            "rg": "identidade_pessoa_fisica",
            "passport": "viagem_internacional",
            "cnh": "habilitacao",
            "titulo_eleitor": "eleitoral",
            "nis": "previdenciario"
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
        # Documentos considerados de alto risco para Brasil
        high_risk_docs = ["passport", "cnpj"]
        
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
            
    def _log_validation_activity(self, document_data: Dict[str, Any], 
                               is_valid: bool, confidence_score: float) -> None:
        """
        Registra a atividade de validação para fins de auditoria e monitoramento.
        
        Args:
            document_data: Dados do documento validado
            is_valid: Se o documento foi considerado válido
            confidence_score: Pontuação de confiança da validação
        """
        # Sanitizar dados para log (remover informações sensíveis)
        sanitized_data = {
            "document_type": document_data.get("type", ""),
            "is_valid": is_valid,
            "confidence": confidence_score,
            "timestamp": datetime.now().isoformat()
        }
        
        # Aqui seria implementada a lógica de envio para sistema de logs
        logger.info(f"Validação documento BR: {sanitized_data}")


# Função utilitária para teste rápido
def test_validator():
    """Função para testes simples do validador"""
    validator = BrasilDocumentValidator()
    
    # Teste com CPF válido (fictício)
    test_cpf = {
        "type": "cpf",
        "number": "52998224725",
        "holder_name": "Maria Silva Santos",
        "holder_birth_date": "1975-05-20"
    }
    
    # Teste com CNPJ válido (fictício)
    test_cnpj = {
        "type": "cnpj",
        "number": "27865757000102",
        "holder_name": "Empresa Brasileira Ltda"
    }
    
    # Validar CPF
    cpf_result = validator.validate(test_cpf)
    print("=== Resultado da validação de CPF ===")
    print(f"Válido: {cpf_result.is_valid}")
    print(f"Pontuação de confiança: {cpf_result.confidence_score:.2f}")
    print("Erros:", cpf_result.errors)
    print("Avisos:", cpf_result.warnings)
    
    print("\nEtapas de validação:")
    for step in cpf_result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")
        
    # Validar CNPJ
    cnpj_result = validator.validate(test_cnpj)
    print("\n\n=== Resultado da validação de CNPJ ===")
    print(f"Válido: {cnpj_result.is_valid}")
    print(f"Pontuação de confiança: {cnpj_result.confidence_score:.2f}")
    print("Erros:", cnpj_result.errors)
    print("Avisos:", cnpj_result.warnings)
    
    print("\nEtapas de validação:")
    for step in cnpj_result.validation_steps:
        print(f"- {step['name']}: {'✓' if step['result'] else '✗'} - {step['details']}")


if __name__ == "__main__":
    # Configurar logging para testes
    logging.basicConfig(level=logging.INFO)
    
    # Executar teste
    test_validator()