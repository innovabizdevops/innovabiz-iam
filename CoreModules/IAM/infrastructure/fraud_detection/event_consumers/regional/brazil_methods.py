#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Métodos adicionais para Análise Comportamental Regional do Brasil

Este arquivo contém implementações complementares para o módulo de análise
comportamental específica para o Brasil.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import re
import logging
from typing import Dict, Any, List, Tuple, Optional, Union

# Configuração do logger
logger = logging.getLogger("iam.trustguard.brazil")


# Métodos adicionais para a classe BrazilBehavioralPatterns

def validate_brazil_phone(self, phone_number: str) -> Dict[str, Any]:
    """
    Valida um número de telefone brasileiro.
    
    Args:
        phone_number: Número de telefone a ser validado
        
    Returns:
        Resultado da validação
    """
    # Limpar formatação
    cleaned = re.sub(r'[^\d+]', '', phone_number)
    
    # Verificar se o número contém o código do país
    if cleaned.startswith('+'):
        has_country_code = True
        if not cleaned.startswith('+55'):
            return {
                "valid": False,
                "reason": "country_code_mismatch",
                "details": {
                    "expected_country_code": "+55",
                    "provided_country_code": cleaned[:3]
                }
            }
    else:
        has_country_code = False
    
    # Verificar se é celular (começa com 9 e tem 9 dígitos após DDD)
    mobile_match = self.phone_patterns["mobile"].search(cleaned)
    if mobile_match:
        return {
            "valid": True,
            "type": "mobile",
            "has_country_code": has_country_code,
            "normalized": f"+55{mobile_match.group(2)}"
        }
    
    # Verificar se é fixo (8 dígitos após DDD)
    landline_match = self.phone_patterns["landline"].search(cleaned)
    if landline_match:
        return {
            "valid": True,
            "type": "landline",
            "has_country_code": has_country_code,
            "normalized": f"+55{landline_match.group(2)}"
        }
    
    # Se não corresponder a nenhum padrão válido
    return {
        "valid": False,
        "reason": "invalid_format",
        "details": {
            "expected_formats": [
                "+55 DDD 9XXXXXXXX (celular)",
                "+55 DDD XXXXXXXX (fixo)"
            ]
        }
    }


def analyze_device_context(self, device_data: Dict[str, Any], 
                        user_history: Dict[str, Any]) -> Dict[str, Any]:
    """
    Analisa o contexto do dispositivo em relação ao mercado brasileiro.
    
    Args:
        device_data: Dados do dispositivo
        user_history: Histórico de dispositivos do usuário
        
    Returns:
        Análise do contexto do dispositivo
    """
    # Extrair dados relevantes
    device_type = device_data.get("type", "unknown")
    os_name = device_data.get("os", {}).get("name", "unknown")
    os_version = device_data.get("os", {}).get("version", "unknown")
    browser = device_data.get("browser", {}).get("name", "unknown")
    is_rooted = device_data.get("is_rooted", False)
    is_emulator = device_data.get("is_emulator", False)
    
    # Considerar dispositivos comuns no Brasil
    common_devices = [
        "Samsung Galaxy", "Motorola Moto", "Xiaomi Redmi", "LG", "Apple iPhone", 
        "Positivo", "Multilaser"
    ]
    
    # Considerar operadoras brasileiras
    operator = device_data.get("network", {}).get("operator", "unknown")
    brazilian_operators = ["Vivo", "Claro", "Tim", "Oi", "Nextel", "Algar"]
    is_brazilian_operator = any(op in operator for op in brazilian_operators)
    
    # Verificar se já é um dispositivo conhecido para o usuário
    known_devices = user_history.get("known_devices", [])
    is_known_device = any(
        device.get("fingerprint") == device_data.get("fingerprint")
        for device in known_devices
    )
    
    # Verificar se o dispositivo é comum no Brasil
    is_common_device = any(
        common_name in device_data.get("model", "")
        for common_name in common_devices
    )
    
    # Verificar idioma do dispositivo
    device_language = device_data.get("language", "unknown")
    is_portuguese = device_language.startswith("pt") or device_language == "pt-BR"
    
    # Calcular score de risco
    risk_score = 0.0
    
    # Fatores de aumento de risco
    if is_rooted:
        risk_score += 0.4
        
    if is_emulator:
        risk_score += 0.5
        
    if not is_brazilian_operator:
        risk_score += 0.2
        
    if not is_known_device:
        risk_score += 0.3
        
    if not is_portuguese:
        risk_score += 0.2
        
    # Fatores de diminuição de risco
    if is_common_device:
        risk_score -= 0.1
        
    if is_known_device:
        risk_score -= 0.3
        
    if is_portuguese:
        risk_score -= 0.1
        
    if is_brazilian_operator:
        risk_score -= 0.1
    
    # Garantir limites
    risk_score = min(1.0, max(0.0, risk_score))
    
    # Determinar se o dispositivo é suspeito
    is_suspicious = risk_score > 0.6
    
    return {
        "risk_score": risk_score,
        "is_suspicious": is_suspicious,
        "is_known_device": is_known_device,
        "recommendations": [
            "Solicitar verificação adicional" if is_suspicious else None,
            "Verificar localização do dispositivo" if not is_brazilian_operator else None,
            "Adicionar à lista de dispositivos confiáveis" if not is_known_device and not is_suspicious else None
        ],
        "details": {
            "device_type": device_type,
            "os_name": os_name,
            "os_version": os_version,
            "browser": browser,
            "is_rooted": is_rooted,
            "is_emulator": is_emulator,
            "is_brazilian_operator": is_brazilian_operator,
            "is_common_device": is_common_device,
            "is_portuguese_language": is_portuguese
        }
    }


def integrate_bureau_credit(self, user_data: Dict[str, Any], 
                         bureau_adapter) -> Dict[str, Any]:
    """
    Integra informações de bureau de crédito brasileiro (Serasa/SPC) na análise.
    
    Args:
        user_data: Dados do usuário
        bureau_adapter: Adaptador para comunicação com o bureau
        
    Returns:
        Análise baseada nos dados do bureau
    """
    try:
        # Obter score de crédito via adaptador do bureau
        credit_result = bureau_adapter.check_credit_score(user_data.get("user_id"))
        
        if not credit_result.get("success", False):
            logger.warning(f"Falha ao obter dados de bureau para usuário {user_data.get('user_id')}: "
                          f"{credit_result.get('error', {}).get('message', 'Erro desconhecido')}")
            return {
                "success": False,
                "risk_score": 0.5,  # Score padrão quando não há dados
                "has_restrictions": False,
                "reason": "api_error"
            }
        
        # Extrair dados relevantes
        credit_data = credit_result.get("data", {})
        credit_score = credit_data.get("credit_score", 0)
        
        # Adaptar para escala brasileira (0-1000)
        scale_min = self.regional_rules["bureau_credit"]["score_scale"]["min"]
        scale_max = self.regional_rules["bureau_credit"]["score_scale"]["max"]
        normalized_score = (credit_score - scale_min) / (scale_max - scale_min)
        
        # Verificar se há restrições (negativado)
        has_restrictions = credit_data.get("has_restrictions", False)
        
        # Verificar CPF em lista de observação
        is_watchlisted = credit_data.get("is_watchlisted", False)
        
        # Calcular score de risco inverso (score alto = risco baixo)
        risk_score = 1 - normalized_score
        
        # Ajustar por restrições e lista de observação
        if has_restrictions:
            risk_score = max(risk_score, 0.8)  # Mínimo de 0.8 para negativados
            
        if is_watchlisted:
            risk_score = max(risk_score, 0.9)  # Mínimo de 0.9 para lista de observação
        
        # Verificar score mínimo de crédito para operações normais
        min_score = self.regional_rules["bureau_credit"]["min_score"]
        below_threshold = credit_score < min_score
        
        return {
            "success": True,
            "risk_score": risk_score,
            "credit_score": credit_score,
            "normalized_score": normalized_score,
            "has_restrictions": has_restrictions,
            "is_watchlisted": is_watchlisted,
            "below_threshold": below_threshold,
            "recommendations": [
                "Bloquear operação" if is_watchlisted else None,
                "Solicitar garantias adicionais" if has_restrictions else None,
                "Limitar valores de transação" if below_threshold else None
            ]
        }
    except Exception as e:
        logger.error(f"Erro ao integrar com bureau de crédito: {str(e)}")
        return {
            "success": False,
            "risk_score": 0.5,
            "error": str(e),
            "reason": "exception"
        }


def get_regional_rules(self) -> Dict[str, Any]:
    """
    Retorna as regras regionais para Brasil.
    
    Returns:
        Regras específicas para Brasil
    """
    return self.regional_rules