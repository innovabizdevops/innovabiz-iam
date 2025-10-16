#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Módulo de Análise Comportamental Regional para Moçambique

Este módulo implementa padrões comportamentais específicos para o mercado moçambicano,
incluindo regras para detecção de anomalias de localização, validação de números
de telefone moçambicanos, análise de transações de mobile money, e integração com
serviços financeiros locais.

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
logger = logging.getLogger("iam.trustguard.mozambique")


class MozambiqueBehavioralPatterns:
    """
    Implementa análise comportamental específica para o mercado moçambicano.
    
    Esta classe contém lógica especializada para detectar anomalias comportamentais
    no contexto moçambicano, considerando particularidades como regiões de alto risco,
    padrões de números de telefone, sistemas de mobile money locais (M-Pesa, mKesh),
    comportamentos comuns de usuários moçambicanos e integrações com sistemas financeiros.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o analisador comportamental para Moçambique.
        
        Args:
            config: Configurações específicas para Moçambique, incluindo:
                - risk_zones: Zonas de alto risco em Moçambique
                - trusted_operators: Operadoras de telefonia confiáveis
                - mobile_money_limits: Limites para operações de mobile money
                - bureau_integration: Configurações para integração com sistemas
        """
        self.config = config
        self.region = "MZ"
        
        # Zonas de risco (províncias e distritos com maiores índices de fraude)
        self.high_risk_zones = config.get("high_risk_zones", [
            "Maputo-Cidade", "Maputo-Matola", "Sofala-Beira", 
            "Nampula-Cidade", "Cabo Delgado-Pemba", "Gaza-Xai-Xai"
        ])
        
        # Regiões urbanas principais (alta densidade de transações)
        self.urban_zones = config.get("urban_zones", [
            "Maputo-Cidade", "Sofala-Beira", "Nampula-Cidade", 
            "Zambézia-Quelimane", "Inhambane-Cidade", "Tete-Cidade",
            "Cabo Delgado-Pemba", "Manica-Chimoio", "Gaza-Xai-Xai", "Niassa-Lichinga"
        ])
        
        # Operadoras de telefonia principais
        self.trusted_telecom = config.get("trusted_telecom", [
            "Vodacom", "Tmcel", "Movitel"
        ])
        
        # Serviços de mobile money principais
        self.mobile_money_services = config.get("mobile_money_services", [
            "M-Pesa", "mKesh", "e-Mola"
        ])
        
        # Padrões de validação de número de telefone moçambicano
        self.phone_patterns = {
            "vodacom": re.compile(r"^(\+258|258|0)?(8[234]\d{7})$"),  # Padrão Vodacom (82, 83, 84)
            "tmcel": re.compile(r"^(\+258|258|0)?(8[56]\d{7})$"),     # Padrão TMcel (85, 86)
            "movitel": re.compile(r"^(\+258|258|0)?(8[79]\d{7})$"),   # Padrão Movitel (87, 89)
            "country_code": "+258"
        }
        
        # Limites de transações mobile money para análise comportamental
        self.mobile_money_limits = config.get("mobile_money_limits", {
            "instant": 5000.00,  # Limite para transações instantâneas em MZN
            "daily": 20000.00,   # Limite diário em MZN
            "unusual_hour_limit": 2000.00  # Limite para horários incomuns em MZN
        })
        
        # Intervalos de tempo para análise de comportamentos temporais
        # Horários típicos para transações em Moçambique
        self.temporal_patterns = {
            "business_hours": (time(8, 0), time(17, 0)),  # 8h às 17h
            "banking_hours": (time(8, 0), time(15, 0)),   # 8h às 15h
            "unusual_hours": [(time(22, 0), time(6, 0))], # 22h às 6h
            "high_risk_hours": [(time(0, 0), time(5, 0))], # 0h às 5h
            "weekend": [5, 6]  # Sábado e domingo (5, 6 no datetime.weekday())
        }
        
        # Padrões de fraude conhecidos
        self.known_fraud_patterns = config.get("known_fraud_patterns", [
            {
                "pattern_type": "mobile_money_multiple_accounts",
                "description": "Múltiplas transferências para diferentes contas em curto período",
                "threshold": 5,  # Número de contas diferentes em um período
                "time_window": 60  # Período em minutos
            },
            {
                "pattern_type": "cross_border_transactions",
                "description": "Transações em áreas fronteiriças em curto período de tempo",
                "severity": "high"
            },
            {
                "pattern_type": "unusual_region_access",
                "description": "Acesso de região geográfica incomum para o usuário",
                "severity": "medium"
            }
        ])
        
        # Bancos principais em Moçambique
        self.trusted_banks = config.get("trusted_banks", [
            "BCI", "Millennium BIM", "Standard Bank", "Absa", 
            "Moza Banco", "FNB Moçambique", "Banco Único", 
            "Ecobank", "Nedbank"
        ])
        
        # Regras regionais de Moçambique para diferentes aspectos da análise
        self.regional_rules = {
            "authentication": {
                "location_change_threshold_km": 250,
                "rapid_location_change_hours": 3,
                "cross_province_mfa_required": True,
                "international_access_mfa_required": True
            },
            "session": {
                "max_session_time_minutes": 240,  # 4 horas
                "inactive_timeout_minutes": 30
            },
            "transaction": {
                "unusual_amount_multiplier": 4,
                "first_international_transfer_delay_hours": 24,
                "mobile_money_to_new_recipient_delay_minutes": 30,
                "high_risk_merchant_categories": [
                    "gambling", "crypto", "digital_goods", 
                    "electronics", "jewelry", "money_transfer"
                ]
            },
            "device": {
                "trusted_device_period_days": 90,
                "max_active_devices": 3,
                "new_device_verification_required": True
            },
            "connection": {
                "tor_exit_nodes_blocked": True,
                "vpn_risk_score": 0.9,
                "public_wifi_risk_score": 0.7
            },
            "mobile_money": {
                "new_recipient_cooling_period_minutes": 30,
                "cross_operator_cooling_period_minutes": 60,
                "high_value_verification_threshold_mzn": 2500,
                "daily_limit_mzn": 25000
            }
        }
        
        logger.info(f"Analisador comportamental para Moçambique inicializado com "
                   f"{len(self.high_risk_zones)} zonas de alto risco e "
                   f"{len(self.trusted_telecom)} operadoras confiáveis")
    
    def analyze_location_risk(self, location_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa o risco associado a uma localização específica em Moçambique.
        
        Args:
            location_data: Dados de localização, incluindo:
                - province: Província
                - district: Distrito
                - coordinates: Coordenadas geográficas
                - ip_address: Endereço IP
                
        Returns:
            Análise de risco da localização
        """
        province = location_data.get("province", "")
        district = location_data.get("district", "")
        location_string = f"{province}-{district}"
        
        # Verificar se é zona de alto risco
        is_high_risk = any(zone in location_string for zone in self.high_risk_zones)
        
        # Verificar se é uma zona urbana principal
        is_urban_center = any(zone in location_string for zone in self.urban_zones)
        
        # Verificar se é uma localização externa a Moçambique
        is_external = location_data.get("country") not in ["MZ", "MOZ", "Moçambique", "Mozambique"]
        
        # Verificar se está próximo a fronteiras
        border_regions = ["Gaza-Chicualacuala", "Niassa-Lago", "Tete-Zumbo", 
                          "Tete-Changara", "Manica-Machipanda", "Cabo Delgado-Mueda"]
        is_border_region = any(region in location_string for region in border_regions)
        
        # Calcular score de risco
        risk_score = 0.5  # Base
        
        if is_high_risk:
            risk_score += 0.3
            
        if is_urban_center:
            risk_score -= 0.1  # Zonas urbanas principais são menos suspeitas (mais comuns)
            
        if is_external:
            risk_score += 0.4  # Localização externa a Moçambique aumenta o risco
            
        if is_border_region:
            risk_score += 0.2  # Regiões de fronteira têm risco elevado
        
        return {
            "risk_score": min(1.0, max(0.0, risk_score)),
            "is_high_risk_zone": is_high_risk,
            "is_urban_center": is_urban_center,
            "is_external": is_external,
            "is_border_region": is_border_region,
            "details": {
                "province": province,
                "district": district,
                "known_risk_area": is_high_risk,
                "region": "MZ" if not is_external else "EXTERNAL"
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
            
            # Definir limites para Moçambique
            # - Voos domésticos: ~700 km/h
            # - Carros em rodovias: ~80-100 km/h (considerando estado das estradas)
            # - Cruzando províncias: Contexto geográfico moçambicano
            impossible_travel = speed > 800  # Acima da velocidade de avião comercial
            unusual_travel = speed > 120  # Acima da velocidade normal de carro em MZ
            
            # Verificar troca de província
            current_province = current_location.get("province", "")
            previous_province = previous_location.get("province", "")
            changed_province = current_province != previous_province and current_province and previous_province
            
            # Verificar se uma das localizações é internacional
            current_country = current_location.get("country", "MZ")
            previous_country = previous_location.get("country", "MZ")
            international_change = current_country != previous_country
            
            return {
                "distance_km": round(distance, 2),
                "time_diff_hours": round(time_diff_hours, 2),
                "speed_kmh": round(speed, 2),
                "impossible_travel": impossible_travel,
                "unusual_travel": unusual_travel,
                "changed_province": changed_province,
                "international_change": international_change,
                "risk_score": 0.9 if impossible_travel else (0.7 if unusual_travel else 0.3),
                "details": {
                    "from_location": f"{previous_location.get('district', '')}, {previous_location.get('province', '')}, {previous_country}",
                    "to_location": f"{current_location.get('district', '')}, {current_location.get('province', '')}, {current_country}"
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
    
    def validate_mozambique_phone(self, phone_number: str) -> Dict[str, Any]:
        """
        Valida um número de telefone moçambicano.
        
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
            if not cleaned.startswith('+258'):
                return {
                    "valid": False,
                    "reason": "country_code_mismatch",
                    "details": {
                        "expected_country_code": "+258",
                        "provided_country_code": cleaned[:4]
                    }
                }
        else:
            has_country_code = False
        
        # Verificar operadora Vodacom (prefixos 82, 83, 84)
        vodacom_match = self.phone_patterns["vodacom"].search(cleaned)
        if vodacom_match:
            return {
                "valid": True,
                "operator": "Vodacom",
                "has_mobile_money": True,
                "mobile_money_service": "M-Pesa",
                "has_country_code": has_country_code,
                "normalized": f"+258{vodacom_match.group(2)}"
            }
        
        # Verificar operadora TMcel (prefixos 85, 86)
        tmcel_match = self.phone_patterns["tmcel"].search(cleaned)
        if tmcel_match:
            return {
                "valid": True,
                "operator": "Tmcel",
                "has_mobile_money": True,
                "mobile_money_service": "mKesh",
                "has_country_code": has_country_code,
                "normalized": f"+258{tmcel_match.group(2)}"
            }
        
        # Verificar operadora Movitel (prefixos 87, 89)
        movitel_match = self.phone_patterns["movitel"].search(cleaned)
        if movitel_match:
            return {
                "valid": True,
                "operator": "Movitel",
                "has_mobile_money": True,
                "mobile_money_service": "e-Mola",
                "has_country_code": has_country_code,
                "normalized": f"+258{movitel_match.group(2)}"
            }
        
        # Se não corresponder a nenhum padrão válido
        return {
            "valid": False,
            "reason": "invalid_format",
            "details": {
                "expected_formats": [
                    "+258 82XXXXXXX (Vodacom)",
                    "+258 83XXXXXXX (Vodacom)",
                    "+258 84XXXXXXX (Vodacom)",
                    "+258 85XXXXXXX (Tmcel)",
                    "+258 86XXXXXXX (Tmcel)",
                    "+258 87XXXXXXX (Movitel)",
                    "+258 89XXXXXXX (Movitel)"
                ]
            }
        }
    
    def analyze_mobile_money_behavior(self, transaction_data: Dict[str, Any], 
                                    user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa comportamento em transações de mobile money, considerando
        limites e padrões moçambicanos.
        
        Args:
            transaction_data: Dados da transação de mobile money
            user_history: Histórico do usuário
            
        Returns:
            Análise de comportamento de mobile money
        """
        # Extrair dados relevantes
        amount = transaction_data.get("amount", 0)
        service = transaction_data.get("service", "")  # M-Pesa, mKesh ou e-Mola
        recipient_number = transaction_data.get("recipient_number", "")
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
        
        # Verificar se o serviço é confiável
        is_trusted_service = service in self.mobile_money_services
        
        # Verificar se o valor está acima dos limites
        is_above_instant_limit = amount > self.mobile_money_limits["instant"]
        is_above_unusual_hour_limit = is_unusual_hour and amount > self.mobile_money_limits["unusual_hour_limit"]
        
        # Verificar histórico do usuário
        avg_transaction = user_history.get("avg_mobile_money_amount", amount / 2)
        max_transaction = user_history.get("max_mobile_money_amount", amount)
        transaction_count = user_history.get("mobile_money_count", 0)
        is_first_transaction = transaction_count == 0
        
        # Verificar operadora do destinatário (cross-operator é mais suspeito)
        recipient_validation = self.validate_mozambique_phone(recipient_number)
        is_cross_operator = False
        recipient_operator = "unknown"
        
        if recipient_validation.get("valid", False):
            recipient_operator = recipient_validation.get("operator", "unknown")
            sender_operator = ""
            
            if service == "M-Pesa":
                sender_operator = "Vodacom"
            elif service == "mKesh":
                sender_operator = "Tmcel"
            elif service == "e-Mola":
                sender_operator = "Movitel"
                
            is_cross_operator = sender_operator != recipient_operator and sender_operator
        
        # Verificar anomalias de valor
        is_unusual_amount = amount > avg_transaction * 3 and amount > 500
        is_record_amount = amount > max_transaction * 1.5 and amount > 1000
        
        # Cálculo do score de risco
        risk_score = 0.0
        
        # Fatores de aumento de risco
        if is_above_instant_limit:
            risk_score += 0.3
            
        if is_unusual_hour:
            risk_score += 0.2
            
        if is_above_unusual_hour_limit:
            risk_score += 0.3
            
        if not is_trusted_service:
            risk_score += 0.2
            
        if recipient_is_new:
            risk_score += 0.2
            
        if is_unusual_amount:
            risk_score += 0.2
            
        if is_record_amount:
            risk_score += 0.3
            
        if is_first_transaction and amount > 500:
            risk_score += 0.3
            
        if is_cross_operator:
            risk_score += 0.2
            
        # Fatores de diminuição de risco
        if transaction_count > 10:
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
            
            if is_unusual_hour:
                recommendations.append("Contato por SMS com cliente")
        
        if requires_additional_verification:
            if recipient_is_new:
                recommendations.append("Aguardar período de segurança para novos destinatários")
                
            if is_unusual_amount or is_record_amount:
                recommendations.append("Solicitar confirmação de valor por outro canal")
                
            if is_cross_operator:
                recommendations.append("Verificar titularidade da conta destino")
        
        return {
            "risk_score": risk_score,
            "is_suspicious": is_suspicious,
            "requires_additional_verification": requires_additional_verification,
            "recommendations": recommendations,
            "details": {
                "amount": amount,
                "service": service,
                "recipient_operator": recipient_operator,
                "is_cross_operator": is_cross_operator,
                "is_unusual_hour": is_unusual_hour,
                "is_weekend": is_weekend,
                "is_trusted_service": is_trusted_service,
                "is_unusual_amount": is_unusual_amount,
                "is_record_amount": is_record_amount,
                "recipient_is_new": recipient_is_new
            }
        }
    
    def analyze_device_context(self, device_data: Dict[str, Any], 
                            user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa o contexto do dispositivo em relação ao mercado moçambicano.
        
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
        
        # Considerar dispositivos comuns em Moçambique
        common_devices = [
            "Samsung Galaxy", "Tecno", "Huawei", "iTel", 
            "Xiaomi Redmi", "Nokia", "Apple iPhone"
        ]
        
        # Considerar operadoras moçambicanas
        operator = device_data.get("network", {}).get("operator", "unknown")
        mozambican_operators = ["Vodacom", "Tmcel", "Movitel"]
        is_mozambican_operator = any(op in operator for op in mozambican_operators)
        
        # Verificar se já é um dispositivo conhecido para o usuário
        known_devices = user_history.get("known_devices", [])
        is_known_device = any(
            device.get("fingerprint") == device_data.get("fingerprint")
            for device in known_devices
        )
        
        # Verificar se o dispositivo é comum em Moçambique
        is_common_device = any(
            common_name in device_data.get("model", "")
            for common_name in common_devices
        )
        
        # Verificar idioma do dispositivo
        device_language = device_data.get("language", "unknown")
        is_portuguese = device_language.startswith("pt") or device_language == "pt-PT"
        
        # Calcular score de risco
        risk_score = 0.0
        
        # Fatores de aumento de risco
        if is_rooted:
            risk_score += 0.4
            
        if is_emulator:
            risk_score += 0.5
            
        if not is_mozambican_operator:
            risk_score += 0.3
            
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
            
        if is_mozambican_operator:
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
                "Verificar localização do dispositivo" if not is_mozambican_operator else None,
                "Adicionar à lista de dispositivos confiáveis" if not is_known_device and not is_suspicious else None
            ],
            "details": {
                "device_type": device_type,
                "os_name": os_name,
                "os_version": os_version,
                "browser": browser,
                "is_rooted": is_rooted,
                "is_emulator": is_emulator,
                "is_mozambican_operator": is_mozambican_operator,
                "is_common_device": is_common_device,
                "is_portuguese_language": is_portuguese
            }
        }
    
    def get_regional_rules(self) -> Dict[str, Any]:
        """
        Retorna as regras regionais para Moçambique.
        
        Returns:
            Regras específicas para Moçambique
        """
        return self.regional_rules


def create_mozambique_analyzer(config: Dict[str, Any]) -> MozambiqueBehavioralPatterns:
    """
    Factory function para criar uma instância do analisador comportamental para Moçambique.
    
    Args:
        config: Configurações específicas para o analisador
        
    Returns:
        Instância configurada do analisador comportamental para Moçambique
    """
    return MozambiqueBehavioralPatterns(config)