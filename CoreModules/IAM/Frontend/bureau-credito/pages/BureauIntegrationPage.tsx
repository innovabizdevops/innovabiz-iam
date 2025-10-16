// ==============================================================================
// Nome: BureauIntegrationPage.tsx
// Descrição: Página principal de integração com o Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React, { useState, useCallback, useEffect } from 'react';
import { 
  Container, 
  Typography, 
  Box, 
  Paper, 
  Tabs, 
  Tab, 
  Grid, 
  Alert,
  AlertTitle,
  Divider,
  Breadcrumbs,
  Link,
  CircularProgress
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import { useMutation, useQuery } from '@apollo/client';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from '../../../hooks/useTranslation';
import { useMultiTenant } from '../../../hooks/useMultiTenant';
import { useNotification } from '../../../hooks/useNotification';
import { BureauIdentitiesList } from '../components/BureauIdentitiesList';
import { BureauIdentityForm } from '../components/BureauIdentityForm';
import { BureauAutorizacaoForm } from '../components/BureauAutorizacaoForm';
import { BureauTokenForm } from '../components/BureauTokenForm';
import { BureauRevokeForm } from '../components/BureauRevokeForm';
import { 
  GET_BUREAU_IDENTITY, 
  LIST_BUREAU_IDENTITIES 
} from '../graphql/bureauQueries';
import {
  CREATE_BUREAU_IDENTITY,
  CREATE_BUREAU_AUTORIZACAO,
  REVOKE_BUREAU_IDENTITY,
  GENERATE_BUREAU_TOKEN
} from '../graphql/bureauMutations';
import { 
  HomeOutlined as HomeIcon,
  AccountBalanceOutlined as BureauIcon,
  AdminPanelSettingsOutlined as AdminIcon
} from '@mui/icons-material';
import { PageTitle } from '../../../components/common/PageTitle';
import { TabPanel } from '../../../components/common/TabPanel';
import { Permission } from '../../../constants/permissions';

// Interface para os parâmetros de rota
interface RouteParams {
  identityId?: string;
}

// Interface para as estatísticas do Bureau
interface BureauStats {
  totalVinculos: number;
  vinculosAtivos: number;
  autorizacoesAtivas: number;
  tokensAtivos: number;
}

export const BureauIntegrationPage: React.FC = () => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { currentTenant, userHasPermission } = useMultiTenant();
  const { showNotification } = useNotification();
  const navigate = useNavigate();
  const { identityId } = useParams<RouteParams>();
  
  const [tabValue, setTabValue] = useState(0);
  const [selectedIdentityId, setSelectedIdentityId] = useState<string | null>(identityId || null);
  const [bureauStats, setBureauStats] = useState<BureauStats>({
    totalVinculos: 0,
    vinculosAtivos: 0,
    autorizacoesAtivas: 0,
    tokensAtivos: 0
  });
  
  // Consultas GraphQL
  const { 
    data: identitiesData, 
    loading: loadingIdentities,
    refetch: refetchIdentities 
  } = useQuery(LIST_BUREAU_IDENTITIES, {
    variables: { 
      tenantId: currentTenant?.id,
      page: 1,
      pageSize: 10,
      sortField: 'dataCriacao',
      sortOrder: 'DESC'
    },
    skip: !currentTenant?.id,
    fetchPolicy: 'network-only',
    onCompleted: (data) => {
      const identities = data?.bureauCredito?.bureauIdentities?.items || [];
      setBureauStats(prevStats => ({
        ...prevStats,
        totalVinculos: data?.bureauCredito?.bureauIdentities?.totalCount || 0,
        vinculosAtivos: identities.filter(i => i.status === 'ATIVO').length
      }));
    }
  });
  
  const { 
    data: identityData, 
    loading: loadingIdentity 
  } = useQuery(GET_BUREAU_IDENTITY, {
    variables: { id: selectedIdentityId },
    skip: !selectedIdentityId,
    fetchPolicy: 'network-only',
    onCompleted: (data) => {
      if (data?.bureauCredito?.bureauIdentity) {
        // Carregar estatísticas associadas ao vínculo
        const identity = data.bureauCredito.bureauIdentity;
        setBureauStats(prevStats => ({
          ...prevStats,
          autorizacoesAtivas: identity.autorizacoesAtivas || 0,
          tokensAtivos: identity.tokensAtivos || 0
        }));
      }
    }
  });
  
  // Mutações GraphQL
  const [createIdentity, { loading: loadingCreateIdentity }] = useMutation(CREATE_BUREAU_IDENTITY);
  const [createAutorizacao, { loading: loadingCreateAutorizacao }] = useMutation(CREATE_BUREAU_AUTORIZACAO);
  const [revokeIdentity, { loading: loadingRevokeIdentity }] = useMutation(REVOKE_BUREAU_IDENTITY);
  const [generateToken, { loading: loadingGenerateToken }] = useMutation(GENERATE_BUREAU_TOKEN);
  
  // Manipuladores de eventos
  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };
  
  const handleViewDetails = useCallback((id: string) => {
    setSelectedIdentityId(id);
    setTabValue(1); // Alterar para a aba de detalhes
  }, []);
  
  const handleCreateIdentity = useCallback(async (formData: any) => {
    try {
      const { data } = await createIdentity({
        variables: { 
          input: {
            ...formData,
            tenantId: currentTenant?.id
          }
        }
      });
      
      if (data?.bureauCredito?.createIdentity?.success) {
        showNotification({
          type: 'success',
          message: t('bureau.vinculo_criado_sucesso')
        });
        
        refetchIdentities();
        setTabValue(0); // Retornar à lista
        return data.bureauCredito.createIdentity;
      } else {
        throw new Error(data?.bureauCredito?.createIdentity?.errors?.[0]?.message || t('bureau.erro_criar_vinculo'));
      }
    } catch (error) {
      showNotification({
        type: 'error',
        message: (error as Error).message
      });
      throw error;
    }
  }, [createIdentity, currentTenant, refetchIdentities, showNotification, t]);
  
  const handleCreateAutorizacao = useCallback(async (formData: any) => {
    try {
      const { data } = await createAutorizacao({
        variables: { 
          input: {
            ...formData,
            tenantId: currentTenant?.id
          }
        }
      });
      
      if (data?.bureauCredito?.createAutorizacao?.success) {
        showNotification({
          type: 'success',
          message: t('bureau.autorizacao_criada_sucesso')
        });
        
        // Atualizar estatísticas
        setBureauStats(prevStats => ({
          ...prevStats,
          autorizacoesAtivas: prevStats.autorizacoesAtivas + 1
        }));
        
        return data.bureauCredito.createAutorizacao;
      } else {
        throw new Error(data?.bureauCredito?.createAutorizacao?.errors?.[0]?.message || t('bureau.erro_criar_autorizacao'));
      }
    } catch (error) {
      showNotification({
        type: 'error',
        message: (error as Error).message
      });
      throw error;
    }
  }, [createAutorizacao, currentTenant, showNotification, t]);
  
  const handleGenerateToken = useCallback(async (formData: any) => {
    try {
      const { data } = await generateToken({
        variables: { 
          input: {
            ...formData,
            tenantId: currentTenant?.id
          }
        }
      });
      
      if (data?.bureauCredito?.generateToken?.success) {
        showNotification({
          type: 'success',
          message: t('bureau.token_gerado_sucesso')
        });
        
        // Atualizar estatísticas
        setBureauStats(prevStats => ({
          ...prevStats,
          tokensAtivos: prevStats.tokensAtivos + 1
        }));
        
        return data.bureauCredito.generateToken;
      } else {
        throw new Error(data?.bureauCredito?.generateToken?.errors?.[0]?.message || t('bureau.erro_gerar_token'));
      }
    } catch (error) {
      showNotification({
        type: 'error',
        message: (error as Error).message
      });
      throw error;
    }
  }, [generateToken, currentTenant, showNotification, t]);
  
  const handleRevokeIdentity = useCallback(async (formData: any) => {
    try {
      const { data } = await revokeIdentity({
        variables: { ...formData }
      });
      
      if (data?.bureauCredito?.revokeIdentity?.success) {
        showNotification({
          type: 'success',
          message: t('bureau.vinculo_revogado_sucesso')
        });
        
        refetchIdentities();
        setTabValue(0); // Retornar à lista
        setSelectedIdentityId(null);
        
        // Atualizar estatísticas
        setBureauStats(prevStats => ({
          ...prevStats,
          vinculosAtivos: prevStats.vinculosAtivos - 1
        }));
        
        return data.bureauCredito.revokeIdentity;
      } else {
        throw new Error(data?.bureauCredito?.revokeIdentity?.errors?.[0]?.message || t('bureau.erro_revogar_vinculo'));
      }
    } catch (error) {
      showNotification({
        type: 'error',
        message: (error as Error).message
      });
      throw error;
    }
  }, [revokeIdentity, refetchIdentities, showNotification, t]);
  
  // Verificar permissões necessárias
  const canViewBureauIdentities = userHasPermission(Permission.BUREAU_VIEW_IDENTITIES);
  const canCreateBureauIdentity = userHasPermission(Permission.BUREAU_CREATE_IDENTITY);
  const canManageBureauAutorizacoes = userHasPermission(Permission.BUREAU_MANAGE_AUTORIZACOES);
  const canGenerateBureauTokens = userHasPermission(Permission.BUREAU_GENERATE_TOKENS);
  const canRevokeBureauIdentity = userHasPermission(Permission.BUREAU_REVOKE_IDENTITY);
  
  // Renderizar mensagem de acesso negado quando não houver permissão
  if (!canViewBureauIdentities) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Alert severity="error">
          <AlertTitle>{t('common.acesso_negado')}</AlertTitle>
          {t('bureau.sem_permissao_acesso')}
        </Alert>
      </Container>
    );
  }
  
  // Identidade selecionada
  const selectedIdentity = identityData?.bureauCredito?.bureauIdentity;
  
  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <Box sx={{ mb: 4 }}>
        <Breadcrumbs aria-label="breadcrumb" sx={{ mb: 2 }}>
          <Link 
            color="inherit"
            underline="hover"
            sx={{ display: 'flex', alignItems: 'center' }}
            onClick={() => navigate('/')}
          >
            <HomeIcon sx={{ mr: 0.5 }} fontSize="inherit" />
            {t('common.inicio')}
          </Link>
          <Link
            color="inherit"
            underline="hover"
            sx={{ display: 'flex', alignItems: 'center' }}
            onClick={() => navigate('/admin')}
          >
            <AdminIcon sx={{ mr: 0.5 }} fontSize="inherit" />
            {t('common.administracao')}
          </Link>
          <Typography
            sx={{ display: 'flex', alignItems: 'center' }}
            color="text.primary"
          >
            <BureauIcon sx={{ mr: 0.5 }} fontSize="inherit" />
            {t('bureau.gestao_bureau')}
          </Typography>
        </Breadcrumbs>
        
        <PageTitle title={t('bureau.integracao_bureau_creditos')} />
        
        <Box sx={{ mb: 4, mt: 2 }}>
          <Grid container spacing={2}>
            <Grid item xs={12} md={3}>
              <Paper
                elevation={2}
                sx={{
                  p: 2,
                  textAlign: 'center',
                  backgroundColor: theme.palette.primary.light,
                  color: theme.palette.primary.contrastText
                }}
              >
                <Typography variant="h4">{bureauStats.totalVinculos}</Typography>
                <Typography variant="body2">{t('bureau.total_vinculos')}</Typography>
              </Paper>
            </Grid>
            <Grid item xs={12} md={3}>
              <Paper
                elevation={2}
                sx={{
                  p: 2,
                  textAlign: 'center',
                  backgroundColor: theme.palette.success.light,
                  color: theme.palette.success.contrastText
                }}
              >
                <Typography variant="h4">{bureauStats.vinculosAtivos}</Typography>
                <Typography variant="body2">{t('bureau.vinculos_ativos')}</Typography>
              </Paper>
            </Grid>
            <Grid item xs={12} md={3}>
              <Paper
                elevation={2}
                sx={{
                  p: 2,
                  textAlign: 'center',
                  backgroundColor: theme.palette.info.light,
                  color: theme.palette.info.contrastText
                }}
              >
                <Typography variant="h4">{bureauStats.autorizacoesAtivas}</Typography>
                <Typography variant="body2">{t('bureau.autorizacoes_ativas')}</Typography>
              </Paper>
            </Grid>
            <Grid item xs={12} md={3}>
              <Paper
                elevation={2}
                sx={{
                  p: 2,
                  textAlign: 'center',
                  backgroundColor: theme.palette.warning.light,
                  color: theme.palette.warning.contrastText
                }}
              >
                <Typography variant="h4">{bureauStats.tokensAtivos}</Typography>
                <Typography variant="body2">{t('bureau.tokens_ativos')}</Typography>
              </Paper>
            </Grid>
          </Grid>
        </Box>
      </Box>
      
      <Paper elevation={3} sx={{ borderRadius: 2 }}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs 
            value={tabValue} 
            onChange={handleTabChange} 
            aria-label="bureau tabs"
            variant="scrollable"
            scrollButtons="auto"
          >
            <Tab label={t('bureau.lista_vinculos')} id="tab-0" aria-controls="tabpanel-0" />
            {selectedIdentity && (
              <Tab label={t('bureau.detalhes_vinculo')} id="tab-1" aria-controls="tabpanel-1" />
            )}
            {canCreateBureauIdentity && (
              <Tab label={t('bureau.novo_vinculo')} id="tab-2" aria-controls="tabpanel-2" />
            )}
          </Tabs>
        </Box>
        
        {/* Tab: Lista de Vínculos */}
        <TabPanel value={tabValue} index={0}>
          <Box sx={{ p: 2 }}>
            <Typography variant="h6" component="h3" gutterBottom>
              {t('bureau.vinculos_bureau')}
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph>
              {t('bureau.descricao_vinculos')}
            </Typography>
            
            {loadingIdentities ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                <CircularProgress />
              </Box>
            ) : (
              <BureauIdentitiesList
                onViewDetails={handleViewDetails}
                onRevokeIdentity={canRevokeBureauIdentity ? handleRevokeIdentity : undefined}
              />
            )}
          </Box>
        </TabPanel>
        
        {/* Tab: Detalhes do Vínculo */}
        {selectedIdentity && (
          <TabPanel value={tabValue} index={1}>
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" component="h3" gutterBottom>
                {t('bureau.detalhes_vinculo')} - {selectedIdentity.usuarioNome}
              </Typography>
              
              {loadingIdentity ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                  <CircularProgress />
                </Box>
              ) : (
                <Grid container spacing={3}>
                  {/* Seção de Informações do Vínculo */}
                  <Grid item xs={12}>
                    <Paper elevation={1} sx={{ p: 2, mb: 3 }}>
                      <Typography variant="subtitle1" gutterBottom>
                        {t('bureau.informacoes_vinculo')}
                      </Typography>
                      <Grid container spacing={2}>
                        <Grid item xs={12} sm={6} md={4}>
                          <Typography variant="body2" color="text.secondary">
                            {t('bureau.tipo_vinculo')}
                          </Typography>
                          <Typography variant="body1">
                            {t(`bureau.tipo_${selectedIdentity.tipoVinculo.toLowerCase()}`)}
                          </Typography>
                        </Grid>
                        <Grid item xs={12} sm={6} md={4}>
                          <Typography variant="body2" color="text.secondary">
                            {t('bureau.nivel_acesso')}
                          </Typography>
                          <Typography variant="body1">
                            {t(`bureau.nivel_${selectedIdentity.nivelAcesso.toLowerCase()}`)}
                          </Typography>
                        </Grid>
                        <Grid item xs={12} sm={6} md={4}>
                          <Typography variant="body2" color="text.secondary">
                            {t('bureau.status_vinculo')}
                          </Typography>
                          <Typography 
                            variant="body1"
                            sx={{ 
                              color: selectedIdentity.status === 'ATIVO' 
                                ? 'success.main' 
                                : 'error.main' 
                            }}
                          >
                            {t(`bureau.status_${selectedIdentity.status.toLowerCase()}`)}
                          </Typography>
                        </Grid>
                      </Grid>
                    </Paper>
                  </Grid>
                  
                  {/* Abas para Gestão de Autorizações, Tokens e Revogação */}
                  <Grid item xs={12}>
                    <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
                      <Tabs 
                        value={selectedIdentity.status === 'ATIVO' ? 0 : 2} 
                        aria-label="identity management tabs"
                      >
                        <Tab 
                          label={t('bureau.criar_autorizacao')} 
                          disabled={selectedIdentity.status !== 'ATIVO' || !canManageBureauAutorizacoes}
                        />
                        <Tab 
                          label={t('bureau.gerar_token')} 
                          disabled={selectedIdentity.status !== 'ATIVO' || !canGenerateBureauTokens}
                        />
                        <Tab 
                          label={t('bureau.revogar_vinculo')} 
                          disabled={!canRevokeBureauIdentity}
                        />
                      </Tabs>
                    </Box>
                    
                    {selectedIdentity.status === 'ATIVO' ? (
                      <>
                        {canManageBureauAutorizacoes && (
                          <Box sx={{ mb: 4 }}>
                            <BureauAutorizacaoForm 
                              identityId={selectedIdentity.id}
                              loading={loadingCreateAutorizacao}
                              onSubmit={handleCreateAutorizacao}
                            />
                            <Divider sx={{ my: 4 }} />
                          </Box>
                        )}
                        
                        {canGenerateBureauTokens && (
                          <Box>
                            <BureauTokenForm
                              identityId={selectedIdentity.id}
                              tipoVinculo={selectedIdentity.tipoVinculo}
                              nivelAcesso={selectedIdentity.nivelAcesso}
                              loading={loadingGenerateToken}
                              onSubmit={handleGenerateToken}
                            />
                          </Box>
                        )}
                      </>
                    ) : (
                      <Alert severity="warning" sx={{ mb: 3 }}>
                        {t('bureau.vinculo_inativo_aviso')}
                      </Alert>
                    )}
                    
                    {canRevokeBureauIdentity && (
                      <>
                        <Divider sx={{ my: 4 }} />
                        <BureauRevokeForm
                          identityId={selectedIdentity.id}
                          loading={loadingRevokeIdentity}
                          onSubmit={handleRevokeIdentity}
                        />
                      </>
                    )}
                  </Grid>
                </Grid>
              )}
            </Box>
          </TabPanel>
        )}
        
        {/* Tab: Novo Vínculo */}
        {canCreateBureauIdentity && (
          <TabPanel value={tabValue} index={selectedIdentity ? 2 : 1}>
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" component="h3" gutterBottom>
                {t('bureau.novo_vinculo')}
              </Typography>
              <Typography variant="body2" color="text.secondary" paragraph>
                {t('bureau.descricao_novo_vinculo')}
              </Typography>
              
              <BureauIdentityForm
                loading={loadingCreateIdentity}
                onSubmit={handleCreateIdentity}
              />
            </Box>
          </TabPanel>
        )}
      </Paper>
    </Container>
  );
};

export default BureauIntegrationPage;