// ==============================================================================
// Nome: BureauIdentityDetails.tsx
// Descrição: Componente para exibir detalhes de um vínculo com Bureau de Créditos
// Autor: Equipa de Desenvolvimento INNOVABIZ
// Data: 19/08/2025
// ==============================================================================

import React from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Chip,
  Divider,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
  Alert,
  Card,
  CardHeader,
  CardContent,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Tabs,
  Tab,
  IconButton,
  Tooltip,
  useTheme
} from '@mui/material';
import {
  Person as PersonIcon,
  History as HistoryIcon,
  Description as DescriptionIcon,
  Link as LinkIcon,
  VpnKey as VpnKeyIcon,
  Schedule as ScheduleIcon,
  MoreHoriz as MoreHorizIcon
} from '@mui/icons-material';
import { useTranslation } from '../../../hooks/useTranslation';
import { formatDateTimeWithTZ, formatRelativeTime } from '../../../utils/dateUtils';
import { JsonViewer } from '../../../components/JsonViewer';

// Tipos
import { 
  BureauIdentity, 
  BureauAutorizacao, 
  BureauEvento,
  BureauVinculoStatus
} from '../../../types/bureau-credito';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`details-tabpanel-${index}`}
      aria-labelledby={`details-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ pt: 2 }}>
          {children}
        </Box>
      )}
    </div>
  );
}

interface BureauIdentityDetailsProps {
  loading: boolean;
  error: any;
  identity?: BureauIdentity;
}

/**
 * Componente para exibir detalhes completos de um vínculo com Bureau de Créditos
 * 
 * Este componente mostra:
 * - Informações básicas do vínculo (identificadores, status, datas)
 * - Detalhes do usuário vinculado
 * - Histórico de autorizações
 * - Eventos de auditoria relacionados
 * - Dados técnicos completos para usuários avançados
 */
export const BureauIdentityDetails: React.FC<BureauIdentityDetailsProps> = ({
  loading,
  error,
  identity
}) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const [tabValue, setTabValue] = React.useState(0);

  // Função para mudar a aba atual
  const handleChangeTab = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  // Função para obter cor do chip de status
  const getStatusColor = (status: BureauVinculoStatus) => {
    switch (status) {
      case 'ativo':
        return 'success';
      case 'pendente':
        return 'warning';
      case 'revogado':
        return 'error';
      case 'expirado':
        return 'default';
      default:
        return 'default';
    }
  };

  // Se estiver carregando, mostrar indicador
  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  // Se houver erro, mostrar mensagem
  if (error) {
    return (
      <Alert severity="error" sx={{ mb: 2 }}>
        {t('bureau.erro_carregar_detalhes')}: {error.message}
      </Alert>
    );
  }

  // Se não houver dados de identidade, mostrar mensagem
  if (!identity) {
    return (
      <Alert severity="info" sx={{ mb: 2 }}>
        {t('bureau.selecione_vinculo')}
      </Alert>
    );
  }

  return (
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', p: 3 }}>
        {/* Cabeçalho com informações básicas */}
        <Grid container spacing={2} alignItems="center" sx={{ mb: 3 }}>
          <Grid item xs={12} md={6}>
            <Typography variant="h6" component="div" sx={{ fontWeight: 'bold' }}>
              {t('bureau.detalhe_vinculo')}
            </Typography>
            <Typography variant="body2" color="textSecondary">
              ID: {identity.id}
            </Typography>
          </Grid>
          <Grid item xs={12} md={6} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Chip 
              label={t(`bureau.status_${identity.status}`)} 
              color={getStatusColor(identity.status) as any}
              sx={{ ml: 1 }}
            />
            <Chip 
              label={t(`bureau.tipo_${identity.tipoVinculo.toLowerCase()}`)} 
              color="primary"
              variant="outlined"
              sx={{ ml: 1 }}
            />
            <Chip 
              label={t(`bureau.nivel_${identity.nivelAcesso.toLowerCase()}`)} 
              color="secondary"
              variant="outlined"
              sx={{ ml: 1 }}
            />
          </Grid>
        </Grid>

        {/* Abas para organizar o conteúdo */}
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs 
            value={tabValue} 
            onChange={handleChangeTab}
            aria-label="detalhes do vínculo"
            variant="scrollable"
            scrollButtons="auto"
          >
            <Tab label={t('bureau.tab_informacoes')} id="details-tab-0" aria-controls="details-tabpanel-0" />
            <Tab label={t('bureau.tab_autorizacoes')} id="details-tab-1" aria-controls="details-tabpanel-1" />
            <Tab label={t('bureau.tab_eventos')} id="details-tab-2" aria-controls="details-tabpanel-2" />
            <Tab label={t('bureau.tab_dados_tecnicos')} id="details-tab-3" aria-controls="details-tabpanel-3" />
          </Tabs>
        </Box>

        {/* Painel de informações básicas */}
        <TabPanel value={tabValue} index={0}>
          <Grid container spacing={3}>
            {/* Informações do vínculo */}
            <Grid item xs={12} md={6}>
              <Card>
                <CardHeader 
                  title={t('bureau.informacoes_vinculo')}
                  avatar={<LinkIcon />}
                />
                <CardContent>
                  <List>
                    <ListItem>
                      <ListItemIcon>
                        <VpnKeyIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.id_interno')} 
                        secondary={identity.id} 
                      />
                    </ListItem>
                    <ListItem>
                      <ListItemIcon>
                        <VpnKeyIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.id_externo')} 
                        secondary={identity.externalId || t('comum.nao_disponivel')} 
                      />
                    </ListItem>
                    <ListItem>
                      <ListItemIcon>
                        <VpnKeyIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.tenant_externo')} 
                        secondary={identity.externalTenantId || t('comum.nao_disponivel')} 
                      />
                    </ListItem>
                    <ListItem>
                      <ListItemIcon>
                        <ScheduleIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.data_criacao')} 
                        secondary={`${formatDateTimeWithTZ(identity.dataCriacao)} (${formatRelativeTime(identity.dataCriacao)})`} 
                      />
                    </ListItem>
                    <ListItem>
                      <ListItemIcon>
                        <ScheduleIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.data_atualizacao')} 
                        secondary={`${formatDateTimeWithTZ(identity.dataAtualizacao)} (${formatRelativeTime(identity.dataAtualizacao)})`} 
                      />
                    </ListItem>
                    {identity.motivoStatus && (
                      <ListItem>
                        <ListItemIcon>
                          <DescriptionIcon />
                        </ListItemIcon>
                        <ListItemText 
                          primary={t('bureau.motivo_status')} 
                          secondary={identity.motivoStatus} 
                        />
                      </ListItem>
                    )}
                  </List>
                </CardContent>
              </Card>
            </Grid>

            {/* Informações do usuário */}
            <Grid item xs={12} md={6}>
              <Card>
                <CardHeader 
                  title={t('bureau.informacoes_usuario')}
                  avatar={<PersonIcon />}
                />
                <CardContent>
                  <List>
                    <ListItem>
                      <ListItemIcon>
                        <VpnKeyIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.usuario_id')} 
                        secondary={identity.usuarioId} 
                      />
                    </ListItem>
                    <ListItem>
                      <ListItemIcon>
                        <VpnKeyIcon />
                      </ListItemIcon>
                      <ListItemText 
                        primary={t('bureau.tenant_id')} 
                        secondary={identity.tenantId} 
                      />
                    </ListItem>
                  </List>
                </CardContent>
              </Card>

              {/* Detalhes adicionais */}
              {identity.detalhes && Object.keys(identity.detalhes).length > 0 && (
                <Card sx={{ mt: 2 }}>
                  <CardHeader 
                    title={t('bureau.detalhes_adicionais')}
                    avatar={<MoreHorizIcon />}
                  />
                  <CardContent>
                    <List>
                      {Object.entries(identity.detalhes).map(([key, value]) => (
                        <ListItem key={key}>
                          <ListItemText 
                            primary={key} 
                            secondary={typeof value === 'string' ? value : JSON.stringify(value)} 
                          />
                        </ListItem>
                      ))}
                    </List>
                  </CardContent>
                </Card>
              )}
            </Grid>
          </Grid>
        </TabPanel>

        {/* Painel de autorizações */}
        <TabPanel value={tabValue} index={1}>
          {!identity.autorizacoes || identity.autorizacoes.length === 0 ? (
            <Alert severity="info">
              {t('bureau.nenhuma_autorizacao')}
            </Alert>
          ) : (
            <TableContainer>
              <Table size="medium">
                <TableHead>
                  <TableRow>
                    <TableCell>{t('comum.id')}</TableCell>
                    <TableCell>{t('bureau.tipo_consulta')}</TableCell>
                    <TableCell>{t('bureau.finalidade')}</TableCell>
                    <TableCell>{t('bureau.data_autorizacao')}</TableCell>
                    <TableCell>{t('bureau.data_validade')}</TableCell>
                    <TableCell>{t('bureau.status')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {identity.autorizacoes.map((autorizacao) => (
                    <TableRow key={autorizacao.id}>
                      <TableCell>{autorizacao.id.substring(0, 8)}...</TableCell>
                      <TableCell>
                        <Chip 
                          label={t(`bureau.consulta_${autorizacao.tipoConsulta.toLowerCase()}`)} 
                          size="small"
                          color="primary"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>{autorizacao.finalidade}</TableCell>
                      <TableCell>{formatDateTimeWithTZ(autorizacao.dataAutorizacao)}</TableCell>
                      <TableCell>{formatDateTimeWithTZ(autorizacao.dataValidade)}</TableCell>
                      <TableCell>
                        <Chip 
                          label={t(`bureau.status_autorizacao_${autorizacao.status}`)} 
                          size="small"
                          color={autorizacao.status === 'ativa' ? 'success' : 
                                 autorizacao.status === 'pendente' ? 'warning' : 'default'} 
                        />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </TabPanel>

        {/* Painel de eventos */}
        <TabPanel value={tabValue} index={2}>
          {!identity.eventos || identity.eventos.length === 0 ? (
            <Alert severity="info">
              {t('bureau.nenhum_evento')}
            </Alert>
          ) : (
            <TableContainer>
              <Table size="medium">
                <TableHead>
                  <TableRow>
                    <TableCell>{t('comum.id')}</TableCell>
                    <TableCell>{t('bureau.tipo_evento')}</TableCell>
                    <TableCell>{t('bureau.descricao')}</TableCell>
                    <TableCell>{t('bureau.timestamp')}</TableCell>
                    <TableCell>{t('bureau.operador')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {identity.eventos.map((evento) => (
                    <TableRow key={evento.id}>
                      <TableCell>{evento.id.substring(0, 8)}...</TableCell>
                      <TableCell>
                        <Chip 
                          label={t(`bureau.evento_${evento.tipo.toLowerCase()}`)} 
                          size="small"
                          color={
                            evento.tipo.includes('CREATE') ? 'success' :
                            evento.tipo.includes('UPDATE') ? 'info' :
                            evento.tipo.includes('DELETE') || evento.tipo.includes('REVOKE') ? 'error' :
                            'default'
                          }
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>{evento.descricao}</TableCell>
                      <TableCell>{formatDateTimeWithTZ(evento.timestamp)}</TableCell>
                      <TableCell>{evento.operadorId || t('comum.sistema')}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </TabPanel>

        {/* Painel de dados técnicos */}
        <TabPanel value={tabValue} index={3}>
          <Typography variant="body2" color="textSecondary" paragraph>
            {t('bureau.dados_tecnicos_descricao')}
          </Typography>
          <JsonViewer data={identity} />
        </TabPanel>
      </Paper>
    </Box>
  );
};