#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Script de exemplo para demonstração do BrasilBehaviorAgent

Este script exemplifica o uso do agente de análise comportamental
do Brasil para detecção de fraudes contextuais, incluindo:
- Análise de risco de conta
- Detecção de anomalias de localização
- Análise de comportamento de dispositivo
- Obtenção de fatores de risco regionais do Brasil
- Cálculo de score combinado de risco

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
import sys
from datetime import datetime, timedelta
from pprint import pprint

# Configurar o caminho para importar módulos do projeto
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Importar o agente de comportamento do Brasil
from src.api.services.fraud_detection.agents.behavioral.brasil_behavior_agent import BrasilBehaviorAgent

# Configurar logging
logging.basicConfig(level=logging.INFO, 
                    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("exemplo_brasil")

def main():
    """Função principal para demonstração do agente Brasil"""
    try:
        print("=" * 80)
        print("INNOVABIZ IAM/TrustGuard - Demonstração do Agente de Comportamento do Brasil")
        print("=" * 80)
        
        # Criar instância do agente Brasil
        brasil_agent = BrasilBehaviorAgent()
        print(f"Agente inicializado: BrasilBehaviorAgent")
        print("-" * 80)
        
        # 1. Análise de risco de conta
        print("\n1. ANÁLISE DE RISCO DE CONTA (Caso legítimo):")
        
        # Dados de uma conta legítima típica brasileira
        conta_legitima = {
            "creation_date": (datetime.now() - timedelta(days=180)).isoformat(),
            "kyc_status": "verified",
            "suspicious_activities": [],
            "has_valid_bank_account": True,
            "id_verification": {
                "verified": True,
                "method": "facial_biometrics",
                "provider": "serpro"
            },
            "devices": [
                {"id": "device1", "last_used": "2025-08-10T10:30:00"},
                {"id": "device2", "last_used": "2025-08-15T14:20:00"}
            ],
            "cpf_number": "123.456.789-00",
            "address": {
                "city": "São Paulo",
                "state": "SP",
                "neighborhood": "Moema"
            },
            "has_credit_restrictions": False,
            "in_cadin": False,
            "economic_activity": "comercio"
        }
        
        # Analisar conta legítima
        resultado_conta_legitima = brasil_agent.evaluate_account_risk("12345", conta_legitima)
        print(f"Score de risco: {resultado_conta_legitima['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_conta_legitima['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_conta_legitima['risk_factors']))
        for fator in resultado_conta_legitima['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n2. ANÁLISE DE RISCO DE CONTA (Caso suspeito):")
        
        # Dados de uma conta suspeita
        conta_suspeita = {
            "creation_date": (datetime.now() - timedelta(days=5)).isoformat(),
            "kyc_status": "pending",
            "suspicious_activities": [
                {"type": "multiple_logins", "date": "2025-08-18T10:30:00"},
                {"type": "payment_failure", "date": "2025-08-19T14:20:00"}
            ],
            "has_valid_bank_account": False,
            "id_verification": {
                "verified": False,
                "method": "pending",
                "provider": ""
            },
            "devices": [
                {"id": "device1", "last_used": "2025-08-18T10:30:00"},
                {"id": "device2", "last_used": "2025-08-18T10:35:00"},
                {"id": "device3", "last_used": "2025-08-19T08:20:00"},
                {"id": "device4", "last_used": "2025-08-19T23:45:00"}
            ],
            "cpf_number": "",
            "address": {
                "city": "São Paulo",
                "state": "SP",
                "neighborhood": "Paraisópolis"
            },
            "has_credit_restrictions": True,
            "in_cadin": True,
            "economic_activity": "criptomoedas",
            "recent_changes": [
                {"field": "email", "field_type": "sensitive", "date": "2025-08-18T22:30:00"},
                {"field": "phone", "field_type": "sensitive", "date": "2025-08-18T22:35:00"},
            ]
        }
        
        # Analisar conta suspeita
        resultado_conta_suspeita = brasil_agent.evaluate_account_risk("67890", conta_suspeita)
        print(f"Score de risco: {resultado_conta_suspeita['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_conta_suspeita['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_conta_suspeita['risk_factors']))
        for fator in resultado_conta_suspeita['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n" + "-" * 80)
        
        # 2. Análise de anomalias de localização
        print("\n3. DETECÇÃO DE ANOMALIAS DE LOCALIZAÇÃO (Caso legítimo):")
        
        # Dados de localização legítima
        localizacao_legitima = {
            "latitude": -23.5505,
            "longitude": -46.6333,
            "ip_latitude": -23.5550,
            "ip_longitude": -46.6370,
            "timestamp": datetime.now().isoformat(),
            "city": "São Paulo",
            "state": "SP",
            "district": "Jardins",
            "country": "BR",
            "ip_country": "BR"
        }
        
        # Histórico de localizações compatível
        historico_localizacao = [
            {
                "latitude": -23.5500,
                "longitude": -46.6350,
                "timestamp": (datetime.now() - timedelta(days=1)).isoformat()
            },
            {
                "latitude": -23.5480,
                "longitude": -46.6340,
                "timestamp": (datetime.now() - timedelta(hours=12)).isoformat()
            }
        ]
        
        # Analisar localização legítima
        resultado_local_legitimo = brasil_agent.detect_location_anomalies(
            "12345", localizacao_legitima, historico_localizacao
        )
        print(f"Score de risco: {resultado_local_legitimo['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_local_legitimo['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_local_legitimo['risk_factors']))
        for fator in resultado_local_legitimo['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n4. DETECÇÃO DE ANOMALIAS DE LOCALIZAÇÃO (Caso suspeito):")
        
        # Dados de localização suspeita (mudança rápida de lugar)
        localizacao_suspeita = {
            "latitude": -22.9068,
            "longitude": -43.1729,  # Rio de Janeiro
            "ip_latitude": -25.4284,
            "ip_longitude": -49.2733,  # Curitiba
            "timestamp": datetime.now().isoformat(),
            "city": "Rio de Janeiro",
            "state": "RJ",
            "district": "Centro",
            "country": "BR",
            "ip_country": "US"  # IP dos EUA, mas GPS no Brasil
        }
        
        # Histórico recente em São Paulo
        historico_recente = [
            {
                "latitude": -23.5505,
                "longitude": -46.6333,  # São Paulo
                "timestamp": (datetime.now() - timedelta(hours=2)).isoformat()
            }
        ]
        
        # Analisar localização suspeita
        resultado_local_suspeito = brasil_agent.detect_location_anomalies(
            "67890", localizacao_suspeita, historico_recente
        )
        print(f"Score de risco: {resultado_local_suspeito['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_local_suspeito['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_local_suspeito['risk_factors']))
        for fator in resultado_local_suspeito['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n" + "-" * 80)
        
        # 3. Análise de comportamento de dispositivo
        print("\n5. ANÁLISE DE COMPORTAMENTO DE DISPOSITIVO (Caso legítimo):")
        
        # Dispositivo legítimo típico brasileiro
        dispositivo_legitimo = {
            "device_id": "device123",
            "fingerprint": "abcdefg123456",
            "type": "mobile",
            "model": "Samsung Galaxy S22",
            "os": "Android",
            "browser": "Chrome Mobile",
            "user_agent": "Mozilla/5.0 (Linux; Android 12) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.41 Mobile Safari/537.36",
            "is_emulator": False,
            "is_rooted": False,
            "is_vpn": False,
            "is_proxy": False,
            "is_tor": False,
            "connection_type": "mobile",
            "isp": "Vivo",
            "has_fraud_history": False
        }
        
        sessao_legitima = {
            "timestamp": datetime.now().isoformat(),
            "authentication_failures": 0,
            "typing_speed": {"is_anomalous": False},
            "mouse_movement": {"is_natural": True}
        }
        
        # Analisar dispositivo legítimo
        resultado_dispositivo_legitimo = brasil_agent.analyze_device_behavior(
            "12345", dispositivo_legitimo, sessao_legitima
        )
        print(f"Score de risco: {resultado_dispositivo_legitimo['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_dispositivo_legitimo['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_dispositivo_legitimo['risk_factors']))
        for fator in resultado_dispositivo_legitimo['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n6. ANÁLISE DE COMPORTAMENTO DE DISPOSITIVO (Caso suspeito):")
        
        # Dispositivo suspeito
        dispositivo_suspeito = {
            "device_id": "device999",
            "fingerprint": "suspicious123",
            "type": "mobile",
            "model": "Generic Android",
            "os": "Android",
            "browser": "Chrome",
            "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko)",
            "is_emulator": True,
            "is_rooted": True,
            "is_vpn": True,
            "is_proxy": False,
            "is_tor": False,
            "connection_type": "mobile",
            "isp": "Unknown Carrier",
            "has_fraud_history": True,
            "fraud_signals": ["app_cloner", "fake_gps"]
        }
        
        sessao_suspeita = {
            "timestamp": datetime.now().replace(hour=3, minute=15).isoformat(),  # 3:15 da manhã
            "authentication_failures": 4,
            "typing_speed": {"is_anomalous": True},
            "mouse_movement": {"is_natural": False}
        }
        
        # Analisar dispositivo suspeito
        resultado_dispositivo_suspeito = brasil_agent.analyze_device_behavior(
            "67890", dispositivo_suspeito, sessao_suspeita
        )
        print(f"Score de risco: {resultado_dispositivo_suspeito['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_dispositivo_suspeito['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_dispositivo_suspeito['risk_factors']))
        for fator in resultado_dispositivo_suspeito['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n" + "-" * 80)
        
        # 4. Fatores de risco regionais do Brasil
        print("\n7. FATORES DE RISCO REGIONAIS DO BRASIL (Caso legítimo):")
        
        # Entidade legítima
        entidade_legitima = {
            "cpf_number": "123.456.789-00",
            "full_name": "João Silva",
            "email": "joao.silva@exemplo.com.br",
            "phone": "+5511999998888"
        }
        
        # Análise regional - caso legítimo
        # Nota: Em ambiente de produção, isto usaria adaptadores reais para Serasa, Receita Federal, etc.
        resultado_regional_legitimo = brasil_agent.get_regional_risk_factors("12345", entidade_legitima)
        print(f"Score de risco: {resultado_regional_legitimo['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_regional_legitimo['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_regional_legitimo['risk_factors']))
        for fator in resultado_regional_legitimo['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n8. FATORES DE RISCO REGIONAIS DO BRASIL (Caso suspeito):")
        
        # Entidade suspeita - dados insuficientes
        entidade_suspeita = {
            "cpf_number": "",
            "full_name": "Usuario Sem Identificação",
            "email": "temp123@mail.com",
            "phone": ""
        }
        
        # Análise regional - caso suspeito
        resultado_regional_suspeito = brasil_agent.get_regional_risk_factors("67890", entidade_suspeita)
        print(f"Score de risco: {resultado_regional_suspeito['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_regional_suspeito['risk_level'].upper()}")
        print("Fatores de risco identificados:", len(resultado_regional_suspeito['risk_factors']))
        for fator in resultado_regional_suspeito['risk_factors']:
            print(f"  - {fator['description']} (peso: {fator['weight']:.2f})")
        
        print("\n" + "-" * 80)
        
        # 5. Análise combinada
        print("\n9. ANÁLISE COMBINADA DE RISCO (Caso legítimo):")
        
        # Combinar resultados legítimos
        resultados_legitimos = {
            "account_risk": resultado_conta_legitima,
            "location_risk": resultado_local_legitimo,
            "device_risk": resultado_dispositivo_legitimo,
            "regional_risk": resultado_regional_legitimo,
            "transaction_risk": {"risk_score": 0.15, "risk_level": "low"}
        }
        
        # Calcular risco combinado - caso legítimo
        resultado_combinado_legitimo = brasil_agent._calculate_combined_risk_score(resultados_legitimos)
        print(f"Score de risco combinado: {resultado_combinado_legitimo['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_combinado_legitimo['risk_level'].upper()}")
        print("Detalhes de scores individuais:")
        print(f"  - Risco de conta: {resultado_combinado_legitimo['details']['account_risk']:.2f}")
        print(f"  - Risco de localização: {resultado_combinado_legitimo['details']['location_risk']:.2f}")
        print(f"  - Risco de dispositivo: {resultado_combinado_legitimo['details']['device_risk']:.2f}")
        print(f"  - Risco regional: {resultado_combinado_legitimo['details']['regional_risk']:.2f}")
        print(f"  - Risco de transação: {resultado_combinado_legitimo['details']['transaction_risk']:.2f}")
        
        print("\n10. ANÁLISE COMBINADA DE RISCO (Caso suspeito):")
        
        # Combinar resultados suspeitos
        resultados_suspeitos = {
            "account_risk": resultado_conta_suspeita,
            "location_risk": resultado_local_suspeito,
            "device_risk": resultado_dispositivo_suspeito,
            "regional_risk": resultado_regional_suspeito,
            "transaction_risk": {"risk_score": 0.75, "risk_level": "high"}
        }
        
        # Calcular risco combinado - caso suspeito
        resultado_combinado_suspeito = brasil_agent._calculate_combined_risk_score(resultados_suspeitos)
        print(f"Score de risco combinado: {resultado_combinado_suspeito['risk_score']:.2f}")
        print(f"Nível de risco: {resultado_combinado_suspeito['risk_level'].upper()}")
        print("Detalhes de scores individuais:")
        print(f"  - Risco de conta: {resultado_combinado_suspeito['details']['account_risk']:.2f}")
        print(f"  - Risco de localização: {resultado_combinado_suspeito['details']['location_risk']:.2f}")
        print(f"  - Risco de dispositivo: {resultado_combinado_suspeito['details']['device_risk']:.2f}")
        print(f"  - Risco regional: {resultado_combinado_suspeito['details']['regional_risk']:.2f}")
        print(f"  - Risco de transação: {resultado_combinado_suspeito['details']['transaction_risk']:.2f}")
        
        print("\n" + "=" * 80)
        print("Demonstração do agente de análise comportamental do Brasil concluída!")
        print("=" * 80)
        
    except Exception as e:
        logger.error(f"Erro durante a demonstração: {str(e)}")
        import traceback
        traceback.print_exc()
        
if __name__ == "__main__":
    main()