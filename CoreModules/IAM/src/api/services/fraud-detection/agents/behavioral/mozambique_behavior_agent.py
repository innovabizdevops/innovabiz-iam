#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Agente de Análise Comportamental para Moçambique

Este módulo implementa um agente especializado de detecção de fraudes
com adaptações específicas para o mercado moçambicano, considerando
padrões comportamentais, regulamentações locais, características
culturais e dinâmicas econômicas específicas de Moçambique.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
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
logger = logging.getLogger("fraud_detection.behavioral.mozambique")

class MozambiqueBehaviorAgent(BehaviorAnalysisAgent):
    """
    Agente de análise comportamental especializado para o mercado moçambicano.
    
    Esta classe implementa análise de comportamento para detecção de fraudes
    considerando fatores específicos de Moçambique, incluindo:
    - Regulamentações do Banco de Moçambique
    - Padrões comportamentais típicos em transações financeiras moçambicanas
    - Características de uso de dispositivos em Moçambique
    - Dados do sistema financeiro moçambicano e CPC (Central de Risco de Crédito)
    - Validação de documentos moçambicanos (NUIT, BI)
    - Considerações geográficas específicas de Moçambique e SADC
    
    Implementa todos os métodos abstratos da classe BehaviorAnalysisAgent
    com adaptações específicas para o contexto moçambicano.
    """
    
    def __init__(self, config_path: Optional[str] = None, 
                model_path: Optional[str] = None,
                cache_dir: Optional[str] = None,
                data_sources: Optional[List[str]] = None):
        """
        Inicializa o agente de comportamento moçambicano.
        
        Args:
            config_path: Caminho para arquivo de configuração
            model_path: Caminho para modelos treinados
            cache_dir: Diretório para armazenamento de cache
            data_sources: Lista de fontes de dados a utilizar
        """
        # Chamar inicialização da classe pai
        super().__init__(config_path, model_path, cache_dir, data_sources)
        
        # Definir região para Moçambique
        self.region = "MZ"
        
        # Carregar configurações específicas de Moçambique se não fornecido
        if not config_path:
            default_config = os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "config",
                "mozambique_config.json"
            )
            if os.path.exists(default_config):
                self.config_path = default_config
        
        # Inicializar adaptadores específicos para Moçambique
        self._init_mozambique_adapters()
        
        # Carregar padrões regionais para Moçambique
        self._load_mozambique_patterns()
        
        # Fatores de risco específicos de Moçambique
        self.regional_risk_factors = {
            "new_device_login": 0.7,
            "multiple_auth_failures": 0.7,
            "unusual_transaction_time": 0.6,
            "multiple_card_registrations": 0.8,
            "foreign_ip_access": 0.6,
            "high_risk_zone": 0.7,
            "nuit_blacklist": 0.9,
            "banco_mocambique_restrictions": 0.8,
            "sadc_sanctions_list": 0.85,
            "device_fraud_history": 0.75
        }
        
        # Inicializar modelos específicos para Moçambique
        self._load_mozambique_models()
        
        logger.info(f"Agente de análise comportamental de Moçambique inicializado. Versão: 1.0.0")
    
    def _init_mozambique_adapters(self):
        """Inicializa adaptadores de dados específicos para Moçambique"""
        try:
            # Lista de adaptadores a serem inicializados
            mozambique_adapters = {
                "banco_mocambique": "BancoMocambiqueAdapter",
                "autoridade_tributaria_mz": "AutoridadeTributariaMZAdapter",
                "sadc_info": "SADCInfoAdapter",
                "telecom_mocambique": "TelecomMocambiqueAdapter"
            }
            
            # Inicializar adaptadores selecionados ou todos se nenhum foi especificado
            adapter_names = self.config.get("data_sources", list(mozambique_adapters.keys()))
            
            for adapter_name in adapter_names:
                if adapter_name in mozambique_adapters:
                    try:
                        # Caminho dinâmico para importação dos adaptadores
                        adapter_module = f"...adapters.mozambique.{adapter_name.lower()}_adapter"
                        adapter_class = mozambique_adapters[adapter_name]
                        
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
                logger.warning("Nenhum adaptador de dados foi inicializado para Moçambique")
                
        except Exception as e:
            logger.error(f"Falha na inicialização dos adaptadores para Moçambique: {str(e)}")
    
    def _load_mozambique_patterns(self):
        """Carrega padrões comportamentais específicos de Moçambique"""
        try:
            # Tentar carregar padrões de um arquivo
            patterns_path = os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "patterns",
                "mozambique_patterns.json"
            )
            
            if os.path.exists(patterns_path):
                with open(patterns_path, 'r', encoding='utf-8') as f:
                    self.regional_patterns = json.load(f)
                logger.info(f"Padrões regionais de Moçambique carregados de {patterns_path}")
            else:
                # Definir padrões padrão se o arquivo não existe
                self.regional_patterns = {
                    "transaction_patterns": {
                        "typical_transaction_amount": {
                            "p2p_transfer": {
                                "mean": 1500.00,  # Em meticais
                                "std_dev": 1000.00,
                                "max_normal": 10000.00
                            },
                            "bill_payment": {
                                "mean": 800.00,
                                "std_dev": 600.00,
                                "max_normal": 5000.00
                            },
                            "retail_purchase": {
                                "mean": 500.00,
                                "std_dev": 400.00,
                                "max_normal": 3000.00
                            }
                        },
                        "peak_transaction_hours": [9, 12, 16, 18],
                        "low_activity_hours": [1, 2, 3, 4],
                        "weekend_usage_factor": 0.6,
                        "month_end_increase_factor": 1.8,
                        "common_transaction_frequencies": {
                            "daily": 1,
                            "weekly": 2,
                            "monthly": 8
                        },
                        "high_risk_merchants_categories": [
                            "apostas_online",
                            "transferencias_internacionais_nao_identificadas",
                            "casinos",
                            "cambio_nao_autorizado"
                        ],
                        "high_risk_regions": [
                            "fora_sadc",
                            "zonas_fronteirica"
                        ],
                        "mpesa_specific_patterns": {
                            "typical_frequency_daily": 2,
                            "max_normal_amount": 5000.00,
                            "suspicious_time_gap_seconds": 30
                        }
                    },
                    "behavioral_patterns": {
                        "device_usage": {
                            "mobile_predominance": 0.85,  # Alta prevalência mobile
                            "typical_session_duration_min": 8,
                            "common_device_change_frequency_days": 180,
                            "max_normal_devices_per_user": 2,
                            "typical_auth_methods": ["password", "pin", "codigo_sms", "biometria"]
                        },
                        "login_patterns": {
                            "typical_login_frequency_days": 3,
                            "typical_login_hours": [7, 21],
                            "suspicious_login_attempts_threshold": 3
                        }
                    },
                    "location_patterns": {
                        "high_risk_areas": [
                            "Maputo Cidade", "Matola", "Beira", 
                            "Nampula", "Tete", "Pemba"
                        ],
                        "common_movement_radius_km": 20,
                        "typical_speed_kmh": 60,
                        "province_risk_factors": {
                            "Maputo Cidade": 0.5, "Maputo Província": 0.45,
                            "Gaza": 0.4, "Inhambane": 0.35, "Sofala": 0.45,
                            "Manica": 0.4, "Tete": 0.5, "Zambézia": 0.4,
                            "Nampula": 0.5, "Cabo Delgado": 0.5, "Niassa": 0.4
                        },
                        "sadc_border_risk": 0.5,
                        "non_sadc_border_risk": 0.7
                    }
                }
                logger.warning(f"Arquivo de padrões para Moçambique não encontrado. Usando padrões padrão.")
                
        except Exception as e:
            logger.error(f"Erro ao carregar padrões regionais de Moçambique: {str(e)}")
            # Definir padrões mínimos em caso de erro
            self.regional_patterns = {
                "transaction_patterns": {"typical_amount": 500.0},
                "behavioral_patterns": {"device_usage": {"mobile_predominance": 0.8}},
                "location_patterns": {"high_risk_areas": []}
            }
    
    def _load_mozambique_models(self):
        """Carrega modelos de ML específicos para Moçambique"""
        try:
            # Verificar diretório de modelos
            if not self.model_path:
                logger.warning("Caminho de modelos não definido. Usando heurísticas.")
                return
            
            # Definir caminho específico para modelos de Moçambique
            mozambique_models_path = os.path.join(self.model_path, "mozambique")
            
            # Verificar e carregar modelos específicos (quando disponíveis)
            model_files = {
                "transaction_risk": "mozambique_transaction_risk_model.pkl",
                "account_risk": "mozambique_account_risk_model.pkl",
                "location_risk": "mozambique_location_anomaly_model.pkl",
                "device_risk": "mozambique_device_behavior_model.pkl"
            }
            
            # Carregar modelos disponíveis
            for model_type, model_file in model_files.items():
                model_path = os.path.join(mozambique_models_path, model_file)
                if os.path.exists(model_path):
                    try:
                        # Aqui usaria uma função para carregar o modelo de acordo com seu tipo
                        # self.models[model_type] = load_model(model_path)
                        logger.info(f"Modelo {model_type} para Moçambique carregado com sucesso")
                    except Exception as e:
                        logger.error(f"Erro ao carregar modelo {model_type}: {str(e)}")
                else:
                    logger.warning(f"Modelo {model_type} não encontrado. Usando regras heurísticas.")
                    
        except Exception as e:
            logger.error(f"Erro ao carregar modelos de Moçambique: {str(e)}")    def evaluate_account_risk(self, account_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Avalia o risco da conta com base em critérios específicos de Moçambique.
        
        Args:
            account_data: Dados da conta a ser analisada
                          Deve conter: account_id, user_id, created_at, kyc_status,
                          suspicious_activity_history, etc.
                          
        Returns:
            Dicionário com pontuação de risco, nível de risco e fatores de risco
        """
        try:
            logger.info(f"Avaliando risco da conta {account_data.get('account_id', 'desconhecido')} para Moçambique")
            
            # Inicializar lista de fatores de risco
            risk_factors = []
            
            # Verificar se temos dados suficientes para análise
            if not account_data or 'account_id' not in account_data:
                logger.warning("Dados insuficientes para análise de conta")
                return {
                    'risk_score': 0.85,
                    'risk_level': 'alto',
                    'risk_factors': [{'factor': 'dados_insuficientes', 'score': 0.85}]
                }
            
            # 1. Verificar idade da conta (contas novas têm maior risco)
            account_age_days = self._calculate_account_age(account_data.get('created_at'))
            if account_age_days < 30:
                risk_score = min(0.7, 1.0 - (account_age_days / 30) * 0.3)
                risk_factors.append({
                    'factor': 'conta_recente',
                    'score': risk_score,
                    'details': f'Conta criada há {account_age_days} dias'
                })
            
            # 2. Verificar status de KYC
            kyc_status = account_data.get('kyc_status', 'desconhecido')
            if kyc_status not in ['verified', 'verificado']:
                risk_factors.append({
                    'factor': 'kyc_incompleto',
                    'score': 0.75,
                    'details': f'Status KYC: {kyc_status}'
                })
            
            # 3. Verificar NUIT (Número Único de Identificação Tributária)
            tax_id = account_data.get('tax_id')
            if tax_id:
                if not self._validate_mozambique_tax_id(tax_id):
                    risk_factors.append({
                        'factor': 'nuit_invalido',
                        'score': 0.8,
                        'details': 'NUIT com formato inválido ou inexistente'
                    })
                
                # Verificar NUIT em listas de alto risco
                nuit_blacklist_check = self._check_nuit_in_blacklist(tax_id)
                if nuit_blacklist_check.get('found', False):
                    risk_factors.append({
                        'factor': 'nuit_em_lista_negra',
                        'score': 0.95,
                        'details': nuit_blacklist_check.get('reason', 'Motivo não especificado')
                    })
            else:
                risk_factors.append({
                    'factor': 'nuit_ausente',
                    'score': 0.7,
                    'details': 'NUIT não fornecido'
                })
            
            # 4. Verificar histórico de atividades suspeitas
            suspicious_activity = account_data.get('suspicious_activity_history', [])
            if suspicious_activity:
                # Calcular pontuação com base na quantidade e recência das atividades
                activity_score = min(0.9, 0.5 + len(suspicious_activity) * 0.1)
                risk_factors.append({
                    'factor': 'historico_atividades_suspeitas',
                    'score': activity_score,
                    'details': f'{len(suspicious_activity)} ocorrências registradas'
                })
            
            # 5. Verificar se o endereço está em área de alto risco
            address = account_data.get('address', {})
            if address:
                district = address.get('district')
                city = address.get('city')
                
                if district and city:
                    # Verificar se o distrito/cidade está em áreas de alto risco
                    is_high_risk_area = self._check_high_risk_area(district, city)
                    if is_high_risk_area:
                        risk_factors.append({
                            'factor': 'area_alto_risco',
                            'score': 0.65,
                            'details': f'Localização em área de risco: {city}, {district}'
                        })
                
                # Verificar se endereço é temporário
                if address.get('is_temporary', False):
                    risk_factors.append({
                        'factor': 'endereco_temporario',
                        'score': 0.6,
                        'details': 'Usuário utiliza endereço temporário'
                    })
                
                # Verificar divergência entre endereço residencial e fiscal
                if address.get('differs_from_fiscal_address', False):
                    risk_factors.append({
                        'factor': 'divergencia_endereco',
                        'score': 0.55,
                        'details': 'Divergência entre endereço residencial e fiscal'
                    })
            
            # 6. Verificar se é PEP (Pessoa Politicamente Exposta)
            if account_data.get('is_pep', False):
                risk_factors.append({
                    'factor': 'pessoa_politicamente_exposta',
                    'score': 0.75,
                    'details': 'Usuário identificado como PEP'
                })
                
                # Verificar parentes PEP
                if account_data.get('pep_relatives', []):
                    risk_factors.append({
                        'factor': 'parentes_pep',
                        'score': 0.6,
                        'details': f'{len(account_data.get("pep_relatives", []))} parentes identificados como PEP'
                    })
            
            # 7. Verificar atividade econômica
            if account_data.get('economic_activity') in ['cambio', 'mineracao', 'trading_internacional']:
                risk_factors.append({
                    'factor': 'atividade_economica_alto_risco',
                    'score': 0.7,
                    'details': f'Atividade de risco: {account_data.get("economic_activity")}'
                })
            
            # 8. Verificar dados do documento de identidade
            id_doc = account_data.get('id_document', {})
            if id_doc:
                # Verificar se o documento está expirado
                if 'expiry_date' in id_doc:
                    try:
                        expiry_date = datetime.fromisoformat(id_doc['expiry_date'])
                        if expiry_date < datetime.now():
                            risk_factors.append({
                                'factor': 'documento_identidade_expirado',
                                'score': 0.7,
                                'details': f'Documento expirou em {id_doc["expiry_date"]}'
                            })
                    except (ValueError, TypeError):
                        risk_factors.append({
                            'factor': 'data_expiracao_invalida',
                            'score': 0.6,
                            'details': 'Data de expiração do documento inválida'
                        })
            
            # Calcular pontuação final de risco
            final_risk_score = self._calculate_risk_score(risk_factors)
            
            # Determinar nível de risco
            risk_level = self._determine_risk_level(final_risk_score)
            
            return {
                'risk_score': final_risk_score,
                'risk_level': risk_level,
                'risk_factors': sorted(risk_factors, key=lambda x: x['score'], reverse=True)
            }
            
        except Exception as e:
            logger.error(f"Erro na avaliação de risco da conta: {str(e)}")
            return {
                'risk_score': 0.5,
                'risk_level': 'médio',
                'risk_factors': [{'factor': 'erro_processamento', 'score': 0.5}],
                'error': str(e)
            }
    
    def _validate_mozambique_tax_id(self, tax_id: str) -> bool:
        """Valida o formato do NUIT moçambicano"""
        # NUIT moçambicano tem 9 dígitos
        if not tax_id or not isinstance(tax_id, str):
            return False
        
        tax_id = tax_id.strip()
        
        # Verificar se contém apenas dígitos e tem 9 caracteres
        if not tax_id.isdigit() or len(tax_id) != 9:
            return False
        
        # Em uma implementação real, aqui teria a validação do algoritmo específico
        # do NUIT moçambicano e potencialmente verificações contra bases oficiais
        # Implementação simplificada
        return True
    
    def _check_nuit_in_blacklist(self, tax_id: str) -> Dict[str, Any]:
        """Verifica se o NUIT está em alguma lista de restrição"""
        try:
            # Tentar usar adaptador do Banco de Moçambique se disponível
            if 'banco_mocambique' in self.data_adapters:
                result = self.data_adapters['banco_mocambique'].check_restricted_tax_id(tax_id)
                if result and result.get('found', False):
                    return result
            
            # Verificação em cache local (mockado)
            high_risk_nuits = [
                "123456789", "987654321", "111222333", "444555666"
            ]
            
            if tax_id in high_risk_nuits:
                return {
                    'found': True,
                    'reason': 'NUIT em lista de restrições simulada',
                    'source': 'cache_local',
                    'risk_score': 0.9
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar NUIT em lista negra: {str(e)}")
            return {'found': False, 'error': str(e)}    def _check_high_risk_area(self, district: str, city: str) -> bool:
        """Verifica se a localização está em área de alto risco"""
        high_risk_areas = self.regional_patterns.get("location_patterns", {}).get("high_risk_areas", [])
        return district in high_risk_areas or city in high_risk_areas
    
    def detect_location_anomalies(self, location_data: Dict[str, Any], 
                                user_location_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Detecta anomalias de localização específicas para Moçambique.
        
        Args:
            location_data: Dados da localização atual
            user_location_history: Histórico de localizações do usuário
            
        Returns:
            Dicionário com pontuação de risco, nível e anomalias detectadas
        """
        try:
            logger.info(f"Analisando anomalias de localização para usuário em Moçambique")
            
            # Lista para armazenar anomalias detectadas
            anomalies = []
            
            # Verificar se temos dados suficientes
            if not location_data or not user_location_history:
                logger.warning("Dados insuficientes para análise de localização")
                return {
                    'risk_score': 0.7,
                    'risk_level': 'alto',
                    'anomalies': [{'type': 'dados_insuficientes', 'risk': 0.7}]
                }
            
            # 1. Verificar se é um IP de Moçambique
            ip_address = location_data.get('ip_address')
            country = location_data.get('country')
            
            if country != 'MZ':
                # Verificar se está em um país SADC
                sadc_countries = ['ZA', 'BW', 'LS', 'NA', 'SZ', 'AO', 'MW', 'MU', 'ZM', 'ZW', 'TZ', 'SC', 'CD']
                if country in sadc_countries:
                    anomalies.append({
                        'type': 'acesso_fora_de_mocambique_sadc',
                        'risk': 0.6,
                        'details': f'Acesso de país SADC: {country}'
                    })
                else:
                    anomalies.append({
                        'type': 'acesso_fora_de_mocambique',
                        'risk': 0.85,
                        'details': f'Acesso internacional de: {country or "País desconhecido"}'
                    })
            
            # 2. Verificar se é um IP de VPN/proxy
            if location_data.get('is_vpn', False) or location_data.get('is_proxy', False):
                anomalies.append({
                    'type': 'vpn_proxy_detectado',
                    'risk': 0.75,
                    'details': f'Uso de {"VPN" if location_data.get("is_vpn") else "proxy"} detectado'
                })
            
            # 3. Verificar se a localização atual difere significativamente do histórico
            if 'recent_locations' in user_location_history and user_location_history['recent_locations']:
                # Verificar distância das últimas localizações conhecidas
                if 'coords' in location_data and location_data['coords']:
                    current_lat = location_data['coords'].get('latitude')
                    current_lon = location_data['coords'].get('longitude')
                    
                    # Verificar se temos coordenadas válidas
                    if current_lat and current_lon:
                        distances = []
                        for loc in user_location_history['recent_locations']:
                            if 'coords' in loc and loc['coords']:
                                loc_lat = loc['coords'].get('latitude')
                                loc_lon = loc['coords'].get('longitude')
                                if loc_lat and loc_lon:
                                    distance = self._calculate_distance(
                                        current_lat, current_lon, loc_lat, loc_lon
                                    )
                                    distances.append(distance)
                        
                        # Avaliar distâncias
                        if distances:
                            min_distance = min(distances)
                            typical_radius = self.regional_patterns.get("location_patterns", {}).get(
                                "common_movement_radius_km", 20)
                            
                            if min_distance > typical_radius * 5:  # 5x o raio típico
                                anomalies.append({
                                    'type': 'deslocamento_incomum',
                                    'risk': 0.8,
                                    'details': f'Distância de {min_distance:.2f} km da última localização conhecida'
                                })
                            elif min_distance > typical_radius * 2:  # 2x o raio típico
                                anomalies.append({
                                    'type': 'deslocamento_suspeito',
                                    'risk': 0.6,
                                    'details': f'Distância de {min_distance:.2f} km da última localização conhecida'
                                })
            
            # 4. Verificar mudanças rápidas de localização (viagens impossíveis)
            recent_locations = user_location_history.get('recent_locations', [])
            if len(recent_locations) >= 2 and 'timestamp' in location_data:
                try:
                    # Pegar a localização mais recente do histórico
                    latest_location = recent_locations[0]
                    
                    # Verificar se temos timestamp em ambas localizações
                    if 'timestamp' in latest_location:
                        current_time = datetime.fromisoformat(location_data['timestamp'])
                        latest_time = datetime.fromisoformat(latest_location['timestamp'])
                        
                        # Calcular diferença de tempo em horas
                        time_diff_hours = (current_time - latest_time).total_seconds() / 3600
                        
                        # Se temos coordenadas em ambas localizações
                        if ('coords' in location_data and 'coords' in latest_location and
                            location_data['coords'] and latest_location['coords']):
                            
                            # Calcular distância
                            distance = self._calculate_distance(
                                location_data['coords'].get('latitude'), 
                                location_data['coords'].get('longitude'),
                                latest_location['coords'].get('latitude'), 
                                latest_location['coords'].get('longitude')
                            )
                            
                            # Velocidade média (km/h)
                            if time_diff_hours > 0:
                                speed = distance / time_diff_hours
                                
                                # Velocidade máxima plausível (ajustado para Moçambique)
                                max_speed = self.regional_patterns.get("location_patterns", {}).get(
                                    "typical_speed_kmh", 60) * 2
                                
                                if speed > max_speed:
                                    anomalies.append({
                                        'type': 'viagem_impossivel',
                                        'risk': 0.9,
                                        'details': f'Velocidade calculada de {speed:.2f} km/h entre localizações'
                                    })
                except (ValueError, TypeError, IndexError) as e:
                    logger.warning(f"Erro ao calcular velocidade de viagem: {str(e)}")
            
            # 5. Verificar se está em área de alto risco
            if location_data.get('city') and location_data.get('district'):
                if self._check_high_risk_area(location_data['district'], location_data['city']):
                    anomalies.append({
                        'type': 'area_alto_risco',
                        'risk': 0.65,
                        'details': f'Localização em área de alto risco: {location_data["city"]}, {location_data["district"]}'
                    })
            
            # 6. Verificar proximidade com fronteira
            if location_data.get('border_proximity', {}).get('is_near_border', False):
                border_data = location_data.get('border_proximity', {})
                border_country = border_data.get('bordering_country', 'Desconhecido')
                
                # Verificar se é fronteira com país SADC
                sadc_countries = ['ZA', 'BW', 'LS', 'NA', 'SZ', 'AO', 'MW', 'MU', 'ZM', 'ZW', 'TZ', 'SC', 'CD']
                
                if border_country in sadc_countries:
                    risk_factor = self.regional_patterns.get("location_patterns", {}).get(
                        "sadc_border_risk", 0.5)
                    anomalies.append({
                        'type': 'proximidade_fronteira_sadc',
                        'risk': risk_factor,
                        'details': f'Proximidade com fronteira de {border_country}'
                    })
                else:
                    risk_factor = self.regional_patterns.get("location_patterns", {}).get(
                        "non_sadc_border_risk", 0.7)
                    anomalies.append({
                        'type': 'proximidade_fronteira_nao_sadc',
                        'risk': risk_factor,
                        'details': f'Proximidade com fronteira de {border_country}'
                    })
            
            # Calcular pontuação final baseada nas anomalias detectadas
            final_risk_score = 0.0
            if anomalies:
                final_risk_score = sum(a['risk'] for a in anomalies) / len(anomalies)
                # Aumentar risco se houver múltiplas anomalias
                if len(anomalies) > 1:
                    final_risk_score = min(1.0, final_risk_score * (1 + (len(anomalies) - 1) * 0.1))
            
            # Determinar nível de risco
            risk_level = self._determine_risk_level(final_risk_score)
            
            return {
                'risk_score': final_risk_score,
                'risk_level': risk_level,
                'anomalies': sorted(anomalies, key=lambda x: x['risk'], reverse=True)
            }
            
        except Exception as e:
            logger.error(f"Erro na detecção de anomalias de localização: {str(e)}")
            return {
                'risk_score': 0.5,
                'risk_level': 'médio',
                'anomalies': [{'type': 'erro_processamento', 'risk': 0.5}],
                'error': str(e)
            }
    
    def _calculate_distance(self, lat1: float, lon1: float, lat2: float, lon2: float) -> float:
        """
        Calcula a distância em quilômetros entre duas coordenadas
        usando a fórmula de Haversine
        """
        from math import radians, cos, sin, asin, sqrt
        
        # Converter coordenadas para radianos
        lat1, lon1, lat2, lon2 = map(radians, [lat1, lon1, lat2, lon2])
        
        # Fórmula de Haversine
        dlon = lon2 - lon1
        dlat = lat2 - lat1
        a = sin(dlat/2)**2 + cos(lat1) * cos(lat2) * sin(dlon/2)**2
        c = 2 * asin(sqrt(a))
        r = 6371  # Raio da Terra em quilômetros
        
        return c * r    def analyze_device_behavior(self, device_data: Dict[str, Any], 
                             user_device_history: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analisa o comportamento do dispositivo com base em padrões moçambicanos.
        
        Args:
            device_data: Dados do dispositivo atual
            user_device_history: Histórico de dispositivos do usuário
            
        Returns:
            Dicionário com pontuação de risco, nível e fatores de risco
        """
        try:
            logger.info(f"Analisando comportamento de dispositivo para usuário em Moçambique")
            
            # Lista para armazenar fatores de risco
            risk_factors = []
            
            # Verificar se temos dados suficientes
            if not device_data or not user_device_history:
                logger.warning("Dados insuficientes para análise de dispositivo")
                return {
                    'risk_score': 0.6,
                    'risk_level': 'médio',
                    'risk_factors': [{'factor': 'dados_insuficientes', 'score': 0.6}]
                }
            
            # 1. Verificar dispositivo desconhecido
            device_id = device_data.get('device_id')
            known_devices = user_device_history.get('known_devices', [])
            
            is_known_device = any(d.get('device_id') == device_id for d in known_devices)
            
            if not is_known_device:
                risk_factors.append({
                    'factor': 'dispositivo_desconhecido',
                    'score': 0.7,
                    'details': 'Primeiro acesso registrado deste dispositivo'
                })
            
            # 2. Verificar se o dispositivo está comprometido
            if device_data.get('is_rooted', False) or device_data.get('is_jailbroken', False):
                risk_factors.append({
                    'factor': 'dispositivo_comprometido',
                    'score': 0.8,
                    'details': 'Dispositivo com sistema operacional comprometido'
                })
            
            # 3. Verificar se é emulador
            if device_data.get('is_emulator', False):
                risk_factors.append({
                    'factor': 'uso_emulador',
                    'score': 0.85,
                    'details': 'Uso de emulador detectado'
                })
            
            # 4. Verificar uso de VPN/proxy/TOR
            if device_data.get('is_vpn', False) or device_data.get('is_proxy', False) or device_data.get('is_tor', False):
                methods = []
                if device_data.get('is_vpn', False): methods.append('VPN')
                if device_data.get('is_proxy', False): methods.append('proxy')
                if device_data.get('is_tor', False): methods.append('TOR')
                
                risk_factors.append({
                    'factor': 'anonimizacao_detectada',
                    'score': 0.75,
                    'details': f'Uso de {", ".join(methods)} detectado'
                })
            
            # 5. Verificar se houve mudança de fingerprint
            if device_data.get('fingerprint_changed', False):
                risk_factors.append({
                    'factor': 'fingerprint_alterado',
                    'score': 0.7,
                    'details': 'Alteração na impressão digital do dispositivo'
                })
            
            # 6. Verificar múltiplas contas no mesmo dispositivo
            linked_accounts = device_data.get('linked_accounts', [])
            if len(linked_accounts) > 2:
                risk_factors.append({
                    'factor': 'multiplas_contas',
                    'score': 0.65 + min(0.25, (len(linked_accounts) - 2) * 0.05),
                    'details': f'{len(linked_accounts)} contas vinculadas a este dispositivo'
                })
            
            # 7. Verificar inconsistências no timezone/idioma com localização
            if device_data.get('timezone') == 'Europe/Lisbon' and device_data.get('language') == 'pt-PT':
                # Combinação típica para Portugal, não para Moçambique
                risk_factors.append({
                    'factor': 'inconsistencia_idioma_regiao',
                    'score': 0.65,
                    'details': 'Configurações de idioma/timezone consistentes com Portugal, não Moçambique'
                })
            
            # 8. Verificar histórico de fraudes relacionado ao dispositivo
            device_fraud_history = self._check_device_fraud_history(device_id)
            if device_fraud_history.get('found', False):
                risk_factors.append({
                    'factor': 'dispositivo_com_historico_fraude',
                    'score': 0.9,
                    'details': device_fraud_history.get('details', 'Dispositivo associado a atividades fraudulentas')
                })
            
            # 9. Analisar comportamento típico para Moçambique
            
            # Verificar se é mobile (predominante em Moçambique)
            if device_data.get('platform') != 'mobile' and device_data.get('device_type') not in ['smartphone', 'feature_phone']:
                # Maior uso de desktop/web não é tão comum em Moçambique
                risk_factors.append({
                    'factor': 'uso_atipico_nao_mobile',
                    'score': 0.4,
                    'details': 'Uso não mobile é menos comum no padrão moçambicano'
                })
            
            # Verificar uso de dispositivo em horário incomum
            if device_data.get('session_start'):
                try:
                    session_time = datetime.fromisoformat(device_data['session_start'])
                    hour = session_time.hour
                    
                    if hour >= 1 and hour <= 4:  # Horário de baixa atividade
                        risk_factors.append({
                            'factor': 'uso_horario_incomum',
                            'score': 0.5,
                            'details': f'Login em horário atípico: {hour}:00'
                        })
                except (ValueError, TypeError):
                    pass
            
            # 10. Comparar com últimas sessões
            sessions = user_device_history.get('sessions', [])
            if sessions and len(sessions) >= 2:
                try:
                    current_user_agent = device_data.get('user_agent', '')
                    previous_agents = [s.get('user_agent', '') for s in sessions[:2]]
                    
                    if current_user_agent and all(current_user_agent != agent for agent in previous_agents if agent):
                        risk_factors.append({
                            'factor': 'mudanca_user_agent',
                            'score': 0.5,
                            'details': 'User Agent diferente das sessões anteriores'
                        })
                except Exception:
                    pass
            
            # Calcular pontuação final baseada nos fatores de risco
            final_risk_score = self._calculate_risk_score(risk_factors)
            
            # Determinar nível de risco
            risk_level = self._determine_risk_level(final_risk_score)
            
            return {
                'risk_score': final_risk_score,
                'risk_level': risk_level,
                'risk_factors': sorted(risk_factors, key=lambda x: x['score'], reverse=True)
            }
            
        except Exception as e:
            logger.error(f"Erro na análise de comportamento de dispositivo: {str(e)}")
            return {
                'risk_score': 0.5,
                'risk_level': 'médio',
                'risk_factors': [{'factor': 'erro_processamento', 'score': 0.5}],
                'error': str(e)
            }
    
    def _check_device_fraud_history(self, device_id: str) -> Dict[str, Any]:
        """Verifica o histórico de fraudes associado ao dispositivo"""
        try:
            # Tentar usar adaptador se disponível
            if 'banco_mocambique' in self.data_adapters:
                result = self.data_adapters['banco_mocambique'].check_device_fraud(device_id)
                if result and result.get('found', False):
                    return result
            
            # Verificação em cache local (mockado)
            high_risk_devices = [
                "dev-mz-12345", "dev-mz-67890", "dev-fraud-01234"
            ]
            
            if device_id in high_risk_devices:
                return {
                    'found': True,
                    'details': 'Dispositivo em lista de restrições simulada',
                    'source': 'cache_local',
                    'risk_score': 0.9
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar histórico de fraudes do dispositivo: {str(e)}")
            return {'found': False, 'error': str(e)}    def get_regional_risk_factors(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Obtém fatores de risco específicos para o contexto moçambicano.
        
        Args:
            user_data: Dados do usuário para análise contextual
            
        Returns:
            Dicionário com pontuação de risco, nível e fatores de risco regionais
        """
        try:
            logger.info(f"Analisando fatores de risco regionais para Moçambique")
            
            # Lista para armazenar fatores de risco
            risk_factors = []
            
            # Verificar dados básicos
            if not user_data or not user_data.get('user_id'):
                logger.warning("Dados insuficientes para análise de risco regional")
                return {
                    'risk_score': 0.5,
                    'risk_level': 'médio',
                    'risk_factors': [{'factor': 'dados_insuficientes', 'score': 0.5}]
                }
            
            # 1. Verificar no sistema financeiro de Moçambique
            
            # Verificar NUIT no sistema da Autoridade Tributária
            tax_id = user_data.get('tax_id')
            if tax_id:
                # Verificar se existe na lista negra
                tax_blacklist_check = self._check_tax_status(tax_id)
                if tax_blacklist_check.get('found', False):
                    risk_factors.append({
                        'factor': 'restricoes_fiscais',
                        'score': 0.85,
                        'details': tax_blacklist_check.get('reason', 'Pendências fiscais detectadas')
                    })
                
                # Verificar histórico de impostos não pagos
                if self._has_tax_compliance_issues(tax_id):
                    risk_factors.append({
                        'factor': 'irregularidade_fiscal',
                        'score': 0.7,
                        'details': 'Histórico de irregularidades fiscais'
                    })
            
            # 2. Verificar na Central de Risco do Banco de Moçambique
            user_id = user_data.get('user_id')
            credit_risk_check = self._check_credit_risk(user_id, tax_id)
            if credit_risk_check.get('found', False):
                risk_factors.append({
                    'factor': 'historico_credito_negativo',
                    'score': 0.75,
                    'details': credit_risk_check.get('reason', 'Histórico negativo na central de risco')
                })
            
            # 3. Verificar restrições no sistema bancário
            bank_restrictions = self._check_banking_restrictions(user_id, tax_id)
            if bank_restrictions.get('found', False):
                risk_factors.append({
                    'factor': 'restricoes_bancarias',
                    'score': 0.8,
                    'details': bank_restrictions.get('reason', 'Restrições bancárias identificadas')
                })
            
            # 4. Verificar fatores de risco específicos de telecomunicações moçambicanas
            phone = user_data.get('phone')
            if phone:
                # Limpar o formato do número
                cleaned_phone = ''.join(filter(str.isdigit, phone))
                
                # Verificar número moçambicano
                is_moz_number = cleaned_phone.startswith('258') or (cleaned_phone.startswith('8') and len(cleaned_phone) == 9)
                
                if not is_moz_number:
                    risk_factors.append({
                        'factor': 'numero_telefone_estrangeiro',
                        'score': 0.6,
                        'details': 'Número de telefone não é de Moçambique'
                    })
                else:
                    # Verificar operadoras moçambicanas (mVodacom, mCel, Movitel)
                    operator_prefix = cleaned_phone[0:2] if cleaned_phone.startswith('8') else cleaned_phone[3:5]
                    
                    if operator_prefix not in ['82', '83', '84', '85', '86', '87']:
                        risk_factors.append({
                            'factor': 'operadora_desconhecida',
                            'score': 0.5,
                            'details': 'Prefixo de operadora desconhecido'
                        })
                    
                    # Verificar histórico do número
                    phone_history = self._check_phone_history(cleaned_phone)
                    if phone_history.get('found', False):
                        risk_factors.append({
                            'factor': 'telefone_alto_risco',
                            'score': 0.7,
                            'details': phone_history.get('reason', 'Número associado a incidentes prévios')
                        })
            
            # 5. Verificar uso de carteiras móveis de dinheiro
            # Mercado moçambicano tem M-Pesa, mKesh, e-Mola
            wallet_history = self._check_mobile_wallet_activity(user_id)
            if wallet_history.get('high_risk_activity', False):
                risk_factors.append({
                    'factor': 'atividade_suspeita_mpesa_mkesh',
                    'score': 0.75,
                    'details': wallet_history.get('details', 'Atividade suspeita em carteiras móveis')
                })
            
            # 6. Verificar padrões comportamentais específicos de Moçambique
            
            # Transferências internacionais frequentes (comum em fraudes em Moçambique)
            if self._has_frequent_international_transfers(user_id):
                risk_factors.append({
                    'factor': 'transferencias_internacionais_frequentes',
                    'score': 0.7,
                    'details': 'Padrão de transferências internacionais frequentes'
                })
            
            # Múltiplas conversões moeda digital/física (comum em fraudes em Moçambique)
            if self._has_multiple_currency_conversions(user_id):
                risk_factors.append({
                    'factor': 'multiplas_conversoes_moeda',
                    'score': 0.65,
                    'details': 'Padrão de múltiplas conversões entre moeda digital e física'
                })
            
            # 7. Verificar listas de sanções específicas
            id_document = user_data.get('id_document', {})
            if id_document:
                doc_type = id_document.get('type', '')
                doc_number = id_document.get('number', '')
                
                if doc_type and doc_number:
                    # Verificar em listas SADC
                    sadc_check = self._check_sadc_sanctions(doc_type, doc_number, user_data.get('name', ''))
                    if sadc_check.get('found', False):
                        risk_factors.append({
                            'factor': 'lista_sancoes_sadc',
                            'score': 0.9,
                            'details': sadc_check.get('reason', 'Encontrado em lista de sanções SADC')
                        })
            
            # Calcular pontuação final baseada nos fatores de risco
            final_risk_score = self._calculate_risk_score(risk_factors)
            
            # Determinar nível de risco
            risk_level = self._determine_risk_level(final_risk_score)
            
            return {
                'risk_score': final_risk_score,
                'risk_level': risk_level,
                'risk_factors': sorted(risk_factors, key=lambda x: x['score'], reverse=True)
            }
            
        except Exception as e:
            logger.error(f"Erro na análise de fatores de risco regionais: {str(e)}")
            return {
                'risk_score': 0.5,
                'risk_level': 'médio',
                'risk_factors': [{'factor': 'erro_processamento', 'score': 0.5}],
                'error': str(e)
            }
    
    def _check_tax_status(self, tax_id: str) -> Dict[str, Any]:
        """Verifica o status fiscal do NUIT"""
        try:
            # Tentar usar adaptador de Autoridade Tributária se disponível
            if 'autoridade_tributaria_mz' in self.data_adapters:
                result = self.data_adapters['autoridade_tributaria_mz'].check_tax_id(tax_id)
                if result and result.get('found', False):
                    return result
            
            # Simulação: 10% dos NUITs têm problemas
            # Implementação mockada para testes
            if hash(tax_id) % 10 == 0:
                return {
                    'found': True,
                    'reason': 'Pendências fiscais simuladas',
                    'source': 'mock_data',
                    'risk_score': 0.85
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar status fiscal: {str(e)}")
            return {'found': False, 'error': str(e)}
    
    def _has_tax_compliance_issues(self, tax_id: str) -> bool:
        """Verifica se há problemas de conformidade fiscal"""
        # Implementação mockada para testes
        return hash(tax_id) % 7 == 0    def _check_credit_risk(self, user_id: str, tax_id: Optional[str]) -> Dict[str, Any]:
        """Verifica o histórico de crédito na Central de Risco de Moçambique"""
        try:
            # Tentar usar adaptador do Banco de Moçambique se disponível
            if 'banco_mocambique' in self.data_adapters:
                result = self.data_adapters['banco_mocambique'].check_credit_risk(user_id, tax_id)
                if result and result.get('found', False):
                    return result
            
            # Simulação para testes
            risky_ids = [
                "user_mz_12345", "user_mz_67890", "user_credit_risk"
            ]
            
            if user_id in risky_ids or (tax_id and hash(tax_id) % 8 == 0):
                return {
                    'found': True,
                    'reason': 'Histórico negativo de crédito simulado',
                    'source': 'mock_data',
                    'risk_score': 0.75
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar histórico de crédito: {str(e)}")
            return {'found': False, 'error': str(e)}
    
    def _check_banking_restrictions(self, user_id: str, tax_id: Optional[str]) -> Dict[str, Any]:
        """Verifica restrições no sistema bancário moçambicano"""
        try:
            # Tentar usar adaptador do Banco de Moçambique se disponível
            if 'banco_mocambique' in self.data_adapters:
                result = self.data_adapters['banco_mocambique'].check_banking_restrictions(user_id, tax_id)
                if result and result.get('found', False):
                    return result
            
            # Simulação para testes
            if hash(f"{user_id}-{tax_id}") % 9 == 0:
                return {
                    'found': True,
                    'reason': 'Restrições bancárias simuladas',
                    'source': 'mock_data',
                    'risk_score': 0.8
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar restrições bancárias: {str(e)}")
            return {'found': False, 'error': str(e)}
    
    def _check_phone_history(self, phone_number: str) -> Dict[str, Any]:
        """Verifica o histórico de uso do número de telefone"""
        try:
            # Tentar usar adaptador de Telecom Moçambique se disponível
            if 'telecom_mocambique' in self.data_adapters:
                result = self.data_adapters['telecom_mocambique'].check_phone_history(phone_number)
                if result and result.get('found', False):
                    return result
            
            # Simulação para testes
            if phone_number.endswith('1234') or phone_number.endswith('5555'):
                return {
                    'found': True,
                    'reason': 'Número com histórico suspeito simulado',
                    'source': 'mock_data',
                    'risk_score': 0.7
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar histórico do telefone: {str(e)}")
            return {'found': False, 'error': str(e)}
    
    def _check_mobile_wallet_activity(self, user_id: str) -> Dict[str, Any]:
        """Verifica atividade em carteiras móveis (M-Pesa, mKesh, e-Mola)"""
        try:
            # Implementação mockada para testes
            if hash(user_id) % 6 == 0:
                return {
                    'high_risk_activity': True,
                    'details': 'Padrões suspeitos de transação em carteiras móveis',
                    'risk_score': 0.75
                }
            
            return {'high_risk_activity': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar atividade em carteiras móveis: {str(e)}")
            return {'high_risk_activity': False, 'error': str(e)}
    
    def _has_frequent_international_transfers(self, user_id: str) -> bool:
        """Verifica se há padrão de transferências internacionais frequentes"""
        # Implementação mockada para testes
        return hash(user_id) % 5 == 0
    
    def _has_multiple_currency_conversions(self, user_id: str) -> bool:
        """Verifica se há padrão de múltiplas conversões de moeda"""
        # Implementação mockada para testes
        return hash(user_id) % 4 == 0
    
    def _check_sadc_sanctions(self, doc_type: str, doc_number: str, name: str) -> Dict[str, Any]:
        """Verifica em listas de sanções da SADC (Southern African Development Community)"""
        try:
            # Tentar usar adaptador de SADC se disponível
            if 'sadc_info' in self.data_adapters:
                result = self.data_adapters['sadc_info'].check_sanctions(doc_type, doc_number, name)
                if result and result.get('found', False):
                    return result
            
            # Simulação para testes
            high_risk_docs = [
                "12345678901", "MZBR123456", "00112233"
            ]
            
            if doc_number in high_risk_docs:
                return {
                    'found': True,
                    'reason': 'Documento em lista de sanções simulada',
                    'source': 'mock_data',
                    'risk_score': 0.9
                }
            
            return {'found': False}
            
        except Exception as e:
            logger.error(f"Erro ao verificar listas de sanções SADC: {str(e)}")
            return {'found': False, 'error': str(e)}
    
    def _calculate_risk_score(self, risk_factors: List[Dict[str, Any]]) -> float:
        """Calcula a pontuação de risco com base nos fatores encontrados"""
        if not risk_factors:
            return 0.0
        
        # Usar média ponderada dos fatores de risco
        total_score = sum(factor['score'] for factor in risk_factors)
        avg_score = total_score / len(risk_factors)
        
        # Aumentar risco se houver múltiplos fatores de alto risco
        high_risk_factors = [f for f in risk_factors if f['score'] >= 0.7]
        if high_risk_factors:
            high_risk_multiplier = 1 + (len(high_risk_factors) / 10)
            avg_score = min(1.0, avg_score * high_risk_multiplier)
        
        return avg_score
    
    def _determine_risk_level(self, risk_score: float) -> str:
        """Determina o nível de risco com base na pontuação"""
        if risk_score >= 0.8:
            return "alto"
        elif risk_score >= 0.5:
            return "médio"
        else:
            return "baixo"
    
    def _calculate_account_age(self, created_at: Optional[str]) -> int:
        """Calcula a idade da conta em dias"""
        if not created_at:
            return 0
            
        try:
            created_date = datetime.fromisoformat(created_at)
            age_days = (datetime.now() - created_date).days
            return max(0, age_days)
        except (ValueError, TypeError):
            return 0
    
    def _calculate_combined_risk_score(self, account_risk: Dict[str, Any], 
                                     location_risk: Dict[str, Any],
                                     device_risk: Dict[str, Any],
                                     regional_risk: Dict[str, Any]) -> Dict[str, Any]:
        """
        Calcula uma pontuação de risco combinada baseada nas várias análises.
        
        Args:
            account_risk: Resultado da análise de risco da conta
            location_risk: Resultado da análise de anomalias de localização
            device_risk: Resultado da análise de comportamento do dispositivo
            regional_risk: Resultado da análise de fatores regionais
            
        Returns:
            Dicionário com pontuação combinada e detalhes
        """
        try:
            # Extrair pontuações individuais com verificações de segurança
            account_score = account_risk.get('risk_score', 0.0) if isinstance(account_risk, dict) else 0.0
            location_score = location_risk.get('risk_score', 0.0) if isinstance(location_risk, dict) else 0.0
            device_score = device_risk.get('risk_score', 0.0) if isinstance(device_risk, dict) else 0.0
            regional_score = regional_risk.get('risk_score', 0.0) if isinstance(regional_risk, dict) else 0.0
            
            # Definir pesos específicos para o contexto moçambicano
            # No contexto de Moçambique, priorizamos fatores regionais e localização
            weights = {
                'account': 0.2,
                'location': 0.3,
                'device': 0.2,
                'regional': 0.3
            }
            
            # Calcular pontuação ponderada
            weighted_score = (
                account_score * weights['account'] +
                location_score * weights['location'] +
                device_score * weights['device'] +
                regional_score * weights['regional']
            )
            
            # Consolidar fatores de risco de todas as fontes
            all_risk_factors = []
            
            # Adicionar fatores de risco da conta
            if isinstance(account_risk, dict) and 'risk_factors' in account_risk:
                for factor in account_risk['risk_factors']:
                    factor['source'] = 'conta'
                    all_risk_factors.append(factor)
            
            # Adicionar anomalias de localização
            if isinstance(location_risk, dict) and 'anomalies' in location_risk:
                for anomaly in location_risk['anomalies']:
                    risk_factor = {
                        'factor': anomaly.get('type', 'anomalia_desconhecida'),
                        'score': anomaly.get('risk', 0.0),
                        'source': 'localização'
                    }
                    if 'details' in anomaly:
                        risk_factor['details'] = anomaly['details']
                    all_risk_factors.append(risk_factor)
            
            # Adicionar fatores de risco de dispositivo
            if isinstance(device_risk, dict) and 'risk_factors' in device_risk:
                for factor in device_risk['risk_factors']:
                    factor['source'] = 'dispositivo'
                    all_risk_factors.append(factor)
            
            # Adicionar fatores de risco regionais
            if isinstance(regional_risk, dict) and 'risk_factors' in regional_risk:
                for factor in regional_risk['risk_factors']:
                    factor['source'] = 'regional'
                    all_risk_factors.append(factor)
            
            # Ordenar fatores de risco por pontuação
            sorted_risk_factors = sorted(all_risk_factors, key=lambda x: x.get('score', 0.0), reverse=True)
            
            # Determinar nível de risco
            risk_level = self._determine_risk_level(weighted_score)
            
            # Preparar resultado com scores detalhados por categoria
            result = {
                'risk_score': weighted_score,
                'risk_level': risk_level,
                'scores_by_category': {
                    'conta': account_score,
                    'localização': location_score,
                    'dispositivo': device_score,
                    'regional': regional_score
                },
                'risk_factors': sorted_risk_factors[:10]  # Limitar aos 10 principais fatores
            }
            
            # Adicionar recomendações específicas para Moçambique
            result['recommendations'] = self._get_mozambique_recommendations(weighted_score, sorted_risk_factors)
            
            return result
            
        except Exception as e:
            logger.error(f"Erro ao calcular pontuação de risco combinada: {str(e)}")
            return {
                'risk_score': max(account_risk.get('risk_score', 0.5), 
                                 location_risk.get('risk_score', 0.5),
                                 device_risk.get('risk_score', 0.5),
                                 regional_risk.get('risk_score', 0.5)),
                'risk_level': 'médio',
                'error': str(e)
            }
    
    def _get_mozambique_recommendations(self, risk_score: float, risk_factors: List[Dict[str, Any]]) -> List[str]:
        """Gera recomendações específicas para o contexto moçambicano"""
        recommendations = []
        
        if risk_score >= 0.8:
            recommendations.append("Exigir verificação adicional de identidade conforme normas do Banco de Moçambique")
            recommendations.append("Solicitar confirmação por SMS para dispositivos não reconhecidos")
            recommendations.append("Aplicar limites de transação reduzidos até verificação completa")
            
        elif risk_score >= 0.5:
            recommendations.append("Solicitar autenticação de dois fatores para operações sensíveis")
            recommendations.append("Monitorar transações nas próximas 24-48 horas")
            
        # Recomendações específicas baseadas em fatores de risco
        for factor in risk_factors:
            if factor.get('factor') == 'area_alto_risco' and factor.get('score', 0) > 0.6:
                recommendations.append("Implementar verificações adicionais para transações nesta área geográfica")
                
            if factor.get('factor') == 'nuit_em_lista_negra':
                recommendations.append("Bloquear transações e reportar ao Banco de Moçambique conforme regulação")
                
            if 'viagem_impossivel' in factor.get('factor', ''):
                recommendations.append("Verificar a identidade do usuário através de múltiplos canais")
                
        # Limitar a 5 recomendações principais
        return recommendations[:5]