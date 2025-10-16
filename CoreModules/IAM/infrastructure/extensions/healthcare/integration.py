"""
INNOVABIZ - Integração entre IAM e Healthcare para Validação e Conformidade
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Módulo de integração que conecta os sistemas de validação IAM
          com o módulo Healthcare, especificamente para conformidade com
          regulamentações do setor de saúde como HIPAA.
==================================================================
"""

import json
import logging
import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional, Union, Set, Tuple

from ...validation.iam_validator import ValidationReport, ValidationStatus
from ....extensions.healthcare.compliance.hipaa_validator import HIPAAHealthcareValidator
from ....extensions.healthcare.models.healthcare_models import HealthcareComplianceData, PolicySyncResult

logger = logging.getLogger(__name__)

class HealthcareComplianceService:
    """
    Serviço para integração de conformidade entre IAM e Healthcare.
    """
    
    def __init__(self, config_path: Optional[Path] = None):
        """
        Inicializa o serviço de integração.
        
        Args:
            config_path: Caminho opcional para configuração personalizada
        """
        self.config_path = config_path
        self.hipaa_validator = HIPAAHealthcareValidator()
        
        # Carregar configuração se fornecida
        self.config = {}
        if config_path and config_path.exists():
            with open(config_path, "r") as f:
                self.config = json.load(f)
    
    def get_compliance_data(self, hipaa_validation_id: str) -> Dict[str, Any]:
        """
        Recupera dados de conformidade HIPAA do módulo Healthcare com base em uma validação IAM.
        
        Args:
            hipaa_validation_id: ID da validação HIPAA do IAM
            
        Returns:
            Dados de conformidade do Healthcare relacionados à validação IAM
        """
        try:
            # Obter resultados da validação HIPAA do IAM
            iam_validation_data = self._get_iam_hipaa_validation(hipaa_validation_id)
            
            if not iam_validation_data:
                logger.error(f"Não foi possível encontrar dados de validação IAM para ID: {hipaa_validation_id}")
                return self._generate_default_compliance_data()
            
            # Recuperar dados do Healthcare relacionados à mesma área
            healthcare_data = self.hipaa_validator.validate_healthcare_hipaa_compliance(
                tenant_id=iam_validation_data.get("tenant_id", ""),
                iam_validation_id=hipaa_validation_id
            )
            
            # Calcular pontuação geral com base em IAM e Healthcare
            overall_score = self._calculate_combined_score(
                iam_score=iam_validation_data.get("score", 0),
                healthcare_score=healthcare_data.get("compliance_score", 0)
            )
            
            # Determinar status com base na pontuação
            status = "compliant" if overall_score >= 80 else "partially_compliant" if overall_score >= 60 else "non_compliant"
            
            # Combinar verificações específicas de saúde com dados de IAM
            return {
                "overallScore": overall_score,
                "status": status,
                "categoryScores": {
                    "privacy": healthcare_data.get("privacy_score", 0),
                    "security": iam_validation_data.get("security_score", 0),
                    "breach": healthcare_data.get("breach_notification_score", 0),
                    "patientRights": healthcare_data.get("patient_rights_score", 0)
                },
                "findings": self._combine_findings(
                    iam_validation_data.get("findings", []),
                    healthcare_data.get("findings", [])
                ),
                "policies": self._combine_policies(
                    iam_validation_data.get("policies", []),
                    healthcare_data.get("policies", [])
                ),
                "services": healthcare_data.get("healthcare_services", []),
                "timestamp": datetime.datetime.now().isoformat(),
                "validUntil": (datetime.datetime.now() + datetime.timedelta(days=90)).isoformat()
            }
            
        except Exception as e:
            logger.exception(f"Erro ao recuperar dados de conformidade HIPAA: {str(e)}")
            return self._generate_default_compliance_data()
    
    def _get_iam_hipaa_validation(self, validation_id: str) -> Dict[str, Any]:
        """
        Recupera os dados de validação HIPAA do IAM.
        
        Args:
            validation_id: ID da validação HIPAA
            
        Returns:
            Dados da validação IAM para HIPAA
        """
        # TODO: Implementar integração com o banco de dados para recuperar dados reais
        # Esta é uma implementação simulada para demonstração
        
        # Simulação de conexão com banco de dados
        from ....dataops.postgresql.db_connector import IAMValidationRepository
        
        try:
            repo = IAMValidationRepository()
            validation_data = repo.get_validation_by_id(validation_id)
            
            if not validation_data:
                return {}
                
            # Processar os dados brutos do banco para o formato adequado
            return {
                "tenant_id": validation_data.get("tenant_id"),
                "framework": "hipaa",
                "score": validation_data.get("score", 0),
                "security_score": validation_data.get("security_score", 0),
                "findings": validation_data.get("findings", []),
                "policies": validation_data.get("policies", []),
                "timestamp": validation_data.get("timestamp"),
                "validation_id": validation_id
            }
        except Exception as e:
            logger.error(f"Erro ao recuperar validação HIPAA: {str(e)}")
            return {}
    
    def _calculate_combined_score(self, iam_score: float, healthcare_score: float) -> float:
        """
        Calcula uma pontuação combinada com pesos para IAM e Healthcare.
        
        Args:
            iam_score: Pontuação da validação IAM
            healthcare_score: Pontuação da validação Healthcare
            
        Returns:
            Pontuação combinada
        """
        # Define pesos para cada componente (configurável)
        iam_weight = 0.6  # IAM tem maior peso pois é mais fundamental para segurança
        healthcare_weight = 0.4
        
        combined_score = (iam_score * iam_weight) + (healthcare_score * healthcare_weight)
        return round(combined_score, 1)
    
    def _combine_findings(self, iam_findings: List[Dict], healthcare_findings: List[Dict]) -> List[Dict]:
        """
        Combina descobertas de IAM e Healthcare, removendo duplicatas.
        
        Args:
            iam_findings: Descobertas da validação IAM
            healthcare_findings: Descobertas específicas de Healthcare
            
        Returns:
            Lista combinada de descobertas
        """
        # Mapear descobertas IAM por ID para evitar duplicatas
        findings_map = {f.get("id"): f for f in iam_findings}
        
        # Adicionar descobertas específicas de Healthcare
        for finding in healthcare_findings:
            finding_id = finding.get("id")
            if finding_id not in findings_map:
                findings_map[finding_id] = finding
            else:
                # Se a descoberta existir em ambos, use a severidade mais alta
                iam_severity = findings_map[finding_id].get("severity", "low")
                hc_severity = finding.get("severity", "low")
                
                severity_rank = {"low": 1, "medium": 2, "high": 3, "critical": 4}
                if severity_rank.get(hc_severity, 0) > severity_rank.get(iam_severity, 0):
                    findings_map[finding_id]["severity"] = hc_severity
                
                # Combine as recomendações
                findings_map[finding_id]["recommendation"] = f"{findings_map[finding_id].get('recommendation', '')} {finding.get('recommendation', '')}"
        
        return list(findings_map.values())
    
    def _combine_policies(self, iam_policies: List[Dict], healthcare_policies: List[Dict]) -> List[Dict]:
        """
        Combina políticas de IAM e Healthcare, identificando sobreposições.
        
        Args:
            iam_policies: Políticas do IAM
            healthcare_policies: Políticas específicas de Healthcare
            
        Returns:
            Lista combinada de políticas
        """
        # Mapear políticas por ID
        policies_map = {p.get("id"): p for p in iam_policies}
        
        # Adicionar políticas específicas de Healthcare
        for policy in healthcare_policies:
            policy_id = policy.get("id")
            if policy_id not in policies_map:
                policies_map[policy_id] = policy
            else:
                # Se a política existir em ambos, marque-a como compartilhada
                policies_map[policy_id]["sharedWithHealthcare"] = True
                
                # Combine as descrições se forem diferentes
                iam_desc = policies_map[policy_id].get("description", "")
                hc_desc = policy.get("description", "")
                if iam_desc != hc_desc:
                    policies_map[policy_id]["description"] = f"{iam_desc} [Healthcare: {hc_desc}]"
        
        return list(policies_map.values())
    
    def _generate_default_compliance_data(self) -> Dict[str, Any]:
        """
        Gera dados de conformidade padrão quando não é possível obter dados reais.
        
        Returns:
            Dados de conformidade padrão
        """
        return {
            "overallScore": 0,
            "status": "unknown",
            "categoryScores": {
                "privacy": 0,
                "security": 0,
                "breach": 0,
                "patientRights": 0
            },
            "findings": [],
            "policies": [],
            "services": [],
            "timestamp": datetime.datetime.now().isoformat(),
            "validUntil": (datetime.datetime.now() + datetime.timedelta(days=90)).isoformat()
        }
    
    def sync_policies(self, hipaa_validation_id: str, options: Dict[str, Any]) -> PolicySyncResult:
        """
        Sincroniza políticas entre IAM e o módulo Healthcare.
        
        Args:
            hipaa_validation_id: ID da validação HIPAA
            options: Opções de sincronização como syncAllPolicies e autoFix
            
        Returns:
            Resultado da sincronização
        """
        try:
            # Obter validação IAM e políticas de Healthcare
            iam_validation = self._get_iam_hipaa_validation(hipaa_validation_id)
            
            if not iam_validation:
                return PolicySyncResult(
                    syncedPoliciesCount=0,
                    successfulSyncs=0,
                    failedSyncs=0,
                    details=[]
                )
            
            # Obter políticas atuais do Healthcare
            healthcare_policies = self.hipaa_validator.get_healthcare_policies(
                tenant_id=iam_validation.get("tenant_id", "")
            )
            
            # Determinar quais políticas sincronizar com base nas opções
            policies_to_sync = []
            
            if options.get("syncAllPolicies", False):
                # Sincronizar todas as políticas HIPAA do IAM
                policies_to_sync = iam_validation.get("policies", [])
            else:
                # Sincronizar apenas políticas que afetam diretamente o Healthcare
                policies_to_sync = [
                    p for p in iam_validation.get("policies", [])
                    if self._is_healthcare_relevant_policy(p)
                ]
            
            # Executar a sincronização
            sync_details = []
            successful_syncs = 0
            failed_syncs = 0
            
            for policy in policies_to_sync:
                try:
                    # Verificar se a política já existe no Healthcare
                    exists = any(p.get("id") == policy.get("id") for p in healthcare_policies)
                    
                    # Aplicar a política ao Healthcare
                    result = self.hipaa_validator.apply_iam_policy_to_healthcare(
                        policy_id=policy.get("id"),
                        policy_data=policy,
                        auto_fix=options.get("autoFix", False)
                    )
                    
                    sync_details.append({
                        "policyId": policy.get("id"),
                        "policyName": policy.get("name"),
                        "status": "success" if result else "failed",
                        "message": f"{'Updated' if exists else 'Created'} policy in Healthcare" if result else "Failed to apply policy"
                    })
                    
                    if result:
                        successful_syncs += 1
                    else:
                        failed_syncs += 1
                        
                except Exception as e:
                    sync_details.append({
                        "policyId": policy.get("id"),
                        "policyName": policy.get("name"),
                        "status": "failed",
                        "message": f"Error: {str(e)}"
                    })
                    failed_syncs += 1
            
            # Retornar resultado da sincronização
            return PolicySyncResult(
                syncedPoliciesCount=len(policies_to_sync),
                successfulSyncs=successful_syncs,
                failedSyncs=failed_syncs,
                details=sync_details
            )
            
        except Exception as e:
            logger.exception(f"Erro ao sincronizar políticas: {str(e)}")
            return PolicySyncResult(
                syncedPoliciesCount=0,
                successfulSyncs=0,
                failedSyncs=1,
                details=[{
                    "policyId": None,
                    "policyName": None,
                    "status": "failed", 
                    "message": f"Error during policy sync: {str(e)}"
                }]
            )
    
    def _is_healthcare_relevant_policy(self, policy: Dict[str, Any]) -> bool:
        """
        Determina se uma política IAM é relevante para o módulo Healthcare.
        
        Args:
            policy: Dados da política
            
        Returns:
            True se a política for relevante para Healthcare
        """
        # Palavras-chave que indicam relevância para Healthcare
        healthcare_keywords = [
            "healthcare", "health", "medical", "patient", "hipaa", 
            "phi", "clinical", "hospital", "diagnosis", "treatment"
        ]
        
        # Verificar nome e descrição
        policy_text = f"{policy.get('name', '')} {policy.get('description', '')}".lower()
        
        # Verificar tags específicas de Healthcare
        policy_tags = [tag.lower() for tag in policy.get("tags", [])]
        if "healthcare" in policy_tags or "medical" in policy_tags or "hipaa" in policy_tags:
            return True
        
        # Verificar se a política é especificamente marcada para HIPAA
        if policy.get("requiredBy") and "hipaa" in [r.lower() for r in policy.get("requiredBy")]:
            return True
            
        # Verificar palavras-chave no texto da política
        return any(keyword in policy_text for keyword in healthcare_keywords)


# Modelos de dados para saída GraphQL
class HealthcareComplianceValidationResult:
    """Resultado de validação de conformidade Healthcare."""
    
    def __init__(
        self,
        validation_id: str,
        status: str,
        score: float,
        findings: List[Dict[str, Any]],
        category_scores: Dict[str, float],
        timestamp: datetime.datetime
    ):
        self.validation_id = validation_id
        self.status = status
        self.score = score
        self.findings = findings
        self.category_scores = category_scores
        self.timestamp = timestamp
        

class HealthcarePolicySyncResult:
    """Resultado de sincronização de políticas de IAM para Healthcare."""
    
    def __init__(
        self,
        synced_count: int,
        successful: int,
        failed: int,
        details: List[Dict[str, Any]]
    ):
        self.synced_count = synced_count
        self.successful = successful
        self.failed = failed
        self.details = details
