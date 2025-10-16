#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Padrões de Análise Comportamental Específicos para Angola

Este módulo implementa regras e análises de comportamento específicas
para o contexto angolano, considerando padrões regionais de fraude,
comportamentos de usuários e regulamentações locais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import logging
import datetime
import re
from typing import Dict, Any, List, Optional, Union, Set

# Configuração de logging
logger = logging.getLogger("angola_behavior_patterns")


class AngolanBehaviorPatterns:
    """
    Implementação de padrões de comportamento específicos para Angola.
    
    Esta classe implementa regras e análises específicas para o contexto angolano,
    incluindo detecção de fraudes em mobile money, padrões de acesso regional,
    e considerações específicas para o mercado angolano.
    """
    
    def __init__(self, config: Optional[Dict[str, Any]] = None):
        """
        Inicializa o analisador de padrões comportamentais para Angola.
        
        Args:
            config: Configurações específicas (opcional)
        """
        self.config = config or {}
        
        # Zonas de alto risco em Angola (áreas fronteiriças e remotas)
        self.high_risk_zones = {
            # Fronteiras com países vizinhos
            "cabinda",      # Enclave de Cabinda (fronteira com Congo e RDC)
            "zaire",        # Fronteira com RDC
            "uige",         # Fronteira com RDC
            "malanje",      # Província remota
            "lunda norte",  # Fronteira com RDC
            "lunda sul",    # Província remota com atividade diamantífera
            "moxico",       # Fronteira com Zâmbia
            "cuando cubango", # Fronteira com Namíbia e Zâmbia
            "cunene",       # Fronteira com Namíbia
            "namibe"        # Área costeira com baixa densidade
        }
        
        # Zonas urbanas principais (menor risco)
        self.urban_zones = {
            "luanda",      # Capital
            "benguela",    # Segunda maior cidade
            "huambo",      # Terceira maior cidade
            "lubango"      # Huíla
        }
        
        # Operadoras de telecomunicações confiáveis em Angola
        self.trusted_telcos = {
            "unitel",
            "movicel",
            "angola telecom",
            "tv cabo angola",
            "multitel",
            "zap"
        }
        
        # Redes bancárias confiáveis
        self.trusted_banks = {
            "bfa",          # Banco de Fomento Angola
            "bai",          # Banco Angolano de Investimentos
            "bpc",          # Banco de Poupança e Crédito
            "bic",          # Banco BIC
            "banco sol",
            "banco económico",
            "standard bank",
            "banco millennium atlântico",
            "banco bai microfinanças",
            "banco valor",
            "banco keve",
            "banco yetu"
        }
        
        # Padrões de números de telefone angolanos (operadoras principais)
        self.phone_patterns = {
            "unitel": r"^\+244(?:99|91)\d{7}$",
            "movicel": r"^\+244(?:92|93)\d{7}$",
            "angola_telecom": r"^\+244(?:94)\d{7}$"
        }
        
        # Padrões temporais específicos para Angola
        self.temporal_patterns = {
            "business_hours_start": 8,  # 8:00
            "business_hours_end": 17,   # 17:00
            "weekend_days": [5, 6],     # Sábado e domingo
            "holidays": [
                # Feriados nacionais de Angola (variáveis por ano)
                "01-01",   # Ano Novo
                "02-04",   # Dia do Início da Luta Armada
                "04-04",   # Dia da Paz
                "01-05",   # Dia do Trabalhador
                "17-09",   # Dia do Fundador da Nação e do Herói Nacional
                "11-11",   # Dia da Independência
                "25-12"    # Natal
            ]
        }
        
        # Limites específicos para transações em Angola
        self.transaction_limits = {
            "mobile_money": {
                "daily_limit_kwanza": 100000,  # 100.000 Kwanzas
                "single_transaction_limit": 50000,  # 50.000 Kwanzas
                "monthly_limit": 500000,  # 500.000 Kwanzas
                "max_daily_transactions": 10
            },
            "bank_transfer": {
                "daily_limit_kwanza": 1000000,  # 1.000.000 Kwanzas
                "suspicious_amount_threshold": 500000,  # 500.000 Kwanzas
                "international_transfer_threshold": 250000  # 250.000 Kwanzas
            }
        }
        
        # Padrões de fraude conhecidos em Angola
        self.fraud_patterns = {
            "mobile_money_fraud": [
                "multiple_sim_registrations",
                "cross_border_transfers",
                "rapid_cash_in_cash_out",
                "agent_collusion",
                "unusual_province_activity"
            ],
            "identity_fraud": [
                "multiple_accounts_same_device",
                "credential_stuffing",
                "fake_id_creation",
                "government_id_misuse"
            ],
            "banking_fraud": [
                "account_takeover",
                "unusual_international_transfers",
                "diamond_trade_suspicious_patterns",
                "oil_sector_unusual_activity"
            ]
        }
        
        # Dispositivos confiáveis mais comuns em Angola
        self.device_patterns = {
            "common_android_devices": [
                "tecno",
                "infinix", 
                "samsung",
                "huawei",
                "xiaomi"
            ],
            "common_ios_penetration": 0.15,  # 15% estimado iOS vs Android
            "trusted_browser_versions": {
                "chrome": 90,
                "firefox": 85,
                "safari": 14,
                "opera": 75  # Muito usado em Angola com data free
            }
        }
        
        logger.info("Módulo de padrões comportamentais de Angola inicializado")
    
    def analyze_location_risk(self, location_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa o risco baseado na localização em Angola.
        
        Args:
            location_data: Dados de localização do evento
            
        Returns:
            Dict: Análise de risco e insights específicos
        """
        result = {
            "risk_score": 0.5,  # Score padrão médio
            "risk_factors": [],
            "insights": [],
            "is_high_risk_zone": False,
            "is_urban_zone": False
        }
        
        try:
            city = location_data.get("city", "").lower() if location_data.get("city") else ""
            province = location_data.get("province", "").lower() if location_data.get("province") else ""
            country = location_data.get("country_code", "").upper() if location_data.get("country_code") else ""
            
            # Verificar se localização está em Angola
            if country and country != "AO":
                result["risk_score"] = 0.7
                result["risk_factors"].append("access_from_outside_angola")
                result["insights"].append("Acesso de fora de Angola. Verificação adicional recomendada.")
            
            # Verificar se é zona de alto risco
            location_match = False
            for zone in self.high_risk_zones:
                if (zone in city or zone in province):
                    result["risk_score"] = min(result["risk_score"] + 0.2, 1.0)
                    result["risk_factors"].append("high_risk_zone")
                    result["insights"].append(f"Acesso de zona de alto risco ({zone}).")
                    result["is_high_risk_zone"] = True
                    location_match = True
                    break
            
            # Verificar se é zona urbana (menor risco)
            if not location_match:
                for zone in self.urban_zones:
                    if (zone in city or zone in province):
                        result["risk_score"] = max(result["risk_score"] - 0.1, 0.1)
                        result["insights"].append(f"Acesso de zona urbana ({zone}).")
                        result["is_urban_zone"] = True
                        break
            
            # Analisar risco de mudança rápida de localização
            if "previous_location" in location_data:
                prev = location_data["previous_location"]
                current = {
                    "city": city,
                    "province": province,
                    "country_code": country
                }
                
                if self._is_rapid_location_change(prev, current):
                    result["risk_score"] = min(result["risk_score"] + 0.3, 1.0)
                    result["risk_factors"].append("rapid_location_change")
                    result["insights"].append("Mudança rápida de localização detectada.")
        
        except Exception as e:
            logger.error(f"Erro ao analisar risco de localização em Angola: {str(e)}")
            result["insights"].append("Erro ao processar análise de localização.")
        
        return result
    
    def _is_rapid_location_change(self, prev_location: Dict[str, Any], current_location: Dict[str, Any]) -> bool:
        """
        Detecta se houve mudança rápida e improvável de localização.
        
        Args:
            prev_location: Localização anterior
            current_location: Localização atual
            
        Returns:
            bool: True se detectou mudança rápida de localização
        """
        # Simplificação para o exemplo
        if not prev_location or not current_location:
            return False
            
        # Mudança de cidade
        if prev_location.get("city") and current_location.get("city") and \
           prev_location.get("city") != current_location.get("city"):
            # Se mudou de província também, é uma mudança significativa
            if prev_location.get("province") != current_location.get("province"):
                # Em Angola, deslocamento entre províncias é lento, então isso é suspeito
                return True
                
        # Mudança de país
        if prev_location.get("country_code") != current_location.get("country_code"):
            # Qualquer mudança internacional em curto período é suspeita
            return True
                
        return False    def analyze_mobile_money_behavior(self, transaction_data: Dict[str, Any], user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa padrões comportamentais específicos para Mobile Money em Angola.
        
        Args:
            transaction_data: Dados da transação atual
            user_history: Histórico do usuário
            
        Returns:
            Dict: Análise de risco e insights específicos
        """
        result = {
            "risk_score": 0.3,  # Score inicial baixo
            "risk_factors": [],
            "insights": [],
            "recommendation": "allow"
        }
        
        try:
            # Extrair dados relevantes
            transaction_amount = transaction_data.get("amount", 0)
            transaction_type = transaction_data.get("transaction_type", "").lower()
            recipient_id = transaction_data.get("recipient_id")
            agent_id = transaction_data.get("agent_id")
            
            # Histórico do usuário
            daily_transactions = user_history.get("daily_transactions", [])
            daily_volume = user_history.get("daily_volume", 0)
            monthly_volume = user_history.get("monthly_volume", 0)
            known_recipients = user_history.get("known_recipients", set())
            
            # 1. Verificar limites diários de transação
            if daily_volume + transaction_amount > self.transaction_limits["mobile_money"]["daily_limit_kwanza"]:
                result["risk_score"] = min(result["risk_score"] + 0.3, 1.0)
                result["risk_factors"].append("exceeded_daily_limit")
                result["insights"].append("Excedido limite diário de transações.")
                result["recommendation"] = "review"
            
            # 2. Verificar transação única acima do limite
            if transaction_amount > self.transaction_limits["mobile_money"]["single_transaction_limit"]:
                result["risk_score"] = min(result["risk_score"] + 0.25, 1.0)
                result["risk_factors"].append("large_single_transaction")
                result["insights"].append("Transação única de valor elevado.")
            
            # 3. Verificar número de transações diárias
            if len(daily_transactions) >= self.transaction_limits["mobile_money"]["max_daily_transactions"]:
                result["risk_score"] = min(result["risk_score"] + 0.2, 1.0)
                result["risk_factors"].append("high_transaction_frequency")
                result["insights"].append("Frequência elevada de transações no dia.")
                
            # 4. Padrões de fraude conhecidos em Mobile Money em Angola
            
            # 4.1 Verificar cash-in/cash-out rápido (popular em fraudes em Angola)
            if transaction_type == "cash_out" and "recent_cash_in" in user_history:
                time_diff = (transaction_data.get("timestamp", datetime.datetime.now()) - 
                             user_history["recent_cash_in"].get("timestamp", datetime.datetime.now()))
                
                # Se fez cash-out em menos de 10 minutos após cash-in
                if time_diff.total_seconds() < 600:  # 10 minutos
                    result["risk_score"] = min(result["risk_score"] + 0.4, 1.0)
                    result["risk_factors"].append("rapid_cash_in_cash_out")
                    result["insights"].append("Padrão suspeito: cash-out imediatamente após cash-in.")
                    result["recommendation"] = "review"
            
            # 4.2 Verificar mesmo agente para cash-in e cash-out
            if transaction_type == "cash_out" and agent_id and \
               user_history.get("recent_cash_in", {}).get("agent_id") == agent_id:
                result["risk_score"] = min(result["risk_score"] + 0.35, 1.0)
                result["risk_factors"].append("same_agent_cash_in_out")
                result["insights"].append("Mesmo agente usado para cash-in e cash-out.")
                result["recommendation"] = "review"
            
            # 4.3 Verificar destinatário desconhecido com valor alto
            if transaction_type == "transfer" and recipient_id and \
               recipient_id not in known_recipients and \
               transaction_amount > (self.transaction_limits["mobile_money"]["single_transaction_limit"] * 0.7):
                result["risk_score"] = min(result["risk_score"] + 0.3, 1.0)
                result["risk_factors"].append("high_value_to_unknown_recipient")
                result["insights"].append("Transação de valor alto para destinatário desconhecido.")
            
            # 5. Decisão final baseada no score de risco
            if result["risk_score"] >= 0.8:
                result["recommendation"] = "block"
            elif result["risk_score"] >= 0.6:
                result["recommendation"] = "review"
            
        except Exception as e:
            logger.error(f"Erro ao analisar comportamento de Mobile Money em Angola: {str(e)}")
            result["insights"].append("Erro ao processar análise de Mobile Money.")
        
        return result
    
    def validate_angola_phone(self, phone_number: str) -> Dict[str, Any]:
        """
        Valida e identifica operadora de número de telefone angolano.
        
        Args:
            phone_number: Número de telefone
            
        Returns:
            Dict: Resultado da validação
        """
        result = {
            "is_valid": False,
            "operator": None,
            "insights": []
        }
        
        try:
            # Normalizar número
            phone = phone_number.strip()
            if phone.startswith("00244"):
                phone = "+244" + phone[5:]
            elif phone.startswith("244"):
                phone = "+" + phone
                
            # Verificar operadora
            for operator, pattern in self.phone_patterns.items():
                if re.match(pattern, phone):
                    result["is_valid"] = True
                    result["operator"] = operator
                    result["insights"].append(f"Número válido da operadora {operator}.")
                    break
                    
            if not result["is_valid"]:
                result["insights"].append("Número não corresponde aos padrões das operadoras angolanas.")
                
        except Exception as e:
            logger.error(f"Erro ao validar número de telefone angolano: {str(e)}")
            result["insights"].append("Erro ao processar validação de número.")
            
        return result
    
    def analyze_device_context(self, device_data: Dict[str, Any], user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa contexto do dispositivo específico para Angola.
        
        Args:
            device_data: Dados do dispositivo
            user_history: Histórico do usuário
            
        Returns:
            Dict: Análise de risco e insights específicos
        """
        result = {
            "risk_score": 0.3,  # Score inicial baixo
            "risk_factors": [],
            "insights": [],
            "is_common_device": False
        }
        
        try:
            # Extrair dados relevantes
            device_model = device_data.get("model", "").lower()
            device_os = device_data.get("os", "").lower()
            browser = device_data.get("browser", {})
            browser_name = browser.get("name", "").lower()
            browser_version = browser.get("version", 0)
            
            # Verificar se é um dispositivo comum em Angola
            device_brand = None
            for brand in self.device_patterns["common_android_devices"]:
                if brand in device_model:
                    device_brand = brand
                    result["is_common_device"] = True
                    result["insights"].append(f"Dispositivo comum em Angola ({brand}).")
                    break
                    
            # Verificar proporção iOS vs. Android (iOS menos comum em Angola)
            if "ios" in device_os:
                # Se histórico mostra Android mas agora é iOS, isso é incomum
                if user_history.get("usual_device_os") == "android":
                    result["risk_score"] = min(result["risk_score"] + 0.2, 1.0)
                    result["risk_factors"].append("os_platform_change")
                    result["insights"].append("Mudança de Android para iOS é incomum.")
            
            # Verificar versão do navegador (versões muito antigas são comuns em Angola)
            if browser_name in self.device_patterns["trusted_browser_versions"]:
                min_version = self.device_patterns["trusted_browser_versions"][browser_name]
                if browser_version < min_version:
                    # Em Angola, navegadores antigos são comuns devido a limitações de dispositivos
                    result["insights"].append(f"Navegador {browser_name} em versão antiga ({browser_version}).")
                    # Não aumentamos o risco, pois é padrão regional
                    
            # Verificar rede de operadora
            network = device_data.get("network", "").lower()
            if network:
                is_trusted_network = False
                for telco in self.trusted_telcos:
                    if telco in network:
                        is_trusted_network = True
                        result["insights"].append(f"Rede confiável detectada ({telco}).")
                        break
                        
                if not is_trusted_network:
                    result["risk_score"] = min(result["risk_score"] + 0.15, 1.0)
                    result["risk_factors"].append("unknown_network")
                    result["insights"].append("Rede de conexão não reconhecida.")
        
        except Exception as e:
            logger.error(f"Erro ao analisar contexto de dispositivo em Angola: {str(e)}")
            result["insights"].append("Erro ao processar análise de dispositivo.")
        
        return result
    
    def analyze_bureau_credito_integration(self, user_data: Dict[str, Any], bureau_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa dados do Bureau de Crédito de Angola para complementar avaliação de risco.
        
        Args:
            user_data: Dados do usuário
            bureau_data: Dados do bureau de crédito
            
        Returns:
            Dict: Análise de risco creditício e insights
        """
        result = {
            "risk_score": 0.5,  # Score médio inicial
            "risk_factors": [],
            "insights": [],
            "credit_status": "unknown"
        }
        
        try:
            # Verificar se há dados do bureau
            if not bureau_data or not isinstance(bureau_data, dict):
                result["insights"].append("Dados do Bureau de Crédito não disponíveis.")
                return result
                
            # Extrair dados relevantes
            credit_score = bureau_data.get("credit_score", 0)
            payment_defaults = bureau_data.get("payment_defaults", 0)
            active_loans = bureau_data.get("active_loans", 0)
            credit_inquiries = bureau_data.get("recent_inquiries", 0)
            
            # Avaliação baseada no score de crédito
            if credit_score > 0:
                # Score de crédito em Angola normalmente vai de 0-1000
                normalized_score = credit_score / 1000  # Normalizar para 0-1
                
                if normalized_score < 0.3:
                    result["risk_score"] = 0.8
                    result["risk_factors"].append("low_credit_score")
                    result["insights"].append("Score de crédito baixo.")
                    result["credit_status"] = "bad"
                elif normalized_score < 0.6:
                    result["risk_score"] = 0.5
                    result["insights"].append("Score de crédito médio.")
                    result["credit_status"] = "medium"
                else:
                    result["risk_score"] = 0.2
                    result["insights"].append("Score de crédito bom.")
                    result["credit_status"] = "good"
            
            # Verificar inadimplências
            if payment_defaults > 0:
                result["risk_score"] = min(result["risk_score"] + 0.2 * payment_defaults, 1.0)
                result["risk_factors"].append("payment_defaults")
                result["insights"].append(f"Histórico com {payment_defaults} inadimplências.")
            
            # Verificar empréstimos ativos
            if active_loans > 3:
                result["risk_score"] = min(result["risk_score"] + 0.15, 1.0)
                result["risk_factors"].append("multiple_active_loans")
                result["insights"].append(f"Usuário com {active_loans} empréstimos ativos.")
            
            # Verificar consultas recentes (possível tentativa de obter múltiplos créditos)
            if credit_inquiries > 5:
                result["risk_score"] = min(result["risk_score"] + 0.2, 1.0)
                result["risk_factors"].append("multiple_credit_inquiries")
                result["insights"].append(f"{credit_inquiries} consultas de crédito recentes.")
                
        except Exception as e:
            logger.error(f"Erro ao analisar dados do Bureau de Crédito de Angola: {str(e)}")
            result["insights"].append("Erro ao processar dados do Bureau de Crédito.")
        
        return result
    
    def get_regional_rules(self) -> Dict[str, Any]:
        """
        Retorna regras específicas para Angola.
        
        Returns:
            Dict: Conjunto de regras regionais
        """
        return {
            "auth_thresholds": {
                "max_failed_attempts": 4,  # Menor que o padrão devido à baixa penetração digital
                "lockout_period_minutes": 20,
                "max_password_changes_per_day": 1,
                "max_device_changes_per_week": 2
            },
            "session_thresholds": {
                "max_concurrent_sessions": 2,  # Menor que o padrão devido a menos dispositivos por usuário
                "max_session_time_hours": 8,  # Jornadas mais curtas
                "max_idle_time_minutes": 20
            },
            "transaction_thresholds": {
                "daily_limit_kwanza": self.transaction_limits["mobile_money"]["daily_limit_kwanza"],
                "single_transaction_limit": self.transaction_limits["mobile_money"]["single_transaction_limit"],
                "monthly_limit": self.transaction_limits["mobile_money"]["monthly_limit"],
                "max_daily_transactions": self.transaction_limits["mobile_money"]["max_daily_transactions"]
            },
            "location_patterns": {
                "high_risk_zones": list(self.high_risk_zones),
                "trusted_locations": list(self.urban_zones),
                "location_change_speed_threshold_kmh": 500  # Infraestrutura de transporte limita velocidades
            },
            "mobile_device_patterns": {
                "primary_device_trust_level": "high",
                "multi_device_threshold": 3,
                "device_change_cooling_period_hours": 48  # Período mais longo devido a menos trocas de dispositivos
            },
            "connection_patterns": {
                "trusted_networks": list(self.trusted_telcos),
                "network_change_threshold_per_day": 2
            },
            "bureau_credit_thresholds": {
                "minimum_credit_score": 300,  # Score mínimo em escala 0-1000
                "max_payment_defaults": 2,
                "max_active_loans": 3
            }
        }


# Função auxiliar para criar instância do analisador
def create_angola_analyzer(config=None):
    """
    Cria uma instância do analisador comportamental para Angola.
    
    Args:
        config: Configuração opcional
        
    Returns:
        AngolanBehaviorPatterns: Instância configurada
    """
    return AngolanBehaviorPatterns(config)