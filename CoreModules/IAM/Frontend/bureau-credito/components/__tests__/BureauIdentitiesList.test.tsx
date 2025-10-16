// ==============================================================================
// Nome: BureauIdentitiesList.test.tsx
// Descrição: Testes unitários para o componente BureauIdentitiesList
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../theme';
import { BureauIdentitiesList } from '../BureauIdentitiesList';
import { LIST_BUREAU_IDENTITIES } from '../../graphql/bureauQueries';
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
const mockIdentities = [
  {
    id: 'id-1',
    usuarioId: 'user-1',
    usuarioNome: 'João Silva',
    usuarioEmail: 'joao.silva@example.com',
    tipoVinculo: 'CONSULTA',
    nivelAcesso: 'BASICO',
    status: 'ATIVO',
    dataCriacao: '2025-08-01T12:00:00Z',
    dataAtualizacao: '2025-08-01T12:00:00Z',
    detalhes: { cpf: '123.456.789-00', numeroIdentificacao: '12345' },
    tenantId: 'tenant-123'
  },
  {
    id: 'id-2',
    usuarioId: 'user-2',
    usuarioNome: 'Maria Oliveira',
    usuarioEmail: 'maria.oliveira@example.com',
    tipoVinculo: 'INTEGRACAO',
    nivelAcesso: 'INTERMEDIARIO',
    status: 'PENDENTE',
    dataCriacao: '2025-08-02T14:30:00Z',
    dataAtualizacao: '2025-08-02T14:30:00Z',
    detalhes: { nif: '987654321', tipoContribuinte: 'SINGULAR' },
    tenantId: 'tenant-123'
  }
];

const mocks = [
  {
    request: {
      query: LIST_BUREAU_IDENTITIES,
      variables: {
        tenantId: 'tenant-123',
        page: 1,
        pageSize: 10,
        sortField: 'dataCriacao',
        sortOrder: 'DESC'
      }
    },
    result: {
      data: {
        bureauCredito: {
          bureauIdentities: {
            items: mockIdentities,
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
  }
];

describe('BureauIdentitiesList', () => {
  const renderComponent = (props = {}) => {
    return render(
      <MockedProvider mocks={mocks} addTypename={false}>
        <ThemeProvider theme={theme}>
          <I18nProvider>
            <MultiTenantProvider>
              <BureauIdentitiesList 
                onViewDetails={jest.fn()} 
                onRevokeIdentity={jest.fn()}
                {...props} 
              />
            </MultiTenantProvider>
          </I18nProvider>
        </ThemeProvider>
      </MockedProvider>
    );
  };

  it('deve renderizar componente de carregamento', () => {
    renderComponent();
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('deve renderizar a lista de vínculos após carregamento', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Verificar se os itens da lista estão presentes
    expect(screen.getByText('João Silva')).toBeInTheDocument();
    expect(screen.getByText('Maria Oliveira')).toBeInTheDocument();
  });

  it('deve filtrar vínculos por texto de pesquisa', async () => {
    renderComponent();
    
    await waitFor(() => {
      expect(screen.getByText('João Silva')).toBeInTheDocument();
    });

    // Simular entrada de texto no campo de pesquisa
    const searchInput = screen.getByPlaceholderText('bureau.busca_placeholder');
    fireEvent.change(searchInput, { target: { value: 'João' } });
    
    // Verificar se apenas o item filtrado está visível
    expect(screen.getByText('João Silva')).toBeInTheDocument();
    expect(screen.queryByText('Maria Oliveira')).not.toBeInTheDocument();
  });

  it('deve chamar onViewDetails quando o botão de detalhes é clicado', async () => {
    const mockViewDetails = jest.fn();
    renderComponent({ onViewDetails: mockViewDetails });
    
    await waitFor(() => {
      expect(screen.getByText('João Silva')).toBeInTheDocument();
    });

    // Encontrar e clicar no botão de detalhes
    const detailsButton = screen.getAllByRole('button', { name: 'bureau.ver_detalhes' })[0];
    fireEvent.click(detailsButton);
    
    // Verificar se a função foi chamada com o ID correto
    expect(mockViewDetails).toHaveBeenCalledWith('id-1');
  });

  it('deve exibir mensagem quando não há vínculos', async () => {
    const emptyMock = [
      {
        request: {
          query: LIST_BUREAU_IDENTITIES,
          variables: {
            tenantId: 'tenant-123',
            page: 1,
            pageSize: 10,
            sortField: 'dataCriacao',
            sortOrder: 'DESC'
          }
        },
        result: {
          data: {
            bureauCredito: {
              bureauIdentities: {
                items: [],
                totalCount: 0,
                pageInfo: {
                  hasNextPage: false,
                  hasPreviousPage: false,
                  currentPage: 1,
                  totalPages: 0
                }
              }
            }
          }
        }
      }
    ];
    
    render(
      <MockedProvider mocks={emptyMock} addTypename={false}>
        <ThemeProvider theme={theme}>
          <I18nProvider>
            <MultiTenantProvider>
              <BureauIdentitiesList 
                onViewDetails={jest.fn()} 
                onRevokeIdentity={jest.fn()}
              />
            </MultiTenantProvider>
          </I18nProvider>
        </ThemeProvider>
      </MockedProvider>
    );
    
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    expect(screen.getByText('bureau.nenhum_vinculo')).toBeInTheDocument();
  });
});