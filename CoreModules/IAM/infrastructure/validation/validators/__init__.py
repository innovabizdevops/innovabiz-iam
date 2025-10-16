"""
INNOVABIZ - Pacote de Validadores Regionais de IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Implementação de validadores regionais de conformidade para o 
           módulo IAM, incluindo validadores para UE, Brasil, EUA e África.
==================================================================
"""

import logging
from typing import Dict, List, Any, Set, Optional
from ..models import ComplianceValidationResult
from ..compliance_engine import ValidationRule
from ..compliance_metadata import Region

# Importação de validadores regionais
from .eu_validator import get_eu_rules
from .brazil_validator import get_brazil_rules
from .us_validator import get_us_rules
from .africa_validator import get_africa_rules

# Configuração de logging
logger = logging.getLogger(__name__)

def get_validators_for_region(region: Region) -> List[ValidationRule]:
    """
    Retorna todas as regras de validação aplicáveis para uma região específica.
    
    Args:
        region: A região para a qual obter validadores
        
    Returns:
        Lista de regras de validação aplicáveis à região especificada
    """
    all_validators = []
    
    # Mapear regiões para seus respectivos validadores
    region_map = {
        Region.EU: get_eu_rules(),
        Region.BRAZIL: get_brazil_rules(),
        Region.USA: get_us_rules(),
        Region.AFRICA: get_africa_rules(),
        Region.ANGOLA: get_africa_rules(),
        Region.GLOBAL: get_all_validators()
    }
    
    # Obter validadores para a região especificada
    validators = region_map.get(region, [])
    if validators:
        all_validators.extend(validators)
    
    return all_validators

def get_all_validators() -> List[ValidationRule]:
    """
    Retorna todas as regras de validação de todas as regiões.
    
    Returns:
        Lista combinada de todas as regras de validação implementadas
    """
    all_validators = []
    all_validators.extend(get_eu_rules())
    all_validators.extend(get_brazil_rules())
    all_validators.extend(get_us_rules())
    all_validators.extend(get_africa_rules())
    return all_validators
