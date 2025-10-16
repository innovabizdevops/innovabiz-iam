from fastapi import APIRouter, Depends, HTTPException, status, Path, Query, Body
from sqlalchemy.orm import Session
from typing import Dict, Any, List, Optional
from uuid import UUID
from datetime import datetime, date
import json

from app.db.session import get_db
from app.schemas.healthcare import (
    HealthcareComplianceValidationRequest, HealthcareComplianceResult, 
    HealthcareComplianceValidationSummary, HealthcareComplianceHistoryFilter,
    ComplianceRequirement
)
from app.services.auth_service import AuthService
from app.core.config import settings
from app.core.logger import logger
from app.middlewares.tenant_middleware import get_current_tenant_id
from app.routers.auth import get_current_user

router = APIRouter(prefix="/healthcare", tags=["Healthcare Compliance"])

@router.post("/compliance/validate/{regulation}", response_model=HealthcareComplianceResult)
async def validate_healthcare_compliance(
    regulation: str = Path(..., description="Regulamentação a ser validada: hipaa, gdpr_health, lgpd_health, pndsb, all"),
    parameters: Optional[Dict[str, Any]] = Body({}, description="Parâmetros adicionais para validação"),
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Executa uma validação de compliance para regulamentações de saúde
    """
    # Verifica se a regulamentação é válida
    valid_regulations = ["hipaa", "gdpr_health", "lgpd_health", "pndsb", "all"]
    if regulation not in valid_regulations:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Regulamentação inválida. Deve ser uma das seguintes: {', '.join(valid_regulations)}"
        )
    
    # Verifica permissão para executar validação
    has_permission = await AuthService.check_permission(
        db=db,
        user_id=UUID(current_user["id"]),
        resource_type="healthcare_compliance",
        action="validate",
        resource_id=regulation
    )
    
    if not has_permission:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Sem permissão para executar validação de compliance"
        )
    
    try:
        tenant_id = get_current_tenant_id()
        organization_id = current_user["organization_id"]
        
        # Executar validação específica de acordo com a regulamentação
        if regulation == "all":
            # Executa todas as validações
            results = {}
            all_passed = True
            total_score = 0
            all_checks = []
            
            for reg in ["hipaa", "gdpr_health", "lgpd_health", "pndsb"]:
                reg_result = await _execute_compliance_validation(
                    db=db,
                    regulation=reg,
                    tenant_id=tenant_id,
                    organization_id=organization_id,
                    user_id=current_user["id"],
                    parameters=parameters
                )
                
                results[reg] = reg_result
                all_passed = all_passed and reg_result["status"] == "passed"
                total_score += reg_result["score"]
                all_checks.extend(reg_result["checks"])
            
            # Calcula o score médio
            avg_score = total_score // 4
            
            # Determina o status geral
            overall_status = "passed" if all_passed else "failed"
            
            # Prepara o resultado agregado
            result = {
                "regulation": "all",
                "timestamp": datetime.utcnow(),
                "validator": "IAM Healthcare Compliance Validator",
                "status": overall_status,
                "score": avg_score,
                "checks": all_checks,
                "remediation_plan": _generate_remediation_plan(all_checks) if not all_passed else None
            }
        else:
            # Executa uma validação específica
            result = await _execute_compliance_validation(
                db=db,
                regulation=regulation,
                tenant_id=tenant_id,
                organization_id=organization_id,
                user_id=current_user["id"],
                parameters=parameters
            )
        
        # Registra resultado da validação
        validation_id = UUID.uuid4()
        
        db.execute(
            """
            INSERT INTO iam.healthcare_compliance_validations (
                id, organization_id, regulation, validator_name, 
                status, score, validated_by, validation_timestamp,
                validation_result
            )
            VALUES (
                :id, :organization_id, :regulation, :validator_name,
                :status, :score, :validated_by, :validation_timestamp,
                :validation_result::jsonb
            )
            """,
            {
                "id": str(validation_id),
                "organization_id": organization_id,
                "regulation": regulation,
                "validator_name": result["validator"],
                "status": result["status"],
                "score": result["score"],
                "validated_by": current_user["id"],
                "validation_timestamp": result["timestamp"],
                "validation_result": json.dumps(result)
            }
        )
        
        # Registra em log para auditoria
        db.execute(
            """
            SELECT iam.log_audit_event(
                :tenant_id, 'healthcare_compliance_validation', 'compliance_validation', :validation_id,
                'validate', :user_id, :metadata::jsonb
            )
            """,
            {
                "tenant_id": tenant_id,
                "validation_id": str(validation_id),
                "user_id": current_user["id"],
                "metadata": json.dumps({
                    "regulation": regulation,
                    "status": result["status"],
                    "score": result["score"],
                    "parameters": parameters
                })
            }
        )
        
        db.commit()
        
        return result
    
    except HTTPException:
        raise
    except Exception as e:
        db.rollback()
        logger.error(f"Error validating healthcare compliance: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao executar validação de compliance"
        )

@router.get("/compliance/history", response_model=Dict[str, Any])
async def get_healthcare_compliance_history(
    page: int = Query(1, description="Página atual"),
    per_page: int = Query(20, description="Itens por página", ge=1, le=100),
    regulation: Optional[str] = Query(None, description="Filtro por regulamentação"),
    start_date: Optional[date] = Query(None, description="Data inicial (YYYY-MM-DD)"),
    end_date: Optional[date] = Query(None, description="Data final (YYYY-MM-DD)"),
    status: Optional[str] = Query(None, description="Filtro por status"),
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> Dict[str, Any]:
    """
    Obtém o histórico de validações de compliance
    """
    if regulation and regulation not in ["hipaa", "gdpr_health", "lgpd_health", "pndsb", "all"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Regulamentação inválida"
        )
    
    if status and status not in ["passed", "failed", "warning", "not_applicable"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Status inválido"
        )
    
    try:
        tenant_id = get_current_tenant_id()
        organization_id = current_user["organization_id"]
        
        # Constrói a consulta base
        query_conditions = [
            "organization_id = :organization_id"
        ]
        query_params = {
            "organization_id": organization_id,
            "offset": (page - 1) * per_page,
            "limit": per_page
        }
        
        # Adiciona filtros condicionais
        if regulation:
            query_conditions.append("regulation = :regulation")
            query_params["regulation"] = regulation
        
        if start_date:
            query_conditions.append("validation_timestamp >= :start_date")
            query_params["start_date"] = start_date
        
        if end_date:
            query_conditions.append("validation_timestamp <= :end_date")
            query_params["end_date"] = end_date
        
        if status:
            query_conditions.append("status = :status")
            query_params["status"] = status
        
        # Constrói a consulta final
        conditions_sql = " AND ".join(query_conditions)
        
        # Consulta o total de itens
        count_sql = f"""
            SELECT COUNT(*) FROM iam.healthcare_compliance_validations
            WHERE {conditions_sql}
        """
        
        total_items = db.execute(count_sql, query_params).scalar()
        
        # Consulta os itens paginados
        items_sql = f"""
            SELECT id, validation_timestamp, regulation, validator_name, status, score, validated_by
            FROM iam.healthcare_compliance_validations
            WHERE {conditions_sql}
            ORDER BY validation_timestamp DESC
            OFFSET :offset LIMIT :limit
        """
        
        result = db.execute(items_sql, query_params).fetchall()
        
        # Converte o resultado para lista de dicionários
        validations = []
        for row in result:
            validation = dict(row._mapping)
            
            # Converte UUID para string
            validation["id"] = str(validation["id"])
            if validation["validated_by"]:
                validation["validated_by"] = str(validation["validated_by"])
            
            # Formata a data
            validation["validation_timestamp"] = validation["validation_timestamp"].isoformat()
            
            validations.append(validation)
        
        # Calcula a paginação
        total_pages = (total_items + per_page - 1) // per_page
        
        # Prepara a resposta
        response = {
            "validations": validations,
            "pagination": {
                "total_items": total_items,
                "total_pages": total_pages,
                "current_page": page,
                "per_page": per_page,
                "next_page": page + 1 if page < total_pages else None,
                "prev_page": page - 1 if page > 1 else None
            }
        }
        
        return response
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error getting healthcare compliance history: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao obter histórico de validações de compliance"
        )

@router.get("/compliance/requirements", response_model=List[ComplianceRequirement])
async def get_compliance_requirements(
    regulation: Optional[str] = Query(None, description="Filtro por regulamentação"),
    db: Session = Depends(get_db),
    current_user: Dict[str, Any] = Depends(get_current_user)
) -> List[Dict[str, Any]]:
    """
    Obtém os requisitos de compliance para regulamentações de saúde
    """
    if regulation and regulation not in ["hipaa", "gdpr_health", "lgpd_health", "pndsb"]:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Regulamentação inválida"
        )
    
    try:
        # Constrói a consulta base
        query_conditions = ["is_active = TRUE"]
        query_params = {}
        
        # Adiciona filtro por regulamentação
        if regulation:
            query_conditions.append("regulation = :regulation")
            query_params["regulation"] = regulation
        
        # Constrói a consulta final
        conditions_sql = " AND ".join(query_conditions)
        
        # Consulta os requisitos
        sql = f"""
            SELECT * FROM iam.compliance_requirements
            WHERE {conditions_sql}
            ORDER BY regulation, code
        """
        
        result = db.execute(sql, query_params).fetchall()
        
        # Converte o resultado para lista de dicionários
        requirements = []
        for row in result:
            requirement = dict(row._mapping)
            
            # Converte UUID para string
            requirement["id"] = str(requirement["id"])
            
            # Formata as datas
            requirement["created_at"] = requirement["created_at"].isoformat()
            if requirement["updated_at"]:
                requirement["updated_at"] = requirement["updated_at"].isoformat()
            
            requirements.append(requirement)
        
        return requirements
    
    except Exception as e:
        logger.error(f"Error getting compliance requirements: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Erro ao obter requisitos de compliance"
        )

# Funções auxiliares

async def _execute_compliance_validation(
    db: Session,
    regulation: str,
    tenant_id: str,
    organization_id: str,
    user_id: str,
    parameters: Dict[str, Any]
) -> Dict[str, Any]:
    """
    Executa a validação de compliance para uma regulamentação específica
    """
    # Obtém os requisitos da regulamentação
    requirements = db.execute(
        """
        SELECT * FROM iam.compliance_requirements
        WHERE regulation = :regulation AND is_active = TRUE
        ORDER BY code
        """,
        {"regulation": regulation}
    ).fetchall()
    
    if not requirements:
        return {
            "regulation": regulation,
            "timestamp": datetime.utcnow(),
            "validator": "IAM Healthcare Compliance Validator",
            "status": "not_applicable",
            "score": 0,
            "checks": [
                {
                    "name": "requirements_not_found",
                    "requirement": f"Requisitos para {regulation} não encontrados",
                    "status": "not_applicable",
                    "details": "Não há requisitos configurados para esta regulamentação"
                }
            ],
            "remediation_plan": None
        }
    
    # Executa as validações para cada requisito
    checks = []
    failed_checks = []
    total_score = 0
    max_score = 0
    
    for req in requirements:
        req_dict = dict(req._mapping)
        check_result = await _validate_requirement(
            db=db,
            requirement=req_dict,
            tenant_id=tenant_id,
            organization_id=organization_id,
            parameters=parameters
        )
        
        checks.append(check_result)
        
        if check_result["status"] == "failed":
            failed_checks.append(check_result)
        
        # Calcula o score
        if check_result["status"] != "not_applicable":
            severity_weight = {
                "critical": 4,
                "high": 3,
                "medium": 2,
                "low": 1
            }
            
            weight = severity_weight.get(req_dict["severity"], 1)
            max_score += weight * 25  # 25 pontos por requisito (ajustado pelo peso)
            
            if check_result["status"] == "passed":
                total_score += weight * 25
            elif check_result["status"] == "warning":
                total_score += weight * 12.5  # Metade dos pontos para avisos
    
    # Calcula o score final (0-100)
    final_score = 100 if max_score == 0 else (total_score * 100) // max_score
    
    # Determina o status geral
    if not checks:
        status = "not_applicable"
    elif failed_checks:
        status = "failed"
    elif any(check["status"] == "warning" for check in checks):
        status = "warning"
    else:
        status = "passed"
    
    # Gera plano de remediação
    remediation_plan = _generate_remediation_plan(failed_checks) if failed_checks else None
    
    # Prepara o resultado
    result = {
        "regulation": regulation,
        "timestamp": datetime.utcnow(),
        "validator": "IAM Healthcare Compliance Validator",
        "status": status,
        "score": final_score,
        "checks": checks,
        "remediation_plan": remediation_plan
    }
    
    return result

async def _validate_requirement(
    db: Session,
    requirement: Dict[str, Any],
    tenant_id: str,
    organization_id: str,
    parameters: Dict[str, Any]
) -> Dict[str, Any]:
    """
    Executa a validação para um requisito específico
    """
    req_code = requirement["code"]
    req_name = requirement["name"]
    validation_method = requirement["validation_method"]
    
    # Resultado padrão
    result = {
        "name": req_code,
        "requirement": req_name,
        "status": "not_applicable",
        "details": "Validação não implementada"
    }
    
    try:
        # Executa a validação de acordo com o método
        if validation_method == "automatic":
            # Chama a função específica para validar o requisito
            function_name = f"validate_{requirement['regulation']}_{req_code.lower()}"
            
            if hasattr(globals().get("_validation_functions", {}), function_name):
                validation_result = await getattr(globals()["_validation_functions"], function_name)(
                    db=db,
                    tenant_id=tenant_id,
                    organization_id=organization_id,
                    parameters=parameters
                )
                
                result.update(validation_result)
            else:
                # Fallback para validação baseada em regras SQL
                sql_validation = db.execute(
                    """
                    SELECT * FROM iam.validate_compliance_requirement(
                        :tenant_id, :organization_id, :regulation, :requirement_code, :parameters::jsonb
                    )
                    """,
                    {
                        "tenant_id": tenant_id,
                        "organization_id": organization_id,
                        "regulation": requirement["regulation"],
                        "requirement_code": req_code,
                        "parameters": json.dumps(parameters)
                    }
                ).fetchone()
                
                if sql_validation:
                    validation_dict = dict(sql_validation._mapping)
                    result.update({
                        "status": validation_dict["status"],
                        "details": validation_dict["details"]
                    })
        elif validation_method == "manual":
            # Para validações manuais, verifica se existe um registro de validação manual
            manual_validation = db.execute(
                """
                SELECT * FROM iam.manual_compliance_validations
                WHERE organization_id = :organization_id
                AND regulation = :regulation
                AND requirement_code = :requirement_code
                AND is_active = TRUE
                ORDER BY validation_date DESC
                LIMIT 1
                """,
                {
                    "organization_id": organization_id,
                    "regulation": requirement["regulation"],
                    "requirement_code": req_code
                }
            ).fetchone()
            
            if manual_validation:
                validation_dict = dict(manual_validation._mapping)
                result.update({
                    "status": validation_dict["status"],
                    "details": validation_dict["details"]
                })
            else:
                result.update({
                    "status": "warning",
                    "details": "Este requisito requer validação manual, mas nenhuma validação foi encontrada"
                })
    except Exception as e:
        logger.error(f"Error validating requirement {req_code}: {str(e)}")
        result.update({
            "status": "error",
            "details": f"Erro ao validar requisito: {str(e)}"
        })
    
    return result

def _generate_remediation_plan(failed_checks: List[Dict[str, Any]]) -> str:
    """
    Gera um plano de remediação baseado nos checks falhados
    """
    if not failed_checks:
        return None
    
    plan = "# Plano de Remediação\n\n"
    plan += "Os seguintes itens de compliance não foram atendidos e requerem ação corretiva:\n\n"
    
    for i, check in enumerate(failed_checks, 1):
        plan += f"## {i}. {check['requirement']}\n\n"
        plan += f"**Status:** {check['status']}\n\n"
        plan += f"**Detalhes:** {check['details']}\n\n"
        
        # Adiciona ações de remediação recomendadas
        # Na prática, estas ações viriam de uma base de conhecimento
        plan += "**Ações Recomendadas:**\n\n"
        
        if "hipaa" in check['name'].lower():
            plan += "- Revise as políticas de controle de acesso aos dados de saúde\n"
            plan += "- Implemente mecanismos de registro de auditoria para acessos a PHI\n"
            plan += "- Assegure que todos os dados em repouso e em trânsito estejam criptografados\n"
        elif "gdpr" in check['name'].lower():
            plan += "- Atualize os termos de consentimento para processamento de dados de saúde\n"
            plan += "- Implemente mecanismos para atender pedidos de exclusão de dados\n"
            plan += "- Documente a base legal para o processamento de dados sensíveis\n"
        elif "lgpd" in check['name'].lower():
            plan += "- Atualize a política de privacidade conforme requisitos da LGPD\n"
            plan += "- Implemente mecanismos para notificação de vazamentos\n"
            plan += "- Documente os fluxos de dados sensíveis no sistema\n"
        elif "pndsb" in check['name'].lower():
            plan += "- Atualize os metadados conforme requisitos da PNDSB\n"
            plan += "- Implemente controles de interoperabilidade para dados de saúde\n"
            plan += "- Revise os processos de consentimento específicos para a legislação brasileira\n"
        
        plan += "\n---\n\n"
    
    plan += "Este plano deve ser implementado o mais rápido possível para garantir conformidade com as regulamentações aplicáveis."
    
    return plan

# Módulo de funções de validação específicas
class _validation_functions:
    @staticmethod
    async def validate_hipaa_access_control(db, tenant_id, organization_id, parameters):
        """Valida controles de acesso HIPAA"""
        # Implementação de exemplo - em produção seria mais complexa
        access_controls = db.execute(
            """
            SELECT COUNT(*) FROM iam.access_policies 
            WHERE organization_id = :organization_id
            AND policy_data->>'scope' = 'healthcare'
            AND is_active = TRUE
            """,
            {"organization_id": organization_id}
        ).scalar()
        
        if access_controls == 0:
            return {
                "status": "failed",
                "details": "Não foram encontradas políticas de controle de acesso específicas para dados de saúde"
            }
        
        return {
            "status": "passed",
            "details": f"Encontradas {access_controls} políticas de controle de acesso para dados de saúde"
        }
    
    @staticmethod
    async def validate_lgpd_consent(db, tenant_id, organization_id, parameters):
        """Valida mecanismos de consentimento LGPD"""
        consent_mechanisms = db.execute(
            """
            SELECT COUNT(*) FROM iam.consent_records
            WHERE organization_id = :organization_id
            AND consent_type = 'health_data'
            AND is_active = TRUE
            """,
            {"organization_id": organization_id}
        ).scalar()
        
        if consent_mechanisms == 0:
            return {
                "status": "failed",
                "details": "Não foram encontrados registros de consentimento para dados de saúde"
            }
        
        return {
            "status": "passed",
            "details": f"Encontrados {consent_mechanisms} registros de consentimento para dados de saúde"
        }
