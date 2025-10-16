#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Script de exemplo para uso do Agente de Análise Comportamental para Moçambique

Este script demonstra a utilização do MozambiqueBehaviorAgent para análise
de fraudes contextuais específicas para o mercado moçambicano.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
import json
import logging
from datetime import datetime, timedelta
from typing import Dict, Any

# Configurar logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("example_mozambique")

# Importar o agente de comportamento para Moçambique
try:
    from src.api.services.fraud_detection.agents.behavioral.mozambique_behavior_agent import MozambiqueBehaviorAgent
except ImportError:
    # Ajustar o caminho conforme a estrutura do projeto
    import sys
    sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
    from src.api.services.fraud_detection.agents.behavioral.mozambique_behavior_agent import MozambiqueBehaviorAgent


def generate_mock_data() -> Dict[str, Any]:
    """
    Gera dados simulados para teste do agente comportamental
    """
    # Dados de conta
    account_data = {
        "account_id": "MZ987654321",
        "user_id": "user_mz_54321",
        "name": "António Armando Guebuza",
        "created_at": (datetime.now() - timedelta(days=60)).isoformat(),
        "kyc_status": "verified",
        "tax_id": "123456789",  # NUIT moçambicano
        "id_document": {
            "type": "bi",
            "number": "12345678M",
            "issuer": "República de Moçambique",
            "issue_date": "2018-03-15",
            "expiry_date": "2028-03-15"
        },
        "suspicious_activity_history": [
            {
                "type": "failed_login",
                "timestamp": (datetime.now() - timedelta(days=10)).isoformat(),
                "details": "Múltiplas tentativas de login malsucedidas"
            }
        ],
        "address": {
            "street": "Av. Julius Nyerere, 702",
            "neighborhood": "Polana",
            "city": "Maputo",
            "district": "Maputo Cidade",
            "postal_code": "1102",
            "country": "Moçambique",
            "is_temporary": False,
            "differs_from_fiscal_address": False
        },
        "economic_activity": "comercio_geral",
        "is_pep": False,
        "pep_relatives": []
    }
    
    # Dados de localização
    location_data = {
        "ip_address": "197.218.27.54",  # IP simulado de Moçambique
        "country": "MZ",
        "city": "Maputo",
        "district": "Maputo Cidade",
        "latitude": -25.9686,
        "longitude": 32.5804,
        "timestamp": datetime.now().isoformat(),
        "is_vpn": False,
        "is_proxy": False,
        "coords": {
            "latitude": -25.9686,
            "longitude": 32.5804
        },
        "border_proximity": {
            "is_near_border": True,
            "distance_to_border_km": 78,
            "bordering_country": "SZ"  # Eswatini (Suazilândia)
        }
    }
    
    # Histórico de localizações
    user_location_history = {
        "user_id": "user_mz_54321",
        "recent_locations": [
            {
                "ip_address": "197.218.27.54",
                "country": "MZ",
                "city": "Maputo",
                "timestamp": (datetime.now() - timedelta(hours=24)).isoformat(),
                "coords": {
                    "latitude": -25.9686,
                    "longitude": 32.5804
                }
            },
            {
                "ip_address": "197.218.96.123",
                "country": "MZ",
                "city": "Beira",
                "timestamp": (datetime.now() - timedelta(days=5)).isoformat(),
                "coords": {
                    "latitude": -19.8436,
                    "longitude": 34.8389
                }
            }
        ],
        "typical_locations": [
            {
                "city": "Maputo",
                "district": "Maputo Cidade",
                "frequency": "high",
                "coords": {
                    "latitude": -25.9686,
                    "longitude": 32.5804
                }
            },
            {
                "city": "Beira",
                "district": "Sofala",
                "frequency": "medium",
                "coords": {
                    "latitude": -19.8436,
                    "longitude": 34.8389
                }
            }
        ]
    }
    
    # Dados do dispositivo
    device_data = {
        "device_id": "dev-mz-12345",
        "device_type": "smartphone",
        "platform": "mobile",
        "os": {
            "name": "Android",
            "version": "11.0"
        },
        "browser": {
            "name": "Chrome Mobile",
            "version": "98.0.4758.101"
        },
        "ip_address": "197.218.27.54",
        "user_agent": "Mozilla/5.0 (Linux; Android 11; SM-A715F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.101 Mobile Safari/537.36",
        "screen_resolution": "1080x2400",
        "is_rooted": False,
        "is_jailbroken": False,
        "is_emulator": False,
        "is_vpn": False,
        "is_proxy": False,
        "is_tor": False,
        "session_start": datetime.now().isoformat(),
        "timezone": "Africa/Maputo",
        "language": "pt-MZ",
        "linked_accounts": ["user_mz_54321"],
        "fingerprint_changed": False,
        "coords": {
            "latitude": -25.9686,
            "longitude": 32.5804
        }
    }
    
    # Histórico de dispositivos
    user_device_history = {
        "user_id": "user_mz_54321",
        "known_devices": [
            {
                "device_id": "dev-mz-12345",
                "first_seen": (datetime.now() - timedelta(days=60)).isoformat(),
                "last_seen": (datetime.now() - timedelta(hours=12)).isoformat()
            }
        ],
        "last_login": {
            "device_id": "dev-mz-12345",
            "timestamp": (datetime.now() - timedelta(hours=12)).isoformat(),
            "ip_address": "197.218.27.54"
        },
        "sessions": [
            {
                "login_time": (datetime.now() - timedelta(hours=12)).isoformat(),
                "ip_address": "197.218.27.54",
                "user_agent": "Mozilla/5.0 (Linux; Android 11; SM-A715F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.101 Mobile Safari/537.36",
                "screen_resolution": "1080x2400",
                "timezone": "Africa/Maputo",
                "language": "pt-MZ",
                "referrer": "https://www.banco.co.mz"
            },
            {
                "login_time": (datetime.now() - timedelta(days=3)).isoformat(),
                "ip_address": "197.218.96.123",
                "user_agent": "Mozilla/5.0 (Linux; Android 11; SM-A715F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.101 Mobile Safari/537.36",
                "screen_resolution": "1080x2400",
                "timezone": "Africa/Maputo",
                "language": "pt-MZ",
                "referrer": "https://www.banco.co.mz/mobile"
            }
        ]
    }
    
    # Dados do usuário para análise regional
    user_data = {
        "user_id": "user_mz_54321",
        "name": "António Armando Guebuza",
        "tax_id": "123456789",
        "id_document": {
            "type": "bi",
            "number": "12345678M"
        },
        "phone": "+258843012345"
    }
    
    return {
        "account_data": account_data,
        "location_data": location_data,
        "user_location_history": user_location_history,
        "device_data": device_data,
        "user_device_history": user_device_history,
        "user_data": user_data
    }


def main():
    """Função principal para demonstração do agente comportamental de Moçambique"""
    try:
        logger.info("Inicializando agente de comportamento para Moçambique...")
        
        # Inicializar agente de Moçambique
        mozambique_agent = MozambiqueBehaviorAgent()
        
        # Gerar dados simulados para teste
        mock_data = generate_mock_data()
        
        logger.info("Executando análises de risco para Moçambique...")
        
        # Análise de risco de conta
        account_risk = mozambique_agent.evaluate_account_risk(mock_data["account_data"])
        print("\n=== ANÁLISE DE RISCO DE CONTA ===")
        print(f"Pontuação: {account_risk['risk_score']:.2f} ({account_risk['risk_level']})")
        print("Fatores de risco identificados:")
        for factor in account_risk.get("risk_factors", []):
            print(f"  • {factor['factor']}: {factor['score']:.2f}")
            if "details" in factor:
                print(f"    - {factor['details']}")
        
        # Análise de anomalias de localização
        location_risk = mozambique_agent.detect_location_anomalies(
            mock_data["location_data"],
            mock_data["user_location_history"]
        )
        print("\n=== ANÁLISE DE ANOMALIAS DE LOCALIZAÇÃO ===")
        print(f"Pontuação: {location_risk['risk_score']:.2f} ({location_risk['risk_level']})")
        print("Anomalias identificadas:")
        for anomaly in location_risk.get("anomalies", []):
            print(f"  • {anomaly['type']}: {anomaly['risk']:.2f}")
            if "details" in anomaly:
                print(f"    - {anomaly['details']}")
        
        # Análise de comportamento de dispositivo
        device_risk = mozambique_agent.analyze_device_behavior(
            mock_data["device_data"],
            mock_data["user_device_history"]
        )
        print("\n=== ANÁLISE DE COMPORTAMENTO DE DISPOSITIVO ===")
        print(f"Pontuação: {device_risk['risk_score']:.2f} ({device_risk['risk_level']})")
        print("Fatores de risco identificados:")
        for factor in device_risk.get("risk_factors", []):
            print(f"  • {factor['factor']}: {factor['score']:.2f}")
            if "details" in factor:
                print(f"    - {factor['details']}")
        
        # Análise de fatores de risco regionais
        regional_risk = mozambique_agent.get_regional_risk_factors(mock_data["user_data"])
        print("\n=== ANÁLISE DE FATORES DE RISCO REGIONAIS ===")
        print(f"Pontuação: {regional_risk['risk_score']:.2f} ({regional_risk['risk_level']})")
        print("Fatores de risco identificados:")
        for factor in regional_risk.get("risk_factors", []):
            print(f"  • {factor['factor']}: {factor['score']:.2f}")
            if "details" in factor:
                print(f"    - {factor['details']}")
        
        # Calcular pontuação combinada
        combined_risk = mozambique_agent._calculate_combined_risk_score(
            account_risk, 
            location_risk, 
            device_risk, 
            regional_risk
        )
        print("\n=== ANÁLISE DE RISCO COMBINADA ===")
        print(f"Pontuação final: {combined_risk['risk_score']:.2f} ({combined_risk['risk_level']})")
        print("Pontuações por categoria:")
        for category, score in combined_risk.get("scores_by_category", {}).items():
            print(f"  • {category}: {score:.2f}")
        
        print("\nPrincipais fatores de risco:")
        for i, factor in enumerate(combined_risk.get("risk_factors", [])[:5]):
            print(f"  {i+1}. [{factor.get('source', '')}] {factor['factor']}: {factor['score']:.2f}")
            if "details" in factor:
                print(f"     - {factor['details']}")
        
        # Exibir recomendações específicas para Moçambique
        if "recommendations" in combined_risk:
            print("\nRecomendações para Moçambique:")
            for i, rec in enumerate(combined_risk["recommendations"]):
                print(f"  {i+1}. {rec}")
        
    except Exception as e:
        logger.error(f"Erro ao executar o exemplo: {str(e)}")


if __name__ == "__main__":
    main()