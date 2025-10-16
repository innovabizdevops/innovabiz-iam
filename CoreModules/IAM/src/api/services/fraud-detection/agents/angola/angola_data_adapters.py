"""
Adaptadores para Fontes de Dados Específicas do Mercado Angolano

Este módulo implementa adaptadores para integração com fontes de dados específicas
do mercado angolano, incluindo sistemas bancários, operadoras de telecomunicações,
instituições governamentais e fontes de dados externas relevantes.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import os
import json
import logging
import requests
import pandas as pd
from datetime import datetime, timedelta
from typing import Dict, List, Tuple, Optional, Union, Any
from abc import ABC, abstractmethod

# Configuração de logging
logger = logging.getLogger("angola_data_adapters")
logger.setLevel(logging.INFO)

# Adaptador base para todas as fontes de dados angolanas
class AngolaDataAdapter(ABC):
    """Classe base para todos os adaptadores de fontes de dados angolanas."""
    
    def __init__(
        self,
        config_path: Optional[str] = None,
        credentials_path: Optional[str] = None,
        cache_dir: Optional[str] = None
    ):
        """
        Inicializa o adaptador base.
        
        Args:
            config_path: Caminho para arquivo de configuração
            credentials_path: Caminho para arquivo de credenciais
            cache_dir: Diretório para cache de dados
        """
        self.config = {}
        self.credentials = {}
        self.cache_dir = cache_dir or os.path.join(
            os.path.dirname(__file__), "cache"
        )
        
        # Criar diretório de cache se não existir
        os.makedirs(self.cache_dir, exist_ok=True)
        
        # Carregar configuração
        if config_path and os.path.exists(config_path):
            with open(config_path, 'r') as f:
                self.config = json.load(f)
                
        # Carregar credenciais
        if credentials_path and os.path.exists(credentials_path):
            with open(credentials_path, 'r') as f:
                self.credentials = json.load(f)
                
        # Inicializar sessão HTTP
        self.session = requests.Session()
        
        # Status da conexão
        self.connection_status = {
            "is_connected": False,
            "last_connection_time": None,
            "connection_errors": []
        }
    
    @abstractmethod
    def connect(self) -> bool:
        """
        Estabelece conexão com a fonte de dados.
        
        Returns:
            True se a conexão foi estabelecida com sucesso, False caso contrário.
        """
        pass
    
    @abstractmethod
    def fetch_data(self, query_params: Dict) -> Dict:
        """
        Busca dados da fonte.
        
        Args:
            query_params: Parâmetros de consulta
            
        Returns:
            Dados obtidos da fonte
        """
        pass
    
    def cache_data(self, key: str, data: Any) -> None:
        """
        Armazena dados em cache para uso futuro.
        
        Args:
            key: Chave de identificação para os dados
            data: Dados a serem armazenados
        """
        cache_file = os.path.join(self.cache_dir, f"{key}.json")
        
        try:
            with open(cache_file, 'w') as f:
                json.dump({
                    "timestamp": datetime.now().isoformat(),
                    "data": data
                }, f)
        except Exception as e:
            logger.error(f"Erro ao armazenar cache para {key}: {str(e)}")
    
    def get_cached_data(self, key: str, max_age_hours: int = 24) -> Optional[Any]:
        """
        Recupera dados do cache se disponíveis e não expirados.
        
        Args:
            key: Chave de identificação dos dados
            max_age_hours: Idade máxima do cache em horas
            
        Returns:
            Dados em cache ou None se não disponíveis ou expirados
        """
        cache_file = os.path.join(self.cache_dir, f"{key}.json")
        
        if not os.path.exists(cache_file):
            return None
            
        try:
            with open(cache_file, 'r') as f:
                cache_data = json.load(f)
                
            # Verificar se o cache expirou
            timestamp = datetime.fromisoformat(cache_data["timestamp"])
            age = datetime.now() - timestamp
            
            if age > timedelta(hours=max_age_hours):
                logger.info(f"Cache expirado para {key}")
                return None
                
            return cache_data["data"]
        except Exception as e:
            logger.error(f"Erro ao ler cache para {key}: {str(e)}")
            return None
    
    def handle_error(self, error_type: str, error_message: str, retry: bool = False) -> Dict:
        """
        Manipula erros de conexão ou processamento.
        
        Args:
            error_type: Tipo de erro
            error_message: Mensagem de erro
            retry: Se deve tentar novamente automaticamente
            
        Returns:
            Dicionário com informações do erro
        """
        error_info = {
            "error_type": error_type,
            "error_message": error_message,
            "timestamp": datetime.now().isoformat(),
            "retry_attempted": retry
        }
        
        self.connection_status["connection_errors"].append(error_info)
        logger.error(f"{error_type}: {error_message}")
        
        return error_info
    
    def get_connection_status(self) -> Dict:
        """
        Retorna o status atual da conexão.
        
        Returns:
            Status da conexão
        """
        return self.connection_status