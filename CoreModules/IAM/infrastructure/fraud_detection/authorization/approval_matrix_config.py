#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
Configurações para a Matriz de Autorização e Aprovação para Alertas Comportamentais

Este módulo contém configurações de exemplo para a matriz de aprovação,
incluindo níveis de autoridade, thresholds, permissões por papel e regras regionais.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 20/08/2025
"""

# Configuração padrão para matriz de aprovação
DEFAULT_APPROVAL_MATRIX_CONFIG = {
    "approval_criteria": {
        # Matriz de thresholds para determinar nível de aprovação
        "threshold_matrix": {
            # Severidade baixa
            "low": {
                "authentication": {
                    "500": "l1_agent",     # Até R$500 - Agente L1
                    "5000": "l2_specialist", # Até R$5000 - Especialista L2
                    "9999999999": "l3_supervisor"  # Acima - Supervisor L3
                },
                "transaction": {
                    "1000": "l1_agent",     # Até R$1000 - Agente L1
                    "10000": "l2_specialist", # Até R$10000 - Especialista L2
                    "50000": "l3_supervisor",  # Até R$50000 - Supervisor L3
                    "9999999999": "l4_manager"   # Acima - Gerente L4
                },
                "session": {
                    "9999999999": "l1_agent"  # Qualquer valor - Agente L1
                },
                "device": {
                    "9999999999": "l1_agent"  # Qualquer valor - Agente L1
                },
                "location": {
                    "9999999999": "l1_agent"  # Qualquer valor - Agente L1
                },
                "profile": {
                    "9999999999": "l1_agent"  # Qualquer valor - Agente L1
                },
                "combined": {
                    "1000": "l1_agent",     # Até R$1000 - Agente L1
                    "10000": "l2_specialist", # Até R$10000 - Especialista L2
                    "9999999999": "l3_supervisor"  # Acima - Supervisor L3
                }
            },
            # Severidade média
            "medium": {
                "authentication": {
                    "100": "l1_agent",     # Até R$100 - Agente L1
                    "1000": "l2_specialist", # Até R$1000 - Especialista L2
                    "10000": "l3_supervisor",  # Até R$10000 - Supervisor L3
                    "9999999999": "l4_manager"   # Acima - Gerente L4
                },
                "transaction": {
                    "500": "l2_specialist", # Até R$500 - Especialista L2
                    "5000": "l3_supervisor",  # Até R$5000 - Supervisor L3
                    "50000": "l4_manager",   # Até R$50000 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "session": {
                    "1000": "l1_agent",     # Até R$1000 - Agente L1
                    "9999999999": "l2_specialist"  # Acima - Especialista L2
                },
                "device": {
                    "1000": "l1_agent",     # Até R$1000 - Agente L1
                    "9999999999": "l2_specialist"  # Acima - Especialista L2
                },
                "location": {
                    "1000": "l1_agent",     # Até R$1000 - Agente L1
                    "10000": "l2_specialist", # Até R$10000 - Especialista L2
                    "9999999999": "l3_supervisor"  # Acima - Supervisor L3
                },
                "profile": {
                    "1000": "l2_specialist", # Até R$1000 - Especialista L2
                    "9999999999": "l3_supervisor"  # Acima - Supervisor L3
                },
                "combined": {
                    "500": "l2_specialist", # Até R$500 - Especialista L2
                    "5000": "l3_supervisor",  # Até R$5000 - Supervisor L3
                    "9999999999": "l4_manager"   # Acima - Gerente L4
                }
            },
            # Severidade alta
            "high": {
                "authentication": {
                    "100": "l2_specialist", # Até R$100 - Especialista L2
                    "1000": "l3_supervisor",  # Até R$1000 - Supervisor L3
                    "10000": "l4_manager",   # Até R$10000 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "transaction": {
                    "100": "l3_supervisor",  # Até R$100 - Supervisor L3
                    "1000": "l4_manager",   # Até R$1000 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "session": {
                    "100": "l2_specialist", # Até R$100 - Especialista L2
                    "1000": "l3_supervisor",  # Até R$1000 - Supervisor L3
                    "9999999999": "l4_manager"   # Acima - Gerente L4
                },
                "device": {
                    "100": "l2_specialist", # Até R$100 - Especialista L2
                    "1000": "l3_supervisor",  # Até R$1000 - Supervisor L3
                    "9999999999": "l4_manager"   # Acima - Gerente L4
                },
                "location": {
                    "100": "l3_supervisor",  # Até R$100 - Supervisor L3
                    "1000": "l4_manager",   # Até R$1000 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "profile": {
                    "100": "l3_supervisor",  # Até R$100 - Supervisor L3
                    "1000": "l4_manager",   # Até R$1000 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "combined": {
                    "100": "l3_supervisor",  # Até R$100 - Supervisor L3
                    "1000": "l4_manager",   # Até R$1000 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                }
            },
            # Severidade crítica
            "critical": {
                "authentication": {
                    "50": "l3_supervisor",  # Até R$50 - Supervisor L3
                    "500": "l4_manager",   # Até R$500 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "transaction": {
                    "50": "l4_manager",   # Até R$50 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "session": {
                    "50": "l3_supervisor",  # Até R$50 - Supervisor L3
                    "500": "l4_manager",   # Até R$500 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "device": {
                    "50": "l3_supervisor",  # Até R$50 - Supervisor L3
                    "500": "l4_manager",   # Até R$500 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "location": {
                    "50": "l4_manager",   # Até R$50 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "profile": {
                    "50": "l4_manager",   # Até R$50 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                },
                "combined": {
                    "50": "l4_manager",   # Até R$50 - Gerente L4
                    "9999999999": "l5_director"  # Acima - Diretor L5
                }
            }
        },
        
        # Permissões específicas por papel
        "role_permissions": {
            "l1_agent": {
                "low": ["approve", "escalate", "monitor"],
                "medium": ["escalate", "monitor"],
                "high": ["escalate"],
                "critical": ["escalate"]
            },
            "l2_specialist": {
                "low": ["approve", "reject", "monitor"],
                "medium": ["approve", "reject", "escalate", "monitor"],
                "high": ["escalate", "investigate"],
                "critical": ["escalate"]
            },
            "l3_supervisor": {
                "low": ["approve", "reject", "monitor"],
                "medium": ["approve", "reject", "monitor"],
                "high": ["approve", "reject", "escalate", "investigate"],
                "critical": ["escalate", "investigate", "block"]
            },
            "l4_manager": {
                "low": ["approve", "reject", "monitor"],
                "medium": ["approve", "reject", "monitor"],
                "high": ["approve", "reject", "investigate", "block", "restrict"],
                "critical": ["approve", "reject", "investigate", "block", "restrict"]
            },
            "l5_director": {
                "default": ["approve", "reject", "escalate", "investigate", "challenge", "block", "restrict", "monitor"]
            }
        },
        
        # Overrides específicos por região
        "region_overrides": {
            # Configurações específicas para Brasil
            "BR": {
                "critical": {
                    "transaction": "l5_director",  # Todas transações críticas no Brasil requerem aprovação de diretor
                    "profile": "l5_director"
                }
            },
            # Configurações específicas para Moçambique
            "MZ": {
                "high": {
                    "transaction": "l4_manager"  # Todas transações de alta severidade em Moçambique requerem gerente
                },
                "critical": {
                    "transaction": "l5_director"  # Todas transações críticas em Moçambique requerem diretor
                }
            }
        },
        
        # Regras por segmento de cliente
        "customer_segment_rules": {
            "premium": {
                "medium": {
                    "5000": "l1_agent",      # Clientes premium têm limites mais altos para aprovação L1
                    "50000": "l2_specialist",
                    "100000": "l3_supervisor"
                },
                "high": {
                    "1000": "l2_specialist",
                    "10000": "l3_supervisor"
                }
            },
            "corporate": {
                "medium": {
                    "10000": "l1_agent",     # Clientes corporativos têm limites mais altos
                    "100000": "l2_specialist",
                    "500000": "l3_supervisor"
                },
                "high": {
                    "5000": "l2_specialist",
                    "50000": "l3_supervisor"
                }
            }
        },
        
        # Regras para auto-aprovação
        "auto_approval_rules": {
            # Limites por severidade
            "severity_limits": {
                "low": 500.0,    # Auto-aprovação até R$500 para severidade baixa
                "medium": 100.0  # Auto-aprovação até R$100 para severidade média
                # Alta e crítica não têm auto-aprovação (ausentes)
            },
            # Limites por categoria
            "category_limits": {
                "authentication": 300.0,
                "session": 300.0,
                "device": 200.0
                # Outras categorias não têm auto-aprovação (ausentes)
            },
            # Score máximo para auto-aprovação
            "max_risk_score": 0.3,
            # Limites específicos por região
            "region_limits": {
                "BR": 200.0,
                "MZ": 100.0,
                "AO": 50.0
            },
            # Limites por segmento de cliente
            "segment_limits": {
                "premium": 1000.0,
                "corporate": 2000.0,
                "default": 200.0
            }
        },
        
        # Regras para escalação automática
        "escalation_rules": {
            # Tempo máximo em minutos antes da escalação automática
            "timeout_minutes": {
                "l1_agent": 30,
                "l2_specialist": 60,
                "l3_supervisor": 120,
                "l4_manager": 240,
                "l5_director": 480
            },
            # Escalação por severidade e valor
            "auto_escalate": {
                "critical": {
                    "any_value": True  # Sempre escalar alertas críticos
                },
                "high": {
                    "min_value": 10000  # Escalar alertas de alta severidade acima de R$10000
                }
            }
        },
        
        # Valor máximo padrão para verificações
        "default_max_amount": 9999999999
    },
    
    # Configurações gerais da matriz
    "max_escalation_levels": 3,
    "auto_approval_enabled": True,
    "require_comments_for_reject": True,
    "store_audit_trail": True
}


# Configurações específicas para diferentes regiões
REGION_CONFIGS = {
    # Configuração para Brasil
    "BR": {
        "approval_criteria": {
            # Override apenas dos elementos específicos para o Brasil
            "threshold_matrix": {
                "high": {
                    "transaction": {
                        "1000": "l3_supervisor",
                        "10000": "l4_manager",
                        "9999999999": "l5_director"
                    }
                }
            },
            "auto_approval_rules": {
                "severity_limits": {
                    "low": 400.0,  # Limite menor para Brasil
                    "medium": 80.0
                },
                "max_risk_score": 0.25  # Score mais rigoroso para Brasil
            }
        },
        "require_comments_for_reject": True
    },
    
    # Configuração para Moçambique
    "MZ": {
        "approval_criteria": {
            "auto_approval_rules": {
                "severity_limits": {
                    "low": 200.0,  # Limite menor para Moçambique
                    "medium": 50.0
                },
                "max_risk_score": 0.2  # Score mais rigoroso para Moçambique
            }
        }
    }
}


def get_approval_matrix_config(region: str = None) -> dict:
    """
    Obtém a configuração da matriz de aprovação para uma região específica.
    
    Args:
        region: Código da região
        
    Returns:
        Configuração da matriz de aprovação
    """
    config = DEFAULT_APPROVAL_MATRIX_CONFIG.copy()
    
    # Se uma região específica foi solicitada e existe configuração regional
    if region and region in REGION_CONFIGS:
        region_config = REGION_CONFIGS[region]
        
        # Realizar merge profundo das configurações
        _deep_merge(config, region_config)
    
    return config


def _deep_merge(base: dict, override: dict) -> dict:
    """
    Realiza merge profundo de dicionários, atualizando o dicionário base.
    
    Args:
        base: Dicionário base
        override: Dicionário com valores a sobrescrever
        
    Returns:
        Dicionário base atualizado
    """
    for key, value in override.items():
        if key in base and isinstance(base[key], dict) and isinstance(value, dict):
            _deep_merge(base[key], value)
        else:
            base[key] = value
    return base