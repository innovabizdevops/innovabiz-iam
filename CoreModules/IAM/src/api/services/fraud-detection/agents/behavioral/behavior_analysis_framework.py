"""
Framework para Análise Comportamental Regional

Este módulo define a estrutura base para os agentes de análise comportamental 
adaptados para diferentes contextos regionais (CPLP, SADC, PALOP, BRICS),
permitindo a detecção contextual de fraudes com sensibilidade cultural e regional.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

import os
import logging
import json
import uuid
import time
import hashlib
from datetime import datetime, timedelta
from typing import Dict, List, Tuple, Optional, Union, Any
from abc import ABC, abstractmethod

# Configuração do logging
logger = logging.getLogger("behavior_analysis")

class BehaviorAnalysisAgent(ABC):
    """
    Classe abstrata base para os agentes de análise comportamental.
    Define a interface e comportamentos comuns para todos os agentes regionais.
    """
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        model_path: Optional[str] = None,
        cache_dir: Optional[str] = None,
        region_code: str = "DEFAULT",
        data_sources: Optional[List[str]] = None,
        language: str = "pt"
    ):
        """
        Inicializa o agente de análise comportamental.
        
        Args:
            config_path: Caminho para arquivo de configuração
            model_path: Caminho para modelos pré-treinados
            cache_dir: Diretório para cache de resultados
            region_code: Código da região (país ou grupo regional)
            data_sources: Lista de fontes de dados a serem utilizadas
            language: Código do idioma principal (pt, en, es, etc.)
        """
        self.region_code = region_code
        self.language = language
        
        # Carregar configuração
        self.config = self._load_config(config_path)
        
        # Definir fontes de dados
        self.data_sources = data_sources or self.config.get("data_sources", [])
        
        # Cache e persistência
        self.cache_dir = cache_dir or self.config.get("cache_dir", "./cache")
        os.makedirs(self.cache_dir, exist_ok=True)
        
        # Carregar modelos de análise
        self.models = {}
        self.model_path = model_path or self.config.get("model_path", "./models")
        self._load_models()
        
        # Métricas e monitoramento
        self.metrics = {
            "analysis_count": 0,
            "cache_hits": 0,
            "fraud_detected": 0,
            "total_processing_time": 0,
            "last_execution_time": None
        }
        
        # Configurações específicas de região
        self.region_config = self._load_region_config()
        
        # Adapters para fontes de dados
        self.data_adapters = {}
        self._initialize_data_adapters()
        
        logger.info(f"Agente de análise comportamental inicializado para região: {self.region_code}")
    
    def _load_config(self, config_path: Optional[str]) -> Dict:
        """
        Carrega a configuração do agente de arquivo JSON.
        
        Args:
            config_path: Caminho para o arquivo de configuração
            
        Returns:
            Dicionário com configurações
        """
        default_config = {
            "threshold_low": 0.3,
            "threshold_medium": 0.6,
            "threshold_high": 0.8,
            "cache_ttl_seconds": 3600,
            "use_contextual_rules": True,
            "enable_adaptive_learning": True,
            "data_sources": [],
            "model_path": "./models",
            "cache_dir": "./cache"
        }
        
        if config_path and os.path.exists(config_path):
            try:
                with open(config_path, 'r', encoding='utf-8') as f:
                    loaded_config = json.load(f)
                    # Mesclar com valores padrão
                    config = {**default_config, **loaded_config}
                    logger.info(f"Configuração carregada de {config_path}")
                    return config
            except Exception as e:
                logger.error(f"Erro ao carregar configuração: {str(e)}")
                
        logger.warning(f"Usando configuração padrão para o agente de análise comportamental")
        return default_config
    
    def _load_region_config(self) -> Dict:
        """
        Carrega a configuração específica da região.
        Estas configurações podem incluir limiares específicos, regras contextuais,
        pesos para diferentes fatores de risco, etc.
        
        Returns:
            Dicionário com configurações regionais
        """
        region_config_path = os.path.join(
            os.path.dirname(self.model_path),
            f"region_config_{self.region_code.lower()}.json"
        )
        
        default_region_config = {
            "behavioral_patterns": {},
            "cultural_factors": {},
            "regional_risk_weights": {},
            "custom_rules": []
        }
        
        if os.path.exists(region_config_path):
            try:
                with open(region_config_path, 'r', encoding='utf-8') as f:
                    loaded_config = json.load(f)
                    # Mesclar com valores padrão
                    config = {**default_region_config, **loaded_config}
                    logger.info(f"Configuração regional carregada para {self.region_code}")
                    return config
            except Exception as e:
                logger.error(f"Erro ao carregar configuração regional: {str(e)}")
                
        logger.warning(f"Usando configuração regional padrão para {self.region_code}")
        return default_region_config
    
    def _load_models(self) -> None:
        """
        Carrega os modelos de análise comportamental.
        Pode incluir modelos de machine learning, regras de negócio, etc.
        """
        # Implementação básica - deve ser expandida nas classes concretas
        model_dir = os.path.join(self.model_path, self.region_code.lower())
        os.makedirs(model_dir, exist_ok=True)
        
        # Log da tentativa de carregar modelos
        logger.info(f"Tentando carregar modelos de {model_dir}")
        
        # Verificar se existem modelos salvos
        model_files = [f for f in os.listdir(model_dir) if f.endswith('.model')] if os.path.exists(model_dir) else []
        
        if not model_files:
            logger.warning(f"Nenhum modelo encontrado em {model_dir}. Usando configurações padrão.")
        else:
            for model_file in model_files:
                model_name = model_file.split('.')[0]
                model_path = os.path.join(model_dir, model_file)
                # A implementação específica de carregamento deve ser feita nas subclasses
                logger.info(f"Modelo {model_name} disponível em {model_path}")
    
    def _initialize_data_adapters(self) -> None:
        """
        Inicializa os adaptadores para as fontes de dados configuradas.
        Este método deve ser implementado pelas subclasses para conectar às fontes de dados específicas.
        """
        # Este método será implementado pelas subclasses
        pass
    
    def get_cache_key(self, entity_id: str, analysis_type: str, context_data: Dict = None) -> str:
        """
        Gera uma chave de cache para armazenar/recuperar resultados de análise.
        
        Args:
            entity_id: ID da entidade analisada
            analysis_type: Tipo de análise
            context_data: Dados de contexto relevantes
            
        Returns:
            Chave de cache única
        """
        context_str = json.dumps(context_data or {}, sort_keys=True) if context_data else ""
        base_string = f"{self.region_code}:{entity_id}:{analysis_type}:{context_str}"
        return hashlib.sha256(base_string.encode('utf-8')).hexdigest()
    
    def get_cached_result(self, cache_key: str, max_age_seconds: int = None) -> Optional[Dict]:
        """
        Recupera um resultado de análise do cache, se disponível e dentro do prazo de validade.
        
        Args:
            cache_key: Chave de cache para buscar
            max_age_seconds: Idade máxima aceitável para o cache em segundos
                             (None usa o padrão da configuração)
            
        Returns:
            Resultado em cache ou None se não encontrado/expirado
        """
        if max_age_seconds is None:
            max_age_seconds = self.config.get("cache_ttl_seconds", 3600)
            
        cache_file = os.path.join(self.cache_dir, f"{cache_key}.json")
        
        if not os.path.exists(cache_file):
            return None
            
        try:
            # Verificar idade do arquivo
            file_age = time.time() - os.path.getmtime(cache_file)
            
            if file_age > max_age_seconds:
                logger.debug(f"Cache expirado para {cache_key} (idade: {file_age:.1f}s)")
                return None
                
            # Carregar dados do cache
            with open(cache_file, 'r', encoding='utf-8') as f:
                cached_data = json.load(f)
                
            self.metrics["cache_hits"] += 1
            logger.debug(f"Cache hit para {cache_key}")
            return cached_data
            
        except Exception as e:
            logger.error(f"Erro ao recuperar cache: {str(e)}")
            return None
    
    def save_to_cache(self, cache_key: str, result: Dict) -> bool:
        """
        Salva um resultado de análise no cache.
        
        Args:
            cache_key: Chave de cache
            result: Resultado da análise a ser armazenado
            
        Returns:
            True se o salvamento foi bem-sucedido, False caso contrário
        """
        cache_file = os.path.join(self.cache_dir, f"{cache_key}.json")
        
        try:
            # Adicionar metadados ao cache
            result_with_metadata = {
                **result,
                "cache_created_at": datetime.now().isoformat(),
                "cache_key": cache_key
            }
            
            # Salvar no arquivo
            with open(cache_file, 'w', encoding='utf-8') as f:
                json.dump(result_with_metadata, f, ensure_ascii=False, indent=2)
                
            logger.debug(f"Resultado salvo em cache: {cache_key}")
            return True
            
        except Exception as e:
            logger.error(f"Erro ao salvar cache: {str(e)}")
            return False
    
    def update_metrics(self, start_time: float, fraud_detected: bool) -> None:
        """
        Atualiza métricas de desempenho e detecção.
        
        Args:
            start_time: Timestamp de início da análise
            fraud_detected: Se fraude foi detectada
        """
        execution_time = time.time() - start_time
        
        self.metrics["analysis_count"] += 1
        self.metrics["total_processing_time"] += execution_time
        self.metrics["last_execution_time"] = execution_time
        
        if fraud_detected:
            self.metrics["fraud_detected"] += 1
    
    @abstractmethod
    def analyze_behavior(self, entity_id: str, entity_type: str, transaction_data: Dict, 
                       context_data: Optional[Dict] = None, use_cache: bool = True) -> Dict:
        """
        Analisa o comportamento de uma entidade para detectar padrões fraudulentos.
        
        Args:
            entity_id: ID da entidade (usuário, conta, dispositivo, etc)
            entity_type: Tipo de entidade (user, account, device, merchant, etc)
            transaction_data: Dados da transação ou ação sendo analisada
            context_data: Informações contextuais adicionais
            use_cache: Se deve usar cache para resultados prévios
            
        Returns:
            Resultado da análise de comportamento com score de risco
        """
        pass
    
    @abstractmethod
    def analyze_transaction_pattern(self, entity_id: str, transactions: List[Dict], 
                                 context_data: Optional[Dict] = None) -> Dict:
        """
        Analisa padrões em múltiplas transações para detectar anomalias.
        
        Args:
            entity_id: ID da entidade
            transactions: Lista de transações para análise
            context_data: Informações contextuais adicionais
            
        Returns:
            Resultado da análise de padrões com anomalias identificadas
        """
        pass
    
    @abstractmethod
    def evaluate_account_risk(self, entity_id: str, account_data: Dict, 
                           history_data: Optional[Dict] = None) -> Dict:
        """
        Avalia o nível de risco de uma conta com base em seu perfil e histórico.
        
        Args:
            entity_id: ID da entidade
            account_data: Dados da conta
            history_data: Dados históricos da conta
            
        Returns:
            Avaliação de risco da conta
        """
        pass
    
    @abstractmethod
    def detect_location_anomalies(self, entity_id: str, location_data: Dict, 
                               history: Optional[List[Dict]] = None) -> Dict:
        """
        Detecta anomalias de localização com sensibilidade ao contexto regional.
        
        Args:
            entity_id: ID da entidade
            location_data: Dados de localização atuais
            history: Histórico de localizações
            
        Returns:
            Resultado da detecção de anomalias de localização
        """
        pass
    
    @abstractmethod
    def analyze_device_behavior(self, entity_id: str, device_data: Dict, 
                             session_data: Optional[Dict] = None) -> Dict:
        """
        Analisa comportamento de dispositivos para detectar padrões suspeitos.
        
        Args:
            entity_id: ID da entidade
            device_data: Dados do dispositivo
            session_data: Dados da sessão atual
            
        Returns:
            Resultado da análise de comportamento do dispositivo
        """
        pass
    
    @abstractmethod
    def get_regional_risk_factors(self, entity_id: str, entity_type: str) -> Dict:
        """
        Obtém fatores de risco específicos da região para uma entidade.
        
        Args:
            entity_id: ID da entidade
            entity_type: Tipo de entidade
            
        Returns:
            Fatores de risco regionais
        """
        pass
    
    def format_response(self, analysis_result: Dict, include_details: bool = True) -> Dict:
        """
        Formata a resposta da análise comportamental em formato padronizado.
        
        Args:
            analysis_result: Resultado bruto da análise
            include_details: Se deve incluir detalhes completos ou versão resumida
            
        Returns:
            Resposta formatada
        """
        # Gerar ID da análise
        analysis_id = str(uuid.uuid4())
        timestamp = datetime.now().isoformat()
        
        # Formatar resposta básica
        response = {
            "analysis_id": analysis_id,
            "timestamp": timestamp,
            "region_code": self.region_code,
            "risk_score": analysis_result.get("risk_score", 0),
            "risk_level": analysis_result.get("risk_level", "low"),
            "recommended_action": analysis_result.get("recommended_action", "allow")
        }
        
        # Incluir detalhes se solicitado
        if include_details:
            response["details"] = analysis_result.get("details", {})
            response["risk_factors"] = analysis_result.get("risk_factors", [])
            response["analysis_type"] = analysis_result.get("analysis_type", "behavioral")
            response["data_sources"] = analysis_result.get("data_sources", [])
        
        return response
    
    def get_metrics(self) -> Dict:
        """
        Retorna métricas atuais de desempenho e detecção.
        
        Returns:
            Dicionário com métricas
        """
        avg_time = (self.metrics["total_processing_time"] / self.metrics["analysis_count"] 
                   if self.metrics["analysis_count"] > 0 else 0)
        
        detection_rate = (self.metrics["fraud_detected"] / self.metrics["analysis_count"] * 100
                         if self.metrics["analysis_count"] > 0 else 0)
        
        cache_rate = (self.metrics["cache_hits"] / self.metrics["analysis_count"] * 100
                     if self.metrics["analysis_count"] > 0 else 0)
        
        return {
            "region_code": self.region_code,
            "analysis_count": self.metrics["analysis_count"],
            "fraud_detected": self.metrics["fraud_detected"],
            "detection_rate_percent": round(detection_rate, 2),
            "cache_hits": self.metrics["cache_hits"],
            "cache_hit_rate_percent": round(cache_rate, 2),
            "average_processing_time": round(avg_time, 3),
            "last_execution_time": round(self.metrics["last_execution_time"], 3) if self.metrics["last_execution_time"] else None
        }