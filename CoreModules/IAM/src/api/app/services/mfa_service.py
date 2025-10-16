import uuid
import pyotp
import base64
import os
import random
import string
from typing import Optional, Dict, Any, List, Tuple
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from sqlalchemy import text
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id

class MFAService:
    """
    Serviço para gerenciamento de autenticação multi-fator (MFA)
    """
    
    @staticmethod
    def generate_totp_secret() -> str:
        """
        Gera um segredo para TOTP (Time-based One-Time Password)
        """
        return pyotp.random_base32()
    
    @staticmethod
    def generate_backup_codes(count: int = None, length: int = None) -> List[str]:
        """
        Gera códigos de backup para MFA
        """
        if count is None:
            count = settings.MFA_BACKUP_CODES_COUNT
        if length is None:
            length = settings.MFA_BACKUP_CODES_LENGTH
            
        codes = []
        characters = string.ascii_uppercase + string.digits
        # Remove caracteres ambíguos
        characters = characters.replace('0', '').replace('O', '').replace('1', '').replace('I', '')
        
        for _ in range(count):
            code = ''.join(random.choice(characters) for _ in range(length))
            # Formata como XXXX-XXXX para facilitar a leitura
            if length >= 8:
                code = code[:4] + '-' + code[4:8]
            codes.append(code)
            
        return codes
    
    @staticmethod
    def verify_totp(secret: str, code: str) -> bool:
        """
        Verifica um código TOTP
        """
        totp = pyotp.TOTP(secret)
        return totp.verify(code)
    
    @staticmethod
    def get_totp_provisioning_uri(secret: str, user_email: str) -> str:
        """
        Gera a URI para provisionar um TOTP em aplicativos como Google Authenticator
        """
        totp = pyotp.TOTP(secret)
        return totp.provisioning_uri(name=user_email, issuer_name=settings.MFA_TOTP_ISSUER)
    
    @staticmethod
    def generate_qr_code_url(secret: str, user_email: str) -> str:
        """
        Gera uma URL para um QR code de TOTP
        """
        provisioning_uri = MFAService.get_totp_provisioning_uri(secret, user_email)
        return f"https://chart.googleapis.com/chart?chs=200x200&chld=M|0&cht=qr&chl={provisioning_uri}"
    
    @staticmethod
    async def register_mfa_method(db: Session, user_id: uuid.UUID, method_type: str, name: str, metadata: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Registra um novo método MFA para um usuário
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Preparando metadados específicos do método
            method_data = {}
            setup_details = {}
            
            # Configuração específica para cada tipo de método
            if method_type == "totp":
                secret = MFAService.generate_totp_secret()
                
                # Obtendo email do usuário para o QR code
                user_email = db.execute(
                    text("SELECT email FROM iam.users WHERE id = :user_id"),
                    {"user_id": str(user_id)}
                ).scalar()
                
                method_data = {
                    "secret": secret,
                    "algorithm": "SHA1",
                    "digits": 6,
                    "period": 30
                }
                
                setup_details = {
                    "secret": secret,
                    "qr_code_url": MFAService.generate_qr_code_url(secret, user_email),
                    "manual_entry_key": secret
                }
            
            elif method_type == "backup_codes":
                codes = MFAService.generate_backup_codes()
                
                # Armazenamos o hash dos códigos, não os códigos em si
                from app.services.auth_service import AuthService
                hashed_codes = [AuthService.get_password_hash(code) for code in codes]
                
                method_data = {
                    "codes": hashed_codes,
                    "count": len(codes),
                    "used": [False] * len(codes)
                }
                
                setup_details = {
                    "codes": codes
                }
            
            elif method_type == "sms":
                # Verificar se o metadata contém o número de telefone
                if not metadata or "phone_number" not in metadata:
                    raise ValueError("Número de telefone obrigatório para método SMS")
                
                method_data = {
                    "phone_number": metadata["phone_number"]
                }
                
                setup_details = {
                    "phone_number": metadata["phone_number"]
                }
            
            elif method_type == "email":
                # Verificar se o metadata contém o email
                if not metadata or "email" not in metadata:
                    # Usar o email do usuário se não especificado
                    user_email = db.execute(
                        text("SELECT email FROM iam.users WHERE id = :user_id"),
                        {"user_id": str(user_id)}
                    ).scalar()
                    
                    method_data = {
                        "email": user_email
                    }
                    
                    setup_details = {
                        "email": user_email
                    }
                else:
                    method_data = {
                        "email": metadata["email"]
                    }
                    
                    setup_details = {
                        "email": metadata["email"]
                    }
            
            elif method_type in ["ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password"]:
                # Verificar se o metadata contém os dados do método
                if not metadata or "method_data" not in metadata:
                    raise ValueError(f"Dados do método obrigatórios para {method_type}")
                
                method_data = {
                    "ar_data": metadata["method_data"],
                    "complexity_score": metadata.get("complexity_score", 50)
                }
                
                setup_details = {
                    "type": method_type.replace("ar_", "").replace("_", " ").title(),
                    "complexity_score": metadata.get("complexity_score", 50)
                }
            
            # Chamada para função SQL que registra o método MFA
            query = text("""
                SELECT * FROM iam.register_mfa_method(
                    :tenant_id, :user_id, :method_type, :name, 
                    :method_data::jsonb, :metadata::jsonb
                )
            """)
            
            full_metadata = metadata or {}
            full_metadata.update({"setup_details": setup_details})
            
            result = db.execute(
                query,
                {
                    "tenant_id": tenant_id,
                    "user_id": str(user_id),
                    "method_type": method_type,
                    "name": name,
                    "method_data": dict(method_data),
                    "metadata": dict(full_metadata)
                }
            ).fetchone()
            
            db.commit()
            
            # Convertendo resultado em dicionário
            mfa_method = dict(result._mapping)
            
            # Incluindo detalhes de configuração na resposta
            mfa_method["setup_details"] = setup_details
            
            return mfa_method
        except Exception as e:
            db.rollback()
            logger.error(f"Error registering MFA method: {str(e)}")
            raise
    
    @staticmethod
    async def verify_mfa(db: Session, user_id: uuid.UUID, method_id: uuid.UUID, verification_code: str) -> bool:
        """
        Verifica um código MFA
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Obtendo o método MFA
            query = text("""
                SELECT * FROM iam.get_mfa_method(
                    :tenant_id, :user_id, :method_id
                )
            """)
            
            result = db.execute(
                query,
                {
                    "tenant_id": tenant_id,
                    "user_id": str(user_id),
                    "method_id": str(method_id)
                }
            ).fetchone()
            
            if not result:
                return False
            
            # Convertendo resultado em dicionário
            mfa_method = dict(result._mapping)
            
            # Verificando o código de acordo com o tipo de método
            method_type = mfa_method["method_type"]
            method_data = mfa_method["method_data"]
            
            if method_type == "totp":
                secret = method_data["secret"]
                return MFAService.verify_totp(secret, verification_code)
            
            elif method_type == "backup_codes":
                # Verifica se o código está na lista de códigos de backup
                from app.services.auth_service import AuthService
                
                for i, hashed_code in enumerate(method_data["codes"]):
                    # Se o código já foi usado, pula
                    if method_data["used"][i]:
                        continue
                    
                    # Verifica se o código corresponde
                    if AuthService.verify_password(verification_code.replace("-", ""), hashed_code):
                        # Atualiza o código como usado
                        method_data["used"][i] = True
                        
                        # Atualiza o método MFA no banco de dados
                        update_query = text("""
                            UPDATE iam.user_mfa_methods
                            SET method_data = :method_data::jsonb
                            WHERE id = :method_id
                        """)
                        
                        db.execute(
                            update_query,
                            {
                                "method_id": str(method_id),
                                "method_data": method_data
                            }
                        )
                        
                        db.commit()
                        return True
                
                return False
            
            elif method_type in ["sms", "email"]:
                # Para SMS e Email, normalmente verificamos se o código enviado
                # corresponde ao que foi gerado e armazenado temporariamente
                
                # Verificar na tabela iam.mfa_verification_codes
                verify_query = text("""
                    SELECT * FROM iam.mfa_verification_codes
                    WHERE user_id = :user_id
                    AND method_id = :method_id
                    AND code = :code
                    AND expires_at > NOW()
                    AND is_used = FALSE
                """)
                
                verify_result = db.execute(
                    verify_query,
                    {
                        "user_id": str(user_id),
                        "method_id": str(method_id),
                        "code": verification_code
                    }
                ).fetchone()
                
                if verify_result:
                    # Marcar o código como usado
                    update_query = text("""
                        UPDATE iam.mfa_verification_codes
                        SET is_used = TRUE
                        WHERE id = :code_id
                    """)
                    
                    db.execute(
                        update_query,
                        {
                            "code_id": str(verify_result["id"])
                        }
                    )
                    
                    db.commit()
                    return True
                
                return False
            
            elif method_type in ["ar_spatial_gesture", "ar_gaze_pattern", "ar_spatial_password"]:
                # Para métodos AR/VR, normalmente verificamos se o gesto, padrão de olhar
                # ou senha espacial corresponde ao que foi registrado
                
                # Este é um mock para demostração - em uma implementação real,
                # precisaríamos de algoritmos específicos para cada tipo de método
                
                # Verificando se o código tem o formato esperado para métodos AR/VR
                if verification_code == "ar_test_code":
                    return True
                
                return False
            
            # Método não suportado
            return False
        except Exception as e:
            logger.error(f"Error verifying MFA: {str(e)}")
            return False
    
    @staticmethod
    async def list_user_mfa_methods(db: Session, user_id: uuid.UUID) -> List[Dict[str, Any]]:
        """
        Lista todos os métodos MFA de um usuário
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Chamada para função SQL que lista os métodos MFA do usuário
            query = text("""
                SELECT * FROM iam.list_user_mfa_methods(
                    :tenant_id, :user_id
                )
            """)
            
            result = db.execute(
                query,
                {
                    "tenant_id": tenant_id,
                    "user_id": str(user_id)
                }
            ).fetchall()
            
            # Convertendo resultado em lista de dicionários
            mfa_methods = [dict(row._mapping) for row in result]
            
            # Removendo informações sensíveis
            for method in mfa_methods:
                if "method_data" in method and method["method_data"]:
                    if "secret" in method["method_data"]:
                        method["method_data"]["secret"] = "********"
                    if "codes" in method["method_data"]:
                        method["method_data"]["codes"] = ["********" for _ in method["method_data"]["codes"]]
            
            return mfa_methods
        except Exception as e:
            logger.error(f"Error listing user MFA methods: {str(e)}")
            return []
    
    @staticmethod
    async def remove_mfa_method(db: Session, user_id: uuid.UUID, method_id: uuid.UUID) -> bool:
        """
        Remove um método MFA de um usuário
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Chamada para função SQL que remove o método MFA
            query = text("""
                SELECT * FROM iam.remove_mfa_method(
                    :tenant_id, :user_id, :method_id
                )
            """)
            
            result = db.execute(
                query,
                {
                    "tenant_id": tenant_id,
                    "user_id": str(user_id),
                    "method_id": str(method_id)
                }
            ).scalar()
            
            db.commit()
            
            return result
        except Exception as e:
            db.rollback()
            logger.error(f"Error removing MFA method: {str(e)}")
            return False
