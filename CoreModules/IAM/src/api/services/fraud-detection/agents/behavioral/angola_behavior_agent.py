"""
Agente de Análise Comportamental para Angola

Este módulo implementa o agente de análise comportamental específico para o mercado angolano,
considerando padrões culturais, comportamentais e regionais para detectar fraudes contextuais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
import logging
import numpy as np
import json
import pandas as pd
from typing import Dict, List, Tuple, Optional, Union, Any
from datetime import datetime, timedelta
from .behavior_analysis_framework import BehaviorAnalysisAgent

# Importação dos adaptadores específicos de Angola
from ..angola.bureau_credito_adapter import BureauCreditoAdapter
from ..angola.senica_adapter import SenicaAdapter
from ..angola.telecom_adapter import TelecomAdapter
from ..angola.bna_adapter import BNAAdapter

# Configuração do logging
logger = logging.getLogger("angola_behavior_agent")

class AngolaBehaviorAgent(BehaviorAnalysisAgent):
    """
    Agente de análise comportamental especializado para o mercado angolano,
    considerando padrões culturais, regulações locais e comportamentos específicos.
    """
    
    # Padrões regionais específicos de Angola
    ANGOLA_PATTERNS = {
        # Padrões de uso de telefonia móvel
        "mobile_patterns": {
            "frequent_carriers": ["Unitel", "Movicel", "Angola Telecom"],
            "common_prefixes": ["91", "92", "93", "94", "99"],
            "typical_recharge_values": [500, 1000, 1500, 2000, 2500, 5000]
        },
        
        # Padrões de transação financeira
        "transaction_patterns": {
            "common_values": [1000, 5000, 10000, 20000, 50000, 100000],
            "high_risk_times": ["22:00-06:00"],
            "high_risk_days": ["domingo"],
            "typical_transaction_frequency": {
                "daily": 0.8,
                "weekly": 3.2,
                "monthly": 12.5
            }
        },
        
        # Padrões de localização
        "location_patterns": {
            "major_urban_centers": ["Luanda", "Benguela", "Huambo", "Lubango", "Malanje"],
            "high_risk_areas": ["Rocha Pinto", "Sambizanga", "Cazenga", "Viana"],
            "common_movement_radius_km": 15.0,
            "typical_speed_kmh": 40.0
        },
        
        # Padrões comportamentais
        "behavioral_patterns": {
            "purchase_categories": ["alimentação", "transporte", "telecomunicações", "entretenimento"],
            "purchasing_hours": {
                "weekday": ["08:00-19:00"],
                "weekend": ["10:00-22:00"]
            },
            "device_usage": {
                "mobile_predominance": 0.75,  # 75% de uso móvel vs desktop
                "typical_session_duration_min": 22
            }
        }
    }
    
    # Fatores de risco específicos de Angola
    ANGOLA_RISK_FACTORS = {
        "cross_border_transactions": 1.5,  # multiplicador de risco para transações internacionais
        "new_device_login": 1.3,  # multiplicador para login de dispositivos novos
        "unusual_location": 1.4,  # multiplicador para localizações não usuais
        "unusual_time": 1.2,  # multiplicador para horários não usuais
        "high_value_transaction": 1.8,  # multiplicador para transações de alto valor
        "multiple_auth_failures": 1.6,  # multiplicador para falhas de autenticação repetidas
        "identity_mismatch": 2.0  # multiplicador para divergências de identidade
    }
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        model_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        data_sources: Optional[List[str]] = None,
        adapters_config_path: Optional[str] = None,
        language: str = "pt"
    ):
        """
        Inicializa o agente de análise comportamental para Angola.
        
        Args:
            config_path: Caminho para arquivo de configuração
            model_path: Caminho para modelos pré-treinados
            cache_dir: Diretório para cache de resultados
            data_sources: Lista de fontes de dados a serem utilizadas
            adapters_config_path: Caminho para configuração dos adaptadores
            language: Código do idioma principal (padrão: português)
        """
        # Inicializar classe base
        super().__init__(
            config_path=config_path,
            model_path=model_path,
            cache_dir=cache_dir,
            region_code="AO",  # Código ISO para Angola
            data_sources=data_sources,
            language=language
        )
        
        self.adapters_config_path = adapters_config_path
        
        # Carregar dados regionais específicos
        self.regional_patterns = self.ANGOLA_PATTERNS
        self.regional_risk_factors = self.ANGOLA_RISK_FACTORS
        
        # Carregar modelos específicos para Angola
        self._load_angola_models()
        
        logger.info("Agente de análise comportamental para Angola inicializado")
    
    def _initialize_data_adapters(self) -> None:
        """
        Inicializa os adaptadores para fontes de dados específicas de Angola.
        """
        try:
            # Carregar configuração dos adaptadores se especificada
            adapters_config = {}
            if self.adapters_config_path and os.path.exists(self.adapters_config_path):
                with open(self.adapters_config_path, 'r', encoding='utf-8') as f:
                    adapters_config = json.load(f)
            
            # Inicializar adaptadores específicos
            bureau_config = adapters_config.get("bureau_credito", {})
            senica_config = adapters_config.get("senica", {})
            telecom_config = adapters_config.get("telecom", {})
            bna_config = adapters_config.get("bna", {})
            
            # Inicializar adaptadores apenas se configurados para uso
            if "bureau_credito" in self.data_sources:
                self.data_adapters["bureau_credito"] = BureauCreditoAdapter(
                    config_path=bureau_config.get("config_path"),
                    credentials_path=bureau_config.get("credentials_path"),
                    cache_dir=bureau_config.get("cache_dir")
                )
                logger.info("Adaptador Bureau Crédito de Angola inicializado")
            
            if "senica" in self.data_sources:
                self.data_adapters["senica"] = SenicaAdapter(
                    config_path=senica_config.get("config_path"),
                    credentials_path=senica_config.get("credentials_path"),
                    cache_dir=senica_config.get("cache_dir")
                )
                logger.info("Adaptador SENICA inicializado")
            
            if "telecom" in self.data_sources:
                self.data_adapters["telecom"] = TelecomAdapter(
                    config_path=telecom_config.get("config_path"),
                    credentials_path=telecom_config.get("credentials_path"),
                    cache_dir=telecom_config.get("cache_dir")
                )
                logger.info("Adaptador Telecom Angola inicializado")
            
            if "bna" in self.data_sources:
                self.data_adapters["bna"] = BNAAdapter(
                    config_path=bna_config.get("config_path"),
                    credentials_path=bna_config.get("credentials_path"),
                    cache_dir=bna_config.get("cache_dir")
                )
                logger.info("Adaptador BNA inicializado")
            
            # Estabelecer conexão com os adaptadores
            for adapter_name, adapter in self.data_adapters.items():
                connected = adapter.connect()
                if connected:
                    logger.info(f"Conexão estabelecida com {adapter_name}")
                else:
                    logger.warning(f"Não foi possível estabelecer conexão com {adapter_name}")
        
        except Exception as e:
            logger.error(f"Erro ao inicializar adaptadores: {str(e)}")
    
    def _load_angola_models(self) -> None:
        """
        Carrega modelos específicos para análise comportamental no contexto angolano.
        """
        try:
            angola_models_path = os.path.join(self.model_path, "ao")
            os.makedirs(angola_models_path, exist_ok=True)
            
            # Definir caminhos dos modelos
            transaction_model_path = os.path.join(angola_models_path, "transaction_patterns.model")
            device_model_path = os.path.join(angola_models_path, "device_behavior.model")
            location_model_path = os.path.join(angola_models_path, "location_patterns.model")
            account_model_path = os.path.join(angola_models_path, "account_risk.model")
            
            # Verificar existência dos modelos e carregar
            # Nota: Esta é uma implementação placeholder para demonstração
            # Em produção, aqui seriam carregados modelos ML reais (scikit-learn, TensorFlow, etc.)
            
            # Se os modelos não existirem, criar placeholders com regras básicas
            if not os.path.exists(transaction_model_path):
                logger.warning("Modelo de padrões de transação não encontrado. Usando regras básicas.")
                self.models["transaction"] = {"type": "rules", "rules": self.regional_patterns["transaction_patterns"]}
            
            if not os.path.exists(device_model_path):
                logger.warning("Modelo de comportamento de dispositivo não encontrado. Usando regras básicas.")
                self.models["device"] = {"type": "rules", "rules": self.regional_patterns["behavioral_patterns"]}
            
            if not os.path.exists(location_model_path):
                logger.warning("Modelo de padrões de localização não encontrado. Usando regras básicas.")
                self.models["location"] = {"type": "rules", "rules": self.regional_patterns["location_patterns"]}
            
            if not os.path.exists(account_model_path):
                logger.warning("Modelo de risco de conta não encontrado. Usando regras básicas.")
                self.models["account"] = {"type": "rules", "rules": self.regional_risk_factors}
            
            logger.info("Modelos específicos para Angola carregados")
            
        except Exception as e:
            logger.error(f"Erro ao carregar modelos para Angola: {str(e)}")
    
    def analyze_behavior(self, entity_id: str, entity_type: str, transaction_data: Dict, 
                       context_data: Optional[Dict] = None, use_cache: bool = True) -> Dict:
        """
        Analisa o comportamento de uma entidade para detectar padrões fraudulentos no contexto angolano.
        
        Args:
            entity_id: ID da entidade (usuário, conta, dispositivo, etc)
            entity_type: Tipo de entidade (user, account, device, merchant, etc)
            transaction_data: Dados da transação ou ação sendo analisada
            context_data: Informações contextuais adicionais
            use_cache: Se deve usar cache para resultados prévios
            
        Returns:
            Resultado da análise de comportamento com score de risco
        """
        # Registrar timestamp de início para métricas
        start_time = time.time()
        
        # Preparar dados de contexto
        context = context_data or {}
        
        # Verificar cache se solicitado
        if use_cache:
            cache_key = self.get_cache_key(entity_id, f"behavior_{entity_type}", context_data)
            cached_result = self.get_cached_result(cache_key)
            
            if cached_result:
                logger.info(f"Resultado em cache encontrado para {entity_id}")
                return cached_result
        
        # Combinar diferentes tipos de análise para obter um score completo
        results = {}
        risk_factors = []
        
        # Analisar dispositivo se houver dados de dispositivo
        if "device_data" in transaction_data or "device_id" in transaction_data:
            device_data = transaction_data.get("device_data", {})
            if "device_id" in transaction_data and not device_data:
                device_data = {"device_id": transaction_data["device_id"]}
                
            device_result = self.analyze_device_behavior(
                entity_id, 
                device_data, 
                session_data=context.get("session_data")
            )
            results["device"] = device_result
            
            if device_result.get("risk_factors"):
                risk_factors.extend(device_result["risk_factors"])
        
        # Analisar localização se houver dados de localização
        if "location" in transaction_data or "location_data" in transaction_data:
            location_data = transaction_data.get("location_data") or transaction_data.get("location", {})
            location_result = self.detect_location_anomalies(
                entity_id, 
                location_data, 
                history=context.get("location_history")
            )
            results["location"] = location_result
            
            if location_result.get("risk_factors"):
                risk_factors.extend(location_result["risk_factors"])
        
        # Analisar dados da conta
        account_data = transaction_data.get("account_data", {})
        if entity_type == "account" or account_data:
            if not account_data and entity_type == "account":
                # Se não temos dados de conta mas o tipo de entidade é conta, tentamos buscar dados
                for adapter_name, adapter in self.data_adapters.items():
                    if hasattr(adapter, "get_account_data"):
                        try:
                            account_data = adapter.get_account_data(entity_id)
                            if account_data:
                                break
                        except:
                            pass
            
            if account_data:
                account_result = self.evaluate_account_risk(
                    entity_id, 
                    account_data, 
                    history_data=context.get("account_history")
                )
                results["account"] = account_result
                
                if account_result.get("risk_factors"):
                    risk_factors.extend(account_result["risk_factors"])
        
        # Analisar transações se tivermos histórico
        if "transactions" in context or "transaction_history" in context:
            transactions = context.get("transactions") or context.get("transaction_history", [])
            # Adicionar a transação atual ao histórico para análise
            if transaction_data.get("transaction_details"):
                transactions = transactions + [transaction_data["transaction_details"]]
                
            transaction_result = self.analyze_transaction_pattern(
                entity_id, 
                transactions, 
                context_data=context
            )
            results["transaction"] = transaction_result
            
            if transaction_result.get("risk_factors"):
                risk_factors.extend(transaction_result["risk_factors"])
        
        # Buscar fatores de risco regionais
        regional_result = self.get_regional_risk_factors(entity_id, entity_type)
        results["regional"] = regional_result
        
        if regional_result.get("risk_factors"):
            risk_factors.extend(regional_result["risk_factors"])
        
        # Calcular score de risco combinado
        risk_score = self._calculate_combined_risk_score(results)
        
        # Determinar nível de risco com base no score
        risk_level = "low"
        if risk_score >= self.config["threshold_high"]:
            risk_level = "high"
        elif risk_score >= self.config["threshold_medium"]:
            risk_level = "medium"
            
        # Determinar ação recomendada
        recommended_action = "allow"
        if risk_level == "high":
            recommended_action = "block"
        elif risk_level == "medium":
            recommended_action = "verify"
        
        # Consolidar resultado final
        result = {
            "entity_id": entity_id,
            "entity_type": entity_type,
            "risk_score": risk_score,
            "risk_level": risk_level,
            "recommended_action": recommended_action,
            "risk_factors": risk_factors,
            "analysis_type": "behavioral",
            "region_code": "AO",
            "analysis_timestamp": datetime.now().isoformat(),
            "data_sources": list(self.data_adapters.keys()),
            "details": results
        }
        
        # Salvar em cache se solicitado
        if use_cache:
            self.save_to_cache(cache_key, result)
        
        # Atualizar métricas
        self.update_metrics(start_time, risk_level in ["medium", "high"])
        
        return result
    
    def analyze_transaction_pattern(self, entity_id: str, transactions: List[Dict], 
                                 context_data: Optional[Dict] = None) -> Dict:
        """
        Analisa padrões em múltiplas transações para detectar anomalias no contexto angolano.
        
        Args:
            entity_id: ID da entidade
            transactions: Lista de transações para análise
            context_data: Informações contextuais adicionais
            
        Returns:
            Resultado da análise de padrões com anomalias identificadas
        """
        # Verificar se temos transações suficientes para análise
        if not transactions or len(transactions) < 2:
            return {
                "risk_score": 0.1,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"reason": "Histórico de transações insuficiente para análise"}
            }
        
        # Preparar dados para análise
        try:
            # Converter lista de transações para DataFrame para facilitar análise
            df = pd.DataFrame(transactions)
            
            # Verificar colunas necessárias
            required_columns = ["amount", "timestamp"]
            for col in required_columns:
                if col not in df.columns:
                    return {
                        "risk_score": 0.2,
                        "risk_level": "low",
                        "risk_factors": [],
                        "details": {"reason": f"Dados de transação incompletos - falta coluna {col}"}
                    }
            
            # Converter timestamps para datetime
            df["timestamp"] = pd.to_datetime(df["timestamp"])
            
            # Ordenar por timestamp
            df = df.sort_values("timestamp")
            
            # Extrair padrões típicos para Angola
            transaction_patterns = self.regional_patterns["transaction_patterns"]
            
            # Inicializar fatores de risco
            risk_factors = []
            
            # 1. Verificar horários das transações
            df["hour"] = df["timestamp"].dt.hour
            high_risk_hours = []
            for time_range in transaction_patterns["high_risk_times"]:
                start_h, end_h = map(int, time_range.split("-"))
                high_risk_hours.extend(list(range(start_h, end_h + 1)))
            
            high_risk_hour_transactions = df[df["hour"].isin(high_risk_hours)]
            if len(high_risk_hour_transactions) > 0:
                risk_factors.append({
                    "factor": "transaction_unusual_hours",
                    "description": f"Transações em horários de alto risco ({len(high_risk_hour_transactions)} ocorrências)",
                    "weight": 0.7
                })
            
            # 2. Verificar frequência anormal
            time_delta = df["timestamp"].max() - df["timestamp"].min()
            days_span = time_delta.total_seconds() / (60 * 60 * 24)
            
            if days_span < 1:
                days_span = 1  # Mínimo de 1 dia para evitar divisão por zero
                
            transaction_count = len(df)
            daily_frequency = transaction_count / days_span
            
            # Comparar com padrões típicos de Angola
            expected_frequency = transaction_patterns["typical_transaction_frequency"]["daily"]
            
            if daily_frequency > (expected_frequency * 2):
                risk_factors.append({
                    "factor": "high_transaction_frequency",
                    "description": f"Frequência de transações muito alta ({daily_frequency:.1f} vs normal {expected_frequency:.1f} por dia)",
                    "weight": 0.8
                })
            
            # 3. Verificar valores anormais
            amount_mean = df["amount"].mean()
            amount_std = df["amount"].std() if len(df) > 1 else 0
            amount_max = df["amount"].max()
            
            # Verificar transações com valores muito acima da média
            if amount_std > 0:
                outliers = df[df["amount"] > (amount_mean + 2 * amount_std)]
                if len(outliers) > 0:
                    risk_factors.append({
                        "factor": "transaction_amount_outliers",
                        "description": f"Transações com valores anormalmente altos ({len(outliers)} ocorrências)",
                        "weight": 0.6
                    })
            
            # 4. Verificar padrões repetitivos exatos (possível fraude automatizada)
            value_counts = df["amount"].value_counts()
            repeated_amounts = value_counts[value_counts > 3].index.tolist()
            
            if repeated_amounts:
                risk_factors.append({
                    "factor": "repetitive_exact_amounts",
                    "description": f"Padrão repetitivo em valores de transação ({len(repeated_amounts)} valores repetidos)",
                    "weight": 0.5
                })
            
            # 5. Verificar transações internacionais (se disponível)
            if "country" in df.columns:
                foreign_transactions = df[df["country"] != "AO"]
                if len(foreign_transactions) > 0:
                    risk_factors.append({
                        "factor": "international_transactions",
                        "description": f"Transações internacionais detectadas ({len(foreign_transactions)} ocorrências)",
                        "weight": 0.9
                    })
            
            # Calcular score de risco com base nos fatores identificados
            risk_score = sum(factor["weight"] for factor in risk_factors) / 5  # Normalizar para 0-1
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
                    "transaction_count": transaction_count,
                    "time_span_days": days_span,
                    "daily_frequency": daily_frequency,
                    "amount_statistics": {
                        "mean": amount_mean,
                        "max": amount_max,
                        "std": amount_std
                    }
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao analisar padrões de transação: {str(e)}")
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
        no contexto angolano.
        
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
                    "weight": 0.9
                })
            
            # Verificar validação bancária/financeira
            has_valid_bank = account_data.get("has_valid_bank_account", False)
            if not has_valid_bank:
                risk_factors.append({
                    "factor": "no_bank_account",
                    "description": "Sem conta bancária verificada",
                    "weight": 0.5
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
            
            # Verificar se há dados específicos de Angola
            # Verificação de BI/NIF angolano
            has_bi = "bi_number" in account_data
            has_nif = "nif_number" in account_data
            
            if not has_bi and not has_nif:
                risk_factors.append({
                    "factor": "missing_angola_id",
                    "description": "Falta identificação angolana (BI/NIF)",
                    "weight": 0.75
                })
            
            # Verificar tipo de atividade econômica da conta (comum em Angola)
            activity_type = account_data.get("economic_activity", "unknown").lower()
            high_risk_activities = ["cambista informal", "comércio informal", "mineração", "importação"]
            
            if activity_type in high_risk_activities or activity_type == "unknown":
                risk_factors.append({
                    "factor": "high_risk_activity",
                    "description": f"Atividade econômica de alto risco: {activity_type}",
                    "weight": 0.7
                })
            
            # Verificar se a conta foi criada em Angola ou no exterior
            account_origin = account_data.get("origin_country", "unknown")
            if account_origin != "AO" and account_origin != "unknown":
                risk_factors.append({
                    "factor": "foreign_account_creation",
                    "description": f"Conta criada fora de Angola ({account_origin})",
                    "weight": 0.6
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
            logger.error(f"Erro ao avaliar risco da conta: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao avaliar risco da conta: {str(e)}"}
            }
    
    def detect_location_anomalies(self, entity_id: str, location_data: Dict, 
                               history: Optional[List[Dict]] = None) -> Dict:
        """
        Detecta anomalias de localização com sensibilidade ao contexto regional angolano.
        
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
            
            # Obter padrões de localização de Angola
            location_patterns = self.regional_patterns["location_patterns"]
            
            # 1. Verificar se a localização atual está em área de alto risco
            current_city = location_data.get("city", "").strip()
            current_district = location_data.get("district", "").strip()
            current_country = location_data.get("country", "").strip()
            
            # Verificar se está fora de Angola
            if current_country and current_country != "AO" and current_country != "Angola":
                risk_factors.append({
                    "factor": "foreign_location",
                    "description": f"Localização fora de Angola ({current_country})",
                    "weight": 0.8
                })
            
            # Verificar áreas de alto risco em Angola
            high_risk_areas = location_patterns["high_risk_areas"]
            
            if current_district and any(area.lower() in current_district.lower() for area in high_risk_areas):
                risk_factors.append({
                    "factor": "high_risk_area",
                    "description": f"Localização em área de alto risco: {current_district}",
                    "weight": 0.7
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
                    current_timestamp = datetime.fromisoformat(location_data.get("timestamp", datetime.now().isoformat()))
                    time_diff = (current_timestamp - last_timestamp).total_seconds() / 3600  # em horas
                    
                    # Evitar divisão por zero
                    if time_diff <= 0:
                        time_diff = 0.01
                    
                    # Calcular velocidade de deslocamento em km/h
                    speed = distance / time_diff
                    
                    # Verificar se a velocidade é anormalmente alta (considerando padrões angolanos)
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
                            "weight": 0.6
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
                if distance > 100:
                    risk_factors.append({
                        "factor": "ip_gps_mismatch",
                        "description": f"Divergência entre localização GPS e IP ({distance:.1f} km)",
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
                        "city": current_city,
                        "district": current_district
                    },
                    "has_history": bool(history and len(history) > 0),
                    "historical_points": len(history) if history else 0
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de localização: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao detectar anomalias de localização: {str(e)}"}
            }
            
    def analyze_device_behavior(self, entity_id: str, device_data: Dict, 
                             session_data: Optional[Dict] = None) -> Dict:
        """
        Analisa comportamento de dispositivos para detectar padrões suspeitos no contexto angolano.
        
        Args:
            entity_id: ID da entidade
            device_data: Dados do dispositivo
            session_data: Dados da sessão atual
            
        Returns:
            Resultado da análise de comportamento do dispositivo
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Verificar se temos dados suficientes
            if not device_data or not isinstance(device_data, dict):
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Dados de dispositivo insuficientes para análise"}
                }
            
            # Obter padrões de uso de dispositivo para Angola
            device_patterns = self.regional_patterns["behavioral_patterns"]["device_usage"]
            
            # 1. Verificar se o dispositivo é conhecido para esta entidade
            device_id = device_data.get("device_id", "unknown")
            device_known = device_data.get("is_known", False)
            
            if not device_known and device_id != "unknown":
                risk_factors.append({
                    "factor": "new_device",
                    "description": "Dispositivo não reconhecido anteriormente",
                    "weight": self.regional_risk_factors["new_device_login"]
                })
            
            # 2. Verificar se há simulação ou emulação de dispositivo
            if device_data.get("is_emulator", False) or device_data.get("is_rooted", False) or device_data.get("is_simulator", False):
                risk_factors.append({
                    "factor": "compromised_device",
                    "description": f"Dispositivo comprometido ou emulado",
                    "weight": 0.9
                })
            
            # 3. Verificar assinaturas de segurança do dispositivo
            security_flags = []
            
            if device_data.get("is_vpn", False):
                security_flags.append("VPN ativo")
            
            if device_data.get("is_proxy", False):
                security_flags.append("Proxy detectado")
            
            if device_data.get("is_tor", False):
                security_flags.append("Rede Tor detectada")
                
            if device_data.get("is_public_network", False):
                security_flags.append("Rede pública")
            
            if security_flags:
                risk_factors.append({
                    "factor": "security_flags",
                    "description": f"Sinais de segurança suspeitos: {', '.join(security_flags)}",
                    "weight": 0.8
                })
            
            # 4. Verificar inconsistências de navegador/dispositivo
            browser_inconsistencies = []
            
            if "browser" in device_data and "os" in device_data:
                browser = device_data["browser"].lower() if isinstance(device_data["browser"], str) else ""
                os_name = device_data["os"].lower() if isinstance(device_data["os"], str) else ""
                
                # Inconsistências conhecidas
                if "safari" in browser and not ("ios" in os_name or "mac" in os_name):
                    browser_inconsistencies.append("Safari em sistema não-Apple")
                
                if "edge" in browser and not "windows" in os_name:
                    browser_inconsistencies.append("Edge em sistema não-Windows")
            
            if browser_inconsistencies:
                risk_factors.append({
                    "factor": "browser_os_mismatch",
                    "description": f"Inconsistências navegador/OS: {', '.join(browser_inconsistencies)}",
                    "weight": 0.75
                })
            
            # 5. Verificar comportamento de sessão atípico (se dados de sessão disponíveis)
            if session_data and isinstance(session_data, dict):
                # Sessão típica em Angola (de acordo com padrões)
                typical_session_duration = device_patterns["typical_session_duration_min"]
                
                # Duração anormal de sessão
                session_duration = session_data.get("duration_minutes", 0)
                if session_duration > (typical_session_duration * 5):
                    risk_factors.append({
                        "factor": "abnormal_session_length",
                        "description": f"Sessão anormalmente longa ({session_duration} min vs típico {typical_session_duration} min)",
                        "weight": 0.6
                    })
                
                # Verificar se há ações simultâneas em dispositivos diferentes
                concurrent_sessions = session_data.get("concurrent_sessions", 0)
                if concurrent_sessions > 1:
                    risk_factors.append({
                        "factor": "concurrent_sessions",
                        "description": f"Sessões simultâneas em múltiplos dispositivos ({concurrent_sessions})",
                        "weight": 0.85
                    })
                
                # Verificar falhas de autenticação recentes
                auth_failures = session_data.get("recent_auth_failures", 0)
                if auth_failures >= 3:
                    risk_factors.append({
                        "factor": "multiple_auth_failures",
                        "description": f"Múltiplas falhas de autenticação recentes ({auth_failures})",
                        "weight": self.regional_risk_factors["multiple_auth_failures"]
                    })
            
            # 6. Verificar contexto específico de Angola
            
            # Verificar compatibilidade com dispositivos/navegadores comuns em Angola
            if "device_type" in device_data and "device_model" in device_data:
                device_type = device_data["device_type"].lower() if isinstance(device_data["device_type"], str) else ""
                
                # Em Angola, dispositivos móveis são mais comuns que desktops
                if device_type == "desktop" and not device_data.get("is_known", False):
                    # Desktop não reconhecido é menos comum em Angola (maior predominância mobile)
                    if device_patterns["mobile_predominance"] > 0.7:
                        risk_factors.append({
                            "factor": "uncommon_device_type",
                            "description": f"Tipo de dispositivo menos comum para o mercado angolano",
                            "weight": 0.5
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
                    "device_id": device_id,
                    "device_known": device_known,
                    "device_type": device_data.get("device_type", "unknown"),
                    "security_flags": security_flags,
                    "has_session_data": bool(session_data)
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao analisar comportamento do dispositivo: {str(e)}")
            return {
                "risk_score": 0.3,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao analisar comportamento do dispositivo: {str(e)}"}
            }
    
    def get_regional_risk_factors(self, entity_id: str, entity_type: str) -> Dict:
        """
        Obtém fatores de risco específicos da região de Angola para uma entidade.
        
        Args:
            entity_id: ID da entidade
            entity_type: Tipo de entidade
            
        Returns:
            Fatores de risco regionais
        """
        # Inicializar fatores de risco
        risk_factors = []
        
        try:
            # Buscar dados específicos de entidade em fontes regionais
            regional_data = {}
            regional_factors_found = False
            
            # Tentar buscar em adaptadores disponíveis
            for adapter_name, adapter in self.data_adapters.items():
                try:
                    # Verificar se o adaptador tem o método adequado
                    if hasattr(adapter, "get_entity_risk_data"):
                        entity_data = adapter.get_entity_risk_data(entity_id, entity_type)
                        if entity_data:
                            regional_data[adapter_name] = entity_data
                            regional_factors_found = True
                except Exception as e:
                    logger.warning(f"Erro ao buscar dados de {adapter_name}: {str(e)}")
            
            # Se não encontramos dados específicos, retornar resultado padrão baixo
            if not regional_factors_found:
                return {
                    "risk_score": 0.1,
                    "risk_level": "low",
                    "risk_factors": [],
                    "details": {"reason": "Sem fatores de risco regionais específicos identificados"}
                }
            
            # Analisar fatores de risco específicos de Angola
            
            # 1. Verificar se a entidade está em listas de restrição (AML, PEP, sanções)
            bna_data = regional_data.get("bna", {})
            bureau_data = regional_data.get("bureau_credito", {})
            
            if bna_data.get("in_restriction_list", False):
                risk_factors.append({
                    "factor": "bna_restriction_list",
                    "description": f"Entidade presente em lista de restrição do BNA",
                    "weight": 0.95
                })
            
            if bna_data.get("is_pep", False):
                risk_factors.append({
                    "factor": "pep_status",
                    "description": f"Entidade classificada como PEP (Pessoa Politicamente Exposta)",
                    "weight": 0.7
                })
            
            # 2. Verificar score de crédito em bureau angolano
            credit_score = bureau_data.get("credit_score", 0)
            if credit_score > 0:
                # Normalizar para escala de 0-100 se necessário
                if credit_score > 100:
                    normalized_score = credit_score / 10
                else:
                    normalized_score = credit_score
                
                # Score baixo é fator de risco
                if normalized_score < 30:
                    risk_factors.append({
                        "factor": "low_credit_score",
                        "description": f"Score de crédito muito baixo ({normalized_score}/100)",
                        "weight": 0.8
                    })
                elif normalized_score < 50:
                    risk_factors.append({
                        "factor": "medium_credit_score",
                        "description": f"Score de crédito abaixo da média ({normalized_score}/100)",
                        "weight": 0.5
                    })
            
            # 3. Verificar histórico de documentos fraudulentos
            senica_data = regional_data.get("senica", {})
            document_risk = senica_data.get("document_risk_level", "unknown")
            
            if document_risk == "high":
                risk_factors.append({
                    "factor": "document_fraud_history",
                    "description": "Histórico de documentação fraudulenta detectado",
                    "weight": 0.9
                })
            elif document_risk == "medium":
                risk_factors.append({
                    "factor": "document_irregularities",
                    "description": "Irregularidades documentais detectadas",
                    "weight": 0.6
                })
            
            # 4. Verificar comportamentos de telefonia
            telecom_data = regional_data.get("telecom", {})
            
            if telecom_data.get("sim_swap_recent", False):
                risk_factors.append({
                    "factor": "recent_sim_swap",
                    "description": "Troca de SIM recente detectada",
                    "weight": 0.8
                })
            
            if telecom_data.get("multiple_sim_registrations", False):
                risk_factors.append({
                    "factor": "multiple_sims",
                    "description": "Múltiplos SIMs registrados no mesmo período",
                    "weight": 0.7
                })
            
            if telecom_data.get("phone_in_blacklist", False):
                risk_factors.append({
                    "factor": "blacklisted_phone",
                    "description": "Número de telefone em lista de restrição",
                    "weight": 0.85
                })
            
            # 5. Verificar fatores de região/localização
            common_regions = telecom_data.get("common_regions", [])
            suspicious_regions = ["Zaire", "Cabinda", "Lunda Norte", "Lunda Sul", "Uíge"]
            
            if any(region in suspicious_regions for region in common_regions):
                risk_factors.append({
                    "factor": "high_risk_region",
                    "description": f"Operações em regiões de alto risco: {', '.join([r for r in common_regions if r in suspicious_regions])}",
                    "weight": 0.6
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
                    "data_sources": list(regional_data.keys()),
                    "regional_data_found": regional_factors_found
                }
            }
            
        except Exception as e:
            logger.error(f"Erro ao obter fatores de risco regionais: {str(e)}")
            return {
                "risk_score": 0.2,
                "risk_level": "low",
                "risk_factors": [],
                "details": {"error": f"Erro ao obter fatores de risco regionais: {str(e)}"}
            }
            
    def _calculate_combined_risk_score(self, results: Dict) -> float:
        """
        Calcula um score de risco combinado a partir de múltiplos resultados de análise.
        
        Args:
            results: Dicionário com resultados de diferentes tipos de análise
            
        Returns:
            Score de risco combinado (0.0 a 1.0)
        """
        # Definir pesos para cada tipo de análise
        # Estes pesos podem ser ajustados com base em dados reais e feedback
        analysis_weights = {
            "device": 0.2,
            "location": 0.25,
            "account": 0.2,
            "transaction": 0.2,
            "regional": 0.15
        }
        
        # Inicializar score ponderado e soma de pesos
        weighted_score = 0.0
        total_weight = 0.0
        
        # Calcular score ponderado
        for analysis_type, weight in analysis_weights.items():
            if analysis_type in results:
                result = results[analysis_type]
                if "risk_score" in result:
                    weighted_score += result["risk_score"] * weight
                    total_weight += weight
        
        # Se não tivermos nenhum resultado, retornar score baixo padrão
        if total_weight == 0:
            return 0.2
        
        # Normalizar para a soma de pesos
        final_score = weighted_score / total_weight
        
        # Verificar se há algum resultado de alto risco, que deve elevar o score final
        has_high_risk = any(result.get("risk_level") == "high" for result in results.values() if isinstance(result, dict))
        
        if has_high_risk and final_score < 0.7:
            # Elevar o score se tivermos pelo menos um resultado de alto risco
            final_score = max(final_score, 0.7)
        
        # Garantir limites
        final_score = max(0.0, min(1.0, final_score))
        
        return final_score