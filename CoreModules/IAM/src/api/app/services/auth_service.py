import uuid
import jwt
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List
from sqlalchemy.orm import Session
from sqlalchemy import text
from passlib.context import CryptContext
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id

# Configuração para hashing de senhas
pwd_context = CryptContext(schemes=["argon2"], deprecated="auto")

class AuthService:
    """
    Serviço para autenticação e autorização
    """
    
    @staticmethod
    def verify_password(plain_password: str, hashed_password: str) -> bool:
        """
        Verifica se a senha em texto plano corresponde à senha hash
        """
        return pwd_context.verify(plain_password, hashed_password)
    
    @staticmethod
    def get_password_hash(password: str) -> str:
        """
        Gera o hash de uma senha
        """
        return pwd_context.hash(password)
    
    @staticmethod
    def create_access_token(data: Dict[str, Any], expires_delta: Optional[timedelta] = None) -> str:
        """
        Cria um token JWT de acesso
        """
        to_encode = data.copy()
        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            expire = datetime.utcnow() + timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
        
        to_encode.update({"exp": expire.timestamp()})
        encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
        return encoded_jwt
    
    @staticmethod
    def create_refresh_token(data: Dict[str, Any]) -> str:
        """
        Cria um token JWT de atualização
        """
        to_encode = data.copy()
        expire = datetime.utcnow() + timedelta(days=settings.REFRESH_TOKEN_EXPIRE_DAYS)
        to_encode.update({"exp": expire.timestamp()})
        encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
        return encoded_jwt
    
    @staticmethod
    def create_mfa_token(data: Dict[str, Any]) -> str:
        """
        Cria um token para MFA
        """
        to_encode = data.copy()
        expire = datetime.utcnow() + timedelta(minutes=settings.MFA_TOKEN_EXPIRE_MINUTES)
        to_encode.update({
            "exp": expire.timestamp(),
            "token_type": "mfa"
        })
        encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
        return encoded_jwt
    
    @staticmethod
    def decode_token(token: str) -> Dict[str, Any]:
        """
        Decodifica um token JWT
        """
        try:
            payload = jwt.decode(token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM])
            return payload
        except jwt.PyJWTError as e:
            logger.error(f"Error decoding token: {str(e)}")
            raise ValueError("Token inválido ou expirado")
    
    @staticmethod
    async def authenticate_user(db: Session, username: str, password: str, organization_code: str) -> Optional[Dict[str, Any]]:
        """
        Autentica um usuário por username/email e senha
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Chamada para função SQL que procura usuário por username ou email e valida a senha
            query = text("""
                SELECT * FROM iam.authenticate_user(
                    :tenant_id, :username, :organization_code
                )
            """)
            
            result = db.execute(
                query,
                {
                    "tenant_id": tenant_id,
                    "username": username,
                    "organization_code": organization_code
                }
            ).fetchone()
            
            if not result:
                return None
            
            # Convertendo resultado em dicionário
            user_data = dict(result._mapping)
            
            # Verificando a senha
            if not AuthService.verify_password(password, user_data["password_hash"]):
                return None
            
            # Removendo a senha hash do resultado
            user_data.pop("password_hash", None)
            
            return user_data
        except Exception as e:
            logger.error(f"Error authenticating user: {str(e)}")
            return None
    
    @staticmethod
    async def get_user_permissions(db: Session, user_id: uuid.UUID) -> List[str]:
        """
        Obtém todas as permissões de um usuário
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Chamada para função SQL que obtém as permissões do usuário
            query = text("""
                SELECT permission_code FROM iam.get_user_permissions(
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
            
            # Extraindo os códigos de permissão do resultado
            permissions = [row[0] for row in result]
            
            return permissions
        except Exception as e:
            logger.error(f"Error getting user permissions: {str(e)}")
            return []
    
    @staticmethod
    async def get_user_roles(db: Session, user_id: uuid.UUID) -> List[Dict[str, Any]]:
        """
        Obtém todos os papéis de um usuário
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Chamada para função SQL que obtém os papéis do usuário
            query = text("""
                SELECT * FROM iam.get_user_roles(
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
            roles = [dict(row._mapping) for row in result]
            
            return roles
        except Exception as e:
            logger.error(f"Error getting user roles: {str(e)}")
            return []
    
    @staticmethod
    async def check_permission(db: Session, user_id: uuid.UUID, resource_type: str, action: str, resource_id: Optional[str] = None) -> bool:
        """
        Verifica se um usuário tem permissão para realizar uma ação em um recurso
        """
        tenant_id = get_current_tenant_id()
        
        try:
            # Chamada para função SQL que verifica a permissão do usuário
            query = text("""
                SELECT * FROM iam.check_permission(
                    :tenant_id, :user_id, :resource_type, :action, :resource_id
                )
            """)
            
            result = db.execute(
                query,
                {
                    "tenant_id": tenant_id,
                    "user_id": str(user_id),
                    "resource_type": resource_type,
                    "action": action,
                    "resource_id": resource_id
                }
            ).scalar()
            
            return bool(result)
        except Exception as e:
            logger.error(f"Error checking permission: {str(e)}")
            return False
