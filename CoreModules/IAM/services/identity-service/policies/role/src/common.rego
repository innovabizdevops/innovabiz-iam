# Funções comuns e utilitárias para políticas do RoleService - INNOVABIZ Platform
#
# Conformidade: ISO/IEC 27001:2022, TOGAF 10.0, COBIT 2019, NIST SP 800-53,
# PCI DSS v4.0, GDPR, APD Angola, BNA, Basel III
package innovabiz.iam.role.common

import data.innovabiz.iam.role.constants

# ---------------------------------------------------------
# Validações básicas de usuário e requisição
# ---------------------------------------------------------

# Verifica se o usuário está autenticado
is_authenticated {
    # Usuário deve ter um ID válido
    input.user.id
    input.user.id != ""
}

# Verifica se a requisição é válida
is_valid_request {
    # A requisição deve ter um método HTTP válido
    input.http_method
    input.http_method != ""
    
    # A requisição deve ter um caminho
    input.path
    input.path != ""
    
    # O recurso deve existir
    input.resource
}

# Verifica se o tenant é válido
has_valid_tenant {
    # O tenant_id deve existir e não ser vazio
    input.tenant_id
    input.tenant_id != ""
    
    # O usuário deve ter um tenant_id
    input.user.tenant_id
    input.user.tenant_id != ""
}

# ---------------------------------------------------------
# Validações de autorização baseadas em usuário
# ---------------------------------------------------------

# Verifica se o usuário é um Super Admin
is_super_admin {
    # Usuário deve ter a função SUPER_ADMIN
    role_names := [role | role = input.user.roles[_]]
    contains(role_names, constants.super_admin_role)
}

# Verifica se o usuário é um Tenant Admin
is_tenant_admin {
    # Usuário deve ter a função TENANT_ADMIN
    role_names := [role | role = input.user.roles[_]]
    contains(role_names, constants.tenant_admin_role)
}

# Verifica se o usuário é um IAM Admin
is_iam_admin {
    # Usuário deve ter a função IAM_ADMIN
    role_names := [role | role = input.user.roles[_]]
    contains(role_names, constants.iam_admin_role)
}

# Verifica se o usuário é um IAM Operator
is_iam_operator {
    # Usuário deve ter a função IAM_OPERATOR
    role_names := [role | role = input.user.roles[_]]
    contains(role_names, constants.iam_operator_role)
}

# Verifica se o usuário pertence ao tenant do recurso
is_in_same_tenant {
    input.user.tenant_id == input.tenant_id
}

# Verifica se o usuário tem uma permissão específica
has_permission(permission) {
    # Permissão específica
    input.user.permissions[_] == permission
}

# Verifica se o usuário tem uma permissão com wildcard
has_wildcard_permission(permission_pattern) {
    # Verifica permissões exatas
    has_permission(permission_pattern)
} else {
    # Verifica permissões com wildcard
    user_permission := input.user.permissions[_]
    glob.match(user_permission, [], permission_pattern)
} else {
    # Verifica se tem permissão de administração total
    has_permission("*")
} else {
    # Verifica padrões de permissão com asterisco
    permission_parts := split(permission_pattern, ":")
    
    # Verifica padrão domínio:*
    permission_parts[1] == "*"
    domain := permission_parts[0]
    pattern := sprintf("%s:*", [domain])
    has_permission(pattern)
}

# ---------------------------------------------------------
# Validações de formato e conteúdo
# ---------------------------------------------------------

# Verifica se uma string é um UUID válido
is_valid_uuid(str) {
    # Formato padrão de UUID: 8-4-4-4-12 caracteres hexadecimais
    regex.match("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$", str)
}

# Verifica se um nome de função é válido
is_valid_role_name(name) {
    # Nome deve ter entre 3 e 50 caracteres
    count(name) >= 3
    count(name) <= 50
    
    # Nome deve conter apenas caracteres alfanuméricos, underscores e pontos
    regex.match("^[A-Za-z0-9_.]+$", name)
}

# Verifica se uma descrição é válida
is_valid_description(description) {
    # Descrição deve ter entre 5 e 500 caracteres
    count(description) >= 5
    count(description) <= 500
}

# ---------------------------------------------------------
# Proteções de segurança
# ---------------------------------------------------------

# Verifica se o IP do cliente está em uma lista de redes confiáveis
is_from_trusted_network {
    # Lista de CIDRs confiáveis (exemplo)
    trusted_cidrs := [
        "192.168.0.0/16",  # Rede interna
        "10.0.0.0/8",      # VPN corporativa
        "172.16.0.0/12"    # Redes parceiras
    ]
    
    client_ip := input.context.client_ip
    
    # Verifica se o IP está em alguma das redes confiáveis
    cidr := trusted_cidrs[_]
    net.cidr_contains(cidr, client_ip)
}

# Verifica se a requisição está vindo de um IP suspeito
is_from_suspicious_ip {
    # Lista de CIDRs bloqueados (exemplo)
    blocked_cidrs := [
        "185.143.223.0/24",  # Conhecida por ataques
        "91.234.36.0/24"     # Origem de spam
    ]
    
    client_ip := input.context.client_ip
    
    # Verifica se o IP está em alguma das redes bloqueadas
    cidr := blocked_cidrs[_]
    net.cidr_contains(cidr, client_ip)
}

# Verifica se o acesso está sendo feito durante o horário comercial
is_outside_business_hours {
    # Obtém a hora atual
    current_time := time.now_ns()
    current_hour := time.date(current_time)[3]  # Índice 3 é a hora (0-23)
    current_weekday := time.weekday(current_time)
    
    # Verifica se é final de semana
    current_weekday == "Saturday"
} else {
    current_weekday := time.weekday(time.now_ns())
    current_weekday == "Sunday"
} else {
    # Verifica se está fora do horário comercial (8:00 - 18:00)
    current_hour := time.date(time.now_ns())[3]
    current_hour < 8
} else {
    current_hour := time.date(time.now_ns())[3]
    current_hour > 18
}

# Verifica limites de taxa para prevenção de abusos
is_rate_limited {
    # Esta é uma implementação simplificada para exemplo
    # Em ambiente real, utilizaria uma integração externa para verificação de limites
    
    # Simula um limite excedido para IPs específicos (para demonstração)
    high_frequency_ips := ["203.0.113.1", "198.51.100.2"]
    input.context.client_ip == high_frequency_ips[_]
    
    # Em um sistema real, aqui teríamos uma chamada para verificar se o usuário ou IP 
    # excedeu seu limite de requisições por minuto/hora
}

# ---------------------------------------------------------
# Funções utilitárias
# ---------------------------------------------------------

# Calcula o nível de sensibilidade da operação
operation_sensitivity_level(http_method, resource_type) = "HIGH" {
    # Métodos que modificam recursos críticos são de alta sensibilidade
    http_method == "POST"
    resource_type == "ROLE"
} else = "HIGH" {
    http_method == "DELETE"
} else = "MEDIUM" {
    http_method == "PUT"
} else = "MEDIUM" {
    http_method == "PATCH"
} else = "LOW" {
    http_method == "GET"
}

# Verifica se uma data de expiração é válida (no futuro)
is_valid_expiration_date(expiration) {
    # Converte a data de expiração para timestamp
    expiration_time := time.parse_rfc3339_ns(expiration)
    
    # Obtém o timestamp atual
    now := time.now_ns()
    
    # Verifica se a expiração está no futuro
    expiration_time > now
}

# Verifica se um valor está presente em uma lista
contains(list, value) {
    list[_] == value
}

# Calcula hash para fins de cache e detecção de duplicidade
calculate_hash(obj) = hash {
    json_str := json.marshal(obj)
    hash := crypto.sha256(json_str)
}