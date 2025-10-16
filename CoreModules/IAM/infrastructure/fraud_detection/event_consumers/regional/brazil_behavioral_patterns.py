#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Módulo de Análise Comportamental Regional para Brasil

Este módulo implementa padrões comportamentais específicos para o mercado brasileiro,
incluindo regras para detecção de anomalias de localização, validação de números
de telefone brasileiros, análise de transações de PIX, e integração com bureaus
de crédito locais (Serasa, SPC, etc).

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import re
import logging
import datetime
from typing import Dict, Any, List, Tuple, Optional, Union
from datetime import datetime, time

# Configuração do logger
logger = logging.getLogger("iam.trustguard.brazil")


class BrazilBehavioralPatterns:
    """
    Implementa análise comportamental específica para o mercado brasileiro.
    
    Esta classe contém lógica especializada para detectar anomalias comportamentais
    no contexto brasileiro, considerando particularidades como regiões de alto risco,
    padrões de números de telefone, sistemas bancários locais (PIX, TED, etc),
    comportamentos comuns de usuários brasileiros e integração com bureaus de crédito.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o analisador comportamental para Brasil.
        
        Args:
            config: Configurações específicas para Brasil, incluindo:
                - risk_zones: Zonas de alto risco no Brasil
                - trusted_banks: Bancos confiáveis brasileiros
                - trusted_operators: Operadoras de telefonia confiáveis
                - pix_limits: Limites para operações PIX
                - bureau_integration: Configurações para integração com bureaus
        """
        self.config = config
        self.region = "BR"
        
        # Zonas de risco (estados e municípios com maiores índices de fraude)
        self.high_risk_zones = config.get("high_risk_zones", [
            "RJ-Rio de Janeiro", "SP-São Paulo", "MG-Uberlândia", 
            "DF-Brasília", "PR-Curitiba", "PE-Recife"
        ])
        
        # Regiões urbanas principais (alta densidade de transações)
        self.urban_zones = config.get("urban_zones", [
            "SP-São Paulo", "RJ-Rio de Janeiro", "DF-Brasília", 
            "MG-Belo Horizonte", "PR-Curitiba", "RS-Porto Alegre",
            "BA-Salvador", "CE-Fortaleza", "AM-Manaus", "GO-Goiânia"
        ])
        
        # Bancos principais no Brasil
        self.trusted_banks = config.get("trusted_banks", [
            "Banco do Brasil", "Bradesco", "Caixa Econômica Federal", 
            "Itaú", "Santander", "Nubank", "Inter", "C6 Bank", 
            "BTG Pactual", "XP"
        ])
        
        # Operadoras de telefonia principais
        self.trusted_telecom = config.get("trusted_telecom", [
            "Vivo", "Tim", "Claro", "Oi", "Algar", "Nextel"
        ])
        
        # Padrões de validação de número de telefone brasileiro
        self.phone_patterns = {
            "mobile": re.compile(r"^(\+55|55|0)?(9\d{8})$"),  # Celulares com 9 dígitos
            "landline": re.compile(r"^(\+55|55|0)?([1-9][1-9]\d{7})$"),  # Fixos com 8 dígitos
            "country_code": "+55"
        }
        
        # Limites de transações PIX para análise comportamental
        self.pix_limits = config.get("pix_limits", {
            "instant": 5000.00,  # Limite para transações instantâneas
            "daily": 20000.00,   # Limite diário
            "unusual_hour_limit": 2000.00  # Limite para horários incomuns
        })
        
        # Intervalos de tempo para análise de comportamentos temporais
        # Horários típicos para transações no Brasil
        self.temporal_patterns = {
            "business_hours": (time(8, 0), time(18, 0)),  # 8h às 18h
            "banking_hours": (time(10, 0), time(16, 0)),  # 10h às 16h
            "unusual_hours": [(time(23, 0), time(5, 0))],  # 23h às 5h
            "high_risk_hours": [(time(0, 0), time(5, 0))],  # 0h às 5h
            "weekend": [5, 6]  # Sábado e domingo (5, 6 no datetime.weekday())
        }
        
        # Padrões de fraude conhecidos
        self.known_fraud_patterns = config.get("known_fraud_patterns", [
            {
                "pattern_type": "pix_multiple_accounts",
                "description": "Múltiplas transferências PIX para diferentes contas em curto período",
                "threshold": 5,  # Número de contas diferentes em um período
                "time_window": 60  # Período em minutos
            },
            {
                "pattern_type": "device_ip_mismatch",
                "description": "Dispositivo reportando geolocalização incompatível com IP",
                "severity": "high"
            },
            {
                "pattern_type": "unusual_region_access",
                "description": "Acesso de região geográfica incomum para o usuário",
                "severity": "medium"
            }
        ])
        
        # Regras regionais do Brasil para diferentes aspectos da análise
        self.regional_rules = {
            "authentication": {
                "location_change_threshold_km": 300,
                "rapid_location_change_hours": 3,
                "cross_state_mfa_required": True,
                "international_access_mfa_required": True
            },
            "session": {
                "max_session_time_minutes": 240,  # 4 horas
                "inactive_timeout_minutes": 30
            },
            "transaction": {
                "unusual_amount_multiplier": 5,
                "first_international_transfer_delay_hours": 24,
                "pix_to_new_account_delay_minutes": 30,
                "high_risk_merchant_categories": [
                    "gambling", "crypto", "digital_goods", 
                    "electronics", "jewelry", "gift_cards"
                ]
            },
            "device": {
                "trusted_device_period_days": 90,
                "max_active_devices": 5,
                "new_device_verification_required": True
            },
            "connection": {
                "tor_exit_nodes_blocked": True,
                "vpn_risk_score": 0.8,
                "public_wifi_risk_score": 0.6
            },
            "bureau_credit": {
                "min_score": 500,  # Escore mínimo para operações normais
                "score_scale": {
                    "min": 0,
                    "max": 1000
                },
                "verification_thresholds": {
                    "high_value_transaction": 0.7,
                    "new_account_registration": 0.6,
                    "credit_request": 0.8
                }
            }
        }
        
        logger.info(f"Analisador comportamental para Brasil inicializado com "
                   f"{len(self.high_risk_zones)} zonas de alto risco e "
                   f"{len(self.trusted_banks)} bancos confiáveis")
    
    def analyze_location_risk(self, location_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa o risco associado a uma localização específica no Brasil.
        
        Args:
            location_data: Dados de localização, incluindo:
                - state: Estado (UF)
                - city: Cidade
                - coordinates: Coordenadas geográficas
                - ip_address: Endereço IP
                
        Returns:
            Análise de risco da localização
        """
        state = location_data.get("state", "")
        city = location_data.get("city", "")
        location_string = f"{state}-{city}"
        
        # Verificar se é zona de alto risco
        is_high_risk = any(zone in location_string for zone in self.high_risk_zones)
        
        # Verificar se é uma zona urbana principal
        is_urban_center = any(zone in location_string for zone in self.urban_zones)
        
        # Verificar se é uma localização externa ao Brasil
        is_external = location_data.get("country") not in ["BR", "Brasil", "Brazil"]
        
        # Calcular score de risco
        risk_score = 0.5  # Base
        
        if is_high_risk:
            risk_score += 0.3
            
        if is_urban_center:
            risk_score -= 0.1  # Zonas urbanas principais são menos suspeitas (mais comuns)
            
        if is_external:
            risk_score += 0.4  # Localização externa ao Brasil aumenta o risco
        
        return {
            "risk_score": min(1.0, max(0.0, risk_score)),
            "is_high_risk_zone": is_high_risk,
            "is_urban_center": is_urban_center,
            "is_external": is_external,
            "details": {
                "state": state,
                "city": city,
                "known_risk_area": is_high_risk,
                "region": "BR" if not is_external else "EXTERNAL"
            }
        }
    
    def detect_rapid_location_change(self, 
                                   current_location: Dict[str, Any],
                                   previous_location: Dict[str, Any],
                                   time_diff_hours: float) -> Dict[str, Any]:
        """
        Detecta mudanças rápidas de localização que seriam fisicamente impossíveis.
        
        Args:
            current_location: Localização atual
            previous_location: Localização anterior
            time_diff_hours: Diferença de tempo em horas
            
        Returns:
            Resultado da detecção de mudança rápida
        """
        # Extrair coordenadas das localizações
        try:
            current_coords = current_location.get("coordinates", {"lat": 0, "lon": 0})
            previous_coords = previous_location.get("coordinates", {"lat": 0, "lon": 0})
            
            # Calcular distância aproximada (km)
            from math import radians, sin, cos, sqrt, atan2
            
            def haversine(lat1, lon1, lat2, lon2):
                # Raio da Terra em km
                R = 6371.0
                
                # Converter para radianos
                lat1 = radians(float(lat1))
                lon1 = radians(float(lon1))
                lat2 = radians(float(lat2))
                lon2 = radians(float(lon2))
                
                # Diferenças
                dlon = lon2 - lon1
                dlat = lat2 - lat1
                
                # Fórmula de Haversine
                a = sin(dlat / 2)**2 + cos(lat1) * cos(lat2) * sin(dlon / 2)**2
                c = 2 * atan2(sqrt(a), sqrt(1 - a))
                
                # Distância
                distance = R * c
                
                return distance
            
            distance = haversine(
                current_coords.get("lat", 0), 
                current_coords.get("lon", 0),
                previous_coords.get("lat", 0), 
                previous_coords.get("lon", 0)
            )
            
            # Velocidade necessária em km/h
            speed = distance / time_diff_hours if time_diff_hours > 0 else float('inf')
            
            # Definir limites para Brasil
            # - Voos domésticos: ~800 km/h
            # - Carros em rodovias: ~120 km/h
            # - Cruzando estados: Contexto geográfico brasileiro
            impossible_travel = speed > 900  # Acima da velocidade de avião comercial
            unusual_travel = speed > 150  # Acima da velocidade normal de carro
            
            # Verificar troca de estado
            current_state = current_location.get("state", "")
            previous_state = previous_location.get("state", "")
            changed_state = current_state != previous_state and current_state and previous_state
            
            # Verificar se uma das localizações é internacional
            current_country = current_location.get("country", "BR")
            previous_country = previous_location.get("country", "BR")
            international_change = current_country != previous_country
            
            return {
                "distance_km": round(distance, 2),
                "time_diff_hours": round(time_diff_hours, 2),
                "speed_kmh": round(speed, 2),
                "impossible_travel": impossible_travel,
                "unusual_travel": unusual_travel,
                "changed_state": changed_state,
                "international_change": international_change,
                "risk_score": 0.9 if impossible_travel else (0.7 if unusual_travel else 0.3),
                "details": {
                    "from_location": f"{previous_location.get('city', '')}, {previous_location.get('state', '')}, {previous_country}",
                    "to_location": f"{current_location.get('city', '')}, {current_location.get('state', '')}, {current_country}"
                }
            }
        except Exception as e:
            logger.error(f"Erro ao calcular mudança de localização: {str(e)}")
            return {
                "error": str(e),
                "risk_score": 0.5,  # Valor padrão em caso de erro
                "impossible_travel": False,
                "unusual_travel": False
            }
    
    def analyze_pix_behavior(self, transaction_data: Dict[str, Any], 
                          user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa comportamento em transações PIX, considerando limites e padrões brasileiros.
        
        Args:
            transaction_data: Dados da transação PIX
            user_history: Histórico do usuário
            
        Returns:
            Análise de comportamento PIX
        """
        # Extrair dados relevantes
        amount = transaction_data.get("amount", 0)
        recipient_key = transaction_data.get("recipient_key", "")
        recipient_bank = transaction_data.get("recipient_bank", "")
        recipient_is_new = transaction_data.get("recipient_is_new", True)
        transaction_time = transaction_data.get("timestamp", datetime.now())
        
        if isinstance(transaction_time, str):
            try:
                transaction_time = datetime.fromisoformat(transaction_time.replace('Z', '+00:00'))
            except:
                transaction_time = datetime.now()
        
        # Verificar se o horário é incomum
        transaction_time_obj = transaction_time.time()
        is_unusual_hour = any(
            start <= transaction_time_obj <= end 
            for start, end in self.temporal_patterns["unusual_hours"]
        )
        
        is_weekend = transaction_time.weekday() in self.temporal_patterns["weekend"]
        
        # Verificar se o banco destinatário é confiável
        is_trusted_bank = recipient_bank in self.trusted_banks
        
        # Verificar se o valor está acima dos limites
        is_above_instant_limit = amount > self.pix_limits["instant"]
        is_above_unusual_hour_limit = is_unusual_hour and amount > self.pix_limits["unusual_hour_limit"]
        
        # Verificar histórico do usuário
        avg_transaction = user_history.get("avg_pix_amount", amount / 2)
        max_transaction = user_history.get("max_pix_amount", amount)
        transaction_count = user_history.get("pix_count", 0)
        is_first_transaction = transaction_count == 0
        
        # Verificar anomalias de valor
        is_unusual_amount = amount > avg_transaction * 3 and amount > 1000
        is_record_amount = amount > max_transaction * 1.5 and amount > 2000
        
        # Cálculo do score de risco
        risk_score = 0.0
        
        # Fatores de aumento de risco
        if is_above_instant_limit:
            risk_score += 0.3
            
        if is_unusual_hour:
            risk_score += 0.2
            
        if is_above_unusual_hour_limit:
            risk_score += 0.3
            
        if not is_trusted_bank:
            risk_score += 0.1
            
        if recipient_is_new:
            risk_score += 0.2
            
        if is_unusual_amount:
            risk_score += 0.2
            
        if is_record_amount:
            risk_score += 0.3
            
        if is_first_transaction and amount > 1000:
            risk_score += 0.3
            
        # Fatores de diminuição de risco
        if transaction_count > 10:
            risk_score -= 0.1
            
        if is_trusted_bank:
            risk_score -= 0.1
            
        # Garantir limite entre 0 e 1
        risk_score = min(1.0, max(0.0, risk_score))
        
        # Determinar se é uma transação suspeita
        is_suspicious = risk_score > 0.7
        requires_additional_verification = risk_score > 0.5
        
        # Gerar recomendações baseadas na análise
        recommendations = []
        
        if is_suspicious:
            recommendations.append("Solicitar confirmação multicanal")
            recommendations.append("Verificar titularidade da conta destino")
            
            if is_unusual_hour:
                recommendations.append("Contato telefônico com cliente")
        
        if requires_additional_verification:
            if recipient_is_new:
                recommendations.append("Aguardar período de segurança para novos destinatários")
                
            if is_unusual_amount or is_record_amount:
                recommendations.append("Verificar origem dos recursos")
        
        return {
            "risk_score": risk_score,
            "is_suspicious": is_suspicious,
            "requires_additional_verification": requires_additional_verification,
            "recommendations": recommendations,
            "details": {
                "amount": amount,
                "is_unusual_hour": is_unusual_hour,
                "is_weekend": is_weekend,
                "is_trusted_bank": is_trusted_bank,
                "is_unusual_amount": is_unusual_amount,
                "is_record_amount": is_record_amount,
                "recipient_is_new": recipient_is_new
            }
        }
        
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


def create_brazil_analyzer(config: Dict[str, Any]) -> BrazilBehavioralPatterns:
    """
    Factory function para criar uma instância do analisador comportamental para Brasil.
    
    Args:
        config: Configurações específicas para o analisador
        
    Returns:
        Instância configurada do analisador comportamental para Brasil
    """
    return BrazilBehavioralPatterns(config)