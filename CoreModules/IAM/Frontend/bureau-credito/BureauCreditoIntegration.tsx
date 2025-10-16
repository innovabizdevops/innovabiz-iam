// ==============================================================================
// Nome: BureauCreditoIntegration.tsx
// Descrição: Componente principal para integração com Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React, { useState, useEffect } from 'react';
import { useQuery, useMutation } from '@apollo/client';
import { 
  GET_BUREAU_IDENTITIES, 
  GET_BUREAU_IDENTITY,
  CREATE_BUREAU_IDENTITY,
  CREATE_BUREAU_AUTORIZACAO,
  GENERATE_BUREAU_TOKEN,
  REVOKE_BUREAU_VINCULO
} from './graphql/bureauQueries';
import { 
  Container, 
  Typography, 
  Box, 
  Grid, 
  Button, 
  Paper, 
  Tabs, 
  Tab, 
  CircularProgress, 
  Alert,
  Snackbar,
  useTheme
} from '@mui/material';

import { useAuth } from '../../hooks/useAuth';
import { usePermissions } from '../../hooks/usePermissions';
import { useFeatureFlags } from '../../hooks/useFeatureFlags';
import { useTranslation } from '../../hooks/useTranslation';
import { useMultiTenant } from '../../hooks/useMultiTenant';
import { useAnalytics } from '../../hooks/useAnalytics';

// Componentes internos
import { BureauIdentitiesList } from './components/BureauIdentitiesList';
import { BureauIdentityDetails } from './components/BureauIdentityDetails';
import { BureauIdentityForm } from './components/BureauIdentityForm';
import { BureauAutorizacaoForm } from './components/BureauAutorizacaoForm';
import { BureauTokenForm } from './components/BureauTokenForm';
import { BureauRevokeForm } from './components/BureauRevokeForm';
import { PageHeader } from '../../components/PageHeader';
import { AccessDenied } from '../../components/AccessDenied';
import { ErrorBoundary } from '../../components/ErrorBoundary';
import { AuditLogViewer } from '../../components/AuditLogViewer';

// Tipos
import { 
  BureauIdentity, 
  BureauAutorizacao, 
  BureauToken, 
  TipoVinculo, 
  NivelAcesso, 
  TipoConsulta,
  BureauVinculoStatus,
  BureauAutorizacaoStatus
} from '../../types/bureau-credito';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

// Componente TabPanel para organização das abas
function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`bureau-tabpanel-${index}`}
      aria-labelledby={`bureau-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

/**
 * Componente principal para gerenciamento da integração com Bureau de Créditos
 * 
 * Este componente fornece uma interface completa para:
 * - Listar vínculos com Bureau de Créditos
 * - Visualizar detalhes de um vínculo
 * - Criar novos vínculos
 * - Gerar autorizações de consulta
 * - Gerar tokens de acesso temporários
 * - Revogar vínculos
 * 
 * Implementa controle de permissões granular e suporta múltiplos contextos
 * de tenant com configurações específicas por organização.
 */
export const BureauCreditoIntegration: React.FC = () => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { user, isAuthenticated, hasValidToken } = useAuth();
  const { currentTenant, tenantConfig } = useMultiTenant();
  const { trackEvent } = useAnalytics();
  const { hasPermission } = usePermissions();
  
  // Estado local
  const [tabValue, setTabValue] = useState(0);
  const [selectedIdentityId, setSelectedIdentityId] = useState<string | null>(null);
  const [filter, setFilter] = useState({
    status: '' as BureauVinculoStatus | '',
    tipoVinculo: '' as TipoVinculo | '',
  });
  const [notification, setNotification] = useState({
    open: false,
    message: '',
    severity: 'info' as 'success' | 'info' | 'warning' | 'error'
  });

  // Verifica se o recurso está habilitado para o tenant atual
  const { isEnabled: isBureauEnabled } = useFeatureFlags('BUREAU_CREDITO_INTEGRATION');

  // Verifica permissões do usuário para este módulo
  const canViewBureau = hasPermission('bureau_credito:list');
  const canCreateVinculo = hasPermission('bureau_credito:create_vinculo');
  const canCreateAutorizacao = hasPermission('bureau_credito:create_autorizacao');
  const canGenerateToken = hasPermission('bureau_credito:generate_token');
  const canRevokeVinculo = hasPermission('bureau_credito:revoke_vinculo');
  
  // Query para buscar vínculos
  const { 
    loading: loadingIdentities, 
    error: errorIdentities, 
    data: dataIdentities,
    refetch: refetchIdentities
  } = useQuery(GET_BUREAU_IDENTITIES, {
    variables: { 
      tenantId: currentTenant?.id,
      ...filter
    },
    skip: !currentTenant || !canViewBureau || !isAuthenticated,
    fetchPolicy: 'network-only'
  });

  // Query para buscar detalhes de um vínculo específico
  const { 
    loading: loadingIdentityDetails, 
    error: errorIdentityDetails, 
    data: dataIdentityDetails 
  } = useQuery(GET_BUREAU_IDENTITY, {
    variables: { id: selectedIdentityId },
    skip: !selectedIdentityId || !canViewBureau,
    fetchPolicy: 'network-only'
  });

  // Mutations
  const [createVinculo, { loading: loadingCreateVinculo }] = useMutation(CREATE_BUREAU_IDENTITY);
  const [createAutorizacao, { loading: loadingCreateAutorizacao }] = useMutation(CREATE_BUREAU_AUTORIZACAO);
  const [generateToken, { loading: loadingGenerateToken }] = useMutation(GENERATE_BUREAU_TOKEN);
  const [revokeVinculo, { loading: loadingRevokeVinculo }] = useMutation(REVOKE_BUREAU_VINCULO);

  // Função para mudar a aba atual
  const handleChangeTab = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  // Selecionar uma identidade para visualizar detalhes
  const handleSelectIdentity = (identityId: string) => {
    setSelectedIdentityId(identityId);
    setTabValue(1); // Muda para a aba de detalhes
    trackEvent('bureau_credito_select_identity', { identityId });
  };

  // Função para criar um novo vínculo
  const handleCreateVinculo = async (data: {
    usuarioId: string;
    tipoVinculo: TipoVinculo;
    nivelAcesso: NivelAcesso;
    detalhes?: Record<string, any>;
  }) => {
    try {
      const response = await createVinculo({
        variables: {
          input: {
            ...data,
            tenantId: currentTenant?.id
          }
        }
      });

      setNotification({
        open: true,
        message: t('bureau.vinculo_criado_sucesso'),
        severity: 'success'
      });

      // Recarregar a lista após criação
      refetchIdentities();
      
      // Redirecionar para detalhes do novo vínculo
      const newIdentityId = response.data?.bureauCredito?.criarVinculoBureau?.id;
      if (newIdentityId) {
        handleSelectIdentity(newIdentityId);
      }
      
      trackEvent('bureau_credito_create_vinculo_success', { usuarioId: data.usuarioId });
      return true;
    } catch (error) {
      console.error('Erro ao criar vínculo:', error);
      setNotification({
        open: true,
        message: t('bureau.erro_criar_vinculo'),
        severity: 'error'
      });
      trackEvent('bureau_credito_create_vinculo_error', { error: (error as Error).message });
      return false;
    }
  };

  // Função para criar nova autorização
  const handleCreateAutorizacao = async (data: {
    identityId: string;
    tipoConsulta: TipoConsulta;
    finalidade: string;
    justificativa: string;
    diasValidade: number;
  }) => {
    try {
      const response = await createAutorizacao({
        variables: {
          input: {
            ...data,
            operadorId: user?.id
          }
        }
      });

      setNotification({
        open: true,
        message: t('bureau.autorizacao_criada_sucesso'),
        severity: 'success'
      });
      
      trackEvent('bureau_credito_create_autorizacao_success', { identityId: data.identityId });
      return response.data?.bureauCredito?.criarAutorizacaoBureau;
    } catch (error) {
      console.error('Erro ao criar autorização:', error);
      setNotification({
        open: true,
        message: t('bureau.erro_criar_autorizacao'),
        severity: 'error'
      });
      trackEvent('bureau_credito_create_autorizacao_error', { error: (error as Error).message });
      return null;
    }
  };

  // Função para gerar token
  const handleGenerateToken = async (data: {
    autorizacaoId: string;
    escopos: string[];
    expiracaoMinutos: number;
  }) => {
    try {
      const response = await generateToken({
        variables: {
          input: {
            ...data,
            operadorId: user?.id
          }
        }
      });

      setNotification({
        open: true,
        message: t('bureau.token_gerado_sucesso'),
        severity: 'success'
      });
      
      trackEvent('bureau_credito_generate_token_success', { autorizacaoId: data.autorizacaoId });
      return response.data?.bureauCredito?.gerarTokenBureau;
    } catch (error) {
      console.error('Erro ao gerar token:', error);
      setNotification({
        open: true,
        message: t('bureau.erro_gerar_token'),
        severity: 'error'
      });
      trackEvent('bureau_credito_generate_token_error', { error: (error as Error).message });
      return null;
    }
  };

  // Função para revogar vínculo
  const handleRevokeVinculo = async (data: {
    identityId: string;
    motivo: string;
  }) => {
    try {
      await revokeVinculo({
        variables: {
          input: {
            ...data
          }
        }
      });

      setNotification({
        open: true,
        message: t('bureau.vinculo_revogado_sucesso'),
        severity: 'success'
      });

      // Recarregar a lista após revogação
      refetchIdentities();
      
      // Limpar seleção se o vínculo revogado for o atualmente selecionado
      if (selectedIdentityId === data.identityId) {
        setSelectedIdentityId(null);
        setTabValue(0); // Retornar para a lista
      }
      
      trackEvent('bureau_credito_revoke_vinculo_success', { identityId: data.identityId });
      return true;
    } catch (error) {
      console.error('Erro ao revogar vínculo:', error);
      setNotification({
        open: true,
        message: t('bureau.erro_revogar_vinculo'),
        severity: 'error'
      });
      trackEvent('bureau_credito_revoke_vinculo_error', { error: (error as Error).message });
      return false;
    }
  };

  // Fechar notificação
  const handleCloseNotification = () => {
    setNotification({
      ...notification,
      open: false
    });
  };

  // Se recurso não estiver habilitado para o tenant
  if (!isBureauEnabled) {
    return (
      <Container maxWidth="lg">
        <PageHeader title={t('bureau.titulo')} />
        <Alert severity="warning">
          {t('bureau.recurso_nao_habilitado')}
        </Alert>
      </Container>
    );
  }

  // Se usuário não tiver permissão
  if (!canViewBureau) {
    return <AccessDenied resource="bureau_credito" />;
  }

  return (
    <ErrorBoundary>
      <Container maxWidth="lg">
        <PageHeader 
          title={t('bureau.titulo')} 
          subtitle={t('bureau.subtitulo')}
          icon="credit_score"
        />

        {/* Sistema de abas para navegação entre funcionalidades */}
        <Paper sx={{ width: '100%', mb: 2 }}>
          <Tabs
            value={tabValue}
            onChange={handleChangeTab}
            indicatorColor="primary"
            textColor="primary"
            variant="scrollable"
            scrollButtons="auto"
            aria-label="abas bureau de crédito"
          >
            <Tab label={t('bureau.tab_vinculos')} id="bureau-tab-0" aria-controls="bureau-tabpanel-0" />
            <Tab 
              label={t('bureau.tab_detalhes')} 
              id="bureau-tab-1" 
              aria-controls="bureau-tabpanel-1"
              disabled={!selectedIdentityId} 
            />
            <Tab 
              label={t('bureau.tab_novo_vinculo')} 
              id="bureau-tab-2" 
              aria-controls="bureau-tabpanel-2" 
              disabled={!canCreateVinculo}
            />
            <Tab 
              label={t('bureau.tab_autorizacao')} 
              id="bureau-tab-3" 
              aria-controls="bureau-tabpanel-3" 
              disabled={!canCreateAutorizacao || !selectedIdentityId}
            />
            <Tab 
              label={t('bureau.tab_gerar_token')} 
              id="bureau-tab-4" 
              aria-controls="bureau-tabpanel-4" 
              disabled={!canGenerateToken || !selectedIdentityId}
            />
            <Tab 
              label={t('bureau.tab_revogar')} 
              id="bureau-tab-5" 
              aria-controls="bureau-tabpanel-5" 
              disabled={!canRevokeVinculo || !selectedIdentityId}
            />
            <Tab 
              label={t('bureau.tab_auditoria')} 
              id="bureau-tab-6" 
              aria-controls="bureau-tabpanel-6"
            />
          </Tabs>
        </Paper>

        {/* Conteúdo de cada aba */}
        <TabPanel value={tabValue} index={0}>
          <BureauIdentitiesList
            loading={loadingIdentities}
            error={errorIdentities}
            identities={dataIdentities?.bureauCredito?.bureauIdentitiesByTenant || []}
            onSelect={handleSelectIdentity}
            filter={filter}
            onFilterChange={setFilter}
            onRefresh={refetchIdentities}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <BureauIdentityDetails
            loading={loadingIdentityDetails}
            error={errorIdentityDetails}
            identity={dataIdentityDetails?.bureauCredito?.bureauIdentity}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          <BureauIdentityForm
            onSubmit={handleCreateVinculo}
            loading={loadingCreateVinculo}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={3}>
          <BureauAutorizacaoForm
            identityId={selectedIdentityId || ''}
            onSubmit={handleCreateAutorizacao}
            loading={loadingCreateAutorizacao}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={4}>
          <BureauTokenForm
            identityId={selectedIdentityId || ''}
            onSubmit={handleGenerateToken}
            loading={loadingGenerateToken}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={5}>
          <BureauRevokeForm
            identityId={selectedIdentityId || ''}
            onSubmit={handleRevokeVinculo}
            loading={loadingRevokeVinculo}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={6}>
          <AuditLogViewer
            moduleId="BUREAU_CREDITO"
            entityId={selectedIdentityId}
            tenantId={currentTenant?.id}
          />
        </TabPanel>
      </Container>

      {/* Sistema de notificações */}
      <Snackbar
        open={notification.open}
        autoHideDuration={6000}
        onClose={handleCloseNotification}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          onClose={handleCloseNotification}
          severity={notification.severity}
          sx={{ width: '100%' }}
        >
          {notification.message}
        </Alert>
      </Snackbar>
    </ErrorBoundary>
  );
};

export default BureauCreditoIntegration;