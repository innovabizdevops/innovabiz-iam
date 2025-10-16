#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Adaptador para Bureau de Créditos

Este módulo implementa o adaptador para integração com Bureaus de Crédito,
permitindo consultas de histórico creditício para análise comportamental
e avaliação de risco.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

import logging
import time
import json
import hashlib
import hmac
import requests
import datetime
from typing import Dict, Any, Optional, Union, List
from abc import ABC, abstractmethod

# Configuração de logging
logger = logging.getLogger("bureau_credito_adapter")


class BaseBureauCreditoAdapter(ABC):
    """
    Classe base para adaptadores de Bureau de Crédito.
    
    Define a interface comum que todos os adaptadores de Bureau de Crédito
    devem implementar, independente da região ou provedor específico.
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        Inicializa o adaptador base.
        
        Args:
            config: Configuração do adaptador com credenciais e endpoints
        """
        self.config = config
        self.base_url = config.get("base_url", "")
        self.api_key = config.get("api_key", "")
        self.api_secret = config.get("api_secret", "")
        self.timeout = config.get("timeout", 5)  # Timeout em segundos
        self.headers = {
            "Content-Type": "application/json",
            "Accept": "application/json"
        }
        
        # Credenciais específicas por implementação
        self._setup_credentials()
        
        logger.info(f"Adaptador Base Bureau de Crédito inicializado: {self.__class__.__name__}")
    
    @abstractmethod
    def _setup_credentials(self) -> None:
        """
        Configura credenciais específicas para o provedor de Bureau de Crédito.
        
        Esta função deve ser implementada por cada adaptador específico.
        """
        pass
    
    @abstractmethod
    def get_credit_report(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Obtém relatório de crédito para o usuário especificado.
        
        Args:
            user_data: Dados do usuário para consulta
            
        Returns:
            Dict: Relatório de crédito ou dados de erro
        """
        pass
    
    @abstractmethod
    def check_credit_score(self, user_id: str) -> Dict[str, Any]:
        """
        Consulta apenas o score de crédito do usuário.
        
        Args:
            user_id: Identificador do usuário
            
        Returns:
            Dict: Score de crédito ou dados de erro
        """
        pass
    
    def validate_response(self, response: requests.Response) -> Dict[str, Any]:
        """
        Valida e processa resposta da API do Bureau de Crédito.
        
        Args:
            response: Objeto de resposta da requisição
            
        Returns:
            Dict: Dados da resposta ou erro
        """
        result = {
            "success": False,
            "data": None,
            "error": None,
            "status_code": response.status_code,
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        try:
            if response.status_code >= 200 and response.status_code < 300:
                result["success"] = True
                result["data"] = response.json()
            else:
                result["error"] = {
                    "code": response.status_code,
                    "message": response.text
                }
                logger.warning(f"Erro na resposta do Bureau de Crédito: {response.status_code} - {response.text}")
        except Exception as e:
            result["error"] = {
                "code": "PARSE_ERROR",
                "message": str(e)
            }
            logger.error(f"Erro ao processar resposta do Bureau de Crédito: {str(e)}")
        
        return result
    
    def _generate_hmac_signature(self, payload: Dict[str, Any], secret: str) -> str:
        """
        Gera assinatura HMAC para autenticação segura com Bureau de Crédito.
        
        Args:
            payload: Dados a serem assinados
            secret: Chave secreta para assinatura
            
        Returns:
            str: Assinatura gerada
        """
        # Converter payload para string JSON ordenada
        payload_str = json.dumps(payload, sort_keys=True)
        # Gerar assinatura HMAC-SHA256
        signature = hmac.new(
            secret.encode('utf-8'),
            payload_str.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()
        
        return signature
    
    def _log_request_metrics(self, 
                            endpoint: str, 
                            start_time: float, 
                            status_code: int, 
                            success: bool) -> None:
        """
        Registra métricas de requisição para observabilidade.
        
        Args:
            endpoint: Endpoint da API acessado
            start_time: Timestamp de início da requisição
            status_code: Código de status HTTP
            success: Indicador de sucesso da requisição
        """
        duration = time.time() - start_time
        logger.info(
            f"Bureau Request: endpoint={endpoint}, "
            f"duration={duration:.3f}s, "
            f"status={status_code}, "
            f"success={success}"
        )


class AngolaBureauCreditoAdapter(BaseBureauCreditoAdapter):
    """
    Adaptador para Bureau de Crédito de Angola.
    
    Implementa a integração específica com o Bureau de Crédito de Angola,
    considerando requisitos regulatórios, formatos de dados e endpoints específicos.
    """
    
    def _setup_credentials(self) -> None:
        """
        Configura credenciais específicas para o Bureau de Crédito de Angola.
        """
        # Adicionar headers específicos para Angola
        self.headers["X-Angola-Bureau-ApiKey"] = self.api_key
        self.headers["X-Angola-Bureau-Version"] = self.config.get("api_version", "1.0")
        
        # Licença bancária específica para Angola (obrigatória para consultas)
        self.angola_banking_license = self.config.get("banking_license", "")
        if self.angola_banking_license:
            self.headers["X-Angola-Banking-License"] = self.angola_banking_license
        
        # Configuração regional
        self.region = "AO"  # Angola
        
        # Endpoints específicos para Angola
        self.endpoints = {
            "credit_report": f"{self.base_url}/credit/angola/report",
            "credit_score": f"{self.base_url}/credit/angola/score",
            "account_status": f"{self.base_url}/credit/angola/account_status",
            "payment_history": f"{self.base_url}/credit/angola/payment_history",
            "loan_summary": f"{self.base_url}/credit/angola/loan_summary"
        }
    
    def get_credit_report(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Obtém relatório de crédito completo para o usuário em Angola.
        
        Args:
            user_data: Dados do usuário para consulta
            
        Returns:
            Dict: Relatório de crédito ou dados de erro
        """
        result = {
            "success": False,
            "data": None,
            "error": None,
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        try:
            # Preparar payload específico para Angola
            payload = self._prepare_angola_payload(user_data)
            
            # Adicionar assinatura HMAC para autenticação
            timestamp = str(int(time.time()))
            self.headers["X-Angola-Bureau-Timestamp"] = timestamp
            signature = self._generate_hmac_signature({
                "payload": payload,
                "timestamp": timestamp
            }, self.api_secret)
            self.headers["X-Angola-Bureau-Signature"] = signature
            
            # Registrar início da requisição
            start_time = time.time()
            
            # Fazer requisição ao Bureau de Crédito de Angola
            response = requests.post(
                self.endpoints["credit_report"],
                headers=self.headers,
                json=payload,
                timeout=self.timeout
            )
            
            # Validar e processar resposta
            result = self.validate_response(response)
            
            # Registrar métricas da requisição
            self._log_request_metrics(
                "credit_report", 
                start_time, 
                response.status_code,
                result["success"]
            )
            
            # Processar dados específicos de Angola se houver sucesso
            if result["success"] and result["data"]:
                result["data"] = self._normalize_angola_report(result["data"])
                
        except Exception as e:
            result["error"] = {
                "code": "REQUEST_ERROR",
                "message": str(e)
            }
            logger.error(f"Erro ao consultar relatório de crédito em Angola: {str(e)}")
        
        return result    def check_credit_score(self, user_id: str) -> Dict[str, Any]:
        """
        Consulta apenas o score de crédito do usuário em Angola.
        
        Args:
            user_id: Identificador do usuário
            
        Returns:
            Dict: Score de crédito ou dados de erro
        """
        result = {
            "success": False,
            "data": None,
            "error": None,
            "timestamp": datetime.datetime.now().isoformat()
        }
        
        try:
            # Preparar payload simplificado para consulta rápida
            payload = {
                "user_id": user_id,
                "region": self.region,
                "request_type": "score_only",
                "request_id": f"score_{int(time.time())}_{user_id}"
            }
            
            # Adicionar assinatura HMAC para autenticação
            timestamp = str(int(time.time()))
            self.headers["X-Angola-Bureau-Timestamp"] = timestamp
            signature = self._generate_hmac_signature({
                "payload": payload,
                "timestamp": timestamp
            }, self.api_secret)
            self.headers["X-Angola-Bureau-Signature"] = signature
            
            # Registrar início da requisição
            start_time = time.time()
            
            # Fazer requisição ao Bureau de Crédito de Angola
            response = requests.post(
                self.endpoints["credit_score"],
                headers=self.headers,
                json=payload,
                timeout=self.timeout
            )
            
            # Validar e processar resposta
            result = self.validate_response(response)
            
            # Registrar métricas da requisição
            self._log_request_metrics(
                "credit_score", 
                start_time, 
                response.status_code,
                result["success"]
            )
            
            # Normalizar dados de score
            if result["success"] and result["data"]:
                result["data"] = {
                    "credit_score": result["data"].get("score", 0),
                    "score_range": result["data"].get("range", {"min": 0, "max": 1000}),
                    "risk_category": result["data"].get("risk_category", "unknown"),
                    "score_date": result["data"].get("evaluation_date", datetime.datetime.now().isoformat())
                }
                
        except Exception as e:
            result["error"] = {
                "code": "REQUEST_ERROR",
                "message": str(e)
            }
            logger.error(f"Erro ao consultar score de crédito em Angola: {str(e)}")
        
        return result
    
    def _prepare_angola_payload(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Prepara payload específico para Bureau de Crédito de Angola.
        
        Args:
            user_data: Dados do usuário
            
        Returns:
            Dict: Payload formatado
        """
        # Extrair dados básicos
        user_id = user_data.get("user_id", "")
        full_name = user_data.get("full_name", "")
        
        # Documentos de identificação angolanos
        id_bilhete = user_data.get("id_bilhete", "")  # Bilhete de Identidade angolano
        nif = user_data.get("nif", "")  # Número de Identificação Fiscal angolano
        passport = user_data.get("passport", "")
        
        # Criar payload base
        payload = {
            "request_id": f"req_{int(time.time())}_{user_id}",
            "request_type": "full_report",
            "region": self.region,
            "client_reference": self.config.get("client_reference", "innovabiz-iam"),
            "requesting_entity": self.config.get("entity_name", "InnovaBiz IAM"),
            "purpose_code": self.config.get("purpose_code", "FRAUD_PREVENTION"),
            "subject": {
                "user_id": user_id,
                "full_name": full_name,
                "identification": []
            }
        }
        
        # Adicionar documentos de identificação disponíveis
        if id_bilhete:
            payload["subject"]["identification"].append({
                "type": "BI",  # Bilhete de Identidade
                "value": id_bilhete
            })
        
        if nif:
            payload["subject"]["identification"].append({
                "type": "NIF",  # Número de Identificação Fiscal
                "value": nif
            })
        
        if passport:
            payload["subject"]["identification"].append({
                "type": "PASSPORT",
                "value": passport
            })
        
        # Dados adicionais opcionais
        if "phone_number" in user_data:
            payload["subject"]["phone_number"] = user_data["phone_number"]
            
        if "address" in user_data:
            payload["subject"]["address"] = user_data["address"]
            
        if "date_of_birth" in user_data:
            payload["subject"]["date_of_birth"] = user_data["date_of_birth"]
        
        return payload
    
    def _normalize_angola_report(self, report_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Normaliza dados de relatório do Bureau de Angola para formato padrão.
        
        Args:
            report_data: Dados brutos do relatório
            
        Returns:
            Dict: Dados normalizados
        """
        normalized = {
            "credit_score": 0,
            "payment_defaults": 0,
            "active_loans": 0,
            "recent_inquiries": 0,
            "account_summary": {},
            "payment_history": [],
            "risk_indicators": [],
            "score_history": []
        }
        
        try:
            # Extrair score de crédito
            if "credit_score" in report_data:
                normalized["credit_score"] = report_data["credit_score"].get("score", 0)
                
            # Extrair inadimplências
            if "payment_history" in report_data:
                payment_history = report_data["payment_history"]
                normalized["payment_history"] = payment_history
                
                # Contar inadimplências
                for payment in payment_history:
                    if payment.get("status") == "DEFAULT":
                        normalized["payment_defaults"] += 1
            
            # Extrair empréstimos ativos
            if "loans" in report_data:
                active_loans = [loan for loan in report_data["loans"] 
                               if loan.get("status") == "ACTIVE"]
                normalized["active_loans"] = len(active_loans)
            
            # Extrair consultas recentes
            if "inquiries" in report_data:
                # Filtrar consultas nos últimos 90 dias
                recent_date = datetime.datetime.now() - datetime.timedelta(days=90)
                recent_inquiries = [
                    inq for inq in report_data["inquiries"]
                    if datetime.datetime.fromisoformat(inq.get("inquiry_date", "")) >= recent_date
                ]
                normalized["recent_inquiries"] = len(recent_inquiries)
            
            # Extrair resumo de contas
            if "account_summary" in report_data:
                normalized["account_summary"] = report_data["account_summary"]
            
            # Extrair indicadores de risco
            if "risk_indicators" in report_data:
                normalized["risk_indicators"] = report_data["risk_indicators"]
            
            # Extrair histórico de score
            if "score_history" in report_data:
                normalized["score_history"] = report_data["score_history"]
                
        except Exception as e:
            logger.error(f"Erro ao normalizar relatório do Bureau de Angola: {str(e)}")
        
        return normalized


class BrazilBureauCreditoAdapter(BaseBureauCreditoAdapter):
    """
    Adaptador para Bureau de Crédito do Brasil (Serasa, Boa Vista, SPC, etc.).
    
    Implementa a integração específica com Bureaus de Crédito do Brasil,
    considerando requisitos regulatórios, formatos de dados e endpoints específicos.
    """
    
    def _setup_credentials(self) -> None:
        """
        Configura credenciais específicas para o Bureau de Crédito do Brasil.
        """
        # Implementação para Brasil será adicionada em uma iteração futura
        pass
    
    def get_credit_report(self, user_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Obtém relatório de crédito para o usuário no Brasil.
        
        Args:
            user_data: Dados do usuário para consulta
            
        Returns:
            Dict: Relatório de crédito ou dados de erro
        """
        # Implementação para Brasil será adicionada em uma iteração futura
        return {
            "success": False,
            "error": {"code": "NOT_IMPLEMENTED", "message": "Brasil Bureau não implementado"}
        }
    
    def check_credit_score(self, user_id: str) -> Dict[str, Any]:
        """
        Consulta apenas o score de crédito do usuário no Brasil.
        
        Args:
            user_id: Identificador do usuário
            
        Returns:
            Dict: Score de crédito ou dados de erro
        """
        # Implementação para Brasil será adicionada em uma iteração futura
        return {
            "success": False,
            "error": {"code": "NOT_IMPLEMENTED", "message": "Brasil Bureau não implementado"}
        }


# Fábrica de adaptadores de Bureau de Crédito
def create_bureau_adapter(region: str, config: Dict[str, Any]) -> BaseBureauCreditoAdapter:
    """
    Cria um adaptador apropriado para o Bureau de Crédito da região especificada.
    
    Args:
        region: Código da região (ISO 3166-1 alpha-2)
        config: Configuração do adaptador
        
    Returns:
        BaseBureauCreditoAdapter: Instância do adaptador configurado
    """
    region_upper = region.upper()
    
    if region_upper == "AO":
        return AngolaBureauCreditoAdapter(config)
    elif region_upper == "BR":
        return BrazilBureauCreditoAdapter(config)
    else:
        logger.warning(f"Região não suportada para Bureau de Crédito: {region}")
        raise ValueError(f"Região não suportada para Bureau de Crédito: {region}")