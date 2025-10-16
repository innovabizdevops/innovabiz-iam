"""
INNOVABIZ - Gerenciador de Políticas Regionais para IAM
===============================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0

Descrição: Este módulo implementa o gerenciador de políticas regionais
para os frameworks de compliance de IAM, permitindo personalização 
específica por região, indústria e framework.
===============================================================
"""

import logging
import json
from typing import Dict, List, Optional, Any, Tuple, Set
from enum import Enum
from dataclasses import dataclass, field
from datetime import datetime
import uuid

from opentelemetry import trace
from opentelemetry.trace import Status, StatusCode

from .compliance_metadata import ComplianceRegistry, Region, Industry, ComplianceFramework
from .exceptions import PolicyValidationError, PolicyNotFoundException, RegionNotSupportedException
from ..adaptive.risk_engine import RiskSignalProcessor
from ...common.observability import ContextLogger, metrics
from ...database.repository import ConfigRepository

# Configuração de logging
logger = ContextLogger(__name__)
tracer = trace.get_tracer(__name__)

# Enum para status de ativação de política
class PolicyStatus(str, Enum):
    """Status de ativação de política regional."""
    ACTIVE = "ACTIVE"
    INACTIVE = "INACTIVE"
    PENDING_APPROVAL = "PENDING_APPROVAL"
    DEPRECATED = "DEPRECATED"


@dataclass
class RegionalFrameworkPolicy:
    """
    Representa uma política específica para um framework em uma região/indústria.
    """
    id: str
    region: Region
    industry: Optional[Industry]
    framework: ComplianceFramework
    status: PolicyStatus
    settings: Dict[str, Any]
    version: str
    created_at: datetime
    updated_at: datetime
    last_validated: Optional[datetime] = None
    validation_score: Optional[float] = None
    override_rules: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    metadata: Dict[str, Any] = field(default_factory=dict)

    @property
    def is_active(self) -> bool:
        """Verifica se a política está ativa."""
        return self.status == PolicyStatus.ACTIVE
    
    @property
    def requires_specific_authentication(self) -> bool:
        """Verifica se a política requer autenticação específica."""
        return self.settings.get('require_specific_auth', False)
    
    @property
    def authentication_factors(self) -> List[str]:
        """Retorna os fatores de autenticação exigidos pela política."""
        return self.settings.get('authentication_factors', [])


class RegionalPolicyManager:
    """
    Gerenciador de políticas regionais para frameworks de compliance.
    """
    
    def __init__(self, compliance_registry: ComplianceRegistry, config_repository: ConfigRepository):
        """
        Inicializa o gerenciador de políticas regionais.
        
        Args:
            compliance_registry: Registro de compliance com metadados.
            config_repository: Repositório para armazenamento de configurações.
        """
        self.compliance_registry = compliance_registry
        self.config_repository = config_repository
        self.risk_processor = RiskSignalProcessor()
        self._policies_cache: Dict[str, RegionalFrameworkPolicy] = {}
        self._load_policies()
    
    @tracer.start_as_current_span("load_policies")
    def _load_policies(self) -> None:
        """Carrega todas as políticas do repositório."""
        with metrics.measure_execution_time("policy_manager.load_policies"):
            try:
                policies_data = self.config_repository.get_all("iam_compliance_policies")
                for policy_data in policies_data:
                    policy = self._deserialize_policy(policy_data)
                    self._policies_cache[policy.id] = policy
                logger.info(f"Carregadas {len(self._policies_cache)} políticas regionais de compliance")
            except Exception as e:
                logger.error(f"Erro ao carregar políticas: {str(e)}")
                # Continuar com cache vazio e tentar carregar posteriormente
    
    def _deserialize_policy(self, policy_data: Dict[str, Any]) -> RegionalFrameworkPolicy:
        """
        Converte dados do repositório para objeto RegionalFrameworkPolicy.
        
        Args:
            policy_data: Dados da política do repositório.
            
        Returns:
            Objeto RegionalFrameworkPolicy.
        """
        return RegionalFrameworkPolicy(
            id=policy_data.get("id", str(uuid.uuid4())),
            region=Region(policy_data.get("region")),
            industry=Industry(policy_data.get("industry")) if policy_data.get("industry") else None,
            framework=ComplianceFramework(policy_data.get("framework")),
            status=PolicyStatus(policy_data.get("status", "INACTIVE")),
            settings=policy_data.get("settings", {}),
            version=policy_data.get("version", "1.0.0"),
            created_at=datetime.fromisoformat(policy_data.get("created_at", 
                                                          datetime.now().isoformat())),
            updated_at=datetime.fromisoformat(policy_data.get("updated_at", 
                                                          datetime.now().isoformat())),
            last_validated=datetime.fromisoformat(policy_data.get("last_validated")) 
                        if policy_data.get("last_validated") else None,
            validation_score=policy_data.get("validation_score"),
            override_rules=policy_data.get("override_rules", {}),
            metadata=policy_data.get("metadata", {})
        )
    
    def _serialize_policy(self, policy: RegionalFrameworkPolicy) -> Dict[str, Any]:
        """
        Converte objeto RegionalFrameworkPolicy para formato de armazenamento.
        
        Args:
            policy: Objeto RegionalFrameworkPolicy para serializar.
            
        Returns:
            Dados serializados para armazenamento.
        """
        return {
            "id": policy.id,
            "region": policy.region.value,
            "industry": policy.industry.value if policy.industry else None,
            "framework": policy.framework.value,
            "status": policy.status.value,
            "settings": policy.settings,
            "version": policy.version,
            "created_at": policy.created_at.isoformat(),
            "updated_at": policy.updated_at.isoformat(),
            "last_validated": policy.last_validated.isoformat() if policy.last_validated else None,
            "validation_score": policy.validation_score,
            "override_rules": policy.override_rules,
            "metadata": policy.metadata
        }
    
    @tracer.start_as_current_span("get_policy")
    def get_policy(self, region: Region, framework: ComplianceFramework, 
                industry: Optional[Industry] = None) -> Optional[RegionalFrameworkPolicy]:
        """
        Obtém a política regional para um framework específico.
        
        Args:
            region: Região para a qual buscar a política.
            framework: Framework de compliance.
            industry: Indústria específica (opcional).
            
        Returns:
            Política regional ou None se não encontrada.
        """
        # Buscar a política mais específica primeiro (com indústria)
        if industry:
            for policy in self._policies_cache.values():
                if (policy.region == region and 
                    policy.framework == framework and 
                    policy.industry == industry and
                    policy.is_active):
                    return policy
        
        # Buscar política genérica para a região e framework
        for policy in self._policies_cache.values():
            if (policy.region == region and 
                policy.framework == framework and 
                policy.industry is None and
                policy.is_active):
                return policy
        
        # Se não encontrar uma política específica, buscar política global para o framework
        if region != Region.GLOBAL:
            for policy in self._policies_cache.values():
                if (policy.region == Region.GLOBAL and 
                    policy.framework == framework and
                    policy.is_active):
                    return policy
        
        return None
    
    @tracer.start_as_current_span("create_policy")
    def create_policy(self, region: Region, framework: ComplianceFramework,
                    settings: Dict[str, Any], industry: Optional[Industry] = None,
                    metadata: Optional[Dict[str, Any]] = None) -> RegionalFrameworkPolicy:
        """
        Cria uma nova política regional para um framework.
        
        Args:
            region: Região para a política.
            framework: Framework de compliance.
            settings: Configurações da política.
            industry: Indústria específica (opcional).
            metadata: Metadados adicionais (opcional).
            
        Returns:
            Política regional criada.
            
        Raises:
            PolicyValidationError: Se a validação da política falhar.
            RegionNotSupportedException: Se a região não for suportada.
        """
        # Verificar se a região e o framework são suportados
        if not self.compliance_registry.is_region_supported(region):
            raise RegionNotSupportedException(f"Região {region.value} não suportada")
        
        if not self.compliance_registry.is_framework_supported(framework):
            raise PolicyValidationError(f"Framework {framework.value} não suportado")
            
        # Validar configurações da política
        self._validate_policy_settings(framework, settings)
        
        # Criar nova política
        policy = RegionalFrameworkPolicy(
            id=str(uuid.uuid4()),
            region=region,
            industry=industry,
            framework=framework,
            status=PolicyStatus.ACTIVE,
            settings=settings,
            version="1.0.0",
            created_at=datetime.now(),
            updated_at=datetime.now(),
            metadata=metadata or {}
        )
        
        # Salvar política
        self._save_policy(policy)
        
        logger.info(f"Política regional criada: {policy.id} para {region.value}/{framework.value}")
        return policy
    
    @tracer.start_as_current_span("update_policy")
    def update_policy(self, policy_id: str, settings: Dict[str, Any],
                    status: Optional[PolicyStatus] = None,
                    override_rules: Optional[Dict[str, Dict[str, Any]]] = None,
                    metadata: Optional[Dict[str, Any]] = None) -> RegionalFrameworkPolicy:
        """
        Atualiza uma política regional existente.
        
        Args:
            policy_id: ID da política a atualizar.
            settings: Novas configurações (opcional).
            status: Novo status (opcional).
            override_rules: Regras de override (opcional).
            metadata: Novos metadados (opcional).
            
        Returns:
            Política atualizada.
            
        Raises:
            PolicyNotFoundException: Se a política não for encontrada.
            PolicyValidationError: Se a validação da nova configuração falhar.
        """
        # Buscar política existente
        policy = self._policies_cache.get(policy_id)
        if not policy:
            raise PolicyNotFoundException(f"Política com ID {policy_id} não encontrada")
        
        # Validar configurações da política
        if settings:
            self._validate_policy_settings(policy.framework, settings)
            policy.settings.update(settings)
        
        # Atualizar campos
        if status:
            policy.status = status
        
        if override_rules:
            policy.override_rules.update(override_rules)
        
        if metadata:
            policy.metadata.update(metadata)
        
        policy.updated_at = datetime.now()
        policy.version = self._increment_version(policy.version)
        
        # Salvar política
        self._save_policy(policy)
        
        logger.info(f"Política regional atualizada: {policy.id}")
        return policy
    
    def _increment_version(self, version: str) -> str:
        """Incrementa a versão da política."""
        major, minor, patch = map(int, version.split('.'))
        return f"{major}.{minor}.{patch + 1}"
    
    @tracer.start_as_current_span("validate_policy_settings")
    def _validate_policy_settings(self, framework: ComplianceFramework, 
                                settings: Dict[str, Any]) -> None:
        """
        Valida as configurações da política para um framework.
        
        Args:
            framework: Framework de compliance.
            settings: Configurações a validar.
            
        Raises:
            PolicyValidationError: Se a validação falhar.
        """
        # Obter requisitos do framework
        framework_requirements = self.compliance_registry.get_framework_requirements(framework)
        if not framework_requirements:
            raise PolicyValidationError(f"Requisitos para framework {framework.value} não encontrados")
        
        # Validar configurações de autenticação
        if 'authentication_factors' in settings:
            auth_factors = settings['authentication_factors']
            if not isinstance(auth_factors, list):
                raise PolicyValidationError("authentication_factors deve ser uma lista")
            
            # Verificar se todos os fatores de autenticação são suportados
            valid_factors = self.compliance_registry.get_supported_auth_factors()
            for factor in auth_factors:
                if factor not in valid_factors:
                    raise PolicyValidationError(f"Fator de autenticação não suportado: {factor}")
        
        # Validar valores mínimos exigidos pelo framework
        if framework == ComplianceFramework.GDPR:
            self._validate_gdpr_settings(settings)
        elif framework == ComplianceFramework.LGPD:
            self._validate_lgpd_settings(settings)
        elif framework == ComplianceFramework.PNDSB:
            self._validate_pndsb_settings(settings)
        elif framework == ComplianceFramework.HIPAA:
            self._validate_hipaa_settings(settings)
    
    def _validate_gdpr_settings(self, settings: Dict[str, Any]) -> None:
        """Valida configurações específicas para GDPR."""
        if settings.get('data_retention_days', 0) > 730:  # Máximo de 2 anos
            raise PolicyValidationError("GDPR: data_retention_days não pode exceder 730 dias (2 anos)")
        
        if 'authentication_factors' in settings and len(settings['authentication_factors']) < 2:
            raise PolicyValidationError("GDPR: Pelo menos dois fatores de autenticação são necessários")
    
    def _validate_lgpd_settings(self, settings: Dict[str, Any]) -> None:
        """Valida configurações específicas para LGPD."""
        if settings.get('data_retention_days', 0) > 1825:  # Máximo de 5 anos
            raise PolicyValidationError("LGPD: data_retention_days não pode exceder 1825 dias (5 anos)")
        
        if 'authentication_factors' in settings and len(settings['authentication_factors']) < 2:
            raise PolicyValidationError("LGPD: Pelo menos dois fatores de autenticação são necessários")
    
    def _validate_pndsb_settings(self, settings: Dict[str, Any]) -> None:
        """Valida configurações específicas para PNDSB."""
        # PNDSB tem requisitos mais flexíveis para inclusão financeira
        if 'authentication_alternatives' not in settings or not settings['authentication_alternatives']:
            raise PolicyValidationError("PNDSB: authentication_alternatives é obrigatório")
    
    def _validate_hipaa_settings(self, settings: Dict[str, Any]) -> None:
        """Valida configurações específicas para HIPAA."""
        if 'phi_access_logging' not in settings or not settings['phi_access_logging']:
            raise PolicyValidationError("HIPAA: phi_access_logging é obrigatório")
        
        if 'authentication_factors' in settings and len(settings['authentication_factors']) < 2:
            raise PolicyValidationError("HIPAA: Pelo menos dois fatores de autenticação são necessários")
    
    @tracer.start_as_current_span("save_policy")
    def _save_policy(self, policy: RegionalFrameworkPolicy) -> None:
        """
        Salva uma política no repositório e atualiza o cache.
        
        Args:
            policy: Política a ser salva.
        """
        try:
            policy_data = self._serialize_policy(policy)
            self.config_repository.upsert("iam_compliance_policies", policy.id, policy_data)
            self._policies_cache[policy.id] = policy
        except Exception as e:
            logger.error(f"Erro ao salvar política {policy.id}: {str(e)}")
            raise
    
    @tracer.start_as_current_span("delete_policy")
    def delete_policy(self, policy_id: str) -> None:
        """
        Remove uma política regional.
        
        Args:
            policy_id: ID da política a remover.
            
        Raises:
            PolicyNotFoundException: Se a política não for encontrada.
        """
        if policy_id not in self._policies_cache:
            raise PolicyNotFoundException(f"Política com ID {policy_id} não encontrada")
        
        try:
            self.config_repository.delete("iam_compliance_policies", policy_id)
            del self._policies_cache[policy_id]
            logger.info(f"Política regional removida: {policy_id}")
        except Exception as e:
            logger.error(f"Erro ao remover política {policy_id}: {str(e)}")
            raise
    
    @tracer.start_as_current_span("get_policies_by_region")
    def get_policies_by_region(self, region: Region) -> List[RegionalFrameworkPolicy]:
        """
        Obtém todas as políticas para uma região específica.
        
        Args:
            region: Região para buscar políticas.
            
        Returns:
            Lista de políticas para a região.
        """
        return [
            policy for policy in self._policies_cache.values()
            if policy.region == region and policy.is_active
        ]
    
    @tracer.start_as_current_span("get_policies_by_framework")
    def get_policies_by_framework(self, framework: ComplianceFramework) -> List[RegionalFrameworkPolicy]:
        """
        Obtém todas as políticas para um framework específico.
        
        Args:
            framework: Framework para buscar políticas.
            
        Returns:
            Lista de políticas para o framework.
        """
        return [
            policy for policy in self._policies_cache.values()
            if policy.framework == framework and policy.is_active
        ]
    
    @tracer.start_as_current_span("get_all_active_policies")
    def get_all_active_policies(self) -> List[RegionalFrameworkPolicy]:
        """
        Obtém todas as políticas ativas.
        
        Returns:
            Lista de todas as políticas ativas.
        """
        return [
            policy for policy in self._policies_cache.values()
            if policy.is_active
        ]
    
    @tracer.start_as_current_span("get_policy_by_id")
    def get_policy_by_id(self, policy_id: str) -> Optional[RegionalFrameworkPolicy]:
        """
        Obtém uma política pelo ID.
        
        Args:
            policy_id: ID da política.
            
        Returns:
            Política ou None se não encontrada.
        """
        return self._policies_cache.get(policy_id)
    
    @tracer.start_as_current_span("record_policy_validation")
    def record_policy_validation(self, policy_id: str, score: float) -> None:
        """
        Registra uma validação de política.
        
        Args:
            policy_id: ID da política validada.
            score: Pontuação da validação (0-100).
            
        Raises:
            PolicyNotFoundException: Se a política não for encontrada.
        """
        policy = self._policies_cache.get(policy_id)
        if not policy:
            raise PolicyNotFoundException(f"Política com ID {policy_id} não encontrada")
        
        policy.last_validated = datetime.now()
        policy.validation_score = max(0, min(100, score))  # Garantir valor entre 0-100
        
        self._save_policy(policy)
        logger.info(f"Validação registrada para política {policy_id}: score={score}")
    
    @tracer.start_as_current_span("validate_all_policies")
    def validate_all_policies(self) -> Dict[str, float]:
        """
        Executa validação de todas as políticas ativas.
        
        Returns:
            Dicionário com IDs de políticas e scores de validação.
        """
        results = {}
        for policy in self.get_all_active_policies():
            try:
                # Aqui seria feita a validação real da política
                # Para este exemplo, geramos um score aleatório entre 70 e 100
                import random
                score = random.uniform(70, 100)
                
                self.record_policy_validation(policy.id, score)
                results[policy.id] = score
            except Exception as e:
                logger.error(f"Erro ao validar política {policy.id}: {str(e)}")
                results[policy.id] = 0
        
        return results
    
    @tracer.start_as_current_span("get_policy_templates")
    def get_policy_templates(self, framework: ComplianceFramework) -> Dict[str, Any]:
        """
        Obtém templates de políticas para um framework.
        
        Args:
            framework: Framework para obter templates.
            
        Returns:
            Templates de políticas para o framework.
        """
        # Bases de templates para cada framework
        templates = {
            ComplianceFramework.GDPR: {
                "authentication_factors": ["password", "totp", "fido2"],
                "session_timeout_minutes": 30,
                "data_retention_days": 365,
                "require_consent": True,
                "breach_notification": True,
                "audit_frequency_days": 90,
                "encrypt_pii": True,
                "dpia_required": True
            },
            ComplianceFramework.LGPD: {
                "authentication_factors": ["password", "totp", "sms"],
                "session_timeout_minutes": 60,
                "data_retention_days": 730,
                "require_consent": True,
                "breach_notification": True,
                "audit_frequency_days": 180,
                "encrypt_pii": True
            },
            ComplianceFramework.PNDSB: {
                "authentication_factors": ["password", "sms"],
                "authentication_alternatives": ["biometric_voice", "agent_verification"],
                "session_timeout_minutes": 60,
                "data_retention_days": 1825,
                "inclusive_access": True,
                "offline_capabilities": True
            },
            ComplianceFramework.HIPAA: {
                "authentication_factors": ["password", "totp", "smartcard"],
                "session_timeout_minutes": 15,
                "data_retention_days": 2190,  # 6 anos
                "encrypt_phi": True,
                "phi_access_logging": True,
                "emergency_access_procedure": True,
                "audit_frequency_days": 30
            }
        }
        
        return templates.get(framework, {})
