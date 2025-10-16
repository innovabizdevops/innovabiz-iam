// ==============================================================================
// Nome: BureauAutorizacaoForm.test.tsx
// Descrição: Testes unitários para o componente BureauAutorizacaoForm
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../theme';
import { BureauAutorizacaoForm } from '../BureauAutorizacaoForm';
import { GET_BUREAU_IDENTITY } from '../../graphql/bureauQueries';
import { MultiTenantProvider } from '../../../contexts/MultiTenantContext';
import { I18nProvider } from '../../../contexts/I18nContext';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { ptBR } from 'date-fns/locale';

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
  }
];

describe('BureauAutorizacaoForm', () => {
  const renderComponent = (props = {}) => {
    return render(
      <LocalizationProvider dateAdapter={AdapterDateFns} adapterLocale={ptBR}>
        <MockedProvider mocks={mocks} addTypename={false}>
          <ThemeProvider theme={theme}>
            <I18nProvider>
              <MultiTenantProvider>
                <BureauAutorizacaoForm
                  identityId="id-1"
                  loading={false}
                  onSubmit={jest.fn()}
                  {...props}
                />
              </MultiTenantProvider>
            </I18nProvider>
          </ThemeProvider>
        </MockedProvider>
      </LocalizationProvider>
    );
  };

  it('deve renderizar o formulário de criação de autorização', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Verificar elementos principais do formulário
    expect(screen.getByText('bureau.criar_autorizacao_titulo')).toBeInTheDocument();
    expect(screen.getByLabelText('bureau.finalidade')).toBeInTheDocument();
    expect(screen.getByLabelText('bureau.justificativa')).toBeInTheDocument();
  });
  
  it('deve exibir informações do vínculo', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados do vínculo
    await waitFor(() => {
      expect(screen.getByText('João Silva')).toBeInTheDocument();
    });
    
    // Verificar se as informações do vínculo estão presentes
    expect(screen.getByText('bureau.tipo_consulta')).toBeInTheDocument();
    expect(screen.getByText('bureau.nivel_basico')).toBeInTheDocument();
  });

  it('deve validar campos obrigatórios', async () => {
    renderComponent();
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Submeter formulário sem preencher campos obrigatórios
    const submitButton = screen.getByRole('button', { name: 'bureau.criar_autorizacao' });
    fireEvent.click(submitButton);
    
    // Verificar mensagens de erro
    await waitFor(() => {
      expect(screen.getByText('Campo obrigatório')).toBeInTheDocument();
    });
  });

  it('deve submeter formulário com dados válidos', async () => {
    const mockSubmit = jest.fn().mockResolvedValue({ id: 'auth-new', dataExpiracao: '2025-09-19T16:30:00Z' });
    renderComponent({ onSubmit: mockSubmit });
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Preencher campos obrigatórios
    fireEvent.click(screen.getByLabelText('bureau.consulta_simples'));
    fireEvent.change(screen.getByLabelText('bureau.finalidade'), { target: { value: 'Avaliação de crédito para novo empréstimo' } });
    fireEvent.change(screen.getByLabelText('bureau.justificativa'), { target: { value: 'Cliente solicitou linha de crédito e autorizou a consulta de seus dados no bureau' } });
    
    // Adicionar tags
    const tagsInput = screen.getByLabelText('bureau.tags');
    fireEvent.change(tagsInput, { target: { value: 'empréstimo' } });
    fireEvent.keyDown(tagsInput, { key: 'Enter' });
    
    // Confirmar uso legítimo
    fireEvent.click(screen.getByLabelText('bureau.confirmar_uso_legitimo'));
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.criar_autorizacao' });
    fireEvent.click(submitButton);
    
    // Verificar se a função foi chamada com os parâmetros corretos
    await waitFor(() => {
      expect(mockSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          identityId: 'id-1',
          tipoConsulta: 'SIMPLES',
          finalidade: 'Avaliação de crédito para novo empréstimo',
          justificativa: expect.any(String),
          diasValidade: expect.any(Number)
        })
      );
    });
    
    // Verificar mensagem de sucesso
    await waitFor(() => {
      expect(screen.getByText('bureau.autorizacao_criada_sucesso')).toBeInTheDocument();
      expect(screen.getByText('bureau.autorizacao_criada_id')).toBeInTheDocument();
    });
  });

  it('deve tratar erro na submissão', async () => {
    const mockSubmit = jest.fn().mockRejectedValue(new Error('Erro ao criar autorização'));
    renderComponent({ onSubmit: mockSubmit });
    
    // Aguardar carregamento dos dados
    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    
    // Preencher campos obrigatórios
    fireEvent.click(screen.getByLabelText('bureau.consulta_simples'));
    fireEvent.change(screen.getByLabelText('bureau.finalidade'), { target: { value: 'Avaliação de crédito' } });
    fireEvent.change(screen.getByLabelText('bureau.justificativa'), { target: { value: 'Cliente solicitou linha de crédito' } });
    fireEvent.click(screen.getByLabelText('bureau.confirmar_uso_legitimo'));
    
    // Submeter formulário
    const submitButton = screen.getByRole('button', { name: 'bureau.criar_autorizacao' });
    fireEvent.click(submitButton);
    
    // Verificar mensagem de erro
    await waitFor(() => {
      expect(screen.getByText('bureau.erro_criar_autorizacao')).toBeInTheDocument();
    });
  });
});