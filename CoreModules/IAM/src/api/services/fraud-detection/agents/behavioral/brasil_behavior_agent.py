#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Agente de Análise Comportamental para Brasil

Este módulo implementa um agente especializado de detecção de fraudes
com adaptações específicas para o mercado brasileiro, considerando
padrões comportamentais, regulamentações, características culturais
e dinâmicas econômicas específicas do Brasil.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
import hashlib
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Union

# Importar classe base de análise comportamental
from .behavior_analysis_framework import BehaviorAnalysisAgent

# Configurar logger
logger = logging.getLogger("fraud_detection.behavioral.brasil")

class BrasilBehaviorAgent(BehaviorAnalysisAgent):
    """
    Agente de análise comportamental especializado para o mercado brasileiro.
    
    Esta classe implementa análise de comportamento para detecção de fraudes
    considerando fatores específicos do Brasil, incluindo:
    - Regulamentações do Banco Central do Brasil (BACEN)
    - Padrões comportamentais típicos em transações bancárias/financeiras brasileiras
    - Características de uso de dispositivos no Brasil (predominância mobile)
    - Dados do Serasa/Boa Vista/SPC e outros bureaus de crédito locais
    - Validação de documentos brasileiros (CPF, CNPJ, RG)
    - Considerações geográficas específicas do Brasil
    
    Implementa todos os métodos abstratos da classe BehaviorAnalysisAgent
    com adaptações específicas para o contexto brasileiro.
    """
    
    def __init__(self, config_path: Optional[str] = None, 
                model_path: Optional[str] = None,
                cache_dir: Optional[str] = None,
                data_sources: Optional[List[str]] = None):
        """
        Inicializa o agente de comportamento brasileiro.
        
        Args:
            config_path: Caminho para arquivo de configuração
            model_path: Caminho para modelos treinados
            cache_dir: Diretório para armazenamento de cache
            data_sources: Lista de fontes de dados a utilizar
        """
        # Chamar inicialização da classe pai
        super().__init__(config_path, model_path, cache_dir, data_sources)
        
        # Definir região para Brasil
        self.region = "BR"
        
        # Carregar configurações específicas do Brasil se não fornecido
        if not config_path:
            default_config = os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "config",
                "brasil_config.json"
            )
            if os.path.exists(default_config):
                self.config_path = default_config
        
        # Inicializar adaptadores específicos para o Brasil
        self._init_brasil_adapters()
        
        # Carregar padrões regionais para o Brasil
        self._load_brasil_patterns()
        
        # Fatores de risco específicos do Brasil
        self.regional_risk_factors = {
            "new_device_login": 0.65,
            "multiple_auth_failures": 0.75,
            "unusual_transaction_time": 0.7,
            "multiple_card_registrations": 0.8,
            "foreign_ip_access": 0.75,
            "high_risk_state": 0.65,
            "cpf_blacklist": 0.9,
            "serasa_restrictions": 0.8,
            "cadin_presence": 0.85,
            "device_fraud_history": 0.85
        }
        
        # Inicializar modelos específicos para o Brasil
        self._load_brasil_models()
        
        logger.info(f"Agente de análise comportamental do Brasil inicializado. Versão: 1.0.0")
    
    def _init_brasil_adapters(self):
        """Inicializa adaptadores de dados específicos para o Brasil"""
        try:
            # Lista de adaptadores a serem inicializados
            brasil_adapters = {
                "serasa": "SerasaAdapter",
                "receita_federal": "ReceitaFederalAdapter",
                "bacen": "BacenAdapter",
                "telecom_brasil": "TelecomBrasilAdapter"
            }
            
            # Inicializar adaptadores selecionados ou todos se nenhum foi especificado
            adapter_names = self.config.get("data_sources", list(brasil_adapters.keys()))
            
            for adapter_name in adapter_names:
                if adapter_name in brasil_adapters:
                    try:
                        # Caminho dinâmico para importação dos adaptadores
                        adapter_module = f"...adapters.brasil.{adapter_name.lower()}_adapter"
                        adapter_class = brasil_adapters[adapter_name]
                        
                        # Importar dinamicamente o adaptador
                        try:
                            # Tentar importação relativa
                            module = __import__(adapter_module, fromlist=[adapter_class])
                            adapter_cls = getattr(module, adapter_class)
                            
                            # Inicializar adaptador
                            self.data_adapters[adapter_name] = adapter_cls(
                                config=self.config.get(f"{adapter_name}_config", {}),
                                cache_dir=os.path.join(self.cache_dir, adapter_name) if self.cache_dir else None
                            )
                            logger.info(f"Adaptador {adapter_name} inicializado com sucesso")
                        except ImportError:
                            logger.warning(f"Não foi possível importar o adaptador {adapter_name}. Usando mock.")
                            # Usar adaptador mock se o real não estiver disponível
                            from ...adapters.mock_adapter import MockAdapter
                            self.data_adapters[adapter_name] = MockAdapter(adapter_name, self.region)
                    except Exception as e:
                        logger.error(f"Erro ao inicializar adaptador {adapter_name}: {str(e)}")
                        
            # Verificar se temos adaptadores suficientes
            if not self.data_adapters:
                logger.warning("Nenhum adaptador de dados foi inicializado para o Brasil")
                
        except Exception as e:
            logger.error(f"Falha na inicialização dos adaptadores para o Brasil: {str(e)}")
    
    def _load_brasil_patterns(self):
        """Carrega padrões comportamentais específicos do Brasil"""
        try:
            # Tentar carregar padrões de um arquivo
            patterns_path = os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "patterns",
                "brasil_patterns.json"
            )
            
            if os.path.exists(patterns_path):
                with open(patterns_path, 'r', encoding='utf-8') as f:
                    self.regional_patterns = json.load(f)
                logger.info(f"Padrões regionais do Brasil carregados de {patterns_path}")
            else:
                # Definir padrões padrão se o arquivo não existe
                self.regional_patterns = {
                    "transaction_patterns": {
                        "typical_transaction_amount": {
                            "p2p_transfer": {
                                "mean": 350.00,
                                "std_dev": 250.00,
                                "max_normal": 2000.00
                            },
                            "bill_payment": {
                                "mean": 220.00,
                                "std_dev": 300.00,
                                "max_normal": 1500.00
                            },
                            "retail_purchase": {
                                "mean": 180.00,
                                "std_dev": 150.00,
                                "max_normal": 1000.00
                            }
                        },
                        "peak_transaction_hours": [10, 12, 15, 19],
                        "low_activity_hours": [0, 1, 2, 3, 4],
                        "weekend_usage_factor": 0.7,
                        "month_end_increase_factor": 1.5,
                        "common_transaction_frequencies": {
                            "daily": 1,
                            "weekly": 5,
                            "monthly": 20
                        },
                        "high_risk_merchants_categories": [
                            "jogos_online",
                            "apostas",
                            "moeda_virtual",
                            "serviços_internacionais_não_identificados"
                        ],
                        "high_risk_regions": [
                            "fronteira_paraguay",
                            "fronteira_bolivia",
                            "tríplice_fronteira"
                        ],
                        "pix_specific_patterns": {
                            "typical_frequency_daily": 3,
                            "max_normal_amount": 5000.00,
                            "suspicious_time_gap_seconds": 30
                        }
                    },
                    "behavioral_patterns": {
                        "device_usage": {
                            "mobile_predominance": 0.75,  # Brasil tem alta penetração de mobile
                            "typical_session_duration_min": 8,
                            "common_device_change_frequency_days": 180,
                            "max_normal_devices_per_user": 3,
                            "typical_auth_methods": ["senha", "biometria", "codigo_sms"]
                        },
                        "login_patterns": {
                            "typical_login_frequency_days": 3,
                            "typical_login_hours": [8, 22],
                            "suspicious_login_attempts_threshold": 3
                        }
                    },
                    "location_patterns": {
                        "high_risk_areas": [
                            "Paraisópolis", "Heliópolis", "Capão Redondo", "Brasilândia", 
                            "Complexo do Alemão", "Rocinha", "Cidade de Deus",
                            "Santa Cruz", "Maré", "Itaquera", "Capão Redondo"
                        ],
                        "common_movement_radius_km": 30,
                        "typical_speed_kmh": 60,
                        "state_risk_factors": {
                            "RO": 0.6, "AC": 0.65, "AM": 0.55, "RR": 0.7, "PA": 0.6,
                            "AP": 0.65, "TO": 0.55, "MA": 0.6, "PI": 0.5, "CE": 0.5,
                            "RN": 0.5, "PB": 0.5, "PE": 0.55, "AL": 0.6, "SE": 0.55,
                            "BA": 0.5, "MG": 0.4, "ES": 0.45, "RJ": 0.65, "SP": 0.45,
                            "PR": 0.4, "SC": 0.35, "RS": 0.4, "MS": 0.5, "MT": 0.55,
                            "GO": 0.45, "DF": 0.4
                        },
                        "border_areas_risk": 0.7
                    }
                }
                logger.warning(f"Arquivo de padrões para o Brasil não encontrado. Usando padrões padrão.")
                
        except Exception as e:
            logger.error(f"Erro ao carregar padrões regionais do Brasil: {str(e)}")
            # Definir padrões mínimos em caso de erro
            self.regional_patterns = {
                "transaction_patterns": {"typical_amount": 200.0},
                "behavioral_patterns": {"device_usage": {"mobile_predominance": 0.7}},
                "location_patterns": {"high_risk_areas": []}
            }
    
    def _load_brasil_models(self):
        """Carrega modelos de ML específicos para o Brasil"""
        try:
            # Verificar diretório de modelos
            if not self.model_path:
                logger.warning("Caminho de modelos não definido. Usando heurísticas.")
                return
            
            # Definir caminho específico para modelos do Brasil
            brasil_models_path = os.path.join(self.model_path, "brasil")
            
            # Verificar e carregar modelos específicos (quando disponíveis)
            model_files = {
                "transaction_risk": "brasil_transaction_risk_model.pkl",
                "account_risk": "brasil_account_risk_model.pkl",
                "location_risk": "brasil_location_anomaly_model.pkl",
                "device_risk": "brasil_device_behavior_model.pkl"
            }
            
            # Carregar modelos disponíveis
            for model_type, model_file in model_files.items():
                model_path = os.path.join(brasil_models_path, model_file)
                if os.path.exists(model_path):
                    try:
                        # Aqui usaria uma função para carregar o modelo de acordo com seu tipo
                        # self.models[model_type] = load_model(model_path)
                        logger.info(f"Modelo {model_type} para Brasil carregado com sucesso")
                    except Exception as e:
                        logger.error(f"Erro ao carregar modelo {model_type}: {str(e)}")
                else:
                    logger.warning(f"Modelo {model_type} não encontrado. Usando regras heurísticas.")
                    
        except Exception as e:
            logger.error(f"Erro ao carregar modelos do Brasil: {str(e)}")
    
    def analyze_transaction_pattern(self, entity_id: str, transaction_data: Dict) -> Dict:
        """
        Analisa padrões de transação para detectar anomalias no contexto brasileiro.
        
        Args:
            entity_id: ID da entidade
            transaction_data: Dados da transação
            
        Returns:
            Resultado da análise de transação
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Extrair detalhes da transação
            tx_details = transaction_data.get("transaction_details", {})
            
            if not tx_details:
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Dados insuficientes para análise"}
                }
            
            # Obter valores relevantes
            tx_amount = float(tx_details.get("amount", 0))
            tx_type = tx_details.get("type", "unknown").lower()
            tx_timestamp = tx_details.get("timestamp")
            tx_description = tx_details.get("description", "").lower()
            tx_merchant = tx_details.get("merchant", "").lower()
            tx_payment_method = tx_details.get("payment_method", "").lower()
            tx_currency = tx_details.get("currency", "BRL")
            
            # Converter timestamp para datetime
            if tx_timestamp:
                if isinstance(tx_timestamp, str):
                    tx_datetime = datetime.fromisoformat(tx_timestamp)
                else:
                    tx_datetime = tx_timestamp
            else:
                tx_datetime = datetime.now()
            
            # Obter padrões de transação para o Brasil
            brasil_patterns = self.regional_patterns["transaction_patterns"]
            
            # 1. Verificar valor anômalo da transação
            # Mapear tipo de transação para categoria padrão
            tx_category = "retail_purchase"  # Padrão
            if "transfer" in tx_type or "pix" in tx_type:
                tx_category = "p2p_transfer"
            elif "bill" in tx_type or "payment" in tx_type or "boleto" in tx_type:
                tx_category = "bill_payment"
            
            # Obter limites para esta categoria
            category_limits = brasil_patterns["typical_transaction_amount"].get(
                tx_category, 
                {"mean": 200.0, "std_dev": 150.0, "max_normal": 1000.0}
            )
            
            # Verificar se o valor é anômalo
            if tx_amount > category_limits["max_normal"]:
                # Calcular quão anômalo é o valor (z-score aproximado)
                anomaly_factor = (tx_amount - category_limits["max_normal"]) / category_limits["std_dev"]
                anomaly_factor = min(5.0, anomaly_factor)  # Limitar o fator
                
                # Calcular peso com base na anomalia
                weight = min(0.9, 0.5 + (anomaly_factor / 10))
                
                risk_factors.append({
                    "factor": "unusual_amount",
                    "description": f"Valor anormal para {tx_category}: R$ {tx_amount:.2f}",
                    "weight": weight
                })
            
            # 2. Verificar horário da transação
            hour = tx_datetime.hour
            
            # Transações em horários de baixa atividade são mais suspeitas no Brasil
            if hour in brasil_patterns["low_activity_hours"]:
                risk_factors.append({
                    "factor": "unusual_hour",
                    "description": f"Transação em horário atípico: {hour}h",
                    "weight": self.regional_risk_factors["unusual_transaction_time"]
                })
            
            # 3. Verificar comportamento típico de final de mês no Brasil
            day_of_month = tx_datetime.day
            is_month_end = day_of_month >= 25 and day_of_month <= 5  # Inclui virada de mês
            
            if is_month_end and tx_amount > (category_limits["mean"] * 2) and tx_category != "bill_payment":
                # Transações grandes que não são pagamentos de contas no fim do mês são menos suspeitas no Brasil
                # devido ao padrão de pagamentos de salários/contas neste período
                pass  # Não adicionamos fator de risco neste caso
            
            # 4. Verificar se a transação é PIX (específico do Brasil)
            if "pix" in tx_type.lower() or "pix" in tx_payment_method.lower() or "pix" in tx_description.lower():
                pix_patterns = brasil_patterns.get("pix_specific_patterns", {})
                
                # PIX de valor muito alto é suspeito
                if tx_amount > pix_patterns.get("max_normal_amount", 5000.0):
                    risk_factors.append({
                        "factor": "high_pix_amount",
                        "description": f"Transação PIX de valor elevado: R$ {tx_amount:.2f}",
                        "weight": 0.7
                    })
                
                # TODO: Verificar frequência de PIX se tivermos histórico
                # Isso requer dados de histórico de transações
            
            # 5. Verificar categoria de comerciante de alto risco
            high_risk_merchants = brasil_patterns.get("high_risk_merchants_categories", [])
            merchant_risk = False
            
            for risk_merchant_type in high_risk_merchants:
                if risk_merchant_type in tx_merchant or risk_merchant_type in tx_description:
                    merchant_risk = True
                    risk_factors.append({
                        "factor": "high_risk_merchant",
                        "description": f"Comerciante/serviço de categoria de alto risco: {risk_merchant_type}",
                        "weight": 0.7
                    })
                    break
            
            # 6. Verificar transação internacional (comum alvo de fraude no Brasil)
            is_international = tx_currency != "BRL"
            if is_international:
                risk_factors.append({
                    "factor": "international_transaction",
                    "description": f"Transação internacional em {tx_currency}",
                    "weight": 0.65
                })
            
            # 7. Verificar termos suspeitos na descrição (específicos do Brasil)
            suspicious_terms = [
                "investimento", "retorno", "garantido", "lucro", "bonus", "promocao", 
                "sorteio", "premio", "resgate", "exterior", "bitcoin", "cripto"
            ]
            
            for term in suspicious_terms:
                if term in tx_description:
                    risk_factors.append({
                        "factor": "suspicious_description",
                        "description": f"Termo suspeito na descrição: '{term}'",
                        "weight": 0.6
                    })
                    break
            
            # Calcular score de risco com base nos fatores identificados
            risk_score = sum(factor["weight"] for factor in risk_factors) / max(len(risk_factors), 1) if risk_factors else 0.1
            risk_score = min(risk_score, 1.0)  # Limitar a 1.0
            
            # Determinar nível de risco
            risk_level = "low"
            if risk_score >= 0.7:
                risk_level = "high"
            elif risk_score >= 0.4:
                risk_level = "medium"
            
            return {
                "risk_score": risk_score,
                "risk_level": risk_level,
                "risk_factors": risk_factors,
                "details": {
                    "transaction_type": tx_type,
                    "amount": tx_amount,
                    "datetime": tx_datetime.isoformat() if hasattr(tx_datetime, 'isoformat') else str(tx_datetime),
                    "currency": tx_currency
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao analisar padrões de transação do Brasil: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao processar análise de transações: {str(e)}"}
            }
    
    def evaluate_account_risk(self, entity_id: str, account_data: Dict, 
                           history_data: Optional[Dict] = None) -> Dict:
        """
        Avalia o nível de risco de uma conta com base em seu perfil e histórico
        no contexto brasileiro.
        
        Args:
            entity_id: ID da entidade
            account_data: Dados da conta
            history_data: Dados históricos da conta
            
        Returns:
            Avaliação de risco da conta
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Verificar idade da conta
            account_age_days = 0
            if "creation_date" in account_data:
                try:
                    creation_date = datetime.fromisoformat(account_data["creation_date"])
                    account_age_days = (datetime.now() - creation_date).days
                except:
                    pass
                    
            # Contas recentes têm risco mais elevado
            if account_age_days < 30:
                risk_factors.append({
                    "factor": "new_account",
                    "description": f"Conta recente (idade: {account_age_days} dias)",
                    "weight": 0.7
                })
            
            # Verificar se a conta tem dados completos de KYC
            kyc_status = account_data.get("kyc_status", "pending")
            if kyc_status != "verified":
                risk_factors.append({
                    "factor": "incomplete_kyc",
                    "description": f"KYC incompleto (status: {kyc_status})",
                    "weight": 0.8
                })
            
            # Verificar histórico de atividades suspeitas
            suspicious_activities = account_data.get("suspicious_activities", [])
            if suspicious_activities and len(suspicious_activities) > 0:
                risk_factors.append({
                    "factor": "suspicious_history",
                    "description": f"Histórico de atividades suspeitas ({len(suspicious_activities)} ocorrências)",
                    "weight": 0.85
                })
            
            # Verificar validação bancária/financeira
            has_valid_bank = account_data.get("has_valid_bank_account", False)
            if not has_valid_bank:
                risk_factors.append({
                    "factor": "no_bank_account",
                    "description": "Sem conta bancária verificada",
                    "weight": 0.6
                })
            
            # Verificar se a conta tem informações de identidade verificadas
            id_verification = account_data.get("id_verification", {})
            id_verified = id_verification.get("verified", False)
            if not id_verified:
                risk_factors.append({
                    "factor": "unverified_identity",
                    "description": "Identidade não verificada",
                    "weight": 0.85
                })
            
            # Verificar se houve alterações recentes de dados sensíveis
            recent_changes = account_data.get("recent_changes", [])
            sensitive_changes = [c for c in recent_changes if c.get("field_type") == "sensitive"]
            if sensitive_changes and len(sensitive_changes) > 0:
                risk_factors.append({
                    "factor": "recent_sensitive_changes",
                    "description": f"Alterações recentes em dados sensíveis ({len(sensitive_changes)} ocorrências)",
                    "weight": 0.65
                })
            
            # Verificar número de dispositivos associados (muitos dispositivos podem ser suspeitos)
            num_devices = len(account_data.get("devices", []))
            if num_devices > 3:
                risk_factors.append({
                    "factor": "multiple_devices",
                    "description": f"Múltiplos dispositivos associados ({num_devices})",
                    "weight": 0.4 + min(0.4, (num_devices - 3) * 0.1)  # Peso aumenta com mais dispositivos
                })
            
            # Fatores específicos do Brasil
            
            # Verificação de CPF/CNPJ
            has_cpf = "cpf_number" in account_data
            has_cnpj = "cnpj_number" in account_data
            
            if not has_cpf and not has_cnpj:
                risk_factors.append({
                    "factor": "missing_brazil_id",
                    "description": "Falta identificação brasileira (CPF/CNPJ)",
                    "weight": 0.8
                })
            
            # Verificação de restrições no Serasa/SPC
            has_credit_restrictions = account_data.get("has_credit_restrictions", False)
            if has_credit_restrictions:
                risk_factors.append({
                    "factor": "credit_restrictions",
                    "description": "Restrições de crédito em bureaus brasileiros",
                    "weight": self.regional_risk_factors["serasa_restrictions"]
                })
            
            # Verificar presença no CADIN (Cadastro Informativo de Créditos não Quitados)
            in_cadin = account_data.get("in_cadin", False)
            if in_cadin:
                risk_factors.append({
                    "factor": "in_cadin",
                    "description": "Entidade presente no CADIN",
                    "weight": self.regional_risk_factors["cadin_presence"]
                })
            
            # Verificar tipo de atividade econômica da conta (específico para Brasil)
            activity_type = account_data.get("economic_activity", "unknown").lower()
            high_risk_activities = ["criptomoedas", "cambio", "jogos", "apostas", "intermediação financeira informal"]
            
            if activity_type in high_risk_activities or activity_type == "unknown":
                risk_factors.append({
                    "factor": "high_risk_activity",
                    "description": f"Atividade econômica de alto risco: {activity_type}",
                    "weight": 0.7
                })
            
            # Verificar se é conta PJ nova com movimentação alta (comum em fraudes no Brasil)
            is_business = account_data.get("is_business", False)
            if is_business and account_age_days < 90:
                high_movement = account_data.get("high_movement", False)
                if high_movement:
                    risk_factors.append({
                        "factor": "new_business_high_movement",
                        "description": "Empresa nova com alta movimentação",
                        "weight": 0.75
                    })
            
            # Verificar se a conta tem endereço em área de alto risco
            address = account_data.get("address", {})
            city = address.get("city", "").lower()
            neighborhood = address.get("neighborhood", "").lower()
            state = address.get("state", "").upper()
            
            # Áreas com maior risco de fraude no Brasil (exemplos)
            high_risk_cities = ["são paulo", "rio de janeiro", "salvador", "fortaleza", "manaus"]
            if city in high_risk_cities:
                # Verificar bairros específicos de alto risco
                high_risk_areas = self.regional_patterns["location_patterns"]["high_risk_areas"]
                if any(area.lower() in neighborhood.lower() for area in high_risk_areas):
                    risk_factors.append({
                        "factor": "high_risk_address",
                        "description": f"Endereço em área de alto risco: {neighborhood}, {city}",
                        "weight": 0.6
                    })
            
            # Verificar se o estado está em região de fronteira (maior risco)
            border_states = ["AC", "RO", "MT", "MS", "PR", "SC", "RS", "AP", "RR", "AM"]
            if state in border_states:
                risk_factors.append({
                    "factor": "border_state",
                    "description": f"Estado em região de fronteira: {state}",
                    "weight": self.regional_patterns["location_patterns"]["border_areas_risk"]
                })
            
            # Calcular score de risco com base nos fatores identificados
            risk_score = sum(factor["weight"] for factor in risk_factors) / max(len(risk_factors), 1)
            risk_score = min(risk_score, 1.0)  # Limitar a 1.0
            
            # Se não houver fatores de risco, usar um score baixo padrão
            if not risk_factors:
                risk_score = 0.1
            
            # Determinar nível de risco
            risk_level = "low"
            if risk_score >= 0.7:
                risk_level = "high"
            elif risk_score >= 0.4:
                risk_level = "medium"
            
            return {
                "risk_score": risk_score,
                "risk_level": risk_level,
                "risk_factors": risk_factors,
                "details": {
                    "account_age_days": account_age_days,
                    "kyc_status": kyc_status,
                    "id_verified": id_verified,
                    "has_valid_bank": has_valid_bank
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao avaliar risco da conta para o Brasil: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao avaliar risco da conta: {str(e)}"}
            }
    
    def detect_location_anomalies(self, entity_id: str, location_data: Dict, 
                               history: Optional[List[Dict]] = None) -> Dict:
        """
        Detecta anomalias de localização com sensibilidade ao contexto regional brasileiro.
        
        Args:
            entity_id: ID da entidade
            location_data: Dados de localização atuais
            history: Histórico de localizações
            
        Returns:
            Resultado da detecção de anomalias de localização
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Validar dados de localização atuais
            if not location_data or not isinstance(location_data, dict):
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Dados de localização insuficientes para análise"}
                }
            
            # Verificar se temos coordenadas
            has_coordinates = "latitude" in location_data and "longitude" in location_data
            
            # Obter padrões de localização do Brasil
            location_patterns = self.regional_patterns["location_patterns"]
            
            # 1. Verificar se a localização atual está em área de alto risco
            current_city = location_data.get("city", "").strip().lower()
            current_state = location_data.get("state", "").strip().upper()
            current_district = location_data.get("district", "").strip().lower()
            current_country = location_data.get("country", "").strip()
            
            # Verificar se está fora do Brasil
            if current_country and current_country != "BR" and current_country != "Brasil" and current_country != "Brazil":
                risk_factors.append({
                    "factor": "foreign_location",
                    "description": f"Localização fora do Brasil ({current_country})",
                    "weight": self.regional_risk_factors["foreign_ip_access"]
                })
            
            # Verificar áreas de alto risco no Brasil
            high_risk_areas = location_patterns["high_risk_areas"]
            
            if current_district and any(area.lower() in current_district for area in high_risk_areas):
                risk_factors.append({
                    "factor": "high_risk_area",
                    "description": f"Localização em área de alto risco: {current_district}",
                    "weight": 0.7
                })
            
            # Verificar estados com risco elevado
            if current_state:
                state_risk = location_patterns["state_risk_factors"].get(current_state, 0.5)
                if state_risk > 0.6:
                    risk_factors.append({
                        "factor": "high_risk_state",
                        "description": f"Estado com fator de risco elevado: {current_state}",
                        "weight": self.regional_risk_factors["high_risk_state"]
                    })
                
                # Verificar se é estado de fronteira e está próximo da fronteira
                border_states = ["AC", "RO", "MT", "MS", "PR", "SC", "RS", "AP", "RR", "AM"]
                if current_state in border_states and has_coordinates:
                    # Verificar proximidade com fronteiras (simplificado)
                    # Uma verificação real usaria dados geográficos precisos
                    border_proximity = False
                    
                    # Áreas de fronteira conhecidas por maior risco
                    if current_state == "MS" and "ponta porã" in current_city:
                        border_proximity = True
                    elif current_state == "PR" and "foz do iguaçu" in current_city:
                        border_proximity = True
                    elif current_state == "RS" and "uruguaiana" in current_city:
                        border_proximity = True
                    elif current_state == "AC" and "brasileia" in current_city:
                        border_proximity = True
                    
                    if border_proximity:
                        risk_factors.append({
                            "factor": "border_area",
                            "description": f"Localização em área de fronteira ({current_city}, {current_state})",
                            "weight": location_patterns["border_areas_risk"]
                        })
            
            # 2. Verificar anomalias de movimento (se houver histórico)
            if history and has_coordinates and len(history) > 0:
                # Obter coordenadas atuais
                current_lat = float(location_data["latitude"])
                current_lon = float(location_data["longitude"])
                
                # Processar histórico para análise
                valid_history = []
                
                for loc in history:
                    if isinstance(loc, dict) and "latitude" in loc and "longitude" in loc and "timestamp" in loc:
                        try:
                            loc_lat = float(loc["latitude"])
                            loc_lon = float(loc["longitude"])
                            timestamp = datetime.fromisoformat(loc["timestamp"]) if isinstance(loc["timestamp"], str) else loc["timestamp"]
                            
                            valid_history.append({
                                "latitude": loc_lat,
                                "longitude": loc_lon,
                                "timestamp": timestamp
                            })
                        except (ValueError, TypeError):
                            continue
                
                # Ordenar histórico por timestamp
                valid_history = sorted(valid_history, key=lambda x: x["timestamp"])
                
                if valid_history:
                    # Obter última localização conhecida
                    last_location = valid_history[-1]
                    last_lat = last_location["latitude"]
                    last_lon = last_location["longitude"]
                    last_timestamp = last_location["timestamp"]
                    
                    # Calcular distância entre localização atual e última conhecida
                    from math import sin, cos, sqrt, atan2, radians
                    
                    # Raio da Terra em km
                    R = 6371.0
                    
                    # Converter coordenadas para radianos
                    lat1 = radians(last_lat)
                    lon1 = radians(last_lon)
                    lat2 = radians(current_lat)
                    lon2 = radians(current_lon)
                    
                    # Fórmula haversine
                    dlon = lon2 - lon1
                    dlat = lat2 - lat1
                    a = sin(dlat / 2)**2 + cos(lat1) * cos(lat2) * sin(dlon / 2)**2
                    c = 2 * atan2(sqrt(a), sqrt(1 - a))
                    distance = R * c  # distância em km
                    
                    # Calcular tempo entre localizações
                    current_timestamp = datetime.fromisoformat(location_data.get("timestamp", datetime.now().isoformat())) if isinstance(location_data.get("timestamp"), str) else location_data.get("timestamp", datetime.now())
                    time_diff = (current_timestamp - last_timestamp).total_seconds() / 3600  # em horas
                    
                    # Evitar divisão por zero
                    if time_diff <= 0:
                        time_diff = 0.01
                    
                    # Calcular velocidade de deslocamento em km/h
                    speed = distance / time_diff
                    
                    # Verificar se a velocidade é anormalmente alta
                    typical_speed = location_patterns["typical_speed_kmh"]
                    max_reasonable_speed = 900  # km/h (aproximadamente velocidade de avião comercial)
                    
                    if speed > max_reasonable_speed:
                        risk_factors.append({
                            "factor": "impossible_travel",
                            "description": f"Velocidade de deslocamento fisicamente impossível ({speed:.1f} km/h)",
                            "weight": 0.95
                        })
                    elif speed > (typical_speed * 3):
                        risk_factors.append({
                            "factor": "unusual_travel_speed",
                            "description": f"Velocidade de deslocamento anormalmente alta ({speed:.1f} km/h)",
                            "weight": 0.65
                        })
                    
                    # Verificar raio de movimento típico
                    typical_radius = location_patterns["common_movement_radius_km"]
                    
                    if distance > (typical_radius * 3) and time_diff < 24:
                        risk_factors.append({
                            "factor": "unusual_movement_radius",
                            "description": f"Deslocamento fora do raio típico ({distance:.1f} km em {time_diff:.1f} horas)",
                            "weight": 0.7
                        })
            
            # 3. Verificar dados de IP vs. dados de GPS (se disponíveis)
            gps_location = {"lat": location_data.get("latitude"), "lon": location_data.get("longitude")}
            ip_location = {"lat": location_data.get("ip_latitude"), "lon": location_data.get("ip_longitude")}
            
            if all(gps_location.values()) and all(ip_location.values()):
                # Calcular distância entre localização GPS e IP
                from math import sin, cos, sqrt, atan2, radians
                
                R = 6371.0  # Raio da Terra em km
                
                lat1 = radians(float(ip_location["lat"]))
                lon1 = radians(float(ip_location["lon"]))
                lat2 = radians(float(gps_location["lat"]))
                lon2 = radians(float(gps_location["lon"]))
                
                dlon = lon2 - lon1
                dlat = lat2 - lat1
                a = sin(dlat / 2)**2 + cos(lat1) * cos(lat2) * sin(dlon / 2)**2
                c = 2 * atan2(sqrt(a), sqrt(1 - a))
                distance = R * c  # distância em km
                
                # Se a distância for significativa, é uma anomalia
                # No Brasil, usamos um limite menor devido à densidade populacional maior
                if distance > 50:
                    risk_factors.append({
                        "factor": "ip_gps_mismatch",
                        "description": f"Divergência entre localização GPS e IP ({distance:.1f} km)",
                        "weight": 0.8
                    })
            
            # 4. Verificar se a localização IP é de país estrangeiro mas GPS indica Brasil
            ip_country = location_data.get("ip_country", "")
            if ip_country and ip_country != "BR" and ip_country != "Brasil" and ip_country != "Brazil":
                if current_country == "BR" or current_country == "Brasil" or current_country == "Brazil":
                    risk_factors.append({
                        "factor": "vpn_location_mismatch",
                        "description": f"IP estrangeiro ({ip_country}) com GPS no Brasil",
                        "weight": 0.85
                    })
            
            # Calcular score de risco com base nos fatores identificados
            risk_score = sum(factor["weight"] for factor in risk_factors) / max(len(risk_factors), 1) if risk_factors else 0.1
            risk_score = min(risk_score, 1.0)  # Limitar a 1.0
            
            # Determinar nível de risco
            risk_level = "low"
            if risk_score >= 0.7:
                risk_level = "high"
            elif risk_score >= 0.4:
                risk_level = "medium"
            
            return {
                "risk_score": risk_score,
                "risk_level": risk_level,
                "risk_factors": risk_factors,
                "details": {
                    "current_location": {
                        "country": current_country,
                        "state": current_state,
                        "city": current_city,
                        "district": current_district
                    },
                    "has_history": bool(history and len(history) > 0),
                    "historical_points": len(history) if history else 0
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de localização para o Brasil: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao detectar anomalias de localização: {str(e)}"}
            }
            
    def analyze_device_behavior(self, entity_id: str, device_data: Dict, 
                             session_data: Optional[Dict] = None,
                             device_history: Optional[List] = None) -> Dict:
        """
        Analisa o comportamento do dispositivo para detectar anomalias no contexto brasileiro.
        
        Args:
            entity_id: ID da entidade
            device_data: Dados do dispositivo atual
            session_data: Dados da sessão atual
            device_history: Histórico de dispositivos
            
        Returns:
            Resultado da análise de comportamento do dispositivo
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Validar dados de dispositivo
            if not device_data or not isinstance(device_data, dict):
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Dados de dispositivo insuficientes para análise"}
                }
            
            # Obter padrões de comportamento de dispositivo para o Brasil
            device_patterns = self.regional_patterns["behavioral_patterns"]["device_usage"]
            
            # 1. Verificar se o dispositivo é conhecido
            device_id = device_data.get("device_id", "")
            device_fingerprint = device_data.get("fingerprint", "")
            is_known_device = False
            
            if device_history:
                device_ids = [d.get("device_id", "") for d in device_history if isinstance(d, dict)]
                device_fingerprints = [d.get("fingerprint", "") for d in device_history if isinstance(d, dict)]
                
                is_known_device = device_id in device_ids or device_fingerprint in device_fingerprints
            
            if not is_known_device:
                risk_factors.append({
                    "factor": "new_device",
                    "description": "Dispositivo não reconhecido",
                    "weight": self.regional_risk_factors["new_device_login"]
                })
            
            # 2. Verificar se o dispositivo é um emulador ou possui indicadores de fraude
            is_emulator = device_data.get("is_emulator", False)
            is_rooted = device_data.get("is_rooted", False) or device_data.get("is_jailbroken", False)
            fraud_signals = device_data.get("fraud_signals", [])
            
            if is_emulator:
                risk_factors.append({
                    "factor": "emulator_detected",
                    "description": "Emulador detectado",
                    "weight": 0.8
                })
            
            if is_rooted:
                risk_factors.append({
                    "factor": "rooted_device",
                    "description": "Dispositivo root/jailbreak detectado",
                    "weight": 0.75
                })
            
            if fraud_signals and len(fraud_signals) > 0:
                risk_factors.append({
                    "factor": "fraud_signals",
                    "description": f"Sinais de fraude detectados: {', '.join(fraud_signals)}",
                    "weight": 0.85
                })
            
            # 3. Verificar flags de segurança (VPN, Proxy, Tor)
            is_vpn = device_data.get("is_vpn", False)
            is_proxy = device_data.get("is_proxy", False)
            is_tor = device_data.get("is_tor", False)
            
            if is_vpn:
                risk_factors.append({
                    "factor": "vpn_detected",
                    "description": "Uso de VPN detectado",
                    "weight": 0.6
                })
            
            if is_proxy:
                risk_factors.append({
                    "factor": "proxy_detected",
                    "description": "Uso de proxy detectado",
                    "weight": 0.65
                })
            
            if is_tor:
                risk_factors.append({
                    "factor": "tor_detected",
                    "description": "Uso da rede Tor detectado",
                    "weight": 0.8
                })
            
            # 4. Verificar inconsistências de navegador e sistema operacional
            browser = device_data.get("browser", "")
            os_name = device_data.get("os", "")
            user_agent = device_data.get("user_agent", "")
            
            # Verificar inconsistências no user agent
            ua_inconsistent = False
            
            # Exemplos de inconsistências
            if "windows" in os_name.lower() and "safari" in browser.lower():
                ua_inconsistent = True
            elif "android" in os_name.lower() and "safari" in browser.lower() and "chrome" not in browser.lower():
                ua_inconsistent = True
            elif "ios" in os_name.lower() and "chrome" in browser.lower():
                ua_inconsistent = True
            elif user_agent:
                if "windows" in user_agent.lower() and "android" in os_name.lower():
                    ua_inconsistent = True
                elif "macintosh" in user_agent.lower() and "windows" in os_name.lower():
                    ua_inconsistent = True
            
            if ua_inconsistent:
                risk_factors.append({
                    "factor": "os_browser_inconsistency",
                    "description": f"Inconsistência entre sistema ({os_name}) e navegador ({browser})",
                    "weight": 0.75
                })
            
            # 5. Analisar comportamento de sessão (se disponível)
            if session_data and isinstance(session_data, dict):
                # Verificar tempo de digitação (se disponível)
                typing_speed = session_data.get("typing_speed", {})
                typing_pattern = session_data.get("typing_pattern", {})
                
                if typing_speed and typing_pattern:
                    # Verificar se a velocidade de digitação é anômala (muito rápida pode indicar automação)
                    if typing_speed.get("is_anomalous", False):
                        risk_factors.append({
                            "factor": "anomalous_typing",
                            "description": "Padrão de digitação anômalo detectado",
                            "weight": 0.7
                        })
                
                # Verificar cliques e movimentação do mouse (se disponível)
                mouse_movement = session_data.get("mouse_movement", {})
                if mouse_movement:
                    is_natural = mouse_movement.get("is_natural", True)
                    if not is_natural:
                        risk_factors.append({
                            "factor": "unnatural_mouse_movement",
                            "description": "Movimentação não natural do mouse",
                            "weight": 0.65
                        })
                
                # Verificar falhas de autenticação na sessão
                auth_failures = session_data.get("authentication_failures", 0)
                if auth_failures > 2:
                    risk_factors.append({
                        "factor": "multiple_auth_failures",
                        "description": f"Múltiplas falhas de autenticação ({auth_failures})",
                        "weight": self.regional_risk_factors["multiple_auth_failures"]
                    })
                
                # Verificar hora da sessão (período noturno no Brasil tem maior risco)
                session_timestamp = session_data.get("timestamp")
                if session_timestamp:
                    if isinstance(session_timestamp, str):
                        try:
                            session_time = datetime.fromisoformat(session_timestamp)
                        except:
                            session_time = datetime.now()
                    else:
                        session_time = session_timestamp
                    
                    hour = session_time.hour
                    # No Brasil, o período de 1h às 5h da manhã é menos comum para transações legítimas
                    if 1 <= hour <= 5:
                        risk_factors.append({
                            "factor": "unusual_session_hour",
                            "description": f"Login em horário atípico ({hour}h)",
                            "weight": 0.55
                        })
            
            # 6. Verificar contexto específico brasileiro
            
            # Verificar se é um dispositivo típico no Brasil
            device_type = device_data.get("type", "").lower()
            device_model = device_data.get("model", "").lower()
            
            # No Brasil, há forte predominância de dispositivos Android e Windows
            is_brazilian_common_device = False
            
            if "android" in os_name.lower():
                is_brazilian_common_device = True
            elif "windows" in os_name.lower():
                is_brazilian_common_device = True
            
            # Verificar se é um modelo popular no Brasil
            popular_brands_brazil = ["samsung", "motorola", "lg", "xiaomi", "huawei"]
            if any(brand in device_model for brand in popular_brands_brazil):
                is_brazilian_common_device = True
            
            # Se não for dispositivo comum, aumentar levemente o risco
            if not is_brazilian_common_device and device_type != "desktop":
                risk_factors.append({
                    "factor": "uncommon_device_brazil",
                    "description": f"Dispositivo incomum para o Brasil: {device_model} ({os_name})",
                    "weight": 0.4
                })
            
            # Verificar histórico de fraude do dispositivo no Brasil (se disponível)
            has_fraud_history = device_data.get("has_fraud_history", False)
            if has_fraud_history:
                risk_factors.append({
                    "factor": "device_fraud_history",
                    "description": "Dispositivo com histórico de fraude",
                    "weight": self.regional_risk_factors["device_fraud_history"]
                })
            
            # Verificar o tipo de conexão (no Brasil, conexão 4G é predominante)
            connection_type = device_data.get("connection_type", "").lower()
            
            # Conexões que não são de operadoras móveis brasileiras conhecidas podem ser suspeitas
            brazilian_isps = ["vivo", "claro", "tim", "oi", "algar", "net", "gvt", "globo", "uol", "sercomtel"]
            connection_isp = device_data.get("isp", "").lower()
            
            if connection_type == "mobile" and connection_isp and not any(isp in connection_isp for isp in brazilian_isps):
                risk_factors.append({
                    "factor": "uncommon_mobile_isp",
                    "description": f"Operadora móvel não comum no Brasil: {connection_isp}",
                    "weight": 0.55
                })
            
            # Calcular score de risco com base nos fatores identificados
            risk_score = sum(factor["weight"] for factor in risk_factors) / max(len(risk_factors), 1) if risk_factors else 0.1
            risk_score = min(risk_score, 1.0)  # Limitar a 1.0
            
            # Determinar nível de risco
            risk_level = "low"
            if risk_score >= 0.7:
                risk_level = "high"
            elif risk_score >= 0.4:
                risk_level = "medium"
            
            return {
                "risk_score": risk_score,
                "risk_level": risk_level,
                "risk_factors": risk_factors,
                "details": {
                    "device_type": device_type,
                    "os": os_name,
                    "browser": browser,
                    "is_known_device": is_known_device,
                    "security_flags": {
                        "is_emulator": is_emulator,
                        "is_rooted": is_rooted,
                        "is_vpn": is_vpn,
                        "is_proxy": is_proxy,
                        "is_tor": is_tor
                    }
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao analisar comportamento do dispositivo para o Brasil: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao analisar comportamento do dispositivo: {str(e)}"}
            }
    
    def get_regional_risk_factors(self, entity_id: str, entity_data: Dict) -> Dict:
        """
        Obtém fatores de risco específicos para o Brasil usando
        adaptadores regionais (Serasa, Receita Federal, BACEN, etc.)
        
        Args:
            entity_id: ID da entidade
            entity_data: Dados da entidade
            
        Returns:
            Fatores de risco regionais
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Verificar se temos adaptadores regionais disponíveis
            if not self.data_adapters:
                logger.warning(f"Sem adaptadores regionais disponíveis para o Brasil")
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Adaptadores regionais não disponíveis"}
                }
            
            # Extrair dados relevantes para consulta
            cpf_number = entity_data.get("cpf_number", "")
            cnpj_number = entity_data.get("cnpj_number", "")
            full_name = entity_data.get("full_name", "")
            email = entity_data.get("email", "")
            phone = entity_data.get("phone", "")
            
            # Verificar dados suficientes para identificação
            has_identification = cpf_number or cnpj_number or (full_name and (email or phone))
            
            if not has_identification:
                return {
                    "risk_score": 0.7,  # Pontuação alta de risco quando não há identificação adequada
                    "risk_level": "high",
                    "risk_factors": [{
                        "factor": "insufficient_identification",
                        "description": "Dados insuficientes para verificação regional",
                        "weight": 0.7
                    }],
                    "details": {"reason": "Dados de identificação insuficientes para Brasil"}
                }
            
            # 1. Verificar listas de restrição (Serasa/SPC/etc.)
            if "serasa" in self.data_adapters:
                try:
                    serasa_data = self.data_adapters["serasa"].verify_entity(
                        cpf=cpf_number,
                        cnpj=cnpj_number,
                        name=full_name
                    )
                    
                    if serasa_data.get("has_restrictions", False):
                        restrictions = serasa_data.get("restrictions", [])
                        restriction_count = len(restrictions)
                        
                        risk_factors.append({
                            "factor": "serasa_restrictions",
                            "description": f"Restrições no Serasa ({restriction_count})",
                            "weight": self.regional_risk_factors["serasa_restrictions"],
                            "source": "serasa"
                        })
                    
                    if serasa_data.get("score", 0) < 300:  # Score Serasa abaixo de 300 é problemático
                        risk_factors.append({
                            "factor": "low_credit_score",
                            "description": f"Score de crédito baixo: {serasa_data.get('score', 0)}",
                            "weight": 0.65,
                            "source": "serasa"
                        })
                except Exception as e:
                    logger.error(f"Erro ao verificar Serasa: {str(e)}")
            
            # 2. Verificar status PEP e CADIN (Receita Federal)
            if "receita_federal" in self.data_adapters:
                try:
                    receita_data = self.data_adapters["receita_federal"].verify_entity(
                        cpf=cpf_number,
                        cnpj=cnpj_number,
                        name=full_name
                    )
                    
                    if receita_data.get("is_pep", False):
                        risk_factors.append({
                            "factor": "is_pep",
                            "description": "Pessoa Politicamente Exposta (PEP)",
                            "weight": 0.6,  # PEP não é necessariamente risco, mas requer maior vigilância
                            "source": "receita_federal"
                        })
                    
                    if receita_data.get("in_cadin", False):
                        risk_factors.append({
                            "factor": "in_cadin",
                            "description": "Presença no CADIN (Cadastro Informativo de Créditos não Quitados)",
                            "weight": self.regional_risk_factors["cadin_presence"],
                            "source": "receita_federal"
                        })
                    
                    if receita_data.get("document_status", "") == "irregular":
                        risk_factors.append({
                            "factor": "document_irregular",
                            "description": "Documento com situação irregular na Receita Federal",
                            "weight": 0.8,
                            "source": "receita_federal"
                        })
                        
                    if receita_data.get("company_status", "") == "baixada" and cnpj_number:
                        risk_factors.append({
                            "factor": "company_closed",
                            "description": "CNPJ com status 'baixada' ou inativo",
                            "weight": 0.85,
                            "source": "receita_federal"
                        })
                except Exception as e:
                    logger.error(f"Erro ao verificar Receita Federal: {str(e)}")
            
            # 3. Verificar histórico de transações financeiras (BACEN)
            if "bacen" in self.data_adapters:
                try:
                    bacen_data = self.data_adapters["bacen"].verify_entity(
                        cpf=cpf_number,
                        cnpj=cnpj_number,
                        name=full_name
                    )
                    
                    if bacen_data.get("has_currency_violations", False):
                        risk_factors.append({
                            "factor": "currency_violations",
                            "description": "Histórico de violações de câmbio/moeda estrangeira",
                            "weight": 0.75,
                            "source": "bacen"
                        })
                    
                    if bacen_data.get("unusual_money_flow", False):
                        risk_factors.append({
                            "factor": "unusual_money_flow",
                            "description": "Fluxo de dinheiro incomum detectado pelo BACEN",
                            "weight": 0.7,
                            "source": "bacen"
                        })
                        
                    foreign_accounts = bacen_data.get("foreign_accounts", 0)
                    if foreign_accounts > 0:
                        risk_factors.append({
                            "factor": "foreign_accounts",
                            "description": f"Contas no exterior registradas no BACEN: {foreign_accounts}",
                            "weight": 0.5,  # Ter conta no exterior não é necessariamente risco alto
                            "source": "bacen"
                        })
                except Exception as e:
                    logger.error(f"Erro ao verificar BACEN: {str(e)}")
            
            # 4. Verificar comportamentos suspeitos com operadoras de telecom
            if "telecom_brasil" in self.data_adapters and phone:
                try:
                    telecom_data = self.data_adapters["telecom_brasil"].verify_entity(
                        phone=phone,
                        cpf=cpf_number,
                        name=full_name
                    )
                    
                    if telecom_data.get("is_burner_phone", False):
                        risk_factors.append({
                            "factor": "burner_phone",
                            "description": "Telefone descartável ou pré-pago recente",
                            "weight": 0.65,
                            "source": "telecom_brasil"
                        })
                    
                    if telecom_data.get("multiple_sim_changes", 0) > 2:
                        risk_factors.append({
                            "factor": "multiple_sim_changes",
                            "description": f"Múltiplas trocas de SIM recentes: {telecom_data.get('multiple_sim_changes')}",
                            "weight": 0.6,
                            "source": "telecom_brasil"
                        })
                    
                    if telecom_data.get("fraud_reports", 0) > 0:
                        risk_factors.append({
                            "factor": "phone_fraud_reports",
                            "description": f"Número com denúncias de fraude: {telecom_data.get('fraud_reports')}",
                            "weight": 0.75,
                            "source": "telecom_brasil"
                        })
                except Exception as e:
                    logger.error(f"Erro ao verificar dados de telecom: {str(e)}")
            
            # Calcular score de risco com base nos fatores identificados
            risk_score = sum(factor["weight"] for factor in risk_factors) / max(len(risk_factors), 1) if risk_factors else 0.1
            risk_score = min(risk_score, 1.0)  # Limitar a 1.0
            
            # Determinar nível de risco
            risk_level = "low"
            if risk_score >= 0.7:
                risk_level = "high"
            elif risk_score >= 0.4:
                risk_level = "medium"
            
            # Se não encontramos fatores de risco mas temos identificação adequada, o risco é baixo
            if not risk_factors and has_identification:
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Verificação regional não identificou fatores de risco"}
                }
            
            return {
                "risk_score": risk_score,
                "risk_level": risk_level,
                "risk_factors": risk_factors,
                "details": {
                    "cpf_verified": bool(cpf_number),
                    "cnpj_verified": bool(cnpj_number),
                    "name_verified": bool(full_name),
                    "adapters_used": list(self.data_adapters.keys())
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao obter fatores de risco regionais do Brasil: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao obter fatores de risco regionais: {str(e)}"}
            }
            
    def _calculate_combined_risk_score(self, results: Dict) -> Dict:
        """
        Calcula um score de risco combinado com base nos resultados
        de diferentes análises, com pesos adaptados ao contexto brasileiro.
        
        Args:
            results: Resultados de diferentes análises
            
        Returns:
            Score de risco combinado e nível
        """
        try:
            # Extrair scores das diferentes análises
            account_risk = results.get("account_risk", {}).get("risk_score", 0.1)
            transaction_risk = results.get("transaction_risk", {}).get("risk_score", 0.1)
            location_risk = results.get("location_risk", {}).get("risk_score", 0.1)
            device_risk = results.get("device_risk", {}).get("risk_score", 0.1)
            regional_risk = results.get("regional_risk", {}).get("risk_score", 0.1)
            
            # Pesos específicos para o Brasil (ajustados para o contexto brasileiro)
            # Fatores regionais e comportamento de conta têm peso maior no Brasil
            weights = {
                "account_risk": 0.25,
                "transaction_risk": 0.20,
                "location_risk": 0.15,
                "device_risk": 0.15,
                "regional_risk": 0.25
            }
            
            # Calcular score ponderado
            weighted_score = (
                account_risk * weights["account_risk"] +
                transaction_risk * weights["transaction_risk"] +
                location_risk * weights["location_risk"] +
                device_risk * weights["device_risk"] +
                regional_risk * weights["regional_risk"]
            )
            
            # Ajustar score se algum dos fatores individuais for de alto risco
            # Este é um princípio de "máximo risco" - se qualquer fator está em vermelho,
            # o score final deve refletir isso
            high_risk_detected = False
            max_individual_score = max([account_risk, transaction_risk, location_risk, device_risk, regional_risk])
            
            if max_individual_score >= 0.8:  # Se qualquer análise individual apresentar risco muito alto
                high_risk_detected = True
                # Aumentar score mantendo a proporção mas garantindo risco alto
                weighted_score = min(0.95, weighted_score * 1.25)
            
            # Determinar nível de risco
            risk_level = "low"
            if weighted_score >= 0.7:
                risk_level = "high"
            elif weighted_score >= 0.4:
                risk_level = "medium"
            
            # Se detectamos alto risco em alguma dimensão individual, garantir que 
            # o risco final seja pelo menos médio
            if high_risk_detected and risk_level == "low":
                risk_level = "medium"
                weighted_score = max(weighted_score, 0.4)
            
            return {
                "risk_score": weighted_score,
                "risk_level": risk_level,
                "details": {
                    "account_risk": account_risk,
                    "transaction_risk": transaction_risk,
                    "location_risk": location_risk,
                    "device_risk": device_risk,
                    "regional_risk": regional_risk,
                    "high_risk_detected": high_risk_detected,
                    "regional_context": "Brasil"
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao calcular score combinado de risco para o Brasil: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "details": {"error": f"Erro ao calcular score combinado de risco: {str(e)}"}
            }