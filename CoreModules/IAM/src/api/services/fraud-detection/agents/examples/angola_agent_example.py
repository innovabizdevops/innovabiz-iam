#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Exemplo de Uso do Agente de Análise Comportamental para Angola

Este script demonstra como utilizar o agente de análise comportamental
especializado para o mercado angolano para detectar fraudes contextuais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
import sys
import json
import logging
from datetime import datetime, timedelta
from typing import Dict, List

# Configurar caminho para importação dos módulos
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Importar o agente de Angola
from behavioral.angola_behavior_agent import AngolaBehaviorAgent

# Configuração de logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

logger = logging.getLogger("exemplo_agente_angola")

def carregar_dados_exemplo(arquivo: str) -> Dict:
    """Carrega dados de exemplo de um arquivo JSON"""
    try:
        caminho = os.path.join(os.path.dirname(os.path.abspath(__file__)), "dados", arquivo)
        with open(caminho, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        logger.error(f"Erro ao carregar dados de exemplo: {str(e)}")
        # Retornar dados mínimos se não conseguir carregar o arquivo
        return {}

def exemplo_analise_transacao():
    """Exemplo de análise de uma transação financeira"""
    logger.info("=== EXEMPLO: ANÁLISE DE TRANSAÇÃO FINANCEIRA ===")
    
    # Configuração do agente
    config = {
        "cache_dir": "./cache/angola",
        "model_path": "./models",
        "data_sources": ["bureau_credito", "telecom", "bna"]
    }
    
    # Criar diretório de cache se não existir
    os.makedirs(config["cache_dir"], exist_ok=True)
    
    # Inicializar o agente
    agente = AngolaBehaviorAgent(
        config_path=None,
        model_path=config["model_path"],
        cache_dir=config["cache_dir"],
        data_sources=config["data_sources"]
    )
    
    # Carregar dados de exemplo
    entity_id = "AO123456789"
    transaction_data = {
        "transaction_details": {
            "amount": 125000,
            "timestamp": datetime.now().isoformat(),
            "type": "payment",
            "description": "Pagamento de serviços",
            "merchant": "Shoprite Angola",
            "payment_method": "mobile_money",
            "currency": "AOA"
        },
        "device_data": {
            "device_id": "d789xyz",
            "is_known": False,
            "device_type": "mobile",
            "os": "Android",
            "browser": "Chrome Mobile",
            "is_emulator": False,
            "is_rooted": False,
            "is_vpn": True
        },
        "location_data": {
            "country": "AO",
            "city": "Luanda",
            "district": "Sambizanga",
            "latitude": -8.8368,
            "longitude": 13.2343,
            "ip_latitude": -8.8120,
            "ip_longitude": 13.2420,
            "timestamp": datetime.now().isoformat()
        }
    }
    
    # Contexto adicional
    context_data = {
        "session_data": {
            "duration_minutes": 45,
            "concurrent_sessions": 1,
            "recent_auth_failures": 0,
            "authentication_method": "password"
        },
        "account_history": {
            "creation_date": (datetime.now() - timedelta(days=75)).isoformat(),
            "kyc_status": "verified",
            "suspicious_activities": [],
            "has_valid_bank_account": True,
            "id_verification": {"verified": True},
            "devices": ["d123abc", "d456def"]
        },
        "location_history": [
            {
                "latitude": -8.8368,
                "longitude": 13.2343,
                "timestamp": (datetime.now() - timedelta(hours=5)).isoformat()
            },
            {
                "latitude": -8.8390,
                "longitude": 13.2350,
                "timestamp": (datetime.now() - timedelta(hours=12)).isoformat()
            }
        ],
        "transaction_history": [
            {
                "amount": 5000,
                "timestamp": (datetime.now() - timedelta(days=3)).isoformat(),
                "type": "payment",
                "description": "Recarrega telefonica"
            },
            {
                "amount": 8500,
                "timestamp": (datetime.now() - timedelta(days=5)).isoformat(),
                "type": "payment",
                "description": "Supermercado"
            },
            {
                "amount": 12000,
                "timestamp": (datetime.now() - timedelta(days=10)).isoformat(),
                "type": "payment",
                "description": "Combustível"
            }
        ]
    }
    
    # Realizar análise
    logger.info(f"Analisando comportamento da entidade {entity_id}...")
    resultado = agente.analyze_behavior(
        entity_id=entity_id,
        entity_type="user",
        transaction_data=transaction_data,
        context_data=context_data,
        use_cache=False
    )
    
    # Exibir resultado
    logger.info(f"Análise concluída com score de risco: {resultado['risk_score']:.2f}")
    logger.info(f"Nível de risco: {resultado['risk_level']}")
    logger.info(f"Ação recomendada: {resultado['recommended_action']}")
    
    if resultado['risk_factors']:
        logger.info("Fatores de risco identificados:")
        for i, fator in enumerate(resultado['risk_factors'], 1):
            logger.info(f"  {i}. {fator['description']}")
    else:
        logger.info("Nenhum fator de risco significativo identificado.")
    
    # Verificar métricas
    metricas = agente.get_metrics()
    logger.info(f"Tempo de execução: {metricas['last_execution_time']:.3f} segundos")
    
    return resultado

def exemplo_analise_dispositivo_suspeito():
    """Exemplo de análise de um dispositivo potencialmente suspeito"""
    logger.info("\n=== EXEMPLO: ANÁLISE DE DISPOSITIVO SUSPEITO ===")
    
    # Inicializar o agente com configuração padrão
    agente = AngolaBehaviorAgent(
        cache_dir="./cache/angola",
        data_sources=["telecom"]
    )
    
    # Dados de dispositivo com características suspeitas
    entity_id = "AO987654321"
    device_data = {
        "device_id": "emu123456",
        "is_known": False,
        "device_type": "mobile",
        "os": "Android",
        "browser": "Firefox Mobile",
        "is_emulator": True,
        "is_rooted": True,
        "is_vpn": True,
        "is_tor": False,
        "is_proxy": True,
        "is_public_network": True,
        "device_model": "Generic Android",
        "ip_address": "196.223.42.X"
    }
    
    # Dados de sessão
    session_data = {
        "duration_minutes": 120,  # Sessão anormalmente longa
        "concurrent_sessions": 2,  # Sessões simultâneas
        "recent_auth_failures": 4,  # Múltiplas falhas recentes
        "authentication_method": "password",
        "login_time": datetime.now().isoformat(),
        "unusual_activity": True
    }
    
    # Analisar dispositivo
    logger.info(f"Analisando dispositivo para entidade {entity_id}...")
    resultado = agente.analyze_device_behavior(
        entity_id=entity_id,
        device_data=device_data,
        session_data=session_data
    )
    
    # Exibir resultado
    logger.info(f"Análise concluída com score de risco: {resultado['risk_score']:.2f}")
    logger.info(f"Nível de risco: {resultado['risk_level']}")
    
    if resultado['risk_factors']:
        logger.info("Fatores de risco identificados:")
        for i, fator in enumerate(resultado['risk_factors'], 1):
            logger.info(f"  {i}. {fator['description']}")
    else:
        logger.info("Nenhum fator de risco significativo identificado.")
    
    return resultado

def exemplo_deteccao_anomalia_localizacao():
    """Exemplo de detecção de anomalias de localização"""
    logger.info("\n=== EXEMPLO: DETECÇÃO DE ANOMALIAS DE LOCALIZAÇÃO ===")
    
    # Inicializar o agente
    agente = AngolaBehaviorAgent(
        cache_dir="./cache/angola"
    )
    
    # Dados de localização atual (suspeita - mudança rápida de localização)
    entity_id = "AO123456789"
    location_data = {
        "country": "AO",
        "city": "Luanda",
        "district": "Viana",
        "latitude": -8.9125,
        "longitude": 13.3604,
        "ip_latitude": -1.2921,    # IP mostra Gabão
        "ip_longitude": 9.4521,    # IP mostra Gabão
        "timestamp": datetime.now().isoformat()
    }
    
    # Histórico de localizações (todas em Luanda)
    location_history = [
        {
            "latitude": -8.8368,
            "longitude": 13.2343,
            "timestamp": (datetime.now() - timedelta(hours=1)).isoformat()
        },
        {
            "latitude": -8.8390,
            "longitude": 13.2350,
            "timestamp": (datetime.now() - timedelta(hours=5)).isoformat()
        },
        {
            "latitude": -8.8415,
            "longitude": 13.2330,
            "timestamp": (datetime.now() - timedelta(hours=12)).isoformat()
        }
    ]
    
    # Analisar localização
    logger.info(f"Analisando localização para entidade {entity_id}...")
    resultado = agente.detect_location_anomalies(
        entity_id=entity_id,
        location_data=location_data,
        history=location_history
    )
    
    # Exibir resultado
    logger.info(f"Análise concluída com score de risco: {resultado['risk_score']:.2f}")
    logger.info(f"Nível de risco: {resultado['risk_level']}")
    
    if resultado['risk_factors']:
        logger.info("Fatores de risco identificados:")
        for i, fator in enumerate(resultado['risk_factors'], 1):
            logger.info(f"  {i}. {fator['description']}")
    else:
        logger.info("Nenhum fator de risco significativo identificado.")
    
    return resultado

def exemplo_avaliacao_risco_conta():
    """Exemplo de avaliação de risco de uma conta"""
    logger.info("\n=== EXEMPLO: AVALIAÇÃO DE RISCO DE CONTA ===")
    
    # Inicializar o agente
    agente = AngolaBehaviorAgent(
        cache_dir="./cache/angola"
    )
    
    # Dados de conta com características de alto risco
    entity_id = "AO876543210"
    account_data = {
        "creation_date": (datetime.now() - timedelta(days=12)).isoformat(),  # Conta recente
        "kyc_status": "pending",  # KYC incompleto
        "suspicious_activities": [
            {"type": "failed_login", "timestamp": (datetime.now() - timedelta(days=5)).isoformat()},
            {"type": "unusual_access", "timestamp": (datetime.now() - timedelta(days=3)).isoformat()}
        ],
        "has_valid_bank_account": False,  # Sem conta bancária verificada
        "id_verification": {"verified": False},  # ID não verificado
        "recent_changes": [
            {"field": "email", "field_type": "sensitive", "timestamp": (datetime.now() - timedelta(days=1)).isoformat()},
            {"field": "phone", "field_type": "sensitive", "timestamp": (datetime.now() - timedelta(days=1)).isoformat()}
        ],
        "devices": ["d111aaa", "d222bbb", "d333ccc", "d444ddd"],  # Muitos dispositivos
        "economic_activity": "cambista informal",  # Atividade de alto risco
        "origin_country": "ZA"  # Origem na África do Sul, não em Angola
    }
    
    # Analisar conta
    logger.info(f"Avaliando risco da conta {entity_id}...")
    resultado = agente.evaluate_account_risk(
        entity_id=entity_id,
        account_data=account_data
    )
    
    # Exibir resultado
    logger.info(f"Análise concluída com score de risco: {resultado['risk_score']:.2f}")
    logger.info(f"Nível de risco: {resultado['risk_level']}")
    
    if resultado['risk_factors']:
        logger.info("Fatores de risco identificados:")
        for i, fator in enumerate(resultado['risk_factors'], 1):
            logger.info(f"  {i}. {fator['description']}")
    else:
        logger.info("Nenhum fator de risco significativo identificado.")
    
    return resultado

def main():
    """Função principal que executa todos os exemplos"""
    logger.info("Iniciando exemplos de uso do Agente de Análise Comportamental para Angola")
    
    # Criar diretório para dados de exemplo (placeholder)
    os.makedirs(os.path.join(os.path.dirname(os.path.abspath(__file__)), "dados"), exist_ok=True)
    
    # Executar exemplos
    try:
        exemplo_analise_transacao()
        exemplo_analise_dispositivo_suspeito()
        exemplo_deteccao_anomalia_localizacao()
        exemplo_avaliacao_risco_conta()
        
        logger.info("\n=== TODOS OS EXEMPLOS EXECUTADOS COM SUCESSO ===")
    except Exception as e:
        logger.error(f"Erro ao executar exemplos: {str(e)}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()