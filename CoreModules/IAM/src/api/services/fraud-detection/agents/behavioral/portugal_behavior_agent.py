#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Agente de Análise Comportamental para Portugal

Este módulo implementa um agente especializado de detecção de fraudes
com adaptações específicas para o mercado português, considerando
padrões comportamentais, regulamentações da União Europeia e Portugal,
características culturais e dinâmicas econômicas específicas de Portugal.

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
logger = logging.getLogger("fraud_detection.behavioral.portugal")

class PortugalBehaviorAgent(BehaviorAnalysisAgent):
    """
    Agente de análise comportamental especializado para o mercado português.
    
    Esta classe implementa análise de comportamento para detecção de fraudes
    considerando fatores específicos de Portugal, incluindo:
    - Regulamentações do Banco de Portugal e União Europeia
    - Padrões comportamentais típicos em transações financeiras portuguesas
    - Características de uso de dispositivos em Portugal
    - Dados do Banco de Portugal e sistemas de crédito europeus
    - Validação de documentos portugueses (NIF, CC)
    - Considerações geográficas específicas de Portugal e UE
    
    Implementa todos os métodos abstratos da classe BehaviorAnalysisAgent
    com adaptações específicas para o contexto português e europeu.
    """
    
    def __init__(self, config_path: Optional[str] = None, 
                model_path: Optional[str] = None,
                cache_dir: Optional[str] = None,
                data_sources: Optional[List[str]] = None):
        """
        Inicializa o agente de comportamento português.
        
        Args:
            config_path: Caminho para arquivo de configuração
            model_path: Caminho para modelos treinados
            cache_dir: Diretório para armazenamento de cache
            data_sources: Lista de fontes de dados a utilizar
        """
        # Chamar inicialização da classe pai
        super().__init__(config_path, model_path, cache_dir, data_sources)
        
        # Definir região para Portugal
        self.region = "PT"
        
        # Carregar configurações específicas de Portugal se não fornecido
        if not config_path:
            default_config = os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "config",
                "portugal_config.json"
            )
            if os.path.exists(default_config):
                self.config_path = default_config
        
        # Inicializar adaptadores específicos para Portugal
        self._init_portugal_adapters()
        
        # Carregar padrões regionais para Portugal
        self._load_portugal_patterns()
        
        # Fatores de risco específicos de Portugal
        self.regional_risk_factors = {
            "new_device_login": 0.6,
            "multiple_auth_failures": 0.7,
            "unusual_transaction_time": 0.65,
            "multiple_card_registrations": 0.75,
            "foreign_ip_access": 0.6,
            "high_risk_zone": 0.6,
            "nif_blacklist": 0.85,
            "banco_portugal_restrictions": 0.8,
            "eu_sanctions_list": 0.9,
            "device_fraud_history": 0.8
        }
        
        # Inicializar modelos específicos para Portugal
        self._load_portugal_models()
        
        logger.info(f"Agente de análise comportamental de Portugal inicializado. Versão: 1.0.0")
    
    def _init_portugal_adapters(self):
        """Inicializa adaptadores de dados específicos para Portugal"""
        try:
            # Lista de adaptadores a serem inicializados
            portugal_adapters = {
                "banco_portugal": "BancoPortugalAdapter",
                "autoridade_tributaria": "AutoridadeTributariaAdapter",
                "schengen_info": "SchengenInfoAdapter",
                "telecom_portugal": "TelecomPortugalAdapter"
            }
            
            # Inicializar adaptadores selecionados ou todos se nenhum foi especificado
            adapter_names = self.config.get("data_sources", list(portugal_adapters.keys()))
            
            for adapter_name in adapter_names:
                if adapter_name in portugal_adapters:
                    try:
                        # Caminho dinâmico para importação dos adaptadores
                        adapter_module = f"...adapters.portugal.{adapter_name.lower()}_adapter"
                        adapter_class = portugal_adapters[adapter_name]
                        
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
                logger.warning("Nenhum adaptador de dados foi inicializado para Portugal")
                
        except Exception as e:
            logger.error(f"Falha na inicialização dos adaptadores para Portugal: {str(e)}")
    
    def _load_portugal_patterns(self):
        """Carrega padrões comportamentais específicos de Portugal"""
        try:
            # Tentar carregar padrões de um arquivo
            patterns_path = os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "patterns",
                "portugal_patterns.json"
            )
            
            if os.path.exists(patterns_path):
                with open(patterns_path, 'r', encoding='utf-8') as f:
                    self.regional_patterns = json.load(f)
                logger.info(f"Padrões regionais de Portugal carregados de {patterns_path}")
            else:
                # Definir padrões padrão se o arquivo não existe
                self.regional_patterns = {
                    "transaction_patterns": {
                        "typical_transaction_amount": {
                            "p2p_transfer": {
                                "mean": 200.00,
                                "std_dev": 150.00,
                                "max_normal": 1500.00
                            },
                            "bill_payment": {
                                "mean": 120.00,
                                "std_dev": 100.00,
                                "max_normal": 1000.00
                            },
                            "retail_purchase": {
                                "mean": 80.00,
                                "std_dev": 70.00,
                                "max_normal": 800.00
                            }
                        },
                        "peak_transaction_hours": [10, 13, 16, 19],
                        "low_activity_hours": [2, 3, 4, 5],
                        "weekend_usage_factor": 0.8,
                        "month_end_increase_factor": 1.4,
                        "common_transaction_frequencies": {
                            "daily": 1,
                            "weekly": 4,
                            "monthly": 15
                        },
                        "high_risk_merchants_categories": [
                            "apostas_online",
                            "casinos_online",
                            "criptomoeda",
                            "transferencias_internacionais_nao_identificadas"
                        ],
                        "high_risk_regions": [
                            "fora_ue",
                            "zonas_fiscais_privilegiadas"
                        ],
                        "mb_way_specific_patterns": {
                            "typical_frequency_daily": 2,
                            "max_normal_amount": 1000.00,
                            "suspicious_time_gap_seconds": 45
                        }
                    },
                    "behavioral_patterns": {
                        "device_usage": {
                            "mobile_predominance": 0.65,
                            "typical_session_duration_min": 10,
                            "common_device_change_frequency_days": 210,
                            "max_normal_devices_per_user": 3,
                            "typical_auth_methods": ["password", "biometria", "codigo_sms", "cartao_cidadao"]
                        },
                        "login_patterns": {
                            "typical_login_frequency_days": 4,
                            "typical_login_hours": [8, 22],
                            "suspicious_login_attempts_threshold": 3
                        }
                    },
                    "location_patterns": {
                        "high_risk_areas": [
                            "Cova da Moura", "Bairro do Zambujal", "Bairro da Boavista",
                            "Bairro do Cerco", "Bairro do Aleixo", "Vale da Amoreira"
                        ],
                        "common_movement_radius_km": 50,
                        "typical_speed_kmh": 100,
                        "district_risk_factors": {
                            "Lisboa": 0.45, "Porto": 0.45, "Setúbal": 0.4, "Faro": 0.35,
                            "Aveiro": 0.3, "Braga": 0.3, "Leiria": 0.25, "Coimbra": 0.25,
                            "Santarém": 0.25, "Viana do Castelo": 0.2, "Vila Real": 0.2,
                            "Bragança": 0.2, "Viseu": 0.25, "Guarda": 0.2, "Castelo Branco": 0.2,
                            "Portalegre": 0.2, "Évora": 0.25, "Beja": 0.25, "Madeira": 0.3, "Açores": 0.25
                        },
                        "eu_border_risk": 0.4,
                        "non_eu_border_risk": 0.6
                    }
                }
                logger.warning(f"Arquivo de padrões para Portugal não encontrado. Usando padrões padrão.")
                
        except Exception as e:
            logger.error(f"Erro ao carregar padrões regionais de Portugal: {str(e)}")
            # Definir padrões mínimos em caso de erro
            self.regional_patterns = {
                "transaction_patterns": {"typical_amount": 100.0},
                "behavioral_patterns": {"device_usage": {"mobile_predominance": 0.6}},
                "location_patterns": {"high_risk_areas": []}
            }
    
    def _load_portugal_models(self):
        """Carrega modelos de ML específicos para Portugal"""
        try:
            # Verificar diretório de modelos
            if not self.model_path:
                logger.warning("Caminho de modelos não definido. Usando heurísticas.")
                return
            
            # Definir caminho específico para modelos de Portugal
            portugal_models_path = os.path.join(self.model_path, "portugal")
            
            # Verificar e carregar modelos específicos (quando disponíveis)
            model_files = {
                "transaction_risk": "portugal_transaction_risk_model.pkl",
                "account_risk": "portugal_account_risk_model.pkl",
                "location_risk": "portugal_location_anomaly_model.pkl",
                "device_risk": "portugal_device_behavior_model.pkl"
            }
            
            # Carregar modelos disponíveis
            for model_type, model_file in model_files.items():
                model_path = os.path.join(portugal_models_path, model_file)
                if os.path.exists(model_path):
                    try:
                        # Aqui usaria uma função para carregar o modelo de acordo com seu tipo
                        # self.models[model_type] = load_model(model_path)
                        logger.info(f"Modelo {model_type} para Portugal carregado com sucesso")
                    except Exception as e:
                        logger.error(f"Erro ao carregar modelo {model_type}: {str(e)}")
                else:
                    logger.warning(f"Modelo {model_type} não encontrado. Usando regras heurísticas.")
                    
        except Exception as e:
            logger.error(f"Erro ao carregar modelos de Portugal: {str(e)}")
    
    def evaluate_account_risk(self, account_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Avalia o risco associado à conta com base nos padrões portugueses.
        
        Considera fatores como:
        - Estado de verificação KYC e idade da conta
        - Validação do NIF e Cartão de Cidadão
        - Restrições no Banco de Portugal
        - Histórico de atividades suspeitas
        - Perfil de risco baseado em dados da União Europeia
        
        Args:
            account_data: Dicionário com dados da conta a analisar
            
        Returns:
            Dicionário com pontuação de risco e fatores de risco identificados
        """
        try:
            logger.info(f"Avaliando risco da conta: {account_data.get('account_id', 'Desconhecido')}")
            
            # Inicializar resultados
            risk_score = 0.0
            risk_factors = []
            
            # 1. Verificar idade da conta
            account_age_days = self._calculate_account_age(account_data.get("created_at"))
            if account_age_days < 30:
                risk_score += 0.7
                risk_factors.append({"factor": "conta_recente", "score": 0.7, "details": f"Conta com {account_age_days} dias"})
            elif account_age_days < 90:
                risk_score += 0.4
                risk_factors.append({"factor": "conta_nova", "score": 0.4, "details": f"Conta com {account_age_days} dias"})
                
            # 2. Verificar status de KYC
            kyc_status = account_data.get("kyc_status", "unverified")
            if kyc_status == "unverified":
                risk_score += 0.8
                risk_factors.append({"factor": "kyc_nao_verificado", "score": 0.8})
            elif kyc_status == "pending":
                risk_score += 0.5
                risk_factors.append({"factor": "kyc_pendente", "score": 0.5})
            elif kyc_status == "rejected":
                risk_score += 0.9
                risk_factors.append({"factor": "kyc_rejeitado", "score": 0.9})
                
            # 3. Verificar presença de NIF
            nif = account_data.get("tax_id")
            if not nif:
                risk_score += 0.7
                risk_factors.append({"factor": "nif_ausente", "score": 0.7})
            elif not self._validate_portuguese_nif(nif):
                risk_score += 0.8
                risk_factors.append({"factor": "nif_invalido", "score": 0.8, "details": f"NIF: {nif}"})
                
            # 4. Verificar presença de Cartão de Cidadão
            cc = account_data.get("id_document", {}).get("number")
            if not cc:
                risk_score += 0.5
                risk_factors.append({"factor": "cc_ausente", "score": 0.5})
            elif not self._validate_portuguese_cc(cc):
                risk_score += 0.7
                risk_factors.append({"factor": "cc_invalido", "score": 0.7, "details": f"CC: {cc}"})
                
            # 5. Verificar histórico de atividade suspeita
            suspicious_activities = account_data.get("suspicious_activity_history", [])
            if suspicious_activities:
                activity_count = len(suspicious_activities)
                recent_activities = [a for a in suspicious_activities 
                                     if self._is_activity_recent(a.get("timestamp"))]
                
                if recent_activities:
                    risk_score += min(0.9, 0.5 + 0.1 * len(recent_activities))
                    risk_factors.append({
                        "factor": "atividades_suspeitas_recentes", 
                        "score": min(0.9, 0.5 + 0.1 * len(recent_activities)),
                        "details": f"{len(recent_activities)} atividades nos últimos 30 dias"
                    })
                elif activity_count > 0:
                    risk_score += 0.3
                    risk_factors.append({
                        "factor": "atividades_suspeitas_historicas", 
                        "score": 0.3,
                        "details": f"{activity_count} atividades históricas"
                    })
                    
            # 6. Verificar restrições no Banco de Portugal
            if "banco_portugal" in self.data_adapters:
                try:
                    bp_restrictions = self.data_adapters["banco_portugal"].check_account_restrictions(
                        nif=nif, 
                        name=account_data.get("name"),
                        document=cc
                    )
                    
                    if bp_restrictions.get("has_restrictions"):
                        restriction_type = bp_restrictions.get("restriction_type", "desconhecida")
                        risk_score += 0.9
                        risk_factors.append({
                            "factor": "restricao_banco_portugal", 
                            "score": 0.9,
                            "details": f"Tipo: {restriction_type}"
                        })
                except Exception as e:
                    logger.error(f"Erro ao verificar restrições no Banco de Portugal: {str(e)}")
            
            # 7. Verificar lista de sanções da União Europeia
            eu_sanctioned = self._check_eu_sanctions_list(
                name=account_data.get("name"),
                nif=nif,
                cc=cc
            )
            
            if eu_sanctioned:
                risk_score += 1.0
                risk_factors.append({
                    "factor": "lista_sancoes_ue",
                    "score": 1.0,
                    "details": "Presente na lista de sanções da UE"
                })
                
            # 8. Verificar perfil de risco do endereço
            address = account_data.get("address", {})
            address_risk = self._evaluate_address_risk_portugal(address)
            
            if address_risk > 0.3:
                risk_score += address_risk
                risk_factors.append({
                    "factor": "risco_endereco",
                    "score": address_risk,
                    "details": f"Endereço em zona de risco: {address.get('district', 'Desconhecido')}"
                })
            
            # 9. Avaliar tipo de atividade econômica
            economic_activity = account_data.get("economic_activity")
            if economic_activity in self._get_high_risk_activities_portugal():
                risk_score += 0.5
                risk_factors.append({
                    "factor": "atividade_economica_alto_risco",
                    "score": 0.5,
                    "details": f"Atividade: {economic_activity}"
                })
            
            # 10. Verificar status PEP (Pessoa Politicamente Exposta)
            is_pep = account_data.get("is_pep", False)
            pep_relatives = account_data.get("pep_relatives", [])
            
            if is_pep:
                risk_score += 0.7
                risk_factors.append({
                    "factor": "cliente_pep",
                    "score": 0.7
                })
            
            if pep_relatives:
                risk_score += 0.5
                risk_factors.append({
                    "factor": "familiar_pep",
                    "score": 0.5,
                    "details": f"{len(pep_relatives)} familiares PEP"
                })
                
            # Normalizar pontuação final entre 0 e 1
            final_risk_score = min(1.0, risk_score / 10.0)
            
            # Preparar resultado
            result = {
                "risk_score": final_risk_score,
                "risk_level": self._get_risk_level(final_risk_score),
                "risk_factors": risk_factors,
                "timestamp": datetime.now().isoformat(),
                "account_id": account_data.get("account_id", "unknown")
            }
            
            logger.info(f"Avaliação de conta concluída: {result['risk_level']} ({final_risk_score:.2f})")
            return result
            
        except Exception as e:
            logger.error(f"Erro na avaliação de risco da conta: {str(e)}")
            # Retornar avaliação de risco conservadora em caso de erro
            return {
                "risk_score": 0.7,
                "risk_level": "high",
                "risk_factors": [{"factor": "erro_avaliacao", "score": 0.7, "details": str(e)}],
                "timestamp": datetime.now().isoformat(),
                "account_id": account_data.get("account_id", "unknown")
            }
            
    def _validate_portuguese_nif(self, nif: str) -> bool:
        """Valida um NIF (Número de Identificação Fiscal) português"""
        # Remove espaços e hifens
        if not nif:
            return False
        
        nif = nif.replace(" ", "").replace("-", "")
        
        # NIF deve ter 9 dígitos
        if not nif.isdigit() or len(nif) != 9:
            return False
            
        # Algoritmo de validação do NIF português
        # Os primeiros dígitos têm significados especiais
        # 1 ou 2: pessoa singular
        # 5: pessoa coletiva
        # 6: pessoa coletiva pública
        # 8: empresário em nome individual
        # 9: pessoa coletiva irregular ou número temporário
        
        if nif[0] not in "125689":
            return False
            
        # Cálculo do dígito de controle
        total = 0
        for i in range(8):
            total += int(nif[i]) * (9 - i)
            
        check_digit = 11 - (total % 11)
        if check_digit >= 10:
            check_digit = 0
            
        return check_digit == int(nif[8])
        
    def _validate_portuguese_cc(self, cc: str) -> bool:
        """Valida um número de Cartão de Cidadão português"""
        if not cc:
            return False
            
        # Remove espaços e converte para maiúsculo
        cc = cc.replace(" ", "").upper()
        
        # Formato básico: 00000000 0 XX0
        # 8 dígitos + 1 dígito de controle + 2 letras + 1 dígito
        if len(cc) != 12:
            return False
            
        # Verificação básica de formato
        # Primeiros 8 caracteres devem ser dígitos
        if not cc[:8].isdigit():
            return False
            
        # 9º caractere deve ser dígito
        if not cc[8].isdigit():
            return False
            
        # 10º e 11º caracteres devem ser letras
        if not (cc[9].isalpha() and cc[10].isalpha()):
            return False
            
        # 12º caractere deve ser dígito
        if not cc[11].isdigit():
            return False
            
        # Nota: a validação completa do CC exigiria algoritmos mais complexos
        # Esta é uma validação simplificada de formato
        return True
        
    def _evaluate_address_risk_portugal(self, address: Dict[str, Any]) -> float:
        """Avalia o risco associado ao endereço em Portugal"""
        if not address:
            return 0.5  # Endereço ausente é risco moderado
            
        risk_score = 0.0
            
        # 1. Verificar distrito
        district = address.get("district", "").lower()
        district_risk = self.regional_patterns.get("location_patterns", {}).get(
            "district_risk_factors", {}).get(district.title(), 0.3)
        risk_score += district_risk
            
        # 2. Verificar áreas de alto risco
        address_str = " ".join([
            str(address.get("street", "")),
            str(address.get("neighborhood", "")),
            str(address.get("city", "")),
            str(address.get("district", ""))
        ]).lower()
            
        high_risk_areas = self.regional_patterns.get("location_patterns", {}).get("high_risk_areas", [])
        for area in high_risk_areas:
            if area.lower() in address_str:
                risk_score += 0.3
                break
                
        # 3. Verificar se o endereço é diferente do endereço fiscal
        if address.get("differs_from_fiscal_address", False):
            risk_score += 0.4
                
        # 4. Verificar proximidade de fronteira
        if self._is_border_area_portugal(district):
            risk_score += 0.3
                
        # 5. Verificar se é endereço temporário/hotel
        if address.get("is_temporary", False) or "hotel" in address_str or "hostel" in address_str:
            risk_score += 0.4
                
        return min(1.0, risk_score)
            
    def _is_border_area_portugal(self, district: str) -> bool:
        """Verifica se o distrito está em área de fronteira"""
        border_districts = ["viana do castelo", "braga", "vila real", "bragança", 
                           "guarda", "castelo branco", "portalegre", "évora", "beja", "faro"]
        return district.lower() in border_districts
            
    def _get_high_risk_activities_portugal(self) -> List[str]:
        """Retorna lista de atividades econômicas de alto risco em Portugal"""
        return [
            "apostas_e_jogos",
            "compra_e_venda_de_imoveis",
            "cambio_de_moeda",
            "comercio_de_bens_de_luxo",
            "comercio_de_ouro_e_metais_preciosos",
            "comercio_de_antiguidades",
            "intermediacao_imobiliaria",
            "casinos_e_jogos_de_fortuna_e_azar",
            "comercio_de_criptomoedas",
            "sociedades_offshore"
        ]
            
    def _check_eu_sanctions_list(self, name: str, nif: str = None, cc: str = None) -> bool:
        """Verifica se o cliente está na lista de sanções da União Europeia"""
        # Implementação básica - em produção, usaria uma API ou fonte de dados oficial
        try:
            if "schengen_info" in self.data_adapters:
                return self.data_adapters["schengen_info"].check_sanctions_list(
                    name=name,
                    document_id=cc,
                    tax_id=nif
                )
            return False
        except Exception as e:
            logger.error(f"Erro ao verificar lista de sanções da UE: {str(e)}")
            return False
            
    def detect_location_anomalies(self, location_data: Dict[str, Any], 
                                 user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Detecta anomalias de localização específicas para Portugal.
        
        Considera:
        - Padrões de movimentação típicos em Portugal
        - Movimentações rápidas entre distritos ou cidades
        - Acesso a partir de zonas de alto risco
        - Diferenças entre localização GPS e IP
        - Considerações de fronteira do Espaço Schengen
        
        Args:
            location_data: Dados atuais de localização do usuário
            user_history: Histórico de localizações e comportamento do usuário
            
        Returns:
            Dicionário com pontuação de risco e anomalias detectadas
        """
        try:
            logger.info(f"Analisando anomalias de localização para o usuário: {user_history.get('user_id', 'Desconhecido')}")
            
            # Inicializar resultados
            risk_score = 0.0
            anomalies = []
            
            # Extrair dados relevantes
            current_location = {
                "ip": location_data.get("ip_address"),
                "country": location_data.get("country"),
                "city": location_data.get("city"),
                "district": location_data.get("district", ""),
                "coords": {
                    "latitude": location_data.get("latitude"),
                    "longitude": location_data.get("longitude")
                },
                "timestamp": location_data.get("timestamp", datetime.now().isoformat())
            }
            
            # 1. Verificar se está em Portugal
            if current_location["country"] != "PT" and current_location["country"] != "Portugal":
                foreign_country_risk = 0.5
                # Verificar se está na UE
                if current_location["country"] in self._get_eu_countries():
                    foreign_country_risk = 0.3
                # Verificar se está em países CPLP
                elif current_location["country"] in self._get_cplp_countries():
                    foreign_country_risk = 0.4
                
                risk_score += foreign_country_risk
                anomalies.append({
                    "type": "acesso_fora_de_portugal",
                    "details": f"Localização: {current_location['country']}",
                    "risk": foreign_country_risk
                })
            
            # 2. Verificar histórico recente de localizações
            recent_locations = user_history.get("recent_locations", [])
            if recent_locations and len(recent_locations) > 0:
                last_location = recent_locations[0]
                
                # Verificar velocidade de deslocamento impossível
                if self._is_impossible_travel(last_location, current_location):
                    risk_score += 0.9
                    anomalies.append({
                        "type": "deslocamento_impossivel",
                        "details": f"De {last_location.get('city', 'desconhecido')} para {current_location['city']} em tempo impossível",
                        "risk": 0.9
                    })
                
                # Verificar mudança repentina de país
                if (last_location.get("country") == "PT" or last_location.get("country") == "Portugal") and \
                   (current_location["country"] != "PT" and current_location["country"] != "Portugal"):
                    risk_score += 0.6
                    anomalies.append({
                        "type": "mudanca_pais",
                        "details": f"De Portugal para {current_location['country']}",
                        "risk": 0.6
                    })
            
            # 3. Verificar distritos de alto risco em Portugal
            if current_location["country"] == "PT" or current_location["country"] == "Portugal":
                district_risk = self.regional_patterns.get("location_patterns", {}).get(
                    "district_risk_factors", {}).get(current_location["district"], 0.3)
                
                if district_risk > 0.4:
                    risk_score += district_risk
                    anomalies.append({
                        "type": "distrito_alto_risco",
                        "details": f"Distrito: {current_location['district']}",
                        "risk": district_risk
                    })
                    
                # Verificar áreas específicas de alto risco
                high_risk_areas = self.regional_patterns.get("location_patterns", {}).get("high_risk_areas", [])
                for area in high_risk_areas:
                    if area.lower() in current_location["city"].lower() or \
                       (current_location.get("neighborhood") and area.lower() in current_location["neighborhood"].lower()):
                        risk_score += 0.6
                        anomalies.append({
                            "type": "zona_alto_risco",
                            "details": f"Local: {area}",
                            "risk": 0.6
                        })
                        break
                        
                # Verificar distritos de fronteira
                if self._is_border_area_portugal(current_location["district"]):
                    risk_score += 0.3
                    anomalies.append({
                        "type": "distrito_fronteira",
                        "details": f"Distrito: {current_location['district']}",
                        "risk": 0.3
                    })
            
            # 4. Verificar discrepância entre localização GPS e IP
            ip_location = self._get_ip_location(current_location["ip"])
            if ip_location and self._is_significant_location_mismatch(
                {"latitude": current_location["coords"]["latitude"], 
                 "longitude": current_location["coords"]["longitude"]},
                {"latitude": ip_location["latitude"], 
                 "longitude": ip_location["longitude"]}
            ):
                risk_score += 0.7
                anomalies.append({
                    "type": "discrepancia_gps_ip",
                    "details": f"GPS: {current_location['city']}, IP: {ip_location.get('city', 'desconhecido')}",
                    "risk": 0.7
                })
            
            # 5. Verificar padrões de localização do usuário
            typical_locations = user_history.get("typical_locations", [])
            if typical_locations and not self._is_location_in_typical_places(
                current_location, typical_locations
            ):
                risk_score += 0.4
                anomalies.append({
                    "type": "localizacao_atipica",
                    "details": f"Local atual: {current_location['city']}",
                    "risk": 0.4
                })
            
            # 6. Verificar uso de VPN/Proxy
            if location_data.get("is_vpn", False) or location_data.get("is_proxy", False):
                risk_score += 0.6
                anomalies.append({
                    "type": "uso_vpn_proxy",
                    "details": f"Detectado {'VPN' if location_data.get('is_vpn') else 'Proxy'}",
                    "risk": 0.6
                })
                
            # 7. Verificar horário atípico para a localização
            if self._is_unusual_time_for_location(current_location):
                risk_score += 0.4
                anomalies.append({
                    "type": "horario_atipico_para_localizacao",
                    "details": "Acesso em horário incomum para o local",
                    "risk": 0.4
                })
            
            # Normalizar pontuação final
            final_risk_score = min(1.0, risk_score / 5.0)
            
            # Preparar resultado
            result = {
                "risk_score": final_risk_score,
                "risk_level": self._get_risk_level(final_risk_score),
                "anomalies": anomalies,
                "timestamp": datetime.now().isoformat(),
                "user_id": user_history.get("user_id", "unknown")
            }
            
            logger.info(f"Análise de localização concluída: {result['risk_level']} ({final_risk_score:.2f})")
            return result
            
        except Exception as e:
            logger.error(f"Erro na detecção de anomalias de localização: {str(e)}")
            # Retornar avaliação de risco conservadora em caso de erro
            return {
                "risk_score": 0.6,
                "risk_level": "medium",
                "anomalies": [{"type": "erro_analise", "details": str(e), "risk": 0.6}],
                "timestamp": datetime.now().isoformat(),
                "user_id": user_history.get("user_id", "unknown")
            }
            
    def _get_eu_countries(self) -> List[str]:
        """Retorna lista de países membros da União Europeia"""
        return [
            "Austria", "Belgium", "Bulgaria", "Croatia", "Cyprus", "Czech Republic",
            "Denmark", "Estonia", "Finland", "France", "Germany", "Greece", "Hungary",
            "Ireland", "Italy", "Latvia", "Lithuania", "Luxembourg", "Malta", "Netherlands",
            "Poland", "Portugal", "Romania", "Slovakia", "Slovenia", "Spain", "Sweden",
            # Códigos de país
            "AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR", "DE", "GR", "HU",
            "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK", "SI", "ES", "SE"
        ]
        
    def _get_cplp_countries(self) -> List[str]:
        """Retorna lista de países da CPLP (Comunidade dos Países de Língua Portuguesa)"""
        return [
            "Angola", "Brazil", "Cabo Verde", "Guinea-Bissau", "Equatorial Guinea",
            "Mozambique", "Portugal", "São Tomé and Príncipe", "Timor-Leste",
            # Códigos de país
            "AO", "BR", "CV", "GW", "GQ", "MZ", "PT", "ST", "TL"
        ]
        
    def _is_impossible_travel(self, last_location: Dict[str, Any], current_location: Dict[str, Any]) -> bool:
        """Verifica se o deslocamento entre duas localizações é fisicamente impossível"""
        try:
            # Extrair coordenadas
            last_coords = {
                "latitude": last_location.get("coords", {}).get("latitude"),
                "longitude": last_location.get("coords", {}).get("longitude")
            }
            
            current_coords = {
                "latitude": current_location.get("coords", {}).get("latitude"),
                "longitude": current_location.get("coords", {}).get("longitude")
            }
            
            # Verificar se temos coordenadas válidas
            if not all([last_coords["latitude"], last_coords["longitude"], 
                       current_coords["latitude"], current_coords["longitude"]]):
                return False
                
            # Calcular distância
            distance_km = self._calculate_distance(
                last_coords["latitude"], last_coords["longitude"],
                current_coords["latitude"], current_coords["longitude"]
            )
            
            # Calcular tempo decorrido
            last_time = datetime.fromisoformat(last_location.get("timestamp", "2023-01-01T00:00:00"))
            current_time = datetime.fromisoformat(current_location.get("timestamp", datetime.now().isoformat()))
            time_diff_hours = (current_time - last_time).total_seconds() / 3600
            
            # Velocidade em km/h
            if time_diff_hours > 0:
                speed = distance_km / time_diff_hours
                
                # Verificar se a velocidade é impossível (900 km/h é aprox. velocidade de avião comercial)
                # Se a distância é grande e o tempo muito curto
                if speed > 900:
                    return True
                    
                # Para distâncias menores (dentro de Portugal), usar um limite mais baixo
                if distance_km < 1000 and speed > 200:  # Velocidade máxima realista para carro/trem
                    return True
            
            return False
        except Exception as e:
            logger.error(f"Erro ao verificar deslocamento impossível: {str(e)}")
            return False
            
    def _calculate_distance(self, lat1: float, lon1: float, lat2: float, lon2: float) -> float:
        """Calcula a distância em km entre duas coordenadas usando a fórmula de Haversine"""
        from math import sin, cos, sqrt, atan2, radians
        
        # Raio da Terra em km
        R = 6371.0
        
        # Converter graus para radianos
        lat1, lon1, lat2, lon2 = map(radians, [lat1, lon1, lat2, lon2])
        
        # Diferença de longitude e latitude
        dlon = lon2 - lon1
        dlat = lat2 - lat1
        
        # Fórmula de Haversine
        a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
        c = 2 * atan2(sqrt(a), sqrt(1-a))
        distance = R * c
        
        return distance
        
    def _get_ip_location(self, ip: str) -> Dict[str, Any]:
        """Obtém a localização a partir do endereço IP"""
        # Em um ambiente de produção, usaria um serviço de geolocalização de IP
        try:
            # Verificar se temos um adaptador para isso
            if "telecom_portugal" in self.data_adapters:
                return self.data_adapters["telecom_portugal"].get_ip_location(ip)
                
            # Simulação básica para desenvolvimento
            return {
                "latitude": 0.0,
                "longitude": 0.0,
                "city": "unknown",
                "country": "unknown"
            }
        except Exception as e:
            logger.error(f"Erro ao obter localização do IP: {str(e)}")
            return None
            
    def _is_significant_location_mismatch(self, gps_location: Dict[str, float], ip_location: Dict[str, float]) -> bool:
        """Verifica se há discrepância significativa entre localização GPS e IP"""
        try:
            # Calcular distância entre as localizações
            distance = self._calculate_distance(
                gps_location["latitude"], gps_location["longitude"],
                ip_location["latitude"], ip_location["longitude"]
            )
            
            # Considerar discrepância se distância for maior que 50km
            return distance > 50
        except Exception as e:
            logger.error(f"Erro ao verificar discrepância de localização: {str(e)}")
            return False
            
    def _is_location_in_typical_places(self, current_location: Dict[str, Any], typical_locations: List[Dict[str, Any]]) -> bool:
        """Verifica se a localização atual está entre os lugares típicos do usuário"""
        try:
            # Verificar coordenadas
            if not current_location.get("coords", {}).get("latitude") or not current_location.get("coords", {}).get("longitude"):
                # Se não temos coordenadas, verificar cidade/distrito
                for location in typical_locations:
                    if location.get("city") == current_location.get("city") or \
                       location.get("district") == current_location.get("district"):
                        return True
                return False
                
            # Se temos coordenadas, verificar proximidade
            current_coords = {
                "latitude": current_location["coords"]["latitude"],
                "longitude": current_location["coords"]["longitude"]
            }
            
            # Considerar como lugar típico se estiver dentro de um raio de 10km de um lugar conhecido
            for location in typical_locations:
                loc_coords = location.get("coords", {})
                if loc_coords.get("latitude") and loc_coords.get("longitude"):
                    distance = self._calculate_distance(
                        current_coords["latitude"], current_coords["longitude"],
                        loc_coords["latitude"], loc_coords["longitude"]
                    )
                    
                    if distance <= 10:  # 10km de raio
                        return True
            
            return False
        except Exception as e:
            logger.error(f"Erro ao verificar lugares típicos: {str(e)}")
            return True  # Retornar True em caso de erro para evitar falsos positivos
            
    def _is_unusual_time_for_location(self, location: Dict[str, Any]) -> bool:
        """Verifica se o acesso está ocorrendo em um horário atípico para a localização"""
        try:
            # Extrair horário da timestamp
            timestamp = location.get("timestamp", datetime.now().isoformat())
            dt = datetime.fromisoformat(timestamp)
            hour = dt.hour
            
            # Horários típicos por tipo de localização
            typical_hours = {
                "residential": (7, 23),  # 7h às 23h
                "commercial": (8, 20),   # 8h às 20h
                "nightlife": (18, 4)     # 18h às 4h
            }
            
            # Determinar tipo de localização (simplificado)
            location_type = location.get("location_type", "residential")
            
            # Verificar se horário está fora do intervalo típico
            if location_type == "nightlife":
                # Para áreas noturnas, horário atípico é durante o dia
                return 5 <= hour <= 17
            else:
                min_hour, max_hour = typical_hours.get(location_type, (7, 23))
                return hour < min_hour or hour > max_hour
            
        except Exception as e:
            logger.error(f"Erro ao verificar horário atípico: {str(e)}")
            return False
            
    def analyze_device_behavior(self, device_data: Dict[str, Any], 
                              user_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa comportamento de dispositivo com foco em particularidades de Portugal.
        
        Considera:
        - Padrões de uso de dispositivos em Portugal
        - Estatísticas de fraude por tipo de dispositivo no mercado português
        - Uso de VPN e proxies em relação às regulamentações da UE
        - Análise de padrões de navegação típicos portugueses
        
        Args:
            device_data: Dados do dispositivo atual
            user_history: Histórico de dispositivos e comportamento do usuário
            
        Returns:
            Dicionário com pontuação de risco e anomalias detectadas
        """
        try:
            logger.info(f"Analisando comportamento de dispositivo para o usuário: {user_history.get('user_id', 'Desconhecido')}")
            
            # Inicializar resultados
            risk_score = 0.0
            risk_factors = []
            
            # Extrair dados relevantes
            device_id = device_data.get("device_id", "unknown")
            device_type = device_data.get("device_type", "unknown")
            os_name = device_data.get("os", {}).get("name", "unknown")
            os_version = device_data.get("os", {}).get("version", "unknown")
            browser = device_data.get("browser", {}).get("name", "unknown")
            browser_version = device_data.get("browser", {}).get("version", "unknown")
            
            # 1. Verificar se é um dispositivo conhecido
            known_devices = user_history.get("known_devices", [])
            is_known_device = any(d.get("device_id") == device_id for d in known_devices)
            
            if not is_known_device:
                risk_score += 0.6
                risk_factors.append({
                    "factor": "dispositivo_desconhecido",
                    "score": 0.6,
                    "details": f"Tipo: {device_type}, OS: {os_name} {os_version}"
                })
                
                # Se for o primeiro login no dispositivo, verificar quão recente foi o último login
                last_login = user_history.get("last_login", {})
                if last_login and self._is_recent_login(last_login.get("timestamp")):
                    risk_score += 0.3
                    risk_factors.append({
                        "factor": "login_recente_outro_dispositivo",
                        "score": 0.3,
                        "details": f"Login anterior: {last_login.get('timestamp')}"
                    })
            
            # 2. Verificar se o dispositivo está rooteado/jailbroken
            is_rooted = device_data.get("is_rooted", False) or device_data.get("is_jailbroken", False)
            if is_rooted:
                risk_score += 0.7
                risk_factors.append({
                    "factor": "dispositivo_comprometido",
                    "score": 0.7,
                    "details": f"{'Rooted' if device_data.get('is_rooted') else 'Jailbroken'}"
                })
            
            # 3. Verificar sinais de emulador
            is_emulator = device_data.get("is_emulator", False)
            if is_emulator:
                risk_score += 0.8
                risk_factors.append({
                    "factor": "emulador_detectado",
                    "score": 0.8
                })
            
            # 4. Verificar uso de VPN/Proxy/Tor
            is_vpn = device_data.get("is_vpn", False)
            is_proxy = device_data.get("is_proxy", False)
            is_tor = device_data.get("is_tor", False)
            
            if is_vpn or is_proxy or is_tor:
                privacy_tool = "VPN" if is_vpn else ("Proxy" if is_proxy else "Tor")
                risk_score += 0.6
                risk_factors.append({
                    "factor": f"uso_{privacy_tool.lower()}",
                    "score": 0.6,
                    "details": f"{privacy_tool} detectado"
                })
            
            # 5. Verificar inconsistências no navegador
            if self._has_browser_inconsistencies(device_data):
                risk_score += 0.5
                risk_factors.append({
                    "factor": "inconsistencia_navegador",
                    "score": 0.5,
                    "details": f"Browser: {browser} {browser_version}, OS: {os_name}"
                })
            
            # 6. Verificar comportamento de sessão
            session_anomalies = self._detect_session_anomalies(device_data, user_history)
            if session_anomalies:
                for anomaly in session_anomalies:
                    risk_score += anomaly.get("score", 0.0)
                    risk_factors.append({
                        "factor": anomaly.get("type"),
                        "score": anomaly.get("score"),
                        "details": anomaly.get("details", "")
                    })
            
            # 7. Verificar compatibilidade com padrões de Portugal
            if not self._is_device_pattern_typical_for_portugal(device_data):
                risk_score += 0.4
                risk_factors.append({
                    "factor": "padrao_dispositivo_atipico",
                    "score": 0.4,
                    "details": f"Configuração incomum para Portugal"
                })
            
            # 8. Verificar histórico de fraude do dispositivo
            if self._has_device_fraud_history(device_id):
                risk_score += 0.8
                risk_factors.append({
                    "factor": "historico_fraude_dispositivo",
                    "score": 0.8,
                    "details": "Dispositivo associado a fraudes anteriores"
                })
                
            # 9. Verificar múltiplas contas no mesmo dispositivo
            linked_accounts = device_data.get("linked_accounts", [])
            if len(linked_accounts) > 3:  # Mais de 3 contas é suspeito
                risk_score += 0.5
                risk_factors.append({
                    "factor": "multiplas_contas_dispositivo",
                    "score": 0.5,
                    "details": f"{len(linked_accounts)} contas no mesmo dispositivo"
                })
                
            # 10. Verificar fingerprint do dispositivo
            if device_data.get("fingerprint_changed", False):
                risk_score += 0.7
                risk_factors.append({
                    "factor": "fingerprint_alterada",
                    "score": 0.7,
                    "details": "Alteração na fingerprint do dispositivo"
                })
            
            # Normalizar pontuação final
            final_risk_score = min(1.0, risk_score / 6.0)
            
            # Preparar resultado
            result = {
                "risk_score": final_risk_score,
                "risk_level": self._get_risk_level(final_risk_score),
                "risk_factors": risk_factors,
                "timestamp": datetime.now().isoformat(),
                "device_id": device_id,
                "user_id": user_history.get("user_id", "unknown")
            }
            
            logger.info(f"Análise de dispositivo concluída: {result['risk_level']} ({final_risk_score:.2f})")
            return result
            
        except Exception as e:
            logger.error(f"Erro na análise de comportamento do dispositivo: {str(e)}")
            # Retornar avaliação de risco conservadora em caso de erro
            return {
                "risk_score": 0.6,
                "risk_level": "medium",
                "risk_factors": [{"factor": "erro_analise", "score": 0.6, "details": str(e)}],
                "timestamp": datetime.now().isoformat(),
                "device_id": device_data.get("device_id", "unknown"),
                "user_id": user_history.get("user_id", "unknown")
            }
    
    def _is_recent_login(self, timestamp_str: str) -> bool:
        """Verifica se um login é recente (menos de 10 minutos)"""
        if not timestamp_str:
            return False
            
        try:
            timestamp = datetime.fromisoformat(timestamp_str)
            time_diff = (datetime.now() - timestamp).total_seconds() / 60
            return time_diff <= 10  # menos de 10 minutos
        except Exception:
            return False
    
    def _has_browser_inconsistencies(self, device_data: Dict[str, Any]) -> bool:
        """Verifica inconsistências no navegador que podem indicar spoofing"""
        try:
            browser = device_data.get("browser", {})
            os_info = device_data.get("os", {})
            user_agent = device_data.get("user_agent", "")
            
            # Verificar inconsistências conhecidas
            
            # 1. Safari em não-Apple
            if browser.get("name", "").lower() == "safari" and \
               os_info.get("name", "").lower() not in ["ios", "macos"]:
                return True
                
            # 2. Internet Explorer em não-Windows
            if browser.get("name", "").lower() == "internet explorer" and \
               os_info.get("name", "").lower() != "windows":
                return True
                
            # 3. User-Agent não corresponde ao navegador declarado
            if browser.get("name") and user_agent:
                if browser.get("name").lower() not in user_agent.lower():
                    return True
                    
            # 4. Inconsistência entre plataforma e OS
            platform = device_data.get("platform")
            if platform and os_info.get("name"):
                if platform == "mobile" and os_info.get("name").lower() not in ["ios", "android"]:
                    return True
                if platform == "desktop" and os_info.get("name").lower() in ["ios", "android"]:
                    return True
                    
            return False
        except Exception as e:
            logger.error(f"Erro ao verificar inconsistências no navegador: {str(e)}")
            return False
    
    def _detect_session_anomalies(self, device_data: Dict[str, Any], 
                                user_history: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Detecta anomalias na sessão atual comparada com o histórico do usuário"""
        anomalies = []
        try:
            # Dados da sessão atual
            current_session = {
                "login_time": device_data.get("session_start"),
                "ip_address": device_data.get("ip_address"),
                "user_agent": device_data.get("user_agent"),
                "screen_resolution": device_data.get("screen_resolution"),
                "timezone": device_data.get("timezone"),
                "language": device_data.get("language"),
                "referrer": device_data.get("referrer")
            }
            
            # Histórico de sessões
            past_sessions = user_history.get("sessions", [])
            
            # Se não temos histórico suficiente, não podemos detectar anomalias
            if len(past_sessions) < 2:
                return []
                
            # 1. Verificar mudança rápida de timezone
            if current_session.get("timezone") and past_sessions[0].get("timezone"):
                if current_session["timezone"] != past_sessions[0]["timezone"]:
                    # Verificar se o login anterior foi recente (menos de 12 horas)
                    last_login_time = past_sessions[0].get("login_time")
                    if last_login_time and self._is_within_hours(last_login_time, 12):
                        anomalies.append({
                            "type": "mudanca_rapida_timezone",
                            "score": 0.5,
                            "details": f"De {past_sessions[0]['timezone']} para {current_session['timezone']}"
                        })
            
            # 2. Verificar mudança de idioma incomum
            if current_session.get("language") and past_sessions[0].get("language"):
                if current_session["language"] != past_sessions[0]["language"]:
                    # Para Portugal, os idiomas comuns são PT, EN, ES
                    common_langs = ["pt", "pt-pt", "pt-br", "en", "en-gb", "en-us", "es", "es-es"]
                    if current_session["language"].lower() not in common_langs:
                        anomalies.append({
                            "type": "idioma_incomum",
                            "score": 0.4,
                            "details": f"Idioma atual: {current_session['language']}"
                        })
            
            # 3. Verificar comportamento de navegação inconsistente
            typical_referrers = self._extract_common_domains([s.get("referrer", "") for s in past_sessions])
            current_referrer_domain = self._extract_domain(current_session.get("referrer", ""))
            
            if current_referrer_domain and typical_referrers and current_referrer_domain not in typical_referrers:
                anomalies.append({
                    "type": "referrer_atipico",
                    "score": 0.3,
                    "details": f"Referrer: {current_referrer_domain}"
                })
                
            # 4. Verificar resolução de tela inconsistente (pode indicar emulador ou dispositivo diferente)
            if current_session.get("screen_resolution"):
                typical_resolutions = set(s.get("screen_resolution", "") for s in past_sessions if s.get("screen_resolution"))
                if current_session["screen_resolution"] not in typical_resolutions and len(typical_resolutions) >= 2:
                    anomalies.append({
                        "type": "resolucao_tela_inconsistente",
                        "score": 0.3,
                        "details": f"Resolução: {current_session['screen_resolution']}"
                    })
                    
            # 5. Verificar padrões de uso inconsistentes
            # (Simplificado - em produção seria mais complexo)
            if len(past_sessions) >= 5:
                # Verificar padrão de horário de acesso
                typical_hours = [datetime.fromisoformat(s.get("login_time", "2023-01-01T00:00:00")).hour 
                                for s in past_sessions if s.get("login_time")]
                
                current_hour = datetime.fromisoformat(current_session.get("login_time", datetime.now().isoformat())).hour
                
                if typical_hours and current_hour not in typical_hours:
                    anomalies.append({
                        "type": "horario_acesso_atipico",
                        "score": 0.4,
                        "details": f"Hora atual: {current_hour}:00"
                    })
            
            return anomalies
        except Exception as e:
            logger.error(f"Erro ao detectar anomalias de sessão: {str(e)}")
            return []
    
    def _is_within_hours(self, timestamp_str: str, hours: int) -> bool:
        """Verifica se um timestamp está dentro de um número específico de horas"""
        try:
            timestamp = datetime.fromisoformat(timestamp_str)
            time_diff = (datetime.now() - timestamp).total_seconds() / 3600
            return time_diff <= hours
        except Exception:
            return False
    
    def _extract_domain(self, url: str) -> str:
        """Extrai o domínio de uma URL"""
        import re
        if not url:
            return ""
            
        # Expressão regular para extrair domínio
        pattern = r"^(?:https?:\/\/)?(?:[^@\n]+@)?(?:www\.)?([^:\/\n?]+)"
        match = re.match(pattern, url)
        return match.group(1) if match else ""
    
    def _extract_common_domains(self, urls: List[str]) -> List[str]:
        """Extrai domínios comuns de uma lista de URLs"""
        domains = [self._extract_domain(url) for url in urls if url]
        # Filtrar domínios vazios e retornar únicos
        return list(set([d for d in domains if d]))
    
    def _is_device_pattern_typical_for_portugal(self, device_data: Dict[str, Any]) -> bool:
        """Verifica se o padrão do dispositivo é típico para Portugal"""
        try:
            # 1. Verificar sistemas operacionais comuns em Portugal
            os_name = device_data.get("os", {}).get("name", "").lower()
            
            # Em Portugal, Windows, Android e iOS são os mais comuns
            common_os = ["windows", "android", "ios", "macos"]
            if os_name and os_name not in common_os:
                return False
                
            # 2. Verificar navegadores comuns em Portugal
            browser = device_data.get("browser", {}).get("name", "").lower()
            
            # Chrome, Safari, Edge e Firefox são os mais comuns
            common_browsers = ["chrome", "safari", "edge", "firefox", "opera"]
            if browser and browser not in common_browsers:
                return False
                
            # 3. Verificar idioma do navegador
            language = device_data.get("language", "").lower()
            common_langs = ["pt", "pt-pt", "pt-br", "en", "en-gb", "en-us", "es", "es-es"]
            
            if language and not any(language.startswith(lang) for lang in common_langs):
                return False
                
            # 4. Verificar fuso horário
            timezone = device_data.get("timezone", "")
            portugal_timezones = ["europe/lisbon", "wrt", "wet", "utc", "gmt"]
            
            if timezone and not any(tz in timezone.lower() for tz in portugal_timezones):
                return False
                
            return True
        except Exception as e:
            logger.error(f"Erro ao verificar padrão de dispositivo para Portugal: {str(e)}")
            return True  # Em caso de erro, considerar típico para evitar falsos positivos
    
    def _has_device_fraud_history(self, device_id: str) -> bool:
        """Verifica se o dispositivo tem histórico de fraude nos sistemas portugueses"""
        try:
            # Verificar nos adaptadores disponíveis
            if "telecom_portugal" in self.data_adapters:
                return self.data_adapters["telecom_portugal"].check_device_fraud(device_id)
                
            return False
        except Exception as e:
            logger.error(f"Erro ao verificar histórico de fraude do dispositivo: {str(e)}")
            return False
            
    def get_regional_risk_factors(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Obtém fatores de risco específicos para o mercado português.
        
        Consulta fontes de dados específicas de Portugal:
        - Banco de Portugal (restrições bancárias)
        - Autoridade Tributária (situação fiscal)
        - Base de dados Schengen (alertas europeus)
        - Operadoras de telecomunicações portuguesas
        
        Args:
            user_data: Dados do usuário para análise
            
        Returns:
            Dicionário com fatores de risco regionais identificados
        """
        try:
            logger.info(f"Obtendo fatores de risco regionais para o usuário: {user_data.get('user_id', 'Desconhecido')}")
            
            # Inicializar resultados
            risk_score = 0.0
            risk_factors = []
            
            # Extrair dados relevantes
            nif = user_data.get("tax_id")
            cc = user_data.get("id_document", {}).get("number")
            name = user_data.get("name", "")
            
            # 1. Verificar restrições no Banco de Portugal
            if "banco_portugal" in self.data_adapters:
                try:
                    bp_data = self.data_adapters["banco_portugal"].check_user_status(
                        nif=nif,
                        name=name,
                        document_id=cc
                    )
                    
                    # Verificar restrições bancárias
                    if bp_data.get("has_restrictions", False):
                        risk_type = bp_data.get("restriction_type", "desconhecida")
                        risk_score += 0.8
                        risk_factors.append({
                            "factor": "restricao_banco_portugal",
                            "score": 0.8,
                            "details": f"Tipo: {risk_type}"
                        })
                    
                    # Verificar incumprimentos de crédito
                    credit_defaults = bp_data.get("credit_defaults", [])
                    if credit_defaults:
                        risk_score += min(0.7, 0.4 + len(credit_defaults) * 0.1)
                        risk_factors.append({
                            "factor": "incumprimentos_credito",
                            "score": min(0.7, 0.4 + len(credit_defaults) * 0.1),
                            "details": f"{len(credit_defaults)} incumprimentos registrados"
                        })
                        
                    # Verificar cheques devolvidos
                    returned_checks = bp_data.get("returned_checks", 0)
                    if returned_checks > 0:
                        risk_score += min(0.8, 0.5 + returned_checks * 0.1)
                        risk_factors.append({
                            "factor": "cheques_devolvidos",
                            "score": min(0.8, 0.5 + returned_checks * 0.1),
                            "details": f"{returned_checks} cheques devolvidos"
                        })
                except Exception as e:
                    logger.error(f"Erro ao consultar Banco de Portugal: {str(e)}")
            
            # 2. Verificar situação fiscal na Autoridade Tributária
            if "autoridade_tributaria" in self.data_adapters:
                try:
                    at_data = self.data_adapters["autoridade_tributaria"].check_tax_status(
                        nif=nif,
                        name=name
                    )
                    
                    # Verificar dívidas fiscais
                    if at_data.get("has_tax_debt", False):
                        debt_level = at_data.get("debt_level", "unknown")
                        risk_score += 0.6
                        risk_factors.append({
                            "factor": "dividas_fiscais",
                            "score": 0.6,
                            "details": f"Nível: {debt_level}"
                        })
                        
                    # Verificar processos de execução fiscal
                    tax_execution = at_data.get("tax_execution_processes", 0)
                    if tax_execution > 0:
                        risk_score += min(0.7, 0.4 + tax_execution * 0.1)
                        risk_factors.append({
                            "factor": "processos_execucao_fiscal",
                            "score": min(0.7, 0.4 + tax_execution * 0.1),
                            "details": f"{tax_execution} processos em execução fiscal"
                        })
                        
                    # Verificar inconsistências fiscais
                    if at_data.get("has_tax_inconsistencies", False):
                        risk_score += 0.5
                        risk_factors.append({
                            "factor": "inconsistencias_fiscais",
                            "score": 0.5,
                            "details": at_data.get("inconsistency_details", "Não especificado")
                        })
                except Exception as e:
                    logger.error(f"Erro ao consultar Autoridade Tributária: {str(e)}")
            
            # 3. Verificar alertas Schengen/EU
            if "schengen_info" in self.data_adapters:
                try:
                    schengen_data = self.data_adapters["schengen_info"].check_alerts(
                        name=name,
                        document_id=cc,
                        tax_id=nif
                    )
                    
                    # Verificar alertas de segurança
                    if schengen_data.get("has_alerts", False):
                        alert_type = schengen_data.get("alert_type", "desconhecido")
                        risk_score += 0.9
                        risk_factors.append({
                            "factor": "alerta_schengen",
                            "score": 0.9,
                            "details": f"Tipo: {alert_type}"
                        })
                        
                    # Verificar PEP europeu
                    if schengen_data.get("is_eu_pep", False):
                        risk_score += 0.7
                        risk_factors.append({
                            "factor": "pep_europeu",
                            "score": 0.7,
                            "details": schengen_data.get("pep_details", "Não especificado")
                        })
                        
                    # Verificar sanções da UE
                    if schengen_data.get("has_eu_sanctions", False):
                        risk_score += 1.0
                        risk_factors.append({
                            "factor": "sancoes_ue",
                            "score": 1.0,
                            "details": schengen_data.get("sanction_details", "Não especificado")
                        })
                except Exception as e:
                    logger.error(f"Erro ao consultar informações Schengen: {str(e)}")
            
            # 4. Verificar informações de telecomunicações
            if "telecom_portugal" in self.data_adapters:
                try:
                    telecom_data = self.data_adapters["telecom_portugal"].check_user_info(
                        name=name,
                        document_id=cc,
                        tax_id=nif,
                        phone=user_data.get("phone")
                    )
                    
                    # Verificar fraudes de SIM swapping
                    if telecom_data.get("sim_swapping_history", False):
                        risk_score += 0.8
                        risk_factors.append({
                            "factor": "historico_sim_swapping",
                            "score": 0.8,
                            "details": telecom_data.get("sim_swapping_details", "Não especificado")
                        })
                        
                    # Verificar múltiplos SIMs ativos
                    active_sims = telecom_data.get("active_sim_count", 1)
                    if active_sims > 3:  # Mais de 3 SIMs ativos é suspeito
                        risk_score += 0.5
                        risk_factors.append({
                            "factor": "multiplos_sims_ativos",
                            "score": 0.5,
                            "details": f"{active_sims} SIMs ativos"
                        })
                        
                    # Verificar histórico de fraude telecom
                    if telecom_data.get("has_fraud_history", False):
                        risk_score += 0.7
                        risk_factors.append({
                            "factor": "fraude_telecom",
                            "score": 0.7,
                            "details": telecom_data.get("fraud_details", "Não especificado")
                        })
                except Exception as e:
                    logger.error(f"Erro ao consultar informações de telecomunicações: {str(e)}")
            
            # Normalizar pontuação final entre 0 e 1
            final_risk_score = min(1.0, risk_score / 5.0)
            
            # Preparar resultado
            result = {
                "risk_score": final_risk_score,
                "risk_level": self._get_risk_level(final_risk_score),
                "risk_factors": risk_factors,
                "timestamp": datetime.now().isoformat(),
                "user_id": user_data.get("user_id", "unknown"),
                "region": "PT"
            }
            
            logger.info(f"Análise de fatores regionais concluída: {result['risk_level']} ({final_risk_score:.2f})")
            return result
            
        except Exception as e:
            logger.error(f"Erro na obtenção de fatores de risco regionais: {str(e)}")
            # Retornar avaliação de risco conservadora em caso de erro
            return {
                "risk_score": 0.5,
                "risk_level": "medium",
                "risk_factors": [{"factor": "erro_analise_regional", "score": 0.5, "details": str(e)}],
                "timestamp": datetime.now().isoformat(),
                "user_id": user_data.get("user_id", "unknown"),
                "region": "PT"
            }
    
    def _calculate_combined_risk_score(self, account_risk: Dict[str, Any],
                                      location_risk: Dict[str, Any],
                                      device_risk: Dict[str, Any],
                                      regional_risk: Dict[str, Any]) -> Dict[str, Any]:
        """
        Calcula a pontuação de risco combinada baseada em todos os fatores analisados.
        
        Aplica pesos específicos para o contexto português:
        - Maior ênfase em alertas do sistema Schengen e UE
        - Consideração de restrições do Banco de Portugal
        - Valorização de fatores de risco típicos em Portugal
        
        Args:
            account_risk: Resultado da análise de risco da conta
            location_risk: Resultado da análise de localização
            device_risk: Resultado da análise de dispositivo
            regional_risk: Resultado da análise de fatores regionais
            
        Returns:
            Dicionário com pontuação de risco final e todos os fatores considerados
        """
        try:
            # Extrair pontuações de cada análise
            account_score = account_risk.get("risk_score", 0.0)
            location_score = location_risk.get("risk_score", 0.0)
            device_score = device_risk.get("risk_score", 0.0)
            regional_score = regional_risk.get("risk_score", 0.0)
            
            # Configurar pesos para Portugal
            portugal_weights = {
                "account": 0.25,
                "location": 0.20,
                "device": 0.25,
                "regional": 0.30
            }
            
            # Calcular pontuação ponderada
            weighted_score = (
                account_score * portugal_weights["account"] +
                location_score * portugal_weights["location"] +
                device_score * portugal_weights["device"] +
                regional_score * portugal_weights["regional"]
            )
            
            # Regra do "máximo risco": se qualquer uma das pontuações for muito alta,
            # aumentar a pontuação final para refletir o alto risco
            max_risk = max(account_score, location_score, device_score, regional_score)
            if max_risk > 0.8:  # Se alguma pontuação for muito alta (>0.8)
                # Aumentar a pontuação final sem ultrapassar 1.0
                weighted_score = min(1.0, weighted_score * 1.2)
            
            # Consolidar todos os fatores de risco
            all_risk_factors = []
            
            for risk_type, risk_data in [
                ("conta", account_risk),
                ("localização", location_risk),
                ("dispositivo", device_risk),
                ("regional", regional_risk)
            ]:
                factors = risk_data.get("risk_factors", [])
                for factor in factors:
                    factor["source"] = risk_type
                    all_risk_factors.append(factor)
            
            # Ordenar fatores por pontuação (do maior para o menor risco)
            all_risk_factors.sort(key=lambda x: x.get("score", 0), reverse=True)
            
            # Preparar resultado
            result = {
                "risk_score": weighted_score,
                "risk_level": self._get_risk_level(weighted_score),
                "risk_factors": all_risk_factors[:10],  # Top 10 fatores de maior risco
                "scores_by_category": {
                    "account": account_score,
                    "location": location_score,
                    "device": device_score,
                    "regional": regional_score
                },
                "weights": portugal_weights,
                "timestamp": datetime.now().isoformat(),
                "user_id": account_risk.get("account_id") or account_risk.get("user_id", "unknown"),
                "region": "PT"
            }
            
            logger.info(f"Análise combinada concluída: {result['risk_level']} ({weighted_score:.2f})")
            return result
            
        except Exception as e:
            logger.error(f"Erro no cálculo do risco combinado: {str(e)}")
            # Retornar a maior pontuação individual em caso de erro
            max_score = max(
                account_risk.get("risk_score", 0.0),
                location_risk.get("risk_score", 0.0),
                device_risk.get("risk_score", 0.0),
                regional_risk.get("risk_score", 0.0),
                0.5  # Valor padrão moderado
            )
            
            return {
                "risk_score": max_score,
                "risk_level": self._get_risk_level(max_score),
                "risk_factors": [{"factor": "erro_calculo_risco", "score": max_score, "details": str(e)}],
                "timestamp": datetime.now().isoformat(),
                "user_id": account_risk.get("account_id") or account_risk.get("user_id", "unknown"),
                "region": "PT"
            }
    
    def _get_risk_level(self, score: float) -> str:
        """Converte uma pontuação numérica para um nível de risco"""
        if score < 0.3:
            return "low"
        elif score < 0.6:
            return "medium"
        elif score < 0.8:
            return "high"
        else:
            return "critical"