// ==============================================================================
// Nome: bureauMutations.ts
// Descrição: Mutations GraphQL para o módulo Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import { gql } from '@apollo/client';
import {
  BUREAU_IDENTITY_FIELDS,
  BUREAU_AUTORIZACAO_FIELDS,
  BUREAU_TOKEN_FIELDS
} from './bureauQueries';

// Mutation para criar um novo vínculo com Bureau de Créditos
export const CREATE_BUREAU_IDENTITY = gql`
  mutation CreateBureauIdentity(
    $input: BureauIdentityInput!
  ) {
    bureauCredito {
      createIdentity(input: $input) {
        identity {
          ...BureauIdentityFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_IDENTITY_FIELDS}
`;

// Mutation para atualizar um vínculo existente
export const UPDATE_BUREAU_IDENTITY = gql`
  mutation UpdateBureauIdentity(
    $id: ID!
    $input: BureauIdentityUpdateInput!
  ) {
    bureauCredito {
      updateIdentity(id: $id, input: $input) {
        identity {
          ...BureauIdentityFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_IDENTITY_FIELDS}
`;

// Mutation para revogar (desativar) um vínculo
export const REVOKE_BUREAU_IDENTITY = gql`
  mutation RevokeBureauIdentity(
    $id: ID!
    $input: BureauIdentityRevokeInput!
  ) {
    bureauCredito {
      revokeIdentity(id: $id, input: $input) {
        identity {
          ...BureauIdentityFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_IDENTITY_FIELDS}
`;

// Mutation para criar uma nova autorização
export const CREATE_BUREAU_AUTORIZACAO = gql`
  mutation CreateBureauAutorizacao(
    $input: BureauAutorizacaoInput!
  ) {
    bureauCredito {
      createAutorizacao(input: $input) {
        autorizacao {
          ...BureauAutorizacaoFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_AUTORIZACAO_FIELDS}
`;

// Mutation para atualizar uma autorização existente
export const UPDATE_BUREAU_AUTORIZACAO = gql`
  mutation UpdateBureauAutorizacao(
    $id: ID!
    $input: BureauAutorizacaoUpdateInput!
  ) {
    bureauCredito {
      updateAutorizacao(id: $id, input: $input) {
        autorizacao {
          ...BureauAutorizacaoFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_AUTORIZACAO_FIELDS}
`;

// Mutation para cancelar uma autorização
export const CANCEL_BUREAU_AUTORIZACAO = gql`
  mutation CancelBureauAutorizacao(
    $id: ID!
    $justificativa: String!
  ) {
    bureauCredito {
      cancelAutorizacao(id: $id, justificativa: $justificativa) {
        autorizacao {
          ...BureauAutorizacaoFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_AUTORIZACAO_FIELDS}
`;

// Mutation para gerar um novo token de acesso
export const GENERATE_BUREAU_TOKEN = gql`
  mutation GenerateBureauToken(
    $input: BureauTokenInput!
  ) {
    bureauCredito {
      generateToken(input: $input) {
        token {
          ...BureauTokenFields
        }
        accessToken
        refreshToken
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_TOKEN_FIELDS}
`;

// Mutation para revogar um token
export const REVOKE_BUREAU_TOKEN = gql`
  mutation RevokeBureauToken(
    $id: ID!
    $justificativa: String
  ) {
    bureauCredito {
      revokeToken(id: $id, justificativa: $justificativa) {
        token {
          ...BureauTokenFields
        }
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_TOKEN_FIELDS}
`;

// Mutation para renovar um token usando refresh token
export const REFRESH_BUREAU_TOKEN = gql`
  mutation RefreshBureauToken(
    $refreshToken: String!
  ) {
    bureauCredito {
      refreshToken(refreshToken: $refreshToken) {
        token {
          ...BureauTokenFields
        }
        accessToken
        refreshToken
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
  ${BUREAU_TOKEN_FIELDS}
`;

// Mutation para validar um token
export const VALIDATE_BUREAU_TOKEN = gql`
  mutation ValidateBureauToken(
    $token: String!
    $ipAddress: String
  ) {
    bureauCredito {
      validateToken(token: $token, ipAddress: $ipAddress) {
        valid
        expiresIn
        escopos
        autorizacaoId
        identityId
        success
        message
        errors {
          field
          message
        }
      }
    }
  }
`;