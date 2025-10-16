# Catálogo Completo: 300+ Métodos de Autenticação INNOVABIZ

![Status](https://img.shields.io/badge/Status-Oficial-success)
![Versão](https://img.shields.io/badge/Versão-2.0.0-blue)
![Módulo](https://img.shields.io/badge/Módulo-IAM-orange)

**Autor:** INNOVABIZ DevOps  
**Data:** 14 de Maio de 2025  
**Classificação:** Técnica / Referência  

## Visão Geral

Este documento apresenta o catálogo completo dos mais de 300 métodos de autenticação suportados pela plataforma INNOVABIZ. Os métodos estão organizados em categorias lógicas, baseadas em tecnologias, casos de uso e indústrias específicas, com considerações regulatórias apropriadas para as regiões prioritárias (UE/Portugal, Brasil, Angola, EUA).

O catálogo serve como referência técnica para implementadores, arquitetos e gestores de segurança que necessitam selecionar os métodos mais apropriados para seus contextos específicos.

## Estrutura de Classificação

Cada método de autenticação é classificado de acordo com:

**• Categoria:** Agrupamento principal baseado na tecnologia ou paradigma  
**• Nível de Segurança:** Básico, Intermediário, Avançado, Muito Avançado, Crítico  
**• IRR (Índice de Risco Residual):** R1 (Muito Alto) a R5 (Muito Baixo)  
**• Complexidade:** Baixa, Média, Alta, Muito Alta  
**• Maturidade:** Experimental, Emergente, Estabelecida  
**• Status de Implementação:**  
  - Implementado e Disponível  
  - Em Implementação  
  - Planejado (Roadmap)  
  - Em Pesquisa/Avaliação  

## Métodos de Autenticação por Categoria

### 7.1. Métodos Baseados em Conhecimento

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| KB-01-01 | Senha Tradicional | Básico | R1 | Baixa | Estabelecida | | Geral, Legacy |
| KB-01-02 | PIN Numérico | Básico | R1 | Baixa | Estabelecida | | Mobile, ATM |
| KB-01-03 | Padrão Gráfico | Básico | R1 | Baixa | Estabelecida | | Mobile, Tablets |
| KB-01-04 | Perguntas de Segurança | Básico | R1 | Baixa | Estabelecida | | Recuperação, Legacy |
| KB-01-05 | Senha Única (OTP) | Intermediário | R2 | Média | Estabelecida | | Geral, Segunda Camada |
| KB-01-06 | Verificação de Conhecimento | Intermediário | R2 | Média | Estabelecida | | Finanças, Bancos |
| KB-01-07 | Passphrase | Intermediário | R2 | Média | Estabelecida | | Alta Segurança, Criptografia |
| KB-01-08 | Senha com Requisitos Complexos | Intermediário | R2 | Média | Estabelecida | | Enterprise, Geral |
| KB-01-09 | Imagem Secreta | Básico | R1 | Baixa | Estabelecida | | Anti-phishing, Bancos |
| KB-01-10 | Senha de Uso Único | Intermediário | R2 | Média | Estabelecida | | Temporário, Emergencial |
| KB-01-11 | Senhas Sem Conexão | Intermediário | R2 | Média | Estabelecida | | Ambientes Isolados |
| KB-01-12 | Gestos Customizados | Intermediário | R2 | Média | Estabelecida | | Mobile, Tablets |
| KB-01-13 | Sequência de Ações | Intermediário | R2 | Média | Emergente | | Interfaces Avançadas |
| KB-01-14 | Localização em Imagem | Intermediário | R2 | Média | Emergente | | Segunda Camada |
| KB-01-15 | PIN Expandido | Intermediário | R2 | Média | Estabelecida | | Sistemas de Alta Segurança |
| KB-01-16 | Rotação de Caracteres | Intermediário | R2 | Média | Emergente | | Anti-keylogging |
| KB-01-17 | Teclado Virtual Randomizado | Intermediário | R2 | Média | Estabelecida | | Banking, Anti-keylogging |
| KB-01-18 | Matriz de Autenticação | Intermediário | R2 | Média | Estabelecida | | Finanças, Segunda Camada |
| KB-01-19 | Desafio-Resposta Baseado em Dados | Intermediário | R2 | Alta | Estabelecida | | Serviços Financeiros |
| KB-01-20 | Senha Dividida Multi-canal | Avançado | R3 | Alta | Emergente | | Alta Segurança, Governo |

### 7.2. Métodos Baseados em Posse

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| PB-02-01 | Aplicativo Autenticador | Avançado | R3 | Média | Estabelecida | | Geral, Enterprise |
| PB-02-02 | SMS OTP | Intermediário | R2 | Baixa | Estabelecida | | Consumidor, Legacy |
| PB-02-03 | Email OTP | Intermediário | R2 | Baixa | Estabelecida | | Consumidor, Recuperação |
| PB-02-04 | Token Físico | Avançado | R3 | Média | Estabelecida | | Enterprise, Banking |
| PB-02-05 | Cartão Inteligente | Avançado | R4 | Alta | Estabelecida | | Governamental, Enterprise |
| PB-02-06 | FIDO2/WebAuthn | Muito Avançado | R4 | Alta | Estabelecida | | Enterprise, Cross-Platform |
| PB-02-07 | Push Notification | Avançado | R3 | Média | Estabelecida | | Mobile, Enterprise |
| PB-02-08 | Certificado Digital | Avançado | R3 | Alta | Estabelecida | | Governamental, Enterprise |
| PB-02-09 | Autenticação por Bluetooth | Avançado | R3 | Média | Estabelecida | | IoT, Proximidade |
| PB-02-10 | NFC Authentication | Avançado | R3 | Média | Estabelecida | | Mobile, Pagamentos |
| PB-02-11 | QR Code Dinâmico | Intermediário | R2 | Média | Estabelecida | | Mobile, Cross-Device |
| PB-02-12 | Token Virtual | Avançado | R3 | Média | Estabelecida | | Cloud, Enterprise |
| PB-02-13 | Secure Element Hardware | Muito Avançado | R4 | Alta | Estabelecida | | Mobile, Pagamentos |
| PB-02-14 | Cartão OTP | Intermediário | R2 | Baixa | Estabelecida | | Banking, Legacy |
| PB-02-15 | Proximidade de Dispositivo | Intermediário | R2 | Média | Emergente | | Proximidade, Escritório |
| PB-02-16 | Autenticação por Rádio | Avançado | R3 | Alta | Estabelecida | | Industrial, Específico |
| PB-02-17 | Validação de SIM/IMEI | Avançado | R3 | Alta | Estabelecida | | Telecomunicações, Banking |
| PB-02-18 | Validação de Endpoint | Avançado | R3 | Alta | Estabelecida | | Enterprise, VPN |
| PB-02-19 | Assinatura com TEE | Muito Avançado | R4 | Alta | Emergente | | Mobile Securizado |
| PB-02-20 | YubiKey e Hardware Similar | Muito Avançado | R4 | Média | Estabelecida | | Alta Segurança, Enterprise |

### 7.3. Autenticação Anti-Fraude e Comportamental

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| AF-03-01 | Análise de Comportamento do Usuário | Avançado | R3 | Alta | Emergente | | Finanças, E-commerce |
| AF-03-02 | Detecção de Bot/Automação | Avançado | R3 | Alta | Estabelecida | | Web Generalista |
| AF-03-03 | Análise de Padrão de Digitação | Intermediário | R2 | Média | Estabelecida | | Finanças, Enterprise |
| AF-03-04 | Posicionamento do Mouse | Básico | R1 | Média | Estabelecida | | Web, E-commerce |
| AF-03-05 | Reconhecimento de Estilo de Escrita | Intermediário | R2 | Alta | Emergente | | Educacional, Criativo |
| AF-03-06 | Gestos em Tela Touchscreen | Intermediário | R2 | Média | Estabelecida | | Mobile |
| AF-03-07 | Padrão de Uso de Aplicativo | Básico | R1 | Média | Emergente | | Consumer, Fraude |
| AF-03-08 | Padrão de Interação com Interface | Básico | R1 | Média | Emergente | | Fraude, UX |
| AF-03-09 | Análise de Comunicação (Linguística) | Intermediário | R2 | Alta | Emergente | | Verificação de Autor |
| AF-03-10 | Análise de Navegação Web | Intermediário | R2 | Alta | Estabelecida | | E-commerce, Fraude |
| AF-03-11 | Detecção de Jailbreak/Root | Intermediário | R2 | Média | Estabelecida | | Banking, Corporativo |
| AF-03-12 | Análise de Toque (Pressão/Velocidade) | Intermediário | R2 | Alta | Emergente | | Mobile Banking |
| AF-03-13 | Análise de Postura em Dispositivos | Avançado | R3 | Alta | Emergente | | Mobile Seguro |
| AF-03-14 | Detecção de Emuladores | Avançado | R3 | Alta | Estabelecida | | Banking, Gaming |
| AF-03-15 | Detecção de Ataques de Replay | Avançado | R3 | Alta | Estabelecida | | Finanças, Crítico |
| AF-03-16 | Análise Temporal de Transações | Avançado | R3 | Alta | Estabelecida | | Banking, Fraude |
| AF-03-17 | Machine Learning Comportamental | Muito Avançado | R4 | Alta | Emergente | | Banking, Fraude |
| AF-03-18 | Análise de Cohort/Peer Group | Avançado | R3 | Alta | Emergente | | Financeiro, Fraude |
| AF-03-19 | Detecção de Phishing Comportamental | Avançado | R3 | Alta | Emergente | | Enterprise, Finanças |
| AF-03-20 | Análise de Intenção Fraudulenta | Avançado | R3 | Alta | Emergente | | Finanças, E-commerce |

### 7.4. Autenticação Biométrica

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| BM-04-01 | Impressão Digital | Avançado | R3 | Média | Estabelecida | | Mobile, Acesso Físico |
| BM-04-02 | Reconhecimento Facial | Avançado | R3 | Alta | Estabelecida | | Mobile, Governamental |
| BM-04-03 | Reconhecimento de Íris | Muito Avançado | R4 | Alta | Estabelecida | | Alta Segurança, Fronteiras |
| BM-04-04 | Reconhecimento de Voz | Avançado | R3 | Alta | Estabelecida | | Call Centers, Consumidor |
| BM-04-05 | Escaneamento de Retina | Muito Avançado | R4 | Alta | Estabelecida | | Militar, Alta Segurança |
| BM-04-06 | Reconhecimento Vascular | Muito Avançado | R4 | Alta | Estabelecida | | Banking, Acesso Físico |
| BM-04-07 | Geometria da Mão | Avançado | R3 | Alta | Estabelecida | | Acesso Físico, Industrial |
| BM-04-08 | Dinâmica de Assinatura | Avançado | R3 | Alta | Estabelecida | | Jurídico, Financeiro |
| BM-04-09 | Batimento Cardíaco | Avançado | R3 | Alta | Emergente | | Saúde, Wearables |
| BM-04-10 | Reconhecimento de Marcha | Intermediário | R2 | Alta | Emergente | | Vigilância, Segurança |
| BM-04-11 | EEG (Eletroencefalograma) | Crítico | R5 | Muito Alta | Experimental | | Pesquisa, Ultra-Seguro |
| BM-04-12 | Análise de DNA Rápida | Crítico | R5 | Muito Alta | Experimental | | Forense, Governo |
| BM-04-13 | Reconhecimento de Orelha | Intermediário | R2 | Alta | Emergente | | Aplicações Especializadas |
| BM-04-14 | Leitura Térmica Facial | Avançado | R3 | Alta | Emergente | | Segurança Anti-Spoofing |
| BM-04-15 | Leitura de Impressão Palmar | Intermediário | R2 | Alta | Emergente | | Acesso Físico |
| BM-04-16 | Multiespectral (Combinação de Biometrias) | Avançado | R3 | Alta | Emergente | | Alta Segurança |
| BM-04-17 | Reconhecimento Facial 3D | Muito Avançado | R4 | Alta | Estabelecida | | Mobile Premium, Governamental |
| BM-04-18 | Reconhecimento Labial | Avançado | R3 | Alta | Emergente | | Complementar, Multimodal |
| BM-04-19 | Odor Corporal | Avançado | R3 | Muito Alta | Experimental | | Segurança Física Especializada |
| BM-04-20 | Pulsação Vascular | Muito Avançado | R4 | Alta | Emergente | | Saúde, Alta Segurança |

### 7.5. Dispositivos e Tokens de Segurança

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| DT-05-01 | YubiKey/Chaves FIDO2 | Muito Avançado | R4 | Média | Estabelecida | | Enterprise, Governamental |
| DT-05-02 | RSA SecurID | Avançado | R3 | Média | Estabelecida | | Corporativo, Legacy |
| DT-05-03 | Google Titan Security Key | Muito Avançado | R4 | Média | Estabelecida | | Enterprise, Cloud |
| DT-05-04 | Cartão Criptográfico Nacional | Muito Avançado | R4 | Alta | Estabelecida | | Governamental, Serviços Públicos |
| DT-05-05 | Dispositivo Mobile como Token | Avançado | R3 | Média | Estabelecida | | Enterprise, Banking |
| DT-05-06 | Cartão de Acesso com Visor E-ink | Avançado | R3 | Alta | Emergente | | Corporativo, Governamental |
| DT-05-07 | Impressora de Códigos OTP | Intermediário | R2 | Média | Estabelecida | | Corporativo, Finanças |
| DT-05-08 | Token de Segurança Bluetooth | Avançado | R3 | Alta | Estabelecida | | Mobile Enterprise |
| DT-05-09 | Token Biométrico Isolado | Muito Avançado | R4 | Alta | Emergente | | Militar, Alta Segurança |
| DT-05-10 | Secure Element Mobile | Avançado | R3 | Alta | Estabelecida | | Finanças, Mobile |
| DT-05-11 | Token Virtual em TPM | Avançado | R3 | Alta | Estabelecida | | Enterprise, Endpoints |
| DT-05-12 | Smart Card com Biometria | Muito Avançado | R4 | Alta | Estabelecida | | Governamental, Crítico |
| DT-05-13 | Token Hardware Descartável | Intermediário | R2 | Média | Estabelecida | | Acesso Temporário |
| DT-05-14 | Dongle de Autenticação USB | Avançado | R3 | Média | Estabelecida | | Enterprise, Educacional |
| DT-05-15 | Token em Secure Enclave | Avançado | R3 | Alta | Estabelecida | | Mobile iOS, Enterprise |
| DT-05-16 | Hardware Security Module (HSM) | Muito Avançado | R5 | Alta | Estabelecida | | Finanças, Infraestrutura |
| DT-05-17 | Cartão Inteligente com eInk | Avançado | R3 | Alta | Emergente | | Governamental, Corporativo |
| DT-05-18 | Token Físico Multi-chave | Muito Avançado | R4 | Alta | Emergente | | Cripto, Alta Segurança |
| DT-05-19 | Autenticador de Voz Dedicado | Avançado | R3 | Alta | Emergente | | Call Centers, SAC |
| DT-05-20 | Token com Display e Teclado | Avançado | R3 | Média | Estabelecida | | Banking, Enterprise |

### 7.6. Modos de Autenticação Federada e Single Sign-On

### 7.13. Métodos de Autenticação para Telemedicina

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| TM-13-01 | Autenticação Multi-fator para Telemedicina | Muito Avançado | R4 | Alta | Estabelecida | | Consultas Médicas Virtuais |
| TM-13-02 | Verificação de Identidade do Profissional de Saúde | Muito Avançado | R4 | Alta | Estabelecida | | Telemedicina Certificada |
| TM-13-03 | Biometria Vocal para Autorização Médica | Avançado | R3 | Alta | Emergente | | Prescrição Remota |
| TM-13-04 | Reconhecimento Facial para Pacientes | Avançado | R3 | Média | Estabelecida | | Identificação de Pacientes |
| TM-13-05 | Assinatura Digital Médica Qualificada | Muito Avançado | R5 | Alta | Estabelecida | | Prescrições, Prontuários |
| TM-13-06 | Autenticação por Carteira de Saúde Digital | Avançado | R3 | Alta | Emergente | | Identificação Integrada |
| TM-13-07 | Verificação Contextual de Dispositivos Médicos | Avançado | R3 | Alta | Emergente | | IoMT (Internet of Medical Things) |
| TM-13-08 | Validação de Credenciais Médicas em Tempo Real | Muito Avançado | R4 | Alta | Emergente | | Segunda Opinião Médica |
| TM-13-09 | Autenticação Federada para Sistemas de Saúde | Avançado | R3 | Alta | Estabelecida | | Interoperabilidade Médica |
| TM-13-10 | Token Físico Médico Especializado | Muito Avançado | R4 | Alta | Emergente | | Acesso a Dados Sensíveis |
| TM-13-11 | Autenticação Adaptativa Baseada em Risco Médico | Avançado | R3 | Alta | Emergente | | Acesso a Prontuários Críticos |
| TM-13-12 | Push Notification com Verificação Biométrica | Avançado | R3 | Média | Estabelecida | | Apps de Telemedicina |
| TM-13-13 | Validação de Localização para Serviços Médicos | Avançado | R3 | Alta | Emergente | | Compliance Regional |
| TM-13-14 | Autenticação para Emergências Médicas | Avançado | R4 | Alta | Emergente | | Acesso Rápido Emergencial |
| TM-13-15 | Proxy de Autenticação para Pacientes Vulneráveis | Avançado | R3 | Alta | Emergente | | Idosos, Pacientes Incapacitados |
| TM-13-16 | Delegação de Autenticação para Responsáveis | Avançado | R3 | Alta | Estabelecida | | Pediatria, Cuidadores |
| TM-13-17 | Validação Cruzada de Planos de Saúde | Intermediário | R2 | Média | Estabelecida | | Autorização de Atendimento |
| TM-13-18 | Autenticação em Dispositivos Wearable Médicos | Avançado | R3 | Alta | Emergente | | Monitoramento Contínuo |
| TM-13-19 | Verificação de Identidade com HIPAA/GDPR/LGPD | Muito Avançado | R4 | Alta | Estabelecida | | Compliance Regulatório |
| TM-13-20 | Registro Biométrico para Ensaios Clínicos | Muito Avançado | R4 | Alta | Emergente | | Pesquisa Médica |

### 7.14. Métodos de Autenticação para Realidade Aumentada e Virtual (AR/VR)

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| AR-14-01 | Autenticação por Gestos Espaciais | Avançado | R3 | Alta | Emergente | | AR/VR Enterprise |
| AR-14-02 | Reconhecimento de Padrão de Olhar | Avançado | R3 | Alta | Emergente | | AR Premium |
| AR-14-03 | Autenticação Baseada em Ambiente 3D | Avançado | R3 | Alta | Emergente | | VR Corporativo |
| AR-14-04 | Validação Biométrica Governamental | Muito Avançado | R4 | Alta | Emergente | | AR Segurança Crítica |
| AR-14-05 | Token Virtual em Ambiente 3D | Avançado | R3 | Alta | Emergente | | Metaverso Corporativo |
| AR-14-06 | Reconhecimento de Movimento Corporal | Avançado | R3 | Alta | Emergente | | VR Experiências Imersivas |
| AR-14-07 | Padrão de Interação com Objetos Virtuais | Avançado | R3 | Alta | Emergente | | AR/VR Gaming |
| AR-14-08 | Validação Contextual de Sessão AR/VR | Avançado | R3 | Alta | Emergente | | AR/VR Multi-usuário |
| AR-14-09 | Reconhecimento Vocal Espacial | Avançado | R3 | Alta | Emergente | | Comandos em AR/VR |
| AR-14-10 | Autenticação por Avatar Persistente | Intermediário | R2 | Média | Emergente | | Metaverso Social |
| AR-14-11 | Certificação de Hardware AR/VR | Avançado | R3 | Alta | Emergente | | Enterprise, Governamental |
| AR-14-12 | Token de Mapeamento Espacial | Avançado | R3 | Alta | Emergente | | AR Industrial |
| AR-14-13 | Verificação por Zonas de Privacidade Virtuais | Avançado | R3 | Alta | Emergente | | AR em Espaços Públicos |
| AR-14-14 | Controle de Acesso Contextual em AR | Avançado | R3 | Alta | Emergente | | AR Corporativo |
| AR-14-15 | Autenticação Multimodal em Ambiente Virtual | Muito Avançado | R4 | Alta | Emergente | | VR Premium |
| AR-14-16 | Detecção de Presença com Marcadores Espaciais | Avançado | R3 | Alta | Emergente | | AR Localizado |
| AR-14-17 | Verificação Anti-Spoofing em AR | Muito Avançado | R4 | Alta | Emergente | | AR Financeiro |
| AR-14-18 | Delegação de Identidade em Ambientes Virtuais | Avançado | R3 | Alta | Emergente | | VR Colaborativo |
| AR-14-19 | Criptografia de Dados Espaciais | Muito Avançado | R4 | Alta | Emergente | | AR/VR Confidencial |
| AR-14-20 | Autenticação por Percepção Sensorial | Avançado | R3 | Muito Alta | Experimental | | AR/VR Avançado |

### 7.15. Métodos de Autenticação para Open Banking

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| OB-15-01 | Autenticação OAuth 2.0 com FAPI | Muito Avançado | R4 | Alta | Estabelecida | | APIs Bancárias |
| OB-15-02 | Redirectionamento Seguro com CIBA | Muito Avançado | R4 | Alta | Emergente | | Experiência Sem Redirecionamento |
| OB-15-03 | Autenticação Decoupled (App-to-App) | Avançado | R3 | Alta | Estabelecida | | Banking Apps |
| OB-15-04 | Validação com Certificado eIDAS QTSP | Muito Avançado | R5 | Alta | Estabelecida | | Open Banking UE |
| OB-15-05 | MFA Adaptativo para Transações Financeiras | Muito Avançado | R4 | Alta | Estabelecida | | Pagamentos, Transferências |
| OB-15-06 | Confirmação de Token Binding | Muito Avançado | R4 | Alta | Emergente | | Prevenção de Roubo de Token |
| OB-15-07 | Autenticação SCA Compliant | Muito Avançado | R4 | Alta | Estabelecida | | PSD2, Open Banking UE |
| OB-15-08 | JWT Assinado para Open Banking | Muito Avançado | R4 | Alta | Estabelecida | | APIs Financeiras |
| OB-15-09 | Biometria com Validação Bancária | Muito Avançado | R4 | Alta | Estabelecida | | Acesso a Contas |
| OB-15-10 | Autorização por Credencial Verificável | Avançado | R3 | Alta | Emergente | | Verificação KYC/AML |
| OB-15-11 | Autenticação com Certificado ICP-Brasil | Muito Avançado | R5 | Alta | Estabelecida | | Open Banking Brasil |
| OB-15-12 | mTLS para APIs Financeiras | Muito Avançado | R5 | Alta | Estabelecida | | Machine-to-Machine Financeiro |
| OB-15-13 | Autenticação por Conta Bancária Validada | Avançado | R3 | Média | Estabelecida | | Verificação de Titular |
| OB-15-14 | Validação Multi-Instituição Financeira | Avançado | R3 | Alta | Emergente | | Consolidação Financeira |
| OB-15-15 | Verificação de Dispositivo Confiável | Avançado | R3 | Alta | Estabelecida | | Mobile Banking |
| OB-15-16 | Autorização Financeira Baseada em Risco | Muito Avançado | R4 | Alta | Estabelecida | | Transações Financeiras |
| OB-15-17 | Autenticação com Validação de Regulatory ID | Muito Avançado | R4 | Alta | Estabelecida | | Conformidade Financeira |
| OB-15-18 | Tokenização de Credenciais Bancárias | Muito Avançado | R4 | Alta | Estabelecida | | Compartilhamento de Dados |
| OB-15-19 | Confirmação de Pagamento em Dois Canais | Muito Avançado | R5 | Alta | Estabelecida | | Transações de Alto Valor |
| OB-15-20 | Federação de Identidade Financeira | Muito Avançado | R4 | Alta | Emergente | | Ecossistema Open Finance |

### 7.16. Métodos de Autenticação para Open Insurance

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| OI-16-01 | Autenticação Federada de Seguros | Muito Avançado | R4 | Alta | Emergente | | Corretoras, Seguradoras |
| OI-16-02 | Verificação KYC/AML para Seguros | Muito Avançado | R4 | Alta | Estabelecida | | Contratação de Seguros |
| OI-16-03 | Confirmação Biométrica para Sinistros | Avançado | R3 | Alta | Emergente | | Processamento de Sinistros |
| OI-16-04 | Tokenização de Apólices | Avançado | R3 | Alta | Emergente | | Compartilhamento Seguro |
| OI-16-05 | Verificação Documental com IA | Avançado | R3 | Alta | Emergente | | Validação de Documentos |
| OI-16-06 | Autenticação OAuth com Escopo Seguros | Muito Avançado | R4 | Alta | Estabelecida | | APIs de Seguros |
| OI-16-07 | Autenticação Baseada em Dispositivo IoT | Avançado | R3 | Alta | Emergente | | Seguros Telemáticos |
| OI-16-08 | Validação de Identidade em Vídeo | Avançado | R3 | Alta | Emergente | | Contratação Remota |
| OI-16-09 | Verificação Cruzada de Seguros | Avançado | R3 | Alta | Emergente | | Prevenção de Fraudes |
| OI-16-10 | Prova de Propriedade Digital | Avançado | R3 | Alta | Emergente | | Seguros Patrimoniais |
| OI-16-11 | Autenticação Certificada SCA | Muito Avançado | R4 | Alta | Estabelecida | | Pagamentos de Prêmios |
| OI-16-12 | Verificação de Localização Segura | Avançado | R3 | Alta | Emergente | | Geolocalização para Sinistros |
| OI-16-13 | Autenticação Adaptativa de Seguros | Avançado | R3 | Alta | Emergente | | Acesso a Plataformas |
| OI-16-14 | Verificação de Identidade eIDAS | Muito Avançado | R5 | Alta | Estabelecida | | Open Insurance UE |
| OI-16-15 | Validação Facial para Vida/Saúde | Avançado | R3 | Alta | Emergente | | Seguros de Vida/Saúde |
| OI-16-16 | Autenticação por Código QR Dinâmico | Intermediário | R2 | Média | Estabelecida | | Apólices Digitais |
| OI-16-17 | Credenciais Verificáveis de Seguros | Avançado | R3 | Alta | Emergente | | Ecossistema de Seguros |
| OI-16-18 | Autenticação Omnicanal de Seguros | Avançado | R3 | Alta | Emergente | | Multi-plataforma |
| OI-16-19 | Verificação de Assinatura Qualificada | Muito Avançado | R5 | Alta | Estabelecida | | Contratos de Seguros |
| OI-16-20 | Confirmação de Identidade em Chamada | Avançado | R3 | Média | Estabelecida | | Call Centers de Seguros |

### 7.17. Métodos de Autenticação para Setor Público/Governamental

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| GP-17-01 | Autenticação por Identidade Digital Cidadã | Muito Avançado | R5 | Alta | Estabelecida | | eGov, Serviços Públicos |
| GP-17-02 | Assinatura Digital Governamental | Muito Avançado | R5 | Alta | Estabelecida | | Documentos Oficiais |
| GP-17-03 | Autenticação por Portal Governamental Federado | Muito Avançado | R4 | Alta | Estabelecida | | Serviços Públicos Integrados |
| GP-17-04 | Verificação de Identidade Presencial + Digital | Muito Avançado | R5 | Alta | Estabelecida | | Serviços de Alto Valor |
| GP-17-05 | Autenticação com Cartão de Cidadão | Muito Avançado | R5 | Alta | Estabelecida | | Governo (UE/Portugal) |
| GP-17-06 | Login Gov/Login Único | Muito Avançado | R4 | Alta | Estabelecida | | Governo (Brasil) |
| GP-17-07 | Autenticação por Certificado ICP-Gov | Muito Avançado | R5 | Alta | Estabelecida | | Documentos Públicos |
| GP-17-08 | Validação Biométrica Governamental | Muito Avançado | R5 | Alta | Estabelecida | | Fronteiras, Identificação |
| GP-17-09 | Autenticação Setorial Governamental | Avançado | R4 | Alta | Estabelecida | | Departamentos Específicos |
| GP-17-10 | Acesso Único Cidadão | Avançado | R4 | Alta | Estabelecida | | Portal de Serviços Unificado |
| GP-17-11 | Verificação por Base de Dados Civil | Muito Avançado | R5 | Alta | Estabelecida | | Validação Civil |
| GP-17-12 | Validação por Documento de Identidade | Avançado | R4 | Alta | Estabelecida | | Verificação Documental |
| GP-17-13 | Autenticação para Processos Judiciais | Muito Avançado | R5 | Alta | Estabelecida | | Sistemas Judiciais |
| GP-17-14 | Validação para Votação Eletrônica | Muito Avançado | R5 | Alta | Emergente | | Eleições, Referendos |
| GP-17-15 | Autenticação para Serviços Críticos | Muito Avançado | R5 | Alta | Estabelecida | | Infraestrutura Nacional |
| GP-17-16 | Verificação de Atributos Governamentais | Avançado | R4 | Alta | Estabelecida | | Permissões, Licenças |
| GP-17-17 | Autenticação por Poder Público | Muito Avançado | R5 | Alta | Estabelecida | | Acesso a Sistemas Internos |
| GP-17-18 | Verificação de Profissionais Regulados | Avançado | R4 | Alta | Estabelecida | | Registro Profissional |
| GP-17-19 | Autenticação por Autoridade Certificadora | Muito Avançado | R5 | Alta | Estabelecida | | Certificados Governamentais |
| GP-17-20 | Validação por Dados Cadastrais Oficiais | Avançado | R4 | Alta | Estabelecida | | Serviços Municipais |

### 7.18. Métodos de Autenticação para Setor Financeiro

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| FS-18-01 | Autenticação por Token de Segurança Bancária | Muito Avançado | R5 | Alta | Estabelecida | | Internet Banking |
| FS-18-02 | Validação Multilateral Financeira | Muito Avançado | R5 | Alta | Estabelecida | | Transações de Alto Valor |
| FS-18-03 | Confirmação em Aplicativo Financeiro | Avançado | R4 | Média | Estabelecida | | Mobile Banking |
| FS-18-04 | Verificação Biométrica para Transações | Muito Avançado | R4 | Alta | Estabelecida | | Aprovação de Pagamentos |
| FS-18-05 | Autenticação Baseada em Contexto | Avançado | R3 | Alta | Emergente | | Prevenção de Fraudes |
| FS-18-06 | Validação por Chave de Segurança | Muito Avançado | R5 | Alta | Estabelecida | | Operações Críticas |
| FS-18-07 | Autenticação Dinâmica com Senha Única | Avançado | R4 | Média | Estabelecida | | Transações Eletrônicas |
| FS-18-08 | Validação por Canal Secundário | Avançado | R4 | Média | Estabelecida | | Confirmação Adicional |
| FS-18-09 | Biometria Comportamental Financeira | Avançado | R3 | Alta | Emergente | | Prevenção Antifraude |
| FS-18-10 | MFA Adaptativo para Serviços Financeiros | Muito Avançado | R4 | Alta | Estabelecida | | Banking, Investimentos |
| FS-18-11 | Validação de Dispositivo de Confiança | Avançado | R3 | Média | Estabelecida | | Mobile Banking |
| FS-18-12 | Verificação KYC/AML Avançada | Muito Avançado | R5 | Alta | Estabelecida | | Onboarding Financeiro |
| FS-18-13 | Autenticação por Cartão EMV | Muito Avançado | R5 | Alta | Estabelecida | | Pagamentos Presenciais |
| FS-18-14 | Validação de Identidade por Videoconferência | Avançado | R4 | Alta | Estabelecida | | Abertura de Contas |
| FS-18-15 | Autenticação Forte com SCA | Muito Avançado | R5 | Alta | Estabelecida | | PSD2, Regulamentação |
| FS-18-16 | Verificação com Assinatura Digital Bancária | Muito Avançado | R5 | Alta | Estabelecida | | Contratos Financeiros |
| FS-18-17 | Autenticação para Banking Corporativo | Muito Avançado | R5 | Alta | Estabelecida | | Banking Empresarial |
| FS-18-18 | Validação por Token Hardware Específico | Muito Avançado | R5 | Alta | Estabelecida | | Transações Corporativas |
| FS-18-19 | Autenticação Certificada PCI-DSS | Muito Avançado | R5 | Alta | Estabelecida | | Pagamentos com Cartão |
| FS-18-20 | Autorização por Análise de Risco em Tempo Real | Muito Avançado | R4 | Alta | Emergente | | Prevenção Proativa |

### 7.19. Métodos de Autenticação para Telecomunicações

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| TC-19-01 | Autenticação por SIM Card | Avançado | R3 | Média | Estabelecida | | Serviços Móveis |
| TC-19-02 | Validação de IMEI/IMSI | Avançado | R3 | Média | Estabelecida | | Identificação de Dispositivo |
| TC-19-03 | Autenticação da Rede Móvel | Muito Avançado | R4 | Alta | Estabelecida | | Redes 4G/5G |
| TC-19-04 | eSIM Authentication | Muito Avançado | R4 | Alta | Emergente | | Dispositivos IoT, Wearables |
| TC-19-05 | Autenticação por SS7 Seguro | Muito Avançado | R5 | Alta | Estabelecida | | Infraestrutura de Telecom |
| TC-19-06 | SMS com TOTP | Intermediário | R2 | Baixa | Estabelecida | | Consumer Services |
| TC-19-07 | Verificação de Número por Chamada | Intermediário | R2 | Baixa | Estabelecida | | Verificação de Conta |
| TC-19-08 | Validação de Assinante em App | Avançado | R3 | Média | Estabelecida | | Apps de Operadora |
| TC-19-09 | Autenticação Telco OAuth | Avançado | R3 | Alta | Estabelecida | | Serviços Digitais Telco |
| TC-19-10 | Autenticação por USSD | Intermediário | R2 | Baixa | Estabelecida | | Mercados Emergentes |
| TC-19-11 | Validação de Portabilidade | Avançado | R3 | Alta | Estabelecida | | Verificação de Operadora |
| TC-19-12 | Autenticação em Cell Broadcast | Avançado | R3 | Alta | Estabelecida | | Alertas de Emergência |
| TC-19-13 | Verificação por Geolocalização Celular | Avançado | R3 | Alta | Estabelecida | | Serviços Baseados em Localização |
| TC-19-14 | OpenID Connect para Telecomunicações | Avançado | R3 | Alta | Emergente | | APIs de Telecom |
| TC-19-15 | Autenticação Federada entre Operadoras | Avançado | R3 | Alta | Emergente | | Roaming, Serviços Compartilhados |
| TC-19-16 | Header Enrichment Authentication | Avançado | R3 | Alta | Emergente | | Serviços Mobile |
| TC-19-17 | Validação de SIM OTA | Avançado | R3 | Alta | Estabelecida | | Atualizações Remotas |
| TC-19-18 | Network Function Virtualization Auth | Muito Avançado | R4 | Alta | Emergente | | Redes 5G, Edge |
| TC-19-19 | Autenticação para IPX/GRX | Muito Avançado | R5 | Alta | Estabelecida | | Interconexão Global |
| TC-19-20 | Mutual Authentication LTE/5G | Muito Avançado | R5 | Alta | Estabelecida | | Segurança de Rede Móvel |

### 7.20. Métodos de Autenticação para Multimídia e Entretenimento

| ID | Método | Nível de Segurança | IRR | Complexidade | Maturidade | Status | Casos de Uso Primários |
|----|--------|---------------------|-----|--------------|------------|--------|------------------------|
| MM-20-01 | Autenticação para Streaming | Intermediário | R2 | Média | Estabelecida | | Plataformas de Streaming |
| MM-20-02 | Verificação de Assinatura Digital | Avançado | R3 | Alta | Estabelecida | | Conteúdo Premium |
| MM-20-03 | Autenticação para Smart TV | Intermediário | R2 | Média | Estabelecida | | TVs Conectadas |
| MM-20-04 | Single Sign-On para Entretenimento | Intermediário | R2 | Média | Estabelecida | | Ecossistemas de Mídia |
| MM-20-05 | Validação de Dispositivo de Streaming | Intermediário | R2 | Média | Estabelecida | | Set-top Boxes, Dongles |
| MM-20-06 | Verificação de Região Geográfica | Intermediário | R2 | Média | Estabelecida | | Geo-restrição de Conteúdo |
| MM-20-07 | Autenticação para Jogos Online | Avançado | R3 | Média | Estabelecida | | Plataformas de Gaming |
| MM-20-08 | Verificação para Eventos ao Vivo | Avançado | R3 | Alta | Estabelecida | | Pay-per-view, Eventos |
| MM-20-09 | Autenticação por Biometria Leve | Intermediário | R2 | Média | Emergente | | Smart TV, Consoles |
| MM-20-10 | Validação por QR em Segunda Tela | Intermediário | R2 | Baixa | Estabelecida | | TV + Smartphone |
| MM-20-11 | Autenticação para Conteúdo Infantil | Intermediário | R2 | Média | Estabelecida | | Controle Parental |
| MM-20-12 | Validação para Realidade Virtual | Avançado | R3 | Alta | Emergente | | Streaming VR |
| MM-20-13 | Verificação para Conteúdo Pago | Avançado | R3 | Média | Estabelecida | | Microtransações |
| MM-20-14 | Autenticação de DRM | Avançado | R3 | Alta | Estabelecida | | Proteção de Conteúdo |
| MM-20-15 | Federação de Identidade de Mídia | Avançado | R3 | Alta | Emergente | | Múltiplas Plataformas |
| MM-20-16 | Autenticação para Redes Sociais | Intermediário | R2 | Média | Estabelecida | | Compartilhamento Social |
| MM-20-17 | Validação para Criadores de Conteúdo | Avançado | R3 | Alta | Emergente | | Plataformas de Criadores |
| MM-20-18 | Autenticação para Áudio em Alta Resolução | Avançado | R3 | Alta | Emergente | | Streaming de Áudio Premium |
| MM-20-19 | Verificação para Múltiplos Perfis | Intermediário | R2 | Média | Estabelecida | | Contas Familiares |
| MM-20-20 | Validação para Conteúdo Adulto | Avançado | R3 | Alta | Estabelecida | | Verificação de Idade |
