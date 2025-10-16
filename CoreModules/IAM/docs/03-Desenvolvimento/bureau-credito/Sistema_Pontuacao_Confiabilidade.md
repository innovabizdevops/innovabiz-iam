# Sistema de Pontuação de Confiabilidade - TrustGuard

## Visão Geral

O Sistema de Pontuação de Confiabilidade (Trust Score Engine) é um componente fundamental da integração com o TrustGuard, responsável por calcular e gerenciar pontuações multifatoriais de confiabilidade para identidades verificadas. O sistema fornece uma métrica quantitativa (0-100) que representa o grau de confiança na identidade do usuário com base em diversos fatores de verificação.

## Objetivos

- **Quantificar a confiabilidade**: Transformar múltiplos sinais de verificação em uma pontuação numérica clara
- **Suporte à decisão**: Auxiliar sistemas automatizados e operadores humanos a tomarem decisões baseadas em risco
- **Transparência**: Fornecer explicações claras sobre os fatores que compõem a pontuação
- **Adaptabilidade**: Permitir ajustes nos pesos e fatores considerados conforme necessidades específicas
- **Conformidade**: Integrar requisitos regulatórios diversos ao processo de avaliação de identidade

## Arquitetura

O sistema de pontuação é composto pelos seguintes componentes:

1. **TrustScoreEngine**: Motor principal responsável pelo cálculo das pontuações
2. **TrustScoreResult**: Estrutura de dados que armazena os resultados do cálculo
3. **TrustScoreCategory**: Classificação qualitativa das pontuações (VERY_HIGH, HIGH, MEDIUM, LOW, VERY_LOW)
4. **TrustScoreFactor**: Fatores individuais considerados no cálculo da pontuação
5. **Resolvers GraphQL**: Interfaces para acesso ao sistema via API GraphQL

## Fatores de Pontuação

O sistema considera os seguintes fatores para calcular a pontuação de confiabilidade:

| Fator | Descrição | Peso Padrão |
|-------|-----------|-------------|
| DOCUMENT_VERIFICATION | Avaliação de autenticidade e validade de documentos oficiais | 25% |
| BIOMETRIC_VERIFICATION | Correspondência entre dados biométricos e identidade declarada | 25% |
| COMPLIANCE_CHECK | Avaliação de conformidade regulatória e status em listas de observação | 15% |
| LIVENESS_DETECTION | Verificação de prova de vida para detectar tentativas de fraude | 15% |
| WATCHLIST_CHECK | Verificação contra listas de sanções, PEPs e outras listas de alerta | 10% |
| GEOGRAPHIC_CONSISTENCY | Análise de consistência geográfica de atividades | 5% |
| ACTIVITY_HISTORY | Análise de padrões históricos de atividade e comportamento | 5% |

Outros fatores que podem ser considerados em implementações futuras:
- DEVICE_REPUTATION
- BEHAVIORAL_PATTERNS
- IDENTITY_LONGEVITY

## Categorias de Pontuação

As pontuações são classificadas nas seguintes categorias:

| Categoria | Intervalo | Interpretação |
|-----------|-----------|---------------|
| VERY_HIGH | 90-100 | Confiança extremamente alta na identidade verificada |
| HIGH | 75-89 | Alta confiança na identidade verificada |
| MEDIUM | 50-74 | Confiança moderada, pode exigir verificações adicionais para operações críticas |
| LOW | 30-49 | Baixa confiança, requer verificações adicionais |
| VERY_LOW | 0-29 | Confiança muito baixa, indica potenciais problemas de identidade |

## Cálculo da Pontuação

O processo de cálculo segue estas etapas:

1. **Coleta de dados**: Agregação de todas as verificações disponíveis para o usuário
2. **Cálculo por fator**: Cada fator é avaliado separadamente com base nas verificações relevantes
3. **Ponderação**: Os scores de cada fator são ponderados conforme os pesos configurados
4. **Normalização**: A pontuação é normalizada para escala de 0-100
5. **Categorização**: A pontuação final é classificada em uma das categorias qualitativas
6. **Recomendações**: São geradas recomendações baseadas na pontuação e nos fatores individuais

## API GraphQL

O sistema expõe as seguintes operações via GraphQL:

### Mutations

#### `calculateTrustScore`

Calcula a pontuação de confiabilidade para um usuário.

```graphql
mutation CalculateTrustScore(
  $userId: String!,
  $verificationIds: [String],
  $includeUserHistory: Boolean = true,
  $contextData: JSONString
) {
  calculateTrustScore(
    userId: $userId,
    verificationIds: $verificationIds,
    includeUserHistory: $includeUserHistory,
    contextData: $contextData
  ) {
    userId
    score
    category
    factorScores {
      factor
      score
      weight
      description
    }
    recommendations
    timestamp
    expiresAt
    verificationIds
    context {
      totalVerifications
      documentVerifications
      biometricVerifications
      additionalData
    }
  }
}
```

### Queries

#### `getUserTrustScore`

Obtém a pontuação de confiabilidade atual de um usuário.

```graphql
query GetUserTrustScore($userId: String!) {
  getUserTrustScore(userId: $userId) {
    userId
    score
    category
    factorScores {
      factor
      score
      weight
      description
    }
    recommendations
    timestamp
    expiresAt
    verificationIds
    context {
      totalVerifications
      documentVerifications
      biometricVerifications
      additionalData
    }
  }
}
```

#### `getTrustScoreHistory`

Obtém o histórico de pontuações de confiabilidade de um usuário.

```graphql
query GetTrustScoreHistory(
  $userId: String!,
  $limit: Int = 10,
  $offset: Int = 0
) {
  getTrustScoreHistory(
    userId: $userId,
    limit: $limit,
    offset: $offset
  ) {
    userId
    score
    category
    factorScores {
      factor
      score
      weight
      description
    }
    recommendations
    timestamp
    expiresAt
    verificationIds
    context {
      totalVerifications
      documentVerifications
      biometricVerifications
      additionalData
    }
  }
}
```

## Permissões Requeridas

| Operação | Permissão Necessária |
|----------|---------------------|
| calculateTrustScore | iam:trustscore:calculate |
| getUserTrustScore | iam:trustscore:read |
| getTrustScoreHistory | iam:trustscore:read |

## Integração com Outros Módulos

O Sistema de Pontuação de Confiabilidade integra-se com:

1. **TrustGuard**: Fonte primária dos dados de verificação de identidade
2. **IAM Core**: Para autenticação e autorização dos usuários
3. **Bureau de Créditos**: Para enriquecer as avaliações de risco
4. **Compliance**: Para aplicar verificações regulatórias específicas de cada região

## Casos de Uso

### Abertura de Conta

Quando um novo usuário se registra, o sistema calcula uma pontuação inicial de confiabilidade com base nas verificações de documento e biometria realizadas durante o onboarding.

### Operações Financeiras

Antes de aprovar transações financeiras de alto valor, o sistema recalcula a pontuação de confiabilidade considerando verificações recentes e o histórico de atividades do usuário.

### Acesso a Recursos Sensíveis

Ao solicitar acesso a recursos sensíveis no sistema, a pontuação de confiabilidade é avaliada para determinar se verificações adicionais são necessárias.

### Monitoramento Contínuo

Periodicamente, as pontuações são recalculadas para todos os usuários ativos, permitindo a detecção de mudanças nos padrões de comportamento que possam indicar comprometimento de contas.

## Considerações Técnicas

### Performance

- As pontuações são calculadas de forma assíncrona para usuários com grande volume de verificações
- Os resultados são armazenados em cache para consultas frequentes
- A expiração padrão da pontuação é de 90 dias, configurável conforme necessidade

### Segurança

- Todas as operações de cálculo e consulta de pontuações são auditadas
- O acesso às pontuações é controlado por permissões granulares
- Componentes sensíveis do cálculo (como avaliação de listas de sanções) possuem logging detalhado

### Resiliência

- O sistema opera em modo degradado se alguns fatores não puderem ser calculados
- Mecanismos de retry estão implementados para falhas temporárias em serviços externos
- A indisponibilidade de histórico de usuário não impede o cálculo, apenas reduz a precisão

## Configuração

As configurações do sistema podem ser ajustadas através de variáveis de ambiente ou arquivo de configuração:

| Configuração | Descrição | Valor Padrão |
|--------------|-----------|--------------|
| TRUST_SCORE_EXPIRATION_DAYS | Dias de validade de uma pontuação | 90 |
| TRUST_SCORE_FACTOR_WEIGHTS | Pesos dos fatores em formato JSON | *Ver tabela de fatores* |
| TRUST_SCORE_MIN_VERIFICATIONS | Número mínimo de verificações para cálculo | 1 |

## Monitoramento e Métricas

O sistema expõe as seguintes métricas para monitoramento:

- **trust_score_calculation_duration_seconds**: Tempo de cálculo das pontuações
- **trust_score_distribution**: Distribuição das pontuações por categoria
- **trust_score_factor_availability**: Disponibilidade de dados para cada fator
- **trust_score_cache_hit_ratio**: Taxa de acertos do cache de pontuações

## Limitações Conhecidas

- O sistema requer no mínimo uma verificação de documento ou biometria para calcular uma pontuação significativa
- Verificações com mais de 1 ano são consideradas obsoletas e têm peso reduzido no cálculo
- Alguns fatores como GEOGRAPHIC_CONSISTENCY dependem de histórico suficiente para serem precisos

## Próximos Passos

- Implementação de modelo de Machine Learning para aprimorar a detecção de padrões anômalos
- Integração com sistemas anti-fraude externos para enriquecimento de dados
- Suporte a múltiplos provedores de verificação de identidade além do TrustGuard
- Implementação de dashboards específicos para visualização e análise das pontuações