import json
import logging
from typing import Dict, Any, Optional

import graphene
from graphql import GraphQLError
from graphene import ObjectType, Field, String, Float, List, Boolean, ID, JSONString

from .common.auth import require_permission, get_current_user
from .common.context import get_trace_context
from ..app.trust_guard_service import TrustGuardService
from ..app.trust_guard_models import (
    VerificationRequest, 
    VerificationResponse, 
    VerificationStatus,
    VerificationScoreLevel
)

logger = logging.getLogger(__name__)

class DocumentDataInput(graphene.InputObjectType):
    document_type = String(required=True, description="Tipo do documento (ID, PASSPORT, DRIVER_LICENSE, etc)")
    document_number = String(required=True, description="Número do documento")
    issue_date = String(description="Data de emissão do documento (formato ISO)")
    expiry_date = String(description="Data de validade do documento (formato ISO)")
    issuing_country = String(description="País emissor do documento")
    issuing_authority = String(description="Autoridade emissora do documento")
    document_images = List(String, description="URLs das imagens do documento")
    additional_fields = JSONString(description="Campos adicionais específicos do tipo de documento")

class BiometricDataInput(graphene.InputObjectType):
    face_image = String(description="URL da imagem do rosto")
    liveness_data = JSONString(description="Dados de prova de vida")
    fingerprint_data = JSONString(description="Dados de impressão digital")
    voice_print = String(description="Dados de impressão de voz")
    iris_scan = String(description="Dados de escaneamento de íris")
    additional_biometrics = JSONString(description="Dados biométricos adicionais")

class VerificationDetailsType(ObjectType):
    document_validity = Boolean(description="Validade do documento")
    biometric_match = Float(description="Percentual de correspondência biométrica")
    liveness_score = Float(description="Pontuação de prova de vida")
    age_verification = Boolean(description="Verificação de idade")
    address_verification = Boolean(description="Verificação de endereço")
    watch_list_status = String(description="Status em listas de observação")
    fraud_signals = List(String, description="Sinais de fraude detectados")
    additional_checks = JSONString(description="Resultados de verificações adicionais")

class ComplianceStatusType(ObjectType):
    pep_status = Boolean(description="Status de Pessoa Politicamente Exposta")
    sanctions_hit = Boolean(description="Presença em listas de sanções")
    aml_status = String(description="Status de Anti-Money Laundering")
    risk_level = String(description="Nível de risco de compliance")
    jurisdiction_checks = List(String, description="Verificações jurisdicionais")
    regulatory_notes = List(String, description="Notas regulatórias")

class VerificationResponseType(ObjectType):
    verification_id = ID(description="ID único da verificação")
    status = String(description="Status da verificação (PENDING, APPROVED, REJECTED, REVIEW_NEEDED)")
    score = Float(description="Pontuação da verificação (0-100)")
    confidence = Float(description="Nível de confiança da verificação (0-100)")
    timestamp = String(description="Data e hora da verificação (ISO format)")
    expires_at = String(description="Data e hora de expiração da verificação (ISO format)")
    verification_type = String(description="Tipo de verificação realizada")
    warnings = List(String, description="Avisos gerados durante a verificação")
    details = Field(VerificationDetailsType, description="Detalhes da verificação")
    compliance_status = Field(ComplianceStatusType, description="Status de conformidade")
    recommended_action = String(description="Ação recomendada pelo sistema")

class TrustGuardMutations(ObjectType):
    verify_document_identity = Field(
        VerificationResponseType,
        user_id=String(required=True),
        document_data=DocumentDataInput(required=True),
        description="Verifica a identidade através de documentos via TrustGuard"
    )
    
    verify_biometric_identity = Field(
        VerificationResponseType,
        user_id=String(required=True),
        biometric_data=BiometricDataInput(required=True),
        description="Verifica a identidade através de biometria via TrustGuard"
    )
    
    @require_permission("iam:identity:verify")
    def resolve_verify_document_identity(self, info, user_id: str, document_data: Dict[str, Any]):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            current_user = get_current_user(info.context)
            
            logger.info(
                f"Iniciando verificação de documento para usuário {user_id} por {current_user['id']}"
            )
            
            service = TrustGuardService()
            
            # Converter input para formato do serviço
            document_dict = {
                "documentType": document_data.document_type,
                "documentNumber": document_data.document_number,
            }
            
            # Adicionar campos opcionais se fornecidos
            if hasattr(document_data, 'issue_date') and document_data.issue_date:
                document_dict["issueDate"] = document_data.issue_date
                
            if hasattr(document_data, 'expiry_date') and document_data.expiry_date:
                document_dict["expiryDate"] = document_data.expiry_date
                
            if hasattr(document_data, 'issuing_country') and document_data.issuing_country:
                document_dict["issuingCountry"] = document_data.issuing_country
                
            if hasattr(document_data, 'issuing_authority') and document_data.issuing_authority:
                document_dict["issuingAuthority"] = document_data.issuing_authority
                
            if hasattr(document_data, 'document_images') and document_data.document_images:
                document_dict["documentImages"] = document_data.document_images
                
            if hasattr(document_data, 'additional_fields') and document_data.additional_fields:
                additional = json.loads(document_data.additional_fields)
                for key, value in additional.items():
                    document_dict[key] = value
            
            # Executar verificação
            result = service.verify_document_identity(ctx, user_id, document_dict)
            
            # Converter resultado para o formato GraphQL
            response = {
                "verification_id": result.verification_id,
                "status": result.status,
                "score": result.score,
                "confidence": result.confidence,
                "timestamp": result.timestamp.isoformat(),
                "expires_at": result.expires_at.isoformat() if result.expires_at else None,
                "verification_type": result.verification_type,
                "warnings": result.warnings,
                "recommended_action": result.recommended_action,
            }
            
            # Adicionar detalhes se existirem
            if result.details:
                response["details"] = result.details
                
            # Adicionar status de compliance se existir
            if result.compliance_status:
                response["compliance_status"] = result.compliance_status
            
            logger.info(
                f"Verificação de documento concluída para usuário {user_id}, "
                f"status: {result.status}, score: {result.score}"
            )
            
            return response
        
        except Exception as e:
            logger.error(f"Erro na verificação de documento: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha na verificação de identidade por documento: {str(e)}")

    @require_permission("iam:identity:verify")
    def resolve_verify_biometric_identity(self, info, user_id: str, biometric_data: Dict[str, Any]):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            current_user = get_current_user(info.context)
            
            logger.info(
                f"Iniciando verificação biométrica para usuário {user_id} por {current_user['id']}"
            )
            
            service = TrustGuardService()
            
            # Converter input para formato do serviço
            biometric_dict = {}
            
            # Adicionar campos se fornecidos
            if hasattr(biometric_data, 'face_image') and biometric_data.face_image:
                biometric_dict["faceImage"] = biometric_data.face_image
                
            if hasattr(biometric_data, 'liveness_data') and biometric_data.liveness_data:
                biometric_dict["livenessData"] = json.loads(biometric_data.liveness_data)
                
            if hasattr(biometric_data, 'fingerprint_data') and biometric_data.fingerprint_data:
                biometric_dict["fingerprintData"] = json.loads(biometric_data.fingerprint_data)
                
            if hasattr(biometric_data, 'voice_print') and biometric_data.voice_print:
                biometric_dict["voicePrint"] = biometric_data.voice_print
                
            if hasattr(biometric_data, 'iris_scan') and biometric_data.iris_scan:
                biometric_dict["irisScan"] = biometric_data.iris_scan
                
            if hasattr(biometric_data, 'additional_biometrics') and biometric_data.additional_biometrics:
                additional = json.loads(biometric_data.additional_biometrics)
                for key, value in additional.items():
                    biometric_dict[key] = value
            
            # Executar verificação
            result = service.verify_biometric_identity(ctx, user_id, biometric_dict)
            
            # Converter resultado para o formato GraphQL
            response = {
                "verification_id": result.verification_id,
                "status": result.status,
                "score": result.score,
                "confidence": result.confidence,
                "timestamp": result.timestamp.isoformat(),
                "expires_at": result.expires_at.isoformat() if result.expires_at else None,
                "verification_type": result.verification_type,
                "warnings": result.warnings,
                "recommended_action": result.recommended_action,
            }
            
            # Adicionar detalhes se existirem
            if result.details:
                response["details"] = result.details
                
            # Adicionar status de compliance se existir
            if result.compliance_status:
                response["compliance_status"] = result.compliance_status
            
            logger.info(
                f"Verificação biométrica concluída para usuário {user_id}, "
                f"status: {result.status}, score: {result.score}"
            )
            
            return response
        
        except Exception as e:
            logger.error(f"Erro na verificação biométrica: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha na verificação de identidade biométrica: {str(e)}")

class TrustGuardQueries(ObjectType):
    verification_status = Field(
        VerificationResponseType,
        verification_id=String(required=True),
        description="Obtém o status de uma verificação de identidade"
    )
    
    verification_history = List(
        VerificationResponseType,
        user_id=String(required=True),
        limit=Int(default_value=10),
        offset=Int(default_value=0),
        description="Obtém o histórico de verificações para um usuário"
    )
    
    @require_permission("iam:identity:read")
    def resolve_verification_status(self, info, verification_id: str):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            
            service = TrustGuardService()
            result = service.get_verification_status(ctx, verification_id)
            
            # Converter resultado para o formato GraphQL
            response = {
                "verification_id": result.verification_id,
                "status": result.status,
                "score": result.score,
                "confidence": result.confidence,
                "timestamp": result.timestamp.isoformat(),
                "expires_at": result.expires_at.isoformat() if result.expires_at else None,
                "verification_type": result.verification_type,
                "warnings": result.warnings,
                "recommended_action": result.recommended_action,
            }
            
            # Adicionar detalhes se existirem
            if result.details:
                response["details"] = result.details
                
            # Adicionar status de compliance se existir
            if result.compliance_status:
                response["compliance_status"] = result.compliance_status
                
            return response
        
        except Exception as e:
            logger.error(f"Erro ao obter status de verificação: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha ao obter status de verificação: {str(e)}")
    
    @require_permission("iam:identity:read")
    def resolve_verification_history(self, info, user_id: str, limit: int, offset: int):
        try:
            # Obter contexto para tracing
            ctx = get_trace_context(info.context)
            
            service = TrustGuardService()
            results = service.get_user_verification_history(ctx, user_id, limit, offset)
            
            # Converter resultados para o formato GraphQL
            responses = []
            for result in results:
                response = {
                    "verification_id": result.verification_id,
                    "status": result.status,
                    "score": result.score,
                    "confidence": result.confidence,
                    "timestamp": result.timestamp.isoformat(),
                    "expires_at": result.expires_at.isoformat() if result.expires_at else None,
                    "verification_type": result.verification_type,
                    "warnings": result.warnings,
                    "recommended_action": result.recommended_action,
                }
                
                # Adicionar detalhes se existirem
                if result.details:
                    response["details"] = result.details
                    
                # Adicionar status de compliance se existir
                if result.compliance_status:
                    response["compliance_status"] = result.compliance_status
                    
                responses.append(response)
                
            return responses
        
        except Exception as e:
            logger.error(f"Erro ao obter histórico de verificações: {str(e)}", exc_info=True)
            raise GraphQLError(f"Falha ao obter histórico de verificações: {str(e)}")