"""
Módulo de Validação de Conformidade

Este pacote contém os validadores de conformidade para diferentes regiões
e mercados, suportando o framework multi-tenant e multi-mercado.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

from .base_compliance_validator import BaseComplianceValidator
from .unified_compliance_validator import UnifiedComplianceValidator
from .brics_compliance_validator import BRICSComplianceValidator

# Exportar classes principais para facilitar importação
__all__ = [
    'BaseComplianceValidator',
    'UnifiedComplianceValidator',
    'BRICSComplianceValidator'
]
