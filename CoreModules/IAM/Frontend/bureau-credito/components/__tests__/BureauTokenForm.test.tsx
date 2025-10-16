// ==============================================================================
// Nome: BureauTokenForm.test.tsx
// Descrição: Testes unitários para o componente BureauTokenForm
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../theme';
import { BureauTokenForm } from '../BureauTokenForm';
import { GET_BUREAU_AUTORIZACOES, GET_BUREAU_ESCOPOS_DISPONIVEIS } from '../../graphql/bureauQueries';
import { GENERATE_BUREAU_TOKEN } from '../../graphql/bureauMutations';
import { MultiTenantProvider } from '../../../contexts/MultiTenantContext';
import { I18nProvider } from '../../../contexts/I18nContext';

// Mock dos dados de contexto
jest.mock('../../../hooks/useMultiTenant', () => ({
  useMultiTenant: () => ({
    currentTenant: { id: 'tenant-123', name: 'Tenant Teste' },
    userHasPermission: () => true
  })
}));

jest.mock('../../../hooks/useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    locale: 'pt-BR'
  })
}));

// Mock de dados para o GraphQL
const mockAutorizacoes = [
  {
    id: 'auth-1',
    identityId: 'id-1',
    tipoConsulta: 'SIMPLES',
    finalidade: 'Avaliação de crédito',
    justificativa: 'Análise de limite para empréstimo',
    dataExpiracao: '2025-12-01T12:00:00Z',
    dataCriacao: '2025-08-01T12:00:00Z',
    status: 'ATIVO',
    diasValidade: 120,
    tags: ['credito', 'avaliacao'],
    observacoes: 'Observação de teste'
  },
  {
    id: 'auth-2',
    identityId: 'id-1',
    tipoConsulta: 'COMPLETA',
    finalidade: 'Abertura de conta',
    justificativa: 'Verificação de restrições',
    dataExpiracao: '2025-11-01T12:00:00Z',
    dataCriacao: '2025-08-02T14:30:00Z',
    status: 'ATIVO',
    diasValidade: 90,
    tags: ['onboarding', 'abertura'],
    observacoes: null
  }
];

const mockEscopos = [
  {
    codigo: 'SCORE_CONSULTA',
    descricao: 'Consulta de score de crédito',
    categoria: 'CONSULTA',
    requerPermissaoEspecial: false,
    disponivel: true
  },
  {
    codigo: 'HISTORICO_CONSULTA',
    descricao: 'Histórico de consultas',
    categoria: 'CONSULTA',
    requerPermissaoEspecial: false,
    disponivel: true
  },
  {
    codigo: 'DETALHES_DIVIDAS',
    descricao: 'Detalhes de dívidas',
    categoria: 'FINANCEIRO',
    requerPermissaoEspecial: true,
    disponivel: true
  }
];

const mockTokenResponse = {
  token: {
    id: 'token-123',
    autorizacaoId: 'auth-1',
    token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
    refreshToken: 'refresh-token-123',
    escopos: ['SCORE_CONSULTA', 'HISTORICO_CONSULTA'],
    dataExpiracao: '2025-09-01T12:00:00Z',
    dataCriacao: '2025-08-19T16:30:00Z',
    rotacaoAutomatica: false,
    usoUnico: false,
    restricaoIp: null,
    ultimoUso: null,
    status: 'ATIVO'
  },
  accessToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
  refreshToken: 'refresh-token-123',
  success: true,
  message: 'Token gerado com sucesso',
  errors: null
};

const mocks = [
  {
    request: {
      query: GET_BUREAU_AUTORIZACOES,
      variables: {
        identityId: 'id-1',
        status: 'ATIVO',
        page: 1,
        pageSize: 50
      }
    },
    result: {
      data: {
        bureauCredito: {
          bureauAutorizacoes: {
            items: mockAutorizacoes,
            totalCount: 2,
            pageInfo: {
              hasNextPage: false,
              hasPreviousPage: false,
              currentPage: 1,
              totalPages: 1
            }
          }
        }
      }
    }
  },
  {
    request: {
      query: GET_BUREAU_ESCOPOS_DISPONIVEIS,
      variables: {
        tipoVinculo: 'CONSULTA',
        nivelAcesso: 'BASICO'
      }
    },
    result: {
      data: {
        bureauCredito: {
          escoposDisponiveis: mockEscopos
        }
      }
    }
  },
  {
    request: {
      query: GENERATE_BUREAU_TOKEN,
      variables: {
        input: {
          autorizacaoId: 'auth-1',
          escopos: ['SCORE_CONSULTA', 'HISTORICO_CONSULTA'],
          minutosValidade: 60,
          rotacaoAutomatica: false,
          usoUnico: false,
          restricaoIp: null
        }
      }
    },
    result: {
      data: {
        bureauCredito: {
          generateToken: mockTokenResponse
        }
      }
    }
  }
];

describe('BureauTokenForm', () => {
  const renderComponent = (props = {}) => {
    return render(
      <MockedProvider mocks={mocks} addTypename={false}>
        <ThemeProvider theme={theme}>
          <I18nProvider>
            <MultiTenantProvider>
              <BureauTokenForm
                identityId="id-1" 
                tipoVinculo="CONSULTA" 
                nivelAcesso="BASICO"
                loading={false}
                onSubmit={jest.fn()}
                {...props}
              />
            </MultiTenantProvider>
          </I18nProvider>
        </ThemeProvider>
      </MockedProvider>
    );
  };

  it('deve renderizar o formulário de geração de tokens', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Verificar elementos principais do formulário
    expect(screen.getByText('bureau.gerar_token_titulo')).toBeInTheDocument();
    expect(screen.getByLabelText('bureau.selecionar_autorizacao')).toBeInTheDocument();
    expect(screen.getByText('bureau.escopos_disponiveis')).toBeInTheDocument();
  });
  
  it('deve mostrar lista de autorizações disponíveis', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Abrir o select de autorizações
    fireEvent.mouseDown(screen.getByLabelText('bureau.selecionar_autorizacao'));
    
    // Verificar se as opções estão presentes
    await waitFor(() => {
      expect(screen.getByText('Avaliação de crédito')).toBeInTheDocument();
      expect(screen.getByText('Abertura de conta')).toBeInTheDocument();
    });
  });

  it('deve selecionar autorizações e escopos', async () => {
    const mockSubmit = jest.fn();
    renderComponent({ onSubmit: mockSubmit });
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Selecionar autorização
    fireEvent.mouseDown(screen.getByLabelText('bureau.selecionar_autorizacao'));
    const option = await screen.findByText('Avaliação de crédito');
    fireEvent.click(option);
    
    // Selecionar escopos
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[0]); // SCORE_CONSULTA
    fireEvent.click(checkboxes[1]); // HISTORICO_CONSULTA
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.gerar_token' });
    fireEvent.click(submitButton);
    
    // Verificar se a função foi chamada com os parâmetros corretos
    await waitFor(() => {
      expect(mockSubmit).toHaveBeenCalledWith(expect.objectContaining({
        autorizacaoId: 'auth-1',
        escopos: ['SCORE_CONSULTA', 'HISTORICO_CONSULTA'],
        minutosValidade: expect.any(Number)
      }));
    });
  });

  it('deve exibir tokens após geração bem-sucedida', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Selecionar autorização
    fireEvent.mouseDown(screen.getByLabelText('bureau.selecionar_autorizacao'));
    const option = await screen.findByText('Avaliação de crédito');
    fireEvent.click(option);
    
    // Selecionar escopos
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[0]); // SCORE_CONSULTA
    fireEvent.click(checkboxes[1]); // HISTORICO_CONSULTA
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.gerar_token' });
    fireEvent.click(submitButton);
    
    // Verificar se os tokens são exibidos
    await waitFor(() => {
      expect(screen.getByText('bureau.token_gerado_sucesso')).toBeInTheDocument();
      expect(screen.getByText('bureau.token_acesso')).toBeInTheDocument();
      expect(screen.getByText('bureau.token_refresh')).toBeInTheDocument();
    });
    
    // Verificar presença dos tokens
    expect(screen.getByText('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...')).toBeInTheDocument();
    expect(screen.getByText('refresh-token-123')).toBeInTheDocument();
  });
});