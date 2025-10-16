"""
INNOVABIZ - Processador de Autenticação WebAuthn/FIDO2
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de processador de autenticação baseado em
           protocolos WebAuthn/FIDO2 para autenticação sem senha
           e integração com Passkeys.
==================================================================
"""

import base64
import datetime
import json
import logging
import uuid
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass
from enum import Enum

# Configuração de logging
logger = logging.getLogger(__name__)

class CredentialType(Enum):
    """Tipos de credenciais WebAuthn/FIDO2"""
    PLATFORM = "platform"  # Credencial do dispositivo (ex: TouchID/FaceID)
    CROSS_PLATFORM = "cross-platform"  # Credencial externa (ex: chave física)
    PASSKEY = "passkey"  # Credencial sincronizável (Apple/Google/Microsoft)
    HYBRID = "hybrid"  # Credencial híbrida

@dataclass
class WebAuthnContext:
    """Contexto para validação de autenticação WebAuthn/FIDO2"""
    user_id: str
    session_id: str
    timestamp: datetime.datetime
    client_data_json: str
    authenticator_data: str
    signature: str
    credential_id: str
    credential_type: CredentialType
    rp_id: str  # Relying Party ID
    origin: str
    user_verification: bool
    extensions: Optional[Dict[str, Any]] = None
    device_info: Optional[Dict[str, Any]] = None

class WebAuthnResult:
    """Resultado da validação de autenticação WebAuthn/FIDO2"""
    def __init__(
        self,
        is_valid: bool,
        risk_level: str,
        anomalies: List[str],
        confidence: float,
        credential_properties: Dict[str, Any]
    ):
        self.is_valid = is_valid
        self.risk_level = risk_level  # "low", "medium", "high", "critical"
        self.anomalies = anomalies
        self.confidence = confidence  # 0.0 a 1.0
        self.credential_properties = credential_properties
        self.timestamp = datetime.datetime.now()
    
    def to_dict(self) -> Dict[str, Any]:
        """Converte o resultado para dicionário"""
        return {
            "is_valid": self.is_valid,
            "risk_level": self.risk_level,
            "confidence": self.confidence,
            "anomalies": self.anomalies,
            "credential_properties": self.credential_properties,
            "timestamp": self.timestamp.isoformat()
        }

class WebAuthnProcessor:
    """Processador para autenticação WebAuthn/FIDO2"""
    
    def __init__(self, config: Optional[Dict[str, Any]] = None):
        """
        Inicializa o processador WebAuthn/FIDO2.
        
        Args:
            config: Configuração opcional para o processador
        """
        self.config = config or {}
        self.name = "webauthn_processor"
        
        # Configurações padrão
        self.rp_id = self.config.get("rp_id", "innovabiz.com")
        self.origins = self.config.get("origins", ["https://innovabiz.com"])
        self.require_user_verification = self.config.get("require_user_verification", True)
        self.passkey_providers = self.config.get("passkey_providers", ["apple", "google", "microsoft"])
        
        logger.info(f"Processador WebAuthn/FIDO2 inicializado com configuração: {self.config}")
    
    def process(self, context: WebAuthnContext) -> WebAuthnResult:
        """
        Valida os dados de autenticação WebAuthn/FIDO2.
        
        Args:
            context: Contexto de autenticação WebAuthn/FIDO2
            
        Returns:
            Resultado da validação
        """
        logger.debug(f"Validando autenticação WebAuthn/FIDO2 para usuário {context.user_id}")
        
        # Verificar credencial
        credential_valid, credential_anomalies = self._validate_credential(context)
        
        # Verificar RP ID
        rp_valid = context.rp_id == self.rp_id
        if not rp_valid:
            credential_anomalies.append("invalid_rp_id")
        
        # Verificar origem
        origin_valid = context.origin in self.origins
        if not origin_valid:
            credential_anomalies.append("invalid_origin")
        
        # Verificar user verification se exigido
        uv_valid = True
        if self.require_user_verification and not context.user_verification:
            uv_valid = False
            credential_anomalies.append("missing_user_verification")
        
        # Verificar assinatura (em implementação real, isso seria feito com criptografia)
        signature_valid = self._validate_signature(
            context.client_data_json,
            context.authenticator_data,
            context.signature,
            context.credential_id
        )
        
        if not signature_valid:
            credential_anomalies.append("invalid_signature")
        
        # Verificação completa
        is_valid = credential_valid and rp_valid and origin_valid and uv_valid and signature_valid
        
        # Extrair propriedades da credencial
        credential_properties = self._extract_credential_properties(context)
        
        # Determinar nível de risco e confiança
        risk_level = self._determine_risk_level(is_valid, credential_anomalies, context)
        confidence = self._calculate_confidence(context)
        
        # Criar resultado
        result = WebAuthnResult(
            is_valid=is_valid,
            risk_level=risk_level,
            anomalies=credential_anomalies,
            confidence=confidence,
            credential_properties=credential_properties
        )
        
        logger.debug(f"Autenticação WebAuthn/FIDO2 concluída: {result.to_dict()}")
        return result
    
    def _validate_credential(self, context: WebAuthnContext) -> Tuple[bool, List[str]]:
        """
        Valida a credencial apresentada.
        
        Args:
            context: Contexto de autenticação WebAuthn/FIDO2
            
        Returns:
            Tupla de (validade, anomalias)
        """
        anomalies = []
        
        # Em uma implementação real, verificaríamos a credencial no banco de dados
        # e faríamos validações criptográficas completas
        
        # Simulação para o propósito desta implementação
        is_valid = True
        
        # Verificar tipo de credencial
        if context.credential_type == CredentialType.PLATFORM:
            # Verificações específicas para credenciais de plataforma
            if not context.device_info:
                anomalies.append("missing_device_info")
                is_valid = False
        
        elif context.credential_type == CredentialType.CROSS_PLATFORM:
            # Verificações específicas para chaves de segurança
            pass
        
        elif context.credential_type == CredentialType.PASSKEY:
            # Verificações específicas para passkeys
            passkey_provider = context.extensions.get("passkey_provider") if context.extensions else None
            if passkey_provider and passkey_provider not in self.passkey_providers:
                anomalies.append("unsupported_passkey_provider")
                is_valid = False
        
        # Verificar formato da credencial
        try:
            credential_id_bytes = base64.b64decode(context.credential_id)
            if len(credential_id_bytes) < 16:  # Tamanho mínimo arbitrário
                anomalies.append("invalid_credential_format")
                is_valid = False
        except Exception as e:
            logger.error(f"Erro ao decodificar credential_id: {e}")
            anomalies.append("credential_decoding_error")
            is_valid = False
        
        return is_valid, anomalies
    
    def _validate_signature(
        self,
        client_data_json: str,
        authenticator_data: str,
        signature: str,
        credential_id: str
    ) -> bool:
        """
        Valida a assinatura da autenticação.
        
        Args:
            client_data_json: Dados do cliente em JSON
            authenticator_data: Dados do autenticador
            signature: Assinatura a validar
            credential_id: ID da credencial
            
        Returns:
            True se a assinatura for válida
        """
        # Em uma implementação real, usaríamos criptografia para verificar a assinatura
        # com a chave pública da credencial
        
        # Simulação para o propósito desta implementação
        try:
            # Verificar se os dados básicos existem
            if not client_data_json or not authenticator_data or not signature:
                return False
            
            # Decodificar dados
            client_data = json.loads(client_data_json) if isinstance(client_data_json, str) else client_data_json
            
            # Verificações básicas
            if not isinstance(client_data, dict):
                return False
            
            if "type" not in client_data or client_data["type"] != "webauthn.get":
                return False
            
            if "challenge" not in client_data:
                return False
            
            # Em uma implementação real, verificaríamos a criptografia aqui
            return True
            
        except Exception as e:
            logger.error(f"Erro ao validar assinatura: {e}")
            return False
    
    def _extract_credential_properties(self, context: WebAuthnContext) -> Dict[str, Any]:
        """
        Extrai propriedades da credencial para análise.
        
        Args:
            context: Contexto de autenticação WebAuthn/FIDO2
            
        Returns:
            Propriedades da credencial
        """
        properties = {
            "credential_id": context.credential_id,
            "credential_type": context.credential_type.value,
            "user_verification": context.user_verification,
            "timestamp": context.timestamp.isoformat(),
        }
        
        # Adicionar extensões se disponíveis
        if context.extensions:
            properties["extensions"] = context.extensions
        
        # Adicionar informações de dispositivo se disponíveis
        if context.device_info:
            properties["device_info"] = {
                "platform": context.device_info.get("platform"),
                "browser": context.device_info.get("browser"),
                "is_mobile": context.device_info.get("is_mobile", False),
                "is_trusted": context.device_info.get("is_trusted", False)
            }
        
        return properties
    
    def _determine_risk_level(
        self,
        is_valid: bool,
        anomalies: List[str],
        context: WebAuthnContext
    ) -> str:
        """
        Determina o nível de risco da autenticação.
        
        Args:
            is_valid: Se a autenticação é válida
            anomalies: Anomalias detectadas
            context: Contexto completo de autenticação
            
        Returns:
            Nível de risco: "low", "medium", "high" ou "critical"
        """
        # Se a autenticação é inválida, o risco é crítico
        if not is_valid:
            return "critical"
        
        # Anomalias críticas
        critical_anomalies = ["invalid_signature", "invalid_credential_format", "credential_decoding_error"]
        if any(a in critical_anomalies for a in anomalies):
            return "critical"
        
        # Anomalias graves
        high_risk_anomalies = ["invalid_rp_id", "invalid_origin"]
        if any(a in high_risk_anomalies for a in anomalies):
            return "high"
        
        # Anomalias moderadas
        medium_risk_anomalies = ["missing_user_verification", "unsupported_passkey_provider"]
        if any(a in medium_risk_anomalies for a in anomalies):
            return "medium"
        
        # Autenticação forte com user verification
        if context.user_verification:
            return "low"
        
        # Autenticação sem user verification
        return "medium"
    
    def _calculate_confidence(self, context: WebAuthnContext) -> float:
        """
        Calcula o nível de confiança na autenticação.
        
        Args:
            context: Contexto de autenticação WebAuthn/FIDO2
            
        Returns:
            Nível de confiança (0.0 a 1.0)
        """
        # Fatores que afetam a confiança
        factors = []
        
        # Tipo de credencial
        if context.credential_type == CredentialType.PLATFORM:
            # Credenciais de plataforma (biométricas) têm alta confiança
            factors.append(0.9)
        elif context.credential_type == CredentialType.PASSKEY:
            # Passkeys também têm alta confiança
            factors.append(0.9)
        elif context.credential_type == CredentialType.CROSS_PLATFORM:
            # Chaves físicas têm confiança muito alta
            factors.append(1.0)
        else:
            factors.append(0.7)
        
        # User verification
        if context.user_verification:
            factors.append(1.0)
        else:
            factors.append(0.6)
        
        # Dispositivo confiável (se informação disponível)
        if context.device_info and context.device_info.get("is_trusted", False):
            factors.append(0.9)
        
        # Origem confiável
        if context.origin in self.origins:
            factors.append(1.0)
        else:
            factors.append(0.3)
        
        # Calcular média dos fatores
        return sum(factors) / len(factors) if factors else 0.5
    
    def register_credential(
        self,
        user_id: str,
        credential_id: str,
        public_key: str,
        credential_type: CredentialType,
        device_info: Optional[Dict[str, Any]] = None,
        extensions: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Registra uma nova credencial WebAuthn.
        
        Args:
            user_id: ID do usuário
            credential_id: ID da credencial
            public_key: Chave pública da credencial
            credential_type: Tipo da credencial
            device_info: Informações do dispositivo
            extensions: Extensões WebAuthn
            
        Returns:
            Dados da credencial registrada
        """
        # Em uma implementação real, salvaríamos estas informações no banco de dados
        
        credential = {
            "id": credential_id,
            "user_id": user_id,
            "public_key": public_key,
            "type": credential_type.value,
            "created_at": datetime.datetime.now().isoformat(),
            "last_used": None,
            "counter": 0
        }
        
        if device_info:
            credential["device_info"] = device_info
        
        if extensions:
            credential["extensions"] = extensions
        
        logger.info(f"Credencial registrada para usuário {user_id}: {credential_id}")
        return credential
