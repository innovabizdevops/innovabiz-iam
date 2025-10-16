"""
Constantes específicas para o módulo TrustGuard

Este arquivo define as constantes utilizadas pelo módulo TrustGuard,
incluindo mercados, regiões, níveis de risco e outras definições
utilizadas para análise contextual multi-tenant.

@author: InnovaBiz DevOps Team
@copyright: InnovaBiz 2025
@version: 1.0.0
"""

# Mercados suportados pela plataforma
MERCADOS = {
    # África
    "AO": "Angola",
    "MZ": "Moçambique",
    "CV": "Cabo Verde",
    "GW": "Guiné-Bissau",
    "ST": "São Tomé e Príncipe",
    "ZA": "África do Sul",
    "NA": "Namíbia",
    "BW": "Botsuana",
    "SZ": "Eswatini",
    "LS": "Lesoto",
    "ZM": "Zâmbia",
    "ZW": "Zimbábue",
    "TZ": "Tanzânia",
    "MU": "Maurícia",
    "SC": "Seychelles",
    "MG": "Madagascar",
    "CD": "República Democrática do Congo",
    "CG": "Congo",
    "GA": "Gabão",
    "GQ": "Guiné Equatorial",
    
    # América
    "BR": "Brasil",
    "US": "Estados Unidos",
    "CA": "Canadá",
    "MX": "México",
    "AR": "Argentina",
    "CL": "Chile",
    "CO": "Colômbia",
    "PE": "Peru",
    "UY": "Uruguai",
    "PY": "Paraguai",
    
    # Europa
    "PT": "Portugal",
    "ES": "Espanha",
    "FR": "França",
    "DE": "Alemanha",
    "IT": "Itália",
    "GB": "Reino Unido",
    "BE": "Bélgica",
    "NL": "Países Baixos",
    "LU": "Luxemburgo",
    "CH": "Suíça",
    
    # Ásia
    "CN": "China",
    "IN": "Índia",
    "RU": "Rússia",
    "JP": "Japão",
    "KR": "Coreia do Sul",
}

# Regiões/Blocos econômicos
REGIOES = {
    "CPLP": "Comunidade dos Países de Língua Portuguesa",
    "SADC": "Comunidade de Desenvolvimento da África Austral",
    "PALOP": "Países Africanos de Língua Oficial Portuguesa",
    "BRICS": "Brasil, Rússia, Índia, China e África do Sul",
    "UE": "União Europeia",
    "MERCOSUL": "Mercado Comum do Sul",
    "NAFTA": "Acordo de Livre Comércio da América do Norte",
    "ASEAN": "Associação das Nações do Sudeste Asiático",
    "UA": "União Africana",
    "CEDEAO": "Comunidade Económica dos Estados da África Ocidental",
}

# Níveis de risco (em ordem crescente)
NIVEIS_RISCO = ["baixo", "medio", "alto", "critico"]

# Tipos de fatores de autenticação
FATORES_AUTENTICACAO = {
    "senha": {
        "tipo": "conhecimento",
        "nivel_seguranca": 1,
        "requisitos": ["cliente_web", "cliente_mobile"]
    },
    "pin": {
        "tipo": "conhecimento",
        "nivel_seguranca": 1,
        "requisitos": ["cliente_web", "cliente_mobile"]
    },
    "otp_sms": {
        "tipo": "posse",
        "nivel_seguranca": 2,
        "requisitos": ["telefone_celular"]
    },
    "otp_email": {
        "tipo": "posse",
        "nivel_seguranca": 2,
        "requisitos": ["email_verificado"]
    },
    "otp_app": {
        "tipo": "posse",
        "nivel_seguranca": 3,
        "requisitos": ["app_autenticador"]
    },
    "biometria_facial": {
        "tipo": "inerencia",
        "nivel_seguranca": 3,
        "requisitos": ["camera", "cliente_biometrico"]
    },
    "biometria_digital": {
        "tipo": "inerencia",
        "nivel_seguranca": 3,
        "requisitos": ["sensor_biometrico"]
    },
    "comportamental": {
        "tipo": "inerencia",
        "nivel_seguranca": 3,
        "requisitos": ["historico_comportamental"]
    },
    "token_fisico": {
        "tipo": "posse",
        "nivel_seguranca": 3,
        "requisitos": ["dispositivo_token"]
    },
    "certificado_digital": {
        "tipo": "posse",
        "nivel_seguranca": 4,
        "requisitos": ["armazenamento_certificado"]
    },
    "webauthn": {
        "tipo": "posse_inerencia",
        "nivel_seguranca": 4,
        "requisitos": ["navegador_compativel"]
    },
    "geolocation": {
        "tipo": "contexto",
        "nivel_seguranca": 2,
        "requisitos": ["gps", "ip_confiavel"]
    },
    "device_id": {
        "tipo": "contexto",
        "nivel_seguranca": 2,
        "requisitos": ["dispositivo_registrado"]
    },
}

# Tipos de transações financeiras com níveis de risco base
TRANSACOES_FINANCEIRAS = {
    "pagamento_comercio": {
        "nivel_risco_base": "baixo",
        "limite_sem_autenticacao": 100,
        "limite_autenticacao_simples": 500,
        "limite_autenticacao_forte": 5000,
    },
    "transferencia_mesma_instituicao": {
        "nivel_risco_base": "baixo",
        "limite_sem_autenticacao": 50,
        "limite_autenticacao_simples": 300,
        "limite_autenticacao_forte": 3000,
    },
    "transferencia_outra_instituicao": {
        "nivel_risco_base": "medio",
        "limite_sem_autenticacao": 0,
        "limite_autenticacao_simples": 200,
        "limite_autenticacao_forte": 2000,
    },
    "saque": {
        "nivel_risco_base": "medio",
        "limite_sem_autenticacao": 0,
        "limite_autenticacao_simples": 100,
        "limite_autenticacao_forte": 1000,
    },
    "pix": {
        "nivel_risco_base": "medio",
        "limite_sem_autenticacao": 20,
        "limite_autenticacao_simples": 200,
        "limite_autenticacao_forte": 2000,
    },
    "compra_online": {
        "nivel_risco_base": "alto",
        "limite_sem_autenticacao": 0,
        "limite_autenticacao_simples": 100,
        "limite_autenticacao_forte": 1000,
    },
    "pagamento_internacional": {
        "nivel_risco_base": "alto",
        "limite_sem_autenticacao": 0,
        "limite_autenticacao_simples": 0,
        "limite_autenticacao_forte": 1000,
    },
    "pagamento_boleto": {
        "nivel_risco_base": "baixo",
        "limite_sem_autenticacao": 50,
        "limite_autenticacao_simples": 500,
        "limite_autenticacao_forte": 5000,
    },
    "recarga_celular": {
        "nivel_risco_base": "baixo",
        "limite_sem_autenticacao": 100,
        "limite_autenticacao_simples": 200,
        "limite_autenticacao_forte": 500,
    },
    "pagamento_fatura": {
        "nivel_risco_base": "baixo",
        "limite_sem_autenticacao": 100,
        "limite_autenticacao_simples": 1000,
        "limite_autenticacao_forte": 10000,
    },
}

# Fatores de risco por país
RISCO_PAIS = {
    # África
    "AO": 60,  # Angola
    "MZ": 65,  # Moçambique
    "CV": 50,  # Cabo Verde
    "GW": 75,  # Guiné-Bissau
    "ST": 65,  # São Tomé e Príncipe
    "ZA": 55,  # África do Sul
    "NA": 60,  # Namíbia
    "BW": 55,  # Botsuana
    "SZ": 65,  # Eswatini
    "LS": 65,  # Lesoto
    "ZM": 70,  # Zâmbia
    "ZW": 70,  # Zimbábue
    "TZ": 65,  # Tanzânia
    "MU": 45,  # Maurícia
    "SC": 45,  # Seychelles
    "MG": 70,  # Madagascar
    "CD": 80,  # República Democrática do Congo
    "CG": 75,  # Congo
    "GA": 70,  # Gabão
    "GQ": 75,  # Guiné Equatorial
    
    # América
    "BR": 55,  # Brasil
    "US": 30,  # Estados Unidos
    "CA": 25,  # Canadá
    "MX": 60,  # México
    "AR": 55,  # Argentina
    "CL": 40,  # Chile
    "CO": 60,  # Colômbia
    "PE": 55,  # Peru
    "UY": 45,  # Uruguai
    "PY": 60,  # Paraguai
    
    # Europa
    "PT": 35,  # Portugal
    "ES": 35,  # Espanha
    "FR": 30,  # França
    "DE": 25,  # Alemanha
    "IT": 40,  # Itália
    "GB": 30,  # Reino Unido
    "BE": 30,  # Bélgica
    "NL": 25,  # Países Baixos
    "LU": 20,  # Luxemburgo
    "CH": 15,  # Suíça
    
    # Ásia
    "CN": 55,  # China
    "IN": 60,  # Índia
    "RU": 65,  # Rússia
    "JP": 20,  # Japão
    "KR": 30,  # Coreia do Sul
}

# Regras de compliance por região
REGRAS_COMPLIANCE = {
    "GDPR": ["UE", "global"],
    "LGPD": ["BR"],
    "POPIA": ["ZA"],
    "CCPA": ["US-CA"],
    "HIPAA": ["US"],
    "PCI_DSS": ["global"],
    "SOX": ["US", "global"],
    "BASEL_II": ["global"],
    "BASEL_III": ["global"],
    "ISO_27001": ["global"],
    "ISO_27701": ["global"],
    "CVM": ["BR"],
    "BACEN": ["BR"],
    "BNA": ["AO"],
    "FCA": ["GB"],
    "SEC": ["US"],
    "MIFID_II": ["UE"],
}

# Configurações padrão para monitoria de risco por tipo de serviço
MONITORAMENTO_PADRAO = {
    "iam": {
        "nivel_alerta": "medio",
        "fatores_obrigatorios": ["senha"],
        "fatores_recomendados": ["otp_app", "biometria_facial"],
        "intervalo_atualizacao_senha": 90,  # dias
        "max_tentativas_falhas": 5,
        "periodo_bloqueio": 30,  # minutos
    },
    "pagamentos": {
        "nivel_alerta": "medio",
        "fatores_obrigatorios": ["senha", "otp_sms"],
        "fatores_recomendados": ["biometria_facial", "otp_app"],
        "valor_maximo_sem_verificacao": 100,
        "valor_maximo_verificacao_simples": 1000,
    },
    "credito": {
        "nivel_alerta": "alto",
        "fatores_obrigatorios": ["senha", "otp_sms"],
        "fatores_recomendados": ["biometria_facial", "certificado_digital"],
        "verificacao_bureau": True,
        "verificacao_comportamental": True,
    },
    "investimento": {
        "nivel_alerta": "alto",
        "fatores_obrigatorios": ["senha", "otp_app"],
        "fatores_recomendados": ["certificado_digital", "token_fisico"],
        "valor_maximo_sem_verificacao": 0,
        "valor_maximo_verificacao_simples": 500,
    },
    "seguros": {
        "nivel_alerta": "medio",
        "fatores_obrigatorios": ["senha"],
        "fatores_recomendados": ["otp_sms", "otp_email"],
        "verificacao_sinistros": True,
    }
}
