from typing import List, Dict, Any, Optional
import os
from pydantic import BaseSettings

class Settings(BaseSettings):
    # Informações do Projeto
    PROJECT_NAME: str = "INNOVABIZ IAM API"
    PROJECT_DESCRIPTION: str = "API para gerenciamento de identidade e acesso da plataforma INNOVABIZ"
    API_VERSION: str = "1.0.0"
    AUTHOR: str = "Eduardo Jeremias"
    AUTHOR_EMAIL: str = "innovabizdevops@gmail.com"
    
    # Configurações do Servidor
    HOST: str = os.getenv("IAM_API_HOST", "0.0.0.0")
    PORT: int = int(os.getenv("IAM_API_PORT", "8080"))
    DEBUG: bool = os.getenv("IAM_API_DEBUG", "False").lower() == "true"
    PRODUCTION: bool = os.getenv("IAM_API_PRODUCTION", "False").lower() == "true"
    
    # Configurações de Segurança
    SECRET_KEY: str = os.getenv("IAM_API_SECRET_KEY", "CHANGE_THIS_IN_PRODUCTION_ENVIRONMENT_TO_SECURE_KEY")
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = int(os.getenv("IAM_API_ACCESS_TOKEN_EXPIRE_MINUTES", "30"))
    REFRESH_TOKEN_EXPIRE_DAYS: int = int(os.getenv("IAM_API_REFRESH_TOKEN_EXPIRE_DAYS", "7"))
    MFA_TOKEN_EXPIRE_MINUTES: int = int(os.getenv("IAM_API_MFA_TOKEN_EXPIRE_MINUTES", "5"))
    
    # Configurações de CORS
    CORS_ORIGINS: List[str] = os.getenv(
        "IAM_API_CORS_ORIGINS", 
        "http://localhost,http://localhost:3000,http://localhost:8080"
    ).split(",")
    
    # Configurações de Banco de Dados
    DATABASE_URL: str = os.getenv(
        "IAM_DATABASE_URL", 
        "postgresql://postgres:postgres@localhost:5432/innovabiz"
    )
    DATABASE_SCHEMA: str = os.getenv("IAM_DATABASE_SCHEMA", "iam")
    DATABASE_POOL_SIZE: int = int(os.getenv("IAM_DATABASE_POOL_SIZE", "5"))
    DATABASE_MAX_OVERFLOW: int = int(os.getenv("IAM_DATABASE_MAX_OVERFLOW", "10"))
    
    # Configurações de Redis para Cache e Sessões
    REDIS_URL: str = os.getenv("IAM_REDIS_URL", "redis://localhost:6379/0")
    REDIS_TTL: int = int(os.getenv("IAM_REDIS_TTL", "3600"))  # 1 hora em segundos
    
    # Configurações MFA
    MFA_TOTP_ISSUER: str = os.getenv("IAM_MFA_TOTP_ISSUER", "INNOVABIZ")
    MFA_BACKUP_CODES_COUNT: int = int(os.getenv("IAM_MFA_BACKUP_CODES_COUNT", "10"))
    MFA_BACKUP_CODES_LENGTH: int = int(os.getenv("IAM_MFA_BACKUP_CODES_LENGTH", "8"))
    MFA_MAX_ATTEMPTS: int = int(os.getenv("IAM_MFA_MAX_ATTEMPTS", "5"))
    
    # Configurações de AR/VR
    AR_VR_CONTINUOUS_AUTH_THRESHOLD: float = float(os.getenv("IAM_AR_VR_CONTINUOUS_AUTH_THRESHOLD", "0.7"))
    AR_VR_AUTH_REEVALUATION_INTERVAL_SECONDS: int = int(os.getenv("IAM_AR_VR_AUTH_REEVALUATION_INTERVAL_SECONDS", "30"))
    
    # Configurações de Federação
    IDP_METADATA_CACHE_TTL: int = int(os.getenv("IAM_IDP_METADATA_CACHE_TTL", "86400"))  # 24 horas em segundos
    
    # Configurações de Logging
    LOG_LEVEL: str = os.getenv("IAM_LOG_LEVEL", "INFO")
    LOG_FILE: Optional[str] = os.getenv("IAM_LOG_FILE")
    
    # Configurações de Auditoria
    AUDIT_ENABLED: bool = os.getenv("IAM_AUDIT_ENABLED", "True").lower() == "true"
    AUDIT_RETENTION_DAYS: int = int(os.getenv("IAM_AUDIT_RETENTION_DAYS", "365"))
    
    # Configurações de Email
    SMTP_SERVER: str = os.getenv("IAM_SMTP_SERVER", "smtp.example.com")
    SMTP_PORT: int = int(os.getenv("IAM_SMTP_PORT", "587"))
    SMTP_USERNAME: str = os.getenv("IAM_SMTP_USERNAME", "")
    SMTP_PASSWORD: str = os.getenv("IAM_SMTP_PASSWORD", "")
    SMTP_FROM_EMAIL: str = os.getenv("IAM_SMTP_FROM_EMAIL", "iam@innovabiz.com")
    
    # Configurações de SMS
    SMS_PROVIDER: str = os.getenv("IAM_SMS_PROVIDER", "twilio")
    SMS_PROVIDER_API_KEY: str = os.getenv("IAM_SMS_PROVIDER_API_KEY", "")
    SMS_PROVIDER_API_SECRET: str = os.getenv("IAM_SMS_PROVIDER_API_SECRET", "")
    SMS_SENDER_ID: str = os.getenv("IAM_SMS_SENDER_ID", "INNOVABIZ")
    
    # Configurações de Compliance para Healthcare
    HEALTHCARE_COMPLIANCE_CHECK_INTERVAL_HOURS: int = int(os.getenv("IAM_HEALTHCARE_COMPLIANCE_CHECK_INTERVAL_HOURS", "24"))
    
    # Configurações Multi-Tenant
    DEFAULT_TENANT_ID: str = os.getenv("IAM_DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000000")
    
    class Config:
        env_file = ".env"
        case_sensitive = True

settings = Settings()
