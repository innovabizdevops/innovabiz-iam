"""
INNOVABIZ - Módulo de Validação e Certificação IAM
==================================================================
Autor: Eduardo Jeremias
Data: 07/05/2025
Versão: 1.0
Descrição: Pacote para validação, certificação e conformidade do módulo IAM,
           incluindo verificações de múltiplos frameworks regulatórios.
==================================================================
"""

from .iam_validator import (
    IAMValidator,
    ValidationReport,
    ValidationResult,
    ValidationStatus,
    ValidationSeverity,
    ValidationType
)

__version__ = "1.0.0"
__all__ = [
    "IAMValidator",
    "ValidationReport",
    "ValidationResult",
    "ValidationStatus",
    "ValidationSeverity",
    "ValidationType"
]
