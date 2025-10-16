// ==============================================================================
// Nome: bureauQueries.ts
// Descrição: Queries GraphQL para o módulo Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import { gql } from '@apollo/client';

// Fragment para campos comuns de BureauIdentity
export const BUREAU_IDENTITY_FIELDS = gql`
  fragment BureauIdentityFields on BureauIdentity {
    id
    usuarioId
    usuarioNome
    usuarioEmail
    tipoVinculo
    nivelAcesso
    status
    dataCriacao
    dataAtualizacao
    detalhes
    tenantId
  }
`;

// Fragment para campos comuns de BureauAutorizacao
export const BUREAU_AUTORIZACAO_FIELDS = gql`
  fragment BureauAutorizacaoFields on BureauAutorizacao {
    id
    identityId
    tipoConsulta
    finalidade
    justificativa
    dataExpiracao
    dataCriacao
    status
    diasValidade
    tags
    observacoes
  }
`;

// Fragment para campos comuns de BureauToken
export const BUREAU_TOKEN_FIELDS = gql`
  fragment BureauTokenFields on BureauToken {
    id
    autorizacaoId
    token
    refreshToken
    escopos
    dataExpiracao
    dataCriacao
    rotacaoAutomatica
    usoUnico
    restricaoIp
    ultimoUso
    status
  }
`;

// Query para buscar um vínculo específico por ID
export const GET_BUREAU_IDENTITY = gql`
  query GetBureauIdentity($id: ID!) {
    bureauCredito {
      bureauIdentity(id: $id) {
        ...BureauIdentityFields
      }
    }
  }
  ${BUREAU_IDENTITY_FIELDS}
`;

// Query para listar vínculos com paginação, filtros e ordenação
export const LIST_BUREAU_IDENTITIES = gql`
  query ListBureauIdentities(
    $tenantId: ID!
    $page: Int
    $pageSize: Int
    $sortField: String
    $sortOrder: SortOrder
    $status: BureauIdentityStatus
    $tipoVinculo: TipoVinculo
    $searchTerm: String
  ) {
    bureauCredito {
      bureauIdentities(
        tenantId: $tenantId
        page: $page
        pageSize: $pageSize
        sortField: $sortField
        sortOrder: $sortOrder
        filter: {
          status: $status
          tipoVinculo: $tipoVinculo
          searchTerm: $searchTerm
        }
      ) {
        items {
          ...BureauIdentityFields
        }
        totalCount
        pageInfo {
          hasNextPage
          hasPreviousPage
          currentPage
          totalPages
        }
      }
    }
  }
  ${BUREAU_IDENTITY_FIELDS}
`;

// Query para buscar autorizações de um vínculo
export const GET_BUREAU_AUTORIZACOES = gql`
  query GetBureauAutorizacoes(
    $identityId: ID!
    $status: BureauAutorizacaoStatus
    $page: Int
    $pageSize: Int
  ) {
    bureauCredito {
      bureauAutorizacoes(
        identityId: $identityId
        filter: { status: $status }
        page: $page
        pageSize: $pageSize
      ) {
        items {
          ...BureauAutorizacaoFields
        }
        totalCount
        pageInfo {
          hasNextPage
          hasPreviousPage
          currentPage
          totalPages
        }
      }
    }
  }
  ${BUREAU_AUTORIZACAO_FIELDS}
`;

// Query para buscar uma autorização específica
export const GET_BUREAU_AUTORIZACAO = gql`
  query GetBureauAutorizacao($id: ID!) {
    bureauCredito {
      bureauAutorizacao(id: $id) {
        ...BureauAutorizacaoFields
        identity {
          ...BureauIdentityFields
        }
        tokens {
          ...BureauTokenFields
        }
      }
    }
  }
  ${BUREAU_AUTORIZACAO_FIELDS}
  ${BUREAU_IDENTITY_FIELDS}
  ${BUREAU_TOKEN_FIELDS}
`;

// Query para buscar tokens de uma autorização
export const GET_BUREAU_TOKENS = gql`
  query GetBureauTokens(
    $autorizacaoId: ID!
    $status: BureauTokenStatus
    $page: Int
    $pageSize: Int
  ) {
    bureauCredito {
      bureauTokens(
        autorizacaoId: $autorizacaoId
        filter: { status: $status }
        page: $page
        pageSize: $pageSize
      ) {
        items {
          ...BureauTokenFields
        }
        totalCount
        pageInfo {
          hasNextPage
          hasPreviousPage
          currentPage
          totalPages
        }
      }
    }
  }
  ${BUREAU_TOKEN_FIELDS}
`;

// Query para buscar escopos disponíveis para autorizações
export const GET_BUREAU_ESCOPOS_DISPONIVEIS = gql`
  query GetBureauEscoposDisponiveis($tipoVinculo: TipoVinculo!, $nivelAcesso: NivelAcesso!) {
    bureauCredito {
      escoposDisponiveis(tipoVinculo: $tipoVinculo, nivelAcesso: $nivelAcesso) {
        codigo
        descricao
        categoria
        requerPermissaoEspecial
        disponivel
      }
    }
  }
`;