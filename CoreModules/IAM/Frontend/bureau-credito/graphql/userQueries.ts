// ==============================================================================
// Nome: userQueries.ts
// Descrição: Queries GraphQL para usuários no contexto do Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import { gql } from '@apollo/client';

// Fragment para campos comuns de User
export const USER_FIELDS = gql`
  fragment UserFields on User {
    id
    name
    email
    status
    roles
    tenantId
    profileImage
    lastLogin
    phoneNumber
  }
`;

// Query para buscar usuários por tenant com filtros
export const SEARCH_USERS_BY_TENANT = gql`
  query SearchUsersByTenant(
    $tenantId: ID!
    $searchTerm: String
    $status: UserStatus
    $page: Int
    $pageSize: Int
  ) {
    users {
      usersByTenant(
        tenantId: $tenantId
        filter: {
          searchTerm: $searchTerm
          status: $status
        }
        page: $page
        pageSize: $pageSize
      ) {
        items {
          ...UserFields
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
  ${USER_FIELDS}
`;

// Query para buscar um usuário específico por ID
export const GET_USER = gql`
  query GetUser($id: ID!) {
    users {
      user(id: $id) {
        ...UserFields
        permissions
        metadata
        bureauIdentities {
          id
          tipoVinculo
          nivelAcesso
          status
          dataCriacao
        }
      }
    }
  }
  ${USER_FIELDS}
`;

// Query para verificar se um usuário já possui vínculo com Bureau de Créditos
export const CHECK_USER_HAS_BUREAU_IDENTITY = gql`
  query CheckUserHasBureauIdentity($userId: ID!, $tenantId: ID!) {
    bureauCredito {
      userHasIdentity(userId: $userId, tenantId: $tenantId) {
        hasIdentity
        identityId
        tipoVinculo
        nivelAcesso
        status
      }
    }
  }
`;