// ==============================================================================
// Nome: BureauRevokeForm.test.tsx
// Descrição: Testes unitários para o componente BureauRevokeForm
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../theme';
import { BureauRevokeForm } from '../BureauRevokeForm';
import { GET_BUREAU_IDENTITY } from '../../graphql/bureauQueries';
import { REVOKE_BUREAU_IDENTITY } from '../../graphql/bureauMutations';
import { MultiTenantProvider } from '../../../contexts/MultiTenantContext';
import { I18nProvider } from '../../../contexts/I18nContext';

// Mock dos dados de contexto
jest.mock('../../../hooks/useMultiTenant', () => ({
  useMultiTenant: () => ({
    currentTenant: { id: 'tenant-123', name: 'Tenant Teste' }
  })
}));

jest.mock('../../../hooks/useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    locale: 'pt-BR'
  })
}));

// Mock de dados para o GraphQL
const mockIdentity = {
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
};

const mocks = [
  {
    request: {
      query: GET_BUREAU_IDENTITY,
      variables: {
        id: 'id-1'
      }
    },
    result: {
      data: {
        bureauCredito: {
          bureauIdentity: mockIdentity
        }
      }
    }
  },
  {
    request: {
      query: REVOKE_BUREAU_IDENTITY,
      variables: {
        id: 'id-1',
        motivoRevogacao: 'Solicitação do cliente',
        detalhesRevogacao: 'Cliente solicitou revogação do acesso ao bureau',
        revogarTokens: true
      }
    },
    result: {
      data: {
        bureauCredito: {
          revokeIdentity: {
            success: true,
            message: 'Vínculo revogado com sucesso',
            identity: {
              id: 'id-1',
              status: 'REVOGADO',
              dataAtualizacao: '2025-08-19T16:30:00Z'
            },
            errors: null
          }
        }
      }
    }
  }
];

describe('BureauRevokeForm', () => {
  const renderComponent = (props = {}) => {
    return render(
      <MockedProvider mocks={mocks} addTypename={false}>
        <ThemeProvider theme={theme}>
          <I18nProvider>
            <MultiTenantProvider>
              <BureauRevokeForm
                identityId="id-1"
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

  it('deve renderizar o formulário de revogação', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Verificar elementos principais do formulário
    expect(screen.getByText('bureau.revogar_vinculo_titulo')).toBeInTheDocument();
    expect(screen.getByLabelText('bureau.motivo_revogacao')).toBeInTheDocument();
    expect(screen.getByLabelText('bureau.detalhes_revogacao')).toBeInTheDocument();
    expect(screen.getByLabelText('bureau.revogar_tokens')).toBeInTheDocument();
  });
  
  it('deve exibir informações do vínculo a ser revogado', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados do vínculo
    await waitFor(() => {
      expect(screen.getByText('João Silva')).toBeInTheDocument();
    });
    
    // Verificar se as informações do vínculo estão presentes
    expect(screen.getByText('bureau.email_label')).toBeInTheDocument();
    expect(screen.getByText('joao.silva@example.com')).toBeInTheDocument();
  });

  it('deve validar campos obrigatórios', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Submeter formulário sem preencher campos obrigatórios
    const submitButton = screen.getByRole('button', { name: 'bureau.revogar_vinculo' });
    fireEvent.click(submitButton);
    
    // Verificar mensagens de erro
    await waitFor(() => {
      expect(screen.getByText('Campo obrigatório')).toBeInTheDocument();
    });
  });

  it('deve submeter formulário com dados válidos', async () => {
    const mockSubmit = jest.fn().mockResolvedValue({ success: true });
    renderComponent({ onSubmit: mockSubmit });
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Preencher campos obrigatórios
    fireEvent.change(screen.getByLabelText('bureau.motivo_revogacao'), { target: { value: 'Solicitação do cliente' } });
    fireEvent.change(screen.getByLabelText('bureau.detalhes_revogacao'), { target: { value: 'Cliente solicitou revogação do acesso ao bureau' } });
    fireEvent.click(screen.getByLabelText('bureau.revogar_tokens'));
    
    // Confirmar entendimento de consequências
    fireEvent.click(screen.getByLabelText('bureau.confirmar_entendimento_revogacao'));
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.revogar_vinculo' });
    fireEvent.click(submitButton);
    
    // Verificar se a função foi chamada com os parâmetros corretos
    await waitFor(() => {
      expect(mockSubmit).toHaveBeenCalledWith({
        id: 'id-1',
        motivoRevogacao: 'Solicitação do cliente',
        detalhesRevogacao: 'Cliente solicitou revogação do acesso ao bureau',
        revogarTokens: true
      });
    });
    
    // Verificar mensagem de sucesso
    await waitFor(() => {
      expect(screen.getByText('bureau.vinculo_revogado_sucesso')).toBeInTheDocument();
    });
  });

  it('deve exibir diálogo de confirmação antes de submeter', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Preencher campos obrigatórios
    fireEvent.change(screen.getByLabelText('bureau.motivo_revogacao'), { target: { value: 'Solicitação do cliente' } });
    fireEvent.change(screen.getByLabelText('bureau.detalhes_revogacao'), { target: { value: 'Cliente solicitou revogação' } });
    fireEvent.click(screen.getByLabelText('bureau.confirmar_entendimento_revogacao'));
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.revogar_vinculo' });
    fireEvent.click(submitButton);
    
    // Verificar se o diálogo de confirmação está visível
    await waitFor(() => {
      expect(screen.getByText('bureau.confirmar_revogacao_titulo')).toBeInTheDocument();
    });
    
    // Confirmar a operação no diálogo
    const confirmButton = screen.getByRole('button', { name: 'bureau.confirmar' });
    fireEvent.click(confirmButton);
    
    // Verificar se a operação foi concluída
    await waitFor(() => {
      expect(screen.getByText('bureau.vinculo_revogado_sucesso')).toBeInTheDocument();
    });
  });

  it('deve cancelar operação quando solicitado', async () => {
    const mockSubmit = jest.fn();
    renderComponent({ onSubmit: mockSubmit });
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Preencher campos obrigatórios
    fireEvent.change(screen.getByLabelText('bureau.motivo_revogacao'), { target: { value: 'Solicitação do cliente' } });
    fireEvent.change(screen.getByLabelText('bureau.detalhes_revogacao'), { target: { value: 'Cliente solicitou revogação' } });
    fireEvent.click(screen.getByLabelText('bureau.confirmar_entendimento_revogacao'));
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.revogar_vinculo' });
    fireEvent.click(submitButton);
    
    // Verificar se o diálogo de confirmação está visível
    await waitFor(() => {
      expect(screen.getByText('bureau.confirmar_revogacao_titulo')).toBeInTheDocument();
    });
    
    // Cancelar a operação no diálogo
    const cancelButton = screen.getByRole('button', { name: 'bureau.cancelar' });
    fireEvent.click(cancelButton);
    
    // Verificar se a operação foi cancelada (diálogo fechado)
    await waitFor(() => {
      expect(screen.queryByText('bureau.confirmar_revogacao_titulo')).not.toBeInTheDocument();
    });
    
    // Confirmar que a função de submissão não foi chamada
    expect(mockSubmit).not.toHaveBeenCalled();
  });
});