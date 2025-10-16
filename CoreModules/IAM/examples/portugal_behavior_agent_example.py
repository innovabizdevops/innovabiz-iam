#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Script de exemplo para uso do Agente de Análise Comportamental para Portugal

Este script demonstra a utilização do PortugalBehaviorAgent para análise
de fraudes contextuais específicas para o mercado português.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
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
logger = logging.getLogger("example_portugal")

# Importar o agente de comportamento para Portugal
try:
    from src.api.services.fraud_detection.agents.behavioral.portugal_behavior_agent import PortugalBehaviorAgent
except ImportError:
    # Ajustar o caminho conforme a estrutura do projeto
    import sys
    sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))
    from src.api.services.fraud_detection.agents.behavioral.portugal_behavior_agent import PortugalBehaviorAgent


def generate_mock_data() -> Dict[str, Any]:
    """
    Gera dados simulados para teste do agente comportamental
    """
    # Dados de conta
    account_data = {
        "account_id": "PT123456789",
        "user_id": "user_pt_12345",
        "name": "António Silva Fernandes",
        "created_at": (datetime.now() - timedelta(days=45)).isoformat(),
        "kyc_status": "verified",
        "tax_id": "234567891",  # NIF português
        "id_document": {
            "type": "cc",
            "number": "12345678901X",
            "issuer": "República Portuguesa",
            "issue_date": "2020-06-10",
            "expiry_date": "2030-06-10"
        },
        "suspicious_activity_history": [
            {
                "type": "failed_login",
                "timestamp": (datetime.now() - timedelta(days=15)).isoformat(),
                "details": "Múltiplas tentativas de login malsucedidas"
            }
        ],
        "address": {
            "street": "Rua Augusta, 95",
            "neighborhood": "Baixa",
            "city": "Lisboa",
            "district": "Lisboa",
            "postal_code": "1100-048",
            "country": "Portugal",
            "is_temporary": False,
            "differs_from_fiscal_address": False
        },
        "economic_activity": "comércio_varejo",
        "is_pep": False,
        "pep_relatives": []
    }
    
    # Dados de localização
    location_data = {
        "ip_address": "85.243.120.45",  # IP simulado de Portugal
        "country": "PT",
        "city": "Lisboa",
        "district": "Lisboa",
        "latitude": 38.7223,
        "longitude": -9.1393,
        "timestamp": datetime.now().isoformat(),
        "is_vpn": False,
        "is_proxy": False
    }
    
    # Histórico de localizações
    user_location_history = {
        "user_id": "user_pt_12345",
        "recent_locations": [
            {
                "ip_address": "85.243.120.45",
                "country": "PT",
                "city": "Lisboa",
                "timestamp": (datetime.now() - timedelta(hours=12)).isoformat(),
                "coords": {
                    "latitude": 38.7223,
                    "longitude": -9.1393
                }
            },
            {
                "ip_address": "85.243.121.67",
                "country": "PT",
                "city": "Porto",
                "timestamp": (datetime.now() - timedelta(days=2)).isoformat(),
                "coords": {
                    "latitude": 41.1579,
                    "longitude": -8.6291
                }
            }
        ],
        "typical_locations": [
            {
                "city": "Lisboa",
                "district": "Lisboa",
                "frequency": "high",
                "coords": {
                    "latitude": 38.7223,
                    "longitude": -9.1393
                }
            },
            {
                "city": "Porto",
                "district": "Porto",
                "frequency": "medium",
                "coords": {
                    "latitude": 41.1579,
                    "longitude": -8.6291
                }
            }
        ]
    }
    
    # Dados do dispositivo
    device_data = {
        "device_id": "dev-pt-78901234",
        "device_type": "smartphone",
        "platform": "mobile",
        "os": {
            "name": "iOS",
            "version": "15.5"
        },
        "browser": {
            "name": "Safari",
            "version": "15.5"
        },
        "ip_address": "85.243.120.45",
        "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
        "screen_resolution": "1170x2532",
        "is_rooted": False,
        "is_jailbroken": False,
        "is_emulator": False,
        "is_vpn": False,
        "is_proxy": False,
        "is_tor": False,
        "session_start": datetime.now().isoformat(),
        "timezone": "Europe/Lisbon",
        "language": "pt-PT",
        "linked_accounts": ["user_pt_12345"],
        "fingerprint_changed": False
    }
    
    # Histórico de dispositivos
    user_device_history = {
        "user_id": "user_pt_12345",
        "known_devices": [
            {
                "device_id": "dev-pt-78901234",
                "first_seen": (datetime.now() - timedelta(days=45)).isoformat(),
                "last_seen": (datetime.now() - timedelta(hours=8)).isoformat()
            }
        ],
        "last_login": {
            "device_id": "dev-pt-78901234",
            "timestamp": (datetime.now() - timedelta(hours=8)).isoformat(),
            "ip_address": "85.243.120.45"
        },
        "sessions": [
            {
                "login_time": (datetime.now() - timedelta(hours=8)).isoformat(),
                "ip_address": "85.243.120.45",
                "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
                "screen_resolution": "1170x2532",
                "timezone": "Europe/Lisbon",
                "language": "pt-PT",
                "referrer": "https://www.banco.pt"
            },
            {
                "login_time": (datetime.now() - timedelta(days=2)).isoformat(),
                "ip_address": "85.243.121.67",
                "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
                "screen_resolution": "1170x2532",
                "timezone": "Europe/Lisbon",
                "language": "pt-PT",
                "referrer": "https://www.banco.pt"
            }
        ]
    }
    
    # Dados do usuário para análise regional
    user_data = {
        "user_id": "user_pt_12345",
        "name": "António Silva Fernandes",
        "tax_id": "234567891",
        "id_document": {
            "type": "cc",
            "number": "12345678901X"
        },
        "phone": "+351912345678"
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
    """Função principal para demonstração do agente comportamental de Portugal"""
    try:
        logger.info("Inicializando agente de comportamento para Portugal...")
        
        # Inicializar agente de Portugal
        portugal_agent = PortugalBehaviorAgent()
        
        # Gerar dados simulados para teste
        mock_data = generate_mock_data()
        
        logger.info("Executando análises de risco para Portugal...")
        
        # Análise de risco de conta
        account_risk = portugal_agent.evaluate_account_risk(mock_data["account_data"])
        print("\n=== ANÁLISE DE RISCO DE CONTA ===")
        print(f"Pontuação: {account_risk['risk_score']:.2f} ({account_risk['risk_level']})")
        print("Fatores de risco identificados:")
        for factor in account_risk.get("risk_factors", []):
            print(f"  • {factor['factor']}: {factor['score']:.2f}")
            if "details" in factor:
                print(f"    - {factor['details']}")
        
        # Análise de anomalias de localização
        location_risk = portugal_agent.detect_location_anomalies(
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
        device_risk = portugal_agent.analyze_device_behavior(
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
        regional_risk = portugal_agent.get_regional_risk_factors(mock_data["user_data"])
        print("\n=== ANÁLISE DE FATORES DE RISCO REGIONAIS ===")
        print(f"Pontuação: {regional_risk['risk_score']:.2f} ({regional_risk['risk_level']})")
        print("Fatores de risco identificados:")
        for factor in regional_risk.get("risk_factors", []):
            print(f"  • {factor['factor']}: {factor['score']:.2f}")
            if "details" in factor:
                print(f"    - {factor['details']}")
        
        # Calcular pontuação combinada
        combined_risk = portugal_agent._calculate_combined_risk_score(
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
        
    except Exception as e:
        logger.error(f"Erro ao executar o exemplo: {str(e)}")


if __name__ == "__main__":
    main()