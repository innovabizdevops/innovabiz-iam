import json
import logging
import os
import time
from typing import Dict, Any, List, Optional, Union, Tuple
from datetime import datetime, timedelta

import requests
from requests.exceptions import RequestException
import backoff

from .trust_guard_models import (
    VerificationRequest,
    VerificationResponse,
    VerificationStatus,
    VerificationScoreLevel,
    VerificationDetails,
    ComplianceStatus
)

logger = logging.getLogger(__name__)

class TrustGuardService:
    """
    Serviço para integração com TrustGuard para verificação de identidade.
    """
    
    def __init__(self):
        """Inicializa o serviço TrustGuard com configurações de ambiente."""
        self.api_base_url = os.environ.get('TRUSTGUARD_API_URL', 'https://api.trustguard.innovabiz.dev')
        self.api_key = os.environ.get('TRUSTGUARD_API_KEY', '')
        self.timeout = int(os.environ.get('TRUSTGUARD_TIMEOUT_SECONDS', '30'))
        
        # Configurações de retry
        self.max_retries = int(os.environ.get('TRUSTGUARD_MAX_RETRIES', '3'))
        self.retry_delay = float(os.environ.get('TRUSTGUARD_RETRY_DELAY_SECONDS', '0.5'))
        
        # Configurar sessão HTTP com retry automático
        self.session = requests.Session()
        
        logger.info(
            f"TrustGuard Service inicializado com API URL: {self.api_base_url}, "
            f"timeout: {self.timeout}s, max retries: {self.max_retries}"
        )
    
    @backoff.on_exception(
        backoff.expo,
        (RequestException, ConnectionError),
        max_tries=3,
        factor=2
    )
    def _make_api_request(
        self, 
        ctx: Dict[str, Any], 
        method: str, 
        endpoint: str, 
        payload: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Realiza uma requisição para a API do TrustGuard com retry automático.
        
        Args:
            ctx: Contexto da requisição com informações de tracing
            method: Método HTTP (GET, POST, etc)
            endpoint: Endpoint da API sem a URL base
            payload: Dados a serem enviados (para POST, PUT)
            
        Returns:
            Resposta da API em formato JSON
            
        Raises:
            Exception: Se houver falha na comunicação ou processamento
        """
        url = f"{self.api_base_url}/{endpoint.lstrip('/')}"
        
        # Preparar headers
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
            "Accept": "application/json",
            "X-Client-ID": "INNOVABIZ-IAM"
        }
        
        # Adicionar headers de rastreamento se disponíveis no contexto
        if ctx and 'trace_id' in ctx:
            headers["X-Trace-ID"] = ctx['trace_id']
        if ctx and 'request_id' in ctx:
            headers["X-Request-ID"] = ctx['request_id']
        
        start_time = time.time()
        try:
            if method.upper() == 'GET':
                response = self.session.get(
                    url, 
                    headers=headers, 
                    timeout=self.timeout
                )
            elif method.upper() == 'POST':
                response = self.session.post(
                    url, 
                    headers=headers,
                    json=payload,
                    timeout=self.timeout
                )
            elif method.upper() == 'PUT':
                response = self.session.put(
                    url, 
                    headers=headers,
                    json=payload,
                    timeout=self.timeout
                )
            else:
                raise ValueError(f"Método HTTP não suportado: {method}")
            
            # Registrar duração da requisição
            duration = time.time() - start_time
            logger.debug(f"TrustGuard API request: {method} {endpoint} - {response.status_code} ({duration:.2f}s)")
            
            # Verificar resposta
            response.raise_for_status()
            
            # Retornar dados JSON
            return response.json()
            
        except RequestException as e:
            duration = time.time() - start_time
            logger.error(
                f"Erro na requisição TrustGuard: {method} {endpoint} falhou após {duration:.2f}s: {str(e)}"
            )
            raise
    
    def verify_document_identity(
        self, 
        ctx: Dict[str, Any], 
        user_id: str, 
        document_data: Dict[str, Any]
    ) -> VerificationResponse:
        """
        Verifica a identidade de um usuário através de documentos.
        
        Args:
            ctx: Contexto da requisição
            user_id: ID do usuário
            document_data: Dados do documento
            
        Returns:
            Resposta da verificação
        """
        logger.info(f"Iniciando verificação de documento para usuário {user_id}")
        
        # Preparar payload da requisição
        payload = {
            "userId": user_id,
            "verificationType": "DOCUMENT",
            "documentData": document_data,
            "metadata": {
                "source": "IAM",
                "module": "BureauCredito",
                "timestamp": datetime.utcnow().isoformat(),
                "environment": os.environ.get("ENVIRONMENT", "development")
            }
        }
        
        # Fazer requisição à API
        try:
            result = self._make_api_request(ctx, "POST", "/api/v1/verify", payload)
            
            # Converter para modelo de domínio
            details = None
            if "details" in result:
                details = VerificationDetails(
                    document_validity=result["details"].get("documentValidity"),
                    biometric_match=result["details"].get("biometricMatch"),
                    age_verification=result["details"].get("ageVerification"),
                    address_verification=result["details"].get("addressVerification"),
                    watch_list_status=result["details"].get("watchListStatus"),
                    fraud_signals=result["details"].get("fraudSignals"),
                    additional_checks=result["details"].get("additionalChecks")
                )
            
            compliance_status = None
            if "complianceStatus" in result:
                compliance_status = ComplianceStatus(
                    pep_status=result["complianceStatus"].get("pepStatus"),
                    sanctions_hit=result["complianceStatus"].get("sanctionsHit"),
                    aml_status=result["complianceStatus"].get("amlStatus"),
                    risk_level=result["complianceStatus"].get("riskLevel"),
                    jurisdiction_checks=result["complianceStatus"].get("jurisdictionChecks"),
                    regulatory_notes=result["complianceStatus"].get("regulatoryNotes")
                )
                
            timestamp = datetime.fromisoformat(result["timestamp"].replace("Z", "+00:00"))
            expires_at = None
            if "expiresAt" in result and result["expiresAt"]:
                expires_at = datetime.fromisoformat(result["expiresAt"].replace("Z", "+00:00"))
            
            response = VerificationResponse(
                verification_id=result["verificationId"],
                status=result["status"],
                score=result["score"],
                confidence=result["confidence"],
                timestamp=timestamp,
                verification_type=result["verificationType"],
                warnings=result.get("warnings", []),
                details=details,
                compliance_status=compliance_status,
                expires_at=expires_at,
                recommended_action=result.get("recommendedAction", "")
            )
            
            logger.info(
                f"Verificação de documento para usuário {user_id} concluída com status {response.status}"
            )
            
            return response
            
        except Exception as e:
            logger.error(f"Falha na verificação de documento para usuário {user_id}: {str(e)}", exc_info=True)
            raise
    
    def verify_biometric_identity(
        self, 
        ctx: Dict[str, Any], 
        user_id: str, 
        biometric_data: Dict[str, Any]
    ) -> VerificationResponse:
        """
        Verifica a identidade de um usuário através de dados biométricos.
        
        Args:
            ctx: Contexto da requisição
            user_id: ID do usuário
            biometric_data: Dados biométricos
            
        Returns:
            Resposta da verificação
        """
        logger.info(f"Iniciando verificação biométrica para usuário {user_id}")
        
        # Preparar payload da requisição
        payload = {
            "userId": user_id,
            "verificationType": "BIOMETRIC",
            "biometricData": biometric_data,
            "metadata": {
                "source": "IAM",
                "module": "BureauCredito",
                "timestamp": datetime.utcnow().isoformat(),
                "environment": os.environ.get("ENVIRONMENT", "development")
            }
        }
        
        # Fazer requisição à API
        try:
            result = self._make_api_request(ctx, "POST", "/api/v1/verify", payload)
            
            # Converter para modelo de domínio
            details = None
            if "details" in result:
                details = VerificationDetails(
                    biometric_match=result["details"].get("biometricMatch"),
                    liveness_score=result["details"].get("livenessScore"),
                    watch_list_status=result["details"].get("watchListStatus"),
                    fraud_signals=result["details"].get("fraudSignals"),
                    additional_checks=result["details"].get("additionalChecks")
                )
            
            compliance_status = None
            if "complianceStatus" in result:
                compliance_status = ComplianceStatus(
                    pep_status=result["complianceStatus"].get("pepStatus"),
                    sanctions_hit=result["complianceStatus"].get("sanctionsHit"),
                    aml_status=result["complianceStatus"].get("amlStatus"),
                    risk_level=result["complianceStatus"].get("riskLevel"),
                    jurisdiction_checks=result["complianceStatus"].get("jurisdictionChecks"),
                    regulatory_notes=result["complianceStatus"].get("regulatoryNotes")
                )
                
            timestamp = datetime.fromisoformat(result["timestamp"].replace("Z", "+00:00"))
            expires_at = None
            if "expiresAt" in result and result["expiresAt"]:
                expires_at = datetime.fromisoformat(result["expiresAt"].replace("Z", "+00:00"))
            
            response = VerificationResponse(
                verification_id=result["verificationId"],
                status=result["status"],
                score=result["score"],
                confidence=result["confidence"],
                timestamp=timestamp,
                verification_type=result["verificationType"],
                warnings=result.get("warnings", []),
                details=details,
                compliance_status=compliance_status,
                expires_at=expires_at,
                recommended_action=result.get("recommendedAction", "")
            )
            
            logger.info(
                f"Verificação biométrica para usuário {user_id} concluída com status {response.status}"
            )
            
            return response
            
        except Exception as e:
            logger.error(f"Falha na verificação biométrica para usuário {user_id}: {str(e)}", exc_info=True)
            raise
    
    def get_verification_status(
        self, 
        ctx: Dict[str, Any], 
        verification_id: str
    ) -> VerificationResponse:
        """
        Obtém o status atual de uma verificação.
        
        Args:
            ctx: Contexto da requisição
            verification_id: ID da verificação
            
        Returns:
            Status atualizado da verificação
        """
        logger.info(f"Obtendo status da verificação {verification_id}")
        
        try:
            # Fazer requisição à API
            result = self._make_api_request(ctx, "GET", f"/api/v1/verification/{verification_id}")
            
            # Converter para modelo de domínio
            details = None
            if "details" in result:
                details = VerificationDetails(
                    document_validity=result["details"].get("documentValidity"),
                    biometric_match=result["details"].get("biometricMatch"),
                    liveness_score=result["details"].get("livenessScore"),
                    age_verification=result["details"].get("ageVerification"),
                    address_verification=result["details"].get("addressVerification"),
                    watch_list_status=result["details"].get("watchListStatus"),
                    fraud_signals=result["details"].get("fraudSignals"),
                    additional_checks=result["details"].get("additionalChecks")
                )
            
            compliance_status = None
            if "complianceStatus" in result:
                compliance_status = ComplianceStatus(
                    pep_status=result["complianceStatus"].get("pepStatus"),
                    sanctions_hit=result["complianceStatus"].get("sanctionsHit"),
                    aml_status=result["complianceStatus"].get("amlStatus"),
                    risk_level=result["complianceStatus"].get("riskLevel"),
                    jurisdiction_checks=result["complianceStatus"].get("jurisdictionChecks"),
                    regulatory_notes=result["complianceStatus"].get("regulatoryNotes")
                )
                
            timestamp = datetime.fromisoformat(result["timestamp"].replace("Z", "+00:00"))
            expires_at = None
            if "expiresAt" in result and result["expiresAt"]:
                expires_at = datetime.fromisoformat(result["expiresAt"].replace("Z", "+00:00"))
            
            response = VerificationResponse(
                verification_id=result["verificationId"],
                status=result["status"],
                score=result["score"],
                confidence=result["confidence"],
                timestamp=timestamp,
                verification_type=result["verificationType"],
                warnings=result.get("warnings", []),
                details=details,
                compliance_status=compliance_status,
                expires_at=expires_at,
                recommended_action=result.get("recommendedAction", "")
            )
            
            logger.info(
                f"Status da verificação {verification_id} obtido: {response.status}"
            )
            
            return response
            
        except Exception as e:
            logger.error(f"Falha ao obter status da verificação {verification_id}: {str(e)}", exc_info=True)
            raise
    
    def get_user_verification_history(
        self, 
        ctx: Dict[str, Any], 
        user_id: str, 
        limit: int = 10, 
        offset: int = 0
    ) -> List[VerificationResponse]:
        """
        Obtém o histórico de verificações de um usuário.
        
        Args:
            ctx: Contexto da requisição
            user_id: ID do usuário
            limit: Número máximo de resultados
            offset: Deslocamento para paginação
            
        Returns:
            Lista de verificações do usuário
        """
        logger.info(f"Obtendo histórico de verificações para usuário {user_id} (limit={limit}, offset={offset})")
        
        try:
            # Fazer requisição à API
            result = self._make_api_request(
                ctx, 
                "GET", 
                f"/api/v1/verifications/user/{user_id}?limit={limit}&offset={offset}"
            )
            
            # Converter para modelo de domínio
            responses = []
            for item in result.get("items", []):
                details = None
                if "details" in item:
                    details = VerificationDetails(
                        document_validity=item["details"].get("documentValidity"),
                        biometric_match=item["details"].get("biometricMatch"),
                        liveness_score=item["details"].get("livenessScore"),
                        age_verification=item["details"].get("ageVerification"),
                        address_verification=item["details"].get("addressVerification"),
                        watch_list_status=item["details"].get("watchListStatus"),
                        fraud_signals=item["details"].get("fraudSignals"),
                        additional_checks=item["details"].get("additionalChecks")
                    )
                
                compliance_status = None
                if "complianceStatus" in item:
                    compliance_status = ComplianceStatus(
                        pep_status=item["complianceStatus"].get("pepStatus"),
                        sanctions_hit=item["complianceStatus"].get("sanctionsHit"),
                        aml_status=item["complianceStatus"].get("amlStatus"),
                        risk_level=item["complianceStatus"].get("riskLevel"),
                        jurisdiction_checks=item["complianceStatus"].get("jurisdictionChecks"),
                        regulatory_notes=item["complianceStatus"].get("regulatoryNotes")
                    )
                    
                timestamp = datetime.fromisoformat(item["timestamp"].replace("Z", "+00:00"))
                expires_at = None
                if "expiresAt" in item and item["expiresAt"]:
                    expires_at = datetime.fromisoformat(item["expiresAt"].replace("Z", "+00:00"))
                
                response = VerificationResponse(
                    verification_id=item["verificationId"],
                    status=item["status"],
                    score=item["score"],
                    confidence=item["confidence"],
                    timestamp=timestamp,
                    verification_type=item["verificationType"],
                    warnings=item.get("warnings", []),
                    details=details,
                    compliance_status=compliance_status,
                    expires_at=expires_at,
                    recommended_action=item.get("recommendedAction", "")
                )
                
                responses.append(response)
            
            logger.info(
                f"Histórico de verificações para usuário {user_id} obtido: {len(responses)} registros"
            )
            
            return responses
            
        except Exception as e:
            logger.error(f"Falha ao obter histórico de verificações para usuário {user_id}: {str(e)}", exc_info=True)
            raise