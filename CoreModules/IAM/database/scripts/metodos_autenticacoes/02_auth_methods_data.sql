-- ==================================================================================
-- INNOVABIZ - Scripts de Dados para Métodos de Autenticação
-- Versão: 1.0.0
-- Data de Criação: 14/05/2025
-- Autor: INNOVABIZ DevOps
-- Descrição: Inserção dos métodos de autenticação no catálogo do IAM (Fase 1)
-- Regiões Suportadas: UE/Portugal, Brasil, Angola, EUA
-- ==================================================================================

-- Insere os métodos de autenticação baseados no Catálogo de Métodos de Autenticação
-- Os métodos estão organizados por categoria, conforme documentação oficial

-- ==================================================================================
-- 7.1. Métodos Baseados em Conhecimento (Knowledge-Based)
-- ==================================================================================

INSERT INTO iam_core.authentication_methods (
    method_id, method_name, category_id, security_level, irr_value, 
    complexity, maturity, implementation_status, primary_use_cases, 
    description, technical_requirements, security_considerations
) VALUES
-- Senhas e PINs
('KB-01-01', 'Senhas Alfanuméricas', 'KB', 'BASIC', 'R3',
 'LOW', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Geral', 'Web', 'Mobile'],
 'Autenticação por senha alfanumérica tradicional', 
 'Sistema de armazenamento seguro de senhas com hash e salt', 
 'Vulnerável a ataques de força bruta, phishing e reutilização de senhas'),

('KB-01-02', 'Senhas com Requisitos Complexos', 'KB', 'INTERMEDIATE', 'R2',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Empresarial', 'Financeiro', 'Saúde'],
 'Senhas que exigem combinação de caracteres especiais, números, variação de caso', 
 'Sistema de validação de políticas de senha e avaliação de força', 
 'Maior resistência a ataques de força bruta, mas ainda vulnerável a phishing'),

('KB-01-03', 'PINs Numéricos', 'KB', 'BASIC', 'R4',
 'LOW', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Mobile', 'ATM', 'POS'],
 'Códigos numéricos curtos para autenticação', 
 'Suporte a entrada numérica e limitação de tentativas', 
 'Altamente vulnerável a observação, força bruta em implementações ruins'),

('KB-01-04', 'Senhas de Uso Único (OTP Estático)', 'KB', 'INTERMEDIATE', 'R3',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Acesso Remoto', 'Recuperação de Conta'],
 'Senha predefinida para uso único em situação específica', 
 'Geração, distribuição e invalidação após uso', 
 'Risco de interceptação durante distribuição'),

-- Perguntas e Respostas de Segurança
('KB-02-01', 'Perguntas Predefinidas', 'KB', 'BASIC', 'R4',
 'LOW', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Recuperação de Conta', 'Segunda Camada de Verificação'],
 'Conjunto de perguntas de segurança padrão com respostas cadastradas pelo usuário', 
 'Armazenamento seguro de respostas, mecanismo de seleção de pergunta', 
 'Respostas podem ser pesquisáveis em redes sociais ou adivinhadas por conhecidos'),

('KB-02-02', 'Perguntas Personalizadas', 'KB', 'INTERMEDIATE', 'R3',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Recuperação de Conta', 'Segunda Camada de Verificação'],
 'Perguntas de segurança definidas pelo próprio usuário', 
 'Interface para definição e gerenciamento de perguntas e respostas', 
 'Usuários podem criar perguntas fáceis de lembrar mas também fáceis de adivinhar'),

('KB-02-03', 'Verificação de Dados Pessoais', 'KB', 'BASIC', 'R4',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Finanças', 'Governo', 'Healthcare'],
 'Confirmação de informações pessoais já registradas no sistema', 
 'Acesso seguro à base de dados de informações pessoais', 
 'Dados podem ser conhecidos por terceiros ou obtidos em violações de dados'),

-- Padrões Cognitivos
('KB-03-01', 'Reconhecimento de Imagens', 'KB', 'INTERMEDIATE', 'R3',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Web', 'Mobile', 'Acessibilidade'],
 'Seleção de imagens previamente escolhidas ou familiares ao usuário', 
 'Sistema de apresentação de imagens e reconhecimento de seleções', 
 'Vulnerável a ataques de observação física e gravação de tela'),

('KB-03-02', 'Padrões Gráficos', 'KB', 'INTERMEDIATE', 'R3',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Mobile', 'Tablet'],
 'Conexão de pontos em uma grade para formar um padrão específico', 
 'Interface de entrada de padrões gráficos e reconhecimento de traços', 
 'Vulnerável a ataques de smudge e observação por cima do ombro'),

-- ==================================================================================
-- 7.2. Métodos Baseados em Posse (Possession-Based)
-- ==================================================================================

-- Tokens e Dispositivos
('PB-01-01', 'Tokens de Hardware OTP', 'PB', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Corporativo', 'Bancário', 'Acesso VPN'],
 'Dispositivos físicos dedicados que geram códigos de uso único', 
 'Sistemas de validação de OTP, sincronização de tempo ou contador', 
 'Risco de perda ou roubo do dispositivo, mas alta segurança contra ataques remotos'),

('PB-01-02', 'Cartões Inteligentes', 'PB', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Governamental', 'Corporativo', 'Saúde'],
 'Cartões com chips que armazenam credenciais e certificados seguros', 
 'Leitores de cartão inteligente, middleware de integração', 
 'Requer hardware adicional, risco de perda física'),

('PB-01-03', 'Chaves de Segurança USB', 'PB', 'ADVANCED', 'R1',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Corporativo', 'Desenvolvedor', 'Contas de alto valor'],
 'Dispositivos USB que implementam protocolos de autenticação como FIDO2', 
 'Suporte a U2F/WebAuthn, integração com navegadores e sistemas', 
 'Resistente a phishing, mas risco de perda física'),

-- Certificados Digitais
('PB-02-01', 'Certificados X.509 em Dispositivo', 'PB', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Corporativo', 'Governamental', 'PKI'],
 'Certificados digitais instalados em dispositivos do usuário', 
 'Infraestrutura de chaves públicas, validação de certificados', 
 'Depende da segurança do dispositivo onde está instalado'),

('PB-02-02', 'Certificados em Hardware Seguro', 'PB', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Governo', 'Militar', 'Saúde'],
 'Certificados armazenados em HSMs ou TPMs', 
 'Hardware especializado, integrações com HSM/TPM', 
 'Alto custo, alta segurança, gerenciamento complexo'),

-- Autenticadores Mobile
('PB-03-01', 'Aplicativos Mobile Autenticadores', 'PB', 'ADVANCED', 'R2',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Consumidor', 'Negócios', 'SaaS'],
 'Aplicativos em smartphones que geram códigos OTP ou fornecem aprovação', 
 'Apps para iOS e Android, APIs de notificação push', 
 'Depende da segurança do dispositivo mobile'),

('PB-03-02', 'SMS OTP', 'PB', 'INTERMEDIATE', 'R3',
 'LOW', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Recuperação de Conta', 'Verificação', 'Pagamentos'],
 'Códigos enviados via mensagem SMS para o telefone registrado', 
 'Integração com gateway SMS, geração e validação de OTP', 
 'Vulnerável a SIM swapping e interceptação de SMS'),

('PB-03-03', 'SIM-Based Authentication', 'PB', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Telecomunicações', 'Mobile Banking', 'Pagamentos'],
 'Autenticação baseada no módulo SIM do dispositivo móvel', 
 'Integração com operadoras de telefonia, APIs específicas', 
 'Forte ligação ao dispositivo, mas requer cooperação da operadora'),

-- ==================================================================================
-- 7.6. Modos de Autenticação Federada e Single Sign-On
-- ==================================================================================

('FS-01-01', 'OAuth 2.0', 'FS', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['APIs', 'Mobile Apps', 'Web Services'],
 'Protocolo de autorização que permite acesso delegado a recursos protegidos', 
 'Implementação completa de OAuth 2.0 com suporte a diversos flows', 
 'Requer implementação correta para evitar vulnerabilidades como CSRF'),

('FS-01-02', 'OpenID Connect', 'FS', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Login Federado', 'SaaS', 'Ecossistemas Digitais'],
 'Camada de identidade sobre OAuth 2.0 para autenticação federada', 
 'Implementação de protocolo OIDC, suporte a JWT, endpoints obrigatórios', 
 'Complexidade de integração, dependência de provedores externos'),

('FS-01-03', 'SAML 2.0', 'FS', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Enterprise', 'Governamental', 'SSO Corporativo'],
 'Protocolo baseado em XML para troca de autenticação e autorização', 
 'Suporte à assinatura XML, processamento de asserções SAML', 
 'Complexo de implementar, mas robusto para ambientes corporativos'),

-- ==================================================================================
-- 7.13. Autenticação para Telemedicina
-- ==================================================================================

('TM-01-01', 'Verificação Médica Multi-Fator', 'TM', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Teleconsultas', 'Prescrição Eletrônica', 'Telemedicina'],
 'Combinação de fatores biométricos, documentais e contextuais para médicos', 
 'Integração com sistemas de verificação profissional e bases de reguladores', 
 'Verificação HIPAA/GDPR/LGPD, pré-cadastro com validação documental'),

('TM-01-02', 'Autenticação de Paciente para Telemedicina', 'TM', 'ADVANCED', 'R2',
 'MEDIUM', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Teleconsultas', 'Saúde Digital', 'Monitoramento Remoto'],
 'Verificação simplificada para pacientes com balanço entre segurança e acessibilidade', 
 'Interfaces acessíveis, suporte a múltiplos dispositivos, verificação adaptativa', 
 'Conformidade com proteção de dados de saúde, acessibilidade WCAG 2.1 AAA'),

-- ==================================================================================
-- 7.14. Autenticação para AR/VR
-- ==================================================================================

('AR-01-01', 'Autenticação por Gesto Espacial', 'AR', 'ADVANCED', 'R2',
 'HIGH', 'EMERGING', 'IN_PROGRESS', 
 ARRAY['Realidade Aumentada', 'Ambientes Imersivos', 'Sistemas Holográficos'],
 'Reconhecimento de padrões de gestos específicos no espaço tridimensional', 
 'Sensores de movimento, algoritmos de reconhecimento de padrões 3D', 
 'Alta usabilidade em AR, porém requer calibração e pode ser afetado por limitações físicas'),

('AR-01-02', 'Autenticação por Padrão de Olhar', 'AR', 'ADVANCED', 'R2',
 'HIGH', 'EMERGING', 'IN_PROGRESS', 
 ARRAY['Realidade Virtual', 'Headsets AR/VR', 'Interfaces Neurais'],
 'Análise do padrão de movimento dos olhos do usuário em resposta a estímulos visuais', 
 'Rastreamento ocular de alta precisão, algoritmos de análise de padrões de olhar', 
 'Altamente resistente a observação externa, natural para interfaces AR/VR'),

-- ==================================================================================
-- 7.15. Autenticação para Open Banking
-- ==================================================================================

('OB-01-01', 'Autenticação Delegada Bancária', 'OB', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Open Banking', 'PSD2', 'APIs Financeiras'],
 'Delegação de autenticação para a instituição financeira detentora da conta', 
 'Implementação de OAuth 2.0 com fluxos específicos para Open Banking, conformidade PSD2', 
 'Conformidade com regulações regionais, segurança em redirecionamentos'),

('OB-01-02', 'Autenticação SCA para PSD2', 'OB', 'VERY_ADVANCED', 'R1',
 'VERY_HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Pagamentos Fortes', 'Open Banking', 'Serviços Financeiros'],
 'Strong Customer Authentication conforme requisitos PSD2', 
 'Combinação de múltiplos fatores independentes, gestão de isenções', 
 'Requisitos específicos para UE, balanceamento entre UX e segurança'),

-- ==================================================================================
-- 7.16. Autenticação para Open Insurance
-- ==================================================================================

('OI-01-01', 'Autenticação para Consentimento de Dados Securitários', 'OI', 'ADVANCED', 'R2',
 'HIGH', 'EMERGING', 'IN_PROGRESS', 
 ARRAY['Open Insurance', 'Seguros Digitais', 'APIs de Seguros'],
 'Autenticação específica para autorização de compartilhamento de dados de seguros', 
 'Gestão granular de consentimento, rastreabilidade de autorizações', 
 'Conformidade com regulações de seguros específicas por região'),

-- ==================================================================================
-- 7.17. Autenticação para Setor Público
-- ==================================================================================

('GP-01-01', 'Autenticação por Identidade Digital Cidadã', 'GP', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Governo Digital', 'Serviços Públicos', 'Documentos Eletrônicos'],
 'Baseada em identidades digitais oficiais emitidas pelo governo', 
 'Integração com sistemas de ID nacionais, validação de certificados governamentais', 
 'Conformidade com eIDAS na UE, ITI no Brasil, protocolos internacionais'),

('GP-01-02', 'Autenticação por Assinatura Digital Governamental', 'GP', 'VERY_ADVANCED', 'R1',
 'VERY_HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Documentos Oficiais', 'Processos Administrativos', 'Petições Digitais'],
 'Utilização de assinaturas digitais qualificadas ou avançadas reconhecidas pelo estado', 
 'Suporte a padrões ICP específicos por país, validação de cadeias de certificação', 
 'Requisitos regulatórios por país, suporte a múltiplas autoridades certificadoras'),

-- ==================================================================================
-- 7.18. Autenticação para Setor Financeiro
-- ==================================================================================

('FI-01-01', 'Autenticação por Token de Segurança Bancária', 'FI', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Banking', 'Investimentos', 'Transações Financeiras'],
 'Baseada em tokens físicos ou virtuais emitidos por instituições financeiras', 
 'Integração com sistemas proprietários bancários, validação de OTP específica', 
 'Conformidade com requisitos de segurança bancária por região'),

('FI-01-02', 'Autenticação por Validação Multilateral Financeira', 'FI', 'VERY_ADVANCED', 'R1',
 'VERY_HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Transações de Alto Valor', 'Trading', 'Corporate Banking'],
 'Combinando verificações financeiras em múltiplas dimensões e entidades', 
 'APIs bancárias seguras, integração com sistemas de compensação', 
 'Alta complexidade, requisitos regulatórios rigorosos');

-- ==================================================================================
-- 7.19. Autenticação para Saúde
-- ==================================================================================

INSERT INTO iam_core.authentication_methods (
    method_id, method_name, category_id, security_level, irr_value, 
    complexity, maturity, implementation_status, primary_use_cases, 
    description, technical_requirements, security_considerations
) VALUES
('HC-01-01', 'Autenticação Multi-Fator para Profissionais de Saúde', 'HC', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['EHR/EMR', 'Prescrição', 'Laboratórios'],
 'Múltiplos fatores específicos para profissionais de saúde com validação de credenciais', 
 'Integração com bases de dados profissionais, verificação de licenças ativas', 
 'Compliance HIPAA (EUA), GDPR (UE), LGPD (Brasil), certificações específicas de saúde'),

('HC-01-02', 'Autenticação para Acesso a Dados Sensíveis de Saúde', 'HC', 'VERY_ADVANCED', 'R1',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Prontuário Eletrônico', 'Exames', 'Diagnósticos'],
 'Verificação elevada para acesso a dados classificados como sensíveis de saúde', 
 'Controles granulares por tipo de dado, registro de auditoria detalhado', 
 'Verificações contextuais, análise de padrões de acesso, segregação de dados'),

('HC-01-03', 'Autenticação Adaptativa para Sistemas de Saúde', 'HC', 'ADVANCED', 'R2',
 'VERY_HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Hospitais', 'Clínicas', 'Telemedicina'],
 'Ajuste dinâmico do nível de autenticação com base no contexto clínico e criticidade', 
 'Análise de risco em tempo real, integração com sistemas de emergência', 
 'Balanço entre rapidez de acesso em emergências e segurança para acessos rotineiros');

-- ==================================================================================
-- 7.20. Autenticação em Contextos Específicos
-- ==================================================================================

INSERT INTO iam_core.authentication_methods (
    method_id, method_name, category_id, security_level, irr_value, 
    complexity, maturity, implementation_status, primary_use_cases, 
    description, technical_requirements, security_considerations
) VALUES
('CX-01-01', 'Autenticação para Dispositivos IoT', 'CX', 'ADVANCED', 'R2',
 'HIGH', 'EMERGING', 'IN_PROGRESS', 
 ARRAY['IoT', 'Smart Devices', 'Sensores'],
 'Mecanismos específicos para dispositivos com recursos limitados', 
 'Protocolos leves, certificados embarcados, autenticação por hardware', 
 'Gerenciamento de dispositivos em massa, atualizações de segurança remotas'),

('CX-01-02', 'Autenticação para APIs em Microserviços', 'CX', 'ADVANCED', 'R2',
 'HIGH', 'ESTABLISHED', 'IMPLEMENTED', 
 ARRAY['Microserviços', 'APIs Internas', 'Cloud Native'],
 'Autenticação serviço-a-serviço em arquiteturas distribuídas', 
 'Service mesh, rotação automática de credenciais, mTLS', 
 'Complexidade de gerenciamento, encadeamento de autenticação'),

('CX-01-03', 'Autenticação para Edge Computing', 'CX', 'ADVANCED', 'R2',
 'VERY_HIGH', 'EMERGING', 'IN_PROGRESS', 
 ARRAY['Edge', 'Fog Computing', 'Processamento Distribuído'],
 'Autenticação em ambientes de borda com conectividade intermitente', 
 'Cache de credenciais, sincronização assíncrona, verificação local-primeiro', 
 'Balanceamento entre disponibilidade e segurança, operação offline segura');

-- ==================================================================================
-- Fim do Script
-- ==================================================================================
